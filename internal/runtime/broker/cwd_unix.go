//go:build linux || darwin || freebsd

package broker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

var ErrCwd = errors.New("Runtime cwd observation is stale, redirected or unsupported")

type directoryLink struct {
	parent, child *os.File
	name          string
}

// AcquiredDirectory retains the entire acquired no-follow chain. File is only
// passed to the private child through its designated descriptor; it is never
// turned into an fd pathname or an Application capability.
type AcquiredDirectory struct {
	chain        []*os.File
	links        []directoryLink
	root, target *os.File
	spec         StartSpec
	closed       bool
	closeErr     error
}

func (a *AcquiredDirectory) File() *os.File {
	if a.closed {
		return nil
	}
	return a.target
}

func (a *AcquiredDirectory) Close() error {
	if a.closed {
		return a.closeErr
	}
	a.closed = true
	for i := len(a.chain) - 1; i >= 0; i-- {
		a.closeErr = errors.Join(a.closeErr, a.chain[i].Close())
	}
	a.chain = nil
	a.links = nil
	a.root = nil
	a.target = nil
	return a.closeErr
}

func AcquireCwd(spec StartSpec) (*AcquiredDirectory, error) { return acquireCwd(spec, nil) }

// The private test seam runs at exact acquisition barriers, never on a live
// registry lock. It is not exposed through Runtime's product/API surface.
func acquireCwd(spec StartSpec, barrier func(string)) (*AcquiredDirectory, error) {
	if !spec.valid() || spec.RootIdentity.Platform() != api.DirectoryUnix || !filepath.IsAbs(spec.RootLocator) || filepath.Clean(spec.RootLocator) != spec.RootLocator {
		return nil, ErrCwd
	}
	spec.Components = append([]string(nil), spec.Components...)
	spec.Arguments = append([]string(nil), spec.Arguments...)
	spec.Environment = append([]string(nil), spec.Environment...)
	a := &AcquiredDirectory{spec: spec}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return a, err
	}
	current := os.NewFile(uintptr(fd), "/")
	a.chain = append(a.chain, current)
	rootParts := strings.Split(strings.TrimPrefix(spec.RootLocator, "/"), "/")
	for _, name := range rootParts {
		if name == "" {
			continue
		}
		next, e := a.openChild(current, name)
		if e != nil {
			return a, e
		}
		current = next
	}
	a.root = current
	if barrier != nil {
		barrier("root-acquired")
	}
	if err = a.matches(a.root, spec.RootIdentity); err != nil {
		return a, err
	}
	for _, name := range spec.Components {
		next, e := a.openChild(current, name)
		if e != nil {
			return a, e
		}
		current = next
	}
	a.target = current
	if barrier != nil {
		barrier("project-acquired")
	}
	if err = a.Revalidate(); err != nil {
		return a, err
	}
	return a, nil
}

func (a *AcquiredDirectory) openChild(parent *os.File, name string) (*os.File, error) {
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	child := os.NewFile(uintptr(fd), name)
	a.chain = append(a.chain, child)
	a.links = append(a.links, directoryLink{parent, child, name})
	return child, nil
}

func (a *AcquiredDirectory) matches(file *os.File, want api.DirectoryIdentity) error {
	actual, err := ObserveDirectory(file, want.Stamp())
	if err != nil {
		return err
	}
	if !actual.Equal(want) {
		return ErrCwd
	}
	return nil
}

func (a *AcquiredDirectory) Revalidate() error {
	if a.closed || a.target == nil || a.root == nil {
		return ErrCwd
	}
	// Compare retained child objects to no-follow entries in retained parents.
	// This detects observed relocation; it does not promise continuous ancestry
	// after this observation if another actor moves the same pinned object.
	for _, link := range a.links {
		var named, held unix.Stat_t
		if err := unix.Fstatat(int(link.parent.Fd()), link.name, &named, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		if err := unix.Fstat(int(link.child.Fd()), &held); err != nil {
			return err
		}
		if named.Mode&unix.S_IFMT != unix.S_IFDIR || held.Mode&unix.S_IFMT != unix.S_IFDIR || named.Dev != held.Dev || named.Ino != held.Ino {
			return ErrCwd
		}
	}
	if err := a.matches(a.root, a.spec.RootIdentity); err != nil {
		return err
	}
	return a.matches(a.target, a.spec.ProjectIdentity)
}

func ObserveDirectory(file *os.File, profile string) (api.DirectoryIdentity, error) {
	if file == nil {
		return api.DirectoryIdentity{}, ErrCwd
	}
	var st unix.Stat_t
	if err := unix.Fstat(int(file.Fd()), &st); err != nil {
		return api.DirectoryIdentity{}, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return api.DirectoryIdentity{}, ErrCwd
	}
	stamp, err := directoryStamp(int(file.Fd()), &st, profile)
	if err != nil {
		return api.DirectoryIdentity{}, err
	}
	var id [16]byte
	binary.LittleEndian.PutUint64(id[:8], uint64(st.Ino))
	return api.NewDirectoryIdentity(api.DirectoryUnix, uint64(st.Dev), id, stamp)
}

func stamp(kind string, sec, nsec int64) string { return fmt.Sprintf("%s:%d:%d", kind, sec, nsec) }
