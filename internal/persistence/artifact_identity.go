package persistence

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// Keep the original four-field disk shape, including old frame hashes. The
// private Stamp tag selects artifact validation; it is never a directory profile.
const winObjectIDStamp = "objectid-v1:"

func (v diskIdentity) artifactValid() bool {
	if strings.HasPrefix(v.Stamp, winObjectIDStamp) {
		raw, err := hex.DecodeString(strings.TrimPrefix(v.Stamp, winObjectIDStamp))
		return err == nil && len(raw) == 64 && hex.EncodeToString(raw) == strings.TrimPrefix(v.Stamp, winObjectIDStamp) &&
			v.Platform == api.DirectoryWindows && v.File != ([16]byte{}) && *(*[16]byte)(raw[:16]) != ([16]byte{})
	}
	return v.valid()
}

func verifyArtifactIdentity(object *nativeObject, recorded diskIdentity) error {
	actual, err := nativeArtifactIdentityAs(object, recorded)
	if err != nil {
		return err
	}
	if actual != recorded {
		return errors.Join(errBindingChanged, errors.New("recorded artifact native identity changed"))
	}
	return nil
}
