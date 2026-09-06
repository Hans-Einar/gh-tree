package api

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// SessionCapabilities is an immutable semantic boundary value. Its zero is invalid.
type SessionCapabilities struct {
	data        SessionCapabilitiesData
	initialized bool
}

// SessionCapabilitiesData is a construction/copy value; NewSessionCapabilities validates and owns a copy.
type SessionCapabilitiesData struct {
	Output      bool
	Input       bool
	Resize      bool
	TerminalETX bool
	TreeStop    bool
	Restart     bool
}

func NewSessionCapabilities(d SessionCapabilitiesData) (SessionCapabilities, error) {
	if err := d.validate(); err != nil {
		return SessionCapabilities{}, err
	}
	return SessionCapabilities{data: d.clone(), initialized: true}, nil
}
func (v SessionCapabilities) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v SessionCapabilities) Data() SessionCapabilitiesData { return v.data.clone() }
func (v SessionCapabilities) Clone() SessionCapabilities    { return v }
func (d SessionCapabilitiesData) clone() SessionCapabilitiesData {

	return d
}
func (d SessionCapabilitiesData) validate() error {

	return nil
}

// SessionExit is an immutable semantic boundary value. Its zero is invalid.
type SessionExit struct {
	data        SessionExitData
	initialized bool
}

// SessionExitData is a construction/copy value; NewSessionExit validates and owns a copy.
type SessionExitData struct {
	Code   Optional[int]
	Signal Optional[string]
	Cause  ExitCause
}

func NewSessionExit(d SessionExitData) (SessionExit, error) {
	if err := d.validate(); err != nil {
		return SessionExit{}, err
	}
	return SessionExit{data: d.clone(), initialized: true}, nil
}
func (v SessionExit) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v SessionExit) Data() SessionExitData { return v.data.clone() }
func (v SessionExit) Clone() SessionExit    { return v }
func (d SessionExitData) clone() SessionExitData {

	return d
}
func (d SessionExitData) validate() error {

	if !d.Cause.Valid() {
		return invalid("d.Cause")
	}
	if !d.Code.Present() && !d.Signal.Present() {
		return invalid("exit evidence")
	}
	return nil
}

// RuntimeResidual is an immutable semantic boundary value. Its zero is invalid.
type RuntimeResidual struct {
	data        RuntimeResidualData
	initialized bool
}

// RuntimeResidualData is a construction/copy value; NewRuntimeResidual validates and owns a copy.
type RuntimeResidualData struct {
	SessionID Optional[domain.SessionID]
	Stage     RuntimeCleanupStage
	Detail    Diagnostic
}

func NewRuntimeResidual(d RuntimeResidualData) (RuntimeResidual, error) {
	if err := d.validate(); err != nil {
		return RuntimeResidual{}, err
	}
	return RuntimeResidual{data: d.clone(), initialized: true}, nil
}
func (v RuntimeResidual) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v RuntimeResidual) Data() RuntimeResidualData { return v.data.clone() }
func (v RuntimeResidual) Clone() RuntimeResidual    { return v }
func (d RuntimeResidualData) clone() RuntimeResidualData {

	return d
}
func (d RuntimeResidualData) validate() error {
	if item, ok := d.SessionID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Stage.Valid() {
		return invalid("d.Stage")
	}
	if !d.Detail.Valid() {
		return invalid("d.Detail")
	}

	return nil
}

// SessionCleanup is an immutable semantic boundary value. Its zero is invalid.
type SessionCleanup struct {
	data        SessionCleanupData
	initialized bool
}

// SessionCleanupData is a construction/copy value; NewSessionCleanup validates and owns a copy.
type SessionCleanupData struct {
	State     CleanupState
	Residuals []RuntimeResidual
}

