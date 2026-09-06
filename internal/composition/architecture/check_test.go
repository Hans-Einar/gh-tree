package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const fixtureModule = "example.test/architecture-fixture"

// These fixtures are compiled and checked, not matched against source strings.
// Every rejection asserts its actual diagnostic so a broken build is not a
// substitute for the architecture policy detecting the intended violation.
func TestPolicyFixtures(t *testing.T) {
	b, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Target, Want string
		Files              map[string]string
	}
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			c := fixture(t)
			for path, source := range tc.Files {
				writeFixture(t, c.root, path, strings.ReplaceAll(source, "$MODULE", fixtureModule))
			}
			err := c.inventory()
			if err == nil {
				target := targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
				if tc.Target != "" {
					p := strings.Split(tc.Target, "/")
					target = targetSpec{GOOS: p[0], GOARCH: p[1]}
				}
				err = c.checkTarget(target)
			}
			if tc.Want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.Want) {
				t.Fatalf("want rejection %q, got %v", tc.Want, err)
			}
		})
	}
}

func TestExactLegacyAllowance(t *testing.T) {
	for _, change := range []string{"unchanged", "crlf", "edited", "renamed", "new-in-legacy-folder", "strict"} {
		t.Run(change, func(t *testing.T) {
			c := fixture(t)
			old := "package app\nconst Value = 1\n"
			c.legacy["internal/app/old.go"] = blobHash([]byte(old))
			path := "internal/app/old.go"
			switch change {
			case "crlf":
				old = strings.ReplaceAll(old, "\n", "\r\n")
				gitFixture(t, c.root, "init", "-q")
				gitFixture(t, c.root, "config", "--local", "core.autocrlf", "true")
			case "edited":
				old += "// changed\n"
			case "renamed":
				path = "internal/application/renamed.go"
			case "new-in-legacy-folder":
				path = "internal/app/new.go"
				old = "package app\nconst New = 2\n"
			case "strict":
				c.mode = "strict"
			}
			writeFixture(t, c.root, path, old)
			err := c.inventory()
			want := map[string]string{"edited": "legacy allowance changed", "renamed": "renamed/copied legacy", "new-in-legacy-folder": "package inventory forbids", "strict": "empty legacy allowance"}[change]
			if want == "" {
				if err != nil {
					t.Fatal(err)
				}
				if !c.exempt[path] {
					t.Fatal("exact baseline not exempted")
				}
			} else if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("want %s, got %v", want, err)
			}
		})
	}
}

func TestSharedEntryMigration(t *testing.T) {
	c := fixture(t)
	old := "package main\nimport _ \"" + fixtureModule + "/internal/app\"\nfunc main(){}\n"
	c.entryBlob = blobHash([]byte(old))
	writeFixture(t, c.root, "cmd/gh-tree/main.go", old)
	if err := c.inventory(); err != nil {
		t.Fatal(err)
	}
	if !c.exempt["cmd/gh-tree/main.go"] {
		t.Fatal("unchanged entry must remain available before M6")
	}
	writeFixture(t, c.root, "cmd/gh-tree/main.go", "package main\nimport _ \""+fixtureModule+"/internal/composition\"\nfunc main(){}\n")
	if err := c.inventory(); err != nil {
		t.Fatal(err)
	}
	if c.exempt["cmd/gh-tree/main.go"] {
		t.Fatal("rewritten entry must obey strict imports")
	}
}

func TestNewCodeCannotImportUnchangedLegacy(t *testing.T) {
	c := fixture(t)
	old := "package github\ntype Facts struct { Value string }\n"
	c.legacy["internal/github/client.go"] = blobHash([]byte(old))
	writeFixture(t, c.root, "internal/github/client.go", old)
	writeFixture(t, c.root, "internal/github/adapter/adapter.go", "package adapter\nimport legacy \""+fixtureModule+"/internal/github\"\nvar Value = legacy.Facts{}\n")
	if err := c.inventory(); err != nil {
		t.Fatal(err)
	}
	err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH})
	if err == nil || !strings.Contains(err.Error(), "forbidden production import") {
		t.Fatalf("unchanged allowance leaked to new caller: %v", err)
	}
}

