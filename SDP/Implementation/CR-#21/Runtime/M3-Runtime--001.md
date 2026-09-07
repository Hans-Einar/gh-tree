# M3 Runtime contribution — #65

State: UNIX NATIVE REVIEW CANDIDATE / #65 OVERALL INCOMPLETE / NOT ACCEPTED

Parent: #21; worker authority: #65. Branch: `codereview-21/layer-runtime`.
Worktree: `C:/Users/hanse/GIT/gh-tree-wt/runtime-implementation`.
Starting exact HEAD: `26d5f4fe7b87c8ad19a8a335bc4af19e06b647e2`.
Master dispatch: `e9983bbea3b50a13032afef5e971656af3e6cc02`, ledger74.
Active Sprint-004-v04 / I-03 / M3, all SLC-01..13 remain selected.
Runtime contribution primarily SLC-10..12, prerequisite relations SLC-09/13.

The independent Git review has an external service hold. This Runtime authoring
is separate already-authorized #65 work; it neither retries nor substitutes for
that blocked review. No serial integration, acceptance, PR, tag or release is
authorized by this checkpoint.

## Authority and physical scope

AGENTS/developmentInstructions, current full #21/#65, accepted Runtime #25,
related State/Discovery/Composition evidence, accepted REFDES/API/Slices/
Verification/MigrationMap/FindingDisposition, CwdAcquisition, WindowsBroker,
Runtime RTF-02, frozen Application--Runtime/BoundaryTypes and related BCs govern
this contribution. Actual accepted M2 API/ports/Domain and M1 helper interface
are consumed directly. Prototypes remain design inputs, never product evidence.

Owned edits are `internal/runtime/` and this report only. No legacy, concrete
adapter, Application/State/View/Composition/cmd, module, workflow, shared SDP or
frozen boundary edits are included. The current source adds no public native
handle, process-number, callback or private-protocol DTO.

## Checkpoint 1: registry and memory transports

Implementation files: registry.go, events.go, output.go, input.go, errors.go.
Tests: registry_test.go, events_test.go, transport_test.go; directory README.

The private registry atomically gates membership, IDs and final reservations;
uses one ID namespace and refuses wrap; retains admitted failed-start identities;
protects all live/cleanup-pending records; bounds cleaned history to 256. Original
M2 invocation values and the resolved environment have independent copy ownership.
Admission closure captures Starting/CleanupFailed records as well as Running.
No native calls or waits occur under these memory locks.

The output ring preserves raw byte/stream/offset/sequence facts and precise gaps
within 256 KiB, including an append larger than its capacity and explicit uint64
overflow refusal. Bounded copied input retains in-flight capacity and separates
queue ownership from delivery. Reliable finals retain independent reservation
and replay ownership until cumulative ACK; unknown/future/regressing cursors
refuse. Hint delivery receipts are capped at 64; further hints coalesce until ACK
while finals remain deliverable. Event sequence positions for pending finals are
reserved before admission. Closed-producer EOF still requires every final ACK.

These are actual in-memory implementations, not a claimed complete Sessions
adapter. Native lifecycle/OS truth is not implied by supplying a structurally
valid snapshot to the private publication primitive. No public Start/Stop or
placeholder native success path is exposed at this checkpoint.

