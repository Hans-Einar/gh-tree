//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

// UnixConfig is private Runtime plumbing. Output must copy into the bounded
// parent ring and return without waiting for a UI consumer. Native helpers always
// execute the actual current image, never a configured/PATH-resolved helper.
type UnixConfig struct {
	SessionID                uint64
	Spec                     StartSpec
	Output                   func(api.OutputStream, []byte)
	GracePeriod, ForcePeriod time.Duration
	hook                     func(string, *UnixClient) error
}

type UnixStartResult struct {
	Established bool
	Cwd         string
}
type UnixResidual struct {
	Stage api.RuntimeCleanupStage
	Err   error
}

// UnixFailure transports a closed safe error class and cleanup stage without
// copying argv, environment, private endpoints or native process numbers.
type UnixFailure struct {
	Code  api.ErrorCode
	Stage api.RuntimeCleanupStage
}

func (e UnixFailure) Error() string {
	return fmt.Sprintf("Unix native failure (code %d, stage %d)", e.Code, e.Stage)
}

type UnixFact struct {
	Established, RootExited    bool
	ExitCode, Signal           int
	Quiescent, CleanupComplete bool
	Err                        error
	Residuals                  []UnixResidual
}
type UnixDelivery struct {
	Accepted, Delivered uint32
	Completed           bool
}

type unixFile struct {
	file   *os.File
	stage  api.RuntimeCleanupStage
	once   sync.Once
	err    error
	closed atomic.Bool
}

func (f *unixFile) close() error {
	if f == nil {
		return nil
	}
	f.once.Do(func() { f.err = f.file.Close(); f.closed.Store(true) })
	return f.err
}

var errUnixStartup = errors.New("Unix private startup failed before verified establishment")
var errUnixCleanup = errors.New("Unix resource cleanup is incomplete")
var errUnixObservation = errors.New("Unix native ownership observation failed")
var errUnixJoinTimeout = errors.New("Unix output owner exceeded its join reporting period")

// UnixClient owns the exact supervisor, retained cwd and every parent pipe/PTY
// and I/O worker. It is one native resource owner, never a SessionID registry.
// Facts are cumulative immutable copies: repeated intermediate failures may
// coalesce, while Wait always preserves the latest complete barrier evidence.
type UnixClient struct {
	config                          UnixConfig
	mu                              sync.Mutex
	latest                          UnixFact
	startup                         UnixStartResult
	startDone                       chan struct{}
	startOnce                       sync.Once
	pending                         bool
	changed                         chan struct{}
	done                            chan struct{}
	stop                            chan struct{}
	stopOnce                        sync.Once
	stopping                        bool
	diagnostics                     map[api.RuntimeCleanupStage]error
	residuals                       map[api.RuntimeCleanupStage]error
	reported                        map[api.RuntimeCleanupStage]error
	files                           []*unixFile
	control, reply, input, terminal *unixFile
	childFiles                      []*unixFile
	outputs                         []*unixFile
	acquired                        *AcquiredDirectory
	cmd                             *exec.Cmd
	processDone                     chan struct{}
	processErr                      error
	channel                         *Channel
	mayHaveStarted                  bool
	outputWG                        sync.WaitGroup
	ioWG                            sync.WaitGroup
	activeIO                        atomic.Int32
	activeCallbacks                 atomic.Int32
	readerActive                    atomic.Bool
	cwdClosed                       atomic.Bool
	outputDone                      chan struct{}
	writeGate                       chan struct{}
	setupErr                        error
}

