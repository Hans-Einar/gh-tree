# BoundaryTypes--001 — shared v0.4 contract vocabulary

State: DRAFT
Version: 1.0.0
Authority: #55 under #21; accepted design #52, merged4a42222f7bfedc1d80693effbb25a1a82fcff65e.
This annex is part of the seven-BC set, not an additional layer or adapter.
No implementation authority exists before the whole-set freeze gate.

## B1. Ownership, validity and copying

Domain owns only the pure validated identities/exact targets in accepted
[API A1](../Design/CR-%2321/API--001.md#a1). API owns semantic requests, facts,
projections and outcomes. Ports owns interfaces, preparation/load wrappers and
private issued plan/receipt handles. The State-owned viewmodel leaf has independent
pure presentation records; it imports Domain and pure standard values, never API,
ports or a backend. Do not create a generic shared types/process/filesystem package.

All IDs and versions are distinct Go types, not aliases of an interchangeable
integer/string at call sites. Domain values have private representation and
validating constructors; zero/nil is invalid unless wrapped as explicit absence.
ExactTarget is the closed CommitTarget/BranchTarget/PullRequestTarget union.
Head is the closed Attached/Detached/Unborn union; unknown is an absent observation
with a diagnostic, not a fourth Head or a zero Revision.

Every boundary copies slices, maps, byte buffers, optional nested values and
interface payloads on admission and return. No callback, lazy iterator, reader,
backend object, OS handle, goroutine or mutable internal record is a DTO. Closed
alternatives use sealed marker interfaces or validated private tagged values;
consumers cannot add arbitrary variants. A typed nil is invalid, not absence.

```go
// package api; construction validates the tag/value invariant.
type Optional[T any] struct { present bool; value T }
// None[T](), Some[T](value), Present() bool, Value() (T, bool)
```

The containing family clones mutable T content before publication/access; a generic
wrapper alone does not make arbitrary T immutable. Getters do not expose owned
buffers. The viewmodel leaf may define its own equivalent optional value rather
than importing API. Ports may alias public opaque versions/cursors as follows;
their issuance/revalidation mechanics remain private to the responsible adapter:

```go
type SourceVersion = api.SourceVersion
type StorageVersion = api.StorageVersion
type RuntimeEventCursor = api.RuntimeEventSequence
```

This resolves public-version transport without API->ports imports. A version is
evidence to revalidate, never a caller-created authority to mutate native state.

## B2. IDs, observations and pages

| API value | Exact semantic representation and rules |
|---|---|
| OperationID | Positive uint64 allocated monotonically by Application for its lifetime; refusal before wrap. Runtime/Git may retain attribution/dedup but never allocate or complete an App operation. |
| Sequence | Positive uint64 allocated by App's public event stream. Distinct from every source/session/query counter. |
| ConfirmationID | Opaque nonempty coordinator-issued single-use capability, bound to one operation, exact intent/plan and allowed choice. Private validation, no decoding by UI or adapters. |
| IntentToken | Client-issued monotonic uint64 for global current user intent; zero initial. Echoed by App; no Git/storage freshness meaning. |
| QueryGeneration | Client-issued monotonic uint64 within a QuerySlot; new accepted request replaces older query in that slot. |
| QuerySlot | Comparable opaque client slot key, including widget/purpose and explicit source scope as needed. App uses equality for supersession; source identity is separately validated. It never grants mutation authority. |
| Correlation | IntentToken plus Optional of a pair QuerySlot/QueryGeneration. Queries require the pair; mutations cannot request query supersession. |
| SourceVersion | Opaque comparable namespace/scope/issuer-lifetime/token value minted by Git, remote or Discovery observation. Equality only; no timestamp ordering or cross-adapter comparison. |
| StorageVersion | Opaque comparable whole-document content value with bound family/store identity, missing/present state, byte length and SHA256. Store identity can represent expected absence below a validated nearest parent; a token is never a filesystem handle. |
| ObservationID | Opaque immutable source-observation identifier. Distinct from object identity, update time and fetch generation. |
| ObservationInterval | StartedAt and FinishedAt UTC time values, supplied by an outer observer; ordered bounds, not a global snapshot/lease or proof of future freshness. Domain and View read no clock. |
| RuntimeEventSequence | Global Runtime-stream uint64; zero read/ACK cursor means beginning, events positive and monotonic. Dropped coalescible hints may leave gaps; retained critical outcomes may not be skipped by ACK. |
| SessionSequence | Positive Runtime per-session version for lifecycle/output facts; no relation to App Sequence or UI generation. |
| ContextVersion | Application-owned monotonic uint64 for authoritative active context; zero uninitialized. A query returns the version it captured, not its later completion time. |
| Completeness | Closed Complete, More, Partial, Unknown, Unavailable. Complete alone can support conclusive absence; More means bounded known remaining data; Partial carries independent valid facts plus failures. |
| PageRequest | Limit uint32 and closed Initial, Cursor(string), or Offset(uint64) continuation. The owning family states permitted mode and cap. A caller cannot combine cursor and offset or use another source's cursor. |
| PageInfo | Returned uint32, Completeness, Optional next cursor/offset, Optional HasMore bool and source version. Unknown/partial never becomes empty complete success. |

All source intervals/versions travel with their corresponding result. A remote PR
UpdatedAt is a remote fact, not ObservationInterval. Runtime output sequence is
not evidence that a query projection is current. FetchGeneration remains the
Git-family opaque generation defined in its BC, not a SourceVersion alias.

## B3. Worktree and remote binding values

```go
type WorktreeScope struct {
    ID           domain.WorktreeID
    RootLocator  string
    RootIdentity DirectoryIdentity
    Source       SourceVersion
}
type DirectoryIdentity struct { /* private platform-tagged observed identity */ }
type CwdObservation struct {
    Worktree          WorktreeScope
    ProjectComponents []string
    ProjectIdentity   DirectoryIdentity
    Source            SourceVersion
}
```

WorktreeScope means an actually observed registered worktree with an available
physically canonical root. Missing/unresolved worktrees retain WorktreeID in
their facts and have no fabricated Scope/Head. Scope's repository agrees with
WorktreeID's LocalCommon repository. RootLocator is evidence/location, not identity.
DirectoryIdentity is Windows volume/file ID or Unix device/inode plus supported
birth/change stamp; unavailable identity cannot be a valid positive value.
It is a short-lived observation, not a permanent Domain ID or native handle.

ProjectComponents is an exact relative component list: empty means root; no empty
interior component, NUL, absolute root or `..`. Preserve accepted whitespace and
bytes rather than trimming into another path. Discovery validates physical no-link
scope; Runtime acquires it through CwdAcquisition/WindowsBroker. Persistence uses
the selected WorktreeScope and its fixed `.gh-tree` child, not ProjectComponents.

RemoteBinding contains LocalRepository, RemoteRepository, configured RemoteName,
sanitized observed fetch/push URL locators, exact refspec mappings and configuration
SourceVersion. Both repository scopes are explicit and different. Credentials in
native URLs remain adapter-private; sanitized locators plus opaque configuration
fingerprints identify the binding. App associates independently verified Git and
GitHub facts. Equal branch names/OID bytes or the word `origin` alone cannot bind
repositories or authorize another clone/worktree. Ref mapping/source changes
invalidate dependent plans; no concrete adapter calls another to perform the join.

## B4. Execution intent and presentation-safe summaries

```go
type Geometry struct { Rows, Columns int }
type Invocation struct {
    Execution   ExecutionIntent
    Environment EnvironmentPolicy
    Cwd         CwdObservation
    Terminal    TerminalMode
    Geometry    Geometry
    Label       string
}
type EnvironmentPolicy struct {
    InheritBase bool
    Set         []EnvironmentEntry
    Remove      []string
}
type EnvironmentEntry struct { Name, Value string }
```

ExecutionIntent is closed ArgvExecution(Executable string, Arguments []string)
or InteractiveShell(ShellPolicy). An ArgvExecution names one nonempty literal
executable and literal argv; no shell-source concatenation is allowed. InteractiveShell
explicitly selects Auto or an explicit executable/argv. Runtime owns platform
shell/ancestry/GH_TREE_SHELL resolution, preserving the main application's ancestry
instead of mistaking a broker for the user's shell. App chooses that workflow,
not native process mechanics. InteractiveShell requires terminal mode.

TerminalMode is Pipes or Terminal. Geometry is positive rows/columns <=32767;
nonterminal use retains a valid value but cannot imply resize/input capability.
EnvironmentPolicy is a copied constructor-environment snapshot plus explicit
replacement/removal intent; keys are unique under platform environment rules,
NUL is invalid, and a value can be intentionally empty. Runtime retains the full
resolved environment privately for restart. Executable resolution uses the accepted
cwd and copied environment, not main-app cwd. Unknown platform operand/driver
semantics refuse explicitly rather than interpreting argv as another command.

InvocationSummary exposes Label, safe executable/argument display, accepted cwd
locator/identity, terminal and geometry facts, never an environment/credential
dump. Full restart specification remains private. Opaque caller arguments are
not automatically copied into diagnostics. User-program output is handled by the
Runtime output contract and is not automatically published to external systems.
Native helper paths/control secrets, Job/HPCON handles and private protocol frames
cannot leak into the product API or viewmodel.

### Windows batch carrier

The reviewed Windows .cmd/.bat carrier is the sole additional mechanical shell
transport for ArgvExecution. Runtime resolves the actual batch extension and uses
the native broker's system cmd.exe with `/D /V:OFF /S /C`; it does not choose a
driver by provider name or accept a free-form command string. Every executable/
argument is quoted, trailing backslash runs are doubled for the downstream native
argv parser, and one enclosing quote pair protects the /S /C command. Reject NUL,
CR/LF, embedded quote or percent in this carrier's operands before execution;
never expand environment syntax or silently trim into another script name.
Delayed expansion is disabled; tested literal exclamation, ampersand, parentheses,
colon, slash, whitespace, Unicode, empty argument and trailing backslashes survive.
Native executable argv keeps its native quoting policy; no other extension gets
an implicit shell fallback. This carries literal arguments into ordinary shims,
not a guarantee about arbitrary code inside a user-provided batch program.
The bounded native carrier fixture is archived under Feasibility/Batch; actual
Windows npm/pnpm/yarn integration remains V-LCH/V-RUN implementation proof.

## B5. Diagnostics, transport and effects

Diagnostic contains Code, stable Reason, safe Message, source/operation context
and typed optional detail. It may implement Go error without losing the structured
value. Code is the accepted Invalid, Unavailable, NotFound, Busy, Canceled,
Superseded, StaleObservation, ConfirmationStale, Conflict, Permission, Unsupported,
IOFailure, ProcessFailure, CleanupIncomplete or Indeterminate. Internal unexpected
failure normalizes with its actual effect uncertainty. Native exit/stderr remains
bounded diagnostic evidence, never the only parser of semantic truth.

```go
type CommandTransportOutcome struct {
    Started, RootReaped, CleanupKnown bool
    ExitCode Optional[int]
    StdoutTruncated, StderrTruncated bool
    CancellationRequested bool
    Diagnostics []Diagnostic
}
type EffectReport struct { Facets []FacetEffect }
type FacetEffect struct {
    Facet EffectFacet
    State EffectState
    PostObservation Optional[ObservationID]
    RecoveryIDs []RecoveryID
}
```

EffectFacet is ObjectAcquisition, Recovery, WorktreeBytes, Index, LocalRefsHead,
RemoteRefsPR, Storage or RuntimeResources. EffectState is NotStarted,
VerifiedNoTargetChange, AppliedVerified, Partial or Indeterminate. They are closed
semantic values, not sortable progress levels. Each post-observation/recovery ID
must reference the corresponding typed returned facts/RecoveryRecord; it cannot
be a dangling promise of evidence. Families may return several records for one
facet with distinct exact subjects. Cancellation request is separate from effect.

RecoveryRecord contains opaque RecoveryID, kind, responsible layer, exact subject
IDs, safe locator, original/proposed versions and conservative next-action text.
It contains no executable recovery callback. Captured objects/bytes/refs survive
unknown outcomes; a notice is not permission to overwrite a current destination.
Operation-level recovery lists are the union of referenced family records, not
another competing recovery authority.

A port's nonnil error supplements its typed result. App must inspect and preserve
valid facts, effects, original/recovery identity and mutation receipts first.
In particular, known mutation plus failed refresh is still known mutation; canceled
native commands are not rollback; source absence requires complete evidence.
Default CommandContext is root-only cancellation, not Runtime descendant proof.
Unresolved command cleanup/remote effects are retained and prevent unsafe automatic
retry. Read-only query supersession cannot erase a completed mutation outcome.

## B6. Family ownership and cross-contract conformance

| Family | Single normative detail owner |
|---|---|
| Application Request/Receipt/Event, closed public commands/queries and projections | BC--TUIState--Application |
| Git facts/requests/results/prepared plans/continuation receipts | BC--Application--Git |
| Remote repository/branch/PR facts, observation, create outcome | BC--Application--GitHub |
| Invocation resolution/discovery definitions/selections and saved intent | BC--Application--LaunchDiscovery |
| Whole documents/StoredField/OpaqueJSON/load/commit results | BC--Application--Persistence |
| Session snapshots/controls/start-stop-restart/output/events/shutdown | BC--Application--Runtime |
| Viewmodel records/layout measurement/rendering input | BC--TUIState--TUIView |

Storage's SavedLaunchDefinition is a retained semantic document value; Discovery
interprets it through App, never by calling Storage. Runtime accepts resolved
Invocation, never a provider DTO. Read-model joins and workflow sequencing remain
Application-owned even where the underlying public values share the API package.

The whole-set reviewer checks every referenced type's owner, method/request/result
pair, option/tag invariant, version scope, error/effect rule and identity/lifecycle
transition. A discrepancy cannot be implemented by a layer-local alias or generic
map. After freeze, incompatible changes follow BC-CHANGE, impact review/refreeze
and affected verification. This draft adds no permission to change accepted design.

## Verification and change history

Prove constructor/invalid/equality/copy invariants and imports under V-DOM-01..03,
V-APP-01..06 and family-specific checks. API cannot import ports/adapter types;
viewmodel cannot import API; source/cursor/operation/intent domains cannot be
interchanged. Formatters never repair invalid semantics into a plausible fact.

1.0.0 DRAFT: shared vocabulary specialized from accepted REFDES/API under #55.
Only the independently reviewed whole-set freeze changes implementation authority.
