package worktree

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Change struct {
	Path           string
	OldPath        string
	IndexStatus    byte
	WorktreeStatus byte
	Untracked      bool
	Conflicted     bool
}

func (m *Manager) Changes(ctx context.Context,path string)([]Change,error){if _,err:=m.Status(ctx,path);err!=nil{return nil,err};out,err:=m.runner.Run(ctx,path,"git","status","--porcelain=v1","-z","--untracked-files=all");if err!=nil{return nil,fmt.Errorf("list worktree changes: %w",err)};tokens:=strings.Split(string(out),"\x00");var changes []Change;for i:=0;i<len(tokens);i++{record:=tokens[i];if record==""{continue};if len(record)<4{return nil,fmt.Errorf("unexpected git status record")};x,y:=record[0],record[1];name:=record[3:];change:=Change{Path:name,IndexStatus:x,WorktreeStatus:y,Untracked:x=='?'&&y=='?',Conflicted:x=='U'||y=='U'||(x=='A'&&y=='A')||(x=='D'&&y=='D')};if (x=='R'||x=='C')&&i+1<len(tokens){change.OldPath=tokens[i+1];i++};changes=append(changes,change)};return changes,nil}

func (m *Manager) StagePaths(ctx context.Context,path string,paths ...string)error{safe,err:=safeRelativePaths(paths);if err!=nil{return err};args:=[]string{"add","--"};args=append(args,safe...);if _,err:=m.runner.Run(ctx,path,"git",args...);err!=nil{return fmt.Errorf("stage files: %w",err)};return nil}
func (m *Manager) UnstagePaths(ctx context.Context,path string,paths ...string)error{safe,err:=safeRelativePaths(paths);if err!=nil{return err};args:=[]string{"restore","--staged","--"};args=append(args,safe...);if _,err:=m.runner.Run(ctx,path,"git",args...);err!=nil{return fmt.Errorf("unstage files: %w",err)};return nil}
func safeRelativePaths(paths []string)([]string,error){if len(paths)==0{return nil,fmt.Errorf("at least one path is required")};out:=make([]string,0,len(paths));for _,p:=range paths{p=strings.TrimSpace(p);if p==""||filepath.IsAbs(p){return nil,fmt.Errorf("path must be repository-relative: %q",p)};clean:=filepath.Clean(p);if clean==".."||strings.HasPrefix(clean,".."+string(filepath.Separator)){return nil,fmt.Errorf("path escapes repository: %q",p)};out=append(out,clean)};return out,nil}

func (m *Manager) StashPush(ctx context.Context,path,message string,includeUntracked bool)(string,error){status,err:=m.Status(ctx,path);if err!=nil{return "",err};if status.Clean{return "",fmt.Errorf("worktree is clean; nothing to stash")};args:=[]string{"stash","push"};if includeUntracked{args=append(args,"--include-untracked")};if strings.TrimSpace(message)!=""{args=append(args,"-m",message)};if _,err:=m.runner.Run(ctx,path,"git",args...);err!=nil{return "",fmt.Errorf("stash changes: %w",err)};sha,err:=m.revParse(ctx,path,"refs/stash^{commit}");if err!=nil{return "",fmt.Errorf("resolve created stash: %w",err)};return sha,nil}
func (m *Manager) StashApply(ctx context.Context,path,ref string,pop bool)error{if strings.TrimSpace(ref)==""{ref="stash@{0}"};command:="apply";if pop{command="pop"};if _,err:=m.runner.Run(ctx,path,"git","stash",command,"--index",ref);err!=nil{return fmt.Errorf("stash %s: %w",command,err)};return nil}
