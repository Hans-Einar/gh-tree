package git

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func scopeFixture(t *testing.T, path string) repository {
	t.Helper()
	common, err := observeDirectory(path)
	if err != nil {
		t.Fatal(err)
	}
	id, err := domain.NewRepositoryID(domain.LocalCommon, path)
	if err != nil {
		t.Fatal(err)
	}
	return repository{id: id, common: common}
}

func TestMutationSchedulerSerializesAndBoundsWaiters(t *testing.T) {
	r := scopeFixture(t, t.TempDir())
	s := new(mutationScheduler)
	first, err := s.acquire(context.Background(), r.id)
	if err != nil {
		t.Fatal(err)
	}
	defer first()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results := make(chan error, 63)
	for i := 0; i < 63; i++ {
		go func() {
			release, err := s.acquire(ctx, r.id)
			if release != nil {
				release()
			}
			results <- err
		}()
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		s.mu.Lock()
		count := s.admitted
		s.mu.Unlock()
		if count == 64 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("waiters not admitted")
		}
		time.Sleep(time.Millisecond)
	}
	if release, err := s.acquire(context.Background(), r.id); err == nil {
		release()
		t.Fatal("unbounded admission")
	}
	cancel()
	for i := 0; i < 63; i++ {
		if err := <-results; !errors.Is(err, context.Canceled) {
			t.Fatalf("waiting cancellation: %v", err)
		}
	}
	first()
	s.mu.Lock()
	remaining, scopes := s.admitted, len(s.scopes)
	s.mu.Unlock()
	if remaining != 0 || scopes != 0 {
		t.Fatalf("retained scheduler records: %d/%d", remaining, scopes)
	}
	var active atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := s.acquire(context.Background(), r.id)
			if err != nil {
				t.Error(err)
				return
			}
			defer release()
			if active.Add(1) != 1 {
				t.Error("concurrent same-repository admission")
			}
			active.Add(-1)
		}()
	}
	wg.Wait()
}

func TestNativeCommonScopePermanentOwnershipAndCancellation(t *testing.T) {
	r := scopeFixture(t, t.TempDir())
	a, b := new(Adapter), new(Adapter)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if g, err := a.acquireMutationScope(ctx, r); err == nil {
		g.close()
		t.Fatal("canceled acquisition")
	}
	if _, err := os.Stat(filepath.Join(r.common.path, scopeLockName)); !os.IsNotExist(err) {
		t.Fatal("canceled acquisition created guard")
	}
	g, err := a.acquireMutationScope(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()
	if second, err := b.acquireMutationScope(context.Background(), r); err == nil {
		second.close()
		t.Fatal("independent adapter bypassed native lock")
	}
	other := scopeFixture(t, t.TempDir())
	g2, err := b.acquireMutationScope(context.Background(), other)
	if err != nil {
		t.Fatal("independent clone blocked", err)
	}
	if err := g2.close(); err != nil {
		t.Fatal(err)
	}
	before, err := g.file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if err := g.close(); err != nil {
		t.Fatal(err)
	}
	// Reobserve fallback change-stamp platforms after this operation's creation.
	r = scopeFixture(t, r.common.path)
	g3, err := b.acquireMutationScope(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	after, err := g3.file.Stat()
	if err != nil || !os.SameFile(before, after) {
		t.Fatal("guard was replaced")
	}
	if err := g3.close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(r.common.path, scopeLockName))
	if err != nil || string(data) != scopeLockMarker {
		t.Fatalf("guard marker changed: %q %v", data, err)
	}
}

func TestNativeCommonScopeRefusesForeignFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, scopeLockName)
	const content = "external original bytes\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	a := new(Adapter)
	if g, err := a.acquireMutationScope(context.Background(), scopeFixture(t, root)); err == nil {
		g.close()
		t.Fatal("foreign guard adopted")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != content {
		t.Fatal("foreign guard changed")
	}
}