| #65 clause / verification | Current source and evidence | Remaining proof |
|---|---|---|
| IDs/admission/spec/environment/history; V-DOM-03/V-RUN-01/06/07 | registry.go; concurrent admission/shutdown, exhaustion/copy/cleaned-history tests | Actual admitted native startup/failure ownership and restart concurrency |
| Output/input; V-RUN-05/06 | output.go/input.go; seeded randomized byte reference, streams/sequences, large/single-byte overflow, copied concurrent whole-request capacity and in-flight close accounting | Native blocked reader/writer drains, delivered-count diagnostics and joins |
| Reliable events; V-RUN-06/07 | events.go; full 256 finals, replay/ACK/cancel/misuse, hint-receipt bound and numerical final reservation | Native producers, complete shutdown residuals, later Application transfer/ACK integration |
| All 12 Sessions methods and lifecycle barriers; V-RUN-01..07 | NOT IMPLEMENTED at this checkpoint | Complete Start/Snapshot/List/ReadOutput/Write/Resize/Interrupt/Stop/Restart/NextEvent/AckEvents/Shutdown and all typed effects/results |
| Unix native cwd/SID/PTY/helper acquisition/cleanup; V-RUN-01/04/05/07/08 | NOT IMPLEMENTED / NOT RUN | Retained cwd Fstat/Fchdir, same-executable private supervisor, acquired group helpers, recensus/quiescent-release and all joins; Linux/macOS/FreeBSD native proof |
| Windows broker/Jobs/cwd/debug/ConPTY/architecture/extraction; V-RUN-01/03/05/07/08 | NOT IMPLEMENTED / NOT RUN | Full frozen protocol, all native profiles/failure unwinds and Windows amd64/ARM64/WOW64/emulation execution |
| Embedded helpers; V-RUN-08/V-COMP-03/04 | M1 prerequisite FAIL as expected for partial source | Actual broker/entry/helpergen source closure, both images, manifest and independent canonical rebuild/check without rewriting |
| Full Slices/new stack/release | NOT RUN | Independent final review, prior-adapter gates, serial integration, M4..M8 and all native/public-asset gates |

## Verification and limitations

Local builder: Windows/amd64 Go1.25.0 from
`C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe`.
On the checkpoint source, `go test ./internal/runtime -count=1`,
`go test -race ./internal/runtime -count=1` and `go vet ./internal/runtime` PASS.
All twelve `go test -c ./internal/runtime` target selections PASS with CGO0:
Darwin amd64/arm64; FreeBSD386/amd64/arm64; Linux386/amd64/arm/arm64;
Windows386/amd64/arm64. These are compile-only test binaries, retained locally at
`C:/Users/hanse/AppData/Local/Temp/gh-tree-runtime-build-d8f88aa458754a688aa4d4bbabf61e3d`.
Tests create no native fixture processes and perform no project script execution.
The tests exercise only memory-level responsibilities and do not claim native
Runtime conformance, all-platform execution or completed V-RUN gates.

`go run ./internal/composition/architecture -runtime-prerequisite` returns nonzero
with `M3 prerequisite missing: internal/runtime/broker Go source`. This is a real
mandatory failure, not a passing skip or waived gate. Full workflow acceptance is
therefore unavailable until the helper contribution is complete. No fake images,
no-op verifier, runtime compiler/download or public helper release asset is added.

Separate fresh independent review remains required at the final exact source,
with actual generated assets and native evidence. Author summaries are not review
evidence. Master alone determines contribution/Slice completion and integration.

## Checkpoint 2: private framing and native Unix cwd acquisition

Preceding checkpoint source `14cba902364977967756cca89a2c3b41105624f8` is committed
and pushed. The first source push correctly activates the mandatory helper gate;
it does not create acceptance eligibility while helper prerequisites are missing.

Added broker/protocol.go and start.go: 64 KiB capped length-prefixed binary frames,
fixed protocol/version/role/opcode/session/nonce/sequence binding, exact copied
native startup payloads, one decoder per direction, complete serialized writes,
and poisoned connections after malformed/replayed/foreign frames or partial I/O
failure. Native endpoint/parent validation remains the future owner requirement.
Original SourceVersion/Invocation remain in the sole parent registry; native wire
records carry exact object/argv/environment facts without decoding opaque source
tokens or modifying the shared API. The broker dependency never imports parent
registry or embedded assets.

Added broker/cwd_unix.go and native stamp selections: independent retained
Openat/O_DIRECTORY/O_NOFOLLOW/O_CLOEXEC root/project chain, actual Fstat identity,
full no-follow Fstatat parent-entry recensus, original object authority and observed
relocation refusal. Partial acquisition returns the retained cleanup owner. No
main-process cwd change, path-check/reopen or fd-pathname substitution occurs.
Directory identity follows the current M3 convention: little-endian inode in the
16-byte ID, device, proved birth stamp when available, otherwise explicit change
profile; matching an existing supplied profile never silently chooses another.

