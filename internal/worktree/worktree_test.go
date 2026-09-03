package worktree

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

type repositoryFixture struct {
	mainRoot   string
	targetRoot string
	remoteRoot string
	initialSHA string
	prSHA      string
	branch     string
}

func TestDeployMovesCleanTestWorktreeOnly(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, false)
	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	result, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.prSHA,
		TargetName:   "Concept1",
		TargetPath:   fixture.targetRoot,
		TargetBranch: fixture.branch,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SHA != fixture.prSHA || result.Branch != fixture.branch || result.PRNumber != 60 {
		t.Fatalf("result = %#v", result)
	}
	if got := git(t, fixture.targetRoot, "rev-parse", "HEAD"); got != fixture.prSHA {
		t.Fatalf("target HEAD = %s, want %s", got, fixture.prSHA)
	}
	if got := git(t, fixture.mainRoot, "rev-parse", "HEAD"); got != fixture.initialSHA {
		t.Fatalf("main worktree changed to %s, want %s", got, fixture.initialSHA)
	}
	if output := git(t, fixture.targetRoot, "status", "--porcelain=v1", "--untracked-files=all"); output != "" {
		t.Fatalf("target became dirty: %s", output)
	}
	command := exec.Command("git", "--git-dir", fixture.remoteRoot, "show-ref", "--verify", "refs/heads/"+fixture.branch)
	if err := command.Run(); err == nil {
		t.Fatalf("local target branch %q was unexpectedly pushed", fixture.branch)
	}
}

func TestDeployRefusesDirtyWorktreeWithoutDataLoss(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, false)
	dirtyPath := filepath.Join(fixture.targetRoot, "do-not-delete.txt")
	content := []byte("uncommitted and important\n")
	if err := os.WriteFile(dirtyPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	_, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.prSHA,
		TargetName:   "Concept1",
		TargetPath:   fixture.targetRoot,
		TargetBranch: fixture.branch,
	})
	if err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("expected dirty refusal, got %v", err)
	}
	got, readErr := os.ReadFile(dirtyPath)
	if readErr != nil || string(got) != string(content) {
		t.Fatalf("dirty data was not preserved: data=%q err=%v", got, readErr)
	}
	if gotSHA := git(t, fixture.targetRoot, "rev-parse", "HEAD"); gotSHA != fixture.initialSHA {
		t.Fatalf("dirty target moved to %s", gotSHA)
	}
}

func TestDeployMigratesCleanDetachedTargetToConfiguredBranch(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, true)
	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	result, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.prSHA,
		TargetName:   "Simulator",
		TargetPath:   fixture.targetRoot,
		TargetBranch: fixture.branch,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Branch != fixture.branch || git(t, fixture.targetRoot, "branch", "--show-current") != fixture.branch {
		t.Fatalf("detached target was not migrated: %#v", result)
	}
}

func TestDeployRefusesUnreferencedDetachedCommit(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, true)
	if err := os.WriteFile(filepath.Join(fixture.targetRoot, "detached.txt"), []byte("preserve this commit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, fixture.targetRoot, "add", "detached.txt")
	git(t, fixture.targetRoot, "commit", "-m", "unreferenced detached work")
	detachedSHA := git(t, fixture.targetRoot, "rev-parse", "HEAD")

	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	_, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.prSHA,
		TargetName:   "Simulator",
		TargetPath:   fixture.targetRoot,
		TargetBranch: fixture.branch,
	})
	if err == nil || !strings.Contains(err.Error(), "create a branch") {
		t.Fatalf("expected actionable detached-HEAD refusal, got %v", err)
	}
	if got := git(t, fixture.targetRoot, "rev-parse", "HEAD"); got != detachedSHA {
		t.Fatalf("unreferenced detached commit moved from %s to %s", detachedSHA, got)
	}
}

func TestDeployRefusesTargetBranchCheckedOutElsewhere(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, false)
	secondTarget := filepath.Join(filepath.Dir(fixture.mainRoot), "second-testbed")
	git(t, fixture.mainRoot, "worktree", "add", "-b", "local/second-test", secondTarget, fixture.initialSHA)

	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	_, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.prSHA,
		TargetName:   "second",
		TargetPath:   secondTarget,
		TargetBranch: fixture.branch,
	})
	if err == nil || !strings.Contains(err.Error(), "already checked out") {
		t.Fatalf("expected checked-out-branch refusal, got %v", err)
	}
}

