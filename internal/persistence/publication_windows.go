package persistence

import (
	"errors"
	"runtime"
	"structs"
	"unsafe"

	"golang.org/x/sys/windows"
)

// These private primitives require the request owner to have completed its
// expected-version, security, retention and manifest barriers. They perform one
// selected native operation and never retry with a different mechanism.
const (
	winRenameInformationEx = 65
	winLinkInformation     = 11        // x/sys FileLinkInformation is the distinct class 72.
	winReplacePresent      = 0x1 | 0x2 // REPLACE_IF_EXISTS | POSIX_SEMANTICS
)

type winNameInformation struct {
	_      structs.HostLayout
	Flags  uint32
	Root   windows.Handle
	Length uint32
	Name   [1]uint16
}

func winSetName(object, parent *winObject, name string, flags, class uint32) error {
	if object == nil || parent == nil || !singleName(name) ||
		(class != winRenameInformationEx && class != winLinkInformation) ||
		(class == winLinkInformation && flags != 0) ||
		(class == winRenameInformationEx && flags != 0 && flags != winReplacePresent) {
		return errors.New("invalid native publication arguments")
	}
	// A retained empty directory can be converted to a reparse point. Refuse
	// observed conversion before the call; native relative resolution also uses
	// that retained object, without reopening its former pathname.
	if _, err := winObserve(parent.handle()); err != nil {
		return err
	}
	name16, err := windows.UTF16FromString(name)
	if err != nil {
		return err
	}
	name16 = name16[:len(name16)-1]
	if len(name16) > 32767 {
		return errors.New("native name too long")
	}
	size := unsafe.Offsetof(winNameInformation{}.Name) + uintptr(len(name16)*2)
	// Allocate native words, not bytes: both the address and field offsets are
	// explicitly pointer-aligned on Windows386, amd64 and ARM64.
	words := make([]uintptr, (size+unsafe.Sizeof(uintptr(0))-1)/unsafe.Sizeof(uintptr(0)))
	info := (*winNameInformation)(unsafe.Pointer(&words[0]))
	info.Flags, info.Root, info.Length = flags, parent.handle(), uint32(len(name16)*2)
	copy(unsafe.Slice(&info.Name[0], len(name16)), name16)
	err = windows.NtSetInformationFile(object.handle(), &windows.IO_STATUS_BLOCK{}, (*byte)(unsafe.Pointer(info)), uint32(size), class)
	runtime.KeepAlive(words)
	if status, ok := err.(windows.NTStatus); ok {
		return errors.Join(status, status.Errno())
	}
	return err
}

func winPublish(payload, parent *winObject, name string, expectedPresent bool) error {
	flags := uint32(0)
	if expectedPresent {
		flags = winReplacePresent
	}
	return winSetName(payload, parent, name, flags, winRenameInformationEx)
}

func winRetainOriginal(original, parent *winObject, name string) (*winObject, error) {
	if err := winSetName(original, parent, name, 0, winLinkInformation); err != nil {
		return nil, err
	}
	retained, err := winOpenDocument(parent, name)
	if err != nil {
		return nil, err
	} // The owned link remains for recovery.
	if !original.observation.sameObject(retained.observation) {
		return nil, errors.Join(errors.New("retained original identity differs"), retained.close())
	}
	return retained, nil
}
