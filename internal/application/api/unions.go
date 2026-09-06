package api

// PageContinuation is closed at every containing constructor using an exact type switch.
type PageContinuation interface {
	pageContinuation()
	Valid() bool
}

func (InitialPage) pageContinuation() {}
func (CursorPage) pageContinuation()  {}
func (OffsetPage) pageContinuation()  {}
func validPageContinuation(v PageContinuation) bool {
	switch x := v.(type) {
	case InitialPage:
		return x.Valid()
	case CursorPage:
		return x.Valid()
	case OffsetPage:
		return x.Valid()
	default:
		return false
	}
}

// ShellPolicy is closed at every containing constructor using an exact type switch.
type ShellPolicy interface {
	shellPolicy()
	Valid() bool
}

func (AutoShell) shellPolicy()       {}
func (ConfiguredShell) shellPolicy() {}
func validShellPolicy(v ShellPolicy) bool {
	switch x := v.(type) {
	case AutoShell:
		return x.Valid()
	case ConfiguredShell:
		return x.Valid()
	default:
		return false
	}
}

// ExecutionIntent is closed at every containing constructor using an exact type switch.
type ExecutionIntent interface {
	executionIntent()
	Valid() bool
}

func (ArgvExecution) executionIntent()    {}
func (InteractiveShell) executionIntent() {}
func validExecutionIntent(v ExecutionIntent) bool {
	switch x := v.(type) {
	case ArgvExecution:
		return x.Valid()
	case InteractiveShell:
		return x.Valid()
	default:
		return false
	}
}

// DiagnosticDetail is closed at every containing constructor using an exact type switch.
type DiagnosticDetail interface {
	diagnosticDetail()
	Valid() bool
}

func (NativeDiagnosticDetail) diagnosticDetail()   {}
func (RemoteDiagnosticDetail) diagnosticDetail()   {}
func (IdentityDiagnosticDetail) diagnosticDetail() {}
func (StorageDiagnosticDetail) diagnosticDetail()  {}
func validDiagnosticDetail(v DiagnosticDetail) bool {
	switch x := v.(type) {
	case NativeDiagnosticDetail:
		return x.Valid()
	case RemoteDiagnosticDetail:
		return x.Valid()
	case IdentityDiagnosticDetail:
		return x.Valid()
	case StorageDiagnosticDetail:
		return x.Valid()
	default:
		return false
	}
}

// RecoveryVersion is closed at every containing constructor using an exact type switch.
type RecoveryVersion interface {
	recoveryVersion()
	Valid() bool
}

func (SourceRecoveryVersion) recoveryVersion()  {}
func (StorageRecoveryVersion) recoveryVersion() {}
func validRecoveryVersion(v RecoveryVersion) bool {
	switch x := v.(type) {
	case SourceRecoveryVersion:
		return x.Valid()
	case StorageRecoveryVersion:
		return x.Valid()
	default:
		return false
	}
}

// WorktreeAvailability is closed at every containing constructor using an exact type switch.
type WorktreeAvailability interface {
	worktreeAvailability()
	Valid() bool
}

func (AvailableWorktree) worktreeAvailability()  {}
func (LockedWorktree) worktreeAvailability()     {}
func (PrunableWorktree) worktreeAvailability()   {}
func (MissingWorktree) worktreeAvailability()    {}
func (UnresolvedWorktree) worktreeAvailability() {}
func validWorktreeAvailability(v WorktreeAvailability) bool {
	switch x := v.(type) {
	case AvailableWorktree:
		return x.Valid()
	case LockedWorktree:
		return x.Valid()
	case PrunableWorktree:
		return x.Valid()
	case MissingWorktree:
		return x.Valid()
	case UnresolvedWorktree:
		return x.Valid()
	default:
		return false
	}
}

// FileState is closed at every containing constructor using an exact type switch.
type FileState interface {
	fileState()
	Valid() bool
}

