package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"time"
)

// RemoteRepositoryLocator is an immutable semantic boundary value. Its zero is invalid.
type RemoteRepositoryLocator struct {
	data        RemoteRepositoryLocatorData
	initialized bool
}

// RemoteRepositoryLocatorData is a construction/copy value; NewRemoteRepositoryLocator validates and owns a copy.
type RemoteRepositoryLocatorData struct {
	Host  string
	Owner string
	Name  string
}

func NewRemoteRepositoryLocator(d RemoteRepositoryLocatorData) (RemoteRepositoryLocator, error) {
	if err := d.validate(); err != nil {
		return RemoteRepositoryLocator{}, err
	}
	return RemoteRepositoryLocator{data: d.clone(), initialized: true}, nil
}
func (v RemoteRepositoryLocator) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v RemoteRepositoryLocator) Data() RemoteRepositoryLocatorData { return v.data.clone() }
func (v RemoteRepositoryLocator) Clone() RemoteRepositoryLocator    { return v }
func (d RemoteRepositoryLocatorData) clone() RemoteRepositoryLocatorData {

	return d
}
func (d RemoteRepositoryLocatorData) validate() error {

	if !validRemoteLocator(d) {
		return invalid("remote locator")
	}
	return nil
}

// RemoteCapabilities is an immutable semantic boundary value. Its zero is invalid.
type RemoteCapabilities struct {
	data        RemoteCapabilitiesData
	initialized bool
}

// RemoteCapabilitiesData is a construction/copy value; NewRemoteCapabilities validates and owns a copy.
type RemoteCapabilitiesData struct {
	ReadBranches           bool
	ReadPullRequests       bool
	CreatePullRequest      bool
	SupportedObjectFormats []domain.ObjectFormat
	Diagnostics            []Diagnostic
}

func NewRemoteCapabilities(d RemoteCapabilitiesData) (RemoteCapabilities, error) {
	if err := d.validate(); err != nil {
		return RemoteCapabilities{}, err
	}
	return RemoteCapabilities{data: d.clone(), initialized: true}, nil
}
func (v RemoteCapabilities) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v RemoteCapabilities) Data() RemoteCapabilitiesData { return v.data.clone() }
func (v RemoteCapabilities) Clone() RemoteCapabilities    { return v }
func (d RemoteCapabilitiesData) clone() RemoteCapabilitiesData {
	d.SupportedObjectFormats = cloneSlice(d.SupportedObjectFormats)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d RemoteCapabilitiesData) validate() error {

	for _, item := range d.SupportedObjectFormats {
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

// RemoteRepository is an immutable semantic boundary value. Its zero is invalid.
type RemoteRepository struct {
	data        RemoteRepositoryData
	initialized bool
}

// RemoteRepositoryData is a construction/copy value; NewRemoteRepository validates and owns a copy.
type RemoteRepositoryData struct {
	ID            domain.RepositoryID
	Locator       RemoteRepositoryLocator
	URL           string
	DefaultBranch Optional[domain.BranchID]
	Capabilities  RemoteCapabilities
}

func NewRemoteRepository(d RemoteRepositoryData) (RemoteRepository, error) {
	if err := d.validate(); err != nil {
		return RemoteRepository{}, err
	}
	return RemoteRepository{data: d.clone(), initialized: true}, nil
}
func (v RemoteRepository) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v RemoteRepository) Data() RemoteRepositoryData { return v.data.clone() }
func (v RemoteRepository) Clone() RemoteRepository    { return v }
func (d RemoteRepositoryData) clone() RemoteRepositoryData {

	return d
}
func (d RemoteRepositoryData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
	}
	if !d.Locator.Valid() {
		return invalid("d.Locator")
	}

	if item, ok := d.DefaultBranch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Capabilities.Valid() {
		return invalid("d.Capabilities")
	}
	if d.ID.Scope() != domain.Remote || !remoteURL(d.URL, d.Locator, 0) {
		return invalid("remote repository")
	}
	if b, ok := d.DefaultBranch.Value(); ok && b.Repository() != d.ID {
		return invalid("default branch scope")
	}
	return nil
}

// LiveRemoteObservation is an immutable semantic boundary value. Its zero is invalid.
type LiveRemoteObservation struct {
	data        LiveRemoteObservationData
	initialized bool
}

// LiveRemoteObservationData is a construction/copy value; NewLiveRemoteObservation validates and owns a copy.
type LiveRemoteObservationData struct {
}

func NewLiveRemoteObservation(d LiveRemoteObservationData) (LiveRemoteObservation, error) {
	if err := d.validate(); err != nil {
		return LiveRemoteObservation{}, err
	}
	return LiveRemoteObservation{data: d.clone(), initialized: true}, nil
}
func (v LiveRemoteObservation) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v LiveRemoteObservation) Data() LiveRemoteObservationData { return v.data.clone() }
func (v LiveRemoteObservation) Clone() LiveRemoteObservation    { return v }
func (d LiveRemoteObservationData) clone() LiveRemoteObservationData {

	return d
}
func (d LiveRemoteObservationData) validate() error {

	return nil
}

