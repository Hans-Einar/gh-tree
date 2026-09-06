package main

import (
	"encoding/json"
	"os"
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
		if got := blobHash(b); got != want {
			t.Fatalf("%s: actual %s baseline %s", path, got, want)
		}
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
