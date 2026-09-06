# REFDES--001 — gh-tree v0.4 architecture and migration

State: DRAFT — not implementation authority
Parent program: #21; design authority: #52
Product baseline: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b
Design base: 58f1cb9eda941db0941cbb8e04e6a0559a3ca896
Role: Master / Refactor Designer; fresh independent review required
Date: 2026-09-06

## Authority and acceptance boundary

This design consumes APS--001, BR--001 and all ten accepted focused LR--001
reports. Master reopened them for this design, including the corrected Domain
unborn-HEAD decision and the Git report's APP29-H03/H06 reference erratum.
The complete user authority is preserved in
../../Sprints/Sprint-004-v04/UserRunContract.md. That contract remains the scope
and safety ceiling. Review acceptance has resolved no product finding.

The design consists of this decision record and the normative API, Storage,
CwdAcquisition, migration, Slice, verification and finding-map appendices. Feasibility appendices record
source/probe evidence and its limits; they are not production implementations.
No worker may implement from a DRAFT design or freeze a BC from unaccepted inputs.
Acceptance requires all appendices complete, one frozen exact HEAD, independent
review/correction, green CI and a merged design-only PR. Required BCs follow in
a separate reviewed/frozen gate.

## Release line and complete scope

Use v0.4 as the architecture line and main as its final release target. Preserve
v0.3.14 and its history as the bounded-stabilization baseline; do not deliver an
intermediate architecture as a v0.3 patch. Canonical product integration branch
is the user-requested codereview-21/refactor, created only after the BC gate.

Select the complete Slice set in Slices--001.md. Preserve every existing capability
in the user contract: remote namespace navigation, branches/commits, worktrees,
exact deployment, dirty safeguards, stash, staging/commit/fetch/pull/push/PR,
graph/diff, saved/discovered launch, F5/multi-console, interactive PTY/ConPTY,
mnemonics/focus and persisted preferences. Correctness refusals for unsupported
or changed state are explicit results; no silent target substitution or feature
deletion is permitted. No unrelated fork management, full Make interpreter,
general terminal emulator or theme editor is added.

## Final physical layers

| Layer | Production owner | Exclusive responsibility |
|---|---|---|
| Domain | internal/domain | Small pure scoped identities, exact OID/Revision, HEAD variants and typed exact targets. |
| Application | internal/application/{api,ports,usecases} and coordinator | Operations, confirmations, active context, vertical workflows and cross-source projections. |
| Git | internal/git | Local Git facts/mechanics, native locking, exact mutations and private recovery protocols. |
| GitHub | internal/github/adapter | Explicitly scoped remote observations and PR creation through an authenticated transport. Legacy parent files coexist unchanged until cutover. |
| Runtime | internal/runtime | One session registry, native process/PTY resources, output and complete cleanup barriers. |
| Launch Discovery | internal/launchdiscovery | Passive bounded provider observation, immutable selection resolution and cwd validation. |
| Persistence | internal/persistence | Settings/state codecs, scoped versions, migration, storage locking and commit outcomes. |
| TUI State | internal/tuistate, including viewmodel leaf | One deterministic interaction reducer, selection/modal/focus and projection correlation. |
| TUI View | internal/tuiview | Pure cell/layout/text/graph/output rendering and measurement from supplied snapshots. |
| Composition | cmd/gh-tree, internal/composition/{host}, internal/version | CLI, constructor wiring, restricted Tea host and root lifecycle/build identity. |

Legacy app/worktree/graph/diff/launch/terminal/process/tree/tui/graphui packages
are retired after their mapped replacements and parity gates. Existing github
stays a horizontal folder. Its new final adapter subpackage never imports the
legacy parent client; MC alone retires the parent files after parity. New code
must not call old implementations as its hidden backend.
MigrationMap.yaml names every old physical file, its sole edit/removal owner,
replacement responsibility and gate. A new source file has exactly one layer.

## Dependency rules

