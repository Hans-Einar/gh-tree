package ports_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	. "github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/application/ports"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func rsMust[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
func rsSource(s string) SourceVersion {
	return rsMust(NewSourceVersion("review", "scope", "lifetime", s))
}
func rsScope(name string, root byte) WorktreeScope {
	repo := rsMust(domain.NewRepositoryID(domain.LocalCommon, "review-repo"))
	w := rsMust(domain.NewWorktreeID(repo, name))
	dir := rsMust(NewDirectoryIdentity(DirectoryWindows, 1, [16]byte{root}, "stamp"))
	return rsMust(NewWorktreeScope(WorktreeScopeData{ID: w, RootLocator: "C:/review/" + name, RootIdentity: dir, Source: rsSource(name)}))
}
func rsEffects(f EffectFacet, s EffectState) EffectReport {
	return rsMust(NewEffectReport(EffectReportData{Facets: []FacetEffect{rsMust(NewFacetEffect(FacetEffectData{Facet: f, State: s}))}}))
}
func rsSnapshot(s WorktreeScope, id uint64, phase SessionPhase) SessionSnapshot {
	cwd := rsMust(NewCwdObservation(CwdObservationData{Worktree: s, ProjectIdentity: s.Data().RootIdentity, Source: rsSource("cwd")}))
	summary := rsMust(NewInvocationSummary(InvocationSummaryData{Label: "review", ExecutableDisplay: "literal.exe", Cwd: cwd, AcceptedLocator: s.Data().RootLocator, Terminal: Pipes, Geometry: rsMust(NewGeometry(GeometryData{Rows: 24, Columns: 80}))}))
	cleanup := CleanupPending
	if phase == Cleaned {
		cleanup = CleanupComplete
	}
	if phase == CleanupFailed {
		cleanup = CleanupFailedState
	}
	d := SessionSnapshotData{SessionID: rsMust(domain.NewSessionID(id)), WorktreeID: s.Data().ID, StartOperation: rsMust(NewOperationID(1)), Display: summary, Capabilities: rsMust(NewSessionCapabilities(SessionCapabilitiesData{Output: true, TreeStop: true, Restart: true})), Phase: phase, Cleanup: rsMust(NewSessionCleanup(SessionCleanupData{State: cleanup})), Sequence: rsMust(NewSessionSequence(1)), OutputRange: rsMust(NewOutputRange(OutputRangeData{}))}
	if phase == Running || phase == Cleaned {
		d.AcquiredCwd = Some(rsMust(NewAcquiredCwd(AcquiredCwdData{Observation: cwd, ActualLocator: Some(s.Data().RootLocator)})))
	}
	return rsMust(NewSessionSnapshot(d))
}
func rsDiagnostic() Diagnostic {
	return rsMust(NewDiagnostic(DiagnosticData{Code: IOFailure, Reason: "review-fault", Message: "Observed fault"}))
}
func rsRecovery(subjectScope, versionScope WorktreeScope, store string) StorageRecovery {
	subject := rsMust(NewRecoverySubject(RecoverySubjectData{Worktree: Some(subjectScope.Data().ID), Store: Some(store), Family: Some(RunConfig)}))
	version := rsMust(NewRunStorageVersion(versionScope, store, true, 12, [32]byte{1}))
	rv := rsMust(NewStorageRecoveryVersion(StorageRecoveryVersionData{Version: version}))
	record := rsMust(NewRecoveryRecord(RecoveryRecordData{RecoveryID: rsMust(NewRecoveryID("review-artifact")), Kind: RecoveryManifest, Layer: LayerPersistence, Subject: subject, Locator: "C:/review/manifest", Original: Some[RecoveryVersion](rv), NextAction: "Inspect retained evidence"}))
	return rsMust(NewStorageRecovery(StorageRecoveryData{Record: record, Family: RunConfig, Locator: "C:/review/manifest", Kind: Manifest, Identity: rsSource("artifact")}))
}

func TestRSPositiveControls(t *testing.T) {
	a := rsScope("a", 1)
	rejected := rsMust(NewSessionStartResult(SessionStartResultData{Effects: rsEffects(RuntimeResources, NotStarted)}))
	if rejected.Data().Session.Present() {
		t.Fatal("refused start acquired identity")
	}
	pending := rsSnapshot(a, 1, CleanupFailed)
	admitted := rsMust(NewSessionStartResult(SessionStartResultData{Session: Some(pending), Effects: rsEffects(RuntimeResources, EffectPartial), Diagnostics: []Diagnostic{rsDiagnostic()}}))
	if !admitted.Valid() || !admitted.Data().Session.Present() || admitted.Data().Established {
		t.Fatal("failed admitted start lost facts")
	}
	immediate := rsMust(NewSessionStartResult(SessionStartResultData{Session: Some(rsSnapshot(a, 2, Cleaned)), Established: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	if !immediate.Valid() {
		t.Fatal("immediately cleaned establishment rejected")
	}
	old := rsMust(NewSessionStopResult(SessionStopResultData{Session: rsSnapshot(a, 2, Cleaned), CleanupComplete: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	replacementData := rsSnapshot(a, 3, Running).Data()
	replacementData.RestartOf = Some(old.Data().Session.Data().SessionID)
	replacement := rsMust(NewSessionStartResult(SessionStartResultData{Session: Some(rsMust(NewSessionSnapshot(replacementData))), Established: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	if _, err := NewSessionRestartResult(SessionRestartResultData{Old: old, Replacement: Some(replacement)}); err != nil {
		t.Fatal(err)
	}
	noReplacement := rsMust(NewSessionRestartResult(SessionRestartResultData{Old: old, CancellationAsked: true}))
	if noReplacement.Data().Replacement.Present() {
		t.Fatal("cancellation fabricated replacement")
	}
	stop := rsMust(NewSessionStopResult(SessionStopResultData{Session: pending, Effects: rsEffects(RuntimeResources, EffectPartial)}))
	if _, err := NewRuntimeShutdownResult(RuntimeShutdownResultData{AdmissionClosed: true, Complete: true, Sessions: []SessionStopResult{stop}}); err == nil {
		t.Fatal("complete aggregate accepted pending cleanup")
	}
	if _, err := NewRuntimeEvent(RuntimeEventData{Sequence: rsMust(NewRuntimeEventSequence(0)), SessionSequence: pending.Data().Sequence, SessionID: pending.Data().SessionID, Kind: StateChanged, Snapshot: pending}); err == nil {
		t.Fatal("zero event admitted")
	}
	if _, err := NewRuntimeEvent(RuntimeEventData{Sequence: rsMust(NewRuntimeEventSequence(1)), SessionSequence: rsMust(NewSessionSequence(2)), SessionID: pending.Data().SessionID, Kind: StateChanged, Snapshot: pending}); err == nil {
		t.Fatal("foreign session sequence admitted")
	}
	if _, err := NewRuntimeEvent(RuntimeEventData{Sequence: rsMust(NewRuntimeEventSequence(1)), SessionSequence: pending.Data().Sequence, SessionID: rsMust(domain.NewSessionID(99)), Kind: StateChanged, Snapshot: pending}); err == nil {
		t.Fatal("foreign event identity admitted")
	}
}

func TestRSStartIdentityAndEstablishment(t *testing.T) {
	t.Run("absent-session-applied-resources", func(t *testing.T) {
		v, e := NewSessionStartResult(SessionStartResultData{Effects: rsEffects(RuntimeResources, AppliedVerified)})
		t.Logf("error=%v valid=%v sessionPresent=%v", e, v.Valid(), v.Data().Session.Present())
		if e == nil {
			t.Error("accepted applied Runtime resources without admitted session identity")
		}
	})
	t.Run("running-but-unestablished", func(t *testing.T) {
		v, e := NewSessionStartResult(SessionStartResultData{Session: Some(rsSnapshot(rsScope("a", 1), 1, Running)), Effects: rsEffects(RuntimeResources, AppliedVerified)})
		t.Logf("error=%v valid=%v established=%v", e, v.Valid(), v.Data().Established)
		if e == nil {
			t.Error("accepted Running session with Established false")
		}
	})
}

func TestRSRestartScopeBinding(t *testing.T) {
	a, b := rsScope("a", 1), rsScope("b", 2)
	old := rsMust(NewSessionStopResult(SessionStopResultData{Session: rsSnapshot(a, 1, Cleaned), CleanupComplete: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	d := rsSnapshot(b, 2, Running).Data()
	d.RestartOf = Some(old.Data().Session.Data().SessionID)
	start := rsMust(NewSessionStartResult(SessionStartResultData{Session: Some(rsMust(NewSessionSnapshot(d))), Established: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	v, e := NewSessionRestartResult(SessionRestartResultData{Old: old, Replacement: Some(start)})
	t.Logf("error=%v valid=%v old-worktree=%s replacement-worktree=%s", e, v.Valid(), a.Data().ID.AdministrativeKey(), b.Data().ID.AdministrativeKey())
	if e == nil {
		t.Error("restart changes original selected worktree and cwd")
	}
}

func TestRSStorageWorktreeRecoveryBinding(t *testing.T) {
	a, b := rsScope("a", 1), rsScope("b", 2)
	positive := rsRecovery(a, a, "store")
	data := positive.Data().Record.Data()
	foreign := rsMust(NewRunStorageVersion(b, "store", true, 12, [32]byte{1}))
	data.Original = Some[RecoveryVersion](rsMust(NewStorageRecoveryVersion(StorageRecoveryVersionData{Version: foreign})))
	// The shared record now rejects the conflicting subject before its wrapper.
	if _, e := NewRecoveryRecord(data); e == nil {
		t.Fatal("foreign document worktree")
	}
}

func TestRSStorageCommitVersionWorktreeBinding(t *testing.T) {
	a, b := rsScope("a", 1), rsScope("b", 2)
	proposed := rsMust(NewRunStorageVersion(a, "store", true, 12, [32]byte{1}))
	current := rsMust(NewRunStorageVersion(b, "store", true, 14, [32]byte{2}))
	v, e := NewStorageCommitResult(StorageCommitResultData{Outcome: Committed, PublicationKnown: true, Durability: SupportedCrashBarrierComplete, ProposedVersion: Some(proposed), CurrentVersion: Some(current), Effects: rsEffects(Storage, AppliedVerified)})
	t.Logf("error=%v valid=%v", e, v.Valid())
	if e == nil {
		t.Error("commit result accepts proposed/current versions from different worktrees")
	}
}

func TestRSStorageRecoveryLoadBinding(t *testing.T) {
	a, b := rsScope("a", 1), rsScope("b", 2)
	versionA := rsMust(NewRunStorageVersion(a, "store-a", true, 12, [32]byte{1}))
	recoveryB := rsRecovery(b, b, "store-b")
	if _, e := NewStorageLoadObservation(StorageLoadObservationData{State: Corrupt, Version: Some(versionA), Recovery: []StorageRecovery{recoveryB}}); e == nil {
		t.Fatal("conflicting observed version/recovery")
	}
	observation := rsMust(NewStorageLoadObservation(StorageLoadObservationData{State: Corrupt, Recovery: []StorageRecovery{recoveryB}}))
	if _, e := ports.NewLoadedRunConfig(a, observation, None[RunConfigDocument]()); e == nil {
		t.Fatal("load scope conflicts with recovery without version")
	}
}

func TestRSOutputAndCopyControls(t *testing.T) {
	id := rsMust(domain.NewSessionID(1))
	sequence := rsMust(NewSessionSequence(4))
	input := []byte{0, 27, 255, 10}
	chunk := rsMust(NewSessionOutputChunk(SessionOutputChunkData{Stream: Stdout, Offset: 10, Bytes: input, Sequence: sequence}))
	input[0] = 99
	copied := chunk.Data()
	copied.Bytes[1] = 99
	gap := rsMust(NewOutputGap(OutputGapData{From: 2, To: 10}))
	result := rsMust(NewSessionOutputResult(SessionOutputResultData{SessionID: id, Sequence: sequence, Chunks: []SessionOutputChunk{chunk}, RetainedStart: 10, End: 14, NextOffset: 14, Gap: Some(gap), Truncated: true}))
	got := result.Data()
	got.Chunks[0] = SessionOutputChunk{}
	if !bytes.Equal(result.Clone().Data().Chunks[0].Data().Bytes, []byte{0, 27, 255, 10}) {
		t.Fatal("output nested copy alias")
	}
	bad := result.Data()
	bad.NextOffset = 13
	if _, err := NewSessionOutputResult(bad); err == nil {
		t.Fatal("invalid next offset admitted")
	}
	bad = result.Data()
	bad.Gap = Some(rsMust(NewOutputGap(OutputGapData{From: 2, To: 9})))
	if _, err := NewSessionOutputResult(bad); err == nil {
		t.Fatal("wrong gap end admitted")
	}
	if _, err := NewSessionWriteRequest(SessionWriteRequestData{SessionID: id, Bytes: make([]byte, 65537)}); err == nil {
		t.Fatal("oversize input admitted")
	}
	if _, err := NewSessionWriteResult(SessionWriteResultData{SessionID: id, Sequence: sequence, AcceptedBytes: 65537}); err == nil {
		t.Fatal("oversize accepted count admitted")
	}
}

func TestRSStorageJSONAndCopyControls(t *testing.T) {
	for _, raw := range []string{`{"a":1,"\u0061":2}`, `{"nested":{"x":1,"x":2}}`, `{"𝄞":1,"\uD834\uDD1E":2}`, `"\uD800"`, string([]byte{34, 255, 34}), strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)} {
		if _, e := NewOpaqueJSON([]byte(raw)); e == nil {
			t.Fatalf("malformed opaque admitted: %q", raw)
		}
	}
	raw := []byte(`{"z":[1,null," "]}`)
	opaque := rsMust(NewOpaqueJSON(raw))
	raw[2] = 'X'
	output := opaque.Bytes()
	output[2] = 'X'
	member := rsMust(NewJSONMember("future", opaque))
	members := rsMust(NewJSONMembers([]JSONMember{member}))
	entries := members.Entries()
	entries[0] = JSONMember{}
	targets := []string{" a ", "b"}
	definition := rsMust(NewSavedLaunchDefinition(SavedLaunchDefinitionData{Provider: "future-provider", Targets: PresentField(targets), UnknownMembers: members}))
	targets[0] = "changed"
	entry := rsMust(NewSavedLaunchEntry(SavedLaunchEntryData{Alias: " Exact ", Definition: definition}))
	list := []SavedLaunchEntry{entry}
	doc := rsMust(NewRunConfigDocument(RunConfigDocumentData{SchemaVersion: 1, Launch: PresentField(list), UnknownMembers: rsMust(NewJSONMembers(nil))}))
	list[0] = SavedLaunchEntry{}
	got, _ := doc.Data().Launch.Value()
	nested, _ := got[0].Data().Definition.Data().Targets.Value()
	nested[0] = "changed"
	again, _ := doc.Clone().Data().Launch.Value()
	nested, _ = again[0].Data().Definition.Data().Targets.Value()
	if nested[0] != " a " || string(again[0].Data().Definition.Data().UnknownMembers.Entries()[0].Value().Bytes()) != `{"z":[1,null," "]}` {
		t.Fatal("stored nested copy/preservation lost")
	}
	collision := rsMust(NewJSONMembers([]JSONMember{rsMust(NewJSONMember("launch", opaque))}))
	if _, e := NewRunConfigDocument(RunConfigDocumentData{SchemaVersion: 1, UnknownMembers: collision}); e == nil {
		t.Fatal("known name collision admitted")
	}
	if _, e := NewRunConfigDocument(RunConfigDocumentData{SchemaVersion: 1, Default: NullField[string](), UnknownMembers: members}); e == nil {
		t.Fatal("known null admitted")
	}
	if _, e := NewUserConfigDocument(UserConfigDocumentData{SchemaVersion: 1, StripPrefixes: NullField[[]string](), UnknownMembers: members}); e != nil {
		t.Fatal("allowed null prefix refused", e)
	}
	if _, e := NewRunConfigDocument(RunConfigDocumentData{SchemaVersion: 1, Launch: PresentField([]SavedLaunchEntry{entry, entry}), UnknownMembers: members}); e == nil {
		t.Fatal("duplicate alias admitted")
	}
}

func TestRSStorageErrorAndNormalizationControls(t *testing.T) {
	a := rsScope("a", 1)
	recovery := rsRecovery(a, a, "store")
	proposed := rsMust(NewRunStorageVersion(a, "store", true, 13, [32]byte{2}))
	current := rsMust(NewRunStorageVersion(a, "store", true, 14, [32]byte{3}))
	effect := rsMust(NewFacetEffect(FacetEffectData{Facet: Storage, State: AppliedVerified, RecoveryIDs: []RecoveryID{recovery.Data().Record.Data().RecoveryID}}))
	effects := rsMust(NewEffectReport(EffectReportData{Facets: []FacetEffect{effect}}))
	result := rsMust(NewStorageCommitResult(StorageCommitResultData{Outcome: CommittedDurabilityUncertain, PublicationKnown: true, Durability: DurabilityUncertain, ProposedVersion: Some(proposed), CurrentVersion: Some(current), Effects: effects, Recovery: []StorageRecovery{recovery}, Diagnostics: []Diagnostic{rsDiagnostic()}}))
	returnWithError := func() (StorageCommitResult, error) { return result, errors.New("postcommit observation failure") }
	value, err := returnWithError()
	if err == nil || value.Data().Outcome != CommittedDurabilityUncertain || !value.Data().PublicationKnown {
		t.Fatal("known result plus error lost")
	}
	normalized, e := NormalizeRecovery([]RecoveryRecord{recovery.Data().Record}, []StorageRecovery{recovery, recovery})
	if e != nil || len(normalized) != 1 {
		t.Fatal("normalization", e)
	}
	detail, ok := normalized[0].Data().StorageDetail.Value()
	if !ok || detail.Data().Identity != recovery.Data().Identity || detail.Data().Family != RunConfig || detail.Data().Kind != Manifest {
		t.Fatal("normalization dropped family detail")
	}
	bad := recovery.Data()
	bad.Identity = rsSource("different-artifact")
	if _, e := NormalizeRecovery(nil, []StorageRecovery{recovery, rsMust(NewStorageRecovery(bad))}); e == nil {
		t.Fatal("conflicting duplicate recovery normalized")
	}
	document := rsMust(NewRunConfigDocument(RunConfigDocumentData{SchemaVersion: 1, UnknownMembers: rsMust(NewJSONMembers(nil))}))
	if _, e := ports.NewRunConfigCommit(rsScope("b", 2), proposed, document); e == nil {
		t.Fatal("foreign expected worktree admitted")
	}
	if _, e := ports.NewRunConfigCommit(a, proposed, document); e != nil {
		t.Fatal("valid expected worktree refused")
	}
}

func TestM259StorageRootAndUnavailableCurrent(t *testing.T) {
	original, changed := rsScope("a", 1), rsScope("a", 2)
	proposal := rsMust(NewRunStorageVersion(original, "store", true, 13, [32]byte{1}))
	current := rsMust(NewRunStorageVersion(changed, "store", true, 13, [32]byte{1}))
	d := StorageCommitResultData{Outcome: CommittedDurabilityUncertain, PublicationKnown: true, Durability: DurabilityUncertain, ProposedVersion: Some(proposal), Effects: rsEffects(Storage, AppliedVerified)}
	if _, e := NewStorageCommitResult(d); e != nil {
		t.Fatal("unavailable current loses known publication", e)
	}
	d.CurrentVersion = Some(current)
	if _, e := NewStorageCommitResult(d); e == nil {
		t.Fatal("same worktree changed root")
	}
	record := rsRecovery(original, original, "store").Data().Record.Data()
	record.Proposed = Some[RecoveryVersion](rsMust(NewStorageRecoveryVersion(StorageRecoveryVersionData{Version: current})))
	if _, e := NewRecoveryRecord(record); e == nil {
		t.Fatal("retained original/proposed roots conflict")
	}
}

func TestM259RestartSpecificationAndObservationControls(t *testing.T) {
	scope := rsScope("a", 1)
	old := rsMust(NewSessionStopResult(SessionStopResultData{Session: rsSnapshot(scope, 1, Cleaned), CleanupComplete: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	next := rsSnapshot(scope, 2, Running).Data()
	next.RestartOf = Some(old.Data().Session.Data().SessionID)
	display := next.Display.Data()
	display.Geometry = rsMust(NewGeometry(GeometryData{Rows: 30, Columns: 100}))
	cwd := display.Cwd.Data()
	cwd.Source = rsSource("fresh-observation")
	display.Cwd = rsMust(NewCwdObservation(cwd))
	next.Display = rsMust(NewInvocationSummary(display))
	replacement := rsMust(NewSessionStartResult(SessionStartResultData{Session: Some(rsMust(NewSessionSnapshot(next))), Established: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	if _, e := NewSessionRestartResult(SessionRestartResultData{Old: old, Replacement: Some(replacement)}); e != nil {
		t.Fatal("geometry/freshness change refused", e)
	}
	display.ArgumentDisplay = []string{"different-argument"}
	next.Display = rsMust(NewInvocationSummary(display))
	replacement = rsMust(NewSessionStartResult(SessionStartResultData{Session: Some(rsMust(NewSessionSnapshot(next))), Established: true, Effects: rsEffects(RuntimeResources, AppliedVerified)}))
	if _, e := NewSessionRestartResult(SessionRestartResultData{Old: old, Replacement: Some(replacement)}); e == nil {
		t.Fatal("restart changed published argv")
	}
}
