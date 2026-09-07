package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestWindowsExplicitJunctionSelectionAndRunRefusal(t *testing.T) {
	root := physicalStoreTemp(t)
	external := physicalStoreTemp(t)
	if err := os.WriteFile(filepath.Join(external, "selected.json"), []byte(`{"stripPrefixes":["Selected"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "choice")
	if err := os.Mkdir(link, 0700); err != nil {
		t.Fatal(err)
	}
	if err := setTestJunction(link, external); err != nil {
		t.Fatal(err)
	}
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(link, "selected.json"), PreferencesPath: filepath.Join(root, "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.LoadUserConfig(context.Background())
	if err != nil || u.Observation().Data().State != api.ValidLegacy {
		t.Fatal(u, err)
	}
	run := filepath.Join(root, ".gh-tree")
	if err := os.Mkdir(run, 0700); err != nil {
		t.Fatal(err)
	}
	if err := setTestJunction(run, external); err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadRunConfig(context.Background(), nativeTestScope(t, root))
	if err == nil || r.Document().Present() {
		t.Fatal("run junction accepted")
	}
}
