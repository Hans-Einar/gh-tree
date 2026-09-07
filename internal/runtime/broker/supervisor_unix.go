//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/unix"
)

const supervisorPrivateMarker = "--gh-tree-runtime-supervisor-v1"

func unixFailurePayload(phase byte, stage api.RuntimeCleanupStage, err error) []byte {
	code := api.IOFailure
	switch {
	case os.IsPermission(err):
		code = api.Permission
	case os.IsNotExist(err) || errors.Is(err, exec.ErrNotFound):
		code = api.NotFound
	case errors.Is(err, ErrCwd):
		code = api.StaleObservation
	case errors.Is(err, ErrProtocol):
		code = api.Invalid
	case errors.Is(err, ErrCensus) || errors.Is(err, context.DeadlineExceeded):
		code = api.Unavailable
	case phase != 1:
		code = api.CleanupIncomplete
	}
	return []byte{phase, byte(stage), byte(code)}
}

type parentFrame struct {
	frame Frame
	err   error
}

func findExecutableUnix(name string, environment []string) (string, error) {
	if strings.ContainsRune(name, os.PathSeparator) {
		return name, nil
	}
	path := ""
	for _, entry := range environment {
		if strings.HasPrefix(entry, "PATH=") {
			path = entry[5:]
		}
	}
	if path == "" {
		return "", exec.ErrNotFound
	}
	var last error
	for _, directory := range filepath.SplitList(path) {
		if directory == "" {
			directory = "."
		}
		candidate := filepath.Join(directory, name)
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			err = unix.Access(candidate, unix.X_OK)
			if err == nil {
				if !filepath.IsAbs(candidate) && !strings.ContainsRune(candidate, os.PathSeparator) {
					candidate = "./" + candidate
				}
				return candidate, nil
			}
		}
		if err != nil && !os.IsNotExist(err) {
			last = err
		}
	}
	if last != nil {
		return "", last
	}
	return "", exec.ErrNotFound
}

func cwdEnvironment(environment []string, locator string) []string {
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "PWD=") {
			result = append(result, entry)
		}
	}
	return append(result, "PWD="+locator)
}

func unixExitPayload(tree *unixTree) []byte {
	if tree.root == nil {
		return nil
	}
	select {
	case <-tree.rootDone:
	default:
		return nil
	}
	var data [8]byte
	binary.BigEndian.PutUint32(data[:4], uint32(int32(tree.root.ProcessState.ExitCode())))
	if status, ok := tree.root.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		binary.BigEndian.PutUint32(data[4:], uint32(status.Signal()))
	}
	return data[:]
}

func unixStartedPayload(pid int, locator string) []byte {
	e := encoder{}
	e.u64(uint64(pid))
	e.string(locator)
	return e.buf
}