func StartUnix(ctx context.Context, config UnixConfig) (*UnixClient, UnixStartResult, error) {
	if ctx == nil || config.SessionID == 0 || !config.Spec.valid() || config.Output == nil || config.Spec.RootIdentity.Platform() != api.DirectoryUnix {
		return nil, UnixStartResult{}, ErrProtocol
	}
	if err := ctx.Err(); err != nil {
		return nil, UnixStartResult{}, err
	}
	grace, force, err := effectiveUnixPeriods(config.GracePeriod, config.ForcePeriod)
	if err != nil {
		return nil, UnixStartResult{}, ErrProtocol
	}
	config.GracePeriod, config.ForcePeriod = grace, force
	config.Spec.ParentID = uint64(os.Getpid())
	config.Spec.Components = append([]string(nil), config.Spec.Components...)
	config.Spec.Arguments = append([]string(nil), config.Spec.Arguments...)
	config.Spec.Environment = append([]string(nil), config.Spec.Environment...)
	c := &UnixClient{config: config, startDone: make(chan struct{}), changed: make(chan struct{}), done: make(chan struct{}), stop: make(chan struct{}), outputDone: make(chan struct{}), writeGate: make(chan struct{}, 1), diagnostics: make(map[api.RuntimeCleanupStage]error), residuals: make(map[api.RuntimeCleanupStage]error), reported: make(map[api.RuntimeCleanupStage]error)}
	c.writeGate <- struct{}{}
	c.setupErr = c.create(ctx)
	go c.run()
	select {
	case <-c.startDone:
		c.mu.Lock()
		result, err := c.startup, c.latest.Err
		c.mu.Unlock()
		if result.Established {
			return c, result, c.setupErr
		}
		if err == nil {
			err = c.setupErr
		}
		if err == nil {
			err = errUnixStartup
		}
		return c, result, err
	case <-ctx.Done():
		c.mu.Lock()
		result := c.startup
		c.mu.Unlock()
		if result.Established {
			return c, result, c.setupErr
		}
		c.Stop()
		return c, result, ctx.Err()
	}
}

func (c *UnixClient) own(file *os.File, stage api.RuntimeCleanupStage) *unixFile {
	f := &unixFile{file: file, stage: stage}
	c.files = append(c.files, f)
	return f
}
func (c *UnixClient) pipe(stage api.RuntimeCleanupStage) (*unixFile, *unixFile, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return c.own(r, stage), c.own(w, stage), nil
}
func (c *UnixClient) checkpoint(stage string) error {
	if c.config.hook != nil {
		return c.config.hook(stage, c)
	}
	return nil
}

func (c *UnixClient) create(ctx context.Context) error {
	a, err := AcquireCwd(c.config.Spec)
	c.acquired = a
	if err != nil {
		return err
	}
	if err := c.checkpoint("cwd"); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	controlRead, controlWrite, err := c.pipe(api.ControlCleanup)
	if err != nil {
		return err
	}
	c.control = controlWrite
	if err := c.checkpoint("control-request-pipe"); err != nil {
		return err
	}
	replyRead, replyWrite, err := c.pipe(api.ControlCleanup)
	if err != nil {
		return err
	}
	c.reply = replyRead
	c.childFiles = append(c.childFiles, controlRead, replyWrite)
	if err := c.checkpoint("control-pipes"); err != nil {
		return err
	}
	var stdin, stdout, stderr *unixFile
	if c.config.Spec.Terminal {
		master, slave, err := pty.Open()
		if err != nil {
			return err
		}
		original := c.own(master, api.TerminalCleanup)
		slaveFile := c.own(slave, api.TerminalCleanup)
		if err := c.checkpoint("pty-pair"); err != nil {
			return err
		}
		pollable, pollErr := pollablePTY(master)
		if pollable != nil {
			c.terminal = c.own(pollable, api.TerminalCleanup)
		}
		if err := errors.Join(pollErr, original.close()); err != nil {
			return err
		}
		if err := c.checkpoint("pty-master"); err != nil {
			return err
		}
		c.input = c.terminal
		c.outputs = []*unixFile{c.terminal}
		stdin, stdout, stderr = slaveFile, slaveFile, slaveFile
		c.childFiles = append(c.childFiles, slaveFile)
		if err := pty.Setsize(c.terminal.file, &pty.Winsize{Rows: c.config.Spec.Rows, Cols: c.config.Spec.Columns}); err != nil {
			return err
		}
	} else {
		inputRead, inputWrite, err := c.pipe(api.InputCleanup)
		if err != nil {
			return err
		}
		c.input = inputWrite
		stdin = inputRead
		c.childFiles = append(c.childFiles, inputRead)
		if err := c.checkpoint("input-pipe"); err != nil {
			return err
		}
		outRead, outWrite, err := c.pipe(api.OutputCleanup)
		if err != nil {
			return err
		}
		stdout = outWrite
		c.outputs = append(c.outputs, outRead)
		c.childFiles = append(c.childFiles, outWrite)
		if err := c.checkpoint("stdout-pipe"); err != nil {
			return err
		}
		errRead, errWrite, err := c.pipe(api.OutputCleanup)
		if err != nil {
			return err
		}
		stderr = errWrite
		c.outputs = append(c.outputs, errRead)
		c.childFiles = append(c.childFiles, errWrite)
	}
	if err := c.checkpoint("io"); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil || !filepath.IsAbs(executable) {
		return errors.Join(err, ErrProtocol)
	}
	cmd := exec.Command(executable, supervisorPrivateMarker)
	cmd.Stdin = stdin.file
	cmd.Stdout = stdout.file
	cmd.Stderr = stderr.file
	cmd.ExtraFiles = []*os.File{controlRead.file, replyWrite.file, a.File()}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if c.config.Spec.Terminal {
		cmd.SysProcAttr.Setctty = true
		cmd.SysProcAttr.Ctty = 0
	}
	if err := a.Revalidate(); err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	c.cmd = cmd
	c.processDone = make(chan struct{})
	go func() { c.processErr = cmd.Wait(); close(c.processDone) }()
	for index, file := range c.childFiles {
		if err := file.close(); err != nil {
			return err
		}
		if err := c.checkpoint(fmt.Sprintf("child-endpoint-%d", index)); err != nil {
			return err
		}
	}
	for index, file := range c.outputs {
		stream := api.Stdout
		if index == 1 {
			stream = api.Stderr
		}
		if c.config.Spec.Terminal {
			stream = api.TerminalOutput
		}
		c.outputWG.Add(1)
		go c.drain(file, stream)
		if err := c.checkpoint(fmt.Sprintf("output-reader-%d", index)); err != nil {
			return err
		}
	}
	if err := c.checkpoint("supervisor-created"); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	nonce, err := FreshNonce()
	if err != nil {
		return err
	}
	c.channel, err = NewChannel(replyRead.file, controlWrite.file, Parent, UnixSupervisor, c.config.SessionID, nonce)
	if err != nil {
		return err
	}
	payload, err := encodeUnixStart(c.config.Spec, c.config.GracePeriod, c.config.ForcePeriod)
	if err != nil {
		return err
	}
	if err := controlWrite.file.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	c.mayHaveStarted = true // a partial send is not proof that no complete Start crossed
	if err := c.channel.Send(Start, payload); err != nil {
		return err
	}
	return c.checkpoint("start-sent")
}

