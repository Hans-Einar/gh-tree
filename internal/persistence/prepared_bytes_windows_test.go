package persistence

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

var testMappedCopy = windows.NewLazySystemDLL("ntdll.dll").NewProc("RtlMoveMemory")

func TestWindowsPreparedBytesGuardLifetime(t *testing.T) {
	for _, present := range []bool{false, true} {
		t.Run(fmt.Sprint(present), func(t *testing.T) {
			root := physicalStoreTemp(t)
			if present {
				if err := os.WriteFile(filepath.Join(root, "config.json"), []byte(`{"schemaVersion":1}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			s := newTestStore(t, root)
			loaded, err := s.LoadUserConfig(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			version, _ := loaded.Observation().Data().Version.Value()
			c, err := nativeAcquire(context.Background(), root)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := c.close(); err != nil {
					t.Error(err)
				}
			}()
			stages := map[string]bool{
				"prepare.payload.write": false, "prepare.payload.flush": false,
				"prepare.payload.journal": false, "prepare.publication.journal": false,
				"prepare.ready.journal.write": false, "manifest-flushed": false,
				"prepare.close": false, "final-check": false, "before-publication": false,
			}
			var mapped uintptr
			defer func() {
				if mapped != 0 {
					if err := windows.UnmapViewOfFile(mapped); err != nil {
						t.Error(err)
					}
				}
			}()
			var late []byte
			attempts := 0
			s.hook = func(at string) error {
				if _, check := stages[at]; check {
					stages[at] = true
					entries, err := os.ReadDir(root)
					if err != nil {
						return err
					}
					seen := 0
					for _, entry := range entries {
						if !strings.HasSuffix(entry.Name(), ".payload") && !strings.HasSuffix(entry.Name(), ".publication") {
							continue
						}
						seen++
						for _, access := range []uint32{windows.FILE_WRITE_DATA, windows.FILE_APPEND_DATA, windows.FILE_GENERIC_WRITE} {
							o, err := winOpen(c.parent().handle(), entry.Name(), windows.FILE_GENERIC_READ|access, winShareAll, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
							if o != nil {
								err = errors.Join(err, o.close())
								return fmt.Errorf("%s admitted prepared write access %#x: %w", at, access, err)
							}
							if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
								return fmt.Errorf("%s unexpected writer refusal: %w", at, err)
							}
							attempts++
						}
						if at == "prepare.payload.flush" && strings.HasSuffix(entry.Name(), ".payload") {
							o, err := nativeOpenDocument(c.parent(), entry.Name())
							if err != nil {
								return err
							}
							defer func() {
								if err := o.close(); err != nil {
									t.Error(err)
								}
							}()
							section, err := windows.CreateFileMapping(o.handle(), nil, windows.PAGE_READWRITE, 0, 0, nil)
							if section != 0 {
								windows.CloseHandle(section)
								return errors.New("read handle created writable section")
							}
							if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
								return fmt.Errorf("unexpected writable section refusal: %w", err)
							}
							section, err = windows.CreateFileMapping(o.handle(), nil, windows.PAGE_READONLY, 0, 0, nil)
							if err != nil {
								return err
							}
							defer func() {
								if err := windows.CloseHandle(section); err != nil {
									t.Error(err)
								}
							}()
							view, err := windows.MapViewOfFile(section, windows.FILE_MAP_WRITE, 0, 0, 0)
							if view != 0 {
								windows.UnmapViewOfFile(view)
								return errors.New("readonly section admitted writable view")
							}
							if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
								return fmt.Errorf("unexpected writable view refusal: %w", err)
							}
							mapped, err = windows.MapViewOfFile(section, windows.FILE_MAP_READ, 0, 0, 0)
							if err != nil {
								return err
							}
						}
					}
					if seen == 0 {
						return fmt.Errorf("%s never observed payload", at)
					}
				}
				if at == "native-return-lost" {
					// This runs before result delivery and request cleanup. The real
					// native call succeeded; a permitted later edit stays a distinct
					// CurrentVersion rather than erasing the known proposal effect.
					o, err := winOpen(c.parent().handle(), "config.json", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, winShareAll, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
					if err != nil {
						return err
					}
					defer func() {
						if err := o.close(); err != nil {
							t.Error(err)
						}
					}()
					raw, err := nativeRead(context.Background(), o)
					if err != nil {
						return err
					}
					late = bytes.ReplaceAll(raw, []byte("proposed"), []byte("changed!"))
					if bytes.Equal(raw, late) {
						return errors.New("late-write fixture has no proposal marker")
					}
					if n, err := o.file.WriteAt(late, 0); err != nil || n != len(late) {
						return fmt.Errorf("late target write %d: %w", n, err)
					}
					return o.file.Sync()
				}
				return nil
			}
			result, err := s.CommitUserConfig(context.Background(), userProposal(t, version, "proposed"))
			assertCommitted(t, result, err)
			for at, hit := range stages {
				if !hit {
					t.Fatalf("unexecuted byte boundary %s", at)
				}
			}
			if mapped == 0 || len(late) == 0 {
				t.Fatal("missing mapping/late-write control")
			}
			actual, err := os.ReadFile(filepath.Join(root, "config.json"))
			mappedBytes := make([]byte, len(late))
			// MapViewOfFile supplies native memory, not a Go pointer. Copy the
			// known mapped file length without converting its uintptr to a Go
			// slice or bypassing vet/checkptr.
			testMappedCopy.Call(uintptr(unsafe.Pointer(&mappedBytes[0])), mapped, uintptr(len(mappedBytes)))
			runtime.KeepAlive(mappedBytes)
			if err != nil || !bytes.Equal(actual, late) || !bytes.Equal(mappedBytes, late) {
				t.Fatalf("late target/mapped-view bytes: %s %v", actual, err)
			}
			d := result.Data()
			if d.CurrentVersion == d.ProposedVersion || !d.PublicationKnown || d.Outcome != api.CommittedDurabilityUncertain {
				t.Fatalf("lost known effect or later observation: %+v", d)
			}
			t.Logf("%d actual writer-sharing refusals across %d stages; read mapping retained; postpublication target write accepted", attempts, len(stages))
		})
	}
}

func TestWindowsEarlierWritableMappingCannotBecomePreparedPayload(t *testing.T) {
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
	o, err := nativeCreateFile(c.parent(), "candidate", true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if o != nil {
			if err := o.close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if err := writeComplete(context.Background(), o.file, []byte("original")); err != nil {
		t.Fatal(err)
	}
	section, err := windows.CreateFileMapping(o.handle(), nil, windows.PAGE_READWRITE, 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if section != 0 {
			if err := windows.CloseHandle(section); err != nil {
				t.Error(err)
			}
		}
	}()
	view, err := windows.MapViewOfFile(section, windows.FILE_MAP_WRITE, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := windows.UnmapViewOfFile(view); err != nil {
			t.Error(err)
		}
	}()
	if err := windows.CloseHandle(section); err != nil {
		t.Fatal(err)
	}
	section = 0
	if err := o.close(); err != nil {
		o = nil
		t.Fatal(err)
	}
	o = nil
	// Closing the writer and section handles does not revoke the mapped view.
	// Its retained write access also prevents a later deny-write-sharing guard;
	// the real prepared creator must never adopt this existing object.
	late, err := winOpen(c.parent().handle(), "candidate", windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_DELETE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
	if late != nil {
		late.close()
		t.Fatal("later guard admitted an earlier writable view")
	}
	if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("unexpected earlier-map refusal: %v", err)
	}
	reader, err := nativeOpenDocument(c.parent(), "candidate")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reader.close(); err != nil {
			t.Error(err)
		}
	}()
	modified := []byte("modified")
	testMappedCopy.Call(view, uintptr(unsafe.Pointer(&modified[0])), uintptr(len(modified)))
	runtime.KeepAlive(modified)
	if err := windows.FlushViewOfFile(view, 8); err != nil {
		t.Fatal(err)
	}
	actual, err := nativeRead(context.Background(), reader)
	if err != nil || string(actual) != "modified" {
		t.Fatalf("earlier mapping negative control: %s %v", actual, err)
	}
	created, err := nativeCreatePayloadMetadata(c.parent(), "candidate", true, nil)
	if created != nil {
		created.close()
		t.Fatal("prepared creator adopted existing mapped file")
	}
	if !errors.Is(err, os.ErrExist) && !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
		t.Fatalf("unexpected exclusive-create refusal: %v", err)
	}
	actual, err = nativeRead(context.Background(), reader)
	if err != nil || string(actual) != "modified" {
		t.Fatalf("exclusive refusal altered prior mapping bytes: %s %v", actual, err)
	}
}
