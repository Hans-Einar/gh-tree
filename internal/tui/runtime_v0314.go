package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/launch"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type v314Backend interface {
	WorktreeStashes(context.Context, string) ([]worktree.Stash, error)
	StashApplyRef(context.Context, string, string, bool) error
	StashDrop(context.Context, string, string) error
	StashPatch(context.Context, string, string) (string, error)
	DeployBranchDetached(context.Context, string, string) (worktree.Info, error)
	DeployPRDetached(context.Context, string, ghapi.PullRequest) (worktree.Info, error)
}

type stashList314Msg struct {
	path  string
	items []worktree.Stash
	err   error
}

type stashMutation314Msg struct {
	kind string
	path string
	err  error
}

type stashPatch314Msg struct {
	ref  string
	text string
	err  error
}

type deployKind314 int

const (
	deployCommit314 deployKind314 = iota
	deployBranch314
	deployPR314
)

type deployTarget314 struct {
	kind     deployKind314
	label    string
	revision string
	branch   string
	pr       ghapi.PullRequest
}

type deployDone314Msg struct {
	info    worktree.Info
	label   string
	stashed bool
	err     error
}

// V314Model makes Active worktree the stable deployment/test target. Selection
// and activation are intentionally different operations: Enter deploys into the
// current Active worktree (after confirmation), while Ctrl+Enter activates an
// already-existing worktree that contains the selected Git object.
type V314Model struct {
	V313Model
	stashes            []worktree.Stash
	stashCursor        int
	stashLoadedPath    string
	stashFocus         bool
	deployConfirm      bool
	deployTarget       deployTarget314
	stashConfirm       string
	stashPatchOpen     bool
	stashPatchText     string
	stashPatchScroll   int
}

func WithV314UX(model Model) V314Model {
	// v0.3.11 accidentally introduced a darker blue for some active elements.
	// v0.3.14 restores one canonical purple accent across headings/selections.
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63"))
	focusedTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Padding(0, 1)
	return V314Model{V313Model: WithV313UX(model)}
}

func (m V314Model) Init() tea.Cmd { return m.V313Model.Init() }

func (m V314Model) v314() (v314Backend, bool) {
	b, ok := m.backend.(v314Backend)
	return b, ok
}

func (m V314Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case stashList314Msg:
		if msg.path != m.activeWorktree {
			return m, nil
		}
		m.stashLoadedPath = msg.path
		if msg.err != nil {
			m.status = "Stash list failed: " + msg.err.Error()
			return m, nil
		}
		m.stashes = msg.items
		if m.stashCursor >= len(m.stashes) {
			m.stashCursor = maxInt(0, len(m.stashes)-1)
		}
		return m, nil
	case stashMutation314Msg:
		m.busy = false
		m.stashConfirm = ""
		if msg.err != nil {
			m.status = titleWord(msg.kind) + " failed: " + msg.err.Error()
			return m, tea.Batch(m.statusCmd(msg.path), m.stashCmd314(msg.path))
		}
		m.status = "✓ " + msg.kind
		return m, tea.Batch(m.statusCmd(msg.path), m.stashCmd314(msg.path), m.changesCmd(msg.path))
	case stashPatch314Msg:
		m.busy = false
		if msg.err != nil {
			m.status = "Show stash failed: " + msg.err.Error()
			return m, nil
		}
		m.stashPatchOpen = true
		m.stashPatchText = msg.text
		m.stashPatchScroll = 0
		m.status = "Inspecting " + msg.ref
		return m, nil
	case deployDone314Msg:
		m.busy = false
		m.deployConfirm = false
		if msg.err != nil {
			m.status = "Deploy failed: " + msg.err.Error()
			return m, tea.Batch(m.statusCmd(m.activeWorktree), m.stashCmd314(m.activeWorktree))
		}
		for i := range m.snapshot.Worktrees {
			if sameDisplayPath(m.snapshot.Worktrees[i].Path, msg.info.Path) {
				m.snapshot.Worktrees[i] = msg.info
			}
		}
		extra := ""
		if msg.stashed {
			extra = " · changes stashed"
		}
		m.status = "✓ deployed " + msg.label + " → " + filepath.Base(msg.info.Path) + extra
		return m, tea.Batch(m.refreshCmd(), m.statusCmd(m.activeWorktree), m.stashCmd314(m.activeWorktree))
	}

	if key, ok := message.(tea.KeyMsg); ok {
		if m.stashPatchOpen {
			return m.handleStashPatch314(key)
		}
		if m.deployConfirm {
			return m.handleDeployConfirm314(key)
		}
		if m.stashConfirm != "" {
			return m.handleStashConfirm314(key)
		}
		if m.dialog == dialogNone && !m.deploying && !m.busy && !m.searching && !m.worktreeModal && !m.confirmDiscard && !m.confirmCheckout {
			if handled, model, cmd := m.handle314Key(key); handled {
				return model, cmd
			}
		}
	}

	oldPath := m.activeWorktree
	updated, cmd := m.V313Model.Update(message)
	inner, ok := updated.(V313Model)
	if !ok {
		return updated, cmd
	}
	m.V313Model = inner
	if oldPath != m.activeWorktree {
		m.stashes = nil
		m.stashCursor = 0
		m.stashLoadedPath = ""
		m.stashFocus = false
		return m, tea.Batch(cmd, m.stashCmd314(m.activeWorktree))
	}
	if msg, ok := message.(wtStatusMsg); ok && msg.path == m.activeWorktree && msg.err == nil && m.stashLoadedPath != m.activeWorktree {
		return m, tea.Batch(cmd, m.stashCmd314(m.activeWorktree))
	}
	return m, cmd
}

