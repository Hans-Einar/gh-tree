package api_test

import (
	"context"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// This deliberately narrow fake is a compile assertion, not coordinator proof.
type clientFixture struct{}

var _ api.Client = clientFixture{}

func (clientFixture) Submit(context.Context, api.Request) (api.Receipt, error) {
	return api.Receipt{}, nil
}
func (clientFixture) Confirm(context.Context, api.ConfirmationID, api.Choice) error { return nil }
func (clientFixture) Cancel(api.OperationID) error                                  { return nil }
func (clientFixture) Next(context.Context) (api.Event, error)                       { return api.Event{}, nil }
func (clientFixture) Shutdown(context.Context) api.ShutdownResult                   { return api.ShutdownResult{} }

var _ api.Command = api.WriteInputCommand{}
var _ api.Query = api.SessionOutputQuery{}
var _ api.EventPayload = api.OperationTerminal{}
var _ api.Result = api.CreatePullRequestResult{}
var _ api.Result = api.FetchResult{}
var _ api.Result = api.SessionOutputProjection{}
