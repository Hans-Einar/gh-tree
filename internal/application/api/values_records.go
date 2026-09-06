package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"time"
)

// QueryCorrelation is an immutable semantic boundary value. Its zero is invalid.
type QueryCorrelation struct {
	data        QueryCorrelationData
	initialized bool
}

// QueryCorrelationData is a construction/copy value; NewQueryCorrelation validates and owns a copy.
type QueryCorrelationData struct {
	Slot       QuerySlot
	Generation QueryGeneration
}

func NewQueryCorrelation(d QueryCorrelationData) (QueryCorrelation, error) {
	if err := d.validate(); err != nil {
		return QueryCorrelation{}, err
	}
	return QueryCorrelation{data: d.clone(), initialized: true}, nil
}
func (v QueryCorrelation) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v QueryCorrelation) Data() QueryCorrelationData { return v.data.clone() }
func (v QueryCorrelation) Clone() QueryCorrelation    { return v }
func (d QueryCorrelationData) clone() QueryCorrelationData {

	return d
}
func (d QueryCorrelationData) validate() error {
	if !d.Slot.Valid() {
		return invalid("d.Slot")
	}
	if !d.Generation.Valid() {
		return invalid("d.Generation")
	}

	return nil
}

// Correlation is an immutable semantic boundary value. Its zero is invalid.
type Correlation struct {
	data        CorrelationData
	initialized bool
}

// CorrelationData is a construction/copy value; NewCorrelation validates and owns a copy.
type CorrelationData struct {
	Intent IntentToken
	Query  Optional[QueryCorrelation]
}

func NewCorrelation(d CorrelationData) (Correlation, error) {
	if err := d.validate(); err != nil {
		return Correlation{}, err
	}
	return Correlation{data: d.clone(), initialized: true}, nil
}
func (v Correlation) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v Correlation) Data() CorrelationData { return v.data.clone() }
func (v Correlation) Clone() Correlation    { return v }
func (d CorrelationData) clone() CorrelationData {

	return d
}
func (d CorrelationData) validate() error {
	if !d.Intent.Valid() {
		return invalid("d.Intent")
	}
	if item, ok := d.Query.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// ObservationInterval is an immutable semantic boundary value. Its zero is invalid.
type ObservationInterval struct {
	data        ObservationIntervalData
	initialized bool
}

// ObservationIntervalData is a construction/copy value; NewObservationInterval validates and owns a copy.
type ObservationIntervalData struct {
	StartedAt  time.Time
	FinishedAt time.Time
}

func NewObservationInterval(d ObservationIntervalData) (ObservationInterval, error) {
	if err := d.validate(); err != nil {
		return ObservationInterval{}, err
	}
	return ObservationInterval{data: d.clone(), initialized: true}, nil
}
func (v ObservationInterval) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ObservationInterval) Data() ObservationIntervalData { return v.data.clone() }
func (v ObservationInterval) Clone() ObservationInterval    { return v }
func (d ObservationIntervalData) clone() ObservationIntervalData {

	return d
}
func (d ObservationIntervalData) validate() error {

	if d.StartedAt.IsZero() || d.FinishedAt.IsZero() || d.FinishedAt.Before(d.StartedAt) {
		return invalid("observation interval")
	}
	_, a := d.StartedAt.Zone()
	_, b := d.FinishedAt.Zone()
	if a != 0 || b != 0 {
		return invalid("UTC interval")
	}
	return nil
}

// InitialPage is an immutable semantic boundary value. Its zero is invalid.
type InitialPage struct {
	data        InitialPageData
	initialized bool
}

// InitialPageData is a construction/copy value; NewInitialPage validates and owns a copy.
type InitialPageData struct {
}

func NewInitialPage(d InitialPageData) (InitialPage, error) {
	if err := d.validate(); err != nil {
		return InitialPage{}, err
	}
	return InitialPage{data: d.clone(), initialized: true}, nil
}
func (v InitialPage) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v InitialPage) Data() InitialPageData { return v.data.clone() }
func (v InitialPage) Clone() InitialPage    { return v }
func (d InitialPageData) clone() InitialPageData {

	return d
}
func (d InitialPageData) validate() error {

	return nil
}

// CursorPage is an immutable semantic boundary value. Its zero is invalid.
type CursorPage struct {
	data        CursorPageData
	initialized bool
}

// CursorPageData is a construction/copy value; NewCursorPage validates and owns a copy.
type CursorPageData struct {
	Cursor string
}

func NewCursorPage(d CursorPageData) (CursorPage, error) {
	if err := d.validate(); err != nil {
		return CursorPage{}, err
	}
	return CursorPage{data: d.clone(), initialized: true}, nil
}
func (v CursorPage) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v CursorPage) Data() CursorPageData { return v.data.clone() }
func (v CursorPage) Clone() CursorPage    { return v }
func (d CursorPageData) clone() CursorPageData {

	return d
}
func (d CursorPageData) validate() error {

	if !nonempty(d.Cursor) {
		return invalid("cursor")
	}
	return nil
}

// OffsetPage is an immutable semantic boundary value. Its zero is invalid.
type OffsetPage struct {
	data        OffsetPageData
	initialized bool
}