// CachedRemoteObservation is an immutable semantic boundary value. Its zero is invalid.
type CachedRemoteObservation struct {
	data        CachedRemoteObservationData
	initialized bool
}

// CachedRemoteObservationData is a construction/copy value; NewCachedRemoteObservation validates and owns a copy.
type CachedRemoteObservationData struct {
	OriginalID       ObservationID
	OriginalInterval ObservationInterval
}

func NewCachedRemoteObservation(d CachedRemoteObservationData) (CachedRemoteObservation, error) {
	if err := d.validate(); err != nil {
		return CachedRemoteObservation{}, err
	}
	return CachedRemoteObservation{data: d.clone(), initialized: true}, nil
}
func (v CachedRemoteObservation) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v CachedRemoteObservation) Data() CachedRemoteObservationData { return v.data.clone() }
func (v CachedRemoteObservation) Clone() CachedRemoteObservation    { return v }
func (d CachedRemoteObservationData) clone() CachedRemoteObservationData {

	return d
}
func (d CachedRemoteObservationData) validate() error {
	if !d.OriginalID.Valid() {
		return invalid("d.OriginalID")
	}
	if !d.OriginalInterval.Valid() {
		return invalid("d.OriginalInterval")
	}

	return nil
}

// RemoteObservation is an immutable semantic boundary value. Its zero is invalid.
type RemoteObservation struct {
	data        RemoteObservationData
	initialized bool
}

// RemoteObservationData is a construction/copy value; NewRemoteObservation validates and owns a copy.
type RemoteObservationData struct {
	ID         ObservationID
	Repository domain.RepositoryID
	Interval   ObservationInterval
	Version    SourceVersion
	Origin     RemoteObservationOrigin
	Page       PageInfo
}

func NewRemoteObservation(d RemoteObservationData) (RemoteObservation, error) {
	if err := d.validate(); err != nil {
		return RemoteObservation{}, err
	}
	return RemoteObservation{data: d.clone(), initialized: true}, nil
}
func (v RemoteObservation) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v RemoteObservation) Data() RemoteObservationData { return v.data.clone() }
func (v RemoteObservation) Clone() RemoteObservation    { return v }
func (d RemoteObservationData) clone() RemoteObservationData {

	return d
}
func (d RemoteObservationData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
	}
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Interval.Valid() {
		return invalid("d.Interval")
	}
	if !d.Version.Valid() {
		return invalid("d.Version")
	}
	if !validRemoteObservationOrigin(d.Origin) {
		return invalid("d.Origin")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.Remote || d.Page.data.Source != d.Version {
		return invalid("remote observation")
	}
	return nil
}

// RemoteBranchFact is an immutable semantic boundary value. Its zero is invalid.
type RemoteBranchFact struct {
	data        RemoteBranchFactData
	initialized bool
}

// RemoteBranchFactData is a construction/copy value; NewRemoteBranchFact validates and owns a copy.
type RemoteBranchFactData struct {
	Branch      domain.BranchID
	Tip         domain.Revision
	Observation RemoteObservation
}

func NewRemoteBranchFact(d RemoteBranchFactData) (RemoteBranchFact, error) {
	if err := d.validate(); err != nil {
		return RemoteBranchFact{}, err
	}
	return RemoteBranchFact{data: d.clone(), initialized: true}, nil
}
func (v RemoteBranchFact) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v RemoteBranchFact) Data() RemoteBranchFactData { return v.data.clone() }
func (v RemoteBranchFact) Clone() RemoteBranchFact    { return v }
func (d RemoteBranchFactData) clone() RemoteBranchFactData {

	return d
}
func (d RemoteBranchFactData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.Tip.Valid() {
		return invalid("d.Tip")
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Branch.Kind() != domain.RemoteHead || d.Branch.Repository() != d.Tip.Repository() || d.Tip.Repository() != d.Observation.data.Repository {
		return invalid("remote branch scope")
	}
	return nil
}

// AvailableEndpoint is an immutable semantic boundary value. Its zero is invalid.
type AvailableEndpoint struct {
	data        AvailableEndpointData
	initialized bool
}

// AvailableEndpointData is a construction/copy value; NewAvailableEndpoint validates and owns a copy.
type AvailableEndpointData struct {
	Repository RemoteRepository
	Branch     domain.BranchID
	Revision   domain.Revision
}

func NewAvailableEndpoint(d AvailableEndpointData) (AvailableEndpoint, error) {
	if err := d.validate(); err != nil {
		return AvailableEndpoint{}, err
	}
	return AvailableEndpoint{data: d.clone(), initialized: true}, nil
}
func (v AvailableEndpoint) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v AvailableEndpoint) Data() AvailableEndpointData { return v.data.clone() }
func (v AvailableEndpoint) Clone() AvailableEndpoint    { return v }
func (d AvailableEndpointData) clone() AvailableEndpointData {

	return d
}
func (d AvailableEndpointData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	if d.Branch.Repository() != d.Repository.data.ID || d.Revision.Repository() != d.Repository.data.ID {
		return invalid("endpoint scope")
	}
	return nil
}

