package git

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func (a *Adapter) ReadGraph(ctx context.Context, request api.ReadGraphRequest) (api.ReadGraphResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.ReadGraphResult{}, diagnostic(api.Invalid, "InvalidRequest", "The graph request is invalid.")
	}
	rd := request.Data()
	d := api.ReadGraphResultData{}
	repo, err := a.registered(s.ctx, rd.Repository)
	var refs refsSnapshot
	var inventory inventoryResult
	if err == nil {
		refs, err = s.refs(repo)
		d.Refs = refs.facts
		d.Diagnostics = append(d.Diagnostics, refs.diagnostics...)
	}
	if err == nil {
		inventory, err = s.inventory(repo)
		d.Heads = inventory.facts
		d.Diagnostics = append(d.Diagnostics, inventory.diagnostics...)
	}
	var paths []api.GitPath
	filter := "all"
	switch v := rd.Filter.(type) {
	case api.AllGraph:
		paths = v.Data().Paths
	case api.ReachableFromRoots:
		paths = v.Data().Paths
		filter = "reachable"
	}
	parts := []string{"graph", rd.Repository.Token(), filter, refs.fingerprint, fmt.Sprint(inventory.observation.Data().Version)}
	contextSource := ""
	if err == nil {
		contextSource, err = historyContext(repo)
	}
	parts = append(parts, contextSource)
	var roots []string
	for _, revision := range rd.Roots {
		if err == nil {
			_, err = s.verifyCommit(repo, revision.OID().String())
		}
		roots = append(roots, revision.OID().String())
		parts = append(parts, revision.OID().String())
	}
	for _, path := range paths {
		parts = append(parts, path.String())
	}
	binding := queryBinding(parts...)
	version := sourceVersion("graph", rd.Repository.Token(), a.lifetime, []byte(binding))
	var offset uint64
	more := false
	complete := api.Complete
	if err == nil {
		offset, err = a.pageOffset(rd.Page, binding)
	}
	if err == nil {
		args := []string{"--git-dir=" + repo.common.path, "rev-list", "--topo-order", "--max-count=" + strconv.FormatUint(uint64(rd.Page.Data().Limit)+1, 10), "--skip=" + strconv.FormatUint(offset, 10)}
		args = append(args, roots...)
		args = append(args, "--")
		for _, path := range paths {
			args = append(args, path.String())
		}
		q := s.command(repo.common.path, args...)
		err = q.err
		if err == nil {
			rows := bytes.Split(bytes.TrimSuffix(q.stdout, []byte{'\n'}), []byte{'\n'})
			if len(rows) == 1 && len(rows[0]) == 0 {
				rows = nil
			}
			if len(rows) > int(rd.Page.Data().Limit) {
				more = true
				complete = api.More
				rows = rows[:rd.Page.Data().Limit]
			}
			for _, row := range rows {
				commit, ce := s.commit(repo, string(row))
				if ce != nil {
					err = ce
					break
				}
				d.Commits = append(d.Commits, commit)
			}
		}
	}
	if err != nil {
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	if err != nil || len(d.Diagnostics) > 0 || inventory.observation.Valid() && inventory.observation.Data().Completeness != api.Complete {
		complete = api.Partial
		more = false
	}
	observation, oe := s.observation(rd.Repository, api.None[domain.WorktreeID](), version, complete)
	if oe != nil {
		return api.ReadGraphResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Page, oe = a.pageInfo(version, binding, offset, len(d.Commits), more, complete)
	if oe != nil {
		return api.ReadGraphResult{}, oe
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewReadGraphResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}
