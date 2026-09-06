package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func sameRemoteBinding(a, b RemoteBinding) bool {
	x, y := a.data, b.data
	if x.LocalRepository != y.LocalRepository || x.RemoteRepository != y.RemoteRepository || x.RemoteName != y.RemoteName || x.Configuration != y.Configuration || !sameStrings(x.FetchURLs, y.FetchURLs) || !sameStrings(x.PushURLs, y.PushURLs) || len(x.Refspecs) != len(y.Refspecs) {
		return false
	}
	for i := range x.Refspecs {
		if x.Refspecs[i].data != y.Refspecs[i].data {
			return false
		}
	}
	return true
}

// ValidatePushBinding checks only supplied context. The receiver remains intact
// on error, including independently valid uncertainty, effects and recovery.
// An unbound standalone partial result does not certify a remote association.
func (v GitMutationResult) ValidatePushBinding(binding RemoteBinding, destination domain.BranchID) error {
	if !v.Valid() || v.data.Kind != PushMutation || !binding.Valid() || !destination.Valid() || destination.Kind() != domain.RemoteHead || destination.Repository() != binding.data.RemoteRepository {
		return invalid("push result validation context")
	}
	repo, worktree, observedRemotes := mutationRecoveryAssociations(v.data)
	if repo.Valid() && repo != binding.data.LocalRepository {
		return invalid("push result local repository context")
	}
	for _, r := range observedRemotes {
		if r != binding.data.RemoteRepository {
			return invalid("push result remote context")
		}
	}
	if p, ok := v.data.Outcome.(Pushed); ok {
		if !sameRemoteBinding(p.data.Binding, binding) || p.data.Destination != destination {
			return invalid("push result destination context")
		}
	}
	subjects := make([]RecoverySubject, 0, len(v.data.Recovery)+len(v.data.Steps))
	for _, r := range v.data.Recovery {
		subjects = append(subjects, r.data.Subject)
	}
	for _, s := range v.data.Steps {
		subjects = append(subjects, s.data.Target)
	}
	if err := gitSubjects(subjects, binding.data.LocalRepository, worktree, []domain.RepositoryID{binding.data.RemoteRepository}, false); err != nil {
		return err
	}
	for _, s := range subjects {
		if b, p := s.data.Branch.Value(); p && b.Kind() == domain.RemoteHead && b != destination {
			return invalid("push retained remote branch context")
		}
	}
	return nil
}
