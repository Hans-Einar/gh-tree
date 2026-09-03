package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

type Info struct {
	Path     string
	Head     string
	Branch   string
	Detached bool
	Primary  bool
	Current  bool
}

type DeployRequest struct {
	PRNumber     int
	HeadSHA      string
	TargetName   string
	TargetPath   string
	TargetBranch string
}

type Deployment struct {
	TargetName string
	Path       string
	Branch     string
	PRNumber   int
	SHA        string
}

type Manager struct {
	runner   process.Runner
	repoRoot string
	remote   string
}

func NewManager(runner process.Runner, repoRoot string) *Manager {
	return &Manager{runner: runner, repoRoot: repoRoot, remote: "origin"}
}

func FindRepositoryRoot(ctx context.Context, runner process.Runner, dir string) (string, error) {
	out, err := runner.Run(ctx, dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("git returned an empty repository root")
	}
	return filepath.Abs(root)
}

func (m *Manager) List(ctx context.Context) ([]Info, error) {
	out, err := m.runner.Run(ctx, m.repoRoot, "git", "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, fmt.Errorf("list git worktrees: %w", err)
	}
	infos, err := parsePorcelain(string(out))
	if err != nil {
		return nil, err
	}
	for i := range infos {
		infos[i].Primary = i == 0
		infos[i].Current = samePath(infos[i].Path, m.repoRoot)
	}
	return infos, nil
}

func parsePorcelain(output string) ([]Info, error) {
	var infos []Info
	var current *Info
	for _, field := range strings.Split(output, "\x00") {
		if field == "" {
			if current != nil {
				infos = append(infos, *current)
				current = nil
			}
			continue
		}
		key, value, _ := strings.Cut(field, " ")
		switch key {
		case "worktree":
			if current != nil {
				infos = append(infos, *current)
			}
			current = &Info{Path: value}
		case "HEAD":
			if current != nil {
				current.Head = value
			}
		case "branch":
			if current != nil {
				current.Branch = strings.TrimPrefix(value, "refs/heads/")
			}
		case "detached":
			if current != nil {
				current.Detached = true
			}
		}
	}
	if current != nil {
		infos = append(infos, *current)
	}
	for _, info := range infos {
		if info.Path == "" || info.Head == "" {
			return nil, fmt.Errorf("git worktree output omitted path or HEAD")
		}
	}
	return infos, nil
}

