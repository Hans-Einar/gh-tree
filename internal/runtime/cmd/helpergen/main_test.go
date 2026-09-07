package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProtocolVersionUsesSource(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   uint16
	}{{"package broker; const ProtocolVersion uint16 = 1", 1}, {"package broker; const ProtocolVersion uint16 = 37", 37}, {"package broker; const Other = 1", 0}, {"package broker; const ProtocolVersion = 0", 0}, {"package broker; const ProtocolVersion = 65536", 0}, {"package broker; var ProtocolVersion = 1", 0}, {"package broker; const ProtocolVersion = 1+1", 0}} {
		got, e := protocolVersion([]byte(tc.source))
		if (e != nil) != (tc.want == 0) || got != tc.want {
			t.Fatalf("%s: %d %v", tc.source, got, e)
		}
	}
}
func TestDeterministicCompressionAndPE(t *testing.T) {
	b := bytes.Repeat([]byte("complete-image\n"), 4096)
	a, e := compress(b)
	if e != nil {
		t.Fatal(e)
	}
	c, e := compress(b)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(a, c) || !bytes.Equal(a[:10], []byte{31, 139, 8, 0, 0, 0, 0, 0, 2, 255}) {
		t.Fatal("noncanonical gzip")
	}
	r, e := gzip.NewReader(bytes.NewReader(a))
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := io.ReadAll(r)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(decoded, b) {
		t.Fatal("round trip")
	}
	r.Close()
	for _, arch := range arches {
		if e := validatePE(b, arch); e == nil {
			t.Fatal("accepted non-PE")
		}
		image, e := os.ReadFile(filepath.Join("..", "..", "brokerassets", "broker-"+arch+".gz"))
		if e != nil {
			t.Fatal(e)
		}
		r, e := gzip.NewReader(bytes.NewReader(image))
		if e != nil {
			t.Fatal(e)
		}
		raw, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		if e := validatePE(raw, arch); e != nil {
			t.Fatal(e)
		}
		if e := validatePE(raw, map[string]string{"amd64": "arm64", "arm64": "amd64"}[arch]); e == nil {
			t.Fatal("accepted wrong machine")
		}
		offset := int(binary.LittleEndian.Uint32(raw[60:]))
		raw[offset+4] = 0
		raw[offset+5] = 0
		if e := validatePE(raw, arch); e == nil {
			t.Fatal("accepted corrupted PE")
		}
	}
}
func TestVerifyNoRewriteAndCorruption(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, assetDir)
	if e := os.MkdirAll(dir, 0700); e != nil {
		t.Fatal(e)
	}
	want := map[string][]byte{"broker-amd64.gz": []byte("amd64"), "broker-arm64.gz": []byte("arm64"), "manifest.json": []byte("manifest")}
	for name, b := range want {
		if e := os.WriteFile(filepath.Join(dir, name), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	if e := verifyOutputs(root, want); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"broker-amd64.gz", "broker-arm64.gz", "manifest.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			original := want[name]
			corrupt := append(append([]byte{}, original...), 0x42)
			if e := os.WriteFile(path, corrupt, 0600); e != nil {
				t.Fatal(e)
			}
			before, e := os.Stat(path)
			if e != nil {
				t.Fatal(e)
			}
			if e := verifyOutputs(root, want); e == nil || !strings.Contains(e.Error(), "stale or corrupt "+name) {
				t.Fatalf("wrong mismatch: %v", e)
			}
			after, e := os.Stat(path)
			if e != nil {
				t.Fatal(e)
			}
			got, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(got, corrupt) || !after.ModTime().Equal(before.ModTime()) {
				t.Fatal("check rewrote corrupt input")
			}
			os.WriteFile(path, original, 0600)
		})
	}
	if e := os.Remove(filepath.Join(dir, "manifest.json")); e != nil {
		t.Fatal(e)
	}
	if e := verifyOutputs(root, want); e == nil {
		t.Fatal("missing manifest accepted")
	}
	if _, e := os.Stat(filepath.Join(dir, "manifest.json")); !os.IsNotExist(e) {
		t.Fatal("check recreated missing manifest")
	}
}
func TestNormalizedClosureAndContainment(t *testing.T) {
	if !bytes.Equal(normalize([]byte("a\r\nb\r\n")), []byte("a\nb\n")) {
		t.Fatal("CRLF")
	}
	a := []source{{"z", hash([]byte("z")), 1}, {"a", hash([]byte("a")), 1}}
	sortSources(a)
	first := hash(jsonBytes(a))
	a[0].SHA256 = hash([]byte("changed source"))
	if hash(jsonBytes(a)) == first {
		t.Fatal("stale source digest")
	}
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.go")
	if _, e := contained(root, outside); e == nil {
		t.Fatal("outside input")
	}
}
func TestPinnedEnvironmentOverridesAmbient(t *testing.T) {
	for _, key := range []string{"GOFLAGS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOWORK", "GOEXPERIMENT", "GOTOOLCHAIN", "GOCACHEPROG"} {
		t.Setenv(key, "untrusted")
	}
	env := environment("arm64", "fresh-cache")
	values := map[string]string{}
	for _, line := range env {
		k, v, _ := strings.Cut(line, "=")
		if _, ok := values[strings.ToUpper(k)]; ok {
			t.Fatalf("duplicate env %s", k)
		}
		values[strings.ToUpper(k)] = v
	}
	for k, v := range map[string]string{"GOARCH": "arm64", "GOAMD64": "v1", "GOARM64": "v8.0", "CGO_ENABLED": "0", "GOFLAGS": "", "GOWORK": "off", "GOEXPERIMENT": "", "GOTOOLCHAIN": "local", "GOCACHE": "fresh-cache", "GOCACHEPROG": ""} {
		if values[k] != v {
			t.Fatalf("%s=%s", k, values[k])
		}
	}
}
