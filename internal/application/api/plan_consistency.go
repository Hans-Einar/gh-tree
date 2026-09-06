package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func consistentMutationPlanSummary(d MutationPlanSummaryData) error {
	w, wp := d.Worktree.Value()
	ew, ewp := d.Expected.data.Worktree.Value()
	h, hp := d.Head.Value()
	eh, ehp := d.Expected.data.Head.Value()
	target, tp := d.Target.Value()
	if wp && (!ewp || w != ew) || hp && (!ehp || h != eh) || wp && hp && !h.MatchesWorktree(w) {
		return invalid("summary expected identities")
	}
	if tp && !localTargetScope(target, d.Repository) {
		return invalid("summary target scope")
	}
	seen := map[Choice]bool{}
	for _, c := range d.Choices {
		if seen[c] {
			return invalid("duplicate summary choice")
		}
		seen[c] = true
	}
	requireWorktree := true
	full := false
	switch d.Kind {
	case CreateMutation:
		requireWorktree = false
		path, p := d.Destination.Value()
		mode, mp := d.CreateMode.Value()
		if !tp || !p || !nonempty(path) || !mp || !createModeScope(mode, d.Repository) {
			return invalid("create summary intent")
		}
	case RetargetMutation:
		full = true
		mode, p := d.RetargetMode.Value()
		if !tp || !p || !retargetModeScope(mode, d.Repository, None[domain.Revision]()) {
			return invalid("retarget summary target/mode")
		}
		if ff, p := mode.(FastForward); p {
			from, p := eh.Revision()
			if !p || ff.data.From != from || ff.data.To.OID() != target.ExpectedRevision().OID() {
				return invalid("summary fast-forward endpoints")
			}
		}
	case StageMutation:
		full = true
		if !d.StageAction.Present() {
			return invalid("stage summary action")
		}
	case CommitMutation:
		full = true
		message, p := d.Message.Value()
		if !p || !nonempty(message) || !d.CommitIndexPolicy.Present() {
			return invalid("commit summary message/policy")
		}
	case RestoreMutation:
		full = true
		if len(d.Paths) == 0 {
			return invalid("restore summary paths")
		}
	case StashMutation:
		full = true
		intent, p := d.StashIntent.Value()
		if !p {
			return invalid("stash summary intent")
		}
		request, err := NewPrepareStashRequest(PrepareStashRequestData{Expected: d.Expected, Intent: intent})
		if err != nil || !request.Valid() {
			return invalid("stash summary expected")
		}
		var stash domain.StashID
		switch x := intent.(type) {
		case CreateStashIntent:
			if d.Stash.Present() {
				return invalid("uncreated stash identity")
			}
			if !wp || w != x.data.Worktree {
				return invalid("stash capture worktree")
			}
		case ApplyStashIntent:
			stash = x.data.Stash
			if !wp || w != x.data.Worktree {
				return invalid("stash apply worktree")
			}
		case PopStashIntent:
			stash = x.data.Stash
			if !wp || w != x.data.Worktree {
				return invalid("stash pop worktree")
			}
		case DropStashIntent:
			requireWorktree = false
			stash = x.data.Stash
		}
		if stash.Valid() {
			s, p := d.Stash.Value()
			if !p || s != stash {
				return invalid("summary stash identity")
			}
		}
	case BranchMutation:
		b, p := d.Branch.Value()
		if !tp || !p || b.Kind() != domain.Local || b.Repository() != d.Repository {
			return invalid("branch summary intent")
		}
	case PushMutation:
		b, p := d.Branch.Value()
		binding, bp := d.PushBinding.Value()
		if !tp || !p || !bp || b.Kind() != domain.RemoteHead || binding.data.LocalRepository != d.Repository || binding.data.RemoteRepository != b.Repository() || target.ExpectedRevision().Repository() != d.Repository {
			return invalid("push summary exact endpoints")
		}
	}
	if requireWorktree && (!wp || !hp || !expectedWorktree(d.Expected, w, full)) {
		return invalid("summary material worktree/preconditions")
	}
	if d.Kind != CreateMutation && (d.Destination.Present() || d.CreateMode.Present()) || d.Kind != RetargetMutation && d.RetargetMode.Present() || d.Kind != StashMutation && d.StashIntent.Present() || d.Kind != StageMutation && d.StageAction.Present() || d.Kind != CommitMutation && (d.Message.Present() || d.CommitIndexPolicy.Present()) || d.Kind != PushMutation && d.PushBinding.Present() || d.Kind != PushMutation && d.Kind != BranchMutation && d.Branch.Present() {
		return invalid("summary foreign operation payload")
	}
	return nil
}
