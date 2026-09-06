package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// StorageProjectionSource is an immutable semantic boundary value. Its zero is invalid.
type StorageProjectionSource struct {
	data        StorageProjectionSourceData
	initialized bool
}

// StorageProjectionSourceData is a construction/copy value; NewStorageProjectionSource validates and owns a copy.
type StorageProjectionSourceData struct {
	Observation StorageLoadObservation
	Family      StorageFamily
}

func NewStorageProjectionSource(d StorageProjectionSourceData) (StorageProjectionSource, error) {
	if err := d.validate(); err != nil {
		return StorageProjectionSource{}, err
	}
	return StorageProjectionSource{data: d.clone(), initialized: true}, nil
}
func (v StorageProjectionSource) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v StorageProjectionSource) Data() StorageProjectionSourceData { return v.data.clone() }
func (v StorageProjectionSource) Clone() StorageProjectionSource    { return v }
func (d StorageProjectionSourceData) clone() StorageProjectionSourceData {

	return d
}
func (d StorageProjectionSourceData) validate() error {
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if !d.Family.Valid() {
		return invalid("d.Family")
	}

	return nil
}

// ProjectionSources is an immutable semantic boundary value. Its zero is invalid.
type ProjectionSources struct {
	data        ProjectionSourcesData
	initialized bool
}

// ProjectionSourcesData is a construction/copy value; NewProjectionSources validates and owns a copy.
type ProjectionSourcesData struct {
	Git         []GitObservation
	Remote      []RemoteObservation
	Storage     []StorageProjectionSource
	Discovery   []DiscoveryObservation
	Diagnostics []Diagnostic
}

func NewProjectionSources(d ProjectionSourcesData) (ProjectionSources, error) {
	if err := d.validate(); err != nil {
		return ProjectionSources{}, err
	}
	return ProjectionSources{data: d.clone(), initialized: true}, nil
}
func (v ProjectionSources) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v ProjectionSources) Data() ProjectionSourcesData { return v.data.clone() }
func (v ProjectionSources) Clone() ProjectionSources    { return v }
func (d ProjectionSourcesData) clone() ProjectionSourcesData {
	d.Git = cloneSlice(d.Git)
	d.Remote = cloneSlice(d.Remote)
	d.Storage = cloneSlice(d.Storage)
	d.Discovery = cloneSlice(d.Discovery)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ProjectionSourcesData) validate() error {
	for _, item := range d.Git {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Remote {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Storage {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Discovery {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// ActiveContext is an immutable semantic boundary value. Its zero is invalid.
type ActiveContext struct {
	data        ActiveContextData
	initialized bool
}

// ActiveContextData is a construction/copy value; NewActiveContext validates and owns a copy.
type ActiveContextData struct {
	WorktreeID        Optional[domain.WorktreeID]
	Scope             Optional[WorktreeScope]
	Status            Optional[StatusFacts]
	PreferenceVersion Optional[StorageVersion]
	CommitDiagnostics []Diagnostic
	Version           ContextVersion
}

func NewActiveContext(d ActiveContextData) (ActiveContext, error) {
	if err := d.validate(); err != nil {
		return ActiveContext{}, err
	}
	return ActiveContext{data: d.clone(), initialized: true}, nil
}
func (v ActiveContext) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v ActiveContext) Data() ActiveContextData { return v.data.clone() }
func (v ActiveContext) Clone() ActiveContext    { return v }
func (d ActiveContextData) clone() ActiveContextData {
	d.CommitDiagnostics = cloneSlice(d.CommitDiagnostics)
	return d
}
func (d ActiveContextData) validate() error {
	if item, ok := d.WorktreeID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Scope.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.PreferenceVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.CommitDiagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Version.Valid() {
		return invalid("d.Version")
	}
	if s, ok := d.Scope.Value(); ok {
		w, p := d.WorktreeID.Value()
		if !p || w != s.data.ID {
			return invalid("active scope")
		}
	}
	if s, ok := d.Status.Value(); ok {
		w, p := d.WorktreeID.Value()
		if !p || w != s.data.Worktree.data.ID {
			return invalid("active status")
		}
	}
	if v, ok := d.PreferenceVersion.Value(); ok && v.Family() != Preferences {
		return invalid("active preference version")
	}
	return nil
}

// PRRelation is an immutable semantic boundary value. Its zero is invalid.
type PRRelation struct {
	data        PRRelationData
	initialized bool
}

// PRRelationData is a construction/copy value; NewPRRelation validates and owns a copy.
type PRRelationData struct {
	PRID        domain.PRID
	Role        PRRole
	Branch      domain.BranchID
	Revision    Optional[domain.Revision]
	Evidence    RelationshipEvidence
	Diagnostics []Diagnostic
}

func NewPRRelation(d PRRelationData) (PRRelation, error) {
	if err := d.validate(); err != nil {
		return PRRelation{}, err
	}
	return PRRelation{data: d.clone(), initialized: true}, nil
}
func (v PRRelation) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v PRRelation) Data() PRRelationData { return v.data.clone() }
func (v PRRelation) Clone() PRRelation    { return v }
func (d PRRelationData) clone() PRRelationData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d PRRelationData) validate() error {
	if !d.PRID.Valid() {
		return invalid("d.PRID")
	}
	if !d.Role.Valid() {
		return invalid("d.Role")
	}
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if item, ok := d.Revision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Evidence.Valid() {
		return invalid("d.Evidence")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Branch.Kind() != domain.RemoteHead {
		return invalid("PR relation branch")
	}
	if r, ok := d.Revision.Value(); ok && r.Repository() != d.Branch.Repository() {
		return invalid("PR relation revision")
	}
	return nil
}

// WorktreeRelation is an immutable semantic boundary value. Its zero is invalid.
type WorktreeRelation struct {
	data        WorktreeRelationData
	initialized bool
}

// WorktreeRelationData is a construction/copy value; NewWorktreeRelation validates and owns a copy.
type WorktreeRelationData struct {
	WorktreeID     domain.WorktreeID
	Head           Optional[domain.Head]
	IdentitySource GitObservation
	Kind           WorktreeRelationKind
	Availability   WorktreeAvailability
	Current        bool
	Primary        bool
	Active         bool
}

func NewWorktreeRelation(d WorktreeRelationData) (WorktreeRelation, error) {
	if err := d.validate(); err != nil {
		return WorktreeRelation{}, err
	}
	return WorktreeRelation{data: d.clone(), initialized: true}, nil
}
func (v WorktreeRelation) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v WorktreeRelation) Data() WorktreeRelationData { return v.data.clone() }
func (v WorktreeRelation) Clone() WorktreeRelation    { return v }
func (d WorktreeRelationData) clone() WorktreeRelationData {

	return d
}
func (d WorktreeRelationData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.IdentitySource.Valid() {
		return invalid("d.IdentitySource")
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !validWorktreeAvailability(d.Availability) {
		return invalid("d.Availability")
	}

	if d.WorktreeID.Repository() != d.IdentitySource.data.Repository {
		return invalid("relation worktree scope")
	}
	if h, ok := d.Head.Value(); ok && !h.MatchesWorktree(d.WorktreeID) {
		return invalid("relation head")
	}
	return nil
}

// ActivationCandidate is an immutable semantic boundary value. Its zero is invalid.
type ActivationCandidate struct {
	data        ActivationCandidateData
	initialized bool
}

// ActivationCandidateData is a construction/copy value; NewActivationCandidate validates and owns a copy.
type ActivationCandidateData struct {
	WorktreeID       domain.WorktreeID
	Reason           WorktreeRelationKind
	SelectedRevision domain.Revision
	WorktreeRevision Optional[domain.Revision]
	Availability     WorktreeAvailability
	ContextVersion   ContextVersion
	SourceVersion    SourceVersion
}

func NewActivationCandidate(d ActivationCandidateData) (ActivationCandidate, error) {
	if err := d.validate(); err != nil {
		return ActivationCandidate{}, err
	}
	return ActivationCandidate{data: d.clone(), initialized: true}, nil
}
func (v ActivationCandidate) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ActivationCandidate) Data() ActivationCandidateData { return v.data.clone() }
func (v ActivationCandidate) Clone() ActivationCandidate    { return v }
func (d ActivationCandidateData) clone() ActivationCandidateData {

	return d
}
func (d ActivationCandidateData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	if !d.SelectedRevision.Valid() {
		return invalid("d.SelectedRevision")
	}
	if item, ok := d.WorktreeRevision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validWorktreeAvailability(d.Availability) {
		return invalid("d.Availability")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	if !d.SourceVersion.Valid() {
		return invalid("d.SourceVersion")
	}

	return nil
}

// BranchRelationship is an immutable semantic boundary value. Its zero is invalid.
type BranchRelationship struct {
	data        BranchRelationshipData
	initialized bool
}

// BranchRelationshipData is a construction/copy value; NewBranchRelationship validates and owns a copy.
type BranchRelationshipData struct {
	Branch           domain.BranchID
	ExpectedRevision domain.Revision
	LocalBranch      Optional[RefFact]
	Upstream         UpstreamFact
	RemoteEndpoints  []RemoteBranchFact
	PullRequests     []PRRelation
	Worktrees        []WorktreeRelation
	Sources          ProjectionSources
	Completeness     Completeness
	Diagnostics      []Diagnostic
}

func NewBranchRelationship(d BranchRelationshipData) (BranchRelationship, error) {
	if err := d.validate(); err != nil {
		return BranchRelationship{}, err
	}
	return BranchRelationship{data: d.clone(), initialized: true}, nil
}
func (v BranchRelationship) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v BranchRelationship) Data() BranchRelationshipData { return v.data.clone() }
func (v BranchRelationship) Clone() BranchRelationship    { return v }
func (d BranchRelationshipData) clone() BranchRelationshipData {
	d.RemoteEndpoints = cloneSlice(d.RemoteEndpoints)
	d.PullRequests = cloneSlice(d.PullRequests)
	d.Worktrees = cloneSlice(d.Worktrees)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d BranchRelationshipData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.ExpectedRevision.Valid() {
		return invalid("d.ExpectedRevision")
	}
	if item, ok := d.LocalBranch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validUpstreamFact(d.Upstream) {
		return invalid("d.Upstream")
	}
	for _, item := range d.RemoteEndpoints {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.PullRequests {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.Completeness.Valid() {
		return invalid("d.Completeness")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Branch.Repository() != d.ExpectedRevision.Repository() {
		return invalid("branch relationship")
	}
	return nil
}

// NavigationIntent is an immutable semantic boundary value. Its zero is invalid.
type NavigationIntent struct {
	data        NavigationIntentData
	initialized bool
}

// NavigationIntentData is a construction/copy value; NewNavigationIntent validates and owns a copy.
type NavigationIntentData struct {
	Repository domain.RepositoryID
	Namespace  []string
	Folder     string
}

func NewNavigationIntent(d NavigationIntentData) (NavigationIntent, error) {
	if err := d.validate(); err != nil {
		return NavigationIntent{}, err
	}
	return NavigationIntent{data: d.clone(), initialized: true}, nil
}
func (v NavigationIntent) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v NavigationIntent) Data() NavigationIntentData { return v.data.clone() }
func (v NavigationIntent) Clone() NavigationIntent    { return v }
func (d NavigationIntentData) clone() NavigationIntentData {
	d.Namespace = cloneSlice(d.Namespace)
	return d
}
func (d NavigationIntentData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	for _, s := range d.Namespace {
		if !textValue(s) {
			return invalid("namespace")
		}
	}
	if !textValue(d.Folder) {
		return invalid("folder")
	}
	return nil
}

// ConfiguredDestination is an immutable semantic boundary value. Its zero is invalid.
type ConfiguredDestination struct {
	data        ConfiguredDestinationData
	initialized bool
}

// ConfiguredDestinationData is a construction/copy value; NewConfiguredDestination validates and owns a copy.
type ConfiguredDestinationData struct {
	Repository      domain.RepositoryID
	Name            string
	ExpectedStorage StorageVersion
	Worktree        Optional[domain.WorktreeID]
	ExpectedSource  Optional[SourceVersion]
}

func NewConfiguredDestination(d ConfiguredDestinationData) (ConfiguredDestination, error) {
	if err := d.validate(); err != nil {
		return ConfiguredDestination{}, err
	}
	return ConfiguredDestination{data: d.clone(), initialized: true}, nil
}
func (v ConfiguredDestination) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v ConfiguredDestination) Data() ConfiguredDestinationData { return v.data.clone() }
func (v ConfiguredDestination) Clone() ConfiguredDestination    { return v }
func (d ConfiguredDestinationData) clone() ConfiguredDestinationData {

	return d
}
func (d ConfiguredDestinationData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	if !d.ExpectedStorage.Valid() {
		return invalid("d.ExpectedStorage")
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ExpectedSource.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Repository.Scope() != domain.Remote || !nonempty(d.Name) || d.ExpectedStorage.Family() != UserConfig {
		return invalid("configured destination")
	}
	return nil
}

// ActiveDestination is an immutable semantic boundary value. Its zero is invalid.
type ActiveDestination struct {
	data        ActiveDestinationData
	initialized bool
}

// ActiveDestinationData is a construction/copy value; NewActiveDestination validates and owns a copy.
type ActiveDestinationData struct {
	ExpectedContext ContextVersion
}

func NewActiveDestination(d ActiveDestinationData) (ActiveDestination, error) {
	if err := d.validate(); err != nil {
		return ActiveDestination{}, err
	}
	return ActiveDestination{data: d.clone(), initialized: true}, nil
}
func (v ActiveDestination) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v ActiveDestination) Data() ActiveDestinationData { return v.data.clone() }
func (v ActiveDestination) Clone() ActiveDestination    { return v }
func (d ActiveDestinationData) clone() ActiveDestinationData {

	return d
}
func (d ActiveDestinationData) validate() error {
	if !d.ExpectedContext.Valid() {
		return invalid("d.ExpectedContext")
	}

	return nil
}

// WorktreeDestination is an immutable semantic boundary value. Its zero is invalid.
type WorktreeDestination struct {
	data        WorktreeDestinationData
	initialized bool
}

// WorktreeDestinationData is a construction/copy value; NewWorktreeDestination validates and owns a copy.
type WorktreeDestinationData struct {
	Worktree       domain.WorktreeID
	ExpectedSource SourceVersion
}

func NewWorktreeDestination(d WorktreeDestinationData) (WorktreeDestination, error) {
	if err := d.validate(); err != nil {
		return WorktreeDestination{}, err
	}
	return WorktreeDestination{data: d.clone(), initialized: true}, nil
}
func (v WorktreeDestination) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v WorktreeDestination) Data() WorktreeDestinationData { return v.data.clone() }
func (v WorktreeDestination) Clone() WorktreeDestination    { return v }
func (d WorktreeDestinationData) clone() WorktreeDestinationData {

	return d
}
func (d WorktreeDestinationData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.ExpectedSource.Valid() {
		return invalid("d.ExpectedSource")
	}

	return nil
}

// CurrentDefaultLaunch is an immutable semantic boundary value. Its zero is invalid.
type CurrentDefaultLaunch struct {
	data        CurrentDefaultLaunchData
	initialized bool
}

// CurrentDefaultLaunchData is a construction/copy value; NewCurrentDefaultLaunch validates and owns a copy.
type CurrentDefaultLaunchData struct {
	ExpectedStorage Optional[StorageVersion]
}

func NewCurrentDefaultLaunch(d CurrentDefaultLaunchData) (CurrentDefaultLaunch, error) {
	if err := d.validate(); err != nil {
		return CurrentDefaultLaunch{}, err
	}
	return CurrentDefaultLaunch{data: d.clone(), initialized: true}, nil
}
func (v CurrentDefaultLaunch) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v CurrentDefaultLaunch) Data() CurrentDefaultLaunchData { return v.data.clone() }
func (v CurrentDefaultLaunch) Clone() CurrentDefaultLaunch    { return v }
func (d CurrentDefaultLaunchData) clone() CurrentDefaultLaunchData {

	return d
}
func (d CurrentDefaultLaunchData) validate() error {
	if item, ok := d.ExpectedStorage.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if v, p := d.ExpectedStorage.Value(); p && v.Family() != RunConfig {
		return invalid("default storage")
	}
	return nil
}

// SelectedLaunch is an immutable semantic boundary value. Its zero is invalid.
type SelectedLaunch struct {
	data        SelectedLaunchData
	initialized bool
}

// SelectedLaunchData is a construction/copy value; NewSelectedLaunch validates and owns a copy.
type SelectedLaunchData struct {
	Selection LaunchSelection
}

func NewSelectedLaunch(d SelectedLaunchData) (SelectedLaunch, error) {
	if err := d.validate(); err != nil {
		return SelectedLaunch{}, err
	}
	return SelectedLaunch{data: d.clone(), initialized: true}, nil
}
func (v SelectedLaunch) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v SelectedLaunch) Data() SelectedLaunchData { return v.data.clone() }
func (v SelectedLaunch) Clone() SelectedLaunch    { return v }
func (d SelectedLaunchData) clone() SelectedLaunchData {

	return d
}
func (d SelectedLaunchData) validate() error {
	if !validLaunchSelection(d.Selection) {
		return invalid("d.Selection")
	}

	return nil
}

// AliasReplacement is an immutable semantic boundary value. Its zero is invalid.
type AliasReplacement struct {
	data        AliasReplacementData
	initialized bool
}

