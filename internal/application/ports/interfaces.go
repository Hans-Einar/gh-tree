// Package ports owns the frozen dependency interfaces and immutable issued handles.
package ports

import (
	"context"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type SourceVersion = api.SourceVersion
type StorageVersion = api.StorageVersion
type RuntimeEventCursor = api.RuntimeEventSequence
type GitFacts interface {
	ResolveLocal(context.Context, api.ResolveLocalRequest) (api.ResolveLocalResult, error)
	ListWorktrees(context.Context, api.ListWorktreesRequest) (api.ListWorktreesResult, error)
	ObserveStatus(context.Context, api.ObserveStatusRequest) (api.ObserveStatusResult, error)
	ResolveExact(context.Context, api.ResolveExactRequest) (api.ResolveExactResult, error)
	ListRefs(context.Context, api.ListRefsRequest) (api.ListRefsResult, error)
	ListStashes(context.Context, api.ListStashesRequest) (api.ListStashesResult, error)
	ReadStashPatch(context.Context, api.ReadStashPatchRequest) (api.ReadStashPatchResult, error)
	MergeBase(context.Context, api.MergeBaseRequest) (api.MergeBaseResult, error)
	ReadCommits(context.Context, api.ReadCommitsRequest) (api.ReadCommitsResult, error)
	ReadGraph(context.Context, api.ReadGraphRequest) (api.ReadGraphResult, error)
	ReadDiff(context.Context, api.ReadDiffRequest) (api.ReadDiffResult, error)
}

type GitMutations interface {
	PrepareCreate(context.Context, GitPrepareContext, api.PrepareCreateRequest) (CreatePlan, api.GitPreparationResult, error)
	PrepareRetarget(context.Context, GitPrepareContext, api.PrepareRetargetRequest) (RetargetPlan, api.GitPreparationResult, error)
	PrepareStage(context.Context, GitPrepareContext, api.PrepareStageRequest) (StagePlan, api.GitPreparationResult, error)
	PrepareCommit(context.Context, GitPrepareContext, api.PrepareCommitRequest) (CommitPlan, api.GitPreparationResult, error)
	PrepareRestore(context.Context, GitPrepareContext, api.PrepareRestoreRequest) (RestorePlan, api.GitPreparationResult, error)
	PrepareStash(context.Context, GitPrepareContext, api.PrepareStashRequest) (StashPlan, api.GitPreparationResult, error)
	PrepareBranch(context.Context, GitPrepareContext, api.PrepareBranchRequest) (BranchPlan, api.GitPreparationResult, error)
	PreparePush(context.Context, GitPrepareContext, api.PreparePushRequest) (PushPlan, api.GitPreparationResult, error)
	ExecutePrepared(context.Context, PreparedGitPlan, ExecutionApproval) (ExecutedGitMutation, error)
	ReleasePlan(PreparedGitPlan) error
	Fetch(context.Context, api.FetchRequest) (api.FetchResult, error)
	Reconcile(context.Context, api.ReconcileRequest) (api.ReconcileResult, error)
}

type RemoteFacts interface {
	ResolveRepository(context.Context, api.ResolveRepositoryRequest) (api.ResolveRepositoryResult, error)
	ListBranches(context.Context, api.ListBranchesRequest) (api.ListBranchesResult, error)
	ListPullRequests(context.Context, api.ListPullRequestsRequest) (api.ListPullRequestsResult, error)
	ObservePullRequest(context.Context, api.ObservePullRequestRequest) (api.ObservePullRequestResult, error)
}

type RemoteMutations interface {
	CreatePullRequest(context.Context, api.CreatePullRequestRequest) (api.CreatePullRequestResult, error)
}

type LaunchDiscovery interface {
	Discover(context.Context, api.DiscoveryRequest) (api.DiscoveryResult, error)
	Resolve(context.Context, api.ResolveLaunchRequest) (api.ResolveLaunchResult, error)
}

type Storage interface {
	LoadUserConfig(context.Context) (LoadedUserConfig, error)
	LoadPreferences(context.Context) (LoadedPreferences, error)
	LoadRunConfig(context.Context, api.WorktreeScope) (LoadedRunConfig, error)
	CommitUserConfig(context.Context, UserConfigCommit) (api.StorageCommitResult, error)
	CommitPreferences(context.Context, PreferencesCommit) (api.StorageCommitResult, error)
	CommitRunConfig(context.Context, RunConfigCommit) (api.StorageCommitResult, error)
}

type Sessions interface {
	Start(context.Context, api.SessionStartRequest) (api.SessionStartResult, error)
	Snapshot(context.Context, domain.SessionID) (api.SessionSnapshot, error)
	List(context.Context, api.SessionFilter) (api.SessionList, error)
	ReadOutput(context.Context, api.SessionOutputRequest) (api.SessionOutputResult, error)
	Write(context.Context, api.SessionWriteRequest) (api.SessionWriteResult, error)
	Resize(context.Context, api.SessionResizeRequest) (api.SessionControlResult, error)
	Interrupt(context.Context, domain.SessionID) (api.SessionControlResult, error)
	Stop(context.Context, api.SessionStopRequest) (api.SessionStopResult, error)
	Restart(context.Context, api.SessionRestartRequest) (api.SessionRestartResult, error)
	NextEvent(context.Context, RuntimeEventCursor) (api.RuntimeEvent, error)
	AckEvents(RuntimeEventCursor) error
	Shutdown(context.Context) api.RuntimeShutdownResult
}
