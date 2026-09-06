package persistence

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsPublicCommitCreatesProtectedPerUserFilesAndLock(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := loaded.Observation().Data().Version.Value()
	r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "private"))
	assertCommitted(t, r, err)
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.close(); err != nil {
			t.Error(err)
		}
	}()
	expected, err := winUserSecurity()
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := expected.DACL()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := winACLBytes(acl)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	seenLock := false
	for _, entry := range entries {
		object, err := nativeOpenDocument(c.parent(), entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		metadata, inspectErr := winInspectSecurity(object.handle())
		closeErr := object.close()
		if inspectErr != nil || closeErr != nil {
			t.Fatal(inspectErr, closeErr)
		}
		if !bytes.Equal(policy, metadata.dacl) || metadata.control&windows.SE_DACL_PROTECTED == 0 {
			t.Fatalf("new per-user object has inherited/broader ACL: %s", entry.Name())
		}
		if entry.Name() == "config.json.lock" {
			seenLock = true
		}
	}
	if !seenLock {
		t.Fatal("permanent lock not inspected")
	}
}

func TestWindowsPublicCommitAndLoadsReleaseAllRequestHandles(t *testing.T) {
	countProc := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetProcessHandleCount")
	count := func() uint32 {
		var n uint32
		ok, _, err := countProc.Call(uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&n)))
		if ok == 0 {
			t.Fatal(err)
		}
		return n
	}
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	before := count()
	for i := 0; i < 12; i++ {
		loaded, err := s.LoadUserConfig(context.Background())
		if err != nil {
			dumpWindowsRecoveryIdentities(t, root)
			t.Fatalf("iteration %d: %v", i, err)
		}
		v, _ := loaded.Observation().Data().Version.Value()
		r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "same bytes, independent operation"))
		assertCommitted(t, r, err)
	}
	after := count()
	if after > before+2 {
		t.Fatalf("request handle growth: before %d after %d", before, after)
	}
}

func dumpWindowsRecoveryIdentities(t testing.TB, root string) {
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Log(err)
		return
	}
	defer c.close()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Log(err)
		return
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".manifest") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Log(err)
			continue
		}
		m, err := decodeManifest(raw)
		if err != nil {
			t.Log(err)
			continue
		}
		for _, a := range m.Artifacts {
			o, err := nativeOpenDocument(c.parent(), a.Name)
			if err != nil {
				continue
			}
			actual, err := nativeArtifactIdentity(o)
			o.close()
			if err != nil || actual != a.Identity {
				t.Logf("kind %d sameFileID=%v recordedStamp=%s actualStamp=%s error=%v", a.Kind, actual.File == a.Identity.File, a.Identity.Stamp, actual.Stamp, err)
			}
		}
	}
}
