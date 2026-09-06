# BC--Application--Git

State: DRAFT
Version: 1.0.0
Parent Issue: #55 under #21; accepted design #52 / PR #54
Applies to: complete v0.4 refactor, SLC-01..08, worktree scope for SLC-09..11 and cancellation in SLC-12
Design authority: merged `4a42222f7bfedc1d80693effbb25a1a82fcff65e`; technical acceptance `664f0c051344e3abdfd7d3c5698e4fbd3f584a83`
Supersedes: none. Freeze authority: none while DRAFT.

## G1. Responsibility and dependency direction

Application owns operation/confirmation identity, selected exact targets, active
worktree context, configured-target interpretation, fetch/pull/deploy/pop/commit
workflow sequencing, local/remote association, canonical projections and retry
admission. Git owns local repository/worktree/ref/index/stash/history/diff facts,
native Git mutation mechanics, common-repository coordination and retained
recovery. Git supplies facts; it does not choose which worktree the user meant.

`internal/git` implements `internal/application/ports`; it may import the
Application API semantic values, Domain and approved low-level libraries.
Application consumes ports and imports no concrete adapter. API imports no ports.
Git imports/calls no GitHub, Runtime, Discovery, Persistence, State, View,
coordinator or legacy worktree/graph/diff/process implementation. Composition
alone constructs/wires adapters. Short-command transport and native recovery
implementation remain private to Git; no new shared process/filesystem layer is
created. No `worktree.*` or other concrete DTO enters a public signature.

[BoundaryTypes--001](BoundaryTypes--001.md) defines shared observation, error,
effect, immutable-copy and identity types. API [A1..A5](../Design/CR-%2321/API--001.md)
and the **superseding protected-publication addendum** in
[Git-Safety](../Design/CR-%2321/Feasibility/Git-Safety.md) are normative here.
Earlier live switch/read-tree/stash-cleanup recommendations in that evidence
file are rejected history. Record definitions below belong to `application/api`;
interfaces, typed plan handles and validating issuer mechanism belong to
`application/ports`. `Optional<T>` denotes explicit absence/presence; closed
alternatives are validated sum values, never optional-field bags or `any`.

## G2. Concrete port inventory

```go
// package ports
type GitFacts interface {
    ResolveLocal(context.Context, api.ResolveLocalRequest) (api.ResolveLocalResult, error)
    ListWorktrees(context.Context, api.ListWorktreesRequest) (api.ListWorktreesResult, error)
    ObserveStatus(context.Context, api.ObserveStatusRequest) (api.ObserveStatusResult, error)
    ResolveExact(context.Context, api.ResolveExactRequest) (api.ResolveExactResult, error)
    ListRefs(context.Context, api.ListRefsRequest) (api.ListRefsResult, error)
    ListStashes(context.Context, api.ListStashesRequest) (api.ListStashesResult, error)
    ReadStashPatch(context.Context, api.ReadStashPatchRequest) (api.ReadStashPatchResult, error)
    MergeBase(context.Context, api.MergeBaseRequest) (api.MergeBaseResult, error)
    ReadCommits(context.Context, api.ReadCommitsRequest) (api.ReadCommitsResult, error)
    ReadGraph(context.Context, api.ReadGraphRequest) (api.ReadGraphResult, error)
    ReadDiff(context.Context, api.ReadDiffRequest) (api.ReadDiffResult, error)
}
type GitMutations interface {
    PrepareCreate(context.Context, GitPrepareContext, api.PrepareCreateRequest) (CreatePlan, api.GitPreparationResult, error)
    PrepareRetarget(context.Context, GitPrepareContext, api.PrepareRetargetRequest) (RetargetPlan, api.GitPreparationResult, error)
    PrepareStage(context.Context, GitPrepareContext, api.PrepareStageRequest) (StagePlan, api.GitPreparationResult, error)
    PrepareCommit(context.Context, GitPrepareContext, api.PrepareCommitRequest) (CommitPlan, api.GitPreparationResult, error)
    PrepareRestore(context.Context, GitPrepareContext, api.PrepareRestoreRequest) (RestorePlan, api.GitPreparationResult, error)
    PrepareStash(context.Context, GitPrepareContext, api.PrepareStashRequest) (StashPlan, api.GitPreparationResult, error)
    PrepareBranch(context.Context, GitPrepareContext, api.PrepareBranchRequest) (BranchPlan, api.GitPreparationResult, error)
    PreparePush(context.Context, GitPrepareContext, api.PreparePushRequest) (PushPlan, api.GitPreparationResult, error)
    ExecutePrepared(context.Context, PreparedGitPlan, ExecutionApproval) (ExecutedGitMutation, error)
    ReleasePlan(PreparedGitPlan) error
    Fetch(context.Context, api.FetchRequest) (api.FetchResult, error)
    Reconcile(context.Context, api.ReconcileRequest) (api.ReconcileResult, error)
}
```

PreparedGitPlan is the closed sum of CreatePlan, RetargetPlan, StagePlan,
CommitPlan, RestorePlan, StashPlan, BranchPlan and PushPlan. Its package-private
marker and opaque immutable fields prohibit arbitrary plan implementations,
foreign adapters and operation-kind reuse. Ports' validating issuer mechanism
lets a conforming adapter produce handles without exposing native objects or
adapter DTOs. Handle equality includes issuer lifetime, assigned OperationID,
plan kind, version and unpredictable token. Plan existence is not user approval.

```go
// package ports; Optional is explicit absence/presence notation.
type GitPrepareContext struct {
    OperationID api.OperationID
    OriginPlan  api.Optional[PreparedGitPlan]
    Predecessor api.Optional[GitMutationReceipt]
}
type ExecutedGitMutation struct {
    Facts   api.GitMutationResult
    Receipt api.Optional[GitMutationReceipt]
}
```

GitMutationReceipt is ports-owned opaque immutable issuer/group/step evidence,
not an API DTO or a public assertion of success. It names one exact executed
plan and its recorded before/after facet versions. It contains no native handle.
Only Git's validated completed-step registry can verify it. Wrappers keep all
private plan/receipt types outside API; shared value aliases never add API->ports
imports. A result with a nonnil error may retain a receipt for its independently
known effects, but G6 restricts which verified predecessor can authorize a next
step. Facts always survive wrapper/error handling.

Every result contains `Observation: Optional<GitObservation>`,
`Diagnostics: []Diagnostic` and `Transport: CommandTransportOutcome`. Results
with possible preparation/fetch/mutation effects also contain `Effects:
EffectReport`, `CancellationRequested: bool`, `Recovery: []RecoveryRecord` and
`ReconciliationRequired: bool`. A nonnil error supplements this result; callers
must retain its independently valid facts/effects/recovery. Zero-value success,
generic `Execute(string, any)`, and result loss on error are forbidden.

## G3. Identity, observation and exact endpoint records

