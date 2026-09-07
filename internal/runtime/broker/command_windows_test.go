package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestWindowsBatchCarrierLiteralArguments(t *testing.T) {
	s := windowsSpec(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s.Environment = append(os.Environ(), "GH_TREE_OWNED_ARGV_FIXTURE="+exe)
	s.Executable = filepath.Join(s.RootLocator, "owned shim 世界.cmd")
	if err = os.WriteFile(s.Executable, []byte("@\"%GH_TREE_OWNED_ARGV_FIXTURE%\" -test.run=^TestWindowsOwnedUserFixture$ -- --owned-windows-fixture pipe %*\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	want := []string{"", "space value", `tail\`, `!&()`, "colon:slash/世界", " leading and trailing "}
	s.Arguments = want
	var mu sync.Mutex
	var output bytes.Buffer
	config := WindowsConfig{SessionID: 41, Spec: s, Image: exe, Output: func(_ api.OutputStream, data []byte) { mu.Lock(); output.Write(data); mu.Unlock() }}
	if _, emulated, e := MachineRoute(); e != nil {
		t.Fatal(e)
	} else if emulated {
		config.Extraction = extractedNativeFixture(t)
		config.Image = config.Extraction.Path()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, start, err := StartWindows(ctx, config)
	if client != nil {
		defer client.Stop()
	}
	if err != nil || !start.Established {
		t.Fatalf("batch start %+v %v", start, err)
	}
	final, err := client.Wait(ctx)
	if err != nil || !final.CleanupComplete || final.ExitCode != 0 {
		t.Fatalf("batch final %+v %v", final, err)
	}
	mu.Lock()
	text := bytes.TrimSpace(append([]byte(nil), output.Bytes()...))
	mu.Unlock()
	var got struct {
		Cwd  string
		Args []string
	}
	if err = json.Unmarshal(bytes.TrimPrefix(text, []byte("FIXTURE_READY ")), &got); err != nil {
		t.Fatalf("batch output %q: %v", text, err)
	}
	if got.Cwd != s.RootLocator || !reflect.DeepEqual(got.Args, want) {
		t.Fatalf("batch literal transport: %+v want=%q", got, want)
	}
	for _, bad := range []string{"embedded\"quote", "percent%", "line\nfeed", "carriage\rreturn"} {
		s.Arguments = []string{bad}
		if _, _, err = resolveWindowsCommand(s, s.RootLocator); err == nil {
			t.Fatalf("unsafe batch operand accepted: %q", bad)
		}
	}
}
