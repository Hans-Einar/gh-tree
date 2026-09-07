# M3 Windows Runtime — #69 under #65/#21

Role: bounded Windows Worker. Branch `codex/cr21-runtime-windows`; base
`46fe59c5eae3f08df6ebbcc7bcd41536f3fd6864`. Ownership is the #69 filename/subfolder
split within Runtime. Frozen Application--Runtime/BoundaryTypes 1.0.0,
WindowsBroker/CwdAcquisition and Verification govern this contribution. No
acceptance, integration, full adapter or native platform completion is claimed.

Current disposition: **W69-M01/M02 corrected source frozen for bounded independent
re-review; current native CI complete**. Latest candidate status and dependent gates are
at the end of this record; earlier partial checkpoints are historical evidence.

## Partial checkpoint 1

New Windows-only broker files establish retained native Jobs, wait/identity
primitives, handle-relative no-reparse cwd acquisition, real read/list ancestor
guards, exclusive delete-on-close data-read anchors and revalidation. The private
debug owner implements pinned ABI offset checks, native/WOW64 event sequencing,
read-only PEB cwd-handle duplication/full FileIdInfo comparison and pending-event
detach. This debug code is not yet exercised by a live user-root launcher.

Native Windows amd64 Go1.25.0 `go test ./internal/runtime/broker -count=1 -v`
passes: existing wire/start tests; rename/delete interlocks; stale object and
native mount-point reparse refusal; native layout/routing. The first reparse
fixture used symlink creation and failed with missing privilege; it was replaced
by an owned mount-point FSCTL fixture which needs no privilege. No test skips.

Remaining: live native startup engine and full cleanup/control/output/client,
ConPTY and exact fixture process tests, embedded source/build/extraction,
ARM64/emulation/native32 evidence, failure/resource/adverse controls, all twelve
compile gates and independent review. The M1 helper prerequisite remains failing
until its real inputs/checker exist. Unknown x64-on-ARM64 startup is explicitly
refused pending its verified ABI profile; this is an unfinished required gate,
not permission to remove that target. Public parent Sessions assembly stays with
the original Runtime author. Shared wire/start files remain unchanged.

Next permitted action: continue owned native implementation in tested/pushed
partial checkpoints; obtain separate fresh Windows review before Master alone
integrates into the Runtime branch. No product PR, merge or release.

## Partial checkpoint 2

The live native user launcher now exercises acquired cwd, inner Job assignment
before resume, the complete debug cwd barrier and pending-event detach. Native
amd64 fixtures pass pipes, immediate user chdir, literal argv (including empty,
quotes, trailing slash, Unicode and shell metacharacters), no user initialization
before the barrier, no debugger/anchor visible afterward, real ConPTY input and
100x30 resize, root wait and terminal/output joining. The initial test fixture
omitted the testing flag terminator and exited 2; this was corrected. Native
tests also exposed premature closure of system-owned debug process/thread event
handles; only image/DLL handles are now explicitly closed, while the separately
returned CreateProcess handles retain Runtime ownership. Microsoft documents
event-handle closure in ContinueDebugEvent; the corrected tests pass.

`broker/cmd/main_windows.go` now calls the fixed real
`broker.RunWindowsPrivate() int` entry. The authenticated bounded engine skeleton
is wired to the actual native implementation, including output transfer, serialized
input/resize, root-exit cleanup and Quiescent/Release. Its whole parent protocol
still requires the private client and end-to-end tests; entry compilation alone
does not verify that protocol. Parent output transfer pulls a duplicate from its
retained broker process handle before ACK, avoiding unknown allocated remote
handles when a transfer message fails.

Validation: Go1.25.0 Windows amd64 broker tests and vet. Build/generator source
closure remains incomplete and will change during the remaining implementation.

## Partial checkpoint 3

The Windows-only private client now retains the suspended broker, assigns its
outer kill-on-close Job before resume, authenticates inherited anonymous control
endpoints plus a retained parent-process capability, receives/ACKs kernel output
handle transfers, drains copied output independently, and joins broker/outer
membership/control/output ownership before its final cleanup fact. No outer
membership-one test precedes Release. Its bounded methods and lifecycle facts
will be adapted by the original Runtime author; no public Sessions assembly is
changed. Role-local Stop payloads permit bounded test grace/force periods without
editing shared wire or StartSpec.

