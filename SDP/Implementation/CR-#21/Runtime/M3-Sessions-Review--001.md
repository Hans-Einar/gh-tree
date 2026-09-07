# M3 Sessions common-parent independent review — #71

Disposition: **CHANGES_REQUIRED**. One HIGH and two MEDIUM implementation
findings remain open at frozen source `78a67e1198466e05bc5f2544b3c4a495dc45a52c`
(technical `fff5ee4a30f3b2d67997fcf74fb864a1174fead0`).

Date: 2026-09-07. Role: fresh independent reviewer, separate from the author.
Authority: #71 first common-engine milestone under #65/#21; Sprint-004-v04 /
I-03 / M3. Branch `codex/cr21-runtime-sessions`, worktree
`C:/Users/hanse/GIT/gh-tree-wt/runtime-sessions`. Source was clean and exactly
pushed at review start. Review edits are confined to this report and the small
additive evidence folder; no product/test correction or native integration.

## Scope and authority

Reviewed the complete root `internal/runtime` implementation and tests, including
the earlier registry/event/input/output foundations, not only the author delta
from `c3f23b89610a8695125fcd881711a929c7828374`. Runtime tree object is
`b6b34b0aaad51a4f07929dfde936c0bb1376b3f4` at the frozen source. Actual source,
M2 API validators/ports and independent execution determine this disposition;
the eleven author tests and author report are not independent acceptance evidence.

Reopened root AGENTS/developmentInstructions, full #21/#65/#69/#70/#71 with
comments, Application--Runtime and BoundaryTypes FROZEN 1.0.0, BCFreeze,
REFDES/API, CwdAcquisition, WindowsBroker, Runtime RTF-02 and accepted
Verification--001. Relevant Runtime review/lifecycle requirements and actual
M2 runtime records, consistency/evidence validators, port and local READMEs
were checked against the common code. No local SDP/Instructions folder exists.

This milestone intentionally has only `newSessions(nativeStarter, ...)` and
test-controlled retained owners. Missing exported production construction,
platform bridges, helper binding and Composition cutover are future gates, not
defects invented by this review. Accepted Windows receipt/failure source at
technical `bd78deaf` / review `9b06f8bd` was inspected read-only. No separately
owned Unix/Windows/helper implementation is re-reviewed or adopted here, and
known M65-U01/U02 on this development base remain in their separate correction
history. The separate blocked Git review remains untouched.

## Findings

### M71-H01 — Restart re-looks up an evictable predecessor and panics

Location: `internal/runtime/lifecycle.go:131` and `:155`; selection begins at
`:75`. `registry.go:139`–`:142` independently evicts cleaned history.

Restart retains the selected `*session`, then its asynchronous worker calls
public Stop by ID again. Another session's natural cleanup can evict that old
cleaned ID between those lookups. Stop returns NotFound with a zero result;
the worker passes that zero as `Old` to `NewSessionRestartResult`, and `owned`
panics. This panic occurs in the Runtime goroutine, outside an Application
call-boundary recovery, and terminates the process.

Independent `TestReviewRestartHistoryEvictionRace` admits and cleans/ACKs 256
sessions through the actual engine, starts one more controlled live owner,
pauses Restart after its successful lookup via a context Err scheduling barrier,
and lets the live owner's natural cleanup evict the oldest record. Resuming
reliably produces `panic: invalid boundary value: d.Old`, stack at
`lifecycle.go:155`. No product state is injected in this test. The paired
`TestReviewRestartHistoryWithoutEviction` passes through the same setup and
returns a valid old/replacement result when the eviction is delayed.

Required correction: make the accepted transition retain/use its exact
predecessor throughout cleanup/replacement/result formation, including history
eviction and canceled replacement attribution. Do not construct a required
result from an invalid failed lookup. Preserve bounded history, deduplication,
new-ID ordering, complete old cleanup and aggregate transition ownership.
Authority: frozen restart serialization/history and typed error-plus-fact
clauses; V-RUN-02/07.

### M71-M01 — Shutdown never resumes observation of a retained failed owner

Location: `internal/runtime/lifecycle.go:200`–`:206`; compare Stop's explicit
re-observation at `:31`–`:40` with `sessions.go`'s incomplete-error exit from
`observe`.

After NextFact returns an incomplete fact plus error, `observe` retains the
owner, clears `observing` and exits. Shutdown only calls `requestStop`; it never
restarts observation. Its wait immediately sees CleanupFailed. Subsequent
Shutdown calls repeat that behavior, so a repaired native cleanup cannot be
observed/finalized through the aggregate API. The session, final reservation
and shutdown residual remain indefinitely unless a separate Stop happens.

Independent `TestReviewShutdownReobservesRetainedFailure` supplies an incomplete
error followed by known complete cleanup. Two Shutdown calls leave exactly one
NextFact call and both report incomplete. Calling Stop on the identical retained
owner observes the second fact and reaches Cleaned, proving this is the missing
aggregate recovery path rather than an unrepairable owner. No false successful
cleanup is claimed; the defect is stranded observation and permanent residuals.

