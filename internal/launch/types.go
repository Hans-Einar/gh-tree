package launch

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

type Candidate struct {
	Provider string
	ID       string
	Path     []string
	Dir      string
	Script   string
	Targets  []string
	Command  string
}

type Invocation struct {
	Provider string
	Name     string
	Command  string
	Args     []string
	Dir      string
}

type Provider interface {
	Name() string
	Detect(root string) bool
	Discover(root string) ([]Candidate, error)
	Build(root string, candidate Candidate) (Invocation, error)
}

type Registry struct{ providers []Provider }

func DefaultRegistry() Registry { return Registry{providers: []Provider{NPMProvider{}, MakeProvider{}}} }
func NewRegistry(providers ...Provider) Registry {
	return Registry{providers: append([]Provider(nil), providers...)}
}

func (r Registry) Discover(root string) ([]Candidate, error) {
	root, err := cleanRoot(root)
	if err != nil {
		return nil, err
	}
	var all []Candidate
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			rel = ""
		}
		if rel != "" {
			parts := strings.Split(filepath.ToSlash(rel), "/")
			if len(parts) > 5 {
				return filepath.SkipDir
			}
			if ignoredLaunchDir(entry.Name()) {
				return filepath.SkipDir
			}
		}
		for _, provider := range r.providers {
			if !provider.Detect(path) {
				continue
			}
			items, err := provider.Discover(path)
			if err != nil {
				return fmt.Errorf("discover %s launch points in %s: %w", provider.Name(), displayLaunchDir(rel), err)
			}
			for i := range items {
				items[i].Dir = filepath.ToSlash(rel)
				if rel != "" {
					namespace := filepath.ToSlash(rel)
					items[i].ID = namespace + "/" + items[i].ID
					prefix := strings.Split(namespace, "/")
					items[i].Path = append(prefix, items[i].Path...)
				}
			}
			all = append(all, items...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Dir != all[j].Dir {
			return all[i].Dir < all[j].Dir
		}
		if all[i].Provider != all[j].Provider {
			return all[i].Provider < all[j].Provider
		}
		return strings.Join(all[i].Path, "/") < strings.Join(all[j].Path, "/")
	})
	return all, nil
}

func (r Registry) Build(root string, c Candidate) (Invocation, error) {
	for _, p := range r.providers {
		if p.Name() == c.Provider {
			return p.Build(root, c)
		}
	}
	return Invocation{}, fmt.Errorf("unknown launch provider %q", c.Provider)
}

func cleanRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("launch root is empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func candidateRoot(root, relative string) (string, error) {
	root, err := cleanRoot(root)
	if err != nil {
		return "", err
	}
	relative = strings.TrimSpace(relative)
	if relative == "" || relative == "." {
		return root, nil
	}
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("launch directory must be repository-relative")
	}
	candidate := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("launch directory escapes worktree: %q", relative)
	}
	return candidate, nil
}

func ignoredLaunchDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".gh-tree", "node_modules", "vendor", "dist", "build", "out", "target", ".next", ".cache":
		return true
	default:
		return false
	}
}

func displayLaunchDir(dir string) string {
	if dir == "" {
		return "."
	}
	return filepath.ToSlash(dir)
}