| API record | Concrete fields / invariant |
|---|---|
| GitObservation | `ID: ObservationID`, `Repository: domain.RepositoryID`, `Worktree: Optional<domain.WorktreeID>`, `Interval: ObservationInterval`, `Version: SourceVersion`, `Completeness: Completeness`; local common scope and exact observed source are explicit. Version is opaque equality, not a global repository transaction or application request generation. |
| GitCapabilities | `ObjectFormat: domain.ObjectFormat`, `RefBackend: Files/Reftable/Other`, `GitVersion: string`, `Profile: string`, `Capabilities: []GitCapabilityFact`; every fact names supported/unsupported operation and prerequisite. It includes normal-index, retained-file/index publication, exact stash delete, native symref and path/filesystem support. |
| LocalRepositoryFacts | `Repository: domain.RepositoryID`, `CommonDirectory: string`, `Worktrees: []WorktreeFacts`, `Remotes: []RemoteBinding`, `Capabilities: GitCapabilities`, `Observation: GitObservation`. Paths are canonical locators/evidence, not Domain identity supplied by the caller. |
| WorktreeFacts | `ID: domain.WorktreeID`, `Scope: Optional<WorktreeScope>`, `Head: Optional<domain.Head>`, `Primary: bool`, `Current: bool`, `Availability: WorktreeAvailability`, `Occupancy: []BranchOccupancy`, `Observation: GitObservation`. Availability is Available, Locked(reason), Prunable(reason), Missing or Unresolved(diagnostic). Scope/Head are present only with actual validated root/Head observations; missing/unresolved state cannot fabricate either. Current means adapter-start context, not Active. |
| BranchOccupancy | `Branch: domain.BranchID`, `Worktrees: []domain.WorktreeID`, `Observation: GitObservation`; incomplete/unresolved inventory cannot prove vacancy. |
| GitPath | exact worktree-relative path bytes with validating constructor; no empty/NUL/rooted/traversal or escaping components. Spaces, colon, leading dash, non-ASCII and permitted newline bytes remain literal. Full path operands use literal/NUL-safe native forms, never trimming or shell interpolation. |
| FileState | `Path: GitPath`, `State: Absent/Present`, and for Present `ObjectIdentity: SourceVersion`, `Kind: Regular/Symlink/Directory/Other`, `Mode: uint32`, `Content: SourceVersion`, `LinkTarget: Optional<string>`, `ParentIdentity: SourceVersion`. Content digest covers full bytes; identity covers the actual filesystem object/parent, not size/mtime alone. |
| IndexEntryFact | `Path: GitPath`, `Stage: uint8`, `Object: domain.OID`, `Mode: uint32`, `SemanticFlags: []IndexFlag`; supported flags/extension state is explicit. Null protocol OIDs are represented by an absent entry, never Domain zero Revision. |
| ChangeFact | `Path: GitPath`, `OldPath: Optional<GitPath>`, `Kind: Added/Modified/Deleted/Renamed/Copied/TypeChanged/Untracked/Unmerged`, `IndexEntries: []IndexEntryFact`, `WorktreeState: FileState`; rename old/new identities remain bound. |
| StatusFacts | `Worktree: WorktreeFacts`, `Changes: []ChangeFact`, `IndexVersion: SourceVersion`, `WorktreeVersion: SourceVersion`, `ConfigurationVersion: SourceVersion`, `Upstream: UpstreamFact`, `Observation: GitObservation`; staged, unstaged, untracked and conflicted causes remain separate, never one clean boolean. |
| UpstreamFact | closed None, NotApplicable, Gone(binding/ref/evidence), Unresolved(binding/ref/diagnostic), or Resolved(binding/remote-branch/cached-local-ref/exact-local-endpoints/comparison/freshness). NotApplicable covers detached/unborn where comparison requires an established local commit. Gone requires conclusive scoped evidence; failed lookup is Unresolved. |
| RevisionComparison | `Left: domain.Revision`, `Right: domain.Revision`, `Ahead: Optional<uint64>`, `Behind: Optional<uint64>`, `Observation: GitObservation`; endpoints are exact verified commits in one local object scope. Counts are both present only when proven, never default zero on failure. |
| FetchGeneration | opaque nonzero adapter-lifetime monotonic identifier keyed by common repository plus RemoteBinding/ref scope; wrap refuses new generation. It is distinct from QueryGeneration and RemoteObservationID. |
| FetchFreshness | `Kind: Unknown/Cached/Refreshed`, `Binding: RemoteBinding`, `RefScope: []GitRefLocator`, `Generation: Optional<FetchGeneration>`, `Observation: GitObservation`; only successful completed exact-scope fetch proves Refreshed for that generation. Remote-tracking refs remain local cached facts after it. |
| GitRefLocator | closed LocalBranch(domain.BranchID), LocalTag(exact full native ref), CachedRemote(RemoteBinding, exact local tracking ref), or RemoteRef(RemoteBinding, exact remote source ref). It is locator metadata, not Revision authority. Git validates/refspec-escapes native syntax independently. |
| RefFact | `Locator: GitRefLocator`, `Revision: Optional<domain.Revision>`, `SymbolicTarget: Optional<GitRefLocator>`, `Freshness: Optional<FetchFreshness>`, `Observation: GitObservation`; annotated tag identity and peeled commit are distinguished; unresolvable/missing object is explicit. |
| ExactLocalResolution | `Requested: domain.ExactTarget`, `Local: domain.Revision`, `Locator: Optional<GitRefLocator>`, `Binding: Optional<RemoteBinding>`, `ObservedRemote: Optional<domain.Revision>`, `Observation: GitObservation`; remote-to-local association is verified and retained, never created from equal OID bytes alone. |
| StashFact | `ID: domain.StashID`, `Parents: []domain.OID`, `Occurrence: SourceVersion`, `DisplayPosition: uint64`, `Message: string`, `Origin: Optional<StashOrigin>`, `Observation: GitObservation`; StashOrigin has descriptive worktree/branch/legacy managed metadata. Position, label, timestamp and managed marker are never destructive authority. |

Domain ExactTarget denotes the accepted CommitTarget(Revision),
BranchTarget(BranchID, ExpectedRevision), PullRequestTarget(PRID,
ExpectedHeadRevision) union. Git does not read PR API data; the PR variant's
scope/expected revision is preserved only as provenance. Application supplies
the explicit generic native remote-ref locator/binding needed for fetch.

Git mints LocalCommon RepositoryID from the physically resolved common directory.
Primary and linked administrative worktree keys are distinct and stable across
retargeting. Windows binds final handle path and volume identity, preserving case
of case-sensitive components; Unix resolves links and preserves path bytes.
Blanket case folding is forbidden. Clones/relocation create separate local scopes;
remote URL equality cannot merge them. WorktreeScope binds the validated root
locator/object observation to WorktreeID for Discovery/Runtime/Storage consumers;
it is not a permanent open native handle.

WorktreeFacts always retains its established administrative WorktreeID when a
registered path is missing or cannot be inspected. Optional Scope is present
only when that actual root was acquired and validated; Optional Head is present
only when its native state was independently established. A missing root may
still have an independently readable administrative Head, but neither field is
fabricated to satisfy inventory shape. Unidentifiable malformed entries produce
inventory diagnostics/Unknown completeness without a fake WorktreeFacts row.
Unavailable Head is API observation absence plus diagnostic, not a fourth
Domain Head variant. Activation and scope-dependent consumers require an
available validated Scope; exact-Head operations also require the appropriate
present Head. Successful mutation postconditions likewise require their proven
scope/Head facts rather than treating these options as optional verification.

RemoteBinding ties a local common repository/configured remote to an explicit
remote RepositoryID and exact observed URL/refspec mapping/version. Replacement
of remote URL, host or mapping invalidates dependent observations/plans; `origin`
alone proves no remote identity. API Remote scopes use GitHub's agreed canonical
host/owner/name rules; Git validates local transport mapping and App associates
the two independently observed scopes. Neither Git nor GitHub calls the other.

Attached/Detached Heads contain established exact local Revisions; Unborn contains
only a local BranchID. Inventory/status remain usable for unborn repositories.
An operation requiring an established departure/target must explicitly refuse
its absence, never construct a protocol-zero Revision or treat unborn as detached.

## G4. Read requests and result payloads

