package viewmodel

import "github.com/Hans-Einar/gh-tree/internal/domain"

type ElementKind uint8

const (
	NamespaceElement ElementKind = iota + 1
	RepositoryElement
	PRElement
	BranchElement
	RevisionElement
	WorktreeElement
	StashElement
	LaunchElement
	SessionElement
)

func (k ElementKind) Valid() bool { return k >= NamespaceElement && k <= SessionElement }

// ElementID is closed and comparable. Labels, list position and source versions
// do not participate. A saved alias remains exact and case sensitive.
type ElementID struct {
	kind           ElementKind
	namespace, key string
	repository     Optional[domain.RepositoryID]
	pr             domain.PRID
	branch         domain.BranchID
	revision       domain.Revision
	worktree       domain.WorktreeID
	stash          domain.StashID
	launch         domain.LaunchPointID
	alias          Optional[string]
	session        domain.SessionID
}

func NewNamespaceElement(namespace, key string, repository Optional[domain.RepositoryID]) (ElementID, error) {
	return checkedElement(ElementID{kind: NamespaceElement, namespace: namespace, key: key, repository: repository})
}
func NewRepositoryElement(id domain.RepositoryID) (ElementID, error) {
	return checkedElement(ElementID{kind: RepositoryElement, repository: Some(id)})
}
func NewPRElement(id domain.PRID) (ElementID, error) {
	return checkedElement(ElementID{kind: PRElement, pr: id})
}
func NewBranchElement(id domain.BranchID) (ElementID, error) {
	return checkedElement(ElementID{kind: BranchElement, branch: id})
}
func NewRevisionElement(id domain.Revision) (ElementID, error) {
	return checkedElement(ElementID{kind: RevisionElement, revision: id})
}
func NewWorktreeElement(id domain.WorktreeID) (ElementID, error) {
	return checkedElement(ElementID{kind: WorktreeElement, worktree: id})
}
func NewStashElement(id domain.StashID) (ElementID, error) {
	return checkedElement(ElementID{kind: StashElement, stash: id})
}
func NewLaunchElement(id domain.LaunchPointID, alias Optional[string]) (ElementID, error) {
	return checkedElement(ElementID{kind: LaunchElement, launch: id, alias: alias})
}
func NewSessionElement(id domain.SessionID) (ElementID, error) {
	return checkedElement(ElementID{kind: SessionElement, session: id})
}
func checkedElement(id ElementID) (ElementID, error) {
	if !id.Valid() {
		return ElementID{}, invalid("element identity")
	}
	return id, nil
}
func (id ElementID) Valid() bool {
	if !id.kind.Valid() {
		return false
	}
	rest := id
	rest.kind = 0
	switch id.kind {
	case NamespaceElement:
		if id.namespace == "" || id.key == "" || !optionalValid(id.repository) {
			return false
		}
		rest.namespace = ""
		rest.key = ""
		rest.repository = None[domain.RepositoryID]()
	case RepositoryElement:
		if !id.repository.present || !id.repository.value.Valid() {
			return false
		}
		rest.repository = None[domain.RepositoryID]()
	case PRElement:
		if !id.pr.Valid() {
			return false
		}
		rest.pr = domain.PRID{}
	case BranchElement:
		if !id.branch.Valid() {
			return false
		}
		rest.branch = domain.BranchID{}
	case RevisionElement:
		if !id.revision.Valid() {
			return false
		}
		rest.revision = domain.Revision{}
	case WorktreeElement:
		if !id.worktree.Valid() {
			return false
		}
		rest.worktree = domain.WorktreeID{}
	case StashElement:
		if !id.stash.Valid() {
			return false
		}
		rest.stash = domain.StashID{}
	case LaunchElement:
		if !id.launch.Valid() || id.alias.present && id.alias.value == "" {
			return false
		}
		rest.launch = domain.LaunchPointID{}
		rest.alias = None[string]()
	case SessionElement:
		if !id.session.Valid() {
			return false
		}
		rest.session = domain.SessionID{}
	}
	return rest == (ElementID{})
}
func (id ElementID) Kind() ElementKind { return id.kind }
func (id ElementID) Namespace() (string, string, Optional[domain.RepositoryID], bool) {
	if id.Valid() && id.kind == NamespaceElement {
		return id.namespace, id.key, id.repository, true
	}
	return "", "", None[domain.RepositoryID](), false
}
func (id ElementID) Repository() (domain.RepositoryID, bool) {
	if id.Valid() && id.kind == RepositoryElement {
		return id.repository.value, true
	}
	return domain.RepositoryID{}, false
}
func (id ElementID) PullRequest() (domain.PRID, bool) {
	return id.pr, id.Valid() && id.kind == PRElement
}
func (id ElementID) Branch() (domain.BranchID, bool) {
	return id.branch, id.Valid() && id.kind == BranchElement
}
func (id ElementID) Revision() (domain.Revision, bool) {
	return id.revision, id.Valid() && id.kind == RevisionElement
}
func (id ElementID) Worktree() (domain.WorktreeID, bool) {
	return id.worktree, id.Valid() && id.kind == WorktreeElement
}
func (id ElementID) Stash() (domain.StashID, bool) {
	return id.stash, id.Valid() && id.kind == StashElement
}
func (id ElementID) Launch() (domain.LaunchPointID, Optional[string], bool) {
	return id.launch, id.alias, id.Valid() && id.kind == LaunchElement
}
func (id ElementID) Session() (domain.SessionID, bool) {
	return id.session, id.Valid() && id.kind == SessionElement
}

// SubjectRepository is identity inspection only, never a cross-source join.
func (id ElementID) SubjectRepository() Optional[domain.RepositoryID] {
	if !id.Valid() {
		return None[domain.RepositoryID]()
	}
	switch id.kind {
	case NamespaceElement, RepositoryElement:
		return id.repository
	case PRElement:
		return Some(id.pr.Repository())
	case BranchElement:
		return Some(id.branch.Repository())
	case RevisionElement:
		return Some(id.revision.Repository())
	case WorktreeElement:
		return Some(id.worktree.Repository())
	case StashElement:
		return Some(id.stash.Repository())
	case LaunchElement:
		return Some(id.launch.Worktree().Repository())
	}
	return None[domain.RepositoryID]()
}
