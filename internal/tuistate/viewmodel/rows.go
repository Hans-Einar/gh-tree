package viewmodel

import "github.com/Hans-Einar/gh-tree/internal/domain"

// StatusNoticeSpec is a constructor/accessor copy; changes cannot mutate StatusNotice.
type StatusNoticeSpec struct {
	Text     string
	Severity Severity
	Subject  Optional[ElementID]
}
type StatusNotice struct {
	data  StatusNoticeSpec
	valid bool
}

func NewStatusNotice(s StatusNoticeSpec) (StatusNotice, error) {
	if !validStatusNotice(s) {
		return StatusNotice{}, invalid("StatusNotice")
	}
	return StatusNotice{cloneStatusNotice(s), true}, nil
}
func validStatusNotice(s StatusNoticeSpec) bool {
	return s.Severity.Valid() && optionalValid(s.Subject)
}
func cloneStatusNotice(s StatusNoticeSpec) StatusNoticeSpec { ; return s }
func (v StatusNotice) Valid() bool                          { return v.valid && validStatusNotice(v.data) }
func (v StatusNotice) Fields() StatusNoticeSpec             { return cloneStatusNotice(v.data) }
func (v StatusNotice) Clone() StatusNotice                  { return StatusNotice{cloneStatusNotice(v.data), v.valid} }

// BadgeSpec is a constructor/accessor copy; changes cannot mutate Badge.
type BadgeSpec struct {
	Label    string
	Severity Severity
	Relation Optional[Relation]
}
type Badge struct {
	data  BadgeSpec
	valid bool
}

func NewBadge(s BadgeSpec) (Badge, error) {
	if !validBadge(s) {
		return Badge{}, invalid("Badge")
	}
	return Badge{cloneBadge(s), true}, nil
}
func validBadge(s BadgeSpec) bool {
	return s.Label != "" && s.Severity.Valid() && (!s.Relation.present || s.Relation.value.Valid())
}
func cloneBadge(s BadgeSpec) BadgeSpec { ; return s }
func (v Badge) Valid() bool            { return v.valid && validBadge(v.data) }
func (v Badge) Fields() BadgeSpec      { return cloneBadge(v.data) }
func (v Badge) Clone() Badge           { return Badge{cloneBadge(v.data), v.valid} }

// SourceStatusSpec is a constructor/accessor copy; changes cannot mutate SourceStatus.
type SourceStatusSpec struct {
	Label        string
	Availability Availability
	Completeness Completeness
	Freshness    Freshness
	Generation   Optional[SourceGeneration]
	Notices      []StatusNotice
}
type SourceStatus struct {
	data  SourceStatusSpec
	valid bool
}

func NewSourceStatus(s SourceStatusSpec) (SourceStatus, error) {
	if !validSourceStatus(s) {
		return SourceStatus{}, invalid("SourceStatus")
	}
	return SourceStatus{cloneSourceStatus(s), true}, nil
}
func validSourceStatus(s SourceStatusSpec) bool {
	return s.Label != "" && s.Availability.Valid() && s.Completeness.Valid() && s.Freshness.Valid() && optionalValid(s.Generation) && allValid(s.Notices)
}
func cloneSourceStatus(s SourceStatusSpec) SourceStatusSpec {
	s.Notices = copyValues(s.Notices)
	return s
}
func (v SourceStatus) Valid() bool              { return v.valid && validSourceStatus(v.data) }
func (v SourceStatus) Fields() SourceStatusSpec { return cloneSourceStatus(v.data) }
func (v SourceStatus) Clone() SourceStatus      { return SourceStatus{cloneSourceStatus(v.data), v.valid} }

// RowMetaSpec is a constructor/accessor copy; changes cannot mutate RowMeta.
type RowMetaSpec struct {
	Label   string
	Details []string
	Badges  []Badge
	Sources []SourceStatus
}
type RowMeta struct {
	data  RowMetaSpec
	valid bool
}

func NewRowMeta(s RowMetaSpec) (RowMeta, error) {
	if !validRowMeta(s) {
		return RowMeta{}, invalid("RowMeta")
	}
	return RowMeta{cloneRowMeta(s), true}, nil
}
func validRowMeta(s RowMetaSpec) bool { return allValid(s.Badges) && allValid(s.Sources) }
func cloneRowMeta(s RowMetaSpec) RowMetaSpec {
	s.Details = copySlice(s.Details)
	s.Badges = copyValues(s.Badges)
	s.Sources = copyValues(s.Sources)
	return s
}
func (v RowMeta) Valid() bool         { return v.valid && validRowMeta(v.data) }
func (v RowMeta) Fields() RowMetaSpec { return cloneRowMeta(v.data) }
func (v RowMeta) Clone() RowMeta      { return RowMeta{cloneRowMeta(v.data), v.valid} }

