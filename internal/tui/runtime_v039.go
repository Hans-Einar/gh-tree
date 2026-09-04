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

func wrapRuntimeModel(model tea.Model) tea.Model {
	if inner, ok := model.(Model); ok {
		return RuntimeModel{Model: inner}
	}
	return model
}

func (m RuntimeModel) View() string {
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
