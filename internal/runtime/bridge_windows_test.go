package runtime

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
	"github.com/Hans-Einar/gh-tree/internal/runtime/brokerassets"
	"golang.org/x/sys/windows"
)

func nativeTestPrivate() (bool, int) {
	if len(os.Args) > 1 && os.Args[1] == broker.WindowsPrivateMode {
		return true, broker.RunWindowsPrivate()
	}
	return false, 0
}
func nativeTestSize() (int, int, error) {
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &info)
	return int(info.Window.Bottom - info.Window.Top + 1), int(info.Window.Right - info.Window.Left + 1), err
}
func nativeTestDirectory(file *os.File) (api.DirectoryIdentity, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &info); err != nil {
		return api.DirectoryIdentity{}, err
	}
	birth := uint64(info.CreationTime.HighDateTime)<<32 | uint64(info.CreationTime.LowDateTime)
	return broker.ObserveDirectory(file, fmt.Sprintf("birth-filetime:%d", birth))
}
func nativeTestStarter() nativeStarter { return startWindows }
func nativeTestSuffix() string         { return ".exe" }
func nativeTestEnter() string          { return "\r" }
func nativeTestForce() time.Duration   { return 5 * time.Second }

func nativeTestInterruptedRead(err error) bool {
	return errors.Is(err, windows.ERROR_OPERATION_ABORTED)
}

func TestWindowsParentEmulatedRoutes(t *testing.T) {
	_, embedded, err := broker.MachineRoute()
	if err != nil {
		t.Fatal(err)
	}
	if embedded {
		t.Skip("emulated driver is exercised by the native parent matrix")
	}
	arches := []string{"386"}
	if goruntime.GOARCH == "arm64" {
		arches = []string{"386", "amd64"}
	}
	if goruntime.GOARCH == "386" {
		t.Skip("native 32-bit platform has no emulated parent route")
	}
	for _, arch := range arches {
		t.Run(arch, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			exe := filepath.Join(t.TempDir(), "parent-"+arch+".exe")
			cmd := exec.CommandContext(ctx, filepath.Join(goruntime.GOROOT(), "bin", "go.exe"), "test", "-c", "-o", exe, ".")
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				if !strings.EqualFold(key, "GOARCH") && !strings.EqualFold(key, "CGO_ENABLED") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "GOARCH="+arch, "CGO_ENABLED=0")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("compile %s: %v\n%s", arch, err, out)
			}
			cmd = exec.CommandContext(ctx, exe, "-test.run=^(TestNativeParent|TestWindowsParentPendingReceipt|TestWindowsParentExtractionPartialOwner)", "-test.v", "-test.timeout=60s")
			out, err := cmd.CombinedOutput()
			t.Logf("actual %s parent:\n%s", arch, out)
			if err != nil || !bytes.Contains(out, []byte("--- PASS: TestNativeParentTerminalControlAndRestart")) {
				t.Fatalf("emulated parent: %v", err)
			}
		})
	}
}

func TestWindowsParentPendingReceipt(t *testing.T) {
	request := nativeParentRequest(t, 41, false).Data()
	inv := request.Invocation.Data()
	argv := inv.Execution.(api.ArgvExecution).Data()
	argv.Arguments = []string{"--runtime-parent-blocked"}
	inv.Execution = must(api.NewArgvExecution(argv))
	request.Invocation = must(api.NewInvocation(inv))
	r := nativeParentEngine(t)
	id := startID(t, r, must(api.NewSessionStartRequest(request)))
	nativeOutput(t, r, id, "READY")
	s := must(r.registry.lookup(id))
	s.mu.Lock()
	owner := s.owner
	s.mu.Unlock()
	var delivery nativeDelivery
	var err error
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		delivery, err = owner.Write(ctx, bytes.Repeat([]byte{'x'}, nativeInputChunk))
		cancel()
		if err != nil {
			break
		}
		if !delivery.Completed || delivery.Delivered != nativeInputChunk {
			t.Fatal("known prior fill write", delivery)
		}
	}
	if err == nil || !delivery.Dispatched || delivery.Receipt == nil {
		t.Fatalf("pending native delivery: %+v / %v", delivery, err)
	}
	// The original receipt remains observable after waiting ends; no second
	// Write is sent. Stop joins the real native blocked writer and broker.
	owner.Stop()
	wait, done := context.WithTimeout(context.Background(), 8*time.Second)
	defer done()
	final, receiptErr := delivery.Receipt.Wait(wait)
	if final.Delivered > nativeInputChunk || !final.Dispatched {
		t.Fatalf("terminal receipt %+v / %v", final, receiptErr)
	}
	again, againErr := delivery.Receipt.Wait(wait)
	if again.Accepted != final.Accepted || again.Delivered != final.Delivered || again.Completed != final.Completed || (againErr == nil) != (receiptErr == nil) {
		t.Fatal("receipt observation changed")
	}
	stop, _ := r.Stop(wait, stopRequest(id))
	if !stop.Data().CleanupComplete {
		t.Fatal("pending writer cleanup", stop.Data())
	}
	drainNativeFinals(t, r, 1)
}

func TestWindowsParentExtractionPartialOwner(t *testing.T) {
	asset, err := brokerassets.Load("arm64")
	if err != nil {
		t.Skip("native ARM64 embeds no image; the mandatory emulated parent runs this case")
	}
	request := nativeParentRequest(t, 51, false)
	root := t.TempDir()
	blocked := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("retained control"), 0600); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(blocked, "child")
	t.Setenv("TMP", bad)
	if filepath.Clean(os.TempDir()) != bad {
		t.Fatal("temporary-path fixture did not select its owned control")
	}
	image, extractErr := broker.ExtractWindowsImage(asset.Bytes, asset.Machine, sha256.Sum256(asset.Bytes), asset.Protocol)
	if image == nil || extractErr == nil {
		t.Fatal("fixture did not acquire partial extraction", extractErr)
	}
	n := nativeStart{ID: sessionID(51), OperationID: request.Data().OperationID, Invocation: request.Data().Invocation}
	owner, _, err := ownWindowsExtraction(n, image, extractErr)
	if owner == nil || err == nil {
		t.Fatal("extraction owner abandoned")
	}
	owner.Stop()
	fact, err := owner.NextFact(context.Background())
	if err != nil || !fact.CleanupComplete || len(fact.Residuals) != 0 {
		t.Fatal("partial extraction did not close exact resources", fact, err)
	}
	if b, err := os.ReadFile(blocked); err != nil || string(b) != "retained control" {
		t.Fatal("fixture original changed", err)
	}
}

func TestWindowsBridgeFailureClassification(t *testing.T) {
	err := windowsError(errors.Join(&broker.WindowsFailure{Cause: broker.WindowsPermissionFailure, Stage: api.HelperExtraction, Cleanup: true}, &broker.WindowsFailure{Cause: broker.WindowsProcessFailure, Stage: api.UserProcessWait}), api.Acquisition)
	values := diagnostics(err)
	if len(values) != 2 || values[0].Data().Code != api.Permission || values[1].Data().Code != api.ProcessFailure {
		t.Fatal(values, err)
	}
	resized := mapWindowsDelivery(broker.WindowsDelivery{Completed: true, Dispatched: true}, true, api.TerminalCleanup)
	if resized.Delivered != 1 || resized.Accepted != 0 {
		t.Fatal("control unit mapping", resized)
	}
}
