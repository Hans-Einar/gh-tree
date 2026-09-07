//go:build linux || darwin || freebsd

package runtime

import (
	"context"
	"errors"
	"io"
	"syscall"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
)

type unixOwner struct {
	bridgeState
	client *broker.UnixClient
	last   nativeFact // NextFact has exactly one parent observer.
}

func unixError(err error, stage api.RuntimeCleanupStage) error {
	return normalizeNativeError(err, stage, func(e error) (api.ErrorCode, api.RuntimeCleanupStage, bool) {
		f, ok := e.(broker.UnixFailure)
		return f.Code, f.Stage, ok
	})
}

func startUnix(ctx context.Context, n nativeStart) (nativeOwner, nativeStartFact, error) {
	spec, err := nativeSpecification(n)
	if err != nil {
		return nil, nativeStartFact{}, err
	}
	client, fact, err := broker.StartUnix(ctx, broker.UnixConfig{SessionID: n.ID.Value(), Spec: spec, Output: n.Output, GracePeriod: n.Grace, ForcePeriod: n.Force})
	if client == nil {
		return nil, nativeStartFact{}, unixError(err, api.Acquisition)
	}
	o := &unixOwner{bridgeState: bridgeState{start: n}, client: client}
	result := nativeStartFact{Established: fact.Established}
	if fact.Established {
		result.Cwd = o.established(fact.Cwd)
	}
	return o, result, unixError(err, api.Acquisition)
}

func (o *unixOwner) NextFact(ctx context.Context) (nativeFact, error) {
	f, err := o.client.NextFact(ctx)
	if errors.Is(err, io.EOF) {
		if o.last.CleanupComplete {
			return o.last, nil
		}
		return o.last, errCleanup
	}
	if err != nil {
		return o.last, unixError(err, api.SupervisorOrBroker)
	}
	result := nativeFact{Established: f.Established, CleanupComplete: f.CleanupComplete}
	if f.Established {
		result.Cwd = o.established("")
	}
	if f.RootExited {
		code, signal := api.Some(f.ExitCode), api.None[string]()
		if f.Signal != 0 {
			code, signal = api.None[int](), api.Some(syscall.Signal(f.Signal).String())
		}
		result.Exit = o.exitFact(f.Established, code, signal)
	}
	if f.Err != nil {
		failure := unixError(f.Err, api.SupervisorOrBroker)
		result.Diagnostics = diagnostics(failure)
		// Native Unix facts include pending active resources. Only a failed
		// observation makes those barriers failure residuals at this boundary.
		for _, residual := range f.Residuals {
			result.Residuals = append(result.Residuals, o.residual(residual.Stage, unixError(residual.Err, residual.Stage)))
		}
	}
	o.last = result
	return result, nil
}

func (o *unixOwner) Write(ctx context.Context, data []byte) (nativeDelivery, error) {
	d, err := o.client.Write(ctx, data)
	return nativeDelivery{Accepted: d.Accepted, Delivered: d.Delivered, Completed: d.Completed, Dispatched: d.Accepted != 0}, unixError(err, api.InputCleanup)
}
func (o *unixOwner) Resize(ctx context.Context, g api.Geometry) (nativeDelivery, error) {
	d, err := o.client.Resize(ctx, uint16(g.Data().Rows), uint16(g.Data().Columns))
	result := nativeDelivery{Completed: d.Completed, Dispatched: d.Completed}
	if d.Completed {
		result.Delivered = 1
	}
	return result, unixError(err, api.TerminalCleanup)
}
func (o *unixOwner) Interrupt(ctx context.Context) (nativeDelivery, error) {
	d, err := o.client.Interrupt(ctx)
	return nativeDelivery{Accepted: d.Accepted, Delivered: d.Delivered, Completed: d.Completed, Dispatched: d.Accepted != 0}, unixError(err, api.InputCleanup)
}
func (o *unixOwner) Stop() { o.requestStop(); o.client.Stop() }