Local `go test ./internal/runtime/... -count=1` and matching `-race` PASS on
Windows/amd64 Go1.25.0. All twelve broker test-binary cross-compiles with CGO0 PASS;
new native Unix code is selected on all nine Unix targets. Native execution of
the actual Linux/amd64 broker test binary under existing openSUSE-Leap-15.5 WSL2,
kernel6.18.33.2-microsoft-standard-WSL2, UID/GID65534 PASS. The full broker suite
includes acquired-original marker reads after replacement, stale identity,
symlink/root-redirect refusal, replacement at root/project acquisition barriers,
partial-close ownership, unknown stamp rejection, protocol truncation/unknown/
trailing fields, nonce/role/session/replay rejection, concurrent short frame writes
and failed-transport replay refusal. No native user process or project script is
launched. No native macOS/FreeBSD execution or Fchdir/PTY/process cleanup proof is
claimed. Artifacts remain in the existing local build directory and native owned
`/tmp/gh-tree-runtime-native.XfyRYc` (test-created child directories cleaned).

One initial inline WSL harness command failed during copy with `/broker.test`
permission denied before tests executed; Windows argument transport did not retain
the intended shell variable. A plain UTF-8 file-backed harness corrected argument
transport and produced the passing ordinary-user run. No elevated retry, global
tooling/config change or product-policy change occurred. These are disclosed
harness outcomes, not native product failure/acceptance.

Windows native broker/entry/helpergen/images/manifest remain absent, so the helper
gate remains FAIL. The complete twelve Sessions methods, native Unix supervisor/
helpers, Windows startup/debug/Jobs/ConPTY/extraction, native platforms and final
independent review/integration still remain in full scope.

## Checkpoint 3: acquired-group native helpers and census

Checkpoint2 is committed/pushed `46fe59c5eae3f08df6ebbcc7bcd41536f3fd6864`.
Master subsequently authorized parallel Windows implementation under #69,
canonical `ddfc23f` / ledger78. That worker owns broker Windows files, broker/cmd,
brokerassets and helpergen in its separate worktree. This worker retains Unix
owners and parent registry/Sessions assembly. Shared protocol.go/start.go/common
tests remain exactly at checkpoint2; no Windows helper source-closure input was
changed by checkpoint3. No cross-branch merge bypasses separate review/Root gates.

New Unix inherited-pipe checks validate FIFO kind and read/write direction, set
nonblocking before os.NewFile so Go deadlines actually work, and mark descriptors
close-on-exec. The signal helper validates actual parent/SID/group after the
kernel's setpgid-before-exec setup, accepts only the closed TERM/STOP/KILL set,
sends authenticated Joined, then requires bound Commit before its own `kill(0)`.
It never changes group/session after exec or signals a census numeric target.
STOP stays parked; KILL expects actual signal termination without a fabricated
post-syscall response. The acquisition owner retains one exact waiter and partial
endpoints; cancellation before Commit cannot signal and partial transport outcome
cannot authorize replay. Unknown/foreign/replayed control refuses.

Linux census strictly parses bounded /proc identities/state/session/group;
Darwin selects bounded native sysctl using pinned x/sys KinfoProc and Getsid plus
same-process identity/group recheck; FreeBSD retains the exact selected bounded
/bin/ps profile and joins its own observer before excluding that observer's
already-exited snapshot row. Reserved supervisor-group outsiders refuse. Darwin's
bounded wrapper avoids x/sys table allocation-before-bound and unlimited ENOMEM
retry by checking reported length first and allowing three attempts. Its MIB and
state constants follow Apple's primary
[sysctl.h](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/sysctl.h)
and [proc.h](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/proc.h).
Native macOS execution is still mandatory; source/layout evidence does not prove it.

