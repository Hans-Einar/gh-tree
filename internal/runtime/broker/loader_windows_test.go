package broker

import (
	"bytes"
	"context"
	"debug/pe"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func TestWindowsStaticDLLTLSAndDebugHeap(t *testing.T) {
	machine, emulated, err := MachineRoute()
	if err != nil {
		t.Fatal(err)
	}
	architecture := map[uint16]string{machine386: "x86", machineAMD64: "x64", machineARM64: "arm64"}[machine]
	s := windowsSpec(t)
	for _, name := range []string{"native_checks.h", "native_dll.c", "native_exe.c"} {
		data, err := os.ReadFile(filepath.Join("cmd", "fixtures", name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(s.RootLocator, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	vswhere := filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft Visual Studio", "Installer", "vswhere.exe")
	query := exec.Command(vswhere, "-latest", "-products", "*", "-property", "installationPath")
	result, err := query.Output()
	if err != nil {
		t.Fatalf("native DLL/TLS gate needs installed Visual Studio toolchain: %v", err)
	}
	installation := strings.TrimSpace(string(result))
	if installation == "" {
		t.Fatal("native DLL/TLS gate: no Visual Studio installation")
	}
	vcvars := filepath.Join(installation, "VC", "Auxiliary", "Build", "vcvarsall.bat")
	if strings.ContainsAny(vcvars, "\r\n\"%") {
		t.Fatal("unsupported compiler setup path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := `""` + vcvars + `" ` + architecture + ` && cl /W4 /WX /MTd /LD native_dll.c /link /OUT:native_loader.dll /IMPLIB:native_loader.lib && cl /W4 /WX /MTd native_exe.c native_loader.lib /link /OUT:native_loader.exe"`
	system, err := windows.GetSystemDirectory()
	if err != nil {
		t.Fatal(err)
	}
	shell := filepath.Join(system, "cmd.exe")
	build := exec.CommandContext(ctx, shell)
	build.SysProcAttr = &syscall.SysProcAttr{CmdLine: windows.EscapeArg(shell) + ` /D /S /C ` + command, HideWindow: true}
	build.Dir = s.RootLocator
	output, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("owned %s DLL/TLS fixture compile: %v\n%s", architecture, err, output)
	}
	t.Logf("installed native compiler/target evidence:\n%s", output)
	s.Executable = filepath.Join(s.RootLocator, "native_loader.exe")
	s.Environment = os.Environ()
	for _, name := range []string{"native_loader.exe", "native_loader.dll"} {
		image, err := pe.Open(filepath.Join(s.RootLocator, name))
		if err != nil {
			t.Fatal(err)
		}
		if image.Machine != machine {
			t.Fatalf("native loader fixture machine=%04x expected=%04x", image.Machine, machine)
		}
		image.Close()
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var captured bytes.Buffer
	config := WindowsConfig{SessionID: 21, Spec: s, Image: exe, Output: func(_ api.OutputStream, data []byte) { mu.Lock(); captured.Write(data); mu.Unlock() }}
	if emulated {
		config.Extraction = extractedNativeFixture(t)
		config.Image = config.Extraction.Path()
	}
	client, start, err := StartWindows(ctx, config)
	if client != nil {
		defer client.Stop()
	}
	if err != nil || !start.Established {
		t.Fatalf("DLL/TLS native startup %+v %v", start, err)
	}
	final, err := client.Wait(ctx)
	if err != nil || !final.CleanupComplete || final.ExitCode != 0 {
		mu.Lock()
		text := captured.String()
		mu.Unlock()
		t.Fatalf("DLL/TLS/debug heap %+v %v output=%s", final, err, text)
	}
	mu.Lock()
	text := captured.String()
	mu.Unlock()
	if !strings.Contains(text, "NATIVE_DLL_TLS_DEBUG_HEAP exe=0 dll=0") {
		t.Fatalf("missing DLL/TLS/debug heap proof: %s", text)
	}
	t.Logf("native %s mapped-machine=%04x DLL+exe TLS callbacks and static debug CRT heap: %s", architecture, start.Architecture.ImageMachine, text)
}