| Request | Exact fields | Matching result payload beyond common envelope |
|---|---|---|
| ResolveLocalRequest | `Locator: string` explicit filesystem entry scope | ResolveLocalResult: `Repository: Optional<LocalRepositoryFacts>`; physically invalid/unresolvable scope refuses. |
| ListWorktreesRequest | `Repository: domain.RepositoryID` | ListWorktreesResult: `Worktrees: []WorktreeFacts`; observation includes complete/unknown inventory status. |
| ObserveStatusRequest | `Worktree: domain.WorktreeID` | ObserveStatusResult: `Status: Optional<StatusFacts>`; inspect with no index refresh/write. |
| ResolveExactRequest | `Repository: domain.RepositoryID`, `Target: domain.ExactTarget`, `Locator: Optional<GitRefLocator>`, `ExpectedSource: Optional<SourceVersion>` | ResolveExactResult: `Resolution: Optional<ExactLocalResolution>`; verifies commit existence/type and equality in intended database. No implicit fetch. |
| ListRefsRequest | `Repository: domain.RepositoryID`, `Kinds: []LocalBranch/LocalTag/CachedRemote`, `Page: PageRequest` | ListRefsResult: `Refs: []RefFact`, `Page: PageInfo`; locally cached remote facts have freshness. |
| ListStashesRequest | `Repository: domain.RepositoryID`, `Page: PageRequest` | ListStashesResult: `Stashes: []StashFact`, `Page: PageInfo`; repository-global across linked worktrees; duplicate OIDs remain observable. |
| ReadCommitsRequest | `Endpoint: domain.Revision`, `Traversal: FirstParent/AllParents`, `Page: PageRequest` | ReadCommitsResult: `Commits: []CommitFact`, `Page: PageInfo`, `Endpoint: domain.Revision`; root/parent relationships and verbatim messages retained. |
| ReadGraphRequest | `Repository: domain.RepositoryID`, `Roots: []domain.Revision`, `Filter: GraphFilter`, `Page: PageRequest` | ReadGraphResult: `Commits: []CommitFact`, `Refs: []RefFact`, `Heads: []WorktreeFacts`, `Page: PageInfo`; exact root/traversal continuation, no PR joins or lane state. |
| MergeBaseRequest | `Left: domain.Revision`, `Right: domain.Revision` | MergeBaseResult: closed Unique(`Base: domain.Revision`), NoCommonAncestor or Ambiguous(`Candidates: []domain.Revision`), plus both input endpoints. No silent first-candidate choice. |
| ReadDiffRequest | `Comparison: GitComparison`, `Paths: []GitPath`, `Limits: PatchLimits` | ReadDiffResult: `Comparison: GitComparison`, `Patch: PatchFacts`; exact actual endpoints/version retained. |
| ReadStashPatchRequest | `Stash: domain.StashID`, `View: StashPatchView`, `Paths: []GitPath`, `Limits: PatchLimits` | ReadStashPatchResult: `Stash: domain.StashID`, `Parents: []domain.OID`, `Comparison: StashPatchComparison`, `Patch: PatchFacts`; inspect full OID even after reflog positions shift. |

CommitFact contains `Revision: domain.Revision`, `Parents: []domain.Revision`,
`AuthorName: string`, `AuthorEmail: string`, `AuthorTime: time.Time`,
`CommitterTime: time.Time`, `Message: string`; metadata bytes are semantic facts,
never rewritten to add annotations. GraphFilter is a copied closed All or
ReachableFromRoots selection plus optional literal path filter. PageRequest and
PageInfo bind source, exact roots, traversal/filter and cursor/skip; later pages
cannot silently switch endpoints or reset accumulated graph context. Explicit
positive limits and Complete/More/Unknown apply; limits never imply absence.

GitComparison is closed CommitParent(commit, selected parent index or Root),
CommitPair(from,to), IndexToWorktree(worktree, expected index/worktree versions)
or HeadToIndex(worktree, expected Head/index versions). Root uses native empty
tree comparison without a fake Domain commit. Arbitrary revision expressions,
raw `A...B` strings or labels are not authority. All commit endpoints are verified
local Revisions in the same object database. PatchLimits contains positive
`MaxBytes: uint64`, `MaxFiles: uint32`; it is capped by adapter construction
bounds. PatchFacts contains copied raw patch bytes, `Files: []DiffFileFact`,
`Truncated: bool`, optional known original byte count, and actual returned bytes.
DiffFileFact contains exact path/optional old path, closed change kind, optional
added/deleted line counts and `Binary: bool`. Binary/rename/truncated/invalid text
remain facts; Git inserts no display labels or truncation prose into patch bytes.

StashPatchView is closed BaseToWorktree, BaseToIndex, IndexToWorktree,
Untracked or Parent(index). StashPatchComparison returns exact selected stash
OID, verified base/index/untracked-parent OIDs where present, selected view and
actual tree endpoints. Validate native stash-like parent structure; missing
untracked parent is a known empty/absent view, malformed structure is diagnostic.
The read never authorizes apply/pop/drop and never substitutes current position.

For PullRequestDiff, Application observes explicit remote PR base/head, preserves
selected expected head and either selected exact base or the request's explicit
observe-current-base choice. It Fetches/resolves both into private verified local
roots, calls MergeBase, then ReadDiff(CommitPair(unique merge-base, exact head)).
The public result retains remote/local provenance and all three exact endpoints.
Missing/deleted/fork/mismatched endpoints or ambiguous merge base refuse instead
of using a branch-name fallback. Git knows no PR service or cross-source join.

## G5. Mutation request and plan records

Every Prepare call includes GitPrepareContext with the Application-assigned
OperationID; its API request includes `Expected: GitExpectedState` and its
row-specific payload. GitExpectedState has
`Repository: domain.RepositoryID`, `Worktree: Optional<domain.WorktreeID>`,
`Observation: SourceVersion`, `Head: Optional<domain.Head>`, `Index: Optional<SourceVersion>`,
`WorktreeState: Optional<SourceVersion>`, `Configuration: SourceVersion` and
`Inventory: SourceVersion`. Required members are enforced by the particular
request, not optional safety checks. Restore/retarget/stash/commit bind full
relevant Head/index/content/occupancy; repository-only drop binds stash occurrence
and storage state. A path name by itself is not a state precondition.

| Prepare request | Operation-specific fields and supported alternatives |
|---|---|
| PrepareCreateRequest | `Destination: string`, `Target: ExactLocalResolution`, `Mode: Detached/CreateNewBranch(domain.BranchID)`; vacant path and validated parent identity, no replacing existing destination. |
| PrepareRetargetRequest | `Worktree: domain.WorktreeID`, `Target: ExactLocalResolution`, `Mode: Detach/AttachExisting(domain.BranchID)/CreateNewBranch(domain.BranchID)/FastForward(domain.BranchID, from, to)`, `Purpose: Retarget/Deploy/Pull`, `DirtyPolicy: Refuse/OfferStashThenDeploy`; selected target/version never becomes latest-ref. |
| PrepareStageRequest | `Worktree: domain.WorktreeID`, `Selection: ExactPaths([]GitPath)/AllObserved`, `Action: Stage/Unstage`; AllObserved explicitly binds the observed changes; no implicit commit. |
| PrepareCommitRequest | `Worktree: domain.WorktreeID`, `Message: string`, `IndexPolicy: ExistingIndex/ObservedStageAll`; literal message and explicit policy, no shell or implied default from UI key. |
| PrepareRestoreRequest | `Worktree: domain.WorktreeID`, `Paths: []GitPath`; source is the confirmed index, never silently HEAD; binds exact selected path states/index source blobs. |
| PrepareStashRequest | closed Create(`Worktree`, literal `Message`, `IncludeUntracked: bool`), Apply or Pop(`Worktree`, `Stash: domain.StashID`, `Occurrence: SourceVersion`), or Drop(`Stash: domain.StashID`, `Occurrence: SourceVersion`); common Expected is retained in every alternative. Pop prepares a sequence-root contract; Application executes its derived Apply -> exact Drop steps. |
| PrepareBranchRequest | `Worktree: domain.WorktreeID`, `Name: domain.BranchID`, `Start: domain.Revision`, `Checkout: bool`; branch create-only, never rewind/reuse an existing name. Optional checkout follows the separately validated native attachment protocol. |
| PreparePushRequest | `Worktree: domain.WorktreeID`, `Source: domain.Revision`, `Destination: domain.BranchID`, `Binding: RemoteBinding`, `SetUpstream: Optional<UpstreamSetup>`; Source local, Destination remote, explicit verified association. UpstreamSetup contains local BranchID and expected current upstream configuration version/value. |

