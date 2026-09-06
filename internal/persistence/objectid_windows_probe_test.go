package persistence

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// Test-only mechanism investigation. Production does not yet select object IDs.
// CREATE_OR_GET is used exclusively on files this fixture just created with
// FILE_CREATE; every later observation is GET-only. No SET/DELETE or timestamp
// normalization, case setting, global setting or alternate publisher is used.
func probeWindowsObjectID(o *winObject, create bool) ([64]byte, error) {
	var id [64]byte // FILE_OBJECTID_BUFFER: four consecutive 16-byte identifiers.
	var n uint32
	code := uint32(windows.FSCTL_GET_OBJECT_ID)
	if create {
		code = windows.FSCTL_CREATE_OR_GET_OBJECT_ID
	}
	err := windows.DeviceIoControl(o.handle(), code, nil, 0, &id[0], uint32(len(id)), &n, nil)
	if err == nil && (n != uint32(len(id)) || *(*[16]byte)(id[:16]) == [16]byte{}) {
		err = errors.New("invalid native object-ID result")
	}
	return id, err
}

type probeWindowsIdentity struct {
	ID     [64]byte
	File   [16]byte
	Volume uint64
	Birth  windows.Filetime
}

func probeWindowsIdentityOf(t testing.TB, o *winObject, create bool) probeWindowsIdentity {
	t.Helper()
	id, err := probeWindowsObjectID(o, create)
	if err != nil {
		t.Fatal(err)
	}
	v, err := winObserve(o.handle())
	if err != nil {
		t.Fatal(err)
	}
	return probeWindowsIdentity{id, v.id.File, v.id.Volume, v.basic.CreationTime}
}

func probeWindowsSameIncarnation(a, b probeWindowsIdentity) bool {
	return a.ID == b.ID && a.File == b.File && a.Volume == b.Volume
}

func TestWindowsObjectIDProbeRepeatedNativePublication(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprintf("present=%v", present), func(t *testing.T) {
			root := physicalStoreTemp(t)
			parent := acquiredWindows(t, root).parent()
			seed := winTestPayload(t, parent, "config.json", []byte("same bytes"))
			seedID := probeWindowsIdentityOf(t, seed, true)
			// Renaming the owned seed seeds the destination-name tunnel cache.
			if err := winSetName(seed, parent, "seed-retained", 0, winRenameInformationEx); err != nil {
				t.Fatal(err)
			}
			identities := map[string]probeWindowsIdentity{"seed-retained": seedID}
			if present {
				initial := winTestPayload(t, parent, "initial", []byte("same bytes"))
				initialID := probeWindowsIdentityOf(t, initial, true)
				retained, err := winRetainOriginal(initial, parent, "initial-retained")
				if err != nil {
					t.Fatal(err)
				}
				if err := retained.close(); err != nil {
					t.Fatal(err)
				}
				if err := winPublish(initial, parent, "config.json", false); err != nil {
					t.Fatal(err)
				}
				identities["initial-retained"] = initialID
			}
			var changedBirth int
			for i := range 12 {
				name := fmt.Sprintf("payload-%02d", i)
				payload := winTestPayload(t, parent, name, []byte("same bytes"))
				id := probeWindowsIdentityOf(t, payload, true)
				if id.ID == seedID.ID {
					t.Fatal("new object reused retained seed object ID")
				}
				retainedName := fmt.Sprintf("retained-%02d", i)
				retained, err := winRetainOriginal(payload, parent, retainedName)
				if err != nil {
					t.Fatal(err)
				}
				if err := retained.close(); err != nil {
					t.Fatal(err)
				}
				if err := payload.file.Sync(); err != nil {
					t.Fatal(err)
				}
				if err := winPublish(payload, parent, "config.json", present); err != nil {
					t.Fatal(err)
				}
				after := probeWindowsIdentityOf(t, payload, false)
				if !probeWindowsSameIncarnation(id, after) {
					t.Fatal("publication changed intrinsic identity")
				}
				if id.Birth != after.Birth {
					changedBirth++
				}
				identities[retainedName] = id
				for name, expected := range identities {
					opened, err := winOpenDocument(parent, name)
					if err != nil {
						t.Fatalf("iteration %d retained name %s: %v", i, name, err)
					}
					if _, getErr := probeWindowsObjectID(opened, false); getErr != nil {
						t.Fatalf("iteration %d GET-only identity %s: %v", i, name, getErr)
					}
					actual := probeWindowsIdentityOf(t, opened, false)
					if err := opened.close(); err != nil {
						t.Fatal(err)
					}
					if !probeWindowsSameIncarnation(expected, actual) {
						t.Fatalf("retained object identity changed: %s", name)
					}
				}
				if !present {
					if err := winSetName(payload, parent, fmt.Sprintf("published-%02d", i), 0, winRenameInformationEx); err != nil {
						t.Fatal(err)
					}
				}
			}
			if changedBirth == 0 {
				t.Fatal("fixture did not exercise native creation-time tunneling")
			}
			t.Logf("12 class65 publications: %d tunneled creation times; complete intrinsic IDs retained across every reopen", changedBirth)
		})
	}
}

