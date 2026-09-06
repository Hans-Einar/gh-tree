package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// ProviderProfile is an immutable semantic boundary value. Its zero is invalid.
type ProviderProfile struct {
	data        ProviderProfileData
	initialized bool
}

// ProviderProfileData is a construction/copy value; NewProviderProfile validates and owns a copy.
type ProviderProfileData struct {
	Provider ProviderKind
	Profile  string
	Version  SourceVersion
}

func NewProviderProfile(d ProviderProfileData) (ProviderProfile, error) {
	if err := d.validate(); err != nil {
		return ProviderProfile{}, err
	}
	return ProviderProfile{data: d.clone(), initialized: true}, nil
}
func (v ProviderProfile) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ProviderProfile) Data() ProviderProfileData { return v.data.clone() }
func (v ProviderProfile) Clone() ProviderProfile    { return v }
func (d ProviderProfileData) clone() ProviderProfileData {

	return d
}
func (d ProviderProfileData) validate() error {
	if !d.Provider.Valid() {
		return invalid("d.Provider")
	}

	if !d.Version.Valid() {
		return invalid("d.Version")
	}
	if !nonempty(d.Profile) {
		return invalid("provider profile")
	}
	return nil
}

// DiscoverySourceDiagnostic is an immutable semantic boundary value. Its zero is invalid.
type DiscoverySourceDiagnostic struct {
	data        DiscoverySourceDiagnosticData
	initialized bool
}

// DiscoverySourceDiagnosticData is a construction/copy value; NewDiscoverySourceDiagnostic validates and owns a copy.
type DiscoverySourceDiagnosticData struct {
	Locator    string
	Diagnostic Diagnostic
}

func NewDiscoverySourceDiagnostic(d DiscoverySourceDiagnosticData) (DiscoverySourceDiagnostic, error) {
	if err := d.validate(); err != nil {
		return DiscoverySourceDiagnostic{}, err
	}
	return DiscoverySourceDiagnostic{data: d.clone(), initialized: true}, nil
}
func (v DiscoverySourceDiagnostic) Valid() bool                         { return v.initialized && v.data.validate() == nil }
func (v DiscoverySourceDiagnostic) Data() DiscoverySourceDiagnosticData { return v.data.clone() }
func (v DiscoverySourceDiagnostic) Clone() DiscoverySourceDiagnostic    { return v }
func (d DiscoverySourceDiagnosticData) clone() DiscoverySourceDiagnosticData {

	return d
}
func (d DiscoverySourceDiagnosticData) validate() error {

	if !d.Diagnostic.Valid() {
		return invalid("d.Diagnostic")
	}

	return nil
}

// DiscoveryObservation is an immutable semantic boundary value. Its zero is invalid.
type DiscoveryObservation struct {
	data        DiscoveryObservationData
	initialized bool
}

// DiscoveryObservationData is a construction/copy value; NewDiscoveryObservation validates and owns a copy.
type DiscoveryObservationData struct {
	ObservationID    ObservationID
	WorktreeID       domain.WorktreeID
	Interval         ObservationInterval
	SourceVersion    SourceVersion
	Completeness     Completeness
	Visited          uint64
	Skipped          uint64
	ProviderProfiles []ProviderProfile
	Sources          []DiscoverySourceDiagnostic
}