func (AbsentFile) fileState()  {}
func (PresentFile) fileState() {}
func validFileState(v FileState) bool {
	switch x := v.(type) {
	case AbsentFile:
		return x.Valid()
	case PresentFile:
		return x.Valid()
	default:
		return false
	}
}

// UpstreamFact is closed at every containing constructor using an exact type switch.
type UpstreamFact interface {
	upstreamFact()
	Valid() bool
}

func (NoUpstream) upstreamFact()            {}
func (UpstreamNotApplicable) upstreamFact() {}
func (GoneUpstream) upstreamFact()          {}
func (UnresolvedUpstream) upstreamFact()    {}
func (ResolvedUpstream) upstreamFact()      {}
func validUpstreamFact(v UpstreamFact) bool {
	switch x := v.(type) {
	case NoUpstream:
		return x.Valid()
	case UpstreamNotApplicable:
		return x.Valid()
	case GoneUpstream:
		return x.Valid()
	case UnresolvedUpstream:
		return x.Valid()
	case ResolvedUpstream:
		return x.Valid()
	default:
		return false
	}
}

// GitRefLocator is closed at every containing constructor using an exact type switch.
type GitRefLocator interface {
	gitRefLocator()
	Valid() bool
}

func (LocalBranchRef) gitRefLocator()  {}
func (LocalTagRef) gitRefLocator()     {}
func (CachedRemoteRef) gitRefLocator() {}
func (RemoteRef) gitRefLocator()       {}
func validGitRefLocator(v GitRefLocator) bool {
	switch x := v.(type) {
	case LocalBranchRef:
		return x.Valid()
	case LocalTagRef:
		return x.Valid()
	case CachedRemoteRef:
		return x.Valid()
	case RemoteRef:
		return x.Valid()
	default:
		return false
	}
}

// GraphFilter is closed at every containing constructor using an exact type switch.
type GraphFilter interface {
	graphFilter()
	Valid() bool
}

func (AllGraph) graphFilter()           {}
func (ReachableFromRoots) graphFilter() {}
func validGraphFilter(v GraphFilter) bool {
	switch x := v.(type) {
	case AllGraph:
		return x.Valid()
	case ReachableFromRoots:
		return x.Valid()
	default:
		return false
	}
}

// CommitParentSelection is closed at every containing constructor using an exact type switch.
type CommitParentSelection interface {
	commitParentSelection()
	Valid() bool
}

func (RootParent) commitParentSelection()     {}
func (SelectedParent) commitParentSelection() {}
func validCommitParentSelection(v CommitParentSelection) bool {
	switch x := v.(type) {
	case RootParent:
		return x.Valid()
	case SelectedParent:
		return x.Valid()
	default:
		return false
	}
}

// GitComparison is closed at every containing constructor using an exact type switch.
type GitComparison interface {
	gitComparison()
	Valid() bool
}

func (CommitParentComparison) gitComparison()    {}
func (CommitPairComparison) gitComparison()      {}
func (IndexToWorktreeComparison) gitComparison() {}
func (HeadToIndexComparison) gitComparison()     {}
func validGitComparison(v GitComparison) bool {
	switch x := v.(type) {
	case CommitParentComparison:
		return x.Valid()
	case CommitPairComparison:
		return x.Valid()
	case IndexToWorktreeComparison:
		return x.Valid()
	case HeadToIndexComparison:
		return x.Valid()
	default:
		return false
	}
}

// StashPatchView is closed at every containing constructor using an exact type switch.
type StashPatchView interface {
	stashPatchView()
	Valid() bool
}

func (StashBaseToWorktree) stashPatchView()  {}
func (StashBaseToIndex) stashPatchView()     {}
func (StashIndexToWorktree) stashPatchView() {}
func (StashUntracked) stashPatchView()       {}
func (StashParent) stashPatchView()          {}
func validStashPatchView(v StashPatchView) bool {
	switch x := v.(type) {
	case StashBaseToWorktree:
		return x.Valid()
	case StashBaseToIndex:
		return x.Valid()
	case StashIndexToWorktree:
		return x.Valid()
	case StashUntracked:
		return x.Valid()
	case StashParent:
		return x.Valid()
	default:
		return false
	}
}

