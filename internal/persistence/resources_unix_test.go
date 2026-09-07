//go:build linux || darwin || freebsd

package persistence

import (
	"errors"
	"io"
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
		// /dev/fd entries need not support lstat (Darwin may report a
		// transient descriptor). Count names while the enumeration descriptor
		// is retained; os.ReadDir's file-info fallback is inappropriate here.
		directory, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		entries, err := directory.Readdirnames(4096)
		if errors.Is(err, io.EOF) {
			err = nil
		}
		err = errors.Join(err, directory.Close())
		if err != nil || len(entries) == 4096 {
			t.Fatalf("bounded native descriptor enumeration: count=%d error=%v", len(entries), err)
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
