package worktree

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Stash describes one repository-local Git stash. Git stashes are shared by
// all worktrees in a repository; Origin* fields are populated when gh-tree
// created the stash with its structured metadata message.
type Stash struct {
	Ref            string
	SHA            string
	Subject        string
	Files          int
	Managed        bool
	OriginWorktree string
	OriginBranch   string
	OriginHead     string
	Created        time.Time
}

func (m *Manager) Stashes(ctx context.Context, path string) ([]Stash, error) {
	if _, err := m.Status(ctx, path); err != nil {
		return nil, err
	}
	out, err := m.runner.Run(ctx, path, "git", "stash", "list", "--format=%gd%x00%H%x00%gs%x00")
	if err != nil {
		return nil, fmt.Errorf("list stashes: %w", err)
	}
	fields := strings.Split(string(out), "\x00")
	items := make([]Stash, 0, len(fields)/3)
	for i := 0; i+2 < len(fields); i += 3 {
		ref := strings.TrimSpace(fields[i])
		if ref == "" {
			continue
		}
		item := Stash{Ref: ref, SHA: strings.TrimSpace(fields[i+1]), Subject: strings.TrimSpace(fields[i+2])}
		parseManagedStash(&item)
		if names, e := m.runner.Run(ctx, path, "git", "stash", "show", "--name-only", "--include-untracked", ref); e == nil {
			for _, name := range strings.Split(strings.ReplaceAll(string(names), "\r\n", "\n"), "\n") {
				if strings.TrimSpace(name) != "" {
					item.Files++
				}
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func ManagedStashMessage(worktreePath, branch, head string, now time.Time) string {
	q := url.Values{}
	q.Set("worktree", filepath.Clean(worktreePath))
	q.Set("branch", strings.TrimSpace(branch))
	q.Set("head", strings.TrimSpace(head))
	q.Set("time", now.UTC().Format(time.RFC3339))
	return "gh-tree?" + q.Encode()
}

func parseManagedStash(item *Stash) {
	idx := strings.Index(item.Subject, "gh-tree?")
	if idx < 0 {
		return
	}
	raw := item.Subject[idx+len("gh-tree?"):]
	values, err := url.ParseQuery(raw)
	if err != nil {
		return
	}
	item.Managed = true
	item.OriginWorktree = values.Get("worktree")
	item.OriginBranch = values.Get("branch")
	item.OriginHead = values.Get("head")
	if t, err := time.Parse(time.RFC3339, values.Get("time")); err == nil {
		item.Created = t
	}
}

func (m *Manager) StashPatch(ctx context.Context, path, ref string) (string, error) {
	if err := validateStashRef(ref); err != nil {
		return "", err
	}
	out, err := m.runner.Run(ctx, path, "git", "stash", "show", "--patch", "--stat", "--include-untracked", ref)
	if err != nil {
		return "", fmt.Errorf("show stash %s: %w", ref, err)
	}
	return string(out), nil
}

func (m *Manager) StashApplyRef(ctx context.Context, path, ref string, pop bool) error {
	if err := validateStashRef(ref); err != nil {
		return err
	}
	return m.StashApply(ctx, path, ref, pop)
}

func (m *Manager) StashDrop(ctx context.Context, path, ref string) error {
	if err := validateStashRef(ref); err != nil {
		return err
	}
	if _, err := m.runner.Run(ctx, path, "git", "stash", "drop", ref); err != nil {
		return fmt.Errorf("drop stash %s: %w", ref, err)
	}
	return nil
}

func validateStashRef(ref string) error {
	ref = strings.TrimSpace(ref)
	if !strings.HasPrefix(ref, "stash@{") || !strings.HasSuffix(ref, "}") {
		return fmt.Errorf("invalid stash ref %q", ref)
	}
	n := strings.TrimSuffix(strings.TrimPrefix(ref, "stash@{"), "}")
	if _, err := strconv.Atoi(n); err != nil || strings.HasPrefix(n, "-") {
		return fmt.Errorf("invalid stash ref %q", ref)
	}
	return nil
}
