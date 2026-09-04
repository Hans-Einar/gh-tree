package launch

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T){if os.Getenv("GO_WANT_GH_TREE_HELPER")!="1"{return};mode:="";for i,a:=range os.Args{if a=="--"&&i+1<len(os.Args){mode=os.Args[i+1];break}};switch mode{case "log":fmt.Println("one");fmt.Println("two");fmt.Println("three");os.Exit(0);case "fail":fmt.Fprintln(os.Stderr,"boom");os.Exit(7);case "sleep":fmt.Println("ready");time.Sleep(30*time.Second);os.Exit(0)};os.Exit(2)}

func helperInvocation(mode string)Invocation{return Invocation{Provider:"test",Name:mode,Command:os.Args[0],Args:[]string{"-test.run=TestHelperProcess","--",mode},Dir:"."}}

func TestSessionCapturesBoundedLogsAndExit(t *testing.T){t.Setenv("GO_WANT_GH_TREE_HELPER","1");m:=NewSessionManager(2);if err:=m.Start(helperInvocation("log"));err!=nil{t.Fatal(err)};snap,err:=m.Wait(5*time.Second);if err!=nil{t.Fatal(err)};if snap.State!=StateExited||snap.ExitCode!=0{t.Fatalf("snapshot=%#v",snap)};joined:=strings.Join(snap.Lines,"\n");if strings.Contains(joined,"one")||!strings.Contains(joined,"two")||!strings.Contains(joined,"three"){t.Fatalf("logs=%q",joined)}}
func TestSessionReportsFailureExitCode(t *testing.T){t.Setenv("GO_WANT_GH_TREE_HELPER","1");m:=NewSessionManager(20);if err:=m.Start(helperInvocation("fail"));err!=nil{t.Fatal(err)};snap,err:=m.Wait(5*time.Second);if err!=nil{t.Fatal(err)};if snap.State!=StateFailed||snap.ExitCode!=7||!strings.Contains(strings.Join(snap.Lines,"\n"),"boom"){t.Fatalf("snapshot=%#v",snap)}}
func TestSessionStopsAttachedProcess(t *testing.T){t.Setenv("GO_WANT_GH_TREE_HELPER","1");m:=NewSessionManager(20);if err:=m.Start(helperInvocation("sleep"));err!=nil{t.Fatal(err)};time.Sleep(100*time.Millisecond);if err:=m.Stop();err!=nil{t.Fatal(err)};snap:=m.Snapshot();if snap.State!=StateStopped{t.Fatalf("state=%s",snap.State)}}