func NewSessionCleanup(d SessionCleanupData) (SessionCleanup, error) {
	if err := d.validate(); err != nil {
		return SessionCleanup{}, err
	}
	return SessionCleanup{data: d.clone(), initialized: true}, nil
}
func (v SessionCleanup) Valid() bool              { return v.initialized && v.data.validate() == nil }
func (v SessionCleanup) Data() SessionCleanupData { return v.data.clone() }
func (v SessionCleanup) Clone() SessionCleanup    { return v }
func (d SessionCleanupData) clone() SessionCleanupData {
	d.Residuals = cloneSlice(d.Residuals)
	return d
}
func (d SessionCleanupData) validate() error {
	if !d.State.Valid() {
		return invalid("d.State")
	}
	for _, item := range d.Residuals {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.State == CleanupComplete && len(d.Residuals) > 0 {
		return invalid("complete cleanup residual")
	}
	return nil
}

// OutputRange is an immutable semantic boundary value. Its zero is invalid.
type OutputRange struct {
	data        OutputRangeData
	initialized bool
}

// OutputRangeData is a construction/copy value; NewOutputRange validates and owns a copy.
type OutputRangeData struct {
	RetainedStart uint64
	End           uint64
	Truncated     bool
}

func NewOutputRange(d OutputRangeData) (OutputRange, error) {
	if err := d.validate(); err != nil {
		return OutputRange{}, err
	}
	return OutputRange{data: d.clone(), initialized: true}, nil
}
func (v OutputRange) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v OutputRange) Data() OutputRangeData { return v.data.clone() }
func (v OutputRange) Clone() OutputRange    { return v }
func (d OutputRangeData) clone() OutputRangeData {

	return d
}
func (d OutputRangeData) validate() error {

	if d.RetainedStart > d.End || d.End-d.RetainedStart > 262144 {
		return invalid("output range")
	}
	return nil
}

// AcquiredCwd is an immutable semantic boundary value. Its zero is invalid.
type AcquiredCwd struct {
	data        AcquiredCwdData
	initialized bool
}

// AcquiredCwdData is a construction/copy value; NewAcquiredCwd validates and owns a copy.
type AcquiredCwdData struct {
	Observation   CwdObservation
	ActualLocator Optional[string]
	Diagnostics   []Diagnostic
}

func NewAcquiredCwd(d AcquiredCwdData) (AcquiredCwd, error) {
	if err := d.validate(); err != nil {
		return AcquiredCwd{}, err
	}
	return AcquiredCwd{data: d.clone(), initialized: true}, nil
}
func (v AcquiredCwd) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v AcquiredCwd) Data() AcquiredCwdData { return v.data.clone() }
func (v AcquiredCwd) Clone() AcquiredCwd    { return v }
func (d AcquiredCwdData) clone() AcquiredCwdData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d AcquiredCwdData) validate() error {
	if !d.Observation.Valid() {
		return invalid("d.Observation")
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// SessionSnapshot is an immutable semantic boundary value. Its zero is invalid.
type SessionSnapshot struct {
	data        SessionSnapshotData
	initialized bool
}

// SessionSnapshotData is a construction/copy value; NewSessionSnapshot validates and owns a copy.
type SessionSnapshotData struct {
	SessionID      domain.SessionID
	WorktreeID     domain.WorktreeID
	StartOperation OperationID
	RestartOf      Optional[domain.SessionID]
	Display        InvocationSummary
	Capabilities   SessionCapabilities
	Phase          SessionPhase
	Exit           Optional[SessionExit]
	Cleanup        SessionCleanup
	Sequence       SessionSequence
	OutputRange    OutputRange
	AcquiredCwd    Optional[AcquiredCwd]
}

func NewSessionSnapshot(d SessionSnapshotData) (SessionSnapshot, error) {
	if err := d.validate(); err != nil {
		return SessionSnapshot{}, err
	}
	return SessionSnapshot{data: d.clone(), initialized: true}, nil
}
func (v SessionSnapshot) Valid() bool               { return v.initialized && v.data.validate() == nil }
func (v SessionSnapshot) Data() SessionSnapshotData { return v.data.clone() }
func (v SessionSnapshot) Clone() SessionSnapshot    { return v }
func (d SessionSnapshotData) clone() SessionSnapshotData {

	return d
}
func (d SessionSnapshotData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.WorktreeID.Valid() {
		return invalid("d.WorktreeID")
	}
	if !d.StartOperation.Valid() {
		return invalid("d.StartOperation")
	}
	if item, ok := d.RestartOf.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Display.Valid() {
		return invalid("d.Display")
	}
	if !d.Capabilities.Valid() {
		return invalid("d.Capabilities")
	}
	if !d.Phase.Valid() {
		return invalid("d.Phase")
	}
	if item, ok := d.Exit.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Cleanup.Valid() {
		return invalid("d.Cleanup")
	}
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}
	if !d.OutputRange.Valid() {
		return invalid("d.OutputRange")
	}
	if item, ok := d.AcquiredCwd.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}
	return validateSession(d)
}

