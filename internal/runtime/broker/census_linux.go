package broker

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
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
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			data, err := io.ReadAll(io.LimitReader(f, 8193))
			closeErr := f.Close()
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if closeErr != nil {
				return nil, closeErr
			}
			if len(data) > 8192 {
				return nil, ErrCensus
			}
			p, err := parseLinuxStat(pid, string(data))
			if err != nil {
				return nil, err
			}
			result = append(result, p)
		}
		if readErr == io.EOF {
			return result, nil
		}
		if readErr != nil {
			return nil, readErr
		}
	}
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
