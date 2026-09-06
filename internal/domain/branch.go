package domain

import (
	"errors"
	"strings"
)

// BranchKind distinguishes a local branch from a head in a remote repository.
// Cached remote-tracking refs are adapter observations, not another BranchKind.
type BranchKind uint8

const (
	Local BranchKind = iota + 1
	RemoteHead
)

func (k BranchKind) Valid() bool { return k == Local || k == RemoteHead }

// BranchID retains the entire exact branch name beneath refs/heads/. The name
// is neither a display label nor a native ref expression. See README.md for the
// pure stored-name grammar and adapter-owned creation restrictions.
type BranchID struct {
	repository RepositoryID
	kind       BranchKind
	name       string
}

func NewBranchID(repository RepositoryID, kind BranchKind, name string) (BranchID, error) {
	if !branchScopeValid(repository, kind) {
		return BranchID{}, errors.New("branch kind must match a valid repository scope")
	}
	if !validBranchName(name) {
		return BranchID{}, errors.New("branch identity requires an exact valid stored branch name")
	}
	return BranchID{repository: repository, kind: kind, name: name}, nil
}

func branchScopeValid(repository RepositoryID, kind BranchKind) bool {
	return repository.Valid() && (kind == Local && repository.Scope() == LocalCommon ||
		kind == RemoteHead && repository.Scope() == Remote)
}

// Validate the exact suffix of refs/heads/, without normalizing or expanding it.
// Git's ASCII byte exclusions deliberately do not impose Unicode normalization
// or a UTF-8 restriction on names observed by an adapter.
func validBranchName(name string) bool {
	if name == "" || strings.HasSuffix(name, ".") || strings.Contains(name, "..") || strings.Contains(name, "@{") {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c <= ' ' || c == 0x7f || strings.ContainsRune("~^:?*[\\", rune(c)) {
			return false
		}
	}
	for _, component := range strings.Split(name, "/") {
		if component == "" || component[0] == '.' || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}

func (id BranchID) Valid() bool {
	return branchScopeValid(id.repository, id.kind) && validBranchName(id.name)
}
func (id BranchID) Repository() RepositoryID  { return id.repository }
func (id BranchID) Kind() BranchKind          { return id.kind }
func (id BranchID) Name() string              { return id.name }
func (id BranchID) Equal(other BranchID) bool { return id == other }
