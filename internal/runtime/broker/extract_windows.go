package broker

import (
	"bytes"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowsImage owns an exclusive extracted image and guarded directory chain.
// The parent passes the pure embedded bytes/hash/machine/protocol from
// brokerassets; broker never imports that package or recursively embeds itself.
type WindowsImage struct {
	machine      uint16
	path         string
	name         string
	parent       windows.Handle
	directory    windows.Handle
	chain        []windows.Handle
	guard        *os.File
	writer       *os.File
	imageID      fileIdentity
	directoryID  fileIdentity
	imageCreated bool
	removed      bool
}

func (i *WindowsImage) Path() string {
	if i == nil || i.removed {
		return ""
	}
	return i.path
}

func createProtected(parent windows.Handle, name string, sd *windows.SECURITY_DESCRIPTOR, access, options uint32) (windows.Handle, error) {
	u, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, err
	}
	oa := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: parent, ObjectName: u, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE, SecurityDescriptor: sd}
	var h windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&h, access|windows.SYNCHRONIZE, &oa, &iosb, nil, windows.FILE_ATTRIBUTE_NORMAL, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_CREATE, options|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	return h, err
}

func (i *WindowsImage) acquireParent(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return ErrCwd
	}
	volume := filepath.VolumeName(path)
	if len(volume) != 2 || volume[1] != ':' || strings.Contains(path[2:], ":") {
		return ErrCwd
	}
	u, err := windows.UTF16PtrFromString(volume + `\`)
	if err != nil {
		return err
	}
	h, err := windows.CreateFile(u, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return err
	}
	i.chain = append(i.chain, h)
	for _, part := range strings.FieldsFunc(path[2:], func(r rune) bool { return r == '\\' || r == '/' }) {
		if strings.Contains(part, ":") || part == "." || part == ".." {
			return ErrCwd
		}
		h, err = openRelative(h, part, windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE)
		if err != nil {
			return err
		}
		i.chain = append(i.chain, h)
	}
	for _, h := range i.chain {
		if err := directoryAttributes(h); err != nil {
			return err
		}
	}
	i.parent = h
	return nil
}

// ExtractWindowsImage returns its cleanup owner even on partial acquisition.
// Callers must retain it with admitted session resources until Cleanup succeeds.
func ExtractWindowsImage(image []byte, machine uint16, digest [32]byte, protocol uint16) (*WindowsImage, error) {
	return extractWindowsImage(image, machine, digest, protocol, nil)
}

func extractWindowsImage(image []byte, machine uint16, digest [32]byte, protocol uint16, fault func(string, *WindowsImage) error) (*WindowsImage, error) {
	if len(image) > 32<<20 {
		return nil, ErrProtocol
	}
	image = append([]byte(nil), image...)
	if len(image) == 0 || len(image) > 32<<20 || protocol != ProtocolVersion || sha256.Sum256(image) != digest {
		return nil, ErrProtocol
	}
	parsed, err := pe.NewFile(bytes.NewReader(image))
	if err != nil {
		return nil, err
	}
	defer parsed.Close()
	if parsed.Machine != machine || (machine != machineAMD64 && machine != machineARM64) || parsed.Characteristics&pe.IMAGE_FILE_EXECUTABLE_IMAGE == 0 {
		return nil, ErrProtocol
	}
	// Require an actual PE32+ native image with a nonzero entry point.
	header, ok := parsed.OptionalHeader.(*pe.OptionalHeader64)
	if !ok || header.AddressOfEntryPoint == 0 {
		return nil, ErrProtocol
	}
	i := &WindowsImage{name: "broker.exe", machine: machine}
	parent := filepath.Clean(os.TempDir())
	if err = i.acquireParent(parent); err != nil {
		return i, err
	}
	if fault != nil {
		if err = fault("temporary-parent-acquired", i); err != nil {
			return i, err
		}
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return i, err
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;SY)(A;;FA;;;" + user.User.Sid.String() + ")")
	if err != nil {
		return i, err
	}
	nonce, err := FreshNonce()
	if err != nil {
		return i, err
	}
	dirName := "gh-tree-broker-" + hex.EncodeToString(nonce[:])
	i.directory, err = createProtected(i.parent, dirName, sd, windows.FILE_GENERIC_READ|windows.DELETE, windows.FILE_DIRECTORY_FILE)
	if err != nil {
		return i, err
	}
	i.path = filepath.Join(parent, dirName, i.name)
	i.directoryID, err = identity(i.directory)
	if err != nil {
		return i, err
	}
	if err = directoryAttributes(i.directory); err != nil {
		return i, err
	}
	if err = validateImageACL(i.directory, user.User.Sid); err != nil {
		return i, err
	}
	if fault != nil {
		if err = fault("helper-directory-created", i); err != nil {
			return i, err
		}
	}
	writer, err := createProtected(i.directory, i.name, sd, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return i, err
	}
	i.writer = os.NewFile(uintptr(writer), "owned-helper-write")
	i.imageCreated = true
	i.imageID, err = identity(writer)
	if err != nil {
		return i, err
	}
	if fault != nil {
		if err = fault("helper-image-created", i); err != nil {
			return i, err
		}
	}
	for left := image; len(left) > 0; {
		n, e := i.writer.Write(left)
		if e != nil {
			return i, e
		}
		if n <= 0 {
			return i, io.ErrShortWrite
		}
		left = left[n:]
	}
	if err = i.writer.Sync(); err != nil {
		return i, err
	}
	if fault != nil {
		if err = fault("helper-image-flushed", i); err != nil {
			return i, err
		}
	}
	if err = closeFile(&i.writer); err != nil {
		return i, err
	}
	if fault != nil {
		if err = fault("helper-writer-closed", i); err != nil {
			return i, err
		}
	}
	guard, err := openRelative(i.directory, i.name, windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return i, err
	}
	i.guard = os.NewFile(uintptr(guard), "owned-helper-readonly-guard")
	if fault != nil {
		if err = fault("helper-read-guard-acquired", i); err != nil {
			return i, err
		}
	}
	if err = validateImageACL(guard, user.User.Sid); err != nil {
		return i, err
	}
	actual, err := identity(guard)
	if err != nil {
		return i, err
	}
	if actual != i.imageID {
		return i, ErrCwd
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(guard, &info); err != nil {
		return i, err
	}
	if info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		return i, ErrCwd
	}
	data, err := io.ReadAll(io.LimitReader(i.guard, int64(len(image))+1))
	if err != nil {
		return i, err
	}
	if len(data) != len(image) || sha256.Sum256(data) != digest {
		return i, ErrProtocol
	}
	for _, h := range i.chain {
		if err = directoryAttributes(h); err != nil {
			return i, err
		}
	}
	if err = directoryAttributes(i.directory); err != nil {
		return i, err
	}
	if err = validateImageACL(i.directory, user.User.Sid); err != nil {
		return i, err
	}
	if fault != nil {
		if err = fault("helper-verified", i); err != nil {
			return i, err
		}
	}
	return i, nil
}

func validateImageACL(h windows.Handle, user *windows.SID) error {
	sd, err := windows.GetSecurityInfo(h, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	acl, defaulted, err := sd.DACL()
	if err != nil {
		return err
	}
	if control&windows.SE_DACL_PROTECTED == 0 || defaulted || acl == nil || acl.AceCount != 2 {
		return errors.New("helper DACL is not exact protected ownership")
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return err
	}
	foundUser, foundSystem := false, false
	for index := uint32(0); index < 2; index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err = windows.GetAce(acl, index, &ace); err != nil {
			return err
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 || ace.Mask != 0x1f01ff {
			return errors.New("unexpected helper access rule")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		isUser, isSystem := sid.Equals(user), sid.Equals(system)
		if !isUser && !isSystem {
			return errors.New("unexpected helper trustee")
		}
		if isUser {
			foundUser = true
		}
		if isSystem {
			foundSystem = true
		}
	}
	runtime.KeepAlive(sd)
	if !foundUser || !foundSystem {
		return errors.New("missing helper owner trustee")
	}
	return nil
}

func deleteExact(h windows.Handle) error {
	var status windows.IO_STATUS_BLOCK
	remove := byte(1)
	return windows.NtSetInformationFile(h, &status, &remove, 1, windows.FileDispositionInformation)
}

// Cleanup is invoked only after broker exit and outer Job-zero. It removes the
// exact acquired image and empty directory; unexpected entries are retained.
func (i *WindowsImage) Cleanup() error {
	if i == nil || i.removed {
		return nil
	}
	if err := closeFile(&i.guard); err != nil {
		return err
	}
	if err := closeFile(&i.writer); err != nil {
		return err
	}
	if i.imageCreated {
		h, err := openRelative(i.directory, i.name, windows.FILE_GENERIC_READ|windows.DELETE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
		if err != nil {
			return err
		}
		actual, idErr := identity(h)
		if idErr != nil || actual != i.imageID {
			_ = windows.CloseHandle(h)
			return errors.Join(idErr, ErrCwd)
		}
		err = deleteExact(h)
		err = errors.Join(err, windows.CloseHandle(h))
		if err != nil {
			return err
		}
		i.imageCreated = false
	}
	if i.directory != 0 {
		actual, err := identity(i.directory)
		if err != nil {
			return err
		}
		if actual != i.directoryID {
			return ErrCwd
		}
		if err = deleteExact(i.directory); err != nil {
			return err
		} // fails closed if nonempty
		if err = closeHandle(&i.directory); err != nil {
			return err
		}
	}
	var result error
	for j := len(i.chain) - 1; j >= 0; j-- {
		result = errors.Join(result, closeHandle(&i.chain[j]))
	}
	if result == nil {
		i.removed = true
		i.parent = 0
		i.chain = nil
	}
	return result
}
