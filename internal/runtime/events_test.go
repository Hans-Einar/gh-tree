package runtime

import (
	"context"
	"errors"
	"io"
	"math"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestReliableEventsReplayACKAndMisuse(t *testing.T) {
	q := newEventBuffer()
	for i := uint64(1); i <= 3; i++ {
		if err := q.reserve(sessionID(i)); err != nil {
			t.Fatal(err)
		}
		if err := q.publish(fixtureSnapshot(i, 1, api.Starting), api.StateChanged); err != nil {
			t.Fatal(err)
		}
		if err := q.publish(fixtureSnapshot(i, 2, api.Cleaned), api.RuntimeCleaned); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.ack(cursor(2)); !errors.Is(err, errCursor) {
		t.Fatal("undelivered ACK accepted")
	}
	if _, err := q.next(context.Background(), cursor(1)); !errors.Is(err, errCursor) {
		t.Fatal("coalesced but undelivered cursor accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := q.next(ctx, cursor(0)); !errors.Is(err, context.Canceled) || len(q.delivered) != 0 {
		t.Fatal("canceled call took delivery")
	}
	first := must(q.next(context.Background(), cursor(0)))
	replay := must(q.next(context.Background(), cursor(0)))
	if first.Data().Sequence != replay.Data().Sequence || len(q.delivered) != 1 {
		t.Fatal("replay differs")
	}
	second := must(q.next(context.Background(), first.Data().Sequence))
	if err := q.ack(second.Data().Sequence); err != nil {
		t.Fatal(err)
	}
	if len(q.reservations) != 1 {
		t.Fatal("cumulative ACK reservation count")
	}
	if err := q.ack(second.Data().Sequence); err != nil {
		t.Fatal("idempotent ACK", err)
	}
	if err := q.ack(first.Data().Sequence); !errors.Is(err, errCursor) {
		t.Fatal("regressing ACK accepted")
	}
	if _, err := q.next(context.Background(), first.Data().Sequence); !errors.Is(err, errCursor) {
		t.Fatal("regressing read cursor accepted")
	}
	if err := q.closeProducers(); err != nil {
		t.Fatal(err)
	}
	last := must(q.next(context.Background(), second.Data().Sequence))
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := q.next(ctx, last.Data().Sequence); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("EOF before final ACK")
	}
	if err := q.ack(last.Data().Sequence); err != nil {
		t.Fatal(err)
	}
	if _, err := q.next(context.Background(), last.Data().Sequence); !errors.Is(err, io.EOF) {
		t.Fatal("missing EOF")
	}
}

func TestFinalCapacityDoesNotDependOnConsumer(t *testing.T) {
	q := newEventBuffer()
	for i := uint64(1); i <= finalCapacity; i++ {
		if err := q.reserve(sessionID(i)); err != nil {
			t.Fatal(err)
		}
		if err := q.publish(fixtureSnapshot(i, 1, api.Cleaned), api.RuntimeCleaned); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.reserve(sessionID(finalCapacity + 1)); !errors.Is(err, errBusy) {
		t.Fatal("final reservation ceiling bypassed")
	}
	after := cursor(0)
	for i := 1; i <= finalCapacity; i++ {
		event := must(q.next(context.Background(), after))
		if event.Data().SessionID.Value() != uint64(i) {
			t.Fatal("final order")
		}
		after = event.Data().Sequence
	}
	if len(q.reservations) != finalCapacity {
		t.Fatal("delivery prematurely released reservation")
	}
	if err := q.ack(after); err != nil {
		t.Fatal(err)
	}
	if err := q.reserve(sessionID(finalCapacity + 1)); err != nil {
		t.Fatal("ACK did not restore capacity", err)
	}
	if err := q.closeProducers(); !errors.Is(err, errBusy) {
		t.Fatal("unjoined final producer closed")
	}
}

func TestHintReceiptBoundCannotStarveReliableFinal(t *testing.T) {
	q := newEventBuffer()
	if err := q.reserve(sessionID(1)); err != nil {
		t.Fatal(err)
	}
	after := cursor(0)
	for i := uint64(1); i <= hintCapacity; i++ {
		if err := q.publish(fixtureSnapshot(1, i, api.Running), api.OutputAvailable); err != nil {
			t.Fatal(err)
		}
		after = must(q.next(context.Background(), after)).Data().Sequence
	}
	for i := uint64(hintCapacity + 1); i <= 500; i++ {
		if err := q.publish(fixtureSnapshot(1, i, api.Running), api.OutputAvailable); err != nil {
			t.Fatal(err)
		}
	}
	if len(q.delivered) != hintCapacity || len(q.hints) != 1 {
		t.Fatal("hint receipt memory grew")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := q.next(ctx, after); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("hint receipt ceiling ignored")
	}
	if err := q.publish(fixtureSnapshot(1, 501, api.Cleaned), api.RuntimeCleaned); err != nil {
		t.Fatal(err)
	}
	final := must(q.next(context.Background(), after))
	if final.Data().Kind != api.RuntimeCleaned || len(q.delivered) != hintCapacity+1 {
		t.Fatal("reliable final starved")
	}
	if err := q.ack(final.Data().Sequence); err != nil || len(q.delivered) != 0 || q.hintReceipts != 0 {
		t.Fatal("receipt ACK accounting")
	}
}

func TestSequenceSpaceReservesReliableFinals(t *testing.T) {
	q := newEventBuffer()
	q.sequence = math.MaxUint64 - 2
	if err := q.reserve(sessionID(1)); err != nil {
		t.Fatal(err)
	}
	if err := q.reserve(sessionID(2)); err != nil {
		t.Fatal(err)
	}
	if err := q.reserve(sessionID(3)); !errors.Is(err, errExhausted) {
		t.Fatal("final numerical reservation absent")
	}
	if err := q.publish(fixtureSnapshot(1, 1, api.Running), api.StateChanged); err != nil || q.sequence != math.MaxUint64-2 {
		t.Fatal("hint consumed final sequence")
	}
	for i := uint64(1); i <= 2; i++ {
		if err := q.publish(fixtureSnapshot(i, 2, api.Cleaned), api.RuntimeCleaned); err != nil {
			t.Fatal(err)
		}
	}
	if q.sequence != math.MaxUint64 || len(q.finals) != 2 {
		t.Fatal("lost final at exhaustion")
	}
}
