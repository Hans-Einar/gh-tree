//go:build windows

package git

import (
	"encoding/binary"
	"fmt"
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
	if strings.ContainsAny(name, "\\:") {
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
	var owner *windows.Tokenuser
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
		owner = user
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
		return nil, ntError(err)
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		windows.CloseHandle(handle)
		return nil, diagnostic(api.Unsupported, "NonRegularNativeFile", "The native child is not a supported no-reparse regular file.")
	}
	if create {
		security, se := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if se == nil {
			se = privateACL(security, owner.User.Sid)
		}
		if se != nil {
			windows.CloseHandle(handle)
			return nil, se
		}
	}
	return os.NewFile(uintptr(handle), filepath.Join(d.expected.path, name)), nil
}

func (d *nativeDirectory) openChild(name string) (*nativeDirectory, error) {
	if err := nativeComponent(name); err != nil {
		return nil, err
	}
	if strings.ContainsAny(name, "\\:") || !utf8.ValidString(name) {
		return nil, diagnostic(api.Unsupported, "UnrepresentableWindowsPath", "The native child component is not representable literally.")
	}
	nativeName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return nil, err
	}
	attrs := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: d.handle, ObjectName: nativeName, Attributes: windows.OBJ_DONT_REPARSE}
	var handle windows.Handle
	var status windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, windows.FILE_READ_ATTRIBUTES|windows.FILE_LIST_DIRECTORY|windows.SYNCHRONIZE, &attrs, &status, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		return nil, ntError(err)
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(handle, &info); err != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		windows.CloseHandle(handle)
		return nil, diagnostic(api.Unsupported, "RedirectedNativeDirectory", "The native child directory is redirected or unavailable.")
	}
	identity, err := windowsObjectIdentity(handle)
	if err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}
	return &nativeDirectory{handle: handle, expected: directoryObservation{path: filepath.Join(d.expected.path, name), identity: identity}}, nil
}

func windowsObjectIdentity(handle windows.Handle) (api.DirectoryIdentity, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return api.DirectoryIdentity{}, err
	}
	var nativeID struct {
		Volume uint64
		File   [16]byte
	}
	if err := windows.GetFileInformationByHandleEx(handle, windows.FileIdInfo, (*byte)(unsafe.Pointer(&nativeID)), uint32(unsafe.Sizeof(nativeID))); err != nil {
		return api.DirectoryIdentity{}, err
	}
	stamp := "birth-filetime:" + strconv.FormatUint(uint64(info.CreationTime.HighDateTime)<<32|uint64(info.CreationTime.LowDateTime), 10)
	return api.NewDirectoryIdentity(api.DirectoryWindows, nativeID.Volume, nativeID.File, stamp)
}

func regularIdentity(f *os.File) (string, uint32, error) {
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", 0, diagnostic(api.Unsupported, "FileIdentityUnavailable", "Native regular file identity is unavailable.")
	}
	identity, err := windowsObjectIdentity(windows.Handle(f.Fd()))
	if err != nil {
		return "", 0, err
	}
	return fmt.Sprintf("%d:%x:%s", identity.Device(), identity.FileID(), identity.Stamp()), uint32(info.Mode().Perm()), nil
}

func (d *nativeDirectory) linkObservation(name string) (string, string, uint32, error) {
	if err := nativeComponent(name); err != nil {
		return "", "", 0, err
	}
	if strings.ContainsAny(name, "\\:") || !utf8.ValidString(name) {
		return "", "", 0, diagnostic(api.Unsupported, "UnrepresentableWindowsPath", "The native link name is not representable literally.")
	}
	nativeName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return "", "", 0, err
	}
	attrs := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: d.handle, ObjectName: nativeName, Attributes: windows.OBJ_DONT_REPARSE}
	var handle windows.Handle
	var status windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE, &attrs, &status, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN, windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		return "", "", 0, ntError(err)
	}
	defer windows.CloseHandle(handle)
	buffer := make([]byte, 16384)
	var returned uint32
	if err = windows.DeviceIoControl(handle, windows.FSCTL_GET_REPARSE_POINT, nil, 0, &buffer[0], uint32(len(buffer)), &returned, nil); err != nil {
		return "", "", 0, err
	}
	if returned < 20 || binary.LittleEndian.Uint32(buffer[:4]) != windows.IO_REPARSE_TAG_SYMLINK {
		return "", "", 0, diagnostic(api.Unsupported, "UnsupportedFileKind", "The native reparse object is not a supported symbolic link.")
	}
	offset, length := int(binary.LittleEndian.Uint16(buffer[8:10])), int(binary.LittleEndian.Uint16(buffer[10:12]))
	if offset%2 != 0 || length%2 != 0 || 20+offset+length > int(returned) {
		return "", "", 0, diagnostic(api.Unavailable, "MalformedSymbolicLink", "The native symbolic-link data is malformed.")
	}
	units := make([]uint16, length/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(buffer[20+offset+i*2:])
	}
	for i := 0; i < len(units); i++ {
		u := units[i]
		if u >= 0xd800 && u <= 0xdbff {
			if i+1 == len(units) || units[i+1] < 0xdc00 || units[i+1] > 0xdfff {
				return "", "", 0, diagnostic(api.Unsupported, "UnrepresentableWindowsPath", "The native symbolic-link target contains an unpaired surrogate.")
			}
			i++
		} else if u >= 0xdc00 && u <= 0xdfff {
			return "", "", 0, diagnostic(api.Unsupported, "UnrepresentableWindowsPath", "The native symbolic-link target contains an unpaired surrogate.")
		}
	}
	target := windows.UTF16ToString(units)
	target = strings.TrimPrefix(target, `\??\`)
	identity, err := windowsObjectIdentity(handle)
	if err != nil {
		return "", "", 0, err
	}
	return target, fmt.Sprintf("%d:%x:%s", identity.Device(), identity.FileID(), identity.Stamp()), 0777, nil
}

func ntError(err error) error {
	if status, ok := err.(windows.NTStatus); ok {
		return status.Errno()
	}
	return err
}
