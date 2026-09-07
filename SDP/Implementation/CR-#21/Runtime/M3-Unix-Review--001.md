# M3 Unix native implementation review — #65 / #21

Current disposition: **ACCEPT — Unix/shared native candidate only**, after bounded
independent confirmation at `d0a7682e761a7c076feeaf2e4d1ab0476edf3ae7`.
M65-U01/U02 are resolved; no open finding remains in this review's scope.
The original **CHANGES_REQUIRED** disposition and both MEDIUM findings below
describe the initial frozen source and remain preserved as historical evidence.

Reviewer: `m3_runtime_unix_review`, fresh independent reviewer; not the author or
Master. Initial review source: `codereview-21/layer-runtime` at
`fe4fd4266d3798132d4cca15ce54dc991ad147b1`, clean and equal to origin at inspection.
Source kickoff: `26d5f4fe7b87c8ad19a8a335bc4af19e06b647e2`.
Master dispatch: `828d316523bd596f481d7c65b774a9e0d012eb7c`, ledger89.
Sprint-004-v04 / I-03 / M3. No product source or existing tests were edited.

## Scope and authority

This reviews the actual Unix-selected `internal/runtime/broker` source and tests,
including common `protocol.go`/`start.go`, against frozen Application--Runtime and
BoundaryTypes 1.0.0, BCFreeze--001, accepted REFDES/API, CwdAcquisition,
WindowsBroker, selected Runtime RTF-02 and Verification--001. Root AGENTS and
developmentInstructions, full Issues #21/#65/#69/#70/#71 and comments, the Runtime
README and current M3-Runtime clause/source/test/platform matrix were inspected.
Historical numeric-signaling and direct Windows-start proposals are superseded.

The root registry, sequences, event reservations/ACK, memory queues and twelve
Sessions methods belong to the separate #71 review. Windows native implementation
is #69; helper generation/assets are #70. Missing constructors, platform bridges,
assets and host cutover are explicit later gates, not invented defects in this
Unix contribution. Nothing here accepts Runtime as a whole, a Slice, integration
or release. The separate blocked Git review was not inspected or retried.

## Findings

### M65-U01 — MEDIUM: directional named FIFOs pass private endpoint admission

Locations: `internal/runtime/broker/endpoints_pipe_unix.go:7` and
`endpoints_unix.go:15` (Linux and Darwin selection).

`inheritedPipe` checks `S_IFIFO`, then `validNativePipeAccess` checks only
O_RDONLY/O_WRONLY. A named FIFO, including one already unlinked, has those same
properties. The independent native test
`TestReviewRejectsNamedAndUnlinkedDirectionalFIFO` creates only a fresh owned
FIFO, opens each direction, duplicates each descriptor and calls the actual
admission function. All four linked/unlinked × read/write cases are accepted
with `err == nil` on Linux as UID/GID65534. No signaling was needed for this
negative. The baseline anonymous-pipe positive still passes.

RTF-02 makes separately inherited **anonymous** endpoints the capability and
explicitly says a nonce from a named file is not authentication. This gate must
reject named transports before either private worker accepts a handshake. The
FreeBSD implementation already distinguishes actual anonymous pipes and tests
named/unlinked negatives; the Linux/Darwin branch omits that distinction. This is
a contract-admission defect, not a claim of a sandbox or arbitrary foreign-SID
signaling: the separate actual-parent/SID/group checks remain present.

Required correction: verify the actual native anonymous endpoint profile on
Linux and Darwin while preserving directional access, poller/deadline and
close-on-exec semantics. Add linked/unlinked named FIFO negatives and real
anonymous positives on both native platforms; preserve FreeBSD's duplex profile.
Linux reproduction is native evidence; Darwin impact follows the same selected
source and still requires native confirmation of its correction.

### M65-U02 — MEDIUM: natural exit ignores construction cleanup periods

Locations: `internal/runtime/broker/supervisor_unix.go:149`, `:235`, `:248`;
`client_unix.go:539` and `:574`.

The client accepts GracePeriod/ForcePeriod but transfers them only in a later
Stop frame. The supervisor initializes its tree with 2s/3s and enters synchronous
cleanup immediately when its exact root waiter completes. That path has not
received the client periods. An observation failure or parent EOF before Stop
has the same configuration gap. Sending Stop after the cleanup loop starts also
cannot change that already-running attempt.

