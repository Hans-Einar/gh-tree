package main

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Regression for H70-H01: the live module AND owned compiler/source inputs are
// changed during BOTH clean builds and restored before recapture. Matching
// before/after manifests must never bless executable bytes from those changes.
func TestBuildConsumesCapturedModuleAndToolchain(t *testing.T) {
	canonicalTest(t)
	root, cache, proxy := moduleFixture(t)
	u := url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(proxy)}
	t.Setenv("GOMODCACHE", cache)
	t.Setenv("GOPROXY", u.String())
	t.Setenv("GOSUMDB", "off")
	if err := admit(root); err != nil {
		t.Fatal(err)
	}
	initial, err := capture(root)
	if err != nil {
		t.Fatal(err)
	}
	// Only the owned toolchain copy is changed. Never touch runtime.GOROOT's
	// shared installation or any other worker's fixture/cache.
	owned, err := materialize(initial, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(owned.root, "source"), "internal/runtime/broker/unrecorded_windows.go", []byte("package broker\nconst Unrecorded = 1\n"))
	if err := owned.verifySelection(initial, "amd64"); err == nil || !strings.Contains(err.Error(), "input directory invalidated") {
		owned.close()
		t.Fatalf("unrecorded selected snapshot file: %v", err)
	}
	owned.close() // This copy is deliberately the mutable external toolchain fixture.
	t.Setenv("GOROOT", filepath.Join(owned.root, "goroot"))
	p, err := capture(root)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{filepath.Join(cache, "example.com", "selected@v1.0.0", "value.go"), filepath.Join(owned.root, "goroot", "src", "internal", "goarch", "goarch.go"), filepath.Join(owned.root, "goroot", "pkg", "tool", "windows_amd64", "compile.exe"), filepath.Join(owned.root, "goroot", "bin", "go.exe")}
	original := map[string][]byte{}
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		original[path] = b
		if err := os.Chmod(path, 0600); err != nil {
			t.Fatal(err)
		}
		changed := []byte("UNRECORDED_INVALID_TOOLCHAIN_BYTES")
		if path == paths[0] {
			changed = bytes.ReplaceAll(b, []byte("RECORDED_MODULE_BYTES_111111"), []byte("UNRECORDED_MODULE_BYTES_222222"))
		}
		if err := os.WriteFile(path, changed, 0600); err != nil {
			t.Fatal(err)
		}
	}
	restore := func() {
		for path, b := range original {
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Errorf("restore owned input %s: %v", path, err)
			}
		}
	}
	t.Cleanup(restore)
	t.Setenv("GOPROXY", "off")
	first, err := build(p)
	if err != nil {
		t.Fatal(err)
	}
	second, err := build(p)
	if err != nil {
		t.Fatal(err)
	}
	restore()
	after, err := capture(root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(jsonBytes(p.manifest), jsonBytes(after.manifest)) {
		t.Fatal("control failed to restore before/after provenance")
	}
	for _, arch := range arches {
		if !bytes.Equal(first[arch], second[arch]) || !bytes.Contains(first[arch], []byte("RECORDED_MODULE_BYTES_111111")) || bytes.Contains(first[arch], []byte("UNRECORDED_MODULE_BYTES_222222")) {
			t.Fatalf("%s did not consume exclusively recorded input bytes", arch)
		}
		t.Logf("%s repeated image=%s; captured module bytes retained despite live module/compiler/source replacement", arch, hash(first[arch]))
	}
	// Verification cannot trust an already populated module cache/ziphash either.
	writeFixture(t, cache, "example.com/selected@v1.0.0/value.go", []byte("package selected\nconst Value=\"changed\"\n"))
	if _, err := capture(root); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("captured inconsistent selected source: %v", err)
	}
}

func TestAssemblyClosureRejectsExternalInputs(t *testing.T) {
	for _, text := range []string{"#include \"/outside.h\"", "#include \"../outside.h\"", "#include \"C:/outside.h\"", "#include NAME", "#include \"sub/../../outside.h\""} {
		if _, err := assemblyIncludes([]byte(text)); err == nil {
			t.Fatalf("accepted %s", text)
		}
	}
	names, err := assemblyIncludes([]byte("// #include \"/ignored.h\"\n#include \"nested/input.h\"\n#include \"go_asm.h\"\n"))
	if err != nil || len(names) != 2 || names[0] != "nested/input.h" || names[1] != "go_asm.h" {
		t.Fatalf("literal local/generated includes: %v %v", names, err)
	}
}

func TestSnapshotRejectsUnrecordedOrChangedBytes(t *testing.T) {
	b := []byte("recorded")
	f := captured{source: source{"repo/go.mod", hash(b), len(b)}, bytes: b, repoPath: "go.mod"}
	p := plan{files: map[string]captured{f.Path: f}, manifest: manifest{Sources: []source{f.source}}}
	p.manifest.SourceDigest = hash(jsonBytes(p.manifest.Sources))
	p.manifest.OptionsDigest = hash(jsonBytes(p.manifest.Options))
	p.manifest.ModuleDigest = hash(jsonBytes(p.manifest.Modules))
	baseline, err := materialize(p, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	baseline.close()
	f.bytes = []byte("unrecorded")
	p.files[f.Path] = f
	if _, err := materialize(p, t.TempDir()); err == nil {
		t.Fatal("accepted changed captured bytes")
	}
	f.bytes = b
	p.files[f.Path] = f
	p.files["repo/extra.go"] = f
	if _, err := materialize(p, t.TempDir()); err == nil {
		t.Fatal("accepted unrecorded input")
	}
}
