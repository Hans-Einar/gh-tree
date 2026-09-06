# BC--TUIState--Application

State: FROZEN
Version: 1.0.0

Freeze: [BCFreeze--001](BCFreeze--001.md); effective after reviewed PR #56 merge.
Authority: #55 under #21; accepted design #52/PR54, merge4a42222f7bfedc1d80693effbb25a1a82fcff65e.
Applies to: all UI SLC-01..12 and SLC-13 startup/migration observations.
Supersedes: none. No implementation authority before whole-set freeze.

## Responsibilities and dependencies

Application owns product use cases, operation/confirmation identity, lifetimes,
source association/reconciliation, authoritative active context, durable intent,
adapter orchestration, read models and normalized events. State owns ephemeral
user mode/focus/selection/modal/forms/filter/scroll, request generations, intent
tokens and copied projections. State never chooses which adapter sequence makes
a deploy, stash, commit, launch, activation or shutdown happen.

`tuistate -> application/api + domain + tuistate/viewmodel`; no ports or concrete
backend import. Application API imports Domain and standard value/context types,
never State/View/ports. The restricted Composition host executes State's declarative
effects through Client and feeds events back; State does not call Client during
its reducer. [BoundaryTypes--001](BoundaryTypes--001.md) is normative shared vocabulary;
[accepted API](../Design/CR-%2321/API--001.md) fixes the closed capability inventory.

## Public Go surface and envelopes

```go
// package api
type Client interface {
    Submit(context.Context, Request) (Receipt, error)
    Confirm(context.Context, ConfirmationID, Choice) error
    Cancel(OperationID) error
    Next(context.Context) (Event, error)
    Shutdown(context.Context) ShutdownResult
}
type Receipt struct { OperationID OperationID; AcceptedSequence Sequence }
type Event struct { Sequence Sequence; Payload EventPayload }
```

Request is a validated private envelope containing one sealed Command or Query
payload and Correlation. Constructors reject both/neither, typed nil, invalid
IDs/tag combinations and query-only correlation on a mutation. Submit validates
cross-field scope/expected versions before accepting. Receipt and Accepted event
refer to the same allocated operation; receiving both cannot start two operations.
Submit context bounds admission only. After acceptance, coordinator lifetime and
explicit Cancel/supersession/shutdown govern execution, not a discarded UI context.

Choice is a closed semantic choice value such as Proceed, Cancel, StashThenDeploy
or the exact workflow-specific variant allowed in the supplied confirmation.
Confirm accepts only that pending token's listed choices. No free-form shell,
plan pointer, native handle or unchecked backend request enters this API.

## Closed commands and terminal payloads

Each row is a distinct sealed command type (name suffixed Command) and result type
(name suffixed Result). Types/fields below are API-owned. Common IDs/versions use
the annex, exact selections use Domain targets, and returned family facts use the
corresponding adapter BC. No command is a concrete-adapter facade method.

