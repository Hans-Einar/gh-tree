package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"strconv"
	"strings"
)

func nonblank(s string) bool       { return strings.TrimSpace(s) != "" }
func containsEquals(s string) bool { return strings.ContainsRune(s, '=') }
func safeLocator(s string) bool    { return nonempty(s) && !strings.ContainsAny(s, "\r\n?#@") }
func validEnvironment(d EnvironmentPolicyData) bool {
	seen := map[string]bool{}
	for _, e := range d.Set {
		k := e.data.Name
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	for _, k := range d.Remove {
		if !nonempty(k) || containsEquals(k) {
			return false
		}
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	return true
}

func windowsEnvironment(d EnvironmentPolicyData) bool {
	seen := map[string]bool{}
	for _, v := range d.Set {
		k := strings.ToUpper(v.data.Name)
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	for _, name := range d.Remove {
		k := strings.ToUpper(name)
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	return true
}
func duplicateRecoveryIDs(ids []RecoveryID) bool {
	seen := map[RecoveryID]bool{}
	for _, id := range ids {
		if seen[id] {
			return true
		}
		seen[id] = true
	}
	return false
}
func duplicatePaths(paths []GitPath) bool {
	seen := map[GitPath]bool{}
	for _, p := range paths {
		if seen[p] {
			return true
		}
		seen[p] = true
	}
	return false
}
func containsChoice(choices []Choice, c Choice) bool {
	for _, v := range choices {
		if v == c {
			return true
		}
	}
	return false
}
func sameLocal(a, b domain.Revision) bool {
	return a.Valid() && b.Valid() && a.Repository().Scope() == domain.LocalCommon && a.Repository() == b.Repository() && a.OID().Format() == b.OID().Format()
}
func validateReadCommitsResult(d ReadCommitsResultData) error {
	repo := d.Endpoint.Repository()
	if repo.Scope() != domain.LocalCommon {
		return invalid("history result endpoint scope")
	}
	for _, c := range d.Commits {
		if c.data.Revision.Repository() != repo {
			return invalid("history result commit scope")
		}
	}
	if o, p := d.Observation.Value(); p && o.data.Repository != repo {
		return invalid("history result observation scope")
	}
	return nil
}
func validateMergeBaseResult(d MergeBaseResultData) error {
	if !sameLocal(d.Left, d.Right) {
		return invalid("merge-base result endpoints")
	}
	if outcome, p := d.Outcome.Value(); p {
		switch b := outcome.(type) {
		case UniqueMergeBase:
			if !sameLocal(d.Left, b.data.Base) {
				return invalid("merge-base result scope")
			}
		case AmbiguousMergeBase:
			for _, r := range b.data.Candidates {
				if !sameLocal(d.Left, r) {
					return invalid("merge-base candidate scope")
				}
			}
		}
	}
	if o, p := d.Observation.Value(); p && o.data.Repository != d.Left.Repository() {
		return invalid("merge-base result observation scope")
	}
	return nil
}
func validFullRef(ref, prefix string) bool {
	if !strings.HasPrefix(ref, prefix) || len(ref) == len(prefix) {
		return false
	}
	repo, _ := domain.NewRepositoryID(domain.LocalCommon, "syntax-only")
	_, err := domain.NewBranchID(repo, domain.Local, strings.TrimPrefix(ref, "refs/"))
	return err == nil
}
func filePath(v FileState) GitPath {
	switch x := v.(type) {
	case AbsentFile:
		return x.data.Path
	case PresentFile:
		return x.data.Path
	}
	return GitPath{}
}
func expectedWorktree(e GitExpectedState, w domain.WorktreeID, full bool) bool {
	if !e.Valid() || !w.Valid() {
		return false
	}
	d := e.data
	id, p := d.Worktree.Value()
	if !p || id != w || d.Repository != w.Repository() {
		return false
	}
	h, p := d.Head.Value()
	if !p || !h.MatchesWorktree(w) {
		return false
	}
	return !full || (d.Index.Present() && d.WorktreeState.Present())
}
func validateWorktree(d WorktreeFactsData) error {
	if d.ID.Repository() != d.Observation.data.Repository {
		return invalid("worktree observation")
	}
	if w, p := d.Observation.data.Worktree.Value(); p && w != d.ID {
		return invalid("worktree observation ID")
	}
	if s, p := d.Scope.Value(); p && s.data.ID != d.ID {
		return invalid("worktree scope ID")
	}
	if h, p := d.Head.Value(); p && !h.MatchesWorktree(d.ID) {
		return invalid("worktree head")
	}
	if _, available := d.Availability.(AvailableWorktree); available && !d.Scope.Present() {
		return invalid("available scope")
	}
	return nil
}
func validateResolution(d ExactLocalResolutionData) error {
	if d.Local.Repository().Scope() != domain.LocalCommon || d.Local.Repository() != d.Observation.data.Repository || d.Local.OID() != d.Requested.ExpectedRevision().OID() {
		return invalid("exact resolved revision")
	}
	expected := d.Requested.ExpectedRevision()
	if expected.Repository().Scope() == domain.LocalCommon {
		if expected != d.Local {
			return invalid("foreign local resolution")
		}
	} else {
		b, ok := d.Binding.Value()
		if !ok || b.data.LocalRepository != d.Local.Repository() || b.data.RemoteRepository != expected.Repository() {
			return invalid("remote local association")
		}
		observed, ok := d.ObservedRemote.Value()
		if !ok || observed != expected {
			return invalid("expected remote endpoint")
		}
	}
	return nil
}
func validateStashPrepare(d PrepareStashRequestData) error {
	switch s := d.Intent.(type) {
	case CreateStashIntent:
		if !expectedWorktree(d.Expected, s.data.Worktree, true) {
			return invalid("stash capture expected")
		}
	case ApplyStashIntent:
		if !expectedWorktree(d.Expected, s.data.Worktree, true) {
			return invalid("stash apply expected")
		}
	case PopStashIntent:
		if !expectedWorktree(d.Expected, s.data.Worktree, true) {
			return invalid("stash pop expected")
		}
	case DropStashIntent:
		if d.Expected.data.Repository != s.data.Stash.Repository() {
			return invalid("stash drop expected")
		}
	default:
		return invalid("stash variant")
	}
	return nil
}
func gitFacet(f EffectFacet) bool {
	return f == ObjectAcquisition || f == Recovery || f == WorktreeBytes || f == Index || f == LocalRefsHead || f == LocalConfiguration || f == RemoteRefsPR
}
func validateReconcile(d ReconcileRequestData) error {
	if d.Repository.Scope() != domain.LocalCommon || d.Operation == d.OriginalOperation || len(d.Facets) == 0 {
		return invalid("reconciliation scope/operation")
	}
	if w, ok := d.Worktree.Value(); ok && w.Repository() != d.Repository {
		return invalid("reconciliation worktree")
	}
	seen := map[EffectFacet]bool{}
	for _, f := range d.Facets {
		if !gitFacet(f) || seen[f] {
			return invalid("reconciliation facet")
		}
		seen[f] = true
	}
	for _, r := range d.Recovery {
		if r.data.Layer != LayerGit {
			return invalid("foreign recovery")
		}
	}
	return validateRecoveryReferences(d.PriorEffects, d.Recovery)
}
func validateGitRecovery(e EffectReport, r []RecoveryRecord) error {
	for _, v := range r {
		if v.data.Layer != LayerGit {
			return invalid("Git recovery owner")
		}
	}
	return validateRecoveryReferences(e, r)
}
func validateGitMutation(d GitMutationResultData) error {
	var kind GitMutationKind
	switch d.Outcome.(type) {
	case WorktreeCreated:
		kind = CreateMutation
	case WorktreeRetargeted:
		kind = RetargetMutation
	case IndexChanged:
		kind = StageMutation
	case CommitCreated:
		kind = CommitMutation
	case TrackedRestored:
		kind = RestoreMutation
	case StashCreated, StashCreatedCleanupRefused, StashApplied, AppliedWithConflicts, StashDropped:
		kind = StashMutation
	case BranchCreated:
		kind = BranchMutation
	case Pushed:
		kind = PushMutation
	}
	if kind.Valid() && kind != d.Kind {
		return invalid("mutation outcome kind")
	}
	return validateGitRecovery(d.Effects, d.Recovery)
}
func validateSession(d SessionSnapshotData) error {
	if d.Display.data.Cwd.data.Worktree.data.ID != d.WorktreeID {
		return invalid("session cwd worktree")
	}
	if old, ok := d.RestartOf.Value(); ok && old == d.SessionID {
		return invalid("restart self")
	}
	if (d.Phase == Cleaned) != (d.Cleanup.data.State == CleanupComplete) {
		return invalid("cleaned barrier")
	}
	if d.Phase == CleanupFailed && d.Cleanup.data.State != CleanupFailedState {
		return invalid("cleanup failure state")
	}
	if a, ok := d.AcquiredCwd.Value(); ok && a.data.Observation.data.Worktree.data.ID != d.WorktreeID {
		return invalid("acquired cwd")
	}
	if a, p := d.AcquiredCwd.Value(); p && !sameCwdSubject(a.data.Observation, d.Display.data.Cwd) {
		return invalid("snapshot acquired cwd specification")
	}
	for _, r := range d.Cleanup.data.Residuals {
		if id, p := r.data.SessionID.Value(); p && id != d.SessionID {
			return invalid("session residual identity")
		}
	}
	if d.Display.data.Terminal == Pipes && (d.Capabilities.data.Resize || d.Capabilities.data.TerminalETX) {
		return invalid("pipe terminal capabilities")
	}
	return nil
}
func validateOutput(d SessionOutputResultData) error {
	if d.RetainedStart > d.End || d.End-d.RetainedStart > 262144 || d.NextOffset < d.RetainedStart || d.NextOffset > d.End {
		return invalid("output bounds")
	}
	next := d.NextOffset
	total := uint64(0)
	if len(d.Chunks) > 0 {
		next = d.Chunks[0].data.Offset
	}
	if g, ok := d.Gap.Value(); ok {
		if g.data.To != d.RetainedStart || next != d.RetainedStart {
			return invalid("output gap")
		}
	}
	for _, chunk := range d.Chunks {
		c := chunk.data
		if c.Offset != next || c.Offset < d.RetainedStart || c.Sequence.Value() > d.Sequence.Value() {
			return invalid("output ordering")
		}
		next += uint64(len(c.Bytes))
		total += uint64(len(c.Bytes))
		if next > d.End {
			return invalid("output end")
		}
	}
	if total > 262144 || next != d.NextOffset {
		return invalid("output next offset")
	}
	return nil
}
func validRemoteLocator(d RemoteRepositoryLocatorData) bool {
	if !nonempty(d.Host) || !nonempty(d.Owner) || !nonempty(d.Name) || d.Host != strings.ToLower(d.Host) || d.Owner != strings.ToLower(d.Owner) || d.Name != strings.ToLower(d.Name) || strings.ContainsAny(d.Host+d.Owner+d.Name, "/@:?#\\ \r\n\t") || strings.HasSuffix(d.Name, ".git") {
		return false
	}
	for _, label := range strings.Split(d.Host, ".") {
		if label == "" || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}
func remoteURL(url string, l RemoteRepositoryLocator, number uint64) bool {
	d := l.data
	expected := "https://" + d.Host + "/" + d.Owner + "/" + d.Name
	if number > 0 {
		expected += "/pull/" + strconv.FormatUint(number, 10)
	}
	return url == expected
}
func remotePage(p PageRequest) bool {
	if !p.Valid() || p.data.Limit > 100 {
		return false
	}
	_, bad := p.data.Continuation.(OffsetPage)
	return !bad
}
func endpointRepository(v PullRequestEndpoint) (domain.RepositoryID, bool) {
	switch e := v.(type) {
	case AvailableEndpoint:
		return e.data.Repository.data.ID, true
	case UnavailableEndpoint:
		if r, p := e.data.KnownRepository.Value(); p {
			return r, true
		}
		if b, p := e.data.KnownBranch.Value(); p {
			return b.Repository(), true
		}
		if r, p := e.data.KnownRevision.Value(); p {
			return r.Repository(), true
		}
		return domain.RepositoryID{}, false
	}
	return domain.RepositoryID{}, false
}
func endpointMatches(v PullRequestEndpoint, e EndpointExpectation) bool {
	a, ok := v.(AvailableEndpoint)
	return ok && a.data.Branch == e.data.Branch && a.data.Revision == e.data.Revision
}
func validateUnavailableEndpoint(d UnavailableEndpointData) error {
	repo, hasRepo := d.KnownRepository.Value()
	if hasRepo && repo.Scope() != domain.Remote {
		return invalid("endpoint repository")
	}
	branch, hasBranch := d.KnownBranch.Value()
	if hasBranch && (branch.Kind() != domain.RemoteHead || (hasRepo && branch.Repository() != repo)) {
		return invalid("endpoint branch")
	}
	revision, hasRevision := d.KnownRevision.Value()
	if hasRevision && (revision.Repository().Scope() != domain.Remote || (hasRepo && revision.Repository() != repo) || (hasBranch && revision.Repository() != branch.Repository())) {
		return invalid("endpoint revision")
	}
	return nil
}
func validSavedBinding(saved []SavedLaunchEntry, version Optional[StorageVersion]) bool {
	v, p := version.Value()
	if len(saved) > 0 && !p {
		return false
	}
	return !p || v.Family() == RunConfig
}
func launchMatchesWorktree(s LaunchSelection, w domain.WorktreeID) bool {
	switch v := s.(type) {
	case DiscoveredLaunch:
		return v.data.Member.data.LaunchPointID.Worktree() == w
	case OrderedMakeLaunch:
		for _, m := range v.data.Members {
			if m.data.LaunchPointID.Worktree() != w {
				return false
			}
		}
		return len(v.data.Members) > 0
	case SavedLaunch:
		return v.data.LaunchPointID.Worktree() == w
	}
	return false
}

// Domain's launch key is length framed, not slash delimited. Only the pure
// semantic components are inspected here; no pathname or manifest is resolved.
func launchKeyParts(id domain.LaunchPointID) (string, string, string, bool) {
	if !id.Valid() {
		return "", "", "", false
	}
	key := id.Key()
	parts := [3]string{}
	for i := range parts {
		colon := strings.IndexByte(key, ':')
		if colon < 1 {
			return "", "", "", false
		}
		n, e := strconv.Atoi(key[:colon])
		key = key[colon+1:]
		if e != nil || n < 0 || n > len(key) {
			return "", "", "", false
		}
		parts[i] = key[:n]
		key = key[n:]
	}
	return parts[0], parts[1], parts[2], key == ""
}
func validOrderedMake(members []MemberSelection) bool {
	if len(members) == 0 {
		return false
	}
	first := members[0].data
	provider, project, _, ok := launchKeyParts(first.LaunchPointID)
	if !ok || provider != "make" {
		return false
	}
	for _, v := range members {
		d := v.data
		p, pr, _, ok := launchKeyParts(d.LaunchPointID)
		if !ok || p != provider || pr != project || d.LaunchPointID.Worktree() != first.LaunchPointID.Worktree() || d.SourceVersion != first.SourceVersion {
			return false
		}
	}
	return true
}
func validLaunchDefinition(d LaunchDefinitionData) bool {
	provider, _, member, ok := launchKeyParts(d.LaunchPointID)
	if !ok || member != d.Member {
		return false
	}
	return d.Provider == Npm && provider == "npm" || d.Provider == Make && provider == "make"
}
func createModeScope(mode CreateMode, repo domain.RepositoryID) bool {
	switch v := mode.(type) {
	case DetachedCreate:
		return true
	case CreateNewBranch:
		return v.data.Branch.Repository() == repo
	}
	return false
}
func retargetModeScope(mode RetargetMode, repo domain.RepositoryID, target Optional[domain.Revision]) bool {
	switch v := mode.(type) {
	case DetachRetarget:
		return true
	case AttachExisting:
		return v.data.Branch.Repository() == repo
	case CreateNewBranch:
		return v.data.Branch.Repository() == repo
	case FastForward:
		if v.data.Branch.Repository() != repo {
			return false
		}
		if r, p := target.Value(); p && r != v.data.To {
			return false
		}
		return true
	}
	return false
}
