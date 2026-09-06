# BC--Application--Persistence

State: DRAFT

Version: 1.0.0

Parent Issue: #55, under #21; accepted design #52 / PR #54

Applies to: complete v0.4 refactor; SLC-01/04/05/09/10/13 and shutdown obligations

Accepted design merge: `4a42222f7bfedc1d80693effbb25a1a82fcff65e`

Supersedes: none; no implementation authority until the whole-set freeze gate.

## Authority and responsibilities

This contract specializes [API A1..A6](../Design/CR-%2321/API--001.md),
[Storage--001](../Design/CR-%2321/Storage--001.md),
[REFDES](../Design/CR-%2321/REFDES--001.md) and the selected native protocol in
[Persistence feasibility sections 1..5](../Design/CR-%2321/Feasibility/Persistence.md).
Storage--001's later actual list/data-read Windows guards supersede the older
metadata-only guard description in that feasibility history. These complete
protocols are normative. [BoundaryTypes--001](BoundaryTypes--001.md) owns shared
scope/value/error rules; the typed Storage family below contains no adapter DTO.

Application owns durable intent, active-worktree context, exact scope association,
configured-target validation, migration decisions, save/alias/default policy,
version-bound confirmation, operation lifecycle and commit-before-publication.
It obtains current Git worktree identity and remote/local associations; Persistence
does not choose them. Discovery interprets provider definitions supplied by
Application; storage neither validates provider execution semantics nor starts it.

Persistence owns physical store binding, no-follow object acquisition, bounded
read/codec, immutable documents and whole-document versions, cooperative locks,
native preparation/publication, supported permissions, precise commit effects,
recovery evidence and cleanup of exact owned request resources. Composition
selects user configuration/state paths and lazy default policy. Persistence has
no background flush queue, session registry or business transaction coordinator.

## Dependency direction and type ownership

`application/usecases -> application/ports -> application/api + domain`;
`persistence -> application/ports + application/api + domain + approved native
filesystem/codec libraries`. Application imports no concrete Persistence. API
imports no ports. Domain has no JSON tags, path/URL parsing, storage schema or
filesystem calls. No concrete adapter imports/calls another.

Ports owns Storage and typed request/load wrappers. API owns complete document
values, semantic settings and commit/load observations. Public `api.StorageVersion`
is an opaque comparable whole-document content value; `ports.StorageVersion`
aliases it. Its private codec/validation state does not become a public adapter
type or Application dependency on ports. Same rule applies to SourceVersion.
Runtime IDs, Application operation IDs, source versions and StorageVersion are
different types and cannot substitute for expected content version.

## Commands / inputs: exact Go surface

User-config and preferences physical bindings are constructor-owned, selected
once through Composition's supplied locations. A single Storage instance may
serve many Git-issued worktree scopes; run storage always receives its explicit
scope. No caller chooses a basename, arbitrary project storage path or native
handle through a commit request.

```go
type StorageVersion = api.StorageVersion

type Storage interface {
    LoadUserConfig(context.Context) (LoadedUserConfig, error)
    LoadPreferences(context.Context) (LoadedPreferences, error)
    LoadRunConfig(context.Context, api.WorktreeScope) (LoadedRunConfig, error)
    CommitUserConfig(context.Context, UserConfigCommit) (api.StorageCommitResult, error)
    CommitPreferences(context.Context, PreferencesCommit) (api.StorageCommitResult, error)
    CommitRunConfig(context.Context, RunConfigCommit) (api.StorageCommitResult, error)
}
type LoadedUserConfig struct {
    Observation api.StorageLoadObservation
    Document    api.Optional[api.UserConfigDocument]
}
type LoadedPreferences struct {
    Observation api.StorageLoadObservation
    Document    api.Optional[api.PreferencesDocument]
}
type LoadedRunConfig struct {
    Scope       api.WorktreeScope
    Observation api.StorageLoadObservation
    Document    api.Optional[api.RunConfigDocument]
}
type UserConfigCommit struct {
    Expected StorageVersion
    Document api.UserConfigDocument
}
type PreferencesCommit struct {
    Expected StorageVersion
    Document api.PreferencesDocument
}
type RunConfigCommit struct {
    Scope    api.WorktreeScope
    Expected StorageVersion
    Document api.RunConfigDocument
}
```

