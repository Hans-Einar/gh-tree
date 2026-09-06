package git

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func primaryID(t *testing.T, a *Adapter, root string) domain.WorktreeID {
	t.Helper()
	facts := resolveFixture(t, a, root).Data()
	for _, w := range facts.Worktrees {
		if w.Data().Primary {
			return w.Data().ID
		}
	}
	t.Fatal("primary unavailable")
	return domain.WorktreeID{}
}
func observeFixtureStatus(t *testing.T, a *Adapter, id domain.WorktreeID) api.StatusFacts {
	t.Helper()
	request, err := api.NewObserveStatusRequest(api.ObserveStatusRequestData{Worktree: id})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ObserveStatus(context.Background(), request)
	if err != nil {
		t.Fatalf("status: %v, diagnostics %+v", err, result.Data().Diagnostics)
	}
	facts, p := result.Data().Status.Value()
	if !p {
		t.Fatal("missing status facts")
	}
	if facts.Data().Observation.Data().Completeness != api.Complete {
		t.Fatalf("incomplete status: %+v", result.Data().Diagnostics)
	}
	return facts
}
func causesFor(status api.StatusFacts, name string) map[api.ChangeCause]api.ChangeFact {
	result := map[api.ChangeCause]api.ChangeFact{}
	for _, row := range status.Data().Changes {
		if row.Data().Path.String() == name {
			result[row.Data().Cause] = row
		}
	}
	return result
}

func TestStatusSeparatesStagedUnstagedAndBothWithoutIndexWrites(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			id := primaryID(t, a, root)
			if clean := observeFixtureStatus(t, a, id); len(clean.Data().Changes) != 0 {
				t.Fatal("clean repository invented changes")
			}
			path := filepath.Join(root, "file.txt")
			if err := os.WriteFile(path, []byte("unstaged\n"), 0600); err != nil {
				t.Fatal(err)
			}
			status := observeFixtureStatus(t, a, id)
			causes := causesFor(status, "file.txt")
			if len(causes) != 1 || !causes[api.WorktreeChangeCause].Valid() {
				t.Fatal("unstaged cause lost")
			}
			fixtureGit(t, a, root, "add", "--", "file.txt")
			status = observeFixtureStatus(t, a, id)
			causes = causesFor(status, "file.txt")
			if len(causes) != 1 || !causes[api.IndexChangeCause].Valid() {
				t.Fatal("staged cause lost")
			}
			if err := os.WriteFile(path, []byte("both causes\n"), 0600); err != nil {
				t.Fatal(err)
			}
			indexPath := filepath.Join(root, ".git", "index")
			before, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			status = observeFixtureStatus(t, a, id)
			causes = causesFor(status, "file.txt")
			if len(causes) != 2 || !causes[api.IndexChangeCause].Valid() || !causes[api.WorktreeChangeCause].Valid() {
				t.Fatal("independent same-path causes collapsed")
			}
			after, err := os.ReadFile(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			now, err := os.Stat(indexPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) || !os.SameFile(info, now) || !info.ModTime().Equal(now.ModTime()) {
				t.Fatal("status refreshed/replaced the native index")
			}
		})
	}
}

func TestStatusRenameDeletionAndUntrackedReplacement(t *testing.T) {
	root, a := nativeFixture(t, "sha256")
	seedCommit(t, a, root)
	id := primaryID(t, a, root)
	fixtureGit(t, a, root, "mv", "--", "file.txt", "é renamed.txt")
	if err := os.WriteFile(filepath.Join(root, "é renamed.txt"), []byte("edited after rename\n"), 0600); err != nil {
		t.Fatal(err)
	}
	status := observeFixtureStatus(t, a, id)
	causes := causesFor(status, "é renamed.txt")
	if len(causes) != 2 || causes[api.IndexChangeCause].Data().Kind != api.Renamed || causes[api.WorktreeChangeCause].Data().Kind != api.Modified {
		t.Fatal("staged rename plus working edit changed meaning")
	}
	old, p := causes[api.IndexChangeCause].Data().OldPath.Value()
	if !p || old.String() != "file.txt" {
		t.Fatal("rename source lost")
	}
	if err := os.Remove(filepath.Join(root, "é renamed.txt")); err != nil {
		t.Fatal(err)
	}
	status = observeFixtureStatus(t, a, id)
	if causesFor(status, "é renamed.txt")[api.WorktreeChangeCause].Data().Kind != api.Deleted {
		t.Fatal("worktree deletion after rename lost")
	}
	fixtureGit(t, a, root, "rm", "--cached", "--", "é renamed.txt")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("untracked replacement\n"), 0600); err != nil {
		t.Fatal(err)
	}
	status = observeFixtureStatus(t, a, id)
	causes = causesFor(status, "file.txt")
	if len(causes) != 2 || causes[api.IndexChangeCause].Data().Kind != api.Deleted || causes[api.UntrackedChangeCause].Data().Kind != api.Untracked {
		t.Fatal("staged deletion plus untracked replacement lost")
	}
	if _, present := causes[api.IndexChangeCause].Data().WorktreeState.(api.PresentFile); !present {
		t.Fatal("staged deletion fabricated absent live bytes")
	}
}

