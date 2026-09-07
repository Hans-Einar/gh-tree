package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This exercises the actual Go dependency selector against an owned source
// snapshot. It never edits the repository or shared module cache.
func TestActualSelectedDependencyClosure(t *testing.T) {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" || runtime.Version() != "go1.25.0" {
		t.Skip("canonical Windows/amd64 Go1.25.0 go-list execution; portable pure controls run separately")
	}
	root, e := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if e != nil {
		t.Fatal(e)
	}
	p, e := capture(root)
	if e != nil {
		t.Fatal(e)
	}
	fixture := t.TempDir()
	write := func(path string, b []byte) {
		t.Helper()
		path = filepath.Join(fixture, path)
		if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	for _, f := range p.files {
		if f.repoPath != "" {
			write(f.repoPath, f.bytes)
		}
	}
	baseline, e := capture(fixture)
	if e != nil {
		t.Fatal(e)
	}
	apiPath := "internal/application/api/runtime.go"
	// Select an actual API source from go list instead of assuming a filename.
	for _, s := range baseline.manifest.Sources {
		if strings.HasPrefix(s.Path, "repo/internal/application/api/") && strings.HasSuffix(s.Path, ".go") {
			apiPath = strings.TrimPrefix(s.Path, "repo/")
			break
		}
	}
	before, e := os.ReadFile(filepath.Join(fixture, apiPath))
	if e != nil {
		t.Fatal(e)
	}
	write(apiPath, append(append([]byte{}, before...), []byte("\n// dependency changed without changing exported symbols\n")...))
	changed, e := capture(fixture)
	if e != nil {
		t.Fatal(e)
	}
	if changed.manifest.SourceDigest == baseline.manifest.SourceDigest {
		t.Fatal("actual dependency source change missed")
	}
	write(apiPath, bytes.ReplaceAll(before, []byte("\n"), []byte("\r\n")))
	normalized, e := capture(fixture)
	if e != nil {
		t.Fatal(e)
	}
	if normalized.manifest.SourceDigest != baseline.manifest.SourceDigest {
		t.Fatal("CRLF checkout changed normalized provenance")
	}
	write(apiPath, before)
	write("internal/runtime/broker/closure_linux.go", []byte("package broker\nconst unusedOnWindows = 1\n"))
	unselected, e := capture(fixture)
	if e != nil {
		t.Fatal(e)
	}
	if unselected.manifest.SourceDigest != baseline.manifest.SourceDigest {
		t.Fatal("unselected source entered Windows closure")
	}
	write("internal/runtime/broker/closurefixture/payload.go", []byte("package closurefixture\nimport _ \"embed\"\n//go:embed payload.bin\nvar Payload []byte\n"))
	write("internal/runtime/broker/closurefixture/payload.bin", []byte{0, 13, 10, 255, 42})
	write("internal/runtime/broker/closure_windows.go", []byte("package broker\nimport \""+modulePath+"/internal/runtime/broker/closurefixture\"\nvar selectedClosurePayload = closurefixture.Payload\n"))
	added, e := capture(fixture)
	if e != nil {
		t.Fatal(e)
	}
	key := "repo/internal/runtime/broker/closurefixture/payload.bin"
	if !bytes.Equal(added.files[key].bytes, []byte{0, 13, 10, 255, 42}) {
		t.Fatal("selected non-Go embed missing/normalized")
	}
	if added.manifest.SourceDigest == baseline.manifest.SourceDigest {
		t.Fatal("new dependency missed")
	}
	write("internal/runtime/broker/closurefixture/payload.bin", []byte{0, 13, 10, 255, 43})
	embedChanged, e := capture(fixture)
	if e != nil {
		t.Fatal(e)
	}
	if added.manifest.SourceDigest == embedChanged.manifest.SourceDigest {
		t.Fatal("stale selected embed missed")
	}
	write("internal/runtime/brokerassets/data.go", []byte("package brokerassets\n"))
	write("internal/runtime/broker/closure_windows.go", []byte("package broker\nimport _ \""+modulePath+"/internal/runtime/brokerassets\"\n"))
	if _, e := capture(fixture); e == nil || !strings.Contains(e.Error(), "recursive/parent helper dependency") {
		t.Fatalf("recursive assets rejection: %v", e)
	}
}
