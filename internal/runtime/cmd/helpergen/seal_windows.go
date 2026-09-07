package main

import (
	"os"
	"syscall"
)

// Retain actual read handles until the compiler/linker have exited. Deny write
// and delete sharing on source/tool bytes; deny delete sharing on every input
// directory. Read-only attributes alone can be cleared and are not the barrier.
func openImmutableInput(path string, directory bool) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	flags := uint32(syscall.FILE_ATTRIBUTE_NORMAL)
	share := uint32(syscall.FILE_SHARE_READ)
	if directory {
		flags = syscall.FILE_FLAG_BACKUP_SEMANTICS
		share |= syscall.FILE_SHARE_WRITE
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ, share, nil, syscall.OPEN_EXISTING, flags, 0)
	if err != nil {
		return nil, &os.PathError{Op: "seal captured input", Path: path, Err: err}
	}
	return os.NewFile(uintptr(h), path), nil
}
