package broker

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var fixtureBuilds struct {
	sync.Mutex
	paths     map[string]string
	directory string
}

// Builds only this package's newly owned test executable, never a user project.
// Cross-ABI fixture compilation is test infrastructure; product startup uses
// committed embedded bytes and never invokes a compiler.
func fixtureExecutable(t *testing.T, arch string) string {
	t.Helper()
	if arch == runtime.GOARCH {
		path, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	fixtureBuilds.Lock()
	defer fixtureBuilds.Unlock()
	if path := fixtureBuilds.paths[arch]; path != "" {
		return path
	}
	if fixtureBuilds.directory == "" {
		dir, err := os.MkdirTemp("", "gh-tree-owned-abi-fixtures-")
		if err != nil {
			t.Fatal(err)
		}
		fixtureBuilds.directory = dir
		fixtureBuilds.paths = make(map[string]string)
	}
	path := filepath.Join(fixtureBuilds.directory, "broker-test-"+arch+".exe")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go.exe"), "test", "-c", "-o", path, ".")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "GOARCH=") && !strings.HasPrefix(strings.ToUpper(entry), "CGO_ENABLED=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "GOARCH="+arch, "CGO_ENABLED=0")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("owned %s fixture build: %v\n%s", arch, err, output)
	}
	fixtureBuilds.paths[arch] = path
	return path
}

func nativeFixture(t *testing.T) (string, uint16) {
	t.Helper()
	machine, _, err := MachineRoute()
	if err != nil {
		t.Fatal(err)
	}
	arch := ""
	switch machine {
	case machine386:
		arch = "386"
	case machineAMD64:
		arch = "amd64"
	case machineARM64:
		arch = "arm64"
	default:
		t.Fatal("unsupported native fixture")
	}
	return fixtureExecutable(t, arch), machine
}

func extractedNativeFixture(t *testing.T) *WindowsImage {
	t.Helper()
	path, machine := nativeFixture(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	image, err := ExtractWindowsImage(data, machine, sha256.Sum256(data), ProtocolVersion)
	if image != nil {
		t.Cleanup(func() {
			if err := image.Cleanup(); err != nil {
				t.Error(err)
			}
		})
	}
	if err != nil {
		t.Fatal(err)
	}
	return image
}

func cleanupBuiltFixtures() error {
	var result error
	for _, path := range fixtureBuilds.paths {
		if err := os.Remove(path); err != nil {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}
	if fixtureBuilds.directory != "" {
		if err := os.Remove(fixtureBuilds.directory); err != nil {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}
	return result
}
