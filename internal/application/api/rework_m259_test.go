package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

func TestM259CompoundEffectsKeepAllStages(t *testing.T) {
	w := rvWork(rvRepo("compound"), "worktree")
	staged := rvEffect(a.Index, a.AppliedVerified)
	untouched := rvEffect(a.Index, a.NotStarted)
	stage := rvMutation(rvMust(a.NewIndexChanged(a.IndexChangedData{Status: rvStatus(w), Action: a.Stage})), a.StageMutation, staged)
	commit := rvMutation(rvMust(a.NewRefusedMutation(a.RefusedMutationData{Reason: rvDiag(), Effects: untouched})), a.CommitMutation, untouched)
	merged := rvMust(a.MergeEffectReports(staged, untouched, staged))
	if len(merged.Data().Facets) != 2 {
		t.Fatal("compound stages collapsed into ordinal state")
	}
	result := rvMust(a.NewStageAllAndCommitResult(a.StageAllAndCommitResultData{Stage: a.Some(stage), Commit: a.Some(commit), Outcome: rvEnvelope(merged)}))
	if _, e := a.NewOperationTerminal(a.OperationTerminalData{OperationID: rvMust(a.NewOperationID(1)), Correlation: rvMust(a.NewCorrelation(a.CorrelationData{})), Disposition: a.Failed, Result: a.Some[a.Result](result), Effects: merged, CancellationRequested: true}); e != nil {
		t.Fatal(e)
	}
	bad := result.Data()
	bad.Outcome = rvEnvelope(staged)
	if _, e := a.NewStageAllAndCommitResult(bad); e == nil {
		t.Fatal("later untouched stage erased")
	}
	if _, e := a.NewOperationTerminal(a.OperationTerminalData{OperationID: rvMust(a.NewOperationID(2)), Correlation: rvMust(a.NewCorrelation(a.CorrelationData{})), Disposition: a.Failed, Result: a.Some[a.Result](result), Effects: merged}); e == nil {
		t.Fatal("terminal substituted operation")
	}
}

func TestM259StashCompoundKindsAndPartialInventory(t *testing.T) {
	w := rvWork(rvRepo("stash"), "worktree")
	stash := rvMust(d.NewStashID(w.Repository(), rvRev(w.Repository(), "1").OID()))
	applyEffects := rvEffect(a.Index, a.AppliedVerified)
	dropEffects := rvEffect(a.LocalRefsHead, a.AppliedVerified)
	apply := rvMutation(rvMust(a.NewStashApplied(a.StashAppliedData{Stash: stash, Status: rvStatus(w), IndexRestored: true, Retained: true})), a.StashMutation, applyEffects)
	drop := rvMutation(rvMust(a.NewStashDropped(a.StashDroppedData{Stash: stash, Occurrence: rvSource("occurrence"), Observation: rvObservation(w), RefCleanup: dropEffects})), a.StashMutation, dropEffects)
	complete := rvMust(a.NewStashPopCompleted(a.StashPopCompletedData{Stash: stash, Apply: apply, Drop: drop}))
	if _, e := a.NewStashPopResult(a.StashPopResultData{Outcome: complete, Envelope: rvEnvelope(rvMust(a.MergeEffectReports(applyEffects, dropEffects)))}); e != nil {
		t.Fatal(e)
	}
	if _, e := a.NewStashDropResult(a.StashDropResultData{Stash: stash, Git: apply, Outcome: rvEnvelope(applyEffects)}); e == nil {
		t.Fatal("drop wrapper accepted apply")
	}
	refused := rvMutation(rvMust(a.NewRefusedMutation(a.RefusedMutationData{Reason: rvDiag(), Effects: rvEffects()})), a.StashMutation, rvEffects())
	if _, e := a.NewAppliedStashRetained(a.AppliedStashRetainedData{Stash: stash, Apply: apply, Drop: a.Some(refused), Reason: rvDiag()}); e != nil {
		t.Fatal("known apply then refused drop", e)
	}
	if _, e := a.NewAppliedStashRetained(a.AppliedStashRetainedData{Stash: stash, Apply: apply, Drop: a.Some(drop), Reason: rvDiag()}); e == nil {
		t.Fatal("known dropped stash retained")
	}
	missing := rvMust(a.NewWorktreeFacts(a.WorktreeFactsData{ID: w, Availability: rvMust(a.NewMissingWorktree(a.MissingWorktreeData{})), Observation: rvObservation(w)}))
	partial := rvMust(a.NewPartialMutation(a.PartialMutationData{Facts: rvMust(a.NewGitPostFacts(a.GitPostFactsData{Worktrees: []a.WorktreeFacts{missing}})), Reason: rvDiag(), ReconciliationRequired: true}))
	if _, e := a.NewGitMutationResult(a.GitMutationResultData{Operation: rvMust(a.NewOperationID(1)), Kind: a.CreateMutation, PlanVersion: rvSource("plan"), Outcome: partial, Observation: a.Some(rvObservation(w)), Transport: rvTransport(), Effects: rvEffects(), ReconciliationRequired: true}); e != nil {
		t.Fatal("partial creation lost unavailable worktree", e)
	}
}

