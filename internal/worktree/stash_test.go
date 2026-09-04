package worktree

import (
	"path/filepath"
	"testing"
	"time"
)

func TestManagedStashMessageRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 4, 20, 30, 0, 0, time.UTC)
	path := filepath.Join("C:", "Users", "Hans Einar", "git", "ponsse-Concept1")
	message := ManagedStashMessage(path, "steering/issue-59-concept1-uibox", "0123456789abcdef", now)
	item := Stash{Subject: "On steering/issue-59-concept1-uibox: " + message}
	parseManagedStash(&item)
	if !item.Managed {
		t.Fatal("expected managed stash metadata")
	}
	if item.OriginWorktree != filepath.Clean(path) {
		t.Fatalf("origin worktree = %q, want %q", item.OriginWorktree, filepath.Clean(path))
	}
	if item.OriginBranch != "steering/issue-59-concept1-uibox" {
		t.Fatalf("origin branch = %q", item.OriginBranch)
	}
	if item.OriginHead != "0123456789abcdef" {
		t.Fatalf("origin head = %q", item.OriginHead)
	}
	if !item.Created.Equal(now) {
		t.Fatalf("created = %v, want %v", item.Created, now)
	}
}

func TestValidateStashRef(t *testing.T) {
	for _, ref := range []string{"stash@{0}", "stash@{12}"} {
		if err := validateStashRef(ref); err != nil {
			t.Fatalf("%s should be valid: %v", ref, err)
		}
	}
	for _, ref := range []string{"stash", "HEAD", "stash@{-1}", "stash@{x}", "stash@{}"} {
		if err := validateStashRef(ref); err == nil {
			t.Fatalf("%s should be invalid", ref)
		}
	}
}
