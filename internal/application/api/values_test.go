package api

import (
	"errors"
	"github.com/Hans-Einar/gh-tree/internal/domain"
	"strings"
	"testing"
	"time"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func source(token string) SourceVersion {
	return must(NewSourceVersion("test", "repository", "observer", token))
}
func operation(n uint64) OperationID { return must(NewOperationID(n)) }
func emptyEffects() EffectReport     { return must(NewEffectReport(EffectReportData{})) }
func unknowns() JSONMembers          { return must(NewJSONMembers(nil)) }
func local(token string) domain.RepositoryID {
	return must(domain.NewRepositoryID(domain.LocalCommon, token))
}
func remoteRepo(token string) domain.RepositoryID {
	return must(domain.NewRepositoryID(domain.Remote, token))
}
func revision(repo domain.RepositoryID, digit string) domain.Revision {
	return must(domain.NewRevision(repo, must(domain.NewOID(strings.Repeat(digit, 40)))))
}
func worktree(repo domain.RepositoryID) domain.WorktreeID {
	return must(domain.NewWorktreeID(repo, "primary"))
}
func directory(platform DirectoryPlatform) DirectoryIdentity {
	return must(NewDirectoryIdentity(platform, 3, [16]byte{4}, "observed-stamp"))
}
func scope(w domain.WorktreeID) WorktreeScope {
	return must(NewWorktreeScope(WorktreeScopeData{ID: w, RootLocator: "C:/fixture", RootIdentity: directory(DirectoryWindows), Source: source("root")}))
}
func interval() ObservationInterval {
	at := time.Date(2026, 9, 6, 1, 0, 0, 0, time.UTC)
	return must(NewObservationInterval(ObservationIntervalData{StartedAt: at, FinishedAt: at.Add(time.Second)}))
}
func observation(repo domain.RepositoryID, w Optional[domain.WorktreeID]) GitObservation {
	return must(NewGitObservation(GitObservationData{ID: must(NewObservationID("git-observation")), Repository: repo, Worktree: w, Interval: interval(), Version: source("git"), Completeness: Complete}))
}
func expected(w domain.WorktreeID) GitExpectedState {
	h := must(domain.NewDetachedHead(revision(w.Repository(), "1")))
	return must(NewGitExpectedState(GitExpectedStateData{Repository: w.Repository(), Worktree: Some(w), Observation: source("status"), Head: Some(h), Index: Some(source("index")), WorktreeState: Some(source("bytes")), Configuration: source("config"), Inventory: source("inventory")}))
}
func diagnostic() Diagnostic {
	return must(NewDiagnostic(DiagnosticData{Code: IOFailure, Reason: "refresh-failed", Message: "Refresh failed after the known effect."}))
}
func storageVersion(f StorageFamily) StorageVersion {
	return must(NewStorageVersion(f, "bound-store", true, 12, [32]byte{1}))
}

func TestIdentityDomainsAndClosedRequests(t *testing.T) {
	if _, e := NewOperationID(0); e == nil {
		t.Fatal("zero operation")
	}
	if _, e := NewSourceVersion("git", "scope", "", "t"); e == nil {
		t.Fatal("empty issuer")
	}
	a := source("a")
	b := must(NewSourceVersion("remote", "repository", "observer", "a"))
	if a == b {
		t.Fatal("source namespaces collapsed")
	}
	q := must(NewQueryCorrelation(QueryCorrelationData{Slot: must(NewQuerySlot("console/output")), Generation: must(NewQueryGeneration(2))}))
	correlation := must(NewCorrelation(CorrelationData{Intent: must(NewIntentToken(3)), Query: Some(q)}))
	session := must(domain.NewSessionID(3))
	query := must(NewSessionOutputQuery(SessionOutputQueryData{SessionID: session, MaxBytes: 32}))
	if _, e := NewQueryRequest(query, correlation); e != nil {
		t.Fatal(e)
	}
	command := must(NewWriteInputCommand(WriteInputCommandData{SessionID: session, Bytes: []byte("input")}))
	if _, e := NewCommandRequest(command, correlation); e == nil {
		t.Fatal("mutation requested supersession")
	}
	mutationCorrelation := must(NewCorrelation(CorrelationData{}))
	if _, e := NewQueryRequest(query, mutationCorrelation); e == nil {
		t.Fatal("query without generation")
	}
	if _, e := NewCommandRequest(nil, mutationCorrelation); e == nil {
		t.Fatal("nil command")
	}
	var typedNil *WriteInputCommand
	if _, e := NewCommandRequest(typedNil, mutationCorrelation); e == nil {
		t.Fatal("typed nil command")
	}
	if _, e := NewCommandRequest(struct{ WriteInputCommand }{command}, mutationCorrelation); e == nil {
		t.Fatal("embedded foreign command")
	}
	input := []byte{1, 2, 3}
	cmd := must(NewWriteInputCommand(WriteInputCommandData{SessionID: session, Bytes: input}))
	request := must(NewCommandRequest(cmd, mutationCorrelation))
	input[0] = 9
	published, _ := request.Command()
	copyData := published.(WriteInputCommand).Data()
	copyData.Bytes[1] = 9
	if got := published.(WriteInputCommand).Data().Bytes; got[0] != 1 || got[1] != 2 {
		t.Fatal("request buffer alias", got)
	}
	if _, e := NewSessionWriteRequest(SessionWriteRequestData{SessionID: session, Bytes: make([]byte, 65537)}); e == nil {
		t.Fatal("input ceiling")
	}
	if _, e := NewGeometry(GeometryData{Rows: 0, Columns: 20}); e == nil {
		t.Fatal("zero geometry")
	}
	if _, e := NewGeometry(GeometryData{Rows: 32768, Columns: 20}); e == nil {
		t.Fatal("ABI overflow")
	}
}

func TestExactScopeAndReconciliationFacets(t *testing.T) {
	repo := local("clone-a")
	other := local("clone-b")
	w := worktree(repo)
	h := must(domain.NewDetachedHead(revision(other, "1")))
	if _, e := NewWorktreeFacts(WorktreeFactsData{ID: w, Scope: Some(scope(w)), Head: Some(h), Availability: must(NewAvailableWorktree(AvailableWorktreeData{})), Observation: observation(repo, Some(w))}); e == nil {
		t.Fatal("foreign head")
	}
	missing := must(NewWorktreeFacts(WorktreeFactsData{ID: w, Availability: must(NewMissingWorktree(MissingWorktreeData{})), Observation: observation(repo, Some(w))}))
	if missing.Data().Scope.Present() || missing.Data().Head.Present() {
		t.Fatal("missing fabricated facts")
	}
	remote := remoteRepo("fork")
	selected := revision(remote, "2")
	binding := must(NewRemoteBinding(RemoteBindingData{LocalRepository: repo, RemoteRepository: remote, RemoteName: "upstream", Configuration: source("binding")}))
	target := must(domain.NewCommitTarget(selected))
	resolution := ExactLocalResolutionData{Requested: target, Local: revision(repo, "2"), Binding: Some(binding), ObservedRemote: Some(selected), Observation: observation(repo, None[domain.WorktreeID]())}
	if _, e := NewExactLocalResolution(resolution); e != nil {
		t.Fatal(e)
	}
	resolution.Local = revision(repo, "3")
	if _, e := NewExactLocalResolution(resolution); e == nil {
		t.Fatal("latest endpoint substitution")
	}
	all := []EffectFacet{ObjectAcquisition, Recovery, WorktreeBytes, Index, LocalRefsHead, LocalConfiguration, RemoteRefsPR}
	d := ReconcileRequestData{Operation: operation(2), OriginalOperation: operation(1), Repository: repo, Kind: PushMutation, PriorEffects: emptyEffects(), Facets: all}
	if _, e := NewReconcileRequest(d); e != nil {
		t.Fatal(e)
	}
	for _, f := range []EffectFacet{Storage, RuntimeResources, 0, 255} {
		d.Facets = []EffectFacet{f}
		if _, e := NewReconcileRequest(d); e == nil {
			t.Fatalf("forbidden reconciliation facet %d", f)
		}
	}
	d.Facets = []EffectFacet{LocalConfiguration, LocalConfiguration}
	if _, e := NewReconcileRequest(d); e == nil {
		t.Fatal("duplicate facet")
	}
	wrong := expected(w).Data()
	wrong.Worktree = Some(worktree(other))
	if _, e := NewGitExpectedState(wrong); e == nil {
		t.Fatal("mismatched expected worktree")
	}
}

func TestOpaqueJSONAndWholeDocumentRetention(t *testing.T) {
	for _, raw := range [][]byte{[]byte(`{"a":1,"\u0061":2}`), []byte(`{"x":{"a":1,"a":2}}`), []byte(`{"\ud800":1}`), {'"', 255, '"'}, []byte(strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)), make([]byte, MaxDocumentBytes+1)} {
		if _, e := NewOpaqueJSON(raw); e == nil {
			t.Fatalf("invalid opaque JSON admitted: %q", raw[:min(len(raw), 100)])
		}
	}
	original := []byte(`{"\ud83d\ude00":[true,null,{"safe":1}]}`)
	opaque := must(NewOpaqueJSON(original))
	original[0] = '['
	got := opaque.Bytes()
	got[0] = '['
	if opaque.Bytes()[0] != '{' {
		t.Fatal("opaque bytes shared")
	}
	if _, e := NewOpaqueJSON([]byte(`{"😀":1,"\ud83d\ude00":2}`)); e == nil {
		t.Fatal("surrogate-equivalent duplicate")
	}
	nested := must(NewJSONMembers([]JSONMember{must(NewJSONMember("future", opaque))}))
	def := must(NewSavedLaunchDefinition(SavedLaunchDefinitionData{Provider: "future-provider", Dir: PresentField(" exact dir "), Targets: PresentField([]string{"a", "b"}), Command: PresentField("custom"), UnknownMembers: nested}))
	entries := []SavedLaunchEntry{must(NewSavedLaunchEntry(SavedLaunchEntryData{Alias: " Exact Alias ", Definition: def}))}
	d := must(NewRunConfigDocument(RunConfigDocumentData{SchemaVersion: 1, Default: PresentField(" Exact Alias "), Launch: PresentField(entries), UnknownMembers: nested}))
	entries[0] = SavedLaunchEntry{}
	copy1, _ := d.Data().Launch.Value()
	definition := copy1[0].Data().Definition
	targets, _ := definition.Data().Targets.Value()
	targets[0] = "changed"
	copy1[0] = SavedLaunchEntry{}
	actual, _ := d.Data().Launch.Value()
	saved := actual[0].Data()
	kept, _ := saved.Definition.Data().Targets.Value()
	if kept[0] != "a" || saved.Alias != " Exact Alias " || !saved.Definition.Data().UnknownMembers.Entries()[0].Value().Equal(opaque) {
		t.Fatal("nested retention/copy")
	}
	invalidDoc := d.Data()
	invalidDoc.Default = NullField[string]()
	if _, e := NewRunConfigDocument(invalidDoc); e == nil {
		t.Fatal("known null default")
	}
	collision := must(NewJSONMembers([]JSONMember{must(NewJSONMember("launch", opaque))}))
	invalidDoc = d.Data()
	invalidDoc.UnknownMembers = collision
	if _, e := NewRunConfigDocument(invalidDoc); e == nil {
		t.Fatal("known member collision")
	}
	for _, field := range []StoredField[[]string]{AbsentField[[]string](), NullField[[]string](), PresentField([]string{})} {
		v := must(NewUserConfigDocument(UserConfigDocumentData{SchemaVersion: 1, StripPrefixes: field, UnknownMembers: unknowns()}))
		if v.Data().StripPrefixes.Presence() != field.Presence() {
			t.Fatal("prefix presence collapsed")
		}
	}
	if _, e := NewPreferencesDocument(PreferencesDocumentData{SchemaVersion: 1, LegacyFolders: NullField[[]LegacyStringPreference](), UnknownMembers: unknowns()}); e == nil {
		t.Fatal("null known map")
	}
}

