package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

var _ ports.RemoteFacts = (*Adapter)(nil)
var _ ports.RemoteMutations = (*Adapter)(nil)

func effects(state api.EffectState, obs api.Optional[api.RemoteObservation]) api.EffectReport {
	f := api.FacetEffectData{Facet: api.RemoteRefsPR, State: state}
	if o, p := obs.Value(); p {
		f.PostObservation = api.Some(o.Data().ID)
	}
	return must(api.NewEffectReport(api.EffectReportData{Facets: []api.FacetEffect{must(api.NewFacetEffect(f))}}))
}
func mergeTransport(a, b api.CommandTransportOutcome) api.CommandTransportOutcome {
	x, y := a.Data(), b.Data()
	x.Started = x.Started || y.Started
	x.RootReaped = (!a.Data().Started || x.RootReaped) && (!y.Started || y.RootReaped) && x.Started
	x.CleanupKnown = x.CleanupKnown && y.CleanupKnown
	x.StdoutTruncated = x.StdoutTruncated || y.StdoutTruncated
	x.StderrTruncated = x.StderrTruncated || y.StderrTruncated
	x.CancellationRequested = x.CancellationRequested || y.CancellationRequested
	x.Diagnostics = append(x.Diagnostics, y.Diagnostics...)
	// A single exit status cannot describe several distinct commands.
	if a.Data().Started && y.Started {
		x.ExitCode = api.None[int]()
	} else if y.Started {
		x.ExitCode = y.ExitCode
	}
	return must(api.NewCommandTransportOutcome(x))
}

// reobserveEndpoint establishes the still-associated repository and the exact
// published branch. The API cannot make these sequential reads a transaction.
func (a *Adapter) reobserveEndpoint(ctx context.Context, repo api.RemoteRepository, want api.EndpointExpectation) (api.CommandTransportOutcome, []api.Diagnostic, error) {
	r := a.request(ctx, repoPath(repo.Data().Locator), "GET", nil)
	transport := r.wire.transport
	diagnostics := r.diagnostics
	if r.err != nil {
		return transport, diagnostics, r.err
	}
	observed, e := parseRepository(r.body, a.config.Host)
	if e != nil || observed.Data().ID != repo.Data().ID {
		return transport, append(diagnostics, diagnostic(api.StaleObservation, "repository-association-changed")), errors.New("repository association changed")
	}
	path := repoPath(repo.Data().Locator) + "/branches/" + url.PathEscape(want.Data().Branch.Name())
	r = a.request(ctx, path, "GET", nil)
	transport = mergeTransport(transport, r.wire.transport)
	diagnostics = append(diagnostics, r.diagnostics...)
	if r.err != nil {
		return transport, diagnostics, r.err
	}
	b, revision, e := parseBranch(r.body, repo.Data().ID)
	if e != nil || b != want.Data().Branch || revision != want.Data().Revision {
		return transport, append(diagnostics, diagnostic(api.StaleObservation, "published-endpoint-changed-or-unresolved")), errors.New("published endpoint changed or unresolved")
	}
	return transport, diagnostics, nil
}

