package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/app"
	"github.com/Hans-Einar/gh-tree/internal/config"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/tree"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type Backend interface {
	Load(ctx context.Context, repo string) (app.Snapshot, error)
	Deploy(ctx context.Context, pr ghapi.PullRequest, target config.WorktreeTarget) (worktree.Deployment, error)
}

type mode int

const (
	modePullRequests mode = iota
	modeBranches
)

type dialog int

const (
	dialogNone dialog = iota
	dialogTargetPicker
	dialogConfirm
)

type snapshotMsg struct {
	snapshot app.Snapshot
	err      error
}

type deploymentMsg struct {
	deployment worktree.Deployment
	err        error
}

type Model struct {
	repo          string
	backend       Backend
	prefixes      []string
	targets       []config.WorktreeTarget
	configPath    string
	saveFolder    func(string) error
	snapshot      app.Snapshot
	mode          mode
	folder        string
	query         string
	searching     bool
	entries       []tree.Entry
	paths         []tree.PathItem
	prsByID       map[string]ghapi.PullRequest
	branchesByID  map[string]ghapi.Branch
	cursor        int
	width         int
	height        int
	loading       bool
	status        string
	dialog        dialog
	targetCursor  int
	pendingTarget config.WorktreeTarget
	pendingPR     ghapi.PullRequest
	deploying     bool
}

func NewModel(
	repo string,
	backend Backend,
	prefixes []string,
	targets []config.WorktreeTarget,
	configPath string,
	savedFolder string,
	saveFolder func(string) error,
) Model {
	return Model{
		repo:         repo,
		backend:      backend,
		prefixes:     append([]string(nil), prefixes...),
		targets:      append([]config.WorktreeTarget(nil), targets...),
		configPath:   configPath,
		saveFolder:   saveFolder,
		folder:       strings.Trim(savedFolder, "/"),
		prsByID:      make(map[string]ghapi.PullRequest),
		branchesByID: make(map[string]ghapi.Branch),
		loading:      true,
		status:       "Loading GitHub state…",
	}
}

func (m Model) Init() tea.Cmd { return m.refreshCmd() }

func (m Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		snapshot, err := m.backend.Load(context.Background(), m.repo)
		return snapshotMsg{snapshot: snapshot, err: err}
	}
}

func (m Model) deployCmd(pr ghapi.PullRequest, target config.WorktreeTarget) tea.Cmd {
	return func() tea.Msg {
		deployment, err := m.backend.Deploy(context.Background(), pr, target)
		return deploymentMsg{deployment: deployment, err: err}
	}
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case snapshotMsg:
		m.loading = false
		if msg.err != nil {
			m.status = "Refresh failed: " + msg.err.Error()
			return m, nil
		}
		m.snapshot = msg.snapshot
		oldFolder := m.folder
		m.rebuild()
		if m.folder != oldFolder {
			m.persistFolder()
		}
		m.status = fmt.Sprintf("Loaded %d open PRs, %d branches", len(m.snapshot.PullRequests), len(m.snapshot.Branches))
		return m, nil
	case deploymentMsg:
		m.deploying = false
		m.dialog = dialogNone
		if msg.err != nil {
			m.status = "Deployment failed: " + msg.err.Error()
			return m, nil
		}
		m.status = fmt.Sprintf("✓ %s · PR #%d · %s · %s",
			msg.deployment.TargetName, msg.deployment.PRNumber, msg.deployment.Branch, msg.deployment.SHA)
		for i := range m.snapshot.Worktrees {
			if sameDisplayPath(m.snapshot.Worktrees[i].Path, msg.deployment.Path) {
				m.snapshot.Worktrees[i].Head = msg.deployment.SHA
				m.snapshot.Worktrees[i].Branch = msg.deployment.Branch
				m.snapshot.Worktrees[i].Detached = false
			}
		}
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.deploying {
		return m, nil
	}
	if m.dialog != dialogNone {
		return m.updateDialog(msg)
	}
	if m.searching {
		return m.updateSearch(msg)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor+1 < len(m.entries) {
			m.cursor++
		}
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		if len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
		}
	case "enter":
		if entry, ok := m.currentEntry(); ok {
			if entry.IsFolder {
				m.folder = entry.Path
				m.cursor = 0
				m.rebuild()
				m.persistFolder()
			} else {
				m.status = m.selectionStatus(entry)
			}
		}
	case "backspace":
		if m.folder != "" {
			m.folder = tree.Parent(m.folder)
			m.cursor = 0
			m.rebuild()
			m.persistFolder()
		}
	case "/":
		m.searching = true
		m.status = "Type to filter; Enter accepts, Ctrl+U clears"
	case "r":
		m.loading = true
		m.status = "Refreshing GitHub state…"
		return m, m.refreshCmd()
	case "p":
		if m.mode != modePullRequests {
			m.mode = modePullRequests
			m.cursor = 0
			m.rebuild()
			m.persistFolder()
		}
	case "b":
		if m.mode != modeBranches {
			m.mode = modeBranches
			m.cursor = 0
			m.rebuild()
			m.persistFolder()
		}
	case "w":
		return m.beginDeployment()
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.searching = false
		m.status = "Filter: " + displayQuery(m.query)
	case "ctrl+u":
		m.query = ""
		m.cursor = 0
		m.rebuild()
	case "backspace":
		runes := []rune(m.query)
		if len(runes) > 0 {
			m.query = string(runes[:len(runes)-1])
			m.cursor = 0
			m.rebuild()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.query += string(msg.Runes)
			m.cursor = 0
			m.rebuild()
		}
	}
	return m, nil
}

