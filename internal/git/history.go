package git

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

const maxReadPage uint32 = 1000

// Cursor payload binds exact immutable endpoints/traversal/filter and source.
// Its MAC prevents transplanting a cursor from another adapter or query. There
// is no cursor registry and no unbounded retention of accumulated pages.
func (a *Adapter) pageOffset(page api.PageRequest, binding string) (uint64, error) {
	if !page.Valid() || page.Data().Limit > maxReadPage {
		return 0, diagnostic(api.Invalid, "PageLimit", "The requested page limit exceeds the adapter bound.")
	}
	switch continuation := page.Data().Continuation.(type) {
	case api.InitialPage:
		return 0, nil
	case api.OffsetPage:
		return continuation.Data().Offset, nil
	case api.CursorPage:
		raw, err := base64.RawURLEncoding.DecodeString(continuation.Data().Cursor)
		if err != nil || len(raw) > 4096 {
			return 0, diagnostic(api.Invalid, "InvalidPageCursor", "The page cursor is invalid.")
		}
		payload, mac, ok := bytes.Cut(raw, []byte{0})
		if !ok {
			return 0, diagnostic(api.Invalid, "InvalidPageCursor", "The page cursor is invalid.")
		}
		h := hmac.New(sha256.New, []byte(a.lifetime))
		h.Write(payload)
		if !hmac.Equal(mac, h.Sum(nil)) {
			return 0, diagnostic(api.StaleObservation, "ForeignPageCursor", "The page cursor belongs to another adapter or is changed.")
		}
		prefix, offset, ok := strings.Cut(string(payload), ":")
		if !ok || prefix != binding {
			return 0, diagnostic(api.StaleObservation, "PageSourceChanged", "The cursor source, endpoints or traversal changed.")
		}
		n, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			return 0, diagnostic(api.Invalid, "InvalidPageCursor", "The page cursor offset is invalid.")
		}
		return n, nil
	default:
		return 0, diagnostic(api.Invalid, "InvalidPageCursor", "The page continuation is invalid.")
	}
}

