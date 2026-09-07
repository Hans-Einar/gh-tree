package persistence

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

// Only nativeCreateFileMetadata calls CREATE_OR_GET, after exclusive FILE_CREATE.
// All observation and recorded-profile validation calls use GET only.
func winArtifactObjectID(handle windows.Handle, code uint32) ([64]byte, error) {
	var id [64]byte
	var n uint32
	err := windows.DeviceIoControl(handle, code, nil, 0, &id[0], uint32(len(id)), &n, nil)
	return winArtifactObjectIDResult(id, n, err)
}

func winArtifactObjectIDResult(id [64]byte, n uint32, err error) ([64]byte, error) {
	if err != nil {
		return [64]byte{}, err
	}
	if n != 64 || *(*[16]byte)(id[:16]) == ([16]byte{}) {
		return [64]byte{}, errors.New("invalid native artifact object-ID buffer")
	}
	return id, nil
}

func winArtifactObservation(object *nativeObject) (winObservation, error) {
	v, err := winObserve(object.handle())
	if err == nil && v.basic.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		err = errUnsupportedProfile
	}
	return v, err
}

func winBirthArtifact(v winObservation) diskIdentity {
	birth := uint64(v.basic.CreationTime.HighDateTime)<<32 | uint64(v.basic.CreationTime.LowDateTime)
	return diskIdentity{api.DirectoryWindows, v.id.Volume, v.id.File, fmt.Sprintf("birth-filetime:%d", birth)}
}

func winSelectArtifactIdentity(v winObservation, id [64]byte, err error) (diskIdentity, error) {
	// This is the exact result of GET on an already acquired regular handle.
	// A wrapped/joined pathname absence or arbitrary error is not this exception.
	if err == windows.ERROR_FILE_NOT_FOUND {
		return winBirthArtifact(v), nil
	}
	if err != nil {
		return diskIdentity{}, err
	}
	return diskIdentity{api.DirectoryWindows, v.id.Volume, v.id.File, winObjectIDStamp + hex.EncodeToString(id[:])}, nil
}

func nativeArtifactIdentity(object *nativeObject) (diskIdentity, error) {
	if object.createdArtifact != (diskIdentity{}) {
		if err := verifyArtifactIdentity(object, object.createdArtifact); err != nil {
			return diskIdentity{}, err
		}
		return object.createdArtifact, nil
	}
	v, err := winArtifactObservation(object)
	if err != nil {
		return diskIdentity{}, err
	}
	id, err := winArtifactObjectID(object.handle(), windows.FSCTL_GET_OBJECT_ID)
	return winSelectArtifactIdentity(v, id, err)
}

func nativeArtifactIdentityAs(object *nativeObject, recorded diskIdentity) (diskIdentity, error) {
	if !recorded.artifactValid() || recorded.Platform != api.DirectoryWindows {
		return diskIdentity{}, errors.New("invalid recorded Windows artifact profile")
	}
	v, err := winArtifactObservation(object)
	if err != nil {
		return diskIdentity{}, err
	}
	if strings.HasPrefix(recorded.Stamp, winObjectIDStamp) {
		id, err := winArtifactObjectID(object.handle(), windows.FSCTL_GET_OBJECT_ID)
		if err != nil {
			return diskIdentity{}, err // Recorded ObjectID profiles never fall back.
		}
		return winSelectArtifactIdentity(v, id, nil)
	}
	if !strings.HasPrefix(recorded.Stamp, "birth-filetime:") {
		return diskIdentity{}, errors.New("unknown recorded Windows artifact profile")
	}
	// Legacy records remain exact, even when the object now has an ID. Do not
	// issue GET, rewrite a frame, or replace its historical observation token.
	return winBirthArtifact(v), nil
}
