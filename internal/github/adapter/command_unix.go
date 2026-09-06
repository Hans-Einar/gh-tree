//go:build linux || darwin || freebsd

package adapter

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
)

// This short noninteractive gh transport owns its root and pipes. A dedicated
// group provides conservative residual observation only, never signal authority.
// There is no Runtime/session containment or escaped-descendant cleanup claim.
type commandOwner struct{}

func prepareCommand(cmd *exec.Cmd) (*commandOwner, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return &commandOwner{}, nil
}
func (o *commandOwner) started(cmd *exec.Cmd) error { return nil }
func (o *commandOwner) stop(cmd *exec.Cmd) {
	// Go owns root process identity (including its platform lifetime safeguards).
	// A numeric PGID can be reused after Wait reaps the root: never signal it.
	_ = cmd.Process.Kill()
}
func (o *commandOwner) finish(cmd *exec.Cmd, budget time.Duration) bool {
	// The root has been reaped. No cleanup claim if any group member remains,
	// including an unreaped descendant. Group escape is outside this profile.
	if e := syscall.Kill(-cmd.Process.Pid, 0); errors.Is(e, syscall.ESRCH) {
		return true
	}
	return false
}
func (o *commandOwner) close() {}
