package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func consistentNavigatorResult(d NavigatorResultData) error {
	if d.Repository.Scope() != domain.Remote {
		return invalid("navigator repository")
	}
	for _, p := range d.PullRequests {
		if p.data.ID.Repository() != d.Repository {
			return invalid("navigator PR scope")
		}
	}
	for _, b := range d.Branches {
		if b.data.Branch.Repository() != d.Repository {
			return invalid("navigator branch scope")
		}
	}
	return nil
}
func endpointProjection(t domain.ExactTarget, res Optional[ExactLocalResolution], commits []CommitFact) error {
	r, p := res.Value()
	if p && r.data.Requested != t {
		return invalid("projection exact target")
	}
	if len(commits) > 0 && !p {
		return invalid("history missing verified endpoint")
	}
	for _, c := range commits {
		if c.data.Revision.Repository() != r.data.Local.Repository() {
			return invalid("projection history repository")
		}
	}
	return nil
}
func consistentBranchContextResult(d BranchContextResultData) error {
	if err := endpointProjection(d.Target, d.Endpoint, d.Commits); err != nil {
		return err
	}
	if relation, p := d.Relationship.Value(); p {
		if relation.data.ExpectedRevision != d.Target.ExpectedRevision() {
			return invalid("branch context expected revision")
		}
		if b, p := d.Target.Branch(); p && relation.data.Branch != b {
			return invalid("branch context identity")
		}
	}
	return nil
}
func consistentCommitsResult(d CommitsResultData) error {
	return endpointProjection(d.Target, d.Endpoint, d.Commits)
}
func consistentGraphResult(d GraphResultData) error {
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("graph local repository")
	}
	for _, r := range d.Roots {
		if r.Repository() != d.Repository {
			return invalid("graph roots scope")
		}
	}
	for _, c := range d.Commits {
		if c.data.Revision.Repository() != d.Repository {
			return invalid("graph commit scope")
		}
	}
	for _, e := range d.Edges {
		if e.data.Child.Repository() != d.Repository {
			return invalid("graph edge scope")
		}
	}
	for _, r := range d.Refs {
		if r.data.Observation.data.Repository != d.Repository {
			return invalid("graph ref scope")
		}
	}
	for _, a := range d.Annotations {
		if a.data.Revision.Repository() != d.Repository {
			return invalid("graph annotation scope")
		}
	}
	return nil
}
func consistentDiffResult(DiffResultData) error { return nil }
func consistentPullRequestDiffResult(d PullRequestDiffResultData) error {
	if d.Target.Kind() != domain.PullRequestTarget {
		return invalid("PR diff target kind")
	}
	id, _ := d.Target.PullRequest()
	if pr, p := d.RemotePR.Value(); p && pr.data.ID != id {
		return invalid("PR diff identity")
	}
	if h, p := d.Head.Value(); p && h.data.Requested.ExpectedRevision() != d.Target.ExpectedRevision() {
		return invalid("PR diff exact head")
	}
	if b, p := d.Base.Value(); p {
		if exact, p := d.RequestedBase.(ExactPRBase); p && b.data.Requested.ExpectedRevision() != exact.data.Revision {
			return invalid("PR diff exact base")
		}
		if h, p := d.Head.Value(); p && !sameLocal(b.data.Local, h.data.Local) {
			return invalid("PR diff local association")
		}
	}
	if m, p := d.MergeBase.Value(); p {
		b, bp := d.Base.Value()
		h, hp := d.Head.Value()
		if !bp || !hp || !sameLocal(m, b.data.Local) || !sameLocal(m, h.data.Local) {
			return invalid("PR diff merge base")
		}
	}
	if d.Patch.Present() && (!d.Base.Present() || !d.Head.Present() || !d.MergeBase.Present()) {
		return invalid("PR diff patch endpoints")
	}
	return nil
}
func consistentWorktreeStatusResult(d WorktreeStatusResultData) error {
	if s, p := d.Status.Value(); p && s.data.Worktree.data.ID != d.WorktreeID {
		return invalid("status projection worktree")
	}
	return nil
}
func consistentStashesResult(d StashesResultData) error {
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("stash projection repository")
	}
	for _, s := range d.Stashes {
		if s.data.ID.Repository() != d.Repository {
			return invalid("stash projection subject")
		}
	}
	return nil
}
func consistentStashPatchResult(d StashPatchResultData) error {
	if c, p := d.Comparison.Value(); p && c.data.Stash != d.Stash.OID() {
		return invalid("stash patch exact identity")
	}
	if d.Patch.Present() && !d.Comparison.Present() {
		return invalid("stash patch comparison")
	}
	return nil
}
func consistentLaunchPointsResult(d LaunchPointsResultData) error {
	for _, v := range d.Definitions {
		if v.data.LaunchPointID.Worktree() != d.WorktreeID {
			return invalid("launch projection subject")
		}
	}
	for _, s := range d.Saved {
		if !versionWorktree(s.data.StorageVersion, d.WorktreeID) {
			return invalid("saved projection subject")
		}
	}
	if v, p := d.StorageVersion.Value(); p && !versionWorktree(v, d.WorktreeID) {
		return invalid("launch projection storage")
	}
	if id, p := d.DefaultID.Value(); p && id.Worktree() != d.WorktreeID {
		return invalid("default projection subject")
	}
	if d.DefaultAlias.Presence() == FieldNull {
		return invalid("default alias null")
	}
	return nil
}
func consistentSessionsResult(SessionsResultData) error                   { return nil }
func consistentSessionOutputProjection(SessionOutputProjectionData) error { return nil }
func consistentPreferencesResult(d PreferencesResultData) error {
	if d.ContextVersion != d.Active.data.Version {
		return invalid("preferences captured context")
	}
	return nil
}
