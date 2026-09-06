package persistence

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func externalReplacement(t testing.TB, root, target string, raw []byte, present bool) {
	t.Helper()
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.close(); err != nil {
			t.Error(err)
		}
	}()
	payload, err := nativeCreateFile(c.parent(), "external-editor-payload", true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := payload.close(); err != nil {
			t.Error(err)
		}
	}()
	if err := writeComplete(context.Background(), payload.file, raw); err != nil {
		t.Fatal(err)
	}
	if err := nativePublish(payload, c.parent(), "external-editor-payload", target, present); err != nil {
		t.Fatal(err)
	}
}

func TestCommitExternalComparisonGapAndAbsenceCompetitor(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(map[bool]string{false: "absence", true: "presence"}[present], func(t *testing.T) {
			root := physicalStoreTemp(t)
			old := []byte(`{"schemaVersion":1,"stripPrefixes":["old"]}`)
			if present {
				if err := os.WriteFile(filepath.Join(root, "config.json"), old, 0600); err != nil {
					t.Fatal(err)
				}
			}
			s := newTestStore(t, root)
			load, err := s.LoadUserConfig(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			v, _ := load.Observation().Data().Version.Value()
			external := []byte(`{"schemaVersion":1,"stripPrefixes":["external"]}`)
			s.hook = func(stage string) error {
				if stage == "before-publication" {
					externalReplacement(t, root, "config.json", external, present)
				}
				return nil
			}
			result, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "ours"))
			if !result.Valid() {
				t.Fatalf("invalid gap result: %v", err)
			}
			current, ok := result.Data().CurrentVersion.Value()
			if !ok {
				t.Fatal("current competitor observation was discarded")
			}
			if !present {
				if err == nil || result.Data().Outcome != api.NotCommitted || current.SHA256() != sha256.Sum256(external) {
					t.Fatalf("absence overwrote competitor or lost current: %v %v", result, err)
				}
			} else {
				assertCommitted(t, result, err)
				// This is the explicitly unsupported arbitrary-editor gap: our
				// present replacement can overwrite an intervening external name.
				// Retention describes the observed original, never every unseen file.
				for _, r := range result.Data().Recovery {
					if r.Data().Kind == api.RawOriginal {
						data, e := os.ReadFile(r.Data().Locator)
						if e != nil || string(data) != string(old) {
							t.Fatalf("retained original was relabeled external: %q %v", data, e)
						}
					}
				}
			}
		})
	}
}

func TestCommitDetectedExternalChangeRefusesBeforePublication(t *testing.T) {
	root := physicalStoreTemp(t)
	if err := os.WriteFile(filepath.Join(root, "config.json"), []byte(`{"schemaVersion":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	s := newTestStore(t, root)
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := loaded.Observation().Data().Version.Value()
	external := []byte(`{"schemaVersion":1,"outside":true}`)
	s.hook = func(stage string) error {
		if stage == "final-check" {
			externalReplacement(t, root, "config.json", external, true)
		}
		return nil
	}
	r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "ours"))
	if !errors.Is(err, errBindingChanged) || !r.Valid() || r.Data().Outcome != api.NotCommitted {
		t.Fatalf("detected conflict: %v %v", r, err)
	}
	current, _ := r.Data().CurrentVersion.Value()
	if current.SHA256() != sha256.Sum256(external) {
		t.Fatal("current external bytes lost")
	}
}

func TestCommitMissingParentsAdoptsConcurrentScopeConformingCreation(t *testing.T) {
	root := physicalStoreTemp(t)
	parent := filepath.Join(root, "missing", "deep")
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(parent, "config.json"), PreferencesPath: filepath.Join(root, "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := loaded.Observation().Data().Version.Value()
	s.hook = func(stage string) error {
		if stage == "parents" {
			return os.MkdirAll(parent, 0700)
		}
		return nil
	}
	r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "adopted"))
	assertCommitted(t, r, err)
	loaded, err = s.LoadUserConfig(context.Background())
	if err != nil || !loaded.Valid() {
		t.Fatalf("adopted original anchor not reloadable: %v", err)
	}
}

func TestCommitOutcomeDeliveryPreservesCancellationAndIndependentCurrent(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := loaded.Observation().Data().Version.Value()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	external := []byte(`{"schemaVersion":1,"outside":"after-publication"}`)
	s.hook = func(stage string) error {
		if stage == "outcome-delivery" {
			cancel()
			externalReplacement(t, root, "config.json", external, true)
		}
		return nil
	}
	r, err := s.CommitUserConfig(ctx, userProposal(t, v, "ours"))
	assertCommitted(t, r, err)
	current, _ := r.Data().CurrentVersion.Value()
	proposal, _ := r.Data().ProposedVersion.Value()
	if !r.Data().CancellationAsked || current.SHA256() != sha256.Sum256(external) || current == proposal {
		t.Fatal("cancellation/current confused with publication")
	}
}
