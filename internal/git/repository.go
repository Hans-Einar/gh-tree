package git

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type readSession struct {
	a         *Adapter
	ctx       context.Context
	started   time.Time
	transport api.CommandTransportOutcomeData
}

func (a *Adapter) readSession(ctx context.Context) (*readSession, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	bounded, cancel := context.WithTimeout(ctx, a.options.ReadTimeout)
	return &readSession{a: a, ctx: bounded, started: time.Now().UTC(), transport: api.CommandTransportOutcomeData{CleanupKnown: true}}, cancel
}

func (s *readSession) command(cwd string, args ...string) commandResult {
	r := s.a.command(s.ctx, cwd, false, args...)
	d := r.transport.Data()
	if d.Started {
		s.transport.Started = true
		s.transport.RootReaped = d.RootReaped
		s.transport.ExitCode = d.ExitCode
	}
	s.transport.CleanupKnown = s.transport.CleanupKnown && d.CleanupKnown
	s.transport.StdoutTruncated = s.transport.StdoutTruncated || d.StdoutTruncated
	s.transport.StderrTruncated = s.transport.StderrTruncated || d.StderrTruncated
	s.transport.CancellationRequested = s.transport.CancellationRequested || d.CancellationRequested
	s.transport.Diagnostics = append(s.transport.Diagnostics, d.Diagnostics...)
	return r
}

func (s *readSession) observation(repo domain.RepositoryID, w api.Optional[domain.WorktreeID], version api.SourceVersion, complete api.Completeness) (api.GitObservation, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return api.GitObservation{}, err
	}
	id, err := api.NewObservationID(s.a.lifetime + ":" + hex.EncodeToString(nonce[:]))
	if err != nil {
		return api.GitObservation{}, err
	}
	return api.NewGitObservation(api.GitObservationData{ID: id, Repository: repo, Worktree: w, Interval: interval(s.started), Version: version, Completeness: complete})
}

func line(bytes []byte) string {
	return strings.TrimSuffix(strings.TrimSuffix(string(bytes), "\n"), "\r")
}

func directoryKey(d directoryObservation) string {
	return fmt.Sprintf("%s\x00%d\x00%x", d.path, d.identity.Device(), d.identity.FileID())
}

