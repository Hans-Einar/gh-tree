package worktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

func TestSelectiveStageAndUnstage(t *testing.T){fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot);path:=filepath.Join(fixture.targetRoot,"selective.txt");if err:=os.WriteFile(path,[]byte("hello\n"),0o600);err!=nil{t.Fatal(err)};changes,err:=m.Changes(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)};if len(changes)!=1||!changes[0].Untracked{t.Fatalf("changes=%#v",changes)};if err:=m.StagePaths(context.Background(),fixture.targetRoot,"selective.txt");err!=nil{t.Fatal(err)};changes,err=m.Changes(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)};if len(changes)!=1||changes[0].IndexStatus!='A'{t.Fatalf("staged=%#v",changes)};if err:=m.UnstagePaths(context.Background(),fixture.targetRoot,"selective.txt");err!=nil{t.Fatal(err)};changes,err=m.Changes(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)};if len(changes)!=1||!changes[0].Untracked{t.Fatalf("unstaged=%#v",changes)}}
func TestStashPushAndPopPreservesChanges(t *testing.T){fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot);path:=filepath.Join(fixture.targetRoot,"payload.txt");if err:=os.WriteFile(path,[]byte("local edit\n"),0o600);err!=nil{t.Fatal(err)};sha,err:=m.StashPush(context.Background(),fixture.targetRoot,"gh-tree test",false);if err!=nil{t.Fatal(err)};if len(sha)!=40{t.Fatalf("stash sha=%q",sha)};status,err:=m.Status(context.Background(),fixture.targetRoot);if err!=nil{t.Fatal(err)};if !status.Clean{t.Fatalf("status after stash=%#v",status)};if err:=m.StashApply(context.Background(),fixture.targetRoot,"stash@{0}",true);err!=nil{t.Fatal(err)};data,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)};if !strings.Contains(string(data),"local edit"){t.Fatalf("data=%q",data)}}
func TestStageRejectsEscapingPath(t *testing.T){fixture:=setupRepository(t,false);m:=NewManager(process.ExecRunner{},fixture.mainRoot);if err:=m.StagePaths(context.Background(),fixture.targetRoot,"../outside");err==nil{t.Fatal("expected escaping path rejection")}}
