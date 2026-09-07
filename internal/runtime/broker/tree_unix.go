//go:build linux || darwin || freebsd

package broker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const maxOwnedSignalHelpers = 1024

type unixTree struct {
	executable      string
	session         uint64
	root            *exec.Cmd
	rootDone        chan struct{}
	rootErr         error
	helpers         []*acquiredSignal
	known           map[int]string
	escape          bool
	nativeQuiescent bool
	grace, force    time.Duration
}

func (t *unixTree) startRoot(cmd *exec.Cmd) error {
	if t.root != nil {
		return ErrProtocol
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	t.root = cmd
	t.rootDone = make(chan struct{})
	go func() { t.rootErr = cmd.Wait(); close(t.rootDone) }()
	return nil
}

func (t *unixTree) observe(ctx context.Context) ([]processFact, []int, error) {
	all, err := census(ctx)
	if err != nil {
		return nil, nil, err
	}
	if t.known == nil {
		t.known = make(map[int]string)
	}
	next := make(map[int]string)
	currentMembers := make(map[int]bool)
	for _, p := range all {
		if p.live && p.session == os.Getpid() {
			currentMembers[p.pid] = true
		}
	}
	for _, p := range all {
		if !p.live {
			continue
		}
		if p.session != os.Getpid() && currentMembers[p.parent] {
			t.escape = true
		}
		if identity, seen := t.known[p.pid]; seen && p.session != os.Getpid() && (identity == "" || identity == p.identity) {
			t.escape = true
		}
		if p.session == os.Getpid() {
			next[p.pid] = p.identity
		}
	}
	t.known = next
	members, groups, err := sessionMembers(all, os.Getpid())
	return members, groups, err
}

func (t *unixTree) pruneHelpers() error {
	var failure error
	retained := t.helpers[:0]
	for _, a := range t.helpers {
		select {
		case <-a.done:
			failure = errors.Join(failure, a.closeEndpoints())
			if a.waitErr != nil {
				var exit *exec.ExitError
				killed := errors.As(a.waitErr, &exit)
				if killed {
					status, ok := exit.Sys().(syscall.WaitStatus)
					killed = ok && status.Signaled() && status.Signal() == syscall.SIGKILL
				}
				if !killed || !a.committed || (a.prepare.signal != killGroup && a.prepare.signal != stopGroup) {
					failure = errors.Join(failure, a.waitErr)
				}
			}
		default:
			retained = append(retained, a)
		}
	}
	clear(t.helpers[len(retained):])
	t.helpers = retained
	return failure
}

func (t *unixTree) pulse(ctx context.Context, group int, kind groupSignal) error {
	prior := t.pruneHelpers()
	if len(t.helpers) >= maxOwnedSignalHelpers {
		return errors.Join(prior, ErrCensus)
	}
	a, err := acquireSignalGroup(ctx, t.executable, t.session, group, kind)
	if a != nil {
		t.helpers = append(t.helpers, a)
	}
	if err != nil {
		if a != nil {
			return errors.Join(prior, err, a.closeEndpoints())
		}
		return errors.Join(prior, err)
	}
	err = errors.Join(prior, a.commit(ctx))
	err = errors.Join(err, a.closeEndpoints())
	if kind != stopGroup && err == nil {
		err = a.join(ctx)
	}
	return err
}

func (t *unixTree) joined() bool {
	if t.root != nil {
		select {
		case <-t.rootDone:
		default:
			return false
		}
	}
	for _, a := range t.helpers {
		select {
		case <-a.done:
		default:
			return false
		}
	}
	return true
}

// cleanup uses no caller cancellation: once accepted, the owned tree remains
// this supervisor's responsibility. The deadlines bound one reporting attempt.
// A failed attempt retains live helpers/root/waiters for a later retry.
func (t *unixTree) cleanup() (bool, error) {
	if t.root == nil && len(t.helpers) == 0 {
		return true, nil
	} // no user/helper acquisition ever occurred
	graceDeadline := time.Now().Add(t.grace)
	deadline := graceDeadline.Add(t.force)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	termSent := map[int]bool{}
	var evidenceError error
	for {
		evidenceError = errors.Join(evidenceError, t.pruneHelpers())
		members, groups, observeErr := t.observe(ctx)
		if observeErr == nil && len(members) == 0 && t.joined() {
			t.nativeQuiescent = true
			if t.escape {
				return false, errors.Join(evidenceError, ErrCensus)
			}
			return true, evidenceError
		}
		if time.Now().After(deadline) {
			return false, errors.Join(evidenceError, observeErr, context.DeadlineExceeded)
		}
		if time.Now().Before(graceDeadline) {
			for _, group := range groups {
				if !termSent[group] {
					termSent[group] = true
					if err := t.pulse(ctx, group, termGroup); err != nil {
						evidenceError = err
					}
				}
			}
		} else {
			// Each pass acquires and parks a STOP helper in every currently
			// running group. Re-census must show every ordinary live member
			// stopped before any acquired KILL helper is committed.
			runningGroups := map[int]bool{}
			parked := map[int]bool{}
			for _, a := range t.helpers {
				if a.committed && a.prepare.signal == stopGroup {
					select {
					case <-a.done:
					default:
						parked[int(a.prepare.group)] = true
					}
				}
			}
			for _, p := range members {
				if p.group != os.Getpid() && !p.stopped {
					runningGroups[p.group] = true
				}
			}
			for _, group := range groups {
				if runningGroups[group] || !parked[group] {
					if err := t.pulse(ctx, group, stopGroup); err != nil {
						evidenceError = err
					}
				}
			}
			frozen, acquiredGroups, freezeErr := t.observe(ctx)
			allStopped := freezeErr == nil
			for _, p := range frozen {
				if !p.stopped {
					allStopped = false
				}
			}
			if allStopped {
				for _, group := range acquiredGroups {
					if err := t.pulse(ctx, group, killGroup); err != nil {
						evidenceError = err
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return false, errors.Join(evidenceError, observeErr, ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
}
