//go:build freebsd

package git

import (
	"fmt"
	"syscall"
)

func directoryStamp(s *syscall.Stat_t) string {
	return fmt.Sprintf("birth:%d:%d", s.Birthtimespec.Sec, s.Birthtimespec.Nsec)
}