// SessionStartRequest is an immutable semantic boundary value. Its zero is invalid.
type SessionStartRequest struct {
	data        SessionStartRequestData
	initialized bool
}

// SessionStartRequestData is a construction/copy value; NewSessionStartRequest validates and owns a copy.
type SessionStartRequestData struct {
	OperationID OperationID
	Invocation  Invocation
}

func NewSessionStartRequest(d SessionStartRequestData) (SessionStartRequest, error) {
	if err := d.validate(); err != nil {
		return SessionStartRequest{}, err
	}
	return SessionStartRequest{data: d.clone(), initialized: true}, nil
}
func (v SessionStartRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v SessionStartRequest) Data() SessionStartRequestData { return v.data.clone() }
func (v SessionStartRequest) Clone() SessionStartRequest    { return v }
func (d SessionStartRequestData) clone() SessionStartRequestData {

	return d
}
func (d SessionStartRequestData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.Invocation.Valid() {
		return invalid("d.Invocation")
	}

	return nil
}

// SessionStopRequest is an immutable semantic boundary value. Its zero is invalid.
type SessionStopRequest struct {
	data        SessionStopRequestData
	initialized bool
}

// SessionStopRequestData is a construction/copy value; NewSessionStopRequest validates and owns a copy.
type SessionStopRequestData struct {
	OperationID OperationID
	SessionID   domain.SessionID
}

func NewSessionStopRequest(d SessionStopRequestData) (SessionStopRequest, error) {
	if err := d.validate(); err != nil {
		return SessionStopRequest{}, err
	}
	return SessionStopRequest{data: d.clone(), initialized: true}, nil
}
func (v SessionStopRequest) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v SessionStopRequest) Data() SessionStopRequestData { return v.data.clone() }
func (v SessionStopRequest) Clone() SessionStopRequest    { return v }
func (d SessionStopRequestData) clone() SessionStopRequestData {

	return d
}
func (d SessionStopRequestData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	return nil
}

// SessionRestartRequest is an immutable semantic boundary value. Its zero is invalid.
type SessionRestartRequest struct {
	data        SessionRestartRequestData
	initialized bool
}

// SessionRestartRequestData is a construction/copy value; NewSessionRestartRequest validates and owns a copy.
type SessionRestartRequestData struct {
	OperationID OperationID
	SessionID   domain.SessionID
	Geometry    Optional[Geometry]
}

