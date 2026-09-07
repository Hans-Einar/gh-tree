package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestDirectoryChangeGuardReleaseOverflowAndRefusal(t *testing.T) {
	root := t.TempDir()
	t.Run("unchanged", func(t *testing.T) {
		w, e := openDirectoryWatch(root)
		if e != nil {
			t.Fatal(e)
		}
		if e := w.check(); e != nil {
			t.Fatal(e)
		}
		if e := w.close(); e != nil {
			t.Fatal(e)
		}
		if e := w.close(); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(root, "released"), nil, 0600); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("overflow", func(t *testing.T) {
		w, e := openDirectoryWatchBuffer(root, 4)
		if e != nil {
			t.Fatal(e)
		}
		defer w.close()
		if e := os.WriteFile(filepath.Join(root, "too-long-for-four-bytes"), nil, 0600); e != nil {
			t.Fatal(e)
		}
		if e := w.check(); e == nil {
			t.Fatal("overflow completed without invalidation")
		}
		if e := w.result(false); e != nil {
			t.Fatal(e)
		}
		if w.completedBytes != 0 {
			t.Fatalf("expected native zero-byte overflow, got %d", w.completedBytes)
		}
		if e := w.close(); e == nil {
			t.Fatal("overflow accepted during release")
		}
	})
	t.Run("unexpected-cancel", func(t *testing.T) {
		w, e := openDirectoryWatch(root)
		if e != nil {
			t.Fatal(e)
		}
		defer w.close()
		if e := syscall.CancelIoEx(w.handle, &w.request); e != nil {
			t.Fatal(e)
		}
		if e := w.result(true); !errors.Is(e, syscall.ERROR_OPERATION_ABORTED) {
			t.Fatal(e)
		}
		if e := w.close(); e == nil {
			t.Fatal("foreign early cancellation accepted")
		}
	})
	t.Run("unsupported", func(t *testing.T) {
		path := filepath.Join(root, "ordinary-file")
		os.WriteFile(path, nil, 0600)
		if w, e := openDirectoryWatch(path); e == nil {
			w.close()
			t.Fatal("ordinary file accepted as directory watcher")
		}
	})
	for i := 0; i < 16; i++ {
		w, e := openDirectoryWatch(root)
		if e != nil {
			t.Fatal(e)
		}
		path := filepath.Join(root, fmt.Sprintf("race-%d", i))
		started := make(chan struct{})
		done := make(chan error, 1)
		go func() { <-started; done <- os.WriteFile(path, nil, 0600) }()
		close(started)
		e = w.close()
		if e != nil && !strings.Contains(e.Error(), "invalidated") {
			t.Fatal(e)
		}
		if e := <-done; e != nil {
			t.Fatal(e)
		}
		if !w.released {
			t.Fatal("raced close retained native resources")
		}
		// Either our cancellation precedes the write, or completion invalidates.
		// A second close must be harmless in both cases.
		if e := w.close(); e != nil {
			t.Fatal(e)
		}
	}
}

func TestPartialDirectoryWatchAcquisitionUnwinds(t *testing.T) {
	b := []byte("recorded")
	f := captured{source: source{"repo/go.mod", hash(b), len(b)}, bytes: b, repoPath: "go.mod"}
	p := plan{files: map[string]captured{f.Path: f}, manifest: manifest{Sources: []source{f.source}}}
	p.manifest.SourceDigest = hash(jsonBytes(p.manifest.Sources))
	p.manifest.OptionsDigest = hash(jsonBytes(p.manifest.Options))
	p.manifest.ModuleDigest = hash(jsonBytes(p.manifest.Modules))
	root := t.TempDir()
	var acquired []*directoryWatch
	s, e := materializeWithWatch(p, root, func(path string) (*directoryWatch, error) {
		if len(acquired) == 2 {
			return nil, fmt.Errorf("injected unsupported directory guard")
		}
		w, e := openDirectoryWatch(path)
		if e == nil {
			acquired = append(acquired, w)
		}
		return w, e
	})
	if e == nil || !strings.Contains(e.Error(), "injected unsupported") {
		s.close()
		t.Fatalf("partial acquisition: %v", e)
	}
	if len(s.guards) != 0 || len(s.watches) != 0 {
		t.Fatal("partial ownership not unwound")
	}
	for _, w := range acquired {
		if !w.released {
			t.Fatal("pending native request after failed acquisition")
		}
	}
	if e := os.Rename(filepath.Join(root, "source"), filepath.Join(root, "released-source")); e != nil {
		t.Fatalf("partial cleanup retained directory handle: %v", e)
	}
}

func TestPostSelectionInputSetChangesRefuseActualBuild(t *testing.T) {
	canonicalTest(t)
	root, cache, proxy := moduleFixture(t)
	u := url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(proxy)}
	t.Setenv("GOMODCACHE", cache)
	t.Setenv("GOPROXY", u.String())
	t.Setenv("GOSUMDB", "off")
	writeFixture(t, root, "internal/runtime/broker/cmd/payload/recorded.txt", []byte("recorded payload"))
	writeFixture(t, root, "internal/runtime/broker/cmd/main_windows.go", []byte("package main\nimport (\"embed\";\"fmt\";\"example.com/selected\";\""+modulePath+"/internal/runtime/broker\")\n//go:embed payload/*.txt\nvar payload embed.FS\nfunc main(){items,_:=payload.ReadDir(\"payload\");fmt.Print(len(items),selected.Value,broker.ProtocolVersion)}\n"))
	if e := admit(root); e != nil {
		t.Fatal(e)
	}
	p, e := capture(root)
	if e != nil {
		t.Fatal(e)
	}
	const marker = "UNRECORDED_POST_SELECTION_FILESET_444444"
	content := []byte("package main\nimport \"fmt\"\nfunc init(){fmt.Print(\"" + marker + "\")}\n")
	for _, mode := range []string{"source", "embed", "hardlink", "rename", "prior-handle"} {
		t.Run(mode, func(t *testing.T) {
			var prior windows.Handle
			s, e := materializeWithWatch(p, t.TempDir(), func(path string) (*directoryWatch, error) {
				if mode == "prior-handle" && strings.HasSuffix(filepath.ToSlash(path), "source/internal/runtime/broker/cmd") {
					name, e := windows.UTF16PtrFromString(path)
					if e != nil {
						return nil, e
					}
					prior, e = windows.CreateFile(name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.WRITE_DAC, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
					if e != nil {
						return nil, e
					}
				}
				return openDirectoryWatch(path)
			})
			if prior != 0 {
				defer windows.CloseHandle(prior)
			}
			if e != nil {
				t.Fatal(e)
			}
			defer s.close()
			for _, arch := range arches {
				if e := s.verifySelection(p, arch); e != nil {
					t.Fatal(e)
				}
			}
			path := filepath.Join(s.root, "source", "internal", "runtime", "broker", "cmd", "unrecorded_windows.go")
			if mode == "embed" {
				path = filepath.Join(filepath.Dir(path), "payload", "unrecorded.txt")
			}
			staged := filepath.Join(s.root, "staged.go")
			if e := os.WriteFile(staged, content, 0600); e != nil {
				t.Fatal(e)
			}
			insert := func() error {
				switch mode {
				case "source", "embed":
					return os.WriteFile(path, content, 0600)
				case "hardlink":
					return os.Link(staged, path)
				case "rename":
					return os.Rename(staged, path)
				case "prior-handle":
					name, e := windows.NewNTUnicodeString(filepath.Base(path))
					if e != nil {
						return e
					}
					oa := windows.OBJECT_ATTRIBUTES{RootDirectory: prior, ObjectName: name, Attributes: windows.OBJ_CASE_INSENSITIVE}
					oa.Length = uint32(unsafe.Sizeof(oa))
					var h windows.Handle
					var status windows.IO_STATUS_BLOCK
					if e := windows.NtCreateFile(&h, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, &oa, &status, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0); e != nil {
						return e
					}
					f := os.NewFile(uintptr(h), path)
					_, e = f.Write(content)
					return errors.Join(e, f.Close())
				}
				return fmt.Errorf("unknown mutation")
			}
			if mode == "source" || mode == "embed" {
				arch := "amd64"
				if mode == "embed" {
					arch = "arm64"
				}
				args := append([]string{"build"}, buildFlags...)
				args = append(args, "-o", filepath.Join(s.root, arch+".exe"), entry)
				var mutationErr error
				started := false
				_, e = s.commandAfterStart(arch, func() {
					started = true
					mutationErr = insert()
					if mutationErr == nil {
						mutationErr = os.Remove(path)
					}
				}, args...)
				if !started || mutationErr != nil {
					t.Fatalf("actual child-start mutation: %v", mutationErr)
				}
				var toolFailure *exec.ExitError
				if errors.As(e, &toolFailure) {
					t.Fatalf("mutation control failed Go compilation rather than guard: %v", e)
				}
				if e == nil || !strings.Contains(e.Error(), "input directory invalidated") {
					t.Fatalf("post-start input change accepted: %v", e)
				}
			} else {
				if e := insert(); e != nil {
					t.Fatal(e)
				}
				if e := os.Remove(path); e != nil {
					t.Fatal(e)
				}
				for _, arch := range arches {
					args := append([]string{"build"}, buildFlags...)
					args = append(args, "-o", filepath.Join(s.root, arch+".exe"), entry)
					if _, e := s.command(arch, args...); e == nil || !strings.Contains(e.Error(), "input directory invalidated") {
						t.Fatalf("restored post-selection change accepted: %v", e)
					}
				}
			}
			if e := s.close(); e == nil {
				t.Fatal("release erased permanent invalidation")
			}
			if e := os.WriteFile(path, content, 0600); e != nil {
				t.Fatalf("released input-directory control: %v", e)
			}
		})
	}
	after, e := capture(root)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(jsonBytes(p.manifest), jsonBytes(after.manifest)) {
		t.Fatal("outer source provenance changed")
	}
}