// OffsetPageData is a construction/copy value; NewOffsetPage validates and owns a copy.
type OffsetPageData struct {
	Offset uint64
}

func NewOffsetPage(d OffsetPageData) (OffsetPage, error) {
	if err := d.validate(); err != nil {
		return OffsetPage{}, err
	}
	return OffsetPage{data: d.clone(), initialized: true}, nil
}
func (v OffsetPage) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v OffsetPage) Data() OffsetPageData { return v.data.clone() }
func (v OffsetPage) Clone() OffsetPage    { return v }
func (d OffsetPageData) clone() OffsetPageData {

	return d
}
func (d OffsetPageData) validate() error {

	return nil
}

// PageRequest is an immutable semantic boundary value. Its zero is invalid.
type PageRequest struct {
	data        PageRequestData
	initialized bool
}

// PageRequestData is a construction/copy value; NewPageRequest validates and owns a copy.
type PageRequestData struct {
	Limit        uint32
	Continuation PageContinuation
}

func NewPageRequest(d PageRequestData) (PageRequest, error) {
	if err := d.validate(); err != nil {
		return PageRequest{}, err
	}
	return PageRequest{data: d.clone(), initialized: true}, nil
}
func (v PageRequest) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v PageRequest) Data() PageRequestData { return v.data.clone() }
func (v PageRequest) Clone() PageRequest    { return v }
func (d PageRequestData) clone() PageRequestData {

	return d
}
func (d PageRequestData) validate() error {

	if !validPageContinuation(d.Continuation) {
		return invalid("d.Continuation")
	}
	if d.Limit == 0 {
		return invalid("page limit")
	}
	return nil
}

// PageInfo is an immutable semantic boundary value. Its zero is invalid.
type PageInfo struct {
	data        PageInfoData
	initialized bool
}

// PageInfoData is a construction/copy value; NewPageInfo validates and owns a copy.
type PageInfoData struct {
	Returned     uint32
	Completeness Completeness
	Next         Optional[PageContinuation]
	HasMore      Optional[bool]
	Source       SourceVersion
}

func NewPageInfo(d PageInfoData) (PageInfo, error) {
	if err := d.validate(); err != nil {
		return PageInfo{}, err
	}
	return PageInfo{data: d.clone(), initialized: true}, nil
}
func (v PageInfo) Valid() bool        { return v.initialized && v.data.validate() == nil }
func (v PageInfo) Data() PageInfoData { return v.data.clone() }
func (v PageInfo) Clone() PageInfo    { return v }
func (d PageInfoData) clone() PageInfoData {

	return d
}
func (d PageInfoData) validate() error {

	if !d.Completeness.Valid() {
		return invalid("d.Completeness")
	}
	if item, ok := d.Next.Value(); ok {
		if !validPageContinuation(item) {
			return invalid("item")
		}
	}

	if !d.Source.Valid() {
		return invalid("d.Source")
	}
	if p, ok := d.Next.Value(); ok {
		if _, initial := p.(InitialPage); initial {
			return invalid("initial continuation")
		}
	}
	if d.Completeness == Complete {
		if d.Next.Present() {
			return invalid("complete next")
		}
		if m, ok := d.HasMore.Value(); ok && m {
			return invalid("complete has more")
		}
	}
	return nil
}

// WorktreeScope is an immutable semantic boundary value. Its zero is invalid.
type WorktreeScope struct {
	data        WorktreeScopeData
	initialized bool
}

// WorktreeScopeData is a construction/copy value; NewWorktreeScope validates and owns a copy.
type WorktreeScopeData struct {
	ID           domain.WorktreeID
	RootLocator  string
	RootIdentity DirectoryIdentity
	Source       SourceVersion
}

func NewWorktreeScope(d WorktreeScopeData) (WorktreeScope, error) {
	if err := d.validate(); err != nil {
		return WorktreeScope{}, err
	}
	return WorktreeScope{data: d.clone(), initialized: true}, nil
}
func (v WorktreeScope) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v WorktreeScope) Data() WorktreeScopeData { return v.data.clone() }
func (v WorktreeScope) Clone() WorktreeScope    { return v }
func (d WorktreeScopeData) clone() WorktreeScopeData {

	return d
}
func (d WorktreeScopeData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
	}

	if !d.RootIdentity.Valid() {
		return invalid("d.RootIdentity")
	}
	if !d.Source.Valid() {
		return invalid("d.Source")
	}
	if !nonempty(d.RootLocator) {
		return invalid("root locator")
	}
	return nil
}

// CwdObservation is an immutable semantic boundary value. Its zero is invalid.
type CwdObservation struct {
	data        CwdObservationData
	initialized bool
}

// CwdObservationData is a construction/copy value; NewCwdObservation validates and owns a copy.
type CwdObservationData struct {
	Worktree          WorktreeScope
	ProjectComponents []string
	ProjectIdentity   DirectoryIdentity
	Source            SourceVersion
}

