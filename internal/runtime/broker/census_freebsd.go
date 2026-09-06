package broker

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type censusWriter struct {
	bytes.Buffer
	limit int
}

func (w *censusWriter) Write(data []byte) (int, error) {
	if len(data) > w.limit-w.Len() {
		return 0, ErrCensus
	}
	return w.Buffer.Write(data)
}

func census(ctx context.Context) ([]processFact, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	bounded, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	// This is the frozen, trusted base-system observer. Its direct child waiter
	// and copied pipe workers are joined by Cmd.Wait; it runs no project code.
	cmd := exec.CommandContext(bounded, "/bin/ps", "-ax", "-o", "pid=", "-o", "ppid=", "-o", "pgid=", "-o", "sid=", "-o", "stat=")
	cmd.Env = []string{"LC_ALL=C", "PATH=/bin:/usr/bin"}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = time.Second
	output := &censusWriter{limit: 4 << 20}
	cmd.Stdout = output
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	observer := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	rows, err := parseFreeBSDPS(output.String())
	if err != nil {
		return nil, err
	}
	result := rows[:0]
	for _, p := range rows {
		// The snapshot contains ps itself. Its exact owned wait above proves it
		// has exited; it cannot become an unexpected live cleanup candidate.
		if p.pid == observer {
			if p.parent != os.Getpid() {
				return nil, ErrCensus
			}
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func parseFreeBSDPS(text string) ([]processFact, error) {
	var result []processFact
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 5 || len(result) >= maxCensusRecords {
			return nil, ErrCensus
		}
		pid, e1 := strconv.Atoi(fields[0])
		parent, e2 := strconv.Atoi(fields[1])
		group, e3 := strconv.Atoi(fields[2])
		sid, e4 := strconv.Atoi(fields[3])
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || pid < 0 || parent < 0 || group < 0 || sid < 0 || len(fields[4]) == 0 {
			return nil, ErrCensus
		}
		state := fields[4][0]
		if !strings.ContainsRune("DILRSTWZ", rune(state)) {
			return nil, ErrCensus
		}
		result = append(result, processFact{pid: pid, parent: parent, group: group, session: sid, live: state != 'Z', stopped: state == 'T'})
	}
	if len(result) == 0 {
		return nil, ErrCensus
	}
	return result, nil
}
