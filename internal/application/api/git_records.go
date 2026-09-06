package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"time"
)

// GitObservation is an immutable semantic boundary value. Its zero is invalid.
type GitObservation struct {
	data        GitObservationData
	initialized bool
}

// GitObservationData is a construction/copy value; NewGitObservation validates and owns a copy.
type GitObservationData struct {
	ID           ObservationID
	Repository   domain.RepositoryID
	Worktree     Optional[domain.WorktreeID]
	Interval     ObservationInterval
	Version      SourceVersion
	Completeness Completeness
}

func NewGitObservation(d GitObservationData) (GitObservation, error) {
	if err := d.validate(); err != nil {
		return GitObservation{}, err
	}
	return GitObservation{data: d.clone(), initialized: true}, nil
}
func (v GitObservation) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v GitObservation) Data() GitObservationData { return v.data.clone() }
func (v GitObservation) Clone() GitObservation    { return v }
func (d GitObservationData) clone() GitObservationData {

	return d
}
func (d GitObservationData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
	}
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Interval.Valid() {
		return invalid("d.Interval")
	}
	if !d.Version.Valid() {
		return invalid("d.Version")
	}
	if !d.Completeness.Valid() {
		return invalid("d.Completeness")
	}
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("Git repository scope")
	}
	if w, ok := d.Worktree.Value(); ok && w.Repository() != d.Repository {
		return invalid("Git observation worktree")
	}
	return nil
}

// GitCapabilityFact is an immutable semantic boundary value. Its zero is invalid.
type GitCapabilityFact struct {
	data        GitCapabilityFactData
	initialized bool
}

// GitCapabilityFactData is a construction/copy value; NewGitCapabilityFact validates and owns a copy.
type GitCapabilityFactData struct {
	Operation    string
	Supported    bool
	Prerequisite string
	Diagnostics  []Diagnostic
}

func NewGitCapabilityFact(d GitCapabilityFactData) (GitCapabilityFact, error) {
	if err := d.validate(); err != nil {
		return GitCapabilityFact{}, err
	}
	return GitCapabilityFact{data: d.clone(), initialized: true}, nil
}
func (v GitCapabilityFact) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v GitCapabilityFact) Data() GitCapabilityFactData { return v.data.clone() }
func (v GitCapabilityFact) Clone() GitCapabilityFact    { return v }
func (d GitCapabilityFactData) clone() GitCapabilityFactData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d GitCapabilityFactData) validate() error {

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !nonempty(d.Operation) {
		return invalid("Git capability")
	}
	return nil
}

// GitCapabilities is an immutable semantic boundary value. Its zero is invalid.
type GitCapabilities struct {
	data        GitCapabilitiesData
	initialized bool
}

// GitCapabilitiesData is a construction/copy value; NewGitCapabilities validates and owns a copy.
type GitCapabilitiesData struct {
	ObjectFormat domain.ObjectFormat
	RefBackend   RefBackend
	GitVersion   string
	Profile      string
	Capabilities []GitCapabilityFact
}

func NewGitCapabilities(d GitCapabilitiesData) (GitCapabilities, error) {
	if err := d.validate(); err != nil {
		return GitCapabilities{}, err
	}
	return GitCapabilities{data: d.clone(), initialized: true}, nil
}
func (v GitCapabilities) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v GitCapabilities) Data() GitCapabilitiesData { return v.data.clone() }
func (v GitCapabilities) Clone() GitCapabilities    { return v }
func (d GitCapabilitiesData) clone() GitCapabilitiesData {
	d.Capabilities = cloneSlice(d.Capabilities)
	return d
}
func (d GitCapabilitiesData) validate() error {
	if !d.ObjectFormat.Valid() {
		return invalid("d.ObjectFormat")
	}
	if !d.RefBackend.Valid() {
		return invalid("d.RefBackend")
	}

	for _, item := range d.Capabilities {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !nonempty(d.GitVersion) || !nonempty(d.Profile) {
		return invalid("Git profile")
	}
	return nil
}

// AvailableWorktree is an immutable semantic boundary value. Its zero is invalid.
type AvailableWorktree struct {
	data        AvailableWorktreeData
	initialized bool
}

// AvailableWorktreeData is a construction/copy value; NewAvailableWorktree validates and owns a copy.
type AvailableWorktreeData struct {
}

func NewAvailableWorktree(d AvailableWorktreeData) (AvailableWorktree, error) {
	if err := d.validate(); err != nil {
		return AvailableWorktree{}, err
	}
	return AvailableWorktree{data: d.clone(), initialized: true}, nil
}
func (v AvailableWorktree) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v AvailableWorktree) Data() AvailableWorktreeData { return v.data.clone() }
func (v AvailableWorktree) Clone() AvailableWorktree    { return v }
func (d AvailableWorktreeData) clone() AvailableWorktreeData {

	return d
}
func (d AvailableWorktreeData) validate() error {

	return nil
}

// LockedWorktree is an immutable semantic boundary value. Its zero is invalid.
type LockedWorktree struct {
	data        LockedWorktreeData
	initialized bool
}

// LockedWorktreeData is a construction/copy value; NewLockedWorktree validates and owns a copy.
type LockedWorktreeData struct {
	Reason string
}

func NewLockedWorktree(d LockedWorktreeData) (LockedWorktree, error) {
	if err := d.validate(); err != nil {
		return LockedWorktree{}, err
	}
	return LockedWorktree{data: d.clone(), initialized: true}, nil
}
func (v LockedWorktree) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v LockedWorktree) Data() LockedWorktreeData { return v.data.clone() }
func (v LockedWorktree) Clone() LockedWorktree    { return v }
func (d LockedWorktreeData) clone() LockedWorktreeData {

	return d
}
func (d LockedWorktreeData) validate() error {

	return nil
}

// PrunableWorktree is an immutable semantic boundary value. Its zero is invalid.
type PrunableWorktree struct {
	data        PrunableWorktreeData
	initialized bool
}

// PrunableWorktreeData is a construction/copy value; NewPrunableWorktree validates and owns a copy.
type PrunableWorktreeData struct {
	Reason string
}

func NewPrunableWorktree(d PrunableWorktreeData) (PrunableWorktree, error) {
	if err := d.validate(); err != nil {
		return PrunableWorktree{}, err
	}
	return PrunableWorktree{data: d.clone(), initialized: true}, nil
}
func (v PrunableWorktree) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v PrunableWorktree) Data() PrunableWorktreeData { return v.data.clone() }
func (v PrunableWorktree) Clone() PrunableWorktree    { return v }
func (d PrunableWorktreeData) clone() PrunableWorktreeData {

	return d
}
func (d PrunableWorktreeData) validate() error {

	return nil
}

// MissingWorktree is an immutable semantic boundary value. Its zero is invalid.
type MissingWorktree struct {
	data        MissingWorktreeData
	initialized bool
}

// MissingWorktreeData is a construction/copy value; NewMissingWorktree validates and owns a copy.
type MissingWorktreeData struct {
}

func NewMissingWorktree(d MissingWorktreeData) (MissingWorktree, error) {
	if err := d.validate(); err != nil {
		return MissingWorktree{}, err
	}
	return MissingWorktree{data: d.clone(), initialized: true}, nil
}
func (v MissingWorktree) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v MissingWorktree) Data() MissingWorktreeData { return v.data.clone() }
func (v MissingWorktree) Clone() MissingWorktree    { return v }
func (d MissingWorktreeData) clone() MissingWorktreeData {

	return d
}
func (d MissingWorktreeData) validate() error {

	return nil
}

// UnresolvedWorktree is an immutable semantic boundary value. Its zero is invalid.
type UnresolvedWorktree struct {
	data        UnresolvedWorktreeData
	initialized bool
}

// UnresolvedWorktreeData is a construction/copy value; NewUnresolvedWorktree validates and owns a copy.
type UnresolvedWorktreeData struct {
	Diagnostic Diagnostic
}

func NewUnresolvedWorktree(d UnresolvedWorktreeData) (UnresolvedWorktree, error) {
	if err := d.validate(); err != nil {
		return UnresolvedWorktree{}, err
	}
	return UnresolvedWorktree{data: d.clone(), initialized: true}, nil
}
func (v UnresolvedWorktree) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v UnresolvedWorktree) Data() UnresolvedWorktreeData { return v.data.clone() }
func (v UnresolvedWorktree) Clone() UnresolvedWorktree    { return v }
func (d UnresolvedWorktreeData) clone() UnresolvedWorktreeData {

	return d
}
func (d UnresolvedWorktreeData) validate() error {
	if !d.Diagnostic.Valid() {
		return invalid("d.Diagnostic")
	}

	return nil
}

// WorktreeFacts is an immutable semantic boundary value. Its zero is invalid.
type WorktreeFacts struct {
	data        WorktreeFactsData
	initialized bool
}

// WorktreeFactsData is a construction/copy value; NewWorktreeFacts validates and owns a copy.
type WorktreeFactsData struct {
	ID           domain.WorktreeID
	Scope        Optional[WorktreeScope]
	Head         Optional[domain.Head]
	Primary      bool
	Current      bool
	Availability WorktreeAvailability
	Occupancy    []BranchOccupancy
	Observation  GitObservation
}

func NewWorktreeFacts(d WorktreeFactsData) (WorktreeFacts, error) {
	if err := d.validate(); err != nil {
		return WorktreeFacts{}, err
	}
	return WorktreeFacts{data: d.clone(), initialized: true}, nil
}
func (v WorktreeFacts) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v WorktreeFacts) Data() WorktreeFactsData { return v.data.clone() }
func (v WorktreeFacts) Clone() WorktreeFacts    { return v }
func (d WorktreeFactsData) clone() WorktreeFactsData {
	d.Occupancy = cloneSlice(d.Occupancy)
	return d
}
func (d WorktreeFactsData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
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

	if !validWorktreeAvailability(d.Availability) {
		return invalid("d.Availability")
	}
	for _, item := range d.Occupancy {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	return validateWorktree(d)
}

// BranchOccupancy is an immutable semantic boundary value. Its zero is invalid.
type BranchOccupancy struct {
	data        BranchOccupancyData
	initialized bool
}

// BranchOccupancyData is a construction/copy value; NewBranchOccupancy validates and owns a copy.
type BranchOccupancyData struct {
	Branch      domain.BranchID
	Worktrees   []domain.WorktreeID
	Observation GitObservation
}

func NewBranchOccupancy(d BranchOccupancyData) (BranchOccupancy, error) {
	if err := d.validate(); err != nil {
		return BranchOccupancy{}, err
	}
	return BranchOccupancy{data: d.clone(), initialized: true}, nil
}
func (v BranchOccupancy) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v BranchOccupancy) Data() BranchOccupancyData { return v.data.clone() }
func (v BranchOccupancy) Clone() BranchOccupancy    { return v }
func (d BranchOccupancyData) clone() BranchOccupancyData {
	d.Worktrees = cloneSlice(d.Worktrees)
	return d
}
func (d BranchOccupancyData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Branch.Kind() != domain.Local || d.Branch.Repository() != d.Observation.data.Repository {
		return invalid("occupancy branch")
	}
	for _, w := range d.Worktrees {
		if w.Repository() != d.Branch.Repository() {
			return invalid("occupancy worktree")
		}
	}
	return nil
}

// LocalRepositoryFacts is an immutable semantic boundary value. Its zero is invalid.
type LocalRepositoryFacts struct {
	data        LocalRepositoryFactsData
	initialized bool
}

// LocalRepositoryFactsData is a construction/copy value; NewLocalRepositoryFacts validates and owns a copy.
type LocalRepositoryFactsData struct {
	Repository      domain.RepositoryID
	CommonDirectory string
	Worktrees       []WorktreeFacts
	Remotes         []RemoteBinding
	Capabilities    GitCapabilities
	Observation     GitObservation
}

func NewLocalRepositoryFacts(d LocalRepositoryFactsData) (LocalRepositoryFacts, error) {
	if err := d.validate(); err != nil {
		return LocalRepositoryFacts{}, err
	}
	return LocalRepositoryFacts{data: d.clone(), initialized: true}, nil
}
func (v LocalRepositoryFacts) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v LocalRepositoryFacts) Data() LocalRepositoryFactsData { return v.data.clone() }
func (v LocalRepositoryFacts) Clone() LocalRepositoryFacts    { return v }
func (d LocalRepositoryFactsData) clone() LocalRepositoryFactsData {
	d.Worktrees = cloneSlice(d.Worktrees)
	d.Remotes = cloneSlice(d.Remotes)
	return d
}
func (d LocalRepositoryFactsData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Remotes {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Capabilities.Valid() {
		return invalid("d.Capabilities")
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Repository != d.Observation.data.Repository || !nonempty(d.CommonDirectory) {
		return invalid("local facts")
	}
	for _, w := range d.Worktrees {
		if w.data.ID.Repository() != d.Repository {
			return invalid("worktree scope")
		}
	}
	for _, r := range d.Remotes {
		if r.data.LocalRepository != d.Repository {
			return invalid("remote binding")
		}
	}
	return nil
}

// AbsentFile is an immutable semantic boundary value. Its zero is invalid.
type AbsentFile struct {
	data        AbsentFileData
	initialized bool
}

// AbsentFileData is a construction/copy value; NewAbsentFile validates and owns a copy.
type AbsentFileData struct {
	Path GitPath
}

func NewAbsentFile(d AbsentFileData) (AbsentFile, error) {
	if err := d.validate(); err != nil {
		return AbsentFile{}, err
	}
	return AbsentFile{data: d.clone(), initialized: true}, nil
}
func (v AbsentFile) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v AbsentFile) Data() AbsentFileData { return v.data.clone() }
func (v AbsentFile) Clone() AbsentFile    { return v }
func (d AbsentFileData) clone() AbsentFileData {

	return d
}
func (d AbsentFileData) validate() error {
	if !d.Path.Valid() {
		return invalid("d.Path")
	}

	return nil
}

// PresentFile is an immutable semantic boundary value. Its zero is invalid.
type PresentFile struct {
	data        PresentFileData
	initialized bool
}

// PresentFileData is a construction/copy value; NewPresentFile validates and owns a copy.
type PresentFileData struct {
	Path           GitPath
	ObjectIdentity SourceVersion
	Kind           FileKind
	Mode           uint32
	Content        SourceVersion
	LinkTarget     Optional[string]
	ParentIdentity SourceVersion
}

func NewPresentFile(d PresentFileData) (PresentFile, error) {
	if err := d.validate(); err != nil {
		return PresentFile{}, err
	}
	return PresentFile{data: d.clone(), initialized: true}, nil
}
func (v PresentFile) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v PresentFile) Data() PresentFileData { return v.data.clone() }
func (v PresentFile) Clone() PresentFile    { return v }
func (d PresentFileData) clone() PresentFileData {

	return d
}
func (d PresentFileData) validate() error {
	if !d.Path.Valid() {
		return invalid("d.Path")
	}
	if !d.ObjectIdentity.Valid() {
		return invalid("d.ObjectIdentity")
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}

	if !d.Content.Valid() {
		return invalid("d.Content")
	}

	if !d.ParentIdentity.Valid() {
		return invalid("d.ParentIdentity")
	}
	if (d.Kind == SymlinkFile) != d.LinkTarget.Present() {
		return invalid("link target")
	}
	return nil
}

// IndexEntryFact is an immutable semantic boundary value. Its zero is invalid.
type IndexEntryFact struct {
	data        IndexEntryFactData
	initialized bool
}

// IndexEntryFactData is a construction/copy value; NewIndexEntryFact validates and owns a copy.
type IndexEntryFactData struct {
	Path          GitPath
	Stage         uint8
	Object        domain.OID
	Mode          uint32
	SemanticFlags []IndexFlag
}

func NewIndexEntryFact(d IndexEntryFactData) (IndexEntryFact, error) {
	if err := d.validate(); err != nil {
		return IndexEntryFact{}, err
	}
	return IndexEntryFact{data: d.clone(), initialized: true}, nil
}
func (v IndexEntryFact) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v IndexEntryFact) Data() IndexEntryFactData { return v.data.clone() }
func (v IndexEntryFact) Clone() IndexEntryFact    { return v }
func (d IndexEntryFactData) clone() IndexEntryFactData {
	d.SemanticFlags = cloneSlice(d.SemanticFlags)
	return d
}
func (d IndexEntryFactData) validate() error {
	if !d.Path.Valid() {
		return invalid("d.Path")
	}

	if !d.Object.Valid() {
		return invalid("d.Object")
	}

	for _, item := range d.SemanticFlags {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Stage > 3 {
		return invalid("index stage")
	}
	return nil
}

// ChangeFact is an immutable semantic boundary value. Its zero is invalid.
type ChangeFact struct {
	data        ChangeFactData
	initialized bool
}

// ChangeFactData is a construction/copy value; NewChangeFact validates and owns a copy.
type ChangeFactData struct {
	Path          GitPath
	OldPath       Optional[GitPath]
	Kind          ChangeKind
	IndexEntries  []IndexEntryFact
	WorktreeState FileState
}

func NewChangeFact(d ChangeFactData) (ChangeFact, error) {
	if err := d.validate(); err != nil {
		return ChangeFact{}, err
	}
	return ChangeFact{data: d.clone(), initialized: true}, nil
}
func (v ChangeFact) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v ChangeFact) Data() ChangeFactData { return v.data.clone() }
func (v ChangeFact) Clone() ChangeFact    { return v }
func (d ChangeFactData) clone() ChangeFactData {
	d.IndexEntries = cloneSlice(d.IndexEntries)
	return d
}
func (d ChangeFactData) validate() error {
	if !d.Path.Valid() {
		return invalid("d.Path")
	}
	if item, ok := d.OldPath.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	for _, item := range d.IndexEntries {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validFileState(d.WorktreeState) {
		return invalid("d.WorktreeState")
	}
	if filePath(d.WorktreeState) != d.Path {
		return invalid("change file path")
	}
	for _, e := range d.IndexEntries {
		if e.data.Path != d.Path {
			return invalid("index path")
		}
	}
	return nil
}

// StatusFacts is an immutable semantic boundary value. Its zero is invalid.
type StatusFacts struct {
	data        StatusFactsData
	initialized bool
}

// StatusFactsData is a construction/copy value; NewStatusFacts validates and owns a copy.
type StatusFactsData struct {
	Worktree             WorktreeFacts
	Changes              []ChangeFact
	IndexVersion         SourceVersion
	WorktreeVersion      SourceVersion
	ConfigurationVersion SourceVersion
	Upstream             UpstreamFact
	Observation          GitObservation
}

func NewStatusFacts(d StatusFactsData) (StatusFacts, error) {
	if err := d.validate(); err != nil {
		return StatusFacts{}, err
	}
	return StatusFacts{data: d.clone(), initialized: true}, nil
}
func (v StatusFacts) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v StatusFacts) Data() StatusFactsData { return v.data.clone() }
func (v StatusFacts) Clone() StatusFacts    { return v }
func (d StatusFactsData) clone() StatusFactsData {
	d.Changes = cloneSlice(d.Changes)
	return d
}
func (d StatusFactsData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	for _, item := range d.Changes {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.IndexVersion.Valid() {
		return invalid("d.IndexVersion")
	}
	if !d.WorktreeVersion.Valid() {
		return invalid("d.WorktreeVersion")
	}
	if !d.ConfigurationVersion.Valid() {
		return invalid("d.ConfigurationVersion")
	}
	if !validUpstreamFact(d.Upstream) {
		return invalid("d.Upstream")
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Worktree.data.ID.Repository() != d.Observation.data.Repository {
		return invalid("status scope")
	}
	return nil
}

// NoUpstream is an immutable semantic boundary value. Its zero is invalid.
type NoUpstream struct {
	data        NoUpstreamData
	initialized bool
}

// NoUpstreamData is a construction/copy value; NewNoUpstream validates and owns a copy.
type NoUpstreamData struct {
}

func NewNoUpstream(d NoUpstreamData) (NoUpstream, error) {
	if err := d.validate(); err != nil {
		return NoUpstream{}, err
	}
	return NoUpstream{data: d.clone(), initialized: true}, nil
}
func (v NoUpstream) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v NoUpstream) Data() NoUpstreamData { return v.data.clone() }
func (v NoUpstream) Clone() NoUpstream    { return v }
func (d NoUpstreamData) clone() NoUpstreamData {

	return d
}
func (d NoUpstreamData) validate() error {

	return nil
}

// UpstreamNotApplicable is an immutable semantic boundary value. Its zero is invalid.
type UpstreamNotApplicable struct {
	data        UpstreamNotApplicableData
	initialized bool
}

// UpstreamNotApplicableData is a construction/copy value; NewUpstreamNotApplicable validates and owns a copy.
type UpstreamNotApplicableData struct {
}

func NewUpstreamNotApplicable(d UpstreamNotApplicableData) (UpstreamNotApplicable, error) {
	if err := d.validate(); err != nil {
		return UpstreamNotApplicable{}, err
	}
	return UpstreamNotApplicable{data: d.clone(), initialized: true}, nil
}
func (v UpstreamNotApplicable) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v UpstreamNotApplicable) Data() UpstreamNotApplicableData { return v.data.clone() }
func (v UpstreamNotApplicable) Clone() UpstreamNotApplicable    { return v }
func (d UpstreamNotApplicableData) clone() UpstreamNotApplicableData {

	return d
}
func (d UpstreamNotApplicableData) validate() error {

	return nil
}

