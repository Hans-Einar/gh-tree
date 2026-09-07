package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

type ownedParentTree struct {
	Broker, Root, Child, Grandchild uint32
	KillOnClose                     bool
}

func runOwnedParentDeathFixture(root string) int {
	u, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return 80
	}
	h, err := windows.CreateFile(u, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return 81
	}
	id, err := observeHandle(h)
	_ = windows.CloseHandle(h)
	if err != nil {
		return 82
	}
	exe, err := os.Executable()
	if err != nil {
		return 83
	}
	s := StartSpec{ParentID: uint64(os.Getpid()), OperationID: 1, RootLocator: root, RootIdentity: id, ProjectIdentity: id, Executable: exe, Arguments: []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", "hold"}, Environment: os.Environ(), Rows: 24, Columns: 80}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var mu sync.Mutex
	var output bytes.Buffer
	client, start, err := StartWindows(ctx, WindowsConfig{SessionID: 1, Spec: s, Image: exe, Output: func(_ api.OutputStream, data []byte) { mu.Lock(); output.Write(data); mu.Unlock() }})
	if err != nil || !start.Established {
		return 84
	}
	patterns := []*regexp.Regexp{regexp.MustCompile(`"PID":([0-9]+)`), regexp.MustCompile(`OWNED_CHILD ([0-9]+)`), regexp.MustCompile(`OWNED_GRANDCHILD ([0-9]+)`)}
	var ids [3]uint32
	for {
		mu.Lock()
		text := output.String()
		mu.Unlock()
		ready := true
		for index, pattern := range patterns {
			match := pattern.FindStringSubmatch(text)
			if len(match) != 2 {
				ready = false
				break
			}
			number, e := strconv.ParseUint(match[1], 10, 32)
			if e != nil {
				return 85
			}
			ids[index] = uint32(number)
		}
		if ready {
			break
		}
		select {
		case <-ctx.Done():
			return 86
		case <-time.After(time.Millisecond):
		}
	}
	var limit windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	if err = windows.QueryInformationJobObject(client.job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limit)), uint32(unsafe.Sizeof(limit)), nil); err != nil {
		return 87
	}
	fact := ownedParentTree{client.process.ProcessId, ids[0], ids[1], ids[2], limit.BasicLimitInformation.LimitFlags == windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE}
	if err = json.NewEncoder(os.Stdout).Encode(fact); err != nil {
		return 88
	}
	var trigger [1]byte
	if _, err = os.Stdin.Read(trigger[:]); err != nil {
		return 89
	}
	// Deliberately no client.Stop/Wait: process exit closes the last parent Job
	// handle. The observing test holds only process wait handles, never the Job.
	return 0
}

func TestWindowsParentDeathClosesLastOuterJobHandle(t *testing.T) {
	s := windowsSpec(t)
	exe, _ := nativeFixture(t)
	cmd := exec.Command(exe, "--owned-job-parent-fixture", s.RootLocator)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		_ = input.Close()
		if !waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	var tree ownedParentTree
	if err = json.NewDecoder(output).Decode(&tree); err != nil {
		t.Fatal(err)
	}
	if !tree.KillOnClose {
		t.Fatal("outer Job lacks exact kill-on-close policy")
	}
	var handles []windows.Handle
	for _, pid := range []uint32{tree.Broker, tree.Root, tree.Child, tree.Grandchild} {
		if pid == 0 {
			t.Fatal("missing owned process identity")
		}
		// Each PID came from this still-live owned parent and its own tree. A
		// retained handle is acquired before asking that parent to exit.
		h, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			t.Fatal(err)
		}
		handles = append(handles, h)
	}
	defer func() {
		for _, h := range handles {
			if err := windows.CloseHandle(h); err != nil {
				t.Error(err)
			}
		}
	}()
	if _, err = input.Write([]byte{1}); err != nil {
		t.Fatal(err)
	}
	_ = input.Close()
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	waited = true
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for index, h := range handles {
		if _, err = waitProcess(ctx, h); err != nil {
			t.Fatal(fmt.Errorf("owned tree member %d survived parent Job close: %w", index, err))
		}
	}
	t.Logf("parent exit closed last outer Job handle; exact broker/root/child/grandchild waits completed: %+v", tree)
}
