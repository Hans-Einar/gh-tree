package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

func TestIndependentKnownOutcomeEffect(t *testing.T) {
	repo := rvRepo("effect")
	branch := rvMust(d.NewBranchID(repo, d.Local, "created"))
	out := rvMust(a.NewBranchCreated(a.BranchCreatedData{Branch: branch, Revision: rvRev(repo, "1")}))
	for _, c := range []struct {
		name   string
		state  a.EffectState
		reject bool
	}{{"positive_known_creation", a.AppliedVerified, false}, {"reject_not_started_creation", a.NotStarted, true}, {"reject_unchanged_creation", a.VerifiedNoTargetChange, true}} {
		t.Run(c.name, func(t *testing.T) {
			_, e := a.NewGitMutationResult(a.GitMutationResultData{Operation: rvMust(a.NewOperationID(1)), Kind: a.BranchMutation, PlanVersion: rvSource("plan"), Outcome: out, Transport: rvTransport(), Effects: rvEffect(a.LocalRefsHead, c.state)})
			if c.reject && e == nil {
				t.Fatal("created branch paired with explicit unchanged/unstarted local-ref effect")
			}
			if !c.reject && e != nil {
				t.Fatal(e)
			}
		})
	}
}
