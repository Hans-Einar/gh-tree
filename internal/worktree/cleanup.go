package worktree

import (
	"context"
	"fmt"
)

// RestorePaths discards working-tree changes for tracked paths only. It refuses
// untracked and conflicted paths so cleanup cannot silently delete data or make
// conflict-resolution choices for the user.
func (m *Manager) RestorePaths(ctx context.Context, path string, paths ...string) error {
	safe, err := safeRelativePaths(paths)
	if err != nil {
		return err
	}
	changes, err := m.Changes(ctx, path)
	if err != nil {
		return err
	}
	byPath := make(map[string]Change, len(changes))
	for _, change := range changes {
		byPath[change.Path] = change
	}
	for _, name := range safe {
		change, ok := byPath[name]
		if !ok {
			return fmt.Errorf("path is not currently changed: %q", name)
		}
		if change.Untracked {
			return fmt.Errorf("refusing to delete untracked path %q", name)
		}
		if change.Conflicted {
			return fmt.Errorf("refusing to restore conflicted path %q", name)
		}
	}
	args := []string{"restore", "--worktree", "--"}
	args = append(args, safe...)
	if _, err := m.runner.Run(ctx, path, "git", args...); err != nil {
		return fmt.Errorf("restore tracked files: %w", err)
	}
	return nil
}
