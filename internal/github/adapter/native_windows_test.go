//go:build windows

package adapter

import (
	"bytes"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"unsafe"
)

func TestNativeWindowsFailedAssignmentNeverRunsChild(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestNativeCommandHelper$")
	cmd.Env = append(os.Environ(), "GH_TREE_ADAPTER_HELPER=descendant")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	owner, e := prepareCommand(cmd)
	if e != nil {
		t.Fatal(e)
	}
	if e = cmd.Start(); e != nil {
		owner.close()
		t.Fatal(e)
	}
	// Invalidate the native Job before assignment: the actual product acquisition
	// method must fail while user code (including immediate spawn) is suspended.
	owner.close()
	if e = owner.started(cmd); e == nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatal("assignment to invalid Job succeeded")
	}
	if e = cmd.Process.Kill(); e != nil {
		t.Fatal(e)
	}
	_ = cmd.Wait()
	if output.Len() != 0 {
		t.Fatal("uncontained user code ran", output.String())
	}
	verifyFixtureChildExited(t, cmd.Process.Pid)
}

func hideFixture(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
}
func verifyFixtureChildExited(t *testing.T, pid int) {
	t.Helper()
	handle, e := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if e == windows.ERROR_INVALID_PARAMETER {
		return
	}
	if e != nil {
		t.Fatal(e)
	}
	defer windows.CloseHandle(handle)
	state, e := windows.WaitForSingleObject(handle, 1000)
	if e != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatalf("child still alive %d: %d %v", pid, state, e)
	}
}
func TestNativeWindowsAccountingABI(t *testing.T) {
	type accounting struct {
		Times                             [4]int64
		Faults, Total, Active, Terminated uint32
	}
	var v accounting
	if unsafe.Sizeof(v) != 48 || unsafe.Offsetof(v.Active) != 40 {
		t.Fatal("Job accounting ABI")
	}
}
func TestNativeWindowsForeignOpenedThreadRefusesResume(t *testing.T) {
	thread, e := windows.GetCurrentThread()
	if e != nil {
		t.Fatal(e)
	}
	if e = resumeRootThread(thread, uint32(os.Getpid())+1); e == nil {
		t.Fatal("foreign thread resumed")
	}
	if e = resumeRootThread(windows.Handle(0), uint32(os.Getpid())); e == nil {
		t.Fatal("invalid thread resumed")
	}
}
