package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RuntimeModel adds small runtime-only UX refinements without changing the
// stable Model API used by existing integrations and tests.
type RuntimeModel struct{ Model }

func WithRuntimeUX(model Model) RuntimeModel { return RuntimeModel{Model: model} }

func (m RuntimeModel) Init() tea.Cmd { return m.Model.Init() }

func (m RuntimeModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := message.(tea.KeyMsg); ok && m.dialog == dialogLaunchPicker && key.String() == " " {
		m.keepMakeStackInsideProjectRoot()
	}

	wasCreate := m.dialog == dialogCreateWorktree
	if wasCreate && m.inputField < 0 {
		if key, ok := message.(tea.KeyMsg); ok {
			switch key.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "q":
				m.dialog = dialogNone
				m.status = "Create worktree cancelled"
				return m, nil
			case "e":
				m.inputField = 0
				m.status = "Editing suggested worktree path/name; Tab also edits the local branch"
				return m, nil
			case "tab":
				m.inputField = 0
				m.status = "Editing suggested worktree path/name"
				return m, nil
			case "enter":
				updated, cmd := m.executeCreateWorktree()
				return wrapRuntimeModel(updated), cmd
			default:
				return m, nil
			}
		}
	}

	updated, cmd := m.Model.Update(message)
	inner, ok := updated.(Model)
	if !ok {
		return updated, cmd
	}
	if !wasCreate && inner.dialog == dialogCreateWorktree {
		// -1 is the review/accept state. The existing dialog editor uses 0/1.
		inner.inputField = -1
		inner.status = "Create worktree: Enter accepts suggestion; e edits path/name"
	}
	return RuntimeModel{Model: inner}, cmd
}

func (m *RuntimeModel) keepMakeStackInsideProjectRoot() {
	if m.launchCursor < 0 || m.launchCursor >= len(m.launchCandidates) {
		return
	}
	current := m.launchCandidates[m.launchCursor]
	if current.Provider != "make" || len(m.launchSelected) == 0 {
		return
	}
	for _, idx := range m.launchSelected {
		if idx < 0 || idx >= len(m.launchCandidates) {
			continue
		}
		selected := m.launchCandidates[idx]
		if selected.Provider == "make" && selected.Dir != current.Dir {
			m.launchSelected = nil
			m.status = "Make stacks are scoped to one project root; started a new stack in " + displayProjectRoot(current.Dir)
			return
		}
	}
}

func wrapRuntimeModel(model tea.Model) tea.Model {
	if inner, ok := model.(Model); ok {
		return RuntimeModel{Model: inner}
	}
	return model
}

func (m RuntimeModel) View() string {
	if m.dialog == dialogLaunchPicker {
		return m.viewWithLaunchPicker()
	}
	if m.dialog != dialogCreateWorktree {
		return m.Model.View()
	}

	width := m.width
	if width < 40 {
		width = 100
	}
	height := m.height
	if height < 18 {
		height = 30
	}
	var out strings.Builder
	out.WriteString(m.renderHeader(width))
	out.WriteString("\n")
	upper := max(8, (height-7)*2/3)
	lower := max(8, height-upper-6)
	switch m.mode {
	case modeCommits:
		out.WriteString(m.renderCommitCreateCockpit(width, upper+lower+1))
	default:
		out.WriteString(m.renderCreateCockpit(width, upper, lower))
	}
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine())
	out.WriteString("\n")
	out.WriteString(dimStyle.Render(m.renderCreateFooter()))
	return out.String()
}

func (m RuntimeModel) viewWithLaunchPicker() string {
	base := m.Model.View()
	old := m.Model.renderLaunchPickerDialog()
	if old == "" {
		return base
	}
	return strings.TrimSuffix(base, old) + m.renderNestedLaunchPickerDialog()
}

