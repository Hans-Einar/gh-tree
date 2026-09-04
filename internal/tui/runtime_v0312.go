package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Hans-Einar/gh-tree/internal/launch"
)

type v312Backend interface {
	RunCandidateConsole(context.Context, string, launch.Candidate) (launch.ProcessSnapshot, error)
	RunDefaultConsole(context.Context, string) (launch.ProcessSnapshot, error)
	ConsoleSnapshots() []launch.ProcessSnapshot
	StopConsole(int) error
	RestartConsole(int) (launch.ProcessSnapshot, error)
	SelectConsole(int) bool
}

type consoleStarted312Msg struct {
	snapshot launch.ProcessSnapshot
	err      error
}

type consoleStopped312Msg struct {
	id  int
	err error
}

type consoleRestarted312Msg struct {
	snapshot launch.ProcessSnapshot
	err      error
}

type activityTick312Msg struct{}

// V312Model adds launch consoles as first-class cockpit panes. It deliberately
// keeps the stable Model/V311 APIs intact so older tests and integrations do not
// need to know about multi-console behavior.
type V312Model struct {
	V311Model
	consoleFocus      bool
	consoles          []launch.ProcessSnapshot
	consoleCursor     int
	consoleScroll     int
	exitConfirm       bool
	activityFrame     int
	activityScheduled bool
}

func WithV312UX(model Model) V312Model {
	return V312Model{V311Model: WithV311UX(model)}
}

func (m V312Model) Init() tea.Cmd {
	return m.V311Model.Init()
}

func (m V312Model) v312() (v312Backend, bool) {
	b, ok := m.backend.(v312Backend)
	return b, ok
}

func (m V312Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if tick, ok := message.(activityTick312Msg); ok {
		_ = tick
		m.activityScheduled = false
		m.activityFrame++
		m.refreshConsoles312()
		return m.withActivity312(nil)
	}

	switch msg := message.(type) {
	case consoleStarted312Msg:
		m.busy = false
		if msg.err != nil {
			m.status = "Launch failed: " + msg.err.Error()
			m.refreshConsoles312()
			return m.withActivity312(nil)
		}
		m.refreshConsoles312()
		m.selectConsoleID312(msg.snapshot.ID)
		m.consoleFocus = true
		m.focus = paneDetails
		m.consoleScroll = 0
		m.status = "Console " + strconv.Itoa(msg.snapshot.ID) + " started: " + msg.snapshot.Invocation.Name
		return m.withActivity312(nil)
	case consoleStopped312Msg:
		m.busy = false
		m.refreshConsoles312()
		if msg.err != nil {
			m.status = "Stop console failed: " + msg.err.Error()
		} else {
			m.status = fmt.Sprintf("Console %d stopped", msg.id)
		}
		return m.withActivity312(nil)
	case consoleRestarted312Msg:
		m.busy = false
		if msg.err != nil {
			m.status = "Restart console failed: " + msg.err.Error()
			m.refreshConsoles312()
			return m.withActivity312(nil)
		}
		m.refreshConsoles312()
		m.selectConsoleID312(msg.snapshot.ID)
		m.consoleFocus = true
		m.status = fmt.Sprintf("Console %d restarted", msg.snapshot.ID)
		return m.withActivity312(nil)
	}

	if key, ok := message.(tea.KeyMsg); ok {
		if m.exitConfirm {
			switch key.String() {
			case "enter":
				return m, tea.Quit
			case "esc":
				m.exitConfirm = false
				m.status = "Exit cancelled"
			}
			return m, nil
		}

		// Ctrl+C is now context-routed. Console focus interrupts the selected
		// process; everywhere else it requests an explicit application exit.
		if key.String() == "ctrl+c" {
			if m.consoleFocus {
				model, cmd := m.stopSelectedConsole312()
				return model, cmd
			}
			m.exitConfirm = true
			m.status = "Exit gh-tree?"
			return m, nil
		}

		if m.dialog == dialogNone && !m.deploying && !m.busy && !m.searching && m.mode != modeCommits && m.mode != modeDiff && !m.worktreeModal && !m.confirmDiscard && !m.confirmCheckout {
			if handled, model, cmd := m.handle312Key(key); handled {
				return model, cmd
			}
		}
	}

	updated, cmd := m.V311Model.Update(message)
	inner, ok := updated.(V311Model)
	if !ok {
		return updated, cmd
	}
	m.V311Model = inner
	if msg, ok := message.(launchMsg); ok {
		m.refreshConsoles312()
		if msg.err == nil && msg.snapshot.ID > 0 {
			m.selectConsoleID312(msg.snapshot.ID)
			m.consoleFocus = true
			m.focus = paneDetails
		}
	}
	return m.withActivity312(cmd)
}