// MergeBaseOutcome is closed at every containing constructor using an exact type switch.
type MergeBaseOutcome interface {
	mergeBaseOutcome()
	Valid() bool
}

func (UniqueMergeBase) mergeBaseOutcome()    {}
func (NoCommonAncestor) mergeBaseOutcome()   {}
func (AmbiguousMergeBase) mergeBaseOutcome() {}
func validMergeBaseOutcome(v MergeBaseOutcome) bool {
	switch x := v.(type) {
	case UniqueMergeBase:
		return x.Valid()
	case NoCommonAncestor:
		return x.Valid()
	case AmbiguousMergeBase:
		return x.Valid()
	default:
		return false
	}
}

// CreateMode is closed at every containing constructor using an exact type switch.
type CreateMode interface {
	createMode()
	Valid() bool
}

func (DetachedCreate) createMode()  {}
func (CreateNewBranch) createMode() {}
func validCreateMode(v CreateMode) bool {
	switch x := v.(type) {
	case DetachedCreate:
		return x.Valid()
	case CreateNewBranch:
		return x.Valid()
	default:
		return false
	}
}

// RetargetMode is closed at every containing constructor using an exact type switch.
type RetargetMode interface {
	retargetMode()
	Valid() bool
}

func (DetachRetarget) retargetMode()  {}
func (AttachExisting) retargetMode()  {}
func (CreateNewBranch) retargetMode() {}
func (FastForward) retargetMode()     {}
func validRetargetMode(v RetargetMode) bool {
	switch x := v.(type) {
	case DetachRetarget:
		return x.Valid()
	case AttachExisting:
		return x.Valid()
	case CreateNewBranch:
		return x.Valid()
	case FastForward:
		return x.Valid()
	default:
		return false
	}
}

// PathSelection is closed at every containing constructor using an exact type switch.
type PathSelection interface {
	pathSelection()
	Valid() bool
}

func (ExactPaths) pathSelection()  {}
func (AllObserved) pathSelection() {}
func validPathSelection(v PathSelection) bool {
	switch x := v.(type) {
	case ExactPaths:
		return x.Valid()
	case AllObserved:
		return x.Valid()
	default:
		return false
	}
}

// StashIntent is closed at every containing constructor using an exact type switch.
type StashIntent interface {
	stashIntent()
	Valid() bool
}

func (CreateStashIntent) stashIntent() {}
func (ApplyStashIntent) stashIntent()  {}
func (PopStashIntent) stashIntent()    {}
func (DropStashIntent) stashIntent()   {}
func validStashIntent(v StashIntent) bool {
	switch x := v.(type) {
	case CreateStashIntent:
		return x.Valid()
	case ApplyStashIntent:
		return x.Valid()
	case PopStashIntent:
		return x.Valid()
	case DropStashIntent:
		return x.Valid()
	default:
		return false
	}
}

// GitMutationOutcome is closed at every containing constructor using an exact type switch.
type GitMutationOutcome interface {
	gitMutationOutcome()
	Valid() bool
}

