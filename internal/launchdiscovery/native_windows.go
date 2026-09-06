package launchdiscovery

import (
	"encoding/binary"
	"fmt"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func nativeComponent(s string) bool {
	return !strings.ContainsRune(s, ':') && !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, " ")
}
func nativePermission(e error) bool {
	return e == windows.STATUS_ACCESS_DENIED || e == windows.ERROR_ACCESS_DENIED || os.IsPermission(e)
}
func nativeRedirect(e error) bool {
	return e == windows.STATUS_REPARSE_POINT_ENCOUNTERED || e == windows.STATUS_NOT_A_DIRECTORY || e == windows.STATUS_FILE_IS_A_DIRECTORY
}
func nativeMissing(e error) bool {
	return e == windows.STATUS_OBJECT_NAME_NOT_FOUND || e == windows.STATUS_OBJECT_PATH_NOT_FOUND || e == windows.ERROR_FILE_NOT_FOUND || e == windows.ERROR_PATH_NOT_FOUND || os.IsNotExist(e)
}
func nativeRoot(path string) (*os.File, []*os.File, error) {
	// A canonical Git root may use the extended DOS prefix; UNC/device/network
	// scopes are outside this initial local native observation profile.
	p := strings.TrimPrefix(path, `\\?\`)
	vol := filepath.VolumeName(p)
	if len(vol) != 2 || vol[1] != ':' || !filepath.IsAbs(p) || filepath.Clean(p) != p {
		return nil, nil, errRedirect
	}
	u, e := windows.UTF16PtrFromString(vol + `\`)
	if e != nil {
		return nil, nil, e
	}
	h, e := windows.CreateFile(u, windows.FILE_LIST_DIRECTORY|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if e != nil {
		return nil, nil, e
	}
	current := os.NewFile(uintptr(h), vol+`\`)
	chain := []*os.File{}
	for _, name := range strings.Split(strings.TrimPrefix(p[len(vol):], `\`), `\`) {
		if name == "" {
			continue
		}
		next, e := nativeChild(current, name, true)
		if e != nil {
			current.Close()
			for _, f := range chain {
				f.Close()
			}
			return nil, nil, e
		}
		chain = append(chain, current)
		current = next
	}
	return current, chain, nil
}
func nativeChild(parent *os.File, name string, dir bool) (*os.File, error) {
	if !component(name) {
		return nil, errRedirect
	}
	u, e := windows.NewNTUnicodeString(name)
	if e != nil {
		return nil, e
	}
	oa := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: windows.Handle(parent.Fd()), ObjectName: u, Attributes: windows.OBJ_DONT_REPARSE}
	var ios windows.IO_STATUS_BLOCK
	var h windows.Handle
	access := uint32(windows.FILE_READ_DATA | windows.FILE_READ_ATTRIBUTES | windows.SYNCHRONIZE)
	opts := uint32(windows.FILE_OPEN_REPARSE_POINT | windows.FILE_SYNCHRONOUS_IO_NONALERT)
	share := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE)
	if dir {
		opts |= windows.FILE_DIRECTORY_FILE
		share = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE
	}
	if e = windows.NtCreateFile(&h, access, &oa, &ios, nil, 0, share, windows.FILE_OPEN, opts, 0, 0); e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(h), name)
	var info windows.ByHandleFileInformation
	if e = windows.GetFileInformationByHandle(h, &info); e != nil {
		f.Close()
		return nil, e
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (dir != (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0)) {
		f.Close()
		return nil, errRedirect
	}
	return f, nil
}
func observeIdentity(f *os.File, profile string) (api.DirectoryIdentity, error) {
	if profile != "" && !strings.HasPrefix(profile, "birth-filetime:") {
		return api.DirectoryIdentity{}, errRedirect
	}
	h := windows.Handle(f.Fd())
	var info windows.ByHandleFileInformation
	if e := windows.GetFileInformationByHandle(h, &info); e != nil {
		return api.DirectoryIdentity{}, e
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return api.DirectoryIdentity{}, errRedirect
	}
	// uint64 storage guarantees FileIdInfo alignment even for a 386 caller.
	words := make([]uint64, 4)
	base := uintptr(unsafe.Pointer(&words[0]))
	offset := (8 - base%8) % 8
	b := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(&words[0]), offset)), 24)
	if e := windows.GetFileInformationByHandleEx(h, windows.FileIdInfo, &b[0], 24); e != nil {
		return api.DirectoryIdentity{}, e
	}
	var id [16]byte
	copy(id[:], b[8:])
	birth := uint64(info.CreationTime.HighDateTime)<<32 | uint64(info.CreationTime.LowDateTime)
	return api.NewDirectoryIdentity(api.DirectoryWindows, binary.LittleEndian.Uint64(b[:8]), id, fmt.Sprintf("birth-filetime:%d", birth))
}