// GoneUpstream is an immutable semantic boundary value. Its zero is invalid.
type GoneUpstream struct {
	data        GoneUpstreamData
	initialized bool
}

// GoneUpstreamData is a construction/copy value; NewGoneUpstream validates and owns a copy.
type GoneUpstreamData struct {
	Binding  RemoteBinding
	Ref      GitRefLocator
	Evidence GitObservation
}

func NewGoneUpstream(d GoneUpstreamData) (GoneUpstream, error) {
	if err := d.validate(); err != nil {
		return GoneUpstream{}, err
	}
	return GoneUpstream{data: d.clone(), initialized: true}, nil
}
func (v GoneUpstream) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v GoneUpstream) Data() GoneUpstreamData { return v.data.clone() }
func (v GoneUpstream) Clone() GoneUpstream    { return v }
func (d GoneUpstreamData) clone() GoneUpstreamData {

	return d
}
func (d GoneUpstreamData) validate() error {
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	if !validGitRefLocator(d.Ref) {
		return invalid("d.Ref")
	}
	if !d.Evidence.Valid() {
		return invalid("d.Evidence")
	}
	if d.Evidence.data.Completeness != Complete {
		return invalid("gone evidence")
	}
	return nil
}

// UnresolvedUpstream is an immutable semantic boundary value. Its zero is invalid.
type UnresolvedUpstream struct {
	data        UnresolvedUpstreamData
	initialized bool
}

// UnresolvedUpstreamData is a construction/copy value; NewUnresolvedUpstream validates and owns a copy.
type UnresolvedUpstreamData struct {
	Binding    Optional[RemoteBinding]
	Ref        Optional[GitRefLocator]
	Diagnostic Diagnostic
}

func NewUnresolvedUpstream(d UnresolvedUpstreamData) (UnresolvedUpstream, error) {
	if err := d.validate(); err != nil {
		return UnresolvedUpstream{}, err
	}
	return UnresolvedUpstream{data: d.clone(), initialized: true}, nil
}
func (v UnresolvedUpstream) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v UnresolvedUpstream) Data() UnresolvedUpstreamData { return v.data.clone() }
func (v UnresolvedUpstream) Clone() UnresolvedUpstream    { return v }
func (d UnresolvedUpstreamData) clone() UnresolvedUpstreamData {

	return d
}
func (d UnresolvedUpstreamData) validate() error {
	if item, ok := d.Binding.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Ref.Value(); ok {
		if !validGitRefLocator(item) {
			return invalid("item")
		}
	}
	if !d.Diagnostic.Valid() {
		return invalid("d.Diagnostic")
	}

	return nil
}

// ResolvedUpstream is an immutable semantic boundary value. Its zero is invalid.
type ResolvedUpstream struct {
	data        ResolvedUpstreamData
	initialized bool
}

// ResolvedUpstreamData is a construction/copy value; NewResolvedUpstream validates and owns a copy.
type ResolvedUpstreamData struct {
	Binding        RemoteBinding
	RemoteBranch   domain.BranchID
	CachedLocalRef GitRefLocator
	Local          domain.Revision
	Upstream       domain.Revision
	Comparison     RevisionComparison
	Freshness      FetchFreshness
}

func NewResolvedUpstream(d ResolvedUpstreamData) (ResolvedUpstream, error) {
	if err := d.validate(); err != nil {
		return ResolvedUpstream{}, err
	}
	return ResolvedUpstream{data: d.clone(), initialized: true}, nil
}
func (v ResolvedUpstream) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v ResolvedUpstream) Data() ResolvedUpstreamData { return v.data.clone() }
func (v ResolvedUpstream) Clone() ResolvedUpstream    { return v }
func (d ResolvedUpstreamData) clone() ResolvedUpstreamData {

	return d
}
func (d ResolvedUpstreamData) validate() error {
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	if !d.RemoteBranch.Valid() {
		return invalid("d.RemoteBranch")
	}
	if !validGitRefLocator(d.CachedLocalRef) {
		return invalid("d.CachedLocalRef")
	}
	if !d.Local.Valid() {
		return invalid("d.Local")
	}
	if !d.Upstream.Valid() {
		return invalid("d.Upstream")
	}
	if !d.Comparison.Valid() {
		return invalid("d.Comparison")
	}
	if !d.Freshness.Valid() {
		return invalid("d.Freshness")
	}
	if d.RemoteBranch.Repository() != d.Binding.data.RemoteRepository || d.Local.Repository() != d.Binding.data.LocalRepository || d.Upstream.Repository() != d.Local.Repository() || d.Comparison.data.Left != d.Local || d.Comparison.data.Right != d.Upstream {
		return invalid("upstream endpoints")
	}
	return nil
}

// RevisionComparison is an immutable semantic boundary value. Its zero is invalid.
type RevisionComparison struct {
	data        RevisionComparisonData
	initialized bool
}

// RevisionComparisonData is a construction/copy value; NewRevisionComparison validates and owns a copy.
type RevisionComparisonData struct {
	Left        domain.Revision
	Right       domain.Revision
	Ahead       Optional[uint64]
	Behind      Optional[uint64]
	Observation GitObservation
}

func NewRevisionComparison(d RevisionComparisonData) (RevisionComparison, error) {
	if err := d.validate(); err != nil {
		return RevisionComparison{}, err
	}
	return RevisionComparison{data: d.clone(), initialized: true}, nil
}
func (v RevisionComparison) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v RevisionComparison) Data() RevisionComparisonData { return v.data.clone() }
func (v RevisionComparison) Clone() RevisionComparison    { return v }
func (d RevisionComparisonData) clone() RevisionComparisonData {

	return d
}
func (d RevisionComparisonData) validate() error {
	if !d.Left.Valid() {
		return invalid("d.Left")
	}
	if !d.Right.Valid() {
		return invalid("d.Right")
	}

	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if !sameLocal(d.Left, d.Right) || d.Left.Repository() != d.Observation.data.Repository || d.Ahead.Present() != d.Behind.Present() {
		return invalid("revision comparison")
	}
	return nil
}

// FetchGeneration is an immutable semantic boundary value. Its zero is invalid.
type FetchGeneration struct {
	data        FetchGenerationData
	initialized bool
}

// FetchGenerationData is a construction/copy value; NewFetchGeneration validates and owns a copy.
type FetchGenerationData struct {
	Repository domain.RepositoryID
	Binding    RemoteBinding
	RefScope   []GitRefLocator
	Issuer     string
	Counter    uint64
}

func NewFetchGeneration(d FetchGenerationData) (FetchGeneration, error) {
	if err := d.validate(); err != nil {
		return FetchGeneration{}, err
	}
	return FetchGeneration{data: d.clone(), initialized: true}, nil
}
func (v FetchGeneration) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v FetchGeneration) Data() FetchGenerationData { return v.data.clone() }
func (v FetchGeneration) Clone() FetchGeneration    { return v }
func (d FetchGenerationData) clone() FetchGenerationData {
	d.RefScope = cloneSlice(d.RefScope)
	return d
}
func (d FetchGenerationData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	for _, item := range d.RefScope {
		if !validGitRefLocator(item) {
			return invalid("item")
		}
	}

	if d.Repository.Scope() != domain.LocalCommon || d.Repository != d.Binding.data.LocalRepository || !nonempty(d.Issuer) || d.Counter == 0 {
		return invalid("fetch generation")
	}
	return nil
}

// FetchFreshness is an immutable semantic boundary value. Its zero is invalid.
type FetchFreshness struct {
	data        FetchFreshnessData
	initialized bool
}

// FetchFreshnessData is a construction/copy value; NewFetchFreshness validates and owns a copy.
type FetchFreshnessData struct {
	Kind        FetchFreshnessKind
	Binding     RemoteBinding
	RefScope    []GitRefLocator
	Generation  Optional[FetchGeneration]
	Observation GitObservation
}

func NewFetchFreshness(d FetchFreshnessData) (FetchFreshness, error) {
	if err := d.validate(); err != nil {
		return FetchFreshness{}, err
	}
	return FetchFreshness{data: d.clone(), initialized: true}, nil
}
func (v FetchFreshness) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v FetchFreshness) Data() FetchFreshnessData { return v.data.clone() }
func (v FetchFreshness) Clone() FetchFreshness    { return v }
func (d FetchFreshnessData) clone() FetchFreshnessData {
	d.RefScope = cloneSlice(d.RefScope)
	return d
}
func (d FetchFreshnessData) validate() error {
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	for _, item := range d.RefScope {
		if !validGitRefLocator(item) {
			return invalid("item")
		}
	}
	if item, ok := d.Generation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Kind == Refreshed && !d.Generation.Present() {
		return invalid("refreshed generation")
	}
	if g, ok := d.Generation.Value(); ok && g.data.Repository != d.Binding.data.LocalRepository {
		return invalid("generation scope")
	}
	return nil
}

// LocalBranchRef is an immutable semantic boundary value. Its zero is invalid.
type LocalBranchRef struct {
	data        LocalBranchRefData
	initialized bool
}

// LocalBranchRefData is a construction/copy value; NewLocalBranchRef validates and owns a copy.
type LocalBranchRefData struct {
	Branch domain.BranchID
}

func NewLocalBranchRef(d LocalBranchRefData) (LocalBranchRef, error) {
	if err := d.validate(); err != nil {
		return LocalBranchRef{}, err
	}
	return LocalBranchRef{data: d.clone(), initialized: true}, nil
}
func (v LocalBranchRef) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v LocalBranchRef) Data() LocalBranchRefData { return v.data.clone() }
func (v LocalBranchRef) Clone() LocalBranchRef    { return v }
func (d LocalBranchRefData) clone() LocalBranchRefData {

	return d
}
func (d LocalBranchRefData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if d.Branch.Kind() != domain.Local {
		return invalid("local branch ref")
	}
	return nil
}

// LocalTagRef is an immutable semantic boundary value. Its zero is invalid.
type LocalTagRef struct {
	data        LocalTagRefData
	initialized bool
}

// LocalTagRefData is a construction/copy value; NewLocalTagRef validates and owns a copy.
type LocalTagRefData struct {
	Repository domain.RepositoryID
	Ref        string
}

func NewLocalTagRef(d LocalTagRefData) (LocalTagRef, error) {
	if err := d.validate(); err != nil {
		return LocalTagRef{}, err
	}
	return LocalTagRef{data: d.clone(), initialized: true}, nil
}
func (v LocalTagRef) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v LocalTagRef) Data() LocalTagRefData { return v.data.clone() }
func (v LocalTagRef) Clone() LocalTagRef    { return v }
func (d LocalTagRefData) clone() LocalTagRefData {

	return d
}
func (d LocalTagRefData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}

	if d.Repository.Scope() != domain.LocalCommon || !validFullRef(d.Ref, "refs/tags/") {
		return invalid("tag ref")
	}
	return nil
}

// CachedRemoteRef is an immutable semantic boundary value. Its zero is invalid.
type CachedRemoteRef struct {
	data        CachedRemoteRefData
	initialized bool
}

// CachedRemoteRefData is a construction/copy value; NewCachedRemoteRef validates and owns a copy.
type CachedRemoteRefData struct {
	Binding RemoteBinding
	Ref     string
}

func NewCachedRemoteRef(d CachedRemoteRefData) (CachedRemoteRef, error) {
	if err := d.validate(); err != nil {
		return CachedRemoteRef{}, err
	}
	return CachedRemoteRef{data: d.clone(), initialized: true}, nil
}
func (v CachedRemoteRef) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v CachedRemoteRef) Data() CachedRemoteRefData { return v.data.clone() }
func (v CachedRemoteRef) Clone() CachedRemoteRef    { return v }
func (d CachedRemoteRefData) clone() CachedRemoteRefData {

	return d
}
func (d CachedRemoteRefData) validate() error {
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}

	if !validFullRef(d.Ref, "refs/") {
		return invalid("cached ref")
	}
	return nil
}

// RemoteRef is an immutable semantic boundary value. Its zero is invalid.
type RemoteRef struct {
	data        RemoteRefData
	initialized bool
}

// RemoteRefData is a construction/copy value; NewRemoteRef validates and owns a copy.
type RemoteRefData struct {
	Binding RemoteBinding
	Ref     string
}

func NewRemoteRef(d RemoteRefData) (RemoteRef, error) {
	if err := d.validate(); err != nil {
		return RemoteRef{}, err
	}
	return RemoteRef{data: d.clone(), initialized: true}, nil
}
func (v RemoteRef) Valid() bool         { return v.initialized && v.data.validate() == nil }
func (v RemoteRef) Data() RemoteRefData { return v.data.clone() }
func (v RemoteRef) Clone() RemoteRef    { return v }
func (d RemoteRefData) clone() RemoteRefData {

	return d
}
func (d RemoteRefData) validate() error {
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}

	if !validFullRef(d.Ref, "refs/") {
		return invalid("remote ref")
	}
	return nil
}

// RefFact is an immutable semantic boundary value. Its zero is invalid.
type RefFact struct {
	data        RefFactData
	initialized bool
}

// RefFactData is a construction/copy value; NewRefFact validates and owns a copy.
type RefFactData struct {
	Locator        GitRefLocator
	Revision       Optional[domain.Revision]
	SymbolicTarget Optional[GitRefLocator]
	Freshness      Optional[FetchFreshness]
	TagObject      Optional[domain.OID]
	Observation    GitObservation
}

func NewRefFact(d RefFactData) (RefFact, error) {
	if err := d.validate(); err != nil {
		return RefFact{}, err
	}
	return RefFact{data: d.clone(), initialized: true}, nil
}
func (v RefFact) Valid() bool       { return v.initialized && v.data.validate() == nil }
func (v RefFact) Data() RefFactData { return v.data.clone() }
func (v RefFact) Clone() RefFact    { return v }
func (d RefFactData) clone() RefFactData {

	return d
}
func (d RefFactData) validate() error {
	if !validGitRefLocator(d.Locator) {
		return invalid("d.Locator")
	}
	if item, ok := d.Revision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.SymbolicTarget.Value(); ok {
		if !validGitRefLocator(item) {
			return invalid("item")
		}
	}
	if item, ok := d.Freshness.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.TagObject.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}

	return nil
}

// ExactLocalResolution is an immutable semantic boundary value. Its zero is invalid.
type ExactLocalResolution struct {
	data        ExactLocalResolutionData
	initialized bool
}

// ExactLocalResolutionData is a construction/copy value; NewExactLocalResolution validates and owns a copy.
type ExactLocalResolutionData struct {
	Requested      domain.ExactTarget
	Local          domain.Revision
	Locator        Optional[GitRefLocator]
	Binding        Optional[RemoteBinding]
	ObservedRemote Optional[domain.Revision]
	Observation    GitObservation
}

func NewExactLocalResolution(d ExactLocalResolutionData) (ExactLocalResolution, error) {
	if err := d.validate(); err != nil {
		return ExactLocalResolution{}, err
	}
	return ExactLocalResolution{data: d.clone(), initialized: true}, nil
}
func (v ExactLocalResolution) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ExactLocalResolution) Data() ExactLocalResolutionData { return v.data.clone() }
func (v ExactLocalResolution) Clone() ExactLocalResolution    { return v }
func (d ExactLocalResolutionData) clone() ExactLocalResolutionData {

	return d
}
func (d ExactLocalResolutionData) validate() error {
	if !d.Requested.Valid() {
		return invalid("d.Requested")
	}
	if !d.Local.Valid() {
		return invalid("d.Local")
	}
	if item, ok := d.Locator.Value(); ok {
		if !validGitRefLocator(item) {
			return invalid("item")
		}
	}
	if item, ok := d.Binding.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ObservedRemote.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	return validateResolution(d)
}

// StashOrigin is an immutable semantic boundary value. Its zero is invalid.
type StashOrigin struct {
	data        StashOriginData
	initialized bool
}