func queryBinding(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(h, "%d:", len(part))
		h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (a *Adapter) pageInfo(version api.SourceVersion, binding string, offset uint64, count int, more bool, complete api.Completeness) (api.PageInfo, error) {
	d := api.PageInfoData{Source: version, Returned: uint32(count), Completeness: complete}
	if complete == api.Complete || complete == api.More {
		d.HasMore = api.Some(more)
	}
	if more {
		if offset > ^uint64(0)-uint64(count) {
			return api.PageInfo{}, diagnostic(api.Invalid, "PageOffsetOverflow", "The page continuation is exhausted.")
		}
		payload := []byte(binding + ":" + strconv.FormatUint(offset+uint64(count), 10))
		h := hmac.New(sha256.New, []byte(a.lifetime))
		h.Write(payload)
		encoded := base64.RawURLEncoding.EncodeToString(append(append(payload, 0), h.Sum(nil)...))
		cursor, err := api.NewCursorPage(api.CursorPageData{Cursor: encoded})
		if err != nil {
			return api.PageInfo{}, err
		}
		d.Next = api.Some[api.PageContinuation](cursor)
	}
	return api.NewPageInfo(d)
}

func (a *Adapter) ReadCommits(ctx context.Context, request api.ReadCommitsRequest) (api.ReadCommitsResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.ReadCommitsResult{}, diagnostic(api.Invalid, "InvalidRequest", "The history request is invalid.")
	}
	rd := request.Data()
	d := api.ReadCommitsResultData{Endpoint: rd.Endpoint}
	repo, err := a.registered(s.ctx, rd.Endpoint.Repository())
	contextSource := ""
	if err == nil {
		contextSource, err = historyContext(repo)
	}
	binding := queryBinding("commits", rd.Endpoint.Repository().Token(), rd.Endpoint.OID().String(), strconv.Itoa(int(rd.Traversal)), contextSource)
	version := sourceVersion("history", rd.Endpoint.Repository().Token(), a.lifetime, []byte(binding))
	complete := api.Complete
	more := false
	var offset uint64
	if err == nil {
		offset, err = a.pageOffset(rd.Page, binding)
	}
	if err == nil {
		_, err = s.verifyCommit(repo, rd.Endpoint.OID().String())
	}
	if err == nil {
		args := []string{"--git-dir=" + repo.common.path, "rev-list", "--topo-order", "--max-count=" + strconv.FormatUint(uint64(rd.Page.Data().Limit)+1, 10), "--skip=" + strconv.FormatUint(offset, 10)}
		if rd.Traversal == api.FirstParent {
			args = append(args, "--first-parent")
		}
		args = append(args, rd.Endpoint.OID().String(), "--")
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
		complete = api.Partial
		more = false
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	observation, oe := s.observation(rd.Endpoint.Repository(), api.None[domain.WorktreeID](), version, complete)
	if oe != nil {
		return api.ReadCommitsResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Page, oe = a.pageInfo(version, binding, offset, len(d.Commits), more, complete)
	if oe != nil {
		return api.ReadCommitsResult{}, oe
	}
	d.Transport = transportValue(s.transport)
	result, ve := api.NewReadCommitsResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}

func historyContext(repo repository) (string, error) {
	parts := []string{"native-history"}
	for _, name := range []string{"shallow", filepath.Join("info", "grafts")} {
		raw, err := readSmallFile(filepath.Join(repo.common.path, name), 1<<20)
		if os.IsNotExist(err) {
			parts = append(parts, name, "absent")
			continue
		}
		if err != nil {
			return "", err
		}
		parts = append(parts, name, "present", string(raw))
	}
	return queryBinding(parts...), nil
}

func (s *readSession) commit(repo repository, oid string) (api.CommitFact, error) {
	revision, err := s.verifyCommit(repo, oid)
	if err != nil {
		return api.CommitFact{}, err
	}
	q := s.command(repo.common.path, "--git-dir="+repo.common.path, "cat-file", "commit", revision.OID().String())
	if q.err != nil {
		return api.CommitFact{}, q.err
	}
	header, message, ok := bytes.Cut(q.stdout, []byte{'\n', '\n'})
	if !ok {
		return api.CommitFact{}, diagnostic(api.Unavailable, "MalformedCommit", "The native commit object has no message delimiter.")
	}
	d := api.CommitFactData{Revision: revision, Message: string(message)}
	var haveAuthor, haveCommitter bool
	for _, row := range bytes.Split(header, []byte{'\n'}) {
		key, value, ok := strings.Cut(string(row), " ")
		if !ok {
			continue
		}
		switch key {
		case "parent":
			oid, err := domain.NewOID(value)
			if err != nil {
				return api.CommitFact{}, err
			}
			parent, err := domain.NewRevision(repo.id, oid)
			if err != nil {
				return api.CommitFact{}, err
			}
			d.Parents = append(d.Parents, parent)
		case "author":
			d.AuthorName, d.AuthorEmail, d.AuthorTime, err = parseSignature(value)
			haveAuthor = err == nil
		case "committer":
			_, _, d.CommitterTime, err = parseSignature(value)
			haveCommitter = err == nil
		}
		if err != nil {
			return api.CommitFact{}, err
		}
	}
	if !haveAuthor || !haveCommitter {
		return api.CommitFact{}, diagnostic(api.Unavailable, "MalformedCommit", "The native commit lacks established author/committer metadata.")
	}
	return api.NewCommitFact(d)
}

func parseSignature(value string) (string, string, time.Time, error) {
	bad := func() (string, string, time.Time, error) {
		return "", "", time.Time{}, diagnostic(api.Unavailable, "MalformedSignature", "Native author or committer metadata is malformed.")
	}
	close := strings.LastIndexByte(value, '>')
	if close < 0 {
		return bad()
	}
	open := strings.LastIndexByte(value[:close], '<')
	if open < 1 || value[open-1] != ' ' {
		return bad()
	}
	fields := strings.Fields(value[close+1:])
	if len(fields) != 2 {
		return bad()
	}
	seconds, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return bad()
	}
	zone := fields[1]
	if len(zone) != 5 || (zone[0] != '+' && zone[0] != '-') {
		return bad()
	}
	hours, err := strconv.Atoi(zone[1:3])
	if err != nil || hours > 23 {
		return bad()
	}
	minutes, err := strconv.Atoi(zone[3:])
	if err != nil || minutes > 59 {
		return bad()
	}
	offset := hours*3600 + minutes*60
	if zone[0] == '-' {
		offset = -offset
	}
	return value[:open-1], value[open+1 : close], time.Unix(seconds, 0).In(time.FixedZone(zone, offset)), nil
}

func (a *Adapter) MergeBase(ctx context.Context, request api.MergeBaseRequest) (api.MergeBaseResult, error) {
	s, cancel := a.readSession(ctx)
	defer cancel()
	if !request.Valid() {
		return api.MergeBaseResult{}, diagnostic(api.Invalid, "InvalidRequest", "The merge-base request is invalid.")
	}
	rd := request.Data()
	d := api.MergeBaseResultData{Left: rd.Left, Right: rd.Right}
	repo, err := a.registered(s.ctx, rd.Left.Repository())
	if err == nil {
		_, err = s.verifyCommit(repo, rd.Left.OID().String())
	}
	if err == nil {
		_, err = s.verifyCommit(repo, rd.Right.OID().String())
	}
	if err == nil {
		q := s.command(repo.common.path, "--git-dir="+repo.common.path, "merge-base", "--all", rd.Left.OID().String(), rd.Right.OID().String())
		exit, _ := q.transport.Data().ExitCode.Value()
		if q.err != nil && exit == 1 && len(q.stdout) == 0 {
			out, oe := api.NewNoCommonAncestor(api.NoCommonAncestorData{})
			if oe != nil {
				err = oe
			} else {
				d.Outcome = api.Some[api.MergeBaseOutcome](out)
			}
		} else if q.err != nil {
			err = q.err
		} else {
			var bases []domain.Revision
			for _, row := range bytes.Split(bytes.TrimSuffix(q.stdout, []byte{'\n'}), []byte{'\n'}) {
				base, be := s.verifyCommit(repo, string(row))
				if be != nil {
					err = be
					break
				}
				bases = append(bases, base)
			}
			if err == nil {
				if len(bases) == 1 {
					out, oe := api.NewUniqueMergeBase(api.UniqueMergeBaseData{Base: bases[0]})
					err = oe
					if oe == nil {
						d.Outcome = api.Some[api.MergeBaseOutcome](out)
					}
				} else if len(bases) > 1 {
					out, oe := api.NewAmbiguousMergeBase(api.AmbiguousMergeBaseData{Candidates: bases})
					err = oe
					if oe == nil {
						d.Outcome = api.Some[api.MergeBaseOutcome](out)
					}
				} else {
					err = diagnostic(api.Unavailable, "MergeBaseUnavailable", "Git returned no classified merge-base result.")
				}
			}
		}
	}
	completeness := api.Complete
	if err != nil {
		completeness = api.Partial
		d.Diagnostics = append(d.Diagnostics, safeError(err))
	}
	version := sourceVersion("merge-base", rd.Left.Repository().Token(), a.lifetime, []byte(queryBinding(rd.Left.OID().String(), rd.Right.OID().String())))
	observation, oe := s.observation(rd.Left.Repository(), api.None[domain.WorktreeID](), version, completeness)
	if oe != nil {
		return api.MergeBaseResult{}, oe
	}
	d.Observation = api.Some(observation)
	d.Transport = transportValue(s.transport)
	result, ve := api.NewMergeBaseResult(d)
	if ve != nil {
		return result, ve
	}
	return result, err
}