~~~
cmd/gh-tree -> composition + version
composition -> application + concrete adapters + composition/host + version
composition/host -> application/api + tuistate + tuiview + tuistate/viewmodel
application -> api + ports + usecases + domain
application/usecases -> api + ports + domain
application/ports -> api contract values + domain
application/api -> domain + standard value/context types
git/github/runtime/launchdiscovery/persistence -> ports + api contract values + domain
tuistate -> application/api + domain + tuistate/viewmodel
tuiview -> tuistate/viewmodel + approved rendering libraries
tuistate/viewmodel -> domain + pure standard value types
domain -> pure standard library only
~~~

Usecases never import their parent coordinator. API never imports ports. Viewmodel
never imports parent State, View or Application. State never imports View; the
host calls both. Concrete adapters never import each other. Adapter-private
short-command helpers may duplicate a small transport primitive; there is no
new shared process/types/filesystem layer. View/State/Domain may not perform OS,
network, filesystem, process or clock I/O. Host performs no Git/launch/storage
workflow and imports no concrete adapter.

CI checks module-qualified direct imports and selected production source for
all twelve target combinations, external-test imports separately, public API
types and final package inventory. During migration only the exact unchanged
legacy paths in MigrationMap may remain. The final allowlist is empty. A versioned
wrapper renamed into a new directory fails the ownership gate even if imports pass.

## Identity and observation decisions

RepositoryID is an opaque comparable token, not a universal assertion that a
remote repository and a local clone are the same object. Two namespaces are
minted at outer boundaries: a canonical host/owner/name remote scope, and a
canonical Git common-directory local scope. Application associates these
explicitly from observed remote bindings. Linked worktrees share local scope;
independent clones do not. Domain parses neither URLs nor filesystem paths.

For this release local scope is derived from the physically resolved common
directory and platform-correct canonical locator. A relocation creates a new
storage scope; old preferences are preserved, not guessed into another clone.
Remote owner/name shorthand resolves its host once; remote rename/transfer is
an explicit new locator association, not blanket casefold migration. Canonical
normalization is documented in API--001 and tested; case-sensitive local paths
must not collapse. Metadata paths are not authorization for Git mutation.

WorktreeID combines local RepositoryID with a Git-issued administrative key;
retargeting does not change it. PRID is scoped to its base remote repository;
head/base branches carry their own remote repository scopes. Local and remote
BranchIDs remain distinct even when their names match. StashID uses local common
RepositoryID plus full stash OID. LaunchPointID uses WorktreeID plus a length-
delimited provider/project/member key, independent of display path and version.
SessionID is a centrally allocated nonzero uint64 value, never reused or decoded
by kind. Runtime rejects allocation exhaustion.

OID supports exactly 40/64 hexadecimal digits, canonical lowercase, rejects
whitespace/abbreviations/nonhex/all-zero and records object format. Revision is
an exact commit identity; Git verifies commit type/existence. A pure three-variant
Head is selected: Attached(BranchID,Revision), Detached(Revision), Unborn(BranchID).
Unborn has no Revision; exact-revision operations reject absent commit identity.
Invalid/unresolvable HEAD is an adapter error, not another disguised branch.

The hard Ref/Revision rule is retained: a selected exact Revision is authoritative.
Mutable names are locators only. Fetch/resolve verifies ExpectedRevision before
target mutation; mutations use the pinned full OID and verify promised post-state.
Freshness never replaces expected-OID comparison. Explicit latest-ref queries
cannot be passed as exact mutation targets.

Git fetch generation, remote observation identity/interval, Application operation
ID, query generation, TUI intent generation and storage/source versions are
different values. Remote-tracking refs are cached local facts. Upstream is
None/Resolved/Gone/Unresolved/NotApplicable; ahead/behind are optional and valid
only for explicit resolved endpoints. Limited remote lists cannot prove absence.

## Application operation and event model

The coordinator validates submissions, allocates monotonically increasing
OperationID, records immutable request/correlation, then starts a use case.
Accepted, progress, confirmation and exactly one terminal event have ordered
operation identity. Invalid synchronous submissions are rejected without claiming
acceptance. API--001 fixes envelopes, typed payloads and error/effect values.

