package api

import "testing"

type foreignPageContinuation struct{ PageContinuation }

func (foreignPageContinuation) Valid() bool { return true }
func TestPageContinuationRejectsForeignAndNil(t *testing.T) {
	if validPageContinuation(nil) || validPageContinuation(foreignPageContinuation{}) {
		t.Fatal("foreign or nil PageContinuation")
	}
	var nilInitialPage *InitialPage
	if validPageContinuation(nilInitialPage) {
		t.Fatal("typed nil InitialPage")
	}
	var nilCursorPage *CursorPage
	if validPageContinuation(nilCursorPage) {
		t.Fatal("typed nil CursorPage")
	}
	var nilOffsetPage *OffsetPage
	if validPageContinuation(nilOffsetPage) {
		t.Fatal("typed nil OffsetPage")
	}
}

type foreignShellPolicy struct{ ShellPolicy }

func (foreignShellPolicy) Valid() bool { return true }
func TestShellPolicyRejectsForeignAndNil(t *testing.T) {
	if validShellPolicy(nil) || validShellPolicy(foreignShellPolicy{}) {
		t.Fatal("foreign or nil ShellPolicy")
	}
	var nilAutoShell *AutoShell
	if validShellPolicy(nilAutoShell) {
		t.Fatal("typed nil AutoShell")
	}
	var nilConfiguredShell *ConfiguredShell
	if validShellPolicy(nilConfiguredShell) {
		t.Fatal("typed nil ConfiguredShell")
	}
}

type foreignExecutionIntent struct{ ExecutionIntent }

func (foreignExecutionIntent) Valid() bool { return true }
func TestExecutionIntentRejectsForeignAndNil(t *testing.T) {
	if validExecutionIntent(nil) || validExecutionIntent(foreignExecutionIntent{}) {
		t.Fatal("foreign or nil ExecutionIntent")
	}
	var nilArgvExecution *ArgvExecution
	if validExecutionIntent(nilArgvExecution) {
		t.Fatal("typed nil ArgvExecution")
	}
	var nilInteractiveShell *InteractiveShell
	if validExecutionIntent(nilInteractiveShell) {
		t.Fatal("typed nil InteractiveShell")
	}
}

type foreignDiagnosticDetail struct{ DiagnosticDetail }

func (foreignDiagnosticDetail) Valid() bool { return true }
func TestDiagnosticDetailRejectsForeignAndNil(t *testing.T) {
	if validDiagnosticDetail(nil) || validDiagnosticDetail(foreignDiagnosticDetail{}) {
		t.Fatal("foreign or nil DiagnosticDetail")
	}
	var nilNativeDiagnosticDetail *NativeDiagnosticDetail
	if validDiagnosticDetail(nilNativeDiagnosticDetail) {
		t.Fatal("typed nil NativeDiagnosticDetail")
	}
	var nilRemoteDiagnosticDetail *RemoteDiagnosticDetail
	if validDiagnosticDetail(nilRemoteDiagnosticDetail) {
		t.Fatal("typed nil RemoteDiagnosticDetail")
	}
	var nilIdentityDiagnosticDetail *IdentityDiagnosticDetail
	if validDiagnosticDetail(nilIdentityDiagnosticDetail) {
		t.Fatal("typed nil IdentityDiagnosticDetail")
	}
	var nilStorageDiagnosticDetail *StorageDiagnosticDetail
	if validDiagnosticDetail(nilStorageDiagnosticDetail) {
		t.Fatal("typed nil StorageDiagnosticDetail")
	}
}

type foreignRecoveryVersion struct{ RecoveryVersion }

func (foreignRecoveryVersion) Valid() bool { return true }
func TestRecoveryVersionRejectsForeignAndNil(t *testing.T) {
	if validRecoveryVersion(nil) || validRecoveryVersion(foreignRecoveryVersion{}) {
		t.Fatal("foreign or nil RecoveryVersion")
	}
	var nilSourceRecoveryVersion *SourceRecoveryVersion
	if validRecoveryVersion(nilSourceRecoveryVersion) {
		t.Fatal("typed nil SourceRecoveryVersion")
	}
	var nilStorageRecoveryVersion *StorageRecoveryVersion
	if validRecoveryVersion(nilStorageRecoveryVersion) {
		t.Fatal("typed nil StorageRecoveryVersion")
	}
}

type foreignWorktreeAvailability struct{ WorktreeAvailability }

func (foreignWorktreeAvailability) Valid() bool { return true }
func TestWorktreeAvailabilityRejectsForeignAndNil(t *testing.T) {
	if validWorktreeAvailability(nil) || validWorktreeAvailability(foreignWorktreeAvailability{}) {
		t.Fatal("foreign or nil WorktreeAvailability")
	}
	var nilAvailableWorktree *AvailableWorktree
	if validWorktreeAvailability(nilAvailableWorktree) {
		t.Fatal("typed nil AvailableWorktree")
	}
	var nilLockedWorktree *LockedWorktree
	if validWorktreeAvailability(nilLockedWorktree) {
		t.Fatal("typed nil LockedWorktree")
	}
	var nilPrunableWorktree *PrunableWorktree
	if validWorktreeAvailability(nilPrunableWorktree) {
		t.Fatal("typed nil PrunableWorktree")
	}
	var nilMissingWorktree *MissingWorktree
	if validWorktreeAvailability(nilMissingWorktree) {
		t.Fatal("typed nil MissingWorktree")
	}
	var nilUnresolvedWorktree *UnresolvedWorktree
	if validWorktreeAvailability(nilUnresolvedWorktree) {
		t.Fatal("typed nil UnresolvedWorktree")
	}
}

