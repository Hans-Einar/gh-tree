package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Hans-Einar/gh-tree/internal/launch"
	ghversion "github.com/Hans-Einar/gh-tree/internal/version"
)

type v313Backend interface {
	OpenTerminalConsole(context.Context, string, int, int) (launch.ProcessSnapshot, error)
	WriteConsole(int, []byte) error
	ResizeConsole(int, int, int) error
	IsInteractiveConsole(int) bool
}

type terminalOpened313Msg struct {
	snapshot launch.ProcessSnapshot
	err      error
}

type terminalWrite313Msg struct{ err error }
type terminalResize313Msg struct{ err error }

// V313Model upgrades the v0.3.12 Console log viewer into a PTY-backed
// interactive terminal while retaining launch-console tabs in the same tab bar.
type V313Model struct {
	V312Model
}

func WithV313UX(model Model) V313Model {
	return V313Model{V312Model: WithV312UX(model)}
}

func (m V313Model) Init() tea.Cmd { return m.V312Model.Init() }

func (m V313Model) v313() (v313Backend, bool) {
	b, ok := m.backend.(v313Backend)
	return b, ok
}

func (m V313Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case terminalOpened313Msg:
		m.busy = false
		m.refreshConsoles312()
		if msg.err != nil {
			m.status = "Open terminal failed: " + msg.err.Error()
			return m.withActivity313(nil)
		}
		m.selectConsoleID312(msg.snapshot.ID)
		m.consoleFocus = true
		m.focus = paneDetails
		m.consoleScroll = 0
		m.status = fmt.Sprintf("Terminal %d: %s", msg.snapshot.ID, msg.snapshot.Invocation.Name)
		return m.withActivity313(m.resizeSelectedTerminal313())
	case terminalWrite313Msg:
		if msg.err != nil {
			m.status = "Terminal input failed: " + msg.err.Error()
		}
		m.refreshConsoles312()
		return m.withActivity313(nil)
	case terminalResize313Msg:
		if msg.err != nil {
			m.status = "Terminal resize failed: " + msg.err.Error()
		}
		return m.withActivity313(nil)
	}

	if key, ok := message.(tea.KeyMsg); ok {
		if m.exitConfirm || m.worktreeModal || m.dialog != dialogNone || m.deploying || m.confirmDiscard || m.confirmCheckout {
			return m.delegate313(message)
		}
		k := key.String()
		if k == "alt+t" {
			return m.openTerminal313()
		}

		// All gh-tree Alt mnemonics stay global even while a shell owns the
		// console keyboard. Alt+1..9 therefore remain reliable tab switches.
		if strings.HasPrefix(k, "alt+") || k == "tab" || k == "ctrl+shift+tab" {
			model, cmd := m.delegate313(message)
			if next, ok := model.(V313Model); ok && next.consoleFocus {
				return next, tea.Batch(cmd, next.resizeSelectedTerminal313())
			}
			return model, cmd
		}

		if m.consoleFocus && m.selectedInteractive313() {
			// F-keys remain application shortcuts. Shift+F5 only stops launch
			// consoles; an interactive shell is closed naturally with `exit`.
			switch k {
			case "f5", "ctrl+f5", "f6":
				return m.delegate313(message)
			case "shift+f5":
				m.status = "Shift+F5 stops launch consoles; type exit to close this shell"
				return m, nil
			}
			data, handled := terminalKeyBytes313(key)
			if handled {
				return m.writeTerminal313(data)
			}
		}
	}

	if _, ok := message.(tea.WindowSizeMsg); ok {
		model, cmd := m.delegate313(message)
		if next, ok := model.(V313Model); ok {
			return next, tea.Batch(cmd, next.resizeSelectedTerminal313())
		}
		return model, cmd
	}
	return m.delegate313(message)
}

func (m V313Model) delegate313(message tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.V312Model.Update(message)
	inner, ok := updated.(V312Model)
	if !ok {
		return updated, cmd
	}
	m.V312Model = inner
	return m, cmd
}

