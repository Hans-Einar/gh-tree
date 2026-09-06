package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type allowance struct{ Path, BaselineBlob string }
type migrationMap struct {
	Baseline                 string
	TemporaryLegacyAllowance []allowance
	Files                    []struct{ Path, Kind, Action, BaselineBlob string }
}
type checker struct {
	root, module, mode string
	gitClean           bool
	autoCRLF           string
	legacy             map[string]string
	entryBlob          string
	exempt             map[string]bool
}

func newChecker(root, mode string) (*checker, error) {
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}
	var module string
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && f[0] == "module" {
			module = strings.Trim(f[1], "\"")
		}
	}
	if module == "" {
		return nil, fmt.Errorf("go.mod has no module declaration")
	}
	b, err = os.ReadFile(filepath.Join(root, "SDP/Design/CR-#21/MigrationMap.yaml"))
	if err != nil {
		return nil, err
	}
	var m migrationMap
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("MigrationMap (JSON-compatible YAML): %w", err)
	}
	if m.Baseline != "f626077ca0e59fbe9ede7ba1116982bb94b2eb6b" {
		return nil, fmt.Errorf("unexpected MigrationMap baseline %q", m.Baseline)
	}
	c := &checker{root: root, module: module, mode: mode, legacy: map[string]string{}, exempt: map[string]bool{}}
	for _, a := range m.TemporaryLegacyAllowance {
		if !strings.HasSuffix(a.Path, ".go") || filepath.ToSlash(filepath.Clean(a.Path)) != a.Path || len(a.BaselineBlob) != 40 || c.legacy[a.Path] != "" {
			return nil, fmt.Errorf("invalid/duplicate exact legacy allowance: %q", a.Path)
		}
		c.legacy[a.Path] = a.BaselineBlob
	}
	for _, f := range m.Files {
		if f.Path == "cmd/gh-tree/main.go" && f.Action == "retain_or_rewrite_serially" {
			c.entryBlob = f.BaselineBlob
		}
	}
	if len(c.legacy) == 0 || len(c.entryBlob) != 40 {
		return nil, fmt.Errorf("MigrationMap lacks legacy inventory or shared entry baseline")
	}
	return c, nil
}

