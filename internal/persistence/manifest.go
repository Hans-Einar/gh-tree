package persistence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

const maxManifestBytes = 64 * 1024

// Disk records are private facts. No field is interpreted as an ambient path or
// executable instruction. Every basename is independently derived and every
// native object is reacquired before observations cross the port.
type diskIdentity struct {
	Platform api.DirectoryPlatform
	Device   uint64
	File     [16]byte
	Stamp    string
}

func directoryRecord(v api.DirectoryIdentity) diskIdentity {
	return diskIdentity{v.Platform(), v.Device(), v.FileID(), v.Stamp()}
}
func (v diskIdentity) directory() (api.DirectoryIdentity, error) {
	return api.NewDirectoryIdentity(v.Platform, v.Device, v.File, v.Stamp)
}
func (v diskIdentity) valid() bool {
	_, err := v.directory()
	return err == nil
}

type diskVersion struct {
	Store   string
	Present bool
	Length  uint64
	Digest  [32]byte
}

func versionRecord(v api.StorageVersion) diskVersion {
	return diskVersion{v.Store(), v.Present(), v.ByteLength(), v.SHA256()}
}
func (v diskVersion) version(family api.StorageFamily, scope api.WorktreeScope) (api.StorageVersion, error) {
	if family == api.RunConfig {
		return api.NewRunStorageVersion(scope, v.Store, v.Present, v.Length, v.Digest)
	}
	return api.NewStorageVersion(family, v.Store, v.Present, v.Length, v.Digest)
}

type diskRunScope struct {
	RepositoryToken   []byte
	AdministrativeKey string
	Root              diskIdentity
}

func runScopeRecord(scope api.WorktreeScope) diskRunScope {
	if !scope.Valid() {
		return diskRunScope{}
	}
	v := scope.Data()
	return diskRunScope{[]byte(v.ID.Repository().Token()), v.ID.AdministrativeKey(), directoryRecord(v.RootIdentity)}
}
func (v diskRunScope) matches(scope api.WorktreeScope) bool {
	other := runScopeRecord(scope)
	return bytes.Equal(v.RepositoryToken, other.RepositoryToken) && v.AdministrativeKey == other.AdministrativeKey && v.Root == other.Root
}

type diskArtifact struct {
	Kind     api.StorageRecoveryKind
	Name     string
	ID       string
	Identity diskIdentity
	Length   uint64
	Digest   [32]byte
}

type recoveryManifest struct {
	Preparing         bool `json:",omitempty"`
	SchemaVersion     uint32
	Nonce             string
	Family            api.StorageFamily
	Basename          string
	Parent            diskIdentity
	Scope             diskRunScope
	ExpectedScope     diskRunScope
	Expected          diskVersion
	ExpectedAnchor    diskIdentity
	ExpectedRemaining []string
	Original          diskVersion
	Proposed          diskVersion
	Artifacts         []diskArtifact
}

func artifactSuffix(kind api.StorageRecoveryKind) string {
	switch kind {
	case api.Manifest:
		return ".manifest"
	case api.RawOriginal:
		return ".raw"
	case api.RetainedOriginal:
		return ".original"
	case api.RetainedPayload:
		return ".payload"
	}
	return ""
}
func (m recoveryManifest) artifactName(kind api.StorageRecoveryKind) string {
	return recoveryPrefix(m.Basename) + m.Nonce + artifactSuffix(kind)
}
func (m recoveryManifest) publicationName() string {
	return recoveryPrefix(m.Basename) + m.Nonce + ".publication"
}

