package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

const WindowsPrivateMode = "--gh-tree-runtime-windows-broker-v1"

type receivedFrame struct {
	frame Frame
	err   error
}

// timedIO always joins its operation. Closing a pipe is a resource barrier,
// implemented by Go's tracked pipe I/O and CancelIoEx, not a dropped goroutine.
func timedIO(file *os.File, limit time.Duration, fn func() error) error {
	done := make(chan error, 1)
	go func() { done <- fn() }()
	timer := time.NewTimer(limit)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		closeErr := file.Close()
		return errors.Join(context.DeadlineExceeded, closeErr, <-done)
	}
}

func sendControl(channel *Channel, output *os.File, op Opcode, payload []byte) error {
	return timedIO(output, 3*time.Second, func() error { return channel.Send(op, payload) })
}

func validateWindowsEndpoints(input, output *os.File, parent windows.Handle) (uint64, error) {
	for _, file := range []*os.File{input, output} {
		kind, err := windows.GetFileType(windows.Handle(file.Fd()))
		if err != nil || kind != windows.FILE_TYPE_PIPE {
			return 0, ErrProtocol
		}
	}
	var own, ancestor windows.PROCESS_BASIC_INFORMATION
	if err := windows.NtQueryInformationProcess(windows.CurrentProcess(), windows.ProcessBasicInformation, unsafe.Pointer(&own), uint32(unsafe.Sizeof(own)), nil); err != nil {
		return 0, err
	}
	if err := windows.NtQueryInformationProcess(parent, windows.ProcessBasicInformation, unsafe.Pointer(&ancestor), uint32(unsafe.Sizeof(ancestor)), nil); err != nil {
		return 0, err
	}
	if ancestor.UniqueProcessId == 0 || ancestor.UniqueProcessId != own.InheritedFromUniqueProcessId {
		return 0, ErrProtocol
	}
	state, err := windows.WaitForSingleObject(parent, 0)
	if err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		return 0, ErrProtocol
	}
	return uint64(ancestor.UniqueProcessId), nil
}

// RunWindowsPrivate is called by the tiny private build entry or Composition's
// early dispatch. It never reads configuration, resolves a provider, or executes
// any user specification before inherited capability and Start authentication.
// A nonzero result means the parent must enforce the retained outer Job barrier.
func RunWindowsPrivate() int {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	input, output := os.Stdin, os.Stdout
	parent := windows.Handle(os.Stderr.Fd())
	defer input.Close()
	defer output.Close()
	defer windows.CloseHandle(parent)
	parentID, err := validateWindowsEndpoints(input, output, parent)
	if err != nil {
		return 70
	}
	var channel *Channel
	var first Frame
	err = timedIO(input, 30*time.Second, func() error {
		var e error
		channel, first, e = AcceptChannel(input, output, WindowsBroker, Parent)
		return e
	})
	if err != nil || first.Opcode != Start {
		return 71
	}
	spec, err := DecodeStart(first.Payload)
	if err != nil || spec.ParentID != parentID {
		return 72
	}
	if _, embedded, err := MachineRoute(); err != nil || embedded {
		return 73
	}
	return runWindowsEngine(channel, input, output, parent, spec)
}

func runWindowsEngine(channel *Channel, input, output *os.File, parent windows.Handle, spec StartSpec) int {
	return runWindowsEngineOwned(channel, input, output, parent, spec, &userProcess{})
}

