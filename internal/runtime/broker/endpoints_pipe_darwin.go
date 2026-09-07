package broker

import "golang.org/x/sys/unix"

func anonymousNativePipe(fd int) bool {
	// XNU anonymous pipes are DTYPE_PIPE, not vnodes: fstatfs64's
	// file_vnode lookup returns EINVAL. A linked or unlinked named FIFO is
	// a vnode and reports its filesystem. S_IFIFO/direction were checked
	// separately, so other non-vnode descriptor types cannot pass.
	// bsd/{kern/sys_pipe.c,kern/kern_descrip.c,vfs/vfs_syscalls.c}.
	var fs unix.Statfs_t
	return unix.Fstatfs(fd, &fs) == unix.EINVAL
}
