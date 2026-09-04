package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/app"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

func TestV310PRDirectionRelativeToBranch(t *testing.T){
	m:=V310Model{RuntimeModel:RuntimeModel{Model:Model{snapshot:app.Snapshot{PullRequests:[]ghapi.PullRequest{
		{Number:60,HeadBranch:"feature/ui",BaseBranch:"main"},
		{Number:61,HeadBranch:"topic",BaseBranch:"feature/ui"},
	}}}}}
	if got:=m.prDirections("feature/ui");!strings.Contains(got,"< [PR #61]")||!strings.Contains(got,"[PR #60] >"){t.Fatalf("direction=%q",got)}
	if got:=m.prDirections("main");got!="  < [PR #60]"{t.Fatalf("main direction=%q",got)}
}

func TestV310BranchContextRendersCommitMessage(t *testing.T){
	base:=NewModel("Hans-Einar/ponsse",&fakeBackend{},nil,nil,"","",nil)
	base.mode=modeBranches;base.width=120;base.height=32;base.snapshot=app.Snapshot{Branches:[]ghapi.Branch{{Name:"feature/ui",SHA:"1234567890abcdef"}},PullRequests:[]ghapi.PullRequest{{Number:60,HeadBranch:"feature/ui",BaseBranch:"main"}}};base.rebuild()
	m:=WithV310UX(base);m.branchContext=true;m.branchName="feature/ui";m.focus=paneWorktrees;m.branchCommits=[]worktree.Commit{{SHA:"abcdef1234567890",Subject:"fix layout",Message:"fix layout\n\nlonger explanation"}}
	view:=m.View();for _,want:=range []string{"Branch context","Commits","Message","fix layout","longer explanation","[PR #60] >"}{if!strings.Contains(view,want){t.Fatalf("view missing %q: %q",want,view)}}
}

func TestV310ShiftTabCyclesBranchSubpaneOnly(t *testing.T){
	base:=NewModel("Hans-Einar/ponsse",&fakeBackend{},nil,nil,"","",nil);base.mode=modeBranches;base.focus=paneWorktrees
	m:=WithV310UX(base);m.branchContext=true;m.branchSubFocus=1
	updated,_:=m.Update(tea.KeyMsg{Type:tea.KeyShiftTab});got:=updated.(V310Model)
	if got.branchSubFocus!=2{t.Fatalf("subfocus=%d",got.branchSubFocus)}
	if got.focus!=paneWorktrees{t.Fatalf("main focus moved: %v",got.focus)}
}

func TestV310ChangeStateExplainsDirtyCause(t *testing.T){
	if got:=changeState(worktree.Change{Path:"new.txt",IndexStatus:'?',WorktreeStatus:'?',Untracked:true});got!="?? untracked"{t.Fatalf("untracked=%q",got)}
	if got:=changeState(worktree.Change{Path:"x",IndexStatus:'M',WorktreeStatus:'M'});!strings.Contains(got,"staged:M")||!strings.Contains(got,"work:M"){t.Fatalf("modified=%q",got)}
	if got:=changeState(worktree.Change{Path:"x",Conflicted:true});got!="UU conflict"{t.Fatalf("conflict=%q",got)}
}