type foreignFileState struct{ FileState }

func (foreignFileState) Valid() bool { return true }
func TestFileStateRejectsForeignAndNil(t *testing.T) {
	if validFileState(nil) || validFileState(foreignFileState{}) {
		t.Fatal("foreign or nil FileState")
	}
	var nilAbsentFile *AbsentFile
	if validFileState(nilAbsentFile) {
		t.Fatal("typed nil AbsentFile")
	}
	var nilPresentFile *PresentFile
	if validFileState(nilPresentFile) {
		t.Fatal("typed nil PresentFile")
	}
}

type foreignUpstreamFact struct{ UpstreamFact }

func (foreignUpstreamFact) Valid() bool { return true }
func TestUpstreamFactRejectsForeignAndNil(t *testing.T) {
	if validUpstreamFact(nil) || validUpstreamFact(foreignUpstreamFact{}) {
		t.Fatal("foreign or nil UpstreamFact")
	}
	var nilNoUpstream *NoUpstream
	if validUpstreamFact(nilNoUpstream) {
		t.Fatal("typed nil NoUpstream")
	}
	var nilUpstreamNotApplicable *UpstreamNotApplicable
	if validUpstreamFact(nilUpstreamNotApplicable) {
		t.Fatal("typed nil UpstreamNotApplicable")
	}
	var nilGoneUpstream *GoneUpstream
	if validUpstreamFact(nilGoneUpstream) {
		t.Fatal("typed nil GoneUpstream")
	}
	var nilUnresolvedUpstream *UnresolvedUpstream
	if validUpstreamFact(nilUnresolvedUpstream) {
		t.Fatal("typed nil UnresolvedUpstream")
	}
	var nilResolvedUpstream *ResolvedUpstream
	if validUpstreamFact(nilResolvedUpstream) {
		t.Fatal("typed nil ResolvedUpstream")
	}
}

type foreignGitRefLocator struct{ GitRefLocator }

func (foreignGitRefLocator) Valid() bool { return true }
func TestGitRefLocatorRejectsForeignAndNil(t *testing.T) {
	if validGitRefLocator(nil) || validGitRefLocator(foreignGitRefLocator{}) {
		t.Fatal("foreign or nil GitRefLocator")
	}
	var nilLocalBranchRef *LocalBranchRef
	if validGitRefLocator(nilLocalBranchRef) {
		t.Fatal("typed nil LocalBranchRef")
	}
	var nilLocalTagRef *LocalTagRef
	if validGitRefLocator(nilLocalTagRef) {
		t.Fatal("typed nil LocalTagRef")
	}
	var nilCachedRemoteRef *CachedRemoteRef
	if validGitRefLocator(nilCachedRemoteRef) {
		t.Fatal("typed nil CachedRemoteRef")
	}
	var nilRemoteRef *RemoteRef
	if validGitRefLocator(nilRemoteRef) {
		t.Fatal("typed nil RemoteRef")
	}
}

type foreignGraphFilter struct{ GraphFilter }

func (foreignGraphFilter) Valid() bool { return true }
func TestGraphFilterRejectsForeignAndNil(t *testing.T) {
	if validGraphFilter(nil) || validGraphFilter(foreignGraphFilter{}) {
		t.Fatal("foreign or nil GraphFilter")
	}
	var nilAllGraph *AllGraph
	if validGraphFilter(nilAllGraph) {
		t.Fatal("typed nil AllGraph")
	}
	var nilReachableFromRoots *ReachableFromRoots
	if validGraphFilter(nilReachableFromRoots) {
		t.Fatal("typed nil ReachableFromRoots")
	}
}

type foreignCommitParentSelection struct{ CommitParentSelection }

func (foreignCommitParentSelection) Valid() bool { return true }
func TestCommitParentSelectionRejectsForeignAndNil(t *testing.T) {
	if validCommitParentSelection(nil) || validCommitParentSelection(foreignCommitParentSelection{}) {
		t.Fatal("foreign or nil CommitParentSelection")
	}
	var nilRootParent *RootParent
	if validCommitParentSelection(nilRootParent) {
		t.Fatal("typed nil RootParent")
	}
	var nilSelectedParent *SelectedParent
	if validCommitParentSelection(nilSelectedParent) {
		t.Fatal("typed nil SelectedParent")
	}
}

type foreignGitComparison struct{ GitComparison }

func (foreignGitComparison) Valid() bool { return true }
func TestGitComparisonRejectsForeignAndNil(t *testing.T) {
	if validGitComparison(nil) || validGitComparison(foreignGitComparison{}) {
		t.Fatal("foreign or nil GitComparison")
	}
	var nilCommitParentComparison *CommitParentComparison
	if validGitComparison(nilCommitParentComparison) {
		t.Fatal("typed nil CommitParentComparison")
	}
	var nilCommitPairComparison *CommitPairComparison
	if validGitComparison(nilCommitPairComparison) {
		t.Fatal("typed nil CommitPairComparison")
	}
	var nilIndexToWorktreeComparison *IndexToWorktreeComparison
	if validGitComparison(nilIndexToWorktreeComparison) {
		t.Fatal("typed nil IndexToWorktreeComparison")
	}
	var nilHeadToIndexComparison *HeadToIndexComparison
	if validGitComparison(nilHeadToIndexComparison) {
		t.Fatal("typed nil HeadToIndexComparison")
	}
}

