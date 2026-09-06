# BC--Application--Runtime

State: DRAFT

Version: 1.0.0

Parent Issue: #55, under #21; accepted design #52 / PR #54

Applies to: complete v0.4 refactor, SLC-10..12 and their SLC-09/13 prerequisites

Accepted design merge: `4a42222f7bfedc1d80693effbb25a1a82fcff65e`

Supersedes: none; no implementation authority until the whole-set freeze gate.

## Authority and responsibilities

This contract specializes [API A1..A6](../Design/CR-%2321/API--001.md),
[REFDES](../Design/CR-%2321/REFDES--001.md),
[CwdAcquisition--001](../Design/CR-%2321/CwdAcquisition--001.md),
[WindowsBroker--001](../Design/CR-%2321/WindowsBroker--001.md) and the selected
[Runtime RTF-02 addendum](../Design/CR-%2321/Feasibility/Runtime.md#addendum-rtf-02--replace-numeric-unix-signaling-with-session-local-helpers).
Those complete native protocols are normative, including their explicit limits;
the older direct-Windows-start and Unix numeric-signaling proposals are superseded.
[BoundaryTypes--001](BoundaryTypes--001.md) fixes common values, copies and errors.
This contract fixes the Sessions family rather than introducing a broker port.

Application owns operation acceptance, OperationID, cancellation intent, one
terminal outcome per accepted operation, active worktree, launch/default/shell
workflow selection, source validation, event normalization and downstream event
capacity. It obtains Git-issued worktree scope and Discovery's validated immutable
Invocation before Start. Application does not own process resources or another
session registry. Changing active worktree never changes an existing session cwd.

Runtime owns the sole registry for launch and interactive sessions, all SessionID
allocation, exact process waiters, acquired cwd, native containment, PTY/ConPTY,
input/output/control resources, start specifications, per-session synchronization,
cleanup barriers, output retention and reliable source events. Native brokers,
supervisors and helpers are private parts of this ownership. Runtime reports
observable lifecycle/exit/output facts; it does not report shell-idle or actively
working based on a provider, PID, terminal existence or output silence.

## Dependency direction and type ownership

Compile-time dependencies are `application/usecases -> application/ports ->
application/api + domain` and `runtime -> application/ports + application/api +
domain + approved OS/terminal libraries`. Application does not import concrete
Runtime. API imports no ports; Domain remains pure. The Composition root wires
one Sessions implementation; its early private-mode dispatch delegates only to
Runtime and precedes ordinary config/auth/CLI bootstrap. Host/State/View do not
call Sessions or import Runtime.

Domain owns SessionID and WorktreeID. API owns Invocation, CwdObservation,
Geometry, all snapshots, operation-neutral result facts and event values below.
Ports owns the Sessions interface. `ports.RuntimeEventCursor` is an alias of
`api.RuntimeEventSequence`; it is not an OS/provider identifier. Shared opaque
source/storage values follow the annex. No exported signature contains a broker
frame, PID/PGID/SID, Job/HPCON handle, OS descriptor, mutable internal record or
adapter-private implementation DTO.

## Commands / inputs: exact Go surface

These declarations belong to `internal/application/ports`. Every argument and
result is copied deeply at the boundary. Method names and typed payloads are
closed; `Execute(string, any)` and provider-string dispatch are prohibited.

```go
type RuntimeEventCursor = api.RuntimeEventSequence

type Sessions interface {
    Start(context.Context, api.SessionStartRequest) (api.SessionStartResult, error)
    Snapshot(context.Context, domain.SessionID) (api.SessionSnapshot, error)
    List(context.Context, api.SessionFilter) (api.SessionList, error)
    ReadOutput(context.Context, api.SessionOutputRequest) (api.SessionOutputResult, error)
    Write(context.Context, api.SessionWriteRequest) (api.SessionWriteResult, error)
    Resize(context.Context, api.SessionResizeRequest) (api.SessionControlResult, error)
    Interrupt(context.Context, domain.SessionID) (api.SessionControlResult, error)
    Stop(context.Context, api.SessionStopRequest) (api.SessionStopResult, error)
    Restart(context.Context, api.SessionRestartRequest) (api.SessionRestartResult, error)
    NextEvent(context.Context, RuntimeEventCursor) (api.RuntimeEvent, error)
    AckEvents(RuntimeEventCursor) error
    Shutdown(context.Context) api.RuntimeShutdownResult
}
```

The following API records specialize common annex values. `Optional[T]` has an
explicit absent state; absent IDs/exits are never fabricated zero-valued facts.

```go
type SessionStartRequest struct {
    OperationID OperationID
    Invocation  Invocation
}
type SessionStopRequest struct {
    OperationID OperationID
    SessionID   domain.SessionID
}
type SessionRestartRequest struct {
    OperationID OperationID
    SessionID   domain.SessionID
    Geometry    Optional[Geometry]
}
type SessionFilter struct { WorktreeID Optional[domain.WorktreeID] }
type SessionList struct {
    Sessions []SessionSnapshot
    Sequence RuntimeEventSequence
}
type SessionOutputRequest struct {
    SessionID domain.SessionID
    Offset    uint64
    MaxBytes  uint32
}
type SessionWriteRequest struct {
    SessionID domain.SessionID
    Bytes     []byte
}
type SessionResizeRequest struct {
    SessionID domain.SessionID
    Geometry  Geometry
}
```

OperationID is allocated by Application before these lifecycle calls. Runtime
only validates, echoes attribution and deduplicates a lifecycle transition; it
never accepts/completes an Application operation. Invalid/foreign IDs refuse.
Invocation contains the annex's closed ArgvExecution or InteractiveShell intent,
explicit environment policy, terminal mode, positive geometry and CwdObservation.
Runtime owns Auto/Configured shell and main-process ancestry resolution, not App
or the private broker's ancestry. Literal executable/argv/environment
reject embedded NUL; command overrides remain executable intent, not implicit
shell source. The annex's explicitly reviewed .cmd/.bat carrier preserves normal
Windows npm shims; other implicit shell fallback is forbidden. npm/Make literal argv and configured
interactive-shell policy survive. Relative executable/PATH resolution happens
after native cwd acquisition, never relative to the main application's cwd.

Geometry is positive rows/columns, each at most 32767 for the supported terminal
ABI. A nonterminal session refuses Resize rather than silently ignoring it.
ReadOutput requires `1 <= MaxBytes <= 262144`; an offset beyond the current end
is Invalid. Offset at the end returns an empty successful bounded observation.
Write copies one nonempty buffer into a bounded admission queue. Its initial
per-session outstanding-input ceiling is 64KiB (including in-flight bytes); a
single request exceeding it is Invalid and a full queue returns Busy before
enqueue. This is a transport budget, not output or operation-event capacity.

## Results, snapshots and event fields

```go
type SessionStartResult struct {
    Session           Optional[SessionSnapshot]
    Established       bool
    CancellationAsked bool
    Effects           EffectReport
    Diagnostics       []Diagnostic
}
type SessionControlResult struct {
    SessionID         domain.SessionID
    Sequence          SessionSequence
    Delivered         bool
    CancellationAsked bool
    Diagnostics       []Diagnostic
}
type SessionWriteResult struct {
    SessionID         domain.SessionID
    Sequence          SessionSequence
    AcceptedBytes     uint32
    CancellationAsked bool
    Diagnostics       []Diagnostic
}
type SessionStopResult struct {
    Session           SessionSnapshot
    CleanupComplete   bool
    CancellationAsked bool
    Effects           EffectReport
    Diagnostics       []Diagnostic
}
type SessionRestartResult struct {
    Old               SessionStopResult
    Replacement       Optional[SessionStartResult]
    CancellationAsked bool
    Diagnostics       []Diagnostic
}
type RuntimeShutdownResult struct {
    AdmissionClosed bool
    Complete        bool
    Sessions        []SessionStopResult
    Residuals       []RuntimeResidual
    Diagnostics     []Diagnostic
}
type RuntimeResidual struct {
    SessionID Optional[domain.SessionID]
    Stage     RuntimeCleanupStage
    Detail    Diagnostic
}
```

`Session` is absent only if Start refused before registry admission. Once an ID
is reserved, Start returns that session even on a failed acquisition with residual
resources. `Established` means the complete startup ownership/cwd barrier passed;
it can be true even if an immediately exiting child is already observed Cleaned
at return. A nonzero root exit is separate from start and cleanup success.
Stop's `CleanupComplete` is equivalent to its snapshot being Cleaned with no
residual resources. A successful repeated Stop returns that same established
cleanup fact. Restart's absent Replacement means no new admission occurred.
A present failed replacement has its own new ID and cleanup facts; the old
session remains Cleaned. None of these facts is an Application terminal event.

Control `Delivered` means the native resize or capability-defined interrupt
delivery returned success; it never means that the child acted on input or exited.
Write accepts the entire buffer or none. `AcceptedBytes` means copied queue
ownership, not child consumption. A later native partial-write/close failure
becomes a session diagnostic with the accepted/delivered byte counts; it is not
safe to replay the input automatically. Resize records the latest successfully
accepted geometry under session synchronization. Interrupt is terminal ETX
delivery where advertised; it cannot become an arbitrary signal API.

`RuntimeCleanupStage` is a closed enum: Acquisition, ProcessContainment,
CwdAcquisition, UserProcessWait, Descendants, Terminal, Input, Output, Control,
SupervisorOrBroker, OuterContainment, HelperExtraction, EventTransfer. Stages
identify failure location without exposing handles. Runtime Shutdown includes
every known residual, including registry-wide resources without a SessionID.

SessionSnapshot's required fields/semantics are:

| Field | Meaning |
|---|---|
| SessionID, WorktreeID | Runtime-allocated identity and original selected Git worktree; restart never reuses identity. |
| StartOperation, RestartOf | Original Application attribution and optional predecessor SessionID, without operation authority. |
| Display/start specification | Copied safe display, executable/argv/cwd/terminal/geometry facts; no credential/environment-value dump. Full immutable environment remains Runtime-private for restart. |
| Capabilities | Explicit output, input, resize, terminal-ETX, tree-stop and restart support; routing never decodes kind/provider/ID ranges. |
| Phase | Exactly Starting, Running, Stopping, Cleaned or CleanupFailed. Root exit does not itself select Cleaned. |
| Exit | Optional observed exit code/signal and natural/requested/failed-start cause. Unknown exit remains absent. |
| Cleanup | Complete/pending/failed facts and copied RuntimeResidual records, independent of exit. |
| Sequence | Positive monotonic SessionSequence; advances for published session state/output-range facts, never TUI generation. |
| OutputRange | Absolute retained-start/end offsets and truncation facts for the current bounded byte ring. |
| Acquired cwd | Accepted directory observations and observed actual locator/diagnostics; an old replaced pathname is not presented as verified current cwd. |

```go
type RuntimeEvent struct {
    Sequence        RuntimeEventSequence
    SessionSequence SessionSequence
    SessionID       domain.SessionID
    Kind            RuntimeEventKind
    Snapshot        SessionSnapshot
}
type SessionOutputResult struct {
    SessionID     domain.SessionID
    Sequence      SessionSequence
    Chunks        []SessionOutputChunk
    RetainedStart uint64
    End           uint64
    NextOffset    uint64
    Gap           Optional[OutputGap]
    Truncated     bool
}
type SessionOutputChunk struct {
    Stream   OutputStream
    Offset   uint64
    Bytes    []byte
    Sequence SessionSequence
}
type OutputGap struct { From, To uint64 }
```

RuntimeEventKind is closed: StateChanged, OutputAvailable, Cleaned. The first two
are coalescible hints; Cleaned is the one reliable final cleanup event. A
CleanupFailed state with retained resources is a nonfinal StateChanged fact.
Each event binds both sequences to its same snapshot and SessionID. Global
RuntimeEventSequence orders this stream, separately from per-session facts and
Application Sequence. No mutation of a previously returned event is permitted.

OutputStream is Stdout, Stderr or Terminal. A PTY's merged stream is Terminal;
separate pipes preserve within-stream order, while their ring order is observed
read order and does not assert a causal total order between OS writers. The ring
retains at most 256KiB per session, counted across all streams. Offsets count raw
bytes in ring order. Chunks are copied and bounded by requested MaxBytes; no
line-scanner ceiling, ANSI stripping or UI formatting affects capture. If requested
Offset precedes RetainedStart, Gap is precisely `[Offset, RetainedStart)` and
reading continues there. NextOffset advances only through returned bytes and the
reported gap. Output/running hints can coalesce; consumers recover via ReadOutput.

## Admission, state authority and lifetime

Before accepting StartLaunch, OpenTerminal or Restart, Application reserves the
operation's critical envelopes and one potential new-session cleanup slot in its
shared 256-envelope budget. Runtime separately reserves a critical cleanup event
before acquiring any new session. Normal live admission is capped at 64 sessions;
256 reserved/unacknowledged cleanup events is the independent Runtime ceiling.
Exhausted capacity/ID space/shutdown refuses before acquisition with Busy or
Unavailable as appropriate. Cleanup, Cancel and Shutdown bypass normal admission.

SessionID is nonzero monotonically allocated uint64, never reused or wrapped for
the registry lifetime. Allocation/gating/insertion share one synchronization
point. An admitted Starting record owns every partial acquisition immediately.
No product code executes before established containment and actual-cwd barriers.
Failure enters Stopping/CleanupFailed until owned cleanup completes; failure of
the requesting operation cannot remove the record or its reserved event.

A Start rejected before registry admission returns no SessionID/event and releases
the unused downstream slot only after proof of no resource/event. An admitted
failed Start remains attributable by ID and produces its single final Cleaned
event if/when cleanup completes. Application retains/transfers that slot like a
successful Start. There is never an orphan whose only identity lived in an error
string. Natural exit starts cleanup; it does not finish the Start operation again.

Stop latches intent once and coalesces concurrent callers onto the same cleanup
attempt. Starting sessions participate. Each Application caller receives its own
call result and later exactly one Application terminal; Runtime produces only one
final Cleaned event. CleanupFailed retains live/pending resources and capacity,
may be safely retried, and cannot be evicted. Later completion does not send a
second terminal result for an earlier failed Stop operation.

Restart serializes on the old ID and deduplicates the Application operation key.
One old-session replacement transition may win. A duplicate key returns/joins
that transition; a different concurrent key observes its result or receives
Conflict, never creates another replacement from the same old ID. Reuse of a
key with different immutable input is Invalid. A deliberate subsequent restart
targets the returned replacement ID. Old cleanup must finish through every
barrier before new ID allocation/acquisition. The replacement uses the original
copied start spec and latest accepted geometry, overridden by explicit valid
restart geometry. Failed old cleanup forbids replacement. Cancellation after old
cleanup but before replacement creates nothing. A new failed admission is a
distinct failed replacement result; there is no overlapping old/new lifetime.

Registry locking protects membership/admission only and is never held over OS
waits. Per-session serialization orders Stop/Restart/Write/Resize/Interrupt and
natural exit. Reads capture coherent immutable snapshots; List orders IDs
ascending and captures one registry observation without pretending all OS facts
were sampled atomically. Cleaned history retains the latest 256 snapshots;
foreign/evicted IDs return NotFound. Evicting a cleaned snapshot never evicts its
unacknowledged final event. Live/cleanup-pending records cannot be history-evicted.

## Ordered event delivery and acknowledgment

There is exactly one Application consumer. Cursor zero starts the stream;
NextEvent(ctx, after) returns the next available event after that successfully
delivered cursor. Runtime retains reliable unacknowledged Cleaned events for
replay independently of snapshot history. A canceled call that did not return an
event changes no delivery ownership. A successful return remains valid if its
context concurrently cancels. An explicit prior delivered/unacknowledged cursor
may replay retained events; unknown/future/ACK-regressing cursor misuse refuses.
Coalesced hints permit sequence gaps and do not promise replay of every hint.

AckEvents is cumulative, monotonic and idempotent for the already-acknowledged
cursor; it cannot advance beyond successfully delivered sequence. It releases
only events through that cursor. Application must first transfer each final
cleanup event through the cursor into the already reserved downstream session
slot and record `(SessionID, SessionSequence)` as transferred. Only then may it
ACK. Transfer is complete when Application owns reliable normalized
SessionChanged delivery, not when the host has seen it.

Replay of a recorded transfer retries ACK without emitting a second normalized
terminal SessionChanged. Pending-ACK records are bounded by Runtime reservations
and survive bridge cancellation until acknowledged or explicitly included as
shutdown residuals. Transfer failure leaves Runtime unacknowledged; neither side
can drop, history-evict or put it into an unbounded side queue. Application releases
its slot only after successful Next delivery or explicit shutdown drain. It keeps
draining when the host detaches, including all active operations and sessions.
Output delivery and a full UI queue cannot block native output draining or Stop.

Shutdown closes Runtime admission atomically, includes in-flight Start/Restart,
attempts every session under one overall budget and aggregates all failures. It
is idempotent; a timed-out wait does not discard cleanup ownership. Initial
budgets are 2s grace, 3s force/resource join and 8s aggregate shutdown, injected
through construction for tests. Unsupported grace is explicitly skipped. These
are deadlines for reporting, not claims that arbitrary kernel failures can be
cleaned on time. A retained event stream stays drainable through resource
shutdown until final outcomes are ACKed; NextEvent returns EOF only after
admission is closed, producers have joined and pending reliable events are
acknowledged. Residual resources/events prevent a false successful shutdown.

## Acquired cwd and native ownership barriers

Runtime independently opens the root/project against CwdObservation, refuses
stale/foreign/unsupported identities, redirects, `..` or unproved components,
and rechecks identities after acquisition. Tokens carry no OS capability.
Checking a path then reopening it is forbidden. The acquired object is authority;
observed relocation before launch refuses. Unix cannot promise continuous path
ancestry if a peer moves the same authorized object. Neither platform promises
a sandbox against later project code, chdir, privileged namespace mutation or
absolute paths synthesized from cached cwd text.

Unix uses retained no-follow Openat directory descriptors and Fstat/Fstatat checks.
The private SID supervisor receives a designated cwd descriptor and authenticated
nonce/role/SessionID/invocation handshake, Fstat-validates and Fchdir's exactly
once, then closes the explicit descriptor. It constructs/resolves the user
command afterward with empty Cmd.Dir, inheriting the acquired object. Main never
changes cwd; no `/proc/self/fd` or `/dev/fd` pathname substitutes. User commands
inherit neither cwd/control descriptors nor secrets. PTY supervisor owns the
controlling tty; the user shell starts in another foreground PGID in the SID.

The supervisor outlives root exit, has one waiter per child, and controls ordinary
foreground/background groups through acquired membership: spawn a private helper
with setpgid before exec, validate its own SID/group and authenticated Joined ->
Commit handshake, then signal its own group with the fixed permitted `kill(0)`
operation. The helper pins the group lifetime. Census-selected numeric PID/PGID
signaling and precheck-to-kill are forbidden. Grace uses TERM; escalation acquires
parked STOP helpers, repeats full census until groups cannot create new members,
then KILL helpers. Expected helper SIGKILL is not proof of whole-session cleanup.

Require no live SID members except supervisor and joined user/helper waiters,
then Quiescent -> parent Release -> supervisor exit -> exact parent wait and
PTY/pipe/control joins. Output continues while Release closes the final slave;
waiting for EOF before Release is forbidden. Supervisor failure, incomplete
census, unacquirable live group, permission failure or unexpected member in the
reserved supervisor group is residual failure, never numeric recovery. Linux
uses bounded `/proc` identity/session census; Darwin uses native kinfo/getsid;
FreeBSD uses the selected bounded base-system `/bin/ps` numeric fields or a
separately reviewed native ABI replacement. Deliberate new-session/daemon escape
and external-service launch are outside portable ownership; observed escape is
residual failure. The corrected design accepts no numeric reuse race.

Windows follows WindowsBroker--001 in full. Native-machine selection preserves
386-to-native64 and ARM64 emulation using the trusted current executable or
embedded native helper. Main assigns a suspended broker to its noninheritable
kill-on-close outer Job before Resume; no breakaway or broker-held outer-Job
handle. Broker owns the nested user Job and HPCON locally. Effective directory
chain guards use actual list/read access with READ|WRITE/no-DELETE sharing and
real data-read nonempty child anchors; metadata-only guards are insufficient.

On one LockOSThread owner, broker creates the user root debugged+suspended,
assigns the inner Job before Resume and begins output draining. Retain guards
through the explicit target-ABI initial runtime breakpoint (initial 30s startup
deadline). CreateProcess return, a sleep, an arbitrary first breakpoint or a
pathname check is insufficient. At the pending supported event, read the correct
PEB/ProcessParameters cwd handle, duplicate it and compare FileIdInfo to the
selected guard; unknown exception/profile, zero/mismatched/incomplete handle
observation refuses. Release only exact owned anchors/guards and detach while
that event is pending, before user initialization proceeds. No handle injection
or process-memory writing substitutes. Only successful barrier/detach establishes
Start; native ABI/loader compatibility is mandatory implementation proof.

Control/output transports are separate. Versioned nonce/sequence-bound control
frames are at most 64KiB with closed opcodes; reject malformed/replayed/foreign
frames. Parent owns the sole output reader/ring after acknowledged kernel-pipe
handle transfer; broker closes its transferred reader copy. HPCON is not a
duplicable handle. Input/resize/close stay in the broker; parent drains through
terminal close. Normal cleanup proves inner Job zero, root wait and terminal/I/O
quiescence; broker Quiescent -> Release -> broker exit precedes final parent
outer-Job termination/zero-membership and parent control/output joins. Requiring
outer membership equal one before Release is forbidden because terminal
auxiliaries cause a circular wait. Forced outer abort reports its actual path
and joins/barriers; it cannot claim a killed broker ran graceful close callbacks.

Every acquisition has immediate rollback ownership, including debug-image/DLL
handles, suspended roots, anonymous endpoints and helper extraction. Native calls
wrapped in a goroutine remain resources until that goroutine joins. Stop success
requires descendant, root-wait, terminal, reader/writer/control/helper/broker and
extraction barriers, not a signal, root exit, Quiescent or close request alone.

## Native helper build and extraction obligations

Runtime owns broker, broker/cmd, brokerassets and cmd/helpergen subpackages;
broker dependencies exclude brokerassets to prevent recursive embedding.
Composition owns only generator/verification invocation in CI/release. Committed
`broker-amd64.gz`, `broker-arm64.gz` and a provenance manifest make ordinary
clean-checkout builds work without generation/download/compiler. Windows386
embeds both; Windowsamd64 embeds ARM64; native hosts use the trusted current
executable as specified. The twelve existing public asset names remain unchanged.

Generate on canonical Windows/amd64 Go1.25.0 with pinned modules, CGO_ENABLED=0,
explicit target/microarchitecture, `-trimpath -buildvcs=false -ldflags=-buildid=`
and deterministic gzip. Manifest binds protocol/schema, machine, lengths, both
hashes, toolchain/options/modules and sorted normalized source-closure digest;
generated outputs are excluded by dependency topology. No self-referential commit
SHA. CI independently regenerates both targets and checks exact bytes/machine/
closure, then ordinary builds. Changed helper closure requires reviewed generated
inputs; cross-host reproduction is claimed only with separate proof.

Extraction uses an exclusive cryptographic-nonce directory, protected current-user/
SYSTEM DACL, native no-reparse acquisition and data-read guards. CREATE_NEW,
complete write/flush, read-only reopen with READ sharing only, original FileIdInfo,
machine/protocol/hash verification and retained no-write/no-delete image guard
precede execution. No trusted filename cache. Delete only exact owned image and
empty directories after complete outer cleanup; preserve unexpected added/replaced
entries as residuals, with no recursive/age/PID cleanup. Helper startup cannot
perform project-cwd work before authenticated private handshake.

## Errors, cancellation and forbidden calls

The shared Diagnostic/ErrorCode and EffectReport rules apply. Invalid capability,
NotFound, Busy, stale cwd/source, Unsupported, Permission, ProcessFailure,
CleanupIncomplete and Indeterminate remain distinguishable. A non-nil Go error
does not erase populated result facts. An error before any admission/acquisition
has no Runtime effect; after possible creation, report the actual retained ID,
phase, effects and residuals. Never infer rollback or Cleaned from context error.
Known AppliedVerified resources followed by failed observation remain known
establishment plus unavailable observation; panic recovery after possible effect
is Indeterminate at the Application boundary.

Start cancellation before establishment initiates owned partial cleanup; after
successful persistent establishment, expiration of its original context does
not kill the session. Stop/Shutdown cancellation limits caller waiting while
owned cleanup continues. Restart checks cancellation before new acquisition;
after replacement begins it follows Start's full cleanup rules. Read cancellation
has no side effects. Write/Resize/Interrupt canceled before admission/delivery
refuse; after acceptance/delivery, cancellation cannot retract bytes, geometry or
interrupt, and the result must retain that known effect. No automatic retries.

Runtime cannot discover providers, load/save storage, call Git/GitHub, allocate
Application IDs, mutate State/focus or infer active-worktree changes. Other
adapters cannot use Runtime's private helper or persistent-session guarantee for
their short commands. No public arbitrary PID/signal/control socket or secrets in
argv/environment. No direct-child-only Stop, guessed broker ABI, broadcast console
Ctrl event, xpty wrapper success as ownership proof, no-op cleanup success, early
ACK, mutable snapshot reuse or unbounded lifecycle/input/output queue is allowed.

## Verification and change control

Proof is required at exact implementation/integration SHA under
[Verification--001](../Design/CR-%2321/Verification--001.md), not inferred from
accepted design or feasibility probes:

| Contract clauses | Mandatory proof |
|---|---|
| Identity/admission/copy/start failures | V-DOM-01/03, V-RUN-01, V-APP-01/02/06 |
| Stop/restart/phase/cleanup/error/cancel/shutdown | V-RUN-02/03/04/05/07, V-APP-01/02/06, V-COMP-02 |
| Reliable transfer/ACK/full queues/detached consumer | V-APP-01, V-RUN-06/07; fill both capacities, replay between transfer and ACK, timeout/detach during drain, no duplicate normalized cleanup |
| Raw output/positive geometry/bounded copied input | V-RUN-05/06, V-VIEW-02/03 |
| Acquired cwd/native helpers/descendants/ABI/partial unwind | V-RUN-01/03/04/05/08, V-LCH-02/03, V-COMP-02 |
| Helper reproducibility/extraction/twelve assets | V-RUN-08, V-COMP-03/04, V-REL-01/04 |
| Complete vertical behavior | V-E2E-10/11/12, relevant V-E2E-09/13, V-APP-04 and V-STATE-02/03 |

Native proof includes Windows amd64 and ARM64 with 386/WOW64 and supported
x86/x64 emulation, native32 layout evidence, Linux, macOS and FreeBSD census/PTY
mechanisms; all twelve cross-builds remain mandatory and separate. Prove actual
root/child/grandchild identity and membership, every acquisition/teardown failure,
blocking ConPTY close, TLS/DLL/debug-heap/immediate-chdir startup, native cwd
replacement at each barrier, and resource counts after repeated cycles. A killed
root or successful cross-build cannot satisfy these gates.

Change history: 2026-09-06, initial DRAFT 1.0.0 under #55; no superseding contract.
Freeze requires fresh independent whole-set review of seven BCs plus shared
annex at exact HEAD, corrections/re-review, explicit metadata freeze review and
configured green CI. A substantive conflict returns through design/BC-CHANGE.
After freeze, ordinary layer workers cannot edit this contract; Master must
assess impact, review/refreeze, notify affected workers and rerun affected proof.