// Hash actual bytes. Path-specific text conversion, when proved, is separate.
func blobHash(b []byte) string {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d%c", len(b), 0)
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *checker) inventory() error {
	// go list reports physical package directories. In particular, macOS's
	// /var temporary-root alias resolves to /private/var. Bind the selected root
	// once to that same physical spelling before any relative ownership checks.
	root, err := filepath.EvalSymlinks(c.root)
	if err != nil {
		return err
	}
	c.root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	if err := c.prepareCleanPolicy(); err != nil {
		return err
	}
	c.exempt = map[string]bool{}
	packages := map[string]bool{}
	legacyHashes := map[string]string{}
	for path, hash := range c.legacy {
		legacyHashes[hash] = path
	}
	returnAfter := filepath.WalkDir(c.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != c.root && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "testdata" || entry.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, err := filepath.Rel(c.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source symlink is not an owned physical file: %s", rel)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash := blobHash(b)
		// Check clean metadata wherever a baseline blob could grant an allowance
		// or identify a renamed baseline copy. Ordinary new files need no clean
		// conversion and cannot acquire an exception from their directory.
		if c.legacy[rel] != "" || rel == "cmd/gh-tree/main.go" || legacyHashes[hash] != "" || legacyHashes[blobHash(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))] != "" {
			hash, err = c.cleanBlobHash(rel, b)
			if err != nil {
				return err
			}
		}
		if want, old := c.legacy[rel]; old {
			if c.mode == "strict" {
				return fmt.Errorf("strict mode requires empty legacy allowance: %s remains", rel)
			}
			if hash != want {
				return fmt.Errorf("legacy allowance changed: %s blob %s, expected %s", rel, hash, want)
			}
			c.exempt[rel] = true
			return nil
		}
		// The shared CLI entry is separately classified, not added to the legacy
		// folder list. M6 may replace it only with strict Composition wiring.
		if c.mode == "staged" && rel == "cmd/gh-tree/main.go" && hash == c.entryBlob {
			c.exempt[rel] = true
			return nil
		}
		if old := legacyHashes[hash]; old != "" {
			return fmt.Errorf("renamed/copied legacy source: %s has baseline blob of %s", rel, old)
		}
		pkg := filepath.ToSlash(filepath.Dir(rel))
		if classify(pkg) == "" {
			return fmt.Errorf("package inventory forbids %s (no accepted owner)", pkg)
		}
		if !strings.HasSuffix(rel, "_test.go") {
			packages[pkg] = true
			selected := false
			for _, t := range targets {
				ctx := build.Default
				ctx.GOOS = t.GOOS
				ctx.GOARCH = t.GOARCH
				ctx.CgoEnabled = false
				ctx.BuildTags = nil
				ok, err := ctx.MatchFile(filepath.Dir(path), filepath.Base(path))
				if err != nil {
					return fmt.Errorf("%s: %w", rel, err)
				}
				selected = selected || ok
			}
			if !selected {
				return fmt.Errorf("production source is not selected by any accepted target/default tags: %s", rel)
			}
		}
		return nil
	})
	if returnAfter != nil {
		return returnAfter
	}
	if c.mode == "strict" {
		for _, pkg := range []string{"cmd/gh-tree", "internal/version", "internal/composition", "internal/composition/host", "internal/domain", "internal/application", "internal/application/api", "internal/application/ports", "internal/application/usecases", "internal/git", "internal/github/adapter", "internal/runtime", "internal/runtime/broker", "internal/runtime/broker/cmd", "internal/runtime/brokerassets", "internal/runtime/cmd/helpergen", "internal/launchdiscovery", "internal/persistence", "internal/tuistate", "internal/tuistate/viewmodel", "internal/tuiview"} {
			if !packages[pkg] {
				return fmt.Errorf("strict final package inventory missing %s", pkg)
			}
		}
	}
	return nil
}

type listedPackage struct {
	Dir, ImportPath, Name, Export                string
	Standard                                     bool
	GoFiles, CgoFiles, TestGoFiles, XTestGoFiles []string
	Error                                        *struct{ Err string }
	DepsErrors                                   []struct{ Err string }
}

func targetEnv(t targetSpec) []string {
	override := map[string]string{"GOOS": t.GOOS, "GOARCH": t.GOARCH, "CGO_ENABLED": "0", "GOFLAGS": "-mod=readonly", "GOWORK": "off", "GOAMD64": "v1", "GO386": "sse2", "GOARM": "7", "GOARM64": "v8.0"}
	env := []string{}
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if _, ok := override[strings.ToUpper(key)]; !ok {
			env = append(env, item)
		}
	}
	for key, value := range override {
		env = append(env, key+"="+value)
	}
	return env
}

func (c *checker) checkTarget(t targetSpec) error {
	cmd := exec.Command("go", "list", "-deps", "-export", "-json", "./...")
	cmd.Dir = c.root
	cmd.Env = targetEnv(t)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("go list %s/%s: %w\n%s", t.GOOS, t.GOARCH, err, stderr.String())
	}
	pkgs := map[string]listedPackage{}
	dec := json.NewDecoder(bytes.NewReader(b))
	for {
		var p listedPackage
		err := dec.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if p.Error != nil || len(p.DepsErrors) != 0 {
			return fmt.Errorf("incomplete Go package metadata: %s", p.ImportPath)
		}
		pkgs[p.ImportPath] = p
	}
	// Resolve test-only standard imports from the actual selected Go command.
	// A trimpath-built checker has no usable embedded build.Default.GOROOT.
	std := exec.Command("go", "list", "-f", "{{.ImportPath}}", "std")
	std.Dir = c.root
	std.Env = targetEnv(t)
	standard, err := std.Output()
	if err != nil {
		return fmt.Errorf("selected Go standard-library metadata: %w", err)
	}
	for _, path := range strings.Fields(string(standard)) {
		p := pkgs[path]
		p.Standard = true
		pkgs[path] = p
	}
	paths := []string{}
	for path := range pkgs {
		if within(path, c.module) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		p := pkgs[path]
		if err := c.checkPackage(p, pkgs, t); err != nil {
			return err
		}
	}
	return nil
}