// StashOriginData is a construction/copy value; NewStashOrigin validates and owns a copy.
type StashOriginData struct {
	Worktree      Optional[domain.WorktreeID]
	Branch        Optional[domain.BranchID]
	LegacyManaged Optional[string]
}

func NewStashOrigin(d StashOriginData) (StashOrigin, error) {
	if err := d.validate(); err != nil {
		return StashOrigin{}, err
	}
	return StashOrigin{data: d.clone(), initialized: true}, nil
}
func (v StashOrigin) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v StashOrigin) Data() StashOriginData { return v.data.clone() }
func (v StashOrigin) Clone() StashOrigin    { return v }
func (d StashOriginData) clone() StashOriginData {

	return d
}
func (d StashOriginData) validate() error {
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Branch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// StashFact is an immutable semantic boundary value. Its zero is invalid.
type StashFact struct {
	data        StashFactData
	initialized bool
}

// StashFactData is a construction/copy value; NewStashFact validates and owns a copy.
type StashFactData struct {
	ID              domain.StashID
	Parents         []domain.OID
	Occurrence      SourceVersion
	DisplayPosition uint64
	Message         string
	AuthorName      string
	AuthorEmail     string
	AuthorTime      Optional[time.Time]
	Origin          Optional[StashOrigin]
	Observation     GitObservation
}

func NewStashFact(d StashFactData) (StashFact, error) {
	if err := d.validate(); err != nil {
		return StashFact{}, err
	}
	return StashFact{data: d.clone(), initialized: true}, nil
}
func (v StashFact) Valid() bool         { return v.initialized && v.data.validate() == nil }
func (v StashFact) Data() StashFactData { return v.data.clone() }
func (v StashFact) Clone() StashFact    { return v }
func (d StashFactData) clone() StashFactData {
	d.Parents = cloneSlice(d.Parents)
	return d
}
func (d StashFactData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
	}
	for _, item := range d.Parents {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}

	if item, ok := d.Origin.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.ID.Repository() != d.Observation.data.Repository || len(d.Parents) < 2 || len(d.Parents) > 3 {
		return invalid("stash facts")
	}
	return nil
}

// CommitFact is an immutable semantic boundary value. Its zero is invalid.
type CommitFact struct {
	data        CommitFactData
	initialized bool
}

// CommitFactData is a construction/copy value; NewCommitFact validates and owns a copy.
type CommitFactData struct {
	Revision      domain.Revision
	Parents       []domain.Revision
	AuthorName    string
	AuthorEmail   string
	AuthorTime    time.Time
	CommitterTime time.Time
	Message       string
}

func NewCommitFact(d CommitFactData) (CommitFact, error) {
	if err := d.validate(); err != nil {
		return CommitFact{}, err
	}
	return CommitFact{data: d.clone(), initialized: true}, nil
}
func (v CommitFact) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v CommitFact) Data() CommitFactData { return v.data.clone() }
func (v CommitFact) Clone() CommitFact    { return v }
func (d CommitFactData) clone() CommitFactData {
	d.Parents = cloneSlice(d.Parents)
	return d
}
func (d CommitFactData) validate() error {
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	for _, item := range d.Parents {
		if !item.Valid() {
			return invalid("item")
		}
	}

	for _, p := range d.Parents {
		if !sameLocal(d.Revision, p) {
			return invalid("commit parent scope")
		}
	}
	return nil
}

// AllGraph is an immutable semantic boundary value. Its zero is invalid.
type AllGraph struct {
	data        AllGraphData
	initialized bool
}

// AllGraphData is a construction/copy value; NewAllGraph validates and owns a copy.
type AllGraphData struct {
	Paths []GitPath
}

func NewAllGraph(d AllGraphData) (AllGraph, error) {
	if err := d.validate(); err != nil {
		return AllGraph{}, err
	}
	return AllGraph{data: d.clone(), initialized: true}, nil
}
func (v AllGraph) Valid() bool        { return v.initialized && v.data.validate() == nil }
func (v AllGraph) Data() AllGraphData { return v.data.clone() }
func (v AllGraph) Clone() AllGraph    { return v }
func (d AllGraphData) clone() AllGraphData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d AllGraphData) validate() error {
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// ReachableFromRoots is an immutable semantic boundary value. Its zero is invalid.
type ReachableFromRoots struct {
	data        ReachableFromRootsData
	initialized bool
}

// ReachableFromRootsData is a construction/copy value; NewReachableFromRoots validates and owns a copy.
type ReachableFromRootsData struct {
	Paths []GitPath
}

func NewReachableFromRoots(d ReachableFromRootsData) (ReachableFromRoots, error) {
	if err := d.validate(); err != nil {
		return ReachableFromRoots{}, err
	}
	return ReachableFromRoots{data: d.clone(), initialized: true}, nil
}
func (v ReachableFromRoots) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v ReachableFromRoots) Data() ReachableFromRootsData { return v.data.clone() }
func (v ReachableFromRoots) Clone() ReachableFromRoots    { return v }
func (d ReachableFromRootsData) clone() ReachableFromRootsData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d ReachableFromRootsData) validate() error {
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// RootParent is an immutable semantic boundary value. Its zero is invalid.
type RootParent struct {
	data        RootParentData
	initialized bool
}

// RootParentData is a construction/copy value; NewRootParent validates and owns a copy.
type RootParentData struct {
}

func NewRootParent(d RootParentData) (RootParent, error) {
	if err := d.validate(); err != nil {
		return RootParent{}, err
	}
	return RootParent{data: d.clone(), initialized: true}, nil
}
func (v RootParent) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v RootParent) Data() RootParentData { return v.data.clone() }
func (v RootParent) Clone() RootParent    { return v }
func (d RootParentData) clone() RootParentData {

	return d
}
func (d RootParentData) validate() error {

	return nil
}

// SelectedParent is an immutable semantic boundary value. Its zero is invalid.
type SelectedParent struct {
	data        SelectedParentData
	initialized bool
}

// SelectedParentData is a construction/copy value; NewSelectedParent validates and owns a copy.
type SelectedParentData struct {
	Index uint32
}

func NewSelectedParent(d SelectedParentData) (SelectedParent, error) {
	if err := d.validate(); err != nil {
		return SelectedParent{}, err
	}
	return SelectedParent{data: d.clone(), initialized: true}, nil
}
func (v SelectedParent) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v SelectedParent) Data() SelectedParentData { return v.data.clone() }
func (v SelectedParent) Clone() SelectedParent    { return v }
func (d SelectedParentData) clone() SelectedParentData {

	return d
}
func (d SelectedParentData) validate() error {

	return nil
}

// CommitParentComparison is an immutable semantic boundary value. Its zero is invalid.
type CommitParentComparison struct {
	data        CommitParentComparisonData
	initialized bool
}

// CommitParentComparisonData is a construction/copy value; NewCommitParentComparison validates and owns a copy.
type CommitParentComparisonData struct {
	Commit domain.Revision
	Parent CommitParentSelection
}

func NewCommitParentComparison(d CommitParentComparisonData) (CommitParentComparison, error) {
	if err := d.validate(); err != nil {
		return CommitParentComparison{}, err
	}
	return CommitParentComparison{data: d.clone(), initialized: true}, nil
}
func (v CommitParentComparison) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v CommitParentComparison) Data() CommitParentComparisonData { return v.data.clone() }
func (v CommitParentComparison) Clone() CommitParentComparison    { return v }
func (d CommitParentComparisonData) clone() CommitParentComparisonData {

	return d
}
func (d CommitParentComparisonData) validate() error {
	if !d.Commit.Valid() {
		return invalid("d.Commit")
	}
	if !validCommitParentSelection(d.Parent) {
		return invalid("d.Parent")
	}
	if d.Commit.Repository().Scope() != domain.LocalCommon {
		return invalid("local commit")
	}
	return nil
}

// CommitPairComparison is an immutable semantic boundary value. Its zero is invalid.
type CommitPairComparison struct {
	data        CommitPairComparisonData
	initialized bool
}

// CommitPairComparisonData is a construction/copy value; NewCommitPairComparison validates and owns a copy.
type CommitPairComparisonData struct {
	From domain.Revision
	To   domain.Revision
}

func NewCommitPairComparison(d CommitPairComparisonData) (CommitPairComparison, error) {
	if err := d.validate(); err != nil {
		return CommitPairComparison{}, err
	}
	return CommitPairComparison{data: d.clone(), initialized: true}, nil
}
func (v CommitPairComparison) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v CommitPairComparison) Data() CommitPairComparisonData { return v.data.clone() }
func (v CommitPairComparison) Clone() CommitPairComparison    { return v }
func (d CommitPairComparisonData) clone() CommitPairComparisonData {

	return d
}
func (d CommitPairComparisonData) validate() error {
	if !d.From.Valid() {
		return invalid("d.From")
	}
	if !d.To.Valid() {
		return invalid("d.To")
	}
	if !sameLocal(d.From, d.To) {
		return invalid("comparison scope")
	}
	return nil
}

// IndexToWorktreeComparison is an immutable semantic boundary value. Its zero is invalid.
type IndexToWorktreeComparison struct {
	data        IndexToWorktreeComparisonData
	initialized bool
}

// IndexToWorktreeComparisonData is a construction/copy value; NewIndexToWorktreeComparison validates and owns a copy.
type IndexToWorktreeComparisonData struct {
	Worktree        domain.WorktreeID
	Index           SourceVersion
	WorktreeVersion SourceVersion
}

func NewIndexToWorktreeComparison(d IndexToWorktreeComparisonData) (IndexToWorktreeComparison, error) {
	if err := d.validate(); err != nil {
		return IndexToWorktreeComparison{}, err
	}
	return IndexToWorktreeComparison{data: d.clone(), initialized: true}, nil
}
func (v IndexToWorktreeComparison) Valid() bool                         { return v.initialized && v.data.validate() == nil }
func (v IndexToWorktreeComparison) Data() IndexToWorktreeComparisonData { return v.data.clone() }
func (v IndexToWorktreeComparison) Clone() IndexToWorktreeComparison    { return v }
func (d IndexToWorktreeComparisonData) clone() IndexToWorktreeComparisonData {

	return d
}
func (d IndexToWorktreeComparisonData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.Index.Valid() {
		return invalid("d.Index")
	}
	if !d.WorktreeVersion.Valid() {
		return invalid("d.WorktreeVersion")
	}

	return nil
}

// HeadToIndexComparison is an immutable semantic boundary value. Its zero is invalid.
type HeadToIndexComparison struct {
	data        HeadToIndexComparisonData
	initialized bool
}

// HeadToIndexComparisonData is a construction/copy value; NewHeadToIndexComparison validates and owns a copy.
type HeadToIndexComparisonData struct {
	Worktree    domain.WorktreeID
	Head        domain.Head
	HeadVersion SourceVersion
	Index       SourceVersion
}

func NewHeadToIndexComparison(d HeadToIndexComparisonData) (HeadToIndexComparison, error) {
	if err := d.validate(); err != nil {
		return HeadToIndexComparison{}, err
	}
	return HeadToIndexComparison{data: d.clone(), initialized: true}, nil
}
func (v HeadToIndexComparison) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v HeadToIndexComparison) Data() HeadToIndexComparisonData { return v.data.clone() }
func (v HeadToIndexComparison) Clone() HeadToIndexComparison    { return v }
func (d HeadToIndexComparisonData) clone() HeadToIndexComparisonData {

	return d
}
func (d HeadToIndexComparisonData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.Head.Valid() {
		return invalid("d.Head")
	}
	if !d.HeadVersion.Valid() {
		return invalid("d.HeadVersion")
	}
	if !d.Index.Valid() {
		return invalid("d.Index")
	}
	if !d.Head.MatchesWorktree(d.Worktree) {
		return invalid("head scope")
	}
	return nil
}

// PatchLimits is an immutable semantic boundary value. Its zero is invalid.
type PatchLimits struct {
	data        PatchLimitsData
	initialized bool
}

// PatchLimitsData is a construction/copy value; NewPatchLimits validates and owns a copy.
type PatchLimitsData struct {
	MaxBytes uint64
	MaxFiles uint32
}

func NewPatchLimits(d PatchLimitsData) (PatchLimits, error) {
	if err := d.validate(); err != nil {
		return PatchLimits{}, err
	}
	return PatchLimits{data: d.clone(), initialized: true}, nil
}
func (v PatchLimits) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v PatchLimits) Data() PatchLimitsData { return v.data.clone() }
func (v PatchLimits) Clone() PatchLimits    { return v }
func (d PatchLimitsData) clone() PatchLimitsData {

	return d
}
func (d PatchLimitsData) validate() error {

	if d.MaxBytes == 0 || d.MaxFiles == 0 {
		return invalid("patch limits")
	}
	return nil
}

// DiffFileFact is an immutable semantic boundary value. Its zero is invalid.
type DiffFileFact struct {
	data        DiffFileFactData
	initialized bool
}

// DiffFileFactData is a construction/copy value; NewDiffFileFact validates and owns a copy.
type DiffFileFactData struct {
	Path         GitPath
	OldPath      Optional[GitPath]
	Kind         ChangeKind
	AddedLines   Optional[uint64]
	DeletedLines Optional[uint64]
	Binary       bool
}

func NewDiffFileFact(d DiffFileFactData) (DiffFileFact, error) {
	if err := d.validate(); err != nil {
		return DiffFileFact{}, err
	}
	return DiffFileFact{data: d.clone(), initialized: true}, nil
}
func (v DiffFileFact) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v DiffFileFact) Data() DiffFileFactData { return v.data.clone() }
func (v DiffFileFact) Clone() DiffFileFact    { return v }
func (d DiffFileFactData) clone() DiffFileFactData {

	return d
}
func (d DiffFileFactData) validate() error {
	if !d.Path.Valid() {
		return invalid("d.Path")
	}
	if item, ok := d.OldPath.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}

	return nil
}

// PatchFacts is an immutable semantic boundary value. Its zero is invalid.
type PatchFacts struct {
	data        PatchFactsData
	initialized bool
}

// PatchFactsData is a construction/copy value; NewPatchFacts validates and owns a copy.
type PatchFactsData struct {
	Bytes         []byte
	Files         []DiffFileFact
	Truncated     bool
	OriginalBytes Optional[uint64]
	ReturnedBytes uint64
}

func NewPatchFacts(d PatchFactsData) (PatchFacts, error) {
	if err := d.validate(); err != nil {
		return PatchFacts{}, err
	}
	return PatchFacts{data: d.clone(), initialized: true}, nil
}
func (v PatchFacts) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v PatchFacts) Data() PatchFactsData { return v.data.clone() }
func (v PatchFacts) Clone() PatchFacts    { return v }
func (d PatchFactsData) clone() PatchFactsData {
	d.Bytes = cloneSlice(d.Bytes)
	d.Files = cloneSlice(d.Files)
	return d
}
func (d PatchFactsData) validate() error {

	for _, item := range d.Files {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if d.ReturnedBytes != uint64(len(d.Bytes)) {
		return invalid("patch length")
	}
	if n, ok := d.OriginalBytes.Value(); ok && n < d.ReturnedBytes {
		return invalid("original patch length")
	}
	return nil
}

// StashBaseToWorktree is an immutable semantic boundary value. Its zero is invalid.
type StashBaseToWorktree struct {
	data        StashBaseToWorktreeData
	initialized bool
}

// StashBaseToWorktreeData is a construction/copy value; NewStashBaseToWorktree validates and owns a copy.
type StashBaseToWorktreeData struct {
}

func NewStashBaseToWorktree(d StashBaseToWorktreeData) (StashBaseToWorktree, error) {
	if err := d.validate(); err != nil {
		return StashBaseToWorktree{}, err
	}
	return StashBaseToWorktree{data: d.clone(), initialized: true}, nil
}
func (v StashBaseToWorktree) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v StashBaseToWorktree) Data() StashBaseToWorktreeData { return v.data.clone() }
func (v StashBaseToWorktree) Clone() StashBaseToWorktree    { return v }
func (d StashBaseToWorktreeData) clone() StashBaseToWorktreeData {

	return d
}
func (d StashBaseToWorktreeData) validate() error {

	return nil
}

// StashBaseToIndex is an immutable semantic boundary value. Its zero is invalid.
type StashBaseToIndex struct {
	data        StashBaseToIndexData
	initialized bool
}

// StashBaseToIndexData is a construction/copy value; NewStashBaseToIndex validates and owns a copy.
type StashBaseToIndexData struct {
}

func NewStashBaseToIndex(d StashBaseToIndexData) (StashBaseToIndex, error) {
	if err := d.validate(); err != nil {
		return StashBaseToIndex{}, err
	}
	return StashBaseToIndex{data: d.clone(), initialized: true}, nil
}
func (v StashBaseToIndex) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v StashBaseToIndex) Data() StashBaseToIndexData { return v.data.clone() }
func (v StashBaseToIndex) Clone() StashBaseToIndex    { return v }
func (d StashBaseToIndexData) clone() StashBaseToIndexData {

	return d
}
func (d StashBaseToIndexData) validate() error {

	return nil
}

// StashIndexToWorktree is an immutable semantic boundary value. Its zero is invalid.
type StashIndexToWorktree struct {
	data        StashIndexToWorktreeData
	initialized bool
}

// StashIndexToWorktreeData is a construction/copy value; NewStashIndexToWorktree validates and owns a copy.
type StashIndexToWorktreeData struct {
}

func NewStashIndexToWorktree(d StashIndexToWorktreeData) (StashIndexToWorktree, error) {
	if err := d.validate(); err != nil {
		return StashIndexToWorktree{}, err
	}
	return StashIndexToWorktree{data: d.clone(), initialized: true}, nil
}
func (v StashIndexToWorktree) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v StashIndexToWorktree) Data() StashIndexToWorktreeData { return v.data.clone() }
func (v StashIndexToWorktree) Clone() StashIndexToWorktree    { return v }
func (d StashIndexToWorktreeData) clone() StashIndexToWorktreeData {

	return d
}
func (d StashIndexToWorktreeData) validate() error {

	return nil
}

// StashUntracked is an immutable semantic boundary value. Its zero is invalid.
type StashUntracked struct {
	data        StashUntrackedData
	initialized bool
}

// StashUntrackedData is a construction/copy value; NewStashUntracked validates and owns a copy.
type StashUntrackedData struct {
}