type foreignStashPatchView struct{ StashPatchView }

func (foreignStashPatchView) Valid() bool { return true }
func TestStashPatchViewRejectsForeignAndNil(t *testing.T) {
	if validStashPatchView(nil) || validStashPatchView(foreignStashPatchView{}) {
		t.Fatal("foreign or nil StashPatchView")
	}
	var nilStashBaseToWorktree *StashBaseToWorktree
	if validStashPatchView(nilStashBaseToWorktree) {
		t.Fatal("typed nil StashBaseToWorktree")
	}
	var nilStashBaseToIndex *StashBaseToIndex
	if validStashPatchView(nilStashBaseToIndex) {
		t.Fatal("typed nil StashBaseToIndex")
	}
	var nilStashIndexToWorktree *StashIndexToWorktree
	if validStashPatchView(nilStashIndexToWorktree) {
		t.Fatal("typed nil StashIndexToWorktree")
	}
	var nilStashUntracked *StashUntracked
	if validStashPatchView(nilStashUntracked) {
		t.Fatal("typed nil StashUntracked")
	}
	var nilStashParent *StashParent
	if validStashPatchView(nilStashParent) {
		t.Fatal("typed nil StashParent")
	}
}

type foreignMergeBaseOutcome struct{ MergeBaseOutcome }

func (foreignMergeBaseOutcome) Valid() bool { return true }
func TestMergeBaseOutcomeRejectsForeignAndNil(t *testing.T) {
	if validMergeBaseOutcome(nil) || validMergeBaseOutcome(foreignMergeBaseOutcome{}) {
		t.Fatal("foreign or nil MergeBaseOutcome")
	}
	var nilUniqueMergeBase *UniqueMergeBase
	if validMergeBaseOutcome(nilUniqueMergeBase) {
		t.Fatal("typed nil UniqueMergeBase")
	}
	var nilNoCommonAncestor *NoCommonAncestor
	if validMergeBaseOutcome(nilNoCommonAncestor) {
		t.Fatal("typed nil NoCommonAncestor")
	}
	var nilAmbiguousMergeBase *AmbiguousMergeBase
	if validMergeBaseOutcome(nilAmbiguousMergeBase) {
		t.Fatal("typed nil AmbiguousMergeBase")
	}
}

type foreignCreateMode struct{ CreateMode }

func (foreignCreateMode) Valid() bool { return true }
func TestCreateModeRejectsForeignAndNil(t *testing.T) {
	if validCreateMode(nil) || validCreateMode(foreignCreateMode{}) {
		t.Fatal("foreign or nil CreateMode")
	}
	var nilDetachedCreate *DetachedCreate
	if validCreateMode(nilDetachedCreate) {
		t.Fatal("typed nil DetachedCreate")
	}
	var nilCreateNewBranch *CreateNewBranch
	if validCreateMode(nilCreateNewBranch) {
		t.Fatal("typed nil CreateNewBranch")
	}
}

type foreignRetargetMode struct{ RetargetMode }

func (foreignRetargetMode) Valid() bool { return true }
func TestRetargetModeRejectsForeignAndNil(t *testing.T) {
	if validRetargetMode(nil) || validRetargetMode(foreignRetargetMode{}) {
		t.Fatal("foreign or nil RetargetMode")
	}
	var nilDetachRetarget *DetachRetarget
	if validRetargetMode(nilDetachRetarget) {
		t.Fatal("typed nil DetachRetarget")
	}
	var nilAttachExisting *AttachExisting
	if validRetargetMode(nilAttachExisting) {
		t.Fatal("typed nil AttachExisting")
	}
	var nilCreateNewBranch *CreateNewBranch
	if validRetargetMode(nilCreateNewBranch) {
		t.Fatal("typed nil CreateNewBranch")
	}
	var nilFastForward *FastForward
	if validRetargetMode(nilFastForward) {
		t.Fatal("typed nil FastForward")
	}
}

type foreignPathSelection struct{ PathSelection }

func (foreignPathSelection) Valid() bool { return true }
func TestPathSelectionRejectsForeignAndNil(t *testing.T) {
	if validPathSelection(nil) || validPathSelection(foreignPathSelection{}) {
		t.Fatal("foreign or nil PathSelection")
	}
	var nilExactPaths *ExactPaths
	if validPathSelection(nilExactPaths) {
		t.Fatal("typed nil ExactPaths")
	}
	var nilAllObserved *AllObserved
	if validPathSelection(nilAllObserved) {
		t.Fatal("typed nil AllObserved")
	}
}

type foreignStashIntent struct{ StashIntent }

func (foreignStashIntent) Valid() bool { return true }
func TestStashIntentRejectsForeignAndNil(t *testing.T) {
	if validStashIntent(nil) || validStashIntent(foreignStashIntent{}) {
		t.Fatal("foreign or nil StashIntent")
	}
	var nilCreateStashIntent *CreateStashIntent
	if validStashIntent(nilCreateStashIntent) {
		t.Fatal("typed nil CreateStashIntent")
	}
	var nilApplyStashIntent *ApplyStashIntent
	if validStashIntent(nilApplyStashIntent) {
		t.Fatal("typed nil ApplyStashIntent")
	}
	var nilPopStashIntent *PopStashIntent
	if validStashIntent(nilPopStashIntent) {
		t.Fatal("typed nil PopStashIntent")
	}
	var nilDropStashIntent *DropStashIntent
	if validStashIntent(nilDropStashIntent) {
		t.Fatal("typed nil DropStashIntent")
	}
}

