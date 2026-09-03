package github

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct { output []byte; err error; name string; args []string }
func (f *fakeRunner) Run(_ context.Context,_ string,name string,args ...string)([]byte,error){f.name=name;f.args=append([]string(nil),args...);return f.output,f.err}

func TestParsePullRequests(t *testing.T){t.Parallel();data:=[]byte(`[{"number":60,"title":"UIBox clean-cut rewrite","state":"OPEN","isDraft":true,"headRefName":"steering/Concept1/ui-box","baseRefName":"main","headRefOid":"3c83ea2d4e3ba071b5a6648129c6ebb136db8912","updatedAt":"2026-09-02T21:00:00Z"}]`);prs,err:=parsePullRequests(data);if err!=nil{t.Fatal(err)};if len(prs)!=1||prs[0].Number!=60||!prs[0].IsDraft||prs[0].UpdatedAt.IsZero(){t.Fatalf("parsed PRs = %#v",prs)}}
func TestParsePullRequestsRejectsMissingDeploymentIdentity(t *testing.T){t.Parallel();_,err:=parsePullRequests([]byte(`[{"number":1,"headRefName":"feature/x"}]`));if err==nil{t.Fatal("expected missing SHA to fail")}}
func TestListBranchesUsesBoundedAPIRequest(t *testing.T){t.Parallel();runner:=&fakeRunner{output:[]byte(`[{"name":"main","commit":{"sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}]`)};client:=NewClient(runner,"/repo");branches,err:=client.ListBranches(context.Background(),"Hans-Einar/gh-tree");if err!=nil{t.Fatal(err)};if !reflect.DeepEqual(branches,[]Branch{{Name:"main",SHA:strings.Repeat("a",40)}}){t.Fatalf("branches = %#v",branches)};joined:=strings.Join(runner.args," ");if runner.name!="gh"||!strings.Contains(joined,"branches?per_page=100")||strings.Contains(joined,"--paginate"){t.Fatalf("unexpected gh invocation: %s %s",runner.name,joined)}}
func TestResolveRepositoryUsesGhAuthenticationContext(t *testing.T){t.Parallel();runner:=&fakeRunner{output:[]byte("Hans-Einar/gh-tree\n")};client:=NewClient(runner,"/repo");repo,err:=client.ResolveRepository(context.Background(),"");if err!=nil{t.Fatal(err)};if repo!="Hans-Einar/gh-tree"{t.Fatalf("repo = %q",repo)};if got:=fmt.Sprintf("%s %s",runner.name,strings.Join(runner.args," "));!strings.Contains(got,"gh repo view"){t.Fatalf("unexpected invocation: %s",got)}}
func TestCreatePullRequestUsesArgumentSafeGhInvocation(t *testing.T){t.Parallel();runner:=&fakeRunner{output:[]byte("https://github.com/Hans-Einar/gh-tree/pull/9\n")};client:=NewClient(runner,"/repo");url,err:=client.CreatePullRequest(context.Background(),"Hans-Einar/gh-tree","feature/name-with-$-chars","main","Title with spaces","Body; not a shell command",true);if err!=nil{t.Fatal(err)};if !strings.Contains(url,"/pull/9"){t.Fatalf("url=%q",url)};want:=[]string{"pr","create","--repo","Hans-Einar/gh-tree","--head","feature/name-with-$-chars","--base","main","--title","Title with spaces","--body","Body; not a shell command","--draft"};if runner.name!="gh"||!reflect.DeepEqual(runner.args,want){t.Fatalf("invocation=%s %#v",runner.name,runner.args)}}
