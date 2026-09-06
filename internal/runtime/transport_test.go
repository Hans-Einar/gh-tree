package runtime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
func sequence(n uint64) api.SessionSequence    { return must(api.NewSessionSequence(n)) }
func cursor(n uint64) api.RuntimeEventSequence { return must(api.NewRuntimeEventSequence(n)) }
func sessionID(n uint64) domain.SessionID      { return must(domain.NewSessionID(n)) }

func fixtureSnapshot(id, seq uint64, phase api.SessionPhase) api.SessionSnapshot {
	repo := must(domain.NewRepositoryID(domain.LocalCommon, "runtime-fixture"))
	worktree := must(domain.NewWorktreeID(repo, "fixture"))
	identity := must(api.NewDirectoryIdentity(api.DirectoryUnix, 1, [16]byte{1}, "change:1"))
	source := must(api.NewSourceVersion("fixture", "runtime", "unit", "1"))
	scope := must(api.NewWorktreeScope(api.WorktreeScopeData{ID: worktree, RootLocator: "/runtime-fixture", RootIdentity: identity, Source: source}))
	cwd := must(api.NewCwdObservation(api.CwdObservationData{Worktree: scope, ProjectIdentity: identity, Source: source}))
	geometry := must(api.NewGeometry(api.GeometryData{Rows: 24, Columns: 80}))
	display := must(api.NewInvocationSummary(api.InvocationSummaryData{Label: "fixture", ExecutableDisplay: "fixture", Cwd: cwd, Terminal: api.Pipes, Geometry: geometry}))
	cleanupState := api.CleanupPending
	if phase == api.Cleaned {
		cleanupState = api.CleanupComplete
	}
	if phase == api.CleanupFailed {
		cleanupState = api.CleanupFailedState
	}
	cleanup := must(api.NewSessionCleanup(api.SessionCleanupData{State: cleanupState}))
	capabilities := must(api.NewSessionCapabilities(api.SessionCapabilitiesData{Output: true, TreeStop: true, Restart: true}))
	return must(api.NewSessionSnapshot(api.SessionSnapshotData{SessionID: sessionID(id), WorktreeID: worktree, StartOperation: must(api.NewOperationID(id)), Display: display, Capabilities: capabilities, Phase: phase, Cleanup: cleanup, Sequence: sequence(seq), OutputRange: must(api.NewOutputRange(api.OutputRangeData{}))}))
}

func TestOutputRandomizedByteReference(t *testing.T) {
	random := rand.New(rand.NewPCG(98137, 58319))
	var ring outputRing
	var all []byte
	var expectedStreams []api.OutputStream
	var expectedSeq []uint64
	for step := 1; step <= 120; step++ {
		length := 1 + random.IntN(13000)
		data := make([]byte, length)
		for i := range data {
			data[i] = byte(random.Uint32())
		}
		stream := api.OutputStream(1 + random.IntN(3))
		all = append(all, data...)
		for range data {
			expectedStreams = append(expectedStreams, stream)
			expectedSeq = append(expectedSeq, uint64(step))
		}
		if err := ring.append(stream, sequence(uint64(step)), data); err != nil {
			t.Fatal(err)
		}
		clear(data) // producer mutation must not alter retained bytes
		for range 5 {
			offset := uint64(random.IntN(len(all) + 1))
			limit := uint32(1 + random.IntN(outputCapacity))
			got, err := ring.read(sessionID(1), sequence(uint64(step)), offset, limit)
			if err != nil {
				t.Fatal(err)
			}
			d := got.Data()
			start := uint64(max(0, len(all)-outputCapacity))
			next := max(offset, start)
			end := min(uint64(len(all)), next+uint64(limit))
			if d.RetainedStart != start || d.End != uint64(len(all)) || d.NextOffset != end || d.Truncated != (start > 0) {
				t.Fatalf("wrong range: %+v", d)
			}
			gap, hasGap := d.Gap.Value()
			if hasGap != (offset < start) || (hasGap && (gap.Data().From != offset || gap.Data().To != start)) {
				t.Fatal("wrong gap")
			}
			var received []byte
			for _, chunk := range d.Chunks {
				c := chunk.Data()
				if c.Offset != next {
					t.Fatal("noncontiguous chunks")
				}
				for range c.Bytes {
					if expectedStreams[next] != c.Stream || expectedSeq[next] != c.Sequence.Value() {
						t.Fatal("lost stream/sequence identity")
					}
					next++
				}
				received = append(received, c.Bytes...)
				clear(c.Bytes) // consumer mutation also must not alter owned output
			}
			if !bytes.Equal(received, all[max(offset, start):end]) {
				t.Fatal("raw bytes differ")
			}
		}
	}
}