| Command | Required input | Required result |
|---|---|---|
| ActivateWorktree | WorktreeID, expected preference StorageVersion | ActiveContext and precise StorageCommitResult |
| SaveNavigation | explicit repository scope, namespace/folder intent, expected StorageVersion | committed effective intent/version and commit diagnostics |
| CreateWorktree | exact target, vacant literal target path, closed Detached/CreateNewBranch policy | optional new WorktreeID/Scope/Head and full creation effects/recovery |
| RetargetWorktree | WorktreeID, exact target, explicit attach/detach policy | exact resulting Head/Scope plus effects and diagnostics |
| Deploy | exact PR/branch/commit target; Active, chosen WorktreeID or named configured target choice | resolved bound target, resulting Head and optional created stash/recovery, even if later step failed |
| CreateBranch | WorktreeID, new branch name, exact starting Revision, checkout choice | created BranchID and optional updated Head, separate effects |
| Fetch | explicit local/remote binding/ref scope | fetch generation and exact observed refs/completeness |
| Pull | WorktreeID, expected Head and observed resolved upstream | exact fetched endpoint and guarded fast-forward/refusal/effect facts |
| Push | WorktreeID, expected local Revision, explicit remote BranchID, set-upstream choice | remote ref effect separately from upstream configuration effect |
| StagePaths / UnstagePaths / StageAll | WorktreeID, explicit path list or All tag, expected index/status versions | resulting index/status and independently known effects |
| Commit / StageAllAndCommit | WorktreeID, expected Head/index/status, literal message, explicit staging policy | exact optional new Revision, candidate/staged/ref/hook effects and recovery |
| RestoreTracked | WorktreeID, exact path set, expected index/worktree versions | restored-from-index facts and retained originals/recovery |
| StashCreate | WorktreeID, expected dirty observation, message, include-untracked choice | exact optional created StashID and cleanup outcome |
| StashApply / StashPop / StashDrop | exact StashID, target WorktreeID where needed, expected observation | exact affected stash/index/worktree effects; AppliedStashRetained and StashCreatedCleanupRefused remain distinct |
| CreatePullRequest | explicit base/head repository+branch+expected published revision, title/body/draft and maintainer-modification policy | typed created PR identity/URL or remote-effect uncertainty |
| SaveLaunch | WorktreeID, exact discovery/saved selection/version, alias, default choice and version-bound replacement approval | complete committed saved definition/default and storage outcome |
| StartLaunch | WorktreeID, exact discovered/saved/default selection and geometry | established new SessionID or retained failed-start session/effects |
| OpenTerminal | WorktreeID, explicit Auto/Configured shell policy and geometry | established interactive SessionID/capabilities or precise failure |
| WriteInput | SessionID and copied bytes | queue-accepted byte count, never child-consumption or retry permission |
| ResizeSession / InterruptSession | SessionID and positive geometry / advertised interrupt intent | delivered-control fact or typed unavailable/unsupported outcome |
| StopSession / RestartSession | SessionID, optional restart geometry | complete old cleanup fact and only then optional new SessionID |

All commands return the common terminal effects/diagnostics/recovery envelope.
No nil/error branch loses a known changed result. A result may be failed/partial
while containing a created stash/branch/commit or committed preference. Distinct
concurrent restart operations cannot create duplicate replacements from one old
transition; later deliberate restart targets the returned replacement identity.

## Closed queries and canonical projections

Queries are sealed types suffixed Query, with their accepted names and matching
projection payloads. Each carries current source selection and QuerySlot/generation.

| Query | Inputs and result ownership |
|---|---|
| Navigator | explicit remote namespace/repository and optional associated local scope; returns PR/branch hierarchy facts, canonical branch relations, worktrees, capabilities, active context and independent source completeness/errors |
| BranchContext / Commits | exact selected target and bounded page; verified endpoint, full commit facts/messages, scoped branch/PR direction and pagination |
| Graph | local scope, exact roots/filter and bounded page; semantic DAG/refs and App-owned PR/worktree/branch annotations, no display lane prefixes in messages |
| Diff | closed CommitParent/CommitPair/IndexToWorktree/HeadToIndex comparison and literal path filter; exact endpoints/source, bounded patch/files/truncation |
| PullRequestDiff | exact PR head target plus exact base or explicit current-base observation choice; exact resolved base/head/merge-base and three-dot-equivalent patch/files |
| WorktreeStatus | WorktreeID; typed Head/status/upstream/availability/occupancy and source versions |
| Stashes / StashPatch | local common scope list, or exact StashID+parent/view/path selection; immutable stash identity/metadata/patch/files independent of reflog position |
| LaunchPoints | WorktreeID; normalized discovered definitions, exact saved aliases/default and source/storage versions, partial diagnostics |
| Sessions / SessionOutput | optional WorktreeID, or SessionID/offset/bound; copied session facts or bounded output/gap observation |
| Preferences | explicit resolved scope; effective settings, retained migration diagnostics/source versions and authoritative ActiveContext |

