package worktree

import (
	"context"
	"fmt"
	"strings"
)

func (m *Manager) ensureBranchStart(ctx context.Context,branch string)(string,bool,error){branch=strings.TrimSpace(branch);if branch==""{return "",false,fmt.Errorf("branch is required")};if _,err:=m.runner.Run(ctx,m.repoRoot,"git","check-ref-format","--branch",branch);err!=nil{return "",false,fmt.Errorf("invalid branch %q: %w",branch,err)};if _,err:=m.runner.Run(ctx,m.repoRoot,"git","show-ref","--verify","--quiet","refs/heads/"+branch);err==nil{return branch,true,nil};remoteRef:="refs/remotes/"+m.remote+"/"+branch;refspec:="+refs/heads/"+branch+":"+remoteRef;if _,err:=m.runner.Run(ctx,m.repoRoot,"git","fetch","--no-tags",m.remote,refspec);err!=nil{return "",false,fmt.Errorf("fetch branch %q: %w",branch,err)};return remoteRef,false,nil}

func (m *Manager) CreateBranchWorktree(ctx context.Context,path,branch string)(Info,error){return m.CreateBranchWorktreeAs(ctx,path,branch,branch)}

// CreateBranchWorktreeAs creates a worktree at source branch history while
// allowing a different local branch name. This avoids Git's one-worktree-per-
// branch conflict when, for example, main is already checked out in the primary.
func (m *Manager) CreateBranchWorktreeAs(ctx context.Context,path,sourceBranch,targetBranch string)(Info,error){
	start,local,err:=m.ensureBranchStart(ctx,sourceBranch);if err!=nil{return Info{},err}
	targetBranch=strings.TrimSpace(targetBranch)
	if targetBranch==""{return m.Create(ctx,CreateRequest{Path:path,StartPoint:start,Detach:true})}
	if targetBranch==sourceBranch&&local{return m.Create(ctx,CreateRequest{Path:path,StartPoint:start})}
	return m.Create(ctx,CreateRequest{Path:path,StartPoint:start,Branch:targetBranch})
}

func (m *Manager) CheckoutBranchWorktree(ctx context.Context,path,branch string)(Info,error){start,local,err:=m.ensureBranchStart(ctx,branch);if err!=nil{return Info{},err};if local{return m.Checkout(ctx,CheckoutRequest{Path:path,Revision:start,Branch:branch})};return m.Checkout(ctx,CheckoutRequest{Path:path,Revision:start,Branch:branch,Create:true})}

// CheckoutBranchDetached deploys the exact selected branch tip into an existing
// secondary worktree without checking out/moving the branch ref itself. This is
// useful for a stable user-owned test worktree even when the branch is already
// checked out in an agent/Codex worktree.
func (m *Manager) CheckoutBranchDetached(ctx context.Context,path,branch string)(Info,error){start,_,err:=m.ensureBranchStart(ctx,branch);if err!=nil{return Info{},err};return m.Checkout(ctx,CheckoutRequest{Path:path,Revision:start,Detach:true})}
