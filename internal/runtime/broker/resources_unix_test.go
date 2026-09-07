//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestNativeUnixClientRepeatedPipesAndPTYDoNotLeak(t *testing.T) {
	config, _ := unixClientFixture(t, "--runtime-fixture-cwd")
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	warmup, _, err := StartUnix(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	requireUnixClientCleanup(t, warmup)
	beforeFD := 0
	if runtime.GOOS == "linux" {
		beforeFD = len(must(os.ReadDir("/proc/self/fd")))
	}
	beforeGo := runtime.NumGoroutine()
	closedHandles := 0
	var clients []*UnixClient
	for i := 0; i < 12; i++ {
		config.Spec.Terminal = i%2 != 0
		c, start, err := StartUnix(ctx, config)
		if err != nil || !start.Established {
			t.Fatal(start, err)
		}
		clients = append(clients, c)
		f, err := c.Wait(ctx)
		if err != nil || !f.CleanupComplete {
			t.Fatal(f, err)
		}
		for _, file := range c.files {
			if !file.closed.Load() {
				t.Fatal("retained native descriptor")
			}
			if _, err := file.file.Stat(); !errors.Is(err, os.ErrClosed) {
				t.Fatal("closed native file remains usable", err)
			}
			closedHandles++
		}
		if !c.cwdClosed.Load() || c.activeIO.Load() != 0 || c.readerActive.Load() {
			t.Fatal("retained native owner worker")
		}
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > beforeGo+1 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	afterGo := runtime.NumGoroutine()
	if afterGo > beforeGo+1 {
		t.Fatalf("goroutines leaked: %d→%d", beforeGo, afterGo)
	}
	if runtime.GOOS == "linux" {
		afterFD := len(must(os.ReadDir("/proc/self/fd")))
		if afterFD != beforeFD {
			t.Fatalf("native descriptors leaked: %d→%d", beforeFD, afterFD)
		}
		t.Logf("Linux process fd count %d→%d", beforeFD, afterFD)
	}
	t.Logf("%s: %d owned native descriptor objects closed; goroutines %d→%d", runtime.GOOS, closedHandles, beforeGo, afterGo)
	runtime.KeepAlive(clients)
}
