package graph

import (
	"context"
	"strings"
	"testing"

	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type fakeRunner struct {
	responses map[string][]byte
}

func (f fakeRunner) Run(_ context.Context, _ string, name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	for prefix, out := range f.responses {
		if strings.HasPrefix(key, prefix) {
			return out, nil
		}
	}
	return nil, nil
}

func TestLoadBuildsStructuredGraphAndAnnotations(t *testing.T) {
	shaA := strings.Repeat("a", 40)
	shaB := strings.Repeat("b", 40)
	shaC := strings.Repeat("c", 40)
	runner := fakeRunner{responses: map[string][]byte{
		"git log --all": []byte(shaC + "\x1f" + shaB + "\x1fAlice\x1f2026-09-04T08:00:00Z\x1fmerge work\x1e\n" + shaB + "\x1f" + shaA + "\x1fBob\x1f2026-09-03T08:00:00Z\x1ffeature\x1e\n" + shaA + "\x1f\x1fAlice\x1f2026-09-02T08:00:00Z\x1finitial\x1e\n"),
		"git for-each-ref": []byte("refs/heads/main\x1f" + shaC + "\x1f\nrefs/heads/feature\x1f" + shaB + "\x1f\nrefs/remotes/origin/main\x1f" + shaC + "\x1f\nrefs/tags/v0.3.0\x1f" + shaA + "\x1f\nrefs/remotes/origin/HEAD\x1f" + shaC + "\x1frefs/remotes/origin/main\n"),
		"git rev-parse HEAD^{commit}": []byte(shaB + "\n"),
		"git symbolic-ref --quiet --short HEAD": []byte("feature\n"),
	}}
	reader := NewReader(runner, "/repo")
	prs := []ghapi.PullRequest{{Number: 60, HeadBranch: "feature", BaseBranch: "main", HeadSHA: shaB, IsDraft: true}}
	wts := []worktree.Info{{Path: "/repo", Branch: "feature", Head: shaB, Primary: true}}
	snapshot, err := reader.Load(context.Background(), 10, 0, prs, wts)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Commits) != 3 || snapshot.SelectedHeadSHA != shaB || snapshot.SelectedHeadName != "feature" {
		t.Fatalf("snapshot=%#v", snapshot)
	}
	if !snapshot.RemoteRefsSeen || !strings.Contains(snapshot.RemoteFreshness, "may be stale") {
		t.Fatalf("remote freshness=%q", snapshot.RemoteFreshness)
	}
	feature := snapshot.Commits[1]
	if feature.SHA != shaB || len(feature.PRs) != 1 || feature.PRs[0].Number != 60 || len(feature.Worktrees) != 1 {
		t.Fatalf("feature annotations=%#v", feature)
	}
	var haveHEAD, haveLocal bool
	for _, d := range feature.Decorations {
		if d.Kind == RefHEAD { haveHEAD = true }
		if d.Kind == RefLocal && d.Name == "feature" { haveLocal = true }
	}
	if !haveHEAD || !haveLocal {
		t.Fatalf("decorations=%#v", feature.Decorations)
	}
	main := snapshot.Commits[0]
	if len(main.Decorations) != 2 || main.Decorations[0].Kind != RefLocal || main.Decorations[1].Kind != RefRemote {
		t.Fatalf("main decorations=%#v", main.Decorations)
	}
}

func TestLoadIsBoundedAndReportsMore(t *testing.T) {
	shaA := strings.Repeat("a", 40)
	shaB := strings.Repeat("b", 40)
	shaC := strings.Repeat("c", 40)
	runner := fakeRunner{responses: map[string][]byte{
		"git log --all": []byte(shaC + "\x1f" + shaB + "\x1fA\x1f2026-09-04T08:00:00Z\x1fc\x1e\n" + shaB + "\x1f" + shaA + "\x1fA\x1f2026-09-03T08:00:00Z\x1fb\x1e\n" + shaA + "\x1f\x1fA\x1f2026-09-02T08:00:00Z\x1fa\x1e\n"),
		"git for-each-ref": []byte{},
		"git rev-parse HEAD^{commit}": []byte(shaC + "\n"),
		"git symbolic-ref --quiet --short HEAD": []byte("main\n"),
	}}
	snapshot, err := NewReader(runner, "/repo").Load(context.Background(), 2, 0, nil, nil)
	if err != nil { t.Fatal(err) }
	if len(snapshot.Commits) != 2 || !snapshot.HasMore {
		t.Fatalf("bounded snapshot=%#v", snapshot)
	}
}

func TestClassifyRef(t *testing.T) {
	cases := []struct{ in string; kind RefKind; name string; ok bool }{
		{"refs/heads/main", RefLocal, "main", true},
		{"refs/remotes/origin/main", RefRemote, "origin/main", true},
		{"refs/tags/v1", RefTag, "v1", true},
		{"refs/notes/x", "", "", false},
	}
	for _, tc := range cases {
		kind, name, ok := classifyRef(tc.in)
		if kind != tc.kind || name != tc.name || ok != tc.ok { t.Fatalf("%s => %q %q %v", tc.in, kind, name, ok) }
	}
}
