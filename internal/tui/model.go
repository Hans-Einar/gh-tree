package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/app"
	"github.com/Hans-Einar/gh-tree/internal/config"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/tree"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

type Backend interface {
	Load(ctx context.Context, repo string) (app.Snapshot, error)
	Deploy(ctx context.Context, pr ghapi.PullRequest, target config.WorktreeTarget) (worktree.Deployment, error)
}

// v2Backend is intentionally optional so the original v1 model tests and
// third-party embedders only implementing Backend keep working.
type v2Backend interface {
	WorktreeStatus(context.Context, string) (worktree.Status, error)
	CreateWorktree(context.Context, worktree.CreateRequest) (worktree.Info, error)
	CreatePRWorktree(context.Context, ghapi.PullRequest, string, string) (worktree.Info, error)
	CreateBranchWorktree(context.Context, string, string) (worktree.Info, error)
	CheckoutWorktree(context.Context, worktree.CheckoutRequest) (worktree.Info, error)
	CheckoutPRWorktree(context.Context, ghapi.PullRequest, string, string) (worktree.Info, error)
	CheckoutBranchWorktree(context.Context, string, string) (worktree.Info, error)
	Fetch(context.Context, string) error
	Pull(context.Context, string) error
	StageAll(context.Context, string) error
	Commit(context.Context, string, string) (string, error)
	Push(context.Context, string, bool) error
	NewBranch(context.Context, string, string, string) (worktree.Info, error)
	Commits(context.Context, string, string, int, int) ([]worktree.Commit, error)
	CommitsForPullRequest(context.Context, ghapi.PullRequest, int, int) ([]worktree.Commit, error)
	CommitsForBranch(context.Context, string, int, int) ([]worktree.Commit, error)
	CreatePullRequest(context.Context, string, string, string, string, string, bool) (string, error)
}

type mode int
const (
	modePullRequests mode = iota
	modeBranches
	modeCommits
)

type pane int
const (
	paneNavigator pane = iota
	paneWorktrees
	paneDetails
)

type dialog int
const (
	dialogNone dialog = iota
	dialogTargetPicker
	dialogConfirm
	dialogCreateWorktree
	dialogCheckoutConfirm
	dialogNewBranch
	dialogCommit
	dialogPushConfirm
	dialogCreatePR
)

type snapshotMsg struct { snapshot app.Snapshot; err error }
type deploymentMsg struct { deployment worktree.Deployment; err error }
type wtStatusMsg struct { path string; status worktree.Status; err error }
type wtOpMsg struct { kind string; info worktree.Info; sha string; text string; err error }
type commitsMsg struct { commits []worktree.Commit; append bool; err error }

type Model struct {
	repo          string
	backend       Backend
	prefixes      []string
	targets       []config.WorktreeTarget
	configPath    string
	saveFolder    func(string) error
	snapshot      app.Snapshot
	mode          mode
	previousMode  mode
	folder        string
	query         string
	searching     bool
	entries       []tree.Entry
	paths         []tree.PathItem
	prsByID       map[string]ghapi.PullRequest
	branchesByID  map[string]ghapi.Branch
	cursor        int
	width         int
	height        int
	loading       bool
	status        string
	dialog        dialog
	targetCursor  int
	pendingTarget config.WorktreeTarget
	pendingPR     ghapi.PullRequest
	pendingBranch ghapi.Branch
	deploying     bool

	focus          pane
	worktreeCursor int
	activeWorktree string
	worktreeStatus worktree.Status
	haveWTStatus   bool
	busy           bool

	commits        []worktree.Commit
	commitCursor   int
	commitScroll   int
	commitSkip     int
	commitSource   string

	inputA         string
	inputB         string
	inputC         string
	inputField     int
	pendingRevision string
	pendingDetach  bool
}

func NewModel(repo string, backend Backend, prefixes []string, targets []config.WorktreeTarget, configPath, savedFolder string, saveFolder func(string) error) Model {
	return Model{
		repo: repo, backend: backend, prefixes: append([]string(nil), prefixes...),
		targets: append([]config.WorktreeTarget(nil), targets...), configPath: configPath,
		saveFolder: saveFolder, folder: strings.Trim(savedFolder, "/"),
		prsByID: make(map[string]ghapi.PullRequest), branchesByID: make(map[string]ghapi.Branch),
		loading: true, status: "Loading GitHub state…", previousMode: modePullRequests,
	}
}

