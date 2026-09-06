package viewmodel

import "github.com/Hans-Einar/gh-tree/internal/domain"

type SnapshotSpec struct {
	PresentationGeneration PresentationGeneration
	Viewport               Viewport
	Mode                   Mode
	Focus                  FocusPath
	Selected               []ElementID
	Panes                  []PaneModel
	Modal                  Optional[Modal]
	Actions                []ActionBinding
	Status                 []StatusNotice
	AnimationFrame         uint64
	VersionDisplay         string
}
type Snapshot struct {
	data  SnapshotSpec
	valid bool
}

func NewSnapshot(s SnapshotSpec) (Snapshot, error) {
	if !validSnapshot(s) {
		return Snapshot{}, invalid("snapshot")
	}
	return Snapshot{cloneSnapshot(s), true}, nil
}
func validSnapshot(s SnapshotSpec) bool {
	if !s.PresentationGeneration.Valid() || !s.Viewport.Valid() || !s.Mode.Valid() || !s.Focus.Valid() || !allValid(s.Selected) || !allValid(s.Panes) || !optionalValid(s.Modal) || !allValid(s.Actions) || !allValid(s.Status) {
		return false
	}
	if s.Modal.present != s.Focus.modal.present || s.Modal.present && s.Modal.value.data.ID != s.Focus.modal.value {
		return false
	}
	if !paneInMode(s.Focus.path.pane, s.Mode) {
		return false
	}
	panes := map[Pane]bool{}
	selected := map[ElementID]bool{}
	sessions := map[domain.SessionID]ConsoleModel{}
	for _, id := range s.Selected {
		if selected[id] {
			return false
		}
		selected[id] = true
	}
	for _, p := range s.Panes {
		if panes[p.kind] || !paneInMode(p.kind, s.Mode) {
			return false
		}
		panes[p.kind] = true
		if id := p.Selection(); id.present && !selected[id.value] {
			return false
		}
		if p.kind == ConsolePane {
			for _, c := range p.console.data.Rows {
				sessions[c.data.SessionID] = c
				if c.data.InputFocused && (s.Focus.path.pane != ConsolePane || s.Focus.path.part != InputPart) {
					return false
				}
			}
		}
	}
	if s.Focus.path.pane != ModalPane && !panes[s.Focus.path.pane] {
		return false
	}
	if s.Focus.path.pane == ConsolePane && s.Focus.path.part == InputPart {
		found := false
		for _, c := range sessions {
			found = found || c.data.InputFocused
		}
		if !found {
			return false
		}
	}
	if s.Modal.present && !modalFocus(s.Focus, s.Modal.value) {
		return false
	}
	for _, a := range s.Actions {
		switch a.data.Scope.kind {
		case PaneScope:
			if !panes[a.data.Scope.path.pane] {
				return false
			}
		case ModalScope:
			if !s.Modal.present || a.data.Scope.modal != s.Modal.value.data.ID {
				return false
			}
		case ConsoleScope:
			if _, ok := sessions[a.data.Scope.session]; !ok {
				return false
			}
		}
	}
	return true
}
func modalFocus(f FocusPath, m Modal) bool {
	b := m.data.Body
	switch f.path.part {
	case RootPart, BodyPart, ChoicesPart:
		return true
	case FormPart:
		if b.kind != FormBody {
			return false
		}
		for _, field := range b.fields {
			if field.id == f.field.value {
				return true
			}
		}
		return false
	case ListPart:
		return b.kind == WorktreeChooserBody || b.kind == DeployChooserBody || b.kind == LaunchChooserBody
	}
	return false
}
func cloneSnapshot(s SnapshotSpec) SnapshotSpec {
	s.Selected = copySlice(s.Selected)
	s.Panes = copyValues(s.Panes)
	s.Modal = copyOptional(s.Modal)
	s.Actions = copyValues(s.Actions)
	s.Status = copyValues(s.Status)
	return s
}
func (s Snapshot) Valid() bool          { return s.valid && validSnapshot(s.data) }
func (s Snapshot) Fields() SnapshotSpec { return cloneSnapshot(s.data) }
func (s Snapshot) Clone() Snapshot      { return Snapshot{cloneSnapshot(s.data), s.valid} }

type PaneRect struct {
	path PanePath
	rect Rect
}

func NewPaneRect(path PanePath, rect Rect) (PaneRect, error) {
	p := PaneRect{path, rect}
	if !p.Valid() {
		return PaneRect{}, invalid("pane rectangle")
	}
	return p, nil
}
func (p PaneRect) Valid() bool    { return p.path.Valid() && p.rect.Valid() }
func (p PaneRect) Path() PanePath { return p.path }
func (p PaneRect) Rect() Rect     { return p.rect }