// ListStateSpec is a constructor/accessor copy; changes cannot mutate ListState.
type ListStateSpec struct {
	Selected      Optional[ElementID]
	Cursor        Optional[int]
	Scroll        Scroll
	Filter        string
	EditingFilter bool
}
type ListState struct {
	data  ListStateSpec
	valid bool
}

func NewListState(s ListStateSpec) (ListState, error) {
	if !validListState(s) {
		return ListState{}, invalid("ListState")
	}
	return ListState{cloneListState(s), true}, nil
}
func validListState(s ListStateSpec) bool {
	return optionalValid(s.Selected) && (!s.Cursor.present || s.Cursor.value >= 0) && s.Scroll.Valid()
}
func cloneListState(s ListStateSpec) ListStateSpec { ; return s }
func (v ListState) Valid() bool                    { return v.valid && validListState(v.data) }
func (v ListState) Fields() ListStateSpec          { return cloneListState(v.data) }
func (v ListState) Clone() ListState               { return ListState{cloneListState(v.data), v.valid} }

// BranchEndpointSpec is a constructor/accessor copy; changes cannot mutate BranchEndpoint.
type BranchEndpointSpec struct {
	Branch    domain.BranchID
	Revision  Optional[domain.Revision]
	Freshness Freshness
	Evidence  Completeness
}
type BranchEndpoint struct {
	data  BranchEndpointSpec
	valid bool
}

func NewBranchEndpoint(s BranchEndpointSpec) (BranchEndpoint, error) {
	if !validBranchEndpoint(s) {
		return BranchEndpoint{}, invalid("BranchEndpoint")
	}
	return BranchEndpoint{cloneBranchEndpoint(s), true}, nil
}
func validBranchEndpoint(s BranchEndpointSpec) bool {
	return s.Branch.Valid() && optionalValid(s.Revision) && (!s.Revision.present || s.Revision.value.Repository() == s.Branch.Repository()) && s.Freshness.Valid() && s.Evidence.Valid()
}
func cloneBranchEndpoint(s BranchEndpointSpec) BranchEndpointSpec { ; return s }
func (v BranchEndpoint) Valid() bool                              { return v.valid && validBranchEndpoint(v.data) }
func (v BranchEndpoint) Fields() BranchEndpointSpec               { return cloneBranchEndpoint(v.data) }
func (v BranchEndpoint) Clone() BranchEndpoint {
	return BranchEndpoint{cloneBranchEndpoint(v.data), v.valid}
}

// UpstreamSpec is a constructor/accessor copy; changes cannot mutate Upstream.
type UpstreamSpec struct {
	State    UpstreamState
	Endpoint Optional[BranchEndpoint]
	Ahead    Optional[uint64]
	Behind   Optional[uint64]
}
type Upstream struct {
	data  UpstreamSpec
	valid bool
}

func NewUpstream(s UpstreamSpec) (Upstream, error) {
	if !validUpstream(s) {
		return Upstream{}, invalid("Upstream")
	}
	return Upstream{cloneUpstream(s), true}, nil
}
func validUpstream(s UpstreamSpec) bool {
	return s.State.Valid() && optionalValid(s.Endpoint) && (s.State != UpstreamResolved || s.Endpoint.present && s.Endpoint.value.data.Revision.present) && (s.State == UpstreamResolved || !s.Ahead.present && !s.Behind.present) && (s.State != UpstreamNone && s.State != UpstreamNotApplicable || !s.Endpoint.present)
}
func cloneUpstream(s UpstreamSpec) UpstreamSpec { s.Endpoint = copyOptional(s.Endpoint); return s }
func (v Upstream) Valid() bool                  { return v.valid && validUpstream(v.data) }
func (v Upstream) Fields() UpstreamSpec         { return cloneUpstream(v.data) }
func (v Upstream) Clone() Upstream              { return Upstream{cloneUpstream(v.data), v.valid} }

// PRAnnotationSpec is a constructor/accessor copy; changes cannot mutate PRAnnotation.
type PRAnnotationSpec struct {
	ID       domain.PRID
	Role     Relation
	Endpoint Optional[BranchEndpoint]
	Evidence Completeness
	Label    string
}
type PRAnnotation struct {
	data  PRAnnotationSpec
	valid bool
}

