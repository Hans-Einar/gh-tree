# BC--Application--LaunchDiscovery

State: DRAFT
Version: 1.0.0
Authority: #55/#21; accepted #52/PR54, merge4a42222f7bfedc1d80693effbb25a1a82fcff65e.
Applies to: SLC-09/10, worktree/cwd prerequisites and cancellation in12.
Supersedes: none; no implementation authority before whole-set freeze.

## Responsibilities and dependency direction

Application chooses worktree and ordered launch intent, loads saved definitions
through Storage, owns alias/default/replace policy, query/operation correlation
and Runtime Start sequencing. Discovery passively observes provider manifests,
constructs stable semantic definitions and resolves an exact selection into a
validated Invocation/CwdObservation. Persistence alone owns run.json codec/write;
Runtime alone owns shell/executable execution, sessions and native acquired cwd.

`launchdiscovery -> application/ports + application/api + domain + standard/native
read-only filesystem helpers`; Application imports ports, not the concrete adapter.
Discovery imports/calls no Runtime, Persistence, Git, GitHub, State or View. It
spawns no process, runs no provider/version command, changes no cwd and writes no
file while discovering/resolving. Provider parsing types remain private; Domain
owns LaunchPointID only, not manifests/argv/provider DTO collections.

[BoundaryTypes--001](BoundaryTypes--001.md),
[API A1/A3/A5/A6](../Design/CR-%2321/API--001.md),
[CwdAcquisition](../Design/CR-%2321/CwdAcquisition--001.md) and
[Storage](../Design/CR-%2321/Storage--001.md) are normative.

## Typed port and records

```go
// package ports
type LaunchDiscovery interface {
    Discover(context.Context, api.DiscoveryRequest) (api.DiscoveryResult, error)
    Resolve(context.Context, api.ResolveLaunchRequest) (api.ResolveLaunchResult, error)
}
// package api
type DiscoveryRequest struct {
    Worktree WorktreeScope
    Saved []SavedLaunchEntry
    SavedVersion Optional[StorageVersion]
}
type DiscoveryResult struct {
    WorktreeID domain.WorktreeID
    Definitions []LaunchDefinition
    Saved []SavedLaunchObservation
    Observation DiscoveryObservation
    Diagnostics []Diagnostic
}
type ResolveLaunchRequest struct {
    Worktree WorktreeScope
    Selection LaunchSelection
    Saved []SavedLaunchEntry
    SavedVersion Optional[StorageVersion]
    Geometry Geometry
}
type ResolveLaunchResult struct {
    Definition Optional[ResolvedLaunchDefinition]
    Invocation Optional[Invocation]
    Observation DiscoveryObservation
    Diagnostics []Diagnostic
}
```

SavedLaunchEntry/Definition are the API semantic retained document values in the
Persistence BC, including unknown provider definitions. App passes copies from
its actual versioned load; Discovery does not fetch the file itself. SavedVersion
is present exactly when supplied saved entries come from a valid load, including
known absence. It is not a discovery source token or an unverified current default.

| Record | Exact fields/semantics |
|---|---|
| DiscoveryObservation | ObservationID, WorktreeID, ObservationInterval, SourceVersion, Completeness, visited/skipped counts, provider-profile versions and per-source diagnostics |
| LaunchDefinition | LaunchPointID, ProviderKind, exact ProjectComponents, member identity, display path/label, ProjectSource, EffectiveExecutable intent, validation/availability diagnostics |
| ProjectSource | exact manifest locator/identity, content SourceVersion, relevant regular lock/config observations, parser profile, physical root/project DirectoryIdentity |
| ProviderKind | closed Npm or Make for executable definitions. Unknown stored providers remain preserved unavailable records, not a runtime routing string. |
| LaunchSelection | closed Discovered(single ID+SourceVersion), OrderedMake(nonempty ordered MemberSelection list), or Saved(exact alias+LaunchPointID+StorageVersion+source expectation) |
| MemberSelection | LaunchPointID and exact source version; ordering is semantic. It cannot select a label/cursor/index as identity. |
| SavedLaunchObservation | exact alias, optional resolvable LaunchPointID/definition, storage/source versions and diagnostics. App joins the stored default separately; Discovery has no default authority. |
| ResolvedLaunchDefinition | exact selected IDs/alias/order, provider/project/manifest/effective-executable facts and the revalidated source versions used to construct Invocation |

LaunchPointID uses WorktreeID plus unambiguous length-delimited provider/project/
member identity. npm root script `a/b` cannot collide with nested `a` script `b`,
and npm `dev` cannot collide with Make `dev`. Label/display grouping, file content
version and executable override are not stable ID equality. Two aliases may name
the same underlying ID; Saved selection retains the exact alias and version.
Duplicate provider registration refuses construction; deterministic sort uses
project bytes, provider key, exact member bytes and stable ID, never map order.

## Passive discovery and bounded source policy

Before saved/default execution, App reloads the run document and checks the selected
StorageVersion (or explicitly resolves the current default choice) before supplying
entries to Resolve. Discovery checks the supplied binding but cannot prove storage
freshness by rereading a file it does not own. A later native Start binds cwd and
literal invocation; it does not promise immutable future project code.

