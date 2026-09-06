// Package domain defines immutable, comparable identities and exact targets.
// Constructors perform deterministic validation only. A zero value is invalid;
// callers must use Valid before accepting values from a boundary.
package domain

import "errors"

// RepositoryScope distinguishes remote repositories from local object stores.
type RepositoryScope uint8

const (
	Remote RepositoryScope = iota + 1
	LocalCommon
)

func (s RepositoryScope) Valid() bool { return s == Remote || s == LocalCommon }

// RepositoryID is an adapter-minted opaque token in a particular scope.
// Domain neither canonicalizes the token nor establishes repository existence.
type RepositoryID struct {
	scope RepositoryScope
	token string
}

func NewRepositoryID(scope RepositoryScope, token string) (RepositoryID, error) {
	if !scope.Valid() || token == "" {
		return RepositoryID{}, errors.New("repository identity requires a valid scope and nonempty token")
	}
	return RepositoryID{scope: scope, token: token}, nil
}

func (id RepositoryID) Valid() bool                   { return id.scope.Valid() && id.token != "" }
func (id RepositoryID) Scope() RepositoryScope        { return id.scope }
func (id RepositoryID) Token() string                 { return id.token }
func (id RepositoryID) Equal(other RepositoryID) bool { return id == other }

// WorktreeID identifies a registered worktree by its local common repository
// and Git-issued administrative key. Retargeting does not change this value.
type WorktreeID struct {
	repository RepositoryID
	key        string
}

func NewWorktreeID(repository RepositoryID, administrativeKey string) (WorktreeID, error) {
	if !repository.Valid() || repository.Scope() != LocalCommon || administrativeKey == "" {
		return WorktreeID{}, errors.New("worktree identity requires a local common repository and nonempty administrative key")
	}
	return WorktreeID{repository: repository, key: administrativeKey}, nil
}

func (id WorktreeID) Valid() bool {
	return id.repository.Valid() && id.repository.Scope() == LocalCommon && id.key != ""
}
func (id WorktreeID) Repository() RepositoryID    { return id.repository }
func (id WorktreeID) AdministrativeKey() string   { return id.key }
func (id WorktreeID) Equal(other WorktreeID) bool { return id == other }

// PRID identifies a pull request in its base remote repository. Its head may
// belong to a different remote repository, including a fork.
type PRID struct {
	repository RepositoryID
	number     uint64
}

func NewPRID(repository RepositoryID, number uint64) (PRID, error) {
	if !repository.Valid() || repository.Scope() != Remote || number == 0 {
		return PRID{}, errors.New("pull request identity requires a remote repository and positive number")
	}
	return PRID{repository: repository, number: number}, nil
}

func (id PRID) Valid() bool {
	return id.repository.Valid() && id.repository.Scope() == Remote && id.number != 0
}
func (id PRID) Repository() RepositoryID { return id.repository }
func (id PRID) Number() uint64           { return id.number }
func (id PRID) Equal(other PRID) bool    { return id == other }