func NewSessionRestartRequest(d SessionRestartRequestData) (SessionRestartRequest, error) {
	if err := d.validate(); err != nil {
		return SessionRestartRequest{}, err
	}
	return SessionRestartRequest{data: d.clone(), initialized: true}, nil
}
func (v SessionRestartRequest) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v SessionRestartRequest) Data() SessionRestartRequestData { return v.data.clone() }
func (v SessionRestartRequest) Clone() SessionRestartRequest    { return v }
func (d SessionRestartRequestData) clone() SessionRestartRequestData {

	return d
}
func (d SessionRestartRequestData) validate() error {
	if !d.OperationID.Valid() {
		return invalid("d.OperationID")
	}
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

// SessionFilter is an immutable semantic boundary value. Its zero is invalid.
type SessionFilter struct {
	data        SessionFilterData
	initialized bool
}

// SessionFilterData is a construction/copy value; NewSessionFilter validates and owns a copy.
type SessionFilterData struct {
	WorktreeID Optional[domain.WorktreeID]
}

func NewSessionFilter(d SessionFilterData) (SessionFilter, error) {
	if err := d.validate(); err != nil {
		return SessionFilter{}, err
	}
	return SessionFilter{data: d.clone(), initialized: true}, nil
}
func (v SessionFilter) Valid() bool             { return v.initialized && v.data.validate() == nil }
func (v SessionFilter) Data() SessionFilterData { return v.data.clone() }
func (v SessionFilter) Clone() SessionFilter    { return v }
func (d SessionFilterData) clone() SessionFilterData {

	return d
}
func (d SessionFilterData) validate() error {
	if item, ok := d.WorktreeID.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// SessionList is an immutable semantic boundary value. Its zero is invalid.
type SessionList struct {
	data        SessionListData
	initialized bool
}

// SessionListData is a construction/copy value; NewSessionList validates and owns a copy.
type SessionListData struct {
	Sessions []SessionSnapshot
	Sequence RuntimeEventSequence
}

func NewSessionList(d SessionListData) (SessionList, error) {
	if err := d.validate(); err != nil {
		return SessionList{}, err
	}
	return SessionList{data: d.clone(), initialized: true}, nil
}
func (v SessionList) Valid() bool           { return v.initialized && v.data.validate() == nil }
func (v SessionList) Data() SessionListData { return v.data.clone() }
func (v SessionList) Clone() SessionList    { return v }
func (d SessionListData) clone() SessionListData {
	d.Sessions = cloneSlice(d.Sessions)
	return d
}
func (d SessionListData) validate() error {
	for _, item := range d.Sessions {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}
	for i := 1; i < len(d.Sessions); i++ {
		if d.Sessions[i-1].data.SessionID.Value() >= d.Sessions[i].data.SessionID.Value() {
			return invalid("session list order")
		}
	}
	return nil
}

// SessionOutputRequest is an immutable semantic boundary value. Its zero is invalid.
type SessionOutputRequest struct {
	data        SessionOutputRequestData
	initialized bool
}

// SessionOutputRequestData is a construction/copy value; NewSessionOutputRequest validates and owns a copy.
type SessionOutputRequestData struct {
	SessionID domain.SessionID
	Offset    uint64
	MaxBytes  uint32
}

func NewSessionOutputRequest(d SessionOutputRequestData) (SessionOutputRequest, error) {
	if err := d.validate(); err != nil {
		return SessionOutputRequest{}, err
	}
	return SessionOutputRequest{data: d.clone(), initialized: true}, nil
}
func (v SessionOutputRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v SessionOutputRequest) Data() SessionOutputRequestData { return v.data.clone() }
func (v SessionOutputRequest) Clone() SessionOutputRequest    { return v }
func (d SessionOutputRequestData) clone() SessionOutputRequestData {

	return d
}
func (d SessionOutputRequestData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	if d.MaxBytes < 1 || d.MaxBytes > 262144 {
		return invalid("output bound")
	}
	return nil
}

// SessionWriteRequest is an immutable semantic boundary value. Its zero is invalid.
type SessionWriteRequest struct {
	data        SessionWriteRequestData
	initialized bool
}

// SessionWriteRequestData is a construction/copy value; NewSessionWriteRequest validates and owns a copy.
type SessionWriteRequestData struct {
	SessionID domain.SessionID
	Bytes     []byte
}

func NewSessionWriteRequest(d SessionWriteRequestData) (SessionWriteRequest, error) {
	if err := d.validate(); err != nil {
		return SessionWriteRequest{}, err
	}
	return SessionWriteRequest{data: d.clone(), initialized: true}, nil
}
func (v SessionWriteRequest) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v SessionWriteRequest) Data() SessionWriteRequestData { return v.data.clone() }
func (v SessionWriteRequest) Clone() SessionWriteRequest    { return v }
func (d SessionWriteRequestData) clone() SessionWriteRequestData {
	d.Bytes = cloneSlice(d.Bytes)
	return d
}
func (d SessionWriteRequestData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}

	if len(d.Bytes) < 1 || len(d.Bytes) > 65536 {
		return invalid("input bound")
	}
	return nil
}

