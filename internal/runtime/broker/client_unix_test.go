//go:build linux || darwin || freebsd

package broker

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func unixClientFixture(t *testing.T, mode string) (UnixConfig, *fixtureOutput) {
	t.Helper()
	spec := nativeSpec(t)
	spec.Executable = must(os.Executable())
	spec.Arguments = []string{mode}
	output := &fixtureOutput{}
	config := UnixConfig{SessionID: 11, Spec: spec, Output: func(_ api.OutputStream, data []byte) {
		if _, err := output.Write(data); err != nil {
			panic(err)
		}
	}, GracePeriod: 20 * time.Millisecond, ForcePeriod: time.Second}
	return config, output
}

func requireUnixClientCleanup(t *testing.T, c *UnixClient) {
	t.Helper()
	c.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	f, err := c.Wait(ctx)
	if !f.CleanupComplete || len(f.Residuals) != 0 {
		t.Errorf("incomplete client cleanup: %+v / %v", f, err)
	}
}

func TestNativeUnixClientNaturalExitAndIndependentStartContext(t *testing.T) {
	config, output := unixClientFixture(t, "--runtime-fixture-cwd")
	ctx, cancel := context.WithCancel(context.Background())
	c, start, err := StartUnix(ctx, config)
	if c != nil {
		t.Cleanup(func() { requireUnixClientCleanup(t, c) })
	}
	if err != nil {
		t.Fatal(err)
	}
	if !start.Established || start.Cwd != filepath.Join(config.Spec.RootLocator, config.Spec.Components[0]) {
		t.Fatal("startup facts", start)
	}
	cancel()
	waitCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	f, err := c.Wait(waitCtx)
	if err != nil || !f.Established || !f.RootExited || !f.CleanupComplete || f.ExitCode != 0 || f.Signal != 0 {
		t.Fatalf("final native facts: %+v / %v", f, err)
	}
	if !strings.Contains(output.text(), "marker=selected-original") {
		t.Fatal("draining lost output")
	}
	if _, err := c.NextFact(waitCtx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.NextFact(waitCtx); !errors.Is(err, io.EOF) {
		t.Fatal("native fact EOF", err)
	}
}

func TestNativeUnixClientLateStartCancelDoesNotKillPersistentRoot(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-hold-ignore")
	ctx, cancel := context.WithCancel(context.Background())
	c, start, err := StartUnix(ctx, config)
	if c != nil {
		t.Cleanup(func() { requireUnixClientCleanup(t, c) })
	}
	if err != nil || !start.Established {
		t.Fatal(start, err)
	}
	cancel()
	short, stop := context.WithTimeout(context.Background(), 30*time.Millisecond)
	f, err := c.Wait(short)
	stop()
	if !errors.Is(err, context.DeadlineExceeded) || f.RootExited || f.CleanupComplete {
		t.Fatal("start context canceled persistent session", f, err)
	}
	began := time.Now()
	c.Stop()
	wait, done := context.WithTimeout(context.Background(), 4*time.Second)
	defer done()
	f, err = c.Wait(wait)
	if !f.CleanupComplete || !f.RootExited || f.Signal == 0 {
		t.Fatalf("owned forced cleanup: %+v / %v", f, err)
	}
	if time.Since(began) > 1500*time.Millisecond {
		t.Fatal("construction cleanup budgets were not applied")
	}
}

func TestNativeUnixClientEveryPartialCheckpointRetainsAndClosesOwner(t *testing.T) {
	injected := errors.New("owned native fixture acquisition failure")
	for _, stage := range []string{"cwd", "control-pipes", "io", "supervisor-created", "start-sent"} {
		t.Run(stage, func(t *testing.T) {
			config, _ := unixClientFixture(t, "--runtime-fixture-hold-ignore")
			config.hook = func(at string, _ *UnixClient) error {
				if at == stage {
					return injected
				}
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			c, _, err := StartUnix(ctx, config)
			if c == nil || !errors.Is(err, injected) {
				t.Fatal("partial owner/error lost", c, err)
			}
			f, waitErr := c.Wait(ctx)
			if !f.CleanupComplete || len(f.Residuals) != 0 || !errors.Is(waitErr, injected) {
				t.Fatalf("partial resource unwind: %+v / %v", f, waitErr)
			}
		})
	}
}

func TestNativeUnixClientInputFullAndCanceledPartialAreNotReplayed(t *testing.T) {
	config, output := unixClientFixture(t, "--runtime-fixture-input-count")
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, config)
	if c != nil {
		t.Cleanup(func() { requireUnixClientCleanup(t, c) })
	}
	if err != nil || !start.Established {
		t.Fatal(err)
	}
	data := bytes.Repeat([]byte{0, 255, 13, 27}, 16384)
	delivery, err := c.Write(ctx, data)
	if err != nil || delivery.Accepted != 65536 || delivery.Delivered != 65536 || !delivery.Completed {
		t.Fatal("whole input delivery", delivery, err)
	}
	expected := fmt.Sprintf("owned-input=65536:%x", sha256.Sum256(data))
	deadline := time.Now().Add(time.Second)
	for !strings.Contains(output.text(), expected) {
		if time.Now().After(deadline) {
			t.Fatal("native input bytes changed", output.text())
		}
		time.Sleep(time.Millisecond)
	}
	requireUnixClientCleanup(t, c)

	blockedConfig, _ := unixClientFixture(t, "--runtime-fixture-hold-ignore")
	blocked, _, err := StartUnix(ctx, blockedConfig)
	if blocked != nil {
		t.Cleanup(func() { requireUnixClientCleanup(t, blocked) })
	}
	if err != nil {
		t.Fatal(err)
	}
	writeCtx, writeCancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer writeCancel()
	var total uint32
	for i := 0; i < 128; i++ {
		part, e := blocked.Write(writeCtx, data)
		total += part.Delivered
		if e != nil {
			if part.Delivered > part.Accepted || part.Completed {
				t.Fatal("partial write fabricated completion", part, e)
			}
			if !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal("write cancellation lost", e)
			}
			break
		}
		if i == 127 {
			t.Fatal("owned input fixture never blocked")
		}
	}
	if total == 0 {
		t.Fatal("native partial-write positive control delivered no bytes")
	}
}

