package runtime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type controlledOwner struct {
	facts  chan nativeFact
	stops  atomic.Int32
	write  func(context.Context, []byte) (nativeDelivery, error)
	resize func(context.Context, api.Geometry) (nativeDelivery, error)
}

func testOwner() *controlledOwner { return &controlledOwner{facts: make(chan nativeFact, 8)} }
func (o *controlledOwner) NextFact(ctx context.Context) (nativeFact, error) {
	select {
	case f := <-o.facts:
		return f, nil
	case <-ctx.Done():
		return nativeFact{}, ctx.Err()
	}
}
func (o *controlledOwner) Write(ctx context.Context, b []byte) (nativeDelivery, error) {
	if o.write != nil {
		return o.write(ctx, b)
	}
	return nativeDelivery{Accepted: uint32(len(b)), Delivered: uint32(len(b)), Completed: true}, nil
}
func (o *controlledOwner) Resize(ctx context.Context, g api.Geometry) (nativeDelivery, error) {
	if o.resize != nil {
		return o.resize(ctx, g)
	}
	return nativeDelivery{Delivered: 1, Completed: true}, nil
}
func (o *controlledOwner) Interrupt(ctx context.Context) (nativeDelivery, error) {
	return nativeDelivery{Delivered: 1, Completed: true}, nil
}
func (o *controlledOwner) Stop() { o.stops.Add(1) }

type controlledReceipt struct {
	done     chan struct{}
	observed chan struct{}
	delivery nativeDelivery
	err      error
	once     sync.Once
}

func (r *controlledReceipt) Wait(ctx context.Context) (nativeDelivery, error) {
	r.once.Do(func() { close(r.observed) })
	select {
	case <-r.done:
		return r.delivery, r.err
	case <-ctx.Done():
		return nativeDelivery{}, ctx.Err()
	}
}

func engineRequest(n uint64, terminal bool) api.SessionStartRequest {
	r := newRegistry()
	s := must(fixtureAdmit(r, nil))
	d := s.start.Data()
	d.OperationID = must(api.NewOperationID(n))
	inv := d.Invocation.Data()
	if terminal {
		inv.Terminal = api.Terminal
	}
	d.Invocation = must(api.NewInvocation(inv))
	return must(api.NewSessionStartRequest(d))
}
func established(c nativeStart) nativeStartFact {
	return nativeStartFact{Established: true, Cwd: api.Some(must(api.NewAcquiredCwd(api.AcquiredCwdData{Observation: c.Invocation.Data().Cwd})))}
}
func testEngine(o *controlledOwner) (*sessions, <-chan nativeStart) {
	configs := make(chan nativeStart, 8)
	return newSessions(func(ctx context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		configs <- c
		return o, established(c), nil
	}, []string{"BASE=original"}, sessionBudgets{grace: time.Millisecond, force: 30 * time.Millisecond, shutdown: 30 * time.Millisecond}), configs
}
func startID(t *testing.T, r *sessions, request api.SessionStartRequest) domain.SessionID {
	t.Helper()
	result, err := r.Start(context.Background(), request)
	if err != nil || !result.Valid() || !result.Data().Established {
		t.Fatalf("Start: %v %v", result, err)
	}
	return mustValue(result.Data().Session).Data().SessionID
}
func mustValue[T any](v api.Optional[T]) T {
	value, ok := v.Value()
	if !ok {
		panic("absent fixture value")
	}
	return value
}
func awaitPhase(t *testing.T, r *sessions, id domain.SessionID, phase api.SessionPhase) api.SessionSnapshot {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	s := must(r.registry.lookup(id))
	for {
		s.mu.Lock()
		snapshot := s.snapshot
		changed := s.changed
		s.mu.Unlock()
		if snapshot.Data().Phase == phase {
			return snapshot
		}
		select {
		case <-changed:
		case <-ctx.Done():
			t.Fatalf("phase wanted %v got %v", phase, snapshot.Data().Phase)
		}
	}
}
func cleanedFact() nativeFact { return nativeFact{CleanupComplete: true} }
func stopRequest(id domain.SessionID) api.SessionStopRequest {
	return must(api.NewSessionStopRequest(api.SessionStopRequestData{OperationID: must(api.NewOperationID(999)), SessionID: id}))
}

