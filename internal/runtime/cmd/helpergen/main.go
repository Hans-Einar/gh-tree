// helpergen builds committed Runtime inputs. It is never invoked by the product.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const modulePath = "github.com/Hans-Einar/gh-tree"
const entry = "./internal/runtime/broker/cmd"
const assetDir = "internal/runtime/brokerassets"

var arches = []string{"amd64", "arm64"}
var buildFlags = []string{"-mod=readonly", "-trimpath", "-buildvcs=false", "-ldflags=-buildid="}

type source struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Length int    `json:"length"`
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
type module struct {
	Path     string
	Version  string
	Sum      string
	GoModSum string
	Main     bool
	Replace  *module `json:",omitempty"`
	Dir      string  `json:"-"`
	GoMod    string  `json:"-"`
}
type listedModule struct {
	Path, Version, Sum, GoModSum, Dir, GoMod string
	Main                                     bool
	Replace                                  *listedModule
}
type pkg struct {
	Dir, ImportPath                                                                                                     string
	Standard                                                                                                            bool
	Module                                                                                                              *listedModule
	GoFiles, CgoFiles, CFiles, CXXFiles, MFiles, HFiles, FFiles, SFiles, SwigFiles, SwigCXXFiles, SysoFiles, EmbedFiles []string
}
type captured struct {
	source
	bytes    []byte
	repoPath string
}
type plan struct {
	manifest      manifest
	files         map[string]captured
	targetSources map[string][]source
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func jsonBytes(v any) []byte {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	return append(b, '\n')
}
func normalize(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }

func main() {
	check := flag.Bool("check", false, "independently rebuild and compare without changing the checkout")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "helpergen accepts only -check")
		os.Exit(1)
	}
	root, e := os.Getwd()
	if e == nil {
		e = run(root, *check)
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, "helpergen:", e)
		os.Exit(1)
	}
}

