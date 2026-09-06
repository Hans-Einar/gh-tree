package api

type OperationID struct{ value uint64 }

func NewOperationID(value uint64) (OperationID, error) {
	if value == 0 {
		return OperationID{}, invalid("OperationID")
	}
	return OperationID{value}, nil
}
func (v OperationID) Value() uint64 { return v.value }
func (v OperationID) Valid() bool   { return v.value > 0 }

type Sequence struct{ value uint64 }

func NewSequence(value uint64) (Sequence, error) {
	if value == 0 {
		return Sequence{}, invalid("Sequence")
	}
	return Sequence{value}, nil
}
func (v Sequence) Value() uint64 { return v.value }
func (v Sequence) Valid() bool   { return v.value > 0 }

type IntentToken struct{ value uint64 }

func NewIntentToken(value uint64) (IntentToken, error) { return IntentToken{value}, nil }
func (v IntentToken) Value() uint64                    { return v.value }
func (v IntentToken) Valid() bool                      { return true }

type QueryGeneration struct{ value uint64 }

func NewQueryGeneration(value uint64) (QueryGeneration, error) { return QueryGeneration{value}, nil }
func (v QueryGeneration) Value() uint64                        { return v.value }
func (v QueryGeneration) Valid() bool                          { return true }

type RuntimeEventSequence struct{ value uint64 }

func NewRuntimeEventSequence(value uint64) (RuntimeEventSequence, error) {
	return RuntimeEventSequence{value}, nil
}
func (v RuntimeEventSequence) Value() uint64 { return v.value }
func (v RuntimeEventSequence) Valid() bool   { return true }

type SessionSequence struct{ value uint64 }

func NewSessionSequence(value uint64) (SessionSequence, error) {
	if value == 0 {
		return SessionSequence{}, invalid("SessionSequence")
	}
	return SessionSequence{value}, nil
}
func (v SessionSequence) Value() uint64 { return v.value }
func (v SessionSequence) Valid() bool   { return v.value > 0 }

type ContextVersion struct{ value uint64 }

func NewContextVersion(value uint64) (ContextVersion, error) { return ContextVersion{value}, nil }
func (v ContextVersion) Value() uint64                       { return v.value }
func (v ContextVersion) Valid() bool                         { return true }

type ConfirmationID struct{ token string }

func NewConfirmationID(token string) (ConfirmationID, error) {
	if !nonempty(token) {
		return ConfirmationID{}, invalid("ConfirmationID")
	}
	return ConfirmationID{token}, nil
}
func (v ConfirmationID) Valid() bool                 { return nonempty(v.token) }
func (v ConfirmationID) Equal(w ConfirmationID) bool { return v == w }

type QuerySlot struct{ token string }

func NewQuerySlot(token string) (QuerySlot, error) {
	if !nonempty(token) {
		return QuerySlot{}, invalid("QuerySlot")
	}
	return QuerySlot{token}, nil
}
func (v QuerySlot) Valid() bool            { return nonempty(v.token) }
func (v QuerySlot) Equal(w QuerySlot) bool { return v == w }

type ObservationID struct{ token string }

func NewObservationID(token string) (ObservationID, error) {
	if !nonempty(token) {
		return ObservationID{}, invalid("ObservationID")
	}
	return ObservationID{token}, nil
}
func (v ObservationID) Valid() bool                { return nonempty(v.token) }
func (v ObservationID) Equal(w ObservationID) bool { return v == w }

type RecoveryID struct{ token string }

func NewRecoveryID(token string) (RecoveryID, error) {
	if !nonempty(token) {
		return RecoveryID{}, invalid("RecoveryID")
	}
	return RecoveryID{token}, nil
}
func (v RecoveryID) Valid() bool             { return nonempty(v.token) }
func (v RecoveryID) Equal(w RecoveryID) bool { return v == w }
