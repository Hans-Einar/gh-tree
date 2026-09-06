package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	machine386    = 0x014c
	machineAMD64  = 0x8664
	machineARM64  = 0xaa64
	debugContinue = 0x00010002
	breakpoint    = 0x80000003
	wowBreakpoint = 0x4000001f
)

// MachineRoute reports only private Runtime implementation facts. Native32
// uses its own executable; emulated callers require the indicated native image.
func MachineRoute() (machine uint16, embedded bool, err error) {
	var process, native uint16
	if err = windows.IsWow64Process2(windows.CurrentProcess(), &process, &native); err != nil {
		return
	}
	if native != machine386 && native != machineAMD64 && native != machineARM64 {
		return 0, false, errors.New("unsupported native Windows machine")
	}
	if process != 0 && process != machine386 && process != machineAMD64 {
		return 0, false, errors.New("unsupported emulated Windows machine")
	}
	return native, process != 0, nil
}

type startupProfile struct {
	pointerSize                  uintptr
	pebParameters, parametersCwd uintptr
	wow                          bool
}

func processProfile(process windows.Handle) (startupProfile, error) {
	var target, native uint16
	if err := windows.IsWow64Process2(process, &target, &native); err != nil {
		return startupProfile{}, err
	}
	if native != machine386 && native != machineAMD64 && native != machineARM64 {
		return startupProfile{}, errors.New("unknown target runtime machine")
	}
	if target == machine386 && native != machine386 {
		return startupProfile{4, 0x10, 0x2c, true}, nil
	}
	if target != 0 {
		// x64-on-ARM64 has a distinct loader ABI. Enable only with its actual
		// native event-sequence evidence; it must never borrow WOW64's profile.
		return startupProfile{}, errors.New("unverified emulation startup profile")
	}
	var peb windows.PEB
	var parameters windows.RTL_USER_PROCESS_PARAMETERS
	return startupProfile{unsafe.Sizeof(uintptr(0)), unsafe.Offsetof(peb.ProcessParameters), unsafe.Offsetof(parameters.CurrentDirectory) + unsafe.Offsetof(parameters.CurrentDirectory.Handle), false}, nil
}

func readPointer(process windows.Handle, address, size uintptr) (uintptr, error) {
	if address == 0 || (size != 4 && size != 8) || size > unsafe.Sizeof(uintptr(0)) {
		return 0, errors.New("unsupported remote pointer")
	}
	var bytes [8]byte
	var got uintptr
	if err := windows.ReadProcessMemory(process, address, &bytes[0], size, &got); err != nil {
		return 0, err
	}
	if got != size {
		return 0, errors.New("incomplete remote pointer read")
	}
	if size == 4 {
		return uintptr(binary.LittleEndian.Uint32(bytes[:])), nil
	}
	return uintptr(binary.LittleEndian.Uint64(bytes[:])), nil
}

func childCwd(process windows.Handle, profile startupProfile) (windows.Handle, error) {
	var peb uintptr
	if profile.wow {
		if err := windows.NtQueryInformationProcess(process, windows.ProcessWow64Information, unsafe.Pointer(&peb), uint32(unsafe.Sizeof(peb)), nil); err != nil {
			return 0, err
		}
	} else {
		var info windows.PROCESS_BASIC_INFORMATION
		if err := windows.NtQueryInformationProcess(process, windows.ProcessBasicInformation, unsafe.Pointer(&info), uint32(unsafe.Sizeof(info)), nil); err != nil {
			return 0, err
		}
		peb = uintptr(unsafe.Pointer(info.PebBaseAddress))
	}
	if peb == 0 {
		return 0, errors.New("missing target PEB")
	}
	parameters, err := readPointer(process, peb+profile.pebParameters, profile.pointerSize)
	if err != nil || parameters == 0 {
		return 0, err
	}
	h, err := readPointer(process, parameters+profile.parametersCwd, profile.pointerSize)
	return windows.Handle(h), err
}

// debugOwner is used only on the OS thread which called CreateProcess. Pending
// events are retained until detach or explicit failed-start teardown.
type debugOwner struct {
	process                windows.ProcessInformation
	attached               bool
	pendingPID, pendingTID uint32
	owned                  []windows.Handle
}

func (d *debugOwner) continuation() error {
	if d.pendingPID == 0 {
		return nil
	}
	err := nativeCall("ContinueDebugEvent", uintptr(d.pendingPID), uintptr(d.pendingTID), debugContinue)
	if err == nil {
		d.pendingPID, d.pendingTID = 0, 0
	}
	return err
}

func (d *debugOwner) detach() error {
	if !d.attached {
		return nil
	}
	err := nativeCall("DebugActiveProcessStop", uintptr(d.process.ProcessId))
	if err == nil {
		d.attached = false
		d.pendingPID, d.pendingTID = 0, 0
	}
	return err
}

