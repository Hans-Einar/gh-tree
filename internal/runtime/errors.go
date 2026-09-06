// Package runtime owns persistent process sessions and their native resources.
package runtime

import "github.com/Hans-Einar/gh-tree/internal/application/api"

func diagnostic(code api.ErrorCode, reason, message string) api.Diagnostic {
	d, err := api.NewDiagnostic(api.DiagnosticData{Code: code, Reason: reason, Message: message})
	if err != nil {
		panic(err) // Only static implementation-owned values enter this helper.
	}
	return d
}

var (
	errInvalid   = diagnostic(api.Invalid, "runtime.invalid", "Invalid Runtime request.")
	errBusy      = diagnostic(api.Busy, "runtime.capacity", "Runtime capacity is awaiting cleanup or acknowledgment.")
	errClosed    = diagnostic(api.Unavailable, "runtime.closed", "Runtime admission is closed.")
	errExhausted = diagnostic(api.Unavailable, "runtime.sequence_exhausted", "Runtime identity or sequence space is exhausted.")
	errNotFound  = diagnostic(api.NotFound, "runtime.session_missing", "The session is unknown or its cleaned history was evicted.")
	errCursor    = diagnostic(api.Invalid, "runtime.event_cursor", "The event cursor was not delivered or precedes acknowledgment.")
)
