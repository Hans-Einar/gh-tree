package broker

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func extractedFixture(t *testing.T) *WindowsImage {
	t.Helper()
	return extractedNativeFixture(t)
}

func TestWindowsExtractionPartialFailuresAndTamper(t *testing.T) {
	path, machine := nativeFixture(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	for _, stage := range []string{"temporary-parent-acquired", "helper-directory-created", "helper-image-created", "helper-image-flushed", "helper-writer-closed", "helper-read-guard-acquired", "helper-verified"} {
		t.Run(stage, func(t *testing.T) {
			injected := errors.New("owned extraction fixture failure")
			image, err := extractWindowsImage(data, machine, digest, ProtocolVersion, func(actual string, _ *WindowsImage) error {
				if actual == stage {
					return injected
				}
				return nil
			})
			if image == nil || !errors.Is(err, injected) {
				t.Fatalf("partial owner lost: image=%v %v", image != nil, err)
			}
			path := image.Path()
			if err = image.Cleanup(); err != nil {
				t.Fatal(err)
			}
			if path != "" {
				if _, err = os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
					t.Fatalf("partial extraction remains: %v", err)
				}
			}
		})
	}
	t.Run("same-object-byte-tamper", func(t *testing.T) {
		image, err := extractWindowsImage(data, machine, digest, ProtocolVersion, func(stage string, image *WindowsImage) error {
			if stage == "helper-writer-closed" {
				return os.WriteFile(image.Path(), []byte("tampered owned image"), 0600)
			}
			return nil
		})
		if image == nil || err == nil {
			t.Fatalf("tampered image accepted: image=%v %v", image != nil, err)
		}
		if err = image.Cleanup(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestWindowsExtractionIdentityACLAndInterlock(t *testing.T) {
	image := extractedFixture(t)
	p, err := windows.UTF16PtrFromString(image.Path())
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.CreateFile(p, windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err == nil {
		windows.CloseHandle(h)
		t.Fatal("readonly image guard allowed write access")
	}
	if err = os.Remove(image.Path()); err == nil {
		t.Fatal("readonly image guard allowed deletion")
	}
	for _, held := range []windows.Handle{image.directory, windows.Handle(image.guard.Fd())} {
		sd, err := windows.GetSecurityInfo(held, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		control, _, err := sd.Control()
		if err != nil {
			t.Fatal(err)
		}
		acl, defaulted, err := sd.DACL()
		if err != nil {
			t.Fatal(err)
		}
		user, err := windows.GetCurrentProcessToken().GetTokenUser()
		if err != nil {
			t.Fatal(err)
		}
		if defaulted || acl == nil || acl.AceCount != 2 || control&windows.SE_DACL_PROTECTED == 0 {
			t.Fatalf("incorrect protected ACL: %s", sd.String())
		}
		system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
		if err != nil {
			t.Fatal(err)
		}
		seenUser, seenSystem := false, false
		for n := uint32(0); n < 2; n++ {
			var ace *windows.ACCESS_ALLOWED_ACE
			if err = windows.GetAce(acl, n, &ace); err != nil {
				t.Fatal(err)
			}
			if ace.Header.AceType != 0 || ace.Header.AceFlags != 0 || ace.Mask != 0x1f01ff {
				t.Fatal("non-full-access or inherited/non-allow ACE")
			}
			sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
			if sid.Equals(user.User.Sid) {
				seenUser = true
			} else if sid.Equals(system) {
				seenSystem = true
			} else {
				t.Fatalf("unexpected trustee %s", sid.String())
			}
		}
		if !seenUser || !seenSystem {
			t.Fatalf("missing exact owner trustees: %s", sd.String())
		}
	}
	path := image.Path()
	if err = image.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatalf("owned directory remains: %v", err)
	}
}

func TestWindowsExtractionPreservesUnexpectedEntries(t *testing.T) {
	image := extractedFixture(t)
	added := filepath.Join(filepath.Dir(image.Path()), "owned-test-unexpected")
	if err := os.WriteFile(added, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := image.Cleanup(); err == nil {
		t.Fatal("unexpected entry silently removed")
	}
	if bytes, err := os.ReadFile(added); err != nil || string(bytes) != "preserve" {
		t.Fatalf("entry lost: %q %v", bytes, err)
	}
	if err := os.Remove(added); err != nil {
		t.Fatal(err)
	}
	if err := image.Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsExtractionRejectsReplacedImage(t *testing.T) {
	image := extractedFixture(t)
	path := image.Path()
	original := path + ".original"
	if err := closeFile(&image.guard); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, original); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := image.Cleanup(); err == nil {
		t.Fatal("replacement identity accepted")
	}
	if bytes, err := os.ReadFile(path); err != nil || string(bytes) != "replacement" {
		t.Fatalf("replacement was lost: %q %v", bytes, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(original, path); err != nil {
		t.Fatal(err)
	}
	if err := image.Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsExtractedBrokerLifecycle(t *testing.T) {
	image := extractedFixture(t)
	s := windowsSpec(t)
	s.Environment = os.Environ()
	s.Executable = image.Path()
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "root-first"}
	path := image.Path()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	proof := &extractedLifecycleProof{}
	t.Cleanup(func() { proof.close(t) })
	client, start, err := StartWindows(ctx, WindowsConfig{SessionID: 2, Spec: s, Image: path, Extraction: image, Output: func(api.OutputStream, []byte) {}, hook: proof.capture(t)})
	if client != nil {
		defer joinExtractedClient(t, client)
	}
	if err != nil || !start.Established {
		t.Fatalf("extracted start %+v %v", start, err)
	}
	final, err := client.Wait(ctx)
	logExtractionNativeCauses(t, err)
	assertExtractedLifecycleGone(t, client, image, path, final, proof)
	if !final.Established || !final.RootExited || final.ExitCode != 0 || !final.Quiescent {
		t.Fatalf("natural extracted lifecycle lost facts: %+v", final)
	}
	// The owner preserves a failed attempt after successful repair. Accept only
	// the exact native refusal proved by the mapped-image control below, after
	// independently proving every final native/file postcondition above.
	if err != nil && !repairedExtractionRefusal(err) {
		t.Fatalf("unexpected extracted lifecycle diagnostic: %+v %v", final, err)
	}
}

func repairedExtractionRefusal(err error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		children := joined.Unwrap()
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !repairedExtractionRefusal(child) {
				return false
			}
		}
		return true
	}
	failure, ok := err.(*WindowsFailure)
	if !ok || failure.Cause != WindowsPermissionFailure || failure.Stage != api.HelperExtraction || !failure.Cleanup {
		return false
	}
	var nativeOnly func(error) bool
	nativeOnly = func(err error) bool {
		if err == windows.STATUS_CANNOT_DELETE {
			return true
		}
		if joined, ok := err.(interface{ Unwrap() []error }); ok {
			children := joined.Unwrap()
			if len(children) == 0 {
				return false
			}
			for _, child := range children {
				if !nativeOnly(child) {
					return false
				}
			}
			return true
		}
		if wrapped, ok := err.(interface{ Unwrap() error }); ok {
			return nativeOnly(wrapped.Unwrap())
		}
		return false
	}
	return nativeOnly(failure.local)
}

func TestWindowsRepairedExtractionDiagnosticIsNarrow(t *testing.T) {
	known := windowsCleanupFailure(windows.STATUS_CANNOT_DELETE, api.HelperExtraction)
	if !repairedExtractionRefusal(known) || !repairedExtractionRefusal(errors.Join(known)) {
		t.Fatal("exact historical extraction cause refused")
	}
	for _, err := range []error{
		nil, windows.STATUS_CANNOT_DELETE,
		windowsFailureAt(windows.STATUS_CANNOT_DELETE, api.HelperExtraction),
		windowsCleanupFailure(windows.STATUS_CANNOT_DELETE, api.TerminalCleanup),
		windowsCleanupFailure(windows.ERROR_ACCESS_DENIED, api.HelperExtraction),
		windowsCleanupFailure(windows.STATUS_ACCESS_DENIED, api.HelperExtraction),
		windowsCleanupFailure(context.DeadlineExceeded, api.HelperExtraction),
		errors.Join(known, windowsFailureAt(ErrWindowsProcess, api.UserProcessWait)),
		windowsCleanupFailure(errors.Join(windows.STATUS_CANNOT_DELETE, context.Canceled), api.HelperExtraction),
	} {
		if repairedExtractionRefusal(err) {
			t.Fatalf("unrelated or mixed failure accepted: %v", err)
		}
	}
}

// Separate proof duplicates retain the exact broker/outer Job identities even
// after the product closes its own handles. They are never signal authority.
type extractedLifecycleProof struct{ process, job windows.Handle }

func (p *extractedLifecycleProof) capture(t *testing.T) func(string, *WindowsClient) {
	t.Helper()
	return func(stage string, client *WindowsClient) {
		if stage != "broker-before-resume" {
			return
		}
		if err := windows.DuplicateHandle(windows.CurrentProcess(), client.process.Process, windows.CurrentProcess(), &p.process, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
			t.Error(err)
		}
		if err := windows.DuplicateHandle(windows.CurrentProcess(), client.job, windows.CurrentProcess(), &p.job, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
			t.Error(err)
		}
	}
}

func joinExtractedClient(t *testing.T, client *WindowsClient) {
	t.Helper()
	client.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	final, err := client.Wait(ctx)
	if !final.CleanupComplete {
		t.Errorf("fixture cleanup did not join: %+v %v", final, err)
	}
}

func (p *extractedLifecycleProof) close(t *testing.T) {
	t.Helper()
	if err := closeHandle(&p.process); err != nil {
		t.Error(err)
	}
	if err := closeHandle(&p.job); err != nil {
		t.Error(err)
	}
}

func (p *extractedLifecycleProof) assertQuiescent(t *testing.T) {
	t.Helper()
	if p.process == 0 || p.job == 0 {
		t.Fatal("missing retained broker/outer proof")
	}
	state, err := windows.WaitForSingleObject(p.process, 0)
	if err != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatalf("retained broker not exited: state=%d err=%v", state, err)
	}
	active, err := activeProcesses(p.job)
	if err != nil || active != 0 {
		t.Fatalf("retained outer Job not empty: active=%d err=%v", active, err)
	}
}

func assertExtractedLifecycleGone(t *testing.T, client *WindowsClient, image *WindowsImage, path string, final WindowsFact, proof *extractedLifecycleProof) {
	t.Helper()
	if !final.CleanupComplete || len(final.Residuals) != 0 {
		t.Fatalf("cleanup has residuals: %+v", final)
	}
	proof.assertQuiescent(t)
	select {
	case <-client.done:
	default:
		t.Fatal("client owner has not joined")
	}
	if client.process.Process != 0 || client.process.Thread != 0 || client.job != 0 || client.read != nil || client.write != nil {
		t.Fatal("client retained native resources after CleanupComplete")
	}
	for _, h := range client.childHandles {
		if h != 0 {
			t.Fatal("client retained inherited resource")
		}
	}
	for _, file := range client.outputFiles {
		if file != nil {
			t.Fatal("client retained output reader")
		}
	}
	if !image.removed || image.imageCreated || image.guard != nil || image.writer != nil || image.directory != 0 || image.parent != 0 || len(image.chain) != 0 {
		t.Fatal("image retained owned extraction resources after CleanupComplete")
	}
	for _, item := range []string{path, filepath.Dir(path)} {
		if _, err := os.Stat(item); !os.IsNotExist(err) {
			t.Fatalf("extracted object remains after cleanup: %q %v", item, err)
		}
	}
	t.Log("proved broker exited, outer Job zero, client/native extraction handles released, exact image and directory absent")
}

func logExtractionNativeCauses(t *testing.T, err error) {
	t.Helper()
	var visit func(error, int)
	visit = func(e error, depth int) {
		if e == nil || depth > 16 {
			return
		}
		switch cause := e.(type) {
		case *WindowsFailure:
			t.Logf("extraction diagnostic cause=%d stage=%d cleanup=%v", cause.Cause, cause.Stage, cause.Cleanup)
		case windows.NTStatus:
			t.Logf("extraction NTSTATUS=0x%08x mapped Win32=%d: %v", uint32(cause), uint32(cause.Errno()), cause)
		default:
			t.Logf("extraction error %T: %v", e, e)
		}
		if joined, ok := e.(interface{ Unwrap() []error }); ok {
			for _, child := range joined.Unwrap() {
				visit(child, depth+1)
			}
		} else if wrapped, ok := e.(interface{ Unwrap() error }); ok {
			visit(wrapped.Unwrap(), depth+1)
		}
	}
	visit(err, 0)
}

func TestWindowsExtractionMappedImageCleanupRepair(t *testing.T) {
	image := extractedFixture(t)
	path := image.Path()
	// SEC_IMAGE is the documented executable image mapping profile. Keep an
	// owned mapped view, without executing it or writing process memory.
	const secImage = 0x01000000
	section, err := windows.CreateFileMapping(windows.Handle(image.guard.Fd()), nil, windows.PAGE_READONLY|secImage, 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := closeHandle(&section); err != nil {
			t.Error(err)
		}
	})
	view, err := windows.MapViewOfFile(section, windows.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if view != 0 {
			if err := windows.UnmapViewOfFile(view); err != nil {
				t.Error(err)
			}
		}
	})
	s := windowsSpec(t)
	s.Environment, s.Executable = os.Environ(), path
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "root-first"}
	proof := &extractedLifecycleProof{}
	t.Cleanup(func() { proof.close(t) })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, start, err := StartWindows(ctx, WindowsConfig{SessionID: 3, Spec: s, Image: path, Extraction: image, Output: func(api.OutputStream, []byte) {}, hook: proof.capture(t)})
	if client != nil {
		defer func() {
			// Release this test's extra refusal condition before joining the
			// ordinary owner, including when a diagnostic assertion fails.
			if view != 0 {
				if err := windows.UnmapViewOfFile(view); err != nil {
					t.Error(err)
				} else {
					view = 0
				}
			}
			if err := closeHandle(&section); err != nil {
				t.Error(err)
			}
			joinExtractedClient(t, client)
		}()
	}
	if err != nil || !start.Established {
		t.Fatalf("mapped-image fixture start %+v %v", start, err)
	}
	var pending WindowsFact
	for {
		pending, err = client.NextFact(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if pending.Err != nil {
			break
		}
		if pending.CleanupComplete {
			t.Fatal("live image mapping did not exercise cleanup refusal")
		}
	}
	logExtractionNativeCauses(t, pending.Err)
	if pending.CleanupComplete || !pending.RootExited || !pending.Quiescent || !errors.Is(pending.Err, windows.STATUS_CANNOT_DELETE) || !errors.Is(pending.Err, os.ErrPermission) || !errors.Is(pending.Err, ErrWindowsCleanup) {
		t.Fatalf("mapped-image refusal lost native cause/ownership: %+v", pending)
	}
	proof.assertQuiescent(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("refused cleanup lost exact mapped image: %v", err)
	}
	if err := windows.UnmapViewOfFile(view); err != nil {
		t.Fatal(err)
	}
	view = 0
	if err := closeHandle(&section); err != nil {
		t.Fatal(err)
	}
	final, err := client.Wait(ctx)
	logExtractionNativeCauses(t, err)
	assertExtractedLifecycleGone(t, client, image, path, final, proof)
	if !repairedExtractionRefusal(err) || !errors.Is(err, windows.STATUS_CANNOT_DELETE) || !errors.Is(err, ErrWindowsCleanup) {
		t.Fatalf("repaired cleanup erased historical cause: %+v %v", final, err)
	}
}