// AliasReplacementData is a construction/copy value; NewAliasReplacement validates and owns a copy.
type AliasReplacementData struct {
	Alias    string
	Expected StorageVersion
	Prior    SavedLaunchDefinition
}

func NewAliasReplacement(d AliasReplacementData) (AliasReplacement, error) {
	if err := d.validate(); err != nil {
		return AliasReplacement{}, err
	}
	return AliasReplacement{data: d.clone(), initialized: true}, nil
}
func (v AliasReplacement) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v AliasReplacement) Data() AliasReplacementData { return v.data.clone() }
func (v AliasReplacement) Clone() AliasReplacement    { return v }
func (d AliasReplacementData) clone() AliasReplacementData {

	return d
}
func (d AliasReplacementData) validate() error {

	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !d.Prior.Valid() {
		return invalid("d.Prior")
	}
	if !nonempty(d.Alias) || d.Expected.Family() != RunConfig {
		return invalid("replacement")
	}
	return nil
}

// ActivateWorktreeCommand is an immutable semantic boundary value. Its zero is invalid.
type ActivateWorktreeCommand struct {
	data        ActivateWorktreeCommandData
	initialized bool
}

// ActivateWorktreeCommandData is a construction/copy value; NewActivateWorktreeCommand validates and owns a copy.
type ActivateWorktreeCommandData struct {
	WorktreeID         domain.WorktreeID
	ExpectedPreference StorageVersion
}

func NewActivateWorktreeCommand(d ActivateWorktreeCommandData) (ActivateWorktreeCommand, error) {
	if err := d.validate(); err != nil {
		return ActivateWorktreeCommand{}, err
	}
	return ActivateWorktreeCommand{data: d.clone(), initialized: true}, nil
}
func (v ActivateWorktreeCommand) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v ActivateWorktreeCommand) Data() ActivateWorktreeCommandData { return v.data.clone() }
func (v ActivateWorktreeCommand) Clone() ActivateWorktreeCommand    { return v }
func (d ActivateWorktreeCommandData) clone() ActivateWorktreeCommandData {

	return d
}
func (d ActivateWorktreeCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.ExpectedPreference.Valid() {
		return invalid("d.ExpectedPreference")
	}
	if d.ExpectedPreference.Family() != Preferences {
		return invalid("preference family")
	}
	return nil
}

// SaveNavigationCommand is an immutable semantic boundary value. Its zero is invalid.
type SaveNavigationCommand struct {
	data        SaveNavigationCommandData
	initialized bool
}

// SaveNavigationCommandData is a construction/copy value; NewSaveNavigationCommand validates and owns a copy.
type SaveNavigationCommandData struct {
	Intent   NavigationIntent
	Expected StorageVersion
}

func NewSaveNavigationCommand(d SaveNavigationCommandData) (SaveNavigationCommand, error) {
	if err := d.validate(); err != nil {
		return SaveNavigationCommand{}, err
	}
	return SaveNavigationCommand{data: d.clone(), initialized: true}, nil
}
func (v SaveNavigationCommand) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v SaveNavigationCommand) Data() SaveNavigationCommandData { return v.data.clone() }
func (v SaveNavigationCommand) Clone() SaveNavigationCommand    { return v }
func (d SaveNavigationCommandData) clone() SaveNavigationCommandData {

	return d
}
func (d SaveNavigationCommandData) validate() error {
	if !d.Intent.Valid() {
		return invalid("d.Intent")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if d.Expected.Family() != Preferences {
		return invalid("navigation family")
	}
	return nil
}

// CreateWorktreeCommand is an immutable semantic boundary value. Its zero is invalid.
type CreateWorktreeCommand struct {
	data        CreateWorktreeCommandData
	initialized bool
}

// CreateWorktreeCommandData is a construction/copy value; NewCreateWorktreeCommand validates and owns a copy.
type CreateWorktreeCommandData struct {
	Target      domain.ExactTarget
	Destination string
	Mode        CreateMode
}

func NewCreateWorktreeCommand(d CreateWorktreeCommandData) (CreateWorktreeCommand, error) {
	if err := d.validate(); err != nil {
		return CreateWorktreeCommand{}, err
	}
	return CreateWorktreeCommand{data: d.clone(), initialized: true}, nil
}
func (v CreateWorktreeCommand) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v CreateWorktreeCommand) Data() CreateWorktreeCommandData { return v.data.clone() }
func (v CreateWorktreeCommand) Clone() CreateWorktreeCommand    { return v }
func (d CreateWorktreeCommandData) clone() CreateWorktreeCommandData {

	return d
}
func (d CreateWorktreeCommandData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}

	if !validCreateMode(d.Mode) {
		return invalid("d.Mode")
	}
	if err := consistentCreateWorktreeCommand(d); err != nil {
		return err
	}
	if !nonempty(d.Destination) {
		return invalid("create path")
	}
	return nil
}

// RetargetWorktreeCommand is an immutable semantic boundary value. Its zero is invalid.
type RetargetWorktreeCommand struct {
	data        RetargetWorktreeCommandData
	initialized bool
}

// RetargetWorktreeCommandData is a construction/copy value; NewRetargetWorktreeCommand validates and owns a copy.
type RetargetWorktreeCommandData struct {
	WorktreeID domain.WorktreeID
	Target     domain.ExactTarget
	Mode       RetargetMode
	Expected   GitExpectedState
}

func NewRetargetWorktreeCommand(d RetargetWorktreeCommandData) (RetargetWorktreeCommand, error) {
	if err := d.validate(); err != nil {
		return RetargetWorktreeCommand{}, err
	}
	return RetargetWorktreeCommand{data: d.clone(), initialized: true}, nil
}
func (v RetargetWorktreeCommand) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v RetargetWorktreeCommand) Data() RetargetWorktreeCommandData { return v.data.clone() }
func (v RetargetWorktreeCommand) Clone() RetargetWorktreeCommand    { return v }
func (d RetargetWorktreeCommandData) clone() RetargetWorktreeCommandData {

	return d
}
func (d RetargetWorktreeCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !validRetargetMode(d.Mode) {
		return invalid("d.Mode")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if err := consistentRetargetWorktreeCommand(d); err != nil {
		return err
	}
	if !expectedWorktree(d.Expected, d.WorktreeID, true) || !retargetModeScope(d.Mode, d.WorktreeID.Repository(), None[domain.Revision]()) {
		return invalid("retarget expected")
	}
	return nil
}

// DeployCommand is an immutable semantic boundary value. Its zero is invalid.
type DeployCommand struct {
	data        DeployCommandData
	initialized bool
}

// DeployCommandData is a construction/copy value; NewDeployCommand validates and owns a copy.
type DeployCommandData struct {
	Target      domain.ExactTarget
	Destination DeployDestination
	Mode        RetargetMode
}

func NewDeployCommand(d DeployCommandData) (DeployCommand, error) {
	if err := d.validate(); err != nil {
		return DeployCommand{}, err
	}
	return DeployCommand{data: d.clone(), initialized: true}, nil
}
func (v DeployCommand) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v DeployCommand) Data() DeployCommandData { return v.data.clone() }
func (v DeployCommand) Clone() DeployCommand    { return v }
func (d DeployCommandData) clone() DeployCommandData {

	return d
}
func (d DeployCommandData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !validDeployDestination(d.Destination) {
		return invalid("d.Destination")
	}
	if !validRetargetMode(d.Mode) {
		return invalid("d.Mode")
	}
	if err := consistentDeployCommand(d); err != nil {
		return err
	}
	return nil
}

// CreateBranchCommand is an immutable semantic boundary value. Its zero is invalid.
type CreateBranchCommand struct {
	data        CreateBranchCommandData
	initialized bool
}

// CreateBranchCommandData is a construction/copy value; NewCreateBranchCommand validates and owns a copy.
type CreateBranchCommandData struct {
	WorktreeID domain.WorktreeID
	Name       domain.BranchID
	Start      domain.Revision
	Checkout   bool
	Expected   GitExpectedState
}

func NewCreateBranchCommand(d CreateBranchCommandData) (CreateBranchCommand, error) {
	if err := d.validate(); err != nil {
		return CreateBranchCommand{}, err
	}
	return CreateBranchCommand{data: d.clone(), initialized: true}, nil
}
func (v CreateBranchCommand) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v CreateBranchCommand) Data() CreateBranchCommandData { return v.data.clone() }
func (v CreateBranchCommand) Clone() CreateBranchCommand    { return v }
func (d CreateBranchCommandData) clone() CreateBranchCommandData {

	return d
}
func (d CreateBranchCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Name.Valid() {
		return invalid("d.Name")
	}
	if !d.Start.Valid() {
		return invalid("d.Start")
	}

	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if err := consistentCreateBranchCommand(d); err != nil {
		return err
	}
	if !expectedWorktree(d.Expected, d.WorktreeID, false) || d.Name.Repository() != d.WorktreeID.Repository() || d.Start.Repository() != d.WorktreeID.Repository() {
		return invalid("branch command scope")
	}
	return nil
}

// FetchCommand is an immutable semantic boundary value. Its zero is invalid.
type FetchCommand struct {
	data        FetchCommandData
	initialized bool
}

// FetchCommandData is a construction/copy value; NewFetchCommand validates and owns a copy.
type FetchCommandData struct {
	Binding         RemoteBinding
	RefScope        []FetchRefSpec
	Prune           bool
	ExpectedBinding SourceVersion
}

func NewFetchCommand(d FetchCommandData) (FetchCommand, error) {
	if err := d.validate(); err != nil {
		return FetchCommand{}, err
	}
	return FetchCommand{data: d.clone(), initialized: true}, nil
}
func (v FetchCommand) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v FetchCommand) Data() FetchCommandData { return v.data.clone() }
func (v FetchCommand) Clone() FetchCommand    { return v }
func (d FetchCommandData) clone() FetchCommandData {
	d.RefScope = cloneSlice(d.RefScope)
	return d
}
func (d FetchCommandData) validate() error {
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	for _, item := range d.RefScope {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !d.ExpectedBinding.Valid() {
		return invalid("d.ExpectedBinding")
	}
	if len(d.RefScope) == 0 || d.ExpectedBinding != d.Binding.data.Configuration {
		return invalid("fetch scope")
	}
	return nil
}

// PullCommand is an immutable semantic boundary value. Its zero is invalid.
type PullCommand struct {
	data        PullCommandData
	initialized bool
}

// PullCommandData is a construction/copy value; NewPullCommand validates and owns a copy.
type PullCommandData struct {
	WorktreeID domain.WorktreeID
	Expected   GitExpectedState
	Head       domain.Head
	Upstream   ResolvedUpstream
}

func NewPullCommand(d PullCommandData) (PullCommand, error) {
	if err := d.validate(); err != nil {
		return PullCommand{}, err
	}
	return PullCommand{data: d.clone(), initialized: true}, nil
}
func (v PullCommand) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v PullCommand) Data() PullCommandData { return v.data.clone() }
func (v PullCommand) Clone() PullCommand    { return v }
func (d PullCommandData) clone() PullCommandData {

	return d
}
func (d PullCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !d.Head.Valid() {
		return invalid("d.Head")
	}
	if !d.Upstream.Valid() {
		return invalid("d.Upstream")
	}
	if err := consistentPullCommand(d); err != nil {
		return err
	}
	if !d.Head.MatchesWorktree(d.WorktreeID) || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("pull scope")
	}
	return nil
}

// PushCommand is an immutable semantic boundary value. Its zero is invalid.
type PushCommand struct {
	data        PushCommandData
	initialized bool
}

// PushCommandData is a construction/copy value; NewPushCommand validates and owns a copy.
type PushCommandData struct {
	WorktreeID  domain.WorktreeID
	Source      domain.Revision
	Destination domain.BranchID
	Binding     RemoteBinding
	SetUpstream Optional[UpstreamSetup]
	Expected    GitExpectedState
}

func NewPushCommand(d PushCommandData) (PushCommand, error) {
	if err := d.validate(); err != nil {
		return PushCommand{}, err
	}
	return PushCommand{data: d.clone(), initialized: true}, nil
}
func (v PushCommand) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v PushCommand) Data() PushCommandData { return v.data.clone() }
func (v PushCommand) Clone() PushCommand    { return v }
func (d PushCommandData) clone() PushCommandData {

	return d
}
func (d PushCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Source.Valid() {
		return invalid("d.Source")
	}
	if !d.Destination.Valid() {
		return invalid("d.Destination")
	}
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	if item, ok := d.SetUpstream.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if err := consistentPushCommand(d); err != nil {
		return err
	}
	if !expectedWorktree(d.Expected, d.WorktreeID, false) || d.Source.Repository() != d.WorktreeID.Repository() || d.Binding.data.LocalRepository != d.WorktreeID.Repository() || d.Destination.Repository() != d.Binding.data.RemoteRepository {
		return invalid("push scope")
	}
	return nil
}

// StagePathsCommand is an immutable semantic boundary value. Its zero is invalid.
type StagePathsCommand struct {
	data        StagePathsCommandData
	initialized bool
}

// StagePathsCommandData is a construction/copy value; NewStagePathsCommand validates and owns a copy.
type StagePathsCommandData struct {
	WorktreeID domain.WorktreeID
	Paths      []GitPath
	Expected   GitExpectedState
}

func NewStagePathsCommand(d StagePathsCommandData) (StagePathsCommand, error) {
	if err := d.validate(); err != nil {
		return StagePathsCommand{}, err
	}
	return StagePathsCommand{data: d.clone(), initialized: true}, nil
}
func (v StagePathsCommand) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v StagePathsCommand) Data() StagePathsCommandData { return v.data.clone() }
func (v StagePathsCommand) Clone() StagePathsCommand    { return v }
func (d StagePathsCommandData) clone() StagePathsCommandData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d StagePathsCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if len(d.Paths) == 0 || duplicatePaths(d.Paths) || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("stage paths")
	}
	return nil
}

// UnstagePathsCommand is an immutable semantic boundary value. Its zero is invalid.
type UnstagePathsCommand struct {
	data        UnstagePathsCommandData
	initialized bool
}

// UnstagePathsCommandData is a construction/copy value; NewUnstagePathsCommand validates and owns a copy.
type UnstagePathsCommandData struct {
	WorktreeID domain.WorktreeID
	Paths      []GitPath
	Expected   GitExpectedState
}

func NewUnstagePathsCommand(d UnstagePathsCommandData) (UnstagePathsCommand, error) {
	if err := d.validate(); err != nil {
		return UnstagePathsCommand{}, err
	}
	return UnstagePathsCommand{data: d.clone(), initialized: true}, nil
}
func (v UnstagePathsCommand) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v UnstagePathsCommand) Data() UnstagePathsCommandData { return v.data.clone() }
func (v UnstagePathsCommand) Clone() UnstagePathsCommand    { return v }
func (d UnstagePathsCommandData) clone() UnstagePathsCommandData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d UnstagePathsCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if len(d.Paths) == 0 || duplicatePaths(d.Paths) || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("unstage paths")
	}
	return nil
}

// StageAllCommand is an immutable semantic boundary value. Its zero is invalid.
type StageAllCommand struct {
	data        StageAllCommandData
	initialized bool
}

// StageAllCommandData is a construction/copy value; NewStageAllCommand validates and owns a copy.
type StageAllCommandData struct {
	WorktreeID domain.WorktreeID
	Expected   GitExpectedState
}

func NewStageAllCommand(d StageAllCommandData) (StageAllCommand, error) {
	if err := d.validate(); err != nil {
		return StageAllCommand{}, err
	}
	return StageAllCommand{data: d.clone(), initialized: true}, nil
}
func (v StageAllCommand) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v StageAllCommand) Data() StageAllCommandData { return v.data.clone() }
func (v StageAllCommand) Clone() StageAllCommand    { return v }
func (d StageAllCommandData) clone() StageAllCommandData {

	return d
}
func (d StageAllCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("stage all expected")
	}
	return nil
}

// CommitCommand is an immutable semantic boundary value. Its zero is invalid.
type CommitCommand struct {
	data        CommitCommandData
	initialized bool
}

// CommitCommandData is a construction/copy value; NewCommitCommand validates and owns a copy.
type CommitCommandData struct {
	WorktreeID  domain.WorktreeID
	Expected    GitExpectedState
	Message     string
	IndexPolicy CommitIndexPolicy
}

func NewCommitCommand(d CommitCommandData) (CommitCommand, error) {
	if err := d.validate(); err != nil {
		return CommitCommand{}, err
	}
	return CommitCommand{data: d.clone(), initialized: true}, nil
}
func (v CommitCommand) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v CommitCommand) Data() CommitCommandData { return v.data.clone() }
func (v CommitCommand) Clone() CommitCommand    { return v }
func (d CommitCommandData) clone() CommitCommandData {

	return d
}
func (d CommitCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}

	if !d.IndexPolicy.Valid() {
		return invalid("d.IndexPolicy")
	}
	if d.IndexPolicy != ExistingIndex || !nonempty(d.Message) || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("commit policy")
	}
	return nil
}

// StageAllAndCommitCommand is an immutable semantic boundary value. Its zero is invalid.
type StageAllAndCommitCommand struct {
	data        StageAllAndCommitCommandData
	initialized bool
}

// StageAllAndCommitCommandData is a construction/copy value; NewStageAllAndCommitCommand validates and owns a copy.
type StageAllAndCommitCommandData struct {
	WorktreeID  domain.WorktreeID
	Expected    GitExpectedState
	Message     string
	IndexPolicy CommitIndexPolicy
}

