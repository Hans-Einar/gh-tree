package ports_test

import (
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
)

// The consumer receives only the opaque plan, never issuance data or its digest.
func opaqueDigestPlan(token string) ports.PreparedGitPlan {
	s := spec(api.StageMutation, token)
	s.SummaryDigest = [32]byte{0: 19, 7: 83, 31: 241}
	return must(must(ports.NewPlanIssuer("digest-fixture")).IssueStage(s))
}

func TestPlanSummaryDigestOpaqueConsumer(t *testing.T) {
	plan := opaqueDigestPlan("original")
	digest := must(ports.PlanSummaryDigest(plan))
	approver := must(ports.NewApprovalIssuer("coordinator"))
	confirmation := api.Some(must(api.NewConfirmationID("consumed-by-coordinator")))
	approval := must(approver.Issue(plan, digest, api.Proceed, confirmation))
	if !approval.ValidFor(plan) || !approver.Issued(approval) || approval.ValidFor(opaqueDigestPlan("other")) {
		t.Fatal("approval lost exact plan binding")
	}
	digest[31] ^= 255
	if must(ports.PlanSummaryDigest(plan)) == digest {
		t.Fatal("returned digest aliases plan")
	}
	if _, err := approver.Issue(plan, digest, api.Proceed, confirmation); err == nil {
		t.Fatal("modified digest approved")
	}
	for _, choice := range []api.Choice{api.Cancel, api.StashThenDeploy, 0} {
		if _, err := approver.Issue(plan, must(ports.PlanSummaryDigest(plan)), choice, confirmation); err == nil {
			t.Fatalf("invalid or disallowed choice %v approved", choice)
		}
	}
	for _, missing := range []api.Optional[api.ConfirmationID]{api.None[api.ConfirmationID](), api.Some(api.ConfirmationID{})} {
		if _, err := approver.Issue(plan, must(ports.PlanSummaryDigest(plan)), api.Proceed, missing); err == nil {
			t.Fatal("missing or invalid confirmation approved")
		}
	}
}

func TestPlanSummaryDigestRejectsInvalidPlans(t *testing.T) {
	for _, plan := range []ports.PreparedGitPlan{
		nil, foreignPlan{}, foreignPlan{opaqueDigestPlan("embedded")},
		ports.CreatePlan{}, ports.RetargetPlan{}, ports.StagePlan{}, ports.CommitPlan{},
		ports.RestorePlan{}, ports.StashPlan{}, ports.BranchPlan{}, ports.PushPlan{},
		(*ports.CreatePlan)(nil), (*ports.RetargetPlan)(nil), (*ports.StagePlan)(nil), (*ports.CommitPlan)(nil),
		(*ports.RestorePlan)(nil), (*ports.StashPlan)(nil), (*ports.BranchPlan)(nil), (*ports.PushPlan)(nil),
	} {
		digest, err := ports.PlanSummaryDigest(plan)
		_, summaryErr := ports.PlanSummary(plan)
		if err == nil || summaryErr == nil || err.Error() != summaryErr.Error() || digest != [32]byte{} {
			t.Errorf("%T: expected zero digest and consistent invalid-plan error, got %x, %v", plan, digest, err)
		}
	}
}
