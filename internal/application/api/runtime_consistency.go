package api

func sameCwdSubject(a, b CwdObservation) bool {
	x, y := a.data, b.data
	return x.Worktree.data.ID == y.Worktree.data.ID && x.Worktree.data.RootIdentity == y.Worktree.data.RootIdentity && x.ProjectIdentity == y.ProjectIdentity && sameStrings(x.ProjectComponents, y.ProjectComponents)
}
func sameRestartSummary(a, b InvocationSummary) bool {
	x, y := a.data, b.data
	return x.Label == y.Label && x.ExecutableDisplay == y.ExecutableDisplay && sameStrings(x.ArgumentDisplay, y.ArgumentDisplay) && x.Terminal == y.Terminal && sameCwdSubject(x.Cwd, y.Cwd)
}
func consistentSessionStartResult(d SessionStartResultData) error {
	s, p := d.Session.Value()
	if !p {
		for _, f := range d.Effects.data.Facets {
			if f.data.State != NotStarted && f.data.State != VerifiedNoTargetChange {
				return invalid("unadmitted start effect")
			}
		}
	} else if s.data.Phase == Running && !d.Established {
		return invalid("running without establishment")
	}
	return nil
}
func consistentSessionRestartResult(d SessionRestartResultData) error {
	if replacement, p := d.Replacement.Value(); p {
		if !replacement.data.Session.Present() {
			return invalid("replacement without registry admission")
		}
		if s, p := replacement.data.Session.Value(); p {
			old := d.Old.data.Session
			if s.data.SessionID.Value() <= old.data.SessionID.Value() {
				return invalid("replacement identity is not newer")
			}
			if s.data.WorktreeID != old.data.WorktreeID || !sameRestartSummary(old.data.Display, s.data.Display) {
				return invalid("restart original specification")
			}
			if a, p := old.data.AcquiredCwd.Value(); p {
				if b, p := s.data.AcquiredCwd.Value(); p && !sameCwdSubject(a.data.Observation, b.data.Observation) {
					return invalid("restart acquired cwd identity")
				}
			}
		}
	}
	return nil
}