func TestStorageRecoveryAndErrorFactsAreLossless(t *testing.T) {
	version := storageVersion(Preferences)
	id := must(NewRecoveryID("persisted-manifest-artifact-id"))
	subject := must(NewRecoverySubject(RecoverySubjectData{Store: Some(version.Store()), Family: Some(Preferences)}))
	typed := must(NewStorageRecoveryVersion(StorageRecoveryVersionData{Version: version}))
	record := must(NewRecoveryRecord(RecoveryRecordData{RecoveryID: id, Kind: RecoveryManifest, Layer: LayerPersistence, Subject: subject, Locator: "owned/recovery", Original: Some[RecoveryVersion](typed), Proposed: Some[RecoveryVersion](typed), NextAction: "Inspect retained bytes before any new action."}))
	detail := must(NewStorageRecovery(StorageRecoveryData{Record: record, Family: Preferences, Locator: "owned/recovery", Kind: Manifest, Identity: source("manifest-observation")}))
	effect := must(NewFacetEffect(FacetEffectData{Facet: Storage, State: AppliedVerified, RecoveryIDs: []RecoveryID{id}}))
	effects := must(NewEffectReport(EffectReportData{Facets: []FacetEffect{effect}}))
	result := must(NewStorageCommitResult(StorageCommitResultData{Outcome: CommittedDurabilityUncertain, ProposedVersion: Some(version), CurrentVersion: Some(version), PublicationKnown: true, Durability: DurabilityUncertain, Effects: effects, Recovery: []StorageRecovery{detail}, Diagnostics: []Diagnostic{diagnostic()}}))
	fake := func() (StorageCommitResult, error) { return result, errors.New("postcommit refresh failed") }
	facts, err := fake()
	if err == nil || !facts.Data().PublicationKnown {
		t.Fatal("result plus error lost publication")
	}
	recovery := must(NormalizeRecovery([]RecoveryRecord{record}, facts.Data().Recovery))
	if len(recovery) != 1 {
		t.Fatal("recovery duplication")
	}
	retained, _ := recovery[0].Data().StorageDetail.Value()
	if retained.Data().Identity != detail.Data().Identity || retained.Data().Record.Data().RecoveryID != id {
		t.Fatal("storage detail lost")
	}
	terminal := must(NewOperationTerminal(OperationTerminalData{OperationID: operation(1), Correlation: must(NewCorrelation(CorrelationData{})), Disposition: Failed, Effects: effects, Recovery: recovery, Diagnostics: []Diagnostic{diagnostic()}, CancellationRequested: true}))
	recovery[0] = NormalizedRecovery{}
	output := terminal.Data()
	output.Recovery[0] = NormalizedRecovery{}
	if !terminal.Data().Recovery[0].Valid() {
		t.Fatal("terminal recovery shared")
	}
	bad := detail.Data()
	bad.Locator = "other"
	if _, e := NewStorageRecovery(bad); e == nil {
		t.Fatal("recovery locator mismatch")
	}
	changed := record.Data()
	changed.NextAction = "inconsistent"
	different := must(NewRecoveryRecord(changed))
	if _, e := NormalizeRecovery([]RecoveryRecord{record, different}, nil); e == nil {
		t.Fatal("inconsistent ID duplicate")
	}
	badRecord := record.Data()
	badRecord.Original = Some[RecoveryVersion](must(NewSourceRecoveryVersion(SourceRecoveryVersionData{Version: source("bad-domain")})))
	bad = detail.Data()
	bad.Record = must(NewRecoveryRecord(badRecord))
	if _, e := NewStorageRecovery(bad); e == nil {
		t.Fatal("document version cast to source version")
	}
	dangling := must(NewFacetEffect(FacetEffectData{Facet: Storage, State: AppliedVerified, RecoveryIDs: []RecoveryID{must(NewRecoveryID("dangling"))}}))
	badCommit := result.Data()
	badCommit.Effects = must(NewEffectReport(EffectReportData{Facets: []FacetEffect{dangling}}))
	if _, e := NewStorageCommitResult(badCommit); e == nil {
		t.Fatal("dangling ID")
	}
	badCommit = result.Data()
	badCommit.Outcome = NotCommitted
	if _, e := NewStorageCommitResult(badCommit); e == nil {
		t.Fatal("known publication relabeled no commit")
	}
	fullRecovery := must(NormalizeRecovery(nil, []StorageRecovery{detail}))
	sharedOnly := must(NormalizeRecovery([]RecoveryRecord{record}, nil))
	intent := must(NewNavigationIntent(NavigationIntentData{Repository: remoteRepo("navigation"), Folder: "folder"}))
	envelope := must(NewOutcomeEnvelope(OutcomeEnvelopeData{Effects: effects, Recovery: fullRecovery}))
	navigation := must(NewSaveNavigationResult(SaveNavigationResultData{Intent: intent, EffectiveVersion: Some(version), Storage: result, Outcome: envelope}))
	incomplete := navigation.Data()
	incomplete.Outcome = must(NewOutcomeEnvelope(OutcomeEnvelopeData{Effects: effects, Recovery: sharedOnly}))
	if _, e := NewSaveNavigationResult(incomplete); e == nil {
		t.Fatal("result normalization lost StorageRecovery detail")
	}
	term := OperationTerminalData{OperationID: operation(2), Correlation: must(NewCorrelation(CorrelationData{})), Disposition: Succeeded, Result: Some[Result](navigation), Effects: effects, Recovery: sharedOnly}
	if _, e := NewOperationTerminal(term); e == nil {
		t.Fatal("terminal lost StorageRecovery detail")
	}
	term.Recovery = fullRecovery
	if _, e := NewOperationTerminal(term); e != nil {
		t.Fatal(e)
	}
}

