//go:build freebsd

package git

import (
	"fmt"
	"os"
	"syscall"
)

func directoryStamp(_ *os.File, s *syscall.Stat_t) string {
	if s.Birthtimespec.Sec == 0 && s.Birthtimespec.Nsec == 0 {
		return fmt.Sprintf("change:%d:%d", s.Ctimespec.Sec, s.Ctimespec.Nsec)
	}
	return fmt.Sprintf("birth:%d:%d", s.Birthtimespec.Sec, s.Birthtimespec.Nsec)
}
