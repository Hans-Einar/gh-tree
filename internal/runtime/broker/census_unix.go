//go:build linux || darwin || freebsd

package broker

import (
	"errors"
	"sort"
)

const maxCensusRecords = 65536

var ErrCensus = errors.New("Runtime session census is incomplete or unsupported")

type processFact struct {
	pid, parent, group, session int
	identity                    string
	live, stopped               bool
}

// sessionMembers rejects unexpected live members of the supervisor's reserved
// group. A numeric group is only an acquisition candidate, never a signal target.
func sessionMembers(all []processFact, supervisor int) ([]processFact, []int, error) {
	if supervisor <= 1 {
		return nil, nil, ErrCensus
	}
	var members []processFact
	groups := map[int]bool{}
	found := false
	for _, p := range all {
		if p.session != supervisor || !p.live {
			continue
		}
		if p.pid == supervisor {
			if p.group != supervisor {
				return nil, nil, ErrCensus
			}
			found = true
			continue
		}
		if p.group <= 0 || p.group == supervisor {
			return nil, nil, ErrCensus
		}
		members = append(members, p)
		groups[p.group] = true
	}
	if !found {
		return nil, nil, ErrCensus
	}
	result := make([]int, 0, len(groups))
	for group := range groups {
		result = append(result, group)
	}
	sort.Ints(result)
	return members, result, nil
}
