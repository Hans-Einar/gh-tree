package api

import (
	"strconv"
	"strings"
)

func consistentPullRequestFact(d PullRequestFactData) error {
	if r, p := endpointRepository(d.Base); p && r != d.ID.Repository() {
		return invalid("PR base independent scope")
	}
	if b, p := d.Base.(AvailableEndpoint); p {
		if !remoteURL(d.URL, b.data.Repository.data.Locator, d.ID.Number()) {
			return invalid("qualified PR URL")
		}
		if h, p := d.Head.(AvailableEndpoint); p && h.data.Repository.data.Locator.data.Host != b.data.Repository.data.Locator.data.Host {
			return invalid("cross-host PR endpoints")
		}
	} else {
		parts := strings.Split(strings.TrimPrefix(d.URL, "https://"), "/")
		if !strings.HasPrefix(d.URL, "https://") || len(parts) != 5 || !safeLocator(d.URL) || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] != "pull" || parts[4] != strconv.FormatUint(d.ID.Number(), 10) {
			return invalid("PR URL identity")
		}
	}
	return nil
}
func consistentListBranchesResult(d ListBranchesResultData) error {
	if int(d.Page.data.Returned) != len(d.Branches) {
		return invalid("remote branch page count")
	}
	if obs, p := d.Observation.Value(); p {
		if obs.data.Version != d.Page.data.Source {
			return invalid("remote branch page version")
		}
		for _, b := range d.Branches {
			if b.data.Branch.Repository() != obs.data.Repository {
				return invalid("remote branch result repository")
			}
		}
	}
	return nil
}
func consistentListPullRequestsResult(d ListPullRequestsResultData) error {
	if int(d.Page.data.Returned) != len(d.PullRequests) {
		return invalid("PR page count")
	}
	if obs, p := d.Observation.Value(); p {
		if obs.data.Version != d.Page.data.Source {
			return invalid("PR page version")
		}
		for _, pr := range d.PullRequests {
			if pr.data.ID.Repository() != obs.data.Repository {
				return invalid("PR list result repository")
			}
		}
	}
	return nil
}
func consistentObservePullRequestResult(d ObservePullRequestResultData) error {
	if pr, p := d.PullRequest.Value(); p {
		if obs, p := d.Observation.Value(); p && obs.data.Repository != pr.data.ID.Repository() {
			return invalid("PR observation result repository")
		}
	}
	return nil
}
func consistentCreatedWithDrift(d CreatedWithDriftData) error {
	if d.Created.data.ID.Repository() != d.RequestedBase.data.Branch.Repository() {
		return invalid("created PR requested base repository")
	}
	return nil
}
func consistentCreationIndeterminate(d CreationIndeterminateData) error {
	if c, p := d.Candidate.Value(); p && c.data.ID.Repository() != d.RequestEvidence.data.RequestedBase.data.Branch.Repository() {
		return invalid("creation candidate base repository")
	}
	return nil
}
