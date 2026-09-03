package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/Hans-Einar/gh-tree/internal/config"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type Snapshot struct {
	PullRequests     []ghapi.PullRequest
	Branches         []ghapi.Branch
	Worktrees        []worktree.Info
	WorktreesEnabled bool
}

type Service struct {
	GitHub    *ghapi.Client
	Worktrees *worktree.Manager
}

func (s *Service) Load(ctx context.Context, repo string) (Snapshot, error) {
	var snapshot Snapshot
	var wait sync.WaitGroup
	errCh := make(chan error, 3)

	wait.Add(2)
	go func() {
		defer wait.Done()
		prs, err := s.GitHub.ListOpenPullRequests(ctx, repo)
		if err != nil {
			errCh <- err
			return
		}
		snapshot.PullRequests = prs
	}()
	go func() {
		defer wait.Done()
		branches, err := s.GitHub.ListBranches(ctx, repo)
		if err != nil {
			errCh <- err
			return
		}
		snapshot.Branches = branches
	}()
	if s.Worktrees != nil {
		snapshot.WorktreesEnabled = true
		wait.Add(1)
		go func() {
			defer wait.Done()
			infos, err := s.Worktrees.List(ctx)
			if err != nil {
				errCh <- err
				return
			}
			snapshot.Worktrees = infos
		}()
	}

	wait.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return Snapshot{}, err
		}
	}
	return snapshot, nil
}

func (s *Service) requireWorktrees() (*worktree.Manager, error) {
	if s.Worktrees == nil {
		return nil, fmt.Errorf("worktree operations are unavailable outside the selected local repository")
	}
	return s.Worktrees, nil
}

func (s *Service) Deploy(ctx context.Context, pr ghapi.PullRequest, target config.WorktreeTarget) (worktree.Deployment, error) {
	manager, err := s.requireWorktrees()
	if err != nil {
		return worktree.Deployment{}, err
	}
	return manager.Deploy(ctx, worktree.DeployRequest{
		PRNumber: pr.Number, HeadSHA: pr.HeadSHA, TargetName: target.Name,
		TargetPath: target.Path, TargetBranch: target.Branch,
	})
}

func (s *Service) WorktreeStatus(ctx context.Context, path string) (worktree.Status, error) {
	manager, err := s.requireWorktrees()
	if err != nil { return worktree.Status{}, err }
	return manager.Status(ctx, path)
}

func (s *Service) CreateWorktree(ctx context.Context, req worktree.CreateRequest) (worktree.Info, error) {
	manager, err := s.requireWorktrees()
	if err != nil { return worktree.Info{}, err }
	return manager.Create(ctx, req)
}

func (s *Service) CheckoutWorktree(ctx context.Context, req worktree.CheckoutRequest) (worktree.Info, error) {
	manager, err := s.requireWorktrees()
	if err != nil { return worktree.Info{}, err }
	return manager.Checkout(ctx, req)
}

func (s *Service) Fetch(ctx context.Context, path string) error {
	manager, err := s.requireWorktrees(); if err != nil { return err }
	return manager.Fetch(ctx, path)
}
func (s *Service) Pull(ctx context.Context, path string) error {
	manager, err := s.requireWorktrees(); if err != nil { return err }
	return manager.Pull(ctx, path)
}
func (s *Service) StageAll(ctx context.Context, path string) error {
	manager, err := s.requireWorktrees(); if err != nil { return err }
	return manager.StageAll(ctx, path)
}
func (s *Service) Commit(ctx context.Context, path, message string) (string, error) {
	manager, err := s.requireWorktrees(); if err != nil { return "", err }
	return manager.Commit(ctx, path, message)
}
func (s *Service) Push(ctx context.Context, path string, setUpstream bool) error {
	manager, err := s.requireWorktrees(); if err != nil { return err }
	return manager.Push(ctx, path, setUpstream)
}
func (s *Service) NewBranch(ctx context.Context, path, name, startPoint string) (worktree.Info, error) {
	manager, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }
	return manager.NewBranch(ctx, path, name, startPoint)
}
func (s *Service) Commits(ctx context.Context, path, revision string, limit, skip int) ([]worktree.Commit, error) {
	manager, err := s.requireWorktrees(); if err != nil { return nil, err }
	return manager.Commits(ctx, path, revision, limit, skip)
}
