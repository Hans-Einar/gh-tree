//go:build linux || darwin || freebsd

package broker

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "--runtime-fixture-check-private-fds":
			for fd := 3; fd <= 5; fd++ {
				var stat unix.Stat_t
				if unix.Fstat(fd, &stat) == nil {
					kind := stat.Mode & unix.S_IFMT
					if kind == unix.S_IFIFO || kind == unix.S_IFDIR {
						os.Exit(1)
					}
				}
			}
			fmt.Println("__owned_no_private_fds__")
			os.Exit(0)
		case "--runtime-fixture-helper-recovery":
			if err := helperRecoveryFixture(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(0)
		case "--runtime-fixture-escape-root":
			terminate := make(chan os.Signal, 1)
			signal.Notify(terminate, syscall.SIGTERM)
			member, err := startFixtureMember(must(os.Executable()), true, true)
			if err != nil {
				if member != nil {
					member.cleanup()
				}
				os.Exit(1)
			}
			fmt.Println("__owned_escape_started__")
			select {
			case <-terminate:
			case <-time.After(5 * time.Second):
			}
			if err := member.cleanup(); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		case "--runtime-fixture-foreground":
			fmt.Println("__owned_foreground_ready__")
			time.Sleep(20 * time.Second)
			os.Exit(0)
		case "--runtime-fixture-input-count":
			data := make([]byte, 65536)
			n, err := io.ReadFull(os.Stdin, data)
			if err != nil {
				os.Exit(1)
			}
			fmt.Printf("owned-input=%d:%x\n", n, sha256.Sum256(data))
			time.Sleep(20 * time.Second)
			os.Exit(0)
		case supervisorPrivateMarker:
			os.Exit(RunSupervisor())
		case "--runtime-fixture-cwd":
			data, err := os.ReadFile("marker")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("owned-cwd=%s;marker=%s\n", must(os.Getwd()), data)
			os.Exit(0)
		case "--runtime-fixture-tree-root":
			reader, writer, err := os.Pipe()
			if err != nil {
				os.Exit(1)
			}
			cmd := exec.Command(must(os.Executable()), "--runtime-fixture-tree-branch")
			cmd.Stdout = writer
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				reader.Close()
				writer.Close()
				os.Exit(1)
			}
			go cmd.Wait()
			writer.Close()
			reader.SetReadDeadline(time.Now().Add(3 * time.Second))
			line, err := bufio.NewReader(reader).ReadString('\n')
			reader.Close()
			if err != nil {
				cmd.Process.Kill()
				os.Exit(1)
			}
			fmt.Print(line)
			os.Exit(0)
		case "--runtime-fixture-tree-branch":
			signal.Ignore(syscall.SIGTERM)
			member, err := startFixtureMember(must(os.Executable()), false, true)
			if err != nil {
				if member != nil {
					member.cleanup()
				}
				os.Exit(1)
			}
			fmt.Printf("owned-tree-child=%d;owned-tree-grandchild=%d\n", os.Getpid(), member.cmd.Process.Pid)
			time.Sleep(20 * time.Second)
			member.cleanup()
			os.Exit(0)
		case signalPrivateMarker:
			os.Exit(RunSignalHelper())
		case "--runtime-fixture-signal-suite":
			if err := signalFixtureSuite(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(0)
		case "--runtime-fixture-hold-ignore":
			signal.Ignore(syscall.SIGTERM)
			fallthrough
		case "--runtime-fixture-hold":
			fmt.Println("owned-fixture-ready")
			time.Sleep(20 * time.Second)
			os.Exit(0)
		case "--runtime-fixture-exit":
			os.Exit(0)
		}
	}
	os.Exit(m.Run())
}

type fixtureMember struct {
	cmd     *exec.Cmd
	done    chan struct{}
	waitErr error
}

