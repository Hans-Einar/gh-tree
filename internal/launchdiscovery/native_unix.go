//go:build linux || darwin || freebsd

package launchdiscovery

import (
	"encoding/binary"
	"fmt"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strings"
)

func nativeComponent(string) bool   { return true }
func nativePermission(e error) bool { return e == unix.EACCES || e == unix.EPERM || os.IsPermission(e) }
func nativeRedirect(e error) bool   { return e == unix.ELOOP || e == unix.ENOTDIR }
func nativeMissing(e error) bool    { return e == unix.ENOENT || os.IsNotExist(e) }
func nativeRoot(path string) (*os.File, []*os.File, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, nil, errRedirect
	}
	fd, e := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if e != nil {
		return nil, nil, e
	}
	current := os.NewFile(uintptr(fd), "/")
	chain := []*os.File{}
	for _, name := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if name == "" {
			continue
		}
		next, e := nativeChild(current, name, true)
		if e != nil {
			current.Close()
			for _, f := range chain {
				f.Close()
			}
			return nil, nil, e
		}
		chain = append(chain, current)
		current = next
	}
	return current, chain, nil
}
func nativeChild(parent *os.File, name string, dir bool) (*os.File, error) {
	if !component(name) {
		return nil, errRedirect
	}
	flags := unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_CLOEXEC | unix.O_NONBLOCK
	if dir {
		flags |= unix.O_DIRECTORY
	}
	fd, e := unix.Openat(int(parent.Fd()), name, flags, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(fd), name)
	var st unix.Stat_t
	if e = unix.Fstat(fd, &st); e != nil {
		f.Close()
		return nil, e
	}
	kind := st.Mode & unix.S_IFMT
	if dir && kind != unix.S_IFDIR || !dir && kind != unix.S_IFREG {
		f.Close()
		return nil, errRedirect
	}
	return f, nil
}
func observeIdentity(f *os.File, profile string) (api.DirectoryIdentity, error) {
	var st unix.Stat_t
	if e := unix.Fstat(int(f.Fd()), &st); e != nil {
		return api.DirectoryIdentity{}, e
	}
	stamp, e := unixStamp(int(f.Fd()), &st, profile)
	if e != nil {
		return api.DirectoryIdentity{}, e
	}
	var id [16]byte
	binary.LittleEndian.PutUint64(id[:8], uint64(st.Ino))
	return api.NewDirectoryIdentity(api.DirectoryUnix, uint64(st.Dev), id, stamp)
}
func stamp(kind string, sec, nsec int64) string { return fmt.Sprintf("%s:%d:%d", kind, sec, nsec) }
