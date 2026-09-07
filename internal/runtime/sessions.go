package runtime

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type sessions struct {
	registry    *registry
	starter     nativeStarter
	environment []string
	budgets     sessionBudgets
	admission   sync.Mutex // memory-only lifecycle-key selection, never native waits
}

var _ ports.Sessions = (*sessions)(nil)

func newSessions(starter nativeStarter, environment []string, budgets sessionBudgets) *sessions {
	if starter == nil {
		panic("Runtime requires a native owner constructor")
	}
	if budgets.grace <= 0 {
		budgets.grace = 2 * time.Second
	}
	if budgets.force <= 0 {
		budgets.force = 3 * time.Second
	}
	if budgets.shutdown <= 0 {
		budgets.shutdown = 8 * time.Second
	}
	return &sessions{registry: newRegistry(), starter: starter, environment: append([]string(nil), environment...), budgets: budgets}
}

func owned[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func runtimeEffects(state api.EffectState) api.EffectReport {
	f := owned(api.NewFacetEffect(api.FacetEffectData{Facet: api.RuntimeResources, State: state}))
	return owned(api.NewEffectReport(api.EffectReportData{Facets: []api.FacetEffect{f}}))
}

func safeDiagnostic(err error) api.Diagnostic {
	var d api.Diagnostic
	if errors.As(err, &d) {
		return d
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return diagnostic(api.Canceled, "runtime.canceled", "Runtime caller waiting ended; accepted effects remain owned.")
	}
	return diagnostic(api.IOFailure, "runtime.native_failure", "The native owner reported a failure.")
}

func diagnostics(err error) []api.Diagnostic {
	if err == nil {
		return nil
	}
	return []api.Diagnostic{safeDiagnostic(err)}
}

func (s *session) diagnosticsLocked() []api.Diagnostic {
	var ds []api.Diagnostic
	for stage := api.Acquisition; stage <= api.EventTransfer; stage++ {
		if d, ok := s.diagnostics[stage]; ok {
			ds = append(ds, d)
		}
	}
	return ds
}

func (s *session) startResult(canceled bool) api.SessionStartResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := api.EffectPartial
	if s.established {
		state = api.AppliedVerified
	} else if s.owner == nil && !s.startPending {
		state = api.VerifiedNoTargetChange
	}
	return owned(api.NewSessionStartResult(api.SessionStartResultData{Session: api.Some(s.snapshot), Established: s.established, CancellationAsked: canceled, Effects: runtimeEffects(state), Diagnostics: s.diagnosticsLocked()}))
}

func (s *session) stopResult(canceled bool) api.SessionStopResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	complete := s.snapshot.Data().Phase == api.Cleaned
	state := api.EffectPartial
	if complete {
		state = api.AppliedVerified
	}
	return owned(api.NewSessionStopResult(api.SessionStopResultData{Session: s.snapshot, CleanupComplete: complete, CancellationAsked: canceled, Effects: runtimeEffects(state), Diagnostics: s.diagnosticsLocked()}))
}

