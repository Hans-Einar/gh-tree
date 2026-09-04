package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/Hans-Einar/gh-tree/internal/launch"
)

type launchList311Msg struct {
	path  string
	items []launch.Candidate
	err   error
}

// V311Model adds a ncurses/YaST-like focus model and floating modal compositor
// on top of the v0.3.10 cockpit without changing the stable Model API.
type V311Model struct {
	V310Model
	activeSubFocus int // 0=root/status, 1=launch, 2=changes
	worktreeModal bool
	worktreeModalCursor int
	launchLoadedPath string
}

func WithV311UX(model Model) V311Model {
	return V311Model{V310Model: WithV310UX(model)}
}

func (m V311Model) Init() tea.Cmd { return m.V310Model.Init() }

func (m V311Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case launchList311Msg:
		if msg.path != m.activeWorktree {
			return m, nil
		}
		m.launchLoadedPath = msg.path
		if msg.err != nil {
			m.status = "Launch discovery failed: " + msg.err.Error()
			return m, nil
		}
		m.launchCandidates = msg.items
		if m.launchCursor >= len(m.launchCandidates) {
			m.launchCursor = maxInt(0, len(m.launchCandidates)-1)
		}
		return m, nil
	}

	if key, ok := message.(tea.KeyMsg); ok {
		if m.worktreeModal {
			return m.updateWorktreeModal311(key)
		}
		if m.dialog == dialogNone && !m.deploying && !m.busy && !m.searching && m.mode != modeCommits && m.mode != modeDiff {
			switch key.String() {
			case "alt+a":
				m.focus = paneDetails
				m.activeSubFocus = 0
				m.status = "Focus: Active worktree"
				return m, nil
			case "alt+l":
				m.focus = paneDetails
				m.activeSubFocus = 1
				m.status = "Focus: Launch"
				if m.activeWorktree != "" && m.launchLoadedPath != m.activeWorktree {
					return m, m.discoverLaunch311(m.activeWorktree)
				}
				return m, nil
			case "alt+b":
				if m.mode == modeBranches && m.branchContext {
					m.focus = paneWorktrees
					m.branchSubFocus = 0
					m.status = "Focus: Branch context"
					return m, nil
				}
			case "alt+c":
				if m.mode == modeBranches && m.branchContext {
					m.focus = paneWorktrees
					m.branchSubFocus = 1
					m.status = "Focus: Commits"
					return m, nil
				}
			case "alt+m":
				if m.mode == modeBranches && m.branchContext {
					m.focus = paneWorktrees
					m.branchSubFocus = 2
					m.status = "Focus: Message"
					return m, nil
				}
			case "enter":
				if m.focus == paneDetails && m.activeSubFocus == 0 && len(m.snapshot.Worktrees) > 0 {
					m.worktreeModal = true
					m.worktreeModalCursor = m.worktreeCursor
					m.status = "Choose active worktree"
					return m, nil
				}
				if m.focus == paneDetails && m.activeSubFocus == 1 {
					return m.runLaunch311()
				}
			case "up", "k":
				if m.focus == paneDetails && m.activeSubFocus == 1 {
					if m.launchCursor > 0 { m.launchCursor-- }
					return m, nil
				}
			case "down", "j":
				if m.focus == paneDetails && m.activeSubFocus == 1 {
					if m.launchCursor+1 < len(m.launchCandidates) { m.launchCursor++ }
					return m, nil
				}
			case "shift+tab":
				if m.focus == paneDetails {
					m.activeSubFocus = (m.activeSubFocus + 1) % 3
					m.status = "Active worktree subfocus: " + []string{"status", "launch", "changes"}[m.activeSubFocus]
					return m, nil
				}
			}
		}
	}

	oldPath := m.activeWorktree
	updated, cmd := m.V310Model.Update(message)
	inner, ok := updated.(V310Model)
	if !ok {
		return updated, cmd
	}
	m.V310Model = inner
	if m.activeWorktree != "" && (m.activeWorktree != oldPath || m.launchLoadedPath != m.activeWorktree) {
		if _, ok := message.(wtStatusMsg); ok {
			return m, tea.Batch(cmd, m.discoverLaunch311(m.activeWorktree))
		}
	}
	return m, cmd
}

