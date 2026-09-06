package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func refScopes(v GitRefLocator) (local, object domain.RepositoryID) {
	switch x := v.(type) {
	case LocalBranchRef:
		return x.data.Branch.Repository(), x.data.Branch.Repository()
	case LocalTagRef:
		return x.data.Repository, x.data.Repository
	case CachedRemoteRef:
		return x.data.Binding.data.LocalRepository, x.data.Binding.data.LocalRepository
	case RemoteRef:
		return x.data.Binding.data.LocalRepository, x.data.Binding.data.RemoteRepository
	}
	return
}
func consistentRefFact(d RefFactData) error {
	local, object := refScopes(d.Locator)
	if local != d.Observation.data.Repository {
		return invalid("ref observation repository")
	}
	if r, p := d.Revision.Value(); p && r.Repository() != object {
		return invalid("ref revision repository")
	}
	if target, p := d.SymbolicTarget.Value(); p {
		l, o := refScopes(target)
		if l != local || o != object {
			return invalid("symbolic ref scope")
		}
	}
	if f, p := d.Freshness.Value(); p && f.data.Binding.data.LocalRepository != local {
		return invalid("ref freshness scope")
	}
	return nil
}
func consistentStatusFacts(d StatusFactsData) error {
	w := d.Worktree.data.ID
	if o, p := d.Observation.data.Worktree.Value(); p && o != w {
		return invalid("status observation worktree")
	}
	if err := upstreamScope(d.Upstream, w.Repository()); err != nil {
		return err
	}
	return consistentStatusChanges(d.Changes)
}
func upstreamScope(v UpstreamFact, r domain.RepositoryID) error {
	switch x := v.(type) {
	case ResolvedUpstream:
		if x.data.Binding.data.LocalRepository != r {
			return invalid("upstream binding repository")
		}
	case GoneUpstream:
		if x.data.Binding.data.LocalRepository != r || x.data.Evidence.data.Repository != r {
			return invalid("gone upstream repository")
		}
	case UnresolvedUpstream:
		if b, p := x.data.Binding.Value(); p && b.data.LocalRepository != r {
			return invalid("unresolved upstream repository")
		}
	}
	return nil
}
func successfulWorktree(w WorktreeFacts, revision Optional[domain.Revision]) bool {
	if _, p := w.data.Availability.(AvailableWorktree); !p {
		return false
	}
	s, sp := w.data.Scope.Value()
	h, hp := w.data.Head.Value()
	if !sp || !hp || s.data.ID != w.data.ID || !h.MatchesWorktree(w.data.ID) {
		return false
	}
	if r, p := revision.Value(); p {
		actual, p := h.Revision()
		return p && actual == r
	}
	return true
}
func successfulStatus(s StatusFacts) bool {
	return successfulWorktree(s.data.Worktree, None[domain.Revision]())
}
func consistentWorktreeCreated(d WorktreeCreatedData) error {
	if !successfulWorktree(d.Worktree, Some(d.Target.data.Local)) {
		return invalid("created worktree postconditions")
	}
	return nil
}
func consistentWorktreeRetargeted(d WorktreeRetargetedData) error {
	if !successfulWorktree(d.Worktree, Some(d.Target.data.Local)) || !d.PriorHead.MatchesWorktree(d.Worktree.data.ID) {
		return invalid("retarget postconditions")
	}
	return nil
}
func consistentIndexChanged(d IndexChangedData) error {
	if !successfulStatus(d.Status) {
		return invalid("index changed status")
	}
	return nil
}
func consistentCommitCreated(d CommitCreatedData) error {
	if d.Revision.Repository().Scope() != domain.LocalCommon {
		return invalid("created commit scope")
	}
	if r, p := d.Candidate.data.Candidate.Value(); p && r != d.Revision {
		return invalid("created candidate revision")
	}
	if r, p := d.Candidate.data.Parent.Value(); p && !sameLocal(r, d.Revision) {
		return invalid("candidate parent scope")
	}
	return nil
}
func consistentTrackedRestored(d TrackedRestoredData) error {
	if len(d.Paths) == 0 || duplicatePaths(d.Paths) || !successfulStatus(d.Status) {
		return invalid("restored postconditions")
	}
	return gitRecoverySubjects(d.Recovery, d.Status.data.Worktree.data.ID.Repository(), Some(d.Status.data.Worktree.data.ID), nil, false)
}
func recoveryWorktree(records []RecoveryRecord, w domain.WorktreeID) error {
	for _, r := range records {
		if x, p := r.data.Subject.data.Worktree.Value(); p && x != w {
			return invalid("recovery worktree")
		}
		if x, p := r.data.Subject.data.Repository.Value(); p && x != w.Repository() {
			return invalid("recovery repository")
		}
	}
	return nil
}
func stashStatus(stash domain.StashID, status StatusFacts, success bool) error {
	if stash.Repository() != status.data.Worktree.data.ID.Repository() {
		return invalid("stash status repository")
	}
	if success && !successfulStatus(status) {
		return invalid("stash success status")
	}
	return nil
}
func consistentStashCreated(d StashCreatedData) error { return stashStatus(d.Stash, d.Status, true) }
func consistentStashCreatedCleanupRefused(d StashCreatedCleanupRefusedData) error {
	if err := gitRecoverySubjects(d.Recovery, d.Stash.Repository(), None[domain.WorktreeID](), nil, false); err != nil {
		return err
	}
	if s, p := d.Status.Value(); p {
		if err := stashStatus(d.Stash, s, false); err != nil {
			return err
		}
		return recoveryWorktree(d.Recovery, s.data.Worktree.data.ID)
	}
	return nil
}
func consistentStashApplied(d StashAppliedData) error {
	if !d.IndexRestored || !d.Retained {
		return invalid("stash apply index/retention postcondition")
	}
	return stashStatus(d.Stash, d.Status, true)
}
func consistentAppliedWithConflicts(d AppliedWithConflictsData) error {
	if err := gitRecoverySubjects(d.Recovery, d.Stash.Repository(), None[domain.WorktreeID](), nil, false); err != nil {
		return err
	}
	if len(d.ConflictPaths) == 0 || len(d.IndexEntries) == 0 {
		return invalid("conflict evidence")
	}
	for _, entry := range d.IndexEntries {
		found := false
		for _, p := range d.ConflictPaths {
			found = found || entry.data.Path == p
		}
		if !found || entry.data.Stage == 0 {
			return invalid("conflict index stage")
		}
	}
	if s, p := d.Status.Value(); p {
		return stashStatus(d.Stash, s, false)
	}
	return nil
}
func consistentStashDropped(d StashDroppedData) error {
	if d.Observation.data.Repository != d.Stash.Repository() {
		return invalid("stash drop observation")
	}
	for _, s := range d.Survivors {
		if s.data.ID.Repository() != d.Stash.Repository() {
			return invalid("stash survivor repository")
		}
		if s.data.ID == d.Stash && s.data.Occurrence == d.Occurrence {
			return invalid("dropped occurrence survives")
		}
	}
	return nil
}
func consistentBranchCreated(d BranchCreatedData) error {
	if w, p := d.Worktree.Value(); p {
		if !successfulWorktree(w, Some(d.Revision)) {
			return invalid("branch checkout postconditions")
		}
		h, _ := w.data.Head.Value()
		b, p := h.Branch()
		if !p || b != d.Branch {
			return invalid("branch checkout identity")
		}
	}
	return nil
}
func consistentPushed(d PushedData) error {
	if d.Source.Repository() != d.Binding.data.LocalRepository || d.Destination.Repository() != d.Binding.data.RemoteRepository || d.Destination.Kind() != domain.RemoteHead {
		return invalid("push result scope")
	}
	if r, p := d.ObservedRemote.Value(); p && r.Repository() != d.Destination.Repository() {
		return invalid("observed push endpoint scope")
	}
	if c, p := d.Configuration.Value(); p && c.data.Branch.Repository() != d.Source.Repository() {
		return invalid("push configuration scope")
	}
	return nil
}
func postRepository(d GitPostFactsData) (domain.RepositoryID, error) {
	var repo domain.RepositoryID
	check := func(r domain.RepositoryID) bool {
		if !repo.Valid() {
			repo = r
		}
		return repo == r
	}
	for _, w := range d.Worktrees {
		if !check(w.data.ID.Repository()) {
			return repo, invalid("post worktree scope")
		}
	}
	for _, s := range d.Status {
		if !check(s.data.Worktree.data.ID.Repository()) {
			return repo, invalid("post status scope")
		}
	}
	for _, r := range d.Refs {
		if !check(r.data.Observation.data.Repository) {
			return repo, invalid("post ref scope")
		}
	}
	for _, s := range d.Stashes {
		if !check(s.data.ID.Repository()) {
			return repo, invalid("post stash scope")
		}
	}
	for _, c := range d.Configuration {
		if !check(c.data.Branch.Repository()) {
			return repo, invalid("post configuration scope")
		}
	}
	if c, p := d.Commit.Value(); p {
		for _, v := range []Optional[domain.Revision]{c.data.Candidate, c.data.Parent} {
			if r, p := v.Value(); p && !check(r.Repository()) {
				return repo, invalid("post commit scope")
			}
		}
	}
	return repo, nil
}
func consistentGitPostFacts(d GitPostFactsData) error { _, e := postRepository(d); return e }
func outcomeRepository(v GitMutationOutcome) (domain.RepositoryID, Optional[domain.WorktreeID]) {
	switch x := v.(type) {
	case WorktreeCreated:
		return x.data.Worktree.data.ID.Repository(), Some(x.data.Worktree.data.ID)
	case WorktreeRetargeted:
		return x.data.Worktree.data.ID.Repository(), Some(x.data.Worktree.data.ID)
	case IndexChanged:
		return x.data.Status.data.Worktree.data.ID.Repository(), Some(x.data.Status.data.Worktree.data.ID)
	case CommitCreated:
		return x.data.Revision.Repository(), None[domain.WorktreeID]()
	case TrackedRestored:
		return x.data.Status.data.Worktree.data.ID.Repository(), Some(x.data.Status.data.Worktree.data.ID)
	case StashCreated:
		return x.data.Stash.Repository(), Some(x.data.Status.data.Worktree.data.ID)
	case StashCreatedCleanupRefused:
		if s, p := x.data.Status.Value(); p {
			return x.data.Stash.Repository(), Some(s.data.Worktree.data.ID)
		}
		return x.data.Stash.Repository(), None[domain.WorktreeID]()
	case StashApplied:
		return x.data.Stash.Repository(), Some(x.data.Status.data.Worktree.data.ID)
	case AppliedWithConflicts:
		if s, p := x.data.Status.Value(); p {
			return x.data.Stash.Repository(), Some(s.data.Worktree.data.ID)
		}
		return x.data.Stash.Repository(), None[domain.WorktreeID]()
	case StashDropped:
		return x.data.Stash.Repository(), None[domain.WorktreeID]()
	case BranchCreated:
		if w, p := x.data.Worktree.Value(); p {
			return x.data.Branch.Repository(), Some(w.data.ID)
		}
		return x.data.Branch.Repository(), None[domain.WorktreeID]()
	case Pushed:
		return x.data.Source.Repository(), None[domain.WorktreeID]()
	case PartialMutation:
		r, _ := postRepository(x.data.Facts.data)
		return r, None[domain.WorktreeID]()
	case MutationIndeterminate:
		r, _ := postRepository(x.data.Facts.data)
		return r, None[domain.WorktreeID]()
	}
	return domain.RepositoryID{}, None[domain.WorktreeID]()
}
func consistentGitMutationResult(d GitMutationResultData) error {
	switch x := d.Outcome.(type) {
	case BranchCreated, CommitCreated, StashCreated, StashCreatedCleanupRefused:
		if err := knownChangedFacet(d.Effects, LocalRefsHead, None[EffectReport]()); err != nil {
			return err
		}
	case StashDropped:
		if err := knownChangedFacet(d.Effects, LocalRefsHead, Some(x.data.RefCleanup)); err != nil {
			return err
		}
	}
	r, w := outcomeRepository(d.Outcome)
	if o, p := d.Observation.Value(); p {
		if r.Valid() && o.data.Repository != r {
			return invalid("mutation observation scope")
		}
		if a, p := w.Value(); p {
			if b, p := o.data.Worktree.Value(); p && a != b {
				return invalid("mutation observation worktree")
			}
		}
	}
	for _, step := range d.Steps {
		if o, p := step.data.PostObservation.Value(); p && r.Valid() && o.data.Repository != r {
			return invalid("step observation scope")
		}
	}
	repo, worktree, remotes := mutationRecoveryAssociations(d)
	subjects := make([]RecoverySubject, 0, len(d.Recovery)+len(d.Steps))
	for _, r := range d.Recovery {
		subjects = append(subjects, r.data.Subject)
	}
	for _, s := range d.Steps {
		subjects = append(subjects, s.data.Target)
	}
	return gitSubjects(subjects, repo, worktree, remotes, d.Kind == PushMutation)
}

// Subject returns the independently represented local subject, without deriving
// identity from an OID, label or opaque token. Absence stays explicit.
func (v GitMutationResult) Subject() (domain.RepositoryID, Optional[domain.WorktreeID]) {
	if !v.Valid() {
		return domain.RepositoryID{}, None[domain.WorktreeID]()
	}
	repo, w := outcomeRepository(v.data.Outcome)
	if obs, p := v.data.Observation.Value(); p {
		if !repo.Valid() {
			repo = obs.data.Repository
		}
		if !w.Present() {
			w = obs.data.Worktree
		}
	}
	return repo, w
}
