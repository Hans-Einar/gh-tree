# Runtime

Current #71 binding milestone: Master combined independently accepted common
and native/helper sources at `4412fe8e`. The private Unix bridge now drives the
common twelve-method engine with actual native clients, including retained cwd,
typed failures, input, terminal controls and complete cleanup. Root native tests
exercise real owned processes; earlier component evidence remains applicable.
Windows parent binding, Auto shell selection, production construction and
Composition cutover are still pending. The historical milestone notes below do
not override this current state; exact evidence is in `M3-Sessions--001.md`.

Issue #65 under #21 owns this directory. The normative authority is the frozen
Application--Runtime 1.0.0 contract, shared BoundaryTypes, CwdAcquisition,
WindowsBroker, and Feasibility/Runtime RTF-02. No old launch/terminal/process
implementation is imported or moved here.

This is an explicit partial M3 checkpoint. Issue #71 adds the common twelve-method
Sessions engine through `newSessions`, an unexported constructor taking a retained
`nativeOwner` startup seam. Only test-controlled owners currently construct that
engine. Real platform bridges, production construction and Composition cutover
remain pending reviewed native clients and their matching helper assets. Passing
the common engine tests does not establish native containment or cleanup.
The mandatory M1 helper prerequisite currently fails because the native broker,
entry, verifier, embedded images, and provenance manifest are not yet implemented.

`registry.go` owns membership, monotonic allocation, final-event reservation,
admission closure and the latest 256 cleaned records. Live and cleanup-failed
records remain protected by the 64-live admission ceiling. Accepted invocation
values use the immutable M2 API directly; the resolved private environment is
copied. The memory-lock order is registry, then session, then event buffer. Native
calls and waits must occur outside these locks. A native lifecycle owner still
must prove a resource barrier before supplying a Cleaned snapshot.

`events.go` retains one reserved final per admitted session until cumulative ACK.
There are at most 256 reserved or unacknowledged finals, one coalesced hint per
live session, and 64 unacknowledged hint delivery receipts. After the receipt
bound, new hint deliveries coalesce until ACK; reliable final delivery continues
using its independent reservation. This bounds memory needed to reject unknown
or regressing cursors without blocking native producers. An already delivered
unacknowledged cursor remains valid. Hint sequence gaps are intentional. Numerical
sequence space is reserved for pending finals so hints cannot consume the last
usable final sequence. Native producer joining and closed admission remain the
future lifecycle owner's prerequisites to closing the event producer side.

`output.go` retains at most 256 KiB of raw combined output, exact byte offsets,
stream identity and source session sequence. Span metadata is also bounded. It
performs no ANSI, text or line interpretation. Reads copy chunks, identify precise
overflow gaps and refuse future offsets; append refuses uint64 offset wrap.

`input.go` copies each nonempty request atomically with a 64 KiB ceiling including
in-flight bytes. A single writer owns a popped buffer until completion accounting.
Closing admission discards and counts queued bytes but preserves in-flight
ownership. The native writer must record accepted/delivered byte counts, close
its native endpoint and actually join; this queue is not native write evidence.

Run `go test ./internal/runtime -count=1`, `go test -race ./internal/runtime
-count=1`, and `go vet ./internal/runtime`. Current tests cover randomized
raw-byte/stream/sequence retention, bounded span storage, input copy/capacity,
event replay/ACK/cancellation and receipt limits, final numerical reservation,
concurrent registry admission/shutdown, ID exhaustion and cleaned-only history.
They create no native user processes or project scripts.

`broker/protocol.go` implements the bounded version/nonce/session/role/sequence
wire envelope, and `broker/start.go` encodes exact private startup inputs with
bounded length fields and no unknown/trailing payload. Each direction has one
decoder. Partial-write or malformed/replayed input poisons that direction; its
native owner must close and join the endpoints. Complete frame writes serialize
on a private transport lock, separate from session/registry locks. Native owners
must provide inherited anonymous endpoints and independently validate the actual
parent and role before any user work. This codec alone authenticates no OS owner.

`broker/cwd_unix.go` acquires the absolute root and relative project by retained
no-follow directory descriptors, checks the supplied root/project identities,
and revalidates every named parent-child relation. Identity uses the established
M3 dev/inode plus available birth stamp or explicit change-stamp profile. The
acquired descriptor preserves the original object after replacement; observed
relocation before launch refuses. Partial acquisition returns its cleanup owner.
The main process never changes cwd and no fd pathname is used. Linux execution
as UID/GID65534 proves the acquisition/substitution tests; Fchdir in the private
supervisor and native macOS/FreeBSD execution remain separate required proof.

