package worktree

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (m *Manager) PreparePullRequest(ctx context.Context, number int, selectedSHA string) (string, error) {
	if number<=0||!validSHA(selectedSHA){return "",fmt.Errorf("invalid pull request identity")}
	pullRef:="refs/pull/"+strconv.Itoa(number)+"/head";localRef:="refs/gh-tree/pr-"+strconv.Itoa(number)
	if _,err:=m.runner.Run(ctx,m.repoRoot,"git","fetch","--no-tags",m.remote,"+"+pullRef+":"+localRef);err!=nil{return "",fmt.Errorf("fetch PR #%d: %w",number,err)}
	sha,err:=m.revParse(ctx,m.repoRoot,localRef+"^{commit}");if err!=nil{return "",fmt.Errorf("resolve PR #%d: %w",number,err)}
	if !strings.EqualFold(sha,selectedSHA){return "",fmt.Errorf("PR #%d changed: selected %s, fetched %s; refresh and retry",number,selectedSHA,sha)}
	return localRef,nil
}

// CheckoutRef advances a named local worktree branch to an already verified
// ref. It only fast-forwards an existing branch. Local commits or a force-pushed
// remote therefore cause a refusal rather than silently losing history.
func (m *Manager) CheckoutRef(ctx context.Context,path,revision,branch string)(Info,error){
	branch=strings.TrimSpace(branch);if branch==""{return Info{},fmt.Errorf("local branch name is required")}
	if _,err:=m.runner.Run(ctx,m.repoRoot,"git","check-ref-format","--branch",branch);err!=nil{return Info{},fmt.Errorf("invalid branch %q: %w",branch,err)}
	status,err:=m.Status(ctx,path);if err!=nil{return Info{},err};if status.Info.Primary{return Info{},fmt.Errorf("refusing to retarget the primary worktree")};if !status.Clean{return Info{},fmt.Errorf("worktree is dirty; refusing checkout: %s",path)}
	targetSHA,err:=m.revParse(ctx,m.repoRoot,revision+"^{commit}");if err!=nil{return Info{},fmt.Errorf("resolve target revision: %w",err)}
	infos,err:=m.List(ctx);if err!=nil{return Info{},err};for _,info:=range infos{if !samePath(info.Path,path)&&sameBranch(info.Branch,branch){return Info{},fmt.Errorf("branch %q is already checked out in %s",branch,info.Path)}}

	oldSHA:="";if sha,e:=m.revParse(ctx,m.repoRoot,"refs/heads/"+branch+"^{commit}");e==nil{oldSHA=sha}
	if oldSHA!=""&&!strings.EqualFold(oldSHA,targetSHA){
		if _,e:=m.runner.Run(ctx,m.repoRoot,"git","merge-base","--is-ancestor",oldSHA,targetSHA);e!=nil{return Info{},fmt.Errorf("refusing to rewind branch %q from %s to %s; it contains commits not present in the selected PR head. Preserve the branch or choose a new local branch",branch,oldSHA,targetSHA)}
	}
	if sameBranch(status.Info.Branch,branch){
		if _,err:=m.runner.Run(ctx,path,"git","reset","--keep",targetSHA);err!=nil{return Info{},fmt.Errorf("move branch %q: %w",branch,err)}
	}else if oldSHA!=""{
		if _,err:=m.runner.Run(ctx,m.repoRoot,"git","branch","-f",branch,targetSHA);err!=nil{return Info{},fmt.Errorf("move branch %q: %w",branch,err)}
		if _,err:=m.runner.Run(ctx,path,"git","switch",branch);err!=nil{return Info{},fmt.Errorf("switch to branch %q: %w",branch,err)}
	}else{
		if _,err:=m.runner.Run(ctx,path,"git","switch","-c",branch,targetSHA);err!=nil{return Info{},fmt.Errorf("create branch %q: %w",branch,err)}
	}
	result,err:=m.revParse(ctx,path,"HEAD^{commit}");if err!=nil{return Info{},err};if !strings.EqualFold(result,targetSHA){return Info{},fmt.Errorf("checkout verification failed: expected %s, got %s",targetSHA,result)}
	if err:=m.requireClean(ctx,path);err!=nil{return Info{},fmt.Errorf("checkout did not remain clean: %w",err)}
	infos,err=m.List(ctx);if err!=nil{return Info{},err};for _,info:=range infos{if samePath(info.Path,path){return info,nil}};return Info{},fmt.Errorf("worktree disappeared after checkout")
}
