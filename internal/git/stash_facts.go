package git

import (
	"bytes"
	"context"
	"encoding/hex"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type stashOccurrence struct {
	old, new           domain.OID
	oldNull            bool
	signature, message string
	raw, identity      []byte
}

func parseStashLog(raw []byte, format domain.ObjectFormat) ([]stashOccurrence, error) {
	bad := func() ([]stashOccurrence, error) {
		return nil, diagnostic(api.Unavailable, "MalformedStashLog", "The native stash reflog is malformed or exceeds its supported bound.")
	}
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[len(raw)-1] != '\n' {
		return bad()
	}
	var records []stashOccurrence
	width := format.ByteLength() * 2
	for _, row := range bytes.Split(raw[:len(raw)-1], []byte{'\n'}) {
		if len(row) <= width*2+2 || row[width] != ' ' || row[width*2+1] != ' ' {
			return bad()
		}
		oldText, newText := string(row[:width]), string(row[width+1:width*2+1])
		r := stashOccurrence{oldNull: oldText == strings.Repeat("0", width)}
		if !r.oldNull {
			oid, err := domain.NewOID(oldText)
			if err != nil || oid.Format() != format {
				return bad()
			}
			r.old = oid
		} else {
			if _, err := hex.DecodeString(oldText); err != nil {
				return bad()
			}
		}
		oid, err := domain.NewOID(newText)
		if err != nil || oid.Format() != format {
			return bad()
		}
		r.new = oid
		signature, message, ok := bytes.Cut(row[width*2+2:], []byte{'\t'})
		if !ok {
			return bad()
		}
		if _, _, _, err := parseSignature(string(signature)); err != nil {
			return bad()
		}
		r.signature = string(signature)
		r.message = string(message)
		r.raw = append(append([]byte(nil), row...), '\n')
		// The previous-link field changes during native survivor chaining. It is
		// deliberately excluded; full newOID plus original author/time/message
		// bytes remain stable across unrelated prepends and positional shifts.
		r.identity = append(append([]byte(nil), row[width+1:]...), '\n')
		records = append(records, r)
		if len(records) > 100000 {
			return bad()
		}
	}
	return records, nil
}

func (s *readSession) stashParents(repo repository, stash domain.OID) ([]domain.OID, error) {
	commit, err := s.commit(repo, stash.String())
	if err != nil {
		return nil, err
	}
	parents := commit.Data().Parents
	if len(parents) < 2 || len(parents) > 3 {
		return nil, diagnostic(api.Invalid, "InvalidStashStructure", "The selected object has no supported native stash parent structure.")
	}
	index, err := s.commit(repo, parents[1].OID().String())
	if err != nil {
		return nil, err
	}
	if len(index.Data().Parents) != 1 || index.Data().Parents[0] != parents[0] {
		return nil, diagnostic(api.Invalid, "InvalidStashIndexParent", "The stash index parent does not share the selected base.")
	}
	if len(parents) == 3 {
		untracked, err := s.commit(repo, parents[2].OID().String())
		if err != nil {
			return nil, err
		}
		if len(untracked.Data().Parents) != 0 {
			return nil, diagnostic(api.Invalid, "InvalidStashUntrackedParent", "The stash untracked parent is not an independent root.")
		}
	}
	result := make([]domain.OID, len(parents))
	for i, p := range parents {
		result[i] = p.OID()
	}
	return result, nil
}

func (a *Adapter) ListStashes(ctx context.Context, request api.ListStashesRequest) (api.ListStashesResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.ListStashesResult{}, diagnostic(api.Invalid, "InvalidRequest", "The stash inventory request is invalid.")
	}
	rd := request.Data()
	d := api.ListStashesResultData{}
	repo, err := a.registered(s.ctx, rd.Repository)
	var raw []byte
	var records []stashOccurrence
	if err == nil {
		if repo.backend == api.FilesRefs {
			raw, err = readSmallFile(filepath.Join(repo.common.path, "logs", "refs", "stash"), int64(a.options.MaxStdoutBytes))
			if os.IsNotExist(err) {
				raw = nil
				err = nil
			}
			if err == nil {
				records, err = parseStashLog(raw, repo.format)
			}
		} else {
			// Native normalized log metadata preserves read capability on reftable;
			// this occurrence profile never grants files-backend drop authority.
			q := s.command(repo.common.path, "--git-dir="+repo.common.path, "for-each-ref", "--format=%(refname)", "--", "refs/stash")
			err = q.err
			if err == nil && len(q.stdout) > 0 {
				q = s.command(repo.common.path, "--git-dir="+repo.common.path, "log", "-g", "--date=raw", "--format=%H%x00%gn%x00%ge%x00%gD%x00%gs%x00", "refs/stash", "--")
				raw = q.stdout
				err = q.err
				if err == nil {
					records, err = parseNativeStashLog(raw, repo.format)
				}
			}
		}
	}
	binding := queryBinding("stashes", rd.Repository.Token(), string(raw))
	version := sourceVersion("stashes", rd.Repository.Token(), a.lifetime, []byte(binding))
	var offset uint64
	complete := api.Complete
	more := false
	if err == nil {
		if _, p := rd.Page.Data().Continuation.(api.OffsetPage); p {
			err = diagnostic(api.Invalid, "SourceBoundCursorRequired", "Mutable stash pages require the returned source-bound cursor.")
		} else {
			offset, err = a.pageOffset(rd.Page, binding)
		}
	}
	observation, oe := s.observation(rd.Repository, api.None[domain.WorktreeID](), version, complete)
	if oe != nil {
		return api.ListStashesResult{}, oe
	}
	if err == nil && offset < uint64(len(records)) {
		end := offset + uint64(rd.Page.Data().Limit)
		if end < uint64(len(records)) {
			more = true
			complete = api.More
		} else {
			end = uint64(len(records))
		}
		for position := offset; position < end; position++ {
			r := records[len(records)-1-int(position)]
			parents, pe := s.stashParents(repo, r.new)
			if pe != nil {
				err = pe
				break
			}
			id, ie := domain.NewStashID(repo.id, r.new)
			if ie != nil {
				err = ie
				break
			}
			name, email, when, se := parseSignature(r.signature)
			if se != nil {
				err = se
				break
			}
			fact := api.StashFactData{ID: id, Parents: parents, Occurrence: sourceVersion("stash-occurrence", repo.id.Token(), a.lifetime, r.identity), DisplayPosition: position, Message: r.message, AuthorName: name, AuthorEmail: email, AuthorTime: api.Some(when), Observation: observation}
			if origin, p := legacyStashOrigin(repo.id, r.message); p {
				fact.Origin = api.Some(origin)
			}
			value, ve := api.NewStashFact(fact)
			if ve != nil {
				err = ve
				break
			}
			d.Stashes = append(d.Stashes, value)
		}
	}
	if err != nil {
		complete = api.Partial
		more = false
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	observation, oe = s.observation(rd.Repository, api.None[domain.WorktreeID](), version, complete)
	if oe != nil {
		return api.ListStashesResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Page, oe = a.pageInfo(version, binding, offset, len(d.Stashes), more, complete)
	if oe != nil {
		return api.ListStashesResult{}, oe
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewListStashesResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func legacyStashOrigin(repo domain.RepositoryID, message string) (api.StashOrigin, bool) {
	position := strings.Index(message, "gh-tree?")
	if position < 0 {
		return api.StashOrigin{}, false
	}
	raw := message[position+len("gh-tree?"):]
	values, err := url.ParseQuery(raw)
	if err != nil {
		return api.StashOrigin{}, false
	}
	d := api.StashOriginData{LegacyManaged: api.Some(message[position:])}
	if branch, err := domain.NewBranchID(repo, domain.Local, values.Get("branch")); err == nil {
		d.Branch = api.Some(branch)
	}
	// Historical path/head/time remain exact descriptive legacy bytes. They do
	// not mint a WorktreeID, a current Revision or destructive authorization.
	value, err := api.NewStashOrigin(d)
	return value, err == nil
}

func parseNativeStashLog(raw []byte, format domain.ObjectFormat) ([]stashOccurrence, error) {
	bad := func() ([]stashOccurrence, error) {
		return nil, diagnostic(api.Unavailable, "MalformedNativeStashLog", "Native reflog metadata could not establish an exact occurrence.")
	}
	var newest []stashOccurrence
	for _, row := range bytes.Split(bytes.TrimSuffix(raw, []byte{'\n'}), []byte{'\n'}) {
		if len(row) == 0 {
			continue
		}
		fields := bytes.Split(row, []byte{0})
		if len(fields) != 6 {
			return bad()
		}
		oid, err := domain.NewOID(string(fields[0]))
		if err != nil || oid.Format() != format {
			return bad()
		}
		selector := string(fields[3])
		if !strings.HasPrefix(selector, "refs/stash@{") || !strings.HasSuffix(selector, "}") {
			return bad()
		}
		when := strings.TrimSuffix(strings.TrimPrefix(selector, "refs/stash@{"), "}")
		signature := string(fields[1]) + " <" + string(fields[2]) + "> " + when
		if _, _, _, err := parseSignature(signature); err != nil {
			return bad()
		}
		identity := []byte(oid.String() + " " + signature + "\t" + string(fields[4]) + "\n")
		newest = append(newest, stashOccurrence{new: oid, signature: signature, message: string(fields[4]), identity: identity})
		if len(newest) > 100000 {
			return bad()
		}
	}
	records := make([]stashOccurrence, len(newest))
	for i, r := range newest {
		records[len(newest)-1-i] = r
	}
	return records, nil
}