Required correction: share the retained-owner retry/observation ownership with
Shutdown, coalescing concurrent observers and preserving its one overall wait
budget. A retry must still retain genuine failures and wait for parent producers;
final ACK remains a separate completion requirement. Authority: frozen retry,
idempotent aggregate cleanup and V-RUN-07.

### M71-M02 — A delivered resize loses its geometry at sequence exhaustion

Location: `internal/runtime/controls.go:132`–`:174` and
`internal/runtime/registry.go:123`–`:124`.

Control admission does not reserve/check session publication space. At
SessionSequence MaxUint64-1, native Resize succeeds, but `registry.change`
rejects the nonfinal snapshot after editing a temporary data copy. The ignored
error leaves the old display geometry. Resize nevertheless returns Delivered
true with no error; finalization consumes the reserved last sequence and still
publishes the old geometry, which Restart reads as its original latest geometry.

Independent `TestReviewResizeAtSequenceLimitPreservesDeliveredGeometry` places
only the counter at the specified boundary (explicit numerical-edge fixture),
delivers 57x119, then supplies native cleanup. The result reports delivery while
the final at MaxUint64 contains 24x80. This is an extreme counter case, not an
ordinary session-duration claim, but wrap/refusal and latest accepted geometry
are explicit requirements of this milestone.

Required correction: refuse before an unrecordable native effect or reserve/
retain the accepted effect through concurrent output and the final publication.
Account also for a resize already in flight when the last hint version is used;
do not merely add an unsynchronized precheck or reuse an old published version.
Authority: frozen sequence/refusal, control-effect and restart-geometry clauses;
V-RUN-01/05.

## Independent verification and retained strengths

Native host: Windows amd64; explicit Go1.25.0 module-toolchain executable,
existing CGO/race compiler, no installs or global changes. Tests use controlled
owners only and create no native user processes. The evidence manifest records
exact commands, return codes and SHA-256 hashes under
`SDP/Verification/CR-#21/Evidence/M3-Sessions-Review--001/`.

| Check | Result |
|---|---|
| Frozen source `go test -race ./internal/runtime -count=1 -timeout=60s` | PASS, all existing root foundation/engine tests. |
| Frozen source `go vet ./internal/runtime` | PASS. |
| Independent eviction adversary, with race instrumentation | FAIL: reproduced process panic (H01). |
| Independent aggregate repair and exhausted-geometry adversaries, with race instrumentation | FAIL: M01 and M02 reproduced. |
| Paired no-eviction restart control | PASS. |
| Pending to terminal-unknown receipt, canceled caller and native cleanup first | PASS: parent final withheld until receipt observation; unknown effect does not invent geometry; typed cause retained. |
| Completed delivered resize plus supplemental error | PASS: known Delivered and geometry retained alongside the error. |
| Native resize reenters List and captures raw NUL/invalid-UTF8/ESC output | PASS: no registry/session/event lock spans the native call; exact bytes retained. |
| Product tree / `git diff --check` | Unchanged / PASS. |

The additive Go test is stored as `.go.txt` and selected using a Go overlay;
it is not a product test edit or part of ordinary source discovery. A first
reviewer-only compile attempt used the terminal-mode constant where the distinct
output-stream constant was required; corrected before the execution above.
It is not product failure evidence. No duplicate source archive was created.

Source audit and passing tests support atomic membership/ID/final reservation,
bounded live/history/final queues, event replay/cumulative ACK and hint receipt
bounds, raw output ownership/offsets, whole copied input admission including
in-flight 65484+52 framing, and retained input/control receipts without replay.
Natural root exit remains distinct from native cleanup; parent producers block
the final. Canceled starts retain identity and established owners survive their
original context; restart and shutdown admission share serialization. These
strengths do not cancel the three findings above.

The private common delivery seam can represent pending, known and terminal-
unknown outcomes. Future bridges must normalize actual platform facts, not copy
field meanings blindly: Windows successful Resize acknowledges Completed with
zero byte counters, whereas this parent represents delivered controls with a
positive Delivered unit. Closed native cause/stage normalization, actual native
resource joins and retained receipt behavior require real binding verification.
No missing bridge is scored as a defect in this bounded candidate.

## Exact next permitted action

Master assigns a fresh author to correct only M71-H01/M01/M02 in the authorized
root Runtime scope, commit/test/push the changed source, then return its exact
SHA to this independent reviewer for affected re-review. Preserve the original
rejected source and this compact evidence. No author correction is included in
this review commit.

Even after these corrections pass, any acceptance is limited to the common
parent engine. Reviewed real native clients/assets, production construction,
twelve-method native integration, all accepted platform/ABI/PTY/ConPTY/helper
and full Runtime gates remain mandatory. No complete #71/#65/Runtime/Slice,
canonical/main integration or release acceptance is granted.
