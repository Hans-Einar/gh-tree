package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// UserConfigDocument is an immutable semantic boundary value. Its zero is invalid.
type UserConfigDocument struct {
	data        UserConfigDocumentData
	initialized bool
}

// UserConfigDocumentData is a construction/copy value; NewUserConfigDocument validates and owns a copy.
type UserConfigDocumentData struct {
	SchemaVersion  uint32
	StripPrefixes  StoredField[[]string]
	LegacyRepos    StoredField[[]LegacyRepositoryConfig]
	ScopedRepos    StoredField[[]ScopedRepositoryConfig]
	UnknownMembers JSONMembers
}

func NewUserConfigDocument(d UserConfigDocumentData) (UserConfigDocument, error) {
	if err := d.validate(); err != nil {
		return UserConfigDocument{}, err
	}
	return UserConfigDocument{data: d.clone(), initialized: true}, nil
}
func (v UserConfigDocument) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v UserConfigDocument) Data() UserConfigDocumentData { return v.data.clone() }
func (v UserConfigDocument) Clone() UserConfigDocument    { return v }
func (d UserConfigDocumentData) clone() UserConfigDocumentData {
	d.StripPrefixes = cloneStoredSlice(d.StripPrefixes)
	d.LegacyRepos = cloneStoredSlice(d.LegacyRepos)
	d.ScopedRepos = cloneStoredSlice(d.ScopedRepos)
	return d
}
func (d UserConfigDocumentData) validate() error {

	if !d.StripPrefixes.Valid() {
		return invalid("field presence")
	}
	if !d.LegacyRepos.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.LegacyRepos.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.ScopedRepos.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.ScopedRepos.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.LegacyRepos.Presence() == FieldNull {
		return invalid("null LegacyRepos")
	}
	if d.ScopedRepos.Presence() == FieldNull {
		return invalid("null ScopedRepos")
	}
	if d.SchemaVersion > 1 {
		return invalid("schema version")
	}
	return validateUserDocument(d)
}

// LegacyRepositoryConfig is an immutable semantic boundary value. Its zero is invalid.
type LegacyRepositoryConfig struct {
	data        LegacyRepositoryConfigData
	initialized bool
}

// LegacyRepositoryConfigData is a construction/copy value; NewLegacyRepositoryConfig validates and owns a copy.
type LegacyRepositoryConfigData struct {
	Key   string
	Value ConfiguredRepository
}

func NewLegacyRepositoryConfig(d LegacyRepositoryConfigData) (LegacyRepositoryConfig, error) {
	if err := d.validate(); err != nil {
		return LegacyRepositoryConfig{}, err
	}
	return LegacyRepositoryConfig{data: d.clone(), initialized: true}, nil
}
func (v LegacyRepositoryConfig) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v LegacyRepositoryConfig) Data() LegacyRepositoryConfigData { return v.data.clone() }
func (v LegacyRepositoryConfig) Clone() LegacyRepositoryConfig    { return v }
func (d LegacyRepositoryConfigData) clone() LegacyRepositoryConfigData {

	return d
}
func (d LegacyRepositoryConfigData) validate() error {

	if !d.Value.Valid() {
		return invalid("d.Value")
	}
	if !storageString(d.Key) || d.Key == "" {
		return invalid("legacy key")
	}
	return nil
}

// ScopedRepositoryConfig is an immutable semantic boundary value. Its zero is invalid.
type ScopedRepositoryConfig struct {
	data        ScopedRepositoryConfigData
	initialized bool
}

// ScopedRepositoryConfigData is a construction/copy value; NewScopedRepositoryConfig validates and owns a copy.
type ScopedRepositoryConfigData struct {
	RepositoryID domain.RepositoryID
	Value        ConfiguredRepository
}

