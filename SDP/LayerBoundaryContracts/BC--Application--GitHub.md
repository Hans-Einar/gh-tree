# BC--Application--GitHub

State: DRAFT
Version: 1.0.0
Parent Issue: #55 under #21; accepted design #52 / PR #54
Applies to: complete v0.4 refactor, SLC-01..03/05/08 and cancellation in SLC-12
Design authority: merged `4a42222f7bfedc1d80693effbb25a1a82fcff65e`; technical acceptance `664f0c051344e3abdfd7d3c5698e4fbd3f584a83`
Supersedes: none. Freeze authority: none while DRAFT.

## GH1. Responsibilities and dependencies

Application owns remote selection, operation identity/lifetime, expected revision
intent, user confirmation where relevant, local/remote association, retry admission
and canonical branch/PR/worktree/graph projections. GitHub owns authenticated
remote repository/branch/PR observation and the one supported remote mutation,
CreatePullRequest. Its short-command transport is private implementation detail.

The concrete new implementation is `internal/github/adapter`. It implements
`internal/application/ports` and may import `application/api`, Domain and approved
low-level transport libraries. Application consumes ports; API imports no ports
or adapter. GitHub imports/calls no Git, Runtime, Discovery, Persistence, State,
View, coordinator or legacy `internal/github` parent client. Composition alone
wires the concrete adapter. No concrete JSON/GraphQL/CLI DTO crosses this boundary.