// Environment selection is private and deterministic from the constructor copy.
// Platform execution/shell resolution will belong to the real native binding.
func (r *sessions) specification(inv api.Invocation) ([]string, api.InvocationSummary, api.SessionCapabilities, error) {
	d := inv.Data()
	windows := d.Cwd.Data().Worktree.Data().RootIdentity.Platform() == api.DirectoryWindows
	key := func(k string) string {
		if windows {
			return strings.ToUpper(k)
		}
		return k
	}
	env := make(map[string]string)
	p := d.Environment.Data()
	if p.InheritBase {
		for _, item := range r.environment {
			at := strings.IndexByte(item, '=')
			if at < 1 || strings.ContainsRune(item, 0) {
				return nil, api.InvocationSummary{}, api.SessionCapabilities{}, errInvalid
			}
			env[key(item[:at])] = item
		}
	}
	for _, k := range p.Remove {
		delete(env, key(k))
	}
	for _, entry := range p.Set {
		e := entry.Data()
		env[key(e.Name)] = e.Name + "=" + e.Value
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	resolved := make([]string, 0, len(keys))
	for _, k := range keys {
		resolved = append(resolved, env[k])
	}
	executable := "interactive shell"
	if argv, ok := d.Execution.(api.ArgvExecution); ok {
		executable = argv.Data().Executable
	}
	display := owned(api.NewInvocationSummary(api.InvocationSummaryData{Label: d.Label, ExecutableDisplay: executable, Cwd: d.Cwd, AcceptedLocator: d.Cwd.Data().Worktree.Data().RootLocator, Terminal: d.Terminal, Geometry: d.Geometry}))
	terminal := d.Terminal == api.Terminal
	caps := owned(api.NewSessionCapabilities(api.SessionCapabilitiesData{Output: true, Input: true, Resize: terminal, TerminalETX: terminal, TreeStop: true, Restart: true}))
	return resolved, display, caps, nil
}

func refusedStart(err error, canceled bool) (api.SessionStartResult, error) {
	return owned(api.NewSessionStartResult(api.SessionStartResultData{CancellationAsked: canceled, Effects: runtimeEffects(api.NotStarted), Diagnostics: diagnostics(err)})), err
}

func (r *sessions) Start(ctx context.Context, req api.SessionStartRequest) (api.SessionStartResult, error) {
	if ctx == nil || !req.Valid() {
		return refusedStart(errInvalid, false)
	}
	return r.start(ctx, req, nil, api.None[domain.SessionID]())
}

func (r *sessions) start(ctx context.Context, req api.SessionStartRequest, environment []string, predecessor api.Optional[domain.SessionID]) (api.SessionStartResult, error) {
	r.admission.Lock()
	// Deduplication lives with the bounded retained records. Evicted IDs/keys do
	// not become authority to resurrect an old transition.
	r.registry.mu.Lock()
	for _, old := range r.registry.sessions {
		old.mu.Lock()
		same := old.start.Data().OperationID == req.Data().OperationID
		collision := old.restart != nil && old.restart.request.Data().OperationID == req.Data().OperationID && !predecessor.Present()
		old.mu.Unlock()
		if same || collision {
			r.registry.mu.Unlock()
			r.admission.Unlock()
			if collision || !reflect.DeepEqual(old.start.Data(), req.Data()) {
				return refusedStart(errInvalid, ctx.Err() != nil)
			}
			return waitStart(ctx, old)
		}
	}
	r.registry.mu.Unlock()
	env, display, caps, err := r.specification(req.Data().Invocation)
	if environment != nil {
		env = append([]string(nil), environment...)
	}
	if err != nil {
		r.admission.Unlock()
		return refusedStart(err, ctx.Err() != nil)
	}
	s, err := r.registry.admit(ctx, req, env, display, caps, predecessor)
	if err == nil {
		if oldID, present := predecessor.Value(); present {
			old, lookupErr := r.registry.lookup(oldID)
			if lookupErr == nil {
				old.mu.Lock()
				if old.restart != nil {
					old.restart.replacement = s
				}
				old.mu.Unlock()
			}
		}
	}
	r.admission.Unlock()
	if err != nil {
		return refusedStart(err, ctx.Err() != nil)
	}
	go r.acquire(ctx, s)
	return waitStart(ctx, s)
}

func waitStart(ctx context.Context, s *session) (api.SessionStartResult, error) {
	select {
	case <-s.startDone:
	case <-ctx.Done():
	}
	s.mu.Lock()
	pending, err := s.startPending, s.startErr
	s.mu.Unlock()
	if pending {
		err = ctx.Err()
	}
	return s.startResult(ctx.Err() != nil), err
}

func (r *sessions) acquire(ctx context.Context, s *session) {
	s.mu.Lock()
	d := s.snapshot.Data()
	inv := s.start.Data().Invocation
	env := append([]string(nil), s.environment...)
	s.mu.Unlock()
	owner, fact, err := r.starter(ctx, nativeStart{ID: d.SessionID, Invocation: inv, Environment: env, Grace: r.budgets.grace, Force: r.budgets.force, Output: func(stream api.OutputStream, data []byte) { r.capture(s, stream, data) }})
	s.mu.Lock()
	s.owner = owner
	s.startErr = err
	if err != nil {
		s.diagnostics[api.Acquisition] = safeDiagnostic(err)
	}
	s.mu.Unlock()
	_ = r.registry.change(s, api.StateChanged, func(d *api.SessionSnapshotData) error {
		s.established = fact.Established
		if fact.Cwd.Present() {
			s.acquired = fact.Cwd
			d.AcquiredCwd = fact.Cwd
		}
		if fact.Established && !s.stopAsked {
			d.Phase = api.Running
		} else {
			d.Phase = api.Stopping
		}
		if !fact.Established {
			s.stopAsked = true
		}
		return nil
	})
	s.mu.Lock()
	s.startPending = false
	if owner == nil {
		s.nativeClean = true
	} else {
		s.producers++
		s.observing = true
		go r.writeLoop(s)
	}
	close(s.startDone)
	stop := s.stopAsked
	s.mu.Unlock()
	if stop {
		r.requestStop(s)
	}
	if owner == nil {
		r.finalize(s)
		return
	}
	r.observe(s, owner)
}

func (r *sessions) capture(s *session, stream api.OutputStream, data []byte) {
	if len(data) == 0 {
		return
	}
	err := r.registry.change(s, api.OutputAvailable, func(d *api.SessionSnapshotData) error {
		if d.Sequence.Value() >= math.MaxUint64-1 || s.startPending && d.Sequence.Value() >= math.MaxUint64-2 {
			return errExhausted
		}
		seq := owned(api.NewSessionSequence(d.Sequence.Value() + 1))
		if err := s.output.append(stream, seq, data); err != nil {
			return err
		}
		d.OutputRange = s.output.rangeValue()
		return nil
	})
	if err != nil {
		s.mu.Lock()
		s.diagnostics[api.OutputCleanup] = safeDiagnostic(err)
		s.mu.Unlock()
		r.requestStop(s)
	}
}

func (r *sessions) observe(s *session, owner nativeOwner) {
	for {
		fact, err := owner.NextFact(context.Background())
		if err != nil {
			fact.Residuals = []api.RuntimeResidual{r.residual(s, api.SupervisorOrBroker, safeDiagnostic(err))}
		}
		_ = r.registry.change(s, api.StateChanged, func(d *api.SessionSnapshotData) error {
			if fact.Established {
				s.established = true
			}
			if fact.Cwd.Present() {
				s.acquired = fact.Cwd
				d.AcquiredCwd = fact.Cwd
			}
			if fact.Exit.Present() {
				s.exit = fact.Exit
				d.Exit = fact.Exit
				s.stopAsked = true
			}
			if diagnostic, ok := fact.Diagnostic.Value(); ok {
				s.diagnostics[api.SupervisorOrBroker] = diagnostic
			}
			if fact.CleanupComplete && len(fact.Residuals) == 0 {
				s.nativeClean = true
				s.stopAsked = true
			}
			if s.stopAsked {
				d.Phase = api.Stopping
			}
			if len(fact.Residuals) > 0 {
				d.Phase = api.CleanupFailed
				d.Cleanup = owned(api.NewSessionCleanup(api.SessionCleanupData{State: api.CleanupFailedState, Residuals: fact.Residuals}))
				s.stopAsked = true
			} else {
				d.Cleanup = owned(api.NewSessionCleanup(api.SessionCleanupData{State: api.CleanupPending}))
			}
			return nil
		})
		s.mu.Lock()
		stop, complete := s.stopAsked, s.nativeClean
		s.mu.Unlock()
		if stop {
			r.requestStop(s)
		}
		if complete {
			s.mu.Lock()
			s.observing = false
			s.mu.Unlock()
			r.finalize(s)
			return
		}
		if err != nil {
			s.mu.Lock()
			s.observing = false
			s.stopSent = false
			s.mu.Unlock()
			return
		} // retained owner; a later Stop may restart observation
	}
}

func (r *sessions) residual(s *session, stage api.RuntimeCleanupStage, d api.Diagnostic) api.RuntimeResidual {
	s.mu.Lock()
	id := s.snapshot.Data().SessionID
	s.mu.Unlock()
	return owned(api.NewRuntimeResidual(api.RuntimeResidualData{SessionID: api.Some(id), Stage: stage, Detail: d}))
}

func (r *sessions) requestStop(s *session) {
	_ = r.registry.change(s, api.StateChanged, func(d *api.SessionSnapshotData) error {
		s.stopAsked = true
		if d.Phase != api.CleanupFailed {
			d.Phase = api.Stopping
		}
		if discarded := s.input.close(); discarded != 0 {
			message := fmt.Sprintf("Accepted input: %d queued bytes discarded before native delivery; do not replay.", discarded)
			if previous, ok := s.diagnostics[api.InputCleanup]; ok {
				message = previous.Data().Message + " " + message
			}
			s.diagnostics[api.InputCleanup] = diagnostic(api.IOFailure, "runtime.input_discarded", message)
		}
		if s.snapshot.Data().Phase == api.Stopping {
			return errBusy
		}
		return nil
	})
	s.mu.Lock()
	owner := s.owner
	send := owner != nil && !s.stopSent && !s.nativeClean
	if send {
		s.stopSent = true
	}
	s.mu.Unlock()
	if send {
		owner.Stop()
	}
}

func (r *sessions) finalize(s *session) {
	_ = r.registry.change(s, api.RuntimeCleaned, func(d *api.SessionSnapshotData) error {
		if s.startPending || !s.nativeClean || s.producers != 0 {
			return errBusy
		}
		d.Phase = api.Cleaned
		if s.acquired.Present() {
			d.AcquiredCwd = s.acquired
		}
		if s.exit.Present() {
			d.Exit = s.exit
		}
		d.Cleanup = owned(api.NewSessionCleanup(api.SessionCleanupData{State: api.CleanupComplete}))
		return nil
	})
}

func (r *sessions) Snapshot(ctx context.Context, id domain.SessionID) (api.SessionSnapshot, error) {
	if ctx == nil {
		return api.SessionSnapshot{}, errInvalid
	}
	if err := ctx.Err(); err != nil {
		return api.SessionSnapshot{}, err
	}
	s, err := r.registry.lookup(id)
	if err != nil {
		return api.SessionSnapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshot.Clone(), nil
}

func (r *sessions) List(ctx context.Context, filter api.SessionFilter) (api.SessionList, error) {
	if ctx == nil {
		return api.SessionList{}, errInvalid
	}
	return r.registry.list(ctx, filter)
}

func (r *sessions) ReadOutput(ctx context.Context, request api.SessionOutputRequest) (api.SessionOutputResult, error) {
	if ctx == nil || !request.Valid() {
		return api.SessionOutputResult{}, errInvalid
	}
	if err := ctx.Err(); err != nil {
		return api.SessionOutputResult{}, err
	}
	d := request.Data()
	s, err := r.registry.lookup(d.SessionID)
	if err != nil {
		return api.SessionOutputResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.snapshot.Data().Capabilities.Data().Output {
		return api.SessionOutputResult{}, errUnsupported
	}
	return s.output.read(d.SessionID, s.snapshot.Data().Sequence, d.Offset, d.MaxBytes)
}

func (r *sessions) NextEvent(ctx context.Context, cursor ports.RuntimeEventCursor) (api.RuntimeEvent, error) {
	if ctx == nil {
		return api.RuntimeEvent{}, errInvalid
	}
	return r.registry.events.next(ctx, cursor)
}
func (r *sessions) AckEvents(cursor ports.RuntimeEventCursor) error {
	return r.registry.events.ack(cursor)
}