// UnavailableEndpoint is an immutable semantic boundary value. Its zero is invalid.
type UnavailableEndpoint struct {
	data        UnavailableEndpointData
	initialized bool
}

// UnavailableEndpointData is a construction/copy value; NewUnavailableEndpoint validates and owns a copy.
type UnavailableEndpointData struct {
	KnownRepository Optional[domain.RepositoryID]
	KnownBranch     Optional[domain.BranchID]
	KnownRevision   Optional[domain.Revision]
	Reason          EndpointUnavailableReason
}

func NewUnavailableEndpoint(d UnavailableEndpointData) (UnavailableEndpoint, error) {
	if err := d.validate(); err != nil {
		return UnavailableEndpoint{}, err
	}
	return UnavailableEndpoint{data: d.clone(), initialized: true}, nil
}
func (v UnavailableEndpoint) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v UnavailableEndpoint) Data() UnavailableEndpointData { return v.data.clone() }
func (v UnavailableEndpoint) Clone() UnavailableEndpoint    { return v }
func (d UnavailableEndpointData) clone() UnavailableEndpointData {

	return d
}
func (d UnavailableEndpointData) validate() error {
	if item, ok := d.KnownRepository.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.KnownBranch.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.KnownRevision.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}
	return validateUnavailableEndpoint(d)
}

// PullRequestFact is an immutable semantic boundary value. Its zero is invalid.
type PullRequestFact struct {
	data        PullRequestFactData
	initialized bool
}

// PullRequestFactData is a construction/copy value; NewPullRequestFact validates and owns a copy.
type PullRequestFactData struct {
	ID                  domain.PRID
	URL                 string
	Title               string
	Body                StoredField[string]
	State               PullRequestState
	Draft               bool
	MaintainerCanModify Optional[bool]
	Base                PullRequestEndpoint
	Head                PullRequestEndpoint
	UpdatedAt           Optional[time.Time]
	Observation         RemoteObservation
	Diagnostics         []Diagnostic
}

func NewPullRequestFact(d PullRequestFactData) (PullRequestFact, error) {
	if err := d.validate(); err != nil {
		return PullRequestFact{}, err
	}
	return PullRequestFact{data: d.clone(), initialized: true}, nil
}
func (v PullRequestFact) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v PullRequestFact) Data() PullRequestFactData { return v.data.clone() }
func (v PullRequestFact) Clone() PullRequestFact    { return v }
func (d PullRequestFactData) clone() PullRequestFactData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d PullRequestFactData) validate() error {
	if !d.ID.Valid() {
		return invalid("d.ID")
	}

	if !d.Body.Valid() {
		return invalid("field presence")
	}
	if !d.State.Valid() {
		return invalid("d.State")
	}

	if !validPullRequestEndpoint(d.Base) {
		return invalid("d.Base")
	}
	if !validPullRequestEndpoint(d.Head) {
		return invalid("d.Head")
	}

	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.ID.Repository() != d.Observation.data.Repository || d.Body.Presence() == FieldNull || (d.State == PRUnknown && len(d.Diagnostics) == 0) {
		return invalid("PR fact")
	}
	if r, ok := endpointRepository(d.Base); ok && r != d.ID.Repository() {
		return invalid("PR base scope")
	}
	return nil
}

// AllRemoteBranches is an immutable semantic boundary value. Its zero is invalid.
type AllRemoteBranches struct {
	data        AllRemoteBranchesData
	initialized bool
}

// AllRemoteBranchesData is a construction/copy value; NewAllRemoteBranches validates and owns a copy.
type AllRemoteBranchesData struct {
}

func NewAllRemoteBranches(d AllRemoteBranchesData) (AllRemoteBranches, error) {
	if err := d.validate(); err != nil {
		return AllRemoteBranches{}, err
	}
	return AllRemoteBranches{data: d.clone(), initialized: true}, nil
}
func (v AllRemoteBranches) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v AllRemoteBranches) Data() AllRemoteBranchesData { return v.data.clone() }
func (v AllRemoteBranches) Clone() AllRemoteBranches    { return v }
func (d AllRemoteBranchesData) clone() AllRemoteBranchesData {

	return d
}
func (d AllRemoteBranchesData) validate() error {

	return nil
}

// RemoteBranchPrefix is an immutable semantic boundary value. Its zero is invalid.
type RemoteBranchPrefix struct {
	data        RemoteBranchPrefixData
	initialized bool
}

// RemoteBranchPrefixData is a construction/copy value; NewRemoteBranchPrefix validates and owns a copy.
type RemoteBranchPrefixData struct {
	Prefix string
}

func NewRemoteBranchPrefix(d RemoteBranchPrefixData) (RemoteBranchPrefix, error) {
	if err := d.validate(); err != nil {
		return RemoteBranchPrefix{}, err
	}
	return RemoteBranchPrefix{data: d.clone(), initialized: true}, nil
}
func (v RemoteBranchPrefix) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v RemoteBranchPrefix) Data() RemoteBranchPrefixData { return v.data.clone() }
func (v RemoteBranchPrefix) Clone() RemoteBranchPrefix    { return v }
func (d RemoteBranchPrefixData) clone() RemoteBranchPrefixData {

	return d
}
func (d RemoteBranchPrefixData) validate() error {

	if !nonempty(d.Prefix) {
		return invalid("branch prefix")
	}
	return nil
}

