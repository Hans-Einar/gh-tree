package viewmodel

import "github.com/Hans-Einar/gh-tree/internal/domain"

// StashComparisonDetail keeps commit-parent OIDs separate from actual tree OIDs.
// The observed ordinary stash has base/index and optional untracked parents;
// malformed structures remain a source diagnostic, never fabricated detail.
type StashComparisonDetailSpec struct {
	Stash       domain.StashID
	View        StashView
	ParentIndex Optional[uint32]
	Parents     []domain.OID
	FromTree    Optional[domain.OID]
	ToTree      Optional[domain.OID]
}
type StashComparisonDetail struct {
	stash            domain.StashID
	view             StashView
	parentIndex      Optional[uint32]
	parents          [3]domain.OID
	count            uint8
	fromTree, toTree Optional[domain.OID]
}

func NewStashComparisonDetail(s StashComparisonDetailSpec) (StashComparisonDetail, error) {
	if len(s.Parents) < 2 || len(s.Parents) > 3 {
		return StashComparisonDetail{}, invalid("stash comparison parents")
	}
	d := StashComparisonDetail{stash: s.Stash, view: s.View, parentIndex: s.ParentIndex, count: uint8(len(s.Parents)), fromTree: s.FromTree, toTree: s.ToTree}
	copy(d.parents[:], s.Parents)
	if !d.Valid() {
		return StashComparisonDetail{}, invalid("stash comparison detail")
	}
	return d, nil
}
func (d StashComparisonDetail) Valid() bool {
	if !d.stash.Valid() || !d.view.Valid() || d.count < 2 || d.count > 3 || (d.view == StashParentView) != d.parentIndex.present || d.parentIndex.present && d.parentIndex.value >= uint32(d.count) || !optionalValid(d.fromTree) || !optionalValid(d.toTree) {
		return false
	}
	for i, p := range d.parents {
		if i < int(d.count) {
			if !p.Valid() || p.Format() != d.stash.OID().Format() {
				return false
			}
		} else if p != (domain.OID{}) {
			return false
		}
	}
	if d.fromTree.present && d.fromTree.value.Format() != d.stash.OID().Format() || d.toTree.present && d.toTree.value.Format() != d.stash.OID().Format() {
		return false
	}
	if d.view == StashUntrackedView && d.count == 2 {
		return !d.fromTree.present && !d.toTree.present
	}
	return d.fromTree.present && d.toTree.present
}
func (d StashComparisonDetail) Fields() StashComparisonDetailSpec {
	var parents []domain.OID
	if d.count <= 3 {
		parents = copySlice(d.parents[:d.count])
	}
	return StashComparisonDetailSpec{Stash: d.stash, View: d.view, ParentIndex: d.parentIndex, Parents: parents, FromTree: d.fromTree, ToTree: d.toTree}
}
