package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/launch"
)

func TestV312AltOFocusesConsole(t *testing.T) {
	m := WithV312UX(Model{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}, Alt: true})
	got := updated.(V312Model)
	if !got.consoleFocus || got.focus != paneDetails {
		t.Fatalf("consoleFocus=%v focus=%v", got.consoleFocus, got.focus)
	}
}

func TestV312CtrlCOutsideConsoleAsksBeforeExit(t *testing.T) {
	m := WithV312UX(Model{})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	got := updated.(V312Model)
	if cmd != nil {
		t.Fatal("Ctrl+C outside console must not quit immediately")
	}
	if !got.exitConfirm {
		t.Fatal("Ctrl+C did not open exit confirmation")
	}
	view := got.View()
	if !strings.Contains(view, "Exit gh-tree") || !strings.Contains(view, "[Enter] Exit") {
		t.Fatalf("exit modal missing from view: %q", view)
	}
}

func TestV312DownFromActiveRootMovesToLaunch(t *testing.T) {
	m := WithV312UX(Model{})
	m.focus = paneDetails
	m.activeSubFocus = 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	got := updated.(V312Model)
	if got.activeSubFocus != 1 || got.consoleFocus {
		t.Fatalf("subfocus=%d console=%v", got.activeSubFocus, got.consoleFocus)
	}
	updated, _ = got.Update(tea.KeyMsg{Type: tea.KeyUp})
	back := updated.(V312Model)
	if back.activeSubFocus != 0 {
		t.Fatalf("up from first launch did not return to root: %d", back.activeSubFocus)
	}
}

func TestV312AltNumberSelectsConsoleTab(t *testing.T) {
	m := WithV312UX(Model{})
	m.consoles = []launch.ProcessSnapshot{
		{ID: 11, Invocation: launch.Invocation{Name: "one"}, State: launch.StateRunning},
		{ID: 12, Invocation: launch.Invocation{Name: "two"}, State: launch.StateRunning},
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}, Alt: true})
	got := updated.(V312Model)
	if got.consoleCursor != 1 || !got.consoleFocus {
		t.Fatalf("cursor=%d console=%v", got.consoleCursor, got.consoleFocus)
	}
}

func TestV312ConsoleRendersTabsAndRollingState(t *testing.T) {
	m := WithV312UX(Model{})
	m.activityFrame = 2
	m.consoles = []launch.ProcessSnapshot{{
		ID: 1,
		Invocation: launch.Invocation{Name: "dev:lan", Command: "npm", Args: []string{"run", "dev:lan"}},
		State: launch.StateRunning,
		PID: 123,
		Lines: []string{"server ready"},
	}}
	plain := stripANSI312(m.renderConsole312(80, 10))
	if !strings.Contains(plain, "[1 dev:lan*]") || !strings.Contains(plain, "running") || !strings.Contains(plain, "server ready") {
		t.Fatalf("console=%q", plain)
	}
}

func stripANSI312(s string) string {
	// Existing render tests generally only need semantic containment. Remove the
	// common CSI color/style sequences without adding another dependency here.
	var out strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) {
				c := s[i]
				i++
				if c >= '@' && c <= '~' {
					break
				}
			}
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}
