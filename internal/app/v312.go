package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/Hans-Einar/gh-tree/internal/launch"
)

// RunCandidateConsole starts one independent launch session and returns its
// stable console id. Existing RunCandidate remains compatible and simply
// selects the newest session through SessionManager's current-session API.
func (s *Service) RunCandidateConsole(ctx context.Context, path string, c launch.Candidate) (launch.ProcessSnapshot, error) {
	if s.Launcher == nil {
		return launch.ProcessSnapshot{}, fmt.Errorf("launch manager is unavailable")
	}
	if _, err := s.WorktreeStatus(ctx, path); err != nil {
		return launch.ProcessSnapshot{}, err
	}
	inv, err := s.LaunchRegistry.Build(path, c)
	if err != nil {
		return launch.ProcessSnapshot{}, err
	}
	id, err := s.Launcher.StartSession(inv)
	if err != nil {
		if snap, ok := s.Launcher.SnapshotSession(id); ok {
			return snap, err
		}
		return launch.ProcessSnapshot{}, err
	}
	snap, _ := s.Launcher.SnapshotSession(id)
	return snap, nil
}

func (s *Service) RunDefaultConsole(ctx context.Context, path string) (launch.ProcessSnapshot, error) {
	if s.Launcher == nil {
		return launch.ProcessSnapshot{}, fmt.Errorf("launch manager is unavailable")
	}
	if _, err := s.WorktreeStatus(ctx, path); err != nil {
		return launch.ProcessSnapshot{}, err
	}
	cfg, err := launch.LoadConfig(path)
	if err != nil {
		return launch.ProcessSnapshot{}, err
	}
	if cfg.Default == "" {
		return launch.ProcessSnapshot{}, fmt.Errorf("no default launch point; use Ctrl+F5 or Alt+L to choose one")
	}
	inv, err := cfg.Invocation(path, cfg.Default, s.LaunchRegistry)
	if err != nil {
		return launch.ProcessSnapshot{}, err
	}
	id, err := s.Launcher.StartSession(inv)
	if err != nil {
		if snap, ok := s.Launcher.SnapshotSession(id); ok {
			return snap, err
		}
		return launch.ProcessSnapshot{}, err
	}
	snap, _ := s.Launcher.SnapshotSession(id)
	return snap, nil
}

// ConsoleSnapshots returns launch and PTY-backed interactive shell consoles in
// one stable id namespace for the cockpit tab bar.
func (s *Service) ConsoleSnapshots() []launch.ProcessSnapshot {
	var out []launch.ProcessSnapshot
	if s.Launcher != nil {
		out = append(out, s.Launcher.Snapshots()...)
	}
	if s.Terminal != nil {
		out = append(out, s.Terminal.Snapshots()...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Service) StopConsole(id int) error {
	if s.Terminal != nil && s.Terminal.IsTerminal(id) {
		return s.Terminal.Stop(id)
	}
	if s.Launcher == nil {
		return nil
	}
	return s.Launcher.StopSession(id)
}

func (s *Service) RestartConsole(id int) (launch.ProcessSnapshot, error) {
	if s.Terminal != nil && s.Terminal.IsTerminal(id) {
		return s.Terminal.Restart(id, 80, 24)
	}
	if s.Launcher == nil {
		return launch.ProcessSnapshot{}, fmt.Errorf("launch manager is unavailable")
	}
	newID, err := s.Launcher.RestartSession(id)
	if err != nil {
		return launch.ProcessSnapshot{}, err
	}
	snap, _ := s.Launcher.SnapshotSession(newID)
	return snap, nil
}

func (s *Service) SelectConsole(id int) bool {
	if s.Terminal != nil && s.Terminal.IsTerminal(id) {
		return true
	}
	if s.Launcher == nil {
		return false
	}
	return s.Launcher.SetCurrent(id)
}

func (s *Service) StopAllConsoles() error {
	var first error
	if s.Launcher != nil {
		if err := s.Launcher.StopAll(); err != nil {
			first = err
		}
	}
	if s.Terminal != nil {
		if err := s.Terminal.StopAll(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
