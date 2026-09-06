package viewmodel_test

import (
	"reflect"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
	vm "github.com/Hans-Einar/gh-tree/internal/tuistate/viewmodel"
)

func TestCockpitRetainsMixedPanesAcrossFocusChanges(t *testing.T) {
	repo := local("cockpit")
	branchID := branch(repo)
	navSelection := must(vm.NewBranchElement(branchID))
	// Both selections deliberately survive outside the supplied partial rows.
	commitSelection := must(vm.NewRevisionElement(revision(repo, 40)))
	folderID := must(vm.NewNamespaceElement("folder", "topic", vm.Some(repo)))
	row := must(vm.NewBranchRow(vm.BranchRowSpec{ID: branchID, ExpectedRevision: vm.Some(revision(repo, 64)), Meta: meta()}))
	navList := must(vm.NewListState(vm.ListStateSpec{Selected: vm.Some(navSelection), Cursor: vm.Some(4), Scroll: must(vm.NewScroll(7, 3)), Filter: "navigator filter"}))
	branchList := must(vm.NewListState(vm.ListStateSpec{Selected: vm.Some(commitSelection), Cursor: vm.Some(8), Scroll: must(vm.NewScroll(11, 5)), Filter: "history filter"}))
	navigator := must(vm.NewNavigatorPane(header(), must(vm.NewNavigatorModel(vm.NavigatorModelSpec{
		Content:  vm.BranchesContent,
		Rows:     []vm.NamespaceRow{must(vm.NewNamespaceRow(vm.NamespaceRowSpec{ID: folderID, Meta: meta()}))},
		Branches: []vm.BranchRow{row}, List: navList, Folder: []vm.ElementID{folderID},
	}))))
	branchPane := must(vm.NewBranchPane(header(), must(vm.NewBranchModel(vm.BranchModelSpec{
		Branch: row, Commits: []vm.CommitRow{commit(repo, 64)}, List: branchList,
		DetailScroll: must(vm.NewScroll(9, 2)), MessageScroll: must(vm.NewScroll(17, 4)),
	}))))
	active := must(vm.NewActivePane(header(), must(vm.NewActiveModel(vm.ActiveModelSpec{Changes: emptyList()}))))
	consolePane := must(vm.NewConsolePane(header(), must(vm.NewConsolesModel(vm.ConsolesModelSpec{Rows: []vm.ConsoleModel{console(1, false)}, List: emptyList()}))))
	spec := vm.SnapshotSpec{
		PresentationGeneration: 10, Viewport: must(vm.NewViewport(160, 70, 3)), Mode: vm.CockpitMode,
		Focus:    must(vm.NewFocusPath(must(vm.NewPanePath(vm.NavigatorPane, vm.ListPart)), vm.None[vm.ModalID](), vm.None[vm.FieldID]())),
		Selected: []vm.ElementID{navSelection, commitSelection}, Panes: []vm.PaneModel{navigator, branchPane, active, consolePane},
	}
	initial := must(vm.NewSnapshot(spec))
	before := initial.Fields()
	current := initial
	// These are supplied successor values, not a reducer or key implementation.
	// They represent the retained Alt+N/B/C/M focus destinations in one cockpit.
	for _, path := range []vm.PanePath{
		must(vm.NewPanePath(vm.BranchPane, vm.DetailsPart)),
		must(vm.NewPanePath(vm.BranchPane, vm.ListPart)),
		must(vm.NewPanePath(vm.BranchPane, vm.MessagePart)),
		must(vm.NewPanePath(vm.NavigatorPane, vm.FilterPart)),
		must(vm.NewPanePath(vm.ActivePane, vm.DetailsPart)),
		must(vm.NewPanePath(vm.ConsolePane, vm.BodyPart)),
		must(vm.NewPanePath(vm.NavigatorPane, vm.ListPart)),
	} {
		next := current.Fields()
		next.PresentationGeneration++
		next.Focus = must(vm.NewFocusPath(path, vm.None[vm.ModalID](), vm.None[vm.FieldID]()))
		current = must(vm.NewSnapshot(next))
		got := current.Fields()
		if got.Mode != vm.CockpitMode || got.Focus.Path() != path || !reflect.DeepEqual(got.Panes, before.Panes) || !reflect.DeepEqual(got.Selected, before.Selected) {
			t.Fatalf("focus %v lost independently retained pane facts or selections", path)
		}
		nav, ok := got.Panes[0].Navigator()
		if !ok || nav.Fields().Content != vm.BranchesContent || !reflect.DeepEqual(nav.Fields().List, navList) {
			t.Fatal("navigator content/list state changed with focus")
		}
		bp, ok := got.Panes[1].Branch()
		if !ok || !reflect.DeepEqual(bp.Fields().List, branchList) || bp.Fields().DetailScroll.Vertical() != 9 || bp.Fields().MessageScroll.Vertical() != 17 {
			t.Fatal("branch selection/detail/message scroll changed with focus")
		}
		if got.Panes[0].Header().Fields().Completeness != vm.Partial || got.Panes[1].Header().Fields().Sources[0].Fields().Completeness != vm.Partial || len(nav.Fields().Rows) != 1 || len(bp.Fields().Commits) != 1 {
			t.Fatal("partial pane data was erased or promoted to complete")
		}
		noSharedMutable(t, reflect.ValueOf(initial), reflect.ValueOf(current))
	}
	// The correction does not permit graph/history/diff focus over the cockpit,
	// or an input focus that no selected console actually owns.
	for _, path := range []vm.PanePath{
		must(vm.NewPanePath(vm.GraphPane, vm.ListPart)),
		must(vm.NewPanePath(vm.HistoryPane, vm.MessagePart)),
		must(vm.NewPanePath(vm.DiffPane, vm.PatchPart)),
		must(vm.NewPanePath(vm.ConsolePane, vm.InputPart)),
	} {
		badSpec := current.Fields()
		badSpec.Focus = must(vm.NewFocusPath(path, vm.None[vm.ModalID](), vm.None[vm.FieldID]()))
		bad, err := vm.NewSnapshot(badSpec)
		reject(t, bad, err)
	}
	badSpec := current.Fields()
	badSpec.Mode = vm.GraphMode
	bad, err := vm.NewSnapshot(badSpec)
	reject(t, bad, err)
}

