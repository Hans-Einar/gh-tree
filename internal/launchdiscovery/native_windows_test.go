package launchdiscovery

import (
	"context"
	"encoding/binary"
	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"testing"
)

// A native mount-point junction requires no symbolic-link privilege. Only this
// fixture builder writes reparse metadata, on test-owned temporary objects.
func junction(t *testing.T, path, target string) error {
	t.Helper()
	if e := os.Mkdir(path, 0700); e != nil && !os.IsExist(e) {
		return e
	}
	sub := must(windows.UTF16FromString(`\??\` + target))
	print := must(windows.UTF16FromString(target))
	data := make([]byte, 16+2*(len(sub)+len(print)))
	binary.LittleEndian.PutUint32(data, windows.IO_REPARSE_TAG_MOUNT_POINT)
	binary.LittleEndian.PutUint16(data[4:], uint16(len(data)-8))
	binary.LittleEndian.PutUint16(data[10:], uint16(2*(len(sub)-1)))
	binary.LittleEndian.PutUint16(data[12:], uint16(2*len(sub)))
	binary.LittleEndian.PutUint16(data[14:], uint16(2*(len(print)-1)))
	offset := 16
	for _, s := range append(sub, print...) {
		binary.LittleEndian.PutUint16(data[offset:], s)
		offset += 2
	}
	h, e := windows.CreateFile(must(windows.UTF16PtrFromString(path)), windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if e != nil {
		return e
	}
	defer windows.CloseHandle(h)
	var returned uint32
	return windows.DeviceIoControl(h, windows.FSCTL_SET_REPARSE_POINT, &data[0], uint32(len(data)), nil, 0, &returned, nil)
}
func TestWindowsNativeJunctionAndConversion(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	put(t, outside, "package.json", `{"scripts":{"external":"x"}}`)
	put(t, root, "package.json", `{"scripts":{"local":"x"}}`)
	if e := junction(t, filepath.Join(root, "linked"), outside); e != nil {
		t.Fatal(e)
	}
	if e := junction(t, filepath.Join(root, "pnpm-lock.yaml"), outside); e != nil {
		t.Fatal(e)
	}
	a := must(New(Config{}))
	scope := fixtureScope(t, root)
	r := discover(t, a, scope)
	if len(r.Definitions) != 1 {
		t.Fatal(r.Definitions)
	}
	def := pick(t, r.Definitions, api.Npm, "", "local")
	if def.Data().EffectiveExecutable.Data().Executable != "npm" {
		t.Fatal("redirected lock qualified")
	}
	entry := savedEntry("external", "npm", "linked", "external", "", nil)
	saved := discover(t, a, scope, entry)
	if saved.Saved[0].Data().Definition.Present() {
		t.Fatal("saved junction escaped")
	}
	// Convert an already opened empty directory. If native guards allow the
	// conversion, the next relative observation and identity check must refuse.
	if e := os.Mkdir(filepath.Join(root, "empty"), 0700); e != nil {
		t.Fatal(e)
	}
	d, e := acquireRoot(scope)
	if e != nil {
		t.Fatal(e)
	}
	defer d.close()
	child, e := childDirectory(d, "empty")
	if e != nil {
		t.Fatal(e)
	}
	defer child.close()
	e = junction(t, filepath.Join(root, "empty"), outside)
	if e == nil {
		if e = sameDirectory(child); e == nil {
			t.Fatal("converted identity accepted")
		}
		if o, e := observeFile(context.Background(), child, "package.json", true, 4096); e == nil && o.state == "regular" {
			t.Fatal("converted parent exposed external source")
		}
	} else {
		t.Logf("native directory guard refused conversion: %v", e)
	}
}
func TestWindowsRootReplacementRefuses(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	put(t, root, "package.json", `{"scripts":{"dev":"x"}}`)
	scope := fixtureScope(t, root)
	a := must(New(Config{}))
	def := pick(t, discover(t, a, scope).Definitions, api.Npm, "", "dev")
	if e := os.Rename(root, filepath.Join(base, "old")); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	put(t, root, "package.json", `{"scripts":{"dev":"x"}}`)
	out, e := resolve(t, a, scope, must(api.NewDiscoveredLaunch(api.DiscoveredLaunchData{Member: memberOf(def)})), nil, api.None[api.StorageVersion]())
	if e == nil || out.Data().Invocation.Present() {
		t.Fatal("replaced root accepted")
	}
}