func (m Model) updateDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.dialog {
	case dialogTargetPicker:
		switch msg.String() {
		case "esc", "q":
			m.dialog = dialogNone
			m.status = "Deployment cancelled"
		case "up", "k":
			if m.targetCursor > 0 {
				m.targetCursor--
			}
		case "down", "j":
			if m.targetCursor+1 < len(m.targets) {
				m.targetCursor++
			}
		case "enter":
			if len(m.targets) > 0 {
				m.pendingTarget = m.targets[m.targetCursor]
				m.dialog = dialogConfirm
			}
		}
	case dialogConfirm:
		switch msg.String() {
		case "y", "Y":
			m.dialog = dialogNone
			m.deploying = true
			m.status = fmt.Sprintf("Deploying PR #%d at %s…", m.pendingPR.Number, m.pendingPR.HeadSHA)
			return m, m.deployCmd(m.pendingPR, m.pendingTarget)
		case "n", "N", "esc", "q":
			m.dialog = dialogNone
			m.status = "Deployment cancelled"
		}
	}
	return m, nil
}

func (m Model) beginDeployment() (tea.Model, tea.Cmd) {
	if m.mode != modePullRequests {
		m.status = "Select a PR before deploying"
		return m, nil
	}
	entry, ok := m.currentEntry()
	if !ok || entry.IsFolder {
		m.status = "Select a PR before deploying"
		return m, nil
	}
	pr, ok := m.prsByID[entry.ID]
	if !ok {
		m.status = "Selected item is not a PR"
		return m, nil
	}
	if len(m.targets) == 0 {
		m.status = "No worktree targets configured; edit " + m.configPath
		return m, nil
	}
	m.pendingPR = pr
	m.targetCursor = 0
	if len(m.targets) == 1 {
		m.pendingTarget = m.targets[0]
		m.dialog = dialogConfirm
	} else {
		m.dialog = dialogTargetPicker
	}
	return m, nil
}

func (m *Model) rebuild() {
	selectedID := ""
	if entry, ok := m.currentEntry(); ok {
		selectedID = entry.ID
	}
	m.paths = m.buildPaths()
	m.folder = tree.ResolveFolder(m.paths, m.folder)
	m.entries = tree.Entries(m.paths, m.folder, m.query)
	m.cursor = 0
	if selectedID != "" {
		for i, entry := range m.entries {
			if entry.ID == selectedID {
				m.cursor = i
				break
			}
		}
	}
	if m.cursor >= len(m.entries) && len(m.entries) > 0 {
		m.cursor = len(m.entries) - 1
	}
}

func (m *Model) buildPaths() []tree.PathItem {
	m.prsByID = make(map[string]ghapi.PullRequest, len(m.snapshot.PullRequests))
	m.branchesByID = make(map[string]ghapi.Branch, len(m.snapshot.Branches))
	if m.mode == modePullRequests {
		paths := make([]tree.PathItem, 0, len(m.snapshot.PullRequests))
		for _, pr := range m.snapshot.PullRequests {
			id := fmt.Sprintf("pr:%d", pr.Number)
			m.prsByID[id] = pr
			state := "OPEN"
			if pr.IsDraft {
				state = "DRAFT"
			}
			paths = append(paths, tree.PathItem{
				ID:    id,
				Path:  tree.NormalizeBranch(pr.HeadBranch, m.prefixes),
				Label: fmt.Sprintf("#%d  %s  [%s]", pr.Number, pr.Title, state),
			})
		}
		return paths
	}
	paths := make([]tree.PathItem, 0, len(m.snapshot.Branches))
	for _, branch := range m.snapshot.Branches {
		id := "branch:" + branch.Name
		m.branchesByID[id] = branch
		paths = append(paths, tree.PathItem{
			ID:    id,
			Path:  tree.NormalizeBranch(branch.Name, m.prefixes),
			Label: branch.Name,
		})
	}
	return paths
}

func (m *Model) persistFolder() {
	if m.saveFolder == nil {
		return
	}
	if err := m.saveFolder(m.folder); err != nil {
		m.status = "Could not persist folder: " + err.Error()
	}
}

func (m Model) currentEntry() (tree.Entry, bool) {
	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return tree.Entry{}, false
	}
	return m.entries[m.cursor], true
}

func (m Model) selectionStatus(entry tree.Entry) string {
	if pr, ok := m.prsByID[entry.ID]; ok {
		return fmt.Sprintf("Selected PR #%d at %s", pr.Number, pr.HeadSHA)
	}
	if branch, ok := m.branchesByID[entry.ID]; ok {
		return fmt.Sprintf("Selected branch %s at %s", branch.Name, branch.SHA)
	}
	return "Selected " + entry.Label
}

func displayQuery(query string) string {
	if query == "" {
		return "(none)"
	}
	return query
}

func sameDisplayPath(left, right string) bool {
	return strings.EqualFold(strings.TrimRight(left, "/\\"), strings.TrimRight(right, "/\\"))
}
