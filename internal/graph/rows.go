package graph

// Row is a display-oriented projection of one structured DAG commit. Prefix is
// derived from parent relationships, while all semantic refs remain on Commit.
type Row struct {
	Commit Commit
	Prefix string
	Lane   int
}

// Rows converts topologically ordered commits into stable lane prefixes without
// parsing `git log --graph` terminal artwork.
func Rows(commits []Commit) []Row {
	lanes := []string{}
	rows := make([]Row, 0, len(commits))
	for _, commit := range commits {
		lane := indexOf(lanes, commit.SHA)
		if lane < 0 {
			lane = len(lanes)
			lanes = append(lanes, commit.SHA)
		}
		prefix := ""
		for i := range lanes {
			if i > 0 { prefix += " " }
			if i == lane { prefix += "*" } else { prefix += "│" }
		}
		rows = append(rows, Row{Commit: commit, Prefix: prefix, Lane: lane})

		switch len(commit.Parents) {
		case 0:
			lanes = append(lanes[:lane], lanes[lane+1:]...)
		default:
			lanes[lane] = commit.Parents[0]
			if len(commit.Parents) > 1 {
				extra := append([]string(nil), commit.Parents[1:]...)
				tail := append([]string(nil), lanes[lane+1:]...)
				lanes = append(lanes[:lane+1], extra...)
				lanes = append(lanes, tail...)
			}
		}
		lanes = dedupe(lanes)
	}
	return rows
}

func indexOf(values []string, value string) int {
	for i, v := range values { if v == value { return i } }
	return -1
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	out := values[:0]
	for _, v := range values {
		if v == "" || seen[v] { continue }
		seen[v] = true
		out = append(out, v)
	}
	return out
}