func (m RuntimeModel) renderNestedLaunchPickerDialog() string {
	lines := []string{
		headerStyle.Render("Launch discovery · active worktree"),
		dimStyle.Render("Project roots come from provider manifests; npm ':' names remain one exact script."),
		"",
	}
	visible := min(len(m.launchCandidates), 12)
	start := 0
	if m.launchCursor >= visible {
		start = m.launchCursor - visible + 1
	}
	end := min(len(m.launchCandidates), start+visible)
	selectedMap := map[int]bool{}
	for _, i := range m.launchSelected {
		selectedMap[i] = true
	}
	for i := start; i < end; i++ {
		c := m.launchCandidates[i]
		cursor := "  "
		if i == m.launchCursor {
			cursor = "> "
		}
		check := "  "
		if selectedMap[i] {
			check = "✓ "
		}
		label := displayProjectRoot(c.Dir) + " · " + c.Provider + "  "
		if c.Provider == "npm" {
			label += c.Script
		} else {
			label += strings.Join(c.Targets, ":")
		}
		line := cursor + check + label
		if i == m.launchCursor {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	if c, ok := m.selectedLaunchCandidate(); ok {
		lines = append(lines, "", dimStyle.Render("selected cwd: "+displayProjectRoot(c.Dir)))
	}
	lines = append(lines, "", dimStyle.Render("[↑/↓] choose  [Space] stack Make in same project  [Enter] run once  [s] save default  [Esc] cancel"))
	return dialogStyle.Render(strings.Join(lines, "\n"))
}

func displayProjectRoot(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return "."
	}
	return strings.ReplaceAll(dir, "\\", "/")
}

func (m RuntimeModel) renderCreateCockpit(width, upperHeight, lowerHeight int) string {
	if width >= 88 {
		lw := width / 2
		rw := width - lw
		left := m.panel("Navigator", m.focus == paneNavigator, m.renderEntries(lw-4, upperHeight-3), lw-1, upperHeight)
		right := m.panel("Local worktrees", true, m.renderWorktreePane(rw-4, upperHeight-3), rw-1, upperHeight)
		lower := m.panel("Create worktree", true, m.renderCreateWorktreePane(width-4, lowerHeight-3), width-1, lowerHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + lower
	}
	left := m.panel("Navigator", m.focus == paneNavigator, m.renderEntries(width-4, max(5, upperHeight/2-2)), width-1, max(7, upperHeight/2))
	right := m.panel("Local worktrees", true, m.renderWorktreePane(width-4, max(4, upperHeight/2-3)), width-1, max(6, upperHeight/2))
	lower := m.panel("Create worktree", true, m.renderCreateWorktreePane(width-4, lowerHeight-3), width-1, lowerHeight)
	return left + "\n" + right + "\n" + lower
}

func (m RuntimeModel) renderCommitCreateCockpit(width, height int) string {
	if width >= 88 {
		lw := width / 2
		rw := width - lw
		left := m.panel("Commits / DAG", true, m.renderCommitList(lw-4, height-3), lw-1, height)
		right := m.panel("Create worktree", true, m.renderCreateWorktreePane(rw-4, height-3), rw-1, height)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	lh := max(8, height/2)
	return m.panel("Commits / DAG", true, m.renderCommitList(width-4, lh-3), width-1, lh) + "\n" +
		m.panel("Create worktree", true, m.renderCreateWorktreePane(width-4, height-lh-3), width-1, max(7, height-lh))
}

func (m RuntimeModel) renderCreateWorktreePane(width, height int) string {
	source := "active commit"
	if m.pendingPR.Number > 0 {
		source = fmt.Sprintf("PR #%d · %s", m.pendingPR.Number, m.pendingPR.HeadBranch)
	} else if m.pendingBranch.Name != "" {
		source = "branch " + m.pendingBranch.Name
	} else if m.pendingRevision != "" {
		source = "commit " + shortSHA(m.pendingRevision)
	}

	lines := []string{truncate("source: "+source, width), ""}
	if m.inputField < 0 {
		lines = append(lines,
			labeledValue("suggested path/name: ", m.inputA, width),
			labeledValue("local branch:        ", coalesce(m.inputB, "(detached)"), width),
			"",
			activeStyle.Render("[Enter] create with suggestion   [e] edit"),
		)
	} else {
		lines = append(lines,
			inputLine("Path/name", m.inputA, m.inputField == 0),
			inputLine("Local branch (blank = detached)", m.inputB, m.inputField == 1),
			"",
			dimStyle.Render("[Tab] field  [Enter] next/create  [Esc] cancel"),
		)
	}
	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

func (m RuntimeModel) renderCreateFooter() string {
	if m.inputField < 0 {
		return "[Enter] create suggested worktree  [e] edit path/name  [Esc] cancel"
	}
	return "[Tab] field  [Enter] next/create  [Backspace] edit  [Esc] cancel"
}