Use one owned event stream with monotonic sequence and immutable snapshots.
Critical acceptance/terminal outcomes are retained until consumed; output hints
may coalesce by offset. Admission is bounded before acceptance, not by dropping
an accepted outcome. Host disposal does not strand workers behind an undrained
event channel: Application shutdown keeps drainage/history alive through barriers.
Every goroutine/subscription has a join or explicit residual outcome.

Queries are safely supersedable within an explicit scope. Mutations serialize
per local common repository; active-context changes serialize per application
session; lifecycle changes serialize per Runtime SessionID. No UI navigation
automatically cancels or forgets a mutation. Cancellation intent and actual
effect are separate. A completed mutation followed by failed refresh remains a
known mutation with unavailable projection, never an unchanged failure.

Confirmation belongs to the original operation. A single-use ConfirmationID
binds exact target, worktree, allowed choice and the Git/storage plan displayed.
Do not hold native locks while waiting for the user. Confirm consumes the token
under coordinator synchronization, reacquires adapter guards and revalidates.
Changed preconditions end with ConfirmationStale and no destructive continuation;
a new request is required. Cancel terminates a waiting operation. A completion
cannot close a newer modal.

Active worktree is Application context. Activation validates current Git inventory,
writes the versioned durable preference, then publishes durable success. A failed
precommit write leaves context unchanged. Committed-but-durability-uncertain state
is published with that precise notice; indeterminate state is reconciled before
another dependent activation. Startup may use a valid saved target, otherwise
Current then a deterministic inventory fallback, without erasing unavailable old
intent merely because a partial observation failed.

Repository overview and graph allow independently valid local/remote sources
with explicit diagnostics/completeness. Application performs branch/upstream/fork,
worktree/PR and graph annotation joins. View gets semantic relations, not raw
SHA fields from which to infer business truth. Bound PR/branch retrieval to a
documented finite page set; report More/Unknown at the cap and retain requested
namespace intent when completeness cannot justify removing it.

## Git safety and mutation mechanics

Feasibility/Git-Safety.md supplies the native API evidence. Git owns operation-
specific plans and serializes cooperating mutations by local common repository;
native Git locks remain required. Plans include exact HEAD/mode/target, index,
paths, content/identity, availability/occupancy and scoped ref observations.
Return per-facet effects: NotStarted, VerifiedNoTargetChange, AppliedVerified,
Partial or Indeterminate, plus cancellation intent, post-observations and recovery.
Do not infer rollback from a generic process error or a matching HEAD alone.

Exact stash apply uses full OID and restores the index as in the baseline. Pop is
apply, verify, then exact drop; a failed/refused drop returns AppliedStashRetained
and never invites automatic apply replay. Native positional drop is forbidden.
For recognized files ref storage, implement only the narrow journaled refs/stash
deletion protocol: lock before reading the reflog, select exact OID/unique record,
retain recovery refs/bytes, publish survivor log/ref with native chaining, and
verify survivors. Empty-ref cleanup uses native prepared expected-OID deletion
and checks the log remains empty while locked, protecting reinsertion/packed refs.
Unknown/reftable/symbolic/malformed storage receives an explicit safe-deletion
unsupported result. No generic hand-written ref database is authorized.

Tracked restore is from the confirmed index version, not HEAD. Refuse untracked,
conflicted or unsupported special paths. Prepare converted replacement through
native Git in scratch storage, bind conversion inputs, atomically capture the
actual original to an owned same-filesystem recovery location, inspect that
captured object, and install without replacing a newly created destination.
Mismatch preserves both the latest original and any new destination. Retain the
captured object for conscious recovery cleanup so an existing open editor handle
cannot lose later writes. No copy-then-delete fallback or automatic recovery purge.