GitPreparationResult contains `Summary: Optional<MutationPlanSummary>` and the
common result envelope. Successful Prepare returns exactly one valid typed plan
and its corresponding immutable summary. Failure yields no executable plan.
MutationPlanSummary contains OperationID, `Kind: GitMutationKind`, `PlanVersion:
SourceVersion`, exact worktree/repository/target/Head/stash identities as applicable,
`Expected: GitExpectedState`, `Paths: []PlannedPathEffect`, expected ref/index
changes, `Choices: []GitPlanChoice`, recovery behavior and required capabilities.
PlannedPathEffect carries exact old FileState and desired Present/Absent state
with source blob/mode/version. It is the same immutable intent represented by the
opaque handle; human-safe summary cannot omit a material destructive identity.
Choices are Proceed, Cancel or StashThenDeploy as allowed for that plan.

ExecutionApproval is ports-owned opaque evidence issued by the coordinator for
the original assigned OperationID, typed plan identity/version, exact summary
digest and chosen allowed alternative. A confirmation-required plan additionally
binds the single consumed ConfirmationID. Non-confirming actions still bind
the validated original request/plan; a raw bool never grants approval. Git checks
binding/version/issuer/kind; it does not issue/consume UI confirmation or decide
what to display. A stale/newly prepared plan never inherits an old approval.

Prepare performs bounded immutable preflight only. It retains no native lock,
OS handle, subprocess, scratch repository, file descriptor or resource lease
across user delay. In-memory copied preconditions are bounded by admitted
operations (maximum 64, no unbounded plan table). Excess bytes/plans return Busy
or explicit limit refusal before admitting the plan. Execute reacquires guards,
revalidates, then performs native scratch preparation and publication. No native
lock is held while waiting for ConfirmationRequested/Confirm.

A plan is single-use: Issued -> Executing -> Consumed, or Issued -> Released.
Each handle has `Role: Executable/SequenceRoot`. Dirty StashThenDeploy, Pop and
ObservedStageAll commit use a SequenceRoot that binds the complete planned
sequence; it cannot be passed directly to ExecutePrepared. Application derives
the allowed executable steps through GitPrepareContext as specified in G6.
Atomic admission rejects stale, foreign, already executing/consumed/released
handles and operation-kind/version mismatch before target change. ReleasePlan is
idempotent for the exact issued handle and frees only ephemeral plan memory;
release while executing marks disposal pending and cannot abandon execution.
Application calls it on confirmation cancel, terminal completion and shutdown.
Releasing the sequence-root disposes its remaining unused child handles/receipts
after admitted execution ends. All children/step receipts share one bounded
operation reservation; no per-step admission expansion or unbounded history.
Recovery journals/retained originals are independent and survive plan release.

## G6. Result facets, sequencing and active context

GitMutationResult carries the common mutation envelope, `Operation: OperationID`,
`Kind: GitMutationKind`, `PlanVersion: SourceVersion`, `Steps: []GitCompletedStep`
and `Outcome: GitMutationOutcome`. GitCompletedStep has closed step kind,
exact target, observed effect and post-observation. GitMutationOutcome is closed:

| Outcome | Exact payload |
|---|---|
| WorktreeCreated | verified WorktreeFacts and ExactLocalResolution; partial registration instead remains a partial effect with recovery. |
| WorktreeRetargeted | verified WorktreeFacts, prior Head and requested ExactLocalResolution. |
| IndexChanged | verified StatusFacts and Stage/Unstage operation kind. |
| CommitCreated | exact new Revision, verified Head/index facts and explicit staged-index effect. |
| TrackedRestored | exact selected paths, resulting StatusFacts and retained-original recovery records. |
| StashCreated | exact StashID and cleanup/resulting StatusFacts. |
| StashCreatedCleanupRefused | exact stored StashID, refusal/partial cleanup facets, current StatusFacts where valid and recovery. |
| StashApplied | exact StashID, resulting StatusFacts, `IndexRestored: bool`; stash retention explicit. |
| AppliedWithConflicts | exact StashID, published conflict paths/index stages and retained stash/recovery. |
| StashDropped | exact StashID/selected occurrence, surviving stash observation and separate ref cleanup result. |
| BranchCreated | exact local BranchID/Revision and optional verified checked-out WorktreeFacts. |
| Pushed | exact local source Revision, explicit remote branch/observed remote tip evidence, separate upstream-configuration effect. |
| Refused | Diagnostic plus independently known facets; reason does not imply global unchanged state. |
| PartialMutation / MutationIndeterminate | independently known post-facts, failed/unclassified steps, diagnostic, recovery and required reconciliation. |

EffectReport independently classifies object/recovery creation, worktree bytes,
index, local refs/HEAD, remote refs and upstream configuration as NotStarted,
VerifiedNoTargetChange, AppliedVerified, Partial or Indeterminate. No ordinal
success calculation or single HEAD equality can prove rollback. Canceled intent
is separate. Known commit/stash/push followed by failed projection refresh retains
that known effect; its projection becomes unavailable instead of reporting no
change. Reconcile observations supplement the original terminal record; they
do not emit another terminal for the original operation or retroactively claim
causal ownership of external actions.

Application sequences Fetch -> ObserveStatus/exact upstream -> prepared
FastForward for Pull. Git never runs a broad pull that reselects mutable upstream.
Application sequences explicit StageAll -> Commit for StageAllAndCommit, preserving
a known staged index if commit/hooks/signing fail. PrepareCommit's ObservedStageAll
policy records this explicit request; it does not authorize an implicit stage-all
inside an ordinary ExistingIndex commit. Dirty deploy binds its original dirty
snapshot and target to one confirmation, obtains the exact created stash, then
revalidates before deployment; created stash/recovery is returned after later
failure. A required second decision ends with a new explicit request requirement.

Pop is exact apply (including index) then exact drop only after verified apply.
Conflicted/partial/indeterminate/canceled apply retains the stash and prevents
automatic deletion. A successful apply followed by a missing, unsupported, stale
or failed drop becomes Application's `AppliedStashRetained` result with both
facets; it is never an apply failure inviting duplicate replay. Continuation
authority follows this concrete ports protocol:

1. Initial Prepare has absent OriginPlan/Predecessor and produces a root bound to
   the whole immutable request. Application consumes any required confirmation
   once. The root fixes exact target, dirty snapshot/stash occurrence, allowed
   choice, complete expected step kinds and operation identity.
2. For StashThenDeploy, Application calls PrepareStash(Create) with the original
   RetargetPlan as OriginPlan and absent Predecessor, then ExecutePrepared with
   the original root-bound approval. Git permits only the declared stash capture
   of the original dirty snapshot. For Pop, the first derived step is Apply of
   the original exact StashID/occurrence. For ObservedStageAll commit it is the
   exact observed StageAll action. No other cross-kind derivation is allowed.