func NewPRAnnotation(s PRAnnotationSpec) (PRAnnotation, error) {
	if !validPRAnnotation(s) {
		return PRAnnotation{}, invalid("PRAnnotation")
	}
	return PRAnnotation{clonePRAnnotation(s), true}, nil
}
func validPRAnnotation(s PRAnnotationSpec) bool {
	return s.ID.Valid() && (s.Role == PRHead || s.Role == PRBase || s.Role == RelationUnknown) && optionalValid(s.Endpoint) && s.Evidence.Valid() && (!s.Endpoint.present || s.Endpoint.value.data.Branch.Repository().Scope() == domain.Remote) && (s.Role != PRBase || !s.Endpoint.present || s.Endpoint.value.data.Branch.Repository() == s.ID.Repository())
}
func clonePRAnnotation(s PRAnnotationSpec) PRAnnotationSpec {
	s.Endpoint = copyOptional(s.Endpoint)
	return s
}
func (v PRAnnotation) Valid() bool              { return v.valid && validPRAnnotation(v.data) }
func (v PRAnnotation) Fields() PRAnnotationSpec { return clonePRAnnotation(v.data) }
func (v PRAnnotation) Clone() PRAnnotation      { return PRAnnotation{clonePRAnnotation(v.data), v.valid} }

// WorktreeAnnotationSpec is a constructor/accessor copy; changes cannot mutate WorktreeAnnotation.
type WorktreeAnnotationSpec struct {
	ID           domain.WorktreeID
	Head         Optional[domain.Head]
	Relation     Relation
	Availability Availability
	Current      bool
	Primary      bool
	Active       bool
	Label        string
}
type WorktreeAnnotation struct {
	data  WorktreeAnnotationSpec
	valid bool
}

func NewWorktreeAnnotation(s WorktreeAnnotationSpec) (WorktreeAnnotation, error) {
	if !validWorktreeAnnotation(s) {
		return WorktreeAnnotation{}, invalid("WorktreeAnnotation")
	}
	return WorktreeAnnotation{cloneWorktreeAnnotation(s), true}, nil
}
func validWorktreeAnnotation(s WorktreeAnnotationSpec) bool {
	return s.ID.Valid() && optionalValid(s.Head) && (!s.Head.present || s.Head.value.MatchesWorktree(s.ID)) && (s.Relation >= RelationUnknown && s.Relation <= Unrelated) && s.Availability.Valid()
}
func cloneWorktreeAnnotation(s WorktreeAnnotationSpec) WorktreeAnnotationSpec { ; return s }
func (v WorktreeAnnotation) Valid() bool                                      { return v.valid && validWorktreeAnnotation(v.data) }
func (v WorktreeAnnotation) Fields() WorktreeAnnotationSpec                   { return cloneWorktreeAnnotation(v.data) }
func (v WorktreeAnnotation) Clone() WorktreeAnnotation {
	return WorktreeAnnotation{cloneWorktreeAnnotation(v.data), v.valid}
}

// BranchRowSpec is a constructor/accessor copy; changes cannot mutate BranchRow.
type BranchRowSpec struct {
	ID               domain.BranchID
	ExpectedRevision Optional[domain.Revision]
	Local            Optional[BranchEndpoint]
	Upstream         Optional[Upstream]
	Remote           []BranchEndpoint
	PullRequests     []PRAnnotation
	Worktrees        []WorktreeAnnotation
	Meta             RowMeta
}
type BranchRow struct {
	data  BranchRowSpec
	valid bool
}

func NewBranchRow(s BranchRowSpec) (BranchRow, error) {
	if !validBranchRow(s) {
		return BranchRow{}, invalid("BranchRow")
	}
	return BranchRow{cloneBranchRow(s), true}, nil
}
func validBranchRow(s BranchRowSpec) bool {
	return s.ID.Valid() && optionalValid(s.ExpectedRevision) && (!s.ExpectedRevision.present || s.ExpectedRevision.value.Repository() == s.ID.Repository()) && optionalValid(s.Local) && (!s.Local.present || s.Local.value.data.Branch.Kind() == domain.Local) && optionalValid(s.Upstream) && allValid(s.Remote) && remoteEndpoints(s.Remote) && allValid(s.PullRequests) && allValid(s.Worktrees) && s.Meta.Valid()
}
func cloneBranchRow(s BranchRowSpec) BranchRowSpec {
	s.Local = copyOptional(s.Local)
	s.Upstream = copyOptional(s.Upstream)
	s.Remote = copyValues(s.Remote)
	s.PullRequests = copyValues(s.PullRequests)
	s.Worktrees = copyValues(s.Worktrees)
	s.Meta = s.Meta.Clone()
	return s
}
func (v BranchRow) Valid() bool           { return v.valid && validBranchRow(v.data) }
func (v BranchRow) Fields() BranchRowSpec { return cloneBranchRow(v.data) }
func (v BranchRow) Clone() BranchRow      { return BranchRow{cloneBranchRow(v.data), v.valid} }

