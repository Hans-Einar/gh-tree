package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

// WindowsConfig is private to the Runtime layer. Image is the trusted current
// executable or an already verified, retained extracted image. Output must copy
// into the registry's bounded ring and return without waiting for a UI consumer.
type WindowsConfig struct {
	SessionID   uint64
	Spec        StartSpec
	Image       string
	Output      func(api.OutputStream, []byte)
	GracePeriod time.Duration
	ForcePeriod time.Duration
}

type WindowsStartResult struct {
	Established bool
	Cwd         string
}
type WindowsFact struct {
	Established     bool
	RootExited      bool
	ExitCode        uint32
	Quiescent       bool
	CleanupComplete bool
	Err             error
}
type WindowsDelivery struct {
	Accepted, Delivered uint32
	Completed           bool
}

type controlReply struct {
	delivery WindowsDelivery
	err      error
}

// WindowsClient owns one exact broker process, its outer Job, and all parent
// transport workers. There is no SessionID allocator or second session registry.
type WindowsClient struct {
	config       WindowsConfig
	job          windows.Handle
	process      windows.ProcessInformation
	read, write  *os.File
	childHandles []windows.Handle
	channel      *Channel
	controlMu    sync.Mutex
	sequence     uint64
	pendingMu    sync.Mutex
	pending      map[uint64]chan controlReply
	outputFiles  []*os.File
	outputDone   sync.WaitGroup
	outputMu     sync.Mutex
	outputErr    error
	facts        chan WindowsFact
	started      chan WindowsStartResult
	stop         chan struct{}
	stopOnce     sync.Once
	done         chan struct{}
	mu           sync.Mutex
	latest       WindowsFact
	setupErr     error
}

func StartWindows(ctx context.Context, config WindowsConfig) (*WindowsClient, WindowsStartResult, error) {
	if ctx == nil || config.SessionID == 0 || config.Image == "" || config.Output == nil || !config.Spec.valid() {
		return nil, WindowsStartResult{}, ErrProtocol
	}
	if err := ctx.Err(); err != nil {
		return nil, WindowsStartResult{}, err
	}
	config.Spec.ParentID = uint64(os.Getpid())
	if config.GracePeriod == 0 {
		config.GracePeriod = 2 * time.Second
	}
	if config.ForcePeriod == 0 {
		config.ForcePeriod = 3 * time.Second
	}
	if config.GracePeriod < time.Millisecond || config.ForcePeriod < time.Millisecond || config.GracePeriod > time.Minute || config.ForcePeriod > time.Minute {
		return nil, WindowsStartResult{}, ErrProtocol
	}
	config.Spec.Components = append([]string(nil), config.Spec.Components...)
	config.Spec.Arguments = append([]string(nil), config.Spec.Arguments...)
	config.Spec.Environment = append([]string(nil), config.Spec.Environment...)
	c := &WindowsClient{config: config, pending: make(map[uint64]chan controlReply), facts: make(chan WindowsFact, 4), started: make(chan WindowsStartResult, 1), stop: make(chan struct{}), done: make(chan struct{})}
	c.setupErr = c.create(ctx)
	go c.run()
	select {
	case result := <-c.started:
		if !result.Established {
			c.mu.Lock()
			err := c.latest.Err
			c.mu.Unlock()
			if err == nil {
				err = c.setupErr
			}
			return c, result, err
		}
		return c, result, nil
	case <-ctx.Done():
		c.Stop()
		return c, WindowsStartResult{}, ctx.Err()
	}
}

