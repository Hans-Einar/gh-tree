package adapter

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

type command struct {
	args     []string
	input    []byte
	mutation bool
}
type wireResult struct {
	stdout, stderr    []byte
	transport         api.CommandTransportOutcome
	started, finished time.Time
	err               error
}
type commandRunner func(context.Context, command) wireResult
type boundedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - b.buffer.Len()
	if len(p) > remaining {
		p = p[:remaining]
		b.truncated = true
	}
	_, _ = b.buffer.Write(p)
	return n, nil
}

func (a *Adapter) runCommand(ctx context.Context, request command) wireResult {
	r := wireResult{started: time.Now().UTC(), transport: noTransport()}
	defer func() {}()
	if err := checkContext(ctx); err != nil {
		r.err = err
		r.finished = time.Now().UTC()
		return r
	}
	budget := a.config.ReadTimeout
	if request.mutation {
		budget = a.config.MutationTimeout
	}
	child, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	cmd := exec.Command(a.config.Executable, append([]string(nil), request.args...)...)
	cmd.Env = commandEnvironment(a.config)
	cmd.Stdin = bytes.NewReader(append([]byte(nil), request.input...))
	out := &boundedBuffer{limit: a.config.StdoutLimit}
	errout := &boundedBuffer{limit: a.config.StderrLimit}
	cmd.Stdout = out
	cmd.Stderr = errout
	cmd.WaitDelay = a.config.DrainTimeout
	owner, err := prepareCommand(cmd)
	if err != nil {
		r.err = errors.New("command ownership setup failed")
		r.finished = time.Now().UTC()
		return r
	}
	defer owner.close()
	if err = cmd.Start(); err != nil {
		r.err = errors.New("command start failed")
		r.finished = time.Now().UTC()
		return r
	}
	td := api.CommandTransportOutcomeData{Started: true}
	if err = owner.started(cmd); err != nil {
		_ = cmd.Process.Kill()
	}
	// One waiter owns root reaping and Go's stdout/stderr/stdin copy joins.
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	var waitErr error
	if err != nil {
		waitErr = <-waited
	} else {
		select {
		case waitErr = <-waited:
		case <-child.Done():
			owner.stop(cmd)
			waitErr = <-waited
		}
	}
	td.RootReaped = cmd.ProcessState != nil
	td.CleanupKnown = owner.finish(cmd, a.config.DrainTimeout) && td.RootReaped && !errors.Is(waitErr, exec.ErrWaitDelay)
	if td.RootReaped {
		td.ExitCode = api.Some(cmd.ProcessState.ExitCode())
	}
	td.CancellationRequested = ctx.Err() != nil
	td.StdoutTruncated = out.truncated
	td.StderrTruncated = errout.truncated
	if !td.CleanupKnown {
		td.Diagnostics = append(td.Diagnostics, diagnostic(api.CleanupIncomplete, "command-cleanup-unproved"))
	}
	if err != nil {
		r.err = errors.New("command ownership establishment failed")
	} else if child.Err() != nil {
		r.err = child.Err()
	} else if waitErr != nil {
		r.err = errors.New("command failed")
	}
	if out.truncated {
		td.Diagnostics = append(td.Diagnostics, diagnostic(api.IOFailure, "machine-output-limit"))
		r.err = errors.Join(r.err, errors.New("machine output truncated"))
	}
	if errout.truncated {
		td.Diagnostics = append(td.Diagnostics, diagnostic(api.IOFailure, "diagnostic-output-limit"))
	}
	if !td.CleanupKnown {
		r.err = errors.Join(r.err, errors.New("command cleanup unproved"))
	}
	r.transport = must(api.NewCommandTransportOutcome(td))
	r.stdout = append([]byte(nil), out.buffer.Bytes()...)
	r.stderr = append([]byte(nil), errout.buffer.Bytes()...)
	r.finished = time.Now().UTC()
	return r
}

func commandEnvironment(c Config) []string {
	env := make([]string, 0, len(c.Environment)+8)
	replace := map[string]string{"GH_HOST": c.Host, "GH_PROMPT_DISABLED": "1", "GH_PAGER": "cat", "PAGER": "cat", "GH_NO_UPDATE_NOTIFIER": "1", "GH_NO_EXTENSION_UPDATE_NOTIFIER": "1", "GH_DEBUG": "", "GH_REPO": "", "BROWSER": ""}
	if c.ConfigDirectory != "" {
		replace["GH_CONFIG_DIR"] = c.ConfigDirectory
	}
	for _, v := range c.Environment {
		k, _, ok := strings.Cut(v, "=")
		if !ok {
			continue
		}
		if _, found := replace[strings.ToUpper(k)]; !found {
			env = append(env, v)
		}
	}
	for k, v := range replace {
		env = append(env, k+"="+v)
	}
	return env
}

