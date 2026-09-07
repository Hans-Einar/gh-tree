//go:build linux || darwin || freebsd

package persistence

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func physicalTemp(t testing.TB) string {
	t.Helper()
	v, err := filepath.EvalSymlinks(t.TempDir())
	return must(t, v, err)
}
func acquiredUnix(t testing.TB, path string) *unixChain {
	t.Helper()
	v, err := unixAcquire(context.Background(), path)
	chain := must(t, v, err)
	t.Cleanup(func() {
		if err := chain.close(); err != nil {
			t.Error(err)
		}
	})
	return chain
}
func TestUnixAcquisitionBoundedReadAndAbsence(t *testing.T) {
	root := physicalTemp(t)
	chain := acquiredUnix(t, root)
	identity, err := chain.parent().observation.directoryIdentity()
	if err != nil {
		t.Fatal(err)
	}
	missing := acquiredUnix(t, filepath.Join(root, "one", "two"))
	if len(missing.remaining) != 2 || !missing.parent().observation.sameObject(chain.parent().observation) {
		t.Fatal("missing anchor not proved")
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Fatal("read acquisition mutated store", err)
	}
	raw := []byte(` {"schemaVersion":1,"unknown":[1,2,3]} `)
	if err := os.WriteFile(filepath.Join(root, "run.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	object, err := unixOpenDocument(chain.parent(), "run.json")
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := unixRead(context.Background(), object)
	if err != nil || !bytes.Equal(got, raw) {
		object.close()
		t.Fatal("native coherent read", err)
	}
	flags, err := unix.FcntlInt(object.file.Fd(), unix.F_GETFD, 0)
	if err != nil || flags&unix.FD_CLOEXEC == 0 {
		object.close()
		t.Fatal("inheritable native descriptor", err)
	}
	if err := object.close(); err != nil {
		t.Fatal(err)
	}
	if _, err := unixOpenDocument(chain.parent(), "../outside"); err == nil {
		t.Fatal("traversal accepted")
	}
	if err := unix.Mkfifo(filepath.Join(root, "fifo"), 0600); err != nil {
		t.Fatal(err)
	}
	if object, err := unixOpenDocument(chain.parent(), "fifo"); err == nil {
		object.close()
		t.Fatal("special object accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := unixAcquire(ctx, root); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled acquisition", err)
	}
	t.Logf("native filesystem=%s identity=%v", chain.fileSystem, identity)
}
func TestUnixMovedObjectAndSubstitutedPathRemainDistinct(t *testing.T) {
	base := physicalTemp(t)
	original := filepath.Join(base, "root")
	moved := filepath.Join(base, "moved")
	outside := filepath.Join(base, "outside")
	for _, path := range []string{original, outside} {
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(original, "marker"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "marker"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	chain := acquiredUnix(t, original)
	if err := chain.revalidate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(original, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, original); err != nil {
		t.Fatal(err)
	}
	if err := chain.revalidate(context.Background()); err == nil {
		t.Fatal("observed path substitution ignored")
	}
	object, err := unixOpenDocument(chain.parent(), "marker")
	if err != nil {
		t.Fatal(err)
	}
	data, _, err := unixRead(context.Background(), object)
	if err != nil || string(data) != "original" {
		object.close()
		t.Fatal("descriptor retargeted via old path", err)
	}
	if err := object.close(); err != nil {
		t.Fatal(err)
	}
	if c, err := unixAcquire(context.Background(), original); err == nil {
		c.close()
		t.Fatal("substituted symlink followed")
	}
	if got, err := os.ReadFile(filepath.Join(outside, "marker")); err != nil || string(got) != "outside" {
		t.Fatal("outside object changed", err)
	}
}
func TestUnixPermanentLockLocalQueueCancellationAndRefusal(t *testing.T) {
	root := physicalTemp(t)
	chain := acquiredUnix(t, root)
	first, err := unixLock(context.Background(), chain.parent(), "state.json", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	identity := first.object.observation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	_, err = unixLock(ctx, chain.parent(), "state.json", time.Second)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		first.close()
		t.Fatal("local lock queue cancellation", err)
	}
	other, err := unixLock(context.Background(), chain.parent(), "run.json", time.Second)
	if err != nil {
		first.close()
		t.Fatal("independent store blocked", err)
	}
	if err := other.close(); err != nil {
		t.Fatal(err)
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
			lock, err := unixLock(context.Background(), chain.parent(), "state.json", time.Second)
			if err != nil {
				t.Error(err)
				return
			}
			if active.Add(1) != 1 {
				t.Error("independent handles bypassed local/kernel locks")
			}
			if !lock.object.observation.sameObject(identity) {
				t.Error("permanent lock changed identity")
			}
			active.Add(-1)
			if err := lock.close(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	unixProcessLocks.Lock()
	left := len(unixProcessLocks.entries)
	unixProcessLocks.Unlock()
	if left != 0 {
		t.Fatal("keyed mutex entries leaked", left)
	}
	if _, err := os.Stat(filepath.Join(root, "state.json.lock")); err != nil {
		t.Fatal("permanent lock unlinked", err)
	}
	if err := os.WriteFile(filepath.Join(root, "outside"), []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("outside", filepath.Join(root, "unsafe.lock")); err != nil {
		t.Fatal(err)
	}
	if lock, err := unixLock(context.Background(), chain.parent(), "unsafe", time.Second); err == nil {
		lock.close()
		t.Fatal("symlink lock followed")
	}
	if err := os.Link(filepath.Join(root, "outside"), filepath.Join(root, "linked.lock")); err != nil {
		t.Fatal(err)
	}
	if lock, err := unixLock(context.Background(), chain.parent(), "linked", time.Second); err == nil {
		lock.close()
		t.Fatal("multiply linked lock accepted")
	}
	if got, err := os.ReadFile(filepath.Join(root, "outside")); err != nil || string(got) != "untouched" {
		t.Fatal("unsafe lock target changed", err)
	}
}
func TestUnixLockProcessFixture(t *testing.T) {
	path := os.Getenv("GH_TREE_PERSISTENCE_UNIX_LOCK_FIXTURE")
	if path == "" {
		return
	}
	chain, err := unixAcquire(context.Background(), path)
	if err != nil {
		os.Exit(71)
	}
	lock, err := unixLock(context.Background(), chain.parent(), "run.json", 50*time.Millisecond)
	if err != nil {
		os.Exit(72)
	}
	if os.Getenv("GH_TREE_PERSISTENCE_UNIX_LOCK_HOLD") == "1" {
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
func TestUnixKilledProcessReleasesKernelLock(t *testing.T) {
	root := physicalTemp(t)
	chain := acquiredUnix(t, root)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	child := exec.CommandContext(ctx, executable, "-test.run=^TestUnixLockProcessFixture$")
	child.Env = append(os.Environ(), "GH_TREE_PERSISTENCE_UNIX_LOCK_FIXTURE="+root, "GH_TREE_PERSISTENCE_UNIX_LOCK_HOLD=1")
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
		t.Fatal("child lock acquisition", line, err)
	}
	if lock, err := unixLock(context.Background(), chain.parent(), "run.json", 30*time.Millisecond); err == nil {
		lock.close()
		t.Fatal("kernel process lock bypassed")
	}
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := child.Wait(); err == nil {
		t.Fatal("expected killed process")
	}
	joined = true
	lock, err := unixLock(context.Background(), chain.parent(), "run.json", time.Second)
	if err != nil {
		t.Fatal("lock remained after process exit", err)
	}
	if err := lock.close(); err != nil {
		t.Fatal(err)
	}
}
