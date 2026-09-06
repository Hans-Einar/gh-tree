package persistence

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"structs"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

const winShareAll = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE

// FileIdInfo requires native alignment as well as its 24-byte field layout.
// A byte array of the right length does not guarantee adequate ABI alignment.
type winFileIDInfo struct {
	_      structs.HostLayout
	Volume uint64
	File   [16]byte
}

var _ [24 - unsafe.Sizeof(winFileIDInfo{})]byte
var _ [unsafe.Sizeof(winFileIDInfo{}) - 24]byte
var _ [8 - unsafe.Offsetof(winFileIDInfo{}.File)]byte
var _ [unsafe.Offsetof(winFileIDInfo{}.File) - 8]byte

type winObservation struct {
	id    winFileIDInfo
	basic windows.ByHandleFileInformation
}

func winObserve(handle windows.Handle) (winObservation, error) {
	var result winObservation
	if err := windows.GetFileInformationByHandleEx(handle, windows.FileIdInfo, (*byte)(unsafe.Pointer(&result.id)), uint32(unsafe.Sizeof(result.id))); err != nil {
		return result, err
	}
	if err := windows.GetFileInformationByHandle(handle, &result.basic); err != nil {
		return result, err
	}
	if result.id.File == ([16]byte{}) {
		return result, errors.New("native file identity unavailable")
	}
	if result.basic.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return result, fmt.Errorf("%w: storage refuses reparse object", errUnsupportedProfile)
	}
	return result, nil
}
func (o winObservation) directoryIdentity() (api.DirectoryIdentity, error) {
	if o.basic.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		return api.DirectoryIdentity{}, errors.New("not a directory")
	}
	birth := uint64(o.basic.CreationTime.HighDateTime)<<32 | uint64(o.basic.CreationTime.LowDateTime)
	return api.NewDirectoryIdentity(api.DirectoryWindows, o.id.Volume, o.id.File, "birth-filetime:"+strconv.FormatUint(birth, 10))
}
func (o winObservation) sameObject(other winObservation) bool {
	return o.id.Volume == other.id.Volume && o.id.File == other.id.File && o.basic.CreationTime == other.basic.CreationTime
}
func (o winObservation) sameRead(other winObservation) bool {
	return o.sameObject(other) && o.basic.FileSizeHigh == other.basic.FileSizeHigh && o.basic.FileSizeLow == other.basic.FileSizeLow && o.basic.LastWriteTime == other.basic.LastWriteTime && o.basic.FileAttributes == other.basic.FileAttributes
}
func (o winObservation) size() uint64 {
	return uint64(o.basic.FileSizeHigh)<<32 | uint64(o.basic.FileSizeLow)
}

type winObject struct {
	file            *os.File
	observation     winObservation
	createdArtifact diskIdentity // Set only by exclusive owned-artifact creation.
}

func (o *winObject) handle() windows.Handle { return windows.Handle(o.file.Fd()) }
func (o *winObject) close() error           { return o.file.Close() }

// winOpen performs exactly one native open relative to a retained parent. Root
// acquisition alone may use an absolute NT name. Every descendant is one literal
// basename. No inheritance or ordinary path-based reopen is permitted here.
func winOpen(parent windows.Handle, name string, access, share, disposition, options uint32) (*winObject, error) {
	return winOpenWithSecurity(parent, name, access, share, disposition, options, nil)
}