func TestSessionsRootExitLateOutputFinalAndACK(t *testing.T) {
	o := testOwner()
	r, configs := testEngine(o)
	id := startID(t, r, engineRequest(1, false))
	config := <-configs
	exit := must(api.NewSessionExit(api.SessionExitData{Code: api.Some(0), Cause: api.NaturalExit}))
	o.facts <- nativeFact{Exit: api.Some(exit)}
	awaitPhase(t, r, id, api.Stopping)
	config.Output(api.Stdout, []byte("after root exit\x00\xff"))
	output := must(r.ReadOutput(context.Background(), must(api.NewSessionOutputRequest(api.SessionOutputRequestData{SessionID: id, MaxBytes: 100})))).Data()
	if len(output.Chunks) != 1 || !bytes.Equal(output.Chunks[0].Data().Bytes, []byte("after root exit\x00\xff")) {
		t.Fatal("raw late output lost")
	}
	o.facts <- cleanedFact()
	snapshot := awaitPhase(t, r, id, api.Cleaned)
	if snapshot.Data().OutputRange.Data().End != 17 || !snapshot.Data().Exit.Present() {
		t.Fatal("final lost facts")
	}
	shutdown := r.Shutdown(context.Background())
	if shutdown.Data().Complete {
		t.Fatal("unacknowledged final hidden")
	}
	final := must(r.NextEvent(context.Background(), cursor(0)))
	replay := must(r.NextEvent(context.Background(), cursor(0)))
	if final.Data().Kind != api.RuntimeCleaned || !reflect.DeepEqual(final, replay) {
		t.Fatal("final replay")
	}
	if err := r.AckEvents(final.Data().Sequence); err != nil {
		t.Fatal(err)
	}
	if !r.Shutdown(context.Background()).Data().Complete {
		t.Fatal("complete shutdown refused")
	}
	if _, err := r.NextEvent(context.Background(), final.Data().Sequence); !errors.Is(err, io.EOF) {
		t.Fatal("EOF", err)
	}
}

func TestSessionsCanceledAdmissionAndEstablishment(t *testing.T) {
	entered := make(chan nativeStart, 1)
	release := make(chan struct{})
	o := testOwner()
	r := newSessions(func(ctx context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		entered <- c
		<-release
		return o, nativeStartFact{}, ctx.Err()
	}, nil, sessionBudgets{})
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := r.Start(canceled, engineRequest(1, false))
	if !errors.Is(err, context.Canceled) || result.Data().Session.Present() {
		t.Fatal("admitted canceled request")
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan api.SessionStartResult, 1)
	go func() { result, _ := r.Start(ctx, engineRequest(2, false)); done <- result }()
	config := <-entered
	cancel()
	result = <-done
	if !result.Data().Session.Present() || result.Data().Established || !result.Data().CancellationAsked {
		t.Fatal("lost canceled admission")
	}
	close(release)
	awaitPhase(t, r, config.ID, api.Stopping)
	o.facts <- cleanedFact()
	awaitPhase(t, r, config.ID, api.Cleaned)
	o2 := testOwner()
	r2, _ := testEngine(o2)
	ctx2, cancel2 := context.WithCancel(context.Background())
	result, err = r2.Start(ctx2, engineRequest(3, false))
	if err != nil {
		t.Fatal(err)
	}
	cancel2()
	id := mustValue(result.Data().Session).Data().SessionID
	if snapshot := must(r2.Snapshot(context.Background(), id)); snapshot.Data().Phase != api.Running || o2.stops.Load() != 0 {
		t.Fatal("original start context killed established owner")
	}
	o2.facts <- cleanedFact()
	awaitPhase(t, r2, id, api.Cleaned)
}

func TestSessionsInputReceiptRetainsCapacityAndFinalBarrier(t *testing.T) {
	o := testOwner()
	receipt := &controlledReceipt{done: make(chan struct{}), observed: make(chan struct{}), delivery: nativeDelivery{Completed: true, Delivered: nativeInputChunk}}
	written := make(chan []byte, 2)
	o.write = func(ctx context.Context, b []byte) (nativeDelivery, error) {
		written <- append([]byte(nil), b...)
		return nativeDelivery{Dispatched: true, Receipt: receipt}, context.Canceled
	}
	r, _ := testEngine(o)
	id := startID(t, r, engineRequest(1, false))
	data := bytes.Repeat([]byte{7}, inputCapacity)
	result := must(r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: data}))))
	data[0] = 9
	if result.Data().AcceptedBytes != inputCapacity {
		t.Fatal("partial parent admission")
	}
	first := <-written
	<-receipt.observed
	if len(first) != nativeInputChunk || first[0] != 7 {
		t.Fatal("split/copy")
	}
	busy, err := r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: []byte{1}})))
	if !errors.Is(err, errBusy) || busy.Data().AcceptedBytes != 0 {
		t.Fatal("in-flight capacity released")
	}
	o.facts <- cleanedFact()
	awaitPhase(t, r, id, api.Stopping)
	stop, err := r.Stop(context.Background(), stopRequest(id))
	if err == nil || stop.Data().CleanupComplete {
		t.Fatal("pending parent receipt fabricated cleanup")
	}
	close(receipt.done)
	awaitPhase(t, r, id, api.Cleaned)
	select {
	case <-written:
		t.Fatal("replayed/sent canceled remainder")
	default:
	}
	s := must(r.registry.lookup(id))
	s.mu.Lock()
	diagnostics := s.diagnosticsLocked()
	s.mu.Unlock()
	if len(diagnostics) == 0 {
		t.Fatal("partial delivery not retained")
	}
}