// pty's OpenFile/Fd ioctl path can switch its Go file to blocking mode. A
// separately owned duplicate is made nonblocking before os.NewFile, so later
// Fd-based resize ioctls retain the poller/deadline/Close behavior. On partial
// setup the returned file still owns the duplicate and must be closed.
func pollablePTY(original *os.File) (*os.File, error) {
	fd, err := unix.FcntlInt(original.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	err = unix.SetNonblock(fd, true)
	return os.NewFile(uintptr(fd), "runtime-pty-master"), err
}

func (c *UnixClient) record(err error, stage api.RuntimeCleanupStage, residual bool) {
	if err == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !errors.Is(c.diagnostics[stage], err) {
		c.diagnostics[stage] = errors.Join(c.diagnostics[stage], err)
	}
	if residual {
		c.residuals[stage] = err
	}
}
func (c *UnixClient) factLocked() UnixFact {
	f := c.latest
	f.Residuals = nil
	f.Err = nil
	pending := make(map[api.RuntimeCleanupStage]error)
	if !f.CleanupComplete {
		for _, file := range c.files {
			if !file.closed.Load() {
				pending[file.stage] = errUnixCleanup
			} else if file.err != nil {
				pending[file.stage] = file.err
			}
		}
		if c.acquired != nil && !c.cwdClosed.Load() {
			pending[api.CwdAcquisition] = errUnixCleanup
		}
		if c.cmd != nil {
			select {
			case <-c.processDone:
			default:
				pending[api.SupervisorOrBroker] = errUnixCleanup
			}
		}
		if f.Established && !f.RootExited {
			pending[api.UserProcessWait] = errUnixCleanup
		}
		if c.mayHaveStarted && !f.Quiescent {
			pending[api.Descendants] = errUnixCleanup
		}
		if c.activeIO.Load() != 0 {
			pending[api.InputCleanup] = errUnixCleanup
		}
		if c.readerActive.Load() {
			pending[api.ControlCleanup] = errUnixCleanup
		}
		select {
		case <-c.outputDone:
		default:
			pending[api.OutputCleanup] = errUnixCleanup
		}
	}
	for stage := api.Acquisition; stage <= api.EventTransfer; stage++ {
		f.Err = errors.Join(f.Err, c.diagnostics[stage])
		if err := errors.Join(c.residuals[stage], c.reported[stage], pending[stage]); err != nil {
			f.Residuals = append(f.Residuals, UnixResidual{stage, err})
		}
	}
	return f
}
func (c *UnixClient) publish() {
	c.mu.Lock()
	c.latest = c.factLocked()
	c.pending = true
	close(c.changed)
	c.changed = make(chan struct{})
	c.mu.Unlock()
}
func (c *UnixClient) finishStart() { c.startOnce.Do(func() { close(c.startDone) }) }

func (c *UnixClient) drain(file *unixFile, stream api.OutputStream) {
	defer c.outputWG.Done()
	defer func() {
		if recover() != nil {
			c.record(errors.New("Unix output owner callback panicked"), api.OutputCleanup, false)
			c.Stop()
		}
	}()
	buffer := make([]byte, 32<<10)
	for {
		n, err := file.file.Read(buffer)
		if n > 0 {
			c.deliverOutput(stream, buffer[:n])
		}
		if err != nil {
			if err != io.EOF && !(c.config.Spec.Terminal && errors.Is(err, syscall.EIO)) && !errors.Is(err, os.ErrClosed) {
				c.record(err, api.OutputCleanup, false)
			}
			return
		}
	}
}

func (c *UnixClient) send(op Opcode, payload []byte) error {
	if c.channel == nil || c.control == nil {
		return ErrProtocol
	}
	if err := c.control.file.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	return c.channel.Send(op, payload)
}

func (c *UnixClient) deliverOutput(stream api.OutputStream, data []byte) {
	c.activeCallbacks.Add(1)
	defer c.activeCallbacks.Add(-1)
	c.config.Output(stream, append([]byte(nil), data...))
}

func (c *UnixClient) run() {
	defer close(c.done)
	go func() { c.outputWG.Wait(); close(c.outputDone) }()
	if c.setupErr != nil {
		c.record(c.setupErr, api.Acquisition, false)
		c.publish()
		c.finishStart()
	}
	if !c.mayHaveStarted {
		// The Start frame never began. No user specification reached this
		// supervisor, so closing its control is a definitive pre-user abort.
		if c.control != nil {
			c.record(c.control.close(), api.ControlCleanup, true)
		}
		c.joinAndClose(false)
		c.mu.Lock()
		c.latest.Quiescent = true
		c.latest.CleanupComplete = len(c.residuals) == 0
		c.mu.Unlock()
		c.publish()
		c.finishStart()
		return
	}
	frames := make(chan parentFrame, 1)
	readerDone := make(chan struct{})
	readerStop := make(chan struct{})
	c.readerActive.Store(true)
	go func() {
		defer close(readerDone)
		for {
			frame, err := c.channel.Receive()
			select {
			case frames <- parentFrame{frame, err}:
			case <-readerStop:
				return
			}
			if err != nil {
				return
			}
		}
	}()
	stop := c.stop
	if c.setupErr != nil {
		c.Stop()
	}
	quiet := false
loop:
	for {
		select {
		case <-stop:
			stop = nil
			if err := c.send(Stop, nil); err != nil {
				c.record(err, api.ControlCleanup, true)
				c.record(c.control.close(), api.ControlCleanup, true)
			}
		case event := <-frames:
			if event.err != nil {
				c.record(event.err, api.ControlCleanup, false)
				break loop
			}
			switch event.frame.Opcode {
			case Started:
				d := decoder{buf: event.frame.Payload}
				pid := d.u64()
				cwd := d.string()
				if d.err != nil || d.pos != len(d.buf) || pid == 0 || cwd == "" {
					c.record(ErrProtocol, api.ControlCleanup, true)
					break loop
				}
				c.mu.Lock()
				if c.latest.Established {
					c.mu.Unlock()
					c.record(ErrProtocol, api.ControlCleanup, true)
					break loop
				}
				c.latest.Established = true
				c.startup = UnixStartResult{true, cwd}
				c.mu.Unlock()
				if c.acquired != nil {
					c.record(c.acquired.Close(), api.CwdAcquisition, true)
					c.cwdClosed.Store(true)
				}
				c.publish()
				c.finishStart()
			case UserExit:
				if len(event.frame.Payload) != 8 {
					c.record(ErrProtocol, api.ControlCleanup, true)
					break loop
				}
				c.mu.Lock()
				c.latest.RootExited = true
				c.latest.ExitCode = int(int32(binary.BigEndian.Uint32(event.frame.Payload)))
				c.latest.Signal = int(binary.BigEndian.Uint32(event.frame.Payload[4:]))
				c.mu.Unlock()
				c.publish()
			case Failure:
				if len(event.frame.Payload) != 1 && len(event.frame.Payload) != 3 {
					c.record(ErrProtocol, api.ControlCleanup, true)
					break loop
				}
				stage := api.Descendants
				err := errUnixCleanup
				if event.frame.Payload[0] == 1 {
					stage = api.Acquisition
					err = errUnixStartup
				}
				if event.frame.Payload[0] == 3 {
					err = errUnixObservation
				}
				if len(event.frame.Payload) == 3 {
					stage = api.RuntimeCleanupStage(event.frame.Payload[1])
					code := api.ErrorCode(event.frame.Payload[2])
					if !stage.Valid() || !code.Valid() {
						c.record(ErrProtocol, api.ControlCleanup, true)
						break loop
					}
					err = UnixFailure{code, stage}
				}
				c.record(err, stage, false)
				c.mu.Lock()
				c.reported[stage] = err
				c.mu.Unlock()
				c.publish()
				c.finishStart()
			case Quiescent:
				if len(event.frame.Payload) != 1 {
					c.record(ErrProtocol, api.ControlCleanup, true)
					break loop
				}
				quiet = true
				c.mu.Lock()
				c.latest.Quiescent = true
				c.mu.Unlock()
				if event.frame.Payload[0] != 0 {
					c.record(errUnixCleanup, api.Descendants, false)
				}
				if err := c.send(Release, nil); err != nil {
					c.record(err, api.ControlCleanup, false)
					c.record(c.control.close(), api.ControlCleanup, true)
				}
				break loop
			default:
				c.record(ErrProtocol, api.ControlCleanup, true)
				break loop
			}
		}
	}
	if !quiet {
		c.record(errUnixCleanup, api.Descendants, true)
		c.record(errUnixCleanup, api.SupervisorOrBroker, true)
		c.publish()
		c.finishStart()
		c.record(c.control.close(), api.ControlCleanup, true)
	}
	close(readerStop)
	c.record(c.reply.close(), api.ControlCleanup, true)
	<-readerDone
	c.readerActive.Store(false)
	c.joinAndClose(quiet)
	c.mu.Lock()
	if quiet {
		// A full native quiescence proof repairs earlier observation/control
		// uncertainty, but an actual retained close/join failure remains.
		clear(c.reported)
	}
	c.latest.CleanupComplete = quiet && len(c.residuals) == 0 && len(c.reported) == 0
	c.mu.Unlock()
	c.publish()
	c.finishStart()
}

func (c *UnixClient) joinAndClose(quiet bool) {
	c.Stop()
	for _, file := range c.childFiles {
		c.record(file.close(), file.stage, true)
	}
	if c.cmd != nil {
		<-c.processDone // exact waiter remains owned even if public Wait times out
		if c.processErr != nil && c.mayHaveStarted {
			var exited *exec.ExitError
			c.record(c.processErr, api.SupervisorOrBroker, !quiet || !errors.As(c.processErr, &exited))
		}
	}
	select {
	case <-c.outputDone:
	case <-time.After(c.config.ForcePeriod):
		blockedCallback := c.activeCallbacks.Load() > 0
		c.record(errUnixJoinTimeout, api.OutputCleanup, true)
		for _, file := range c.outputs {
			c.record(file.close(), file.stage, true)
		}
		<-c.outputDone // callbacks/readers remain owned until they actually join
		if blockedCallback {
			for _, file := range c.outputs {
				if file.err != nil {
					blockedCallback = false
				}
			}
		}
		if blockedCallback {
			c.mu.Lock()
			delete(c.residuals, api.OutputCleanup)
			c.mu.Unlock()
		}
	}
	for _, file := range c.files {
		c.record(file.close(), file.stage, true)
	}
	c.ioWG.Wait()
	if c.acquired != nil {
		c.record(c.acquired.Close(), api.CwdAcquisition, true)
		c.cwdClosed.Store(true)
	}
}

func (c *UnixClient) Stop() {
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.stopping = true
		c.mu.Unlock()
		close(c.stop)
		if c.input != nil {
			err := c.input.file.SetWriteDeadline(time.Now())
			if !errors.Is(err, os.ErrClosed) {
				c.record(err, api.InputCleanup, false)
			}
		}
	})
}

