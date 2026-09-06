//go:build linux

package git

import (
	"fmt"
	"syscall"
)

func directoryStamp(s *syscall.Stat_t) string {
	return fmt.Sprintf("change:%d:%d", s.Ctim.Sec, s.Ctim.Nsec)
}