func (m Model) Init() tea.Cmd { return m.refreshCmd() }
func (m Model) refreshCmd() tea.Cmd { return func() tea.Msg { snapshot, err := m.backend.Load(context.Background(), m.repo); return snapshotMsg{snapshot: snapshot, err: err} } }
func (m Model) deployCmd(pr ghapi.PullRequest, target config.WorktreeTarget) tea.Cmd { return func() tea.Msg { d, err := m.backend.Deploy(context.Background(), pr, target); return deploymentMsg{deployment: d, err: err} } }

func (m Model) v2() (v2Backend, bool) { b, ok := m.backend.(v2Backend); return b, ok }
func (m Model) statusCmd(path string) tea.Cmd {
	b, ok := m.v2(); if !ok || path == "" { return nil }
	return func() tea.Msg { s, err := b.WorktreeStatus(context.Background(), path); return wtStatusMsg{path: path, status: s, err: err} }
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case snapshotMsg:
		m.loading = false
		if msg.err != nil { m.status = "Refresh failed: " + msg.err.Error(); return m, nil }
		m.snapshot = msg.snapshot
		oldFolder := m.folder
		m.rebuild()
		if m.folder != oldFolder { m.persistFolder() }
		m.syncActiveWorktree()
		m.status = fmt.Sprintf("Loaded %d open PRs, %d branches, %d worktrees", len(m.snapshot.PullRequests), len(m.snapshot.Branches), len(m.snapshot.Worktrees))
		return m, m.statusCmd(m.activeWorktree)
	case deploymentMsg:
		m.deploying = false; m.dialog = dialogNone
		if msg.err != nil { m.status = "Deployment failed: " + msg.err.Error(); return m, nil }
		m.status = fmt.Sprintf("✓ %s · PR #%d · %s · %s", msg.deployment.TargetName, msg.deployment.PRNumber, msg.deployment.Branch, msg.deployment.SHA)
		for i := range m.snapshot.Worktrees { if sameDisplayPath(m.snapshot.Worktrees[i].Path, msg.deployment.Path) { m.snapshot.Worktrees[i].Head = msg.deployment.SHA; m.snapshot.Worktrees[i].Branch = msg.deployment.Branch; m.snapshot.Worktrees[i].Detached = false } }
		return m, m.statusCmd(m.activeWorktree)
	case wtStatusMsg:
		if msg.path != m.activeWorktree { return m, nil }
		if msg.err != nil { m.haveWTStatus = false; m.status = "Worktree status failed: " + msg.err.Error(); return m, nil }
		m.worktreeStatus, m.haveWTStatus = msg.status, true
		return m, nil
	case wtOpMsg:
		m.busy = false; m.dialog = dialogNone
		if msg.err != nil { m.status = strings.Title(msg.kind) + " failed: " + msg.err.Error(); return m, nil }
		if msg.info.Path != "" { m.activeWorktree = msg.info.Path }
		if msg.text != "" { m.status = "✓ " + msg.text } else if msg.sha != "" { m.status = "✓ " + msg.kind + " · " + msg.sha } else { m.status = "✓ " + msg.kind }
		return m, m.refreshCmd()
	case commitsMsg:
		m.busy = false
		if msg.err != nil { m.status = "Commit history failed: " + msg.err.Error(); return m, nil }
		if msg.append { m.commits = append(m.commits, msg.commits...) } else { m.commits = msg.commits; m.commitCursor = 0; m.commitScroll = 0 }
		m.commitSkip = len(m.commits)
		m.status = fmt.Sprintf("Loaded %d commits", len(m.commits))
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m *Model) syncActiveWorktree() {
	if len(m.snapshot.Worktrees) == 0 { m.activeWorktree = ""; m.worktreeCursor = 0; return }
	for i, wt := range m.snapshot.Worktrees {
		if m.activeWorktree != "" && sameDisplayPath(wt.Path, m.activeWorktree) { m.worktreeCursor = i; return }
	}
	for i, wt := range m.snapshot.Worktrees { if wt.Current { m.activeWorktree = wt.Path; m.worktreeCursor = i; return } }
	m.activeWorktree = m.snapshot.Worktrees[0].Path; m.worktreeCursor = 0
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" { return m, tea.Quit }
	if m.deploying || m.busy { return m, nil }
	if m.dialog != dialogNone { return m.updateDialog(msg) }
	if m.searching { return m.updateSearch(msg) }

	if msg.String() == "tab" { m.focus = (m.focus + 1) % 3; m.status = "Focus: " + m.focusName(); return m, nil }
	if msg.String() == "shift+tab" { m.focus = (m.focus + 2) % 3; m.status = "Focus: " + m.focusName(); return m, nil }
	if msg.String() == "q" { return m, tea.Quit }
	if msg.String() == "r" { m.loading = true; m.status = "Refreshing GitHub and worktree state…"; return m, m.refreshCmd() }
	if msg.String() == "p" { m.mode = modePullRequests; m.focus = paneNavigator; m.cursor = 0; m.rebuild(); m.persistFolder(); return m, nil }
	if msg.String() == "b" { m.mode = modeBranches; m.focus = paneNavigator; m.cursor = 0; m.rebuild(); m.persistFolder(); return m, nil }

	if m.mode == modeCommits { return m.updateCommitKeys(msg) }
	if m.focus == paneWorktrees { return m.updateWorktreeKeys(msg) }
	if m.focus == paneDetails { return m.updateDetailsKeys(msg) }
	return m.updateNavigatorKeys(msg)
}

