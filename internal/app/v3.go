package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/diff"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/graph"
	"github.com/Hans-Einar/gh-tree/internal/launch"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

func (s *Service) GraphHistory(ctx context.Context,repo string,limit,skip int)([]worktree.Commit,error){snapshot,err:=s.LoadGraph(ctx,repo,limit,skip);if err!=nil{return nil,err};rows:=graph.Rows(snapshot.Commits);out:=make([]worktree.Commit,0,len(rows));for _,row:=range rows{c:=row.Commit;subject:=strings.TrimSpace(row.Prefix+" "+c.Subject+" "+graphDecorations(c));message:=[]string{c.Subject,"","Git graph decorations:"};for _,d:=range c.Decorations{message=append(message,fmt.Sprintf("- %s: %s",d.Kind,d.Name))};for _,pr:=range c.PRs{state:="OPEN";if pr.Draft{state="DRAFT"};message=append(message,fmt.Sprintf("- PR #%d %s: %s -> %s",pr.Number,state,pr.Head,pr.Base))};for _,wt:=range c.Worktrees{state:=wt.Branch;if wt.Detached{state="DETACHED HEAD"};message=append(message,fmt.Sprintf("- worktree %s: %s",filepath.Base(wt.Path),state))};message=append(message,"",snapshot.RemoteFreshness);out=append(out,worktree.Commit{SHA:c.SHA,Parents:append([]string(nil),c.Parents...),Author:c.Author,Authored:c.Authored,Subject:subject,Message:strings.Join(message,"\n")})};return out,nil}
func graphDecorations(c graph.Commit)string{var parts []string;for _,d:=range c.Decorations{switch d.Kind{case graph.RefHEAD:parts=append(parts,"HEAD");case graph.RefLocal:parts=append(parts,"[L:"+d.Name+"]");case graph.RefRemote:parts=append(parts,"[R:"+d.Name+"]");case graph.RefTag:parts=append(parts,"[T:"+d.Name+"]")}};for _,pr:=range c.PRs{parts=append(parts,fmt.Sprintf("[PR#%d]",pr.Number))};for _,wt:=range c.Worktrees{parts=append(parts,"[WT:"+filepath.Base(wt.Path)+"]")};return strings.Join(parts," ")}

func (s *Service) DiscoverLaunch(ctx context.Context,path string)([]launch.Candidate,error){if _,err:=s.WorktreeStatus(ctx,path);err!=nil{return nil,err};return s.LaunchRegistry.Discover(path)}
func (s *Service) SaveLaunch(ctx context.Context,path,name string,c launch.Candidate,makeDefault bool)error{if _,err:=s.WorktreeStatus(ctx,path);err!=nil{return err};return launch.SaveCandidate(path,name,c,makeDefault)}
func (s *Service) RunLaunch(ctx context.Context,path,name string)(launch.ProcessSnapshot,error){if s.Launcher==nil{return launch.ProcessSnapshot{},fmt.Errorf("launch manager is unavailable")};if _,err:=s.WorktreeStatus(ctx,path);err!=nil{return launch.ProcessSnapshot{},err};cfg,err:=launch.LoadConfig(path);if err!=nil{return launch.ProcessSnapshot{},err};if name==""{name=cfg.Default};if name==""{return launch.ProcessSnapshot{},fmt.Errorf("no default launch point; use Ctrl+F5 to discover one")};inv,err:=cfg.Invocation(path,name,s.LaunchRegistry);if err!=nil{return launch.ProcessSnapshot{},err};if err:=s.Launcher.Start(inv);err!=nil{return launch.ProcessSnapshot{},err};return s.Launcher.Snapshot(),nil}
func (s *Service) RunCandidate(ctx context.Context,path string,c launch.Candidate)(launch.ProcessSnapshot,error){if s.Launcher==nil{return launch.ProcessSnapshot{},fmt.Errorf("launch manager is unavailable")};if _,err:=s.WorktreeStatus(ctx,path);err!=nil{return launch.ProcessSnapshot{},err};inv,err:=s.LaunchRegistry.Build(path,c);if err!=nil{return launch.ProcessSnapshot{},err};if err:=s.Launcher.Start(inv);err!=nil{return launch.ProcessSnapshot{},err};return s.Launcher.Snapshot(),nil}
func (s *Service) StopLaunch()error{if s.Launcher==nil{return nil};return s.Launcher.Stop()}
func (s *Service) RestartLaunch()(launch.ProcessSnapshot,error){if s.Launcher==nil{return launch.ProcessSnapshot{},fmt.Errorf("launch manager is unavailable")};if err:=s.Launcher.Restart();err!=nil{return launch.ProcessSnapshot{},err};return s.Launcher.Snapshot(),nil}
func (s *Service) LaunchSnapshot()launch.ProcessSnapshot{if s.Launcher==nil{return launch.ProcessSnapshot{State:launch.StateExited,ExitCode:-1}};return s.Launcher.Snapshot()}