func NewCwdObservation(d CwdObservationData) (CwdObservation, error) {
	if err := d.validate(); err != nil {
		return CwdObservation{}, err
	}
	return CwdObservation{data: d.clone(), initialized: true}, nil
}
func (v CwdObservation) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v CwdObservation) Data() CwdObservationData { return v.data.clone() }
func (v CwdObservation) Clone() CwdObservation    { return v }
func (d CwdObservationData) clone() CwdObservationData {
	d.ProjectComponents = cloneSlice(d.ProjectComponents)
	return d
}
func (d CwdObservationData) validate() error {
	if !d.Worktree.Valid() {
		return invalid("d.Worktree")
	}

	if !d.ProjectIdentity.Valid() {
		return invalid("d.ProjectIdentity")
	}
	if !d.Source.Valid() {
		return invalid("d.Source")
	}
	if !components(d.ProjectComponents) {
		return invalid("project components")
	}
	return nil
}

// RefspecMapping is an immutable semantic boundary value. Its zero is invalid.
type RefspecMapping struct {
	data        RefspecMappingData
	initialized bool
}

// RefspecMappingData is a construction/copy value; NewRefspecMapping validates and owns a copy.
type RefspecMappingData struct {
	Source      string
	Destination string
	Force       bool
}

func NewRefspecMapping(d RefspecMappingData) (RefspecMapping, error) {
	if err := d.validate(); err != nil {
		return RefspecMapping{}, err
	}
	return RefspecMapping{data: d.clone(), initialized: true}, nil
}
func (v RefspecMapping) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v RefspecMapping) Data() RefspecMappingData { return v.data.clone() }
func (v RefspecMapping) Clone() RefspecMapping    { return v }
func (d RefspecMappingData) clone() RefspecMappingData {

	return d
}
func (d RefspecMappingData) validate() error {

	if !nonempty(d.Source) || !nonempty(d.Destination) {
		return invalid("refspec")
	}
	return nil
}

// RemoteBinding is an immutable semantic boundary value. Its zero is invalid.
type RemoteBinding struct {
	data        RemoteBindingData
	initialized bool
}

// RemoteBindingData is a construction/copy value; NewRemoteBinding validates and owns a copy.
type RemoteBindingData struct {
	LocalRepository  domain.RepositoryID
	RemoteRepository domain.RepositoryID
	RemoteName       string
	FetchURLs        []string
	PushURLs         []string
	Refspecs         []RefspecMapping
	Configuration    SourceVersion
}

func NewRemoteBinding(d RemoteBindingData) (RemoteBinding, error) {
	if err := d.validate(); err != nil {
		return RemoteBinding{}, err
	}
	return RemoteBinding{data: d.clone(), initialized: true}, nil
}
func (v RemoteBinding) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v RemoteBinding) Data() RemoteBindingData { return v.data.clone() }
func (v RemoteBinding) Clone() RemoteBinding    { return v }
func (d RemoteBindingData) clone() RemoteBindingData {
	d.FetchURLs = cloneSlice(d.FetchURLs)
	d.PushURLs = cloneSlice(d.PushURLs)
	d.Refspecs = cloneSlice(d.Refspecs)
	return d
}
func (d RemoteBindingData) validate() error {
	if !d.LocalRepository.Valid() {
		return invalid("d.LocalRepository")
	}
	if !d.RemoteRepository.Valid() {
		return invalid("d.RemoteRepository")
	}

	for _, item := range d.Refspecs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Configuration.Valid() {
		return invalid("d.Configuration")
	}
	if d.LocalRepository.Scope() != domain.LocalCommon || d.RemoteRepository.Scope() != domain.Remote || !nonempty(d.RemoteName) {
		return invalid("remote binding scope")
	}
	for _, u := range append(cloneSlice(d.FetchURLs), d.PushURLs...) {
		if !safeLocator(u) {
			return invalid("sanitized URL")
		}
	}
	return nil
}

// Geometry is an immutable semantic boundary value. Its zero is invalid.
type Geometry struct {
	data        GeometryData
	initialized bool
}

// GeometryData is a construction/copy value; NewGeometry validates and owns a copy.
type GeometryData struct {
	Rows    int
	Columns int
}

func NewGeometry(d GeometryData) (Geometry, error) {
	if err := d.validate(); err != nil {
		return Geometry{}, err
	}
	return Geometry{data: d.clone(), initialized: true}, nil
}
func (v Geometry) Valid() bool        { return v.initialized && v.data.validate() == nil }
func (v Geometry) Data() GeometryData { return v.data.clone() }
func (v Geometry) Clone() Geometry    { return v }
func (d GeometryData) clone() GeometryData {

	return d
}
func (d GeometryData) validate() error {

	if d.Rows < 1 || d.Columns < 1 || d.Rows > 32767 || d.Columns > 32767 {
		return invalid("geometry")
	}
	return nil
}

// EnvironmentEntry is an immutable semantic boundary value. Its zero is invalid.
type EnvironmentEntry struct {
	data        EnvironmentEntryData
	initialized bool
}

// EnvironmentEntryData is a construction/copy value; NewEnvironmentEntry validates and owns a copy.
type EnvironmentEntryData struct {
	Name  string
	Value string
}