func (WorktreeCreated) gitMutationOutcome()            {}
func (WorktreeRetargeted) gitMutationOutcome()         {}
func (IndexChanged) gitMutationOutcome()               {}
func (CommitCreated) gitMutationOutcome()              {}
func (TrackedRestored) gitMutationOutcome()            {}
func (StashCreated) gitMutationOutcome()               {}
func (StashCreatedCleanupRefused) gitMutationOutcome() {}
func (StashApplied) gitMutationOutcome()               {}
func (AppliedWithConflicts) gitMutationOutcome()       {}
func (StashDropped) gitMutationOutcome()               {}
func (BranchCreated) gitMutationOutcome()              {}
func (Pushed) gitMutationOutcome()                     {}
func (RefusedMutation) gitMutationOutcome()            {}
func (PartialMutation) gitMutationOutcome()            {}
func (MutationIndeterminate) gitMutationOutcome()      {}
func validGitMutationOutcome(v GitMutationOutcome) bool {
	switch x := v.(type) {
	case WorktreeCreated:
		return x.Valid()
	case WorktreeRetargeted:
		return x.Valid()
	case IndexChanged:
		return x.Valid()
	case CommitCreated:
		return x.Valid()
	case TrackedRestored:
		return x.Valid()
	case StashCreated:
		return x.Valid()
	case StashCreatedCleanupRefused:
		return x.Valid()
	case StashApplied:
		return x.Valid()
	case AppliedWithConflicts:
		return x.Valid()
	case StashDropped:
		return x.Valid()
	case BranchCreated:
		return x.Valid()
	case Pushed:
		return x.Valid()
	case RefusedMutation:
		return x.Valid()
	case PartialMutation:
		return x.Valid()
	case MutationIndeterminate:
		return x.Valid()
	default:
		return false
	}
}

// RemoteObservationOrigin is closed at every containing constructor using an exact type switch.
type RemoteObservationOrigin interface {
	remoteObservationOrigin()
	Valid() bool
}

func (LiveRemoteObservation) remoteObservationOrigin()   {}
func (CachedRemoteObservation) remoteObservationOrigin() {}
func validRemoteObservationOrigin(v RemoteObservationOrigin) bool {
	switch x := v.(type) {
	case LiveRemoteObservation:
		return x.Valid()
	case CachedRemoteObservation:
		return x.Valid()
	default:
		return false
	}
}

// PullRequestEndpoint is closed at every containing constructor using an exact type switch.
type PullRequestEndpoint interface {
	pullRequestEndpoint()
	Valid() bool
}

func (AvailableEndpoint) pullRequestEndpoint()   {}
func (UnavailableEndpoint) pullRequestEndpoint() {}
func validPullRequestEndpoint(v PullRequestEndpoint) bool {
	switch x := v.(type) {
	case AvailableEndpoint:
		return x.Valid()
	case UnavailableEndpoint:
		return x.Valid()
	default:
		return false
	}
}

// RemoteBranchFilter is closed at every containing constructor using an exact type switch.
type RemoteBranchFilter interface {
	remoteBranchFilter()
	Valid() bool
}

func (AllRemoteBranches) remoteBranchFilter()  {}
func (RemoteBranchPrefix) remoteBranchFilter() {}
func validRemoteBranchFilter(v RemoteBranchFilter) bool {
	switch x := v.(type) {
	case AllRemoteBranches:
		return x.Valid()
	case RemoteBranchPrefix:
		return x.Valid()
	default:
		return false
	}
}

// PullRequestCreationOutcome is closed at every containing constructor using an exact type switch.
type PullRequestCreationOutcome interface {
	pullRequestCreationOutcome()
	Valid() bool
}

func (NotSubmitted) pullRequestCreationOutcome()          {}
func (RejectedNoCreation) pullRequestCreationOutcome()    {}
func (ExistingCandidate) pullRequestCreationOutcome()     {}
func (CreatedVerified) pullRequestCreationOutcome()       {}
func (CreatedWithDrift) pullRequestCreationOutcome()      {}
func (CreationIndeterminate) pullRequestCreationOutcome() {}
func validPullRequestCreationOutcome(v PullRequestCreationOutcome) bool {
	switch x := v.(type) {
	case NotSubmitted:
		return x.Valid()
	case RejectedNoCreation:
		return x.Valid()
	case ExistingCandidate:
		return x.Valid()
	case CreatedVerified:
		return x.Valid()
	case CreatedWithDrift:
		return x.Valid()
	case CreationIndeterminate:
		return x.Valid()
	default:
		return false
	}
}

// LaunchSelection is closed at every containing constructor using an exact type switch.
type LaunchSelection interface {
	launchSelection()
	Valid() bool
}

