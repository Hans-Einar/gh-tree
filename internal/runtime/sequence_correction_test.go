package runtime

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// Only the numerical counter is positioned directly; every subsequent effect,
// observation, publication, cleanup and restart runs through the common engine.
func setSessionSequence(s *session, n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.snapshot.Data()
	d.Sequence = sequence(n)
	s.snapshot = must(api.NewSessionSnapshot(d))
}

type countedControlOwner struct {
	*controlledOwner
	interrupts atomic.Int32
}

func (o *countedControlOwner) Interrupt(context.Context) (nativeDelivery, error) {
	o.interrupts.Add(1)
	return nativeDelivery{Completed: true, Delivered: 1}, nil
}

func TestSessionsSequenceRefusesOptionalEffectsBeforeNativeDelivery(t *testing.T) {
	o := &countedControlOwner{controlledOwner: testOwner()}
	var resizes, writes atomic.Int32
	o.resize = func(context.Context, api.Geometry) (nativeDelivery, error) {
		resizes.Add(1)
		return nativeDelivery{Completed: true, Delivered: 1}, nil
	}
	o.write = func(_ context.Context, b []byte) (nativeDelivery, error) {
		writes.Add(1)
		return nativeDelivery{Completed: true, Delivered: uint32(len(b))}, nil
	}
	r := newSessions(func(_ context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		return o, established(c), nil
	}, nil, sessionBudgets{})
	id := startID(t, r, engineRequest(6000, true))
	setSessionSequence(must(r.registry.lookup(id)), math.MaxUint64-1)
	before := must(r.Snapshot(context.Background(), id))
	g := must(api.NewGeometry(api.GeometryData{Rows: 57, Columns: 119}))
	resize, err := r.Resize(context.Background(), must(api.NewSessionResizeRequest(api.SessionResizeRequestData{SessionID: id, Geometry: g})))
	if !errors.Is(err, errExhausted) || resize.Data().Delivered {
		t.Fatal("unrecordable resize admitted", err)
	}
	interrupt, err := r.Interrupt(context.Background(), id)
	if !errors.Is(err, errExhausted) || interrupt.Data().Delivered {
		t.Fatal("unrecordable interrupt admitted", err)
	}
	write, err := r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: []byte("refused")})))
	if !errors.Is(err, errExhausted) || write.Data().AcceptedBytes != 0 {
		t.Fatal("exhausted input admission accepted bytes", err)
	}
	if resizes.Load() != 0 || o.interrupts.Load() != 0 || writes.Load() != 0 {
		t.Fatal("refusal happened after native effect")
	}
	o.facts <- cleanedFact()
	final := awaitPhase(t, r, id, api.Cleaned)
	if final.Data().Sequence.Value() != math.MaxUint64 || final.Data().Display.Data().Geometry != before.Data().Display.Data().Geometry {
		t.Fatal("final sequence or refusal truth changed")
	}
}

