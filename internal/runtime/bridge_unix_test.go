//go:build linux || darwin || freebsd

package runtime

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
	"github.com/creack/pty"
)

func TestMain(m *testing.M) {
	if handled, code := broker.RunUnixPrivate(); handled {
		os.Exit(code)
	}
	if len(os.Args) == 2 && os.Args[1] == "--runtime-parent-fixture" {
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, syscall.SIGINT)
		go func() {
			for range interrupts {
				fmt.Println("INTERRUPT-OBSERVED")
			}
		}()
		cwd, err := os.Getwd()
		if err != nil {
			os.Exit(41)
		}
		marker, err := os.ReadFile("marker")
		if err != nil {
			os.Exit(42)
		}
		fmt.Printf("READY cwd=%s marker=%s env=%s\n", cwd, marker, os.Getenv("PARENT_TEST"))
		fmt.Fprint(os.Stderr, "STDERR-READY\n")
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) != 0 {
				fmt.Printf("DELIVERED:%s", line)
				if string(line) == "size\n" {
					rows, cols, e := pty.Getsize(os.Stdin)
					fmt.Printf("GEOMETRY:%d,%d,%v\n", rows, cols, e)
				}
			}
			if err != nil {
				os.Exit(0)
			}
		}
	}
	os.Exit(m.Run())
}

func nativeUnixRequest(t *testing.T, operation uint64, terminal bool) api.SessionStartRequest {
	t.Helper()
	t.Setenv("GORACE", strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"))
	root := must(filepath.EvalSymlinks(t.TempDir()))
	project := filepath.Join(root, " project")
	if err := os.Mkdir(project, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "marker"), []byte("selected-original"), 0600); err != nil {
		t.Fatal(err)
	}
	// Resolve a literal relative executable only inside the acquired directory.
	exe := must(os.Executable())
	input := must(os.Open(exe))
	output := must(os.OpenFile(filepath.Join(project, "user-fixture"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700))
	_, copyErr := io.Copy(output, input)
	if err := errors.Join(copyErr, input.Close(), output.Close()); err != nil {
		t.Fatal(err)
	}
	rootFile, projectFile := must(os.Open(root)), must(os.Open(project))
	rootID, projectID := must(broker.ObserveDirectory(rootFile, "")), must(broker.ObserveDirectory(projectFile, ""))
	if err := errors.Join(rootFile.Close(), projectFile.Close()); err != nil {
		t.Fatal(err)
	}
	request := engineRequest(operation, terminal).Data()
	inv := request.Invocation.Data()
	cwd := inv.Cwd.Data()
	scope := cwd.Worktree.Data()
	scope.RootLocator, scope.RootIdentity = root, rootID
	cwd.Worktree = must(api.NewWorktreeScope(scope))
	cwd.ProjectComponents, cwd.ProjectIdentity = []string{" project"}, projectID
	inv.Cwd = must(api.NewCwdObservation(cwd))
	inv.Execution = must(api.NewArgvExecution(api.ArgvExecutionData{Executable: "./user-fixture", Arguments: []string{"--runtime-parent-fixture"}}))
	inv.Environment = must(api.NewEnvironmentPolicy(api.EnvironmentPolicyData{Set: []api.EnvironmentEntry{
		must(api.NewEnvironmentEntry(api.EnvironmentEntryData{Name: "PARENT_TEST", Value: "literal value"})),
		must(api.NewEnvironmentEntry(api.EnvironmentEntryData{Name: "GORACE", Value: os.Getenv("GORACE")})),
	}}))
	request.Invocation = must(api.NewInvocation(inv))
	return must(api.NewSessionStartRequest(request))
}

func nativeUnixEngine(t *testing.T) *sessions {
	t.Helper()
	r := newSessions(startUnix, os.Environ(), sessionBudgets{grace: 30 * time.Millisecond, force: time.Second, shutdown: 8 * time.Second})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result := r.Shutdown(ctx).Data()
		for _, session := range result.Sessions {
			if !session.Data().CleanupComplete {
				t.Errorf("native cleanup incomplete: %+v", session.Data())
			}
		}
	})
	return r
}

func nativeOutput(t *testing.T, r *sessions, id domain.SessionID, marker string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(2 * time.Millisecond)
	defer ticker.Stop()
	for {
		result := must(r.ReadOutput(ctx, must(api.NewSessionOutputRequest(api.SessionOutputRequestData{SessionID: id, MaxBytes: 262144})))).Data()
		var data []byte
		for _, chunk := range result.Chunks {
			data = append(data, chunk.Data().Bytes...)
		}
		if bytes.Contains(data, []byte(marker)) {
			return data
		}
		select {
		case <-ctx.Done():
			t.Fatalf("missing output %q; observed %q", marker, data)
		case <-ticker.C:
		}
	}
}

func drainNativeFinals(t *testing.T, r *sessions, count int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cursor := cursor(0)
	seen := make(map[domain.SessionID]bool)
	for len(seen) < count {
		event := must(r.NextEvent(ctx, cursor)).Data()
		if event.Kind == api.RuntimeCleaned {
			if seen[event.SessionID] {
				t.Fatal("duplicate final")
			}
			seen[event.SessionID] = true
		}
		cursor = event.Sequence
		if err := r.AckEvents(cursor); err != nil {
			t.Fatal(err)
		}
	}
	if result := r.Shutdown(ctx).Data(); !result.Complete {
		t.Fatalf("acknowledged shutdown: %+v", result)
	}
	if _, err := r.NextEvent(ctx, cursor); !errors.Is(err, io.EOF) {
		t.Fatal("event EOF", err)
	}
}

