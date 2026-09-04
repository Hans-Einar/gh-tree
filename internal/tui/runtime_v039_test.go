package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
)

func TestV039CreateWorktreeReviewUsesEToEdit(t *testing.T) {
	m := RuntimeModel{Model: Model{dialog: dialogCreateWorktree, inputField: -1, inputA: `C:\repo-pr-60`, inputB: "gh-tree/pr-60"}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	runtime, ok := updated.(RuntimeModel)
	if !ok {
		t.Fatalf("model type=%T", updated)
	}
	if runtime.inputField != 0 {
		t.Fatalf("inputField=%d want 0", runtime.inputField)
	}
	if runtime.inputA != `C:\repo-pr-60` {
		t.Fatalf("suggested path changed: %q", runtime.inputA)
	}
}

func TestV039CreateWorktreeReplacesLowerPane(t *testing.T) {
	m := RuntimeModel{Model: Model{
		repo:       "Hans-Einar/ponsse",
		width:      120,
		height:     32,
		dialog:     dialogCreateWorktree,
		inputField: -1,
		inputA:     `C:\Users\Hans Einar\git\ponsse-pr-60`,
		inputB:     "gh-tree/pr-60",
		pendingPR:  ghapi.PullRequest{Number: 60, HeadBranch: "steering/Concept1/ui-box"},
	}}
	view := m.View()
	for _, want := range []string{"Create worktree", "suggested path/name", "[e] edit path/name"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
	}
	if strings.Contains(view, "\n\n╭") {
		t.Fatalf("create dialog was appended below cockpit: %q", view)
	}
}

func TestV039WrapperTurnsNewCreateDialogIntoReviewState(t *testing.T) {
	inner := Model{dialog: dialogCreateWorktree, inputField: 0, inputA: "suggested"}
	m := WithRuntimeUX(inner)
	// Simulate the transition behavior directly by entering from a non-create
	// state then returning a create-dialog Model through the stable model API.
	m.dialog = dialogNone
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if runtime, ok := updated.(RuntimeModel); ok && runtime.dialog == dialogCreateWorktree && runtime.inputField != -1 {
		t.Fatalf("create dialog inputField=%d want review state -1", runtime.inputField)
	}
}
