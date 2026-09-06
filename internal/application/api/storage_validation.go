package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"unicode/utf8"
)

// Stored semantic strings preserve UTF-8 JSON contents, including controls that
// a later execution/path validator may refuse. Storage is not a provider parser.
func storageString(s string) bool { return utf8.ValidString(s) }

func storedText(v StoredField[string]) bool {
	if s, ok := v.Value(); ok {
		return storageString(s)
	}
	return v.Presence() != FieldNull
}
func validateUserDocument(d UserConfigDocumentData) error {
	if !d.UnknownMembers.Excludes("schemaVersion", "stripPrefixes", "repos", "scopedRepos") || !d.UnknownMembers.WithinDepth(1) {
		return invalid("user unknown members")
	}
	size := 64 + d.UnknownMembers.Size()
	if values, p := d.StripPrefixes.Value(); p {
		for _, s := range values {
			if !storageString(s) {
				return invalid("prefix bytes")
			}
			size += quotedSize(s) + 1
		}
	}
	if values, p := d.LegacyRepos.Value(); p {
		seen := map[string]bool{}
		for _, v := range values {
			if seen[v.data.Key] {
				return invalid("legacy repository duplicate")
			}
			seen[v.data.Key] = true
			size += quotedSize(v.data.Key) + configuredSize(v.data.Value)
		}
	}
	if values, p := d.ScopedRepos.Value(); p {
		seen := map[domain.RepositoryID]bool{}
		for _, v := range values {
			if seen[v.data.RepositoryID] {
				return invalid("scoped repository duplicate")
			}
			seen[v.data.RepositoryID] = true
			size += quotedSize(v.data.RepositoryID.Token()) + 16 + configuredSize(v.data.Value)
		}
	}
	if size > MaxDocumentBytes {
		return invalid("user document size")
	}
	return nil
}
func configuredSize(v ConfiguredRepository) int {
	d := v.data
	if !d.UnknownMembers.WithinDepth(3) {
		return MaxDocumentBytes + 1
	}
	size := 24 + d.UnknownMembers.Size()
	if targets, p := d.Worktrees.Value(); p {
		for _, t := range targets {
			x := t.data
			if !x.UnknownMembers.WithinDepth(5) {
				return MaxDocumentBytes + 1
			}
			size += quotedSize(x.Name) + 32 + x.UnknownMembers.Size()
			if s, p := x.Path.Value(); p {
				size += quotedSize(s)
			}
			if s, p := x.Branch.Value(); p {
				size += quotedSize(s)
			}
		}
	}
	return size
}
func validatePreferencesDocument(d PreferencesDocumentData) error {
	if !d.UnknownMembers.Excludes("schemaVersion", "lastFolders", "lastWorktrees", "scopedPreferences") || !d.UnknownMembers.WithinDepth(1) {
		return invalid("preference unknown members")
	}
	size := 80 + d.UnknownMembers.Size()
	for _, field := range []StoredField[[]LegacyStringPreference]{d.LegacyFolders, d.LegacyWorktrees} {
		if values, p := field.Value(); p {
			seen := map[string]bool{}
			for _, v := range values {
				if seen[v.data.Key] {
					return invalid("legacy preference duplicate")
				}
				seen[v.data.Key] = true
				size += quotedSize(v.data.Key) + quotedSize(v.data.Value) + 2
			}
		}
	}
	if values, p := d.ScopedPreferences.Value(); p {
		seen := map[domain.RepositoryID]bool{}
		for _, v := range values {
			x := v.data
			if seen[x.RepositoryID] || !x.UnknownMembers.WithinDepth(3) {
				return invalid("scoped preference duplicate/depth")
			}
			seen[x.RepositoryID] = true
			size += quotedSize(x.RepositoryID.Token()) + 64 + x.UnknownMembers.Size()
			if s, p := x.LastFolder.Value(); p {
				size += quotedSize(s)
			}
			if a, p := x.ActiveWorktree.Value(); p {
				if !a.data.UnknownMembers.WithinDepth(4) {
					return invalid("active unknown depth")
				}
				size += 64 + quotedSize(a.data.AdministrativeKey) + quotedSize(a.data.LastKnownPath) + a.data.UnknownMembers.Size()
			}
		}
	}
	if size > MaxDocumentBytes {
		return invalid("preferences document size")
	}
	return nil
}
func validateRunDocument(d RunConfigDocumentData) error {
	if !storedText(d.Default) || !d.UnknownMembers.Excludes("schemaVersion", "default", "launch") || !d.UnknownMembers.WithinDepth(1) {
		return invalid("run unknown members/default")
	}
	size := 48 + d.UnknownMembers.Size()
	if s, p := d.Default.Value(); p {
		size += quotedSize(s)
	}
	if values, p := d.Launch.Value(); p {
		seen := map[string]bool{}
		for _, v := range values {
			if seen[v.data.Alias] {
				return invalid("alias duplicate")
			}
			seen[v.data.Alias] = true
			x := v.data.Definition.data
			if !x.UnknownMembers.WithinDepth(3) {
				return invalid("saved unknown depth")
			}
			size += quotedSize(v.data.Alias) + 96 + quotedSize(x.Provider) + x.UnknownMembers.Size()
			for _, f := range []StoredField[string]{x.Dir, x.Script, x.Command} {
				if s, p := f.Value(); p {
					size += quotedSize(s)
				}
			}
			if t, p := x.Targets.Value(); p {
				for _, s := range t {
					size += quotedSize(s) + 1
				}
			}
		}
	}
	if size > MaxDocumentBytes {
		return invalid("run document size")
	}
	return nil
}
func validRecoverySubject(d RecoverySubjectData) bool {
	repo, rp := d.Repository.Value()
	if w, p := d.Worktree.Value(); p && rp && w.Repository() != repo {
		return false
	}
	if s, p := d.Stash.Value(); p && rp && s.Repository() != repo {
		return false
	}
	if b, p := d.Branch.Value(); p && rp && b.Repository() != repo {
		return false
	}
	if r, p := d.Revision.Value(); p && rp && r.Repository() != repo {
		return false
	}
	store, sp := d.Store.Value()
	family, fp := d.Family.Value()
	if sp != fp || sp && !nonempty(store) {
		return false
	}
	if fp {
		if family == RunConfig && !d.Worktree.Present() {
			return false
		}
		if family != RunConfig && (rp || d.Worktree.Present()) {
			return false
		}
	}
	return rp || d.Worktree.Present() || d.Stash.Present() || d.Revision.Present() || d.Branch.Present() || d.Session.Present() || sp
}
func sameRecoveryRecord(a, b RecoveryRecord) bool { return a.Valid() && b.Valid() && a.data == b.data }
func validateRecoveryReferences(e EffectReport, r []RecoveryRecord) error {
	records := map[RecoveryID]RecoveryRecord{}
	for _, v := range r {
		if !v.Valid() {
			return invalid("invalid recovery record")
		}
		id := v.data.RecoveryID
		if old, p := records[id]; p && !sameRecoveryRecord(old, v) {
			return invalid("inconsistent recovery duplicate")
		}
		records[id] = v
	}
	for _, facet := range e.data.Facets {
		for _, id := range facet.data.RecoveryIDs {
			if _, p := records[id]; !p {
				return invalid("dangling recovery ID")
			}
		}
	}
	return nil
}
func validateStorageRecovery(d StorageRecoveryData) error {
	r := d.Record.data
	s := r.Subject.data
	if r.Layer != LayerPersistence || r.Locator != d.Locator {
		return invalid("storage recovery identity/owner")
	}
	kinds := map[StorageRecoveryKind]RecoveryKind{Manifest: RecoveryManifest, RawOriginal: RecoveryRawOriginal, RetainedOriginal: RecoveryRetainedOriginal, RetainedPayload: RecoveryRetainedPayload}
	if r.Kind != kinds[d.Kind] {
		return invalid("storage recovery kind")
	}
	f, p := s.Family.Value()
	store, sp := s.Store.Value()
	if !p || !sp || f != d.Family {
		return invalid("storage recovery subject")
	}
	for _, v := range []Optional[RecoveryVersion]{r.Original, r.Proposed} {
		if item, ok := v.Value(); ok {
			x, p := item.(StorageRecoveryVersion)
			if !p || x.data.Version.Family() != d.Family || x.data.Version.Store() != store {
				return invalid("storage recovery version domain")
			}
		}
	}
	return nil
}
func validateStorageRecoveryList(v []StorageRecovery) error {
	records := map[RecoveryID]StorageRecovery{}
	for _, r := range v {
		if old, p := records[r.data.Record.data.RecoveryID]; p && old.data != r.data {
			return invalid("inconsistent storage recovery duplicate")
		}
		records[r.data.Record.data.RecoveryID] = r
	}
	return nil
}
func validateStorageCommit(d StorageCommitResultData) error {
	if err := validateStorageRecoveryList(d.Recovery); err != nil {
		return err
	}
	known := d.Outcome == Committed || d.Outcome == CommittedDurabilityUncertain
	if d.PublicationKnown != known {
		return invalid("storage publication fact")
	}
	if d.Outcome == Committed && d.Durability != SupportedCrashBarrierComplete {
		return invalid("commit durability")
	}
	if d.Outcome == CommittedDurabilityUncertain && d.Durability != DurabilityUncertain {
		return invalid("uncertain durability")
	}
	if d.Outcome == NotCommitted && d.Durability != DurabilityNotApplicable {
		return invalid("uncommitted durability")
	}
	if known && !d.ProposedVersion.Present() {
		return invalid("committed proposed version")
	}
	if known {
		p, _ := d.ProposedVersion.Value()
		if !p.Present() {
			return invalid("committed document absent")
		}
	}
	records := make([]RecoveryRecord, len(d.Recovery))
	for i, r := range d.Recovery {
		records[i] = r.data.Record
	}
	if err := validateRecoveryReferences(d.Effects, records); err != nil {
		return err
	}
	storageEffect := false
	for _, f := range d.Effects.data.Facets {
		if f.data.Facet == Storage {
			storageEffect = true
			if known && f.data.State != AppliedVerified {
				return invalid("known storage effect")
			}
			if d.Outcome == NotCommitted && f.data.State != NotStarted && f.data.State != VerifiedNoTargetChange {
				return invalid("uncommitted storage effect")
			}
		}
	}
	if !storageEffect {
		return invalid("missing storage effect")
	}
	if a, p := d.ProposedVersion.Value(); p {
		if b, q := d.CurrentVersion.Value(); q && (a.Family() != b.Family() || a.Store() != b.Store()) {
			return invalid("commit version scope")
		}
	}
	return nil
}
func validateNormalizedRecovery(e EffectReport, r []NormalizedRecovery) error {
	records := make([]RecoveryRecord, len(r))
	seen := map[RecoveryID]NormalizedRecovery{}
	for i, v := range r {
		records[i] = v.data.Record
		id := v.data.Record.data.RecoveryID
		if prior, p := seen[id]; p && prior.data != v.data {
			return invalid("inconsistent normalized recovery")
		}
		seen[id] = v
	}
	return validateRecoveryReferences(e, records)
}

