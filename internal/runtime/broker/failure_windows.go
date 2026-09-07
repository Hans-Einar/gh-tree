package broker

import (
	"context"
	"errors"
	"os"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

// These values belong only to the Windows Runtime seam. They carry no native
// status number, path, argument, environment value or handle over the channel.
type WindowsFailureCause uint8

const (
	WindowsCwdFailure WindowsFailureCause = iota + 1
	WindowsNotFoundFailure
	WindowsPermissionFailure
	WindowsUnsupportedFailure
	WindowsProcessFailure
	WindowsProtocolFailure
	WindowsCanceledFailure
	WindowsTimeoutFailure
	WindowsIOFailure
	WindowsBusyFailure
)

var (
	ErrWindowsUnsupported = errors.New("unsupported Windows runtime profile")
	ErrWindowsProcess     = errors.New("Windows runtime process failure")
	ErrWindowsCleanup     = errors.New("Windows runtime cleanup incomplete")
	ErrWindowsBusy        = errors.New("Windows runtime control busy")
)

// WindowsFailure preserves the source stage and cause independently of whether
// a subsequent cleanup succeeds. Cleanup marks a failed cleanup operation; it
// is historical diagnostics, not evidence that resources still remain after
// the parent's final outer Job and I/O barriers. Only locally created failures
// retain the original error. Error never formats that private underlying value.
type WindowsFailure struct {
	Cause   WindowsFailureCause
	Stage   api.RuntimeCleanupStage
	Cleanup bool
	local   error
}

func (f *WindowsFailure) valid() bool {
	return f != nil && f.Cause >= WindowsCwdFailure && f.Cause <= WindowsBusyFailure && f.Stage.Valid()
}

func (f *WindowsFailure) Error() string {
	if !f.valid() {
		return "invalid Windows runtime failure"
	}
	causes := [...]string{"", "stale cwd", "not found", "permission denied", "unsupported profile", "process failure", "protocol failure", "canceled", "timeout", "I/O failure", "busy"}
	stages := [...]string{"", "acquisition", "process containment", "cwd acquisition", "user process wait", "descendants", "terminal", "input", "output", "control", "broker", "outer containment", "helper extraction", "event transfer"}
	message := "Windows runtime " + causes[f.Cause] + " at " + stages[f.Stage]
	if f.Cleanup {
		message += " during cleanup"
	}
	return message
}

func (f *WindowsFailure) Unwrap() error { return f.local }

func (f *WindowsFailure) Is(target error) bool {
	if !f.valid() {
		return false
	}
	if target == ErrWindowsCleanup && f.Cleanup {
		return true
	}
	switch f.Cause {
	case WindowsCwdFailure:
		return target == ErrCwd
	case WindowsNotFoundFailure:
		return target == os.ErrNotExist
	case WindowsPermissionFailure:
		return target == os.ErrPermission
	case WindowsUnsupportedFailure:
		return target == ErrWindowsUnsupported
	case WindowsProcessFailure:
		return target == ErrWindowsProcess
	case WindowsProtocolFailure:
		return target == ErrProtocol
	case WindowsCanceledFailure:
		return target == context.Canceled
	case WindowsTimeoutFailure:
		return target == context.DeadlineExceeded
	case WindowsBusyFailure:
		return target == ErrWindowsBusy
	}
	return false
}

func windowsFailureAt(err error, stage api.RuntimeCleanupStage) error {
	if err == nil {
		return nil
	}
	var existing *WindowsFailure
	if errors.As(err, &existing) {
		return err
	}
	cause := WindowsIOFailure
	if stage == api.ProcessContainment || stage == api.UserProcessWait || stage == api.SupervisorOrBroker {
		cause = WindowsProcessFailure
	}
	// NtCreateFile returns NTSTATUS, whereas Win32 calls and os return errno.
	// Convert that typed status using the OS mapping, never its error text.
	semantic := err
	var status windows.NTStatus
	if errors.As(err, &status) {
		semantic = errors.Join(err, status.Errno())
	}
	switch {
	case errors.Is(semantic, ErrCwd):
		cause = WindowsCwdFailure
	case errors.Is(semantic, os.ErrNotExist):
		cause = WindowsNotFoundFailure
	case errors.Is(semantic, os.ErrPermission), errors.Is(semantic, windows.ERROR_PRIVILEGE_NOT_HELD):
		cause = WindowsPermissionFailure
	case errors.Is(semantic, ErrWindowsUnsupported), errors.Is(semantic, windows.ERROR_NOT_SUPPORTED), errors.Is(semantic, windows.ERROR_CALL_NOT_IMPLEMENTED), errors.Is(semantic, windows.ERROR_PROC_NOT_FOUND):
		cause = WindowsUnsupportedFailure
	case errors.Is(semantic, ErrProtocol):
		cause = WindowsProtocolFailure
	case errors.Is(semantic, context.Canceled):
		cause = WindowsCanceledFailure
	case errors.Is(semantic, context.DeadlineExceeded):
		cause = WindowsTimeoutFailure
	case errors.Is(semantic, ErrWindowsBusy):
		cause = WindowsBusyFailure
	case errors.Is(semantic, ErrWindowsProcess):
		cause = WindowsProcessFailure
	}
	return &WindowsFailure{Cause: cause, Stage: stage, local: err}
}

func windowsCleanupFailure(err error, stage api.RuntimeCleanupStage) error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		children := joined.Unwrap()
		failures := make([]error, len(children))
		for i, child := range children {
			failures[i] = windowsCleanupFailure(child, stage)
		}
		return errors.Join(failures...)
	}
	if typed, ok := err.(*WindowsFailure); ok {
		copy := *typed
		copy.Cleanup = true
		return &copy
	}
	var existing *WindowsFailure
	if errors.As(err, &existing) {
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			return windowsCleanupFailure(wrapped.Unwrap(), stage)
		}
	}
	failure := windowsFailureAt(err, stage)
	var typed *WindowsFailure
	if errors.As(failure, &typed) {
		copy := *typed
		copy.Cleanup = true
		return &copy
	}
	return failure
}

