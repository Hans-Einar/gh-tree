package runtime

import (
	"context"
	"debug/pe"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
	"github.com/Hans-Einar/gh-tree/internal/runtime/brokerassets"
	"golang.org/x/sys/windows"
)

type windowsOwner struct {
	bridgeState
	client *broker.WindowsClient
	last   nativeFact
}

func windowsError(err error, stage api.RuntimeCleanupStage) error {
	return normalizeNativeError(err, stage, func(e error) (api.ErrorCode, api.RuntimeCleanupStage, bool) {
		if f, ok := e.(*broker.WindowsFailure); ok && f != nil {
			codes := [...]api.ErrorCode{0, api.StaleObservation, api.NotFound, api.Permission, api.Unsupported, api.ProcessFailure, api.Invalid, api.Canceled, api.Canceled, api.IOFailure, api.Busy}
			if int(f.Cause) < len(codes) && f.Cause > 0 {
				return codes[f.Cause], f.Stage, true
			}
			return api.Indeterminate, stage, true
		}
		if status, ok := e.(windows.NTStatus); ok {
			e = status.Errno()
		}
		switch {
		case errors.Is(e, broker.ErrWindowsUnsupported), errors.Is(e, windows.ERROR_NOT_SUPPORTED), errors.Is(e, windows.ERROR_CALL_NOT_IMPLEMENTED), errors.Is(e, windows.ERROR_PROC_NOT_FOUND):
			return api.Unsupported, stage, true
		case errors.Is(e, broker.ErrWindowsBusy), errors.Is(e, broker.ErrWindowsControlsBusy):
			return api.Busy, stage, true
		case errors.Is(e, broker.ErrWindowsProcess):
			return api.ProcessFailure, stage, true
		case errors.Is(e, os.ErrNotExist):
			return api.NotFound, stage, true
		case errors.Is(e, os.ErrPermission), errors.Is(e, windows.ERROR_PRIVILEGE_NOT_HELD):
			return api.Permission, stage, true
		}
		return 0, stage, false
	})
}

func startWindows(ctx context.Context, n nativeStart) (nativeOwner, nativeStartFact, error) {
	spec, err := nativeSpecification(n)
	if err != nil {
		return nil, nativeStartFact{}, err
	}
	if ctx == nil {
		return nil, nativeStartFact{}, errInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, nativeStartFact{}, err
	}
	machine, embedded, err := broker.MachineRoute()
	if err != nil {
		return nil, nativeStartFact{}, windowsError(err, api.SupervisorOrBroker)
	}
	var extraction *broker.WindowsImage
	image := ""
	if embedded {
		arch := ""
		switch machine {
		case pe.IMAGE_FILE_MACHINE_AMD64:
			arch = "amd64"
		case pe.IMAGE_FILE_MACHINE_ARM64:
			arch = "arm64"
		default:
			return nil, nativeStartFact{}, errUnsupported
		}
		asset, err := brokerassets.Load(arch)
		if err != nil {
			return nil, nativeStartFact{}, windowsError(err, api.HelperExtraction)
		}
		bytes, err := hex.DecodeString(asset.SHA256)
		if err != nil || len(bytes) != 32 {
			return nil, nativeStartFact{}, errInvalid
		}
		var digest [32]byte
		copy(digest[:], bytes)
		extraction, err = broker.ExtractWindowsImage(asset.Bytes, machine, digest, asset.Protocol)
		if err != nil {
			return ownWindowsExtraction(n, extraction, err)
		}
		image = extraction.Path()
	} else {
		image, err = os.Executable()
		if err != nil {
			return nil, nativeStartFact{}, windowsError(err, api.SupervisorOrBroker)
		}
	}
	client, fact, err := broker.StartWindows(ctx, broker.WindowsConfig{SessionID: n.ID.Value(), Spec: spec, Image: image, Extraction: extraction, Output: n.Output, GracePeriod: n.Grace, ForcePeriod: n.Force})
	if client == nil {
		if extraction != nil {
			return ownWindowsExtraction(n, extraction, err)
		}
		return nil, nativeStartFact{}, windowsError(err, api.Acquisition)
	}
	o := &windowsOwner{bridgeState: bridgeState{start: n}, client: client}
	result := nativeStartFact{Established: fact.Established}
	if fact.Established {
		result.Cwd = o.established(fact.Cwd)
	}
	return o, result, windowsError(err, api.Acquisition)
}

// An extracted image is already an acquired resource even if client admission
// fails. This retained owner retries exact cleanup on an explicit parent cleanup
// observation; no user process exists, and nil ownership would be dishonest.
type windowsExtractionOwner struct {
	bridgeState
	image   *broker.WindowsImage
	failure error
}

