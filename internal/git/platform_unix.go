//go:build linux || darwin || freebsd

package git

import (
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func hideWindow(cmd *exec.Cmd) {}

func observeDirectory(path string) (directoryObservation, error) {
	final, err := filepath.EvalSymlinks(path)
	if err != nil {
		return directoryObservation{}, err
	}
	final, err = filepath.Abs(final)
	if err != nil {
		return directoryObservation{}, err
	}
	f, err := os.Open(final)
	if err != nil {
		return directoryObservation{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return directoryObservation{}, err
	}
	if !info.IsDir() {
		return directoryObservation{}, errors.New("not a directory")
	}
	s, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return directoryObservation{}, errors.New("directory identity unavailable")
	}
	// The open-object/path pairing is checked again; this is a short-lived read
	// observation, never a retained mutation capability or continuous ancestry.
	current, err := os.Stat(final)
	if err != nil || !os.SameFile(info, current) {
		return directoryObservation{}, errors.New("directory changed during observation")
	}
	var file [16]byte
	binary.LittleEndian.PutUint64(file[:8], uint64(s.Ino))
	id, err := api.NewDirectoryIdentity(api.DirectoryUnix, uint64(s.Dev), file, directoryStamp(s))
	return directoryObservation{path: final, identity: id}, err
}