func NewStageAllAndCommitCommand(d StageAllAndCommitCommandData) (StageAllAndCommitCommand, error) {
	if err := d.validate(); err != nil {
		return StageAllAndCommitCommand{}, err
	}
	return StageAllAndCommitCommand{data: d.clone(), initialized: true}, nil
}
func (v StageAllAndCommitCommand) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v StageAllAndCommitCommand) Data() StageAllAndCommitCommandData { return v.data.clone() }
func (v StageAllAndCommitCommand) Clone() StageAllAndCommitCommand    { return v }
func (d StageAllAndCommitCommandData) clone() StageAllAndCommitCommandData {

	return d
}
func (d StageAllAndCommitCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}

	if !d.IndexPolicy.Valid() {
		return invalid("d.IndexPolicy")
	}
	if d.IndexPolicy != ObservedStageAll || !nonempty(d.Message) || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("staged commit policy")
	}
	return nil
}

// RestoreTrackedCommand is an immutable semantic boundary value. Its zero is invalid.
type RestoreTrackedCommand struct {
	data        RestoreTrackedCommandData
	initialized bool
}

// RestoreTrackedCommandData is a construction/copy value; NewRestoreTrackedCommand validates and owns a copy.
type RestoreTrackedCommandData struct {
	WorktreeID domain.WorktreeID
	Paths      []GitPath
	Expected   GitExpectedState
}

func NewRestoreTrackedCommand(d RestoreTrackedCommandData) (RestoreTrackedCommand, error) {
	if err := d.validate(); err != nil {
		return RestoreTrackedCommand{}, err
	}
	return RestoreTrackedCommand{data: d.clone(), initialized: true}, nil
}
func (v RestoreTrackedCommand) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v RestoreTrackedCommand) Data() RestoreTrackedCommandData { return v.data.clone() }
func (v RestoreTrackedCommand) Clone() RestoreTrackedCommand    { return v }
func (d RestoreTrackedCommandData) clone() RestoreTrackedCommandData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d RestoreTrackedCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if len(d.Paths) == 0 || duplicatePaths(d.Paths) || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("restore paths")
	}
	return nil
}

// StashCreateCommand is an immutable semantic boundary value. Its zero is invalid.
type StashCreateCommand struct {
	data        StashCreateCommandData
	initialized bool
}

// StashCreateCommandData is a construction/copy value; NewStashCreateCommand validates and owns a copy.
type StashCreateCommandData struct {
	WorktreeID       domain.WorktreeID
	Expected         GitExpectedState
	Message          string
	IncludeUntracked bool
}

func NewStashCreateCommand(d StashCreateCommandData) (StashCreateCommand, error) {
	if err := d.validate(); err != nil {
		return StashCreateCommand{}, err
	}
	return StashCreateCommand{data: d.clone(), initialized: true}, nil
}
func (v StashCreateCommand) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v StashCreateCommand) Data() StashCreateCommandData { return v.data.clone() }
func (v StashCreateCommand) Clone() StashCreateCommand    { return v }
func (d StashCreateCommandData) clone() StashCreateCommandData {

	return d
}
func (d StashCreateCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}

	if !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("stash create expected")
	}
	return nil
}

// StashApplyCommand is an immutable semantic boundary value. Its zero is invalid.
type StashApplyCommand struct {
	data        StashApplyCommandData
	initialized bool
}

// StashApplyCommandData is a construction/copy value; NewStashApplyCommand validates and owns a copy.
type StashApplyCommandData struct {
	WorktreeID domain.WorktreeID
	Stash      domain.StashID
	Occurrence SourceVersion
	Expected   GitExpectedState
}

func NewStashApplyCommand(d StashApplyCommandData) (StashApplyCommand, error) {
	if err := d.validate(); err != nil {
		return StashApplyCommand{}, err
	}
	return StashApplyCommand{data: d.clone(), initialized: true}, nil
}
func (v StashApplyCommand) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v StashApplyCommand) Data() StashApplyCommandData { return v.data.clone() }
func (v StashApplyCommand) Clone() StashApplyCommand    { return v }
func (d StashApplyCommandData) clone() StashApplyCommandData {

	return d
}
func (d StashApplyCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if d.Stash.Repository() != d.WorktreeID.Repository() || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("stash apply scope")
	}
	return nil
}

// StashPopCommand is an immutable semantic boundary value. Its zero is invalid.
type StashPopCommand struct {
	data        StashPopCommandData
	initialized bool
}

// StashPopCommandData is a construction/copy value; NewStashPopCommand validates and owns a copy.
type StashPopCommandData struct {
	WorktreeID domain.WorktreeID
	Stash      domain.StashID
	Occurrence SourceVersion
	Expected   GitExpectedState
}

func NewStashPopCommand(d StashPopCommandData) (StashPopCommand, error) {
	if err := d.validate(); err != nil {
		return StashPopCommand{}, err
	}
	return StashPopCommand{data: d.clone(), initialized: true}, nil
}
func (v StashPopCommand) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v StashPopCommand) Data() StashPopCommandData { return v.data.clone() }
func (v StashPopCommand) Clone() StashPopCommand    { return v }
func (d StashPopCommandData) clone() StashPopCommandData {

	return d
}
func (d StashPopCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if d.Stash.Repository() != d.WorktreeID.Repository() || !expectedWorktree(d.Expected, d.WorktreeID, true) {
		return invalid("stash pop scope")
	}
	return nil
}

// StashDropCommand is an immutable semantic boundary value. Its zero is invalid.
type StashDropCommand struct {
	data        StashDropCommandData
	initialized bool
}

// StashDropCommandData is a construction/copy value; NewStashDropCommand validates and owns a copy.
type StashDropCommandData struct {
	Stash      domain.StashID
	Occurrence SourceVersion
	Expected   GitExpectedState
}

func NewStashDropCommand(d StashDropCommandData) (StashDropCommand, error) {
	if err := d.validate(); err != nil {
		return StashDropCommand{}, err
	}
	return StashDropCommand{data: d.clone(), initialized: true}, nil
}
func (v StashDropCommand) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v StashDropCommand) Data() StashDropCommandData { return v.data.clone() }
func (v StashDropCommand) Clone() StashDropCommand    { return v }
func (d StashDropCommandData) clone() StashDropCommandData {

	return d
}
func (d StashDropCommandData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if d.Stash.Repository() != d.Expected.data.Repository {
		return invalid("stash drop scope")
	}
	return nil
}

// CreatePullRequestCommand is an immutable semantic boundary value. Its zero is invalid.
type CreatePullRequestCommand struct {
	data        CreatePullRequestCommandData
	initialized bool
}

// CreatePullRequestCommandData is a construction/copy value; NewCreatePullRequestCommand validates and owns a copy.
type CreatePullRequestCommandData struct {
	Base                EndpointExpectation
	Head                EndpointExpectation
	Title               string
	Body                string
	Draft               bool
	MaintainerCanModify bool
}

func NewCreatePullRequestCommand(d CreatePullRequestCommandData) (CreatePullRequestCommand, error) {
	if err := d.validate(); err != nil {
		return CreatePullRequestCommand{}, err
	}
	return CreatePullRequestCommand{data: d.clone(), initialized: true}, nil
}
func (v CreatePullRequestCommand) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v CreatePullRequestCommand) Data() CreatePullRequestCommandData { return v.data.clone() }
func (v CreatePullRequestCommand) Clone() CreatePullRequestCommand    { return v }
func (d CreatePullRequestCommandData) clone() CreatePullRequestCommandData {

	return d
}
func (d CreatePullRequestCommandData) validate() error {
	if !d.Base.Valid() {
		return invalid("d.Base")
	}
	if !d.Head.Valid() {
		return invalid("d.Head")
	}

	if !nonblank(d.Title) || !textValue(d.Title) || !textValue(d.Body) {
		return invalid("PR text")
	}
	return nil
}

// SaveLaunchCommand is an immutable semantic boundary value. Its zero is invalid.
type SaveLaunchCommand struct {
	data        SaveLaunchCommandData
	initialized bool
}

// SaveLaunchCommandData is a construction/copy value; NewSaveLaunchCommand validates and owns a copy.
type SaveLaunchCommandData struct {
	WorktreeID      domain.WorktreeID
	Selection       LaunchSelection
	Alias           string
	MakeDefault     bool
	ExpectedStorage StorageVersion
	Replacement     Optional[AliasReplacement]
}

func NewSaveLaunchCommand(d SaveLaunchCommandData) (SaveLaunchCommand, error) {
	if err := d.validate(); err != nil {
		return SaveLaunchCommand{}, err
	}
	return SaveLaunchCommand{data: d.clone(), initialized: true}, nil
}
func (v SaveLaunchCommand) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v SaveLaunchCommand) Data() SaveLaunchCommandData { return v.data.clone() }
func (v SaveLaunchCommand) Clone() SaveLaunchCommand    { return v }
func (d SaveLaunchCommandData) clone() SaveLaunchCommandData {

	return d
}
func (d SaveLaunchCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !validLaunchSelection(d.Selection) {
		return invalid("d.Selection")
	}

	if !d.ExpectedStorage.Valid() {
		return invalid("d.ExpectedStorage")
	}
	if item, ok := d.Replacement.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentSaveLaunchCommand(d); err != nil {
		return err
	}
	if !launchMatchesWorktree(d.Selection, d.WorktreeID) || !nonempty(d.Alias) || d.ExpectedStorage.Family() != RunConfig {
		return invalid("save launch")
	}
	if r, ok := d.Replacement.Value(); ok && (r.data.Alias != d.Alias || r.data.Expected != d.ExpectedStorage) {
		return invalid("replacement binding")
	}
	return nil
}

// StartLaunchCommand is an immutable semantic boundary value. Its zero is invalid.
type StartLaunchCommand struct {
	data        StartLaunchCommandData
	initialized bool
}

// StartLaunchCommandData is a construction/copy value; NewStartLaunchCommand validates and owns a copy.
type StartLaunchCommandData struct {
	WorktreeID domain.WorktreeID
	Selection  LaunchIntent
	Geometry   Geometry
}

func NewStartLaunchCommand(d StartLaunchCommandData) (StartLaunchCommand, error) {
	if err := d.validate(); err != nil {
		return StartLaunchCommand{}, err
	}
	return StartLaunchCommand{data: d.clone(), initialized: true}, nil
}
func (v StartLaunchCommand) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v StartLaunchCommand) Data() StartLaunchCommandData { return v.data.clone() }
func (v StartLaunchCommand) Clone() StartLaunchCommand    { return v }
func (d StartLaunchCommandData) clone() StartLaunchCommandData {

	return d
}
func (d StartLaunchCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !validLaunchIntent(d.Selection) {
		return invalid("d.Selection")
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}
	if err := consistentStartLaunchCommand(d); err != nil {
		return err
	}
	if s, ok := d.Selection.(SelectedLaunch); ok && !launchMatchesWorktree(s.data.Selection, d.WorktreeID) {
		return invalid("start launch scope")
	}
	return nil
}

// OpenTerminalCommand is an immutable semantic boundary value. Its zero is invalid.
type OpenTerminalCommand struct {
	data        OpenTerminalCommandData
	initialized bool
}

// OpenTerminalCommandData is a construction/copy value; NewOpenTerminalCommand validates and owns a copy.
type OpenTerminalCommandData struct {
	WorktreeID domain.WorktreeID
	Shell      ShellPolicy
	Geometry   Geometry
}

func NewOpenTerminalCommand(d OpenTerminalCommandData) (OpenTerminalCommand, error) {
	if err := d.validate(); err != nil {
		return OpenTerminalCommand{}, err
	}
	return OpenTerminalCommand{data: d.clone(), initialized: true}, nil
}
func (v OpenTerminalCommand) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v OpenTerminalCommand) Data() OpenTerminalCommandData { return v.data.clone() }
func (v OpenTerminalCommand) Clone() OpenTerminalCommand    { return v }
func (d OpenTerminalCommandData) clone() OpenTerminalCommandData {

	return d
}
func (d OpenTerminalCommandData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !validShellPolicy(d.Shell) {
		return invalid("d.Shell")
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}

	return nil
}

// WriteInputCommand is an immutable semantic boundary value. Its zero is invalid.
type WriteInputCommand struct {
	data        WriteInputCommandData
	initialized bool
}

// WriteInputCommandData is a construction/copy value; NewWriteInputCommand validates and owns a copy.
type WriteInputCommandData struct {
	SessionID domain.SessionID
	Bytes     []byte
}

func NewWriteInputCommand(d WriteInputCommandData) (WriteInputCommand, error) {
	if err := d.validate(); err != nil {
		return WriteInputCommand{}, err
	}
	return WriteInputCommand{data: d.clone(), initialized: true}, nil
}
func (v WriteInputCommand) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v WriteInputCommand) Data() WriteInputCommandData { return v.data.clone() }
func (v WriteInputCommand) Clone() WriteInputCommand    { return v }
func (d WriteInputCommandData) clone() WriteInputCommandData {
	d.Bytes = cloneSlice(d.Bytes)
	return d
}
func (d WriteInputCommandData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	if len(d.Bytes) == 0 || len(d.Bytes) > 65536 {
		return invalid("write input bound")
	}
	return nil
}

// ResizeSessionCommand is an immutable semantic boundary value. Its zero is invalid.
type ResizeSessionCommand struct {
	data        ResizeSessionCommandData
	initialized bool
}

// ResizeSessionCommandData is a construction/copy value; NewResizeSessionCommand validates and owns a copy.
type ResizeSessionCommandData struct {
	SessionID domain.SessionID
	Geometry  Geometry
}

func NewResizeSessionCommand(d ResizeSessionCommandData) (ResizeSessionCommand, error) {
	if err := d.validate(); err != nil {
		return ResizeSessionCommand{}, err
	}
	return ResizeSessionCommand{data: d.clone(), initialized: true}, nil
}
func (v ResizeSessionCommand) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ResizeSessionCommand) Data() ResizeSessionCommandData { return v.data.clone() }
func (v ResizeSessionCommand) Clone() ResizeSessionCommand    { return v }
func (d ResizeSessionCommandData) clone() ResizeSessionCommandData {

	return d
}
func (d ResizeSessionCommandData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}

	return nil
}

// InterruptSessionCommand is an immutable semantic boundary value. Its zero is invalid.
type InterruptSessionCommand struct {
	data        InterruptSessionCommandData
	initialized bool
}

// InterruptSessionCommandData is a construction/copy value; NewInterruptSessionCommand validates and owns a copy.
type InterruptSessionCommandData struct {
	SessionID domain.SessionID
}

func NewInterruptSessionCommand(d InterruptSessionCommandData) (InterruptSessionCommand, error) {
	if err := d.validate(); err != nil {
		return InterruptSessionCommand{}, err
	}
	return InterruptSessionCommand{data: d.clone(), initialized: true}, nil
}
func (v InterruptSessionCommand) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v InterruptSessionCommand) Data() InterruptSessionCommandData { return v.data.clone() }
func (v InterruptSessionCommand) Clone() InterruptSessionCommand    { return v }
func (d InterruptSessionCommandData) clone() InterruptSessionCommandData {

	return d
}
func (d InterruptSessionCommandData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	return nil
}

// StopSessionCommand is an immutable semantic boundary value. Its zero is invalid.
type StopSessionCommand struct {
	data        StopSessionCommandData
	initialized bool
}

// StopSessionCommandData is a construction/copy value; NewStopSessionCommand validates and owns a copy.
type StopSessionCommandData struct {
	SessionID domain.SessionID
}

func NewStopSessionCommand(d StopSessionCommandData) (StopSessionCommand, error) {
	if err := d.validate(); err != nil {
		return StopSessionCommand{}, err
	}
	return StopSessionCommand{data: d.clone(), initialized: true}, nil
}
func (v StopSessionCommand) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v StopSessionCommand) Data() StopSessionCommandData { return v.data.clone() }
func (v StopSessionCommand) Clone() StopSessionCommand    { return v }
func (d StopSessionCommandData) clone() StopSessionCommandData {

	return d
}
func (d StopSessionCommandData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	return nil
}

// RestartSessionCommand is an immutable semantic boundary value. Its zero is invalid.
type RestartSessionCommand struct {
	data        RestartSessionCommandData
	initialized bool
}

// RestartSessionCommandData is a construction/copy value; NewRestartSessionCommand validates and owns a copy.
type RestartSessionCommandData struct {
	SessionID domain.SessionID
	Geometry  Optional[Geometry]
}

