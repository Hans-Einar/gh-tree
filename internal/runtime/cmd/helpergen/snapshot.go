package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type buildSnapshot struct {
	root    string
	files   map[string]captured
	guards  []*os.File
	watches []*directoryWatch
}

func (s *buildSnapshot) close() error {
	var result error
	for i := len(s.watches) - 1; i >= 0; i-- {
		result = errors.Join(result, s.watches[i].close())
	}
	s.watches = nil
	for i := len(s.guards) - 1; i >= 0; i-- {
		result = errors.Join(result, s.guards[i].Close())
	}
	s.guards = nil
	return result
}

func snapshotName(key string, modules []module) (string, error) {
	for prefix, dest := range map[string]string{"repo/": "source/", "toolchain/": "goroot/", "modulemeta/": "modcache/cache/download/"} {
		if strings.HasPrefix(key, prefix) {
			return dest + strings.TrimPrefix(key, prefix), nil
		}
	}
	for _, m := range modules {
		prefix := "module/" + m.Path + "@" + m.Version + "/"
		if !m.Main && strings.HasPrefix(key, prefix) {
			return "modcache/" + escapeModule(m.Path) + "@" + escapeModule(m.Version) + "/" + strings.TrimPrefix(key, prefix), nil
		}
	}
	return "", fmt.Errorf("unattributed snapshot input %q", key)
}

func materialize(p plan, tmp string) (s buildSnapshot, resultErr error) {
	return materializeWithWatch(p, tmp, openDirectoryWatch)
}

func materializeWithWatch(p plan, tmp string, watchDirectory func(string) (*directoryWatch, error)) (s buildSnapshot, resultErr error) {
	s = buildSnapshot{root: tmp, files: map[string]captured{}}
	complete := false
	defer func() {
		if !complete {
			resultErr = errors.Join(resultErr, s.close())
		}
	}()
	if p.manifest.SourceDigest != hash(jsonBytes(p.manifest.Sources)) || p.manifest.OptionsDigest != hash(jsonBytes(p.manifest.Options)) || p.manifest.ModuleDigest != hash(jsonBytes(p.manifest.Modules)) {
		return s, fmt.Errorf("captured provenance mismatch")
	}
	if len(p.files) != len(p.manifest.Sources) {
		return s, fmt.Errorf("unrecorded captured input")
	}
	for _, record := range p.manifest.Sources {
		f, ok := p.files[record.Path]
		if !ok || f.source != record || hash(f.bytes) != record.SHA256 || len(f.bytes) != record.Length {
			return s, fmt.Errorf("captured input integrity mismatch: %s", record.Path)
		}
		if strings.HasPrefix(record.Path, "recipe/") {
			continue
		}
		name, err := snapshotName(record.Path, p.manifest.Modules)
		if err != nil {
			return s, err
		}
		if !filepath.IsLocal(filepath.FromSlash(name)) || strings.Contains(name, "\\") {
			return s, fmt.Errorf("invalid snapshot name %q", name)
		}
		path := filepath.Join(tmp, filepath.FromSlash(name))
		if _, exists := s.files[path]; exists {
			return s, fmt.Errorf("duplicate snapshot path %q", path)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return s, err
		}
		mode := os.FileMode(0400)
		if strings.HasSuffix(name, ".exe") {
			mode = 0500
		}
		if err := os.WriteFile(path, f.bytes, mode); err != nil {
			return s, err
		}
		s.files[path] = f
	}
	for _, name := range []string{"work", "cache", "gopath", "source", "goroot", "modcache"} {
		if err := os.MkdirAll(filepath.Join(tmp, name), 0700); err != nil {
			return s, err
		}
	}
	// Parent-first retained directory guards bind the names used by Go. Each
	// retained file is hashed through its guarded handle before build selection.
	dirs := map[string]bool{tmp: true}
	for _, name := range []string{"source", "goroot", "modcache"} {
		dirs[filepath.Join(tmp, name)] = true
	}
	for path := range s.files {
		for dir := filepath.Dir(path); dir != tmp; dir = filepath.Dir(dir) {
			dirs[dir] = true
		}
	}
	names := []string{}
	for dir := range dirs {
		names = append(names, dir)
	}
	sort.Strings(names)
	for _, dir := range names {
		guard, err := openImmutableInput(dir, true)
		if err != nil {
			return s, err
		}
		s.guards = append(s.guards, guard)
	}
	for path, f := range s.files {
		guard, err := openImmutableInput(path, false)
		if err != nil {
			return s, err
		}
		s.guards = append(s.guards, guard)
		b, err := io.ReadAll(guard)
		if err != nil {
			return s, err
		}
		if hash(b) != f.SHA256 || len(b) != f.Length {
			return s, fmt.Errorf("sealed input integrity mismatch: %s", path)
		}
	}
	// Acquire after opening the retained byte guards. Children precede parents
	// so our initial native directory acquisition is outside its parent's watch.
	// The entire set is validated only after ALL watches have been granted.
	for i := len(names) - 1; i >= 0; i-- {
		if names[i] == tmp {
			continue
		}
		watch, err := watchDirectory(names[i])
		if err != nil {
			return s, err
		}
		s.watches = append(s.watches, watch)
	}
	// Once every directory is continuously watched, validate the initial entire
	// input set, including non-Go/embed/include candidates. This closes the
	// materialization-to-watch interval; later changes latch kernel invalidation.
	for _, name := range []string{"source", "goroot", "modcache"} {
		err := filepath.WalkDir(filepath.Join(tmp, name), func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("redirected snapshot input %q", path)
			}
			if d.IsDir() {
				if !dirs[path] {
					return fmt.Errorf("unrecorded snapshot directory %q", path)
				}
				return nil
			}
			if _, ok := s.files[path]; !ok {
				return fmt.Errorf("unrecorded snapshot file %q", path)
			}
			return nil
		})
		if err != nil {
			return s, err
		}
	}
	if err := s.checkDirectories(); err != nil {
		return s, err
	}
	complete = true
	return s, nil
}

