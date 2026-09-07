//go:build linux || darwin || freebsd

package persistence

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"golang.org/x/sys/unix"
)

func unixTestPayload(t testing.TB, parent *unixObject, name string, data []byte) *unixObject {
	t.Helper()
	p, err := unixOpen(parent.fd(), name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL, 0600, false)
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

func TestUnixNativePublicationRetentionAndAbsence(t *testing.T) {
	root := physicalTemp(t)
	parent := acquiredUnix(t, root).parent()
	old := unixTestPayload(t, parent, "run.json", []byte("old"))
	retained, err := unixRetainOriginal(old, parent, "run.json", "retained-original")
	if err != nil {
		t.Fatal(err)
	}
	defer retained.close()
	if _, err := unixRetainOriginal(old, parent, "run.json", "retained-original"); err == nil {
		t.Fatal("retention replaced existing object")
	}
	unixTestPayload(t, parent, "raw-backup", []byte("old"))
	unixTestPayload(t, parent, "payload", []byte("new"))
	if err := unixPublish(parent, "payload", "run.json", false); err == nil {
		t.Fatal("absence publication replaced competitor")
	}
	if got, _ := os.ReadFile(filepath.Join(root, "run.json")); string(got) != "old" {
		t.Fatal("failed no-replace changed original")
	}
	if err := unixPublish(parent, "payload", "run.json", true); err != nil {
		t.Fatal(err)
	}
	if err := unix.Fsync(parent.fd()); err != nil {
		t.Fatal(err)
	}
	if _, err := old.file.WriteAt([]byte("late"), 0); err != nil {
		t.Fatal(err)
	}
	if err := old.file.Sync(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"run.json": "new", "retained-original": "late", "raw-backup": "old"} {
		got, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(got) != want {
			t.Fatalf("%s: %q, %v", name, got, err)
		}
	}
	payload := unixTestPayload(t, parent, "payload2", []byte("absent"))
	if err := unixPublish(parent, "payload2", "new.json", false); err != nil {
		t.Fatal(err)
	}
	f, err := unixOpenDocument(parent, "new.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.close()
	if f.observation.stat.Ino != payload.observation.stat.Ino {
		t.Fatal("no-replace did not publish acquired payload")
	}
	if err := unix.Fsync(parent.fd()); err != nil {
		t.Fatal(err)
	}
}

func TestUnixNativePublicationConcurrentReaders(t *testing.T) {
	parent := acquiredUnix(t, physicalTemp(t)).parent()
	values := [][]byte{bytes.Repeat([]byte("a"), 32769), bytes.Repeat([]byte("b"), 65539)}
	unixTestPayload(t, parent, "run.json", values[0])
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
				f, err := unixOpenDocument(parent, "run.json")
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
		name := fmt.Sprintf("payload-%d", i)
		unixTestPayload(t, parent, name, values[i%2])
		if err := unixPublish(parent, name, "run.json", true); err != nil {
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
	if err := unix.Fsync(parent.fd()); err != nil {
		t.Fatal(err)
	}
}