func (c *WindowsClient) create(ctx context.Context) error {
	var err error
	c.job, err = newJob()
	if err != nil {
		return err
	}
	var childRead, parentWrite, parentRead, childWrite windows.Handle
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), InheritHandle: 1}
	if err = windows.CreatePipe(&childRead, &parentWrite, &sa, 64<<10); err != nil {
		return err
	}
	c.childHandles = append(c.childHandles, childRead)
	c.write = os.NewFile(uintptr(parentWrite), "broker-control-write")
	if err = windows.SetHandleInformation(parentWrite, windows.HANDLE_FLAG_INHERIT, 0); err != nil {
		return err
	}
	if err = windows.CreatePipe(&parentRead, &childWrite, &sa, 64<<10); err != nil {
		return err
	}
	c.childHandles = append(c.childHandles, childWrite)
	c.read = os.NewFile(uintptr(parentRead), "broker-control-read")
	if err = windows.SetHandleInformation(parentRead, windows.HANDLE_FLAG_INHERIT, 0); err != nil {
		return err
	}
	var parent windows.Handle
	if err = windows.DuplicateHandle(windows.CurrentProcess(), windows.CurrentProcess(), windows.CurrentProcess(), &parent, windows.PROCESS_QUERY_INFORMATION|windows.SYNCHRONIZE, true, 0); err != nil {
		return err
	}
	c.childHandles = append(c.childHandles, parent)
	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return err
	}
	defer attrs.Delete()
	handles := []windows.Handle{childRead, childWrite, parent}
	if err = attrs.Update(windows.PROC_THREAD_ATTRIBUTE_HANDLE_LIST, unsafe.Pointer(&handles[0]), uintptr(len(handles))*unsafe.Sizeof(handles[0])); err != nil {
		return err
	}
	si := windows.StartupInfoEx{}
	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = windows.STARTF_USESHOWWINDOW | windows.STARTF_USESTDHANDLES
	si.ShowWindow = windows.SW_HIDE
	si.StdInput, si.StdOutput, si.StdErr = childRead, childWrite, parent
	si.ProcThreadAttributeList = attrs.List()
	app, err := windows.UTF16PtrFromString(c.config.Image)
	if err != nil {
		return err
	}
	cmd, err := windows.UTF16PtrFromString(windows.ComposeCommandLine([]string{c.config.Image, WindowsPrivateMode}))
	if err != nil {
		return err
	}
	// Bootstrap in the native system directory, never the user's project.
	system, err := windows.GetSystemDirectory()
	if err != nil {
		return err
	}
	dir, err := windows.UTF16PtrFromString(system)
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	err = windows.CreateProcess(app, cmd, nil, nil, true, windows.CREATE_SUSPENDED|windows.CREATE_NO_WINDOW|windows.CREATE_UNICODE_ENVIRONMENT|windows.EXTENDED_STARTUPINFO_PRESENT, nil, dir, &si.StartupInfo, &c.process)
	runtime.KeepAlive(handles)
	runtime.KeepAlive(attrs)
	if err != nil {
		return err
	}
	if err = windows.AssignProcessToJobObject(c.job, c.process.Process); err != nil {
		return err
	}
	for i := range c.childHandles {
		if err = closeHandle(&c.childHandles[i]); err != nil {
			return err
		}
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	count, err := windows.ResumeThread(c.process.Thread)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("unexpected broker suspend count")
	}
	if err = closeHandle(&c.process.Thread); err != nil {
		return err
	}
	nonce, err := FreshNonce()
	if err != nil {
		return err
	}
	c.channel, err = NewChannel(c.read, c.write, Parent, WindowsBroker, c.config.SessionID, nonce)
	if err != nil {
		return err
	}
	payload, err := EncodeStart(c.config.Spec)
	if err != nil {
		return err
	}
	_, err = c.send(Start, payload, nil)
	return err
}

func (c *WindowsClient) send(op Opcode, payload []byte, reply chan controlReply) (uint64, error) {
	c.controlMu.Lock()
	defer c.controlMu.Unlock()
	if c.channel == nil || c.write == nil {
		return 0, ErrProtocol
	}
	sequence := c.sequence + 1
	if sequence == 0 {
		return 0, ErrProtocol
	}
	if reply != nil {
		c.pendingMu.Lock()
		if len(c.pending) >= 64 {
			c.pendingMu.Unlock()
			return 0, errors.New("native controls busy")
		}
		c.pending[sequence] = reply
		c.pendingMu.Unlock()
	}
	err := sendControl(c.channel, c.write, op, payload)
	if err != nil {
		c.pendingMu.Lock()
		delete(c.pending, sequence)
		c.pendingMu.Unlock()
		return 0, err
	}
	c.sequence = sequence
	return sequence, nil
}

func (c *WindowsClient) publish(f WindowsFact) {
	c.mu.Lock()
	c.latest = f
	c.mu.Unlock()
	// At most Established, root/quiescent, retained failure and final cleanup
	// facts are published. Registry observation does not depend on consumption.
	c.facts <- f
}