func (m V311Model) discoverLaunch311(path string) tea.Cmd {
	b, ok := m.v3()
	if !ok || path == "" { return nil }
	return func() tea.Msg {
		items, err := b.DiscoverLaunch(context.Background(), path)
		return launchList311Msg{path: path, items: items, err: err}
	}
}

func (m V311Model) runLaunch311() (tea.Model, tea.Cmd) {
	if m.launchCursor < 0 || m.launchCursor >= len(m.launchCandidates) || m.activeWorktree == "" {
		m.status = "No launch point selected"
		return m, nil
	}
	b, ok := m.v3()
	if !ok { return m, nil }
	path := m.activeWorktree
	candidate := m.launchCandidates[m.launchCursor]
	m.status = "Starting " + launchLabel311(candidate)
	return m, func() tea.Msg {
		snapshot, err := b.RunCandidate(context.Background(), path, candidate)
		return launchMsg{snapshot: snapshot, err: err}
	}
}

func (m V311Model) updateWorktreeModal311(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "q":
		m.worktreeModal = false
		m.status = "Worktree change cancelled"
	case "up", "k":
		if m.worktreeModalCursor > 0 { m.worktreeModalCursor-- }
	case "down", "j":
		if m.worktreeModalCursor+1 < len(m.snapshot.Worktrees) { m.worktreeModalCursor++ }
	case "enter":
		if m.worktreeModalCursor >= 0 && m.worktreeModalCursor < len(m.snapshot.Worktrees) {
			wt := m.snapshot.Worktrees[m.worktreeModalCursor]
			m.activeWorktree = wt.Path
			m.worktreeCursor = m.worktreeModalCursor
			m.persistWorktree()
			m.worktreeModal = false
			m.launchLoadedPath = ""
			m.launchCandidates = nil
			m.launchCursor = 0
			m.status = "Active worktree: " + wt.Path
			return m, tea.Batch(m.statusCmd(wt.Path), m.discoverLaunch311(wt.Path))
		}
	}
	return m, nil
}

func (m V311Model) View() string {
	base := m.viewBase311()
	if m.worktreeModal {
		return overlay311(base, m.renderWorktreeChooser311(), m.width, m.height)
	}
	if m.confirmDiscard {
		return overlay311(base, m.modalFrame311("Discard tracked change", "This will restore the selected tracked path from Git.\nUntracked files are never deleted.\n\n[y] Discard    [Esc] Cancel"), m.width, m.height)
	}
	if m.confirmCheckout {
		return overlay311(base, m.modalFrame311("Historical checkout", "Checkout the selected historical commit detached into the active secondary worktree?\n\n[y] Checkout    [Esc] Cancel"), m.width, m.height)
	}
	if m.dialog != dialogNone || m.deploying {
		return overlay311(m.viewBackground311(), m.renderDialog311(), m.width, m.height)
	}
	return base
}

func (m V311Model) viewBase311() string {
	if m.mode == modeCommits || m.mode == modeDiff {
		return m.RuntimeModel.View()
	}
	width := m.width
	if width < 40 { width = 100 }
	height := m.height
	if height < 18 { height = 30 }
	var out strings.Builder
	out.WriteString(m.renderHeader(width))
	out.WriteString("\n")
	upper := max(8, (height-7)*2/3)
	lower := max(8, height-upper-6)
	out.WriteString(m.renderCockpit311(width, upper, lower))
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine())
	out.WriteString("\n")
	out.WriteString(dimStyle.Render(m.footer311()))
	return out.String()
}