type response struct {
	host        string
	body        []byte
	status      int
	headers     http.Header
	wire        wireResult
	diagnostics []api.Diagnostic
	err         error
}

func (a *Adapter) request(ctx context.Context, path, method string, payload []byte) response {
	args := []string{"api", "--hostname", a.config.Host, "--method", method, "--include", "--header", "Accept: application/vnd.github+json", path}
	if payload != nil {
		args = append(args, "--input", "-")
	}
	w := a.run(ctx, command{args: args, input: append([]byte(nil), payload...), mutation: method != "GET"})
	r := response{host: a.config.Host, wire: w, err: w.err, diagnostics: w.transport.Data().Diagnostics}
	// --include emits a status line and MIME header block before exactly one body.
	reader := bufio.NewReader(bytes.NewReader(w.stdout))
	line, err := reader.ReadString('\n')
	fields := strings.Fields(line)
	if err != nil || len(fields) < 2 || !strings.HasPrefix(fields[0], "HTTP/") {
		r.err = errors.Join(r.err, errors.New("missing machine HTTP envelope"))
	} else {
		r.status, err = strconv.Atoi(fields[1])
		if err != nil || r.status < 100 || r.status > 599 {
			r.status = 0
			r.err = errors.Join(r.err, errors.New("invalid HTTP status"))
		}
		h, e := textproto.NewReader(reader).ReadMIMEHeader()
		if e != nil {
			r.err = errors.Join(r.err, errors.New("invalid HTTP headers"))
		} else {
			r.headers = http.Header(h)
			r.body, _ = io.ReadAll(reader)
		}
	}
	if r.status < 200 || r.status >= 300 {
		r.err = errors.Join(r.err, errors.New("remote request unavailable"))
	}
	if r.err != nil {
		code := api.ProcessFailure
		reason := "remote-transport-failed"
		switch r.status {
		case 401, 403:
			code = api.Permission
			reason = "remote-permission-or-rate-limit"
		case 404:
			code = api.Unavailable
			reason = "remote-private-or-missing"
		case 429:
			code = api.Busy
			reason = "remote-rate-limit"
		case 422:
			code = api.Conflict
			reason = "remote-validation-rejected"
		}
		if errors.Is(w.err, context.Canceled) || errors.Is(w.err, context.DeadlineExceeded) {
			code = api.Canceled
			reason = "remote-command-canceled-or-timed-out"
		}
		d := api.RemoteDiagnosticDetailData{}
		if r.status != 0 {
			d.HTTPStatus = api.Some(uint32(r.status))
		}
		if id := r.headers.Get("X-Github-Request-Id"); id != "" && len(id) <= 256 {
			d.RequestID = api.Some(id)
		}
		if v, e := strconv.ParseUint(r.headers.Get("X-Ratelimit-Remaining"), 10, 64); e == nil {
			d.RateRemaining = api.Some(v)
		}
		if v, e := strconv.ParseUint(r.headers.Get("Retry-After"), 10, 64); e == nil {
			d.RetryAfterSeconds = api.Some(v)
		}
		if v, e := strconv.ParseInt(r.headers.Get("X-Ratelimit-Reset"), 10, 64); e == nil && v > 0 {
			d.ResetAt = api.Some(time.Unix(v, 0).UTC())
		}
		detail := must(api.NewRemoteDiagnosticDetail(d))
		r.diagnostics = append(r.diagnostics, must(api.NewDiagnostic(api.DiagnosticData{Code: code, Reason: reason, Message: reason, Detail: api.Some[api.DiagnosticDetail](detail)})))
	}
	if len(w.stderr) > 0 {
		r.diagnostics = append(r.diagnostics, diagnostic(api.ProcessFailure, "remote-command-diagnostic-stream-present"))
	}
	return r
}

func (r response) interval() api.ObservationInterval {
	return must(api.NewObservationInterval(api.ObservationIntervalData{StartedAt: r.wire.started, FinishedAt: r.wire.finished}))
}
func protocolError(what string) error { return fmt.Errorf("remote protocol: %s", what) }