func TestStrictInventory(t *testing.T) {
	c := fixture(t)
	c.mode = "strict"
	if err := c.inventory(); err == nil || !strings.Contains(err.Error(), "strict final package inventory missing") {
		t.Fatalf("incomplete M1 tree must not pass final gate: %v", err)
	}
	for path, pkg := range map[string]string{
		"cmd/gh-tree": "main", "internal/version": "version", "internal/composition": "composition", "internal/composition/host": "host", "internal/domain": "domain", "internal/application": "application", "internal/application/api": "api", "internal/application/ports": "ports", "internal/application/usecases": "usecases", "internal/git": "git", "internal/github/adapter": "adapter", "internal/runtime": "runtime", "internal/runtime/broker": "broker", "internal/runtime/broker/cmd": "main", "internal/runtime/brokerassets": "brokerassets", "internal/runtime/cmd/helpergen": "main", "internal/launchdiscovery": "launchdiscovery", "internal/persistence": "persistence", "internal/tuistate": "tuistate", "internal/tuistate/viewmodel": "viewmodel", "internal/tuiview": "tuiview",
	} {
		source := "package " + pkg + "\n"
		if pkg == "main" {
			source += "func main(){}\n"
		}
		writeFixture(t, c.root, path+"/main.go", source)
	}
	if err := c.inventory(); err != nil {
		t.Fatal(err)
	}
	if err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	if len(c.exempt) != 0 {
		t.Fatal("strict mode must have no exemptions")
	}
}

func TestRuntimePrerequisites(t *testing.T) {
	c := fixture(t)
	if state, err := runtimePrerequisite(c.root); err != nil || state != "pending-m3" {
		t.Fatalf("missing Runtime must be pending, not conformant: %s %v", state, err)
	}
	writeFixture(t, c.root, "internal/runtime/registry.go", "package runtime\n")
	if _, err := runtimePrerequisite(c.root); err == nil {
		t.Fatal("Runtime without real helper inputs passed")
	}
	for _, path := range []string{"broker/engine.go", "broker/cmd/main.go", "cmd/helpergen/main.go", "brokerassets/broker-amd64.gz", "brokerassets/broker-arm64.gz", "brokerassets/manifest.json"} {
		writeFixture(t, c.root, "internal/runtime/"+path, "fixture prerequisite only; not conformance evidence")
	}
	if state, err := runtimePrerequisite(c.root); err != nil || state != "ready" {
		t.Fatalf("complete inputs should require the real M3 check: %s %v", state, err)
	}
	writeFixture(t, c.root, "internal/runtime/brokerassets/manifest.json", "")
	if _, err := runtimePrerequisite(c.root); err == nil {
		t.Fatal("empty manifest passed")
	}
}

func TestOpaqueJSONValidationAndCallbackBoundary(t *testing.T) {
	c := fixture(t)
	writeFixture(t, c.root, "internal/application/api/opaque.go", "package api\nimport \"encoding/json\"\ntype OpaqueJSON struct { raw []byte }\nfunc Valid(b []byte) bool { return json.Valid(b) }\ntype Recursive *Recursive\n")
	if err := c.inventory(); err != nil {
		t.Fatal(err)
	}
	if err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, c.root, "internal/application/api/callback.go", "package api\nfunc Bad(callback func()) {}\n")
	err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH})
	if err == nil || !strings.Contains(err.Error(), "callback argument/result") {
		t.Fatalf("callback boundary passed: %v", err)
	}
}

