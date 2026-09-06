package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func resultWorktree(g GitMutationResult, w domain.WorktreeID) error {
	if obs, p := g.data.Observation.Value(); p {
		if actual, p := obs.data.Worktree.Value(); p && actual != w {
			return invalid("result request observed worktree")
		}
	}
	r, id := outcomeRepository(g.data.Outcome)
	if r.Valid() && r != w.Repository() {
		return invalid("result request repository")
	}
	if actual, p := id.Value(); p && actual != w {
		return invalid("result request worktree")
	}
	_, _, remotes := mutationRecoveryAssociations(g.data)
	return gitRecoverySubjects(g.data.Recovery, w.Repository(), Some(w), remotes, g.data.Kind == PushMutation)
}
func requestResultSubject(r Request, result Result) error {
	switch q := r.command.(type) {
	case RetargetWorktreeCommand:
		v := result.(RetargetWorktreeResult).data
		if err := resultWorktree(v.Git, q.data.WorktreeID); err != nil {
			return err
		}
		if x, p := v.Git.data.Outcome.(WorktreeRetargeted); p && x.data.Target.data.Requested != q.data.Target {
			return invalid("retarget request exact target")
		}
	case CreateWorktreeCommand:
		v := result.(CreateWorktreeResult).data
		if x, p := v.Git.data.Outcome.(WorktreeCreated); p && x.data.Target.data.Requested != q.data.Target {
			return invalid("create request exact target")
		}
	case DeployCommand:
		v := result.(DeployResult).data
		if v.Target != q.data.Target {
			return invalid("deploy request exact target")
		}
		if dst, p := q.data.Destination.(WorktreeDestination); p {
			for _, g := range v.Steps {
				if err := resultWorktree(g, dst.data.Worktree); err != nil {
					return err
				}
			}
		}
	case CreateBranchCommand:
		v := result.(CreateBranchResult).data
		if err := resultWorktree(v.Git, q.data.WorktreeID); err != nil {
			return err
		}
		if x, p := v.Git.data.Outcome.(BranchCreated); p && (x.data.Branch != q.data.Name || x.data.Revision != q.data.Start) {
			return invalid("branch request identity")
		}
	case PullCommand:
		if g, p := result.(PullResult).data.FastForward.Value(); p {
			return resultWorktree(g, q.data.WorktreeID)
		}
	case PushCommand:
		g := result.(PushResult).data.Git
		if err := resultWorktree(g, q.data.WorktreeID); err != nil {
			return err
		}
		if err := g.ValidatePushBinding(q.data.Binding, q.data.Destination); err != nil {
			return err
		}
		if p, ok := g.data.Outcome.(Pushed); ok && (p.data.Source != q.data.Source || p.data.Destination != q.data.Destination) {
			return invalid("push request endpoints")
		}
	case StagePathsCommand:
		return resultWorktree(result.(StagePathsResult).data.Git, q.data.WorktreeID)
	case UnstagePathsCommand:
		return resultWorktree(result.(UnstagePathsResult).data.Git, q.data.WorktreeID)
	case StageAllCommand:
		return resultWorktree(result.(StageAllResult).data.Git, q.data.WorktreeID)
	case CommitCommand:
		return resultWorktree(result.(CommitResult).data.Git, q.data.WorktreeID)
	case StageAllAndCommitCommand:
		v := result.(StageAllAndCommitResult).data
		for _, o := range []Optional[GitMutationResult]{v.Stage, v.Commit} {
			if g, p := o.Value(); p {
				if err := resultWorktree(g, q.data.WorktreeID); err != nil {
					return err
				}
			}
		}
	case RestoreTrackedCommand:
		return resultWorktree(result.(RestoreTrackedResult).data.Git, q.data.WorktreeID)
	case StashCreateCommand:
		return resultWorktree(result.(StashCreateResult).data.Git, q.data.WorktreeID)
	case StashApplyCommand:
		v := result.(StashApplyResult).data
		if v.Stash != q.data.Stash {
			return invalid("stash apply request identity")
		}
		return resultWorktree(v.Git, q.data.WorktreeID)
	case StashDropCommand:
		if result.(StashDropResult).data.Stash != q.data.Stash {
			return invalid("stash drop request identity")
		}
	case StashPopCommand:
		v := result.(StashPopResult).data
		var id domain.StashID
		var apply GitMutationResult
		switch x := v.Outcome.(type) {
		case StashPopCompleted:
			id = x.data.Stash
			apply = x.data.Apply
		case AppliedStashRetained:
			id = x.data.Stash
			apply = x.data.Apply
		case StashPopNotApplied:
			id = x.data.Stash
			apply = x.data.Apply
		}
		if id != q.data.Stash {
			return invalid("pop request identity")
		}
		return resultWorktree(apply, q.data.WorktreeID)
	case StartLaunchCommand:
		if s, p := result.(StartLaunchResult).data.Start.data.Session.Value(); p && s.data.WorktreeID != q.data.WorktreeID {
			return invalid("start request worktree")
		}
	case OpenTerminalCommand:
		if s, p := result.(OpenTerminalResult).data.Start.data.Session.Value(); p && s.data.WorktreeID != q.data.WorktreeID {
			return invalid("terminal request worktree")
		}
	case WriteInputCommand:
		v := result.(WriteInputResult).data.Write.data
		if v.SessionID != q.data.SessionID || v.AcceptedBytes != 0 && int(v.AcceptedBytes) != len(q.data.Bytes) {
			return invalid("write request identity/whole-buffer admission")
		}
	case ResizeSessionCommand:
		if result.(ResizeSessionResult).data.Control.data.SessionID != q.data.SessionID {
			return invalid("resize request session")
		}
	case InterruptSessionCommand:
		if result.(InterruptSessionResult).data.Control.data.SessionID != q.data.SessionID {
			return invalid("interrupt request session")
		}
	case StopSessionCommand:
		if result.(StopSessionResult).data.Stop.data.Session.data.SessionID != q.data.SessionID {
			return invalid("stop request session")
		}
	case RestartSessionCommand:
		if result.(RestartSessionResult).data.Restart.data.Old.data.Session.data.SessionID != q.data.SessionID {
			return invalid("restart request old session")
		}
	}
	switch q := r.query.(type) {
	case NavigatorQuery:
		if result.(NavigatorResult).data.Repository != q.data.Repository {
			return invalid("navigator query scope")
		}
	case BranchContextQuery:
		if result.(BranchContextResult).data.Target != q.data.Target {
			return invalid("branch query target")
		}
	case CommitsQuery:
		if result.(CommitsResult).data.Target != q.data.Target {
			return invalid("commits query target")
		}
	case GraphQuery:
		if result.(GraphResult).data.Repository != q.data.Repository {
			return invalid("graph query repository")
		}
		if !sameRevisions(result.(GraphResult).data.Roots, q.data.Roots) {
			return invalid("graph query exact roots")
		}
	case DiffQuery:
		if !sameGitComparison(q.data.Comparison, result.(DiffResult).data.Comparison) {
			return invalid("diff query exact comparison")
		}
	case PullRequestDiffQuery:
		if result.(PullRequestDiffResult).data.Target != q.data.Target {
			return invalid("PR diff query target")
		}
		if !samePRBase(q.data.Base, result.(PullRequestDiffResult).data.RequestedBase) {
			return invalid("PR diff requested base")
		}
		for _, o := range []Optional[ExactLocalResolution]{result.(PullRequestDiffResult).data.Base, result.(PullRequestDiffResult).data.Head} {
			if v, p := o.Value(); p && v.data.Local.Repository() != q.data.Local {
				return invalid("PR diff requested local object scope")
			}
		}
	case WorktreeStatusQuery:
		if result.(WorktreeStatusResult).data.WorktreeID != q.data.WorktreeID {
			return invalid("status query worktree")
		}
	case StashesQuery:
		if result.(StashesResult).data.Repository != q.data.Repository {
			return invalid("stashes query repository")
		}
	case StashPatchQuery:
		if result.(StashPatchResult).data.Stash != q.data.Stash {
			return invalid("stash patch query identity")
		}
		if c, p := result.(StashPatchResult).data.Comparison.Value(); p && !sameStashView(q.data.View, c.data.View) {
			return invalid("stash patch query exact view")
		}
	case LaunchPointsQuery:
		if result.(LaunchPointsResult).data.WorktreeID != q.data.WorktreeID {
			return invalid("launch query worktree")
		}
	case SessionsQuery:
		if w, p := q.data.WorktreeID.Value(); p {
			for _, s := range result.(SessionsResult).data.Sessions.data.Sessions {
				if s.data.WorktreeID != w {
					return invalid("sessions query scope")
				}
			}
		}
	case SessionOutputQuery:
		if result.(SessionOutputProjection).data.Output.data.SessionID != q.data.SessionID {
			return invalid("output query session")
		}
	case PreferencesQuery:
		if result.(PreferencesResult).data.Effective.data.Repository != q.data.Repository {
			return invalid("preferences query scope")
		}
	}
	return nil
}