func (DiscoveredLaunch) launchSelection()  {}
func (OrderedMakeLaunch) launchSelection() {}
func (SavedLaunch) launchSelection()       {}
func validLaunchSelection(v LaunchSelection) bool {
	switch x := v.(type) {
	case DiscoveredLaunch:
		return x.Valid()
	case OrderedMakeLaunch:
		return x.Valid()
	case SavedLaunch:
		return x.Valid()
	default:
		return false
	}
}

// DeployDestination is closed at every containing constructor using an exact type switch.
type DeployDestination interface {
	deployDestination()
	Valid() bool
}

func (ActiveDestination) deployDestination()     {}
func (WorktreeDestination) deployDestination()   {}
func (ConfiguredDestination) deployDestination() {}
func validDeployDestination(v DeployDestination) bool {
	switch x := v.(type) {
	case ActiveDestination:
		return x.Valid()
	case WorktreeDestination:
		return x.Valid()
	case ConfiguredDestination:
		return x.Valid()
	default:
		return false
	}
}

// LaunchIntent is closed at every containing constructor using an exact type switch.
type LaunchIntent interface {
	launchIntent()
	Valid() bool
}

func (CurrentDefaultLaunch) launchIntent() {}
func (SelectedLaunch) launchIntent()       {}
func validLaunchIntent(v LaunchIntent) bool {
	switch x := v.(type) {
	case CurrentDefaultLaunch:
		return x.Valid()
	case SelectedLaunch:
		return x.Valid()
	default:
		return false
	}
}

// Command is closed at every containing constructor using an exact type switch.
type Command interface {
	command()
	Valid() bool
}

func (ActivateWorktreeCommand) command()  {}
func (SaveNavigationCommand) command()    {}
func (CreateWorktreeCommand) command()    {}
func (RetargetWorktreeCommand) command()  {}
func (DeployCommand) command()            {}
func (CreateBranchCommand) command()      {}
func (FetchCommand) command()             {}
func (PullCommand) command()              {}
func (PushCommand) command()              {}
func (StagePathsCommand) command()        {}
func (UnstagePathsCommand) command()      {}
func (StageAllCommand) command()          {}
func (CommitCommand) command()            {}
func (StageAllAndCommitCommand) command() {}
func (RestoreTrackedCommand) command()    {}
func (StashCreateCommand) command()       {}
func (StashApplyCommand) command()        {}
func (StashPopCommand) command()          {}
func (StashDropCommand) command()         {}
func (CreatePullRequestCommand) command() {}
func (SaveLaunchCommand) command()        {}
func (StartLaunchCommand) command()       {}
func (OpenTerminalCommand) command()      {}
func (WriteInputCommand) command()        {}
func (ResizeSessionCommand) command()     {}
func (InterruptSessionCommand) command()  {}
func (StopSessionCommand) command()       {}
func (RestartSessionCommand) command()    {}
func validCommand(v Command) bool {
	switch x := v.(type) {
	case ActivateWorktreeCommand:
		return x.Valid()
	case SaveNavigationCommand:
		return x.Valid()
	case CreateWorktreeCommand:
		return x.Valid()
	case RetargetWorktreeCommand:
		return x.Valid()
	case DeployCommand:
		return x.Valid()
	case CreateBranchCommand:
		return x.Valid()
	case FetchCommand:
		return x.Valid()
	case PullCommand:
		return x.Valid()
	case PushCommand:
		return x.Valid()
	case StagePathsCommand:
		return x.Valid()
	case UnstagePathsCommand:
		return x.Valid()
	case StageAllCommand:
		return x.Valid()
	case CommitCommand:
		return x.Valid()
	case StageAllAndCommitCommand:
		return x.Valid()
	case RestoreTrackedCommand:
		return x.Valid()
	case StashCreateCommand:
		return x.Valid()
	case StashApplyCommand:
		return x.Valid()
	case StashPopCommand:
		return x.Valid()
	case StashDropCommand:
		return x.Valid()
	case CreatePullRequestCommand:
		return x.Valid()
	case SaveLaunchCommand:
		return x.Valid()
	case StartLaunchCommand:
		return x.Valid()
	case OpenTerminalCommand:
		return x.Valid()
	case WriteInputCommand:
		return x.Valid()
	case ResizeSessionCommand:
		return x.Valid()
	case InterruptSessionCommand:
		return x.Valid()
	case StopSessionCommand:
		return x.Valid()
	case RestartSessionCommand:
		return x.Valid()
	default:
		return false
	}
}

