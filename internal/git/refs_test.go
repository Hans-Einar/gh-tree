package git

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func listFixtureRefs(t *testing.T, a *Adapter, repo domain.RepositoryID, page api.PageRequest) api.ListRefsResult {
	t.Helper()
	request, err := api.NewListRefsRequest(api.ListRefsRequestData{Repository: repo, Kinds: []api.RefKind{api.LocalBranchKind, api.LocalTagKind, api.CachedRemoteKind}, Page: page})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ListRefs(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRefsKeepLiteralBranchSuffixAndTagObject(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			root, a := nativeFixture(t, format)
			seedCommit(t, a, root)
			head := exactHead(t, a, root)
			fixtureGit(t, a, root, "update-ref", "refs/heads/refs/heads/literal", head.OID().String())
			fixtureGit(t, a, root, "tag", "-a", "v-fixture", "-m", "Annotation")
			result := listFixtureRefs(t, a, head.Repository(), firstPage(t, 100))
			seenBranch, seenTag := false, false
			for _, fact := range result.Data().Refs {
				switch locator := fact.Data().Locator.(type) {
				case api.LocalBranchRef:
					if locator.Data().Branch.Name() == "refs/heads/literal" {
						seenBranch = true
						if nativeRef(locator) != "refs/heads/refs/heads/literal" {
							t.Fatal("branch prefix stripped twice")
						}
					}
				case api.LocalTagRef:
					seenTag = true
					tag, p := fact.Data().TagObject.Value()
					revision, r := fact.Data().Revision.Value()
					if !p || !r || revision != head || tag == head.OID() {
						t.Fatal("annotated tag object and commit conflated")
					}
				}
			}
			if !seenBranch || !seenTag {
				t.Fatal("native ref facts omitted")
			}
		})
	}
}

func TestExactBranchRefusesMovementAndSameOIDWrongLocator(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	head := exactHead(t, a, root)
	branch, _ := domain.NewBranchID(head.Repository(), domain.Local, "main")
	target, _ := domain.NewBranchTarget(branch, head)
	request, _ := api.NewResolveExactRequest(api.ResolveExactRequestData{Repository: head.Repository(), Target: target})
	resolved, err := a.ResolveExact(context.Background(), request)
	if err != nil || !resolved.Data().Resolution.Present() {
		t.Fatal("exact branch failed", err)
	}
	fixtureGit(t, a, root, "update-ref", "refs/heads/other", head.OID().String())
	other, _ := domain.NewBranchID(head.Repository(), domain.Local, "other")
	locator, _ := api.NewLocalBranchRef(api.LocalBranchRefData{Branch: other})
	request, _ = api.NewResolveExactRequest(api.ResolveExactRequestData{Repository: head.Repository(), Target: target, Locator: api.Some[api.GitRefLocator](locator)})
	refused, err := a.ResolveExact(context.Background(), request)
	if err == nil || refused.Data().Resolution.Present() {
		t.Fatal("same OID substituted another selected branch")
	}
	fixtureGit(t, a, root, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Advance")
	request, _ = api.NewResolveExactRequest(api.ResolveExactRequestData{Repository: head.Repository(), Target: target})
	refused, err = a.ResolveExact(context.Background(), request)
	if err == nil || refused.Data().Resolution.Present() {
		t.Fatal("advanced ref replaced selected exact revision")
	}
}

func TestRefCursorRejectsChangedSource(t *testing.T) {
	root, a := nativeFixture(t, "sha1")
	seedCommit(t, a, root)
	head := exactHead(t, a, root)
	fixtureGit(t, a, root, "update-ref", "refs/heads/z", head.OID().String())
	first := listFixtureRefs(t, a, head.Repository(), firstPage(t, 1))
	next, p := first.Data().Page.Data().Next.Value()
	if !p {
		t.Fatal("missing ref continuation")
	}
	fixtureGit(t, a, root, "update-ref", "refs/heads/a", head.OID().String())
	page, _ := api.NewPageRequest(api.PageRequestData{Limit: 1, Continuation: next})
	request, err := api.NewListRefsRequest(api.ListRefsRequestData{Repository: head.Repository(), Kinds: []api.RefKind{api.LocalBranchKind, api.LocalTagKind, api.CachedRemoteKind}, Page: page})
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.ListRefs(context.Background(), request)
	if err == nil || len(result.Data().Refs) != 0 {
		t.Fatal("mutable ref cursor silently switched source")
	}
}

func TestRemoteExactNeedsObservedMappingAndCurrentBinding(t *testing.T) {
	root, a := nativeFixture(t, "sha256")
	seedCommit(t, a, root)
	head := exactHead(t, a, root)
	remote := filepath.Join(t.TempDir(), "remote.git")
	fixtureGit(t, a, root, "clone", "--bare", "--", root, remote)
	fixtureGit(t, a, root, "remote", "add", "fixture", remote)
	fixtureGit(t, a, root, "fetch", "fixture")
	refs := listFixtureRefs(t, a, head.Repository(), firstPage(t, 100))
	var locator api.CachedRemoteRef
	for _, fact := range refs.Data().Refs {
		if cached, p := fact.Data().Locator.(api.CachedRemoteRef); p && cached.Data().Ref == "refs/remotes/fixture/main" {
			locator = cached
			fresh, p := fact.Data().Freshness.Value()
			if !p || fresh.Data().Kind != api.Cached || fresh.Data().Generation.Present() {
				t.Fatal("cached native ref fabricated refreshed generation")
			}
		}
	}
	if !locator.Valid() {
		t.Fatal("cached remote mapping absent")
	}
	remoteRevision, _ := domain.NewRevision(locator.Data().Binding.Data().RemoteRepository, head.OID())
	target, _ := domain.NewCommitTarget(remoteRevision)
	request, _ := api.NewResolveExactRequest(api.ResolveExactRequestData{Repository: head.Repository(), Target: target})
	unbound, err := a.ResolveExact(context.Background(), request)
	if err == nil || unbound.Data().Resolution.Present() {
		t.Fatal("equal object bytes invented remote association")
	}
	request, _ = api.NewResolveExactRequest(api.ResolveExactRequestData{Repository: head.Repository(), Target: target, Locator: api.Some[api.GitRefLocator](locator)})
	bound, err := a.ResolveExact(context.Background(), request)
	if err != nil || !bound.Data().Resolution.Present() {
		t.Fatal("explicit native mapping failed", err)
	}
	fixtureGit(t, a, root, "config", "remote.fixture.pushurl", remote)
	stale, err := a.ResolveExact(context.Background(), request)
	if err == nil || stale.Data().Resolution.Present() {
		t.Fatal("configuration drift renewed stale binding")
	}
}
