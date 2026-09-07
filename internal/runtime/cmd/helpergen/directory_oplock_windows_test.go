package main

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestDirectoryOplockContinuousInvalidationProbe(t *testing.T) {
	for _, change := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged-release", true: "preexisting-handle-insertion-and-restore"}[change], func(t *testing.T) {
			path := t.TempDir()
			name, err := windows.UTF16PtrFromString(path)
			if err != nil {
				t.Fatal(err)
			}
			mutator, err := windows.CreateFile(name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.WRITE_DAC, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer windows.CloseHandle(mutator)
			h, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OVERLAPPED, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer windows.CloseHandle(h)
			event, err := windows.CreateEvent(nil, 1, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer windows.CloseHandle(event)
			ov := windows.Overlapped{HEvent: event}
			var in [12]byte
			binary.LittleEndian.PutUint16(in[0:], 1)
			binary.LittleEndian.PutUint16(in[2:], 12)
			binary.LittleEndian.PutUint32(in[4:], 1)
			binary.LittleEndian.PutUint32(in[8:], 1)
			var out [24]byte
			var size uint32
			err = windows.DeviceIoControl(h, 0x90240, &in[0], 12, &out[0], 24, &size, &ov)
			if !errors.Is(err, windows.ERROR_IO_PENDING) {
				t.Fatalf("grant directory R oplock: %v", err)
			}
			defer func() {
				windows.CancelIoEx(h, &ov)
				err := windows.GetOverlappedResult(h, &ov, &size, true)
				if err != nil && !errors.Is(err, windows.ERROR_OPERATION_ABORTED) {
					t.Errorf("join directory watch: %v", err)
				}
			}()
			if wait, err := windows.WaitForSingleObject(event, 0); err != nil || wait != uint32(windows.WAIT_TIMEOUT) {
				t.Fatalf("initial watch: %d %v", wait, err)
			}
			if change {
				child, err := windows.NewNTUnicodeString("unrecorded_windows.go")
				if err != nil {
					t.Fatal(err)
				}
				oa := windows.OBJECT_ATTRIBUTES{RootDirectory: mutator, ObjectName: child, Attributes: windows.OBJ_CASE_INSENSITIVE}
				oa.Length = uint32(unsafe.Sizeof(oa))
				var file windows.Handle
				var status windows.IO_STATUS_BLOCK
				if err := windows.NtCreateFile(&file, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE, &oa, &status, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0); err != nil {
					t.Fatal(err)
				}
				windows.CloseHandle(file)
				if err := os.Remove(filepath.Join(path, "unrecorded_windows.go")); err != nil {
					t.Fatal(err)
				}
				if wait, err := windows.WaitForSingleObject(event, 5000); err != nil || wait != windows.WAIT_OBJECT_0 {
					t.Fatalf("insertion break: %d %v", wait, err)
				}
				if err := windows.GetOverlappedResult(h, &ov, &size, false); err != nil {
					t.Fatal(err)
				}
				t.Logf("preexisting-handle insert/remove is durably invalidated: original=%#x new=%#x flags=%#x", binary.LittleEndian.Uint32(out[4:]), binary.LittleEndian.Uint32(out[8:]), binary.LittleEndian.Uint32(out[12:]))
			} else {
				if err := windows.CancelIoEx(h, &ov); err != nil {
					t.Fatal(err)
				}
				if err := windows.GetOverlappedResult(h, &ov, &size, true); !errors.Is(err, windows.ERROR_OPERATION_ABORTED) {
					t.Fatalf("release result: %v", err)
				}
				if err := os.WriteFile(filepath.Join(path, "after-release.txt"), []byte("allowed"), 0600); err != nil {
					t.Fatal(err)
				}
				t.Log("unchanged watch remains pending until canceled/joined; writes after release succeed")
			}
		})
	}
}
