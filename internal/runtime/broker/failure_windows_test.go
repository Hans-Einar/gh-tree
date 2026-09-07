package broker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func assertWindowsFailure(t *testing.T, err error, cause WindowsFailureCause, stage api.RuntimeCleanupStage, cleanup bool, sentinel error) {
	t.Helper()
	var failure *WindowsFailure
	if !errors.As(err, &failure) || failure.Cause != cause || failure.Stage != stage || failure.Cleanup != cleanup || (sentinel != nil && !errors.Is(err, sentinel)) {
		t.Fatalf("failure lost classification: %v, typed=%+v", err, failure)
	}
}

func TestWindowsFailureCodecAndNativeStatuses(t *testing.T) {
	for _, item := range []struct {
		name     string
		err      error
		cause    WindowsFailureCause
		sentinel error
	}{
		{"cwd", ErrCwd, WindowsCwdFailure, ErrCwd},
		{"not-found", os.ErrNotExist, WindowsNotFoundFailure, os.ErrNotExist},
		{"nt-not-found", windows.STATUS_OBJECT_NAME_NOT_FOUND, WindowsNotFoundFailure, os.ErrNotExist},
		{"permission", windows.ERROR_ACCESS_DENIED, WindowsPermissionFailure, os.ErrPermission},
		{"nt-permission", windows.STATUS_ACCESS_DENIED, WindowsPermissionFailure, os.ErrPermission},
		{"profile", ErrWindowsUnsupported, WindowsUnsupportedFailure, ErrWindowsUnsupported},
		{"process", windows.ERROR_BAD_EXE_FORMAT, WindowsProcessFailure, ErrWindowsProcess},
		{"protocol", ErrProtocol, WindowsProtocolFailure, ErrProtocol},
		{"cancel", context.Canceled, WindowsCanceledFailure, context.Canceled},
		{"timeout", context.DeadlineExceeded, WindowsTimeoutFailure, context.DeadlineExceeded},
		{"busy", ErrWindowsBusy, WindowsBusyFailure, ErrWindowsBusy},
	} {
		t.Run(item.name, func(t *testing.T) {
			original := &os.PathError{Op: "private-operation", Path: "secret-path-argv-environment", Err: item.err}
			local := windowsFailureAt(original, api.ProcessContainment)
			assertWindowsFailure(t, local, item.cause, api.ProcessContainment, false, item.sentinel)
			var retained *os.PathError
			if !errors.As(local, &retained) || retained != original || strings.Contains(local.Error(), original.Path) {
				t.Fatal("local cause lost or private text exposed")
			}
			payload, err := encodeWindowsFailure(local)
			if err != nil || len(payload) != 5 {
				t.Fatalf("encode %x %v", payload, err)
			}
			remote := decodeWindowsFailure(payload)
			assertWindowsFailure(t, remote, item.cause, api.ProcessContainment, false, item.sentinel)
			if errors.As(remote, &retained) || strings.Contains(remote.Error(), original.Path) {
				t.Fatal("private cause crossed wire")
			}
		})
	}
	primary := windowsFailureAt(ErrCwd, api.CwdAcquisition)
	cleanup := windowsCleanupFailure(windows.ERROR_ACCESS_DENIED, api.Descendants)
	payload, err := encodeWindowsFailure(errors.Join(primary, cleanup))
	if err != nil {
		t.Fatal(err)
	}
	remote := decodeWindowsFailure(payload)
	if !errors.Is(remote, ErrCwd) || !errors.Is(remote, os.ErrPermission) || !errors.Is(remote, ErrWindowsCleanup) {
		t.Fatalf("joined failures erased: %v", remote)
	}
	wrapped := windowsCleanupFailure(errors.Join(primary, cleanup), api.SupervisorOrBroker)
	payload, err = encodeWindowsFailure(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	joined := decodeWindowsFailure(payload).(interface{ Unwrap() []error }).Unwrap()
	if len(joined) != 2 {
		t.Fatalf("cleanup wrapper lost errors: %v", joined)
	}
	assertWindowsFailure(t, joined[0], WindowsCwdFailure, api.CwdAcquisition, true, ErrCwd)
	assertWindowsFailure(t, joined[1], WindowsPermissionFailure, api.Descendants, true, os.ErrPermission)
}

func TestWindowsFailureNativeProfileStage(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	s := windowsSpec(t)
	s.Environment = os.Environ()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Executable = exe
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}
	p := &userProcess{fault: func(stage string) error {
		if stage == "target-profile-acquired" {
			return ErrWindowsUnsupported
		}
		return nil
	}}
	defer func() {
		if e := boundedCleanup(p); e != nil {
			t.Error(e)
		}
		for i := range p.outputs {
			if e := closeFile(&p.outputs[i].file); e != nil {
				t.Error(e)
			}
		}
	}()
	if err := p.prepare(s); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = p.start(ctx, s)
	assertWindowsFailure(t, err, WindowsUnsupportedFailure, api.SupervisorOrBroker, false, ErrWindowsUnsupported)
	if p.debug.process.Process == 0 {
		t.Fatal("native profile failure was not reached")
	}
	if _, err := os.Stat(filepath.Join(s.RootLocator, "user-ran")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("user initialization ran before rejected profile: %v", err)
	}
}

