package persistence

import "github.com/Hans-Einar/gh-tree/internal/application/api"

// A retained payload is the published inode and may receive subsequent writes.
// That observation changes its SourceVersion, never its persisted RecoveryID or
// the historical request's Original/Proposed. It is not a conflict with a newly
// loaded current-document Expected version.
type recoveryPayloadChanged struct{}

func (*recoveryPayloadChanged) Error() string {
	return "retained payload bytes changed after preparation"
}
func (*recoveryPayloadChanged) diagnostic() api.Diagnostic {
	v, _ := api.NewDiagnostic(api.DiagnosticData{Code: api.Conflict, Reason: "storage-retained-payload-changed", Message: "A retained payload object has changed since preparation; recorded document versions describe the earlier request."})
	return v
}

func recoveryNotices(err error) ([]api.Diagnostic, bool) {
	if err == nil {
		return nil, true
	}
	if notice, ok := err.(*recoveryPayloadChanged); ok {
		return []api.Diagnostic{notice.diagnostic()}, true
	}
	if group, ok := err.(interface{ Unwrap() []error }); ok {
		var notices []api.Diagnostic
		for _, child := range group.Unwrap() {
			found, only := recoveryNotices(child)
			if !only {
				return nil, false
			}
			notices = append(notices, found...)
		}
		return notices, true
	}
	return nil, false
}
