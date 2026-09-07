package persistence

import (
	"encoding/binary"
	"testing"

	"golang.org/x/sys/unix"
)

func TestLinuxNativePOSIXACLCopy(t *testing.T) {
	parent := acquiredUnix(t, physicalTemp(t)).parent()
	source := unixTestPayload(t, parent, "source-acl", []byte("old"))
	// Version2 POSIX access ACL with an explicit named user, owner/group/mask/
	// other entries. The native kernel validates this fixture and updates mode.
	entries := []struct {
		tag, perm uint16
		id        uint32
	}{{1, 6, 0xffffffff}, {2, 4, 65533}, {4, 0, 0xffffffff}, {16, 4, 0xffffffff}, {32, 0, 0xffffffff}}
	acl := make([]byte, 4+8*len(entries))
	binary.LittleEndian.PutUint32(acl[:4], 2)
	for i, v := range entries {
		p := acl[4+8*i:]
		binary.LittleEndian.PutUint16(p[:2], v.tag)
		binary.LittleEndian.PutUint16(p[2:4], v.perm)
		binary.LittleEndian.PutUint32(p[4:8], v.id)
	}
	if err := unix.Fsetxattr(source.fd(), "system.posix_acl_access", acl, 0); err != nil {
		t.Fatal(err)
	}
	m, err := unixInspectMetadata(source)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, v := range m.attributes {
		if v.name == "system.posix_acl_access" {
			found = true
		}
	}
	if !found {
		t.Fatal("native ACL was not detected")
	}
	payload := unixTestPayload(t, parent, "payload-acl", []byte("new"))
	if err := unixApplyMetadata(payload, m); err != nil {
		t.Fatal(err)
	}
	got, err := unixInspectMetadata(payload)
	if err != nil || !m.equal(got) {
		t.Fatal("ACL/mode/ownership not preserved", err)
	}
	if unixSupportedAttribute("security.capability") || unixSupportedAttribute("trusted.policy") {
		t.Fatal("unsupported security namespace accepted")
	}
}
