package launch

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type ProcessState string
const(
	StateStarting ProcessState="starting"
	StateRunning ProcessState="running"
	StateExited ProcessState="exited"
	StateFailed ProcessState="failed"
	StateStopped ProcessState="stopped"
)

type ProcessSnapshot struct{ Invocation Invocation; State ProcessState; PID int; ExitCode int; Started time.Time; Ended time.Time; Lines []string }

type SessionManager struct{ mu sync.Mutex; current *session; maxLines int }
type session struct{ mu sync.Mutex; invocation Invocation; cmd *exec.Cmd; state ProcessState; exitCode int; started time.Time; ended time.Time; logs *lineBuffer; done chan struct{} }

func NewSessionManager(maxLines int)*SessionManager{if maxLines<=0{maxLines=500};return &SessionManager{maxLines:maxLines}}

func (m *SessionManager) Start(inv Invocation) error{
	if inv.Command==""{return fmt.Errorf("launch command is empty")};m.mu.Lock();defer m.mu.Unlock();if m.current!=nil{snap:=m.current.snapshot();if snap.State==StateRunning||snap.State==StateStarting{return fmt.Errorf("a launch process is already running")}}
	logs:=newLineBuffer(m.maxLines);cmd:=exec.Command(inv.Command,inv.Args...);cmd.Dir=inv.Dir;cmd.Stdout=logs;cmd.Stderr=logs;s:=&session{invocation:inv,cmd:cmd,state:StateStarting,exitCode:-1,started:time.Now(),logs:logs,done:make(chan struct{})};m.current=s
	if err:=cmd.Start();err!=nil{s.mu.Lock();s.state=StateFailed;s.ended=time.Now();s.mu.Unlock();close(s.done);return fmt.Errorf("start launch point: %w",err)}
	s.mu.Lock();s.state=StateRunning;s.mu.Unlock();go s.wait();return nil
}
func (s *session) wait(){err:=s.cmd.Wait();s.mu.Lock();defer s.mu.Unlock();s.ended=time.Now();if err==nil{s.state=StateExited;s.exitCode=0}else if ee,ok:=err.(*exec.ExitError);ok{s.exitCode=ee.ExitCode();if s.state!=StateStopped{s.state=StateFailed}}else{s.exitCode=-1;if s.state!=StateStopped{s.state=StateFailed}};close(s.done)}

func (m *SessionManager) Stop() error{m.mu.Lock();s:=m.current;m.mu.Unlock();if s==nil{return nil};s.mu.Lock();if s.state!=StateRunning&&s.state!=StateStarting{s.mu.Unlock();return nil};s.state=StateStopped;proc:=s.cmd.Process;s.mu.Unlock();if proc==nil{return nil};if runtime.GOOS!="windows"{_ = proc.Signal(os.Interrupt);select{case<-s.done:return nil;case<-time.After(1500*time.Millisecond):}};if err:=proc.Kill();err!=nil{return fmt.Errorf("stop launch process: %w",err)};select{case<-s.done:case<-time.After(2*time.Second):return fmt.Errorf("launch process did not stop")};return nil}
func (m *SessionManager) Restart() error{m.mu.Lock();s:=m.current;m.mu.Unlock();if s==nil{return fmt.Errorf("no launch process to restart")};inv:=s.snapshot().Invocation;if err:=m.Stop();err!=nil{return err};return m.Start(inv)}
func (m *SessionManager) Snapshot() ProcessSnapshot{m.mu.Lock();s:=m.current;m.mu.Unlock();if s==nil{return ProcessSnapshot{State:StateExited,ExitCode:-1}};return s.snapshot()}
func (m *SessionManager) Wait(timeout time.Duration)(ProcessSnapshot,error){m.mu.Lock();s:=m.current;m.mu.Unlock();if s==nil{return ProcessSnapshot{},fmt.Errorf("no launch process")};select{case<-s.done:return s.snapshot(),nil;case<-time.After(timeout):return s.snapshot(),fmt.Errorf("timed out waiting for launch process")}}
func (s *session) snapshot()ProcessSnapshot{s.mu.Lock();defer s.mu.Unlock();pid:=0;if s.cmd!=nil&&s.cmd.Process!=nil{pid=s.cmd.Process.Pid};return ProcessSnapshot{Invocation:s.invocation,State:s.state,PID:pid,ExitCode:s.exitCode,Started:s.started,Ended:s.ended,Lines:s.logs.lines()}}

type lineBuffer struct{mu sync.Mutex;max int;pending bytes.Buffer;data []string}
func newLineBuffer(max int)*lineBuffer{return &lineBuffer{max:max}}
func (b *lineBuffer) Write(p []byte)(int,error){b.mu.Lock();defer b.mu.Unlock();n:=len(p);_,_=b.pending.Write(p);scanner:=bufio.NewScanner(bytes.NewReader(b.pending.Bytes()));var complete []string;for scanner.Scan(){complete=append(complete,scanner.Text())};endsNewline:=b.pending.Len()>0&&(b.pending.Bytes()[b.pending.Len()-1]=='\n');b.pending.Reset();if len(complete)>0{last:=len(complete);if !endsNewline{last--;if last>=0{_,_=io.WriteString(&b.pending,complete[len(complete)-1])}};for i:=0;i<last;i++{b.data=append(b.data,complete[i])}};if len(b.data)>b.max{b.data=append([]string(nil),b.data[len(b.data)-b.max:]...)};return n,nil}
func (b *lineBuffer) lines()[]string{b.mu.Lock();defer b.mu.Unlock();out:=append([]string(nil),b.data...);if b.pending.Len()>0{out=append(out,b.pending.String())};return out}
