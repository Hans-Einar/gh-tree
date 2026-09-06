package launchdiscovery

import (
	"golang.org/x/sys/unix"
	"strings"
)

func unixStamp(_ int, st *unix.Stat_t, profile string) (string, error) {
	if profile == "" || strings.HasPrefix(profile, "birth:") {
		return stamp("birth", int64(st.Btim.Sec), int64(st.Btim.Nsec)), nil
	}
	if strings.HasPrefix(profile, "change:") {
		return stamp("change", int64(st.Ctim.Sec), int64(st.Ctim.Nsec)), nil
	}
	return "", errRedirect
}
