//go:build linux || darwin

package broker

import "golang.org/x/sys/unix"

func validNativePipeAccess(_ int, flags int, write bool) bool {
	want := unix.O_RDONLY
	if write {
		want = unix.O_WRONLY
	}
	return flags&unix.O_ACCMODE == want
}
