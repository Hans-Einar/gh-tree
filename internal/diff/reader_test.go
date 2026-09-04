package diff

import (
	"context"
	"strings"
	"testing"
)

type fakeRunner struct{ responses map[string][]byte }
func (f fakeRunner)Run(_ context.Context,_ string,name string,args ...string)([]byte,error){key:=name+" "+strings.Join(args," ");for prefix,out:=range f.responses{if strings.HasPrefix(key,prefix){return out,nil}};return nil,nil}

func TestParseNameStatusRename(t *testing.T){data:=[]byte("M\x00a.go\x00R100\x00old.txt\x00new.txt\x00");files,err:=parseNameStatus(data);if err!=nil{t.Fatal(err)};if len(files)!=2||files[0].Path!="a.go"||files[1].OldPath!="old.txt"||files[1].Path!="new.txt"{t.Fatalf("files=%#v",files)}}
func TestReaderBoundsPatchAndStats(t *testing.T){runner:=fakeRunner{responses:map[string][]byte{"git diff --name-status":[]byte("M\x00a.go\x00"),"git diff --numstat":[]byte("12\t3\ta.go\x00"),"git diff --patch":[]byte(strings.Repeat("x",100))}};result,err:=NewReader(runner).Worktree(context.Background(),"/repo",20);if err!=nil{t.Fatal(err)};if len(result.Files)!=1||result.Files[0].Additions!=12||result.Files[0].Deletions!=3{t.Fatalf("files=%#v",result.Files)};if !result.Truncated||!strings.Contains(result.Patch,"diff truncated"){t.Fatalf("result=%#v",result)}}
func TestCommitUsesFirstParent(t *testing.T){sha:=strings.Repeat("a",40);parent:=strings.Repeat("b",40);runner:=fakeRunner{responses:map[string][]byte{"git rev-list --parents -n 1":[]byte(sha+" "+parent+"\n"),"git diff "+parent+" "+sha+" --name-status":[]byte{},"git diff "+parent+" "+sha+" --numstat":[]byte{},"git diff "+parent+" "+sha+" --patch":[]byte("patch\n")}};result,err:=NewReader(runner).Commit(context.Background(),"/repo",sha,100);if err!=nil{t.Fatal(err)};if result.Label!="commit "+sha+" vs parent"{t.Fatalf("label=%q",result.Label)}}