// PullRequestFilter is an immutable semantic boundary value. Its zero is invalid.
type PullRequestFilter struct {
	data        PullRequestFilterData
	initialized bool
}

// PullRequestFilterData is a construction/copy value; NewPullRequestFilter validates and owns a copy.
type PullRequestFilterData struct {
	State PullRequestFilterState
	Head  Optional[domain.BranchID]
	Base  Optional[domain.BranchID]
}

func NewPullRequestFilter(d PullRequestFilterData) (PullRequestFilter, error) {
	if err := d.validate(); err != nil {
		return PullRequestFilter{}, err
	}
	return PullRequestFilter{data: d.clone(), initialized: true}, nil
}
func (v PullRequestFilter) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v PullRequestFilter) Data() PullRequestFilterData { return v.data.clone() }
func (v PullRequestFilter) Clone() PullRequestFilter    { return v }
func (d PullRequestFilterData) clone() PullRequestFilterData {

	return d
}
func (d PullRequestFilterData) validate() error {
	if !d.State.Valid() {
		return invalid("d.State")
	}
	if item, ok := d.Head.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.Base.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, v := range []Optional[domain.BranchID]{d.Head, d.Base} {
		if b, ok := v.Value(); ok && b.Kind() != domain.RemoteHead {
			return invalid("PR filter scope")
		}
	}
	return nil
}

// EndpointExpectation is an immutable semantic boundary value. Its zero is invalid.
type EndpointExpectation struct {
	data        EndpointExpectationData
	initialized bool
}

// EndpointExpectationData is a construction/copy value; NewEndpointExpectation validates and owns a copy.
type EndpointExpectationData struct {
	Branch      domain.BranchID
	Revision    domain.Revision
	Observation RemoteObservation
}

func NewEndpointExpectation(d EndpointExpectationData) (EndpointExpectation, error) {
	if err := d.validate(); err != nil {
		return EndpointExpectation{}, err
	}
	return EndpointExpectation{data: d.clone(), initialized: true}, nil
}
func (v EndpointExpectation) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v EndpointExpectation) Data() EndpointExpectationData { return v.data.clone() }
func (v EndpointExpectation) Clone() EndpointExpectation    { return v }
func (d EndpointExpectationData) clone() EndpointExpectationData {

	return d
}
func (d EndpointExpectationData) validate() error {
	if !d.Branch.Valid() {
		return invalid("d.Branch")
	}
	if !d.Revision.Valid() {
		return invalid("d.Revision")
	}
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}
	if d.Branch.Kind() != domain.RemoteHead || d.Branch.Repository() != d.Revision.Repository() || d.Branch.Repository() != d.Observation.data.Repository {
		return invalid("endpoint expectation")
	}
	return nil
}

// ResolveRepositoryRequest is an immutable semantic boundary value. Its zero is invalid.
type ResolveRepositoryRequest struct {
	data        ResolveRepositoryRequestData
	initialized bool
}

// ResolveRepositoryRequestData is a construction/copy value; NewResolveRepositoryRequest validates and owns a copy.
type ResolveRepositoryRequestData struct {
	Locator RemoteRepositoryLocator
}

func NewResolveRepositoryRequest(d ResolveRepositoryRequestData) (ResolveRepositoryRequest, error) {
	if err := d.validate(); err != nil {
		return ResolveRepositoryRequest{}, err
	}
	return ResolveRepositoryRequest{data: d.clone(), initialized: true}, nil
}
func (v ResolveRepositoryRequest) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v ResolveRepositoryRequest) Data() ResolveRepositoryRequestData { return v.data.clone() }
func (v ResolveRepositoryRequest) Clone() ResolveRepositoryRequest    { return v }
func (d ResolveRepositoryRequestData) clone() ResolveRepositoryRequestData {

	return d
}
func (d ResolveRepositoryRequestData) validate() error {
	if !d.Locator.Valid() {
		return invalid("d.Locator")
	}

	return nil
}

// ResolveRepositoryResult is an immutable semantic boundary value. Its zero is invalid.
type ResolveRepositoryResult struct {
	data        ResolveRepositoryResultData
	initialized bool
}

