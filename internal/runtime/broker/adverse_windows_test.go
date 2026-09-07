package broker

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func handleCount(t *testing.T) uint32 {
	t.Helper()
	var count uint32
	if err := nativeCall("GetProcessHandleCount", uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&count))); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestWindowsBlockedConPTYCloseRetainsOwner(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	s := windowsSpec(t)
	s.Environment = os.Environ()
	s.Terminal = true
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Executable = exe
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "pipe"}
	release := make(chan struct{})
	p := &userProcess{closeTerminal: func(h windows.Handle) { <-release; windows.ClosePseudoConsole(h) }}
	if err = p.prepare(s); err != nil {
		t.Fatal(err)
	}
	var readers sync.WaitGroup
	for _, out := range p.outputs {
		readers.Add(1)
		go func(out nativeOutput) {
			defer readers.Done()
			buffer := make([]byte, 1024)
			for {
				if _, e := out.file.Read(buffer); e != nil {
					return
				}
			}
		}(out)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = p.start(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err = waitProcess(ctx, p.debug.process.Process); err != nil {
		t.Fatal(err)
	}
	short, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	err = p.cleanup(short)
	stop()
	if !errors.Is(err, context.DeadlineExceeded) || p.hpc == 0 || p.job == 0 || p.terminalClosed == nil {
		t.Errorf("blocked close lost ownership: error=%v hpc=%x job=%x", err, p.hpc, p.job)
	}
	close(release)
	if err = p.cleanup(ctx); err != nil {
		t.Fatal(err)
	}
	readers.Wait()
	for i := range p.outputs {
		if err = closeFile(&p.outputs[i].file); err != nil {
			t.Error(err)
		}
	}
	if p.hpc != 0 || p.job != 0 {
		t.Fatal("released close did not join")
	}
}

func TestWindowsStopJoinsBlockedInputAndReportsPartial(t *testing.T) {
	s := windowsSpec(t)
	s.Environment = os.Environ()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Executable = exe
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}}
	if _, embedded, e := MachineRoute(); e != nil {
		t.Fatal(e)
	} else if embedded {
		config.Extraction = extractedNativeFixture(t)
		config.Image = config.Extraction.Path()
	}
	client, start, err := StartWindows(ctx, config)
	if err != nil || !start.Established {
		t.Fatalf("start %+v %v", start, err)
	}
	defer client.Stop()
	data := bytes.Repeat([]byte{0x41}, MaxFrame-headerSize)
	first, err := client.Write(ctx, data)
	if err != nil || first.Delivered != uint32(len(data)) {
		t.Fatalf("initial fill %+v %v", first, err)
	}
	type delivered struct {
		fact WindowsDelivery
		err  error
	}
	done := make(chan delivered, 1)
	go func() { fact, e := client.Write(ctx, data); done <- delivered{fact, e} }()
	select {
	case result := <-done:
		t.Fatalf("input was not blocked: %+v", result)
	case <-time.After(50 * time.Millisecond):
	}
	client.Stop()
	result := <-done
	if result.err == nil || result.fact.Accepted != uint32(len(data)) || result.fact.Delivered >= result.fact.Accepted {
		t.Fatalf("missing copied/partial-write facts: %+v", result)
	}
	final, err := client.Wait(ctx)
	if err != nil || !final.CleanupComplete {
		t.Fatalf("blocked input did not join: %+v %v", final, err)
	}
}

