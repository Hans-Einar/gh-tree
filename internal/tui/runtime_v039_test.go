package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/launch"
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

func TestV039MakeStackDoesNotCrossProjectRoots(t *testing.T) {
	m := RuntimeModel{Model: Model{
		dialog: dialogLaunchPicker,
		launchCandidates: []launch.Candidate{
			{Provider: "make", Dir: "frontend", Targets: []string{"clean"}},
			{Provider: "make", Dir: "backend", Targets: []string{"all"}},
		},
		launchCursor:   1,
		launchSelected: []int{0},
	}}
	m.keepMakeStackInsideProjectRoot()
	if len(m.launchSelected) != 0 {
		t.Fatalf("selection=%v want cleared", m.launchSelected)
	}
	if !strings.Contains(m.status, "backend") {
		t.Fatalf("status=%q", m.status)
	}
}

func TestV039LaunchPickerShowsProjectRoot(t *testing.T) {
	m := RuntimeModel{Model: Model{
		dialog: dialogLaunchPicker,
		launchCandidates: []launch.Candidate{
			{Provider: "npm", Dir: "Concept1", Script: "dev:wan"},
		},
	}}
	view := m.renderNestedLaunchPickerDialog()
	if !strings.Contains(view, "Concept1") || !strings.Contains(view, "dev:wan") {
		t.Fatalf("view=%q", view)
	}
}