func (m Model) updateNavigatorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k": if m.cursor > 0 { m.cursor-- }
	case "down", "j": if m.cursor+1 < len(m.entries) { m.cursor++ }
	case "home", "g": m.cursor = 0
	case "end", "G": if len(m.entries) > 0 { m.cursor = len(m.entries)-1 }
	case "enter":
		if entry, ok := m.currentEntry(); ok { if entry.IsFolder { m.folder = entry.Path; m.cursor = 0; m.rebuild(); m.persistFolder() } else { m.status = m.selectionStatus(entry) } }
	case "backspace": if m.folder != "" { m.folder = tree.Parent(m.folder); m.cursor = 0; m.rebuild(); m.persistFolder() }
	case "/": m.searching = true; m.status = "Type to filter; Enter accepts, Ctrl+U clears"
	case "w": return m.beginDeployment()
	case "h": return m.beginHistory(false)
	}
	return m, nil
}

func (m Model) updateWorktreeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.snapshot.WorktreesEnabled { m.status = "Worktree operations require running inside this repository"; return m, nil }
	switch msg.String() {
	case "up", "k": if m.worktreeCursor > 0 { m.worktreeCursor-- }
	case "down", "j": if m.worktreeCursor+1 < len(m.snapshot.Worktrees) { m.worktreeCursor++ }
	case "enter", "a":
		if wt, ok := m.selectedWorktree(); ok { m.activeWorktree = wt.Path; m.status = "Active worktree: " + wt.Path; return m, m.statusCmd(wt.Path) }
	case "c": return m.beginCreateWorktree()
	case "x": return m.beginCheckout()
	case "f": return m.runSimpleWT("fetch")
	case "p": return m.runSimpleWT("pull")
	case "P": return m.beginPush()
	case "m": return m.beginCommit()
	case "n": return m.beginNewBranch()
	case "o": return m.beginCreatePR()
	case "h": return m.beginHistory(true)
	}
	return m, nil
}

func (m Model) updateDetailsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h": return m.beginHistory(true)
	case "enter": if m.activeWorktree != "" { return m, m.statusCmd(m.activeWorktree) }
	}
	return m, nil
}