// Snapshot only this test process's kernel handle table. Classify the three
// owned native resource types using newly created probe capabilities; no
// globally enumerated process or numeric process identity is touched.
func nativeResourceCounts(t *testing.T) map[string]int {
	t.Helper()
	job, err := newJob()
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(job)
	var r, w, process windows.Handle
	if err = windows.CreatePipe(&r, &w, nil, 0); err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(r)
	defer windows.CloseHandle(w)
	if err = windows.DuplicateHandle(windows.CurrentProcess(), windows.CurrentProcess(), windows.CurrentProcess(), &process, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(process)
	type entry struct {
		Handle                             windows.Handle
		HandleCount, PointerCount          uintptr
		Access, Type, Attributes, Reserved uint32
	}
	storage := make([]uintptr, 1<<17)
	if err = windows.NtQueryInformationProcess(windows.CurrentProcess(), windows.ProcessHandleInformation, unsafe.Pointer(&storage[0]), uint32(len(storage))*uint32(unsafe.Sizeof(uintptr(0))), nil); err != nil {
		t.Fatal(err)
	}
	count := storage[0]
	if count > 4096 {
		t.Fatal("unbounded process handle snapshot")
	}
	entries := unsafe.Slice((*entry)(unsafe.Pointer(&storage[2])), count)
	types := make(map[uint32]string)
	for _, item := range entries {
		switch item.Handle {
		case job:
			types[item.Type] = "Job"
		case r:
			types[item.Type] = "File/Pipe"
		case process:
			types[item.Type] = "Process"
		}
	}
	if len(types) != 3 {
		t.Fatal("incomplete native handle type probes")
	}
	result := make(map[string]int)
	for _, item := range entries {
		if name := types[item.Type]; name != "" {
			result[name]++
		}
	}
	return result
}

func TestWindowsCanceledStartupUnwinds(t *testing.T) {
	for _, stage := range []string{"root-acquired", "project-acquired", "anchor-acquired", "inner-job-created", "conpty-created", "user-created-suspended", "user-assigned-inner-job", "before-user-resume"} {
		t.Run(stage, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			s := windowsSpec(t)
			s.Environment = os.Environ()
			s.Terminal = true
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var p *userProcess
			var original, proof windows.Handle
			var threadID uint32
			p = &userProcess{hook: func(actual string) {
				if actual == "user-created-suspended" {
					original = p.debug.process.Thread
					threadID = p.debug.process.ThreadId
					if e := windows.DuplicateHandle(windows.CurrentProcess(), original, windows.CurrentProcess(), &proof, 0, false, windows.DUPLICATE_SAME_ACCESS); e != nil {
						t.Error(e)
					}
				}
				if actual == stage {
					cancel()
				}
			}}
			if err = p.prepare(s); err != nil {
				t.Fatal(err)
			}
			var readers sync.WaitGroup
			for _, output := range p.outputs {
				readers.Add(1)
				go func(out nativeOutput) {
					defer readers.Done()
					buffer := make([]byte, 1024)
					for {
						if _, e := out.file.Read(buffer); e != nil {
							return
						}
					}
				}(output)
			}
			if err = p.start(ctx, s); !errors.Is(err, context.Canceled) {
				t.Errorf("cancel at %s returned %v", stage, err)
			}
			if err = boundedCleanup(p); err != nil {
				t.Fatal(err)
			}
			if proof != 0 {
				assertOwnedThreadHandleClosed(t, original, threadID)
				if e := windows.CloseHandle(proof); e != nil {
					t.Error(e)
				}
			}
			readers.Wait()
			for i := range p.outputs {
				if err = closeFile(&p.outputs[i].file); err != nil {
					t.Error(err)
				}
			}
			if _, err = os.Stat(filepath.Join(s.RootLocator, "user-ran")); !os.IsNotExist(err) {
				t.Fatalf("canceled startup ran user initialization: %v", err)
			}
			if p.job != 0 || p.debug.process.Process != 0 || p.debug.attached || p.hpc != 0 {
				t.Fatalf("retained native resources after cancellation: %+v", p)
			}
		})
	}
}

func assertOwnedThreadHandleClosed(t *testing.T, original windows.Handle, threadID uint32) {
	t.Helper()
	// A separately retained proof duplicate pins the original thread identity
	// throughout this assertion. Numeric thread-ID reuse therefore cannot make
	// an unrelated recycled handle look like this particular owned thread.
	id, _, _ := kernel.NewProc("GetThreadId").Call(uintptr(original))
	if uint32(id) == threadID {
		t.Fatalf("original owned primary-thread handle still open for thread %d", threadID)
	}
}

