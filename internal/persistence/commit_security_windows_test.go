package persistence

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"structs"
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
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	// Runtime worker threads/events can grow during a cold test. Count the
	// native File type (including directory guards and locks) instead, with
	// automatic GC disabled so finalizers cannot conceal request-owned leaks.
	probe, err := os.Create(filepath.Join(root, "handle-count-reference"))
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Close()
	previousGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previousGC)
	count := func() int { return windowsTestFileHandleCount(t, windows.Handle(probe.Fd())) }
	before := count()
	var duplicate windows.Handle
	if err := windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(probe.Fd()), windows.CurrentProcess(), &duplicate, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		t.Fatal(err)
	}
	withExtra := count()
	if err := windows.CloseHandle(duplicate); err != nil {
		t.Fatal(err)
	}
	if withExtra != before+1 || count() != before {
		t.Fatal("native file-handle counter failed the deliberate open/close control")
	}
	for i := 0; i < 12; i++ {
		loaded, err := s.LoadUserConfig(context.Background())
		if err != nil {
			dumpWindowsRecoveryIdentities(t, root)
			t.Fatalf("iteration %d: %v", i, err)
		}
		v, _ := loaded.Observation().Data().Version.Value()
		r, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "same bytes, independent operation"))
		assertCommitted(t, r, err)
		if got := count(); got != before {
			t.Fatalf("iteration %d request file-handle growth: before %d after %d", i, before, got)
		}
	}
	after := count()
	if after != before {
		t.Fatalf("request handle growth: before %d after %d", before, after)
	}
}

// ProcessHandleInformation=51, current process only. Native layout source:
// https://github.com/winsiderss/phnt/blob/master/ntpsapi.h
type windowsTestHandleEntry struct {
	_                                  structs.HostLayout
	Handle                             windows.Handle
	HandleCount, PointerCount          uintptr
	Access, Type, Attributes, Reserved uint32
}

func windowsTestFileHandleCount(t testing.TB, reference windows.Handle) int {
	t.Helper()
	// A fixed 1 MiB bound comfortably covers this test process; truncation fails.
	words := make([]uintptr, (1<<20)/unsafe.Sizeof(uintptr(0)))
	var length uint32
	if err := windows.NtQueryInformationProcess(windows.CurrentProcess(), 51, unsafe.Pointer(&words[0]), uint32(len(words)*int(unsafe.Sizeof(uintptr(0)))), &length); err != nil {
		t.Fatal(err)
	}
	header := 2 * unsafe.Sizeof(uintptr(0))
	entrySize := unsafe.Sizeof(windowsTestHandleEntry{})
	if uintptr(length) < header || words[0] > (uintptr(length)-header)/entrySize {
		t.Fatal("invalid current-process handle snapshot length")
	}
	entries := unsafe.Slice((*windowsTestHandleEntry)(unsafe.Pointer(&words[2])), int(words[0]))
	var fileType uint32
	for _, entry := range entries {
		if entry.Handle == reference {
			fileType = entry.Type
		}
	}
	if fileType == 0 {
		t.Fatal("native reference file is missing from current-process snapshot")
	}
	n := 0
	for _, entry := range entries {
		if entry.Type == fileType {
			n++
		}
	}
	return n
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