// PRRowSpec is a constructor/accessor copy; changes cannot mutate PRRow.
type PRRowSpec struct {
	ID            domain.PRID
	Head          Optional[BranchEndpoint]
	Base          Optional[BranchEndpoint]
	Title         string
	Body          string
	Author        string
	URL           string
	UpdatedAtText string
	UpdatedAt     Optional[Timestamp]
	Draft         bool
	State         PRState
	Worktrees     []WorktreeAnnotation
	Meta          RowMeta
}
type PRRow struct {
	data  PRRowSpec
	valid bool
}

func NewPRRow(s PRRowSpec) (PRRow, error) {
	if !validPRRow(s) {
		return PRRow{}, invalid("PRRow")
	}
	return PRRow{clonePRRow(s), true}, nil
}
func validPRRow(s PRRowSpec) bool {
	return s.ID.Valid() && optionalValid(s.Head) && optionalValid(s.Base) && optionalValid(s.UpdatedAt) && (!s.Head.present || s.Head.value.data.Branch.Repository().Scope() == domain.Remote) && (!s.Base.present || s.Base.value.data.Branch.Repository() == s.ID.Repository()) && s.State.Valid() && allValid(s.Worktrees) && s.Meta.Valid()
}
func clonePRRow(s PRRowSpec) PRRowSpec {
	s.Head = copyOptional(s.Head)
	s.Base = copyOptional(s.Base)
	s.Worktrees = copyValues(s.Worktrees)
	s.Meta = s.Meta.Clone()
	return s
}
func (v PRRow) Valid() bool       { return v.valid && validPRRow(v.data) }
func (v PRRow) Fields() PRRowSpec { return clonePRRow(v.data) }
func (v PRRow) Clone() PRRow      { return PRRow{clonePRRow(v.data), v.valid} }

// FileChangeSpec is a constructor/accessor copy; changes cannot mutate FileChange.
type FileChangeSpec struct {
	Path           string
	OldPath        Optional[string]
	Kind           ChangeKind
	IndexStatus    ChangeStatus
	WorktreeStatus ChangeStatus
	Binary         bool
	Additions      Optional[uint64]
	Deletions      Optional[uint64]
	Meta           RowMeta
}
type FileChange struct {
	data  FileChangeSpec
	valid bool
}

func NewFileChange(s FileChangeSpec) (FileChange, error) {
	if !validFileChange(s) {
		return FileChange{}, invalid("FileChange")
	}
	return FileChange{cloneFileChange(s), true}, nil
}
func validFileChange(s FileChangeSpec) bool {
	return s.Path != "" && (!s.OldPath.present || s.OldPath.value != "") && s.Kind.Valid() && (s.Kind != RenamedChange && s.Kind != CopiedChange || s.OldPath.present) && (s.Kind == RenamedChange || s.Kind == CopiedChange || !s.OldPath.present) && s.IndexStatus.Valid() && s.WorktreeStatus.Valid() && s.Meta.Valid()
}
func cloneFileChange(s FileChangeSpec) FileChangeSpec { s.Meta = s.Meta.Clone(); return s }
func (v FileChange) Valid() bool                      { return v.valid && validFileChange(v.data) }
func (v FileChange) Fields() FileChangeSpec           { return cloneFileChange(v.data) }
func (v FileChange) Clone() FileChange                { return FileChange{cloneFileChange(v.data), v.valid} }

// WorktreeRowSpec is a constructor/accessor copy; changes cannot mutate WorktreeRow.
type WorktreeRowSpec struct {
	ID           domain.WorktreeID
	Head         Optional[domain.Head]
	Locator      string
	Availability Availability
	Current      bool
	Primary      bool
	Active       bool
	Locked       bool
	LockReason   Optional[string]
	Prunable     bool
	Upstream     Optional[Upstream]
	Changes      []FileChange
	Meta         RowMeta
}
type WorktreeRow struct {
	data  WorktreeRowSpec
	valid bool
}