Each Document is the complete proposed immutable result of one Application-
validated durable intent patch against the loaded expected document. This is
typed whole-document commit, not a generic string-key dispatcher. Application
preserves unrelated known fields and all retained unknown fields, applies only
the authorized intent, and supplies schemaVersion 1. Persistence structurally
validates and verifies preserved unknown members against the guarded loaded
document; it never chooses alias/default/active context or silently merges a stale
snapshot. A conflict requires fresh load and revalidated Application intent.

Expected version is mandatory, including expected absence. A zero/foreign-store/
foreign-family token is Invalid before publication. A stale content version is
Conflict/NotCommitted. A project token from another WorktreeID/root object cannot
authorize this run document. User config and preferences remain different store
identities even if a caller attempts overlapping path configuration; conflicting
family bindings refuse rather than reinterpret one document as the other.

No mutation is inferred from Load, startup, help/version or shutdown. Expected
versions and document copies cross calls; open OS handles and native locks do
not. There is no callback, lazy reader, file pointer or storage transaction held
across a UI delay.

## Typed documents and codec field semantics

The following API shapes are semantic whole documents, not the concrete JSON
codec's DTOs. JSON naming/encoding belongs only to Persistence. `StoredField[T]`
has a closed presence tag Absent, Null or Present and carries a value only in
Present; invalid tag/value combinations refuse. It preserves distinctions such
as missing/null prefixes versus explicit empty lists and exact existing optional
fields without turning null into fabricated defaults.

`JSONMembers` is a copied ordered sequence of unique exact member names plus
`OpaqueJSON` values. OpaqueJSON is an immutable validated bounded JSON value
with private bytes and a copying accessor; callers cannot mutate decoder buffers.
These are narrowly for retention of unknown schema members, not a general API
command payload. Duplicate names, known-name collisions and invalid bytes refuse.
Every known object level below retains its own unknown members.

```go
type UserConfigDocument struct {
    SchemaVersion  uint32
    StripPrefixes  StoredField[[]string]
    LegacyRepos    StoredField[[]LegacyRepositoryConfig]
    ScopedRepos    StoredField[[]ScopedRepositoryConfig]
    UnknownMembers JSONMembers
}
type LegacyRepositoryConfig struct {
    Key   string
    Value ConfiguredRepository
}
type ScopedRepositoryConfig struct {
    RepositoryID domain.RepositoryID
    Value        ConfiguredRepository
}
type ConfiguredRepository struct {
    Worktrees      StoredField[[]ConfiguredWorktreeTarget]
    UnknownMembers JSONMembers
}
type ConfiguredWorktreeTarget struct {
    Name           string
    Path           StoredField[string]
    Branch         StoredField[string]
    UnknownMembers JSONMembers
}
type PreferencesDocument struct {
    SchemaVersion    uint32
    LegacyFolders    StoredField[[]LegacyStringPreference]
    LegacyWorktrees  StoredField[[]LegacyStringPreference]
    ScopedPreferences StoredField[[]ScopedPreference]
    UnknownMembers   JSONMembers
}
type LegacyStringPreference struct { Key, Value string }
type ScopedPreference struct {
    RepositoryID   domain.RepositoryID
    LastFolder     StoredField[string]
    ActiveWorktree  StoredField[StoredActiveWorktree]
    UnknownMembers JSONMembers
}
type StoredActiveWorktree struct {
    AdministrativeKey string
    LastKnownPath     string
    UnknownMembers    JSONMembers
}
type RunConfigDocument struct {
    SchemaVersion  uint32
    Default        StoredField[string]
    Launch         StoredField[[]SavedLaunchEntry]
    UnknownMembers JSONMembers
}
type SavedLaunchEntry struct {
    Alias      string
    Definition SavedLaunchDefinition
}
type SavedLaunchDefinition struct {
    Provider       string
    Dir            StoredField[string]
    Script         StoredField[string]
    Targets        StoredField[[]string]
    Command        StoredField[string]
    UnknownMembers JSONMembers
}
```

`SchemaVersion` is 0 only for a decoded supported legacy document and 1 for the
current schema; writers emit 1. A family load carries the original observed
schema separately, including unsupported versions. Known fields encode exactly:

