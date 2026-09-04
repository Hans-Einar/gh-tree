package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/xpty"

	"github.com/Hans-Einar/gh-tree/internal/launch"
)

const terminalIDBase = 1000

// Manager owns interactive PTY/ConPTY shell sessions. Terminal ids live in a
// separate range from launch-console ids so both can share the cockpit tab bar.
type Manager struct {
	mu       sync.Mutex
	nextID   int
	sessions map[int]*session
	maxBytes int
}

type session struct {
	mu      sync.Mutex
	id      int
	shell   Shell
	dir     string
	pty     xpty.Pty
	cmd     *exec.Cmd
	state   launch.ProcessState
	started time.Time
	ended   time.Time
	exit    int
	stopped bool
	buf     *byteRing
}

func NewManager(maxBytes int) *Manager {
	if maxBytes <= 0 {
		maxBytes = 256 * 1024
	}
	return &Manager{nextID: terminalIDBase, sessions: map[int]*session{}, maxBytes: maxBytes}
}

func (m *Manager) StartShell(dir string, width, height int) (launch.ProcessSnapshot, error) {
	sh, err := DetectShell()
	if err != nil {
		return launch.ProcessSnapshot{}, err
	}
	return m.StartShellWith(dir, width, height, sh)
}

func (m *Manager) StartShellWith(dir string, width, height int, sh Shell) (launch.ProcessSnapshot, error) {
	if strings.TrimSpace(dir) == "" {
		return launch.ProcessSnapshot{}, fmt.Errorf("terminal working directory is empty")
	}
	if width < 20 {
		width = 80
	}
	if height < 4 {
		height = 24
	}
	p, err := xpty.NewPty(width, height)
	if err != nil {
		return launch.ProcessSnapshot{}, fmt.Errorf("open PTY: %w", err)
	}
	cmd := exec.Command(sh.Path, sh.Args...)
	cmd.Dir = dir
	cmd.Env = terminalEnv()

	m.mu.Lock()
	m.nextID++
	id := m.nextID
	s := &session{id: id, shell: sh, dir: dir, pty: p, cmd: cmd, state: launch.StateStarting, started: time.Now(), exit: -1, buf: newByteRing(m.maxBytes)}
	m.sessions[id] = s
	m.mu.Unlock()

	if err := p.Start(cmd); err != nil {
		_ = p.Close()
		s.mu.Lock()
		s.state = launch.StateFailed
		s.ended = time.Now()
		s.mu.Unlock()
		return s.snapshot(), fmt.Errorf("start %s terminal: %w", sh.Name, err)
	}
	s.mu.Lock()
	s.state = launch.StateRunning
	s.mu.Unlock()
	go s.readLoop()
	go s.waitLoop()
	return s.snapshot(), nil
}

func terminalEnv() []string {
	env := append([]string(nil), os.Environ()...)
	env = setEnv(env, "TERM", "xterm-256color")
	env = setEnv(env, "COLORTERM", "truecolor")
	return env
}

func setEnv(env []string, key, value string) []string {
	prefix := strings.ToUpper(key) + "="
	for i, item := range env {
		if strings.HasPrefix(strings.ToUpper(item), prefix) {
			env[i] = key + "=" + value
			return env
		}
	}
	return append(env, key+"="+value)
}

