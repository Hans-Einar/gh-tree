package broker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func testSpec() StartSpec {
	identity := must(api.NewDirectoryIdentity(api.DirectoryUnix, 1, [16]byte{1}, "birth:1:0"))
	return StartSpec{ParentID: 1, OperationID: 1, RootLocator: "/fixture", Components: []string{" project"}, RootIdentity: identity, ProjectIdentity: identity, Executable: "./literal", Arguments: []string{"", "dev:wan", "x\x1b[31m", "日本語"}, Environment: []string{"NAME=explicit value", "EMPTY="}, Rows: 24, Columns: 80}
}

func TestStartCodecRoundtripCopyAndRefusals(t *testing.T) {
	spec := testSpec()
	buf := must(EncodeStart(spec))
	got := must(DecodeStart(buf))
	if !reflect.DeepEqual(spec, got) {
		t.Fatal("start semantics changed")
	}
	clear(buf)
	if got.Components[0] != " project" || got.Arguments[3] != "日本語" {
		t.Fatal("decoder retained input buffer")
	}
	for _, change := range []func(*StartSpec){
		func(s *StartSpec) { s.ParentID = 0 }, func(s *StartSpec) { s.OperationID = 0 }, func(s *StartSpec) { s.Components = []string{".."} },
		func(s *StartSpec) { s.Components = []string{"/absolute"} }, func(s *StartSpec) { s.Components = []string{"C:drive"} }, func(s *StartSpec) { s.Executable = "zero\x00tail" },
		func(s *StartSpec) { s.Arguments = []string{string(make([]byte, MaxFrame))} }, func(s *StartSpec) { s.Environment = []string{"NO_EQUALS"} }, func(s *StartSpec) { s.Rows = 32768 },
	} {
		spec := testSpec()
		change(&spec)
		if _, err := EncodeStart(spec); !errors.Is(err, ErrProtocol) {
			t.Fatal("invalid StartSpec accepted")
		}
	}
	valid := must(EncodeStart(testSpec()))
	for i := 0; i < len(valid); i++ {
		if _, err := DecodeStart(valid[:i]); !errors.Is(err, ErrProtocol) {
			t.Fatalf("truncation %d accepted", i)
		}
	}
	if _, err := DecodeStart(append(valid, 0)); !errors.Is(err, ErrProtocol) {
		t.Fatal("trailing unknown payload accepted")
	}
	valid[len(valid)-5] = 2
	if _, err := DecodeStart(valid); !errors.Is(err, ErrProtocol) {
		t.Fatal("invalid terminal tag accepted")
	}
}

func TestFramesRejectMalformedLengthsVersionsAndOpcodes(t *testing.T) {
	nonce := Nonce{1}
	frame := Frame{Role: Parent, Opcode: Start, SessionID: 5, Sequence: 1, Nonce: nonce, Payload: must(EncodeStart(testSpec()))}
	encoded := must(EncodeFrame(frame))
	decoded := must(DecodeFrame(bytes.NewReader(encoded)))
	if !reflect.DeepEqual(frame, decoded) {
		t.Fatal("frame differs")
	}
	for i := 0; i < len(encoded); i++ {
		if _, err := DecodeFrame(bytes.NewReader(encoded[:i])); err == nil {
			t.Fatalf("truncated frame %d accepted", i)
		}
	}
	for _, length := range []uint32{0, headerSize - 1, MaxFrame + 1, ^uint32(0)} {
		var prefix [4]byte
		binary.BigEndian.PutUint32(prefix[:], length)
		if _, err := DecodeFrame(bytes.NewReader(prefix[:])); !errors.Is(err, ErrProtocol) {
			t.Fatal("unbounded length accepted")
		}
	}
	for _, at := range []int{4, 6, 7} {
		bad := append([]byte(nil), encoded...)
		bad[at] = 255
		if _, err := DecodeFrame(bytes.NewReader(bad)); !errors.Is(err, ErrProtocol) {
			t.Fatal("unknown version/role/opcode accepted")
		}
	}
	frame.Payload = make([]byte, MaxFrame-headerSize)
	if len(must(EncodeFrame(frame))) != MaxFrame+4 {
		t.Fatal("maximum frame")
	}
	frame.Payload = append(frame.Payload, 0)
	if _, err := EncodeFrame(frame); !errors.Is(err, ErrProtocol) {
		t.Fatal("oversized frame")
	}
}