func (m V313Model) openTerminal313() (tea.Model, tea.Cmd) {
	if m.activeWorktree == "" {
		m.status = "No active worktree"
		return m, nil
	}
	b, ok := m.v313()
	if !ok {
		m.status = "Interactive terminal backend unavailable"
		return m, nil
	}
	path := m.activeWorktree
	w, h := m.consoleGeometry313()
	m.busy = true
	m.status = "Opening interactive terminal"
	return m.withActivity313(func() tea.Msg {
		snap, err := b.OpenTerminalConsole(context.Background(), path, w, h)
		return terminalOpened313Msg{snapshot: snap, err: err}
	})
}

func (m V313Model) writeTerminal313(data []byte) (tea.Model, tea.Cmd) {
	snap, ok := m.selectedConsole312()
	if !ok || len(data) == 0 {
		return m, nil
	}
	b, ok := m.v313()
	if !ok {
		return m, nil
	}
	id := snap.ID
	return m.withActivity313(func() tea.Msg { return terminalWrite313Msg{err: b.WriteConsole(id, data)} })
}

func (m V313Model) resizeSelectedTerminal313() tea.Cmd {
	snap, ok := m.selectedConsole312()
	if !ok || snap.Invocation.Provider != "terminal" {
		return nil
	}
	b, ok := m.v313()
	if !ok {
		return nil
	}
	w, h := m.consoleGeometry313()
	id := snap.ID
	return func() tea.Msg { return terminalResize313Msg{err: b.ResizeConsole(id, w, h)} }
}

func (m V313Model) selectedInteractive313() bool {
	snap, ok := m.selectedConsole312()
	if !ok {
		return false
	}
	if snap.Invocation.Provider == "terminal" {
		return true
	}
	if b, ok := m.v313(); ok {
		return b.IsInteractiveConsole(snap.ID)
	}
	return false
}

func (m V313Model) consoleGeometry313() (int, int) {
	width := m.width
	if width < 40 {
		width = 100
	}
	height := m.height
	if height < 18 {
		height = 30
	}
	upper := max(8, (height-7)*2/3)
	lower := max(8, height-upper-6)
	_ = upper
	if width >= 88 {
		consoleW := width - width*11/20
		return max(20, consoleW-6), max(4, lower-8)
	}
	consoleH := max(7, lower-max(7, lower/2))
	return max(20, width-6), max(4, consoleH-8)
}

func (m V313Model) withActivity313(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	model, activity := m.withActivity312(nil)
	inner, ok := model.(V312Model)
	if ok {
		m.V312Model = inner
	}
	if cmd == nil {
		return m, activity
	}
	if activity == nil {
		return m, cmd
	}
	return m, tea.Batch(cmd, activity)
}

func (m V313Model) View() string {
	base := m.viewBase313()
	if m.exitConfirm {
		return overlay311(base, m.modalFrame311("Exit gh-tree", "Stop all attached consoles and return to the shell?\n\n[Enter] Exit    [Esc] Cancel"), m.width, m.height)
	}
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
		return overlay311(base, m.renderDialog311(), m.width, m.height)
	}
	return base
}

func (m V313Model) viewBase313() string {
	width := m.width
	if width < 40 {
		width = 100
	}
	height := m.height
	if height < 18 {
		height = 30
	}
	var out strings.Builder
	out.WriteString(m.renderHeader313(width))
	out.WriteString("\n")
	upper := max(8, (height-7)*2/3)
	lower := max(8, height-upper-6)
	switch m.mode {
	case modeCommits:
		out.WriteString(m.renderCommitCockpit(width, upper+lower+1))
	case modeDiff:
		out.WriteString(m.renderDiffCockpit(width, upper+lower+1))
	default:
		out.WriteString(m.renderCockpit313(width, upper, lower))
	}
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine312())
	out.WriteString("\n")
	out.WriteString(dimStyle.Render(m.footer313()))
	return out.String()
}

