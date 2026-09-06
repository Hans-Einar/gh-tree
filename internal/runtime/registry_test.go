package runtime

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func fixtureAdmit(r *registry, env []string) (*session, error) {
	snapshot := fixtureSnapshot(1, 1, api.Starting).Data()
	execution := must(api.NewArgvExecution(api.ArgvExecutionData{Executable: "fixture", Arguments: []string{"literal", "\x1b", ""}}))
	policy := must(api.NewEnvironmentPolicy(api.EnvironmentPolicyData{InheritBase: true}))
	display := snapshot.Display.Data()
	invocation := must(api.NewInvocation(api.InvocationData{Execution: execution, Environment: policy, Cwd: display.Cwd, Terminal: display.Terminal, Geometry: display.Geometry, Label: display.Label}))
	request := must(api.NewSessionStartRequest(api.SessionStartRequestData{OperationID: snapshot.StartOperation, Invocation: invocation}))
	return r.admit(context.Background(), request, env, snapshot.Display, snapshot.Capabilities, api.None[domain.SessionID]())
}

func fixturePublish(r *registry, s *session, phase api.SessionPhase) error {
	s.mu.Lock()
	d := s.snapshot.Data()
	s.mu.Unlock()
	d.Sequence = sequence(d.Sequence.Value() + 1)
	d.Phase = phase
	state := api.CleanupPending
	kind := api.StateChanged
	if phase == api.Cleaned {
		state = api.CleanupComplete
		kind = api.RuntimeCleaned
	}
	if phase == api.CleanupFailed {
		state = api.CleanupFailedState
	}
	d.Cleanup = must(api.NewSessionCleanup(api.SessionCleanupData{State: state}))
	return r.publish(s, must(api.NewSessionSnapshot(d)), kind)
}

func TestRegistryConcurrentAdmissionIsBoundedAndCopied(t *testing.T) {
	r := newRegistry()
	var wg sync.WaitGroup
	results := make(chan *session, 100)
	failures := make(chan error, 100)
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			env := []string{"TOKEN=owned-private-value"}
			s, err := fixtureAdmit(r, env)
			env[0] = "TOKEN=mutated"
			if err != nil {
				failures <- err
				return
			}
			results <- s
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	seen := map[domain.SessionID]bool{}
	for s := range results {
		d := s.snapshot.Data()
		if seen[d.SessionID] || !d.SessionID.Valid() || s.environment[0] != "TOKEN=owned-private-value" {
			t.Fatal("identity/copy invariant")
		}
		seen[d.SessionID] = true
	}
	if len(seen) != liveCapacity || r.live != liveCapacity || len(r.events.reservations) != liveCapacity {
		t.Fatal("admission not atomic")
	}
	for err := range failures {
		if !errors.Is(err, errBusy) {
			t.Fatal(err)
		}
	}
	filter := must(api.NewSessionFilter(api.SessionFilterData{}))
	list := must(r.list(context.Background(), filter)).Data()
	for i, s := range list.Sessions {
		if s.Data().SessionID.Value() != uint64(i+1) {
			t.Fatal("list not ascending")
		}
	}
	captured := r.closeAdmission()
	if len(captured) != liveCapacity {
		t.Fatal("shutdown missed Starting sessions")
	}
	if _, err := fixtureAdmit(r, nil); !errors.Is(err, errClosed) {
		t.Fatal("shutdown admission race")
	}
}

func TestRegistryNeverReusesOrWrapsIDs(t *testing.T) {
	r := newRegistry()
	r.nextID = math.MaxUint64 - 1
	s := must(fixtureAdmit(r, nil))
	if s.snapshot.Data().SessionID.Value() != math.MaxUint64 {
		t.Fatal("last ID not allocated")
	}
	if err := fixturePublish(r, s, api.Cleaned); err != nil {
		t.Fatal(err)
	}
	if _, err := fixtureAdmit(r, nil); !errors.Is(err, errExhausted) || r.nextID != math.MaxUint64 {
		t.Fatal("session ID wrapped/reused")
	}
	if _, err := r.lookup(sessionID(1)); !errors.Is(err, errNotFound) {
		t.Fatal("foreign ID accepted")
	}
}

func TestHistoryEvictsOnlyCleanedAndFinalsRequireACK(t *testing.T) {
	r := newRegistry()
	retained := must(fixtureAdmit(r, nil))
	if err := fixturePublish(r, retained, api.CleanupFailed); err != nil {
		t.Fatal(err)
	}
	after := cursor(0)
	for i := 0; i < 2*historyCapacity; i++ {
		s := must(fixtureAdmit(r, nil))
		if err := fixturePublish(r, s, api.Cleaned); err != nil {
			t.Fatal(err)
		}
		if err := fixturePublish(r, s, api.Cleaned); !errors.Is(err, errInvalid) {
			t.Fatal("duplicate final accepted")
		}
		for {
			event := must(r.events.next(context.Background(), after))
			after = event.Data().Sequence
			if event.Data().Kind == api.RuntimeCleaned {
				break
			}
		}
		if err := r.events.ack(after); err != nil {
			t.Fatal(err)
		}
	}
	if r.live != 1 || len(r.history) != historyCapacity || len(r.sessions) != historyCapacity+1 {
		t.Fatal("history/live limits")
	}
	if s, err := r.lookup(retained.snapshot.Data().SessionID); err != nil || s != retained {
		t.Fatal("cleanup-pending resource evicted")
	}
	if _, err := r.lookup(sessionID(2)); !errors.Is(err, errNotFound) {
		t.Fatal("old cleaned snapshot retained")
	}
	if err := r.events.closeProducers(); !errors.Is(err, errBusy) {
		t.Fatal("cleanup-failed reservation lost")
	}
	if err := fixturePublish(r, retained, api.Cleaned); err != nil {
		t.Fatal(err)
	}
	if len(r.events.reservations) != 1 {
		t.Fatal("cleanup delivery released reservation")
	}
	final := must(r.events.next(context.Background(), after))
	if final.Data().SessionID != retained.snapshot.Data().SessionID {
		t.Fatal("retained resource final lost")
	}
}

func TestRegistryConcurrentShutdownNeverMissesAdmittedSession(t *testing.T) {
	for range 20 {
		r := newRegistry()
		start := make(chan struct{})
		var wg sync.WaitGroup
		for range 30 {
			wg.Add(1)
			go func() { defer wg.Done(); <-start; _, _ = fixtureAdmit(r, nil) }()
		}
		close(start)
		captured := r.closeAdmission()
		wg.Wait()
		if len(captured) != len(r.sessions) {
			t.Fatal("session admitted after shutdown captured set")
		}
	}
}
