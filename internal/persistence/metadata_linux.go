package persistence

import (
	"errors"
	"strings"

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
	return strings.HasPrefix(name, "user.") || name == "system.posix_acl_access"
}

func unixInspectNativePolicy(fd int, st *unix.Stat_t) error {
	// Linux ACLs are enumerated/read through the native system.posix_acl_access
	// xattr. Security/trusted/unknown returned namespaces are explicit refusals.
	flags, err := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
	if err != nil {
		return err
	}
	// EXTENTS is an allocation implementation detail. Any other inode flag
	// requires a separately proved copy/security/durability profile.
	if uint32(flags)&^uint32(0x00080000) != 0 {
		return errors.New("unsupported native inode flags")
	}
	return nil
}
