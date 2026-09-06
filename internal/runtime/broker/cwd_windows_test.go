package broker

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

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

func makeJunction(t *testing.T, path, target string) {
	t.Helper()
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(p, windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	sub, err := windows.UTF16FromString(`\??\` + target)
	if err != nil {
		t.Fatal(err)
	}
	printName, err := windows.UTF16FromString(target)
	if err != nil {
		t.Fatal(err)
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
	if err := windows.DeviceIoControl(h, windows.FSCTL_SET_REPARSE_POINT, &b[0], uint32(len(b)), nil, 0, &returned, nil); err != nil {
		t.Fatal(err)
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
