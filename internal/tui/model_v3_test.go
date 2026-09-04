package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/diff"
	"github.com/Hans-Einar/gh-tree/internal/launch"
)

func TestV3MakeLaunchStackPreservesSelectionOrder(t *testing.T){
	m:=Model{launchCandidates:[]launch.Candidate{
		{Provider:"make",ID:"all",Targets:[]string{"all"}},
		{Provider:"make",ID:"clean",Targets:[]string{"clean"}},
		{Provider:"make",ID:"install",Targets:[]string{"install"}},
	},launchCursor:2,launchSelected:[]int{1,0,2}}
	c,ok:=m.selectedLaunchCandidate();if!ok{t.Fatal("no candidate")};got:=strings.Join(c.Targets,":");if got!="clean:all:install"{t.Fatalf("stack=%q",got)}
}

func TestV3DiffViewRendersFilesAndPatch(t *testing.T){
	m:=NewModel("Hans-Einar/gh-tree",&fakeBackend{},nil,nil,"","",nil);m.mode=modeDiff;m.width=120;m.height=30;m.diffSource="worktree";m.diffResult=diff.Result{Files:[]diff.FileChange{{Path:"internal/tui/model.go",Status:"M",Additions:3,Deletions:1}},Patch:"@@ -1 +1 @@\n-old\n+new\n"}
	view:=m.View();for _,want:=range []string{"Changed files","internal/tui/model.go","Patch","+new"}{if!strings.Contains(view,want){t.Fatalf("view missing %q: %q",want,view)}}
}

func TestV3FooterAdvertisesLaunchKeys(t *testing.T){m:=NewModel("Hans-Einar/gh-tree",&fakeBackend{},nil,nil,"","",nil);footer:=m.renderFooter();if!strings.Contains(footer,"[F5] run")||!strings.Contains(footer,"[Ctrl+F5] launch"){t.Fatalf("footer=%q",footer)}}

func TestV3DiffEscapeReturnsToPreviousMode(t *testing.T){m:=NewModel("Hans-Einar/gh-tree",&fakeBackend{},nil,nil,"","",nil);m.previousMode=modePullRequests;m.mode=modeDiff;m=updateModel(t,m,keyMsg(tea.KeyEsc,""));if m.mode!=modePullRequests{t.Fatalf("mode=%v want PR mode",m.mode)}}