func (m Model) updateCommitKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.focus == paneDetails {
		switch msg.String() {
		case "up", "k": if m.commitScroll > 0 { m.commitScroll-- }
		case "down", "j": m.commitScroll++
		case "pgup": m.commitScroll = maxInt(0, m.commitScroll-10)
		case "pgdown": m.commitScroll += 10
		case "esc", "backspace": m.mode = m.previousMode; m.focus = paneNavigator; m.rebuild()
		}
		return m, nil
	}
	switch msg.String() {
	case "up", "k": if m.commitCursor > 0 { m.commitCursor--; m.commitScroll = 0 }
	case "down", "j": if m.commitCursor+1 < len(m.commits) { m.commitCursor++; m.commitScroll = 0 }
	case "home", "g": m.commitCursor = 0; m.commitScroll = 0
	case "end", "G": if len(m.commits)>0 { m.commitCursor=len(m.commits)-1; m.commitScroll=0 }
	case "L": return m.loadMoreCommits()
	case "c": return m.beginCreateFromCommit()
	case "n": return m.beginBranchFromCommit()
	case "esc", "backspace": m.mode = m.previousMode; m.focus = paneNavigator; m.rebuild()
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter": m.searching = false; m.status = "Filter: " + displayQuery(m.query)
	case "ctrl+u": m.query = ""; m.cursor=0; m.rebuild()
	case "backspace": runes:=[]rune(m.query); if len(runes)>0 { m.query=string(runes[:len(runes)-1]); m.cursor=0; m.rebuild() }
	default: if msg.Type == tea.KeyRunes { m.query += string(msg.Runes); m.cursor=0; m.rebuild() }
	}
	return m,nil
}

func (m Model) updateDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.dialog == dialogTargetPicker || m.dialog == dialogConfirm { return m.updateDeployDialog(msg) }
	if msg.String()=="esc" { m.dialog=dialogNone; m.status="Cancelled"; return m,nil }
	if m.dialog == dialogCheckoutConfirm {
		if msg.String()=="y" || msg.String()=="Y" { return m.executeCheckout() }
		if msg.String()=="n" || msg.String()=="N" || msg.String()=="q" { m.dialog=dialogNone; m.status="Checkout cancelled" }
		return m,nil
	}
	if m.dialog == dialogPushConfirm {
		if msg.String()=="y" || msg.String()=="Y" { return m.executePush() }
		if msg.String()=="n" || msg.String()=="N" || msg.String()=="q" { m.dialog=dialogNone; m.status="Push cancelled" }
		return m,nil
	}
	if msg.String()=="tab" { m.inputField = (m.inputField+1)%m.dialogFields(); return m,nil }
	if msg.String()=="shift+tab" { m.inputField = (m.inputField+m.dialogFields()-1)%m.dialogFields(); return m,nil }
	if msg.String()=="enter" {
		if m.inputField+1 < m.dialogFields() { m.inputField++; return m,nil }
		switch m.dialog {
		case dialogCreateWorktree: return m.executeCreateWorktree()
		case dialogNewBranch: return m.executeNewBranch()
		case dialogCommit: return m.executeCommit()
		case dialogCreatePR: return m.executeCreatePR()
		}
	}
	m.editDialogText(msg)
	return m,nil
}

func (m Model) updateDeployDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.dialog {
	case dialogTargetPicker:
		switch msg.String() {
		case "esc","q": m.dialog=dialogNone; m.status="Deployment cancelled"
		case "up","k": if m.targetCursor>0 { m.targetCursor-- }
		case "down","j": if m.targetCursor+1<len(m.targets) { m.targetCursor++ }
		case "enter": if len(m.targets)>0 { m.pendingTarget=m.targets[m.targetCursor]; m.dialog=dialogConfirm }
		}
	case dialogConfirm:
		switch msg.String() {
		case "y","Y": m.dialog=dialogNone; m.deploying=true; m.status=fmt.Sprintf("Deploying PR #%d at %s…",m.pendingPR.Number,m.pendingPR.HeadSHA); return m,m.deployCmd(m.pendingPR,m.pendingTarget)
		case "n","N","esc","q": m.dialog=dialogNone; m.status="Deployment cancelled"
		}
	}
	return m,nil
}

func (m *Model) editDialogText(msg tea.KeyMsg) {
	var target *string
	switch m.inputField { case 0: target=&m.inputA; case 1: target=&m.inputB; default: target=&m.inputC }
	if msg.String()=="backspace" { r:=[]rune(*target); if len(r)>0 { *target=string(r[:len(r)-1]) }; return }
	if msg.Type==tea.KeyRunes { *target += string(msg.Runes) }
}
func (m Model) dialogFields() int { if m.dialog==dialogCreateWorktree { return 2 }; if m.dialog==dialogCreatePR { return 3 }; return 1 }

