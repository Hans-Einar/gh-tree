package persistence

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

type nativeMetadata = winMetadata
type nativeStoreLock = winStoreLock

const nativePreparedBytesGuarded = true

func nativeLock(ctx context.Context, parent *nativeObject, basename string, wait time.Duration) (*nativeStoreLock, error) {
	return winLock(ctx, parent, basename, wait)
}
func nativeLockForStore(ctx context.Context, parent *nativeObject, basename string, wait time.Duration, userOnly bool) (*nativeStoreLock, error) {
	var security *windows.SECURITY_DESCRIPTOR
	var err error
	if userOnly {
		security, err = winUserSecurity()
		if err != nil {
			return nil, err
		}
	}
	return winLockSecurity(ctx, parent, basename, wait, true, security)
}
func nativeLinkCount(object *nativeObject) (uint64, error) {
	v, err := winObserve(object.handle())
	return uint64(v.basic.NumberOfLinks), err
}
func nativeExistingLock(ctx context.Context, parent *nativeObject, basename string, wait time.Duration) (*nativeStoreLock, error) {
	return winLockMode(ctx, parent, basename, wait, false)
}
func nativePublish(payload, parent *nativeObject, name, target string, present bool) error {
	return winPublish(payload, parent, target, present)
}
func nativeDirectoryBarrier(parent *nativeObject) error  { return nil }
func nativePublicationDurability() api.StorageDurability { return api.DurabilityUncertain }
func nativeDirectoryIdentityAs(object *nativeObject, profile api.DirectoryIdentity) (api.DirectoryIdentity, error) {
	return nativeDirectoryIdentity(object)
}
func nativeAppendCreated(c *nativeChain, child *nativeObject, name string) {
	c.guards = append(c.guards, child)
	c.remaining = c.remaining[1:]
}
func nativeAdoptDirectory(parent *nativeObject, name string) (*nativeObject, error) {
	child, err := winOpenDirectory(parent.handle(), name)
	if err != nil {
		return nil, err
	}
	if err := nativeInspectDirectory(child); err != nil {
		return nil, errors.Join(err, child.close())
	}
	return child, nil
}
func nativeRetainOriginal(original, parent *nativeObject, target, name string) (*nativeObject, error) {
	return winRetainOriginal(original, parent, name)
}

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
	return nativeCreateFileMetadata(parent, name, userOnly, nil)
}
func nativeCreateFileMetadata(parent *nativeObject, name string, userOnly bool, metadata *nativeMetadata) (*nativeObject, error) {
	return winCreateArtifact(parent, name, userOnly, metadata, winShareAll)
}

func nativeCreatePayloadMetadata(parent *nativeObject, name string, userOnly bool, metadata *nativeMetadata) (*nativeObject, error) {
	// Exclusion begins with FILE_CREATE, before another handle or writable
	// section can exist. Keep this exact creator open through publication: a
	// later no-write-sharing reopen could not revoke an earlier writable map.
	// DELETE sharing permits the retained-handle hardlink/rename protocol.
	return winCreateArtifact(parent, name, userOnly, metadata, windows.FILE_SHARE_READ|windows.FILE_SHARE_DELETE)
}

func winCreateArtifact(parent *nativeObject, name string, userOnly bool, metadata *nativeMetadata, share uint32) (*nativeObject, error) {
	var security *windows.SECURITY_DESCRIPTOR
	var err error
	if userOnly {
		security, err = winUserSecurity()
		if err != nil {
			return nil, err
		}
	}
	if metadata != nil {
		security = metadata.sd
	}
	object, err := winOpenWithSecurity(parent.handle(), name,
		windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE|windows.WRITE_DAC|windows.WRITE_OWNER,
		share, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE, security)
	if err != nil {
		return nil, err
	}
	// This path exclusively created this owned regular artifact. Initialize its
	// native identity before any bytes, journal/data flush, hardlink or publication.
	id, err := winArtifactObjectID(object.handle(), windows.FSCTL_CREATE_OR_GET_OBJECT_ID)
	if err != nil {
		return nil, errors.Join(errors.New("exclusively created artifact remains without initialized identity: "+name), err, object.close())
	}
	object.createdArtifact, err = winSelectArtifactIdentity(object.observation, id, nil)
	if err != nil {
		return nil, errors.Join(err, object.close())
	}
	return object, nil
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
