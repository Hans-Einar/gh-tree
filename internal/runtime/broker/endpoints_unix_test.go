//go:build linux || darwin || freebsd

package broker

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"testing"
	"time"
)

func TestNativeAnonymousPipeEndpointProfile(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	for _, side := range []struct {
		file  *os.File
		write bool
	}{{reader, false}, {writer, true}} {
		flags := must(unix.FcntlInt(side.file.Fd(), unix.F_GETFL, 0))
		var fs unix.Statfs_t
		fsErr := unix.Fstatfs(int(side.file.Fd()), &fs)
		t.Logf("native pipe write=%t access=%d fstatfs=%v", side.write, flags&unix.O_ACCMODE, fsErr)
		fd := must(unix.FcntlInt(side.file.Fd(), unix.F_DUPFD_CLOEXEC, 0))
		owned, err := inheritedPipe(fd, side.write)
		if err != nil {
			unix.Close(fd)
			t.Fatal(err)
		}
		if descriptorFlags := must(unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)); descriptorFlags&unix.FD_CLOEXEC == 0 {
			owned.Close()
			t.Fatal("private endpoint lost close-on-exec")
		}
		if err := owned.SetDeadline(time.Now().Add(5 * time.Millisecond)); err != nil {
			owned.Close()
			t.Fatal("private endpoint is not pollable", err)
		}
		if !side.write {
			_, err = owned.Read(make([]byte, 1))
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				owned.Close()
				t.Fatal("empty private reader did not honor deadline", err)
			}
		}
		if err := owned.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
