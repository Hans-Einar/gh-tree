// Command architecture checks the accepted CR-21 package boundaries. It is build
// tooling owned by Composition, and is never imported by the product.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", ".", "repository root")
	mode := flag.String("mode", "staged", "staged or strict (empty legacy allowance)")
	target := flag.String("target", "", "one supported GOOS/GOARCH; default checks all twelve")
	matrix := flag.Bool("targets", false, "print the complete accepted CI matrix as JSON")
	helper := flag.Bool("runtime-prerequisite", false, "print pending-m3 or ready; reject incomplete Runtime helper inputs")
	flag.Parse()
	if flag.NArg() != 0 || (*mode != "staged" && *mode != "strict") {
		fail(fmt.Errorf("expected no operands and -mode staged|strict"))
	}
	abs, err := filepath.Abs(*root)
	if err != nil {
		fail(err)
	}
	if err := validateTargets(abs); err != nil {
		fail(err)
	}
	if *matrix {
		if err := json.NewEncoder(os.Stdout).Encode(struct {
			Include []targetSpec `json:"include"`
		}{targets}); err != nil {
			fail(err)
		}
		return
	}
	if *helper {
		state, err := runtimePrerequisite(abs)
		if err != nil {
			fail(err)
		}
		if state == "pending-m3" {
			fmt.Fprintln(os.Stderr, "Runtime helper conformance NOT RUN: M3 Runtime/helper inputs have not been integrated")
		}
		fmt.Println(state)
		return
	}
	selected := targets
	if *target != "" {
		selected = nil
		for _, t := range targets {
			if t.GOOS+"/"+t.GOARCH == *target {
				selected = append(selected, t)
			}
		}
		if len(selected) == 0 {
			fail(fmt.Errorf("unsupported target %q", *target))
		}
	}
	c, err := newChecker(abs, *mode)
	if err != nil {
		fail(err)
	}
	if err := c.inventory(); err != nil {
		fail(err)
	}
	for _, t := range selected {
		fmt.Printf("architecture: %s/%s (%s)\n", t.GOOS, t.GOARCH, *mode)
		if err := c.checkTarget(t); err != nil {
			fail(err)
		}
	}
	fmt.Printf("architecture: PASS, %d target selections; %d exact legacy/shared-entry allowances used; ownership review still required\n", len(selected), len(c.exempt))
}

func fail(err error) { fmt.Fprintln(os.Stderr, "architecture:", err); os.Exit(1) }

type targetSpec struct {
	GOOS   string `json:"goos"`
	GOARCH string `json:"goarch"`
	Asset  string `json:"asset"`
}

var targets = []targetSpec{
	{"darwin", "amd64", "darwin-amd64"}, {"darwin", "arm64", "darwin-arm64"},
	{"freebsd", "386", "freebsd-386"}, {"freebsd", "amd64", "freebsd-amd64"}, {"freebsd", "arm64", "freebsd-arm64"},
	{"linux", "386", "linux-386"}, {"linux", "amd64", "linux-amd64"}, {"linux", "arm", "linux-arm"}, {"linux", "arm64", "linux-arm64"},
	{"windows", "386", "windows-386.exe"}, {"windows", "amd64", "windows-amd64.exe"}, {"windows", "arm64", "windows-arm64.exe"},
}

// The accepted table is the independent inventory authority, so dropping a row
// from this executable cannot silently reduce the dynamic CI matrix.
func validateTargets(root string) error {
	b, err := os.ReadFile(filepath.Join(root, "SDP/Design/CR-#21/Verification--001.md"))
	if err != nil {
		return err
	}
	want := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) == 4 {
			pair := strings.TrimSpace(cells[1])
			if strings.Count(pair, "/") == 1 && !strings.ContainsAny(pair, " `") {
				want[pair] = strings.TrimSpace(cells[2])
			}
		}
	}
	if len(want) != 12 || len(targets) != 12 {
		return fmt.Errorf("accepted release matrix must contain exactly twelve targets")
	}
	seen := map[string]bool{}
	for _, t := range targets {
		key := t.GOOS + "/" + t.GOARCH
		if seen[key] || want[key] != t.Asset {
			return fmt.Errorf("release matrix differs from Verification--001: %s (%s)", key, t.Asset)
		}
		seen[key] = true
	}
	return nil
}

func runtimePrerequisite(root string) (string, error) {
	dir := filepath.Join(root, "internal/runtime")
	present := false
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if os.IsNotExist(err) && path == dir {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == "testdata" {
			return filepath.SkipDir
		}
		if !entry.IsDir() && (strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".gz")) {
			present = true
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if !present {
		return "pending-m3", nil
	}
	for _, rel := range []string{"broker", "broker/cmd", "cmd/helpergen"} {
		files, err := filepath.Glob(filepath.Join(dir, rel, "*.go"))
		if err != nil || len(files) == 0 {
			return "", fmt.Errorf("M3 prerequisite missing: internal/runtime/%s Go source", rel)
		}
	}
	for _, rel := range []string{"brokerassets/broker-amd64.gz", "brokerassets/broker-arm64.gz", "brokerassets/manifest.json"} {
		info, err := os.Stat(filepath.Join(dir, rel))
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return "", fmt.Errorf("M3 prerequisite missing/empty: internal/runtime/%s", rel)
		}
	}
	return "ready", nil
}
