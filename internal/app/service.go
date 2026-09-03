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

type Service struct { GitHub *ghapi.Client; Worktrees *worktree.Manager }

func (s *Service) Load(ctx context.Context, repo string) (Snapshot, error) {
	var snapshot Snapshot
	var wait sync.WaitGroup
	errCh := make(chan error, 3)
	wait.Add(2)
	go func() { defer wait.Done(); prs, err := s.GitHub.ListOpenPullRequests(ctx, repo); if err != nil { errCh <- err; return }; snapshot.PullRequests = prs }()
	go func() { defer wait.Done(); branches, err := s.GitHub.ListBranches(ctx, repo); if err != nil { errCh <- err; return }; snapshot.Branches = branches }()
	if s.Worktrees != nil {
		snapshot.WorktreesEnabled = true
		wait.Add(1)
		go func() { defer wait.Done(); infos, err := s.Worktrees.List(ctx); if err != nil { errCh <- err; return }; snapshot.Worktrees = infos }()
	}
	wait.Wait(); close(errCh)
	for err := range errCh { if err != nil { return Snapshot{}, err } }
	return snapshot, nil
}

func (s *Service) requireWorktrees() (*worktree.Manager, error) {
	if s.Worktrees == nil { return nil, fmt.Errorf("worktree operations are unavailable outside the selected local repository") }
	return s.Worktrees, nil
}

func (s *Service) Deploy(ctx context.Context, pr ghapi.PullRequest, target config.WorktreeTarget) (worktree.Deployment, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Deployment{}, err }; return m.Deploy(ctx, worktree.DeployRequest{PRNumber: pr.Number, HeadSHA: pr.HeadSHA, TargetName: target.Name, TargetPath: target.Path, TargetBranch: target.Branch}) }
func (s *Service) WorktreeStatus(ctx context.Context, path string) (worktree.Status, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Status{}, err }; return m.Status(ctx, path) }
func (s *Service) CreateWorktree(ctx context.Context, req worktree.CreateRequest) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; return m.Create(ctx, req) }
func (s *Service) CreatePRWorktree(ctx context.Context, pr ghapi.PullRequest, path, branch string) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; ref, err := m.PreparePullRequest(ctx, pr.Number, pr.HeadSHA); if err != nil { return worktree.Info{}, err }; return m.Create(ctx, worktree.CreateRequest{Path: path, StartPoint: ref, Branch: branch}) }
func (s *Service) CreateBranchWorktree(ctx context.Context, path, branch string) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; return m.CreateBranchWorktree(ctx, path, branch) }
func (s *Service) CheckoutWorktree(ctx context.Context, req worktree.CheckoutRequest) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; return m.Checkout(ctx, req) }
func (s *Service) CheckoutPRWorktree(ctx context.Context, pr ghapi.PullRequest, path, branch string) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; ref, err := m.PreparePullRequest(ctx, pr.Number, pr.HeadSHA); if err != nil { return worktree.Info{}, err }; return m.Checkout(ctx, worktree.CheckoutRequest{Path: path, Revision: ref, Branch: branch, Create: true}) }
func (s *Service) CheckoutBranchWorktree(ctx context.Context, path, branch string) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; return m.CheckoutBranchWorktree(ctx, path, branch) }
func (s *Service) Fetch(ctx context.Context, path string) error { m, err := s.requireWorktrees(); if err != nil { return err }; return m.Fetch(ctx, path) }
func (s *Service) Pull(ctx context.Context, path string) error { m, err := s.requireWorktrees(); if err != nil { return err }; return m.Pull(ctx, path) }
func (s *Service) StageAll(ctx context.Context, path string) error { m, err := s.requireWorktrees(); if err != nil { return err }; return m.StageAll(ctx, path) }
func (s *Service) Commit(ctx context.Context, path, message string) (string, error) { m, err := s.requireWorktrees(); if err != nil { return "", err }; return m.Commit(ctx, path, message) }
func (s *Service) Push(ctx context.Context, path string, setUpstream bool) error { m, err := s.requireWorktrees(); if err != nil { return err }; return m.Push(ctx, path, setUpstream) }
func (s *Service) NewBranch(ctx context.Context, path, name, startPoint string) (worktree.Info, error) { m, err := s.requireWorktrees(); if err != nil { return worktree.Info{}, err }; return m.NewBranch(ctx, path, name, startPoint) }
func (s *Service) Commits(ctx context.Context, path, revision string, limit, skip int) ([]worktree.Commit, error) { m, err := s.requireWorktrees(); if err != nil { return nil, err }; return m.Commits(ctx, path, revision, limit, skip) }
func (s *Service) CommitsForPullRequest(ctx context.Context, pr ghapi.PullRequest, limit, skip int) ([]worktree.Commit, error) { m, err := s.requireWorktrees(); if err != nil { return nil, err }; return m.CommitsForPullRequest(ctx, pr.Number, pr.HeadSHA, limit, skip) }
func (s *Service) CommitsForBranch(ctx context.Context, branch string, limit, skip int) ([]worktree.Commit, error) { m, err := s.requireWorktrees(); if err != nil { return nil, err }; return m.CommitsForBranch(ctx, branch, limit, skip) }
func (s *Service) CreatePullRequest(ctx context.Context, repo, head, base, title, body string, draft bool) (string, error) { return s.GitHub.CreatePullRequest(ctx, repo, head, base, title, body, draft) }
