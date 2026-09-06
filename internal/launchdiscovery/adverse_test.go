package launchdiscovery

import (
	"context"
	"fmt"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

func TestExactDefaultBoundsAndDeterministicPartialScan(t *testing.T) {
	if DefaultLimits() != (Limits{5, 10000, 10000, 4 << 20, 1 << 20}) {
		t.Fatal(DefaultLimits())
	}
	root := t.TempDir()
	for _, name := range []string{"z", "b", "a", "y"} {
		put(t, root, name+"/package.json", `{"scripts":{"dev":"x"}}`)
	}
	limits := DefaultLimits()
	limits.Directories = 3
	a := must(New(Config{Limits: limits}))
	scope := fixtureScope(t, root)
	r := discover(t, a, scope)
	if len(r.Definitions) != 2 || strings.Join(r.Definitions[0].Data().ProjectComponents, "/") != "a" || strings.Join(r.Definitions[1].Data().ProjectComponents, "/") != "b" || r.Observation.Data().Visited != 3 || r.Observation.Data().Completeness == api.Complete || len(r.Diagnostics) == 0 {
		t.Fatal(r)
	}
	deep := t.TempDir()
	put(t, deep, "a/b/c/d/e/package.json", `{"scripts":{"five":"x"}}`)
	put(t, deep, "a/b/c/d/e/f/package.json", `{"scripts":{"six":"x"}}`)
	a = must(New(Config{}))
	r = discover(t, a, fixtureScope(t, deep))
	if len(r.Definitions) != 1 || r.Definitions[0].Data().Member != "five" || r.Observation.Data().Completeness == api.Complete {
		t.Fatal(r)
	}
	capRoot := t.TempDir()
	var b strings.Builder
	b.WriteString(`{"scripts":{`)
	for i := 0; i < 10001; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%q:%q", fmt.Sprintf("script%05d", i), "x")
	}
	b.WriteString("}}")
	put(t, capRoot, "package.json", b.String())
	r = discover(t, a, fixtureScope(t, capRoot))
	if len(r.Definitions) != 10000 || r.Observation.Data().Completeness == api.Complete {
		t.Fatal(len(r.Definitions), r.Observation.Data())
	}
}
func TestManifestAndMakeLineByteCaps(t *testing.T) {
	root := t.TempDir()
	prefix, suffix := `{"scripts":{"dev":"`, `"}}`
	body := prefix + strings.Repeat("x", DefaultLimits().ManifestBytes-len(prefix)-len(suffix)) + suffix
	put(t, root, "package.json", body)
	a := must(New(Config{}))
	scope := fixtureScope(t, root)
	r := discover(t, a, scope)
	if len(r.Definitions) != 1 {
		t.Fatal("exact manifest cap refused", r.Diagnostics)
	}
	put(t, root, "package.json", body+" ")
	r = discover(t, a, scope)
	if len(r.Definitions) != 0 || r.Observation.Data().Completeness == api.Complete {
		t.Fatal("oversize manifest accepted")
	}
	line := "all:" + strings.Repeat(" ", DefaultLimits().MakeLineBytes-4)
	if p, e := parseMake(context.Background(), []byte(line), DefaultLimits().MakeLineBytes); e != nil || len(p.members) != 1 {
		t.Fatal(e, p)
	}
	if _, e := parseMake(context.Background(), []byte(line+" "), DefaultLimits().MakeLineBytes); e == nil {
		t.Fatal("oversize Make line accepted")
	}
}

type cancelAtContext struct {
	context.Context
	cancel context.CancelFunc
	calls  atomic.Int64
	at     int64
}

