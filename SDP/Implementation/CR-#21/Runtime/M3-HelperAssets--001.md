# M3 Runtime helper assets — Issue #70

Disposition: corrected bounded candidate frozen for independent re-review.
Technical source: `b6f161f5189d70b66b95129237b51f9984d58e35`.
Branch: `codex/cr21-runtime-helper-assets`; worktree:
`C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets`.
No helper, Runtime, Slice, integration or release acceptance is claimed.

Authority: full #21/#65/#69/#70, root instructions, frozen Application--Runtime
and BoundaryTypes1.0.0, WindowsBroker/CwdAcquisition/Runtime feasibility and
Verification--001; Sprint-004-v04 / I-03 / M3. Current correction dispatch is
Master94f00a9e / ledger88, with the subsequently approved private continuous
name-change mechanism. Master owns final gates. Actual broker source remains
`ab608327e63727f66ffb1aa7b3200c2865307cf5`; no newer Windows source is adopted.

## Independent findings and correction history

The independent report and small controls in M3-HelperAssets-Review--001 and
Evidence/Runtime-HelperAssets-Review--001 are unchanged by this worker.

| Finding | Actual reviewed state and current change |
|---|---|
| H70-M02 | Independently RESOLVED at732fd9bb/f6128251. The existing command prepares pinned dependencies through Go's checksum mechanism using unchanged temporary go.mod/go.sum copies. Fresh/selected-only/offline and inconsistent-pin/no-rewrite controls pass. |
| H70-M03 | Independently RESOLVED at732fd9bb/f6128251. Native final-path resolution accepts a selected root junction and rejects child redirection outside its physical root, with path/stage diagnostics. |
| H70-H01 | OPEN at732fd9bb/f6128251: captured external byte replacement was corrected, but Go could consume a new source added after selection. Current b6f161f corrects the remaining input-set acceptance gap; independent re-review is pending. |

Coherent pushed checkpoints:02bb1c8 root alias; a9559d99 preparation;9a7d326
isolated inputs;732fd9bb retained byte guards;2b51026 native mechanism probes;
b6f161f continuous input-set invalidation. Original b72ad94 CI34068851101
attempt2 had18 success/2 failures. Exact a9559d99 CI34070628791 and732fd9bb
CI34071597307 each passed20/20, but neither covered the later H01 insertion
finding. Canceled runs are not evidence. Final b6f161f CI remains a Master gate.

## Enforced helper contract

The interface remains `go run ./internal/runtime/cmd/helpergen -check` on native
Windows amd64 Go1.25.0, including actual IsWow64Process2/executable admission.
Exactly broker-amd64.gz, broker-arm64.gz and manifest.json are generated.
Ordinary build/install/run invokes no generator and downloads no helper/compiler;
all twelve public assets and the pure architecture-selected loader are preserved.

Actual target Go dependency selection includes repository/API/Domain, pinned
module and standard Go/assembly/header/object/embed inputs. Repository text is
normalized; external/embedded bytes stay raw. Selected module buffers are bound
to root h1 pins by complete module-directory hashing and exact captured-buffer/
module-go.mod comparison. Cached ziphash or earlier verification is insufficient.
Offline metadata derives from verified pins. Its .info record now matches Go1.25's
compact RevInfo encoding, avoiding a cache canonicalization attempt during read.
Recursive literal assembly includes capture nested Runtime ABI headers; absolute
or traversing include operands refuse. Generated go_asm.h derives from captured
source. Parent Runtime/assets/generator cannot enter the broker dependency graph;
recipe source is separate provenance without a self-referential commit SHA.

Each of two independent builds uses only captured buffers in owned source,
GOROOT/module/build/work roots, copied Go/compiler/linker/assembler and support
files, fresh build caches and offline subprocesses. Captured bytes are hashed
before materialization and through retained no-write/no-delete file handles.
Retained native directory handles bind root/ancestor names. Fixed environment,
CGO0, explicit architectures, PGO-off/internal-link, trimpath/buildvcs=false/empty
build ID and deterministic gzip remain. Both PE32+ machine images must match.
Check compares all three exact outputs and never enters the writing branch.

The current input-set guard queues one asynchronous
[ReadDirectoryChangesW](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-readdirectorychangesw)
FILE_NAME|DIR_NAME request for every input directory. After all requests are
queued, the initial entire input set is validated, closing the materialization
interval. Every completion/error/zero-byte overflow permanently invalidates
output, including insertion followed by removal using an already granted handle.
No request is rearmed. Aligned buffers, OVERLAPPED records, events and handles
remain live through actual copied-Go exit. Release distinguishes our after-build
cancellation from raced/early completion, cancels/joins all requests, then closes
resources. Acquisition/release/cleanup errors reject output. This proves whether
the produced output is admissible; it does not claim additions are prevented.
No ACL, privilege, global setting or trust-scope exception is used.