func (s *Service) DiffWorktree(ctx context.Context,path string)(diff.Result,error){if s.DiffReader==nil{return diff.Result{},fmt.Errorf("diff reader is unavailable")};if _,err:=s.WorktreeStatus(ctx,path);err!=nil{return diff.Result{},err};return s.DiffReader.Worktree(ctx,path,diff.DefaultMaxPatchBytes)}
func (s *Service) DiffStaged(ctx context.Context,path string)(diff.Result,error){if s.DiffReader==nil{return diff.Result{},fmt.Errorf("diff reader is unavailable")};if _,err:=s.WorktreeStatus(ctx,path);err!=nil{return diff.Result{},err};return s.DiffReader.Staged(ctx,path,diff.DefaultMaxPatchBytes)}
func (s *Service) DiffCommit(ctx context.Context,path,sha string)(diff.Result,error){if s.DiffReader==nil{return diff.Result{},fmt.Errorf("diff reader is unavailable")};return s.DiffReader.Commit(ctx,path,sha,diff.DefaultMaxPatchBytes)}
func (s *Service) DiffPullRequest(ctx context.Context,pr ghapi.PullRequest)(diff.Result,error){m,err:=s.requireWorktrees();if err!=nil{return diff.Result{},err};if s.DiffReader==nil||s.LocalRoot==""{return diff.Result{},fmt.Errorf("diff reader is unavailable")};ref,err:=m.PreparePullRequest(ctx,pr.Number,pr.HeadSHA);if err!=nil{return diff.Result{},err};if err:=m.Fetch(ctx,s.LocalRoot);err!=nil{return diff.Result{},err};base:="origin/"+pr.BaseBranch;return s.DiffReader.Range(ctx,s.LocalRoot,base,ref,diff.DefaultMaxPatchBytes)}
func (s *Service) StagePaths(ctx context.Context,path string,paths ...string)error{m,err:=s.requireWorktrees();if err!=nil{return err};return m.StagePaths(ctx,path,paths...)}
func (s *Service) UnstagePaths(ctx context.Context,path string,paths ...string)error{m,err:=s.requireWorktrees();if err!=nil{return err};return m.UnstagePaths(ctx,path,paths...)}
func (s *Service) StashPush(ctx context.Context,path,message string,includeUntracked bool)(string,error){m,err:=s.requireWorktrees();if err!=nil{return "",err};return m.StashPush(ctx,path,message,includeUntracked)}
func (s *Service) StashPop(ctx context.Context,path string)error{m,err:=s.requireWorktrees();if err!=nil{return err};return m.StashApply(ctx,path,"stash@{0}",true)}
func (s *Service) WorktreeChanges(ctx context.Context,path string)([]worktree.Change,error){m,err:=s.requireWorktrees();if err!=nil{return nil,err};return m.Changes(ctx,path)}
func (s *Service) RestorePaths(ctx context.Context,path string,paths ...string)error{m,err:=s.requireWorktrees();if err!=nil{return err};return m.RestorePaths(ctx,path,paths...)}
