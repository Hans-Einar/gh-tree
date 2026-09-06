//go:build linux || darwin || freebsd

package broker

import (
	"golang.org/x/sys/unix"
	"os"
	"time"
)

func inheritedPipe(fd int, write bool) (*os.File, error) {
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return nil, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFIFO {
		return nil, ErrProtocol
	}
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		return nil, err
	}
	if !validNativePipeAccess(fd, flags, write) {
		return nil, ErrProtocol
	}
	// Inherited pipe descriptors require nonblocking setup before os.NewFile
	// to join the Go poller; otherwise read deadlines are unsupported.
	if err := unix.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	unix.CloseOnExec(fd)
	f := os.NewFile(uintptr(fd), "runtime-private-pipe")
	if err := f.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		closeErr := f.Close()
		if closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	return f, nil
}
