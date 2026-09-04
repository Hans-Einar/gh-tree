package launch

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNPMDiscoveryKeepsColonScriptAsOneInvocation(t *testing.T){
	root:=t.TempDir();pkg:=`{"scripts":{"dev":"vite","dev:wan":"vite --host 0.0.0.0","test:unit":"vitest"}}`;if err:=os.WriteFile(filepath.Join(root,"package.json"),[]byte(pkg),0o644);err!=nil{t.Fatal(err)};if err:=os.WriteFile(filepath.Join(root,"pnpm-lock.yaml"),[]byte("lockfileVersion: 9\n"),0o644);err!=nil{t.Fatal(err)}
	items,err:=(NPMProvider{}).Discover(root);if err!=nil{t.Fatal(err)};var wan Candidate;for _,c:=range items{if c.Script=="dev:wan"{wan=c}}
	if wan.Script!="dev:wan"||!reflect.DeepEqual(wan.Path,[]string{"npm","dev","wan"}){t.Fatalf("candidate=%#v",wan)}
	inv,err:=(NPMProvider{}).Build(root,wan);if err!=nil{t.Fatal(err)};if inv.Command!="pnpm"||!reflect.DeepEqual(inv.Args,[]string{"run","dev:wan"}){t.Fatalf("invocation=%#v",inv)}
}

func TestRunConfigRoundTripAndDefault(t *testing.T){
	root:=t.TempDir();c:=Candidate{Provider:"npm",Script:"dev:wan",Command:"npm"};if err:=SaveCandidate(root,"dev-wan",c,true);err!=nil{t.Fatal(err)};cfg,err:=LoadConfig(root);if err!=nil{t.Fatal(err)};if cfg.Default!="dev-wan"{t.Fatalf("default=%q",cfg.Default)};inv,err:=cfg.Invocation(root,cfg.Default,DefaultRegistry());if err!=nil{t.Fatal(err)};if inv.Name!="dev-wan"||inv.Command!="npm"||len(inv.Args)!=2||inv.Args[1]!="dev:wan"{t.Fatalf("invocation=%#v",inv)}
}

func TestNPMConfigRejectsTargetStack(t *testing.T){cfg:=RunConfig{Default:"bad",Launch:map[string]LaunchPoint{"bad":{Provider:"npm",Script:"dev",Targets:[]string{"wan"}}}};if err:=cfg.Validate();err==nil{t.Fatal("expected validation error")}}