func NewWorktreeRow(s WorktreeRowSpec) (WorktreeRow, error) {
	if !validWorktreeRow(s) {
		return WorktreeRow{}, invalid("WorktreeRow")
	}
	return WorktreeRow{cloneWorktreeRow(s), true}, nil
}
func validWorktreeRow(s WorktreeRowSpec) bool {
	return s.ID.Valid() && optionalValid(s.Head) && (!s.Head.present || s.Head.value.MatchesWorktree(s.ID)) && s.Availability.Valid() && (s.Locked || !s.LockReason.present) && optionalValid(s.Upstream) && allValid(s.Changes) && s.Meta.Valid()
}
func cloneWorktreeRow(s WorktreeRowSpec) WorktreeRowSpec {
	s.Upstream = copyOptional(s.Upstream)
	s.Changes = copyValues(s.Changes)
	s.Meta = s.Meta.Clone()
	return s
}
func (v WorktreeRow) Valid() bool             { return v.valid && validWorktreeRow(v.data) }
func (v WorktreeRow) Fields() WorktreeRowSpec { return cloneWorktreeRow(v.data) }
func (v WorktreeRow) Clone() WorktreeRow      { return WorktreeRow{cloneWorktreeRow(v.data), v.valid} }

// CommitRowSpec is a constructor/accessor copy; changes cannot mutate CommitRow.
type CommitRowSpec struct {
	Revision        domain.Revision
	Parents         []domain.Revision
	Subject         string
	Message         string
	Author          string
	AuthorEmail     string
	Committer       string
	AuthoredAt      Optional[Timestamp]
	CommittedAt     Optional[Timestamp]
	AuthoredAtText  string
	CommittedAtText string
	Meta            RowMeta
}
type CommitRow struct {
	data  CommitRowSpec
	valid bool
}

func NewCommitRow(s CommitRowSpec) (CommitRow, error) {
	if !validCommitRow(s) {
		return CommitRow{}, invalid("CommitRow")
	}
	return CommitRow{cloneCommitRow(s), true}, nil
}
func validCommitRow(s CommitRowSpec) bool {
	return s.Revision.Valid() && revisionsInScope(s.Parents, s.Revision.Repository()) && optionalValid(s.AuthoredAt) && optionalValid(s.CommittedAt) && s.Meta.Valid()
}
func cloneCommitRow(s CommitRowSpec) CommitRowSpec {
	s.Parents = copySlice(s.Parents)
	s.Meta = s.Meta.Clone()
	return s
}
func (v CommitRow) Valid() bool           { return v.valid && validCommitRow(v.data) }
func (v CommitRow) Fields() CommitRowSpec { return cloneCommitRow(v.data) }
func (v CommitRow) Clone() CommitRow      { return CommitRow{cloneCommitRow(v.data), v.valid} }

// StashRowSpec is a constructor/accessor copy; changes cannot mutate StashRow.
type StashRowSpec struct {
	ID             domain.StashID
	PositionLabel  string
	Message        string
	OriginLabel    string
	OriginWorktree Optional[domain.WorktreeID]
	Parents        []domain.Revision
	Managed        Optional[bool]
	CreatedAtText  string
	CreatedAt      Optional[Timestamp]
	Meta           RowMeta
}
type StashRow struct {
	data  StashRowSpec
	valid bool
}

func NewStashRow(s StashRowSpec) (StashRow, error) {
	if !validStashRow(s) {
		return StashRow{}, invalid("StashRow")
	}
	return StashRow{cloneStashRow(s), true}, nil
}
func validStashRow(s StashRowSpec) bool {
	return s.ID.Valid() && optionalValid(s.OriginWorktree) && optionalValid(s.CreatedAt) && (!s.OriginWorktree.present || s.OriginWorktree.value.Repository() == s.ID.Repository()) && revisionsInScope(s.Parents, s.ID.Repository()) && s.Meta.Valid()
}
func cloneStashRow(s StashRowSpec) StashRowSpec {
	s.Parents = copySlice(s.Parents)
	s.Meta = s.Meta.Clone()
	return s
}
func (v StashRow) Valid() bool          { return v.valid && validStashRow(v.data) }
func (v StashRow) Fields() StashRowSpec { return cloneStashRow(v.data) }
func (v StashRow) Clone() StashRow      { return StashRow{cloneStashRow(v.data), v.valid} }

// LaunchRowSpec is a constructor/accessor copy; changes cannot mutate LaunchRow.
type LaunchRowSpec struct {
	ID             domain.LaunchPointID
	SavedAlias     Optional[string]
	SourceLabel    string
	ProjectLabel   string
	ProviderLabel  string
	Default        bool
	Availability   Availability
	OrderedMembers []LaunchMember
	Meta           RowMeta
}
type LaunchRow struct {
	data  LaunchRowSpec
	valid bool
}

