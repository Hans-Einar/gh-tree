package viewmodel

import "github.com/Hans-Einar/gh-tree/internal/domain"

// ActionID names an affordance. It carries no request, callback or executable
// payload; State decides applicability and binds actual product intent privately.
type ActionID uint8

const (
	NavigateUp ActionID = iota + 1
	NavigateDown
	NavigateLeft
	NavigateRight
	NavigateFirst
	NavigateLast
	PageUp
	PageDown
	FocusNext
	FocusPrevious
	FocusNavigator
	FocusWorktrees
	FocusActive
	FocusBranch
	FocusCommits
	FocusMessage
	FocusLaunch
	FocusStashes
	FocusConsole
	SelectConsole
	Search
	ClearSearch
	Back
	Refresh
	SwitchNavigator
	OpenSelection
	LoadMore
	ShowHistory
	ShowGraph
	ShowDiff
	ShowStagedDiff
	ShowDetails
	ActivateWorktree
	ChooseWorktree
	CreateWorktree
	RetargetWorktree
	DeploySelection
	CreateBranch
	Fetch
	Pull
	Push
	StagePath
	UnstagePath
	StageAll
	Commit
	StageAllAndCommit
	RestoreTracked
	CreateStash
	ApplyStash
	PopStash
	DropStash
	ShowStashPatch
	CreatePullRequest
	ChooseLaunch
	SelectLaunchMember
	SaveLaunch
	StartLaunch
	StartDefaultLaunch
	OpenTerminal
	StopSession
	RestartSession
	InterruptSession
	TerminalInput
	ConfirmChoice
	CancelModal
	SubmitForm
	Quit
)

func (v ActionID) Valid() bool { return v >= NavigateUp && v <= Quit }

type Modifier uint8

const (
	ControlModifier Modifier = 1 << iota
	AltModifier
	ShiftModifier
)

type KeyStroke struct {
	key       string
	modifiers Modifier
}

func NewKeyStroke(key string, modifiers Modifier) (KeyStroke, error) {
	k := KeyStroke{key, modifiers}
	if !k.Valid() {
		return KeyStroke{}, invalid("key stroke")
	}
	return k, nil
}
func (k KeyStroke) Valid() bool {
	return k.key != "" && safeLine(k.key) && k.modifiers & ^(ControlModifier|AltModifier|ShiftModifier) == 0
}
func (k KeyStroke) Key() string         { return k.key }
func (k KeyStroke) Modifiers() Modifier { return k.modifiers }

type ScopeKind uint8

const (
	GlobalScope ScopeKind = iota + 1
	PaneScope
	ModalScope
	ConsoleScope
)

func (v ScopeKind) Valid() bool { return v >= GlobalScope && v <= ConsoleScope }

type ActionScope struct {
	kind    ScopeKind
	path    PanePath
	modal   ModalID
	session domain.SessionID
}

func GlobalActionScope() ActionScope { return ActionScope{kind: GlobalScope} }
func NewPaneActionScope(path PanePath) (ActionScope, error) {
	return checkedScope(ActionScope{kind: PaneScope, path: path})
}
func NewModalActionScope(id ModalID) (ActionScope, error) {
	return checkedScope(ActionScope{kind: ModalScope, modal: id})
}
func NewConsoleActionScope(id domain.SessionID) (ActionScope, error) {
	return checkedScope(ActionScope{kind: ConsoleScope, session: id})
}
func checkedScope(s ActionScope) (ActionScope, error) {
	if !s.Valid() {
		return ActionScope{}, invalid("action scope")
	}
	return s, nil
}
func (s ActionScope) Valid() bool {
	rest := s
	rest.kind = 0
	switch s.kind {
	case GlobalScope:
	case PaneScope:
		if !s.path.Valid() || s.path.pane == ModalPane {
			return false
		}
		rest.path = PanePath{}
	case ModalScope:
		if !s.modal.Valid() {
			return false
		}
		rest.modal = ModalID{}
	case ConsoleScope:
		if !s.session.Valid() {
			return false
		}
		rest.session = domain.SessionID{}
	default:
		return false
	}
	return rest == (ActionScope{})
}
func (s ActionScope) Kind() ScopeKind        { return s.kind }
func (s ActionScope) Path() (PanePath, bool) { return s.path, s.kind == PaneScope && s.Valid() }
func (s ActionScope) Modal() (ModalID, bool) { return s.modal, s.kind == ModalScope && s.Valid() }
func (s ActionScope) Session() (domain.SessionID, bool) {
	return s.session, s.kind == ConsoleScope && s.Valid()
}

