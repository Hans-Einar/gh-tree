package git

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func readFixtureDiff(t *testing.T, a *Adapter, comparison api.GitComparison) api.PatchFacts {
	t.Helper()
	request, err := api.NewReadDiffRequest(api.ReadDiffRequestData{Comparison: comparison, Limits: patchLimits(t, 1<<20, 100)})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ReadDiff(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	patch, p := result.Data().Patch.Value()
	if !p {
		t.Fatal("missing verified patch")
	}
	return patch
}

func TestDiffRootSelectedParentAndExactPair(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			first := exactHead(t, a, root)
			rootParent, _ := api.NewRootParent(api.RootParentData{})
			rootComparison, _ := api.NewCommitParentComparison(api.CommitParentComparisonData{Commit: first, Parent: rootParent})
			patch := readFixtureDiff(t, a, rootComparison)
			if !bytes.Contains(patch.Data().Bytes, []byte("+first")) {
				t.Fatal("root diff fabricated a commit parent")
			}
			if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("second\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fixtureGit(t, a, root, "add", "--", "file.txt")
			fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "-m", "Second")
			second := exactHead(t, a, root)
			selected, _ := api.NewSelectedParent(api.SelectedParentData{Index: 0})
			parentComparison, _ := api.NewCommitParentComparison(api.CommitParentComparisonData{Commit: second, Parent: selected})
			pairComparison, _ := api.NewCommitPairComparison(api.CommitPairComparisonData{From: first, To: second})
			parentPatch, pairPatch := readFixtureDiff(t, a, parentComparison), readFixtureDiff(t, a, pairComparison)
			if !bytes.Equal(parentPatch.Data().Bytes, pairPatch.Data().Bytes) || !bytes.Contains(pairPatch.Data().Bytes, []byte("+second")) {
				t.Fatal("selected parent/pair endpoints diverged")
			}
			invalid, _ := api.NewSelectedParent(api.SelectedParentData{Index: ^uint32(0)})
			comparison, _ := api.NewCommitParentComparison(api.CommitParentComparisonData{Commit: second, Parent: invalid})
			request, _ := api.NewReadDiffRequest(api.ReadDiffRequestData{Comparison: comparison, Limits: patchLimits(t, 1024, 10)})
			result, err := a.ReadDiff(context.Background(), request)
			if err == nil || result.Data().Patch.Present() {
				t.Fatal("oversized parent selection did not refuse")
			}
		})
	}
}

func TestDiffSeparatesIndexAndWorktreeVersions(t *testing.T) {
	root, a := nativeFixture(t, "sha256")
	seedCommit(t, a, root)
	id := primaryID(t, a, root)
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("index B\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "add", "--", "file.txt")
	if err := os.WriteFile(file, []byte("worktree C\n"), 0600); err != nil {
		t.Fatal(err)
	}
	status := observeFixtureStatus(t, a, id).Data()
	head, _ := status.Worktree.Data().Head.Value()
	staged, err := api.NewHeadToIndexComparison(api.HeadToIndexComparisonData{Worktree: id, Head: head, HeadVersion: status.Worktree.Data().Observation.Data().Version, Index: status.IndexVersion})
	if err != nil {
		t.Fatal(err)
	}
	working, err := api.NewIndexToWorktreeComparison(api.IndexToWorktreeComparisonData{Worktree: id, Index: status.IndexVersion, WorktreeVersion: status.WorktreeVersion})
	if err != nil {
		t.Fatal(err)
	}
	stagedPatch, workingPatch := readFixtureDiff(t, a, staged), readFixtureDiff(t, a, working)
	if !bytes.Contains(stagedPatch.Data().Bytes, []byte("+index B")) || bytes.Contains(stagedPatch.Data().Bytes, []byte("worktree C")) {
		t.Fatal("staged diff used working bytes")
	}
	if !bytes.Contains(workingPatch.Data().Bytes, []byte("+worktree C")) || !bytes.Contains(workingPatch.Data().Bytes, []byte("-index B")) {
		t.Fatal("working diff used Head instead of index")
	}
	if err := os.WriteFile(file, []byte("later D\n"), 0600); err != nil {
		t.Fatal(err)
	}
	request, _ := api.NewReadDiffRequest(api.ReadDiffRequestData{Comparison: working, Limits: patchLimits(t, 1024, 10)})
	stale, err := a.ReadDiff(context.Background(), request)
	if err == nil || stale.Data().Patch.Present() {
		t.Fatal("stale working diff substituted current bytes")
	}
	// Worktree edits do not change the exact Head/index comparison.
	_ = readFixtureDiff(t, a, staged)
}
