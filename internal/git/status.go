package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type statusSnapshot struct {
	facts       api.StatusFacts
	index       []byte
	entries     map[api.GitPath][]api.IndexEntryFact
	inventory   api.SourceVersion
	admin       string
	diagnostics []api.Diagnostic
}

func indexBytes(admin string) ([]byte, error) {
	raw, err := readSmallFile(filepath.Join(admin, "index"), 16<<20)
	if os.IsNotExist(err) {
		return []byte{0}, nil
	}
	if err != nil {
		return nil, err
	}
	return append([]byte{1}, raw...), nil
}

func parseIndexFacts(raw []byte, format domain.ObjectFormat) (map[api.GitPath][]api.IndexEntryFact, error) {
	entries := make(map[api.GitPath][]api.IndexEntryFact)
	for offset := 0; offset < len(raw); {
		nul := bytes.IndexByte(raw[offset:], 0)
		if nul < 0 {
			return nil, diagnostic(api.Unavailable, "MalformedIndexFacts", "Native index output lacks a path terminator.")
		}
		header := raw[offset : offset+nul]
		offset += nul + 1
		end := offset
		// Git's supported --debug profile emits five fixed metadata lines after
		// the NUL-terminated path. Path bytes never enter this line parser.
		for n := 0; n < 5; n++ {
			newline := bytes.IndexByte(raw[end:], '\n')
			if newline < 0 {
				return nil, diagnostic(api.Unavailable, "MalformedIndexFlags", "Native index debug metadata is incomplete.")
			}
			end += newline + 1
		}
		metadata := string(raw[offset:end])
		offset = end
		position := strings.LastIndex(metadata, "flags: ")
		if position < 0 {
			return nil, diagnostic(api.Unavailable, "MalformedIndexFlags", "Native index flags are unavailable.")
		}
		flags, err := strconv.ParseUint(strings.TrimSpace(metadata[position+len("flags: "):]), 16, 32)
		if err != nil {
			return nil, diagnostic(api.Unavailable, "MalformedIndexFlags", "Native index flags are invalid.")
		}
		prefix, pathBytes, ok := bytes.Cut(header, []byte{'\t'})
		if !ok {
			return nil, diagnostic(api.Unavailable, "MalformedIndexFacts", "Native index entry lacks a literal path.")
		}
		fields := strings.Fields(string(prefix))
		if len(fields) != 3 {
			return nil, diagnostic(api.Unavailable, "MalformedIndexFacts", "Native index entry fields are invalid.")
		}
		path, err := api.NewGitPath(string(pathBytes))
		if err != nil {
			return nil, err
		}
		mode, err := strconv.ParseUint(fields[0], 8, 32)
		if err != nil {
			return nil, err
		}
		oid, err := domain.NewOID(fields[1])
		if err != nil || oid.Format() != format {
			return nil, diagnostic(api.Unavailable, "MalformedIndexObject", "Native index object identity is invalid.")
		}
		stage, err := strconv.ParseUint(fields[2], 10, 8)
		if err != nil || stage > 3 {
			return nil, diagnostic(api.Unavailable, "MalformedIndexStage", "Native index stage is invalid.")
		}
		var semantic []api.IndexFlag
		if flags&0x8000 != 0 {
			semantic = append(semantic, api.AssumeUnchanged)
		}
		if flags&0x40000000 != 0 {
			semantic = append(semantic, api.SkipWorktree)
		}
		if flags&0x20000000 != 0 {
			semantic = append(semantic, api.IntentToAdd)
		}
		if flags & ^uint64(0x6000f000|0x40000) != 0 {
			semantic = append(semantic, api.ExtendedIndexFlag)
		}
		entry, err := api.NewIndexEntryFact(api.IndexEntryFactData{Path: path, Stage: uint8(stage), Object: oid, Mode: uint32(mode), SemanticFlags: semantic})
		if err != nil {
			return nil, err
		}
		entries[path] = append(entries[path], entry)
	}
	return entries, nil
}

