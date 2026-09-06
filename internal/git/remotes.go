package git

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

type configuredRemote struct {
	name  string
	fetch []string
}

func (s *readSession) remoteBindings(repo repository) ([]api.RemoteBinding, []api.Diagnostic, error) {
	c := s.command(repo.cwd(), "--git-dir="+repo.gitDir(), "config", "--null", "--list", "--show-origin", "--show-scope")
	if c.err != nil {
		return nil, nil, c.err
	}
	configuration := sourceVersion("configuration", repo.id.Token(), s.a.lifetime, c.stdout)
	q := s.command(repo.cwd(), "--git-dir="+repo.gitDir(), "config", "--null", "--get-regexp", `^remote\..*\.(url|pushurl|fetch)$`)
	if q.err != nil {
		if exit, _ := q.transport.Data().ExitCode.Value(); exit == 1 {
			return nil, nil, nil
		}
		return nil, nil, q.err
	}
	remoteByName := make(map[string]*configuredRemote)
	for _, entry := range bytes.Split(q.stdout, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		key, value, ok := bytes.Cut(entry, []byte{'\n'})
		if !ok {
			return nil, nil, diagnostic(api.Unavailable, "MalformedRemoteConfiguration", "Native remote configuration could not be decoded.")
		}
		keyName := strings.TrimPrefix(string(key), "remote.")
		i := strings.LastIndexByte(keyName, '.')
		if i <= 0 {
			return nil, nil, diagnostic(api.Unavailable, "MalformedRemoteConfiguration", "Native remote configuration has an invalid key.")
		}
		name, field := keyName[:i], keyName[i+1:]
		r := remoteByName[name]
		if r == nil {
			r = &configuredRemote{name: name}
			remoteByName[name] = r
		}
		if field == "fetch" {
			r.fetch = append(r.fetch, string(value))
		}
	}
	names := make([]string, 0, len(remoteByName))
	for name := range remoteByName {
		names = append(names, name)
	}
	sort.Strings(names)
	var bindings []api.RemoteBinding
	var diagnostics []api.Diagnostic
	for _, name := range names {
		r := remoteByName[name]
		fetch := s.command(repo.cwd(), "--git-dir="+repo.gitDir(), "remote", "get-url", "--all", "--", name)
		push := s.command(repo.cwd(), "--git-dir="+repo.gitDir(), "remote", "get-url", "--push", "--all", "--", name)
		if fetch.err != nil || push.err != nil {
			diagnostics = append(diagnostics, diagnostic(api.Unavailable, "RemoteMappingUnavailable", "A configured remote's effective native transport mapping is unavailable."))
			continue
		}
		var id domain.RepositoryID
		var fetchURLs, pushURLs []string
		var mappings []api.RefspecMapping
		var invalidMapping bool
		for group, raw := range [][]byte{fetch.stdout, push.stdout} {
			for _, locator := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
				remote, sanitized, err := remoteIdentity(repo, locator)
				if err != nil || id.Valid() && id != remote {
					invalidMapping = true
					break
				}
				id = remote
				if group == 0 {
					fetchURLs = append(fetchURLs, sanitized)
				} else {
					pushURLs = append(pushURLs, sanitized)
				}
			}
		}
		for _, raw := range r.fetch {
			force := strings.HasPrefix(raw, "+")
			raw = strings.TrimPrefix(raw, "+")
			source, destination, ok := strings.Cut(raw, ":")
			// Negative or multiple-destination mappings need explicit support;
			// they cannot become an invented positive transport association.
			if !ok || strings.HasPrefix(source, "^") || strings.Contains(destination, ":") {
				invalidMapping = true
				break
			}
			mapping, err := api.NewRefspecMapping(api.RefspecMappingData{Source: source, Destination: destination, Force: force})
			if err != nil {
				invalidMapping = true
				break
			}
			mappings = append(mappings, mapping)
		}
		if invalidMapping || !id.Valid() {
			diagnostics = append(diagnostics, diagnostic(api.Unsupported, "UnsupportedRemoteMapping", "A configured remote has ambiguous or unsupported scope/refspec mapping."))
			continue
		}
		binding, err := api.NewRemoteBinding(api.RemoteBindingData{LocalRepository: repo.id, RemoteRepository: id, RemoteName: name, FetchURLs: fetchURLs, PushURLs: pushURLs, Refspecs: mappings, Configuration: configuration})
		if err != nil {
			return bindings, diagnostics, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, diagnostics, nil
}

func remoteIdentity(repo repository, locator string) (domain.RepositoryID, string, error) {
	bad := func() (domain.RepositoryID, string, error) {
		return domain.RepositoryID{}, "", diagnostic(api.Unsupported, "UnsupportedRemoteScope", "The native remote locator cannot be associated with an explicit supported repository scope.")
	}
	if locator == "" || strings.ContainsAny(locator, "\x00\r\n?#") {
		return bad()
	}
	var host, path string
	if strings.Contains(locator, "://") {
		u, err := url.Parse(locator)
		if err != nil {
			return bad()
		}
		if u.Scheme == "file" {
			if u.Host != "" && u.Host != "localhost" {
				return bad()
			}
			p, err := url.PathUnescape(u.Path)
			if err != nil {
				return bad()
			}
			if len(p) > 2 && p[0] == '/' && p[2] == ':' {
				p = p[1:]
			}
			return localRemoteIdentity(repo, p)
		}
		if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "ssh" && u.Scheme != "git" {
			return bad()
		}
		if u.Port() != "" {
			return bad()
		} // Distinct ports require an explicit transport profile.
		host = u.Hostname()
		path = strings.TrimPrefix(u.EscapedPath(), "/")
	} else if colon := strings.IndexByte(locator, ':'); colon > 0 && !filepath.IsAbs(locator) && !(colon == 1 && len(locator) > 2) {
		host = locator[:colon]
		if at := strings.LastIndexByte(host, '@'); at >= 0 {
			host = host[at+1:]
		}
		path = locator[colon+1:]
	} else {
		return localRemoteIdentity(repo, locator)
	}
	parts := strings.Split(path, "/")
	if host == "" || len(parts) != 2 || strings.ContainsAny(host, "/@\\ ") || strings.ContainsAny(path, "%\\") {
		return bad()
	}
	owner := strings.ToLower(parts[0])
	name := strings.ToLower(strings.TrimSuffix(parts[1], ".git"))
	host = strings.ToLower(host)
	if owner == "" || name == "" || owner == "." || owner == ".." || name == "." || name == ".." {
		return bad()
	}
	if _, err := api.NewRemoteRepositoryLocator(api.RemoteRepositoryLocatorData{Host: host, Owner: owner, Name: name}); err != nil {
		return bad()
	}
	id, err := domain.NewRepositoryID(domain.Remote, host+"/"+owner+"/"+name)
	return id, "https://" + host + "/" + owner + "/" + name, err
}

func localRemoteIdentity(repo repository, path string) (domain.RepositoryID, string, error) {
	if !filepath.IsAbs(path) {
		return domain.RepositoryID{}, "", diagnostic(api.Unsupported, "RelativeRemoteScope", "A relative filesystem remote needs an explicit stable transport base.")
	}
	d, err := observeDirectory(path)
	if err != nil {
		return domain.RepositoryID{}, "", diagnostic(api.Unavailable, "LocalRemoteUnavailable", "The configured local remote directory cannot be inspected.")
	}
	// File transports are separately scoped native repositories; they never
	// masquerade as a GitHub host or as their caller's LocalCommon repository.
	id, err := domain.NewRepositoryID(domain.Remote, fmt.Sprintf("file:%s:%d:%x", d.path, d.identity.Device(), d.identity.FileID()))
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(d.path)}
	return id, u.String(), err
}
