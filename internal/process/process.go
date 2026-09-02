package process

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes a command without involving a shell.
type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) ([]byte, error)
}

// ExecRunner is the production Runner implementation.
type ExecRunner struct{}

// Run executes name with an argument vector and returns combined stdout/stderr.
func (ExecRunner) Run(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, &CommandError{
			Name:   name,
			Args:   append([]string(nil), args...),
			Dir:    dir,
			Output: strings.TrimSpace(string(out)),
			Err:    err,
		}
	}
	return out, nil
}

// CommandError retains enough context to produce an actionable UI error.
type CommandError struct {
	Name   string
	Args   []string
	Dir    string
	Output string
	Err    error
}

func (e *CommandError) Error() string {
	detail := e.Output
	if detail == "" {
		detail = e.Err.Error()
	}
	return fmt.Sprintf("%s failed: %s", e.Name, detail)
}

func (e *CommandError) Unwrap() error { return e.Err }
