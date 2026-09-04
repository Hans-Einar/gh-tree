package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type branchContextMsg struct {
	branch  string
	commits []worktree.Commit
	err     error
}

type worktreeChangesMsg struct {
	path    string
	changes []worktree.Change
	err     error
}

type cleanup310Msg struct {
	kind string
	path string
	err  error
}

type checkoutCommit310Msg struct {
	info worktree.Info
	err  error
}

type v310Backend interface {
	WorktreeChanges(context.Context, string) ([]worktree.Change, error)
	RestorePaths(context.Context, string, ...string) error
}

// V310Model layers the v0.3.10 branch/dirty-worktree cockpit on top of the
// stable v0.3.9 runtime wrapper. Keeping it as a wrapper makes the UX slice
// isolated while the underlying Git safety contracts stay unchanged.
type V310Model struct {
	RuntimeModel
	branchContext      bool
	branchName         string
	branchCommits      []worktree.Commit
	branchCommitCursor int
	branchMessageScroll int
	branchSubFocus     int // 0=info, 1=commits, 2=message
	changes            []worktree.Change
	changeCursor       int
	confirmDiscard     bool
	confirmCheckout    bool
}

func WithV310UX(model Model) V310Model { return V310Model{RuntimeModel: WithRuntimeUX(model), branchSubFocus: 1} }
func (m V310Model) Init() tea.Cmd { return m.RuntimeModel.Init() }

func (m V310Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case branchContextMsg:
		if msg.err != nil {
			m.status = "Branch history failed: " + msg.err.Error()
			return m, nil
		}
		if m.mode != modeBranches || m.branchName != msg.branch {
			return m, nil
		}
		m.branchCommits = msg.commits
		m.branchCommitCursor = 0
		m.branchMessageScroll = 0
		m.branchContext = true
		m.branchSubFocus = 1
		m.focus = paneWorktrees
		m.status = fmt.Sprintf("Loaded %d commits for %s", len(msg.commits), msg.branch)
		return m, nil
	case worktreeChangesMsg:
		if msg.path != m.activeWorktree {
			return m, nil
		}
		if msg.err != nil {
			m.status = "Changed-file list failed: " + msg.err.Error()
			return m, nil
		}
		m.changes = msg.changes
		if m.changeCursor >= len(m.changes) {
			m.changeCursor = maxInt(0, len(m.changes)-1)
		}
		return m, nil
	case cleanup310Msg:
		m.confirmDiscard = false
		if msg.err != nil {
			m.status = titleWord(msg.kind) + " failed: " + msg.err.Error()
			return m, nil
		}
		m.status = "✓ " + msg.kind
		return m, tea.Batch(m.statusCmd(msg.path), m.changesCmd(msg.path))
	case checkoutCommit310Msg:
		m.confirmCheckout = false
		if msg.err != nil {
			m.status = "Checkout commit failed: " + msg.err.Error()
			return m, nil
		}
		m.activeWorktree = msg.info.Path
		m.persistWorktree()
		m.status = "✓ checked out " + shortSHA(msg.info.Head) + " detached in " + filepath.Base(msg.info.Path)
		return m, tea.Batch(m.refreshCmd(), m.statusCmd(msg.info.Path))
	}

	if key, ok := message.(tea.KeyMsg); ok {
		if m.confirmDiscard {
			switch key.String() {
			case "y", "Y":
				return m.executeDiscardSelected()
			case "n", "N", "esc", "q":
				m.confirmDiscard = false
				m.status = "Discard cancelled"
				return m, nil
			}
			return m, nil
		}
		if m.confirmCheckout {
			switch key.String() {
			case "y", "Y":
				return m.executeCheckoutSelectedCommit()
			case "n", "N", "esc", "q":
				m.confirmCheckout = false
				m.status = "Historical checkout cancelled"
				return m, nil
			}
			return m, nil
		}
		if handled, model, cmd := m.handleV310Key(key); handled {
			return model, cmd
		}
	}

	wasDirtyPath := m.activeWorktree
	updated, cmd := m.RuntimeModel.Update(message)
	runtime, ok := updated.(RuntimeModel)
	if !ok {
		return updated, cmd
	}
	m.RuntimeModel = runtime
	if msg, ok := message.(wtStatusMsg); ok && msg.path == wasDirtyPath {
		if msg.err == nil && !msg.status.Clean {
			return m, tea.Batch(cmd, m.changesCmd(msg.path))
		}
		if msg.err == nil && msg.status.Clean {
			m.changes = nil
			m.changeCursor = 0
		}
	}
	return m, cmd
}

