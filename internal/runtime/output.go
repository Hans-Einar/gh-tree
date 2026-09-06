package runtime

import (
	"math"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

const outputCapacity = 256 << 10

type outputSpan struct {
	start, end uint64
	stream     api.OutputStream
	sequence   api.SessionSequence
}

// outputRing is protected by its session's mutex. Only retained bytes are kept;
// metadata has at most one span per byte and discarded spans are zeroed.
type outputRing struct {
	bytes      [outputCapacity]byte
	start, end uint64
	spans      []outputSpan
	head       int
}

func (r *outputRing) append(stream api.OutputStream, seq api.SessionSequence, data []byte) error {
	if !stream.Valid() || !seq.Valid() || len(data) == 0 {
		return errInvalid
	}
	if uint64(len(data)) > math.MaxUint64-r.end {
		return errExhausted
	}
	end := r.end + uint64(len(data))
	start := r.start
	if end-start > outputCapacity {
		start = end - outputCapacity
	}
	writeStart := r.end
	if writeStart < start {
		data = data[start-writeStart:]
		writeStart = start
	}
	index := int(writeStart % outputCapacity)
	n := copy(r.bytes[index:], data)
	copy(r.bytes[:], data[n:])
	for r.head < len(r.spans) && r.spans[r.head].end <= start {
		r.spans[r.head] = outputSpan{}
		r.head++
	}
	if r.head < len(r.spans) && r.spans[r.head].start < start {
		r.spans[r.head].start = start
	}
	// Compact geometrically; single-byte producers cannot grow discarded metadata.
	if r.head >= len(r.spans)-r.head {
		r.spans = append(r.spans[:0], r.spans[r.head:]...)
		r.head = 0
	}
	span := outputSpan{writeStart, end, stream, seq}
	if len(r.spans) > r.head && r.spans[len(r.spans)-1].stream == stream && r.spans[len(r.spans)-1].sequence == seq {
		r.spans[len(r.spans)-1].end = end
	} else {
		r.spans = append(r.spans, span)
	}
	r.start, r.end = start, end
	return nil
}

func (r *outputRing) rangeValue() api.OutputRange {
	v, err := api.NewOutputRange(api.OutputRangeData{RetainedStart: r.start, End: r.end, Truncated: r.start != 0})
	if err != nil {
		panic(err)
	}
	return v
}

func (r *outputRing) read(id domain.SessionID, seq api.SessionSequence, offset uint64, limit uint32) (api.SessionOutputResult, error) {
	if !id.Valid() || !seq.Valid() || limit < 1 || limit > outputCapacity || offset > r.end {
		return api.SessionOutputResult{}, errInvalid
	}
	d := api.SessionOutputResultData{SessionID: id, Sequence: seq, RetainedStart: r.start, End: r.end, NextOffset: offset, Truncated: r.start != 0}
	if offset < r.start {
		gap, err := api.NewOutputGap(api.OutputGapData{From: offset, To: r.start})
		if err != nil {
			return api.SessionOutputResult{}, err
		}
		d.Gap, d.NextOffset = api.Some(gap), r.start
	}
	remaining := uint64(limit)
	for _, s := range r.spans[r.head:] {
		if s.end <= d.NextOffset {
			continue
		}
		if remaining == 0 {
			break
		}
		length := min(remaining, s.end-d.NextOffset)
		data := make([]byte, int(length))
		index := int(d.NextOffset % outputCapacity)
		n := copy(data, r.bytes[index:])
		copy(data[n:], r.bytes[:])
		chunk, err := api.NewSessionOutputChunk(api.SessionOutputChunkData{Stream: s.stream, Offset: d.NextOffset, Bytes: data, Sequence: s.sequence})
		if err != nil {
			return api.SessionOutputResult{}, err
		}
		d.Chunks = append(d.Chunks, chunk)
		d.NextOffset += length
		remaining -= length
	}
	return api.NewSessionOutputResult(d)
}
