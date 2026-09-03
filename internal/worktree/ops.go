package worktree

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Status is the live state of one registered worktree.
type Status struct {
	Info       Info
	Clean      bool
	Staged     int
	Modified   int
	Untracked  int
	Conflicted int
	Upstream   string
	Ahead      int
	Behind     int
}

// Commit is one commit shown by the bounded history browser.
type Commit struct {
	SHA       string
	Parents   []string
	Author    string
	Email     string
	Authored  time.Time
	Subject   string
	Message   string
}

type CreateRequest struct {
	Path       string
	StartPoint string
	Branch     string
	Detach     bool
}

type CheckoutRequest struct {
	Path       string
	Revision   string
	Branch     string
	Detach     bool
	Create     bool
}

func (m *Manager) Status(ctx context.Context, path string) (Status, error) {
	infos, err := m.List(ctx)
	if err != nil {
		return Status{}, err
	}
	var info *Info
	for i := range infos {
		if samePath(infos[i].Path, path) {
			info = &infos[i]
			break
		}
	}
	if info == nil {
		return Status{}, fmt.Errorf("path is not a registered worktree: %s", path)
	}
	out, err := m.runner.Run(ctx, info.Path, "git", "status", "--porcelain=v1", "--branch", "--untracked-files=all")
	if err != nil {
		return Status{}, fmt.Errorf("inspect worktree status: %w", err)
	}
	status := Status{Info: *info, Clean: true}
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, "## ") {
			status.Upstream, status.Ahead, status.Behind = parseBranchStatus(line)
			continue
		}
		if len(line) < 2 {
			continue
		}
		status.Clean = false
		x, y := line[0], line[1]
		if x == '?' && y == '?' {
			status.Untracked++
			continue
		}
		if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
			status.Conflicted++
			continue
		}
		if x != ' ' {
			status.Staged++
		}
		if y != ' ' {
			status.Modified++
		}
	}
	return status, nil
}

func parseBranchStatus(line string) (upstream string, ahead, behind int) {
	line = strings.TrimPrefix(line, "## ")
	if i := strings.Index(line, "..."); i >= 0 {
		rest := line[i+3:]
		if j := strings.Index(rest, " ["); j >= 0 {
			upstream = strings.TrimSpace(rest[:j])
			meta := strings.TrimSuffix(strings.TrimPrefix(rest[j+2:], ""), "]")
			for _, part := range strings.Split(meta, ",") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "ahead ") {
				ahead, _ = strconv.Atoi(strings.TrimPrefix(part, "ahead "))
				}
				if strings.HasPrefix(part, "behind ") {
				behind, _ = strconv.Atoi(strings.TrimPrefix(part, "behind "))
				}
			}
		} else {
			upstream = strings.TrimSpace(rest)
		}
	}
	return upstream, ahead, behind
}

// Create creates and registers a new worktree. It never overwrites an existing path.
func (m *Manager) Create(ctx context.Context, req CreateRequest) (Info, error) {
	if req.Path == "" || !filepath.IsAbs(req.Path) {
		return Info{}, fmt.Errorf("worktree path must be absolute")
	}
	if strings.TrimSpace(req.StartPoint) == "" {
		req.StartPoint = "HEAD"
	}
	if _, err := filepath.Abs(req.Path); err != nil {
		return Info{}, err
	}
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "rev-parse", "--verify", req.StartPoint+"^{commit}"); err != nil {
		return Info{}, fmt.Errorf("invalid start point %q: %w", req.StartPoint, err)
	}
	args := []string{"worktree", "add"}
	if req.Detach {
		args = append(args, "--detach", req.Path, req.StartPoint)
	} else if strings.TrimSpace(req.Branch) != "" {
		if _, err := m.runner.Run(ctx, m.repoRoot, "git", "check-ref-format", "--branch", req.Branch); err != nil {
			return Info{}, fmt.Errorf("invalid branch %q: %w", req.Branch, err)
		}
		args = append(args, "-b", req.Branch, req.Path, req.StartPoint)
	} else {
		args = append(args, req.Path, req.StartPoint)
	}
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", args...); err != nil {
		return Info{}, fmt.Errorf("create worktree: %w", err)
	}
	infos, err := m.List(ctx)
	if err != nil {
		return Info{}, err
	}
	for _, info := range infos {
		if samePath(info.Path, req.Path) {
			return info, nil
		}
	}
	return Info{}, fmt.Errorf("created worktree was not registered by git: %s", req.Path)
}

// Checkout safely retargets a clean, non-primary worktree.
func (m *Manager) Checkout(ctx context.Context, req CheckoutRequest) (Info, error) {
	status, err := m.Status(ctx, req.Path)
	if err != nil {
		return Info{}, err
	}
	if status.Info.Primary {
		return Info{}, fmt.Errorf("refusing to retarget the primary worktree")
	}
	if !status.Clean {
		return Info{}, fmt.Errorf("worktree is dirty; refusing checkout: %s", req.Path)
	}
	if strings.TrimSpace(req.Revision) == "" {
		return Info{}, fmt.Errorf("revision is required")
	}
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "rev-parse", "--verify", req.Revision+"^{commit}"); err != nil {
		return Info{}, fmt.Errorf("invalid revision %q: %w", req.Revision, err)
	}
	var args []string
	if req.Detach {
		args = []string{"switch", "--detach", req.Revision}
	} else if req.Create {
		if strings.TrimSpace(req.Branch) == "" {
			return Info{}, fmt.Errorf("new branch name is required")
		}
		args = []string{"switch", "-c", req.Branch, req.Revision}
	} else if strings.TrimSpace(req.Branch) != "" {
		args = []string{"switch", req.Branch}
	} else {
		args = []string{"switch", req.Revision}
	}
	if _, err := m.runner.Run(ctx, req.Path, "git", args...); err != nil {
		return Info{}, fmt.Errorf("checkout in worktree: %w", err)
	}
	infos, err := m.List(ctx)
	if err != nil {
		return Info{}, err
	}
	for _, info := range infos {
		if samePath(info.Path, req.Path) {
			return info, nil
		}
	}
	return Info{}, fmt.Errorf("worktree disappeared after checkout")
}