func (m V311Model) viewBackground311() string {
	clone := m
	clone.dialog = dialogNone
	clone.deploying = false
	clone.confirmDiscard = false
	clone.confirmCheckout = false
	clone.worktreeModal = false
	return clone.viewBase311()
}

func (m V311Model) renderCockpit311(width, upperHeight, lowerHeight int) string {
	if width >= 88 {
		lw, rw := width/2, width-width/2
		leftBody := m.renderEntries(lw-4, upperHeight-3)
		if m.mode == modeBranches { leftBody = m.renderDirectionalBranches(lw-4, upperHeight-3) }
		left := m.panel(m.heading311("Navigator", 'N', m.focus == paneNavigator), m.focus == paneNavigator, leftBody, lw-1, upperHeight)
		rightTitle := m.heading311("Worktrees", 'W', m.focus == paneWorktrees)
		rightBody := m.renderWorktreePane(rw-4, upperHeight-3)
		if m.mode == modeBranches && m.branchContext {
			rightTitle = m.heading311("Branch context", 'B', m.focus == paneWorktrees && m.branchSubFocus == 0)
			rightBody = m.renderBranchContext311(rw-4, upperHeight-3)
		}
		right := m.panel(rightTitle, m.focus == paneWorktrees, rightBody, rw-1, upperHeight)
		lower := m.panel(m.heading311("Active worktree", 'A', m.focus == paneDetails && m.activeSubFocus == 0), m.focus == paneDetails, m.renderActive311(width-4, lowerHeight-3), width-1, lowerHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + lower
	}
	leftBody := m.renderEntries(width-4, max(5, upperHeight/2-2))
	if m.mode == modeBranches { leftBody = m.renderDirectionalBranches(width-4, max(5, upperHeight/2-2)) }
	left := m.panel(m.heading311("Navigator", 'N', m.focus == paneNavigator), m.focus == paneNavigator, leftBody, width-1, max(7, upperHeight/2))
	rightTitle := m.heading311("Worktrees", 'W', m.focus == paneWorktrees)
	rightBody := m.renderWorktreePane(width-4, max(4, upperHeight/2-3))
	if m.mode == modeBranches && m.branchContext {
		rightTitle = m.heading311("Branch context", 'B', m.focus == paneWorktrees && m.branchSubFocus == 0)
		rightBody = m.renderBranchContext311(width-4, max(4, upperHeight/2-3))
	}
	right := m.panel(rightTitle, m.focus == paneWorktrees, rightBody, width-1, max(6, upperHeight/2))
	lower := m.panel(m.heading311("Active worktree", 'A', m.focus == paneDetails && m.activeSubFocus == 0), m.focus == paneDetails, m.renderActive311(width-4, lowerHeight-3), width-1, lowerHeight)
	return left + "\n" + right + "\n" + lower
}

func (m V311Model) heading311(title string, mnemonic rune, active bool) string {
	idx := strings.Index(strings.ToLower(title), strings.ToLower(string(mnemonic)))
	if idx < 0 { return title }
	pre, hit, post := title[:idx], title[idx:idx+1], title[idx+1:]
	if active {
		base := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
		key := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("11")).Bold(true)
		return base.Render(pre) + key.Render(hit) + base.Render(post)
	}
	key := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	return pre + key.Render(hit) + post
}

func (m V311Model) subheading311(title string, active bool) string {
	if active {
		return lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15")).Bold(true).Render(" " + title + " ")
	}
	return headerStyle.Render(title)
}

