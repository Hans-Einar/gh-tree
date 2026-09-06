// Package api contains the immutable semantic vocabulary of the frozen v0.4
// boundaries. Constructors validate values; they do not observe native state,
// allocate operation/session identities, or grant execution authority.
package api

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func invalid(field string) error { return errors.New("invalid boundary value: " + field) }

// Optional expresses absence. Some does not validate or clone arbitrary T;
// containing family constructors validate and copy every mutable instantiation.
type Optional[T any] struct {
	present bool
	value   T
}

func None[T any]() Optional[T]         { return Optional[T]{} }
func Some[T any](value T) Optional[T]  { return Optional[T]{true, value} }
func (v Optional[T]) Present() bool    { return v.present }
func (v Optional[T]) Value() (T, bool) { return v.value, v.present }
func cloneSlice[T any](v []T) []T {
	if v == nil {
		return nil
	}
	r := make([]T, len(v))
	copy(r, v)
	return r
}
func cloneOptionalSlice[T any](v Optional[[]T]) Optional[[]T] {
	if x, ok := v.Value(); ok {
		return Some(cloneSlice(x))
	}
	return None[[]T]()
}

type FieldPresence uint8

const (
	FieldAbsent FieldPresence = iota
	FieldNull
	FieldPresent
)

type StoredField[T any] struct {
	presence FieldPresence
	value    T
}

func AbsentField[T any]() StoredField[T]         { return StoredField[T]{} }
func NullField[T any]() StoredField[T]           { return StoredField[T]{presence: FieldNull} }
func PresentField[T any](v T) StoredField[T]     { return StoredField[T]{FieldPresent, v} }
func (v StoredField[T]) Presence() FieldPresence { return v.presence }
func (v StoredField[T]) Value() (T, bool)        { return v.value, v.presence == FieldPresent }
func (v StoredField[T]) Valid() bool             { return v.presence <= FieldPresent }
func cloneStoredSlice[T any](v StoredField[[]T]) StoredField[[]T] {
	if x, ok := v.Value(); ok {
		return PresentField(cloneSlice(x))
	}
	return v
}

func literal(v string) bool   { return !strings.ContainsRune(v, 0) }
func textValue(v string) bool { return utf8.ValidString(v) && literal(v) }
func nonempty(v string) bool  { return v != "" && literal(v) }
func component(v string) bool {
	return nonempty(v) && v != "." && v != ".." && !strings.ContainsAny(v, "/\\") && !(len(v) >= 2 && v[1] == ':')
}
func components(v []string) bool {
	for _, s := range v {
		if !component(s) {
			return false
		}
	}
	return true
}

// GitPath preserves literal relative path bytes, including whitespace/newlines.
type GitPath struct{ value string }

func NewGitPath(v string) (GitPath, error) {
	if !nonempty(v) || strings.HasPrefix(v, "/") || strings.HasPrefix(v, "\\") || (len(v) > 2 && v[1] == ':' && (v[2] == '/' || v[2] == '\\')) {
		return GitPath{}, invalid("GitPath")
	}
	for _, p := range strings.FieldsFunc(v, func(r rune) bool { return r == '/' || r == '\\' }) {
		if p == "." || p == ".." {
			return GitPath{}, invalid("GitPath traversal")
		}
	}
	if strings.Contains(v, "//") || strings.Contains(v, "\\\\") || strings.HasSuffix(v, "/") || strings.HasSuffix(v, "\\") {
		return GitPath{}, invalid("GitPath component")
	}
	return GitPath{v}, nil
}
func (v GitPath) Valid() bool    { _, e := NewGitPath(v.value); return e == nil }
func (v GitPath) String() string { return v.value }

// SourceVersion is comparable evidence, never a native capability. Its fields
// are supplied by the responsible observer without allocation or I/O here.
type SourceVersion struct{ namespace, scope, issuer, token string }

