package tui

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTerminalKeyBytes313(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
		want []byte
	}{
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}, []byte{'\r'}},
		{"tab", tea.KeyMsg{Type: tea.KeyTab}, []byte{'\t'}},
		{"ctrl-c", tea.KeyMsg{Type: tea.KeyCtrlC}, []byte{0x03}},
		{"up", tea.KeyMsg{Type: tea.KeyUp}, []byte("\x1b[A")},
		{"rune", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ø")}, []byte("ø")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := terminalKeyBytes313(tt.key)
			if !ok || !bytes.Equal(got, tt.want) {
				t.Fatalf("terminalKeyBytes313(%q)=(%v,%v), want %v,true", tt.key.String(), got, ok, tt.want)
			}
		})
	}
}

func TestV313HeaderShowsVersion(t *testing.T) {
	m := V313Model{}
	m.repo = "Hans-Einar/gh-tree"
	m.width = 120
	header := m.renderHeader313(120)
	if !strings.Contains(header, "gh-tree v0.3.13") {
		t.Fatalf("header missing version: %q", header)
	}
	first := strings.SplitN(header, "\n", 2)[0]
	if !strings.Contains(first, "gh-tree v0.3.13") {
		t.Fatalf("version not in top line: %q", first)
	}
}

func TestParseAltNumber313(t *testing.T) {
	if n, ok := parseAltNumber313("alt+3"); !ok || n != 3 {
		t.Fatalf("alt+3=(%d,%v)", n, ok)
	}
	if _, ok := parseAltNumber313("alt+0"); ok {
		t.Fatal("alt+0 unexpectedly accepted")
	}
}
