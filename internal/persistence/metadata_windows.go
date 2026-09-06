package persistence

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The selected ordinary-account profile preserves access policy, not audit-only
// SACL replication. READ_CONTROL does not prove audit ACE absence. In addition
// to DACL/owner/group, explicitly query every access-affecting SACL class; only
// mandatory labels are currently copy-supported. Unreadable or nonempty CAP,
// resource attributes, trust labels and access filters refuse publication.
const winAccessSecurity windows.SECURITY_INFORMATION = windows.OWNER_SECURITY_INFORMATION |
	windows.GROUP_SECURITY_INFORMATION | windows.DACL_SECURITY_INFORMATION |
	windows.LABEL_SECURITY_INFORMATION | windows.ATTRIBUTE_SECURITY_INFORMATION |
	windows.SCOPE_SECURITY_INFORMATION | 0x80 | 0x100

var winQueryFileInformation = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtQueryInformationFile")

type winMetadata struct {
	sd           *windows.SECURITY_DESCRIPTOR
	owner, group string
	dacl, label  []byte
	control      windows.SECURITY_DESCRIPTOR_CONTROL
	attributes   uint32
}

func winACLBytes(acl *windows.ACL) ([]byte, error) {
	if acl == nil {
		return nil, nil
	}
	// Header padding/free bytes are not policy. Preserve revision, ACE ordering,
	// every ACE flag and exact complete ACE bytes; never normalize with SDDL.
	result := []byte{*(*byte)(unsafe.Pointer(acl))}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			return nil, err
		}
		if ace == nil || ace.Header.AceSize < 4 {
			return nil, errors.New("invalid native ACL")
		}
		result = append(result, unsafe.Slice((*byte)(unsafe.Pointer(ace)), int(ace.Header.AceSize))...)
	}
	runtime.KeepAlive(acl)
	return result, nil
}

func winInspectSecurity(handle windows.Handle) (winMetadata, error) {
	var m winMetadata
	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, winAccessSecurity)
	if err != nil {
		return m, err
	}
	m.sd = sd
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		return m, errors.Join(err, errors.New("missing native owner"))
	}
	m.owner = owner.String()
	group, _, err := sd.Group()
	if err != nil || group == nil {
		return m, errors.Join(err, errors.New("missing native group"))
	}
	m.group = group.String()
	m.control, _, err = sd.Control()
	if err != nil {
		return m, err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return m, err
	}
	m.dacl, err = winACLBytes(dacl)
	if err != nil {
		return m, err
	}
	sacl, _, err := sd.SACL()
	if err != nil && !errors.Is(err, windows.ERROR_OBJECT_NOT_FOUND) {
		return m, err
	}
	if sacl != nil {
		for i := uint32(0); i < uint32(sacl.AceCount); i++ {
			var ace *windows.ACCESS_ALLOWED_ACE
			if err := windows.GetAce(sacl, i, &ace); err != nil {
				return m, err
			}
			if ace == nil || ace.Header.AceType != 0x11 {
				return m, errors.New("unsupported access-affecting SACL profile")
			}
		}
	}
	m.label, err = winACLBytes(sacl)
	return m, err
}

func winInspectMetadata(object *winObject) (winMetadata, error) {
	var m winMetadata
	before, err := winObserve(object.handle())
	if err != nil {
		return m, err
	}
	// Special, read-only, EFS, compression, sparse/cloud/recall and unrecognized
	// attributes are unsupported before any destination-changing operation.
	if before.basic.FileAttributes & ^uint32(windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_ATTRIBUTE_ARCHIVE) != 0 {
		return m, errors.New("unsupported or read-only file attributes")
	}
	if err := winInspectStreams(object.handle()); err != nil {
		return m, err
	}
	var eaSize uint32
	var ioStatus windows.IO_STATUS_BLOCK
	status, _, _ := winQueryFileInformation.Call(uintptr(object.handle()), uintptr(unsafe.Pointer(&ioStatus)), uintptr(unsafe.Pointer(&eaSize)), unsafe.Sizeof(eaSize), 7 /* FileEaInformation */)
	if native := windows.NTStatus(status); native != 0 {
		return m, errors.Join(native, native.Errno())
	}
	if eaSize != 0 {
		return m, errors.New("unsupported native extended attributes")
	}
	m, err = winInspectSecurity(object.handle())
	if err != nil {
		return m, err
	}
	m.attributes = before.basic.FileAttributes
	after, err := winObserve(object.handle())
	if err != nil {
		return m, err
	}
	if !before.sameRead(after) {
		return m, errors.New("file changed during metadata observation")
	}
	confirmed, err := winInspectSecurity(object.handle())
	if err != nil {
		return m, err
	}
	if !m.equal(confirmed) {
		return m, errors.New("security changed during metadata observation")
	}
	return m, nil
}