func startFixtureMember(executable string, foreign, ignore bool) (*fixtureMember, error) {
	mode := "--runtime-fixture-hold"
	if ignore {
		mode += "-ignore"
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(executable, mode)
	cmd.Stdout = writer
	if foreign {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	if err := cmd.Start(); err != nil {
		return nil, errors.Join(err, reader.Close(), writer.Close())
	}
	m := &fixtureMember{cmd: cmd, done: make(chan struct{})}
	go func() { m.waitErr = cmd.Wait(); close(m.done) }()
	closeErr := writer.Close()
	if closeErr != nil {
		reader.Close()
		return m, closeErr
	}
	if err := reader.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		reader.Close()
		return m, err
	}
	line, err := bufio.NewReader(reader).ReadString('\n')
	closeErr = reader.Close()
	if err != nil || closeErr != nil {
		return m, errors.Join(err, closeErr)
	}
	if strings.TrimSpace(line) != "owned-fixture-ready" {
		return m, ErrProtocol
	}
	return m, nil
}

func (m *fixtureMember) cleanup() error {
	select {
	case <-m.done:
		return nil
	default:
	}
	// This kills only the newly created test fixture's directly owned child,
	// never a census-selected process/group or a user's process.
	err := m.cmd.Process.Kill()
	select {
	case <-m.done:
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return err
	case <-time.After(4 * time.Second):
		return errors.Join(err, errors.New("owned fixture waiter did not join"))
	}
}

func signalFixtureSuite() (result error) {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	var members []*fixtureMember
	var helpers []*acquiredSignal
	defer func() {
		for _, m := range members {
			result = errors.Join(result, m.cleanup())
		}
		for _, a := range helpers {
			result = errors.Join(result, a.closeEndpoints())
			select {
			case <-a.done:
				continue
			default:
			}
			// Failure-only fixture teardown owns this exact directly created
			// helper child. This never substitutes for the product's group
			// acquisition protocol or passing native cleanup assertions.
			result = errors.Join(result, errors.New("fixture helper needed failure-only teardown"), a.cmd.Process.Kill())
			select {
			case <-a.done:
			case <-time.After(4 * time.Second):
				result = errors.Join(result, errors.New("owned helper waiter did not join"))
			}
		}
	}()
	newMember := func(foreign, ignore bool) (*fixtureMember, error) {
		m, e := startFixtureMember(executable, foreign, ignore)
		if m != nil {
			members = append(members, m)
		}
		return m, e
	}
	acquire := func(group int, kind groupSignal) (*acquiredSignal, error) {
		a, e := acquireSignalGroup(ctx, executable, 1, group, kind)
		if a != nil {
			helpers = append(helpers, a)
		}
		return a, e
	}
	foreign, err := newMember(true, false)
	if err != nil {
		return err
	}
	if a, e := acquire(foreign.cmd.Process.Pid, termGroup); e == nil {
		if a != nil {
			a.closeEndpoints()
			a.join(ctx)
		}
		return errors.New("foreign session group acquired")
	}
	select {
	case <-foreign.done:
		return errors.New("foreign owned control fixture terminated")
	default:
	}
	fmt.Println("PASS foreign SID acquisition refused; control fixture remains alive")

	gone := exec.Command(executable, "--runtime-fixture-exit")
	gone.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := gone.Start(); err != nil {
		return err
	}
	goneGroup := gone.Process.Pid
	if err := gone.Wait(); err != nil {
		return err
	}
	if a, err := acquire(goneGroup, killGroup); err == nil {
		// A vanished candidate numerically equal to the new helper PID can
		// only create that helper's own group inside this owned SID.
		if a.cmd.Process.Pid != goneGroup {
			return errors.New("vanished group unexpectedly acquired a different helper identity")
		}
		if err := a.commit(ctx); err != nil {
			return err
		}
		if err := a.join(ctx); err != nil {
			return err
		}
		if err := a.closeEndpoints(); err != nil {
			return err
		}
	}
	fmt.Println("PASS vanished candidate refuses or acquires only the new helper's own group")

	member, err := newMember(false, false)
	if err != nil {
		return err
	}
	canceled, err := acquire(member.cmd.Process.Pid, termGroup)
	if err != nil {
		return err
	}
	if err := canceled.closeEndpoints(); err != nil {
		return err
	}
	if err := canceled.join(ctx); err != nil {
		return err
	}
	select {
	case <-member.done:
		return errors.New("pre-Commit cancellation signaled member")
	default:
	}
	fmt.Println("PASS close before Commit leaves member alive")

	badNonce, err := acquire(member.cmd.Process.Pid, termGroup)
	if err != nil {
		return err
	}
	badNonce.channel.nonce[0] ^= 1
	if err := badNonce.commit(ctx); err != nil {
		return err
	}
	if err := badNonce.join(ctx); err == nil {
		return errors.New("foreign Commit nonce accepted")
	}
	if err := badNonce.closeEndpoints(); err != nil {
		return err
	}
	select {
	case <-member.done:
		return errors.New("foreign Commit signaled member")
	default:
	}
	fmt.Println("PASS wrong Commit nonce rejected without signal")

	joined, err := acquire(member.cmd.Process.Pid, killGroup)
	if err != nil {
		return err
	}
	if err := member.cleanup(); err != nil {
		return err
	}
	if err := joined.commit(ctx); err != nil {
		return err
	}
	if err := joined.join(ctx); err != nil {
		return err
	}
	if err := joined.closeEndpoints(); err != nil {
		return err
	}
	fmt.Println("PASS last original member exits after Joined; acquired helper signals only its pinned group")

	stopped, err := newMember(false, true)
	if err != nil {
		return err
	}
	stopper, err := acquire(stopped.cmd.Process.Pid, stopGroup)
	if err != nil {
		return err
	}
	if err := stopper.commit(ctx); err != nil {
		return err
	}
	for {
		all, e := census(ctx)
		if e != nil {
			return e
		}
		found := 0
		for _, p := range all {
			if p.live && p.group == stopped.cmd.Process.Pid {
				if !p.stopped {
					found = -100
					break
				}
				found++
			}
		}
		if found == 2 {
			break
		}
		select {
		case <-ctx.Done():
			return errors.New("STOP helper did not remain parked with group")
		case <-time.After(10 * time.Millisecond):
		}
	}
	killer, err := acquire(stopped.cmd.Process.Pid, killGroup)
	if err != nil {
		return err
	}
	if err := killer.commit(ctx); err != nil {
		return err
	}
	if err := killer.join(ctx); err != nil {
		return err
	}
	if err := killer.closeEndpoints(); err != nil {
		return err
	}
	if err := stopper.join(ctx); err == nil {
		return errors.New("parked helper did not report actual signal termination")
	}
	if err := stopper.closeEndpoints(); err != nil {
		return err
	}
	select {
	case <-stopped.done:
	case <-ctx.Done():
		return errors.New("stopped member survived acquired KILL")
	}
	fmt.Println("PASS STOP helper parked, KILL helper joined stopped group, all owned waiters joined")

	if err := foreign.cleanup(); err != nil {
		return err
	}
	all, err := census(ctx)
	if err != nil {
		return err
	}
	live, groups, err := sessionMembers(all, os.Getpid())
	if err != nil {
		return err
	}
	if len(live) != 0 || len(groups) != 0 {
		return errors.New("fixture SID retained a live member")
	}
	fmt.Println("PASS full SID census contains only owning supervisor fixture")
	return nil
}

func TestNativeAcquiredSignalHelpers(t *testing.T) {
	t.Setenv("GORACE", strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"))
	executable := must(os.Executable())
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "--runtime-fixture-signal-suite")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.WaitDelay = time.Second
	output, err := cmd.CombinedOutput()
	t.Logf("native %s uid=%d gid=%d\n%s", syscallName(), os.Getuid(), os.Getgid(), output)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"foreign SID", "close before Commit", "wrong Commit nonce", "last original member", "STOP helper parked", "full SID census"} {
		if !strings.Contains(string(output), expected) {
			t.Fatalf("missing native evidence %q", expected)
		}
	}
}

func TestNativeSignalPrivateEntryFailsWithoutEndpoints(t *testing.T) {
	cmd := exec.Command(must(os.Executable()), signalPrivateMarker)
	if err := cmd.Run(); err == nil {
		t.Fatal("private helper ran without inherited endpoints")
	}
}

func TestNativeCensusObservesCurrentIdentity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	all := must(census(ctx))
	found := false
	sid := must(unix.Getsid(0))
	for _, p := range all {
		if p.pid == os.Getpid() {
			found = true
			if p.parent != os.Getppid() || p.group != unix.Getpgrp() || p.session != sid || !p.live {
				t.Fatalf("wrong native membership %+v", p)
			}
		}
	}
	if !found {
		t.Fatal("current process omitted")
	}
}

func syscallName() string {
	var name unix.Utsname
	if unix.Uname(&name) != nil {
		return "unix"
	}
	return string(bytesUntilZero(name.Sysname[:]))
}
func bytesUntilZero[T ~int8 | ~uint8](value []T) []byte {
	var result []byte
	for _, v := range value {
		if v == 0 {
			break
		}
		result = append(result, byte(v))
	}
	return result
}
