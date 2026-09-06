package domain_test

import (
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func local(token string) domain.RepositoryID {
	return must(domain.NewRepositoryID(domain.LocalCommon, token))
}
func remote(token string) domain.RepositoryID {
	return must(domain.NewRepositoryID(domain.Remote, token))
}
func oid() domain.OID { return must(domain.NewOID(strings.Repeat("a1", 20))) }
func revision(repository domain.RepositoryID) domain.Revision {
	return must(domain.NewRevision(repository, oid()))
}
func worktree(repository domain.RepositoryID, key string) domain.WorktreeID {
	return must(domain.NewWorktreeID(repository, key))
}
func branch(repository domain.RepositoryID, kind domain.BranchKind, name string) domain.BranchID {
	return must(domain.NewBranchID(repository, kind, name))
}

func TestRepositoryScopeAndExactToken(t *testing.T) {
	a := local("same-token")
	b := remote("same-token")
	if a == b || a.Equal(b) {
		t.Fatal("local and remote repository scopes aliased")
	}
	if !a.Valid() || a.Scope() != domain.LocalCommon || a.Token() != "same-token" || !a.Equal(local("same-token")) {
		t.Fatal("local repository did not preserve identity")
	}
	for _, token := range []string{" Name ", "name", "Name", "x/y.git", "x\\y", "x\x00y", "\xff"} {
		got := local(token)
		if got.Token() != token {
			t.Fatalf("opaque token %q was transformed", token)
		}
	}
	if local("Name") == local("name") || local(" Name ") == local("Name") {
		t.Fatal("domain canonicalized an opaque token")
	}
	for _, tc := range []struct {
		scope domain.RepositoryScope
		token string
	}{
		{0, "repo"}, {domain.RepositoryScope(255), "repo"}, {domain.Remote, ""}, {domain.LocalCommon, ""},
	} {
		got, err := domain.NewRepositoryID(tc.scope, tc.token)
		if err == nil || got.Valid() || got != (domain.RepositoryID{}) {
			t.Fatalf("accepted invalid repository: %+v", tc)
		}
	}
}

func TestWorktreeAndStashScope(t *testing.T) {
	common := local("clone-A")
	primary := worktree(common, "primary")
	linked := worktree(common, "linked-id")
	clone := worktree(local("clone-B"), "linked-id")
	if primary == linked || linked == clone || primary.Repository() != linked.Repository() {
		t.Fatal("administrative or common-repository identity was lost")
	}
	if !linked.Valid() || linked.AdministrativeKey() != "linked-id" || !linked.Equal(worktree(common, "linked-id")) {
		t.Fatal("worktree accessors/equality lost identity")
	}
	// Retargeting, display paths and reflog positions are observations. There is
	// no parameter for them in these constructors; the same supplied identity
	// remains equal across primary/linked worktree observations.
	a := must(domain.NewStashID(primary.Repository(), oid()))
	b := must(domain.NewStashID(linked.Repository(), oid()))
	c := must(domain.NewStashID(clone.Repository(), oid()))
	if !a.Valid() || a != b || !a.Equal(b) || a == c || a.OID() != oid() || a.Repository() != common {
		t.Fatal("stash did not retain exact common-repository/object identity")
	}
	for _, repository := range []domain.RepositoryID{{}, remote("clone-A")} {
		got, err := domain.NewWorktreeID(repository, "key")
		if err == nil || got != (domain.WorktreeID{}) {
			t.Fatal("invalid/remote worktree accepted")
		}
		stash, err := domain.NewStashID(repository, oid())
		if err == nil || stash != (domain.StashID{}) {
			t.Fatal("invalid/remote stash accepted")
		}
	}
	if got, err := domain.NewWorktreeID(common, ""); err == nil || got != (domain.WorktreeID{}) {
		t.Fatal("empty administrative key accepted")
	}
	if got, err := domain.NewStashID(common, domain.OID{}); err == nil || got != (domain.StashID{}) {
		t.Fatal("zero stash OID accepted")
	}
	if worktree(common, " key ").AdministrativeKey() != " key " {
		t.Fatal("administrative key was trimmed")
	}
}

func TestBranchKindsAndExactStoredNames(t *testing.T) {
	l := branch(local("repo"), domain.Local, "main")
	r := branch(remote("repo"), domain.RemoteHead, "main")
	if l == r || l.Equal(r) || !l.Valid() || l.Repository() != local("repo") || l.Kind() != domain.Local || l.Name() != "main" {
		t.Fatal("branch scope/kind/name lost")
	}
	if !l.Equal(branch(local("repo"), domain.Local, "main")) {
		t.Fatal("same branch identity differs")
	}
	for _, tc := range []struct {
		repo domain.RepositoryID
		kind domain.BranchKind
	}{
		{domain.RepositoryID{}, domain.Local}, {local("repo"), 0}, {remote("repo"), domain.BranchKind(255)},
		{local("repo"), domain.RemoteHead}, {remote("repo"), domain.Local},
	} {
		got, err := domain.NewBranchID(tc.repo, tc.kind, "main")
		if err == nil || got != (domain.BranchID{}) {
			t.Fatalf("invalid branch scope/kind accepted: %+v", tc)
		}
	}
	// These are literal stored suffixes beneath refs/heads/, not porcelain
	// creation operands. Prefix-like and special-looking bytes stay literal.
	for _, name := range []string{"main", "topic/nested", "-topic", "HEAD", "@", "refs/heads/main", "origin/main", "a./b", "x.LOCK", "über/枝", "a\xffb", "brace}ok", "a+b"} {
		for _, tc := range []struct {
			repo domain.RepositoryID
			kind domain.BranchKind
		}{{local("repo"), domain.Local}, {remote("repo"), domain.RemoteHead}} {
			got, err := domain.NewBranchID(tc.repo, tc.kind, name)
			if err != nil || got.Name() != name {
				t.Fatalf("exact valid stored suffix %q: value=%q error=%v", name, got.Name(), err)
			}
		}
	}
	if branch(local("repo"), domain.Local, "refs/heads/main") == l || branch(local("repo"), domain.Local, "origin/main") == l {
		t.Fatal("display/native-looking prefix was stripped")
	}
	invalid := []string{"", "/main", "main/", "a//b", ".hidden", "a/.hidden/b", "x.lock", "x.lock/y", "x.lock.lock", "a..b", "main.", "a@{0}", "@{-1}", "@{-12}", "a b", " main", "main ", "a~b", "a^b", "a:b", "a?b", "a*b", "a[b", "a\\b"}
	for c := byte(0); c <= 0x20; c++ {
		invalid = append(invalid, "a"+string([]byte{c})+"b")
	}
	invalid = append(invalid, "a\x7fb")
	for _, name := range invalid {
		got, err := domain.NewBranchID(local("repo"), domain.Local, name)
		if err == nil || got != (domain.BranchID{}) {
			t.Fatalf("invalid stored branch %q accepted", name)
		}
	}
}

func TestPRAndSessionIdentity(t *testing.T) {
	p := must(domain.NewPRID(remote("base"), 7))
	if !p.Valid() || p.Number() != 7 || p.Repository() != remote("base") || !p.Equal(must(domain.NewPRID(remote("base"), 7))) {
		t.Fatal("PR accessors/equality lost identity")
	}
	if p == must(domain.NewPRID(remote("fork"), 7)) || p == must(domain.NewPRID(remote("base"), 8)) {
		t.Fatal("PR repository or number omitted")
	}
	for _, tc := range []struct {
		repo   domain.RepositoryID
		number uint64
	}{
		{domain.RepositoryID{}, 7}, {local("base"), 7}, {remote("base"), 0},
	} {
		got, err := domain.NewPRID(tc.repo, tc.number)
		if err == nil || got != (domain.PRID{}) {
			t.Fatal("invalid PR accepted")
		}
	}
	if got, err := domain.NewSessionID(0); err == nil || got != (domain.SessionID{}) {
		t.Fatal("zero session accepted")
	}
	for _, number := range []uint64{1, 2, 1 << 63, ^uint64(0)} {
		id := must(domain.NewSessionID(number))
		if !id.Valid() || id.Value() != number || !id.Equal(must(domain.NewSessionID(number))) {
			t.Fatal("session identity changed")
		}
	}
	if must(domain.NewSessionID(1)) == must(domain.NewSessionID(2)) {
		t.Fatal("session numbers aliased")
	}
	// Accepting max uint64 is value validation, not allocation or wrap handling.
	// A second construction with the same supplied number remains equal; Runtime
	// must independently refuse reuse/exhaustion in its registry contract tests.
}

func TestLaunchIdentityHasUnambiguousBoundaries(t *testing.T) {
	w := worktree(local("repo"), "primary")
	type parts struct{ provider, project, member string }
	unique := []parts{
		{"npm", "", "a/b"}, {"npm", "a", "b"}, {"make", "", "a/b"},
		{"a", "bc", "d"}, {"ab", "c", "d"}, {"ab", "", "cd"},
		{"a:b", "c", "d"}, {"a", "b:c", "d"}, {"a", "b", "c:d"},
		{"npm", "", "dev"}, {"npm", "", "dev:wan"}, {"npm", "", " dev "},
		{"npm", "枝", "λ"}, {"npm", "\xff", "\x00"},
	}
	ids := map[domain.LaunchPointID]parts{}
	keys := map[string]parts{}
	for _, p := range unique {
		id := must(domain.NewLaunchPointID(w, p.provider, p.project, p.member))
		if !id.Valid() || id.Worktree() != w || !id.Equal(must(domain.NewLaunchPointID(w, p.provider, p.project, p.member))) {
			t.Fatal("launch identity not retained")
		}
		if previous, ok := ids[id]; ok {
			t.Fatalf("identity collision: %+v vs %+v", previous, p)
		}
		if previous, ok := keys[id.Key()]; ok {
			t.Fatalf("key collision: %+v vs %+v", previous, p)
		}
		ids[id], keys[id.Key()] = p, p
	}
	unicode := must(domain.NewLaunchPointID(w, "npm", "枝", "λ"))
	if unicode.Key() != "3:npm3:枝2:λ" {
		t.Fatalf("key lengths are not exact byte lengths: %q", unicode.Key())
	}
	a := must(domain.NewLaunchPointID(w, "npm", "", "dev"))
	for _, other := range []domain.WorktreeID{worktree(local("repo"), "linked"), worktree(local("clone"), "primary")} {
		if a == must(domain.NewLaunchPointID(other, "npm", "", "dev")) {
			t.Fatal("launch worktree scope aliased")
		}
	}
	for _, tc := range []struct {
		worktree                  domain.WorktreeID
		provider, project, member string
	}{
		{domain.WorktreeID{}, "npm", "", "dev"}, {w, "", "", "dev"}, {w, "npm", "", ""},
	} {
		got, err := domain.NewLaunchPointID(tc.worktree, tc.provider, tc.project, tc.member)
		if err == nil || got != (domain.LaunchPointID{}) {
			t.Fatal("invalid launch identity accepted")
		}
	}
}
