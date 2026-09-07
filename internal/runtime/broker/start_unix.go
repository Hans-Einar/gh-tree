//go:build linux || darwin || freebsd

package broker

import "time"

const unixStartVersion = 1
const unixStartHeaderSize = 17 // version and two uint64 nanosecond periods

func validUnixPeriods(grace, force time.Duration) bool {
	return grace >= time.Millisecond && grace <= time.Minute && force >= time.Millisecond && force <= time.Minute
}

func effectiveUnixPeriods(grace, force time.Duration) (time.Duration, time.Duration, error) {
	if grace == 0 {
		grace = 2 * time.Second
	}
	if force == 0 {
		force = 3 * time.Second
	}
	if !validUnixPeriods(grace, force) {
		return 0, 0, ErrProtocol
	}
	return grace, force, nil
}

// Unix-only startup envelope binds effective immutable cleanup periods before
// cwd acquisition, user execution, observation failure or control EOF cleanup.
// The shared StartSpec/Windows helper closure is unchanged. Nanoseconds preserve
// the exact accepted construction values rather than truncating milliseconds.
func encodeUnixStart(spec StartSpec, grace, force time.Duration) ([]byte, error) {
	if !validUnixPeriods(grace, force) {
		return nil, ErrProtocol
	}
	payload, err := EncodeStart(spec)
	if err != nil || len(payload) > MaxFrame-headerSize-unixStartHeaderSize {
		return nil, ErrProtocol
	}
	e := encoder{}
	e.u8(unixStartVersion)
	e.u64(uint64(grace))
	e.u64(uint64(force))
	return append(e.buf, payload...), nil
}

func decodeUnixStart(payload []byte) (StartSpec, time.Duration, time.Duration, error) {
	if len(payload) < unixStartHeaderSize || len(payload) > MaxFrame-headerSize {
		return StartSpec{}, 0, 0, ErrProtocol
	}
	d := decoder{buf: payload}
	version := d.u8()
	grace, force := time.Duration(d.u64()), time.Duration(d.u64())
	if version != unixStartVersion || !validUnixPeriods(grace, force) {
		return StartSpec{}, 0, 0, ErrProtocol
	}
	spec, err := DecodeStart(payload[d.pos:])
	if err != nil {
		return StartSpec{}, 0, 0, err
	}
	return spec, grace, force, nil
}
