//go:build linux || darwin || freebsd

package broker

import (
	"golang.org/x/sys/unix"
	"os"
	"testing"
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
		if err := owned.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