func (s *session) readLoop() {
	buf := make([]byte, 8192)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			s.buf.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func (s *session) waitLoop() {
	err := xpty.WaitProcess(context.Background(), s.cmd)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ended = time.Now()
	if s.cmd != nil && s.cmd.ProcessState != nil {
		s.exit = s.cmd.ProcessState.ExitCode()
	}
	if s.stopped {
		s.state = launch.StateStopped
		return
	}
	if err != nil {
		s.state = launch.StateFailed
		return
	}
	s.state = launch.StateExited
	if s.exit < 0 {
		s.exit = 0
	}
}

func (m *Manager) Write(id int, data []byte) error {
	s, ok := m.get(id)
	if !ok {
		return fmt.Errorf("terminal %d not found", id)
	}
	s.mu.Lock()
	state := s.state
	p := s.pty
	s.mu.Unlock()
	if state != launch.StateRunning && state != launch.StateStarting {
		return fmt.Errorf("terminal %d is not running", id)
	}
	_, err := p.Write(data)
	if err != nil {
		return fmt.Errorf("write terminal %d: %w", id, err)
	}
	return nil
}

func (m *Manager) Resize(id, width, height int) error {
	s, ok := m.get(id)
	if !ok {
		return nil
	}
	if width < 2 || height < 2 {
		return nil
	}
	if err := s.pty.Resize(width, height); err != nil {
		return fmt.Errorf("resize terminal %d: %w", id, err)
	}
	return nil
}

func (m *Manager) Stop(id int) error {
	s, ok := m.get(id)
	if !ok {
		return nil
	}
	s.mu.Lock()
	if s.state != launch.StateRunning && s.state != launch.StateStarting {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	s.state = launch.StateStopped
	proc := s.cmd.Process
	p := s.pty
	s.mu.Unlock()
	_ = p.Close()
	if proc != nil {
		_ = proc.Kill()
	}
	return nil
}

func (m *Manager) Restart(id int, width, height int) (launch.ProcessSnapshot, error) {
	s, ok := m.get(id)
	if !ok {
		return launch.ProcessSnapshot{}, fmt.Errorf("terminal %d not found", id)
	}
	s.mu.Lock()
	dir := s.dir
	sh := s.shell
	s.mu.Unlock()
	if err := m.Stop(id); err != nil {
		return launch.ProcessSnapshot{}, err
	}
	return m.StartShellWith(dir, width, height, sh)
}

func (m *Manager) Snapshot(id int) (launch.ProcessSnapshot, bool) {
	s, ok := m.get(id)
	if !ok {
		return launch.ProcessSnapshot{}, false
	}
	return s.snapshot(), true
}

func (m *Manager) Snapshots() []launch.ProcessSnapshot {
	m.mu.Lock()
	ids := make([]int, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	sort.Ints(ids)
	out := make([]launch.ProcessSnapshot, 0, len(ids))
	for _, id := range ids {
		if snap, ok := m.Snapshot(id); ok {
			out = append(out, snap)
		}
	}
	return out
}

func (m *Manager) StopAll() error {
	m.mu.Lock()
	ids := make([]int, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	var first error
	for _, id := range ids {
		if err := m.Stop(id); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m *Manager) IsTerminal(id int) bool {
	_, ok := m.get(id)
	return ok
}

func (m *Manager) get(id int) (*session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (s *session) snapshot() launch.ProcessSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	pid := 0
	if s.cmd != nil && s.cmd.Process != nil {
		pid = s.cmd.Process.Pid
	}
	return launch.ProcessSnapshot{
		ID: s.id,
		Invocation: launch.Invocation{Provider: "terminal", Name: s.shell.Name, Command: s.shell.Path, Args: append([]string(nil), s.shell.Args...), Dir: s.dir},
		State: s.state,
		PID: pid,
		ExitCode: s.exit,
		Started: s.started,
		Ended: s.ended,
		Lines: terminalLines(s.buf.Bytes()),
	}
}

type byteRing struct {
	mu  sync.Mutex
	max int
	buf []byte
}

func newByteRing(max int) *byteRing { return &byteRing{max: max} }
func (b *byteRing) Write(p []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	if len(b.buf) > b.max {
		copy(b.buf, b.buf[len(b.buf)-b.max:])
		b.buf = b.buf[:b.max]
	}
}
func (b *byteRing) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.buf...)
}

func terminalLines(raw []byte) []string {
	text := ansi.Strip(string(raw))
	var lines []string
	line := []rune{}
	cr := false
	for _, r := range text {
		switch r {
		case '\r':
			cr = true
		case '\n':
			lines = append(lines, string(line))
			line = line[:0]
			cr = false
		case '\b':
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		default:
			if r < 0x20 && r != '\t' {
				continue
			}
			if cr {
				line = line[:0]
				cr = false
			}
			line = append(line, r)
		}
	}
	if len(line) > 0 {
		lines = append(lines, string(line))
	}
	if len(lines) > 500 {
		lines = append([]string(nil), lines[len(lines)-500:]...)
	}
	return lines
}
