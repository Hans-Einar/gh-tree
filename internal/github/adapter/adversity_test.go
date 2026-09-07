package adapter

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestMoreThan100BranchesAndForeignLink(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "owner", "project")
	calls := 0
	a.run = func(context.Context, command) wireResult {
		calls++
		if calls == 1 {
			rows := make([]string, 100)
			for i := range rows {
				rows[i] = fmt.Sprintf(`{"name":"branch-%d","commit":{"sha":%q}}`, i, oidA)
			}
			return wire(200, "Link: <https://api.github.com/repos/owner/project/branches?page=2&per_page=100>; rel=\"next\"\r\n", "["+strings.Join(rows, ",")+"]")
		}
		return wire(200, "", fmt.Sprintf(`[{"name":"last","commit":{"sha":%q}}]`, oidB))
	}
	r, e := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 100, initial()))
	if e != nil || len(r.Data().Branches) != 100 || r.Data().Page.Data().Completeness != api.More {
		t.Fatal(r, e)
	}
	next, _ := r.Data().Page.Data().Next.Value()
	last, e := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 100, next))
	if e != nil || len(last.Data().Branches) != 1 || last.Data().Page.Data().Completeness != api.Complete {
		t.Fatal(last, e)
	}
	oldObs, _ := r.Data().Observation.Value()
	newObs, _ := last.Data().Observation.Value()
	if oldObs.Data().ID == newObs.Data().ID || oldObs.Data().Version == newObs.Data().Version {
		t.Fatal("page provenance merged")
	}
	a.run = func(context.Context, command) wireResult {
		return wire(200, "Link: <https://evil.example/repos/owner/project/branches?page=2&per_page=100>; rel=\"next\"\r\n", fmt.Sprintf(`[{"name":"good","commit":{"sha":%q}}]`, oidA))
	}
	r, e = a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 100, initial()))
	if e == nil || len(r.Data().Branches) != 1 || r.Data().Page.Data().Completeness != api.Unknown || r.Data().Page.Data().Next.Present() {
		t.Fatal("foreign link trusted", r, e)
	}
}
func TestProviderFormatAndMalformedPartialPR(t *testing.T) {
	a := testAdapter(t)
	base := resolveFixture(t, a, "base", "project")
	if len(base.Data().Capabilities.Data().SupportedObjectFormats) != 1 {
		t.Fatal("invented provider format")
	}
	if _, e := remoteRevision(base.Data().ID, strings.Repeat("1", 64)); e == nil {
		t.Fatal("provider assumed SHA256")
	}
	good := prJSON("base", "fork", oidB)
	bad := strings.Replace(good, `"number":7`, `"number":0`, 1)
	deleted := strings.Replace(good, `"head":{"repo":`+repositoryJSON("fork", "project"), `"head":{"repo":null`, 1)
	deleted = strings.ReplaceAll(deleted, `"number":7`, `"number":8`)
	deleted = strings.ReplaceAll(deleted, `/pull/7`, `/pull/8`)
	a.run = func(context.Context, command) wireResult { return wire(200, "", "["+good+","+bad+","+deleted+"]") }
	filter := must(api.NewPullRequestFilter(api.PullRequestFilterData{State: api.FilterAll}))
	r, e := a.ListPullRequests(context.Background(), must(api.NewListPullRequestsRequest(api.ListPullRequestsRequestData{Repository: base.Data().ID, Filter: filter, Page: must(api.NewPageRequest(api.PageRequestData{Limit: 100, Continuation: initial()}))})))
	if e != nil || len(r.Data().PullRequests) != 2 || r.Data().Page.Data().Completeness != api.Unknown {
		t.Fatal(r, e)
	}
}
func TestCreateTruncatedAndCanceledAdmission(t *testing.T) {
	a, request := createFixture(t)
	posts := scriptedCreate(t, a, request, "verified", func() {})
	run := a.run
	a.run = func(ctx context.Context, c command) wireResult {
		w := run(ctx, c)
		if c.mutation {
			d := w.transport.Data()
			d.StdoutTruncated = true
			w.transport = must(api.NewCommandTransportOutcome(d))
		}
		return w
	}
	r, e := a.CreatePullRequest(context.Background(), request)
	if _, ok := r.Data().Outcome.(api.CreationIndeterminate); !ok || e == nil || *posts != 1 {
		t.Fatalf("%T %v %d", r.Data().Outcome, e, *posts)
	}
	a, request = createFixture(t)
	posts = scriptedCreate(t, a, request, "verified", func() {})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, e = a.CreatePullRequest(ctx, request)
	if _, ok := r.Data().Outcome.(api.NotSubmitted); !ok || e == nil || *posts != 0 || !r.Data().CancellationRequested {
		t.Fatal("canceled admission sent mutation")
	}
}
func TestUnknownServerFailureNeverProvesNoCreation(t *testing.T) {
	for _, status := range []int{200, 201, 301, 400, 401, 403, 404, 422, 429, 500, 502} {
		r := response{status: status, wire: wire(status, "", "")}
		r.body = []byte(`{"message":"localized only"}`)
		if explicitRejection(r) {
			t.Fatal("unstructured rejection", status)
		}
	}
}

func TestLocatorTraversalRefusesBeforeInvocation(t *testing.T) {
	a := testAdapter(t)
	a.run = func(context.Context, command) wireResult { t.Fatal("invalid locator invoked"); return wireResult{} }
	for _, pair := range [][2]string{{"..", "project"}, {"owner", ".."}, {"owner", "%2e%2e"}, {"owner", "."}} {
		l, e := api.NewRemoteRepositoryLocator(api.RemoteRepositoryLocatorData{Host: "github.com", Owner: pair[0], Name: pair[1]})
		if e != nil {
			continue
		}
		r, e := a.ResolveRepository(context.Background(), must(api.NewResolveRepositoryRequest(api.ResolveRepositoryRequestData{Locator: l})))
		if e == nil || r.Data().Transport.Data().Started {
			t.Fatal("traversal admitted")
		}
	}
}