func TestSessionsWriteSplitsOneRequestExactlyOnce(t *testing.T) {
	o := testOwner()
	sizes := make(chan int, 2)
	o.write = func(ctx context.Context, b []byte) (nativeDelivery, error) {
		sizes <- len(b)
		return nativeDelivery{Completed: true, Delivered: uint32(len(b))}, nil
	}
	r, _ := testEngine(o)
	id := startID(t, r, engineRequest(1, false))
	must(r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: make([]byte, inputCapacity)}))))
	if a, b := <-sizes, <-sizes; a != 65484 || b != 52 {
		t.Fatalf("splits %d %d", a, b)
	}
	o.facts <- cleanedFact()
	awaitPhase(t, r, id, api.Cleaned)
}

func TestSessionsPendingResizeJoinedBeforeRestart(t *testing.T) {
	first, second := testOwner(), testOwner()
	receipt := &controlledReceipt{done: make(chan struct{}), observed: make(chan struct{}), delivery: nativeDelivery{Completed: true, Delivered: 1}}
	first.resize = func(ctx context.Context, g api.Geometry) (nativeDelivery, error) {
		return nativeDelivery{Dispatched: true, Receipt: receipt}, ctx.Err()
	}
	configs := make(chan nativeStart, 2)
	var count atomic.Int32
	r := newSessions(func(ctx context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		configs <- c
		if count.Add(1) == 1 {
			return first, established(c), nil
		}
		return second, established(c), nil
	}, nil, sessionBudgets{grace: time.Millisecond, force: 20 * time.Millisecond})
	id := startID(t, r, engineRequest(1, true))
	<-configs
	geometry := must(api.NewGeometry(api.GeometryData{Rows: 40, Columns: 100}))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan api.SessionControlResult, 1)
	go func() {
		result, _ := r.Resize(ctx, must(api.NewSessionResizeRequest(api.SessionResizeRequestData{SessionID: id, Geometry: geometry})))
		done <- result
	}()
	<-receipt.observed
	cancel()
	result := <-done
	if !result.Data().CancellationAsked || result.Data().Delivered {
		t.Fatal("pending control truthfulness")
	}
	first.facts <- cleanedFact()
	awaitPhase(t, r, id, api.Stopping)
	close(receipt.done)
	final := awaitPhase(t, r, id, api.Cleaned)
	if final.Data().Display.Data().Geometry != geometry {
		t.Fatal("eventual accepted geometry lost")
	}
	req := must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(2)), SessionID: id}))
	restart := must(r.Restart(context.Background(), req))
	replacement := mustValue(mustValue(restart.Data().Replacement).Data().Session)
	config := <-configs
	if replacement.Data().SessionID == id || config.Invocation.Data().Geometry != geometry {
		t.Fatal("restart geometry/identity")
	}
	again := must(r.Restart(context.Background(), req))
	if !reflect.DeepEqual(restart, again) || count.Load() != 2 {
		t.Fatal("restart duplicate")
	}
	changed := req.Data()
	changed.Geometry = api.Some(must(api.NewGeometry(api.GeometryData{Rows: 10, Columns: 10})))
	if _, err := r.Restart(context.Background(), must(api.NewSessionRestartRequest(changed))); !errors.Is(err, errInvalid) {
		t.Fatal("changed restart key")
	}
	second.facts <- cleanedFact()
	awaitPhase(t, r, replacement.Data().SessionID, api.Cleaned)
}

func TestSessionsShutdownCapturesBlockedStartAndClosesAdmission(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	o := testOwner()
	r := newSessions(func(ctx context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		close(entered)
		<-release
		return o, established(c), nil
	}, nil, sessionBudgets{shutdown: time.Millisecond})
	started := make(chan api.SessionStartResult, 1)
	go func() { result, _ := r.Start(context.Background(), engineRequest(1, false)); started <- result }()
	<-entered
	result := r.Shutdown(context.Background())
	if result.Data().Complete || len(result.Data().Sessions) != 1 || !result.Data().AdmissionClosed {
		t.Fatal("Starting omitted")
	}
	if refused, err := r.Start(context.Background(), engineRequest(2, false)); !errors.Is(err, errClosed) || refused.Data().Session.Present() {
		t.Fatal("late admission")
	}
	close(release)
	start := <-started
	id := mustValue(start.Data().Session).Data().SessionID
	awaitPhase(t, r, id, api.Stopping)
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, _ = r.Stop(ctx, stopRequest(id))
		}()
	}
	wg.Wait()
	if o.stops.Load() != 1 {
		t.Fatal("Stop did not coalesce", o.stops.Load())
	}
	o.facts <- cleanedFact()
	awaitPhase(t, r, id, api.Cleaned)
}