func (m V310Model) handleV310Key(key tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	k := key.String()
	if m.dialog != dialogNone || m.deploying || m.busy || m.searching || m.mode == modeCommits || m.mode == modeDiff {
		return false, m, nil
	}
	switch k {
	case "alt+n":
		m.focus = paneNavigator
		m.status = "Focus: Navigator"
		return true, m, nil
	case "alt+w":
		m.branchContext = false
		m.focus = paneWorktrees
		m.status = "Focus: Worktrees"
		return true, m, nil
	case "alt+b":
		if m.mode == modeBranches && m.branchContext {
			m.focus = paneWorktrees
			m.branchSubFocus = 0
			return true, m, nil
		}
	case "alt+c":
		if m.mode == modeBranches && m.branchContext {
			m.focus = paneWorktrees
			m.branchSubFocus = 1
			return true, m, nil
		}
	case "alt+m":
		if m.mode == modeBranches && m.branchContext {
			m.focus = paneWorktrees
			m.branchSubFocus = 2
			return true, m, nil
		}
	case "alt+a":
		m.focus = paneDetails
		m.status = "Focus: Active worktree"
		return true, m, nil
	case "ctrl+shift+tab":
		m.focus = (m.focus + 2) % 3
		return true, m, nil
	}

	if k == "tab" {
		m.focus = (m.focus + 1) % 3
		m.status = "Focus: " + m.focusName310()
		return true, m, nil
	}
	if k == "shift+tab" && m.mode == modeBranches && m.branchContext && m.focus == paneWorktrees {
		m.branchSubFocus = (m.branchSubFocus + 1) % 3
		m.status = "Branch subfocus: " + []string{"info", "commits", "message"}[m.branchSubFocus]
		return true, m, nil
	}

	if m.mode == modeBranches && m.focus == paneNavigator && k == "enter" {
		br, ok := m.currentBranch()
		if !ok {
			return false, m, nil
		}
		m.branchName = br.Name
		m.branchCommits = nil
		m.branchCommitCursor = 0
		m.branchMessageScroll = 0
		m.branchContext = true
		m.branchSubFocus = 1
		m.focus = paneWorktrees
		m.status = "Loading branch commits…"
		return true, m, m.branchContextCmd(br.Name)
	}

	if m.mode == modeBranches && m.branchContext && m.focus == paneWorktrees {
		switch m.branchSubFocus {
		case 1:
			switch k {
			case "up", "k":
				if m.branchCommitCursor > 0 { m.branchCommitCursor--; m.branchMessageScroll = 0 }
				return true, m, nil
			case "down", "j":
				if m.branchCommitCursor+1 < len(m.branchCommits) { m.branchCommitCursor++; m.branchMessageScroll = 0 }
				return true, m, nil
			case "home", "g":
				m.branchCommitCursor = 0; m.branchMessageScroll = 0; return true, m, nil
			case "end", "G":
				if len(m.branchCommits) > 0 { m.branchCommitCursor = len(m.branchCommits)-1; m.branchMessageScroll = 0 }
				return true, m, nil
			case "c":
				return true, m.beginCreateFromBranchContextCommit()
			case "x":
				if _, ok := m.selectedBranchContextCommit(); !ok { return true, m, nil }
				m.confirmCheckout = true
				m.status = "Checkout selected historical commit detached into active secondary worktree? [y/N]"
				return true, m, nil
			}
		case 2:
			switch k {
			case "up", "k": if m.branchMessageScroll > 0 { m.branchMessageScroll-- }; return true, m, nil
			case "down", "j": m.branchMessageScroll++; return true, m, nil
			case "pgup": m.branchMessageScroll = maxInt(0, m.branchMessageScroll-8); return true, m, nil
			case "pgdown": m.branchMessageScroll += 8; return true, m, nil
			}
		}
	}

	if m.focus == paneDetails && len(m.changes) > 0 {
		switch k {
		case "up", "k": if m.changeCursor > 0 { m.changeCursor-- }; return true, m, nil
		case "down", "j": if m.changeCursor+1 < len(m.changes) { m.changeCursor++ }; return true, m, nil
		case "s": return true, m.stageSelectedChange(true)
		case "u": return true, m.stageSelectedChange(false)
		case "z": return true, m.stashDirtyWorktree()
		case "r":
			change, ok := m.selectedChange(); if !ok { return true, m, nil }
			if change.Untracked { m.status = "Untracked files are never deleted by cleanup"; return true, m, nil }
			if change.Conflicted { m.status = "Conflicted files must be resolved explicitly; cleanup refuses them"; return true, m, nil }
			m.confirmDiscard = true
			m.status = "Discard working-tree changes for " + change.Path + "? [y/N]"
			return true, m, nil
		case "d":
			model, cmd := m.Model.beginWorktreeDiff(false)
			if inner, ok := model.(Model); ok { m.RuntimeModel.Model = inner; return true, m, cmd }
			return true, model, cmd
		}
	}
	return false, m, nil
}

