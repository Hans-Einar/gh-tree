//go:build linux || darwin || freebsd

package broker

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

type fixtureOutput struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *fixtureOutput) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.buffer.Len()+len(data) > 65536 {
		return 0, ErrProtocol
	}
	return b.buffer.Write(data)
}
func (b *fixtureOutput) text() string { b.mu.Lock(); defer b.mu.Unlock(); return b.buffer.String() }

type supervisorFixture struct {
	cmd                             *exec.Cmd
	control, reply, input, terminal *os.File
	readers                         []*os.File
	channel                         *Channel
	waitDone                        chan struct{}
	waitErr                         error
	output                          fixtureOutput
	outputWG                        sync.WaitGroup
	released                        bool
	acquired                        *AcquiredDirectory
}

func startSupervisorFixture(t *testing.T, spec StartSpec) *supervisorFixture {
	t.Helper()
	a, err := AcquireCwd(spec)
	if err != nil {
		closeAcquired(t, a)
		t.Fatal(err)
	}
	f := &supervisorFixture{acquired: a, waitDone: make(chan struct{})}
	controlRead, controlWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	replyRead, replyWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	f.control = controlWrite
	f.reply = replyRead
	var stdin, stdout, stderr *os.File
	if spec.Terminal {
		master, slave, err := pty.Open()
		if err != nil {
			t.Fatal(err)
		}
		if err := pty.Setsize(master, &pty.Winsize{Rows: spec.Rows, Cols: spec.Columns}); err != nil {
			t.Fatal(err)
		}
		f.terminal = master
		f.input = master
		f.readers = []*os.File{master}
		stdin, stdout, stderr = slave, slave, slave
	} else {
		inputRead, inputWrite, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		outRead, outWrite, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		errRead, errWrite, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		f.input = inputWrite
		f.readers = []*os.File{outRead, errRead}
		stdin, stdout, stderr = inputRead, outWrite, errWrite
	}
	cmd := exec.Command(must(os.Executable()), supervisorPrivateMarker)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.ExtraFiles = []*os.File{controlRead, replyWrite, a.File()}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if spec.Terminal {
		cmd.SysProcAttr.Setctty = true
		cmd.SysProcAttr.Ctty = 0
	}
	f.cmd = cmd
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() { f.waitErr = cmd.Wait(); close(f.waitDone) }()
	for _, file := range []*os.File{controlRead, replyWrite, stdin} {
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if !spec.Terminal {
		if err := errors.Join(stdout.Close(), stderr.Close()); err != nil {
			t.Fatal(err)
		}
	}
	for _, reader := range f.readers {
		f.outputWG.Add(1)
		go func(r *os.File) { defer f.outputWG.Done(); _, _ = io.Copy(&f.output, r) }(reader)
	}
	nonce := must(FreshNonce())
	f.channel = must(NewChannel(replyRead, controlWrite, Parent, UnixSupervisor, 1, nonce))
	t.Cleanup(func() { f.close(t) })
	spec.ParentID = uint64(os.Getpid())
	if err := controlWrite.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := f.channel.Send(Start, must(encodeUnixStart(spec, 2*time.Second, 3*time.Second))); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *supervisorFixture) receive(t *testing.T) Frame {
	t.Helper()
	if err := f.reply.SetReadDeadline(time.Now().Add(8 * time.Second)); err != nil {
		t.Fatal(err)
	}
	frame, err := f.channel.Receive()
	if err != nil {
		t.Fatalf("private frame: %v; output=%q", err, f.output.text())
	}
	return frame
}

func (f *supervisorFixture) release(t *testing.T) {
	t.Helper()
	if f.released {
		return
	}
	if err := f.control.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := f.channel.Send(Release, nil); err != nil {
		t.Fatal(err)
	}
	f.released = true
	select {
	case <-f.waitDone:
		if f.waitErr != nil {
			t.Fatal(f.waitErr)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("supervisor exact waiter did not join")
	}
	f.outputWG.Wait()
}

func (f *supervisorFixture) close(t *testing.T) {
	t.Helper()
	if !f.released {
		f.control.SetWriteDeadline(time.Now().Add(time.Second))
		sendErr := f.channel.Send(Abort, nil)
		if sendErr == nil {
			f.reply.SetReadDeadline(time.Now().Add(8 * time.Second))
			for {
				frame, err := f.channel.Receive()
				if err != nil {
					t.Errorf("fixture teardown: %v", err)
					break
				}
				if frame.Opcode == Quiescent {
					f.release(t)
					break
				}
			}
		} else {
			t.Errorf("fixture teardown Abort: %v", sendErr)
		}
	}
	closeErr := errors.Join(f.control.Close(), f.reply.Close(), f.acquired.Close())
	if f.terminal == nil {
		closeErr = errors.Join(closeErr, f.input.Close())
	}
	for _, reader := range f.readers {
		closeErr = errors.Join(closeErr, reader.Close())
	}
	if closeErr != nil {
		t.Error(closeErr)
	}
}

func TestNativeSupervisorAcquiredCwdNaturalExitAndRelease(t *testing.T) {
	spec := nativeSpec(t)
	spec.Executable = must(os.Executable())
	spec.Arguments = []string{"--runtime-fixture-cwd"}
	f := startSupervisorFixture(t, spec)
	started := f.receive(t)
	if started.Opcode != Started {
		t.Fatalf("startup opcode %v", started.Opcode)
	}
	rootExit := f.receive(t)
	if rootExit.Opcode != UserExit || len(rootExit.Payload) != 8 {
		t.Fatal("root exit missing")
	}
	quiet := f.receive(t)
	if quiet.Opcode != Quiescent {
		t.Fatalf("quiescence opcode %v", quiet.Opcode)
	}
	select {
	case <-f.waitDone:
		t.Fatal("supervisor exited before Release")
	default:
	}
	f.release(t)
	if output := f.output.text(); !strings.Contains(output, "marker=selected-original") || !strings.Contains(output, "owned-cwd="+filepath.Join(spec.RootLocator, spec.Components[0])) {
		t.Fatalf("actual cwd/output: %q", output)
	}
}

func TestNativeSupervisorRootExitsBeforeGrandchildCleanup(t *testing.T) {
	spec := nativeSpec(t)
	spec.Executable = must(os.Executable())
	spec.Arguments = []string{"--runtime-fixture-tree-root"}
	f := startSupervisorFixture(t, spec)
	if frame := f.receive(t); frame.Opcode != Started {
		t.Fatal("start", frame.Opcode)
	}
	if frame := f.receive(t); frame.Opcode != UserExit {
		t.Fatal("root exit", frame.Opcode)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	members, _, err := sessionMembers(must(census(ctx)), f.cmd.Process.Pid)
	if err != nil || len(members) < 2 {
		t.Fatal("descendant was not observed after root exit", err)
	}
	if frame := f.receive(t); frame.Opcode != Quiescent {
		t.Fatal("quiescence", frame.Opcode)
	}
	f.release(t)
	for _, p := range must(census(context.Background())) {
		if p.live && p.session == f.cmd.Process.Pid {
			t.Fatalf("residual member %+v", p)
		}
	}
	if !strings.Contains(f.output.text(), "owned-tree-child=") || !strings.Contains(f.output.text(), "owned-tree-grandchild=") {
		t.Fatal("missing actual child identity evidence")
	}
}

func TestNativeSupervisorPTYJobControlResizeAndRelease(t *testing.T) {
	shell := "/bin/sh"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "/bin/bash"
	}
	spec := nativeSpec(t)
	spec.Terminal = true
	spec.Executable = shell
	if filepath.Base(shell) == "bash" {
		spec.Arguments = []string{"--noprofile", "--norc", "-i"}
	} else {
		spec.Arguments = []string{"-i"}
	}
	spec.Environment = []string{"PATH=/bin:/usr/bin", "HISTFILE=/dev/null", "ENV=/dev/null", "HOME=" + spec.RootLocator, "TERM=xterm"}
	f := startSupervisorFixture(t, spec)
	if frame := f.receive(t); frame.Opcode != Started {
		t.Fatal("terminal start", frame.Opcode)
	}
	if err := pty.Setsize(f.terminal, &pty.Winsize{Rows: 30, Cols: 100}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.input.Write([]byte("stty size; printf '__owned_pty_ready__\\n'\n")); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for !strings.Contains(f.output.text(), "30 100\r\n") || !strings.Contains(f.output.text(), "__owned_pty_ready__\r\n") {
		if time.Now().After(deadline) {
			t.Fatal("native terminal output", f.output.text())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := f.input.Write([]byte("sleep 20 | cat &\nsleep 20\n")); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(3 * time.Second)
	for {
		_, groups, err := sessionMembers(must(census(context.Background())), f.cmd.Process.Pid)
		foreground := must(unix.IoctlGetInt(int(f.terminal.Fd()), unix.TIOCGPGRP))
		if err == nil && len(groups) >= 3 && foreground != f.cmd.Process.Pid {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("foreground/background PTY groups not observed: %v %v", groups, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := f.control.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := f.channel.Send(Stop, nil); err != nil {
		t.Fatal(err)
	}
	for {
		frame := f.receive(t)
		if frame.Opcode == Quiescent {
			break
		}
		if frame.Opcode != UserExit {
			t.Fatal("terminal cleanup", frame.Opcode)
		}
	}
	f.release(t)
}
