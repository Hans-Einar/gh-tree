package api_test

import (
	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
	"testing"
)

func TestReviewRemoteEndpointIdentity(t *testing.T) {
	ra := rvMust(d.NewRepositoryID(d.Remote, "host/owner/A"))
	rb := rvMust(d.NewRepositoryID(d.Remote, "host/owner/B"))
	src := rvSource("remote")
	page := rvMust(a.NewPageInfo(a.PageInfoData{Completeness: a.Complete, Source: src}))
	obs := rvMust(a.NewRemoteObservation(a.RemoteObservationData{ID: rvMust(a.NewObservationID("remote-A")), Repository: ra, Interval: rvObservation(rvWork(rvRepo("local"), "one")).Data().Interval, Version: src, Origin: rvMust(a.NewLiveRemoteObservation(a.LiveRemoteObservationData{})), Page: page}))
	baseB := rvMust(a.NewUnavailableEndpoint(a.UnavailableEndpointData{KnownBranch: a.Some(rvMust(d.NewBranchID(rb, d.RemoteHead, "main"))), Reason: a.EndpointInaccessible}))
	head := rvMust(a.NewUnavailableEndpoint(a.UnavailableEndpointData{Reason: a.EndpointDeleted}))
	t.Run("unavailable_base_retains_foreign_branch", func(t *testing.T) {
		_, e := a.NewPullRequestFact(a.PullRequestFactData{ID: rvMust(d.NewPRID(ra, 1)), URL: "https://host/owner/A/pull/1", State: a.PROpen, Base: baseB, Head: head, Observation: obs})
		rvReject(t, e)
	})
	locator := rvMust(a.NewRemoteRepositoryLocator(a.RemoteRepositoryLocatorData{Host: "example.test", Owner: "owner", Name: "repo"}))
	repo := rvMust(a.NewRemoteRepository(a.RemoteRepositoryData{ID: ra, Locator: locator, URL: "https://example.test/owner/repo", Capabilities: rvMust(a.NewRemoteCapabilities(a.RemoteCapabilitiesData{}))}))
	endpoint := rvMust(a.NewAvailableEndpoint(a.AvailableEndpointData{Repository: repo, Branch: rvMust(d.NewBranchID(ra, d.RemoteHead, "main")), Revision: rvRev(ra, "1")}))
	t.Run("positive_qualified_PR", func(t *testing.T) {
		_, e := a.NewPullRequestFact(a.PullRequestFactData{ID: rvMust(d.NewPRID(ra, 1)), URL: "https://example.test/owner/repo/pull/1", State: a.PROpen, Base: endpoint, Head: head, Observation: obs})
		if e != nil {
			t.Fatal(e)
		}
	})
	t.Run("PR_URL_does_not_match_known_repository_or_number", func(t *testing.T) {
		_, e := a.NewPullRequestFact(a.PullRequestFactData{ID: rvMust(d.NewPRID(ra, 1)), URL: "https://elsewhere.test/someone/other/pull/999", State: a.PROpen, Base: endpoint, Head: head, Observation: obs})
		rvReject(t, e)
	})
}
