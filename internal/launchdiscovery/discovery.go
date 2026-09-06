package launchdiscovery

import (
	"container/heap"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"io"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

// Limits are immutable construction policy. Smaller values support deliberately
// bounded installations and tests; larger values cannot bypass the frozen caps.
type Limits struct{ Depth, Directories, Candidates, ManifestBytes, MakeLineBytes int }

func DefaultLimits() Limits { return Limits{5, 10000, 10000, 4 << 20, 1 << 20} }

// Config registers the closed built-in providers. Nil selects npm and Make.
// Duplicate or unknown keys refuse construction, before filesystem observation.
type Config struct {
	Limits    Limits
	Providers []api.ProviderKind
}
type Adapter struct {
	limits       Limits
	providers    []api.ProviderKind
	issuer       string
	observations atomic.Uint64
}

var _ ports.LaunchDiscovery = (*Adapter)(nil)

func New(config Config) (*Adapter, error) {
	limits := config.Limits
	if limits == (Limits{}) {
		limits = DefaultLimits()
	}
	max := DefaultLimits()
	if limits.Depth < 0 || limits.Depth > max.Depth || limits.Directories < 1 || limits.Directories > max.Directories || limits.Candidates < 1 || limits.Candidates > max.Candidates || limits.ManifestBytes < 1 || limits.ManifestBytes > max.ManifestBytes || limits.MakeLineBytes < 1 || limits.MakeLineBytes > max.MakeLineBytes {
		return nil, errors.New("invalid discovery construction limits")
	}
	providers := append([]api.ProviderKind(nil), config.Providers...)
	if config.Providers == nil {
		providers = []api.ProviderKind{api.Npm, api.Make}
	}
	seen := map[api.ProviderKind]bool{}
	for _, p := range providers {
		if !p.Valid() || seen[p] {
			return nil, errors.New("invalid or duplicate provider registration")
		}
		seen[p] = true
	}
	if len(providers) == 0 {
		return nil, errors.New("empty provider registry")
	}
	sort.Slice(providers, func(i, j int) bool { return providerKey(providers[i]) < providerKey(providers[j]) })
	var nonce [16]byte
	if _, e := rand.Read(nonce[:]); e != nil {
		return nil, e
	}
	return &Adapter{limits: limits, providers: providers, issuer: hex.EncodeToString(nonce[:])}, nil
}
func providerKey(p api.ProviderKind) string {
	if p == api.Npm {
		return "npm"
	}
	return "make"
}
func providerProfile(p api.ProviderKind) string {
	if p == api.Npm {
		return npmProfile
	}
	return makeProfile
}
func (d *Adapter) version(scope string, parts ...any) api.SourceVersion {
	h := sha256.New()
	for _, p := range parts {
		s := fmt.Sprintf("%#v", p)
		fmt.Fprintf(h, "%d:%s", len(s), s)
	}
	v, e := api.NewSourceVersion("launchdiscovery-v1", scope, d.issuer, hex.EncodeToString(h.Sum(nil)))
	if e != nil {
		panic(e)
	}
	return v
}
func diagnostic(code api.ErrorCode, reason string) api.Diagnostic {
	v, e := api.NewDiagnostic(api.DiagnosticData{Code: code, Reason: reason, Message: reason})
	if e != nil {
		panic(e)
	}
	return v
}
func codeFor(e error) api.ErrorCode {
	switch {
	case errors.Is(e, context.Canceled), errors.Is(e, context.DeadlineExceeded):
		return api.Canceled
	case errors.Is(e, errChanged):
		return api.StaleObservation
	case errors.Is(e, errInvalid):
		return api.Invalid
	case nativePermission(e):
		return api.Permission
	case errors.Is(e, errRedirect), errors.Is(e, errLimit), nativeRedirect(e):
		return api.Unsupported
	default:
		return api.IOFailure
	}
}

type scan struct {
	adapter          *Adapter
	scope            api.WorktreeScope
	started          time.Time
	definitions      []api.LaunchDefinition
	diagnostics      []api.Diagnostic
	sources          []api.DiscoverySourceDiagnostic
	visited, skipped int
	partial, more    bool
	physical         map[api.DirectoryIdentity]bool
}

func (s *scan) notice(locator string, code api.ErrorCode, reason string) {
	cap := s.adapter.limits.Directories + s.adapter.limits.Candidates
	if len(s.sources) > cap {
		s.partial = true
		return
	}
	if len(s.sources) == cap {
		locator = ""
		code = api.Unsupported
		reason = "diagnostic-limit"
	}
	n := diagnostic(code, reason)
	s.diagnostics = append(s.diagnostics, n)
	v, e := api.NewDiscoverySourceDiagnostic(api.DiscoverySourceDiagnosticData{Locator: locator, Diagnostic: n})
	if e != nil {
		panic(e)
	}
	s.sources = append(s.sources, v)
	if code != api.Conflict {
		s.partial = true
	}
}
func (s *scan) observation() api.DiscoveryObservation {
	finished := time.Now().UTC()
	if finished.Before(s.started) {
		finished = s.started
		s.partial = true
	}
	interval, e := api.NewObservationInterval(api.ObservationIntervalData{StartedAt: s.started, FinishedAt: finished})
	if e != nil {
		panic(e)
	}
	var profiles []api.ProviderProfile
	for _, p := range s.adapter.providers {
		v, e := api.NewProviderProfile(api.ProviderProfileData{Provider: p, Profile: providerProfile(p), Version: s.adapter.version("profile", providerProfile(p), s.adapter.limits)})
		if e != nil {
			panic(e)
		}
		profiles = append(profiles, v)
	}
	complete := api.Complete
	if s.partial {
		complete = api.Partial
	} else if s.more {
		complete = api.More
	}
	version := s.adapter.version("scan", s.scope.Data(), s.definitions, s.sources, s.visited, s.skipped, complete)
	id, e := api.NewObservationID(s.adapter.issuer + ":" + fmt.Sprint(s.adapter.observations.Add(1)))
	if e != nil {
		panic(e)
	}
	v, e := api.NewDiscoveryObservation(api.DiscoveryObservationData{ObservationID: id, WorktreeID: s.scope.Data().ID, Interval: interval, SourceVersion: version, Completeness: complete, Visited: uint64(s.visited), Skipped: uint64(s.skipped), ProviderProfiles: profiles, Sources: s.sources})
	if e != nil {
		panic(e)
	}
	return v
}

var excluded = map[string]bool{".git": true, ".gh-tree": true, "node_modules": true, "vendor": true, "dist": true, "build": true, "out": true, "target": true, ".next": true, ".cache": true}

func (d *Adapter) Discover(ctx context.Context, req api.DiscoveryRequest) (api.DiscoveryResult, error) {
	if d == nil || d.issuer == "" || ctx == nil || !req.Valid() {
		return api.DiscoveryResult{}, errors.New("invalid discovery request")
	}
	r := req.Data()
	s := &scan{adapter: d, scope: r.Worktree, started: time.Now().UTC(), physical: map[api.DirectoryIdentity]bool{}}
	var err error
	if err = ctx.Err(); err == nil {
		var root *directory
		root, err = acquireRoot(r.Worktree)
		if err == nil {
			err = s.walk(ctx, root, nil)
			if e := sameDirectory(root); e != nil {
				err = e
				s.definitions = nil
			}
			fresh, e := acquireRoot(r.Worktree)
			if e != nil {
				err = e
				s.definitions = nil
			} else {
				fresh.close()
			}
			root.close()
		}
	}
	if err != nil {
		s.notice(r.Worktree.Data().RootLocator, codeFor(err), "discovery-observation-failed")
	}
	sortDefinitions(s.definitions)
	var saved []api.SavedLaunchObservation
	if len(r.Saved) > d.limits.Candidates {
		s.skipped += len(r.Saved)
		s.notice("", api.Unsupported, "saved-candidate-limit")
	} else if version, ok := r.SavedVersion.Value(); ok {
		aliases := map[string]int{}
		for _, entry := range r.Saved {
			aliases[entry.Data().Alias]++
		}
		for _, entry := range r.Saved {
			if ctx.Err() != nil {
				s.notice(entry.Data().Alias, api.Canceled, "saved-observation-canceled")
				break
			}
			if aliases[entry.Data().Alias] > 1 {
				n := diagnostic(api.Invalid, "duplicate-saved-alias")
				v, _ := api.NewSavedLaunchObservation(api.SavedLaunchObservationData{Alias: entry.Data().Alias, StorageVersion: version, Diagnostics: []api.Diagnostic{n}})
				saved = append(saved, v)
				continue
			}
			saved = append(saved, d.observeSaved(ctx, r.Worktree, entry, version))
		}
	}
	result, e := api.NewDiscoveryResult(api.DiscoveryResultData{WorktreeID: r.Worktree.Data().ID, Definitions: s.definitions, Saved: saved, Observation: s.observation(), Diagnostics: s.diagnostics})
	if e != nil {
		return api.DiscoveryResult{}, e
	}
	return result, err
}
func sortDefinitions(defs []api.LaunchDefinition) {
	sort.Slice(defs, func(i, j int) bool {
		a, b := defs[i].Data(), defs[j].Data()
		x, y := strings.Join(a.ProjectComponents, "/"), strings.Join(b.ProjectComponents, "/")
		if x != y {
			return x < y
		}
		if a.Provider != b.Provider {
			return providerKey(a.Provider) < providerKey(b.Provider)
		}
		if a.Member != b.Member {
			return a.Member < b.Member
		}
		return a.LaunchPointID.Key() < b.LaunchPointID.Key()
	})
}
func (s *scan) walk(ctx context.Context, dir *directory, parts []string) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if s.visited >= s.adapter.limits.Directories {
		s.more = true
		s.skipped++
		s.notice(dir.path, api.Unsupported, "directory-limit")
		return nil
	}
	s.visited++
	if s.physical[dir.identity] {
		s.skipped++
		s.notice(dir.path, api.Unsupported, "duplicate-physical-project")
		return nil
	}
	s.physical[dir.identity] = true
	for _, p := range s.adapter.providers {
		defs, notices, e := s.adapter.project(ctx, s.scope, dir, parts, p, "")
		for _, n := range notices {
			s.notice(dir.path, n.Data().Code, n.Data().Reason)
		}
		if e != nil {
			s.notice(dir.path+"/"+providerKey(p), codeFor(e), "provider-observation-failed")
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		for _, def := range defs {
			if len(s.definitions) >= s.adapter.limits.Candidates {
				s.more = true
				s.notice(dir.path, api.Unsupported, "candidate-limit")
				break
			}
			s.definitions = append(s.definitions, def)
		}
	}
	// Directory enumeration is streaming; there is no all-tree/all-directory
	// unbounded allocation. Context is checked between each bounded OS batch.
	children := nameHeap{}
	directoryLimit := false
	for {
		if e := ctx.Err(); e != nil {
			return e
		}
		entries, e := dir.file.ReadDir(64)
		for _, entry := range entries {
			if excluded[strings.ToLower(entry.Name())] {
				s.skipped++
				continue
			}
			if entry.Type()&os.ModeSymlink != 0 {
				s.skipped++
				s.notice(dir.path+"/"+entry.Name(), api.Unsupported, "child-link-skipped")
				continue
			}
			if !entry.IsDir() {
				continue
			}
			if len(parts) >= s.adapter.limits.Depth {
				s.skipped++
				s.more = true
				s.notice(dir.path+"/"+entry.Name(), api.Unsupported, "depth-limit")
				continue
			}
			if len(children)+s.visited >= s.adapter.limits.Directories {
				s.skipped++
				s.more = true
				directoryLimit = true
				if len(children) > 0 && entry.Name() < children[0] {
					children[0] = entry.Name()
					heap.Fix(&children, 0)
				}
				continue
			}
			heap.Push(&children, entry.Name())
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			s.notice(dir.path, codeFor(e), "directory-enumeration-failed")
			break
		}
	}
	if directoryLimit {
		s.notice(dir.path, api.Unsupported, "directory-limit")
	}
	sort.Strings(children)
	for _, name := range children {
		if e := ctx.Err(); e != nil {
			return e
		}
		if s.visited >= s.adapter.limits.Directories {
			s.skipped++
			s.more = true
			continue
		}
		child, e := childDirectory(dir, name)
		if e != nil {
			s.skipped++
			s.notice(dir.path+"/"+name, codeFor(e), "child-directory-unproved")
			continue
		}
		e = s.walk(ctx, child, append(append([]string(nil), parts...), name))
		check := sameNamedDirectory(dir, name, child)
		child.close()
		if check != nil {
			s.notice(dir.path+"/"+name, api.StaleObservation, "child-directory-changed")
			s.removeProject(append(append([]string(nil), parts...), name))
		}
		if e != nil {
			return e
		}
	}
	return nil
}
func (s *scan) removeProject(parts []string) {
	prefix := strings.Join(parts, "/")
	keep := s.definitions[:0]
	for _, def := range s.definitions {
		p := strings.Join(def.Data().ProjectComponents, "/")
		if p != prefix && !strings.HasPrefix(p, prefix+"/") {
			keep = append(keep, def)
		}
	}
	s.definitions = keep
}

func (d *Adapter) project(ctx context.Context, scope api.WorktreeScope, dir *directory, parts []string, provider api.ProviderKind, override string) ([]api.LaunchDefinition, []api.Diagnostic, error) {
	var observed []fileObservation
	var notices []api.Diagnostic
	var manifest fileObservation
	var parsed parsed
	var err error
	executable := providerKey(provider)
	if provider == api.Npm {
		manifest, err = observeFile(ctx, dir, "package.json", true, d.limits.ManifestBytes)
		if err != nil || manifest.state == "absent" {
			return nil, nil, err
		}
		parsed, err = parseNpm(ctx, manifest.data)
		if err != nil {
			if ctx.Err() != nil {
				return nil, nil, ctx.Err()
			}
			return nil, nil, fmt.Errorf("%w: %v", errInvalid, err)
		}
		regular := map[string]bool{}
		for _, name := range []string{"pnpm-lock.yaml", "yarn.lock", "package-lock.json", "npm-shrinkwrap.json"} {
			o, e := observeFile(ctx, dir, name, false, d.limits.ManifestBytes)
			observed = append(observed, o)
			if e != nil {
				if errors.Is(e, errRedirect) || nativeRedirect(e) {
					notices = append(notices, diagnostic(api.Unsupported, "redirected-or-nonregular-lock"))
					continue
				}
				return nil, notices, e
			}
			regular[name] = o.state == "regular"
		}
		if regular["pnpm-lock.yaml"] {
			executable = "pnpm"
		} else if regular["yarn.lock"] {
			executable = "yarn"
		}
		count := 0
		for _, present := range regular {
			if present {
				count++
			}
		}
		if count > 1 {
			notices = append(notices, diagnostic(api.Conflict, "conflicting-colocated-locks"))
		}
		if parsed.manager != "" && strings.SplitN(parsed.manager, "@", 2)[0] != executable {
			notices = append(notices, diagnostic(api.Conflict, "package-manager-disagreement"))
		}
	} else {
		for _, name := range []string{"GNUmakefile", "makefile", "Makefile"} {
			o, e := observeFile(ctx, dir, name, true, d.limits.ManifestBytes)
			observed = append(observed, o)
			if e != nil {
				return nil, nil, e
			}
			if manifest.state == "" && o.state == "regular" {
				manifest = o
				break
			}
		}
		if manifest.state == "" {
			return nil, nil, nil
		}
		parsed, err = parseMake(ctx, manifest.data, d.limits.MakeLineBytes)
		if err != nil {
			if ctx.Err() != nil {
				return nil, nil, ctx.Err()
			}
			return nil, nil, fmt.Errorf("%w: %v", errInvalid, err)
		}
	}
	for _, n := range parsed.notices {
		notices = append(notices, diagnostic(api.Unsupported, n))
	}
	if override != "" {
		if !validExecutable(override) {
			return nil, notices, errRedirect
		}
		executable = override
	}
	versionParts := []any{scope.Data(), parts, dir.identity, providerProfile(provider), manifest.versionBytes(), executable}
	var inputs []api.ProjectInputObservation
	for _, o := range observed {
		versionParts = append(versionParts, o.versionBytes())
		input, e := api.NewProjectInputObservation(api.ProjectInputObservationData{Locator: o.name, Identity: d.version("input-object", o.name, o.state, o.identity), Content: d.version("input-content", o.name, o.state, o.digest), Regular: o.state == "regular"})
		if e != nil {
			return nil, notices, e
		}
		inputs = append(inputs, input)
	}
	version := d.version("project", versionParts...)
	source, e := api.NewProjectSource(api.ProjectSourceData{ManifestLocator: manifest.name, ManifestIdentity: d.version("manifest-object", manifest.identity), Content: version, Inputs: inputs, ParserProfile: providerProfile(provider), RootIdentity: scope.Data().RootIdentity, ProjectIdentity: dir.identity})
	if e != nil {
		return nil, notices, e
	}
	if len(parsed.members) > d.limits.Candidates {
		parsed.members = parsed.members[:d.limits.Candidates]
		notices = append(notices, diagnostic(api.Unsupported, "candidate-limit"))
	}
	defs := make([]api.LaunchDefinition, 0, len(parsed.members))
	for _, m := range parsed.members {
		if ctx.Err() != nil {
			return defs, notices, ctx.Err()
		}
		if strings.ContainsRune(m.name, 0) {
			notices = append(notices, diagnostic(api.Invalid, "invalid-script-member"))
			continue
		}
		id, e := domain.NewLaunchPointID(scope.Data().ID, providerKey(provider), strings.Join(parts, "/"), m.name)
		if e != nil {
			return nil, notices, e
		}
		args := []string{"run", m.name}
		if provider == api.Make {
			args = []string{"-f", manifest.name, m.name}
		}
		execution, e := api.NewArgvExecution(api.ArgvExecutionData{Executable: executable, Arguments: args})
		if e != nil {
			return nil, notices, e
		}
		ns := append([]api.Diagnostic(nil), notices...)
		if !m.valid {
			ns = append(ns, diagnostic(api.Invalid, "invalid-script-member"))
		}
		def, e := api.NewLaunchDefinition(api.LaunchDefinitionData{LaunchPointID: id, Provider: provider, ProjectComponents: parts, Member: m.name, DisplayPath: strings.Join(parts, "/"), Label: m.name, ProjectSource: source, EffectiveExecutable: execution, Available: m.valid, Diagnostics: ns})
		if e != nil {
			return nil, notices, e
		}
		defs = append(defs, def)
	}
	return defs, notices, nil
}
func validExecutable(s string) bool { return s != "" && !strings.ContainsAny(s, "\x00\r\n") }

type nameHeap []string

func (h nameHeap) Len() int           { return len(h) }
func (h nameHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h nameHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *nameHeap) Push(v any)        { *h = append(*h, v.(string)) }
func (h *nameHeap) Pop() any          { last := len(*h) - 1; v := (*h)[last]; *h = (*h)[:last]; return v }
