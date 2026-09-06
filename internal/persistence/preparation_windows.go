package persistence

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/sys/windows"
)

type nativeMetadata = winMetadata

func nativeNameKey(name string) string {
	// Use the same equivalence relation as nativeSameName/EqualFold. Lowercase
	// alone gives different namespace hashes for e.g. final and ordinary sigma.
	return strings.Map(func(r rune) rune {
		least := r
		for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
			if next < least {
				least = next
			}
		}
		return unicode.ToLower(least)
	}, name)
}
func nativeObjectSize(object *nativeObject) (int64, error) {
	v, err := winObserve(object.handle())
	if err != nil {
		return 0, err
	}
	if v.basic.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 || v.size() > 1<<63-1 {
		return 0, errUnsupportedProfile
	}
	return int64(v.size()), nil
}

func nativeCreateFile(parent *nativeObject, name string, userOnly bool) (*nativeObject, error) {
	var security *windows.SECURITY_DESCRIPTOR
	var err error
	if userOnly {
		security, err = winUserSecurity()
		if err != nil {
			return nil, err
		}
	}
	return winOpenWithSecurity(parent.handle(), name,
		windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER,
		winShareAll, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE, security)
}

func nativeInspectMetadata(object *nativeObject) (nativeMetadata, error) {
	return winInspectMetadata(object)
}
func nativeApplyMetadata(object *nativeObject, metadata nativeMetadata) error {
	return winApplyMetadata(object, metadata)
}

func nativeOpenOriginal(parent *nativeObject, name string) (*nativeObject, error) {
	return winOpen(parent.handle(), name, windows.FILE_GENERIC_READ|windows.DELETE, winShareAll, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
}

func nativeInspectDirectory(parent *nativeObject) error {
	v, err := winObserve(parent.handle())
	if err != nil {
		return err
	}
	if v.basic.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 ||
		v.basic.FileAttributes & ^uint32(windows.FILE_ATTRIBUTE_DIRECTORY|windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_ATTRIBUTE_ARCHIVE) != 0 {
		return errUnsupportedProfile
	}
	_, err = winInspectSecurity(parent.handle())
	return err
}

func nativeCreateDirectory(parent *nativeObject, name string, userOnly bool) (*nativeObject, error) {
	if err := nativeInspectDirectory(parent); err != nil {
		return nil, err
	}
	var security *windows.SECURITY_DESCRIPTOR
	var err error
	if userOnly {
		security, err = winUserSecurity()
		if err != nil {
			return nil, err
		}
	}
	child, err := winOpenWithSecurity(parent.handle(), name, windows.FILE_GENERIC_READ|windows.FILE_TRAVERSE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, windows.FILE_CREATE, windows.FILE_DIRECTORY_FILE, security)
	if err != nil {
		return nil, err
	}
	if err := nativeInspectDirectory(child); err != nil {
		return nil, errors.Join(err, child.close())
	}
	return child, nil
}
