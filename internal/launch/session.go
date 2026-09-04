package launch

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"sync"
	"time"
)

type ProcessState string

const (
	StateStarting ProcessState = "starting"
	StateRunning  ProcessState = "running"
	StateExited   ProcessState = "exited"
	StateFailed   ProcessState = "failed"
	StateStopped  ProcessState = "stopped"
)

type ProcessSnapshot struct {
	ID         int
	Invocation Invocation
	State      ProcessState
	PID        int
	ExitCode   int
	Started    time.Time
	Ended      time.Time
	Lines      []string
}

type SessionManager struct {
	mu        sync.Mutex
	sessions  map[int]*session
	order     []int
	currentID int
	nextID    int
	maxLines  int
}

type session struct {
	mu         sync.Mutex
	id         int
	invocation Invocation
	cmd        *exec.Cmd
	state      ProcessState
	exitCode   int
	started    time.Time
	ended      time.Time
	logs       *lineBuffer
	done       chan struct{}
}

func NewSessionManager(maxLines int) *SessionManager {
	if maxLines <= 0 {
		maxLines = 500
	}
	return &SessionManager{sessions: map[int]*session{}, nextID: 1, maxLines: maxLines}
}

// Start preserves the original API while now opening a new independent session.
func (m *SessionManager) Start(inv Invocation) error {
	_, err := m.StartSession(inv)
	return err
}

func (m *SessionManager) StartSession(inv Invocation) (int, error) {
	if inv.Command == "" {
		return 0, fmt.Errorf("launch command is empty")
	}
	m.mu.Lock()
	id := m.nextID
	m.nextID++
	logs := newLineBuffer(m.maxLines)
	cmd := exec.Command(inv.Command, inv.Args...)
	cmd.Dir = inv.Dir
	cmd.Stdout = logs
	cmd.Stderr = logs
	s := &session{id: id, invocation: inv, cmd: cmd, state: StateStarting, exitCode: -1, started: time.Now(), logs: logs, done: make(chan struct{})}
	m.sessions[id] = s
	m.order = append(m.order, id)
	m.currentID = id
	m.mu.Unlock()

	if err := cmd.Start(); err != nil {
		s.mu.Lock()
		s.state = StateFailed
		s.ended = time.Now()
		s.mu.Unlock()
		close(s.done)
		return id, fmt.Errorf("start launch point: %w", err)
	}
	s.mu.Lock()
	s.state = StateRunning
	s.mu.Unlock()
	go s.wait()
	return id, nil
}

func (s *session) wait() {
	err := s.cmd.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ended = time.Now()
	if err == nil {
		s.state = StateExited
		s.exitCode = 0
	} else if ee, ok := err.(*exec.ExitError); ok {
		s.exitCode = ee.ExitCode()
		if s.state != StateStopped {
			s.state = StateFailed
		}
	} else {
		s.exitCode = -1
		if s.state != StateStopped {
			s.state = StateFailed
		}
	}
	close(s.done)
}

func (m *SessionManager) Stop() error {
	m.mu.Lock()
	id := m.currentID
	m.mu.Unlock()
	if id == 0 {
		return nil
	}
	return m.StopSession(id)
}

func (m *SessionManager) StopSession(id int) error {
	m.mu.Lock()
	s := m.sessions[id]
	m.mu.Unlock()
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.state != StateRunning && s.state != StateStarting {
		s.mu.Unlock()
		return nil
	}
	s.state = StateStopped
	proc := s.cmd.Process
	s.mu.Unlock()
	if proc == nil {
		return nil
	}
	if runtime.GOOS != "windows" {
		_ = proc.Signal(os.Interrupt)
		select {
		case <-s.done:
			return nil
		case <-time.After(1500 * time.Millisecond):
		}
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("stop launch process: %w", err)
	}
	select {
	case <-s.done:
	case <-time.After(2 * time.Second):
		return fmt.Errorf("launch process did not stop")
	}
	return nil
}

func (m *SessionManager) Restart() error {
	m.mu.Lock()
	id := m.currentID
	m.mu.Unlock()
	if id == 0 {
		return fmt.Errorf("no launch process to restart")
	}
	_, err := m.RestartSession(id)
	return err
}

func (m *SessionManager) RestartSession(id int) (int, error) {
	snap, ok := m.SnapshotSession(id)
	if !ok {
		return 0, fmt.Errorf("no launch process to restart")
	}
	if err := m.StopSession(id); err != nil {
		return 0, err
	}
	return m.StartSession(snap.Invocation)
}

