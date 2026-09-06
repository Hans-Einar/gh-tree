package broker

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeFreeBSDRejectsNamedAndUnlinkedFIFO(t *testing.T) {
	name := filepath.Join(t.TempDir(), "owned-control-fifo")
	if err := unix.Mkfifo(name, 0600); err != nil {
		t.Fatal(err)
	}
	fd := must(unix.Open(name, unix.O_RDWR|unix.O_NONBLOCK|unix.O_CLOEXEC, 0))
	defer unix.Close(fd)
	if _, err := inheritedPipe(fd, false); !errors.Is(err, ErrProtocol) {
		t.Fatal("named FIFO accepted", err)
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	if _, err := inheritedPipe(fd, false); !errors.Is(err, ErrProtocol) {
		t.Fatal("unlinked named FIFO accepted", err)
	}
}