func (m Model) beginDeployment() (tea.Model, tea.Cmd) {
	if m.mode != modePullRequests { m.status="Select a PR before deploying"; return m,nil }
	entry,ok:=m.currentEntry(); if !ok||entry.IsFolder { m.status="Select a PR before deploying"; return m,nil }
	pr,ok:=m.prsByID[entry.ID]; if !ok { m.status="Selected item is not a PR"; return m,nil }
	// v2: when there is no legacy target, guide the user into the interactive
	// worktree pane instead of asking them to edit JSON.
	if len(m.targets)==0 { m.pendingPR=pr; m.focus=paneWorktrees; m.status="No legacy target needed: press c to create a worktree or select one and x to check out this PR"; return m,nil }
	m.pendingPR=pr; m.targetCursor=0
	if len(m.targets)==1 { m.pendingTarget=m.targets[0]; m.dialog=dialogConfirm } else { m.dialog=dialogTargetPicker }
	return m,nil
}

func (m Model) beginCreateWorktree() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); _=b; if !ok { m.status="This backend does not support interactive worktrees"; return m,nil }
	basePath := ""
	if len(m.snapshot.Worktrees)>0 { basePath = m.snapshot.Worktrees[0].Path }
	suffix := "work"
	m.pendingPR=ghapi.PullRequest{}; m.pendingBranch=ghapi.Branch{}; m.pendingRevision=""; m.pendingDetach=false
	if pr,ok:=m.currentPR(); ok { m.pendingPR=pr; suffix="pr-"+strconv.Itoa(pr.Number); m.inputB="gh-tree/pr-"+strconv.Itoa(pr.Number) } else if br,ok:=m.currentBranch(); ok { m.pendingBranch=br; suffix=sanitizeSuffix(br.Name); m.inputB=br.Name } else if m.haveWTStatus { m.pendingRevision=m.worktreeStatus.Info.Head; suffix="snapshot"; m.inputB="gh-tree/snapshot" }
	if basePath!="" { m.inputA=filepath.Join(filepath.Dir(basePath), filepath.Base(basePath)+"-"+suffix) } else { m.inputA="" }
	m.inputField=0; m.dialog=dialogCreateWorktree; m.status="Create worktree: edit path and branch"
	return m,nil
}

func (m Model) executeCreateWorktree() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { m.dialog=dialogNone; m.status="Interactive worktrees unavailable"; return m,nil }
	path,branch:=strings.TrimSpace(m.inputA),strings.TrimSpace(m.inputB)
	if path=="" { m.status="Worktree path is required"; return m,nil }
	m.busy=true; m.dialog=dialogNone; m.status="Creating worktree…"
	pr:=m.pendingPR; br:=m.pendingBranch; revision:=m.pendingRevision
	return m,func() tea.Msg {
		var info worktree.Info; var err error
		switch { case pr.Number>0: info,err=b.CreatePRWorktree(context.Background(),pr,path,branch); case br.Name!="": info,err=b.CreateBranchWorktree(context.Background(),path,br.Name); default: info,err=b.CreateWorktree(context.Background(),worktree.CreateRequest{Path:path,StartPoint:revision,Branch:branch,Detach:branch==""}) }
		return wtOpMsg{kind:"create worktree",info:info,text:"created "+path,err:err}
	}
}

func (m Model) beginCheckout() (tea.Model, tea.Cmd) {
	if m.activeWorktree=="" { m.status="Activate a worktree first"; return m,nil }
	if m.haveWTStatus && m.worktreeStatus.Info.Primary { m.status="Primary worktree is protected; create/select a secondary worktree"; return m,nil }
	m.pendingPR=ghapi.PullRequest{}; m.pendingBranch=ghapi.Branch{}; m.pendingRevision=""
	if pr,ok:=m.currentPR(); ok { m.pendingPR=pr; m.inputA="gh-tree/pr-"+strconv.Itoa(pr.Number) } else if br,ok:=m.currentBranch(); ok { m.pendingBranch=br; m.inputA=br.Name } else { m.status="Select a PR or branch in the navigator first"; return m,nil }
	m.dialog=dialogCheckoutConfirm; m.status="Confirm checkout into active worktree"
	return m,nil
}

func (m Model) executeCheckout() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { m.dialog=dialogNone; return m,nil }
	path:=m.activeWorktree; pr:=m.pendingPR; br:=m.pendingBranch; localBranch:=m.inputA
	m.busy=true; m.dialog=dialogNone; m.status="Changing worktree checkout…"
	return m,func() tea.Msg { var info worktree.Info; var err error; if pr.Number>0 { info,err=b.CheckoutPRWorktree(context.Background(),pr,path,localBranch) } else { info,err=b.CheckoutBranchWorktree(context.Background(),path,br.Name) }; return wtOpMsg{kind:"checkout",info:info,text:"checked out "+coalesce(info.Branch,"detached"),err:err} }
}