func (m V313Model) renderHeader313(width int) string {
	base := m.renderHeader(width)
	parts := strings.SplitN(base, "\n", 2)
	first := parts[0]
	ver := "gh-tree v" + ghversion.Version
	gap := width - lipgloss.Width(first) - len(ver)
	if gap < 2 {
		gap = 2
	}
	first = first + strings.Repeat(" ", gap) + dimStyle.Render(ver)
	first = truncate(first, width)
	if len(parts) == 2 {
		return first + "\n" + parts[1]
	}
	return first
}

func (m V313Model) renderCockpit313(width, upperHeight, lowerHeight int) string {
	if width >= 88 {
		lw, rw := width/2, width-width/2
		leftBody := m.renderEntries(lw-4, upperHeight-3)
		if m.mode == modeBranches {
			leftBody = m.renderDirectionalBranches(lw-4, upperHeight-3)
		}
		left := m.panel(m.heading311("Navigator", 'N', m.focus == paneNavigator && !m.consoleFocus), m.focus == paneNavigator && !m.consoleFocus, leftBody, lw-1, upperHeight)
		rightTitle := m.heading311("Worktrees", 'W', m.focus == paneWorktrees && !m.consoleFocus)
		rightBody := m.renderWorktreePane(rw-4, upperHeight-3)
		if m.mode == modeBranches && m.branchContext {
			rightTitle = m.heading311("Branch context", 'B', m.focus == paneWorktrees && m.branchSubFocus == 0 && !m.consoleFocus)
			rightBody = m.renderBranchContext311(rw-4, upperHeight-3)
		}
		right := m.panel(rightTitle, m.focus == paneWorktrees && !m.consoleFocus, rightBody, rw-1, upperHeight)
		activeW := width * 11 / 20
		consoleW := width - activeW
		activeOn := m.focus == paneDetails && !m.consoleFocus
		active := m.panel(m.heading311("Active worktree", 'A', activeOn && m.activeSubFocus == 0), activeOn, m.renderActive312(activeW-4, lowerHeight-3), activeW-1, lowerHeight)
		console := m.panel(m.heading311("Console", 'O', m.consoleFocus), m.consoleFocus, m.renderConsole313(consoleW-4, lowerHeight-3), consoleW-1, lowerHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, active, console)
	}
	leftBody := m.renderEntries(width-4, max(5, upperHeight/2-2))
	if m.mode == modeBranches {
		leftBody = m.renderDirectionalBranches(width-4, max(5, upperHeight/2-2))
	}
	left := m.panel(m.heading311("Navigator", 'N', m.focus == paneNavigator && !m.consoleFocus), m.focus == paneNavigator && !m.consoleFocus, leftBody, width-1, max(7, upperHeight/2))
	rightTitle := m.heading311("Worktrees", 'W', m.focus == paneWorktrees && !m.consoleFocus)
	rightBody := m.renderWorktreePane(width-4, max(4, upperHeight/2-3))
	if m.mode == modeBranches && m.branchContext {
		rightTitle = m.heading311("Branch context", 'B', m.focus == paneWorktrees && m.branchSubFocus == 0 && !m.consoleFocus)
		rightBody = m.renderBranchContext311(width-4, max(4, upperHeight/2-3))
	}
	right := m.panel(rightTitle, m.focus == paneWorktrees && !m.consoleFocus, rightBody, width-1, max(6, upperHeight/2))
	activeH := max(7, lowerHeight/2)
	consoleH := max(7, lowerHeight-activeH)
	activeOn := m.focus == paneDetails && !m.consoleFocus
	active := m.panel(m.heading311("Active worktree", 'A', activeOn && m.activeSubFocus == 0), activeOn, m.renderActive312(width-4, activeH-3), width-1, activeH)
	console := m.panel(m.heading311("Console", 'O', m.consoleFocus), m.consoleFocus, m.renderConsole313(width-4, consoleH-3), width-1, consoleH)
	return left + "\n" + right + "\n" + active + "\n" + console
}

