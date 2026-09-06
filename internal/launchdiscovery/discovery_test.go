package launchdiscovery

import (
	"context"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func put(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(path, []byte(body), 0600); e != nil {
		t.Fatal(e)
	}
}
func discover(t *testing.T, a *Adapter, s api.WorktreeScope, saved ...api.SavedLaunchEntry) api.DiscoveryResultData {
	t.Helper()
	var version api.Optional[api.StorageVersion]
	if saved != nil {
		version = api.Some(must(api.NewRunStorageVersion(s, "fixture-store", true, 1, [32]byte{1})))
	}
	r, e := a.Discover(context.Background(), must(api.NewDiscoveryRequest(api.DiscoveryRequestData{Worktree: s, Saved: saved, SavedVersion: version})))
	if e != nil {
		t.Fatal(e)
	}
	if !r.Valid() {
		t.Fatal("invalid result")
	}
	return r.Data()
}
func pick(t *testing.T, defs []api.LaunchDefinition, provider api.ProviderKind, project, member string) api.LaunchDefinition {
	t.Helper()
	for _, d := range defs {
		v := d.Data()
		if v.Provider == provider && strings.Join(v.ProjectComponents, "/") == project && v.Member == member {
			return d
		}
	}
	t.Fatalf("missing %s %s %s in %v", providerKey(provider), project, member, defs)
	return api.LaunchDefinition{}
}
func memberOf(d api.LaunchDefinition) api.MemberSelection {
	return must(api.NewMemberSelection(api.MemberSelectionData{LaunchPointID: d.Data().LaunchPointID, SourceVersion: d.Data().ProjectSource.Data().Content}))
}
func resolve(t *testing.T, a *Adapter, s api.WorktreeScope, selection api.LaunchSelection, saved []api.SavedLaunchEntry, version api.Optional[api.StorageVersion]) (api.ResolveLaunchResult, error) {
	t.Helper()
	return a.Resolve(context.Background(), must(api.NewResolveLaunchRequest(api.ResolveLaunchRequestData{Worktree: s, Selection: selection, Saved: saved, SavedVersion: version, Geometry: must(api.NewGeometry(api.GeometryData{Rows: 24, Columns: 80}))})))
}
func TestDiscoverIdentityPartialAndLiteralResolve(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev:wan":"echo never-run","a/b":"x","dev":"x"}}`)
	put(t, root, "Makefile", "dev:\n")
	put(t, root, "a/package.json", `{"scripts":{"b":"x"}}`)
	put(t, root, " project/package.json", `{"scripts":{" dev:wan ":"x"}}`)
	put(t, root, "broken/package.json", "{")
	put(t, root, "NoDe_Modules/package.json", `{"scripts":{"excluded":"x"}}`)
	put(t, root, ".gh-tree/run.json", "must never be loaded")
	a := must(New(Config{}))
	s := fixtureScope(t, root)
	r := discover(t, a, s)
	if len(r.Definitions) != 6 || r.Observation.Data().Completeness != api.Partial || len(r.Diagnostics) == 0 {
		t.Fatal(len(r.Definitions), r.Observation.Data(), r.Diagnostics)
	}
	x, y := pick(t, r.Definitions, api.Npm, "", "a/b"), pick(t, r.Definitions, api.Npm, "a", "b")
	if x.Data().LaunchPointID == y.Data().LaunchPointID {
		t.Fatal("path/member collision")
	}
	x, y = pick(t, r.Definitions, api.Npm, "", "dev"), pick(t, r.Definitions, api.Make, "", "dev")
	if x.Data().LaunchPointID == y.Data().LaunchPointID {
		t.Fatal("provider collision")
	}
	def := pick(t, r.Definitions, api.Npm, " project", " dev:wan ")
	out, e := resolve(t, a, s, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(def)})), nil, api.None[api.StorageVersion]())
	if e != nil {
		t.Fatal(e, out.Data())
	}
	inv, ok := out.Data().Invocation.Value()
	if !ok {
		t.Fatal(out.Data())
	}
	exec := inv.Data().Execution.(api.ArgvExecution).Data()
	if exec.Executable != "npm" || !reflect.DeepEqual(exec.Arguments, []string{"run", " dev:wan "}) || !reflect.DeepEqual(inv.Data().Cwd.Data().ProjectComponents, []string{" project"}) {
		t.Fatal(inv.Data())
	}
	copy := r.Definitions[0].Data()
	copy.ProjectComponents = append(copy.ProjectComponents, "evil")
	args := def.Data().EffectiveExecutable.Data()
	args.Arguments[1] = "evil"
	if def.Data().EffectiveExecutable.Data().Arguments[1] != " dev:wan " {
		t.Fatal("mutable output")
	}
	put(t, root, " project/package.json", `{"scripts":{}}`)
	out, e = resolve(t, a, s, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(def)})), nil, api.None[api.StorageVersion]())
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("removed member accepted")
	}
}
func TestManagerObservationsAndDrift(t *testing.T) {
	for _, tc := range []struct {
		locks []string
		want  string
	}{{nil, "npm"}, {[]string{"yarn.lock"}, "yarn"}, {[]string{"yarn.lock", "pnpm-lock.yaml"}, "pnpm"}, {[]string{"package-lock.json"}, "npm"}} {
		t.Run(tc.want+strings.Join(tc.locks, "-"), func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "package.json", `{"packageManager":"yarn@4","scripts":{"dev:wan":"x"}}`)
			for _, lock := range tc.locks {
				put(t, root, lock, "lock")
			}
			a := must(New(Config{}))
			s := fixtureScope(t, root)
			r := discover(t, a, s)
			def := pick(t, r.Definitions, api.Npm, "", "dev:wan")
			if def.Data().EffectiveExecutable.Data().Executable != tc.want {
				t.Fatal(def.Data())
			}
			put(t, root, "pnpm-lock.yaml", "new")
			if tc.want != "pnpm" {
				out, e := resolve(t, a, s, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(def)})), nil, api.None[api.StorageVersion]())
				if e == nil || out.Data().Invocation.Present() {
					t.Fatal("manager drift accepted")
				}
			}
		})
	}
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev":"x"}}`)
	if e := os.Mkdir(filepath.Join(root, "pnpm-lock.yaml"), 0700); e != nil {
		t.Fatal(e)
	}
	a := must(New(Config{}))
	r := discover(t, a, fixtureScope(t, root))
	def := pick(t, r.Definitions, api.Npm, "", "dev")
	if def.Data().EffectiveExecutable.Data().Executable != "npm" || len(def.Data().Diagnostics) == 0 {
		t.Fatal(def.Data())
	}
}
func TestMakePrecedenceOrderAndExactVersion(t *testing.T) {
	root := t.TempDir()
	put(t, root, "GNUmakefile", "all clean: dep\n")
	put(t, root, "Makefile", "wrong:\n")
	a := must(New(Config{}))
	s := fixtureScope(t, root)
	r := discover(t, a, s)
	all, clean := pick(t, r.Definitions, api.Make, "", "all"), pick(t, r.Definitions, api.Make, "", "clean")
	selection := must(api.NewOrderedMakeLaunch(api.OrderedMakeLaunchData{Members: []api.MemberSelection{memberOf(clean), memberOf(all)}}))
	out, e := resolve(t, a, s, selection, nil, api.None[api.StorageVersion]())
	if e != nil {
		t.Fatal(e)
	}
	inv, _ := out.Data().Invocation.Value()
	if !reflect.DeepEqual(inv.Data().Execution.(api.ArgvExecution).Data().Arguments, []string{"-f", "GNUmakefile", "clean", "all"}) {
		t.Fatal(inv.Data())
	}
	if len(r.Definitions) != 2 {
		t.Fatal(r.Definitions)
	}
	put(t, root, "GNUmakefile", "all clean: changed\n")
	out, e = resolve(t, a, s, selection, nil, api.None[api.StorageVersion]())
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("stale Make accepted")
	}
}
func savedEntry(alias, provider, dir, script, command string, targets []string) api.SavedLaunchEntry {
	d := api.SavedLaunchDefinitionData{Provider: provider, Dir: api.PresentField(dir), Script: api.AbsentField[string](), Targets: api.AbsentField[[]string](), Command: api.AbsentField[string](), UnknownMembers: must(api.NewJSONMembers(nil))}
	if script != "" {
		d.Script = api.PresentField(script)
	}
	if targets != nil {
		d.Targets = api.PresentField(targets)
	}
	if command != "" {
		d.Command = api.PresentField(command)
	}
	return must(api.NewSavedLaunchEntry(api.SavedLaunchEntryData{Alias: alias, Definition: must(api.NewSavedLaunchDefinition(d))}))
}
func TestSavedOverridesUnknownAndOrderedBinding(t *testing.T) {
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"dev":"x"}}`)
	put(t, root, "Makefile", "all clean:\n")
	a := must(New(Config{}))
	s := fixtureScope(t, root)
	entries := []api.SavedLaunchEntry{savedEntry(" alias ", "npm", "", "dev", " custom executable ", nil), savedEntry("stack", "make", "", "", "custom-make", []string{"clean", "all"}), savedEntry("unknown", "future", "", "", "", nil)}
	r := discover(t, a, s, entries...)
	if len(r.Saved) != 3 || r.Saved[2].Data().Definition.Present() || len(r.Saved[2].Data().Diagnostics) == 0 {
		t.Fatal(r.Saved)
	}
	for _, i := range []int{0, 1} {
		o := r.Saved[i].Data()
		id, _ := o.LaunchPointID.Value()
		source, _ := o.SourceVersion.Value()
		selection := must(api.NewSavedLaunch(api.SavedLaunchData{Alias: o.Alias, LaunchPointID: id, StorageVersion: o.StorageVersion, SourceExpectation: source}))
		out, e := resolve(t, a, s, selection, entries, api.Some(o.StorageVersion))
		if e != nil {
			t.Fatal(e, out.Data())
		}
		inv, _ := out.Data().Invocation.Value()
		exec := inv.Data().Execution.(api.ArgvExecution).Data()
		if i == 0 && exec.Executable != " custom executable " {
			t.Fatal(exec)
		}
		manifest := "Makefile"
		if runtime.GOOS == "windows" {
			manifest = "makefile"
		}
		if i == 1 && !reflect.DeepEqual(exec.Arguments, []string{"-f", manifest, "clean", "all"}) {
			t.Fatal(exec)
		}
		resolved, _ := out.Data().Definition.Value()
		alias, _ := resolved.Data().Alias.Value()
		if alias != o.Alias {
			t.Fatal(alias)
		}
		changed := append([]api.SavedLaunchEntry(nil), entries...)
		if i == 0 {
			changed[i] = savedEntry(o.Alias, "npm", "", "dev", "other", nil)
		} else {
			changed[i] = savedEntry(o.Alias, "make", "", "", "custom-make", []string{"all", "clean"})
		}
		out, e = resolve(t, a, s, selection, changed, api.Some(o.StorageVersion))
		if e == nil || out.Data().Invocation.Present() {
			t.Fatal("saved intent drift accepted")
		}
	}
}
func TestConstructionLimitsCancellationAndConcurrentCopies(t *testing.T) {
	if _, e := New(Config{Providers: []api.ProviderKind{api.Npm, api.Npm}}); e == nil {
		t.Fatal("duplicate registry")
	}
	limits := DefaultLimits()
	limits.Directories = 0
	if _, e := New(Config{Limits: limits}); e == nil {
		t.Fatal("invalid limits")
	}
	root := t.TempDir()
	put(t, root, "package.json", `{"scripts":{"a":"x","b":"x","c":"x"}}`)
	put(t, root, "nested/package.json", `{"scripts":{"d":"x"}}`)
	s := fixtureScope(t, root)
	limits = DefaultLimits()
	limits.Candidates = 1
	a := must(New(Config{Limits: limits}))
	r := discover(t, a, s)
	if len(r.Definitions) != 1 || r.Observation.Data().Completeness == api.Complete {
		t.Fatal(r)
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	out, e := a.Discover(c, must(api.NewDiscoveryRequest(api.DiscoveryRequestData{Worktree: s})))
	if e != context.Canceled || !out.Valid() || out.Data().Observation.Data().Completeness == api.Complete {
		t.Fatal(out, e)
	}
	a = must(New(Config{}))
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := a.Discover(context.Background(), must(api.NewDiscoveryRequest(api.DiscoveryRequestData{Worktree: s})))
			if e != nil || len(r.Data().Definitions) != 4 {
				t.Errorf("concurrent scan: %v", e)
			}
			if len(r.Data().Definitions) > 0 {
				b := r.Data().Definitions[0].Data().EffectiveExecutable.Data().Arguments
				b[0] = "mutated"
			}
		}()
	}
	wg.Wait()
}