func (m V314Model) handle314Key(key tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	k := key.String()
	switch k {
	case "alt+s":
		m.consoleFocus = false
		m.focus = paneDetails
		m.activeSubFocus = 2
		m.stashFocus = true
		m.status = "Focus: Stash"
		return true, m, m.stashCmd314(m.activeWorktree)
	case "alt+l":
		m.consoleFocus = false
		m.focus = paneDetails
		m.activeSubFocus = 1
		m.stashFocus = false
		m.status = "Focus: Launch"
		if m.activeWorktree != "" && m.launchLoadedPath != m.activeWorktree {
			return true, m, m.discoverLaunch311(m.activeWorktree)
		}
		return true, m, nil
	case "alt+a":
		m.stashFocus = false
	case "alt+o":
		m.stashFocus = false
	case "ctrl+enter":
		return true, m.activateSelectedWorktree314()
	}

	if m.mode == modeBranches && m.branchContext && m.focus == paneWorktrees {
		if m.branchSubFocus == 0 && (k == "down" || k == "j") {
			m.branchSubFocus = 1
			m.status = "Focus: Commits"
			return true, m, nil
		}
		if m.branchSubFocus == 1 && m.branchCommitCursor == 0 && (k == "up" || k == "k") {
			m.branchSubFocus = 0
			m.status = "Focus: Branch context"
			return true, m, nil
		}
		if k == "enter" {
			if m.branchSubFocus == 0 {
				if br, ok := m.currentBranchByName311(); ok {
					return true, m.beginDeploy314(deployTarget314{kind: deployBranch314, label: br.Name, branch: br.Name, revision: br.SHA})
				}
			}
			if m.branchSubFocus == 1 {
				if c, ok := m.selectedBranchCommit(); ok {
					return true, m.beginDeploy314(deployTarget314{kind: deployCommit314, label: shortSHA(c.SHA) + " " + c.Subject, revision: c.SHA})
				}
			}
		}
	}

	if m.mode == modePullRequests && m.focus == paneNavigator && k == "enter" {
		if pr, ok := m.currentPR(); ok {
			return true, m.beginDeploy314(deployTarget314{kind: deployPR314, label: fmt.Sprintf("PR #%d %s", pr.Number, pr.HeadBranch), revision: pr.HeadSHA, pr: pr})
		}
	}

	if !m.consoleFocus && m.focus == paneDetails {
		if m.activeSubFocus == 0 && (k == "right" || k == "l") {
			m.activeSubFocus = 2
			m.stashFocus = true
			m.status = "Focus: Stash"
			return true, m, m.stashCmd314(m.activeWorktree)
		}
		if m.activeSubFocus == 1 && (k == "right" || k == "l") {
			m.activeSubFocus = 2
			m.stashFocus = true
			m.status = "Focus: Stash"
			return true, m, m.stashCmd314(m.activeWorktree)
		}
		if m.activeSubFocus == 2 || m.stashFocus {
			return m.handleStashKeys314(key)
		}
	}
	return false, m, nil
}

