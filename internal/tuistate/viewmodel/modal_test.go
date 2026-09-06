package viewmodel_test

import (
	"reflect"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
	vm "github.com/Hans-Einar/gh-tree/internal/tuistate/viewmodel"
)

func textField(id vm.FieldID) vm.FormField {
	return must(vm.NewTextField(id, "label", "value", 2, id == vm.MessageField || id == vm.BodyField, nil))
}
func boolField(id vm.FieldID) vm.FormField { return must(vm.NewBooleanField(id, "label", true, nil)) }
func TestEveryModalPurposeBodyAndAllowedChoice(t *testing.T) {
	r := local("clone")
	w := worktree(r)
	rev := revision(r, 64)
	sid := must(domain.NewStashID(r, rev.OID()))
	lr := launch(w)
	lid := must(vm.NewLaunchElement(lr.Fields().ID, vm.Some("Exact Alias")))
	target := must(vm.NewTargetDetail(vm.TargetDetailSpec{Target: vm.Some(must(domain.NewCommitTarget(rev))), Worktree: vm.Some(w), ExpectedHead: vm.Some(must(domain.NewDetachedHead(rev))), Subjects: []vm.ElementID{lid}, Stashes: []domain.StashID{sid}, Paths: []string{"literal path"}, Fields: []vm.DetailField{must(vm.NewDetailField(vm.DetailFieldSpec{Label: "OID", Value: rev.OID().String()}))}}))
	confirmation := must(vm.NewConfirmationBody(target, []string{"safe supplied summary"}))
	form := func(fs ...vm.FormField) vm.ModalBody { return must(vm.NewFormBody(target, fs)) }
	cases := []struct {
		purpose vm.ModalPurpose
		body    vm.ModalBody
		choice  vm.ChoiceID
	}{
		{vm.ConfirmDeploy, confirmation, vm.StashThenDeployChoice}, {vm.ConfirmRetarget, confirmation, vm.ProceedChoice}, {vm.ConfirmRestore, confirmation, vm.ProceedChoice}, {vm.ConfirmStashCreate, confirmation, vm.ProceedChoice}, {vm.ConfirmStashApply, confirmation, vm.ProceedChoice}, {vm.ConfirmStashPop, confirmation, vm.ProceedChoice}, {vm.ConfirmStashDrop, confirmation, vm.ProceedChoice}, {vm.ConfirmPush, confirmation, vm.ProceedChoice}, {vm.ConfirmQuit, confirmation, vm.ProceedChoice}, {vm.ConfirmAliasReplacement, confirmation, vm.ProceedChoice},
		{vm.CreateWorktreeForm, form(textField(vm.PathField), textField(vm.BranchNameField), boolField(vm.DetachField)), vm.ProceedChoice},
		{vm.NewBranchForm, form(textField(vm.BranchNameField), boolField(vm.CheckoutField)), vm.ProceedChoice},
		{vm.CommitForm, form(textField(vm.MessageField), boolField(vm.StageAllField)), vm.ProceedChoice},
		{vm.PullRequestForm, form(textField(vm.BaseBranchField), textField(vm.TitleField), textField(vm.BodyField), boolField(vm.DraftField), boolField(vm.MaintainerModificationField)), vm.ProceedChoice},
		{vm.SaveLaunchForm, form(textField(vm.AliasField), textField(vm.ExecutableField), boolField(vm.MakeDefaultField)), vm.SaveChoice},
		{vm.WorktreeChooser, must(vm.NewWorktreeChooserBody(target, []vm.WorktreeRow{must(vm.NewWorktreeRow(vm.WorktreeRowSpec{ID: w, Availability: vm.Unavailable, Meta: meta()}))}, emptyList())), vm.SelectChoice},
		{vm.DeployTargetChooser, must(vm.NewDeployChooserBody(target, []vm.ConfiguredTargetRow{must(vm.NewConfiguredTargetRow(vm.ConfiguredTargetRowSpec{ID: must(vm.NewNamespaceElement("configured-target", "production", vm.Some(r))), Label: "Production", Path: "C:/target", Availability: vm.Available, Meta: meta()}))}, emptyList())), vm.SelectChoice},
		{vm.LaunchChooser, must(vm.NewLaunchChooserBody(target, must(vm.NewLaunchModel(vm.LaunchModelSpec{Worktree: w, Rows: []vm.LaunchRow{lr}, List: emptyList()})))), vm.RunChoice},
		{vm.StashPatchDetail, must(vm.NewStashPatchBody(target, must(vm.NewDiffModel(vm.DiffModelSpec{Comparison: must(vm.NewStashComparison(stashDetail(sid, vm.StashWorktreeView))), Patch: must(vm.NewPatch(vm.PatchSpec{Bytes: []byte("full patch")})), List: emptyList()})))), vm.CancelChoice},
		{vm.InspectDetail, must(vm.NewDetailBody(target, must(vm.NewTextDetail(vm.TextDetailSpec{Subject: vm.Some(must(vm.NewRevisionElement(rev))), Text: "full detail"})))), vm.CancelChoice},
	}
	for _, tc := range cases {
		spec := vm.ModalSpec{ID: must(vm.NewModalID("modal")), Owner: must(vm.NewOwnerKey("intent-key")), Purpose: tc.purpose, Title: "Exact target", Body: tc.body, Choices: choices(tc.choice)}
		modal := must(vm.NewModal(spec))
		if !modal.Valid() || modal.Confirmation() != (tc.body.Kind() == vm.ConfirmationBody) {
			t.Fatalf("purpose %v invalid", tc.purpose)
		}
		noSharedMutable(t, reflect.ValueOf(modal), reflect.ValueOf(modal.Clone()))
		spec.Choices[0] = must(vm.NewModalChoice(vm.CancelChoice, "Cancel", false))
		bad, err := vm.NewModal(spec)
		reject(t, bad, err)
		spec = modal.Fields()
		spec.Choices = append(spec.Choices, spec.Choices[0])
		bad, err = vm.NewModal(spec)
		reject(t, bad, err)
	}
	missing := must(vm.NewConfirmationBody(must(vm.NewTargetDetail(vm.TargetDetailSpec{})), nil))
	bad, err := vm.NewModal(vm.ModalSpec{ID: must(vm.NewModalID("id")), Owner: must(vm.NewOwnerKey("owner")), Purpose: vm.ConfirmDeploy, Title: "deploy", Body: missing, Choices: choices(vm.ProceedChoice)})
	reject(t, bad, err)
	bad, err = vm.NewModal(vm.ModalSpec{ID: must(vm.NewModalID("id")), Owner: must(vm.NewOwnerKey("owner")), Purpose: vm.CommitForm, Title: "commit", Body: form(textField(vm.MessageField)), Choices: choices(vm.ProceedChoice)})
	reject(t, bad, err)
	field, err := vm.NewTextField(vm.DraftField, "draft", "true", 0, false, nil)
	reject(t, field, err)
	field, err = vm.NewTextField(vm.MessageField, "message", "界", 2, true, nil)
	reject(t, field, err)
	field, err = vm.NewBooleanField(vm.MessageField, "message", true, nil)
	reject(t, field, err)
}