func NewLaunchRow(s LaunchRowSpec) (LaunchRow, error) {
	if !validLaunchRow(s) {
		return LaunchRow{}, invalid("LaunchRow")
	}
	return LaunchRow{cloneLaunchRow(s), true}, nil
}
func validLaunchRow(s LaunchRowSpec) bool {
	return s.ID.Valid() && (!s.SavedAlias.present || s.SavedAlias.value != "") && s.Availability.Valid() && allValid(s.OrderedMembers) && launchMembersInScope(s.OrderedMembers, s.ID.Worktree()) && s.Meta.Valid()
}
func cloneLaunchRow(s LaunchRowSpec) LaunchRowSpec {
	s.OrderedMembers = copyValues(s.OrderedMembers)
	s.Meta = s.Meta.Clone()
	return s
}
func (v LaunchRow) Valid() bool           { return v.valid && validLaunchRow(v.data) }
func (v LaunchRow) Fields() LaunchRowSpec { return cloneLaunchRow(v.data) }
func (v LaunchRow) Clone() LaunchRow      { return LaunchRow{cloneLaunchRow(v.data), v.valid} }

// LaunchMemberSpec is a constructor/accessor copy; changes cannot mutate LaunchMember.
type LaunchMemberSpec struct {
	ID            domain.LaunchPointID
	Label         string
	SelectedOrder Optional[uint32]
}
type LaunchMember struct {
	data  LaunchMemberSpec
	valid bool
}

func NewLaunchMember(s LaunchMemberSpec) (LaunchMember, error) {
	if !validLaunchMember(s) {
		return LaunchMember{}, invalid("LaunchMember")
	}
	return LaunchMember{cloneLaunchMember(s), true}, nil
}
func validLaunchMember(s LaunchMemberSpec) bool {
	return s.ID.Valid() && (!s.SelectedOrder.present || s.SelectedOrder.value > 0)
}
func cloneLaunchMember(s LaunchMemberSpec) LaunchMemberSpec { ; return s }
func (v LaunchMember) Valid() bool                          { return v.valid && validLaunchMember(v.data) }
func (v LaunchMember) Fields() LaunchMemberSpec             { return cloneLaunchMember(v.data) }
func (v LaunchMember) Clone() LaunchMember                  { return LaunchMember{cloneLaunchMember(v.data), v.valid} }

// NamespaceRowSpec is a constructor/accessor copy; changes cannot mutate NamespaceRow.
type NamespaceRowSpec struct {
	ID         ElementID
	Parent     Optional[ElementID]
	Depth      int
	Expanded   bool
	ChildCount Optional[uint64]
	Meta       RowMeta
}
type NamespaceRow struct {
	data  NamespaceRowSpec
	valid bool
}

func NewNamespaceRow(s NamespaceRowSpec) (NamespaceRow, error) {
	if !validNamespaceRow(s) {
		return NamespaceRow{}, invalid("NamespaceRow")
	}
	return NamespaceRow{cloneNamespaceRow(s), true}, nil
}
func validNamespaceRow(s NamespaceRowSpec) bool {
	return s.ID.Valid() && (s.ID.kind == NamespaceElement || s.ID.kind == RepositoryElement || s.ID.kind == PRElement || s.ID.kind == BranchElement) && optionalValid(s.Parent) && (!s.Parent.present || s.Parent.value.kind == NamespaceElement || s.Parent.value.kind == RepositoryElement) && s.Depth >= 0 && s.Meta.Valid()
}
func cloneNamespaceRow(s NamespaceRowSpec) NamespaceRowSpec { s.Meta = s.Meta.Clone(); return s }
func (v NamespaceRow) Valid() bool                          { return v.valid && validNamespaceRow(v.data) }
func (v NamespaceRow) Fields() NamespaceRowSpec             { return cloneNamespaceRow(v.data) }
func (v NamespaceRow) Clone() NamespaceRow                  { return NamespaceRow{cloneNamespaceRow(v.data), v.valid} }

// GraphRefSpec is a constructor/accessor copy; changes cannot mutate GraphRef.
type GraphRefSpec struct {
	Name     string
	Kind     RefKind
	Revision domain.Revision
	Branch   Optional[domain.BranchID]
}
type GraphRef struct {
	data  GraphRefSpec
	valid bool
}

