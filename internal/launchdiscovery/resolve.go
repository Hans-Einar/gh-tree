package launchdiscovery

import (
	"context"
	"errors"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"strings"
	"time"
)

// acquiredProject holds every intermediate project descriptor until the complete
// observation is checked. It owns no capability beyond this one request.
type acquiredProject struct {
	root  *directory
	dirs  []*directory
	parts []string
}

func (p *acquiredProject) close() {
	for i := len(p.dirs) - 1; i >= 0; i-- {
		p.dirs[i].close()
	}
	p.root.close()
}
func (p *acquiredProject) last() *directory {
	if len(p.dirs) == 0 {
		return p.root
	}
	return p.dirs[len(p.dirs)-1]
}
func (p *acquiredProject) check(scope api.WorktreeScope) error {
	if e := sameDirectory(p.root); e != nil {
		return e
	}
	parent := p.root
	for i, d := range p.dirs {
		if e := sameDirectory(d); e != nil {
			return e
		}
		if e := sameNamedDirectory(parent, p.parts[i], d); e != nil {
			return e
		}
		parent = d
	}
	fresh, e := acquireRoot(scope)
	if e != nil {
		return e
	}
	fresh.close()
	return nil
}
func openProject(ctx context.Context, scope api.WorktreeScope, parts []string) (*acquiredProject, error) {
	root, e := acquireRoot(scope)
	if e != nil {
		return nil, e
	}
	p := &acquiredProject{root: root, parts: append([]string(nil), parts...)}
	for _, name := range parts {
		if e = ctx.Err(); e != nil {
			p.close()
			return nil, e
		}
		next, e := childDirectory(p.last(), name)
		if e != nil {
			p.close()
			return nil, e
		}
		p.dirs = append(p.dirs, next)
	}
	return p, nil
}

func (d *Adapter) registered(p api.ProviderKind) bool {
	for _, v := range d.providers {
		if p == v {
			return true
		}
	}
	return false
}
func (d *Adapter) savedDefinitions(ctx context.Context, scope api.WorktreeScope, entry api.SavedLaunchEntry) ([]api.LaunchDefinition, error) {
	e := entry.Data()
	stored := e.Definition.Data()
	var provider api.ProviderKind
	switch stored.Provider {
	case "npm":
		provider = api.Npm
	case "make":
		provider = api.Make
	default:
		return nil, errRedirect
	}
	if !d.registered(provider) {
		return nil, errRedirect
	}
	dir, _ := stored.Dir.Value()
	parts, err := projectParts(dir)
	if err != nil || len(parts) > d.limits.Depth {
		return nil, errRedirect
	}
	for _, part := range parts {
		if excluded[strings.ToLower(part)] {
			return nil, errRedirect
		}
	}
	override, present := stored.Command.Value()
	if present && !validExecutable(override) {
		return nil, errRedirect
	}
	var names []string
	if provider == api.Npm {
		name, ok := stored.Script.Value()
		targets, _ := stored.Targets.Value()
		if !ok || !safeScript(name) || len(targets) != 0 {
			return nil, errRedirect
		}
		names = []string{name}
	} else {
		script, _ := stored.Script.Value()
		names, _ = stored.Targets.Value()
		if len(names) == 0 || len(names) > d.limits.Candidates || script != "" {
			return nil, errRedirect
		}
		for _, name := range names {
			if !safeTarget(name) {
				return nil, errRedirect
			}
		}
	}
	p, err := openProject(ctx, scope, parts)
	if err != nil {
		return nil, err
	}
	defer p.close()
	defs, _, err := d.project(ctx, scope, p.last(), parts, provider, override)
	if err != nil {
		return nil, err
	}
	if err = p.check(scope); err != nil {
		return nil, err
	}
	byName := map[string]api.LaunchDefinition{}
	for _, def := range defs {
		byName[def.Data().Member] = def
	}
	selected := make([]api.LaunchDefinition, 0, len(names))
	for _, name := range names {
		def, ok := byName[name]
		if !ok || !def.Data().Available {
			return nil, errChanged
		}
		selected = append(selected, def)
	}
	// The selection version also binds ordered saved members and explicit override
	// presence. Alias/storage stay their separate exact API binding dimensions.
	for i, def := range selected {
		dd := def.Data()
		source := dd.ProjectSource.Data()
		source.Content = d.version("saved-project", source.Content, stored.Provider, parts, names, present, override)
		dd.ProjectSource, err = api.NewProjectSource(source)
		if err != nil {
			return nil, err
		}
		selected[i], err = api.NewLaunchDefinition(dd)
		if err != nil {
			return nil, err
		}
	}
	return selected, nil
}
func (d *Adapter) observeSaved(ctx context.Context, scope api.WorktreeScope, entry api.SavedLaunchEntry, version api.StorageVersion) api.SavedLaunchObservation {
	data := api.SavedLaunchObservationData{Alias: entry.Data().Alias, StorageVersion: version}
	defs, e := d.savedDefinitions(ctx, scope, entry)
	if e != nil {
		data.Diagnostics = []api.Diagnostic{diagnostic(codeFor(e), "saved-launch-unavailable")}
	} else {
		first := defs[0]
		data.LaunchPointID = api.Some(first.Data().LaunchPointID)
		data.Definition = api.Some(first)
		data.SourceVersion = api.Some(first.Data().ProjectSource.Data().Content)
		data.Diagnostics = first.Data().Diagnostics
	}
	v, e := api.NewSavedLaunchObservation(data)
	if e != nil {
		panic(e)
	}
	return v
}

