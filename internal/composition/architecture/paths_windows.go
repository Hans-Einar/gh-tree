package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var finalPathName = syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")

const fileReadAttributes = 0x0080

// Go 1.25's EvalSymlinks does not resolve every Windows directory junction.
// A metadata-only, fully shared handle asks Windows for the actual final object
// pathname. It is closed before returning; this is inspection, not a retained
// Runtime/Persistence capability or a promise against concurrent tree changes.
func physicalPath(path string) (resolved string, err error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(abs, `\\?\`) {
		if strings.HasPrefix(abs, `\\`) {
			abs = `\\?\UNC\` + strings.TrimPrefix(abs, `\\`)
		} else {
			abs = `\\?\` + abs
		}
	}
	name, err := syscall.UTF16PtrFromString(abs)
	if err != nil {
		return "", err
	}
	handle, err := syscall.CreateFile(name, fileReadAttributes, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", fmt.Errorf("open source metadata %s: %w", path, err)
	}
	defer func() { err = errors.Join(err, syscall.CloseHandle(handle)) }()
	buffer := make([]uint16, 512)
	for {
		n, _, callErr := finalPathName.Call(uintptr(handle), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0)
		if n == 0 {
			return "", fmt.Errorf("resolve source object %s: %w", path, callErr)
		}
		if n >= uintptr(len(buffer)) {
			if n > 65536 {
				return "", fmt.Errorf("source object path exceeds inspection bound: %s", path)
			}
			buffer = make([]uint16, int(n)+1)
			continue
		}
		resolved = syscall.UTF16ToString(buffer[:int(n)])
		if strings.HasPrefix(resolved, `\\?\UNC\`) {
			resolved = `\\` + strings.TrimPrefix(resolved, `\\?\UNC\`)
		} else {
			resolved = strings.TrimPrefix(resolved, `\\?\`)
		}
		return filepath.Clean(resolved), nil
	}
}
