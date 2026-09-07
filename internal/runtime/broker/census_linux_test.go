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
	path := fmt.Sprintf("/proc/%d/stat", cmd.Process.Pid)
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
	t.Log("live positive; exact exit0; native retained read ESRCH; absent census record; descriptor closed")
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
