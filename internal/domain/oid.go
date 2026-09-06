package domain

import (
	"encoding/hex"
	"errors"
)

// ObjectFormat is the retained object hashing algorithm, distinct from an OID.
type ObjectFormat uint8

const (
	SHA1 ObjectFormat = iota + 1
	SHA256
)

func (f ObjectFormat) Valid() bool { return f == SHA1 || f == SHA256 }

// ByteLength returns zero for an invalid format.
func (f ObjectFormat) ByteLength() int {
	switch f {
	case SHA1:
		return 20
	case SHA256:
		return 32
	default:
		return 0
	}
}

// OID is a nonzero full object identity. Only its format and immutable bytes
// participate in equality. It does not assert object type or existence.
type OID struct {
	format ObjectFormat
	bytes  [32]byte
}

func NewOID(fullHex string) (OID, error) {
	var oid OID
	switch len(fullHex) {
	case 40:
		oid.format = SHA1
	case 64:
		oid.format = SHA256
	default:
		return OID{}, errors.New("object identity requires exactly 40 or 64 hex digits")
	}
	if _, err := hex.Decode(oid.bytes[:oid.format.ByteLength()], []byte(fullHex)); err != nil {
		return OID{}, errors.New("object identity contains nonhex bytes")
	}
	if !oid.Valid() {
		return OID{}, errors.New("zero object identity is invalid")
	}
	return oid, nil
}

func (oid OID) Valid() bool {
	length := oid.format.ByteLength()
	if length == 0 {
		return false
	}
	nonzero := false
	for _, b := range oid.bytes[:length] {
		nonzero = nonzero || b != 0
	}
	for _, b := range oid.bytes[length:] {
		if b != 0 {
			return false
		}
	}
	return nonzero
}
func (oid OID) Format() ObjectFormat { return oid.format }
func (oid OID) Equal(other OID) bool { return oid == other }

// String returns full canonical lowercase hex, or empty for an invalid value.
func (oid OID) String() string {
	if !oid.Valid() {
		return ""
	}
	return hex.EncodeToString(oid.bytes[:oid.format.ByteLength()])
}

// Bytes returns an owned copy; invalid values return nil.
func (oid OID) Bytes() []byte {
	if !oid.Valid() {
		return nil
	}
	return append([]byte(nil), oid.bytes[:oid.format.ByteLength()]...)
}

// Revision is an exact commit identity in one repository scope. The observing
// adapter must verify commit type and existence in that scope's object database.
// Mutable names, protocol null OIDs and latest-ref queries are not Revisions.
type Revision struct {
	repository RepositoryID
	oid        OID
}

func NewRevision(repository RepositoryID, oid OID) (Revision, error) {
	if !repository.Valid() || !oid.Valid() {
		return Revision{}, errors.New("revision requires a valid repository and exact nonzero object identity")
	}
	return Revision{repository: repository, oid: oid}, nil
}

func (r Revision) Valid() bool               { return r.repository.Valid() && r.oid.Valid() }
func (r Revision) Repository() RepositoryID  { return r.repository }
func (r Revision) OID() OID                  { return r.oid }
func (r Revision) Equal(other Revision) bool { return r == other }

// StashID names an exact stash object independently of its current reflog
// position, label, source worktree or other observations.
type StashID struct {
	repository RepositoryID
	oid        OID
}

func NewStashID(repository RepositoryID, oid OID) (StashID, error) {
	if !repository.Valid() || repository.Scope() != LocalCommon || !oid.Valid() {
		return StashID{}, errors.New("stash identity requires a local common repository and exact nonzero object identity")
	}
	return StashID{repository: repository, oid: oid}, nil
}

func (id StashID) Valid() bool {
	return id.repository.Valid() && id.repository.Scope() == LocalCommon && id.oid.Valid()
}
func (id StashID) Repository() RepositoryID { return id.repository }
func (id StashID) OID() OID                 { return id.oid }
func (id StashID) Equal(other StashID) bool { return id == other }
