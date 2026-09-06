package git

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type refsSnapshot struct {
	facts       []api.RefFact
	byName      map[string]api.RefFact
	remotes     []api.RemoteBinding
	observation api.GitObservation
	fingerprint string
	diagnostics []api.Diagnostic
}

func (s *readSession) refs(repo repository) (refsSnapshot, error) {
	result := refsSnapshot{byName: make(map[string]api.RefFact)}
	remoteBindings, diags, err := s.remoteBindings(repo)
	if err != nil {
		return result, err
	}
	result.remotes = remoteBindings
	result.diagnostics = diags
	config := s.command(repo.common.path, "--git-dir="+repo.common.path, "config", "--null", "--list", "--show-origin", "--show-scope")
	if config.err != nil {
		return result, config.err
	}
	q := s.command(repo.common.path, "--git-dir="+repo.common.path, "for-each-ref", "--sort=refname", "--format=%(refname)%00%(objectname)%00%(objecttype)%00%(*objectname)%00%(*objecttype)%00%(symref)%00", "--", "refs/heads/", "refs/tags/", "refs/remotes/")
	if q.err != nil {
		return result, q.err
	}
	result.fingerprint = queryBinding(string(q.stdout), string(config.stdout))
	version := sourceVersion("refs", repo.id.Token(), s.a.lifetime, []byte(result.fingerprint))
	completeness := api.Complete
	if len(diags) > 0 {
		completeness = api.Partial
	}
	result.observation, err = s.observation(repo.id, api.None[domain.WorktreeID](), version, completeness)
	if err != nil {
		return result, err
	}
	for _, row := range bytes.Split(bytes.TrimSuffix(q.stdout, []byte{'\n'}), []byte{'\n'}) {
		if len(row) == 0 {
			continue
		}
		fields := bytes.Split(row, []byte{0})
		if len(fields) != 7 {
			return result, diagnostic(api.Unavailable, "MalformedRefs", "Native reference output has an invalid record shape.")
		}
		name := string(fields[0])
		locator, le := refLocator(repo.id, name, remoteBindings)
		if le != nil {
			result.diagnostics = append(result.diagnostics, safeError(le))
			continue
		}
		d := api.RefFactData{Locator: locator, Observation: result.observation}
		object, oe := domain.NewOID(string(fields[1]))
		if oe != nil || object.Format() != repo.format {
			return result, diagnostic(api.Unavailable, "MalformedRefObject", "A native reference object has an invalid full identity.")
		}
		typeName := string(fields[2])
		commitObject := object
		if typeName == "tag" {
			d.TagObject = api.Some(object)
			typeName = string(fields[4])
			if typeName == "commit" {
				commitObject, oe = domain.NewOID(string(fields[3]))
				if oe != nil {
					return result, oe
				}
			}
			if typeName == "tag" {
				peel := s.command(repo.common.path, "--git-dir="+repo.common.path, "rev-parse", "--verify", "--end-of-options", name+"^{commit}")
				if peel.err == nil {
					commitObject, oe = domain.NewOID(line(peel.stdout))
					if oe != nil {
						return result, oe
					}
					typeName = "commit"
				} else {
					result.diagnostics = append(result.diagnostics, diagnostic(api.Unavailable, "TagCommitUnavailable", "An annotated tag has no established peeled commit."))
				}
			}
		}
		if typeName == "commit" {
			revision, re := domain.NewRevision(repo.id, commitObject)
			if re != nil {
				return result, re
			}
			d.Revision = api.Some(revision)
		}
		if len(fields[5]) > 0 {
			target, te := refLocator(repo.id, string(fields[5]), remoteBindings)
			if te != nil {
				result.diagnostics = append(result.diagnostics, safeError(te))
			} else {
				d.SymbolicTarget = api.Some(target)
			}
		}
		if cached, ok := locator.(api.CachedRemoteRef); ok {
			freshness, fe := api.NewFetchFreshness(api.FetchFreshnessData{Kind: api.Cached, Binding: cached.Data().Binding, RefScope: []api.GitRefLocator{locator}, Observation: result.observation})
			if fe != nil {
				return result, fe
			}
			d.Freshness = api.Some(freshness)
		}
		fact, fe := api.NewRefFact(d)
		if fe != nil {
			return result, fe
		}
		result.facts = append(result.facts, fact)
		result.byName[name] = fact
	}
	if len(result.diagnostics) > 0 {
		d := result.observation.Data()
		d.Completeness = api.Partial
		result.observation, err = api.NewGitObservation(d)
	}
	return result, err
}

