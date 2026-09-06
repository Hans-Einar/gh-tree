package launchdiscovery

import (
	"context"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"os"
	"path/filepath"
	"testing"
)

func fixtureScope(t *testing.T, path string) api.WorktreeScope {
	t.Helper()
	path, e := filepath.EvalSymlinks(path)
	if e != nil {
		t.Fatal(e)
	}
	f, chain, e := nativeRoot(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	for _, a := range chain {
		defer a.Close()
	}
	id, e := observeIdentity(f, "")
	if e != nil {
		t.Fatal(e)
	}
	r, e := domain.NewRepositoryID(domain.LocalCommon, "discovery-fixture")
	if e != nil {
		t.Fatal(e)
	}
	w, e := domain.NewWorktreeID(r, "primary")
	if e != nil {
		t.Fatal(e)
	}
	v, e := api.NewSourceVersion("fixture", "root", "tests", "1")
	if e != nil {
		t.Fatal(e)
	}
	s, e := api.NewWorktreeScope(api.WorktreeScopeData{ID: w, RootLocator: path, RootIdentity: id, Source: v})
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestNativeRetainedObservation(t *testing.T) {
	root := t.TempDir()
	if e := os.Mkdir(filepath.Join(root, " project"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, " project", "package.json"), []byte(`{"scripts":{"dev":"x"}}`), 0600); e != nil {
		t.Fatal(e)
	}
	s := fixtureScope(t, root)
	d, e := acquireRoot(s)
	if e != nil {
		t.Fatal(e)
	}
	defer d.close()
	p, e := childDirectory(d, " project")
	if e != nil {
		t.Fatal(e)
	}
	defer p.close()
	o, e := observeFile(context.Background(), p, "package.json", true, 4096)
	if e != nil || o.state != "regular" {
		t.Fatal(o, e)
	}
	if e = sameNamedDirectory(d, " project", p); e != nil {
		t.Fatal(e)
	}
	if _, e = observeFile(context.Background(), p, "package.json", true, 2); e != errLimit {
		t.Fatal(e)
	}
	missing, e := observeFile(context.Background(), p, "absent", false, 4096)
	if e != nil || missing.state != "absent" {
		t.Fatal(missing, e)
	}
	bad := s.Data()
	bad.RootIdentity, _ = api.NewDirectoryIdentity(bad.RootIdentity.Platform(), bad.RootIdentity.Device(), bad.RootIdentity.FileID(), "unsupported-stamp")
	wrong, _ := api.NewWorktreeScope(bad)
	if x, e := acquireRoot(wrong); e == nil {
		x.close()
		t.Fatal("unsupported stamp accepted")
	}
}
func TestNativeLinksAndReplacement(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if e := os.WriteFile(filepath.Join(outside, "package.json"), []byte(`{"scripts":{"outside":"x"}}`), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(filepath.Join(root, "project"), 0700); e != nil {
		t.Fatal(e)
	}
	s := fixtureScope(t, root)
	d, e := acquireRoot(s)
	if e != nil {
		t.Fatal(e)
	}
	p, e := childDirectory(d, "project")
	if e != nil {
		t.Fatal(e)
	}
	p.close()
	d.close()
	if e = os.Rename(filepath.Join(root, "project"), filepath.Join(root, "old")); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(filepath.Join(root, "project"), 0700); e != nil {
		t.Fatal(e)
	}
	d, e = acquireRoot(s)
	if e != nil { // change-stamp profiles correctly invalidate child namespace changes.
		if s.Data().RootIdentity.Stamp()[:6] == "change" {
			return
		}
		t.Fatal(e)
	}
	defer d.close()
	replacement, e := childDirectory(d, "project")
	if e != nil {
		t.Fatal(e)
	}
	defer replacement.close()
	if replacement.identity == p.identity {
		t.Fatal("replacement identity reused")
	}
	if e = os.Symlink(outside, filepath.Join(root, "linked")); e != nil {
		t.Logf("symlink privilege unavailable: %v", e)
		return
	}
	if x, e := childDirectory(d, "linked"); e == nil {
		x.close()
		t.Fatal("followed child link")
	}
	if e = os.Symlink(filepath.Join(outside, "package.json"), filepath.Join(root, "package.json")); e != nil {
		t.Fatal(e)
	}
	if _, e = observeFile(context.Background(), d, "package.json", true, 4096); e == nil {
		t.Fatal("followed manifest link")
	}
}