func NewEnvironmentEntry(d EnvironmentEntryData) (EnvironmentEntry, error) {
	if err := d.validate(); err != nil {
		return EnvironmentEntry{}, err
	}
	return EnvironmentEntry{data: d.clone(), initialized: true}, nil
}
func (v EnvironmentEntry) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v EnvironmentEntry) Data() EnvironmentEntryData { return v.data.clone() }
func (v EnvironmentEntry) Clone() EnvironmentEntry    { return v }
func (d EnvironmentEntryData) clone() EnvironmentEntryData {

	return d
}
func (d EnvironmentEntryData) validate() error {

	if !nonempty(d.Name) || !literal(d.Value) || containsEquals(d.Name) {
		return invalid("environment entry")
	}
	return nil
}

// EnvironmentPolicy is an immutable semantic boundary value. Its zero is invalid.
type EnvironmentPolicy struct {
	data        EnvironmentPolicyData
	initialized bool
}

// EnvironmentPolicyData is a construction/copy value; NewEnvironmentPolicy validates and owns a copy.
type EnvironmentPolicyData struct {
	InheritBase bool
	Set         []EnvironmentEntry
	Remove      []string
}

func NewEnvironmentPolicy(d EnvironmentPolicyData) (EnvironmentPolicy, error) {
	if err := d.validate(); err != nil {
		return EnvironmentPolicy{}, err
	}
	return EnvironmentPolicy{data: d.clone(), initialized: true}, nil
}
func (v EnvironmentPolicy) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v EnvironmentPolicy) Data() EnvironmentPolicyData { return v.data.clone() }
func (v EnvironmentPolicy) Clone() EnvironmentPolicy    { return v }
func (d EnvironmentPolicyData) clone() EnvironmentPolicyData {
	d.Set = cloneSlice(d.Set)
	d.Remove = cloneSlice(d.Remove)
	return d
}
func (d EnvironmentPolicyData) validate() error {

	for _, item := range d.Set {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !validEnvironment(d) {
		return invalid("environment uniqueness")
	}
	return nil
}

// ArgvExecution is an immutable semantic boundary value. Its zero is invalid.
type ArgvExecution struct {
	data        ArgvExecutionData
	initialized bool
}

// ArgvExecutionData is a construction/copy value; NewArgvExecution validates and owns a copy.
type ArgvExecutionData struct {
	Executable string
	Arguments  []string
}

func NewArgvExecution(d ArgvExecutionData) (ArgvExecution, error) {
	if err := d.validate(); err != nil {
		return ArgvExecution{}, err
	}
	return ArgvExecution{data: d.clone(), initialized: true}, nil
}
func (v ArgvExecution) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v ArgvExecution) Data() ArgvExecutionData { return v.data.clone() }
func (v ArgvExecution) Clone() ArgvExecution    { return v }
func (d ArgvExecutionData) clone() ArgvExecutionData {
	d.Arguments = cloneSlice(d.Arguments)
	return d
}
func (d ArgvExecutionData) validate() error {

	if !nonempty(d.Executable) {
		return invalid("executable")
	}
	for _, a := range d.Arguments {
		if !literal(a) {
			return invalid("argv NUL")
		}
	}
	return nil
}

// AutoShell is an immutable semantic boundary value. Its zero is invalid.
type AutoShell struct {
	data        AutoShellData
	initialized bool
}

// AutoShellData is a construction/copy value; NewAutoShell validates and owns a copy.
type AutoShellData struct {
}

func NewAutoShell(d AutoShellData) (AutoShell, error) {
	if err := d.validate(); err != nil {
		return AutoShell{}, err
	}
	return AutoShell{data: d.clone(), initialized: true}, nil
}
func (v AutoShell) Valid() bool         { return v.initialized && v.data.validate() == nil }
func (v AutoShell) Data() AutoShellData { return v.data.clone() }
func (v AutoShell) Clone() AutoShell    { return v }
func (d AutoShellData) clone() AutoShellData {

	return d
}
func (d AutoShellData) validate() error {

	return nil
}

// ConfiguredShell is an immutable semantic boundary value. Its zero is invalid.
type ConfiguredShell struct {
	data        ConfiguredShellData
	initialized bool
}

// ConfiguredShellData is a construction/copy value; NewConfiguredShell validates and owns a copy.
type ConfiguredShellData struct {
	Execution ArgvExecution
}

func NewConfiguredShell(d ConfiguredShellData) (ConfiguredShell, error) {
	if err := d.validate(); err != nil {
		return ConfiguredShell{}, err
	}
	return ConfiguredShell{data: d.clone(), initialized: true}, nil
}
func (v ConfiguredShell) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v ConfiguredShell) Data() ConfiguredShellData { return v.data.clone() }
func (v ConfiguredShell) Clone() ConfiguredShell    { return v }
func (d ConfiguredShellData) clone() ConfiguredShellData {

	return d
}
func (d ConfiguredShellData) validate() error {
	if !d.Execution.Valid() {
		return invalid("d.Execution")
	}

	return nil
}

// InteractiveShell is an immutable semantic boundary value. Its zero is invalid.
type InteractiveShell struct {
	data        InteractiveShellData
	initialized bool
}

// InteractiveShellData is a construction/copy value; NewInteractiveShell validates and owns a copy.
type InteractiveShellData struct {
	Policy ShellPolicy
}