func NewRestartSessionCommand(d RestartSessionCommandData) (RestartSessionCommand, error) {
	if err := d.validate(); err != nil {
		return RestartSessionCommand{}, err
	}
	return RestartSessionCommand{data: d.clone(), initialized: true}, nil
}
func (v RestartSessionCommand) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v RestartSessionCommand) Data() RestartSessionCommandData { return v.data.clone() }
func (v RestartSessionCommand) Clone() RestartSessionCommand    { return v }
func (d RestartSessionCommandData) clone() RestartSessionCommandData {

	return d
}
func (d RestartSessionCommandData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if item, ok := d.Geometry.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// NormalizedRecovery is an immutable semantic boundary value. Its zero is invalid.
type NormalizedRecovery struct {
	data        NormalizedRecoveryData
	initialized bool
}

// NormalizedRecoveryData is a construction/copy value; NewNormalizedRecovery validates and owns a copy.
type NormalizedRecoveryData struct {
	Record        RecoveryRecord
	StorageDetail Optional[StorageRecovery]
}

func NewNormalizedRecovery(d NormalizedRecoveryData) (NormalizedRecovery, error) {
	if err := d.validate(); err != nil {
		return NormalizedRecovery{}, err
	}
	return NormalizedRecovery{data: d.clone(), initialized: true}, nil
}
func (v NormalizedRecovery) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v NormalizedRecovery) Data() NormalizedRecoveryData { return v.data.clone() }
func (v NormalizedRecovery) Clone() NormalizedRecovery    { return v }
func (d NormalizedRecoveryData) clone() NormalizedRecoveryData {

	return d
}
func (d NormalizedRecoveryData) validate() error {
	if !d.Record.Valid() {
		return invalid("d.Record")
	}
	if item, ok := d.StorageDetail.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if s, ok := d.StorageDetail.Value(); ok && !sameRecoveryRecord(d.Record, s.data.Record) {
		return invalid("recovery detail association")
	}
	return nil
}

// OutcomeEnvelope is an immutable semantic boundary value. Its zero is invalid.
type OutcomeEnvelope struct {
	data        OutcomeEnvelopeData
	initialized bool
}

// OutcomeEnvelopeData is a construction/copy value; NewOutcomeEnvelope validates and owns a copy.
type OutcomeEnvelopeData struct {
	Effects               EffectReport
	Diagnostics           []Diagnostic
	Recovery              []NormalizedRecovery
	CancellationRequested bool
}

func NewOutcomeEnvelope(d OutcomeEnvelopeData) (OutcomeEnvelope, error) {
	if err := d.validate(); err != nil {
		return OutcomeEnvelope{}, err
	}
	return OutcomeEnvelope{data: d.clone(), initialized: true}, nil
}
func (v OutcomeEnvelope) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v OutcomeEnvelope) Data() OutcomeEnvelopeData { return v.data.clone() }
func (v OutcomeEnvelope) Clone() OutcomeEnvelope    { return v }
func (d OutcomeEnvelopeData) clone() OutcomeEnvelopeData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d OutcomeEnvelopeData) validate() error {
	if !d.Effects.Valid() {
		return invalid("d.Effects")
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

	return validateNormalizedRecovery(d.Effects, d.Recovery)
}

// StashPopCompleted is an immutable semantic boundary value. Its zero is invalid.
type StashPopCompleted struct {
	data        StashPopCompletedData
	initialized bool
}

// StashPopCompletedData is a construction/copy value; NewStashPopCompleted validates and owns a copy.
type StashPopCompletedData struct {
	Stash domain.StashID
	Apply GitMutationResult
	Drop  GitMutationResult
}

func NewStashPopCompleted(d StashPopCompletedData) (StashPopCompleted, error) {
	if err := d.validate(); err != nil {
		return StashPopCompleted{}, err
	}
	return StashPopCompleted{data: d.clone(), initialized: true}, nil
}
func (v StashPopCompleted) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v StashPopCompleted) Data() StashPopCompletedData { return v.data.clone() }
func (v StashPopCompleted) Clone() StashPopCompleted    { return v }
func (d StashPopCompletedData) clone() StashPopCompletedData {

	return d
}
func (d StashPopCompletedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Apply.Valid() {
		return invalid("d.Apply")
	}
	if !d.Drop.Valid() {
		return invalid("d.Drop")
	}
	if err := consistentStashPopCompleted(d); err != nil {
		return err
	}
	return nil
}

// AppliedStashRetained is an immutable semantic boundary value. Its zero is invalid.
type AppliedStashRetained struct {
	data        AppliedStashRetainedData
	initialized bool
}

// AppliedStashRetainedData is a construction/copy value; NewAppliedStashRetained validates and owns a copy.
type AppliedStashRetainedData struct {
	Stash  domain.StashID
	Apply  GitMutationResult
	Drop   Optional[GitMutationResult]
	Reason Diagnostic
}

func NewAppliedStashRetained(d AppliedStashRetainedData) (AppliedStashRetained, error) {
	if err := d.validate(); err != nil {
		return AppliedStashRetained{}, err
	}
	return AppliedStashRetained{data: d.clone(), initialized: true}, nil
}
func (v AppliedStashRetained) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v AppliedStashRetained) Data() AppliedStashRetainedData { return v.data.clone() }
func (v AppliedStashRetained) Clone() AppliedStashRetained    { return v }
func (d AppliedStashRetainedData) clone() AppliedStashRetainedData {

	return d
}
func (d AppliedStashRetainedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Apply.Valid() {
		return invalid("d.Apply")
	}
	if item, ok := d.Drop.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	if err := consistentAppliedStashRetained(d); err != nil {
		return err
	}
	return nil
}

// StashPopNotApplied is an immutable semantic boundary value. Its zero is invalid.
type StashPopNotApplied struct {
	data        StashPopNotAppliedData
	initialized bool
}

// StashPopNotAppliedData is a construction/copy value; NewStashPopNotApplied validates and owns a copy.
type StashPopNotAppliedData struct {
	Stash domain.StashID
	Apply GitMutationResult
}

func NewStashPopNotApplied(d StashPopNotAppliedData) (StashPopNotApplied, error) {
	if err := d.validate(); err != nil {
		return StashPopNotApplied{}, err
	}
	return StashPopNotApplied{data: d.clone(), initialized: true}, nil
}
func (v StashPopNotApplied) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v StashPopNotApplied) Data() StashPopNotAppliedData { return v.data.clone() }
func (v StashPopNotApplied) Clone() StashPopNotApplied    { return v }
func (d StashPopNotAppliedData) clone() StashPopNotAppliedData {

	return d
}
func (d StashPopNotAppliedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Apply.Valid() {
		return invalid("d.Apply")
	}
	if err := consistentStashPopNotApplied(d); err != nil {
		return err
	}
	return nil
}

// ActivateWorktreeResult is an immutable semantic boundary value. Its zero is invalid.
type ActivateWorktreeResult struct {
	data        ActivateWorktreeResultData
	initialized bool
}

// ActivateWorktreeResultData is a construction/copy value; NewActivateWorktreeResult validates and owns a copy.
type ActivateWorktreeResultData struct {
	Context ActiveContext
	Storage StorageCommitResult
	Outcome OutcomeEnvelope
}

func NewActivateWorktreeResult(d ActivateWorktreeResultData) (ActivateWorktreeResult, error) {
	if err := d.validate(); err != nil {
		return ActivateWorktreeResult{}, err
	}
	return ActivateWorktreeResult{data: d.clone(), initialized: true}, nil
}
func (v ActivateWorktreeResult) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v ActivateWorktreeResult) Data() ActivateWorktreeResultData { return v.data.clone() }
func (v ActivateWorktreeResult) Clone() ActivateWorktreeResult    { return v }
func (d ActivateWorktreeResultData) clone() ActivateWorktreeResultData {

	return d
}
func (d ActivateWorktreeResultData) validate() error {
	if !d.Context.Valid() {
		return invalid("d.Context")
	}
	if !d.Storage.Valid() {
		return invalid("d.Storage")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateActivateWorktreeResultEvidence(d); err != nil {
		return err
	}
	if err := consistentActivateWorktreeResult(d); err != nil {
		return err
	}
	return nil
}

// SaveNavigationResult is an immutable semantic boundary value. Its zero is invalid.
type SaveNavigationResult struct {
	data        SaveNavigationResultData
	initialized bool
}

// SaveNavigationResultData is a construction/copy value; NewSaveNavigationResult validates and owns a copy.
type SaveNavigationResultData struct {
	Intent           NavigationIntent
	EffectiveVersion Optional[StorageVersion]
	Storage          StorageCommitResult
	Outcome          OutcomeEnvelope
}

func NewSaveNavigationResult(d SaveNavigationResultData) (SaveNavigationResult, error) {
	if err := d.validate(); err != nil {
		return SaveNavigationResult{}, err
	}
	return SaveNavigationResult{data: d.clone(), initialized: true}, nil
}
func (v SaveNavigationResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v SaveNavigationResult) Data() SaveNavigationResultData { return v.data.clone() }
func (v SaveNavigationResult) Clone() SaveNavigationResult    { return v }
func (d SaveNavigationResultData) clone() SaveNavigationResultData {

	return d
}
func (d SaveNavigationResultData) validate() error {
	if !d.Intent.Valid() {
		return invalid("d.Intent")
	}
	if item, ok := d.EffectiveVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Storage.Valid() {
		return invalid("d.Storage")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateSaveNavigationResultEvidence(d); err != nil {
		return err
	}
	if err := consistentSaveNavigationResult(d); err != nil {
		return err
	}
	return nil
}

// CreateWorktreeResult is an immutable semantic boundary value. Its zero is invalid.
type CreateWorktreeResult struct {
	data        CreateWorktreeResultData
	initialized bool
}

// CreateWorktreeResultData is a construction/copy value; NewCreateWorktreeResult validates and owns a copy.
type CreateWorktreeResultData struct {
	Git        GitMutationResult
	WorktreeID Optional[domain.WorktreeID]
	Scope      Optional[WorktreeScope]
	Head       Optional[domain.Head]
	Outcome    OutcomeEnvelope
}

func NewCreateWorktreeResult(d CreateWorktreeResultData) (CreateWorktreeResult, error) {
	if err := d.validate(); err != nil {
		return CreateWorktreeResult{}, err
	}
	return CreateWorktreeResult{data: d.clone(), initialized: true}, nil
}
func (v CreateWorktreeResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v CreateWorktreeResult) Data() CreateWorktreeResultData { return v.data.clone() }
func (v CreateWorktreeResult) Clone() CreateWorktreeResult    { return v }
func (d CreateWorktreeResultData) clone() CreateWorktreeResultData {

	return d
}
func (d CreateWorktreeResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.WorktreeID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Scope.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateCreateWorktreeResultEvidence(d); err != nil {
		return err
	}
	if err := consistentCreateWorktreeResult(d); err != nil {
		return err
	}
	return nil
}

// RetargetWorktreeResult is an immutable semantic boundary value. Its zero is invalid.
type RetargetWorktreeResult struct {
	data        RetargetWorktreeResultData
	initialized bool
}

// RetargetWorktreeResultData is a construction/copy value; NewRetargetWorktreeResult validates and owns a copy.
type RetargetWorktreeResultData struct {
	Git     GitMutationResult
	Scope   Optional[WorktreeScope]
	Head    Optional[domain.Head]
	Outcome OutcomeEnvelope
}

func NewRetargetWorktreeResult(d RetargetWorktreeResultData) (RetargetWorktreeResult, error) {
	if err := d.validate(); err != nil {
		return RetargetWorktreeResult{}, err
	}
	return RetargetWorktreeResult{data: d.clone(), initialized: true}, nil
}
func (v RetargetWorktreeResult) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v RetargetWorktreeResult) Data() RetargetWorktreeResultData { return v.data.clone() }
func (v RetargetWorktreeResult) Clone() RetargetWorktreeResult    { return v }
func (d RetargetWorktreeResultData) clone() RetargetWorktreeResultData {

	return d
}
func (d RetargetWorktreeResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Scope.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateRetargetWorktreeResultEvidence(d); err != nil {
		return err
	}
	if err := consistentRetargetWorktreeResult(d); err != nil {
		return err
	}
	return nil
}

// DeployResult is an immutable semantic boundary value. Its zero is invalid.
type DeployResult struct {
	data        DeployResultData
	initialized bool
}

// DeployResultData is a construction/copy value; NewDeployResult validates and owns a copy.
type DeployResultData struct {
	Target       domain.ExactTarget
	Resolution   Optional[ExactLocalResolution]
	Head         Optional[domain.Head]
	CreatedStash Optional[domain.StashID]
	Steps        []GitMutationResult
	Outcome      OutcomeEnvelope
}

func NewDeployResult(d DeployResultData) (DeployResult, error) {
	if err := d.validate(); err != nil {
		return DeployResult{}, err
	}
	return DeployResult{data: d.clone(), initialized: true}, nil
}
func (v DeployResult) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v DeployResult) Data() DeployResultData { return v.data.clone() }
func (v DeployResult) Clone() DeployResult    { return v }
func (d DeployResultData) clone() DeployResultData {
	d.Steps = cloneSlice(d.Steps)
	return d
}
func (d DeployResultData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if item, ok := d.Resolution.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.CreatedStash.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Steps {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateDeployResultEvidence(d); err != nil {
		return err
	}
	if err := consistentDeployResult(d); err != nil {
		return err
	}
	return nil
}

// CreateBranchResult is an immutable semantic boundary value. Its zero is invalid.
type CreateBranchResult struct {
	data        CreateBranchResultData
	initialized bool
}

// CreateBranchResultData is a construction/copy value; NewCreateBranchResult validates and owns a copy.
type CreateBranchResultData struct {
	Git     GitMutationResult
	Branch  Optional[domain.BranchID]
	Head    Optional[domain.Head]
	Outcome OutcomeEnvelope
}

func NewCreateBranchResult(d CreateBranchResultData) (CreateBranchResult, error) {
	if err := d.validate(); err != nil {
		return CreateBranchResult{}, err
	}
	return CreateBranchResult{data: d.clone(), initialized: true}, nil
}
func (v CreateBranchResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v CreateBranchResult) Data() CreateBranchResultData { return v.data.clone() }
func (v CreateBranchResult) Clone() CreateBranchResult    { return v }
func (d CreateBranchResultData) clone() CreateBranchResultData {

	return d
}
func (d CreateBranchResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Branch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateCreateBranchResultEvidence(d); err != nil {
		return err
	}
	if err := consistentCreateBranchResult(d); err != nil {
		return err
	}
	return nil
}

// PullResult is an immutable semantic boundary value. Its zero is invalid.
type PullResult struct {
	data        PullResultData
	initialized bool
}

// PullResultData is a construction/copy value; NewPullResult validates and owns a copy.
type PullResultData struct {
	Fetch       Optional[FetchResult]
	Endpoint    Optional[domain.Revision]
	FastForward Optional[GitMutationResult]
	Outcome     OutcomeEnvelope
}

func NewPullResult(d PullResultData) (PullResult, error) {
	if err := d.validate(); err != nil {
		return PullResult{}, err
	}
	return PullResult{data: d.clone(), initialized: true}, nil
}
func (v PullResult) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v PullResult) Data() PullResultData { return v.data.clone() }
func (v PullResult) Clone() PullResult    { return v }
func (d PullResultData) clone() PullResultData {

	return d
}
func (d PullResultData) validate() error {
	if item, ok := d.Fetch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Endpoint.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.FastForward.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validatePullResultEvidence(d); err != nil {
		return err
	}
	if err := consistentPullResult(d); err != nil {
		return err
	}
	return nil
}

// PushResult is an immutable semantic boundary value. Its zero is invalid.
type PushResult struct {
	data        PushResultData
	initialized bool
}

// PushResultData is a construction/copy value; NewPushResult validates and owns a copy.
type PushResultData struct {
	Git            GitMutationResult
	RemoteEffect   EffectReport
	UpstreamEffect EffectReport
	Outcome        OutcomeEnvelope
}

func NewPushResult(d PushResultData) (PushResult, error) {
	if err := d.validate(); err != nil {
		return PushResult{}, err
	}
	return PushResult{data: d.clone(), initialized: true}, nil
}
func (v PushResult) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v PushResult) Data() PushResultData { return v.data.clone() }
func (v PushResult) Clone() PushResult    { return v }
func (d PushResultData) clone() PushResultData {

	return d
}
func (d PushResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if !d.RemoteEffect.Valid() {
		return invalid("d.RemoteEffect")
	}
	if !d.UpstreamEffect.Valid() {
		return invalid("d.UpstreamEffect")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validatePushResultEvidence(d); err != nil {
		return err
	}
	if err := consistentPushResult(d); err != nil {
		return err
	}
	return nil
}

// StagePathsResult is an immutable semantic boundary value. Its zero is invalid.
type StagePathsResult struct {
	data        StagePathsResultData
	initialized bool
}

// StagePathsResultData is a construction/copy value; NewStagePathsResult validates and owns a copy.
type StagePathsResultData struct {
	Git     GitMutationResult
	Status  Optional[StatusFacts]
	Outcome OutcomeEnvelope
}

func NewStagePathsResult(d StagePathsResultData) (StagePathsResult, error) {
	if err := d.validate(); err != nil {
		return StagePathsResult{}, err
	}
	return StagePathsResult{data: d.clone(), initialized: true}, nil
}
func (v StagePathsResult) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v StagePathsResult) Data() StagePathsResultData { return v.data.clone() }
func (v StagePathsResult) Clone() StagePathsResult    { return v }
func (d StagePathsResultData) clone() StagePathsResultData {

	return d
}
func (d StagePathsResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStagePathsResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStagePathsResult(d); err != nil {
		return err
	}
	return nil
}

// UnstagePathsResult is an immutable semantic boundary value. Its zero is invalid.
type UnstagePathsResult struct {
	data        UnstagePathsResultData
	initialized bool
}

// UnstagePathsResultData is a construction/copy value; NewUnstagePathsResult validates and owns a copy.
type UnstagePathsResultData struct {
	Git     GitMutationResult
	Status  Optional[StatusFacts]
	Outcome OutcomeEnvelope
}

