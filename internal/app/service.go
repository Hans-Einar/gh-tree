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

func (s *Service) Deploy(ctx context.Context, pr ghapi.PullRequest, target config.WorktreeTarget) (worktree.Deployment, error) {
	if s.Worktrees == nil {
		return worktree.Deployment{}, fmt.Errorf("worktree deployment is unavailable outside the selected local repository")
	}
	return s.Worktrees.Deploy(ctx, worktree.DeployRequest{
		PRNumber:     pr.Number,
		HeadSHA:      pr.HeadSHA,
		TargetName:   target.Name,
		TargetPath:   target.Path,
		TargetBranch: target.Branch,
	})
}
