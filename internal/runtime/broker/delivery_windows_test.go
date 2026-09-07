package broker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func windowsDeliveryFixture(t *testing.T, mode string, hook func(string, *WindowsClient)) (*WindowsClient, context.Context, func() string) {
	t.Helper()
	s := windowsSpec(t)
	s.Environment = os.Environ()
	s.Terminal = mode == "terminal"
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Executable = exe
	s.Arguments = []string{"-test.run=^TestWindowsOwnedUserFixture$", "--", "--owned-windows-fixture", mode}
	var mu sync.Mutex
	var output bytes.Buffer
	config := WindowsConfig{SessionID: 801, Spec: s, Image: exe, GracePeriod: 20 * time.Millisecond, ForcePeriod: 3 * time.Second, hook: hook, Output: func(_ api.OutputStream, b []byte) { mu.Lock(); output.Write(b); mu.Unlock() }}
	if _, embedded, e := MachineRoute(); e != nil {
		t.Fatal(e)
	} else if embedded {
		config.Extraction = extractedNativeFixture(t)
		config.Image = config.Extraction.Path()
	}
	live, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	client, started, err := StartWindows(live, config)
	if client != nil {
		t.Cleanup(func() {
			client.Stop()
			join, stop := context.WithTimeout(context.Background(), 10*time.Second)
			defer stop()
			final, e := client.Wait(join)
			if !final.CleanupComplete {
				t.Errorf("fixture cleanup %+v %v", final, e)
			}
			cancel()
		})
	} else {
		cancel()
	}
	if err != nil || !started.Established {
		t.Fatalf("start %+v %v", started, err)
	}
	return client, live, func() string { mu.Lock(); defer mu.Unlock(); return output.String() }
}

func TestWindowsCanceledControlAdmission(t *testing.T) {
	for _, mode := range []string{"before-call", "waiting-for-serialization", "stop-wins"} {
		t.Run(mode, func(t *testing.T) {
			client, live, output := windowsDeliveryFixture(t, "terminal", nil)
			for _, op := range []string{"write", "resize", "interrupt"} {
				ctx, cancel := context.WithCancel(context.Background())
				call := func() (WindowsDelivery, error) {
					switch op {
					case "write":
						return client.Write(ctx, []byte("CANCELED_CONTROL_EFFECT\r"))
					case "resize":
						return client.Resize(ctx, 31, 101)
					default:
						return client.Interrupt(ctx)
					}
				}
				var result controlReply
				if mode == "before-call" {
					cancel()
					result.delivery, result.err = call()
				} else {
					client.controlMu.Lock()
					done := make(chan controlReply, 1)
					go func() { d, e := call(); done <- controlReply{d, e} }()
					if mode == "stop-wins" {
						client.Stop()
					} else {
						cancel()
					}
					select {
					case result = <-done:
					case <-time.After(time.Second):
						client.controlMu.Unlock()
						t.Fatal("canceled admission waited for serialized native work")
					}
					client.controlMu.Unlock()
				}
				cancel()
				want := context.Canceled
				if mode == "stop-wins" {
					want = io.ErrClosedPipe
				}
				if !errors.Is(result.err, want) || result.delivery.Dispatched || result.delivery.Receipt != nil {
					t.Fatalf("%s admitted canceled/stopped control %+v %v", op, result.delivery, result.err)
				}
			}
			if mode != "stop-wins" {
				delivery, err := client.Write(live, []byte("UNCHANGED_OBSERVER\r"))
				if err != nil || !delivery.Completed {
					t.Fatalf("observer %+v %v", delivery, err)
				}
			}
			final, err := client.Wait(live)
			if err != nil || !final.CleanupComplete {
				t.Fatalf("join %+v %v", final, err)
			}
			got := output()
			if strings.Contains(got, "CANCELED_CONTROL_EFFECT") || strings.Contains(got, "SIZE 101 31") {
				t.Fatalf("canceled effect reached native user: %q", got)
			}
			if mode != "stop-wins" && !strings.Contains(got, "INPUT UNCHANGED_OBSERVER SIZE 80 24") {
				t.Fatalf("native observer did not survive with unchanged geometry: %q", got)
			}
		})
	}
}

func TestWindowsCompletionWinsCanceledWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, live, output := windowsDeliveryFixture(t, "terminal", func(stage string, c *WindowsClient) {
		if stage != "control-dispatched" {
			return
		}
		c.pendingMu.Lock()
		var receipt *WindowsReceipt
		for _, r := range c.pending {
			receipt = r
		}
		c.pendingMu.Unlock()
		if receipt == nil {
			t.Error("missing admitted receipt")
			return
		}
		select {
		case <-receipt.done:
			cancel()
		case <-time.After(5 * time.Second):
			t.Error("native delivery did not finish")
		}
	})
	result, err := client.Write(ctx, []byte("RACED_COMPLETION\r"))
	if ctx.Err() == nil || err != nil || !result.Completed || !result.Dispatched || result.Receipt == nil || result.Delivered != uint32(len("RACED_COMPLETION\r")) {
		t.Fatalf("known reply lost to cancellation: %+v %v", result, err)
	}
	final, err := client.Wait(live)
	if err != nil || !final.CleanupComplete || !strings.Contains(output(), "INPUT RACED_COMPLETION") {
		t.Fatalf("native join/output %+v %v %q", final, err, output())
	}
	again, err := result.Receipt.Wait(ctx)
	if err != nil || again != result {
		t.Fatalf("repeat observation changed known delivery: %+v %v", again, err)
	}
}

func windowsBlockedDelivery(t *testing.T) (*WindowsClient, context.Context, WindowsDelivery, []byte) {
	t.Helper()
	client, live, _ := windowsDeliveryFixture(t, "hold", nil)
	data := bytes.Repeat([]byte{'A'}, MaxFrame-headerSize)
	first, err := client.Write(live, data)
	if err != nil || first.Delivered != uint32(len(data)) {
		t.Fatalf("fill %+v %v", first, err)
	}
	short, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()
	result, err := client.Write(short, data)
	if !errors.Is(err, context.DeadlineExceeded) || result.Completed || !result.Dispatched || result.Receipt == nil {
		t.Fatalf("post-send timeout lost pending effect: %+v %v", result, err)
	}
	again, err := result.Receipt.Wait(short)
	if !errors.Is(err, context.DeadlineExceeded) || again != result {
		t.Fatalf("canceled join changed pending receipt: %+v %v", again, err)
	}
	refused, err := client.Write(live, []byte{'B'})
	if !errors.Is(err, ErrWindowsControlsBusy) || refused.Dispatched || refused.Receipt != nil {
		t.Fatalf("second in-flight input bypassed native bound: %+v %v", refused, err)
	}
	return client, live, result, data
}

func TestWindowsCanceledInputRetainsEventualDelivery(t *testing.T) {
	client, live, returned, data := windowsBlockedDelivery(t)
	client.Stop()
	final, err := client.Wait(live)
	if err != nil || !final.CleanupComplete {
		t.Fatalf("Stop waited for receipt consumption: %+v %v", final, err)
	}
	client.pendingMu.Lock()
	count := len(client.pending)
	client.pendingMu.Unlock()
	if count != 1 {
		t.Fatalf("unobserved receipt was discarded: %d", count)
	}
	var observers sync.WaitGroup
	for n := 0; n < 16; n++ {
		observers.Add(1)
		go func() {
			defer observers.Done()
			actual, e := returned.Receipt.Wait(live)
			if !errors.Is(e, io.ErrShortWrite) || !actual.Completed || !actual.Dispatched || actual.Receipt != returned.Receipt || actual.Accepted != uint32(len(data)) || actual.Delivered >= actual.Accepted {
				t.Errorf("eventual native partial facts %+v %v", actual, e)
			}
		}()
	}
	observers.Wait()
	client.pendingMu.Lock()
	defer client.pendingMu.Unlock()
	if len(client.pending) != 0 || client.inputPending {
		t.Fatal("receipt observation did not retire bounded ownership")
	}
}

func TestWindowsCanceledInputRetainsTerminalUnknown(t *testing.T) {
	client, live, returned, _ := windowsBlockedDelivery(t)
	if err := windows.TerminateProcess(client.process.Process, 77); err != nil {
		t.Fatal(err)
	}
	final, err := client.Wait(live)
	if err == nil || !final.CleanupComplete {
		t.Fatalf("outer containment failed: %+v %v", final, err)
	}
	actual, err := returned.Receipt.Wait(live)
	if !errors.Is(err, io.ErrClosedPipe) || actual.Completed || !actual.Dispatched || actual.Receipt != returned.Receipt {
		t.Fatalf("broker death fabricated delivery/no effect: %+v %v", actual, err)
	}
}