// The fixed small record set preserves the primary failure plus independent
// cleanup-stage failures. The bound exceeds every engine cleanup call site.
const maxWindowsFailures = 16

func encodeWindowsFailure(err error) ([]byte, error) {
	var failures []*WindowsFailure
	var visit func(error) bool
	visit = func(err error) bool {
		if err == nil {
			return true
		}
		if typed, ok := err.(*WindowsFailure); ok {
			if !typed.valid() {
				return false
			}
			for _, prior := range failures {
				if prior.Cause == typed.Cause && prior.Stage == typed.Stage && prior.Cleanup == typed.Cleanup {
					return true
				}
			}
			failures = append(failures, typed)
			return len(failures) <= maxWindowsFailures
		}
		if joined, ok := err.(interface{ Unwrap() []error }); ok {
			for _, child := range joined.Unwrap() {
				if !visit(child) {
					return false
				}
			}
			return true
		}
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			return visit(wrapped.Unwrap())
		}
		return visit(windowsFailureAt(err, api.SupervisorOrBroker))
	}
	if !visit(err) || len(failures) == 0 {
		return nil, ErrProtocol
	}
	data := []byte{1, byte(len(failures))}
	for _, failure := range failures {
		cleanup := byte(0)
		if failure.Cleanup {
			cleanup = 1
		}
		data = append(data, byte(failure.Cause), byte(failure.Stage), cleanup)
	}
	return data, nil
}

// decodeWindowsFailure returns only validated, safe typed failures. Malformed,
// unknown, duplicate or trailing fields are a protocol failure, never success.
func decodeWindowsFailure(data []byte) error {
	if len(data) < 2 || data[0] != 1 || data[1] == 0 || data[1] > maxWindowsFailures || len(data) != 2+3*int(data[1]) {
		return ErrProtocol
	}
	failures := make([]error, 0, int(data[1]))
	seen := make(map[[3]byte]bool, int(data[1]))
	for i := 2; i < len(data); i += 3 {
		key := [3]byte{data[i], data[i+1], data[i+2]}
		failure := &WindowsFailure{Cause: WindowsFailureCause(key[0]), Stage: api.RuntimeCleanupStage(key[1]), Cleanup: key[2] == 1}
		if !failure.valid() || key[2] > 1 || seen[key] {
			return ErrProtocol
		}
		seen[key] = true
		failures = append(failures, failure)
	}
	if len(failures) == 1 {
		return failures[0]
	}
	return errors.Join(failures...)
}
