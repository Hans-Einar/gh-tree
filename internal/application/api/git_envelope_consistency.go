package api

import "github.com/Hans-Einar/gh-tree/internal/domain"

func observationSubject(o Optional[GitObservation], repo domain.RepositoryID, w Optional[domain.WorktreeID]) error {
	if v, p := o.Value(); p {
		if repo.Valid() && v.data.Repository != repo {
			return invalid("Git result observation repository")
		}
		if expected, p := w.Value(); p {
			if actual, p := v.data.Worktree.Value(); p && actual != expected {
				return invalid("Git result observation worktree")
			}
		}
	}
	return nil
}
func gitPageCount(page PageInfo, o Optional[GitObservation], count int) error {
	if uint64(page.data.Returned) != uint64(count) {
		return invalid("returned Git page count")
	}
	if obs, p := o.Value(); p && page.data.Source != obs.data.Version {
		return invalid("Git page source version")
	}
	return nil
}
func consistentResolveLocalResult(d ResolveLocalResultData) error {
	if r, p := d.Repository.Value(); p {
		return observationSubject(d.Observation, r.data.Repository, None[domain.WorktreeID]())
	}
	return nil
}
func consistentListWorktreesResult(d ListWorktreesResultData) error {
	var repo domain.RepositoryID
	for _, w := range d.Worktrees {
		if repo.Valid() && w.data.ID.Repository() != repo {
			return invalid("inventory cross-repository row")
		}
		repo = w.data.ID.Repository()
	}
	return observationSubject(d.Observation, repo, None[domain.WorktreeID]())
}
func consistentObserveStatusResult(d ObserveStatusResultData) error {
	if s, p := d.Status.Value(); p {
		return observationSubject(d.Observation, s.data.Worktree.data.ID.Repository(), Some(s.data.Worktree.data.ID))
	}
	return nil
}
func consistentResolveExactResult(d ResolveExactResultData) error {
	if r, p := d.Resolution.Value(); p {
		return observationSubject(d.Observation, r.data.Local.Repository(), None[domain.WorktreeID]())
	}
	return nil
}
func consistentListRefsResult(d ListRefsResultData) error {
	var repo domain.RepositoryID
	for _, r := range d.Refs {
		if repo.Valid() && repo != r.data.Observation.data.Repository {
			return invalid("ref page repositories")
		}
		repo = r.data.Observation.data.Repository
	}
	if err := observationSubject(d.Observation, repo, None[domain.WorktreeID]()); err != nil {
		return err
	}
	return gitPageCount(d.Page, d.Observation, len(d.Refs))
}
func consistentListStashesResult(d ListStashesResultData) error {
	var repo domain.RepositoryID
	for _, s := range d.Stashes {
		if repo.Valid() && repo != s.data.ID.Repository() {
			return invalid("stash page repositories")
		}
		repo = s.data.ID.Repository()
	}
	if err := observationSubject(d.Observation, repo, None[domain.WorktreeID]()); err != nil {
		return err
	}
	return gitPageCount(d.Page, d.Observation, len(d.Stashes))
}
func consistentReadGraphResult(d ReadGraphResultData) error {
	r, err := postRepository(GitPostFactsData{Worktrees: d.Heads, Refs: d.Refs})
	if err != nil {
		return err
	}
	for _, c := range d.Commits {
		if r.Valid() && c.data.Revision.Repository() != r {
			return invalid("graph read repositories")
		}
		r = c.data.Revision.Repository()
	}
	if err := observationSubject(d.Observation, r, None[domain.WorktreeID]()); err != nil {
		return err
	}
	return gitPageCount(d.Page, d.Observation, len(d.Commits))
}
func comparisonSubject(c GitComparison) (domain.RepositoryID, Optional[domain.WorktreeID]) {
	switch x := c.(type) {
	case CommitParentComparison:
		return x.data.Commit.Repository(), None[domain.WorktreeID]()
	case CommitPairComparison:
		return x.data.From.Repository(), None[domain.WorktreeID]()
	case IndexToWorktreeComparison:
		return x.data.Worktree.Repository(), Some(x.data.Worktree)
	case HeadToIndexComparison:
		return x.data.Worktree.Repository(), Some(x.data.Worktree)
	}
	return domain.RepositoryID{}, None[domain.WorktreeID]()
}
func consistentReadDiffResult(d ReadDiffResultData) error {
	r, w := comparisonSubject(d.Comparison)
	return observationSubject(d.Observation, r, w)
}
func consistentReadStashPatchResult(d ReadStashPatchResultData) error {
	if c, p := d.Comparison.Value(); p && c.data.Stash != d.Stash.OID() {
		return invalid("read stash exact object")
	}
	if d.Patch.Present() && !d.Comparison.Present() {
		return invalid("read stash comparison")
	}
	return observationSubject(d.Observation, d.Stash.Repository(), None[domain.WorktreeID]())
}
func consistentFetchResult(d FetchResultData) error {
	repo := d.Freshness.data.Binding.data.LocalRepository
	if g, p := d.Generation.Value(); p {
		if g.data.Repository != repo || g.data.Binding.data.Configuration != d.Freshness.data.Binding.data.Configuration {
			return invalid("fetch generation binding")
		}
	}
	for _, r := range d.Refs {
		if r.data.Observation.data.Repository != repo {
			return invalid("fetched ref repository")
		}
	}
	return observationSubject(d.Observation, repo, None[domain.WorktreeID]())
}
