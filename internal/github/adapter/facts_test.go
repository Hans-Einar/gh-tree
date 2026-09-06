package adapter

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

const oidA = "1111111111111111111111111111111111111111"
const oidB = "2222222222222222222222222222222222222222"

func testAdapter(t *testing.T) *Adapter {
	t.Helper()
	a, e := New(Config{Host: "github.com"})
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func repositoryJSON(owner, name string) string {
	return fmt.Sprintf(`{"id":123,"name":%q,"full_name":%q,"owner":{"login":%q},"html_url":%q,"default_branch":"main"}`, name, owner+"/"+name, owner, "https://github.com/"+owner+"/"+name)
}
func wire(status int, headers, body string) wireResult {
	return wireResult{stdout: []byte(fmt.Sprintf("HTTP/2.0 %d Status\r\n%s\r\n%s", status, headers, body)), transport: must(api.NewCommandTransportOutcome(api.CommandTransportOutcomeData{Started: true, RootReaped: true, CleanupKnown: true, ExitCode: api.Some(0)})), started: time.Now().UTC(), finished: time.Now().UTC()}
}
func resolveFixture(t *testing.T, a *Adapter, owner, name string) api.RemoteRepository {
	t.Helper()
	a.run = func(_ context.Context, c command) wireResult { return wire(200, "", repositoryJSON(owner, name)) }
	l := must(ParseLocator("github.com", owner+"/"+name))
	r, e := a.ResolveRepository(context.Background(), must(api.NewResolveRepositoryRequest(api.ResolveRepositoryRequestData{Locator: l})))
	if e != nil {
		t.Fatal(e)
	}
	repo, ok := r.Data().Repository.Value()
	if !ok {
		t.Fatal("missing repository")
	}
	return repo
}
func branchRequest(repo domain.RepositoryID, limit uint32, continuation api.PageContinuation) api.ListBranchesRequest {
	return must(api.NewListBranchesRequest(api.ListBranchesRequestData{Repository: repo, Filter: must(api.NewAllRemoteBranches(api.AllRemoteBranchesData{})), Page: must(api.NewPageRequest(api.PageRequestData{Limit: limit, Continuation: continuation}))}))
}
func initial() api.PageContinuation { return must(api.NewInitialPage(api.InitialPageData{})) }
func TestRepositoryScopeAndStrictParsing(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "Owner", "Project")
	if repo.Data().ID.Token() != "github.com/owner/project" {
		t.Fatal(repo.Data())
	}
	if _, e := ParseLocator("github.com", "https://user:secret@github.com/owner/project"); e == nil {
		t.Fatal("credentials accepted")
	}
	if _, e := ParseLocator("github.com", "owner/project.git.git"); e == nil {
		t.Fatal("multiple suffix stripping")
	}
	a.run = func(context.Context, command) wireResult {
		t.Fatal("foreign scope invoked transport")
		return wireResult{}
	}
	unknown := must(domain.NewRepositoryID(domain.Remote, "github.com/foreign/project"))
	r, e := a.ListBranches(context.Background(), branchRequest(unknown, 100, initial()))
	if e == nil || !r.Valid() || r.Data().Transport.Data().Started {
		t.Fatal(r, e)
	}
	for _, raw := range []string{`{"name":"a","name":"b"}`, `{"name":"\ud800"}`, `{"name":"a"} {}`, `{"name":true}`} {
		var d branchDTO
		if strictJSON([]byte(raw), &d) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
func TestBranchPagingMalformedConflictAndCap(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "owner", "project")
	calls := 0
	a.run = func(_ context.Context, c command) wireResult {
		calls++
		for _, arg := range c.args {
			if arg == "--paginate" {
				t.Fatal("unbounded traversal")
			}
		}
		headers := fmt.Sprintf("Link: <https://api.github.com/repos/owner/project/branches?page=%d&per_page=2>; rel=\"next\"\r\n", calls+1)
		body := fmt.Sprintf(`[{"name":"same","commit":{"sha":%q}},{"name":"invalid","commit":{"sha":"short"}}]`, oidA)
		if calls > 1 {
			body = fmt.Sprintf(`[{"name":"same","commit":{"sha":%q}},{"name":"other-%d","commit":{"sha":%q}}]`, oidB, calls, oidB)
		}
		return wire(200, headers, body)
	}
	cont := initial()
	var first api.RemoteBranchFact
	for n := 1; n <= 10; n++ {
		r, e := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 2, cont))
		if e != nil || !r.Valid() {
			t.Fatal(r, e)
		}
		d := r.Data()
		if n == 1 {
			if len(d.Branches) != 1 || d.Page.Data().Completeness != api.Unknown {
				t.Fatal(d)
			}
			first = d.Branches[0]
		}
		if n == 2 && d.Page.Data().Completeness != api.Unknown {
			t.Fatal("conflict claimed complete")
		}
		next, ok := d.Page.Data().Next.Value()
		if n < 10 {
			if !ok {
				t.Fatal("no continuation", n)
			}
			cont = next
		} else if ok {
			t.Fatal("cap continuation")
		}
	}
	if first.Data().Tip.OID().String() != oidA {
		t.Fatal("prior observation mutated")
	}
	if calls != 10 {
		t.Fatal(calls)
	}
	r, e := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 1, cont))
	if e == nil || r.Data().Transport.Data().Started {
		t.Fatal("cursor query reuse admitted")
	}
}
func TestBranchMalformedRecordRetainsIndependentFact(t *testing.T) {
	a := testAdapter(t)
	repo := resolveFixture(t, a, "owner", "project")
	a.run = func(context.Context, command) wireResult {
		return wire(200, "", fmt.Sprintf(`[{"name":"bad","name":"other","commit":{"sha":%q}},{"name":"good","commit":{"sha":%q}}]`, oidA, oidB))
	}
	r, e := a.ListBranches(context.Background(), branchRequest(repo.Data().ID, 100, initial()))
	if e != nil || len(r.Data().Branches) != 1 || r.Data().Page.Data().Completeness != api.Unknown {
		t.Fatal(r, e)
	}
}
func TestHTTPDiagnosticMetadataAndPrivate404(t *testing.T) {
	a := testAdapter(t)
	a.run = func(context.Context, command) wireResult {
		return wire(404, "X-Github-Request-Id: request-1\r\nX-Ratelimit-Remaining: 0\r\nRetry-After: 3\r\n", `{"message":"localized"}`)
	}
	r := a.request(context.Background(), "repos/a/b", "GET", nil)
	if r.err == nil || r.diagnostics[0].Data().Code != api.Unavailable {
		t.Fatal(r)
	}
	detail, _ := r.diagnostics[0].Data().Detail.Value()
	v := detail.(api.RemoteDiagnosticDetail).Data()
	if n, _ := v.HTTPStatus.Value(); n != 404 {
		t.Fatal(v)
	}
	if n, _ := v.RetryAfterSeconds.Value(); n != 3 {
		t.Fatal(v)
	}
}