func TestNavigatorContentIsExplicitWithEmptyAndCachedRows(t *testing.T) {
	repo := remote("navigator")
	branchRow := must(vm.NewBranchRow(vm.BranchRowSpec{ID: branch(repo), ExpectedRevision: vm.Some(revision(repo, 64)), Meta: meta()}))
	prRow := must(vm.NewPRRow(vm.PRRowSpec{ID: must(domain.NewPRID(repo, 5)), State: vm.PROpen, Meta: meta()}))
	folder := must(vm.NewNamespaceElement("folder", "same folder", vm.Some(repo)))
	for _, withCache := range []bool{false, true} {
		var last vm.NavigatorModel
		for _, content := range []vm.NavigatorContent{vm.PullRequestsContent, vm.BranchesContent} {
			spec := vm.NavigatorModelSpec{Content: content, List: list(vm.Some(folder)), Folder: []vm.ElementID{folder}}
			if withCache {
				spec.Rows = []vm.NamespaceRow{must(vm.NewNamespaceRow(vm.NamespaceRowSpec{ID: folder, Meta: meta()}))}
				spec.Branches = []vm.BranchRow{branchRow}
				spec.PullRequests = []vm.PRRow{prRow}
			}
			model := must(vm.NewNavigatorModel(spec))
			// Identical titles, selection and cached rows cannot decide content.
			h := header().Fields()
			h.Title = "same title for either content"
			s := snapshot(must(vm.NewNavigatorPane(must(vm.NewPaneHeader(h)), model)), vm.CockpitMode, vm.ListPart)
			got, _ := s.Fields().Panes[0].Navigator()
			if got.Fields().Content != content {
				t.Fatal("navigator content was inferred or lost")
			}
			if last.Valid() {
				old := last.Fields()
				next := got.Fields()
				old.Content = next.Content
				if !reflect.DeepEqual(old, next) {
					t.Fatal("content switch changed retained row/selection state")
				}
			}
			// Explicit navigator content also survives branch-pane focus while
			// both cached content families remain in the shared cockpit.
			mixed := s.Fields()
			mixed.Panes = append(mixed.Panes, must(vm.NewBranchPane(header(), must(vm.NewBranchModel(vm.BranchModelSpec{Branch: branchRow, List: emptyList()})))))
			mixed.Focus = must(vm.NewFocusPath(must(vm.NewPanePath(vm.BranchPane, vm.DetailsPart)), vm.None[vm.ModalID](), vm.None[vm.FieldID]()))
			combined := must(vm.NewSnapshot(mixed))
			retained, _ := combined.Fields().Panes[0].Navigator()
			if !reflect.DeepEqual(retained, got) {
				t.Fatal("branch focus changed explicit navigator content or caches")
			}
			last = got
		}
	}
	for _, content := range []vm.NavigatorContent{0, 255} {
		bad, err := vm.NewNavigatorModel(vm.NavigatorModelSpec{Content: content, List: emptyList()})
		reject(t, bad, err)
		if content.Valid() {
			t.Fatal("invalid navigator content tag accepted")
		}
	}
}