func TestPhysicalRepositoryRoot(t *testing.T) {
	c := fixture(t)
	physical, err := filepath.EvalSymlinks(c.root)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "root-alias")
	createDirectoryAlias(t, c.root, alias)
	c.root = alias
	if err := c.inventory(); err != nil {
		t.Fatal(err)
	}
	if c.root != physical {
		t.Fatalf("selected root not physically normalized: %s, expected %s", c.root, physical)
	}
	if err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	// Normalizing the selected root does not allow an out-of-root source.
	outside := t.TempDir()
	writeFixture(t, outside, "outside.go", "package domain\n")
	err = c.checkPackage(listedPackage{Dir: outside, ImportPath: fixtureModule + "/internal/domain", GoFiles: []string{"outside.go"}}, nil, targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH})
	if err == nil || !strings.Contains(err.Error(), "module source outside repository") {
		t.Fatalf("outside source passed after root normalization: %v", err)
	}
}

func TestOutsideChildDirectoryAlias(t *testing.T) {
	c := fixture(t)
	outside := t.TempDir()
	writeFixture(t, outside, "native.go", "package native\nconst Value = \"outside source\"\n")
	writeFixture(t, c.root, "internal/git/git.go", "package git\nimport \""+fixtureModule+"/internal/git/native\"\nfunc Value() string {return native.Value}\n")
	createDirectoryAlias(t, outside, filepath.Join(c.root, "internal/git/native"))
	if err := c.inventory(); err == nil || !strings.Contains(err.Error(), "child alias") {
		t.Fatalf("inventory admitted source directory alias: %v", err)
	}
	// Exercise the selected-source check independently too: Go can report a
	// lexical Dir for a package reached through the junction/symlink import.
	err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH})
	if err == nil || !strings.Contains(err.Error(), "module source outside repository") {
		t.Fatalf("selected source admitted outside object: %v", err)
	}
}

func TestInsideChildAliasCannotRelabelOwner(t *testing.T) {
	for _, targetPath := range []string{"internal/persistence/native", "internal/app"} {
		t.Run(targetPath, func(t *testing.T) {
			c := fixture(t)
			source := "package native\nconst Value = \"different physical owner\"\n"
			writeFixture(t, c.root, targetPath+"/native.go", source)
			if targetPath == "internal/app" {
				c.legacy[targetPath+"/native.go"] = blobHash([]byte(source))
			}
			writeFixture(t, c.root, "internal/git/git.go", "package git\nimport \""+fixtureModule+"/internal/git/native\"\nfunc Value()string{return native.Value}\n")
			createDirectoryAlias(t, filepath.Join(c.root, targetPath), filepath.Join(c.root, "internal/git/native"))
			if err := c.inventory(); err == nil || !strings.Contains(err.Error(), "child alias") {
				t.Fatalf("inventory relabeled another owner: %v", err)
			}
			err := c.checkTarget(targetSpec{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH})
			if err == nil || !strings.Contains(err.Error(), "physical ownership") {
				t.Fatalf("selected source relabeled another owner: %v", err)
			}
		})
	}
}

func createDirectoryAlias(t *testing.T, target, alias string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		// Native junctions need no symlink privilege. Both locations were created
		// under this test's disposable root; no user configuration is involved.
		cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", "New-Item -ItemType Junction -Path $env:GH_TREE_ARCH_LINK -Target $env:GH_TREE_ARCH_LINK_TARGET -ErrorAction Stop | Out-Null")
		cmd.Env = append(os.Environ(), "GH_TREE_ARCH_LINK="+alias, "GH_TREE_ARCH_LINK_TARGET="+target)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("native directory junction: %v\n%s", err, out)
		}
	} else if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}
	// Remove only the link before the temporary-root cleanups run. Never walk
	// through it to remove the target; that target has its own test owner.
	t.Cleanup(func() {
		if err := os.Remove(alias); err != nil {
			t.Errorf("remove owned directory alias: %v", err)
		}
	})
}