func NewDiscoveryObservation(d DiscoveryObservationData) (DiscoveryObservation, error) {
	if err := d.validate(); err != nil {
		return DiscoveryObservation{}, err
	}
	return DiscoveryObservation{data: d.clone(), initialized: true}, nil
}
func (v DiscoveryObservation) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v DiscoveryObservation) Data() DiscoveryObservationData { return v.data.clone() }
func (v DiscoveryObservation) Clone() DiscoveryObservation    { return v }
func (d DiscoveryObservationData) clone() DiscoveryObservationData {
	d.ProviderProfiles = cloneSlice(d.ProviderProfiles)
	d.Sources = cloneSlice(d.Sources)
	return d
}
func (d DiscoveryObservationData) validate() error {
	if !d.ObservationID.Valid() {
		return invalid("d.ObservationID")
	}
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Interval.Valid() {
		return invalid("d.Interval")
	}
	if !d.SourceVersion.Valid() {
		return invalid("d.SourceVersion")
	}
	if !d.Completeness.Valid() {
		return invalid("d.Completeness")
	}

	for _, item := range d.ProviderProfiles {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Sources {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// ProjectInputObservation is an immutable semantic boundary value. Its zero is invalid.
type ProjectInputObservation struct {
	data        ProjectInputObservationData
	initialized bool
}

// ProjectInputObservationData is a construction/copy value; NewProjectInputObservation validates and owns a copy.
type ProjectInputObservationData struct {
	Locator     string
	Identity    SourceVersion
	Content     SourceVersion
	Regular     bool
	Diagnostics []Diagnostic
}

func NewProjectInputObservation(d ProjectInputObservationData) (ProjectInputObservation, error) {
	if err := d.validate(); err != nil {
		return ProjectInputObservation{}, err
	}
	return ProjectInputObservation{data: d.clone(), initialized: true}, nil
}
func (v ProjectInputObservation) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v ProjectInputObservation) Data() ProjectInputObservationData { return v.data.clone() }
func (v ProjectInputObservation) Clone() ProjectInputObservation    { return v }
func (d ProjectInputObservationData) clone() ProjectInputObservationData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ProjectInputObservationData) validate() error {

	if !d.Identity.Valid() {
		return invalid("d.Identity")
	}
	if !d.Content.Valid() {
		return invalid("d.Content")
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// ProjectSource is an immutable semantic boundary value. Its zero is invalid.
type ProjectSource struct {
	data        ProjectSourceData
	initialized bool
}

// ProjectSourceData is a construction/copy value; NewProjectSource validates and owns a copy.
type ProjectSourceData struct {
	ManifestLocator  string
	ManifestIdentity SourceVersion
	Content          SourceVersion
	Inputs           []ProjectInputObservation
	ParserProfile    string
	RootIdentity     DirectoryIdentity
	ProjectIdentity  DirectoryIdentity
}

func NewProjectSource(d ProjectSourceData) (ProjectSource, error) {
	if err := d.validate(); err != nil {
		return ProjectSource{}, err
	}
	return ProjectSource{data: d.clone(), initialized: true}, nil
}
func (v ProjectSource) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v ProjectSource) Data() ProjectSourceData { return v.data.clone() }
func (v ProjectSource) Clone() ProjectSource    { return v }
func (d ProjectSourceData) clone() ProjectSourceData {
	d.Inputs = cloneSlice(d.Inputs)
	return d
}
func (d ProjectSourceData) validate() error {

	if !d.ManifestIdentity.Valid() {
		return invalid("d.ManifestIdentity")
	}
	if !d.Content.Valid() {
		return invalid("d.Content")
	}
	for _, item := range d.Inputs {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !d.RootIdentity.Valid() {
		return invalid("d.RootIdentity")
	}
	if !d.ProjectIdentity.Valid() {
		return invalid("d.ProjectIdentity")
	}
	if !nonempty(d.ManifestLocator) || !nonempty(d.ParserProfile) {
		return invalid("project source")
	}
	return nil
}

// LaunchDefinition is an immutable semantic boundary value. Its zero is invalid.
type LaunchDefinition struct {
	data        LaunchDefinitionData
	initialized bool
}

// LaunchDefinitionData is a construction/copy value; NewLaunchDefinition validates and owns a copy.
type LaunchDefinitionData struct {
	LaunchPointID       domain.LaunchPointID
	Provider            ProviderKind
	ProjectComponents   []string
	Member              string
	DisplayPath         string
	Label               string
	ProjectSource       ProjectSource
	EffectiveExecutable ArgvExecution
	Available           bool
	Diagnostics         []Diagnostic
}

func NewLaunchDefinition(d LaunchDefinitionData) (LaunchDefinition, error) {
	if err := d.validate(); err != nil {
		return LaunchDefinition{}, err
	}
	return LaunchDefinition{data: d.clone(), initialized: true}, nil
}
func (v LaunchDefinition) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v LaunchDefinition) Data() LaunchDefinitionData { return v.data.clone() }
func (v LaunchDefinition) Clone() LaunchDefinition    { return v }
func (d LaunchDefinitionData) clone() LaunchDefinitionData {
	d.ProjectComponents = cloneSlice(d.ProjectComponents)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d LaunchDefinitionData) validate() error {
	if !d.LaunchPointID.Valid() {
		return invalid("d.LaunchPointID")
	}
	if !d.Provider.Valid() {
		return invalid("d.Provider")
	}

	if !d.ProjectSource.Valid() {
		return invalid("d.ProjectSource")
	}
	if !d.EffectiveExecutable.Valid() {
		return invalid("d.EffectiveExecutable")
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !components(d.ProjectComponents) || !nonempty(d.Member) || !validLaunchDefinition(d) {
		return invalid("launch definition")
	}
	return nil
}

// MemberSelection is an immutable semantic boundary value. Its zero is invalid.
type MemberSelection struct {
	data        MemberSelectionData
	initialized bool
}

// MemberSelectionData is a construction/copy value; NewMemberSelection validates and owns a copy.
type MemberSelectionData struct {
	LaunchPointID domain.LaunchPointID
	SourceVersion SourceVersion
}

func NewMemberSelection(d MemberSelectionData) (MemberSelection, error) {
	if err := d.validate(); err != nil {
		return MemberSelection{}, err
	}
	return MemberSelection{data: d.clone(), initialized: true}, nil
}
func (v MemberSelection) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v MemberSelection) Data() MemberSelectionData { return v.data.clone() }
func (v MemberSelection) Clone() MemberSelection    { return v }
func (d MemberSelectionData) clone() MemberSelectionData {

	return d
}
func (d MemberSelectionData) validate() error {
	if !d.LaunchPointID.Valid() {
		return invalid("d.LaunchPointID")
	}
	if !d.SourceVersion.Valid() {
		return invalid("d.SourceVersion")
	}

	return nil
}

// DiscoveredLaunch is an immutable semantic boundary value. Its zero is invalid.
type DiscoveredLaunch struct {
	data        DiscoveredLaunchData
	initialized bool
}

// DiscoveredLaunchData is a construction/copy value; NewDiscoveredLaunch validates and owns a copy.
type DiscoveredLaunchData struct {
	Member MemberSelection
}

func NewDiscoveredLaunch(d DiscoveredLaunchData) (DiscoveredLaunch, error) {
	if err := d.validate(); err != nil {
		return DiscoveredLaunch{}, err
	}
	return DiscoveredLaunch{data: d.clone(), initialized: true}, nil
}
func (v DiscoveredLaunch) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v DiscoveredLaunch) Data() DiscoveredLaunchData { return v.data.clone() }
func (v DiscoveredLaunch) Clone() DiscoveredLaunch    { return v }
func (d DiscoveredLaunchData) clone() DiscoveredLaunchData {

	return d
}
func (d DiscoveredLaunchData) validate() error {
	if !d.Member.Valid() {
		return invalid("d.Member")
	}

	return nil
}

// OrderedMakeLaunch is an immutable semantic boundary value. Its zero is invalid.
type OrderedMakeLaunch struct {
	data        OrderedMakeLaunchData
	initialized bool
}

// OrderedMakeLaunchData is a construction/copy value; NewOrderedMakeLaunch validates and owns a copy.
type OrderedMakeLaunchData struct {
	Members []MemberSelection
}

func NewOrderedMakeLaunch(d OrderedMakeLaunchData) (OrderedMakeLaunch, error) {
	if err := d.validate(); err != nil {
		return OrderedMakeLaunch{}, err
	}
	return OrderedMakeLaunch{data: d.clone(), initialized: true}, nil
}
func (v OrderedMakeLaunch) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v OrderedMakeLaunch) Data() OrderedMakeLaunchData { return v.data.clone() }
func (v OrderedMakeLaunch) Clone() OrderedMakeLaunch    { return v }
func (d OrderedMakeLaunchData) clone() OrderedMakeLaunchData {
	d.Members = cloneSlice(d.Members)
	return d
}
func (d OrderedMakeLaunchData) validate() error {
	for _, item := range d.Members {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validOrderedMake(d.Members) {
		return invalid("ordered Make project/source/provider")
	}
	return nil
}

// SavedLaunch is an immutable semantic boundary value. Its zero is invalid.
type SavedLaunch struct {
	data        SavedLaunchData
	initialized bool
}

// SavedLaunchData is a construction/copy value; NewSavedLaunch validates and owns a copy.
type SavedLaunchData struct {
	Alias             string
	LaunchPointID     domain.LaunchPointID
	StorageVersion    StorageVersion
	SourceExpectation SourceVersion
}

func NewSavedLaunch(d SavedLaunchData) (SavedLaunch, error) {
	if err := d.validate(); err != nil {
		return SavedLaunch{}, err
	}
	return SavedLaunch{data: d.clone(), initialized: true}, nil
}
func (v SavedLaunch) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v SavedLaunch) Data() SavedLaunchData { return v.data.clone() }
func (v SavedLaunch) Clone() SavedLaunch    { return v }
func (d SavedLaunchData) clone() SavedLaunchData {

	return d
}
func (d SavedLaunchData) validate() error {

	if !d.LaunchPointID.Valid() {
		return invalid("d.LaunchPointID")
	}
	if !d.StorageVersion.Valid() {
		return invalid("d.StorageVersion")
	}
	if !d.SourceExpectation.Valid() {
		return invalid("d.SourceExpectation")
	}
	if !nonempty(d.Alias) || d.StorageVersion.Family() != RunConfig {
		return invalid("saved selection")
	}
	return nil
}

