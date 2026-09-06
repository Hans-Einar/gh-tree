//go:build linux || darwin || freebsd

package persistence

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestUnixNativeMetadataCopy(t *testing.T) {
	parent := acquiredUnix(t, physicalTemp(t)).parent()
	source := unixTestPayload(t, parent, "source", []byte("old"))
	if err := unix.Fchmod(source.fd(), 0640); err != nil {
		t.Fatal(err)
	}
	if err := unix.Fsetxattr(source.fd(), "user.gh-tree-fixture", []byte("exact metadata"), 0); err != nil {
		t.Fatal(err)
	}
	metadata, err := unixInspectMetadata(source)
	if err != nil {
		t.Fatal(err)
	}
	payload := unixTestPayload(t, parent, "payload", []byte("new"))
	if err := unixApplyMetadata(payload, metadata); err != nil {
		t.Fatal(err)
	}
	actual, err := unixInspectMetadata(payload)
	if err != nil || !metadata.equal(actual) {
		t.Fatal("metadata changed", err)
	}
	if err := payload.file.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := unix.Fsetxattr(payload.fd(), "user.additional", []byte("preserve"), 0); err != nil {
		t.Fatal(err)
	}
	if err := unixApplyMetadata(payload, metadata); err == nil {
		t.Fatal("additional inherited attribute silently removed")
	}
	if err := unix.Fchmod(source.fd(), 0740); err != nil {
		t.Fatal(err)
	}
	executable, err := unixInspectMetadata(source)
	if err != nil {
		t.Fatal(err)
	}
	copy := unixTestPayload(t, parent, "copy-mode", []byte("new"))
	if err := unixApplyMetadata(copy, executable); err != nil {
		t.Fatal(err)
	}
}

func TestUnixNativeMetadataInvalidDescriptor(t *testing.T) {
	if _, err := unixAttributeNames(-1); err == nil {
		t.Fatal("native query error became attribute absence")
	}
	if _, err := unixReadAttributes(-1); err == nil {
		t.Fatal("metadata observation erased native query error")
	}
}

func TestUnixNativeAttributeListValidation(t *testing.T) {
	for _, bad := range [][]byte{[]byte("unterminated"), []byte("a\x00b")} {
		if _, err := unixNullTerminatedNames(bad); err == nil {
			t.Fatal("truncated list accepted")
		}
	}
	got, err := unixNullTerminatedNames([]byte("user.a\x00user.b\x00"))
	if err != nil || len(got) != 2 || got[0] != "user.a" || got[1] != "user.b" {
		t.Fatal(got, err)
	}
	if unixAttributesEqual([]unixAttribute{{"x", []byte("1")}}, []unixAttribute{{"x", []byte("2")}}) {
		t.Fatal("changed attribute equals old")
	}
}
