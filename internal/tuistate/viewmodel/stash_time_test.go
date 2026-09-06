package viewmodel_test

import (
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/domain"
	vm "github.com/Hans-Einar/gh-tree/internal/tuistate/viewmodel"
)

func TestFullStashViewsAndSuppliedSourceTime(t *testing.T) {
	r := local("clone")
	oid := revision(r, 64).OID()
	id := must(domain.NewStashID(r, oid))
	for _, view := range []vm.StashView{vm.StashWorktreeView, vm.StashIndexView, vm.StashUntrackedView, vm.StashIndexToWorktreeView, vm.StashParentView} {
		d := stashDetail(id, view)
		c := must(vm.NewStashComparison(d))
		got, ok := c.Stash()
		if !ok || got != d || got.Fields().View != view {
			t.Fatal("stash view lost exact comparison")
		}
		f := got.Fields()
		f.Parents[0] = domain.OID{}
		if !got.Valid() || got.Fields().Parents[0] != oid {
			t.Fatal("stash parent copy escaped")
		}
	}
	d := stashDetail(id, vm.StashParentView).Fields()
	d.ParentIndex = vm.Some(uint32(2))
	bad, e := vm.NewStashComparisonDetail(d)
	reject(t, bad, e)
	d = stashDetail(id, vm.StashIndexToWorktreeView).Fields()
	d.ParentIndex = vm.Some(uint32(0))
	bad, e = vm.NewStashComparisonDetail(d)
	reject(t, bad, e)
	d = stashDetail(id, vm.StashWorktreeView).Fields()
	d.ToTree = vm.None[domain.OID]()
	bad, e = vm.NewStashComparisonDetail(d)
	reject(t, bad, e)
	d = stashDetail(id, vm.StashWorktreeView).Fields()
	d.Parents[0] = revision(r, 40).OID()
	bad, e = vm.NewStashComparisonDetail(d)
	reject(t, bad, e)
	d = stashDetail(id, vm.StashUntrackedView).Fields()
	d.Parents = append(d.Parents, oid)
	d.FromTree = vm.Some(oid)
	d.ToTree = vm.Some(oid)
	if !must(vm.NewStashComparisonDetail(d)).Valid() {
		t.Fatal("present untracked tree rejected")
	}
	// Repeated stash objects remain observable with distinct positional labels.
	a := must(vm.NewStashRow(vm.StashRowSpec{ID: id, PositionLabel: "stash@{0}", Meta: meta()}))
	b := a.Fields()
	b.PositionLabel = "stash@{1}"
	rows := []vm.StashRow{a, must(vm.NewStashRow(b))}
	if !must(vm.NewStashesModel(vm.StashesModelSpec{Repository: r, Rows: rows, List: emptyList()})).Valid() {
		t.Fatal("duplicate object observation hidden")
	}

	original := time.Date(2026, 9, 6, 12, 34, 56, 123, time.FixedZone("source", 5*60*60+30*60))
	stamp := must(vm.NewTimestamp(original, 5*60*60+30*60))
	if !stamp.Valid() || !stamp.Time().Equal(original) || stamp.OriginalOffsetSeconds() != 19800 {
		t.Fatal("source instant/original offset lost")
	}
	s := commit(r, 64).Fields()
	s.AuthorEmail = "exact@example.test"
	s.AuthoredAt = vm.Some(stamp)
	s.CommittedAt = vm.Some(stamp)
	c := must(vm.NewCommitRow(s))
	gotTime, _ := c.Fields().AuthoredAt.Value()
	if !gotTime.Time().Equal(original) || gotTime.OriginalOffsetSeconds() != 19800 || c.Fields().AuthorEmail != s.AuthorEmail {
		t.Fatal("commit source facts lost")
	}
	s.AuthoredAt = vm.Some(vm.Timestamp{})
	badCommit, err := vm.NewCommitRow(s)
	reject(t, badCommit, err)
	badTime, err := vm.NewTimestamp(time.Time{}, 0)
	reject(t, badTime, err)
	badTime, err = vm.NewTimestamp(original, 24*60*60)
	reject(t, badTime, err)
	if (vm.StashComparisonDetail{}).Valid() || (vm.Timestamp{}).Valid() {
		t.Fatal("invalid zero source detail accepted")
	}
}
