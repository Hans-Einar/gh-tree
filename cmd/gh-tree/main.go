package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Hans-Einar/gh-tree/internal/app"
	"github.com/Hans-Einar/gh-tree/internal/config"
	ghapi "github.com/Hans-Einar/gh-tree/internal/github"
	"github.com/Hans-Einar/gh-tree/internal/process"
	"github.com/Hans-Einar/gh-tree/internal/tui"
	"github.com/Hans-Einar/gh-tree/internal/worktree"
)

func main(){if err:=run(context.Background(),os.Args[1:]);err!=nil{fmt.Fprintln(os.Stderr,"gh-tree:",err);os.Exit(1)}}

func run(ctx context.Context,args []string)error{
	configDefault,err:=config.DefaultConfigPath();if err!=nil{return err};stateDefault,err:=config.DefaultStatePath();if err!=nil{return err}
	flags:=flag.NewFlagSet("gh-tree",flag.ContinueOnError);flags.SetOutput(os.Stderr);repoFlag:=flags.String("repo","","GitHub repository in owner/name form");configFlag:=flags.String("config",configDefault,"configuration file path");stateFlag:=flags.String("state",stateDefault,"navigation state file path")
	if err:=flags.Parse(args);err!=nil{if errors.Is(err,flag.ErrHelp){return nil};return err};if flags.NArg()!=0{return fmt.Errorf("unexpected arguments: %s",strings.Join(flags.Args()," "))}
	if _,err:=exec.LookPath("gh");err!=nil{return fmt.Errorf("GitHub CLI (gh) is required; install it and run 'gh auth login'")};if _,err:=exec.LookPath("git");err!=nil{return fmt.Errorf("git is required: %w",err)}
	loadedConfig,err:=config.Load(*configFlag);if err!=nil{return err};state,err:=config.OpenStateStore(*stateFlag);if err!=nil{return err}
	runner:=process.ExecRunner{};localRoot,rootErr:=worktree.FindRepositoryRoot(ctx,runner,"");clientDir:="";if rootErr==nil{clientDir=localRoot};client:=ghapi.NewClient(runner,clientDir);repo,err:=client.ResolveRepository(ctx,*repoFlag);if err!=nil{if *repoFlag==""&&rootErr!=nil{return errors.New("run inside a GitHub repository or pass --repo owner/name")};return err}
	var manager *worktree.Manager;if rootErr==nil{localClient:=ghapi.NewClient(runner,localRoot);localRepo,localErr:=localClient.ResolveRepository(ctx,"");if localErr==nil&&strings.EqualFold(localRepo,repo){manager=worktree.NewManager(runner,localRoot)}}
	service:=&app.Service{GitHub:client,Worktrees:manager}
	model:=tui.NewModel(repo,service,loadedConfig.StripPrefixes,loadedConfig.Targets(repo),*configFlag,state.LastFolder(repo),func(folder string)error{return state.SetLastFolder(repo,folder)})
	model=model.WithWorktreeState(state.LastWorktree(repo),func(path string)error{return state.SetLastWorktree(repo,path)})
	program:=tea.NewProgram(model,tea.WithAltScreen());if _,err:=program.Run();err!=nil{return fmt.Errorf("run terminal UI: %w",err)};return nil
}