// StashPopOutcome is closed at every containing constructor using an exact type switch.
type StashPopOutcome interface {
	stashPopOutcome()
	Valid() bool
}

func (StashPopCompleted) stashPopOutcome()    {}
func (AppliedStashRetained) stashPopOutcome() {}
func (StashPopNotApplied) stashPopOutcome()   {}
func validStashPopOutcome(v StashPopOutcome) bool {
	switch x := v.(type) {
	case StashPopCompleted:
		return x.Valid()
	case AppliedStashRetained:
		return x.Valid()
	case StashPopNotApplied:
		return x.Valid()
	default:
		return false
	}
}

// PRBaseSelection is closed at every containing constructor using an exact type switch.
type PRBaseSelection interface {
	pRBaseSelection()
	Valid() bool
}

func (ExactPRBase) pRBaseSelection()          {}
func (ObserveCurrentPRBase) pRBaseSelection() {}
func validPRBaseSelection(v PRBaseSelection) bool {
	switch x := v.(type) {
	case ExactPRBase:
		return x.Valid()
	case ObserveCurrentPRBase:
		return x.Valid()
	default:
		return false
	}
}

// Query is closed at every containing constructor using an exact type switch.
type Query interface {
	query()
	Valid() bool
}

func (NavigatorQuery) query()       {}
func (BranchContextQuery) query()   {}
func (CommitsQuery) query()         {}
func (GraphQuery) query()           {}
func (DiffQuery) query()            {}
func (PullRequestDiffQuery) query() {}
func (WorktreeStatusQuery) query()  {}
func (StashesQuery) query()         {}
func (StashPatchQuery) query()      {}
func (LaunchPointsQuery) query()    {}
func (SessionsQuery) query()        {}
func (SessionOutputQuery) query()   {}
func (PreferencesQuery) query()     {}
func validQuery(v Query) bool {
	switch x := v.(type) {
	case NavigatorQuery:
		return x.Valid()
	case BranchContextQuery:
		return x.Valid()
	case CommitsQuery:
		return x.Valid()
	case GraphQuery:
		return x.Valid()
	case DiffQuery:
		return x.Valid()
	case PullRequestDiffQuery:
		return x.Valid()
	case WorktreeStatusQuery:
		return x.Valid()
	case StashesQuery:
		return x.Valid()
	case StashPatchQuery:
		return x.Valid()
	case LaunchPointsQuery:
		return x.Valid()
	case SessionsQuery:
		return x.Valid()
	case SessionOutputQuery:
		return x.Valid()
	case PreferencesQuery:
		return x.Valid()
	default:
		return false
	}
}

// Result is closed at every containing constructor using an exact type switch.
type Result interface {
	result()
	Valid() bool
}