func (a *Adapter) CreatePullRequest(ctx context.Context, request api.CreatePullRequestRequest) (api.CreatePullRequestResult, error) {
	result := api.CreatePullRequestResultData{Transport: noTransport()}
	notSubmitted := func(d api.Diagnostic, cause error) (api.CreatePullRequestResult, error) {
		result.Outcome = must(api.NewNotSubmitted(api.NotSubmittedData{Reason: d}))
		result.Effects = effects(api.NotStarted, api.None[api.RemoteObservation]())
		result.Diagnostics = append(result.Diagnostics, d)
		if ctx != nil {
			result.CancellationRequested = ctx.Err() != nil
		}
		return must(api.NewCreatePullRequestResult(result)), errors.Join(cause, diagError(d))
	}
	if !request.Valid() {
		return notSubmitted(diagnostic(api.Invalid, "invalid-create-request"), nil)
	}
	d := request.Data()
	base, bok := a.lookup(d.Base.Data().Branch.Repository())
	head, hok := a.lookup(d.Head.Data().Branch.Repository())
	if !bok || !hok {
		return notSubmitted(diagnostic(api.Invalid, "unregistered-create-endpoint"), nil)
	}
	if base.Data().Locator.Data().Host != head.Data().Locator.Data().Host || base.Data().Locator.Data().Host != a.config.Host {
		return notSubmitted(diagnostic(api.Unsupported, "cross-host-pull-request"), nil)
	}
	if !base.Data().Capabilities.Data().CreatePullRequest {
		return notSubmitted(diagnostic(api.Unsupported, "repository-create-unavailable"), nil)
	}
	if e := checkContext(ctx); e != nil {
		return notSubmitted(diagnostic(api.Canceled, "create-canceled-before-submission"), e)
	}
	for i, repo := range []api.RemoteRepository{base, head} {
		want := d.Base
		if i == 1 {
			want = d.Head
		}
		transport, diagnostics, e := a.reobserveEndpoint(ctx, repo, want)
		result.Transport = mergeTransport(result.Transport, transport)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		if e != nil {
			return notSubmitted(diagnostic(api.StaleObservation, "create-endpoint-precondition-failed"), e)
		}
	}
	// This bounded existing-candidate query supplies evidence, never causal
	// ownership. Incomplete absence is not used as an exactly-once guarantee.
	filter := must(api.NewPullRequestFilter(api.PullRequestFilterData{State: api.FilterOpen, Base: api.Some(d.Base.Data().Branch), Head: api.Some(d.Head.Data().Branch)}))
	page := must(api.NewPageRequest(api.PageRequestData{Limit: 100, Continuation: must(api.NewInitialPage(api.InitialPageData{}))}))
	existing, e := a.ListPullRequests(ctx, must(api.NewListPullRequestsRequest(api.ListPullRequestsRequestData{Repository: base.Data().ID, Filter: filter, Page: page})))
	ed := existing.Data()
	result.Transport = mergeTransport(result.Transport, ed.Transport)
	result.Diagnostics = append(result.Diagnostics, ed.Diagnostics...)
	if len(ed.PullRequests) > 0 {
		result.Outcome = must(api.NewExistingCandidate(api.ExistingCandidateData{Candidates: ed.PullRequests, Page: ed.Page}))
		result.Observation = ed.Observation
		result.Effects = effects(api.NotStarted, api.None[api.RemoteObservation]())
		result.CancellationRequested = ctx.Err() != nil
		return must(api.NewCreatePullRequestResult(result)), e
	}
	if e != nil {
		return notSubmitted(diagnostic(api.Unavailable, "create-preflight-unavailable"), e)
	}
	if e := ctx.Err(); e != nil {
		return notSubmitted(diagnostic(api.Canceled, "create-canceled-before-submission"), e)
	}
	// Reobserve after the candidate query as close to send as this non-atomic
	// service allows. Never substitute a freshly observed tip for either request.
	for i, repo := range []api.RemoteRepository{base, head} {
		want := d.Base
		if i == 1 {
			want = d.Head
		}
		transport, diagnostics, e := a.reobserveEndpoint(ctx, repo, want)
		result.Transport = mergeTransport(result.Transport, transport)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		if e != nil {
			return notSubmitted(diagnostic(api.StaleObservation, "create-endpoint-precondition-failed"), e)
		}
	}
	payload := struct {
		Title      string `json:"title"`
		Body       string `json:"body"`
		Head       string `json:"head"`
		HeadRepo   string `json:"head_repo"`
		Base       string `json:"base"`
		Draft      bool   `json:"draft"`
		Maintainer bool   `json:"maintainer_can_modify"`
	}{d.Title, d.Body, head.Data().Locator.Data().Owner + ":" + d.Head.Data().Branch.Name(), head.Data().Locator.Data().Name, d.Base.Data().Branch.Name(), d.Draft, d.MaintainerCanModify}
	bytes, e := json.Marshal(payload)
	if e != nil {
		return notSubmitted(diagnostic(api.Invalid, "create-payload-invalid"), e)
	}
	r := a.request(ctx, repoPath(base.Data().Locator)+"/pulls", "POST", bytes)
	result.Transport = mergeTransport(result.Transport, r.wire.transport)
	result.Diagnostics = append(result.Diagnostics, r.diagnostics...)
	result.CancellationRequested = ctx.Err() != nil
	if !r.wire.transport.Data().Started {
		return notSubmitted(diagnostic(api.Canceled, "create-command-not-started"), r.err)
	}
	if explicitRejection(r) {
		reason := diagnostic(api.Conflict, "server-rejected-create")
		result.Outcome = must(api.NewRejectedNoCreation(api.RejectedNoCreationData{Reason: reason}))
		result.Effects = effects(api.VerifiedNoTargetChange, api.None[api.RemoteObservation]())
		result.Diagnostics = append(result.Diagnostics, reason)
		return must(api.NewCreatePullRequestResult(result)), errors.Join(r.err, diagError(reason))
	}
	version := a.version(base.Data().ID.Token() + "/create")
	obs := a.observe(base.Data().ID, r, must(api.NewPageInfo(api.PageInfoData{Returned: 1, Completeness: api.Unknown, Source: version})))
	if r.status == 201 && !r.wire.transport.Data().StdoutTruncated {
		created, parseErr := a.mapPR(r.body, base, obs)
		if parseErr == nil {
			// Retain causal 201 identity even if the command exits nonzero, context
			// cancels, or an independently bounded follow-up cannot refresh it.
			result.Observation = api.Some(obs)
			result.Effects = effects(api.AppliedVerified, result.Observation)
			follow, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.config.ReadTimeout)
			observed, followErr := a.ObservePullRequest(follow, must(api.NewObservePullRequestRequest(api.ObservePullRequestRequestData{Target: created.Data().ID, ExpectedBase: api.Some(d.Base.Data().Revision), ExpectedHead: api.Some(d.Head.Data().Revision)})))
			cancel()
			od := observed.Data()
			result.Transport = mergeTransport(result.Transport, od.Transport)
			result.Diagnostics = append(result.Diagnostics, od.Diagnostics...)
			responseMatched := endpointMatches(created.Data().Base, d.Base) && endpointMatches(created.Data().Head, d.Head)
			if actual, p := od.PullRequest.Value(); p {
				created = actual
				result.Observation = od.Observation
			}
			if responseMatched && followErr == nil && endpointMatches(created.Data().Base, d.Base) && endpointMatches(created.Data().Head, d.Head) {
				result.Outcome = must(api.NewCreatedVerified(api.CreatedVerifiedData{Created: created, RequestedBase: d.Base, RequestedHead: d.Head}))
			} else {
				reason := diagnostic(api.StaleObservation, "created-endpoint-drift-or-postcheck-unavailable")
				result.Outcome = must(api.NewCreatedWithDrift(api.CreatedWithDriftData{Created: created, RequestedBase: d.Base, RequestedHead: d.Head, Reason: reason}))
				result.Diagnostics = append(result.Diagnostics, reason)
				followErr = errors.Join(followErr, diagError(reason))
			}
			// Use the returned created observation, not a dangling earlier observation.
			result.Effects = effects(api.AppliedVerified, result.Observation)
			result.CancellationRequested = ctx.Err() != nil
			return must(api.NewCreatePullRequestResult(result)), errors.Join(r.err, followErr)
		}
	}
	evidence := api.RemoteCreateEvidenceData{OperationID: d.Operation, RequestedBase: d.Base, RequestedHead: d.Head, Interval: r.interval()}
	if id := r.headers.Get("X-Github-Request-Id"); id != "" && len(id) <= 256 {
		evidence.ProviderRequestID = api.Some(id)
	}
	// Partial identity is independently validated, never an arbitrary output URL.
	var returned struct {
		Number uint64 `json:"number"`
		URL    string `json:"html_url"`
	}
	if strictJSON(r.body, &returned) == nil && returned.Number > 0 && validURL(returned.URL, base.Data().Locator, returned.Number) {
		evidence.ReturnedID = api.Some(must(domain.NewPRID(base.Data().ID, returned.Number)))
		evidence.ReturnedURL = api.Some(repoURL(base.Data().Locator) + "/pull/" + fmt.Sprint(returned.Number))
	}
	reason := diagnostic(api.Indeterminate, "creation-may-have-committed-reconcile-before-retry")
	result.Outcome = must(api.NewCreationIndeterminate(api.CreationIndeterminateData{RequestEvidence: must(api.NewRemoteCreateEvidence(evidence)), Reason: reason}))
	result.Effects = effects(api.EffectIndeterminate, api.None[api.RemoteObservation]())
	result.Diagnostics = append(result.Diagnostics, reason)
	return must(api.NewCreatePullRequestResult(result)), errors.Join(r.err, diagError(reason))
}
func endpointMatches(endpoint api.PullRequestEndpoint, want api.EndpointExpectation) bool {
	e, ok := endpoint.(api.AvailableEndpoint)
	return ok && e.Data().Branch == want.Data().Branch && e.Data().Revision == want.Data().Revision
}
func explicitRejection(r response) bool {
	// A complete structured 4xx validation/auth response establishes rejection;
	// an exit code, localized stderr, 5xx or truncated body does not.
	if r.status != 400 && r.status != 401 && r.status != 403 && r.status != 404 && r.status != 422 && r.status != 429 {
		return false
	}
	if r.wire.transport.Data().StdoutTruncated {
		return false
	}
	var body struct {
		Message          *string `json:"message"`
		DocumentationURL *string `json:"documentation_url"`
	}
	return strictJSON(r.body, &body) == nil && body.Message != nil && body.DocumentationURL != nil
}
