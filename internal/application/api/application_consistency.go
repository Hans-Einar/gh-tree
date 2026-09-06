package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func mutationKind(g GitMutationResult, k GitMutationKind) error {
	if g.data.Kind != k {
		return invalid("public mutation result kind")
	}
	return nil
}
func sameMutationOperation(a, b GitMutationResult) bool { return a.data.Operation == b.data.Operation }
func mutationStash(g GitMutationResult) (domain.StashID, bool) {
	switch s := g.data.Outcome.(type) {
	case StashCreated:
		return s.data.Stash, true
	case StashCreatedCleanupRefused:
		return s.data.Stash, true
	case StashApplied:
		return s.data.Stash, true
	case AppliedWithConflicts:
		return s.data.Stash, true
	case StashDropped:
		return s.data.Stash, true
	}
	return domain.StashID{}, false
}
func stashResult(g GitMutationResult, s domain.StashID) error {
	if err := mutationKind(g, StashMutation); err != nil {
		return err
	}
	if id, p := mutationStash(g); p && id != s {
		return invalid("public stash result identity")
	}
	r, _ := outcomeRepository(g.data.Outcome)
	if r.Valid() && r != s.Repository() {
		return invalid("public stash result repository")
	}
	return nil
}
func consistentStashPopCompleted(d StashPopCompletedData) error {
	if err := stashResult(d.Apply, d.Stash); err != nil {
		return err
	}
	if err := stashResult(d.Drop, d.Stash); err != nil {
		return err
	}
	if !sameMutationOperation(d.Apply, d.Drop) || d.Apply.data.CancellationRequested {
		return invalid("pop operation/continuation")
	}
	if _, p := d.Apply.data.Outcome.(StashApplied); !p {
		return invalid("pop verified apply")
	}
	if _, p := d.Drop.data.Outcome.(StashDropped); !p {
		return invalid("pop verified drop")
	}
	return nil
}
func consistentAppliedStashRetained(d AppliedStashRetainedData) error {
	if err := stashResult(d.Apply, d.Stash); err != nil {
		return err
	}
	if _, p := d.Apply.data.Outcome.(StashApplied); !p {
		return invalid("retained stash verified apply")
	}
	if drop, p := d.Drop.Value(); p {
		if err := stashResult(drop, d.Stash); err != nil {
			return err
		}
		if !sameMutationOperation(d.Apply, drop) {
			return invalid("retained stash operation")
		}
		if _, p := drop.data.Outcome.(StashDropped); p {
			return invalid("dropped stash reported retained")
		}
	}
	return nil
}
func consistentStashPopNotApplied(d StashPopNotAppliedData) error {
	if err := stashResult(d.Apply, d.Stash); err != nil {
		return err
	}
	switch d.Apply.data.Outcome.(type) {
	case StashApplied, StashCreated, StashCreatedCleanupRefused, StashDropped:
		return invalid("incomplete pop outcome")
	}
	return nil
}
func consistentActivateWorktreeResult(d ActivateWorktreeResultData) error {
	if v, p := d.Context.data.PreferenceVersion.Value(); p {
		if current, p := d.Storage.data.CurrentVersion.Value(); p && !v.SameBinding(current) {
			return invalid("active storage subject")
		}
		if proposed, p := d.Storage.data.ProposedVersion.Value(); p && !v.SameBinding(proposed) {
			return invalid("active storage proposal subject")
		}
	}
	return nil
}
func consistentSaveNavigationResult(d SaveNavigationResultData) error {
	if v, p := d.EffectiveVersion.Value(); p {
		if v.Family() != Preferences {
			return invalid("navigation storage family")
		}
		for _, o := range []Optional[StorageVersion]{d.Storage.data.ProposedVersion, d.Storage.data.CurrentVersion} {
			if x, p := o.Value(); p && !v.SameBinding(x) {
				return invalid("navigation storage binding")
			}
		}
	}
	return nil
}
func consistentCreateWorktreeResult(d CreateWorktreeResultData) error {
	if err := mutationKind(d.Git, CreateMutation); err != nil {
		return err
	}
	if x, p := d.Git.data.Outcome.(WorktreeCreated); p {
		return publicWorktreePost(d.WorktreeID, d.Scope, d.Head, x.data.Worktree)
	}
	if id, p := d.WorktreeID.Value(); p {
		if s, p := d.Scope.Value(); p && s.data.ID != id {
			return invalid("create result scope")
		}
		if h, p := d.Head.Value(); p && !h.MatchesWorktree(id) {
			return invalid("create result head")
		}
	}
	return nil
}
func publicWorktreePost(id Optional[domain.WorktreeID], scope Optional[WorktreeScope], head Optional[domain.Head], facts WorktreeFacts) error {
	if w, p := id.Value(); p && w != facts.data.ID {
		return invalid("public worktree identity")
	}
	if s, p := scope.Value(); p {
		actual, p := facts.data.Scope.Value()
		if !p || s.data.ID != actual.data.ID || s.data.RootIdentity != actual.data.RootIdentity {
			return invalid("public worktree scope")
		}
	}
	if h, p := head.Value(); p {
		actual, p := facts.data.Head.Value()
		if !p || h != actual {
			return invalid("public worktree head")
		}
	}
	return nil
}
func consistentRetargetWorktreeResult(d RetargetWorktreeResultData) error {
	if err := mutationKind(d.Git, RetargetMutation); err != nil {
		return err
	}
	if x, p := d.Git.data.Outcome.(WorktreeRetargeted); p {
		return publicWorktreePost(None[domain.WorktreeID](), d.Scope, d.Head, x.data.Worktree)
	}
	if s, p := d.Scope.Value(); p {
		if h, p := d.Head.Value(); p && !h.MatchesWorktree(s.data.ID) {
			return invalid("retarget public scope")
		}
	}
	return nil
}
func consistentDeployResult(d DeployResultData) error {
	if len(d.Steps) > 2 {
		return invalid("deploy step count")
	}
	if len(d.Steps) == 2 {
		first, second := d.Steps[0], d.Steps[1]
		if first.data.Kind != StashMutation || second.data.Kind != RetargetMutation || first.data.CancellationRequested {
			return invalid("deploy continuation order")
		}
		if _, p := first.data.Outcome.(StashCreated); !p {
			return invalid("deploy after unsuccessful stash cleanup")
		}
		_, a := outcomeRepository(first.data.Outcome)
		_, b := outcomeRepository(second.data.Outcome)
		if x, p := a.Value(); p {
			if y, p := b.Value(); p && x != y {
				return invalid("deploy continuation worktree")
			}
		}
	}

	if r, p := d.Resolution.Value(); p {
		if r.data.Requested != d.Target {
			return invalid("deploy exact target")
		}
		if h, p := d.Head.Value(); p && h.Repository() != r.data.Local.Repository() {
			return invalid("deploy resulting head scope")
		}
	}
	var prior Optional[GitMutationResult]
	for _, g := range d.Steps {
		if g.data.Kind != StashMutation && g.data.Kind != RetargetMutation {
			return invalid("deploy step kind")
		}
		if p, ok := prior.Value(); ok && !sameMutationOperation(p, g) {
			return invalid("deploy step operation")
		}
		prior = Some(g)
		if x, p := g.data.Outcome.(WorktreeRetargeted); p {
			if x.data.Target.data.Requested != d.Target {
				return invalid("deploy step target")
			}
			if h, p := d.Head.Value(); p {
				actual, _ := x.data.Worktree.data.Head.Value()
				if h != actual {
					return invalid("deploy head step")
				}
			}
		}
		if s, p := d.CreatedStash.Value(); p {
			if got, p := mutationStash(g); p && got != s {
				return invalid("deploy retained stash")
			}
		}
	}
	return nil
}
func headRevision(h domain.Head, r domain.Revision) bool { x, p := h.Revision(); return p && x == r }
func consistentCreateBranchResult(d CreateBranchResultData) error {
	if err := mutationKind(d.Git, BranchMutation); err != nil {
		return err
	}
	if x, p := d.Git.data.Outcome.(BranchCreated); p {
		if b, p := d.Branch.Value(); p && b != x.data.Branch {
			return invalid("created branch label")
		}
		if h, p := d.Head.Value(); p {
			b, bp := h.Branch()
			if !bp || b != x.data.Branch || !headRevision(h, x.data.Revision) {
				return invalid("created branch head")
			}
		}
	}
	return nil
}
func consistentPullResult(d PullResultData) error {
	if g, p := d.FastForward.Value(); p {
		if err := mutationKind(g, RetargetMutation); err != nil {
			return err
		}
		if x, p := g.data.Outcome.(WorktreeRetargeted); p {
			fetch, fp := d.Fetch.Value()
			if !fp || !fetch.data.Generation.Present() || fetch.data.Freshness.data.Kind != Refreshed {
				return invalid("pull fast-forward lacks completed fetch")
			}
			if endpoint, p := d.Endpoint.Value(); p && endpoint != x.data.Target.data.Local {
				return invalid("pull exact endpoint")
			}
		}
	}
	return nil
}

