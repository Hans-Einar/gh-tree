package app

import (
	"context"

	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

func (s *Service) WorktreeStashes(ctx context.Context, path string) ([]worktree.Stash, error) {
	m, err := s.requireWorktrees()
	if err != nil {
		return nil, err
	}
	return m.Stashes(ctx, path)
}

func (s *Service) StashApplyRef(ctx context.Context, path, ref string, pop bool) error {
	m, err := s.requireWorktrees()
	if err != nil {
		return err
	}
	return m.StashApplyRef(ctx, path, ref, pop)
}

func (s *Service) StashDrop(ctx context.Context, path, ref string) error {
	m, err := s.requireWorktrees()
	if err != nil {
		return err
	}
	return m.StashDrop(ctx, path, ref)
}

func (s *Service) StashPatch(ctx context.Context, path, ref string) (string, error) {
	m, err := s.requireWorktrees()
	if err != nil {
		return "", err
	}
	return m.StashPatch(ctx, path, ref)
}

func (s *Service) DeployBranchDetached(ctx context.Context, path, branch string) (worktree.Info, error) {
	m, err := s.requireWorktrees()
	if err != nil {
		return worktree.Info{}, err
	}
	return m.CheckoutBranchDetached(ctx, path, branch)
}

func (s *Service) DeployPRDetached(ctx context.Context, path string, pr ghapi.PullRequest) (worktree.Info, error) {
	m, err := s.requireWorktrees()
	if err != nil {
		return worktree.Info{}, err
	}
	ref, err := m.PreparePullRequest(ctx, pr.Number, pr.HeadSHA)
	if err != nil {
		return worktree.Info{}, err
	}
	return m.Checkout(ctx, worktree.CheckoutRequest{Path: path, Revision: ref, Detach: true})
}