func (m V314Model) beginDeploy314(target deployTarget314) (tea.Model, tea.Cmd) {
	if m.activeWorktree == "" {
		m.status = "No active worktree"
		return m, nil
	}
	m.deployTarget = target
	m.deployConfirm = true
	m.status = "Confirm deploy to Active worktree"
	return m, nil
}

func (m V314Model) handleDeployConfirm314(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := key.String()
	if k == "esc" || k == "q" || k == "n" {
		m.deployConfirm = false
		m.status = "Deploy cancelled"
		return m, nil
	}
	dirty := m.haveWTStatus && !m.worktreeStatus.Clean
	if dirty {
		if k == "s" || k == "S" {
			return m.executeDeploy314(true)
		}
		return m, nil
	}
	if k == "enter" || k == "y" || k == "Y" {
		return m.executeDeploy314(false)
	}
	return m, nil
}

func (m V314Model) executeDeploy314(stashFirst bool) (tea.Model, tea.Cmd) {
	b, ok := m.v314()
	v2, ok2 := m.v2()
	v3, ok3 := m.v3()
	if !ok || !ok2 || !ok3 || m.activeWorktree == "" {
		m.deployConfirm = false
		m.status = "Deploy backend unavailable"
		return m, nil
	}
	path := m.activeWorktree
	target := m.deployTarget
	status := m.worktreeStatus
	m.busy = true
	m.status = "Deploying " + target.label
	return m, func() tea.Msg {
		stashed := false
		if stashFirst {
			message := worktree.ManagedStashMessage(path, status.Info.Branch, status.Info.Head, time.Now())
			if _, err := v3.StashPush(context.Background(), path, message, true); err != nil {
				return deployDone314Msg{label: target.label, err: fmt.Errorf("stash before deploy: %w", err)}
			}
			stashed = true
		}
		var info worktree.Info
		var err error
		switch target.kind {
		case deployBranch314:
			info, err = b.DeployBranchDetached(context.Background(), path, target.branch)
		case deployPR314:
			info, err = b.DeployPRDetached(context.Background(), path, target.pr)
		default:
			info, err = v2.CheckoutWorktree(context.Background(), worktree.CheckoutRequest{Path: path, Revision: target.revision, Detach: true})
		}
		return deployDone314Msg{info: info, label: target.label, stashed: stashed, err: err}
	}
}

func (m V314Model) activateSelectedWorktree314() (tea.Model, tea.Cmd) {
	branch, sha := "", ""
	if m.mode == modePullRequests {
		if pr, ok := m.currentPR(); ok {
			branch, sha = pr.HeadBranch, pr.HeadSHA
		}
	} else if m.mode == modeBranches {
		if m.branchContext && m.focus == paneWorktrees && m.branchSubFocus == 1 {
			if c, ok := m.selectedBranchCommit(); ok {
				sha = c.SHA
			}
		} else if br, ok := m.currentBranch(); ok {
			branch, sha = br.Name, br.SHA
		}
	}
	if branch == "" && sha == "" {
		m.status = "No Git object selected"
		return m, nil
	}
	best := -1
	for i, wt := range m.snapshot.Worktrees {
		if branch != "" && !wt.Detached && wt.Branch == branch {
			best = i
			break
		}
		if best < 0 && sha != "" && strings.EqualFold(wt.Head, sha) {
			best = i
		}
	}
	if best < 0 {
		m.status = "No local worktree contains the selected branch/commit"
		return m, nil
	}
	wt := m.snapshot.Worktrees[best]
	m.worktreeCursor = best
	m.activeWorktree = wt.Path
	m.persistWorktree()
	m.stashes = nil
	m.stashLoadedPath = ""
	m.status = "Active worktree: " + wt.Path
	return m, tea.Batch(m.statusCmd(wt.Path), m.stashCmd314(wt.Path))
}

