package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
)

func assertCommitted(t testing.TB, result api.StorageCommitResult, err error) {
	t.Helper()
	if err != nil || !result.Valid() || !result.Data().PublicationKnown || result.Data().Outcome == api.NotCommitted || !result.Data().ProposedVersion.Present() {
		t.Fatalf("commit: valid=%v outcome=%v known=%v err=%v", result.Valid(), result.Data().Outcome, result.Data().PublicationKnown, err)
	}
}

func userProposal(t testing.TB, expected api.StorageVersion, prefixes ...string) ports.UserConfigCommit {
	t.Helper()
	doc, err := decodeUserConfig([]byte(`{"schemaVersion":1}`))
	if err != nil {
		t.Fatal(err)
	}
	data := doc.Data()
	data.StripPrefixes = api.PresentField(prefixes)
	doc, err = api.NewUserConfigDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ports.NewUserConfigCommit(expected, doc)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestCommitAllFamiliesAndRestartRecovery(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	u, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	uv, _ := u.Observation().Data().Version.Value()
	r, err := s.CommitUserConfig(context.Background(), userProposal(t, uv, "first"))
	assertCommitted(t, r, err)
	if len(r.Data().Recovery) < 2 {
		t.Fatal("missing prepared recovery")
	}
	fresh := newTestStore(t, root)
	u, err = fresh.LoadUserConfig(context.Background())
	if err != nil || !u.Valid() || len(u.Observation().Data().Recovery) < 2 {
		t.Fatalf("recovery load: %v %v", u, err)
	}
	ids := map[api.RecoveryID]bool{}
	for _, record := range r.Data().Recovery {
		ids[record.Data().Record.Data().RecoveryID] = true
	}
	for _, record := range u.Observation().Data().Recovery {
		if !ids[record.Data().Record.Data().RecoveryID] {
			t.Fatal("restart minted new artifact ID")
		}
	}
	uv, _ = u.Observation().Data().Version.Value()
	r, err = fresh.CommitUserConfig(context.Background(), userProposal(t, uv, "second"))
	assertCommitted(t, r, err)
	p, err := s.LoadPreferences(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	pv, _ := p.Observation().Data().Version.Value()
	pd, _ := p.Document().Value()
	pc, err := ports.NewPreferencesCommit(pv, pd)
	if err != nil {
		t.Fatal(err)
	}
	r, err = s.CommitPreferences(context.Background(), pc)
	assertCommitted(t, r, err)
	scope := nativeTestScope(t, root)
	l, err := s.LoadRunConfig(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	lv, _ := l.Observation().Data().Version.Value()
	ld, _ := l.Document().Value()
	rc, err := ports.NewRunConfigCommit(scope, lv, ld)
	if err != nil {
		t.Fatal(err)
	}
	r, err = s.CommitRunConfig(context.Background(), rc)
	assertCommitted(t, r, err)
	l, err = fresh.LoadRunConfig(context.Background(), nativeTestScope(t, root))
	if err != nil || !l.Valid() || len(l.Observation().Data().Recovery) < 2 {
		t.Fatalf("run recovery: %v", err)
	}
}

func TestCommitStaleWholeVersionAndForeignStoreRefuse(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	version, _ := loaded.Observation().Data().Version.Value()
	r, err := s.CommitUserConfig(context.Background(), userProposal(t, version, "first"))
	assertCommitted(t, r, err)
	r, err = s.CommitUserConfig(context.Background(), userProposal(t, version, "stale"))
	if err == nil || !r.Valid() || r.Data().Outcome != api.NotCommitted {
		t.Fatalf("stale result: %v %v", r, err)
	}
	other := newTestStore(t, physicalStoreTemp(t))
	r, err = other.CommitUserConfig(context.Background(), userProposal(t, version, "foreign"))
	if err == nil || !r.Valid() || r.Data().Outcome != api.NotCommitted {
		t.Fatalf("foreign result: %v %v", r, err)
	}
}

func TestCommitMissingUserParentsPreservesExpectedAnchor(t *testing.T) {
	root := physicalStoreTemp(t)
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(root, "missing", "deep", "config.json"), PreferencesPath: filepath.Join(root, "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	version, _ := loaded.Observation().Data().Version.Value()
	r, err := s.CommitUserConfig(context.Background(), userProposal(t, version, "created"))
	assertCommitted(t, r, err)
	var manifestPath string
	for _, recovery := range r.Data().Recovery {
		if recovery.Data().Kind == api.Manifest {
			manifestPath = recovery.Data().Locator
		}
	}
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m, err := decodeManifest(raw)
	if err != nil || m.Expected != versionRecord(version) || len(m.ExpectedRemaining) != 2 || m.Expected.Store == m.Original.Store || m.Original.Store != m.Proposed.Store {
		t.Fatalf("missing-parent transition erased: %+v %v", m, err)
	}
	loaded, err = s.LoadUserConfig(context.Background())
	if err != nil || !loaded.Valid() {
		t.Fatalf("created binding not reloadable: %v", err)
	}
}

func TestCommitFaultsPreserveKnownOutcomeAndCancellation(t *testing.T) {
	for _, stage := range []string{"lock", "prepare.payload", "manifest-flushed", "before-publication", "native-return-lost", "directory-flush", "outcome-delivery"} {
		t.Run(stage, func(t *testing.T) {
			root := physicalStoreTemp(t)
			s := newTestStore(t, root)
			loaded, err := s.LoadUserConfig(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			v, _ := loaded.Observation().Data().Version.Value()
			injected := errors.New("injected " + stage)
			s.hook = func(at string) error {
				if at == stage {
					return injected
				}
				return nil
			}
			r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "proposal"))
			if !errors.Is(err, injected) || !r.Valid() {
				t.Fatalf("fault was lost: %v %v", r, err)
			}
			if stage == "native-return-lost" {
				if r.Data().Outcome != api.StorageIndeterminate || r.Data().PublicationKnown {
					t.Fatal("lost return became known")
				}
			} else if stage == "directory-flush" || stage == "outcome-delivery" {
				if !r.Data().PublicationKnown || r.Data().Outcome == api.NotCommitted {
					t.Fatal("postcommit error erased effect")
				}
			} else if r.Data().Outcome != api.NotCommitted || r.Data().PublicationKnown {
				t.Fatal("precommit fault published")
			}
		})
	}
}