func TestRuntimeOutputAndResultFamily(t *testing.T) {
	id := must(domain.NewSessionID(1))
	sequence := must(NewSessionSequence(4))
	bytes := []byte{0, 255, 'x'}
	chunk := must(NewSessionOutputChunk(SessionOutputChunkData{Stream: Stdout, Offset: 10, Bytes: bytes, Sequence: sequence}))
	bytes[2] = 'z'
	gap := must(NewOutputGap(OutputGapData{From: 4, To: 10}))
	result := must(NewSessionOutputResult(SessionOutputResultData{SessionID: id, Sequence: sequence, Chunks: []SessionOutputChunk{chunk}, RetainedStart: 10, End: 13, NextOffset: 13, Gap: Some(gap), Truncated: true}))
	copy := result.Data().Chunks[0].Data().Bytes
	copy[2] = 'y'
	if result.Data().Chunks[0].Data().Bytes[2] != 'x' {
		t.Fatal("output byte alias")
	}
	d := result.Data()
	d.NextOffset = 12
	if _, e := NewSessionOutputResult(d); e == nil {
		t.Fatal("incorrect next offset")
	}
	d = result.Data()
	d.Gap = Some(must(NewOutputGap(OutputGapData{From: 4, To: 9})))
	if _, e := NewSessionOutputResult(d); e == nil {
		t.Fatal("incorrect gap")
	}
	correlation := must(NewCorrelation(CorrelationData{}))
	request := must(NewCommandRequest(must(NewWriteInputCommand(WriteInputCommandData{SessionID: id, Bytes: []byte{1}})), correlation))
	outcome := must(NewOutcomeEnvelope(OutcomeEnvelopeData{Effects: emptyEffects()}))
	write := must(NewSessionWriteResult(SessionWriteResultData{SessionID: id, Sequence: sequence, AcceptedBytes: 1}))
	typed := must(NewWriteInputResult(WriteInputResultData{Write: write, Outcome: outcome}))
	terminal := must(NewOperationTerminal(OperationTerminalData{OperationID: operation(1), Correlation: correlation, Disposition: Succeeded, Result: Some[Result](typed), Effects: emptyEffects()}))
	if e := ValidateTerminalFor(request, terminal); e != nil {
		t.Fatal(e)
	}
	resize := must(NewResizeSessionResult(ResizeSessionResultData{Control: must(NewSessionControlResult(SessionControlResultData{SessionID: id, Sequence: sequence, Delivered: true})), Outcome: outcome}))
	dterm := terminal.Data()
	dterm.Result = Some[Result](resize)
	if e := ValidateTerminalFor(request, must(NewOperationTerminal(dterm))); e == nil {
		t.Fatal("wrong result family")
	}
}