func (m V311Model) renderBranchContext311(width, height int) string {
	lines := []string{
		m.subheading311("Branch", m.focus == paneWorktrees && m.branchSubFocus == 0),
		truncate(m.branchName, width),
	}
	if br, ok := m.currentBranchByName311(); ok {
		lines = append(lines, "HEAD: "+shortSHA(br.SHA), "PR: "+strings.TrimSpace(m.prDirections(br.Name)))
	}
	lines = append(lines, "", m.subheading311("Commits", m.focus == paneWorktrees && m.branchSubFocus == 1))
	available := max(3, height-11)
	start := 0
	if m.branchCommitCursor >= available { start = m.branchCommitCursor-available+1 }
	end := min(len(m.branchCommits), start+available)
	for i := start; i < end; i++ {
		c := m.branchCommits[i]
		line := "  " + shortSHA(c.SHA) + "  " + c.Subject
		if i == m.branchCommitCursor { line = selectedStyle.Render("> "+shortSHA(c.SHA)+"  "+c.Subject) }
		lines = append(lines, truncate(line, width))
	}
	lines = append(lines, "", m.subheading311("Message", m.focus == paneWorktrees && m.branchSubFocus == 2))
	if c, ok := m.selectedBranchCommit(); ok {
		message := strings.Split(c.Message, "\n")
		if m.branchMessageScroll < len(message) {
			message = message[m.branchMessageScroll:]
		}
		for _, line := range message {
			if len(lines) >= height { break }
			lines = append(lines, truncate(line, width))
		}
	}
	if len(lines) > height { lines = lines[:height] }
	return strings.Join(lines, "\n")
}

func (m V311Model) currentBranchByName311() (struct{Name, SHA string}, bool) {
	for _, br := range m.snapshot.Branches {
		if br.Name == m.branchName { return struct{Name, SHA string}{br.Name, br.SHA}, true }
	}
	return struct{Name, SHA string}{}, false
}

func (m V311Model) renderActive311(width, height int) string {
	base := m.render310Lower(width, max(5, height-5))
	lines := strings.Split(base, "\n")
	if len(lines) < height {
		lines = append(lines, "", m.subheading311("Launch", m.focus == paneDetails && m.activeSubFocus == 1))
		if len(m.launchCandidates) == 0 {
			lines = append(lines, dimStyle.Render("No launch points discovered"))
		} else {
			room := max(1, height-len(lines))
			start := 0
			if m.launchCursor >= room { start = m.launchCursor-room+1 }
			end := min(len(m.launchCandidates), start+room)
			for i := start; i < end; i++ {
				line := "  " + launchLabel311(m.launchCandidates[i])
				if i == m.launchCursor && m.focus == paneDetails && m.activeSubFocus == 1 { line = selectedStyle.Render("> "+launchLabel311(m.launchCandidates[i])) }
				lines = append(lines, truncate(line, width))
			}
		}
	}
	if len(lines) > height { lines = lines[:height] }
	return strings.Join(lines, "\n")
}

func launchLabel311(c launch.Candidate) string {
	root := displayProjectRoot(c.Dir)
	if c.Provider == "npm" || c.Provider == "pnpm" || c.Provider == "yarn" {
		return root + " · " + c.Provider + "  " + c.Script
	}
	return root + " · " + c.Provider + "  " + strings.Join(c.Targets, " : ")
}

func (m V311Model) footer311() string {
	if m.focus == paneDetails && m.activeSubFocus == 1 {
		return "[↑/↓] launch  [Enter] run  [F5] default  [Ctrl+F5] choose  [Shift+F5] stop  [Alt+A] active worktree"
	}
	if m.focus == paneDetails && m.activeSubFocus == 0 {
		return "[Enter] change active worktree  [Alt+L] launch  [h] history  [d] diff  [Tab] focus  [q] quit"
	}
	if m.mode == modeBranches && m.branchContext && m.focus == paneWorktrees {
		return "[Alt+B] branch  [Alt+C] commits  [Alt+M] message  [Shift+Tab] subpane  [Tab] main pane  [q] quit"
	}
	return m.footer310() + "  [Alt+L] launch"
}