| Family | JSON schema1 mapping and scope |
|---|---|
| User config | `schemaVersion`, `stripPrefixes`, legacy `repos`, new `scopedRepos`. Repositories contain `worktrees`; exact target-name keys contain path/branch and retained members. Scoped repo keys are Remote RepositoryID tokens. |
| Preferences | `schemaVersion`, legacy `lastFolders`/`lastWorktrees`, new `scopedPreferences`. Remote scopes may carry `lastFolder`; LocalCommon scopes may carry `activeWorktree:{administrativeKey,lastKnownPath}`. Parent scope supplies WorktreeID repository identity. |
| Run config | `schemaVersion`, existing `default`, `launch` exact alias map and each existing provider/dir/script/targets/command value. It is worktree-scoped and alias/default publish in one document. |

Canonical scoped keys serialize the supplied Remote/LocalCommon opaque token
unambiguously. Codec decoding of a canonical token does not establish an observed
repository association or validate a live Git worktree. Domain does not parse
JSON/paths. Unknown noncurrent scope entries are retained; lack of current
inventory is no reason to prune them. Scope/family-invalid known field placement
is a structural diagnostic, not permission to drop it during migration.

Maximum input and newly encoded document size is 4MiB, maximum nesting depth 64.
Known strings retain exact UTF-8 bytes, spaces and case; invalid UTF-8 refuses
rather than decoder replacement. Top-level null/array/nonobject, duplicate keys
at any depth, invalid known-field types, negative/noninteger schemaVersion and
ambiguous reserved-new-field shapes refuse. Missing schemaVersion or integer 0
is legacy; unknown positive versions are UnsupportedVersion and never rewritten.
Of the fields listed here, only StripPrefixes accepts Null, preserving the
documented legacy/default-list policy. Other absent optional fields use Absent;
a numeric/string type mismatch or null known object is not an empty/default
document. Unsupported version integers too large for uint32 remain unsupported
raw observations with an exact diagnostic, never a wrapped current version.

Unknown members at every known object level and unknown-provider definitions
are preserved. Persistence validates known field types but does not reject an
otherwise well-formed definition solely because its provider is unknown. Discovery
reports whether that entry can resolve. Formatting may normalize in a new file;
the entire original bytes remain separately retained. Unknown-member equality
is lossless JSON value preservation, not a promise to retain whitespace in newly
formatted known objects. Original backup is the exact byte-for-byte authority.

Missing/null legacy stripPrefixes uses the existing copied default list;
explicit `[]` is intentionally empty. Alias/target name equality is exact and
case-sensitive; display order uses `(folded label, exact label bytes, stable
identity)` as a total order. Absent/empty run default means none. First save
becomes default only if none exists or make-default is explicit. Existing alias
replacement requires Application's explicit approval bound to prior exact alias
and whole-document StorageVersion; two aliases may identify the same LaunchPointID.
Saved selection retains alias plus StorageVersion through Discovery resolution.

Command is the preserved executable override, not shell code. Ordered Make
targets and singular npm script bytes survive load/save/migration. Missing/empty
project dir denotes worktree root; nonempty bytes go unchanged to Discovery for
validation. Empty saved active path is invalid intent and never cleans to `.`.
Persistence returns canonical settings and legacy candidates without deciding
which one is effective for the current application context.

## Load results and versions

```go
type StorageLoadObservation struct {
    State          StorageLoadState
    Version        Optional[StorageVersion]
    SchemaVersion  Optional[uint32]
    Diagnostics    []Diagnostic
    Recovery       []StorageRecovery
}
```

StorageLoadState is Absent, ValidLegacy, ValidCurrent, Corrupt,
UnsupportedVersion, Unavailable or UnsupportedProfile. Version, when known,
binds the family, actual parent/store object identity and fixed basename,
missing/present state, exact complete byte length and SHA-256. The token exposes
equality, not filesystem authority or native handles. Recreating identical bytes
under the same bound store is the same content version; an edit restored exactly
is not claimed as a semantic conflict. Document-file identity/security is checked
separately at native publication. Parent/store replacement invalidates binding.
When an authorized parent does not yet exist, an absence token binds its nearest
validated existing anchor and exact remaining literal component sequence, with
observed absence explicit. Load creates no directories merely to mint a token.
Commit reacquires that same anchor, validates/creates each missing component
without following links, binds the actual resulting parent and rechecks document
absence under its permanent lock. A component already observed present cannot
be silently replaced; a newly appeared component must satisfy the same scope/
profile checks before it can be adopted. Unavailable or unprovable ancestry is
not absence. The result's version binds the now-established actual store.

