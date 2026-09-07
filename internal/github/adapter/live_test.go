package adapter

import (
	"context"
	"os"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// Explicit opt-in read-only CLI compatibility observation. Ordinary tests use
// controlled raw fixtures and never create a remote pull request.
func TestLiveReadOnly(t *testing.T) {
	locator := os.Getenv("GH_TREE_ADAPTER_LIVE_READ")
	if locator == "" {
		t.Skip("opt-in read-only remote observation")
	}
	a := testAdapter(t)
	l, e := ParseLocator("github.com", locator)
	if e != nil {
		t.Fatal(e)
	}
	resolved, e := a.ResolveRepository(context.Background(), must(api.NewResolveRepositoryRequest(api.ResolveRepositoryRequestData{Locator: l})))
	if e != nil {
		t.Fatal(e)
	}
	repo, ok := resolved.Data().Repository.Value()
	if !ok {
		t.Fatal("no repository")
	}
	branches, e := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 100, initial()))
	if e != nil {
		t.Fatal(e)
	}
	prs, e := a.ListPullRequests(context.Background(), must(api.NewListPullRequestsRequest(api.ListPullRequestsRequestData{Repository: repo.Data().ID, Filter: must(api.NewPullRequestFilter(api.PullRequestFilterData{State: api.FilterAll})), Page: must(api.NewPageRequest(api.PageRequestData{Limit: 100, Continuation: initial()}))})))
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("read-only repository %s; branches %d completeness %v; PRs %d completeness %v", repo.Data().ID.Token(), len(branches.Data().Branches), branches.Data().Page.Data().Completeness, len(prs.Data().PullRequests), prs.Data().Page.Data().Completeness)
	if len(prs.Data().PullRequests) > 0 {
		pr := prs.Data().PullRequests[0]
		observed, e := a.ObservePullRequest(context.Background(), must(api.NewObservePullRequestRequest(api.ObservePullRequestRequestData{Target: pr.Data().ID})))
		if e != nil || !observed.Data().PullRequest.Present() {
			t.Fatal(e)
		}
	}
}