func TestDeployRejectsChangedPRBeforeMovingTarget(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, false)
	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	_, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.initialSHA,
		TargetName:   "Concept1",
		TargetPath:   fixture.targetRoot,
		TargetBranch: fixture.branch,
	})
	if err == nil || !strings.Contains(err.Error(), "changed during deployment") {
		t.Fatalf("expected exact-SHA mismatch refusal, got %v", err)
	}
	if got := git(t, fixture.targetRoot, "rev-parse", "HEAD"); got != fixture.initialSHA {
		t.Fatalf("target moved despite SHA mismatch: %s", got)
	}
}

func TestDeployRefusesPrimaryWorktree(t *testing.T) {
	t.Parallel()
	fixture := setupRepository(t, false)
	manager := NewManager(process.ExecRunner{}, fixture.mainRoot)
	_, err := manager.Deploy(context.Background(), DeployRequest{
		PRNumber:     60,
		HeadSHA:      fixture.prSHA,
		TargetName:   "main",
		TargetPath:   fixture.mainRoot,
		TargetBranch: "local/never",
	})
	if err == nil || !strings.Contains(err.Error(), "primary worktree") {
		t.Fatalf("expected primary-worktree refusal, got %v", err)
	}
}

func TestParsePorcelain(t *testing.T) {
	t.Parallel()
	output := "worktree /repo/main\x00HEAD " + strings.Repeat("a", 40) + "\x00branch refs/heads/main\x00\x00" +
		"worktree /repo/tést bed\x00HEAD " + strings.Repeat("b", 40) + "\x00detached\x00\x00"
	infos, err := parsePorcelain(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 2 || infos[0].Branch != "main" || !infos[1].Detached || infos[1].Path != "/repo/tést bed" {
		t.Fatalf("infos = %#v", infos)
	}
}

func setupRepository(t *testing.T, detached bool) repositoryFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	base := t.TempDir()
	mainRoot := filepath.Join(base, "main")
	targetRoot := filepath.Join(base, "testbed")
	remoteRoot := filepath.Join(base, "origin.git")
	git(t, base, "init", "--bare", remoteRoot)
	git(t, base, "init", "-b", "main", mainRoot)
	git(t, mainRoot, "config", "user.name", "gh-tree test")
	git(t, mainRoot, "config", "user.email", "gh-tree@example.invalid")
	if err := os.WriteFile(filepath.Join(mainRoot, "payload.txt"), []byte("initial\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, mainRoot, "add", "payload.txt")
	git(t, mainRoot, "commit", "-m", "initial")
	initialSHA := git(t, mainRoot, "rev-parse", "HEAD")
	git(t, mainRoot, "remote", "add", "origin", remoteRoot)
	git(t, mainRoot, "push", "-u", "origin", "main")

	git(t, mainRoot, "switch", "-c", "feature/Concept1/ui-box")
	if err := os.WriteFile(filepath.Join(mainRoot, "payload.txt"), []byte("pull request\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, mainRoot, "add", "payload.txt")
	git(t, mainRoot, "commit", "-m", "pull request")
	prSHA := git(t, mainRoot, "rev-parse", "HEAD")
	git(t, mainRoot, "push", "origin", "HEAD:refs/pull/60/head")
	git(t, mainRoot, "switch", "main")

	branch := "local/concept1-test"
	if detached {
		git(t, mainRoot, "worktree", "add", "--detach", targetRoot, initialSHA)
	} else {
		git(t, mainRoot, "worktree", "add", "-b", branch, targetRoot, initialSHA)
	}
	return repositoryFixture{
		mainRoot:   mainRoot,
		targetRoot: targetRoot,
		remoteRoot: remoteRoot,
		initialSHA: initialSHA,
		prSHA:      prSHA,
		branch:     branch,
	}
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			t.Fatalf("git %v failed (%d): %s", args, exitError.ExitCode(), output)
		}
		t.Fatalf("git %v failed: %v: %s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}
