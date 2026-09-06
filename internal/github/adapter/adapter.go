// Package adapter implements the frozen Application remote ports using gh api.
// It performs no local Git operation and never imports the legacy client.
package adapter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// Config selects one authenticated gh host/profile for the adapter lifetime.
// Environment is copied at construction; nil snapshots the current environment.
// ConfigDirectory selects an existing gh authentication profile without editing it.
// Zero limits select the frozen defaults. Larger limits are refused.
type Config struct {
	Host, Executable, ConfigDirectory          string
	Environment                                []string
	ReadTimeout, MutationTimeout, DrainTimeout time.Duration
	StdoutLimit, StderrLimit                   int
}

type Adapter struct {
	config       Config
	run          commandRunner
	issuer       string
	sequence     atomic.Uint64
	mu           sync.Mutex
	repositories map[domain.RepositoryID]api.RemoteRepository
	cursors      map[string]continuation
}

func New(c Config) (*Adapter, error) {
	c.Host = strings.ToLower(c.Host)
	if _, err := api.NewRemoteRepositoryLocator(api.RemoteRepositoryLocatorData{Host: c.Host, Owner: "probe", Name: "probe"}); err != nil {
		return nil, errors.New("invalid authenticated host")
	}
	if c.Executable == "" {
		c.Executable = "gh"
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 30 * time.Second
	}
	if c.MutationTimeout == 0 {
		c.MutationTimeout = 120 * time.Second
	}
	if c.DrainTimeout == 0 {
		c.DrainTimeout = 2 * time.Second
	}
	if c.StdoutLimit == 0 {
		c.StdoutLimit = 16 << 20
	}
	if c.StderrLimit == 0 {
		c.StderrLimit = 256 << 10
	}
	if c.ReadTimeout <= 0 || c.ReadTimeout > 30*time.Second || c.MutationTimeout <= 0 || c.MutationTimeout > 120*time.Second || c.DrainTimeout <= 0 || c.DrainTimeout > 2*time.Second || c.StdoutLimit <= 0 || c.StdoutLimit > 16<<20 || c.StderrLimit <= 0 || c.StderrLimit > 256<<10 {
		return nil, errors.New("invalid command limits")
	}
	if c.Environment == nil {
		c.Environment = os.Environ()
	}
	c.Environment = append([]string(nil), c.Environment...)
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil, err
	}
	a := &Adapter{config: c, issuer: hex.EncodeToString(random[:]), repositories: make(map[domain.RepositoryID]api.RemoteRepository), cursors: make(map[string]continuation)}
	a.run = a.runCommand
	return a, nil
}

// ParseLocator accepts a fully specified HTTPS URL or owner/name on the supplied
// explicit host. Only one transport .git suffix is removed. Credentials and
// ambiguous URL forms are refused rather than repaired.
func ParseLocator(host, input string) (api.RemoteRepositoryLocator, error) {
	owner, name := "", ""
	if strings.HasPrefix(input, "https://") {
		u, e := url.Parse(input)
		if e != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || u.Port() != "" {
			return api.RemoteRepositoryLocator{}, errors.New("invalid repository URL")
		}
		if host != "" && !strings.EqualFold(host, u.Host) {
			return api.RemoteRepositoryLocator{}, errors.New("host mismatch")
		}
		host = u.Host
		input = strings.TrimPrefix(u.Path, "/")
	}
	p := strings.Split(input, "/")
	if len(p) == 2 {
		owner, name = p[0], strings.TrimSuffix(p[1], ".git")
	}
	return api.NewRemoteRepositoryLocator(api.RemoteRepositoryLocatorData{Host: strings.ToLower(host), Owner: strings.ToLower(owner), Name: strings.ToLower(name)})
}

func repositoryID(l api.RemoteRepositoryLocator) domain.RepositoryID {
	d := l.Data()
	return must(domain.NewRepositoryID(domain.Remote, d.Host+"/"+d.Owner+"/"+d.Name))
}
func repoPath(l api.RemoteRepositoryLocator) string {
	d := l.Data()
	return "repos/" + url.PathEscape(d.Owner) + "/" + url.PathEscape(d.Name)
}
func repoURL(l api.RemoteRepositoryLocator) string {
	d := l.Data()
	return "https://" + d.Host + "/" + d.Owner + "/" + d.Name
}
func (a *Adapter) lookup(id domain.RepositoryID) (api.RemoteRepository, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	r, ok := a.repositories[id]
	return r, ok
}
func (a *Adapter) register(r api.RemoteRepository) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.repositories[r.Data().ID]; !ok && len(a.repositories) >= 4096 {
		return errors.New("repository association limit")
	}
	a.repositories[r.Data().ID] = r
	return nil
}
func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func diagnostic(code api.ErrorCode, reason string) api.Diagnostic {
	return must(api.NewDiagnostic(api.DiagnosticData{Code: code, Reason: reason, Message: reason}))
}
func noTransport() api.CommandTransportOutcome {
	return must(api.NewCommandTransportOutcome(api.CommandTransportOutcomeData{CleanupKnown: true}))
}
func diagError(d api.Diagnostic) error { return errors.New(d.Data().Reason) }
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	return ctx.Err()
}
