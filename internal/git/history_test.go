package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func firstPage(t *testing.T, limit uint32) api.PageRequest {
	t.Helper()
	initial, _ := api.NewInitialPage(api.InitialPageData{})
	page, err := api.NewPageRequest(api.PageRequestData{Limit: limit, Continuation: initial})
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func exactHead(t *testing.T, a *Adapter, root string) domain.Revision {
	t.Helper()
	facts := resolveFixture(t, a, root).Data()
	for _, w := range facts.Worktrees {
		if w.Data().Primary {
			h, p := w.Data().Head.Value()
			if !p {
				t.Fatal("Head absent")
			}
			r, p := h.Revision()
			if !p {
				t.Fatal("Revision absent")
			}
			return r
		}
	}
	t.Fatal("primary missing")
	return domain.Revision{}
}

func TestExactHistoryPagesAndVerbatimMessages(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			first := exactHead(t, a, root)
			if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("second\n"), 0600); err != nil {
				t.Fatal(err)
			}
			fixtureGit(t, a, root, "add", "--", "file.txt")
			fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "-m", "Subject\n\nVerbatim body: [L:] é")
			second := exactHead(t, a, root)
			request, _ := api.NewReadCommitsRequest(api.ReadCommitsRequestData{Endpoint: second, Traversal: api.AllParents, Page: firstPage(t, 1)})
			result, err := a.ReadCommits(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			d := result.Data()
			if len(d.Commits) != 1 || d.Commits[0].Data().Revision != second || d.Commits[0].Data().Message != "Subject\n\nVerbatim body: [L:] é\n" || d.Page.Data().Completeness != api.More {
				t.Fatalf("history changed facts: %+v", d)
			}
			next, p := d.Page.Data().Next.Value()
			if !p {
				t.Fatal("missing continuation")
			}
			page, _ := api.NewPageRequest(api.PageRequestData{Limit: 1, Continuation: next})
			request, _ = api.NewReadCommitsRequest(api.ReadCommitsRequestData{Endpoint: second, Traversal: api.AllParents, Page: page})
			last, err := a.ReadCommits(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if len(last.Data().Commits) != 1 || last.Data().Commits[0].Data().Revision != first || len(last.Data().Commits[0].Data().Parents) != 0 || last.Data().Page.Data().Completeness != api.Complete {
				t.Fatal("root history incorrect")
			}
			request, _ = api.NewReadCommitsRequest(api.ReadCommitsRequestData{Endpoint: first, Traversal: api.AllParents, Page: page})
			mismatch, err := a.ReadCommits(context.Background(), request)
			if err == nil || len(mismatch.Data().Commits) != 0 {
				t.Fatal("cursor transplanted to different exact endpoint")
			}
			// Returned API slices cannot change an earlier snapshot or continuation.
			d.Commits[0] = api.CommitFact{}
			if !result.Data().Commits[0].Valid() {
				t.Fatal("history snapshot mutated")
			}
		})
	}
}

func TestMergeBaseRetainsExactEndpointsAndNoCommonAncestor(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			first := exactHead(t, a, root)
			fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Second")
			second := exactHead(t, a, root)
			request, _ := api.NewMergeBaseRequest(api.MergeBaseRequestData{Left: first, Right: second})
			result, err := a.MergeBase(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			out, p := result.Data().Outcome.Value()
			if !p {
				t.Fatal("missing merge base")
			}
			unique, ok := out.(api.UniqueMergeBase)
			if !ok || unique.Data().Base != first {
				t.Fatal("wrong exact merge base")
			}
			tree := line(fixtureGit(t, a, root, "rev-parse", "HEAD^{tree}"))
			rootOID := line(fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit-tree", tree, "-m", "Unrelated root"))
			oid, _ := domain.NewOID(rootOID)
			unrelated, _ := domain.NewRevision(first.Repository(), oid)
			request, _ = api.NewMergeBaseRequest(api.MergeBaseRequestData{Left: second, Right: unrelated})
			result, err = a.MergeBase(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			out, p = result.Data().Outcome.Value()
			if !p {
				t.Fatal("unrelated outcome missing")
			}
			if _, ok = out.(api.NoCommonAncestor); !ok {
				t.Fatal("unrelated history fabricated a base")
			}
		})
	}
}

func TestMergeBaseReturnsEveryCrissCrossCandidate(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			base := exactHead(t, a, root)
			tree := line(fixtureGit(t, a, root, "rev-parse", "HEAD^{tree}"))
			create := func(message string, parents ...string) domain.Revision {
				args := []string{"-c", "commit.gpgsign=false", "commit-tree", tree, "-m", message}
				for _, parent := range parents {
					args = append(args, "-p", parent)
				}
				text := line(fixtureGit(t, a, root, args...))
				oid, err := domain.NewOID(text)
				if err != nil {
					t.Fatal(err)
				}
				revision, err := domain.NewRevision(base.Repository(), oid)
				if err != nil {
					t.Fatal(err)
				}
				return revision
			}
			left := create("Left", base.OID().String())
			right := create("Right", base.OID().String())
			m1 := create("Merge1", left.OID().String(), right.OID().String())
			m2 := create("Merge2", right.OID().String(), left.OID().String())
			request, _ := api.NewMergeBaseRequest(api.MergeBaseRequestData{Left: m1, Right: m2})
			result, err := a.MergeBase(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			out, p := result.Data().Outcome.Value()
			if !p {
				t.Fatal("missing criss-cross outcome")
			}
			ambiguous, ok := out.(api.AmbiguousMergeBase)
			if !ok {
				t.Fatal("silently selected one merge base")
			}
			seen := map[domain.Revision]bool{}
			for _, r := range ambiguous.Data().Candidates {
				seen[r] = true
			}
			if len(seen) != 2 || !seen[left] || !seen[right] {
				t.Fatal("candidate set changed")
			}
		})
	}
}

func TestFactsTransportDeniesImplicitAcquisition(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	remote := filepath.Join(t.TempDir(), "remote.git")
	fixtureGit(t, a, root, "clone", "--bare", "--", root, remote)
	before := line(fixtureGit(t, a, root, "rev-parse", "HEAD"))
	r := a.command(context.Background(), root, false, "fetch", "--", remote)
	if r.err == nil {
		t.Fatal("facts transport allowed object acquisition")
	}
	if after := line(fixtureGit(t, a, root, "rev-parse", "HEAD")); after != before {
		t.Fatal("read-profile transport changed Head")
	}
}
