package git

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func (s *readSession) tree(repo repository, commit domain.OID) (domain.OID, error) {
	if _, err := s.verifyCommit(repo, commit.String()); err != nil {
		return domain.OID{}, err
	}
	q := s.command(repo.common.path, "--git-dir="+repo.common.path, "rev-parse", "--verify", "--end-of-options", commit.String()+"^{tree}")
	if q.err != nil {
		return domain.OID{}, q.err
	}
	oid, err := domain.NewOID(line(q.stdout))
	if err != nil || oid.Format() != repo.format {
		return domain.OID{}, diagnostic(api.Unavailable, "InvalidTreeObject", "The native commit has no established tree object.")
	}
	return oid, nil
}

func (s *readSession) emptyTree(repo repository) (domain.OID, error) {
	// No -w: the native object-format encoder calculates this read-only.
	q := s.command(repo.common.path, "--git-dir="+repo.common.path, "hash-object", "-t", "tree", "--stdin")
	if q.err != nil {
		return domain.OID{}, q.err
	}
	return domain.NewOID(line(q.stdout))
}

func changeKind(status byte) (api.ChangeKind, error) {
	switch status {
	case 'A':
		return api.Added, nil
	case 'M':
		return api.Modified, nil
	case 'D':
		return api.Deleted, nil
	case 'R':
		return api.Renamed, nil
	case 'C':
		return api.Copied, nil
	case 'T':
		return api.TypeChanged, nil
	case 'U':
		return api.Unmerged, nil
	}
	return 0, diagnostic(api.Unavailable, "UnknownChangeKind", "Native Git returned an unclassified change kind.")
}

func parseNameStatus(raw []byte) ([]api.DiffFileFactData, error) {
	fields := bytes.Split(raw, []byte{0})
	if len(fields) > 0 && len(fields[len(fields)-1]) == 0 {
		fields = fields[:len(fields)-1]
	}
	var files []api.DiffFileFactData
	for i := 0; i < len(fields); {
		if len(fields[i]) == 0 || i+1 >= len(fields) {
			return nil, diagnostic(api.Unavailable, "MalformedDiffNames", "Native diff names have an invalid record shape.")
		}
		kind, err := changeKind(fields[i][0])
		if err != nil {
			return nil, err
		}
		i++
		d := api.DiffFileFactData{Kind: kind}
		if kind == api.Renamed || kind == api.Copied {
			if i+1 >= len(fields) {
				return nil, diagnostic(api.Unavailable, "MalformedDiffRename", "Native rename lacks an old/new path pair.")
			}
			old, err := api.NewGitPath(string(fields[i]))
			if err != nil {
				return nil, err
			}
			d.OldPath = api.Some(old)
			i++
		}
		path, err := api.NewGitPath(string(fields[i]))
		if err != nil {
			return nil, err
		}
		d.Path = path
		i++
		files = append(files, d)
	}
	return files, nil
}

func applyNumstat(raw []byte, files []api.DiffFileFactData) error {
	byPath := make(map[api.GitPath]int, len(files))
	for i, file := range files {
		byPath[file.Path] = i
	}
	fields := bytes.Split(raw, []byte{0})
	if len(fields) > 0 && len(fields[len(fields)-1]) == 0 {
		fields = fields[:len(fields)-1]
	}
	seen := make(map[api.GitPath]bool)
	for i := 0; i < len(fields); i++ {
		values := bytes.SplitN(fields[i], []byte{'\t'}, 3)
		if len(values) != 3 {
			return diagnostic(api.Unavailable, "MalformedDiffStats", "Native diff statistics have an invalid record shape.")
		}
		pathText := string(values[2])
		var old string
		if pathText == "" {
			if i+2 >= len(fields) {
				return diagnostic(api.Unavailable, "MalformedDiffStats", "Native renamed statistics lack exact paths.")
			}
			old = string(fields[i+1])
			pathText = string(fields[i+2])
			i += 2
		}
		path, err := api.NewGitPath(pathText)
		if err != nil {
			return err
		}
		index, p := byPath[path]
		if !p || seen[path] {
			return diagnostic(api.StaleObservation, "DiffSourcesChanged", "Native diff name/stat sets disagree.")
		}
		seen[path] = true
		if oldPath, p := files[index].OldPath.Value(); p && oldPath.String() != old {
			return diagnostic(api.StaleObservation, "DiffSourcesChanged", "Native diff rename endpoints disagree.")
		}
		if string(values[0]) == "-" && string(values[1]) == "-" {
			files[index].Binary = true
			continue
		}
		added, err := strconv.ParseUint(string(values[0]), 10, 64)
		if err != nil {
			return diagnostic(api.Unavailable, "MalformedDiffStats", "Native added-line count is invalid.")
		}
		deleted, err := strconv.ParseUint(string(values[1]), 10, 64)
		if err != nil {
			return diagnostic(api.Unavailable, "MalformedDiffStats", "Native deleted-line count is invalid.")
		}
		files[index].AddedLines = api.Some(added)
		files[index].DeletedLines = api.Some(deleted)
	}
	if len(seen) != len(files) {
		return diagnostic(api.StaleObservation, "DiffSourcesChanged", "Native diff name/stat sets disagree.")
	}
	return nil
}