func NewStashUntracked(d StashUntrackedData) (StashUntracked, error) {
	if err := d.validate(); err != nil {
		return StashUntracked{}, err
	}
	return StashUntracked{data: d.clone(), initialized: true}, nil
}
func (v StashUntracked) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v StashUntracked) Data() StashUntrackedData { return v.data.clone() }
func (v StashUntracked) Clone() StashUntracked    { return v }
func (d StashUntrackedData) clone() StashUntrackedData {

	return d
}
func (d StashUntrackedData) validate() error {

	return nil
}

// StashParent is an immutable semantic boundary value. Its zero is invalid.
type StashParent struct {
	data        StashParentData
	initialized bool
}

// StashParentData is a construction/copy value; NewStashParent validates and owns a copy.
type StashParentData struct {
	Index uint32
}

func NewStashParent(d StashParentData) (StashParent, error) {
	if err := d.validate(); err != nil {
		return StashParent{}, err
	}
	return StashParent{data: d.clone(), initialized: true}, nil
}
func (v StashParent) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v StashParent) Data() StashParentData { return v.data.clone() }
func (v StashParent) Clone() StashParent    { return v }
func (d StashParentData) clone() StashParentData {

	return d
}
func (d StashParentData) validate() error {

	return nil
}

// StashPatchComparison is an immutable semantic boundary value. Its zero is invalid.
type StashPatchComparison struct {
	data        StashPatchComparisonData
	initialized bool
}

// StashPatchComparisonData is a construction/copy value; NewStashPatchComparison validates and owns a copy.
type StashPatchComparisonData struct {
	Stash           domain.OID
	Base            Optional[domain.OID]
	IndexParent     Optional[domain.OID]
	UntrackedParent Optional[domain.OID]
	View            StashPatchView
	FromTree        Optional[domain.OID]
	ToTree          Optional[domain.OID]
}

func NewStashPatchComparison(d StashPatchComparisonData) (StashPatchComparison, error) {
	if err := d.validate(); err != nil {
		return StashPatchComparison{}, err
	}
	return StashPatchComparison{data: d.clone(), initialized: true}, nil
}
func (v StashPatchComparison) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v StashPatchComparison) Data() StashPatchComparisonData { return v.data.clone() }
func (v StashPatchComparison) Clone() StashPatchComparison    { return v }
func (d StashPatchComparisonData) clone() StashPatchComparisonData {

	return d
}
func (d StashPatchComparisonData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if item, ok := d.Base.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.IndexParent.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.UntrackedParent.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validStashPatchView(d.View) {
		return invalid("d.View")
	}
	if item, ok := d.FromTree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ToTree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// UniqueMergeBase is an immutable semantic boundary value. Its zero is invalid.
type UniqueMergeBase struct {
	data        UniqueMergeBaseData
	initialized bool
}

// UniqueMergeBaseData is a construction/copy value; NewUniqueMergeBase validates and owns a copy.
type UniqueMergeBaseData struct {
	Base domain.Revision
}

func NewUniqueMergeBase(d UniqueMergeBaseData) (UniqueMergeBase, error) {
	if err := d.validate(); err != nil {
		return UniqueMergeBase{}, err
	}
	return UniqueMergeBase{data: d.clone(), initialized: true}, nil
}
func (v UniqueMergeBase) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v UniqueMergeBase) Data() UniqueMergeBaseData { return v.data.clone() }
func (v UniqueMergeBase) Clone() UniqueMergeBase    { return v }
func (d UniqueMergeBaseData) clone() UniqueMergeBaseData {

	return d
}
func (d UniqueMergeBaseData) validate() error {
	if !d.Base.Valid() {
		return invalid("d.Base")
	}

	return nil
}

// NoCommonAncestor is an immutable semantic boundary value. Its zero is invalid.
type NoCommonAncestor struct {
	data        NoCommonAncestorData
	initialized bool
}

// NoCommonAncestorData is a construction/copy value; NewNoCommonAncestor validates and owns a copy.
type NoCommonAncestorData struct {
}

func NewNoCommonAncestor(d NoCommonAncestorData) (NoCommonAncestor, error) {
	if err := d.validate(); err != nil {
		return NoCommonAncestor{}, err
	}
	return NoCommonAncestor{data: d.clone(), initialized: true}, nil
}
func (v NoCommonAncestor) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v NoCommonAncestor) Data() NoCommonAncestorData { return v.data.clone() }
func (v NoCommonAncestor) Clone() NoCommonAncestor    { return v }
func (d NoCommonAncestorData) clone() NoCommonAncestorData {

	return d
}
func (d NoCommonAncestorData) validate() error {

	return nil
}

// AmbiguousMergeBase is an immutable semantic boundary value. Its zero is invalid.
type AmbiguousMergeBase struct {
	data        AmbiguousMergeBaseData
	initialized bool
}

// AmbiguousMergeBaseData is a construction/copy value; NewAmbiguousMergeBase validates and owns a copy.
type AmbiguousMergeBaseData struct {
	Candidates []domain.Revision
}