func NewUnstagePathsResult(d UnstagePathsResultData) (UnstagePathsResult, error) {
	if err := d.validate(); err != nil {
		return UnstagePathsResult{}, err
	}
	return UnstagePathsResult{data: d.clone(), initialized: true}, nil
}
func (v UnstagePathsResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v UnstagePathsResult) Data() UnstagePathsResultData { return v.data.clone() }
func (v UnstagePathsResult) Clone() UnstagePathsResult    { return v }
func (d UnstagePathsResultData) clone() UnstagePathsResultData {

	return d
}
func (d UnstagePathsResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateUnstagePathsResultEvidence(d); err != nil {
		return err
	}
	if err := consistentUnstagePathsResult(d); err != nil {
		return err
	}
	return nil
}

// StageAllResult is an immutable semantic boundary value. Its zero is invalid.
type StageAllResult struct {
	data        StageAllResultData
	initialized bool
}

// StageAllResultData is a construction/copy value; NewStageAllResult validates and owns a copy.
type StageAllResultData struct {
	Git     GitMutationResult
	Status  Optional[StatusFacts]
	Outcome OutcomeEnvelope
}

func NewStageAllResult(d StageAllResultData) (StageAllResult, error) {
	if err := d.validate(); err != nil {
		return StageAllResult{}, err
	}
	return StageAllResult{data: d.clone(), initialized: true}, nil
}
func (v StageAllResult) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v StageAllResult) Data() StageAllResultData { return v.data.clone() }
func (v StageAllResult) Clone() StageAllResult    { return v }
func (d StageAllResultData) clone() StageAllResultData {

	return d
}
func (d StageAllResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStageAllResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStageAllResult(d); err != nil {
		return err
	}
	return nil
}

// CommitResult is an immutable semantic boundary value. Its zero is invalid.
type CommitResult struct {
	data        CommitResultData
	initialized bool
}

// CommitResultData is a construction/copy value; NewCommitResult validates and owns a copy.
type CommitResultData struct {
	Git       GitMutationResult
	Revision  Optional[domain.Revision]
	Candidate Optional[CommitCandidateFacts]
	Outcome   OutcomeEnvelope
}

func NewCommitResult(d CommitResultData) (CommitResult, error) {
	if err := d.validate(); err != nil {
		return CommitResult{}, err
	}
	return CommitResult{data: d.clone(), initialized: true}, nil
}
func (v CommitResult) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v CommitResult) Data() CommitResultData { return v.data.clone() }
func (v CommitResult) Clone() CommitResult    { return v }
func (d CommitResultData) clone() CommitResultData {

	return d
}
func (d CommitResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Revision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Candidate.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateCommitResultEvidence(d); err != nil {
		return err
	}
	if err := consistentCommitResult(d); err != nil {
		return err
	}
	return nil
}

// StageAllAndCommitResult is an immutable semantic boundary value. Its zero is invalid.
type StageAllAndCommitResult struct {
	data        StageAllAndCommitResultData
	initialized bool
}

// StageAllAndCommitResultData is a construction/copy value; NewStageAllAndCommitResult validates and owns a copy.
type StageAllAndCommitResultData struct {
	Stage     Optional[GitMutationResult]
	Commit    Optional[GitMutationResult]
	Revision  Optional[domain.Revision]
	Candidate Optional[CommitCandidateFacts]
	Outcome   OutcomeEnvelope
}

func NewStageAllAndCommitResult(d StageAllAndCommitResultData) (StageAllAndCommitResult, error) {
	if err := d.validate(); err != nil {
		return StageAllAndCommitResult{}, err
	}
	return StageAllAndCommitResult{data: d.clone(), initialized: true}, nil
}
func (v StageAllAndCommitResult) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v StageAllAndCommitResult) Data() StageAllAndCommitResultData { return v.data.clone() }
func (v StageAllAndCommitResult) Clone() StageAllAndCommitResult    { return v }
func (d StageAllAndCommitResultData) clone() StageAllAndCommitResultData {

	return d
}
func (d StageAllAndCommitResultData) validate() error {
	if item, ok := d.Stage.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Commit.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Revision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Candidate.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStageAllAndCommitResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStageAllAndCommitResult(d); err != nil {
		return err
	}
	return nil
}

// RestoreTrackedResult is an immutable semantic boundary value. Its zero is invalid.
type RestoreTrackedResult struct {
	data        RestoreTrackedResultData
	initialized bool
}

// RestoreTrackedResultData is a construction/copy value; NewRestoreTrackedResult validates and owns a copy.
type RestoreTrackedResultData struct {
	Git     GitMutationResult
	Status  Optional[StatusFacts]
	Outcome OutcomeEnvelope
}

func NewRestoreTrackedResult(d RestoreTrackedResultData) (RestoreTrackedResult, error) {
	if err := d.validate(); err != nil {
		return RestoreTrackedResult{}, err
	}
	return RestoreTrackedResult{data: d.clone(), initialized: true}, nil
}
func (v RestoreTrackedResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v RestoreTrackedResult) Data() RestoreTrackedResultData { return v.data.clone() }
func (v RestoreTrackedResult) Clone() RestoreTrackedResult    { return v }
func (d RestoreTrackedResultData) clone() RestoreTrackedResultData {

	return d
}
func (d RestoreTrackedResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateRestoreTrackedResultEvidence(d); err != nil {
		return err
	}
	if err := consistentRestoreTrackedResult(d); err != nil {
		return err
	}
	return nil
}

// StashCreateResult is an immutable semantic boundary value. Its zero is invalid.
type StashCreateResult struct {
	data        StashCreateResultData
	initialized bool
}

// StashCreateResultData is a construction/copy value; NewStashCreateResult validates and owns a copy.
type StashCreateResultData struct {
	Git     GitMutationResult
	Stash   Optional[domain.StashID]
	Outcome OutcomeEnvelope
}

func NewStashCreateResult(d StashCreateResultData) (StashCreateResult, error) {
	if err := d.validate(); err != nil {
		return StashCreateResult{}, err
	}
	return StashCreateResult{data: d.clone(), initialized: true}, nil
}
func (v StashCreateResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v StashCreateResult) Data() StashCreateResultData { return v.data.clone() }
func (v StashCreateResult) Clone() StashCreateResult    { return v }
func (d StashCreateResultData) clone() StashCreateResultData {

	return d
}
func (d StashCreateResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if item, ok := d.Stash.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStashCreateResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStashCreateResult(d); err != nil {
		return err
	}
	return nil
}

// StashApplyResult is an immutable semantic boundary value. Its zero is invalid.
type StashApplyResult struct {
	data        StashApplyResultData
	initialized bool
}

// StashApplyResultData is a construction/copy value; NewStashApplyResult validates and owns a copy.
type StashApplyResultData struct {
	Git     GitMutationResult
	Stash   domain.StashID
	Outcome OutcomeEnvelope
}

func NewStashApplyResult(d StashApplyResultData) (StashApplyResult, error) {
	if err := d.validate(); err != nil {
		return StashApplyResult{}, err
	}
	return StashApplyResult{data: d.clone(), initialized: true}, nil
}
func (v StashApplyResult) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v StashApplyResult) Data() StashApplyResultData { return v.data.clone() }
func (v StashApplyResult) Clone() StashApplyResult    { return v }
func (d StashApplyResultData) clone() StashApplyResultData {

	return d
}
func (d StashApplyResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStashApplyResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStashApplyResult(d); err != nil {
		return err
	}
	return nil
}

// StashPopResult is an immutable semantic boundary value. Its zero is invalid.
type StashPopResult struct {
	data        StashPopResultData
	initialized bool
}

// StashPopResultData is a construction/copy value; NewStashPopResult validates and owns a copy.
type StashPopResultData struct {
	Outcome  StashPopOutcome
	Envelope OutcomeEnvelope
}

func NewStashPopResult(d StashPopResultData) (StashPopResult, error) {
	if err := d.validate(); err != nil {
		return StashPopResult{}, err
	}
	return StashPopResult{data: d.clone(), initialized: true}, nil
}
func (v StashPopResult) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v StashPopResult) Data() StashPopResultData { return v.data.clone() }
func (v StashPopResult) Clone() StashPopResult    { return v }
func (d StashPopResultData) clone() StashPopResultData {

	return d
}
func (d StashPopResultData) validate() error {
	if !validStashPopOutcome(d.Outcome) {
		return invalid("d.Outcome")
	}
	if !d.Envelope.Valid() {
		return invalid("d.Envelope")
	}
	if err := validateStashPopResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStashPopResult(d); err != nil {
		return err
	}
	return nil
}

// StashDropResult is an immutable semantic boundary value. Its zero is invalid.
type StashDropResult struct {
	data        StashDropResultData
	initialized bool
}

// StashDropResultData is a construction/copy value; NewStashDropResult validates and owns a copy.
type StashDropResultData struct {
	Git     GitMutationResult
	Stash   domain.StashID
	Outcome OutcomeEnvelope
}

func NewStashDropResult(d StashDropResultData) (StashDropResult, error) {
	if err := d.validate(); err != nil {
		return StashDropResult{}, err
	}
	return StashDropResult{data: d.clone(), initialized: true}, nil
}
func (v StashDropResult) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v StashDropResult) Data() StashDropResultData { return v.data.clone() }
func (v StashDropResult) Clone() StashDropResult    { return v }
func (d StashDropResultData) clone() StashDropResultData {

	return d
}
func (d StashDropResultData) validate() error {
	if !d.Git.Valid() {
		return invalid("d.Git")
	}
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStashDropResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStashDropResult(d); err != nil {
		return err
	}
	return nil
}

// SaveLaunchResult is an immutable semantic boundary value. Its zero is invalid.
type SaveLaunchResult struct {
	data        SaveLaunchResultData
	initialized bool
}

// SaveLaunchResultData is a construction/copy value; NewSaveLaunchResult validates and owns a copy.
type SaveLaunchResultData struct {
	Entry   SavedLaunchEntry
	Default StoredField[string]
	Storage StorageCommitResult
	Outcome OutcomeEnvelope
}

func NewSaveLaunchResult(d SaveLaunchResultData) (SaveLaunchResult, error) {
	if err := d.validate(); err != nil {
		return SaveLaunchResult{}, err
	}
	return SaveLaunchResult{data: d.clone(), initialized: true}, nil
}
func (v SaveLaunchResult) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v SaveLaunchResult) Data() SaveLaunchResultData { return v.data.clone() }
func (v SaveLaunchResult) Clone() SaveLaunchResult    { return v }
func (d SaveLaunchResultData) clone() SaveLaunchResultData {

	return d
}
func (d SaveLaunchResultData) validate() error {
	if !d.Entry.Valid() {
		return invalid("d.Entry")
	}
	if !d.Default.Valid() {
		return invalid("field presence")
	}
	if !d.Storage.Valid() {
		return invalid("d.Storage")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateSaveLaunchResultEvidence(d); err != nil {
		return err
	}
	if err := consistentSaveLaunchResult(d); err != nil {
		return err
	}
	return nil
}

// StartLaunchResult is an immutable semantic boundary value. Its zero is invalid.
type StartLaunchResult struct {
	data        StartLaunchResultData
	initialized bool
}

// StartLaunchResultData is a construction/copy value; NewStartLaunchResult validates and owns a copy.
type StartLaunchResultData struct {
	Start   SessionStartResult
	Outcome OutcomeEnvelope
}

func NewStartLaunchResult(d StartLaunchResultData) (StartLaunchResult, error) {
	if err := d.validate(); err != nil {
		return StartLaunchResult{}, err
	}
	return StartLaunchResult{data: d.clone(), initialized: true}, nil
}
func (v StartLaunchResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v StartLaunchResult) Data() StartLaunchResultData { return v.data.clone() }
func (v StartLaunchResult) Clone() StartLaunchResult    { return v }
func (d StartLaunchResultData) clone() StartLaunchResultData {

	return d
}
func (d StartLaunchResultData) validate() error {
	if !d.Start.Valid() {
		return invalid("d.Start")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStartLaunchResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStartLaunchResult(d); err != nil {
		return err
	}
	return nil
}

// OpenTerminalResult is an immutable semantic boundary value. Its zero is invalid.
type OpenTerminalResult struct {
	data        OpenTerminalResultData
	initialized bool
}

// OpenTerminalResultData is a construction/copy value; NewOpenTerminalResult validates and owns a copy.
type OpenTerminalResultData struct {
	Start   SessionStartResult
	Outcome OutcomeEnvelope
}

func NewOpenTerminalResult(d OpenTerminalResultData) (OpenTerminalResult, error) {
	if err := d.validate(); err != nil {
		return OpenTerminalResult{}, err
	}
	return OpenTerminalResult{data: d.clone(), initialized: true}, nil
}
func (v OpenTerminalResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v OpenTerminalResult) Data() OpenTerminalResultData { return v.data.clone() }
func (v OpenTerminalResult) Clone() OpenTerminalResult    { return v }
func (d OpenTerminalResultData) clone() OpenTerminalResultData {

	return d
}
func (d OpenTerminalResultData) validate() error {
	if !d.Start.Valid() {
		return invalid("d.Start")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateOpenTerminalResultEvidence(d); err != nil {
		return err
	}
	if err := consistentOpenTerminalResult(d); err != nil {
		return err
	}
	return nil
}

// WriteInputResult is an immutable semantic boundary value. Its zero is invalid.
type WriteInputResult struct {
	data        WriteInputResultData
	initialized bool
}

// WriteInputResultData is a construction/copy value; NewWriteInputResult validates and owns a copy.
type WriteInputResultData struct {
	Write   SessionWriteResult
	Outcome OutcomeEnvelope
}

func NewWriteInputResult(d WriteInputResultData) (WriteInputResult, error) {
	if err := d.validate(); err != nil {
		return WriteInputResult{}, err
	}
	return WriteInputResult{data: d.clone(), initialized: true}, nil
}
func (v WriteInputResult) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v WriteInputResult) Data() WriteInputResultData { return v.data.clone() }
func (v WriteInputResult) Clone() WriteInputResult    { return v }
func (d WriteInputResultData) clone() WriteInputResultData {

	return d
}
func (d WriteInputResultData) validate() error {
	if !d.Write.Valid() {
		return invalid("d.Write")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateWriteInputResultEvidence(d); err != nil {
		return err
	}
	if err := consistentWriteInputResult(d); err != nil {
		return err
	}
	return nil
}

// ResizeSessionResult is an immutable semantic boundary value. Its zero is invalid.
type ResizeSessionResult struct {
	data        ResizeSessionResultData
	initialized bool
}

// ResizeSessionResultData is a construction/copy value; NewResizeSessionResult validates and owns a copy.
type ResizeSessionResultData struct {
	Control SessionControlResult
	Outcome OutcomeEnvelope
}

func NewResizeSessionResult(d ResizeSessionResultData) (ResizeSessionResult, error) {
	if err := d.validate(); err != nil {
		return ResizeSessionResult{}, err
	}
	return ResizeSessionResult{data: d.clone(), initialized: true}, nil
}
func (v ResizeSessionResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ResizeSessionResult) Data() ResizeSessionResultData { return v.data.clone() }
func (v ResizeSessionResult) Clone() ResizeSessionResult    { return v }
func (d ResizeSessionResultData) clone() ResizeSessionResultData {

	return d
}
func (d ResizeSessionResultData) validate() error {
	if !d.Control.Valid() {
		return invalid("d.Control")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateResizeSessionResultEvidence(d); err != nil {
		return err
	}
	if err := consistentResizeSessionResult(d); err != nil {
		return err
	}
	return nil
}

// InterruptSessionResult is an immutable semantic boundary value. Its zero is invalid.
type InterruptSessionResult struct {
	data        InterruptSessionResultData
	initialized bool
}

// InterruptSessionResultData is a construction/copy value; NewInterruptSessionResult validates and owns a copy.
type InterruptSessionResultData struct {
	Control SessionControlResult
	Outcome OutcomeEnvelope
}

func NewInterruptSessionResult(d InterruptSessionResultData) (InterruptSessionResult, error) {
	if err := d.validate(); err != nil {
		return InterruptSessionResult{}, err
	}
	return InterruptSessionResult{data: d.clone(), initialized: true}, nil
}
func (v InterruptSessionResult) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v InterruptSessionResult) Data() InterruptSessionResultData { return v.data.clone() }
func (v InterruptSessionResult) Clone() InterruptSessionResult    { return v }
func (d InterruptSessionResultData) clone() InterruptSessionResultData {

	return d
}
func (d InterruptSessionResultData) validate() error {
	if !d.Control.Valid() {
		return invalid("d.Control")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateInterruptSessionResultEvidence(d); err != nil {
		return err
	}
	if err := consistentInterruptSessionResult(d); err != nil {
		return err
	}
	return nil
}

// StopSessionResult is an immutable semantic boundary value. Its zero is invalid.
type StopSessionResult struct {
	data        StopSessionResultData
	initialized bool
}

// StopSessionResultData is a construction/copy value; NewStopSessionResult validates and owns a copy.
type StopSessionResultData struct {
	Stop    SessionStopResult
	Outcome OutcomeEnvelope
}

func NewStopSessionResult(d StopSessionResultData) (StopSessionResult, error) {
	if err := d.validate(); err != nil {
		return StopSessionResult{}, err
	}
	return StopSessionResult{data: d.clone(), initialized: true}, nil
}
func (v StopSessionResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v StopSessionResult) Data() StopSessionResultData { return v.data.clone() }
func (v StopSessionResult) Clone() StopSessionResult    { return v }
func (d StopSessionResultData) clone() StopSessionResultData {

	return d
}
func (d StopSessionResultData) validate() error {
	if !d.Stop.Valid() {
		return invalid("d.Stop")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateStopSessionResultEvidence(d); err != nil {
		return err
	}
	if err := consistentStopSessionResult(d); err != nil {
		return err
	}
	return nil
}

// RestartSessionResult is an immutable semantic boundary value. Its zero is invalid.
type RestartSessionResult struct {
	data        RestartSessionResultData
	initialized bool
}

// RestartSessionResultData is a construction/copy value; NewRestartSessionResult validates and owns a copy.
type RestartSessionResultData struct {
	Restart SessionRestartResult
	Outcome OutcomeEnvelope
}

func NewRestartSessionResult(d RestartSessionResultData) (RestartSessionResult, error) {
	if err := d.validate(); err != nil {
		return RestartSessionResult{}, err
	}
	return RestartSessionResult{data: d.clone(), initialized: true}, nil
}
func (v RestartSessionResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v RestartSessionResult) Data() RestartSessionResultData { return v.data.clone() }
func (v RestartSessionResult) Clone() RestartSessionResult    { return v }
func (d RestartSessionResultData) clone() RestartSessionResultData {

	return d
}
func (d RestartSessionResultData) validate() error {
	if !d.Restart.Valid() {
		return invalid("d.Restart")
	}
	if !d.Outcome.Valid() {
		return invalid("d.Outcome")
	}
	if err := validateRestartSessionResultEvidence(d); err != nil {
		return err
	}
	if err := consistentRestartSessionResult(d); err != nil {
		return err
	}
	return nil
}

// ExactPRBase is an immutable semantic boundary value. Its zero is invalid.
type ExactPRBase struct {
	data        ExactPRBaseData
	initialized bool
}

// ExactPRBaseData is a construction/copy value; NewExactPRBase validates and owns a copy.
type ExactPRBaseData struct {
	Revision domain.Revision
}

func NewExactPRBase(d ExactPRBaseData) (ExactPRBase, error) {
	if err := d.validate(); err != nil {
		return ExactPRBase{}, err
	}
	return ExactPRBase{data: d.clone(), initialized: true}, nil
}
func (v ExactPRBase) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v ExactPRBase) Data() ExactPRBaseData { return v.data.clone() }
func (v ExactPRBase) Clone() ExactPRBase    { return v }
func (d ExactPRBaseData) clone() ExactPRBaseData {

	return d
}
func (d ExactPRBaseData) validate() error {
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	if d.Revision.Repository().Scope() != domain.Remote {
		return invalid("PR base revision")
	}
	return nil
}

// ObserveCurrentPRBase is an immutable semantic boundary value. Its zero is invalid.
type ObserveCurrentPRBase struct {
	data        ObserveCurrentPRBaseData
	initialized bool
}

// ObserveCurrentPRBaseData is a construction/copy value; NewObserveCurrentPRBase validates and owns a copy.
type ObserveCurrentPRBaseData struct {
}

func NewObserveCurrentPRBase(d ObserveCurrentPRBaseData) (ObserveCurrentPRBase, error) {
	if err := d.validate(); err != nil {
		return ObserveCurrentPRBase{}, err
	}
	return ObserveCurrentPRBase{data: d.clone(), initialized: true}, nil
}
func (v ObserveCurrentPRBase) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ObserveCurrentPRBase) Data() ObserveCurrentPRBaseData { return v.data.clone() }
func (v ObserveCurrentPRBase) Clone() ObserveCurrentPRBase    { return v }
func (d ObserveCurrentPRBaseData) clone() ObserveCurrentPRBaseData {

	return d
}
func (d ObserveCurrentPRBaseData) validate() error {

	return nil
}

// NavigatorQuery is an immutable semantic boundary value. Its zero is invalid.
type NavigatorQuery struct {
	data        NavigatorQueryData
	initialized bool
}

// NavigatorQueryData is a construction/copy value; NewNavigatorQuery validates and owns a copy.
type NavigatorQueryData struct {
	Repository domain.RepositoryID
	Namespace  []string
	Local      Optional[domain.RepositoryID]
}

func NewNavigatorQuery(d NavigatorQueryData) (NavigatorQuery, error) {
	if err := d.validate(); err != nil {
		return NavigatorQuery{}, err
	}
	return NavigatorQuery{data: d.clone(), initialized: true}, nil
}
func (v NavigatorQuery) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v NavigatorQuery) Data() NavigatorQueryData { return v.data.clone() }
func (v NavigatorQuery) Clone() NavigatorQuery    { return v }
func (d NavigatorQueryData) clone() NavigatorQueryData {
	d.Namespace = cloneSlice(d.Namespace)
	return d
}
func (d NavigatorQueryData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	if item, ok := d.Local.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Repository.Scope() != domain.Remote {
		return invalid("navigator remote")
	}
	if l, ok := d.Local.Value(); ok && l.Scope() != domain.LocalCommon {
		return invalid("navigator local")
	}
	return nil
}

// BranchContextQuery is an immutable semantic boundary value. Its zero is invalid.
type BranchContextQuery struct {
	data        BranchContextQueryData
	initialized bool
}

// BranchContextQueryData is a construction/copy value; NewBranchContextQuery validates and owns a copy.
type BranchContextQueryData struct {
	Target domain.ExactTarget
	Page   PageRequest
}

func NewBranchContextQuery(d BranchContextQueryData) (BranchContextQuery, error) {
	if err := d.validate(); err != nil {
		return BranchContextQuery{}, err
	}
	return BranchContextQuery{data: d.clone(), initialized: true}, nil
}
func (v BranchContextQuery) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v BranchContextQuery) Data() BranchContextQueryData { return v.data.clone() }
func (v BranchContextQuery) Clone() BranchContextQuery    { return v }
func (d BranchContextQueryData) clone() BranchContextQueryData {

	return d
}
func (d BranchContextQueryData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}

	return nil
}