func (s *readSession) resolveRepository(locator string) (repository, error) {
	d, err := observeDirectory(locator)
	if err != nil {
		return repository{}, diagnostic(api.NotFound, "RepositoryLocatorUnavailable", "The selected directory cannot be inspected.")
	}
	c := s.command(d.path, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if c.err != nil {
		return repository{}, c.err
	}
	common, err := observeDirectory(line(c.stdout))
	if err != nil {
		return repository{}, diagnostic(api.Unavailable, "CommonDirectoryUnavailable", "The native common directory cannot be inspected.")
	}
	f := s.command(common.path, "--git-dir="+common.path, "rev-parse", "--show-object-format")
	if f.err != nil {
		return repository{}, f.err
	}
	var format domain.ObjectFormat
	switch line(f.stdout) {
	case "sha1":
		format = domain.SHA1
	case "sha256":
		format = domain.SHA256
	default:
		return repository{}, diagnostic(api.Unsupported, "UnsupportedObjectFormat", "Git reported an unsupported object format.")
	}
	v := s.command(common.path, "--version")
	if v.err != nil {
		return repository{}, v.err
	}
	backend := api.FilesRefs
	b := s.command(common.path, "--git-dir="+common.path, "config", "--local", "--get", "extensions.refStorage")
	if b.err == nil {
		switch line(b.stdout) {
		case "files":
		case "reftable":
			backend = api.ReftableRefs
		default:
			backend = api.OtherRefs
		}
	} else if exit, _ := b.transport.Data().ExitCode.Value(); exit != 1 {
		return repository{}, b.err
	}
	// Locator and observed volume/object identity both participate. Clones and
	// relocation cannot inherit the same common-repository scope.
	h := fmt.Sprintf("%s:%d:%x", common.path, common.identity.Device(), common.identity.FileID())
	id, err := domain.NewRepositoryID(domain.LocalCommon, h)
	if err != nil {
		return repository{}, err
	}
	r := repository{id: id, common: common, format: format, version: line(v.stdout), backend: backend}
	s.a.mu.Lock()
	defer s.a.mu.Unlock()
	if _, ok := s.a.repositories[id]; !ok && len(s.a.repositories) >= s.a.options.MaxRepositories {
		return repository{}, diagnostic(api.Busy, "RepositoryAdmissionFull", "The adapter repository registry is full.")
	}
	s.a.repositories[id] = r
	return r, nil
}

func (a *Adapter) ResolveLocal(ctx context.Context, request api.ResolveLocalRequest) (api.ResolveLocalResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	d := api.ResolveLocalResultData{}
	var err error
	if !request.Valid() {
		err = diagnostic(api.Invalid, "InvalidRequest", "The local resolution request is invalid.")
	} else {
		var repo repository
		repo, err = s.resolveRepository(request.Data().Locator)
		if err == nil {
			var inventory inventoryResult
			inventory, err = s.inventory(repo)
			d.Diagnostics = append(d.Diagnostics, inventory.diagnostics...)
			if inventory.observation.Valid() {
				d.Observation = api.Some(inventory.observation)
				caps, ce := capabilities(repo)
				if ce != nil {
					err = ce
				} else {
					remotes, diags, re := s.remoteBindings(repo)
					d.Diagnostics = append(d.Diagnostics, diags...)
					if re != nil && err == nil {
						err = re
					}
					if len(diags) > 0 || re != nil {
						observed := inventory.observation.Data()
						observed.Completeness = api.Partial
						observed.Interval = interval(s.started)
						inventory.observation, ce = api.NewGitObservation(observed)
						if ce != nil {
							return api.ResolveLocalResult{}, ce
						}
						d.Observation = api.Some(inventory.observation)
					}
					facts, fe := api.NewLocalRepositoryFacts(api.LocalRepositoryFactsData{Repository: repo.id, CommonDirectory: repo.common.path, Worktrees: inventory.facts, Remotes: remotes, Capabilities: caps, Observation: inventory.observation})
					if fe != nil {
						err = fe
					} else {
						d.Repository = api.Some(facts)
					}
				}
			}
		}
	}
	if err != nil {
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewResolveLocalResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func (a *Adapter) ListWorktrees(ctx context.Context, request api.ListWorktreesRequest) (api.ListWorktreesResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	d := api.ListWorktreesResultData{}
	var err error
	if !request.Valid() {
		err = diagnostic(api.Invalid, "InvalidRequest", "The worktree inventory request is invalid.")
	} else {
		var repo repository
		repo, err = a.registered(s.ctx, request.Data().Repository)
		if err == nil {
			var inventory inventoryResult
			inventory, err = s.inventory(repo)
			d.Worktrees = inventory.facts
			d.Diagnostics = inventory.diagnostics
			if inventory.observation.Valid() {
				d.Observation = api.Some(inventory.observation)
			}
		}
	}
	if err != nil {
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewListWorktreesResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func capabilities(r repository) (api.GitCapabilities, error) {
	var facts []api.GitCapabilityFact
	for _, op := range []string{"LocalRepositoryInventory"} {
		f, err := api.NewGitCapabilityFact(api.GitCapabilityFactData{Operation: op, Supported: true})
		if err != nil {
			return api.GitCapabilities{}, err
		}
		facts = append(facts, f)
	}
	// Until each mutation implementation is complete this partial checkpoint
	// makes no mutation support claim. It does not implement refusal stubs.
	return api.NewGitCapabilities(api.GitCapabilitiesData{ObjectFormat: r.format, RefBackend: r.backend, GitVersion: r.version, Profile: "native-read-m3-in-progress", Capabilities: facts})
}

type inventoryResult struct {
	facts       []api.WorktreeFacts
	observation api.GitObservation
	diagnostics []api.Diagnostic
}
type worktreeRecord struct {
	root, admin, key, branch, oid, locked, prunable string
	primary, bare, isLocked, isPrunable             bool
}

func parseInventory(raw []byte) ([]worktreeRecord, error) {
	if len(raw) == 0 {
		return nil, diagnostic(api.Unavailable, "EmptyInventory", "Native Git returned no inventory records.")
	}
	var records []worktreeRecord
	for _, record := range bytes.Split(raw, []byte{0, 0}) {
		if len(record) == 0 {
			continue
		}
		var r worktreeRecord
		for _, field := range bytes.Split(record, []byte{0}) {
			key, value, _ := strings.Cut(string(field), " ")
			switch key {
			case "worktree":
				if r.root != "" {
					return nil, diagnostic(api.Unavailable, "MalformedInventory", "A worktree record has duplicate roots.")
				}
				r.root = value
			case "HEAD":
				r.oid = value
			case "branch":
				r.branch = value
			case "bare":
				r.bare = true
			case "detached":
			case "locked":
				r.isLocked = true
				r.locked = value
			case "prunable":
				r.isPrunable = true
				r.prunable = value
			default:
				return nil, diagnostic(api.Unavailable, "UnknownInventoryField", "Native worktree inventory contains an unrecognized field.")
			}
		}
		if r.root == "" || len(records) >= 10000 {
			return nil, diagnostic(api.Unavailable, "MalformedInventory", "Native worktree inventory is malformed or exceeds its bound.")
		}
		r.primary = len(records) == 0
		records = append(records, r)
	}
	return records, nil
}

func (s *readSession) inventory(repo repository) (inventoryResult, error) {
	var result inventoryResult
	c := s.command(repo.common.path, "--git-dir="+repo.common.path, "worktree", "list", "--porcelain", "-z")
	if c.err != nil {
		return result, c.err
	}
	records, err := parseInventory(c.stdout)
	if err != nil {
		return result, err
	}
	adminByRoot, err := administrativeRoots(repo.common.path)
	if err != nil {
		return result, err
	}
	currentAdmin := ""
	if current, err := observeDirectory(s.a.current.path); err == nil && sameDirectoryObject(current, s.a.current) {
		q := s.command(current.path, "rev-parse", "--absolute-git-dir")
		if q.err == nil {
			if d, e := observeDirectory(line(q.stdout)); e == nil {
				currentAdmin = d.path
			}
		}
	}
	version := sourceVersion("inventory", repo.id.Token(), s.a.lifetime, c.stdout)
	complete := api.Complete
	var data []api.WorktreeFactsData
	for _, r := range records {
		if r.bare {
			continue
		} // Bare registration has no WorktreeID/root scope.
		if r.primary {
			r.admin = repo.common.path
			r.key = "primary"
		} else {
			for path, admin := range adminByRoot {
				if filepath.Clean(path) == filepath.Clean(r.root) {
					r.admin = admin
					r.key = "linked:" + filepath.Base(admin)
					break
				}
			}
		}
		if r.admin == "" {
			complete = api.Partial
			result.diagnostics = append(result.diagnostics, diagnostic(api.Unavailable, "UnidentifiedWorktree", "A registered worktree has no established administrative identity."))
			continue
		}
		id, err := domain.NewWorktreeID(repo.id, r.key)
		if err != nil {
			return result, err
		}
		d := api.WorktreeFactsData{ID: id, Primary: r.primary, Current: currentAdmin == r.admin}
		available, _ := api.NewAvailableWorktree(api.AvailableWorktreeData{})
		d.Availability = available
		root, e := observeDirectory(r.root)
		if e != nil {
			complete = api.Partial
			if os.IsNotExist(e) {
				missing, _ := api.NewMissingWorktree(api.MissingWorktreeData{})
				d.Availability = missing
			} else {
				diag := diagnostic(api.Unavailable, "WorktreeRootUnavailable", "A registered worktree root could not be acquired.")
				un, _ := api.NewUnresolvedWorktree(api.UnresolvedWorktreeData{Diagnostic: diag})
				d.Availability = un
				result.diagnostics = append(result.diagnostics, diag)
			}
		} else {
			// Native acquisition from this root must return the registered admin
			// and common identities; a transplanted directory is not its scope.
			actual := s.command(root.path, "rev-parse", "--absolute-git-dir")
			ad, ae := observeDirectory(line(actual.stdout))
			if actual.err != nil || ae != nil || ad.path != r.admin {
				complete = api.Partial
				diag := diagnostic(api.StaleObservation, "WorktreeRegistrationChanged", "The root no longer resolves to its registered administrative directory.")
				un, _ := api.NewUnresolvedWorktree(api.UnresolvedWorktreeData{Diagnostic: diag})
				d.Availability = un
				result.diagnostics = append(result.diagnostics, diag)
			} else {
				scope, err := api.NewWorktreeScope(api.WorktreeScopeData{ID: id, RootLocator: root.path, RootIdentity: root.identity, Source: sourceVersion("root", repo.id.Token(), s.a.lifetime, []byte(directoryKey(root)))})
				if err != nil {
					return result, err
				}
				d.Scope = api.Some(scope)
				if r.isLocked {
					locked, _ := api.NewLockedWorktree(api.LockedWorktreeData{Reason: r.locked})
					d.Availability = locked
				} else if r.isPrunable {
					prunable, _ := api.NewPrunableWorktree(api.PrunableWorktreeData{Reason: r.prunable})
					d.Availability = prunable
				}
			}
		}
		head, err := s.readHead(repo, r.admin)
		if err != nil {
			complete = api.Partial
			result.diagnostics = append(result.diagnostics, safeError(err))
		} else {
			d.Head = api.Some(head)
		}
		data = append(data, d)
	}
	result.observation, err = s.observation(repo.id, api.None[domain.WorktreeID](), version, complete)
	if err != nil {
		return result, err
	}
	for i := range data {
		d := data[i]
		d.Observation, err = s.observation(repo.id, api.Some(d.ID), version, complete)
		if err != nil {
			return result, err
		}
		if head, p := d.Head.Value(); p {
			if branch, p := head.Branch(); p {
				var occupants []domain.WorktreeID
				for _, other := range data {
					if h, p := other.Head.Value(); p {
						if b, p := h.Branch(); p && b == branch {
							occupants = append(occupants, other.ID)
						}
					}
				}
				occupancy, err := api.NewBranchOccupancy(api.BranchOccupancyData{Branch: branch, Worktrees: occupants, Observation: result.observation})
				if err != nil {
					return result, err
				}
				d.Occupancy = []api.BranchOccupancy{occupancy}
			}
		}
		fact, err := api.NewWorktreeFacts(d)
		if err != nil {
			return result, err
		}
		result.facts = append(result.facts, fact)
	}
	return result, nil
}

func administrativeRoots(common string) (map[string]string, error) {
	result := make(map[string]string)
	parent := filepath.Join(common, "worktrees")
	f, err := os.Open(parent)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := f.ReadDir(10001)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(entries) > 10000 {
		return nil, diagnostic(api.Unavailable, "InventoryLimit", "The administrative worktree inventory exceeds its bound.")
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		admin, err := observeDirectory(filepath.Join(parent, entry.Name()))
		if err != nil {
			return nil, err
		}
		path, err := readSmallFile(filepath.Join(admin.path, "gitdir"), 32768)
		if err != nil {
			return nil, err
		}
		root := filepath.Dir(line(path))
		if !filepath.IsAbs(root) {
			return nil, diagnostic(api.Unavailable, "MalformedRegistration", "A native worktree registration has no absolute root.")
		}
		if _, present := result[root]; present {
			return nil, diagnostic(api.Unavailable, "AmbiguousRegistration", "More than one native administration record claims the same root.")
		}
		result[root] = admin.path
	}
	return result, nil
}

func (s *readSession) readHead(repo repository, admin string) (domain.Head, error) {
	raw, err := readSmallFile(filepath.Join(admin, "HEAD"), 4096)
	if err != nil {
		return domain.Head{}, err
	}
	value := line(raw)
	if strings.HasPrefix(value, "ref: ") {
		ref := strings.TrimPrefix(value, "ref: ")
		if !strings.HasPrefix(ref, "refs/heads/") {
			return domain.Head{}, diagnostic(api.Unavailable, "UnsupportedHeadRef", "HEAD does not name a local branch.")
		}
		branch, err := domain.NewBranchID(repo.id, domain.Local, strings.TrimPrefix(ref, "refs/heads/"))
		if err != nil {
			return domain.Head{}, err
		}
		q := s.command(repo.common.path, "--git-dir="+admin, "for-each-ref", "--format=%(refname)%00%(objectname)", "--", ref)
		if q.err != nil {
			return domain.Head{}, q.err
		}
		var oid string
		for _, row := range bytes.Split(q.stdout, []byte{'\n'}) {
			name, obj, ok := bytes.Cut(row, []byte{0})
			if ok && string(name) == ref {
				oid = string(obj)
			}
		}
		if oid == "" {
			return domain.NewUnbornHead(branch)
		}
		revision, err := s.verifyCommit(repo, oid)
		if err != nil {
			return domain.Head{}, err
		}
		return domain.NewAttachedHead(branch, revision)
	}
	revision, err := s.verifyCommit(repo, value)
	if err != nil {
		return domain.Head{}, err
	}
	return domain.NewDetachedHead(revision)
}

func (s *readSession) verifyCommit(repo repository, oidText string) (domain.Revision, error) {
	oid, err := domain.NewOID(oidText)
	if err != nil || oid.Format() != repo.format {
		return domain.Revision{}, diagnostic(api.Invalid, "InvalidExactObject", "The requested object is not a full object identity in this repository format.")
	}
	q := s.command(repo.common.path, "--git-dir="+repo.common.path, "cat-file", "-t", oid.String())
	if q.err != nil {
		return domain.Revision{}, q.err
	}
	if line(q.stdout) != "commit" {
		return domain.Revision{}, diagnostic(api.Invalid, "ObjectIsNotCommit", "The exact object is not a commit.")
	}
	return domain.NewRevision(repo.id, oid)
}

func readSmallFile(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, diagnostic(api.Unsupported, "UnsupportedMetadataFile", "Native metadata is not a bounded regular file.")
	}
	buffer, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(buffer)) > limit {
		return nil, diagnostic(api.Unavailable, "MetadataLimit", "Native metadata exceeds its bound.")
	}
	return buffer, nil
}
