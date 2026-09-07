# Runtime

Issue #65 under #21 owns this directory. The normative authority is the frozen
Application--Runtime 1.0.0 contract, shared BoundaryTypes, CwdAcquisition,
WindowsBroker, and Feasibility/Runtime RTF-02. No old launch/terminal/process
implementation is imported or moved here.

This is an explicit partial M3 checkpoint. The current implementation contains
the private registry, bounded memory/framed transports and native Unix cwd,
helper and supervisor components. It has no exported Sessions constructor,
complete parent native client, public lifecycle stubs, or Composition
cutover. Passing these tests does not establish process containment or cleanup.
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
Sessions methods remain required by #65 before M3 Runtime acceptance.

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
