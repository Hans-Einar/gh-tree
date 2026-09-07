# M3 Windows Runtime — #69 under #65/#21

Role: bounded Windows Worker. Branch `codex/cr21-runtime-windows`; base
`46fe59c5eae3f08df6ebbcc7bcd41536f3fd6864`. Ownership is the #69 filename/subfolder
split within Runtime. Frozen Application--Runtime/BoundaryTypes 1.0.0,
WindowsBroker/CwdAcquisition and Verification govern this contribution. No
acceptance, integration, full adapter or native platform completion is claimed.

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
