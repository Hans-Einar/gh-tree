package viewmodel

type ModalPurpose uint8

const (
	ConfirmDeploy ModalPurpose = iota + 1
	ConfirmRetarget
	ConfirmRestore
	ConfirmStashCreate
	ConfirmStashApply
	ConfirmStashPop
	ConfirmStashDrop
	ConfirmPush
	ConfirmQuit
	ConfirmAliasReplacement
	CreateWorktreeForm
	NewBranchForm
	CommitForm
	PullRequestForm
	SaveLaunchForm
	WorktreeChooser
	DeployTargetChooser
	LaunchChooser
	StashPatchDetail
	InspectDetail
)

func (v ModalPurpose) Valid() bool { return v >= ConfirmDeploy && v <= InspectDetail }

type BodyKind uint8

const (
	ConfirmationBody BodyKind = iota + 1
	FormBody
	WorktreeChooserBody
	DeployChooserBody
	LaunchChooserBody
	StashPatchBody
	DetailBody
)

func (v BodyKind) Valid() bool { return v >= ConfirmationBody && v <= DetailBody }

type ChoiceID uint8

const (
	CancelChoice ChoiceID = iota + 1
	ProceedChoice
	StashThenDeployChoice
	SelectChoice
	RunChoice
	SaveChoice
)

func (v ChoiceID) Valid() bool { return v >= CancelChoice && v <= SaveChoice }

type ModalChoice struct {
	choice  ChoiceID
	label   string
	enabled bool
}

func NewModalChoice(choice ChoiceID, label string, enabled bool) (ModalChoice, error) {
	c := ModalChoice{choice, label, enabled}
	if !c.Valid() {
		return ModalChoice{}, invalid("modal choice")
	}
	return c, nil
}
func (c ModalChoice) Valid() bool      { return c.choice.Valid() && c.label != "" }
func (c ModalChoice) Choice() ChoiceID { return c.choice }
func (c ModalChoice) Label() string    { return c.label }
func (c ModalChoice) Enabled() bool    { return c.enabled }

// ModalBody is a closed presentation union. All constructors retain exact
// target detail; a local owner/key never grants Application approval authority.
type ModalBody struct {
	kind       BodyKind
	target     TargetDetail
	paragraphs []string
	fields     []FormField
	worktrees  []WorktreeRow
	targets    []ConfiguredTargetRow
	launch     LaunchModel
	list       ListState
	detail     TextDetail
	diff       DiffModel
}

