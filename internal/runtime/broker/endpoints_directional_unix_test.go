//go:build linux || darwin

package broker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestNativeRejectsNamedAndUnlinkedDirectionalFIFO(t *testing.T) {
	name := filepath.Join(t.TempDir(), "owned-control-fifo")
	if err := unix.Mkfifo(name, 0600); err != nil {
		t.Fatal(err)
	}
	reader := must(unix.Open(name, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0))
	defer unix.Close(reader)
	writer := must(unix.Open(name, unix.O_WRONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0))
	defer unix.Close(writer)
	for _, unlinked := range []bool{false, true} {
		if unlinked {
			if err := os.Remove(name); err != nil {
				t.Fatal(err)
			}
		}
		for _, side := range []struct {
			fd    int
			write bool
		}{{reader, false}, {writer, true}} {
			fd := must(unix.FcntlInt(uintptr(side.fd), unix.F_DUPFD_CLOEXEC, 0))
			owned, err := inheritedPipe(fd, side.write)
			if owned != nil {
				owned.Close()
			} else {
				unix.Close(fd)
			}
			if !errors.Is(err, ErrProtocol) {
				t.Errorf("named FIFO admitted: unlinked=%t write=%t err=%v", unlinked, side.write, err)
			}
		}
	}
}

func TestNativeRejectsReversedAnonymousPipeDirection(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	for _, side := range []struct {
		file  *os.File
		write bool
	}{{reader, true}, {writer, false}} {
		fd := must(unix.FcntlInt(side.file.Fd(), unix.F_DUPFD_CLOEXEC, 0))
		owned, err := inheritedPipe(fd, side.write)
		if owned != nil {
			owned.Close()
		} else {
			unix.Close(fd)
		}
		if !errors.Is(err, ErrProtocol) {
			t.Fatal("reversed pipe role admitted", err)
		}
	}
}