type ActionBindingSpec struct {
	Action         ActionID
	Chord          []KeyStroke
	Scope          ActionScope
	Applicable     bool
	Enabled        bool
	Label          string
	DisabledReason Optional[string]
}
type ActionBinding struct {
	data  ActionBindingSpec
	valid bool
}

func NewActionBinding(s ActionBindingSpec) (ActionBinding, error) {
	if !validActionBinding(s) {
		return ActionBinding{}, invalid("action binding")
	}
	return ActionBinding{cloneActionBinding(s), true}, nil
}
func validActionBinding(s ActionBindingSpec) bool {
	return s.Action.Valid() && len(s.Chord) > 0 && allValid(s.Chord) && s.Scope.Valid() && s.Label != "" && (!s.Enabled || s.Applicable && !s.DisabledReason.present)
}
func cloneActionBinding(s ActionBindingSpec) ActionBindingSpec {
	s.Chord = copySlice(s.Chord)
	return s
}
func (a ActionBinding) Valid() bool               { return a.valid && validActionBinding(a.data) }
func (a ActionBinding) Fields() ActionBindingSpec { return cloneActionBinding(a.data) }
func (a ActionBinding) Clone() ActionBinding {
	return ActionBinding{cloneActionBinding(a.data), a.valid}
}

type FieldID uint8

const (
	PathField FieldID = iota + 1
	BranchNameField
	MessageField
	BaseBranchField
	TitleField
	BodyField
	DraftField
	AliasField
	MakeDefaultField
	ExecutableField
	DetachField
	CheckoutField
	IncludeUntrackedField
	StageAllField
	MaintainerModificationField
)

func (v FieldID) Valid() bool { return v >= PathField && v <= MaintainerModificationField }

type FieldKind uint8

const (
	TextField FieldKind = iota + 1
	BooleanField
)

func (v FieldKind) Valid() bool { return v == TextField || v == BooleanField }

type FormField struct {
	id                 FieldID
	kind               FieldKind
	label, text        string
	cursor             int
	checked, multiline bool
	notices            []StatusNotice
}

func NewTextField(id FieldID, label, text string, cursor int, multiline bool, notices []StatusNotice) (FormField, error) {
	return checkedField(FormField{id: id, kind: TextField, label: label, text: text, cursor: cursor, multiline: multiline, notices: notices})
}
func NewBooleanField(id FieldID, label string, checked bool, notices []StatusNotice) (FormField, error) {
	return checkedField(FormField{id: id, kind: BooleanField, label: label, checked: checked, notices: notices})
}
func checkedField(f FormField) (FormField, error) {
	if !f.Valid() {
		return FormField{}, invalid("form field")
	}
	return f.Clone(), nil
}
func (f FormField) Valid() bool {
	if !f.id.Valid() || !f.kind.Valid() || f.label == "" || !allValid(f.notices) {
		return false
	}
	boolean := f.id == DraftField || f.id == MakeDefaultField || f.id == DetachField || f.id == CheckoutField || f.id == IncludeUntrackedField || f.id == StageAllField || f.id == MaintainerModificationField
	if boolean {
		return f.kind == BooleanField && f.text == "" && f.cursor == 0 && !f.multiline
	}
	return f.kind == TextField && !f.checked && f.cursor >= 0 && f.cursor <= len([]rune(f.text)) && (!f.multiline || f.id == MessageField || f.id == BodyField)
}
func (f FormField) ID() FieldID     { return f.id }
func (f FormField) Kind() FieldKind { return f.kind }
func (f FormField) Label() string   { return f.label }
func (f FormField) Text() (string, int, bool, bool) {
	return f.text, f.cursor, f.multiline, f.kind == TextField && f.Valid()
}
func (f FormField) Checked() (bool, bool)   { return f.checked, f.kind == BooleanField && f.Valid() }
func (f FormField) Notices() []StatusNotice { return copyValues(f.notices) }
func (f FormField) Clone() FormField        { f.notices = copyValues(f.notices); return f }
