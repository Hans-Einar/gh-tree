package viewmodel_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
	vm "github.com/Hans-Einar/gh-tree/internal/tuistate/viewmodel"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
func reject[T any](t *testing.T, v T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("accepted invalid %T: %#v", v, v)
	}
	if !reflect.DeepEqual(v, *new(T)) {
		t.Fatalf("failed constructor returned nonzero %T", v)
	}
}
func local(s string) domain.RepositoryID  { return must(domain.NewRepositoryID(domain.LocalCommon, s)) }
func remote(s string) domain.RepositoryID { return must(domain.NewRepositoryID(domain.Remote, s)) }
func revision(r domain.RepositoryID, n int) domain.Revision {
	return must(domain.NewRevision(r, must(domain.NewOID(strings.Repeat("ab", n/2)))))
}
func worktree(r domain.RepositoryID) domain.WorktreeID {
	return must(domain.NewWorktreeID(r, "primary"))
}
func branch(r domain.RepositoryID) domain.BranchID {
	k := domain.Local
	if r.Scope() == domain.Remote {
		k = domain.RemoteHead
	}
	return must(domain.NewBranchID(r, k, "topic"))
}
func meta() vm.RowMeta {
	return must(vm.NewRowMeta(vm.RowMetaSpec{Label: "title", Details: []string{"full detail"}, Badges: []vm.Badge{must(vm.NewBadge(vm.BadgeSpec{Label: "unknown", Severity: vm.Warning, Relation: vm.Some(vm.RelationUnknown)}))}, Sources: []vm.SourceStatus{source()}}))
}
func source() vm.SourceStatus {
	return must(vm.NewSourceStatus(vm.SourceStatusSpec{Label: "remote", Availability: vm.Available, Completeness: vm.Partial, Freshness: vm.Cached, Generation: vm.Some(vm.SourceGeneration(2)), Notices: []vm.StatusNotice{must(vm.NewStatusNotice(vm.StatusNoticeSpec{Text: "bounded partial source", Severity: vm.Warning}))}}))
}
func list(id vm.Optional[vm.ElementID]) vm.ListState {
	return must(vm.NewListState(vm.ListStateSpec{Selected: id, Scroll: must(vm.NewScroll(2, 3)), Filter: "exact filter"}))
}
func emptyList() vm.ListState { return list(vm.None[vm.ElementID]()) }
func header() vm.PaneHeader {
	return must(vm.NewPaneHeader(vm.PaneHeaderSpec{Title: "pane", Availability: vm.Available, Completeness: vm.Partial, ContentGeneration: 3, Sources: []vm.SourceStatus{source()}}))
}
func endpoint(r domain.RepositoryID, n int) vm.BranchEndpoint {
	return must(vm.NewBranchEndpoint(vm.BranchEndpointSpec{Branch: branch(r), Revision: vm.Some(revision(r, n)), Freshness: vm.Fresh, Evidence: vm.Complete}))
}
func commit(r domain.RepositoryID, n int) vm.CommitRow {
	return must(vm.NewCommitRow(vm.CommitRowSpec{Revision: revision(r, n), Subject: "subject\x1b[2J stays verbatim", Message: "first line\n\nfull message\n", Parents: []domain.Revision{}, Meta: meta()}))
}
func file() vm.FileChange {
	return must(vm.NewFileChange(vm.FileChangeSpec{Path: "new name", OldPath: vm.Some("old name"), Kind: vm.RenamedChange, IndexStatus: vm.StatusClean, WorktreeStatus: vm.StatusRenamed, Meta: meta()}))
}
func launch(w domain.WorktreeID) vm.LaunchRow {
	id := must(domain.NewLaunchPointID(w, "make", "nested", "build"))
	return must(vm.NewLaunchRow(vm.LaunchRowSpec{ID: id, SavedAlias: vm.Some("Build"), SourceLabel: "Makefile", ProjectLabel: "nested", ProviderLabel: "make", Availability: vm.Available, OrderedMembers: []vm.LaunchMember{must(vm.NewLaunchMember(vm.LaunchMemberSpec{ID: id, Label: "build", SelectedOrder: vm.Some(uint32(1))}))}, Meta: meta()}))
}
func console(id uint64, input bool) vm.ConsoleModel {
	sid := must(domain.NewSessionID(id))
	r := must(vm.NewByteRange(0, 4))
	line := must(vm.NewConsoleLine(vm.ConsoleLineSpec{Text: "safe", Stream: vm.TerminalStream, SourceRange: r}))
	p := must(vm.NewConsolePresentation(vm.ConsolePresentationSpec{SessionID: sid, SourceGeneration: 5, PresentationGeneration: 6, ContentGeneration: 7, Range: r, Lines: []vm.ConsoleLine{line}}))
	return must(vm.NewConsoleModel(vm.ConsoleModelSpec{SessionID: sid, SourceGeneration: 5, PresentationGeneration: 6, ContentGeneration: 7, Phase: vm.RunningSession, Cleanup: vm.CleanupPending, Activity: vm.ActivityUnknown, Capabilities: vm.NewConsoleCapabilities(true, true, true, true), InputFocused: input, Summary: must(vm.NewConsoleSummary(vm.ConsoleSummarySpec{Label: "shell", WorktreeID: worktree(local("clone")), Terminal: true, ArgumentDisplay: []string{"safe argument"}})), Presentation: vm.Some(p)}))
}
func snapshot(p vm.PaneModel, mode vm.Mode, part vm.Part) vm.Snapshot {
	ids := []vm.ElementID{}
	if id, ok := p.Selection().Value(); ok {
		ids = append(ids, id)
	}
	return must(vm.NewSnapshot(vm.SnapshotSpec{PresentationGeneration: 10, Viewport: must(vm.NewViewport(80, 24, 20)), Mode: mode, Focus: must(vm.NewFocusPath(must(vm.NewPanePath(p.Kind(), part)), vm.None[vm.ModalID](), vm.None[vm.FieldID]())), Selected: ids, Panes: []vm.PaneModel{p}, VersionDisplay: "0.4.0", AnimationFrame: 12}))
}

