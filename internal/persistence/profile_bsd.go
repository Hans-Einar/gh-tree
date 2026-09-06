//go:build darwin || freebsd

package persistence

import (
	"bytes"
	"fmt"
	"golang.org/x/sys/unix"
)

func unixBirthStamp(_ int, st *unix.Stat_t) string {
	if st.Btim.Sec != 0 || st.Btim.Nsec != 0 {
		return fmt.Sprintf("birth:%d:%d", st.Btim.Sec, st.Btim.Nsec)
	}
	return fmt.Sprintf("change:%d:%d", st.Ctim.Sec, st.Ctim.Nsec)
}
func unixLocalFileSystem(fd int) (string, error) {
	var fs unix.Statfs_t
	if err := unix.Fstatfs(fd, &fs); err != nil {
		return "", err
	}
	name := string(bytes.TrimRight(fs.Fstypename[:], "\x00"))
	switch name {
	case "apfs", "hfs", "ufs", "zfs", "tmpfs":
		return name, nil
	}
	return "", fmt.Errorf("unsupported storage filesystem %s", name)
}
