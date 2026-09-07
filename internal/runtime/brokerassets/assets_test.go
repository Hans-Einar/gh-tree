package brokerassets

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

func fixture(t *testing.T, arch string) ([]byte, []byte) {
	t.Helper()
	m, e := os.ReadFile("manifest.json")
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile("broker-" + arch + ".gz")
	if e != nil {
		t.Fatal(e)
	}
	return m, z
}
func TestBothCommittedImages(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			m, z := fixture(t, arch)
			image, e := validate(m, z, arch)
			if e != nil {
				t.Fatal(e)
			}
			if image.Machine != machine(arch) || image.Protocol == 0 || image.SHA256 != hash(image.Bytes) {
				t.Fatal("wrong returned values")
			}
			image.Bytes[0] ^= 1
			again, e := validate(m, z, arch)
			if e != nil {
				t.Fatal(e)
			}
			if again.Bytes[0] == image.Bytes[0] {
				t.Fatal("returned bytes alias")
			}
		})
	}
}
func TestArchitectureSelection(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64", "386", "ARM64", "", "../arm64"} {
		image, e := Load(arch)
		available := runtime.GOOS == "windows" && (runtime.GOARCH == "386" && (arch == "amd64" || arch == "arm64") || runtime.GOARCH == "amd64" && arch == "arm64")
		if available {
			if e != nil {
				t.Fatal(e)
			}
			if len(image.Bytes) == 0 {
				t.Fatal("empty image")
			}
		} else if e == nil || len(image.Bytes) != 0 {
			t.Fatalf("unavailable %s accepted", arch)
		}
	}
}
func TestRejectManifestAndImageCorruption(t *testing.T) {
	m, z := fixture(t, "amd64")
	parsed, e := parseManifest(m)
	if e != nil {
		t.Fatal(e)
	}
	for name, change := range map[string]func(*manifest){"schema": func(m *manifest) { m.Schema++ }, "protocol": func(m *manifest) { m.Protocol = 0 }, "toolchain": func(m *manifest) { m.Toolchain = "go1.26.0" }, "builder": func(m *manifest) { m.Builder = "linux/amd64" }, "source": func(m *manifest) { m.Sources[0].SHA256 = strings.Repeat("0", 64) }, "modules": func(m *manifest) { m.Modules[0].Version = "changed" }, "options": func(m *manifest) { m.Options[0] = "CGO_ENABLED=1" }, "target-duplicate": func(m *manifest) { m.Targets[1] = m.Targets[0] }, "target-machine": func(m *manifest) { m.Targets[0].Machine = 0xaa64 }, "image-hash": func(m *manifest) { m.Targets[0].SHA256 = strings.Repeat("0", 64) }, "image-length": func(m *manifest) { m.Targets[0].Length-- }, "allocation-bound": func(m *manifest) { m.Targets[0].Length = maxImage + 1 }} {
		t.Run(name, func(t *testing.T) {
			copy, e := parseManifest(m)
			if e != nil {
				t.Fatal(e)
			}
			change(&copy)
			if _, e := validate(canonical(copy), z, "amd64"); e == nil {
				t.Fatal("corrupt manifest accepted")
			}
		})
	}
	for _, bad := range [][]byte{append(append([]byte{}, m...), m...), bytes.Replace(m, []byte("\"schema\": 1"), []byte("\"schema\": 1, \"schema\": 1"), 1), bytes.Replace(m, []byte("\"schema\": 1"), []byte("\"schema\": 1, \"unknown\": 1"), 1)} {
		if _, e := validate(bad, z, "amd64"); e == nil {
			t.Fatal("ambiguous manifest accepted")
		}
	}
	bad := append([]byte{}, z...)
	bad[len(bad)/2] ^= 1
	if _, e := validate(m, bad, "amd64"); e == nil {
		t.Fatal("compressed corruption")
	}
	if _, e := validate(m, z[:len(z)-1], "amd64"); e == nil {
		t.Fatal("compressed truncation")
	}
	// Update the matching outer hash so controls exercise deeper gzip/PE rules.
	check := func(name string, bad []byte) {
		t.Helper()
		copy := parsed
		copy.Targets = append([]target{}, parsed.Targets...)
		copy.Targets[0].CompressedSHA256 = hash(bad)
		copy.Targets[0].CompressedLength = len(bad)
		if _, e := validate(canonical(copy), bad, "amd64"); e == nil {
			t.Fatal(name)
		}
	}
	bad = append(append([]byte{}, z...), z...)
	check("concatenated members", bad)
	bad = append([]byte{}, z...)
	bad[4] = 1
	check("timestamp", bad)
	r, e := gzip.NewReader(bytes.NewReader(z))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := io.ReadAll(r)
	r.Close()
	if e != nil {
		t.Fatal(e)
	}
	var alternate bytes.Buffer
	w, _ := gzip.NewWriterLevel(&alternate, gzip.NoCompression)
	w.Write(raw)
	w.Close()
	check("noncanonical compression", alternate.Bytes())
	offset := int(binary.LittleEndian.Uint32(raw[60:]))
	raw[offset+4] = 0x64
	raw[offset+5] = 0xaa
	var corrupted bytes.Buffer
	w, _ = gzip.NewWriterLevel(&corrupted, gzip.BestCompression)
	w.Write(raw)
	w.Close()
	copy := parsed
	copy.Targets = append([]target{}, parsed.Targets...)
	copy.Targets[0].SHA256 = hash(raw)
	copy.Targets[0].CompressedSHA256 = hash(corrupted.Bytes())
	copy.Targets[0].CompressedLength = corrupted.Len()
	if _, e := validate(canonical(copy), corrupted.Bytes(), "amd64"); e == nil {
		t.Fatal("wrong actual PE machine accepted")
	}
}
