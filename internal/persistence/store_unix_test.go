//go:build linux || darwin || freebsd

package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestUnixExplicitLinkSelectionAndRunChildRefusal(t *testing.T) {
	root := physicalStoreTemp(t)
	external := physicalStoreTemp(t)
	selected := filepath.Join(external, "selected.json")
	if err := os.WriteFile(selected, []byte(`{"stripPrefixes":["Selected"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "choice.json")
	if err := os.Symlink(selected, link); err != nil {
		t.Fatal(err)
	}
	s, err := New(context.Background(), Options{UserConfigPath: link, PreferencesPath: filepath.Join(root, "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(external, "other.json"), link); err != nil {
		t.Fatal(err)
	}
	u, err := s.LoadUserConfig(context.Background())
	if err != nil || u.Observation().Data().State != api.ValidLegacy {
		t.Fatal(u, err)
	}
	d, _ := u.Document().Value()
	prefixes, _ := d.Data().StripPrefixes.Value()
	if len(prefixes) != 1 || prefixes[0] != "Selected" {
		t.Fatal("explicit selection was retargeted")
	}
	if err := os.Symlink(external, filepath.Join(root, ".gh-tree")); err != nil {
		t.Fatal(err)
	}
	r, err := s.LoadRunConfig(context.Background(), nativeTestScope(t, root))
	if err == nil || r.Document().Present() {
		t.Fatal("run child redirect accepted")
	}
}

func TestUnixRunMatchesSuppliedChangeStampWithoutUpgrade(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	c, err := unixAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := unixObserve(c.parent().fd())
	if err != nil {
		c.close()
		t.Fatal(err)
	}
	observed.stamp = fmt.Sprintf("change:%d:%d", observed.stat.Ctim.Sec, observed.stat.Ctim.Nsec)
	id, err := observed.directoryIdentity()
	closeErr := c.close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	data := nativeTestScope(t, root).Data()
	data.RootIdentity = id
	scope, err := api.NewWorktreeScope(data)
	if err != nil {
		t.Fatal(err)
	}
	first, err := s.LoadRunConfig(context.Background(), scope)
	if err != nil || !first.Document().Present() {
		t.Fatal("supplied change profile was silently upgraded", err)
	}
	if err := os.Mkdir(filepath.Join(root, "changed"), 0700); err != nil {
		t.Fatal(err)
	}
	second, err := s.LoadRunConfig(context.Background(), scope)
	if err == nil || second.Document().Present() {
		t.Fatal("change-profile drift ignored")
	}
}