Real amd64 broker-client tests pass pipes, ConPTY input/resize, forced stop with
an owned child, and root-first exit with a surviving child. All four repeated
three times passed; full broker tests and vet subsequently pass. The first forced
stop exposed TerminateProcess racing an already-terminating Job member; cleanup
now accepts the exact signaled process wait as proof of completion and retains
termination errors if that proof fails. This does not weaken Job-zero evidence.

Incomplete gates remain: reviewed helper generation/extraction, native WOW64 and
ARM64/emulation, comprehensive fault/cancel/resources/TLS/DLL/debug-heap controls,
all twelve builds, independent review and parent adapter integration. No M1 helper
or full native acceptance is claimed.

## Partial checkpoint 4

Native extraction now accepts independently supplied pure image bytes, machine,
SHA256 and protocol metadata. It validates PE32+/entry/hash, acquires the entire
temporary-directory chain with handle-relative no-reparse data/list-read handles,
creates an exclusive nonce directory and image under protected current-user and
SYSTEM DACLs, fully writes/flushes, reopens a READ-share-only image guard and checks
the full original FileIdInfo/hash again. The client retains that guard through
broker/outer cleanup, then removes only the exact image and empty owned directory.
Replacement identity and unexpected entries preserve residuals. No broker source
imports brokerassets; the separate #70 worker owns pure generated data/build code.

Native amd64 tests pass protected ACL inspection, write/delete interlocks, exact
cleanup, preservation of an unexpected file, rejection/preservation of a replaced
image, and a real extracted test broker's root-first lifecycle through final
directory removal. Full broker tests pass. Test-only cross-ABI executable builders
now exist for upcoming native386/WOW64/ARM64 controls; their implementation is
excluded from ordinary product startup. They build only this owned test package,
and remove only the exact generated executables/empty fixture directory.

## Partial checkpoint 5

Local full native386/WOW64 suite passes, including386 parent -> extracted native64
test broker -> native64 user and native64 broker ->386 user; native32 direct
startup layout also runs. Native amd64 race suite passes (23.666s), native386
suite passes (16.921s), and vet passes at this checkpoint's source.

Cancellation passes at eight cwd/Job/ConPTY/CreateProcess/assignment/pre-resume
barriers, with no user initialization and joined resources. Missing executable,
stale cwd and parent Start cancellation retain cleanup ownership. Separate
primary-thread checks retain a test proof duplicate that prevents thread-ID reuse,
then assert the actual original CreateProcess thread handle no longer denotes
that owned thread, for user and broker success/canceled startup paths. Aggregate
kernel counts cover only the current test process's File/Pipe, Job and Process
objects, classified through new probe capabilities; they do not claim Thread/Event
pool coverage. Those owned object counts and goroutine counts stay flat across
eight repeated ConPTY cycles. Earlier total-handle comparisons included Go's
expanding scheduler Thread/Event pool and were replaced with explicit type counts;
primary-thread closure remains independently tested, not excluded from proof.

Existing readable children can now serve as preserved data/list-read anchors,
selected by bounded enumeration of the acquired directory handle. Empty/no-usable-
child cases retain exclusive temporary anchors. Native tests prove existing child
preservation and delete interlock.

CI at prior2f388202 exposed SDDL alias `LA` in the ACL test; that test now compares
binary trustees, ACE masks/types/flags and protection. Product extraction also
checks those exact properties after directory creation and read-only image reopen.
This preserves the current-user plus SYSTEM requirement; no alias spelling is
treated as a different trustee. Corrected full native386/race tests pass locally;
the new native Windows CI result is still pending.

Native ARM64/x64-emulation, static DLL/TLS/debug-heap fixtures and remaining adverse
transport/close cases are still required; no complete Windows contribution or
helper-source freeze is claimed. The final helper closure will need regeneration
after these broker changes.

## ABI runner diagnostic checkpoint

`TestWindowsTargetABIMatrix` now exercises every supported target machine on the
actual native host, using only newly built owned test executables. Local native
amd64 ->386 and amd64 ->amd64 both establish and join. On native ARM64 the matrix
also requires386/amd64/ARM64; the currently unsupported amd64-emulation path is
expected to fail until its real loader profile is resolved. A bounded test-only
diagnostic captures loader events/native64 PEB cwd identity at a candidate barrier
and kills its owned Job without approving/releasing user startup. This is evidence
gathering for the missing profile, not a passing product substitute or waived gate.