func NewConfirmationBody(target TargetDetail, paragraphs []string) (ModalBody, error) {
	return checkedBody(ModalBody{kind: ConfirmationBody, target: target, paragraphs: paragraphs})
}
func NewFormBody(target TargetDetail, fields []FormField) (ModalBody, error) {
	return checkedBody(ModalBody{kind: FormBody, target: target, fields: fields})
}
func NewWorktreeChooserBody(target TargetDetail, rows []WorktreeRow, list ListState) (ModalBody, error) {
	return checkedBody(ModalBody{kind: WorktreeChooserBody, target: target, worktrees: rows, list: list})
}
func NewDeployChooserBody(target TargetDetail, rows []ConfiguredTargetRow, list ListState) (ModalBody, error) {
	return checkedBody(ModalBody{kind: DeployChooserBody, target: target, targets: rows, list: list})
}
func NewLaunchChooserBody(target TargetDetail, launch LaunchModel) (ModalBody, error) {
	return checkedBody(ModalBody{kind: LaunchChooserBody, target: target, launch: launch})
}
func NewStashPatchBody(target TargetDetail, diff DiffModel) (ModalBody, error) {
	return checkedBody(ModalBody{kind: StashPatchBody, target: target, diff: diff})
}
func NewDetailBody(target TargetDetail, detail TextDetail) (ModalBody, error) {
	return checkedBody(ModalBody{kind: DetailBody, target: target, detail: detail})
}
func checkedBody(b ModalBody) (ModalBody, error) {
	if !b.Valid() {
		return ModalBody{}, invalid("modal body")
	}
	return b.Clone(), nil
}
func (b ModalBody) Valid() bool {
	if !b.kind.Valid() || !b.target.Valid() {
		return false
	}
	if b.kind != ConfirmationBody && len(b.paragraphs) != 0 || b.kind != FormBody && len(b.fields) != 0 || b.kind != WorktreeChooserBody && len(b.worktrees) != 0 || b.kind != DeployChooserBody && len(b.targets) != 0 || b.kind != LaunchChooserBody && b.launch.valid || b.kind != DetailBody && b.detail.valid || b.kind != StashPatchBody && b.diff.valid || b.kind != WorktreeChooserBody && b.kind != DeployChooserBody && b.list.valid {
		return false
	}
	switch b.kind {
	case ConfirmationBody:
		return true
	case FormBody:
		return len(b.fields) > 0 && allValid(b.fields) && uniqueFields(b.fields)
	case WorktreeChooserBody:
		return allValid(b.worktrees) && b.list.Valid() && selectedKind(b.list, WorktreeElement)
	case DeployChooserBody:
		return allValid(b.targets) && b.list.Valid() && selectedKind(b.list, NamespaceElement) && b.target.data.Target.present
	case LaunchChooserBody:
		return b.launch.Valid()
	case StashPatchBody:
		return b.diff.Valid() && b.diff.data.Comparison.kind == StashComparison && len(b.target.data.Stashes) == 1 && b.target.data.Stashes[0] == b.diff.data.Comparison.stashDetail.stash
	case DetailBody:
		return b.detail.Valid()
	}
	return false
}
func (b ModalBody) Kind() BodyKind       { return b.kind }
func (b ModalBody) Target() TargetDetail { return b.target.Clone() }
func (b ModalBody) Paragraphs() []string { return copySlice(b.paragraphs) }
func (b ModalBody) Fields() []FormField  { return copyValues(b.fields) }
func (b ModalBody) Worktrees() ([]WorktreeRow, ListState, bool) {
	return copyValues(b.worktrees), b.list.Clone(), b.kind == WorktreeChooserBody && b.Valid()
}
func (b ModalBody) DeployTargets() ([]ConfiguredTargetRow, ListState, bool) {
	return copyValues(b.targets), b.list.Clone(), b.kind == DeployChooserBody && b.Valid()
}
func (b ModalBody) Launch() (LaunchModel, bool) {
	return b.launch.Clone(), b.kind == LaunchChooserBody && b.Valid()
}
func (b ModalBody) Detail() (TextDetail, bool) {
	return b.detail.Clone(), b.kind == DetailBody && b.Valid()
}
func (b ModalBody) Diff() (DiffModel, bool) {
	return b.diff.Clone(), b.kind == StashPatchBody && b.Valid()
}
func (b ModalBody) Clone() ModalBody {
	b.target = b.target.Clone()
	b.paragraphs = copySlice(b.paragraphs)
	b.fields = copyValues(b.fields)
	b.worktrees = copyValues(b.worktrees)
	b.targets = copyValues(b.targets)
	b.launch = b.launch.Clone()
	b.list = b.list.Clone()
	b.detail = b.detail.Clone()
	b.diff = b.diff.Clone()
	return b
}

type ModalSpec struct {
	ID      ModalID
	Owner   OwnerKey
	Purpose ModalPurpose
	Title   string
	Body    ModalBody
	Choices []ModalChoice
	Scroll  Scroll
}
type Modal struct {
	data  ModalSpec
	valid bool
}

func NewModal(s ModalSpec) (Modal, error) {
	if !validModal(s) {
		return Modal{}, invalid("modal")
	}
	return Modal{cloneModal(s), true}, nil
}
func validModal(s ModalSpec) bool {
	return s.ID.Valid() && s.Owner.Valid() && s.Purpose.Valid() && s.Title != "" && s.Body.Valid() && s.Scroll.Valid() && modalPurpose(s.Purpose, s.Body) && modalChoices(s.Purpose, s.Choices)
}
func cloneModal(s ModalSpec) ModalSpec {
	s.Body = s.Body.Clone()
	s.Choices = copySlice(s.Choices)
	return s
}
func (m Modal) Valid() bool        { return m.valid && validModal(m.data) }
func (m Modal) Fields() ModalSpec  { return cloneModal(m.data) }
func (m Modal) Clone() Modal       { return Modal{cloneModal(m.data), m.valid} }
func (m Modal) Confirmation() bool { return m.Valid() && m.data.Body.kind == ConfirmationBody }