func TestNativeUnixClientStopUnblocksOwnedWriterAndRejectsNewControls(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-hold-ignore")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := StartUnix(ctx, config)
	if c != nil {
		t.Cleanup(func() { requireUnixClientCleanup(t, c) })
	}
	if err != nil {
		t.Fatal(err)
	}
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for i := 0; i < 128; i++ {
			if _, err := c.Write(context.Background(), make([]byte, 65536)); err != nil {
				return
			}
		}
	}()
	time.Sleep(20 * time.Millisecond)
	c.Stop()
	f, err := c.Wait(ctx)
	if !f.CleanupComplete {
		t.Fatal("writer prevented owned cleanup", f, err)
	}
	select {
	case <-writerDone:
	case <-time.After(time.Second):
		t.Fatal("writer did not join")
	}
	if _, err := c.Write(ctx, []byte{1}); err == nil {
		t.Fatal("late input accepted")
	}
	if _, err := c.Resize(ctx, 30, 100); err == nil {
		t.Fatal("pipe resize accepted")
	}
	if _, err := c.Interrupt(ctx); err == nil {
		t.Fatal("pipe ETX accepted")
	}
}

func TestNativeUnixClientFailedStartPreservesTypedFacts(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-cwd")
	config.Spec.Executable = "./missing-owned-fixture"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, config)
	if c == nil || start.Established || err == nil {
		t.Fatal("failed start facts", start, err)
	}
	f, waitErr := c.Wait(ctx)
	var failure UnixFailure
	if !f.CleanupComplete || !errors.As(waitErr, &failure) || failure.Code != api.NotFound {
		t.Fatal("typed failed-start cleanup", f, waitErr)
	}
}

func TestNativeUnixClientPTYResizeInputAndETX(t *testing.T) {
	config, output := unixClientFixture(t, "unused")
	config.Spec.Terminal = true
	config.Spec.Executable = "/bin/sh"
	config.Spec.Arguments = []string{"-i"}
	if _, err := os.Stat("/bin/bash"); err == nil {
		config.Spec.Executable = "/bin/bash"
		config.Spec.Arguments = []string{"--noprofile", "--norc", "-i"}
	}
	config.Spec.Environment = []string{"PATH=/bin:/usr/bin", "HISTFILE=/dev/null", "ENV=/dev/null", "HOME=" + config.Spec.RootLocator, "TERM=xterm"}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, config)
	if c != nil {
		t.Cleanup(func() { requireUnixClientCleanup(t, c) })
	}
	if err != nil || !start.Established {
		t.Fatal(start, err)
	}
	if d, err := c.Resize(ctx, 27, 92); err != nil || !d.Completed {
		t.Fatal("native resize", d, err)
	}
	if d, err := c.Write(ctx, []byte("stty size; sleep 20\n")); err != nil || !d.Completed {
		t.Fatal("native PTY input", d, err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(output.text(), "27 92\r\n") {
		if time.Now().After(deadline) {
			t.Fatal("actual resize absent", output.text())
		}
		time.Sleep(time.Millisecond)
	}
	if d, err := c.Interrupt(ctx); err != nil || !d.Completed || d.Delivered != 1 {
		t.Fatal("ETX delivery", d, err)
	}
	if _, err := c.Write(ctx, []byte("printf '__owned_interrupt__\\n'\n")); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for !strings.Contains(output.text(), "\r\n__owned_interrupt__\r\n") {
		if time.Now().After(deadline) {
			t.Fatal("foreground interrupt control absent", output.text())
		}
		time.Sleep(time.Millisecond)
	}
	c.mu.Lock()
	exited := c.latest.RootExited
	c.mu.Unlock()
	if exited {
		t.Fatal("terminal ETX became whole-session stop")
	}
	requireUnixClientCleanup(t, c)
}
