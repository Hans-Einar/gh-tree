package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	"testing"
)

func TestIndependentPlanChoiceKind(t *testing.T) {
	w := rvWork(rvRepo("choice"), "one")
	expected := rvExpected(w)
	data := a.MutationPlanSummaryData{OperationID: rvMust(a.NewOperationID(1)), Kind: a.StageMutation, PlanVersion: rvSource("plan"), Repository: w.Repository(), Worktree: a.Some(w), Head: expected.Data().Head, Expected: expected, Choices: []a.Choice{a.Proceed, a.Cancel}, StageAction: a.Some(a.Stage)}
	t.Run("positive_stage_choices", func(t *testing.T) {
		if _, e := a.NewMutationPlanSummary(data); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("reject_stash_then_deploy_for_stage", func(t *testing.T) {
		data.Choices = []a.Choice{a.StashThenDeploy}
		if _, e := a.NewMutationPlanSummary(data); e == nil {
			t.Fatal("Stage summary advertises a StashThenDeploy continuation")
		}
	})
}
