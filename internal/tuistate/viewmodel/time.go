package viewmodel

import "time"

// Timestamp preserves a supplied source instant and its explicitly supplied
// original offset. It is not an observation interval or freshness signal. The
// private instant is UTC to avoid implicit caller-local timezone discovery;
// the meaningful original offset remains available for View formatting.
type Timestamp struct {
	value         time.Time
	offsetSeconds int
}

func NewTimestamp(value time.Time, originalOffsetSeconds int) (Timestamp, error) {
	if value.IsZero() || originalOffsetSeconds <= -24*60*60 || originalOffsetSeconds >= 24*60*60 {
		return Timestamp{}, invalid("timestamp")
	}
	return Timestamp{value.UTC(), originalOffsetSeconds}, nil
}
func (t Timestamp) Valid() bool {
	return !t.value.IsZero() && t.value.Location() == time.UTC && t.offsetSeconds > -24*60*60 && t.offsetSeconds < 24*60*60
}
func (t Timestamp) Time() time.Time            { return t.value }
func (t Timestamp) OriginalOffsetSeconds() int { return t.offsetSeconds }
