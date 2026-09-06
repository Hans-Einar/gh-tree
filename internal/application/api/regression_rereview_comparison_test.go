package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

func icorr() a.Correlation {
	return rvMust(a.NewCorrelation(a.CorrelationData{Query: a.Some(rvMust(a.NewQueryCorrelation(a.QueryCorrelationData{Slot: rvMust(a.NewQuerySlot("comparison")), Generation: rvMust(a.NewQueryGeneration(1))})))}))
}
func iterm(c a.Correlation, r a.Result) a.OperationTerminal {
	return rvMust(a.NewOperationTerminal(a.OperationTerminalData{OperationID: rvMust(a.NewOperationID(9)), Correlation: c, Disposition: a.Failed, Result: a.Some[a.Result](r), Effects: rvEffects()}))
}
func TestIndependentExactRequestResults(t *testing.T) {
	repo := rvRepo("comparison")
	one, two, three := rvRev(repo, "1"), rvRev(repo, "2"), rvRev(repo, "3")
	corr := icorr()
	sources := rvMust(a.NewProjectionSources(a.ProjectionSourcesData{}))
	first := rvMust(a.NewCommitPairComparison(a.CommitPairComparisonData{From: one, To: two}))
	different := rvMust(a.NewCommitPairComparison(a.CommitPairComparisonData{From: one, To: three}))
	limits := rvMust(a.NewPatchLimits(a.PatchLimitsData{MaxBytes: 32, MaxFiles: 1}))
	request := rvMust(a.NewQueryRequest(rvMust(a.NewDiffQuery(a.DiffQueryData{Comparison: first, Limits: limits})), corr))
	correct := rvMust(a.NewDiffResult(a.DiffResultData{Comparison: first, Sources: sources}))
	t.Run("positive_diff_comparison", func(t *testing.T) {
		if e := a.ValidateTerminalFor(request, iterm(corr, correct)); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("reject_changed_diff_endpoint", func(t *testing.T) {
		result := rvMust(a.NewDiffResult(a.DiffResultData{Comparison: different, Sources: sources}))
		if e := a.ValidateTerminalFor(request, iterm(corr, result)); e == nil {
			t.Fatal("different exact diff endpoint accepted")
		}
	})
	remote := rvMust(d.NewRepositoryID(d.Remote, "remote"))
	pr := rvMust(d.NewPRID(remote, 1))
	target := rvMust(d.NewPullRequestTarget(pr, rvRev(remote, "1")))
	baseA := rvMust(a.NewExactPRBase(a.ExactPRBaseData{Revision: rvRev(remote, "2")}))
	baseB := rvMust(a.NewExactPRBase(a.ExactPRBaseData{Revision: rvRev(remote, "3")}))
	preq := rvMust(a.NewQueryRequest(rvMust(a.NewPullRequestDiffQuery(a.PullRequestDiffQueryData{Target: target, Base: baseA, Local: repo, Limits: limits})), corr))
	t.Run("reject_changed_requested_PR_base", func(t *testing.T) {
		r := rvMust(a.NewPullRequestDiffResult(a.PullRequestDiffResultData{Target: target, RequestedBase: baseB, Sources: sources}))
		if e := a.ValidateTerminalFor(preq, iterm(corr, r)); e == nil {
			t.Fatal("different requested PR base accepted")
		}
	})
	page := rvMust(a.NewPageRequest(a.PageRequestData{Limit: 10, Continuation: rvMust(a.NewInitialPage(a.InitialPageData{}))}))
	filter := rvMust(a.NewAllGraph(a.AllGraphData{}))
	greq := rvMust(a.NewQueryRequest(rvMust(a.NewGraphQuery(a.GraphQueryData{Repository: repo, Roots: []d.Revision{one}, Filter: filter, Page: page})), corr))
	t.Run("reject_changed_graph_roots", func(t *testing.T) {
		r := rvMust(a.NewGraphResult(a.GraphResultData{Repository: repo, Roots: []d.Revision{three}, Sources: sources, Page: rvMust(a.NewPageInfo(a.PageInfoData{Completeness: a.Unknown, Source: rvSource("graph")}))}))
		if e := a.ValidateTerminalFor(greq, iterm(corr, r)); e == nil {
			t.Fatal("different exact graph roots accepted")
		}
	})
}