// SessionResizeRequest is an immutable semantic boundary value. Its zero is invalid.
type SessionResizeRequest struct {
	data        SessionResizeRequestData
	initialized bool
}

// SessionResizeRequestData is a construction/copy value; NewSessionResizeRequest validates and owns a copy.
type SessionResizeRequestData struct {
	SessionID domain.SessionID
	Geometry  Geometry
}

func NewSessionResizeRequest(d SessionResizeRequestData) (SessionResizeRequest, error) {
	if err := d.validate(); err != nil {
		return SessionResizeRequest{}, err
	}
	return SessionResizeRequest{data: d.clone(), initialized: true}, nil
}
func (v SessionResizeRequest) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v SessionResizeRequest) Data() SessionResizeRequestData { return v.data.clone() }
func (v SessionResizeRequest) Clone() SessionResizeRequest    { return v }
func (d SessionResizeRequestData) clone() SessionResizeRequestData {

	return d
}
func (d SessionResizeRequestData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.Geometry.Valid() {
		return invalid("d.Geometry")
	}

	return nil
}

// SessionStartResult is an immutable semantic boundary value. Its zero is invalid.
type SessionStartResult struct {
	data        SessionStartResultData
	initialized bool
}

// SessionStartResultData is a construction/copy value; NewSessionStartResult validates and owns a copy.
type SessionStartResultData struct {
	Session           Optional[SessionSnapshot]
	Established       bool
	CancellationAsked bool
	Effects           EffectReport
	Diagnostics       []Diagnostic
}

func NewSessionStartResult(d SessionStartResultData) (SessionStartResult, error) {
	if err := d.validate(); err != nil {
		return SessionStartResult{}, err
	}
	return SessionStartResult{data: d.clone(), initialized: true}, nil
}
func (v SessionStartResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v SessionStartResult) Data() SessionStartResultData { return v.data.clone() }
func (v SessionStartResult) Clone() SessionStartResult    { return v }
func (d SessionStartResultData) clone() SessionStartResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionStartResultData) validate() error {
	if item, ok := d.Session.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := validateSessionStartResultEvidence(d); err != nil {
		return err
	}
	if d.Established {
		if s, p := d.Session.Value(); !p || !s.data.AcquiredCwd.Present() || s.data.Phase == Starting {
			return invalid("established session identity/cwd barrier")
		}
	}
	return nil
}

// SessionControlResult is an immutable semantic boundary value. Its zero is invalid.
type SessionControlResult struct {
	data        SessionControlResultData
	initialized bool
}

// SessionControlResultData is a construction/copy value; NewSessionControlResult validates and owns a copy.
type SessionControlResultData struct {
	SessionID         domain.SessionID
	Sequence          SessionSequence
	Delivered         bool
	CancellationAsked bool
	Diagnostics       []Diagnostic
}

func NewSessionControlResult(d SessionControlResultData) (SessionControlResult, error) {
	if err := d.validate(); err != nil {
		return SessionControlResult{}, err
	}
	return SessionControlResult{data: d.clone(), initialized: true}, nil
}
func (v SessionControlResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v SessionControlResult) Data() SessionControlResultData { return v.data.clone() }
func (v SessionControlResult) Clone() SessionControlResult    { return v }
func (d SessionControlResultData) clone() SessionControlResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionControlResultData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return nil
}

// SessionWriteResult is an immutable semantic boundary value. Its zero is invalid.
type SessionWriteResult struct {
	data        SessionWriteResultData
	initialized bool
}