// SavedLaunchObservation is an immutable semantic boundary value. Its zero is invalid.
type SavedLaunchObservation struct {
	data        SavedLaunchObservationData
	initialized bool
}

// SavedLaunchObservationData is a construction/copy value; NewSavedLaunchObservation validates and owns a copy.
type SavedLaunchObservationData struct {
	Alias          string
	LaunchPointID  Optional[domain.LaunchPointID]
	Definition     Optional[LaunchDefinition]
	StorageVersion StorageVersion
	SourceVersion  Optional[SourceVersion]
	Diagnostics    []Diagnostic
}

func NewSavedLaunchObservation(d SavedLaunchObservationData) (SavedLaunchObservation, error) {
	if err := d.validate(); err != nil {
		return SavedLaunchObservation{}, err
	}
	return SavedLaunchObservation{data: d.clone(), initialized: true}, nil
}
func (v SavedLaunchObservation) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v SavedLaunchObservation) Data() SavedLaunchObservationData { return v.data.clone() }
func (v SavedLaunchObservation) Clone() SavedLaunchObservation    { return v }
func (d SavedLaunchObservationData) clone() SavedLaunchObservationData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SavedLaunchObservationData) validate() error {

	if item, ok := d.LaunchPointID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Definition.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.StorageVersion.Valid() {
		return invalid("d.StorageVersion")
	}
	if item, ok := d.SourceVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !nonempty(d.Alias) || d.StorageVersion.Family() != RunConfig {
		return invalid("saved observation")
	}
	return nil
}

