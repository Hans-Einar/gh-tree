package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type endpointDTO struct {
	Repo json.RawMessage `json:"repo"`
	Ref  *string         `json:"ref"`
	SHA  *string         `json:"sha"`
}
type pullRequestDTO struct {
	Number     uint64          `json:"number"`
	HTMLURL    string          `json:"html_url"`
	Title      *string         `json:"title"`
	Body       json.RawMessage `json:"body"`
	State      *string         `json:"state"`
	Draft      *bool           `json:"draft"`
	Maintainer *bool           `json:"maintainer_can_modify"`
	Merged     *bool           `json:"merged"`
	MergedAt   *time.Time      `json:"merged_at"`
	UpdatedAt  *time.Time      `json:"updated_at"`
	Base       json.RawMessage `json:"base"`
	Head       json.RawMessage `json:"head"`
}

func (a *Adapter) endpoint(raw []byte) (api.PullRequestEndpoint, []api.Diagnostic) {
	un := api.UnavailableEndpointData{Reason: api.EndpointMissingField}
	bad := func(reason api.EndpointUnavailableReason) (api.PullRequestEndpoint, []api.Diagnostic) {
		un.Reason = reason
		return must(api.NewUnavailableEndpoint(un)), []api.Diagnostic{diagnostic(api.Unavailable, "pull-request-endpoint-unavailable")}
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return bad(api.EndpointMissingField)
	}
	var d endpointDTO
	if e := strictJSON(raw, &d); e != nil {
		return bad(api.EndpointUnresolved)
	}
	if len(d.Repo) == 0 {
		return bad(api.EndpointMissingField)
	}
	if bytes.Equal(d.Repo, []byte("null")) {
		return bad(api.EndpointDeleted)
	}
	repo, e := parseRepository(d.Repo, a.config.Host)
	if e != nil {
		return bad(api.EndpointUnresolved)
	}
	un.KnownRepository = api.Some(repo.Data().ID)
	var b domain.BranchID
	var revision domain.Revision
	if d.Ref != nil {
		b, e = domain.NewBranchID(repo.Data().ID, domain.RemoteHead, *d.Ref)
		if e == nil {
			un.KnownBranch = api.Some(b)
		}
	}
	if d.SHA != nil {
		revision, e = remoteRevision(repo.Data().ID, *d.SHA)
		if e == nil {
			un.KnownRevision = api.Some(revision)
		}
	}
	if !b.Valid() || !revision.Valid() {
		return bad(api.EndpointMissingField)
	}
	endpoint, e := api.NewAvailableEndpoint(api.AvailableEndpointData{Repository: repo, Branch: b, Revision: revision})
	if e != nil {
		return bad(api.EndpointUnresolved)
	}
	return endpoint, nil
}

func (a *Adapter) mapPR(raw []byte, repo api.RemoteRepository, obs api.RemoteObservation) (api.PullRequestFact, error) {
	var d pullRequestDTO
	if e := strictJSON(raw, &d); e != nil {
		return api.PullRequestFact{}, e
	}
	if d.Number == 0 || d.Title == nil || d.State == nil || d.Draft == nil || !validURL(d.HTMLURL, repo.Data().Locator, d.Number) {
		return api.PullRequestFact{}, protocolError("pull request required identity/fields")
	}
	base, bd := a.endpoint(d.Base)
	head, hd := a.endpoint(d.Head)
	diagnostics := append(bd, hd...)
	state := api.PRUnknown
	switch *d.State {
	case "open":
		state = api.PROpen
	case "closed":
		state = api.PRClosed
	}
	if d.MergedAt != nil || (d.Merged != nil && *d.Merged) {
		state = api.PRMerged
	}
	if state == api.PRUnknown {
		diagnostics = append(diagnostics, diagnostic(api.Unavailable, "unknown-pull-request-state"))
	}
	body := api.AbsentField[string]()
	if len(d.Body) > 0 {
		var text string
		if !bytes.Equal(d.Body, []byte("null")) {
			if e := strictJSON(d.Body, &text); e != nil {
				return api.PullRequestFact{}, e
			}
		}
		body = api.PresentField(text)
	}
	result := api.PullRequestFactData{ID: must(domain.NewPRID(repo.Data().ID, d.Number)), URL: repoURL(repo.Data().Locator) + "/pull/" + strconv.FormatUint(d.Number, 10), Title: *d.Title, Body: body, State: state, Draft: *d.Draft, Base: base, Head: head, Observation: obs, Diagnostics: diagnostics}
	if d.Maintainer != nil {
		result.MaintainerCanModify = api.Some(*d.Maintainer)
	}
	if d.UpdatedAt != nil {
		result.UpdatedAt = api.Some(d.UpdatedAt.UTC())
	}
	return api.NewPullRequestFact(result)
}