Real supervisor/helper, Job/ConPTY, complete failure-unwind and all twelve
Sessions methods bound to real native clients remain required by #65 before M3
Runtime acceptance.

`broker/signal_unix.go` now implements the actual session-local signal helper and
its acquisition owner: inherited anonymous pipe direction/type and poller checks,
setpgid before exec, own parent/SID/group validation, authenticated Joined/Commit,
the sole permitted self-group `kill(0)` call, parked STOP helpers and exactly one
retained waiter. No numeric census signal fallback exists. Native Linux fixture
tests exercise foreign-SID rejection, cancellation/nonce failure before signaling,
departure of the last original member after Joined, parked STOP and acquired KILL,
all wait joins and final full SID census. Test failure teardown may kill only its
own directly created fixture children and records that as a test failure; it is
never product cleanup evidence.

Unix census selections are Linux bounded `/proc` stat reads, Darwin bounded native
sysctl with pinned x/sys KinfoProc layout plus Getsid and identity recheck, and the
frozen FreeBSD bounded `/bin/ps` numeric profile with its exact observer waiter.
These observations choose acquisition candidates and classify residuals. They do
not authorize numeric PID/PGID signaling. Native Darwin/FreeBSD execution remains
required; cross-compilation only verifies selected code/layout availability.

`broker/supervisor_unix.go` and `tree_unix.go` provide the SID owner: authenticated
Start, designated descriptor Fstat/Fchdir exactly once then close before command
lookup, inherited cwd and copied PWD/environment, different user-root PGID and
PTY foreground group, one root/helper waiter, TERM then parked STOP/full recensus/
KILL, root-exit-triggered cleanup, parent EOF ownership, Quiescent/Release and
supervisor exit. Failed ownership retains resources and refuses successful
quiescence. Native Linux tests prove cwd marker access, root-before-child-and-
grandchild cleanup, a real shell with foreground/background pipeline groups,
100x30 terminal resize and reader/wait release ordering. Whole-client partial
acquisition/error/cancel/input/resize ownership and all twelve Sessions methods
still need assembly and full failure verification. The CLI has no private-mode
dispatch cutover; tests dispatch the actual native functions in their own binary.

`sessions.go`, `controls.go` and `lifecycle.go` now assemble the common engine.
Admitted startup retains an ID and reliable final reservation even when native
acquisition fails. Start contexts govern establishment; persistent established
owners survive later request cancellation. Root exit closes input admission while
output continues until the native output/cleanup barrier. A public Cleaned fact
also requires the parent input writer and every accepted control observer to
finish. Native receipt cancellation never releases parent ownership or authorizes
replay. A 65536-byte parent write keeps one queue reservation while splitting into
at most 65484-byte native chunks. One control admission at a time bounds pending
resize/interrupt work; overlapping controls return Busy, and Stop bypasses that
admission. Event consumers never participate in these native waits.

Lifecycle-key checks use retained session records. Restart retains its replacement
identity as soon as admission occurs, including cancellation during startup, and
reuses the private original environment plus the latest delivered geometry. An
old session with an existing replacement transition cannot create another one.
Shutdown closes admission atomically, retains unfinished Start/Restart work and
reports native/parent/event-transfer residuals. Cleanup and final ACK remain
separate: resource completion alone does not make aggregate shutdown successful.

Eleven deterministic engine tests in `sessions_test.go` add startup cancellation,
late raw output, pending input/control receipt barriers, exact input splitting,
restart identity/geometry/deduplication, shutdown races, repaired cleanup,
exhausted final sequence space and known cleanup with a historical error. These
supplement the buffer/registry tests. The bounded report and next native binding
gate are in `SDP/Implementation/CR-#21/Runtime/M3-Sessions--001.md`.

The independent common-engine corrections retain the admitted Restart subject
and operation key through history eviction, including canceled replacement facts
and Shutdown capture. Active transitions have a separate 256-entry bound; public
history still retains the latest 256 cleaned sessions. Stop and Shutdown share
one native observer and coalesce a recovery request across an incomplete-error
handoff without autonomous retries. An admitted control reserves a publication
version before native delivery. Other optional work refuses exhausted space;
unavoidable native facts remain private until a fresh version or the reserved
final publishes them. Nine correction tests cover these interleavings, numerical
boundaries and producer barriers. Native binding and full acceptance remain pending.

Startup now reserves its own publication version from admission through the
native startup result. Output and Stop cannot consume that version while startup
is pending; acquisition publishes known establishment/cwd or failed-start facts
coherently before releasing the reservation. Two further boundary tests cover
early output, concurrent Stop, failed/canceled startup, immutable snapshots and
the subsequent input/control receipt handoff to the final cleanup barrier.