func (m *SessionManager) Snapshot() ProcessSnapshot {
	m.mu.Lock()
	id := m.currentID
	m.mu.Unlock()
	if id == 0 {
		return ProcessSnapshot{State: StateExited, ExitCode: -1}
	}
	snap, ok := m.SnapshotSession(id)
	if !ok {
		return ProcessSnapshot{State: StateExited, ExitCode: -1}
	}
	return snap
}

func (m *SessionManager) SnapshotSession(id int) (ProcessSnapshot, bool) {
	m.mu.Lock()
	s := m.sessions[id]
	m.mu.Unlock()
	if s == nil {
		return ProcessSnapshot{}, false
	}
	return s.snapshot(), true
}

func (m *SessionManager) Snapshots() []ProcessSnapshot {
	m.mu.Lock()
	ids := append([]int(nil), m.order...)
	sessions := make(map[int]*session, len(m.sessions))
	for id, s := range m.sessions {
		sessions[id] = s
	}
	m.mu.Unlock()
	out := make([]ProcessSnapshot, 0, len(ids))
	for _, id := range ids {
		if s := sessions[id]; s != nil {
			out = append(out, s.snapshot())
		}
	}
	return out
}

func (m *SessionManager) SetCurrent(id int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[id]; !ok {
		return false
	}
	m.currentID = id
	return true
}

func (m *SessionManager) CurrentID() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentID
}

func (m *SessionManager) StopAll() error {
	m.mu.Lock()
	ids := append([]int(nil), m.order...)
	m.mu.Unlock()
	var first error
	for _, id := range ids {
		if err := m.StopSession(id); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m *SessionManager) Wait(timeout time.Duration) (ProcessSnapshot, error) {
	m.mu.Lock()
	id := m.currentID
	m.mu.Unlock()
	if id == 0 {
		return ProcessSnapshot{}, fmt.Errorf("no launch process")
	}
	return m.WaitSession(id, timeout)
}

func (m *SessionManager) WaitSession(id int, timeout time.Duration) (ProcessSnapshot, error) {
	m.mu.Lock()
	s := m.sessions[id]
	m.mu.Unlock()
	if s == nil {
		return ProcessSnapshot{}, fmt.Errorf("no launch process")
	}
	select {
	case <-s.done:
		return s.snapshot(), nil
	case <-time.After(timeout):
		return s.snapshot(), fmt.Errorf("timed out waiting for launch process")
	}
}

func (s *session) snapshot() ProcessSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	pid := 0
	if s.cmd != nil && s.cmd.Process != nil {
		pid = s.cmd.Process.Pid
	}
	return ProcessSnapshot{ID: s.id, Invocation: s.invocation, State: s.state, PID: pid, ExitCode: s.exitCode, Started: s.started, Ended: s.ended, Lines: s.logs.lines()}
}

// SortedSnapshots is useful to callers that need a stable ID order independent
// of creation-order bookkeeping.
func SortedSnapshots(items []ProcessSnapshot) []ProcessSnapshot {
	out := append([]ProcessSnapshot(nil), items...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

type lineBuffer struct {
	mu      sync.Mutex
	max     int
	pending bytes.Buffer
	data    []string
}

func newLineBuffer(max int) *lineBuffer { return &lineBuffer{max: max} }
func (b *lineBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	_, _ = b.pending.Write(p)
	scanner := bufio.NewScanner(bytes.NewReader(b.pending.Bytes()))
	var complete []string
	for scanner.Scan() {
		complete = append(complete, scanner.Text())
	}
	endsNewline := b.pending.Len() > 0 && (b.pending.Bytes()[b.pending.Len()-1] == '\n')
	b.pending.Reset()
	if len(complete) > 0 {
		last := len(complete)
		if !endsNewline {
			last--
			if last >= 0 {
				_, _ = io.WriteString(&b.pending, complete[len(complete)-1])
			}
		}
		for i := 0; i < last; i++ {
			b.data = append(b.data, complete[i])
		}
	}
	if len(b.data) > b.max {
		b.data = append([]string(nil), b.data[len(b.data)-b.max:]...)
	}
	return n, nil
}
func (b *lineBuffer) lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := append([]string(nil), b.data...)
	if b.pending.Len() > 0 {
		out = append(out, b.pending.String())
	}
	return out
}
