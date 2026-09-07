package runtime

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

const liveCapacity = 64
const historyCapacity = 256

type session struct {
	mu           sync.Mutex
	start        api.SessionStartRequest
	environment  []string
	snapshot     api.SessionSnapshot
	output       outputRing
	input        *inputQueue
	owner        nativeOwner
	changed      chan struct{}
	startDone    chan struct{}
	startPending bool
	established  bool
	startErr     error
	stopAsked    bool
	stopSent     bool
	nativeClean  bool
	producers    int
	controlBusy  bool
	observing    bool
	acquired     api.Optional[api.AcquiredCwd]
	exit         api.Optional[api.SessionExit]
	diagnostics  map[api.RuntimeCleanupStage]api.Diagnostic
	restart      *restartTransition
}

// registry synchronizes membership, allocation, final reservation and admission.
// The lock order is registry -> session -> events; no path reverses that order.
// OS calls/waits and native callbacks must occur after releasing these locks.
type registry struct {
	mu          sync.Mutex
	nextID      uint64
	closed      bool
	live        int
	transitions int // admitted restart producers, independent of native sessions
	sessions    map[domain.SessionID]*session
	history     []domain.SessionID // order of completed cleanup, oldest first
	events      *eventBuffer
}

func newRegistry() *registry {
	return &registry{sessions: make(map[domain.SessionID]*session), events: newEventBuffer()}
}

// admit receives an already resolved private environment and safe summary. It
// performs no native acquisition. A successful return creates durable ownership
// of the identity and its final-event reservation, including failed starts.
func (r *registry) admit(ctx context.Context, request api.SessionStartRequest, env []string, display api.InvocationSummary, capabilities api.SessionCapabilities, restartOf api.Optional[domain.SessionID]) (*session, error) {
	if !request.Valid() || !display.Valid() || !capabilities.Valid() {
		return nil, errInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.closed {
		return nil, errClosed
	}
	if r.live >= liveCapacity {
		return nil, errBusy
	}
	if r.nextID == math.MaxUint64 {
		return nil, errExhausted
	}
	id, _ := domain.NewSessionID(r.nextID + 1)
	seq, _ := api.NewSessionSequence(1)
	cleanup, _ := api.NewSessionCleanup(api.SessionCleanupData{State: api.CleanupPending})
	output, _ := api.NewOutputRange(api.OutputRangeData{})
	d := request.Data()
	snapshot, err := api.NewSessionSnapshot(api.SessionSnapshotData{SessionID: id, WorktreeID: d.Invocation.Data().Cwd.Data().Worktree.Data().ID, StartOperation: d.OperationID, RestartOf: restartOf, Display: display, Capabilities: capabilities, Phase: api.Starting, Cleanup: cleanup, Sequence: seq, OutputRange: output})
	if err != nil {
		return nil, err
	}
	if err := r.events.reserve(id); err != nil {
		return nil, err
	}
	s := &session{start: request.Clone(), environment: append([]string(nil), env...), snapshot: snapshot, input: newInputQueue(), changed: make(chan struct{}), startDone: make(chan struct{}), startPending: true, diagnostics: make(map[api.RuntimeCleanupStage]api.Diagnostic)}
	r.nextID++
	r.live++
	r.sessions[id] = s
	if err := r.events.publish(snapshot, api.StateChanged); err != nil {
		panic(err)
	} // validated admission owns a reservation
	return s, nil
}

// change performs a coherent memory-only read/modify/publication. Callers must
// not call native code or wait in edit. Numerical space for the final is kept
// even when an effectively infinite producer exhausts the session sequence.
func (r *registry) change(s *session, kind api.RuntimeEventKind, edit func(*api.SessionSnapshotData) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.snapshot.Data()
	if d.Phase == api.Cleaned {
		return errClosed
	}
	if d.Sequence.Value() == math.MaxUint64 {
		return errExhausted
	}
	if err := edit(&d); err != nil {
		return err
	}
	// Private ownership still accepts cleanup facts after hint sequence space is
	// exhausted. Only the final can consume the last version; output refuses in
	// its edit before touching the ring. Never reuse a published snapshot version.
	if kind != api.RuntimeCleaned && d.Sequence.Value() == math.MaxUint64-1 {
		return errExhausted
	}
	d.Sequence, _ = api.NewSessionSequence(d.Sequence.Value() + 1)
	next, err := api.NewSessionSnapshot(d)
	if err != nil {
		return err
	}
	if err = r.events.publish(next, kind); err != nil {
		return err
	}
	s.snapshot = next
	close(s.changed)
	s.changed = make(chan struct{})
	if d.Phase == api.Cleaned {
		r.live--
		r.history = append(r.history, d.SessionID)
		if len(r.history) > historyCapacity {
			delete(r.sessions, r.history[0])
			r.history = r.history[1:]
		}
		if r.closed && r.live == 0 && r.transitions == 0 {
			_ = r.events.closeProducers()
		}
	}
	return nil
}

func (r *registry) lookup(id domain.SessionID) (*session, error) {
	if !id.Valid() {
		return nil, errInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s, exists := r.sessions[id]
	if !exists {
		return nil, errNotFound
	}
	return s, nil
}

func (r *registry) list(ctx context.Context, filter api.SessionFilter) (api.SessionList, error) {
	if !filter.Valid() {
		return api.SessionList{}, errInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return api.SessionList{}, err
	}
	worktree, filtered := filter.Data().WorktreeID.Value()
	var snapshots []api.SessionSnapshot
	for _, s := range r.sessions {
		s.mu.Lock()
		snapshot := s.snapshot.Clone()
		s.mu.Unlock()
		if !filtered || snapshot.Data().WorktreeID == worktree {
			snapshots = append(snapshots, snapshot)
		}
	}
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Data().SessionID.Value() < snapshots[j].Data().SessionID.Value()
	})
	return api.NewSessionList(api.SessionListData{Sessions: snapshots, Sequence: r.events.position()})
}

// closeAdmission atomically captures every admitted session, including Starting
// and CleanupFailed records. Cleaned records remain available for aggregation.
func (r *registry) closeAdmission() []*session {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	ids := make([]domain.SessionID, 0, len(r.sessions))
	for id := range r.sessions {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].Value() < ids[j].Value() })
	result := make([]*session, 0, len(ids))
	for _, id := range ids {
		result = append(result, r.sessions[id])
	}
	return result
}

// publish records an already observed native fact. Only the future native
// lifecycle owner may construct Cleaned after all of its resource joins. This
// method enforces structural/ordering facts, not the truth of OS observations.
func (r *registry) publish(s *session, snapshot api.SessionSnapshot, kind api.RuntimeEventKind) error {
	if !snapshot.Valid() {
		return errInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	old, next := s.snapshot.Data(), snapshot.Data()
	if r.sessions[old.SessionID] != s || next.SessionID != old.SessionID || next.WorktreeID != old.WorktreeID || next.StartOperation != old.StartOperation || next.RestartOf != old.RestartOf || next.Sequence.Value() <= old.Sequence.Value() || old.Phase == api.Cleaned {
		return errInvalid
	}
	if err := r.events.publish(snapshot, kind); err != nil {
		return err
	}
	s.snapshot = snapshot.Clone()
	if next.Phase == api.Cleaned {
		r.live--
		r.history = append(r.history, next.SessionID)
		if len(r.history) > historyCapacity {
			delete(r.sessions, r.history[0])
			r.history = r.history[1:]
		}
	}
	return nil
}
