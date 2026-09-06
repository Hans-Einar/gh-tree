package git

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// These are pure portable registry controls. Supplied semantic identities are
// fixture values, not assertions of native repository/preparation evidence.
func registrySource(t *testing.T, token string) api.SourceVersion {
	t.Helper()
	v, err := api.NewSourceVersion("registry-test", "fixture", "public-observer", token)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func registrySummary(t *testing.T, operation uint64, kind api.GitMutationKind) api.MutationPlanSummary {
	t.Helper()
	repo, _ := domain.NewRepositoryID(domain.LocalCommon, "registry-fixture")
	worktree, _ := domain.NewWorktreeID(repo, "primary")
	branch, _ := domain.NewBranchID(repo, domain.Local, "main")
	oid, _ := domain.NewOID(strings.Repeat("1", 40))
	revision, _ := domain.NewRevision(repo, oid)
	head, _ := domain.NewAttachedHead(branch, revision)
	op, _ := api.NewOperationID(operation)
	expected, err := api.NewGitExpectedState(api.GitExpectedStateData{Repository: repo, Worktree: api.Some(worktree), Observation: registrySource(t, "observation"), Head: api.Some(head), Index: api.Some(registrySource(t, "index")), WorktreeState: api.Some(registrySource(t, "worktree")), Configuration: registrySource(t, "config"), Inventory: registrySource(t, "inventory")})
	if err != nil {
		t.Fatal(err)
	}
	d := api.MutationPlanSummaryData{OperationID: op, Kind: kind, PlanVersion: registrySource(t, "plan-"+strconv.FormatUint(operation, 10)), Repository: repo, Worktree: api.Some(worktree), Head: api.Some(head), Expected: expected, Choices: []api.Choice{api.Proceed, api.Cancel}, RecoveryBehavior: "native recovery is not exercised by this fixture"}
	switch kind {
	case api.BranchMutation:
		target, _ := domain.NewCommitTarget(revision)
		name, _ := domain.NewBranchID(repo, domain.Local, "new")
		d.Target = api.Some(target)
		d.Branch = api.Some(name)
	case api.CommitMutation:
		d.Message = api.Some("literal message")
		d.CommitIndexPolicy = api.Some(api.ObservedStageAll)
	case api.StageMutation:
		d.StageAction = api.Some(api.Stage)
	}
	value, err := api.NewMutationPlanSummary(d)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
func registryApproval(t *testing.T, authority ports.ApprovalIssuer, plan ports.PreparedGitPlan) ports.ExecutionApproval {
	t.Helper()
	digest, err := ports.PlanSummaryDigest(plan)
	if err != nil {
		t.Fatal(err)
	}
	approval, err := authority.Issue(plan, digest, api.Proceed, api.None[api.ConfirmationID]())
	if err != nil {
		t.Fatal(err)
	}
	return approval
}
func registryFixture(t *testing.T) (*planRegistry, ports.ApprovalIssuer, time.Time) {
	t.Helper()
	authority, _ := ports.NewApprovalIssuer("fixture-coordinator")
	registry, err := newPlanRegistry(authority, time.Minute, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	return registry, authority, time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
}

func TestPlanRegistryChecksCoordinatorAndSingleUse(t *testing.T) {
	r, authority, now := registryFixture(t)
	plan, err := r.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now)
	if err != nil {
		t.Fatal(err)
	}
	foreign, _ := ports.NewApprovalIssuer("different-coordinator")
	wrong := registryApproval(t, foreign, plan)
	if !wrong.ValidFor(plan) {
		t.Fatal("portable control should bind the same plan")
	}
	if err = r.begin(plan, wrong, now); err == nil {
		t.Fatal("accepted a different coordinator issuer")
	}
	approval := registryApproval(t, authority, plan)
	var successes atomic.Int32
	var workers sync.WaitGroup
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if r.begin(plan, approval, now) == nil {
				successes.Add(1)
			}
		}()
	}
	workers.Wait()
	if successes.Load() != 1 {
		t.Fatalf("native execution admitted %d times", successes.Load())
	}
	if _, err = r.finish(plan, nil, true, false); err != nil {
		t.Fatal(err)
	}
	if err = r.begin(plan, approval, now); err == nil {
		t.Fatal("replayed consumed plan")
	}
	if err = r.release(plan); err != nil {
		t.Fatal(err)
	}
	if err = r.release(plan); err != nil {
		t.Fatal("release not idempotent", err)
	}
	if r.bytes != 0 || len(r.groups) != 0 {
		t.Fatal("released root reservation leaked")
	}
}