func NewInteractiveShell(d InteractiveShellData) (InteractiveShell, error) {
	if err := d.validate(); err != nil {
		return InteractiveShell{}, err
	}
	return InteractiveShell{data: d.clone(), initialized: true}, nil
}
func (v InteractiveShell) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v InteractiveShell) Data() InteractiveShellData { return v.data.clone() }
func (v InteractiveShell) Clone() InteractiveShell    { return v }
func (d InteractiveShellData) clone() InteractiveShellData {

	return d
}
func (d InteractiveShellData) validate() error {
	if !validShellPolicy(d.Policy) {
		return invalid("d.Policy")
	}

	return nil
}

// Invocation is an immutable semantic boundary value. Its zero is invalid.
type Invocation struct {
	data        InvocationData
	initialized bool
}

// InvocationData is a construction/copy value; NewInvocation validates and owns a copy.
type InvocationData struct {
	Execution   ExecutionIntent
	Environment EnvironmentPolicy
	Cwd         CwdObservation
	Terminal    TerminalMode
	Geometry    Geometry
	Label       string
}

func NewInvocation(d InvocationData) (Invocation, error) {
	if err := d.validate(); err != nil {
		return Invocation{}, err
	}
	return Invocation{data: d.clone(), initialized: true}, nil
}
func (v Invocation) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v Invocation) Data() InvocationData { return v.data.clone() }
func (v Invocation) Clone() Invocation    { return v }
func (d InvocationData) clone() InvocationData {

	return d
}
func (d InvocationData) validate() error {
	if !validExecutionIntent(d.Execution) {
		return invalid("d.Execution")
	}
	if !d.Environment.Valid() {
		return invalid("d.Environment")
	}
	if !d.Cwd.Valid() {
		return invalid("d.Cwd")
	}
	if !d.Terminal.Valid() {
		return invalid("d.Terminal")
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}

	if _, ok := d.Execution.(InteractiveShell); ok && d.Terminal != Terminal {
		return invalid("interactive terminal")
	}
	if !textValue(d.Label) {
		return invalid("label")
	}
	if d.Cwd.data.ProjectIdentity.Platform() == DirectoryWindows && !windowsEnvironment(d.Environment.data) {
		return invalid("Windows environment keys")
	}
	return nil
}

// InvocationSummary is an immutable semantic boundary value. Its zero is invalid.
type InvocationSummary struct {
	data        InvocationSummaryData
	initialized bool
}

// InvocationSummaryData is a construction/copy value; NewInvocationSummary validates and owns a copy.
type InvocationSummaryData struct {
	Label             string
	ExecutableDisplay string
	ArgumentDisplay   []string
	Cwd               CwdObservation
	AcceptedLocator   string
	Terminal          TerminalMode
	Geometry          Geometry
}

func NewInvocationSummary(d InvocationSummaryData) (InvocationSummary, error) {
	if err := d.validate(); err != nil {
		return InvocationSummary{}, err
	}
	return InvocationSummary{data: d.clone(), initialized: true}, nil
}
func (v InvocationSummary) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v InvocationSummary) Data() InvocationSummaryData { return v.data.clone() }
func (v InvocationSummary) Clone() InvocationSummary    { return v }
func (d InvocationSummaryData) clone() InvocationSummaryData {
	d.ArgumentDisplay = cloneSlice(d.ArgumentDisplay)
	return d
}
func (d InvocationSummaryData) validate() error {

	if !d.Cwd.Valid() {
		return invalid("d.Cwd")
	}

	if !d.Terminal.Valid() {
		return invalid("d.Terminal")
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}

	return nil
}

// DiagnosticContext is an immutable semantic boundary value. Its zero is invalid.
type DiagnosticContext struct {
	data        DiagnosticContextData
	initialized bool
}

// DiagnosticContextData is a construction/copy value; NewDiagnosticContext validates and owns a copy.
type DiagnosticContextData struct {
	Operation   Optional[OperationID]
	Repository  Optional[domain.RepositoryID]
	Worktree    Optional[domain.WorktreeID]
	Session     Optional[domain.SessionID]
	Observation Optional[ObservationID]
	Source      Optional[SourceVersion]
}

func NewDiagnosticContext(d DiagnosticContextData) (DiagnosticContext, error) {
	if err := d.validate(); err != nil {
		return DiagnosticContext{}, err
	}
	return DiagnosticContext{data: d.clone(), initialized: true}, nil
}
func (v DiagnosticContext) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v DiagnosticContext) Data() DiagnosticContextData { return v.data.clone() }
func (v DiagnosticContext) Clone() DiagnosticContext    { return v }
func (d DiagnosticContextData) clone() DiagnosticContextData {

	return d
}
func (d DiagnosticContextData) validate() error {
	if item, ok := d.Operation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Repository.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Session.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Observation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Source.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// NativeDiagnosticDetail is an immutable semantic boundary value. Its zero is invalid.
type NativeDiagnosticDetail struct {
	data        NativeDiagnosticDetailData
	initialized bool
}

// NativeDiagnosticDetailData is a construction/copy value; NewNativeDiagnosticDetail validates and owns a copy.
type NativeDiagnosticDetailData struct {
	ExitCode  Optional[int]
	Stderr    string
	Stage     string
	Truncated bool
}

func NewNativeDiagnosticDetail(d NativeDiagnosticDetailData) (NativeDiagnosticDetail, error) {
	if err := d.validate(); err != nil {
		return NativeDiagnosticDetail{}, err
	}
	return NativeDiagnosticDetail{data: d.clone(), initialized: true}, nil
}
func (v NativeDiagnosticDetail) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v NativeDiagnosticDetail) Data() NativeDiagnosticDetailData { return v.data.clone() }
func (v NativeDiagnosticDetail) Clone() NativeDiagnosticDetail    { return v }
func (d NativeDiagnosticDetailData) clone() NativeDiagnosticDetailData {

	return d
}
func (d NativeDiagnosticDetailData) validate() error {

	return nil
}

