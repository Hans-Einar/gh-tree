package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type continuation struct {
	scope   string
	page    int
	expires time.Time
	seen    map[string]string
}

func (a *Adapter) pageRequest(scope string, p api.PageRequest) (continuation, error) {
	if !p.Valid() {
		return continuation{}, protocolError("invalid page")
	}
	d := p.Data()
	if d.Limit == 0 || d.Limit > 100 {
		return continuation{}, protocolError("page size")
	}
	switch c := d.Continuation.(type) {
	case api.InitialPage:
		return continuation{scope: scope, page: 1, seen: map[string]string{}}, nil
	case api.CursorPage:
		a.mu.Lock()
		defer a.mu.Unlock()
		v, ok := a.cursors[c.Data().Cursor]
		if !ok || v.scope != scope || !time.Now().Before(v.expires) || v.page > 10 {
			return continuation{}, protocolError("foreign expired or capped cursor")
		}
		v.seen = copySeen(v.seen)
		return v, nil
	default:
		return continuation{}, protocolError("offset pagination unsupported")
	}
}
func copySeen(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
func (a *Adapter) version(scope string) api.SourceVersion {
	token := fmt.Sprint(a.sequence.Add(1))
	return must(api.NewSourceVersion("github", scope, a.issuer, token))
}
func (a *Adapter) observe(repo domain.RepositoryID, r response, page api.PageInfo) api.RemoteObservation {
	id := must(api.NewObservationID(a.issuer + "/" + fmt.Sprint(a.sequence.Add(1))))
	return must(api.NewRemoteObservation(api.RemoteObservationData{ID: id, Repository: repo, Interval: r.interval(), Version: page.Data().Source, Origin: must(api.NewLiveRemoteObservation(api.LiveRemoteObservationData{})), Page: page}))
}
func unknownPage(version api.SourceVersion) api.PageInfo {
	return must(api.NewPageInfo(api.PageInfoData{Completeness: api.Unknown, Source: version}))
}
func (a *Adapter) finishPage(v continuation, version api.SourceVersion, count int, more, unknown bool) (api.PageInfo, []api.Diagnostic) {
	d := api.PageInfoData{Returned: uint32(count), Completeness: api.Complete, HasMore: api.Some(more), Source: version}
	var diagnostics []api.Diagnostic
	if more {
		d.Completeness = api.More
		if v.page < 10 {
			a.mu.Lock()
			now := time.Now()
			for k, c := range a.cursors {
				if !now.Before(c.expires) {
					delete(a.cursors, k)
				}
			}
			if len(a.cursors) < 256 {
				token := a.issuer + "/" + fmt.Sprint(a.sequence.Add(1))
				v.page++
				v.expires = now.Add(10 * time.Minute)
				v.seen = copySeen(v.seen)
				a.cursors[token] = v
				d.Next = api.Some[api.PageContinuation](must(api.NewCursorPage(api.CursorPageData{Cursor: token})))
			} else {
				diagnostics = append(diagnostics, diagnostic(api.Busy, "cursor-capacity-reached"))
			}
			a.mu.Unlock()
		} else {
			diagnostics = append(diagnostics, diagnostic(api.Unavailable, "refresh-ten-page-cap"))
		}
	}
	if unknown {
		d.Completeness = api.Unknown
		d.HasMore = api.None[bool]()
	}
	return must(api.NewPageInfo(d)), diagnostics
}

// Link targets are validated as evidence only; never follow an untrusted URL.
// The next invocation reconstructs the exact endpoint/query from its bound scope.
func nextPage(r response, path string, page, limit int, rawCount int) (bool, error) {
	if rawCount > limit {
		return false, protocolError("oversized page")
	}
	link := r.headers.Get("Link")
	if link == "" {
		return false, nil
	}
	found := false
	for _, part := range strings.Split(link, ",") {
		bits := strings.Split(strings.TrimSpace(part), ";")
		if len(bits) < 2 {
			return false, protocolError("invalid Link")
		}
		next := false
		for _, b := range bits[1:] {
			if strings.TrimSpace(b) == `rel="next"` {
				next = true
			}
		}
		if !next {
			continue
		}
		if found {
			return false, protocolError("duplicate next Link")
		}
		found = true
		target := strings.TrimSpace(bits[0])
		if len(target) < 2 || target[0] != '<' || target[len(target)-1] != '>' {
			return false, protocolError("invalid next Link")
		}
		u, e := url.Parse(target[1 : len(target)-1])
		if e != nil || u.User != nil || u.Fragment != "" || u.Scheme != "https" {
			return false, protocolError("invalid next URL")
		}
		expected, e := url.Parse(path)
		if e != nil {
			return false, e
		}
		apiHost := r.wireHost()
		if u.Host != apiHost && u.Host != "api."+apiHost {
			return false, protocolError("next host mismatch")
		}
		actualPath := strings.TrimPrefix(u.Path, "/api/v3/")
		actualPath = strings.TrimPrefix(actualPath, "/")
		if actualPath != expected.Path {
			return false, protocolError("next endpoint mismatch")
		}
		q := expected.Query()
		q.Set("page", strconv.Itoa(page+1))
		if u.Query().Encode() != q.Encode() {
			return false, protocolError("next query mismatch")
		}
	}
	return found, nil
}
func (r response) wireHost() string { return r.host }
func fingerprint(v string) string   { sum := sha256.Sum256([]byte(v)); return hex.EncodeToString(sum[:]) }
func seenFact(c *continuation, key, value string) (duplicate, conflict bool) {
	key = fingerprint(key)
	h := fingerprint(value)
	if previous, ok := c.seen[key]; ok {
		return previous == h, previous != h
	}
	c.seen[key] = h
	return false, false
}
