package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func manifestFixture(t testing.TB) (string, recoveryManifest) {
	t.Helper()
	root := physicalStoreTemp(t)
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.close(); err != nil {
			t.Error(err)
		}
	}()
	lock, err := nativeLock(context.Background(), c.parent(), "config.json", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.close(); err != nil {
			t.Error(err)
		}
	}()
	parent, err := nativeDirectoryIdentity(c.parent())
	if err != nil {
		t.Fatal(err)
	}
	store, err := bindingToken(parent, nil, "config.json")
	if err != nil {
		t.Fatal(err)
	}
	nonce, err := operationNonce()
	if err != nil {
		t.Fatal(err)
	}
	original := []byte(" {\"repos\":{}} ")
	proposed := []byte(`{"schemaVersion":1,"repos":{}}`)
	m := recoveryManifest{
		SchemaVersion: 1, Nonce: nonce, Family: api.UserConfig, Basename: "config.json", Parent: directoryRecord(parent), ExpectedAnchor: directoryRecord(parent),
		Original: diskVersion{store, true, uint64(len(original)), sha256.Sum256(original)},
		Proposed: diskVersion{store, true, uint64(len(proposed)), sha256.Sum256(proposed)},
	}
	m.Expected = m.Original
	if err := os.WriteFile(filepath.Join(root, "config.json"), original, 0600); err != nil {
		t.Fatal(err)
	}
	old, err := nativeOpenOriginal(c.parent(), "config.json")
	if err != nil {
		t.Fatal(err)
	}
	retained, retainErr := nativeRetainOriginal(old, c.parent(), "config.json", m.artifactName(api.RetainedOriginal))
	closeErr := old.close()
	if retainErr != nil || closeErr != nil {
		t.Fatal(retainErr, closeErr)
	}
	var manifestObject *nativeObject
	for _, kind := range []api.StorageRecoveryKind{api.Manifest, api.RetainedPayload, api.RawOriginal, api.RetainedOriginal} {
		object := retained
		var content []byte
		if kind != api.RetainedOriginal {
			object, err = nativeCreateFile(c.parent(), m.artifactName(kind), true)
			if err != nil {
				t.Fatal(err)
			}
		}
		if kind == api.RetainedPayload {
			content = proposed
		}
		if kind == api.RawOriginal {
			content = original
		}
		if len(content) != 0 {
			if err := writeComplete(context.Background(), object.file, content); err != nil {
				t.Fatal(err)
			}
			if err := object.file.Sync(); err != nil {
				t.Fatal(err)
			}
		}
		id, err := operationNonce()
		if err != nil {
			t.Fatal(err)
		}
		identity, err := nativeArtifactIdentity(object)
		if err != nil {
			t.Fatal(err)
		}
		m.Artifacts = append(m.Artifacts, diskArtifact{kind, m.artifactName(kind), id, identity, uint64(len(content)), sha256.Sum256(content)})
		if kind == api.Manifest {
			manifestObject = object
		} else if err := object.close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.validate(api.UserConfig, api.WorktreeScope{}, "config.json", directoryRecord(parent)); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeComplete(context.Background(), manifestObject.file, raw); err != nil {
		t.Fatal(err)
	}
	if err := manifestObject.file.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := manifestObject.close(); err != nil {
		t.Fatal(err)
	}
	return root, m
}

func observedFixture(t testing.TB, root string, m recoveryManifest) ([]api.StorageRecovery, error) {
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
	lock, err := nativeLock(context.Background(), c.parent(), "config.json", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.close(); err != nil {
			t.Error(err)
		}
	}()
	return observeManifest(context.Background(), c, m.artifactName(api.Manifest), "config.json", root, api.UserConfig, api.WorktreeScope{})
}

func TestManifestRestartStableRecoveryIDsAndLateOriginalWrites(t *testing.T) {
	root, m := manifestFixture(t)
	first, err := observedFixture(t, root, m)
	if err != nil || len(first) != 4 {
		t.Fatalf("initial recovery: %d %v", len(first), err)
	}
	second, err := observedFixture(t, root, m)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("restart minted or changed records: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.json"), []byte("late writer changed original inode"), 0600); err != nil {
		t.Fatal(err)
	}
	third, err := observedFixture(t, root, m)
	if err != nil || !reflect.DeepEqual(first, third) {
		t.Fatalf("late original write erased exact retained association: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(root, m.artifactName(api.RawOriginal)))
	if err != nil || sha256.Sum256(raw) != m.Original.Digest {
		t.Fatalf("late writer changed immutable backup: %q, %v", raw, err)
	}
}

func TestManifestRefusesSameByteArtifactReplacement(t *testing.T) {
	root, m := manifestFixture(t)
	name := filepath.Join(root, m.artifactName(api.RawOriginal))
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(name, filepath.Join(root, "original-backup-object")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	records, err := observedFixture(t, root, m)
	if err == nil || len(records) != 3 {
		t.Fatalf("same bytes authorized substituted artifact: %d %v", len(records), err)
	}
	for _, record := range records {
		if record.Data().Kind == api.RawOriginal {
			t.Fatal("substituted raw artifact reported with retained ID")
		}
	}
}

func TestManifestRefusesForeignBindingAndMalformedData(t *testing.T) {
	_, m := manifestFixture(t)
	for _, mutate := range []func(*recoveryManifest){
		func(x *recoveryManifest) { x.Family = api.Preferences },
		func(x *recoveryManifest) { x.Parent.Device++ },
		func(x *recoveryManifest) { x.Expected.Store = "foreign" },
		func(x *recoveryManifest) { x.ExpectedAnchor.Stamp += "changed" },
		func(x *recoveryManifest) { x.Original.Digest[0]++ },
		func(x *recoveryManifest) { x.Artifacts[0].Name = "../outside" },
		func(x *recoveryManifest) { x.Artifacts[0].ID = x.Artifacts[1].ID },
		func(x *recoveryManifest) { x.Artifacts = x.Artifacts[:2] },
	} {
		copy := m
		copy.Artifacts = append([]diskArtifact(nil), m.Artifacts...)
		mutate(&copy)
		if err := copy.validate(api.UserConfig, api.WorktreeScope{}, "config.json", m.Parent); err == nil {
			t.Fatal("invalid manifest accepted")
		}
	}
	for _, raw := range [][]byte{[]byte(`{"SchemaVersion":1,"SchemaVersion":1}`), []byte(`{"Unknown":1}`), []byte(`null`)} {
		if got, err := decodeManifest(raw); err == nil && got.validate(api.UserConfig, api.WorktreeScope{}, "config.json", m.Parent) == nil {
			t.Fatalf("malformed manifest accepted: %q", raw)
		}
	}
}
