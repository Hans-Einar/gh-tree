package launchdiscovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var errRedirect = errors.New("redirected or unsupported filesystem object")
var errChanged = errors.New("filesystem observation changed")
var errLimit = errors.New("observation limit reached")
var errInvalid = errors.New("invalid passive source")

type directory struct {
	file     *os.File
	identity api.DirectoryIdentity
	chain    []*os.File
	path     string
}

func (d *directory) close() {
	if d == nil {
		return
	}
	_ = d.file.Close()
	for i := len(d.chain) - 1; i >= 0; i-- {
		_ = d.chain[i].Close()
	}
}
func component(s string) bool {
	return s != "" && s != "." && s != ".." && utf8.ValidString(s) && !strings.ContainsAny(s, "/\\\x00") && nativeComponent(s)
}
func projectParts(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	if filepath.IsAbs(s) || filepath.VolumeName(s) != "" {
		return nil, errRedirect
	}
	s = filepath.ToSlash(s)
	p := strings.Split(s, "/")
	for _, v := range p {
		if !component(v) {
			return nil, errRedirect
		}
	}
	return p, nil
}

func acquireRoot(scope api.WorktreeScope) (*directory, error) {
	d := scope.Data()
	f, chain, err := nativeRoot(d.RootLocator)
	if err != nil {
		return nil, err
	}
	dir := &directory{file: f, chain: chain, path: d.RootLocator}
	dir.identity, err = observeIdentity(f, d.RootIdentity.Stamp())
	if err != nil || dir.identity != d.RootIdentity {
		dir.close()
		if err == nil {
			err = errChanged
		}
		return nil, err
	}
	return dir, nil
}
func childDirectory(parent *directory, name string) (*directory, error) {
	if !component(name) {
		return nil, errRedirect
	}
	f, err := nativeChild(parent.file, name, true)
	if err != nil {
		return nil, err
	}
	id, err := observeIdentity(f, "")
	if err != nil {
		f.Close()
		return nil, err
	}
	return &directory{file: f, identity: id, path: filepath.Join(parent.path, name)}, nil
}
func sameNamedDirectory(parent *directory, name string, dir *directory) error {
	f, e := nativeChild(parent.file, name, true)
	if e != nil {
		return e
	}
	defer f.Close()
	id, e := observeIdentity(f, dir.identity.Stamp())
	if e != nil {
		return e
	}
	if id != dir.identity {
		return errChanged
	}
	return nil
}
func sameDirectory(dir *directory) error {
	id, e := observeIdentity(dir.file, dir.identity.Stamp())
	if e != nil {
		return e
	}
	if id != dir.identity {
		return errChanged
	}
	return nil
}

type fileObservation struct {
	name, state string
	identity    api.DirectoryIdentity
	data        []byte
	digest      [32]byte
}

func (f fileObservation) versionBytes() string {
	return fmt.Sprintf("%s:%s:%v:%x", f.name, f.state, f.identity, f.digest)
}

func observeFile(ctx context.Context, dir *directory, name string, read bool, cap int) (fileObservation, error) {
	o := fileObservation{name: name, state: "absent"}
	if err := ctx.Err(); err != nil {
		return o, err
	}
	f, err := nativeChild(dir.file, name, false)
	if err != nil {
		if nativeMissing(err) {
			return o, nil
		}
		o.state = "unavailable"
		return o, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return o, err
	}
	if !info.Mode().IsRegular() {
		o.state = "nonregular"
		return o, errRedirect
	}
	o.identity, err = observeIdentity(f, "")
	if err != nil {
		return o, err
	}
	o.state = "regular"
	if read {
		if info.Size() > int64(cap) {
			return o, errLimit
		}
		buf := make([]byte, 32768)
		for {
			if err := ctx.Err(); err != nil {
				return o, err
			}
			n, e := f.Read(buf)
			if len(o.data)+n > cap {
				return o, errLimit
			}
			o.data = append(o.data, buf[:n]...)
			if e == io.EOF {
				break
			}
			if e != nil {
				return o, e
			}
		}
		o.digest = sha256.Sum256(o.data)
	}
	after, err := f.Stat()
	if err != nil {
		return o, err
	}
	id, err := observeIdentity(f, o.identity.Stamp())
	if err != nil {
		return o, err
	}
	if id != o.identity || info.Size() != after.Size() || info.ModTime() != after.ModTime() {
		return o, errChanged
	}
	// Reopen by retained parent after reading, refusing pathname replacement.
	current, e := nativeChild(dir.file, name, false)
	if e != nil {
		return o, e
	}
	defer current.Close()
	id, e = observeIdentity(current, o.identity.Stamp())
	if e != nil {
		return o, e
	}
	if id != o.identity {
		return o, errChanged
	}
	if read {
		// A second bounded read detects same-size in-place writes even when an
		// editor restores mtime. It cannot promise immutable future project code.
		h := sha256.New()
		buf := make([]byte, 32768)
		total := 0
		for {
			if e := ctx.Err(); e != nil {
				return o, e
			}
			n, e := current.Read(buf)
			total += n
			if total > cap {
				return o, errLimit
			}
			_, _ = h.Write(buf[:n])
			if e == io.EOF {
				break
			}
			if e != nil {
				return o, e
			}
		}
		if total != len(o.data) || !bytes.Equal(h.Sum(nil), o.digest[:]) {
			return o, errChanged
		}
	}
	return o, nil
}
