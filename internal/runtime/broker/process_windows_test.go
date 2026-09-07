package broker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 3 && os.Args[1] == "--owned-job-parent-fixture" {
		os.Exit(runOwnedParentDeathFixture(os.Args[2]))
	}
	if len(os.Args) > 1 && os.Args[1] == WindowsPrivateMode {
		os.Exit(RunWindowsPrivate())
	}
	code := m.Run()
	if err := cleanupBuiltFixtures(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

func TestWindowsOwnedUserFixture(t *testing.T) {
	for i, arg := range os.Args {
		if arg == "--owned-windows-fixture" {
			mode := os.Args[i+1]
			if mode == "leaf" {
				fmt.Println("OWNED_LEAF_READY")
				time.Sleep(15 * time.Second)
				os.Exit(0)
			}
			if mode == "branch" {
				exe, _ := os.Executable()
				child := exec.Command(exe, "-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "leaf")
				child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
				if err := child.Start(); err != nil {
					os.Exit(39)
				}
				fmt.Printf("OWNED_GRANDCHILD %d\n", child.Process.Pid)
				if err := os.WriteFile("grandchild-ready", []byte(fmt.Sprint(child.Process.Pid)), 0600); err != nil {
					os.Exit(40)
				}
				time.Sleep(15 * time.Second)
				os.Exit(0)
			}
			cwd, err := os.Getwd()
			if err != nil {
				os.Exit(30)
			}
			entries, err := os.ReadDir(".")
			if err != nil {
				os.Exit(31)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".gh-tree-start-") {
					os.Exit(32)
				}
			}
			debug, _, _ := kernel.NewProc("IsDebuggerPresent").Call()
			if debug != 0 {
				os.Exit(33)
			}
			if err = os.WriteFile("user-ran", []byte("after-detach"), 0600); err != nil {
				os.Exit(34)
			}
			if mode == "chdir" {
				if err = os.Chdir("child"); err != nil {
					os.Exit(35)
				}
			}
			b, _ := json.Marshal(struct {
				Cwd  string
				Args []string
				PID  int
			}{cwd, os.Args[i+2:], os.Getpid()})
			fmt.Printf("FIXTURE_READY %s\n", b)
			if mode == "flood" {
				block := bytes.Repeat([]byte{0, 0xff, 0x1b, 0x0d, 0x0a, 0x80, 0x41, 0x42}, 32<<10)
				for n := 0; n < 4; n++ {
					if _, err = os.Stdout.Write(block); err != nil {
						os.Exit(42)
					}
					if _, err = os.Stderr.Write(block); err != nil {
						os.Exit(43)
					}
				}
				os.Exit(0)
			}
			if mode == "hold" || mode == "root-first" {
				exe, _ := os.Executable()
				child := exec.Command(exe, "-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "branch")
				child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
				if err = child.Start(); err != nil {
					os.Exit(38)
				}
				deadline := time.Now().Add(5 * time.Second)
				for {
					if _, err = os.Stat("grandchild-ready"); err == nil {
						break
					}
					if time.Now().After(deadline) {
						os.Exit(41)
					}
					time.Sleep(time.Millisecond)
				}
				fmt.Printf("OWNED_CHILD %d\n", child.Process.Pid)
				if mode == "hold" {
					time.Sleep(15 * time.Second)
				}
				os.Exit(0)
			}
			if mode == "terminal" {
				line, err := bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					os.Exit(36)
				}
				var info windows.ConsoleScreenBufferInfo
				if err = windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &info); err != nil {
					os.Exit(37)
				}
				fmt.Printf("INPUT %s SIZE %d %d\n", strings.TrimSpace(line), info.Window.Right-info.Window.Left+1, info.Window.Bottom-info.Window.Top+1)
			}
			os.Exit(0)
		}
	}
}

func TestWindowsBrokerClientLifecycle(t *testing.T) {
	for _, mode := range []string{"pipe", "terminal", "hold", "root-first"} {
		t.Run(mode, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			s.Terminal = mode == "terminal"
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", mode}
			var mu sync.Mutex
			var output bytes.Buffer
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer cancel()
			config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(_ api.OutputStream, data []byte) { mu.Lock(); output.Write(data); mu.Unlock() }}
			if _, embedded, e := MachineRoute(); e != nil {
				t.Fatal(e)
			} else if embedded {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			client, started, err := StartWindows(ctx, config)
			if client != nil {
				defer func() {
					client.Stop()
					f, e := client.Wait(ctx)
					if !f.CleanupComplete {
						t.Errorf("incomplete client cleanup: %+v %v", f, e)
					}
				}()
			}
			if err != nil {
				t.Fatal(err)
			}
			if !started.Established || started.Cwd != s.RootLocator {
				t.Fatalf("bad started facts: %+v", started)
			}
			if mode == "terminal" {
				if got, e := client.Resize(ctx, 30, 100); e != nil || !got.Completed {
					t.Fatalf("resize: %+v %v", got, e)
				}
				if got, e := client.Write(ctx, []byte("broker input\r")); e != nil || got.Delivered != 13 {
					t.Fatalf("input: %+v %v", got, e)
				}
			}
			if mode == "hold" {
				for {
					mu.Lock()
					ready := strings.Contains(output.String(), "OWNED_CHILD")
					mu.Unlock()
					if ready {
						break
					}
					select {
					case <-ctx.Done():
						t.Fatal("no child ready")
					case <-time.After(10 * time.Millisecond):
					}
				}
				client.Stop()
			}
			if mode == "hold" || mode == "root-first" {
				mu.Lock()
				hasGrandchild := strings.Contains(output.String(), "OWNED_GRANDCHILD")
				mu.Unlock()
				if mode == "hold" && !hasGrandchild {
					t.Error("root/child/grandchild fixture did not establish")
				}
			}
			final, err := client.Wait(ctx)
			if err != nil || !final.CleanupComplete || !final.RootExited || !final.Quiescent || !final.Established {
				t.Fatalf("final: %+v %v", final, err)
			}
			mu.Lock()
			captured := output.String()
			mu.Unlock()
			if !strings.Contains(captured, "FIXTURE_READY") {
				t.Fatalf("missing output: %q", captured)
			}
			if (mode == "hold" || mode == "root-first") && !strings.Contains(captured, "OWNED_GRANDCHILD") {
				t.Fatalf("missing actual grandchild identity: %s", captured)
			}
			if mode == "terminal" && !strings.Contains(captured, "SIZE 100 30") {
				t.Fatalf("terminal output: %q", captured)
			}
			t.Logf("native broker %s: full inner/Release/outer barrier, %d output bytes", mode, len(captured))
		})
	}
}

