package git

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

const scopeLockName = "gh-tree-mutation.lock"
const scopeLockMarker = "gh-tree common mutation coordination v1\n"

// mutationScheduler reserves bounded waiting/active calls, never goroutines or
// native resources. One adapter serializes every linked worktree by LocalCommon.
// Independent adapters/processes additionally acquire nativeScopeGuard below.
type mutationScheduler struct {
	mu       sync.Mutex
	scopes   map[domain.RepositoryID]*scheduledScope
	admitted int
}

type scheduledScope struct {
	turn  chan struct{}
	users int
}

func (s *mutationScheduler) acquire(ctx context.Context, repository domain.RepositoryID) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	if s.admitted >= 64 {
		s.mu.Unlock()
		return nil, diagnostic(api.Busy, "MutationAdmissionFull", "The bounded Git mutation scheduler is full.")
	}
	if s.scopes == nil {
		s.scopes = make(map[domain.RepositoryID]*scheduledScope)
	}
	entry := s.scopes[repository]
	if entry == nil {
		entry = &scheduledScope{turn: make(chan struct{}, 1)}
		entry.turn <- struct{}{}
		s.scopes[repository] = entry
	}
	entry.users++
	s.admitted++
	s.mu.Unlock()
	forget := func() {
		s.mu.Lock()
		entry.users--
		s.admitted--
		if entry.users == 0 {
			delete(s.scopes, repository)
		}
		s.mu.Unlock()
	}
	select {
	case <-ctx.Done():
		forget()
		return nil, ctx.Err()
	case <-entry.turn:
	}
	// An available token and canceled context can become ready together.
	if err := ctx.Err(); err != nil {
		entry.turn <- struct{}{}
		forget()
		return nil, err
	}
	var once sync.Once
	return func() { once.Do(func() { entry.turn <- struct{}{}; forget() }) }, nil
}

// A permanent advisory-lock file is never unlinked or truncated. Closing this
// operation's handle releases only its OS lock, including after process death.
// It is not a native index/ref lock and no nested Git command adopts it.
type nativeScopeGuard struct {
	directory       *nativeDirectory
	file            *os.File
	created         bool
	releaseSchedule func()
	once            sync.Once
	closeError      error
}

// A nonnil guard accompanying an error is already closed, but retains created
// and closeError facts: a canceled/failed acquisition can have created the
// permanent coordination object. Mutation result assembly must retain them.
func (a *Adapter) acquireMutationScope(ctx context.Context, r repository) (*nativeScopeGuard, error) {
	release, err := a.mutations.acquire(ctx, r.id)
	if err != nil {
		return nil, err
	}
	g := &nativeScopeGuard{releaseSchedule: release}
	ok := false
	defer func() {
		if !ok {
			g.close()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	g.directory, err = acquireDirectory(r.common)
	if err != nil {
		return nil, err
	}
	g.file, g.created, err = g.directory.openScopeLock(scopeLockName)
	if err != nil {
		return nil, err
	}
	if err = lockScopeFile(g.file); err != nil {
		if scopeLockContention(err) {
			return g, diagnostic(api.Busy, "CommonMutationLocked", "Another process owns the common-repository Git mutation guard.")
		}
		return g, err
	}
	if g.created {
		if _, err = io.WriteString(g.file, scopeLockMarker); err == nil {
			err = g.file.Sync()
		}
		if err != nil {
			return g, err
		}
	} else {
		marker, e := io.ReadAll(io.LimitReader(g.file, int64(len(scopeLockMarker)+1)))
		if e != nil {
			return nil, e
		}
		if string(marker) != scopeLockMarker {
			return nil, diagnostic(api.Conflict, "UnrecognizedCommonGuard", "The existing coordination file is incomplete or belongs to an unsupported protocol; it is preserved.")
		}
	}
	if err = g.validate(); err != nil {
		return g, err
	}
	if err = ctx.Err(); err != nil {
		return g, err
	}
	ok = true
	return g, nil
}

func (g *nativeScopeGuard) validate() error {
	// A cooperating process always locks this same permanent object. Detect a
	// replaced named entry; never adopt it or delete either object as cleanup.
	actual, err := g.directory.openRegular(scopeLockName)
	if err != nil {
		return err
	}
	defer actual.Close()
	heldInfo, err := g.file.Stat()
	if err != nil {
		return err
	}
	namedInfo, err := actual.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(heldInfo, namedInfo) {
		return diagnostic(api.Indeterminate, "CommonGuardReplaced", "The named common guard differs from the locked object; both are preserved.")
	}
	return nil
}

func (g *nativeScopeGuard) close() error {
	g.once.Do(func() {
		if g.file != nil {
			g.closeError = errors.Join(g.closeError, g.file.Close())
		}
		if g.directory != nil {
			g.closeError = errors.Join(g.closeError, g.directory.close())
		}
		if g.releaseSchedule != nil {
			g.releaseSchedule()
		}
	})
	return g.closeError
}
