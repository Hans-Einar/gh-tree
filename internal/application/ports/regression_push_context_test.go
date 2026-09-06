package ports_test

import (
	"errors"
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

func TestPushUnknownEvidenceAndSuppliedContext(t *testing.T) {
	expected := fixtureExpected()
	local := expected.Data().Repository
	w, _ := expected.Data().Worktree.Value()
	head, _ := expected.Data().Head.Value()
	source, _ := head.Revision()
	remoteA := must(d.NewRepositoryID(d.Remote, "remote-A"))
	remoteB := must(d.NewRepositoryID(d.Remote, "remote-B"))
	branchA := must(d.NewBranchID(remoteA, d.RemoteHead, "main"))
	branchB := must(d.NewBranchID(remoteB, d.RemoteHead, "main"))
	binding := func(r d.RepositoryID) a.RemoteBinding {
		return must(a.NewRemoteBinding(a.RemoteBindingData{LocalRepository: local, RemoteRepository: r, RemoteName: "origin", Configuration: version("remote-config")}))
	}
	bindingA, bindingB := binding(remoteA), binding(remoteB)
	record := must(a.NewRecoveryRecord(a.RecoveryRecordData{RecoveryID: must(a.NewRecoveryID("push-recovery")), Kind: a.RecoveryJournal, Layer: a.LayerGit, Subject: must(a.NewRecoverySubject(a.RecoverySubjectData{Repository: a.Some(remoteA), Branch: a.Some(branchA)})), Locator: "owned/push-journal", NextAction: "Inspect retained uncertainty"}))
	effect := must(a.NewEffectReport(a.EffectReportData{Facets: []a.FacetEffect{must(a.NewFacetEffect(a.FacetEffectData{Facet: a.RemoteRefsPR, State: a.EffectIndeterminate, RecoveryIDs: []a.RecoveryID{record.Data().RecoveryID}}))}}))
	empty := must(a.NewEffectReport(a.EffectReportData{}))
	reason := must(a.NewDiagnostic(a.DiagnosticData{Code: a.Indeterminate, Reason: "lost-response", Message: "Retain the original remote evidence."}))
	outcome := must(a.NewMutationIndeterminate(a.MutationIndeterminateData{Facts: must(a.NewGitPostFacts(a.GitPostFactsData{})), Reason: reason, Recovery: []a.RecoveryRecord{record}, ReconciliationRequired: true}))
	facts := must(a.NewGitMutationResult(a.GitMutationResultData{Operation: op(1), Kind: a.PushMutation, PlanVersion: version("push-plan"), Outcome: outcome, Effects: effect, Recovery: []a.RecoveryRecord{record}, ReconciliationRequired: true, Transport: must(a.NewCommandTransportOutcome(a.CommandTransportOutcomeData{}))}))
	if !facts.Valid() || facts.Data().Observation.Present() {
		t.Fatal("standalone uncertainty invented post-binding")
	}
	if e := facts.ValidatePushBinding(bindingA, branchA); e != nil {
		t.Fatal(e)
	}
	if e := facts.ValidatePushBinding(bindingB, branchB); e == nil {
		t.Fatal("known destination B adopted remote A recovery")
	}
	otherBranch := must(d.NewBranchID(remoteA, d.RemoteHead, "other"))
	if e := facts.ValidatePushBinding(bindingA, otherBranch); e == nil {
		t.Fatal("known different branch adopted recovery")
	}
	normalized := must(a.NormalizeRecovery([]a.RecoveryRecord{record}, nil))
	result := must(a.NewPushResult(a.PushResultData{Git: facts, RemoteEffect: effect, UpstreamEffect: empty, Outcome: must(a.NewOutcomeEnvelope(a.OutcomeEnvelopeData{Effects: effect, Recovery: normalized}))}))
	corr := must(a.NewCorrelation(a.CorrelationData{}))
	terminal := must(a.NewOperationTerminal(a.OperationTerminalData{OperationID: op(1), Correlation: corr, Disposition: a.Failed, Result: a.Some[a.Result](result), Effects: effect, Recovery: normalized}))
	request := func(b a.RemoteBinding, branch d.BranchID) a.Request {
		return must(a.NewCommandRequest(must(a.NewPushCommand(a.PushCommandData{WorktreeID: w, Source: source, Destination: branch, Binding: b, Expected: expected})), corr))
	}
	if e := a.ValidateTerminalFor(request(bindingA, branchA), terminal); e != nil {
		t.Fatal(e)
	}
	if e := a.ValidateTerminalFor(request(bindingB, branchB), terminal); e == nil {
		t.Fatal("original request mismatch accepted")
	}
	issuer := must(ports.NewPlanIssuer("push-issuer"))
	makePlan := func(b a.RemoteBinding, branch d.BranchID) ports.PushPlan {
		summary := must(a.NewMutationPlanSummary(a.MutationPlanSummaryData{OperationID: op(1), Kind: a.PushMutation, PlanVersion: version("push-plan"), Repository: local, Worktree: a.Some(w), Head: a.Some(head), Target: a.Some(must(d.NewCommitTarget(source))), Branch: a.Some(branch), PushBinding: a.Some(b), Expected: expected, Choices: []a.Choice{a.Proceed, a.Cancel}}))
		return must(issuer.IssuePush(ports.PlanSpec{Operation: op(1), Version: version("push-plan"), Token: branch.Repository().Token(), Group: "group", Role: ports.Executable, Summary: summary, SummaryDigest: [32]byte{1}}))
	}
	receiptA := must(issuer.IssueReceipt(makePlan(bindingA, branchA), "receipt-a", nil, false, false))
	receiptB := must(issuer.IssueReceipt(makePlan(bindingB, branchB), "receipt-b", nil, false, false))
	wrapped := must(ports.NewExecutedGitMutation(facts, a.Some(receiptA)))
	if _, e := ports.NewExecutedGitMutation(facts, a.Some(receiptB)); e == nil {
		t.Fatal("receipt summary mismatch accepted")
	}
	call := func() (ports.ExecutedGitMutation, error) { return wrapped, errors.New("response unknown") }
	retained, e := call()
	if e == nil || retained.Facts().Data().Recovery[0].Data().RecoveryID != record.Data().RecoveryID {
		t.Fatal("validation/error handling erased recovery")
	}
	explicit := must(a.NewPushed(a.PushedData{Source: source, Destination: branchB, Binding: bindingB, RemoteEffect: effect, UpstreamEffect: empty}))
	bad := facts.Data()
	bad.Outcome = explicit
	if _, e := a.NewGitMutationResult(bad); e == nil {
		t.Fatal("returned typed binding B adopted remote A")
	}
	noop := must(a.NewEffectReport(a.EffectReportData{Facets: []a.FacetEffect{must(a.NewFacetEffect(a.FacetEffectData{Facet: a.RemoteRefsPR, State: a.VerifiedNoTargetChange}))}}))
	success := must(a.NewPushed(a.PushedData{Source: source, Destination: branchA, Binding: bindingA, RemoteEffect: noop, UpstreamEffect: empty}))
	good := facts.Data()
	good.Outcome = success
	good.Effects = noop
	good.Recovery = nil
	if _, e := a.NewGitMutationResult(good); e != nil {
		t.Fatal("already-up-to-date push refused", e)
	}
}
