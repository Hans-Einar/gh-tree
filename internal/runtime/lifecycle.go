package runtime

import (
	"context"
	"reflect"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

type restartTransition struct {
	request     api.SessionRestartRequest
	done        chan struct{}
	result      api.SessionRestartResult
	err         error
	replacement *session // protected by predecessor session mutex
}

func (r *sessions) Stop(ctx context.Context, request api.SessionStopRequest) (api.SessionStopResult, error) {
	if ctx == nil || !request.Valid() {
		return api.SessionStopResult{}, errInvalid
	}
	s, err := r.registry.lookup(request.Data().SessionID)
	if err != nil {
		return api.SessionStopResult{}, err
	}
	return r.stop(ctx, s)
}

// A lifecycle transition owns its admitted subject even after cleaned history
// evicts the public ID. Never resolve that subject again during the transition.
func (r *sessions) stop(ctx context.Context, s *session) (api.SessionStopResult, error) {
	s.mu.Lock()
	supported := s.snapshot.Data().Capabilities.Data().TreeStop
	s.mu.Unlock()
	if !supported {
		return s.stopResult(ctx.Err() != nil), errUnsupported
	}
	r.requestCleanup(s)
	wait, cancel := context.WithTimeout(ctx, r.budgets.grace+r.budgets.force)
	defer cancel()
	err := r.waitCleanup(wait, s)
	return s.stopResult(ctx.Err() != nil), err
}

func (r *sessions) waitCleanup(ctx context.Context, s *session) error {
	for {
		s.mu.Lock()
		phase := s.latestLocked().Phase
		observing := s.observing
		changed := s.changed
		s.mu.Unlock()
		if phase == api.Cleaned {
			return nil
		}
		if phase == api.CleanupFailed && !observing {
			return errCleanup
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *sessions) Restart(ctx context.Context, request api.SessionRestartRequest) (api.SessionRestartResult, error) {
	if ctx == nil || !request.Valid() {
		return api.SessionRestartResult{}, errInvalid
	}
	d := request.Data()
	r.admission.Lock()
	s, err := r.registry.lookup(d.SessionID)
	if err != nil {
		r.admission.Unlock()
		return api.SessionRestartResult{}, err
	}
	s.mu.Lock()
	if !s.snapshot.Data().Capabilities.Data().Restart {
		s.mu.Unlock()
		r.admission.Unlock()
		return r.restartRefused(s, ctx, errUnsupported)
	}
	if prior := s.restart; prior != nil {
		sameKey := prior.request.Data().OperationID == d.OperationID
		sameInput := reflect.DeepEqual(prior.request.Data(), d)
		s.mu.Unlock()
		r.admission.Unlock()
		if sameKey && !sameInput {
			return r.restartRefused(s, ctx, errInvalid)
		}
		if !sameKey {
			return r.restartRefused(s, ctx, errConflict)
		}
		return r.waitRestart(ctx, s, prior)
	}
	s.mu.Unlock()
	// An OperationID already retained by another lifecycle cannot be repurposed.
	r.registry.mu.Lock()
	collision := r.registry.transitions[d.OperationID] != nil
	closed := r.registry.closed
	full := len(r.registry.transitions) >= finalCapacity
	for _, other := range r.registry.sessions {
		other.mu.Lock()
		collision = collision || other.start.Data().OperationID == d.OperationID || other.restart != nil && other.restart.request.Data().OperationID == d.OperationID
		other.mu.Unlock()
	}
	r.registry.mu.Unlock()
	if closed {
		r.admission.Unlock()
		return r.restartRefused(s, ctx, errClosed)
	}
	if collision {
		r.admission.Unlock()
		return r.restartRefused(s, ctx, errInvalid)
	}
	if full {
		r.admission.Unlock()
		return r.restartRefused(s, ctx, errBusy)
	}
	if err := ctx.Err(); err != nil {
		r.admission.Unlock()
		return r.restartRefused(s, ctx, err)
	}
	transition := &restartTransition{request: request.Clone(), done: make(chan struct{})}
	s.mu.Lock()
	s.restart = transition
	s.mu.Unlock()
	r.registry.mu.Lock()
	r.registry.transitions[d.OperationID] = s
	r.registry.mu.Unlock()
	r.admission.Unlock()
	go func() {
		old, err := r.stop(ctx, s)
		result := api.SessionRestartResultData{Old: old, CancellationAsked: ctx.Err() != nil}
		if err == nil {
			err = ctx.Err()
		}
		if err == nil {
			s.mu.Lock()
			inv := s.start.Data().Invocation.Data()
			env := append([]string{}, s.environment...)
			geometry := s.snapshot.Data().Display.Data().Geometry
			s.mu.Unlock()
			if override, ok := d.Geometry.Value(); ok {
				geometry = override
			}
			inv.Geometry = geometry
			req := owned(api.NewSessionStartRequest(api.SessionStartRequestData{OperationID: d.OperationID, Invocation: owned(api.NewInvocation(inv))}))
			replacement, startErr := r.start(ctx, req, env, s)
			err = startErr
			if replacement.Data().Session.Present() {
				result.Replacement = api.Some(replacement)
			}
		}
		result.CancellationAsked = ctx.Err() != nil
		result.Diagnostics = diagnostics(err)
		transition.result = owned(api.NewSessionRestartResult(result))
		transition.err = err
		r.registry.mu.Lock()
		delete(r.registry.transitions, d.OperationID)
		close(transition.done)
		if r.registry.closed && r.registry.live == 0 && len(r.registry.transitions) == 0 {
			_ = r.registry.events.closeProducers()
		}
		r.registry.mu.Unlock()
	}()
	return r.waitRestart(ctx, s, transition)
}

func (r *sessions) restartRefused(s *session, ctx context.Context, err error) (api.SessionRestartResult, error) {
	return owned(api.NewSessionRestartResult(api.SessionRestartResultData{Old: s.stopResult(ctx.Err() != nil), CancellationAsked: ctx.Err() != nil, Diagnostics: diagnostics(err)})), err
}

func (r *sessions) waitRestart(ctx context.Context, s *session, transition *restartTransition) (api.SessionRestartResult, error) {
	select {
	case <-transition.done:
		return transition.result.Clone(), transition.err
	case <-ctx.Done():
		select {
		case <-transition.done:
			return transition.result.Clone(), transition.err
		default:
		}
		s.mu.Lock()
		replacement := transition.replacement
		s.mu.Unlock()
		result := api.SessionRestartResultData{Old: s.stopResult(true), CancellationAsked: true, Diagnostics: diagnostics(ctx.Err())}
		if replacement != nil {
			result.Replacement = api.Some(replacement.startResult(true))
		}
		return owned(api.NewSessionRestartResult(result)), ctx.Err()
	}
}

func (r *sessions) Shutdown(ctx context.Context) api.RuntimeShutdownResult {
	if ctx == nil {
		return owned(api.NewRuntimeShutdownResult(api.RuntimeShutdownResultData{Diagnostics: diagnostics(errInvalid)}))
	}
	r.admission.Lock()
	all := r.registry.closeAdmission()
	r.admission.Unlock()
	for _, s := range all {
		r.requestCleanup(s)
	}
	wait, cancel := context.WithTimeout(ctx, r.budgets.shutdown)
	defer cancel()
	for _, s := range all {
		_ = r.waitCleanup(wait, s)
	}
	for _, s := range all {
		s.mu.Lock()
		transition := s.restart
		s.mu.Unlock()
		if transition != nil {
			select {
			case <-transition.done:
			case <-wait.Done():
			}
		}
	}
	result := api.RuntimeShutdownResultData{AdmissionClosed: true, Complete: true}
	for _, s := range all {
		stop := s.stopResult(ctx.Err() != nil)
		result.Sessions = append(result.Sessions, stop)
		if !stop.Data().CleanupComplete {
			result.Complete = false
			s.mu.Lock()
			residuals := s.latestLocked().Cleanup.Data().Residuals
			s.mu.Unlock()
			if len(residuals) == 0 {
				s.mu.Lock()
				nativeClean, control := s.nativeClean, s.controlBusy
				producers := s.producers
				s.mu.Unlock()
				if nativeClean && producers > 0 {
					if control {
						residuals = append(residuals, r.residual(s, api.ControlCleanup, errCleanup))
					}
					if producers > 1 || !control {
						residuals = append(residuals, r.residual(s, api.InputCleanup, errCleanup))
					}
				} else {
					residuals = []api.RuntimeResidual{r.residual(s, api.Acquisition, errCleanup)}
				}
			}
			result.Residuals = append(result.Residuals, residuals...)
		}
	}
	r.registry.mu.Lock()
	if len(r.registry.transitions) == 0 {
		_ = r.registry.events.closeProducers()
	} else {
		result.Complete = false
		result.Residuals = append(result.Residuals, owned(api.NewRuntimeResidual(api.RuntimeResidualData{Stage: api.ControlCleanup, Detail: errCleanup})))
	}
	r.registry.mu.Unlock()
	r.registry.events.mu.Lock()
	pending := len(r.registry.events.reservations)
	if pending != 0 {
		result.Complete = false
		result.Residuals = append(result.Residuals, owned(api.NewRuntimeResidual(api.RuntimeResidualData{Stage: api.EventTransfer, Detail: diagnostic(api.CleanupIncomplete, "runtime.events_pending", "Reliable Runtime cleanup events still await transfer and acknowledgment.")})))
	}
	r.registry.events.mu.Unlock()
	return owned(api.NewRuntimeShutdownResult(result))
}
