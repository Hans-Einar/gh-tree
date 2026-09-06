package persistence

import (
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsNativeMetadataCopy(t *testing.T) {
	parent := acquiredWindows(t, t.TempDir()).parent()
	for i, sddl := range []string{"D:P(A;;FA;;;OW)", "D:AI(A;;FA;;;OW)", "D:P(A;;FA;;;OW)S:(ML;;NW;;;LW)", "D:P(D;;GW;;;WD)(A;;FA;;;OW)"} {
		t.Run(sddl, func(t *testing.T) {
			p := winTestPayload(t, parent, "source-"+string(rune('a'+i)), []byte("source"))
			resetTestSecurity(t, p)
			sd, err := windows.SecurityDescriptorFromString(sddl)
			if err != nil {
				t.Fatal(err)
			}
			dacl, _, err := sd.DACL()
			if err != nil {
				t.Fatal(err)
			}
			control, _, _ := sd.Control()
			flags := windows.SECURITY_INFORMATION(windows.DACL_SECURITY_INFORMATION)
			if control&windows.SE_DACL_PROTECTED != 0 {
				flags |= windows.PROTECTED_DACL_SECURITY_INFORMATION
			} else {
				flags |= windows.UNPROTECTED_DACL_SECURITY_INFORMATION
			}
			if err := windows.SetSecurityInfo(p.handle(), windows.SE_FILE_OBJECT, flags, nil, nil, dacl, nil); err != nil {
				t.Fatal(err)
			}
			if sacl, _, err := sd.SACL(); err == nil {
				if err := windows.SetSecurityInfo(p.handle(), windows.SE_FILE_OBJECT, windows.LABEL_SECURITY_INFORMATION, nil, nil, nil, sacl); err != nil {
					t.Fatal(err)
				}
			}
			m, err := winInspectMetadata(p)
			if err != nil {
				t.Fatal(err)
			}
			payload := winTestPayload(t, parent, "copy-"+string(rune('a'+i)), []byte("copy"))
			resetTestSecurity(t, payload)
			if err := winApplyMetadata(payload, m); err != nil {
				t.Fatal(err)
			}
			got, err := winInspectMetadata(payload)
			if err != nil || !m.equal(got) {
				t.Fatal("policy not retained", err)
			}
		})
	}
}

func resetTestSecurity(t *testing.T, p *winObject) {
	t.Helper()
	t.Cleanup(func() {
		sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;OW)")
		if err != nil {
			t.Error(err)
			return
		}
		acl, _, err := sd.DACL()
		if err != nil {
			t.Error(err)
			return
		}
		if err := windows.SetSecurityInfo(p.handle(), windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestWindowsNativeResourceAttributeRefusal(t *testing.T) {
	parent := acquiredWindows(t, t.TempDir()).parent()
	p := winTestPayload(t, parent, "resource", []byte("original"))
	sd, err := windows.SecurityDescriptorFromString(`S:(RA;;;;;WD;("Project",TS,0,"gh-tree"))`)
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.SACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetSecurityInfo(p.handle(), windows.SE_FILE_OBJECT, windows.ATTRIBUTE_SECURITY_INFORMATION, nil, nil, nil, acl); err != nil {
		t.Fatal(err)
	}
	if _, err := winInspectMetadata(p); err == nil {
		t.Fatal("resource attribute policy accepted")
	}
	if string(winTestRead(t, parent, "resource")) != "original" {
		t.Fatal("unsupported metadata refusal changed bytes")
	}
}

func TestWindowsNativeMetadataRefusesReadOnlyAndStreams(t *testing.T) {
	root := t.TempDir()
	parent := acquiredWindows(t, root).parent()
	p := winTestPayload(t, parent, "run.json", []byte("old"))
	path, _ := windows.UTF16PtrFromString(filepath.Join(root, "run.json"))
	if err := windows.SetFileAttributes(path, windows.FILE_ATTRIBUTE_READONLY); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { windows.SetFileAttributes(path, windows.FILE_ATTRIBUTE_NORMAL) })
	if _, err := winInspectMetadata(p); err == nil {
		t.Fatal("read-only accepted")
	}
	if err := windows.SetFileAttributes(path, windows.FILE_ATTRIBUTE_NORMAL); err != nil {
		t.Fatal(err)
	}
	streamName, _ := windows.UTF16PtrFromString(filepath.Join(root, "run.json:extra"))
	stream, err := windows.CreateFile(streamName, windows.FILE_GENERIC_WRITE, winShareAll, nil, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	var wrote uint32
	err = windows.WriteFile(stream, []byte("stream"), &wrote, nil)
	closeErr := windows.CloseHandle(stream)
	if err != nil || closeErr != nil || wrote != 6 {
		t.Fatal(err, closeErr, wrote)
	}
	if _, err := winInspectMetadata(p); err == nil {
		t.Fatal("alternate stream accepted")
	}
}

func TestWindowsNativeProtectedUserCreation(t *testing.T) {
	parent := acquiredWindows(t, t.TempDir()).parent()
	sd, err := winUserSecurity()
	if err != nil {
		t.Fatal(err)
	}
	p, err := winOpenWithSecurity(parent.handle(), "private", windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER, winShareAll, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE, sd)
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
	m, err := winInspectSecurity(p.handle())
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := m.sd.DACL()
	if err != nil || dacl == nil || dacl.AceCount != 1 || m.control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatal("not a protected single-user DACL", err)
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(dacl, 0, &ace); err != nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || uint32(ace.Mask) != 0x1f01ff {
		t.Fatal("wrong sole ACE", err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil || !(*windows.SID)(unsafe.Pointer(&ace.SidStart)).Equals(user.User.Sid) {
		t.Fatal("sole ACE does not identify current user", err)
	}
}
