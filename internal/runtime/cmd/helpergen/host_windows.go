package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func nativeHost() error {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("IsWow64Process2")
	if e := proc.Find(); e != nil {
		return e
	}
	var process, native uint16
	r, _, e := proc.Call(^uintptr(0), uintptr(unsafe.Pointer(&process)), uintptr(unsafe.Pointer(&native)))
	if r == 0 {
		return e
	}
	if process != 0 || native != 0x8664 {
		return fmt.Errorf("canonical helper builder is emulated or not native amd64: process=%#x native=%#x", process, native)
	}
	return nil
}
