package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMissingConfigUsesDefaults(t *testing.T) {
	t.Parallel()
	config, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(config.StripPrefixes, DefaultStripPrefixes) { t.Fatalf("prefixes = %v, want %v", config.StripPrefixes, DefaultStripPrefixes) }
}
func TestLoadHonorsExplicitEmptyPrefixOverride(t *testing.T) {
	t.Parallel(); path:=filepath.Join(t.TempDir(),"config.json")
	if err:=os.WriteFile(path,[]byte(`{"stripPrefixes":[],"repos":{}}`),0o600);err!=nil{t.Fatal(err)}
	config,err:=Load(path);if err!=nil{t.Fatal(err)}
	if config.StripPrefixes==nil||len(config.StripPrefixes)!=0{t.Fatalf("explicit empty prefixes were replaced: %#v",config.StripPrefixes)}
}
func TestTargetsAreCaseInsensitiveAndSorted(t *testing.T) {
	t.Parallel(); config:=Config{Repos:map[string]RepoConfig{"Hans-Einar/ponsse":{Worktrees:map[string]WorktreeTarget{"MVP1":{Path:"/tmp/mvp",Branch:"local/mvp"},"Concept1":{Path:"/tmp/concept",Branch:"local/concept"}}}}}
	targets:=config.Targets("hans-einar/PONSSE");if len(targets)!=2||targets[0].Name!="Concept1"||targets[1].Name!="MVP1"{t.Fatalf("targets = %#v",targets)}
}
func TestStateStorePersistsLastFolderAndWorktreePerRepository(t *testing.T) {
	t.Parallel(); path:=filepath.Join(t.TempDir(),"nested","state.json");store,err:=OpenStateStore(path);if err!=nil{t.Fatal(err)}
	if err:=store.SetLastFolder("Hans-Einar/ponsse","MVP1/machine-service");err!=nil{t.Fatal(err)}
	if err:=store.SetLastFolder("Hans-Einar/TerrainAnalyzer","Geometry");err!=nil{t.Fatal(err)}
	worktree:=filepath.Join(t.TempDir(),"ponsse-MVP1");if err:=store.SetLastWorktree("Hans-Einar/ponsse",worktree);err!=nil{t.Fatal(err)}
	reloaded,err:=OpenStateStore(path);if err!=nil{t.Fatal(err)}
	if got:=reloaded.LastFolder("hans-einar/PONSSE");got!="MVP1/machine-service"{t.Fatalf("ponsse folder = %q",got)}
	if got:=reloaded.LastFolder("Hans-Einar/TerrainAnalyzer");got!="Geometry"{t.Fatalf("TerrainAnalyzer folder = %q",got)}
	if got:=reloaded.LastWorktree("hans-einar/PONSSE");got!=filepath.Clean(worktree){t.Fatalf("worktree = %q want %q",got,filepath.Clean(worktree))}
}
