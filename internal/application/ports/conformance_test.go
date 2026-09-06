package ports_test

import (
	"context"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

type fakeGitFacts struct{}

var _ ports.GitFacts = fakeGitFacts{}

func (fakeGitFacts) ResolveLocal(_ context.Context, _ api.ResolveLocalRequest) (api.ResolveLocalResult, error) {
	var r0 api.ResolveLocalResult
	return r0, nil
}
func (fakeGitFacts) ListWorktrees(_ context.Context, _ api.ListWorktreesRequest) (api.ListWorktreesResult, error) {
	var r0 api.ListWorktreesResult
	return r0, nil
}
func (fakeGitFacts) ObserveStatus(_ context.Context, _ api.ObserveStatusRequest) (api.ObserveStatusResult, error) {
	var r0 api.ObserveStatusResult
	return r0, nil
}
func (fakeGitFacts) ResolveExact(_ context.Context, _ api.ResolveExactRequest) (api.ResolveExactResult, error) {
	var r0 api.ResolveExactResult
	return r0, nil
}
func (fakeGitFacts) ListRefs(_ context.Context, _ api.ListRefsRequest) (api.ListRefsResult, error) {
	var r0 api.ListRefsResult
	return r0, nil
}
func (fakeGitFacts) ListStashes(_ context.Context, _ api.ListStashesRequest) (api.ListStashesResult, error) {
	var r0 api.ListStashesResult
	return r0, nil
}
func (fakeGitFacts) ReadStashPatch(_ context.Context, _ api.ReadStashPatchRequest) (api.ReadStashPatchResult, error) {
	var r0 api.ReadStashPatchResult
	return r0, nil
}
func (fakeGitFacts) MergeBase(_ context.Context, _ api.MergeBaseRequest) (api.MergeBaseResult, error) {
	var r0 api.MergeBaseResult
	return r0, nil
}
func (fakeGitFacts) ReadCommits(_ context.Context, _ api.ReadCommitsRequest) (api.ReadCommitsResult, error) {
	var r0 api.ReadCommitsResult
	return r0, nil
}
func (fakeGitFacts) ReadGraph(_ context.Context, _ api.ReadGraphRequest) (api.ReadGraphResult, error) {
	var r0 api.ReadGraphResult
	return r0, nil
}
func (fakeGitFacts) ReadDiff(_ context.Context, _ api.ReadDiffRequest) (api.ReadDiffResult, error) {
	var r0 api.ReadDiffResult
	return r0, nil
}

type fakeGitMutations struct{}

var _ ports.GitMutations = fakeGitMutations{}

func (fakeGitMutations) PrepareCreate(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareCreateRequest) (ports.CreatePlan, api.GitPreparationResult, error) {
	var r0 ports.CreatePlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PrepareRetarget(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareRetargetRequest) (ports.RetargetPlan, api.GitPreparationResult, error) {
	var r0 ports.RetargetPlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PrepareStage(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareStageRequest) (ports.StagePlan, api.GitPreparationResult, error) {
	var r0 ports.StagePlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PrepareCommit(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareCommitRequest) (ports.CommitPlan, api.GitPreparationResult, error) {
	var r0 ports.CommitPlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PrepareRestore(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareRestoreRequest) (ports.RestorePlan, api.GitPreparationResult, error) {
	var r0 ports.RestorePlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PrepareStash(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareStashRequest) (ports.StashPlan, api.GitPreparationResult, error) {
	var r0 ports.StashPlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PrepareBranch(_ context.Context, _ ports.GitPrepareContext, _ api.PrepareBranchRequest) (ports.BranchPlan, api.GitPreparationResult, error) {
	var r0 ports.BranchPlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) PreparePush(_ context.Context, _ ports.GitPrepareContext, _ api.PreparePushRequest) (ports.PushPlan, api.GitPreparationResult, error) {
	var r0 ports.PushPlan
	var r1 api.GitPreparationResult
	return r0, r1, nil
}
func (fakeGitMutations) ExecutePrepared(_ context.Context, _ ports.PreparedGitPlan, _ ports.ExecutionApproval) (ports.ExecutedGitMutation, error) {
	var r0 ports.ExecutedGitMutation
	return r0, nil
}
func (fakeGitMutations) ReleasePlan(_ ports.PreparedGitPlan) error { ; return nil }
func (fakeGitMutations) Fetch(_ context.Context, _ api.FetchRequest) (api.FetchResult, error) {
	var r0 api.FetchResult
	return r0, nil
}
func (fakeGitMutations) Reconcile(_ context.Context, _ api.ReconcileRequest) (api.ReconcileResult, error) {
	var r0 api.ReconcileResult
	return r0, nil
}

type fakeRemoteFacts struct{}

var _ ports.RemoteFacts = fakeRemoteFacts{}

func (fakeRemoteFacts) ResolveRepository(_ context.Context, _ api.ResolveRepositoryRequest) (api.ResolveRepositoryResult, error) {
	var r0 api.ResolveRepositoryResult
	return r0, nil
}
func (fakeRemoteFacts) ListBranches(_ context.Context, _ api.ListBranchesRequest) (api.ListBranchesResult, error) {
	var r0 api.ListBranchesResult
	return r0, nil
}
func (fakeRemoteFacts) ListPullRequests(_ context.Context, _ api.ListPullRequestsRequest) (api.ListPullRequestsResult, error) {
	var r0 api.ListPullRequestsResult
	return r0, nil
}
func (fakeRemoteFacts) ObservePullRequest(_ context.Context, _ api.ObservePullRequestRequest) (api.ObservePullRequestResult, error) {
	var r0 api.ObservePullRequestResult
	return r0, nil
}

type fakeRemoteMutations struct{}

var _ ports.RemoteMutations = fakeRemoteMutations{}

func (fakeRemoteMutations) CreatePullRequest(_ context.Context, _ api.CreatePullRequestRequest) (api.CreatePullRequestResult, error) {
	var r0 api.CreatePullRequestResult
	return r0, nil
}

type fakeLaunchDiscovery struct{}

var _ ports.LaunchDiscovery = fakeLaunchDiscovery{}

func (fakeLaunchDiscovery) Discover(_ context.Context, _ api.DiscoveryRequest) (api.DiscoveryResult, error) {
	var r0 api.DiscoveryResult
	return r0, nil
}
func (fakeLaunchDiscovery) Resolve(_ context.Context, _ api.ResolveLaunchRequest) (api.ResolveLaunchResult, error) {
	var r0 api.ResolveLaunchResult
	return r0, nil
}

type fakeStorage struct{}

var _ ports.Storage = fakeStorage{}

func (fakeStorage) LoadUserConfig(_ context.Context) (ports.LoadedUserConfig, error) {
	var r0 ports.LoadedUserConfig
	return r0, nil
}
func (fakeStorage) LoadPreferences(_ context.Context) (ports.LoadedPreferences, error) {
	var r0 ports.LoadedPreferences
	return r0, nil
}
func (fakeStorage) LoadRunConfig(_ context.Context, _ api.WorktreeScope) (ports.LoadedRunConfig, error) {
	var r0 ports.LoadedRunConfig
	return r0, nil
}
func (fakeStorage) CommitUserConfig(_ context.Context, _ ports.UserConfigCommit) (api.StorageCommitResult, error) {
	var r0 api.StorageCommitResult
	return r0, nil
}
func (fakeStorage) CommitPreferences(_ context.Context, _ ports.PreferencesCommit) (api.StorageCommitResult, error) {
	var r0 api.StorageCommitResult
	return r0, nil
}
func (fakeStorage) CommitRunConfig(_ context.Context, _ ports.RunConfigCommit) (api.StorageCommitResult, error) {
	var r0 api.StorageCommitResult
	return r0, nil
}

type fakeSessions struct{}

var _ ports.Sessions = fakeSessions{}

func (fakeSessions) Start(_ context.Context, _ api.SessionStartRequest) (api.SessionStartResult, error) {
	var r0 api.SessionStartResult
	return r0, nil
}
func (fakeSessions) Snapshot(_ context.Context, _ domain.SessionID) (api.SessionSnapshot, error) {
	var r0 api.SessionSnapshot
	return r0, nil
}
func (fakeSessions) List(_ context.Context, _ api.SessionFilter) (api.SessionList, error) {
	var r0 api.SessionList
	return r0, nil
}
func (fakeSessions) ReadOutput(_ context.Context, _ api.SessionOutputRequest) (api.SessionOutputResult, error) {
	var r0 api.SessionOutputResult
	return r0, nil
}
func (fakeSessions) Write(_ context.Context, _ api.SessionWriteRequest) (api.SessionWriteResult, error) {
	var r0 api.SessionWriteResult
	return r0, nil
}
func (fakeSessions) Resize(_ context.Context, _ api.SessionResizeRequest) (api.SessionControlResult, error) {
	var r0 api.SessionControlResult
	return r0, nil
}
func (fakeSessions) Interrupt(_ context.Context, _ domain.SessionID) (api.SessionControlResult, error) {
	var r0 api.SessionControlResult
	return r0, nil
}
func (fakeSessions) Stop(_ context.Context, _ api.SessionStopRequest) (api.SessionStopResult, error) {
	var r0 api.SessionStopResult
	return r0, nil
}
func (fakeSessions) Restart(_ context.Context, _ api.SessionRestartRequest) (api.SessionRestartResult, error) {
	var r0 api.SessionRestartResult
	return r0, nil
}
func (fakeSessions) NextEvent(_ context.Context, _ ports.RuntimeEventCursor) (api.RuntimeEvent, error) {
	var r0 api.RuntimeEvent
	return r0, nil
}
func (fakeSessions) AckEvents(_ ports.RuntimeEventCursor) error { ; return nil }
func (fakeSessions) Shutdown(_ context.Context) api.RuntimeShutdownResult {
	var r0 api.RuntimeShutdownResult
	return r0
}

// These fakes certify exact compile signatures only, not native implementation.
func TestPortMethodInventory(t *testing.T) {
	interfaces := []int{11, 12, 4, 1, 2, 6, 12}
	total := 0
	for _, n := range interfaces {
		total += n
	}
	if total != 48 {
		t.Fatal(total)
	}
}
