package broker

import "golang.org/x/sys/unix"

func validNativePipeAccess(fd, flags int, _ bool) bool {
	// FreeBSD kern_pipe initializes BOTH anonymous endpoints FREAD|FWRITE.
	// The two inherited endpoints still have distinct logical protocol roles.
	// Require a non-vnode FIFO: fstatfs returns EINVAL for DTYPE_PIPE, while a
	// named (or unlinked named) FIFO is a vnode with a real filesystem result.
	// FreeBSD releng/15.0 sys/kern/{sys_pipe.c,vfs_syscalls.c} define this profile.
	if flags&unix.O_ACCMODE != unix.O_RDWR {
		return false
	}
	var fs unix.Statfs_t
	return unix.Fstatfs(fd, &fs) == unix.EINVAL
}
