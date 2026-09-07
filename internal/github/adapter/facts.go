package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func (a *Adapter) ResolveRepository(ctx context.Context, request api.ResolveRepositoryRequest) (api.ResolveRepositoryResult, error) {
	fail := func(d api.Diagnostic) (api.ResolveRepositoryResult, error) {
		return must(api.NewResolveRepositoryResult(api.ResolveRepositoryResultData{Transport: noTransport(), Diagnostics: []api.Diagnostic{d}})), diagError(d)
	}
	if !request.Valid() {
		return fail(diagnostic(api.Invalid, "invalid-repository-request"))
	}
	l := request.Data().Locator
	if !providerLocator(l) {
		return fail(diagnostic(api.Invalid, "unsupported-repository-components"))
	}
	if l.Data().Host != a.config.Host {
		return fail(diagnostic(api.Unsupported, "foreign-authenticated-host"))
	}
	if e := checkContext(ctx); e != nil {
		return fail(diagnostic(api.Canceled, "repository-request-canceled"))
	}
	r := a.request(ctx, repoPath(l), "GET", nil)
	d := api.ResolveRepositoryResultData{Transport: r.wire.transport, Diagnostics: r.diagnostics}
	if r.status == 200 && !r.wire.transport.Data().StdoutTruncated {
		repo, e := parseRepository(r.body, a.config.Host)
		if e == nil {
			// A verified redirect/rename produces an explicit new association and a
			// diagnostic. It never changes an existing scope's locator.
			if repo.Data().Locator != l {
				d.Diagnostics = append(d.Diagnostics, diagnostic(api.StaleObservation, "repository-renamed-or-transferred"))
			}
			e = a.register(repo)
			if e == nil {
				page := must(api.NewPageInfo(api.PageInfoData{Returned: 1, Completeness: api.Complete, Source: a.version(repo.Data().ID.Token() + "/repository")}))
				obs := a.observe(repo.Data().ID, r, page)
				d.Repository = api.Some(repo)
				d.Observation = api.Some(obs)
			}
		}
		if e != nil {
			d.Diagnostics = append(d.Diagnostics, diagnostic(api.Unavailable, "invalid-repository-response"))
			r.err = errors.Join(r.err, e)
		}
	}
	if !d.Repository.Present() && r.err == nil {
		r.err = protocolError("repository unavailable")
	}
	return must(api.NewResolveRepositoryResult(d)), r.err
}

func (a *Adapter) ListBranches(ctx context.Context, request api.ListBranchesRequest) (api.ListBranchesResult, error) {
	scope := "invalid/branches"
	if request.Valid() {
		d := request.Data()
		scope = d.Repository.Token() + "/branches/provider-default/" + fmt.Sprint(d.Page.Data().Limit)
		if p, ok := d.Filter.(api.RemoteBranchPrefix); ok {
			scope += "/prefix-sha256/" + fingerprint(p.Data().Prefix)
		}
	}
	version := a.version(scope)
	fail := func(d api.Diagnostic) (api.ListBranchesResult, error) {
		return must(api.NewListBranchesResult(api.ListBranchesResultData{Page: unknownPage(version), Transport: noTransport(), Diagnostics: []api.Diagnostic{d}})), diagError(d)
	}
	if !request.Valid() {
		return fail(diagnostic(api.Invalid, "invalid-branch-request"))
	}
	d := request.Data()
	repo, ok := a.lookup(d.Repository)
	if !ok {
		return fail(diagnostic(api.Invalid, "unregistered-remote-scope"))
	}
	c, e := a.pageRequest(scope, d.Page)
	if e != nil {
		return fail(diagnostic(api.Invalid, "invalid-branch-cursor"))
	}
	if e := checkContext(ctx); e != nil {
		return fail(diagnostic(api.Canceled, "branch-request-canceled"))
	}
	path := repoPath(repo.Data().Locator) + fmt.Sprintf("/branches?per_page=%d&page=%d", d.Page.Data().Limit, c.page)
	r := a.request(ctx, path, "GET", nil)
	result := api.ListBranchesResultData{Page: unknownPage(version), Transport: r.wire.transport, Diagnostics: r.diagnostics}
	type mapped struct {
		branch   domain.BranchID
		revision domain.Revision
	}
	var mappedFacts []mapped
	more, unknown := false, r.err != nil
	if r.status == 200 && !r.wire.transport.Data().StdoutTruncated {
		records, parseErr := decodeList(r.body)
		e = parseErr
		if e == nil {
			more, e = nextPage(r, path, c.page, int(d.Page.Data().Limit), len(records))
		}
		if e != nil {
			unknown = true
			result.Diagnostics = append(result.Diagnostics, diagnostic(api.Unavailable, "invalid-branch-page"))
			r.err = errors.Join(r.err, e)
		}
		if records != nil && len(records) <= int(d.Page.Data().Limit) {
			for _, record := range records {
				b, tip, e := parseBranch(record, d.Repository)
				if e != nil {
					unknown = true
					result.Diagnostics = append(result.Diagnostics, diagnostic(api.Unavailable, "invalid-branch-record"))
					continue
				}
				if p, ok := d.Filter.(api.RemoteBranchPrefix); ok && !strings.HasPrefix(b.Name(), p.Data().Prefix) {
					continue
				}
				duplicate, conflict := seenFact(&c, b.Name(), tip.OID().String())
				if duplicate {
					result.Diagnostics = append(result.Diagnostics, diagnostic(api.Conflict, "duplicate-branch-observation"))
					continue
				}
				if conflict {
					unknown = true
					result.Diagnostics = append(result.Diagnostics, diagnostic(api.StaleObservation, "conflicting-branch-page-observations"))
				}
				mappedFacts = append(mappedFacts, mapped{b, tip})
			}
		}
	} else {
		unknown = true
	}
	page, diagnostics := a.finishPage(c, version, len(mappedFacts), more, unknown)
	result.Page = page
	result.Diagnostics = append(result.Diagnostics, diagnostics...)
	if r.wire.transport.Data().Started {
		obs := a.observe(d.Repository, r, page)
		result.Observation = api.Some(obs)
		for _, f := range mappedFacts {
			result.Branches = append(result.Branches, must(api.NewRemoteBranchFact(api.RemoteBranchFactData{Branch: f.branch, Tip: f.revision, Observation: obs})))
		}
	}
	return must(api.NewListBranchesResult(result)), r.err
}