type foreignGitMutationOutcome struct{ GitMutationOutcome }

func (foreignGitMutationOutcome) Valid() bool { return true }
func TestGitMutationOutcomeRejectsForeignAndNil(t *testing.T) {
	if validGitMutationOutcome(nil) || validGitMutationOutcome(foreignGitMutationOutcome{}) {
		t.Fatal("foreign or nil GitMutationOutcome")
	}
	var nilWorktreeCreated *WorktreeCreated
	if validGitMutationOutcome(nilWorktreeCreated) {
		t.Fatal("typed nil WorktreeCreated")
	}
	var nilWorktreeRetargeted *WorktreeRetargeted
	if validGitMutationOutcome(nilWorktreeRetargeted) {
		t.Fatal("typed nil WorktreeRetargeted")
	}
	var nilIndexChanged *IndexChanged
	if validGitMutationOutcome(nilIndexChanged) {
		t.Fatal("typed nil IndexChanged")
	}
	var nilCommitCreated *CommitCreated
	if validGitMutationOutcome(nilCommitCreated) {
		t.Fatal("typed nil CommitCreated")
	}
	var nilTrackedRestored *TrackedRestored
	if validGitMutationOutcome(nilTrackedRestored) {
		t.Fatal("typed nil TrackedRestored")
	}
	var nilStashCreated *StashCreated
	if validGitMutationOutcome(nilStashCreated) {
		t.Fatal("typed nil StashCreated")
	}
	var nilStashCreatedCleanupRefused *StashCreatedCleanupRefused
	if validGitMutationOutcome(nilStashCreatedCleanupRefused) {
		t.Fatal("typed nil StashCreatedCleanupRefused")
	}
	var nilStashApplied *StashApplied
	if validGitMutationOutcome(nilStashApplied) {
		t.Fatal("typed nil StashApplied")
	}
	var nilAppliedWithConflicts *AppliedWithConflicts
	if validGitMutationOutcome(nilAppliedWithConflicts) {
		t.Fatal("typed nil AppliedWithConflicts")
	}
	var nilStashDropped *StashDropped
	if validGitMutationOutcome(nilStashDropped) {
		t.Fatal("typed nil StashDropped")
	}
	var nilBranchCreated *BranchCreated
	if validGitMutationOutcome(nilBranchCreated) {
		t.Fatal("typed nil BranchCreated")
	}
	var nilPushed *Pushed
	if validGitMutationOutcome(nilPushed) {
		t.Fatal("typed nil Pushed")
	}
	var nilRefusedMutation *RefusedMutation
	if validGitMutationOutcome(nilRefusedMutation) {
		t.Fatal("typed nil RefusedMutation")
	}
	var nilPartialMutation *PartialMutation
	if validGitMutationOutcome(nilPartialMutation) {
		t.Fatal("typed nil PartialMutation")
	}
	var nilMutationIndeterminate *MutationIndeterminate
	if validGitMutationOutcome(nilMutationIndeterminate) {
		t.Fatal("typed nil MutationIndeterminate")
	}
}

type foreignRemoteObservationOrigin struct{ RemoteObservationOrigin }

func (foreignRemoteObservationOrigin) Valid() bool { return true }
func TestRemoteObservationOriginRejectsForeignAndNil(t *testing.T) {
	if validRemoteObservationOrigin(nil) || validRemoteObservationOrigin(foreignRemoteObservationOrigin{}) {
		t.Fatal("foreign or nil RemoteObservationOrigin")
	}
	var nilLiveRemoteObservation *LiveRemoteObservation
	if validRemoteObservationOrigin(nilLiveRemoteObservation) {
		t.Fatal("typed nil LiveRemoteObservation")
	}
	var nilCachedRemoteObservation *CachedRemoteObservation
	if validRemoteObservationOrigin(nilCachedRemoteObservation) {
		t.Fatal("typed nil CachedRemoteObservation")
	}
}

type foreignPullRequestEndpoint struct{ PullRequestEndpoint }

func (foreignPullRequestEndpoint) Valid() bool { return true }
func TestPullRequestEndpointRejectsForeignAndNil(t *testing.T) {
	if validPullRequestEndpoint(nil) || validPullRequestEndpoint(foreignPullRequestEndpoint{}) {
		t.Fatal("foreign or nil PullRequestEndpoint")
	}
	var nilAvailableEndpoint *AvailableEndpoint
	if validPullRequestEndpoint(nilAvailableEndpoint) {
		t.Fatal("typed nil AvailableEndpoint")
	}
	var nilUnavailableEndpoint *UnavailableEndpoint
	if validPullRequestEndpoint(nilUnavailableEndpoint) {
		t.Fatal("typed nil UnavailableEndpoint")
	}
}

type foreignRemoteBranchFilter struct{ RemoteBranchFilter }

func (foreignRemoteBranchFilter) Valid() bool { return true }
func TestRemoteBranchFilterRejectsForeignAndNil(t *testing.T) {
	if validRemoteBranchFilter(nil) || validRemoteBranchFilter(foreignRemoteBranchFilter{}) {
		t.Fatal("foreign or nil RemoteBranchFilter")
	}
	var nilAllRemoteBranches *AllRemoteBranches
	if validRemoteBranchFilter(nilAllRemoteBranches) {
		t.Fatal("typed nil AllRemoteBranches")
	}
	var nilRemoteBranchPrefix *RemoteBranchPrefix
	if validRemoteBranchFilter(nilRemoteBranchPrefix) {
		t.Fatal("typed nil RemoteBranchPrefix")
	}
}