func (s buildSnapshot) command(arch string, args ...string) ([]byte, error) {
	return s.commandAfterStart(arch, nil, args...)
}

// The observer is a private test seam at the actual child-start boundary.
// Production always passes nil; no test substitutes a fake Go build result.
func (s buildSnapshot) commandAfterStart(arch string, started func(), args ...string) ([]byte, error) {
	if err := s.checkDirectories(); err != nil {
		return nil, err
	}
	overrides := map[string]string{"GOROOT": filepath.Join(s.root, "goroot"), "GOPATH": filepath.Join(s.root, "gopath"), "GOMODCACHE": filepath.Join(s.root, "modcache"), "GOTMPDIR": filepath.Join(s.root, "work"), "TEMP": filepath.Join(s.root, "work"), "TMP": filepath.Join(s.root, "work")}
	env := []string{}
	for _, line := range environment(arch, filepath.Join(s.root, "cache")) {
		key, _, _ := strings.Cut(line, "=")
		if _, ok := overrides[strings.ToUpper(key)]; !ok {
			env = append(env, line)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	cmd := exec.Command(filepath.Join(s.root, "goroot", "bin", "go.exe"), args...)
	cmd.Dir, cmd.Env = filepath.Join(s.root, "source"), env
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Start()
	if err == nil {
		if started != nil {
			started()
		}
		err = cmd.Wait()
	}
	guardErr := s.checkDirectories()
	if guardErr != nil {
		return nil, errors.Join(guardErr, err)
	}
	b := output.Bytes()
	if err != nil {
		return nil, fmt.Errorf("isolated go %s: %w\n%s", strings.Join(args, " "), err, b)
	}
	return b, nil
}

func (s buildSnapshot) checkDirectories() error {
	for _, watch := range s.watches {
		if err := watch.check(); err != nil {
			return err
		}
	}
	return nil
}

// Re-select using the copied Go executable and only the copied source roots.
// Every file the target selector chooses must be in the captured closure with
// those exact bytes. No host GOROOT/cache, network, auto-PGO or external linker
// participates in either build; missing inputs are errors, never fallbacks.
func (s buildSnapshot) verifySelection(p plan, arch string) error {
	b, err := s.command(arch, "list", "-mod=readonly", "-deps", "-json", entry)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	selected := map[string]bool{"repo/go.mod": true, "repo/go.sum": true}
	for {
		var q pkg
		err := d.Decode(&q)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(q.CgoFiles)+len(q.CFiles)+len(q.CXXFiles)+len(q.MFiles)+len(q.FFiles)+len(q.SwigFiles)+len(q.SwigCXXFiles) != 0 {
			return fmt.Errorf("unrecorded native compiler input %s", q.ImportPath)
		}
		for _, group := range [][]string{q.GoFiles, q.HFiles, q.SFiles, q.SysoFiles, q.EmbedFiles} {
			for _, name := range group {
				path := filepath.Join(q.Dir, name)
				f, ok := s.files[path]
				if !ok {
					return fmt.Errorf("unrecorded selected build input %q", path)
				}
				if _, err := contained(s.root, path); err != nil {
					return err
				}
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if !bytes.Equal(b, f.bytes) {
					return fmt.Errorf("isolated input changed: %s", path)
				}
				selected[f.Path] = true
			}
		}
	}
	// Include-only headers do not appear in go list. They were recursively
	// captured and validated against the same restricted include search roots.
	for key := range p.targetIncludes[arch] {
		selected[key] = true
	}
	if len(selected) != len(p.targetSources[arch]) {
		return fmt.Errorf("isolated %s source selection changed", arch)
	}
	for _, f := range p.targetSources[arch] {
		if !selected[f.Path] {
			return fmt.Errorf("isolated %s omitted captured input %s", arch, f.Path)
		}
	}
	return nil
}
