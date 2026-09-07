package persistence

import (
	"fmt"
	"golang.org/x/sys/unix"
)

func unixBirthStamp(fd int, st *unix.Stat_t) string {
	var extended unix.Statx_t
	if err := unix.Statx(fd, "", unix.AT_EMPTY_PATH|unix.AT_SYMLINK_NOFOLLOW, unix.STATX_BTIME, &extended); err == nil && extended.Mask&unix.STATX_BTIME != 0 {
		return fmt.Sprintf("birth:%d:%d", extended.Btime.Sec, extended.Btime.Nsec)
	}
	return fmt.Sprintf("change:%d:%d", st.Ctim.Sec, st.Ctim.Nsec)
}
func unixLocalFileSystem(fd int) (string, error) {
	var fs unix.Statfs_t
	if err := unix.Fstatfs(fd, &fs); err != nil {
		return "", err
	}
	switch uint64(uint32(fs.Type)) {
	case unix.EXT4_SUPER_MAGIC:
		return "ext4", nil
	case unix.XFS_SUPER_MAGIC:
		return "xfs", nil
	case unix.BTRFS_SUPER_MAGIC:
		return "btrfs", nil
	case unix.TMPFS_MAGIC:
		return "tmpfs", nil
	}
	return "", fmt.Errorf("%w: filesystem type %x", errUnsupportedProfile, fs.Type)
}
