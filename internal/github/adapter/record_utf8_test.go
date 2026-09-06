package adapter

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestMalformedBranchUTF8RetainsValidSibling(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "base", "project")
	good := fmt.Sprintf(`{"name":"日本語/topic","commit":{"sha":%q}}`, oidB)
	bad := `{"name":"bad` + string([]byte{0xff}) + `","commit":{"sha":"` + oidA + `"}}`
	a.run = func(context.Context, command) wireResult { return wire(200, "", "["+bad+","+good+"]") }
	r, err := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 100, initial()))
	if err != nil || !r.Valid() {
		t.Fatalf("independent branch unavailable: %v", err)
	}
	d := r.Data()
	if len(d.Branches) != 1 || d.Branches[0].Data().Branch.Name() != "日本語/topic" || d.Branches[0].Data().Tip.OID().String() != oidB {
		t.Fatal("valid branch changed or malformed branch retained")
	}
	assertMalformedRecordPage(t, d.Page, d.Diagnostics, "invalid-branch-record")
}

func TestMalformedPullRequestUTF8RetainsValidSibling(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "base", "project")
	good := prJSON("base", "fork", oidB)
	bad := strings.Replace(good, `"title":"literal title"`, `"title":"bad`+string([]byte{0xff})+`"`, 1)
	a.run = func(context.Context, command) wireResult { return wire(200, "", "["+bad+","+good+"]") }
	request := must(api.NewListPullRequestsRequest(api.ListPullRequestsRequestData{
		Repository: repo.Data().ID,
		Filter:     must(api.NewPullRequestFilter(api.PullRequestFilterData{State: api.FilterAll})),
		Page:       must(api.NewPageRequest(api.PageRequestData{Limit: 100, Continuation: initial()})),
	}))
	r, err := a.ListPullRequests(context.Background(), request)
	if err != nil || !r.Valid() {
		t.Fatalf("independent PR unavailable: %v", err)
	}
	d := r.Data()
	if len(d.PullRequests) != 1 {
		t.Fatalf("want only valid sibling, got %d", len(d.PullRequests))
	}
	pr := d.PullRequests[0].Data()
	if pr.ID.Repository() != repo.Data().ID || pr.ID.Number() != 7 || pr.Title != "literal title" || pr.URL != "https://github.com/base/project/pull/7" {
		t.Fatal("valid PR changed or malformed title replacement accepted")
	}
	head := pr.Head.(api.AvailableEndpoint).Data()
	if head.Revision.OID().String() != oidB || head.Repository.Data().ID.Token() != "github.com/fork/project" {
		t.Fatal("valid PR endpoint changed")
	}
	assertMalformedRecordPage(t, d.Page, d.Diagnostics, "invalid-pull-request-record")
}

func assertMalformedRecordPage(t *testing.T, page api.PageInfo, diagnostics []api.Diagnostic, reason string) {
	t.Helper()
	if page.Data().Returned != 1 || page.Data().Completeness != api.Unknown {
		t.Fatal("malformed record claimed complete or lost valid count", page.Data())
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Data().Reason == reason {
			return
		}
	}
	t.Fatalf("missing malformed-record diagnostic %s", reason)
}

func TestListFramingPreservesRawBytesAndRejectsBrokenEnvelope(t *testing.T) {
	bad := []byte(`{"name":"bad` + string([]byte{0xff}) + `"}`)
	records, err := decodeList(append(append([]byte{'['}, bad...), ']'))
	if err != nil || len(records) != 1 || !bytes.Equal(records[0], bad) {
		t.Fatal("framing rewrote invalid scalar bytes", err)
	}
	var decoded branchDTO
	if err := strictJSON(records[0], &decoded); err == nil {
		t.Fatal("invalid UTF-8 reached scalar decoding")
	}
	for _, raw := range []string{`[{"name":"good"}`, `[{"name" "good"}]`, `[{}] []`, `{"name":"good"}`, `null`, "[{} ," + string([]byte{0xff}) + "]"} {
		if records, err := decodeList([]byte(raw)); err == nil || records != nil {
			t.Fatalf("broken envelope retained records: %q", raw)
		}
	}
	if records, err := decodeList([]byte(`[]`)); err != nil || records == nil || len(records) != 0 {
		t.Fatal("valid empty array refused", err)
	}
}
