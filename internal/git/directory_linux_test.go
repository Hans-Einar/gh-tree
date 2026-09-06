//go:build linux

package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// This native mechanism control exercises the future ref-command scope gap:
// a common-directory descriptor must continue naming the acquired repository
// after its old pathname is replaced by a clone with the same expected OID.
// It is not yet a production mutation port or the full hook/publication proof.
func TestNativeGitRefCommandAcquiredDirectoryControl(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	old := line(fixtureGit(t, a, root, "rev-parse", "HEAD"))
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "New root object")
	next := line(fixtureGit(t, a, root, "rev-parse", "HEAD"))
	fixtureGit(t, a, root, "update-ref", "refs/heads/main", old, next)
	fixtureGit(t, a, root, "update-ref", "refs/heads/fixture-new", next)
	parent := t.TempDir()
	clone := filepath.Join(parent, "clone")
	fixtureGit(t, a, root, "clone", "--no-local", "--", root, clone)
	// Preserve the new object in the clone too, so an accidental pathname CAS
	// could otherwise succeed against its identical old endpoint.
	observed, err := observeDirectory(filepath.Join(root, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	directory, err := acquireDirectory(observed)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.close()
	retained := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-retained")
	if err := os.Rename(root, retained); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(retained)
	if err := os.Rename(clone, root); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.options.GitExecutable, "--git-dir=/proc/self/fd/3", "update-ref", "refs/heads/main", next, old)
	cmd.Dir = parent
	cmd.ExtraFiles = []*os.File{directory.file}
	cmd.Env = append(commandEnvironment(a.options.Environment), "GIT_COMMON_DIR=/proc/self/fd/3", "GIT_IMPLICIT_WORK_TREE=0")
	cmd.WaitDelay = 100 * time.Millisecond
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("native acquired-directory ref command: %v %s", err, output)
	}
	if actual := line(fixtureGit(t, a, retained, "rev-parse", "HEAD")); actual != next {
		t.Fatal("native command lost its acquired original repository")
	}
	if actual := line(fixtureGit(t, a, root, "rev-parse", "HEAD")); actual != old {
		t.Fatal("native command changed the same-OID replacement clone")
	}
}