The independent `TestReviewNaturalExitUsesConstructionBudgets` uses the existing
owned `--runtime-fixture-tree-root` (ordinary child/grandchild groups ignoring
TERM), with 20ms grace and 200ms force. At 1.202s the real root has exited but
quiescence is still false; the 1.2s observation times out. Normal acquired cleanup
eventually completes at 2.091s with zero residuals, matching the hardcoded grace.
The test always retains and joins its client; no numeric fallback or fixture
signal is used. This is not a claim that arbitrary kernel failures must finish
within a budget: the source omits the configured periods on this lifecycle path.

REFDES and Application--Runtime fix central defaults with construction-injected
periods; V-RUN-02/07 require root-before-descendant and bounded cleanup semantics.
The existing client test checks configured periods only after explicit Stop, so
it misses this combination.

Required correction: bind validated native cleanup periods before user execution
and any natural/EOF/failure cleanup can start, keeping defaults 2s/3s and retaining
truthful timeout/residual ownership. Coordinate any shared protocol/source-closure
change with Master and the other native owners. Cover natural root exit plus
retained descendants with construction periods; do not send a premature Stop or
relax the budgets merely to satisfy the test.

## Independent inspection and execution

All Unix production owners, platform census/stamp/endpoint selections and their
actual tests were inspected. The following substantive behavior is present and
the existing native controls pass; this does not override the findings above.

| Area | Evidence assessed |
|---|---|
| Cwd ownership | Retained no-follow directory chain, supplied identity/profile checks, parent-entry revalidation; supervisor Fstat/Fchdir once, descriptor close before lookup, empty Cmd.Dir and copied environment/PWD; original-object, stale/link/substitution and private-fd positive/negative fixtures. |
| Process/session lifetime | Private supervisor retains SID; user root starts in another PGID; real shell foreground/background pipelines, PTY geometry/ETX; exact root/helper waiters and root-before-child/grandchild cleanup. |
| Signaling | Setpgid before helper exec, own parent/SID/group validation, fresh inherited channel binding and Joined→Commit; only `kill(0, permittedSignal)`, no census-number fallback; vanished/foreign/last-original-member cases, parked STOP and KILL helpers. |
| Escalation/release | TERM then acquired STOP/full stopped-member recensus/KILL; complete SID census and waiters before Quiescent; Release before final supervisor/PTY EOF; observed escape and broken control remain residual. |
| Native parent/concurrency | Copied full/partial input and cancellation facts, serialized writes, cancellation watchers, no callback under client fact lock; Stop prevents further I/O admission, actual wait/decoder/input/output joins; blocked output owner/panic, helper permission failure then repair, 17 partial acquisition checkpoints. |
| Common codec | Bounded length-prefixed version/role/opcode/session/nonce/sequence checks, exact bounded StartSpec decoding, copy ownership, partial-write poisoning and serialized complete frames. U01 is the native endpoint gate, outside the codec's own authority. |
| Repeated resources | Twelve pipe/PTY cycles: 102 registered native descriptors closed/unusable, Linux fd count 6→6, goroutines 2→2; other OS global inventories are not inferred from `/proc`. |

Independent builder: Windows/amd64 Go1.25.0, CGO_ENABLED=0; compiled the unchanged
broker tests and a separate Go `-overlay` reviewer binary. Native execution:
existing openSUSE-Leap-15.5 WSL, Linux
`6.18.33.2-microsoft-standard-WSL2` amd64, UID/GID65534, owned ext4
`/tmp/gh-tree-unix-review-fe4fd426.jLhLxC`, not a Windows-mounted working directory.
The unmodified full broker suite (`-test.v -test.count=1 -test.timeout=120s`)
**PASS**. The independent two-test overlay (`-test.run=TestReview`)
**FAIL**, exactly as recorded above; both fixture owners clean successfully.

Critical reviewer evidence is in
`SDP/Verification/CR-#21/Evidence/M65-Unix-Review--001/`: native logs, the external
`.go.txt` overlay and a small hash/command manifest. No source tree, binary, cache,
generated helper image or duplicate CI archive is committed.

## Exact CI and limits

