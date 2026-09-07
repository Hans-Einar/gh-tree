//go:build linux || darwin || freebsd

package persistence

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

type nativeChain = unixChain
type nativeObject = unixObject

func nativeResolveExplicit(path string) (string, error) { return filepath.EvalSymlinks(path) }

func nativeUnsupported(err error) bool {
	return errors.Is(err, errUnsupportedProfile) || errors.Is(err, unix.ELOOP)
}

func nativeAcquire(ctx context.Context, path string) (*nativeChain, error) {
	return unixAcquire(ctx, path)
}
func nativeDirectoryIdentity(o *nativeObject) (api.DirectoryIdentity, error) {
	v, err := unixObserve(o.fd())
	if err != nil {
		return api.DirectoryIdentity{}, err
	}
	return v.directoryIdentity()
}
func nativeMatchesDirectory(o *nativeObject, want api.DirectoryIdentity) bool {
	v, err := unixObserve(o.fd())
	if err != nil {
		return false
	}
	// Match the supplied observation profile, never silently upgrade a supplied
	// change stamp to an available birth stamp or ignore its observed drift.
	if strings.HasPrefix(want.Stamp(), "change:") {
		v.stamp = fmt.Sprintf("change:%d:%d", v.stat.Ctim.Sec, v.stat.Ctim.Nsec)
	}
	got, err := v.directoryIdentity()
	return err == nil && got == want
}
func nativeSameObject(a, b *nativeObject) bool { return a.observation.sameObject(b.observation) }
func nativeSameName(a, b string) bool          { return a == b }
func nativeRead(ctx context.Context, o *nativeObject) ([]byte, error) {
	raw, _, err := unixRead(ctx, o)
	return raw, err
}
func nativeOpenDocument(parent *nativeObject, name string) (*nativeObject, error) {
	return unixOpenDocument(parent, name)
}
func nativeRevalidate(ctx context.Context, c *nativeChain) error { return c.revalidate(ctx) }
func nativeAppendDirectory(c *nativeChain, name string) error {
	if len(c.remaining) != 0 {
		return errors.New("cannot append below missing scope")
	}
	child, err := unixOpenDirectory(c.parent().fd(), name)
	if errors.Is(err, os.ErrNotExist) {
		c.remaining = []string{name}
		return nil
	}
	if err != nil {
		return err
	}
	c.guards = append(c.guards, child)
	c.names = append(c.names, name)
	c.fileSystem, err = unixLocalFileSystem(child.fd())
	return err
}
