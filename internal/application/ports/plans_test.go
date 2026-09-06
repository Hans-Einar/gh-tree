package ports_test

import (
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"strings"
	"testing"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func version(t string) api.SourceVersion {
	return must(api.NewSourceVersion("git", "repo", "native-issuer", t))
}
func op(n uint64) api.OperationID { return must(api.NewOperationID(n)) }
func fixtureExpected() api.GitExpectedState {
	repo := must(domain.NewRepositoryID(domain.LocalCommon, "repo"))
	w := must(domain.NewWorktreeID(repo, "primary"))
	r := must(domain.NewRevision(repo, must(domain.NewOID(strings.Repeat("1", 40)))))
	h := must(domain.NewDetachedHead(r))
	return must(api.NewGitExpectedState(api.GitExpectedStateData{Repository: repo, Worktree: api.Some(w), Observation: version("status"), Head: api.Some(h), Index: api.Some(version("index")), WorktreeState: api.Some(version("bytes")), Configuration: version("config"), Inventory: version("inventory")}))
}
func spec(kind api.GitMutationKind, token string) ports.PlanSpec {
	expected := fixtureExpected()
	summary := must(api.NewMutationPlanSummary(api.MutationPlanSummaryData{OperationID: op(1), Kind: kind, PlanVersion: version(token), Repository: expected.Data().Repository, Expected: expected, Choices: []api.Choice{api.Proceed, api.Cancel}, ConfirmationRequired: true}))
	return ports.PlanSpec{Operation: op(1), Version: version(token), Token: token, Group: "group", Role: ports.Executable, Summary: summary, SummaryDigest: [32]byte{1}}
}

type foreignPlan struct{ ports.PreparedGitPlan }

func (foreignPlan) Valid() bool { return true }

func TestPrivatePlanIdentityAndApproval(t *testing.T) {
	issuer := must(ports.NewPlanIssuer("Git-instance-1"))
	other := must(ports.NewPlanIssuer("Git-instance-2"))
	s := spec(api.StageMutation, "stage-1")
	plan := must(issuer.IssueStage(s))
	if e := issuer.ValidateBinding(plan, op(1), api.StageMutation, s.Version); e != nil {
		t.Fatal(e)
	}
	for _, e := range []error{other.ValidateBinding(plan, op(1), api.StageMutation, s.Version), issuer.ValidateBinding(plan, op(2), api.StageMutation, s.Version), issuer.ValidateBinding(plan, op(1), api.CommitMutation, s.Version), issuer.ValidateBinding(plan, op(1), api.StageMutation, version("different")), issuer.ValidateBinding(foreignPlan{}, op(1), api.StageMutation, s.Version)} {
		if e == nil {
			t.Fatal("foreign binding accepted")
		}
	}
	var nilPlan *ports.StagePlan
	if e := issuer.ValidateBinding(nilPlan, op(1), api.StageMutation, s.Version); e == nil {
		t.Fatal("typed nil plan")
	}
	approver := must(ports.NewApprovalIssuer("coordinator-1"))
	confirmation := api.Some(must(api.NewConfirmationID("consumed-once-by-coordinator")))
	if _, e := approver.Issue(plan, s.SummaryDigest, api.Proceed, api.None[api.ConfirmationID]()); e == nil {
		t.Fatal("missing required confirmation")
	}
	approval := must(approver.Issue(plan, s.SummaryDigest, api.Proceed, confirmation))
	if !approval.ValidFor(plan) || !approver.Issued(approval) {
		t.Fatal("valid approval refused")
	}
	if _, e := approver.Issue(plan, [32]byte{2}, api.Proceed, confirmation); e == nil {
		t.Fatal("summary digest mismatch")
	}
	if _, e := approver.Issue(plan, s.SummaryDigest, api.StashThenDeploy, confirmation); e == nil {
		t.Fatal("disallowed choice")
	}
	different := must(other.IssueStage(s))
	if approval.ValidFor(different) || ports.PlanEqual(plan, different) {
		t.Fatal("issuer lifetime lost")
	}
	summary := must(ports.PlanSummary(plan)).Data()
	summary.Choices[0] = api.StashThenDeploy
	if must(ports.PlanSummary(plan)).Data().Choices[0] != api.Proceed {
		t.Fatal("summary mutable")
	}
	wrong := s
	wrong.Summary = spec(api.CommitMutation, "stage-1").Summary
	if _, e := issuer.IssueStage(wrong); e == nil {
		t.Fatal("summary operation kind")
	}
}

func TestContinuationGroupStepAndReceipt(t *testing.T) {
	issuer := must(ports.NewPlanIssuer("git"))
	rootSpec := spec(api.CommitMutation, "root")
	rootSpec.Role = ports.SequenceRoot
	root := must(issuer.IssueCommit(rootSpec))
	approver := must(ports.NewApprovalIssuer("app"))
	approval := must(approver.Issue(root, rootSpec.SummaryDigest, api.Proceed, api.Some(must(api.NewConfirmationID("confirmation")))))
	if approval.ValidFor(root) {
		t.Fatal("sequence root executable")
	}
	childSpec := spec(api.StageMutation, "child")
	childSpec.Origin = api.Some[ports.PreparedGitPlan](root)
	childSpec.Step = 1
	data := childSpec.Summary.Data()
	data.OriginVersion = api.Some(rootSpec.Version)
	childSpec.Summary = must(api.NewMutationPlanSummary(data))
	child := must(issuer.IssueStage(childSpec))
	if !approval.ValidFor(child) {
		t.Fatal("original approval lost")
	}
	facet := must(api.NewFacetVersion(api.FacetVersionData{Facet: api.Index, Before: version("before"), After: version("after")}))
	versions := []api.FacetVersion{facet}
	receipt := must(issuer.IssueReceipt(child, "completed-step-record", versions, true, false))
	versions[0] = api.FacetVersion{}
	copy := receipt.Versions()
	copy[0] = api.FacetVersion{}
	if !receipt.Versions()[0].Valid() {
		t.Fatal("receipt version alias")
	}
	ctx := must(ports.NewGitPrepareContext(op(1), api.Some[ports.PreparedGitPlan](root), api.Some(receipt)))
	if !ctx.Valid() {
		t.Fatal("valid continuation")
	}
	for _, state := range []struct{ verified, canceled bool }{{false, false}, {true, true}} {
		r := must(issuer.IssueReceipt(child, "partial", []api.FacetVersion{facet}, state.verified, state.canceled))
		if _, e := ports.NewGitPrepareContext(op(1), api.Some[ports.PreparedGitPlan](root), api.Some(r)); e == nil {
			t.Fatal("failed/canceled predecessor continued")
		}
	}
	if _, e := ports.NewGitPrepareContext(op(2), api.Some[ports.PreparedGitPlan](root), api.Some(receipt)); e == nil {
		t.Fatal("cross-operation continuation")
	}
	if _, e := ports.NewGitPrepareContext(op(1), api.None[ports.PreparedGitPlan](), api.Some(receipt)); e == nil {
		t.Fatal("orphan predecessor")
	}
	wrong := childSpec
	wrong.Group = "foreign"
	if _, e := issuer.IssueStage(wrong); e == nil {
		t.Fatal("foreign group")
	}
	wrong = childSpec
	wrong.Step = 2
	if _, e := issuer.IssueStage(wrong); e == nil {
		t.Fatal("wrong step kind")
	}
}

func TestStorageWrappersBindFamilyRootAndLoadState(t *testing.T) {
	empty := must(api.NewJSONMembers(nil))
	doc := must(api.NewRunConfigDocument(api.RunConfigDocumentData{SchemaVersion: 1, UnknownMembers: empty}))
	w, _ := fixtureExpected().Data().Worktree.Value()
	root := must(api.NewDirectoryIdentity(api.DirectoryWindows, 1, [16]byte{1}, "stamp"))
	scope := must(api.NewWorktreeScope(api.WorktreeScopeData{ID: w, RootLocator: "C:/test", RootIdentity: root, Source: version("root")}))
	v := must(api.NewRunStorageVersion(scope, "run-store", false, 0, [32]byte{}))
	commit := must(ports.NewRunConfigCommit(scope, v, doc))
	if !commit.Valid() {
		t.Fatal("valid commit")
	}
	otherRoot := must(api.NewDirectoryIdentity(api.DirectoryWindows, 1, [16]byte{2}, "stamp"))
	changed := scope.Data()
	changed.RootIdentity = otherRoot
	foreign := must(api.NewWorktreeScope(changed))
	if _, e := ports.NewRunConfigCommit(foreign, v, doc); e == nil {
		t.Fatal("replaced root accepted")
	}
	userVersion := must(api.NewStorageVersion(api.UserConfig, "run-store", false, 0, [32]byte{}))
	if _, e := ports.NewRunConfigCommit(scope, userVersion, doc); e == nil {
		t.Fatal("foreign family")
	}
	obs := must(api.NewStorageLoadObservation(api.StorageLoadObservationData{State: api.LoadAbsent, Version: api.Some(v)}))
	if _, e := ports.NewLoadedRunConfig(scope, obs, api.Some(doc)); e != nil {
		t.Fatal(e)
	}
	if _, e := ports.NewLoadedRunConfig(foreign, obs, api.Some(doc)); e == nil {
		t.Fatal("load foreign root")
	}
	if _, e := ports.NewLoadedRunConfig(scope, obs, api.None[api.RunConfigDocument]()); e == nil {
		t.Fatal("usable load no document")
	}
	bad := must(api.NewStorageLoadObservation(api.StorageLoadObservationData{State: api.Corrupt, Version: api.Some(v)}))
	if _, e := ports.NewLoadedRunConfig(scope, bad, api.Some(doc)); e == nil {
		t.Fatal("corrupt load became defaults")
	}
	legacy := doc.Data()
	legacy.SchemaVersion = 0
	if _, e := ports.NewRunConfigCommit(scope, v, must(api.NewRunConfigDocument(legacy))); e == nil {
		t.Fatal("legacy writer")
	}
	present := must(api.NewRunStorageVersion(scope, "run-store", true, 32, [32]byte{1}))
	current := must(api.NewStorageLoadObservation(api.StorageLoadObservationData{State: api.ValidCurrent, Version: api.Some(present), SchemaVersion: api.Some(uint32(1))}))
	if _, e := ports.NewLoadedRunConfig(scope, current, api.Some(must(api.NewRunConfigDocument(legacy)))); e == nil {
		t.Fatal("legacy document mislabeled current")
	}
}