Actual Linux/amd64 execution as UID/GID65534 on the same WSL/ext4 fixture environment
PASS: foreign group refusal with independent owned control still alive; EOF before
Commit; wrong nonce with no member signal; last original group member exits after
Joined while helper pins membership; STOP helper remains parked; a new KILL helper
joins the stopped group and all member/helper waiters join; complete SID census
contains only the owning supervisor fixture. Private entry without endpoints
refuses. Tests create only their own fixture executables/processes/directories,
with failure-only direct-child teardown that makes the test fail if needed.
No user project script or unrelated process is affected. Final passing owned native
directory: `/tmp/gh-tree-runtime-native.r2Ue02`; prior passing native helper run
directory: `/tmp/gh-tree-runtime-native.p9LSq0`.

All nine Unix broker test-binary cross-build selections PASS. The first bounded
Darwin wrapper compilation revealed x/sys omits KERN_PROC/KERN_PROC_ALL names;
the official header values were checked and both corrected Darwin architectures
compile PASS. Linux native tests were rerun after fixture teardown hardening and
PASS. macOS/FreeBSD native behavior, full supervision/Fchdir/PTY/root-exit lifecycle,
escape/helper failure residual handling and all twelve Sessions methods remain
uncompleted. Expected helper prerequisite failure remains; no native acceptance
or completed Runtime/Slice claim.

## Native FreeBSD correction after checkpoint 3

Exact source da295696 CI run34067206261 is FAILURE: target inventory correctly
fails missing helper inputs, and actual Native FreeBSD job101577954465 separately
fails TestNativeAcquiredSignalHelpers. Race detector and Windows amd64/ARM64,
Ubuntu and macOS native suite jobs PASS. The cross-build/helper jobs skip behind
inventory. This is not a passing acceptance run.

