package git

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

type commandResult struct {
	stdout, stderr []byte
	transport      api.CommandTransportOutcome
	err            error
}

// cappedOutput drains without allocating beyond the admitted bound. Exceeding
// it cancels the root immediately; WaitDelay bounds descendant-held pipes.
type cappedOutput struct {
	bytes     []byte
	limit     int
	truncated bool
	cancel    context.CancelFunc
	once      sync.Once
}

func (w *cappedOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := w.limit - len(w.bytes)
	if n > remaining {
		w.bytes = append(w.bytes, p[:remaining]...)
		w.truncated = true
		w.once.Do(w.cancel)
	} else {
		w.bytes = append(w.bytes, p...)
	}
	return n, nil
}

func transportValue(d api.CommandTransportOutcomeData) api.CommandTransportOutcome {
	v, err := api.NewCommandTransportOutcome(d)
	if err != nil {
		panic(err)
	}
	return v
}

func (a *Adapter) command(ctx context.Context, cwd string, mutation bool, args ...string) commandResult {
	return a.commandInput(ctx, cwd, mutation, nil, args...)
}

func (a *Adapter) commandInput(ctx context.Context, cwd string, mutation bool, input []byte, args ...string) commandResult {
	if ctx == nil {
		return commandResult{transport: transportValue(api.CommandTransportOutcomeData{CleanupKnown: true}), err: diagnostic(api.Invalid, "InvalidContext", "A context is required.")}
	}
	if err := ctx.Err(); err != nil {
		return commandResult{transport: transportValue(api.CommandTransportOutcomeData{CleanupKnown: true, CancellationRequested: true}), err: err}
	}
	budget := a.options.ReadTimeout
	if mutation {
		budget = a.options.MutationTimeout
	}
	bounded, timeout := context.WithTimeout(ctx, budget)
	defer timeout()
	runctx, cancel := context.WithCancel(bounded)
	defer cancel()
	stdout := &cappedOutput{limit: a.options.MaxStdoutBytes, cancel: cancel}
	stderr := &cappedOutput{limit: a.options.MaxStderrBytes, cancel: cancel}
	nativeArgs := []string{"--no-pager", "--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "maintenance.auto=false", "-c", "gc.auto=0"}
	nativeArgs = append(nativeArgs, args...)
	cmd := exec.CommandContext(runctx, a.options.GitExecutable, nativeArgs...)
	cmd.Dir = cwd
	cmd.Env = commandEnvironment(a.options.Environment)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if input != nil {
		cmd.Stdin = strings.NewReader(string(input))
	}
	cmd.WaitDelay = a.options.DrainTimeout
	hideWindow(cmd)
	if err := cmd.Start(); err != nil {
		return commandResult{transport: transportValue(api.CommandTransportOutcomeData{CleanupKnown: true, CancellationRequested: bounded.Err() != nil}), err: diagnostic(api.ProcessFailure, "CommandStartFailed", "The Git process could not start.")}
	}
	err := cmd.Wait() // Sole root waiter; joins os/exec's private copy goroutines.
	d := api.CommandTransportOutcomeData{Started: true, RootReaped: true, CleanupKnown: true, ExitCode: api.Some(cmd.ProcessState.ExitCode()), StdoutTruncated: stdout.truncated, StderrTruncated: stderr.truncated, CancellationRequested: bounded.Err() != nil}
	if runctx.Err() != nil || errors.Is(err, exec.ErrWaitDelay) {
		// Root termination or forced pipe closure proves no descendant barrier.
		d.CleanupKnown = false
		d.Diagnostics = append(d.Diagnostics, diagnostic(api.CleanupIncomplete, "CommandDescendantsUnknown", "The root was reaped; helper cleanup is not established."))
	}
	if stdout.truncated || stderr.truncated {
		err = diagnostic(api.Unavailable, "CommandOutputLimit", "Git output exceeded the configured bound.")
	} else if bounded.Err() != nil {
		err = bounded.Err()
	} else if errors.Is(err, exec.ErrWaitDelay) {
		err = diagnostic(api.CleanupIncomplete, "CommandDrainDeadline", "Git pipe drainage exceeded its deadline.")
	} else if err != nil {
		err = diagnostic(api.ProcessFailure, "CommandFailed", "Git returned a nonzero exit status.")
	}
	return commandResult{stdout: stdout.bytes, stderr: stderr.bytes, transport: transportValue(d), err: err}
}

// Native scope is always explicit. Never inherit an alternate index, git-dir,
// replace namespace, object store, or pathspec interpretation from the parent.
// Effective configuration and signer/SSH environment remain copied inputs.
func commandEnvironment(base []string) []string {
	drop := map[string]bool{"GIT_DIR": true, "GIT_COMMON_DIR": true, "GIT_WORK_TREE": true, "GIT_INDEX_FILE": true, "GIT_OBJECT_DIRECTORY": true, "GIT_ALTERNATE_OBJECT_DIRECTORIES": true, "GIT_NAMESPACE": true, "GIT_PREFIX": true, "GIT_CEILING_DIRECTORIES": true, "GIT_DISCOVERY_ACROSS_FILESYSTEM": true, "GIT_OPTIONAL_LOCKS": true, "GIT_LITERAL_PATHSPECS": true, "GIT_GLOB_PATHSPECS": true, "GIT_NOGLOB_PATHSPECS": true, "GIT_ICASE_PATHSPECS": true, "GIT_NO_REPLACE_OBJECTS": true, "GIT_TERMINAL_PROMPT": true, "GIT_PAGER": true, "LC_ALL": true}
	env := make([]string, 0, len(base)+6)
	for _, e := range base {
		name, _, _ := strings.Cut(e, "=")
		if !drop[strings.ToUpper(name)] {
			env = append(env, e)
		}
	}
	return append(env, "GIT_OPTIONAL_LOCKS=0", "GIT_LITERAL_PATHSPECS=1", "GIT_NO_REPLACE_OBJECTS=1", "GIT_TERMINAL_PROMPT=0", "GIT_PAGER=cat", "LC_ALL=C")
}

// observation intervals use wall-clock UTC acquisition bounds, not Git author
// times and not an ordering relation for SourceVersion.
func interval(start time.Time) api.ObservationInterval {
	end := time.Now().UTC()
	if end.Before(start) {
		end = start
	}
	v, err := api.NewObservationInterval(api.ObservationIntervalData{StartedAt: start, FinishedAt: end})
	if err != nil {
		panic(err)
	}
	return v
}