func NewScopedRepositoryConfig(d ScopedRepositoryConfigData) (ScopedRepositoryConfig, error) {
	if err := d.validate(); err != nil {
		return ScopedRepositoryConfig{}, err
	}
	return ScopedRepositoryConfig{data: d.clone(), initialized: true}, nil
}
func (v ScopedRepositoryConfig) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v ScopedRepositoryConfig) Data() ScopedRepositoryConfigData { return v.data.clone() }
func (v ScopedRepositoryConfig) Clone() ScopedRepositoryConfig    { return v }
func (d ScopedRepositoryConfigData) clone() ScopedRepositoryConfigData {

	return d
}
func (d ScopedRepositoryConfigData) validate() error {
	if !d.RepositoryID.Valid() {
		return invalid("d.RepositoryID")
	}
	if !d.Value.Valid() {
		return invalid("d.Value")
	}
	if d.RepositoryID.Scope() != domain.Remote {
		return invalid("configured scope")
	}
	return nil
}

// ConfiguredRepository is an immutable semantic boundary value. Its zero is invalid.
type ConfiguredRepository struct {
	data        ConfiguredRepositoryData
	initialized bool
}

// ConfiguredRepositoryData is a construction/copy value; NewConfiguredRepository validates and owns a copy.
type ConfiguredRepositoryData struct {
	Worktrees      StoredField[[]ConfiguredWorktreeTarget]
	UnknownMembers JSONMembers
}

func NewConfiguredRepository(d ConfiguredRepositoryData) (ConfiguredRepository, error) {
	if err := d.validate(); err != nil {
		return ConfiguredRepository{}, err
	}
	return ConfiguredRepository{data: d.clone(), initialized: true}, nil
}
func (v ConfiguredRepository) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ConfiguredRepository) Data() ConfiguredRepositoryData { return v.data.clone() }
func (v ConfiguredRepository) Clone() ConfiguredRepository    { return v }
func (d ConfiguredRepositoryData) clone() ConfiguredRepositoryData {
	d.Worktrees = cloneStoredSlice(d.Worktrees)
	return d
}
func (d ConfiguredRepositoryData) validate() error {
	if !d.Worktrees.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.Worktrees.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.Worktrees.Presence() == FieldNull {
		return invalid("null Worktrees")
	}
	if !d.UnknownMembers.Excludes("worktrees") {
		return invalid("known member collision")
	}
	if x, ok := d.Worktrees.Value(); ok {
		seen := map[string]bool{}
		for _, v := range x {
			n := v.data.Name
			if seen[n] {
				return invalid("target duplicate")
			}
			seen[n] = true
		}
	}
	return nil
}

// ConfiguredWorktreeTarget is an immutable semantic boundary value. Its zero is invalid.
type ConfiguredWorktreeTarget struct {
	data        ConfiguredWorktreeTargetData
	initialized bool
}

// ConfiguredWorktreeTargetData is a construction/copy value; NewConfiguredWorktreeTarget validates and owns a copy.
type ConfiguredWorktreeTargetData struct {
	Name           string
	Path           StoredField[string]
	Branch         StoredField[string]
	UnknownMembers JSONMembers
}

