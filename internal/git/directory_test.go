package git

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNativeDirectoryAcquisitionAndExclusivePrivateFiles(t *testing.T) {
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
	f, err := directory.createPrivate("payload")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write([]byte("original")); err != nil {
		t.Fatal(err)
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if f, err := directory.createPrivate("payload"); err == nil {
		f.Close()
		t.Fatal("exclusive creation overwrote an existing child")
	}
	opened, err := directory.openRegular("payload")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(opened)
	opened.Close()
	if err != nil || string(raw) != "original" {
		t.Fatal("original private payload changed", err)
	}
	for _, name := range []string{"../outside", "nested/path", "", "."} {
		if f, err := directory.createPrivate(name); err == nil {
			f.Close()
			t.Fatalf("accepted escaping component %q", name)
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(root, "payload"))
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatal("private file mode", err)
		}
	}
}

func TestNativeDirectoryRejectsReplacementBeforeAcquisition(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	observed, err := observeDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(root, filepath.Join(parent, "retained")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if directory, err := acquireDirectory(observed); err == nil {
		directory.close()
		t.Fatal("acquired a replacement through stale physical identity")
	}
}

func TestNativeDirectoryChildRedirectRefuses(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "redirect")); err != nil {
		t.Skipf("native symlink creation unavailable: %v", err)
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
		t.Fatal("followed a redirected native child")
	}
	if f, err := directory.createPrivate("redirect"); err == nil {
		f.Close()
		t.Fatal("exclusive create followed a redirected native child")
	}
	bytes, err := os.ReadFile(outside)
	if err != nil || string(bytes) != "outside" {
		t.Fatal("outside data changed", err)
	}
}
