package launchdiscovery

import (
	"golang.org/x/sys/unix"
	"strings"
)

func unixStamp(fd int, st *unix.Stat_t, profile string) (string, error) {
	if profile == "" || strings.HasPrefix(profile, "birth:") {
		var x unix.Statx_t
		e := unix.Statx(fd, "", unix.AT_EMPTY_PATH|unix.AT_SYMLINK_NOFOLLOW, unix.STATX_BTIME, &x)
		if e == nil && x.Mask&unix.STATX_BTIME != 0 {
			return stamp("birth", x.Btime.Sec, int64(x.Btime.Nsec)), nil
		}
		if profile != "" {
			return "", errRedirect
		}
	}
	if profile == "" || strings.HasPrefix(profile, "change:") {
		return stamp("change", st.Ctim.Sec, st.Ctim.Nsec), nil
	}
	return "", errRedirect
}