type foreignPullRequestCreationOutcome struct{ PullRequestCreationOutcome }

func (foreignPullRequestCreationOutcome) Valid() bool { return true }
func TestPullRequestCreationOutcomeRejectsForeignAndNil(t *testing.T) {
	if validPullRequestCreationOutcome(nil) || validPullRequestCreationOutcome(foreignPullRequestCreationOutcome{}) {
		t.Fatal("foreign or nil PullRequestCreationOutcome")
	}
	var nilNotSubmitted *NotSubmitted
	if validPullRequestCreationOutcome(nilNotSubmitted) {
		t.Fatal("typed nil NotSubmitted")
	}
	var nilRejectedNoCreation *RejectedNoCreation
	if validPullRequestCreationOutcome(nilRejectedNoCreation) {
		t.Fatal("typed nil RejectedNoCreation")
	}
	var nilExistingCandidate *ExistingCandidate
	if validPullRequestCreationOutcome(nilExistingCandidate) {
		t.Fatal("typed nil ExistingCandidate")
	}
	var nilCreatedVerified *CreatedVerified
	if validPullRequestCreationOutcome(nilCreatedVerified) {
		t.Fatal("typed nil CreatedVerified")
	}
	var nilCreatedWithDrift *CreatedWithDrift
	if validPullRequestCreationOutcome(nilCreatedWithDrift) {
		t.Fatal("typed nil CreatedWithDrift")
	}
	var nilCreationIndeterminate *CreationIndeterminate
	if validPullRequestCreationOutcome(nilCreationIndeterminate) {
		t.Fatal("typed nil CreationIndeterminate")
	}
}

type foreignLaunchSelection struct{ LaunchSelection }

func (foreignLaunchSelection) Valid() bool { return true }
func TestLaunchSelectionRejectsForeignAndNil(t *testing.T) {
	if validLaunchSelection(nil) || validLaunchSelection(foreignLaunchSelection{}) {
		t.Fatal("foreign or nil LaunchSelection")
	}
	var nilDiscoveredLaunch *DiscoveredLaunch
	if validLaunchSelection(nilDiscoveredLaunch) {
		t.Fatal("typed nil DiscoveredLaunch")
	}
	var nilOrderedMakeLaunch *OrderedMakeLaunch
	if validLaunchSelection(nilOrderedMakeLaunch) {
		t.Fatal("typed nil OrderedMakeLaunch")
	}
	var nilSavedLaunch *SavedLaunch
	if validLaunchSelection(nilSavedLaunch) {
		t.Fatal("typed nil SavedLaunch")
	}
}

type foreignDeployDestination struct{ DeployDestination }

func (foreignDeployDestination) Valid() bool { return true }
func TestDeployDestinationRejectsForeignAndNil(t *testing.T) {
	if validDeployDestination(nil) || validDeployDestination(foreignDeployDestination{}) {
		t.Fatal("foreign or nil DeployDestination")
	}
	var nilActiveDestination *ActiveDestination
	if validDeployDestination(nilActiveDestination) {
		t.Fatal("typed nil ActiveDestination")
	}
	var nilWorktreeDestination *WorktreeDestination
	if validDeployDestination(nilWorktreeDestination) {
		t.Fatal("typed nil WorktreeDestination")
	}
	var nilConfiguredDestination *ConfiguredDestination
	if validDeployDestination(nilConfiguredDestination) {
		t.Fatal("typed nil ConfiguredDestination")
	}
}

type foreignLaunchIntent struct{ LaunchIntent }

func (foreignLaunchIntent) Valid() bool { return true }
func TestLaunchIntentRejectsForeignAndNil(t *testing.T) {
	if validLaunchIntent(nil) || validLaunchIntent(foreignLaunchIntent{}) {
		t.Fatal("foreign or nil LaunchIntent")
	}
	var nilCurrentDefaultLaunch *CurrentDefaultLaunch
	if validLaunchIntent(nilCurrentDefaultLaunch) {
		t.Fatal("typed nil CurrentDefaultLaunch")
	}
	var nilSelectedLaunch *SelectedLaunch
	if validLaunchIntent(nilSelectedLaunch) {
		t.Fatal("typed nil SelectedLaunch")
	}
}

type foreignCommand struct{ Command }

