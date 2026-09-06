package broker

import (
	"golang.org/x/sys/unix"
	"strings"
)

func directoryStamp(fd int, st *unix.Stat_t, profile string) (string, error) {
	if profile == "" || strings.HasPrefix(profile, "birth:") {
		var sx unix.Statx_t
		err := unix.Statx(fd, "", unix.AT_EMPTY_PATH|unix.AT_SYMLINK_NOFOLLOW, unix.STATX_BTIME, &sx)
		if err == nil && sx.Mask&unix.STATX_BTIME != 0 {
			return stamp("birth", sx.Btime.Sec, int64(sx.Btime.Nsec)), nil
		}
		if profile != "" {
			return "", ErrCwd
		}
	}
	if profile == "" || strings.HasPrefix(profile, "change:") {
		return stamp("change", int64(st.Ctim.Sec), int64(st.Ctim.Nsec)), nil
	}
	return "", ErrCwd
}