func NewAmbiguousMergeBase(d AmbiguousMergeBaseData) (AmbiguousMergeBase, error) {
	if err := d.validate(); err != nil {
		return AmbiguousMergeBase{}, err
	}
	return AmbiguousMergeBase{data: d.clone(), initialized: true}, nil
}
func (v AmbiguousMergeBase) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v AmbiguousMergeBase) Data() AmbiguousMergeBaseData { return v.data.clone() }
func (v AmbiguousMergeBase) Clone() AmbiguousMergeBase    { return v }
func (d AmbiguousMergeBaseData) clone() AmbiguousMergeBaseData {
	d.Candidates = cloneSlice(d.Candidates)
	return d
}
func (d AmbiguousMergeBaseData) validate() error {
	for _, item := range d.Candidates {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if len(d.Candidates) < 2 {
		return invalid("ambiguous bases")
	}
	for _, r := range d.Candidates {
		if !sameLocal(r, d.Candidates[0]) {
			return invalid("base scope")
		}
	}
	return nil
}

// ResolveLocalRequest is an immutable semantic boundary value. Its zero is invalid.
type ResolveLocalRequest struct {
	data        ResolveLocalRequestData
	initialized bool
}

// ResolveLocalRequestData is a construction/copy value; NewResolveLocalRequest validates and owns a copy.
type ResolveLocalRequestData struct {
	Locator string
}

func NewResolveLocalRequest(d ResolveLocalRequestData) (ResolveLocalRequest, error) {
	if err := d.validate(); err != nil {
		return ResolveLocalRequest{}, err
	}
	return ResolveLocalRequest{data: d.clone(), initialized: true}, nil
}
func (v ResolveLocalRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ResolveLocalRequest) Data() ResolveLocalRequestData { return v.data.clone() }
func (v ResolveLocalRequest) Clone() ResolveLocalRequest    { return v }
func (d ResolveLocalRequestData) clone() ResolveLocalRequestData {

	return d
}
func (d ResolveLocalRequestData) validate() error {

	if !nonempty(d.Locator) {
		return invalid("local locator")
	}
	return nil
}

// ResolveLocalResult is an immutable semantic boundary value. Its zero is invalid.
type ResolveLocalResult struct {
	data        ResolveLocalResultData
	initialized bool
}

// ResolveLocalResultData is a construction/copy value; NewResolveLocalResult validates and owns a copy.
type ResolveLocalResultData struct {
	Repository  Optional[LocalRepositoryFacts]
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewResolveLocalResult(d ResolveLocalResultData) (ResolveLocalResult, error) {
	if err := d.validate(); err != nil {
		return ResolveLocalResult{}, err
	}
	return ResolveLocalResult{data: d.clone(), initialized: true}, nil
}
func (v ResolveLocalResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v ResolveLocalResult) Data() ResolveLocalResultData { return v.data.clone() }
func (v ResolveLocalResult) Clone() ResolveLocalResult    { return v }
func (d ResolveLocalResultData) clone() ResolveLocalResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ResolveLocalResultData) validate() error {
	if item, ok := d.Repository.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ListWorktreesRequest is an immutable semantic boundary value. Its zero is invalid.
type ListWorktreesRequest struct {
	data        ListWorktreesRequestData
	initialized bool
}

// ListWorktreesRequestData is a construction/copy value; NewListWorktreesRequest validates and owns a copy.
type ListWorktreesRequestData struct {
	Repository domain.RepositoryID
}

func NewListWorktreesRequest(d ListWorktreesRequestData) (ListWorktreesRequest, error) {
	if err := d.validate(); err != nil {
		return ListWorktreesRequest{}, err
	}
	return ListWorktreesRequest{data: d.clone(), initialized: true}, nil
}
func (v ListWorktreesRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ListWorktreesRequest) Data() ListWorktreesRequestData { return v.data.clone() }
func (v ListWorktreesRequest) Clone() ListWorktreesRequest    { return v }
func (d ListWorktreesRequestData) clone() ListWorktreesRequestData {

	return d
}
func (d ListWorktreesRequestData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("local scope")
	}
	return nil
}

// ListWorktreesResult is an immutable semantic boundary value. Its zero is invalid.
type ListWorktreesResult struct {
	data        ListWorktreesResultData
	initialized bool
}

// ListWorktreesResultData is a construction/copy value; NewListWorktreesResult validates and owns a copy.
type ListWorktreesResultData struct {
	Worktrees   []WorktreeFacts
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewListWorktreesResult(d ListWorktreesResultData) (ListWorktreesResult, error) {
	if err := d.validate(); err != nil {
		return ListWorktreesResult{}, err
	}
	return ListWorktreesResult{data: d.clone(), initialized: true}, nil
}
func (v ListWorktreesResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ListWorktreesResult) Data() ListWorktreesResultData { return v.data.clone() }
func (v ListWorktreesResult) Clone() ListWorktreesResult    { return v }
func (d ListWorktreesResultData) clone() ListWorktreesResultData {
	d.Worktrees = cloneSlice(d.Worktrees)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ListWorktreesResultData) validate() error {
	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ObserveStatusRequest is an immutable semantic boundary value. Its zero is invalid.
type ObserveStatusRequest struct {
	data        ObserveStatusRequestData
	initialized bool
}

// ObserveStatusRequestData is a construction/copy value; NewObserveStatusRequest validates and owns a copy.
type ObserveStatusRequestData struct {
	Worktree domain.WorktreeID
}

func NewObserveStatusRequest(d ObserveStatusRequestData) (ObserveStatusRequest, error) {
	if err := d.validate(); err != nil {
		return ObserveStatusRequest{}, err
	}
	return ObserveStatusRequest{data: d.clone(), initialized: true}, nil
}
func (v ObserveStatusRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ObserveStatusRequest) Data() ObserveStatusRequestData { return v.data.clone() }
func (v ObserveStatusRequest) Clone() ObserveStatusRequest    { return v }
func (d ObserveStatusRequestData) clone() ObserveStatusRequestData {

	return d
}
func (d ObserveStatusRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}

	return nil
}

// ObserveStatusResult is an immutable semantic boundary value. Its zero is invalid.
type ObserveStatusResult struct {
	data        ObserveStatusResultData
	initialized bool
}

// ObserveStatusResultData is a construction/copy value; NewObserveStatusResult validates and owns a copy.
type ObserveStatusResultData struct {
	Status      Optional[StatusFacts]
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewObserveStatusResult(d ObserveStatusResultData) (ObserveStatusResult, error) {
	if err := d.validate(); err != nil {
		return ObserveStatusResult{}, err
	}
	return ObserveStatusResult{data: d.clone(), initialized: true}, nil
}
func (v ObserveStatusResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ObserveStatusResult) Data() ObserveStatusResultData { return v.data.clone() }
func (v ObserveStatusResult) Clone() ObserveStatusResult    { return v }
func (d ObserveStatusResultData) clone() ObserveStatusResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ObserveStatusResultData) validate() error {
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ResolveExactRequest is an immutable semantic boundary value. Its zero is invalid.
type ResolveExactRequest struct {
	data        ResolveExactRequestData
	initialized bool
}

// ResolveExactRequestData is a construction/copy value; NewResolveExactRequest validates and owns a copy.
type ResolveExactRequestData struct {
	Repository     domain.RepositoryID
	Target         domain.ExactTarget
	Locator        Optional[GitRefLocator]
	ExpectedSource Optional[SourceVersion]
}

func NewResolveExactRequest(d ResolveExactRequestData) (ResolveExactRequest, error) {
	if err := d.validate(); err != nil {
		return ResolveExactRequest{}, err
	}
	return ResolveExactRequest{data: d.clone(), initialized: true}, nil
}
func (v ResolveExactRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ResolveExactRequest) Data() ResolveExactRequestData { return v.data.clone() }
func (v ResolveExactRequest) Clone() ResolveExactRequest    { return v }
func (d ResolveExactRequestData) clone() ResolveExactRequestData {

	return d
}
func (d ResolveExactRequestData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if item, ok := d.Locator.Value(); ok {
		if !validGitRefLocator(item) {
			return invalid("item")
		}
	}
	if item, ok := d.ExpectedSource.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("local scope")
	}
	return nil
}

// ResolveExactResult is an immutable semantic boundary value. Its zero is invalid.
type ResolveExactResult struct {
	data        ResolveExactResultData
	initialized bool
}

// ResolveExactResultData is a construction/copy value; NewResolveExactResult validates and owns a copy.
type ResolveExactResultData struct {
	Resolution  Optional[ExactLocalResolution]
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewResolveExactResult(d ResolveExactResultData) (ResolveExactResult, error) {
	if err := d.validate(); err != nil {
		return ResolveExactResult{}, err
	}
	return ResolveExactResult{data: d.clone(), initialized: true}, nil
}
func (v ResolveExactResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v ResolveExactResult) Data() ResolveExactResultData { return v.data.clone() }
func (v ResolveExactResult) Clone() ResolveExactResult    { return v }
func (d ResolveExactResultData) clone() ResolveExactResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ResolveExactResultData) validate() error {
	if item, ok := d.Resolution.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ListRefsRequest is an immutable semantic boundary value. Its zero is invalid.
type ListRefsRequest struct {
	data        ListRefsRequestData
	initialized bool
}

// ListRefsRequestData is a construction/copy value; NewListRefsRequest validates and owns a copy.
type ListRefsRequestData struct {
	Repository domain.RepositoryID
	Kinds      []RefKind
	Page       PageRequest
}

func NewListRefsRequest(d ListRefsRequestData) (ListRefsRequest, error) {
	if err := d.validate(); err != nil {
		return ListRefsRequest{}, err
	}
	return ListRefsRequest{data: d.clone(), initialized: true}, nil
}
func (v ListRefsRequest) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ListRefsRequest) Data() ListRefsRequestData { return v.data.clone() }
func (v ListRefsRequest) Clone() ListRefsRequest    { return v }
func (d ListRefsRequestData) clone() ListRefsRequestData {
	d.Kinds = cloneSlice(d.Kinds)
	return d
}
func (d ListRefsRequestData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	for _, item := range d.Kinds {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.LocalCommon || len(d.Kinds) == 0 {
		return invalid("ref scope")
	}
	return nil
}

// ListRefsResult is an immutable semantic boundary value. Its zero is invalid.
type ListRefsResult struct {
	data        ListRefsResultData
	initialized bool
}

// ListRefsResultData is a construction/copy value; NewListRefsResult validates and owns a copy.
type ListRefsResultData struct {
	Refs        []RefFact
	Page        PageInfo
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewListRefsResult(d ListRefsResultData) (ListRefsResult, error) {
	if err := d.validate(); err != nil {
		return ListRefsResult{}, err
	}
	return ListRefsResult{data: d.clone(), initialized: true}, nil
}
func (v ListRefsResult) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v ListRefsResult) Data() ListRefsResultData { return v.data.clone() }
func (v ListRefsResult) Clone() ListRefsResult    { return v }
func (d ListRefsResultData) clone() ListRefsResultData {
	d.Refs = cloneSlice(d.Refs)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ListRefsResultData) validate() error {
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ListStashesRequest is an immutable semantic boundary value. Its zero is invalid.
type ListStashesRequest struct {
	data        ListStashesRequestData
	initialized bool
}

// ListStashesRequestData is a construction/copy value; NewListStashesRequest validates and owns a copy.
type ListStashesRequestData struct {
	Repository domain.RepositoryID
	Page       PageRequest
}

func NewListStashesRequest(d ListStashesRequestData) (ListStashesRequest, error) {
	if err := d.validate(); err != nil {
		return ListStashesRequest{}, err
	}
	return ListStashesRequest{data: d.clone(), initialized: true}, nil
}
func (v ListStashesRequest) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v ListStashesRequest) Data() ListStashesRequestData { return v.data.clone() }
func (v ListStashesRequest) Clone() ListStashesRequest    { return v }
func (d ListStashesRequestData) clone() ListStashesRequestData {

	return d
}
func (d ListStashesRequestData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("stash scope")
	}
	return nil
}

// ListStashesResult is an immutable semantic boundary value. Its zero is invalid.
type ListStashesResult struct {
	data        ListStashesResultData
	initialized bool
}

// ListStashesResultData is a construction/copy value; NewListStashesResult validates and owns a copy.
type ListStashesResultData struct {
	Stashes     []StashFact
	Page        PageInfo
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewListStashesResult(d ListStashesResultData) (ListStashesResult, error) {
	if err := d.validate(); err != nil {
		return ListStashesResult{}, err
	}
	return ListStashesResult{data: d.clone(), initialized: true}, nil
}
func (v ListStashesResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v ListStashesResult) Data() ListStashesResultData { return v.data.clone() }
func (v ListStashesResult) Clone() ListStashesResult    { return v }
func (d ListStashesResultData) clone() ListStashesResultData {
	d.Stashes = cloneSlice(d.Stashes)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ListStashesResultData) validate() error {
	for _, item := range d.Stashes {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ReadCommitsRequest is an immutable semantic boundary value. Its zero is invalid.
type ReadCommitsRequest struct {
	data        ReadCommitsRequestData
	initialized bool
}

// ReadCommitsRequestData is a construction/copy value; NewReadCommitsRequest validates and owns a copy.
type ReadCommitsRequestData struct {
	Endpoint  domain.Revision
	Traversal CommitTraversal
	Page      PageRequest
}

func NewReadCommitsRequest(d ReadCommitsRequestData) (ReadCommitsRequest, error) {
	if err := d.validate(); err != nil {
		return ReadCommitsRequest{}, err
	}
	return ReadCommitsRequest{data: d.clone(), initialized: true}, nil
}
func (v ReadCommitsRequest) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v ReadCommitsRequest) Data() ReadCommitsRequestData { return v.data.clone() }
func (v ReadCommitsRequest) Clone() ReadCommitsRequest    { return v }
func (d ReadCommitsRequestData) clone() ReadCommitsRequestData {

	return d
}
func (d ReadCommitsRequestData) validate() error {
	if !d.Endpoint.Valid() {
		return invalid("d.Endpoint")
	}
	if !d.Traversal.Valid() {
		return invalid("d.Traversal")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Endpoint.Repository().Scope() != domain.LocalCommon {
		return invalid("history scope")
	}
	return nil
}

// ReadCommitsResult is an immutable semantic boundary value. Its zero is invalid.
type ReadCommitsResult struct {
	data        ReadCommitsResultData
	initialized bool
}

// ReadCommitsResultData is a construction/copy value; NewReadCommitsResult validates and owns a copy.
type ReadCommitsResultData struct {
	Commits     []CommitFact
	Page        PageInfo
	Endpoint    domain.Revision
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewReadCommitsResult(d ReadCommitsResultData) (ReadCommitsResult, error) {
	if err := d.validate(); err != nil {
		return ReadCommitsResult{}, err
	}
	return ReadCommitsResult{data: d.clone(), initialized: true}, nil
}
func (v ReadCommitsResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v ReadCommitsResult) Data() ReadCommitsResultData { return v.data.clone() }
func (v ReadCommitsResult) Clone() ReadCommitsResult    { return v }
func (d ReadCommitsResultData) clone() ReadCommitsResultData {
	d.Commits = cloneSlice(d.Commits)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ReadCommitsResultData) validate() error {
	for _, item := range d.Commits {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if !d.Endpoint.Valid() {
		return invalid("d.Endpoint")
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ReadGraphRequest is an immutable semantic boundary value. Its zero is invalid.
type ReadGraphRequest struct {
	data        ReadGraphRequestData
	initialized bool
}

// ReadGraphRequestData is a construction/copy value; NewReadGraphRequest validates and owns a copy.
type ReadGraphRequestData struct {
	Repository domain.RepositoryID
	Roots      []domain.Revision
	Filter     GraphFilter
	Page       PageRequest
}

func NewReadGraphRequest(d ReadGraphRequestData) (ReadGraphRequest, error) {
	if err := d.validate(); err != nil {
		return ReadGraphRequest{}, err
	}
	return ReadGraphRequest{data: d.clone(), initialized: true}, nil
}
func (v ReadGraphRequest) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v ReadGraphRequest) Data() ReadGraphRequestData { return v.data.clone() }
func (v ReadGraphRequest) Clone() ReadGraphRequest    { return v }
func (d ReadGraphRequestData) clone() ReadGraphRequestData {
	d.Roots = cloneSlice(d.Roots)
	return d
}
func (d ReadGraphRequestData) validate() error {
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
		return invalid("graph scope")
	}
	for _, r := range d.Roots {
		if r.Repository() != d.Repository {
			return invalid("graph root scope")
		}
	}
	return nil
}

// ReadGraphResult is an immutable semantic boundary value. Its zero is invalid.
type ReadGraphResult struct {
	data        ReadGraphResultData
	initialized bool
}

// ReadGraphResultData is a construction/copy value; NewReadGraphResult validates and owns a copy.
type ReadGraphResultData struct {
	Commits     []CommitFact
	Refs        []RefFact
	Heads       []WorktreeFacts
	Page        PageInfo
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewReadGraphResult(d ReadGraphResultData) (ReadGraphResult, error) {
	if err := d.validate(); err != nil {
		return ReadGraphResult{}, err
	}
	return ReadGraphResult{data: d.clone(), initialized: true}, nil
}
func (v ReadGraphResult) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ReadGraphResult) Data() ReadGraphResultData { return v.data.clone() }
func (v ReadGraphResult) Clone() ReadGraphResult    { return v }
func (d ReadGraphResultData) clone() ReadGraphResultData {
	d.Commits = cloneSlice(d.Commits)
	d.Refs = cloneSlice(d.Refs)
	d.Heads = cloneSlice(d.Heads)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ReadGraphResultData) validate() error {
	for _, item := range d.Commits {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Heads {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// MergeBaseRequest is an immutable semantic boundary value. Its zero is invalid.
type MergeBaseRequest struct {
	data        MergeBaseRequestData
	initialized bool
}

// MergeBaseRequestData is a construction/copy value; NewMergeBaseRequest validates and owns a copy.
type MergeBaseRequestData struct {
	Left  domain.Revision
	Right domain.Revision
}

func NewMergeBaseRequest(d MergeBaseRequestData) (MergeBaseRequest, error) {
	if err := d.validate(); err != nil {
		return MergeBaseRequest{}, err
	}
	return MergeBaseRequest{data: d.clone(), initialized: true}, nil
}
func (v MergeBaseRequest) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v MergeBaseRequest) Data() MergeBaseRequestData { return v.data.clone() }
func (v MergeBaseRequest) Clone() MergeBaseRequest    { return v }
func (d MergeBaseRequestData) clone() MergeBaseRequestData {

	return d
}
func (d MergeBaseRequestData) validate() error {
	if !d.Left.Valid() {
		return invalid("d.Left")
	}
	if !d.Right.Valid() {
		return invalid("d.Right")
	}
	if !sameLocal(d.Left, d.Right) {
		return invalid("merge base scope")
	}
	return nil
}

// MergeBaseResult is an immutable semantic boundary value. Its zero is invalid.
type MergeBaseResult struct {
	data        MergeBaseResultData
	initialized bool
}

// MergeBaseResultData is a construction/copy value; NewMergeBaseResult validates and owns a copy.
type MergeBaseResultData struct {
	Left        domain.Revision
	Right       domain.Revision
	Outcome     Optional[MergeBaseOutcome]
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewMergeBaseResult(d MergeBaseResultData) (MergeBaseResult, error) {
	if err := d.validate(); err != nil {
		return MergeBaseResult{}, err
	}
	return MergeBaseResult{data: d.clone(), initialized: true}, nil
}
func (v MergeBaseResult) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v MergeBaseResult) Data() MergeBaseResultData { return v.data.clone() }
func (v MergeBaseResult) Clone() MergeBaseResult    { return v }
func (d MergeBaseResultData) clone() MergeBaseResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d MergeBaseResultData) validate() error {
	if !d.Left.Valid() {
		return invalid("d.Left")
	}
	if !d.Right.Valid() {
		return invalid("d.Right")
	}
	if item, ok := d.Outcome.Value(); ok {
		if !validMergeBaseOutcome(item) {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ReadDiffRequest is an immutable semantic boundary value. Its zero is invalid.
type ReadDiffRequest struct {
	data        ReadDiffRequestData
	initialized bool
}

// ReadDiffRequestData is a construction/copy value; NewReadDiffRequest validates and owns a copy.
type ReadDiffRequestData struct {
	Comparison GitComparison
	Paths      []GitPath
	Limits     PatchLimits
}

func NewReadDiffRequest(d ReadDiffRequestData) (ReadDiffRequest, error) {
	if err := d.validate(); err != nil {
		return ReadDiffRequest{}, err
	}
	return ReadDiffRequest{data: d.clone(), initialized: true}, nil
}
func (v ReadDiffRequest) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ReadDiffRequest) Data() ReadDiffRequestData { return v.data.clone() }
func (v ReadDiffRequest) Clone() ReadDiffRequest    { return v }
func (d ReadDiffRequestData) clone() ReadDiffRequestData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d ReadDiffRequestData) validate() error {
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

// ReadDiffResult is an immutable semantic boundary value. Its zero is invalid.
type ReadDiffResult struct {
	data        ReadDiffResultData
	initialized bool
}

// ReadDiffResultData is a construction/copy value; NewReadDiffResult validates and owns a copy.
type ReadDiffResultData struct {
	Comparison  GitComparison
	Patch       Optional[PatchFacts]
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewReadDiffResult(d ReadDiffResultData) (ReadDiffResult, error) {
	if err := d.validate(); err != nil {
		return ReadDiffResult{}, err
	}
	return ReadDiffResult{data: d.clone(), initialized: true}, nil
}
func (v ReadDiffResult) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v ReadDiffResult) Data() ReadDiffResultData { return v.data.clone() }
func (v ReadDiffResult) Clone() ReadDiffResult    { return v }
func (d ReadDiffResultData) clone() ReadDiffResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ReadDiffResultData) validate() error {
	if !validGitComparison(d.Comparison) {
		return invalid("d.Comparison")
	}
	if item, ok := d.Patch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// ReadStashPatchRequest is an immutable semantic boundary value. Its zero is invalid.
type ReadStashPatchRequest struct {
	data        ReadStashPatchRequestData
	initialized bool
}

// ReadStashPatchRequestData is a construction/copy value; NewReadStashPatchRequest validates and owns a copy.
type ReadStashPatchRequestData struct {
	Stash  domain.StashID
	View   StashPatchView
	Paths  []GitPath
	Limits PatchLimits
}

func NewReadStashPatchRequest(d ReadStashPatchRequestData) (ReadStashPatchRequest, error) {
	if err := d.validate(); err != nil {
		return ReadStashPatchRequest{}, err
	}
	return ReadStashPatchRequest{data: d.clone(), initialized: true}, nil
}
func (v ReadStashPatchRequest) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v ReadStashPatchRequest) Data() ReadStashPatchRequestData { return v.data.clone() }
func (v ReadStashPatchRequest) Clone() ReadStashPatchRequest    { return v }
func (d ReadStashPatchRequestData) clone() ReadStashPatchRequestData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d ReadStashPatchRequestData) validate() error {
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

// ReadStashPatchResult is an immutable semantic boundary value. Its zero is invalid.
type ReadStashPatchResult struct {
	data        ReadStashPatchResultData
	initialized bool
}

// ReadStashPatchResultData is a construction/copy value; NewReadStashPatchResult validates and owns a copy.
type ReadStashPatchResultData struct {
	Stash       domain.StashID
	Parents     []domain.OID
	Comparison  Optional[StashPatchComparison]
	Patch       Optional[PatchFacts]
	Observation Optional[GitObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewReadStashPatchResult(d ReadStashPatchResultData) (ReadStashPatchResult, error) {
	if err := d.validate(); err != nil {
		return ReadStashPatchResult{}, err
	}
	return ReadStashPatchResult{data: d.clone(), initialized: true}, nil
}
func (v ReadStashPatchResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v ReadStashPatchResult) Data() ReadStashPatchResultData { return v.data.clone() }
func (v ReadStashPatchResult) Clone() ReadStashPatchResult    { return v }
func (d ReadStashPatchResultData) clone() ReadStashPatchResultData {
	d.Parents = cloneSlice(d.Parents)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ReadStashPatchResultData) validate() error {
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
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}

	return nil
}

// GitExpectedState is an immutable semantic boundary value. Its zero is invalid.
type GitExpectedState struct {
	data        GitExpectedStateData
	initialized bool
}

// GitExpectedStateData is a construction/copy value; NewGitExpectedState validates and owns a copy.
type GitExpectedStateData struct {
	Repository    domain.RepositoryID
	Worktree      Optional[domain.WorktreeID]
	Observation   SourceVersion
	Head          Optional[domain.Head]
	Index         Optional[SourceVersion]
	WorktreeState Optional[SourceVersion]
	Configuration SourceVersion
	Inventory     SourceVersion
}

func NewGitExpectedState(d GitExpectedStateData) (GitExpectedState, error) {
	if err := d.validate(); err != nil {
		return GitExpectedState{}, err
	}
	return GitExpectedState{data: d.clone(), initialized: true}, nil
}
func (v GitExpectedState) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v GitExpectedState) Data() GitExpectedStateData { return v.data.clone() }
func (v GitExpectedState) Clone() GitExpectedState    { return v }
func (d GitExpectedStateData) clone() GitExpectedStateData {

	return d
}
func (d GitExpectedStateData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Index.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.WorktreeState.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Configuration.Valid() {
		return invalid("d.Configuration")
	}
	if !d.Inventory.Valid() {
		return invalid("d.Inventory")
	}
	if d.Repository.Scope() != domain.LocalCommon {
		return invalid("expected local scope")
	}
	if w, ok := d.Worktree.Value(); ok && w.Repository() != d.Repository {
		return invalid("expected worktree")
	}
	if h, ok := d.Head.Value(); ok && h.Repository() != d.Repository {
		return invalid("expected head")
	}
	return nil
}

// DetachedCreate is an immutable semantic boundary value. Its zero is invalid.
type DetachedCreate struct {
	data        DetachedCreateData
	initialized bool
}

// DetachedCreateData is a construction/copy value; NewDetachedCreate validates and owns a copy.
type DetachedCreateData struct {
}

func NewDetachedCreate(d DetachedCreateData) (DetachedCreate, error) {
	if err := d.validate(); err != nil {
		return DetachedCreate{}, err
	}
	return DetachedCreate{data: d.clone(), initialized: true}, nil
}
func (v DetachedCreate) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v DetachedCreate) Data() DetachedCreateData { return v.data.clone() }
func (v DetachedCreate) Clone() DetachedCreate    { return v }
func (d DetachedCreateData) clone() DetachedCreateData {

	return d
}
func (d DetachedCreateData) validate() error {

	return nil
}

// CreateNewBranch is an immutable semantic boundary value. Its zero is invalid.
type CreateNewBranch struct {
	data        CreateNewBranchData
	initialized bool
}

// CreateNewBranchData is a construction/copy value; NewCreateNewBranch validates and owns a copy.
type CreateNewBranchData struct {
	Branch domain.BranchID
}

func NewCreateNewBranch(d CreateNewBranchData) (CreateNewBranch, error) {
	if err := d.validate(); err != nil {
		return CreateNewBranch{}, err
	}
	return CreateNewBranch{data: d.clone(), initialized: true}, nil
}
func (v CreateNewBranch) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v CreateNewBranch) Data() CreateNewBranchData { return v.data.clone() }
func (v CreateNewBranch) Clone() CreateNewBranch    { return v }
func (d CreateNewBranchData) clone() CreateNewBranchData {

	return d
}
func (d CreateNewBranchData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if d.Branch.Kind() != domain.Local {
		return invalid("new branch scope")
	}
	return nil
}

// DetachRetarget is an immutable semantic boundary value. Its zero is invalid.
type DetachRetarget struct {
	data        DetachRetargetData
	initialized bool
}

// DetachRetargetData is a construction/copy value; NewDetachRetarget validates and owns a copy.
type DetachRetargetData struct {
}

func NewDetachRetarget(d DetachRetargetData) (DetachRetarget, error) {
	if err := d.validate(); err != nil {
		return DetachRetarget{}, err
	}
	return DetachRetarget{data: d.clone(), initialized: true}, nil
}
func (v DetachRetarget) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v DetachRetarget) Data() DetachRetargetData { return v.data.clone() }
func (v DetachRetarget) Clone() DetachRetarget    { return v }
func (d DetachRetargetData) clone() DetachRetargetData {

	return d
}
func (d DetachRetargetData) validate() error {

	return nil
}

// AttachExisting is an immutable semantic boundary value. Its zero is invalid.
type AttachExisting struct {
	data        AttachExistingData
	initialized bool
}

// AttachExistingData is a construction/copy value; NewAttachExisting validates and owns a copy.
type AttachExistingData struct {
	Branch domain.BranchID
}

func NewAttachExisting(d AttachExistingData) (AttachExisting, error) {
	if err := d.validate(); err != nil {
		return AttachExisting{}, err
	}
	return AttachExisting{data: d.clone(), initialized: true}, nil
}
func (v AttachExisting) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v AttachExisting) Data() AttachExistingData { return v.data.clone() }
func (v AttachExisting) Clone() AttachExisting    { return v }
func (d AttachExistingData) clone() AttachExistingData {

	return d
}
func (d AttachExistingData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if d.Branch.Kind() != domain.Local {
		return invalid("attachment scope")
	}
	return nil
}

// FastForward is an immutable semantic boundary value. Its zero is invalid.
type FastForward struct {
	data        FastForwardData
	initialized bool
}

// FastForwardData is a construction/copy value; NewFastForward validates and owns a copy.
type FastForwardData struct {
	Branch domain.BranchID
	From   domain.Revision
	To     domain.Revision
}

func NewFastForward(d FastForwardData) (FastForward, error) {
	if err := d.validate(); err != nil {
		return FastForward{}, err
	}
	return FastForward{data: d.clone(), initialized: true}, nil
}
func (v FastForward) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v FastForward) Data() FastForwardData { return v.data.clone() }
func (v FastForward) Clone() FastForward    { return v }
func (d FastForwardData) clone() FastForwardData {

	return d
}
func (d FastForwardData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.From.Valid() {
		return invalid("d.From")
	}
	if !d.To.Valid() {
		return invalid("d.To")
	}
	if !sameLocal(d.From, d.To) || d.Branch.Repository() != d.From.Repository() {
		return invalid("fast forward scope")
	}
	return nil
}

// ExactPaths is an immutable semantic boundary value. Its zero is invalid.
type ExactPaths struct {
	data        ExactPathsData
	initialized bool
}

// ExactPathsData is a construction/copy value; NewExactPaths validates and owns a copy.
type ExactPathsData struct {
	Paths []GitPath
}

func NewExactPaths(d ExactPathsData) (ExactPaths, error) {
	if err := d.validate(); err != nil {
		return ExactPaths{}, err
	}
	return ExactPaths{data: d.clone(), initialized: true}, nil
}
func (v ExactPaths) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v ExactPaths) Data() ExactPathsData { return v.data.clone() }
func (v ExactPaths) Clone() ExactPaths    { return v }
func (d ExactPathsData) clone() ExactPathsData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d ExactPathsData) validate() error {
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if len(d.Paths) == 0 || duplicatePaths(d.Paths) {
		return invalid("exact paths")
	}
	return nil
}

// AllObserved is an immutable semantic boundary value. Its zero is invalid.
type AllObserved struct {
	data        AllObservedData
	initialized bool
}

// AllObservedData is a construction/copy value; NewAllObserved validates and owns a copy.
type AllObservedData struct {
}

func NewAllObserved(d AllObservedData) (AllObserved, error) {
	if err := d.validate(); err != nil {
		return AllObserved{}, err
	}
	return AllObserved{data: d.clone(), initialized: true}, nil
}
func (v AllObserved) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v AllObserved) Data() AllObservedData { return v.data.clone() }
func (v AllObserved) Clone() AllObserved    { return v }
func (d AllObservedData) clone() AllObservedData {

	return d
}
func (d AllObservedData) validate() error {

	return nil
}

// PrepareCreateRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareCreateRequest struct {
	data        PrepareCreateRequestData
	initialized bool
}