func NewConfiguredWorktreeTarget(d ConfiguredWorktreeTargetData) (ConfiguredWorktreeTarget, error) {
	if err := d.validate(); err != nil {
		return ConfiguredWorktreeTarget{}, err
	}
	return ConfiguredWorktreeTarget{data: d.clone(), initialized: true}, nil
}
func (v ConfiguredWorktreeTarget) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v ConfiguredWorktreeTarget) Data() ConfiguredWorktreeTargetData { return v.data.clone() }
func (v ConfiguredWorktreeTarget) Clone() ConfiguredWorktreeTarget    { return v }
func (d ConfiguredWorktreeTargetData) clone() ConfiguredWorktreeTargetData {

	return d
}
func (d ConfiguredWorktreeTargetData) validate() error {

	if !d.Path.Valid() {
		return invalid("field presence")
	}
	if !d.Branch.Valid() {
		return invalid("field presence")
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.Path.Presence() == FieldNull {
		return invalid("null Path")
	}
	if d.Branch.Presence() == FieldNull {
		return invalid("null Branch")
	}
	if d.Name == "" || !storageString(d.Name) || !storedText(d.Path) || !storedText(d.Branch) || !d.UnknownMembers.Excludes("path", "branch") {
		return invalid("configured target")
	}
	return nil
}

// PreferencesDocument is an immutable semantic boundary value. Its zero is invalid.
type PreferencesDocument struct {
	data        PreferencesDocumentData
	initialized bool
}

// PreferencesDocumentData is a construction/copy value; NewPreferencesDocument validates and owns a copy.
type PreferencesDocumentData struct {
	SchemaVersion     uint32
	LegacyFolders     StoredField[[]LegacyStringPreference]
	LegacyWorktrees   StoredField[[]LegacyStringPreference]
	ScopedPreferences StoredField[[]ScopedPreference]
	UnknownMembers    JSONMembers
}

func NewPreferencesDocument(d PreferencesDocumentData) (PreferencesDocument, error) {
	if err := d.validate(); err != nil {
		return PreferencesDocument{}, err
	}
	return PreferencesDocument{data: d.clone(), initialized: true}, nil
}
func (v PreferencesDocument) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v PreferencesDocument) Data() PreferencesDocumentData { return v.data.clone() }
func (v PreferencesDocument) Clone() PreferencesDocument    { return v }
func (d PreferencesDocumentData) clone() PreferencesDocumentData {
	d.LegacyFolders = cloneStoredSlice(d.LegacyFolders)
	d.LegacyWorktrees = cloneStoredSlice(d.LegacyWorktrees)
	d.ScopedPreferences = cloneStoredSlice(d.ScopedPreferences)
	return d
}
func (d PreferencesDocumentData) validate() error {

	if !d.LegacyFolders.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.LegacyFolders.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.LegacyWorktrees.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.LegacyWorktrees.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.ScopedPreferences.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.ScopedPreferences.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.LegacyFolders.Presence() == FieldNull {
		return invalid("null LegacyFolders")
	}
	if d.LegacyWorktrees.Presence() == FieldNull {
		return invalid("null LegacyWorktrees")
	}
	if d.ScopedPreferences.Presence() == FieldNull {
		return invalid("null ScopedPreferences")
	}
	if d.SchemaVersion > 1 {
		return invalid("schema version")
	}
	return validatePreferencesDocument(d)
}

// LegacyStringPreference is an immutable semantic boundary value. Its zero is invalid.
type LegacyStringPreference struct {
	data        LegacyStringPreferenceData
	initialized bool
}

// LegacyStringPreferenceData is a construction/copy value; NewLegacyStringPreference validates and owns a copy.
type LegacyStringPreferenceData struct {
	Key   string
	Value string
}

func NewLegacyStringPreference(d LegacyStringPreferenceData) (LegacyStringPreference, error) {
	if err := d.validate(); err != nil {
		return LegacyStringPreference{}, err
	}
	return LegacyStringPreference{data: d.clone(), initialized: true}, nil
}
func (v LegacyStringPreference) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v LegacyStringPreference) Data() LegacyStringPreferenceData { return v.data.clone() }
func (v LegacyStringPreference) Clone() LegacyStringPreference    { return v }
func (d LegacyStringPreferenceData) clone() LegacyStringPreferenceData {

	return d
}
func (d LegacyStringPreferenceData) validate() error {

	if d.Key == "" || !storageString(d.Key) || !storageString(d.Value) {
		return invalid("legacy preference")
	}
	return nil
}

// ScopedPreference is an immutable semantic boundary value. Its zero is invalid.
type ScopedPreference struct {
	data        ScopedPreferenceData
	initialized bool
}

// ScopedPreferenceData is a construction/copy value; NewScopedPreference validates and owns a copy.
type ScopedPreferenceData struct {
	RepositoryID   domain.RepositoryID
	LastFolder     StoredField[string]
	ActiveWorktree StoredField[StoredActiveWorktree]
	UnknownMembers JSONMembers
}

