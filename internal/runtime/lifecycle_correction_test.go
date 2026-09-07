package runtime

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// A context's Err barrier exposes an ordinary scheduling gap after Restart has
// looked up an old cleaned record but before its worker calls public Stop.
type evictionRestartContext struct {
	context.Context
	entered, resume chan struct{}
	once            sync.Once
}

func awaitObservationIdle(t *testing.T, s *session) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		s.mu.Lock()
		idle := !s.observing && s.latestLocked().Phase == api.CleanupFailed
		s.mu.Unlock()
		if idle {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("retained failed observer did not become idle")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSessionsShutdownRecoveryAcrossObservationError(t *testing.T) {
	for _, requestBeforeError := range []bool{false, true} {
		t.Run(map[bool]string{false: "after-error", true: "before-error"}[requestBeforeError], func(t *testing.T) {
			o := testOwner()
			failed, failNow := make(chan struct{}), make(chan struct{})
			repair, repairNow := make(chan struct{}), make(chan struct{})
			var calls, active, overlap atomic.Int32
			o.next = func(context.Context) (nativeFact, error) {
				if active.Add(1) != 1 {
					overlap.Add(1)
				}
				defer active.Add(-1)
				if calls.Add(1) == 1 {
					close(failed)
					<-failNow
					return nativeFact{}, errUnsupported
				}
				close(repair)
				<-repairNow
				return cleanedFact(), nil
			}
			receipt := &controlledReceipt{done: make(chan struct{}), observed: make(chan struct{}), delivery: nativeDelivery{Completed: true, Delivered: 1}}
			o.resize = func(context.Context, api.Geometry) (nativeDelivery, error) {
				return nativeDelivery{Dispatched: true, Receipt: receipt}, nil
			}
			r, _ := testEngine(o)
			id := startID(t, r, engineRequest(5000, true))
			<-failed
			ctx, cancel := context.WithCancel(context.Background())
			controlDone := make(chan struct{})
			go func() {
				_, _ = r.Resize(ctx, must(api.NewSessionResizeRequest(api.SessionResizeRequestData{SessionID: id, Geometry: must(api.NewGeometry(api.GeometryData{Rows: 57, Columns: 119}))})))
				close(controlDone)
			}()
			<-receipt.observed
			cancel()
			<-controlDone
			if !requestBeforeError {
				close(failNow)
				awaitObservationIdle(t, must(r.registry.lookup(id)))
			}
			// The aggregate call alone must own recovery, including the gap before
			// the active observer returns an incomplete error. Its canceled wait
			// does not cancel the retained owner or its receipt producer.
			if result := r.Shutdown(ctx); result.Data().Complete {
				t.Fatal("failed or pending owner was reported complete")
			}
			var callers sync.WaitGroup
			for i := range 12 {
				callers.Add(1)
				go func() {
					defer callers.Done()
					if i%2 == 0 {
						_, _ = r.Stop(ctx, stopRequest(id))
					} else {
						r.Shutdown(ctx)
					}
				}()
			}
			callers.Wait()
			if requestBeforeError {
				close(failNow)
			}
			<-repair
			close(repairNow)
			awaitPhase(t, r, id, api.Stopping)
			if result := r.Shutdown(context.Background()); result.Data().Complete {
				t.Fatal("native repair abandoned parent receipt")
			}
			close(receipt.done)
			awaitPhase(t, r, id, api.Cleaned)
			if calls.Load() != 2 || overlap.Load() != 0 {
				t.Fatalf("native observations calls=%d overlaps=%d", calls.Load(), overlap.Load())
			}
			final := must(r.NextEvent(context.Background(), cursor(0)))
			if final.Data().Kind != api.RuntimeCleaned {
				t.Fatal("final missing")
			}
			mustAck := r.AckEvents(final.Data().Sequence)
			if mustAck != nil || !r.Shutdown(context.Background()).Data().Complete {
				t.Fatal("repaired aggregate cleanup did not complete", mustAck)
			}
			if _, err := r.NextEvent(context.Background(), final.Data().Sequence); !errors.Is(err, io.EOF) {
				t.Fatal("producer closure did not follow final ACK", err)
			}
		})
	}
}

func TestSessionsShutdownFailureDoesNotPollOwner(t *testing.T) {
	o := testOwner()
	var calls atomic.Int32
	o.next = func(context.Context) (nativeFact, error) {
		calls.Add(1)
		return nativeFact{}, errUnsupported
	}
	r, _ := testEngine(o)
	id := startID(t, r, engineRequest(5001, false))
	s := must(r.registry.lookup(id))
	awaitObservationIdle(t, s)
	for want := int32(2); want <= 4; want++ {
		result := r.Shutdown(context.Background())
		awaitObservationIdle(t, s)
		if result.Data().Complete || calls.Load() != want || result.Data().Sessions[0].Data().Session.Data().Phase != api.CleanupFailed {
			t.Fatalf("failed repair lost residual or polled: calls=%d want=%d", calls.Load(), want)
		}
	}
}

func (c *evictionRestartContext) Err() error {
	c.once.Do(func() { close(c.entered); <-c.resume })
	return c.Context.Err()
}

func TestSessionsRestartHistoryEvictionRace(t *testing.T) {
	sessionsRestartHistory(t, true, false)
}
func TestSessionsRestartHistoryWithoutEviction(t *testing.T) {
	sessionsRestartHistory(t, false, false)
}
func TestSessionsRestartEvictionRetainsCanceledReplacement(t *testing.T) {
	sessionsRestartHistory(t, true, true)
}
func sessionsRestartHistory(t *testing.T, evict, cancelReplacement bool) {
	o := testOwner()
	replacementEntered := make(chan nativeStart, 1)
	replacementRelease := make(chan struct{})
	r := newSessions(func(ctx context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		if c.ID.Value() == historyCapacity+1 {
			return o, established(c), nil
		}
		if cancelReplacement && c.ID.Value() == historyCapacity+2 {
			replacementEntered <- c
			<-replacementRelease
			return nil, nativeStartFact{}, ctx.Err()
		}
		return nil, nativeStartFact{}, errUnsupported
	}, nil, sessionBudgets{})
	after := cursor(0)
	for i := uint64(1); i <= historyCapacity; i++ {
		result, _ := r.Start(context.Background(), engineRequest(2000+i, false))
		id := mustValue(result.Data().Session).Data().SessionID
		awaitPhase(t, r, id, api.Cleaned)
		e := must(r.NextEvent(context.Background(), after))
		after = e.Data().Sequence
		if e.Data().Kind != api.RuntimeCleaned {
			t.Fatal("fixture final missing")
		}
		if err := r.AckEvents(after); err != nil {
			t.Fatal(err)
		}
	}
	live := startID(t, r, engineRequest(3000, false))
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &evictionRestartContext{Context: base, entered: make(chan struct{}), resume: make(chan struct{})}
	done := make(chan api.SessionRestartResult, 1)
	go func() {
		result, _ := r.Restart(ctx, must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(4000)), SessionID: sessionID(1)})))
		done <- result
	}()
	<-ctx.entered
	if evict {
		o.facts <- cleanedFact()
		awaitPhase(t, r, live, api.Cleaned)
		if _, err := r.Snapshot(context.Background(), sessionID(1)); !errors.Is(err, errNotFound) {
			t.Fatal("old record did not history-evict")
		}
	}
	close(ctx.resume)
	if cancelReplacement {
		config := <-replacementEntered
		if config.ID.Value() != historyCapacity+2 {
			t.Fatal("replacement identity changed")
		}
		cancel()
	}
	result := <-done
	if !result.Valid() || !result.Data().Old.Data().CleanupComplete || !result.Data().Replacement.Present() {
		t.Fatal("retained restart result lost")
	}
	replacement := mustValue(mustValue(result.Data().Replacement).Data().Session)
	if result.Data().Old.Data().Session.Data().SessionID != sessionID(1) || replacement.Data().SessionID.Value() != historyCapacity+2 {
		t.Fatal("retained subject or replacement mismatch")
	}
	if cancelReplacement {
		if !result.Data().CancellationAsked || replacement.Data().Phase != api.Starting {
			t.Fatal("canceled admitted replacement missing")
		}
		close(replacementRelease)
		awaitPhase(t, r, replacement.Data().SessionID, api.Cleaned)
	}
	if evict {
		if _, err := r.Restart(context.Background(), must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(4001)), SessionID: sessionID(1)}))); !errors.Is(err, errNotFound) {
			t.Fatal("evicted public ID regained restart authority")
		}
	}
	if !evict {
		o.facts <- cleanedFact()
		awaitPhase(t, r, live, api.Cleaned)
	}
}
