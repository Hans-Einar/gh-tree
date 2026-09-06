package persistence

import (
	"encoding/binary"
	"errors"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

func unixAttributeNames(fd int) ([]string, error) {
	n, err := unix.Flistxattr(fd, nil)
	if err != nil {
		return nil, err
	}
	if n < 0 || n > maxNativeMetadata {
		return nil, errors.New("native attribute list limit")
	}
	raw := make([]byte, n)
	got, err := unix.Flistxattr(fd, raw)
	if err != nil {
		return nil, err
	}
	if got != n {
		return nil, errors.New("native attribute list changed")
	}
	return unixNullTerminatedNames(raw)
}
func unixSupportedAttribute(name string) bool {
	// User metadata copies verbatim; Apple's access/security/resource/compression
	// namespaces require their own profile and are never silently discarded.
	return strings.HasPrefix(name, "user.")
}

func unixInspectNativePolicy(fd int, st *unix.Stat_t) error {
	if st.Flags != 0 {
		return errors.New("unsupported native BSD file flags")
	}
	// fgetattrlist's returned bitmap proves this attribute was actually supplied.
	// Flistxattr alone does not expose Darwin's ACL, so it cannot prove absence.
	type attrList struct {
		count, pad                            uint16
		common, volume, directory, file, fork uint32
	}
	attrs := attrList{count: 5, common: unix.ATTR_CMN_RETURNED_ATTRS | unix.ATTR_CMN_EXTENDED_SECURITY}
	buf := make([]uint32, 16384)
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)*4), unix.FSOPT_REPORT_FULLSIZE, 0)
	runtime.KeepAlive(attrs)
	if errno != 0 {
		return errno
	}
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), len(buf)*4)
	size := binary.LittleEndian.Uint32(raw[:4])
	if size < 32 || size > uint32(len(raw)) {
		return errors.New("invalid native extended-security length")
	}
	returned := binary.LittleEndian.Uint32(raw[4:8])
	if returned&unix.ATTR_CMN_EXTENDED_SECURITY == 0 {
		return errors.New("native extended security unavailable")
	}
	offset := int64(int32(binary.LittleEndian.Uint32(raw[24:28]))) + 24
	length := int64(binary.LittleEndian.Uint32(raw[28:32]))
	if length == 0 {
		return nil
	}
	if offset < 32 || length < 44 || offset+length > int64(size) {
		return errors.New("invalid native extended-security reference")
	}
	security := raw[offset : offset+length]
	if binary.LittleEndian.Uint32(security[:4]) != 0x012cc16d {
		return errors.New("invalid native filesec magic")
	}
	// KAUTH_FILESEC_NOACL is distinct from a valid empty deny-all ACL.
	if binary.LittleEndian.Uint32(security[36:40]) != 0xffffffff {
		return errors.New("unsupported native Darwin ACL")
	}
	return nil
}
