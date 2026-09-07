package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// A context's Err barrier exposes an ordinary scheduling gap after Restart has
// looked up an old cleaned record but before its worker calls public Stop.
type evictionRestartContext struct {
	context.Context
	entered, resume chan struct{}
	once            sync.Once
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