func TestRuntimeLifecycleRetainsFailedAdmissionAndReplacementIdentity(t *testing.T) {
	w := worktree(local("runtime"))
	cwd := must(NewCwdObservation(CwdObservationData{Worktree: scope(w), ProjectIdentity: directory(DirectoryWindows), Source: source("cwd")}))
	geometry := must(NewGeometry(GeometryData{Rows: 24, Columns: 80}))
	argv := must(NewArgvExecution(ArgvExecutionData{Executable: "tool", Arguments: []string{"literal:arg"}}))
	env := must(NewEnvironmentPolicy(EnvironmentPolicyData{InheritBase: true}))
	invocation := must(NewInvocation(InvocationData{Execution: argv, Environment: env, Cwd: cwd, Terminal: Terminal, Geometry: geometry, Label: "label"}))
	if !invocation.Valid() {
		t.Fatal("invocation")
	}
	display := must(NewInvocationSummary(InvocationSummaryData{Label: "label", ExecutableDisplay: "tool", Cwd: cwd, AcceptedLocator: "C:/fixture", Terminal: Terminal, Geometry: geometry}))
	capabilities := must(NewSessionCapabilities(SessionCapabilitiesData{Output: true, Input: true, Resize: true, TreeStop: true, Restart: true}))
	id := must(domain.NewSessionID(1))
	seq := must(NewSessionSequence(1))
	pending := must(NewSessionCleanup(SessionCleanupData{State: CleanupPending}))
	output := must(NewOutputRange(OutputRangeData{}))
	data := SessionSnapshotData{SessionID: id, WorktreeID: w, StartOperation: operation(1), Display: display, Capabilities: capabilities, Phase: Starting, Cleanup: pending, Sequence: seq, OutputRange: output}
	starting := must(NewSessionSnapshot(data))
	failed := must(NewSessionStartResult(SessionStartResultData{Session: Some(starting), Effects: emptyEffects(), Diagnostics: []Diagnostic{diagnostic()}}))
	if !failed.Data().Session.Present() {
		t.Fatal("failed admitted start lost identity")
	}
	bad := failed.Data()
	bad.Established = true
	if _, e := NewSessionStartResult(bad); e == nil {
		t.Fatal("start without acquired cwd barrier")
	}
	data.Phase = Cleaned
	data.Cleanup = must(NewSessionCleanup(SessionCleanupData{State: CleanupComplete}))
	cleaned := must(NewSessionSnapshot(data))
	old := must(NewSessionStopResult(SessionStopResultData{Session: cleaned, CleanupComplete: true, Effects: emptyEffects()}))
	if _, e := NewSessionStopResult(SessionStopResultData{Session: starting, CleanupComplete: true, Effects: emptyEffects()}); e == nil {
		t.Fatal("root-only cleanup claimed")
	}
	data.SessionID = must(domain.NewSessionID(2))
	data.StartOperation = operation(2)
	data.RestartOf = Some(id)
	data.Phase = Running
	data.Cleanup = pending
	data.AcquiredCwd = Some(must(NewAcquiredCwd(AcquiredCwdData{Observation: cwd, ActualLocator: Some("C:/fixture")})))
	replacement := must(NewSessionStartResult(SessionStartResultData{Session: Some(must(NewSessionSnapshot(data))), Established: true, Effects: emptyEffects()}))
	restart := must(NewSessionRestartResult(SessionRestartResultData{Old: old, Replacement: Some(replacement)}))
	if !restart.Valid() {
		t.Fatal("replacement identity")
	}
	badRestart := restart.Data()
	badRestart.Old = must(NewSessionStopResult(SessionStopResultData{Session: starting, Effects: emptyEffects()}))
	if _, e := NewSessionRestartResult(badRestart); e == nil {
		t.Fatal("replacement before old barrier")
	}
	event := must(NewRuntimeEvent(RuntimeEventData{Sequence: must(NewRuntimeEventSequence(9)), SessionSequence: seq, SessionID: id, Kind: RuntimeCleaned, Snapshot: cleaned}))
	badEvent := event.Data()
	badEvent.SessionID = data.SessionID
	if _, e := NewRuntimeEvent(badEvent); e == nil {
		t.Fatal("foreign event snapshot")
	}
	residual := must(NewRuntimeResidual(RuntimeResidualData{SessionID: Some(id), Stage: EventTransfer, Detail: diagnostic()}))
	if _, e := NewRuntimeShutdownResult(RuntimeShutdownResultData{AdmissionClosed: true, Complete: true, Sessions: []SessionStopResult{old}, Residuals: []RuntimeResidual{residual}}); e == nil {
		t.Fatal("shutdown discarded ACK residual")
	}
}

