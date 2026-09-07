//go:build linux || darwin || freebsd

package broker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func malformedUnixStarts(t *testing.T, spec StartSpec) map[string][]byte {
	t.Helper()
	valid := must(encodeUnixStart(spec, 20*time.Millisecond, 200*time.Millisecond))
	result := map[string][]byte{
		"truncated":              valid[:unixStartHeaderSize-1],
		"trailing":               append(append([]byte(nil), valid...), 0),
		"shared-without-periods": must(EncodeStart(spec)),
		"oversized":              make([]byte, MaxFrame-headerSize+1),
	}
	badVersion := append([]byte(nil), valid...)
	badVersion[0]++
	result["version"] = badVersion
	for name, period := range map[string]uint64{"zero": 0, "too-short": uint64(time.Millisecond - 1), "too-long": uint64(time.Minute + 1), "overflow": ^uint64(0)} {
		for _, offset := range []int{1, 9} {
			value := append([]byte(nil), valid...)
			binary.BigEndian.PutUint64(value[offset:], period)
			side := "grace-"
			if offset == 9 {
				side = "force-"
			}
			result[side+name] = value
		}
	}
	return result
}

func TestUnixStartEnvelopeEffectivePeriodsAndMalformedConfiguration(t *testing.T) {
	spec := nativeSpec(t)
	grace, force, err := effectiveUnixPeriods(0, 0)
	if err != nil || grace != 2*time.Second || force != 3*time.Second {
		t.Fatal(grace, force, err)
	}
	for _, periods := range [][2]time.Duration{{grace, force}, {time.Millisecond, time.Minute}, {20*time.Millisecond + 1, 200*time.Millisecond + 3}} {
		payload := must(encodeUnixStart(spec, periods[0], periods[1]))
		decoded, gotGrace, gotForce, err := decodeUnixStart(payload)
		if err != nil || gotGrace != periods[0] || gotForce != periods[1] || !bytes.Equal(must(EncodeStart(decoded)), must(EncodeStart(spec))) {
			t.Fatal("startup values changed", gotGrace, gotForce, err)
		}
	}
	for name, payload := range malformedUnixStarts(t, spec) {
		t.Run(name, func(t *testing.T) {
			if _, _, _, err := decodeUnixStart(payload); !errors.Is(err, ErrProtocol) {
				t.Fatal("malformed startup accepted", err)
			}
		})
	}
	for _, invalid := range []time.Duration{-1, time.Millisecond - 1, time.Minute + 1} {
		for _, periods := range [][2]time.Duration{{invalid, force}, {grace, invalid}} {
			if _, _, err := effectiveUnixPeriods(periods[0], periods[1]); !errors.Is(err, ErrProtocol) {
				t.Fatal("invalid configuration accepted")
			}
			if _, err := encodeUnixStart(spec, periods[0], periods[1]); !errors.Is(err, ErrProtocol) {
				t.Fatal("invalid wire periods encoded")
			}
			config, _ := unixClientFixture(t, "--runtime-fixture-cwd")
			config.GracePeriod, config.ForcePeriod = periods[0], periods[1]
			owner, start, err := StartUnix(context.Background(), config)
			if owner != nil || start.Established || !errors.Is(err, ErrProtocol) {
				t.Fatal("invalid construction acquired resources", owner, start, err)
			}
		}
	}
	// The Unix wrapper participates in the same total frame ceiling.
	spec.Environment = nil
	spec.Environment = []string{"A=" + strings.Repeat("x", MaxFrame-headerSize-len(must(EncodeStart(spec)))-7)}
	if _, err := EncodeStart(spec); err != nil {
		t.Fatal("shared payload control must fit", err)
	}
	if _, err := encodeUnixStart(spec, grace, force); !errors.Is(err, ErrProtocol) {
		t.Fatal("Unix header exceeded frame ceiling", err)
	}
}

func TestNativeUnixMalformedStartupNeverExecutesUser(t *testing.T) {
	spec := nativeSpec(t)
	spec.Executable = must(os.Executable())
	spec.Arguments = []string{"--runtime-fixture-cwd"}
	spec.ParentID = uint64(os.Getpid())
	for name, payload := range malformedUnixStarts(t, spec) {
		if len(payload) > MaxFrame-headerSize {
			continue
		} // rejected by frame codec above native dispatch
		t.Run(name, func(t *testing.T) {
			acquired, err := AcquireCwd(spec)
			if err != nil {
				closeAcquired(t, acquired)
				t.Fatal(err)
			}
			defer acquired.Close()
			controlRead, controlWrite, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer controlRead.Close()
			defer controlWrite.Close()
			replyRead, replyWrite, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer replyRead.Close()
			defer replyWrite.Close()
			var output bytes.Buffer
			cmd := exec.Command(spec.Executable, supervisorPrivateMarker)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			cmd.ExtraFiles = []*os.File{controlRead, replyWrite, acquired.File()}
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			controlRead.Close()
			replyWrite.Close()
			channel := must(NewChannel(replyRead, controlWrite, Parent, UnixSupervisor, 13, must(FreshNonce())))
			controlWrite.SetWriteDeadline(time.Now().Add(time.Second))
			sendErr := channel.Send(Start, payload)
			controlWrite.Close() // also ensures an erroneously accepted fixture owns EOF cleanup
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case waitErr := <-done:
				var exit *exec.ExitError
				if sendErr != nil || !errors.As(waitErr, &exit) || exit.ExitCode() != 131 || output.Len() != 0 {
					t.Fatal("malformed private configuration reached user execution", sendErr, waitErr, output.String())
				}
			case <-time.After(8 * time.Second):
				t.Fatal("malformed private startup did not join its exact waiter")
			}
		})
	}
}

