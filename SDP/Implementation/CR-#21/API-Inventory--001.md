# M2 API and ports completeness inventory

State: read-only preparation review; no implementation acceptance or contract change.
Reviewer: m2_api_inventory_review at75a214200d1ab192f491e11c7165e1d5def4712d.
Authority: full #21, accepted design and seven FROZEN1.0.0 BCs/BoundaryTypes.
No material signature/type-ownership contradiction was found. #59 is the bounded
implementation authority; actual frozen documents remain normative over this index.

## Closed public inventory

Client has exactly Submit, Confirm, Cancel, Next and Shutdown. Preserve validated
Request/Receipt/Event/Correlation/Choice, command/query payloads, typed terminals
and aggregate shutdown. The28 command/result semantics are ActivateWorktree,
SaveNavigation, CreateWorktree, RetargetWorktree, Deploy, CreateBranch, Fetch, Pull,
Push, StagePaths, UnstagePaths, StageAll, Commit, StageAllAndCommit, RestoreTracked,
StashCreate, StashApply, StashPop, StashDrop, CreatePullRequest, SaveLaunch,
StartLaunch, OpenTerminal, WriteInput, ResizeSession, InterruptSession, StopSession
and RestartSession. No workflow-specific partial outcome may disappear in wrapping.

Thirteen query families: Navigator, BranchContext, Commits, Graph, Diff,
PullRequestDiff, WorktreeStatus, Stashes, StashPatch, LaunchPoints, Sessions,
SessionOutput and Preferences, with complete canonical projections/relations.
Events: Accepted, Progress, ConfirmationRequested, OperationTerminal,
SessionChanged and ProjectionInvalidated. Terminal disposition is Succeeded,
Failed, Canceled or Superseded, independently of effects/cancellation/recovery.

## Seven interfaces,48 exact methods

| Port | Methods |
|---|---|
| GitFacts (11) | ResolveLocal, ListWorktrees, ObserveStatus, ResolveExact, ListRefs, ListStashes, ReadStashPatch, MergeBase, ReadCommits, ReadGraph, ReadDiff |
| GitMutations (12) | PrepareCreate, PrepareRetarget, PrepareStage, PrepareCommit, PrepareRestore, PrepareStash, PrepareBranch, PreparePush, ExecutePrepared, ReleasePlan, Fetch, Reconcile |
| RemoteFacts (4) | ResolveRepository, ListBranches, ListPullRequests, ObservePullRequest |
| RemoteMutations (1) | CreatePullRequest |
| LaunchDiscovery (2) | Discover, Resolve |
| Storage (6) | LoadUserConfig, LoadPreferences, LoadRunConfig, CommitUserConfig, CommitPreferences, CommitRunConfig |
| Sessions (12) | Start, Snapshot, List, ReadOutput, Write, Resize, Interrupt, Stop, Restart, NextEvent, AckEvents, Shutdown |

Exact signatures override generic naming: Git Prepare returns specific ports plan
plus API result/error; Execute returns ExecutedGitMutation. ReleasePlan/ACK have
no context. Storage loads/commit requests are ports wrappers. Runtime Snapshot/
Interrupt take Domain SessionID; Resize/Interrupt share SessionControlResult;
Runtime Shutdown has its aggregate result without a separate error.

## Required complete families and ownership

- BoundaryTypes B1-B5: options, all distinct IDs/sequences/opaque versions, source/
  interval/completeness/pages, WorktreeScope/DirectoryIdentity/CwdObservation,
  RemoteBinding, Invocation/execution/shell/environment/geometry, diagnostics and
  transport. Fifteen error codes, nine effect facets, five states, shared recovery
  IDs/records plus lossless family detail. API owns SourceVersion/StorageVersion/
  RuntimeEventSequence; ports aliases never permit API->ports. Issuance/native
  revalidation/registries/allocators remain outside these pure values.
- Git G2-G7: every typed fact/read/prepare/fetch/reconcile request/result, exact
  object/ref/stash/merge-base/patch/comparison/freshness/profile/occupancy records,
  expected state/plan summaries/completed steps and all closed mutation outcomes.
  Eight specific plan types, PreparedGitPlan, GitPrepareContext, receipt, execution
  approval/result wrappers remain ports-owned. Keep origin/predecessor and executable
  versus sequence-root distinctions for stash/deploy, pop and staged commit.
- GitHub GH2-GH5: full qualified repository/branch/PR endpoints, observations,
  filters/expectations/creation evidence; explicit independent unavailable fields.
  Six create outcomes retain requested and observed base/head facts separately.
- Discovery: observations/definitions/project sources, provider kind, member/saved/
  resolved values and Discovered/OrderedMake/Saved selections. Preserve exact
  ordered members and alias/source/storage binding; Application resolves default.
- Storage: complete UserConfig/Preferences/RunConfig documents, every specified
  nested known/unknown/legacy value, StoredField/JSONMembers/OpaqueJSON and typed
  load/commit/recovery state. Among specified known fields only StripPrefixes
  permits Null; seven load states and four commit outcomes remain distinct.
- Runtime: all requests/results, session snapshots/capabilities/exit/cleanup,
  output/chunk/gap/ranges, controls and aggregate shutdown/residuals. Five phases,
  three event kinds and all cleanup stages through EventTransfer. Failed admitted
  Start retains identity; Restart retains old outcome separately from replacement.
- Canonical projections: BranchRelationship, PRRelation, WorktreeRelation,
  ActivationCandidate, ProjectionSources and ActiveContext; Active deploy carries
  expected ContextVersion. Exact PR base/head/merge-base and stash parents remain
  typed evidence. State/viewmodel records/options/modal IDs do not enter API/ports.

## Validation and proof boundaries

Optional alone does not deep-copy nested data. All admission/return/access paths,
documents, summaries, diagnostics, recovery and closed variants need explicit
copy/validation behavior. Preserve valid result plus nonnil error, LocalConfiguration
independently from refs/Storage, seven allowed Git reconcile facets and resolvable
recovery records. Storage normalization retains shared Record plus typed detail;
document versions stay StorageVersion and artifact identity stays SourceVersion.

API may use json.Valid and permitted pure predicates for bounded immutable
OpaqueJSON. That predicate alone does not prove UTF-8, duplicate-key, depth,
known-name collision or whole-document validity. Appropriate value/document
validation and Persistence's strict codec must enforce those requirements. JSON
schema naming/decoding/encoding stays Persistence-owned; no API codec or checker
relaxation is justified. Constructor/helper names and file layout inside the
assigned leaves remain ordinary implementation details.

Compile narrow external fakes for every exact interface and test invalid/foreign/
zero/typed-nil/tag/copy/request-result cases. M1 checks all12 target selections.
Leaf tests do not prove coordinator exactly-once/workflows or native adapter
behavior. Source citations are the complete seven BCs and BoundaryTypes sections
identified above, plus accepted API--001; do not implement from this checklist alone.