3. ExecutePrepared returns Facts plus an opaque GitMutationReceipt. Git records
   exact planned source versions, completed changed facets and observed post-
   versions; Application then decides whether to request the next declared step.
4. Application passes the same OriginPlan and exact receipt as Predecessor to
   PrepareRetarget, PrepareStash(Drop) or PrepareCommit(ExistingIndex), respectively.
   Git validates issuer/operation/group/kind/step order/single-use predecessor,
   checks the recorded required success (no partial/conflicted/indeterminate or
   canceled predecessor), and advances only facets changed by that verified
   planned step. It rechecks current scope, target and all unrelated versions.
   A new external edit, even matching an intended broad status, refuses; no fresh
   current-state plan can inherit the original approval.
5. Derived plan summaries retain origin identity and show original plus exact
   own-step after-versions. Execution validates the original root approval and
   child binding. Target/choice/paths cannot change, no second UI approval is
   silently invented, and any required new decision ends this operation.

The step receipt is not a generic rollback/retry token. Git owns the mechanical
record and guards; Application alone sequences and decides to continue. Refused
drop after successful apply retains the prior exact applied facts and yields
AppliedStashRetained. Releasing the root ends this continuation authority.

Active preference remains Application/Persistence authority. Git never reads a
preferred target to select a worktree or persists activation. ListWorktrees is
the truth source for App's validated saved/current/deterministic fallback; partial
inventory cannot erase saved intent. Existing sessions retain their acquired cwd
when App activates another worktree.

## G7. Common coordination and exact revision safeguards

Serialize cooperating gh-tree mutations per LocalCommon RepositoryID across all
linked worktrees, with an in-process scheduler and cooperative cross-process
scope lock. This does not serialize independent clones or arbitrary external
editors/Git. Native Git ref/index interlocks remain required. Scope guard first;
acquire remaining native guards nonblockingly in the documented implementation
order and release only owned guards on every path. Never precreate a lock and
invoke ordinary nested Git expecting lock adoption. Existing locks return Busy/
Conflict; no age/PID-only unlink or forced lock deletion.

Before publication revalidate actual common/worktree administrative identity,
exact Head mode/ref/OID, index contents, configuration/conversion versions,
selected full target OID, file/parent identities and branch occupancy. Same-OID
HEAD switch to another branch is stale identity. Ref movement after selection or
fetch must equal ExpectedRevision; no local-tip/latest remote fallback. Pin
verified source/target/departure history using create-only owned recovery refs.
Private fetched roots use operation-nonce namespaces; never replace a shared
`refs/gh-tree/pr-N` that another operation can repurpose. No opportunistic GC,
reflog expiration or branch rewind cleans up those roots.

Retarget/deploy refuse primary/current worktrees according to the retained
secondary-target safety contract, unavailable/locked/prunable/unresolved paths,
unsafe dirty state, occupied branches and loss of departure history. User-initiated
branch creation/checkout is explicitly separate from secondary-target deploy;
it still requires exact native target/occupancy/history guards. Attach existing
branch at its exact current target without rewinding; advance only a proven
fast-forward. Configured legacy targets that cannot advance safely return refusal
with the detached/new-branch alternatives above Application, not force parity.

FetchRequest contains `Operation: OperationID`, `Binding: RemoteBinding`,
`RefScope: []FetchRefSpec`, `Prune: bool`, `ExpectedBinding: SourceVersion`.
FetchRefSpec contains explicit remote source ref, operation-owned destination
policy and optional authoritative expected Revision. Git validates native syntax
and exact scope; no guessed origin/refspec. Prefer native atomic ref updates,
while separately reporting downloaded objects. FetchResult contains
`Generation: Optional<FetchGeneration>`, `Refs: []RefFact`, `Freshness:
FetchFreshness` and the common mutation envelope. Failed/partial fetch or binding
replacement cannot advance an exhaustive successful generation. A completed
scoped fetch proves an observation at its interval, never permanent remote truth.

Push is pinned full source OID -> explicit remote branch, non-force with no `+`
or force-with-lease shortcut. Server rejection is authoritative. Optional upstream
setup is a subsequent checked native configuration step because source OID is
not a local branch argument to `push -u`. Revalidate expected local BranchID,
Head and config; preserve changed upstream rather than replacing it. Return
pushed-but-upstream-failed when appropriate. Remote timeout is uncertain even
when later observed tip equals the requested OID; that does not prove causality
or no intervening changes. Never retry automatically.

ReconcileRequest contains `Operation: OperationID` for this fresh query,
`OriginalOperation: OperationID`, `Repository: domain.RepositoryID`, exact
optional WorktreeID, original typed GitMutationKind/targets, prior EffectReport,
owned recovery locators and `Facets: []GitEffectFacet`. ReconcileResult returns
fresh local/remote-ref/status/stash observations per requested facet, unresolved
diagnostics and command-cleanup facts. It may inspect a journal and classify its
stage; it cannot replay a mutation, delete recovery, reset files or adopt foreign
locks. Remediation requiring a changed target/plan/decision is a new explicit
operation under the accepted confirmation protocol.

## G8. Protected native preparation and live publication

All supported worktree-file changes use this protocol. Native switch,
`read-tree -u`, restore, stash push/apply/pop, reset and clean must not rewrite or
remove actual worktree paths. Native Git remains authoritative for object/index
encoding, merge/conversion behavior and reference transactions.

1. During Execute, create a private **standalone** scratch repository of the same
   object format, with independent git directory, HEAD, refs/reflogs, index and
   object writes. Read-only alternates may read pinned real objects. A linked
   scratch worktree or alternate index with a real GIT_DIR does not isolate stash/
   ref effects. Copy live inputs; never hardlink live input objects into scratch.
2. Reproduce the exact confirmed Head, complete understood native index and
   relevant tracked/untracked/ignored state, including collisions. Bind effective
   attributes/ignore, info/global exclusions, filemode/symlink/case/EOL and supported
   filters/merge inputs. Disable scratch lifecycle hooks except the explicit
   native commit bridge in G8a, and disable fsmonitor/cache shortcuts
   and auto-maintenance. Do not copy live git/worktree/hook paths or extensions
   blindly. Re-stat changed inputs; unsupported/drifting conversions refuse.
3. Native scratch work computes immutable semantic path delta, resulting complete
   native index image/stages and required object roots. Only classified successful
   preparation or supported stash merge-conflict output may publish; generic
   failed scratch state grants no live write authority. Ordinary conflicts may
   publish guarded marker files plus unmerged index as AppliedWithConflicts.
4. Transfer bounded native packs into the real object database and pin every
   required source/target/departure/output object, including conflict-stage blobs
   which `write-tree` cannot root alone. Separately report preparatory native
   stash-store/new-branch effects. Native real ref hooks may change files; preserve
   them through the subsequent checks/capture rather than undoing them.
5. Prepare required native ref/HEAD transactions; acquire actual worktree index
   interlock exclusively and verify the full expected index version. Native
   competing add must fail/wait. Build all images in operation-owned alternate
   indexes; no nested command may expect to reuse the owned real lock.
6. Persist immutable manifest and journal before every irreversible entry
   transition. Manifest binds common/worktree/root/parent identity, exact Head/
   refs/target/index/config versions, expected-present/absent path facts, desired
   delta, object roots, transaction/hook plan, operation/plan/approval and recovery
   locations. Prefer capturing all originals before installing outputs.
7. For every affected expected-present entry, atomically capture the **actual
   object** by same-filesystem rename into an exclusive retained recovery name;
   validate the captured object, not the old pathname. A late edit/type/identity
   mismatch remains retained and refuses/partially stops publication. Expected-
   absent entries are never captured/deleted. Unchanged tracked paths, unrelated
   ignored/untracked entries and newly appearing files never join a cleanup set.
