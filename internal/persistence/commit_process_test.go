package persistence

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// This fixture owns no descendants or user configuration. Its only external
// state is the explicit parent-owned temporary store; the parent kills/joins it.
func TestCommitProtocolProcessFixture(t *testing.T) {
	root := os.Getenv("GH_TREE_TEST_COMMIT_ROOT")
	if root == "" {
		t.Skip("private process fixture")
	}
	s, err := New(context.Background(), Options{UserConfigPath: filepath.Join(root, "config.json"), PreferencesPath: filepath.Join(root, "state.json")})
	if err != nil {
		t.Fatal(err)
	}
	load, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := load.Observation().Data().Version.Value()
	s.hook = func(stage string) error {
		if stage == os.Getenv("GH_TREE_TEST_COMMIT_STOP") {
			fmt.Println("AT " + stage)
			for {
				time.Sleep(time.Hour)
			}
		}
		return nil
	}
	fmt.Println("READY")
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		t.Fatal(err)
	}
	result, commitErr := s.CommitUserConfig(context.Background(), userProposal(t, v, "child-proposal"))
	if !result.Valid() {
		t.Fatalf("invalid result: %v", commitErr)
	}
	if commitErr != nil && result.Data().Outcome != api.NotCommitted {
		t.Fatal(commitErr)
	}
	fmt.Printf("RESULT %d\n", result.Data().Outcome)
}

type commitChild struct {
	cmd    *exec.Cmd
	out    *bufio.Scanner
	stderr bytes.Buffer
	input  *os.File
	write  func(string) error
}

func startCommitChild(t testing.TB, root, stop string) *commitChild {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	child := &commitChild{cmd: exec.CommandContext(ctx, exe, "-test.run=^TestCommitProtocolProcessFixture$", "-test.timeout=20s")}
	child.cmd.Env = append(os.Environ(), "GH_TREE_TEST_COMMIT_ROOT="+root, "GH_TREE_TEST_COMMIT_STOP="+stop)
	child.cmd.Stderr = &child.stderr
	output, err := child.cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	input, err := child.cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	child.write = func(s string) error { _, err := input.Write([]byte(s)); return err }
	child.out = bufio.NewScanner(output)
	child.out.Buffer(make([]byte, 4096), 64*1024)
	if err := child.cmd.Start(); err != nil {
		input.Close()
		output.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		input.Close()
		if child.cmd.ProcessState == nil {
			child.cmd.Process.Kill()
			child.cmd.Wait()
		}
	})
	child.line(t, "READY")
	return child
}

func (c *commitChild) line(t testing.TB, prefix string) string {
	t.Helper()
	for c.out.Scan() {
		line := c.out.Text()
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("child missing %q: scan=%v stderr=%s", prefix, c.out.Err(), c.stderr.String())
	return ""
}

func TestCommitSeparateProcessesUseWholeDocumentCAS(t *testing.T) {
	root := physicalStoreTemp(t)
	first := startCommitChild(t, root, "")
	second := startCommitChild(t, root, "")
	if err := first.write("GO\n"); err != nil {
		t.Fatal(err)
	}
	won := first.line(t, "RESULT ")
	if won != fmt.Sprintf("RESULT %d", api.Committed) && won != fmt.Sprintf("RESULT %d", api.CommittedDurabilityUncertain) {
		t.Fatal(won)
	}
	if err := first.cmd.Wait(); err != nil {
		t.Fatalf("first child: %v %s", err, first.stderr.String())
	}
	if err := second.write("GO\n"); err != nil {
		t.Fatal(err)
	}
	if line := second.line(t, "RESULT "); line != fmt.Sprintf("RESULT %d", api.NotCommitted) {
		t.Fatal(line)
	}
	if err := second.cmd.Wait(); err != nil {
		t.Fatalf("second child: %v %s", err, second.stderr.String())
	}
	loaded, err := newTestStore(t, root).LoadUserConfig(context.Background())
	if err != nil || loaded.Observation().Data().State != api.ValidCurrent {
		t.Fatalf("winner/recovery lost: %v", err)
	}
}

func TestCommitProcessCrashBoundariesReleaseLockAndRetainFacts(t *testing.T) {
	for _, stage := range []string{"lock", "prepare.payload", "prepare.original.journal", "manifest-flushed", "before-publication", "native-return-lost", "directory-flush", "outcome-delivery"} {
		t.Run(stage, func(t *testing.T) {
			root := physicalStoreTemp(t)
			old := []byte(`{"schemaVersion":1,"stripPrefixes":["old"]}`)
			if err := os.WriteFile(filepath.Join(root, "config.json"), old, 0600); err != nil {
				t.Fatal(err)
			}
			child := startCommitChild(t, root, stage)
			if err := child.write("GO\n"); err != nil {
				t.Fatal(err)
			}
			child.line(t, "AT "+stage)
			if err := child.cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err := child.cmd.Wait(); err == nil {
				t.Fatal("killed child unexpectedly succeeded")
			}
			current, err := os.ReadFile(filepath.Join(root, "config.json"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeUserConfig(current); err != nil {
				t.Fatalf("crash left partial target: %v", err)
			}
			published := stage == "native-return-lost" || stage == "directory-flush" || stage == "outcome-delivery"
			if bytes.Equal(old, current) == published {
				t.Fatalf("wrong native boundary contents: %s", current)
			}
			loaded, err := newTestStore(t, root).LoadUserConfig(context.Background())
			if !loaded.Valid() || loaded.Observation().Data().State != api.ValidCurrent {
				t.Fatalf("crash load: %v %v", loaded, err)
			}
			if stage != "lock" && len(loaded.Observation().Data().Recovery) == 0 {
				t.Fatalf("crash lost all persisted recovery: %v", err)
			}
			// Reacquired lock and a completed load prove killed request ownership
			// was released. Restart facts never label the dead request committed.
		})
	}
}
