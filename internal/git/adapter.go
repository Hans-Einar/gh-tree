// Package git implements local Git observations and guarded native mechanics.
// It owns its command transport; it does not use the legacy implementation.
package git

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// Options fixes transport and admission bounds for one adapter lifetime.
// Environment is copied; nil snapshots the process environment. CurrentDirectory
// identifies the launch context, independent of later ResolveLocal calls.
type Options struct {
	GitExecutable    string
	CurrentDirectory string
	Environment      []string
	ReadTimeout      time.Duration
	MutationTimeout  time.Duration
	DrainTimeout     time.Duration
	MaxStdoutBytes   int
	MaxStderrBytes   int
	MaxRepositories  int
}

// Adapter retains only copied identity/locator observations between calls.
// No constructor or read obtains native mutation locks.
type Adapter struct {
	options      Options
	lifetime     string
	current      directoryObservation
	mu           sync.Mutex
	repositories map[domain.RepositoryID]repository
}

type repository struct {
	id           domain.RepositoryID
	common       directoryObservation
	format       domain.ObjectFormat
	version      string
	backend      api.RefBackend
	contextAdmin string
	contextRoot  string
}

func (r repository) gitDir() string {
	if r.contextAdmin != "" {
		return r.contextAdmin
	}
	return r.common.path
}
func (r repository) cwd() string {
	if r.contextRoot != "" {
		return r.contextRoot
	}
	return r.common.path
}

type directoryObservation struct {
	path     string
	identity api.DirectoryIdentity
}

// New snapshots construction context without observing or changing a repository.
func New(options Options) (*Adapter, error) {
	if options.GitExecutable == "" {
		options.GitExecutable = "git"
	}
	executable, err := exec.LookPath(options.GitExecutable)
	if err != nil {
		return nil, diagnostic(api.NotFound, "GitExecutableUnavailable", "The configured Git executable is unavailable.")
	}
	options.GitExecutable, err = filepath.Abs(executable)
	if err != nil {
		return nil, diagnostic(api.Invalid, "GitExecutableInvalid", "The configured Git executable has no absolute locator.")
	}
	if options.CurrentDirectory == "" {
		options.CurrentDirectory, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	current, err := observeDirectory(options.CurrentDirectory)
	if err != nil {
		return nil, diagnostic(api.Unavailable, "CurrentDirectoryUnavailable", "The adapter launch directory could not be observed.")
	}
	if options.Environment == nil {
		options.Environment = os.Environ()
	}
	options.Environment = append([]string(nil), options.Environment...)
	for _, entry := range options.Environment {
		name, _, ok := strings.Cut(entry, "=")
		if !ok || name == "" || strings.ContainsRune(entry, 0) {
			return nil, diagnostic(api.Invalid, "InvalidEnvironment", "An environment entry is invalid.")
		}
	}
	if options.ReadTimeout == 0 {
		options.ReadTimeout = 30 * time.Second
	}
	if options.MutationTimeout == 0 {
		options.MutationTimeout = 120 * time.Second
	}
	if options.DrainTimeout == 0 {
		options.DrainTimeout = 2 * time.Second
	}
	if options.MaxStdoutBytes == 0 {
		options.MaxStdoutBytes = 16 << 20
	}
	if options.MaxStderrBytes == 0 {
		options.MaxStderrBytes = 256 << 10
	}
	if options.MaxRepositories == 0 {
		options.MaxRepositories = 64
	}
	if options.ReadTimeout < 0 || options.MutationTimeout < 0 || options.DrainTimeout < 0 || options.MaxStdoutBytes < 1 || options.MaxStdoutBytes > 16<<20 || options.MaxStderrBytes < 1 || options.MaxStderrBytes > 256<<10 || options.MaxRepositories < 1 || options.MaxRepositories > 64 {
		return nil, diagnostic(api.Invalid, "InvalidAdapterBounds", "Git adapter construction bounds are invalid.")
	}
	var nonce [32]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	return &Adapter{options: options, lifetime: hex.EncodeToString(nonce[:]), current: current, repositories: make(map[domain.RepositoryID]repository)}, nil
}

func diagnostic(code api.ErrorCode, reason, message string) api.Diagnostic {
	d, err := api.NewDiagnostic(api.DiagnosticData{Code: code, Reason: reason, Message: message})
	if err != nil {
		panic(err)
	} // Constant programmer-owned vocabulary only.
	return d
}

func sourceVersion(namespace, scope, issuer string, bytes []byte) api.SourceVersion {
	h := sha256.Sum256(bytes)
	v, err := api.NewSourceVersion(namespace, scope, issuer, hex.EncodeToString(h[:]))
	if err != nil {
		panic(err)
	}
	return v
}

func (a *Adapter) registered(ctx context.Context, id domain.RepositoryID) (repository, error) {
	if err := ctx.Err(); err != nil {
		return repository{}, err
	}
	a.mu.Lock()
	r, ok := a.repositories[id]
	a.mu.Unlock()
	if !ok {
		return repository{}, diagnostic(api.NotFound, "RepositoryNotRegistered", "Resolve the local repository with this adapter before using its identity.")
	}
	actual, err := observeDirectory(r.common.path)
	if err != nil || !sameDirectoryObject(actual, r.common) {
		return repository{}, diagnostic(api.StaleObservation, "CommonDirectoryChanged", "The registered common directory no longer matches its physical observation.")
	}
	return r, nil
}

func sameDirectoryObject(a, b directoryObservation) bool {
	return a.path == b.path && a.identity.Platform() == b.identity.Platform() && a.identity.Device() == b.identity.Device() && a.identity.FileID() == b.identity.FileID()
}

func safeError(err error) api.Diagnostic {
	var d api.Diagnostic
	if errors.As(err, &d) {
		return d
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return diagnostic(api.Canceled, "CommandCanceled", "Git command cancellation was requested.")
	}
	return diagnostic(api.IOFailure, "ObservationFailed", "The local Git observation could not be completed.")
}
