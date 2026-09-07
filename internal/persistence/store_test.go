package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func physicalStoreTemp(t testing.TB) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	return must(t, p, err)
}
func newTestStore(t testing.TB, root string) *Store {
	t.Helper()
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(root, "config.json"), PreferencesPath: filepath.Join(root, "state.json")})
	return must(t, s, err)
}
func nativeTestScope(t testing.TB, root string) api.WorktreeScope {
	t.Helper()
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	id, err := nativeDirectoryIdentity(c.parent())
	closeErr := c.close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	v := testScope(t, "native-root", id).Data()
	v.RootLocator = root
	s, err := api.NewWorktreeScope(v)
	return must(t, s, err)
}

func TestStoreLoadsMissingParentsWithoutWriting(t *testing.T) {
	root := physicalStoreTemp(t)
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(root, "missing", "config", "config.json"), PreferencesPath: filepath.Join(root, "missing", "state", "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.LoadUserConfig(context.Background())
	if err != nil || !u.Valid() || u.Observation().Data().State != api.LoadAbsent || !u.Document().Present() {
		t.Fatal(u, err)
	}
	p, err := s.LoadPreferences(context.Background())
	if err != nil || !p.Valid() || p.Observation().Data().State != api.LoadAbsent || !p.Document().Present() {
		t.Fatal(p, err)
	}
	r, err := s.LoadRunConfig(context.Background(), nativeTestScope(t, root))
	if err != nil || !r.Valid() || r.Observation().Data().State != api.LoadAbsent || !r.Document().Present() {
		t.Fatal(r, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("load/constructor created state", entries, err)
	}
}

func TestStoreLoadsStrictFamiliesAndRawVersions(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	for _, fixture := range []struct {
		raw    string
		state  api.StorageLoadState
		usable bool
	}{{` {"repos":{}} `, api.ValidLegacy, true}, {`{"schemaVersion":1,"stripPrefixes":[]}`, api.ValidCurrent, true}, {`{"schemaVersion":99}`, api.UnsupportedVersion, false}, {`{"repos":null}`, api.Corrupt, false}, {`{"schemaVersion":1,"x":1,"x":2}`, api.Corrupt, false}} {
		if err := os.WriteFile(filepath.Join(root, "config.json"), []byte(fixture.raw), 0600); err != nil {
			t.Fatal(err)
		}
		loaded, err := s.LoadUserConfig(context.Background())
		if !loaded.Valid() || loaded.Observation().Data().State != fixture.state || loaded.Document().Present() != fixture.usable || (err == nil) != fixture.usable {
			t.Fatalf("%q: %v %v", fixture.raw, loaded, err)
		}
		version, present := loaded.Observation().Data().Version.Value()
		if !present || version.ByteLength() != uint64(len(fixture.raw)) {
			t.Fatal("whole raw version missing")
		}
		actual, _ := os.ReadFile(filepath.Join(root, "config.json"))
		if string(actual) != fixture.raw {
			t.Fatal("load changed source")
		}
	}
	if err := os.WriteFile(filepath.Join(root, "state.json"), []byte(`{"lastFolders":{"Owner/Repo":"exact "}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := s.LoadPreferences(context.Background())
	if err != nil || p.Observation().Data().State != api.ValidLegacy {
		t.Fatal(p, err)
	}
	if err := os.Mkdir(filepath.Join(root, ".gh-tree"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gh-tree", "run.json"), []byte(`{"default":"Exact","launch":{"Exact":{"provider":"future","command":"literal"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadRunConfig(context.Background(), nativeTestScope(t, root))
	if err != nil || r.Observation().Data().State != api.ValidLegacy {
		t.Fatal(r, err)
	}
}

func TestStoreBindingsRejectOverlapAndMovement(t *testing.T) {
	root := physicalStoreTemp(t)
	path := filepath.Join(root, "shared.json")
	if _, err := New(context.Background(), Options{UserConfigPath: path, PreferencesPath: path}); err == nil {
		t.Fatal("same family location accepted")
	}
	if _, err := New(context.Background(), Options{UserConfigPath: path, PreferencesPath: path + ".lock"}); err == nil {
		t.Fatal("document overlaps another family's permanent lock")
	}
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(root, "linked.json")
	if err := os.Link(path, linked); err != nil {
		t.Fatal(err)
	}
	if _, err := New(context.Background(), Options{UserConfigPath: path, PreferencesPath: linked}); err == nil {
		t.Fatal("same file object accepted across families")
	}
	parent := filepath.Join(root, "selected")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	s := newTestStore(t, parent)
	if err := os.Rename(parent, parent+"-old"); err != nil {
		t.Fatal("request leaked parent guard", err)
	}
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	v, err := s.LoadUserConfig(context.Background())
	if err == nil || v.Document().Present() {
		t.Fatal("replaced parent was rebound")
	}
}

func TestStoreAdoptionCannotLaterReplaceObservedParent(t *testing.T) {
	root := physicalStoreTemp(t)
	path := filepath.Join(root, "new-parent")
	s := newTestStore(t, path)
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	first, err := s.LoadUserConfig(context.Background())
	if err != nil || !first.Document().Present() {
		t.Fatal(first, err)
	}
	if err := os.Rename(path, path+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	second, err := s.LoadUserConfig(context.Background())
	if err == nil || second.Document().Present() {
		t.Fatal("observed adoption silently replaced")
	}
}

func TestStoreRunRootBindingCancellationAndConcurrency(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	scope := nativeTestScope(t, root)
	wrong := scope.Data()
	wrong.RootIdentity = testDirectory(t, 9)
	bad, err := api.NewWorktreeScope(wrong)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadRunConfig(context.Background(), bad)
	if err == nil || r.Document().Present() {
		t.Fatal("foreign root accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u, err := s.LoadUserConfig(ctx)
	if !errors.Is(err, context.Canceled) || u.Document().Present() {
		t.Fatal("cancellation lost", err)
	}
	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				v, err := s.LoadRunConfig(context.Background(), scope)
				if err != nil || !v.Valid() {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	if _, err := New(context.Background(), Options{UserConfigPath: "relative.json", PreferencesPath: filepath.Join(root, "state.json")}); err == nil {
		t.Fatal("ambient relative location accepted")
	}
}

func TestStoreRunCannotOverlapExplicitUserStore(t *testing.T) {
	root := physicalStoreTemp(t)
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(root, ".gh-tree", "run.json"), PreferencesPath: filepath.Join(root, "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadRunConfig(context.Background(), nativeTestScope(t, root))
	if err == nil || r.Document().Present() {
		t.Fatal("run and user families overlap")
	}
}

func TestStoreOversizedNativeFileIsCorruptWithoutDefaults(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	f, err := os.Create(filepath.Join(root, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Truncate(api.MaxDocumentBytes + 1)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	loaded, err := s.LoadUserConfig(context.Background())
	if err == nil || !loaded.Valid() || loaded.Document().Present() || loaded.Observation().Data().State != api.Corrupt {
		t.Fatal("oversized input became defaults or unclassified absence", loaded, err)
	}
}
