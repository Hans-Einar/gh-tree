package git

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

var _ ports.GitFacts = (*Adapter)(nil)

func (a *Adapter) ReadDiff(ctx context.Context, request api.ReadDiffRequest) (api.ReadDiffResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.ReadDiffResult{}, diagnostic(api.Invalid, "InvalidRequest", "The diff request is invalid.")
	}
	rd := request.Data()
	d := api.ReadDiffResultData{Comparison: rd.Comparison}
	var repoID domain.RepositoryID
	var worktree api.Optional[domain.WorktreeID]
	switch c := rd.Comparison.(type) {
	case api.CommitPairComparison:
		repoID = c.Data().From.Repository()
	case api.CommitParentComparison:
		repoID = c.Data().Commit.Repository()
	case api.IndexToWorktreeComparison:
		repoID = c.Data().Worktree.Repository()
		worktree = api.Some(c.Data().Worktree)
	case api.HeadToIndexComparison:
		repoID = c.Data().Worktree.Repository()
		worktree = api.Some(c.Data().Worktree)
	}
	repo, err := a.registered(s.ctx, repoID)
	var patch api.PatchFacts
	var version api.SourceVersion
	if err == nil {
		switch c := rd.Comparison.(type) {
		case api.CommitPairComparison:
			var from, to domain.OID
			from, err = s.tree(repo, c.Data().From.OID())
			if err == nil {
				to, err = s.tree(repo, c.Data().To.OID())
			}
			if err == nil {
				patch, err = s.treePatch(repo, from, to, rd.Paths, rd.Limits)
			}
			version = sourceVersion("diff", repoID.Token(), a.lifetime, []byte(queryBinding(from.String(), to.String(), fmt.Sprint(rd.Paths))))
		case api.CommitParentComparison:
			var commit api.CommitFact
			commit, err = s.commit(repo, c.Data().Commit.OID().String())
			var from, to domain.OID
			if err == nil {
				to, err = s.tree(repo, c.Data().Commit.OID())
			}
			if err == nil {
				switch parent := c.Data().Parent.(type) {
				case api.RootParent:
					if len(commit.Data().Parents) != 0 {
						err = diagnostic(api.Invalid, "CommitIsNotRoot", "An explicit root comparison requires a parentless commit.")
					} else {
						from, err = s.emptyTree(repo)
					}
				case api.SelectedParent:
					if uint64(parent.Data().Index) >= uint64(len(commit.Data().Parents)) {
						err = diagnostic(api.Invalid, "CommitParentMissing", "The selected exact parent does not exist.")
					} else {
						from, err = s.tree(repo, commit.Data().Parents[parent.Data().Index].OID())
					}
				}
			}
			if err == nil {
				patch, err = s.treePatch(repo, from, to, rd.Paths, rd.Limits)
			}
			version = sourceVersion("diff", repoID.Token(), a.lifetime, []byte(queryBinding(from.String(), to.String(), fmt.Sprint(rd.Paths))))
		case api.IndexToWorktreeComparison:
			before, se := s.status(repo, c.Data().Worktree)
			err = se
			d.Diagnostics = append(d.Diagnostics, before.diagnostics...)
			if before.facts.Valid() {
				version = before.facts.Data().Observation.Data().Version
			}
			if err == nil {
				bd := before.facts.Data()
				if bd.IndexVersion != c.Data().Index || bd.WorktreeVersion != c.Data().WorktreeVersion {
					err = diagnostic(api.StaleObservation, "DiffSourceChanged", "The exact index/worktree versions changed before diff.")
				} else {
					scope, _ := bd.Worktree.Data().Scope.Value()
					root := scope.Data().RootLocator
					run := func(args ...string) commandResult { return s.command(root, args...) }
					base := []string{"--git-dir=" + before.admin, "--work-tree=" + root, "diff", "--no-ext-diff", "--no-textconv", "--no-color", "--find-renames"}
					patch, err = nativePatch(run, base, nil, rd.Paths, rd.Limits)
				}
			}
			if patch.Valid() {
				after, se := s.status(repo, c.Data().Worktree)
				d.Diagnostics = append(d.Diagnostics, after.diagnostics...)
				if se != nil || !after.facts.Valid() || after.facts.Data().IndexVersion != c.Data().Index || after.facts.Data().WorktreeVersion != c.Data().WorktreeVersion {
					patch = api.PatchFacts{}
					err = diagnostic(api.StaleObservation, "DiffSourceChanged", "The index/worktree source changed during diff acquisition.")
				}
				if after.facts.Valid() {
					version = after.facts.Data().Observation.Data().Version
				}
			}
		case api.HeadToIndexComparison:
			var result headIndexSnapshot
			result, err = s.headIndex(repo, c.Data().Worktree)
			version = result.version
			if err == nil && (result.head != c.Data().Head || result.version != c.Data().HeadVersion || result.indexVersion != c.Data().Index) {
				err = diagnostic(api.StaleObservation, "DiffSourceChanged", "The exact Head/index versions changed before staged diff.")
			}
			if err == nil {
				var baseOID domain.OID
				if revision, p := result.head.Revision(); p {
					baseOID = revision.OID()
				} else {
					baseOID, err = s.emptyTree(repo)
				}
				if err == nil {
					run := func(args ...string) commandResult { return s.command(result.root, args...) }
					base := []string{"--git-dir=" + result.admin, "--work-tree=" + result.root, "diff", "--cached", "--no-ext-diff", "--no-textconv", "--no-color", "--find-renames"}
					patch, err = nativePatch(run, base, []string{baseOID.String()}, rd.Paths, rd.Limits)
				}
			}
			if patch.Valid() {
				after, se := s.headIndex(repo, c.Data().Worktree)
				if se != nil || after.head != result.head || !bytes.Equal(after.index, result.index) {
					patch = api.PatchFacts{}
					err = diagnostic(api.StaleObservation, "DiffSourceChanged", "The Head/index source changed during staged diff.")
				}
				if after.version.Valid() {
					version = after.version
				}
			}
		}
	}
	if !version.Valid() {
		version = sourceVersion("diff", repoID.Token(), a.lifetime, []byte(fmt.Sprint(rd.Comparison)))
	}
	complete := api.Complete
	if err != nil {
		complete = api.Partial
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	if patch.Valid() {
		d.Patch = api.Some(patch)
		if patch.Data().Truncated && complete == api.Complete {
			complete = api.More
		}
	}
	observation, oe := s.observation(repoID, worktree, version, complete)
	if oe != nil {
		return api.ReadDiffResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Transport = transportValue(s.transport)
	result, ve := api.NewReadDiffResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

type headIndexSnapshot struct {
	head                  domain.Head
	version, indexVersion api.SourceVersion
	index                 []byte
	root, admin           string
}

func (s *readSession) headIndex(repo repository, id domain.WorktreeID) (headIndexSnapshot, error) {
	var result headIndexSnapshot
	inventory, err := s.inventory(repo)
	if err != nil {
		return result, err
	}
	for _, w := range inventory.facts {
		if w.Data().ID != id {
			continue
		}
		scope, p := w.Data().Scope.Value()
		head, h := w.Data().Head.Value()
		if !p || !h {
			return result, diagnostic(api.Unavailable, "WorktreeUnavailable", "Staged diff needs an established worktree root and Head.")
		}
		result.root = scope.Data().RootLocator
		result.head = head
		result.version = w.Data().Observation.Data().Version
		result.admin = repo.common.path
		if !w.Data().Primary {
			result.admin = filepath.Join(repo.common.path, "worktrees", strings.TrimPrefix(id.AdministrativeKey(), "linked:"))
		}
		result.index, err = indexBytes(result.admin)
		if err != nil {
			return result, err
		}
		result.indexVersion = sourceVersion("index", queryBinding(repo.id.Token(), id.AdministrativeKey()), s.a.lifetime, result.index)
		return result, nil
	}
	return result, diagnostic(api.NotFound, "WorktreeNotRegistered", "The staged diff worktree is not registered.")
}