8. Freeze each desired output into a separate retained payload object that native
   scratch work will never touch again. Regular files install by no-replace
   hardlink from the retained object, or an independently proved equivalent that
   retains the installed object from first visibility. Supported symlinks use
   no-follow/exclusive publication with retained original link state. Never
   copy-overwrite/truncate a reappeared destination. Directory/file transitions
   affect named entries; remove empty directories only via rmdir, never recursion.
9. Capture the actual old index under its native lock, revalidate it and retain
   both old and installed index objects. Publish the complete understood native
   image with no-replace semantics to an absent live index name. Normalize only
   the alternate image using native read-only live-worktree inspection; scratch
   stat/cache fields never prove live cleanliness. Do not synthesize index binary
   fields or discard unknown semantic flags/extensions. Worktree-only Restore
   leaves the exact original index unchanged.
10. Release index guard; commit prepared native ref transaction. Preserve partial
    file/index publication after failed/unknown ref commit, with no reset/switch
    repair. Invoke applicable native real post-checkout/post-merge hooks once at
    their true commit point with correct context/arguments; preserve reference-
    transaction hooks. No cleanup/file/index replacement follows committed-ref
    or post-operation callbacks. Prepared-ref hook effects are revalidated before
    protected publication, as specified in G8a.
    Stash apply/restore do not invent post-checkout hooks.
11. Observe exact resulting Head/mode/ref/index/stages/files and occupancy. Record
    known changed/conflicted/partial/indeterminate facets, all retained originals
    and recovery locations. A post-hook edit is changed truth, not clean success.

Use validated directory-relative no-follow/native handle operations to bind
parents and reject substitutions. Recovery prefers the actual Git administrative
directory if same filesystem; otherwise exclusively create an operation directory
under a validated sibling `.gh-tree-recovery` on the affected filesystem. Bind
common repository, owner marker, parent identity and cryptographic nonce. Never
adopt another operation/owner directory. Cross-filesystem mounts or unavailable
safe retention/no-replace primitives refuse before capture; no copy-then-delete
fallback. Space limits refuse more publication rather than evict originals.

Keep captured originals and installed payload objects reachable, including late
writes through held-open descriptors after validation/publication. On failure,
restore a missing name only with no-replace while retaining an alias; never
overwrite a newly created live destination. Every recovery/retry replacement
captures the currently named object anew and installs a fresh payload exclusively.
An old digest or ownership marker cannot authorize truncation/deletion. There
is no automatic retention expiry or purge; cleanup requires an explicit conscious
safe disposition under product policy, and no new cleanup command is added here.

Worktree create keeps native registration, exact create-only destination/branch
semantics and the same protected payload protocol; native registration must not
perform an unguarded content checkout. Revalidate the acquired new root and
registration after interruption. Partially created destinations/administrative
records are explicit effects; never recursively remove uncertain user contents.

Stage/unstage creates an operation-owned native alternate index from exact planned
input and uses protected index publication. ExistingIndex commit and explicitly
sequenced StageAll commit use G8a's native engine, real hook/signing bridge and
guarded publication. No native command can adopt held guards, silently refresh
to another Head, overwrite a live index, or bypass this publisher to satisfy a
hook. A mechanism needing a contract/design change is raised before implementation.

No atomic visibility across all files/index/refs or global external-writer
serializability is claimed. The no-silent-loss guarantee comes from retaining
actual objects and no-replace publication, not quiescence assumptions or a final
precheck. Arbitrary converters/hooks are not sandboxed; outside effects/unknown
cleanup are reported honestly and never undone blindly.

## G8a. Commit specialization — proposed BC55-GIT-01 resolution

BC55-GIT-01's missing mechanical sequence is supplied below as a **proposed
specialization awaiting whole-set independent review**, not a self-approved
freeze. It preserves accepted exact-Head/index authority, native hooks/signing,
Application stage/commit sequencing and retained publication. The private engine
and bridge are Git internals; no new public port or generic command runner is
introduced.

1. Bind exact expected Head mode/name and parent P, complete original index
   bytes/semantic flags/object identity (or explicit expected absence), observed
   StageAll path/content set where requested, configuration/conversion/signing
   and hook-source versions. Established attached/detached commit uses exactly P.
   Initial commit uses the exact Unborn branch plus verified ref absence and no
   parent, never a Domain null OID. The files/full-index/no-sequencer profile and
   common/worktree/occupancy/path guards remain required.
2. Build the candidate index natively in operation-owned storage. ExistingIndex
   starts from the confirmed complete index; ObservedStageAll uses the original
   accepted path/content staging plan, never unconstrained later live `git add
   -A`. Application's protected StageAll substep may already have published the
   staged index and produced the exact G6 receipt. Its successful effect remains
   reported if Commit fails; no automatic unstage or inseparable success claim.
3. Establish a private standalone repository with independent refs/object writes,
   exact source Head/branch and copied candidate index; read alternates only from
   pinned real objects. Native ordinary `git commit -m <literal>` is the message/
   tree/parent/signing engine, with no amend, `-a`, pathspec or allow-empty escape.
   Map versioned supported effective configuration; disable private maintenance
   and unneeded rerere. It commits only private Head and performs no public
   worktree/index/ref publication. Unsupported context mapping refuses rather
   than dropping configuration or invoking native live commit.
4. Private core.hooksPath contains only adapter wrappers for pre-commit,
   prepare-commit-msg and commit-msg. Each immediately dispatches native
   `git hook run --ignore-missing <hook> -- <native args>` in the **actual
   worktree root**, restoring actual GIT_DIR, GIT_COMMON_DIR, GIT_WORK_TREE,
   original effective configuration/environment policy, native author variables
   and GIT_EDITOR=:. Every bridge receives the same absolute operation-owned
   GIT_INDEX_FILE. Convert only the native message filename to its absolute owned
   location; preserve other native arguments, including prepare source `message`
   for literal `-m`. Hook Git commands observe the real repository/Head and stage
   into the candidate index. Worktree edits remain real hook effects with no
   later worktree reset/checkout/cleanup. Scratch does not run the user's
   reference-transaction or post-commit hooks.
5. Preserve native hook order, operands, author environment, failure propagation
   and message cleanup. A hook using a hardcoded COMMIT_EDITMSG path instead of
   the passed filename, ignoring GIT_INDEX_FILE or depending on an unmapped
   private-context detail requires explicitly proved support or an unsupported
   profile before target publication. A private Git directory is not presumed
   equivalent to the real one. Arbitrary hook effects are not sandboxed and may
   leave changed/uncertain work even when the hook rejects the commit.
6. For configured signing, an adapter-owned transparent launcher runs the
   original configured signer in actual cwd/restored real Git environment with
   the candidate index. Preserve stdin/stdout/stderr, native format-specific
   arguments, exit semantics and final exact payload supplied by Git. Never
   synthesize signature headers, remove signing settings or substitute an
   unsigned commit. Resolve relative programs/key paths in their real context.
   OpenPGP/X.509, SSH literal/path keys and defaultKeyCommand need their explicit
   mappings/native profile checks; unreproducible profiles refuse before public
   publication. Signing helpers have the same bounded command/residual contract.
7. Native success returns exact candidate N, exact P/no-parent, final message and
   final candidate index as **distinct** outputs. Git rereads the index after
   pre-commit but builds N's tree before prepare-commit-msg/commit-msg. A later
   message hook's stage changes can legitimately be outside N; preserve them,
   never force final index tree to equal N's tree. Natively normalize only the
   alternate index against live files, then freeze its complete supported image.
   Transfer/pin N and every final index object root; N alone does not preserve
   later staged blobs. Retain/report candidate edits on hook/signing/native
   failure. Publishing a valid failed-commit staging result would require its own
   guarded classified stage substep, never unclassified failed-image publication.
