package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/config"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

func TestCommitModeTabTogglesOnlyListAndDetails(t *testing.T) {
	t.Parallel()
	m:=NewModel("Hans-Einar/gh-tree",&fakeBackend{},config.DefaultStripPrefixes,nil,"","",nil)
	m.mode=modeCommits;m.focus=paneNavigator
	m=updateModel(t,m,keyMsg(tea.KeyTab,""))
	if m.focus!=paneDetails{t.Fatalf("first Tab focus=%v want details",m.focus)}
	m=updateModel(t,m,keyMsg(tea.KeyTab,""))
	if m.focus!=paneNavigator{t.Fatalf("second Tab focus=%v want commit list",m.focus)}
}

func TestDeployWithoutLegacyTargetsEntersInteractiveWorktreeFlow(t *testing.T) {
	t.Parallel()
	m:=NewModel("Hans-Einar/ponsse",&fakeBackend{},config.DefaultStripPrefixes,nil,"","Concept1",nil)
	m=updateModel(t,m,snapshotMsg{snapshot:testSnapshot()})
	m=updateModel(t,m,runeKey("w"))
	if m.dialog!=dialogNone||m.focus!=paneWorktrees{t.Fatalf("dialog=%v focus=%v",m.dialog,m.focus)}
	if !strings.Contains(strings.ToLower(m.status),"interactive"){t.Fatalf("status=%q",m.status)}
}

func TestWorktreePullKeyDoesNotSwitchToPRMode(t *testing.T) {
	t.Parallel()
	m:=NewModel("Hans-Einar/ponsse",&fakeBackend{},config.DefaultStripPrefixes,nil,"","",nil)
	m.mode=modeBranches
	m.focus=paneWorktrees
	m.snapshot.WorktreesEnabled=true
	m.snapshot.Worktrees=[]worktree.Info{{Path:`C:\work\ponsse-test`,Branch:"local/test",Head:strings.Repeat("a",40)}}
	m.activeWorktree=m.snapshot.Worktrees[0].Path
	m=updateModel(t,m,runeKey("p"))
	if m.mode!=modeBranches{t.Fatalf("worktree p changed mode=%v, want branches",m.mode)}
	if m.focus!=paneWorktrees{t.Fatalf("worktree p changed focus=%v",m.focus)}
}

func TestWorktreeStatusViewShowsRemoteSHAAndDirtyCounts(t *testing.T) {
	t.Parallel()
	sha:=strings.Repeat("d",40);upstreamSHA:=strings.Repeat("e",40)
	m:=NewModel("Hans-Einar/ponsse",&fakeBackend{},config.DefaultStripPrefixes,nil,"","Concept1",nil)
	m=updateModel(t,m,snapshotMsg{snapshot:testSnapshot()})
	m.width=120;m.height=32;m.activeWorktree=`C:\work\ponsse-C1`;m.haveWTStatus=true
	m.worktreeStatus=worktree.Status{Info:worktree.Info{Path:m.activeWorktree,Branch:"local/test",Head:sha},Clean:false,Modified:2,Untracked:1,Upstream:"origin/local/test",UpstreamSHA:upstreamSHA,Ahead:1,Behind:3}
	view:=m.View()
	for _,want:=range []string{sha,upstreamSHA,"modified=2","untracked=1","ahead 1 / behind 3"}{if !strings.Contains(view,want){t.Fatalf("view missing %q: %q",want,view)}}
}

func TestPRWorktreeStatusAssociatesByExactHeadSHA(t *testing.T) {
	t.Parallel()
	sha:=strings.Repeat("f",40)
	m:=NewModel("Hans-Einar/ponsse",&fakeBackend{},config.DefaultStripPrefixes,nil,"","",nil)
	m.snapshot.PullRequests=[]ghapi.PullRequest{{Number:60,HeadBranch:"steering/Concept1/ui-box",BaseBranch:"main",HeadSHA:sha,IsDraft:true}}
	m.width=120;m.height=32;m.activeWorktree=`C:\work\ponsse-pr-60`;m.haveWTStatus=true
	m.worktreeStatus=worktree.Status{Info:worktree.Info{Path:m.activeWorktree,Branch:"gh-tree/pr-60",Head:sha},Clean:true}
	view:=m.View()
	for _,want:=range []string{"#60","DRAFT","HEAD matches PR"}{if !strings.Contains(view,want){t.Fatalf("view missing %q: %q",want,view)}}
}

func TestNarrowLayoutStacksNavigatorWorktreesAndStatus(t *testing.T) {
	t.Parallel()
	m:=NewModel("Hans-Einar/ponsse",&fakeBackend{},config.DefaultStripPrefixes,nil,"","Concept1",nil)
	m=updateModel(t,m,snapshotMsg{snapshot:testSnapshot()});m.width=70;m.height=32
	view:=m.View()
	for _,want:=range []string{"Navigator","Local worktrees","Selection + active worktree"}{if !strings.Contains(view,want){t.Fatalf("narrow view missing %q",want)}}
}
