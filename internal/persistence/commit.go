package persistence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
)

var _ ports.Storage = (*Store)(nil)

func (s *Store) CommitUserConfig(ctx context.Context, request ports.UserConfigCommit) (api.StorageCommitResult, error) {
	return s.commit(ctx, request.Valid(), request.Expected(), document{family: api.UserConfig, user: request.Document()}, api.WorktreeScope{})
}
func (s *Store) CommitPreferences(ctx context.Context, request ports.PreferencesCommit) (api.StorageCommitResult, error) {
	return s.commit(ctx, request.Valid(), request.Expected(), document{family: api.Preferences, preferences: request.Document()}, api.WorktreeScope{})
}
func (s *Store) CommitRunConfig(ctx context.Context, request ports.RunConfigCommit) (api.StorageCommitResult, error) {
	return s.commit(ctx, request.Valid(), request.Expected(), document{family: api.RunConfig, run: request.Document()}, request.Scope())
}

// checkpoint is private fault instrumentation: there is no callback or native
// authority at the Storage port. Production stores have no hook. Tests set it
// once before concurrent use and can stop an actual process at each boundary.
func (s *Store) checkpoint(ctx context.Context, stage string) error {
	if s.hook != nil {
		if err := s.hook(stage); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (s *Store) acquireStore(ctx context.Context, family api.StorageFamily, scope api.WorktreeScope) (c *nativeChain, name, locator string, resultErr error) {
	if s == nil {
		return nil, "", "", errInvalidRequest
	}
	switch family {
	case api.UserConfig:
		c, resultErr = s.user.acquire(ctx)
		name, locator = s.user.basename, s.user.parentPath
	case api.Preferences:
		c, resultErr = s.preferences.acquire(ctx)
		name, locator = s.preferences.basename, s.preferences.parentPath
	case api.RunConfig:
		c, resultErr = acquireRun(ctx, scope)
		name, locator = "run.json", filepath.Join(scope.Data().RootLocator, ".gh-tree")
	default:
		resultErr = errInvalidRequest
	}
	if resultErr != nil {
		return c, name, locator, resultErr
	}
	if family == api.RunConfig {
		for _, binding := range []storeBinding{s.user, s.preferences} {
			other, err := binding.acquire(ctx)
			if err != nil {
				resultErr = err
				break
			}
			overlap, err := acquiredOverlap(c, name, other, binding.basename)
			err = errors.Join(err, other.close())
			if overlap {
				err = errors.Join(err, errInvalidRequest)
			}
			if err != nil {
				resultErr = err
				break
			}
		}
	}
	if resultErr != nil {
		resultErr = errors.Join(resultErr, c.close())
		c = nil
	}
	return c, name, locator, resultErr
}

// expectedAnchor resolves only already-acquired ancestors. It never decodes an
// opaque token into a pathname. An absence loaded before another writer created
// components can still match the same acquired anchor and literal remainder.
func (s *Store) expectedAnchor(c *nativeChain, name string, expected api.StorageVersion, scope api.WorktreeScope) (diskIdentity, []string, error) {
	start := len(c.guards) - 1
	var tail []string
	if expected.Family() == api.RunConfig {
		start = len(c.guards) - 2
		if len(c.remaining) != 0 {
			start++
		}
		tail = []string{".gh-tree"}
	} else {
		binding := s.user
		if expected.Family() == api.Preferences {
			binding = s.preferences
		}
		start = binding.anchorIndex
		tail = binding.remaining
	}
	for i := len(c.guards) - 1; i >= start; i-- {
		identity, err := nativeDirectoryIdentity(c.guards[i])
		if err != nil {
			return diskIdentity{}, nil, err
		}
		remaining := c.remaining
		if i < len(c.guards)-1 {
			offset := i - start
			if offset < 0 || offset >= len(tail) {
				continue
			}
			remaining = tail[offset:]
		}
		token, err := bindingToken(identity, remaining, name)
		if err != nil {
			return diskIdentity{}, nil, err
		}
		if token == expected.Store() {
			if len(remaining) != 0 && expected.Present() {
				return diskIdentity{}, nil, errInvalidRequest
			}
			return directoryRecord(identity), append([]string(nil), remaining...), nil
		}
	}
	return diskIdentity{}, nil, errors.Join(errInvalidRequest, errors.New("expected version belongs to another physical store"))
}

func (s *Store) commit(ctx context.Context, valid bool, expected api.StorageVersion, proposed document, scope api.WorktreeScope) (result api.StorageCommitResult, resultErr error) {
	r := api.StorageCommitResultData{Outcome: api.NotCommitted, Durability: api.DurabilityNotApplicable}
	stage := "validate"
	var c *nativeChain
	var lock *nativeStoreLock
	var owned []*nativeObject
	started := false
	var name, locator string
	effectiveScope := scope
	defer func() {
		if lock != nil && c != nil {
			observationContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, observed, err := loadAcquired(observationContext, c, proposed.family, effectiveScope, name)
			cancel()
			r.CurrentVersion = observed.Version
			if err != nil {
				r.Diagnostics = append(r.Diagnostics, storageDiagnostic("current-observation", err))
				resultErr = errors.Join(resultErr, err)
			}
		}
		for i := len(owned) - 1; i >= 0; i-- {
			resultErr = errors.Join(resultErr, owned[i].close())
		}
		if lock != nil {
			resultErr = errors.Join(resultErr, lock.close())
		}
		if c != nil {
			resultErr = errors.Join(resultErr, c.close())
		}
		r.CancellationAsked = ctx.Err() != nil
		state := api.NotStarted
		if started {
			state = api.VerifiedNoTargetChange
		}
		if r.PublicationKnown {
			state = api.AppliedVerified
		}
		if r.Outcome == api.StorageIndeterminate {
			state = api.EffectIndeterminate
		}
		var ids []api.RecoveryID
		for _, recovery := range r.Recovery {
			ids = append(ids, recovery.Data().Record.Data().RecoveryID)
		}
		facet, err := api.NewFacetEffect(api.FacetEffectData{Facet: api.Storage, State: state, RecoveryIDs: ids})
		if err != nil {
			resultErr = errors.Join(resultErr, err)
		}
		effects, err := api.NewEffectReport(api.EffectReportData{Facets: []api.FacetEffect{facet}})
		if err != nil {
			resultErr = errors.Join(resultErr, err)
		}
		r.Effects = effects
		if resultErr != nil {
			diagnostic := storageDiagnostic(stage, resultErr)
			r.Diagnostics = append(r.Diagnostics, diagnostic)
			resultErr = errors.Join(diagnostic, resultErr)
		}
		result, err = api.NewStorageCommitResult(r)
		resultErr = errors.Join(resultErr, err)
	}()
	if s == nil || !valid || !expected.Valid() || expected.Family() != proposed.family {
		return result, errInvalidRequest
	}
	raw, err := proposed.encode()
	if err != nil {
		return result, errors.Join(errInvalidRequest, err)
	}
	if err := s.checkpoint(ctx, stage); err != nil {
		return result, err
	}
	c, name, locator, err = s.acquireStore(ctx, proposed.family, scope)
	if err != nil {
		return result, err
	}
	started = true
	anchor, remainder, err := s.expectedAnchor(c, name, expected, scope)
	if err != nil {
		return result, err
	}
	// Birth-less identities are retained by reads but do not silently receive
	// restart incarnation guarantees. This write profile needs separate native
	// proof; changing a store token to ignore ctime would weaken its authority.
	if strings.HasPrefix(anchor.Stamp, "change:") {
		return result, errors.Join(errUnsupportedProfile, errors.New("unproved no-birth recovery incarnation profile"))
	}
	stage = "parents"
	for len(c.remaining) != 0 {
		if err := s.checkpoint(ctx, stage); err != nil {
			return result, err
		}
		if err := nativeRevalidate(ctx, c); err != nil {
			return result, err
		}
		part := c.remaining[0]
		child, createErr := nativeCreateDirectory(c.parent(), part, proposed.family != api.RunConfig)
		if errors.Is(createErr, os.ErrExist) {
			// Adoption is independently checked; never remove/recreate a peer's
			// component or claim ownership from its matching literal name.
			child, createErr = nativeAdoptDirectory(c.parent(), part)
		}
		if createErr != nil {
			return result, createErr
		}
		parent := c.parent()
		nativeAppendCreated(c, child, part)
		if err := nativeDirectoryBarrier(parent); err != nil {
			return result, err
		}
	}
	if proposed.family == api.RunConfig {
		v := scope.Data()
		v.RootIdentity, err = nativeDirectoryIdentityAs(c.guards[len(c.guards)-2], v.RootIdentity)
		if err != nil {
			return result, err
		}
		effectiveScope, err = api.NewWorktreeScope(v)
		if err != nil {
			return result, err
		}
	}
	stage = "lock"
	lock, err = nativeLockForStore(ctx, c.parent(), name, s.options.LockWait, proposed.family != api.RunConfig)
	if err != nil {
		return result, err
	}
	if err := s.checkpoint(ctx, stage); err != nil {
		return result, err
	}
	if err := nativeRevalidate(ctx, c); err != nil {
		return result, err
	}
	parentID, err := nativeDirectoryIdentity(c.parent())
	if err != nil {
		return result, err
	}
	if strings.HasPrefix(parentID.Stamp(), "change:") {
		return result, errUnsupportedProfile
	}
	store, err := bindingToken(parentID, nil, name)
	if err != nil {
		return result, err
	}
	current, observation, err := loadAcquired(ctx, c, proposed.family, effectiveScope, name)
	r.CurrentVersion = observation.Version
	if err != nil {
		return result, err
	}
	version, ok := observation.Version.Value()
	if !ok || current == nil {
		return result, errInvalidRequest
	}
	if expected.Store() == store {
		if versionRecord(expected) != versionRecord(version) {
			return result, errBindingChanged
		}
	} else if expected.Present() || version.Present() || len(remainder) == 0 {
		return result, errBindingChanged
	}
	if err := verifyRetention(*current, proposed); err != nil {
		return result, errors.Join(errInvalidRequest, err)
	}
	proposedVersion, err := contentVersion(proposed.family, store, effectiveScope, true, raw)
	if err != nil {
		return result, err
	}
	r.ProposedVersion = api.Some(proposedVersion)
	stage = "admission"
	names, records, retainedBytes, err := inventoryRecovery(ctx, c.parent(), name, s.options.RecoveryMaxRecords, s.options.RecoveryMaxBytes)
	inventoryErr := err
	for _, entry := range names {
		if !manifestName(entry) {
			continue
		}
		retained, observeErr := observeManifest(ctx, c, entry, name, locator, proposed.family, effectiveScope)
		r.Recovery = append(r.Recovery, retained...)
		if observeErr != nil {
			if notices, only := recoveryNotices(observeErr); only {
				r.Diagnostics = append(r.Diagnostics, notices...)
			} else {
				inventoryErr = errors.Join(inventoryErr, observeErr)
			}
		}
	}
	if inventoryErr != nil {
		return result, inventoryErr
	}
	reserve := int64(maxManifestJournalBytes + 2*len(raw) + 2*len(current.raw))
	if records >= s.options.RecoveryMaxRecords || reserve > s.options.RecoveryMaxBytes-retainedBytes {
		return result, errRecoveryCapacity
	}
	if err := nativeInspectDirectory(c.parent()); err != nil {
		return result, err
	}
	var original *nativeObject
	var metadata nativeMetadata
	if version.Present() {
		original, err = nativeOpenOriginal(c.parent(), name)
		if err != nil {
			return result, err
		}
		owned = append(owned, original)
		observed, err := nativeRead(ctx, original)
		if err != nil {
			return result, err
		}
		if !bytes.Equal(observed, []byte(current.raw)) {
			return result, errBindingChanged
		}
		metadata, err = nativeInspectMetadata(original)
		if err != nil {
			return result, err
		}
	}
	nonce, err := operationNonce()
	if err != nil {
		return result, err
	}
	m := recoveryManifest{Preparing: true, SchemaVersion: 1, Nonce: nonce, Family: proposed.family, Basename: name, Parent: directoryRecord(parentID), Scope: runScopeRecord(effectiveScope), ExpectedScope: runScopeRecord(scope), Expected: versionRecord(expected), ExpectedAnchor: anchor, ExpectedRemaining: remainder, Original: versionRecord(version), Proposed: versionRecord(proposedVersion)}
	for _, kind := range []api.StorageRecoveryKind{api.Manifest, api.RetainedPayload, api.RawOriginal, api.RetainedOriginal, api.RetainedPayload} {
		if !version.Present() && (kind == api.RawOriginal || kind == api.RetainedOriginal) {
			continue
		}
		artifactName := m.artifactName(kind)
		if kind == api.RetainedPayload && len(m.Artifacts) > 1 {
			artifactName = m.publicationName()
		}
		id, err := operationNonce()
		if err != nil {
			return result, err
		}
		var content []byte
		if kind == api.RetainedPayload {
			content = raw
		}
		if kind == api.RawOriginal {
			content = []byte(current.raw)
		}
		m.Artifacts = append(m.Artifacts, diskArtifact{kind, artifactName, id, diskIdentity{}, uint64(len(content)), sha256.Sum256(content)})
	}
	stage = "prepare.manifest"
	if err := s.checkpoint(ctx, stage); err != nil {
		return result, err
	}
	manifest, err := nativeCreateFile(c.parent(), m.artifactName(api.Manifest), proposed.family != api.RunConfig)
	if err != nil {
		return result, err
	}
	owned = append(owned, manifest)
	m.Artifacts[0].Identity, err = nativeArtifactIdentity(manifest)
	if err != nil {
		return result, err
	}
	if err := m.validate(proposed.family, effectiveScope, name, directoryRecord(parentID)); err != nil {
		return result, err
	}
	journal := manifestJournal{object: manifest}
	if err := journal.append(ctx, m); err != nil {
		return result, err
	}
	manifestRecovery, err := m.recovery(m.Artifacts[0], effectiveScope, locator)
	if err != nil {
		return result, err
	}
	r.Recovery = append(r.Recovery, manifestRecovery)
	var payload *nativeObject
	for i := 1; i < len(m.Artifacts); i++ {
		artifact := &m.Artifacts[i]
		stage = "prepare" + artifactSuffix(artifact.Kind)
		if artifact.Name == m.publicationName() {
			stage = "prepare.publication"
		}
		if err := s.checkpoint(ctx, stage); err != nil {
			return result, err
		}
		var object *nativeObject
		if artifact.Kind == api.RetainedOriginal {
			object, err = nativeRetainOriginal(original, c.parent(), name, artifact.Name)
		} else if artifact.Name == m.publicationName() {
			object, err = nativeRetainOriginal(payload, c.parent(), m.artifactName(api.RetainedPayload), artifact.Name)
		} else {
			var initial *nativeMetadata
			if original != nil {
				initial = &metadata
			}
			object, err = nativeCreateFileMetadata(c.parent(), artifact.Name, proposed.family != api.RunConfig, initial)
		}
		if err != nil {
			return result, err
		}
		owned = append(owned, object)
		if artifact.Name == m.artifactName(api.RetainedPayload) {
			payload = object
		}
		if artifact.Kind != api.RetainedOriginal && artifact.Name != m.publicationName() {
			if original != nil {
				if err := nativeApplyMetadata(object, metadata); err != nil {
					return result, err
				}
			}
			content := raw
			if artifact.Kind == api.RawOriginal {
				content = []byte(current.raw)
			}
			if err := s.checkpoint(ctx, stage+".write"); err != nil {
				return result, err
			}
			if err := writeComplete(ctx, object.file, content); err != nil {
				return result, err
			}
			// Writing can clear set-ID bits on Unix. Initial metadata protects
			// preparation confidentiality; exact metadata is established again
			// after all data writes and independently verified before flushing.
			if original != nil {
				if err := nativeApplyMetadata(object, metadata); err != nil {
					return result, err
				}
			}
			if _, err := nativeInspectMetadata(object); err != nil {
				return result, err
			}
			if err := s.checkpoint(ctx, stage+".flush"); err != nil {
				return result, err
			}
			if err := object.file.Sync(); err != nil {
				return result, err
			}
			verified, err := nativeRead(ctx, object)
			if err != nil || !bytes.Equal(verified, content) {
				return result, errors.Join(err, errors.New("prepared bytes differ"))
			}
		}
		artifact.Identity, err = nativeArtifactIdentity(object)
		if err != nil {
			return result, err
		}
		if err := s.checkpoint(ctx, stage+".journal"); err != nil {
			return result, err
		}
		if err := journal.append(ctx, m); err != nil {
			return result, err
		}
		recovery, err := m.recovery(*artifact, effectiveScope, locator)
		if err != nil {
			return result, err
		}
		r.Recovery = append(r.Recovery, recovery)
	}
	publisher, err := nativeOpenOriginal(c.parent(), m.publicationName())
	if err != nil {
		return result, err
	}
	owned = append(owned, publisher)
	publicationIdentity, err := nativeArtifactIdentity(publisher)
	if err != nil {
		return result, err
	}
	payloadPolicy, err := nativeInspectMetadata(publisher)
	if err != nil {
		return result, err
	}
	m.Preparing = false
	if err := m.validate(proposed.family, effectiveScope, name, directoryRecord(parentID)); err != nil {
		return result, err
	}
	if err := journal.append(ctx, m); err != nil {
		return result, err
	}
	stage = "manifest-flushed"
	if err := nativeDirectoryBarrier(c.parent()); err != nil {
		return result, err
	}
	if err := s.checkpoint(ctx, stage); err != nil {
		return result, err
	}
	// Close every preparation handle except the payload and exact original;
	// their retained handles are needed for final identity checks/publication.
	for i := len(owned) - 1; i >= 0; i-- {
		if owned[i] == publisher || owned[i] == original {
			continue
		}
		object := owned[i]
		owned = append(owned[:i], owned[i+1:]...)
		if err := object.close(); err != nil {
			return result, err
		}
	}
	stage = "final-check"
	if err := s.checkpoint(ctx, stage); err != nil {
		return result, err
	}
	if err := nativeRevalidate(ctx, c); err != nil {
		return result, err
	}
	if err := verifyPermanentLock(c.parent(), name, lock); err != nil {
		return result, err
	}
	if proposed.family == api.RunConfig && !nativeMatchesDirectory(c.guards[len(c.guards)-2], effectiveScope.Data().RootIdentity) {
		return result, errBindingChanged
	}
	if err := verifyCurrent(ctx, c.parent(), name, original, []byte(current.raw), metadata); err != nil {
		return result, err
	}
	preparedID, err := nativeArtifactIdentity(publisher)
	if err != nil || preparedID != publicationIdentity {
		return result, errors.Join(err, errBindingChanged)
	}
	if err := verifyNativeEntry(c.parent(), m.publicationName(), publisher); err != nil {
		return result, err
	}
	if err := verifyNativeEntry(c.parent(), m.artifactName(api.RetainedPayload), publisher); err != nil {
		return result, err
	}
	preparedRaw, err := nativeRead(ctx, publisher)
	if err != nil || !bytes.Equal(preparedRaw, raw) {
		return result, errors.Join(err, errBindingChanged)
	}
	preparedPolicy, err := nativeInspectMetadata(publisher)
	if err != nil || !payloadPolicy.equal(preparedPolicy) {
		return result, errors.Join(err, errBindingChanged)
	}
	if err := s.checkpoint(ctx, "before-publication"); err != nil {
		return result, err
	}
	stage = "publication"
	if err := nativePublish(publisher, c.parent(), m.publicationName(), name, version.Present()); err != nil {
		return result, err
	}
	r.Outcome, r.PublicationKnown, r.Durability = api.CommittedDurabilityUncertain, true, api.DurabilityUncertain
	// Losing delivery of the native return is an explicit test seam. Production
	// reaches this point only after the actual selected native call succeeded.
	var lostReturnErr error
	if s.hook != nil {
		if err := s.hook("native-return-lost"); err != nil {
			r.Outcome, r.PublicationKnown = api.StorageIndeterminate, false
			lostReturnErr = err
		}
	}
	stage = "directory-flush"
	barrierErr := s.checkpoint(context.WithoutCancel(ctx), stage)
	if barrierErr == nil {
		barrierErr = nativeDirectoryBarrier(c.parent())
	}
	if r.PublicationKnown && barrierErr == nil && nativePublicationDurability() == api.SupportedCrashBarrierComplete {
		r.Outcome, r.Durability = api.Committed, api.SupportedCrashBarrierComplete
	}
	// Outcome and current observation are independent. A later editor can make
	// current differ from proposed without erasing this known publication.
	resultErr = errors.Join(lostReturnErr, barrierErr)
	stage = "outcome-delivery"
	if s.hook != nil {
		resultErr = errors.Join(resultErr, s.hook(stage))
	}
	return result, resultErr
}

func verifyNativeEntry(parent *nativeObject, name string, expected *nativeObject) (resultErr error) {
	entry, err := nativeOpenDocument(parent, name)
	if err != nil {
		return errors.Join(err, errBindingChanged)
	}
	defer func() { resultErr = errors.Join(resultErr, entry.close()) }()
	actualID, err := nativeArtifactIdentity(entry)
	if err != nil {
		return err
	}
	expectedID, err := nativeArtifactIdentity(expected)
	if err != nil {
		return err
	}
	if actualID != expectedID {
		return errors.Join(errBindingChanged, errors.New("publication source entry no longer names the retained payload"))
	}
	return nil
}

func verifyPermanentLock(parent *nativeObject, basename string, lock *nativeStoreLock) (resultErr error) {
	object, err := nativeOpenDocument(parent, basename+".lock")
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, object.close()) }()
	if !nativeSameObject(object, lock.object) {
		return errBindingChanged
	}
	links, err := nativeLinkCount(object)
	if err != nil {
		return err
	}
	if links != 1 {
		return errors.Join(errBindingChanged, errors.New("permanent lock acquired additional links"))
	}
	return nil
}

func verifyCurrent(ctx context.Context, parent *nativeObject, name string, original *nativeObject, raw []byte, metadata nativeMetadata) (resultErr error) {
	current, err := nativeOpenDocument(parent, name)
	if original == nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if current != nil {
			err = errors.Join(err, current.close())
		}
		return errors.Join(err, errBindingChanged)
	}
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, current.close()) }()
	originalID, err := nativeArtifactIdentity(original)
	if err != nil {
		return err
	}
	currentID, err := nativeArtifactIdentity(current)
	if err != nil {
		return err
	}
	if originalID != currentID {
		return errBindingChanged
	}
	content, err := nativeRead(ctx, current)
	if err != nil {
		return err
	}
	if !bytes.Equal(content, raw) {
		return errBindingChanged
	}
	policy, err := nativeInspectMetadata(current)
	if err != nil {
		return err
	}
	if !metadata.equal(policy) {
		return errBindingChanged
	}
	return nil
}