func refLocator(repo domain.RepositoryID, name string, bindings []api.RemoteBinding) (api.GitRefLocator, error) {
	switch {
	case strings.HasPrefix(name, "refs/heads/"):
		branch, err := domain.NewBranchID(repo, domain.Local, strings.TrimPrefix(name, "refs/heads/"))
		if err != nil {
			return nil, err
		}
		return api.NewLocalBranchRef(api.LocalBranchRefData{Branch: branch})
	case strings.HasPrefix(name, "refs/tags/"):
		return api.NewLocalTagRef(api.LocalTagRefData{Repository: repo, Ref: name})
	case strings.HasPrefix(name, "refs/remotes/"):
		var matches []api.RemoteBinding
		for _, binding := range bindings {
			for _, mapping := range binding.Data().Refspecs {
				if _, ok := mapRef(mapping.Data().Destination, mapping.Data().Source, name); ok {
					matches = append(matches, binding)
					break
				}
			}
		}
		if len(matches) == 1 {
			return api.NewCachedRemoteRef(api.CachedRemoteRefData{Binding: matches[0], Ref: name})
		}
		return nil, diagnostic(api.Unavailable, "CachedRefBindingAmbiguous", "A cached ref has no unique observed remote mapping.")
	default:
		return nil, diagnostic(api.Invalid, "InvalidRefNamespace", "The locator is outside the supported native reference namespaces.")
	}
}

func mapRef(pattern, replacement, ref string) (string, bool) {
	if strings.Count(pattern, "*") == 0 {
		return replacement, pattern == ref
	}
	if strings.Count(pattern, "*") != 1 || strings.Count(replacement, "*") != 1 {
		return "", false
	}
	pre, post, _ := strings.Cut(pattern, "*")
	if !strings.HasPrefix(ref, pre) || !strings.HasSuffix(ref, post) || len(ref) < len(pre)+len(post) {
		return "", false
	}
	return strings.Replace(replacement, "*", ref[len(pre):len(ref)-len(post)], 1), true
}

func nativeRef(locator api.GitRefLocator) string {
	switch v := locator.(type) {
	case api.LocalBranchRef:
		return "refs/heads/" + v.Data().Branch.Name()
	case api.LocalTagRef:
		return v.Data().Ref
	case api.CachedRemoteRef:
		return v.Data().Ref
	case api.RemoteRef:
		return v.Data().Ref
	}
	return ""
}