func (a *Adapter) ListPullRequests(ctx context.Context, request api.ListPullRequestsRequest) (api.ListPullRequestsResult, error) {
	scope := "invalid/pulls"
	if request.Valid() {
		d := request.Data()
		scope = d.Repository.Token() + "/pulls/created-asc/" + fmt.Sprint(d.Page.Data().Limit) + "/" + fingerprint(filterScope(d.Filter))
	}
	version := a.version(scope)
	fail := func(d api.Diagnostic) (api.ListPullRequestsResult, error) {
		return must(api.NewListPullRequestsResult(api.ListPullRequestsResultData{Page: unknownPage(version), Transport: noTransport(), Diagnostics: []api.Diagnostic{d}})), diagError(d)
	}
	if !request.Valid() {
		return fail(diagnostic(api.Invalid, "invalid-pull-request-list"))
	}
	d := request.Data()
	repo, ok := a.lookup(d.Repository)
	if !ok {
		return fail(diagnostic(api.Invalid, "unregistered-remote-scope"))
	}
	filter := d.Filter.Data()
	q := url.Values{"sort": {"created"}, "direction": {"asc"}, "per_page": {fmt.Sprint(d.Page.Data().Limit)}}
	state := "open"
	switch filter.State {
	case api.FilterClosed, api.FilterMerged:
		state = "closed"
	case api.FilterAll:
		state = "all"
	}
	q.Set("state", state)
	if b, p := filter.Base.Value(); p {
		if b.Repository() != d.Repository {
			return fail(diagnostic(api.Invalid, "foreign-base-filter"))
		}
		q.Set("base", b.Name())
	}
	if b, p := filter.Head.Value(); p {
		head, p := a.lookup(b.Repository())
		if !p {
			return fail(diagnostic(api.Invalid, "unregistered-head-filter"))
		}
		if head.Data().Locator.Data().Host != a.config.Host {
			return fail(diagnostic(api.Unsupported, "cross-host-filter"))
		}
		q.Set("head", head.Data().Locator.Data().Owner+":"+b.Name())
	}
	c, e := a.pageRequest(scope, d.Page)
	if e != nil {
		return fail(diagnostic(api.Invalid, "invalid-pull-request-cursor"))
	}
	if e := checkContext(ctx); e != nil {
		return fail(diagnostic(api.Canceled, "pull-request-list-canceled"))
	}
	q.Set("page", fmt.Sprint(c.page))
	path := repoPath(repo.Data().Locator) + "/pulls?" + q.Encode()
	r := a.request(ctx, path, "GET", nil)
	result := api.ListPullRequestsResultData{Page: unknownPage(version), Transport: r.wire.transport, Diagnostics: r.diagnostics}
	more, unknown := false, r.err != nil
	var facts []api.PullRequestFact
	if r.wire.transport.Data().Started {
		obs := a.observe(d.Repository, r, unknownPage(version))
		result.Observation = api.Some(obs)
		if r.status == 200 && !r.wire.transport.Data().StdoutTruncated {
			records, e := decodeList(r.body)
			if e == nil {
				more, e = nextPage(r, path, c.page, int(d.Page.Data().Limit), len(records))
			}
			if e != nil {
				unknown = true
				r.err = errors.Join(r.err, e)
				result.Diagnostics = append(result.Diagnostics, diagnostic(api.Unavailable, "invalid-pull-request-page"))
			}
			if len(records) <= int(d.Page.Data().Limit) {
				for _, raw := range records {
					pr, e := a.mapPR(raw, repo, obs)
					if e != nil {
						unknown = true
						result.Diagnostics = append(result.Diagnostics, diagnostic(api.Unavailable, "invalid-pull-request-record"))
						continue
					}
					matched, known := matchesFilter(pr, filter)
					if !known {
						unknown = true
						result.Diagnostics = append(result.Diagnostics, diagnostic(api.Unavailable, "unresolved-pull-request-filter"))
					}
					if !matched {
						continue
					}
					duplicate, conflict := seenFact(&c, fmt.Sprint(pr.Data().ID.Number()), semanticPR(pr))
					if duplicate {
						result.Diagnostics = append(result.Diagnostics, diagnostic(api.Conflict, "duplicate-pull-request-observation"))
						continue
					}
					if conflict {
						unknown = true
						result.Diagnostics = append(result.Diagnostics, diagnostic(api.StaleObservation, "conflicting-pull-request-page-observations"))
					}
					facts = append(facts, pr)
				}
			}
		} else {
			unknown = true
		}
	}
	page, diags := a.finishPage(c, version, len(facts), more, unknown)
	result.Page = page
	result.Diagnostics = append(result.Diagnostics, diags...)
	if old, p := result.Observation.Value(); p {
		od := old.Data()
		od.Page = page
		obs := must(api.NewRemoteObservation(od))
		result.Observation = api.Some(obs)
		for _, pr := range facts {
			pd := pr.Data()
			pd.Observation = obs
			result.PullRequests = append(result.PullRequests, must(api.NewPullRequestFact(pd)))
		}
	}
	return must(api.NewListPullRequestsResult(result)), r.err
}