func TestWindowsBrokerPrimaryThreadClosure(t *testing.T) {
	for _, stage := range []string{"success", "broker-created-suspended", "broker-before-resume"} {
		t.Run(stage, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "pipe"}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			join, cancelJoin := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelJoin()
			var original, proof windows.Handle
			var threadID uint32
			config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}}
			if _, embedded, e := MachineRoute(); e != nil {
				t.Fatal(e)
			} else if embedded {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			config.hook = func(actual string, client *WindowsClient) {
				if actual == "broker-created-suspended" {
					original = client.process.Thread
					threadID = client.process.ThreadId
					if e := windows.DuplicateHandle(windows.CurrentProcess(), original, windows.CurrentProcess(), &proof, 0, false, windows.DUPLICATE_SAME_ACCESS); e != nil {
						t.Error(e)
					}
				}
				if actual == "broker-before-resume" {
					if count, e := activeProcesses(client.job); e != nil || count != 1 {
						t.Errorf("broker not contained before resume: %d %v", count, e)
					}
				}
				if actual == stage {
					cancel()
				}
			}
			client, start, err := StartWindows(ctx, config)
			if client == nil {
				t.Fatal("lost acquired broker owner")
			}
			if stage == "success" {
				if err != nil || !start.Established {
					t.Fatalf("start %+v %v", start, err)
				}
			} else {
				if err == nil || start.Established {
					t.Fatalf("canceled start %+v %v", start, err)
				}
			}
			client.Stop()
			final, _ := client.Wait(join)
			if !final.CleanupComplete {
				t.Fatalf("cleanup incomplete %+v", final)
			}
			if proof == 0 {
				t.Fatal("no owned primary thread proof")
			}
			assertOwnedThreadHandleClosed(t, original, threadID)
			if err = windows.CloseHandle(proof); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWindowsBrokerCanceledAndFailedStart(t *testing.T) {
	for _, which := range []string{"missing-executable", "stale-cwd", "cancel-start"} {
		t.Run(which, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
			if which == "missing-executable" {
				s.Executable = filepath.Join(s.RootLocator, "nonexistent-owned-fixture.exe")
			}
			if which == "stale-cwd" {
				other := windowsSpec(t)
				s.RootIdentity = other.RootIdentity
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			startCtx := ctx
			if which == "cancel-start" {
				short, c := context.WithTimeout(ctx, time.Millisecond)
				defer c()
				startCtx = short
			}
			config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}, GracePeriod: time.Millisecond, ForcePeriod: time.Second}
			if _, embedded, e := MachineRoute(); e != nil {
				t.Fatal(e)
			} else if embedded {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			client, started, err := StartWindows(startCtx, config)
			if err == nil || started.Established {
				t.Fatalf("failed start accepted: %+v %v", started, err)
			}
			if client == nil {
				if which != "cancel-start" {
					t.Fatal("partial acquisition lost owner")
				}
				return
			}
			client.Stop()
			final, _ := client.Wait(ctx)
			if !final.CleanupComplete {
				t.Fatalf("failed start did not join: %+v", final)
			}
		})
	}
}

func TestWindowsNativeResourceCycles(t *testing.T) {
	// Fix the scheduler width so background GC does not grow an unrelated
	// pool of OS-thread/event handles between native ownership snapshots.
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)
	// Warm native Go/loader/terminal facilities before comparing retained handles.
	cycle := func() {
		s := windowsSpec(t)
		s.Environment = os.Environ()
		exe, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		s.Executable = exe
		s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "pipe"}
		s.Terminal = true
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}}
		if _, embedded, e := MachineRoute(); e != nil {
			t.Fatal(e)
		} else if embedded {
			config.Extraction = extractedNativeFixture(t)
			config.Image = config.Extraction.Path()
		}
		client, _, err := StartWindows(ctx, config)
		if err != nil {
			t.Fatal(err)
		}
		final, err := client.Wait(ctx)
		if err != nil || !final.CleanupComplete {
			t.Fatalf("resource cycle %+v %v", final, err)
		}
	}
	for i := 0; i < 3; i++ {
		cycle()
		runtime.GC()
	}
	runtime.GC()
	before := handleCount(t)
	ownedBefore := nativeResourceCounts(t)
	goroutines := runtime.NumGoroutine()
	for i := 0; i < 8; i++ {
		cycle()
	}
	runtime.GC()
	after := handleCount(t)
	ownedAfter := nativeResourceCounts(t)
	for name, count := range ownedBefore {
		if ownedAfter[name] != count {
			t.Fatalf("native %s handle leak: before=%v after=%v", name, ownedBefore, ownedAfter)
		}
	}
	if got := runtime.NumGoroutine(); got > goroutines {
		t.Fatalf("goroutine leak: before=%d after=%d", goroutines, got)
	}
	t.Logf("joined ConPTY resource cycles: owned handles %v -> %v; total including Go thread/event pool %d -> %d, goroutines %d -> %d", ownedBefore, ownedAfter, before, after, goroutines, runtime.NumGoroutine())
}