Initial construction limits: depth5 relative project components;10,000 directories;
10,000 candidates;4MiB per manifest;1MiB per Make line. Limits are explicit tested
configuration, not silent truncation. Cancellation is checked during walk/read/
parse and returns independent partial facts plus Canceled/diagnostics as appropriate.
Skipped/failed source is not marked successfully loaded or complete.

Retain existing case-folded generated-directory exclusions: `.git`, `.gh-tree`,
`node_modules`, `vendor`, `dist`, `build`, `out`, `target`, `.next`, `.cache`.
Do not follow unproven child symlinks/reparse points, leave worktree scope, scan
ancestors for workspace manifests or execute Make to discover dynamic targets.
One malformed/permission-failed project does not erase independent valid projects.
Completeness refers to the stated passive profile and reports unsupported source
constructs/limits; it does not claim exhaustive native build-tool semantics.

Npm reads package.json scripts as exact string keys/values. Colon grouping is
display-only: `dev:wan` resolves as one argv element after `run`, never two scripts.
Preserve leading/trailing accepted script/path whitespace. Empty or malformed
members get explicit invalid/unavailable diagnostics, not a trimmed alternative.
Reject ambiguous/malformed JSON and invalid UTF-8 rather than silently replacing
identity bytes. Unknown unrelated package members do not affect provider ownership.

Manager selection preserves colocated regular `pnpm-lock.yaml` precedence, then
regular `yarn.lock`, otherwise npm (including package-lock/npm-shrinkwrap/no lock).
Directories or redirected lock paths do not count as valid lock files. Conflicting
locks/packageManager declarations produce diagnostics; the stated compatibility
precedence remains deterministic. No ancestor-lock inheritance is added. An explicit
saved executable override remains authoritative after validation; do not silently
replace it with detection. Record all inputs affecting this decision in versioning.

Make chooses GNUmakefile, then makefile, then Makefile for the default supported
GNU Make profile, and binds the selected file explicitly with `-f`. Discover simple
textual rule targets without recipes, dot-special/pattern/assignment/option-shaped
operands or unsupported slash/macros. Includes/conditionals/dynamic targets produce
declared profile limitations; textual membership is not a promise native execution
will succeed. Accepted target grammar retains letters/digits/underscore/dash/dot,
but rejects empty, leading dash/dot, assignment and shell/Make-control syntax.
Do not silently broaden into a full Make interpreter.

## Resolution, exact argv and cwd handoff

Resolve rereads/revalidates root/project/manifest/member/manager/override versions
against the selected identity. Changed, foreign, missing or unproved selection
returns StaleObservation/Invalid/Unsupported and no Invocation. It never substitutes
the current cursor, another project or newly detected manager under an old token.
For OrderedMake every member must share the same worktree, physical project,
selected manifest/version and effective executable policy; order remains exactly
the user's list. Reject incompatible selection even if the UI previously allowed it.

Produce ArgvExecution with literal executable plus `run, exactScript` for npm,
or `-f, selectedManifest, orderedTargets...` for Make. No user shell source is
constructed by Discovery. Runtime handles an explicitly reviewed native Windows
batch transport when the chosen executable resolves to .cmd/.bat; supported argv
must remain literal, and unsupported grammar refuses before execution. Other
platform/executable kinds cannot silently fall back to a shell.

CwdObservation binds the supplied available WorktreeScope and exact relative
project components with newly observed root/project identities/source versions.
Physical no-link validation occurs in Discovery, followed by independent Runtime
acquisition through Unix descriptor/Fchdir or the accepted Windows broker/barrier.
No open handle is passed to another adapter. These are staged observations and
an acquisition contract, not immutable project code or a filesystem sandbox after
Start. Runtime never looks up a current UI cursor/provider manager to execute it.

Save/default orchestration remains App -> Storage. Discovery can return normalized
definition values but never writes JSON, picks a replacement alias, decides whether
first save becomes default or changes active worktree. First-save/default and
version-bound alias replacement are proven under the Persistence/Application BCs.

## Errors, concurrency and forbidden behavior

Discover/Resolve are reentrant read operations with no live registry/process.
Any private cache is bounded, immutable/versioned, and cannot convert a failed
read into freshness. Context cancel stops owned observation and joins any readers;
no detached scan continues after the operation returns. Result+error preserves
independently valid partial facts; absence requires complete applicable evidence.
App owns query supersession and exactly one terminal event.

Forbidden: calling providers to discover targets, running install/build commands,
Runtime session ownership, JSON writes, Git/GitHub lookup, active-worktree choice,
foreign-worktree acceptance, caller-fabricated unversioned Candidate execution,
trimming/colon splitting/option injection, or display sorting as launch identity.

## Verification and history

V-LCH-01..03, V-PER-01/04, V-APP-02/06, V-RUN-01/08 and V-E2E-09/10 prove this
boundary. Include nested projects/exclusions/limits/partial errors, exact colon/
whitespace/member identity, source/cwd replacement, conflicting locks/overrides,
all selected Make members from different projects, selected manifest precedence,
ordered targets, saved/default roundtrip and provider-free passive scans. Native
execution tests verify actual supported argv/cwd, including Windows batch drivers;
construction-only assertions do not prove shell transport.

1.0.0 DRAFT under #55. Whole-set review precedes freeze; any incompatible change
afterwards follows BC-CHANGE/impact/refreeze and affected verification.