// CommitsQuery is an immutable semantic boundary value. Its zero is invalid.
type CommitsQuery struct {
	data        CommitsQueryData
	initialized bool
}

// CommitsQueryData is a construction/copy value; NewCommitsQuery validates and owns a copy.
type CommitsQueryData struct {
	Target domain.ExactTarget
	Page   PageRequest
}

func NewCommitsQuery(d CommitsQueryData) (CommitsQuery, error) {
	if err := d.validate(); err != nil {
		return CommitsQuery{}, err
	}
	return CommitsQuery{data: d.clone(), initialized: true}, nil
}
func (v CommitsQuery) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v CommitsQuery) Data() CommitsQueryData { return v.data.clone() }
func (v CommitsQuery) Clone() CommitsQuery    { return v }
func (d CommitsQueryData) clone() CommitsQueryData {

	return d
}
func (d CommitsQueryData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}

	return nil
}

// GraphQuery is an immutable semantic boundary value. Its zero is invalid.
type GraphQuery struct {
	data        GraphQueryData
	initialized bool
}

// GraphQueryData is a construction/copy value; NewGraphQuery validates and owns a copy.
type GraphQueryData struct {
	Repository domain.RepositoryID
	Roots      []domain.Revision
	Filter     GraphFilter
	Page       PageRequest
}

func NewGraphQuery(d GraphQueryData) (GraphQuery, error) {
	if err := d.validate(); err != nil {
		return GraphQuery{}, err
	}
	return GraphQuery{data: d.clone(), initialized: true}, nil
}
func (v GraphQuery) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v GraphQuery) Data() GraphQueryData { return v.data.clone() }
func (v GraphQuery) Clone() GraphQuery    { return v }
func (d GraphQueryData) clone() GraphQueryData {
	d.Roots = cloneSlice(d.Roots)
	return d
}
func (d GraphQueryData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	for _, item := range d.Roots {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validGraphFilter(d.Filter) {
		return invalid("d.Filter")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.LocalCommon || len(d.Roots) == 0 {
		return invalid("graph query")
	}
	for _, r := range d.Roots {
		if r.Repository() != d.Repository {
			return invalid("graph roots")
		}
	}
	return nil
}

// DiffQuery is an immutable semantic boundary value. Its zero is invalid.
type DiffQuery struct {
	data        DiffQueryData
	initialized bool
}

// DiffQueryData is a construction/copy value; NewDiffQuery validates and owns a copy.
type DiffQueryData struct {
	Comparison GitComparison
	Paths      []GitPath
	Limits     PatchLimits
}

func NewDiffQuery(d DiffQueryData) (DiffQuery, error) {
	if err := d.validate(); err != nil {
		return DiffQuery{}, err
	}
	return DiffQuery{data: d.clone(), initialized: true}, nil
}
func (v DiffQuery) Valid() bool         { return v.initialized && v.data.validate() == nil }
func (v DiffQuery) Data() DiffQueryData { return v.data.clone() }
func (v DiffQuery) Clone() DiffQuery    { return v }
func (d DiffQueryData) clone() DiffQueryData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d DiffQueryData) validate() error {
	if !validGitComparison(d.Comparison) {
		return invalid("d.Comparison")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Limits.Valid() {
		return invalid("d.Limits")
	}

	return nil
}

// PullRequestDiffQuery is an immutable semantic boundary value. Its zero is invalid.
type PullRequestDiffQuery struct {
	data        PullRequestDiffQueryData
	initialized bool
}

// PullRequestDiffQueryData is a construction/copy value; NewPullRequestDiffQuery validates and owns a copy.
type PullRequestDiffQueryData struct {
	Target domain.ExactTarget
	Base   PRBaseSelection
	Local  domain.RepositoryID
	Paths  []GitPath
	Limits PatchLimits
}

func NewPullRequestDiffQuery(d PullRequestDiffQueryData) (PullRequestDiffQuery, error) {
	if err := d.validate(); err != nil {
		return PullRequestDiffQuery{}, err
	}
	return PullRequestDiffQuery{data: d.clone(), initialized: true}, nil
}
func (v PullRequestDiffQuery) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v PullRequestDiffQuery) Data() PullRequestDiffQueryData { return v.data.clone() }
func (v PullRequestDiffQuery) Clone() PullRequestDiffQuery    { return v }
func (d PullRequestDiffQueryData) clone() PullRequestDiffQueryData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d PullRequestDiffQueryData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !validPRBaseSelection(d.Base) {
		return invalid("d.Base")
	}
	if !d.Local.Valid() {
		return invalid("d.Local")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Limits.Valid() {
		return invalid("d.Limits")
	}
	if d.Target.Kind() != domain.PullRequestTarget || d.Local.Scope() != domain.LocalCommon {
		return invalid("PR diff target")
	}
	return nil
}

// WorktreeStatusQuery is an immutable semantic boundary value. Its zero is invalid.
type WorktreeStatusQuery struct {
	data        WorktreeStatusQueryData
	initialized bool
}

// WorktreeStatusQueryData is a construction/copy value; NewWorktreeStatusQuery validates and owns a copy.
type WorktreeStatusQueryData struct {
	WorktreeID domain.WorktreeID
}

func NewWorktreeStatusQuery(d WorktreeStatusQueryData) (WorktreeStatusQuery, error) {
	if err := d.validate(); err != nil {
		return WorktreeStatusQuery{}, err
	}
	return WorktreeStatusQuery{data: d.clone(), initialized: true}, nil
}
func (v WorktreeStatusQuery) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v WorktreeStatusQuery) Data() WorktreeStatusQueryData { return v.data.clone() }
func (v WorktreeStatusQuery) Clone() WorktreeStatusQuery    { return v }
func (d WorktreeStatusQueryData) clone() WorktreeStatusQueryData {

	return d
}
func (d WorktreeStatusQueryData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}

	return nil
}

// StashesQuery is an immutable semantic boundary value. Its zero is invalid.
type StashesQuery struct {
	data        StashesQueryData
	initialized bool
}

// StashesQueryData is a construction/copy value; NewStashesQuery validates and owns a copy.
type StashesQueryData struct {
	Repository domain.RepositoryID
	Page       PageRequest
}

func NewStashesQuery(d StashesQueryData) (StashesQuery, error) {
	if err := d.validate(); err != nil {
		return StashesQuery{}, err
	}
	return StashesQuery{data: d.clone(), initialized: true}, nil
}
func (v StashesQuery) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v StashesQuery) Data() StashesQueryData { return v.data.clone() }
func (v StashesQuery) Clone() StashesQuery    { return v }
func (d StashesQueryData) clone() StashesQueryData {

	return d
}
func (d StashesQueryData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("stashes scope")
	}
	return nil
}

// StashPatchQuery is an immutable semantic boundary value. Its zero is invalid.
type StashPatchQuery struct {
	data        StashPatchQueryData
	initialized bool
}

// StashPatchQueryData is a construction/copy value; NewStashPatchQuery validates and owns a copy.
type StashPatchQueryData struct {
	Stash  domain.StashID
	View   StashPatchView
	Paths  []GitPath
	Limits PatchLimits
}

func NewStashPatchQuery(d StashPatchQueryData) (StashPatchQuery, error) {
	if err := d.validate(); err != nil {
		return StashPatchQuery{}, err
	}
	return StashPatchQuery{data: d.clone(), initialized: true}, nil
}
func (v StashPatchQuery) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v StashPatchQuery) Data() StashPatchQueryData { return v.data.clone() }
func (v StashPatchQuery) Clone() StashPatchQuery    { return v }
func (d StashPatchQueryData) clone() StashPatchQueryData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d StashPatchQueryData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !validStashPatchView(d.View) {
		return invalid("d.View")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Limits.Valid() {
		return invalid("d.Limits")
	}

	return nil
}

// LaunchPointsQuery is an immutable semantic boundary value. Its zero is invalid.
type LaunchPointsQuery struct {
	data        LaunchPointsQueryData
	initialized bool
}

// LaunchPointsQueryData is a construction/copy value; NewLaunchPointsQuery validates and owns a copy.
type LaunchPointsQueryData struct {
	WorktreeID domain.WorktreeID
}

func NewLaunchPointsQuery(d LaunchPointsQueryData) (LaunchPointsQuery, error) {
	if err := d.validate(); err != nil {
		return LaunchPointsQuery{}, err
	}
	return LaunchPointsQuery{data: d.clone(), initialized: true}, nil
}
func (v LaunchPointsQuery) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v LaunchPointsQuery) Data() LaunchPointsQueryData { return v.data.clone() }
func (v LaunchPointsQuery) Clone() LaunchPointsQuery    { return v }
func (d LaunchPointsQueryData) clone() LaunchPointsQueryData {

	return d
}
func (d LaunchPointsQueryData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}

	return nil
}

// SessionsQuery is an immutable semantic boundary value. Its zero is invalid.
type SessionsQuery struct {
	data        SessionsQueryData
	initialized bool
}

// SessionsQueryData is a construction/copy value; NewSessionsQuery validates and owns a copy.
type SessionsQueryData struct {
	WorktreeID Optional[domain.WorktreeID]
}

func NewSessionsQuery(d SessionsQueryData) (SessionsQuery, error) {
	if err := d.validate(); err != nil {
		return SessionsQuery{}, err
	}
	return SessionsQuery{data: d.clone(), initialized: true}, nil
}
func (v SessionsQuery) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v SessionsQuery) Data() SessionsQueryData { return v.data.clone() }
func (v SessionsQuery) Clone() SessionsQuery    { return v }
func (d SessionsQueryData) clone() SessionsQueryData {

	return d
}
func (d SessionsQueryData) validate() error {
	if item, ok := d.WorktreeID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// SessionOutputQuery is an immutable semantic boundary value. Its zero is invalid.
type SessionOutputQuery struct {
	data        SessionOutputQueryData
	initialized bool
}

// SessionOutputQueryData is a construction/copy value; NewSessionOutputQuery validates and owns a copy.
type SessionOutputQueryData struct {
	SessionID domain.SessionID
	Offset    uint64
	MaxBytes  uint32
}

func NewSessionOutputQuery(d SessionOutputQueryData) (SessionOutputQuery, error) {
	if err := d.validate(); err != nil {
		return SessionOutputQuery{}, err
	}
	return SessionOutputQuery{data: d.clone(), initialized: true}, nil
}
func (v SessionOutputQuery) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v SessionOutputQuery) Data() SessionOutputQueryData { return v.data.clone() }
func (v SessionOutputQuery) Clone() SessionOutputQuery    { return v }
func (d SessionOutputQueryData) clone() SessionOutputQueryData {

	return d
}
func (d SessionOutputQueryData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	if d.MaxBytes == 0 || d.MaxBytes > 262144 {
		return invalid("output bound")
	}
	return nil
}

// PreferencesQuery is an immutable semantic boundary value. Its zero is invalid.
type PreferencesQuery struct {
	data        PreferencesQueryData
	initialized bool
}

// PreferencesQueryData is a construction/copy value; NewPreferencesQuery validates and owns a copy.
type PreferencesQueryData struct {
	Repository domain.RepositoryID
}

