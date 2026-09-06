package launchdiscovery

import (
	"golang.org/x/sys/unix"
	"strings"
)

func unixStamp(_ int, st *unix.Stat_t, profile string) (string, error) {
	if profile == "" || strings.HasPrefix(profile, "birth:") {
		return stamp("birth", st.Btim.Sec, st.Btim.Nsec), nil
	}
	if strings.HasPrefix(profile, "change:") {
		return stamp("change", st.Ctim.Sec, st.Ctim.Nsec), nil
	}
	return "", errRedirect
}