func (d *debugOwner) barrier(ctx context.Context, cwd *AcquiredDirectory, hook func(string)) error {
	profile, err := processProfile(d.process.Process)
	if err != nil {
		return err
	}
	if err = cwd.Revalidate(); err != nil {
		return err
	}
	if hook != nil {
		hook("before-user-resume")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	count, err := windows.ResumeThread(d.process.Thread)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("unexpected user suspend count")
	}
	nativeBreakSeen, created := false, false
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		// DEBUG_EVENT is 176 bytes on native64 and 96 on native32. A uintptr
		// array provides correct union alignment for both selected architectures.
		var storage [176 / 4]uintptr
		bytes := unsafe.Slice((*byte)(unsafe.Pointer(&storage[0])), int(unsafe.Sizeof(storage)))
		r, _, callErr := kernel.NewProc("WaitForDebugEventEx").Call(uintptr(unsafe.Pointer(&storage[0])), 50)
		if r == 0 {
			if callErr == windows.ERROR_SEM_TIMEOUT {
				continue
			}
			return callErr
		}
		code := binary.LittleEndian.Uint32(bytes)
		d.pendingPID, d.pendingTID = binary.LittleEndian.Uint32(bytes[4:]), binary.LittleEndian.Uint32(bytes[8:])
		if d.pendingPID != d.process.ProcessId {
			return errors.New("foreign debug event")
		}
		offset := 16
		if unsafe.Sizeof(uintptr(0)) == 4 {
			offset = 12
		}
		word := func(at int) windows.Handle {
			if unsafe.Sizeof(uintptr(0)) == 4 {
				return windows.Handle(binary.LittleEndian.Uint32(bytes[at:]))
			}
			return windows.Handle(binary.LittleEndian.Uint64(bytes[at:]))
		}
		switch code {
		case 3: // CREATE_PROCESS_DEBUG_EVENT; image, process and thread handles.
			if created {
				return errors.New("duplicate debug creation")
			}
			created = true
			for i := 0; i < 3; i++ {
				h := word(offset + i*int(unsafe.Sizeof(uintptr(0))))
				if h != 0 && h != d.process.Process && h != d.process.Thread {
					d.owned = append(d.owned, h)
				}
			}
		case 2: // CREATE_THREAD_DEBUG_EVENT
			h := word(offset)
			if h != 0 && h != d.process.Thread {
				d.owned = append(d.owned, h)
			}
		case 6: // LOAD_DLL_DEBUG_EVENT
			h := word(offset)
			if h != 0 {
				d.owned = append(d.owned, h)
			}
		case 1:
			if !created {
				return errors.New("breakpoint before process creation")
			}
			exception := binary.LittleEndian.Uint32(bytes[offset:])
			firstChanceOffset := offset + 152
			if unsafe.Sizeof(uintptr(0)) == 4 {
				firstChanceOffset = offset + 80
			}
			if binary.LittleEndian.Uint32(bytes[firstChanceOffset:]) != 1 {
				return errors.New("second-chance startup exception")
			}
			if profile.wow && !nativeBreakSeen && exception == breakpoint {
				nativeBreakSeen = true
				break
			}
			want := uint32(breakpoint)
			if profile.wow {
				want = wowBreakpoint
			}
			if exception != want || (profile.wow && !nativeBreakSeen) {
				return fmt.Errorf("unexpected startup exception 0x%x", exception)
			}
			if hook != nil {
				hook("cwd-breakpoint")
			}
			remote, e := childCwd(d.process.Process, profile)
			if e != nil {
				return e
			}
			if remote == 0 {
				return errors.New("runtime has not acquired cwd")
			}
			var duplicate windows.Handle
			if e = windows.DuplicateHandle(d.process.Process, remote, windows.CurrentProcess(), &duplicate, 0, false, windows.DUPLICATE_SAME_ACCESS); e != nil {
				return e
			}
			d.owned = append(d.owned, duplicate)
			actual, e := identity(duplicate)
			if e != nil {
				return e
			}
			expected, e := identity(cwd.target)
			if e != nil {
				return e
			}
			if actual != expected {
				return ErrCwd
			}
			if e = cwd.Revalidate(); e != nil {
				return e
			}
			if e = cwd.Close(); e != nil {
				return e
			}
			if hook != nil {
				hook("guards-released-pending-event")
			}
			// Detach at the still-pending approved runtime event. Continuing it
			// before detach would allow user initialization to race guard release.
			if e = d.detach(); e != nil {
				return e
			}
			return d.closeDebugHandles()
		case 4, 7, 8: // EXIT_THREAD, UNLOAD_DLL, OUTPUT_DEBUG_STRING
		case 5:
			return errors.New("user exited before verified runtime startup")
		default:
			return errors.New("unsupported loader debug event")
		}
		if err = d.closeDebugHandles(); err != nil {
			return err
		}
		if err = d.continuation(); err != nil {
			return err
		}
	}
}

func (d *debugOwner) closeDebugHandles() error {
	var err error
	for i := range d.owned {
		err = errors.Join(err, closeHandle(&d.owned[i]))
	}
	if err == nil {
		d.owned = nil
	}
	return err
}

// Assert the pinned x/sys native layouts against the selected Windows ABI.
// WOW64 values above are the same 32-bit types exercised by the 386 build/test.
func assertStartupLayout() error {
	var peb windows.PEB
	var parameters windows.RTL_USER_PROCESS_PARAMETERS
	p, c := uintptr(0x20), uintptr(0x48)
	if runtime.GOARCH == "386" {
		p, c = 0x10, 0x2c
	}
	if unsafe.Offsetof(peb.ProcessParameters) != p || unsafe.Offsetof(parameters.CurrentDirectory)+unsafe.Offsetof(parameters.CurrentDirectory.Handle) != c {
		return errors.New("pinned x/sys startup layout mismatch")
	}
	return nil
}
