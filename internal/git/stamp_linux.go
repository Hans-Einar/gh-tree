//go:build linux

package git

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

func directoryStamp(f *os.File, s *syscall.Stat_t) string {
	var extended unix.Statx_t
	if err := unix.Statx(int(f.Fd()), "", unix.AT_EMPTY_PATH|unix.AT_SYMLINK_NOFOLLOW, unix.STATX_BTIME, &extended); err == nil && extended.Mask&unix.STATX_BTIME != 0 {
		return fmt.Sprintf("birth:%d:%d", extended.Btime.Sec, extended.Btime.Nsec)
	}
	return fmt.Sprintf("change:%d:%d", s.Ctim.Sec, s.Ctim.Nsec)
}
