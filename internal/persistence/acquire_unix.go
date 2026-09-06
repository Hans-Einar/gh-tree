//go:build linux || darwin || freebsd

package persistence

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

type unixObservation struct {
	stat  unix.Stat_t
	stamp string
}

func unixObserve(fd int) (unixObservation, error) {
	var v unixObservation
	if err := unix.Fstat(fd, &v.stat); err != nil {
		return v, err
	}
	if v.stat.Ino == 0 {
		return v, errors.New("native inode unavailable")
	}
	v.stamp = unixBirthStamp(fd, &v.stat)
	return v, nil
}
func (o unixObservation) directoryIdentity() (api.DirectoryIdentity, error) {
	if uint32(o.stat.Mode)&unix.S_IFMT != unix.S_IFDIR {
		return api.DirectoryIdentity{}, errors.New("not a directory")
	}
	var fileID [16]byte
	binary.LittleEndian.PutUint64(fileID[:8], uint64(o.stat.Ino))
	return api.NewDirectoryIdentity(api.DirectoryUnix, uint64(o.stat.Dev), fileID, o.stamp)
}
func (o unixObservation) sameObject(v unixObservation) bool {
	return o.stat.Dev == v.stat.Dev && o.stat.Ino == v.stat.Ino && o.stamp == v.stamp
}
func (o unixObservation) sameRead(v unixObservation) bool {
	return o.sameObject(v) && o.stat.Size == v.stat.Size && o.stat.Mtim == v.stat.Mtim && o.stat.Ctim == v.stat.Ctim && o.stat.Mode == v.stat.Mode && o.stat.Uid == v.stat.Uid && o.stat.Gid == v.stat.Gid
}

type unixObject struct {
	file        *os.File
	observation unixObservation
}

func (o *unixObject) fd() int      { return int(o.file.Fd()) }
func (o *unixObject) close() error { return o.file.Close() }
func unixOpen(parent int, name string, flags int, mode uint32, directory bool) (*unixObject, error) {
	if parent != unix.AT_FDCWD && !singleName(name) {
		return nil, errors.New("invalid native basename")
	}
	flags |= unix.O_NOFOLLOW | unix.O_CLOEXEC | unix.O_NONBLOCK
	if directory {
		flags |= unix.O_DIRECTORY
	}
	fd, err := unix.Openat(parent, name, flags, mode)
	if err != nil {
		return nil, err
	}
	observation, err := unixObserve(fd)
	if err == nil && uint32(observation.stat.Mode)&unix.S_IFMT != unix.S_IFREG && !directory {
		err = fmt.Errorf("%w: nonregular object", errUnsupportedProfile)
	}
	if err != nil {
		return nil, errors.Join(err, unix.Close(fd))
	}
	file := os.NewFile(uintptr(fd), name)
	if file == nil {
		return nil, errors.Join(errors.New("cannot own native file"), unix.Close(fd))
	}
	return &unixObject{file, observation}, nil
}
func unixOpenDirectory(parent int, name string) (*unixObject, error) {
	return unixOpen(parent, name, unix.O_RDONLY, 0, true)
}
func unixOpenDocument(parent *unixObject, name string) (*unixObject, error) {
	return unixOpen(parent.fd(), name, unix.O_RDONLY, 0, false)
}

type unixChain struct {
	guards     []*unixObject
	names      []string
	remaining  []string
	fileSystem string
}

