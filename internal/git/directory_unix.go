//go:build linux || darwin || freebsd

package git

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

// nativeDirectory is a private acquired object, not an API scope or pathname
// lease. The caller must close it before returning a preparation to the user.
type nativeDirectory struct {
	file     *os.File
	expected directoryObservation
}

func acquireDirectory(expected directoryObservation) (*nativeDirectory, error) {
	fd, err := unix.Open(expected.path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	d := &nativeDirectory{file: os.NewFile(uintptr(fd), expected.path), expected: expected}
	if err = d.validate(); err != nil {
		d.close()
		return nil, err
	}
	return d, nil
}

func (d *nativeDirectory) validate() error {
	info, err := d.file.Stat()
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() {
		return diagnostic(api.Unsupported, "DirectoryIdentityUnavailable", "The acquired native object is not a supported directory.")
	}
	var id [16]byte
	binary.LittleEndian.PutUint64(id[:8], uint64(stat.Ino))
	actual, err := api.NewDirectoryIdentity(api.DirectoryUnix, uint64(stat.Dev), id, directoryStamp(d.file, stat))
	if err != nil {
		return err
	}
	if actual != d.expected.identity {
		return diagnostic(api.StaleObservation, "DirectoryIdentityChanged", "The acquired directory differs from the confirmed physical observation.")
	}
	return nil
}

func (d *nativeDirectory) close() error { return d.file.Close() }

func (d *nativeDirectory) openRegular(name string) (*os.File, error) {
	if err := nativeComponent(name); err != nil {
		return nil, err
	}
	fd, err := unix.Openat(int(d.file.Fd()), name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), filepath.Join(d.expected.path, name))
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, diagnostic(api.Unsupported, "NonRegularNativeFile", "The native child is not a supported regular file.")
	}
	return f, nil
}

func (d *nativeDirectory) createPrivate(name string) (*os.File, error) {
	if err := nativeComponent(name); err != nil {
		return nil, err
	}
	fd, err := unix.Openat(int(d.file.Fd()), name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), filepath.Join(d.expected.path, name)), nil
}
