//go:build linux || darwin || freebsd

package adapter

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
)

// This short noninteractive gh transport owns its process group, not Runtime's
// interactive session model. It makes no claim about explicitly escaped sessions.
type commandOwner struct{}

func prepareCommand(cmd *exec.Cmd) (*commandOwner, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return &commandOwner{}, nil
}
func (o *commandOwner) started(cmd *exec.Cmd) error { return nil }
func (o *commandOwner) stop(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
}
func (o *commandOwner) finish(cmd *exec.Cmd, budget time.Duration) bool {
	// The root has been reaped. No cleanup claim if any group member remains,
	// including an unreaped descendant. Group escape is outside this profile.
	if e := syscall.Kill(-cmd.Process.Pid, 0); errors.Is(e, syscall.ESRCH) {
		return true
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	deadline := time.Now().Add(budget)
	for {
		if e := syscall.Kill(-cmd.Process.Pid, 0); errors.Is(e, syscall.ESRCH) {
			return true
		}
		if !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}
func (o *commandOwner) close() {}