func (m V312Model) handle312Key(key tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	k := key.String()
	switch k {
	case "alt+o":
		m.consoleFocus = true
		m.focus = paneDetails
		m.status = "Focus: Console"
		m.refreshConsoles312()
		return true, m, nil
	case "alt+a":
		m.consoleFocus = false
		m.focus = paneDetails
		m.activeSubFocus = 0
		m.status = "Focus: Active worktree"
		return true, m, nil
	case "tab":
		switch {
		case m.consoleFocus:
			m.consoleFocus = false
			m.focus = paneNavigator
		case m.focus == paneNavigator:
			m.focus = paneWorktrees
		case m.focus == paneWorktrees:
			m.focus = paneDetails
			m.activeSubFocus = 0
		default:
			m.consoleFocus = true
			m.focus = paneDetails
		}
		m.status = "Focus: " + m.focus312()
		return true, m, nil
	case "ctrl+shift+tab":
		switch {
		case m.consoleFocus:
			m.consoleFocus = false
			m.focus = paneDetails
			m.activeSubFocus = 0
		case m.focus == paneDetails:
			m.focus = paneWorktrees
		case m.focus == paneWorktrees:
			m.focus = paneNavigator
		default:
			m.consoleFocus = true
			m.focus = paneDetails
		}
		m.status = "Focus: " + m.focus312()
		return true, m, nil
	case "f5":
		if !m.consoleFocus && m.focus == paneDetails && m.activeSubFocus == 1 {
			model, cmd := m.runSelectedConsole312()
			return true, model, cmd
		}
		model, cmd := m.runDefaultConsole312()
		return true, model, cmd
	case "shift+f5":
		model, cmd := m.stopSelectedConsole312()
		return true, model, cmd
	case "f6":
		if m.consoleFocus {
			model, cmd := m.restartSelectedConsole312()
			return true, model, cmd
		}
	}

	if strings.HasPrefix(k, "alt+") && len(k) == len("alt+1") {
		n := int(k[len(k)-1] - '0')
		if n >= 1 && n <= 9 {
			m.refreshConsoles312()
			if n <= len(m.consoles) {
				m.consoleCursor = n - 1
				m.consoleScroll = 0
				m.consoleFocus = true
				m.focus = paneDetails
				if b, ok := m.v312(); ok {
					b.SelectConsole(m.consoles[m.consoleCursor].ID)
				}
				m.status = fmt.Sprintf("Console %d", n)
			}
			return true, m, nil
		}
	}

	if m.consoleFocus {
		switch k {
		case "left", "h":
			if m.consoleCursor > 0 {
				m.consoleCursor--
				m.consoleScroll = 0
				m.selectCurrentConsoleBackend312()
			}
			return true, m, nil
		case "right", "l":
			if m.consoleCursor+1 < len(m.consoles) {
				m.consoleCursor++
				m.consoleScroll = 0
				m.selectCurrentConsoleBackend312()
			}
			return true, m, nil
		case "up", "k":
			m.consoleScroll++
			return true, m, nil
		case "down", "j":
			if m.consoleScroll > 0 {
				m.consoleScroll--
			}
			return true, m, nil
		case "pgup":
			m.consoleScroll += 8
			return true, m, nil
		case "pgdown":
			m.consoleScroll = maxInt(0, m.consoleScroll-8)
			return true, m, nil
		}
	}

	if !m.consoleFocus && m.focus == paneDetails {
		switch k {
		case "down", "j":
			if m.activeSubFocus == 0 {
				m.activeSubFocus = 1
				m.status = "Focus: Launch"
				if m.activeWorktree != "" && m.launchLoadedPath != m.activeWorktree {
					return true, m, m.discoverLaunch311(m.activeWorktree)
				}
				return true, m, nil
			}
		case "up", "k":
			if m.activeSubFocus == 1 && m.launchCursor == 0 {
				m.activeSubFocus = 0
				m.status = "Focus: Active worktree"
				return true, m, nil
			}
		}
	}
	return false, m, nil
}

func (m V312Model) runSelectedConsole312() (tea.Model, tea.Cmd) {
	if m.launchCursor < 0 || m.launchCursor >= len(m.launchCandidates) || m.activeWorktree == "" {
		m.status = "No launch point selected"
		return m, nil
	}
	b, ok := m.v312()
	if !ok {
		m.status = "Multi-console backend unavailable"
		return m, nil
	}
	path := m.activeWorktree
	candidate := m.launchCandidates[m.launchCursor]
	m.busy = true
	m.status = "Starting " + launchLabel311(candidate)
	return m.withActivity312(func() tea.Msg {
		snap, err := b.RunCandidateConsole(context.Background(), path, candidate)
		return consoleStarted312Msg{snapshot: snap, err: err}
	})
}

