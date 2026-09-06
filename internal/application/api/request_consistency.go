package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func versionWorktree(v StorageVersion, w domain.WorktreeID) bool {
	id, p := v.Worktree().Value()
	return p && id == w
}
func savedScope(v Optional[StorageVersion], s WorktreeScope) bool {
	x, p := v.Value()
	return !p || x.MatchesRunScope(s)
}
func localTargetScope(t domain.ExactTarget, r domain.RepositoryID) bool {
	return t.Valid() && (t.ExpectedRevision().Repository().Scope() == domain.Remote || t.ExpectedRevision().Repository() == r)
}
func consistentSavedLaunch(d SavedLaunchData) error {
	if !versionWorktree(d.StorageVersion, d.LaunchPointID.Worktree()) {
		return invalid("saved version worktree")
	}
	return nil
}
func consistentSavedLaunchObservation(d SavedLaunchObservationData) error {
	if id, p := d.LaunchPointID.Value(); p && !versionWorktree(d.StorageVersion, id.Worktree()) {
		return invalid("saved observation worktree")
	}
	if v, p := d.Definition.Value(); p {
		if !versionWorktree(d.StorageVersion, v.data.LaunchPointID.Worktree()) {
			return invalid("saved definition worktree")
		}
		if id, p := d.LaunchPointID.Value(); p && id != v.data.LaunchPointID {
			return invalid("saved definition identity")
		}
	}
	return nil
}
func consistentDiscoveryRequest(d DiscoveryRequestData) error {
	if !savedScope(d.SavedVersion, d.Worktree) {
		return invalid("discovery storage scope")
	}
	return nil
}
func consistentDiscoveryResult(d DiscoveryResultData) error {
	for _, s := range d.Saved {
		if !versionWorktree(s.data.StorageVersion, d.WorktreeID) {
			return invalid("discovery saved scope")
		}
	}
	return nil
}
func consistentResolveLaunchRequest(d ResolveLaunchRequestData) error {
	if !savedScope(d.SavedVersion, d.Worktree) {
		return invalid("resolve storage scope")
	}
	if s, p := d.Selection.(SavedLaunch); p && !s.data.StorageVersion.MatchesRunScope(d.Worktree) {
		return invalid("saved selection root")
	}
	return nil
}
func consistentResolveLaunchResult(d ResolveLaunchResultData) error {
	if def, p := d.Definition.Value(); p {
		for _, m := range def.data.Selected {
			if m.data.LaunchPointID.Worktree() != d.Observation.data.WorktreeID {
				return invalid("resolved definition subject")
			}
		}
		if inv, p := d.Invocation.Value(); p {
			if !sameStrings(def.data.ProjectComponents, inv.data.Cwd.data.ProjectComponents) || def.data.ProjectSource.data.RootIdentity != inv.data.Cwd.data.Worktree.data.RootIdentity || def.data.ProjectSource.data.ProjectIdentity != inv.data.Cwd.data.ProjectIdentity {
				return invalid("resolved project cwd")
			}
			if v, p := def.data.SavedVersion.Value(); p && !v.MatchesRunScope(inv.data.Cwd.data.Worktree) {
				return invalid("resolved saved root")
			}
		}
	}
	return nil
}
func consistentResolvedLaunchDefinition(d ResolvedLaunchDefinitionData) error {
	if len(d.Selected) == 0 {
		return nil
	}
	w := d.Selected[0].data.LaunchPointID.Worktree()
	for _, m := range d.Selected {
		if m.data.LaunchPointID.Worktree() != w {
			return invalid("resolved member scope")
		}
	}
	if v, p := d.SavedVersion.Value(); p && !versionWorktree(v, w) {
		return invalid("resolved saved worktree")
	}
	return nil
}
func consistentSaveLaunchCommand(d SaveLaunchCommandData) error {
	if !versionWorktree(d.ExpectedStorage, d.WorktreeID) {
		return invalid("save launch worktree")
	}
	if s, p := d.Selection.(SavedLaunch); p && !s.data.StorageVersion.SameBinding(d.ExpectedStorage) {
		return invalid("save selection storage binding")
	}
	return nil
}
func consistentStartLaunchCommand(d StartLaunchCommandData) error {
	switch x := d.Selection.(type) {
	case CurrentDefaultLaunch:
		if v, p := x.data.ExpectedStorage.Value(); p && !versionWorktree(v, d.WorktreeID) {
			return invalid("default launch worktree")
		}
	case SelectedLaunch:
		if !launchMatchesWorktree(x.data.Selection, d.WorktreeID) {
			return invalid("selected launch worktree")
		}
	}
	return nil
}
func consistentRetargetWorktreeCommand(d RetargetWorktreeCommandData) error {
	if !localTargetScope(d.Target, d.WorktreeID.Repository()) {
		return invalid("retarget local target scope")
	}
	return nil
}
func consistentCreateWorktreeCommand(d CreateWorktreeCommandData) error {
	if branch, p := d.Mode.(CreateNewBranch); p && !localTargetScope(d.Target, branch.data.Branch.Repository()) {
		return invalid("create local target branch scope")
	}
	return nil
}
func consistentDeployCommand(d DeployCommandData) error {
	switch dst := d.Destination.(type) {
	case WorktreeDestination:
		if !localTargetScope(d.Target, dst.data.Worktree.Repository()) || !retargetModeScope(d.Mode, dst.data.Worktree.Repository(), None[domain.Revision]()) {
			return invalid("deploy worktree binding")
		}
	case ConfiguredDestination:
		if w, p := dst.data.Worktree.Value(); p && (!localTargetScope(d.Target, w.Repository()) || !retargetModeScope(d.Mode, w.Repository(), None[domain.Revision]())) {
			return invalid("configured deploy binding")
		}
	}
	return nil
}
func consistentPullCommand(d PullCommandData) error {
	h, p := d.Expected.data.Head.Value()
	if !p || h != d.Head || d.Upstream.data.Binding.data.LocalRepository != d.WorktreeID.Repository() {
		return invalid("pull expected head/upstream")
	}
	r, p := d.Head.Revision()
	if !p || r != d.Upstream.data.Local {
		return invalid("pull exact local endpoint")
	}
	return nil
}
func consistentPushCommand(d PushCommandData) error {
	if s, p := d.SetUpstream.Value(); p {
		h, hp := d.Expected.data.Head.Value()
		b, bp := h.Branch()
		if !hp || !bp || b != s.data.Branch || s.data.ExpectedConfiguration != d.Expected.data.Configuration {
			return invalid("push upstream precondition")
		}
	}
	return nil
}
func consistentCreateBranchCommand(d CreateBranchCommandData) error {
	if d.Name.Kind() != domain.Local {
		return invalid("new branch kind")
	}
	return nil
}
func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
