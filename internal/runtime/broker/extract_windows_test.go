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
	client, start, err := StartWindows(ctx, WindowsConfig{SessionID: 2, Spec: s, Image: path, Extraction: image, Output: func(api.OutputStream, []byte) {}})
	if client != nil {
		defer client.Stop()
	}
	if err != nil || !start.Established {
		t.Fatalf("extracted start %+v %v", start, err)
	}
	final, err := client.Wait(ctx)
	if err != nil || !final.CleanupComplete {
		t.Fatalf("extracted final %+v %v", final, err)
	}
	if _, err = os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatalf("helper survived outer cleanup: %v", err)
	}
}
