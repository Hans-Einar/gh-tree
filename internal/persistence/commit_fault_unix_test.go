//go:build linux || darwin || freebsd

package persistence

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

func TestUnixCommitNativePermissionFailures(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("ordinary-account permission control")
	}
	for _, stage := range []string{"prepare.manifest", "prepare.payload", "prepare.raw", "prepare.original", "prepare.publication", "before-publication"} {
		t.Run(stage, func(t *testing.T) {
			root := physicalStoreTemp(t)
			original := []byte(`{"schemaVersion":1,"stripPrefixes":["old"]}`)
			if err := os.WriteFile(filepath.Join(root, "config.json"), original, 0600); err != nil {
				t.Fatal(err)
			}
			// Retain the exact owned directory for restoration. It remains
			// searchable/readable; no denied subtree or pathname-based repair.
			parent, err := os.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			before, err := parent.Stat()
			if err != nil {
				parent.Close()
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := parent.Chmod(before.Mode()); err != nil {
					t.Error(err)
				}
				if err := parent.Close(); err != nil {
					t.Error(err)
				}
			})
			s := newTestStore(t, root)
			loaded, err := s.LoadUserConfig(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			version, _ := loaded.Observation().Data().Version.Value()
			hit := false
			s.hook = func(at string) error {
				if at == stage {
					hit = true
					if err := parent.Chmod(0500); err != nil {
						t.Fatal(err)
					}
				}
				return nil // the actual following native operation must fail
			}
			result, err := s.CommitUserConfig(context.Background(), userProposal(t, version, "new"))
			if restoreErr := parent.Chmod(before.Mode()); restoreErr != nil {
				t.Fatal(restoreErr)
			}
			if !hit || !errors.Is(err, unix.EACCES) || !result.Valid() || result.Data().Outcome != api.NotCommitted || result.Data().PublicationKnown || result.Data().CurrentVersion != loaded.Observation().Data().Version {
				t.Fatalf("native permission refusal lost facts: hit=%v result=%+v err=%v", hit, result.Data(), err)
			}
			current, err := os.ReadFile(filepath.Join(root, "config.json"))
			if err != nil || !bytes.Equal(current, original) {
				t.Fatalf("native failure changed original: %q %v", current, err)
			}
			restarted, err := newTestStore(t, root).LoadUserConfig(context.Background())
			if !restarted.Valid() || restarted.Observation().Data().Version != loaded.Observation().Data().Version {
				t.Fatalf("native failure stranded request lock or current: %v", err)
			}
		})
	}
}

func TestUnixCommitNativePartialWriteFixture(t *testing.T) {
	root := os.Getenv("GH_TREE_TEST_PARTIAL_WRITE_ROOT")
	if root == "" {
		t.Skip("private process file-size-limit fixture")
	}
	s := newTestStore(t, root)
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	version, _ := loaded.Observation().Data().Version.Value()
	checkResources := testRequestResources(t)
	// Only this joined child process receives the limit. The native write first
	// puts a real prefix on disk, then fails; no mock writer supplies the error.
	var limit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &limit); err != nil {
		t.Fatal(err)
	}
	limit.Cur = 40000
	if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &limit); err != nil {
		t.Fatal(err)
	}
	result, err := s.CommitUserConfig(context.Background(), userProposal(t, version, strings.Repeat("x", 200000)))
	if (!errors.Is(err, unix.EFBIG) && !errors.Is(err, io.ErrShortWrite)) || !result.Valid() || result.Data().Outcome != api.NotCommitted || result.Data().PublicationKnown || result.Data().CurrentVersion != loaded.Observation().Data().Version || len(result.Data().Recovery) != 1 {
		t.Fatalf("partial native write lost original/manifest facts: %+v %v", result.Data(), err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	partial := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".payload") {
			info, err := entry.Info()
			if err != nil || info.Size() != 40000 {
				t.Fatalf("native short write was not observed: %v %v", info, err)
			}
			partial = true
		}
	}
	if !partial {
		t.Fatal("native partial payload missing")
	}
	checkResources(t)
	fmt.Println("NATIVE_PARTIAL_WRITE=40000 NOT_COMMITTED MANIFEST_RETAINED RESOURCES_CLOSED")
}

func TestUnixCommitNativePartialWrite(t *testing.T) {
	root := physicalStoreTemp(t)
	original := []byte(`{"schemaVersion":1}`)
	if err := os.WriteFile(filepath.Join(root, "config.json"), original, 0600); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	child := exec.CommandContext(ctx, exe, "-test.run=^TestUnixCommitNativePartialWriteFixture$", "-test.timeout=12s")
	child.Env = append(os.Environ(), "GH_TREE_TEST_PARTIAL_WRITE_ROOT="+root)
	output, err := child.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "NATIVE_PARTIAL_WRITE=40000 NOT_COMMITTED MANIFEST_RETAINED RESOURCES_CLOSED") {
		t.Fatalf("native partial-write process: %v %s", err, output)
	}
	current, err := os.ReadFile(filepath.Join(root, "config.json"))
	if err != nil || !bytes.Equal(current, original) {
		t.Fatal("native partial write changed current bytes")
	}
	loaded, err := newTestStore(t, root).LoadUserConfig(context.Background())
	if !errors.Is(err, errIncompletePreparation) || !loaded.Valid() || loaded.Observation().Data().State != api.ValidCurrent || len(loaded.Observation().Data().Recovery) != 1 {
		t.Fatalf("restart lost independently proved original/manifest: %v", err)
	}
	t.Log(strings.TrimSpace(string(output)))
}
