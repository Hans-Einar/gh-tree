package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"strings"
	"testing"
	"time"
)

func rvMust[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func rvSource(s string) a.SourceVersion {
	return rvMust(a.NewSourceVersion("git", "scope", "reviewer", s))
}
func rvRev(r d.RepositoryID, s string) d.Revision {
	return rvMust(d.NewRevision(r, rvMust(d.NewOID(strings.Repeat(s, 40)))))
}
func rvRepo(s string) d.RepositoryID                 { return rvMust(d.NewRepositoryID(d.LocalCommon, s)) }
func rvWork(r d.RepositoryID, s string) d.WorktreeID { return rvMust(d.NewWorktreeID(r, s)) }
func rvScope(w d.WorktreeID) a.WorktreeScope {
	return rvMust(a.NewWorktreeScope(a.WorktreeScopeData{ID: w, RootLocator: "C:/review/" + w.AdministrativeKey(), RootIdentity: rvMust(a.NewDirectoryIdentity(a.DirectoryWindows, 1, [16]byte{1}, "stamp")), Source: rvSource("root")}))
}
func rvObservation(w d.WorktreeID) a.GitObservation {
	at := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	return rvMust(a.NewGitObservation(a.GitObservationData{ID: rvMust(a.NewObservationID(w.Repository().Token() + w.AdministrativeKey())), Repository: w.Repository(), Worktree: a.Some(w), Interval: rvMust(a.NewObservationInterval(a.ObservationIntervalData{StartedAt: at, FinishedAt: at})), Version: rvSource("observation"), Completeness: a.Complete}))
}
func rvFacts(w d.WorktreeID) a.WorktreeFacts {
	return rvMust(a.NewWorktreeFacts(a.WorktreeFactsData{ID: w, Scope: a.Some(rvScope(w)), Head: a.Some(rvMust(d.NewDetachedHead(rvRev(w.Repository(), "1")))), Availability: rvMust(a.NewAvailableWorktree(a.AvailableWorktreeData{})), Observation: rvObservation(w)}))
}
func rvStatus(w d.WorktreeID) a.StatusFacts {
	return rvMust(a.NewStatusFacts(a.StatusFactsData{Worktree: rvFacts(w), IndexVersion: rvSource("index"), WorktreeVersion: rvSource("bytes"), ConfigurationVersion: rvSource("config"), Upstream: rvMust(a.NewNoUpstream(a.NoUpstreamData{})), Observation: rvObservation(w)}))
}
func rvExpected(w d.WorktreeID) a.GitExpectedState {
	return rvMust(a.NewGitExpectedState(a.GitExpectedStateData{Repository: w.Repository(), Worktree: a.Some(w), Head: a.Some(rvMust(d.NewDetachedHead(rvRev(w.Repository(), "1")))), Index: a.Some(rvSource("index")), WorktreeState: a.Some(rvSource("bytes")), Observation: rvSource("status"), Configuration: rvSource("config"), Inventory: rvSource("inventory")}))
}
func rvEffects() a.EffectReport { return rvMust(a.NewEffectReport(a.EffectReportData{})) }
func rvEffect(f a.EffectFacet, s a.EffectState) a.EffectReport {
	return rvMust(a.NewEffectReport(a.EffectReportData{Facets: []a.FacetEffect{rvMust(a.NewFacetEffect(a.FacetEffectData{Facet: f, State: s}))}}))
}
func rvEnvelope(e a.EffectReport) a.OutcomeEnvelope {
	return rvMust(a.NewOutcomeEnvelope(a.OutcomeEnvelopeData{Effects: e}))
}
func rvTransport() a.CommandTransportOutcome {
	return rvMust(a.NewCommandTransportOutcome(a.CommandTransportOutcomeData{}))
}
func rvDiag() a.Diagnostic {
	return rvMust(a.NewDiagnostic(a.DiagnosticData{Code: a.Conflict, Reason: "review", Message: "review"}))
}
func rvReject(t *testing.T, e error) {
	t.Helper()
	if e == nil {
		t.Fatal("contradictory value was admitted as valid")
	}
}
func rvMutation(out a.GitMutationOutcome, kind a.GitMutationKind, e a.EffectReport) a.GitMutationResult {
	return rvMust(a.NewGitMutationResult(a.GitMutationResultData{Operation: rvMust(a.NewOperationID(1)), Kind: kind, PlanVersion: rvSource("plan"), Outcome: out, Transport: rvTransport(), Effects: e}))
}
func rvRunVersion(w d.WorktreeID) a.StorageVersion {
	return rvMust(a.NewRunStorageVersion(rvScope(w), "run-store", true, 20, [32]byte{1}))
}
func rvSelection(w d.WorktreeID) a.DiscoveredLaunch {
	return rvMust(a.NewDiscoveredLaunch(a.DiscoveredLaunchData{Member: rvMust(a.NewMemberSelection(a.MemberSelectionData{LaunchPointID: rvMust(d.NewLaunchPointID(w, "npm", "", "dev")), SourceVersion: rvSource("manifest")}))}))
}

func TestReviewGitCrossFieldScope(t *testing.T) {
	ra, rb := rvRepo("A"), rvRepo("B")
	wa, wb := rvWork(ra, "one"), rvWork(rb, "two")
	rev := rvRev(ra, "1")
	branch := rvMust(d.NewBranchID(ra, d.Local, "main"))
	ref := rvMust(a.NewLocalBranchRef(a.LocalBranchRefData{Branch: branch}))
	t.Run("positive_ref", func(t *testing.T) {
		if _, e := a.NewRefFact(a.RefFactData{Locator: ref, Revision: a.Some(rev), Observation: rvObservation(wa)}); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("foreign_ref_revision", func(t *testing.T) {
		_, e := a.NewRefFact(a.RefFactData{Locator: ref, Revision: a.Some(rvRev(rb, "1")), Observation: rvObservation(wa)})
		rvReject(t, e)
	})
	t.Run("same_repo_wrong_status_worktree", func(t *testing.T) {
		x := rvStatus(wa).Data()
		x.Observation = rvObservation(rvWork(ra, "other"))
		_, e := a.NewStatusFacts(x)
		rvReject(t, e)
	})
	target := rvMust(d.NewCommitTarget(rev))
	resolved := rvMust(a.NewExactLocalResolution(a.ExactLocalResolutionData{Requested: target, Local: rev, Observation: rvObservation(wa)}))
	t.Run("created_worktree_different_target_repo", func(t *testing.T) {
		_, e := a.NewWorktreeCreated(a.WorktreeCreatedData{Worktree: rvFacts(wb), Target: resolved})
		rvReject(t, e)
	})
	t.Run("created_worktree_no_head_or_scope", func(t *testing.T) {
		missing := rvMust(a.NewWorktreeFacts(a.WorktreeFactsData{ID: wa, Availability: rvMust(a.NewMissingWorktree(a.MissingWorktreeData{})), Observation: rvObservation(wa)}))
		_, e := a.NewWorktreeCreated(a.WorktreeCreatedData{Worktree: missing, Target: resolved})
		rvReject(t, e)
	})
	t.Run("stash_applied_other_repository", func(t *testing.T) {
		stash := rvMust(d.NewStashID(ra, rev.OID()))
		_, e := a.NewStashApplied(a.StashAppliedData{Stash: stash, Status: rvStatus(wb), IndexRestored: true, Retained: true})
		rvReject(t, e)
	})
}

func TestReviewRequestVersionBindings(t *testing.T) {
	wa, wb := rvWork(rvRepo("A"), "one"), rvWork(rvRepo("B"), "two")
	geom := rvMust(a.NewGeometry(a.GeometryData{Rows: 24, Columns: 80}))
	sel := rvSelection(wa)
	t.Run("positive_saved", func(t *testing.T) {
		_, e := a.NewSavedLaunch(a.SavedLaunchData{Alias: "alias", LaunchPointID: sel.Data().Member.Data().LaunchPointID, StorageVersion: rvRunVersion(wa), SourceExpectation: rvSource("manifest")})
		if e != nil {
			t.Fatal(e)
		}
	})
	t.Run("saved_selection_foreign_run_version", func(t *testing.T) {
		_, e := a.NewSavedLaunch(a.SavedLaunchData{Alias: "alias", LaunchPointID: sel.Data().Member.Data().LaunchPointID, StorageVersion: rvRunVersion(wb), SourceExpectation: rvSource("manifest")})
		rvReject(t, e)
	})
	t.Run("discovery_foreign_version", func(t *testing.T) {
		_, e := a.NewDiscoveryRequest(a.DiscoveryRequestData{Worktree: rvScope(wa), SavedVersion: a.Some(rvRunVersion(wb))})
		rvReject(t, e)
	})
	t.Run("resolve_foreign_version", func(t *testing.T) {
		_, e := a.NewResolveLaunchRequest(a.ResolveLaunchRequestData{Worktree: rvScope(wa), Selection: sel, SavedVersion: a.Some(rvRunVersion(wb)), Geometry: geom})
		rvReject(t, e)
	})
	t.Run("save_foreign_version", func(t *testing.T) {
		_, e := a.NewSaveLaunchCommand(a.SaveLaunchCommandData{WorktreeID: wa, Selection: sel, Alias: "alias", ExpectedStorage: rvRunVersion(wb)})
		rvReject(t, e)
	})
	t.Run("default_foreign_version", func(t *testing.T) {
		choice := rvMust(a.NewCurrentDefaultLaunch(a.CurrentDefaultLaunchData{ExpectedStorage: a.Some(rvRunVersion(wb))}))
		_, e := a.NewStartLaunchCommand(a.StartLaunchCommandData{WorktreeID: wa, Selection: choice, Geometry: geom})
		rvReject(t, e)
	})
	t.Run("retarget_exact_target_foreign_local_repository", func(t *testing.T) {
		_, e := a.NewRetargetWorktreeCommand(a.RetargetWorktreeCommandData{WorktreeID: wa, Target: rvMust(d.NewCommitTarget(rvRev(wb.Repository(), "1"))), Mode: rvMust(a.NewDetachRetarget(a.DetachRetargetData{})), Expected: rvExpected(wa)})
		rvReject(t, e)
	})
}

func TestReviewPublicResultSemantics(t *testing.T) {
	wa, wb := rvWork(rvRepo("A"), "one"), rvWork(rvRepo("B"), "two")
	sources := rvMust(a.NewProjectionSources(a.ProjectionSourcesData{}))
	rev := rvRev(wa.Repository(), "1")
	target := rvMust(d.NewCommitTarget(rev))
	t.Run("worktree_status_foreign_subject", func(t *testing.T) {
		_, e := a.NewWorktreeStatusResult(a.WorktreeStatusResultData{WorktreeID: wa, Status: a.Some(rvStatus(wb)), Sources: sources})
		rvReject(t, e)
	})
	t.Run("PR_result_accepts_commit_target", func(t *testing.T) {
		_, e := a.NewPullRequestDiffResult(a.PullRequestDiffResultData{Target: target, RequestedBase: rvMust(a.NewObserveCurrentPRBase(a.ObserveCurrentPRBaseData{})), Sources: sources})
		rvReject(t, e)
	})
	t.Run("deploy_resolution_replaces_exact_target", func(t *testing.T) {
		r2 := rvRev(wa.Repository(), "2")
		res := rvMust(a.NewExactLocalResolution(a.ExactLocalResolutionData{Requested: rvMust(d.NewCommitTarget(r2)), Local: r2, Observation: rvObservation(wa)}))
		_, e := a.NewDeployResult(a.DeployResultData{Target: target, Resolution: a.Some(res), Outcome: rvEnvelope(rvEffects())})
		rvReject(t, e)
	})
	refused := rvMutation(rvMust(a.NewRefusedMutation(a.RefusedMutationData{Reason: rvDiag(), Effects: rvEffects()})), a.PushMutation, rvEffects())
	t.Run("completed_pop_wraps_refused_pushes", func(t *testing.T) {
		_, e := a.NewStashPopCompleted(a.StashPopCompletedData{Stash: rvMust(d.NewStashID(wa.Repository(), rev.OID())), Apply: refused, Drop: refused})
		rvReject(t, e)
	})
	t.Run("retarget_confirmation_omits_exact_target_and_worktree", func(t *testing.T) {
		_, e := a.NewMutationPlanSummary(a.MutationPlanSummaryData{OperationID: rvMust(a.NewOperationID(1)), Kind: a.RetargetMutation, PlanVersion: rvSource("plan"), Repository: wa.Repository(), Expected: rvExpected(wa), Choices: []a.Choice{a.Proceed}, ConfirmationRequired: true})
		rvReject(t, e)
	})
}

func TestReviewEffectNormalization(t *testing.T) {
	wa := rvWork(rvRepo("A"), "one")
	effect := rvEffect(a.Index, a.AppliedVerified)
	indexed := rvMust(a.NewIndexChanged(a.IndexChangedData{Status: rvStatus(wa), Action: a.Stage}))
	facts := rvMutation(indexed, a.StageMutation, effect)
	t.Run("positive_stage", func(t *testing.T) {
		_, e := a.NewStageAllResult(a.StageAllResultData{Git: facts, Outcome: rvEnvelope(effect)})
		if e != nil {
			t.Fatal(e)
		}
	})
	t.Run("command_envelope_discards_known_effect", func(t *testing.T) {
		_, e := a.NewStageAllResult(a.StageAllResultData{Git: facts, Outcome: rvEnvelope(rvEffects())})
		rvReject(t, e)
	})
	stage := rvMust(a.NewStageAllResult(a.StageAllResultData{Git: facts, Outcome: rvEnvelope(effect)}))
	t.Run("terminal_says_no_change_after_applied_index", func(t *testing.T) {
		_, e := a.NewOperationTerminal(a.OperationTerminalData{OperationID: rvMust(a.NewOperationID(1)), Correlation: rvMust(a.NewCorrelation(a.CorrelationData{})), Disposition: a.Succeeded, Result: a.Some[a.Result](stage), Effects: rvEffect(a.Index, a.NotStarted)})
		rvReject(t, e)
	})
}

func TestReviewStorageRecoveryScope(t *testing.T) {
	wa, wb := rvWork(rvRepo("A"), "one"), rvWork(rvRepo("B"), "two")
	version := rvRunVersion(wb)
	subject := rvMust(a.NewRecoverySubject(a.RecoverySubjectData{Worktree: a.Some(wa), Store: a.Some(version.Store()), Family: a.Some(a.RunConfig)}))
	// The corrected shared record refuses this contradiction before a family wrapper.
	data := a.RecoveryRecordData{RecoveryID: rvMust(a.NewRecoveryID("artifact")), Kind: a.RecoveryManifest, Layer: a.LayerPersistence, Subject: subject, Locator: "retained", Original: a.Some[a.RecoveryVersion](rvMust(a.NewStorageRecoveryVersion(a.StorageRecoveryVersionData{Version: version}))), NextAction: "inspect"}
	t.Run("artifact_subject_and_version_different_worktree", func(t *testing.T) {
		_, e := a.NewRecoveryRecord(data)
		rvReject(t, e)
	})
}
