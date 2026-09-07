package broker

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

var ErrCwd = errors.New("Runtime cwd observation is stale, redirected or unsupported")

// AcquiredDirectory retains real data/list-read handles to every effective
// ancestor and an exclusively created nonempty child until the debug barrier.
// Its private native handle never crosses the Application boundary.
type AcquiredDirectory struct {
	chain          []windows.Handle
	root, target   windows.Handle
	anchor         windows.Handle
	anchorName     string
	anchorIdentity fileIdentity
	anchorOwned    bool
	spec           StartSpec
	path           string
}

func (a *AcquiredDirectory) Path() string { return a.path }

func (a *AcquiredDirectory) Close() error {
	var result error
	if a.anchor != 0 {
		// DELETE_ON_CLOSE refers to this retained owned object, never its name.
		result = errors.Join(result, closeHandle(&a.anchor))
	}
	for i := len(a.chain) - 1; i >= 0; i-- {
		result = errors.Join(result, closeHandle(&a.chain[i]))
	}
	if result == nil {
		a.root, a.target = 0, 0
		a.chain = nil
	}
	return result
}

func AcquireCwd(spec StartSpec) (*AcquiredDirectory, error) { return acquireCwd(spec, nil) }

func acquireCwd(spec StartSpec, barrier func(string)) (*AcquiredDirectory, error) {
	if !spec.valid() || spec.RootIdentity.Platform() != api.DirectoryWindows || !filepath.IsAbs(spec.RootLocator) || filepath.Clean(spec.RootLocator) != spec.RootLocator {
		return nil, ErrCwd
	}
	volume := filepath.VolumeName(spec.RootLocator)
	// Local drive roots are the supported observation profile. Device namespaces
	// and UNC shares need separate verified filesystem identity/interlock profiles.
	if len(volume) != 2 || volume[1] != ':' || strings.ContainsAny(spec.RootLocator[2:], ":") {
		return nil, ErrCwd
	}
	a := &AcquiredDirectory{spec: spec, path: spec.RootLocator}
	base, err := windows.UTF16PtrFromString(volume + `\`)
	if err != nil {
		return a, err
	}
	h, err := windows.CreateFile(base, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return a, err
	}
	a.chain = append(a.chain, h)
	if err = directoryAttributes(h); err != nil {
		return a, err
	}
	parts := strings.FieldsFunc(spec.RootLocator[len(volume):], func(r rune) bool { return r == '\\' || r == '/' })
	for _, name := range parts {
		h, err = a.child(h, name)
		if err != nil {
			return a, err
		}
	}
	a.root = h
	if barrier != nil {
		barrier("root-acquired")
	}
	if err = matchesDirectory(h, spec.RootIdentity); err != nil {
		return a, err
	}
	for _, name := range spec.Components {
		if strings.ContainsAny(name, ":") || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
			return a, ErrCwd
		}
		h, err = a.child(h, name)
		if err != nil {
			return a, err
		}
		a.path = filepath.Join(a.path, name)
	}
	a.target = h
	if barrier != nil {
		barrier("project-acquired")
	}
	if err = a.pinExisting(); err != nil {
		return a, err
	}
	if a.anchor == 0 {
		n, e := FreshNonce()
		if e != nil {
			return a, e
		}
		a.anchorName = ".gh-tree-start-" + hex.EncodeToString(n[:])
		a.anchor, err = openRelative(a.target, a.anchorName, windows.FILE_GENERIC_READ|windows.DELETE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE|windows.FILE_DELETE_ON_CLOSE)
		if err != nil {
			return a, err
		}
		a.anchorOwned = true
	}
	a.anchorIdentity, err = identity(a.anchor)
	if err != nil {
		return a, err
	}
	if barrier != nil {
		barrier("anchor-acquired")
	}
	return a, a.Revalidate()
}

// Enumerate through the acquired directory handle, with a fixed 64KiB bound.
// Entry names are observations only; each candidate is independently acquired
// handle-relative with actual data/list-read access and no delete sharing.
func (a *AcquiredDirectory) pinExisting() error {
	var aligned [8192]uint64
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(&aligned[0])), 65536)
	err := windows.GetFileInformationByHandleEx(a.target, windows.FileIdBothDirectoryRestartInfo, &buffer[0], uint32(len(buffer)))
	if err == windows.ERROR_NO_MORE_FILES {
		return nil
	}
	if err != nil {
		return err
	}
	for offset, entries := 0, 0; entries < 256; entries++ {
		if offset > len(buffer)-104 {
			return ErrCwd
		}
		entry := buffer[offset:]
		next := int(binary.LittleEndian.Uint32(entry))
		length := int(binary.LittleEndian.Uint32(entry[60:]))
		if length%2 != 0 || length < 0 || length > len(entry)-104 {
			return ErrCwd
		}
		units := make([]uint16, length/2)
		for j := range units {
			units[j] = binary.LittleEndian.Uint16(entry[104+j*2:])
		}
		name := windows.UTF16ToString(units)
		if name != "" && name != "." && name != ".." && !strings.ContainsAny(name, ":/\\\x00") {
			h, e := openRelative(a.target, name, windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_OPEN, 0)
			if e == nil {
				var info windows.ByHandleFileInformation
				e = windows.GetFileInformationByHandle(h, &info)
				if e == nil && info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
					a.anchor = h
					a.anchorName = name
					return nil
				}
				if e = windows.CloseHandle(h); e != nil {
					return e
				}
			}
		}
		if next == 0 {
			return nil
		}
		if next < 104 || next > len(entry) {
			return ErrCwd
		}
		offset += next
	}
	return nil
}

func (a *AcquiredDirectory) child(parent windows.Handle, name string) (windows.Handle, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, ":/\\\x00") {
		return 0, ErrCwd
	}
	h, err := openRelative(parent, name, windows.FILE_GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_OPEN, windows.FILE_DIRECTORY_FILE)
	if err != nil {
		return 0, err
	}
	a.chain = append(a.chain, h)
	return h, directoryAttributes(h)
}

func directoryAttributes(h windows.Handle) error {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return ErrCwd
	}
	return nil
}

func matchesDirectory(h windows.Handle, expected api.DirectoryIdentity) error {
	v, err := observeHandle(h)
	if err != nil {
		return err
	}
	if !v.Equal(expected) {
		return ErrCwd
	}
	return nil
}

func (a *AcquiredDirectory) Revalidate() error {
	if a.root == 0 || a.target == 0 || a.anchor == 0 {
		return ErrCwd
	}
	for _, h := range a.chain {
		if err := directoryAttributes(h); err != nil {
			return err
		}
	}
	if err := matchesDirectory(a.root, a.spec.RootIdentity); err != nil {
		return err
	}
	if err := matchesDirectory(a.target, a.spec.ProjectIdentity); err != nil {
		return err
	}
	id, err := identity(a.anchor)
	if err != nil {
		return err
	}
	if id != a.anchorIdentity {
		return ErrCwd
	}
	return nil
}

func observeHandle(h windows.Handle) (api.DirectoryIdentity, error) {
	if err := directoryAttributes(h); err != nil {
		return api.DirectoryIdentity{}, err
	}
	id, err := identity(h)
	if err != nil {
		return api.DirectoryIdentity{}, err
	}
	var info windows.ByHandleFileInformation
	if err = windows.GetFileInformationByHandle(h, &info); err != nil {
		return api.DirectoryIdentity{}, err
	}
	birth := uint64(info.CreationTime.HighDateTime)<<32 | uint64(info.CreationTime.LowDateTime)
	return api.NewDirectoryIdentity(api.DirectoryWindows, id.Volume, id.ID, fmt.Sprintf("birth-filetime:%d", birth))
}

func ObserveDirectory(file *os.File, profile string) (api.DirectoryIdentity, error) {
	if file == nil {
		return api.DirectoryIdentity{}, ErrCwd
	}
	v, err := observeHandle(windows.Handle(file.Fd()))
	if err == nil && v.Stamp() != profile {
		return api.DirectoryIdentity{}, ErrCwd
	}
	return v, err
}
