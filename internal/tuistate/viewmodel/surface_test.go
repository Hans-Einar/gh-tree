package viewmodel_test

import (
	"github.com/Hans-Einar/gh-tree/internal/domain"
	vm "github.com/Hans-Einar/gh-tree/internal/tuistate/viewmodel"
	"reflect"
	"testing"
)

func TestInvalidZeroAndClosedPublicRepresentation(t *testing.T) {
	values := []interface{ Valid() bool }{vm.StatusNotice{}, vm.Badge{}, vm.SourceStatus{}, vm.RowMeta{}, vm.ListState{}, vm.BranchEndpoint{}, vm.Upstream{}, vm.PRAnnotation{}, vm.WorktreeAnnotation{}, vm.BranchRow{}, vm.PRRow{}, vm.FileChange{}, vm.WorktreeRow{}, vm.CommitRow{}, vm.StashRow{}, vm.LaunchRow{}, vm.LaunchMember{}, vm.NamespaceRow{}, vm.GraphRef{}, vm.GraphAnnotation{}, vm.GraphModel{}, vm.Patch{}, vm.DiffModel{}, vm.TextDetail{}, vm.DetailField{}, vm.TargetDetail{}, vm.ConfiguredTargetRow{}, vm.NavigatorModel{}, vm.WorktreesModel{}, vm.ActiveModel{}, vm.BranchModel{}, vm.HistoryModel{}, vm.LaunchModel{}, vm.StashesModel{}, vm.ConsolesModel{}, vm.PaneHeader{}, vm.OutputInput{}, vm.ConsoleLine{}, vm.ConsolePresentation{}, vm.ConsoleSummary{}, vm.ConsoleModel{}, vm.ActionBinding{}, vm.PaneModel{}, vm.Snapshot{}, vm.Measurement{}, vm.ElementID{}, vm.Comparison{}, vm.Modal{}, vm.ModalBody{}, vm.ModalID{}, vm.OwnerKey{}, vm.FocusPath{}, vm.PanePath{}, vm.FormField{}, vm.ModalChoice{}, vm.ConsoleCapabilities{}, vm.PaneRect{}, vm.ConsoleRect{}, vm.KeyStroke{}, vm.ActionScope{}}
	for _, v := range values {
		if v.Valid() {
			t.Fatalf("zero %T accepted", v)
		}
		typ := reflect.TypeOf(v)
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).IsExported() || typ.Field(i).Anonymous {
				t.Fatalf("%s exposes mutable representation", typ)
			}
		}
		for i := 0; i < typ.NumMethod(); i++ {
			if typ.Method(i).Name == "Set" {
				t.Fatalf("%s has public mutation", typ)
			}
		}
	}
	enums := []interface{ Valid() bool }{vm.Mode(0), vm.Mode(255), vm.Pane(0), vm.Pane(255), vm.Part(0), vm.Part(255), vm.Relation(0), vm.Relation(255), vm.Availability(0), vm.Availability(255), vm.Completeness(0), vm.Completeness(255), vm.SourceGeneration(0), vm.ContentGeneration(0), vm.ViewportGeneration(0), vm.PresentationGeneration(0), vm.ModalPurpose(0), vm.ModalPurpose(255), vm.ChoiceID(0), vm.ChoiceID(255), vm.ActionID(0), vm.ActionID(255), vm.StreamKind(0), vm.StreamKind(255), vm.CleanupState(0), vm.CleanupState(255), vm.BodyKind(0), vm.BodyKind(255)}
	for _, v := range enums {
		if v.Valid() {
			t.Fatalf("invalid enum %T accepted", v)
		}
	}
}

// Inspect actual nested backing arrays as well as semantic mutation controls.
// Strings/Domain fixed arrays may share immutable values; no mutable slice/map
// backing is shared by a family's admission/access/clone boundary.
func noSharedMutable(t *testing.T, a, b reflect.Value) {
	t.Helper()
	if a.Type() != b.Type() {
		t.Fatal("copy type changed")
	}
	switch a.Kind() {
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			noSharedMutable(t, a.Field(i), b.Field(i))
		}
	case reflect.Slice:
		if a.Len() > 0 && a.Pointer() == b.Pointer() {
			t.Fatalf("shared slice backing: %s", a.Type())
		}
		for i := 0; i < a.Len(); i++ {
			noSharedMutable(t, a.Index(i), b.Index(i))
		}
	case reflect.Map, reflect.Pointer:
		if !a.IsNil() && a.Pointer() == b.Pointer() {
			t.Fatalf("shared mutable backing: %s", a.Type())
		}
	}
}
func TestSnapshotRecursiveBackingCopies(t *testing.T) {
	pane := must(vm.NewConsolePane(header(), must(vm.NewConsolesModel(vm.ConsolesModelSpec{Rows: []vm.ConsoleModel{console(1, false)}, List: list(vm.Some(must(vm.NewSessionElement(must(domain.NewSessionID(1))))))}))))
	s := snapshot(pane, vm.PullRequestsMode, vm.BodyPart)
	f := s.Fields()
	clone := s.Clone()
	noSharedMutable(t, reflect.ValueOf(s), reflect.ValueOf(clone))
	noSharedMutable(t, reflect.ValueOf(f), reflect.ValueOf(s.Fields()))
	f.Panes[0] = vm.PaneModel{}
	f.Selected[0] = vm.ElementID{}
	if !s.Valid() || !clone.Valid() {
		t.Fatal("snapshot mutated by accessor")
	}
	key := must(vm.NewKeyStroke("q", 0))
	keys := []vm.KeyStroke{key}
	a := must(vm.NewActionBinding(vm.ActionBindingSpec{Action: vm.Quit, Chord: keys, Scope: vm.GlobalActionScope(), Applicable: false, Enabled: false, Label: "Quit"}))
	keys[0] = vm.KeyStroke{}
	if !a.Valid() {
		t.Fatal("key chord escaped")
	}
	af := a.Fields()
	af.Chord[0] = vm.KeyStroke{}
	if !a.Valid() {
		t.Fatal("key chord accessor escaped")
	}
	as := a.Fields()
	as.Enabled = true
	bad, err := vm.NewActionBinding(as)
	reject(t, bad, err)
}
