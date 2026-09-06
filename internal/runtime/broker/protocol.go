// Package broker contains Runtime-private native workers and their transports.
// Its wire types never appear in the public Sessions surface.
package broker

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
)

const ProtocolVersion uint16 = 1
const MaxFrame = 64 << 10
const headerSize = 52

type Role uint8

const (
	Parent           Role = 1
	UnixSupervisor   Role = 2
	UnixSignalHelper Role = 3
	WindowsBroker    Role = 4
)

type Opcode uint8

const (
	Start          Opcode = 1
	WriteInput     Opcode = 2
	Resize         Opcode = 3
	Interrupt      Opcode = 4
	Stop           Opcode = 5
	Abort          Opcode = 6
	Release        Opcode = 7
	Started        Opcode = 8
	UserExit       Opcode = 9
	Quiescent      Opcode = 10
	Failure        Opcode = 11
	Prepare        Opcode = 12
	Joined         Opcode = 13
	Commit         Opcode = 14
	OutputTransfer Opcode = 15
	OutputAccepted Opcode = 16
	Delivered      Opcode = 17
)

type Nonce [32]byte

func FreshNonce() (Nonce, error) {
	var n Nonce
	_, err := rand.Read(n[:])
	if n == (Nonce{}) && err == nil {
		return n, errors.New("zero private nonce")
	}
	return n, err
}

// Frame contains one owned, bounded body. Role identifies the sending endpoint;
// each direction has its own strictly monotonic sequence starting at one.
type Frame struct {
	Role      Role
	Opcode    Opcode
	SessionID uint64
	Sequence  uint64
	Nonce     Nonce
	Payload   []byte
}

var ErrProtocol = errors.New("invalid Runtime private protocol")

func (f Frame) valid() bool {
	return f.Role >= Parent && f.Role <= WindowsBroker && f.Opcode >= Start && f.Opcode <= Delivered && f.SessionID != 0 && f.Sequence != 0 && f.Nonce != (Nonce{}) && len(f.Payload) <= MaxFrame-headerSize
}

func EncodeFrame(f Frame) ([]byte, error) {
	if !f.valid() {
		return nil, ErrProtocol
	}
	buf := make([]byte, 4+headerSize+len(f.Payload))
	binary.BigEndian.PutUint32(buf, uint32(len(buf)-4))
	binary.BigEndian.PutUint16(buf[4:], ProtocolVersion)
	buf[6], buf[7] = byte(f.Role), byte(f.Opcode)
	binary.BigEndian.PutUint64(buf[8:], f.SessionID)
	binary.BigEndian.PutUint64(buf[16:], f.Sequence)
	copy(buf[24:56], f.Nonce[:])
	copy(buf[56:], f.Payload)
	return buf, nil
}

func DecodeFrame(reader io.Reader) (Frame, error) {
	var prefix [4]byte
	if _, err := io.ReadFull(reader, prefix[:]); err != nil {
		return Frame{}, err
	}
	n := binary.BigEndian.Uint32(prefix[:])
	if n < headerSize || n > MaxFrame {
		return Frame{}, ErrProtocol
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return Frame{}, err
	}
	if binary.BigEndian.Uint16(buf) != ProtocolVersion {
		return Frame{}, ErrProtocol
	}
	f := Frame{Role: Role(buf[2]), Opcode: Opcode(buf[3]), SessionID: binary.BigEndian.Uint64(buf[4:]), Sequence: binary.BigEndian.Uint64(buf[12:])}
	copy(f.Nonce[:], buf[20:52])
	f.Payload = buf[52:]
	if !f.valid() {
		return Frame{}, ErrProtocol
	}
	return f, nil
}

// Channel has exactly one Receive owner. Send serializes complete frames;
// native owners configure deadlines/cancellation and close/join endpoints.
// It creates no goroutine and does not take any registry/session lock.
type Channel struct {
	reader         io.Reader
	writer         io.Writer
	role, peer     Role
	session        uint64
	nonce          Nonce
	received       uint64
	sendMu         sync.Mutex
	sent           uint64
	sendFailure    error
	receiveFailure error
}

func NewChannel(reader io.Reader, writer io.Writer, role, peer Role, session uint64, nonce Nonce) (*Channel, error) {
	if reader == nil || writer == nil || role < Parent || role > WindowsBroker || peer < Parent || peer > WindowsBroker || role == peer || session == 0 || nonce == (Nonce{}) {
		return nil, ErrProtocol
	}
	return &Channel{reader: reader, writer: writer, role: role, peer: peer, session: session, nonce: nonce}, nil
}

// AcceptChannel consumes the first frame only. Endpoint kind, exact parent and
// native role/SID are independently verified by the native owner before calling
// this constructor. A nonce from a pathname/socket is never sufficient authority.
func AcceptChannel(reader io.Reader, writer io.Writer, role, peer Role) (*Channel, Frame, error) {
	f, err := DecodeFrame(reader)
	if err != nil {
		return nil, Frame{}, err
	}
	if f.Role != peer || f.Sequence != 1 {
		return nil, Frame{}, ErrProtocol
	}
	c, err := NewChannel(reader, writer, role, peer, f.SessionID, f.Nonce)
	if err != nil {
		return nil, Frame{}, err
	}
	c.received = 1
	return c, f, nil
}

func (c *Channel) Send(op Opcode, payload []byte) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.sendFailure != nil {
		return c.sendFailure
	}
	if c.sent == math.MaxUint64 {
		return ErrProtocol
	}
	buf, err := EncodeFrame(Frame{Role: c.role, Opcode: op, SessionID: c.session, Sequence: c.sent + 1, Nonce: c.nonce, Payload: payload})
	if err != nil {
		return err
	}
	for len(buf) > 0 {
		n, e := c.writer.Write(buf)
		if n < 0 || n > len(buf) {
			e = io.ErrShortWrite
			n = 0
		}
		buf = buf[n:]
		if e != nil {
			c.sendFailure = e
			return e
		}
		if n == 0 {
			c.sendFailure = io.ErrShortWrite
			return c.sendFailure
		}
	}
	c.sent++
	return nil
}

func (c *Channel) Receive() (Frame, error) {
	if c.receiveFailure != nil {
		return Frame{}, c.receiveFailure
	}
	f, err := DecodeFrame(c.reader)
	if err == nil && (f.Role != c.peer || f.SessionID != c.session || c.received == math.MaxUint64 || f.Sequence != c.received+1 || subtle.ConstantTimeCompare(f.Nonce[:], c.nonce[:]) != 1) {
		err = ErrProtocol
	}
	if err != nil {
		c.receiveFailure = err
		return Frame{}, err
	}
	c.received++
	return f, nil
}