func (foreignCommand) Valid() bool { return true }
func TestCommandRejectsForeignAndNil(t *testing.T) {
	if validCommand(nil) || validCommand(foreignCommand{}) {
		t.Fatal("foreign or nil Command")
	}
	var nilActivateWorktreeCommand *ActivateWorktreeCommand
	if validCommand(nilActivateWorktreeCommand) {
		t.Fatal("typed nil ActivateWorktreeCommand")
	}
	var nilSaveNavigationCommand *SaveNavigationCommand
	if validCommand(nilSaveNavigationCommand) {
		t.Fatal("typed nil SaveNavigationCommand")
	}
	var nilCreateWorktreeCommand *CreateWorktreeCommand
	if validCommand(nilCreateWorktreeCommand) {
		t.Fatal("typed nil CreateWorktreeCommand")
	}
	var nilRetargetWorktreeCommand *RetargetWorktreeCommand
	if validCommand(nilRetargetWorktreeCommand) {
		t.Fatal("typed nil RetargetWorktreeCommand")
	}
	var nilDeployCommand *DeployCommand
	if validCommand(nilDeployCommand) {
		t.Fatal("typed nil DeployCommand")
	}
	var nilCreateBranchCommand *CreateBranchCommand
	if validCommand(nilCreateBranchCommand) {
		t.Fatal("typed nil CreateBranchCommand")
	}
	var nilFetchCommand *FetchCommand
	if validCommand(nilFetchCommand) {
		t.Fatal("typed nil FetchCommand")
	}
	var nilPullCommand *PullCommand
	if validCommand(nilPullCommand) {
		t.Fatal("typed nil PullCommand")
	}
	var nilPushCommand *PushCommand
	if validCommand(nilPushCommand) {
		t.Fatal("typed nil PushCommand")
	}
	var nilStagePathsCommand *StagePathsCommand
	if validCommand(nilStagePathsCommand) {
		t.Fatal("typed nil StagePathsCommand")
	}
	var nilUnstagePathsCommand *UnstagePathsCommand
	if validCommand(nilUnstagePathsCommand) {
		t.Fatal("typed nil UnstagePathsCommand")
	}
	var nilStageAllCommand *StageAllCommand
	if validCommand(nilStageAllCommand) {
		t.Fatal("typed nil StageAllCommand")
	}
	var nilCommitCommand *CommitCommand
	if validCommand(nilCommitCommand) {
		t.Fatal("typed nil CommitCommand")
	}
	var nilStageAllAndCommitCommand *StageAllAndCommitCommand
	if validCommand(nilStageAllAndCommitCommand) {
		t.Fatal("typed nil StageAllAndCommitCommand")
	}
	var nilRestoreTrackedCommand *RestoreTrackedCommand
	if validCommand(nilRestoreTrackedCommand) {
		t.Fatal("typed nil RestoreTrackedCommand")
	}
	var nilStashCreateCommand *StashCreateCommand
	if validCommand(nilStashCreateCommand) {
		t.Fatal("typed nil StashCreateCommand")
	}
	var nilStashApplyCommand *StashApplyCommand
	if validCommand(nilStashApplyCommand) {
		t.Fatal("typed nil StashApplyCommand")
	}
	var nilStashPopCommand *StashPopCommand
	if validCommand(nilStashPopCommand) {
		t.Fatal("typed nil StashPopCommand")
	}
	var nilStashDropCommand *StashDropCommand
	if validCommand(nilStashDropCommand) {
		t.Fatal("typed nil StashDropCommand")
	}
	var nilCreatePullRequestCommand *CreatePullRequestCommand
	if validCommand(nilCreatePullRequestCommand) {
		t.Fatal("typed nil CreatePullRequestCommand")
	}
	var nilSaveLaunchCommand *SaveLaunchCommand
	if validCommand(nilSaveLaunchCommand) {
		t.Fatal("typed nil SaveLaunchCommand")
	}
	var nilStartLaunchCommand *StartLaunchCommand
	if validCommand(nilStartLaunchCommand) {
		t.Fatal("typed nil StartLaunchCommand")
	}
	var nilOpenTerminalCommand *OpenTerminalCommand
	if validCommand(nilOpenTerminalCommand) {
		t.Fatal("typed nil OpenTerminalCommand")
	}
	var nilWriteInputCommand *WriteInputCommand
	if validCommand(nilWriteInputCommand) {
		t.Fatal("typed nil WriteInputCommand")
	}
	var nilResizeSessionCommand *ResizeSessionCommand
	if validCommand(nilResizeSessionCommand) {
		t.Fatal("typed nil ResizeSessionCommand")
	}
	var nilInterruptSessionCommand *InterruptSessionCommand
	if validCommand(nilInterruptSessionCommand) {
		t.Fatal("typed nil InterruptSessionCommand")
	}
	var nilStopSessionCommand *StopSessionCommand
	if validCommand(nilStopSessionCommand) {
		t.Fatal("typed nil StopSessionCommand")
	}
	var nilRestartSessionCommand *RestartSessionCommand
	if validCommand(nilRestartSessionCommand) {
		t.Fatal("typed nil RestartSessionCommand")
	}
}

type foreignStashPopOutcome struct{ StashPopOutcome }

func (foreignStashPopOutcome) Valid() bool { return true }
func TestStashPopOutcomeRejectsForeignAndNil(t *testing.T) {
	if validStashPopOutcome(nil) || validStashPopOutcome(foreignStashPopOutcome{}) {
		t.Fatal("foreign or nil StashPopOutcome")
	}
	var nilStashPopCompleted *StashPopCompleted
	if validStashPopOutcome(nilStashPopCompleted) {
		t.Fatal("typed nil StashPopCompleted")
	}
	var nilAppliedStashRetained *AppliedStashRetained
	if validStashPopOutcome(nilAppliedStashRetained) {
		t.Fatal("typed nil AppliedStashRetained")
	}
	var nilStashPopNotApplied *StashPopNotApplied
	if validStashPopOutcome(nilStashPopNotApplied) {
		t.Fatal("typed nil StashPopNotApplied")
	}
}

type foreignPRBaseSelection struct{ PRBaseSelection }

func (foreignPRBaseSelection) Valid() bool { return true }
func TestPRBaseSelectionRejectsForeignAndNil(t *testing.T) {
	if validPRBaseSelection(nil) || validPRBaseSelection(foreignPRBaseSelection{}) {
		t.Fatal("foreign or nil PRBaseSelection")
	}
	var nilExactPRBase *ExactPRBase
	if validPRBaseSelection(nilExactPRBase) {
		t.Fatal("typed nil ExactPRBase")
	}
	var nilObserveCurrentPRBase *ObserveCurrentPRBase
	if validPRBaseSelection(nilObserveCurrentPRBase) {
		t.Fatal("typed nil ObserveCurrentPRBase")
	}
}