func (d *Adapter) Resolve(ctx context.Context, req api.ResolveLaunchRequest) (api.ResolveLaunchResult, error) {
	if d == nil || d.issuer == "" || ctx == nil || !req.Valid() {
		return api.ResolveLaunchResult{}, errors.New("invalid resolve request")
	}
	r := req.Data()
	s := &scan{adapter: d, scope: r.Worktree, started: time.Now().UTC()}
	refuse := func(e error) (api.ResolveLaunchResult, error) {
		s.notice(r.Worktree.Data().RootLocator, codeFor(e), "launch-selection-unproved")
		v, ce := api.NewResolveLaunchResult(api.ResolveLaunchResultData{Observation: s.observation(), Diagnostics: s.diagnostics})
		if ce != nil {
			return api.ResolveLaunchResult{}, ce
		}
		return v, e
	}
	if e := ctx.Err(); e != nil {
		return refuse(e)
	}
	var expected []api.MemberSelection
	var selected []api.LaunchDefinition
	var alias api.Optional[string]
	var savedVersion api.Optional[api.StorageVersion]
	switch selection := r.Selection.(type) {
	case api.DiscoveredLaunch:
		expected = []api.MemberSelection{selection.Data().Member}
	case api.OrderedMakeLaunch:
		expected = selection.Data().Members
	case api.SavedLaunch:
		pick := selection.Data()
		if len(r.Saved) > d.limits.Candidates {
			return refuse(errLimit)
		}
		count := 0
		for _, entry := range r.Saved {
			if entry.Data().Alias == pick.Alias {
				count++
			}
		}
		if count != 1 {
			return refuse(errInvalid)
		}
		var found bool
		for _, entry := range r.Saved {
			if entry.Data().Alias != pick.Alias {
				continue
			}
			found = true
			defs, e := d.savedDefinitions(ctx, r.Worktree, entry)
			if e != nil {
				return refuse(e)
			}
			selected = defs
			break
		}
		if !found || selected[0].Data().LaunchPointID != pick.LaunchPointID || selected[0].Data().ProjectSource.Data().Content != pick.SourceExpectation {
			return refuse(errChanged)
		}
		for _, def := range selected {
			m, e := api.NewMemberSelection(api.MemberSelectionData{LaunchPointID: def.Data().LaunchPointID, SourceVersion: def.Data().ProjectSource.Data().Content})
			if e != nil {
				return refuse(e)
			}
			expected = append(expected, m)
		}
		alias = api.Some(pick.Alias)
		savedVersion = api.Some(pick.StorageVersion)
	default:
		return refuse(errRedirect)
	}
	if len(expected) == 0 || len(expected) > d.limits.Candidates {
		return refuse(errLimit)
	}
	if len(selected) == 0 {
		// Decode only our own stable length-framed key, then independently observe
		// the exact named project. No scan cursor or adapter cache is authority.
		providerKey, project, _, ok := keyParts(expected[0].Data().LaunchPointID)
		if !ok {
			return refuse(errRedirect)
		}
		provider := api.Npm
		if providerKey == "make" {
			provider = api.Make
		} else if providerKey != "npm" {
			return refuse(errRedirect)
		}
		if !d.registered(provider) {
			return refuse(errRedirect)
		}
		parts, e := projectParts(project)
		if e != nil || len(parts) > d.limits.Depth {
			return refuse(errRedirect)
		}
		for _, part := range parts {
			if excluded[strings.ToLower(part)] {
				return refuse(errRedirect)
			}
		}
		p, e := openProject(ctx, r.Worktree, parts)
		if e != nil {
			return refuse(e)
		}
		defs, notices, e := d.project(ctx, r.Worktree, p.last(), parts, provider, "")
		s.diagnostics = append(s.diagnostics, notices...)
		if e == nil {
			e = p.check(r.Worktree)
		}
		p.close()
		if e != nil {
			return refuse(e)
		}
		byID := map[domain.LaunchPointID]api.LaunchDefinition{}
		for _, def := range defs {
			byID[def.Data().LaunchPointID] = def
		}
		for _, member := range expected {
			def, ok := byID[member.Data().LaunchPointID]
			if !ok || !def.Data().Available || def.Data().ProjectSource.Data().Content != member.Data().SourceVersion {
				return refuse(errChanged)
			}
			selected = append(selected, def)
		}
	}
	first := selected[0].Data()
	for i, def := range selected {
		v := def.Data()
		if v.Provider != first.Provider || v.ProjectSource.Data().Content != first.ProjectSource.Data().Content || v.ProjectSource.Data().ProjectIdentity != first.ProjectSource.Data().ProjectIdentity || v.ProjectSource.Data().ManifestIdentity != first.ProjectSource.Data().ManifestIdentity || v.EffectiveExecutable.Data().Executable != first.EffectiveExecutable.Data().Executable || v.LaunchPointID != expected[i].Data().LaunchPointID {
			return refuse(errChanged)
		}
	}
	if first.Provider == api.Npm && len(selected) != 1 {
		return refuse(errRedirect)
	}
	// Newly acquire and compare the actual cwd after source/member validation.
	p, e := openProject(ctx, r.Worktree, first.ProjectComponents)
	if e != nil {
		return refuse(e)
	}
	if p.last().identity != first.ProjectSource.Data().ProjectIdentity {
		p.close()
		return refuse(errChanged)
	}
	e = p.check(r.Worktree)
	projectIdentity := p.last().identity
	p.close()
	if e != nil {
		return refuse(e)
	}
	if e = ctx.Err(); e != nil {
		return refuse(e)
	}
	executionData := first.EffectiveExecutable.Data()
	if first.Provider == api.Make {
		executionData.Arguments = []string{"-f", first.ProjectSource.Data().ManifestLocator}
		for _, def := range selected {
			executionData.Arguments = append(executionData.Arguments, def.Data().Member)
		}
	}
	execution, e := api.NewArgvExecution(executionData)
	if e != nil {
		return refuse(e)
	}
	var versions []api.SourceVersion
	for _, member := range expected {
		versions = append(versions, member.Data().SourceVersion)
	}
	resolved, e := api.NewResolvedLaunchDefinition(api.ResolvedLaunchDefinitionData{Selected: expected, Alias: alias, Provider: first.Provider, ProjectComponents: first.ProjectComponents, ProjectSource: first.ProjectSource, EffectiveExecutable: execution, SourceVersions: versions, SavedVersion: savedVersion})
	if e != nil {
		return refuse(e)
	}
	cwd, e := api.NewCwdObservation(api.CwdObservationData{Worktree: r.Worktree, ProjectComponents: first.ProjectComponents, ProjectIdentity: projectIdentity, Source: d.version("cwd", r.Worktree.Data(), projectIdentity, first.ProjectComponents, versions)})
	if e != nil {
		return refuse(e)
	}
	env, e := api.NewEnvironmentPolicy(api.EnvironmentPolicyData{InheritBase: true})
	if e != nil {
		return refuse(e)
	}
	inv, e := api.NewInvocation(api.InvocationData{Execution: execution, Environment: env, Cwd: cwd, Terminal: api.Pipes, Geometry: r.Geometry, Label: first.Label})
	if e != nil {
		return refuse(e)
	}
	s.visited = len(first.ProjectComponents) + 1
	s.definitions = selected
	out, e := api.NewResolveLaunchResult(api.ResolveLaunchResultData{Definition: api.Some(resolved), Invocation: api.Some(inv), Observation: s.observation(), Diagnostics: s.diagnostics})
	return out, e
}
func keyParts(id domain.LaunchPointID) (string, string, string, bool) {
	key := id.Key()
	p := [3]string{}
	for i := range p {
		colon := strings.IndexByte(key, ':')
		if colon < 1 {
			return "", "", "", false
		}
		n := 0
		for _, b := range key[:colon] {
			if b < '0' || b > '9' || n > len(key) {
				return "", "", "", false
			}
			n = n*10 + int(b-'0')
		}
		key = key[colon+1:]
		if n > len(key) {
			return "", "", "", false
		}
		p[i] = key[:n]
		key = key[n:]
	}
	return p[0], p[1], p[2], key == ""
}