func TestWindowsPartialNativeAcquisitionFailures(t *testing.T) {
	for _, stage := range []string{"cwd-volume-opened", "cwd-child-opened", "root-acquired", "project-acquired", "anchor-acquired", "inner-job-created", "pipe-created", "conpty-created", "attributes-created", "user-created-suspended", "user-assigned-inner-job", "target-profile-acquired", "before-user-resume", "debug-image-acquired", "cwd-handle-duplicated", "cwd-identity-verified"} {
		t.Run(stage, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			s := windowsSpec(t)
			s.Environment = os.Environ()
			s.Terminal = true
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
			injected := errors.New("owned fixture injected acquisition failure")
			p := &userProcess{fault: func(actual string) error {
				if actual == stage {
					return injected
				}
				return nil
			}}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err = p.prepare(s)
			var readers sync.WaitGroup
			for _, out := range p.outputs {
				readers.Add(1)
				go func(out nativeOutput) {
					defer readers.Done()
					buffer := make([]byte, 1024)
					for {
						if _, e := out.file.Read(buffer); e != nil {
							return
						}
					}
				}(out)
			}
			if err == nil {
				err = p.start(ctx, s)
			}
			if !errors.Is(err, injected) {
				t.Errorf("failure stage %s returned %v", stage, err)
			}
			if err = p.cleanup(ctx); err != nil {
				t.Fatal(err)
			}
			readers.Wait()
			for i := range p.outputs {
				if err = closeFile(&p.outputs[i].file); err != nil {
					t.Error(err)
				}
			}
			if _, err = os.Stat(filepath.Join(s.RootLocator, "user-ran")); !os.IsNotExist(err) {
				t.Fatalf("failed acquisition ran user initialization: %v", err)
			}
			if p.job != 0 || p.hpc != 0 || p.debug.process.Process != 0 || p.debug.process.Thread != 0 || p.debug.attached {
				t.Fatal("partial owner retained native resources")
			}
		})
	}
}

func TestWindowsPartialParentAcquisitionFailures(t *testing.T) {
	for _, stage := range []string{"outer-job-created", "control-input-pipe-created", "control-output-pipe-created", "parent-capability-created", "broker-attributes-created", "broker-created-suspended", "broker-before-resume", "broker-resumed"} {
		t.Run(stage, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
			injected := errors.New("owned fixture injected parent failure")
			config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}, fault: func(actual string) error {
				if actual == stage {
					return injected
				}
				return nil
			}}
			if _, emulated, e := MachineRoute(); e != nil {
				t.Fatal(e)
			} else if emulated {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, start, err := StartWindows(ctx, config)
			if client == nil || start.Established || !errors.Is(err, injected) {
				t.Fatalf("partial Start lost injected facts: client=%v %+v %v", client != nil, start, err)
			}
			final, _ := client.Wait(ctx)
			if !final.CleanupComplete || len(final.Residuals) != 0 {
				t.Fatalf("partial parent cleanup %+v", final)
			}
			if _, err = os.Stat(filepath.Join(s.RootLocator, "user-ran")); !os.IsNotExist(err) {
				t.Fatalf("partial parent startup ran user initialization: %v", err)
			}
		})
	}
}

func TestWindowsBrokerFailureAndForeignFrameCleanup(t *testing.T) {
	for _, mode := range []string{"broker-crash", "foreign-frame", "control-eof"} {
		t.Run(mode, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			s.Executable = exe
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
			var mu sync.Mutex
			var output bytes.Buffer
			config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(_ api.OutputStream, b []byte) { mu.Lock(); output.Write(b); mu.Unlock() }}
			if _, emulated, e := MachineRoute(); e != nil {
				t.Fatal(e)
			} else if emulated {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, start, err := StartWindows(ctx, config)
			if err != nil || !start.Established {
				t.Fatalf("start %+v %v", start, err)
			}
			defer client.Stop()
			for {
				mu.Lock()
				ready := bytes.Contains(output.Bytes(), []byte("OWNED_CHILD"))
				mu.Unlock()
				if ready {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("tree not ready")
				case <-time.After(time.Millisecond):
				}
			}
			var retainedJob windows.Handle
			if err = windows.DuplicateHandle(windows.CurrentProcess(), client.job, windows.CurrentProcess(), &retainedJob, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
				t.Fatal(err)
			}
			defer windows.CloseHandle(retainedJob)
			if active, e := activeProcesses(retainedJob); e != nil || active < 4 {
				t.Fatalf("missing broker/root/child/grandchild containment: %d %v", active, e)
			}
			switch mode {
			case "broker-crash":
				if err = windows.TerminateProcess(client.process.Process, 77); err != nil {
					t.Fatal(err)
				}
			case "foreign-frame":
				client.controlMu.Lock()
				nonce := client.channel.nonce
				nonce[0] ^= 1
				frame, e := EncodeFrame(Frame{Role: Parent, Opcode: Stop, SessionID: config.SessionID, Sequence: client.sequence + 1, Nonce: nonce})
				if e == nil {
					_, e = client.write.Write(frame)
				}
				client.controlMu.Unlock()
				if e != nil {
					t.Fatal(e)
				}
			case "control-eof":
				client.controlMu.Lock()
				err = closeFile(&client.write)
				client.controlMu.Unlock()
				if err != nil {
					t.Fatal(err)
				}
			}
			final, err := client.Wait(ctx)
			if !final.CleanupComplete || err == nil || len(final.Residuals) != 0 {
				t.Fatalf("failed protocol/broker cleanup facts %+v %v", final, err)
			}
			if active, e := activeProcesses(retainedJob); e != nil || active != 0 {
				t.Fatalf("outer residual survived failure: %d %v", active, e)
			}
		})
	}
}