func (c *UnixClient) Write(ctx context.Context, data []byte) (UnixDelivery, error) {
	if ctx == nil || len(data) == 0 || len(data) > 65536 {
		return UnixDelivery{}, ErrProtocol
	}
	select {
	case <-ctx.Done():
		return UnixDelivery{}, ctx.Err()
	case <-c.writeGate:
	}
	defer func() { c.writeGate <- struct{}{} }()
	c.mu.Lock()
	allowed := c.latest.Established && !c.stopping && !c.latest.Quiescent
	if allowed {
		c.ioWG.Add(1)
		c.activeIO.Add(1)
	}
	c.mu.Unlock()
	if !allowed || c.input == nil {
		return UnixDelivery{}, ErrProtocol
	}
	defer func() { c.activeIO.Add(-1); c.ioWG.Done() }()
	if err := ctx.Err(); err != nil {
		return UnixDelivery{}, err
	}
	deadline := time.Time{}
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	}
	if err := c.input.file.SetWriteDeadline(deadline); err != nil {
		return UnixDelivery{}, err
	}
	select {
	case <-c.stop:
		return UnixDelivery{}, ErrProtocol
	default:
	}
	copyData := append([]byte(nil), data...)
	finished := make(chan struct{})
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			err := c.input.file.SetWriteDeadline(time.Now())
			if !errors.Is(err, os.ErrClosed) {
				c.record(err, api.InputCleanup, false)
			}
		case <-finished:
		}
	}()
	n, err := c.input.file.Write(copyData)
	close(finished)
	<-watcherDone
	cancelErr := ctx.Err()
	if cancelErr == nil && errors.Is(err, os.ErrDeadlineExceeded) && !deadline.IsZero() && !time.Now().Before(deadline) {
		cancelErr = context.DeadlineExceeded
	}
	return UnixDelivery{Accepted: uint32(len(copyData)), Delivered: uint32(n), Completed: err == nil && n == len(copyData)}, errors.Join(err, cancelErr)
}