func NewScopedPreference(d ScopedPreferenceData) (ScopedPreference, error) {
	if err := d.validate(); err != nil {
		return ScopedPreference{}, err
	}
	return ScopedPreference{data: d.clone(), initialized: true}, nil
}
func (v ScopedPreference) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v ScopedPreference) Data() ScopedPreferenceData { return v.data.clone() }
func (v ScopedPreference) Clone() ScopedPreference    { return v }
func (d ScopedPreferenceData) clone() ScopedPreferenceData {

	return d
}
func (d ScopedPreferenceData) validate() error {
	if !d.RepositoryID.Valid() {
		return invalid("d.RepositoryID")
	}
	if !d.LastFolder.Valid() {
		return invalid("field presence")
	}
	if !d.ActiveWorktree.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.ActiveWorktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.LastFolder.Presence() == FieldNull {
		return invalid("null LastFolder")
	}
	if d.ActiveWorktree.Presence() == FieldNull {
		return invalid("null ActiveWorktree")
	}
	if (d.RepositoryID.Scope() == domain.Remote && d.ActiveWorktree.Presence() != FieldAbsent) || (d.RepositoryID.Scope() == domain.LocalCommon && d.LastFolder.Presence() != FieldAbsent) || !storedText(d.LastFolder) || !d.UnknownMembers.Excludes("lastFolder", "activeWorktree") {
		return invalid("preference scope")
	}
	return nil
}

// StoredActiveWorktree is an immutable semantic boundary value. Its zero is invalid.
type StoredActiveWorktree struct {
	data        StoredActiveWorktreeData
	initialized bool
}

// StoredActiveWorktreeData is a construction/copy value; NewStoredActiveWorktree validates and owns a copy.
type StoredActiveWorktreeData struct {
	AdministrativeKey string
	LastKnownPath     string
	UnknownMembers    JSONMembers
}

func NewStoredActiveWorktree(d StoredActiveWorktreeData) (StoredActiveWorktree, error) {
	if err := d.validate(); err != nil {
		return StoredActiveWorktree{}, err
	}
	return StoredActiveWorktree{data: d.clone(), initialized: true}, nil
}
func (v StoredActiveWorktree) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v StoredActiveWorktree) Data() StoredActiveWorktreeData { return v.data.clone() }
func (v StoredActiveWorktree) Clone() StoredActiveWorktree    { return v }
func (d StoredActiveWorktreeData) clone() StoredActiveWorktreeData {

	return d
}
func (d StoredActiveWorktreeData) validate() error {

	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.AdministrativeKey == "" || d.LastKnownPath == "" || !storageString(d.AdministrativeKey) || !storageString(d.LastKnownPath) || !d.UnknownMembers.Excludes("administrativeKey", "lastKnownPath") {
		return invalid("stored active worktree")
	}
	return nil
}

// RunConfigDocument is an immutable semantic boundary value. Its zero is invalid.
type RunConfigDocument struct {
	data        RunConfigDocumentData
	initialized bool
}

// RunConfigDocumentData is a construction/copy value; NewRunConfigDocument validates and owns a copy.
type RunConfigDocumentData struct {
	SchemaVersion  uint32
	Default        StoredField[string]
	Launch         StoredField[[]SavedLaunchEntry]
	UnknownMembers JSONMembers
}

func NewRunConfigDocument(d RunConfigDocumentData) (RunConfigDocument, error) {
	if err := d.validate(); err != nil {
		return RunConfigDocument{}, err
	}
	return RunConfigDocument{data: d.clone(), initialized: true}, nil
}
func (v RunConfigDocument) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v RunConfigDocument) Data() RunConfigDocumentData { return v.data.clone() }
func (v RunConfigDocument) Clone() RunConfigDocument    { return v }
func (d RunConfigDocumentData) clone() RunConfigDocumentData {
	d.Launch = cloneStoredSlice(d.Launch)
	return d
}
func (d RunConfigDocumentData) validate() error {

	if !d.Default.Valid() {
		return invalid("field presence")
	}
	if !d.Launch.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.Launch.Value(); ok {
		for _, item := range item {
			if !item.Valid() {
				return invalid("item")
			}
		}
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.Default.Presence() == FieldNull {
		return invalid("null Default")
	}
	if d.Launch.Presence() == FieldNull {
		return invalid("null Launch")
	}
	if d.SchemaVersion > 1 {
		return invalid("schema version")
	}
	return validateRunDocument(d)
}

// SavedLaunchEntry is an immutable semantic boundary value. Its zero is invalid.
type SavedLaunchEntry struct {
	data        SavedLaunchEntryData
	initialized bool
}

// SavedLaunchEntryData is a construction/copy value; NewSavedLaunchEntry validates and owns a copy.
type SavedLaunchEntryData struct {
	Alias      string
	Definition SavedLaunchDefinition
}

func NewSavedLaunchEntry(d SavedLaunchEntryData) (SavedLaunchEntry, error) {
	if err := d.validate(); err != nil {
		return SavedLaunchEntry{}, err
	}
	return SavedLaunchEntry{data: d.clone(), initialized: true}, nil
}
func (v SavedLaunchEntry) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v SavedLaunchEntry) Data() SavedLaunchEntryData { return v.data.clone() }
func (v SavedLaunchEntry) Clone() SavedLaunchEntry    { return v }
func (d SavedLaunchEntryData) clone() SavedLaunchEntryData {

	return d
}
func (d SavedLaunchEntryData) validate() error {

	if !d.Definition.Valid() {
		return invalid("d.Definition")
	}
	if d.Alias == "" || !storageString(d.Alias) {
		return invalid("alias")
	}
	return nil
}