type foreignQuery struct{ Query }

func (foreignQuery) Valid() bool { return true }
func TestQueryRejectsForeignAndNil(t *testing.T) {
	if validQuery(nil) || validQuery(foreignQuery{}) {
		t.Fatal("foreign or nil Query")
	}
	var nilNavigatorQuery *NavigatorQuery
	if validQuery(nilNavigatorQuery) {
		t.Fatal("typed nil NavigatorQuery")
	}
	var nilBranchContextQuery *BranchContextQuery
	if validQuery(nilBranchContextQuery) {
		t.Fatal("typed nil BranchContextQuery")
	}
	var nilCommitsQuery *CommitsQuery
	if validQuery(nilCommitsQuery) {
		t.Fatal("typed nil CommitsQuery")
	}
	var nilGraphQuery *GraphQuery
	if validQuery(nilGraphQuery) {
		t.Fatal("typed nil GraphQuery")
	}
	var nilDiffQuery *DiffQuery
	if validQuery(nilDiffQuery) {
		t.Fatal("typed nil DiffQuery")
	}
	var nilPullRequestDiffQuery *PullRequestDiffQuery
	if validQuery(nilPullRequestDiffQuery) {
		t.Fatal("typed nil PullRequestDiffQuery")
	}
	var nilWorktreeStatusQuery *WorktreeStatusQuery
	if validQuery(nilWorktreeStatusQuery) {
		t.Fatal("typed nil WorktreeStatusQuery")
	}
	var nilStashesQuery *StashesQuery
	if validQuery(nilStashesQuery) {
		t.Fatal("typed nil StashesQuery")
	}
	var nilStashPatchQuery *StashPatchQuery
	if validQuery(nilStashPatchQuery) {
		t.Fatal("typed nil StashPatchQuery")
	}
	var nilLaunchPointsQuery *LaunchPointsQuery
	if validQuery(nilLaunchPointsQuery) {
		t.Fatal("typed nil LaunchPointsQuery")
	}
	var nilSessionsQuery *SessionsQuery
	if validQuery(nilSessionsQuery) {
		t.Fatal("typed nil SessionsQuery")
	}
	var nilSessionOutputQuery *SessionOutputQuery
	if validQuery(nilSessionOutputQuery) {
		t.Fatal("typed nil SessionOutputQuery")
	}
	var nilPreferencesQuery *PreferencesQuery
	if validQuery(nilPreferencesQuery) {
		t.Fatal("typed nil PreferencesQuery")
	}
}

type foreignResult struct{ Result }

