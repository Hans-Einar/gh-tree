package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func canonicalTest(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" || runtime.Version() != "go1.25.0" {
		t.Skip("canonical native Windows amd64 Go1.25.0 execution")
	}
}

func writeFixture(t *testing.T, root, name string, b []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
}

// Independent fixture checksum in the Go h1 directory format. The real Go
// downloader verifies these pins; changing a fixture pin must fail admission.
func fixtureSum(files map[string][]byte) string {
	names := []string{}
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		fmt.Fprintf(h, "%x  %s\n", sha256.Sum256(files[name]), name)
	}
	return "h1:" + base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func moduleFixture(t *testing.T) (root, cache, proxy string) {
	t.Helper()
	owned := t.TempDir()
	root, cache, proxy = filepath.Join(owned, "source"), filepath.Join(owned, "cache"), filepath.Join(owned, "proxy")
	var requires, sums strings.Builder
	for _, name := range []string{"selected", "unused"} {
		mod := "example.com/" + name
		modfile := []byte("module " + mod + "\n\ngo 1.25.0\n")
		files := map[string][]byte{mod + "@v1.0.0/go.mod": modfile, mod + "@v1.0.0/value.go": []byte("package " + name + "\nconst Value = \"RECORDED_MODULE_BYTES_111111\"\n")}
		var archive bytes.Buffer
		z := zip.NewWriter(&archive)
		for key, b := range files {
			w, err := z.Create(key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(b); err != nil {
				t.Fatal(err)
			}
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		for ext, b := range map[string][]byte{"mod": modfile, "info": []byte(`{"Version":"v1.0.0","Time":"2026-01-01T00:00:00Z"}`), "zip": archive.Bytes()} {
			writeFixture(t, proxy, mod+"/@v/v1.0.0."+ext, b)
		}
		fmt.Fprintf(&requires, "require %s v1.0.0\n", mod)
		fmt.Fprintf(&sums, "%s v1.0.0 %s\n%s v1.0.0/go.mod %s\n", mod, fixtureSum(files), mod, fixtureSum(map[string][]byte{"go.mod": modfile}))
	}
	writeFixture(t, root, "go.mod", []byte("module "+modulePath+"\n\ngo 1.25.0\n\n"+requires.String()))
	writeFixture(t, root, "go.sum", []byte(sums.String()))
	writeFixture(t, root, "internal/runtime/broker/protocol.go", []byte("package broker\nconst ProtocolVersion uint16 = 1\n"))
	writeFixture(t, root, "internal/runtime/broker/cmd/main_windows.go", []byte("package main\nimport (\"fmt\"; \"example.com/selected\"; \""+modulePath+"/internal/runtime/broker\")\nfunc main(){fmt.Print(selected.Value, broker.ProtocolVersion)}\n"))
	return
}

func TestPinnedPreparationFreshSelectedOnlyOfflineAndInconsistent(t *testing.T) {
	canonicalTest(t)
	root, cache, proxy := moduleFixture(t)
	u := url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(proxy)}
	t.Setenv("GOMODCACHE", cache)
	t.Setenv("GOPROXY", u.String())
	t.Setenv("GOSUMDB", "off") // All synthetic module content has preexisting h1 pins.
	before := map[string][]byte{}
	for _, name := range []string{"go.mod", "go.sum"} {
		before[name], _ = os.ReadFile(filepath.Join(root, name))
	}
	// Begin with only the actually selected module, omitting unrelated metadata.
	if err := prepareModules(root); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(cache, "example.com", "unused@v1.0.0"), filepath.Join(cache, "cache", "download", "example.com", "unused")} {
		// Every path is a literal child of this test's new cache.
		if err := filepath.WalkDir(path, func(p string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			return os.Chmod(p, 0700)
		}); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := admit(root); err != nil {
		t.Fatalf("selected-only cache: %v", err)
	}
	t.Setenv("GOPROXY", "off")
	if err := admit(root); err != nil {
		t.Fatalf("populated offline cache: %v", err)
	}
	for name, b := range before {
		after, _ := os.ReadFile(filepath.Join(root, name))
		if !bytes.Equal(b, after) {
			t.Fatalf("rewrote %s", name)
		}
	}
	bad := bytes.Replace(before["go.sum"], []byte("h1:"), []byte("h1:wrong"), 1)
	writeFixture(t, root, "go.sum", bad)
	if err := admit(root); err == nil {
		t.Fatal("accepted inconsistent pins")
	}
	after, _ := os.ReadFile(filepath.Join(root, "go.sum"))
	if !bytes.Equal(after, bad) {
		t.Fatal("rewrote inconsistent pins")
	}
}