// Validation checks complete request/family/root/parent/absence transition and
// each derived name/ID. Current document equality is not publication evidence.
func (m recoveryManifest) validate(family api.StorageFamily, scope api.WorktreeScope, basename string, parent diskIdentity) error {
	nonce, err := hex.DecodeString(m.Nonce)
	if err != nil || len(nonce) != 32 || hex.EncodeToString(nonce) != m.Nonce || m.SchemaVersion != 1 || m.Family != family || !nativeSameName(m.Basename, basename) || m.Parent != parent || !parent.valid() {
		return errors.New("recovery manifest subject or native parent mismatch")
	}
	if !m.Scope.matches(scope) {
		return errors.New("recovery manifest worktree/root mismatch")
	}
	parentValue, err := parent.directory()
	if err != nil {
		return err
	}
	store, err := bindingToken(parentValue, nil, basename)
	if err != nil || m.Original.Store != store || m.Proposed.Store != store || !m.Proposed.Present || m.Proposed.Length > api.MaxDocumentBytes {
		return errors.New("recovery manifest document binding mismatch")
	}
	anchor, err := m.ExpectedAnchor.directory()
	if err != nil {
		return err
	}
	expectedStore, err := bindingToken(anchor, m.ExpectedRemaining, basename)
	if err != nil || m.Expected.Store != expectedStore {
		return errors.New("recovery manifest expected anchor mismatch")
	}
	for _, v := range []diskVersion{m.Expected, m.Original, m.Proposed} {
		if _, err := v.version(family, scope); err != nil {
			return err
		}
	}
	if m.Expected.Store == m.Original.Store {
		if m.Expected != m.Original || m.ExpectedAnchor != m.Parent || len(m.ExpectedRemaining) != 0 {
			return errors.New("recovery manifest original expected mismatch")
		}
	} else if m.Expected.Present || m.Original.Present || len(m.ExpectedRemaining) == 0 {
		return errors.New("recovery manifest invalid missing-parent transition")
	}
	seenKinds := map[api.StorageRecoveryKind]int{}
	seenNames := map[string]bool{}
	seenIDs := map[string]bool{}
	for _, artifact := range m.Artifacts {
		validName := artifact.Name == m.artifactName(artifact.Kind) || artifact.Kind == api.RetainedPayload && artifact.Name == m.publicationName()
		identityValid := artifact.Identity.valid() || m.Preparing && artifact.Identity == (diskIdentity{})
		if !artifact.Kind.Valid() || seenNames[artifact.Name] || seenIDs[artifact.ID] || !validName || !singleName(artifact.Name) || !identityValid {
			return errors.New("recovery artifact name, identity or uniqueness mismatch")
		}
		// IDs are allocated separately, then persisted; no locator-derived IDs.
		id, err := hex.DecodeString(artifact.ID)
		if err != nil || len(id) != 32 || hex.EncodeToString(id) != artifact.ID {
			return errors.New("invalid persisted recovery ID")
		}
		seenKinds[artifact.Kind]++
		seenIDs[artifact.ID], seenNames[artifact.Name] = true, true
		if artifact.Kind == api.RawOriginal && (artifact.Length != m.Original.Length || artifact.Digest != m.Original.Digest) ||
			artifact.Kind == api.RetainedPayload && (artifact.Length != m.Proposed.Length || artifact.Digest != m.Proposed.Digest) {
			return errors.New("recovery artifact byte binding mismatch")
		}
	}
	if seenKinds[api.Manifest] != 1 || seenKinds[api.RetainedPayload] < 1 || seenKinds[api.RetainedPayload] > 2 || seenKinds[api.RawOriginal] != boolInt(m.Original.Present) || seenKinds[api.RetainedOriginal] != boolInt(m.Original.Present) || len(m.Artifacts) != 1+seenKinds[api.RetainedPayload]+2*boolInt(m.Original.Present) {
		return errors.New("incomplete recovery artifact set")
	}
	return nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func decodeManifest(raw []byte) (recoveryManifest, error) {
	if bytes.HasPrefix(raw, []byte(manifestJournalMagic)) {
		return decodeManifestJournal(raw)
	}
	var m recoveryManifest
	if len(raw) > maxManifestBytes {
		return m, errors.New("recovery manifest size limit")
	}
	if _, err := api.NewOpaqueJSON(raw); err != nil {
		return m, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&m); err != nil {
		return m, err
	}
	canonical, err := json.Marshal(m)
	if err != nil || !sameJSONValue(raw, canonical) {
		return m, errors.New("manifest requires exact field names and complete shape")
	}
	return m, nil
}

func (m recoveryManifest) recovery(artifact diskArtifact, scope api.WorktreeScope, parentLocator string) (api.StorageRecovery, error) {
	id, err := api.NewRecoveryID("persistence:" + artifact.ID)
	if err != nil {
		return api.StorageRecovery{}, err
	}
	subjectData := api.RecoverySubjectData{Store: api.Some(m.Original.Store), Family: api.Some(m.Family)}
	if m.Family == api.RunConfig {
		subjectData.Worktree = api.Some(scope.Data().ID)
	}
	subject, err := api.NewRecoverySubject(subjectData)
	if err != nil {
		return api.StorageRecovery{}, err
	}
	original, err := m.Original.version(m.Family, scope)
	if err != nil {
		return api.StorageRecovery{}, err
	}
	proposed, err := m.Proposed.version(m.Family, scope)
	if err != nil {
		return api.StorageRecovery{}, err
	}
	oldVersion, err := api.NewStorageRecoveryVersion(api.StorageRecoveryVersionData{Version: original})
	if err != nil {
		return api.StorageRecovery{}, err
	}
	newVersion, err := api.NewStorageRecoveryVersion(api.StorageRecoveryVersionData{Version: proposed})
	if err != nil {
		return api.StorageRecovery{}, err
	}
	kinds := map[api.StorageRecoveryKind]api.RecoveryKind{api.Manifest: api.RecoveryManifest, api.RawOriginal: api.RecoveryRawOriginal, api.RetainedOriginal: api.RecoveryRetainedOriginal, api.RetainedPayload: api.RecoveryRetainedPayload}
	locator := filepath.Join(parentLocator, artifact.Name) // display only, never native authority
	record, err := api.NewRecoveryRecord(api.RecoveryRecordData{
		RecoveryID: id, Kind: kinds[artifact.Kind], Layer: api.LayerPersistence, Subject: subject, Locator: locator,
		Original: api.Some[api.RecoveryVersion](oldVersion), Proposed: api.Some[api.RecoveryVersion](newVersion),
		NextAction: "Inspect retained facts and current document before any explicit recovery; publication attribution is unknown after restart.",
	})
	if err != nil {
		return api.StorageRecovery{}, err
	}
	raw, err := json.Marshal(artifact)
	if err != nil {
		return api.StorageRecovery{}, err
	}
	digest := sha256.Sum256(raw)
	identity, err := api.NewSourceVersion("persistence-artifact", m.Original.Store, m.Nonce, hex.EncodeToString(digest[:]))
	if err != nil {
		return api.StorageRecovery{}, err
	}
	return api.NewStorageRecovery(api.StorageRecoveryData{Record: record, Family: m.Family, Locator: locator, Kind: artifact.Kind, Identity: identity})
}

func observeManifest(ctx context.Context, parent *nativeObject, name, basename, locator string, family api.StorageFamily, scope api.WorktreeScope) (out []api.StorageRecovery, resultErr error) {
	object, err := nativeOpenDocument(parent, name)
	if err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, object.close()) }()
	raw, err := nativeRead(ctx, object)
	if err != nil {
		return nil, err
	}
	m, err := decodeManifest(raw)
	if err != nil && m.SchemaVersion == 0 {
		return nil, err
	}
	resultErr = errors.Join(resultErr, err)
	if m.Preparing {
		resultErr = errors.Join(resultErr, errIncompletePreparation)
	}
	parentID, err := nativeDirectoryIdentity(parent)
	if err != nil {
		return nil, err
	}
	if err := m.validate(family, scope, basename, directoryRecord(parentID)); err != nil || name != m.artifactName(api.Manifest) {
		return nil, errors.Join(err, errors.New("manifest native name/subject mismatch"))
	}
	manifestIdentity, err := nativeArtifactIdentity(object)
	if err != nil {
		return nil, err
	}
	for _, artifact := range m.Artifacts {
		if artifact.Kind == api.Manifest && manifestIdentity != artifact.Identity {
			return nil, errors.New("read manifest native identity differs from recorded self")
		}
	}
	for _, artifact := range m.Artifacts {
		if artifact.Identity == (diskIdentity{}) {
			continue
		}
		observed, err := nativeOpenDocument(parent, artifact.Name)
		if err != nil {
			// A renamed payload may no longer have its preparation name. This
			// is an independent absence; never relabel the current target as the
			// retained artifact or infer publication from matching target bytes.
			if !(artifact.Kind == api.RetainedPayload && artifact.Name == m.publicationName() && errors.Is(err, os.ErrNotExist)) {
				resultErr = errors.Join(resultErr, err)
			}
			continue
		}
		identity, err := nativeArtifactIdentity(observed)
		if err == nil && identity != artifact.Identity {
			err = errors.New("recovery artifact native identity changed")
		}
		if err == nil && (artifact.Kind == api.RawOriginal || artifact.Kind == api.RetainedPayload) {
			content, readErr := nativeRead(ctx, observed)
			err = readErr
			if err == nil && (uint64(len(content)) != artifact.Length || sha256.Sum256(content) != artifact.Digest) {
				err = errors.New("recovery artifact content changed")
			}
		}
		err = errors.Join(err, observed.close())
		if err != nil {
			resultErr = errors.Join(resultErr, err)
			continue
		}
		recovery, err := m.recovery(artifact, scope, locator)
		if err != nil {
			resultErr = errors.Join(resultErr, err)
			continue
		}
		out = append(out, recovery)
	}
	return out, resultErr
}

func manifestName(name string) bool { return strings.HasSuffix(name, ".manifest") }
