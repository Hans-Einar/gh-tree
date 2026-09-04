package launch

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMakeDiscoveryAndNativeOrderedStack(t *testing.T){root:=t.TempDir();content:=".PHONY: clean all install\nclean:\n\t@echo clean\nall: build\ninstall: all\n%.o: %.c\nVAR := value\n";if err:=os.WriteFile(filepath.Join(root,"Makefile"),[]byte(content),0o644);err!=nil{t.Fatal(err)};items,err:=(MakeProvider{}).Discover(root);if err!=nil{t.Fatal(err)};var names []string;for _,c:=range items{names=append(names,c.ID)};want:=[]string{"all","clean","install"};if !reflect.DeepEqual(names,want){t.Fatalf("targets=%v want %v",names,want)};candidate:=Candidate{Provider:"make",Targets:[]string{"clean","all","install"}};inv,err:=(MakeProvider{}).Build(root,candidate);if err!=nil{t.Fatal(err)};if inv.Command!="make"||!reflect.DeepEqual(inv.Args,[]string{"clean","all","install"}){t.Fatalf("invocation=%#v",inv)}}
func TestMakeConfigRoundTrip(t *testing.T){root:=t.TempDir();cfg:=RunConfig{Default:"release",Launch:map[string]LaunchPoint{"release":{Provider:"make",Targets:[]string{"clean","all","install"}}}};if err:=SaveConfig(root,cfg);err!=nil{t.Fatal(err)};loaded,err:=LoadConfig(root);if err!=nil{t.Fatal(err)};inv,err:=loaded.Invocation(root,"release",DefaultRegistry());if err!=nil{t.Fatal(err)};if !reflect.DeepEqual(inv.Args,[]string{"clean","all","install"}){t.Fatalf("args=%v",inv.Args)}}