func (m V314Model) stashCmd314(path string) tea.Cmd {
	b, ok := m.v314()
	if !ok || path == "" {
		return nil
	}
	return func() tea.Msg {
		items, err := b.WorktreeStashes(context.Background(), path)
		return stashList314Msg{path: path, items: items, err: err}
	}
}

func (m V314Model) selectedStash314() (worktree.Stash, bool) {
	if m.stashCursor < 0 || m.stashCursor >= len(m.stashes) {
		return worktree.Stash{}, false
	}
	return m.stashes[m.stashCursor], true
}

func (m V314Model) handleStashKeys314(key tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch key.String() {
	case "left", "h":
		m.stashFocus = false
		m.activeSubFocus = 1
		m.status = "Focus: Launch"
		return true, m, nil
	case "up", "k":
		if m.stashCursor > 0 {
			m.stashCursor--
		} else {
			m.stashFocus = false
			m.activeSubFocus = 0
			m.status = "Focus: Active worktree"
		}
		return true, m, nil
	case "down", "j":
		if m.stashCursor+1 < len(m.stashes) {
			m.stashCursor++
		}
		return true, m, nil
	case "enter", "d":
		stash, ok := m.selectedStash314()
		if !ok {
			m.status = "No stash selected"
			return true, m, nil
		}
		b, ok := m.v314()
		if !ok {
			return true, m, nil
		}
		m.busy = true
		return true, m, func() tea.Msg {
			text, err := b.StashPatch(context.Background(), m.activeWorktree, stash.Ref)
			return stashPatch314Msg{ref: stash.Ref, text: text, err: err}
		}
	case "a":
		return m.mutateStash314("apply")
	case "p":
		if _, ok := m.selectedStash314(); ok {
			m.stashConfirm = "pop"
			m.status = "Pop selected stash?"
		}
		return true, m, nil
	case "x", "delete":
		if _, ok := m.selectedStash314(); ok {
			m.stashConfirm = "drop"
			m.status = "Drop selected stash permanently?"
		}
		return true, m, nil
	}
	return true, m, nil
}

func (m V314Model) handleStashConfirm314(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.String() == "esc" || key.String() == "n" || key.String() == "q" {
		m.stashConfirm = ""
		m.status = "Stash action cancelled"
		return m, nil
	}
	if key.String() != "enter" && key.String() != "y" {
		return m, nil
	}
	kind := m.stashConfirm
	m.stashConfirm = ""
	_, model, cmd := m.mutateStash314(kind)
	return model, cmd
}

func (m V314Model) mutateStash314(kind string) (bool, tea.Model, tea.Cmd) {
	stash, ok := m.selectedStash314()
	b, have := m.v314()
	if !ok || !have || m.activeWorktree == "" {
		m.status = "No stash selected"
		return true, m, nil
	}
	path := m.activeWorktree
	m.busy = true
	return true, m, func() tea.Msg {
		var err error
		switch kind {
		case "apply":
			err = b.StashApplyRef(context.Background(), path, stash.Ref, false)
		case "pop":
			err = b.StashApplyRef(context.Background(), path, stash.Ref, true)
		case "drop":
			err = b.StashDrop(context.Background(), path, stash.Ref)
		default:
			err = fmt.Errorf("unknown stash action %q", kind)
		}
		return stashMutation314Msg{kind: kind + " " + stash.Ref, path: path, err: err}
	}
}

func (m V314Model) handleStashPatch314(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "q", "backspace":
		m.stashPatchOpen = false
		m.stashPatchText = ""
		m.stashPatchScroll = 0
		m.status = "Focus: Stash"
	case "up", "k":
		if m.stashPatchScroll > 0 {
			m.stashPatchScroll--
		}
	case "down", "j":
		m.stashPatchScroll++
	case "pgup":
		m.stashPatchScroll = maxInt(0, m.stashPatchScroll-10)
	case "pgdown":
		m.stashPatchScroll += 10
	}
	return m, nil
}

