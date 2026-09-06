package api

// evidenceSet checks only referential integrity of returned semantic facts.
// Actual native observation/retention and causal attribution remain adapter work.
type evidenceSet struct {
	observations map[ObservationID]bool
	recovery     map[RecoveryID]RecoveryRecord
	storage      map[RecoveryID]StorageRecovery
	effects      []FacetEffect
	inconsistent bool
}

func newEvidenceSet() evidenceSet {
	return evidenceSet{observations: map[ObservationID]bool{}, recovery: map[RecoveryID]RecoveryRecord{}, storage: map[RecoveryID]StorageRecovery{}}
}
func (e *evidenceSet) collectGitObservation(v GitObservation)       { e.observations[v.data.ID] = true }
func (e *evidenceSet) collectRemoteObservation(v RemoteObservation) { e.observations[v.data.ID] = true }
func (e *evidenceSet) collectDiscoveryObservation(v DiscoveryObservation) {
	e.observations[v.data.ObservationID] = true
}
func (e *evidenceSet) collectFacetEffect(v FacetEffect) { e.effects = append(e.effects, v) }
func (e *evidenceSet) collectRecoveryRecord(v RecoveryRecord) {
	id := v.data.RecoveryID
	if old, p := e.recovery[id]; p && old.data != v.data {
		e.inconsistent = true
	}
	e.recovery[id] = v
}
func (e *evidenceSet) collectStorageRecovery(v StorageRecovery) {
	e.collectRecoveryRecord(v.data.Record)
	id := v.data.Record.data.RecoveryID
	if old, p := e.storage[id]; p && old.data != v.data {
		e.inconsistent = true
	}
	e.storage[id] = v
}
func (e evidenceSet) validate() error {
	if e.inconsistent {
		return invalid("inconsistent returned recovery identity")
	}
	for _, facet := range e.effects {
		if id, p := facet.data.PostObservation.Value(); p && !e.observations[id] {
			return invalid("dangling post observation")
		}
		for _, id := range facet.data.RecoveryIDs {
			if _, p := e.recovery[id]; !p {
				return invalid("dangling returned recovery")
			}
		}
	}
	return nil
}

func (e evidenceSet) recordUnion(records []RecoveryRecord) error {
	provided := map[RecoveryID]RecoveryRecord{}
	for _, r := range records {
		provided[r.data.RecoveryID] = r
	}
	for id, record := range e.recovery {
		r, p := provided[id]
		if !p || r.data != record.data {
			return invalid("incomplete recovery union")
		}
	}
	return nil
}
func (e evidenceSet) normalizedUnion(records []NormalizedRecovery) error {
	provided := map[RecoveryID]NormalizedRecovery{}
	for _, r := range records {
		provided[r.data.Record.data.RecoveryID] = r
	}
	for id, record := range e.recovery {
		r, p := provided[id]
		if !p || r.data.Record.data != record.data {
			return invalid("incomplete normalized recovery union")
		}
		if detail, ok := e.storage[id]; ok {
			stored, p := r.data.StorageDetail.Value()
			if !p || stored.data != detail.data {
				return invalid("lost storage recovery detail")
			}
		}
	}
	return nil
}