func TestExternalConsumerCompleteTypedPaneInventory(t *testing.T) {
	r := local("clone")
	rr := remote("base")
	w := worktree(r)
	rev := revision(r, 64)
	b := branch(r)
	pr := must(domain.NewPRID(rr, 4))
	target := must(domain.NewBranchTarget(b, rev))
	head := must(domain.NewAttachedHead(b, rev))
	wr := must(vm.NewWorktreeRow(vm.WorktreeRowSpec{ID: w, Head: vm.Some(head), Locator: "C:/clone", Availability: vm.Available, Current: true, Primary: true, Active: true, Changes: []vm.FileChange{file()}, Meta: meta()}))
	pa := must(vm.NewPRAnnotation(vm.PRAnnotationSpec{ID: pr, Role: vm.PRHead, Endpoint: vm.Some(endpoint(remote("fork"), 64)), Evidence: vm.Partial}))
	wa := must(vm.NewWorktreeAnnotation(vm.WorktreeAnnotationSpec{ID: w, Head: vm.Some(head), Relation: vm.ExactSelectedRevision, Availability: vm.Available, Active: true}))
	up := must(vm.NewUpstream(vm.UpstreamSpec{State: vm.UpstreamResolved, Endpoint: vm.Some(endpoint(rr, 64)), Ahead: vm.Some(uint64(0))}))
	br := must(vm.NewBranchRow(vm.BranchRowSpec{ID: b, ExpectedRevision: vm.Some(rev), Local: vm.Some(endpoint(r, 64)), Remote: []vm.BranchEndpoint{endpoint(rr, 64)}, Upstream: vm.Some(up), PullRequests: []vm.PRAnnotation{pa}, Worktrees: []vm.WorktreeAnnotation{wa}, Meta: meta()}))
	prrow := must(vm.NewPRRow(vm.PRRowSpec{ID: pr, Head: vm.Some(endpoint(remote("fork"), 64)), Base: vm.Some(endpoint(rr, 64)), Title: "full PR title", Body: "full PR body", State: vm.PROpen, Worktrees: []vm.WorktreeAnnotation{wa}, Meta: meta()}))
	wid := must(vm.NewWorktreeElement(w))
	rid := must(vm.NewRevisionElement(rev))
	bid := must(vm.NewBranchElement(b))
	sid := must(domain.NewStashID(r, rev.OID()))
	stash := must(vm.NewStashRow(vm.StashRowSpec{ID: sid, PositionLabel: "stash@{19}", Message: "exact stash", Parents: []domain.Revision{rev}, OriginWorktree: vm.Some(w), Meta: meta()}))
	launchrow := launch(w)
	graph := must(vm.NewGraphModel(vm.GraphModelSpec{Repository: r, SourceGeneration: 4, Roots: []domain.Revision{rev}, Commits: []vm.CommitRow{commit(r, 64)}, Refs: []vm.GraphRef{must(vm.NewGraphRef(vm.GraphRefSpec{Name: "refs/heads/topic", Kind: vm.BranchRef, Revision: rev, Branch: vm.Some(b)}))}, Annotations: []vm.GraphAnnotation{must(vm.NewGraphAnnotation(vm.GraphAnnotationSpec{Revision: rev, PullRequests: []vm.PRAnnotation{pa}, Worktrees: []vm.WorktreeAnnotation{wa}}))}, Sources: []vm.SourceStatus{source()}, List: list(vm.Some(rid))}))
	diff := must(vm.NewDiffModel(vm.DiffModelSpec{Comparison: must(vm.NewIndexToWorktreeComparison(w)), Files: []vm.FileChange{file()}, Patch: must(vm.NewPatch(vm.PatchSpec{Bytes: []byte("diff\n\x1b[2J"), Truncated: true, OriginalByteCount: vm.Some(uint64(100))})), Sources: []vm.SourceStatus{source()}, List: emptyList(), CanStage: true}))
	cases := []struct {
		pane vm.PaneModel
		mode vm.Mode
		part vm.Part
	}{
		{must(vm.NewNavigatorPane(header(), must(vm.NewNavigatorModel(vm.NavigatorModelSpec{Rows: []vm.NamespaceRow{must(vm.NewNamespaceRow(vm.NamespaceRowSpec{ID: bid, Meta: meta()}))}, Branches: []vm.BranchRow{br}, PullRequests: []vm.PRRow{prrow}, List: list(vm.Some(bid))})))), vm.BranchesMode, vm.ListPart},
		{must(vm.NewWorktreesPane(header(), must(vm.NewWorktreesModel(vm.WorktreesModelSpec{Rows: []vm.WorktreeRow{wr}, List: list(vm.Some(wid))})))), vm.PullRequestsMode, vm.ListPart},
		{must(vm.NewActivePane(header(), must(vm.NewActiveModel(vm.ActiveModelSpec{Worktree: vm.Some(wr), Changes: emptyList()})))), vm.PullRequestsMode, vm.ChangesPart},
		{must(vm.NewBranchPane(header(), must(vm.NewBranchModel(vm.BranchModelSpec{Branch: br, Commits: []vm.CommitRow{commit(r, 64)}, List: list(vm.Some(rid))})))), vm.BranchContextMode, vm.MessagePart},
		{must(vm.NewHistoryPane(header(), must(vm.NewHistoryModel(vm.HistoryModelSpec{Source: bid, Target: vm.Some(target), Commits: []vm.CommitRow{commit(r, 64)}, List: list(vm.Some(rid))})))), vm.HistoryMode, vm.ListPart},
		{must(vm.NewGraphPane(header(), graph)), vm.GraphMode, vm.ListPart},
		{must(vm.NewDiffPane(header(), diff)), vm.DiffMode, vm.PatchPart},
		{must(vm.NewLaunchPane(header(), must(vm.NewLaunchModel(vm.LaunchModelSpec{Worktree: w, Rows: []vm.LaunchRow{launchrow}, List: list(vm.Some(must(vm.NewLaunchElement(launchrow.Fields().ID, launchrow.Fields().SavedAlias))))})))), vm.PullRequestsMode, vm.ListPart},
		{must(vm.NewStashesPane(header(), must(vm.NewStashesModel(vm.StashesModelSpec{Repository: r, Rows: []vm.StashRow{stash}, List: list(vm.Some(must(vm.NewStashElement(sid))))})))), vm.PullRequestsMode, vm.ListPart},
		{must(vm.NewConsolePane(header(), must(vm.NewConsolesModel(vm.ConsolesModelSpec{Rows: []vm.ConsoleModel{console(1, true)}, List: list(vm.Some(must(vm.NewSessionElement(must(domain.NewSessionID(1))))))})))), vm.PullRequestsMode, vm.InputPart},
	}
	for _, tc := range cases {
		s := snapshot(tc.pane, tc.mode, tc.part)
		if !s.Valid() || !s.Clone().Valid() || s.Fields().VersionDisplay != "0.4.0" || s.Fields().AnimationFrame != 12 {
			t.Fatalf("pane %v cannot publish immutable snapshot", tc.pane.Kind())
		}
	}
	if graph.Fields().Commits[0].Fields().Message != "first line\n\nfull message\n" || len(stash.Fields().ID.OID().String()) != 64 || prrow.Fields().Head.Present() == false || br.Fields().ExpectedRevision.Present() == false {
		t.Fatal("detail identity/message lost")
	}
}

