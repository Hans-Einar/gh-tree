package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	activeStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dialogStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
)

func (m Model) View() string {
	width := m.width
	if width < 40 {
		width = 100
	}
	listWidth := width
	detailWidth := width
	if width >= 88 {
		listWidth = width/2 - 2
		detailWidth = width - listWidth - 3
	}

	var out strings.Builder
	out.WriteString(headerStyle.Render("gh tree"))
	out.WriteString("  ")
	out.WriteString(m.repo)
	out.WriteString("  ")
	if m.mode == modePullRequests {
		out.WriteString(activeStyle.Render("[PRs]"))
		out.WriteString(dimStyle.Render("  branches"))
	} else {
		out.WriteString(dimStyle.Render("PRs  "))
		out.WriteString(activeStyle.Render("[branches]"))
	}
	out.WriteString("\n")
	breadcrumb := "/"
	if m.folder != "" {
		breadcrumb += m.folder + "/"
	}
	out.WriteString(dimStyle.Render(breadcrumb))
	if m.query != "" || m.searching {
		out.WriteString("  ")
		out.WriteString(activeStyle.Render("filter: " + m.query + cursorMarker(m.searching)))
	}
	out.WriteString("\n\n")

	left := m.renderEntries(listWidth)
	detail := m.renderDetails(max(40, detailWidth))
	if width >= 88 {
		out.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(listWidth).Render(left),
			" │ ",
			lipgloss.NewStyle().Width(width-listWidth-3).Render(detail),
		))
	} else {
		out.WriteString(left)
		out.WriteString("\n\n")
		out.WriteString(detail)
	}
	out.WriteString("\n\n")
	out.WriteString(m.renderWorktrees(width))
	out.WriteString("\n")
	if strings.HasPrefix(m.status, "✓") {
		out.WriteString(successStyle.Render(m.status))
	} else if strings.Contains(strings.ToLower(m.status), "fail") || strings.Contains(strings.ToLower(m.status), "could not") {
		out.WriteString(errorStyle.Render(m.status))
	} else {
		out.WriteString(dimStyle.Render(m.status))
	}
	out.WriteString("\n")
	out.WriteString(dimStyle.Render("[Enter] open  [Backspace] parent  [p] PRs  [b] branches  [w] deploy  [/] search  [r] refresh  [q] quit"))

	if m.dialog != dialogNone || m.deploying {
		out.WriteString("\n\n")
		out.WriteString(m.renderDialog())
	}
	return out.String()
}