// PrepareCreateRequestData is a construction/copy value; NewPrepareCreateRequest validates and owns a copy.
type PrepareCreateRequestData struct {
	Destination string
	Target      ExactLocalResolution
	Mode        CreateMode
	Expected    GitExpectedState
}

func NewPrepareCreateRequest(d PrepareCreateRequestData) (PrepareCreateRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareCreateRequest{}, err
	}
	return PrepareCreateRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareCreateRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v PrepareCreateRequest) Data() PrepareCreateRequestData { return v.data.clone() }
func (v PrepareCreateRequest) Clone() PrepareCreateRequest    { return v }
func (d PrepareCreateRequestData) clone() PrepareCreateRequestData {

	return d
}
func (d PrepareCreateRequestData) validate() error {

	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !validCreateMode(d.Mode) {
		return invalid("d.Mode")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !nonempty(d.Destination) || d.Target.data.Local.Repository() != d.Expected.data.Repository || !createModeScope(d.Mode, d.Expected.data.Repository) {
		return invalid("create destination/target")
	}
	return nil
}

// PrepareRetargetRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareRetargetRequest struct {
	data        PrepareRetargetRequestData
	initialized bool
}

// PrepareRetargetRequestData is a construction/copy value; NewPrepareRetargetRequest validates and owns a copy.
type PrepareRetargetRequestData struct {
	Worktree    domain.WorktreeID
	Target      ExactLocalResolution
	Mode        RetargetMode
	Purpose     RetargetPurpose
	DirtyPolicy DirtyPolicy
	Expected    GitExpectedState
}

func NewPrepareRetargetRequest(d PrepareRetargetRequestData) (PrepareRetargetRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareRetargetRequest{}, err
	}
	return PrepareRetargetRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareRetargetRequest) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v PrepareRetargetRequest) Data() PrepareRetargetRequestData { return v.data.clone() }
func (v PrepareRetargetRequest) Clone() PrepareRetargetRequest    { return v }
func (d PrepareRetargetRequestData) clone() PrepareRetargetRequestData {

	return d
}
func (d PrepareRetargetRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !validRetargetMode(d.Mode) {
		return invalid("d.Mode")
	}
	if !d.Purpose.Valid() {
		return invalid("d.Purpose")
	}
	if !d.DirtyPolicy.Valid() {
		return invalid("d.DirtyPolicy")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !expectedWorktree(d.Expected, d.Worktree, true) || d.Target.data.Local.Repository() != d.Worktree.Repository() || !retargetModeScope(d.Mode, d.Worktree.Repository(), Some(d.Target.data.Local)) {
		return invalid("retarget expectation")
	}
	return nil
}

// PrepareStageRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareStageRequest struct {
	data        PrepareStageRequestData
	initialized bool
}

// PrepareStageRequestData is a construction/copy value; NewPrepareStageRequest validates and owns a copy.
type PrepareStageRequestData struct {
	Worktree  domain.WorktreeID
	Selection PathSelection
	Action    StageAction
	Expected  GitExpectedState
}

func NewPrepareStageRequest(d PrepareStageRequestData) (PrepareStageRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareStageRequest{}, err
	}
	return PrepareStageRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareStageRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v PrepareStageRequest) Data() PrepareStageRequestData { return v.data.clone() }
func (v PrepareStageRequest) Clone() PrepareStageRequest    { return v }
func (d PrepareStageRequestData) clone() PrepareStageRequestData {

	return d
}
func (d PrepareStageRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !validPathSelection(d.Selection) {
		return invalid("d.Selection")
	}
	if !d.Action.Valid() {
		return invalid("d.Action")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !expectedWorktree(d.Expected, d.Worktree, true) {
		return invalid("stage expectation")
	}
	return nil
}

// PrepareCommitRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareCommitRequest struct {
	data        PrepareCommitRequestData
	initialized bool
}

// PrepareCommitRequestData is a construction/copy value; NewPrepareCommitRequest validates and owns a copy.
type PrepareCommitRequestData struct {
	Worktree    domain.WorktreeID
	Message     string
	IndexPolicy CommitIndexPolicy
	Expected    GitExpectedState
}

func NewPrepareCommitRequest(d PrepareCommitRequestData) (PrepareCommitRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareCommitRequest{}, err
	}
	return PrepareCommitRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareCommitRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v PrepareCommitRequest) Data() PrepareCommitRequestData { return v.data.clone() }
func (v PrepareCommitRequest) Clone() PrepareCommitRequest    { return v }
func (d PrepareCommitRequestData) clone() PrepareCommitRequestData {

	return d
}
func (d PrepareCommitRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}

	if !d.IndexPolicy.Valid() {
		return invalid("d.IndexPolicy")
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !expectedWorktree(d.Expected, d.Worktree, true) || !nonempty(d.Message) {
		return invalid("commit expectation/message")
	}
	return nil
}

// PrepareRestoreRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareRestoreRequest struct {
	data        PrepareRestoreRequestData
	initialized bool
}

// PrepareRestoreRequestData is a construction/copy value; NewPrepareRestoreRequest validates and owns a copy.
type PrepareRestoreRequestData struct {
	Worktree domain.WorktreeID
	Paths    []GitPath
	Expected GitExpectedState
}

func NewPrepareRestoreRequest(d PrepareRestoreRequestData) (PrepareRestoreRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareRestoreRequest{}, err
	}
	return PrepareRestoreRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareRestoreRequest) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v PrepareRestoreRequest) Data() PrepareRestoreRequestData { return v.data.clone() }
func (v PrepareRestoreRequest) Clone() PrepareRestoreRequest    { return v }
func (d PrepareRestoreRequestData) clone() PrepareRestoreRequestData {
	d.Paths = cloneSlice(d.Paths)
	return d
}
func (d PrepareRestoreRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !expectedWorktree(d.Expected, d.Worktree, true) || len(d.Paths) == 0 || duplicatePaths(d.Paths) {
		return invalid("restore expectation")
	}
	return nil
}

// PrepareBranchRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareBranchRequest struct {
	data        PrepareBranchRequestData
	initialized bool
}

// PrepareBranchRequestData is a construction/copy value; NewPrepareBranchRequest validates and owns a copy.
type PrepareBranchRequestData struct {
	Worktree domain.WorktreeID
	Name     domain.BranchID
	Start    domain.Revision
	Checkout bool
	Expected GitExpectedState
}

func NewPrepareBranchRequest(d PrepareBranchRequestData) (PrepareBranchRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareBranchRequest{}, err
	}
	return PrepareBranchRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareBranchRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v PrepareBranchRequest) Data() PrepareBranchRequestData { return v.data.clone() }
func (v PrepareBranchRequest) Clone() PrepareBranchRequest    { return v }
func (d PrepareBranchRequestData) clone() PrepareBranchRequestData {

	return d
}
func (d PrepareBranchRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
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
	if !expectedWorktree(d.Expected, d.Worktree, false) || d.Name.Repository() != d.Worktree.Repository() || d.Start.Repository() != d.Worktree.Repository() {
		return invalid("branch expectation")
	}
	return nil
}

// PreparePushRequest is an immutable semantic boundary value. Its zero is invalid.
type PreparePushRequest struct {
	data        PreparePushRequestData
	initialized bool
}

// PreparePushRequestData is a construction/copy value; NewPreparePushRequest validates and owns a copy.
type PreparePushRequestData struct {
	Worktree    domain.WorktreeID
	Source      domain.Revision
	Destination domain.BranchID
	Binding     RemoteBinding
	SetUpstream Optional[UpstreamSetup]
	Expected    GitExpectedState
}

func NewPreparePushRequest(d PreparePushRequestData) (PreparePushRequest, error) {
	if err := d.validate(); err != nil {
		return PreparePushRequest{}, err
	}
	return PreparePushRequest{data: d.clone(), initialized: true}, nil
}
func (v PreparePushRequest) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v PreparePushRequest) Data() PreparePushRequestData { return v.data.clone() }
func (v PreparePushRequest) Clone() PreparePushRequest    { return v }
func (d PreparePushRequestData) clone() PreparePushRequestData {

	return d
}
func (d PreparePushRequestData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
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
	if !expectedWorktree(d.Expected, d.Worktree, false) || d.Source.Repository() != d.Worktree.Repository() || d.Binding.data.LocalRepository != d.Worktree.Repository() || d.Destination.Repository() != d.Binding.data.RemoteRepository {
		return invalid("push expectation")
	}
	return nil
}

// UpstreamSetup is an immutable semantic boundary value. Its zero is invalid.
type UpstreamSetup struct {
	data        UpstreamSetupData
	initialized bool
}

// UpstreamSetupData is a construction/copy value; NewUpstreamSetup validates and owns a copy.
type UpstreamSetupData struct {
	Branch                domain.BranchID
	ExpectedConfiguration SourceVersion
	ExpectedValue         UpstreamFact
}

func NewUpstreamSetup(d UpstreamSetupData) (UpstreamSetup, error) {
	if err := d.validate(); err != nil {
		return UpstreamSetup{}, err
	}
	return UpstreamSetup{data: d.clone(), initialized: true}, nil
}
func (v UpstreamSetup) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v UpstreamSetup) Data() UpstreamSetupData { return v.data.clone() }
func (v UpstreamSetup) Clone() UpstreamSetup    { return v }
func (d UpstreamSetupData) clone() UpstreamSetupData {

	return d
}
func (d UpstreamSetupData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.ExpectedConfiguration.Valid() {
		return invalid("d.ExpectedConfiguration")
	}
	if !validUpstreamFact(d.ExpectedValue) {
		return invalid("d.ExpectedValue")
	}
	if d.Branch.Kind() != domain.Local {
		return invalid("upstream local branch")
	}
	return nil
}

// CreateStashIntent is an immutable semantic boundary value. Its zero is invalid.
type CreateStashIntent struct {
	data        CreateStashIntentData
	initialized bool
}

// CreateStashIntentData is a construction/copy value; NewCreateStashIntent validates and owns a copy.
type CreateStashIntentData struct {
	Worktree         domain.WorktreeID
	Message          string
	IncludeUntracked bool
}

func NewCreateStashIntent(d CreateStashIntentData) (CreateStashIntent, error) {
	if err := d.validate(); err != nil {
		return CreateStashIntent{}, err
	}
	return CreateStashIntent{data: d.clone(), initialized: true}, nil
}
func (v CreateStashIntent) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v CreateStashIntent) Data() CreateStashIntentData { return v.data.clone() }
func (v CreateStashIntent) Clone() CreateStashIntent    { return v }
func (d CreateStashIntentData) clone() CreateStashIntentData {

	return d
}
func (d CreateStashIntentData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}

	return nil
}

// ApplyStashIntent is an immutable semantic boundary value. Its zero is invalid.
type ApplyStashIntent struct {
	data        ApplyStashIntentData
	initialized bool
}

// ApplyStashIntentData is a construction/copy value; NewApplyStashIntent validates and owns a copy.
type ApplyStashIntentData struct {
	Worktree   domain.WorktreeID
	Stash      domain.StashID
	Occurrence SourceVersion
}

func NewApplyStashIntent(d ApplyStashIntentData) (ApplyStashIntent, error) {
	if err := d.validate(); err != nil {
		return ApplyStashIntent{}, err
	}
	return ApplyStashIntent{data: d.clone(), initialized: true}, nil
}
func (v ApplyStashIntent) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v ApplyStashIntent) Data() ApplyStashIntentData { return v.data.clone() }
func (v ApplyStashIntent) Clone() ApplyStashIntent    { return v }
func (d ApplyStashIntentData) clone() ApplyStashIntentData {

	return d
}
func (d ApplyStashIntentData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}
	if d.Worktree.Repository() != d.Stash.Repository() {
		return invalid("stash scope")
	}
	return nil
}

// PopStashIntent is an immutable semantic boundary value. Its zero is invalid.
type PopStashIntent struct {
	data        PopStashIntentData
	initialized bool
}

// PopStashIntentData is a construction/copy value; NewPopStashIntent validates and owns a copy.
type PopStashIntentData struct {
	Worktree   domain.WorktreeID
	Stash      domain.StashID
	Occurrence SourceVersion
}

func NewPopStashIntent(d PopStashIntentData) (PopStashIntent, error) {
	if err := d.validate(); err != nil {
		return PopStashIntent{}, err
	}
	return PopStashIntent{data: d.clone(), initialized: true}, nil
}
func (v PopStashIntent) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v PopStashIntent) Data() PopStashIntentData { return v.data.clone() }
func (v PopStashIntent) Clone() PopStashIntent    { return v }
func (d PopStashIntentData) clone() PopStashIntentData {

	return d
}
func (d PopStashIntentData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}
	if d.Worktree.Repository() != d.Stash.Repository() {
		return invalid("stash scope")
	}
	return nil
}

// DropStashIntent is an immutable semantic boundary value. Its zero is invalid.
type DropStashIntent struct {
	data        DropStashIntentData
	initialized bool
}

// DropStashIntentData is a construction/copy value; NewDropStashIntent validates and owns a copy.
type DropStashIntentData struct {
	Stash      domain.StashID
	Occurrence SourceVersion
}

func NewDropStashIntent(d DropStashIntentData) (DropStashIntent, error) {
	if err := d.validate(); err != nil {
		return DropStashIntent{}, err
	}
	return DropStashIntent{data: d.clone(), initialized: true}, nil
}
func (v DropStashIntent) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v DropStashIntent) Data() DropStashIntentData { return v.data.clone() }
func (v DropStashIntent) Clone() DropStashIntent    { return v }
func (d DropStashIntentData) clone() DropStashIntentData {

	return d
}
func (d DropStashIntentData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}

	return nil
}

// PrepareStashRequest is an immutable semantic boundary value. Its zero is invalid.
type PrepareStashRequest struct {
	data        PrepareStashRequestData
	initialized bool
}

// PrepareStashRequestData is a construction/copy value; NewPrepareStashRequest validates and owns a copy.
type PrepareStashRequestData struct {
	Expected GitExpectedState
	Intent   StashIntent
}

func NewPrepareStashRequest(d PrepareStashRequestData) (PrepareStashRequest, error) {
	if err := d.validate(); err != nil {
		return PrepareStashRequest{}, err
	}
	return PrepareStashRequest{data: d.clone(), initialized: true}, nil
}
func (v PrepareStashRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v PrepareStashRequest) Data() PrepareStashRequestData { return v.data.clone() }
func (v PrepareStashRequest) Clone() PrepareStashRequest    { return v }
func (d PrepareStashRequestData) clone() PrepareStashRequestData {

	return d
}
func (d PrepareStashRequestData) validate() error {
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	if !validStashIntent(d.Intent) {
		return invalid("d.Intent")
	}
	return validateStashPrepare(d)
}

// PlannedPathEffect is an immutable semantic boundary value. Its zero is invalid.
type PlannedPathEffect struct {
	data        PlannedPathEffectData
	initialized bool
}

// PlannedPathEffectData is a construction/copy value; NewPlannedPathEffect validates and owns a copy.
type PlannedPathEffectData struct {
	Old           FileState
	Desired       FileState
	SourceBlob    Optional[domain.OID]
	SourceVersion SourceVersion
}

func NewPlannedPathEffect(d PlannedPathEffectData) (PlannedPathEffect, error) {
	if err := d.validate(); err != nil {
		return PlannedPathEffect{}, err
	}
	return PlannedPathEffect{data: d.clone(), initialized: true}, nil
}
func (v PlannedPathEffect) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v PlannedPathEffect) Data() PlannedPathEffectData { return v.data.clone() }
func (v PlannedPathEffect) Clone() PlannedPathEffect    { return v }
func (d PlannedPathEffectData) clone() PlannedPathEffectData {

	return d
}
func (d PlannedPathEffectData) validate() error {
	if !validFileState(d.Old) {
		return invalid("d.Old")
	}
	if !validFileState(d.Desired) {
		return invalid("d.Desired")
	}
	if item, ok := d.SourceBlob.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.SourceVersion.Valid() {
		return invalid("d.SourceVersion")
	}
	if filePath(d.Old) != filePath(d.Desired) {
		return invalid("planned path")
	}
	return nil
}

// PlannedRefEffect is an immutable semantic boundary value. Its zero is invalid.
type PlannedRefEffect struct {
	data        PlannedRefEffectData
	initialized bool
}

// PlannedRefEffectData is a construction/copy value; NewPlannedRefEffect validates and owns a copy.
type PlannedRefEffectData struct {
	Ref    GitRefLocator
	Before Optional[domain.Revision]
	After  Optional[domain.Revision]
}