// SessionWriteResultData is a construction/copy value; NewSessionWriteResult validates and owns a copy.
type SessionWriteResultData struct {
	SessionID         domain.SessionID
	Sequence          SessionSequence
	AcceptedBytes     uint32
	CancellationAsked bool
	Diagnostics       []Diagnostic
}

func NewSessionWriteResult(d SessionWriteResultData) (SessionWriteResult, error) {
	if err := d.validate(); err != nil {
		return SessionWriteResult{}, err
	}
	return SessionWriteResult{data: d.clone(), initialized: true}, nil
}
func (v SessionWriteResult) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v SessionWriteResult) Data() SessionWriteResultData { return v.data.clone() }
func (v SessionWriteResult) Clone() SessionWriteResult    { return v }
func (d SessionWriteResultData) clone() SessionWriteResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionWriteResultData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if d.AcceptedBytes > 65536 {
		return invalid("accepted input")
	}
	return nil
}

// SessionStopResult is an immutable semantic boundary value. Its zero is invalid.
type SessionStopResult struct {
	data        SessionStopResultData
	initialized bool
}

// SessionStopResultData is a construction/copy value; NewSessionStopResult validates and owns a copy.
type SessionStopResultData struct {
	Session           SessionSnapshot
	CleanupComplete   bool
	CancellationAsked bool
	Effects           EffectReport
	Diagnostics       []Diagnostic
}