func TestSessionsResizeReservationSurvivesOutputAndNativeCleanup(t *testing.T) {
	first, second := testOwner(), testOwner()
	receipt := &controlledReceipt{done: make(chan struct{}), observed: make(chan struct{}), delivery: nativeDelivery{Completed: true, Delivered: 1}, err: errUnsupported}
	first.resize = func(context.Context, api.Geometry) (nativeDelivery, error) {
		return nativeDelivery{Dispatched: true, Receipt: receipt}, nil
	}
	configs := make(chan nativeStart, 2)
	var starts atomic.Int32
	r := newSessions(func(_ context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		configs <- c
		if starts.Add(1) == 1 {
			return first, established(c), nil
		}
		return second, established(c), nil
	}, nil, sessionBudgets{})
	id := startID(t, r, engineRequest(6001, true))
	config := <-configs
	s := must(r.registry.lookup(id))
	setSessionSequence(s, math.MaxUint64-3)
	before := must(r.Snapshot(context.Background(), id))
	g := must(api.NewGeometry(api.GeometryData{Rows: 57, Columns: 119}))
	done := make(chan controlCompletion, 1)
	go func() {
		result, err := r.Resize(context.Background(), must(api.NewSessionResizeRequest(api.SessionResizeRequestData{SessionID: id, Geometry: g})))
		done <- controlCompletion{result, err}
	}()
	<-receipt.observed
	config.Output(api.TerminalOutput, []byte{0, 255, 27})
	outputSnapshot := must(r.Snapshot(context.Background(), id))
	if outputSnapshot.Data().Sequence.Value() != math.MaxUint64-2 || outputSnapshot.Data().OutputRange.Data().End != 3 {
		t.Fatal("available output version was not used")
	}
	// The second callback has no unreserved hint version. It must refuse raw
	// offset movement, while the delivered resize and final retain their slots.
	config.Output(api.TerminalOutput, []byte("cannot consume control reservation"))
	exit := must(api.NewSessionExit(api.SessionExitData{Code: api.Some(7), Cause: api.RequestedExit}))
	first.facts <- nativeFact{Exit: api.Some(exit), CleanupComplete: true}
	limit := time.Now().Add(time.Second)
	for {
		s.mu.Lock()
		clean := s.nativeClean && !s.observing
		s.mu.Unlock()
		if clean {
			break
		}
		if time.Now().After(limit) {
			t.Fatal("native cleanup fact was not consumed")
		}
		time.Sleep(time.Millisecond)
	}
	if snapshot := must(r.Snapshot(context.Background(), id)); !reflect.DeepEqual(snapshot, outputSnapshot) {
		t.Fatal("unpublished facts reused or spent a reserved sequence")
	}
	close(receipt.done)
	completion := <-done
	if !completion.result.Data().Delivered || !errors.Is(completion.err, errUnsupported) || completion.result.Data().Sequence.Value() != math.MaxUint64-1 {
		t.Fatal("control slot or known effect plus supplemental error lost", completion.err)
	}
	final := awaitPhase(t, r, id, api.Cleaned)
	if final.Data().Sequence.Value() != math.MaxUint64 || final.Data().Display.Data().Geometry != g || mustValue(final.Data().Exit) != exit || final.Data().OutputRange.Data().End != 3 {
		t.Fatal("final did not merge delivered geometry and unavoidable native facts")
	}
	if before.Data().Sequence.Value() != math.MaxUint64-3 || before.Data().Display.Data().Geometry == g {
		t.Fatal("previously returned snapshot mutated")
	}
	restart := must(r.Restart(context.Background(), must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(6002)), SessionID: id}))))
	replacement := <-configs
	if replacement.Invocation.Data().Geometry != g || starts.Load() != 2 || !restart.Data().Old.Data().CleanupComplete {
		t.Fatal("restart lost latest delivered geometry or duplicated replacement")
	}
	second.facts <- cleanedFact()
	awaitPhase(t, r, replacement.ID, api.Cleaned)
}

func TestSessionsExhaustedNativeResidualRemainsOwnedUntilRepair(t *testing.T) {
	o := testOwner()
	failed, failNow := make(chan struct{}), make(chan struct{})
	repair, repairNow := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	var residual api.RuntimeResidual
	exit := must(api.NewSessionExit(api.SessionExitData{Code: api.Some(19), Cause: api.NaturalExit}))
	o.next = func(context.Context) (nativeFact, error) {
		if calls.Add(1) == 1 {
			close(failed)
			<-failNow
			return nativeFact{Exit: api.Some(exit), Residuals: []api.RuntimeResidual{residual}}, errUnsupported
		}
		close(repair)
		<-repairNow
		return cleanedFact(), nil
	}
	r, _ := testEngine(o)
	id := startID(t, r, engineRequest(6003, false))
	s := must(r.registry.lookup(id))
	<-failed
	residual = r.residual(s, api.Descendants, errCleanup)
	setSessionSequence(s, math.MaxUint64-1)
	before := must(r.Snapshot(context.Background(), id))
	close(failNow)
	awaitObservationIdle(t, s)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := r.Shutdown(ctx)
	<-repair
	found := false
	for _, item := range result.Data().Residuals {
		found = found || reflect.DeepEqual(item, residual)
	}
	if result.Data().Complete || !found || !reflect.DeepEqual(before, must(r.Snapshot(context.Background(), id))) {
		t.Fatal("exhaustion dropped retained residual or reused public sequence")
	}
	close(repairNow)
	final := awaitPhase(t, r, id, api.Cleaned)
	if final.Data().Sequence.Value() != math.MaxUint64 || mustValue(final.Data().Exit) != exit {
		t.Fatal("repaired final lost earlier unavoidable exit")
	}
}