func (m Model) runSimpleWT(kind string) (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok||m.activeWorktree=="" { m.status="Activate a worktree first"; return m,nil }
	path:=m.activeWorktree; m.busy=true; m.status=strings.Title(kind)+"…"
	return m,func() tea.Msg { var err error; if kind=="fetch" { err=b.Fetch(context.Background(),path) } else { err=b.Pull(context.Background(),path) }; return wtOpMsg{kind:kind,text:kind+" complete",err:err} }
}

func (m Model) beginPush() (tea.Model, tea.Cmd) {
	if m.activeWorktree==""||!m.haveWTStatus { m.status="Activate a worktree and refresh status first"; return m,nil }
	if m.worktreeStatus.Info.Detached { m.status="Cannot push detached HEAD; create a branch first"; return m,nil }
	m.dialog=dialogPushConfirm; return m,nil
}
func (m Model) executePush() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { m.dialog=dialogNone; return m,nil }
	path:=m.activeWorktree; setUpstream:=m.worktreeStatus.Upstream==""; m.busy=true; m.dialog=dialogNone; m.status="Pushing branch…"
	return m,func() tea.Msg { err:=b.Push(context.Background(),path,setUpstream); return wtOpMsg{kind:"push",text:"push complete",err:err} }
}

func (m Model) beginCommit() (tea.Model, tea.Cmd) {
	if m.activeWorktree=="" { m.status="Activate a worktree first"; return m,nil }
	m.inputA=""; m.inputField=0; m.dialog=dialogCommit; return m,nil
}
func (m Model) executeCommit() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { return m,nil }; message:=strings.TrimSpace(m.inputA); if message=="" { m.status="Commit message is required"; return m,nil }
	path:=m.activeWorktree; m.busy=true; m.dialog=dialogNone; m.status="Staging and committing changes…"
	return m,func() tea.Msg { if err:=b.StageAll(context.Background(),path); err!=nil { return wtOpMsg{kind:"commit",err:err} }; sha,err:=b.Commit(context.Background(),path,message); return wtOpMsg{kind:"commit",sha:sha,err:err} }
}

func (m Model) beginNewBranch() (tea.Model, tea.Cmd) {
	if m.activeWorktree=="" { m.status="Activate a worktree first"; return m,nil }
	m.inputA=""; m.pendingRevision="HEAD"; m.inputField=0; m.dialog=dialogNewBranch; return m,nil
}
func (m Model) beginBranchFromCommit() (tea.Model, tea.Cmd) {
	c,ok:=m.currentCommit(); if !ok { return m,nil }; if m.activeWorktree=="" { m.status="Activate the worktree where the new branch should be checked out"; return m,nil }
	m.inputA=""; m.pendingRevision=c.SHA; m.inputField=0; m.dialog=dialogNewBranch; return m,nil
}
func (m Model) executeNewBranch() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { return m,nil }; name:=strings.TrimSpace(m.inputA); if name=="" { m.status="Branch name is required"; return m,nil }
	path,start:=m.activeWorktree,m.pendingRevision; m.busy=true; m.dialog=dialogNone; m.status="Creating branch…"
	return m,func() tea.Msg { info,err:=b.NewBranch(context.Background(),path,name,start); return wtOpMsg{kind:"new branch",info:info,text:"created branch "+name,err:err} }
}

func (m Model) beginCreatePR() (tea.Model, tea.Cmd) {
	if !m.haveWTStatus||m.worktreeStatus.Info.Detached||m.worktreeStatus.Info.Branch=="" { m.status="Active worktree must be on a branch"; return m,nil }
	m.inputA=m.defaultBaseBranch(); m.inputB=""; m.inputC=""; m.inputField=0; m.dialog=dialogCreatePR; return m,nil
}
func (m Model) executeCreatePR() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { return m,nil }; base,title,body:=strings.TrimSpace(m.inputA),strings.TrimSpace(m.inputB),m.inputC; if base==""||title=="" { m.status="Base branch and title are required"; return m,nil }
	head:=m.worktreeStatus.Info.Branch; repo:=m.repo; m.busy=true; m.dialog=dialogNone; m.status="Creating pull request…"
	return m,func() tea.Msg { url,err:=b.CreatePullRequest(context.Background(),repo,head,base,title,body,true); return wtOpMsg{kind:"create PR",text:url,err:err} }
}

