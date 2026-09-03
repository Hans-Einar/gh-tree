package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

const (
	maxPullRequests = 100
	maxBranches     = 100
)

// PullRequest is the subset of GitHub PR metadata used by the TUI and deployer.
type PullRequest struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	State      string    `json:"state"`
	IsDraft    bool      `json:"isDraft"`
	HeadBranch string    `json:"headRefName"`
	BaseBranch string    `json:"baseRefName"`
	HeadSHA    string    `json:"headRefOid"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Branch is the subset of branch metadata used by the TUI.
type Branch struct {
	Name string
	SHA  string
}

// Client reuses the authenticated gh executable for all GitHub access.
type Client struct {
	runner process.Runner
	dir    string
}

func NewClient(runner process.Runner, dir string) *Client {
	return &Client{runner: runner, dir: dir}
}

// ResolveRepository validates an explicit owner/repo or asks gh to resolve the
// repository associated with the current git directory.
func (c *Client) ResolveRepository(ctx context.Context, explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return validateRepository(explicit)
	}

	out, err := c.runner.Run(ctx, c.dir, "gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil {
		return "", fmt.Errorf("resolve current GitHub repository: %w", err)
	}
	return validateRepository(strings.TrimSpace(string(out)))
}

func validateRepository(repo string) (string, error) {
	repo = strings.TrimSpace(strings.Trim(repo, "/"))
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" ||
		strings.ContainsAny(repo, " \\?#") {
		return "", fmt.Errorf("repository must be in owner/name form, got %q", repo)
	}
	return repo, nil
}

// ListOpenPullRequests returns a bounded list of open PRs.
func (c *Client) ListOpenPullRequests(ctx context.Context, repo string) ([]PullRequest, error) {
	out, err := c.runner.Run(ctx, c.dir, "gh", "pr", "list",
		"--repo", repo,
		"--state", "open",
		"--limit", fmt.Sprintf("%d", maxPullRequests),
		"--json", "number,title,state,isDraft,headRefName,baseRefName,headRefOid,updatedAt",
	)
	if err != nil {
		return nil, fmt.Errorf("load open pull requests: %w", err)
	}
	prs, err := parsePullRequests(out)
	if err != nil {
		return nil, fmt.Errorf("parse pull requests returned by gh: %w", err)
	}
	return prs, nil
}

func parsePullRequests(data []byte) ([]PullRequest, error) {
	var prs []PullRequest
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, err
	}
	for _, pr := range prs {
		if pr.Number <= 0 || pr.HeadBranch == "" || pr.HeadSHA == "" {
			return nil, fmt.Errorf("PR response is missing number, head branch, or head SHA")
		}
	}
	return prs, nil
}

// ListBranches returns the first API page (up to 100 branches). Keeping this
// bounded avoids an unexpectedly expensive exhaustive traversal in large repos.
func (c *Client) ListBranches(ctx context.Context, repo string) ([]Branch, error) {
	validated, err := validateRepository(repo)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(validated, "/")
	endpoint := fmt.Sprintf("repos/%s/%s/branches?per_page=%d",
		url.PathEscape(parts[0]), url.PathEscape(parts[1]), maxBranches)
	out, err := c.runner.Run(ctx, c.dir, "gh", "api", "--method", "GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("load branches: %w", err)
	}
	branches, err := parseBranches(out)
	if err != nil {
		return nil, fmt.Errorf("parse branches returned by gh: %w", err)
	}
	return branches, nil
}

func parseBranches(data []byte) ([]Branch, error) {
	var raw []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	branches := make([]Branch, 0, len(raw))
	for _, branch := range raw {
		if branch.Name == "" || branch.Commit.SHA == "" {
			return nil, fmt.Errorf("branch response is missing name or SHA")
		}
		branches = append(branches, Branch{Name: branch.Name, SHA: branch.Commit.SHA})
	}
	return branches, nil
}