// NormalizeRecovery preserves a stable shared ID and complete typed Storage
// detail. Equal repeated records coalesce; inconsistent duplicates refuse.
// This pure normalization executes no recovery or workflow.
func NormalizeRecovery(records []RecoveryRecord, storage []StorageRecovery) ([]NormalizedRecovery, error) {
	result := []NormalizedRecovery{}
	indices := map[RecoveryID]int{}
	for _, r := range records {
		if !r.Valid() {
			return nil, invalid("recovery record")
		}
		id := r.data.RecoveryID
		if i, p := indices[id]; p {
			if !sameRecoveryRecord(result[i].data.Record, r) {
				return nil, invalid("duplicate recovery")
			}
			continue
		}
		v, e := NewNormalizedRecovery(NormalizedRecoveryData{Record: r})
		if e != nil {
			return nil, e
		}
		indices[id] = len(result)
		result = append(result, v)
	}
	for _, r := range storage {
		if !r.Valid() {
			return nil, invalid("storage recovery")
		}
		id := r.data.Record.data.RecoveryID
		v, e := NewNormalizedRecovery(NormalizedRecoveryData{Record: r.data.Record, StorageDetail: Some(r)})
		if e != nil {
			return nil, e
		}
		if i, p := indices[id]; p {
			if !sameRecoveryRecord(result[i].data.Record, r.data.Record) {
				return nil, invalid("duplicate recovery")
			}
			if old, p := result[i].data.StorageDetail.Value(); p && old.data != r.data {
				return nil, invalid("duplicate storage detail")
			}
			result[i] = v
		} else {
			indices[id] = len(result)
			result = append(result, v)
		}
	}
	return cloneSlice(result), nil
}