Absent alone returns an immutable empty/default candidate Document with an
expected-absence Version; Load still creates no persistent state. ValidLegacy/
ValidCurrent return the complete typed Document and version. Corrupt/unsupported/
unavailable results have no usable Document; existing bytes cannot masquerade as
defaults. They may return known byte/version/recovery observations for diagnosis,
but those tokens cannot authorize rewriting unsupported data. Non-nil errors
preserve any independent observations; permission/sharing failure is not absence.
Reads that cannot establish a coherent bounded byte/version observation refuse
rather than publishing mixed concurrent bytes as a valid snapshot.

Load never migrates on disk. Application prefers a current validated canonical
entry, then considers unambiguous legacy candidates. For owner/name maps it
collects all exact/case-equivalent candidates: conflicting values or uncertain
host association remain diagnosed candidates. A legacy active path migrates only
when fresh Git inventory uniquely identifies a WorktreeID in this LocalCommon
scope. No host invention, arbitrary map winner, clone fan-out or casefold deletion.
Relocation creates a new local scope; old preferences remain available for
explicit recovery/migration. Without a unique mapping, Application uses the
documented Current then deterministic inventory fallback, retaining old intent;
partial/unavailable sources cannot justify deleting it.

## State authority and publication

Application serializes active-context writes for its app session. Activation
validates fresh Git inventory, commits the expected version, and publishes only
known committed intent. NotCommitted leaves prior context intact. A Windows
CommittedDurabilityUncertain effect publishes the actual committed intent with
its notice. Indeterminate blocks the dependent context transition and unsafe
retry until reconciliation establishes facts; it cannot announce a successful
new active context or destroy the original intent. A known committed write
followed by failed refresh remains committed with unavailable projection.

Application owns the resulting Preferences/Launch/config projections and emits
one operation terminal. Persistence allocates no OperationID and emits no UI or
Runtime events. Context cancellation and a disappearing modal do not change a
commit result. No in-memory candidate replaces a previously committed cache
before the native commit point. There is no transaction across config/state/run
files; a workflow with multiple writes reports each effect independently.

## Physical scope and native lock protocol

Explicit `--config`/`--state` is resolved once against startup cwd to the selected
physical parent/fixed basename, including an explicitly selected external or
link target. It is not subject to project ancestry. Later substituted links are
never followed. Defaults retain platform locations: config in UserConfigDir
`gh-tree/config.json`; state there on Windows/macOS, otherwise absolute valid
XDG_STATE_HOME `gh-tree/state.json` or home `.local/state/gh-tree/state.json`.
Evaluate defaults lazily after help/version/overrides; no invalid relative XDG
root or subsequent active-worktree change reinterprets them. Create needed
application-owned components relative to the nearest validated existing parent,
checking each created/opened object without following substitutions.

Project run.json is exactly one literal `.gh-tree` child below the Git-issued
selected physical worktree root, then fixed `run.json`. Acquire/compare supplied
root identity; open/create `.gh-tree` relative to retained root, reject every
link/reparse/non-directory substitution, and bind its identity. All document,
lock/recovery/payload names are fixed or generated single basenames relative to
that retained parent. No caller-supplied project path traversal or ordinary
path-join/Open fallback. Reject link/reparse/nonregular document or lock.

Unix uses no-follow directory Openat/Fstat/Fstatat and retained descriptors;
subsequent Renameat/Linkat are directory-relative. This authorizes the originally
opened directory object, even if an uncooperative peer later moves it; it never
follows the substituted old pathname. Observed drift refuses, but continuous
current-path ancestry is not promised. Windows uses NtCreateFile RootDirectory,
OBJ_DONT_REPARSE/FILE_OPEN_REPARSE_POINT plus explicit reparse attribute checks;
actual directory list/data-read guards use READ|WRITE sharing without DELETE.
Metadata-only guards are not an interlock. Parent conversion must safely refuse
handle-relative operations rather than redirect outside. All handles are
noninheritable and joined/closed by the request owner.

Unsupported special/device/network/cloud/FUSE filesystems do not receive local
guarantees through their pathname syntax. Writes refuse UnsupportedProfile until
their native semantics/metadata are proven; no automatic alternate primitive.
Privileged mount/volume changes and hardware power loss are outside the stated
object-scope/local-filesystem guarantee.

