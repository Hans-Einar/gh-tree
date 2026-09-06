//go:build windows

package git

import (
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

// Compare native ACE/SID data; SDDL spelling may replace the actual current
// user's SID with an alias (for example LA or SY) on another native host.
func privateACL(sd *windows.SECURITY_DESCRIPTOR, user *windows.SID) error {
	bad := func() error {
		return diagnostic(api.Permission, "PrivateACLNotEstablished", "The native private file ACL is not limited to the current user and System.")
	}
	control, _, err := sd.Control()
	if err != nil {
		return err
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		return bad()
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	if acl == nil || acl.AceCount == 0 || acl.AceCount > 2 {
		return bad()
	}
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return err
	}
	seenUser, seenSystem := false, false
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(acl, i, &ace); err != nil {
			return err
		}
		if ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceFlags != 0 || ace.Header.AceSize < 16 || uint32(ace.Mask) != 0x001f01ff {
			return bad()
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !sid.IsValid() || sid.Len()+8 > int(ace.Header.AceSize) {
			return bad()
		}
		isUser, isSystem := sid.Equals(user), sid.Equals(system)
		if !isUser && !isSystem {
			return bad()
		}
		seenUser = seenUser || isUser
		seenSystem = seenSystem || isSystem
	}
	if !seenUser || !seenSystem {
		return bad()
	}
	return nil
}