func (s *readSession) treePatch(repo repository, from, to domain.OID, paths []api.GitPath, limits api.PatchLimits) (api.PatchFacts, error) {
	if !limits.Valid() {
		return api.PatchFacts{}, diagnostic(api.Invalid, "InvalidPatchLimits", "Positive bounded patch limits are required.")
	}
	env := []string{"GIT_ATTR_SOURCE=" + to.String()}
	run := func(args ...string) commandResult { return s.commandEnv(repo.common.path, env, args...) }
	base := []string{"--git-dir=" + repo.common.path, "diff", "--no-ext-diff", "--no-textconv", "--no-color", "--find-renames"}
	return nativePatch(run, base, []string{from.String(), to.String()}, paths, limits)
}

func nativePatch(run func(...string) commandResult, base, endpoints []string, paths []api.GitPath, limits api.PatchLimits) (api.PatchFacts, error) {
	comparison := append(append([]string(nil), endpoints...), "--")
	for _, path := range paths {
		comparison = append(comparison, path.String())
	}
	names := run(append(append(append([]string(nil), base...), "--name-status", "-z"), comparison...)...)
	if names.err != nil {
		return api.PatchFacts{}, names.err
	}
	files, err := parseNameStatus(names.stdout)
	if err != nil {
		return api.PatchFacts{}, err
	}
	stats := run(append(append(append([]string(nil), base...), "--numstat", "-z"), comparison...)...)
	if stats.err != nil {
		return api.PatchFacts{}, stats.err
	}
	if err = applyNumstat(stats.stdout, files); err != nil {
		return api.PatchFacts{}, err
	}
	d := api.PatchFactsData{}
	selected := files
	maxFiles := limits.Data().MaxFiles
	if maxFiles > 1000 {
		maxFiles = 1000
	}
	if uint64(len(selected)) > uint64(maxFiles) {
		selected = selected[:maxFiles]
		d.Truncated = true
	}
	// Bound patch file scope as well as metadata scope. Rename sources remain
	// literal operands and are included with their destinations.
	patchComparison := append(append([]string(nil), endpoints...), "--")
	for _, file := range selected {
		patchComparison = append(patchComparison, file.Path.String())
		if old, p := file.OldPath.Value(); p {
			patchComparison = append(patchComparison, old.String())
		}
	}
	if len(selected) > 0 {
		patch := run(append(append(append([]string(nil), base...), "--patch", "--binary"), patchComparison...)...)
		if patch.err != nil && !patch.transport.Data().StdoutTruncated {
			return api.PatchFacts{}, patch.err
		}
		d.Bytes = patch.stdout
		d.Truncated = d.Truncated || patch.transport.Data().StdoutTruncated
		if !patch.transport.Data().StdoutTruncated && !d.Truncated {
			d.OriginalBytes = api.Some(uint64(len(d.Bytes)))
		}
		if uint64(len(d.Bytes)) > limits.Data().MaxBytes {
			d.Bytes = d.Bytes[:int(limits.Data().MaxBytes)]
			d.Truncated = true
		}
		err = patch.err
	} else {
		d.Bytes = []byte{}
		d.OriginalBytes = api.Some(uint64(0))
	}
	for _, file := range selected {
		fact, fe := api.NewDiffFileFact(file)
		if fe != nil {
			return api.PatchFacts{}, fe
		}
		d.Files = append(d.Files, fact)
	}
	d.ReturnedBytes = uint64(len(d.Bytes))
	value, ve := api.NewPatchFacts(d)
	if ve != nil {
		return value, ve
	}
	return value, err
}