func TestPlanRegistryReadOnlyBoundsAndExpiry(t *testing.T) {
	r, _, now := registryFixture(t)
	readonly, err := newPlanRegistry(ports.ApprovalIssuer{}, time.Minute, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readonly.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now); err == nil {
		t.Fatal("read-only adapter issued executable authority")
	}
	for i := uint64(1); i <= 64; i++ {
		if _, err = r.issueRoot(registrySummary(t, i, api.BranchMutation), ports.Executable, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = r.issueRoot(registrySummary(t, 65, api.BranchMutation), ports.Executable, now); err == nil {
		t.Fatal("exceeded the64 operation reservation bound")
	}
	if _, err = r.issueRoot(registrySummary(t, 65, api.BranchMutation), ports.Executable, now.Add(time.Minute)); err != nil {
		t.Fatal("expired ephemeral plans did not release admission", err)
	}
	small, authority, _ := registryFixture(t)
	small.maxBytes = 1
	if _, err = small.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now); err == nil || small.bytes != 0 {
		t.Fatal("retained-byte refusal admitted a plan")
	}
	ordinary, _, _ := registryFixture(t)
	plan, err := ordinary.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now)
	if err != nil {
		t.Fatal(err)
	}
	approval := registryApproval(t, authority, plan)
	if err = ordinary.begin(plan, approval, now.Add(time.Minute)); err == nil {
		t.Fatal("expired consent executed")
	}
}

func TestPlanRegistryReleaseCannotAbandonExecution(t *testing.T) {
	r, authority, now := registryFixture(t)
	plan, err := r.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = r.begin(plan, registryApproval(t, authority, plan), now); err != nil {
		t.Fatal(err)
	}
	if err = r.release(plan); err != nil {
		t.Fatal(err)
	}
	if len(r.groups) != 1 || r.bytes == 0 {
		t.Fatal("release abandoned admitted execution")
	}
	if _, err = r.finish(plan, nil, true, false); err != nil {
		t.Fatal(err)
	}
	if len(r.groups) != 0 || r.bytes != 0 {
		t.Fatal("barrier did not dispose pending release")
	}
}

func TestPlanRegistryForeignPlansAndReceiptFailureStayConsumed(t *testing.T) {
	r, authority, now := registryFixture(t)
	other, _, _ := registryFixture(t)
	foreign, err := other.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = r.begin(foreign, registryApproval(t, authority, foreign), now); err == nil {
		t.Fatal("accepted another native issuer lifetime")
	}
	if err = r.release(foreign); err == nil {
		t.Fatal("released another native issuer's plan")
	}
	plan, err := r.issueRoot(registrySummary(t, 1, api.BranchMutation), ports.Executable, now)
	if err != nil {
		t.Fatal(err)
	}
	approval := registryApproval(t, authority, plan)
	if err = r.begin(plan, approval, now); err != nil {
		t.Fatal(err)
	}
	if _, err = r.finish(plan, []api.FacetVersion{{}}, true, false); err == nil {
		t.Fatal("invalid receipt versions accepted")
	}
	if err = r.begin(plan, approval, now); err == nil {
		t.Fatal("receipt construction failure allowed native replay")
	}
	if err = r.release(plan); err != nil {
		t.Fatal(err)
	}
	if r.bytes != 0 || len(r.groups) != 0 {
		t.Fatal("receipt failure stranded execution memory")
	}
}

