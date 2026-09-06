package adapter

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// Validate JSON before decoding: duplicate members, invalid UTF-8, excessive
// nesting, trailing values and malformed surrogate escapes cannot silently alter
// identity. Unknown provider fields are allowed but subject to the same bounds.
func strictJSON(raw []byte, into any) error {
	if !utf8.Valid(raw) {
		return protocolError("invalid UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := jsonValue(d, 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return protocolError("trailing JSON")
	}
	// encoding/json replaces invalid UTF-16 surrogate escapes, so reject those
	// explicitly instead of accepting a rewritten branch/repository identity.
	for i := 0; i < len(raw); i++ {
		if raw[i] != '"' {
			continue
		}
		i++
		for i < len(raw) && raw[i] != '"' {
			if raw[i] == '\\' {
				i++
				if i < len(raw) && raw[i] == 'u' {
					if i+4 >= len(raw) {
						return protocolError("unicode escape")
					}
					n, e := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
					if e != nil {
						return e
					}
					i += 4
					if n >= 0xd800 && n <= 0xdbff {
						if i+6 >= len(raw) || string(raw[i+1:i+3]) != "\\u" {
							return protocolError("unpaired surrogate")
						}
						m, e := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
						if e != nil || m < 0xdc00 || m > 0xdfff {
							return protocolError("unpaired surrogate")
						}
						i += 6
					} else if n >= 0xdc00 && n <= 0xdfff {
						return protocolError("unpaired surrogate")
					}
				}
			}
			i++
		}
	}
	if err := json.Unmarshal(raw, into); err != nil {
		return protocolError("invalid response fields")
	}
	return nil
}

func decodeList(raw []byte) ([]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, protocolError("invalid UTF-8")
	}
	var records []json.RawMessage
	if e := json.Unmarshal(raw, &records); e != nil || records == nil {
		return nil, protocolError("invalid list shape")
	}
	return records, nil
}
func jsonValue(d *json.Decoder, depth int) error {
	if depth > 64 {
		return protocolError("JSON nesting limit")
	}
	t, e := d.Token()
	if e != nil {
		return protocolError("malformed JSON")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			k, e := d.Token()
			if e != nil {
				return e
			}
			s, ok := k.(string)
			if !ok || seen[s] {
				return protocolError("duplicate JSON member")
			}
			seen[s] = true
			if e = jsonValue(d, depth+1); e != nil {
				return e
			}
		}
	case '[':
		for d.More() {
			if e = jsonValue(d, depth+1); e != nil {
				return e
			}
		}
	default:
		return protocolError("JSON delimiter")
	}
	_, e = d.Token()
	return e
}

type repositoryDTO struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	DefaultBranch *string `json:"default_branch"`
	Archived      bool    `json:"archived"`
	Disabled      bool    `json:"disabled"`
}

func parseRepository(raw []byte, host string) (api.RemoteRepository, error) {
	var d repositoryDTO
	if e := strictJSON(raw, &d); e != nil {
		return api.RemoteRepository{}, e
	}
	return mapRepository(d, host)
}
func mapRepository(d repositoryDTO, host string) (api.RemoteRepository, error) {
	if d.ID == 0 || d.Name == "" || d.Owner.Login == "" || d.FullName != d.Owner.Login+"/"+d.Name {
		return api.RemoteRepository{}, protocolError("repository identity")
	}
	l, e := api.NewRemoteRepositoryLocator(api.RemoteRepositoryLocatorData{Host: host, Owner: strings.ToLower(d.Owner.Login), Name: strings.ToLower(d.Name)})
	if e != nil {
		return api.RemoteRepository{}, e
	}
	if !providerLocator(l) {
		return api.RemoteRepository{}, protocolError("unsupported repository components")
	}
	if !validURL(d.HTMLURL, l, 0) {
		return api.RemoteRepository{}, protocolError("repository URL")
	}
	id := repositoryID(l)
	var branch api.Optional[domain.BranchID]
	if d.DefaultBranch != nil && *d.DefaultBranch != "" {
		b, e := domain.NewBranchID(id, domain.RemoteHead, *d.DefaultBranch)
		if e != nil {
			return api.RemoteRepository{}, e
		}
		branch = api.Some(b)
	}
	var diagnostics []api.Diagnostic
	if d.Archived || d.Disabled {
		diagnostics = append(diagnostics, diagnostic(api.Unsupported, "repository-read-only"))
	}
	caps := must(api.NewRemoteCapabilities(api.RemoteCapabilitiesData{ReadBranches: !d.Disabled, ReadPullRequests: !d.Disabled, CreatePullRequest: !d.Archived && !d.Disabled, SupportedObjectFormats: []domain.ObjectFormat{domain.SHA1}, Diagnostics: diagnostics}))
	return api.NewRemoteRepository(api.RemoteRepositoryData{ID: id, Locator: l, URL: repoURL(l), DefaultBranch: branch, Capabilities: caps})
}
func validURL(s string, l api.RemoteRepositoryLocator, number uint64) bool {
	u, e := url.Parse(s)
	if e != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || u.Port() != "" {
		return false
	}
	d := l.Data()
	path := "/" + d.Owner + "/" + d.Name
	if number != 0 {
		path += "/pull/" + strconv.FormatUint(number, 10)
	}
	if number == 0 {
		return strings.EqualFold(u.Host, d.Host) && strings.EqualFold(u.Path, path)
	}
	parts := strings.Split(u.Path, "/")
	return strings.EqualFold(u.Host, d.Host) && len(parts) == 5 && strings.EqualFold(parts[1], d.Owner) && strings.EqualFold(parts[2], d.Name) && parts[3] == "pull" && parts[4] == strconv.FormatUint(number, 10)
}

type branchDTO struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func parseBranch(raw []byte, repo domain.RepositoryID) (domain.BranchID, domain.Revision, error) {
	var d branchDTO
	if e := strictJSON(raw, &d); e != nil {
		return domain.BranchID{}, domain.Revision{}, e
	}
	b, e := domain.NewBranchID(repo, domain.RemoteHead, d.Name)
	if e != nil {
		return b, domain.Revision{}, e
	}
	r, e := remoteRevision(repo, d.Commit.SHA)
	return b, r, e
}
func remoteRevision(repo domain.RepositoryID, s string) (domain.Revision, error) {
	o, e := domain.NewOID(s)
	if e != nil {
		return domain.Revision{}, e
	}
	if o.Format() != domain.SHA1 {
		return domain.Revision{}, protocolError("provider object format unsupported")
	}
	return domain.NewRevision(repo, o)
}
