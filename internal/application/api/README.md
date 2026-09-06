# Application semantic values

Issue #59 supplies the complete M2 API leaf under the seven frozen boundaries
and BoundaryTypes--001. Issue #67 applies Application--Git 1.1.0's bounded status
cause correction; the other boundaries remain 1.0.0. It imports the accepted Domain foundation
directly. It performs no orchestration, allocation, observation, clock read,
filesystem access, process execution, or network activity.

Every semantic record has a private representation, `NewName(NameData)` validating
constructor, `Valid`, `Data`, and `Clone`. Its zero is invalid. `NameData` is an
ordinary construction/copy record, never a mutable reference to the stored value.
Constructors and Data copy all slices, including slices inside Optional and
StoredField. Nested records, identities, and strings are already immutable;
Clone can safely share their private storage. No getter exposes owned mutable
storage. Constructors return invalid zero plus error on failure.

Optional explicitly distinguishes absence; it does not promise arbitrary generic
T is valid or immutable. Every actual containing family checks its concrete T and
copies mutable containers. StoredField distinguishes Absent, Null, and Present;
only StripPrefixes accepts Null among specified known document fields. An explicit
empty list stays present. Unknown-provider strings and every known/unknown member
are retained; a storage string is not execution authorization.

All 32 union families are sealed interfaces whose containing constructors use
exact concrete type switches. Nil, typed nil, pointer variants, and foreign
embedding implementations refuse even if they implement Valid. Public Command
has 28 alternatives, Query has 13, Result has their 41 matching alternatives,
and EventPayload has six. Request contains exactly one command or query, with
the proper correlation domain. `Request.AcceptsResult` and `ValidateTerminalFor`
check matching families and original correlation. SessionOutputProjection wraps
the Runtime-owned semantic SessionOutputResult plus captured projection sources
and ContextVersion. FetchResult and CreatePullRequestResult are the same semantic
facts at the port and public terminal; they are not duplicate adapter DTOs.

| Files | Complete family |
|---|---|
| primitives, ids, enums, values_records | Distinct IDs/versions, explicit presence/pages, observations, worktree/remote/cwd binding, invocation/environment, diagnostics, effects and typed recovery versions |
| git_records | Git facts, exact reads/comparisons/stash parents and tree OIDs, all eight preparations, safe summaries, candidate-versus-final-index facts, closed outcomes, fetch and seven-facet reconciliation |
| remote_records | Qualified remote repositories/branches/PR endpoints, unavailable fields, observations/pages, exact creation expectations and six creation outcomes |
| storage_records, json_values, storage_validation | All three whole documents, every specified known/nested/legacy field and unknown member, load/commit outcomes and stable StorageRecovery detail |
| discovery_records | Passive observations, project inputs/profiles, definitions, exact discovered/saved/ordered Make selections and resolved invocation facts |
| runtime_records | Invocation requests, retained failed-start identity, all session phases/capabilities/exit/cleanup/control/output/event/ACK/shutdown values |
| application_records, client, result_binding | Five-method Client, full commands/queries/results/events, canonical relations, active context, pending ACK transfer and aggregate residuals |
| unions, evidence, evidence_records | Closed-family admission and explicit evidence traversal; no reflection or generic command dispatcher |

SourceVersion is opaque comparable namespace/scope/issuer/token equality. API
does not order or interpret it. StorageVersion retains family, store identity,
missing/present state, length, SHA-256 and, for RunConfig, exact WorktreeID/root
identity. `NewRunStorageVersion` and `MatchesRunScope` prevent cross-worktree/root
use without resolving a path. Observer-supplied store identity represents the
native parent or validated missing-ancestry anchor; Persistence must revalidate it.
Different repositories, version families and sequence domains never interchange.

ObservationInterval takes ordered UTC acquisition bounds. Commit/stash author
times, commit committer times, and PR UpdatedAt retain the supplied source property
and offset; they are not observation freshness. Stash/commit parent selections
use zero-based indices into the actual returned parent vector. Root comparisons
are explicit alternatives. Tree objects are OIDs, never fabricated Revisions.

