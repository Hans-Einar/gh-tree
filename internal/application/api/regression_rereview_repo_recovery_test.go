package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

func iRecoveryFor(repo d.RepositoryID) a.RecoveryRecord {
	return rvMust(a.NewRecoveryRecord(a.RecoveryRecordData{RecoveryID: rvMust(a.NewRecoveryID("recovery")), Kind: a.RecoveryObjects, Layer: a.LayerGit, Subject: rvMust(a.NewRecoverySubject(a.RecoverySubjectData{Repository: a.Some(repo)})), Locator: "owned/recovery", NextAction: "inspect"}))
}
func TestIndependentRepoOnlyMutationRecovery(t *testing.T) {
	repo := rvRepo("A")
	other := rvRepo("B")
	branch := rvMust(d.NewBranchID(repo, d.Local, "new"))
	out := rvMust(a.NewBranchCreated(a.BranchCreatedData{Branch: branch, Revision: rvRev(repo, "1")}))
	data := a.GitMutationResultData{Operation: rvMust(a.NewOperationID(1)), Kind: a.BranchMutation, PlanVersion: rvSource("plan"), Outcome: out, Transport: rvTransport(), Effects: rvEffects(), Recovery: []a.RecoveryRecord{iRecoveryFor(repo)}}
	t.Run("positive_same_repository", func(t *testing.T) {
		if _, e := a.NewGitMutationResult(data); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("reject_foreign_repository_recovery", func(t *testing.T) {
		data.Recovery = []a.RecoveryRecord{iRecoveryFor(other)}
		if _, e := a.NewGitMutationResult(data); e == nil {
			t.Fatal("known repository-A mutation retained repository-B recovery")
		}
	})
}
