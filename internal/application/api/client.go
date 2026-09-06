package api

import "context"

type Client interface {
	Submit(context.Context, Request) (Receipt, error)
	Confirm(context.Context, ConfirmationID, Choice) error
	Cancel(OperationID) error
	Next(context.Context) (Event, error)
	Shutdown(context.Context) ShutdownResult
}

// Request admits exactly one known, valid concrete command/query. Exact switches
// reject nil, typed nil, pointers and foreign types that embed a sealed variant.
type Request struct {
	command     Command
	query       Query
	correlation Correlation
}

func NewCommandRequest(command Command, correlation Correlation) (Request, error) {
	if !validCommand(command) || !correlation.Valid() || correlation.data.Query.Present() {
		return Request{}, invalid("command request/correlation")
	}
	return Request{command: command, correlation: correlation}, nil
}
func NewQueryRequest(query Query, correlation Correlation) (Request, error) {
	if !validQuery(query) || !correlation.Valid() || !correlation.data.Query.Present() {
		return Request{}, invalid("query request/correlation")
	}
	return Request{query: query, correlation: correlation}, nil
}
func (v Request) Valid() bool {
	if v.command != nil && v.query == nil {
		_, e := NewCommandRequest(v.command, v.correlation)
		return e == nil
	}
	if v.query != nil && v.command == nil {
		_, e := NewQueryRequest(v.query, v.correlation)
		return e == nil
	}
	return false
}
func (v Request) Command() (Command, bool) { return v.command, v.command != nil }
func (v Request) Query() (Query, bool)     { return v.query, v.query != nil }
func (v Request) Correlation() Correlation { return v.correlation }
func (v Request) Clone() Request           { return v }

func (d Diagnostic) Error() string { return d.data.Reason + ": " + d.data.Message }
