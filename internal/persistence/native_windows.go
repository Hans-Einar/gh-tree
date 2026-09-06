package persistence

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

type nativeChain = winChain
type nativeObject = winObject

func nativeResolveExplicit(path string) (resolved string, resultErr error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	// This one-time explicit user-scope selection deliberately follows its
	// chosen links/junctions. Every later binding/request uses no-reparse opens.
	h, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES, winShareAll, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	defer func() { resultErr = errors.Join(resultErr, windows.CloseHandle(h)) }()
	buf := make([]uint16, 32768)
	n, err := windows.GetFinalPathNameByHandle(h, &buf[0], uint32(len(buf)), 0)
	if err != nil {
		return "", err
	}
	if n == 0 || n >= uint32(len(buf)) {
		return "", errors.New("explicit physical locator length limit")
	}
	resolved = windows.UTF16ToString(buf[:n])
	resolved = strings.TrimPrefix(resolved, `\\?\`)
	return resolved, nil
}

func nativeUnsupported(err error) bool {
	return errors.Is(err, errUnsupportedProfile) || errors.Is(err, windows.ERROR_CANT_ACCESS_FILE) || errors.Is(err, windows.ERROR_CANT_RESOLVE_FILENAME) || errors.Is(err, windows.STATUS_REPARSE_POINT_NOT_RESOLVED)
}

func nativeAcquire(ctx context.Context, path string) (*nativeChain, error) {
	return winAcquire(ctx, path)
}
func nativeDirectoryIdentity(o *nativeObject) (api.DirectoryIdentity, error) {
	v, err := winObserve(o.handle())
	if err != nil {
		return api.DirectoryIdentity{}, err
	}
	return v.directoryIdentity()
}
func nativeMatchesDirectory(o *nativeObject, want api.DirectoryIdentity) bool {
	got, err := nativeDirectoryIdentity(o)
	return err == nil && got == want
}
func nativeSameObject(a, b *nativeObject) bool { return a.observation.sameObject(b.observation) }
func nativeSameName(a, b string) bool          { return strings.EqualFold(a, b) }
func nativeRead(ctx context.Context, o *nativeObject) ([]byte, error) {
	raw, _, err := winRead(ctx, o)
	return raw, err
}
func nativeOpenDocument(parent *nativeObject, name string) (*nativeObject, error) {
	return winOpenDocument(parent, name)
}
func nativeRevalidate(ctx context.Context, c *nativeChain) error {
	for _, o := range c.guards {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := winObserve(o.handle()); err != nil {
			return err
		}
	}
	return nil
}
func nativeAppendDirectory(c *nativeChain, name string) error {
	if len(c.remaining) != 0 {
		return errors.New("cannot append below missing scope")
	}
	child, err := winOpenDirectory(c.parent().handle(), name)
	if errors.Is(err, os.ErrNotExist) {
		c.remaining = []string{name}
		return nil
	}
	if err != nil {
		return err
	}
	c.guards = append(c.guards, child)
	return nil
}