func NewPreferencesQuery(d PreferencesQueryData) (PreferencesQuery, error) {
	if err := d.validate(); err != nil {
		return PreferencesQuery{}, err
	}
	return PreferencesQuery{data: d.clone(), initialized: true}, nil
}
func (v PreferencesQuery) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v PreferencesQuery) Data() PreferencesQueryData { return v.data.clone() }
func (v PreferencesQuery) Clone() PreferencesQuery    { return v }
func (d PreferencesQueryData) clone() PreferencesQueryData {

	return d
}
func (d PreferencesQueryData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	return nil
}

// GraphEdge is an immutable semantic boundary value. Its zero is invalid.
type GraphEdge struct {
	data        GraphEdgeData
	initialized bool
}

// GraphEdgeData is a construction/copy value; NewGraphEdge validates and owns a copy.
type GraphEdgeData struct {
	Child  domain.Revision
	Parent domain.Revision
}

func NewGraphEdge(d GraphEdgeData) (GraphEdge, error) {
	if err := d.validate(); err != nil {
		return GraphEdge{}, err
	}
	return GraphEdge{data: d.clone(), initialized: true}, nil
}
func (v GraphEdge) Valid() bool         { return v.initialized && v.data.validate() == nil }
func (v GraphEdge) Data() GraphEdgeData { return v.data.clone() }
func (v GraphEdge) Clone() GraphEdge    { return v }
func (d GraphEdgeData) clone() GraphEdgeData {

	return d
}
func (d GraphEdgeData) validate() error {
	if !d.Child.Valid() {
		return invalid("d.Child")
	}
	if !d.Parent.Valid() {
		return invalid("d.Parent")
	}
	if !sameLocal(d.Child, d.Parent) {
		return invalid("graph edge")
	}
	return nil
}

// GraphAnnotation is an immutable semantic boundary value. Its zero is invalid.
type GraphAnnotation struct {
	data        GraphAnnotationData
	initialized bool
}

// GraphAnnotationData is a construction/copy value; NewGraphAnnotation validates and owns a copy.
type GraphAnnotationData struct {
	Revision     domain.Revision
	Branches     []BranchRelationship
	PullRequests []PRRelation
	Worktrees    []WorktreeRelation
}

