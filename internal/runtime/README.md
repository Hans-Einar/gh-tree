# Runtime

Issue #65 under #21 owns this directory. The normative authority is the frozen
Application--Runtime 1.0.0 contract, shared BoundaryTypes, CwdAcquisition,
WindowsBroker, and Feasibility/Runtime RTF-02. No old launch/terminal/process
implementation is imported or moved here.

This is an explicit partial M3 checkpoint. The current implementation contains
the private registry, bounded memory/framed transports and native Unix cwd
acquisition. It has no exported Sessions constructor, public lifecycle stubs,
native startup backend, or Composition
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