func TestPlanRegistrySequenceUsesActualVerifiedCompletion(t *testing.T) {
	r, authority, now := registryFixture(t)
	rootSummary := registrySummary(t, 1, api.CommitMutation)
	root, err := r.issueRoot(rootSummary, ports.SequenceRoot, now)
	if err != nil {
		t.Fatal(err)
	}
	approval := registryApproval(t, authority, root)
	if err = r.begin(root, approval, now); err == nil {
		t.Fatal("sequence root executed directly")
	}
	stageData := registrySummary(t, 1, api.StageMutation).Data()
	stageData.PlanVersion = registrySource(t, "child1")
	stageData.OriginVersion = api.Some(rootSummary.Data().PlanVersion)
	stage, err := api.NewMutationPlanSummary(stageData)
	if err != nil {
		t.Fatal(err)
	}
	first, err := r.issueChild(root, stage, api.None[ports.GitMutationReceipt](), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.issueChild(root, stage, api.None[ports.GitMutationReceipt](), now); err == nil {
		t.Fatal("first step issued twice")
	}
	if err = r.begin(first, approval, now); err != nil {
		t.Fatal(err)
	}
	version, err := api.NewFacetVersion(api.FacetVersionData{Facet: api.Index, Before: registrySource(t, "index"), After: registrySource(t, "staged")})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := r.finish(first, []api.FacetVersion{version}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = r.release(first); err != nil {
		t.Fatal(err)
	} // Root owns continuation authority.
	commitData := rootSummary.Data()
	commitData.PlanVersion = registrySource(t, "child2")
	commitData.OriginVersion = api.Some(rootSummary.Data().PlanVersion)
	commitData.CommitIndexPolicy = api.Some(api.ExistingIndex)
	commitData.OwnStepVersions = []api.FacetVersion{version}
	commit, err := api.NewMutationPlanSummary(commitData)
	if err != nil {
		t.Fatal(err)
	}
	changed := commitData
	changed.Message = api.Some("another message")
	changedSummary, err := api.NewMutationPlanSummary(changed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.issueChild(root, changedSummary, api.Some(receipt), now); err == nil {
		t.Fatal("continuation changed the approved literal commit message")
	}
	second, err := r.issueChild(root, commit, api.Some(receipt), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.issueChild(root, commit, api.Some(receipt), now); err == nil {
		t.Fatal("predecessor was replayed into another step")
	}
	if err = r.begin(second, approval, now); err != nil {
		t.Fatal(err)
	}
	if _, err = r.finish(second, nil, true, false); err != nil {
		t.Fatal(err)
	}
	if err = r.release(root); err != nil {
		t.Fatal(err)
	}
	if len(r.groups) != 0 {
		t.Fatal("sequence release leaked history")
	}
}

func TestPlanRegistryRejectsUnverifiedAndCanceledPredecessor(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		r, authority, now := registryFixture(t)
		summary := registrySummary(t, 1, api.CommitMutation)
		root, err := r.issueRoot(summary, ports.SequenceRoot, now)
		if err != nil {
			t.Fatal(err)
		}
		stageData := registrySummary(t, 1, api.StageMutation).Data()
		stageData.PlanVersion = registrySource(t, "stage")
		stageData.OriginVersion = api.Some(summary.Data().PlanVersion)
		stage, _ := api.NewMutationPlanSummary(stageData)
		first, err := r.issueChild(root, stage, api.None[ports.GitMutationReceipt](), now)
		if err != nil {
			t.Fatal(err)
		}
		if err = r.begin(first, registryApproval(t, authority, root), now); err != nil {
			t.Fatal(err)
		}
		receipt, err := r.finish(first, nil, false, canceled)
		if err != nil {
			t.Fatal(err)
		}
		child := summary.Data()
		child.PlanVersion = registrySource(t, "commit")
		child.OriginVersion = api.Some(summary.Data().PlanVersion)
		child.CommitIndexPolicy = api.Some(api.ExistingIndex)
		commit, _ := api.NewMutationPlanSummary(child)
		if _, err = r.issueChild(root, commit, api.Some(receipt), now); err == nil {
			t.Fatal("failed/canceled predecessor authorized continuation")
		}
		r.release(root)
	}
}