func NewPlannedRefEffect(d PlannedRefEffectData) (PlannedRefEffect, error) {
	if err := d.validate(); err != nil {
		return PlannedRefEffect{}, err
	}
	return PlannedRefEffect{data: d.clone(), initialized: true}, nil
}
func (v PlannedRefEffect) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v PlannedRefEffect) Data() PlannedRefEffectData { return v.data.clone() }
func (v PlannedRefEffect) Clone() PlannedRefEffect    { return v }
func (d PlannedRefEffectData) clone() PlannedRefEffectData {

	return d
}
func (d PlannedRefEffectData) validate() error {
	if !validGitRefLocator(d.Ref) {
		return invalid("d.Ref")
	}
	if item, ok := d.Before.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.After.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// MutationPlanSummary is an immutable semantic boundary value. Its zero is invalid.
type MutationPlanSummary struct {
	data        MutationPlanSummaryData
	initialized bool
}

// MutationPlanSummaryData is a construction/copy value; NewMutationPlanSummary validates and owns a copy.
type MutationPlanSummaryData struct {
	OperationID          OperationID
	Kind                 GitMutationKind
	PlanVersion          SourceVersion
	Repository           domain.RepositoryID
	Worktree             Optional[domain.WorktreeID]
	Target               Optional[domain.ExactTarget]
	Head                 Optional[domain.Head]
	Stash                Optional[domain.StashID]
	Expected             GitExpectedState
	Paths                []PlannedPathEffect
	Refs                 []PlannedRefEffect
	IndexChange          Optional[SourceVersion]
	Choices              []Choice
	ConfirmationRequired bool
	RecoveryBehavior     string
	RequiredCapabilities []GitCapabilityFact
	OriginVersion        Optional[SourceVersion]
	OwnStepVersions      []FacetVersion
}

func NewMutationPlanSummary(d MutationPlanSummaryData) (MutationPlanSummary, error) {
	if err := d.validate(); err != nil {
		return MutationPlanSummary{}, err
	}
	return MutationPlanSummary{data: d.clone(), initialized: true}, nil
}
func (v MutationPlanSummary) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v MutationPlanSummary) Data() MutationPlanSummaryData { return v.data.clone() }
func (v MutationPlanSummary) Clone() MutationPlanSummary    { return v }
func (d MutationPlanSummaryData) clone() MutationPlanSummaryData {
	d.Paths = cloneSlice(d.Paths)
	d.Refs = cloneSlice(d.Refs)
	d.Choices = cloneSlice(d.Choices)
	d.RequiredCapabilities = cloneSlice(d.RequiredCapabilities)
	d.OwnStepVersions = cloneSlice(d.OwnStepVersions)
	return d
}
func (d MutationPlanSummaryData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.PlanVersion.Valid() {
		return invalid("d.PlanVersion")
	}
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Target.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Stash.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expected.Valid() {
		return invalid("d.Expected")
	}
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.IndexChange.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Choices {
		if !item.Valid() {
			return invalid("item")
		}
	}

	for _, item := range d.RequiredCapabilities {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.OriginVersion.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.OwnStepVersions {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Repository != d.Expected.data.Repository || len(d.Choices) == 0 {
		return invalid("plan summary")
	}
	return nil
}

// FacetVersion is an immutable semantic boundary value. Its zero is invalid.
type FacetVersion struct {
	data        FacetVersionData
	initialized bool
}

// FacetVersionData is a construction/copy value; NewFacetVersion validates and owns a copy.
type FacetVersionData struct {
	Facet  EffectFacet
	Before SourceVersion
	After  SourceVersion
}

func NewFacetVersion(d FacetVersionData) (FacetVersion, error) {
	if err := d.validate(); err != nil {
		return FacetVersion{}, err
	}
	return FacetVersion{data: d.clone(), initialized: true}, nil
}
func (v FacetVersion) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v FacetVersion) Data() FacetVersionData { return v.data.clone() }
func (v FacetVersion) Clone() FacetVersion    { return v }
func (d FacetVersionData) clone() FacetVersionData {

	return d
}
func (d FacetVersionData) validate() error {
	if !d.Facet.Valid() {
		return invalid("d.Facet")
	}
	if !d.Before.Valid() {
		return invalid("d.Before")
	}
	if !d.After.Valid() {
		return invalid("d.After")
	}

	return nil
}

// GitPreparationResult is an immutable semantic boundary value. Its zero is invalid.
type GitPreparationResult struct {
	data        GitPreparationResultData
	initialized bool
}

// GitPreparationResultData is a construction/copy value; NewGitPreparationResult validates and owns a copy.
type GitPreparationResultData struct {
	Summary                Optional[MutationPlanSummary]
	Observation            Optional[GitObservation]
	Diagnostics            []Diagnostic
	Transport              CommandTransportOutcome
	Effects                EffectReport
	CancellationRequested  bool
	Recovery               []RecoveryRecord
	ReconciliationRequired bool
}

func NewGitPreparationResult(d GitPreparationResultData) (GitPreparationResult, error) {
	if err := d.validate(); err != nil {
		return GitPreparationResult{}, err
	}
	return GitPreparationResult{data: d.clone(), initialized: true}, nil
}
func (v GitPreparationResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v GitPreparationResult) Data() GitPreparationResultData { return v.data.clone() }
func (v GitPreparationResult) Clone() GitPreparationResult    { return v }
func (d GitPreparationResultData) clone() GitPreparationResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d GitPreparationResultData) validate() error {
	if item, ok := d.Summary.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}

	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if err := validateGitPreparationResultEvidence(d); err != nil {
		return err
	}
	return validateGitRecovery(d.Effects, d.Recovery)
}

// GitCompletedStep is an immutable semantic boundary value. Its zero is invalid.
type GitCompletedStep struct {
	data        GitCompletedStepData
	initialized bool
}

// GitCompletedStepData is a construction/copy value; NewGitCompletedStep validates and owns a copy.
type GitCompletedStepData struct {
	Kind            GitStepKind
	Target          RecoverySubject
	Effect          FacetEffect
	PostObservation Optional[GitObservation]
}

func NewGitCompletedStep(d GitCompletedStepData) (GitCompletedStep, error) {
	if err := d.validate(); err != nil {
		return GitCompletedStep{}, err
	}
	return GitCompletedStep{data: d.clone(), initialized: true}, nil
}
func (v GitCompletedStep) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v GitCompletedStep) Data() GitCompletedStepData { return v.data.clone() }
func (v GitCompletedStep) Clone() GitCompletedStep    { return v }
func (d GitCompletedStepData) clone() GitCompletedStepData {

	return d
}
func (d GitCompletedStepData) validate() error {
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if !d.Effect.Valid() {
		return invalid("d.Effect")
	}
	if item, ok := d.PostObservation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// CommitCandidateFacts is an immutable semantic boundary value. Its zero is invalid.
type CommitCandidateFacts struct {
	data        CommitCandidateFactsData
	initialized bool
}

// CommitCandidateFactsData is a construction/copy value; NewCommitCandidateFacts validates and owns a copy.
type CommitCandidateFactsData struct {
	Candidate         Optional[domain.Revision]
	Parent            Optional[domain.Revision]
	CommitTree        Optional[domain.OID]
	FinalIndex        Optional[SourceVersion]
	FinalIndexTree    Optional[domain.OID]
	FinalIndexObjects []domain.OID
	Message           string
	CandidateEffect   EffectReport
	HookEffects       EffectReport
	RefEffects        EffectReport
	StagedIndexEffect EffectReport
}

func NewCommitCandidateFacts(d CommitCandidateFactsData) (CommitCandidateFacts, error) {
	if err := d.validate(); err != nil {
		return CommitCandidateFacts{}, err
	}
	return CommitCandidateFacts{data: d.clone(), initialized: true}, nil
}
func (v CommitCandidateFacts) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v CommitCandidateFacts) Data() CommitCandidateFactsData { return v.data.clone() }
func (v CommitCandidateFacts) Clone() CommitCandidateFacts    { return v }
func (d CommitCandidateFactsData) clone() CommitCandidateFactsData {
	d.FinalIndexObjects = cloneSlice(d.FinalIndexObjects)
	return d
}
func (d CommitCandidateFactsData) validate() error {
	if item, ok := d.Candidate.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Parent.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.CommitTree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.FinalIndex.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.FinalIndexTree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.FinalIndexObjects {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !d.CandidateEffect.Valid() {
		return invalid("d.CandidateEffect")
	}
	if !d.HookEffects.Valid() {
		return invalid("d.HookEffects")
	}
	if !d.RefEffects.Valid() {
		return invalid("d.RefEffects")
	}
	if !d.StagedIndexEffect.Valid() {
		return invalid("d.StagedIndexEffect")
	}

	return nil
}

// WorktreeCreated is an immutable semantic boundary value. Its zero is invalid.
type WorktreeCreated struct {
	data        WorktreeCreatedData
	initialized bool
}

// WorktreeCreatedData is a construction/copy value; NewWorktreeCreated validates and owns a copy.
type WorktreeCreatedData struct {
	Worktree WorktreeFacts
	Target   ExactLocalResolution
}

func NewWorktreeCreated(d WorktreeCreatedData) (WorktreeCreated, error) {
	if err := d.validate(); err != nil {
		return WorktreeCreated{}, err
	}
	return WorktreeCreated{data: d.clone(), initialized: true}, nil
}
func (v WorktreeCreated) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v WorktreeCreated) Data() WorktreeCreatedData { return v.data.clone() }
func (v WorktreeCreated) Clone() WorktreeCreated    { return v }
func (d WorktreeCreatedData) clone() WorktreeCreatedData {

	return d
}
func (d WorktreeCreatedData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.Target.Valid() {
		return invalid("d.Target")
	}

	return nil
}

// WorktreeRetargeted is an immutable semantic boundary value. Its zero is invalid.
type WorktreeRetargeted struct {
	data        WorktreeRetargetedData
	initialized bool
}

// WorktreeRetargetedData is a construction/copy value; NewWorktreeRetargeted validates and owns a copy.
type WorktreeRetargetedData struct {
	Worktree  WorktreeFacts
	PriorHead domain.Head
	Target    ExactLocalResolution
}

func NewWorktreeRetargeted(d WorktreeRetargetedData) (WorktreeRetargeted, error) {
	if err := d.validate(); err != nil {
		return WorktreeRetargeted{}, err
	}
	return WorktreeRetargeted{data: d.clone(), initialized: true}, nil
}
func (v WorktreeRetargeted) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v WorktreeRetargeted) Data() WorktreeRetargetedData { return v.data.clone() }
func (v WorktreeRetargeted) Clone() WorktreeRetargeted    { return v }
func (d WorktreeRetargetedData) clone() WorktreeRetargetedData {

	return d
}
func (d WorktreeRetargetedData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}
	if !d.PriorHead.Valid() {
		return invalid("d.PriorHead")
	}
	if !d.Target.Valid() {
		return invalid("d.Target")
	}

	return nil
}

// IndexChanged is an immutable semantic boundary value. Its zero is invalid.
type IndexChanged struct {
	data        IndexChangedData
	initialized bool
}

// IndexChangedData is a construction/copy value; NewIndexChanged validates and owns a copy.
type IndexChangedData struct {
	Status StatusFacts
	Action StageAction
}

func NewIndexChanged(d IndexChangedData) (IndexChanged, error) {
	if err := d.validate(); err != nil {
		return IndexChanged{}, err
	}
	return IndexChanged{data: d.clone(), initialized: true}, nil
}
func (v IndexChanged) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v IndexChanged) Data() IndexChangedData { return v.data.clone() }
func (v IndexChanged) Clone() IndexChanged    { return v }
func (d IndexChangedData) clone() IndexChangedData {

	return d
}
func (d IndexChangedData) validate() error {
	if !d.Status.Valid() {
		return invalid("d.Status")
	}
	if !d.Action.Valid() {
		return invalid("d.Action")
	}

	return nil
}

// CommitCreated is an immutable semantic boundary value. Its zero is invalid.
type CommitCreated struct {
	data        CommitCreatedData
	initialized bool
}

// CommitCreatedData is a construction/copy value; NewCommitCreated validates and owns a copy.
type CommitCreatedData struct {
	Revision          domain.Revision
	Head              domain.Head
	Index             SourceVersion
	StagedIndexEffect EffectReport
	Candidate         CommitCandidateFacts
}

func NewCommitCreated(d CommitCreatedData) (CommitCreated, error) {
	if err := d.validate(); err != nil {
		return CommitCreated{}, err
	}
	return CommitCreated{data: d.clone(), initialized: true}, nil
}
func (v CommitCreated) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v CommitCreated) Data() CommitCreatedData { return v.data.clone() }
func (v CommitCreated) Clone() CommitCreated    { return v }
func (d CommitCreatedData) clone() CommitCreatedData {

	return d
}
func (d CommitCreatedData) validate() error {
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	if !d.Head.Valid() {
		return invalid("d.Head")
	}
	if !d.Index.Valid() {
		return invalid("d.Index")
	}
	if !d.StagedIndexEffect.Valid() {
		return invalid("d.StagedIndexEffect")
	}
	if !d.Candidate.Valid() {
		return invalid("d.Candidate")
	}
	r, ok := d.Head.Revision()
	if !ok || r != d.Revision {
		return invalid("commit head")
	}
	return nil
}

// TrackedRestored is an immutable semantic boundary value. Its zero is invalid.
type TrackedRestored struct {
	data        TrackedRestoredData
	initialized bool
}

// TrackedRestoredData is a construction/copy value; NewTrackedRestored validates and owns a copy.
type TrackedRestoredData struct {
	Paths    []GitPath
	Status   StatusFacts
	Recovery []RecoveryRecord
}

