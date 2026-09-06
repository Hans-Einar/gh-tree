package main

import "strings"

type role string

const (
	domain       role = "domain"
	api          role = "api"
	ports        role = "ports"
	application  role = "application"
	usecases     role = "usecases"
	git          role = "git"
	github       role = "github"
	runtimeLayer role = "runtime"
	broker       role = "broker"
	brokerEntry  role = "broker-entry"
	assets       role = "broker-assets"
	helpergen    role = "helper-generator"
	discovery    role = "launchdiscovery"
	persistence  role = "persistence"
	state        role = "tuistate"
	viewmodel    role = "viewmodel"
	view         role = "tuiview"
	composition  role = "composition"
	host         role = "host"
	architecture role = "architecture-tool"
	entry        role = "entry"
	version      role = "version"
)

func within(path, root string) bool { return path == root || strings.HasPrefix(path, root+"/") }

// Longest/specialized roots come first. A private child retains its owner's
// restrictions; it does not gain the wider privileges of a neighboring package.
func classify(path string) role {
	for _, p := range []struct {
		path string
		role role
	}{
		{"internal/composition/architecture", architecture}, {"internal/composition/host", host},
		{"internal/composition", composition}, {"internal/application/api", api},
		{"internal/application/ports", ports}, {"internal/application/usecases", usecases},
		{"internal/application", application}, {"internal/tuistate/viewmodel", viewmodel},
		{"internal/tuistate", state}, {"internal/tuiview", view}, {"internal/domain", domain},
		{"internal/git", git}, {"internal/github/adapter", github},
		{"internal/runtime/broker/cmd", brokerEntry}, {"internal/runtime/broker", broker},
		{"internal/runtime/brokerassets", assets}, {"internal/runtime/cmd/helpergen", helpergen},
		{"internal/runtime", runtimeLayer}, {"internal/launchdiscovery", discovery},
		{"internal/persistence", persistence},
	} {
		if within(path, p.path) {
			return p.role
		}
	}
	if path == "cmd/gh-tree" {
		return entry
	}
	if path == "internal/version" {
		return version
	}
	return ""
}

func isAdapter(r role) bool {
	return r == git || r == github || r == runtimeLayer || r == discovery || r == persistence
}
func isPure(r role) bool {
	return r == domain || r == api || r == ports || r == state || r == viewmodel || r == view || r == version
}
func isTool(r role) bool { return r == architecture || r == helpergen }

func internalAllowed(from, to role, test bool) bool {
	if from == "" || to == "" {
		return false
	}
	// The only broad integration-test seam is Composition. Layer-local tests
	// cannot introduce cross-adapter or presentation/backend dependencies.
	if test && from == composition {
		return to != architecture && to != helpergen && to != brokerEntry && to != entry
	}
	if from == architecture {
		return to == architecture
	}
	if to == architecture || to == helpergen || to == brokerEntry || to == entry {
		return false
	}
	if from == to {
		return true
	}
	switch from {
	case entry:
		return to == composition || to == version
	case composition:
		return to == application || to == api || to == ports || to == domain || isAdapter(to) || to == host || to == version
	case host:
		return to == api || to == state || to == view || to == viewmodel
	case application:
		return to == api || to == ports || to == usecases || to == domain
	case usecases:
		return to == api || to == ports || to == domain
	case ports:
		return to == api || to == domain
	case api:
		return to == domain
	case state:
		return to == api || to == domain || to == viewmodel
	case view:
		return to == viewmodel
	case viewmodel:
		return to == domain
	case runtimeLayer:
		return to == ports || to == api || to == domain || to == broker || to == assets
	case broker, brokerEntry:
		return to == broker || to == api || to == domain
	case helpergen:
		return to == broker || to == api || to == domain
	}
	if isAdapter(from) {
		return to == ports || to == api || to == domain
	}
	return false
}

// No operational stdlib is admitted just because it is standard. These packages
// provide deterministic in-memory algorithms; fmt/time need symbol checks too.
var pureStandard = set("bytes", "cmp", "errors", "fmt", "math", "math/bits", "math/big", "regexp", "regexp/syntax", "slices", "sort", "strconv", "strings", "unicode", "unicode/utf8", "unicode/utf16", "encoding/binary", "encoding/hex", "encoding/base64")
var rendering = set("github.com/charmbracelet/lipgloss", "github.com/charmbracelet/x/ansi", "github.com/rivo/uniseg", "github.com/mattn/go-runewidth", "github.com/lucasb-eyer/go-colorful")
var nativeLibraries = []string{"golang.org/x/sys", "golang.org/x/text", "github.com/charmbracelet/x/xpty", "github.com/charmbracelet/x/conpty", "github.com/charmbracelet/x/term", "github.com/charmbracelet/x/termios", "github.com/creack/pty"}

func set(items ...string) map[string]bool {
	m := map[string]bool{}
	for _, s := range items {
		m[s] = true
	}
	return m
}

func externalAllowed(r role, path string, standard, test bool) bool {
	if test {
		return standard || rendering[path] || path == "github.com/charmbracelet/bubbletea" || nativeLibrary(path)
	}
	if r == architecture {
		return standard
	}
	if isPure(r) {
		if pureStandard[path] {
			return true
		}
		if path == "time" {
			return r != domain && r != version
		}
		if path == "context" {
			return r == api || r == ports
		}
		return r == view && rendering[path]
	}
	if r == assets {
		return standard && (pureStandard[path] || set("embed", "compress/gzip", "crypto/sha256", "io", "io/fs")[path])
	}
	if r == discovery && (path == "os/exec" || path == "net" || strings.HasPrefix(path, "net/") || path == "plugin") {
		return false
	}
	if standard {
		return path != "C"
	}
	if nativeLibrary(path) {
		return isAdapter(r) || r == broker || r == brokerEntry || r == helpergen || r == composition
	}
	if r == host || r == composition {
		return path == "github.com/charmbracelet/bubbletea" || rendering[path]
	}
	return false
}

func nativeLibrary(path string) bool {
	for _, p := range nativeLibraries {
		if within(path, p) {
			return true
		}
	}
	return false
}

// Cross-layer dependencies terminate at the published package root. A private
// subpackage can be decomposed freely by its layer owner but is not another API.
// Application's coordinator owns its usecase children; Runtime owns its private
// broker/asset graph. Their explicit edges do not authorize other adapters.
func privateImportAllowed(from role, path string) bool {
	to := classify(path)
	if from == to {
		return true
	}
	if from == application && to == usecases {
		return true
	}
	if (from == runtimeLayer || from == brokerEntry || from == helpergen) && (to == broker || to == assets) {
		return true
	}
	return set("internal/domain", "internal/application", "internal/application/api", "internal/application/ports", "internal/application/usecases", "internal/git", "internal/github/adapter", "internal/runtime", "internal/launchdiscovery", "internal/persistence", "internal/tuistate", "internal/tuistate/viewmodel", "internal/tuiview", "internal/composition", "internal/composition/host", "internal/version")[path]
}

// Imported named types in public boundary surfaces must be values/control types,
// never backend handles, readers or hidden renderer/process implementation DTOs.
func publicExternalType(path, name string, r role) bool {
	if path == "context" && name == "Context" {
		return r == api || r == ports || r == application || r == usecases || isAdapter(r)
	}
	if path == "time" {
		return r != domain && set("Time", "Duration", "Month", "Weekday")[name]
	}
	return r == view && rendering[path]
}
