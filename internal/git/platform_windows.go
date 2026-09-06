//go:build windows

package git

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func hideWindow(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }

func observeDirectory(path string) (directoryObservation, error) {
	if !utf8.ValidString(path) {
		return directoryObservation{}, errors.New("unsupported non-UTF-8 Windows directory locator")
	}
	absolute, err := nativeWindowsPath(path)
	if err != nil {
		return directoryObservation{}, err
	}
	p, err := windows.UTF16PtrFromString(absolute)
	if err != nil {
		return directoryObservation{}, err
	}
	h, err := windows.CreateFile(p, windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return directoryObservation{}, err
	}
	defer windows.CloseHandle(h)
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(h, &info); err != nil {
		return directoryObservation{}, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		return directoryObservation{}, errors.New("not a directory")
	}
	buffer := make([]uint16, 32768)
	n, err := windows.GetFinalPathNameByHandle(h, &buffer[0], uint32(len(buffer)), 0)
	if err != nil || n >= uint32(len(buffer)) {
		return directoryObservation{}, errors.New("canonical directory unavailable")
	}
	final := windows.UTF16ToString(buffer[:n])
	if strings.HasPrefix(final, `\\?\UNC\`) {
		final = `\\` + strings.TrimPrefix(final, `\\?\UNC\`)
	} else {
		final = strings.TrimPrefix(final, `\\?\`)
	}
	var nativeID struct {
		Volume uint64
		File   [16]byte
	}
	if err = windows.GetFileInformationByHandleEx(h, windows.FileIdInfo, (*byte)(unsafe.Pointer(&nativeID)), uint32(unsafe.Sizeof(nativeID))); err != nil {
		return directoryObservation{}, err
	}
	var file [16]byte
	copy(file[:], nativeID.File[:])
	stamp := "birth-filetime:" + strconv.FormatUint(uint64(info.CreationTime.HighDateTime)<<32|uint64(info.CreationTime.LowDateTime), 10)
	id, err := api.NewDirectoryIdentity(api.DirectoryWindows, nativeID.Volume, file, stamp)
	return directoryObservation{path: final, identity: id}, err
}

func nativeWindowsPath(path string) (string, error) {
	if !utf8.ValidString(path) {
		return "", errors.New("unsupported non-UTF-8 Windows path")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(absolute, `\\?\`) {
		return absolute, nil
	}
	if strings.HasPrefix(absolute, `\\`) {
		return `\\?\UNC\` + strings.TrimPrefix(absolute, `\\`), nil
	}
	return `\\?\` + absolute, nil
}
