//go:build linux || darwin || freebsd

package git

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

func (d *nativeDirectory) openScopeLock(name string) (*os.File, bool, error) {
	if err := nativeComponent(name); err != nil {
		return nil, false, err
	}
	flags := unix.O_RDWR | unix.O_NOFOLLOW | unix.O_CLOEXEC | unix.O_NONBLOCK
	fd, err := unix.Openat(int(d.file.Fd()), name, flags|unix.O_CREAT|unix.O_EXCL, 0600)
	created := err == nil
	if errors.Is(err, unix.EEXIST) {
		fd, err = unix.Openat(int(d.file.Fd()), name, flags, 0)
	}
	if err != nil {
		return nil, false, err
	}
	f := os.NewFile(uintptr(fd), filepath.Join(d.expected.path, name))
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, false, diagnostic(api.Unsupported, "UnsupportedCommonGuard", "The common mutation guard is not an ordinary file.")
	}
	return f, created, nil
}

func lockScopeFile(f *os.File) error { return unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB) }

func scopeLockContention(err error) bool {
	return errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)
}