func (a *Adapter) ObserveStatus(ctx context.Context, request api.ObserveStatusRequest) (api.ObserveStatusResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	d := api.ObserveStatusResultData{}
	var err error
	if !request.Valid() {
		err = diagnostic(api.Invalid, "InvalidRequest", "The status request is invalid.")
	} else {
		repo, re := a.registered(s.ctx, request.Data().Worktree.Repository())
		err = re
		if err == nil {
			snapshot, se := s.status(repo, request.Data().Worktree)
			err = se
			d.Diagnostics = snapshot.diagnostics
			if snapshot.facts.Valid() {
				d.Status = api.Some(snapshot.facts)
				d.Observation = api.Some(snapshot.facts.Data().Observation)
			}
		}
	}
	if err != nil {
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewObserveStatusResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func (s *readSession) status(repo repository, id domain.WorktreeID) (statusSnapshot, error) {
	var result statusSnapshot
	inventory, err := s.inventory(repo)
	result.diagnostics = append(result.diagnostics, inventory.diagnostics...)
	if err != nil {
		return result, err
	}
	result.inventory = inventory.observation.Data().Version
	var worktree api.WorktreeFacts
	for _, w := range inventory.facts {
		if w.Data().ID == id {
			worktree = w
			break
		}
	}
	if !worktree.Valid() {
		return result, diagnostic(api.NotFound, "WorktreeNotRegistered", "The requested administrative worktree identity is not in the native inventory.")
	}
	scope, sp := worktree.Data().Scope.Value()
	head, hp := worktree.Data().Head.Value()
	if !sp || !hp {
		return result, diagnostic(api.Unavailable, "WorktreeUnavailable", "The selected worktree has no established root and Head observation.")
	}
	root := scope.Data().RootLocator
	admin := repo.common.path
	if !worktree.Data().Primary {
		admin = filepath.Join(repo.common.path, "worktrees", strings.TrimPrefix(id.AdministrativeKey(), "linked:"))
	}
	result.admin = admin
	repo.contextAdmin = admin
	repo.contextRoot = root
	result.index, err = indexBytes(admin)
	if err != nil {
		return result, err
	}
	indexVersion := sourceVersion("index", queryBinding(repo.id.Token(), id.AdministrativeKey()), s.a.lifetime, result.index)
	prefix := []string{"--git-dir=" + admin, "--work-tree=" + root}
	run := func(args ...string) commandResult {
		return s.command(root, append(append([]string(nil), prefix...), args...)...)
	}
	index := run("ls-files", "--stage", "--debug", "-z")
	if index.err != nil {
		return result, index.err
	}
	result.entries, err = parseIndexFacts(index.stdout, repo.format)
	if err != nil {
		return result, err
	}
	config := run("config", "--null", "--list", "--show-origin", "--show-scope")
	if config.err != nil {
		return result, config.err
	}
	configuration := sourceVersion("configuration", repo.id.Token(), s.a.lifetime, config.stdout)
	var base domain.OID
	if revision, p := head.Revision(); p {
		base = revision.OID()
	} else {
		base, err = s.emptyTree(repo)
		if err != nil {
			return result, err
		}
	}
	staged := run("diff", "--cached", "--no-ext-diff", "--no-textconv", "--find-renames", "--name-status", "-z", base.String(), "--")
	if staged.err != nil {
		return result, staged.err
	}
	working := run("diff", "--no-ext-diff", "--no-textconv", "--find-renames", "--name-status", "-z", "--")
	if working.err != nil {
		return result, working.err
	}
	untracked := run("ls-files", "--others", "--exclude-standard", "-z")
	if untracked.err != nil {
		return result, untracked.err
	}
	var candidates []api.ChangeFactData
	conflicts := make(map[api.GitPath]bool)
	for path, entries := range result.entries {
		for _, entry := range entries {
			if entry.Data().Stage != 0 {
				conflicts[path] = true
			}
		}
	}
	for _, source := range []struct {
		bytes []byte
		cause api.ChangeCause
	}{{staged.stdout, api.IndexChangeCause}, {working.stdout, api.WorktreeChangeCause}} {
		files, pe := parseNameStatus(source.bytes)
		if pe != nil {
			return result, pe
		}
		for _, file := range files {
			if conflicts[file.Path] || file.Kind == api.Unmerged {
				continue
			}
			candidates = append(candidates, api.ChangeFactData{Path: file.Path, OldPath: file.OldPath, Cause: source.cause, Kind: file.Kind})
		}
	}
	for _, name := range bytes.Split(untracked.stdout, []byte{0}) {
		if len(name) == 0 {
			continue
		}
		path, pe := api.NewGitPath(string(name))
		if pe != nil {
			return result, pe
		}
		if !conflicts[path] {
			candidates = append(candidates, api.ChangeFactData{Path: path, Cause: api.UntrackedChangeCause, Kind: api.Untracked})
		}
	}
	for path := range conflicts {
		candidates = append(candidates, api.ChangeFactData{Path: path, Cause: api.ConflictChangeCause, Kind: api.Unmerged})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Path.String() != candidates[j].Path.String() {
			return candidates[i].Path.String() < candidates[j].Path.String()
		}
		return candidates[i].Cause < candidates[j].Cause
	})
	states := make(map[api.GitPath]api.FileState)
	failed := make(map[api.GitPath]bool)
	var changes []api.ChangeFact
	var sourceParts []string
	for _, candidate := range candidates {
		if failed[candidate.Path] {
			continue
		}
		state, p := states[candidate.Path]
		if !p {
			state, err = s.fileState(scope, candidate.Path)
			if err != nil {
				failed[candidate.Path] = true
				result.diagnostics = append(result.diagnostics, safeError(err))
				continue
			}
			states[candidate.Path] = state
		}
		candidate.WorktreeState = state
		candidate.IndexEntries = result.entries[candidate.Path]
		if candidate.Cause == api.UntrackedChangeCause {
			candidate.IndexEntries = nil
		}
		value, ce := api.NewChangeFact(candidate)
		if ce != nil {
			return result, ce
		}
		changes = append(changes, value)
		sourceParts = append(sourceParts, fmt.Sprint(value.Data()))
	}
	var drift error
	endIndex, ie := indexBytes(admin)
	if ie != nil {
		drift = ie
	} else if !bytes.Equal(result.index, endIndex) {
		drift = diagnostic(api.StaleObservation, "IndexChangedDuringStatus", "The index changed while status facts were acquired.")
	}
	endHead, he := s.readHead(repo, admin)
	if he != nil {
		drift = he
	} else if endHead != head {
		drift = diagnostic(api.StaleObservation, "HeadChangedDuringStatus", "The exact Head changed while status facts were acquired.")
	}
	endRoot, re := observeDirectory(root)
	if re != nil || endRoot.identity != scope.Data().RootIdentity {
		drift = diagnostic(api.StaleObservation, "RootChangedDuringStatus", "The physical worktree root changed during status observation.")
	}
	if drift != nil {
		result.diagnostics = append(result.diagnostics, safeError(drift))
	}
	worktreeVersion := sourceVersion("worktree", queryBinding(repo.id.Token(), id.AdministrativeKey()), s.a.lifetime, []byte(queryBinding(string(working.stdout), string(untracked.stdout), strings.Join(sourceParts, "\x00"))))
	version := sourceVersion("status", queryBinding(repo.id.Token(), id.AdministrativeKey()), s.a.lifetime, []byte(queryBinding(fmt.Sprint(head), fmt.Sprint(indexVersion), fmt.Sprint(worktreeVersion), fmt.Sprint(configuration), fmt.Sprint(result.inventory))))
	complete := api.Complete
	if len(result.diagnostics) > 0 || inventory.observation.Data().Completeness != api.Complete {
		complete = api.Partial
	}
	observation, oe := s.observation(repo.id, api.Some(id), version, complete)
	if oe != nil {
		return result, oe
	}
	upstream, diags := s.upstream(repo, head, observation)
	result.diagnostics = append(result.diagnostics, diags...)
	if len(diags) > 0 && complete == api.Complete {
		data := observation.Data()
		data.Completeness = api.Partial
		observation, oe = api.NewGitObservation(data)
		if oe != nil {
			return result, oe
		}
	}
	result.facts, err = api.NewStatusFacts(api.StatusFactsData{Worktree: worktree, Changes: changes, IndexVersion: indexVersion, WorktreeVersion: worktreeVersion, ConfigurationVersion: configuration, Upstream: upstream, Observation: observation})
	if err != nil {
		return result, err
	}
	if drift != nil {
		return result, drift
	}
	if len(failed) > 0 {
		return result, diagnostic(api.Unavailable, "PartialStatusFiles", "Some changed paths could not supply complete current file facts.")
	}
	return result, nil
}

func (s *readSession) upstream(repo repository, head domain.Head, observation api.GitObservation) (api.UpstreamFact, []api.Diagnostic) {
	unresolved := func(err error, binding api.Optional[api.RemoteBinding], ref api.Optional[api.GitRefLocator]) (api.UpstreamFact, []api.Diagnostic) {
		diag := safeError(err)
		value, _ := api.NewUnresolvedUpstream(api.UnresolvedUpstreamData{Binding: binding, Ref: ref, Diagnostic: diag})
		return value, []api.Diagnostic{diag}
	}
	branch, bp := head.Branch()
	local, lp := head.Revision()
	if !bp || !lp {
		value, _ := api.NewUpstreamNotApplicable(api.UpstreamNotApplicableData{})
		return value, nil
	}
	q := s.command(repo.cwd(), "--git-dir="+repo.gitDir(), "for-each-ref", "--format=%(refname)%00%(upstream)%00%(upstream:remotename)%00%(upstream:remoteref)%00", "--", "refs/heads/"+branch.Name())
	if q.err != nil {
		return unresolved(q.err, api.None[api.RemoteBinding](), api.None[api.GitRefLocator]())
	}
	var fields [][]byte
	for _, row := range bytes.Split(q.stdout, []byte{'\n'}) {
		parts := bytes.Split(row, []byte{0})
		if len(parts) == 5 && string(parts[0]) == "refs/heads/"+branch.Name() {
			fields = parts
			break
		}
	}
	if fields == nil {
		return unresolved(diagnostic(api.StaleObservation, "UpstreamHeadMissing", "The observed branch disappeared before upstream inspection."), api.None[api.RemoteBinding](), api.None[api.GitRefLocator]())
	}
	if len(fields[1]) == 0 && len(fields[2]) == 0 && len(fields[3]) == 0 {
		value, _ := api.NewNoUpstream(api.NoUpstreamData{})
		return value, nil
	}
	refs, err := s.refs(repo)
	if err != nil {
		return unresolved(err, api.None[api.RemoteBinding](), api.None[api.GitRefLocator]())
	}
	var binding api.RemoteBinding
	for _, candidate := range refs.remotes {
		if candidate.Data().RemoteName == string(fields[2]) {
			binding = candidate
			break
		}
	}
	if !binding.Valid() {
		return unresolved(diagnostic(api.Unavailable, "UpstreamBindingUnavailable", "The configured upstream has no established remote binding."), api.None[api.RemoteBinding](), api.None[api.GitRefLocator]())
	}
	locator, err := api.NewCachedRemoteRef(api.CachedRemoteRefData{Binding: binding, Ref: string(fields[1])})
	if err != nil {
		return unresolved(err, api.Some(binding), api.None[api.GitRefLocator]())
	}
	fact, found := refs.byName[string(fields[1])]
	if !found {
		if refs.observation.Data().Completeness != api.Complete {
			return unresolved(diagnostic(api.Unavailable, "UpstreamUnresolved", "Incomplete reference observation cannot prove upstream absence."), api.Some(binding), api.Some[api.GitRefLocator](locator))
		}
		value, _ := api.NewGoneUpstream(api.GoneUpstreamData{Binding: binding, Ref: locator, Evidence: refs.observation})
		return value, nil
	}
	remote, known := fact.Data().Revision.Value()
	if !known {
		return unresolved(diagnostic(api.Unavailable, "UpstreamUnresolved", "The upstream ref has no established commit."), api.Some(binding), api.Some[api.GitRefLocator](locator))
	}
	if !strings.HasPrefix(string(fields[3]), "refs/heads/") {
		return unresolved(diagnostic(api.Unsupported, "UpstreamSourceNamespace", "The configured upstream source is not a remote branch."), api.Some(binding), api.Some[api.GitRefLocator](locator))
	}
	remoteBranch, err := domain.NewBranchID(binding.Data().RemoteRepository, domain.RemoteHead, strings.TrimPrefix(string(fields[3]), "refs/heads/"))
	if err != nil {
		return unresolved(err, api.Some(binding), api.Some[api.GitRefLocator](locator))
	}
	comparison := api.RevisionComparisonData{Left: local, Right: remote, Observation: observation}
	var diagnostics []api.Diagnostic
	counts := s.command(repo.common.path, "--git-dir="+repo.common.path, "rev-list", "--left-right", "--count", local.OID().String()+"..."+remote.OID().String())
	if counts.err == nil {
		numbers := strings.Fields(string(counts.stdout))
		if len(numbers) == 2 {
			ahead, ae := strconv.ParseUint(numbers[0], 10, 64)
			behind, be := strconv.ParseUint(numbers[1], 10, 64)
			if ae == nil && be == nil {
				comparison.Ahead = api.Some(ahead)
				comparison.Behind = api.Some(behind)
			}
		}
	}
	if !comparison.Ahead.Present() {
		diagnostics = append(diagnostics, diagnostic(api.Unavailable, "UpstreamCountsUnavailable", "The exact upstream endpoints are known but ahead/behind counts were not established."))
	}
	comp, err := api.NewRevisionComparison(comparison)
	if err != nil {
		return unresolved(err, api.Some(binding), api.Some[api.GitRefLocator](locator))
	}
	freshness, err := api.NewFetchFreshness(api.FetchFreshnessData{Kind: api.Cached, Binding: binding, RefScope: []api.GitRefLocator{locator}, Observation: refs.observation})
	if err != nil {
		return unresolved(err, api.Some(binding), api.Some[api.GitRefLocator](locator))
	}
	value, err := api.NewResolvedUpstream(api.ResolvedUpstreamData{Binding: binding, RemoteBranch: remoteBranch, CachedLocalRef: locator, Local: local, Upstream: remote, Comparison: comp, Freshness: freshness})
	if err != nil {
		return unresolved(err, api.Some(binding), api.Some[api.GitRefLocator](locator))
	}
	return value, diagnostics
}