## Mapped-machine and adverse transport checkpoint

Actual40e7e103 CI34068927027 passed native Windows amd64 job101582615826 and
native ARM64 job101582615856, including the previously expected-to-fail target
matrix. That is positive exact-source evidence, not evidence that the anticipated
failure happened. The next implementation now reads the actual mapped PE machine
through the retained process's PEB image base, and chooses embedded routing from
that image versus the native machine. IsWow64Process2's processMachine alone is
insufficient. Target profiles explicitly distinguish native, x86 WOW64 and x64 on
ARM64, with mapped-machine/initial-breakpoint facts in private Started messages.
Tests assert the compiled PE machine and actual child mapped machine, and run
real emulated parent executables through their own complete matrix. Local native
amd64 and386-parent matrices pass; corrected ARM64 evidence is pending.

Real blocked input now stops and joins with accepted/native-delivered partial
counts reported before Quiescent. A real ConPTY with an explicitly injected
blocking close retains its owner after the reporting deadline, then completes
the actual ClosePseudoConsole and joins after test release. This tests the injected
blocking path, not an assertion that this host runs an older Windows close ABI.
The already recorded native ConPTY lifecycle remains independent positive proof.

The client checks known establishment before responding to a raced Start context
cancellation; an already established session is not stopped/retracted by that
path. Role-private copied typed residuals identify still-unproved cleanup barriers;
full cleanup clears them while retaining historical diagnostics separately.
Full native amd64 broker tests and vet pass before this checkpoint. No helper
freeze, complete native fault matrix, independent implementation acceptance or
parent Sessions integration is claimed.

## Native loader and batch fixtures

Test-only C sources under `broker/cmd/fixtures` compile an owned DLL and executable
using the already installed Visual Studio toolchain in a temporary directory.
Both images have actual TLS callbacks; the static import's DllMain and both TLS
callbacks inspect debugger/anchor visibility and native heap operations. Both
images link the static debug CRT (`/MTd`); explicit debug-heap allocation and
`_CrtCheckMemory` checks run. Native x64 MSVC19.43.34810/link14.43.34810.0 passes:
`NATIVE_DLL_TLS_DEBUG_HEAP exe=0 dll=0`. The first compiler wrapper used ordinary
Go argv quoting for cmd's `/C` carrier and failed before compilation; explicit
reviewable CmdLine quoting corrected the test wrapper. No compiler install/global
environment change or product runtime compilation is introduced. Hosted ARM64
toolchain execution remains a real required gate, with missing tools failing.

The actual reviewed cmd/bat carrier also passes literal empty/whitespace/Unicode,
metacharacter and trailing-backslash arguments through an owned `.cmd` shim into
the native fixture. Quotes, percent and CR/LF operands refuse before execution.
Targeted loader/batch tests pass; these add no product source-closure change.

## Frozen standalone Windows native candidate

The candidate is the commit adding this section; Master records its resolved exact
SHA for independent review and source-closure regeneration. Product code is confined
to Windows-selected broker files and its private entry. Shared wire/start, Unix,
parent registry/Sessions, other layers, module/workflow and frozen governance remain
unchanged. The source dependency graph contains no brokerassets import.

Latest local validation before freeze: Go1.25.0 Windows amd64 full broker race
suite **PASS 41.402s**, native386/WOW64 full broker suite **PASS 31.117s**, broker
vet **PASS**, `git diff --check` **PASS**, and `CGO_ENABLED=0 go build ./...` **PASS
for all twelve release OS/architecture selections**. These builds compile package
selections; they are not native execution or public release packaging proof.

Native CI already establishes earlier material source checkpoints: bed507e
CI34069650439 and e0520a8 CI34069880581 each passed all six independent jobs,
including Windows amd64 and Windows ARM64. The only failure was the expected M1
inventory/helper prerequisite. That proves mapped-machine/target/parent routing
and installed native DLL/TLS/debug-heap execution on those exact earlier sources;
the current frozen candidate still requires its own native CI and fresh review.

