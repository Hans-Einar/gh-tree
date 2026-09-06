package persistence

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestWindowsArtifactIdentityExactNativeResults(t *testing.T) {
	var id [64]byte
	for i := range id {
		id[i] = byte(i + 1)
	}
	for _, n := range []uint32{0, 16, 63, 65} {
		if _, err := winArtifactObjectIDResult(id, n, nil); err == nil {
			t.Fatalf("accepted output length %d", n)
		}
	}
	if _, err := winArtifactObjectIDResult([64]byte{}, 64, nil); err == nil {
		t.Fatal("accepted zero ObjectID")
	}
	if got, err := winArtifactObjectIDResult(id, 64, nil); err != nil || got != id {
		t.Fatal("lost opaque native bytes", err)
	}
	// Optional native birth fields may all be zero; only the ObjectID is required.
	var optional [64]byte
	optional[0] = 1
	if _, err := winArtifactObjectIDResult(optional, 64, nil); err != nil {
		t.Fatal(err)
	}
	v := winObservation{id: winFileIDInfo{Volume: 9, File: [16]byte{1}}}
	for _, cause := range []error{windows.ERROR_ACCESS_DENIED, windows.ERROR_SHARING_VIOLATION, windows.ERROR_NOT_SUPPORTED, windows.ERROR_INVALID_HANDLE, windows.ERROR_INVALID_PARAMETER, windows.ERROR_PATH_NOT_FOUND, fmt.Errorf("wrapped: %w", windows.ERROR_FILE_NOT_FOUND), os.ErrNotExist} {
		_, got := winArtifactObjectIDResult(id, 64, cause)
		if got != cause {
			t.Fatal("native failure was replaced")
		}
		if _, err := winSelectArtifactIdentity(v, [64]byte{}, got); err != cause {
			t.Fatal("arbitrary error became birth fallback", cause, err)
		}
	}
	if got, err := winSelectArtifactIdentity(v, [64]byte{}, windows.ERROR_FILE_NOT_FOUND); err != nil || got != winBirthArtifact(v) {
		t.Fatal("exact native absence did not select strict birth", err)
	}
	if _, err := winArtifactObjectID(windows.InvalidHandle, windows.FSCTL_GET_OBJECT_ID); err != windows.ERROR_INVALID_HANDLE {
		t.Fatal("invalid handle error", err)
	}
}

func TestWindowsArtifactIdentityFullTupleAndReadOnlyProfiles(t *testing.T) {
	parent := acquiredWindows(t, physicalStoreTemp(t)).parent()
	owned, err := nativeCreateFile(parent, "owned", true)
	if err != nil {
		t.Fatal(err)
	}
	defer owned.close()
	var short [63]byte
	var returned uint32
	if err := windows.DeviceIoControl(owned.handle(), windows.FSCTL_GET_OBJECT_ID, nil, 0, &short[0], uint32(len(short)), &returned, nil); err != windows.ERROR_INVALID_PARAMETER || returned != 0 {
		t.Fatal("native malformed-buffer reply", err, returned)
	}
	if size, err := nativeObjectSize(owned); err != nil || size != 0 {
		t.Fatal("identity initialization wrote data", err)
	}
	recorded, err := nativeArtifactIdentity(owned)
	if err != nil || !strings.HasPrefix(recorded.Stamp, winObjectIDStamp) {
		t.Fatal("owned creation lacks intrinsic profile", err)
	}
	if _, err := recorded.directory(); err == nil {
		t.Fatal("artifact ID admitted as directory profile")
	}
	check := func(want diskIdentity) {
		t.Helper()
		if err := verifyArtifactIdentity(owned, want); err == nil {
			t.Fatal("altered tuple admitted", want)
		}
	}
	volume := recorded
	volume.Device++
	check(volume)
	for i := range recorded.File {
		changed := recorded
		changed.File[i] ^= 0x80
		check(changed)
	}
	raw, _ := hex.DecodeString(strings.TrimPrefix(recorded.Stamp, winObjectIDStamp))
	for i := range raw {
		altered := bytes.Clone(raw)
		altered[i] ^= 0x80
		changed := recorded
		changed.Stamp = winObjectIDStamp + hex.EncodeToString(altered)
		check(changed)
	}
	for _, stamp := range []string{winObjectIDStamp, winObjectIDStamp + strings.Repeat("0", 128), winObjectIDStamp + strings.Repeat("ff", 63), winObjectIDStamp + strings.Repeat("FF", 64), "objectid-v2:" + hex.EncodeToString(raw)} {
		changed := recorded
		changed.Stamp = stamp
		check(changed)
	}
	untouched := winTestPayload(t, parent, "untouched-original", []byte("user bytes"))
	before, err := winInspectMetadata(untouched)
	if err != nil {
		t.Fatal(err)
	}
	birth := winBirthArtifact(untouched.observation)
	for range 2 {
		got, err := nativeArtifactIdentity(untouched)
		if err != nil || got != birth {
			t.Fatal("untouched original profile", err)
		}
		if _, err := winArtifactObjectID(untouched.handle(), windows.FSCTL_GET_OBJECT_ID); err != windows.ERROR_FILE_NOT_FOUND {
			t.Fatal("GET created original ID", err)
		}
		missing := recorded
		missing.Device, missing.File = birth.Device, birth.File
		if err := verifyArtifactIdentity(untouched, missing); err != windows.ERROR_FILE_NOT_FOUND {
			t.Fatal("recorded intrinsic ID fell back", err)
		}
	}
	after, err := winInspectMetadata(untouched)
	content, readErr := nativeRead(context.Background(), untouched)
	if err != nil || readErr != nil || !before.equal(after) || string(content) != "user bytes" {
		t.Fatal("original bytes/security changed", err, readErr)
	}
	// Existing birth profiles remain exact even on an ID-bearing file.
	legacy := winBirthArtifact(owned.observation)
	if err := verifyArtifactIdentity(owned, legacy); err != nil {
		t.Fatal(err)
	}
	legacy.Stamp += "0"
	if err := verifyArtifactIdentity(owned, legacy); !errors.Is(err, errBindingChanged) {
		t.Fatal("drifting legacy birth accepted", err)
	}
	if _, err := nativeArtifactIdentity(parent); !errors.Is(err, errUnsupportedProfile) {
		t.Fatal("directory acquired artifact profile", err)
	}
}

func TestWindowsArtifactLegacyIdentityGoldenShape(t *testing.T) {
	// Literal legacy serializer output, independent of today's marshaled shape.
	raw := []byte(`{"Platform":1,"Device":9,"File":[1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"Stamp":"birth-filetime:133700000000000000"}`)
	var record diskIdentity
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(record)
	if err != nil || !bytes.Equal(raw, encoded) {
		t.Fatal("legacy canonical identity bytes changed", err)
	}
}