func TestOutputLargeWriteAndOverflow(t *testing.T) {
	var ring outputRing
	data := bytes.Repeat([]byte{0, 27, 255, 13, 8, 10}, outputCapacity)
	if err := ring.append(api.Stderr, sequence(2), data); err != nil {
		t.Fatal(err)
	}
	got := must(ring.read(sessionID(1), sequence(2), 0, outputCapacity)).Data()
	if len(got.Chunks) != 1 || !bytes.Equal(got.Chunks[0].Data().Bytes, data[len(data)-outputCapacity:]) {
		t.Fatal("oversized append did not preserve tail")
	}
	if _, err := ring.read(sessionID(1), sequence(2), ring.end+1, 1); !errors.Is(err, errInvalid) {
		t.Fatal("future offset accepted")
	}
	if _, err := ring.read(sessionID(1), sequence(2), ring.end, 0); !errors.Is(err, errInvalid) {
		t.Fatal("zero bound accepted")
	}
	empty := must(ring.read(sessionID(1), sequence(2), ring.end, 1)).Data()
	if len(empty.Chunks) != 0 || empty.NextOffset != ring.end {
		t.Fatal("end offset")
	}
	ring.end = math.MaxUint64
	if err := ring.append(api.Stdout, sequence(3), []byte{1}); !errors.Is(err, errExhausted) || ring.end != math.MaxUint64 {
		t.Fatal("offset wrapped")
	}
}

func TestOutputSingleByteMetadataBound(t *testing.T) {
	var ring outputRing
	for i := 1; i <= 3*outputCapacity; i++ {
		if err := ring.append(api.Stdout, sequence(uint64(i)), []byte{byte(i)}); err != nil {
			t.Fatal(err)
		}
	}
	if len(ring.spans)-ring.head != outputCapacity || cap(ring.spans) > 4*outputCapacity {
		t.Fatalf("metadata not bounded: len=%d head=%d cap=%d", len(ring.spans), ring.head, cap(ring.spans))
	}
}

func TestInputOwnershipIncludesInflightAndClose(t *testing.T) {
	q := newInputQueue()
	input := bytes.Repeat([]byte{7}, inputCapacity)
	if err := q.accept(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	clear(input)
	inflight := must(q.next(context.Background()))
	if inflight[0] != 7 {
		t.Fatal("accepted buffer was not copied")
	}
	if err := q.accept(context.Background(), []byte{1}); !errors.Is(err, errBusy) {
		t.Fatal("popping released in-flight capacity")
	}
	if n := q.close(); n != 0 || q.outstanding != inputCapacity {
		t.Fatal("close lost in-flight ownership")
	}
	if err := q.finish(); err != nil || q.outstanding != 0 {
		t.Fatal("writer ownership not released")
	}
	if _, err := q.next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatal("closed queue not EOF")
	}
	if err := q.accept(context.Background(), []byte{1}); !errors.Is(err, errClosed) {
		t.Fatal("late input accepted")
	}
}

func TestInputCancelAndConcurrentWholeRequests(t *testing.T) {
	q := newInputQueue()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := q.accept(ctx, []byte{1}); !errors.Is(err, context.Canceled) || q.outstanding != 0 {
		t.Fatal("canceled input admitted")
	}
	var wg sync.WaitGroup
	accepted := make(chan bool, 100)
	for range 100 {
		wg.Add(1)
		go func() { defer wg.Done(); accepted <- q.accept(context.Background(), make([]byte, 1024)) == nil }()
	}
	wg.Wait()
	close(accepted)
	n := 0
	for ok := range accepted {
		if ok {
			n++
		}
	}
	if n != 64 || q.outstanding != inputCapacity {
		t.Fatalf("whole-request bound: %d %d", n, q.outstanding)
	}
	if dropped := q.close(); dropped != inputCapacity || q.outstanding != 0 {
		t.Fatal("queued close accounting")
	}
	deadline, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	empty := newInputQueue()
	if _, err := empty.next(deadline); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("blocked next not canceled")
	}
}
