package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func listFixtureStashes(t *testing.T, a *Adapter, repo domain.RepositoryID) []api.StashFact {
	t.Helper()
	request, err := api.NewListStashesRequest(api.ListStashesRequestData{Repository: repo, Page: firstPage(t, 100)})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ListStashes(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	return result.Data().Stashes
}

func TestStashFactsSurviveShiftAndPreserveAllParents(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			head := exactHead(t, a, root)
			if empty := listFixtureStashes(t, a, head.Repository()); len(empty) != 0 {
				t.Fatal("unborn stack not empty")
			}
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
			fixtureGit(t, a, root, "stash", "push", "-u", "-m", "gh-tree?branch=main&worktree=historical%20path")
			first := listFixtureStashes(t, a, head.Repository())
			if len(first) != 1 || len(first[0].Data().Parents) != 3 {
				t.Fatalf("stash parents: %+v", first)
			}
			if !first[0].Data().Origin.Present() {
				t.Fatal("legacy descriptive metadata lost")
			}
			if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("later\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fixtureGit(t, a, root, "stash", "push", "-m", "Later")
			next := listFixtureStashes(t, a, head.Repository())
			if len(next) != 2 || next[1].Data().ID != first[0].Data().ID || next[1].Data().Occurrence != first[0].Data().Occurrence || next[1].Data().DisplayPosition != 1 {
				t.Fatal("positional shift changed stable stash occurrence")
			}
			// Native normalized reflog read is exercised separately, providing the
			// read-only backend-neutral occurrence profile used for reftable.
			raw := fixtureGit(t, a, root, "log", "-g", "--date=raw", "--format=%H%x00%gn%x00%ge%x00%gD%x00%gs%x00", "refs/stash", "--")
			records, err := parseNativeStashLog(raw, head.OID().Format())
			if err != nil || len(records) != 2 {
				t.Fatal("native normalized log", err)
			}
			if records[0].new != first[0].Data().ID.OID() {
				t.Fatal("normalized log reordered identity")
			}
		})
	}
}

func TestDuplicateStashOIDsRemainDistinctObservableOccurrences(t *testing.T) {
	root, a := nativeFixture(t, "sha256")
	seedCommit(t, a, root)
	head := exactHead(t, a, root)
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("dirty\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "stash", "push", "-m", "Initial")
	initial := listFixtureStashes(t, a, head.Repository())[0]
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("different dirty\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, a, root, "stash", "push", "-m", "Different")
	other := listFixtureStashes(t, a, head.Repository())[0]
	a.options.Environment = append(a.options.Environment, "GIT_COMMITTER_DATE=1700000000 +0000")
	fixtureGit(t, a, root, "stash", "store", "-m", "Identical occurrence", initial.Data().ID.OID().String())
	fixtureGit(t, a, root, "stash", "store", "-m", "Intervening occurrence", other.Data().ID.OID().String())
	fixtureGit(t, a, root, "stash", "store", "-m", "Identical occurrence", initial.Data().ID.OID().String())
	stashes := listFixtureStashes(t, a, head.Repository())
	if len(stashes) != 5 {
		t.Fatal("duplicate OIDs collapsed")
	}
	if stashes[0].Data().ID != stashes[2].Data().ID || stashes[0].Data().Occurrence != stashes[2].Data().Occurrence || stashes[0].Data().DisplayPosition == stashes[2].Data().DisplayPosition {
		t.Fatal("indistinguishable duplicates must remain visible and ambiguous")
	}
}
