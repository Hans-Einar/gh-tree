package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func junction(t *testing.T, link, target string) {
	t.Helper()
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
		"$ErrorActionPreference = 'Stop'; $null = New-Item -ItemType Junction -Path "+quote(link)+" -Target "+quote(target))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create owned junction: %v: %s", err, out)
	}
	// Remove the junction itself, never recurse through a selected toolchain.
	t.Cleanup(func() {
		if err := os.Remove(link); err != nil {
			t.Errorf("remove owned junction %q: %v", link, err)
		}
	})
}

func TestSelectedRootJunctionAndRedirectedChild(t *testing.T) {
	owned := t.TempDir()
	link := filepath.Join(owned, "go-root-alias")
	junction(t, link, runtime.GOROOT())
	file := filepath.Join(link, "src", "internal", "goarch", "goarch.go")
	if got, err := contained(filepath.Join(link, "src"), file); err != nil || got != "internal/goarch/goarch.go" {
		t.Fatalf("selected toolchain root alias: relative=%q error=%v", got, err)
	}
	root := filepath.Join(owned, "selected")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "redirected")
	junction(t, child, filepath.Join(runtime.GOROOT(), "src"))
	if _, err := contained(root, filepath.Join(child, "internal", "goarch", "goarch.go")); err == nil || !strings.Contains(err.Error(), "redirected input outside root") {
		t.Fatalf("child outside selected root: %v", err)
	}
	missing := filepath.Join(root, "missing.go")
	if _, err := contained(root, missing); err == nil || !strings.Contains(err.Error(), missing) || !strings.Contains(err.Error(), "resolve selected input") {
		t.Fatalf("missing-input stage/path diagnostic: %v", err)
	}
}
