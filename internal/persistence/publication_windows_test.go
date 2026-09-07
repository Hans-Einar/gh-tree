package persistence

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func winTestPayload(t testing.TB, parent *winObject, name string, data []byte) *winObject {
	t.Helper()
	p, err := winOpen(parent.handle(), name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER, winShareAll, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := p.close(); err != nil {
			t.Error(err)
		}
	})
	if n, err := p.file.Write(data); err != nil || n != len(data) {
		t.Fatal(n, err)
	}
	if err := p.file.Sync(); err != nil {
		t.Fatal(err)
	}
	return p
}

func winTestRead(t testing.TB, parent *winObject, name string) []byte {
	t.Helper()
	f, err := winOpenDocument(parent, name)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(f.file)
	closeErr := f.close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
	return raw
}

func TestWindowsNativePublicationLayout(t *testing.T) {
	wantRoot, wantLength, wantName := uintptr(8), uintptr(16), uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		wantRoot, wantLength, wantName = 4, 8, 12
	}
	if unsafe.Offsetof(winNameInformation{}.Root) != wantRoot || unsafe.Offsetof(winNameInformation{}.Length) != wantLength || unsafe.Offsetof(winNameInformation{}.Name) != wantName {
		t.Fatal("wrong native class65/class11 ABI")
	}
}

func TestWindowsNativePublicationRetentionAndAbsence(t *testing.T) {
	root := t.TempDir()
	parent := acquiredWindows(t, root).parent()
	old := winTestPayload(t, parent, "run.json", []byte("old"))
	retained, err := winRetainOriginal(old, parent, "retained-original")
	if err != nil {
		t.Fatal(err)
	}
	defer retained.close()
	if _, err := winRetainOriginal(old, parent, "retained-original"); err == nil {
		t.Fatal("retention replaced existing object")
	}
	raw := winTestPayload(t, parent, "raw-backup", []byte("old"))
	_ = raw
	payload := winTestPayload(t, parent, "payload", []byte("new"))
	if err := winPublish(payload, parent, "run.json", false); err == nil {
		t.Fatal("absence publication replaced competitor")
	}
	if got := winTestRead(t, parent, "run.json"); string(got) != "old" {
		t.Fatal("failed no-replace changed original")
	}
	if err := winPublish(payload, parent, "run.json", true); err != nil {
		t.Fatal(err)
	}
	if _, err := old.file.WriteAt([]byte("late"), 0); err != nil {
		t.Fatal(err)
	}
	if err := old.file.Sync(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"run.json": "new", "retained-original": "late", "raw-backup": "old"} {
		got := winTestRead(t, parent, name)
		if string(got) != want {
			t.Fatalf("%s: %q", name, got)
		}
	}
	payload2 := winTestPayload(t, parent, "payload2", []byte("absent"))
	if err := winPublish(payload2, parent, "new.json", false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "payload2")); !os.IsNotExist(err) {
		t.Fatal("native rename did not consume temp name", err)
	}
}

func TestWindowsNativePublicationConcurrentReaders(t *testing.T) {
	parent := acquiredWindows(t, t.TempDir()).parent()
	values := [][]byte{bytes.Repeat([]byte("a"), 32769), bytes.Repeat([]byte("b"), 65539)}
	winTestPayload(t, parent, "run.json", values[0])
	stop, failures := make(chan struct{}), make(chan error, 1)
	var readers sync.WaitGroup
	for range 3 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				f, err := winOpenDocument(parent, "run.json")
				if err == nil {
					var raw []byte
					raw, err = io.ReadAll(f.file)
					closeErr := f.close()
					if err == nil {
						err = closeErr
					}
					if err == nil && !bytes.Equal(raw, values[0]) && !bytes.Equal(raw, values[1]) {
						err = fmt.Errorf("partial document: %d bytes", len(raw))
					}
				}
				if err != nil {
					select {
					case failures <- err:
					default:
					}
					return
				}
			}
		}()
	}
	for i := range 30 {
		p := winTestPayload(t, parent, fmt.Sprintf("payload-%d", i), values[i%2])
		if err := winPublish(p, parent, "run.json", true); err != nil {
			close(stop)
			readers.Wait()
			t.Fatal(err)
		}
	}
	close(stop)
	readers.Wait()
	select {
	case err := <-failures:
		t.Fatal(err)
	default:
	}
	// Read helper stays valid after publication as well as ordinary native reads.
	f, err := winOpenDocument(parent, "run.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.close()
	if _, _, err := winRead(context.Background(), f); err != nil {
		t.Fatal(err)
	}
}
