//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestNativeUnixClientControlEOFIsResidualWithoutQuiescenceProof(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-hold-ignore")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, config)
	if c == nil || err != nil || !start.Established {
		t.Fatal(start, err)
	}
	if err := c.control.close(); err != nil {
		t.Fatal(err)
	}
	f, err := c.Wait(ctx)
	if err == nil || f.CleanupComplete || len(f.Residuals) == 0 {
		t.Fatal("lost control manufactured full cleanup", f, err)
	}
	for _, file := range c.files {
		if !file.closed.Load() {
			t.Fatal("parent handle unjoined after EOF", file.stage)
		}
	}
	if c.activeIO.Load() != 0 || c.readerActive.Load() {
		t.Fatal("parent workers not joined")
	}
	for _, p := range must(census(ctx)) {
		if p.live && p.session == c.cmd.Process.Pid {
			t.Fatalf("EOF failed to trigger actual owned teardown: %+v", p)
		}
	}
}

func TestNativeUnixClientObservedEscapeRemainsResidual(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-escape-root")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, config)
	if c == nil || err != nil || !start.Established {
		t.Fatal(start, err)
	}
	f, err := c.Wait(ctx)
	if err == nil || f.CleanupComplete {
		t.Fatal("observed escape was called cleaned", f, err)
	}
	found := false
	for _, r := range f.Residuals {
		if r.Stage == api.Descendants {
			found = true
		}
	}
	if !found {
		t.Fatal("escape residual identity lost")
	}
	for _, file := range c.files {
		if !file.closed.Load() {
			t.Fatal("local resource abandoned after residual", file.stage)
		}
	}
	// The fixture root itself owns/terminates/reaps its deliberately escaped
	// direct child on TERM. Product code never numerically signals that child.
	for _, p := range must(census(ctx)) {
		if p.live && p.session == c.cmd.Process.Pid {
			t.Fatal("ordinary owned SID was not quiesced")
		}
	}
}

func TestNativeUnixClientCancellationAtAcquisitionBarriers(t *testing.T) {
	for _, stage := range []string{"cwd", "control-pipes", "io", "supervisor-created", "start-sent"} {
		t.Run(stage, func(t *testing.T) {
			config, _ := unixClientFixture(t, "--runtime-fixture-hold-ignore")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			config.hook = func(at string, _ *UnixClient) error {
				if at == stage {
					cancel()
				}
				return nil
			}
			c, start, err := StartUnix(ctx, config)
			if c == nil {
				t.Fatal("admitted native acquisition lost its owner")
			}
			if !start.Established && !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation fact lost", start, err)
			}
			requireUnixClientCleanup(t, c)
		})
	}
}

func TestNativeUnixClientPreAcquisitionCancellationOwnsNothing(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-cwd")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c, start, err := StartUnix(ctx, config)
	if c != nil || start.Established || !errors.Is(err, context.Canceled) {
		t.Fatal("pre-acquisition cancellation", c, start, err)
	}
	if _, err := os.Stat(config.Spec.RootLocator); err != nil {
		t.Fatal("refusal affected fixture directory", err)
	}
}

func TestNativeUnixUserRootDoesNotInheritPrivateEndpoints(t *testing.T) {
	config, output := unixClientFixture(t, "--runtime-fixture-check-private-fds")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, config)
	if c == nil || err != nil || !start.Established {
		t.Fatal(start, err)
	}
	f, err := c.Wait(ctx)
	if err != nil || !f.CleanupComplete || f.ExitCode != 0 || !strings.Contains(output.text(), "__owned_no_private_fds__") {
		t.Fatal("private descriptor leaked to user root", f, err)
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	// Positive detector control: only this newly owned fixture gets an
	// intentionally inherited fd3 pipe, and it must reject that inheritance.
	control := exec.Command(must(os.Executable()), "--runtime-fixture-check-private-fds")
	control.ExtraFiles = []*os.File{reader}
	if err := control.Run(); err == nil {
		t.Fatal("descriptor leak detector accepted intentional fixture leak")
	}
}
