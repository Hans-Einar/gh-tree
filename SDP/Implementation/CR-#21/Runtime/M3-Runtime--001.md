# M3 Runtime contribution — #65

State: IN PROGRESS / PARTIAL CHECKPOINT / NOT ACCEPTED

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

## Next permitted work

Commit/push this coherent partial checkpoint before further substantial work.
Then implement inherited-endpoint/parent validation, private Fchdir/SID supervisor
and acquired-group helpers, plus Windows native broker/embedded-helper closure, with
explicit partial checkpoints and native fixture verification. Assemble all twelve
Sessions methods around actual resource owners; no shortened replacement scope.
Request Master coordination before any shared-interface/workflow/private-host
dispatch change or frozen-contract gap. Preserve every native requirement and
all failed/not-run verification facts until exact-source evidence resolves them.