func (c *checker) checkPackage(p listedPackage, all map[string]listedPackage, target targetSpec) error {
	r := classify(strings.TrimPrefix(p.ImportPath, c.module+"/"))
	fset := token.NewFileSet()
	var files []*ast.File
	checked := false
	for _, name := range append(append(append([]string{}, p.GoFiles...), p.TestGoFiles...), p.XTestGoFiles...) {
		path := filepath.Join(p.Dir, name)
		rel, err := filepath.Rel(c.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "../") {
			return fmt.Errorf("module source outside repository: %s", path)
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		test := strings.HasSuffix(name, "_test.go")
		if !test {
			files = append(files, file)
		}
		if c.exempt[rel] {
			continue
		}
		if r == "" {
			return fmt.Errorf("package inventory forbids %s", p.ImportPath)
		}
		for _, spec := range file.Imports {
			imp, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if spec.Name != nil && spec.Name.Name == "." && !test {
				return fmt.Errorf("%s: dot imports hide symbol ownership", rel)
			}
			allowed := false
			if within(imp, c.module) {
				toPath := strings.TrimPrefix(imp, c.module+"/")
				allowed = internalAllowed(r, classify(toPath), test) && privateImportAllowed(r, toPath)
			} else {
				// Includes test-only stdlib ownership from the selected Go command.
				standard := all[imp].Standard
				allowed = externalAllowed(r, imp, standard, test)
			}
			if !allowed {
				return fmt.Errorf("%s: forbidden %s import %q from %s", rel, map[bool]string{true: "test", false: "production"}[test], imp, r)
			}
		}
		if !test {
			checked = true
			if err := syntaxPolicy(file, r); err != nil {
				return fmt.Errorf("%s: %w", rel, err)
			}
		}
	}
	if !checked {
		return nil
	}
	// Go's actual target export data resolves aliases, embedded methods, generics
	// and inferred exported variable/function types without a host-ABI assumption.
	lookup := func(path string) (io.ReadCloser, error) {
		if p := all[path]; p.Export != "" {
			return os.Open(p.Export)
		}
		return nil, fmt.Errorf("missing target export data for %s", path)
	}
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	config := types.Config{Importer: importer.ForCompiler(fset, "gc", lookup), Sizes: types.SizesFor("gc", target.GOARCH)}
	checkedPkg, err := config.Check(p.ImportPath, fset, files, info)
	if err != nil {
		return fmt.Errorf("target type check %s: %w", p.ImportPath, err)
	}
	if publicSurface(strings.TrimPrefix(p.ImportPath, c.module+"/")) {
		if err := publicTypes(checkedPkg, r, c.module); err != nil {
			return fmt.Errorf("%s: %w", p.ImportPath, err)
		}
	}
	if err := symbolPolicy(info, r); err != nil {
		return fmt.Errorf("%s: %w", p.ImportPath, err)
	}
	return nil
}

func syntaxPolicy(file *ast.File, r role) error {
	var violation error
	ast.Inspect(file, func(node ast.Node) bool {
		if violation != nil {
			return false
		}
		if isPure(r) {
			switch n := node.(type) {
			case *ast.GoStmt:
				violation = fmt.Errorf("pure layer cannot start a goroutine")
			case *ast.ChanType:
				violation = fmt.Errorf("pure layer cannot own channels")
			case *ast.CallExpr:
				if id, ok := n.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
					violation = fmt.Errorf("pure layer cannot print")
				}
			}
		}
		if r == domain {
			if f, ok := node.(*ast.Field); ok && f.Tag != nil && strings.Contains(f.Tag.Value, "json:") {
				violation = fmt.Errorf("Domain cannot own JSON schemas/tags")
			}
		}
		if isPure(r) {
			if c, ok := node.(*ast.Comment); ok && strings.HasPrefix(c.Text, "//go:linkname") {
				violation = fmt.Errorf("pure layer cannot bypass imports with go:linkname")
			}
		}
		return true
	})
	return violation
}