// RemoteDiagnosticDetail is an immutable semantic boundary value. Its zero is invalid.
type RemoteDiagnosticDetail struct {
	data        RemoteDiagnosticDetailData
	initialized bool
}

// RemoteDiagnosticDetailData is a construction/copy value; NewRemoteDiagnosticDetail validates and owns a copy.
type RemoteDiagnosticDetailData struct {
	HTTPStatus        Optional[uint32]
	RequestID         Optional[string]
	RateRemaining     Optional[uint64]
	RetryAfterSeconds Optional[uint64]
	ResetAt           Optional[time.Time]
	Repository        Optional[domain.RepositoryID]
}

func NewRemoteDiagnosticDetail(d RemoteDiagnosticDetailData) (RemoteDiagnosticDetail, error) {
	if err := d.validate(); err != nil {
		return RemoteDiagnosticDetail{}, err
	}
	return RemoteDiagnosticDetail{data: d.clone(), initialized: true}, nil
}
func (v RemoteDiagnosticDetail) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v RemoteDiagnosticDetail) Data() RemoteDiagnosticDetailData { return v.data.clone() }
func (v RemoteDiagnosticDetail) Clone() RemoteDiagnosticDetail    { return v }
func (d RemoteDiagnosticDetailData) clone() RemoteDiagnosticDetailData {

	return d
}
func (d RemoteDiagnosticDetailData) validate() error {

	if item, ok := d.Repository.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// IdentityDiagnosticDetail is an immutable semantic boundary value. Its zero is invalid.
type IdentityDiagnosticDetail struct {
	data        IdentityDiagnosticDetailData
	initialized bool
}

// IdentityDiagnosticDetailData is a construction/copy value; NewIdentityDiagnosticDetail validates and owns a copy.
type IdentityDiagnosticDetailData struct {
	ExpectedTarget   Optional[domain.ExactTarget]
	ObservedRevision Optional[domain.Revision]
	ExpectedHead     Optional[domain.Head]
	ObservedHead     Optional[domain.Head]
	RecoveryIDs      []RecoveryID
}

func NewIdentityDiagnosticDetail(d IdentityDiagnosticDetailData) (IdentityDiagnosticDetail, error) {
	if err := d.validate(); err != nil {
		return IdentityDiagnosticDetail{}, err
	}
	return IdentityDiagnosticDetail{data: d.clone(), initialized: true}, nil
}
func (v IdentityDiagnosticDetail) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v IdentityDiagnosticDetail) Data() IdentityDiagnosticDetailData { return v.data.clone() }
func (v IdentityDiagnosticDetail) Clone() IdentityDiagnosticDetail    { return v }
func (d IdentityDiagnosticDetailData) clone() IdentityDiagnosticDetailData {
	d.RecoveryIDs = cloneSlice(d.RecoveryIDs)
	return d
}
func (d IdentityDiagnosticDetailData) validate() error {
	if item, ok := d.ExpectedTarget.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ObservedRevision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ExpectedHead.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ObservedHead.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.RecoveryIDs {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// StorageDiagnosticDetail is an immutable semantic boundary value. Its zero is invalid.
type StorageDiagnosticDetail struct {
	data        StorageDiagnosticDetailData
	initialized bool
}

// StorageDiagnosticDetailData is a construction/copy value; NewStorageDiagnosticDetail validates and owns a copy.
type StorageDiagnosticDetailData struct {
	Family           StorageFamily
	Store            string
	Stage            string
	RawSchemaVersion Optional[string]
}

func NewStorageDiagnosticDetail(d StorageDiagnosticDetailData) (StorageDiagnosticDetail, error) {
	if err := d.validate(); err != nil {
		return StorageDiagnosticDetail{}, err
	}
	return StorageDiagnosticDetail{data: d.clone(), initialized: true}, nil
}
func (v StorageDiagnosticDetail) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v StorageDiagnosticDetail) Data() StorageDiagnosticDetailData { return v.data.clone() }
func (v StorageDiagnosticDetail) Clone() StorageDiagnosticDetail    { return v }
func (d StorageDiagnosticDetailData) clone() StorageDiagnosticDetailData {

	return d
}
func (d StorageDiagnosticDetailData) validate() error {
	if !d.Family.Valid() {
		return invalid("d.Family")
	}

	return nil
}

// Diagnostic is an immutable semantic boundary value. Its zero is invalid.
type Diagnostic struct {
	data        DiagnosticData
	initialized bool
}

// DiagnosticData is a construction/copy value; NewDiagnostic validates and owns a copy.
type DiagnosticData struct {
	Code    ErrorCode
	Reason  string
	Message string
	Context Optional[DiagnosticContext]
	Detail  Optional[DiagnosticDetail]
}

func NewDiagnostic(d DiagnosticData) (Diagnostic, error) {
	if err := d.validate(); err != nil {
		return Diagnostic{}, err
	}
	return Diagnostic{data: d.clone(), initialized: true}, nil
}
func (v Diagnostic) Valid() bool          { return v.initialized && v.data.validate() == nil }
func (v Diagnostic) Data() DiagnosticData { return v.data.clone() }
func (v Diagnostic) Clone() Diagnostic    { return v }
func (d DiagnosticData) clone() DiagnosticData {

	return d
}
func (d DiagnosticData) validate() error {
	if !d.Code.Valid() {
		return invalid("d.Code")
	}

	if item, ok := d.Context.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Detail.Value(); ok {
		if !validDiagnosticDetail(item) {
			return invalid("item")
		}
	}
	if !nonempty(d.Reason) || !textValue(d.Message) {
		return invalid("diagnostic")
	}
	return nil
}

// CommandTransportOutcome is an immutable semantic boundary value. Its zero is invalid.
type CommandTransportOutcome struct {
	data        CommandTransportOutcomeData
	initialized bool
}

// CommandTransportOutcomeData is a construction/copy value; NewCommandTransportOutcome validates and owns a copy.
type CommandTransportOutcomeData struct {
	Started               bool
	RootReaped            bool
	CleanupKnown          bool
	ExitCode              Optional[int]
	StdoutTruncated       bool
	StderrTruncated       bool
	CancellationRequested bool
	Diagnostics           []Diagnostic
}

func NewCommandTransportOutcome(d CommandTransportOutcomeData) (CommandTransportOutcome, error) {
	if err := d.validate(); err != nil {
		return CommandTransportOutcome{}, err
	}
	return CommandTransportOutcome{data: d.clone(), initialized: true}, nil
}
func (v CommandTransportOutcome) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v CommandTransportOutcome) Data() CommandTransportOutcomeData { return v.data.clone() }
func (v CommandTransportOutcome) Clone() CommandTransportOutcome    { return v }
func (d CommandTransportOutcomeData) clone() CommandTransportOutcomeData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d CommandTransportOutcomeData) validate() error {

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Started && (d.RootReaped || d.ExitCode.Present()) {
		return invalid("unstarted transport")
	}
	return nil
}

