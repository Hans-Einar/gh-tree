package domain

import "errors"

type HeadKind uint8

const (
	Attached HeadKind = iota + 1
	Detached
	Unborn
)

func (k HeadKind) Valid() bool { return k == Attached || k == Detached || k == Unborn }

// Head is a closed private tagged value: Attached(branch, revision),
// Detached(revision), or Unborn(branch). No unknown or nil alternative exists.
// Missing observations belong to the consuming API's explicit optional value.
type Head struct {
	kind     HeadKind
	branch   BranchID
	revision Revision
}

func NewAttachedHead(branch BranchID, revision Revision) (Head, error) {
	h := Head{kind: Attached, branch: branch, revision: revision}
	if !h.Valid() {
		return Head{}, errors.New("attached head requires a local branch and exact revision in the same repository")
	}
	return h, nil
}

func NewDetachedHead(revision Revision) (Head, error) {
	h := Head{kind: Detached, revision: revision}
	if !h.Valid() {
		return Head{}, errors.New("detached head requires an exact local revision")
	}
	return h, nil
}

func NewUnbornHead(branch BranchID) (Head, error) {
	h := Head{kind: Unborn, branch: branch}
	if !h.Valid() {
		return Head{}, errors.New("unborn head requires a local branch without a revision")
	}
	return h, nil
}

func (h Head) Valid() bool {
	switch h.kind {
	case Attached:
		return h.branch.Valid() && h.branch.Kind() == Local && h.revision.Valid() && h.branch.Repository() == h.revision.Repository()
	case Detached:
		return h.branch == (BranchID{}) && h.revision.Valid() && h.revision.Repository().Scope() == LocalCommon
	case Unborn:
		return h.branch.Valid() && h.branch.Kind() == Local && h.revision == (Revision{})
	default:
		return false
	}
}

func (h Head) Kind() HeadKind        { return h.kind }
func (h Head) Equal(other Head) bool { return h == other }

// Branch returns absence for detached or invalid Head.
func (h Head) Branch() (BranchID, bool) {
	if h.Valid() && (h.kind == Attached || h.kind == Unborn) {
		return h.branch, true
	}
	return BranchID{}, false
}

// Revision returns absence for unborn or invalid Head; no null revision is minted.
func (h Head) Revision() (Revision, bool) {
	if h.Valid() && (h.kind == Attached || h.kind == Detached) {
		return h.revision, true
	}
	return Revision{}, false
}

func (h Head) Repository() RepositoryID {
	if !h.Valid() {
		return RepositoryID{}
	}
	if h.kind == Detached {
		return h.revision.Repository()
	}
	return h.branch.Repository()
}

// MatchesWorktree checks local repository scope only. The adapter must still
// establish that this Head was actually observed for the specified worktree.
func (h Head) MatchesWorktree(worktree WorktreeID) bool {
	return h.Valid() && worktree.Valid() && h.Repository() == worktree.Repository()
}