func TestWindowsRawOutputAndDetachedConsumer(t *testing.T) {
	s := windowsSpec(t)
	s.Environment = os.Environ()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Executable = exe
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "flood"}
	var mu sync.Mutex
	counts := make(map[api.OutputStream]int)
	var retained []byte
	var retainedCopy []byte
	config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(stream api.OutputStream, data []byte) {
		mu.Lock()
		defer mu.Unlock()
		counts[stream] += len(data)
		if retained == nil && bytes.Contains(data, []byte{0, 0xff, 0x1b}) {
			retained = data
			retainedCopy = append([]byte(nil), data...)
		}
	}}
	if _, emulated, e := MachineRoute(); e != nil {
		t.Fatal(e)
	} else if emulated {
		config.Extraction = extractedNativeFixture(t)
		config.Image = config.Extraction.Path()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, start, err := StartWindows(ctx, config)
	if err != nil || !start.Established {
		t.Fatalf("start %+v %v", start, err)
	}
	// Do not consume any lifecycle hint while two MiB of raw output drains.
	final, err := client.Wait(ctx)
	if err != nil || !final.CleanupComplete {
		t.Fatalf("detached consumer %+v %v", final, err)
	}
	mu.Lock()
	defer mu.Unlock()
	if counts[api.Stdout] < 1<<20 || counts[api.Stderr] != 1<<20 {
		t.Fatalf("raw stream byte loss: %v", counts)
	}
	if retained == nil || !bytes.Equal(retained, retainedCopy) {
		t.Fatal("output callback's retained copy mutated across native reads")
	}
}

func TestWindowsWaitTimeoutKeepsTypedResidualsAndCoalescesStop(t *testing.T) {
	s := windowsSpec(t)
	s.Environment = os.Environ()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Executable = exe
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
	config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}}
	if _, emulated, e := MachineRoute(); e != nil {
		t.Fatal(e)
	} else if emulated {
		config.Extraction = extractedNativeFixture(t)
		config.Image = config.Extraction.Path()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, start, err := StartWindows(ctx, config)
	if err != nil || !start.Established {
		t.Fatalf("start %+v %v", start, err)
	}
	defer client.Stop()
	short, stop := context.WithTimeout(context.Background(), time.Millisecond)
	defer stop()
	partial, err := client.Wait(short)
	if !errors.Is(err, context.DeadlineExceeded) || partial.CleanupComplete || len(partial.Residuals) == 0 {
		t.Fatalf("timeout lost retained barriers %+v %v", partial, err)
	}
	partial.Residuals[0].Stage = 0
	again, _ := client.Wait(short)
	for _, residual := range again.Residuals {
		if !residual.Stage.Valid() {
			t.Fatal("returned residual aliases owner state")
		}
	}
	var callers sync.WaitGroup
	for n := 0; n < 32; n++ {
		callers.Add(1)
		go func() { defer callers.Done(); client.Stop() }()
	}
	callers.Wait()
	final, err := client.Wait(ctx)
	if err != nil || !final.CleanupComplete || len(final.Residuals) != 0 {
		t.Fatalf("coalesced Stop %+v %v", final, err)
	}
}
