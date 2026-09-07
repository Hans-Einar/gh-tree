package broker

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func TestWindowsTargetABIMatrix(t *testing.T) {
	machine, emulated, err := MachineRoute()
	if err != nil {
		t.Fatal(err)
	}
	arches := []string{"386"}
	if machine == machineAMD64 || machine == machineARM64 {
		arches = append(arches, "amd64")
	}
	if machine == machineARM64 {
		arches = append(arches, "arm64")
	}
	for _, arch := range arches {
		t.Run(arch, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			s.Executable = fixtureExecutable(t, arch)
			s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "pipe"}
			current, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			config := WindowsConfig{SessionID: 17, Spec: s, Image: current, Output: func(api.OutputStream, []byte) {}}
			if emulated {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			client, start, err := StartWindows(ctx, config)
			if client != nil {
				client.Stop()
				final, _ := client.Wait(ctx)
				if !final.CleanupComplete {
					t.Fatalf("ABI fixture ownership did not join: %+v", final)
				}
			}
			if err != nil || !start.Established {
				if machine == machineARM64 && arch == "amd64" && runtime.GOARCH == "arm64" {
					traceARM64X64(t, s)
				}
				t.Fatalf("native machine=%04x parent=%s target=%s not established: %+v %v", machine, runtime.GOARCH, arch, start, err)
			}
			t.Logf("actual native machine=%04x parent=%s target=%s established and joined", machine, runtime.GOARCH, arch)
		})
	}
}

// Diagnostic-only bounded loader probe for the as-yet-unverified x64-on-ARM64
// profile. It never treats a breakpoint as production Start and never releases
// the user root: at the observed candidate it terminates the owned test Job.
func traceARM64X64(t *testing.T, spec StartSpec) {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	p := &userProcess{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.prepare(spec); err != nil {
		t.Logf("ABI_TRACE prepare %v", err)
		return
	}
	defer func() {
		if err := boundedCleanup(p); err != nil {
			t.Errorf("ABI_TRACE cleanup %v", err)
		}
		for i := range p.outputs {
			_ = closeFile(&p.outputs[i].file)
		}
	}()
	if err := p.start(ctx, spec); err == nil {
		t.Error("ABI_TRACE unexpectedly established production profile")
		return
	} else {
		t.Logf("ABI_TRACE production refusal %v", err)
	}
	if p.debug.process.Process == 0 {
		return
	}
	if _, err := windows.ResumeThread(p.debug.process.Thread); err != nil {
		t.Logf("ABI_TRACE resume %v", err)
		return
	}
	profile := startupProfile{pointerSize: 8, pebParameters: 0x20, parametersCwd: 0x48}
	for events := 0; events < 128; events++ {
		if ctx.Err() != nil {
			t.Log("ABI_TRACE deadline")
			return
		}
		var storage [22]uint64
		b := unsafe.Slice((*byte)(unsafe.Pointer(&storage[0])), 176)
		r, _, callErr := kernel.NewProc("WaitForDebugEventEx").Call(uintptr(unsafe.Pointer(&storage[0])), 100)
		if r == 0 {
			if callErr == windows.ERROR_SEM_TIMEOUT {
				events--
				continue
			}
			t.Logf("ABI_TRACE wait %v", callErr)
			return
		}
		code := binary.LittleEndian.Uint32(b)
		p.debug.pendingPID, p.debug.pendingTID = binary.LittleEndian.Uint32(b[4:]), binary.LittleEndian.Uint32(b[8:])
		if p.debug.pendingPID != p.debug.process.ProcessId {
			t.Error("ABI_TRACE foreign event")
			return
		}
		cwd, e := childCwd(p.debug.process.Process, profile)
		exc := uint32(0)
		if code == 1 {
			exc = binary.LittleEndian.Uint32(b[16:])
		}
		t.Logf("ABI_TRACE event=%d exception=%08x native64Cwd=%x readError=%v", code, exc, cwd, e)
		if code == 3 || code == 6 {
			h := windows.Handle(binary.LittleEndian.Uint64(b[16:]))
			if h != 0 {
				if e := windows.CloseHandle(h); e != nil {
					t.Errorf("ABI_TRACE image close %v", e)
					return
				}
			}
		}
		if code == 1 {
			if exc != breakpoint && exc != wowBreakpoint {
				t.Log("ABI_TRACE unsupported exception; refusing continuation")
				return
			}
			if cwd != 0 && e == nil {
				var duplicate windows.Handle
				if e = windows.DuplicateHandle(p.debug.process.Process, cwd, windows.CurrentProcess(), &duplicate, 0, false, windows.DUPLICATE_SAME_ACCESS); e != nil {
					t.Logf("ABI_TRACE duplicate %v", e)
					return
				}
				actual, e := identity(duplicate)
				_ = windows.CloseHandle(duplicate)
				expected, ee := identity(p.cwd.target)
				_, markerErr := os.Stat(filepath.Join(spec.RootLocator, "user-ran"))
				t.Logf("ABI_TRACE CANDIDATE exception=%08x fullIDMatch=%v identityErrors=%v/%v userInitializationAbsent=%v", exc, actual == expected, e, ee, os.IsNotExist(markerErr))
				return
			}
		}
		if code == 5 {
			return
		}
		if code < 1 || code > 8 {
			t.Logf("ABI_TRACE unsupported event %s", fmt.Sprint(code))
			return
		}
		if e = p.debug.continuation(); e != nil {
			t.Logf("ABI_TRACE continue %v", e)
			return
		}
	}
}
