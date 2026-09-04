package graph

import "testing"

func TestRowsTracksBranchesAndMerges(t *testing.T) {
	commits := []Commit{
		{SHA:"m",Parents:[]string{"a","b"}},
		{SHA:"a",Parents:[]string{"r"}},
		{SHA:"b",Parents:[]string{"r"}},
		{SHA:"r"},
	}
	rows:=Rows(commits)
	if len(rows)!=4{t.Fatalf("rows=%#v",rows)}
	if rows[0].Prefix!="*"{t.Fatalf("merge prefix=%q",rows[0].Prefix)}
	if rows[1].Prefix!="* │"{t.Fatalf("first parent prefix=%q",rows[1].Prefix)}
	if rows[2].Prefix!="│ *"{t.Fatalf("second parent prefix=%q",rows[2].Prefix)}
	if rows[3].Prefix!="*"{t.Fatalf("root prefix=%q",rows[3].Prefix)}
}
