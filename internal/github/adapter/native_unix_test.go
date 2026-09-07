//go:build linux || darwin || freebsd

package adapter

import (
	"errors"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func hideFixture(cmd *exec.Cmd) {}
func verifyFixtureChildExited(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if e := syscall.Kill(pid, 0); errors.Is(e, syscall.ESRCH) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	// An orphan zombie may remain visible until the platform init reaps it. The
	// product must preserve CleanupKnown=false in that case; do not call it clean.
	t.Logf("child %d remains visible to kill(0); conservative group cleanup is required", pid)
}