func TestAcceptedTargetsAndBlobIdentity(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateTargets(root); err != nil {
		t.Fatal(err)
	}
	c, err := newChecker(root, "staged")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.prepareCleanPolicy(); err != nil {
		t.Fatal(err)
	}
	// Compare every still-present allowance with the accepted Git blob, including
	// Windows text normalization. This is independent of fixture-generated hashes.
	for path, want := range c.legacy {
		b, err := os.ReadFile(filepath.Join(root, path))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.cleanBlobHash(path, b)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s: actual %s baseline %s", path, got, want)
		}
	}
}

func TestPathSpecificCleanProfile(t *testing.T) {
	for _, profile := range []string{"minus-text", "plus-text", "autocrlf-false", "autocrlf-true", "filter-refused", "legacy-minus-crlf", "legacy-plus-crlf", "legacy-crlf-input"} {
		t.Run(profile, func(t *testing.T) {
			c := fixture(t)
			gitFixture(t, c.root, "init", "-q")
			gitFixture(t, c.root, "config", "--local", "core.autocrlf", "false")
			old := "package app\nconst Value = 1\n"
			c.legacy["internal/app/old.go"] = blobHash([]byte(old))
			writeFixture(t, c.root, "internal/app/old.go", strings.ReplaceAll(old, "\n", "\r\n"))
			switch profile {
			case "minus-text":
				writeFixture(t, c.root, ".gitattributes", "internal/app/old.go -text\n")
			case "plus-text":
				writeFixture(t, c.root, ".gitattributes", "internal/app/old.go text\n")
			case "autocrlf-true":
				gitFixture(t, c.root, "config", "--local", "core.autocrlf", "true")
			case "legacy-minus-crlf", "legacy-plus-crlf", "legacy-crlf-input":
				gitFixture(t, c.root, "config", "--local", "core.autocrlf", "true")
				attr := map[string]string{"legacy-minus-crlf": "-crlf", "legacy-plus-crlf": "crlf", "legacy-crlf-input": "crlf=input"}[profile]
				writeFixture(t, c.root, ".gitattributes", "internal/app/old.go "+attr+"\n")
			case "filter-refused":
				writeFixture(t, c.root, ".gitattributes", "internal/app/old.go filter=probe\n")
				gitFixture(t, c.root, "config", "--local", "filter.probe.clean", "echo FILTER-RAN > filter-ran")
			}
			err := c.inventory()
			want := map[string]string{"minus-text": "legacy allowance changed", "autocrlf-false": "legacy allowance changed", "filter-refused": "unsupported Git clean profile"}[profile]
			if strings.HasPrefix(profile, "legacy-") {
				want = "unsupported Git legacy crlf profile"
			}
			if want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("want %s, got %v", want, err)
			}
			if _, err := os.Stat(filepath.Join(c.root, "filter-ran")); !os.IsNotExist(err) {
				t.Fatal("checker executed a clean filter")
			}
		})
	}
}

func TestTrimpathToolStdlib(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", "-trimpath", "./internal/composition/architecture", "-target", runtime.GOOS+"/"+runtime.GOARCH)
	cmd.Dir = root
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if !strings.EqualFold(key, "GOROOT") && !strings.EqualFold(key, "GOFLAGS") {
			cmd.Env = append(cmd.Env, item)
		}
	}
	cmd.Env = append(cmd.Env, "GOFLAGS=-trimpath")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("trimpath checker without explicit GOROOT: %v\n%s", err, out)
	}
}

func gitFixture(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture git %v: %v\n%s", args, err, out)
	}
}

func fixture(t *testing.T) *checker {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, "go.mod", "module "+fixtureModule+"\n\ngo 1.25.0\n")
	for path, pkg := range map[string]string{"internal/domain": "domain", "internal/application/api": "api", "internal/application/ports": "ports", "internal/tuistate/viewmodel": "viewmodel"} {
		writeFixture(t, root, path+"/value.go", "package "+pkg+"\ntype Value struct { N int }\n")
	}
	return &checker{root: root, module: fixtureModule, mode: "staged", legacy: map[string]string{}, exempt: map[string]bool{}}
}

func writeFixture(t *testing.T, root, rel, source string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}
