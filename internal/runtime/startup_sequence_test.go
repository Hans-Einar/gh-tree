package runtime

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestSessionsStartupReservationSurvivesOutputStopAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name        string
		established bool
		owner       bool
		cancel      bool
		failure     error
	}{
		{name: "established", established: true, owner: true},
		{name: "known-start-plus-error", established: true, owner: true, failure: errUnsupported},
		{name: "failed-partial-acquisition", owner: true, failure: errUnsupported},
		{name: "failed-without-native-resource", failure: errUnsupported},
		{name: "canceled-known-start", established: true, owner: true, cancel: true, failure: context.Canceled},
		{name: "canceled-partial-acquisition", owner: true, cancel: true, failure: context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := testOwner()
			entered := make(chan nativeStart, 1)
			release := make(chan struct{})
			r := newSessions(func(_ context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
				entered <- c
				<-release
				if !tc.owner {
					return nil, nativeStartFact{}, tc.failure
				}
				fact := established(c)
				fact.Established = tc.established
				return o, fact, tc.failure
			}, nil, sessionBudgets{})
			request := engineRequest(9001, false)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan api.SessionStartResult, 1)
			go func() { result, _ := r.Start(ctx, request); done <- result }()
			config := <-entered
			s := must(r.registry.lookup(config.ID))
			setSessionSequence(s, math.MaxUint64-2)
			before := must(r.Snapshot(context.Background(), config.ID))
			config.Output(api.Stdout, []byte("cannot spend the startup version"))
			stopContext, endStopWait := context.WithCancel(context.Background())
			endStopWait()
			var callers sync.WaitGroup
			for range 8 {
				callers.Add(1)
				go func() { defer callers.Done(); _, _ = r.Stop(stopContext, stopRequest(config.ID)) }()
			}
			callers.Wait()
			if current := must(r.Snapshot(context.Background(), config.ID)); !reflect.DeepEqual(before, current) {
				t.Fatal("output/Stop spent startup's version or changed the published snapshot")
			}
			if tc.cancel {
				cancel()
				partial := <-done
				if !partial.Valid() || partial.Data().Established || !partial.Data().CancellationAsked || mustValue(partial.Data().Session).Data().SessionID != config.ID {
					t.Fatal("cancellation before native return lost admitted identity")
				}
			}
			close(release)
			<-s.startDone
			if !tc.cancel {
				result := <-done
				if !result.Valid() || result.Data().Established != tc.established {
					t.Fatal("known native start was misrepresented")
				}
			}
			// Repeated Start returns the native result already owned by this ID;
			// it must be coherent even when the original caller canceled earlier.
			result, err := r.Start(context.Background(), request)
			if !errors.Is(err, tc.failure) || !result.Valid() || result.Data().Established != tc.established {
				t.Fatal("native startup facts/error were lost", err)
			}
			published := mustValue(result.Data().Session)
			if published.Data().AcquiredCwd.Present() != tc.owner || published.Data().Phase == api.Starting || published.Data().Sequence.Value() <= before.Data().Sequence.Value() {
				t.Fatal("startup was not published coherently at a new sequence")
			}
			if tc.owner {
				if published.Data().Sequence.Value() != math.MaxUint64-1 || published.Data().Phase != api.Stopping {
					t.Fatal("startup and final reservations were not separate")
				}
				o.facts <- cleanedFact()
			}
			final := awaitPhase(t, r, config.ID, api.Cleaned)
			if final.Data().Sequence.Value() != math.MaxUint64 || final.Data().AcquiredCwd.Present() != tc.owner || final.Data().OutputRange.Data().End != 0 {
				t.Fatal("final truth, output refusal or sequence reservation lost")
			}
			if before.Data().Phase != api.Starting || before.Data().AcquiredCwd.Present() {
				t.Fatal("previously returned startup snapshot mutated")
			}
		})
	}
}

func TestSessionsStartupReservationHandsOffToPendingReceipts(t *testing.T) {
	o := testOwner()
	input := &controlledReceipt{done: make(chan struct{}), observed: make(chan struct{}), delivery: nativeDelivery{Completed: true, Delivered: 1}}
	control := &controlledReceipt{done: make(chan struct{}), observed: make(chan struct{}), delivery: nativeDelivery{Dispatched: true}, err: errUnsupported}
	o.write = func(context.Context, []byte) (nativeDelivery, error) {
		return nativeDelivery{Dispatched: true, Receipt: input}, nil
	}
	o.resize = func(context.Context, api.Geometry) (nativeDelivery, error) {
		return nativeDelivery{Dispatched: true, Receipt: control}, nil
	}
	entered := make(chan nativeStart, 1)
	release := make(chan struct{})
	r := newSessions(func(_ context.Context, c nativeStart) (nativeOwner, nativeStartFact, error) {
		entered <- c
		<-release
		return o, established(c), nil
	}, nil, sessionBudgets{})
	started := make(chan api.SessionStartResult, 1)
	go func() { result, _ := r.Start(context.Background(), engineRequest(9002, true)); started <- result }()
	config := <-entered
	s := must(r.registry.lookup(config.ID))
	setSessionSequence(s, math.MaxUint64-3)
	close(release)
	result := <-started
	before := mustValue(result.Data().Session)
	if !result.Data().Established || before.Data().Sequence.Value() != math.MaxUint64-2 {
		t.Fatal("startup did not consume exactly its publication")
	}
	id := config.ID
	must(r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: []byte{1}}))))
	<-input.observed
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_, _ = r.Resize(ctx, must(api.NewSessionResizeRequest(api.SessionResizeRequestData{SessionID: id, Geometry: must(api.NewGeometry(api.GeometryData{Rows: 70, Columns: 120}))})))
		close(done)
	}()
	<-control.observed
	cancel()
	<-done
	config.Output(api.TerminalOutput, []byte("control and final own remaining versions"))
	o.facts <- cleanedFact()
	// The first final attempt must wait for both admitted receipt owners.
	if r.Shutdown(ctx).Data().Complete {
		t.Fatal("startup handoff lost outstanding producers")
	}
	close(control.done)
	if r.Shutdown(ctx).Data().Complete {
		t.Fatal("terminal-unknown control abandoned input ownership")
	}
	close(input.done)
	final := awaitPhase(t, r, id, api.Cleaned)
	if final.Data().Sequence.Value() != math.MaxUint64 || final.Data().Display.Data().Geometry != before.Data().Display.Data().Geometry || final.Data().OutputRange.Data().End != 0 || !final.Data().AcquiredCwd.Present() {
		t.Fatal("final lost startup/control/output facts")
	}
}