func TestM259PlanSummaryCompleteKindContracts(t *testing.T) {
	w := rvWork(rvRepo("plans"), "worktree")
	expected := rvExpected(w)
	revision := rvRev(w.Repository(), "1")
	target := rvMust(d.NewCommitTarget(revision))
	branch := rvMust(d.NewBranchID(w.Repository(), d.Local, "new"))
	remote := rvMust(d.NewRepositoryID(d.Remote, "remote"))
	remoteBranch := rvMust(d.NewBranchID(remote, d.RemoteHead, "main"))
	binding := rvMust(a.NewRemoteBinding(a.RemoteBindingData{LocalRepository: w.Repository(), RemoteRepository: remote, RemoteName: "origin", Configuration: rvSource("remote")}))
	path := rvMust(a.NewGitPath("literal"))
	absent := rvMust(a.NewAbsentFile(a.AbsentFileData{Path: path}))
	planned := rvMust(a.NewPlannedPathEffect(a.PlannedPathEffectData{Old: absent, Desired: absent, SourceVersion: rvSource("path")}))
	for _, kind := range []a.GitMutationKind{a.CreateMutation, a.RetargetMutation, a.StageMutation, a.CommitMutation, a.RestoreMutation, a.StashMutation, a.BranchMutation, a.PushMutation} {
		t.Run(string(rune('0'+kind)), func(t *testing.T) {
			s := a.MutationPlanSummaryData{OperationID: rvMust(a.NewOperationID(1)), Kind: kind, PlanVersion: rvSource("plan"), Repository: w.Repository(), Worktree: a.Some(w), Head: expected.Data().Head, Expected: expected, Choices: []a.Choice{a.Proceed, a.Cancel}, ConfirmationRequired: true}
			switch kind {
			case a.CreateMutation:
				s.Target = a.Some(target)
				s.Destination = a.Some("C:/vacant")
				s.CreateMode = a.Some[a.CreateMode](rvMust(a.NewDetachedCreate(a.DetachedCreateData{})))
			case a.RetargetMutation:
				s.Target = a.Some(target)
				s.RetargetMode = a.Some[a.RetargetMode](rvMust(a.NewDetachRetarget(a.DetachRetargetData{})))
			case a.StageMutation:
				s.StageAction = a.Some(a.Stage)
			case a.CommitMutation:
				s.Message = a.Some("literal message")
				s.CommitIndexPolicy = a.Some(a.ExistingIndex)
			case a.RestoreMutation:
				s.Paths = []a.PlannedPathEffect{planned}
			case a.StashMutation:
				s.StashIntent = a.Some[a.StashIntent](rvMust(a.NewCreateStashIntent(a.CreateStashIntentData{Worktree: w})))
			case a.BranchMutation:
				s.Target = a.Some(target)
				s.Branch = a.Some(branch)
			case a.PushMutation:
				s.Target = a.Some(target)
				s.Branch = a.Some(remoteBranch)
				s.PushBinding = a.Some(binding)
			}
			if _, e := a.NewMutationPlanSummary(s); e != nil {
				t.Fatal("complete summary refused", e)
			}
			if kind != a.CreateMutation {
				s.Worktree = a.None[d.WorktreeID]()
				if _, e := a.NewMutationPlanSummary(s); e == nil {
					t.Fatal("missing material worktree")
				}
			}
		})
	}
}