func NewSessionStopResult(d SessionStopResultData) (SessionStopResult, error) {
	if err := d.validate(); err != nil {
		return SessionStopResult{}, err
	}
	return SessionStopResult{data: d.clone(), initialized: true}, nil
}
func (v SessionStopResult) Valid() bool                 { return v.initialized && v.data.validate() == nil }
func (v SessionStopResult) Data() SessionStopResultData { return v.data.clone() }
func (v SessionStopResult) Clone() SessionStopResult    { return v }
func (d SessionStopResultData) clone() SessionStopResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionStopResultData) validate() error {
	if !d.Session.Valid() {
		return invalid("d.Session")
	}

	if !d.Effects.Valid() {
		return invalid("d.Effects")
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := validateSessionStopResultEvidence(d); err != nil {
		return err
	}
	if d.CleanupComplete != (d.Session.data.Phase == Cleaned) {
		return invalid("stop cleanup")
	}
	return nil
}

// SessionRestartResult is an immutable semantic boundary value. Its zero is invalid.
type SessionRestartResult struct {
	data        SessionRestartResultData
	initialized bool
}

// SessionRestartResultData is a construction/copy value; NewSessionRestartResult validates and owns a copy.
type SessionRestartResultData struct {
	Old               SessionStopResult
	Replacement       Optional[SessionStartResult]
	CancellationAsked bool
	Diagnostics       []Diagnostic
}

func NewSessionRestartResult(d SessionRestartResultData) (SessionRestartResult, error) {
	if err := d.validate(); err != nil {
		return SessionRestartResult{}, err
	}
	return SessionRestartResult{data: d.clone(), initialized: true}, nil
}
func (v SessionRestartResult) Valid() bool                    { return v.initialized && v.data.validate() == nil }
func (v SessionRestartResult) Data() SessionRestartResultData { return v.data.clone() }
func (v SessionRestartResult) Clone() SessionRestartResult    { return v }
func (d SessionRestartResultData) clone() SessionRestartResultData {
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d SessionRestartResultData) validate() error {
	if !d.Old.Valid() {
		return invalid("d.Old")
	}
	if item, ok := d.Replacement.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := validateSessionRestartResultEvidence(d); err != nil {
		return err
	}
	if r, ok := d.Replacement.Value(); ok {
		if !d.Old.data.CleanupComplete {
			return invalid("restart before cleanup")
		}
		if s, ok := r.data.Session.Value(); ok {
			if s.data.SessionID == d.Old.data.Session.data.SessionID {
				return invalid("reused session")
			}
			p, ok := s.data.RestartOf.Value()
			if !ok || p != d.Old.data.Session.data.SessionID {
				return invalid("restart predecessor")
			}
		}
	}
	return nil
}

// RuntimeShutdownResult is an immutable semantic boundary value. Its zero is invalid.
type RuntimeShutdownResult struct {
	data        RuntimeShutdownResultData
	initialized bool
}

// RuntimeShutdownResultData is a construction/copy value; NewRuntimeShutdownResult validates and owns a copy.
type RuntimeShutdownResultData struct {
	AdmissionClosed bool
	Complete        bool
	Sessions        []SessionStopResult
	Residuals       []RuntimeResidual
	Diagnostics     []Diagnostic
}

func NewRuntimeShutdownResult(d RuntimeShutdownResultData) (RuntimeShutdownResult, error) {
	if err := d.validate(); err != nil {
		return RuntimeShutdownResult{}, err
	}
	return RuntimeShutdownResult{data: d.clone(), initialized: true}, nil
}
func (v RuntimeShutdownResult) Valid() bool                     { return v.initialized && v.data.validate() == nil }
func (v RuntimeShutdownResult) Data() RuntimeShutdownResultData { return v.data.clone() }
func (v RuntimeShutdownResult) Clone() RuntimeShutdownResult    { return v }
func (d RuntimeShutdownResultData) clone() RuntimeShutdownResultData {
	d.Sessions = cloneSlice(d.Sessions)
	d.Residuals = cloneSlice(d.Residuals)
	d.Diagnostics = cloneSlice(d.Diagnostics)
	return d
}
func (d RuntimeShutdownResultData) validate() error {

	for _, item := range d.Sessions {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Residuals {
		if !item.Valid() {
			return invalid("item")
		}
	}
	for _, item := range d.Diagnostics {
		if !item.Valid() {
			return invalid("item")
		}
	}
	if err := validateRuntimeShutdownResultEvidence(d); err != nil {
		return err
	}
	if d.Complete {
		if !d.AdmissionClosed || len(d.Residuals) > 0 {
			return invalid("shutdown residual")
		}
		for _, s := range d.Sessions {
			if !s.data.CleanupComplete {
				return invalid("shutdown session")
			}
		}
	}
	return nil
}

// RuntimeEvent is an immutable semantic boundary value. Its zero is invalid.
type RuntimeEvent struct {
	data        RuntimeEventData
	initialized bool
}

// RuntimeEventData is a construction/copy value; NewRuntimeEvent validates and owns a copy.
type RuntimeEventData struct {
	Sequence        RuntimeEventSequence
	SessionSequence SessionSequence
	SessionID       domain.SessionID
	Kind            RuntimeEventKind
	Snapshot        SessionSnapshot
}

func NewRuntimeEvent(d RuntimeEventData) (RuntimeEvent, error) {
	if err := d.validate(); err != nil {
		return RuntimeEvent{}, err
	}
	return RuntimeEvent{data: d.clone(), initialized: true}, nil
}
func (v RuntimeEvent) Valid() bool            { return v.initialized && v.data.validate() == nil }
func (v RuntimeEvent) Data() RuntimeEventData { return v.data.clone() }
func (v RuntimeEvent) Clone() RuntimeEvent    { return v }
func (d RuntimeEventData) clone() RuntimeEventData {

	return d
}
func (d RuntimeEventData) validate() error {
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}
	if !d.SessionSequence.Valid() {
		return invalid("d.SessionSequence")
	}
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.Kind.Valid() {
		return invalid("d.Kind")
	}
	if !d.Snapshot.Valid() {
		return invalid("d.Snapshot")
	}
	if d.Sequence.Value() == 0 || d.SessionSequence != d.Snapshot.data.Sequence || d.SessionID != d.Snapshot.data.SessionID || (d.Kind == RuntimeCleaned && d.Snapshot.data.Phase != Cleaned) {
		return invalid("runtime event binding")
	}
	return nil
}

// SessionOutputChunk is an immutable semantic boundary value. Its zero is invalid.
type SessionOutputChunk struct {
	data        SessionOutputChunkData
	initialized bool
}

// SessionOutputChunkData is a construction/copy value; NewSessionOutputChunk validates and owns a copy.
type SessionOutputChunkData struct {
	Stream   OutputStream
	Offset   uint64
	Bytes    []byte
	Sequence SessionSequence
}

func NewSessionOutputChunk(d SessionOutputChunkData) (SessionOutputChunk, error) {
	if err := d.validate(); err != nil {
		return SessionOutputChunk{}, err
	}
	return SessionOutputChunk{data: d.clone(), initialized: true}, nil
}
func (v SessionOutputChunk) Valid() bool                  { return v.initialized && v.data.validate() == nil }
func (v SessionOutputChunk) Data() SessionOutputChunkData { return v.data.clone() }
func (v SessionOutputChunk) Clone() SessionOutputChunk    { return v }
func (d SessionOutputChunkData) clone() SessionOutputChunkData {
	d.Bytes = cloneSlice(d.Bytes)
	return d
}
func (d SessionOutputChunkData) validate() error {
	if !d.Stream.Valid() {
		return invalid("d.Stream")
	}

	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}
	if len(d.Bytes) == 0 || len(d.Bytes) > 262144 || d.Offset > ^uint64(0)-uint64(len(d.Bytes)) {
		return invalid("output chunk")
	}
	return nil
}