func (m V310Model) branchContextCmd(branch string) tea.Cmd {
	b, ok := m.v2()
	if !ok { return nil }
	return func() tea.Msg {
		commits, err := b.CommitsForBranch(context.Background(), branch, 80, 0)
		return branchContextMsg{branch: branch, commits: commits, err: err}
	}
}

func (m V310Model) changesCmd(path string) tea.Cmd {
	b, ok := m.backend.(v310Backend)
	if !ok || path == "" { return nil }
	return func() tea.Msg {
		changes, err := b.WorktreeChanges(context.Background(), path)
		return worktreeChangesMsg{path: path, changes: changes, err: err}
	}
}

func (m V310Model) beginCreateFromBranchContextCommit() (tea.Model, tea.Cmd) {
	commit, ok := m.selectedBranchContextCommit()
	if !ok { return m, nil }
	base := ""
	if len(m.snapshot.Worktrees) > 0 { base = m.snapshot.Worktrees[0].Path }
	m.pendingPR.Number = 0
	m.pendingBranch.Name = ""
	m.pendingRevision = commit.SHA
	m.inputB = ""
	if base != "" { m.inputA = filepath.Join(filepath.Dir(base), filepath.Base(base)+"-"+shortSHA(commit.SHA)) }
	m.inputField = -1
	m.dialog = dialogCreateWorktree
	m.status = "Create worktree from " + shortSHA(commit.SHA) + ": Enter accepts suggestion; e edits path/name"
	return m, nil
}

func (m V310Model) executeCheckoutSelectedCommit() (tea.Model, tea.Cmd) {
	b, ok := m.v2()
	commit, have := m.selectedBranchContextCommit()
	if !ok || !have || m.activeWorktree == "" { m.confirmCheckout = false; return m, nil }
	path := m.activeWorktree
	m.confirmCheckout = false
	m.status = "Checking out historical commit…"
	return m, func() tea.Msg {
		info, err := b.CheckoutWorktree(context.Background(), worktree.CheckoutRequest{Path: path, Revision: commit.SHA, Detach: true})
		return checkoutCommit310Msg{info: info, err: err}
	}
}