func (foreignResult) Valid() bool { return true }
func TestResultRejectsForeignAndNil(t *testing.T) {
	if validResult(nil) || validResult(foreignResult{}) {
		t.Fatal("foreign or nil Result")
	}
	var nilActivateWorktreeResult *ActivateWorktreeResult
	if validResult(nilActivateWorktreeResult) {
		t.Fatal("typed nil ActivateWorktreeResult")
	}
	var nilSaveNavigationResult *SaveNavigationResult
	if validResult(nilSaveNavigationResult) {
		t.Fatal("typed nil SaveNavigationResult")
	}
	var nilCreateWorktreeResult *CreateWorktreeResult
	if validResult(nilCreateWorktreeResult) {
		t.Fatal("typed nil CreateWorktreeResult")
	}
	var nilRetargetWorktreeResult *RetargetWorktreeResult
	if validResult(nilRetargetWorktreeResult) {
		t.Fatal("typed nil RetargetWorktreeResult")
	}
	var nilDeployResult *DeployResult
	if validResult(nilDeployResult) {
		t.Fatal("typed nil DeployResult")
	}
	var nilCreateBranchResult *CreateBranchResult
	if validResult(nilCreateBranchResult) {
		t.Fatal("typed nil CreateBranchResult")
	}
	var nilFetchResult *FetchResult
	if validResult(nilFetchResult) {
		t.Fatal("typed nil FetchResult")
	}
	var nilPullResult *PullResult
	if validResult(nilPullResult) {
		t.Fatal("typed nil PullResult")
	}
	var nilPushResult *PushResult
	if validResult(nilPushResult) {
		t.Fatal("typed nil PushResult")
	}
	var nilStagePathsResult *StagePathsResult
	if validResult(nilStagePathsResult) {
		t.Fatal("typed nil StagePathsResult")
	}
	var nilUnstagePathsResult *UnstagePathsResult
	if validResult(nilUnstagePathsResult) {
		t.Fatal("typed nil UnstagePathsResult")
	}
	var nilStageAllResult *StageAllResult
	if validResult(nilStageAllResult) {
		t.Fatal("typed nil StageAllResult")
	}
	var nilCommitResult *CommitResult
	if validResult(nilCommitResult) {
		t.Fatal("typed nil CommitResult")
	}
	var nilStageAllAndCommitResult *StageAllAndCommitResult
	if validResult(nilStageAllAndCommitResult) {
		t.Fatal("typed nil StageAllAndCommitResult")
	}
	var nilRestoreTrackedResult *RestoreTrackedResult
	if validResult(nilRestoreTrackedResult) {
		t.Fatal("typed nil RestoreTrackedResult")
	}
	var nilStashCreateResult *StashCreateResult
	if validResult(nilStashCreateResult) {
		t.Fatal("typed nil StashCreateResult")
	}
	var nilStashApplyResult *StashApplyResult
	if validResult(nilStashApplyResult) {
		t.Fatal("typed nil StashApplyResult")
	}
	var nilStashPopResult *StashPopResult
	if validResult(nilStashPopResult) {
		t.Fatal("typed nil StashPopResult")
	}
	var nilStashDropResult *StashDropResult
	if validResult(nilStashDropResult) {
		t.Fatal("typed nil StashDropResult")
	}
	var nilCreatePullRequestResult *CreatePullRequestResult
	if validResult(nilCreatePullRequestResult) {
		t.Fatal("typed nil CreatePullRequestResult")
	}
	var nilSaveLaunchResult *SaveLaunchResult
	if validResult(nilSaveLaunchResult) {
		t.Fatal("typed nil SaveLaunchResult")
	}
	var nilStartLaunchResult *StartLaunchResult
	if validResult(nilStartLaunchResult) {
		t.Fatal("typed nil StartLaunchResult")
	}
	var nilOpenTerminalResult *OpenTerminalResult
	if validResult(nilOpenTerminalResult) {
		t.Fatal("typed nil OpenTerminalResult")
	}
	var nilWriteInputResult *WriteInputResult
	if validResult(nilWriteInputResult) {
		t.Fatal("typed nil WriteInputResult")
	}
	var nilResizeSessionResult *ResizeSessionResult
	if validResult(nilResizeSessionResult) {
		t.Fatal("typed nil ResizeSessionResult")
	}
	var nilInterruptSessionResult *InterruptSessionResult
	if validResult(nilInterruptSessionResult) {
		t.Fatal("typed nil InterruptSessionResult")
	}
	var nilStopSessionResult *StopSessionResult
	if validResult(nilStopSessionResult) {
		t.Fatal("typed nil StopSessionResult")
	}
	var nilRestartSessionResult *RestartSessionResult
	if validResult(nilRestartSessionResult) {
		t.Fatal("typed nil RestartSessionResult")
	}
	var nilNavigatorResult *NavigatorResult
	if validResult(nilNavigatorResult) {
		t.Fatal("typed nil NavigatorResult")
	}
	var nilBranchContextResult *BranchContextResult
	if validResult(nilBranchContextResult) {
		t.Fatal("typed nil BranchContextResult")
	}
	var nilCommitsResult *CommitsResult
	if validResult(nilCommitsResult) {
		t.Fatal("typed nil CommitsResult")
	}
	var nilGraphResult *GraphResult
	if validResult(nilGraphResult) {
		t.Fatal("typed nil GraphResult")
	}
	var nilDiffResult *DiffResult
	if validResult(nilDiffResult) {
		t.Fatal("typed nil DiffResult")
	}
	var nilPullRequestDiffResult *PullRequestDiffResult
	if validResult(nilPullRequestDiffResult) {
		t.Fatal("typed nil PullRequestDiffResult")
	}
	var nilWorktreeStatusResult *WorktreeStatusResult
	if validResult(nilWorktreeStatusResult) {
		t.Fatal("typed nil WorktreeStatusResult")
	}
	var nilStashesResult *StashesResult
	if validResult(nilStashesResult) {
		t.Fatal("typed nil StashesResult")
	}
	var nilStashPatchResult *StashPatchResult
	if validResult(nilStashPatchResult) {
		t.Fatal("typed nil StashPatchResult")
	}
	var nilLaunchPointsResult *LaunchPointsResult
	if validResult(nilLaunchPointsResult) {
		t.Fatal("typed nil LaunchPointsResult")
	}
	var nilSessionsResult *SessionsResult
	if validResult(nilSessionsResult) {
		t.Fatal("typed nil SessionsResult")
	}
	var nilSessionOutputProjection *SessionOutputProjection
	if validResult(nilSessionOutputProjection) {
		t.Fatal("typed nil SessionOutputProjection")
	}
	var nilPreferencesResult *PreferencesResult
	if validResult(nilPreferencesResult) {
		t.Fatal("typed nil PreferencesResult")
	}
}

type foreignEventPayload struct{ EventPayload }

func (foreignEventPayload) Valid() bool { return true }
func TestEventPayloadRejectsForeignAndNil(t *testing.T) {
	if validEventPayload(nil) || validEventPayload(foreignEventPayload{}) {
		t.Fatal("foreign or nil EventPayload")
	}
	var nilAccepted *Accepted
	if validEventPayload(nilAccepted) {
		t.Fatal("typed nil Accepted")
	}
	var nilProgress *Progress
	if validEventPayload(nilProgress) {
		t.Fatal("typed nil Progress")
	}
	var nilConfirmationRequested *ConfirmationRequested
	if validEventPayload(nilConfirmationRequested) {
		t.Fatal("typed nil ConfirmationRequested")
	}
	var nilOperationTerminal *OperationTerminal
	if validEventPayload(nilOperationTerminal) {
		t.Fatal("typed nil OperationTerminal")
	}
	var nilSessionChanged *SessionChanged
	if validEventPayload(nilSessionChanged) {
		t.Fatal("typed nil SessionChanged")
	}
	var nilProjectionInvalidated *ProjectionInvalidated
	if validEventPayload(nilProjectionInvalidated) {
		t.Fatal("typed nil ProjectionInvalidated")
	}
}