func (a *Adapter) ReadStashPatch(ctx context.Context, request api.ReadStashPatchRequest) (api.ReadStashPatchResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.ReadStashPatchResult{}, diagnostic(api.Invalid, "InvalidRequest", "The stash patch request is invalid.")
	}
	rd := request.Data()
	d := api.ReadStashPatchResultData{Stash: rd.Stash}
	repo, err := a.registered(s.ctx, rd.Stash.Repository())
	var parents []domain.OID
	if err == nil {
		parents, err = s.stashParents(repo, rd.Stash.OID())
		d.Parents = parents
	}
	var from, to domain.OID
	var absent bool
	if err == nil {
		comparison := api.StashPatchComparisonData{Stash: rd.Stash.OID(), Base: api.Some(parents[0]), IndexParent: api.Some(parents[1]), View: rd.View}
		if len(parents) == 3 {
			comparison.UntrackedParent = api.Some(parents[2])
		}
		var fromCommit, toCommit domain.OID
		toCommit = rd.Stash.OID()
		switch view := rd.View.(type) {
		case api.StashBaseToWorktree:
			fromCommit = parents[0]
		case api.StashBaseToIndex:
			fromCommit = parents[0]
			toCommit = parents[1]
		case api.StashIndexToWorktree:
			fromCommit = parents[1]
		case api.StashUntracked:
			if len(parents) < 3 {
				absent = true
			} else {
				toCommit = parents[2]
			}
		case api.StashParent:
			if uint64(view.Data().Index) >= uint64(len(parents)) {
				err = diagnostic(api.Invalid, "StashParentMissing", "The selected stash parent does not exist.")
			} else {
				fromCommit = parents[view.Data().Index]
			}
		}
		if err == nil && !absent {
			if fromCommit.Valid() {
				from, err = s.tree(repo, fromCommit)
			} else {
				from, err = s.emptyTree(repo)
			}
			if err == nil {
				to, err = s.tree(repo, toCommit)
			}
		}
		if err == nil {
			if !absent {
				comparison.FromTree = api.Some(from)
				comparison.ToTree = api.Some(to)
			}
			value, ce := api.NewStashPatchComparison(comparison)
			err = ce
			if ce == nil {
				d.Comparison = api.Some(value)
			}
		}
	}
	if err == nil {
		var patch api.PatchFacts
		if absent {
			patch, err = api.NewPatchFacts(api.PatchFactsData{Bytes: []byte{}, OriginalBytes: api.Some(uint64(0))})
		} else {
			patch, err = s.treePatch(repo, from, to, rd.Paths, rd.Limits)
		}
		if patch.Valid() {
			d.Patch = api.Some(patch)
		}
	}
	complete := api.Complete
	if err != nil {
		complete = api.Partial
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	version := sourceVersion("stash-patch", rd.Stash.Repository().Token(), a.lifetime, []byte(queryBinding(rd.Stash.OID().String(), from.String(), to.String(), strings.Join(pathStrings(rd.Paths), "\x00"))))
	observation, oe := s.observation(rd.Stash.Repository(), api.None[domain.WorktreeID](), version, complete)
	if oe != nil {
		return api.ReadStashPatchResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Transport = transportValue(s.transport)
	result, ve := api.NewReadStashPatchResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func pathStrings(paths []api.GitPath) []string {
	result := make([]string, len(paths))
	for i, path := range paths {
		result[i] = path.String()
	}
	return result
}