// Deploy fetches the exact GitHub PR ref and moves only the configured local
// test branch in the configured non-primary worktree.
func (m *Manager) Deploy(ctx context.Context, request DeployRequest) (Deployment, error) {
	if request.PRNumber <= 0 {
		return Deployment{}, fmt.Errorf("invalid PR number %d", request.PRNumber)
	}
	if !validSHA(request.HeadSHA) {
		return Deployment{}, fmt.Errorf("invalid PR head SHA %q", request.HeadSHA)
	}
	if request.TargetPath == "" || !filepath.IsAbs(request.TargetPath) {
		return Deployment{}, fmt.Errorf("configured worktree path must be absolute: %q", request.TargetPath)
	}
	if strings.TrimSpace(request.TargetBranch) == "" {
		return Deployment{}, fmt.Errorf("configured local test branch is empty")
	}
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "check-ref-format", "--branch", request.TargetBranch); err != nil {
		return Deployment{}, fmt.Errorf("invalid configured local test branch %q: %w", request.TargetBranch, err)
	}

	infos, err := m.List(ctx)
	if err != nil {
		return Deployment{}, err
	}
	var target *Info
	for i := range infos {
		if samePath(infos[i].Path, request.TargetPath) {
			target = &infos[i]
			break
		}
	}
	if target == nil {
		return Deployment{}, fmt.Errorf("configured path is not a registered worktree for this repository: %s", request.TargetPath)
	}
	stat, err := os.Stat(target.Path)
	if err != nil {
		return Deployment{}, fmt.Errorf("configured worktree path is unavailable %s: %w", target.Path, err)
	}
	if !stat.IsDir() {
		return Deployment{}, fmt.Errorf("configured worktree path is not a directory: %s", target.Path)
	}
	if target.Primary {
		return Deployment{}, fmt.Errorf("refusing to deploy into the primary worktree: %s", target.Path)
	}
	if target.Current {
		return Deployment{}, fmt.Errorf("refusing to deploy into the current worktree: %s", target.Path)
	}
	for _, info := range infos {
		if samePath(info.Path, target.Path) {
			continue
		}
		if sameBranch(info.Branch, request.TargetBranch) {
			return Deployment{}, fmt.Errorf("local test branch %q is already checked out in %s", request.TargetBranch, info.Path)
		}
	}
	if err := m.requireClean(ctx, target.Path); err != nil {
		return Deployment{}, err
	}
	if target.Detached {
		if err := m.requireReachableDetachedHEAD(ctx, target); err != nil {
			return Deployment{}, err
		}
	}

	pullRef := "refs/pull/" + strconv.Itoa(request.PRNumber) + "/head"
	localRef := "refs/gh-tree/pr-" + strconv.Itoa(request.PRNumber)
	refspec := "+" + pullRef + ":" + localRef
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "fetch", "--no-tags", m.remote, refspec); err != nil {
		return Deployment{}, fmt.Errorf("fetch PR #%d from %s: %w", request.PRNumber, m.remote, err)
	}
	fetchedSHA, err := m.revParse(ctx, m.repoRoot, localRef+"^{commit}")
	if err != nil {
		return Deployment{}, fmt.Errorf("resolve fetched PR #%d: %w", request.PRNumber, err)
	}
	if !strings.EqualFold(fetchedSHA, request.HeadSHA) {
		return Deployment{}, fmt.Errorf("PR #%d changed during deployment: selected %s, fetched %s; refresh and try again",
			request.PRNumber, request.HeadSHA, fetchedSHA)
	}
	// Recheck after the network operation so changes made while fetching are not lost.
	if err := m.requireClean(ctx, target.Path); err != nil {
		return Deployment{}, err
	}

	if sameBranch(target.Branch, request.TargetBranch) {
		// --keep adds a final Git-level guard against a file changing in the
		// narrow interval after the second cleanliness check.
		if _, err := m.runner.Run(ctx, target.Path, "git", "reset", "--keep", fetchedSHA); err != nil {
			return Deployment{}, fmt.Errorf("move local test branch %q: %w", request.TargetBranch, err)
		}
	} else {
		if _, err := m.runner.Run(ctx, target.Path, "git", "switch", "-C", request.TargetBranch, fetchedSHA); err != nil {
			return Deployment{}, fmt.Errorf("check out local test branch %q in %s: %w", request.TargetBranch, target.Path, err)
		}
	}

	resultSHA, err := m.revParse(ctx, target.Path, "HEAD^{commit}")
	if err != nil {
		return Deployment{}, fmt.Errorf("verify deployed HEAD: %w", err)
	}
	if !strings.EqualFold(resultSHA, request.HeadSHA) {
		return Deployment{}, fmt.Errorf("deployment verification failed: expected %s, got %s", request.HeadSHA, resultSHA)
	}
	out, err := m.runner.Run(ctx, target.Path, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return Deployment{}, fmt.Errorf("verify deployed branch: %w", err)
	}
	resultBranch := strings.TrimSpace(string(out))
	if !sameBranch(resultBranch, request.TargetBranch) {
		return Deployment{}, fmt.Errorf("deployment verification failed: expected branch %q, got %q", request.TargetBranch, resultBranch)
	}
	if err := m.requireClean(ctx, target.Path); err != nil {
		return Deployment{}, fmt.Errorf("deployed worktree did not remain clean: %w", err)
	}
	return Deployment{
		TargetName: request.TargetName,
		Path:       target.Path,
		Branch:     resultBranch,
		PRNumber:   request.PRNumber,
		SHA:        resultSHA,
	}, nil
}

func (m *Manager) requireReachableDetachedHEAD(ctx context.Context, target *Info) error {
	out, err := m.runner.Run(ctx, target.Path, "git", "for-each-ref",
		"--format=%(refname)", "--contains", target.Head, "refs/heads", "refs/remotes")
	if err != nil {
		return fmt.Errorf("inspect detached worktree HEAD %s: %w", target.Head, err)
	}
	if strings.TrimSpace(string(out)) == "" {
		return fmt.Errorf("detached worktree HEAD %s is not reachable from a local or remote branch; create a branch for it before deployment", target.Head)
	}
	return nil
}

func (m *Manager) requireClean(ctx context.Context, path string) error {
	out, err := m.runner.Run(ctx, path, "git", "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return fmt.Errorf("inspect worktree %s: %w", path, err)
	}
	if strings.TrimSpace(string(out)) != "" {
		return fmt.Errorf("worktree is dirty; refusing to discard changes: %s", path)
	}
	return nil
}

func (m *Manager) revParse(ctx context.Context, dir, revision string) (string, error) {
	out, err := m.runner.Run(ctx, dir, "git", "rev-parse", "--verify", revision)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func validSHA(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, char := range value {
		if !unicode.Is(unicode.ASCII_Hex_Digit, char) {
			return false
		}
	}
	return true
}

func samePath(left, right string) bool {
	left = canonicalPath(left)
	right = canonicalPath(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func sameBranch(left, right string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func canonicalPath(value string) string {
	absolute, err := filepath.Abs(value)
	if err == nil {
		value = absolute
	}
	if evaluated, err := filepath.EvalSymlinks(value); err == nil {
		value = evaluated
	}
	return filepath.Clean(value)
}