func (m V310Model) stageSelectedChange(stage bool) (tea.Model, tea.Cmd) {
	b, ok := m.v3()
	change, have := m.selectedChange()
	if !ok || !have || m.activeWorktree == "" { return m, nil }
	path := m.activeWorktree
	kind := "unstaged " + change.Path
	m.status = "Unstaging " + change.Path + "…"
	return m, func() tea.Msg {
		var err error
		if stage { err = b.StagePaths(context.Background(), path, change.Path); kind = "staged " + change.Path } else { err = b.UnstagePaths(context.Background(), path, change.Path) }
		return cleanup310Msg{kind: kind, path: path, err: err}
	}
}

func (m V310Model) stashDirtyWorktree() (tea.Model, tea.Cmd) {
	b, ok := m.v3()
	if !ok || m.activeWorktree == "" { return m, nil }
	path := m.activeWorktree
	m.status = "Stashing tracked + untracked changes…"
	return m, func() tea.Msg {
		_, err := b.StashPush(context.Background(), path, "gh-tree cleanup stash", true)
		return cleanup310Msg{kind: "stashed worktree changes", path: path, err: err}
	}
}

func (m V310Model) executeDiscardSelected() (tea.Model, tea.Cmd) {
	b, ok := m.backend.(v310Backend)
	change, have := m.selectedChange()
	if !ok || !have || m.activeWorktree == "" { m.confirmDiscard = false; return m, nil }
	path := m.activeWorktree
	m.confirmDiscard = false
	m.status = "Discarding tracked working-tree change…"
	return m, func() tea.Msg {
		err := b.RestorePaths(context.Background(), path, change.Path)
		return cleanup310Msg{kind: "discarded working-tree change for " + change.Path, path: path, err: err}
	}
}

func (m V310Model) selectedBranchContextCommit() (worktree.Commit, bool) {
	if m.branchCommitCursor < 0 || m.branchCommitCursor >= len(m.branchCommits) { return worktree.Commit{}, false }
	return m.branchCommits[m.branchCommitCursor], true
}
func (m V310Model) selectedChange() (worktree.Change, bool) {
	if m.changeCursor < 0 || m.changeCursor >= len(m.changes) { return worktree.Change{}, false }
	return m.changes[m.changeCursor], true
}

func (m V310Model) View() string {
	if m.dialog == dialogCreateWorktree { return m.RuntimeModel.View() }
	if m.mode == modeCommits || m.mode == modeDiff { return m.RuntimeModel.View() }
	width := m.width; if width < 40 { width = 100 }
	height := m.height; if height < 18 { height = 30 }
	var out strings.Builder
	out.WriteString(m.renderHeader(width)); out.WriteString("\n")
	upper := max(8, (height-7)*2/3); lower := max(8, height-upper-6)
	out.WriteString(m.renderV310Cockpit(width, upper, lower))
	out.WriteString("\n"); out.WriteString(m.renderStatusLine()); out.WriteString("\n")
	out.WriteString(dimStyle.Render(m.renderV310Footer()))
	if m.dialog != dialogNone || m.deploying { out.WriteString("\n\n"); out.WriteString(m.renderDialog()) }
	return out.String()
}