// FacetEffect is an immutable semantic boundary value. Its zero is invalid.
type FacetEffect struct {
	data        FacetEffectData
	initialized bool
}

// FacetEffectData is a construction/copy value; NewFacetEffect validates and owns a copy.
type FacetEffectData struct {
	Facet           EffectFacet
	State           EffectState
	PostObservation Optional[ObservationID]
	RecoveryIDs     []RecoveryID
}

func NewFacetEffect(d FacetEffectData) (FacetEffect, error) {
	if err := d.validate(); err != nil {
		return FacetEffect{}, err
	}
	return FacetEffect{data: d.clone(), initialized: true}, nil
}
func (v FacetEffect) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v FacetEffect) Data() FacetEffectData { return v.data.clone() }
func (v FacetEffect) Clone() FacetEffect    { return v }
func (d FacetEffectData) clone() FacetEffectData {
	d.RecoveryIDs = cloneSlice(d.RecoveryIDs)
	return d
}
func (d FacetEffectData) validate() error {
	if !d.Facet.Valid() {
		return invalid("d.Facet")
	}
	if !d.State.Valid() {
		return invalid("d.State")
	}
	if item, ok := d.PostObservation.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.RecoveryIDs {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if duplicateRecoveryIDs(d.RecoveryIDs) {
		return invalid("duplicate recovery ID")
	}
	return nil
}

// EffectReport is an immutable semantic boundary value. Its zero is invalid.
type EffectReport struct {
	data        EffectReportData
	initialized bool
}

// EffectReportData is a construction/copy value; NewEffectReport validates and owns a copy.
type EffectReportData struct {
	Facets []FacetEffect
}

func NewEffectReport(d EffectReportData) (EffectReport, error) {
	if err := d.validate(); err != nil {
		return EffectReport{}, err
	}
	return EffectReport{data: d.clone(), initialized: true}, nil
}
func (v EffectReport) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v EffectReport) Data() EffectReportData { return v.data.clone() }
func (v EffectReport) Clone() EffectReport    { return v }
func (d EffectReportData) clone() EffectReportData {
	d.Facets = cloneSlice(d.Facets)
	return d
}
func (d EffectReportData) validate() error {
	for _, item := range d.Facets {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// SourceRecoveryVersion is an immutable semantic boundary value. Its zero is invalid.
type SourceRecoveryVersion struct {
	data        SourceRecoveryVersionData
	initialized bool
}

// SourceRecoveryVersionData is a construction/copy value; NewSourceRecoveryVersion validates and owns a copy.
type SourceRecoveryVersionData struct {
	Version SourceVersion
}

func NewSourceRecoveryVersion(d SourceRecoveryVersionData) (SourceRecoveryVersion, error) {
	if err := d.validate(); err != nil {
		return SourceRecoveryVersion{}, err
	}
	return SourceRecoveryVersion{data: d.clone(), initialized: true}, nil
}
func (v SourceRecoveryVersion) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v SourceRecoveryVersion) Data() SourceRecoveryVersionData { return v.data.clone() }
func (v SourceRecoveryVersion) Clone() SourceRecoveryVersion    { return v }
func (d SourceRecoveryVersionData) clone() SourceRecoveryVersionData {

	return d
}
func (d SourceRecoveryVersionData) validate() error {
	if !d.Version.Valid() {
		return invalid("d.Version")
	}

	return nil
}

// StorageRecoveryVersion is an immutable semantic boundary value. Its zero is invalid.
type StorageRecoveryVersion struct {
	data        StorageRecoveryVersionData
	initialized bool
}

// StorageRecoveryVersionData is a construction/copy value; NewStorageRecoveryVersion validates and owns a copy.
type StorageRecoveryVersionData struct {
	Version StorageVersion
}

func NewStorageRecoveryVersion(d StorageRecoveryVersionData) (StorageRecoveryVersion, error) {
	if err := d.validate(); err != nil {
		return StorageRecoveryVersion{}, err
	}
	return StorageRecoveryVersion{data: d.clone(), initialized: true}, nil
}
func (v StorageRecoveryVersion) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v StorageRecoveryVersion) Data() StorageRecoveryVersionData { return v.data.clone() }
func (v StorageRecoveryVersion) Clone() StorageRecoveryVersion    { return v }
func (d StorageRecoveryVersionData) clone() StorageRecoveryVersionData {

	return d
}
func (d StorageRecoveryVersionData) validate() error {
	if !d.Version.Valid() {
		return invalid("d.Version")
	}

	return nil
}