func TestWindowsFailureRejectsMalformedPayload(t *testing.T) {
	for _, payload := range [][]byte{
		nil, {}, {1}, {2, 1, 1, 1, 0}, {1, 0}, {1, 17}, {1, 1, 1, 1},
		{1, 1, 0, 1, 0}, {1, 1, 255, 1, 0}, {1, 1, 1, 0, 0}, {1, 1, 1, 255, 0},
		{1, 1, 1, 1, 2}, {1, 1, 1, 1, 0, 0}, {1, 2, 1, 1, 0, 1, 1, 0},
	} {
		if err := decodeWindowsFailure(payload); err != ErrProtocol {
			t.Errorf("accepted malformed payload %x: %v", payload, err)
		}
	}
	if _, err := encodeWindowsFailure(&WindowsFailure{Cause: 255, Stage: api.Acquisition}); err != ErrProtocol {
		t.Fatal("encoded invalid local enum")
	}
	if _, err := encodeWindowsFailure(nil); err != ErrProtocol {
		t.Fatal("encoded absent failure")
	}
}

// Restrict only one newly created fixture image. The retained handle restores
// its exact original DACL before the ordinary temporary-directory cleanup.
func denyOwnedExecutable(t *testing.T, path string) {
	t.Helper()
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(name, windows.READ_CONTROL|windows.WRITE_DAC, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		windows.CloseHandle(h)
		t.Fatal(err)
	}
	acl, _, err := saved.DACL()
	if err != nil {
		windows.CloseHandle(h)
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
			t.Error(err)
		}
		if err := windows.CloseHandle(h); err != nil {
			t.Error(err)
		}
	})
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(D;;GX;;;" + user.User.Sid.String() + ")(A;;FA;;;" + user.User.Sid.String() + ")(A;;FA;;;SY)")
	if err != nil {
		t.Fatal(err)
	}
	denied, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, denied, nil); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsFailureNativeClientClassification(t *testing.T) {
	for _, which := range []string{"stale", "missing", "permission", "process", "unsupported"} {
		t.Run(which, func(t *testing.T) {
			s := windowsSpec(t)
			s.Environment = os.Environ()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			cause, stage, sentinel := WindowsNotFoundFailure, api.ProcessContainment, os.ErrNotExist
			s.Executable = filepath.Join(s.RootLocator, "missing-owned-fixture.exe")
			switch which {
			case "stale":
				s.RootIdentity = windowsSpec(t).RootIdentity
				cause, stage, sentinel = WindowsCwdFailure, api.CwdAcquisition, ErrCwd
			case "permission":
				data, e := os.ReadFile(exe)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(s.Executable, data, 0600); e != nil {
					t.Fatal(e)
				}
				denyOwnedExecutable(t, s.Executable)
				cause, sentinel = WindowsPermissionFailure, os.ErrPermission
			case "process":
				if err := os.WriteFile(s.Executable, []byte("owned invalid executable"), 0600); err != nil {
					t.Fatal(err)
				}
				cause, sentinel = WindowsProcessFailure, ErrWindowsProcess
			case "unsupported":
				s.Executable = `C:unsupported-drive-relative.exe`
				cause, sentinel = WindowsUnsupportedFailure, ErrWindowsUnsupported
			}
			config := WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(api.OutputStream, []byte) {}}
			if _, embedded, e := MachineRoute(); e != nil {
				t.Fatal(e)
			} else if embedded {
				config.Extraction = extractedNativeFixture(t)
				config.Image = config.Extraction.Path()
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			client, started, startErr := StartWindows(ctx, config)
			if client == nil {
				t.Fatalf("lost partial owner: %v", startErr)
			}
			defer client.Stop()
			final, _ := client.Wait(ctx)
			if started.Established || !final.CleanupComplete {
				t.Fatalf("failure lifecycle: start=%+v final=%+v", started, final)
			}
			assertWindowsFailure(t, startErr, cause, stage, false, sentinel)
			assertWindowsFailure(t, final.Err, cause, stage, false, sentinel)
			if strings.Contains(startErr.Error(), s.RootLocator) {
				t.Fatal("native failure exposes cwd")
			}
		})
	}
}