func NewSourceVersion(namespace, scope, issuer, token string) (SourceVersion, error) {
	v := SourceVersion{namespace, scope, issuer, token}
	if !v.Valid() {
		return SourceVersion{}, invalid("source version")
	}
	return v, nil
}
func (v SourceVersion) Valid() bool {
	return nonempty(v.namespace) && nonempty(v.scope) && nonempty(v.issuer) && nonempty(v.token)
}
func (v SourceVersion) Namespace() string          { return v.namespace }
func (v SourceVersion) Scope() string              { return v.scope }
func (v SourceVersion) Issuer() string             { return v.issuer }
func (v SourceVersion) Equal(w SourceVersion) bool { return v == w }

type StorageVersion struct {
	family   StorageFamily
	store    string
	present  bool
	length   uint64
	digest   [32]byte
	worktree domain.WorktreeID
	root     DirectoryIdentity
}

func NewStorageVersion(family StorageFamily, store string, present bool, length uint64, digest [32]byte) (StorageVersion, error) {
	v := StorageVersion{family: family, store: store, present: present, length: length, digest: digest}
	if !v.Valid() {
		return StorageVersion{}, invalid("storage version")
	}
	return v, nil
}
func (v StorageVersion) Valid() bool {
	base := v.family.Valid() && nonempty(v.store) && (v.present || (v.length == 0 && v.digest == [32]byte{}))
	if v.family == RunConfig {
		return base && v.worktree.Valid() && v.root.Valid()
	}
	return base && v.worktree == (domain.WorktreeID{}) && v.root == (DirectoryIdentity{})
}

// NewRunStorageVersion additionally binds the selected worktree and root object.
// Store records the observer's actual parent/absence-anchor identity, not a path
// to resolve here. Native binding and missing ancestry are revalidated by Storage.
func NewRunStorageVersion(scope WorktreeScope, store string, present bool, length uint64, digest [32]byte) (StorageVersion, error) {
	if !scope.Valid() {
		return StorageVersion{}, invalid("run scope")
	}
	v := StorageVersion{family: RunConfig, store: store, present: present, length: length, digest: digest, worktree: scope.data.ID, root: scope.data.RootIdentity}
	if !v.Valid() {
		return StorageVersion{}, invalid("run storage version")
	}
	return v, nil
}
func (v StorageVersion) Worktree() Optional[domain.WorktreeID] {
	if v.family == RunConfig && v.Valid() {
		return Some(v.worktree)
	}
	return None[domain.WorktreeID]()
}
func (v StorageVersion) MatchesRunScope(scope WorktreeScope) bool {
	return v.Valid() && scope.Valid() && v.family == RunConfig && v.worktree == scope.data.ID && v.root == scope.data.RootIdentity
}
func (v StorageVersion) Family() StorageFamily       { return v.family }
func (v StorageVersion) Store() string               { return v.store }
func (v StorageVersion) Present() bool               { return v.present }
func (v StorageVersion) ByteLength() uint64          { return v.length }
func (v StorageVersion) SHA256() [32]byte            { return v.digest }
func (v StorageVersion) Equal(w StorageVersion) bool { return v == w }

type DirectoryIdentity struct {
	platform DirectoryPlatform
	device   uint64
	file     [16]byte
	stamp    string
}

func NewDirectoryIdentity(platform DirectoryPlatform, device uint64, file [16]byte, stamp string) (DirectoryIdentity, error) {
	v := DirectoryIdentity{platform, device, file, stamp}
	if !v.Valid() {
		return DirectoryIdentity{}, invalid("directory identity")
	}
	return v, nil
}
func (v DirectoryIdentity) Valid() bool {
	return v.platform.Valid() && v.file != [16]byte{} && nonempty(v.stamp)
}
func (v DirectoryIdentity) Platform() DirectoryPlatform    { return v.platform }
func (v DirectoryIdentity) Device() uint64                 { return v.device }
func (v DirectoryIdentity) FileID() [16]byte               { return v.file }
func (v DirectoryIdentity) Stamp() string                  { return v.stamp }
func (v DirectoryIdentity) Equal(w DirectoryIdentity) bool { return v == w }