OpaqueJSON uses json.Valid plus pure UTF-8, bounded shape/depth and duplicate-name
predicates; escaped-name equivalence is checked, including surrogate pairs. It
exposes copied bytes and never decodes or encodes a document. JSONMembers retain
exact ordered unique names and values. Whole-document constructors reject known
name collisions, invalid field placement/presence, duplicate semantic map keys
and invalid UTF-8; they apply bounded retention/depth budgets. Persistence owns
the strict complete input/schema codec, exact encoded 4 MiB limit, unknown-member
preservation against guarded prior bytes, and actual load/publication proof.

Structured errors supplement results. Evidence traversal rejects dangling
post-observation/recovery references and inconsistent reused RecoveryIDs.
NormalizeRecovery preserves each shared record plus complete StorageRecovery
family/kind/locator/artifact identity and typed document versions. Result and
terminal recovery unions must retain these facts. LocalConfiguration is an
independent effect; Git Reconcile accepts exactly the seven specified shared
facets and rejects Storage, RuntimeResources and duplicates.

The types enforce structural identity, variant, copy, and representable factual
consistency. They do not establish that an observation occurred, validate an
adapter registry, consume confirmation, prove native cleanup, guarantee network
creation, or implement exactly-once operation/workflow behavior. Environment
keys use exact equality generally and Windows case-insensitive collision checks
when admitted into an invocation for a Windows observed cwd.

Verification: `go test ./internal/application/... -count=1`, the corresponding
race run, full test/vet/build, and M1's architecture checker across all twelve
target selections. Closed-family negatives cover every union; external fakes
compile all five Client and all 48 port methods. These are M2 leaf checks, not
completed Slices, native adapter evidence, or a release claim.

The M259 correction checks relationships across individually valid values:
request/result worktree and exact-target identity; Git read envelopes and successful
postconditions; complete operation-specific plan intent; compound operation/step
identity; full Storage run/root/store binding; Runtime admission and original
restart specification; and every independently known PR endpoint/qualified URL.
Unavailable/partial observations remain representable and retain their facts.

MutationPlanSummary now includes the supplied operation-specific destination,
create/retarget mode, stage action, commit message/index policy, closed stash
intent, branch and push binding where applicable. Validity requires the relevant
material intent and its consistent Expected identities. These fields describe the
frozen plan; they neither prepare native state nor grant approval.

MergeEffectReports forms an ordered union of complete facet facts. Aggregates must
retain every distinct child fact, including observations and recovery identities.
It never ranks states or reduces a facet to a last writer: an earlier applied
index and a later unstarted commit-index step can coexist, while typed children
retain their exact operation/subject/step provenance. Thus a terminal cannot erase
an applied child or replace it with NotStarted/Partial/Indeterminate. Identical
facts may coalesce; differing factual states/stages remain visible.

Each ChangeFact requires a ChangeCause: IndexChangeCause describes HEAD-to-index
(unborn HEAD compares with the empty tree), WorktreeChangeCause describes
index-to-filesystem, and UntrackedChangeCause/ConflictChangeCause retain separate
native causes. Kind and optional OldPath describe that cause, while IndexEntries
and WorktreeState always describe current observed facts. Renamed/Copied require
an exact, different OldPath; ordinary causes permit only stage 0 or no entry,
Untracked permits no entry, and Conflict requires a nonempty subset of stages 1..3.

StatusFacts rejects duplicate (Path, Cause), conflict coexistence at the same exact
Path, and contradictory current facts for that Path. Index stages and semantic
flags compare without order; every file identity/content/mode/link/parent fact
compares exactly. Inputs and returned row order remain intact. Staged deletion
with an untracked replacement and staged rename with a later edit/deletion or
rename remain representable. Deleted does not imply an absent current file, and
a rename destination need not be indexed. No cause is inferred from opaque
versions. Partial/Unknown empty observations retain their completeness and never
become Complete through admission. Native status truth and lossless Application/
State projection remain the Git #61 and M4/M5 verification responsibilities.
