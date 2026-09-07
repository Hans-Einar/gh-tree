//go:build linux || darwin || freebsd

package persistence

import (
	"os"
	"runtime"
	"testing"

	"golang.org/x/sys/unix"
)

func testRequestResources(t testing.TB) func(testing.TB) {
	t.Helper()
	path := "/dev/fd"
	if runtime.GOOS == "linux" {
		path = "/proc/self/fd"
	}
	count := func(t testing.TB) int {
		t.Helper()
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		return len(entries)
	}
	baseline := count(t)
	fd, err := unix.Open("/dev/null", unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	extra := count(t)
	if err := unix.Close(fd); err != nil {
		t.Fatal(err)
	}
	if extra != baseline+1 || count(t) != baseline {
		t.Fatal("native descriptor counter failed deliberate open/close control")
	}
	return func(t testing.TB) {
		t.Helper()
		if got := count(t); got != baseline {
			t.Fatalf("request leaked descriptors: got %d want %d", got, baseline)
		}
	}
}
