package broker

import (
	"encoding/binary"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// StartSpec is a native-only wire record. The main registry retains complete
// immutable API observations and source attribution; this record carries only
// the exact facts needed to acquire cwd and execute the accepted invocation.
// ParentID is validated against the actual inherited-endpoint owner/parent.
type StartSpec struct {
	ParentID        uint64
	OperationID     uint64
	RootLocator     string
	Components      []string
	RootIdentity    api.DirectoryIdentity
	ProjectIdentity api.DirectoryIdentity
	Executable      string
	Arguments       []string
	Environment     []string
	Terminal        bool
	Rows, Columns   uint16
}

func (s StartSpec) valid() bool {
	if s.ParentID == 0 || s.OperationID == 0 || s.RootLocator == "" || strings.ContainsRune(s.RootLocator, 0) || !s.RootIdentity.Valid() || !s.ProjectIdentity.Valid() || s.RootIdentity.Platform() != s.ProjectIdentity.Platform() || s.Executable == "" || strings.ContainsRune(s.Executable, 0) || s.Rows == 0 || s.Columns == 0 || s.Rows > 32767 || s.Columns > 32767 {
		return false
	}
	for _, part := range s.Components {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, "/\\\x00") || (len(part) >= 2 && part[1] == ':') {
			return false
		}
	}
	for _, arg := range s.Arguments {
		if strings.ContainsRune(arg, 0) {
			return false
		}
	}
	for _, entry := range s.Environment {
		if strings.ContainsRune(entry, 0) || !strings.Contains(entry, "=") {
			return false
		}
	}
	return true
}

type encoder struct {
	buf []byte
	err error
}

func (e *encoder) u8(v byte)    { e.buf = append(e.buf, v) }
func (e *encoder) u16(v uint16) { e.buf = binary.BigEndian.AppendUint16(e.buf, v) }
func (e *encoder) u64(v uint64) { e.buf = binary.BigEndian.AppendUint64(e.buf, v) }
func (e *encoder) string(v string) {
	if len(v) > MaxFrame {
		e.err = ErrProtocol
		return
	}
	e.buf = binary.BigEndian.AppendUint32(e.buf, uint32(len(v)))
	e.buf = append(e.buf, v...)
	if len(e.buf) > MaxFrame-headerSize {
		e.err = ErrProtocol
	}
}
func (e *encoder) strings(v []string) {
	if len(v) > MaxFrame/4 {
		e.err = ErrProtocol
		return
	}
	e.u16(uint16(len(v)))
	for _, s := range v {
		if e.err != nil {
			return
		}
		e.string(s)
	}
}
func (e *encoder) identity(v api.DirectoryIdentity) {
	e.u8(byte(v.Platform()))
	e.u64(v.Device())
	id := v.FileID()
	e.buf = append(e.buf, id[:]...)
	e.string(v.Stamp())
}

type decoder struct {
	buf []byte
	pos int
	err error
}

func (d *decoder) take(n int) []byte {
	if d.err != nil || n < 0 || n > len(d.buf)-d.pos {
		d.err = ErrProtocol
		return nil
	}
	b := d.buf[d.pos : d.pos+n]
	d.pos += n
	return b
}
func (d *decoder) u8() byte {
	b := d.take(1)
	if b == nil {
		return 0
	}
	return b[0]
}
func (d *decoder) u16() uint16 {
	b := d.take(2)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint16(b)
}
func (d *decoder) u64() uint64 {
	b := d.take(8)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}
func (d *decoder) string() string {
	b := d.take(4)
	if b == nil {
		return ""
	}
	n := binary.BigEndian.Uint32(b)
	if n > MaxFrame {
		d.err = ErrProtocol
		return ""
	}
	return string(d.take(int(n)))
}
func (d *decoder) strings() []string {
	n := int(d.u16())
	if n > (len(d.buf)-d.pos)/4 {
		d.err = ErrProtocol
		return nil
	}
	v := make([]string, n)
	for i := range v {
		v[i] = d.string()
	}
	return v
}
func (d *decoder) identity() api.DirectoryIdentity {
	platform := api.DirectoryPlatform(d.u8())
	device := d.u64()
	var id [16]byte
	copy(id[:], d.take(16))
	stamp := d.string()
	v, err := api.NewDirectoryIdentity(platform, device, id, stamp)
	if err != nil {
		d.err = ErrProtocol
	}
	return v
}

func EncodeStart(s StartSpec) ([]byte, error) {
	if !s.valid() {
		return nil, ErrProtocol
	}
	e := encoder{}
	e.u64(s.ParentID)
	e.u64(s.OperationID)
	e.string(s.RootLocator)
	e.strings(s.Components)
	e.identity(s.RootIdentity)
	e.identity(s.ProjectIdentity)
	e.string(s.Executable)
	e.strings(s.Arguments)
	e.strings(s.Environment)
	if s.Terminal {
		e.u8(1)
	} else {
		e.u8(0)
	}
	e.u16(s.Rows)
	e.u16(s.Columns)
	if e.err != nil || len(e.buf) > MaxFrame-headerSize {
		return nil, ErrProtocol
	}
	return e.buf, nil
}

func DecodeStart(buf []byte) (StartSpec, error) {
	if len(buf) > MaxFrame-headerSize {
		return StartSpec{}, ErrProtocol
	}
	d := decoder{buf: buf}
	s := StartSpec{ParentID: d.u64(), OperationID: d.u64(), RootLocator: d.string(), Components: d.strings(), RootIdentity: d.identity(), ProjectIdentity: d.identity(), Executable: d.string(), Arguments: d.strings(), Environment: d.strings()}
	terminal := d.u8()
	s.Terminal = terminal == 1
	s.Rows = d.u16()
	s.Columns = d.u16()
	if d.err != nil || d.pos != len(buf) || terminal > 1 || !s.valid() {
		return StartSpec{}, ErrProtocol
	}
	return s, nil
}