Each store has one permanent never-unlinked `<basename>.lock` object. Windows
holds a nonblocking LockFileEx exclusive byte-0 length-1 lock; Unix combines a
keyed in-process mutex with flock(LOCK_EX|LOCK_NB). Bounded cancellable backoff
initially allows at most 5s acquisition; timeout is Busy/NotCommitted. Kernel
handle/process lifetime releases the lock. PID/age/mtime is never authority to
unlink, replace or steal it. Reject changed lock identity/link/nonregular object.
All cooperating instances and processes use that same lock and whole-document
comparison. A retained request lock cannot span UI confirmation.

## Preparation, publication and recovery

Under the stable lock, load/validate the complete current document and compare
Expected, including unknown fields and absence. Prepare a completely encoded
schema1 payload and exclusive cryptographic-nonce manifest, raw-byte original
backup and payload in the same bound directory. Check every write length/error,
supported metadata, flush and precommit close. Keep Windows payload handle alive
through publication. No destination truncation and no predictable temp reuse.

For a present original, retain a no-replace hardlink to that exact observed object
and verify identity. Unix uses parent-relative Linkat; Windows uses native class11
FileLinkInformation on the actual old handle, not the x/sys class72 constant
with a class11 layout. The immutable raw original-byte backup is separate: late
writes through an editor's old handle can change the retained inode. Both remain
recoverable. If retention cannot be established, refuse before publication.

The manifest records schema, operation nonce, family/store/parent/original/payload
identities, expected/proposed versions and generated names. It is flushed before
publication, as are required Unix directory entries. It contains facts, never
executable recovery instructions. Immediately before commit recompare target
identity/content/security under the lock and refuse detected external changes.
Then use only the selected native commit point:

| Platform/profile | Publication and permission contract |
|---|---|
| Windows local NTFS | Retained payload handle, NtSetInformationFile class65 FileRenameInformationEx, retained RootDirectory and one UTF-16 target basename. Expected-present flags REPLACE_IF_EXISTS\|POSIX; expected-absent flags0 never overwrites a competing creator. Require reviewed architecture-correct buffer offsets/layout and supported native behavior; no ReplaceFileW/path-reopen/fallback. |
| Unix supported local filesystem | Prepared same-directory Renameat for expected presence; no-replace Linkat publication (then removal of only owned temp name), or a separately proved native no-replace equivalent, for expected absence. Fchown before Fchmod, file flush before publication and parent-directory fsync afterward. |

Windows preserves/verifies owner/group, ordered DACL/protection/inheritance and
supported label/resource attributes using actual handle GetSecurityInfo/
SetSecurityInfo; inspect CAP and refuse unsupported nonempty CAP/access-affecting
metadata rather than drop it. Existing read-only destination refuses without
clearing attributes or IGNORE_READONLY. New per-user files have protected
user-only ACL; new project files inherit supported project access. Go mode0600
is not a Windows ACL. Unsupported audit/SACL/security-policy, EFS/compression/
alternate-stream or special profiles require explicit refusal/profile evidence;
the retained original is not a claim the new payload copied unsupported metadata.
Native readers use READ|WRITE|DELETE sharing or return genuine sharing errors.

Unix preserves/verifies uid/gid/mode and positively inspects ACL/xattr profile.
Linux Flistxattr/Fgetxattr/Fsetxattr, Darwin Flistxattr/native extended-security
queries and FreeBSD Extattr*Fd/native ACL queries implement the selected detection
paths. Copy/verify supported metadata or refuse unchanged; mode bits alone cannot
prove ACL absence. Native platform evidence is required for each claimed profile.

Native publication return is the commit point. Pre-invocation cancel/failure is
NotCommitted, though exactly owned preparation artifacts may remain. Successful
publication is a known committed effect despite later cancellation, flush, close
or observation failure. Unix file and directory fsync success yields Committed
under its supported crash contract; Windows success yields
CommittedDurabilityUncertain because flushed payload does not prove namespace
power-loss durability. Lost native result/uncertain attribution is Indeterminate.
Current bytes observed after restart cannot prove historical causality.

