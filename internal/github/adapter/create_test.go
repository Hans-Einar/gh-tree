package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func prJSON(baseOwner, headOwner, headOID string) string {
	return fmt.Sprintf(`{"number":7,"html_url":"https://github.com/%s/project/pull/7","title":"literal title","body":"literal body","state":"open","draft":true,"maintainer_can_modify":true,"updated_at":"2026-09-06T00:00:00Z","base":{"repo":%s,"ref":"main","sha":%q},"head":{"repo":%s,"ref":"topic","sha":%q}}`, baseOwner, repositoryJSON(baseOwner, "project"), oidA, repositoryJSON(headOwner, "project"), headOID)
}
func endpointExpectation(a *Adapter, r api.RemoteRepository, branch, oid string) api.EndpointExpectation {
	response := response{wire: wire(200, "", "")}
	page := must(api.NewPageInfo(api.PageInfoData{Completeness: api.Complete, Source: a.version(r.Data().ID.Token() + "/fixture")}))
	obs := a.observe(r.Data().ID, response, page)
	return must(api.NewEndpointExpectation(api.EndpointExpectationData{Branch: must(domain.NewBranchID(r.Data().ID, domain.RemoteHead, branch)), Revision: must(remoteRevision(r.Data().ID, oid)), Observation: obs}))
}
func createFixture(t *testing.T) (*Adapter, api.CreatePullRequestRequest) {
	t.Helper()
	a := testAdapter(t)
	base := resolveFixture(t, a, "base", "project")
	head := resolveFixture(t, a, "fork", "project")
	r := must(api.NewCreatePullRequestRequest(api.CreatePullRequestRequestData{Operation: must(api.NewOperationID(1)), Base: endpointExpectation(a, base, "main", oidA), Head: endpointExpectation(a, head, "topic", oidB), Title: "title 日本語 $(do not run) &", Body: "body\n`literal` @- --arg", Draft: true, MaintainerCanModify: true}))
	return a, r
}
func scriptedCreate(t *testing.T, a *Adapter, request api.CreatePullRequestRequest, mode string, cancel context.CancelFunc) *int {
	t.Helper()
	posts := new(int)
	a.run = func(_ context.Context, c command) wireResult {
		if len(c.args) < 9 || c.args[0] != "api" || c.args[1] != "--hostname" || c.args[2] != "github.com" {
			t.Fatalf("bad argv %v", c.args)
		}
		method, path := c.args[4], c.args[8]
		if method == "POST" {
			*posts++
			var payload map[string]any
			if e := json.Unmarshal(c.input, &payload); e != nil {
				t.Fatal(e)
			}
			d := request.Data()
			if payload["head"] != "fork:topic" || payload["head_repo"] != "project" || payload["base"] != "main" || payload["title"] != d.Title || payload["body"] != d.Body || payload["draft"] != true || payload["maintainer_can_modify"] != true {
				t.Fatal(payload)
			}
			switch mode {
			case "rejected":
				return wire(422, "", `{"message":"localized error","documentation_url":"https://docs.github.com/rest/pulls/pulls"}`)
			case "lost":
				r := wire(0, "", "")
				r.stdout = nil
				r.err = errors.New("response lost")
				return r
			case "partial":
				return wire(201, "", `{"number":7,"html_url":"https://github.com/base/project/pull/7"}`)
			case "cancel":
				cancel()
				r := wire(201, "", prJSON("base", "fork", oidB))
				r.err = context.Canceled
				return r
			case "drift":
				return wire(201, "", prJSON("base", "fork", oidA))
			}
			return wire(201, "", prJSON("base", "fork", oidB))
		}
		switch {
		case strings.HasSuffix(path, "/pulls/7"):
			if mode == "refresh-failed" {
				return wire(500, "", `{"message":"server unavailable"}`)
			}
			oid := oidB
			if mode == "drift" {
				oid = oidA
			}
			return wire(200, "", prJSON("base", "fork", oid))
		case strings.Contains(path, "/pulls?"):
			if mode == "existing" {
				return wire(200, "", "["+prJSON("base", "fork", oidB)+"]")
			}
			return wire(200, "", "[]")
		case strings.HasSuffix(path, "/branches/main"):
			return wire(200, "", fmt.Sprintf(`{"name":"main","commit":{"sha":%q}}`, oidA))
		case strings.HasSuffix(path, "/branches/topic"):
			oid := oidB
			if mode == "pre-drift" {
				oid = oidA
			}
			return wire(200, "", fmt.Sprintf(`{"name":"topic","commit":{"sha":%q}}`, oid))
		case path == "repos/base/project":
			return wire(200, "", repositoryJSON("base", "project"))
		case path == "repos/fork/project":
			return wire(200, "", repositoryJSON("fork", "project"))
		default:
			t.Fatalf("unexpected request %s %s", method, path)
			return wireResult{}
		}
	}
	return posts
}
func TestCreateSixAlternativesAndFailurePreservation(t *testing.T) {
	for _, mode := range []string{"verified", "pre-drift", "existing", "rejected", "lost", "partial", "drift", "refresh-failed", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			a, req := createFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			posts := scriptedCreate(t, a, req, mode, cancel)
			r, e := a.CreatePullRequest(ctx, req)
			if !r.Valid() {
				t.Fatal("invalid result", e)
			}
			d := r.Data()
			state := d.Effects.Data().Facets[0].Data().State
			switch mode {
			case "verified":
				if _, ok := d.Outcome.(api.CreatedVerified); !ok || e != nil || state != api.AppliedVerified {
					t.Fatalf("%T %v %v", d.Outcome, state, e)
				}
			case "pre-drift":
				if _, ok := d.Outcome.(api.NotSubmitted); !ok || e == nil || state != api.NotStarted {
					t.Fatalf("%T %v", d.Outcome, e)
				}
			case "existing":
				if _, ok := d.Outcome.(api.ExistingCandidate); !ok || state != api.NotStarted {
					t.Fatalf("%T", d.Outcome)
				}
			case "rejected":
				if _, ok := d.Outcome.(api.RejectedNoCreation); !ok || state != api.VerifiedNoTargetChange {
					t.Fatalf("%T", d.Outcome)
				}
			case "lost", "partial":
				ind, ok := d.Outcome.(api.CreationIndeterminate)
				if !ok || state != api.EffectIndeterminate || e == nil {
					t.Fatalf("%T %v", d.Outcome, e)
				}
				if mode == "partial" && !ind.Data().RequestEvidence.Data().ReturnedID.Present() {
					t.Fatal("lost returned identity")
				}
			case "drift", "refresh-failed":
				if _, ok := d.Outcome.(api.CreatedWithDrift); !ok || state != api.AppliedVerified || e == nil {
					t.Fatalf("%T %v", d.Outcome, e)
				}
			case "cancel":
				if _, ok := d.Outcome.(api.CreatedVerified); !ok || state != api.AppliedVerified || e == nil || !d.CancellationRequested {
					t.Fatalf("%T %v", d.Outcome, e)
				}
			}
			expected := 1
			if mode == "existing" || mode == "pre-drift" {
				expected = 0
			}
			if *posts != expected {
				t.Fatal("unexpected replay", *posts)
			}
		})
	}
}
func TestPRMissingForkAndExactExpectations(t *testing.T) {
	a := testAdapter(t)
	base := resolveFixture(t, a, "base", "project")
	target := must(domain.NewPRID(base.Data().ID, 7))
	expected := must(remoteRevision(must(domain.NewRepositoryID(domain.Remote, "github.com/fork/project")), oidB))
	raw := prJSON("base", "fork", oidB)
	a.run = func(context.Context, command) wireResult { return wire(200, "", raw) }
	request := must(api.NewObservePullRequestRequest(api.ObservePullRequestRequestData{Target: target, ExpectedHead: api.Some(expected)}))
	r, e := a.ObservePullRequest(context.Background(), request)
	if e != nil || r.Data().Expectation != api.Matched {
		t.Fatal(r, e)
	}
	fact, _ := r.Data().PullRequest.Value()
	head := fact.Data().Head.(api.AvailableEndpoint)
	if head.Data().Branch.Repository() == base.Data().ID {
		t.Fatal("fork collapsed")
	}
	raw = strings.Replace(raw, `"head":{"repo":`+repositoryJSON("fork", "project"), `"head":{"repo":null`, 1)
	r, e = a.ObservePullRequest(context.Background(), request)
	if e == nil || r.Data().Expectation != api.ExpectationUnresolved {
		t.Fatal(r, e)
	}
	fact, _ = r.Data().PullRequest.Value()
	unknown := fact.Data().Head.(api.UnavailableEndpoint).Data()
	if unknown.KnownRepository.Present() || unknown.KnownRevision.Present() {
		t.Fatal("invented deleted fork scope")
	}
	if request.Data().ExpectedHead != api.Some(expected) {
		t.Fatal("expectation replaced")
	}
}
func TestConcurrentRemoteValuesAreIsolated(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "base", "project")
	a.run = func(context.Context, command) wireResult { return wire(200, "", "["+prJSON("base", "fork", oidB)+"]") }
	filter := must(api.NewPullRequestFilter(api.PullRequestFilterData{State: api.FilterOpen}))
	request := must(api.NewListPullRequestsRequest(api.ListPullRequestsRequestData{Repository: repo.Data().ID, Filter: filter, Page: must(api.NewPageRequest(api.PageRequestData{Limit: 100, Continuation: initial()}))}))
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := a.ListPullRequests(context.Background(), request)
			if e != nil || len(r.Data().PullRequests) != 1 {
				t.Error(e)
				return
			}
			data := r.Data()
			data.PullRequests[0] = api.PullRequestFact{}
			if !r.Data().PullRequests[0].Valid() {
				t.Error("mutable publication")
			}
		}()
	}
	wg.Wait()
}
