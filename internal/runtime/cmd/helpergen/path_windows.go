package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// EvalSymlinks in Go1.25 cannot resolve some canonical setup-go junctions.
// Resolve the opened object through the kernel instead: an alias of the selected
// root is supported, but contained still compares each child's physical path.
func physicalPath(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")
	buf := make([]uint16, 32768)
	n, _, err := proc.Call(f.Fd(), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
	if n == 0 {
		return "", err
	}
	if n >= uintptr(len(buf)) {
		return "", fmt.Errorf("physical path exceeds Windows path limit")
	}
	resolved := syscall.UTF16ToString(buf[:n])
	if strings.HasPrefix(resolved, `\\?\UNC\`) {
		return `\\` + strings.TrimPrefix(resolved, `\\?\UNC\`), nil
	}
	return strings.TrimPrefix(resolved, `\\?\`), nil
}
