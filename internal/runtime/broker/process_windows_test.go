package broker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestWindowsOwnedUserFixture(t *testing.T) {
	for i, arg := range os.Args {
		if arg == "--owned-windows-fixture" {
			mode := os.Args[i+1]
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
			}{cwd, os.Args[i+2:]})
			fmt.Printf("FIXTURE_READY %s\n", b)
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