func TestClosedIdentityAndScopeInvalidity(t *testing.T) {
	r := local("a")
	other := local("b")
	w := worktree(r)
	cases := []vm.ElementID{
		must(vm.NewNamespaceElement("folder", "/", vm.Some(remote("base")))), must(vm.NewRepositoryElement(r)), must(vm.NewPRElement(must(domain.NewPRID(remote("base"), 2)))), must(vm.NewBranchElement(branch(r))), must(vm.NewRevisionElement(revision(r, 40))), must(vm.NewWorktreeElement(w)), must(vm.NewStashElement(must(domain.NewStashID(r, revision(r, 64).OID())))), must(vm.NewLaunchElement(launch(w).Fields().ID, vm.Some("Alias"))), must(vm.NewSessionElement(must(domain.NewSessionID(1)))),
	}
	seen := map[vm.ElementID]bool{}
	for _, id := range cases {
		if !id.Valid() || seen[id] {
			t.Fatal("closed element invalid or aliased")
		}
		seen[id] = true
	}
	a := must(vm.NewLaunchElement(launch(w).Fields().ID, vm.Some("Alias")))
	b := must(vm.NewLaunchElement(launch(w).Fields().ID, vm.Some("alias")))
	c := must(vm.NewLaunchElement(launch(w).Fields().ID, vm.None[string]()))
	if a == b || a == c {
		t.Fatal("saved aliases collapsed")
	}
	v, e := vm.NewLaunchElement(launch(w).Fields().ID, vm.Some(""))
	reject(t, v, e)
	v, e = vm.NewNamespaceElement("n", "key", vm.Some(domain.RepositoryID{}))
	reject(t, v, e)
	be, err := vm.NewBranchEndpoint(vm.BranchEndpointSpec{Branch: branch(r), Revision: vm.Some(revision(other, 40)), Freshness: vm.Fresh, Evidence: vm.Complete})
	reject(t, be, err)
	wr, err := vm.NewWorktreeRow(vm.WorktreeRowSpec{ID: w, Head: vm.Some(must(domain.NewDetachedHead(revision(other, 40)))), Availability: vm.Available, Meta: meta()})
	reject(t, wr, err)
	br, err := vm.NewBranchRow(vm.BranchRowSpec{ID: branch(r), ExpectedRevision: vm.Some(domain.Revision{}), Meta: meta()})
	reject(t, br, err)
	pr, err := vm.NewPRRow(vm.PRRowSpec{ID: must(domain.NewPRID(remote("base"), 1)), Base: vm.Some(endpoint(remote("fork"), 40)), State: vm.PROpen, Meta: meta()})
	reject(t, pr, err)
	up, err := vm.NewUpstream(vm.UpstreamSpec{State: vm.UpstreamUnresolved, Ahead: vm.Some(uint64(0))})
	reject(t, up, err)
	if !must(vm.NewUpstream(vm.UpstreamSpec{State: vm.UpstreamUnresolved})).Valid() {
		t.Fatal("unknown count incorrectly forced to zero")
	}
	fc, err := vm.NewFileChange(vm.FileChangeSpec{Path: "file", Kind: vm.RenamedChange, IndexStatus: vm.StatusClean, WorktreeStatus: vm.StatusRenamed, Meta: meta()})
	reject(t, fc, err)
}