func ownWindowsExtraction(n nativeStart, image *broker.WindowsImage, err error) (nativeOwner, nativeStartFact, error) {
	err = windowsError(err, api.HelperExtraction)
	if image == nil {
		return nil, nativeStartFact{}, err
	}
	return &windowsExtractionOwner{bridgeState: bridgeState{start: n}, image: image, failure: err}, nativeStartFact{}, err
}

func (o *windowsExtractionOwner) NextFact(ctx context.Context) (nativeFact, error) {
	if err := ctx.Err(); err != nil {
		return nativeFact{}, err
	}
	err := windowsError(o.image.Cleanup(), api.HelperExtraction)
	f := nativeFact{CleanupComplete: err == nil, Diagnostics: diagnostics(o.failure)}
	if err != nil {
		f.Diagnostics = append(f.Diagnostics, diagnostics(err)...)
		f.Residuals = []api.RuntimeResidual{o.residual(api.HelperExtraction, err)}
		return f, errCleanup
	}
	return f, nil
}
func (o *windowsExtractionOwner) Write(context.Context, []byte) (nativeDelivery, error) {
	return nativeDelivery{}, errClosed
}
func (o *windowsExtractionOwner) Resize(context.Context, api.Geometry) (nativeDelivery, error) {
	return nativeDelivery{}, errClosed
}
func (o *windowsExtractionOwner) Interrupt(context.Context) (nativeDelivery, error) {
	return nativeDelivery{}, errClosed
}
func (o *windowsExtractionOwner) Stop() { o.requestStop() }

func (o *windowsOwner) NextFact(ctx context.Context) (nativeFact, error) {
	f, err := o.client.NextFact(ctx)
	if errors.Is(err, io.EOF) {
		if o.last.CleanupComplete {
			return o.last, nil
		}
		return o.last, errCleanup
	}
	if err != nil {
		return o.last, windowsError(err, api.SupervisorOrBroker)
	}
	result := nativeFact{Established: f.Established, CleanupComplete: f.CleanupComplete, Diagnostics: diagnostics(windowsError(f.Err, api.SupervisorOrBroker))}
	if f.Established {
		cwd := o.start.Invocation.Data().Cwd.Data()
		// The native Windows startup guards bind this initial locator through
		// the verified child-cwd barrier; no post-start pathname claim is made.
		result.Cwd = o.established(filepath.Join(append([]string{cwd.Worktree.Data().RootLocator}, cwd.ProjectComponents...)...))
	}
	if f.RootExited {
		result.Exit = o.exitFact(f.Established, api.Some(int(int32(f.ExitCode))), api.None[string]())
	}
	for _, residual := range f.Residuals {
		result.Residuals = append(result.Residuals, o.residual(residual.Stage, windowsError(residual.Err, residual.Stage)))
	}
	o.last = result
	return result, nil
}

type windowsReceipt struct {
	receipt *broker.WindowsReceipt
	resize  bool
	stage   api.RuntimeCleanupStage
}

func mapWindowsDelivery(d broker.WindowsDelivery, resize bool, stage api.RuntimeCleanupStage) nativeDelivery {
	result := nativeDelivery{Accepted: d.Accepted, Delivered: d.Delivered, Completed: d.Completed, Dispatched: d.Dispatched}
	if resize && d.Completed {
		result.Delivered = 1
	}
	if d.Receipt != nil {
		result.Receipt = &windowsReceipt{d.Receipt, resize, stage}
	}
	return result
}
func (r *windowsReceipt) Wait(ctx context.Context) (nativeDelivery, error) {
	d, err := r.receipt.Wait(ctx)
	return mapWindowsDelivery(d, r.resize, r.stage), windowsError(err, r.stage)
}
func (o *windowsOwner) Write(ctx context.Context, b []byte) (nativeDelivery, error) {
	d, e := o.client.Write(ctx, b)
	return mapWindowsDelivery(d, false, api.InputCleanup), windowsError(e, api.InputCleanup)
}
func (o *windowsOwner) Resize(ctx context.Context, g api.Geometry) (nativeDelivery, error) {
	d, e := o.client.Resize(ctx, uint16(g.Data().Rows), uint16(g.Data().Columns))
	return mapWindowsDelivery(d, true, api.TerminalCleanup), windowsError(e, api.TerminalCleanup)
}
func (o *windowsOwner) Interrupt(ctx context.Context) (nativeDelivery, error) {
	d, e := o.client.Interrupt(ctx)
	return mapWindowsDelivery(d, false, api.InputCleanup), windowsError(e, api.InputCleanup)
}
func (o *windowsOwner) Stop() { o.requestStop(); o.client.Stop() }
