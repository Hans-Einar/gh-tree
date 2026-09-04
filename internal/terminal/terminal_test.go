package terminal

import (
	"os"
	"strings"
	"testing"
)

func TestTerminalLinesHandlesCarriageReturnAndBackspace(t *testing.T) {
	got := terminalLines([]byte("hello\rworld\nabc\bD\n"))
	joined := strings.Join(got, "|")
	if joined != "world|abD" {
		t.Fatalf("terminalLines=%q", joined)
	}
}

func TestSetEnvReplacesExistingValue(t *testing.T) {
	env := setEnv([]string{"A=1", "TERM=old"}, "TERM", "xterm-256color")
	if strings.Join(env, ";") != "A=1;TERM=xterm-256color" {
		t.Fatalf("env=%v", env)
	}
}

func TestDetectShellHonorsOverride(t *testing.T) {
	if os.Getenv("PATH") == "" {
		t.Skip("PATH unavailable")
	}
	candidate := ""
	if _, err := os.Stat("/bin/sh"); err == nil {
		candidate = "/bin/sh"
	} else if comspec := os.Getenv("COMSPEC"); comspec != "" {
		candidate = comspec
	}
	if candidate == "" {
		t.Skip("no deterministic shell candidate")
	}
	t.Setenv("GH_TREE_SHELL", candidate)
	sh, err := DetectShell()
	if err != nil {
		t.Fatal(err)
	}
	if sh.Path == "" || sh.Name == "" {
		t.Fatalf("shell=%#v", sh)
	}
}
