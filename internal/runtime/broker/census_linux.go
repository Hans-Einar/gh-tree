package broker

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func census(ctx context.Context) ([]processFact, error) {
	directory, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	var result []processFact
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entries, readErr := directory.ReadDir(256)
		for _, entry := range entries {
			pid, err := strconv.Atoi(entry.Name())
			if err != nil || pid <= 0 {
				continue
			}
			if len(result) >= maxCensusRecords {
				return nil, ErrCensus
			}
			f, err := os.Open("/proc/" + entry.Name() + "/stat")
			if linuxProcEntryGone(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			p, present, err := readLinuxStat(pid, f)
			if err != nil {
				return nil, err
			}
			if present {
				result = append(result, p)
			}
		}
		if readErr == io.EOF {
			return result, nil
		}
		if readErr != nil {
			return nil, readErr
		}
	}
}

// proc_pid_permission can return ESRCH after pathname lookup if the task
// disappeared. Accept only the native errno (optionally wrapped by os.Open),
// never a combined error that may also contain an unrelated acquisition failure.
func linuxProcEntryGone(err error) bool {
	if pathErr, ok := err.(*os.PathError); ok {
		err = pathErr.Err
	}
	return err == syscall.ENOENT || err == syscall.ESRCH
}

// readLinuxStat owns one opened proc stat file, including its close result.
func readLinuxStat(pid int, f io.ReadCloser) (processFact, bool, error) {
	data, readErr := io.ReadAll(io.LimitReader(f, 8193))
	closeErr := f.Close()
	if closeErr != nil {
		return processFact{}, false, errors.Join(readErr, closeErr)
	}
	// An opened proc stat inode does not keep its task alive. Linux's
	// proc_single_show returns ESRCH if that task exited after the open.
	// This record disappeared; other failures still invalidate the census.
	if os.IsNotExist(readErr) || errors.Is(readErr, syscall.ESRCH) {
		return processFact{}, false, nil
	}
	if readErr != nil {
		return processFact{}, false, readErr
	}
	if len(data) > 8192 {
		return processFact{}, false, ErrCensus
	}
	p, err := parseLinuxStat(pid, string(data))
	return p, err == nil, err
}

func parseLinuxStat(want int, line string) (processFact, error) {
	left := strings.IndexByte(line, '(')
	right := strings.LastIndexByte(line, ')')
	if left < 1 || right <= left {
		return processFact{}, ErrCensus
	}
	pid, err := strconv.Atoi(strings.TrimSpace(line[:left]))
	if err != nil || pid != want {
		return processFact{}, ErrCensus
	}
	fields := strings.Fields(line[right+1:])
	if len(fields) < 20 || len(fields[0]) != 1 {
		return processFact{}, ErrCensus
	}
	parent, e1 := strconv.Atoi(fields[1])
	group, e2 := strconv.Atoi(fields[2])
	sid, e3 := strconv.Atoi(fields[3])
	birth, e4 := strconv.ParseUint(fields[19], 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || parent < 0 || group < 0 || sid < 0 {
		return processFact{}, ErrCensus
	}
	state := fields[0][0]
	if !strings.ContainsRune("RSDZTXtKWPIN", rune(state)) {
		return processFact{}, ErrCensus
	}
	return processFact{pid: pid, parent: parent, group: group, session: sid, identity: strconv.FormatUint(birth, 10), live: state != 'Z' && state != 'X', stopped: state == 'T' || state == 't'}, nil
}
