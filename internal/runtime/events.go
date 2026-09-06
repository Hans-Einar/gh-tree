package runtime

import (
	"context"
	"io"
	"math"
	"sync"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

const finalCapacity = 256
const hintCapacity = 64

// eventBuffer owns source-event delivery, not session lifetime. A reservation
// remains from admission until cumulative ACK of its final event, independently
// of snapshot eviction. Native producers only take this short memory lock.
type eventBuffer struct {
	mu           sync.Mutex
	changed      chan struct{}
	sequence     uint64
	acked        uint64
	reservations map[domain.SessionID]bool // false: a final still must be produced
	finals       map[uint64]api.RuntimeEvent
	hints        map[domain.SessionID]api.RuntimeEvent
	delivered    map[uint64]bool // true: hint receipt; false: reliable final receipt
	hintReceipts int
	closed       bool // only set after all session producers have joined
}

func newEventBuffer() *eventBuffer {
	return &eventBuffer{changed: make(chan struct{}), reservations: make(map[domain.SessionID]bool), finals: make(map[uint64]api.RuntimeEvent), hints: make(map[domain.SessionID]api.RuntimeEvent), delivered: make(map[uint64]bool)}
}

func (q *eventBuffer) wakeLocked() { close(q.changed); q.changed = make(chan struct{}) }

func (q *eventBuffer) pendingFinalsLocked() uint64 {
	var n uint64
	for _, produced := range q.reservations {
		if !produced {
			n++
		}
	}
	return n
}

func (q *eventBuffer) reserve(id domain.SessionID) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !id.Valid() {
		return errInvalid
	}
	if q.closed {
		return errClosed
	}
	if _, exists := q.reservations[id]; exists {
		return errInvalid
	}
	if len(q.reservations) >= finalCapacity {
		return errBusy
	}
	// Reserve one numerical sequence as well as one memory slot for each final.
	if math.MaxUint64-q.sequence <= q.pendingFinalsLocked() {
		return errExhausted
	}
	q.reservations[id] = false
	return nil
}

func (q *eventBuffer) publish(snapshot api.SessionSnapshot, kind api.RuntimeEventKind) error {
	if !snapshot.Valid() || !kind.Valid() {
		return errInvalid
	}
	s := snapshot.Data()
	if (kind == api.RuntimeCleaned) != (s.Phase == api.Cleaned) {
		return errInvalid
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	produced, reserved := q.reservations[s.SessionID]
	if !reserved || produced || q.closed {
		return errInvalid
	}
	if kind != api.RuntimeCleaned {
		if math.MaxUint64-q.sequence <= q.pendingFinalsLocked() {
			return nil
		} // coalesce away hints, preserve every final
		if _, exists := q.hints[s.SessionID]; !exists && len(q.hints) >= hintCapacity {
			return nil
		}
	}
	if q.sequence == math.MaxUint64 {
		return errExhausted
	}
	seq, _ := api.NewRuntimeEventSequence(q.sequence + 1)
	event, err := api.NewRuntimeEvent(api.RuntimeEventData{Sequence: seq, SessionSequence: s.Sequence, SessionID: s.SessionID, Kind: kind, Snapshot: snapshot})
	if err != nil {
		return err
	}
	q.sequence++
	if kind == api.RuntimeCleaned {
		q.reservations[s.SessionID] = true
		q.finals[q.sequence] = event
		delete(q.hints, s.SessionID)
	} else {
		q.hints[s.SessionID] = event
	}
	q.wakeLocked()
	return nil
}

func (q *eventBuffer) cursorValidLocked(cursor uint64) bool {
	if cursor < q.acked {
		return false
	}
	if cursor == q.acked {
		return true
	}
	_, delivered := q.delivered[cursor]
	return delivered
}

func (q *eventBuffer) next(ctx context.Context, after api.RuntimeEventSequence) (api.RuntimeEvent, error) {
	for {
		q.mu.Lock()
		if err := ctx.Err(); err != nil {
			q.mu.Unlock()
			return api.RuntimeEvent{}, err
		}
		cursor := after.Value()
		if !q.cursorValidLocked(cursor) {
			q.mu.Unlock()
			return api.RuntimeEvent{}, errCursor
		}
		var selected api.RuntimeEvent
		best := uint64(0)
		for seq, event := range q.finals {
			if seq > cursor && (best == 0 || seq < best) {
				best, selected = seq, event
			}
		}
		for _, event := range q.hints {
			seq := event.Data().Sequence.Value()
			_, known := q.delivered[seq]
			// Without ACK, infinitely many delivered hint cursors would require
			// unbounded replay validation memory. Suppress new hint deliveries
			// at the receipt bound; reliable finals have independent slots.
			if !known && q.hintReceipts >= hintCapacity {
				continue
			}
			if seq > cursor && (best == 0 || seq < best) {
				best, selected = seq, event
			}
		}
		if best != 0 {
			if _, known := q.delivered[best]; !known {
				hint := selected.Data().Kind != api.RuntimeCleaned
				q.delivered[best] = hint
				if hint {
					q.hintReceipts++
				}
			}
			q.mu.Unlock()
			return selected, nil
		}
		if q.closed && len(q.reservations) == 0 {
			q.mu.Unlock()
			return api.RuntimeEvent{}, io.EOF
		}
		changed := q.changed
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return api.RuntimeEvent{}, ctx.Err()
		case <-changed:
		}
	}
}

func (q *eventBuffer) ack(cursor api.RuntimeEventSequence) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := cursor.Value()
	if !q.cursorValidLocked(n) {
		return errCursor
	}
	if n == q.acked {
		return nil
	}
	for seq := range q.finals {
		if seq <= n {
			if _, delivered := q.delivered[seq]; !delivered {
				return errCursor
			}
		}
	}
	for seq, event := range q.finals {
		if seq <= n {
			delete(q.reservations, event.Data().SessionID)
			delete(q.finals, seq)
		}
	}
	for id, event := range q.hints {
		if event.Data().Sequence.Value() <= n {
			delete(q.hints, id)
		}
	}
	for seq, hint := range q.delivered {
		if seq <= n {
			if hint {
				q.hintReceipts--
			}
			delete(q.delivered, seq)
		}
	}
	q.acked = n
	q.wakeLocked()
	return nil
}

func (q *eventBuffer) closeProducers() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.pendingFinalsLocked() != 0 {
		return errBusy
	}
	q.closed = true
	q.wakeLocked()
	return nil
}

func (q *eventBuffer) position() api.RuntimeEventSequence {
	q.mu.Lock()
	defer q.mu.Unlock()
	seq, _ := api.NewRuntimeEventSequence(q.sequence)
	return seq
}