// SavedLaunchDefinition is an immutable semantic boundary value. Its zero is invalid.
type SavedLaunchDefinition struct {
	data        SavedLaunchDefinitionData
	initialized bool
}

// SavedLaunchDefinitionData is a construction/copy value; NewSavedLaunchDefinition validates and owns a copy.
type SavedLaunchDefinitionData struct {
	Provider       string
	Dir            StoredField[string]
	Script         StoredField[string]
	Targets        StoredField[[]string]
	Command        StoredField[string]
	UnknownMembers JSONMembers
}

func NewSavedLaunchDefinition(d SavedLaunchDefinitionData) (SavedLaunchDefinition, error) {
	if err := d.validate(); err != nil {
		return SavedLaunchDefinition{}, err
	}
	return SavedLaunchDefinition{data: d.clone(), initialized: true}, nil
}
func (v SavedLaunchDefinition) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v SavedLaunchDefinition) Data() SavedLaunchDefinitionData { return v.data.clone() }
func (v SavedLaunchDefinition) Clone() SavedLaunchDefinition    { return v }
func (d SavedLaunchDefinitionData) clone() SavedLaunchDefinitionData {
	d.Targets = cloneStoredSlice(d.Targets)
	return d
}
func (d SavedLaunchDefinitionData) validate() error {

	if !d.Dir.Valid() {
		return invalid("field presence")
	}
	if !d.Script.Valid() {
		return invalid("field presence")
	}
	if !d.Targets.Valid() {
		return invalid("field presence")
	}
	if !d.Command.Valid() {
		return invalid("field presence")
	}
	if !d.UnknownMembers.Valid() {
		return invalid("d.UnknownMembers")
	}
	if d.Dir.Presence() == FieldNull {
		return invalid("null Dir")
	}
	if d.Script.Presence() == FieldNull {
		return invalid("null Script")
	}
	if d.Targets.Presence() == FieldNull {
		return invalid("null Targets")
	}
	if d.Command.Presence() == FieldNull {
		return invalid("null Command")
	}
	if d.Provider == "" || !storageString(d.Provider) || !storedText(d.Dir) || !storedText(d.Script) || !storedText(d.Command) || !d.UnknownMembers.Excludes("provider", "dir", "script", "targets", "command") {
		return invalid("saved definition")
	}
	if x, ok := d.Targets.Value(); ok {
		for _, s := range x {
			if !storageString(s) {
				return invalid("target bytes")
			}
		}
	}
	return nil
}

// StorageRecovery is an immutable semantic boundary value. Its zero is invalid.
type StorageRecovery struct {
	data        StorageRecoveryData
	initialized bool
}

// StorageRecoveryData is a construction/copy value; NewStorageRecovery validates and owns a copy.
type StorageRecoveryData struct {
	Record   RecoveryRecord
	Family   StorageFamily
	Locator  string
	Kind     StorageRecoveryKind
	Identity SourceVersion
}

