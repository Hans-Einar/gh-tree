package app

import (
	"context"
	"fmt"

	"github.com/Hans-Einar/gh-tree/internal/launch"
)

func (s *Service) OpenTerminalConsole(ctx context.Context, path string, width, height int) (launch.ProcessSnapshot, error) {
	if s.Terminal == nil {
		return launch.ProcessSnapshot{}, fmt.Errorf("interactive terminal manager is unavailable")
	}
	if _, err := s.WorktreeStatus(ctx, path); err != nil {
		return launch.ProcessSnapshot{}, err
	}
	return s.Terminal.StartShell(path, width, height)
}

func (s *Service) WriteConsole(id int, data []byte) error {
	if s.Terminal == nil || !s.Terminal.IsTerminal(id) {
		return fmt.Errorf("console %d is not an interactive terminal", id)
	}
	return s.Terminal.Write(id, data)
}

func (s *Service) ResizeConsole(id, width, height int) error {
	if s.Terminal == nil || !s.Terminal.IsTerminal(id) {
		return nil
	}
	return s.Terminal.Resize(id, width, height)
}

func (s *Service) IsInteractiveConsole(id int) bool {
	return s.Terminal != nil && s.Terminal.IsTerminal(id)
}