func environment(arch, cache string) []string {
	values := map[string]string{"GOROOT": runtime.GOROOT(), "GOCACHEPROG": "", "GODEBUG": "", "GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOFLAGS": "", "CGO_ENABLED": "0", "GOOS": "windows", "GOARCH": arch, "GOAMD64": "v1", "GOARM64": "v8.0", "GO386": "sse2", "GOARM": "7", "GOEXPERIMENT": "", "GOFIPS140": "off", "GOPROXY": "off", "GOSUMDB": "off", "GOTELEMETRY": "off"}
	if cache != "" {
		values["GOCACHE"] = cache
	}
	out := []string{}
	for _, v := range os.Environ() {
		k, _, _ := strings.Cut(v, "=")
		if _, ok := values[strings.ToUpper(k)]; !ok {
			out = append(out, v)
		}
	}
	for k, v := range values {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}
func goCommand(root, arch, cache string, args ...string) ([]byte, error) {
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go.exe"), args...)
	cmd.Dir = root
	cmd.Env = environment(arch, cache)
	b, e := cmd.CombinedOutput()
	if e != nil {
		return nil, fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), e, b)
	}
	return b, nil
}
func admit(root string) error {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" || runtime.Version() != "go1.25.0" {
		return fmt.Errorf("canonical builder requires native windows/amd64 go1.25.0, got %s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	}
	if e := nativeHost(); e != nil {
		return e
	}
	b, e := goCommand(root, "amd64", "", "env", "GOVERSION", "GOHOSTOS", "GOHOSTARCH")
	if e != nil {
		return e
	}
	if strings.Join(strings.Fields(string(b)), " ") != "go1.25.0 windows amd64" {
		return fmt.Errorf("wrong Go executable: %s", b)
	}
	_, e = goCommand(root, "amd64", "", "mod", "verify")
	return e
}

func capture(root string) (plan, error) {
	p := plan{files: map[string]captured{}, targetSources: map[string][]source{}}
	p.manifest = manifest{Schema: 1, Toolchain: "go1.25.0", Builder: "windows/amd64", Options: append([]string{"CGO_ENABLED=0", "GOOS=windows", "GOAMD64=v1", "GOARM64=v8.0", "GO386=sse2", "GOARM=7", "GOEXPERIMENT=", "GOFIPS140=off", "GOWORK=off", "GOTOOLCHAIN=local", "gzip=best-compression,mtime=0,os=255,name=,comment="}, buildFlags...)}
	p.manifest.OptionsDigest = hash(jsonBytes(p.manifest.Options))
	add := func(key, path, repoPath string, text bool) error {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if text {
			b = normalize(b)
		}
		c := captured{source{key, hash(b), len(b)}, b, repoPath}
		if old, ok := p.files[key]; ok && old.SHA256 != c.SHA256 {
			return fmt.Errorf("input changed during capture: %s", key)
		}
		p.files[key] = c
		return nil
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		if e := add("repo/"+name, filepath.Join(root, name), name, true); e != nil {
			return p, e
		}
	}
	selectedModules := map[string]bool{modulePath: true}
	for _, arch := range arches {
		b, e := goCommand(root, arch, "", "list", "-mod=readonly", "-deps", "-json", entry)
		if e != nil {
			return p, e
		}
		dec := json.NewDecoder(bytes.NewReader(b))
		selected := map[string]bool{"repo/go.mod": true, "repo/go.sum": true}
		for {
			var q pkg
			e = dec.Decode(&q)
			if e == io.EOF {
				break
			}
			if e != nil {
				return p, e
			}
			if q.Module != nil {
				selectedModules[q.Module.Path] = true
			}
			if q.ImportPath == modulePath+"/internal/runtime" || strings.HasPrefix(q.ImportPath, modulePath+"/"+assetDir) || strings.HasPrefix(q.ImportPath, modulePath+"/internal/runtime/cmd/helpergen") {
				return p, fmt.Errorf("recursive/parent helper dependency: %s", q.ImportPath)
			}
			if len(q.CgoFiles)+len(q.CFiles)+len(q.CXXFiles)+len(q.MFiles)+len(q.FFiles)+len(q.SwigFiles)+len(q.SwigCXXFiles) != 0 {
				return p, fmt.Errorf("unexpected native compiler input: %s", q.ImportPath)
			}
			groups := [][]string{q.GoFiles, q.HFiles, q.SFiles, q.SysoFiles, q.EmbedFiles}
			for group, files := range groups {
				for _, name := range files {
					path := filepath.Join(q.Dir, name)
					var key, repoPath string
					if q.Standard {
						rel, e := contained(filepath.Join(runtime.GOROOT(), "src"), path)
						if e != nil {
							return p, e
						}
						key = "toolchain/src/" + rel
					} else if q.Module != nil && q.Module.Main {
						rel, e := contained(root, path)
						if e != nil {
							return p, e
						}
						key = "repo/" + rel
						repoPath = rel
					} else if q.Module != nil && q.Module.Replace == nil && q.Module.Version != "" {
						rel, e := contained(q.Module.Dir, path)
						if e != nil {
							return p, e
						}
						key = "module/" + q.Module.Path + "@" + q.Module.Version + "/" + rel
					} else {
						return p, fmt.Errorf("unattributed input %s", path)
					}
					if e := add(key, path, repoPath, repoPath != "" && group < 3); e != nil {
						return p, e
					}
					selected[key] = true
				}
			}
		}
		for key := range selected {
			p.targetSources[arch] = append(p.targetSources[arch], p.files[key].source)
		}
		sortSources(p.targetSources[arch])
	}
	moduleNames := []string{}
	for name := range selectedModules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)
	args := append([]string{"list", "-mod=readonly", "-m", "-json"}, moduleNames...)
	mods, e := goCommand(root, "amd64", "", args...)
	if e != nil {
		return p, e
	}
	d := json.NewDecoder(bytes.NewReader(mods))
	for {
		var m listedModule
		e = d.Decode(&m)
		if e == io.EOF {
			break
		}
		if e != nil {
			return p, e
		}
		if m.Replace != nil || (!m.Main && (m.Version == "" || m.Sum == "" || m.GoModSum == "")) {
			return p, fmt.Errorf("unpinned/replaced module %s", m.Path)
		}
		if m.Main && m.Path != modulePath {
			return p, fmt.Errorf("wrong module %s", m.Path)
		}
		p.manifest.Modules = append(p.manifest.Modules, module{Path: m.Path, Version: m.Version, Sum: m.Sum, GoModSum: m.GoModSum, Main: m.Main})
		if !m.Main {
			if e := add("module/"+m.Path+"@"+m.Version+"/go.mod", m.GoMod, "", false); e != nil {
				return p, e
			}
		}
	}
	p.manifest.ModuleDigest = hash(jsonBytes(p.manifest.Modules))
	version, e := protocolVersion(p.files["repo/internal/runtime/broker/protocol.go"].bytes)
	if e != nil {
		return p, e
	}
	p.manifest.Protocol = version
	// The recipe is provenance, never part of a helper's executable dependency graph.
	recipe, e := filepath.Glob(filepath.Join(root, "internal/runtime/cmd/helpergen/*.go"))
	if e != nil {
		return p, e
	}
	for _, path := range recipe {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		rel, e := contained(root, path)
		if e != nil {
			return p, e
		}
		if e := add("recipe/"+rel, path, "", true); e != nil {
			return p, e
		}
	}
	// Bind the actual canonical compiler/linker/assembler, not just a version label.
	for _, name := range []string{"VERSION", "bin/go.exe", "pkg/tool/windows_amd64/compile.exe", "pkg/tool/windows_amd64/link.exe", "pkg/tool/windows_amd64/asm.exe", "pkg/include/asm_amd64.h", "pkg/include/asm_ppc64x.h", "pkg/include/funcdata.h", "pkg/include/textflag.h"} {
		if e := add("toolchain/"+name, filepath.Join(runtime.GOROOT(), name), "", false); e != nil {
			return p, e
		}
	}
	var tc []source
	for _, f := range p.files {
		p.manifest.Sources = append(p.manifest.Sources, f.source)
		if strings.HasPrefix(f.Path, "toolchain/") {
			tc = append(tc, f.source)
		}
	}
	sortSources(p.manifest.Sources)
	sortSources(tc)
	p.manifest.SourceDigest = hash(jsonBytes(p.manifest.Sources))
	p.manifest.ToolchainDigest = hash(jsonBytes(tc))
	return p, nil
}
func protocolVersion(b []byte) (uint16, error) {
	f, e := parser.ParseFile(token.NewFileSet(), "protocol.go", b, 0)
	if e != nil {
		return 0, e
	}
	for _, decl := range f.Decls {
		g, ok := decl.(*ast.GenDecl)
		if !ok || g.Tok != token.CONST {
			continue
		}
		for _, spec := range g.Specs {
			s := spec.(*ast.ValueSpec)
			for i, name := range s.Names {
				if name.Name != "ProtocolVersion" {
					continue
				}
				if i >= len(s.Values) {
					return 0, fmt.Errorf("protocol version requires explicit integer")
				}
				literal, ok := s.Values[i].(*ast.BasicLit)
				if !ok || literal.Kind != token.INT {
					return 0, fmt.Errorf("protocol version requires explicit integer")
				}
				n, e := strconv.ParseUint(literal.Value, 0, 16)
				if e != nil || n == 0 {
					return 0, fmt.Errorf("invalid protocol version")
				}
				return uint16(n), nil
			}
		}
	}
	return 0, fmt.Errorf("missing protocol version")
}
func sortSources(s []source) { sort.Slice(s, func(i, j int) bool { return s[i].Path < s[j].Path }) }
func contained(root, path string) (string, error) {
	r, e := filepath.Rel(root, path)
	if e != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) || filepath.IsAbs(r) {
		return "", fmt.Errorf("input outside declared root: %s", path)
	}
	physical, e := physicalPath(path)
	if e != nil {
		return "", fmt.Errorf("resolve selected input %q: %w", path, e)
	}
	base, e := physicalPath(root)
	if e != nil {
		return "", fmt.Errorf("resolve declared root %q for %q: %w", root, path, e)
	}
	rel, e := filepath.Rel(base, physical)
	if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("redirected input outside root: %s", path)
	}
	return filepath.ToSlash(r), nil
}