func TestM259PartialDeployAndSavedRootControls(t *testing.T) {
	w := rvWork(rvRepo("deployed"), "worktree")
	target := rvMust(d.NewCommitTarget(rvRev(w.Repository(), "2")))
	resolution := rvMust(a.NewExactLocalResolution(a.ExactLocalResolutionData{Requested: target, Local: target.ExpectedRevision(), Observation: rvObservation(w)}))
	old := rvMust(d.NewDetachedHead(rvRev(w.Repository(), "1")))
	refusal := rvMutation(rvMust(a.NewRefusedMutation(a.RefusedMutationData{Reason: rvDiag(), Effects: rvEffects()})), a.RetargetMutation, rvEffects())
	if _, e := a.NewDeployResult(a.DeployResultData{Target: target, Resolution: a.Some(resolution), Head: a.Some(old), Steps: []a.GitMutationResult{refusal}, Outcome: rvEnvelope(rvEffects())}); e != nil {
		t.Fatal("failed deploy must retain observed old head", e)
	}
	root := rvScope(w).Data()
	root.RootIdentity = rvMust(a.NewDirectoryIdentity(a.DirectoryWindows, 1, [16]byte{2}, "stamp"))
	foreignRoot := rvMust(a.NewWorktreeScope(root))
	v := rvMust(a.NewRunStorageVersion(foreignRoot, "run-store", true, 20, [32]byte{1}))
	if _, e := a.NewDiscoveryRequest(a.DiscoveryRequestData{Worktree: rvScope(w), SavedVersion: a.Some(v)}); e == nil {
		t.Fatal("same worktree replaced root")
	}
}

func TestM259RequestResultAndRemoteEvidenceControls(t *testing.T) {
	correlation := rvMust(a.NewCorrelation(a.CorrelationData{}))
	one := rvMust(d.NewSessionID(1))
	two := rvMust(d.NewSessionID(2))
	request := rvMust(a.NewCommandRequest(rvMust(a.NewWriteInputCommand(a.WriteInputCommandData{SessionID: one, Bytes: []byte{1, 2, 3}})), correlation))
	write := rvMust(a.NewSessionWriteResult(a.SessionWriteResultData{SessionID: two, Sequence: rvMust(a.NewSessionSequence(1)), AcceptedBytes: 3}))
	result := rvMust(a.NewWriteInputResult(a.WriteInputResultData{Write: write, Outcome: rvEnvelope(rvEffects())}))
	terminal := rvMust(a.NewOperationTerminal(a.OperationTerminalData{OperationID: rvMust(a.NewOperationID(1)), Correlation: correlation, Disposition: a.Succeeded, Result: a.Some[a.Result](result), Effects: rvEffects()}))
	if e := a.ValidateTerminalFor(request, terminal); e == nil {
		t.Fatal("same result type substituted session")
	}
	wd := write.Data()
	wd.SessionID = one
	wd.AcceptedBytes = 2
	rd := result.Data()
	rd.Write = rvMust(a.NewSessionWriteResult(wd))
	td := terminal.Data()
	td.Result = a.Some[a.Result](rvMust(a.NewWriteInputResult(rd)))
	if e := a.ValidateTerminalFor(request, rvMust(a.NewOperationTerminal(td))); e == nil {
		t.Fatal("partial queue admission accepted")
	}
	baseRepo := rvMust(d.NewRepositoryID(d.Remote, "base"))
	forkRepo := rvMust(d.NewRepositoryID(d.Remote, "fork"))
	source := rvSource("remote")
	obs := rvMust(a.NewRemoteObservation(a.RemoteObservationData{ID: rvMust(a.NewObservationID("base")), Repository: baseRepo, Interval: rvObservation(rvWork(rvRepo("local"), "w")).Data().Interval, Version: source, Origin: rvMust(a.NewLiveRemoteObservation(a.LiveRemoteObservationData{})), Page: rvMust(a.NewPageInfo(a.PageInfoData{Source: source, Completeness: a.Complete}))}))
	base := rvMust(a.NewUnavailableEndpoint(a.UnavailableEndpointData{KnownRevision: a.Some(rvRev(forkRepo, "1")), Reason: a.EndpointInaccessible}))
	head := rvMust(a.NewUnavailableEndpoint(a.UnavailableEndpointData{KnownRepository: a.Some(forkRepo), Reason: a.EndpointDeleted}))
	if _, e := a.NewPullRequestFact(a.PullRequestFactData{ID: rvMust(d.NewPRID(baseRepo, 1)), URL: "https://example.test/base/repo/pull/1", State: a.PROpen, Base: base, Head: head, Observation: obs}); e == nil {
		t.Fatal("unavailable base revision conflicts with PR scope")
	}
	bd := base.Data()
	bd.KnownRevision = a.Some(rvRev(baseRepo, "1"))
	if _, e := a.NewPullRequestFact(a.PullRequestFactData{ID: rvMust(d.NewPRID(baseRepo, 1)), URL: "https://example.test/base/repo/pull/1", State: a.PROpen, Base: rvMust(a.NewUnavailableEndpoint(bd)), Head: head, Observation: obs}); e != nil {
		t.Fatal("independent unavailable fork retained", e)
	}
}