func runWindowsEngineOwned(channel *Channel, input, output *os.File, parent windows.Handle, spec StartSpec, p *userProcess) int {
	startup, cancelStartup := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelStartup()
	incoming := make(chan receivedFrame, 1)
	receiverDone := make(chan struct{})
	receiverStop := make(chan struct{})
	go func() {
		defer close(receiverDone)
		for {
			f, err := channel.Receive()
			if err != nil || f.Opcode == Abort || f.Opcode == Stop {
				cancelStartup()
			}
			select {
			case incoming <- receivedFrame{f, err}:
			case <-receiverStop:
				return
			}
			if err != nil {
				return
			}
		}
	}()
	defer func() { close(receiverStop); _ = input.Close(); <-receiverDone }()
	// The engine exits only after joined local cleanup or a failed result which
	// delegates residual containment to the still-retained parent outer Job.
	defer func() {
		_ = boundedCleanup(p)
		for i := range p.outputs {
			_ = closeFile(&p.outputs[i].file)
		}
	}()
	stage := api.Acquisition
	var joinInput func() error
	fail := func(e error) int {
		// Preserve the original source failure and actual independent cleanup
		// failures in one bounded frame. The outer Job remains the parent's
		// authority if local cleanup cannot complete within its reporting bound.
		e = windowsFailureAt(e, stage)
		if joinInput != nil {
			e = errors.Join(e, joinInput())
		}
		e = errors.Join(e, boundedCleanup(p))
		for i := range p.outputs {
			e = errors.Join(e, windowsCleanupFailure(closeFile(&p.outputs[i].file), api.OutputCleanup))
		}
		payload, encodeErr := encodeWindowsFailure(e)
		if encodeErr != nil {
			payload, _ = encodeWindowsFailure(windowsFailureAt(ErrProtocol, api.ControlCleanup))
		}
		_ = sendControl(channel, output, Failure, payload)
		return 74
	}
	if err := p.prepare(spec); err != nil {
		return fail(err)
	}
	for i := range p.outputs {
		stage = api.OutputCleanup
		out := &p.outputs[i]
		// The parent duplicates from its retained exact broker process handle.
		// This avoids allocating an unknown remote handle if a transfer frame
		// fails before reaching the parent. Neither side reads before transfer.
		payload := []byte{out.stream}
		payload = binary.BigEndian.AppendUint64(payload, uint64(out.file.Fd()))
		if err := sendControl(channel, output, OutputTransfer, payload); err != nil {
			return fail(windowsFailureAt(err, api.ControlCleanup))
		}
		select {
		case got := <-incoming:
			if got.err != nil || got.frame.Opcode != OutputAccepted || len(got.frame.Payload) != 1 || got.frame.Payload[0] != out.stream {
				return fail(ErrProtocol)
			}
		case <-startup.Done():
			return fail(startup.Err())
		}
		if err := closeFile(&out.file); err != nil {
			return fail(err)
		}
	}
	stage = api.ProcessContainment
	if err := p.start(startup, spec); err != nil {
		return fail(err)
	}
	architecture := p.debug.architecture
	started := binary.BigEndian.AppendUint16(nil, architecture.NativeMachine)
	started = binary.BigEndian.AppendUint16(started, architecture.ProcessMachine)
	started = binary.BigEndian.AppendUint16(started, architecture.ImageMachine)
	started = binary.BigEndian.AppendUint32(started, architecture.InitialBreakpoint)
	stage = api.ControlCleanup
	if err := sendControl(channel, output, Started, started); err != nil {
		return fail(err)
	}
	// One bounded input operation may be in flight. Receive and Stop remain
	// independent while the child declines input; closing input joins it.
	type writeResult struct {
		sequence            uint64
		accepted, delivered uint32
		err                 error
	}
	writes := make(chan writeResult, 1)
	reportWrite := func(result writeResult) error {
		body := binary.BigEndian.AppendUint64(nil, result.sequence)
		body = binary.BigEndian.AppendUint32(body, result.accepted)
		body = binary.BigEndian.AppendUint32(body, result.delivered)
		if result.err != nil {
			body = append(body, 1)
		} else {
			body = append(body, 0)
		}
		return sendControl(channel, output, Delivered, body)
	}
	var writing bool
	defer func() {
		_ = closeFile(&p.input)
		if writing {
			<-writes
		}
	}()
	write := func(f Frame, data []byte) {
		writing = true
		owned := append([]byte(nil), data...)
		file := p.input
		go func() {
			n, err := file.Write(owned)
			if err == nil && n != len(owned) {
				err = io.ErrShortWrite
			}
			writes <- writeResult{f.Sequence, uint32(len(owned)), uint32(n), err}
		}()
	}
	joinInput = func() error {
		err := windowsCleanupFailure(closeFile(&p.input), api.InputCleanup)
		if writing {
			result := <-writes
			writing = false
			// Even a later native failure cannot erase a known partial write.
			// The parent receives this before the terminal Failure when the
			// authenticated control direction still permits delivery.
			err = errors.Join(err, windowsFailureAt(reportWrite(result), api.ControlCleanup))
		}
		return err
	}
	stopping := false
	controlFailed := false
	gracePeriod, forcePeriod := 2*time.Second, 3*time.Second
	var stopAfter time.Time
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		state, err := windows.WaitForSingleObject(p.debug.process.Process, 0)
		if err != nil {
			return fail(windowsFailureAt(err, api.UserProcessWait))
		}
		if state == windows.WAIT_OBJECT_0 {
			stopping = true
			stopAfter = time.Time{}
		}
		if stopping && (stopAfter.IsZero() || !time.Now().Before(stopAfter)) {
			break
		}
		select {
		case got := <-incoming:
			if got.err != nil {
				controlFailed = true
				stopping = true
				stopAfter = time.Time{}
				continue
			}
			f := got.frame
			switch f.Opcode {
			case Stop, Abort:
				if len(f.Payload) != 0 && (f.Opcode != Stop || len(f.Payload) != 8) {
					return fail(ErrProtocol)
				}
				if len(f.Payload) == 8 {
					grace, force := binary.BigEndian.Uint32(f.Payload), binary.BigEndian.Uint32(f.Payload[4:])
					if grace == 0 || force == 0 || grace > 60000 || force > 60000 {
						return fail(ErrProtocol)
					}
					gracePeriod, forcePeriod = time.Duration(grace)*time.Millisecond, time.Duration(force)*time.Millisecond
				}
				if !stopping {
					stopping = true
					if f.Opcode == Stop && spec.Terminal && !writing {
						write(f, []byte{3})
						stopAfter = time.Now().Add(gracePeriod)
					}
				}
				if f.Opcode == Abort {
					stopAfter = time.Time{}
				}
			case WriteInput, Interrupt:
				data := f.Payload
				if f.Opcode == Interrupt {
					if len(data) != 0 || !spec.Terminal {
						return fail(ErrProtocol)
					}
					data = []byte{3}
				}
				if len(data) == 0 {
					return fail(ErrProtocol)
				}
				if stopping || writing {
					return fail(windowsFailureAt(ErrWindowsBusy, api.InputCleanup))
				}
				write(f, data)
			case Resize:
				if len(f.Payload) != 4 || !spec.Terminal {
					return fail(ErrProtocol)
				}
				rows, columns := binary.BigEndian.Uint16(f.Payload), binary.BigEndian.Uint16(f.Payload[2:])
				if rows == 0 || columns == 0 || rows > 32767 || columns > 32767 {
					return fail(ErrProtocol)
				}
				if stopping {
					return fail(ErrProtocol)
				}
				if err = windows.ResizePseudoConsole(p.hpc, windows.Coord{X: int16(columns), Y: int16(rows)}); err != nil {
					return fail(windowsFailureAt(err, api.TerminalCleanup))
				}
				if err = sendControl(channel, output, Delivered, binary.BigEndian.AppendUint64(nil, f.Sequence)); err != nil {
					return fail(err)
				}
			default:
				return fail(ErrProtocol)
			}
		case result := <-writes:
			writing = false
			if err = reportWrite(result); err != nil {
				return fail(err)
			}
		case <-ticker.C:
		}
	}
	if err := closeFile(&p.input); err != nil {
		return fail(windowsCleanupFailure(err, api.InputCleanup))
	}
	if writing {
		result := <-writes
		writing = false
		if err := reportWrite(result); err != nil {
			return fail(err)
		}
	}
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), forcePeriod)
	cleanupErr := p.cleanup(cleanupCtx)
	cleanupCancel()
	if cleanupErr != nil {
		return fail(cleanupErr)
	}
	exit := binary.BigEndian.AppendUint32(nil, p.exit)
	if err := sendControl(channel, output, UserExit, exit); err != nil {
		return 0
	}
	if controlFailed {
		return fail(ErrProtocol)
	}
	if err := sendControl(channel, output, Quiescent, nil); err != nil {
		return 0
	}
	// Terminal auxiliaries may remain in the outer Job here. Release must not
	// depend on their disappearance; parent cleans outer membership after exit.
	for {
		select {
		case got := <-incoming:
			if got.err != nil {
				return 0
			}
			if got.frame.Opcode == Release && len(got.frame.Payload) == 0 {
				return 0
			}
			if got.frame.Opcode == Stop || got.frame.Opcode == Abort {
				continue
			}
			return 75
		case <-ticker.C:
			state, err := windows.WaitForSingleObject(parent, 0)
			if err != nil || state == windows.WAIT_OBJECT_0 {
				return 0
			}
		}
	}
}