func TestAllComparisonVariantsPreserveExactEndpoints(t *testing.T) {
	r := local("clone")
	rr := remote("base")
	fork := remote("fork")
	w := worktree(r)
	rev := revision(r, 64)
	base := revision(rr, 64)
	target := must(domain.NewPullRequestTarget(must(domain.NewPRID(rr, 5)), revision(fork, 64)))
	cases := []vm.Comparison{must(vm.NewCommitParentComparison(rev, vm.None[domain.Revision]())), must(vm.NewCommitPairComparison(rev, revision(r, 40))), must(vm.NewIndexToWorktreeComparison(w)), must(vm.NewHeadToIndexComparison(w, must(domain.NewUnbornHead(branch(r))))), must(vm.NewPullRequestComparison(target, base, rev, rev, revision(r, 40))), must(vm.NewStashComparison(stashDetail(must(domain.NewStashID(r, rev.OID())), vm.StashIndexView)))}
	for _, c := range cases {
		if !c.Valid() {
			t.Fatal("valid comparison refused")
		}
	}
	old, exact, ok := cases[0].CommitEndpoints()
	if !ok || old.Present() || exact != rev {
		t.Fatal("root commit lost explicit absence")
	}
	pr, b, lb, lh, mb, ok := cases[4].PullRequest()
	if !ok || pr != target || b != base || lb != rev || lh != rev || mb != revision(r, 40) {
		t.Fatal("PR exact/resolved scopes lost")
	}
	bad, err := vm.NewCommitPairComparison(rev, revision(local("foreign"), 64))
	reject(t, bad, err)
	bad, err = vm.NewPullRequestComparison(target, revision(fork, 64), rev, rev, rev)
	reject(t, bad, err)
	bad, err = vm.NewPullRequestComparison(target, base, rev, revision(r, 40), rev)
	reject(t, bad, err)
	bad, err = vm.NewHeadToIndexComparison(w, must(domain.NewDetachedHead(revision(local("foreign"), 64))))
	reject(t, bad, err)
}

