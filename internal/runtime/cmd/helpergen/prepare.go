package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Only this build-time preparation may fetch pinned modules. Go's checksum
// mechanism authenticates the downloads. Its writable module files are owned
// copies, and any attempted pin update is a refusal, never a checkout rewrite.
// Product build, install and execution do not invoke this command.
func prepareModules(root string) error {
	tmp, err := os.MkdirTemp("", "gh-tree-helper-prepare-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	pins := map[string][]byte{}
	for _, name := range []string{"go.mod", "go.sum"} {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return err
		}
		pins[name] = b
		if err := os.WriteFile(filepath.Join(tmp, name), b, 0600); err != nil {
			return err
		}
	}
	env := environment("amd64", "")
	for i, line := range env {
		key, _, _ := strings.Cut(line, "=")
		if key == "GOPROXY" || key == "GOSUMDB" {
			value, set := os.LookupEnv(key)
			if !set || value == "" {
				value = map[string]string{"GOPROXY": "https://proxy.golang.org,direct", "GOSUMDB": "sum.golang.org"}[key]
			}
			env[i] = key + "=" + value
		}
	}
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go.exe"), "mod", "download")
	cmd.Dir, cmd.Env = tmp, env
	out, commandErr := cmd.CombinedOutput()
	for name, before := range pins {
		after, err := os.ReadFile(filepath.Join(tmp, name))
		if err != nil || !bytes.Equal(before, after) {
			return fmt.Errorf("pinned dependency preparation attempted to change %s", name)
		}
	}
	if commandErr != nil {
		return fmt.Errorf("prepare pinned dependency cache: %w\n%s", commandErr, out)
	}
	return nil
}
