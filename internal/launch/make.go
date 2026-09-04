package launch

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type MakeProvider struct{}
func (MakeProvider) Name() string { return "make" }
func (MakeProvider) Detect(root string) bool { _,ok:=findMakefile(root);return ok }
func (MakeProvider) Discover(root string)([]Candidate,error){
	root,err:=cleanRoot(root);if err!=nil{return nil,err};path,ok:=findMakefile(root);if !ok{return nil,nil};file,err:=os.Open(path);if err!=nil{return nil,err};defer file.Close()
	set:=map[string]bool{};scanner:=bufio.NewScanner(file);for scanner.Scan(){raw:=scanner.Text();line:=strings.TrimSpace(raw);if line==""||strings.HasPrefix(line,"#")||strings.HasPrefix(raw,"\t")||strings.HasPrefix(line,"."){continue};idx:=strings.Index(line,":");if idx<=0{continue};left:=strings.TrimSpace(line[:idx]);right:=line[idx+1:];if strings.HasPrefix(strings.TrimSpace(right),"=")||strings.ContainsAny(left,"=$%")||strings.Contains(left,"/"){continue};for _,name:=range strings.Fields(left){if validMakeTarget(name){set[name]=true}}};if err:=scanner.Err();err!=nil{return nil,err}
	names:=make([]string,0,len(set));for name:=range set{names=append(names,name)};sort.Strings(names);out:=make([]Candidate,0,len(names));for _,name:=range names{out=append(out,Candidate{Provider:"make",ID:name,Path:[]string{"make",name},Targets:[]string{name},Command:"make"})};return out,nil
}
func (MakeProvider) Build(root string,c Candidate)(Invocation,error){root,err:=cleanRoot(root);if err!=nil{return Invocation{},err};if len(c.Targets)==0{return Invocation{},fmt.Errorf("make target stack is empty")};targets:=append([]string(nil),c.Targets...);for _,target:=range targets{if !validMakeTarget(target){return Invocation{},fmt.Errorf("invalid make target %q",target)}};cmd:=c.Command;if cmd==""{cmd="make"};return Invocation{Provider:"make",Name:strings.Join(targets,":"),Command:cmd,Args:targets,Dir:root},nil}
func findMakefile(root string)(string,bool){for _,name:=range []string{"Makefile","makefile","GNUmakefile"}{path:=filepath.Join(root,name);if info,err:=os.Stat(path);err==nil&&!info.IsDir(){return path,true}};return "",false}
func validMakeTarget(name string)bool{name=strings.TrimSpace(name);if name==""||strings.HasPrefix(name,"."){return false};for _,r:=range name{if unicode.IsLetter(r)||unicode.IsDigit(r)||r=='_'||r=='-'||r=='.'{continue};return false};return true}
