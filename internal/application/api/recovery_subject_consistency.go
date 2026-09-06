package api

// Missing optional evidence asserts nothing. Compatible observations can add a
// known fact without rewriting either original artifact record or its identity.
func mergeSubjectOption[T comparable](a, b Optional[T]) (Optional[T], error) {
	x, xp := a.Value()
	y, yp := b.Value()
	if xp && yp && x != y {
		return Optional[T]{}, invalid("conflicting recovery subject evidence")
	}
	if xp {
		return a, nil
	}
	return b, nil
}
func mergeRecoverySubjects(a, b RecoverySubject) (RecoverySubject, error) {
	x, y := a.data, b.data
	var err error
	if x.Repository, err = mergeSubjectOption(x.Repository, y.Repository); err != nil {
		return RecoverySubject{}, err
	}
	if x.Worktree, err = mergeSubjectOption(x.Worktree, y.Worktree); err != nil {
		return RecoverySubject{}, err
	}
	if x.Stash, err = mergeSubjectOption(x.Stash, y.Stash); err != nil {
		return RecoverySubject{}, err
	}
	if x.Revision, err = mergeSubjectOption(x.Revision, y.Revision); err != nil {
		return RecoverySubject{}, err
	}
	if x.Branch, err = mergeSubjectOption(x.Branch, y.Branch); err != nil {
		return RecoverySubject{}, err
	}
	if x.Session, err = mergeSubjectOption(x.Session, y.Session); err != nil {
		return RecoverySubject{}, err
	}
	if x.Store, err = mergeSubjectOption(x.Store, y.Store); err != nil {
		return RecoverySubject{}, err
	}
	if x.Family, err = mergeSubjectOption(x.Family, y.Family); err != nil {
		return RecoverySubject{}, err
	}
	return NewRecoverySubject(x)
}
