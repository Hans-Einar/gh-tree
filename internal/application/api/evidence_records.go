package api

func (e *evidenceSet) collectActivateWorktreeResult(v ActivateWorktreeResult) {
	e.collectActiveContext(v.data.Context)
	e.collectStorageCommitResult(v.data.Storage)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectActiveContext(v ActiveContext) {

	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}

}
func (e *evidenceSet) collectAppliedStashRetained(v AppliedStashRetained) {

	e.collectGitMutationResult(v.data.Apply)
	if item, p := v.data.Drop.Value(); p {
		e.collectGitMutationResult(item)
	}

}
func (e *evidenceSet) collectAppliedWithConflicts(v AppliedWithConflicts) {

	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}
}
func (e *evidenceSet) collectBranchContextResult(v BranchContextResult) {

	if item, p := v.data.Endpoint.Value(); p {
		e.collectExactLocalResolution(item)
	}
	if item, p := v.data.Relationship.Value(); p {
		e.collectBranchRelationship(item)
	}

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectBranchCreated(v BranchCreated) {

	if item, p := v.data.Worktree.Value(); p {
		e.collectWorktreeFacts(item)
	}
}
func (e *evidenceSet) collectBranchOccupancy(v BranchOccupancy) {

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectBranchRelationship(v BranchRelationship) {

	if item, p := v.data.LocalBranch.Value(); p {
		e.collectRefFact(item)
	}
	e.collectUpstreamFact(v.data.Upstream)
	for _, item := range v.data.RemoteEndpoints {
		e.collectRemoteBranchFact(item)
	}

	for _, item := range v.data.Worktrees {
		e.collectWorktreeRelation(item)
	}
	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectCommand(v Command) {
	switch x := v.(type) {
	case PullCommand:
		e.collectPullCommand(x)
	case PushCommand:
		e.collectPushCommand(x)
	case CreatePullRequestCommand:
		e.collectCreatePullRequestCommand(x)
	}
}
func (e *evidenceSet) collectCommitCandidateFacts(v CommitCandidateFacts) {

	e.collectEffectReport(v.data.CandidateEffect)
	e.collectEffectReport(v.data.HookEffects)
	e.collectEffectReport(v.data.RefEffects)
	e.collectEffectReport(v.data.StagedIndexEffect)
}
func (e *evidenceSet) collectCommitCreated(v CommitCreated) {

	e.collectEffectReport(v.data.StagedIndexEffect)
	e.collectCommitCandidateFacts(v.data.Candidate)
}
func (e *evidenceSet) collectCommitResult(v CommitResult) {
	e.collectGitMutationResult(v.data.Git)

	if item, p := v.data.Candidate.Value(); p {
		e.collectCommitCandidateFacts(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectCommitsResult(v CommitsResult) {

	if item, p := v.data.Endpoint.Value(); p {
		e.collectExactLocalResolution(item)
	}

	for _, item := range v.data.Relationships {
		e.collectBranchRelationship(item)
	}
	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectCreateBranchResult(v CreateBranchResult) {
	e.collectGitMutationResult(v.data.Git)

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectCreatePullRequestCommand(v CreatePullRequestCommand) {
	e.collectEndpointExpectation(v.data.Base)
	e.collectEndpointExpectation(v.data.Head)

}
func (e *evidenceSet) collectCreatePullRequestRequest(v CreatePullRequestRequest) {

	e.collectEndpointExpectation(v.data.Base)
	e.collectEndpointExpectation(v.data.Head)

}
func (e *evidenceSet) collectCreatePullRequestResult(v CreatePullRequestResult) {
	e.collectPullRequestCreationOutcome(v.data.Outcome)
	e.collectEffectReport(v.data.Effects)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectRemoteObservation(item)
	}

}
func (e *evidenceSet) collectCreateWorktreeResult(v CreateWorktreeResult) {
	e.collectGitMutationResult(v.data.Git)

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectCreatedVerified(v CreatedVerified) {
	e.collectPullRequestFact(v.data.Created)
	e.collectEndpointExpectation(v.data.RequestedBase)
	e.collectEndpointExpectation(v.data.RequestedHead)
}
func (e *evidenceSet) collectCreatedWithDrift(v CreatedWithDrift) {
	e.collectPullRequestFact(v.data.Created)
	e.collectEndpointExpectation(v.data.RequestedBase)
	e.collectEndpointExpectation(v.data.RequestedHead)

}
func (e *evidenceSet) collectCreationIndeterminate(v CreationIndeterminate) {
	e.collectRemoteCreateEvidence(v.data.RequestEvidence)
	if item, p := v.data.Candidate.Value(); p {
		e.collectPullRequestFact(item)
	}

}
func (e *evidenceSet) collectDeployResult(v DeployResult) {

	if item, p := v.data.Resolution.Value(); p {
		e.collectExactLocalResolution(item)
	}

	for _, item := range v.data.Steps {
		e.collectGitMutationResult(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectDiffResult(v DiffResult) {

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectDiscoveryResult(v DiscoveryResult) {

	e.collectDiscoveryObservation(v.data.Observation)

}
func (e *evidenceSet) collectEffectReport(v EffectReport) {
	for _, item := range v.data.Facets {
		e.collectFacetEffect(item)
	}
}
func (e *evidenceSet) collectEndpointExpectation(v EndpointExpectation) {

	e.collectRemoteObservation(v.data.Observation)
}
func (e *evidenceSet) collectEvent(v Event) {

	e.collectEventPayload(v.data.Payload)
}
func (e *evidenceSet) collectEventPayload(v EventPayload) {
	switch x := v.(type) {
	case OperationTerminal:
		e.collectOperationTerminal(x)
	}
}
func (e *evidenceSet) collectExactLocalResolution(v ExactLocalResolution) {

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectExistingCandidate(v ExistingCandidate) {
	for _, item := range v.data.Candidates {
		e.collectPullRequestFact(item)
	}

}
func (e *evidenceSet) collectFetchFreshness(v FetchFreshness) {

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectFetchResult(v FetchResult) {

	for _, item := range v.data.Refs {
		e.collectRefFact(item)
	}
	e.collectFetchFreshness(v.data.Freshness)
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

	e.collectEffectReport(v.data.Effects)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectGitCompletedStep(v GitCompletedStep) {

	e.collectFacetEffect(v.data.Effect)
	if item, p := v.data.PostObservation.Value(); p {
		e.collectGitObservation(item)
	}
}
func (e *evidenceSet) collectGitMutationOutcome(v GitMutationOutcome) {
	switch x := v.(type) {
	case WorktreeCreated:
		e.collectWorktreeCreated(x)
	case WorktreeRetargeted:
		e.collectWorktreeRetargeted(x)
	case IndexChanged:
		e.collectIndexChanged(x)
	case CommitCreated:
		e.collectCommitCreated(x)
	case TrackedRestored:
		e.collectTrackedRestored(x)
	case StashCreated:
		e.collectStashCreated(x)
	case StashCreatedCleanupRefused:
		e.collectStashCreatedCleanupRefused(x)
	case StashApplied:
		e.collectStashApplied(x)
	case AppliedWithConflicts:
		e.collectAppliedWithConflicts(x)
	case StashDropped:
		e.collectStashDropped(x)
	case BranchCreated:
		e.collectBranchCreated(x)
	case Pushed:
		e.collectPushed(x)
	case RefusedMutation:
		e.collectRefusedMutation(x)
	case PartialMutation:
		e.collectPartialMutation(x)
	case MutationIndeterminate:
		e.collectMutationIndeterminate(x)
	}
}
func (e *evidenceSet) collectGitMutationResult(v GitMutationResult) {

	for _, item := range v.data.Steps {
		e.collectGitCompletedStep(item)
	}
	e.collectGitMutationOutcome(v.data.Outcome)
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

	e.collectEffectReport(v.data.Effects)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectGitPostFacts(v GitPostFacts) {
	for _, item := range v.data.Worktrees {
		e.collectWorktreeFacts(item)
	}
	for _, item := range v.data.Status {
		e.collectStatusFacts(item)
	}
	for _, item := range v.data.Refs {
		e.collectRefFact(item)
	}
	for _, item := range v.data.Stashes {
		e.collectStashFact(item)
	}
	if item, p := v.data.Commit.Value(); p {
		e.collectCommitCandidateFacts(item)
	}
	for _, item := range v.data.Configuration {
		e.collectLocalConfigurationObservation(item)
	}
}
func (e *evidenceSet) collectGitPreparationResult(v GitPreparationResult) {

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

	e.collectEffectReport(v.data.Effects)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectGoneUpstream(v GoneUpstream) {

	e.collectGitObservation(v.data.Evidence)
}
func (e *evidenceSet) collectGraphAnnotation(v GraphAnnotation) {

	for _, item := range v.data.Branches {
		e.collectBranchRelationship(item)
	}

	for _, item := range v.data.Worktrees {
		e.collectWorktreeRelation(item)
	}
}
func (e *evidenceSet) collectGraphResult(v GraphResult) {

	for _, item := range v.data.Refs {
		e.collectRefFact(item)
	}
	for _, item := range v.data.Annotations {
		e.collectGraphAnnotation(item)
	}

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectIndexChanged(v IndexChanged) {
	e.collectStatusFacts(v.data.Status)

}
func (e *evidenceSet) collectInterruptSessionResult(v InterruptSessionResult) {

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectLaunchPointsResult(v LaunchPointsResult) {

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectListBranchesResult(v ListBranchesResult) {
	for _, item := range v.data.Branches {
		e.collectRemoteBranchFact(item)
	}

	if item, p := v.data.Observation.Value(); p {
		e.collectRemoteObservation(item)
	}

}
func (e *evidenceSet) collectListPullRequestsResult(v ListPullRequestsResult) {
	for _, item := range v.data.PullRequests {
		e.collectPullRequestFact(item)
	}

	if item, p := v.data.Observation.Value(); p {
		e.collectRemoteObservation(item)
	}

}
func (e *evidenceSet) collectListRefsResult(v ListRefsResult) {
	for _, item := range v.data.Refs {
		e.collectRefFact(item)
	}

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectListStashesResult(v ListStashesResult) {
	for _, item := range v.data.Stashes {
		e.collectStashFact(item)
	}

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectListWorktreesResult(v ListWorktreesResult) {
	for _, item := range v.data.Worktrees {
		e.collectWorktreeFacts(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectLocalConfigurationObservation(v LocalConfigurationObservation) {

	e.collectUpstreamFact(v.data.Upstream)

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectLocalRepositoryFacts(v LocalRepositoryFacts) {

	for _, item := range v.data.Worktrees {
		e.collectWorktreeFacts(item)
	}

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectMergeBaseResult(v MergeBaseResult) {

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectMutationIndeterminate(v MutationIndeterminate) {
	e.collectGitPostFacts(v.data.Facts)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectNavigatorResult(v NavigatorResult) {

	for _, item := range v.data.PullRequests {
		e.collectPullRequestFact(item)
	}
	for _, item := range v.data.Branches {
		e.collectBranchRelationship(item)
	}
	for _, item := range v.data.Worktrees {
		e.collectWorktreeFacts(item)
	}

	e.collectActiveContext(v.data.Active)

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectNormalizedRecovery(v NormalizedRecovery) {
	e.collectRecoveryRecord(v.data.Record)
	if item, p := v.data.StorageDetail.Value(); p {
		e.collectStorageRecovery(item)
	}
}
func (e *evidenceSet) collectObservePullRequestResult(v ObservePullRequestResult) {
	if item, p := v.data.PullRequest.Value(); p {
		e.collectPullRequestFact(item)
	}

	if item, p := v.data.Observation.Value(); p {
		e.collectRemoteObservation(item)
	}

}
func (e *evidenceSet) collectObserveStatusResult(v ObserveStatusResult) {
	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectOpenTerminalResult(v OpenTerminalResult) {
	e.collectSessionStartResult(v.data.Start)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectOperationResidual(v OperationResidual) {

	e.collectEffectReport(v.data.Effects)
	for _, item := range v.data.Recovery {
		e.collectNormalizedRecovery(item)
	}

}
func (e *evidenceSet) collectOperationTerminal(v OperationTerminal) {

	if item, p := v.data.Result.Value(); p {
		e.collectResult(item)
	}

	e.collectEffectReport(v.data.Effects)
	for _, item := range v.data.Recovery {
		e.collectNormalizedRecovery(item)
	}

}
func (e *evidenceSet) collectOutcomeEnvelope(v OutcomeEnvelope) {
	e.collectEffectReport(v.data.Effects)

	for _, item := range v.data.Recovery {
		e.collectNormalizedRecovery(item)
	}

}
func (e *evidenceSet) collectPartialMutation(v PartialMutation) {
	e.collectGitPostFacts(v.data.Facts)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectPreferencesResult(v PreferencesResult) {

	e.collectActiveContext(v.data.Active)
	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectPrepareCreateRequest(v PrepareCreateRequest) {

	e.collectExactLocalResolution(v.data.Target)

}
func (e *evidenceSet) collectPreparePushRequest(v PreparePushRequest) {

	if item, p := v.data.SetUpstream.Value(); p {
		e.collectUpstreamSetup(item)
	}

}
func (e *evidenceSet) collectPrepareRetargetRequest(v PrepareRetargetRequest) {

	e.collectExactLocalResolution(v.data.Target)

}
func (e *evidenceSet) collectProjectionSources(v ProjectionSources) {
	for _, item := range v.data.Git {
		e.collectGitObservation(item)
	}
	for _, item := range v.data.Remote {
		e.collectRemoteObservation(item)
	}
	for _, item := range v.data.Storage {
		e.collectStorageProjectionSource(item)
	}
	for _, item := range v.data.Discovery {
		e.collectDiscoveryObservation(item)
	}

}
func (e *evidenceSet) collectPullCommand(v PullCommand) {

	e.collectResolvedUpstream(v.data.Upstream)
}
func (e *evidenceSet) collectPullRequestCreationOutcome(v PullRequestCreationOutcome) {
	switch x := v.(type) {
	case ExistingCandidate:
		e.collectExistingCandidate(x)
	case CreatedVerified:
		e.collectCreatedVerified(x)
	case CreatedWithDrift:
		e.collectCreatedWithDrift(x)
	case CreationIndeterminate:
		e.collectCreationIndeterminate(x)
	}
}
func (e *evidenceSet) collectPullRequestDiffResult(v PullRequestDiffResult) {

	if item, p := v.data.RemotePR.Value(); p {
		e.collectPullRequestFact(item)
	}
	if item, p := v.data.Base.Value(); p {
		e.collectExactLocalResolution(item)
	}
	if item, p := v.data.Head.Value(); p {
		e.collectExactLocalResolution(item)
	}

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectPullRequestFact(v PullRequestFact) {

	e.collectRemoteObservation(v.data.Observation)

}
func (e *evidenceSet) collectPullResult(v PullResult) {
	if item, p := v.data.Fetch.Value(); p {
		e.collectFetchResult(item)
	}

	if item, p := v.data.FastForward.Value(); p {
		e.collectGitMutationResult(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectPushCommand(v PushCommand) {

	if item, p := v.data.SetUpstream.Value(); p {
		e.collectUpstreamSetup(item)
	}

}
func (e *evidenceSet) collectPushResult(v PushResult) {
	e.collectGitMutationResult(v.data.Git)
	e.collectEffectReport(v.data.RemoteEffect)
	e.collectEffectReport(v.data.UpstreamEffect)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectPushed(v Pushed) {

	e.collectEffectReport(v.data.RemoteEffect)
	e.collectEffectReport(v.data.UpstreamEffect)
	if item, p := v.data.Configuration.Value(); p {
		e.collectLocalConfigurationObservation(item)
	}
}
func (e *evidenceSet) collectReadCommitsResult(v ReadCommitsResult) {

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectReadDiffResult(v ReadDiffResult) {

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectReadGraphResult(v ReadGraphResult) {

	for _, item := range v.data.Refs {
		e.collectRefFact(item)
	}
	for _, item := range v.data.Heads {
		e.collectWorktreeFacts(item)
	}

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectReadStashPatchResult(v ReadStashPatchResult) {

	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectReconcileRequest(v ReconcileRequest) {

	e.collectEffectReport(v.data.PriorEffects)
	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectReconcileResult(v ReconcileResult) {
	for _, item := range v.data.Facets {
		e.collectReconciledFacet(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

	e.collectEffectReport(v.data.Effects)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}

}
func (e *evidenceSet) collectReconciledFacet(v ReconciledFacet) {

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}
	for _, item := range v.data.Refs {
		e.collectRefFact(item)
	}
	for _, item := range v.data.Statuses {
		e.collectStatusFacts(item)
	}
	for _, item := range v.data.Stashes {
		e.collectStashFact(item)
	}
	for _, item := range v.data.Configuration {
		e.collectLocalConfigurationObservation(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectRefFact(v RefFact) {

	if item, p := v.data.Freshness.Value(); p {
		e.collectFetchFreshness(item)
	}

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectRefusedMutation(v RefusedMutation) {

	e.collectEffectReport(v.data.Effects)
}
func (e *evidenceSet) collectRemoteBranchFact(v RemoteBranchFact) {

	e.collectRemoteObservation(v.data.Observation)
}
func (e *evidenceSet) collectRemoteCreateEvidence(v RemoteCreateEvidence) {

	e.collectEndpointExpectation(v.data.RequestedBase)
	e.collectEndpointExpectation(v.data.RequestedHead)

}
func (e *evidenceSet) collectResizeSessionResult(v ResizeSessionResult) {

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectResolveExactResult(v ResolveExactResult) {
	if item, p := v.data.Resolution.Value(); p {
		e.collectExactLocalResolution(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectResolveLaunchResult(v ResolveLaunchResult) {

	e.collectDiscoveryObservation(v.data.Observation)

}
func (e *evidenceSet) collectResolveLocalResult(v ResolveLocalResult) {
	if item, p := v.data.Repository.Value(); p {
		e.collectLocalRepositoryFacts(item)
	}
	if item, p := v.data.Observation.Value(); p {
		e.collectGitObservation(item)
	}

}
func (e *evidenceSet) collectResolveRepositoryResult(v ResolveRepositoryResult) {

	if item, p := v.data.Observation.Value(); p {
		e.collectRemoteObservation(item)
	}

}
func (e *evidenceSet) collectResolvedUpstream(v ResolvedUpstream) {

	e.collectRevisionComparison(v.data.Comparison)
	e.collectFetchFreshness(v.data.Freshness)
}
func (e *evidenceSet) collectRestartSessionResult(v RestartSessionResult) {
	e.collectSessionRestartResult(v.data.Restart)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectRestoreTrackedResult(v RestoreTrackedResult) {
	e.collectGitMutationResult(v.data.Git)
	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectResult(v Result) {
	switch x := v.(type) {
	case ActivateWorktreeResult:
		e.collectActivateWorktreeResult(x)
	case SaveNavigationResult:
		e.collectSaveNavigationResult(x)
	case CreateWorktreeResult:
		e.collectCreateWorktreeResult(x)
	case RetargetWorktreeResult:
		e.collectRetargetWorktreeResult(x)
	case DeployResult:
		e.collectDeployResult(x)
	case CreateBranchResult:
		e.collectCreateBranchResult(x)
	case FetchResult:
		e.collectFetchResult(x)
	case PullResult:
		e.collectPullResult(x)
	case PushResult:
		e.collectPushResult(x)
	case StagePathsResult:
		e.collectStagePathsResult(x)
	case UnstagePathsResult:
		e.collectUnstagePathsResult(x)
	case StageAllResult:
		e.collectStageAllResult(x)
	case CommitResult:
		e.collectCommitResult(x)
	case StageAllAndCommitResult:
		e.collectStageAllAndCommitResult(x)
	case RestoreTrackedResult:
		e.collectRestoreTrackedResult(x)
	case StashCreateResult:
		e.collectStashCreateResult(x)
	case StashApplyResult:
		e.collectStashApplyResult(x)
	case StashPopResult:
		e.collectStashPopResult(x)
	case StashDropResult:
		e.collectStashDropResult(x)
	case CreatePullRequestResult:
		e.collectCreatePullRequestResult(x)
	case SaveLaunchResult:
		e.collectSaveLaunchResult(x)
	case StartLaunchResult:
		e.collectStartLaunchResult(x)
	case OpenTerminalResult:
		e.collectOpenTerminalResult(x)
	case WriteInputResult:
		e.collectWriteInputResult(x)
	case ResizeSessionResult:
		e.collectResizeSessionResult(x)
	case InterruptSessionResult:
		e.collectInterruptSessionResult(x)
	case StopSessionResult:
		e.collectStopSessionResult(x)
	case RestartSessionResult:
		e.collectRestartSessionResult(x)
	case NavigatorResult:
		e.collectNavigatorResult(x)
	case BranchContextResult:
		e.collectBranchContextResult(x)
	case CommitsResult:
		e.collectCommitsResult(x)
	case GraphResult:
		e.collectGraphResult(x)
	case DiffResult:
		e.collectDiffResult(x)
	case PullRequestDiffResult:
		e.collectPullRequestDiffResult(x)
	case WorktreeStatusResult:
		e.collectWorktreeStatusResult(x)
	case StashesResult:
		e.collectStashesResult(x)
	case StashPatchResult:
		e.collectStashPatchResult(x)
	case LaunchPointsResult:
		e.collectLaunchPointsResult(x)
	case SessionsResult:
		e.collectSessionsResult(x)
	case SessionOutputProjection:
		e.collectSessionOutputProjection(x)
	case PreferencesResult:
		e.collectPreferencesResult(x)
	}
}
func (e *evidenceSet) collectRetargetWorktreeResult(v RetargetWorktreeResult) {
	e.collectGitMutationResult(v.data.Git)

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectRevisionComparison(v RevisionComparison) {

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectRuntimeShutdownResult(v RuntimeShutdownResult) {

	for _, item := range v.data.Sessions {
		e.collectSessionStopResult(item)
	}

}
func (e *evidenceSet) collectSaveLaunchResult(v SaveLaunchResult) {

	e.collectStorageCommitResult(v.data.Storage)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectSaveNavigationResult(v SaveNavigationResult) {

	e.collectStorageCommitResult(v.data.Storage)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectSessionOutputProjection(v SessionOutputProjection) {

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectSessionRestartResult(v SessionRestartResult) {
	e.collectSessionStopResult(v.data.Old)
	if item, p := v.data.Replacement.Value(); p {
		e.collectSessionStartResult(item)
	}

}
func (e *evidenceSet) collectSessionStartResult(v SessionStartResult) {

	e.collectEffectReport(v.data.Effects)

}
func (e *evidenceSet) collectSessionStopResult(v SessionStopResult) {

	e.collectEffectReport(v.data.Effects)

}
func (e *evidenceSet) collectSessionsResult(v SessionsResult) {

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectShutdownResult(v ShutdownResult) {

	for _, item := range v.data.Operations {
		e.collectOperationResidual(item)
	}
	e.collectRuntimeShutdownResult(v.data.Sessions)

	for _, item := range v.data.Recovery {
		e.collectNormalizedRecovery(item)
	}

}
func (e *evidenceSet) collectStageAllAndCommitResult(v StageAllAndCommitResult) {
	if item, p := v.data.Stage.Value(); p {
		e.collectGitMutationResult(item)
	}
	if item, p := v.data.Commit.Value(); p {
		e.collectGitMutationResult(item)
	}

	if item, p := v.data.Candidate.Value(); p {
		e.collectCommitCandidateFacts(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStageAllResult(v StageAllResult) {
	e.collectGitMutationResult(v.data.Git)
	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStagePathsResult(v StagePathsResult) {
	e.collectGitMutationResult(v.data.Git)
	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStartLaunchResult(v StartLaunchResult) {
	e.collectSessionStartResult(v.data.Start)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStashApplied(v StashApplied) {

	e.collectStatusFacts(v.data.Status)

}
func (e *evidenceSet) collectStashApplyResult(v StashApplyResult) {
	e.collectGitMutationResult(v.data.Git)

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStashCreateResult(v StashCreateResult) {
	e.collectGitMutationResult(v.data.Git)

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStashCreated(v StashCreated) {

	e.collectStatusFacts(v.data.Status)
	e.collectEffectReport(v.data.Cleanup)
}
func (e *evidenceSet) collectStashCreatedCleanupRefused(v StashCreatedCleanupRefused) {

	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	e.collectEffectReport(v.data.Cleanup)

	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}
}
func (e *evidenceSet) collectStashDropResult(v StashDropResult) {
	e.collectGitMutationResult(v.data.Git)

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStashDropped(v StashDropped) {

	for _, item := range v.data.Survivors {
		e.collectStashFact(item)
	}
	e.collectGitObservation(v.data.Observation)
	e.collectEffectReport(v.data.RefCleanup)
}
func (e *evidenceSet) collectStashFact(v StashFact) {

	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectStashPatchResult(v StashPatchResult) {

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectStashPopCompleted(v StashPopCompleted) {

	e.collectGitMutationResult(v.data.Apply)
	e.collectGitMutationResult(v.data.Drop)
}
func (e *evidenceSet) collectStashPopNotApplied(v StashPopNotApplied) {

	e.collectGitMutationResult(v.data.Apply)
}
func (e *evidenceSet) collectStashPopOutcome(v StashPopOutcome) {
	switch x := v.(type) {
	case StashPopCompleted:
		e.collectStashPopCompleted(x)
	case AppliedStashRetained:
		e.collectAppliedStashRetained(x)
	case StashPopNotApplied:
		e.collectStashPopNotApplied(x)
	}
}
func (e *evidenceSet) collectStashPopResult(v StashPopResult) {
	e.collectStashPopOutcome(v.data.Outcome)
	e.collectOutcomeEnvelope(v.data.Envelope)
}
func (e *evidenceSet) collectStashesResult(v StashesResult) {

	for _, item := range v.data.Stashes {
		e.collectStashFact(item)
	}

	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectStatusFacts(v StatusFacts) {
	e.collectWorktreeFacts(v.data.Worktree)

	e.collectUpstreamFact(v.data.Upstream)
	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectStopSessionResult(v StopSessionResult) {
	e.collectSessionStopResult(v.data.Stop)
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectStorageCommitResult(v StorageCommitResult) {

	e.collectEffectReport(v.data.Effects)
	for _, item := range v.data.Recovery {
		e.collectStorageRecovery(item)
	}

}
func (e *evidenceSet) collectStorageLoadObservation(v StorageLoadObservation) {

	for _, item := range v.data.Recovery {
		e.collectStorageRecovery(item)
	}
}
func (e *evidenceSet) collectStorageProjectionSource(v StorageProjectionSource) {
	e.collectStorageLoadObservation(v.data.Observation)

}
func (e *evidenceSet) collectTrackedRestored(v TrackedRestored) {

	e.collectStatusFacts(v.data.Status)
	for _, item := range v.data.Recovery {
		e.collectRecoveryRecord(item)
	}
}
func (e *evidenceSet) collectUnstagePathsResult(v UnstagePathsResult) {
	e.collectGitMutationResult(v.data.Git)
	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	e.collectOutcomeEnvelope(v.data.Outcome)
}
func (e *evidenceSet) collectUpstreamFact(v UpstreamFact) {
	switch x := v.(type) {
	case GoneUpstream:
		e.collectGoneUpstream(x)
	case ResolvedUpstream:
		e.collectResolvedUpstream(x)
	}
}
func (e *evidenceSet) collectUpstreamSetup(v UpstreamSetup) {

	e.collectUpstreamFact(v.data.ExpectedValue)
}
func (e *evidenceSet) collectWorktreeCreated(v WorktreeCreated) {
	e.collectWorktreeFacts(v.data.Worktree)
	e.collectExactLocalResolution(v.data.Target)
}
func (e *evidenceSet) collectWorktreeFacts(v WorktreeFacts) {

	for _, item := range v.data.Occupancy {
		e.collectBranchOccupancy(item)
	}
	e.collectGitObservation(v.data.Observation)
}
func (e *evidenceSet) collectWorktreeRelation(v WorktreeRelation) {

	e.collectGitObservation(v.data.IdentitySource)

}
func (e *evidenceSet) collectWorktreeRetargeted(v WorktreeRetargeted) {
	e.collectWorktreeFacts(v.data.Worktree)

	e.collectExactLocalResolution(v.data.Target)
}
func (e *evidenceSet) collectWorktreeStatusResult(v WorktreeStatusResult) {

	if item, p := v.data.Status.Value(); p {
		e.collectStatusFacts(item)
	}
	e.collectProjectionSources(v.data.Sources)

}
func (e *evidenceSet) collectWriteInputResult(v WriteInputResult) {

	e.collectOutcomeEnvelope(v.data.Outcome)
}
func validateActivateWorktreeResultEvidence(d ActivateWorktreeResultData) error {
	e := newEvidenceSet()
	e.collectActivateWorktreeResult(ActivateWorktreeResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateCommitResultEvidence(d CommitResultData) error {
	e := newEvidenceSet()
	e.collectCommitResult(CommitResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateCreateBranchResultEvidence(d CreateBranchResultData) error {
	e := newEvidenceSet()
	e.collectCreateBranchResult(CreateBranchResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateCreatePullRequestResultEvidence(d CreatePullRequestResultData) error {
	e := newEvidenceSet()
	e.collectCreatePullRequestResult(CreatePullRequestResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.recordUnion(d.Recovery)
}
func validateCreateWorktreeResultEvidence(d CreateWorktreeResultData) error {
	e := newEvidenceSet()
	e.collectCreateWorktreeResult(CreateWorktreeResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateDeployResultEvidence(d DeployResultData) error {
	e := newEvidenceSet()
	e.collectDeployResult(DeployResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateFetchResultEvidence(d FetchResultData) error {
	e := newEvidenceSet()
	e.collectFetchResult(FetchResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.recordUnion(d.Recovery)
}
func validateGitMutationResultEvidence(d GitMutationResultData) error {
	e := newEvidenceSet()
	e.collectGitMutationResult(GitMutationResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.recordUnion(d.Recovery)
}
func validateGitPreparationResultEvidence(d GitPreparationResultData) error {
	e := newEvidenceSet()
	e.collectGitPreparationResult(GitPreparationResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.recordUnion(d.Recovery)
}
func validateInterruptSessionResultEvidence(d InterruptSessionResultData) error {
	e := newEvidenceSet()
	e.collectInterruptSessionResult(InterruptSessionResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateOpenTerminalResultEvidence(d OpenTerminalResultData) error {
	e := newEvidenceSet()
	e.collectOpenTerminalResult(OpenTerminalResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateOperationTerminalEvidence(d OperationTerminalData) error {
	e := newEvidenceSet()
	e.collectOperationTerminal(OperationTerminal{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Recovery)
}
func validatePullResultEvidence(d PullResultData) error {
	e := newEvidenceSet()
	e.collectPullResult(PullResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validatePushResultEvidence(d PushResultData) error {
	e := newEvidenceSet()
	e.collectPushResult(PushResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateReconcileResultEvidence(d ReconcileResultData) error {
	e := newEvidenceSet()
	e.collectReconcileResult(ReconcileResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.recordUnion(d.Recovery)
}
func validateResizeSessionResultEvidence(d ResizeSessionResultData) error {
	e := newEvidenceSet()
	e.collectResizeSessionResult(ResizeSessionResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateRestartSessionResultEvidence(d RestartSessionResultData) error {
	e := newEvidenceSet()
	e.collectRestartSessionResult(RestartSessionResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateRestoreTrackedResultEvidence(d RestoreTrackedResultData) error {
	e := newEvidenceSet()
	e.collectRestoreTrackedResult(RestoreTrackedResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateRetargetWorktreeResultEvidence(d RetargetWorktreeResultData) error {
	e := newEvidenceSet()
	e.collectRetargetWorktreeResult(RetargetWorktreeResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateRuntimeShutdownResultEvidence(d RuntimeShutdownResultData) error {
	e := newEvidenceSet()
	e.collectRuntimeShutdownResult(RuntimeShutdownResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return nil
}
func validateSaveLaunchResultEvidence(d SaveLaunchResultData) error {
	e := newEvidenceSet()
	e.collectSaveLaunchResult(SaveLaunchResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateSaveNavigationResultEvidence(d SaveNavigationResultData) error {
	e := newEvidenceSet()
	e.collectSaveNavigationResult(SaveNavigationResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateSessionRestartResultEvidence(d SessionRestartResultData) error {
	e := newEvidenceSet()
	e.collectSessionRestartResult(SessionRestartResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return nil
}
func validateSessionStartResultEvidence(d SessionStartResultData) error {
	e := newEvidenceSet()
	e.collectSessionStartResult(SessionStartResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return nil
}
func validateSessionStopResultEvidence(d SessionStopResultData) error {
	e := newEvidenceSet()
	e.collectSessionStopResult(SessionStopResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return nil
}
func validateShutdownResultEvidence(d ShutdownResultData) error {
	e := newEvidenceSet()
	e.collectShutdownResult(ShutdownResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Recovery)
}
func validateStageAllAndCommitResultEvidence(d StageAllAndCommitResultData) error {
	e := newEvidenceSet()
	e.collectStageAllAndCommitResult(StageAllAndCommitResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStageAllResultEvidence(d StageAllResultData) error {
	e := newEvidenceSet()
	e.collectStageAllResult(StageAllResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStagePathsResultEvidence(d StagePathsResultData) error {
	e := newEvidenceSet()
	e.collectStagePathsResult(StagePathsResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStartLaunchResultEvidence(d StartLaunchResultData) error {
	e := newEvidenceSet()
	e.collectStartLaunchResult(StartLaunchResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStashApplyResultEvidence(d StashApplyResultData) error {
	e := newEvidenceSet()
	e.collectStashApplyResult(StashApplyResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStashCreateResultEvidence(d StashCreateResultData) error {
	e := newEvidenceSet()
	e.collectStashCreateResult(StashCreateResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStashDropResultEvidence(d StashDropResultData) error {
	e := newEvidenceSet()
	e.collectStashDropResult(StashDropResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStashPopResultEvidence(d StashPopResultData) error {
	e := newEvidenceSet()
	e.collectStashPopResult(StashPopResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Envelope.data.Recovery)
}
func validateStopSessionResultEvidence(d StopSessionResultData) error {
	e := newEvidenceSet()
	e.collectStopSessionResult(StopSessionResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateStorageCommitResultEvidence(d StorageCommitResultData) error {
	e := newEvidenceSet()
	e.collectStorageCommitResult(StorageCommitResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return nil
}
func validateUnstagePathsResultEvidence(d UnstagePathsResultData) error {
	e := newEvidenceSet()
	e.collectUnstagePathsResult(UnstagePathsResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
func validateWriteInputResultEvidence(d WriteInputResultData) error {
	e := newEvidenceSet()
	e.collectWriteInputResult(WriteInputResult{data: d})
	if err := e.validate(); err != nil {
		return err
	}
	return e.normalizedUnion(d.Outcome.data.Recovery)
}