func (m V311Model) renderWorktreeChooser311() string {
	lines := []string{m.subheading311("Change active worktree", true), ""}
	visible := min(10, len(m.snapshot.Worktrees))
	start := 0
	if m.worktreeModalCursor >= visible { start = m.worktreeModalCursor-visible+1 }
	end := min(len(m.snapshot.Worktrees), start+visible)
	for i := start; i < end; i++ {
		wt := m.snapshot.Worktrees[i]
		branch := wt.Branch
		if wt.Detached { branch = "DETACHED" }
		line := fmt.Sprintf("  %-24s %-28s %s", filepathBase311(wt.Path), branch, shortSHA(wt.Head))
		if i == m.worktreeModalCursor { line = selectedStyle.Render("> "+strings.TrimSpace(line)) }
		lines = append(lines, line)
	}
	lines = append(lines, "", dimStyle.Render("[↑/↓] select    [Enter] activate    [Esc] cancel"))
	return m.modalFrame311("Worktrees", strings.Join(lines, "\n"))
}

func filepathBase311(path string) string {
	path = strings.TrimRight(strings.ReplaceAll(path, "\\", "/"), "/")
	if i := strings.LastIndex(path, "/"); i >= 0 { return path[i+1:] }
	return path
}

func (m V311Model) renderDialog311() string {
	w := m.modalWidth311()
	switch m.dialog {
	case dialogCreateWorktree:
		return m.modalFrame311("Create worktree", m.RuntimeModel.renderCreateWorktreePane(w-6, 12))
	case dialogLaunchPicker:
		return m.modalFrame311("Launch", ansi.Strip(m.RuntimeModel.renderNestedLaunchPickerDialog()))
	default:
		body := ansi.Strip(m.Model.renderDialog())
		if strings.TrimSpace(body) == "" { body = m.status }
		return m.modalFrame311("gh-tree", body)
	}
}

func (m V311Model) modalWidth311() int {
	w := m.width
	if w < 40 { w = 100 }
	target := w*3/4
	if target < 58 { target = min(w-4, 58) }
	if target > 96 { target = 96 }
	if target > w-4 { target = w-4 }
	return max(30, target)
}

func (m V311Model) modalFrame311(title, body string) string {
	w := m.modalWidth311()
	titleLine := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render(title)
	content := titleLine + "\n\n" + body
	return lipgloss.NewStyle().Width(w).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("12")).Padding(0, 2).Render(content)
}

func overlay311(base, dialog string, width, height int) string {
	if width < 40 { width = 100 }
	if height < 18 { height = 30 }
	plain := ansi.Strip(base)
	baseLines := strings.Split(plain, "\n")
	for len(baseLines) < height { baseLines = append(baseLines, "") }
	dialogLines := strings.Split(dialog, "\n")
	dw := 0
	for _, line := range dialogLines { if w := lipgloss.Width(line); w > dw { dw = w } }
	dh := len(dialogLines)
	left := maxInt(0, (width-dw)/2)
	top := maxInt(0, (height-dh)/2)
	shade := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	out := make([]string, 0, height)
	for row := 0; row < height; row++ {
		line := padPlain311(baseLines[row], width)
		if row >= top && row < top+dh {
			d := dialogLines[row-top]
			prefix := slicePlain311(line, 0, left)
			suffixStart := min(width, left+dw)
			suffix := slicePlain311(line, suffixStart, width)
			out = append(out, shade.Render(prefix)+d+shade.Render(suffix))
		} else {
			out = append(out, shade.Render(line))
		}
	}
	return strings.Join(out, "\n")
}

func padPlain311(s string, width int) string {
	r := []rune(s)
	if len(r) > width { r = r[:width] }
	if len(r) < width { return string(r)+strings.Repeat(" ", width-len(r)) }
	return string(r)
}

func slicePlain311(s string, start, end int) string {
	r := []rune(s)
	if start > len(r) { start = len(r) }
	if end > len(r) { end = len(r) }
	if start > end { start = end }
	return string(r[start:end])
}