func TestWindowsFailureEnginePreservesOriginalAndActualCleanupFailure(t *testing.T) {
	s := windowsSpec(t)
	s.Terminal = true
	input, parentWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer parentWrite.Close()
	parentRead, output, err := os.Pipe()
	if err != nil {
		input.Close()
		t.Fatal(err)
	}
	defer parentRead.Close()
	defer output.Close()
	nonce, err := FreshNonce()
	if err != nil {
		t.Fatal(err)
	}
	brokerSide, err := NewChannel(input, output, WindowsBroker, Parent, 1, nonce)
	if err != nil {
		t.Fatal(err)
	}
	parentSide, err := NewChannel(parentRead, parentWrite, Parent, WindowsBroker, 1, nonce)
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	p := &userProcess{fault: func(stage string) error {
		if stage == "conpty-created" {
			return ErrWindowsUnsupported
		}
		return nil
	}, closeTerminal: func(h windows.Handle) { <-release; windows.ClosePseudoConsole(h) }}
	done := make(chan int, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		done <- runWindowsEngineOwned(brokerSide, input, output, windows.CurrentProcess(), s, p)
	}()
	var frame Frame
	err = timedIO(parentRead, 10*time.Second, func() error { var e error; frame, e = parentSide.Receive(); return e })
	close(release)
	if err != nil {
		t.Fatal(err)
	}
	if frame.Opcode != Failure {
		t.Fatalf("unexpected frame %v", frame.Opcode)
	}
	failure := decodeWindowsFailure(frame.Payload)
	if !errors.Is(failure, ErrWindowsUnsupported) || !errors.Is(failure, ErrWindowsCleanup) || !errors.Is(failure, context.DeadlineExceeded) {
		t.Fatalf("engine erased primary/cleanup failure: %v", failure)
	}
	joined, ok := failure.(interface{ Unwrap() []error })
	if !ok || len(joined.Unwrap()) != 2 {
		t.Fatalf("missing two stage facts: %v", failure)
	}
	assertWindowsFailure(t, joined.Unwrap()[0], WindowsUnsupportedFailure, api.TerminalCleanup, false, ErrWindowsUnsupported)
	assertWindowsFailure(t, joined.Unwrap()[1], WindowsTimeoutFailure, api.TerminalCleanup, true, ErrWindowsCleanup)
	select {
	case result := <-done:
		if result != 74 {
			t.Fatalf("engine exit %d", result)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("engine cleanup did not join")
	}
	if p.hpc != 0 || p.job != 0 {
		t.Fatal("engine discarded cleanup owner")
	}
}