func (m Model) beginHistory(fromWorktree bool) (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { m.status="Commit browser unavailable"; return m,nil }
	m.previousMode=m.mode; m.mode=modeCommits; m.focus=paneNavigator; m.commits=nil; m.commitSkip=0; m.commitCursor=0; m.commitScroll=0; m.busy=true
	if fromWorktree {
		if m.activeWorktree=="" { m.mode=m.previousMode; m.busy=false; m.status="Activate a worktree first"; return m,nil }
		path:=m.activeWorktree; m.commitSource="worktree HEAD"; return m,func() tea.Msg { cs,err:=b.Commits(context.Background(),path,"HEAD",50,0); return commitsMsg{commits:cs,err:err} }
	}
	if pr,ok:=m.currentPR(); ok { m.commitSource=fmt.Sprintf("PR #%d",pr.Number); return m,func() tea.Msg { cs,err:=b.CommitsForPullRequest(context.Background(),pr,50,0); return commitsMsg{commits:cs,err:err} } }
	if br,ok:=m.currentBranch(); ok { m.commitSource="branch "+br.Name; return m,func() tea.Msg { cs,err:=b.CommitsForBranch(context.Background(),br.Name,50,0); return commitsMsg{commits:cs,err:err} } }
	m.mode=m.previousMode; m.busy=false; m.status="Select a PR or branch first"; return m,nil
}
func (m Model) loadMoreCommits() (tea.Model, tea.Cmd) {
	b,ok:=m.v2(); if !ok { return m,nil }; skip:=len(m.commits); m.busy=true
	// Loading more from the active worktree is always deterministic. For PR or
	// branch history, the source is re-derived from the previous navigator selection.
	if strings.HasPrefix(m.commitSource,"worktree") { path:=m.activeWorktree; return m,func() tea.Msg { cs,err:=b.Commits(context.Background(),path,"HEAD",50,skip); return commitsMsg{commits:cs,append:true,err:err} } }
	if pr,ok:=m.currentPR(); ok { return m,func() tea.Msg { cs,err:=b.CommitsForPullRequest(context.Background(),pr,50,skip); return commitsMsg{commits:cs,append:true,err:err} } }
	if br,ok:=m.currentBranch(); ok { return m,func() tea.Msg { cs,err:=b.CommitsForBranch(context.Background(),br.Name,50,skip); return commitsMsg{commits:cs,append:true,err:err} } }
	m.busy=false; return m,nil
}
func (m Model) beginCreateFromCommit() (tea.Model, tea.Cmd) {
	c,ok:=m.currentCommit(); if !ok { return m,nil }; basePath:=""; if len(m.snapshot.Worktrees)>0 { basePath=m.snapshot.Worktrees[0].Path }
	m.pendingPR=ghapi.PullRequest{}; m.pendingBranch=ghapi.Branch{}; m.pendingRevision=c.SHA; m.inputB=""; if basePath!="" { m.inputA=filepath.Join(filepath.Dir(basePath),filepath.Base(basePath)+"-"+shortSHA(c.SHA)) }; m.inputField=0; m.dialog=dialogCreateWorktree; return m,nil
}

func (m *Model) rebuild() {
	if m.mode==modeCommits { return }
	selectedID:=""; if entry,ok:=m.currentEntry();ok { selectedID=entry.ID }
	m.paths=m.buildPaths(); m.folder=tree.ResolveFolder(m.paths,m.folder); m.entries=tree.Entries(m.paths,m.folder,m.query); m.cursor=0
	if selectedID!="" { for i,e:=range m.entries { if e.ID==selectedID { m.cursor=i; break } } }
	if m.cursor>=len(m.entries)&&len(m.entries)>0 { m.cursor=len(m.entries)-1 }
}

