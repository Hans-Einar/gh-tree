//go:build linux || darwin || freebsd

package runtime

import (
	"errors"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
	"github.com/creack/pty"
	"os"
	"strings"
	"testing"
	"time"
)

func nativeTestPrivate() (bool, int)    { return broker.RunUnixPrivate() }
func nativeTestSize() (int, int, error) { return pty.Getsize(os.Stdin) }
func nativeTestDirectory(file *os.File) (api.DirectoryIdentity, error) {
	return broker.ObserveDirectory(file, "")
}
func nativeTestStarter() nativeStarter { return startUnix }
func nativeTestSuffix() string         { return "" }
func nativeTestEnter() string          { return "\n" }
func nativeTestForce() time.Duration   { return time.Second }

func nativeTestInterruptedRead(error) bool { return false }

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
