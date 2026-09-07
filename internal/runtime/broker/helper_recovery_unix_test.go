//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func helperRecoveryFixture() (result error) {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	// This entry runs only from the caller's fresh, exclusively created copy.
	if filepath.Base(executable) != "owned-runtime-recovery.test" {
		return errors.New("recovery fixture image is not the owned copy")
	}
	root, err := os.MkdirTemp("", "owned-runtime-recovery-")
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, os.Remove(root)) }()
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	directory, err := os.Open(root)
	if err != nil {
		return err
	}
	identity, err := ObserveDirectory(directory, "")
	closeErr := directory.Close()
	if err != nil || closeErr != nil {
		return errors.Join(err, closeErr)
	}
	spec := testSpec()
	spec.RootLocator = root
	spec.Components = nil
	spec.RootIdentity = identity
	spec.ProjectIdentity = identity
	spec.Executable = executable
	spec.Arguments = []string{"--runtime-fixture-hold-ignore"}
	spec.Environment = append(spec.Environment, "GORACE="+os.Getenv("GORACE"))
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	c, start, err := StartUnix(ctx, UnixConfig{SessionID: 17, Spec: spec, Output: func(api.OutputStream, []byte) {}, GracePeriod: 20 * time.Millisecond, ForcePeriod: 200 * time.Millisecond})
	if err != nil || c == nil || !start.Established {
		return errors.Join(err, errors.New("recovery fixture startup failed"))
	}
	defer func() {
		restoreErr := os.Chmod(executable, 0700)
		c.Stop()
		f, e := c.Wait(ctx)
		if !f.CleanupComplete {
			result = errors.Join(result, e, errors.New("recovery fixture did not clean"))
		}
		result = errors.Join(result, restoreErr)
	}()
	if err := os.Chmod(executable, 0600); err != nil {
		return err
	}
	c.Stop()
	short, stop := context.WithTimeout(context.Background(), 500*time.Millisecond)
	f, err := c.Wait(short)
	stop()
	if !errors.Is(err, context.DeadlineExceeded) || f.CleanupComplete {
		return errors.New("unacquirable helper falsely cleaned")
	}
	var failure UnixFailure
	if !errors.As(f.Err, &failure) || failure.Code != api.Permission {
		return fmt.Errorf("helper permission failure evidence missing: %v", f.Err)
	}
	fmt.Println("PASS unacquirable helper retained native owner and Permission residual")
	if err := os.Chmod(executable, 0700); err != nil {
		return err
	}
	f, err = c.Wait(ctx)
	if !f.CleanupComplete || len(f.Residuals) != 0 {
		return fmt.Errorf("restored helper failed owned cleanup: %v", err)
	}
	if !errors.As(err, &failure) || failure.Code != api.Permission {
		return errors.New("recovered cleanup erased historical permission fact")
	}
	fmt.Println("PASS restored owned helper completed cleanup and retained historical diagnostic")
	return nil
}

func TestNativeUnixClientUnacquirableHelperCanRecover(t *testing.T) {
	config, _ := unixClientFixture(t, "unused")
	copyPath := filepath.Join(t.TempDir(), "owned-runtime-recovery.test")
	input, err := os.Open(must(os.Executable()))
	if err != nil {
		t.Fatal(err)
	}
	output, err := os.OpenFile(copyPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		input.Close()
		t.Fatal(err)
	}
	_, copyErr := io.Copy(output, input)
	closeErr := errors.Join(input.Close(), output.Close())
	if copyErr != nil || closeErr != nil {
		t.Fatal(copyErr, closeErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, copyPath, "--runtime-fixture-helper-recovery")
	cmd.Env = append(os.Environ(), "TMPDIR="+config.Spec.RootLocator)
	cmd.WaitDelay = time.Second
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native helper recovery: %v\n%s", err, data)
	}
	if !strings.Contains(string(data), "PASS unacquirable helper") || !strings.Contains(string(data), "PASS restored owned helper") {
		t.Fatal("native recovery controls omitted", string(data))
	}
}