func TestChannelReplayForeignBindingPoisonsConnection(t *testing.T) {
	nonce := Nonce{1}
	for _, alter := range []func(*Frame){func(f *Frame) { f.Role = UnixSignalHelper }, func(f *Frame) { f.SessionID++ }, func(f *Frame) { f.Nonce[0]++ }, func(f *Frame) { f.Sequence = 2 }} {
		frame := Frame{Role: UnixSupervisor, Opcode: Started, SessionID: 5, Sequence: 1, Nonce: nonce}
		alter(&frame)
		bad := must(EncodeFrame(frame))
		good := must(EncodeFrame(Frame{Role: UnixSupervisor, Opcode: Started, SessionID: 5, Sequence: 1, Nonce: nonce}))
		c := must(NewChannel(bytes.NewReader(append(bad, good...)), io.Discard, Parent, UnixSupervisor, 5, nonce))
		if _, err := c.Receive(); !errors.Is(err, ErrProtocol) {
			t.Fatal("foreign frame accepted")
		}
		if _, err := c.Receive(); !errors.Is(err, ErrProtocol) {
			t.Fatal("connection recovered past malformed input")
		}
	}
	frame := must(EncodeFrame(Frame{Role: UnixSupervisor, Opcode: Started, SessionID: 5, Sequence: 1, Nonce: nonce}))
	c := must(NewChannel(bytes.NewReader(append(frame, frame...)), io.Discard, Parent, UnixSupervisor, 5, nonce))
	if _, err := c.Receive(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Receive(); !errors.Is(err, ErrProtocol) {
		t.Fatal("replayed frame accepted")
	}
}

type shortWriter struct {
	bytes.Buffer
	limit int
}

func (w *shortWriter) Write(p []byte) (int, error) { return w.Buffer.Write(p[:min(w.limit, len(p))]) }

func TestConcurrentSendFramesRemainWholeWithShortWrites(t *testing.T) {
	w := &shortWriter{limit: 3}
	nonce := Nonce{1}
	c := must(NewChannel(bytes.NewReader(nil), w, Parent, WindowsBroker, 5, nonce))
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(b byte) {
			defer wg.Done()
			if err := c.Send(WriteInput, bytes.Repeat([]byte{b}, 100)); err != nil {
				t.Error(err)
			}
		}(byte(i))
	}
	wg.Wait()
	receiver := must(NewChannel(bytes.NewReader(w.Bytes()), io.Discard, WindowsBroker, Parent, 5, nonce))
	seen := map[byte]bool{}
	for i := uint64(1); i <= 100; i++ {
		f := must(receiver.Receive())
		if f.Sequence != i || len(f.Payload) != 100 || seen[f.Payload[0]] {
			t.Fatal("torn or duplicate frame")
		}
		seen[f.Payload[0]] = true
		for _, b := range f.Payload {
			if b != f.Payload[0] {
				t.Fatal("interleaved payload")
			}
		}
	}
}

func TestFirstFrameAndWriteFailure(t *testing.T) {
	nonce := Nonce{1}
	buf := must(EncodeFrame(Frame{Role: Parent, Opcode: Start, SessionID: 5, Sequence: 1, Nonce: nonce}))
	c, first, err := AcceptChannel(bytes.NewReader(buf), io.Discard, UnixSupervisor, Parent)
	if err != nil || first.Opcode != Start || c.received != 1 {
		t.Fatal("first frame handshake")
	}
	writer := &shortWriter{limit: 0}
	c = must(NewChannel(bytes.NewReader(nil), writer, Parent, UnixSupervisor, 5, nonce))
	if err := c.Send(Stop, nil); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal("zero write accepted")
	}
	writer.limit = 100
	if err := c.Send(Stop, nil); !errors.Is(err, io.ErrShortWrite) || writer.Len() != 0 {
		t.Fatal("partial transport failure retried")
	}
}
