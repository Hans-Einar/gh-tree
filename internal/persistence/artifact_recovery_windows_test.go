package persistence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

type windowsArtifactSnapshot struct {
	Identity diskIdentity
	Birth    windows.Filetime
}

func windowsArtifactSnapshots(t testing.TB, root string) map[string]windowsArtifactSnapshot {
	t.Helper()
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]windowsArtifactSnapshot{}
	for _, entry := range entries {
		if !manifestName(entry.Name()) {
			continue
		}
		manifest, err := nativeOpenDocument(c.parent(), entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		raw, err := nativeRead(context.Background(), manifest)
		closeErr := manifest.close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
		m, err := decodeManifest(raw)
		if err != nil || m.Preparing {
			t.Fatal("expected complete native manifest", err)
		}
		for _, artifact := range m.Artifacts {
			object, err := nativeOpenDocument(c.parent(), artifact.Name)
			if artifact.Name == m.publicationName() && errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			identityErr := verifyArtifactIdentity(object, artifact.Identity)
			v, observeErr := winObserve(object.handle())
			closeErr := object.close()
			if identityErr != nil || observeErr != nil || closeErr != nil {
				t.Fatal(identityErr, observeErr, closeErr)
			}
			out[artifact.Name] = windowsArtifactSnapshot{artifact.Identity, v.basic.CreationTime}
		}
	}
	return out
}

func windowsCheckStableRecovery(t testing.TB, root string, earlier map[api.RecoveryID]api.StorageRecovery) {
	t.Helper()
	loaded, err := newTestStore(t, root).LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	current := map[api.RecoveryID]api.StorageRecovery{}
	for _, record := range loaded.Observation().Data().Recovery {
		current[record.Data().Record.Data().RecoveryID] = record
	}
	// Account for every persisted artifact except the independently absent
	// publication name. A stable subset is insufficient recovery evidence.
	expected := map[api.RecoveryID]bool{}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !manifestName(entry.Name()) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		m, err := decodeManifest(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, artifact := range m.Artifacts {
			if artifact.Name == m.publicationName() {
				continue
			}
			id, err := api.NewRecoveryID("persistence:" + artifact.ID)
			if err != nil {
				t.Fatal(err)
			}
			expected[id] = true
			if _, ok := current[id]; !ok {
				t.Fatal("persisted retained artifact ID missing from load", artifact.Name)
			}
		}
	}
	if len(expected) != len(current) {
		t.Fatal("unexpected recovery records", len(expected), len(current))
	}
	for id, old := range earlier {
		if now, ok := current[id]; !ok || !reflect.DeepEqual(old, now) {
			t.Fatal("previously issued recovery changed or disappeared", id)
		}
	}
	for id, record := range current {
		earlier[id] = record
	}
}