func TestWindowsFailureRetainsKnownBlockedInput(t *testing.T) {
	client, live, returned, data := windowsBlockedDelivery(t)
	// A malformed role-local resize reaches the real broker while its native
	// input is blocked. Failure cleanup must publish that write's known result.
	if _, err := client.send(Resize, []byte{1}, nil); err != nil {
		t.Fatal(err)
	}
	final, err := client.Wait(live)
	if !errors.Is(err, ErrProtocol) || !final.CleanupComplete {
		t.Fatalf("protocol failure cleanup: %+v %v", final, err)
	}
	actual, err := returned.Receipt.Wait(live)
	if !errors.Is(err, io.ErrShortWrite) || !actual.Completed || actual.Accepted != uint32(len(data)) || actual.Delivered >= actual.Accepted {
		t.Fatalf("failure erased known partial input: %+v %v", actual, err)
	}
}

type windowsAmbiguousWriter struct{ io.Writer }

func (w windowsAmbiguousWriter) Write(data []byte) (int, error) {
	n, err := w.Writer.Write(data)
	if err == nil {
		err = io.ErrShortWrite
	}
	return n, err
}

func TestWindowsAmbiguousSendRetainsNativeReply(t *testing.T) {
	client, live, _ := windowsDeliveryFixture(t, "hold", nil)
	// Inject an error after the actual full native pipe write. The local error
	// cannot prove no effect, even though the authenticated frame reached input.
	client.controlMu.Lock()
	client.channel.writer = windowsAmbiguousWriter{client.write}
	client.controlMu.Unlock()
	result, err := client.Write(live, []byte("OWNED_AMBIGUOUS_DELIVERY"))
	if !result.Dispatched || result.Receipt == nil {
		t.Fatalf("ambiguous send lost receipt: %+v %v", result, err)
	}
	final, waitErr := client.Wait(live)
	if waitErr == nil || !final.CleanupComplete {
		t.Fatalf("send failure cleanup: %+v %v", final, waitErr)
	}
	actual, err := result.Receipt.Wait(live)
	if !actual.Completed || actual.Accepted != uint32(len("OWNED_AMBIGUOUS_DELIVERY")) || actual.Delivered > actual.Accepted {
		t.Fatalf("ordered native answer discarded on send failure: %+v %v", actual, err)
	}
}

// The receipt capacity/invalid-payload checks isolate memory admission from OS
// work; the tests above independently prove the same receive/Stop path natively.
func TestWindowsDeliveryReceiptBoundsAndValidation(t *testing.T) {
	c := &WindowsClient{pending: make(map[uint64]*WindowsReceipt), stop: make(chan struct{}), done: make(chan struct{}), config: WindowsConfig{Spec: StartSpec{Terminal: true}}}
	for seq := uint64(1); seq <= 64; seq++ {
		r := &WindowsReceipt{owner: c, sequence: seq, op: Resize, done: make(chan struct{})}
		c.pending[seq] = r
		if err := c.receiveDelivery(binary.BigEndian.AppendUint64(nil, seq)); err != nil {
			t.Fatal(err)
		}
	}
	if len(c.pending) != 64 {
		t.Fatal("receive retired unobserved receipts")
	}
	for _, payload := range [][]byte{nil, {1}, binary.BigEndian.AppendUint64(nil, 0), binary.BigEndian.AppendUint64(nil, 1), binary.BigEndian.AppendUint64(nil, 65)} {
		if !errors.Is(c.receiveDelivery(payload), ErrProtocol) {
			t.Fatalf("accepted invalid/foreign/duplicate delivery: %x", payload)
		}
	}
	// A live native client's 65th unobserved receipt must refuse admission.
	client, live, _ := windowsDeliveryFixture(t, "terminal", nil)
	client.pendingMu.Lock()
	for seq, r := range c.pending {
		client.pending[seq+1000] = r
	}
	client.pendingMu.Unlock()
	result, err := client.Resize(live, 30, 100)
	if !errors.Is(err, ErrWindowsControlsBusy) || result.Dispatched || result.Receipt != nil {
		t.Fatalf("unobserved capacity bypass: %+v %v", result, err)
	}
	client.pendingMu.Lock()
	for seq := range c.pending {
		delete(client.pending, seq+1000)
	}
	client.pendingMu.Unlock()
	for _, r := range c.pending {
		if _, err := r.Wait(live); err != nil {
			t.Fatal(err)
		}
	}
	if len(c.pending) != 0 {
		t.Fatal("observed receipts retained")
	}
	client.Stop()
}