func uniqueFields(fs []FormField) bool {
	seen := map[FieldID]bool{}
	for _, f := range fs {
		if seen[f.id] {
			return false
		}
		seen[f.id] = true
	}
	return true
}
func modalPurpose(p ModalPurpose, b ModalBody) bool {
	t := b.target.data
	if p >= ConfirmDeploy && p <= ConfirmAliasReplacement {
		if b.kind != ConfirmationBody {
			return false
		}
		switch p {
		case ConfirmDeploy, ConfirmRetarget:
			return t.Target.present && t.Worktree.present
		case ConfirmRestore:
			return t.Worktree.present && len(t.Paths) > 0
		case ConfirmStashCreate:
			return t.Worktree.present
		case ConfirmStashApply, ConfirmStashPop:
			return t.Worktree.present && len(t.Stashes) == 1
		case ConfirmStashDrop:
			return len(t.Stashes) == 1
		case ConfirmPush:
			return t.Worktree.present && t.Target.present
		case ConfirmQuit:
			return true
		case ConfirmAliasReplacement:
			for _, s := range t.Subjects {
				if s.kind == LaunchElement && s.alias.present {
					return true
				}
			}
			return false
		}
	}
	switch p {
	case CreateWorktreeForm, NewBranchForm, CommitForm, PullRequestForm, SaveLaunchForm:
		return b.kind == FormBody && formPurpose(p, b)
	case WorktreeChooser:
		return b.kind == WorktreeChooserBody
	case DeployTargetChooser:
		return b.kind == DeployChooserBody
	case LaunchChooser:
		return b.kind == LaunchChooserBody
	case StashPatchDetail:
		return b.kind == StashPatchBody
	case InspectDetail:
		return b.kind == DetailBody
	}
	return false
}
func formPurpose(p ModalPurpose, b ModalBody) bool {
	var allowed, required []FieldID
	switch p {
	case CreateWorktreeForm:
		allowed = []FieldID{PathField, BranchNameField, DetachField}
		required = []FieldID{PathField, BranchNameField}
		if !b.target.data.Target.present {
			return false
		}
	case NewBranchForm:
		allowed = []FieldID{BranchNameField, CheckoutField}
		required = []FieldID{BranchNameField}
		if !b.target.data.Target.present || !b.target.data.Worktree.present {
			return false
		}
	case CommitForm:
		allowed = []FieldID{MessageField, StageAllField}
		required = []FieldID{MessageField, StageAllField}
		if !b.target.data.Worktree.present {
			return false
		}
	case PullRequestForm:
		allowed = []FieldID{BaseBranchField, TitleField, BodyField, DraftField, MaintainerModificationField}
		required = []FieldID{BaseBranchField, TitleField, BodyField, DraftField}
		if !b.target.data.Target.present {
			return false
		}
	case SaveLaunchForm:
		allowed = []FieldID{AliasField, MakeDefaultField, ExecutableField}
		required = []FieldID{AliasField, MakeDefaultField}
		found := false
		for _, s := range b.target.data.Subjects {
			found = found || s.kind == LaunchElement
		}
		if !found {
			return false
		}
	}
	for _, f := range b.fields {
		found := false
		for _, id := range allowed {
			found = found || f.id == id
		}
		if !found {
			return false
		}
	}
	for _, id := range required {
		found := false
		for _, f := range b.fields {
			found = found || f.id == id
		}
		if !found {
			return false
		}
	}
	return true
}
func modalChoices(p ModalPurpose, cs []ModalChoice) bool {
	if len(cs) == 0 || !allValid(cs) {
		return false
	}
	seen := map[ChoiceID]bool{}
	cancel := false
	for _, c := range cs {
		if seen[c.choice] {
			return false
		}
		seen[c.choice] = true
		if c.choice == CancelChoice {
			cancel = c.enabled
			continue
		}
		allowed := false
		switch {
		case p >= ConfirmDeploy && p <= ConfirmAliasReplacement:
			allowed = c.choice == ProceedChoice || p == ConfirmDeploy && c.choice == StashThenDeployChoice
		case p >= CreateWorktreeForm && p <= SaveLaunchForm:
			allowed = c.choice == SaveChoice || c.choice == ProceedChoice
		case p == WorktreeChooser || p == DeployTargetChooser:
			allowed = c.choice == SelectChoice
		case p == LaunchChooser:
			allowed = c.choice == RunChoice || c.choice == SaveChoice
		}
		if !allowed {
			return false
		}
	}
	return cancel
}
