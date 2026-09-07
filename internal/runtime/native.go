package runtime

import (
	"context"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// These are internal ownership contracts, not a public backend API. There is
// deliberately no production constructor/binding until reviewed native clients
// and their matching embedded assets have been adopted.
type nativeStart struct {
	ID           domain.SessionID
	OperationID  api.OperationID
	Invocation   api.Invocation
	Environment  []string
	Output       func(api.OutputStream, []byte)
	Grace, Force time.Duration
}

type nativeStartFact struct {
	Established bool
	Cwd         api.Optional[api.AcquiredCwd]
}

// A nil owner proves that no native resource was acquired. Any partial native
// acquisition MUST return its owner, even with an error or expired context.
type nativeStarter func(context.Context, nativeStart) (nativeOwner, nativeStartFact, error)

type nativeOwner interface {
	NextFact(context.Context) (nativeFact, error)
	Write(context.Context, []byte) (nativeDelivery, error)
	Resize(context.Context, api.Geometry) (nativeDelivery, error)
	Interrupt(context.Context) (nativeDelivery, error)
	Stop()
}

// CleanupComplete proves ALL native resources and output callbacks joined.
// Err may describe historical failure on a complete fact; residuals alone are
// current ownership. Incomplete observations can later be repaired.
type nativeFact struct {
	Established     bool
	Cwd             api.Optional[api.AcquiredCwd]
	Exit            api.Optional[api.SessionExit]
	CleanupComplete bool
	Residuals       []api.RuntimeResidual
	Diagnostic      api.Optional[api.Diagnostic]
	Diagnostics     []api.Diagnostic // bounded native code/stage facts; no error text
}

type nativeDelivery struct {
	Accepted, Delivered uint32
	Completed           bool
	Dispatched          bool
	Receipt             nativeReceipt
}

// Wait returns stable terminal known/unknown delivery facts. Cancellation of a
// public caller cannot abandon a receipt. Parent observation uses its own
// lifetime, holds producer/queue ownership and never replays dispatched bytes.
type nativeReceipt interface {
	Wait(context.Context) (nativeDelivery, error)
}

type sessionBudgets struct{ grace, force, shutdown time.Duration }
