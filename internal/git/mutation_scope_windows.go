//go:build windows

package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func (d *nativeDirectory) openScopeLock(name string) (*os.File, bool, error) {
	if err := nativeComponent(name); err != nil {
		return nil, false, err
	}
	if strings.ContainsAny(name, "\\:") || !utf8.ValidString(name) {
		return nil, false, diagnostic(api.Unsupported, "UnrepresentableWindowsPath", "The guard name cannot be represented literally.")
	}
	if err := d.validate(); err != nil {
		return nil, false, err
	}
	nativeName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, false, err
	}
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return nil, false, err
	}
	owner, err := token.GetTokenUser()
	token.Close()
	if err != nil {
		return nil, false, err
	}
	security, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;" + owner.User.Sid.String() + ")(A;;FA;;;SY)")
	if err != nil {
		return nil, false, err
	}
	attrs := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: d.handle, ObjectName: nativeName, Attributes: windows.OBJ_DONT_REPARSE, SecurityDescriptor: security}
	var handle windows.Handle
	var status windows.IO_STATUS_BLOCK
	open := func(disposition uint32) error {
		return ntError(windows.NtCreateFile(&handle, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, &attrs, &status, nil, windows.FILE_ATTRIBUTE_NORMAL, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, disposition, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0))
	}
	err = open(windows.FILE_CREATE)
	created := err == nil
	if errors.Is(err, windows.ERROR_FILE_EXISTS) || errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		err = open(windows.FILE_OPEN)
	}
	if err != nil {
		return nil, false, err
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		windows.CloseHandle(handle)
		return nil, false, diagnostic(api.Unsupported, "UnsupportedCommonGuard", "The common mutation guard is not an ordinary no-reparse file.")
	}
	if created {
		sd, e := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if e == nil {
			e = privateACL(sd, owner.User.Sid)
		}
		if e != nil {
			windows.CloseHandle(handle)
			return nil, false, e
		}
	}
	return os.NewFile(uintptr(handle), filepath.Join(d.expected.path, name)), created, nil
}

func lockScopeFile(f *os.File) error {
	return windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
}

func scopeLockContention(err error) bool {
	return errors.Is(err, windows.ERROR_LOCK_VIOLATION) || errors.Is(err, windows.ERROR_IO_PENDING)
}