Checkout pins selected and departing history, uses full OID and native scratch
--no-overwrite-ignore, protects ignored/untracked collisions and primary/current
restrictions, and never force-rewinds user branches. Attached advancement is
fast-forward only with native prepared expected-old ref guards and exact post-
state; branch occupancy is checked and ordinary native guards retained. Pull is
fetch plus exact observed-upstream fast-forward; push sends an explicit pinned
OID to an explicit remote ref without force. Upstream configuration is a separate
checked effect after successful push.

Select the superseding protected-publication addendum in Feasibility/Git-Safety.md.
Native file-changing commands execute only in a private standalone scratch repo
with its own HEAD/refs/index/object writes and copied inputs. A linked scratch
worktree or alternate index alone is insufficient for stash/ref isolation. Bind
effective attributes/EOL/ignore/conversion settings and index semantics; never
hardlink live inputs into scratch or silently drop unknown flags/extensions.
Ordinary full indexes/files and supported symlinks/converters are required;
unproved sparse/split/skip-worktree/intent-to-add/gitlink/sequencer/special-file
states refuse before live publication with explicit UnsupportedMutationState.

Prepare an immutable changed-path manifest; transfer native objects and preserve
all roots, including conflicted index stages, then perform separately reported
preparatory stash-store/new-branch effects. Acquire real native index/ref guards,
revalidate, and journal before each path transition. Atomically capture actual
expected-present entries, validate those captured objects, and publish retained
payload objects through no-replace creation. New/unrelated untracked entries are
never swept into a cleanup set. Empty-directory removal uses rmdir, never recursive
deletion. Capture and preserve the real old index and publish a complete understood
native-generated index image without replacing a newly created index destination.
Normalize alternate index stat information against live files through read-only
worktree inspection; scratch cache fields do not prove the live files clean.

Retain both original and installed file objects, including late writes through
held-open handles. Recovery/retry must capture the currently named object anew;
no ownership marker or old hash authorizes truncating it. Prefer recovery below
the actual Git administrative directory when it shares the affected filesystem;
otherwise create an exclusive operation directory below a validated sibling
.gh-tree-recovery directory on that filesystem. Bind its repository/owner marker,
parent identity and fresh cryptographic nonce; never adopt/delete another owner's
directory. Refuse when a safe same-filesystem location or primitive is unavailable.
Return explicit recovery locations at completion. No automatic retention purge;
space limits refuse further publication rather than evict originals.

Commit native ref/HEAD transactions after protected file/index publication. Native
reference hooks and applicable post-checkout/post-merge hooks run once in the real
operation context; no cleanup/replacement follows them. Scratch lifecycle hooks
are disabled to prevent premature/double callbacks. Reconcile hook/concurrent
edits and report known changed/conflicted/partial/indeterminate facets. Scratch
stash conflicts may publish guarded marker files and unmerged index as
AppliedWithConflicts; a generic unclassified scratch failure never publishes.
StashCreatedCleanupRefused and AppliedStashRetained are explicit distinct outcomes.

Attached/new-branch transitions use native symref transactions, with tested
Git 2.48.1 capability profile. Create a new target branch first, then prepare
no-deref symref-update HEAD with expected old ref/oid and target verification,
retaining native HEAD reflog. In the files backend, the departure branch's OID
uses an exclusive verify-only ref lock because adding its verify to the same
native HEAD transaction conflicts with Git's implicit HEAD log update. This
guard never writes ref bytes. Already-attached fast-forward uses native expected-
old branch update and associated HEAD lock; detached updates use native no-deref
CAS plus required departure guard. No manual HEAD writer or unsafe symbolic-ref
fallback. Older Git refuses only affected capabilities with a stated prerequisite;
read/supported detached operations remain available. Native symref introduction
in 2.46 is source evidence, not a claim of a tested older minimum.

No multi-file atomic visibility or global occupancy transaction is claimed.
Cooperating operations serialize; changed occupancy refuses or reports partial
state without touching another worktree. The no-silent-loss rule is established
by retained actual objects and exclusive publication, not external-writer
quiescence or a narrow pre/post-check window. Eighteen Windows/NTFS mechanism
probes support feasibility; native platform/path/index/hook/crash verification
at actual implementation SHA remains mandatory.

## Runtime resource model

