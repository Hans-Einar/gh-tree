package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// storeBinding stores equality evidence and a construction-resolved locator,
// never an OS capability. Every request reacquires the native objects. The
// original nearest-existing anchor is kept even when missing descendants appear.
type storeBinding struct {
	family               api.StorageFamily
	parentPath, basename string
	anchor               api.DirectoryIdentity
	anchorIndex          int
	remaining            []string
	observed             *bindingObservation
}

type bindingObservation struct {
	sync.Mutex
	anchor api.DirectoryIdentity
	index  int
}

func resolveExplicitFile(path string) (string, error) {
	if !utf8.ValidString(path) || strings.IndexByte(path, 0) >= 0 || !filepath.IsAbs(path) {
		return "", errors.New("explicit storage location must be an absolute UTF-8 path")
	}
	path = filepath.Clean(path)
	if !singleName(filepath.Base(path)) {
		return "", errors.New("invalid storage basename")
	}
	// Only explicit user-selected scope permits one-time link resolution. The
	// native acquisition that follows freezes the actual resulting parent.
	probe := path
	var suffix []string
	for {
		resolved, err := nativeResolveExplicit(probe)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			return "", err
		}
		suffix = append(suffix, filepath.Base(probe))
		probe = parent
	}
}

func bindExplicit(ctx context.Context, family api.StorageFamily, path string) (b storeBinding, resultErr error) {
	resolved, err := resolveExplicitFile(path)
	if err != nil {
		return b, err
	}
	b.family, b.parentPath, b.basename = family, filepath.Dir(resolved), filepath.Base(resolved)
	c, err := nativeAcquire(ctx, b.parentPath)
	if err != nil {
		return b, err
	}
	defer func() { resultErr = errors.Join(resultErr, c.close()) }()
	b.anchor, err = nativeDirectoryIdentity(c.parent())
	if err != nil {
		return b, err
	}
	b.anchorIndex = len(c.guards) - 1
	b.remaining = append([]string(nil), c.remaining...)
	b.observed = &bindingObservation{anchor: b.anchor, index: b.anchorIndex}
	return b, nil
}

func (b storeBinding) acquire(ctx context.Context) (c *nativeChain, resultErr error) {
	if b.observed == nil {
		return nil, errors.Join(errInvalidRequest, errors.New("uninitialized storage binding"))
	}
	c, err := nativeAcquire(ctx, b.parentPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, c.close())
			c = nil
		}
	}()
	b.observed.Lock()
	defer b.observed.Unlock()
	if len(c.guards) <= b.observed.index || !nativeMatchesDirectory(c.guards[b.observed.index], b.observed.anchor) {
		return c, errors.Join(errBindingChanged, errors.New("bound storage anchor changed"))
	}
	if err := nativeRevalidate(ctx, c); err != nil {
		return c, err
	}
	if len(c.guards)-1 > b.observed.index {
		identity, err := nativeDirectoryIdentity(c.parent())
		if err != nil {
			return c, err
		}
		b.observed.anchor, b.observed.index = identity, len(c.guards)-1
	}
	return c, nil
}

func acquireRun(ctx context.Context, scope api.WorktreeScope) (c *nativeChain, resultErr error) {
	if !scope.Valid() {
		return nil, errors.Join(errInvalidRequest, errors.New("invalid run scope"))
	}
	v := scope.Data()
	c, err := nativeAcquire(ctx, v.RootLocator)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, c.close())
			c = nil
		}
	}()
	if len(c.remaining) != 0 || !nativeMatchesDirectory(c.parent(), v.RootIdentity) {
		return c, errors.Join(errBindingChanged, errors.New("run root identity changed or unavailable"))
	}
	if err := nativeAppendDirectory(c, ".gh-tree"); err != nil {
		return c, err
	}
	if err := nativeRevalidate(ctx, c); err != nil {
		return c, err
	}
	return c, nil
}

func acquiredToken(c *nativeChain, basename string) (string, error) {
	id, err := nativeDirectoryIdentity(c.parent())
	if err != nil {
		return "", err
	}
	return bindingToken(id, c.remaining, basename)
}

func bindingsOverlap(ctx context.Context, a, b storeBinding) (overlap bool, resultErr error) {
	x, err := a.acquire(ctx)
	if err != nil {
		return false, err
	}
	defer func() { resultErr = errors.Join(resultErr, x.close()) }()
	y, err := b.acquire(ctx)
	if err != nil {
		return false, err
	}
	defer func() { resultErr = errors.Join(resultErr, y.close()) }()
	return acquiredOverlap(x, a.basename, y, b.basename)
}

func acquiredOverlap(x *nativeChain, xname string, y *nativeChain, yname string) (overlap bool, resultErr error) {
	namesOverlap := nativeSameName(xname, yname) || nativeSameName(xname+".lock", yname) || nativeSameName(xname, yname+".lock")
	if nativeSameObject(x.parent(), y.parent()) && len(x.remaining) == len(y.remaining) && namesOverlap {
		same := true
		for i := range x.remaining {
			if !nativeSameName(x.remaining[i], y.remaining[i]) {
				same = false
			}
		}
		if same {
			return true, nil
		}
	}
	if len(x.remaining) != 0 || len(y.remaining) != 0 {
		return false, nil
	}
	xfile, err := nativeOpenDocument(x.parent(), xname)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer func() { resultErr = errors.Join(resultErr, xfile.close()) }()
	yfile, err := nativeOpenDocument(y.parent(), yname)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer func() { resultErr = errors.Join(resultErr, yfile.close()) }()
	return nativeSameObject(xfile, yfile), nil
}