// OutputGap is an immutable semantic boundary value. Its zero is invalid.
type OutputGap struct {
	data        OutputGapData
	initialized bool
}

// OutputGapData is a construction/copy value; NewOutputGap validates and owns a copy.
type OutputGapData struct {
	From uint64
	To   uint64
}

func NewOutputGap(d OutputGapData) (OutputGap, error) {
	if err := d.validate(); err != nil {
		return OutputGap{}, err
	}
	return OutputGap{data: d.clone(), initialized: true}, nil
}
func (v OutputGap) Valid() bool         { return v.initialized && v.data.validate() == nil }
func (v OutputGap) Data() OutputGapData { return v.data.clone() }
func (v OutputGap) Clone() OutputGap    { return v }
func (d OutputGapData) clone() OutputGapData {

	return d
}
func (d OutputGapData) validate() error {

	if d.From >= d.To {
		return invalid("gap")
	}
	return nil
}

// SessionOutputResult is an immutable semantic boundary value. Its zero is invalid.
type SessionOutputResult struct {
	data        SessionOutputResultData
	initialized bool
}

// SessionOutputResultData is a construction/copy value; NewSessionOutputResult validates and owns a copy.
type SessionOutputResultData struct {
	SessionID     domain.SessionID
	Sequence      SessionSequence
	Chunks        []SessionOutputChunk
	RetainedStart uint64
	End           uint64
	NextOffset    uint64
	Gap           Optional[OutputGap]
	Truncated     bool
}

func NewSessionOutputResult(d SessionOutputResultData) (SessionOutputResult, error) {
	if err := d.validate(); err != nil {
		return SessionOutputResult{}, err
	}
	return SessionOutputResult{data: d.clone(), initialized: true}, nil
}
func (v SessionOutputResult) Valid() bool                   { return v.initialized && v.data.validate() == nil }
func (v SessionOutputResult) Data() SessionOutputResultData { return v.data.clone() }
func (v SessionOutputResult) Clone() SessionOutputResult    { return v }
func (d SessionOutputResultData) clone() SessionOutputResultData {
	d.Chunks = cloneSlice(d.Chunks)
	return d
}
func (d SessionOutputResultData) validate() error {
	if !d.SessionID.Valid() {
		return invalid("d.SessionID")
	}
	if !d.Sequence.Valid() {
		return invalid("d.Sequence")
	}
	for _, item := range d.Chunks {
		if !item.Valid() {
			return invalid("item")
		}
	}

	if item, ok := d.Gap.Value(); ok {
		if !item.Valid() {
			return invalid("item")
		}
	}

	return validateOutput(d)
}