func NewGraphRef(s GraphRefSpec) (GraphRef, error) {
	if !validGraphRef(s) {
		return GraphRef{}, invalid("GraphRef")
	}
	return GraphRef{cloneGraphRef(s), true}, nil
}
func validGraphRef(s GraphRefSpec) bool {
	return s.Name != "" && s.Kind.Valid() && s.Revision.Valid() && optionalValid(s.Branch) && (!s.Branch.present || s.Branch.value.Repository() == s.Revision.Repository()) && (s.Kind != BranchRef || s.Branch.present)
}
func cloneGraphRef(s GraphRefSpec) GraphRefSpec { ; return s }
func (v GraphRef) Valid() bool                  { return v.valid && validGraphRef(v.data) }
func (v GraphRef) Fields() GraphRefSpec         { return cloneGraphRef(v.data) }
func (v GraphRef) Clone() GraphRef              { return GraphRef{cloneGraphRef(v.data), v.valid} }

// GraphAnnotationSpec is a constructor/accessor copy; changes cannot mutate GraphAnnotation.
type GraphAnnotationSpec struct {
	Revision     domain.Revision
	PullRequests []PRAnnotation
	Worktrees    []WorktreeAnnotation
	Badges       []Badge
}
type GraphAnnotation struct {
	data  GraphAnnotationSpec
	valid bool
}

func NewGraphAnnotation(s GraphAnnotationSpec) (GraphAnnotation, error) {
	if !validGraphAnnotation(s) {
		return GraphAnnotation{}, invalid("GraphAnnotation")
	}
	return GraphAnnotation{cloneGraphAnnotation(s), true}, nil
}
func validGraphAnnotation(s GraphAnnotationSpec) bool {
	return s.Revision.Valid() && allValid(s.PullRequests) && allValid(s.Worktrees) && allValid(s.Badges)
}
func cloneGraphAnnotation(s GraphAnnotationSpec) GraphAnnotationSpec {
	s.PullRequests = copyValues(s.PullRequests)
	s.Worktrees = copyValues(s.Worktrees)
	s.Badges = copyValues(s.Badges)
	return s
}
func (v GraphAnnotation) Valid() bool                 { return v.valid && validGraphAnnotation(v.data) }
func (v GraphAnnotation) Fields() GraphAnnotationSpec { return cloneGraphAnnotation(v.data) }
func (v GraphAnnotation) Clone() GraphAnnotation {
	return GraphAnnotation{cloneGraphAnnotation(v.data), v.valid}
}

// GraphModelSpec is a constructor/accessor copy; changes cannot mutate GraphModel.
type GraphModelSpec struct {
	Repository       domain.RepositoryID
	SourceGeneration SourceGeneration
	Roots            []domain.Revision
	Commits          []CommitRow
	Refs             []GraphRef
	Annotations      []GraphAnnotation
	BoundaryParents  []domain.Revision
	Sources          []SourceStatus
	List             ListState
	DetailScroll     Scroll
}
type GraphModel struct {
	data  GraphModelSpec
	valid bool
}

func NewGraphModel(s GraphModelSpec) (GraphModel, error) {
	if !validGraphModel(s) {
		return GraphModel{}, invalid("GraphModel")
	}
	return GraphModel{cloneGraphModel(s), true}, nil
}
func validGraphModel(s GraphModelSpec) bool {
	return s.Repository.Valid() && s.Repository.Scope() == domain.LocalCommon && s.SourceGeneration.Valid() && revisionsInScope(s.Roots, s.Repository) && revisionsInScope(s.BoundaryParents, s.Repository) && allValid(s.Commits) && graphScope(s) && allValid(s.Refs) && allValid(s.Annotations) && allValid(s.Sources) && s.List.Valid() && selectedKind(s.List, RevisionElement) && s.DetailScroll.Valid()
}
func cloneGraphModel(s GraphModelSpec) GraphModelSpec {
	s.Roots = copySlice(s.Roots)
	s.Commits = copyValues(s.Commits)
	s.Refs = copyValues(s.Refs)
	s.Annotations = copyValues(s.Annotations)
	s.BoundaryParents = copySlice(s.BoundaryParents)
	s.Sources = copyValues(s.Sources)
	s.List = s.List.Clone()
	return s
}
func (v GraphModel) Valid() bool            { return v.valid && validGraphModel(v.data) }
func (v GraphModel) Fields() GraphModelSpec { return cloneGraphModel(v.data) }
func (v GraphModel) Clone() GraphModel      { return GraphModel{cloneGraphModel(v.data), v.valid} }

// PatchSpec is a constructor/accessor copy; changes cannot mutate Patch.
type PatchSpec struct {
	Bytes             []byte
	Binary            bool
	Truncated         bool
	OriginalByteCount Optional[uint64]
	Notices           []StatusNotice
}
type Patch struct {
	data  PatchSpec
	valid bool
}

