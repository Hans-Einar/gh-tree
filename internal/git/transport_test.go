package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if mode := os.Getenv("GH_TREE_GIT_TEST_HELPER"); mode != "" {
		switch mode {
		case "literal":
			for _, arg := range os.Args[1:] {
				fmt.Fprint(os.Stdout, arg, "\x00")
			}
			fmt.Fprint(os.Stderr, "private diagnostic bytes")
		case "flood":
			for i := 0; i < 1024; i++ {
				_, _ = os.Stdout.Write(bytes.Repeat([]byte("x"), 4096))
			}
		case "wait":
			time.Sleep(5 * time.Second)
		case "descendant":
			cmd := exec.Command(os.Args[0])
			cmd.Env = append(commandEnvironment(os.Environ()), "GH_TREE_GIT_TEST_HELPER=hold")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			hideWindow(cmd)
			if err := cmd.Start(); err != nil {
				os.Exit(2)
			}
		case "hold":
			time.Sleep(500 * time.Millisecond)
			_ = os.WriteFile(os.Getenv("GH_TREE_GIT_TEST_MARKER"), []byte("done"), 0600)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func testAdapter(t *testing.T, extra ...string) *Adapter {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	a, err := New(Options{GitExecutable: exe, CurrentDirectory: t.TempDir(), Environment: append(os.Environ(), extra...), ReadTimeout: time.Second, DrainTimeout: 50 * time.Millisecond, MaxStdoutBytes: 1024, MaxStderrBytes: 256})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestTransportLiteralBoundsAndCopy(t *testing.T) {
	a := testAdapter(t, "GH_TREE_GIT_TEST_HELPER=literal")
	args := []string{" space ", "-dash", "é:file", "line\nbreak", "$(not-a-shell)", ""}
	r := a.command(context.Background(), a.current.path, false, args...)
	if r.err != nil {
		t.Fatal(r.err)
	}
	want := []byte(strings.Join(args, "\x00") + "\x00")
	if !bytes.HasSuffix(r.stdout, want) {
		t.Fatalf("argv changed: %q", r.stdout)
	}
	d := r.transport.Data()
	if !d.Started || !d.RootReaped || !d.CleanupKnown || d.StdoutTruncated {
		t.Fatalf("bad transport: %+v", d)
	}
	if string(r.stderr) != "private diagnostic bytes" {
		t.Fatalf("stderr merged or lost: %q", r.stderr)
	}
	r2 := a.command(context.Background(), a.current.path, false, "next")
	r2.stdout[0] = 'x'
	if !bytes.HasSuffix(r.stdout, want) {
		t.Fatal("earlier output mutated")
	}
}

func TestTransportLimitCancelsAndReaps(t *testing.T) {
	a := testAdapter(t, "GH_TREE_GIT_TEST_HELPER=flood")
	start := time.Now()
	r := a.command(context.Background(), a.current.path, false)
	if r.err == nil || len(r.stdout) != 1024 || !r.transport.Data().StdoutTruncated || !r.transport.Data().RootReaped {
		t.Fatalf("bad limit result: %v %+v", r.err, r.transport.Data())
	}
	if r.transport.Data().CleanupKnown {
		t.Fatal("killed root fabricated descendant proof")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("output limit did not bound execution")
	}
}

func TestTransportCancellationBeforeAndAfterStart(t *testing.T) {
	a := testAdapter(t, "GH_TREE_GIT_TEST_HELPER=wait")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := a.command(ctx, a.current.path, false)
	if r.err == nil || r.transport.Data().Started || !r.transport.Data().CleanupKnown {
		t.Fatalf("bad prestart result: %+v", r.transport.Data())
	}
	ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	r = a.command(ctx, a.current.path, false)
	d := r.transport.Data()
	if r.err == nil || !d.Started || !d.RootReaped || d.CleanupKnown || !d.CancellationRequested {
		t.Fatalf("bad canceled result: %v %+v", r.err, d)
	}
}

func TestTransportDescendantHeldPipeIsExplicitResidual(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "finished")
	a := testAdapter(t, "GH_TREE_GIT_TEST_HELPER=descendant", "GH_TREE_GIT_TEST_MARKER="+marker)
	r := a.command(context.Background(), a.current.path, false)
	if r.err == nil || !r.transport.Data().RootReaped || r.transport.Data().CleanupKnown {
		t.Fatalf("held pipe falsely clean: %v %+v", r.err, r.transport.Data())
	}
	// The owned helper deliberately exits itself; wait for its explicit final
	// marker so the test leaves no long-running descendant behind.
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("test helper did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCommandEnvironmentScopeOverridesAreRemoved(t *testing.T) {
	env := commandEnvironment([]string{"GIT_DIR=foreign", "GIT_INDEX_FILE=foreign", "GIT_NO_REPLACE_OBJECTS=0", "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=user.name", "GIT_CONFIG_VALUE_0=Fixture", "SSH_AUTH_SOCK=fixture"})
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "foreign") || !strings.Contains(joined, "GIT_CONFIG_VALUE_0=Fixture") || !strings.Contains(joined, "SSH_AUTH_SOCK=fixture") || !strings.Contains(joined, "GIT_NO_REPLACE_OBJECTS=1") {
		t.Fatalf("scope/config mapping: %s", joined)
	}
}

func TestDirectoryObservationBindsPhysicalObject(t *testing.T) {
	root := t.TempDir()
	original := filepath.Join(root, "original")
	if err := os.Mkdir(original, 0700); err != nil {
		t.Fatal(err)
	}
	first, err := observeDirectory(original)
	if err != nil {
		t.Fatal(err)
	}
	again, err := observeDirectory(original)
	if err != nil || again != first {
		t.Fatal("same directory changed identity", err)
	}
	child := filepath.Join(original, "child")
	if err := os.WriteFile(child, []byte("child"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(child, filepath.Join(original, "renamed")); err != nil {
		t.Fatal(err)
	}
	contentsChanged, err := observeDirectory(original)
	if err != nil {
		t.Fatal(err)
	}
	if !sameDirectoryObject(first, contentsChanged) {
		t.Fatal("child edit changed physical object")
	}
	if strings.HasPrefix(first.identity.Stamp(), "birth") && first.identity != contentsChanged.identity {
		t.Fatal("child edit changed stable birth identity")
	}
	if err := os.Rename(original, filepath.Join(root, "retained")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(original, 0700); err != nil {
		t.Fatal(err)
	}
	next, err := observeDirectory(original)
	if err != nil {
		t.Fatal(err)
	}
	if next.identity == first.identity {
		t.Fatal("replacement collapsed to original identity")
	}
	if _, err := observeDirectory(filepath.Join(root, "missing")); err == nil {
		t.Fatal("missing directory fabricated")
	}
}
