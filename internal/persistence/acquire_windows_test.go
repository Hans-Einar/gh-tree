package persistence

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func acquiredWindows(t testing.TB, path string) *winChain {
	t.Helper()
	v, err := winAcquire(context.Background(), path)
	result := must(t, v, err)
	t.Cleanup(func() {
		if err := result.close(); err != nil {
			t.Error(err)
		}
	})
	return result
}
func TestWindowsNativeAcquisitionReadAndAbsence(t *testing.T) {
	root := t.TempDir()
	existing := acquiredWindows(t, root)
	if len(existing.remaining) != 0 {
		t.Fatal("existing parent reported absent")
	}
	identity, err := existing.parent().observation.directoryIdentity()
	if err != nil || identity.FileID() == ([16]byte{}) {
		t.Fatal("native identity unavailable", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gh-tree")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	missing := acquiredWindows(t, filepath.Join(root, "one", "two"))
	if len(missing.remaining) != 2 || !missing.parent().observation.sameObject(existing.parent().observation) {
		t.Fatal("missing ancestor not bound to actual parent")
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Fatal("acquisition wrote persistent state", err)
	}
	raw := []byte(` {"schemaVersion":1,"unknown":[1,2,3]} `)
	if err := os.WriteFile(filepath.Join(root, "run.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	object, err := winOpenDocument(existing.parent(), "run.json")
	if err != nil {
		t.Fatal(err)
	}
	got, observation, err := winRead(context.Background(), object)
	if err != nil || !bytes.Equal(got, raw) || observation.size() != uint64(len(raw)) {
		t.Fatal("bounded native read", err)
	}
	if err := object.close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(root, root+"-moved"); err == nil {
		t.Fatal("guard permitted ordinary directory movement")
	}
	if _, err := winOpenDocument(existing.parent(), `..\outside`); err == nil {
		t.Fatal("relative traversal allowed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := winAcquire(ctx, root); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled acquisition", err)
	}
	if _, err := winAcquire(context.Background(), `\\server\share\dir`); err == nil {
		t.Fatal("network path accepted")
	}
	t.Logf("native FileIdInfo size=%d file offset=%d identity=%v", unsafe.Sizeof(winFileIDInfo{}), unsafe.Offsetof(winFileIDInfo{}.File), identity)
}

// Make an owned directory into a junction using the actual FSCTL. This needs
// no symlink privilege, PowerShell process, or changes outside the temp fixture.
func setTestJunction(path, target string) error {
	h, err := windows.CreateFile(windows.StringToUTF16Ptr(path), windows.GENERIC_WRITE, winShareAll, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	sub, err := windows.UTF16FromString(`\??\` + target)
	if err != nil {
		return err
	}
	print, err := windows.UTF16FromString(target)
	if err != nil {
		return err
	}
	data := make([]byte, 16+2*(len(sub)+len(print)))
	binary.LittleEndian.PutUint32(data, windows.IO_REPARSE_TAG_MOUNT_POINT)
	binary.LittleEndian.PutUint16(data[4:], uint16(len(data)-8))
	binary.LittleEndian.PutUint16(data[10:], uint16((len(sub)-1)*2))
	binary.LittleEndian.PutUint16(data[12:], uint16(len(sub)*2))
	binary.LittleEndian.PutUint16(data[14:], uint16((len(print)-1)*2))
	for i, c := range append(sub, print...) {
		binary.LittleEndian.PutUint16(data[16+2*i:], c)
	}
	var returned uint32
	return windows.DeviceIoControl(h, windows.FSCTL_SET_REPARSE_POINT, &data[0], uint32(len(data)), nil, 0, &returned, nil)
}
func TestWindowsActualDataGuardsAndConvertedParentRefusal(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, ".gh-tree")
	outside := filepath.Join(root, "outside")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0700); err != nil {
		t.Fatal(err)
	}
	guard, err := winAcquire(context.Background(), child)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := winLock(context.Background(), guard.parent(), "run.json", time.Second)
	if err != nil {
		guard.close()
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(child, "run.json.lock")); err == nil {
		lock.close()
		guard.close()
		t.Fatal("data-read lock child could be deleted")
	}
	if err := setTestJunction(child, outside); err == nil {
		lock.close()
		guard.close()
		t.Fatal("nonempty parent converted")
	}
	if err := lock.close(); err != nil {
		guard.close()
		t.Fatal(err)
	}
	// An uncooperative actor can empty the directory after the request releases
	// its lock. Retained directory handles alone do not prevent this conversion.
	if err := os.Remove(filepath.Join(child, "run.json.lock")); err != nil {
		guard.close()
		t.Fatal(err)
	}
	if err := setTestJunction(child, outside); err != nil {
		guard.close()
		t.Fatal("positive junction conversion control", err)
	}
	object, err := winOpen(guard.parent().handle(), "must-not-escape", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, winShareAll, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE)
	if err == nil {
		object.close()
		guard.close()
		t.Fatal("converted retained parent redirected relative creation")
	}
	if err := guard.close(); err != nil {
		t.Fatal(err)
	}
	if guard, err := winAcquire(context.Background(), child); err == nil {
		guard.close()
		t.Fatal("followed converted parent")
	}
	if err := os.Remove(child); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("reparse redirection touched outside fixture", err)
	}
}

func TestWindowsPermanentLocksCancellationAndSeparateStores(t *testing.T) {
	root := t.TempDir()
	chain := acquiredWindows(t, root)
	first, err := winLock(context.Background(), chain.parent(), "state.json", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	identity := first.object.observation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	_, err = winLock(ctx, chain.parent(), "state.json", time.Second)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		first.close()
		t.Fatal("lock cancellation", err)
	}
	other, err := winLock(context.Background(), chain.parent(), "run.json", time.Second)
	if err != nil {
		first.close()
		t.Fatal("unrelated store blocked", err)
	}
	if err := other.close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "state.json.lock")); err == nil {
		first.close()
		t.Fatal("held lock object unlinked")
	}
	if err := first.close(); err != nil {
		t.Fatal(err)
	}
	var active atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock, err := winLock(context.Background(), chain.parent(), "state.json", time.Second)
			if err != nil {
				t.Error(err)
				return
			}
			if active.Add(1) != 1 {
				t.Error("independent handles bypassed permanent lock")
			}
			if !lock.object.observation.sameObject(identity) {
				t.Error("lock identity changed")
			}
			active.Add(-1)
			if err := lock.close(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if _, err := os.Stat(filepath.Join(root, "state.json.lock")); err != nil {
		t.Fatal("permanent lock removed", err)
	}
}

func TestWindowsLockProcessFixture(t *testing.T) {
	path := os.Getenv("GH_TREE_PERSISTENCE_LOCK_FIXTURE")
	if path == "" {
		return
	}
	chain, err := winAcquire(context.Background(), path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "child acquisition:", err)
		os.Exit(71)
	}
	lock, err := winLock(context.Background(), chain.parent(), "run.json", time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "child lock:", err)
		os.Exit(72)
	}
	if os.Getenv("GH_TREE_PERSISTENCE_LOCK_HOLD") == "1" {
		fmt.Println("locked")
		time.Sleep(10 * time.Second)
		os.Exit(75)
	}
	if err := lock.close(); err != nil {
		os.Exit(73)
	}
	if err := chain.close(); err != nil {
		os.Exit(74)
	}
	os.Exit(0)
}
func TestWindowsLockProcessExitReleasesKernelOwnership(t *testing.T) {
	root := t.TempDir()
	chain := acquiredWindows(t, root)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	child := exec.CommandContext(ctx, executable, "-test.run=^TestWindowsLockProcessFixture$")
	child.Env = append(os.Environ(), "GH_TREE_PERSISTENCE_LOCK_FIXTURE="+root, "GH_TREE_PERSISTENCE_LOCK_HOLD=1")
	var childErrors bytes.Buffer
	child.Stderr = &childErrors
	stdout, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	joined := false
	defer func() {
		if !joined {
			child.Process.Kill()
			child.Wait()
		}
	}()
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || line != "locked\n" {
		waitErr := child.Wait()
		joined = true
		t.Fatal("child did not acquire native lock", line, err, waitErr, childErrors.String())
	}
	if lock, err := winLock(context.Background(), chain.parent(), "run.json", 30*time.Millisecond); err == nil {
		lock.close()
		t.Fatal("process lock bypassed")
	}
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := child.Wait(); err == nil {
		t.Fatal("crash fixture unexpectedly exited normally")
	}
	joined = true
	lock, err := winLock(context.Background(), chain.parent(), "run.json", time.Second)
	if err != nil {
		t.Fatal("kernel did not release dead process lock", err)
	}
	if err := lock.close(); err != nil {
		t.Fatal(err)
	}
}