8. Prepare the actual native reference transaction **through HEAD**:

   ```text
   git update-ref --stdin --create-reflog -m <native commit reflog reason>
   start
   update HEAD <N> <P-or-format-null>
   prepare
   ```

   Attached/unborn uses normal dereferencing; detached adds `--no-deref`.
   Prepared native transaction locks actual HEAD and its selected referent while
   checking expected P/null. After prepare, inspect exact actual raw Head
   mode/name under HEAD.lock and require the original expected Head. If another
   same-OID branch became Head before prepare, the native transaction may lock
   that branch but this identity check aborts before index/ref publication.
   Revalidate scope/config/occupancy. Do not combine explicit symref-verify HEAD
   with branch update: native implicit Head splitting rejects that composition.
9. While native ref guards remain owned, acquire actual index.lock exclusively
   and nonblockingly; compare complete original version/absence. Native
   reference-transaction prepared hooks have run in real context; their live
   index edits invalidate the original precondition. Construct/freeze images
   only in alternate storage; no nested command reuses the owned real lock.
   Apply G8's journal/capture/revalidate/no-replace publisher to the actual index,
   retaining old and installed objects, including late direct edits/recreated
   destinations. An unchanged logical index avoids unnecessary replacement;
   missing initial index uses expected-absence/no-replace, not a fabricated old
   object. Commit publishes no worktree-file delta.
10. Release only the owned index interlock, commit the prepared native ref
    transaction, then invoke real post-commit once after **known public commit
    success**, in real cwd/environment with the now-real index. Native public
    reference-transaction callbacks remain enabled. Recovery/pin transactions
    are separately reported and may have their own native callbacks. Prepared-
    hook rejection occurs before index publication. Committed-reference and
    post-commit edits occur afterward; no index/worktree replacement or cleanup
    follows them. A post-hook failure does not erase known committed Revision N.
11. Reconcile Head/ref/index/status and classify stage, hook/helper, object/
    recovery, index, ref and callback facets independently. Cancellation before
    index publication does not undo real hook effects. After index but before
    ref publication it can leave known staged/partial state with old Head; after
    known ref commit retain N even when cancellation is requested. Lost native
    result remains indeterminate unless actual evidence suffices. Keep candidate/
    journal/recovery, reap/join owned commands and do not reset, unstage, rerun
    post-commit blindly or replay the consumed commit.

Bounded source/model evidence is the [archived commit protocol](Feasibility/Commit/PROTOCOL.md),
its 25-case `results.json`, `probe_commit.py`, `bridge.py`, `signer_fixture.py`, tagged native
sources and `hashes.json` produced under #55. Author inspected those actual files;
all eight recorded hashes match, and the captured cases report PASS. The archive
retains exact source bytes (source filenames use `.txt`), original hashes and
its manifest; Master owns archival and independent whole-set disposition.
Native HEAD-first lock/identity
cases executed on Git 2.43.0.windows.1 and 2.48.1.windows.1; this commit mechanism
does not impose G9's newer attach-operation symref-update prerequisite. Ordinary,
initial and detached bridged commits execute in both SHA-1/SHA-256. The rejected
explicit symref composition and initial harness broken-pipe failure remain
evidence, not hidden passing production tests.

These are ordinary NTFS Python feasibility cases, not product acceptance or full
crash/process/permissions/path proof. The OpenPGP fixture proves native signer
invocation/payload/context and signature embedding; it is **not a valid
cryptographic signature** and proves neither real verification nor all signing
providers. V-GIT-08/09 still require full native Windows/Linux/macOS production
tests, real signing profiles, linked/absent-index paths, semantic flags/extensions,
configuration/hooks drift, Unicode paths, actual killed/crash/descendant states
and independently reviewed captured publication. Record source/freeze/evidence
SHAs and the reviewed BC55-GIT-01 disposition before FROZEN 1.0.0.

## G9. Native ref profile, occupancy and unsupported states

Affected attached/new-branch transitions require the tested native symref command
profile of Git 2.48.1 (the bounded evidence used 2.48.1.windows.1). Git 2.46 source
shows introduction, not a tested minimum. Capability detection uses supported
version/semantic profile, not guessed localized unknown-option prose. Older Git
receives UnsupportedGitCapability only for affected capabilities, with explicit
prerequisite; reads and supported detached/other operations remain available.
No installation/upgrade, manual HEAD writer or unguarded symbolic-ref fallback.

Create a new branch at exact T using native create-only update first; retain its
separate preparatory effect if later attachment fails. Then native
`update-ref --stdin --no-deref --create-reflog` starts and prepares
`symref-update HEAD <new> ref <old>` (or supported `oid <B>` for detached old Head)
and `verify <new> <T>`. Inspect old Head mode/ref/OID under lock. For files backend,
hold the departure branch's exclusive **verify-only** native ref lock and compare
B; adding its verify to the same HEAD transaction conflicts with native implicit
HEAD-log behavior. The guard never writes ref bytes. Protected file/index
publication precedes native transaction commit and exact symbolic Head/T/reflog
verification. Already-attached fast-forward uses native expected-old branch B->T
with associated Head lock, without duplicate explicit Head update. Detached
updates use native no-deref CAS and required departure-branch guard.

Revalidate native inventory before publication and after commit. Ref locks are
not occupancy locks; external checkout can race inventory. Occupied or incomplete
occupancy refuses before target change where observed; later change reports
partial/changed/uncertain state and retained recovery. No force flag, foreign-
worktree write or cleanup reconciles it. Existing branches are never rewound.

Initial supported mutations require files refs, known SHA-1/SHA-256, ordinary
full understood index, regular files/supported symlinks and reproducible native
conversion inputs. Unproved sparse/split index, skip-worktree, assume-unchanged,
intent-to-add, gitlinks/submodules, unknown required extensions, active merge/
rebase/cherry-pick/sequencer, special files, redirected/unproved parent paths or
unsupported retention primitives return UnsupportedMutationState before unsafe
publication. Ordinary unmerged stages produced by classified scratch apply are
supported output, not permission to overwrite a preexisting conflict. Read
capabilities remain available. Cwd-dependent or unproved external converters
refuse explicitly; ordinary supported attributes/EOL semantics remain required.
No silent feature removal or fallback to live checkout is permitted.

## G10. Stable stash create/apply/drop and journal recovery

StashID is LocalCommon repository + exact stash OID. Verify native stash-like
object/parents and pin it. Apply reads the pinned full OID with native scratch
`stash apply --index`, only from a confirmed clean/nonconflicted target, preserving
staged/worktree/untracked structure. Position and managed-origin metadata are
display only. A new destructive operation requires the exact observed live
occurrence; missing log entry is StashMissing even when recovery still retains
the object. Repeated identical OIDs require uniquely matching occurrence evidence;
ambiguous/mismatched selection refuses without deleting all duplicates.

Occurrence SourceVersion binds the observed record's exact new OID and retained
author/message/occurrence evidence independently of its numeric display position.
An unrelated prepend that shifts positions may remain compatible when the same
unique record is proved under the ref lock; a whole-list generation change alone
does not choose another occurrence. Changed/duplicated indistinguishable record
evidence refuses safely. ReadStashPatch needs the exact object, not live occurrence.

