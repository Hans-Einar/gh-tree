package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

func TestRestorePathsDiscardsTrackedWorkingTreeEdit(t *testing.T){
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot);path:=filepath.Join(fixture.targetRoot,"payload.txt")
	if err:=os.WriteFile(path,[]byte("temporary edit\n"),0o600);err!=nil{t.Fatal(err)}
	if err:=m.RestorePaths(context.Background(),fixture.targetRoot,"payload.txt");err!=nil{t.Fatal(err)}
	status,err:=m.Status(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)};if !status.Clean{t.Fatalf("status=%#v",status)}
}

func TestRestorePathsRefusesUntrackedDeletion(t *testing.T){
	fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot);path:=filepath.Join(fixture.targetRoot,"scratch.txt")
	if err:=os.WriteFile(path,[]byte("keep me\n"),0o600);err!=nil{t.Fatal(err)}
	err:=m.RestorePaths(context.Background(),fixture.targetRoot,"scratch.txt");if err==nil||!strings.Contains(err.Error(),"refusing to delete untracked"){t.Fatalf("err=%v",err)}
	if _,err:=os.Stat(path);err!=nil{t.Fatalf("untracked file disappeared: %v",err)}
}