Independently inspected actual run/job metadata and relevant raw logs for
[run34073961901 attempt1](https://github.com/Hans-Einar/gh-tree/actions/runs/34073961901),
whose checkout/source is exactly `fe4fd4266d3798132d4cca15ce54dc991ad147b1`:

| Job | Result and evidence |
|---|---|
| 101596402774 macOS | SUCCESS; Go1.25.0 darwin/arm64, actual broker package execution 9.661s, native test/vet/build. |
| 101596402834 Linux race | SUCCESS; Go1.25.0 linux/amd64, actual `go test -race ./...`, broker 11.225s. Test-only subprocess `atexit_sleep_ms=0` preserves race instrumentation and product defaults. |
| 101596402647 FreeBSD | SUCCESS; FreeBSD15.0-RELEASE-p12 amd64 guest, UID/GID1001, Go1.25.0; exact broker tests execute and pass (9.504s), including acquisition, EOF/escape, resource cycles, PTY and helper controls. |
| 101596402864 Linux; 101596402841/846 Windows | SUCCESS native full-suite jobs. Windows jobs on this branch do not prove the separately owned #69 implementation. |
| 101596402839 prerequisite | FAILURE: `M3 prerequisite missing: internal/runtime/broker/cmd Go source`. Cross-build/helper-dependent jobs SKIP. This is an explicit #69/#70 gate, not an aggregate passing workflow. |

FreeBSD full artifact: `10001433337`,
`freebsd-native-fe4fd4266d3798132d4cca15ce54dc991ad147b1-1`; 59,205 bytes,
uploaded ZIP SHA256
`43d89f478f57d57b79afc246bce19860269c745a7a6e6fc326182ee16adafd0b`.
The source matrix's earlier failures remain historical failures: da295696 needed
the actual FreeBSD duplex-pipe correction a6e79eaa; 9a73ffc failed PTY/race
fixtures; 0dfd887 still failed the Darwin KILL helper lifetime; a57d496 repaired
that lifetime and passed six native/race jobs. Their passing successors do not
erase those failures or prove this review's new negative controls.

All-nine-Unix compilation recorded by the author is distinct from this review's
Linux execution and current hosted native proof; cross-build success is never
native proof. Full parent Sessions/ACK/restart/aggregate shutdown, Windows/native
ABI/assets, composition cutover, all-twelve packaging, exact integrated-source and
vertical/release acceptance remain their existing later gates.

## Initial handoff

No source edits, merges, reset, cleanup of unrelated fixtures or new review
subagents. Findings were reported promptly to Master. Product tree and origin
remain frozen at fe4fd426; the review commit adds only this report and its small
critical evidence. Exact next permitted action: Master dispatches the author for
bounded U01/U02 corrections, obtains actual corrected native evidence and a new
frozen source SHA, then requests independent affected-source re-review. Preserve
the already-passing unaffected evidence and all separate program gates.

## Bounded independent correction confirmation

Master dispatch `e43297acbbff385315062bbaec7b5a58c79678a8`, ledger94. Reviewed
clean/pushed candidate `d0a7682e761a7c076feeaf2e4d1ab0476edf3ae7`, matching origin;
complete executable/test tree matches tested
`71b95a32265270c43da58ff41f67ab370a39a245`. Product source is unchanged from
`4df4385fbb2481a0c0007dbfd489c93e15f3170c`; the final technical delta only retains
the malformed-fixture writer through the exact supervisor wait. The required
exit131/no-user assertion remains, and timeout still closes control and joins.
No product implementation was performed by this reviewer.

Reopened the actual complete diff since `c3f23b8`, changed Unix files/tests and
current author record. The common protocol/StartSpec codec/tests, FreeBSD duplex
admission, public/private exported client signatures and root Runtime files are
unchanged. New startup code is selected only on Linux/Darwin/FreeBSD, so this
correction does not change the Windows helper dependency source closure.

**M65-U01 resolved.** Linux now verifies the anonymous pipefs superblock in
addition to FIFO type/direction. Darwin adds its non-vnode Fstatfs profile after
those checks. Independently inspected pinned
[Linux pipe creation/pipefs source](https://github.com/torvalds/linux/blob/df2908090cda368b01ff43709f51890076c56157/fs/pipe.c),
[XNU file_vnode type selection](https://github.com/apple-oss-distributions/xnu/blob/f6217f891ac0bb64f3d375211650a4c1ff8ca1ea/bsd/kern/kern_descrip.c)
and [XNU fstatfs64 path](https://github.com/apple-oss-distributions/xnu/blob/f6217f891ac0bb64f3d375211650a4c1ff8ca1ea/bsd/vfs/vfs_syscalls.c).
The original reviewer linked/unlinked read/write negative now passes unchanged.
Actual anonymous positives still prove both directions, close-on-exec and read
deadline; reversed directions and named/unlinked FIFOs refuse. Darwin native
package execution confirms its selected implementation/tests; FreeBSD retains
its existing anonymous duplex and named/unlinked controls.

**M65-U02 resolved.** The Unix-only versioned 17-byte Start prefix binds both
effective uint64 nanosecond periods before user acquisition or any cleanup path.
Bounds reject zero/negative-overflow/out-of-range wire values; exact admitted
nanoseconds survive. Construction zero still means 2s/3s. The total frame ceiling
includes the prefix; the unchanged StartSpec decoder refuses trailing/unknown
data. Stop/Abort carry intent only and cannot update construction periods.
Natural exit, early/established control EOF and invalid control all use the
bound configuration while retaining the original cleanup/residual semantics.

The exact original two-test reviewer overlay was reused unchanged on corrected
source, with the endpoint and startup correction tests selected alongside it.
Windows Go1.25.0 built Linux/amd64 with CGO0; actual execution used UID/GID65534
on the same WSL Linux kernel and fresh owned ext4 directory
`/tmp/gh-tree-unix-confirm-d0a7682e.HIBIpU`. All selected tests **PASS**:

| Independent/native control | Observed result |
|---|---|
| Original U01 negative | All four named/unlinked × read/write endpoints refuse. |
| Original U02 negative | Natural root-before-descendants cleanup 92.598ms, complete, zero residuals. |
| New natural-exit controls | Injected 89.559ms; default 2.071s with effective 2s/3s. |
| Early EOF / established EOF / invalid control | Exact owners join in 35–36ms; control uncertainty remains residual, no false CleanupComplete or live owned SID member. |
| Startup admission | All thirteen malformed codec cases refuse; twelve native frame-admissible cases return131 with no user output; invalid construction acquires no owner; frame ceiling includes the wrapper. |
| Endpoint positive/negative controls | Anonymous profiles, CLOEXEC and deadline pass; reversed and named/unlinked endpoints refuse. |

All owned clients/waiters complete, and the final scoped executable inventory
finds no live confirmation fixture. `linux-confirmation.log` and
`confirmation.json` in the existing evidence directory preserve commands/hashes.
The original failed logs/overlay/manifest are unchanged. Unaffected full native,
common-codec and resource-cycle evidence from the initial review is reused.

Independently fetched current job metadata and raw relevant logs for
[CI34076324950 attempt1](https://github.com/Hans-Einar/gh-tree/actions/runs/34076324950):
exact tested source71b95a3, six native/race SUCCESS jobs. macOS101603045618 runs
Go1.25.0 darwin/arm64 broker tests (10.119s); Linux race101603045595 executes the
actual race suite (broker12.403s). FreeBSD101603045610 runs
15.0-RELEASE-p12 amd64 as UID/GID1001 with Go1.25.0 (broker9.573s); all twelve
native malformed-start cases and natural/default/EOF/failure controls pass.
Artifact10002195473 is62,502 bytes, ZIP SHA256
`0662e65b97b12219dfaf9303eaa792a512d6e09c7e60ea863624caf66991237e`.
The earlier4df4385 FreeBSD fixture failure remains disclosed in the author record;
it is not passing evidence. The corrected exact131 expectation is independently
confirmed locally and in the final FreeBSD log.

Inventory101603045569 still fails the separate missing broker/cmd prerequisite;
dependent build/helper jobs skip. Thus the aggregate workflow is not claimed
green. Windows jobs in this source are not #69 native implementation evidence.
All-nine compilation and all-twelve architecture checks in the author record
remain distinct from the independent native/race confirmation above.

**Final disposition: ACCEPT this Unix/shared native candidate.** Both findings
are resolved without changing frozen public contracts or the unrelated shared
Windows source closure. No additional material finding arose in the bounded
correction review. Parent Sessions/#71, #69 Windows, #70 assets, #72 and their
real-client composition, exact integrated source, complete Runtime/Slices/M8 and
release gates remain separate. Master alone chooses and records permitted
integration; no integration, product PR or release is performed by this review.