func (m V310Model) renderV310Cockpit(width, upperHeight, lowerHeight int) string {
	if width >= 88 {
		lw := width/2; rw := width-lw
		leftBody := m.renderEntries(lw-4, upperHeight-3)
		if m.mode == modeBranches { leftBody = m.renderDirectionalBranches(lw-4, upperHeight-3) }
		left := m.panel(m.mnemonicTitle("Navigator", 'N'), m.focus == paneNavigator, leftBody, lw-1, upperHeight)
		rightTitle := m.mnemonicTitle("Worktrees", 'W'); rightBody := m.renderWorktreePane(rw-4, upperHeight-3)
		if m.mode == modeBranches && m.branchContext { rightTitle = "Branch context"; rightBody = m.renderBranchContext(rw-4, upperHeight-3) }
		right := m.panel(rightTitle, m.focus == paneWorktrees, rightBody, rw-1, upperHeight)
		lower := m.panel(m.mnemonicTitle("Active worktree", 'A'), m.focus == paneDetails, m.renderV310Lower(width-4, lowerHeight-3), width-1, lowerHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + lower
	}
	leftBody := m.renderEntries(width-4, max(5, upperHeight/2-2)); if m.mode == modeBranches { leftBody = m.renderDirectionalBranches(width-4, max(5, upperHeight/2-2)) }
	left := m.panel(m.mnemonicTitle("Navigator", 'N'), m.focus == paneNavigator, leftBody, width-1, max(7, upperHeight/2))
	rightTitle := m.mnemonicTitle("Worktrees", 'W'); rightBody := m.renderWorktreePane(width-4, max(4, upperHeight/2-3)); if m.mode == modeBranches && m.branchContext { rightTitle = "Branch context"; rightBody = m.renderBranchContext(width-4, max(4, upperHeight/2-3)) }
	right := m.panel(rightTitle, m.focus == paneWorktrees, rightBody, width-1, max(6, upperHeight/2))
	lower := m.panel(m.mnemonicTitle("Active worktree", 'A'), m.focus == paneDetails, m.renderV310Lower(width-4, lowerHeight-3), width-1, lowerHeight)
	return left + "\n" + right + "\n" + lower
}

func (m V310Model) renderDirectionalBranches(width, height int) string {
	if len(m.entries) == 0 { return dimStyle.Render("No matching items") }
	visible := max(3, height); start := 0; if m.cursor >= visible { start = m.cursor-visible+1 }; end := min(len(m.entries), start+visible)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		e := m.entries[i]; marker := "  "; if i == m.cursor { marker = "> " }
		label := e.Label
		if e.IsFolder { label = e.Name+"/" } else if br, ok := m.branchesByID[e.ID]; ok { label = br.Name + m.prDirections(br.Name) }
		line := truncate(marker+label, max(8, width)); if i == m.cursor { line = selectedStyle.Render(line) }; lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m V310Model) prDirections(branch string) string {
	var incoming, outgoing []string
	for _, pr := range m.snapshot.PullRequests {
		if pr.BaseBranch == branch { incoming = append(incoming, fmt.Sprintf("< [PR #%d]", pr.Number)) }
		if pr.HeadBranch == branch { outgoing = append(outgoing, fmt.Sprintf("[PR #%d] >", pr.Number)) }
	}
	parts := append(incoming, outgoing...); if len(parts) == 0 { return "" }; return "  " + strings.Join(parts, "  ")
}

func (m V310Model) renderBranchContext(width, height int) string {
	br, _ := m.branchByName(m.branchName)
	infoTitle := m.subTitle("Branch", 'B', m.branchSubFocus == 0)
	commitTitle := m.subTitle("Commits", 'C', m.branchSubFocus == 1)
	messageTitle := m.subTitle("Message", 'M', m.branchSubFocus == 2)
	lines := []string{infoTitle, truncate(br.Name, width), truncate("HEAD: "+br.SHA, width), truncate("PR: "+strings.TrimSpace(m.prDirections(br.Name)), width), "", commitTitle}
	commitRows := max(3, min(8, height/3)); start := 0; if m.branchCommitCursor >= commitRows { start = m.branchCommitCursor-commitRows+1 }; end := min(len(m.branchCommits), start+commitRows)
	if len(m.branchCommits) == 0 { lines = append(lines, dimStyle.Render("Loading/no commits")) } else { for i := start; i < end; i++ { c := m.branchCommits[i]; mark := "  "; if i == m.branchCommitCursor { mark = "> " }; row := truncate(fmt.Sprintf("%s%s  %s", mark, shortSHA(c.SHA), c.Subject), width); if i == m.branchCommitCursor { row = selectedStyle.Render(row) }; lines = append(lines, row) } }
	lines = append(lines, "", messageTitle)
	if c, ok := m.selectedBranchContextCommit(); ok {
		msgLines := strings.Split(c.Message, "\n"); available := max(2, height-len(lines)); maxScroll := maxInt(0, len(msgLines)-available); scroll := m.branchMessageScroll; if scroll > maxScroll { scroll = maxScroll }; endMsg := min(len(msgLines), scroll+available); for _, line := range msgLines[scroll:endMsg] { lines = append(lines, truncate(line, width)) }
	}
	if len(lines) > height && height > 0 { lines = lines[:height] }
	return strings.Join(lines, "\n")
}

func (m V310Model) renderV310Lower(width, height int) string {
	base := m.renderLowerPane(width, max(3, height))
	if len(m.changes) == 0 { return base }
	lines := strings.Split(base, "\n")
	lines = append(lines, dimStyle.Render(strings.Repeat("─", min(width, 48))), activeStyle.Render("Dirty files"))
	room := max(2, height-len(lines)); visible := min(room, len(m.changes)); start := 0; if m.changeCursor >= visible { start = m.changeCursor-visible+1 }; end := min(len(m.changes), start+visible)
	for i := start; i < end; i++ { ch := m.changes[i]; mark := "  "; if i == m.changeCursor && m.focus == paneDetails { mark = "> " }; row := truncate(mark+changeStatus(ch)+"  "+ch.Path, width); if i == m.changeCursor && m.focus == paneDetails { row = selectedStyle.Render(row) }; lines = append(lines, row) }
	if height > 0 && len(lines) > height { lines = lines[len(lines)-height:] }
	return strings.Join(lines, "\n")
}

func changeStatus(ch worktree.Change) string {
	if ch.Conflicted { return "UU conflict" }
	if ch.Untracked { return "?? untracked" }
	parts := []string{}
	if ch.IndexStatus != ' ' { parts = append(parts, "staged:"+string(ch.IndexStatus)) }
	if ch.WorktreeStatus != ' ' { parts = append(parts, "work:"+string(ch.WorktreeStatus)) }
	if len(parts) == 0 { return "changed" }
	return strings.Join(parts, ",")
}

func (m V310Model) renderV310Footer() string {
	if m.confirmDiscard || m.confirmCheckout { return "[y] confirm  [n/Esc] cancel" }
	if m.mode == modeBranches && m.branchContext && m.focus == paneWorktrees {
		base := "[Shift+Tab] Branch/Commits/Message  [Alt+N/W/B/C/M/A] jump"
		if m.branchSubFocus == 1 { base += "  [↑/↓] commit  [c] new worktree  [x] checkout detached" }
		if m.branchSubFocus == 2 { base += "  [↑/↓ PgUp/PgDn] scroll message" }
		return base + "  [Tab] main pane  [q] quit"
	}
	if m.focus == paneDetails && len(m.changes) > 0 { return "[↑/↓] dirty file  [s/u] stage/unstage  [d] diff  [z] stash  [r] discard tracked  [Alt+N/W/A] jump  [Tab] main pane  [q] quit" }
	return m.renderFooter() + "  [Alt+N/W/A] jump"
}

func (m V310Model) mnemonicTitle(title string, mnemonic rune) string {
	idx := strings.Index(strings.ToLower(title), strings.ToLower(string(mnemonic))); if idx < 0 { return title }
	runes := []rune(title); for i, r := range runes { if strings.EqualFold(string(r), string(mnemonic)) { return string(runes[:i]) + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")).Render(string(r)) + string(runes[i+1:]) } }
	return title
}
func (m V310Model) subTitle(title string, mnemonic rune, focused bool) string { t := m.mnemonicTitle(title, mnemonic); if focused { return activeStyle.Render(t) }; return dimStyle.Render(t) }
func (m V310Model) branchByName(name string) (struct{ Name, SHA string }, bool) { for _, br := range m.snapshot.Branches { if br.Name == name { return struct{ Name, SHA string }{br.Name, br.SHA}, true } }; return struct{ Name, SHA string }{}, false }
func (m V310Model) focusName310() string { if m.focus == paneNavigator { return "Navigator" }; if m.focus == paneWorktrees { if m.branchContext { return "Branch context" }; return "Worktrees" }; return "Active worktree" }