func TestStatusIntentToAddAndPreservedMtimeContentVersion(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	id := primaryID(t, a, root)
	if err := os.WriteFile(filepath.Join(root, "intent.txt"), []byte("intent\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "add", "-N", "--", "intent.txt")
	status := observeFixtureStatus(t, a, id)
	found := false
	for _, row := range status.Data().Changes {
		if row.Data().Path.String() == "intent.txt" {
			for _, entry := range row.Data().IndexEntries {
				for _, flag := range entry.Data().SemanticFlags {
					found = found || flag == api.IntentToAdd
				}
			}
		}
	}
	if !found {
		t.Fatal("native intent-to-add semantic flag lost")
	}
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("BBBBB\n"), 0600); err != nil {
		t.Fatal(err)
	}
	before := observeFixtureStatus(t, a, id)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("CCCCC\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	after := observeFixtureStatus(t, a, id)
	if before.Data().WorktreeVersion == after.Data().WorktreeVersion {
		t.Fatal("same-size preserved-mtime edit escaped the full content version")
	}
}

func TestStatusConflictKeepsNativeStagesAndUnbornNoRevision(t *testing.T) {
	root, a := nativeFixture(t, "sha256")
	id := primaryID(t, a, root)
	unborn := observeFixtureStatus(t, a, id)
	head, p := unborn.Data().Worktree.Data().Head.Value()
	if !p || head.Kind() != domain.Unborn {
		t.Fatal("unborn status fabricated a revision")
	}
	seedCommit(t, a, root)
	fixtureGit(t, a, root, "checkout", "-b", "other")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("other\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "add", "--", "file.txt")
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "-m", "Other")
	fixtureGit(t, a, root, "checkout", "main")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "add", "--", "file.txt")
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "-m", "Main")
	if result := a.command(context.Background(), root, true, "merge", "other"); result.err == nil {
		t.Fatal("fixture did not conflict")
	}
	status := observeFixtureStatus(t, a, id)
	causes := causesFor(status, "file.txt")
	row := causes[api.ConflictChangeCause]
	if len(causes) != 1 || !row.Valid() || len(row.Data().IndexEntries) != 3 {
		t.Fatal("conflict stage/cause facts lost")
	}
}

func TestStatusUpstreamNoneResolvedAheadDivergedGoneAndUnresolved(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	id := primaryID(t, a, root)
	if _, ok := observeFixtureStatus(t, a, id).Data().Upstream.(api.NoUpstream); !ok {
		t.Fatal("unconfigured upstream is not explicit none")
	}
	remote := filepath.Join(t.TempDir(), "remote.git")
	fixtureGit(t, a, root, "clone", "--bare", "--", root, remote)
	fixtureGit(t, a, root, "remote", "add", "origin", remote)
	fixtureGit(t, a, root, "fetch", "origin")
	fixtureGit(t, a, root, "branch", "--set-upstream-to=origin/main", "main")
	checkCounts := func(ahead, behind uint64) {
		t.Helper()
		upstream, ok := observeFixtureStatus(t, a, id).Data().Upstream.(api.ResolvedUpstream)
		if !ok {
			t.Fatal("upstream not resolved")
		}
		ad, ap := upstream.Data().Comparison.Data().Ahead.Value()
		bd, bp := upstream.Data().Comparison.Data().Behind.Value()
		if !ap || !bp || ad != ahead || bd != behind || upstream.Data().Freshness.Data().Kind != api.Cached {
			t.Fatalf("counts/freshness %d %d", ad, bd)
		}
	}
	checkCounts(0, 0)
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Local advance")
	checkCounts(1, 0)
	other := filepath.Join(t.TempDir(), "other")
	fixtureGit(t, a, root, "clone", "--", remote, other)
	fixtureGit(t, a, other, "config", "user.name", "Fixture")
	fixtureGit(t, a, other, "config", "user.email", "fixture@example.invalid")
	fixtureGit(t, a, other, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Remote advance")
	fixtureGit(t, a, other, "push", "origin", "main")
	fixtureGit(t, a, root, "fetch", "origin")
	checkCounts(1, 1)
	fixtureGit(t, a, root, "update-ref", "-d", "refs/remotes/origin/main")
	if _, ok := observeFixtureStatus(t, a, id).Data().Upstream.(api.GoneUpstream); !ok {
		t.Fatal("conclusive missing cache is not gone")
	}
	fixtureGit(t, a, root, "config", "--unset-all", "remote.origin.url")
	request, _ := api.NewObserveStatusRequest(api.ObserveStatusRequestData{Worktree: id})
	result, err := a.ObserveStatus(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	facts, p := result.Data().Status.Value()
	if !p {
		t.Fatal("known status lost on unresolved upstream")
	}
	if _, ok := facts.Data().Upstream.(api.UnresolvedUpstream); !ok {
		t.Fatal("failed binding became none/gone/resolved")
	}
	if facts.Data().Observation.Data().Completeness != api.Partial || len(result.Data().Diagnostics) == 0 {
		t.Fatal("unresolved upstream lacked incomplete diagnostics")
	}
}