// ResolveRepositoryResultData is a construction/copy value; NewResolveRepositoryResult validates and owns a copy.
type ResolveRepositoryResultData struct {
	Repository  Optional[RemoteRepository]
	Observation Optional[RemoteObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewResolveRepositoryResult(d ResolveRepositoryResultData) (ResolveRepositoryResult, error) {
	if err := d.validate(); err != nil {
		return ResolveRepositoryResult{}, err
	}
	return ResolveRepositoryResult{data: d.clone(), initialized: true}, nil
}
func (v ResolveRepositoryResult) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v ResolveRepositoryResult) Data() ResolveRepositoryResultData { return v.data.clone() }
func (v ResolveRepositoryResult) Clone() ResolveRepositoryResult    { return v }
func (d ResolveRepositoryResultData) clone() ResolveRepositoryResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ResolveRepositoryResultData) validate() error {
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

// ListBranchesRequest is an immutable semantic boundary value. Its zero is invalid.
type ListBranchesRequest struct {
	data        ListBranchesRequestData
	initialized bool
}

// ListBranchesRequestData is a construction/copy value; NewListBranchesRequest validates and owns a copy.
type ListBranchesRequestData struct {
	Repository domain.RepositoryID
	Filter     RemoteBranchFilter
	Page       PageRequest
}

func NewListBranchesRequest(d ListBranchesRequestData) (ListBranchesRequest, error) {
	if err := d.validate(); err != nil {
		return ListBranchesRequest{}, err
	}
	return ListBranchesRequest{data: d.clone(), initialized: true}, nil
}
func (v ListBranchesRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v ListBranchesRequest) Data() ListBranchesRequestData { return v.data.clone() }
func (v ListBranchesRequest) Clone() ListBranchesRequest    { return v }
func (d ListBranchesRequestData) clone() ListBranchesRequestData {

	return d
}
func (d ListBranchesRequestData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !validRemoteBranchFilter(d.Filter) {
		return invalid("d.Filter")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.Remote || !remotePage(d.Page) {
		return invalid("remote page scope")
	}
	return nil
}

// ListBranchesResult is an immutable semantic boundary value. Its zero is invalid.
type ListBranchesResult struct {
	data        ListBranchesResultData
	initialized bool
}

// ListBranchesResultData is a construction/copy value; NewListBranchesResult validates and owns a copy.
type ListBranchesResultData struct {
	Branches    []RemoteBranchFact
	Page        PageInfo
	Observation Optional[RemoteObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewListBranchesResult(d ListBranchesResultData) (ListBranchesResult, error) {
	if err := d.validate(); err != nil {
		return ListBranchesResult{}, err
	}
	return ListBranchesResult{data: d.clone(), initialized: true}, nil
}
func (v ListBranchesResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v ListBranchesResult) Data() ListBranchesResultData { return v.data.clone() }
func (v ListBranchesResult) Clone() ListBranchesResult    { return v }
func (d ListBranchesResultData) clone() ListBranchesResultData {
	d.Branches = cloneSlice(d.Branches)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ListBranchesResultData) validate() error {
	for _, item := range d.Branches {
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

// ListPullRequestsRequest is an immutable semantic boundary value. Its zero is invalid.
type ListPullRequestsRequest struct {
	data        ListPullRequestsRequestData
	initialized bool
}

// ListPullRequestsRequestData is a construction/copy value; NewListPullRequestsRequest validates and owns a copy.
type ListPullRequestsRequestData struct {
	Repository domain.RepositoryID
	Filter     PullRequestFilter
	Page       PageRequest
}

func NewListPullRequestsRequest(d ListPullRequestsRequestData) (ListPullRequestsRequest, error) {
	if err := d.validate(); err != nil {
		return ListPullRequestsRequest{}, err
	}
	return ListPullRequestsRequest{data: d.clone(), initialized: true}, nil
}
func (v ListPullRequestsRequest) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v ListPullRequestsRequest) Data() ListPullRequestsRequestData { return v.data.clone() }
func (v ListPullRequestsRequest) Clone() ListPullRequestsRequest    { return v }
func (d ListPullRequestsRequestData) clone() ListPullRequestsRequestData {

	return d
}
func (d ListPullRequestsRequestData) validate() error {
	if !d.Repository.Valid() {
		return invalid("d.Repository")
	}
	if !d.Filter.Valid() {
		return invalid("d.Filter")
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if d.Repository.Scope() != domain.Remote || !remotePage(d.Page) {
		return invalid("PR page scope")
	}
	if b, ok := d.Filter.data.Base.Value(); ok && b.Repository() != d.Repository {
		return invalid("PR base filter")
	}
	return nil
}

// ListPullRequestsResult is an immutable semantic boundary value. Its zero is invalid.
type ListPullRequestsResult struct {
	data        ListPullRequestsResultData
	initialized bool
}

// ListPullRequestsResultData is a construction/copy value; NewListPullRequestsResult validates and owns a copy.
type ListPullRequestsResultData struct {
	PullRequests []PullRequestFact
	Page         PageInfo
	Observation  Optional[RemoteObservation]
	Diagnostics  []Diagnostic
	Transport    CommandTransportOutcome
}

func NewListPullRequestsResult(d ListPullRequestsResultData) (ListPullRequestsResult, error) {
	if err := d.validate(); err != nil {
		return ListPullRequestsResult{}, err
	}
	return ListPullRequestsResult{data: d.clone(), initialized: true}, nil
}
func (v ListPullRequestsResult) Valid() bool                      { return v.initialized && v.data.validate() == nil }
func (v ListPullRequestsResult) Data() ListPullRequestsResultData { return v.data.clone() }
func (v ListPullRequestsResult) Clone() ListPullRequestsResult    { return v }
func (d ListPullRequestsResultData) clone() ListPullRequestsResultData {
	d.PullRequests = cloneSlice(d.PullRequests)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ListPullRequestsResultData) validate() error {
	for _, item := range d.PullRequests {
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

// ObservePullRequestRequest is an immutable semantic boundary value. Its zero is invalid.
type ObservePullRequestRequest struct {
	data        ObservePullRequestRequestData
	initialized bool
}

// ObservePullRequestRequestData is a construction/copy value; NewObservePullRequestRequest validates and owns a copy.
type ObservePullRequestRequestData struct {
	Target       domain.PRID
	ExpectedHead Optional[domain.Revision]
	ExpectedBase Optional[domain.Revision]
}

func NewObservePullRequestRequest(d ObservePullRequestRequestData) (ObservePullRequestRequest, error) {
	if err := d.validate(); err != nil {
		return ObservePullRequestRequest{}, err
	}
	return ObservePullRequestRequest{data: d.clone(), initialized: true}, nil
}
func (v ObservePullRequestRequest) Valid() bool                         { return v.initialized && v.data.validate() == nil }
func (v ObservePullRequestRequest) Data() ObservePullRequestRequestData { return v.data.clone() }
func (v ObservePullRequestRequest) Clone() ObservePullRequestRequest    { return v }
func (d ObservePullRequestRequestData) clone() ObservePullRequestRequestData {

	return d
}
func (d ObservePullRequestRequestData) validate() error {
	if !d.Target.Valid() {
		return invalid("d.Target")
	}
	if item, ok := d.ExpectedHead.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if item, ok := d.ExpectedBase.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if r, ok := d.ExpectedHead.Value(); ok && r.Repository().Scope() != domain.Remote {
		return invalid("expected head")
	}
	if r, ok := d.ExpectedBase.Value(); ok && r.Repository() != d.Target.Repository() {
		return invalid("expected base")
	}
	return nil
}

// ObservePullRequestResult is an immutable semantic boundary value. Its zero is invalid.
type ObservePullRequestResult struct {
	data        ObservePullRequestResultData
	initialized bool
}

// ObservePullRequestResultData is a construction/copy value; NewObservePullRequestResult validates and owns a copy.
type ObservePullRequestResultData struct {
	PullRequest Optional[PullRequestFact]
	Expectation ExpectationResult
	Observation Optional[RemoteObservation]
	Diagnostics []Diagnostic
	Transport   CommandTransportOutcome
}

func NewObservePullRequestResult(d ObservePullRequestResultData) (ObservePullRequestResult, error) {
	if err := d.validate(); err != nil {
		return ObservePullRequestResult{}, err
	}
	return ObservePullRequestResult{data: d.clone(), initialized: true}, nil
}
func (v ObservePullRequestResult) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v ObservePullRequestResult) Data() ObservePullRequestResultData { return v.data.clone() }
func (v ObservePullRequestResult) Clone() ObservePullRequestResult    { return v }
func (d ObservePullRequestResultData) clone() ObservePullRequestResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d ObservePullRequestResultData) validate() error {
	if item, ok := d.PullRequest.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Expectation.Valid() {
		return invalid("d.Expectation")
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

// CreatePullRequestRequest is an immutable semantic boundary value. Its zero is invalid.
type CreatePullRequestRequest struct {
	data        CreatePullRequestRequestData
	initialized bool
}

// CreatePullRequestRequestData is a construction/copy value; NewCreatePullRequestRequest validates and owns a copy.
type CreatePullRequestRequestData struct {
	Operation           OperationID
	Base                EndpointExpectation
	Head                EndpointExpectation
	Title               string
	Body                string
	Draft               bool
	MaintainerCanModify bool
}

func NewCreatePullRequestRequest(d CreatePullRequestRequestData) (CreatePullRequestRequest, error) {
	if err := d.validate(); err != nil {
		return CreatePullRequestRequest{}, err
	}
	return CreatePullRequestRequest{data: d.clone(), initialized: true}, nil
}
func (v CreatePullRequestRequest) Valid() bool                        { return v.initialized && v.data.validate() == nil }
func (v CreatePullRequestRequest) Data() CreatePullRequestRequestData { return v.data.clone() }
func (v CreatePullRequestRequest) Clone() CreatePullRequestRequest    { return v }
func (d CreatePullRequestRequestData) clone() CreatePullRequestRequestData {

	return d
}
func (d CreatePullRequestRequestData) validate() error {
	if !d.Operation.Valid() {
		return invalid("d.Operation")
	}
	if !d.Base.Valid() {
		return invalid("d.Base")
	}
	if !d.Head.Valid() {
		return invalid("d.Head")
	}

	if !nonblank(d.Title) || !textValue(d.Title) || !textValue(d.Body) {
		return invalid("PR literal text")
	}
	return nil
}

// RemoteCreateEvidence is an immutable semantic boundary value. Its zero is invalid.
type RemoteCreateEvidence struct {
	data        RemoteCreateEvidenceData
	initialized bool
}

// RemoteCreateEvidenceData is a construction/copy value; NewRemoteCreateEvidence validates and owns a copy.
type RemoteCreateEvidenceData struct {
	OperationID       OperationID
	RequestedBase     EndpointExpectation
	RequestedHead     EndpointExpectation
	Interval          ObservationInterval
	ProviderRequestID Optional[string]
	ReturnedID        Optional[domain.PRID]
	ReturnedURL       Optional[string]
}

func NewRemoteCreateEvidence(d RemoteCreateEvidenceData) (RemoteCreateEvidence, error) {
	if err := d.validate(); err != nil {
		return RemoteCreateEvidence{}, err
	}
	return RemoteCreateEvidence{data: d.clone(), initialized: true}, nil
}
func (v RemoteCreateEvidence) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v RemoteCreateEvidence) Data() RemoteCreateEvidenceData { return v.data.clone() }
func (v RemoteCreateEvidence) Clone() RemoteCreateEvidence    { return v }
func (d RemoteCreateEvidenceData) clone() RemoteCreateEvidenceData {

	return d
}
func (d RemoteCreateEvidenceData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.RequestedBase.Valid() {
		return invalid("d.RequestedBase")
	}
	if !d.RequestedHead.Valid() {
		return invalid("d.RequestedHead")
	}
	if !d.Interval.Valid() {
		return invalid("d.Interval")
	}

	if item, ok := d.ReturnedID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// NotSubmitted is an immutable semantic boundary value. Its zero is invalid.
type NotSubmitted struct {
	data        NotSubmittedData
	initialized bool
}

// NotSubmittedData is a construction/copy value; NewNotSubmitted validates and owns a copy.
type NotSubmittedData struct {
	Reason Diagnostic
}

func NewNotSubmitted(d NotSubmittedData) (NotSubmitted, error) {
	if err := d.validate(); err != nil {
		return NotSubmitted{}, err
	}
	return NotSubmitted{data: d.clone(), initialized: true}, nil
}
func (v NotSubmitted) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v NotSubmitted) Data() NotSubmittedData { return v.data.clone() }
func (v NotSubmitted) Clone() NotSubmitted    { return v }
func (d NotSubmittedData) clone() NotSubmittedData {

	return d
}
func (d NotSubmittedData) validate() error {
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}

	return nil
}

// RejectedNoCreation is an immutable semantic boundary value. Its zero is invalid.
type RejectedNoCreation struct {
	data        RejectedNoCreationData
	initialized bool
}

// RejectedNoCreationData is a construction/copy value; NewRejectedNoCreation validates and owns a copy.
type RejectedNoCreationData struct {
	Reason Diagnostic
}

func NewRejectedNoCreation(d RejectedNoCreationData) (RejectedNoCreation, error) {
	if err := d.validate(); err != nil {
		return RejectedNoCreation{}, err
	}
	return RejectedNoCreation{data: d.clone(), initialized: true}, nil
}
func (v RejectedNoCreation) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v RejectedNoCreation) Data() RejectedNoCreationData { return v.data.clone() }
func (v RejectedNoCreation) Clone() RejectedNoCreation    { return v }
func (d RejectedNoCreationData) clone() RejectedNoCreationData {

	return d
}
func (d RejectedNoCreationData) validate() error {
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}

	return nil
}

// ExistingCandidate is an immutable semantic boundary value. Its zero is invalid.
type ExistingCandidate struct {
	data        ExistingCandidateData
	initialized bool
}

// ExistingCandidateData is a construction/copy value; NewExistingCandidate validates and owns a copy.
type ExistingCandidateData struct {
	Candidates []PullRequestFact
	Page       PageInfo
}

func NewExistingCandidate(d ExistingCandidateData) (ExistingCandidate, error) {
	if err := d.validate(); err != nil {
		return ExistingCandidate{}, err
	}
	return ExistingCandidate{data: d.clone(), initialized: true}, nil
}
func (v ExistingCandidate) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v ExistingCandidate) Data() ExistingCandidateData { return v.data.clone() }
func (v ExistingCandidate) Clone() ExistingCandidate    { return v }
func (d ExistingCandidateData) clone() ExistingCandidateData {
	d.Candidates = cloneSlice(d.Candidates)
	return d
}
func (d ExistingCandidateData) validate() error {
	for _, item := range d.Candidates {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Page.Valid() {
		return invalid("d.Page")
	}
	if len(d.Candidates) == 0 {
		return invalid("existing candidate")
	}
	return nil
}

// CreatedVerified is an immutable semantic boundary value. Its zero is invalid.
type CreatedVerified struct {
	data        CreatedVerifiedData
	initialized bool
}

// CreatedVerifiedData is a construction/copy value; NewCreatedVerified validates and owns a copy.
type CreatedVerifiedData struct {
	Created       PullRequestFact
	RequestedBase EndpointExpectation
	RequestedHead EndpointExpectation
}

func NewCreatedVerified(d CreatedVerifiedData) (CreatedVerified, error) {
	if err := d.validate(); err != nil {
		return CreatedVerified{}, err
	}
	return CreatedVerified{data: d.clone(), initialized: true}, nil
}
func (v CreatedVerified) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v CreatedVerified) Data() CreatedVerifiedData { return v.data.clone() }
func (v CreatedVerified) Clone() CreatedVerified    { return v }
func (d CreatedVerifiedData) clone() CreatedVerifiedData {

	return d
}
func (d CreatedVerifiedData) validate() error {
	if !d.Created.Valid() {
		return invalid("d.Created")
	}
	if !d.RequestedBase.Valid() {
		return invalid("d.RequestedBase")
	}
	if !d.RequestedHead.Valid() {
		return invalid("d.RequestedHead")
	}
	if !endpointMatches(d.Created.data.Base, d.RequestedBase) || !endpointMatches(d.Created.data.Head, d.RequestedHead) {
		return invalid("created endpoint mismatch")
	}
	return nil
}

// CreatedWithDrift is an immutable semantic boundary value. Its zero is invalid.
type CreatedWithDrift struct {
	data        CreatedWithDriftData
	initialized bool
}

// CreatedWithDriftData is a construction/copy value; NewCreatedWithDrift validates and owns a copy.
type CreatedWithDriftData struct {
	Created       PullRequestFact
	RequestedBase EndpointExpectation
	RequestedHead EndpointExpectation
	Reason        Diagnostic
}

func NewCreatedWithDrift(d CreatedWithDriftData) (CreatedWithDrift, error) {
	if err := d.validate(); err != nil {
		return CreatedWithDrift{}, err
	}
	return CreatedWithDrift{data: d.clone(), initialized: true}, nil
}
func (v CreatedWithDrift) Valid() bool                { return v.initialized && v.data.validate() == nil }
func (v CreatedWithDrift) Data() CreatedWithDriftData { return v.data.clone() }
func (v CreatedWithDrift) Clone() CreatedWithDrift    { return v }
func (d CreatedWithDriftData) clone() CreatedWithDriftData {

	return d
}
func (d CreatedWithDriftData) validate() error {
	if !d.Created.Valid() {
		return invalid("d.Created")
	}
	if !d.RequestedBase.Valid() {
		return invalid("d.RequestedBase")
	}
	if !d.RequestedHead.Valid() {
		return invalid("d.RequestedHead")
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}

	return nil
}

// CreationIndeterminate is an immutable semantic boundary value. Its zero is invalid.
type CreationIndeterminate struct {
	data        CreationIndeterminateData
	initialized bool
}

// CreationIndeterminateData is a construction/copy value; NewCreationIndeterminate validates and owns a copy.
type CreationIndeterminateData struct {
	RequestEvidence RemoteCreateEvidence
	Candidate       Optional[PullRequestFact]
	Reason          Diagnostic
}

func NewCreationIndeterminate(d CreationIndeterminateData) (CreationIndeterminate, error) {
	if err := d.validate(); err != nil {
		return CreationIndeterminate{}, err
	}
	return CreationIndeterminate{data: d.clone(), initialized: true}, nil
}
func (v CreationIndeterminate) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v CreationIndeterminate) Data() CreationIndeterminateData { return v.data.clone() }
func (v CreationIndeterminate) Clone() CreationIndeterminate    { return v }
func (d CreationIndeterminateData) clone() CreationIndeterminateData {

	return d
}
func (d CreationIndeterminateData) validate() error {
	if !d.RequestEvidence.Valid() {
		return invalid("d.RequestEvidence")
	}
	if item, ok := d.Candidate.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Reason.Valid() {
		return invalid("d.Reason")
	}

	return nil
}

// CreatePullRequestResult is an immutable semantic boundary value. Its zero is invalid.
type CreatePullRequestResult struct {
	data        CreatePullRequestResultData
	initialized bool
}

// CreatePullRequestResultData is a construction/copy value; NewCreatePullRequestResult validates and owns a copy.
type CreatePullRequestResultData struct {
	Outcome               PullRequestCreationOutcome
	Effects               EffectReport
	CancellationRequested bool
	Recovery              []RecoveryRecord
	Observation           Optional[RemoteObservation]
	Diagnostics           []Diagnostic
	Transport             CommandTransportOutcome
}

func NewCreatePullRequestResult(d CreatePullRequestResultData) (CreatePullRequestResult, error) {
	if err := d.validate(); err != nil {
		return CreatePullRequestResult{}, err
	}
	return CreatePullRequestResult{data: d.clone(), initialized: true}, nil
}
func (v CreatePullRequestResult) Valid() bool                       { return v.initialized && v.data.validate() == nil }
func (v CreatePullRequestResult) Data() CreatePullRequestResultData { return v.data.clone() }
func (v CreatePullRequestResult) Clone() CreatePullRequestResult    { return v }
func (d CreatePullRequestResultData) clone() CreatePullRequestResultData {
	d.Recovery = cloneSlice(d.Recovery)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d CreatePullRequestResultData) validate() error {
	if !validPullRequestCreationOutcome(d.Outcome) {
		return invalid("d.Outcome")
	}
	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}

	for _, item := range d.Recovery {
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
	if err := validateCreatePullRequestResultEvidence(d); err != nil {
		return err
	}
	return validateRecoveryReferences(d.Effects, d.Recovery)
}