Select the mechanism in Feasibility/Runtime.md: one registry, immutable start
specifications, explicit capabilities and separate phase/exit/cleanup facts.
Windows uses private native x/sys CreateProcess suspended, Job assignment before
ResumeThread, retained thread/process handles and private ConPTY acquisition/
rollback. Do not wrap Start-then-assign or retain the pinned wrapper leaks.
Job kill request is followed by zero active Job membership, root wait, terminal
close, pipe closure and joined readers/writers before terminal success.

Unix uses a dedicated SID for launch and PTY roots, controlling tty for PTY and
whole-session census across ordinary foreground/background job-control groups.
Select Feasibility/Runtime.md addendum RTF-02: a private same-executable supervisor
is the SID leader, and the user root starts in a different group. Supervisor
remains alive after user-root exit. For each census group it spawns a helper with
setpgid before exec; the kernel permits only same-session membership. After exec
the helper validates its own acquired SID/group and private handshake, waits for
authenticated Commit, then signals its own group with kill(0, permittedSignal).
Its membership pins that group even if its original members exit. Numeric
precheck-to-kill is forbidden and has no fallback. Deliberately scoped helper
kill(0) is the only zero-target exception; no arbitrary public signal API exists.

Grace sends TERM through acquired helpers; escalation acquires parked STOP helpers,
re-censuses until ordinary groups cannot create new members, then KILL helpers
terminate their acquired groups including themselves. Expected helper SIGKILL
exit is not terminal cleanup proof. Require no remaining live SID members other
than supervisor and all owned waiters joined, then Quiescent -> parent Release ->
supervisor exit -> parent wait/PTY/pipe/control joins. Keep draining while Release
closes the final slave. A crashed supervisor, unacquirable live group, credential
failure or unexpected member in the reserved supervisor group is residual failure;
never restart or signal remembered numbers as recovery.

Composition has one early Runtime-owned private-entry dispatch before normal
flags/config/bootstrap. Runtime owns the implementation, inherited anonymous
control/reply endpoints, capped versioned nonce/sequence protocol and lifecycle.
No PID/group/signal or bearer secret is accepted through public argv/environment;
no control descriptors reach user commands. Missing/invalid endpoints fail closed.
Cancel before helper Commit sends no signal; cancellation afterward cannot abandon
cleanup. Parent-channel EOF initiates supervisor teardown independently of a dead
UI. RTF-02 establishes core native Linux feasibility as UID65534 and nine Unix
cross-builds; final framed-protocol/fault and native macOS/FreeBSD proof remains
mandatory at implementation acceptance.

Supported ownership is the created OS session;
deliberate daemon/new-session escape or external-service launch is not a portable
sandbox guarantee. Observed escape/uncertainty is residual failure, never stopped
success. Native verification must include root exit before descendants and
ordinary foreground/background/stopped jobs, not only root-PGID cleanup.

Start establishes ownership before success; failed partial acquisition remains
cleanup-pending until actually cleaned. Stop is idempotent and does not abandon
cleanup when a caller stops waiting. Restart waits for the entire old barrier,
then uses a new SessionID with copied prior spec and latest explicit geometry;
failed cleanup prevents replacement. Session-cleaned event and operation terminal
result are distinct: later natural exit never completes Start twice.

Central initial budgets: grace 2s, force/resource join 3s, aggregate shutdown 8s;
tests inject shorter/longer bounds. Unsupported graceful mechanisms skip grace
explicitly (no broadcast console signal). Failure to reach a barrier retains the
session/resources and reports residual failure. No terminal record evicts a live
or cleanup-pending resource. Retain at most 256 cleaned sessions and cap normal
live admission at 64; exhaustion refuses before unowned acquisition.

Output is a bounded 256KiB per-session byte ring with stream kind, absolute offsets,
truncation and per-session sequence. Output hints may coalesce; OS draining cannot
wait on UI delivery. Input uses a bounded copied queue; resize validates positive
supported dimensions and serializes with stopping. ETX is terminal input, not
proof of whole-tree stop. No working/idle inference from provider or process
existence. The presenter interprets safe line output; Runtime owns transport only.