func windowsSeedTunnel(t testing.TB, root string, present bool) {
	t.Helper()
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	seed, err := nativeCreateFile(c.parent(), "config.json", true)
	if err != nil {
		t.Fatal(err)
	}
	defer seed.close()
	if err := writeComplete(context.Background(), seed.file, []byte(`{"schemaVersion":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := seed.file.Sync(); err != nil {
		t.Fatal(err)
	}
	if !present {
		if err := winSetName(seed, c.parent(), "tunnel-seed-preserved", 0, winRenameInformationEx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWindowsArtifactPublicRepeatedTunnelingAndFreshProcesses(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprintf("present=%v", present), func(t *testing.T) {
			root := physicalStoreTemp(t)
			windowsSeedTunnel(t, root, present)
			stable := map[api.RecoveryID]api.StorageRecovery{}
			changedBirth := 0
			for range 4 {
				s := newTestStore(t, root)
				loaded, err := s.LoadUserConfig(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				v, _ := loaded.Observation().Data().Version.Value()
				var before map[string]windowsArtifactSnapshot
				s.hook = func(stage string) error {
					if stage == "manifest-flushed" {
						before = windowsArtifactSnapshots(t, root)
					}
					return nil
				}
				r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "identical repeated content"))
				assertCommitted(t, r, err)
				after := windowsArtifactSnapshots(t, root)
				for name, old := range before {
					now, exists := after[name]
					if !exists && strings.HasSuffix(name, ".publication") {
						continue
					}
					if !exists || now.Identity != old.Identity {
						t.Fatal("recorded native tuple changed", name)
					}
					if now.Birth != old.Birth {
						changedBirth++
					}
				}
				windowsCheckStableRecovery(t, root, stable)
			}
			if changedBirth == 0 {
				t.Fatal("public fixture did not execute creation-time tunneling")
			}
			for range 4 {
				child := startCommitChild(t, root, "")
				if err := child.write("GO\n"); err != nil {
					t.Fatal(err)
				}
				if got := child.line(t, "RESULT "); got != fmt.Sprintf("RESULT %d", api.CommittedDurabilityUncertain) {
					t.Fatal(got)
				}
				if err := child.cmd.Wait(); err != nil {
					t.Fatal(err, child.stderr.String())
				}
				windowsCheckStableRecovery(t, root, stable)
				windowsArtifactSnapshots(t, root)
			}
			if expected := 30 + 2*boolInt(present); len(stable) != expected {
				t.Fatalf("expected %d IDs, observed %d", expected, len(stable))
			}
			t.Logf("%d tunneled retained payload observations; %d stable recovery IDs across fresh stores and four child processes", changedBirth, len(stable))
		})
	}
}

func TestWindowsArtifactPublicCrashBeforeOutcome(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprintf("present=%v", present), func(t *testing.T) {
			root := physicalStoreTemp(t)
			windowsSeedTunnel(t, root, present)
			child := startCommitChild(t, root, "native-return-lost")
			if err := child.write("GO\n"); err != nil {
				t.Fatal(err)
			}
			child.line(t, "AT native-return-lost")
			// These are already flushed prepublication frames. The child has not
			// executed any postpublication query, journal write or outcome delivery.
			beforeDeath := windowsArtifactSnapshots(t, root)
			if err := child.cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err := child.cmd.Wait(); err == nil {
				t.Fatal("killed child returned success")
			}
			if got := windowsArtifactSnapshots(t, root); !reflect.DeepEqual(beforeDeath, got) {
				t.Fatal("owner death changed persisted native association")
			}
			stable := map[api.RecoveryID]api.StorageRecovery{}
			windowsCheckStableRecovery(t, root, stable)
			windowsCheckStableRecovery(t, root, stable)
			if len(stable) != 2+2*boolInt(present) {
				t.Fatal("lost death-before-outcome recovery", len(stable))
			}
		})
	}
}

func TestWindowsArtifactPublicOriginalRemainsUntouched(t *testing.T) {
	for _, tagged := range []bool{false, true} {
		t.Run(fmt.Sprintf("tagged=%v", tagged), func(t *testing.T) {
			root := physicalStoreTemp(t)
			parent := acquiredWindows(t, root).parent()
			directoryID, directoryErr := winArtifactObjectID(parent.handle(), windows.FSCTL_GET_OBJECT_ID)
			original := winTestPayload(t, parent, "config.json", []byte(` {"schemaVersion":1,"stripPrefixes":["old"]} `))
			if tagged {
				if _, err := probeWindowsObjectID(original, true); err != nil {
					t.Fatal(err)
				}
			}
			id, idErr := winArtifactObjectID(original.handle(), windows.FSCTL_GET_OBJECT_ID)
			before, err := winInspectMetadata(original)
			if err != nil {
				t.Fatal(err)
			}
			content, err := nativeRead(context.Background(), original)
			if err != nil {
				t.Fatal(err)
			}
			s := newTestStore(t, root)
			loaded, err := s.LoadUserConfig(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			v, _ := loaded.Observation().Data().Version.Value()
			r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "new"))
			assertCommitted(t, r, err)
			currentDirectoryID, currentDirectoryErr := winArtifactObjectID(parent.handle(), windows.FSCTL_GET_OBJECT_ID)
			if directoryID != currentDirectoryID || directoryErr != currentDirectoryErr {
				t.Fatal("directory ObjectID mutated")
			}
			lock, err := nativeOpenDocument(parent, "config.json.lock")
			if err != nil {
				t.Fatal(err)
			}
			_, lockIDErr := winArtifactObjectID(lock.handle(), windows.FSCTL_GET_OBJECT_ID)
			closeErr := lock.close()
			if lockIDErr != windows.ERROR_FILE_NOT_FOUND || closeErr != nil {
				t.Fatal("lock acquired an artifact ID", lockIDErr, closeErr)
			}
			afterID, afterIDErr := winArtifactObjectID(original.handle(), windows.FSCTL_GET_OBJECT_ID)
			after, err := winInspectMetadata(original)
			afterContent, readErr := nativeRead(context.Background(), original)
			if err != nil || readErr != nil || id != afterID || idErr != afterIDErr || !before.equal(after) || !bytes.Equal(content, afterContent) {
				t.Fatal("public commit mutated original bytes/security/ID", err, readErr, idErr, afterIDErr)
			}
			for name, snapshot := range windowsArtifactSnapshots(t, root) {
				if strings.HasSuffix(name, ".original") && !tagged {
					if !strings.HasPrefix(snapshot.Identity.Stamp, "birth-filetime:") {
						t.Fatal("untagged original upgraded")
					}
				} else if !strings.HasPrefix(snapshot.Identity.Stamp, winObjectIDStamp) {
					t.Fatal("owned artifact or tagged original lacks ObjectID profile", name)
				}
			}
		})
	}
}

func TestWindowsArtifactAllRolesRejectSameByteSubstitution(t *testing.T) {
	for _, kind := range []api.StorageRecoveryKind{api.Manifest, api.RawOriginal, api.RetainedOriginal, api.RetainedPayload} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			root, m := manifestFixture(t)
			name := filepath.Join(root, m.artifactName(kind))
			raw, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(name, filepath.Join(root, "preserved-replaced-object")); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			got, err := observedFixture(t, root, m)
			if err == nil {
				t.Fatal("same-byte substitution accepted")
			}
			if kind == api.Manifest && len(got) != 0 || kind != api.Manifest && len(got) != 3 {
				t.Fatal("independent recovery records lost or substituted", len(got))
			}
			for _, record := range got {
				if record.Data().Kind == kind {
					t.Fatal("replacement inherited original recovery ID")
				}
			}
			preserved, err := os.ReadFile(filepath.Join(root, "preserved-replaced-object"))
			if err != nil || !bytes.Equal(raw, preserved) {
				t.Fatal("preserved original changed", err)
			}
		})
	}
}

func TestWindowsArtifactLegacyFlatAndJournalRemainExact(t *testing.T) {
	for _, journal := range []bool{false, true} {
		for _, drift := range []bool{false, true} {
			t.Run(fmt.Sprintf("journal=%v/drift=%v", journal, drift), func(t *testing.T) {
				root, m := manifestFixture(t)
				c, err := nativeAcquire(context.Background(), root)
				if err != nil {
					t.Fatal(err)
				}
				for i := range m.Artifacts {
					o, err := nativeOpenDocument(c.parent(), m.Artifacts[i].Name)
					if err != nil {
						t.Fatal(err)
					}
					m.Artifacts[i].Identity = winBirthArtifact(o.observation)
					if err := o.close(); err != nil {
						t.Fatal(err)
					}
				}
				if err := c.close(); err != nil {
					t.Fatal(err)
				}
				if drift {
					m.Artifacts[0].Identity.Stamp += "0"
				}
				raw, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				if journal {
					body := manifestFrameBody{Sequence: 1, Snapshot: m}
					bodyRaw, _ := json.Marshal(body)
					frameRaw, _ := json.Marshal(manifestFrame{Body: body, Digest: sha256.Sum256(bodyRaw)})
					raw = append(append([]byte(manifestJournalMagic), frameRaw...), '\n')
				}
				path := filepath.Join(root, m.artifactName(api.Manifest))
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
				decoded, err := decodeManifest(raw)
				if err != nil || !reflect.DeepEqual(decoded, m) {
					t.Fatal("old record shape/hash no longer readable", err)
				}
				first, err := observedFixture(t, root, m)
				if drift {
					if err == nil || len(first) != 0 {
						t.Fatal("drifting legacy self identity accepted")
					}
				} else {
					if err != nil || len(first) != 4 {
						t.Fatal("matching legacy profile refused", err)
					}
					second, err := observedFixture(t, root, m)
					if err != nil || !reflect.DeepEqual(first, second) {
						t.Fatal("old IDs/observations changed", err)
					}
				}
				unchanged, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(raw, unchanged) {
					t.Fatal("legacy frame was rewritten", err)
				}
				prior := m
				prior.Preparing = true
				next := prior
				next.Artifacts = append([]diskArtifact(nil), prior.Artifacts...)
				next.Artifacts[0].Identity.Stamp = winObjectIDStamp + strings.Repeat("01", 64)
				if err := manifestSuccessor(prior, next); err == nil {
					t.Fatal("journal permitted identity upgrade")
				}
			})
		}
	}
}
