package viewmodel

import "github.com/Hans-Einar/gh-tree/internal/domain"

type PRState uint8

const (
	PROpen PRState = iota + 1
	PRClosed
	PRMerged
	PRStateUnknown
)

func (v PRState) Valid() bool { return v >= PROpen && v <= PRStateUnknown }

type ChangeKind uint8

const (
	ModifiedChange ChangeKind = iota + 1
	AddedChange
	DeletedChange
	RenamedChange
	CopiedChange
	TypeChanged
	UntrackedChange
	ConflictedChange
	ChangeUnknown
)

func (v ChangeKind) Valid() bool { return v >= ModifiedChange && v <= ChangeUnknown }

type ChangeStatus uint8

const (
	StatusClean ChangeStatus = iota + 1
	StatusAdded
	StatusModified
	StatusDeleted
	StatusRenamed
	StatusCopied
	StatusUnmerged
	StatusUntracked
	StatusTypeChanged
	StatusUnknown
)

func (v ChangeStatus) Valid() bool { return v >= StatusClean && v <= StatusUnknown }

type RefKind uint8

const (
	BranchRef RefKind = iota + 1
	TagRef
	HeadRef
	RemoteTrackingRef
	OtherRef
)

func (v RefKind) Valid() bool { return v >= BranchRef && v <= OtherRef }

type StashView uint8

const (
	StashWorktreeView StashView = iota + 1
	StashIndexView
	StashUntrackedView
	StashIndexToWorktreeView
	StashParentView
)

func (v StashView) Valid() bool { return v >= StashWorktreeView && v <= StashParentView }

type ComparisonKind uint8

const (
	CommitParentComparison ComparisonKind = iota + 1
	CommitPairComparison
	IndexToWorktreeComparison
	HeadToIndexComparison
	PullRequestComparison
	StashComparison
)

func (v ComparisonKind) Valid() bool { return v >= CommitParentComparison && v <= StashComparison }

// Comparison preserves exact original and resolved endpoints. No constructor
// resolves refs, associates repositories or invents a merge base.
type Comparison struct {
	kind                                  ComparisonKind
	from, to                              Optional[domain.Revision]
	worktree                              domain.WorktreeID
	head                                  domain.Head
	target                                domain.ExactTarget
	base, localBase, localHead, mergeBase domain.Revision
	stashDetail                           StashComparisonDetail
}