func build(p plan) (map[string][]byte, error) {
	tmp, e := os.MkdirTemp("", "gh-tree-helpergen-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(tmp)
	root := filepath.Join(tmp, "source")
	for _, f := range p.files {
		if f.repoPath == "" {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(f.repoPath))
		if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			return nil, e
		}
		if e := os.WriteFile(path, f.bytes, 0600); e != nil {
			return nil, e
		}
	}
	images := map[string][]byte{}
	for _, arch := range arches {
		out := filepath.Join(tmp, arch+".exe")
		args := append([]string{"build"}, buildFlags...)
		args = append(args, "-o", out, entry)
		if _, e := goCommand(root, arch, filepath.Join(tmp, "cache"), args...); e != nil {
			return nil, e
		}
		b, e := os.ReadFile(out)
		if e != nil {
			return nil, e
		}
		if e := validatePE(b, arch); e != nil {
			return nil, e
		}
		images[arch] = b
	}
	return images, nil
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
func validatePE(b []byte, arch string) error {
	f, e := pe.NewFile(bytes.NewReader(b))
	if e != nil {
		return e
	}
	defer f.Close()
	if f.Machine != machine(arch) || f.Machine == 0 {
		return fmt.Errorf("wrong PE machine for %s", arch)
	}
	if _, ok := f.OptionalHeader.(*pe.OptionalHeader64); !ok {
		return fmt.Errorf("expected PE32+ %s", arch)
	}
	if f.Characteristics&pe.IMAGE_FILE_EXECUTABLE_IMAGE == 0 || f.Characteristics&pe.IMAGE_FILE_DLL != 0 {
		return fmt.Errorf("not an executable image")
	}
	return nil
}
func compress(b []byte) ([]byte, error) {
	var out bytes.Buffer
	w, e := gzip.NewWriterLevel(&out, gzip.BestCompression)
	if e != nil {
		return nil, e
	}
	w.Header.OS = 255
	if _, e = w.Write(b); e != nil {
		return nil, e
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
func outputs(p plan, images map[string][]byte) (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, arch := range arches {
		b := images[arch]
		z, e := compress(b)
		if e != nil {
			return nil, e
		}
		out["broker-"+arch+".gz"] = z
		p.manifest.Targets = append(p.manifest.Targets, target{arch, machine(arch), len(b), hash(b), len(z), hash(z), hash(jsonBytes(p.targetSources[arch]))})
	}
	out["manifest.json"] = jsonBytes(p.manifest)
	return out, nil
}
func verifyOutputs(root string, want map[string][]byte) error {
	for _, name := range []string{"broker-amd64.gz", "broker-arm64.gz", "manifest.json"} {
		got, e := os.ReadFile(filepath.Join(root, assetDir, name))
		if e != nil {
			return e
		}
		if !bytes.Equal(got, want[name]) {
			return fmt.Errorf("stale or corrupt %s: expected SHA256 %s, got %s", name, hash(want[name]), hash(got))
		}
	}
	return nil
}
func run(root string, check bool) error {
	if e := admit(root); e != nil {
		return e
	}
	p, e := capture(root)
	if e != nil {
		return e
	}
	first, e := build(p)
	if e != nil {
		return e
	}
	second, e := build(p)
	if e != nil {
		return e
	}
	for _, arch := range arches {
		if !bytes.Equal(first[arch], second[arch]) {
			return fmt.Errorf("independent builds differ for %s", arch)
		}
	}
	after, e := capture(root)
	if e != nil {
		return e
	}
	if !bytes.Equal(jsonBytes(p.manifest), jsonBytes(after.manifest)) {
		return fmt.Errorf("source closure changed during independent builds")
	}
	want, e := outputs(p, first)
	if e != nil {
		return e
	}
	if check {
		if e := verifyOutputs(root, want); e != nil {
			return e
		}
	} else {
		if e := os.MkdirAll(filepath.Join(root, assetDir), 0755); e != nil {
			return e
		}
		for _, name := range []string{"broker-amd64.gz", "broker-arm64.gz", "manifest.json"} {
			if e := os.WriteFile(filepath.Join(root, assetDir, name), want[name], 0644); e != nil {
				return e
			}
		}
	}
	fmt.Printf("helpers verified: canonical windows/amd64 go1.25.0; two clean builds per target; source closure %s (%d inputs)\n", p.manifest.SourceDigest, len(p.manifest.Sources))
	return nil
}