func (c *cancelAtContext) Err() error {
	if c.calls.Add(1) == c.at {
		c.cancel()
	}
	return c.Context.Err()
}
func TestCancellationPreservesIndependentRootFacts(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"root":"x"}}`)
	var b strings.Builder
	b.WriteString(`{"scripts":{`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%q:\"x\"", fmt.Sprint(i))
	}
	b.WriteString("}}")
	put(t, root, "nested/package.json", b.String())
	scope := fixtureScope(t, root)
	a := must(New(Config{}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	controlled := &cancelAtContext{Context: ctx, cancel: cancel, at: 100}
	out, e := a.Discover(controlled, must(api.NewDiscoveryRequest(api.DiscoveryRequestData{Worktree: scope})))
	if e != context.Canceled || len(out.Data().Definitions) != 1 || out.Data().Definitions[0].Data().Member != "root" || out.Data().Observation.Data().Completeness == api.Complete {
		t.Fatal(e, out.Data())
	}
}
func TestForeignSourcesSavedAliasesAndOrderRefuse(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev":"x"}}`)
	put(t, root, "one/Makefile", "all:\n")
	put(t, root, "two/Makefile", "all:\n")
	a := must(New(Config{}))
	scope := fixtureScope(t, root)
	r := discover(t, a, scope)
	def := pick(t, r.Definitions, api.Npm, "", "dev")
	selection := must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(def)}))
	other := must(New(Config{}))
	out, e := resolve(t, other, scope, selection, nil, api.None[api.StorageVersion]())
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("foreign issuer accepted")
	}
	one, two := pick(t, r.Definitions, api.Make, "one", "all"), pick(t, r.Definitions, api.Make, "two", "all")
	if _, e := api.NewOrderedMakeLaunch(api.OrderedMakeLaunchData{Members: []api.MemberSelection{memberOf(one), memberOf(two)}}); e == nil {
		t.Fatal("cross-project selection constructor")
	}
	entry := savedEntry("alias", "npm", "", "dev", "", nil)
	saved := discover(t, a, scope, entry).Saved[0].Data()
	id, _ := saved.LaunchPointID.Value()
	source, _ := saved.SourceVersion.Value()
	savedSelection := must(api.NewSavedLaunch(api.SavedLaunchData{Alias: "alias", LaunchPointID: id, StorageVersion: saved.StorageVersion, SourceExpectation: source}))
	out, e = resolve(t, a, scope, savedSelection, []api.SavedLaunchEntry{entry, entry}, api.Some(saved.StorageVersion))
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("ambiguous alias accepted")
	}
	dup := discover(t, a, scope, entry, entry)
	for _, obs := range dup.Saved {
		if obs.Data().Definition.Present() {
			t.Fatal("duplicate saved authority")
		}
	}
	foreign := scope.Data()
	foreign.ID = must(domain.NewWorktreeID(scope.Data().ID.Repository(), "foreign"))
	if _, e := api.NewResolveLaunchRequest(api.ResolveLaunchRequestData{Worktree: must(api.NewWorktreeScope(foreign)), Selection: selection, Geometry: must(api.NewGeometry(api.GeometryData{Rows: 1, Columns: 1}))}); e == nil {
		t.Fatal("foreign worktree admitted")
	}
}

func TestSameSizeSourceEditAndNewPreferredMakefile(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev":"old"}}`)
	put(t, root, "Makefile", "all:\n")
	a := must(New(Config{}))
	scope := fixtureScope(t, root)
	r := discover(t, a, scope)
	npm := pick(t, r.Definitions, api.Npm, "", "dev")
	makeDef := pick(t, r.Definitions, api.Make, "", "all")
	info := must(os.Stat(filepath.Join(root, "package.json")))
	put(t, root, "package.json", `{"scripts":{"dev":"new"}}`)
	if e := os.Chtimes(filepath.Join(root, "package.json"), info.ModTime(), info.ModTime()); e != nil {
		t.Fatal(e)
	}
	out, e := resolve(t, a, scope, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(npm)})), nil, api.None[api.StorageVersion]())
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("same-size restored-mtime edit accepted")
	}
	put(t, root, "GNUmakefile", "all:\n")
	out, e = resolve(t, a, scope, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(makeDef)})), nil, api.None[api.StorageVersion]())
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("preferred manifest substitution accepted")
	}
}
func TestPassiveOperationsAndProductionSurface(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev":"touch MUST_NOT_EXIST","predev":"touch MUST_NOT_EXIST"}}`)
	put(t, root, "Makefile", "all:\n\ttouch MUST_NOT_EXIST\n")
	put(t, root, ".gh-tree/run.json", "invalid: this is never a Discovery input")
	put(t, root, "npm", "touch MUST_NOT_EXIST\n")
	before := treeBytes(t, root)
	cwd := must(os.Getwd())
	a := must(New(Config{}))
	scope := fixtureScope(t, root)
	r := discover(t, a, scope)
	for _, def := range r.Definitions {
		out, e := resolve(t, a, scope, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(def)})), nil, api.None[api.StorageVersion]())
		if e != nil || !out.Data().Invocation.Present() {
			t.Fatal(e)
		}
	}
	if !reflect.DeepEqual(before, treeBytes(t, root)) || must(os.Getwd()) != cwd {
		t.Fatal("passive operation changed files/cwd")
	}
	files := must(filepath.Glob("*.go"))
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f := must(parser.ParseFile(token.NewFileSet(), path, nil, 0))
		for _, i := range f.Imports {
			if i.Path.Value == `"os/exec"` || strings.Contains(i.Path.Value, "/internal/runtime") || strings.Contains(i.Path.Value, "/internal/persistence") || strings.Contains(i.Path.Value, "/internal/git") {
				t.Fatal("forbidden dependency", path, i.Path.Value)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			s, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			base, ok := s.X.(*ast.Ident)
			if !ok {
				return true
			}
			name := base.Name + "." + s.Sel.Name
			switch name {
			case "os.Chdir", "os.WriteFile", "os.Create", "os.Mkdir", "os.MkdirAll", "syscall.Exec", "syscall.ForkExec", "windows.CreateProcess", "windows.ShellExecute":
				t.Error("forbidden operation", path, name)
			}
			return true
		})
	}
}
func treeBytes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	e := filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			out[path] = "directory"
			return nil
		}
		b, e := os.ReadFile(path)
		out[path] = string(b)
		return e
	})
	if e != nil {
		t.Fatal(e)
	}
	return out
}