func (m Model) renderEntries(width int) string {
	if m.loading && len(m.entries) == 0 {
		return "Loading…"
	}
	if len(m.entries) == 0 {
		return dimStyle.Render("No matching items")
	}
	visible := m.height - 15
	if visible < 6 {
		visible = 12
	}
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := min(len(m.entries), start+visible)
	lines := make([]string, 0, end-start)
	for index := start; index < end; index++ {
		entry := m.entries[index]
		marker := "  "
		if index == m.cursor {
			marker = "> "
		}
		label := entry.Label
		if entry.IsFolder {
			label = entry.Name + "/"
		}
		line := truncate(marker+label, max(10, width-1))
		if index == m.cursor {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderDetails(width int) string {
	entry, ok := m.currentEntry()
	if !ok {
		return dimStyle.Render("No selection")
	}
	if entry.IsFolder {
		return headerStyle.Render(entry.Path+"/") + "\n" + dimStyle.Render("Enter opens this namespace")
	}
	if pr, ok := m.prsByID[entry.ID]; ok {
		state := "OPEN"
		if pr.IsDraft {
			state = "DRAFT"
		}
		return strings.Join([]string{
			headerStyle.Render(wrapText(fmt.Sprintf("#%d  %s", pr.Number, pr.Title), width)),
			labeledValue("head: ", pr.HeadBranch, width),
			labeledValue("base: ", pr.BaseBranch, width),
			labeledValue("sha:  ", pr.HeadSHA, width),
			state,
		}, "\n")
	}
	if branch, ok := m.branchesByID[entry.ID]; ok {
		return strings.Join([]string{
			headerStyle.Render(wrapText(branch.Name, width)),
			labeledValue("sha: ", branch.SHA, width),
		}, "\n")
	}
	return entry.Label
}

func (m Model) renderWorktrees(width int) string {
	if !m.snapshot.WorktreesEnabled {
		return headerStyle.Render("Worktrees") + "\n" + dimStyle.Render("Unavailable: run inside the selected local repository")
	}
	if len(m.snapshot.Worktrees) == 0 {
		return headerStyle.Render("Worktrees") + "\n" + dimStyle.Render("No worktrees found")
	}
	lines := []string{headerStyle.Render(fmt.Sprintf("Worktrees (%d)", len(m.snapshot.Worktrees)))}
	limit := min(4, len(m.snapshot.Worktrees))
	for _, info := range m.snapshot.Worktrees[:limit] {
		branch := info.Branch
		if info.Detached {
			branch = "detached"
		}
		flags := ""
		if info.Current {
			flags = " [current]"
		} else if info.Primary {
			flags = " [primary]"
		}
		line := fmt.Sprintf("  %s → %s @ %s%s", filepath.Base(info.Path), branch, shortSHA(info.Head), flags)
		lines = append(lines, truncate(line, width))
	}
	if len(m.snapshot.Worktrees) > limit {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  … and %d more", len(m.snapshot.Worktrees)-limit)))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderDialog() string {
	if m.deploying {
		return dialogStyle.Render(fmt.Sprintf("Deploying PR #%d\n%s\n\nPlease wait…", m.pendingPR.Number, m.pendingPR.HeadSHA))
	}
	switch m.dialog {
	case dialogTargetPicker:
		lines := []string{headerStyle.Render(fmt.Sprintf("Deploy PR #%d to which worktree?", m.pendingPR.Number))}
		for index, target := range m.targets {
			marker := "  "
			if index == m.targetCursor {
				marker = "> "
			}
			lines = append(lines, fmt.Sprintf("%s%s → %s (%s)", marker, target.Name, target.Path, target.Branch))
		}
		lines = append(lines, "", dimStyle.Render("[Enter] choose  [Esc] cancel"))
		return dialogStyle.Render(strings.Join(lines, "\n"))
	case dialogConfirm:
		return dialogStyle.Render(strings.Join([]string{
			headerStyle.Render("Confirm local test deployment"),
			fmt.Sprintf("PR:     #%d", m.pendingPR.Number),
			"SHA:    " + m.pendingPR.HeadSHA,
			"Target: " + m.pendingTarget.Name,
			"Path:   " + m.pendingTarget.Path,
			"Branch: " + m.pendingTarget.Branch,
			"",
			"The target must be clean. Only this local test branch will move.",
			activeStyle.Render("Deploy? [y/N]"),
		}, "\n"))
	default:
		return ""
	}
}

func labeledValue(label, value string, width int) string {
	if lipgloss.Width(label+value) <= width {
		return label + value
	}
	return label + "\n" + wrapText(value, width)
}

// wrapText preserves every rune. Exact SHAs and branch names must never be
// ellipsized because they are deployment identity and safety information.
func wrapText(value string, width int) string {
	if width < 1 || lipgloss.Width(value) <= width {
		return value
	}
	var lines []string
	runes := []rune(value)
	for len(runes) > 0 {
		cut := len(runes)
		for cut > 1 && lipgloss.Width(string(runes[:cut])) > width {
			cut--
		}
		lines = append(lines, string(runes[:cut]))
		runes = runes[cut:]
	}
	return strings.Join(lines, "\n")
}

func truncate(value string, width int) string {
	if width < 2 || lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func shortSHA(sha string) string {
	if len(sha) <= 8 {
		return sha
	}
	return sha[:8]
}

func cursorMarker(active bool) string {
	if active {
		return "▏"
	}
	return ""
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}
