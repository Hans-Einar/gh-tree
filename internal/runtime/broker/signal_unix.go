//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const signalPrivateMarker = "--gh-tree-runtime-signal-v1"

type groupSignal byte

const (
	termGroup groupSignal = 1
	stopGroup groupSignal = 2
	killGroup groupSignal = 3
)

type signalPreparation struct {
	parent, sid, group uint64
	signal             groupSignal
}

func (p signalPreparation) encode() []byte {
	data := make([]byte, 25)
	binary.BigEndian.PutUint64(data, p.parent)
	binary.BigEndian.PutUint64(data[8:], p.sid)
	binary.BigEndian.PutUint64(data[16:], p.group)
	data[24] = byte(p.signal)
	return data
}
func decodePreparation(data []byte) (signalPreparation, error) {
	if len(data) != 25 {
		return signalPreparation{}, ErrProtocol
	}
	p := signalPreparation{binary.BigEndian.Uint64(data), binary.BigEndian.Uint64(data[8:]), binary.BigEndian.Uint64(data[16:]), groupSignal(data[24])}
	if p.parent <= 1 || p.parent != p.sid || p.group == 0 || p.group == p.sid || p.signal < termGroup || p.signal > killGroup {
		return signalPreparation{}, ErrProtocol
	}
	return p, nil
}

// RunSignalHelper is entered only by the fixed private dispatch marker. The
// working authority is separately inherited anonymous pipes plus actual kernel
// parent/SID/group membership. No target or bearer secret occurs in argv/env.
func RunSignalHelper() int {
	control, err := inheritedPipe(3, false)
	if err != nil {
		return 120
	}
	defer control.Close()
	reply, err := inheritedPipe(4, true)
	if err != nil {
		return 120
	}
	defer reply.Close()
	channel, frame, err := AcceptChannel(control, reply, UnixSignalHelper, UnixSupervisor)
	if err != nil || frame.Opcode != Prepare {
		return 121
	}
	prepare, err := decodePreparation(frame.Payload)
	if err != nil {
		return 121
	}
	sid, err := unix.Getsid(0)
	if err != nil {
		return 121
	}
	if uint64(os.Getppid()) != prepare.parent || uint64(sid) != prepare.sid || uint64(unix.Getpgrp()) != prepare.group {
		return 121
	}
	signal.Ignore(syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	if err := channel.Send(Joined, prepare.encode()); err != nil {
		return 122
	}
	if err := control.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return 122
	}
	commit, err := channel.Receive()
	if errors.Is(err, io.EOF) {
		return 0
	}
	if err != nil {
		return 123
	} // EOF before Commit never signals.
	if commit.Opcode != Commit || len(commit.Payload) != 0 {
		return 123
	}
	var sig syscall.Signal
	switch prepare.signal {
	case termGroup:
		sig = syscall.SIGTERM
	case stopGroup:
		sig = syscall.SIGSTOP
	case killGroup:
		sig = syscall.SIGKILL
	default:
		return 123
	}
	// This is the sole deliberate zero-target signal operation. The helper's
	// acquired membership pins its group; it never changes group/session after
	// exec and never signals a remembered numeric census target.
	if err := unix.Kill(0, sig); err != nil {
		return 124
	}
	if prepare.signal == stopGroup || prepare.signal == killGroup {
		for {
			time.Sleep(time.Hour)
		}
	} // never race a pending STOP/KILL with voluntary process exit
	return 0
}

type acquiredSignal struct {
	cmd            *exec.Cmd
	control, reply *os.File
	channel        *Channel
	prepare        signalPreparation
	done           chan struct{}
	waitErr        error
	committed      bool
	closeErr       error
	closed         bool
}

func acquireSignalGroup(ctx context.Context, executable string, session uint64, group int, request groupSignal) (*acquiredSignal, error) {
	if !filepath.IsAbs(executable) {
		return nil, ErrProtocol
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sid, err := unix.Getsid(0)
	if err != nil {
		return nil, err
	}
	if sid != os.Getpid() || unix.Getpgrp() != sid || group <= 0 || group == sid || request < termGroup || request > killGroup {
		return nil, ErrProtocol
	}
	controlRead, controlWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	replyRead, replyWrite, err := os.Pipe()
	if err != nil {
		return nil, errors.Join(err, controlRead.Close(), controlWrite.Close())
	}
	closeAll := func() error {
		return errors.Join(controlRead.Close(), controlWrite.Close(), replyRead.Close(), replyWrite.Close())
	}
	nonce, err := FreshNonce()
	if err != nil {
		return nil, errors.Join(err, closeAll())
	}
	channel, err := NewChannel(replyRead, controlWrite, UnixSupervisor, UnixSignalHelper, session, nonce)
	if err != nil {
		return nil, errors.Join(err, closeAll())
	}
	cmd := exec.Command(executable, signalPrivateMarker)
	cmd.ExtraFiles = []*os.File{controlRead, replyWrite}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: group}
	if err := cmd.Start(); err != nil {
		return nil, errors.Join(err, closeAll())
	}
	a := &acquiredSignal{cmd: cmd, control: controlWrite, reply: replyRead, channel: channel, prepare: signalPreparation{uint64(sid), uint64(sid), uint64(group), request}, done: make(chan struct{})}
	go func() { a.waitErr = cmd.Wait(); close(a.done) }() // exactly one waiter, retained until done joins
	if err := errors.Join(controlRead.Close(), replyWrite.Close()); err != nil {
		return a, err
	}
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := errors.Join(controlWrite.SetWriteDeadline(deadline), replyRead.SetReadDeadline(deadline)); err != nil {
		return a, err
	}
	if err := channel.Send(Prepare, a.prepare.encode()); err != nil {
		return a, err
	}
	joined, err := channel.Receive()
	if err != nil {
		return a, err
	}
	proof, err := decodePreparation(joined.Payload)
	if err != nil || joined.Opcode != Joined || proof != a.prepare {
		return a, ErrProtocol
	}
	return a, nil
}

func (a *acquiredSignal) commit(ctx context.Context) error {
	if a.committed || a.closed {
		return ErrProtocol
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := a.control.SetWriteDeadline(deadline); err != nil {
		return err
	}
	// Once a complete Commit may have crossed the pipe, ownership cannot be
	// canceled or replayed. A transport error is reconciled by wait + census.
	a.committed = true
	return a.channel.Send(Commit, nil)
}

func (a *acquiredSignal) closeEndpoints() error {
	if a.closed {
		return a.closeErr
	}
	a.closed = true
	a.closeErr = errors.Join(a.control.Close(), a.reply.Close())
	return a.closeErr
}

func (a *acquiredSignal) join(ctx context.Context) error {
	select {
	case <-a.done:
		if a.committed && a.prepare.signal == killGroup {
			var exit *exec.ExitError
			if errors.As(a.waitErr, &exit) {
				if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() && status.Signal() == syscall.SIGKILL {
					return nil
				}
			}
			code := -1
			if a.cmd.ProcessState != nil {
				code = a.cmd.ProcessState.ExitCode()
			}
			return fmt.Errorf("acquired KILL helper exit %d did not prove SIGKILL: %w", code, errors.Join(ErrProtocol, a.waitErr))
		}
		return a.waitErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
