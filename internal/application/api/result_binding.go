package api

// AcceptsResult checks the closed request/result family, independently of effect disposition.
func (r Request) AcceptsResult(result Result) bool {
	if !r.Valid() || !validResult(result) {
		return false
	}
	if r.command != nil {
		switch r.command.(type) {
		case ActivateWorktreeCommand:
			_, ok := result.(ActivateWorktreeResult)
			return ok
		case SaveNavigationCommand:
			_, ok := result.(SaveNavigationResult)
			return ok
		case CreateWorktreeCommand:
			_, ok := result.(CreateWorktreeResult)
			return ok
		case RetargetWorktreeCommand:
			_, ok := result.(RetargetWorktreeResult)
			return ok
		case DeployCommand:
			_, ok := result.(DeployResult)
			return ok
		case CreateBranchCommand:
			_, ok := result.(CreateBranchResult)
			return ok
		case FetchCommand:
			_, ok := result.(FetchResult)
			return ok
		case PullCommand:
			_, ok := result.(PullResult)
			return ok
		case PushCommand:
			_, ok := result.(PushResult)
			return ok
		case StagePathsCommand:
			_, ok := result.(StagePathsResult)
			return ok
		case UnstagePathsCommand:
			_, ok := result.(UnstagePathsResult)
			return ok
		case StageAllCommand:
			_, ok := result.(StageAllResult)
			return ok
		case CommitCommand:
			_, ok := result.(CommitResult)
			return ok
		case StageAllAndCommitCommand:
			_, ok := result.(StageAllAndCommitResult)
			return ok
		case RestoreTrackedCommand:
			_, ok := result.(RestoreTrackedResult)
			return ok
		case StashCreateCommand:
			_, ok := result.(StashCreateResult)
			return ok
		case StashApplyCommand:
			_, ok := result.(StashApplyResult)
			return ok
		case StashPopCommand:
			_, ok := result.(StashPopResult)
			return ok
		case StashDropCommand:
			_, ok := result.(StashDropResult)
			return ok
		case CreatePullRequestCommand:
			_, ok := result.(CreatePullRequestResult)
			return ok
		case SaveLaunchCommand:
			_, ok := result.(SaveLaunchResult)
			return ok
		case StartLaunchCommand:
			_, ok := result.(StartLaunchResult)
			return ok
		case OpenTerminalCommand:
			_, ok := result.(OpenTerminalResult)
			return ok
		case WriteInputCommand:
			_, ok := result.(WriteInputResult)
			return ok
		case ResizeSessionCommand:
			_, ok := result.(ResizeSessionResult)
			return ok
		case InterruptSessionCommand:
			_, ok := result.(InterruptSessionResult)
			return ok
		case StopSessionCommand:
			_, ok := result.(StopSessionResult)
			return ok
		case RestartSessionCommand:
			_, ok := result.(RestartSessionResult)
			return ok
		}
	}
	switch r.query.(type) {
	case NavigatorQuery:
		_, ok := result.(NavigatorResult)
		return ok
	case BranchContextQuery:
		_, ok := result.(BranchContextResult)
		return ok
	case CommitsQuery:
		_, ok := result.(CommitsResult)
		return ok
	case GraphQuery:
		_, ok := result.(GraphResult)
		return ok
	case DiffQuery:
		_, ok := result.(DiffResult)
		return ok
	case PullRequestDiffQuery:
		_, ok := result.(PullRequestDiffResult)
		return ok
	case WorktreeStatusQuery:
		_, ok := result.(WorktreeStatusResult)
		return ok
	case StashesQuery:
		_, ok := result.(StashesResult)
		return ok
	case StashPatchQuery:
		_, ok := result.(StashPatchResult)
		return ok
	case LaunchPointsQuery:
		_, ok := result.(LaunchPointsResult)
		return ok
	case SessionsQuery:
		_, ok := result.(SessionsResult)
		return ok
	case SessionOutputQuery:
		_, ok := result.(SessionOutputProjection)
		return ok
	case PreferencesQuery:
		_, ok := result.(PreferencesResult)
		return ok
	}
	return false
}
func ValidateTerminalFor(request Request, terminal OperationTerminal) error {
	if !request.Valid() || !terminal.Valid() {
		return invalid("terminal/request")
	}
	if request.Correlation().data != terminal.data.Correlation.data {
		return invalid("terminal correlation")
	}
	if r, p := terminal.data.Result.Value(); p {
		if !request.AcceptsResult(r) {
			return invalid("terminal result family")
		}
		if err := requestResultSubject(request, r); err != nil {
			return err
		}
	}
	return nil
}
