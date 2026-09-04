package launch

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const configDir = ".gh-tree"
const configFile = "run.json"

type RunConfig struct { Default string `json:"default,omitempty"`; Launch map[string]LaunchPoint `json:"launch"` }
type LaunchPoint struct { Provider string `json:"provider"`; Script string `json:"script,omitempty"`; Targets []string `json:"targets,omitempty"`; Command string `json:"command,omitempty"` }

func ConfigPath(root string) string { return filepath.Join(root,configDir,configFile) }
func LoadConfig(root string)(RunConfig,error){root,err:=cleanRoot(root);if err!=nil{return RunConfig{},err};path:=ConfigPath(root);data,err:=os.ReadFile(path);if errors.Is(err,os.ErrNotExist){return RunConfig{Launch:map[string]LaunchPoint{}},nil};if err!=nil{return RunConfig{},fmt.Errorf("read %s: %w",path,err)};var cfg RunConfig;if err:=json.Unmarshal(data,&cfg);err!=nil{return RunConfig{},fmt.Errorf("parse %s: %w",path,err)};if cfg.Launch==nil{cfg.Launch=map[string]LaunchPoint{}};if err:=cfg.Validate();err!=nil{return RunConfig{},err};return cfg,nil}
func SaveConfig(root string,cfg RunConfig)error{root,err:=cleanRoot(root);if err!=nil{return err};if cfg.Launch==nil{cfg.Launch=map[string]LaunchPoint{}};if err:=cfg.Validate();err!=nil{return err};path:=ConfigPath(root);if err:=os.MkdirAll(filepath.Dir(path),0o755);err!=nil{return fmt.Errorf("create launch config directory: %w",err)};data,err:=json.MarshalIndent(cfg,"","  ");if err!=nil{return err};data=append(data,'\n');if err:=os.WriteFile(path,data,0o644);err!=nil{return fmt.Errorf("write launch config: %w",err)};return nil}
func (c RunConfig) Validate() error {if c.Default!=""{if _,ok:=c.Launch[c.Default];!ok{return fmt.Errorf("default launch point %q does not exist",c.Default)}};for name,p:=range c.Launch{if strings.TrimSpace(name)==""{return fmt.Errorf("launch point name is empty")};switch p.Provider{case "npm":if strings.TrimSpace(p.Script)==""{return fmt.Errorf("launch point %q: npm script is required",name)};if len(p.Targets)>0{return fmt.Errorf("launch point %q: npm does not accept target stacks",name)};case "make":if p.Script!=""{return fmt.Errorf("launch point %q: make does not accept npm scripts",name)};if len(p.Targets)==0{return fmt.Errorf("launch point %q: at least one make target is required",name)};for _,target:=range p.Targets{if !validMakeTarget(target){return fmt.Errorf("launch point %q: invalid make target %q",name,target)}};default:return fmt.Errorf("launch point %q: unsupported provider %q",name,p.Provider)}};return nil}
func (c RunConfig) Names() []string {names:=make([]string,0,len(c.Launch));for name:=range c.Launch{names=append(names,name)};sort.Strings(names);return names}
func (c RunConfig) Candidate(name string)(Candidate,error){p,ok:=c.Launch[name];if !ok{return Candidate{},fmt.Errorf("launch point %q not found",name)};return Candidate{Provider:p.Provider,ID:name,Path:[]string{p.Provider,name},Script:p.Script,Targets:append([]string(nil),p.Targets...),Command:p.Command},nil}
func (c RunConfig) Invocation(root,name string,registry Registry)(Invocation,error){candidate,err:=c.Candidate(name);if err!=nil{return Invocation{},err};inv,err:=registry.Build(root,candidate);if err!=nil{return Invocation{},err};inv.Name=name;return inv,nil}
func SaveCandidate(root,name string,candidate Candidate,makeDefault bool)error{cfg,err:=LoadConfig(root);if err!=nil{return err};cfg.Launch[name]=LaunchPoint{Provider:candidate.Provider,Script:candidate.Script,Targets:append([]string(nil),candidate.Targets...),Command:candidate.Command};if makeDefault||cfg.Default==""{cfg.Default=name};return SaveConfig(root,cfg)}
