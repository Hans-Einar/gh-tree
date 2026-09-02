package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/app"
	"github.com/Hans-Einar/gh-tree/internal/config"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type fakeBackend struct {
	snapshot   app.Snapshot
	deployment worktree.Deployment
	deployErr  error
	deployedPR ghapi.PullRequest
	target     config.WorktreeTarget
}

func (f *fakeBackend) Load(context.Context, string) (app.Snapshot, error) {
	return f.snapshot, nil
}

func (f *fakeBackend) Deploy(_ context.Context, pr ghapi.PullRequest, target config.WorktreeTarget) (worktree.Deployment, error) {
	f.deployedPR = pr
	f.target = target
	return f.deployment, f.deployErr
}

func testSnapshot() app.Snapshot {
	return app.Snapshot{
		PullRequests: []ghapi.PullRequest{
			{Number: 60, Title: "UIBox", State: "OPEN", HeadBranch: "steering/Concept1/ui-box", BaseBranch: "main", HeadSHA: strings.Repeat("a", 40)},
			{Number: 61, Title: "Simulator", State: "OPEN", HeadBranch: "codex/MVP1/simulator/slc004", BaseBranch: "main", HeadSHA: strings.Repeat("b", 40)},
		},
		Branches: []ghapi.Branch{
			{Name: "feature/Geometry/overlay", SHA: strings.Repeat("c", 40)},
		},
		WorktreesEnabled: true,
	}
}

func TestModelRestoresAndNavigatesNamespace(t *testing.T) {
	t.Parallel()
	var saved []string
	backend := &fakeBackend{}
	model := NewModel(
		"Hans-Einar/ponsse",
		backend,
		config.DefaultStripPrefixes,
		nil,
		"/config.json",
		"Concept1",
		func(folder string) error { saved = append(saved, folder); return nil },
	)
	model = updateModel(t, model, snapshotMsg{snapshot: testSnapshot()})
	if model.folder != "Concept1" || len(model.entries) != 1 || model.entries[0].ID != "pr:60" {
		t.Fatalf("restored model: folder=%q entries=%#v", model.folder, model.entries)
	}

	model = updateModel(t, model, keyMsg(tea.KeyBackspace, ""))
	if model.folder != "" || len(model.entries) != 2 || len(saved) == 0 || saved[len(saved)-1] != "" {
		t.Fatalf("parent navigation: folder=%q entries=%#v saved=%v", model.folder, model.entries, saved)
	}
	model = updateModel(t, model, keyMsg(tea.KeyEnter, ""))
	if model.folder != "Concept1" {
		t.Fatalf("Enter did not open selected folder: %q", model.folder)
	}
}

func TestModelModeAndSearchTransitions(t *testing.T) {
	t.Parallel()
	model := NewModel("Hans-Einar/gh-tree", &fakeBackend{}, config.DefaultStripPrefixes, nil, "/config.json", "", nil)
	model = updateModel(t, model, snapshotMsg{snapshot: testSnapshot()})
	model = updateModel(t, model, runeKey("b"))
	if model.mode != modeBranches || len(model.entries) != 1 || model.entries[0].Path != "Geometry" {
		t.Fatalf("branch mode = %v entries=%#v", model.mode, model.entries)
	}
	model = updateModel(t, model, runeKey("/"))
	model = updateModel(t, model, runeKey("overlay"))
	if !model.searching || model.query != "overlay" || len(model.entries) != 1 {
		t.Fatalf("search state: searching=%v query=%q entries=%#v", model.searching, model.query, model.entries)
	}
	model = updateModel(t, model, keyMsg(tea.KeyEnter, ""))
	if model.searching {
		t.Fatal("Enter should accept the filter")
	}
}

func TestModelRequiresConfirmationAndReportsExactDeployment(t *testing.T) {
	t.Parallel()
	sha := strings.Repeat("a", 40)
	target := config.WorktreeTarget{Name: "Concept1", Path: "/tmp/ponsse-Concept1", Branch: "local/concept1-test"}
	backend := &fakeBackend{deployment: worktree.Deployment{
		TargetName: target.Name,
		Path:       target.Path,
		Branch:     target.Branch,
		PRNumber:   60,
		SHA:        sha,
	}}
	model := NewModel("Hans-Einar/ponsse", backend, config.DefaultStripPrefixes, []config.WorktreeTarget{target}, "/config.json", "Concept1", nil)
	model = updateModel(t, model, snapshotMsg{snapshot: testSnapshot()})
	model = updateModel(t, model, runeKey("w"))
	if model.dialog != dialogConfirm || model.deploying {
		t.Fatalf("deployment was not gated by confirmation: dialog=%v deploying=%v", model.dialog, model.deploying)
	}

	updated, command := model.Update(runeKey("y"))
	model = updated.(Model)
	if !model.deploying || command == nil {
		t.Fatal("confirmation did not start deployment")
	}
	model = updateModel(t, model, command())
	if model.deploying || !strings.Contains(model.status, sha) || backend.deployedPR.Number != 60 {
		t.Fatalf("deployment result: status=%q PR=%#v", model.status, backend.deployedPR)
	}
}

func TestViewNeverEllipsizesExactSHA(t *testing.T) {
	t.Parallel()
	sha := strings.Repeat("a", 40)
	model := NewModel("Hans-Einar/ponsse", &fakeBackend{}, config.DefaultStripPrefixes, nil, "/config.json", "Concept1", nil)
	model = updateModel(t, model, snapshotMsg{snapshot: testSnapshot()})
	model.width = 80
	model.height = 24
	view := model.View()
	if !strings.Contains(view, sha) {
		t.Fatalf("full SHA is missing from detail view: %q", view)
	}
}

func updateModel(t *testing.T, model Model, message tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(message)
	return updated.(Model)
}

func runeKey(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func keyMsg(keyType tea.KeyType, alt string) tea.KeyMsg {
	return tea.KeyMsg{Type: keyType, Alt: alt != ""}
}
