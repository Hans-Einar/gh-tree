// Package brokerassets contains only committed, in-memory Runtime helper inputs.
// The parent Runtime passes copied values to its Windows extraction owner.
package brokerassets

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const maxImage = 64 << 20

// Image owns Bytes. No returned slice aliases embedded data or another result.
// This is a Runtime-private value; it is never exposed by the Sessions API.
type Image struct {
	Bytes    []byte
	SHA256   string
	Machine  uint16
	Protocol uint16
}

type source struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Length int    `json:"length"`
}
type module struct {
	Path     string
	Version  string
	Sum      string
	GoModSum string
	Main     bool
	Replace  *module `json:",omitempty"`
}
type target struct {
	Arch             string `json:"arch"`
	Machine          uint16 `json:"machine"`
	Length           int    `json:"length"`
	SHA256           string `json:"sha256"`
	CompressedLength int    `json:"compressedLength"`
	CompressedSHA256 string `json:"compressedSHA256"`
	SourceDigest     string `json:"sourceDigest"`
}
type manifest struct {
	Schema          int      `json:"schema"`
	Protocol        uint16   `json:"protocol"`
	Toolchain       string   `json:"toolchain"`
	Builder         string   `json:"builder"`
	ToolchainDigest string   `json:"toolchainDigest"`
	ModuleDigest    string   `json:"moduleDigest"`
	Options         []string `json:"options"`
	OptionsDigest   string   `json:"optionsDigest"`
	SourceDigest    string   `json:"sourceDigest"`
	Sources         []source `json:"sources"`
	Modules         []module `json:"modules"`
	Targets         []target `json:"targets"`
}

// Load selects only images embedded for the current product architecture.
// Native routing, extraction and execution belong to the parent/native owners.
func Load(arch string) (Image, error) {
	z := payload(arch)
	if len(z) == 0 {
		return Image{}, fmt.Errorf("native helper asset unavailable for %q", arch)
	}
	return validate(manifestBytes, z, arch)
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func canonical(v any) []byte {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return nil
	}
	return append(b, '\n')
}
func validHash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && strings.ToLower(s) == s
}
func machine(arch string) uint16 {
	if arch == "amd64" {
		return pe.IMAGE_FILE_MACHINE_AMD64
	}
	if arch == "arm64" {
		return pe.IMAGE_FILE_MACHINE_ARM64
	}
	return 0
}

func parseManifest(b []byte) (manifest, error) {
	var m manifest
	if len(b) == 0 || len(b) > 2<<20 {
		return m, fmt.Errorf("invalid manifest length")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(&m); e != nil {
		return m, e
	}
	// A unique canonical representation rejects duplicate keys/trailing values.
	if !bytes.Equal(b, canonical(m)) {
		return m, fmt.Errorf("noncanonical manifest")
	}
	if m.Schema != 1 || m.Protocol == 0 || m.Toolchain != "go1.25.0" || m.Builder != "windows/amd64" || len(m.Targets) != 2 || len(m.Sources) == 0 || len(m.Modules) == 0 || len(m.Options) == 0 {
		return m, fmt.Errorf("unsupported/incomplete manifest")
	}
	if m.SourceDigest != hash(canonical(m.Sources)) || m.ModuleDigest != hash(canonical(m.Modules)) || m.OptionsDigest != hash(canonical(m.Options)) {
		return m, fmt.Errorf("manifest provenance digest mismatch")
	}
	previous := ""
	var tc []source
	for _, s := range m.Sources {
		if s.Path <= previous || s.Length < 0 || !validHash(s.SHA256) {
			return m, fmt.Errorf("invalid sorted source closure")
		}
		previous = s.Path
		if strings.HasPrefix(s.Path, "toolchain/") {
			tc = append(tc, s)
		}
	}
	if len(tc) == 0 || m.ToolchainDigest != hash(canonical(tc)) {
		return m, fmt.Errorf("toolchain digest mismatch")
	}
	previous = ""
	mains := 0
	for _, v := range m.Modules {
		if v.Path <= previous || v.Replace != nil {
			return m, fmt.Errorf("invalid module provenance")
		}
		previous = v.Path
		if v.Main {
			mains++
		} else if v.Version == "" || !strings.HasPrefix(v.Sum, "h1:") || !strings.HasPrefix(v.GoModSum, "h1:") {
			return m, fmt.Errorf("unpinned module")
		}
	}
	if mains != 1 {
		return m, fmt.Errorf("invalid main module count")
	}
	for i, t := range m.Targets {
		expected := []string{"amd64", "arm64"}[i]
		if t.Arch != expected || t.Machine != machine(expected) || t.Length < 1 || t.Length > maxImage || t.CompressedLength < 1 || t.CompressedLength > maxImage || !validHash(t.SHA256) || !validHash(t.CompressedSHA256) || !validHash(t.SourceDigest) {
			return m, fmt.Errorf("invalid target manifest")
		}
	}
	return m, nil
}
func validate(manifestData, z []byte, arch string) (Image, error) {
	if machine(arch) == 0 {
		return Image{}, fmt.Errorf("unsupported native helper architecture")
	}
	m, e := parseManifest(manifestData)
	if e != nil {
		return Image{}, e
	}
	var selected target
	for _, t := range m.Targets {
		if t.Arch == arch {
			selected = t
		}
	}
	if len(z) != selected.CompressedLength || hash(z) != selected.CompressedSHA256 {
		return Image{}, fmt.Errorf("compressed image integrity mismatch")
	}
	raw := bytes.NewReader(z)
	r, e := gzip.NewReader(raw)
	if e != nil {
		return Image{}, e
	}
	r.Multistream(false)
	if !r.ModTime.IsZero() || r.Name != "" || r.Comment != "" || len(r.Extra) != 0 || r.OS != 255 {
		r.Close()
		return Image{}, fmt.Errorf("noncanonical gzip header")
	}
	b, e := io.ReadAll(io.LimitReader(r, int64(selected.Length)+1))
	closeErr := r.Close()
	if e != nil {
		return Image{}, e
	}
	if closeErr != nil {
		return Image{}, closeErr
	}
	if len(b) != selected.Length || raw.Len() != 0 || hash(b) != selected.SHA256 {
		return Image{}, fmt.Errorf("image integrity/length mismatch")
	}
	f, e := pe.NewFile(bytes.NewReader(b))
	if e != nil {
		return Image{}, e
	}
	defer f.Close()
	if f.Machine != selected.Machine || f.Characteristics&pe.IMAGE_FILE_EXECUTABLE_IMAGE == 0 || f.Characteristics&pe.IMAGE_FILE_DLL != 0 {
		return Image{}, fmt.Errorf("wrong PE image machine/type")
	}
	if _, ok := f.OptionalHeader.(*pe.OptionalHeader64); !ok {
		return Image{}, fmt.Errorf("expected PE32+")
	}
	var canonicalGzip bytes.Buffer
	w, e := gzip.NewWriterLevel(&canonicalGzip, gzip.BestCompression)
	if e != nil {
		return Image{}, e
	}
	w.Header.OS = 255
	if _, e = w.Write(b); e != nil {
		return Image{}, e
	}
	if e = w.Close(); e != nil {
		return Image{}, e
	}
	if !bytes.Equal(z, canonicalGzip.Bytes()) {
		return Image{}, fmt.Errorf("noncanonical gzip bytes")
	}
	return Image{Bytes: b, SHA256: selected.SHA256, Machine: selected.Machine, Protocol: m.Protocol}, nil
}
