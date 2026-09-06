package domain

import "testing"

// Exercise invalid tag/payload combinations that public constructors cannot
// express. This guards Valid if package-local representations evolve later.
func TestContradictoryClosedValuesAreInvalid(t *testing.T) {
	repo, _ := NewRepositoryID(LocalCommon, "local")
	remote, _ := NewRepositoryID(Remote, "remote")
	oid, _ := NewOID("1234567890123456789012345678901234567890")
	rev, _ := NewRevision(repo, oid)
	remoteRev, _ := NewRevision(remote, oid)
	branch, _ := NewBranchID(repo, Local, "main")
	pr, _ := NewPRID(remote, 1)
	for _, h := range []Head{
		{}, {kind: HeadKind(255), branch: branch, revision: rev},
		{kind: Attached, revision: rev}, {kind: Attached, branch: branch},
		{kind: Detached, branch: branch, revision: rev},
		{kind: Unborn, branch: branch, revision: rev},
	} {
		if h.Valid() || h.Repository().Valid() {
			t.Fatalf("contradictory HEAD accepted: %+v", h)
		}
		if value, ok := h.Revision(); ok || value.Valid() {
			t.Fatal("invalid HEAD exposed revision")
		}
		if value, ok := h.Branch(); ok || value.Valid() {
			t.Fatal("invalid HEAD exposed branch")
		}
	}
	for _, target := range []ExactTarget{
		{}, {kind: TargetKind(255), expected: rev},
		{kind: CommitTarget}, {kind: CommitTarget, expected: rev, branch: branch},
		{kind: CommitTarget, expected: rev, pr: pr},
		{kind: BranchTarget, expected: rev}, {kind: BranchTarget, branch: branch},
		{kind: BranchTarget, branch: branch, expected: rev, pr: pr},
		{kind: PullRequestTarget, expected: remoteRev}, {kind: PullRequestTarget, pr: pr},
		{kind: PullRequestTarget, pr: pr, expected: remoteRev, branch: branch},
	} {
		if target.Valid() || target.ExpectedRevision().Valid() {
			t.Fatalf("contradictory target accepted: %+v", target)
		}
		if value, ok := target.Branch(); ok || value.Valid() {
			t.Fatal("invalid target exposed branch")
		}
		if value, ok := target.PullRequest(); ok || value.Valid() {
			t.Fatal("invalid target exposed PR")
		}
	}
	badTail := oid
	badTail.bytes[31] = 1
	badFormat := oid
	badFormat.format = ObjectFormat(255)
	for _, invalid := range []OID{{}, badTail, badFormat} {
		if invalid.Valid() || invalid.String() != "" || invalid.Bytes() != nil {
			t.Fatal("noncanonical OID accepted")
		}
	}
}
