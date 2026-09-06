//go:build linux || darwin || freebsd

package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"golang.org/x/sys/unix"
)

func TestUnixCommitRejectsReplacedPublicationEntryBeforeFinalChecks(t *testing.T) {
	for _, present := range []bool{false, true} {
		for _, symlink := range []bool{false, true} {
			t.Run(map[bool]string{false: "absent", true: "present"}[present]+map[bool]string{false: "-regular", true: "-link"}[symlink], func(t *testing.T) {
				root := physicalStoreTemp(t)
				old := []byte(`{"schemaVersion":1}`)
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
				s.hook = func(stage string) error {
					if stage != "final-check" {
						return nil
					}
					entries, err := os.ReadDir(root)
					if err != nil {
						return err
					}
					for _, entry := range entries {
						if !strings.HasSuffix(entry.Name(), ".publication") {
							continue
						}
						path := filepath.Join(root, entry.Name())
						if err := os.Rename(path, filepath.Join(root, "original-publication-entry")); err != nil {
							return err
						}
						if symlink {
							return os.Symlink("outside", path)
						}
						return os.WriteFile(path, []byte(`{"external":true}`), 0600)
					}
					t.Fatal("publication name not found")
					return nil
				}
				r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "ours"))
				if err == nil || !r.Valid() || r.Data().PublicationKnown || r.Data().Outcome != api.NotCommitted {
					t.Fatalf("replaced source published: %v %v", r, err)
				}
				current, readErr := os.ReadFile(filepath.Join(root, "config.json"))
				if present && (readErr != nil || string(current) != string(old)) || !present && !os.IsNotExist(readErr) {
					t.Fatalf("original target changed: %q %v", current, readErr)
				}
			})
		}
	}
}

func TestUnixCommitPreservesSetIDModeAfterDataWrites(t *testing.T) {
	for _, mode := range []uint32{04750, 02750, 0750, 0640} {
		root := physicalStoreTemp(t)
		path := filepath.Join(root, "config.json")
		if err := os.WriteFile(path, []byte(`{"schemaVersion":1}`), 0600); err != nil {
			t.Fatal(err)
		}
		if err := unix.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
		var before unix.Stat_t
		if err := unix.Stat(path, &before); err != nil {
			t.Fatal(err)
		}
		if uint32(before.Mode)&07777 != mode {
			t.Fatalf("fixture mode not established: %o/%o", before.Mode, mode)
		}
		s := newTestStore(t, root)
		loaded, err := s.LoadUserConfig(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		v, _ := loaded.Observation().Data().Version.Value()
		r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "mode"))
		assertCommitted(t, r, err)
		var after unix.Stat_t
		if err := unix.Stat(path, &after); err != nil {
			t.Fatal(err)
		}
		if uint32(after.Mode)&07777 != mode {
			t.Fatalf("set-ID/mode changed: want %o got %o", mode, after.Mode)
		}
	}
}

func TestUnixRecoveryUnsupportedEntryPreservesAllTypedLoadFamilies(t *testing.T) {
	for _, family := range []api.StorageFamily{api.UserConfig, api.Preferences, api.RunConfig} {
		root := physicalStoreTemp(t)
		s := newTestStore(t, root)
		scope := nativeTestScope(t, root)
		var name, parent string
		var result api.StorageCommitResult
		var err error
		switch family {
		case api.UserConfig:
			l, e := s.LoadUserConfig(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			v, _ := l.Observation().Data().Version.Value()
			result, err = s.CommitUserConfig(context.Background(), userProposal(t, v, "valid"))
			name, parent = "config.json", root
		case api.Preferences:
			l, e := s.LoadPreferences(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			v, _ := l.Observation().Data().Version.Value()
			d, _ := l.Document().Value()
			r, e := ports.NewPreferencesCommit(v, d)
			if e != nil {
				t.Fatal(e)
			}
			result, err = s.CommitPreferences(context.Background(), r)
			name, parent = "state.json", root
		case api.RunConfig:
			l, e := s.LoadRunConfig(context.Background(), scope)
			if e != nil {
				t.Fatal(e)
			}
			v, _ := l.Observation().Data().Version.Value()
			d, _ := l.Document().Value()
			r, e := ports.NewRunConfigCommit(scope, v, d)
			if e != nil {
				t.Fatal(e)
			}
			result, err = s.CommitRunConfig(context.Background(), r)
			name, parent = "run.json", filepath.Join(root, ".gh-tree")
		}
		assertCommitted(t, result, err)
		if err := os.Symlink("outside", filepath.Join(parent, recoveryPrefix(name)+"unsupported")); err != nil {
			t.Fatal(err)
		}
		var observation api.StorageLoadObservation
		valid, document := false, false
		switch family {
		case api.UserConfig:
			l, e := s.LoadUserConfig(context.Background())
			observation, err, valid, document = l.Observation(), e, l.Valid(), l.Document().Present()
		case api.Preferences:
			l, e := s.LoadPreferences(context.Background())
			observation, err, valid, document = l.Observation(), e, l.Valid(), l.Document().Present()
		case api.RunConfig:
			l, e := s.LoadRunConfig(context.Background(), scope)
			observation, err, valid, document = l.Observation(), e, l.Valid(), l.Document().Present()
		}
		if err == nil || !valid || !document || observation.Data().State != api.ValidCurrent || !observation.Data().Version.Present() || len(observation.Data().Recovery) == 0 {
			t.Fatalf("recovery error erased family %v load: valid=%v document=%v state=%v error=%v", family, valid, document, observation.Data().State, err)
		}
	}
}
