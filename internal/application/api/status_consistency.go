package api

func consistentChangeFact(d ChangeFactData) error {
	switch d.Cause {
	case IndexChangeCause, WorktreeChangeCause:
		if d.Kind < Added || d.Kind > TypeChanged {
			return invalid("ordinary change kind")
		}
	case UntrackedChangeCause:
		if d.Kind != Untracked || len(d.IndexEntries) != 0 {
			return invalid("untracked change facts")
		}
	case ConflictChangeCause:
		if d.Kind != Unmerged || len(d.IndexEntries) == 0 {
			return invalid("conflict change facts")
		}
	default:
		return invalid("change cause")
	}
	old, hasOld := d.OldPath.Value()
	if hasOld != (d.Kind == Renamed || d.Kind == Copied) || (hasOld && old == d.Path) {
		return invalid("change old path")
	}
	var stages [4]bool
	for _, entry := range d.IndexEntries {
		stage := entry.data.Stage // IndexEntryFact already validates the range.
		if stages[stage] || (d.Cause == ConflictChangeCause) != (stage != 0) {
			return invalid("change index stages")
		}
		stages[stage] = true
	}
	return nil
}

func consistentStatusChanges(changes []ChangeFact) error {
	type pathFacts struct {
		first  ChangeFactData
		causes [5]bool
	}
	paths := make(map[GitPath]pathFacts, len(changes))
	for _, change := range changes {
		d := change.data // ChangeFact admission validates the cause and current facts.
		seen, exists := paths[d.Path]
		if exists {
			if seen.causes[d.Cause] {
				return invalid("duplicate status path and cause")
			}
			if d.Cause == ConflictChangeCause || seen.causes[ConflictChangeCause] {
				return invalid("conflict status coexistence")
			}
			if !sameStatusCurrentFacts(seen.first, d) {
				return invalid("contradictory status current facts")
			}
		} else {
			seen.first = d
		}
		seen.causes[d.Cause] = true
		paths[d.Path] = seen
	}
	return nil
}

// Compare observed facts only. Kind and OldPath belong to each separate cause;
// neither row order nor opaque SourceVersion bytes encode that cause.
func sameStatusCurrentFacts(a, b ChangeFactData) bool {
	return sameFileState(a.WorktreeState, b.WorktreeState) && sameIndexEntries(a.IndexEntries, b.IndexEntries)
}

func sameFileState(a, b FileState) bool {
	switch x := a.(type) {
	case AbsentFile:
		y, ok := b.(AbsentFile)
		return ok && x.data == y.data
	case PresentFile:
		y, ok := b.(PresentFile)
		return ok && x.data == y.data
	}
	return false
}

// Callers admit unique stages first. Semantic flags form an unordered set;
// the existing IndexEntryFact permits repeated flags, which do not add a fact.
func sameIndexEntries(a, b []IndexEntryFact) bool {
	if len(a) != len(b) {
		return false
	}
	var byStage [4]*IndexEntryFactData
	for i := range a {
		byStage[a[i].data.Stage] = &a[i].data
	}
	for _, entry := range b {
		y := entry.data
		x := byStage[y.Stage]
		if x == nil || x.Path != y.Path || x.Object != y.Object || x.Mode != y.Mode || indexFlagSet(x.SemanticFlags) != indexFlagSet(y.SemanticFlags) {
			return false
		}
	}
	return true
}

func indexFlagSet(flags []IndexFlag) uint8 {
	var set uint8
	for _, flag := range flags {
		set |= 1 << flag
	}
	return set
}