Normative shared types and copying rules are in [BoundaryTypes--001](BoundaryTypes--001.md).
Domain identities follow [API--001 A1](../Design/CR-%2321/API--001.md#a1).
The record definitions below belong to `application/api`; interfaces belong to
`application/ports`. Notation `Optional<T>` means an explicit absent/present
value, never a zero Domain ID. Closed alternatives require validating constructors;
they are not an open string discriminator or an `any` payload.

## GH2. Concrete port signatures

```go
// package ports; all request/result records are semantic api values.
type RemoteFacts interface {
    ResolveRepository(context.Context, api.ResolveRepositoryRequest) (api.ResolveRepositoryResult, error)
    ListBranches(context.Context, api.ListBranchesRequest) (api.ListBranchesResult, error)
    ListPullRequests(context.Context, api.ListPullRequestsRequest) (api.ListPullRequestsResult, error)
    ObservePullRequest(context.Context, api.ObservePullRequestRequest) (api.ObservePullRequestResult, error)
}
type RemoteMutations interface {
    CreatePullRequest(context.Context, api.CreatePullRequestRequest) (api.CreatePullRequestResult, error)
}
```

Every result carries `Observation: Optional<RemoteObservation>`,
`Diagnostics: []Diagnostic`, and `Transport: CommandTransportOutcome`. Mutation
results additionally carry `Effects: EffectReport`, `CancellationRequested: bool`
and `Recovery: []RecoveryRecord`. A nonnil Go error supplements the typed result;
it never invalidates independently known facts/effects or permits dropping a
created identity. Invalid admission has no invocation and explicitly NotStarted
effects. Successful reads contain valid typed records; failures never return
fabricated zero-value records as success. Application normalizes the single public
operation terminal; these ports emit no operation or UI events.

## GH3. Remote identity and semantic records

| API record | Fields and required semantics |
|---|---|
| RemoteRepositoryLocator | `Host: string`, `Owner: string`, `Name: string`; explicit nonempty authenticated DNS host and repository components. Owner/name shorthand has its host resolved once before use. Transport URL syntax and one terminal `.git` suffix may be removed; DNS host and GitHub owner/name canonicalize to lower case. Credentials, query fragments, ambiguous components and malformed identity refuse. |
| RemoteRepository | `ID: domain.RepositoryID`, `Locator: RemoteRepositoryLocator`, `URL: string`, `DefaultBranch: Optional<domain.BranchID>`, `Capabilities: RemoteCapabilities`; ID is Remote scope, never LocalCommon. URL is validated against the returned host/repository, not arbitrary stdout. Rename/transfer is an explicit new association, not silent migration. |
| RemoteCapabilities | `ReadBranches: bool`, `ReadPullRequests: bool`, `CreatePullRequest: bool`, `SupportedObjectFormats: []domain.ObjectFormat`, `Diagnostics: []Diagnostic`; capability absence includes its prerequisite/unsupported reason. A valid Domain SHA-256 does not assert provider support. |
| RemoteObservation | `ID: ObservationID`, `Repository: domain.RepositoryID`, `Interval: ObservationInterval`, `Version: SourceVersion`, `Origin: RemoteObservationOrigin`, `Page: PageInfo`; Origin is Live or Cached with original observation identity/interval. SourceVersion is equality for this observation, not a provider transaction or application query generation. |
| RemoteBranchFact | `Branch: domain.BranchID`, `Tip: domain.Revision`, `Observation: RemoteObservation`; Branch is RemoteHead in its own remote repository and Tip has that same scope. A remote-tracking local ref is never this record. |
| PullRequestEndpoint | closed `Available(Repository: RemoteRepository, Branch: domain.BranchID, Revision: domain.Revision)` or `Unavailable(KnownRepository: Optional<domain.RepositoryID>, KnownBranch: Optional<domain.BranchID>, KnownRevision: Optional<domain.Revision>, Reason: EndpointUnavailableReason)`. Reason is Deleted, Inaccessible, MissingField or Unresolved. Each retained optional identity is independently validated. No missing fork/head is defaulted to the base repository. |
| PullRequestFact | `ID: domain.PRID`, `URL: string`, `Title: string`, `Body: string`, `State: PullRequestState`, `Draft: bool`, `MaintainerCanModify: Optional<bool>`, `Base: PullRequestEndpoint`, `Head: PullRequestEndpoint`, `UpdatedAt: Optional<time.Time>`, `Observation: RemoteObservation`. State is Open, Closed, Merged or Unknown with diagnostic. ID is scoped to base remote repository, positive number. UpdatedAt is a remote property timestamp, never observation time. Body may be omitted by the requested projection only through an explicit field-presence result. |
| RemoteBranchFilter | copied exact branch-name prefix or All; filter affects PageInfo scope. It never converts branch names into local refs. |
| PullRequestFilter | `State: Open/Closed/Merged/All`, `Head: Optional<domain.BranchID>`, `Base: Optional<domain.BranchID>`; qualifying head includes its repository so fork heads remain distinct. The default navigator requests Open including drafts. |
| EndpointExpectation | `Branch: domain.BranchID`, `Revision: domain.Revision`, `Observation: RemoteObservation`; Branch and Revision share remote scope and come from an explicitly verified published endpoint. |

All OIDs are full nonzero SHA-1/SHA-256 Domain values. A branch name is preserved
as exact validated identity bytes. Labels, URL text, server timestamps and PR
state do not enter stable equality. Repository resolution requires remote
identity/authentication evidence; syntactically valid input alone is not proof
of repository existence or permission. Remote-only browsing and fully specified
PR creation do not require a local checkout.

| Request | Additional exact fields | Matching result payload |
|---|---|---|
| ResolveRepositoryRequest | `Locator: RemoteRepositoryLocator` | ResolveRepositoryResult: `Repository: Optional<RemoteRepository>`; unavailable/auth/ambiguous input is diagnostic, not another host fallback. |
| ListBranchesRequest | `Repository: domain.RepositoryID`, `Filter: RemoteBranchFilter`, `Page: PageRequest` | ListBranchesResult: `Branches: []RemoteBranchFact`, `Page: PageInfo`; observation includes actual queried scope. |
| ListPullRequestsRequest | `Repository: domain.RepositoryID`, `Filter: PullRequestFilter`, `Page: PageRequest` | ListPullRequestsResult: `PullRequests: []PullRequestFact`, `Page: PageInfo`; repository is the PR/base scope. |
| ObservePullRequestRequest | `Target: domain.PRID`, `ExpectedHead: Optional<domain.Revision>`, `ExpectedBase: Optional<domain.Revision>` | ObservePullRequestResult: `PullRequest: Optional<PullRequestFact>`, `Expectation: NotRequested/Matched/Mismatched/Unresolved`; mismatching current facts may be returned with StaleObservation, but never replace either expected endpoint. |

`ResolveRepository` registers a verified scope-to-locator association inside the
adapter lifetime. Other methods reject foreign/unknown scopes rather than decode
an opaque token into an unchecked URL. Calls carry no local cwd, remote name or
GitHub DTO. App associates a remote repository with Git's independently observed
`RemoteBinding`; host/name or equal OID bytes alone do not merge local scopes.

## GH4. Paging, freshness and immutable observations

PageRequest binds repository, query/filter/order, positive limit and an explicit
cursor. Default page size is 100; at most ten pages belong to one refresh. Bounds
are applied before invocation; no unbounded `--paginate` traversal is permitted.
The adapter uses API-supported stable ordering where available and returns the
actual ordering/cursor semantics. Cursor reuse with a different repository,
filter or ordering refuses. A page is not a transaction with preceding pages.

PageInfo carries returned count, optional next cursor, observation scope and
Completeness: Complete, More or Unknown. At a known cap with more data use More;
interruption or unprovable exhaustion uses Unknown. Empty success is Complete
only when exhaustive valid response evidence establishes it. Partial/cached/capped
or failed lists never prove that a branch/PR is gone. Duplicates with identical
identity/facts may be explicitly deduplicated; conflicting observations retain
their intervals and report changed/inconsistent pagination, never overwrite one
silently and assert a snapshot. Malformed individual records are diagnostics;
only independently valid records survive, with Unknown completeness.

Each actual request brackets its acquisition interval. Aggregation covers the
earliest start/latest finish and preserves constituent page provenance. Server
updated times and completion time alone cannot stand in for acquisition time.
Git FetchGeneration, remote ObservationID, SourceVersion, Application OperationID,
QueryGeneration and TUI intent are independent. A local fetch cannot refresh a
GitHub observation. App alone chooses partial projections and rejects stale query
completions; GitHub returns no focus or selection instruction.

Requests/results, including nested endpoint records, pages, diagnostics and
buffers, are copied on both boundaries. No API object points into transport DTOs,
parser buffers or an adapter cache. Concurrent calls cannot mutate an earlier
published value.

## GH5. CreatePullRequest scope and result protocol

```go
// package api; semantic records, no transport fields or local Git handles.
type CreatePullRequestRequest struct {
    Operation             OperationID
    Base                  EndpointExpectation
    Head                  EndpointExpectation
    Title                 string
    Body                  string
    Draft                 bool
    MaintainerCanModify   bool
}
```

The coordinator passes its assigned OperationID for correlation, not server
idempotency. Title is nonblank; body is literal and may be empty. Both endpoints
are remote repository-qualified branches with exact expected published revisions
and their observations. Head/base repository hosts must agree; a cross-host PR
is Unsupported. Base repository owns the resulting PRID. Existing create behavior
allows maintainer modification; Application supplies true unless an independently
authorized product change changes that policy. This contract adds no UI setting.

Application first validates its intended local/upstream association through Git
when the use case began locally, then observes the actual published head/base
through RemoteFacts. GitHub reobserves expected remote endpoints before sending
the request; missing/unpublished/mismatched/unresolved endpoints refuse before
submission. No selected local branch name may substitute for the explicit remote
head. The authenticated `gh api`/GraphQL request includes explicit host, base
repository, repository-qualified head/base, literal title/body and draft state.
It must not invoke interactive `gh pr create` behavior, push, fork, select a
checkout or edit local remote/upstream configuration. Provider restrictions on
fork qualification refuse explicitly without choosing a same-named base branch.

The server create API has no atomic expected-head-OID predicate. Preobservation
does not remove the race with remote ref movement. On a valid response, inspect
the actual created PR endpoints and identity. Post-create drift is a real remote
effect with mismatch/uncertainty, never a claimed rollback. No automatic close,
delete, edit or corrective push follows it.

CreatePullRequestResult has the common mutation envelope plus
`Outcome: PullRequestCreationOutcome`, a closed alternative:

| Outcome | Payload and evidence |
|---|---|
| NotSubmitted | `Reason: Diagnostic`; no create request started, PR effect NotStarted. |
| RejectedNoCreation | `Reason: Diagnostic`; explicit server rejection establishes no creation by this request. An exit status alone is insufficient. |
| ExistingCandidate | `Candidates: []PullRequestFact`, `Page: PageInfo`; an existing-PR response/query establishes candidates, not ownership by this operation or permission to adopt another user's PR. |
| CreatedVerified | `Created: PullRequestFact`, `RequestedBase: EndpointExpectation`, `RequestedHead: EndpointExpectation`; validated response/post-observation prove identity and matching endpoint facts. |
| CreatedWithDrift | `Created: PullRequestFact`, `RequestedBase: EndpointExpectation`, `RequestedHead: EndpointExpectation`, `Reason: Diagnostic`; creation known but one requested endpoint changed/unresolved. Preserve known PR creation effect separately from endpoint uncertainty. |
| CreationIndeterminate | `RequestEvidence: RemoteCreateEvidence`, `Candidate: Optional<PullRequestFact>`, `Reason: Diagnostic`; request may have committed, identity or outcome cannot be proved. |

RemoteCreateEvidence contains OperationID, exact requested head/base expectations,
bounded observation interval, optional provider request ID and independently
validated returned PR identity/URL evidence. It contains no credentials or raw
environment. It is diagnostic/reconciliation data, not a replay token.

Application publishes exactly one operation terminal and retains the requested
scopes/effects when follow-up refresh fails. It reconciles uncertain creation
through ObservePullRequest when identity is known, otherwise bounded
ListPullRequests with exact qualified head/base criteria and expected revision
evidence. The port has no generic retry or mutation dispatcher. Incomplete lists,
current equal refs, similar title/body or a response-lost timeout cannot prove no
creation or causal ownership. A fresh operation may retry only after Application
establishes the permitted effect; no blind resubmission or generic exactly-once
server guarantee is claimed.

## GH6. Transport lifecycle, concurrency, cancellation and diagnostics

Adapter construction supplies supported authenticated gh host/profile and limits.
Requests are deterministic and noninteractive, without auth-login/auth-refresh,
setup-git or global user configuration writes. Use explicit host on every API
request; never re-read a different implicit default host midway. Credentials and
sensitive inherited environment are excluded from diagnostics/logs.

Machine stdout and diagnostic stderr are separate copied bounded streams: initial
ceilings 16MiB and 256KiB respectively. JSON warnings never become payload; a
nonempty output line is not a created PR URL. Strictly validate response shape,
required positive identities, enums/optional nulls, full OIDs and canonical scope.
Limit refusal and truncation are explicit; partial valid observations are kept
only when structurally independent. Mutation payload truncation after send may
require CreationIndeterminate, never RejectedNoCreation.

Read commands have a 30s budget; mutating/network/hook commands have 120s;
forced pipe-drain/join after cancellation is separately bounded by 2s. These are
adapter construction limits, not UI waiting periods. Each command owns one root
waiter and joined pipe readers. CommandTransportOutcome reports whether it
started, was reaped, and has known transport cleanup, with limit/timeout details.
Killing the direct root and closing pipes cannot prove descendant cleanup or
remote rollback. GitHub cannot call Runtime's private supervisor to imply that
guarantee. Stronger transport ownership requires implementation/review inside
this adapter or an accepted design/BC change.

Read cancellation returns Canceled plus truthful incomplete observations.
Cancellation before mutation submission is distinct from cancellation after it
may have reached the server. The adapter owns bounded reap/reconciliation even
after the request context cancels. Application retains unresolved mutation and
command-cleanup notices, blocks unsafe retry in the affected scope, and includes
residuals in shutdown. A closed modal does not release that responsibility.

Reads may run concurrently; values and caches remain race-safe. Application
orders its own same-intent create/reconciliation operations; neither this ordering
nor an adapter mutex creates a remote transaction against other actors. The
adapter has no event subscription, session registry, unbounded work queue or
independent background mutation retry. Shutdown joins all admitted port work
through Application's operation barrier and constructor-owned resource teardown.

Diagnostic uses the shared ErrorCode with structured remote details: Invalid,
Unavailable, NotFound, Busy, Canceled, StaleObservation, Conflict, Permission,
Unsupported, IOFailure, ProcessFailure, CleanupIncomplete or Indeterminate as
applicable. Remote detail includes optional HTTP status, provider request ID,
rate-limit/retry-after/reset facts and endpoint scope. Authentication/private-404
ambiguity is not proven absence; 403/429 retain rate/permission evidence. Bad
schema/unsupported CLI/API response is explicit protocol detail, not an empty
repository. Localized prose or native exit status alone never manufacture facts.

## GH7. Forbidden behavior

- No local fetch/push/checkout/ref/configuration operation, implicit fork, implicit
  repository guessing or replacement of exact Revision with latest branch tip.
- No PR merge/close/edit/delete, issue/review/label/release feature through this
  boundary; its only product mutation is CreatePullRequest.
- No concrete adapter DTOs in API, cross-source reconciliation in GitHub, mutable
  publication, State callback, Application Client or active-worktree authority.
- No capped-list absence, transport-canceled rollback, arbitrary stdout success,
  ignored nonnil-error effect data, or retry before uncertainty reconciliation.

## GH8. Verification and freeze control

Required exact-product checks are [Verification--001](../Design/CR-%2321/Verification--001.md)
V-GH-01..03, V-APP-01..03/05/06, V-DOM-01..03, V-GIT-02/03/09/10 and
V-COMP-01/03. SLC-01..03/05/08 retain their corresponding V-E2E scenarios;
SLC-12 covers canceled/detached-host transport work. Use raw API fixtures for
controlled creation/response-loss cases; label them as fixtures, not live remote
publication proof. Include host/fork/same-name ambiguity, null/deleted endpoints,
malformed OIDs, overlapping pages, >100 results, capped absence, pre/post-create
drift, duplicated candidates and stderr separation. Windows/Linux/macOS command
resource evidence and twelve selected-platform import/build checks are required.

Read this with [BC--Application--Git](BC--Application--Git.md), shared types and
the State/Application boundary; no component can claim conformance in isolation.
Accepted review/feasibility records do not prove future product implementation.

Change history: 2026-09-06, DRAFT 1.0.0 authored under #55. Separate fresh
whole-set review, correction and exact-HEAD re-review precede REVIEWED/FROZEN.
Master records the source/freeze/merge SHAs and configured CI. Once FROZEN, an
incompatible signature, scope, effect or lifecycle change requires BC-CHANGE,
design impact review, revised freeze and affected re-verification; a worker must
stop affected work rather than bypass the boundary.
