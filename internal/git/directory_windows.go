//go:build windows

package git

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

type nativeDirectory struct {
	handle   windows.Handle
	expected directoryObservation
}

func acquireDirectory(expected directoryObservation) (*nativeDirectory, error) {
	path, err := nativeWindowsPath(expected.path)
	if err != nil {
		return nil, err
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.FILE_READ_ATTRIBUTES|windows.FILE_LIST_DIRECTORY|windows.SYNCHRONIZE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	d := &nativeDirectory{handle: handle, expected: expected}
	if err = d.validate(); err != nil {
		d.close()
		return nil, err
	}
	return d, nil
}

func (d *nativeDirectory) validate() error {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(d.handle, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return diagnostic(api.Unsupported, "RedirectedNativeDirectory", "The acquired directory is not an ordinary no-reparse object.")
	}
	var nativeID struct {
		Volume uint64
		File   [16]byte
	}
	if err := windows.GetFileInformationByHandleEx(d.handle, windows.FileIdInfo, (*byte)(unsafe.Pointer(&nativeID)), uint32(unsafe.Sizeof(nativeID))); err != nil {
		return err
	}
	stamp := "birth-filetime:" + strconv.FormatUint(uint64(info.CreationTime.HighDateTime)<<32|uint64(info.CreationTime.LowDateTime), 10)
	actual, err := api.NewDirectoryIdentity(api.DirectoryWindows, nativeID.Volume, nativeID.File, stamp)
	if err != nil {
		return err
	}
	if actual != d.expected.identity {
		return diagnostic(api.StaleObservation, "DirectoryIdentityChanged", "The acquired directory differs from the confirmed physical observation.")
	}
	return nil
}

func (d *nativeDirectory) close() error { return windows.CloseHandle(d.handle) }

func (d *nativeDirectory) openRegular(name string) (*os.File, error) {
	return d.relativeFile(name, false)
}
func (d *nativeDirectory) createPrivate(name string) (*os.File, error) {
	return d.relativeFile(name, true)
}

func (d *nativeDirectory) relativeFile(name string, create bool) (*os.File, error) {
	if err := nativeComponent(name); err != nil {
		return nil, err
	}
	if !utf8.ValidString(name) {
		return nil, diagnostic(api.Unsupported, "UnrepresentableWindowsPath", "The native Windows component is not representable without replacing bytes.")
	}
	if strings.ContainsRune(name, ':') {
		return nil, diagnostic(api.Unsupported, "AlternateDataStreamPath", "Native Windows alternate data-stream operands are outside this file profile.")
	}
	if err := d.validate(); err != nil {
		return nil, err
	}
	nativeName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	attributes := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: d.handle, ObjectName: nativeName, Attributes: windows.OBJ_DONT_REPARSE}
	access := uint32(windows.FILE_GENERIC_READ)
	disposition := uint32(windows.FILE_OPEN)
	if create {
		access = windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE
		disposition = windows.FILE_CREATE
		token, err := windows.OpenCurrentProcessToken()
		if err != nil {
			return nil, err
		}
		user, err := token.GetTokenUser()
		token.Close()
		if err != nil {
			return nil, err
		}
		descriptor, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;" + user.User.Sid.String() + ")(A;;FA;;;SY)")
		if err != nil {
			return nil, err
		}
		attributes.SecurityDescriptor = descriptor
	}
	var handle windows.Handle
	var status windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, access, &attributes, &status, nil, windows.FILE_ATTRIBUTE_NORMAL, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, disposition, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		windows.CloseHandle(handle)
		return nil, diagnostic(api.Unsupported, "NonRegularNativeFile", "The native child is not a supported no-reparse regular file.")
	}
	return os.NewFile(uintptr(handle), filepath.Join(d.expected.path, name)), nil
}
