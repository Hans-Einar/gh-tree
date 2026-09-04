package launch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type NPMProvider struct{}

func (NPMProvider) Name() string { return "npm" }
func (NPMProvider) Detect(root string) bool {
	info, err := os.Stat(filepath.Join(root, "package.json"))
	return err == nil && !info.IsDir()
}

func (NPMProvider) Discover(root string) ([]Candidate, error) {
	root, err := cleanRoot(root)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil, err
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}
	manager := detectPackageManager(root)
	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]Candidate, 0, len(names))
	for _, name := range names {
		parts := strings.Split(name, ":")
		out = append(out, Candidate{Provider: "npm", ID: name, Path: append([]string{"npm"}, parts...), Script: name, Command: manager})
	}
	return out, nil
}

func (NPMProvider) Build(root string, c Candidate) (Invocation, error) {
	projectRoot, err := candidateRoot(root, c.Dir)
	if err != nil {
		return Invocation{}, err
	}
	script := strings.TrimSpace(c.Script)
	if script == "" {
		return Invocation{}, fmt.Errorf("npm script is empty")
	}
	manager := c.Command
	if manager == "" {
		manager = detectPackageManager(projectRoot)
	}
	return Invocation{Provider: "npm", Name: script, Command: manager, Args: []string{"run", script}, Dir: projectRoot}, nil
}

func detectPackageManager(root string) string {
	if _, err := os.Stat(filepath.Join(root, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(root, "yarn.lock")); err == nil {
		return "yarn"
	}
	return "npm"
}
