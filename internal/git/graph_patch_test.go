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

func patchLimits(t *testing.T, bytes uint64, files uint32) api.PatchLimits {
	t.Helper()
	v, err := api.NewPatchLimits(api.PatchLimitsData{MaxBytes: bytes, MaxFiles: files})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func stashPatch(t *testing.T, a *Adapter, stash domain.StashID, view api.StashPatchView, limits api.PatchLimits) api.ReadStashPatchResult {
	t.Helper()
	request, err := api.NewReadStashPatchRequest(api.ReadStashPatchRequestData{Stash: stash, View: view, Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ReadStashPatch(context.Background(), request)
	if err != nil {
		t.Fatalf("stash patch %T: %v", view, err)
	}
	return result
}

func TestStashPatchExactStagedWorktreeAndUntrackedAfterShift(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			head := exactHead(t, a, root)
			if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("index B\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fixtureGit(t, a, root, "add", "--", "file.txt")
			if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("worktree C\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "untracked.txt"), []byte("untracked U\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fixtureGit(t, a, root, "stash", "push", "-u", "-m", "Exact")
			selected := listFixtureStashes(t, a, head.Repository())[0].Data().ID
			if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("later\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fixtureGit(t, a, root, "stash", "push", "-m", "Later")
			// Read the selected committed attribute context even when the live
			// worktree subsequently declares the same text files binary.
			if err := os.WriteFile(filepath.Join(root, ".gitattributes"), []byte("*.txt -diff\n"), 0600); err != nil {
				t.Fatal(err)
			}
			baseIndex, _ := api.NewStashBaseToIndex(api.StashBaseToIndexData{})
			indexWorking, _ := api.NewStashIndexToWorktree(api.StashIndexToWorktreeData{})
			untracked, _ := api.NewStashUntracked(api.StashUntrackedData{})
			for _, test := range []struct {
				view     api.StashPatchView
				contains string
			}{{baseIndex, "+index B"}, {indexWorking, "+worktree C"}, {untracked, "+untracked U"}} {
				result := stashPatch(t, a, selected, test.view, patchLimits(t, 1<<20, 100))
				patch, p := result.Data().Patch.Value()
				if !p || !bytes.Contains(patch.Data().Bytes, []byte(test.contains)) || result.Data().Stash != selected {
					t.Fatalf("view content changed: %T", test.view)
				}
				comparison, p := result.Data().Comparison.Value()
				if !p || comparison.Data().Stash != selected.OID() {
					t.Fatal("stash comparison lost exact selected identity")
				}
			}
			// Fixture-only native deletion removes the original live occurrence;
			// its exact retained object remains independently readable without replay.
			fixtureGit(t, a, root, "stash", "drop", "stash@{1}")
			result := stashPatch(t, a, selected, indexWorking, patchLimits(t, 32, 1))
			patch, _ := result.Data().Patch.Value()
			if !patch.Data().Truncated || len(patch.Data().Bytes) > 32 || bytes.Contains(patch.Data().Bytes, []byte("[gh-tree")) {
				t.Fatal("bounded patch mutated native payload")
			}
		})
	}
}

func TestStashPatchRenameBinaryAndAbsentUntracked(t *testing.T) {
	root, a := nativeFixture(t, "sha256")
	seedCommit(t, a, root)
	head := exactHead(t, a, root)
	fixtureGit(t, a, root, "mv", "--", "file.txt", "é renamed.txt")
	if err := os.WriteFile(filepath.Join(root, "binary.bin"), []byte{0, 1, 2, 3, 4}, 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "add", "--", "binary.bin")
	fixtureGit(t, a, root, "stash", "push", "-m", "Rename and binary")
	selected := listFixtureStashes(t, a, head.Repository())[0].Data().ID
	view, _ := api.NewStashBaseToIndex(api.StashBaseToIndexData{})
	result := stashPatch(t, a, selected, view, patchLimits(t, 1<<20, 100))
	patch, _ := result.Data().Patch.Value()
	rename, binary := false, false
	for _, file := range patch.Data().Files {
		if file.Data().Kind == api.Renamed {
			old, p := file.Data().OldPath.Value()
			rename = p && old.String() == "file.txt" && file.Data().Path.String() == "é renamed.txt"
		}
		if file.Data().Path.String() == "binary.bin" {
			binary = file.Data().Binary && !file.Data().AddedLines.Present()
		}
	}
	if !rename || !binary {
		t.Fatalf("rename/binary facts lost: %+v", patch.Data().Files)
	}
	limited := stashPatch(t, a, selected, view, patchLimits(t, 1<<20, 1))
	lp, _ := limited.Data().Patch.Value()
	if !lp.Data().Truncated || len(lp.Data().Files) != 1 || bytes.Contains(lp.Data().Bytes, []byte("renamed.txt")) {
		t.Fatal("file-count bound did not constrain the native patch")
	}
	// Large caller uint32 limits cannot overflow an int on Windows386.
	_ = stashPatch(t, a, selected, view, patchLimits(t, 1<<20, ^uint32(0)))
	untracked, _ := api.NewStashUntracked(api.StashUntrackedData{})
	empty := stashPatch(t, a, selected, untracked, patchLimits(t, 1024, 10))
	p, _ := empty.Data().Patch.Value()
	comparison, _ := empty.Data().Comparison.Value()
	if len(p.Data().Bytes) != 0 || comparison.Data().UntrackedParent.Present() || comparison.Data().ToTree.Present() {
		t.Fatal("absent untracked view fabricated an endpoint")
	}
}

func TestGraphPaginationKeepsExactRootsRefsAndParents(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	first := exactHead(t, a, root)
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Second graph message")
	head := exactHead(t, a, root)
	filter, _ := api.NewReachableFromRoots(api.ReachableFromRootsData{})
	request, err := api.NewReadGraphRequest(api.ReadGraphRequestData{Repository: head.Repository(), Roots: []domain.Revision{head}, Filter: filter, Page: firstPage(t, 1)})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ReadGraph(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	d := result.Data()
	if len(d.Commits) != 1 || d.Commits[0].Data().Revision != head || len(d.Refs) == 0 || len(d.Heads) != 1 {
		t.Fatal("graph projections omitted exact native facts")
	}
	next, p := d.Page.Data().Next.Value()
	if !p {
		t.Fatal("missing graph continuation")
	}
	page, _ := api.NewPageRequest(api.PageRequestData{Limit: 1, Continuation: next})
	request, err = api.NewReadGraphRequest(api.ReadGraphRequestData{Repository: head.Repository(), Roots: []domain.Revision{head}, Filter: filter, Page: page})
	if err != nil {
		t.Fatal(err)
	}
	last, err := a.ReadGraph(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(last.Data().Commits) != 1 || last.Data().Commits[0].Data().Revision != first || len(last.Data().Commits[0].Data().Parents) != 0 {
		t.Fatal("graph continuation reset roots/parents")
	}
}
