package worktree

import "context"

func (m *Manager) CommitsForPullRequest(ctx context.Context, number int, sha string, limit, skip int) ([]Commit, error) {
	ref, err := m.PreparePullRequest(ctx, number, sha)
	if err != nil { return nil, err }
	return m.Commits(ctx, m.repoRoot, ref, limit, skip)
}

func (m *Manager) CommitsForBranch(ctx context.Context, branch string, limit, skip int) ([]Commit, error) {
	ref, _, err := m.ensureBranchStart(ctx, branch)
	if err != nil { return nil, err }
	return m.Commits(ctx, m.repoRoot, ref, limit, skip)
}