// RunSupervisor owns one SID. It must be dispatched before application
// bootstrap; only its inherited private endpoints and designated cwd descriptor
// can authorize user work. Main Runtime never changes its own process cwd.
func RunSupervisor() int {
	control, err := inheritedPipe(3, false)
	if err != nil {
		return 130
	}
	defer control.Close()
	reply, err := inheritedPipe(4, true)
	if err != nil {
		return 130
	}
	defer reply.Close()
	sid, err := unix.Getsid(0)
	if err != nil || sid != os.Getpid() || unix.Getpgrp() != sid {
		return 131
	}
	channel, initial, err := AcceptChannel(control, reply, UnixSupervisor, Parent)
	if err != nil || initial.Opcode != Start {
		return 131
	}
	spec, err := DecodeStart(initial.Payload)
	if err != nil || spec.ParentID != uint64(os.Getppid()) {
		return 131
	}
	executable, err := os.Executable()
	if err != nil || !filepath.IsAbs(executable) {
		return 131
	}
	tree := &unixTree{executable: executable, session: initial.SessionID, grace: 2 * time.Second, force: 3 * time.Second}
	// Retain the designated descriptor until Fstat and the one Fchdir complete,
	// then close it before command lookup/creation. User roots inherit no fd5.
	cwd := os.NewFile(5, "runtime-acquired-cwd")
	unix.CloseOnExec(5)
	identity, identityErr := ObserveDirectory(cwd, spec.ProjectIdentity.Stamp())
	startupErr := identityErr
	startupStage := api.CwdAcquisition
	if identityErr == nil && !identity.Equal(spec.ProjectIdentity) {
		startupErr = ErrCwd
	}
	if startupErr == nil {
		startupErr = unix.Fchdir(5)
	}
	startupErr = errors.Join(startupErr, cwd.Close())
	actual := ""
	if startupErr == nil {
		actual, startupErr = os.Getwd()
		parts := append([]string{spec.RootLocator}, spec.Components...)
		if startupErr == nil && actual != filepath.Join(parts...) {
			startupErr = ErrCwd
		}
	}
	if err := control.SetReadDeadline(time.Time{}); err != nil {
		startupErr = errors.Join(startupErr, err)
	}
	send := func(op Opcode, payload []byte) error {
		if err := reply.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
			return err
		}
		return channel.Send(op, payload)
	}
	frames := make(chan parentFrame, 1)
	readDone := make(chan struct{})
	readStop := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			f, e := channel.Receive()
			select {
			case frames <- parentFrame{f, e}:
			case <-readStop:
				return
			}
			if e != nil {
				return
			}
		}
	}()
	defer func() { close(readStop); control.Close(); <-readDone }()
	if startupErr == nil {
		startupStage = api.ProcessContainment
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, _, startupErr = tree.observe(ctx)
		cancel()
	}
	if startupErr == nil {
		environment := cwdEnvironment(spec.Environment, actual)
		path, lookupErr := findExecutableUnix(spec.Executable, environment)
		startupErr = lookupErr
		if startupErr == nil {
			cmd := &exec.Cmd{Path: path, Args: append([]string{spec.Executable}, spec.Arguments...), Env: environment, Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if spec.Terminal {
				cmd.SysProcAttr.Foreground = true
				cmd.SysProcAttr.Ctty = 0
			}
			startupErr = tree.startRoot(cmd)
		}
	}
	parentGone := false
	stopping := startupErr != nil
	if startupErr != nil {
		if err := send(Failure, unixFailurePayload(1, startupStage, startupErr)); err != nil {
			parentGone = true
		}
	} else {
		if err := send(Started, unixStartedPayload(tree.root.Process.Pid, actual)); err != nil {
			parentGone = true
			stopping = true
		}
	}
	exitSent := false
	observeTick := time.NewTicker(50 * time.Millisecond)
	defer observeTick.Stop()
	for {
		if tree.root != nil && !exitSent {
			select {
			case <-tree.rootDone:
				exitSent = true
				stopping = true
				if !parentGone {
					if err := send(UserExit, unixExitPayload(tree)); err != nil {
						parentGone = true
					}
				}
			default:
			}
		}
		if stopping {
			quiescent, cleanupErr := tree.cleanup()
			if quiescent {
				if tree.root != nil && !exitSent {
					exitSent = true
					if !parentGone {
						if err := send(UserExit, unixExitPayload(tree)); err != nil {
							parentGone = true
						}
					}
				}
				if !parentGone {
					payload := []byte{0}
					if cleanupErr != nil {
						payload[0] = 1
					}
					if err := send(Quiescent, payload); err != nil {
						parentGone = true
					}
				}
				if !parentGone {
					for {
						event := <-frames
						if event.err != nil {
							parentGone = true
							break
						}
						if event.frame.Opcode == Release && len(event.frame.Payload) == 0 {
							break
						}
						if event.frame.Opcode != Stop && event.frame.Opcode != Abort {
							return 134
						}
					}
				}
				// No live owned SID member remains, all root/helper waiters have
				// joined. Parent continues draining while these final slave/pipe
				// references disappear on private-process exit.
				return 0
			}
			if !parentGone {
				if err := send(Failure, unixFailurePayload(2, api.Descendants, cleanupErr)); err != nil {
					parentGone = true
				}
			}
			if tree.nativeQuiescent && tree.escape {
				return 135
			}
		}
		select {
		case event := <-frames:
			if event.err != nil {
				parentGone = true
				stopping = true
				continue
			}
			switch event.frame.Opcode {
			case Stop, Abort:
				if len(event.frame.Payload) == 8 {
					grace := time.Duration(binary.BigEndian.Uint32(event.frame.Payload)) * time.Millisecond
					force := time.Duration(binary.BigEndian.Uint32(event.frame.Payload[4:])) * time.Millisecond
					if grace < time.Millisecond || grace > time.Minute || force < time.Millisecond || force > time.Minute {
						parentGone = true
					} else {
						tree.grace = grace
						tree.force = force
					}
				} else if len(event.frame.Payload) != 0 {
					parentGone = true
				}
				stopping = true
			default:
				parentGone = true
				stopping = true
			}
		case <-observeTick.C:
			if !stopping {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_, _, err := tree.observe(ctx)
				cancel()
				if err != nil || tree.escape {
					stopping = true
					if !parentGone {
						if e := send(Failure, unixFailurePayload(3, api.Descendants, err)); e != nil {
							parentGone = true
						}
					}
				}
			}
		}
	}
}
