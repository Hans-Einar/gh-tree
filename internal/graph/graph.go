package graph

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/process"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

const (
	defaultLimit = 200
	maxLimit     = 1000
)

type RefKind string

const (
	RefLocal  RefKind = "local"
	RefRemote RefKind = "remote"
	RefTag    RefKind = "tag"
	RefHEAD   RefKind = "head"
)

type Decoration struct {
	Kind RefKind
	Name string
}

type PRAnnotation struct {
	Number int
	Head   string
	Base   string
	Draft  bool
}

type WorktreeAnnotation struct {
	Path     string
	Branch   string
	Primary  bool
	Detached bool
}

type Commit struct {
	SHA         string
	Parents     []string
	Author      string
	Authored    time.Time
	Subject     string
	Decorations []Decoration
	PRs         []PRAnnotation
	Worktrees   []WorktreeAnnotation
}

type Snapshot struct {
	Commits          []Commit
	HasMore          bool
	RemoteRefsSeen   bool
	RemoteFreshness  string
	SelectedHeadSHA  string
	SelectedHeadName string
}

type Reader struct {
	runner process.Runner
	dir    string
}

func NewReader(runner process.Runner, dir string) *Reader {
	return &Reader{runner: runner, dir: dir}
}

func (r *Reader) Load(ctx context.Context, limit, skip int, prs []ghapi.PullRequest, worktrees []worktree.Info) (Snapshot, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if skip < 0 {
		skip = 0
	}

	commits, hasMore, err := r.loadCommits(ctx, limit, skip)
	if err != nil {
		return Snapshot{}, err
	}
	refs, remoteSeen, err := r.loadRefs(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	headSHA, headName, err := r.loadHEAD(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	bySHA := make(map[string]int, len(commits))
	for i := range commits {
		bySHA[commits[i].SHA] = i
	}
	for sha, decorations := range refs {
		if i, ok := bySHA[sha]; ok {
			commits[i].Decorations = append(commits[i].Decorations, decorations...)
		}
	}
	if i, ok := bySHA[headSHA]; ok {
		commits[i].Decorations = append(commits[i].Decorations, Decoration{Kind: RefHEAD, Name: "HEAD"})
	}
	for _, pr := range prs {
		if i, ok := bySHA[pr.HeadSHA]; ok {
			commits[i].PRs = append(commits[i].PRs, PRAnnotation{Number: pr.Number, Head: pr.HeadBranch, Base: pr.BaseBranch, Draft: pr.IsDraft})
		}
	}
	for _, wt := range worktrees {
		if i, ok := bySHA[wt.Head]; ok {
			commits[i].Worktrees = append(commits[i].Worktrees, WorktreeAnnotation{Path: wt.Path, Branch: wt.Branch, Primary: wt.Primary, Detached: wt.Detached})
		}
	}
	for i := range commits {
		sortDecorations(commits[i].Decorations)
		sort.Slice(commits[i].PRs, func(a, b int) bool { return commits[i].PRs[a].Number < commits[i].PRs[b].Number })
		sort.Slice(commits[i].Worktrees, func(a, b int) bool { return strings.ToLower(commits[i].Worktrees[a].Path) < strings.ToLower(commits[i].Worktrees[b].Path) })
	}

	freshness := "No remote-tracking refs are present. Fetch to establish remote state."
	if remoteSeen {
		freshness = "Remote-tracking refs are local observations and may be stale; fetch to refresh them."
	}
	return Snapshot{Commits: commits, HasMore: hasMore, RemoteRefsSeen: remoteSeen, RemoteFreshness: freshness, SelectedHeadSHA: headSHA, SelectedHeadName: headName}, nil
}

func (r *Reader) loadCommits(ctx context.Context, limit, skip int) ([]Commit, bool, error) {
	const rs = "\x1e"
	const fs = "\x1f"
	format := "%H%x1f%P%x1f%an%x1f%aI%x1f%s%x1e"
	args := []string{"log", "--all", "--topo-order", "--date=iso-strict", "--format=" + format, "--max-count=" + strconv.Itoa(limit+1), "--skip=" + strconv.Itoa(skip)}
	out, err := r.runner.Run(ctx, r.dir, "git", args...)
	if err != nil {
		return nil, false, fmt.Errorf("load git graph commits: %w", err)
	}
	var commits []Commit
	for _, record := range strings.Split(string(out), rs) {
		record = strings.Trim(record, "\r\n")
		if record == "" {
			continue
		}
		fields := strings.SplitN(record, fs, 5)
		if len(fields) != 5 {
			return nil, false, fmt.Errorf("unexpected git graph record")
		}
		authored, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[3]))
		if err != nil {
			return nil, false, fmt.Errorf("parse graph commit time: %w", err)
		}
		commits = append(commits, Commit{SHA: fields[0], Parents: strings.Fields(fields[1]), Author: fields[2], Authored: authored, Subject: fields[4]})
	}
	hasMore := len(commits) > limit
	if hasMore {
		commits = commits[:limit]
	}
	return commits, hasMore, nil
}

func (r *Reader) loadRefs(ctx context.Context) (map[string][]Decoration, bool, error) {
	out, err := r.runner.Run(ctx, r.dir, "git", "for-each-ref", "--format=%(refname)%1f%(objectname)%1f%(symref)", "refs/heads", "refs/remotes", "refs/tags")
	if err != nil {
		return nil, false, fmt.Errorf("load git graph refs: %w", err)
	}
	refs := make(map[string][]Decoration)
	remoteSeen := false
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.SplitN(line, "\x1f", 3)
		if len(fields) < 2 {
			return nil, false, fmt.Errorf("unexpected git ref record")
		}
		refname, sha := fields[0], fields[1]
		if len(fields) == 3 && fields[2] != "" {
			continue
		}
		kind, name, ok := classifyRef(refname)
		if !ok {
			continue
		}
		if kind == RefRemote {
			remoteSeen = true
		}
		refs[sha] = append(refs[sha], Decoration{Kind: kind, Name: name})
	}
	return refs, remoteSeen, nil
}

func (r *Reader) loadHEAD(ctx context.Context) (sha, name string, err error) {
	out, err := r.runner.Run(ctx, r.dir, "git", "rev-parse", "HEAD^{commit}")
	if err != nil {
		return "", "", fmt.Errorf("resolve graph HEAD: %w", err)
	}
	sha = strings.TrimSpace(string(out))
	branch, branchErr := r.runner.Run(ctx, r.dir, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if branchErr == nil {
		name = strings.TrimSpace(string(branch))
	}
	return sha, name, nil
}

func classifyRef(ref string) (RefKind, string, bool) {
	switch {
	case strings.HasPrefix(ref, "refs/heads/"):
		return RefLocal, strings.TrimPrefix(ref, "refs/heads/"), true
	case strings.HasPrefix(ref, "refs/remotes/"):
		return RefRemote, strings.TrimPrefix(ref, "refs/remotes/"), true
	case strings.HasPrefix(ref, "refs/tags/"):
		return RefTag, strings.TrimPrefix(ref, "refs/tags/"), true
	default:
		return "", "", false
	}
}

func sortDecorations(values []Decoration) {
	rank := func(k RefKind) int {
		switch k {
		case RefHEAD:
			return 0
		case RefLocal:
			return 1
		case RefRemote:
			return 2
		case RefTag:
			return 3
		default:
			return 4
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if rank(values[i].Kind) != rank(values[j].Kind) {
			return rank(values[i].Kind) < rank(values[j].Kind)
		}
		return strings.ToLower(values[i].Name) < strings.ToLower(values[j].Name)
	})
}
