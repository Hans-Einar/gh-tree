package api

func publicResultOperations(v Result) []OperationID {
	var ids []OperationID
	add := func(g GitMutationResult) { ids = append(ids, g.data.Operation) }
	start := func(s SessionStartResult) {
		if x, p := s.data.Session.Value(); p {
			ids = append(ids, x.data.StartOperation)
		}
	}
	switch d := v.(type) {
	case CreateWorktreeResult:
		add(d.data.Git)
	case RetargetWorktreeResult:
		add(d.data.Git)
	case DeployResult:
		for _, g := range d.data.Steps {
			add(g)
		}
	case CreateBranchResult:
		add(d.data.Git)
	case PullResult:
		if g, p := d.data.FastForward.Value(); p {
			add(g)
		}
	case PushResult:
		add(d.data.Git)
	case StagePathsResult:
		add(d.data.Git)
	case UnstagePathsResult:
		add(d.data.Git)
	case StageAllResult:
		add(d.data.Git)
	case CommitResult:
		add(d.data.Git)
	case StageAllAndCommitResult:
		if g, p := d.data.Stage.Value(); p {
			add(g)
		}
		if g, p := d.data.Commit.Value(); p {
			add(g)
		}
	case RestoreTrackedResult:
		add(d.data.Git)
	case StashCreateResult:
		add(d.data.Git)
	case StashApplyResult:
		add(d.data.Git)
	case StashDropResult:
		add(d.data.Git)
	case StashPopResult:
		switch p := d.data.Outcome.(type) {
		case StashPopCompleted:
			add(p.data.Apply)
			add(p.data.Drop)
		case AppliedStashRetained:
			add(p.data.Apply)
			if g, p := p.data.Drop.Value(); p {
				add(g)
			}
		case StashPopNotApplied:
			add(p.data.Apply)
		}
	case StartLaunchResult:
		start(d.data.Start)
	case OpenTerminalResult:
		start(d.data.Start)
	case RestartSessionResult:
		if s, p := d.data.Restart.data.Replacement.Value(); p {
			start(s)
		}
	case CreatePullRequestResult:
		if c, p := d.data.Outcome.(CreationIndeterminate); p {
			ids = append(ids, c.data.RequestEvidence.data.OperationID)
		}
	}
	return ids
}
func consistentOperationTerminal(d OperationTerminalData) error {
	if r, p := d.Result.Value(); p {
		for _, id := range publicResultOperations(r) {
			if id != d.OperationID {
				return invalid("terminal result operation")
			}
		}
	}
	return nil
}