func (m V314Model) View() string {
	base := m.viewBase314()
	if m.deployConfirm {
		return overlay311(base, m.renderDeployConfirm314(), m.width, m.height)
	}
	if m.stashConfirm != "" {
		stash, _ := m.selectedStash314()
		verb := titleWord(m.stashConfirm)
		body := fmt.Sprintf("%s %s?\n\n[Enter] Confirm    [Esc] Cancel", verb, stash.Ref)
		if m.stashConfirm == "drop" {
			body = fmt.Sprintf("Drop %s permanently?\nThis removes the stash reference.\n\n[Enter] Drop    [Esc] Cancel", stash.Ref)
		}
		return overlay311(base, m.modalFrame311("Stash", body), m.width, m.height)
	}
	if m.stashPatchOpen {
		return overlay311(base, m.renderStashPatch314(), m.width, m.height)
	}
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

func (m V314Model) viewBase314() string {
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
	if m.mode == modeCommits {
		out.WriteString(m.renderCommitCockpit(width, upper+lower+1))
	} else if m.mode == modeDiff {
		out.WriteString(m.renderDiffCockpit(width, upper+lower+1))
	} else {
		out.WriteString(m.renderCockpit314(width, upper, lower))
	}
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine314())
	out.WriteString("\n")
	out.WriteString(dimStyle.Render("[Tab] next pane  [Ctrl+Shift+Tab] previous pane  [Alt+N/W/A/O] jump  [q] quit"))
	return out.String()
}

func (m V314Model) renderCockpit314(width, upperHeight, lowerHeight int) string {
	if width >= 88 {
		lw, rw := width/2, width-width/2
		navOn := m.focus == paneNavigator && !m.consoleFocus
		leftBody := m.renderEntries(lw-4, upperHeight-5)
		if m.mode == modeBranches {
			leftBody = m.renderDirectionalBranches314(lw-4, upperHeight-5)
		}
		left := m.panel314("Navigator", 'N', navOn, leftBody, m.helpNavigator314(), lw-1, upperHeight)

		wtOn := m.focus == paneWorktrees && !m.consoleFocus
		rightTitle := "Worktrees"
		rightMnemonic := 'W'
		rightBody := m.renderWorktreePane(rw-4, upperHeight-5)
		if m.mode == modeBranches && m.branchContext {
			rightTitle, rightMnemonic = "Branch context", 'B'
			rightBody = m.renderBranchContext314(rw-4, upperHeight-5)
		}
		right := m.panel314(rightTitle, rightMnemonic, wtOn, rightBody, m.helpRight314(), rw-1, upperHeight)

		activeW := width * 11 / 20
		consoleW := width - activeW
		activeOn := m.focus == paneDetails && !m.consoleFocus
		active := m.panel314("Active worktree", 'A', activeOn && m.activeSubFocus == 0, m.renderActive314(activeW-4, lowerHeight-5), m.helpActive314(), activeW-1, lowerHeight)
		console := m.panel314("Console", 'O', m.consoleFocus, m.renderConsole314(consoleW-4, lowerHeight-5), m.helpConsole314(), consoleW-1, lowerHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, active, console)
	}
	return m.V313Model.renderCockpit313(width, upperHeight, lowerHeight)
}

func (m V314Model) panel314(title string, mnemonic rune, active bool, content, help string, width, height int) string {
	style := panelStyle
	if active || (title == "Branch context" && m.focus == paneWorktrees && !m.consoleFocus) || (title == "Active worktree" && m.focus == paneDetails && !m.consoleFocus) || (title == "Console" && m.consoleFocus) {
		style = focusPanelStyle
	}
	titleText := m.heading314(title, mnemonic, active)
	rows := max(1, height-4)
	body := fitRows314(content, rows)
	helpLine := ""
	paneFocused := false
	switch title {
	case "Navigator":
		paneFocused = m.focus == paneNavigator && !m.consoleFocus
	case "Worktrees", "Branch context":
		paneFocused = m.focus == paneWorktrees && !m.consoleFocus
	case "Active worktree":
		paneFocused = m.focus == paneDetails && !m.consoleFocus
	case "Console":
		paneFocused = m.consoleFocus
	}
	if paneFocused {
		helpLine = dimStyle.Render(truncate(help, max(8, width-4)))
	}
	return style.Width(max(10, width-2)).Height(max(3, height-2)).Render(titleText + "\n" + body + "\n" + helpLine)
}