func (c *unixChain) parent() *unixObject { return c.guards[len(c.guards)-1] }
func (c *unixChain) close() error {
	var err error
	for i := len(c.guards) - 1; i >= 0; i-- {
		err = errors.Join(err, c.guards[i].close())
	}
	c.guards = nil
	return err
}
func unixAcquire(ctx context.Context, path string) (_ *unixChain, resultErr error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(path) {
		return nil, errors.New("absolute physical directory required")
	}
	parts := strings.Split(path[1:], "/")
	if len(parts) == 1 && parts[0] == "" {
		parts = nil
	}
	for _, part := range parts {
		if !singleName(part) {
			return nil, errors.New("invalid directory component")
		}
	}
	root, err := unixOpenDirectory(unix.AT_FDCWD, "/")
	if err != nil {
		return nil, err
	}
	c := &unixChain{guards: []*unixObject{root}}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, c.close())
		}
	}()
	c.fileSystem, err = unixLocalFileSystem(root.fd())
	if err != nil {
		return nil, err
	}
	for i, part := range parts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		child, err := unixOpenDirectory(c.parent().fd(), part)
		if errors.Is(err, os.ErrNotExist) {
			c.remaining = append([]string{}, parts[i:]...)
			return c, nil
		}
		if err != nil {
			return nil, err
		}
		c.guards = append(c.guards, child)
		c.names = append(c.names, part)
		c.fileSystem, err = unixLocalFileSystem(child.fd())
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Movement of an acquired directory does not change descriptor authority. An
// observed entry substitution refuses; no path reopen can retarget the handle.
func (c *unixChain) revalidate(ctx context.Context) error {
	for i, name := range c.names {
		if err := ctx.Err(); err != nil {
			return err
		}
		current, err := unixOpenDirectory(c.guards[i].fd(), name)
		if err != nil {
			return err
		}
		same := current.observation.sameObject(c.guards[i+1].observation)
		closeErr := current.close()
		if !same {
			return errors.Join(errors.New("acquired directory entry changed"), closeErr)
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
func unixRead(ctx context.Context, object *unixObject) ([]byte, unixObservation, error) {
	read := func() ([]byte, unixObservation, error) {
		before, err := unixObserve(object.fd())
		if err != nil {
			return nil, before, err
		}
		if uint32(before.stat.Mode)&unix.S_IFMT != unix.S_IFREG || before.stat.Size < 0 {
			return nil, before, errUnsupportedProfile
		}
		if before.stat.Size > api.MaxDocumentBytes {
			return nil, before, corrupt("document exceeds 4 MiB")
		}
		if _, err := object.file.Seek(0, io.SeekStart); err != nil {
			return nil, before, err
		}
		data := make([]byte, 0, int(before.stat.Size))
		buffer := make([]byte, 32*1024)
		for {
			if err := ctx.Err(); err != nil {
				return nil, before, err
			}
			n, err := object.file.Read(buffer)
			if n > 0 {
				if len(data)+n > api.MaxDocumentBytes {
					return nil, before, corrupt("document exceeds 4 MiB")
				}
				data = append(data, buffer[:n]...)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, before, err
			}
			if n == 0 {
				return nil, before, io.ErrNoProgress
			}
		}
		after, err := unixObserve(object.fd())
		if err != nil {
			return nil, before, err
		}
		if !before.sameRead(after) || int64(len(data)) != after.stat.Size {
			return nil, after, errors.New("document changed during read")
		}
		return data, after, nil
	}
	first, observation, err := read()
	if err != nil {
		return nil, observation, err
	}
	second, final, err := read()
	if err != nil {
		return nil, final, err
	}
	if !observation.sameRead(final) || !bytes.Equal(first, second) {
		return nil, final, errors.New("document changed between bounded reads")
	}
	return first, final, nil
}

type unixLockKey struct{ device, inode uint64 }
type unixLockEntry struct {
	gate  chan struct{}
	users int
}

var unixProcessLocks = struct {
	sync.Mutex
	entries map[unixLockKey]*unixLockEntry
}{entries: map[unixLockKey]*unixLockEntry{}}

func referenceUnixLock(key unixLockKey) *unixLockEntry {
	unixProcessLocks.Lock()
	defer unixProcessLocks.Unlock()
	entry := unixProcessLocks.entries[key]
	if entry == nil {
		entry = &unixLockEntry{gate: make(chan struct{}, 1)}
		entry.gate <- struct{}{}
		unixProcessLocks.entries[key] = entry
	}
	entry.users++
	return entry
}
func releaseUnixLock(key unixLockKey) {
	unixProcessLocks.Lock()
	defer unixProcessLocks.Unlock()
	entry := unixProcessLocks.entries[key]
	entry.users--
	if entry.users == 0 {
		delete(unixProcessLocks.entries, key)
	}
}

type unixStoreLock struct {
	object *unixObject
	key    unixLockKey
	entry  *unixLockEntry
}

func unixLock(ctx context.Context, parent *unixObject, basename string, budget time.Duration) (_ *unixStoreLock, resultErr error) {
	return unixLockMode(ctx, parent, basename, budget, true)
}
func unixLockMode(ctx context.Context, parent *unixObject, basename string, budget time.Duration, create bool) (_ *unixStoreLock, resultErr error) {
	if !singleName(basename) || budget <= 0 || budget > 5*time.Second {
		return nil, errors.New("invalid lock parameters")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(budget)
	flags := unix.O_RDWR
	if create {
		flags |= unix.O_CREAT
	}
	object, err := unixOpen(parent.fd(), basename+".lock", flags, 0600, false)
	if err != nil {
		return nil, err
	}
	var key unixLockKey
	var entry *unixLockEntry
	localOwned, kernelOwned := false, false
	defer func() {
		if resultErr != nil {
			if kernelOwned {
				resultErr = errors.Join(resultErr, unix.Flock(object.fd(), unix.LOCK_UN))
			}
			resultErr = errors.Join(resultErr, object.close())
			if entry != nil {
				if localOwned {
					entry.gate <- struct{}{}
				}
				releaseUnixLock(key)
			}
		}
	}()
	if object.observation.stat.Nlink != 1 {
		return nil, errors.New("unsupported linked lock object")
	}
	key = unixLockKey{uint64(object.observation.stat.Dev), uint64(object.observation.stat.Ino)}
	entry = referenceUnixLock(key)
	left := time.Until(deadline)
	if left <= 0 {
		return nil, errLockBusy
	}
	timer := time.NewTimer(left)
	select {
	case <-ctx.Done():
		timer.Stop()
		return nil, ctx.Err()
	case <-timer.C:
		return nil, errLockBusy
	case <-entry.gate:
		timer.Stop()
		localOwned = true
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if time.Until(deadline) <= 0 {
			return nil, fmt.Errorf("%w after %s", errLockBusy, budget)
		}
		err := unix.Flock(object.fd(), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			kernelOwned = true
			current, checkErr := unixOpen(parent.fd(), basename+".lock", unix.O_RDWR, 0, false)
			if checkErr != nil {
				return nil, checkErr
			}
			same := object.observation.sameObject(current.observation) && current.observation.stat.Nlink == 1
			closeErr := current.close()
			if !same || closeErr != nil {
				return nil, errors.Join(errors.New("stable lock entry changed"), closeErr)
			}
			return &unixStoreLock{object, key, entry}, nil
		}
		if err != unix.EWOULDBLOCK && err != unix.EAGAIN {
			return nil, err
		}
		left := time.Until(deadline)
		if left > 10*time.Millisecond {
			left = 10 * time.Millisecond
		}
		if left <= 0 {
			return nil, errLockBusy
		}
		timer := time.NewTimer(left)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
func (l *unixStoreLock) close() error {
	err := errors.Join(unix.Flock(l.object.fd(), unix.LOCK_UN), l.object.close())
	l.entry.gate <- struct{}{}
	releaseUnixLock(l.key)
	return err
}