func TestDiscoverySelectionAndSourceFacts(t *testing.T) {
	repo := local("discovery")
	w := worktree(repo)
	a := must(domain.NewLaunchPointID(w, "make", "project", "a"))
	b := must(domain.NewLaunchPointID(w, "make", "project", "b"))
	members := []MemberSelection{must(NewMemberSelection(MemberSelectionData{LaunchPointID: a, SourceVersion: source("makefile")})), must(NewMemberSelection(MemberSelectionData{LaunchPointID: b, SourceVersion: source("makefile")}))}
	ordered := must(NewOrderedMakeLaunch(OrderedMakeLaunchData{Members: members}))
	members[0] = members[1]
	read := ordered.Data().Members
	read[0] = read[1]
	if ordered.Data().Members[0].Data().LaunchPointID != a {
		t.Fatal("ordered selection changed")
	}
	other := worktree(local("other"))
	if _, e := NewResolveLaunchRequest(ResolveLaunchRequestData{Worktree: scope(other), Selection: ordered, Geometry: must(NewGeometry(GeometryData{Rows: 24, Columns: 80}))}); e == nil {
		t.Fatal("foreign launch worktree")
	}
	stash := must(domain.NewStashID(repo, must(domain.NewOID(strings.Repeat("2", 40)))))
	offset := time.FixedZone("recorded-offset", 2*60*60)
	at := time.Date(2020, 1, 2, 3, 4, 5, 0, offset)
	fact := must(NewStashFact(StashFactData{ID: stash, Parents: []domain.OID{must(domain.NewOID(strings.Repeat("3", 40))), must(domain.NewOID(strings.Repeat("4", 40)))}, Occurrence: source("occurrence"), AuthorName: "name", AuthorEmail: "mail@example.test", AuthorTime: Some(at), Message: "literal", Observation: observation(repo, None[domain.WorktreeID]())}))
	kept, _ := fact.Data().AuthorTime.Value()
	if kept != at {
		t.Fatal("source timestamp normalized")
	}
	selected := must(NewStashParent(StashParentData{Index: 1}))
	tree := must(domain.NewOID(strings.Repeat("5", 40)))
	comparison := must(NewStashPatchComparison(StashPatchComparisonData{Stash: stash.OID(), Base: Some(fact.Data().Parents[0]), IndexParent: Some(fact.Data().Parents[1]), View: selected, FromTree: Some(tree)}))
	if oid, p := comparison.Data().FromTree.Value(); !p || oid != tree {
		t.Fatal("tree OID lost")
	}
}