func (ActivateWorktreeResult) result()  {}
func (SaveNavigationResult) result()    {}
func (CreateWorktreeResult) result()    {}
func (RetargetWorktreeResult) result()  {}
func (DeployResult) result()            {}
func (CreateBranchResult) result()      {}
func (FetchResult) result()             {}
func (PullResult) result()              {}
func (PushResult) result()              {}
func (StagePathsResult) result()        {}
func (UnstagePathsResult) result()      {}
func (StageAllResult) result()          {}
func (CommitResult) result()            {}
func (StageAllAndCommitResult) result() {}
func (RestoreTrackedResult) result()    {}
func (StashCreateResult) result()       {}
func (StashApplyResult) result()        {}
func (StashPopResult) result()          {}
func (StashDropResult) result()         {}
func (CreatePullRequestResult) result() {}
func (SaveLaunchResult) result()        {}
func (StartLaunchResult) result()       {}
func (OpenTerminalResult) result()      {}
func (WriteInputResult) result()        {}
func (ResizeSessionResult) result()     {}
func (InterruptSessionResult) result()  {}
func (StopSessionResult) result()       {}
func (RestartSessionResult) result()    {}
func (NavigatorResult) result()         {}
func (BranchContextResult) result()     {}
func (CommitsResult) result()           {}
func (GraphResult) result()             {}
func (DiffResult) result()              {}
func (PullRequestDiffResult) result()   {}
func (WorktreeStatusResult) result()    {}
func (StashesResult) result()           {}
func (StashPatchResult) result()        {}
func (LaunchPointsResult) result()      {}
func (SessionsResult) result()          {}
func (SessionOutputProjection) result() {}
func (PreferencesResult) result()       {}
func validResult(v Result) bool {
	switch x := v.(type) {
	case ActivateWorktreeResult:
		return x.Valid()
	case SaveNavigationResult:
		return x.Valid()
	case CreateWorktreeResult:
		return x.Valid()
	case RetargetWorktreeResult:
		return x.Valid()
	case DeployResult:
		return x.Valid()
	case CreateBranchResult:
		return x.Valid()
	case FetchResult:
		return x.Valid()
	case PullResult:
		return x.Valid()
	case PushResult:
		return x.Valid()
	case StagePathsResult:
		return x.Valid()
	case UnstagePathsResult:
		return x.Valid()
	case StageAllResult:
		return x.Valid()
	case CommitResult:
		return x.Valid()
	case StageAllAndCommitResult:
		return x.Valid()
	case RestoreTrackedResult:
		return x.Valid()
	case StashCreateResult:
		return x.Valid()
	case StashApplyResult:
		return x.Valid()
	case StashPopResult:
		return x.Valid()
	case StashDropResult:
		return x.Valid()
	case CreatePullRequestResult:
		return x.Valid()
	case SaveLaunchResult:
		return x.Valid()
	case StartLaunchResult:
		return x.Valid()
	case OpenTerminalResult:
		return x.Valid()
	case WriteInputResult:
		return x.Valid()
	case ResizeSessionResult:
		return x.Valid()
	case InterruptSessionResult:
		return x.Valid()
	case StopSessionResult:
		return x.Valid()
	case RestartSessionResult:
		return x.Valid()
	case NavigatorResult:
		return x.Valid()
	case BranchContextResult:
		return x.Valid()
	case CommitsResult:
		return x.Valid()
	case GraphResult:
		return x.Valid()
	case DiffResult:
		return x.Valid()
	case PullRequestDiffResult:
		return x.Valid()
	case WorktreeStatusResult:
		return x.Valid()
	case StashesResult:
		return x.Valid()
	case StashPatchResult:
		return x.Valid()
	case LaunchPointsResult:
		return x.Valid()
	case SessionsResult:
		return x.Valid()
	case SessionOutputProjection:
		return x.Valid()
	case PreferencesResult:
		return x.Valid()
	default:
		return false
	}
}

// EventPayload is closed at every containing constructor using an exact type switch.
type EventPayload interface {
	eventPayload()
	Valid() bool
}

func (Accepted) eventPayload()              {}
func (Progress) eventPayload()              {}
func (ConfirmationRequested) eventPayload() {}
func (OperationTerminal) eventPayload()     {}
func (SessionChanged) eventPayload()        {}
func (ProjectionInvalidated) eventPayload() {}
func validEventPayload(v EventPayload) bool {
	switch x := v.(type) {
	case Accepted:
		return x.Valid()
	case Progress:
		return x.Valid()
	case ConfirmationRequested:
		return x.Valid()
	case OperationTerminal:
		return x.Valid()
	case SessionChanged:
		return x.Valid()
	case ProjectionInvalidated:
		return x.Valid()
	default:
		return false
	}
}