func TestNativeCommandHelper(t *testing.T) {
	if os.Getenv("GH_TREE_ADAPTER_HELPER") == "" {
		return
	}
	mode := os.Getenv("GH_TREE_ADAPTER_HELPER")
	if os.Args[len(os.Args)-1] == "owned-child" {
		mode = "leased-child"
	}
	switch mode {
	case "streams":
		fmt.Print("[]")
		fmt.Fprint(os.Stderr, "private warning")
	case "overflow":
		fmt.Print(strings.Repeat("x", 8192))
	case "wait":
		time.Sleep(time.Minute)
	case "leased-child":
		// Unix deliberately promises only root/pipe ownership. Bound this owned
		// adversarial fixture's lifetime without granting product PGID authority.
		time.Sleep(1200 * time.Millisecond)
	case "input":
		_, _ = io.Copy(os.Stdout, os.Stdin)
	case "descendant":
		child := exec.Command(os.Args[0], "-test.run=^TestNativeCommandHelper$", "--", "owned-child")
		child.Env = os.Environ()
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		hideFixture(child)
		if e := child.Start(); e != nil {
			os.Exit(2)
		}
		fmt.Print(child.Process.Pid)
	}
	os.Exit(0)
}

func TestNativeDescendantPipeJoinAndRepeatedResources(t *testing.T) {
	before := runtime.NumGoroutine()
	a := nativeAdapter(t, "descendant")
	started := time.Now()
	r := a.runCommand(context.Background(), command{args: []string{"-test.run=^TestNativeCommandHelper$"}})
	if !r.transport.Data().RootReaped || time.Since(started) > 3*time.Second || r.err == nil {
		t.Fatalf("root/drain facts %+v %v", r.transport.Data(), r.err)
	}
	pid, e := strconv.Atoi(string(r.stdout))
	if e != nil {
		t.Fatalf("child PID evidence %q", r.stdout)
	}
	verifyFixtureChildExited(t, pid)
	if runtime.GOOS == "windows" && !r.transport.Data().CleanupKnown {
		t.Fatal("Job did not prove active0", r.transport.Data())
	}
	if runtime.GOOS != "windows" && r.transport.Data().CleanupKnown {
		t.Fatal("root-only profile erased known child residual")
	}
	for i := 0; i < 5; i++ {
		a = nativeAdapter(t, "streams")
		r = a.runCommand(context.Background(), command{args: []string{"-test.run=^TestNativeCommandHelper$"}})
		if r.err != nil || !r.transport.Data().CleanupKnown {
			t.Fatal(r.err)
		}
	}
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("unjoined goroutines %d -> %d", before, after)
	}
}
func nativeAdapter(t *testing.T, mode string) *Adapter {
	t.Helper()
	env := append(os.Environ(), "GH_TREE_ADAPTER_HELPER="+mode)
	a, e := New(Config{Host: "github.com", Executable: os.Args[0], Environment: env, ReadTimeout: 2 * time.Second, DrainTimeout: 500 * time.Millisecond, StdoutLimit: 1024})
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func TestNativeCommandStreamsLimitsAndCancel(t *testing.T) {
	args := []string{"-test.run=^TestNativeCommandHelper$"}
	literal := []byte("日本語 $(literal) & `backticks`\n@- --flag")
	echo := nativeAdapter(t, "input").runCommand(context.Background(), command{args: args, input: literal, mutation: true})
	if echo.err != nil || string(echo.stdout) != string(literal) {
		t.Fatal("literal input altered", echo.err)
	}
	a := nativeAdapter(t, "streams")
	r := a.runCommand(context.Background(), command{args: args})
	if r.err != nil || string(r.stdout) != "[]" || string(r.stderr) != "private warning" || !r.transport.Data().RootReaped || !r.transport.Data().CleanupKnown {
		t.Fatalf("%+v %s %s", r, r.stdout, r.stderr)
	}
	a = nativeAdapter(t, "overflow")
	r = a.runCommand(context.Background(), command{args: args})
	if r.err == nil || len(r.stdout) != 1024 || !r.transport.Data().StdoutTruncated {
		t.Fatal(r)
	}
	a = nativeAdapter(t, "wait")
	a.config.ReadTimeout = 100 * time.Millisecond
	start := time.Now()
	r = a.runCommand(context.Background(), command{args: args})
	if r.err == nil || !r.transport.Data().RootReaped || !r.transport.Data().CleanupKnown || time.Since(start) > 2*time.Second {
		t.Fatal(r)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	r = a.runCommand(canceled, command{args: args})
	if r.transport.Data().Started {
		t.Fatal("canceled start")
	}
}