func TestSessionsFailedCleanupRepairAndCanceledRestart(t *testing.T) {
	o := testOwner()
	r, _ := testEngine(o)
	id := startID(t, r, engineRequest(1, false))
	residual := r.residual(must(r.registry.lookup(id)), api.OutputCleanup, errCleanup)
	o.facts <- nativeFact{Residuals: []api.RuntimeResidual{residual}}
	awaitPhase(t, r, id, api.CleanupFailed)
	if stopped, err := r.Stop(context.Background(), stopRequest(id)); err == nil || stopped.Data().CleanupComplete {
		t.Fatal("residual hidden")
	}
	o.facts <- cleanedFact()
	awaitPhase(t, r, id, api.Cleaned)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(2)), SessionID: id}))
	if restart, err := r.Restart(ctx, request); !errors.Is(err, context.Canceled) || restart.Data().Replacement.Present() {
		t.Fatal("canceled replacement admission")
	}
}

func TestSessionsSequenceExhaustionKeepsFinalAndNativeOwnership(t *testing.T) {
	o := testOwner()
	r, configs := testEngine(o)
	id := startID(t, r, engineRequest(1, false))
	config := <-configs
	s := must(r.registry.lookup(id))
	s.mu.Lock()
	d := s.snapshot.Data()
	d.Sequence = sequence(math.MaxUint64 - 1)
	s.snapshot = must(api.NewSessionSnapshot(d))
	s.mu.Unlock()
	config.Output(api.Stdout, []byte("must refuse without offset movement"))
	if o.stops.Load() != 1 {
		t.Fatal("exhaustion did not latch owned Stop")
	}
	exit := must(api.NewSessionExit(api.SessionExitData{Code: api.Some(5), Cause: api.RequestedExit}))
	o.facts <- nativeFact{Exit: api.Some(exit), CleanupComplete: true}
	final := awaitPhase(t, r, id, api.Cleaned).Data()
	if final.Sequence.Value() != math.MaxUint64 || final.OutputRange.Data().End != 0 || !final.Exit.Present() {
		t.Fatal("exhausted state/final reservation")
	}
}

func TestSessionsCanceledRestartRetainsAlreadyAdmittedReplacement(t *testing.T) {
	first, second := testOwner(), testOwner()
	entered := make(chan nativeStart, 1)
	release := make(chan struct{})
	var starts atomic.Int32
	r := newSessions(func(ctx context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		if starts.Add(1) == 1 {
			return first, established(c), nil
		}
		entered <- c
		<-release
		return second, nativeStartFact{}, ctx.Err()
	}, nil, sessionBudgets{})
	id := startID(t, r, engineRequest(1, false))
	first.facts <- cleanedFact()
	awaitPhase(t, r, id, api.Cleaned)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan api.SessionRestartResult, 1)
	req := must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(2)), SessionID: id}))
	go func() { result, _ := r.Restart(ctx, req); done <- result }()
	replacementConfig := <-entered
	cancel()
	result := <-done
	if !result.Data().Old.Data().CleanupComplete || !result.Data().Replacement.Present() {
		t.Fatal("canceled restart lost admitted replacement")
	}
	replacement := mustValue(mustValue(result.Data().Replacement).Data().Session)
	if replacement.Data().SessionID != replacementConfig.ID || replacement.Data().Phase != api.Starting {
		t.Fatal("wrong replacement facts")
	}
	close(release)
	awaitPhase(t, r, replacementConfig.ID, api.Stopping)
	second.facts <- cleanedFact()
	awaitPhase(t, r, replacementConfig.ID, api.Cleaned)
}

func TestSessionsFailureBeforeNativeAcquisitionStillOwnsFinal(t *testing.T) {
	var starts atomic.Int32
	r := newSessions(func(context.Context, nativeStart) (nativeOwner, nativeStartFact, error) {
		starts.Add(1)
		return nil, nativeStartFact{}, errUnsupported
	}, nil, sessionBudgets{})
	request := engineRequest(1, false)
	result, err := r.Start(context.Background(), request)
	if !errors.Is(err, errUnsupported) || !result.Data().Session.Present() || result.Data().Established {
		t.Fatal("failed admitted start")
	}
	id := mustValue(result.Data().Session).Data().SessionID
	awaitPhase(t, r, id, api.Cleaned)
	repeated, err := r.Start(context.Background(), request)
	if !errors.Is(err, errUnsupported) || mustValue(repeated.Data().Session).Data().SessionID != id || starts.Load() != 1 {
		t.Fatal("failed start duplicate reran native acquisition")
	}
}
