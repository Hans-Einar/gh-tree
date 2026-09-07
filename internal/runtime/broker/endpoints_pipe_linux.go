package broker

import "golang.org/x/sys/unix"

func anonymousNativePipe(fd int) bool {
	// Linux allocates anonymous pipes in its internal, unmountable pipefs.
	// A filesystem FIFO retains its own superblock even after unlinking.
	// fs/pipe.c: get_pipe_inode, pipefs_init_fs_context.
	var fs unix.Statfs_t
	return unix.Fstatfs(fd, &fs) == nil && fs.Type == unix.PIPEFS_MAGIC
}
