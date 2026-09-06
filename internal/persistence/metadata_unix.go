//go:build linux || darwin || freebsd

package persistence

import (
	"bytes"
	"errors"
	"sort"

	"golang.org/x/sys/unix"
)

const maxNativeMetadata = 1024 * 1024

type unixAttribute struct {
	name  string
	value []byte
}
type unixMetadata struct {
	uid, gid, mode uint32
	attributes     []unixAttribute
}

func unixInspectMetadata(object *unixObject) (unixMetadata, error) {
	var m unixMetadata
	before, err := unixObserve(object.fd())
	if err != nil {
		return m, err
	}
	if uint32(before.stat.Mode)&unix.S_IFMT != unix.S_IFREG {
		return m, errors.New("nonregular metadata profile")
	}
	if err := unixInspectNativePolicy(object.fd(), &before.stat); err != nil {
		return m, err
	}
	m.uid, m.gid, m.mode = before.stat.Uid, before.stat.Gid, uint32(before.stat.Mode)&07777
	m.attributes, err = unixReadAttributes(object.fd())
	if err != nil {
		return m, err
	}
	confirmed, err := unixReadAttributes(object.fd())
	if err != nil {
		return m, err
	}
	if !unixAttributesEqual(m.attributes, confirmed) {
		return m, errors.New("extended metadata changed during observation")
	}
	after, err := unixObserve(object.fd())
	if err != nil {
		return m, err
	}
	if !before.sameRead(after) {
		return m, errors.New("file changed during metadata observation")
	}
	return m, nil
}

func unixAttributesEqual(a, b []unixAttribute) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].name != b[i].name || !bytes.Equal(a[i].value, b[i].value) {
			return false
		}
	}
	return true
}
func (m unixMetadata) equal(n unixMetadata) bool {
	return m.uid == n.uid && m.gid == n.gid && m.mode == n.mode && unixAttributesEqual(m.attributes, n.attributes)
}

func unixReadAttributes(fd int) ([]unixAttribute, error) {
	names, err := unixAttributeNames(fd)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	result := make([]unixAttribute, 0, len(names))
	total := 0
	for i, name := range names {
		if name == "" || i > 0 && name == names[i-1] || !unixSupportedAttribute(name) {
			return nil, errors.New("unsupported or duplicate native extended attribute")
		}
		n, err := unix.Fgetxattr(fd, name, nil)
		if err != nil {
			return nil, err
		}
		if n < 0 || n > maxNativeMetadata-total {
			return nil, errors.New("native metadata size limit")
		}
		value := make([]byte, n)
		got, err := unix.Fgetxattr(fd, name, value)
		if err != nil {
			return nil, err
		}
		if got != n {
			return nil, errors.New("native attribute changed during read")
		}
		total += n + len(name)
		if total > maxNativeMetadata {
			return nil, errors.New("native metadata size limit")
		}
		result = append(result, unixAttribute{name, value})
	}
	return result, nil
}

func unixApplyMetadata(payload *unixObject, m unixMetadata) error {
	// Preserve existing inherited metadata only when it is exactly part of the
	// selected source profile. Never remove an inherited access-policy attribute
	// as a quiet way to make a copy pass.
	current, err := unixReadAttributes(payload.fd())
	if err != nil {
		return err
	}
	for _, item := range current {
		found := false
		for _, source := range m.attributes {
			if source.name == item.name {
				found = true
				break
			}
		}
		if !found {
			return errors.New("payload has additional inherited metadata")
		}
	}
	if err := unix.Fchown(payload.fd(), int(m.uid), int(m.gid)); err != nil {
		return err
	}
	if err := unix.Fchmod(payload.fd(), m.mode); err != nil {
		return err
	}
	for _, item := range m.attributes {
		if err := unix.Fsetxattr(payload.fd(), item.name, item.value, 0); err != nil {
			return err
		}
	}
	got, err := unixInspectMetadata(payload)
	if err != nil {
		return err
	}
	if !m.equal(got) {
		return errors.New("native metadata copy verification failed")
	}
	return nil
}

func unixNullTerminatedNames(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[len(raw)-1] != 0 {
		return nil, errors.New("unterminated native attribute list")
	}
	parts := bytes.Split(raw[:len(raw)-1], []byte{0})
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = string(p)
	}
	return out, nil
}
