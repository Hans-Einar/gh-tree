package broker

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func windowsSpec(t *testing.T) StartSpec {
	t.Helper()
	dir := t.TempDir()
	p, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	id, err := observeHandle(h)
	if closeErr := windows.CloseHandle(h); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	return StartSpec{ParentID: uint64(os.Getpid()), OperationID: 1, RootLocator: dir, RootIdentity: id, ProjectIdentity: id, Executable: "owned-fixture.exe", Rows: 24, Columns: 80}
}

func TestWindowsCwdDataGuards(t *testing.T) {
	s := windowsSpec(t)
	a, err := AcquireCwd(s)
	if a != nil {
		defer func() {
			if e := a.Close(); e != nil {
				t.Error(e)
			}
		}()
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(s.RootLocator, s.RootLocator+"-moved"); err == nil {
		t.Fatal("acquired directory could be renamed")
	}
	if err = os.Remove(filepath.Join(s.RootLocator, a.anchorName)); err == nil {
		t.Fatal("data-read anchor could be deleted")
	}
	if err = a.Revalidate(); err != nil {
		t.Fatal(err)
	}
	name := a.anchorName
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(s.RootLocator, name)); !os.IsNotExist(err) {
		t.Fatalf("owned anchor remains: %v", err)
	}
}

func TestWindowsCwdStaleAndReparseRefused(t *testing.T) {
	s := windowsSpec(t)
	old := s.RootLocator + "-original"
	if err := os.Rename(s.RootLocator, old); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(old)
	if err := os.Mkdir(s.RootLocator, 0700); err != nil {
		t.Fatal(err)
	}
	a, err := AcquireCwd(s)
	if a != nil {
		if e := a.Close(); e != nil {
			t.Fatal(e)
		}
	}
	if err == nil {
		t.Fatal("stale root identity accepted")
	}
	if err := os.Remove(s.RootLocator); err != nil {
		t.Fatal(err)
	}
	makeJunction(t, s.RootLocator, old)
	a, err = AcquireCwd(s)
	if a != nil {
		if e := a.Close(); e != nil {
			t.Fatal(e)
		}
	}
	if err == nil {
		t.Fatal("reparse root accepted")
	}
}

func TestWindowsExistingDataAnchorPreserved(t *testing.T) {
	s := windowsSpec(t)
	path := filepath.Join(s.RootLocator, "existing-data")
	if err := os.WriteFile(path, []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	a, err := AcquireCwd(s)
	if a != nil {
		defer a.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	if a.anchorOwned || a.anchorName != "existing-data" {
		t.Fatalf("existing readable child not pinned: %+v", a)
	}
	if err = os.Remove(path); err == nil {
		t.Fatal("existing data pin allowed deletion")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "preserved" {
		t.Fatalf("existing anchor lost: %q %v", data, err)
	}
}

func makeJunction(t *testing.T, path, target string) {
	t.Helper()
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := setJunction(path, target); err != nil {
		t.Fatal(err)
	}
}

func setJunction(path, target string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(p, windows.FILE_WRITE_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	sub, err := windows.UTF16FromString(`\??\` + target)
	if err != nil {
		return err
	}
	printName, err := windows.UTF16FromString(target)
	if err != nil {
		return err
	}
	b := make([]byte, 16+(len(sub)+len(printName))*2)
	binary.LittleEndian.PutUint32(b, windows.IO_REPARSE_TAG_MOUNT_POINT)
	binary.LittleEndian.PutUint16(b[4:], uint16(len(b)-8))
	binary.LittleEndian.PutUint16(b[10:], uint16((len(sub)-1)*2))
	binary.LittleEndian.PutUint16(b[12:], uint16(len(sub)*2))
	binary.LittleEndian.PutUint16(b[14:], uint16((len(printName)-1)*2))
	for i, v := range append(sub, printName...) {
		binary.LittleEndian.PutUint16(b[16+i*2:], v)
	}
	var returned uint32
	return windows.DeviceIoControl(h, windows.FSCTL_SET_REPARSE_POINT, &b[0], uint32(len(b)), nil, 0, &returned, nil)
}

func TestWindowsMetadataControlAndInPlaceReparseBarriers(t *testing.T) {
	t.Run("metadata-only-negative-control", func(t *testing.T) {
		s := windowsSpec(t)
		outside := t.TempDir()
		p, err := windows.UTF16PtrFromString(s.RootLocator)
		if err != nil {
			t.Fatal(err)
		}
		h, err := windows.CreateFile(p, windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer windows.CloseHandle(h)
		if err = setJunction(s.RootLocator, outside); err != nil {
			t.Fatalf("negative control did not demonstrate metadata weakness: %v", err)
		}
		if err = directoryAttributes(h); err == nil {
			t.Fatal("metadata guard hid reparse mutation")
		}
	})
	for _, stage := range []string{"anchor-acquired", "user-created-suspended", "before-user-resume", "cwd-breakpoint"} {
		t.Run(stage, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			s := windowsSpec(t)
			outside := t.TempDir()
			s.Environment = os.Environ()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "pipe"}
			attempted := false
			p := &userProcess{hook: func(actual string) {
				if actual == stage {
					attempted = true
					if e := setJunction(s.RootLocator, outside); e == nil {
						t.Errorf("in-place reparse succeeded at %s", stage)
					}
				}
			}}
			if err = p.prepare(s); err != nil {
				t.Fatal(err)
			}
			var readers sync.WaitGroup
			for _, out := range p.outputs {
				readers.Add(1)
				go func(out nativeOutput) {
					defer readers.Done()
					buffer := make([]byte, 1024)
					for {
						if _, e := out.file.Read(buffer); e != nil {
							return
						}
					}
				}(out)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err = p.start(ctx, s); err != nil {
				t.Fatal(err)
			}
			if _, err = waitProcess(ctx, p.debug.process.Process); err != nil {
				t.Fatal(err)
			}
			if err = p.cleanup(ctx); err != nil {
				t.Fatal(err)
			}
			readers.Wait()
			for i := range p.outputs {
				if err = closeFile(&p.outputs[i].file); err != nil {
					t.Error(err)
				}
			}
			if !attempted {
				t.Fatal("named startup barrier not exercised")
			}
			if _, err = os.Stat(filepath.Join(outside, "user-ran")); !os.IsNotExist(err) {
				t.Fatalf("outside replacement ran: %v", err)
			}
		})
	}
}

func TestWindowsStartupLayoutAndRouting(t *testing.T) {
	if err := assertStartupLayout(); err != nil {
		t.Fatal(err)
	}
	machine, embedded, err := MachineRoute()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("native machine=%04x embedded=%v", machine, embedded)
}