func fitRows314(content string, rows int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > rows {
		lines = lines[:rows]
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m V314Model) heading314(title string, mnemonic rune, active bool) string {
	idx := strings.Index(strings.ToLower(title), strings.ToLower(string(mnemonic)))
	if idx < 0 {
		return title
	}
	pre, hit, post := title[:idx], title[idx:idx+1], title[idx+1:]
	if active {
		base := lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("230"))
		key := lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("11")).Bold(true)
		return base.Render(pre) + key.Render(hit) + base.Render(post)
	}
	key := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	return pre + key.Render(hit) + post
}

func (m V314Model) subheading314(title string, active bool) string {
	if active {
		return lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("230")).Bold(true).Render(" " + title + " ")
	}
	return headerStyle.Render(title)
}

func (m V314Model) renderDirectionalBranches314(width, height int) string {
	if len(m.entries) == 0 {
		return dimStyle.Render("No matching items")
	}
	checked := map[string]bool{}
	for _, wt := range m.snapshot.Worktrees {
		if !wt.Detached && wt.Branch != "" {
			checked[wt.Branch] = true
		}
	}
	visible := max(3, height)
	start := 0
	if m.cursor >= visible {
		start = m.cursor-visible+1
	}
	end := min(len(m.entries), start+visible)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		e := m.entries[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		wtMarker := " "
		label := e.Label
		if e.IsFolder {
			label = e.Name + "/"
		} else if br, ok := m.branchesByID[e.ID]; ok {
			if checked[br.Name] {
				wtMarker = "*"
			}
			label = br.Name + m.prDirections(br.Name)
		}
		line := truncate(cursor+wtMarker+" "+label, max(8, width))
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m V314Model) renderBranchContext314(width, height int) string {
	lines := []string{m.subheading314("Branch", m.focus == paneWorktrees && m.branchSubFocus == 0), truncate(m.branchName, width)}
	if br, ok := m.currentBranchByName311(); ok {
		lines = append(lines, "HEAD: "+shortSHA(br.SHA), "PR: "+strings.TrimSpace(m.prDirections(br.Name)))
	}
	lines = append(lines, "", m.subheading314("Commits", m.focus == paneWorktrees && m.branchSubFocus == 1))
	available := max(3, height-10)
	start := 0
	if m.branchCommitCursor >= available {
		start = m.branchCommitCursor-available+1
	}
	end := min(len(m.branchCommits), start+available)
	for i := start; i < end; i++ {
		c := m.branchCommits[i]
		line := "  " + shortSHA(c.SHA) + "  " + c.Subject
		if i == m.branchCommitCursor {
			line = selectedStyle.Render("> " + shortSHA(c.SHA) + "  " + c.Subject)
		}
		lines = append(lines, truncate(line, width))
	}
	lines = append(lines, "", m.subheading314("Message", m.focus == paneWorktrees && m.branchSubFocus == 2))
	if c, ok := m.selectedBranchCommit(); ok {
		message := strings.Split(c.Message, "\n")
		if m.branchMessageScroll < len(message) {
			message = message[m.branchMessageScroll:]
		}
		for _, line := range message {
			if len(lines) >= height {
				break
			}
			lines = append(lines, truncate(line, width))
		}
	}
	return strings.Join(lines, "\n")
}

func (m V314Model) renderActive314(width, height int) string {
	meta := []string{}
	if s := m.renderSelectionIdentity(width); s != "" {
		meta = append(meta, s)
	}
	if s := m.renderWorktreeStatus(width); s != "" {
		if len(meta) > 0 {
			meta = append(meta, dimStyle.Render(strings.Repeat("─", min(width, 42))))
		}
		meta = append(meta, s)
	}
	metaText := strings.Join(meta, "\n")
	metaLines := strings.Split(metaText, "\n")
	listHeight := max(3, height-len(metaLines)-1)
	if len(metaLines) >= height-2 {
		return fitRows314(metaText, height)
	}
	launchW := max(18, width*3/5)
	stashW := max(14, width-launchW-2)
	launchText := m.renderLaunchList314(launchW, listHeight)
	stashText := m.renderStashList314(stashW, listHeight)
	lists := lipgloss.JoinHorizontal(lipgloss.Top, launchText, "  ", stashText)
	return strings.TrimRight(metaText+"\n"+lists, "\n")
}

func (m V314Model) renderLaunchList314(width, height int) string {
	lines := []string{m.subheading314("Launch", m.focus == paneDetails && !m.consoleFocus && m.activeSubFocus == 1 && !m.stashFocus)}
	room := max(1, height-1)
	if len(m.launchCandidates) == 0 {
		lines = append(lines, dimStyle.Render("No launch points"))
	} else {
		start := 0
		if m.launchCursor >= room {
			start = m.launchCursor-room+1
		}
		end := min(len(m.launchCandidates), start+room)
		for i := start; i < end; i++ {
			line := "  " + launchLabel311(m.launchCandidates[i])
			if i == m.launchCursor && m.activeSubFocus == 1 && !m.stashFocus {
				line = selectedStyle.Render("> " + launchLabel311(m.launchCandidates[i]))
			}
			lines = append(lines, truncate(line, width))
		}
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

func (m V314Model) renderStashList314(width, height int) string {
	lines := []string{m.subheading314("Stash", m.focus == paneDetails && !m.consoleFocus && (m.activeSubFocus == 2 || m.stashFocus))}
	room := max(1, height-1)
	if len(m.stashes) == 0 {
		lines = append(lines, dimStyle.Render("No local stashes"))
	} else {
		start := 0
		if m.stashCursor >= room {
			start = m.stashCursor-room+1
		}
		end := min(len(m.stashes), start+room)
		for i := start; i < end; i++ {
			s := m.stashes[i]
			origin := s.OriginBranch
			if origin == "" {
				origin = "unmanaged"
			}
			label := fmt.Sprintf("%s  %s  %d files", s.Ref, origin, s.Files)
			if i == m.stashCursor && (m.activeSubFocus == 2 || m.stashFocus) {
				label = selectedStyle.Render("> " + label)
			} else {
				label = "  " + label
			}
			lines = append(lines, truncate(label, width))
		}
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

func (m V314Model) renderConsole314(width, height int) string {
	if len(m.consoles) == 0 {
		return dimStyle.Render("No consoles yet.\n[Alt+T] open interactive shell\nLaunch a script with Enter or F5.")
	}
	lines := []string{m.renderConsoleTabs312(width)}
	snap, ok := m.selectedConsole312()
	if !ok {
		return strings.Join(lines, "\n")
	}
	state := m.renderProcessState312(snap.State)
	if snap.Invocation.Provider == "terminal" && (snap.State == launch.StateRunning || snap.State == launch.StateStarting) {
		state = "idle"
	}
	if snap.Invocation.Provider == "terminal" {
		lines = append(lines, truncate(fmt.Sprintf("%s · interactive · %s · pid %d", snap.Invocation.Name, state, snap.PID), width), truncate("cwd: "+snap.Invocation.Dir, width), "")
	} else {
		lines = append(lines, truncate(fmt.Sprintf("%s · %s · pid %d", snap.Invocation.Name, state, snap.PID), width), truncate("$ "+snap.Invocation.Command+" "+strings.Join(snap.Invocation.Args, " "), width), "")
	}
	room := max(1, height-len(lines))
	logs := snap.Lines
	end := len(logs)-m.consoleScroll
	if end < 0 {
		end = 0
	}
	start := maxInt(0, end-room)
	for _, line := range logs[start:end] {
		lines = append(lines, truncate(line, width))
	}
	return strings.Join(lines, "\n")
}

func (m V314Model) helpNavigator314() string {
	if m.mode == modeBranches {
		return "[Enter] context  [Ctrl+Enter] activate existing WT  [h] history  [d] diff"
	}
	return "[Enter] deploy → Active  [Ctrl+Enter] activate existing WT  [h] history  [d] diff"
}

func (m V314Model) helpRight314() string {
	if m.mode == modeBranches && m.branchContext {
		switch m.branchSubFocus {
		case 0:
			return "[Enter] deploy branch → Active  [↓] commits  [Alt+C] commits  [Alt+M] message"
		case 1:
			return "[Enter] deploy commit → Active  [Ctrl+Enter] activate WT  [↑/↓] select"
		default:
			return "[↑/↓ PgUp/PgDn] scroll message  [Alt+B] branch  [Alt+C] commits"
		}
	}
	return "[↑/↓] select  [Enter] details  [Ctrl+Enter] make Active  [c] create"
}

func (m V314Model) helpActive314() string {
	if m.activeSubFocus == 2 || m.stashFocus {
		return "[↑/↓] stash  [Enter/d] inspect  [a] apply  [p] pop  [x] drop  [←] launch"
	}
	if m.activeSubFocus == 1 {
		return "[↑/↓] launch  [Enter/F5] run  [→] stash  [Alt+O] console"
	}
	return "[Enter] change worktree  [↓/Alt+L] launch  [→/Alt+S] stash  [h] history  [d] diff"
}

func (m V314Model) helpConsole314() string {
	if m.selectedInteractive313() {
		return "[Alt+1..9] tabs  [Alt+T] terminal  [Ctrl+C] interrupt  [Alt+A/N/W] leave console"
	}
	return "[Alt+1..9] tabs  [Alt+T] terminal  [Ctrl+C/Shift+F5] stop  [F6] restart"
}

func (m V314Model) renderStatusLine314() string {
	// Persistent interactive shells are alive but idle. They must not make the
	// whole application look busy or animate the word "running" forever.
	active := m.busy || m.loading || m.deploying
	for _, snap := range m.consoles {
		if snap.Invocation.Provider == "terminal" {
			continue
		}
		if snap.State == launch.StateRunning || snap.State == launch.StateStarting {
			active = true
			break
		}
	}
	if active {
		return m.renderStatusLine312()
	}
	return m.renderStatusLine()
}

func (m V314Model) renderDeployConfirm314() string {
	status := "status unavailable"
	dirty := false
	if m.haveWTStatus {
		if m.worktreeStatus.Clean {
			status = "CLEAN"
		} else {
			dirty = true
			status = fmt.Sprintf("DIRTY · staged %d · modified %d · untracked %d · conflicts %d", m.worktreeStatus.Staged, m.worktreeStatus.Modified, m.worktreeStatus.Untracked, m.worktreeStatus.Conflicted)
		}
	}
	body := fmt.Sprintf("Deploy selected Git object into Active worktree?\n\nTarget: %s\nCurrent: %s @ %s\nWorking: %s\n\nSelected: %s", m.activeWorktree, coalesce(m.worktreeStatus.Info.Branch, "DETACHED HEAD"), shortSHA(m.worktreeStatus.Info.Head), status, m.deployTarget.label)
	if dirty {
		body += "\n\n[s] Stash tracked + untracked changes and deploy\n[Esc] Cancel"
	} else {
		body += "\n\n[Enter] Deploy    [Esc] Cancel"
	}
	return m.modalFrame311("Deploy to Active worktree", body)
}

func (m V314Model) renderStashPatch314() string {
	width := m.modalWidth311()
	lines := strings.Split(strings.ReplaceAll(m.stashPatchText, "\r\n", "\n"), "\n")
	visible := min(24, max(8, m.height-10))
	maxScroll := maxInt(0, len(lines)-visible)
	if m.stashPatchScroll > maxScroll {
		m.stashPatchScroll = maxScroll
	}
	end := min(len(lines), m.stashPatchScroll+visible)
	out := make([]string, 0, visible+2)
	for _, line := range lines[m.stashPatchScroll:end] {
		out = append(out, truncate(line, width-6))
	}
	out = append(out, "", dimStyle.Render("[↑/↓ PgUp/PgDn] scroll    [Esc] close"))
	return m.modalFrame311("Stash diff", strings.Join(out, "\n"))
}