func TestDeepCopiesAcrossRowsGraphPatchModalAndConsole(t *testing.T) {
	r := local("clone")
	w := worktree(r)
	rev := revision(r, 64)
	fields := []vm.DetailField{must(vm.NewDetailField(vm.DetailFieldSpec{Label: "revision", Value: rev.OID().String()}))}
	paths := []string{"original"}
	subjects := []vm.ElementID{must(vm.NewRevisionElement(rev))}
	target := must(vm.NewTargetDetail(vm.TargetDetailSpec{Target: vm.Some(must(domain.NewCommitTarget(rev))), Worktree: vm.Some(w), Subjects: subjects, Paths: paths, Fields: fields}))
	body := must(vm.NewConfirmationBody(target, []string{"original body"}))
	modal := must(vm.NewModal(vm.ModalSpec{ID: must(vm.NewModalID("modal")), Owner: must(vm.NewOwnerKey("intent")), Purpose: vm.ConfirmDeploy, Title: "Deploy", Body: body, Choices: choices(vm.ProceedChoice)}))
	paths[0] = "changed"
	subjects[0] = must(vm.NewWorktreeElement(w))
	fields[0] = must(vm.NewDetailField(vm.DetailFieldSpec{Label: "bad", Value: "bad"}))
	mf := modal.Fields()
	tf := mf.Body.Target().Fields()
	tf.Paths[0] = "again"
	tf.Fields[0] = fields[0]
	mf.Choices[0] = must(vm.NewModalChoice(vm.ProceedChoice, "bad", true))
	paragraphs := mf.Body.Paragraphs()
	paragraphs[0] = "bad"
	got := modal.Clone().Fields()
	if got.Body.Target().Fields().Paths[0] != "original" || got.Body.Target().Fields().Fields[0].Fields().Value != rev.OID().String() || got.Body.Paragraphs()[0] != "original body" || got.Choices[0].Choice() != vm.CancelChoice {
		t.Fatal("deep modal data escaped")
	}
	bytes := []byte("original\x1b[2J")
	patch := must(vm.NewPatch(vm.PatchSpec{Bytes: bytes, Notices: []vm.StatusNotice{must(vm.NewStatusNotice(vm.StatusNoticeSpec{Text: "truncated", Severity: vm.Warning}))}}))
	bytes[0] = 'X'
	pf := patch.Fields()
	pf.Bytes[1] = 'X'
	pf.Notices[0] = vm.StatusNotice{}
	if string(patch.Clone().Fields().Bytes) != "original\x1b[2J" || !patch.Fields().Notices[0].Valid() {
		t.Fatal("patch copy escaped")
	}
	cm := console(1, false)
	cf := cm.Fields()
	ps, _ := cf.Presentation.Value()
	lines := ps.Fields()
	lines.Lines[0] = vm.ConsoleLine{}
	summ := cf.Summary.Fields()
	summ.ArgumentDisplay[0] = "changed"
	if !cm.Fields().Presentation.Present() || cm.Fields().Summary.Fields().ArgumentDisplay[0] != "safe argument" {
		t.Fatal("console nested copy escaped")
	}
	graphSpec := vm.GraphModelSpec{Repository: r, SourceGeneration: 1, Roots: []domain.Revision{rev}, Commits: []vm.CommitRow{commit(r, 64)}, Sources: []vm.SourceStatus{source()}, List: emptyList()}
	g := must(vm.NewGraphModel(graphSpec))
	graphSpec.Roots[0] = revision(r, 40)
	graphSpec.Commits[0] = vm.CommitRow{}
	gf := g.Fields()
	gf.Roots[0] = revision(r, 40)
	cmf := gf.Commits[0].Fields()
	mm := cmf.Meta.Fields()
	mm.Details[0] = "bad"
	mm.Badges[0] = vm.Badge{}
	ss := mm.Sources[0].Fields()
	ss.Notices[0] = vm.StatusNotice{}
	if g.Fields().Roots[0] != rev || g.Fields().Commits[0].Fields().Meta.Fields().Details[0] != "full detail" || !g.Fields().Commits[0].Fields().Meta.Fields().Sources[0].Fields().Notices[0].Valid() {
		t.Fatal("graph deep copy escaped")
	}
}