| Native requirement | Actual candidate evidence |
|---|---|
| Suspended broker/outer Job and user/inner Job before resume | Client/user hooks inspect actual containment before resume; primary-thread original-handle closure uses retained identity proof duplicates. |
| Real cwd guards/nonempty anchors | Handle-relative acquisition and bounded existing-child pinning; stale/reparse refusal; preserved existing child; metadata-only FSCTL negative control succeeds while real protected startup barriers refuse it. |
| Correct startup ABI/cwd object/pending detach | Compiled and mapped PE machines asserted; native/WOW64/emulated-parent matrix; full duplicated cwd FileIdInfo; no user marker before pending-event guard release; no debugger/anchor in user initialization; immediate user chdir. |
| DLL/TLS/debug heap/batch | Actual DLL static import and DLL/exe TLS callbacks plus static debug CRT/native heap; literal owned cmd shim and unsupported operand refusal. |
| Every partial ownership family | 16 cwd/user/debug acquisition-failure stages, eight parent acquisition-failure stages, seven extraction stages and same-object byte tamper; eight native cancellation stages. All retain an owner and join without user initialization. |
| Root/child/grandchild cleanup and owner failure | Actual three-level fixture, root-first and Stop; retained outer Job proves zero after broker crash, foreign frame and control EOF; actual parent exit closes its last Job handle while separately retained process handles prove all four known tree members exit. |
| Transport/terminal barriers | Real ConPTY input/100x30 resize; blocked input reports accepted/partial native delivery before Quiescent; injected blocked real ConPTY close retains resources at deadline and joins after release; two MiB raw multi-stream output drains without event consumption and callback copies remain stable. |
| Concurrency/residual/resource facts | 32 concurrent Stop callers coalesce; Wait timeout retains typed copied unproved barriers; completed cleanup clears residuals separately from historical errors; repeated ConPTY cycles keep current-process File/Pipe/Job/Process counts and goroutine counts stable. |
| Extraction | Protected binary-trustee ACL checks, PE/hash/protocol, no-reparse full chain, exclusive create/write/flush/read-only guard, exact identity/hash, no-write/no-delete interlocks, replacement/unexpected-entry preservation, and post-outer removal. |

One protocol-EOF case found during adverse testing required a correction: a poisoned
receiver cannot consume Release. The engine now completes owned user cleanup and
reports failure/exits rather than reporting Quiescent and waiting for an unreadable
Release. The parent's outer barrier remains authoritative. Product fault seams are
private and nil in normal construction; tests fail immediately after real native
acquisitions, not by replacing those operations with mocks.

Outstanding dependent gates, not waived: fresh independent native source review;
current-candidate Windows amd64/ARM64 CI; reviewed #70 generator/assets regenerated
against this actual final selected source closure and real committed-image execution
binding; later parent Sessions integration and combined review; full-stack actual
package-manager/Discovery/host scenarios, aggregate registry behavior and final
twelve-asset release/install proof. Local `yarn.cmd` is not installed, so no actual
local yarn integration is claimed. The ordinary native cmd-shim carrier is tested.
No helper source/provenance acceptance can use the older ab608-based #70 bytes as
if they matched this source. Master alone integrates after the required independent
reviews and evidence; this candidate does not close #65/#69, any Slice, or M3.

The working tree is to remain clean and the candidate held while the fresh reviewer
inspects it. Next permitted action is independent review, followed only by bounded
corrections/re-review or Master-coordinated helper binding/integration. The unrelated
blocked Git foundation review has not been retrieved, retried or replaced.

## Bounded W69-M01/M02 correction — 2026-09-07

Fresh correcting author under Master87581558/ledger87 and the latest #69 comment;
starting clean/pushed review HEAD35a7632c97920b68e5c428bbbbf37494a056d8ea. The two
findings and unaffected native evidence remain in M3-Windows-Review--001. A bounded
implementation subagent authored only the failure codec/engine/process/debug
correction and its tests; it performed no independent review or integration.

