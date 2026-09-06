package api

func sameFacetFact(a, b FacetEffect) bool {
	x, y := a.data, b.data
	if x.Facet != y.Facet || x.State != y.State || x.PostObservation != y.PostObservation || len(x.RecoveryIDs) != len(y.RecoveryIDs) {
		return false
	}
	for _, id := range x.RecoveryIDs {
		found := false
		for _, other := range y.RecoveryIDs {
			found = found || id == other
		}
		if !found {
			return false
		}
	}
	return true
}
func containsFacetFact(report EffectReport, fact FacetEffect) bool {
	for _, v := range report.data.Facets {
		if sameFacetFact(v, fact) {
			return true
		}
	}
	return false
}

// Aggregate reports are unions of complete facts, not an ordinal state fold or
// a map keyed only by facet. Native step/subject provenance remains in the typed
// child results and its observation/recovery IDs. Identical facts may coalesce.
func MergeEffectReports(reports ...EffectReport) (EffectReport, error) {
	var facts []FacetEffect
	for _, r := range reports {
		if !r.Valid() {
			return EffectReport{}, invalid("effect report")
		}
		for _, f := range r.data.Facets {
			found := false
			for _, existing := range facts {
				found = found || sameFacetFact(existing, f)
			}
			if !found {
				facts = append(facts, f)
			}
		}
	}
	return NewEffectReport(EffectReportData{Facets: facts})
}
func (e evidenceSet) requireEffects(aggregate EffectReport) error {
	for _, f := range e.effects {
		if !containsFacetFact(aggregate, f) {
			return invalid("aggregate erased or replaced child effect fact")
		}
	}
	return nil
}

// An omitted facet makes no assertion. Explicit reports cannot characterize a
// known creation/establishment solely as never started or unchanged. Mixed,
// partial and unknown stage facts are retained without an ordinal merge.
func knownChangedFacet(report EffectReport, facet EffectFacet, narrower Optional[EffectReport]) error {
	reported := false
	onlyUnchanged := true
	for _, f := range report.data.Facets {
		if f.data.Facet != facet {
			continue
		}
		if child, p := narrower.Value(); p && containsFacetFact(child, f) {
			continue
		}
		reported = true
		if f.data.State != NotStarted && f.data.State != VerifiedNoTargetChange {
			onlyUnchanged = false
		}
	}
	if reported && onlyUnchanged {
		return invalid("known outcome reported solely unstarted/unchanged")
	}
	return nil
}
