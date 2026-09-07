package persistence

import (
	"errors"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

func unixAttributeNames(fd int) ([]string, error) {
	var result []string
	// Do not use x/sys Flistxattr/FlistxattrNS: the pinned wrappers suppress system
	// namespace EPERM and a wrapper error path returns nil. Direct native errors
	// are mandatory; unreadable access metadata is not an empty profile.
	for _, ns := range []int{unix.EXTATTR_NAMESPACE_USER, unix.EXTATTR_NAMESPACE_SYSTEM} {
		n, err := unix.ExtattrListFd(fd, ns, 0, 0)
		if err != nil {
			return nil, err
		}
		if n < 0 || n > maxNativeMetadata {
			return nil, errors.New("native attribute list limit")
		}
		if n == 0 {
			continue
		}
		raw := make([]byte, n)
		got, err := unix.ExtattrListFd(fd, ns, uintptr(unsafe.Pointer(&raw[0])), len(raw))
		runtime.KeepAlive(raw)
		if err != nil {
			return nil, err
		}
		if got != n {
			return nil, errors.New("native attribute list changed")
		}
		prefix := "user."
		if ns == unix.EXTATTR_NAMESPACE_SYSTEM {
			prefix = "system."
		}
		for pos := 0; pos < len(raw); {
			size := int(raw[pos])
			pos++
			if size == 0 || size > len(raw)-pos {
				return nil, errors.New("invalid native attribute list")
			}
			result = append(result, prefix+string(raw[pos:pos+size]))
			pos += size
		}
	}
	return result, nil
}
func unixSupportedAttribute(name string) bool { return strings.HasPrefix(name, "user.") }

func unixInspectNativePolicy(fd int, st *unix.Stat_t) error {
	if st.Flags != 0 {
		return errors.New("unsupported native BSD file flags")
	}
	// FreeBSD's native ACL header/entries use fixed 32/16-bit fields on every
	// released target ABI. Query both current POSIX and NFS4 types. A nontrivial
	// supported query refuses rather than assuming st_mode proves ACL absence.
	for _, kind := range []uintptr{2, 4} {
		// Current ACL_MAX_ENTRIES is254; 24+254*16=4088 bytes on all target ABIs.
		buf := make([]uint32, 1022)
		buf[0] = 254
		_, _, errno := unix.Syscall(unix.SYS___ACL_GET_FD, uintptr(fd), kind, uintptr(unsafe.Pointer(&buf[0])))
		runtime.KeepAlive(buf)
		if errno == unix.EOPNOTSUPP {
			continue
		}
		if errno != 0 {
			return errno
		}
		if kind == 4 {
			return errors.New("unsupported native NFS4 ACL")
		}
		count := buf[1]
		if count != 3 {
			return errors.New("unsupported native POSIX ACL")
		}
		// Three ordinary owner/group/other entries must agree with mode exactly.
		want := map[uint32]uint32{1: uint32(st.Mode>>6) & 7, 4: uint32(st.Mode>>3) & 7, 32: uint32(st.Mode) & 7}
		for i := uint32(0); i < count; i++ {
			base := 6 + i*4
			tag, permission := buf[base], buf[base+2]
			mode, ok := want[tag]
			if !ok || mode != permission {
				return errors.New("nontrivial native POSIX ACL")
			}
			delete(want, tag)
		}
	}
	return nil
}