func choices(extra vm.ChoiceID) []vm.ModalChoice {
	out := []vm.ModalChoice{must(vm.NewModalChoice(vm.CancelChoice, "Cancel", true))}
	if extra != vm.CancelChoice {
		out = append(out, must(vm.NewModalChoice(extra, "Continue", true)))
	}
	return out
}

func stashDetail(id domain.StashID, view vm.StashView) vm.StashComparisonDetail {
	s := vm.StashComparisonDetailSpec{Stash: id, View: view, Parents: []domain.OID{id.OID(), id.OID()}, FromTree: vm.Some(id.OID()), ToTree: vm.Some(id.OID())}
	if view == vm.StashParentView {
		s.ParentIndex = vm.Some(uint32(0))
	}
	if view == vm.StashUntrackedView {
		s.FromTree = vm.None[domain.OID]()
		s.ToTree = vm.None[domain.OID]()
	}
	return must(vm.NewStashComparisonDetail(s))
}

func TestOutputOffsetsSafetyAndSessionCorrelation(t *testing.T) {
	sid := must(domain.NewSessionID(1))
	input := vm.OutputInputSpec{SessionID: sid, SourceGeneration: 1, PresentationGeneration: 2, ContentGeneration: 3, Offset: 8, Bytes: []byte("raw\x1b"), Stream: vm.StdoutStream, Gaps: []vm.ByteRange{must(vm.NewByteRange(2, 8))}, End: true}
	v := must(vm.NewOutputInput(input))
	input.Bytes[0] = 'X'
	input.Gaps[0] = must(vm.NewByteRange(0, 2))
	f := v.Fields()
	f.Bytes[0] = 'Y'
	f.Gaps[0] = vm.ByteRange{}
	if string(v.Fields().Bytes) != "raw\x1b" || v.Fields().Gaps[0].Start() != 2 {
		t.Fatal("raw output ownership escaped")
	}
	for _, text := range []string{"\x1b[2J", "unsafe\nline", "tab\t", "\xff", "\u0085"} {
		line, err := vm.NewConsoleLine(vm.ConsoleLineSpec{Text: text, Stream: vm.StdoutStream})
		reject(t, line, err)
	}
	for _, text := range []string{"界 e\u0301 😀", ""} {
		if !must(vm.NewConsoleLine(vm.ConsoleLineSpec{Text: text, Stream: vm.StdoutStream})).Valid() {
			t.Fatal("safe text refused")
		}
	}
	bad := v.Fields()
	bad.Offset = ^uint64(0)
	out, err := vm.NewOutputInput(bad)
	reject(t, out, err)
	bad = v.Fields()
	bad.SourceGeneration = 0
	out, err = vm.NewOutputInput(bad)
	reject(t, out, err)
	bad = v.Fields()
	bad.Gaps = []vm.ByteRange{must(vm.NewByteRange(0, 5)), must(vm.NewByteRange(4, 8))}
	out, err = vm.NewOutputInput(bad)
	reject(t, out, err)
	cm := console(1, false).Fields()
	cm.SessionID = must(domain.NewSessionID(2))
	model, err := vm.NewConsoleModel(cm)
	reject(t, model, err)
	cm = console(1, false).Fields()
	cm.ContentGeneration++
	model, err = vm.NewConsoleModel(cm)
	reject(t, model, err)
	cm = console(1, false).Fields()
	cm.Phase = vm.CleanedSession
	model, err = vm.NewConsoleModel(cm)
	reject(t, model, err)
	cm = console(1, true).Fields()
	cm.Phase = vm.StoppingSession
	model, err = vm.NewConsoleModel(cm)
	reject(t, model, err)
}

