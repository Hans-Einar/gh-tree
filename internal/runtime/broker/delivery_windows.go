package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

var ErrWindowsControlsBusy = errors.New("native control capacity exhausted")

// windowsControlMutex supports cancellable admission without a waiter goroutine.
// Lifecycle sends and teardown use Lock; ordinary controls use lockContext.
type windowsControlMutex struct {
	once sync.Once
	gate chan struct{}
}

func (m *windowsControlMutex) init()   { m.once.Do(func() { m.gate = make(chan struct{}, 1) }) }
func (m *windowsControlMutex) Lock()   { m.init(); m.gate <- struct{}{} }
func (m *windowsControlMutex) Unlock() { <-m.gate }
func (m *windowsControlMutex) lockContext(ctx context.Context, stop, done <-chan struct{}) error {
	m.init()
	select {
	case m.gate <- struct{}{}:
		return nil // admission rechecks cancellation and Stop under pendingMu
	case <-ctx.Done():
		return ctx.Err()
	case <-stop:
		return io.ErrClosedPipe
	case <-done:
		return io.ErrClosedPipe
	}
}

// WindowsReceipt is a bounded Runtime-private observation capability, not a
// public DTO. A nonnil receipt always means dispatch was possible. Wait joins
// the original operation, never sends again. At most 64 unobserved operations
// are retained by a client, including completed replies. Native cleanup proceeds
// without their consumption. Results contain no copied input or native handles.
type WindowsReceipt struct {
	owner    *WindowsClient
	sequence uint64
	op       Opcode
	expected uint32
	done     chan struct{}
	result   controlReply // written once before close(done), immutable afterward
	retired  sync.Once
}

// Wait prefers known terminal facts over simultaneous cancellation. Completed
// distinguishes known native completion from a terminal transport uncertainty;
// either terminal observation retires admission once, even with a nonnil error.
// Repeated/concurrent observations return the same facts after retirement.
func (r *WindowsReceipt) Wait(ctx context.Context) (WindowsDelivery, error) {
	if r == nil || ctx == nil {
		return WindowsDelivery{}, ErrProtocol
	}
	select {
	case <-r.done:
		return r.observe()
	default:
	}
	select {
	case <-r.done:
		return r.observe()
	case <-ctx.Done():
		select {
		case <-r.done:
			return r.observe()
		default:
			return WindowsDelivery{Dispatched: true, Receipt: r}, ctx.Err()
		}
	}
}

func (r *WindowsReceipt) observe() (WindowsDelivery, error) {
	r.retired.Do(func() {
		r.owner.pendingMu.Lock()
		delete(r.owner.pending, r.sequence)
		r.owner.pendingMu.Unlock()
		r.owner = nil
	})
	result := r.result.delivery
	result.Receipt = r
	return result, r.result.err
}

// Called only under pendingMu, including cleanup retirement. A known reply is
// immutable and cannot be overwritten by later EOF, Stop or cleanup failures.
func (c *WindowsClient) completeReceiptLocked(r *WindowsReceipt, result controlReply) {
	select {
	case <-r.done:
		return
	default:
	}
	r.result = result
	if r.op == WriteInput || r.op == Interrupt {
		c.inputPending = false
	}
	close(r.done)
}

func (c *WindowsClient) receiveDelivery(payload []byte) error {
	if len(payload) != 8 && len(payload) != 17 {
		return ErrProtocol
	}
	sequence := binary.BigEndian.Uint64(payload)
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	r := c.pending[sequence]
	if r == nil {
		// A Stop may send one courtesy ETX. It has no public control caller.
		if sequence == 0 || sequence != c.stopSequence || len(payload) != 17 {
			return ErrProtocol
		}
		c.stopSequence = 0
	} else {
		select {
		case <-r.done:
			return ErrProtocol // duplicate delivery cannot overwrite evidence
		default:
		}
		if (r.op == Resize) != (len(payload) == 8) {
			return ErrProtocol
		}
	}
	result := controlReply{delivery: WindowsDelivery{Completed: true, Dispatched: true}}
	if len(payload) == 17 {
		accepted := binary.BigEndian.Uint32(payload[8:])
		delivered := binary.BigEndian.Uint32(payload[12:])
		expected := uint32(1)
		if r != nil {
			expected = r.expected
		}
		if accepted != expected || delivered > accepted || payload[16] > 1 || (payload[16] == 0 && delivered != accepted) {
			return ErrProtocol
		}
		result.delivery.Accepted, result.delivery.Delivered = accepted, delivered
		if payload[16] == 1 {
			result.err = io.ErrShortWrite
		}
	}
	if r != nil {
		c.completeReceiptLocked(r, result)
	}
	return nil
}
