package broker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func extractedFixture(t *testing.T) *WindowsImage {
	t.Helper()
	return extractedNativeFixture(t)
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
		text := sd.String()
		user, err := windows.GetCurrentProcessToken().GetTokenUser()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(text, "D:P") || !strings.Contains(text, "(A;;FA;;;SY)") || !strings.Contains(text, "(A;;FA;;;"+user.User.Sid.String()+")") {
			t.Fatalf("incorrect protected ACL: %s", text)
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
