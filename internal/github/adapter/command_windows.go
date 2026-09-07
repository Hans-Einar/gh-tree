//go:build windows

package adapter

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type commandOwner struct {
	job      windows.Handle
	assigned bool
}

func prepareCommand(cmd *exec.Cmd) (*commandOwner, error) {
	job, e := windows.CreateJobObject(nil, nil)
	if e != nil {
		return nil, e
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, e = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); e != nil {
		windows.CloseHandle(job)
		return nil, e
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_SUSPENDED | windows.CREATE_NO_WINDOW}
	return &commandOwner{job: job}, nil
}
func (o *commandOwner) started(cmd *exec.Cmd) error {
	p, e := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if e != nil {
		return e
	}
	defer windows.CloseHandle(p)
	if e = windows.AssignProcessToJobObject(o.job, p); e != nil {
		return e
	}
	o.assigned = true
	// The suspended root has one initial thread. Resume only after nonbreakaway
	// Job ownership is established; failure leaves all user code suspended.
	snapshot, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if e != nil {
		return e
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	for e = windows.Thread32First(snapshot, &entry); e == nil; e = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != uint32(cmd.Process.Pid) {
			continue
		}
		thread, e := windows.OpenThread(windows.THREAD_SUSPEND_RESUME|windows.THREAD_QUERY_LIMITED_INFORMATION, false, entry.ThreadID)
		if e != nil {
			return e
		}
		defer windows.CloseHandle(thread)
		return resumeRootThread(thread, uint32(cmd.Process.Pid))
	}
	return errors.New("initial thread unavailable")
}

var processIDOfThread = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetProcessIdOfThread")

func resumeRootThread(thread windows.Handle, root uint32) error {
	// Enumeration is only a locator. Validate the opened handle before resuming:
	// an exited/reused TID must never authorize another process's thread.
	pid, _, _ := processIDOfThread.Call(uintptr(thread))
	if pid == 0 || uint32(pid) != root {
		return errors.New("opened thread does not belong to retained root")
	}
	previous, e := windows.ResumeThread(thread)
	if e != nil {
		return e
	}
	if previous != 1 {
		return errors.New("unexpected initial suspend count")
	}
	return nil
}
func (o *commandOwner) stop(cmd *exec.Cmd) {
	if o.assigned {
		_ = windows.TerminateJobObject(o.job, 1)
	}
	_ = cmd.Process.Kill()
}
func (o *commandOwner) finish(cmd *exec.Cmd, budget time.Duration) bool {
	if !o.assigned {
		return false
	}
	_ = windows.TerminateJobObject(o.job, 1)
	// JOBOBJECT_BASIC_ACCOUNTING_INFORMATION has four LARGE_INTEGERs followed
	// by four DWORDs on all supported Windows ABIs (48 bytes).
	type accounting struct {
		times                                 [4]int64
		PageFaults, Total, Active, Terminated uint32
	}
	deadline := time.Now().Add(budget)
	for {
		var info accounting
		e := windows.QueryInformationJobObject(o.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info)), nil)
		if e != nil {
			return false
		}
		if info.Active == 0 {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}
func (o *commandOwner) close() { _ = windows.CloseHandle(o.job) }
