//go:build windows

package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestNativePrivateFileHasProtectedUserACLAndCorrectABI(t *testing.T) {
	ptr := unsafe.Sizeof(uintptr(0))
	if unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{}) != 6*ptr || unsafe.Sizeof(windows.IO_STATUS_BLOCK{}) != 2*ptr {
		t.Fatal("native NT structures do not match this architecture")
	}
	root := t.TempDir()
	observed, err := observeDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	directory, err := acquireDirectory(observed)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.close()
	f, err := directory.createPrivate("private")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sd, err := windows.GetSecurityInfo(windows.Handle(f.Fd()), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	control, _, err := sd.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatal("new private file inherited an unprotected DACL", err)
	}
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		t.Fatal(err)
	}
	user, err := token.GetTokenUser()
	token.Close()
	if err != nil {
		t.Fatal(err)
	}
	text := sd.String()
	if !strings.Contains(text, user.User.Sid.String()) || strings.Contains(text, ";;;WD)") || strings.Contains(text, ";;;AU)") || strings.Contains(text, ";;;BU)") {
		t.Fatal("private file ACL is not limited to the user and system")
	}
	if f, err := directory.createPrivate("private:alternate"); err == nil {
		f.Close()
		t.Fatal("created an alternate data stream")
	}
}

func TestNativeDirectoryJunctionRefusesWithoutSymlinkPrivilege(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	marker := filepath.Join(outside, "marker")
	if err := os.WriteFile(marker, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "redirect")
	if strings.ContainsAny(link+outside, "\"%\r\n") {
		t.Skip("test cmd carrier cannot represent this temporary path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cmd.exe", "/d", "/c", "mklink", "/J", link, outside)
	hideWindow(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("owned junction fixture: %v %s", err, output)
	}
	observed, err := observeDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	directory, err := acquireDirectory(observed)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.close()
	if f, err := directory.openRegular("redirect"); err == nil {
		f.Close()
		t.Fatal("followed the native child junction")
	}
	if f, err := directory.createPrivate("redirect"); err == nil {
		f.Close()
		t.Fatal("replaced the native child junction")
	}
	if value, err := os.ReadFile(marker); err != nil || string(value) != "outside" {
		t.Fatal("junction target changed", err)
	}
}
