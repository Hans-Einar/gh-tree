package runtime

import (
	"context"
	"io"
	"sync"
)

const inputCapacity = 64 << 10

// inputQueue owns accepted buffers until a single native writer calls finish.
// Popping does not release capacity: in-flight bytes count against admission.
// It starts no goroutine and never closes a native resource under its mutex.
type inputQueue struct {
	mu          sync.Mutex
	changed     chan struct{}
	queue       [][]byte
	outstanding int
	inflight    int
	closed      bool
}

func newInputQueue() *inputQueue { return &inputQueue{changed: make(chan struct{})} }

func (q *inputQueue) wakeLocked() { close(q.changed); q.changed = make(chan struct{}) }

func (q *inputQueue) accept(ctx context.Context, data []byte) error {
	if len(data) == 0 || len(data) > inputCapacity {
		return errInvalid
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if q.closed {
		return errClosed
	}
	if len(data) > inputCapacity-q.outstanding {
		return errBusy
	}
	q.queue = append(q.queue, append([]byte(nil), data...))
	q.outstanding += len(data)
	q.wakeLocked()
	return nil
}

func (q *inputQueue) next(ctx context.Context) ([]byte, error) {
	for {
		q.mu.Lock()
		if err := ctx.Err(); err != nil {
			q.mu.Unlock()
			return nil, err
		}
		if q.inflight != 0 {
			q.mu.Unlock()
			return nil, errBusy
		}
		if len(q.queue) > 0 {
			data := q.queue[0]
			q.queue[0] = nil
			q.queue = q.queue[1:]
			q.inflight = len(data)
			q.mu.Unlock()
			return data, nil
		}
		if q.closed {
			q.mu.Unlock()
			return nil, io.EOF
		}
		changed := q.changed
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-changed:
		}
	}
}

// finish is called after the writer has recorded accepted versus delivered
// bytes, including a partial write. It never authorizes a replay of any bytes.
func (q *inputQueue) finish() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.inflight == 0 {
		return errInvalid
	}
	q.outstanding -= q.inflight
	q.inflight = 0
	q.wakeLocked()
	return nil
}

// close rejects new input, returns the number of queued bytes whose delivery
// was canceled, and retains in-flight ownership until the native writer joins.
func (q *inputQueue) close() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	discarded := q.outstanding - q.inflight
	q.outstanding = q.inflight
	clear(q.queue)
	q.queue = nil
	q.wakeLocked()
	return discarded
}