## Launch Discovery and Persistence

Discovery is passive and context-aware, preserves exact script/path bytes, and
returns immutable definitions with identity separate from source fingerprint.
npm scripts remain singular; Make members preserve order and must share project,
manifest and executable policy. Use GNUmakefile/makefile/Makefile precedence for
default Make, bind the selected file explicitly, and reject option/assignment-
shaped operands. Retain colocated pnpm/yarn/npm lock precedence with diagnostics
for conflicts; explicit stored executable overrides remain supported. No ancestor
workspace discovery or build-tool execution is added.

Bound scans to depth 5, 10,000 directories, 10,000 candidates, 4MiB per manifest
and 1MiB Make line; expose incomplete scope/diagnostics. Keep existing generated
directory exclusions. Resolve revalidates worktree/project/member/version before
Runtime Start; changed sources return stale selection, not another invocation.
Project cwd is relative to the selected worktree; physically validate root/path
and reject redirects outside it. Reject unproven child link/reparse paths in the
initial implementation; linked root canonicalization remains the Git boundary.
Runtime follows CwdAcquisition--001: acquire/validate the actual directory object,
then Unix supervisor Fchdir/inherited cwd or Windows effective no-delete-share
component guards through CreateProcess. A pathname recheck followed by an unguarded
reopen is forbidden. Refuse stale/unsupported scope; fixed-object authority is
distinct from continuous pathname ancestry when another actor moves a directory.
There is no filesystem sandbox claim against later arbitrary project code.

Persistence owns user config, navigation/active preferences and worktree run.json
bytes/schema; Discovery interprets provider intent through Application. Retain
known legacy fields, aliases/defaults, command overrides and explicit empty prefix
settings. Preserve unknown JSON fields or refuse unsupported versions, never
rewrite a forward schema as empty defaults. Retain originals on migration and
report ambiguous case/host/clone mappings without choosing a destructive winner.
No automatic corruption repair or global user configuration changes.

Select Storage--001 and Feasibility/Persistence.md: whole-document versions and
stable never-unlinked LockFileEx/flock sibling locks; exclusive flushed same-dir
payload/manifest/raw backup plus retained observed original hardlink. Windows
local NTFS uses handle-relative NtSetInformationFile class65 replacement, class11
hardlinks, supported verified security metadata and no-delete-share directory
guards. Unix uses no-follow directory-relative handles, Renameat or no-replace
Linkat, supported metadata and file/parent fsync. Unsupported semantics refuse;
no truncate, unsafe fallback or automatic original purge. Native publication is
the commit point, and Windows namespace durability remains explicitly uncertain.

Storage--001 fixes version1 schemas, legacy retention/unknown fields, duplicate/
corrupt/forward-version refusal and unambiguous Application-owned migration.
Six native Windows and six unprivileged Linux mechanism cases support feasibility;
full exact-product native metadata/crash/recovery proof remains required. Atomic
visibility, process-crash recovery and power-loss durability are distinct. External
writers that ignore locks remain outside cooperative CAS; detected drift refuses,
and the design explicitly does not claim every unobserved external replacement
is captured. Unix scope is the pinned authorized directory object, not continuous
pathname ancestry after another actor moves it. Context publishes actual commit
outcome; alias/default commit together. Explicit --state/--config locations retain
user-selected scope; project storage never follows a substituted child object.

## TUI State, View and restricted host

Use one reducer with a single focus path, one modal variant, semantic selected
IDs, local scroll/search state, per-projection request generations and one global
monotonic user-intent generation. Navigation/selection/modal changes invalidate
old conditional follow-ups. Matching OperationID alone never grants focus.
Preserve current auto-follow for a newly started console only when its captured
intent still matches. Query results require current slot/source/generation;
session facts use per-session sequence independently of tab selection.