W69-M01: ordinary controls now await serialization cancellably, then recheck
cancellation and Stop at one atomic memory admission point. At most64 dispatched
but unobserved receipts remain; only one native input is in flight, bounded to
65484 bytes (the parent's65536-byte queue still needs its existing split/join).
No input bytes remain in a completed receipt. A rejected operation returns no
receipt. Every possibly dispatched operation returns a WindowsReceipt through
WindowsDelivery, including an ambiguous send error. Dispatched means possible
effect; Completed alone means a known native result. Receipt.Wait observes the
original immutable result, prefers terminal facts over raced cancellation, and
retires admission exactly once even on error. Cancellation preserves the receipt;
concurrent/repeated observers retain the result. Native cleanup never waits for
receipt consumption. The parent must join retained receipts before its public
final; a completed native-cleanup fact alone does not consume them or authorize
input replay. This private seam was coordinated/approved by Master for #71.

The receive owner drains the ordered broker stream through EOF after process exit
and after send-direction failure, preserving late write/failure replies. Fatal
engine cleanup now joins/reports an in-flight native write before its Failure
frame. Unknown terminal transport outcomes remain explicitly Dispatched with
Completed false, rather than asserted zero effect. Delivered frames validate
request kind, exact accepted length, delivered bounds, status byte and sequence.

W69-M02: a versioned bounded payload carries up to16 closed cause/stage/cleanup
records, with strict enum/count/duplicate/trailing-byte validation. WindowsFailure
preserves errors.Is for stale cwd, not-found, permission, unsupported profile,
process, protocol, cancellation and cleanup semantics. Local typed failures retain
their original cause while safe Error text and wire fields expose no path, argv,
environment, status number or native handle. Original and independent cleanup
failures remain distinct; later cleanup success retains historical diagnostics.
Existing public contracts, shared framing/start, Unix, parent, helper assets,
modules/workflows and independent reviewer files are unchanged.

Coherent corrected source checks before this commit: pinned Go1.25.0 Windows
amd64 full broker race suite PASS47.826s; executing386/WOW64 full broker suite
PASS40.474s; broker vet and diff check PASS. Meaningful new real native controls
cover all three pre-canceled controls, waiting serialization, Stop-winning
admission, completion/cancel priority, blocked-write eventual partial delivery,
broker-death terminal uncertainty, fatal protocol failure during blocked input,
and error after an actual full control-pipe write. The latter two targeted controls
PASS0.877s. Bounded-memory and malformed-payload controls are separately identified
in tests. Failure tests execute stale identity, missing image, execute-denied owned
image (exact retained DACL restored), invalid executable, unsupported command
profile, target-profile refusal before user initialization, and a real ConPTY
whose injected blocked close reports primary plus cleanup failure before joining.
No excluded earlier denied-cleanup residue was touched.

This commit freezes a coherent correction milestone for all-twelve compilation,
exact current-source native CI and the existing independent reviewer's bounded
re-review. No finding is independently closed by this author report. Unaffected
original native loader/DLL/TLS/debug-CRT/outer-Job evidence is reused, and the full
race/386 suites also execute those existing tests. #70's older ab608 images still
do not bind this changed source closure; reviewed regeneration and real committed-
image execution, #71 parent assembly and every full Runtime/Slice/M8 gate remain.
Master alone integrates. No separate Git review was retrieved/retried/substituted.

Final correction evidence: clean/pushed technical source
`bd78deafd4dd36e22d5b106eb7ef9c4edcd2e832`, broker subtree
`04afda252be325beaa6bf1f22c154a094b8daed9`. All twelve local
`CGO_ENABLED=0 go build ./...` selections passed at that exact source after its
commit. This is compilation evidence, separately from native execution.

[CI34073256923](https://github.com/Hans-Einar/gh-tree/actions/runs/34073256923),
attempt1, completed at exact bd78dea with six successful independent jobs:
Windows amd64101594439386, native Windows ARM64101594439292,
Linux101594439330, macOS101594439389, FreeBSD101594439302 and race101594439205.
Windows job logs bind the source and actual Go1.25.0 host architecture; real broker
tests pass54.808s on amd64 and56.583s on ARM64, followed by successful full vet/build.
The run's overall conclusion is FAILURE: inventory101594439384 explicitly reports
`M3 prerequisite missing: internal/runtime/cmd/helpergen Go source`; dependent
cross-build101594535941 and helper-reproducibility101594536162 are SKIPPED.
This standalone-source prerequisite remains separate #70 work and was not waived.

The final report-only commit uses `[skip ci]`, leaves the tested broker subtree
unchanged and must not cancel source CI. Keep the candidate/worktree clean and
stopped for the existing independent reviewer's bounded correction review.
No author acceptance, integration, complete native adapter, Slice or release
disposition is granted by these passing source checks.

## Integrated extraction cleanup investigation — 2026-09-07

Authority: Masterd55d56f0a50e7c27e8c8b15e22e397826f7601f3 / ledger98. Work is on
`codereview-21/layer-runtime`, starting clean/pushed
`bab42e903f94edba3fe35d20187087aa0e2eddb1`; the accepted standalone Windows branch
remains untouched. Windows product source remains bd78dea. Only
`extract_windows_test.go` and this report change; no native product, shared/Unix,
parent, generator, assets, helper closure or integration metadata changes.

Original failure remains attributable to source
`941f4be195bf1a0fefdb13d41e13435e0628bf76`,
[CI34081403751/job101617232026](https://github.com/Hans-Einar/gh-tree/actions/runs/34081403751/job/101617232026),
attempt1: native Go1.25.0 Windows ARM64 / imagewin11-arm64 20260830.155.1 /
Windows10.0.26200. Nineteen other jobs passed. TestWindowsExtractedBrokerLifecycle
failed before inspecting object absence: Established, RootExited/exit0, Quiescent
and CleanupComplete were true, Residuals empty, with historical
`Windows runtime permission denied at helper extraction during cleanup`.
The original raw UTF-8 job log is23466 bytes, SHA256
`dee16351e09ab2d29b1b68d05120c1928b73ef22a2daf3517f25286589a75485`.
Its underlying NTSTATUS was not logged; no claim about its external trigger or
antivirus is supported by that historical log.

Additive local diagnostics reproduced the same natural lifecycle failure on the
first amd64 run and identified actual `STATUS_CANNOT_DELETE`0xc0000121, mapped
Win32ERROR_ACCESS_DENIED5. That initial diagnostic run intentionally still failed
the original nil-error assertion (exit1,2.177s); its independent native proofs
nevertheless established broker exit, outer Job zero, all client/extraction
handles released, and exact image plus directory absent. The owner had repaired
a refused cleanup attempt and correctly retained its historical diagnostic.

A deterministic owned SEC_IMAGE mapped-view control now forces that same native
refusal through real extraction cleanup after root exit/broker exit/outer Job0.
While the mapped view remains, cleanup is incomplete and the exact image remains.
After releasing only the test-owned view/section, the ordinary owner completes
cleanup and preserves the original typed cause. Separate broker/process and Job
proof duplicates remain open through successful deletion, so those process-handle
references are not demonstrated to be the blocking resource. The control does
not infer the historical hosted runner's exact source of a retained image view.
[Microsoft's mapping profile](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createfilemappinga)
defines SEC_IMAGE; the native controlled result is the implementation evidence.

The lifecycle assertion now checks final root/start/quiescence facts, exact broker
wait and outer Job-zero through retained proof capabilities, joined client owner,
closed client/child/output/image/chain handles, and actual absence of both exact
objects before considering an error. It accepts only an entirely matching error
tree: WindowsFailure Permission / HelperExtraction / Cleanup, whose original
native leaf is STATUS_CANNOT_DELETE. Generic Win32/NT access denial, cancellation,
timeouts, wrong stage, non-cleanup errors and any mixed process/other error fail.
The historical error is retained and logged, not erased or normalized to success.
Negative diagnostic controls and the mapped-view refusal/release/repair control
exercise this narrow rule. No product defect or product change is established.

Pinned Go1.25.0 targeted selection
`TestWindows(ExtractionMappedImageCleanupRepair|ExtractedBrokerLifecycle|RepairedExtractionDiagnosticIsNarrow)$`:
amd64 PASS1.724s, final amd64 race PASS5.244s, executing386/WOW64 PASS3.016s;
broker vet and diff check PASS. The 386 natural lifecycle independently observed
the same0xc0000121 historical refusal with all final proofs passing. Test failure
teardown releases only its own mapping and joins the native owner before existing
exact cleanup runs. Earlier excluded ACL/cache residues remain untouched.

Next: push this coherent test-only refinement, run exact integrated native CI
(including the deterministic ARM64 control), then freeze evidence for Master's
verification disposition. Existing helper source/bytes stay unchanged. This
investigation grants no full Runtime, parent, Slice, main or release acceptance.