Downloaded actual artifact `freebsd-native-da2956960cbae22d7e4735b6a20896cd018b1bf5-1`,
artifact9999340169 (54,211 bytes), from
[run34067206261](https://github.com/Hans-Einar/gh-tree/actions/runs/34067206261).
Its tests.jsonl SHA256 is
`BCA31839C91923C8EC691BA7A41222DEC0A6D68070C540CD4E45BCA69141D2D1`.
Lines2207..2211 show FreeBSD UID/GID1001, passing foreign-SID refusal, then EOF
before the first valid own-SID helper reaches Joined. The intentional infrastructure
TestExpectedFailure is separate and is not the Runtime failure.

Source-backed cause: FreeBSD15 kern_pipe initializes both anonymous endpoints
FREAD|FWRITE. The original strict O_RDONLY/O_WRONLY gate incorrectly rejects that
native duplex profile. The correction uses a platform-specific endpoint rule:
FreeBSD requires O_RDWR and non-vnode pipe evidence (Fstatfs EINVAL); this refuses
named and unlinked named FIFOs. Linux/Darwin retain their actual directional
profile. Nonce, role, parent, own SID/group, setpgid-before-exec and Joined/Commit
checks remain mandatory. Primary sources:
[kern_pipe](https://github.com/freebsd/freebsd-src/blob/releng/15.0/sys/kern/sys_pipe.c)
and [kern_fstatfs/getvnode_path](https://github.com/freebsd/freebsd-src/blob/releng/15.0/sys/kern/vfs_syscalls.c).

Added native actual pipe flags/Fstatfs/poller tests on all Unix platforms and a
FreeBSD named/unlinked FIFO negative. Corrected three FreeBSD cross-compiles and
Linux native full broker tests as UID/GID65534 PASS locally; actual corrected
FreeBSD CI remains pending, so the diagnosis is not yet a claimed native resolution.
The local source also has separately unstaged Unix supervisor/tree work in
progress (not executed by these tests and not part of this narrow fix commit).
Final exact committed-source CI must confirm the correction. Downloaded evidence
is retained under the existing local build directory, `freebsd-da29569/`.

Corrected actual FreeBSD job101581052093 in run34068364364 at exact a6e79eaa is
SUCCESS. Artifact9999688515 (55,128 bytes), tests.jsonl SHA256
`A0F0A28F837B74AFF156AED9F74DC025EE401AA52F3AB0CF3ED346468B74469B`, confirms both
anonymous endpoints access=2/Fstatfs EINVAL, named/unlinked FIFO negative PASS,
all six acquired-helper control/group/wait/census assertions PASS. Thus the
profile diagnosis/correction now has native evidence. The aggregate still has
the mandatory expected helper-prerequisite failure and is not an acceptance run.

## Checkpoint 4: Unix SID supervisor and resource-barrier engine

Added actual supervisor/tree native code and tests. The private same-executable
supervisor authenticates inherited endpoints and actual parent/SID, Fstats the
designated acquired cwd descriptor, Fchdirs once, closes that descriptor, and
only then resolves argv/PATH with copied environment and observed PWD. User root
inherits cwd/standard streams without private descriptors and starts in a separate
PGID; terminal roots obtain the foreground group. A single native owner watches
root exit, parent control/EOF and complete census. Cleanup sends acquired TERM,
parks acquired STOP helpers, performs full stopped-member recensus, then acquired
KILL; complete root/helper waits and no live SID members precede Quiescent.
Release is required before normal supervisor exit while parent output drains.
Unknown/escaped/unacquirable membership remains failure, never numeric recovery.

Actual ordinary-user Linux source tests PASS for native acquired-cwd marker access,
root exiting before both child and grandchild in separate groups, and a real
isolated interactive shell with foreground sleep/background pipeline groups,
100x30 PTY resize, input/output and explicit quiescence/release/reader/wait joins.
Final local run directory `/tmp/gh-tree-runtime-native.rLJzHI`. All nine selected
Unix broker test-binary cross-compiles PASS; Linux-selected broker vet PASS. The
full supervisor source still needs native macOS/FreeBSD execution and adverse
startup/helper/EOF/escape/unwind/race verification at committed source.

The first new natural-exit test failed because its concurrent fixture collector
embedded bytes.Buffer, promoting ReaderFrom and bypassing its synchronized Write
method. Private collector now contains the buffer as a field; actual marker
content and all three native lifecycle fixtures PASS after correction. Other
fixture failures are not counted as native passes. No project script or user
process was invoked/manipulated. The broader pre-existing helper tests continue
to verify control rejection and actual acquired signaling.

This remains a partial native component milestone: parent Unix client resource
unwinds/controls, Runtime Sessions API assembly, complete result/diagnostic facts,
construction-injected budgets and the full failure/platform gates remain. Raw
supervisor entry codes/failure frames still need truthful parent normalization.
The test harness executes actual native code but does not replace the missing
production parent ownership path. Windows #69 and assets #70 remain separately
owned/reviewed. No shared protocol/start source or Windows helper closure changed.

Checkpoint4 exact source6be865b CI34069098071 is terminal with six independent
native/race jobs SUCCESS (including actual Linux/macOS/FreeBSD package execution).
The three new native supervisor tests select linux/darwin/freebsd and contain no
Skip/GOOS bypass; this is native execution evidence for those exact controls.
The expected missing-helper inventory failure remains, and dependent build/helper
jobs skip. It is not complete Runtime or full-platform fault acceptance.

## Checkpoint 5: concrete Unix parent client seam

Ownership split under queued #71 is now authoritative: this worker edits only
Unix broker/client/native hardening and this report. Root Runtime Go/tests/README
remain at the common6be865b base for the fresh #71 worker. Windows #69 and assets
#70 are independent contributions; no unreviewed branch integration occurred.

Added real `broker/client_unix.go` with these exact private signatures:

```go
StartUnix(context.Context, UnixConfig) (*UnixClient, UnixStartResult, error)
(*UnixClient).Write(context.Context, []byte) (UnixDelivery, error)
(*UnixClient).Resize(context.Context, uint16, uint16) (UnixDelivery, error)
(*UnixClient).Interrupt(context.Context) (UnixDelivery, error)
(*UnixClient).Stop()
(*UnixClient).NextFact(context.Context) (UnixFact, error)
(*UnixClient).Wait(context.Context) (UnixFact, error)
RunUnixPrivate() (bool, int)
```

UnixConfig carries SessionID, StartSpec, copied-output callback and construction
GracePeriod/ForcePeriod. UnixStartResult has Established/Cwd. UnixDelivery has
Accepted/Delivered uint32 and Completed. UnixFact has Established/RootExited,
ExitCode/Signal int, Quiescent/CleanupComplete, Err and copied []UnixResidual;
each residual has api.RuntimeCleanupStage and Err. UnixFailure carries safe closed
api.ErrorCode/cleanup stage, preserving NotFound/Permission/StaleObservation etc.
No public SessionID allocation, registry, native handle or Application operation
authority is introduced by this native seam.

The retained client owns acquired cwd, exact supervisor, all pipes/PTY handles,
independent output workers, decoder and native I/O joins. Failed post-acquisition
construction returns its owner. Start remains Established after immediate exit
or later initiating-context cancellation. Stop is an asynchronous coalesced latch;
Wait timeout leaves ownership active. Input returns actual accepted/delivered
counts, copies bytes, joins cancellation watchers, refuses late controls and never
replays partial writes. The configured native budgets use a Unix-only validated
Stop payload, leaving shared protocol.go/start.go and Windows source closure
unchanged. The early dispatch checks only fixed markers and delegates to existing
authenticated supervisor/helper code before any normal application bootstrap.

PTY master gets a separately owned duplicate, made nonblocking before os.NewFile,
so Fd-based resize ioctls retain poller/deadline/Close behavior. The original master
and slave both retain immediate cleanup ownership. Remote startup/census failure
reports are distinguished from local close/join residuals: a later complete native
barrier clears repaired reported uncertainty while historical diagnostics remain;
actual local close/wait errors cannot be accidentally cleared.

Actual ordinary-user Linux full broker tests PASS at the new source, including
natural exit/cwd/output with late Start-context cancellation, persistent-root late
cancel, all five exposed acquisition checkpoints with retained owners, full64KiB
input checksum, bounded partial writes and cancellation, Stop unblocking an owned
writer, late control/capability refusal, typed failed executable startup and cleanup,
and real PTY resize/input/foreground ETX distinct from whole-session Stop. Earlier
supervisor/helper/cwd/protocol controls also PASS. Final local native directory:
`/tmp/gh-tree-runtime-native.sLv3KU`. All nine Unix client test-binary cross-compiles,
Linux-selected broker vet and staged architecture checks PASS.

Two initial client regressions were exposed and corrected: a native write deadline
could fire just before context.Err published, losing errors.Is DeadlineExceeded;
and a repaired failed-start ProcessContainment report was retained as a resource
residual after complete cleanup. Corrections preserve known byte delivery and
separate reported versus local residual ownership. The FreeBSD observer writer
also no longer embeds bytes.Buffer: promoted ReaderFrom could bypass its byte
ceiling. A dedicated io.Copy negative prevents that regression; corrected native
FreeBSD execution of this new test remains pending at the new commit.

This is a coherent, callable native-client checkpoint, not final #65 acceptance.
Remaining Unix-owned work includes comprehensive partial native failure seams,
callback/close blocking and timeout residual precision, parent-control failure,
observed escape/unacquirable/helper failure recovery, repeated resource counts,
and new-client exact native macOS/FreeBSD/race evidence. #71 separately owns all
twelve public Sessions methods, parent input/event/restart/shutdown orchestration
and final real-client assembly. No shortened adapter or public stub is supplied.

## Checkpoint 6: client residual precision and native CI corrections

Exact9a73ffc CI34071112204 fails native Linux/macOS/FreeBSD and race, independently
of the expected inventory failure. Inspected raw Linux job101588512087 and race
job101588512130 logs plus actual FreeBSD artifact10000508196 (57,180 bytes),
tests.jsonl SHA256 `EC0344F9B89767704FBED41F700E6E1F431A9866E03B25760B197D5AE2152D73`.
The PTY assertion failed after size27 92 and echoed input; one race output already
contained the executed marker after bracketed-paste ANSI. The fixture had sent ETX
before proving the long-lived foreground command had started, and incorrectly
required a particular surrounding CR/LF/ANSI shape.

PTY test now runs a newly owned same-binary foreground fixture, waits for its
non-echoable readiness marker and an actual changed terminal foreground group,
then sends ETX. The followup marker is assembled by printf, so echoed source does
not contain the expected full bytes; detection does not interpret/strip terminal
ANSI. Product output remains raw. This proves delivery/observed fixture response
without claiming every arbitrary child must act on ETX.

Race failures in short native cleanup tests correspond to the instrumented
helper's default one-second atexit sleep, which exceeds the injected short fixture
budgets. Native test fixtures preserve existing GORACE options and append
`atexit_sleep_ms=0` for their own subprocesses. Race instrumentation/reporting
remains enabled; no tests are skipped and product2s/3s reporting defaults are
unchanged. The default/meaning is documented by the primary
[Go race-detector documentation](https://go.dev/doc/articles/race_detector).
Corrected native race CI is still required before claiming these failures resolved.

Client facts now include all still-unproved parent/native barriers while waiting,
not just recorded failures. Atomic closed-file/callback/I/O/decoder/cwd facts
prevent observations racing native close. Output callback panic is retained as a
safe diagnostic and triggers owned cleanup. An injected blocked output owner
remains a typed OutputCleanup residual through timeout; after its actual join,
the repaired pending residual clears while the historical timeout remains. Local
close errors are kept distinct and cannot be cleared by this repair.

Actual ordinary-user Linux full broker suite PASS, including new blocked-output,
late-join and callback-panic controls and the corrected foreground/ETX fixture;
native directory `/tmp/gh-tree-runtime-native.99ZprM`. All nine Unix cross-compiles
and Linux-selected vet PASS. Actual new-head native/race CI remains pending.
Private client types/signatures and shared protocol/start source are unchanged.
Remaining native failure/resource-count coverage and independent review stay
separate from #71's twelve-method parent implementation.

Exact0dfd887 CI34072100039 now confirms corrected FreeBSD/Linux/race and both
Windows suite jobs PASS. macOS job101591203418 has a separate helper failure:
foreign SID, pre-Commit close and wrong nonce controls pass, followed by
ErrProtocol in the last-original-member-after-Joined KILL case. Inspected actual
job output; this is not the repaired terminal assertion failure.

## Checkpoint 7: KILL helper lifetime

After self-group KILL returns, the helper now remains parked until native signal
termination, matching the existing STOP lifetime rule. It cannot voluntarily
exit0 before deferred self-signal handling, which would defeat the required exact
SIGKILL waiter evidence on Darwin. Failed termination evidence now reports the
actual observed helper exit class. No numerical recovery/fallback is introduced.
Added vanished-before-acquisition control, allowing only the documented harmless
case in which the newly allocated helper PID itself equals the vanished group
number. Whole-session census and all joins remain the final cleanup authority.
Ordinary-user Linux full broker suite PASS at the correction, native directory
`/tmp/gh-tree-runtime-native.QQUAiQ`; actual corrected macOS CI remains required.

## Next permitted work

Freeze/push the Unix candidate below, obtain new-head native/race CI and a fresh
independent source/test/contract review, correct actual findings and re-review.
#71 owns parent Sessions/root files, #69 owns Windows native code and #70 owns
helper assets; their reviewed integration remains Master-owned. Preserve every
native gate. This worker cannot accept/integrate the contribution or close #65.
Request Master coordination before any shared-interface/workflow/private-host
dispatch change or frozen-contract gap. Preserve every native requirement and
all failed/not-run verification facts until exact-source evidence resolves them.

## Unix native candidate: adversarial ownership and recovery

Prior exact a57d496 CI34072812802 passes all six independent native/race jobs,
including macOS corrected KILL lifetime. Only the known helper-input inventory
failure remains; this is not all-target/helper conformance or full Runtime acceptance.

The candidate adds actual native parent-control EOF and observed setsid escape
tests: both retain explicit residuals and never fabricate full cleanup; their
parent handles/workers close/join. The escape fixture itself owns and cleans its
deliberately escaped direct child; product code performs no numeric recovery.
Cancellation before acquisition and at all five lifecycle acquisition barriers
retains correct owner/no-owner facts. A user-root fd probe proves private pipe/cwd
descriptors do not reach it; an intentionally inherited fixture fd3 is the positive
detector control. Existing private-entry, frame/nonce and group-acquisition controls
remain in place.

Native acquisition injection now covers17 checkpoints: cwd, each control/input/
stdout pipe stage, PTY pair/master transfer, I/O setup, all five child endpoint
closures, both output-reader starts, supervisor creation and Start transfer. Each
returns a retained owner and proves actual cleanup. The helper recovery fixture
uses an exclusively created copy of the test executable; removing execute permission
causes a typed Permission residual while ownership remains active, restoring that
owned copy allows cleanup, and historical permission facts survive. It changes
neither the installed executable nor user project code. This test exposed joined
error classification using os.IsPermission; errors.Is now correctly preserves
permission/not-found facts through aggregate cleanup errors.

Twelve alternating pipe/PTY cycles retain their closed client objects and verify
every registered native descriptor is closed/unusable, all private owner counters
are joined, and goroutine counts return. Linux additionally verifies process fd
inventory6→6; local evidence has102 closed owned descriptor objects and goroutines
2→2. The same repeated ownership/worker test runs on macOS/FreeBSD; their global
process fd inventory is not asserted from an unavailable Linux proc filesystem.

Final local ordinary-user Linux full broker suite PASS after all candidate changes,
directory `/tmp/gh-tree-runtime-native.QsL9z5`. All nine Unix test-binary cross-builds,
Linux-selected vet and staged architecture PASS. New candidate exact native CI
remains pending. Prior full source/native/race results do not substitute for it.

| Unix-native obligation | Actual source/test evidence | Remaining acceptance scope |
|---|---|---|
| Independent acquired cwd and before-user Fchdir | cwd_unix/stamp variants; supervisor; original-object, link/stale/barrier-replacement tests and actual marker fixture | Fresh review and new-head native gate; no continuous ancestry/sandbox promise |
| SID/controlling tty/different user PGID | supervisor/tree; native root-before-child/grandchild and foreground/background PTY fixtures | Fresh review; full parent Sessions integration belongs #71 |
| Joined/Commit/acquired self-signals and identity lifetime | signal/endpoints/census variants; foreign/vanished/last-original/nonce/parked STOP/KILL fixtures | Fresh review; global numerical PID recycling is not forced by tests |
| Startup/partial ownership and copied transports | client;17 acquisition fault points, cancellation, checksum/partial input/late controls and typed failed startup | Fresh review and final native fault selection; malformed/unsupported native profiles always refuse |
| Complete release/joins and truthful residuals | client/tree/supervisor; EOF/escape, blocked writer/output, callback panic, helper failure/recovery and repeated resource counts | Fresh review; deliberate unobserved new-session escape is outside portable SID containment |
| Native Linux/macOS/FreeBSD execution | Exact prior CI records above and candidate fixtures without platform bypass | Candidate new-head CI, independent review and eventual integrated SHA |
| Twelve Sessions methods/event ACK/restart/aggregate shutdown | Existing common root foundations, queued #71 | #71 real-client assembly and review, not completed by this Unix native candidate |
| Windows broker/images/reproducibility | Independently owned #69/#70 | Separate accepted source/assets and Master integration; this branch still fails missing-helper inventory |

No native/resource correctness limitation is hidden as a passing skip. Parent
composition/private dispatch, Windows/image integration, complete public lifecycle
and full Slice/M3/native-release proof remain gated outside this isolated candidate.
