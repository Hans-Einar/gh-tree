//go:build linux || darwin || freebsd

package persistence

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

type nativeMetadata = unixMetadata
type nativeStoreLock = unixStoreLock

func nativeLock(ctx context.Context, parent *nativeObject, basename string, wait time.Duration) (*nativeStoreLock, error) {
	return unixLock(ctx, parent, basename, wait)
}
func nativeRetainOriginal(original, parent *nativeObject, target, name string) (*nativeObject, error) {
	return unixRetainOriginal(original, parent, target, name)
}

func nativeArtifactIdentity(object *nativeObject) (diskIdentity, error) {
	v, err := unixObserve(object.fd())
	if err != nil {
		return diskIdentity{}, err
	}
	var id [16]byte
	binary.LittleEndian.PutUint64(id[:8], uint64(v.stat.Ino))
	return diskIdentity{api.DirectoryUnix, uint64(v.stat.Dev), id, v.stamp}, nil
}

func nativeNameKey(name string) string { return name }
func nativeObjectSize(object *nativeObject) (int64, error) {
	v, err := unixObserve(object.fd())
	if err != nil {
		return 0, err
	}
	if uint32(v.stat.Mode)&unix.S_IFMT != unix.S_IFREG || v.stat.Size < 0 {
		return 0, errUnsupportedProfile
	}
	return v.stat.Size, nil
}
func nativeCreateFile(parent *nativeObject, name string, userOnly bool) (*nativeObject, error) {
	mode := uint32(0666)
	if userOnly {
		mode = 0600
	}
	return unixOpen(parent.fd(), name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL, mode, false)
}
func nativeInspectMetadata(object *nativeObject) (nativeMetadata, error) {
	return unixInspectMetadata(object)
}
func nativeApplyMetadata(object *nativeObject, metadata nativeMetadata) error {
	return unixApplyMetadata(object, metadata)
}
func nativeOpenOriginal(parent *nativeObject, name string) (*nativeObject, error) {
	return unixOpenDocument(parent, name)
}

func nativeInspectDirectory(parent *nativeObject) error {
	v, err := unixObserve(parent.fd())
	if err != nil {
		return err
	}
	if uint32(v.stat.Mode)&unix.S_IFMT != unix.S_IFDIR {
		return errUnsupportedProfile
	}
	if _, err := unixLocalFileSystem(parent.fd()); err != nil {
		return err
	}
	if err := unixInspectNativePolicy(parent.fd(), &v.stat); err != nil {
		return err
	}
	first, err := unixReadAttributes(parent.fd())
	if err != nil {
		return err
	}
	second, err := unixReadAttributes(parent.fd())
	if err != nil {
		return err
	}
	after, err := unixObserve(parent.fd())
	if err != nil {
		return err
	}
	if !v.sameRead(after) || !unixAttributesEqual(first, second) {
		return errors.New("directory metadata changed during inspection")
	}
	return nil
}

func nativeCreateDirectory(parent *nativeObject, name string, userOnly bool) (*nativeObject, error) {
	if !singleName(name) {
		return nil, errors.New("invalid directory basename")
	}
	if err := nativeInspectDirectory(parent); err != nil {
		return nil, err
	}
	mode := uint32(0777)
	if userOnly {
		mode = 0700
	}
	if err := unix.Mkdirat(parent.fd(), name, mode); err != nil {
		return nil, err
	}
	child, err := unixOpenDirectory(parent.fd(), name)
	if err != nil {
		return nil, err
	}
	if err := nativeInspectDirectory(child); err != nil {
		return nil, errors.Join(err, child.close())
	}
	return child, nil
}
