package runtime

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
)

// nativeSpecification translates values only. Executable/PATH lookup belongs
// to the native client after cwd acquisition; it must never run in this parent.
func nativeSpecification(n nativeStart) (broker.StartSpec, error) {
	if !n.ID.Valid() || !n.OperationID.Valid() || !n.Invocation.Valid() || n.Output == nil {
		return broker.StartSpec{}, errInvalid
	}
	d := n.Invocation.Data()
	var argv api.ArgvExecution
	switch execution := d.Execution.(type) {
	case api.ArgvExecution:
		argv = execution
	case api.InteractiveShell:
		configured, ok := execution.Data().Policy.(api.ConfiguredShell)
		if !ok {
			// Auto is resolved by the parent constructor before this private
			// translation. No native broker may inspect its own ancestry.
			return broker.StartSpec{}, errInvalid
		}
		argv = configured.Data().Execution
	default:
		return broker.StartSpec{}, errInvalid
	}
	cwd := d.Cwd.Data()
	worktree := cwd.Worktree.Data()
	g := d.Geometry.Data()
	a := argv.Data()
	return broker.StartSpec{ParentID: uint64(os.Getpid()), OperationID: n.OperationID.Value(), RootLocator: worktree.RootLocator, Components: cwd.ProjectComponents, RootIdentity: worktree.RootIdentity, ProjectIdentity: cwd.ProjectIdentity, Executable: a.Executable, Arguments: a.Arguments, Environment: append([]string(nil), n.Environment...), Terminal: d.Terminal == api.Terminal, Rows: uint16(g.Rows), Columns: uint16(g.Columns)}, nil
}

// bridgeState retains only immutable public facts. The native client continues
// owning every OS resource; this mutex never spans a native call or wait.
type bridgeState struct {
	mu        sync.Mutex
	start     nativeStart
	cwd       api.Optional[api.AcquiredCwd]
	exit      api.Optional[api.SessionExit]
	requested bool
}

func (b *bridgeState) established(locator string) api.Optional[api.AcquiredCwd] {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.cwd.Present() {
		d := api.AcquiredCwdData{Observation: b.start.Invocation.Data().Cwd}
		if locator != "" {
			d.ActualLocator = api.Some(locator)
		} else {
			d.Diagnostics = []api.Diagnostic{diagnostic(api.Unavailable, "runtime.cwd_locator_unavailable", "Initial cwd identity was acquired; its actual locator was unavailable when startup waiting ended.")}
		}
		b.cwd = api.Some(owned(api.NewAcquiredCwd(d)))
	}
	return b.cwd
}

func (b *bridgeState) exitFact(established bool, code api.Optional[int], signal api.Optional[string]) api.Optional[api.SessionExit] {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.exit.Present() {
		cause := api.NaturalExit
		if !established {
			cause = api.FailedStartExit
		} else if b.requested {
			cause = api.RequestedExit
		}
		b.exit = api.Some(owned(api.NewSessionExit(api.SessionExitData{Code: code, Signal: signal, Cause: cause})))
	}
	return b.exit
}

func (b *bridgeState) requestStop() { b.mu.Lock(); b.requested = true; b.mu.Unlock() }

func (b *bridgeState) residual(stage api.RuntimeCleanupStage, err error) api.RuntimeResidual {
	return owned(api.NewRuntimeResidual(api.RuntimeResidualData{SessionID: api.Some(b.start.ID), Stage: stage, Detail: safeDiagnostic(err)}))
}

// Native failures are converted to safe API values before reaching the common
// engine. Joining multiple failures preserves each code/stage as a diagnostic;
// neither Error nor the API copies native paths, arguments or error text.
type bridgeError struct{ values []api.Diagnostic }

func (e *bridgeError) Error() string {
	parts := make([]string, len(e.values))
	for i, d := range e.values {
		parts[i] = d.Data().Message
	}
	return strings.Join(parts, " ")
}
func (e *bridgeError) Unwrap() []error {
	result := make([]error, len(e.values))
	for i, d := range e.values {
		result[i] = d
	}
	return result
}

func normalizeNativeError(err error, stage api.RuntimeCleanupStage, classify func(error) (api.ErrorCode, api.RuntimeCleanupStage, bool)) error {
	if err == nil {
		return nil
	}
	var values []api.Diagnostic
	seen := make(map[string]bool)
	var visit func(error)
	visit = func(e error) {
		if e == nil {
			return
		}
		code, at, known := classify(e)
		if !known {
			if joined, ok := e.(interface{ Unwrap() []error }); ok {
				for _, child := range joined.Unwrap() {
					visit(child)
				}
				return
			}
			if wrapped, ok := e.(interface{ Unwrap() error }); ok {
				visit(wrapped.Unwrap())
				return
			}
			at, code = stage, api.IOFailure
			switch {
			case errors.Is(e, context.Canceled), errors.Is(e, context.DeadlineExceeded):
				code = api.Canceled
			case errors.Is(e, os.ErrNotExist):
				code = api.NotFound
			case errors.Is(e, os.ErrPermission):
				code = api.Permission
			case errors.Is(e, broker.ErrCwd):
				code = api.StaleObservation
			case errors.Is(e, broker.ErrProtocol):
				code = api.Invalid
			}
		}
		if !at.Valid() {
			at = stage
		}
		if !code.Valid() {
			code = api.Indeterminate
		}
		reason := "runtime.native." + strconv.Itoa(int(at)) + "." + strconv.Itoa(int(code))
		if !seen[reason] {
			seen[reason] = true
			stages := [...]string{"", "session acquisition", "process containment", "working directory acquisition", "process waiting", "descendant cleanup", "terminal cleanup", "input", "output", "session control", "session supervisor", "outer containment", "helper extraction", "event transfer"}
			causes := [...]string{"", "invalid request or protocol", "unavailable", "not found", "busy", "canceled", "superseded", "changed observation", "stale confirmation", "conflict", "permission denied", "unsupported", "I/O failure", "process failure", "cleanup incomplete", "outcome unknown"}
			values = append(values, diagnostic(code, reason, "Runtime "+stages[at]+": "+causes[code]+"."))
		}
	}
	visit(err)
	if len(values) == 0 {
		return nil
	}
	return &bridgeError{values}
}