func symbolPolicy(info *types.Info, r role) error {
	if !isPure(r) {
		return nil
	}
	allowedTime := set("Time", "Duration", "Month", "Weekday", "UTC", "Nanosecond", "Microsecond", "Millisecond", "Second", "Minute", "Hour", "Date", "Unix", "UnixMilli", "UnixMicro", "ParseDuration", "RFC3339", "RFC3339Nano", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday")
	for _, obj := range info.Uses {
		if obj.Pkg() == nil {
			continue
		}
		path, name := obj.Pkg().Path(), obj.Name()
		if path == "encoding/json" && r == api && name != "Valid" {
			return fmt.Errorf("API permits only the pure json.Valid predicate for OpaqueJSON, not json.%s", name)
		}
		if path == "fmt" && (strings.Contains(name, "Print") && !strings.HasPrefix(name, "S") || strings.Contains(name, "Scan")) {
			return fmt.Errorf("pure layer forbids I/O symbol fmt.%s", name)
		}
		if path == "time" {
			if fn, ok := obj.(*types.Func); ok && fn.Type().(*types.Signature).Recv() != nil {
				if name == "Local" || name == "UnmarshalJSON" || name == "UnmarshalText" || name == "Parse" {
					return fmt.Errorf("pure layer forbids environment-sensitive time.%s", name)
				}
			} else if !allowedTime[name] {
				return fmt.Errorf("pure layer forbids clock/environment symbol time.%s", name)
			}
		}
		if path == "context" && !set("Context", "Canceled", "DeadlineExceeded")[name] {
			return fmt.Errorf("pure contract layer forbids lifecycle symbol context.%s", name)
		}
		if path == "github.com/charmbracelet/lipgloss" && set("NewRenderer", "DefaultRenderer", "SetDefaultRenderer", "ColorProfile", "SetColorProfile", "HasDarkBackground", "SetHasDarkBackground")[name] {
			return fmt.Errorf("pure View forbids environment/global renderer symbol lipgloss.%s", name)
		}
	}
	return nil
}

func publicTypes(pkg *types.Package, r role, module string) error {
	var visit func(types.Type, map[types.Type]bool) error
	// A declared API function or interface method is callable. A signature reached
	// inside one of its values, however deeply nested, is an exposed callback.
	callable := func(sig *types.Signature, seen map[types.Type]bool) error {
		if err := visit(sig.Params(), seen); err != nil {
			return err
		}
		return visit(sig.Results(), seen)
	}
	visit = func(t types.Type, seen map[types.Type]bool) error {
		if t == nil || seen[t] {
			return nil
		}
		seen[t] = true
		children := []types.Type{}
		checkOwner := func(obj *types.TypeName) error {
			if obj.Pkg() == nil || obj.Pkg() == pkg {
				return nil
			}
			path := obj.Pkg().Path()
			if within(path, module) {
				rel := strings.TrimPrefix(path, module+"/")
				to := classify(rel)
				if !internalAllowed(r, to, false) || (isAdapter(to) && to != r) {
					return fmt.Errorf("exported API exposes forbidden type %s.%s", path, obj.Name())
				}
			} else if !publicExternalType(path, obj.Name(), r) {
				return fmt.Errorf("exported API exposes forbidden type %s.%s", path, obj.Name())
			}
			return nil
		}
		switch t := t.(type) {
		case *types.Alias:
			if err := checkOwner(t.Obj()); err != nil {
				return err
			}
			children = append(children, types.Unalias(t))
		case *types.Named:
			if err := checkOwner(t.Obj()); err != nil {
				return err
			}
			if _, callback := t.Underlying().(*types.Signature); callback {
				return fmt.Errorf("exported boundary exposes callback type %s", t.Obj().Name())
			}
			for i := 0; i < t.TypeArgs().Len(); i++ {
				children = append(children, t.TypeArgs().At(i))
			}
			owner := t.Obj().Pkg()
			if owner == pkg || (owner != nil && within(owner.Path(), module) && classify(strings.TrimPrefix(owner.Path(), module+"/")) == r) {
				children = append(children, t.Underlying())
				for i := 0; i < t.NumMethods(); i++ {
					if t.Method(i).Exported() {
						if err := callable(t.Method(i).Type().(*types.Signature), seen); err != nil {
							return err
						}
					}
				}
			}
		case *types.Pointer:
			children = append(children, t.Elem())
		case *types.Slice:
			children = append(children, t.Elem())
		case *types.Array:
			children = append(children, t.Elem())
		case *types.Map:
			children = append(children, t.Key(), t.Elem())
		case *types.Chan:
			return fmt.Errorf("exported boundary exposes a channel")
		case *types.Struct:
			for i := 0; i < t.NumFields(); i++ {
				f := t.Field(i)
				if f.Exported() || f.Embedded() {
					if _, ok := types.Unalias(f.Type()).Underlying().(*types.Signature); ok {
						return fmt.Errorf("exported boundary exposes callback field %s", f.Name())
					}
					children = append(children, f.Type())
				}
			}
		case *types.Signature:
			return fmt.Errorf("exported boundary exposes callback inside a value graph")
		case *types.Tuple:
			for i := 0; i < t.Len(); i++ {
				if callbackType(t.At(i).Type()) {
					return fmt.Errorf("exported boundary exposes callback argument/result")
				}
				children = append(children, t.At(i).Type())
			}
		case *types.Interface:
			if t.Empty() {
				return fmt.Errorf("exported boundary exposes untyped any/interface{} payload")
			}
			for i := 0; i < t.NumMethods(); i++ {
				if err := callable(t.Method(i).Type().(*types.Signature), seen); err != nil {
					return err
				}
			}
			for i := 0; i < t.NumEmbeddeds(); i++ {
				children = append(children, t.EmbeddedType(i))
			}
		case *types.Basic:
			if t.Kind() == types.UnsafePointer || t.Kind() == types.Uintptr {
				return fmt.Errorf("exported boundary exposes native pointer/handle representation")
			}
		case *types.TypeParam: // Constraint any permits typed generic values, not an any DTO.
		}
		for _, child := range children {
			if err := visit(child, seen); err != nil {
				return err
			}
		}
		return nil
	}
	for _, name := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(name)
		if obj.Exported() {
			if function, ok := obj.(*types.Func); ok {
				if err := callable(function.Type().(*types.Signature), map[types.Type]bool{}); err != nil {
					return fmt.Errorf("%s: %w", name, err)
				}
				continue
			}
			if _, function := obj.(*types.Func); !function && callbackType(obj.Type()) {
				return fmt.Errorf("%s: exported boundary exposes callback value/type", name)
			}
			if err := visit(obj.Type(), map[types.Type]bool{}); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}

func callbackType(t types.Type) bool {
	seen := map[types.Type]bool{}
	for {
		t = types.Unalias(t)
		if seen[t] {
			return false
		}
		seen[t] = true
		if p, ok := t.Underlying().(*types.Pointer); ok {
			t = p.Elem()
			continue
		}
		_, ok := t.Underlying().(*types.Signature)
		return ok
	}
}