State derives immutable viewmodel snapshots and a semantic applicable keymap.
Both graph entries use the same reducer/renderer and accumulated graph projection.
View owns lane cells/labels and never alters semantic commit messages. View has
no backend handle, effect callback, timer or clock. One pure layout is shared by
rendering and measured console rectangles; host dispatches changed dimensions
through Application with current SessionID/content/viewport correlation.

Known small viewports are real bounds. Wide cockpit preserves existing panes;
compact layouts prioritize the focused pane and its navigation while retaining
explicit access to every other pane. Modal target/choices are reserved and its
body scrolls; all exact confirmation identities remain inspectable. Unknown
initial size alone uses a fallback. Cell/grapheme-aware clipping handles wide
glyphs and invalid input safely. Ordinary metadata/patch/log controls are rendered
as safe text; only renderer-owned ANSI styling reaches the host. Preserve baseline
terminal CR/backspace line interpretation, not an unrequested full VT emulator.

When a viewport cannot show both a target identity and an approval action, use
a bounded resize/read-only notice and retain cancellation/navigation; do not offer
an unseen destructive approval. Returning to adequate size restores the same
modal and exact identity. Theme and color profile are immutable supplied values;
rendering libraries must not perform implicit terminal/environment detection
during construction/rendering or mutate package-global styles.

Host maps Tea input/events, executes declarative effects through API, and owns
one disposable presentation timer plus session-output interpretation caches.
No rendering call executes a command. Application shutdown retains event drainage
through resource barriers even after Tea Run exits. Root passes WithContext,
unwinds partial construction, preserves primary and cleanup errors and returns
nonzero for residual cleanup; only main performs os.Exit.

## Migration and release gate

First reviewed Composition contribution enables codereview-21 CI pushes, expands
all twelve builds, disables automatic bootstrap-branch publication and establishes
staged dependency checks. Then reviewed Domain/API/ports/viewmodel leaves precede
dependent workers. Adapter contributions are disjoint; serial order is Git,
GitHub, Persistence, Launch Discovery, Runtime, followed by Application workflows,
State, View and shared host/cutover. Slices have separately traceable contributions
and end-to-end harness evidence; they are not complete merely because one folder
compiles. Final cutover/legacy deletion requires capability parity and strict imports.

Master coordinates shared paths and integrates only exact reviewed commits.
Fresh workers implement all product changes, including explicitly delegated
Composition/CI/cutover work; separate fresh reviewers inspect exact source/tests.
Do not independently move the same old file in multiple branches. A faulty frozen
BC triggers BC-CHANGE, impact review, refreeze and affected re-verification.

Verification--001.md defines the exact integrated-HEAD contract. Final product PR
is opened only after it passes, including native Windows/Linux/macOS and required
additional platform mechanisms, all twelve cross-builds, real Git repositories,
operation/state/renderer tests, migration and retained capability. Main then needs
its exact release-commit gate. Version belongs in that tested commit.

Build/stage artifacts before publication, assert the complete twelve-name manifest,
tie artifacts to exact source/workflow SHA and reject existing asset conflicts.
Use an already verified tag pointing to the tested release commit; never let an
action invent it from default branch or clobber published assets. Prefer a gated
draft lifecycle and explicit final publish after asset inspection. Controlled
install/upgrade verifies delivered architecture/version and compatible state.
No HIGH/BLOCKER safety finding may remain unresolved at release; no success is
inferred from review acceptance, cross-compilation or an author's summary.

## Required BC inventory

After this design is accepted/merged, create/review/freeze version 1.0.0 of:
BC--TUIState--TUIView, BC--TUIState--Application, BC--Application--Git,
BC--Application--GitHub, BC--Application--Runtime,
BC--Application--LaunchDiscovery, BC--Application--Persistence.
Domain invariants and Composition wiring policy are governed here and in API/CI;
they do not add an adapter layer or giant Composition runtime BC.

## Draft disposition

This record is not accepted. Complete and cross-review all normative appendices,
resolve the two early safety findings with concrete feasible protocols, then
freeze the complete design HEAD for the independent design gate. Feasibility
proposals and a machine-complete inventory do not themselves authorize product work.
