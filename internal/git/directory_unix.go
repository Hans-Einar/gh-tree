//go:build linux || darwin || freebsd

package git

import (
	"encoding/binary"
	"fmt"
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

func (d *nativeDirectory) openChild(name string) (*nativeDirectory, error) {
	if err := nativeComponent(name); err != nil {
		return nil, err
	}
	fd, err := unix.Openat(int(d.file.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), filepath.Join(d.expected.path, name))
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		f.Close()
		return nil, diagnostic(api.Unsupported, "DirectoryIdentityUnavailable", "Native child directory identity is unavailable.")
	}
	var file [16]byte
	binary.LittleEndian.PutUint64(file[:8], uint64(stat.Ino))
	identity, err := api.NewDirectoryIdentity(api.DirectoryUnix, uint64(stat.Dev), file, directoryStamp(f, stat))
	if err != nil {
		f.Close()
		return nil, err
	}
	return &nativeDirectory{file: f, expected: directoryObservation{path: filepath.Join(d.expected.path, name), identity: identity}}, nil
}

func regularIdentity(f *os.File) (string, uint32, error) {
	info, err := f.Stat()
	if err != nil {
		return "", 0, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() {
		return "", 0, diagnostic(api.Unsupported, "FileIdentityUnavailable", "Native regular file identity is unavailable.")
	}
	return fmt.Sprintf("%d:%d:%s", uint64(stat.Dev), uint64(stat.Ino), directoryStamp(f, stat)), uint32(info.Mode().Perm()), nil
}

func (d *nativeDirectory) linkObservation(name string) (string, string, uint32, error) {
	if err := nativeComponent(name); err != nil {
		return "", "", 0, err
	}
	var before, after unix.Stat_t
	if err := unix.Fstatat(int(d.file.Fd()), name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return "", "", 0, err
	}
	if before.Mode&unix.S_IFMT != unix.S_IFLNK {
		return "", "", 0, diagnostic(api.Unsupported, "UnsupportedFileKind", "The observed child is not a supported regular file or symbolic link.")
	}
	buffer := make([]byte, 32769)
	n, err := unix.Readlinkat(int(d.file.Fd()), name, buffer)
	if err != nil {
		return "", "", 0, err
	}
	if n >= len(buffer) {
		return "", "", 0, diagnostic(api.Unavailable, "LinkTargetLimit", "The symbolic-link target exceeds its observation bound.")
	}
	if err := unix.Fstatat(int(d.file.Fd()), name, &after, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return "", "", 0, err
	}
	if before.Dev != after.Dev || before.Ino != after.Ino || before.Mode != after.Mode || before.Size != after.Size {
		return "", "", 0, diagnostic(api.StaleObservation, "LinkChanged", "The symbolic-link object changed while being observed.")
	}
	return string(buffer[:n]), fmt.Sprintf("%d:%d", uint64(before.Dev), uint64(before.Ino)), uint32(before.Mode & 07777), nil
}
