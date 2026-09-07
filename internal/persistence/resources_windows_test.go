package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func testRequestResources(t testing.TB) func(testing.TB) {
	t.Helper()
	probe, err := os.Create(filepath.Join(t.TempDir(), "file-count-reference"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := probe.Close(); err != nil {
			t.Error(err)
		}
	})
	count := func(t testing.TB) int { return windowsTestFileHandleCount(t, windows.Handle(probe.Fd())) }
	baseline := count(t)
	return func(t testing.TB) {
		t.Helper()
		if got := count(t); got != baseline {
			t.Fatalf("request leaked native file/directory/lock handles: got %d want %d", got, baseline)
		}
	}
}