func TestViewportMeasurementFocusAndGeneration(t *testing.T) {
	if !vm.UnknownViewport().Valid() || vm.UnknownViewport().Known() {
		t.Fatal("unknown startup viewport invalid")
	}
	for _, size := range []int{0, 1, 2, 80} {
		v := must(vm.NewViewport(size, size, 1))
		if !v.Known() || v.Width() != size || !must(vm.NewRect(0, 0, size, size)).Within(v) {
			t.Fatal("known size replaced by fallback")
		}
		if must(vm.NewRect(size, 0, 1, 1)).Within(v) {
			t.Fatal("rectangle exceeds bound")
		}
	}
	vp, e := vm.NewViewport(80, 24, 0)
	reject(t, vp, e)
	vp, e = vm.NewViewport(-1, 24, 1)
	reject(t, vp, e)
	rect, e := vm.NewRect(int(^uint(0)>>1), 0, 1, 1)
	reject(t, rect, e)
	pane := must(vm.NewConsolePane(header(), must(vm.NewConsolesModel(vm.ConsolesModelSpec{Rows: []vm.ConsoleModel{console(1, false)}, List: emptyList()}))))
	s := snapshot(pane, vm.PullRequestsMode, vm.BodyPart)
	m := must(vm.NewMeasurement(vm.MeasurementSpec{Viewport: s.Fields().Viewport, PresentationGeneration: 10, Panes: []vm.PaneRect{must(vm.NewPaneRect(must(vm.NewPanePath(vm.ConsolePane, vm.BodyPart)), must(vm.NewRect(0, 0, 80, 24))))}, Consoles: []vm.ConsoleRect{must(vm.NewConsoleRect(must(domain.NewSessionID(1)), 7, must(vm.NewRect(1, 1, 78, 22))))}}))
	if !m.Matches(s) {
		t.Fatal("current measurement did not match")
	}
	for _, change := range []func(*vm.MeasurementSpec){func(x *vm.MeasurementSpec) { x.PresentationGeneration++ }, func(x *vm.MeasurementSpec) { x.Viewport = must(vm.NewViewport(80, 24, 21)) }, func(x *vm.MeasurementSpec) {
		x.Consoles[0] = must(vm.NewConsoleRect(must(domain.NewSessionID(2)), 7, must(vm.NewRect(0, 0, 1, 1))))
	}, func(x *vm.MeasurementSpec) {
		x.Consoles[0] = must(vm.NewConsoleRect(must(domain.NewSessionID(1)), 8, must(vm.NewRect(0, 0, 1, 1))))
	}} {
		f := m.Fields()
		change(&f)
		if must(vm.NewMeasurement(f)).Matches(s) {
			t.Fatal("stale measurement accepted")
		}
	}
	f := m.Fields()
	f.Consoles[0] = must(vm.NewConsoleRect(must(domain.NewSessionID(1)), 7, must(vm.NewRect(0, 0, 81, 24))))
	bad, err := vm.NewMeasurement(f)
	reject(t, bad, err)
	f = m.Fields()
	f.ConfirmationPresentable = true
	bad, err = vm.NewMeasurement(f)
	reject(t, bad, err)
	sf := s.Fields()
	sf.Mode = vm.GraphMode
	sf.Focus = must(vm.NewFocusPath(must(vm.NewPanePath(vm.GraphPane, vm.ListPart)), vm.None[vm.ModalID](), vm.None[vm.FieldID]()))
	bs, err := vm.NewSnapshot(sf)
	reject(t, bs, err)
	sf = s.Fields()
	sf.Focus = must(vm.NewFocusPath(must(vm.NewPanePath(vm.ConsolePane, vm.InputPart)), vm.None[vm.ModalID](), vm.None[vm.FieldID]()))
	bs, err = vm.NewSnapshot(sf)
	reject(t, bs, err)
}