func (m V312Model) runDefaultConsole312() (tea.Model, tea.Cmd) {
	if m.activeWorktree == "" {
		m.status = "No active worktree"
		return m, nil
	}
	b, ok := m.v312()
	if !ok {
		m.status = "Multi-console backend unavailable"
		return m, nil
	}
	path := m.activeWorktree
	m.busy = true
	m.status = "Starting default launch"
	return m.withActivity312(func() tea.Msg {
		snap, err := b.RunDefaultConsole(context.Background(), path)
		return consoleStarted312Msg{snapshot: snap, err: err}
	})
}

func (m V312Model) stopSelectedConsole312() (tea.Model, tea.Cmd) {
	m.refreshConsoles312()
	snap, ok := m.selectedConsole312()
	if !ok {
		m.status = "No console selected"
		return m, nil
	}
	if snap.State != launch.StateRunning && snap.State != launch.StateStarting {
		m.status = fmt.Sprintf("Console %d is not running", snap.ID)
		return m, nil
	}
	b, ok := m.v312()
	if !ok {
		return m, nil
	}
	id := snap.ID
	m.busy = true
	m.status = fmt.Sprintf("Stopping console %d", id)
	return m.withActivity312(func() tea.Msg { return consoleStopped312Msg{id: id, err: b.StopConsole(id)} })
}

func (m V312Model) restartSelectedConsole312() (tea.Model, tea.Cmd) {
	m.refreshConsoles312()
	snap, ok := m.selectedConsole312()
	if !ok {
		m.status = "No console selected"
		return m, nil
	}
	b, ok := m.v312()
	if !ok {
		return m, nil
	}
	id := snap.ID
	m.busy = true
	m.status = fmt.Sprintf("Restarting console %d", id)
	return m.withActivity312(func() tea.Msg {
		started, err := b.RestartConsole(id)
		return consoleRestarted312Msg{snapshot: started, err: err}
	})
}

func (m *V312Model) refreshConsoles312() {
	b, ok := m.v312()
	if !ok {
		return
	}
	selectedID := 0
	if m.consoleCursor >= 0 && m.consoleCursor < len(m.consoles) {
		selectedID = m.consoles[m.consoleCursor].ID
	}
	m.consoles = b.ConsoleSnapshots()
	if len(m.consoles) == 0 {
		m.consoleCursor = 0
		return
	}
	if selectedID != 0 {
		for i := range m.consoles {
			if m.consoles[i].ID == selectedID {
				m.consoleCursor = i
				return
			}
		}
	}
	if m.consoleCursor >= len(m.consoles) {
		m.consoleCursor = len(m.consoles) - 1
	}
}

func (m *V312Model) selectConsoleID312(id int) {
	for i := range m.consoles {
		if m.consoles[i].ID == id {
			m.consoleCursor = i
			if b, ok := m.v312(); ok {
				b.SelectConsole(id)
			}
			return
		}
	}
}

func (m V312Model) selectCurrentConsoleBackend312() {
	if snap, ok := m.selectedConsole312(); ok {
		if b, have := m.v312(); have {
			b.SelectConsole(snap.ID)
		}
	}
}

func (m V312Model) selectedConsole312() (launch.ProcessSnapshot, bool) {
	if m.consoleCursor < 0 || m.consoleCursor >= len(m.consoles) {
		return launch.ProcessSnapshot{}, false
	}
	return m.consoles[m.consoleCursor], true
}

func (m V312Model) withActivity312(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if !m.activity312() || m.activityScheduled {
		return m, cmd
	}
	m.activityScheduled = true
	tick := tea.Tick(140*time.Millisecond, func(time.Time) tea.Msg { return activityTick312Msg{} })
	if cmd == nil {
		return m, tick
	}
	return m, tea.Batch(cmd, tick)
}

func (m V312Model) activity312() bool {
	if m.busy || m.loading || m.deploying {
		return true
	}
	for _, snap := range m.consoles {
		if snap.State == launch.StateRunning || snap.State == launch.StateStarting {
			return true
		}
	}
	return false
}

func (m V312Model) focus312() string {
	if m.consoleFocus {
		return "Console"
	}
	return m.focus310()
}