func TestNativeCommonScopeReplacementDoesNotDeleteEitherObject(t *testing.T) {
	root := t.TempDir()
	g, err := new(Adapter).acquireMutationScope(context.Background(), scopeFixture(t, root))
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()
	path := filepath.Join(root, scopeLockName)
	retained := filepath.Join(root, "external-retained-lock")
	err = os.Rename(path, retained)
	if runtime.GOOS == "windows" {
		if err == nil {
			t.Fatal("Windows guard allowed replacement while held")
		}
		if err := g.validate(); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	const replacement = "newly appeared guard contents\n"
	if err := os.WriteFile(path, []byte(replacement), 0600); err != nil {
		t.Fatal(err)
	}
	if err := g.validate(); err == nil {
		t.Fatal("replacement not detected")
	}
	if err := g.close(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != replacement {
		t.Fatal("replacement deleted or overwritten")
	}
	if data, err := os.ReadFile(retained); err != nil || string(data) != scopeLockMarker {
		t.Fatal("held original deleted or overwritten")
	}
}

func TestNativeCommonScopeRefusesLink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "external")
	const content = "unrelated bytes\n"
	if err := os.WriteFile(target, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, scopeLockName)); err != nil {
		t.Skipf("native symlink privilege unavailable: %v", err)
	}
	if g, err := new(Adapter).acquireMutationScope(context.Background(), scopeFixture(t, root)); err == nil {
		g.close()
		t.Fatal("redirected lock followed")
	}
	if data, err := os.ReadFile(target); err != nil || string(data) != content {
		t.Fatal("external link target changed")
	}
}

// A separate native process owns the OS lock and reports readiness over a pipe.
// Killing it must release its OS lock while preserving the permanent file.
func TestNativeCommonScopeChild(t *testing.T) {
	root := os.Getenv("GH_TREE_SCOPE_CHILD")
	if root == "" {
		return
	}
	g, err := new(Adapter).acquireMutationScope(context.Background(), scopeFixture(t, root))
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()
	fmt.Fprintln(os.Stdout, "scope-held")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

func TestNativeCommonScopeProcessDeath(t *testing.T) {
	root := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestNativeCommonScopeChild$")
	cmd.Env = append(os.Environ(), "GH_TREE_SCOPE_CHILD="+root, "GORACE="+os.Getenv("GORACE")+" atexit_sleep_ms=0")
	hideWindow(cmd)
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	reader := bufio.NewReader(output)
	line, err := reader.ReadString('\n')
	if err != nil || line != "scope-held\n" {
		t.Fatalf("child readiness: %q %v", line, err)
	}
	r := scopeFixture(t, root)
	if g, err := new(Adapter).acquireMutationScope(context.Background(), r); err == nil {
		g.close()
		t.Fatal("cross-process lock ignored")
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("child was not killed")
	}
	g, err := new(Adapter).acquireMutationScope(context.Background(), scopeFixture(t, root))
	if err != nil {
		t.Fatal("crash did not release native ownership", err)
	}
	if err := g.close(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(root, scopeLockName)); err != nil || string(data) != scopeLockMarker {
		t.Fatalf("crash removed/changed permanent guard: %v", err)
	}
}

func TestNativeLinkedWorktreesUseOneCommonGuard(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			linked := filepath.Join(t.TempDir(), "linked")
			fixtureGit(t, a, root, "worktree", "add", "--detach", "--", linked, "HEAD")
			first := resolveFixture(t, a, root).Data()
			b, err := New(Options{GitExecutable: a.options.GitExecutable, CurrentDirectory: linked, Environment: a.options.Environment})
			if err != nil {
				t.Fatal(err)
			}
			second := resolveFixture(t, b, linked).Data()
			if first.Repository != second.Repository {
				t.Fatal("linked common identity mismatch")
			}
			r, err := a.registered(context.Background(), first.Repository)
			if err != nil {
				t.Fatal(err)
			}
			g, err := a.acquireMutationScope(context.Background(), r)
			if err != nil {
				t.Fatal(err)
			}
			defer g.close()
			r2, err := b.registered(context.Background(), second.Repository)
			if err != nil {
				t.Fatal(err)
			}
			if duplicate, err := b.acquireMutationScope(context.Background(), r2); err == nil {
				duplicate.close()
				t.Fatal("linked native guard bypassed")
			} else {
				var d api.Diagnostic
				if !errors.As(err, &d) || d.Data().Reason != "CommonMutationLocked" {
					t.Fatalf("wrong refusal: %v", err)
				}
			}
		})
	}
}