func TestWindowsNativeStartupBarrier(t *testing.T) {
	for _, mode := range []string{"pipe", "chdir", "terminal"} {
		t.Run(mode, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			s := windowsSpec(t)
			s.Environment = os.Environ()
			s.Terminal = mode == "terminal"
			var err error
			s.Executable, err = os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			want := []string{"", "space value", `tail\`, `quote"inside`, "&()!世界"}
			s.Arguments = append([]string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", mode}, want...)
			if err = os.Mkdir(filepath.Join(s.RootLocator, "child"), 0700); err != nil {
				t.Fatal(err)
			}
			p := &userProcess{}
			var stages []string
			p.hook = func(stage string) {
				stages = append(stages, stage)
				if stage == "user-created-suspended" || stage == "before-user-resume" || stage == "cwd-breakpoint" || stage == "guards-released-pending-event" {
					if _, e := os.Stat(filepath.Join(s.RootLocator, "user-ran")); !os.IsNotExist(e) {
						t.Errorf("user code ran at %s: %v", stage, e)
					}
					if stage == "before-user-resume" {
						if n, e := activeProcesses(p.job); e != nil || n != 1 {
							t.Errorf("uncontained suspended user: %d %v", n, e)
						}
					}
				}
			}
			defer func() {
				if e := boundedCleanup(p); e != nil {
					t.Error(e)
				}
			}()
			if err = p.prepare(s); err != nil {
				t.Fatal(err)
			}
			type captured struct {
				stream byte
				data   []byte
				err    error
			}
			output := make(chan captured, len(p.outputs))
			for _, out := range p.outputs {
				go func(out nativeOutput) {
					b, e := io.ReadAll(out.file)
					ce := out.file.Close()
					if e == nil {
						e = ce
					}
					output <- captured{out.stream, b, e}
				}(out)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err = p.start(ctx, s); err != nil {
				t.Fatalf("start: %v; stages %v", err, stages)
			}
			if s.Terminal {
				if err = windows.ResizePseudoConsole(p.hpc, windows.Coord{X: 100, Y: 30}); err != nil {
					t.Fatal(err)
				}
				if _, err = p.input.Write([]byte("literal input\r")); err != nil {
					t.Fatal(err)
				}
			}
			exit, err := waitProcess(ctx, p.debug.process.Process)
			if err != nil {
				t.Fatal(err)
			}
			if exit != 0 {
				t.Fatalf("fixture exit %d", exit)
			}
			p.exit, p.rootWaited = exit, true
			if err = p.cleanup(ctx); err != nil {
				t.Fatal(err)
			}
			var all []byte
			for range p.outputs {
				select {
				case out := <-output:
					if out.err != nil {
						t.Error(out.err)
					}
					all = append(all, out.data...)
				case <-ctx.Done():
					t.Fatal("output owner did not join")
				}
			}
			if !bytes.Contains(all, []byte("FIXTURE_READY")) {
				t.Fatalf("missing fixture output: %s", all)
			}
			if s.Terminal && !bytes.Contains(all, []byte("SIZE 100 30")) {
				t.Fatalf("resize/input failed: %s", all)
			}
			if !s.Terminal {
				line := bytes.SplitN(all, []byte("\n"), 2)[0]
				var got struct {
					Cwd  string
					Args []string
				}
				if err = json.Unmarshal(bytes.TrimPrefix(line, []byte("FIXTURE_READY ")), &got); err != nil {
					t.Fatal(err)
				}
				if got.Cwd != s.RootLocator || fmt.Sprint(got.Args) != fmt.Sprint(want) {
					t.Fatalf("literal cwd/argv mismatch: %#v", got)
				}
			}
			t.Logf("native=%s terminal=%v barriers=%v output=%s", runtime.GOARCH, s.Terminal, stages, all)
		})
	}
}