// ResolvedLaunchDefinition is an immutable semantic boundary value. Its zero is invalid.
type ResolvedLaunchDefinition struct {
	data        ResolvedLaunchDefinitionData
	initialized bool
}

// ResolvedLaunchDefinitionData is a construction/copy value; NewResolvedLaunchDefinition validates and owns a copy.
type ResolvedLaunchDefinitionData struct {
	Selected            []MemberSelection
	Alias               Optional[string]
	Provider            ProviderKind
	ProjectComponents   []string
	ProjectSource       ProjectSource
	EffectiveExecutable ArgvExecution
	SourceVersions      []SourceVersion
	SavedVersion        Optional[StorageVersion]
}

func NewResolvedLaunchDefinition(d ResolvedLaunchDefinitionData) (ResolvedLaunchDefinition, error) {
	if err := d.validate(); err != nil {
		return ResolvedLaunchDefinition{}, err
	}
	return ResolvedLaunchDefinition{data: d.clone(), initialized: true}, nil
}
func (v ResolvedLaunchDefinition) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v ResolvedLaunchDefinition) Data() ResolvedLaunchDefinitionData { return v.data.clone() }
func (v ResolvedLaunchDefinition) Clone() ResolvedLaunchDefinition    { return v }
func (d ResolvedLaunchDefinitionData) clone() ResolvedLaunchDefinitionData {
	d.Selected = cloneSlice(d.Selected)
	d.ProjectComponents = cloneSlice(d.ProjectComponents)
	d.SourceVersions = cloneSlice(d.SourceVersions)
	return d
}
func (d ResolvedLaunchDefinitionData) validate() error {
	for _, item := range d.Selected {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !d.Provider.Valid() {
		return invalid("d.Provider")
	}

	if !d.ProjectSource.Valid() {
		return invalid("d.ProjectSource")
	}
	if !d.EffectiveExecutable.Valid() {
		return invalid("d.EffectiveExecutable")
	}
	for _, item := range d.SourceVersions {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.SavedVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if len(d.Selected) == 0 || !components(d.ProjectComponents) {
		return invalid("resolved launch")
	}
	if d.Provider == Npm && len(d.Selected) != 1 {
		return invalid("npm singular member")
	}
	return nil
}

// DiscoveryRequest is an immutable semantic boundary value. Its zero is invalid.
type DiscoveryRequest struct {
	data        DiscoveryRequestData
	initialized bool
}

// DiscoveryRequestData is a construction/copy value; NewDiscoveryRequest validates and owns a copy.
type DiscoveryRequestData struct {
	Worktree     WorktreeScope
	Saved        []SavedLaunchEntry
	SavedVersion Optional[StorageVersion]
}

func NewDiscoveryRequest(d DiscoveryRequestData) (DiscoveryRequest, error) {
	if err := d.validate(); err != nil {
		return DiscoveryRequest{}, err
	}
	return DiscoveryRequest{data: d.clone(), initialized: true}, nil
}
func (v DiscoveryRequest) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v DiscoveryRequest) Data() DiscoveryRequestData { return v.data.clone() }
func (v DiscoveryRequest) Clone() DiscoveryRequest    { return v }
func (d DiscoveryRequestData) clone() DiscoveryRequestData {
	d.Saved = cloneSlice(d.Saved)
	return d
}
func (d DiscoveryRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	for _, item := range d.Saved {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.SavedVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validSavedBinding(d.Saved, d.SavedVersion) {
		return invalid("saved load binding")
	}
	return nil
}

// DiscoveryResult is an immutable semantic boundary value. Its zero is invalid.
type DiscoveryResult struct {
	data        DiscoveryResultData
	initialized bool
}

// DiscoveryResultData is a construction/copy value; NewDiscoveryResult validates and owns a copy.
type DiscoveryResultData struct {
	WorktreeID  domain.WorktreeID
	Definitions []LaunchDefinition
	Saved       []SavedLaunchObservation
	Observation DiscoveryObservation
	Diagnostics []Diagnostic
}

func NewDiscoveryResult(d DiscoveryResultData) (DiscoveryResult, error) {
	if err := d.validate(); err != nil {
		return DiscoveryResult{}, err
	}
	return DiscoveryResult{data: d.clone(), initialized: true}, nil
}
func (v DiscoveryResult) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v DiscoveryResult) Data() DiscoveryResultData { return v.data.clone() }
func (v DiscoveryResult) Clone() DiscoveryResult    { return v }
func (d DiscoveryResultData) clone() DiscoveryResultData {
	d.Definitions = cloneSlice(d.Definitions)
	d.Saved = cloneSlice(d.Saved)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d DiscoveryResultData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	for _, item := range d.Definitions {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Saved {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.WorktreeID != d.Observation.data.WorktreeID {
		return invalid("discovery observation scope")
	}
	for _, v := range d.Definitions {
		if v.data.LaunchPointID.Worktree() != d.WorktreeID {
			return invalid("definition scope")
		}
	}
	return nil
}

// ResolveLaunchRequest is an immutable semantic boundary value. Its zero is invalid.
type ResolveLaunchRequest struct {
	data        ResolveLaunchRequestData
	initialized bool
}

// ResolveLaunchRequestData is a construction/copy value; NewResolveLaunchRequest validates and owns a copy.
type ResolveLaunchRequestData struct {
	Worktree     WorktreeScope
	Selection    LaunchSelection
	Saved        []SavedLaunchEntry
	SavedVersion Optional[StorageVersion]
	Geometry     Geometry
}

func NewResolveLaunchRequest(d ResolveLaunchRequestData) (ResolveLaunchRequest, error) {
	if err := d.validate(); err != nil {
		return ResolveLaunchRequest{}, err
	}
	return ResolveLaunchRequest{data: d.clone(), initialized: true}, nil
}
func (v ResolveLaunchRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ResolveLaunchRequest) Data() ResolveLaunchRequestData { return v.data.clone() }
func (v ResolveLaunchRequest) Clone() ResolveLaunchRequest    { return v }
func (d ResolveLaunchRequestData) clone() ResolveLaunchRequestData {
	d.Saved = cloneSlice(d.Saved)
	return d
}
func (d ResolveLaunchRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !validLaunchSelection(d.Selection) {
		return invalid("d.Selection")
	}
	for _, item := range d.Saved {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.SavedVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}
	if !validSavedBinding(d.Saved, d.SavedVersion) || !launchMatchesWorktree(d.Selection, d.Worktree.data.ID) {
		return invalid("launch binding")
	}
	if s, ok := d.Selection.(SavedLaunch); ok {
		v, p := d.SavedVersion.Value()
		if !p || s.data.StorageVersion != v {
			return invalid("saved version")
		}
	}
	return nil
}

// ResolveLaunchResult is an immutable semantic boundary value. Its zero is invalid.
type ResolveLaunchResult struct {
	data        ResolveLaunchResultData
	initialized bool
}

// ResolveLaunchResultData is a construction/copy value; NewResolveLaunchResult validates and owns a copy.
type ResolveLaunchResultData struct {
	Definition  Optional[ResolvedLaunchDefinition]
	Invocation  Optional[Invocation]
	Observation DiscoveryObservation
	Diagnostics []Diagnostic
}

func NewResolveLaunchResult(d ResolveLaunchResultData) (ResolveLaunchResult, error) {
	if err := d.validate(); err != nil {
		return ResolveLaunchResult{}, err
	}
	return ResolveLaunchResult{data: d.clone(), initialized: true}, nil
}
func (v ResolveLaunchResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ResolveLaunchResult) Data() ResolveLaunchResultData { return v.data.clone() }
func (v ResolveLaunchResult) Clone() ResolveLaunchResult    { return v }
func (d ResolveLaunchResultData) clone() ResolveLaunchResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ResolveLaunchResultData) validate() error {
	if item, ok := d.Definition.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Invocation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Definition.Present() != d.Invocation.Present() {
		return invalid("resolved invocation")
	}
	if i, ok := d.Invocation.Value(); ok && i.data.Cwd.data.Worktree.data.ID != d.Observation.data.WorktreeID {
		return invalid("resolved cwd scope")
	}
	return nil
}
