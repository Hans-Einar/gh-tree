package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

func TestStatusAndCommitHistory(t *testing.T) {
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot)
	status,err:=m.Status(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)}
	if !status.Clean||status.Info.Branch!=fixture.branch{t.Fatalf("status=%#v",status)}
	if err:=os.WriteFile(filepath.Join(fixture.targetRoot,"new.txt"),[]byte("new\n"),0o600);err!=nil{t.Fatal(err)}
	status,err=m.Status(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)}
	if status.Clean||status.Untracked!=1{t.Fatalf("dirty status=%#v",status)}
	if err:=m.StageAll(context.Background(),fixture.targetRoot);err!=nil{t.Fatal(err)}
	sha,err:=m.Commit(context.Background(),fixture.targetRoot,"v2 commit");if err!=nil{t.Fatal(err)}
	if len(sha)!=40{t.Fatalf("sha=%q",sha)}
	commits,err:=m.Commits(context.Background(),fixture.targetRoot,"HEAD",10,0);if err!=nil{t.Fatal(err)}
	if len(commits)==0||commits[0].Subject!="v2 commit"||commits[0].SHA!=sha{t.Fatalf("commits=%#v",commits)}
}

func TestCreateAndRetargetWorktree(t *testing.T) {
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot);path:=filepath.Join(filepath.Dir(fixture.mainRoot),"created")
	info,err:=m.Create(context.Background(),CreateRequest{Path:path,StartPoint:fixture.initialSHA,Branch:"local/created"});if err!=nil{t.Fatal(err)}
	if info.Branch!="local/created"||info.Head!=fixture.initialSHA{t.Fatalf("created=%#v",info)}
	if err:=os.WriteFile(filepath.Join(path,"dirty.txt"),[]byte("keep\n"),0o600);err!=nil{t.Fatal(err)}
	if _,err:=m.Checkout(context.Background(),CheckoutRequest{Path:path,Revision:fixture.prSHA,Detach:true});err==nil||!strings.Contains(err.Error(),"dirty"){t.Fatalf("expected dirty refusal, got %v",err)}
	if err:=os.Remove(filepath.Join(path,"dirty.txt"));err!=nil{t.Fatal(err)}
	info,err=m.Checkout(context.Background(),CheckoutRequest{Path:path,Revision:fixture.prSHA,Detach:true});if err!=nil{t.Fatal(err)}
	if !info.Detached||info.Head!=fixture.prSHA{t.Fatalf("retargeted=%#v",info)}
}

func TestPreparePullRequestAndCreatePRWorktree(t *testing.T) {
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot)
	ref,err:=m.PreparePullRequest(context.Background(),60,fixture.prSHA);if err!=nil{t.Fatal(err)}
	if ref!="refs/gh-tree/pr-60"{t.Fatalf("ref=%q",ref)}
	path:=filepath.Join(filepath.Dir(fixture.mainRoot),"pr-worktree")
	info,err:=m.Create(context.Background(),CreateRequest{Path:path,StartPoint:ref,Branch:"local/pr-60"});if err!=nil{t.Fatal(err)}
	if info.Head!=fixture.prSHA||info.Branch!="local/pr-60"{t.Fatalf("info=%#v",info)}
}

func TestCheckoutRefReusesBranchForNewPRHead(t *testing.T) {
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot)
	path:=filepath.Join(filepath.Dir(fixture.mainRoot),"pr-follow")
	ref,err:=m.PreparePullRequest(context.Background(),60,fixture.prSHA);if err!=nil{t.Fatal(err)}
	if _,err:=m.Create(context.Background(),CreateRequest{Path:path,StartPoint:ref,Branch:"local/pr-follow"});err!=nil{t.Fatal(err)}

	git(t,fixture.mainRoot,"switch","feature/Concept1/ui-box")
	if err:=os.WriteFile(filepath.Join(fixture.mainRoot,"second.txt"),[]byte("second PR commit\n"),0o600);err!=nil{t.Fatal(err)}
	git(t,fixture.mainRoot,"add","second.txt");git(t,fixture.mainRoot,"commit","-m","second PR commit")
	newSHA:=git(t,fixture.mainRoot,"rev-parse","HEAD")
	git(t,fixture.mainRoot,"push","origin","+HEAD:refs/pull/60/head");git(t,fixture.mainRoot,"switch","main")
	ref,err=m.PreparePullRequest(context.Background(),60,newSHA);if err!=nil{t.Fatal(err)}
	info,err:=m.CheckoutRef(context.Background(),path,ref,"local/pr-follow");if err!=nil{t.Fatal(err)}
	if info.Head!=newSHA||info.Branch!="local/pr-follow"{t.Fatalf("followed info=%#v want %s",info,newSHA)}
}

func TestCreateBranchWorktreeAsAvoidsPrimaryBranchConflict(t *testing.T) {
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot)
	path:=filepath.Join(filepath.Dir(fixture.mainRoot),"main-snapshot")
	info,err:=m.CreateBranchWorktreeAs(context.Background(),path,"main","gh-tree/main-snapshot");if err!=nil{t.Fatal(err)}
	if info.Branch!="gh-tree/main-snapshot"||info.Head!=fixture.initialSHA{t.Fatalf("info=%#v",info)}
	if got:=git(t,fixture.mainRoot,"branch","--show-current");got!="main"{t.Fatalf("primary worktree changed branch to %q",got)}
}

func TestParseBranchStatus(t *testing.T) {
	upstream,ahead,behind:=parseBranchStatus("## feature...origin/feature [ahead 2, behind 3]")
	if upstream!="origin/feature"||ahead!=2||behind!=3{t.Fatalf("got %q %d %d",upstream,ahead,behind)}
}