Create uses native scratch stash push, including selected `-u`, preserving its
exact W/index/untracked parents, then native pack transfer and real `stash store`
of that known OID. Operation-unique metadata plus actual log/commit observation
proves the created occurrence; never resolve a later refs/stash tip and assume
it belongs to this operation. Store precedes protected cleanup. A real ref hook
that writes tracked/untracked data causes retained mismatch/refusal rather than
loss: StashCreatedCleanupRefused returns the exact stored stash and preserved
newer work. Never re-store/apply automatically after uncertainty.

Exact drop is the deliberately narrow files-backend writer below; no native
positional drop/pop can express expected occurrence atomically. It is private
to `refs/stash` and its reflog, not a general hand-written ref backend.

1. Verify known object/ref format, non-symbolic refs/stash, bounded valid log,
   native common directory and tested no-redirection filesystem profile. Unknown/
   reftable/symbolic/malformed state refuses exact deletion.
2. Under common guard, exclusively acquire actual `refs/stash.lock`, not merely
   `logs/refs/stash.lock`. After this native writer interlock, read full current
   log/top and identify the unique expected OID/occurrence. New prior entries
   remain; missing/ambiguous/mismatched selection aborts before log publication.
3. Create-only native recovery refs preserve selected/surviving objects. Flush
   owned journal/original log bytes/old loose-resolved top/storage identities/
   intended new bytes/top/lock owner/stage before changing live log. Report
   preparatory objects/refs independently if the operation subsequently refuses.
4. Remove only the selected record. Preserve survivor order/new OIDs/author/
   message bytes exactly; rewrite only old-OID chaining as native `--rewrite`
   requires. First old OID is native protocol null, not Domain Revision.
5. Flush owned complete log-lock image and nonempty new-top ref-lock image.
   Publish log then ref using tested same-directory replacement/native ordering.
   A loose new top may shadow packed old ref; never hand-edit packed-refs. Journal
   stages and verify surviving vector/ref while ownership remains where possible.
6. For an empty survivor log, retain old ref temporarily, publish empty log,
   release manual ref lock, then prepare native expected-old delete transaction.
   While native locks are held, recheck the current log is still empty; commit
   only then. Abort on A->B->A/new-entry reinsertion; native Git handles packed
   deletion. Failed empty-ref cleanup is a distinct effect, never authority to
   delete a newly created stack.
7. Retain recovery/journal until an explicit safe disposition. Any owned ref
   cleanup uses native expected-OID deletion, never generic reflog expiry or GC.

The log/ref pair is not atomic. Before publication old state survives; new-log/
old-ref may leave a journal-owned lock. Recovery may finish the exact known
intermediate state only with verified original process identity/liveness and
matching journal/log/ref/lock identities/hashes. New-log/new-ref can be classified
completed. Missing/changed lock, uncertain owner, manual repair or unmatched
bytes means RecoveryRequired; preserve originals, refuse further destructive
stash operations and expose locations. Never age-delete locks, replay old bytes
over unknown state, or claim unflushed power-loss/disk-corruption guarantees.
Read-only Reconcile reports this state; destructive recovery is not implicit.

## G11. Cancellation, immutable results and command resources

All input/output slices, maps, path buffers, nested manifests, diagnostics and
post-facts are copied. Adapter memory never changes a previously returned snapshot
or plan summary. Read commands disable optional index writes. No port callback,
UI event subscription or progress handler can mutate State. Application alone
publishes one terminal event with its original operation/correlation.

Git's adapter-private transport separates bounded machine stdout (initial 16MiB)
and stderr (256KiB); narrower patch/data bounds apply. Stream native object packs
with explicit byte bounds. Read budget is 30s; mutating/network/hook budget 120s;
forced pipe-drain/join after cancel has a separate 2s bound, configurable at
construction. One owner reaps each root and joins pipe readers. Record started,
reaped, transport-cleanup-known, truncation/deadline/limit facts in
CommandTransportOutcome. Localized stderr/exit code alone cannot prove stale,
missing, untouched or successfully mutated semantic state.

Canceled-before-acquisition returns NotStarted; after any possible effect retain
bounded adapter-owned reap/reconciliation independent of the canceled caller
wait. A killed root, pipe closure, context timeout or navigation change is not
rollback or descendant cleanup proof. Normal root-only CommandContext semantics
cannot borrow Runtime's private supervisor. If stronger short-command descendant
ownership is needed, it requires implementation/review in Git or accepted
dependency change. Unknown helpers/hooks/transport cleanup remain residuals.

Application retains notices and blocks unsafe automatic retry in affected scope,
even after modal disposal, and includes residuals in aggregate shutdown. No
unbounded side queue, discarded known effects, or second operation terminal after
late reconciliation. Cancel/Shutdown can reach admitted operations when normal
admission is full; neither releases native ownership before the resource barrier.

Shared ErrorCode plus structured Git reason identifies Invalid, NotFound, Busy,
Canceled, StaleObservation/ConfirmationStale, Conflict, Permission, Unsupported,
IOFailure, ProcessFailure, CleanupIncomplete or Indeterminate. Git-specific
reasons include DirtyWorktree, ConflictPresent, UntrackedCollision, BranchOccupied,
PrimaryWorktreeRefusal, CurrentWorktreeRefusal, DetachedHead, UnbornHead, NoUpstream,
UpstreamGone, UpstreamUnresolved, NonFastForward, ExpectedRevisionMismatch,
AmbiguousStashOccurrence, StashMissing, UnsupportedGitCapability,
UnsupportedMutationState and RecoveryRequired. Structured detail includes exact
scope/expected/observed identity and recovery where safe, excluding credentials.

## G12. Forbidden calls, verification and change control

Forbidden: concrete GitHub/Runtime calls, active preference selection/write,
Application Client/UI callbacks, cross-source joins, provider execution/session
ownership, generic mutation dispatch, positional stash authority, force rewind/
force push, symbolic-ref/manual HEAD fallback, recursive user-path cleanup,
live checkout/reset/clean writes, premature lock release, forged freshness, stale
approval refresh, automatic uncertain retry and untracked/dirty byte loss.

Required checks are [Verification--001](../Design/CR-%2321/Verification--001.md)
V-GIT-01..10, V-DOM-01..03, V-APP-01..06, V-GH-01..03 and V-COMP-01/03/04,
plus V-E2E-01..08, worktree scope in 09..11 and cancel/cleanup in 12.
Use real temporary repositories, bare remotes, linked worktrees and independent
clones in SHA-1/SHA-256. Prove every exact read/Prepare method, closed plan kind,
wrong issuer/version/operation, release/replay/admission, no locks during user
delay and result preservation with nonnil error.

Mandatory safety proof includes post-confirmation same-size/mtime edits, hooks,
held-open writers, recreated destination, captured/index/ref/registration crash
stages, concurrent native stash/store/drop, packed/last-entry ABA, survivor byte
preservation, source/config/occupancy drift, partial fetch, exact merge-base and
StashPatch, stage/commit hooks/signing failure, pinned push/upstream effects and
bounded canceled transport. Native Windows/Linux/macOS path/index/ref/hook/
retention checks and all twelve selected-platform builds remain required.
The eighteen feasibility cases are mechanism evidence only; they do not satisfy
the future implementation/platform/fault gates.

Read this contract with shared types and Application/GitHub, State/Application,
Persistence and Discovery/Runtime worktree-scope contracts. The whole seven-BC
set receives a separate fresh independent review before freeze. No baseline
product finding is closed by this draft.

Change history: 2026-09-06, DRAFT 1.0.0 authored under #55. Master records
review/correction, source/freeze/merge SHA and configured exact-HEAD CI before
FROZEN. A later insufficient signature/protocol or incompatible capability is
BC-CHANGE: stop affected work, analyze design impact, independently review,
refreeze and reverify every affected layer/Slice. No worker workaround.
