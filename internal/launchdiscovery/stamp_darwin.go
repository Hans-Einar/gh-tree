package launchdiscovery

import (
	"golang.org/x/sys/unix"
	"strings"
)

func unixStamp(_ int, st *unix.Stat_t, profile string) (string, error) {
	if profile == "" || strings.HasPrefix(profile, "birth:") {
		return stamp("birth", st.Birthtimespec.Sec, st.Birthtimespec.Nsec), nil
	}
	if strings.HasPrefix(profile, "change:") {
		return stamp("change", st.Ctimespec.Sec, st.Ctimespec.Nsec), nil
	}
	return "", errRedirect
}