func NewStorageRecovery(d StorageRecoveryData) (StorageRecovery, error) {
	if err := d.validate(); err != nil {
		return StorageRecovery{}, err
	}
	return StorageRecovery{data: d.clone(), initialized: true}, nil
}
func (v StorageRecovery) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v StorageRecovery) Data() StorageRecoveryData { return v.data.clone() }
func (v StorageRecovery) Clone() StorageRecovery    { return v }
func (d StorageRecoveryData) clone() StorageRecoveryData {

	return d
}
func (d StorageRecoveryData) validate() error {
	if !d.Record.Valid() {
		return invalid("d.Record")
	}
	if !d.Family.Valid() {
		return invalid("d.Family")
	}

	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.Identity.Valid() {
		return invalid("d.Identity")
	}
	return validateStorageRecovery(d)
}

// StorageLoadObservation is an immutable semantic boundary value. Its zero is invalid.
type StorageLoadObservation struct {
	data        StorageLoadObservationData
	initialized bool
}

// StorageLoadObservationData is a construction/copy value; NewStorageLoadObservation validates and owns a copy.
type StorageLoadObservationData struct {
	State         StorageLoadState
	Version       Optional[StorageVersion]
	SchemaVersion Optional[uint32]
	Diagnostics   []Diagnostic
	Recovery      []StorageRecovery
}

func NewStorageLoadObservation(d StorageLoadObservationData) (StorageLoadObservation, error) {
	if err := d.validate(); err != nil {
		return StorageLoadObservation{}, err
	}
	return StorageLoadObservation{data: d.clone(), initialized: true}, nil
}
func (v StorageLoadObservation) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v StorageLoadObservation) Data() StorageLoadObservationData { return v.data.clone() }
func (v StorageLoadObservation) Clone() StorageLoadObservation    { return v }
func (d StorageLoadObservationData) clone() StorageLoadObservationData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d StorageLoadObservationData) validate() error {
	if !d.State.Valid() {
		return invalid("d.State")
	}
	if item, ok := d.Version.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}
	versions := []StorageVersion{}
	if v, p := d.Version.Value(); p {
		versions = append(versions, v)
	}
	if err := storageAssociations(versions, d.Recovery); err != nil {
		return err
	}
	if (d.State == LoadAbsent || d.State == ValidLegacy || d.State == ValidCurrent) && !d.Version.Present() {
		return invalid("usable load requires version")
	}
	if d.State == LoadAbsent {
		v, _ := d.Version.Value()
		if v.Present() {
			return invalid("absent load")
		}
	}
	if d.State == ValidLegacy || d.State == ValidCurrent {
		v, _ := d.Version.Value()
		s, p := d.SchemaVersion.Value()
		if !v.Present() || !p || (d.State == ValidLegacy && s != 0) || (d.State == ValidCurrent && s != 1) {
			return invalid("loaded schema/content")
		}
	}
	return validateStorageRecoveryList(d.Recovery)
}

// StorageCommitResult is an immutable semantic boundary value. Its zero is invalid.
type StorageCommitResult struct {
	data        StorageCommitResultData
	initialized bool
}

// StorageCommitResultData is a construction/copy value; NewStorageCommitResult validates and owns a copy.
type StorageCommitResultData struct {
	Outcome           StorageCommitOutcome
	ProposedVersion   Optional[StorageVersion]
	CurrentVersion    Optional[StorageVersion]
	PublicationKnown  bool
	Durability        StorageDurability
	CancellationAsked bool
	Effects           EffectReport
	Recovery          []StorageRecovery
	Diagnostics       []Diagnostic
}

func NewStorageCommitResult(d StorageCommitResultData) (StorageCommitResult, error) {
	if err := d.validate(); err != nil {
		return StorageCommitResult{}, err
	}
	return StorageCommitResult{data: d.clone(), initialized: true}, nil
}
func (v StorageCommitResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v StorageCommitResult) Data() StorageCommitResultData { return v.data.clone() }
func (v StorageCommitResult) Clone() StorageCommitResult    { return v }
func (d StorageCommitResultData) clone() StorageCommitResultData {
	d.Recovery = cloneSlice(d.Recovery)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d StorageCommitResultData) validate() error {
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if item, ok := d.ProposedVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.CurrentVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !d.Durability.Valid() {
		return invalid("d.Durability")
	}

	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := validateStorageCommitResultEvidence(d); err != nil {
		return err
	}
	return validateStorageCommit(d)
}