func (m *Manager) Fetch(ctx context.Context, path string) error {
	if _, err := m.runner.Run(ctx, path, "git", "fetch", "--prune", m.remote); err != nil {
		return fmt.Errorf("fetch %s: %w", m.remote, err)
	}
	return nil
}

func (m *Manager) Pull(ctx context.Context, path string) error {
	status, err := m.Status(ctx, path)
	if err != nil {
		return err
	}
	if status.Info.Detached {
		return fmt.Errorf("cannot pull while HEAD is detached")
	}
	if !status.Clean {
		return fmt.Errorf("worktree is dirty; commit or stash before pulling")
	}
	if status.Upstream == "" {
		return fmt.Errorf("branch %q has no upstream", status.Info.Branch)
	}
	if _, err := m.runner.Run(ctx, path, "git", "pull", "--ff-only"); err != nil {
		return fmt.Errorf("pull --ff-only: %w", err)
	}
	return nil
}

func (m *Manager) StageAll(ctx context.Context, path string) error {
	if _, err := m.runner.Run(ctx, path, "git", "add", "--all"); err != nil {
		return fmt.Errorf("stage changes: %w", err)
	}
	return nil
}

func (m *Manager) Commit(ctx context.Context, path, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("commit message is empty")
	}
	if _, err := m.runner.Run(ctx, path, "git", "commit", "-m", message); err != nil {
		return "", fmt.Errorf("commit changes: %w", err)
	}
	return m.revParse(ctx, path, "HEAD^{commit}")
}

func (m *Manager) Push(ctx context.Context, path string, setUpstream bool) error {
	status, err := m.Status(ctx, path)
	if err != nil {
		return err
	}
	if status.Info.Detached || status.Info.Branch == "" {
		return fmt.Errorf("cannot push detached HEAD")
	}
	args := []string{"push"}
	if setUpstream || status.Upstream == "" {
		args = append(args, "-u", m.remote, status.Info.Branch)
	}
	if _, err := m.runner.Run(ctx, path, "git", args...); err != nil {
		return fmt.Errorf("push branch: %w", err)
	}
	return nil
}

func (m *Manager) NewBranch(ctx context.Context, path, name, startPoint string) (Info, error) {
	if strings.TrimSpace(name) == "" {
		return Info{}, fmt.Errorf("branch name is empty")
	}
	if _, err := m.runner.Run(ctx, path, "git", "check-ref-format", "--branch", name); err != nil {
		return Info{}, fmt.Errorf("invalid branch %q: %w", name, err)
	}
	if startPoint == "" {
		startPoint = "HEAD"
	}
	status, err := m.Status(ctx, path)
	if err != nil {
		return Info{}, err
	}
	if !status.Clean {
		return Info{}, fmt.Errorf("worktree is dirty; refusing branch switch")
	}
	if _, err := m.runner.Run(ctx, path, "git", "switch", "-c", name, startPoint); err != nil {
		return Info{}, fmt.Errorf("create branch: %w", err)
	}
	infos, err := m.List(ctx)
	if err != nil {
		return Info{}, err
	}
	for _, info := range infos {
		if samePath(info.Path, path) {
			return info, nil
		}
	}
	return Info{}, fmt.Errorf("worktree disappeared after branch creation")
}

// Commits returns a bounded first page of history for revision.
func (m *Manager) Commits(ctx context.Context, path, revision string, limit, skip int) ([]Commit, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if skip < 0 {
		skip = 0
	}
	if strings.TrimSpace(revision) == "" {
		revision = "HEAD"
	}
	const recordSep = "\x1e"
	const fieldSep = "\x1f"
	format := "%H%x1f%P%x1f%an%x1f%ae%x1f%aI%x1f%s%x1f%B%x1e"
	out, err := m.runner.Run(ctx, path, "git", "log", "--date=iso-strict", "--format="+format,
		"--max-count="+strconv.Itoa(limit), "--skip="+strconv.Itoa(skip), revision)
	if err != nil {
		return nil, fmt.Errorf("load commit history: %w", err)
	}
	var commits []Commit
	for _, record := range strings.Split(string(out), recordSep) {
		record = strings.Trim(record, "\r\n")
		if record == "" {
			continue
		}
		fields := strings.SplitN(record, fieldSep, 7)
		if len(fields) != 7 {
			return nil, fmt.Errorf("unexpected git log record")
		}
		authored, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[4]))
		if err != nil {
			return nil, fmt.Errorf("parse commit time: %w", err)
		}
		parents := strings.Fields(fields[1])
		commits = append(commits, Commit{
			SHA: fields[0], Parents: parents, Author: fields[2], Email: fields[3], Authored: authored,
			Subject: fields[5], Message: strings.TrimSpace(fields[6]),
		})
	}
	return commits, nil
}