func winInspectStreams(handle windows.Handle) error {
	// One bounded native stream query; truncation/unsupported queries refuse.
	buf := make([]uint64, 8192)
	err := windows.GetFileInformationByHandleEx(handle, windows.FileStreamInfo, (*byte)(unsafe.Pointer(&buf[0])), uint32(len(buf)*8))
	if err != nil {
		return err
	}
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), len(buf)*8)
	next, size := binary.LittleEndian.Uint32(raw[:4]), binary.LittleEndian.Uint32(raw[4:8])
	if next != 0 || size != 14 {
		return errors.New("unsupported alternate stream profile")
	}
	name := make([]uint16, 7)
	for i := range name {
		name[i] = binary.LittleEndian.Uint16(raw[24+i*2 : 26+i*2])
	}
	if windows.UTF16ToString(name) != "::$DATA" {
		return errors.New("unsupported alternate stream name")
	}
	return nil
}

func (m winMetadata) equal(n winMetadata) bool {
	const controls = windows.SE_DACL_PRESENT | windows.SE_DACL_PROTECTED | windows.SE_DACL_AUTO_INHERIT_REQ | windows.SE_DACL_AUTO_INHERITED |
		windows.SE_SACL_PRESENT | windows.SE_SACL_PROTECTED | windows.SE_SACL_AUTO_INHERIT_REQ | windows.SE_SACL_AUTO_INHERITED
	return m.owner == n.owner && m.group == n.group && bytes.Equal(m.dacl, n.dacl) && bytes.Equal(m.label, n.label) && m.control&controls == n.control&controls
}

func winApplyMetadata(payload *winObject, m winMetadata) error {
	if m.sd == nil {
		return errors.New("missing source security")
	}
	owner, _, err := m.sd.Owner()
	if err != nil {
		return err
	}
	group, _, err := m.sd.Group()
	if err != nil {
		return err
	}
	dacl, _, err := m.sd.DACL()
	if err != nil {
		return err
	}
	flags := windows.SECURITY_INFORMATION(windows.OWNER_SECURITY_INFORMATION | windows.GROUP_SECURITY_INFORMATION | windows.DACL_SECURITY_INFORMATION)
	if m.control&windows.SE_DACL_PROTECTED != 0 {
		flags |= windows.PROTECTED_DACL_SECURITY_INFORMATION
	} else {
		flags |= windows.UNPROTECTED_DACL_SECURITY_INFORMATION
	}
	if err := windows.SetSecurityInfo(payload.handle(), windows.SE_FILE_OBJECT, flags, owner, group, dacl, nil); err != nil {
		return err
	}
	sacl, _, err := m.sd.SACL()
	if err != nil && !errors.Is(err, windows.ERROR_OBJECT_NOT_FOUND) {
		return err
	}
	if sacl != nil {
		if err := windows.SetSecurityInfo(payload.handle(), windows.SE_FILE_OBJECT, windows.LABEL_SECURITY_INFORMATION, nil, nil, nil, sacl); err != nil {
			return err
		}
	}
	got, err := winInspectSecurity(payload.handle())
	if err != nil {
		return err
	}
	if !m.equal(got) {
		return fmt.Errorf("native security copy verification failed: control %x/%x", m.control, got.control)
	}
	return nil
}

func winUserSecurity() (*windows.SECURITY_DESCRIPTOR, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	// Set at exclusive creation time, before any sensitive bytes are written.
	return windows.SecurityDescriptorFromString("D:P(A;;FA;;;" + user.User.Sid.String() + ")")
}