func TestNativeParentUnixPipesAndPersistentStart(t *testing.T) {
	r := nativeUnixEngine(t)
	request := nativeUnixRequest(t, 11, false)
	ctx, cancel := context.WithCancel(context.Background())
	start := must(r.Start(ctx, request)).Data()
	if !start.Established {
		t.Fatal("not established")
	}
	id := mustValue(start.Session).Data().SessionID
	cancel()
	data := nativeOutput(t, r, id, "STDERR-READY")
	if !bytes.Contains(data, []byte("marker=selected-original env=literal value")) {
		t.Fatalf("cwd/env: %q", data)
	}
	snapshot := must(r.Snapshot(context.Background(), id)).Data()
	if snapshot.Phase != api.Running || !snapshot.AcquiredCwd.Present() {
		t.Fatalf("healthy native ownership stopped: %+v", snapshot)
	}
	if got := must(r.List(context.Background(), must(api.NewSessionFilter(api.SessionFilterData{})))).Data(); len(got.Sessions) != 1 || got.Sessions[0].Data().SessionID != id {
		t.Fatal("list", got)
	}
	if result := must(r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: []byte("literal \x00\xff bytes\n")})))).Data(); result.AcceptedBytes != 17 {
		t.Fatal("input admission", result)
	}
	nativeOutput(t, r, id, "DELIVERED:literal \x00\xff bytes\n")
	stop := must(r.Stop(context.Background(), stopRequest(id))).Data()
	if !stop.CleanupComplete || !stop.Session.Data().Exit.Present() {
		t.Fatal("cleanup", stop)
	}
	if stop.Session.Data().StartOperation != request.Data().OperationID {
		t.Fatal("operation attribution")
	}
	drainNativeFinals(t, r, 1)
}

func TestNativeParentUnixTerminalControlAndRestart(t *testing.T) {
	r := nativeUnixEngine(t)
	request := nativeUnixRequest(t, 21, true)
	id := startID(t, r, request)
	nativeOutput(t, r, id, "READY")
	geometry := must(api.NewGeometry(api.GeometryData{Rows: 37, Columns: 109}))
	if result := must(r.Resize(context.Background(), must(api.NewSessionResizeRequest(api.SessionResizeRequestData{SessionID: id, Geometry: geometry})))).Data(); !result.Delivered {
		t.Fatal("resize did not preserve native success")
	}
	must(r.Write(context.Background(), must(api.NewSessionWriteRequest(api.SessionWriteRequestData{SessionID: id, Bytes: []byte("size\n")}))))
	nativeOutput(t, r, id, "GEOMETRY:37,109,<nil>")
	if result := must(r.Interrupt(context.Background(), id)).Data(); !result.Delivered {
		t.Fatal("interrupt delivery")
	}
	nativeOutput(t, r, id, "INTERRUPT-OBSERVED")
	restarted := must(r.Restart(context.Background(), must(api.NewSessionRestartRequest(api.SessionRestartRequestData{OperationID: must(api.NewOperationID(22)), SessionID: id})))).Data()
	if !restarted.Old.Data().CleanupComplete || !restarted.Replacement.Present() {
		t.Fatal("restart", restarted)
	}
	replacement := mustValue(restarted.Replacement).Data()
	newSnapshot := mustValue(replacement.Session).Data()
	if !replacement.Established || newSnapshot.SessionID == id || newSnapshot.Display.Data().Geometry != geometry {
		t.Fatal("replacement facts", newSnapshot)
	}
	nativeOutput(t, r, newSnapshot.SessionID, "READY")
	must(r.Stop(context.Background(), stopRequest(newSnapshot.SessionID)))
	drainNativeFinals(t, r, 2)
}

func TestNativeParentUnixFailedStartPreservesCodeAndIdentity(t *testing.T) {
	r := nativeUnixEngine(t)
	request := nativeUnixRequest(t, 31, false).Data()
	inv := request.Invocation.Data()
	inv.Execution = must(api.NewArgvExecution(api.ArgvExecutionData{Executable: "./absent-owned-fixture"}))
	request.Invocation = must(api.NewInvocation(inv))
	result, err := r.Start(context.Background(), must(api.NewSessionStartRequest(request)))
	if err == nil || result.Data().Established || !result.Data().Session.Present() {
		t.Fatal("failed-start facts", result, err)
	}
	found := false
	for _, d := range result.Data().Diagnostics {
		found = found || d.Data().Code == api.NotFound
	}
	if !found {
		t.Fatal("lost native NotFound", result.Data().Diagnostics, err)
	}
	id := mustValue(result.Data().Session).Data().SessionID
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	stop, _ := r.Stop(ctx, stopRequest(id))
	if !stop.Data().CleanupComplete {
		t.Fatal("failed acquisition retained resources", stop.Data())
	}
	drainNativeFinals(t, r, 1)
}

func TestUnixBridgePreservesIndependentFailureCodes(t *testing.T) {
	err := unixError(errors.Join(broker.UnixFailure{Code: api.NotFound, Stage: api.ProcessContainment}, broker.UnixFailure{Code: api.Permission, Stage: api.CwdAcquisition}, &os.PathError{Op: "read", Path: "private-argument-secret", Err: os.ErrPermission}), api.OutputCleanup)
	values := diagnostics(err)
	if len(values) != 3 || strings.Contains(err.Error(), "private-argument-secret") {
		t.Fatal(values, err)
	}
	s := &session{diagnostics: make(map[api.RuntimeCleanupStage]api.Diagnostic)}
	s.retainNativeDiagnosticsLocked(values)
	s.retainNativeDiagnosticsLocked(values)
	if len(s.diagnosticsLocked()) != 3 {
		t.Fatal("native facts duplicated or lost")
	}
	if values[0].Data().Code != api.NotFound || values[1].Data().Code != api.Permission || values[2].Data().Code != api.Permission {
		t.Fatal("native causes lost", values)
	}
}