type ConsoleRect struct {
	session    domain.SessionID
	generation ContentGeneration
	rect       Rect
}

func NewConsoleRect(session domain.SessionID, generation ContentGeneration, rect Rect) (ConsoleRect, error) {
	c := ConsoleRect{session, generation, rect}
	if !c.Valid() {
		return ConsoleRect{}, invalid("console rectangle")
	}
	return c, nil
}
func (c ConsoleRect) Valid() bool                          { return c.session.Valid() && c.generation.Valid() && c.rect.Valid() }
func (c ConsoleRect) SessionID() domain.SessionID          { return c.session }
func (c ConsoleRect) ContentGeneration() ContentGeneration { return c.generation }
func (c ConsoleRect) Rect() Rect                           { return c.rect }

// The constructor checks supplied rectangles, not layout policy. The View owns
// geometry, and State/host must match correlation before approval/native resize.
type MeasurementSpec struct {
	Viewport                Viewport
	PresentationGeneration  PresentationGeneration
	ModalID                 Optional[ModalID]
	ConfirmationPresentable bool
	Panes                   []PaneRect
	Consoles                []ConsoleRect
}
type Measurement struct {
	data  MeasurementSpec
	valid bool
}

func NewMeasurement(s MeasurementSpec) (Measurement, error) {
	if !validMeasurement(s) {
		return Measurement{}, invalid("measurement")
	}
	return Measurement{cloneMeasurement(s), true}, nil
}
func validMeasurement(s MeasurementSpec) bool {
	if !s.Viewport.Valid() || !s.PresentationGeneration.Valid() || !optionalValid(s.ModalID) || s.ConfirmationPresentable && !s.ModalID.present || !allValid(s.Panes) || !allValid(s.Consoles) {
		return false
	}
	panes := map[PanePath]bool{}
	consoles := map[domain.SessionID]bool{}
	for _, p := range s.Panes {
		if panes[p.path] || !p.rect.Within(s.Viewport) {
			return false
		}
		panes[p.path] = true
	}
	for _, c := range s.Consoles {
		if consoles[c.session] || !c.rect.Within(s.Viewport) {
			return false
		}
		consoles[c.session] = true
	}
	// A zero-area known viewport cannot show any confirmation, regardless of
	// caller assertion. Adequate positive geometry is the later View's proof.
	if s.ConfirmationPresentable && s.Viewport.known && (s.Viewport.width == 0 || s.Viewport.height == 0) {
		return false
	}
	return true
}
func cloneMeasurement(s MeasurementSpec) MeasurementSpec {
	s.Panes = copySlice(s.Panes)
	s.Consoles = copySlice(s.Consoles)
	return s
}
func (m Measurement) Valid() bool                            { return m.valid && validMeasurement(m.data) }
func (m Measurement) Fields() MeasurementSpec                { return cloneMeasurement(m.data) }
func (m Measurement) Clone() Measurement                     { return Measurement{cloneMeasurement(m.data), m.valid} }
func (m Measurement) ViewportGeneration() ViewportGeneration { return m.data.Viewport.generation }
func (m Measurement) PresentationGeneration() PresentationGeneration {
	return m.data.PresentationGeneration
}
func (m Measurement) Matches(s Snapshot) bool {
	if !m.Valid() || !s.Valid() || m.data.PresentationGeneration != s.data.PresentationGeneration || m.data.Viewport != s.data.Viewport || m.data.ModalID.present != s.data.Modal.present {
		return false
	}
	if m.data.ModalID.present && m.data.ModalID.value != s.data.Modal.value.data.ID {
		return false
	}
	if m.data.ConfirmationPresentable && (!s.data.Modal.present || !s.data.Modal.value.Confirmation()) {
		return false
	}
	panes := map[Pane]bool{}
	sessions := map[domain.SessionID]ContentGeneration{}
	for _, p := range s.data.Panes {
		panes[p.kind] = true
		if p.kind == ConsolePane {
			for _, c := range p.console.data.Rows {
				sessions[c.data.SessionID] = c.data.ContentGeneration
			}
		}
	}
	for _, p := range m.data.Panes {
		if p.path.pane == ModalPane {
			if !s.data.Modal.present {
				return false
			}
		} else if !panes[p.path.pane] {
			return false
		}
	}
	for _, c := range m.data.Consoles {
		if generation, ok := sessions[c.session]; !ok || generation != c.generation {
			return false
		}
	}
	return true
}