func (c *WindowsClient) run() {
	defer close(c.done)
	defer close(c.facts)
	var current WindowsFact
	startedSent := false
	defer func() {
		if !startedSent {
			c.started <- WindowsStartResult{}
		}
	}()
	frames := make(chan receivedFrame, 1)
	readerDone := make(chan struct{})
	readerStop := make(chan struct{})
	if c.setupErr == nil {
		go func() {
			defer close(readerDone)
			for {
				f, e := c.channel.Receive()
				select {
				case frames <- receivedFrame{f, e}:
				case <-readerStop:
					return
				}
				if e != nil {
					return
				}
			}
		}()
	} else {
		close(readerDone)
	}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	err := c.setupErr
	stopSignal := c.stop
	var forceAt time.Time
	seenStreams := make(map[byte]bool)
	for err == nil {
		if c.process.Process != 0 {
			state, e := windows.WaitForSingleObject(c.process.Process, 0)
			if e != nil {
				err = e
				break
			}
			if state == windows.WAIT_OBJECT_0 {
				break
			}
		}
		if !forceAt.IsZero() && !time.Now().Before(forceAt) {
			err = context.DeadlineExceeded
			break
		}
		select {
		case got := <-frames:
			if got.err != nil {
				err = got.err
				break
			}
			f := got.frame
			switch f.Opcode {
			case OutputTransfer:
				if current.Established || len(f.Payload) != 9 || seenStreams[f.Payload[0]] {
					err = ErrProtocol
					break
				}
				stream := f.Payload[0]
				if stream < 1 || stream > 3 || (c.config.Spec.Terminal != (stream == 3)) {
					err = ErrProtocol
					break
				}
				remote := binary.BigEndian.Uint64(f.Payload[1:])
				if remote == 0 || uint64(windows.Handle(remote)) != remote {
					err = ErrProtocol
					break
				}
				var local windows.Handle
				if err = windows.DuplicateHandle(c.process.Process, windows.Handle(remote), windows.CurrentProcess(), &local, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
					break
				}
				file := os.NewFile(uintptr(local), "broker-transferred-output")
				c.outputFiles = append(c.outputFiles, file)
				kind, e := windows.GetFileType(local)
				if e != nil || kind != windows.FILE_TYPE_PIPE {
					err = ErrProtocol
					break
				}
				seenStreams[stream] = true
				c.outputDone.Add(1)
				go c.drain(file, api.OutputStream(stream))
				_, err = c.send(OutputAccepted, []byte{stream}, nil)
			case Started:
				want := 2
				if c.config.Spec.Terminal {
					want = 1
				}
				if current.Established || len(f.Payload) != 0 || len(seenStreams) != want {
					err = ErrProtocol
					break
				}
				current.Established = true
				c.publish(current)
				cwd := c.config.Spec.RootLocator
				for _, part := range c.config.Spec.Components {
					cwd += `\` + part
				}
				c.started <- WindowsStartResult{true, cwd}
				startedSent = true
			case Delivered:
				if len(f.Payload) != 8 && len(f.Payload) != 17 {
					err = ErrProtocol
					break
				}
				seq := binary.BigEndian.Uint64(f.Payload)
				c.pendingMu.Lock()
				reply := c.pending[seq]
				delete(c.pending, seq)
				c.pendingMu.Unlock()
				// Stop's optional ETX can also report a delivery without a caller.
				if reply != nil {
					result := controlReply{delivery: WindowsDelivery{Completed: true}}
					if len(f.Payload) == 17 {
						result.delivery.Accepted = binary.BigEndian.Uint32(f.Payload[8:])
						result.delivery.Delivered = binary.BigEndian.Uint32(f.Payload[12:])
						if f.Payload[16] != 0 {
							result.err = io.ErrShortWrite
						}
					}
					reply <- result
				}
			case UserExit:
				if len(f.Payload) != 4 || current.RootExited {
					err = ErrProtocol
					break
				}
				current.RootExited = true
				current.ExitCode = binary.BigEndian.Uint32(f.Payload)
			case Quiescent:
				if len(f.Payload) != 0 || !current.RootExited || current.Quiescent {
					err = ErrProtocol
					break
				}
				current.Quiescent = true
				c.publish(current)
				_, err = c.send(Release, nil, nil)
				forceAt = time.Now().Add(3 * time.Second)
			case Failure:
				err = errors.New("native broker reported failure")
			default:
				err = ErrProtocol
			}
		case <-stopSignal:
			stopSignal = nil
			forceAt = time.Now().Add(c.config.GracePeriod + c.config.ForcePeriod)
			payload := binary.BigEndian.AppendUint32(nil, uint32(c.config.GracePeriod/time.Millisecond))
			payload = binary.BigEndian.AppendUint32(payload, uint32(c.config.ForcePeriod/time.Millisecond))
			_, err = c.send(Stop, payload, nil)
		case <-ticker.C:
		}
	}
	if errors.Is(err, io.EOF) && current.Quiescent {
		err = nil
	}
	if !current.Quiescent && err == nil {
		err = errors.New("broker exited before user quiescence")
	}
	current.Err = err
	close(readerStop)
	// Every pending request receives a terminal delivery uncertainty; none is
	// replayed. The parent registry retains its own accepted queue accounting.
	c.pendingMu.Lock()
	for seq, reply := range c.pending {
		reply <- controlReply{err: errors.Join(err, io.ErrClosedPipe)}
		delete(c.pending, seq)
	}
	c.pendingMu.Unlock()
	var cleanupErr error
	reportedFailure := false
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		cleanupErr = c.finish(ctx, readerDone)
		cancel()
		if cleanupErr == nil {
			break
		}
		if !reportedFailure {
			current.Err = errors.Join(err, cleanupErr)
			c.publish(current)
			reportedFailure = true
		}
		time.Sleep(20 * time.Millisecond)
	}
	current.CleanupComplete = true
	current.Err = err
	c.publish(current)
}

func (c *WindowsClient) drain(file *os.File, stream api.OutputStream) {
	defer c.outputDone.Done()
	buffer := make([]byte, 32<<10)
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			c.config.Output(stream, append([]byte(nil), buffer[:n]...))
		}
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				c.outputMu.Lock()
				c.outputErr = errors.Join(c.outputErr, err)
				c.outputMu.Unlock()
			}
			return
		}
	}
}

func (c *WindowsClient) finish(ctx context.Context, readerDone <-chan struct{}) error {
	if c.job != 0 {
		if err := windows.TerminateJobObject(c.job, 1); err != nil {
			return err
		}
	}
	if c.process.Process != 0 {
		state, err := windows.WaitForSingleObject(c.process.Process, 0)
		if err != nil {
			return err
		}
		if state != windows.WAIT_OBJECT_0 {
			_ = windows.TerminateProcess(c.process.Process, 1)
		}
		if _, err = waitProcess(ctx, c.process.Process); err != nil {
			return err
		}
	}
	if c.job != 0 {
		if err := waitJob(ctx, c.job); err != nil {
			return err
		}
	}
	for i := range c.childHandles {
		if err := closeHandle(&c.childHandles[i]); err != nil {
			return err
		}
	}
	// No broker or outer member can retain pipe writers now. EOF and all
	// callbacks must join before the final observation is published.
	if c.read != nil {
		_ = c.read.Close()
	}
	<-readerDone
	c.controlMu.Lock()
	err := closeFile(&c.write)
	c.controlMu.Unlock()
	if err != nil {
		return err
	}
	c.outputDone.Wait()
	for i := range c.outputFiles {
		if err := closeFile(&c.outputFiles[i]); err != nil {
			return err
		}
	}
	return errors.Join(closeHandle(&c.process.Thread), closeHandle(&c.process.Process), closeHandle(&c.job))
}

func (c *WindowsClient) request(ctx context.Context, op Opcode, payload []byte) (WindowsDelivery, error) {
	if ctx == nil {
		return WindowsDelivery{}, ErrProtocol
	}
	select {
	case <-c.stop:
		return WindowsDelivery{}, io.ErrClosedPipe
	case <-c.done:
		return WindowsDelivery{}, io.ErrClosedPipe
	default:
	}
	reply := make(chan controlReply, 1)
	if _, err := c.send(op, append([]byte(nil), payload...), reply); err != nil {
		return WindowsDelivery{}, err
	}
	select {
	case result := <-reply:
		return result.delivery, result.err
	case <-ctx.Done():
		return WindowsDelivery{}, ctx.Err()
	case <-c.done:
		return WindowsDelivery{}, io.ErrClosedPipe
	}
}

func (c *WindowsClient) Write(ctx context.Context, data []byte) (WindowsDelivery, error) {
	if len(data) == 0 || len(data) > MaxFrame-headerSize {
		return WindowsDelivery{}, ErrProtocol
	}
	return c.request(ctx, WriteInput, data)
}
func (c *WindowsClient) Resize(ctx context.Context, rows, columns uint16) (WindowsDelivery, error) {
	if !c.config.Spec.Terminal || rows == 0 || columns == 0 || rows > 32767 || columns > 32767 {
		return WindowsDelivery{}, ErrProtocol
	}
	payload := binary.BigEndian.AppendUint16(nil, rows)
	payload = binary.BigEndian.AppendUint16(payload, columns)
	return c.request(ctx, Resize, payload)
}
func (c *WindowsClient) Interrupt(ctx context.Context) (WindowsDelivery, error) {
	if !c.config.Spec.Terminal {
		return WindowsDelivery{}, ErrProtocol
	}
	return c.request(ctx, Interrupt, nil)
}
func (c *WindowsClient) Stop() { c.stopOnce.Do(func() { close(c.stop) }) }
func (c *WindowsClient) NextFact(ctx context.Context) (WindowsFact, error) {
	select {
	case f, ok := <-c.facts:
		if !ok {
			return WindowsFact{}, io.EOF
		}
		return f, nil
	case <-ctx.Done():
		return WindowsFact{}, ctx.Err()
	}
}
func (c *WindowsClient) Wait(ctx context.Context) (WindowsFact, error) {
	select {
	case <-c.done:
		c.mu.Lock()
		f := c.latest
		c.mu.Unlock()
		return f, f.Err
	case <-ctx.Done():
		c.mu.Lock()
		f := c.latest
		c.mu.Unlock()
		return f, errors.Join(f.Err, ctx.Err())
	}
}