// RecoverySubject is an immutable semantic boundary value. Its zero is invalid.
type RecoverySubject struct {
	data        RecoverySubjectData
	initialized bool
}

// RecoverySubjectData is a construction/copy value; NewRecoverySubject validates and owns a copy.
type RecoverySubjectData struct {
	Repository Optional[domain.RepositoryID]
	Worktree   Optional[domain.WorktreeID]
	Stash      Optional[domain.StashID]
	Revision   Optional[domain.Revision]
	Branch     Optional[domain.BranchID]
	Session    Optional[domain.SessionID]
	Store      Optional[string]
	Family     Optional[StorageFamily]
}

func NewRecoverySubject(d RecoverySubjectData) (RecoverySubject, error) {
	if err := d.validate(); err != nil {
		return RecoverySubject{}, err
	}
	return RecoverySubject{data: d.clone(), initialized: true}, nil
}
func (v RecoverySubject) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v RecoverySubject) Data() RecoverySubjectData { return v.data.clone() }
func (v RecoverySubject) Clone() RecoverySubject    { return v }
func (d RecoverySubjectData) clone() RecoverySubjectData {

	return d
}
func (d RecoverySubjectData) validate() error {
	if item, ok := d.Repository.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Worktree.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Stash.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Revision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Branch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Session.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if item, ok := d.Family.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !validRecoverySubject(d) {
		return invalid("recovery subject")
	}
	return nil
}

// RecoveryRecord is an immutable semantic boundary value. Its zero is invalid.
type RecoveryRecord struct {
	data        RecoveryRecordData
	initialized bool
}

// RecoveryRecordData is a construction/copy value; NewRecoveryRecord validates and owns a copy.
type RecoveryRecordData struct {
	RecoveryID RecoveryID
	Kind       RecoveryKind
	Layer      ResponsibleLayer
	Subject    RecoverySubject
	Locator    string
	Original   Optional[RecoveryVersion]
	Proposed   Optional[RecoveryVersion]
	NextAction string
}

func NewRecoveryRecord(d RecoveryRecordData) (RecoveryRecord, error) {
	if err := d.validate(); err != nil {
		return RecoveryRecord{}, err
	}
	return RecoveryRecord{data: d.clone(), initialized: true}, nil
}
func (v RecoveryRecord) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v RecoveryRecord) Data() RecoveryRecordData { return v.data.clone() }
func (v RecoveryRecord) Clone() RecoveryRecord    { return v }
func (d RecoveryRecordData) clone() RecoveryRecordData {

	return d
}
func (d RecoveryRecordData) validate() error {
	if !d.RecoveryID.Valid() {
		return invalid("d.RecoveryID")
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.Layer.Valid() {
		return invalid("d.Layer")
	}
	if !d.Subject.Valid() {
		return invalid("d.Subject")
	}

	if item, ok := d.Original.Value(); ok {
		if !validRecoveryVersion(item) {
			return invalid("item")
		}
	}
	if item, ok := d.Proposed.Value(); ok {
		if !validRecoveryVersion(item) {
			return invalid("item")
		}
	}

	if err := consistentRecoveryRecord(d); err != nil {
		return err
	}
	if !nonempty(d.Locator) || !textValue(d.NextAction) {
		return invalid("recovery locator/action")
	}
	return nil
}