Projection records carry typed semantic rows/edges/endpoints and source-specific
ObservationID/SourceVersion/Completeness. Local, cached upstream and remote facts
are distinct; App performs scope/PR/fork/worktree joins. Optional counts remain
unknown when unproved. One failed/capped source cannot erase independently valid
others or prove a selected namespace/branch/PR absent. Display labels never decide
relationships. Read models contain no service handle, formatter callback or native DTO.

Canonical relation records are API-owned:

| Record | Required semantic fields |
|---|---|
| BranchRelationship | selected scoped BranchID/exact expected Revision, optional matching local branch fact, explicit UpstreamFact, scoped remote endpoint facts, PRRelation list, WorktreeRelation list, observation/version/completeness bundle and diagnostics |
| PRRelation | PRID, Head/Base role, exact scoped branch/optional revision endpoint and Complete/Unknown relationship evidence. Multiple/fork relations remain a list; same spelling never joins different repository scopes. |
| WorktreeRelation | WorktreeID, optional Head, exact identity source, relation kind ExactSelectedRevision/SameScopedBranchDifferentRevision/Unrelated/Unknown, availability and current/primary/active flags from their actual authorities |
| ActivationCandidate | App-selected WorktreeID, explicit relation reason/revision difference, availability and required context/source versions. UI never searches plain branch names to manufacture this candidate. |
| ProjectionSources | independently valid Git/remote/storage observation IDs, versions, completeness and diagnostics. A conclusive negative relation requires complete evidence from every required source. |

App supplies applicable action destinations with these projections. A Deploy
destination of Active includes the caller's expected ContextVersion; mismatch
refuses rather than quietly resolving a newer active target. Explicit worktree and
configured-target destinations retain their own exact binding/preconditions.
Activation uses the supplied available WorktreeID and fresh inventory validation;
it is not checkout and cannot imply that a differing revision has been deployed.

ActiveContext contains optional active WorktreeID, available scope/status,
preference StorageVersion/commit diagnostics and a monotonic ContextVersion owned
by App. Startup validates saved intent, otherwise Current then deterministic
inventory fallback without erasing unavailable old intent. Startup/initialization
is Application lifecycle work; Composition only invokes/wires it. Queries capture
the context version they actually read. State never replaces a newer context
with an older query snapshot, even if the query slot itself is current.

## Events, admission and terminal invariants

EventPayload is sealed Accepted, Progress, ConfirmationRequested, OperationTerminal,
SessionChanged or ProjectionInvalidated. Every operation event contains OperationID
and the original copied Correlation. SessionChanged contains SessionID and the
source SessionSequence independently of current tab; it never completes Start again.
ProjectionInvalidated is a coalescible source/context hint prompting an appropriate
query, not an authoritative deletion or focus instruction.

OperationTerminal contains closed Succeeded/Failed/Canceled/Superseded disposition,
optional typed result, structured diagnostics, EffectReport, recovery records and
CancellationRequested. It is emitted once for every accepted operation, including
panic, shutdown, supersession and lost refresh. Disposition does not replace effect
facts: cancellation may coexist with applied/partial/indeterminate changes. A fully
completed mutation may report success with a later cancellation request recorded.

Admission reserves Accepted, at most one confirmation and Terminal; Start/Open/
Restart additionally reserve a potential new-session cleanup slot. Initial limits:
64 live operations and shared256 unconsumed critical slots, including outstanding
session reservations. Refuse Busy before acceptance when either cannot reserve.
Critical outcomes never drop; progress/invalidation/nonterminal hints may coalesce.
Terminal delivery releases its operation reservation. Cleaned session delivery or
explicit shutdown drainage releases the session reservation. API A6/Runtime BC
govern transfer-before-ACK, dedup and bounded pending acknowledgments.

Next has one host consumer, returning strictly increasing App Sequence. Context
cancellation before successful delivery leaves an event pending. A returned event
remains valid if cancellation races afterward; host applies/drains it before Next.
Keep the latest256 delivered operation terminals as bounded diagnostic history;
no unbounded overflow queue and no eviction hiding unresolved native resources.

