package domain_test

import (
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func TestFullOIDFormatsCaseAndCopy(t *testing.T) {
	for _, tc := range []struct {
		format domain.ObjectFormat
		size   int
	}{{domain.SHA1, 20}, {domain.SHA256, 32}} {
		text := strings.Repeat("aB", tc.size)
		o := must(domain.NewOID(text))
		canonical := strings.ToLower(text)
		if !o.Valid() || o.Format() != tc.format || o.Format().ByteLength() != tc.size || o.String() != canonical {
			t.Fatalf("OID format/string mismatch: %v", o)
		}
		if !o.Equal(must(domain.NewOID(canonical))) || o != must(domain.NewOID(o.String())) {
			t.Fatal("OID case/roundtrip mismatch")
		}
		bytes := o.Bytes()
		if len(bytes) != tc.size || bytes[0] != 0xab {
			t.Fatal("OID byte accessor lost data")
		}
		for i := range bytes {
			bytes[i] = 0
		}
		if o.String() != canonical || o.Bytes()[0] != 0xab {
			t.Fatal("OID bytes escaped by reference")
		}
	}
	sha1 := must(domain.NewOID(strings.Repeat("12", 20)))
	sha256 := must(domain.NewOID(strings.Repeat("12", 20) + strings.Repeat("00", 12)))
	if sha1 == sha256 || sha1.Equal(sha256) {
		t.Fatal("object formats aliased")
	}
	for _, length := range []int{40, 64} {
		for _, text := range []string{"1" + strings.Repeat("0", length-1), strings.Repeat("0", length-1) + "1"} {
			if !must(domain.NewOID(text)).Valid() {
				t.Fatal("nonzero edge byte rejected")
			}
		}
	}
}

func TestOIDRejectsExpressionsAndInvalidBytes(t *testing.T) {
	invalid := []string{"", "abc1234", "HEAD", "HEAD~2", "origin/main", "refs/heads/main", "0x" + strings.Repeat("a", 40)}
	for _, n := range []int{1, 7, 39, 41, 63, 65, 128} {
		invalid = append(invalid, strings.Repeat("a", n))
	}
	for _, n := range []int{40, 64} {
		invalid = append(invalid, strings.Repeat("0", n), " "+strings.Repeat("a", n), strings.Repeat("a", n)+"\n")
		for _, bad := range []string{" ", "\t", "\n", "\r", "\x00", "g", "G", "-", "\xff"} {
			invalid = append(invalid, bad+strings.Repeat("a", n-1), strings.Repeat("a", n/2)+bad+strings.Repeat("a", n-n/2-1))
		}
	}
	for _, text := range invalid {
		got, err := domain.NewOID(text)
		if err == nil || got.Valid() || got != (domain.OID{}) {
			t.Fatalf("accepted invalid OID %q", text)
		}
	}
}

func TestRevisionScopeCannotBeInferredFromOID(t *testing.T) {
	a := revision(local("clone-A"))
	b := revision(local("clone-B"))
	c := revision(remote("clone-A"))
	if a == b || a == c || a.Equal(c) || a.OID() != b.OID() || a.OID() != c.OID() {
		t.Fatal("equal bytes erased repository scope")
	}
	if !a.Valid() || a.Repository() != local("clone-A") || !a.Equal(revision(local("clone-A"))) {
		t.Fatal("revision identity lost")
	}
	for _, tc := range []struct {
		repo domain.RepositoryID
		oid  domain.OID
	}{{domain.RepositoryID{}, oid()}, {local("a"), domain.OID{}}} {
		got, err := domain.NewRevision(tc.repo, tc.oid)
		if err == nil || got != (domain.Revision{}) {
			t.Fatal("invalid revision accepted")
		}
	}
}

func TestHeadAlternativesAndWorktreeScope(t *testing.T) {
	r := local("repo")
	b := branch(r, domain.Local, "main")
	exact := revision(r)
	attached := must(domain.NewAttachedHead(b, exact))
	detached := must(domain.NewDetachedHead(exact))
	unborn := must(domain.NewUnbornHead(b))
	for _, tc := range []struct {
		head             domain.Head
		kind             domain.HeadKind
		branch, revision bool
	}{
		{attached, domain.Attached, true, true}, {detached, domain.Detached, false, true}, {unborn, domain.Unborn, true, false},
	} {
		if !tc.head.Valid() || tc.head.Kind() != tc.kind || tc.head.Repository() != r {
			t.Fatal("HEAD alternative invalid")
		}
		gotB, presentB := tc.head.Branch()
		gotR, presentR := tc.head.Revision()
		if presentB != tc.branch || presentR != tc.revision {
			t.Fatal("HEAD payload presence mismatch")
		}
		if presentB && gotB != b || !presentB && gotB != (domain.BranchID{}) {
			t.Fatal("HEAD branch mismatch")
		}
		if presentR && gotR != exact || !presentR && gotR != (domain.Revision{}) {
			t.Fatal("HEAD revision mismatch")
		}
		for _, key := range []string{"primary", "linked"} {
			if !tc.head.MatchesWorktree(worktree(r, key)) {
				t.Fatal("linked common scope rejected")
			}
		}
		if tc.head.MatchesWorktree(worktree(local("clone"), "primary")) || tc.head.MatchesWorktree(domain.WorktreeID{}) {
			t.Fatal("foreign/invalid worktree scope accepted")
		}
	}
	if attached == detached || attached == unborn || detached == unborn || !attached.Equal(must(domain.NewAttachedHead(b, exact))) {
		t.Fatal("HEAD alternatives aliased")
	}
	if value, ok := unborn.Revision(); ok || value.Valid() {
		t.Fatal("unborn fabricated an exact commit")
	}
	if _, err := domain.NewCommitTarget(func() domain.Revision { r, _ := unborn.Revision(); return r }()); err == nil {
		t.Fatal("unborn entered exact-target operation")
	}
	for _, tc := range []struct {
		branch   domain.BranchID
		revision domain.Revision
	}{
		{domain.BranchID{}, exact}, {b, domain.Revision{}}, {b, revision(local("foreign"))}, {b, revision(remote("repo"))},
		{branch(remote("repo"), domain.RemoteHead, "main"), revision(remote("repo"))},
	} {
		got, err := domain.NewAttachedHead(tc.branch, tc.revision)
		if err == nil || got != (domain.Head{}) {
			t.Fatal("invalid attached HEAD accepted")
		}
	}
	for _, exact := range []domain.Revision{{}, revision(remote("repo"))} {
		if got, err := domain.NewDetachedHead(exact); err == nil || got != (domain.Head{}) {
			t.Fatal("invalid detached HEAD accepted")
		}
	}
	for _, b := range []domain.BranchID{{}, branch(remote("repo"), domain.RemoteHead, "main")} {
		if got, err := domain.NewUnbornHead(b); err == nil || got != (domain.Head{}) {
			t.Fatal("invalid unborn HEAD accepted")
		}
	}
}

func TestExactTargetPreservesIntentAndForkHead(t *testing.T) {
	base, fork, clone := remote("base"), remote("fork"), local("clone")
	pr := must(domain.NewPRID(base, 8))
	localBranch := branch(clone, domain.Local, "main")
	remoteBranch := branch(fork, domain.RemoteHead, "main")
	commit := must(domain.NewCommitTarget(revision(clone)))
	branchTarget := must(domain.NewBranchTarget(localBranch, revision(clone)))
	remoteTarget := must(domain.NewBranchTarget(remoteBranch, revision(fork)))
	prTarget := must(domain.NewPullRequestTarget(pr, revision(fork)))
	if prTarget.ExpectedRevision().Repository() != fork {
		t.Fatal("PR head was constrained/rewritten to base")
	}
	if prTarget == must(domain.NewPullRequestTarget(pr, revision(base))) {
		t.Fatal("PR fork scope was dropped")
	}
	if prTarget == must(domain.NewPullRequestTarget(must(domain.NewPRID(remote("other-base"), 8)), revision(fork))) {
		t.Fatal("PR base scope was dropped")
	}
	for _, tc := range []struct {
		target     domain.ExactTarget
		kind       domain.TargetKind
		expected   domain.Revision
		branch, pr bool
	}{
		{commit, domain.CommitTarget, revision(clone), false, false},
		{branchTarget, domain.BranchTarget, revision(clone), true, false},
		{remoteTarget, domain.BranchTarget, revision(fork), true, false},
		{prTarget, domain.PullRequestTarget, revision(fork), false, true},
	} {
		copy := tc.target
		if !copy.Valid() || copy.Kind() != tc.kind || copy.ExpectedRevision() != tc.expected || !copy.Equal(tc.target) {
			t.Fatal("copied target lost expected revision")
		}
		b, hasB := copy.Branch()
		p, hasP := copy.PullRequest()
		if hasB != tc.branch || hasP != tc.pr || hasB && b.Repository() != tc.expected.Repository() || hasP && p != pr {
			t.Fatal("target payload mismatch")
		}
		if !hasB && b.Valid() || !hasP && p.Valid() {
			t.Fatal("target exposed irrelevant payload")
		}
	}
	if commit == branchTarget || branchTarget == remoteTarget {
		t.Fatal("target alternatives/scopes aliased")
	}
	newTip := must(domain.NewRevision(fork, must(domain.NewOID(strings.Repeat("b2", 20)))))
	if prTarget == must(domain.NewPullRequestTarget(pr, newTip)) || remoteTarget == must(domain.NewBranchTarget(remoteBranch, newTip)) {
		t.Fatal("expected revision omitted from equality")
	}
	if got, err := domain.NewCommitTarget(domain.Revision{}); err == nil || got != (domain.ExactTarget{}) {
		t.Fatal("zero commit target accepted")
	}
	for _, tc := range []struct {
		branch   domain.BranchID
		revision domain.Revision
	}{
		{domain.BranchID{}, revision(clone)}, {localBranch, domain.Revision{}}, {localBranch, revision(local("other"))}, {localBranch, revision(fork)}, {remoteBranch, revision(base)},
	} {
		if got, err := domain.NewBranchTarget(tc.branch, tc.revision); err == nil || got != (domain.ExactTarget{}) {
			t.Fatal("invalid branch target accepted")
		}
	}
	for _, tc := range []struct {
		pr       domain.PRID
		revision domain.Revision
	}{{domain.PRID{}, revision(fork)}, {pr, domain.Revision{}}, {pr, revision(clone)}} {
		if got, err := domain.NewPullRequestTarget(tc.pr, tc.revision); err == nil || got != (domain.ExactTarget{}) {
			t.Fatal("invalid PR target accepted")
		}
	}
}