func (m V313Model) renderConsole313(width, height int) string {
	if len(m.consoles) == 0 {
		return dimStyle.Render("No consoles yet.\n[Alt+T] open interactive shell\nLaunch a script with Enter or F5.")
	}
	lines := []string{m.renderConsoleTabs312(width)}
	snap, ok := m.selectedConsole312()
	if !ok {
		return strings.Join(lines, "\n")
	}
	state := m.renderProcessState312(snap.State)
	if snap.Invocation.Provider == "terminal" {
		lines = append(lines,
			truncate(fmt.Sprintf("%s · interactive · %s · pid %d", snap.Invocation.Name, state, snap.PID), width),
			truncate("cwd: "+snap.Invocation.Dir, width),
			"",
		)
	} else {
		lines = append(lines,
			truncate(fmt.Sprintf("%s · %s · pid %d", snap.Invocation.Name, state, snap.PID), width),
			truncate("$ "+snap.Invocation.Command+" "+strings.Join(snap.Invocation.Args, " "), width),
			"",
		)
	}
	room := max(1, height-len(lines))
	logs := snap.Lines
	end := len(logs) - m.consoleScroll
	if end < 0 {
		end = 0
	}
	start := maxInt(0, end-room)
	for _, line := range logs[start:end] {
		lines = append(lines, truncate(line, width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

func (m V313Model) footer313() string {
	if m.consoleFocus {
		if m.selectedInteractive313() {
			return "[Alt+1..9] tabs  [Alt+T] new terminal  [Ctrl+C] interrupt  [Alt+N/W/A/O] leave console"
		}
		return "[←/→] console tabs  [Alt+1..9] direct  [Alt+T] terminal  [Ctrl+C/Shift+F5] stop  [F6] restart"
	}
	return m.footer312() + "  [Alt+T] terminal"
}

func terminalKeyBytes313(key tea.KeyMsg) ([]byte, bool) {
	k := key.String()
	switch k {
	case "enter":
		return []byte{'\r'}, true
	case "tab":
		return []byte{'\t'}, true
	case "backspace", "ctrl+h":
		return []byte{0x7f}, true
	case "esc":
		return []byte{0x1b}, true
	case "up":
		return []byte("\x1b[A"), true
	case "down":
		return []byte("\x1b[B"), true
	case "right":
		return []byte("\x1b[C"), true
	case "left":
		return []byte("\x1b[D"), true
	case "home":
		return []byte("\x1b[H"), true
	case "end":
		return []byte("\x1b[F"), true
	case "pgup":
		return []byte("\x1b[5~"), true
	case "pgdown":
		return []byte("\x1b[6~"), true
	case "delete":
		return []byte("\x1b[3~"), true
	case "insert":
		return []byte("\x1b[2~"), true
	case "ctrl+c":
		return []byte{0x03}, true
	case "ctrl+d":
		return []byte{0x04}, true
	case "ctrl+l":
		return []byte{0x0c}, true
	case "ctrl+z":
		return []byte{0x1a}, true
	}
	if strings.HasPrefix(k, "ctrl+") && len(k) == 6 {
		r := rune(k[5])
		if r >= 'a' && r <= 'z' {
			return []byte{byte(r - 'a' + 1)}, true
		}
	}
	if len(key.Runes) > 0 {
		var b strings.Builder
		for _, r := range key.Runes {
			if unicode.IsPrint(r) || r == '\t' {
				b.WriteRune(r)
			}
		}
		if b.Len() > 0 {
			return []byte(b.String()), true
		}
	}
	// Bubble Tea normalizes many printable keys through String even if Runes
	// is empty in synthetic tests.
	if len([]rune(k)) == 1 && !strings.HasPrefix(k, "ctrl+") {
		return []byte(k), true
	}
	return nil, false
}

func parseAltNumber313(k string) (int, bool) {
	if !strings.HasPrefix(k, "alt+") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(k, "alt+"))
	return n, err == nil && n >= 1 && n <= 9
}