On restart, reacquire the permanent lock and inspect exact manifest/object/current
version facts. Never replay, overwrite target, adopt another writer's artifact,
or age-purge recovery. Keep raw original/migration backups and retained originals
until conscious safe cleanup. Owned abandoned payload cleanup requires proven
exact identity and lack of references. Initial per-store admission is bounded
to 256 retained operation records or 1GiB retained recovery bytes (whichever is
reached first); count by owned manifests/actual bytes under the lock. Refuse new
commits before preparation when capacity would exceed the configured bound;
never evict originals to fit. Limits are construction settings and test-injected,
not a new automatic cleanup feature.

## Commit results, errors and cancellation

```go
type StorageCommitResult struct {
    Outcome           StorageCommitOutcome
    ProposedVersion   Optional[StorageVersion]
    CurrentVersion    Optional[StorageVersion]
    PublicationKnown  bool
    Durability        StorageDurability
    CancellationAsked bool
    Effects           EffectReport
    Recovery          []StorageRecovery
    Diagnostics       []Diagnostic
}
type StorageRecovery struct {
    Record   RecoveryRecord
    Family   StorageFamily
    Locator  string
    Kind     StorageRecoveryKind
    Identity SourceVersion
}
```

StorageCommitOutcome is NotCommitted, Committed,
CommittedDurabilityUncertain or Indeterminate. StorageDurability is NotApplicable,
SupportedCrashBarrierComplete or Uncertain. StorageFamily is UserConfig,
Preferences or RunConfig. RecoveryKind is Manifest, RawOriginal, RetainedOriginal
or RetainedPayload. Locator is a concrete safe recovery location and identity is
the recorded observation token; neither authorizes automatic deletion/execution.
Record association to the bound store is part of the opaque token/manifest.

Record is the shared B5 RecoveryRecord, not a second recovery authority. Persistence
issues its opaque RecoveryID once for each retained artifact, records it in the
owned recovery manifest before reporting it, and retains the same ID when loading
that manifest after restart or returning the artifact again. Distinct artifacts
(including a manifest and its payload) have distinct IDs; a locator or mutable
SourceVersion alone cannot supply identity. Record's responsible layer is
Persistence, kind/locator agree with Kind/Locator, and exact subject identifies
the physically bound store and Family plus WorktreeID for RunConfig. Its original/
proposed document versions use StorageVersion, while Identity remains the artifact
observation SourceVersion. No manufactured repository ID is used for a per-user
store. All fields and family detail are copied immutable values.

FacetEffect.RecoveryIDs refer to Recovery[i].Record.RecoveryID in this result.
Application's lossless normalization preserves both the shared Record and its
StorageRecovery family detail, including Family, Kind, Locator and Identity; the
operation recovery union deduplicates only equal IDs with consistent records.
A nonnil error, failed refresh, cancellation or aggregate shutdown cannot discard
these records or mint replacement IDs. V-PER-02 and V-APP-01/04/06 must cover
committed-with-error and indeterminate outcomes, multiple artifacts, repeated
manifest observations and propagation through operation terminal/shutdown results.

ProposedVersion identifies the intended fully encoded document when known;
CurrentVersion is only the independently observed current version. The two can
differ after a concurrent external edit. PublicationKnown means this operation's
native success was observed, not merely that later bytes equal its proposal.
Outcome and effect facets are authoritative even when Go error is non-nil.
Diagnostics identify stage and native cause without configuration/credential
dumps. No raw localized error text manufactures commit status.

| Established fact | Required result |
|---|---|
| Invalid/stale scope/version, corrupt/unsupported document, lock/preparation/metadata/flush error or cancel before native call | NotCommitted, PublicationKnown=false, storage facet VerifiedNoTargetChange or NotStarted as actually established; exact owned artifact cleanup diagnostics retained. |
| Native publication success plus required Unix file/directory barriers | Committed, PublicationKnown=true, AppliedVerified storage facet; postcommit diagnostics remain separate. |
| Windows native success, or known native success with a failed/unproved postcommit durability barrier | CommittedDurabilityUncertain, PublicationKnown=true, AppliedVerified storage effect plus precise durability/close/observation notice. |
| Publication may have run but return was lost, or attribution cannot be established | Indeterminate with actual current/recovery observations, never fabricated rollback or automatic retry. |

