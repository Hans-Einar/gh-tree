//go:build linux || darwin || freebsd

package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
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

func TestUnixRunCommitOwnEffectRebindsChangeStampAndPreservesOriginalRequest(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	scopeWithChange := func() api.WorktreeScope {
		c, err := nativeAcquire(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		observation, err := unixObserve(c.parent().fd())
		if err != nil {
			c.close()
			t.Fatal(err)
		}
		observation.stamp = fmt.Sprintf("change:%d:%d", observation.stat.Ctim.Sec, observation.stat.Ctim.Nsec)
		identity, err := observation.directoryIdentity()
		if err := c.close(); err != nil {
			t.Fatal(err)
		}
		if err != nil {
			t.Fatal(err)
		}
		data := nativeTestScope(t, root).Data()
		data.RootIdentity = identity
		scope, err := api.NewWorktreeScope(data)
		if err != nil {
			t.Fatal(err)
		}
		return scope
	}
	scope := scopeWithChange()
	loaded, err := s.LoadRunConfig(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := loaded.Observation().Data().Version.Value()
	doc, _ := loaded.Document().Value()
	request, err := ports.NewRunConfigCommit(scope, expected, doc)
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.CommitRunConfig(context.Background(), request)
	assertCommitted(t, result, err)
	freshScope := scopeWithChange()
	proposed, _ := result.Data().ProposedVersion.Value()
	if proposed.MatchesRunScope(scope) || !proposed.MatchesRunScope(freshScope) {
		t.Fatal("own root creation did not rebind the changed observation")
	}
	for _, recovery := range result.Data().Recovery {
		if recovery.Data().Kind != api.Manifest {
			continue
		}
		raw, err := os.ReadFile(recovery.Data().Locator)
		if err != nil {
			t.Fatal(err)
		}
		manifest, err := decodeManifest(raw)
		if err != nil || manifest.Expected != versionRecord(expected) || !manifest.ExpectedScope.matches(scope) || !manifest.Scope.matches(freshScope) {
			t.Fatalf("original/rebound run evidence lost: %v", err)
		}
	}
	if _, err := s.LoadRunConfig(context.Background(), scope); err == nil {
		t.Fatal("stale supplied change stamp silently accepted")
	}
	loaded, err = s.LoadRunConfig(context.Background(), freshScope)
	if err != nil || !loaded.Valid() || len(loaded.Observation().Data().Recovery) < 2 {
		t.Fatalf("rebound manifest association failed: %v", err)
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