func TestWindowsObjectIDProbeSameByteSubstitutionAndReadOnlyQuery(t *testing.T) {
	root := physicalStoreTemp(t)
	parent := acquiredWindows(t, root).parent()
	original := winTestPayload(t, parent, "original", []byte("same"))
	if _, err := probeWindowsObjectID(original, false); err == nil {
		t.Fatal("GET unexpectedly found an ID on fresh untagged file")
	}
	id := probeWindowsIdentityOf(t, original, true)
	if err := original.file.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := winSetName(original, parent, "preserved-original", 0, winRenameInformationEx); err != nil {
		t.Fatal(err)
	}
	replacement := winTestPayload(t, parent, "original", []byte("same"))
	replacementID := probeWindowsIdentityOf(t, replacement, true)
	if probeWindowsSameIncarnation(id, replacementID) {
		t.Fatal("same-byte replacement inherited complete original intrinsic identity")
	}
	t.Logf("replacement: sameObjectID=%v sameFileID=%v", id.ID == replacementID.ID, id.File == replacementID.File)
	// A GET on an untagged user original must fail without creating an ID.
	untagged := winTestPayload(t, parent, "untagged-user-original", []byte("user"))
	for range 2 {
		if _, err := probeWindowsObjectID(untagged, false); err == nil {
			t.Fatal("GET mutated untagged original")
		}
	}
}

func TestWindowsObjectIDProbeCrashFixture(t *testing.T) {
	root := os.Getenv("GH_TREE_TEST_OBJECT_ID_ROOT")
	if root == "" {
		t.Skip("private process fixture")
	}
	parent := acquiredWindows(t, root).parent()
	payload := winTestPayload(t, parent, "payload", []byte("proposal"))
	id := probeWindowsIdentityOf(t, payload, true)
	retained, err := winRetainOriginal(payload, parent, "retained")
	if err != nil {
		t.Fatal(err)
	}
	if err := retained.close(); err != nil {
		t.Fatal(err)
	}
	evidence := winTestPayload(t, parent, "identity.json", nil)
	raw, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeComplete(context.Background(), evidence.file, raw); err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(payload.file.Sync(), evidence.file.Sync()); err != nil {
		t.Fatal(err)
	}
	if err := winPublish(payload, parent, "config.json", false); err != nil {
		t.Fatal(err)
	}
	// Identity evidence precedes the native call; no post-publication observation,
	// outcome write or timestamp/ID mutation occurs before the parent kills us.
	fmt.Println("PUBLISHED")
	for {
		time.Sleep(time.Hour)
	}
}

func TestWindowsObjectIDProbeCrashBeforeOutcome(t *testing.T) {
	root := physicalStoreTemp(t)
	{
		parent := acquiredWindows(t, root).parent()
		seed := winTestPayload(t, parent, "config.json", []byte("prior"))
		probeWindowsIdentityOf(t, seed, true)
		if err := winSetName(seed, parent, "seed-retained", 0, winRenameInformationEx); err != nil {
			t.Fatal(err)
		}
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	child := &commitChild{cmd: exec.CommandContext(ctx, exe, "-test.run=^TestWindowsObjectIDProbeCrashFixture$", "-test.timeout=20s")}
	child.cmd.Env = append(os.Environ(), "GH_TREE_TEST_OBJECT_ID_ROOT="+root)
	child.cmd.Stderr = &child.stderr
	out, err := child.cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	child.out = bufio.NewScanner(out)
	if err := child.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if child.cmd.ProcessState == nil {
			child.cmd.Process.Kill()
			child.cmd.Wait()
		}
	}()
	child.line(t, "PUBLISHED")
	if err := child.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := child.cmd.Wait(); err == nil {
		t.Fatal("killed child returned success")
	}
	raw, err := os.ReadFile(filepath.Join(root, "identity.json"))
	if err != nil {
		t.Fatal(err)
	}
	var before probeWindowsIdentity
	if err := json.Unmarshal(raw, &before); err != nil {
		t.Fatal(err)
	}
	parent := acquiredWindows(t, root).parent()
	for _, name := range []string{"retained", "config.json"} {
		o, err := winOpenDocument(parent, name)
		if err != nil {
			t.Fatal(err)
		}
		after := probeWindowsIdentityOf(t, o, false)
		if err := o.close(); err != nil {
			t.Fatal(err)
		}
		if !probeWindowsSameIncarnation(before, after) {
			t.Fatalf("prepublication identity lost after killed owner: %s", name)
		}
		if before.Birth == after.Birth {
			t.Fatal("crash fixture did not exercise native creation-time tunneling")
		}
	}
}
