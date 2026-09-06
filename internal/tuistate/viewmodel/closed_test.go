package viewmodel

import (
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func TestPrivateTagsRefuseContradictoryPayloads(t *testing.T) {
	repo, _ := domain.NewRepositoryID(domain.LocalCommon, "repo")
	oid, _ := domain.NewOID(strings.Repeat("ab", 20))
	rev, _ := domain.NewRevision(repo, oid)
	wt, _ := domain.NewWorktreeID(repo, "main")
	sid, _ := domain.NewSessionID(1)
	element, _ := NewRevisionElement(rev)
	element.worktree = wt
	if element.Valid() {
		t.Fatal("closed element accepted irrelevant identity")
	}
	element, _ = NewSessionElement(sid)
	element.alias = Some("forged")
	if element.Valid() {
		t.Fatal("closed element accepted unrelated saved alias")
	}
	comparison, _ := NewCommitParentComparison(rev, None[domain.Revision]())
	comparison.worktree = wt
	if comparison.Valid() {
		t.Fatal("comparison accepted two alternatives")
	}
	comparison, _ = NewIndexToWorktreeComparison(wt)
	comparison.kind = HeadToIndexComparison
	if comparison.Valid() {
		t.Fatal("comparison accepted missing required HEAD")
	}
	scope := GlobalActionScope()
	scope.session = sid
	if scope.Valid() {
		t.Fatal("global scope accepted unrelated session")
	}
	path, _ := NewPanePath(ConsolePane, InputPart)
	focus, _ := NewFocusPath(path, None[ModalID](), None[FieldID]())
	focus.modal = Some(ModalID{key: "modal"})
	if focus.Valid() {
		t.Fatal("console focus accepted modal identity")
	}
	target, _ := NewTargetDetail(TargetDetailSpec{})
	body, _ := NewConfirmationBody(target, nil)
	body.fields = []FormField{{id: MessageField, kind: TextField, label: "message"}}
	if body.Valid() {
		t.Fatal("confirmation body accepted form payload")
	}
	lm, _ := NewListState(ListStateSpec{})
	nm, _ := NewNavigatorModel(NavigatorModelSpec{List: lm})
	h, _ := NewPaneHeader(PaneHeaderSpec{Title: "pane", Availability: Available, Completeness: Complete, ContentGeneration: 1})
	pane, _ := NewNavigatorPane(h, nm)
	wm, _ := NewWorktreesModel(WorktreesModelSpec{List: lm})
	pane.worktrees = wm
	if pane.Valid() {
		t.Fatal("pane accepted two body variants")
	}
}