func TestModalSnapshotAndMeasurementCorrelation(t *testing.T) {
	m := must(vm.NewModal(vm.ModalSpec{ID: must(vm.NewModalID("modal-a")), Owner: must(vm.NewOwnerKey("owner-a")), Purpose: vm.ConfirmQuit, Title: "Quit", Body: must(vm.NewConfirmationBody(must(vm.NewTargetDetail(vm.TargetDetailSpec{})), []string{"stop sessions"})), Choices: choices(vm.ProceedChoice)}))
	nav := must(vm.NewNavigatorPane(header(), must(vm.NewNavigatorModel(vm.NavigatorModelSpec{List: emptyList()}))))
	spec := snapshot(nav, vm.PullRequestsMode, vm.ListPart).Fields()
	spec.Modal = vm.Some(m)
	spec.Focus = must(vm.NewFocusPath(must(vm.NewPanePath(vm.ModalPane, vm.ChoicesPart)), vm.Some(m.Fields().ID), vm.None[vm.FieldID]()))
	s := must(vm.NewSnapshot(spec))
	measure := must(vm.NewMeasurement(vm.MeasurementSpec{Viewport: s.Fields().Viewport, PresentationGeneration: 10, ModalID: vm.Some(m.Fields().ID), ConfirmationPresentable: true, Panes: []vm.PaneRect{must(vm.NewPaneRect(must(vm.NewPanePath(vm.ModalPane, vm.RootPart)), must(vm.NewRect(0, 0, 80, 24))))}}))
	if !measure.Matches(s) {
		t.Fatal("current modal measurement refused")
	}
	mf := measure.Fields()
	mf.ModalID = vm.Some(must(vm.NewModalID("modal-b")))
	if must(vm.NewMeasurement(mf)).Matches(s) {
		t.Fatal("old modal measurement accepted")
	}
	mf = measure.Fields()
	mf.Viewport = must(vm.NewViewport(0, 0, 20))
	mf.Panes = nil
	bad, err := vm.NewMeasurement(mf)
	reject(t, bad, err)
	sf := s.Fields()
	sf.Focus = must(vm.NewFocusPath(must(vm.NewPanePath(vm.ModalPane, vm.FormPart)), vm.Some(m.Fields().ID), vm.Some(vm.MessageField)))
	bs, err := vm.NewSnapshot(sf)
	reject(t, bs, err)
	sf = s.Fields()
	sf.Focus = must(vm.NewFocusPath(must(vm.NewPanePath(vm.NavigatorPane, vm.ListPart)), vm.None[vm.ModalID](), vm.None[vm.FieldID]()))
	bs, err = vm.NewSnapshot(sf)
	reject(t, bs, err)
	sf = s.Fields()
	sf.Actions = []vm.ActionBinding{must(vm.NewActionBinding(vm.ActionBindingSpec{Action: vm.ConfirmChoice, Scope: must(vm.NewModalActionScope(must(vm.NewModalID("modal-b")))), Chord: []vm.KeyStroke{must(vm.NewKeyStroke("enter", 0))}, Applicable: true, Enabled: true, Label: "Proceed"}))}
	bs, err = vm.NewSnapshot(sf)
	reject(t, bs, err)
	noSharedMutable(t, reflect.ValueOf(s), reflect.ValueOf(s.Clone()))
	noSharedMutable(t, reflect.ValueOf(measure), reflect.ValueOf(measure.Clone()))
}