func filterScope(f api.PullRequestFilter) string {
	d := f.Data()
	key := fmt.Sprint(d.State)
	for _, value := range []api.Optional[domain.BranchID]{d.Head, d.Base} {
		if b, p := value.Value(); p {
			key += fmt.Sprintf("/%q/%q", b.Repository().Token(), b.Name())
		} else {
			key += "/none"
		}
	}
	return key
}
func matchesFilter(pr api.PullRequestFact, f api.PullRequestFilterData) (bool, bool) {
	d := pr.Data()
	if d.State == api.PRUnknown && f.State != api.FilterAll {
		return false, false
	}
	if f.State == api.FilterOpen && d.State != api.PROpen || f.State == api.FilterMerged && d.State != api.PRMerged || f.State == api.FilterClosed && d.State != api.PRClosed {
		return false, true
	}
	for i, want := range []api.Optional[domain.BranchID]{f.Head, f.Base} {
		if b, p := want.Value(); p {
			endpoint := d.Head
			if i == 1 {
				endpoint = d.Base
			}
			available, ok := endpoint.(api.AvailableEndpoint)
			if !ok {
				return false, false
			}
			if available.Data().Branch != b {
				return false, true
			}
		}
	}
	return true, true
}
func semanticPR(pr api.PullRequestFact) string {
	d := pr.Data()
	body, p := d.Body.Value()
	return fmt.Sprintf("%q|%q|%d|%t|%v|%d:%q:%t|%s|%s|%v", d.URL, d.Title, d.State, d.Draft, d.MaintainerCanModify, d.Body.Presence(), body, p, semanticEndpoint(d.Base), semanticEndpoint(d.Head), d.UpdatedAt)
}
func semanticEndpoint(e api.PullRequestEndpoint) string {
	switch e := e.(type) {
	case api.AvailableEndpoint:
		d := e.Data()
		return fmt.Sprintf("available:%q:%q:%s", d.Repository.Data().ID.Token(), d.Branch.Name(), d.Revision.OID().String())
	case api.UnavailableEndpoint:
		return fmt.Sprintf("unavailable:%v", e.Data())
	}
	return "invalid"
}
func expectation(pr api.PullRequestFact, head, base api.Optional[domain.Revision]) api.ExpectationResult {
	if !head.Present() && !base.Present() {
		return api.NotRequested
	}
	unresolved := false
	for i, want := range []api.Optional[domain.Revision]{head, base} {
		if revision, p := want.Value(); p {
			endpoint := pr.Data().Head
			if i == 1 {
				endpoint = pr.Data().Base
			}
			available, ok := endpoint.(api.AvailableEndpoint)
			if !ok {
				unresolved = true
				continue
			}
			if available.Data().Revision != revision {
				return api.Mismatched
			}
		}
	}
	if unresolved {
		return api.ExpectationUnresolved
	}
	return api.Matched
}
func (a *Adapter) ObservePullRequest(ctx context.Context, request api.ObservePullRequestRequest) (api.ObservePullRequestResult, error) {
	fail := func(d api.Diagnostic) (api.ObservePullRequestResult, error) {
		return must(api.NewObservePullRequestResult(api.ObservePullRequestResultData{Expectation: api.ExpectationUnresolved, Transport: noTransport(), Diagnostics: []api.Diagnostic{d}})), diagError(d)
	}
	if !request.Valid() {
		return fail(diagnostic(api.Invalid, "invalid-pull-request-observation"))
	}
	d := request.Data()
	repo, ok := a.lookup(d.Target.Repository())
	if !ok {
		return fail(diagnostic(api.Invalid, "unregistered-remote-scope"))
	}
	if e := checkContext(ctx); e != nil {
		return fail(diagnostic(api.Canceled, "pull-request-observation-canceled"))
	}
	path := repoPath(repo.Data().Locator) + "/pulls/" + fmt.Sprint(d.Target.Number())
	r := a.request(ctx, path, "GET", nil)
	result := api.ObservePullRequestResultData{Expectation: api.ExpectationUnresolved, Transport: r.wire.transport, Diagnostics: r.diagnostics}
	if r.status == 200 && !r.wire.transport.Data().StdoutTruncated {
		page := must(api.NewPageInfo(api.PageInfoData{Returned: 1, Completeness: api.Complete, Source: a.version(d.Target.Repository().Token() + "/pull/" + fmt.Sprint(d.Target.Number()))}))
		obs := a.observe(d.Target.Repository(), r, page)
		pr, e := a.mapPR(r.body, repo, obs)
		if e == nil && pr.Data().ID != d.Target {
			e = protocolError("wrong pull request identity")
		}
		if e != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic(api.Unavailable, "invalid-pull-request-response"))
			r.err = errors.Join(r.err, e)
		} else {
			result.PullRequest = api.Some(pr)
			result.Observation = api.Some(obs)
			result.Expectation = expectation(pr, d.ExpectedHead, d.ExpectedBase)
			if result.Expectation == api.Mismatched || result.Expectation == api.ExpectationUnresolved {
				result.Diagnostics = append(result.Diagnostics, diagnostic(api.StaleObservation, "pull-request-endpoint-expectation-unmatched"))
				r.err = errors.Join(r.err, errors.New("requested endpoints unresolved or changed"))
			}
		}
	}
	if !result.PullRequest.Present() && r.err == nil {
		r.err = protocolError("pull request unavailable")
	}
	return must(api.NewObservePullRequestResult(result)), r.err
}