func winOpenWithSecurity(parent windows.Handle, name string, access, share, disposition, options uint32, security *windows.SECURITY_DESCRIPTOR) (*winObject, error) {
	if parent != 0 && !singleName(name) {
		return nil, errors.New("invalid native basename")
	}
	u, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	a := windows.OBJECT_ATTRIBUTES{RootDirectory: parent, ObjectName: u, Attributes: windows.OBJ_DONT_REPARSE}
	a.SecurityDescriptor = security
	a.Length = uint32(unsafe.Sizeof(a))
	var handle windows.Handle
	err = windows.NtCreateFile(&handle, access|windows.SYNCHRONIZE, &a, &windows.IO_STATUS_BLOCK{}, nil, windows.FILE_ATTRIBUTE_NORMAL, share, disposition, options|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	runtime.KeepAlive(security)
	if err != nil {
		if status, ok := err.(windows.NTStatus); ok {
			err = errors.Join(status, status.Errno())
		}
		return nil, err
	}
	observation, err := winObserve(handle)
	if err != nil {
		return nil, errors.Join(err, windows.CloseHandle(handle))
	}
	file := os.NewFile(uintptr(handle), name)
	if file == nil {
		return nil, errors.Join(errors.New("cannot own native file"), windows.CloseHandle(handle))
	}
	return &winObject{file: file, observation: observation}, nil
}
func winOpenDirectory(parent windows.Handle, name string) (*winObject, error) {
	// FILE_GENERIC_READ includes actual FILE_LIST_DIRECTORY/data-read access.
	// Metadata-only handles cannot interlock an in-place reparse conversion.
	return winOpen(parent, name, windows.FILE_GENERIC_READ|windows.FILE_TRAVERSE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE)
}
func winOpenDocument(parent *winObject, name string) (*winObject, error) {
	return winOpen(parent.handle(), name, windows.FILE_GENERIC_READ, winShareAll, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
}

// Every ancestor guard is owned by this one request, released in reverse order.
// Missing ancestry returns its proved anchor and remaining literal components;
// the read path creates no directory, document or lock.
type winChain struct {
	guards    []*winObject
	remaining []string
}

func (c *winChain) parent() *winObject { return c.guards[len(c.guards)-1] }
func (c *winChain) close() error {
	var err error
	for i := len(c.guards) - 1; i >= 0; i-- {
		err = errors.Join(err, c.guards[i].close())
	}
	c.guards = nil
	return err
}
func winAcquire(ctx context.Context, path string) (_ *winChain, resultErr error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.HasPrefix(path, `\\?\`) {
		path = path[4:]
	}
	if !filepath.IsAbs(path) || len(path) < 3 || path[1] != ':' || path[2] != '\\' {
		return nil, errors.New("storage requires an absolute local volume path")
	}
	if path[0] < 'A' || path[0] > 'Z' && path[0] < 'a' || path[0] > 'z' {
		return nil, errors.New("unsupported volume path")
	}
	parts := strings.Split(path[3:], `\`)
	if len(parts) == 1 && parts[0] == "" {
		parts = nil
	}
	for _, part := range parts {
		if !singleName(part) {
			return nil, errors.New("invalid directory component")
		}
	}
	root, err := winOpenDirectory(0, `\??\`+path[:3])
	if err != nil {
		return nil, err
	}
	c := &winChain{guards: []*winObject{root}}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, c.close())
		}
	}()
	var fileSystem [32]uint16
	if err := windows.GetVolumeInformationByHandle(root.handle(), nil, 0, nil, nil, nil, &fileSystem[0], uint32(len(fileSystem))); err != nil {
		return nil, err
	}
	if windows.UTF16ToString(fileSystem[:]) != "NTFS" {
		return nil, fmt.Errorf("%w: local NTFS required", errUnsupportedProfile)
	}
	rootPath, _ := windows.UTF16PtrFromString(path[:3])
	if driveType := windows.GetDriveType(rootPath); driveType != windows.DRIVE_FIXED && driveType != windows.DRIVE_REMOVABLE {
		return nil, fmt.Errorf("%w: nonlocal volume", errUnsupportedProfile)
	}
	for i, part := range parts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		child, err := winOpenDirectory(c.parent().handle(), part)
		if errors.Is(err, os.ErrNotExist) {
			c.remaining = append([]string{}, parts[i:]...)
			return c, nil
		}
		if err != nil {
			return nil, err
		}
		c.guards = append(c.guards, child)
	}
	return c, nil
}

func winRead(ctx context.Context, object *winObject) ([]byte, winObservation, error) {
	read := func() ([]byte, winObservation, error) {
		before, err := winObserve(object.handle())
		if err != nil {
			return nil, before, err
		}
		if before.basic.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
			return nil, before, errUnsupportedProfile
		}
		if before.size() > api.MaxDocumentBytes {
			return nil, before, corrupt("document exceeds 4 MiB")
		}
		if _, err := object.file.Seek(0, io.SeekStart); err != nil {
			return nil, before, err
		}
		data := make([]byte, 0, int(before.size()))
		buffer := make([]byte, 32*1024)
		for {
			if err := ctx.Err(); err != nil {
				return nil, before, err
			}
			n, err := object.file.Read(buffer)
			if n > 0 {
				if len(data)+n > api.MaxDocumentBytes {
					return nil, before, corrupt("document exceeds 4 MiB")
				}
				data = append(data, buffer[:n]...)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, before, err
			}
			if n == 0 {
				return nil, before, io.ErrNoProgress
			}
		}
		after, err := winObserve(object.handle())
		if err != nil {
			return nil, before, err
		}
		if !before.sameRead(after) || uint64(len(data)) != after.size() {
			return nil, after, errors.New("document changed during read")
		}
		return data, after, nil
	}
	first, observation, err := read()
	if err != nil {
		return nil, observation, err
	}
	second, final, err := read()
	if err != nil {
		return nil, final, err
	}
	if !observation.sameRead(final) || string(first) != string(second) {
		return nil, final, errors.New("document changed between bounded reads")
	}
	return first, final, nil
}

type winStoreLock struct{ object *winObject }

func winLock(ctx context.Context, parent *winObject, basename string, budget time.Duration) (_ *winStoreLock, resultErr error) {
	return winLockMode(ctx, parent, basename, budget, true)
}
func winLockMode(ctx context.Context, parent *winObject, basename string, budget time.Duration, create bool) (_ *winStoreLock, resultErr error) {
	return winLockSecurity(ctx, parent, basename, budget, create, nil)
}
func winLockSecurity(ctx context.Context, parent *winObject, basename string, budget time.Duration, create bool, security *windows.SECURITY_DESCRIPTOR) (_ *winStoreLock, resultErr error) {
	if !singleName(basename) || budget <= 0 || budget > 5*time.Second {
		return nil, errors.New("invalid lock parameters")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(budget)
	disposition := uint32(windows.FILE_OPEN)
	if create {
		disposition = windows.FILE_OPEN_IF
	}
	object, err := winOpenWithSecurity(parent.handle(), basename+".lock", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, disposition, windows.FILE_NON_DIRECTORY_FILE, security)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, object.close())
		}
	}()
	if object.observation.basic.NumberOfLinks != 1 {
		return nil, errors.New("unsupported linked lock object")
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if time.Until(deadline) <= 0 {
			return nil, fmt.Errorf("%w after %s", errLockBusy, budget)
		}
		err := windows.LockFileEx(object.handle(), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
		if err == nil {
			return &winStoreLock{object}, nil
		}
		if !errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, err
		}
		left := time.Until(deadline)
		if left <= 0 {
			return nil, fmt.Errorf("%w after %s", errLockBusy, budget)
		}
		if left > 10*time.Millisecond {
			left = 10 * time.Millisecond
		}
		timer := time.NewTimer(left)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
func (l *winStoreLock) close() error {
	return errors.Join(windows.UnlockFileEx(l.object.handle(), 0, 1, 0, &windows.Overlapped{}), l.object.close())
}