Shared ErrorCode covers Invalid, Busy, Canceled, Conflict, Permission, Unavailable,
Unsupported, IOFailure, CleanupIncomplete and Indeterminate. Corrupt/
UnsupportedVersion are explicit StorageLoadState/diagnostic detail, distinct from
NotFound. After a known committed effect, a postcommit error does not relabel it
NotCommitted. CleanupIncomplete can coexist with committed storage because an
unclosed handle/preparation artifact is separate from target bytes.

Context bounds lock waits/read/precommit work. Already canceled requests do not
publish. At the native-call boundary, cancellation is separately recorded and
the actual result is reconciled using the request owner's bounded cleanup context;
the caller cannot abandon handle/lock/manifest ownership. Cancellation after
publication is never rollback. There are no unbounded background goroutines or
deferred shutdown writes. Application drains all accepted operation results and
reports unresolved storage/recovery notices in aggregate shutdown.

## Concurrency limits and forbidden behavior

Cooperating writers serialize and compare complete document versions, preventing
stale gh-tree snapshots overwriting newer snapshots. A complete old/new visibility
guarantee concerns fresh successful opens; existing handles may observe the old
object, and genuine sharing/unavailability errors remain possible. Multiple
stores do not form one transaction or simultaneously switch every reader.

An arbitrary external editor can write/replace after the final comparison. This
is explicitly outside cooperative CAS; retained original means the observed
object, not every unobserved replaced object. Never use a postcheck to prove no
intervening effect. Detected conflicts preserve recovery; uncertain attribution
is reported. The stronger Git worktree capture/no-replace publisher is a separate
boundary and is not silently imported as this store's guarantee.

No direct Git/GitHub/Runtime/Discovery/TUI/Composition calls, workflow callbacks,
provider start, active-worktree choice, path casefold identity, TrimSpace
normalization of scripts/aliases/paths, corrupt-to-default write, unknown-field
discard, stale broad merge, in-place truncate, unsafe native fallback, lock-file
lease deletion, automatic manifest replay or retained-original purge is allowed.
No schema write on Load/help/version, JSON tags in Domain, mutable document sharing
or successful-save notice before the selected native commit point.

## Verification and change control

Exact implementation/integration SHA evidence must satisfy
[Verification--001](../Design/CR-%2321/Verification--001.md). Accepted mechanism
probes are design evidence, not future product proof.

| Contract clauses | Mandatory proof |
|---|---|
| Typed family/codec/schema/defaults/unknowns/migration | V-PER-01/03, V-APP-06, V-LCH-03; exercise invalid UTF-8, duplicate/deep/oversized JSON and every nested retained member. |
| Whole-version lock/preparation/publication/permissions/recovery | V-PER-02; real multiple processes/store instances, retained hardlink late writes, expected-absence competitor, native readers, exact stage faults and crash recovery. |
| Bound scope/no-follow/path/defaults/overrides | V-PER-03/04, V-COMP-01; Windows conversion/ancestor replacement, Unix moved original versus substituted path, explicit external/link target, missing/relative XDG and default-free help/version. |
| Alias/default/source/version coordination | V-PER-04, V-LCH-02/03, V-APP-03/06; exact alias conflict, first save, command override and ordered Make intent. |
| Commit-before-publish/cancel/effects/immutable concurrency | V-APP-01/02/03/04/06, V-PER-02/04; no active context change on NotCommitted, committed durability notice, indeterminate reconciliation before dependent retry. |
| Full retained vertical paths | V-E2E-01/04/05/09/10/12/13, V-COMP-02/03/04, V-REL-04. |

Native Windows/Linux/macOS tests plus FreeBSD metadata/mechanism proof for claimed
profiles are required, with architecture-sensitive layout assertions/native
evidence and all twelve cross-builds separately. Inject lock, original-retained,
payload-complete, manifest-flush, before/issued/known native publication,
directory-flush, close and outcome-delivery boundaries. Cover supported ACL/label
copy and unsupported-profile refusal, resource leaks, actual visibility and the
documented external-editor gap. Never claim power-cut durability from mode/race
tests, successful Rename alone or compile-only evidence.

Change history: 2026-09-06, initial DRAFT 1.0.0 under #55; no superseding contract.
Freeze requires fresh independent review of the complete seven-contract/shared-
annex set at exact HEAD, correction/re-review, explicit metadata freeze review and
configured green CI. Substantive accepted-design conflicts return through
design/BC-CHANGE. After freeze only Master-coordinated reviewed/refrozen changes
may alter this boundary, with affected workers notified and verification rerun.