func (m V312Model) View() string {
	base := m.viewBase312()
	if m.exitConfirm {
		return overlay311(base, m.modalFrame311("Exit gh-tree", "Stop viewing gh-tree and return to the shell?\n\n[Enter] Exit    [Esc] Cancel"), m.width, m.height)
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

func (m V312Model) viewBase312() string {
	if m.mode == modeCommits || m.mode == modeDiff {
		return m.V311Model.View()
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
	out.WriteString(m.renderCockpit312(width, upper, lower))
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine312())
	out.WriteString("\n")
	out.WriteString(dimStyle.Render(m.footer312()))
	return out.String()
}

func (m V312Model) renderCockpit312(width, upperHeight, lowerHeight int) string {
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
		console := m.panel(m.heading311("Console", 'O', m.consoleFocus), m.consoleFocus, m.renderConsole312(consoleW-4, lowerHeight-3), consoleW-1, lowerHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, active, console)
	}

	// Narrow terminals stack Console below Active rather than sacrificing a
	// usable log viewport.
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
	console := m.panel(m.heading311("Console", 'O', m.consoleFocus), m.consoleFocus, m.renderConsole312(width-4, consoleH-3), width-1, consoleH)
	return left + "\n" + right + "\n" + active + "\n" + console
}

func (m V312Model) renderActive312(width, height int) string {
	return m.animateActivityText312(m.renderActive311(width, height))
}

func (m V312Model) renderConsole312(width, height int) string {
	if len(m.consoles) == 0 {
		return dimStyle.Render("No consoles yet.\nLaunch a script with Enter or F5.")
	}
	lines := []string{m.renderConsoleTabs312(width)}
	snap, ok := m.selectedConsole312()
	if !ok {
		return strings.Join(lines, "\n")
	}
	state := m.renderProcessState312(snap.State)
	lines = append(lines,
		truncate(fmt.Sprintf("%s · %s · pid %d", snap.Invocation.Name, state, snap.PID), width),
		truncate("$ "+snap.Invocation.Command+" "+strings.Join(snap.Invocation.Args, " "), width),
		"",
	)
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

func (m V312Model) renderConsoleTabs312(width int) string {
	var tabs []string
	limit := min(9, len(m.consoles))
	for i := 0; i < limit; i++ {
		snap := m.consoles[i]
		name := snap.Invocation.Name
		if name == "" {
			name = snap.Invocation.Provider
		}
		if len(name) > 12 {
			name = name[:12]
		}
		marker := ""
		if snap.State == launch.StateRunning || snap.State == launch.StateStarting {
			marker = "*"
		}
		label := fmt.Sprintf("[%d %s%s]", i+1, name, marker)
		if i == m.consoleCursor {
			label = selectedStyle.Render(label)
		}
		tabs = append(tabs, label)
	}
	return truncate(strings.Join(tabs, " "), width)
}

func (m V312Model) renderProcessState312(state launch.ProcessState) string {
	switch state {
	case launch.StateRunning:
		return m.rollingWord312("running")
	case launch.StateStarting:
		return m.rollingWord312("starting")
	case launch.StateFailed:
		return "failed"
	case launch.StateStopped:
		return "stopped"
	default:
		return string(state)
	}
}

func (m V312Model) rollingWord312(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return word
	}
	bright := m.activityFrame % len(runes)
	var out strings.Builder
	for i, r := range runes {
		color := lipgloss.Color("8")
		if i == bright {
			color = lipgloss.Color("15")
		} else if (i+1)%len(runes) == bright || (i+len(runes)-1)%len(runes) == bright {
			color = lipgloss.Color("7")
		}
		out.WriteString(lipgloss.NewStyle().Foreground(color).Bold(i == bright).Render(string(r)))
	}
	return out.String()
}

func (m V312Model) animateActivityText312(text string) string {
	for _, word := range []string{"running", "Running", "loading", "Loading", "working", "Working", "starting", "Starting"} {
		if strings.Contains(text, word) {
			text = strings.ReplaceAll(text, word, m.rollingWord312(word))
		}
	}
	return text
}

func (m V312Model) renderStatusLine312() string {
	return m.animateActivityText312(m.renderStatusLine())
}

func (m V312Model) footer312() string {
	if m.consoleFocus {
		return "[Alt+1..9] tab  [←/→] console  [↑/↓] scroll  [Ctrl+C] stop  [Shift+F5] stop  [F6] restart  [Alt+A] active"
	}
	if m.focus == paneDetails && m.activeSubFocus == 1 {
		return "[↑/↓] launch  [Enter/F5] run in new console  [Alt+O] console  [Ctrl+F5] choose  [Shift+F5] stop"
	}
	if m.focus == paneDetails && m.activeSubFocus == 0 {
		return "[↓] launch  [Enter] change worktree  [Alt+L] launch  [Alt+O] console  [h] history  [d] diff"
	}
	return m.footer311() + "  [Alt+O] console"
}
