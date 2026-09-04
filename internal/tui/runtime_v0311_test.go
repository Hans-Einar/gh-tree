package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/Hans-Einar/gh-tree/internal/app"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

func TestOverlay311CentersDialogWithoutGrowingScreen(t *testing.T) {
	base := "one\ntwo\nthree"
	dialog := "+----+\n| hi |\n+----+"
	got := overlay311(base, dialog, 80, 30)
	if lines := strings.Split(got, "\n"); len(lines) != 30 {
		t.Fatalf("overlay line count=%d, want 30", len(lines))
	}
	if !strings.Contains(ansi.Strip(got), "| hi |") {
		t.Fatalf("overlay missing dialog: %q", ansi.Strip(got))
	}
}

func TestHeading311PreservesWholeTitleAndMnemonic(t *testing.T) {
	m := V311Model{}
	got := m.heading311("Branch context", 'B', true)
	if plain := ansi.Strip(got); plain != "Branch context" {
		t.Fatalf("heading=%q", plain)
	}
}

func TestAltCAndAltMSelectVisibleBranchSubpanes(t *testing.T) {
	m := WithV311UX(Model{mode: modeBranches})
	m.branchContext = true
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}, Alt: true})
	mc := updated.(V311Model)
	if mc.focus != paneWorktrees || mc.branchSubFocus != 1 {
		t.Fatalf("Alt+C focus=%v sub=%d", mc.focus, mc.branchSubFocus)
	}
	updated, _ = mc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}, Alt: true})
	mm := updated.(V311Model)
	if mm.focus != paneWorktrees || mm.branchSubFocus != 2 {
		t.Fatalf("Alt+M focus=%v sub=%d", mm.focus, mm.branchSubFocus)
	}
}

func TestAltAEnterOpensWorktreeChooserModal(t *testing.T) {
	m := WithV311UX(Model{})
	m.snapshot = app.Snapshot{Worktrees: []worktree.Info{{Path: "C:/repo", Branch: "main", Head: strings.Repeat("a", 40)}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: true})
	ma := updated.(V311Model)
	if ma.focus != paneDetails || ma.activeSubFocus != 0 {
		t.Fatalf("Alt+A focus=%v sub=%d", ma.focus, ma.activeSubFocus)
	}
	updated, _ = ma.Update(tea.KeyMsg{Type: tea.KeyEnter})
	chooser := updated.(V311Model)
	if !chooser.worktreeModal {
		t.Fatal("Enter on Active worktree root did not open chooser modal")
	}
}