// Fetch has no echoed Operation field in its frozen result; provenance is carried
// by the request/adapter. Do not fabricate operation equality from generations.

func consistentPushResult(d PushResultData) error { return mutationKind(d.Git, PushMutation) }
func stageResult(g GitMutationResult, s Optional[StatusFacts], action StageAction) error {
	if err := mutationKind(g, StageMutation); err != nil {
		return err
	}
	if x, p := g.data.Outcome.(IndexChanged); p {
		if x.data.Action != action {
			return invalid("stage action result")
		}
		if status, p := s.Value(); p && (status.data.Worktree.data.ID != x.data.Status.data.Worktree.data.ID || status.data.IndexVersion != x.data.Status.data.IndexVersion || status.data.WorktreeVersion != x.data.Status.data.WorktreeVersion) {
			return invalid("stage status duplicate")
		}
	}
	return nil
}
func consistentStagePathsResult(d StagePathsResultData) error {
	return stageResult(d.Git, d.Status, Stage)
}
func consistentUnstagePathsResult(d UnstagePathsResultData) error {
	return stageResult(d.Git, d.Status, Unstage)
}
func consistentStageAllResult(d StageAllResultData) error { return stageResult(d.Git, d.Status, Stage) }
func commitResult(g GitMutationResult, r Optional[domain.Revision], c Optional[CommitCandidateFacts]) error {
	if err := mutationKind(g, CommitMutation); err != nil {
		return err
	}
	if x, p := g.data.Outcome.(CommitCreated); p {
		if v, p := r.Value(); p && v != x.data.Revision {
			return invalid("commit public revision")
		}
	}
	if x, p := c.Value(); p {
		if candidate, p := x.data.Candidate.Value(); p {
			if revision, p := r.Value(); p && revision != candidate {
				return invalid("public commit candidate")
			}
		}
	}
	return nil
}
func consistentCommitResult(d CommitResultData) error {
	return commitResult(d.Git, d.Revision, d.Candidate)
}
func consistentStageAllAndCommitResult(d StageAllAndCommitResultData) error {
	if stage, p := d.Stage.Value(); p {
		if err := stageResult(stage, None[StatusFacts](), Stage); err != nil {
			return err
		}
	}
	if commit, p := d.Commit.Value(); p {
		if err := commitResult(commit, d.Revision, d.Candidate); err != nil {
			return err
		}
		if stage, p := d.Stage.Value(); p {
			if !sameMutationOperation(stage, commit) || stage.data.CancellationRequested {
				return invalid("stage/commit continuation")
			}
			_, worktree := outcomeRepository(stage.data.Outcome)
			if w, p := worktree.Value(); p {
				if err := resultWorktree(commit, w); err != nil {
					return err
				}
			}
			if _, p := stage.data.Outcome.(IndexChanged); !p {
				return invalid("commit after failed stage")
			}
		} else {
			return invalid("commit without staged predecessor")
		}
	}
	return nil
}
func consistentRestoreTrackedResult(d RestoreTrackedResultData) error {
	if err := mutationKind(d.Git, RestoreMutation); err != nil {
		return err
	}
	if x, p := d.Git.data.Outcome.(TrackedRestored); p {
		if s, p := d.Status.Value(); p && s.data.Worktree.data.ID != x.data.Status.data.Worktree.data.ID {
			return invalid("restore status subject")
		}
	}
	return nil
}
func consistentStashCreateResult(d StashCreateResultData) error {
	switch d.Git.data.Outcome.(type) {
	case StashApplied, AppliedWithConflicts, StashDropped:
		return invalid("StashCreate outcome alternative")
	}
	if err := mutationKind(d.Git, StashMutation); err != nil {
		return err
	}
	if s, p := d.Stash.Value(); p {
		return stashResult(d.Git, s)
	}
	return nil
}
func consistentStashApplyResult(d StashApplyResultData) error {
	switch d.Git.data.Outcome.(type) {
	case StashCreated, StashCreatedCleanupRefused, StashDropped:
		return invalid("StashApply outcome alternative")
	}
	return stashResult(d.Git, d.Stash)
}
func consistentStashDropResult(d StashDropResultData) error {
	switch d.Git.data.Outcome.(type) {
	case StashCreated, StashCreatedCleanupRefused, StashApplied, AppliedWithConflicts:
		return invalid("StashDrop outcome alternative")
	}
	return stashResult(d.Git, d.Stash)
}
func consistentStashPopResult(StashPopResultData) error { return nil }
func consistentSaveLaunchResult(d SaveLaunchResultData) error {
	if d.Default.Presence() == FieldNull {
		return invalid("saved default null")
	}
	for _, o := range []Optional[StorageVersion]{d.Storage.data.ProposedVersion, d.Storage.data.CurrentVersion} {
		if v, p := o.Value(); p && v.Family() != RunConfig {
			return invalid("saved launch storage family")
		}
	}
	return nil
}
func consistentStartLaunchResult(StartLaunchResultData) error { return nil }
func consistentOpenTerminalResult(d OpenTerminalResultData) error {
	if s, p := d.Start.data.Session.Value(); p && s.data.Display.data.Terminal != Terminal {
		return invalid("opened nonterminal")
	}
	return nil
}
func consistentWriteInputResult(WriteInputResultData) error             { return nil }
func consistentResizeSessionResult(ResizeSessionResultData) error       { return nil }
func consistentInterruptSessionResult(InterruptSessionResultData) error { return nil }
func consistentStopSessionResult(StopSessionResultData) error           { return nil }
func consistentRestartSessionResult(RestartSessionResultData) error     { return nil }
