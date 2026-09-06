//go:build linux || darwin || freebsd

package launchdiscovery

import (
	"context"
	"encoding/binary"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnixNativeProfileAndMovedObject(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	put(t, root, "project/package.json", `{"scripts":{"original":"x"}}`)
	scope := fixtureScope(t, root)
	p, e := openProject(context.Background(), scope, []string{"project"})
	if e != nil {
		t.Fatal(e)
	}
	defer p.close()
	var st unix.Stat_t
	if e = unix.Fstat(int(p.last().file.Fd()), &st); e != nil {
		t.Fatal(e)
	}
	id := p.last().identity
	fileID := id.FileID()
	if id.Device() != uint64(st.Dev) || binary.LittleEndian.Uint64(fileID[:8]) != uint64(st.Ino) {
		t.Fatal("native identity mismatch")
	}
	if e = os.Rename(root, filepath.Join(base, "moved")); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	put(t, root, "project/package.json", `{"scripts":{"replacement":"x"}}`)
	o, e := observeFile(context.Background(), p.last(), "package.json", true, 4096)
	if e != nil || !strings.Contains(string(o.data), "original") {
		t.Fatal("lost retained object", e, string(o.data))
	}
	if e = p.check(scope); e == nil {
		t.Fatal("observed root relocation accepted")
	}
}
func TestUnixPermissionAndRedirectedLocks(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission refusal requires unprivileged process")
	}
	root := t.TempDir()
	outside := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"valid":"x"}}`)
	put(t, root, "denied/package.json", `{"scripts":{"hidden":"x"}}`)
	put(t, outside, "lock", "lock")
	if e := os.Symlink(filepath.Join(outside, "lock"), filepath.Join(root, "pnpm-lock.yaml")); e != nil {
		t.Fatal(e)
	}
	if e := os.Chmod(filepath.Join(root, "denied"), 0000); e != nil {
		t.Fatal(e)
	}
	defer os.Chmod(filepath.Join(root, "denied"), 0700)
	a := must(New(Config{}))
	s := fixtureScope(t, root)
	r := discover(t, a, s)
	if len(r.Definitions) != 1 || r.Observation.Data().Completeness != api.Partial {
		t.Fatal(r)
	}
	def := pick(t, r.Definitions, api.Npm, "", "valid")
	if def.Data().EffectiveExecutable.Data().Executable != "npm" {
		t.Fatal("link lock qualified")
	}
	found := false
	for _, diag := range r.Diagnostics {
		if diag.Data().Code == api.Permission {
			found = true
		}
	}
	if !found {
		t.Fatal("permission collapsed", r.Diagnostics)
	}
}
func TestUnixSuppliedChangeProfileDoesNotUpgrade(t *testing.T) {
	root := t.TempDir()
	scope := fixtureScope(t, root)
	f, chain, e := nativeRoot(root)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	for _, v := range chain {
		defer v.Close()
	}
	id, e := observeIdentity(f, "change:0:0")
	if e != nil {
		t.Fatal(e)
	}
	sd := scope.Data()
	sd.RootIdentity = id
	scope = must(api.NewWorktreeScope(sd))
	d, e := acquireRoot(scope)
	if e != nil {
		t.Fatal(e)
	}
	d.close()
	if e = os.Mkdir(filepath.Join(root, "new"), 0700); e != nil {
		t.Fatal(e)
	}
	if d, e = acquireRoot(scope); e == nil {
		d.close()
		t.Fatal("change-profile drift ignored")
	}
}
