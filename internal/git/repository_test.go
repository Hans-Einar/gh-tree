package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func resolveFixture(t *testing.T, a *Adapter, path string) api.LocalRepositoryFacts {
	t.Helper()
	request, err := api.NewResolveLocalRequest(api.ResolveLocalRequestData{Locator: path})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ResolveLocal(context.Background(), request)
	if err != nil {
		t.Fatalf("ResolveLocal: %v; diagnostics %+v", err, result.Data().Diagnostics)
	}
	facts, ok := result.Data().Repository.Value()
	if !ok {
		t.Fatal("missing repository facts")
	}
	return facts
}

func fixtureGit(t *testing.T, a *Adapter, root string, args ...string) []byte {
	t.Helper()
	r := a.command(context.Background(), root, true, args...)
	if r.err != nil {
		t.Fatalf("fixture Git %q: %v, %s", args, r.err, r.stderr)
	}
	return r.stdout
}

func seedCommit(t *testing.T, a *Adapter, root string) {
	t.Helper()
	fixtureGit(t, a, root, "config", "user.name", "Fixture")
	fixtureGit(t, a, root, "config", "user.email", "fixture@example.invalid")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("first\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "add", "--", "file.txt")
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "-m", "First")
}

func TestResolveLocalUnbornAndCloneScopes(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			facts := resolveFixture(t, a, root).Data()
			if len(facts.Worktrees) != 1 {
				t.Fatalf("worktrees: %d", len(facts.Worktrees))
			}
			w := facts.Worktrees[0].Data()
			h, p := w.Head.Value()
			if !p || h.Kind() != domain.Unborn || !w.Primary || !w.Current || !w.Scope.Present() {
				t.Fatalf("unborn worktree: %+v", w)
			}
			if _, present := h.Revision(); present {
				t.Fatal("unborn HEAD exposed a Revision")
			}
			seedCommit(t, a, root)
			second := resolveFixture(t, a, root).Data()
			if second.Repository != facts.Repository || second.Worktrees[0].Data().ID != w.ID {
				t.Fatal("commit changed repository/worktree identity")
			}
			clone := filepath.Join(t.TempDir(), "clone")
			fixtureGit(t, a, root, "clone", "--", root, clone)
			cloned := resolveFixture(t, a, clone).Data()
			if cloned.Repository == facts.Repository {
				t.Fatal("clone scopes collapsed")
			}
		})
	}
}

func TestLinkedWorktreeIdentityMissingAndLocked(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			linked := filepath.Join(t.TempDir(), "linked é")
			fixtureGit(t, a, root, "worktree", "add", "--detach", "--", linked, "HEAD")
			facts := resolveFixture(t, a, linked).Data()
			if len(facts.Worktrees) != 2 {
				t.Fatalf("inventory count %d", len(facts.Worktrees))
			}
			var linkedID domain.WorktreeID
			for _, w := range facts.Worktrees {
				if !w.Data().Primary {
					linkedID = w.Data().ID
					h, p := w.Data().Head.Value()
					if !p || h.Kind() != domain.Detached || w.Data().Current {
						t.Fatalf("linked observation %+v", w.Data())
					}
				}
			}
			fixtureGit(t, a, root, "worktree", "lock", "--reason", "fixture lock", "--", linked)
			locked := resolveFixture(t, a, root).Data()
			found := false
			for _, w := range locked.Worktrees {
				if w.Data().ID == linkedID {
					_, found = w.Data().Availability.(api.LockedWorktree)
				}
			}
			if !found {
				t.Fatal("native lock not observed")
			}
			moved := linked + "-retained"
			if err := os.Rename(linked, moved); err != nil {
				t.Fatal(err)
			}
			missing := resolveFixture(t, a, root).Data()
			found = false
			for _, w := range missing.Worktrees {
				if w.Data().ID == linkedID {
					_, found = w.Data().Availability.(api.MissingWorktree)
					if w.Data().Scope.Present() {
						t.Fatal("missing worktree fabricated root")
					}
					if !w.Data().Head.Present() {
						t.Fatal("independent admin Head lost")
					}
				}
			}
			if !found {
				t.Fatal("missing registered identity lost")
			}
		})
	}
}

func TestRemoteBindingsSanitizeAndPreserveLiteralNames(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	fixtureGit(t, a, root, "remote", "add", "MiXeD.name", "https://fixture-user:fixture-secret@GitHub.com/Owner/Repo.git")
	facts := resolveFixture(t, a, root).Data()
	if len(facts.Remotes) != 1 {
		t.Fatalf("remote mappings %+v", facts.Remotes)
	}
	r := facts.Remotes[0].Data()
	if r.RemoteName != "MiXeD.name" || r.RemoteRepository.Token() != "github.com/owner/repo" || len(r.FetchURLs) != 1 || r.FetchURLs[0] != "https://github.com/owner/repo" {
		t.Fatalf("remote binding %+v", r)
	}
	fixtureGit(t, a, root, "remote", "set-url", "MiXeD.name", "https://github.com/Owner/Another.git")
	next := resolveFixture(t, a, root).Data().Remotes[0].Data()
	if next.RemoteRepository == r.RemoteRepository || next.Configuration == r.Configuration {
		t.Fatal("remote replacement reused identity/version")
	}
}
