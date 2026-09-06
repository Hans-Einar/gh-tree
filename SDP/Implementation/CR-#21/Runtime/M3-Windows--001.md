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
