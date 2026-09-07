package broker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestNativeLinuxCensusRetainedStatAfterExactExit(t *testing.T) {
	// The pipe owns the fixture lifetime: no sleep, guessed PID signal or
	// scheduling window is needed to make the opened proc object lose its task.
	input, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	defer writer.Close()
	cmd := exec.Command("/bin/cat")
	cmd.Stdin = input
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			writer.Close()
			if err := cmd.Wait(); err != nil {
				t.Error("fixture cleanup", err)
			}
		}
	}()
	procPath := fmt.Sprintf("/proc/%d", cmd.Process.Pid)
	path := procPath + "/stat"
	directory, err := os.Open(procPath)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	live, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	p, present, err := readLinuxStat(cmd.Process.Pid, live)
	if err != nil || !present || p.pid != cmd.Process.Pid || !p.live {
		t.Fatal("live positive control", p, present, err)
	}
	retained, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer retained.Close()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	waited = true
	if err != nil {
		t.Fatal("exact fixture wait", err)
	}
	if f, err := os.Open(path); !errors.Is(err, syscall.ENOENT) || !linuxProcEntryGone(err) {
		if f != nil {
			f.Close()
		}
		t.Fatal("fresh proc path after exact wait", err)
	}
	// Reuse the reviewer's retained-directory control to make the native
	// acquisition race deterministic; it is not a synthetic read failure.
	fd, openErr := unix.Openat(int(directory.Fd()), "stat", unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if fd >= 0 {
		unix.Close(fd)
	}
	if !errors.Is(openErr, syscall.ESRCH) {
		t.Fatal("retained proc directory acquisition", fd, openErr)
	}
	if !linuxProcEntryGone(&os.PathError{Op: "open", Path: path, Err: openErr}) {
		t.Fatal("native acquisition disappearance was not recognized", openErr)
	}
	// First establish the native errno on this retained object. Failed seq_file
	// reads do not advance its offset, so the actual reader sees the same errno.
	var probe [1]byte
	if n, err := retained.Read(probe[:]); n != 0 || !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("retained proc read: bytes=%d error=%v", n, err)
	}
	p, present, err = readLinuxStat(cmd.Process.Pid, retained)
	if err != nil || present || p != (processFact{}) {
		t.Fatal("exited retained proc object was not omitted", p, present, err)
	}
	if _, err := retained.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatal("retained descriptor not closed", err)
	}
	t.Log("live positive; exact exit0; fresh open ENOENT; retained-directory open ESRCH; retained read ESRCH; absent census record; descriptor closed")
}

func TestLinuxCensusAcquisitionClassifiesOnlyNativeDisappearance(t *testing.T) {
	for _, nativeErr := range []error{nil, syscall.ENOENT, syscall.ESRCH, syscall.EACCES, syscall.EPERM, syscall.EIO, syscall.EBADF, syscall.ENOTDIR, ErrCensus, errors.Join(syscall.ESRCH, syscall.EACCES)} {
		want := nativeErr == syscall.ENOENT || nativeErr == syscall.ESRCH
		for _, err := range []error{nativeErr, &os.PathError{Op: "open", Path: "/proc/owned/stat", Err: nativeErr}} {
			if got := linuxProcEntryGone(err); got != want {
				t.Fatalf("acquisition error %v: gone=%v, want %v", err, got, want)
			}
		}
	}
}

type linuxStatTestReader struct {
	io.Reader
	readErr, closeErr error
	closes            int
}

func (r *linuxStatTestReader) Read(p []byte) (int, error) {
	if r.readErr != nil {
		return 0, r.readErr
	}
	return r.Reader.Read(p)
}

func (r *linuxStatTestReader) Close() error { r.closes++; return r.closeErr }

func TestLinuxCensusReadAndCloseFailuresRemainErrors(t *testing.T) {
	for _, readErr := range []error{syscall.ENOENT, syscall.ESRCH, syscall.EACCES, syscall.EIO, syscall.EBADF} {
		for _, closeErr := range []error{nil, syscall.EIO} {
			t.Run(fmt.Sprintf("read-%v-close-%v", readErr, closeErr), func(t *testing.T) {
				f := &linuxStatTestReader{readErr: &os.PathError{Op: "read", Path: "/proc/owned/stat", Err: readErr}, closeErr: closeErr}
				p, present, err := readLinuxStat(1, f)
				gone := readErr == syscall.ENOENT || readErr == syscall.ESRCH
				if p != (processFact{}) || present || f.closes != 1 {
					t.Fatal("failed read produced record or missed close", p, present, f.closes)
				}
				if gone && closeErr == nil {
					if err != nil {
						t.Fatal("exact native disappearance refused", err)
					}
				} else if !errors.Is(err, readErr) || (closeErr != nil && !errors.Is(err, closeErr)) {
					t.Fatal("independent read/close error was erased", err)
				}
			})
		}
	}
	for name, data := range map[string][]byte{"empty": nil, "malformed": []byte("not stat"), "oversized": bytes.Repeat([]byte{'x'}, 8193)} {
		t.Run(name, func(t *testing.T) {
			f := &linuxStatTestReader{Reader: bytes.NewReader(data)}
			if _, present, err := readLinuxStat(1, f); present || !errors.Is(err, ErrCensus) || f.closes != 1 {
				t.Fatal("invalid census record accepted", present, err, f.closes)
			}
		})
	}
	valid := "1 (owned fixture) S 0 1 1 " + strings.Repeat("0 ", 15) + "100"
	f := &linuxStatTestReader{Reader: strings.NewReader(valid), closeErr: syscall.EIO}
	if _, present, err := readLinuxStat(1, f); present || !errors.Is(err, syscall.EIO) || f.closes != 1 {
		t.Fatal("valid bytes hid close failure", present, err, f.closes)
	}
}
