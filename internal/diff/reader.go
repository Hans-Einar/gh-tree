package diff

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/process"
)

const DefaultMaxPatchBytes = 512 * 1024

type FileChange struct {
	Path      string
	OldPath   string
	Status    string
	Additions int
	Deletions int
	Binary    bool
}

type Result struct {
	Label     string
	Files     []FileChange
	Patch     string
	Truncated bool
}

type Reader struct { runner process.Runner }
func NewReader(r process.Runner)*Reader{return &Reader{runner:r}}

func (r *Reader) Worktree(ctx context.Context,path string,maxBytes int)(Result,error){return r.run(ctx,path,"worktree vs HEAD",[]string{"diff"},maxBytes)}
func (r *Reader) Staged(ctx context.Context,path string,maxBytes int)(Result,error){return r.run(ctx,path,"staged vs HEAD",[]string{"diff","--cached"},maxBytes)}
func (r *Reader) Range(ctx context.Context,path,base,head string,maxBytes int)(Result,error){if strings.TrimSpace(base)==""||strings.TrimSpace(head)==""{return Result{},fmt.Errorf("base and head are required")};return r.run(ctx,path,base+"..."+head,[]string{"diff",base+"..."+head},maxBytes)}
func (r *Reader) Commit(ctx context.Context,path,sha string,maxBytes int)(Result,error){if strings.TrimSpace(sha)==""{return Result{},fmt.Errorf("commit SHA/ref is required")};out,err:=r.runner.Run(ctx,path,"git","rev-list","--parents","-n","1",sha);if err!=nil{return Result{},fmt.Errorf("resolve commit parent: %w",err)};fields:=strings.Fields(string(out));if len(fields)==0{return Result{},fmt.Errorf("commit %q was not found",sha)};if len(fields)>=2{return r.run(ctx,path,"commit "+sha+" vs parent",[]string{"diff",fields[1],fields[0]},maxBytes)};return r.run(ctx,path,"root commit "+sha,[]string{"diff-tree","--root","--no-commit-id","-r",fields[0]},maxBytes)}

func (r *Reader) run(ctx context.Context,path,label string,base []string,maxBytes int)(Result,error){if maxBytes<=0{maxBytes=DefaultMaxPatchBytes};nameArgs:=append(append([]string{},base...),"--name-status","-z","--find-renames","--no-ext-diff");nameOut,err:=r.runner.Run(ctx,path,"git",nameArgs...);if err!=nil{return Result{},fmt.Errorf("load diff file list: %w",err)};files,err:=parseNameStatus(nameOut);if err!=nil{return Result{},err};numArgs:=append(append([]string{},base...),"--numstat","-z","--find-renames","--no-ext-diff");numOut,err:=r.runner.Run(ctx,path,"git",numArgs...);if err!=nil{return Result{},fmt.Errorf("load diff statistics: %w",err)};applyNumstat(files,numOut);patchArgs:=append(append([]string{},base...),"--patch","--no-color","--no-ext-diff","--unified=3");patchOut,err:=r.runner.Run(ctx,path,"git",patchArgs...);if err!=nil{return Result{},fmt.Errorf("load diff patch: %w",err)};truncated:=len(patchOut)>maxBytes;if truncated{patchOut=append(append([]byte(nil),patchOut[:maxBytes]...),[]byte("\n\n[gh-tree: diff truncated; narrow the selection or inspect externally]\n")...)};return Result{Label:label,Files:files,Patch:string(patchOut),Truncated:truncated},nil}

func parseNameStatus(data []byte)([]FileChange,error){tokens:=splitNUL(data);var files []FileChange;for i:=0;i<len(tokens);{status:=tokens[i];i++;if status==""{continue};code:=status[:1];if code=="R"||code=="C"{if i+1>=len(tokens){return nil,fmt.Errorf("malformed rename/copy diff record")};oldPath,newPath:=tokens[i],tokens[i+1];i+=2;files=append(files,FileChange{Path:newPath,OldPath:oldPath,Status:status});continue};if i>=len(tokens){return nil,fmt.Errorf("malformed diff record")};files=append(files,FileChange{Path:tokens[i],Status:status});i++};return files,nil}
func splitNUL(data []byte)[]string{raw:=strings.Split(string(data),"\x00");if len(raw)>0&&raw[len(raw)-1]==""{raw=raw[:len(raw)-1]};return raw}
func applyNumstat(files []FileChange,data []byte){tokens:=splitNUL(data);for i:=0;i<len(tokens);i++{fields:=strings.Split(tokens[i],"\t");if len(fields)<3{continue};add,del,path:=fields[0],fields[1],strings.Join(fields[2:],"\t");if path==""&&i+2<len(tokens){path=tokens[i+2];i+=2};for j:=range files{if files[j].Path!=path&&files[j].OldPath!=path{continue};if add=="-"||del=="-"{files[j].Binary=true}else{files[j].Additions,_=strconv.Atoi(add);files[j].Deletions,_=strconv.Atoi(del)};break}}}