## Exact executed evidence

Native toolchain: Go1.25.0 windows/amd64 at
`C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe`.
External observer: Python3.14, explicit UTF-8. Only owned temporary fixtures were
mutated; shared source/toolchains, existing native fixtures and user data were
untouched. No duplicate archive or additional generated output was introduced.

- At exact b6f161f, full helpergen/brokerassets/Composition architecture tests and
  generator/assets vet pass. The exact native checker passes, independently
  rebuilding both targets twice. External SHA256/length/nanosecond-mtime/file-set
  observation confirms all60 Runtime-plus-go.mod/go.sum files unchanged.
- Real Go child-start controls insert/remove source and wildcard-embedded data
  AFTER the actual child starts; otherwise successful amd64/ARM64 compilations
  are rejected by permanent invalidation. Hardlink/rename insertion+restoration
  after successful selection refuses both targets. A directory mutator handle
  granted BEFORE watch acquisition likewise cannot evade invalidation. Original
  source recapture remains identical throughout these controls.
- Unchanged repeated builds retain only recorded module bytes despite replacement
  of an owned external module, standard source, compile.exe and go.exe during
  BOTH builds, then restoration before identical source recapture.
- Native zero-byte overflow, unexpected early cancellation,16 close/write races,
  ordinary Stat/Open/ReadDir stability, unsupported refusal and partial watch
  acquisition cancel/join/handle-release controls pass. Released-directory
  mutation controls succeed. No sleep/rearm or before/after-only proof is used.
- Final source closure:917 inputs; digest
  `3817c9a3a9bde9aff8413e2404066b790e721f26f3ebd614d95a367b26335968`.
  Both regenerated image/gzip arrays remain the exact fixed ab608 bytes:

| Target | Image bytes / SHA256 | Gzip bytes / SHA256 |
|---|---|---|
| amd64 /0x8664 |3493376 / `9188bd5063040a2a39b7d5d550b06cd22bebcad73c9406a14f53314b54b44c5b` |2038259 / `437f614f3ae04390c47f547fd0a7f60608870fb1f63cb136379f0a1b725f0183` |
| ARM64 /0xAA64 |3308544 / `2676e8efdcd53f1a87de64790d22ec2fa2f190419005eba25213978e29696e5d` |1902326 / `c9952ce13f6f4846c7445226d2bfc91a2826f135623ce0fac965452b2d90c8f4` |

Unchanged independent loader/parser/pe.Open-policy/WOW64/twelve-build proof may
be reused. Policy prerequisite4b31ac4 and the three Composition files, loader
code/tests/embed selection, broker/native source, root pins and workflow are
unchanged by the correction passes.

## Rejected probes, residue and next action

The small atomic-directory DACL probe is preserved in
`helpergen/testdata/atomic-directory-capability.go.txt`: OWNER RIGHTS blocked
fresh mutation/security access but the creator received mask0x13019f without
restoration rights; relative creation and restoration refused. Its new t.TempDir
cleanup completed. Directory R oplocks correctly detected insertion/removal,
but actual copied-tree wiring also invalidated benign child reads with native
system-managed Last Access updates enabled. That candidate was not accepted or
rearmed, and no global timestamp policy changed; the filtered mechanism replaces
it. The tested native probe checkpoint2b51026 preserves the prior evidence.

No rejected-path cleanup was retried. Previously policy-rejected caches remain
under `C:/Users/hanse/AppData/Local/Temp/`:
`gh-tree-helper-fresh-43f6e3a638ca4bf5afc5ec390cb2f542`,
`gh-tree-helper-check-fresh-krknen5u`, `gh-tree-helper-final-xdx2pbor`.
The previous toolchain junction was removed before that rejection. These caches
contain no unpublished product/evidence. The unrelated denied-ACL fixture and
blocked Git review remain untouched. New probes and guarded builds cleaned their
owned temporary resources without changing permissions.

Next permitted action: the same independent reviewer reopens this exact candidate
and H01 controls; Master checks exact-source CI and records disposition. Newer
Windows-source adoption/regeneration, Sessions/parent/native ABI/ConPTY/ARM64/
emulation, complete V-RUN/Slice, serial integration and release remain separate
mandatory gates. The worker stops at this clean pushed report checkpoint.
