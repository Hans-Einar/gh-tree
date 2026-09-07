package runtime

import (
	"context"
	"fmt"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// This bound is the Windows frame's actual maximum payload. Retain one parent
// queue reservation through all native splits and receipts of a public request.
const nativeInputChunk = 65484

func (r *sessions) Write(ctx context.Context, request api.SessionWriteRequest) (api.SessionWriteResult, error) {
	if ctx == nil || !request.Valid() {
		return api.SessionWriteResult{}, errInvalid
	}
	d := request.Data()
	s, err := r.registry.lookup(d.SessionID)
	if err != nil {
		return api.SessionWriteResult{}, err
	}
	s.mu.Lock()
	snapshot := s.snapshot.Data()
	if !snapshot.Capabilities.Data().Input {
		err = errUnsupported
	} else if snapshot.Phase != api.Running || s.stopAsked {
		err = errClosed
	} else {
		err = s.input.accept(ctx, d.Bytes)
	}
	n := uint32(0)
	if err == nil {
		n = uint32(len(d.Bytes))
	}
	s.mu.Unlock()
	return owned(api.NewSessionWriteResult(api.SessionWriteResultData{SessionID: d.SessionID, Sequence: snapshot.Sequence, AcceptedBytes: n, CancellationAsked: ctx.Err() != nil, Diagnostics: diagnostics(err)})), err
}

func awaitDelivery(d nativeDelivery, err error) (nativeDelivery, error) {
	if d.Receipt != nil {
		return d.Receipt.Wait(context.Background())
	}
	return d, err
}

func (r *sessions) writeLoop(s *session) {
	defer r.producerDone(s)
	for {
		data, err := s.input.next(context.Background())
		if err != nil {
			return
		}
		delivered := uint32(0)
		for offset := 0; offset < len(data); {
			s.mu.Lock()
			owner, stopping := s.owner, s.stopAsked
			s.mu.Unlock()
			if stopping {
				err = errClosed
				break
			}
			end := min(offset+nativeInputChunk, len(data))
			fact, callErr := owner.Write(context.Background(), data[offset:end])
			fact, callErr = awaitDelivery(fact, callErr)
			delivered += fact.Delivered
			if callErr != nil || !fact.Completed || fact.Delivered != uint32(end-offset) {
				err = callErr
				if err == nil {
					err = diagnostic(api.Indeterminate, "runtime.input_unknown", "Accepted input lacks a complete native delivery observation; do not replay.")
				}
				break
			}
			offset = end
		}
		if err != nil {
			d := diagnostic(safeDiagnostic(err).Data().Code, "runtime.input_delivery", fmt.Sprintf("Accepted input: %d bytes; observed native delivery: %d bytes. No automatic replay.", len(data), delivered))
			s.mu.Lock()
			if previous, ok := s.diagnostics[api.InputCleanup]; ok {
				d = diagnostic(d.Data().Code, "runtime.input_delivery", previous.Data().Message+" "+d.Data().Message)
			}
			s.diagnostics[api.InputCleanup] = d
			s.mu.Unlock()
			_ = r.registry.change(s, api.StateChanged, func(d *api.SessionSnapshotData) error { return nil })
		}
		_ = s.input.finish()
		if err != nil {
			r.requestStop(s)
			return
		}
	}
}

func (r *sessions) producerDone(s *session) {
	s.mu.Lock()
	s.producers--
	close(s.changed)
	s.changed = make(chan struct{})
	s.mu.Unlock()
	r.finalize(s)
}

func (r *sessions) Resize(ctx context.Context, request api.SessionResizeRequest) (api.SessionControlResult, error) {
	if !request.Valid() {
		return api.SessionControlResult{}, errInvalid
	}
	d := request.Data()
	return r.control(ctx, d.SessionID, api.Some(d.Geometry))
}
func (r *sessions) Interrupt(ctx context.Context, id domain.SessionID) (api.SessionControlResult, error) {
	return r.control(ctx, id, api.None[api.Geometry]())
}

type controlCompletion struct {
	result api.SessionControlResult
	err    error
}

func (r *sessions) control(ctx context.Context, id domain.SessionID, geometry api.Optional[api.Geometry]) (api.SessionControlResult, error) {
	if ctx == nil {
		return api.SessionControlResult{}, errInvalid
	}
	s, err := r.registry.lookup(id)
	if err != nil {
		return api.SessionControlResult{}, err
	}
	s.mu.Lock()
	d := s.snapshot.Data()
	caps := d.Capabilities.Data()
	g, resize := geometry.Value()
	if ctx.Err() != nil {
		err = ctx.Err()
	} else if resize && !caps.Resize || !resize && !caps.TerminalETX {
		err = errUnsupported
	} else if d.Phase != api.Running || s.stopAsked {
		err = errClosed
	} else if s.controlBusy {
		err = errBusy
	}
	if err != nil {
		s.mu.Unlock()
		return owned(api.NewSessionControlResult(api.SessionControlResultData{SessionID: id, Sequence: d.Sequence, CancellationAsked: ctx.Err() != nil, Diagnostics: diagnostics(err)})), err
	}
	s.controlBusy = true
	s.producers++
	owner := s.owner
	s.mu.Unlock()
	done := make(chan controlCompletion, 1)
	go func() {
		defer func() { s.mu.Lock(); s.controlBusy = false; s.mu.Unlock(); r.producerDone(s) }()
		var fact nativeDelivery
		var err error
		if resize {
			fact, err = owner.Resize(ctx, g)
		} else {
			fact, err = owner.Interrupt(ctx)
		}
		fact, err = awaitDelivery(fact, err)
		if !fact.Completed && fact.Dispatched && err == nil {
			err = diagnostic(api.Indeterminate, "runtime.control_unknown", "Native control delivery remains unknown; do not replay.")
		}
		delivered := fact.Completed && fact.Delivered > 0
		_ = r.registry.change(s, api.StateChanged, func(d *api.SessionSnapshotData) error {
			if resize && delivered {
				display := d.Display.Data()
				display.Geometry = g
				d.Display = owned(api.NewInvocationSummary(display))
			}
			if err != nil {
				s.diagnostics[api.ControlCleanup] = safeDiagnostic(err)
			}
			return nil
		})
		s.mu.Lock()
		seq := s.snapshot.Data().Sequence
		s.mu.Unlock()
		done <- controlCompletion{owned(api.NewSessionControlResult(api.SessionControlResultData{SessionID: id, Sequence: seq, Delivered: delivered, CancellationAsked: ctx.Err() != nil, Diagnostics: diagnostics(err)})), err}
	}()
	select {
	case completion := <-done:
		return completion.result, completion.err
	case <-ctx.Done():
		select {
		case completion := <-done:
			return completion.result, completion.err
		default:
		}
		s.mu.Lock()
		seq := s.snapshot.Data().Sequence
		s.mu.Unlock()
		diagnostic := diagnostic(api.Indeterminate, "runtime.control_pending", "Caller waiting ended while Runtime retains native delivery observation; do not replay.")
		return owned(api.NewSessionControlResult(api.SessionControlResultData{SessionID: id, Sequence: seq, CancellationAsked: true, Diagnostics: []api.Diagnostic{diagnostic}})), ctx.Err()
	}
}