func NewGraphAnnotation(d GraphAnnotationData) (GraphAnnotation, error) {
	if err := d.validate(); err != nil {
		return GraphAnnotation{}, err
	}
	return GraphAnnotation{data: d.clone(), initialized: true}, nil
}
func (v GraphAnnotation) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v GraphAnnotation) Data() GraphAnnotationData { return v.data.clone() }
func (v GraphAnnotation) Clone() GraphAnnotation    { return v }
func (d GraphAnnotationData) clone() GraphAnnotationData {
	d.Branches = cloneSlice(d.Branches)
	d.PullRequests = cloneSlice(d.PullRequests)
	d.Worktrees = cloneSlice(d.Worktrees)
	return d
}
func (d GraphAnnotationData) validate() error {
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	for _, item := range d.Branches {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.PullRequests {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// EffectivePreferences is an immutable semantic boundary value. Its zero is invalid.
type EffectivePreferences struct {
	data        EffectivePreferencesData
	initialized bool
}

// EffectivePreferencesData is a construction/copy value; NewEffectivePreferences validates and owns a copy.
type EffectivePreferencesData struct {
	Repository           domain.RepositoryID
	StripPrefixes        []string
	Folder               Optional[string]
	ConfiguredTargets    []ConfiguredWorktreeTarget
	UserDocument         Optional[UserConfigDocument]
	PreferencesDocument  Optional[PreferencesDocument]
	MigrationDiagnostics []Diagnostic
}

func NewEffectivePreferences(d EffectivePreferencesData) (EffectivePreferences, error) {
	if err := d.validate(); err != nil {
		return EffectivePreferences{}, err
	}
	return EffectivePreferences{data: d.clone(), initialized: true}, nil
}
func (v EffectivePreferences) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v EffectivePreferences) Data() EffectivePreferencesData { return v.data.clone() }
func (v EffectivePreferences) Clone() EffectivePreferences    { return v }
func (d EffectivePreferencesData) clone() EffectivePreferencesData {
	d.StripPrefixes = cloneSlice(d.StripPrefixes)
	d.ConfiguredTargets = cloneSlice(d.ConfiguredTargets)
	d.MigrationDiagnostics = cloneSlice(d.MigrationDiagnostics)
	return d
}
func (d EffectivePreferencesData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	for _, item := range d.ConfiguredTargets {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.UserDocument.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.PreferencesDocument.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.MigrationDiagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// NavigatorResult is an immutable semantic boundary value. Its zero is invalid.
type NavigatorResult struct {
	data        NavigatorResultData
	initialized bool
}

// NavigatorResultData is a construction/copy value; NewNavigatorResult validates and owns a copy.
type NavigatorResultData struct {
	Repository     domain.RepositoryID
	Namespace      []string
	PullRequests   []PullRequestFact
	Branches       []BranchRelationship
	Worktrees      []WorktreeFacts
	Capabilities   []RemoteCapabilities
	Active         ActiveContext
	Activations    []ActivationCandidate
	Page           Optional[PageInfo]
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewNavigatorResult(d NavigatorResultData) (NavigatorResult, error) {
	if err := d.validate(); err != nil {
		return NavigatorResult{}, err
	}
	return NavigatorResult{data: d.clone(), initialized: true}, nil
}
func (v NavigatorResult) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v NavigatorResult) Data() NavigatorResultData { return v.data.clone() }
func (v NavigatorResult) Clone() NavigatorResult    { return v }
func (d NavigatorResultData) clone() NavigatorResultData {
	d.Namespace = cloneSlice(d.Namespace)
	d.PullRequests = cloneSlice(d.PullRequests)
	d.Branches = cloneSlice(d.Branches)
	d.Worktrees = cloneSlice(d.Worktrees)
	d.Capabilities = cloneSlice(d.Capabilities)
	d.Activations = cloneSlice(d.Activations)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d NavigatorResultData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	for _, item := range d.PullRequests {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Branches {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Capabilities {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Active.Valid() {
		return invalid("d.Active")
	}
	for _, item := range d.Activations {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Page.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentNavigatorResult(d); err != nil {
		return err
	}
	return nil
}

// BranchContextResult is an immutable semantic boundary value. Its zero is invalid.
type BranchContextResult struct {
	data        BranchContextResultData
	initialized bool
}

// BranchContextResultData is a construction/copy value; NewBranchContextResult validates and owns a copy.
type BranchContextResultData struct {
	Target         domain.ExactTarget
	Endpoint       Optional[ExactLocalResolution]
	Relationship   Optional[BranchRelationship]
	Commits        []CommitFact
	Page           PageInfo
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewBranchContextResult(d BranchContextResultData) (BranchContextResult, error) {
	if err := d.validate(); err != nil {
		return BranchContextResult{}, err
	}
	return BranchContextResult{data: d.clone(), initialized: true}, nil
}
func (v BranchContextResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v BranchContextResult) Data() BranchContextResultData { return v.data.clone() }
func (v BranchContextResult) Clone() BranchContextResult    { return v }
func (d BranchContextResultData) clone() BranchContextResultData {
	d.Commits = cloneSlice(d.Commits)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d BranchContextResultData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if item, ok := d.Endpoint.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Relationship.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Commits {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentBranchContextResult(d); err != nil {
		return err
	}
	return nil
}

// CommitsResult is an immutable semantic boundary value. Its zero is invalid.
type CommitsResult struct {
	data        CommitsResultData
	initialized bool
}

// CommitsResultData is a construction/copy value; NewCommitsResult validates and owns a copy.
type CommitsResultData struct {
	Target         domain.ExactTarget
	Endpoint       Optional[ExactLocalResolution]
	Commits        []CommitFact
	Page           PageInfo
	Relationships  []BranchRelationship
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewCommitsResult(d CommitsResultData) (CommitsResult, error) {
	if err := d.validate(); err != nil {
		return CommitsResult{}, err
	}
	return CommitsResult{data: d.clone(), initialized: true}, nil
}
func (v CommitsResult) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v CommitsResult) Data() CommitsResultData { return v.data.clone() }
func (v CommitsResult) Clone() CommitsResult    { return v }
func (d CommitsResultData) clone() CommitsResultData {
	d.Commits = cloneSlice(d.Commits)
	d.Relationships = cloneSlice(d.Relationships)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d CommitsResultData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if item, ok := d.Endpoint.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Commits {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	for _, item := range d.Relationships {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentCommitsResult(d); err != nil {
		return err
	}
	return nil
}

// GraphResult is an immutable semantic boundary value. Its zero is invalid.
type GraphResult struct {
	data        GraphResultData
	initialized bool
}

// GraphResultData is a construction/copy value; NewGraphResult validates and owns a copy.
type GraphResultData struct {
	Repository     domain.RepositoryID
	Roots          []domain.Revision
	Commits        []CommitFact
	Edges          []GraphEdge
	Refs           []RefFact
	Annotations    []GraphAnnotation
	Page           PageInfo
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewGraphResult(d GraphResultData) (GraphResult, error) {
	if err := d.validate(); err != nil {
		return GraphResult{}, err
	}
	return GraphResult{data: d.clone(), initialized: true}, nil
}
func (v GraphResult) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v GraphResult) Data() GraphResultData { return v.data.clone() }
func (v GraphResult) Clone() GraphResult    { return v }
func (d GraphResultData) clone() GraphResultData {
	d.Roots = cloneSlice(d.Roots)
	d.Commits = cloneSlice(d.Commits)
	d.Edges = cloneSlice(d.Edges)
	d.Refs = cloneSlice(d.Refs)
	d.Annotations = cloneSlice(d.Annotations)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d GraphResultData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	for _, item := range d.Roots {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Commits {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Edges {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Annotations {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentGraphResult(d); err != nil {
		return err
	}
	return nil
}

// DiffResult is an immutable semantic boundary value. Its zero is invalid.
type DiffResult struct {
	data        DiffResultData
	initialized bool
}

// DiffResultData is a construction/copy value; NewDiffResult validates and owns a copy.
type DiffResultData struct {
	Comparison     GitComparison
	Patch          Optional[PatchFacts]
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewDiffResult(d DiffResultData) (DiffResult, error) {
	if err := d.validate(); err != nil {
		return DiffResult{}, err
	}
	return DiffResult{data: d.clone(), initialized: true}, nil
}
func (v DiffResult) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v DiffResult) Data() DiffResultData { return v.data.clone() }
func (v DiffResult) Clone() DiffResult    { return v }
func (d DiffResultData) clone() DiffResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d DiffResultData) validate() error {
	if !validGitComparison(d.Comparison) {
		return invalid("d.Comparison")
	}
	if item, ok := d.Patch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentDiffResult(d); err != nil {
		return err
	}
	return nil
}

// PullRequestDiffResult is an immutable semantic boundary value. Its zero is invalid.
type PullRequestDiffResult struct {
	data        PullRequestDiffResultData
	initialized bool
}

// PullRequestDiffResultData is a construction/copy value; NewPullRequestDiffResult validates and owns a copy.
type PullRequestDiffResultData struct {
	Target         domain.ExactTarget
	RequestedBase  PRBaseSelection
	RemotePR       Optional[PullRequestFact]
	Base           Optional[ExactLocalResolution]
	Head           Optional[ExactLocalResolution]
	MergeBase      Optional[domain.Revision]
	Patch          Optional[PatchFacts]
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewPullRequestDiffResult(d PullRequestDiffResultData) (PullRequestDiffResult, error) {
	if err := d.validate(); err != nil {
		return PullRequestDiffResult{}, err
	}
	return PullRequestDiffResult{data: d.clone(), initialized: true}, nil
}
func (v PullRequestDiffResult) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v PullRequestDiffResult) Data() PullRequestDiffResultData { return v.data.clone() }
func (v PullRequestDiffResult) Clone() PullRequestDiffResult    { return v }
func (d PullRequestDiffResultData) clone() PullRequestDiffResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d PullRequestDiffResultData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !validPRBaseSelection(d.RequestedBase) {
		return invalid("d.RequestedBase")
	}
	if item, ok := d.RemotePR.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Base.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.MergeBase.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Patch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentPullRequestDiffResult(d); err != nil {
		return err
	}
	return nil
}

// WorktreeStatusResult is an immutable semantic boundary value. Its zero is invalid.
type WorktreeStatusResult struct {
	data        WorktreeStatusResultData
	initialized bool
}

// WorktreeStatusResultData is a construction/copy value; NewWorktreeStatusResult validates and owns a copy.
type WorktreeStatusResultData struct {
	WorktreeID     domain.WorktreeID
	Status         Optional[StatusFacts]
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewWorktreeStatusResult(d WorktreeStatusResultData) (WorktreeStatusResult, error) {
	if err := d.validate(); err != nil {
		return WorktreeStatusResult{}, err
	}
	return WorktreeStatusResult{data: d.clone(), initialized: true}, nil
}
func (v WorktreeStatusResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v WorktreeStatusResult) Data() WorktreeStatusResultData { return v.data.clone() }
func (v WorktreeStatusResult) Clone() WorktreeStatusResult    { return v }
func (d WorktreeStatusResultData) clone() WorktreeStatusResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d WorktreeStatusResultData) validate() error {
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentWorktreeStatusResult(d); err != nil {
		return err
	}
	return nil
}

// StashesResult is an immutable semantic boundary value. Its zero is invalid.
type StashesResult struct {
	data        StashesResultData
	initialized bool
}

// StashesResultData is a construction/copy value; NewStashesResult validates and owns a copy.
type StashesResultData struct {
	Repository     domain.RepositoryID
	Stashes        []StashFact
	Page           PageInfo
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewStashesResult(d StashesResultData) (StashesResult, error) {
	if err := d.validate(); err != nil {
		return StashesResult{}, err
	}
	return StashesResult{data: d.clone(), initialized: true}, nil
}
func (v StashesResult) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v StashesResult) Data() StashesResultData { return v.data.clone() }
func (v StashesResult) Clone() StashesResult    { return v }
func (d StashesResultData) clone() StashesResultData {
	d.Stashes = cloneSlice(d.Stashes)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d StashesResultData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	for _, item := range d.Stashes {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentStashesResult(d); err != nil {
		return err
	}
	return nil
}

// StashPatchResult is an immutable semantic boundary value. Its zero is invalid.
type StashPatchResult struct {
	data        StashPatchResultData
	initialized bool
}

// StashPatchResultData is a construction/copy value; NewStashPatchResult validates and owns a copy.
type StashPatchResultData struct {
	Stash          domain.StashID
	Parents        []domain.OID
	Comparison     Optional[StashPatchComparison]
	Patch          Optional[PatchFacts]
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewStashPatchResult(d StashPatchResultData) (StashPatchResult, error) {
	if err := d.validate(); err != nil {
		return StashPatchResult{}, err
	}
	return StashPatchResult{data: d.clone(), initialized: true}, nil
}
func (v StashPatchResult) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v StashPatchResult) Data() StashPatchResultData { return v.data.clone() }
func (v StashPatchResult) Clone() StashPatchResult    { return v }
func (d StashPatchResultData) clone() StashPatchResultData {
	d.Parents = cloneSlice(d.Parents)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d StashPatchResultData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	for _, item := range d.Parents {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Comparison.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Patch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentStashPatchResult(d); err != nil {
		return err
	}
	return nil
}

// LaunchPointsResult is an immutable semantic boundary value. Its zero is invalid.
type LaunchPointsResult struct {
	data        LaunchPointsResultData
	initialized bool
}

// LaunchPointsResultData is a construction/copy value; NewLaunchPointsResult validates and owns a copy.
type LaunchPointsResultData struct {
	WorktreeID     domain.WorktreeID
	Definitions    []LaunchDefinition
	Saved          []SavedLaunchObservation
	DefaultAlias   StoredField[string]
	DefaultID      Optional[domain.LaunchPointID]
	StorageVersion Optional[StorageVersion]
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewLaunchPointsResult(d LaunchPointsResultData) (LaunchPointsResult, error) {
	if err := d.validate(); err != nil {
		return LaunchPointsResult{}, err
	}
	return LaunchPointsResult{data: d.clone(), initialized: true}, nil
}
func (v LaunchPointsResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v LaunchPointsResult) Data() LaunchPointsResultData { return v.data.clone() }
func (v LaunchPointsResult) Clone() LaunchPointsResult    { return v }
func (d LaunchPointsResultData) clone() LaunchPointsResultData {
	d.Definitions = cloneSlice(d.Definitions)
	d.Saved = cloneSlice(d.Saved)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d LaunchPointsResultData) validate() error {
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
	if !d.DefaultAlias.Valid() {
		return invalid("field presence")
	}
	if item, ok := d.DefaultID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.StorageVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentLaunchPointsResult(d); err != nil {
		return err
	}
	return nil
}

// SessionsResult is an immutable semantic boundary value. Its zero is invalid.
type SessionsResult struct {
	data        SessionsResultData
	initialized bool
}

// SessionsResultData is a construction/copy value; NewSessionsResult validates and owns a copy.
type SessionsResultData struct {
	Sessions       SessionList
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewSessionsResult(d SessionsResultData) (SessionsResult, error) {
	if err := d.validate(); err != nil {
		return SessionsResult{}, err
	}
	return SessionsResult{data: d.clone(), initialized: true}, nil
}
func (v SessionsResult) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v SessionsResult) Data() SessionsResultData { return v.data.clone() }
func (v SessionsResult) Clone() SessionsResult    { return v }
func (d SessionsResultData) clone() SessionsResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionsResultData) validate() error {
	if !d.Sessions.Valid() {
		return invalid("d.Sessions")
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentSessionsResult(d); err != nil {
		return err
	}
	return nil
}

// SessionOutputProjection is an immutable semantic boundary value. Its zero is invalid.
type SessionOutputProjection struct {
	data        SessionOutputProjectionData
	initialized bool
}

// SessionOutputProjectionData is a construction/copy value; NewSessionOutputProjection validates and owns a copy.
type SessionOutputProjectionData struct {
	Output         SessionOutputResult
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewSessionOutputProjection(d SessionOutputProjectionData) (SessionOutputProjection, error) {
	if err := d.validate(); err != nil {
		return SessionOutputProjection{}, err
	}
	return SessionOutputProjection{data: d.clone(), initialized: true}, nil
}
func (v SessionOutputProjection) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v SessionOutputProjection) Data() SessionOutputProjectionData { return v.data.clone() }
func (v SessionOutputProjection) Clone() SessionOutputProjection    { return v }
func (d SessionOutputProjectionData) clone() SessionOutputProjectionData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionOutputProjectionData) validate() error {
	if !d.Output.Valid() {
		return invalid("d.Output")
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentSessionOutputProjection(d); err != nil {
		return err
	}
	return nil
}

// PreferencesResult is an immutable semantic boundary value. Its zero is invalid.
type PreferencesResult struct {
	data        PreferencesResultData
	initialized bool
}

// PreferencesResultData is a construction/copy value; NewPreferencesResult validates and owns a copy.
type PreferencesResultData struct {
	Effective      EffectivePreferences
	Active         ActiveContext
	Sources        ProjectionSources
	ContextVersion ContextVersion
	Diagnostics    []Diagnostic
}

func NewPreferencesResult(d PreferencesResultData) (PreferencesResult, error) {
	if err := d.validate(); err != nil {
		return PreferencesResult{}, err
	}
	return PreferencesResult{data: d.clone(), initialized: true}, nil
}
func (v PreferencesResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v PreferencesResult) Data() PreferencesResultData { return v.data.clone() }
func (v PreferencesResult) Clone() PreferencesResult    { return v }
func (d PreferencesResultData) clone() PreferencesResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d PreferencesResultData) validate() error {
	if !d.Effective.Valid() {
		return invalid("d.Effective")
	}
	if !d.Active.Valid() {
		return invalid("d.Active")
	}
	if !d.Sources.Valid() {
		return invalid("d.Sources")
	}
	if !d.ContextVersion.Valid() {
		return invalid("d.ContextVersion")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := consistentPreferencesResult(d); err != nil {
		return err
	}
	return nil
}

// Receipt is an immutable semantic boundary value. Its zero is invalid.
type Receipt struct {
	data        ReceiptData
	initialized bool
}

// ReceiptData is a construction/copy value; NewReceipt validates and owns a copy.
type ReceiptData struct {
	OperationID      OperationID
	AcceptedSequence Sequence
}

func NewReceipt(d ReceiptData) (Receipt, error) {
	if err := d.validate(); err != nil {
		return Receipt{}, err
	}
	return Receipt{data: d.clone(), initialized: true}, nil
}
func (v Receipt) Valid() bool       { return v.initialized && v.data.validate() == nil }
func (v Receipt) Data() ReceiptData { return v.data.clone() }
func (v Receipt) Clone() Receipt    { return v }
func (d ReceiptData) clone() ReceiptData {

	return d
}
func (d ReceiptData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.AcceptedSequence.Valid() {
		return invalid("d.AcceptedSequence")
	}

	return nil
}

// Accepted is an immutable semantic boundary value. Its zero is invalid.
type Accepted struct {
	data        AcceptedData
	initialized bool
}

// AcceptedData is a construction/copy value; NewAccepted validates and owns a copy.
type AcceptedData struct {
	OperationID OperationID
	Correlation Correlation
	Receipt     Receipt
}

func NewAccepted(d AcceptedData) (Accepted, error) {
	if err := d.validate(); err != nil {
		return Accepted{}, err
	}
	return Accepted{data: d.clone(), initialized: true}, nil
}
func (v Accepted) Valid() bool        { return v.initialized && v.data.validate() == nil }
func (v Accepted) Data() AcceptedData { return v.data.clone() }
func (v Accepted) Clone() Accepted    { return v }
func (d AcceptedData) clone() AcceptedData {

	return d
}
func (d AcceptedData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Correlation.Valid() {
		return invalid("d.Correlation")
	}
	if !d.Receipt.Valid() {
		return invalid("d.Receipt")
	}
	if d.OperationID != d.Receipt.data.OperationID {
		return invalid("acceptance identity")
	}
	return nil
}

// Progress is an immutable semantic boundary value. Its zero is invalid.
type Progress struct {
	data        ProgressData
	initialized bool
}

// ProgressData is a construction/copy value; NewProgress validates and owns a copy.
type ProgressData struct {
	OperationID OperationID
	Correlation Correlation
	Stage       string
	Message     string
	Completed   Optional[uint64]
	Total       Optional[uint64]
}

func NewProgress(d ProgressData) (Progress, error) {
	if err := d.validate(); err != nil {
		return Progress{}, err
	}
	return Progress{data: d.clone(), initialized: true}, nil
}
func (v Progress) Valid() bool        { return v.initialized && v.data.validate() == nil }
func (v Progress) Data() ProgressData { return v.data.clone() }
func (v Progress) Clone() Progress    { return v }
func (d ProgressData) clone() ProgressData {

	return d
}
func (d ProgressData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Correlation.Valid() {
		return invalid("d.Correlation")
	}

	if a, ok := d.Completed.Value(); ok {
		if b, p := d.Total.Value(); p && a > b {
			return invalid("progress bounds")
		}
	}
	return nil
}

// ConfirmationRequested is an immutable semantic boundary value. Its zero is invalid.
type ConfirmationRequested struct {
	data        ConfirmationRequestedData
	initialized bool
}

// ConfirmationRequestedData is a construction/copy value; NewConfirmationRequested validates and owns a copy.
type ConfirmationRequestedData struct {
	OperationID    OperationID
	Correlation    Correlation
	ConfirmationID ConfirmationID
	Summary        MutationPlanSummary
	Choices        []Choice
}

func NewConfirmationRequested(d ConfirmationRequestedData) (ConfirmationRequested, error) {
	if err := d.validate(); err != nil {
		return ConfirmationRequested{}, err
	}
	return ConfirmationRequested{data: d.clone(), initialized: true}, nil
}
func (v ConfirmationRequested) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v ConfirmationRequested) Data() ConfirmationRequestedData { return v.data.clone() }
func (v ConfirmationRequested) Clone() ConfirmationRequested    { return v }
func (d ConfirmationRequestedData) clone() ConfirmationRequestedData {
	d.Choices = cloneSlice(d.Choices)
	return d
}
func (d ConfirmationRequestedData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Correlation.Valid() {
		return invalid("d.Correlation")
	}
	if !d.ConfirmationID.Valid() {
		return invalid("d.ConfirmationID")
	}
	if !d.Summary.Valid() {
		return invalid("d.Summary")
	}
	for _, item := range d.Choices {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.OperationID != d.Summary.data.OperationID || len(d.Choices) == 0 {
		return invalid("confirmation binding")
	}
	for _, c := range d.Choices {
		if !containsChoice(d.Summary.data.Choices, c) {
			return invalid("confirmation choices")
		}
	}
	return nil
}

// OperationTerminal is an immutable semantic boundary value. Its zero is invalid.
type OperationTerminal struct {
	data        OperationTerminalData
	initialized bool
}

// OperationTerminalData is a construction/copy value; NewOperationTerminal validates and owns a copy.
type OperationTerminalData struct {
	OperationID           OperationID
	Correlation           Correlation
	Disposition           TerminalDisposition
	Result                Optional[Result]
	Diagnostics           []Diagnostic
	Effects               EffectReport
	Recovery              []NormalizedRecovery
	CancellationRequested bool
}

func NewOperationTerminal(d OperationTerminalData) (OperationTerminal, error) {
	if err := d.validate(); err != nil {
		return OperationTerminal{}, err
	}
	return OperationTerminal{data: d.clone(), initialized: true}, nil
}
func (v OperationTerminal) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v OperationTerminal) Data() OperationTerminalData { return v.data.clone() }
func (v OperationTerminal) Clone() OperationTerminal    { return v }
func (d OperationTerminalData) clone() OperationTerminalData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d OperationTerminalData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Correlation.Valid() {
		return invalid("d.Correlation")
	}
	if !d.Disposition.Valid() {
		return invalid("d.Disposition")
	}
	if item, ok := d.Result.Value(); ok {
		if !validResult(item) {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if err := validateOperationTerminalEvidence(d); err != nil {
		return err
	}
	if err := consistentOperationTerminal(d); err != nil {
		return err
	}
	if d.Disposition == Succeeded && !d.Result.Present() {
		return invalid("success result")
	}
	return validateNormalizedRecovery(d.Effects, d.Recovery)
}

// SessionChanged is an immutable semantic boundary value. Its zero is invalid.
type SessionChanged struct {
	data        SessionChangedData
	initialized bool
}

// SessionChangedData is a construction/copy value; NewSessionChanged validates and owns a copy.
type SessionChangedData struct {
	SessionID       domain.SessionID
	SessionSequence SessionSequence
	RuntimeSequence RuntimeEventSequence
	Snapshot        SessionSnapshot
	FinalCleanup    bool
}

func NewSessionChanged(d SessionChangedData) (SessionChanged, error) {
	if err := d.validate(); err != nil {
		return SessionChanged{}, err
	}
	return SessionChanged{data: d.clone(), initialized: true}, nil
}
func (v SessionChanged) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v SessionChanged) Data() SessionChangedData { return v.data.clone() }
func (v SessionChanged) Clone() SessionChanged    { return v }
func (d SessionChangedData) clone() SessionChangedData {

	return d
}
func (d SessionChangedData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.SessionSequence.Valid() {
		return invalid("d.SessionSequence")
	}
	if !d.RuntimeSequence.Valid() {
		return invalid("d.RuntimeSequence")
	}
	if !d.Snapshot.Valid() {
		return invalid("d.Snapshot")
	}

	if d.SessionID != d.Snapshot.data.SessionID || d.SessionSequence != d.Snapshot.data.Sequence || d.FinalCleanup != (d.Snapshot.data.Phase == Cleaned) {
		return invalid("normalized session binding")
	}
	return nil
}

// ProjectionInvalidated is an immutable semantic boundary value. Its zero is invalid.
type ProjectionInvalidated struct {
	data        ProjectionInvalidatedData
	initialized bool
}

// ProjectionInvalidatedData is a construction/copy value; NewProjectionInvalidated validates and owns a copy.
type ProjectionInvalidatedData struct {
	Kind           InvalidationKind
	Repository     Optional[domain.RepositoryID]
	WorktreeID     Optional[domain.WorktreeID]
	Source         Optional[SourceVersion]
	ContextVersion Optional[ContextVersion]
}

func NewProjectionInvalidated(d ProjectionInvalidatedData) (ProjectionInvalidated, error) {
	if err := d.validate(); err != nil {
		return ProjectionInvalidated{}, err
	}
	return ProjectionInvalidated{data: d.clone(), initialized: true}, nil
}
func (v ProjectionInvalidated) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v ProjectionInvalidated) Data() ProjectionInvalidatedData { return v.data.clone() }
func (v ProjectionInvalidated) Clone() ProjectionInvalidated    { return v }
func (d ProjectionInvalidatedData) clone() ProjectionInvalidatedData {

	return d
}
func (d ProjectionInvalidatedData) validate() error {
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if item, ok := d.Repository.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.WorktreeID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Source.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ContextVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// Event is an immutable semantic boundary value. Its zero is invalid.
type Event struct {
	data        EventData
	initialized bool
}

// EventData is a construction/copy value; NewEvent validates and owns a copy.
type EventData struct {
	Sequence Sequence
	Payload  EventPayload
}

func NewEvent(d EventData) (Event, error) {
	if err := d.validate(); err != nil {
		return Event{}, err
	}
	return Event{data: d.clone(), initialized: true}, nil
}
func (v Event) Valid() bool     { return v.initialized && v.data.validate() == nil }
func (v Event) Data() EventData { return v.data.clone() }
func (v Event) Clone() Event    { return v }
func (d EventData) clone() EventData {

	return d
}
func (d EventData) validate() error {
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}
	if !validEventPayload(d.Payload) {
		return invalid("d.Payload")
	}
	if a, ok := d.Payload.(Accepted); ok && a.data.Receipt.data.AcceptedSequence != d.Sequence {
		return invalid("accepted sequence")
	}
	return nil
}

// OperationResidual is an immutable semantic boundary value. Its zero is invalid.
type OperationResidual struct {
	data        OperationResidualData
	initialized bool
}

// OperationResidualData is a construction/copy value; NewOperationResidual validates and owns a copy.
type OperationResidualData struct {
	OperationID OperationID
	Correlation Correlation
	Effects     EffectReport
	Recovery    []NormalizedRecovery
	Diagnostics []Diagnostic
}

func NewOperationResidual(d OperationResidualData) (OperationResidual, error) {
	if err := d.validate(); err != nil {
		return OperationResidual{}, err
	}
	return OperationResidual{data: d.clone(), initialized: true}, nil
}
func (v OperationResidual) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v OperationResidual) Data() OperationResidualData { return v.data.clone() }
func (v OperationResidual) Clone() OperationResidual    { return v }
func (d OperationResidualData) clone() OperationResidualData {
	d.Recovery = cloneSlice(d.Recovery)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d OperationResidualData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Correlation.Valid() {
		return invalid("d.Correlation")
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
	return validateNormalizedRecovery(d.Effects, d.Recovery)
}

// PendingRuntimeAcknowledgment is an immutable semantic boundary value. Its zero is invalid.
type PendingRuntimeAcknowledgment struct {
	data        PendingRuntimeAcknowledgmentData
	initialized bool
}

// PendingRuntimeAcknowledgmentData is a construction/copy value; NewPendingRuntimeAcknowledgment validates and owns a copy.
type PendingRuntimeAcknowledgmentData struct {
	RuntimeSequence RuntimeEventSequence
	SessionID       domain.SessionID
	SessionSequence SessionSequence
	Transferred     bool
	Diagnostic      Optional[Diagnostic]
}

func NewPendingRuntimeAcknowledgment(d PendingRuntimeAcknowledgmentData) (PendingRuntimeAcknowledgment, error) {
	if err := d.validate(); err != nil {
		return PendingRuntimeAcknowledgment{}, err
	}
	return PendingRuntimeAcknowledgment{data: d.clone(), initialized: true}, nil
}
func (v PendingRuntimeAcknowledgment) Valid() bool                            { return v.initialized && v.data.validate() == nil }
func (v PendingRuntimeAcknowledgment) Data() PendingRuntimeAcknowledgmentData { return v.data.clone() }
func (v PendingRuntimeAcknowledgment) Clone() PendingRuntimeAcknowledgment    { return v }
func (d PendingRuntimeAcknowledgmentData) clone() PendingRuntimeAcknowledgmentData {

	return d
}
func (d PendingRuntimeAcknowledgmentData) validate() error {
	if !d.RuntimeSequence.Valid() {
		return invalid("d.RuntimeSequence")
	}
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.SessionSequence.Valid() {
		return invalid("d.SessionSequence")
	}

	if item, ok := d.Diagnostic.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.RuntimeSequence.Value() == 0 {
		return invalid("pending ACK sequence")
	}
	return nil
}

// ShutdownResult is an immutable semantic boundary value. Its zero is invalid.
type ShutdownResult struct {
	data        ShutdownResultData
	initialized bool
}

// ShutdownResultData is a construction/copy value; NewShutdownResult validates and owns a copy.
type ShutdownResultData struct {
	Complete               bool
	Operations             []OperationResidual
	Sessions               RuntimeShutdownResult
	PendingAcknowledgments []PendingRuntimeAcknowledgment
	Recovery               []NormalizedRecovery
	Diagnostics            []Diagnostic
}

func NewShutdownResult(d ShutdownResultData) (ShutdownResult, error) {
	if err := d.validate(); err != nil {
		return ShutdownResult{}, err
	}
	return ShutdownResult{data: d.clone(), initialized: true}, nil
}
func (v ShutdownResult) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v ShutdownResult) Data() ShutdownResultData { return v.data.clone() }
func (v ShutdownResult) Clone() ShutdownResult    { return v }
func (d ShutdownResultData) clone() ShutdownResultData {
	d.Operations = cloneSlice(d.Operations)
	d.PendingAcknowledgments = cloneSlice(d.PendingAcknowledgments)
	d.Recovery = cloneSlice(d.Recovery)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ShutdownResultData) validate() error {

	for _, item := range d.Operations {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sessions.Valid() {
		return invalid("d.Sessions")
	}
	for _, item := range d.PendingAcknowledgments {
		if !item.Valid() {
			return invalid("item")
		}
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
	if err := validateShutdownResultEvidence(d); err != nil {
		return err
	}
	if d.Complete && (len(d.Operations) > 0 || !d.Sessions.data.Complete || len(d.PendingAcknowledgments) > 0) {
		return invalid("shutdown residuals")
	}
	return nil
}