func (a *Adapter) ListRefs(ctx context.Context, request api.ListRefsRequest) (api.ListRefsResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.ListRefsResult{}, diagnostic(api.Invalid, "InvalidRequest", "The reference request is invalid.")
	}
	rd := request.Data()
	d := api.ListRefsResultData{}
	repo, err := a.registered(s.ctx, rd.Repository)
	var snapshot refsSnapshot
	if err == nil {
		snapshot, err = s.refs(repo)
	}
	binding := queryBinding("refs", rd.Repository.Token(), snapshot.fingerprint, fmt.Sprint(rd.Kinds))
	version := sourceVersion("refs", rd.Repository.Token(), a.lifetime, []byte(snapshot.fingerprint))
	var offset uint64
	var more bool
	complete := api.Complete
	if err == nil {
		if _, p := rd.Page.Data().Continuation.(api.OffsetPage); p {
			err = diagnostic(api.Invalid, "SourceBoundCursorRequired", "Mutable reference pages require the returned source-bound cursor.")
		} else {
			offset, err = a.pageOffset(rd.Page, binding)
		}
	}
	if err == nil {
		wanted := map[api.RefKind]bool{}
		for _, kind := range rd.Kinds {
			wanted[kind] = true
		}
		var selected []api.RefFact
		for _, fact := range snapshot.facts {
			kind := api.LocalBranchKind
			switch fact.Data().Locator.(type) {
			case api.LocalTagRef:
				kind = api.LocalTagKind
			case api.CachedRemoteRef:
				kind = api.CachedRemoteKind
			}
			if len(wanted) == 0 || wanted[kind] {
				selected = append(selected, fact)
			}
		}
		if offset < uint64(len(selected)) {
			end := offset + uint64(rd.Page.Data().Limit)
			if end < uint64(len(selected)) {
				more = true
			} else {
				end = uint64(len(selected))
			}
			d.Refs = append(d.Refs, selected[int(offset):int(end)]...)
		}
		if more {
			complete = api.More
		}
		if len(snapshot.diagnostics) > 0 {
			complete = api.Partial
			more = false
		}
		d.Diagnostics = append(d.Diagnostics, snapshot.diagnostics...)
	}
	if err != nil {
		complete = api.Partial
		more = false
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	observation, oe := s.observation(rd.Repository, api.None[domain.WorktreeID](), version, complete)
	if oe != nil {
		return api.ListRefsResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Page, oe = a.pageInfo(version, binding, offset, len(d.Refs), more, complete)
	if oe != nil {
		return api.ListRefsResult{}, oe
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewListRefsResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func (a *Adapter) ResolveExact(ctx context.Context, request api.ResolveExactRequest) (api.ResolveExactResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	d := api.ResolveExactResultData{}
	var err error
	if !request.Valid() {
		err = diagnostic(api.Invalid, "InvalidRequest", "The exact resolution request is invalid.")
	} else {
		rd := request.Data()
		repo, re := a.registered(s.ctx, rd.Repository)
		err = re
		var snapshot refsSnapshot
		if err == nil {
			snapshot, err = s.refs(repo)
		}
		if err == nil {
			d.Observation = api.Some(snapshot.observation)
			d.Diagnostics = append(d.Diagnostics, snapshot.diagnostics...)
			if source, p := rd.ExpectedSource.Value(); p && source != snapshot.observation.Data().Version {
				err = diagnostic(api.StaleObservation, "ExactSourceChanged", "The selected native reference/configuration source changed.")
			}
		}
		if err == nil {
			expected := rd.Target.ExpectedRevision()
			local, ve := s.verifyCommit(repo, expected.OID().String())
			err = ve
			if err == nil {
				resolution := api.ExactLocalResolutionData{Requested: rd.Target, Local: local, Observation: snapshot.observation, Locator: rd.Locator}
				locator, p := rd.Locator.Value()
				if !p {
					if branch, b := rd.Target.Branch(); b && branch.Kind() == domain.Local {
						locator, err = api.NewLocalBranchRef(api.LocalBranchRefData{Branch: branch})
						resolution.Locator = api.Some(locator)
						p = true
					}
				}
				if err == nil && p {
					name := nativeRef(locator)
					if branch, b := rd.Target.Branch(); b && branch.Kind() == domain.Local && name != "refs/heads/"+branch.Name() {
						err = diagnostic(api.Invalid, "TargetLocatorMismatch", "The locator does not name the exact selected branch.")
					}
					var supplied api.Optional[api.RemoteBinding]
					switch v := locator.(type) {
					case api.CachedRemoteRef:
						supplied = api.Some(v.Data().Binding)
					case api.RemoteRef:
						supplied = api.Some(v.Data().Binding)
					}
					if binding, has := supplied.Value(); has && err == nil {
						valid := false
						for _, actual := range snapshot.remotes {
							bd, ad := binding.Data(), actual.Data()
							if bd.LocalRepository == ad.LocalRepository && bd.RemoteRepository == ad.RemoteRepository && bd.RemoteName == ad.RemoteName && bd.Configuration == ad.Configuration {
								valid = true
							}
						}
						if !valid {
							err = diagnostic(api.StaleObservation, "RemoteBindingChanged", "The selected remote binding no longer matches current configuration.")
						}
						if branch, has := rd.Target.Branch(); has && branch.Kind() == domain.RemoteHead && err == nil {
							source := name
							if _, cached := locator.(api.CachedRemoteRef); cached {
								var sources []string
								for _, mapping := range binding.Data().Refspecs {
									if source, ok := mapRef(mapping.Data().Destination, mapping.Data().Source, name); ok {
										sources = append(sources, source)
									}
								}
								if len(sources) != 1 {
									err = diagnostic(api.Invalid, "RemoteMappingAmbiguous", "The selected cached ref has no unique remote source.")
								} else {
									source = sources[0]
								}
							}
							if err == nil && source != "refs/heads/"+branch.Name() {
								err = diagnostic(api.Invalid, "TargetLocatorMismatch", "The locator does not map the exact selected remote branch.")
							}
						}
					}
					if remote, ok := locator.(api.RemoteRef); ok {
						var mapped []string
						for _, binding := range snapshot.remotes {
							if binding.Data().RemoteName == remote.Data().Binding.Data().RemoteName && binding.Data().Configuration == remote.Data().Binding.Data().Configuration && binding.Data().RemoteRepository == expected.Repository() {
								for _, mapping := range binding.Data().Refspecs {
									if destination, ok := mapRef(mapping.Data().Source, mapping.Data().Destination, name); ok {
										mapped = append(mapped, destination)
									}
								}
							}
						}
						if len(mapped) != 1 {
							err = diagnostic(api.NotFound, "RemoteRootNotAcquired", "No uniquely mapped acquired local root exists for this exact remote ref.")
						} else {
							name = mapped[0]
						}
					}
					if err == nil {
						observed, found := snapshot.byName[name]
						revision, has := observed.Data().Revision.Value()
						if !found || !has {
							err = diagnostic(api.NotFound, "ExactLocatorMissing", "The selected locator has no established local commit.")
						} else if revision != local {
							err = diagnostic(api.StaleObservation, "ExpectedRevisionMismatch", "The locator differs from the selected exact commit.")
						} else if expected.Repository().Scope() == domain.Remote {
							cached, ok := observed.Data().Locator.(api.CachedRemoteRef)
							if !ok || cached.Data().Binding.Data().RemoteRepository != expected.Repository() {
								err = diagnostic(api.Invalid, "RemoteAssociationMissing", "Equal object bytes do not establish the requested remote association.")
							} else {
								resolution.Binding = api.Some(cached.Data().Binding)
								resolution.ObservedRemote = api.Some(expected)
							}
						}
					}
				}
				if err == nil && expected.Repository().Scope() == domain.Remote && !resolution.Binding.Present() {
					err = diagnostic(api.Invalid, "RemoteAssociationMissing", "An explicit acquired remote locator and binding are required.")
				}
				if err == nil {
					value, ce := api.NewExactLocalResolution(resolution)
					err = ce
					if ce == nil {
						d.Resolution = api.Some(value)
					}
				}
			}
		}
	}
	if err != nil {
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewResolveExactResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}