func TestGitKnownPushAndFailedUpstreamKeepSeparateEffects(t *testing.T) {
	repo := local("push")
	remote := remoteRepo("host/owner/repo")
	obs := observation(repo, None[domain.WorktreeID]())
	binding := must(NewRemoteBinding(RemoteBindingData{LocalRepository: repo, RemoteRepository: remote, RemoteName: "origin", Configuration: source("remote-config")}))
	destination := must(domain.NewBranchID(remote, domain.RemoteHead, "main"))
	localRevision := revision(repo, "1")
	remoteRevision := revision(remote, "1")
	pushed := must(NewFacetEffect(FacetEffectData{Facet: RemoteRefsPR, State: AppliedVerified, PostObservation: Some(obs.Data().ID)}))
	upstream := must(NewFacetEffect(FacetEffectData{Facet: LocalConfiguration, State: VerifiedNoTargetChange, PostObservation: Some(obs.Data().ID)}))
	remoteEffects := must(NewEffectReport(EffectReportData{Facets: []FacetEffect{pushed}}))
	upstreamEffects := must(NewEffectReport(EffectReportData{Facets: []FacetEffect{upstream}}))
	effects := must(NewEffectReport(EffectReportData{Facets: []FacetEffect{pushed, upstream}}))
	outcome := must(NewPushed(PushedData{Source: localRevision, Destination: destination, Binding: binding, ObservedRemote: Some(remoteRevision), RemoteEffect: remoteEffects, UpstreamEffect: upstreamEffects}))
	facts := must(NewGitMutationResult(GitMutationResultData{Operation: operation(1), Kind: PushMutation, PlanVersion: source("plan"), Outcome: outcome, Observation: Some(obs), Transport: must(NewCommandTransportOutcome(CommandTransportOutcomeData{})), Effects: effects, Diagnostics: []Diagnostic{diagnostic()}}))
	call := func() (GitMutationResult, error) {
		return facts, errors.New("upstream configuration refused after push")
	}
	result, err := call()
	if err == nil || result.Data().Effects.Data().Facets[0].Data().State != AppliedVerified || result.Data().Effects.Data().Facets[1].Data().State != VerifiedNoTargetChange {
		t.Fatal("independent effects lost")
	}
	missing := facts.Data()
	missing.Observation = None[GitObservation]()
	if _, e := NewGitMutationResult(missing); e == nil {
		t.Fatal("dangling post-observation accepted")
	}
	commandResult := must(NewPushResult(PushResultData{Git: facts, RemoteEffect: remoteEffects, UpstreamEffect: upstreamEffects, Outcome: must(NewOutcomeEnvelope(OutcomeEnvelopeData{Effects: effects, Diagnostics: []Diagnostic{diagnostic()}}))}))
	terminal := must(NewOperationTerminal(OperationTerminalData{OperationID: operation(1), Correlation: must(NewCorrelation(CorrelationData{})), Disposition: Failed, Result: Some[Result](commandResult), Effects: effects, Diagnostics: []Diagnostic{diagnostic()}}))
	if !terminal.Data().Result.Present() {
		t.Fatal("failed terminal dropped typed result")
	}
	wrong := facts.Data()
	wrong.Kind = CommitMutation
	if _, e := NewGitMutationResult(wrong); e == nil {
		t.Fatal("cross-kind mutation outcome")
	}
}