func NewTrackedRestored(d TrackedRestoredData) (TrackedRestored, error) {
	if err := d.validate(); err != nil {
		return TrackedRestored{}, err
	}
	return TrackedRestored{data: d.clone(), initialized: true}, nil
}
func (v TrackedRestored) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v TrackedRestored) Data() TrackedRestoredData { return v.data.clone() }
func (v TrackedRestored) Clone() TrackedRestored    { return v }
func (d TrackedRestoredData) clone() TrackedRestoredData {
	d.Paths = cloneSlice(d.Paths)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d TrackedRestoredData) validate() error {
	for _, item := range d.Paths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Status.Valid() {
		return invalid("d.Status")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// StashCreated is an immutable semantic boundary value. Its zero is invalid.
type StashCreated struct {
	data        StashCreatedData
	initialized bool
}

// StashCreatedData is a construction/copy value; NewStashCreated validates and owns a copy.
type StashCreatedData struct {
	Stash   domain.StashID
	Status  StatusFacts
	Cleanup EffectReport
}

func NewStashCreated(d StashCreatedData) (StashCreated, error) {
	if err := d.validate(); err != nil {
		return StashCreated{}, err
	}
	return StashCreated{data: d.clone(), initialized: true}, nil
}
func (v StashCreated) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v StashCreated) Data() StashCreatedData { return v.data.clone() }
func (v StashCreated) Clone() StashCreated    { return v }
func (d StashCreatedData) clone() StashCreatedData {

	return d
}
func (d StashCreatedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Status.Valid() {
		return invalid("d.Status")
	}
	if !d.Cleanup.Valid() {
		return invalid("d.Cleanup")
	}

	return nil
}

// StashCreatedCleanupRefused is an immutable semantic boundary value. Its zero is invalid.
type StashCreatedCleanupRefused struct {
	data        StashCreatedCleanupRefusedData
	initialized bool
}

// StashCreatedCleanupRefusedData is a construction/copy value; NewStashCreatedCleanupRefused validates and owns a copy.
type StashCreatedCleanupRefusedData struct {
	Stash    domain.StashID
	Status   Optional[StatusFacts]
	Cleanup  EffectReport
	Reason   Diagnostic
	Recovery []RecoveryRecord
}

func NewStashCreatedCleanupRefused(d StashCreatedCleanupRefusedData) (StashCreatedCleanupRefused, error) {
	if err := d.validate(); err != nil {
		return StashCreatedCleanupRefused{}, err
	}
	return StashCreatedCleanupRefused{data: d.clone(), initialized: true}, nil
}
func (v StashCreatedCleanupRefused) Valid() bool                          { return v.initialized && v.data.validate() == nil }
func (v StashCreatedCleanupRefused) Data() StashCreatedCleanupRefusedData { return v.data.clone() }
func (v StashCreatedCleanupRefused) Clone() StashCreatedCleanupRefused    { return v }
func (d StashCreatedCleanupRefusedData) clone() StashCreatedCleanupRefusedData {
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d StashCreatedCleanupRefusedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Cleanup.Valid() {
		return invalid("d.Cleanup")
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// StashApplied is an immutable semantic boundary value. Its zero is invalid.
type StashApplied struct {
	data        StashAppliedData
	initialized bool
}

// StashAppliedData is a construction/copy value; NewStashApplied validates and owns a copy.
type StashAppliedData struct {
	Stash         domain.StashID
	Status        StatusFacts
	IndexRestored bool
	Retained      bool
}

func NewStashApplied(d StashAppliedData) (StashApplied, error) {
	if err := d.validate(); err != nil {
		return StashApplied{}, err
	}
	return StashApplied{data: d.clone(), initialized: true}, nil
}
func (v StashApplied) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v StashApplied) Data() StashAppliedData { return v.data.clone() }
func (v StashApplied) Clone() StashApplied    { return v }
func (d StashAppliedData) clone() StashAppliedData {

	return d
}
func (d StashAppliedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Status.Valid() {
		return invalid("d.Status")
	}

	return nil
}

// AppliedWithConflicts is an immutable semantic boundary value. Its zero is invalid.
type AppliedWithConflicts struct {
	data        AppliedWithConflictsData
	initialized bool
}

// AppliedWithConflictsData is a construction/copy value; NewAppliedWithConflicts validates and owns a copy.
type AppliedWithConflictsData struct {
	Stash         domain.StashID
	ConflictPaths []GitPath
	IndexEntries  []IndexEntryFact
	Status        Optional[StatusFacts]
	Recovery      []RecoveryRecord
}

func NewAppliedWithConflicts(d AppliedWithConflictsData) (AppliedWithConflicts, error) {
	if err := d.validate(); err != nil {
		return AppliedWithConflicts{}, err
	}
	return AppliedWithConflicts{data: d.clone(), initialized: true}, nil
}
func (v AppliedWithConflicts) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v AppliedWithConflicts) Data() AppliedWithConflictsData { return v.data.clone() }
func (v AppliedWithConflicts) Clone() AppliedWithConflicts    { return v }
func (d AppliedWithConflictsData) clone() AppliedWithConflictsData {
	d.ConflictPaths = cloneSlice(d.ConflictPaths)
	d.IndexEntries = cloneSlice(d.IndexEntries)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d AppliedWithConflictsData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	for _, item := range d.ConflictPaths {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.IndexEntries {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Status.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// StashDropped is an immutable semantic boundary value. Its zero is invalid.
type StashDropped struct {
	data        StashDroppedData
	initialized bool
}

// StashDroppedData is a construction/copy value; NewStashDropped validates and owns a copy.
type StashDroppedData struct {
	Stash       domain.StashID
	Occurrence  SourceVersion
	Survivors   []StashFact
	Observation GitObservation
	RefCleanup  EffectReport
}

func NewStashDropped(d StashDroppedData) (StashDropped, error) {
	if err := d.validate(); err != nil {
		return StashDropped{}, err
	}
	return StashDropped{data: d.clone(), initialized: true}, nil
}
func (v StashDropped) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v StashDropped) Data() StashDroppedData { return v.data.clone() }
func (v StashDropped) Clone() StashDropped    { return v }
func (d StashDroppedData) clone() StashDroppedData {
	d.Survivors = cloneSlice(d.Survivors)
	return d
}
func (d StashDroppedData) validate() error {
	if !d.Stash.Valid() {
		return invalid("d.Stash")
	}
	if !d.Occurrence.Valid() {
		return invalid("d.Occurrence")
	}
	for _, item := range d.Survivors {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if !d.RefCleanup.Valid() {
		return invalid("d.RefCleanup")
	}

	return nil
}

// BranchCreated is an immutable semantic boundary value. Its zero is invalid.
type BranchCreated struct {
	data        BranchCreatedData
	initialized bool
}

// BranchCreatedData is a construction/copy value; NewBranchCreated validates and owns a copy.
type BranchCreatedData struct {
	Branch   domain.BranchID
	Revision domain.Revision
	Worktree Optional[WorktreeFacts]
}

func NewBranchCreated(d BranchCreatedData) (BranchCreated, error) {
	if err := d.validate(); err != nil {
		return BranchCreated{}, err
	}
	return BranchCreated{data: d.clone(), initialized: true}, nil
}
func (v BranchCreated) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v BranchCreated) Data() BranchCreatedData { return v.data.clone() }
func (v BranchCreated) Clone() BranchCreated    { return v }
func (d BranchCreatedData) clone() BranchCreatedData {

	return d
}
func (d BranchCreatedData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.Branch.Kind() != domain.Local || d.Branch.Repository() != d.Revision.Repository() {
		return invalid("branch result")
	}
	return nil
}

// Pushed is an immutable semantic boundary value. Its zero is invalid.
type Pushed struct {
	data        PushedData
	initialized bool
}

// PushedData is a construction/copy value; NewPushed validates and owns a copy.
type PushedData struct {
	Source         domain.Revision
	Destination    domain.BranchID
	Binding        RemoteBinding
	ObservedRemote Optional[domain.Revision]
	RemoteEffect   EffectReport
	UpstreamEffect EffectReport
	Configuration  Optional[LocalConfigurationObservation]
}

func NewPushed(d PushedData) (Pushed, error) {
	if err := d.validate(); err != nil {
		return Pushed{}, err
	}
	return Pushed{data: d.clone(), initialized: true}, nil
}
func (v Pushed) Valid() bool      { return v.initialized && v.data.validate() == nil }
func (v Pushed) Data() PushedData { return v.data.clone() }
func (v Pushed) Clone() Pushed    { return v }
func (d PushedData) clone() PushedData {

	return d
}
func (d PushedData) validate() error {
	if !d.Source.Valid() {
		return invalid("d.Source")
	}
	if !d.Destination.Valid() {
		return invalid("d.Destination")
	}
	if !d.Binding.Valid() {
		return invalid("d.Binding")
	}
	if item, ok := d.ObservedRemote.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.RemoteEffect.Valid() {
		return invalid("d.RemoteEffect")
	}
	if !d.UpstreamEffect.Valid() {
		return invalid("d.UpstreamEffect")
	}
	if item, ok := d.Configuration.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// RefusedMutation is an immutable semantic boundary value. Its zero is invalid.
type RefusedMutation struct {
	data        RefusedMutationData
	initialized bool
}

// RefusedMutationData is a construction/copy value; NewRefusedMutation validates and owns a copy.
type RefusedMutationData struct {
	Reason  Diagnostic
	Effects EffectReport
}

func NewRefusedMutation(d RefusedMutationData) (RefusedMutation, error) {
	if err := d.validate(); err != nil {
		return RefusedMutation{}, err
	}
	return RefusedMutation{data: d.clone(), initialized: true}, nil
}
func (v RefusedMutation) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v RefusedMutation) Data() RefusedMutationData { return v.data.clone() }
func (v RefusedMutation) Clone() RefusedMutation    { return v }
func (d RefusedMutationData) clone() RefusedMutationData {

	return d
}
func (d RefusedMutationData) validate() error {
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}

	return nil
}

// GitPostFacts is an immutable semantic boundary value. Its zero is invalid.
type GitPostFacts struct {
	data        GitPostFactsData
	initialized bool
}

// GitPostFactsData is a construction/copy value; NewGitPostFacts validates and owns a copy.
type GitPostFactsData struct {
	Worktrees     []WorktreeFacts
	Status        []StatusFacts
	Refs          []RefFact
	Stashes       []StashFact
	Commit        Optional[CommitCandidateFacts]
	Configuration []LocalConfigurationObservation
}

func NewGitPostFacts(d GitPostFactsData) (GitPostFacts, error) {
	if err := d.validate(); err != nil {
		return GitPostFacts{}, err
	}
	return GitPostFacts{data: d.clone(), initialized: true}, nil
}
func (v GitPostFacts) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v GitPostFacts) Data() GitPostFactsData { return v.data.clone() }
func (v GitPostFacts) Clone() GitPostFacts    { return v }
func (d GitPostFactsData) clone() GitPostFactsData {
	d.Worktrees = cloneSlice(d.Worktrees)
	d.Status = cloneSlice(d.Status)
	d.Refs = cloneSlice(d.Refs)
	d.Stashes = cloneSlice(d.Stashes)
	d.Configuration = cloneSlice(d.Configuration)
	return d
}
func (d GitPostFactsData) validate() error {
	for _, item := range d.Worktrees {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Status {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Stashes {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Commit.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Configuration {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// PartialMutation is an immutable semantic boundary value. Its zero is invalid.
type PartialMutation struct {
	data        PartialMutationData
	initialized bool
}

// PartialMutationData is a construction/copy value; NewPartialMutation validates and owns a copy.
type PartialMutationData struct {
	Facts                  GitPostFacts
	FailedSteps            []GitStepKind
	Reason                 Diagnostic
	Recovery               []RecoveryRecord
	ReconciliationRequired bool
}

func NewPartialMutation(d PartialMutationData) (PartialMutation, error) {
	if err := d.validate(); err != nil {
		return PartialMutation{}, err
	}
	return PartialMutation{data: d.clone(), initialized: true}, nil
}
func (v PartialMutation) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v PartialMutation) Data() PartialMutationData { return v.data.clone() }
func (v PartialMutation) Clone() PartialMutation    { return v }
func (d PartialMutationData) clone() PartialMutationData {
	d.FailedSteps = cloneSlice(d.FailedSteps)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d PartialMutationData) validate() error {
	if !d.Facts.Valid() {
		return invalid("d.Facts")
	}
	for _, item := range d.FailedSteps {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// MutationIndeterminate is an immutable semantic boundary value. Its zero is invalid.
type MutationIndeterminate struct {
	data        MutationIndeterminateData
	initialized bool
}

// MutationIndeterminateData is a construction/copy value; NewMutationIndeterminate validates and owns a copy.
type MutationIndeterminateData struct {
	Facts                  GitPostFacts
	UnclassifiedSteps      []GitStepKind
	Reason                 Diagnostic
	Recovery               []RecoveryRecord
	ReconciliationRequired bool
}

func NewMutationIndeterminate(d MutationIndeterminateData) (MutationIndeterminate, error) {
	if err := d.validate(); err != nil {
		return MutationIndeterminate{}, err
	}
	return MutationIndeterminate{data: d.clone(), initialized: true}, nil
}
func (v MutationIndeterminate) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v MutationIndeterminate) Data() MutationIndeterminateData { return v.data.clone() }
func (v MutationIndeterminate) Clone() MutationIndeterminate    { return v }
func (d MutationIndeterminateData) clone() MutationIndeterminateData {
	d.UnclassifiedSteps = cloneSlice(d.UnclassifiedSteps)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d MutationIndeterminateData) validate() error {
	if !d.Facts.Valid() {
		return invalid("d.Facts")
	}
	for _, item := range d.UnclassifiedSteps {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// GitMutationResult is an immutable semantic boundary value. Its zero is invalid.
type GitMutationResult struct {
	data        GitMutationResultData
	initialized bool
}

// GitMutationResultData is a construction/copy value; NewGitMutationResult validates and owns a copy.
type GitMutationResultData struct {
	Operation              OperationID
	Kind                   GitMutationKind
	PlanVersion            SourceVersion
	Steps                  []GitCompletedStep
	Outcome                GitMutationOutcome
	Observation            Optional[GitObservation]
	Diagnostics            []Diagnostic
	Transport              CommandTransportOutcome
	Effects                EffectReport
	CancellationRequested  bool
	Recovery               []RecoveryRecord
	ReconciliationRequired bool
}

func NewGitMutationResult(d GitMutationResultData) (GitMutationResult, error) {
	if err := d.validate(); err != nil {
		return GitMutationResult{}, err
	}
	return GitMutationResult{data: d.clone(), initialized: true}, nil
}
func (v GitMutationResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v GitMutationResult) Data() GitMutationResultData { return v.data.clone() }
func (v GitMutationResult) Clone() GitMutationResult    { return v }
func (d GitMutationResultData) clone() GitMutationResultData {
	d.Steps = cloneSlice(d.Steps)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d GitMutationResultData) validate() error {
	if !d.Operation.Valid() {
		return invalid("d.Operation")
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.PlanVersion.Valid() {
		return invalid("d.PlanVersion")
	}
	for _, item := range d.Steps {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validGitMutationOutcome(d.Outcome) {
		return invalid("d.Outcome")
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}

	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if err := validateGitMutationResultEvidence(d); err != nil {
		return err
	}
	return validateGitMutation(d)
}

// FetchRefSpec is an immutable semantic boundary value. Its zero is invalid.
type FetchRefSpec struct {
	data        FetchRefSpecData
	initialized bool
}

// FetchRefSpecData is a construction/copy value; NewFetchRefSpec validates and owns a copy.
type FetchRefSpecData struct {
	SourceRef      string
	Destination    FetchDestinationPolicy
	DestinationRef Optional[string]
	Expected       Optional[domain.Revision]
}

func NewFetchRefSpec(d FetchRefSpecData) (FetchRefSpec, error) {
	if err := d.validate(); err != nil {
		return FetchRefSpec{}, err
	}
	return FetchRefSpec{data: d.clone(), initialized: true}, nil
}
func (v FetchRefSpec) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v FetchRefSpec) Data() FetchRefSpecData { return v.data.clone() }
func (v FetchRefSpec) Clone() FetchRefSpec    { return v }
func (d FetchRefSpecData) clone() FetchRefSpecData {

	return d
}
func (d FetchRefSpecData) validate() error {

	if !d.Destination.Valid() {
		return invalid("d.Destination")
	}

	if item, ok := d.Expected.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validFullRef(d.SourceRef, "refs/") {
		return invalid("fetch source")
	}
	if r, ok := d.DestinationRef.Value(); ok && !validFullRef(r, "refs/") {
		return invalid("fetch destination")
	}
	return nil
}

// FetchRequest is an immutable semantic boundary value. Its zero is invalid.
type FetchRequest struct {
	data        FetchRequestData
	initialized bool
}

// FetchRequestData is a construction/copy value; NewFetchRequest validates and owns a copy.
type FetchRequestData struct {
	Operation       OperationID
	Binding         RemoteBinding
	RefScope        []FetchRefSpec
	Prune           bool
	ExpectedBinding SourceVersion
}

func NewFetchRequest(d FetchRequestData) (FetchRequest, error) {
	if err := d.validate(); err != nil {
		return FetchRequest{}, err
	}
	return FetchRequest{data: d.clone(), initialized: true}, nil
}
func (v FetchRequest) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v FetchRequest) Data() FetchRequestData { return v.data.clone() }
func (v FetchRequest) Clone() FetchRequest    { return v }
func (d FetchRequestData) clone() FetchRequestData {
	d.RefScope = cloneSlice(d.RefScope)
	return d
}
func (d FetchRequestData) validate() error {
	if !d.Operation.Valid() {
		return invalid("d.Operation")
	}
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
		return invalid("fetch binding")
	}
	return nil
}

// FetchResult is an immutable semantic boundary value. Its zero is invalid.
type FetchResult struct {
	data        FetchResultData
	initialized bool
}

// FetchResultData is a construction/copy value; NewFetchResult validates and owns a copy.
type FetchResultData struct {
	Generation             Optional[FetchGeneration]
	Refs                   []RefFact
	Freshness              FetchFreshness
	Observation            Optional[GitObservation]
	Diagnostics            []Diagnostic
	Transport              CommandTransportOutcome
	Effects                EffectReport
	CancellationRequested  bool
	Recovery               []RecoveryRecord
	ReconciliationRequired bool
}

func NewFetchResult(d FetchResultData) (FetchResult, error) {
	if err := d.validate(); err != nil {
		return FetchResult{}, err
	}
	return FetchResult{data: d.clone(), initialized: true}, nil
}
func (v FetchResult) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v FetchResult) Data() FetchResultData { return v.data.clone() }
func (v FetchResult) Clone() FetchResult    { return v }
func (d FetchResultData) clone() FetchResultData {
	d.Refs = cloneSlice(d.Refs)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d FetchResultData) validate() error {
	if item, ok := d.Generation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Freshness.Valid() {
		return invalid("d.Freshness")
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}

	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if err := validateFetchResultEvidence(d); err != nil {
		return err
	}
	return validateGitRecovery(d.Effects, d.Recovery)
}

// LocalConfigurationObservation is an immutable semantic boundary value. Its zero is invalid.
type LocalConfigurationObservation struct {
	data        LocalConfigurationObservationData
	initialized bool
}

// LocalConfigurationObservationData is a construction/copy value; NewLocalConfigurationObservation validates and owns a copy.
type LocalConfigurationObservationData struct {
	Branch               domain.BranchID
	Upstream             UpstreamFact
	ConfigurationVersion SourceVersion
	Observation          GitObservation
}

func NewLocalConfigurationObservation(d LocalConfigurationObservationData) (LocalConfigurationObservation, error) {
	if err := d.validate(); err != nil {
		return LocalConfigurationObservation{}, err
	}
	return LocalConfigurationObservation{data: d.clone(), initialized: true}, nil
}
func (v LocalConfigurationObservation) Valid() bool { return v.initialized && v.data.validate() == nil }
func (v LocalConfigurationObservation) Data() LocalConfigurationObservationData {
	return v.data.clone()
}
func (v LocalConfigurationObservation) Clone() LocalConfigurationObservation { return v }
func (d LocalConfigurationObservationData) clone() LocalConfigurationObservationData {

	return d
}
func (d LocalConfigurationObservationData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !validUpstreamFact(d.Upstream) {
		return invalid("d.Upstream")
	}
	if !d.ConfigurationVersion.Valid() {
		return invalid("d.ConfigurationVersion")
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Branch.Kind() != domain.Local || d.Branch.Repository() != d.Observation.data.Repository {
		return invalid("configuration scope")
	}
	return nil
}

// ReconcileRequest is an immutable semantic boundary value. Its zero is invalid.
type ReconcileRequest struct {
	data        ReconcileRequestData
	initialized bool
}

// ReconcileRequestData is a construction/copy value; NewReconcileRequest validates and owns a copy.
type ReconcileRequestData struct {
	Operation         OperationID
	OriginalOperation OperationID
	Repository        domain.RepositoryID
	Worktree          Optional[domain.WorktreeID]
	Kind              GitMutationKind
	Targets           []domain.ExactTarget
	PriorEffects      EffectReport
	Recovery          []RecoveryRecord
	Facets            []EffectFacet
}

func NewReconcileRequest(d ReconcileRequestData) (ReconcileRequest, error) {
	if err := d.validate(); err != nil {
		return ReconcileRequest{}, err
	}
	return ReconcileRequest{data: d.clone(), initialized: true}, nil
}
func (v ReconcileRequest) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v ReconcileRequest) Data() ReconcileRequestData { return v.data.clone() }
func (v ReconcileRequest) Clone() ReconcileRequest    { return v }
func (d ReconcileRequestData) clone() ReconcileRequestData {
	d.Targets = cloneSlice(d.Targets)
	d.Recovery = cloneSlice(d.Recovery)
	d.Facets = cloneSlice(d.Facets)
	return d
}
func (d ReconcileRequestData) validate() error {
	if !d.Operation.Valid() {
		return invalid("d.Operation")
	}
	if !d.OriginalOperation.Valid() {
		return invalid("d.OriginalOperation")
	}
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	for _, item := range d.Targets {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.PriorEffects.Valid() {
		return invalid("d.PriorEffects")
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Facets {
		if !item.Valid() {
			return invalid("item")
		}
	}
	return validateReconcile(d)
}

// ReconciledFacet is an immutable semantic boundary value. Its zero is invalid.
type ReconciledFacet struct {
	data        ReconciledFacetData
	initialized bool
}

// ReconciledFacetData is a construction/copy value; NewReconciledFacet validates and owns a copy.
type ReconciledFacetData struct {
	Facet         EffectFacet
	Objects       []domain.OID
	Recovery      []RecoveryRecord
	Refs          []RefFact
	Statuses      []StatusFacts
	Stashes       []StashFact
	Configuration []LocalConfigurationObservation
	Observation   Optional[GitObservation]
	Diagnostics   []Diagnostic
}

func NewReconciledFacet(d ReconciledFacetData) (ReconciledFacet, error) {
	if err := d.validate(); err != nil {
		return ReconciledFacet{}, err
	}
	return ReconciledFacet{data: d.clone(), initialized: true}, nil
}
func (v ReconciledFacet) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ReconciledFacet) Data() ReconciledFacetData { return v.data.clone() }
func (v ReconciledFacet) Clone() ReconciledFacet    { return v }
func (d ReconciledFacetData) clone() ReconciledFacetData {
	d.Objects = cloneSlice(d.Objects)
	d.Recovery = cloneSlice(d.Recovery)
	d.Refs = cloneSlice(d.Refs)
	d.Statuses = cloneSlice(d.Statuses)
	d.Stashes = cloneSlice(d.Stashes)
	d.Configuration = cloneSlice(d.Configuration)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ReconciledFacetData) validate() error {
	if !d.Facet.Valid() {
		return invalid("d.Facet")
	}
	for _, item := range d.Objects {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Refs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Statuses {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Stashes {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Configuration {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !gitFacet(d.Facet) {
		return invalid("reconcile facet")
	}
	return nil
}

// ReconcileResult is an immutable semantic boundary value. Its zero is invalid.
type ReconcileResult struct {
	data        ReconcileResultData
	initialized bool
}

// ReconcileResultData is a construction/copy value; NewReconcileResult validates and owns a copy.
type ReconcileResultData struct {
	Facets                 []ReconciledFacet
	Observation            Optional[GitObservation]
	Diagnostics            []Diagnostic
	Transport              CommandTransportOutcome
	Effects                EffectReport
	CancellationRequested  bool
	Recovery               []RecoveryRecord
	ReconciliationRequired bool
}

func NewReconcileResult(d ReconcileResultData) (ReconcileResult, error) {
	if err := d.validate(); err != nil {
		return ReconcileResult{}, err
	}
	return ReconcileResult{data: d.clone(), initialized: true}, nil
}
func (v ReconcileResult) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ReconcileResult) Data() ReconcileResultData { return v.data.clone() }
func (v ReconcileResult) Clone() ReconcileResult    { return v }
func (d ReconcileResultData) clone() ReconcileResultData {
	d.Facets = cloneSlice(d.Facets)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	d.Recovery = cloneSlice(d.Recovery)
	return d
}
func (d ReconcileResultData) validate() error {
	for _, item := range d.Facets {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Transport.Valid() {
		return invalid("d.Transport")
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}

	for _, item := range d.Recovery {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if err := validateReconcileResultEvidence(d); err != nil {
		return err
	}
	return validateGitRecovery(d.Effects, d.Recovery)
}