func TestNativeUnixNaturalRootExitUsesConstructionPeriods(t *testing.T) {
	for _, defaults := range []bool{false, true} {
		name := "injected"
		if defaults {
			name = "defaults"
		}
		t.Run(name, func(t *testing.T) {
			config, output := unixClientFixture(t, "--runtime-fixture-tree-root")
			config.GracePeriod, config.ForcePeriod = 20*time.Millisecond, 200*time.Millisecond
			bound := 1200 * time.Millisecond
			if defaults {
				config.GracePeriod, config.ForcePeriod, bound = 0, 0, 8*time.Second
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			began := time.Now()
			c, start, err := StartUnix(ctx, config)
			if c != nil {
				defer requireUnixClientCleanup(t, c)
			}
			if err != nil || !start.Established {
				t.Fatal(start, err)
			}
			observation, stop := context.WithTimeout(context.Background(), bound)
			fact, waitErr := c.Wait(observation)
			stop()
			elapsed := time.Since(began)
			if !fact.CleanupComplete || !fact.RootExited || !fact.Quiescent || len(fact.Residuals) != 0 {
				t.Errorf("natural cleanup periods: elapsed=%v fact=%+v err=%v", elapsed, fact, waitErr)
			}
			// Always join the retained native owner before reporting a timing failure.
			final, finalErr := c.Wait(ctx)
			if !final.CleanupComplete {
				t.Fatal("fixture cleanup incomplete", final, finalErr)
			}
			if !strings.Contains(output.text(), "owned-tree-child=") || !strings.Contains(output.text(), "owned-tree-grandchild=") {
				t.Fatal("missing actual descendants", output.text())
			}
			if defaults && (elapsed < 1500*time.Millisecond || c.config.GracePeriod != 2*time.Second || c.config.ForcePeriod != 3*time.Second) {
				t.Fatal("production defaults were changed", elapsed, c.config.GracePeriod, c.config.ForcePeriod)
			}
			t.Logf("natural root-before-descendants %s cleanup=%v", name, elapsed)
		})
	}
}

func TestNativeUnixFailureAndControlEOFUseConstructionPeriods(t *testing.T) {
	for _, trigger := range []string{"early-eof", "control-eof", "invalid-command"} {
		t.Run(trigger, func(t *testing.T) {
			config, output := unixClientFixture(t, "--runtime-fixture-hold-ignore")
			config.GracePeriod, config.ForcePeriod = 20*time.Millisecond, 200*time.Millisecond
			if trigger == "early-eof" {
				config.hook = func(stage string, c *UnixClient) error {
					if stage == "start-sent" {
						return c.control.close()
					}
					return nil
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			c, start, err := StartUnix(ctx, config)
			if c == nil {
				t.Fatal("missing native owner", start, err)
			}
			if trigger != "early-eof" {
				if !start.Established || err != nil {
					t.Fatal(start, err)
				}
				until := time.Now().Add(time.Second)
				for !strings.Contains(output.text(), "owned-fixture-ready") && time.Now().Before(until) {
					time.Sleep(time.Millisecond)
				}
				if !strings.Contains(output.text(), "owned-fixture-ready") {
					t.Fatal("TERM-ignore fixture did not initialize")
				}
			}
			began := time.Now()
			switch trigger {
			case "control-eof":
				if err := c.control.close(); err != nil {
					t.Fatal(err)
				}
			case "invalid-command":
				if err := c.send(Resize, nil); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case <-c.done:
			case <-time.After(1200 * time.Millisecond):
				t.Error("native failure/EOF cleanup ignored construction periods")
			}
			fact, waitErr := c.Wait(ctx)
			if fact.CleanupComplete || waitErr == nil || len(fact.Residuals) == 0 {
				t.Fatal("lost private control manufactured cleanup proof", fact, waitErr)
			}
			select {
			case <-c.done:
			default:
				t.Fatal("retained owner has not joined")
			}
			for _, p := range must(census(ctx)) {
				if p.live && p.session == c.cmd.Process.Pid {
					t.Fatal("native SID member survived owned cleanup")
				}
			}
			for _, file := range c.files {
				if !file.closed.Load() {
					t.Fatal("native descriptor not closed")
				}
			}
			t.Logf("%s cleanup owners joined in %v; control residual retained", trigger, time.Since(began))
		})
	}
}
