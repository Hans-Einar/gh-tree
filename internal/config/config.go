package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var DefaultStripPrefixes = []string{
	"steering",
	"codex",
	"worker",
	"review",
	"agent",
	"fix",
	"feature",
}

type Config struct {
	StripPrefixes []string              `json:"stripPrefixes"`
	Repos         map[string]RepoConfig `json:"repos"`
}

type RepoConfig struct {
	Worktrees map[string]WorktreeTarget `json:"worktrees"`
}

type WorktreeTarget struct {
	Name   string `json:"-"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

type State struct {
	LastFolders map[string]string `json:"lastFolders"`
}

func DefaultConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user config directory: %w", err)
	}
	return filepath.Join(base, "gh-tree", "config.json"), nil
}

func DefaultStatePath() (string, error) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		if base := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); base != "" {
			return filepath.Join(base, "gh-tree", "state.json"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find user home directory: %w", err)
		}
		return filepath.Join(home, ".local", "state", "gh-tree", "state.json"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user state directory: %w", err)
	}
	return filepath.Join(base, "gh-tree", "state.json"), nil
}

func Load(path string) (Config, error) {
	config := Config{}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return withDefaults(config), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return withDefaults(config), nil
}

func withDefaults(config Config) Config {
	if config.StripPrefixes == nil {
		config.StripPrefixes = append([]string(nil), DefaultStripPrefixes...)
	}
	if config.Repos == nil {
		config.Repos = make(map[string]RepoConfig)
	}
	return config
}

func (c Config) Targets(repo string) []WorktreeTarget {
	repoConfig, ok := lookupRepo(c.Repos, repo)
	if !ok {
		return nil
	}
	targets := make([]WorktreeTarget, 0, len(repoConfig.Worktrees))
	for name, target := range repoConfig.Worktrees {
		target.Name = name
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool {
		return strings.ToLower(targets[i].Name) < strings.ToLower(targets[j].Name)
	})
	return targets
}

func lookupRepo[T any](repos map[string]T, repo string) (T, bool) {
	if value, ok := repos[repo]; ok {
		return value, true
	}
	for key, value := range repos {
		if strings.EqualFold(key, repo) {
			return value, true
		}
	}
	var zero T
	return zero, false
}

// StateStore serializes per-repository navigation updates and persists them.
type StateStore struct {
	mu    sync.Mutex
	path  string
	state State
}

func OpenStateStore(path string) (*StateStore, error) {
	state := State{LastFolders: make(map[string]string)}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &state); err != nil {
			return nil, fmt.Errorf("parse state %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read state %s: %w", path, err)
	}
	if state.LastFolders == nil {
		state.LastFolders = make(map[string]string)
	}
	return &StateStore{path: path, state: state}, nil
}

func (s *StateStore) LastFolder(repo string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	folder, _ := lookupRepo(s.state.LastFolders, repo)
	return folder
}

func (s *StateStore) SetLastFolder(repo, folder string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.state.LastFolders {
		if strings.EqualFold(key, repo) && key != repo {
			delete(s.state.LastFolders, key)
		}
	}
	s.state.LastFolders[repo] = strings.Trim(folder, "/")
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write state %s: %w", s.path, err)
	}
	if err := os.Chmod(s.path, 0o600); err != nil {
		return fmt.Errorf("protect state %s: %w", s.path, err)
	}
	return nil
}