func (m *Model) buildPaths() []tree.PathItem {
	m.prsByID=make(map[string]ghapi.PullRequest,len(m.snapshot.PullRequests)); m.branchesByID=make(map[string]ghapi.Branch,len(m.snapshot.Branches))
	if m.mode==modePullRequests {
		paths:=make([]tree.PathItem,0,len(m.snapshot.PullRequests)); for _,pr:=range m.snapshot.PullRequests { id:=fmt.Sprintf("pr:%d",pr.Number); m.prsByID[id]=pr; state:="OPEN"; if pr.IsDraft { state="DRAFT" }; paths=append(paths,tree.PathItem{ID:id,Path:tree.NormalizeBranch(pr.HeadBranch,m.prefixes),Label:fmt.Sprintf("#%d  %s  [%s]",pr.Number,pr.Title,state)}) }; return paths
	}
	paths:=make([]tree.PathItem,0,len(m.snapshot.Branches)); for _,br:=range m.snapshot.Branches { id:="branch:"+br.Name; m.branchesByID[id]=br; label:=br.Name; if pr,ok:=m.prForBranch(br.Name);ok { label+=fmt.Sprintf("  [PR #%d]",pr.Number) }; paths=append(paths,tree.PathItem{ID:id,Path:tree.NormalizeBranch(br.Name,m.prefixes),Label:label}) }; return paths
}

func (m *Model) persistFolder() { if m.saveFolder!=nil { if err:=m.saveFolder(m.folder);err!=nil { m.status="Could not persist folder: "+err.Error() } } }
func (m Model) currentEntry() (tree.Entry,bool) { if m.cursor<0||m.cursor>=len(m.entries){return tree.Entry{},false};return m.entries[m.cursor],true }
func (m Model) currentPR() (ghapi.PullRequest,bool) { e,ok:=m.currentEntry(); if !ok||e.IsFolder{return ghapi.PullRequest{},false};pr,ok:=m.prsByID[e.ID];return pr,ok }
func (m Model) currentBranch() (ghapi.Branch,bool) { e,ok:=m.currentEntry(); if !ok||e.IsFolder{return ghapi.Branch{},false};br,ok:=m.branchesByID[e.ID];return br,ok }
func (m Model) currentCommit() (worktree.Commit,bool) { if m.commitCursor<0||m.commitCursor>=len(m.commits){return worktree.Commit{},false};return m.commits[m.commitCursor],true }
func (m Model) selectedWorktree() (worktree.Info,bool) { if m.worktreeCursor<0||m.worktreeCursor>=len(m.snapshot.Worktrees){return worktree.Info{},false};return m.snapshot.Worktrees[m.worktreeCursor],true }
func (m Model) prForBranch(name string) (ghapi.PullRequest,bool) { for _,pr:=range m.snapshot.PullRequests { if pr.HeadBranch==name { return pr,true } }; return ghapi.PullRequest{},false }
func (m Model) selectionStatus(entry tree.Entry) string { if pr,ok:=m.prsByID[entry.ID];ok{return fmt.Sprintf("Selected PR #%d · %s → %s · %s",pr.Number,pr.HeadBranch,pr.BaseBranch,pr.HeadSHA)};if br,ok:=m.branchesByID[entry.ID];ok{return fmt.Sprintf("Selected branch %s at %s",br.Name,br.SHA)};return "Selected "+entry.Label }
func (m Model) focusName() string { switch m.focus {case paneWorktrees:return "worktrees";case paneDetails:return "details";default:return "navigator"} }
func (m Model) defaultBaseBranch() string { for _,b:=range m.snapshot.Branches { if b.Name=="main" {return "main"} }; for _,b:=range m.snapshot.Branches {if b.Name=="master"{return "master"}}; return "main" }
func displayQuery(q string) string { if q==""{return "(none)"};return q }
func sameDisplayPath(a,b string) bool { return strings.EqualFold(strings.TrimRight(a,"/\\"),strings.TrimRight(b,"/\\")) }
func sanitizeSuffix(v string) string { var out []rune; for _,r:=range v { if unicode.IsLetter(r)||unicode.IsDigit(r)||r=='-'||r=='_' {out=append(out,r)} else if r=='/' {out=append(out,'-')} }; s:=strings.Trim(string(out),"-"); if s==""{return "work"}; if len(s)>32{return s[:32]}; return s }
func coalesce(v, fallback string) string { if strings.TrimSpace(v)==""{return fallback};return v }
func maxInt(a,b int) int {if a>b{return a};return b}