func NewPatch(s PatchSpec) (Patch, error) {
	if !validPatch(s) {
		return Patch{}, invalid("Patch")
	}
	return Patch{clonePatch(s), true}, nil
}
func validPatch(s PatchSpec) bool {
	return (!s.OriginalByteCount.present || s.OriginalByteCount.value >= uint64(len(s.Bytes))) && allValid(s.Notices)
}
func clonePatch(s PatchSpec) PatchSpec {
	s.Bytes = copySlice(s.Bytes)
	s.Notices = copyValues(s.Notices)
	return s
}
func (v Patch) Valid() bool       { return v.valid && validPatch(v.data) }
func (v Patch) Fields() PatchSpec { return clonePatch(v.data) }
func (v Patch) Clone() Patch      { return Patch{clonePatch(v.data), v.valid} }

// DiffModelSpec is a constructor/accessor copy; changes cannot mutate DiffModel.
type DiffModelSpec struct {
	Comparison  Comparison
	Files       []FileChange
	Patch       Patch
	Sources     []SourceStatus
	List        ListState
	PatchScroll Scroll
	CanStage    bool
	CanUnstage  bool
}
type DiffModel struct {
	data  DiffModelSpec
	valid bool
}

func NewDiffModel(s DiffModelSpec) (DiffModel, error) {
	if !validDiffModel(s) {
		return DiffModel{}, invalid("DiffModel")
	}
	return DiffModel{cloneDiffModel(s), true}, nil
}
func validDiffModel(s DiffModelSpec) bool {
	return s.Comparison.Valid() && allValid(s.Files) && s.Patch.Valid() && allValid(s.Sources) && s.List.Valid() && selectedKind(s.List, NamespaceElement) && s.PatchScroll.Valid() && (!s.CanStage || s.Comparison.kind == IndexToWorktreeComparison) && (!s.CanUnstage || s.Comparison.kind == HeadToIndexComparison)
}
func cloneDiffModel(s DiffModelSpec) DiffModelSpec {
	s.Files = copyValues(s.Files)
	s.Patch = s.Patch.Clone()
	s.Sources = copyValues(s.Sources)
	s.List = s.List.Clone()
	return s
}
func (v DiffModel) Valid() bool           { return v.valid && validDiffModel(v.data) }
func (v DiffModel) Fields() DiffModelSpec { return cloneDiffModel(v.data) }
func (v DiffModel) Clone() DiffModel      { return DiffModel{cloneDiffModel(v.data), v.valid} }

// TextDetailSpec is a constructor/accessor copy; changes cannot mutate TextDetail.
type TextDetailSpec struct {
	Subject Optional[ElementID]
	Title   string
	Text    string
	Fields  []DetailField
	Scroll  Scroll
}
type TextDetail struct {
	data  TextDetailSpec
	valid bool
}

func NewTextDetail(s TextDetailSpec) (TextDetail, error) {
	if !validTextDetail(s) {
		return TextDetail{}, invalid("TextDetail")
	}
	return TextDetail{cloneTextDetail(s), true}, nil
}
func validTextDetail(s TextDetailSpec) bool {
	return optionalValid(s.Subject) && allValid(s.Fields) && s.Scroll.Valid()
}
func cloneTextDetail(s TextDetailSpec) TextDetailSpec { s.Fields = copyValues(s.Fields); return s }
func (v TextDetail) Valid() bool                      { return v.valid && validTextDetail(v.data) }
func (v TextDetail) Fields() TextDetailSpec           { return cloneTextDetail(v.data) }
func (v TextDetail) Clone() TextDetail                { return TextDetail{cloneTextDetail(v.data), v.valid} }

// DetailFieldSpec is a constructor/accessor copy; changes cannot mutate DetailField.
type DetailFieldSpec struct {
	Label string
	Value string
}
type DetailField struct {
	data  DetailFieldSpec
	valid bool
}

func NewDetailField(s DetailFieldSpec) (DetailField, error) {
	if !validDetailField(s) {
		return DetailField{}, invalid("DetailField")
	}
	return DetailField{cloneDetailField(s), true}, nil
}
func validDetailField(s DetailFieldSpec) bool            { return s.Label != "" }
func cloneDetailField(s DetailFieldSpec) DetailFieldSpec { ; return s }
func (v DetailField) Valid() bool                        { return v.valid && validDetailField(v.data) }
func (v DetailField) Fields() DetailFieldSpec            { return cloneDetailField(v.data) }
func (v DetailField) Clone() DetailField                 { return DetailField{cloneDetailField(v.data), v.valid} }
