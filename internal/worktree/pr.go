package worktree

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// PreparePullRequest fetches a GitHub pull-request head into a private local ref
// and verifies that the fetched object still matches the SHA selected in the UI.
func (m *Manager) PreparePullRequest(ctx context.Context, number int, selectedSHA string) (string, error) {
	if number <= 0 || !validSHA(selectedSHA) {
		return "", fmt.Errorf("invalid pull request identity")
	}
	pullRef := "refs/pull/" + strconv.Itoa(number) + "/head"
	localRef := "refs/gh-tree/pr-" + strconv.Itoa(number)
	if _, err := m.runner.Run(ctx, m.repoRoot, "git", "fetch", "--no-tags", m.remote, "+"+pullRef+":"+localRef); err != nil {
		return "", fmt.Errorf("fetch PR #%d: %w", number, err)
	}
	sha, err := m.revParse(ctx, m.repoRoot, localRef+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve PR #%d: %w", number, err)
	}
	if !strings.EqualFold(sha, selectedSHA) {
		return "", fmt.Errorf("PR #%d changed: selected %s, fetched %s; refresh and retry", number, selectedSHA, sha)
	}
	return localRef, nil
}
