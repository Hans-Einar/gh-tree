package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

func TestCheckoutRefRefusesToLoseLocalCommits(t *testing.T) {
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot)
	path:=filepath.Join(filepath.Dir(fixture.mainRoot),"pr-local-work")
	ref,err:=m.PreparePullRequest(context.Background(),60,fixture.prSHA);if err!=nil{t.Fatal(err)}
	if _,err:=m.Create(context.Background(),CreateRequest{Path:path,StartPoint:ref,Branch:"gh-tree/pr-60-local"});err!=nil{t.Fatal(err)}
	if err:=os.WriteFile(filepath.Join(path,"local.txt"),[]byte("do not lose\n"),0o600);err!=nil{t.Fatal(err)}
	git(t,path,"add","local.txt");git(t,path,"commit","-m","local unpushed work")
	localSHA:=git(t,path,"rev-parse","HEAD")
	_,err=m.CheckoutRef(context.Background(),path,ref,"gh-tree/pr-60-local")
	if err==nil||!strings.Contains(err.Error(),"refusing to rewind"){t.Fatalf("expected history-preserving refusal, got %v",err)}
	if got:=git(t,path,"rev-parse","HEAD");got!=localSHA{t.Fatalf("local branch moved from %s to %s",localSHA,got)}
}