## Confirmation and workflow state authority

App prepares typed preconditions and issues one single-use ConfirmationID binding
operation, exact target/worktree/versions, immutable safe summary and allowed
choices. No native lock/handle/process remains held during user delay. User approval
belongs only to that original operation; Confirm atomically consumes it once.
Wrong/stale/consumed token or disallowed choice refuses. Cancel ends pending approval.
No completion from A can dismiss/open a later modal B.

Reacquire guards and validate before each irreversible continuation. Changed
preconditions produce ConfirmationStale and require new intent, never fresh state
silently substituted under old approval. App orchestrates approved compound steps;
Git's typed origin/receipt mechanism may advance only verified authorized own-step
versions. An applied stash with failed cleanup/drop stays visible; no automatic
apply/commit/push replay. StageAllAndCommit preserves a known staged effect when
the later commit fails. No workflow lives in a reducer or host callback.

Activate validates current inventory, commits expected-version preference, then
publishes ActiveContext. NotCommitted leaves the old context. Committed durability
uncertainty publishes actual committed intent with its diagnostic. Indeterminate
storage is reconciled before dependent activation. ContextVersion prevents stale
query/mutation delivery from reversing newer authoritative state. Display selection
and current process cwd are different from active application context.

## Cancellation, ordering and State correlation

Queries supersede only older queries in the same explicit slot and still receive
terminals. Mutations serialize per local common repository across whole authorized
compound execution; active changes serialize per App session and Runtime lifecycle
per SessionID. Navigation does not kill, forget or imply rollback of mutation.
Waiting confirmation owns only coordinator state, not native resources.

State tracks global IntentToken, current generation/source per query slot, pending
operation/confirmation ownership and last SessionSequence/ContextVersion. Query
data is accepted only for matching slot/generation/source; semantic session/context
facts additionally use their own versions. Conditional focus/selection/modal
follow-up requires the original intent still match. A valid operation result may
update facts while its now-stale navigation side effect is suppressed. Mutations'
effects/errors remain discoverable even if their initiating modal is gone.

One focus path and one modal variant replace all release wrapper booleans. Console
selection belongs only to State. A vanished selected semantic item gets deterministic
fallback only when complete evidence proves disappearance; partial sources retain
pending intent. Both graph entry modes use the same accumulated reducer/projection.

## Shutdown and forbidden behavior

Cancel and Shutdown bypass normal admission. Shutdown closes admission, cancels
queries/pending approvals, retains native mutation reconciliation, drains outcomes
even after host detachment, joins use cases/subscriptions and Runtime barriers,
then closes the stream. Caller deadline only stops waiting; cleanup remains owned.
ShutdownResult includes Complete, operation/session residuals and aggregate errors.
Composition combines its primary error with cleanup result and exits nonzero for
residuals. No UI exit callback silently discards that result.

Forbidden: concrete backend/ports calls from State, UI-managed workflow sequences,
global busy booleans as operation identity, provider-string resource routing,
optimistic durable active mutation, silent effect loss on errors, callback-shaped
commands, uncorrelated modal/focus follow and unbounded critical event storage.

## Verification and change history

V-APP-01..06, V-STATE-01..04, relevant Runtime/Git/Storage checks and V-E2E-01..12
prove this boundary. Include async Start then Alt/Tab, restart A while B selected,
worktree switch during stale status/stash/launch query, stale diff/branch/page/modal,
active commit failure/uncertainty and full queues with natural exits/host detach.
Test denied/consumed confirmations and all accepted operation terminal paths.

1.0.0 DRAFT under #55. Freeze requires whole-set independent review. Later
incompatible changes require BC-CHANGE, reviewed impact/refreeze and affected tests.

Freeze history: 2026-09-06, whole set independently REVIEWED at
7685494e45c0ef44fbccf9b49a589a90a78026d0, then marked FROZEN 1.0.0 by Master.
BCFreeze--001 governs effective authority after final metadata review/CI/PR56 merge.