func (c *UnixClient) Resize(ctx context.Context, rows, columns uint16) (UnixDelivery, error) {
	if ctx == nil || !c.config.Spec.Terminal || rows == 0 || columns == 0 || rows > 32767 || columns > 32767 {
		return UnixDelivery{}, ErrProtocol
	}
	if err := ctx.Err(); err != nil {
		return UnixDelivery{}, err
	}
	c.mu.Lock()
	allowed := c.latest.Established && !c.stopping && !c.latest.Quiescent
	if allowed {
		c.ioWG.Add(1)
		c.activeIO.Add(1)
	}
	c.mu.Unlock()
	if !allowed {
		return UnixDelivery{}, ErrProtocol
	}
	defer func() { c.activeIO.Add(-1); c.ioWG.Done() }()
	err := pty.Setsize(c.terminal.file, &pty.Winsize{Rows: rows, Cols: columns})
	return UnixDelivery{Completed: err == nil}, err
}
func (c *UnixClient) Interrupt(ctx context.Context) (UnixDelivery, error) {
	if !c.config.Spec.Terminal {
		return UnixDelivery{}, ErrProtocol
	}
	return c.Write(ctx, []byte{3})
}

func (c *UnixClient) NextFact(ctx context.Context) (UnixFact, error) {
	if ctx == nil {
		return UnixFact{}, ErrProtocol
	}
	for {
		c.mu.Lock()
		if err := ctx.Err(); err != nil {
			c.mu.Unlock()
			return UnixFact{}, err
		}
		if c.pending {
			f := c.factLocked()
			c.pending = false
			c.mu.Unlock()
			return f, nil
		}
		changed := c.changed
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return UnixFact{}, ctx.Err()
		case <-changed:
		case <-c.done:
			c.mu.Lock()
			pending := c.pending
			c.mu.Unlock()
			if !pending {
				return UnixFact{}, io.EOF
			}
		}
	}
}
func (c *UnixClient) Wait(ctx context.Context) (UnixFact, error) {
	if ctx == nil {
		return UnixFact{}, ErrProtocol
	}
	select {
	case <-c.done:
		c.mu.Lock()
		f := c.factLocked()
		c.mu.Unlock()
		return f, f.Err
	case <-ctx.Done():
		c.mu.Lock()
		f := c.factLocked()
		c.mu.Unlock()
		return f, errors.Join(f.Err, ctx.Err())
	}
}

// RunUnixPrivate is the bounded early dispatch seam for Composition. A fixed
// marker with invalid/missing inherited endpoints fails without normal bootstrap.
func RunUnixPrivate() (bool, int) {
	if len(os.Args) != 2 {
		return false, 0
	}
	switch os.Args[1] {
	case supervisorPrivateMarker:
		return true, RunSupervisor()
	case signalPrivateMarker:
		return true, RunSignalHelper()
	default:
		return false, 0
	}
}