func NewCommitParentComparison(commit domain.Revision, parent Optional[domain.Revision]) (Comparison, error) {
	return checkedComparison(Comparison{kind: CommitParentComparison, from: parent, to: Some(commit)})
}
func NewCommitPairComparison(from, to domain.Revision) (Comparison, error) {
	return checkedComparison(Comparison{kind: CommitPairComparison, from: Some(from), to: Some(to)})
}
func NewIndexToWorktreeComparison(worktree domain.WorktreeID) (Comparison, error) {
	return checkedComparison(Comparison{kind: IndexToWorktreeComparison, worktree: worktree})
}
func NewHeadToIndexComparison(worktree domain.WorktreeID, head domain.Head) (Comparison, error) {
	return checkedComparison(Comparison{kind: HeadToIndexComparison, worktree: worktree, head: head})
}
func NewPullRequestComparison(target domain.ExactTarget, base, localBase, localHead, mergeBase domain.Revision) (Comparison, error) {
	return checkedComparison(Comparison{kind: PullRequestComparison, target: target, base: base, localBase: localBase, localHead: localHead, mergeBase: mergeBase})
}
func NewStashComparison(detail StashComparisonDetail) (Comparison, error) {
	return checkedComparison(Comparison{kind: StashComparison, stashDetail: detail})
}
func checkedComparison(c Comparison) (Comparison, error) {
	if !c.Valid() {
		return Comparison{}, invalid("comparison")
	}
	return c, nil
}
func (c Comparison) Valid() bool {
	rest := c
	rest.kind = 0
	switch c.kind {
	case CommitParentComparison, CommitPairComparison:
		if !c.to.present || !c.to.value.Valid() || !optionalValid(c.from) || c.kind == CommitPairComparison && !c.from.present || c.from.present && c.from.value.Repository() != c.to.value.Repository() {
			return false
		}
		rest.from = None[domain.Revision]()
		rest.to = None[domain.Revision]()
	case IndexToWorktreeComparison:
		if !c.worktree.Valid() {
			return false
		}
		rest.worktree = domain.WorktreeID{}
	case HeadToIndexComparison:
		if !c.head.MatchesWorktree(c.worktree) {
			return false
		}
		rest.worktree = domain.WorktreeID{}
		rest.head = domain.Head{}
	case PullRequestComparison:
		if !c.target.Valid() || c.target.Kind() != domain.PullRequestTarget || !c.base.Valid() || c.base.Repository().Scope() != domain.Remote || !c.localBase.Valid() || c.localBase.Repository().Scope() != domain.LocalCommon || !c.localHead.Valid() || !c.mergeBase.Valid() || c.localBase.Repository() != c.localHead.Repository() || c.localBase.Repository() != c.mergeBase.Repository() || c.localBase.OID() != c.base.OID() || c.localHead.OID() != c.target.ExpectedRevision().OID() {
			return false
		}
		pr, _ := c.target.PullRequest()
		if c.base.Repository() != pr.Repository() {
			return false
		}
		rest.target = domain.ExactTarget{}
		rest.base = domain.Revision{}
		rest.localBase = domain.Revision{}
		rest.localHead = domain.Revision{}
		rest.mergeBase = domain.Revision{}
	case StashComparison:
		if !c.stashDetail.Valid() {
			return false
		}
		rest.stashDetail = StashComparisonDetail{}
	default:
		return false
	}
	return rest == (Comparison{})
}
func (c Comparison) Kind() ComparisonKind { return c.kind }
func (c Comparison) CommitEndpoints() (Optional[domain.Revision], domain.Revision, bool) {
	if c.Valid() && (c.kind == CommitParentComparison || c.kind == CommitPairComparison) {
		return c.from, c.to.value, true
	}
	return None[domain.Revision](), domain.Revision{}, false
}
func (c Comparison) Worktree() (domain.WorktreeID, Optional[domain.Head], bool) {
	if c.Valid() && (c.kind == IndexToWorktreeComparison || c.kind == HeadToIndexComparison) {
		h := None[domain.Head]()
		if c.kind == HeadToIndexComparison {
			h = Some(c.head)
		}
		return c.worktree, h, true
	}
	return domain.WorktreeID{}, None[domain.Head](), false
}
func (c Comparison) PullRequest() (domain.ExactTarget, domain.Revision, domain.Revision, domain.Revision, domain.Revision, bool) {
	if c.Valid() && c.kind == PullRequestComparison {
		return c.target, c.base, c.localBase, c.localHead, c.mergeBase, true
	}
	return domain.ExactTarget{}, domain.Revision{}, domain.Revision{}, domain.Revision{}, domain.Revision{}, false
}
func (c Comparison) Stash() (StashComparisonDetail, bool) {
	if c.Valid() && c.kind == StashComparison {
		return c.stashDetail, true
	}
	return StashComparisonDetail{}, false
}

func revisionsInScope(xs []domain.Revision, repo domain.RepositoryID) bool {
	for _, r := range xs {
		if !r.Valid() || r.Repository() != repo {
			return false
		}
	}
	return true
}
func remoteEndpoints(xs []BranchEndpoint) bool {
	for _, x := range xs {
		if x.data.Branch.Kind() != domain.RemoteHead {
			return false
		}
	}
	return true
}
func launchMembersInScope(xs []LaunchMember, wt domain.WorktreeID) bool {
	ids := map[domain.LaunchPointID]bool{}
	order := map[uint32]bool{}
	for _, x := range xs {
		if x.data.ID.Worktree() != wt || ids[x.data.ID] {
			return false
		}
		ids[x.data.ID] = true
		if x.data.SelectedOrder.present {
			n := x.data.SelectedOrder.value
			if order[n] {
				return false
			}
			order[n] = true
		}
	}
	for n := uint32(1); n <= uint32(len(order)); n++ {
		if !order[n] {
			return false
		}
	}
	return true
}
func selectedKind(s ListState, k ElementKind) bool {
	return !s.data.Selected.present || s.data.Selected.value.kind == k
}
func graphScope(s GraphModelSpec) bool {
	seen := map[domain.Revision]bool{}
	for _, c := range s.Commits {
		if c.data.Revision.Repository() != s.Repository || seen[c.data.Revision] {
			return false
		}
		seen[c.data.Revision] = true
	}
	for _, r := range s.Refs {
		if r.data.Revision.Repository() != s.Repository {
			return false
		}
	}
	for _, a := range s.Annotations {
		if a.data.Revision.Repository() != s.Repository {
			return false
		}
	}
	return !s.List.data.Selected.present || s.List.data.Selected.value.revision.Repository() == s.Repository
}
