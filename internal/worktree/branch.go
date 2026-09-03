package worktree

import (
	"context"
	"fmt"
	"strings"
)

func (m *Manager) ensureBranchStart(ctx context.Context, branch string) (string, bool, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" { return "", false, fmt.Errorf("branch is required") }
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "check-ref-format", "--branch", branch); err != nil {
		return "", false, fmt.Errorf("invalid branch %q: %w", branch, err)
	}
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		return branch, true, nil
	}
	remoteRef := "refs/remotes/" + m.remote + "/" + branch
	refspec := "+refs/heads/" + branch + ":" + remoteRef
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "fetch", "--no-tags", m.remote, refspec); err != nil {
		return "", false, fmt.Errorf("fetch branch %q: %w", branch, err)
	}
	return remoteRef, false, nil
}

func (m *Manager) CreateBranchWorktree(ctx context.Context, path, branch string) (Info, error) {
	start, local, err := m.ensureBranchStart(ctx, branch)
	if err != nil { return Info{}, err }
	if local {
		return m.Create(ctx, CreateRequest{Path: path, StartPoint: start})
	}
	return m.Create(ctx, CreateRequest{Path: path, StartPoint: start, Branch: branch})
}

func (m *Manager) CheckoutBranchWorktree(ctx context.Context, path, branch string) (Info, error) {
	start, local, err := m.ensureBranchStart(ctx, branch)
	if err != nil { return Info{}, err }
	if local {
		return m.Checkout(ctx, CheckoutRequest{Path: path, Revision: start, Branch: branch})
	}
	return m.Checkout(ctx, CheckoutRequest{Path: path, Revision: start, Branch: branch, Create: true})
}
