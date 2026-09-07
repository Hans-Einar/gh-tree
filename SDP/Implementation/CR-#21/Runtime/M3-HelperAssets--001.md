# M3 Runtime helper assets — Issue #70

## Current input-set correction (Master94f00a9e / ledger88)

The fresh re-review at `f6128251a0510021191880be658b705528fcc606` keeps H70-H01
OPEN at its reviewed source: post-selection source insertion was consumed by
Go's subsequent selection. This checkpoint corrects that continuation and is
pending final exact-source checks and independent re-review.
Existing-byte protection remains valid, and H70-M02/M03 are independently
RESOLVED. Exact technical732fd9bb CI34071597307 passed20/20, but that does not
exercise or resolve this continuation. The historical candidate summary below
is superseded by this section until a new reviewed correction exists.

Bounded native mechanism evidence on Windows amd64 Go1.25.0, preserved at
`2b510264b448ed5cfd3c5e660f24b035cc22a948`:

- An atomic NtCreateFile directory with OWNER RIGHTS and denied fresh
  WRITE_DAC/WRITE_OWNER/addition access blocks ordinary creation and permission
  override. It returns only granted mask0x13019f, omitting requested security
  modification rights; both relative child creation and retained restoration
  refuse AccessDenied. This rejected approach is preserved in the small
  `helpergen/testdata/atomic-directory-capability.go.txt` probe. Its owned
  t.TempDir cleanup completed; no permission change outside those new fixtures
  or rejected-cache cleanup was attempted. It is not product code or a passing
  immutable-materialization mechanism.
- `TestDirectoryOplockContinuousInvalidationProbe` passes its bounded cases. A native directory R
  oplock stays pending without changes. A new file added and removed using a
  directory mutator handle opened BEFORE watcher acquisition durably breaks
  it (level1 to0); restoration cannot erase that notification. An unchanged
  request is canceled and joined; writes after release succeed. Directory
  watches are advisory. Actual copied-tree wiring additionally exposed benign
  child Stat/Open/ReadDir invalidations with system-managed Last Access updates
  enabled. Reordering acquisition did not repair that limitation; no rearming
  or global timestamp-policy change was adopted. The native
  [oplock contract](https://learn.microsoft.com/en-us/windows/win32/api/winioctl/ni-winioctl-fsctl_request_oplock)
  supplies continuous invalidation rather than a before/after scan.

Master approved the bounded replacement: one queued asynchronous
[ReadDirectoryChangesW](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-readdirectorychangesw)
FILE_NAME|DIR_NAME request for every input directory. The request is never
rearmed. Every completion/error/zero-byte overflow permanently invalidates the
build, including insertion followed by removal through a previously granted
handle. Existing native file-byte and root/ancestor guards remain held. After
all requests are queued, the initial entire input set is checked, covering the
materialization interval; requests then remain alive through actual copied-Go
exit. Release distinguishes our own after-build cancellation from raced/early
completion and cancels/joins every request before freeing aligned buffers,
events or handles. No ACL, privilege, policy or trust-scope exception is used.

Targeted actual controls pass: unchanged two-build amd64/ARM64 images retain
recorded external bytes; source and wildcard-embedded data inserted/removed
AFTER the real Go child starts invalidate otherwise successful builds; after-
selection hardlink/rename and pre-watch granted directory-handle insertion/
restoration refuse both targets. Native zero-byte overflow, unexpected early
cancellation,16 close/write races, ordinary read stability, unsupported refusal
and partial-acquisition cancel/join/handle-release controls pass. The guard
also exposed Go1.25 attempting to canonicalize our generated .info metadata;
it now uses exact compact RevInfo encoding, retaining the same verified pins.

Actual guarded regeneration passes with917 inputs, digest
`3817c9a3a9bde9aff8413e2404066b790e721f26f3ebd614d95a367b26335968`.
Both gzip/image byte arrays remain the fixed ab608 outputs below. Next: final
exact-source checker/no-rewrite/package checks, then the same independent
reviewer inspects this corrected source. No helper/Runtime/Slice acceptance or
newer Windows-source adoption is claimed. Historical details below apply only
at their explicitly recorded prior SHAs.

Disposition: corrected bounded candidate frozen for independent re-review.
Technical source: `732fd9bbfe1dad7430f71ceca8270283b82a29d7`.
No Runtime contribution, Slice, integration or release
is accepted here. Worker branch: `codex/cr21-runtime-helper-assets`, worktree
`C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets`.

Authority: full #21/#65/#69/#70, root instructions, frozen Application--Runtime
and BoundaryTypes 1.0.0, WindowsBroker/CwdAcquisition/Runtime feasibility and
Verification--001. Sprint-004-v04 / I-03 / M3. Correction dispatch:
`aa7feeda7625087f08f75f1a59221c4bb5a37b90`, ledger84. Master owns completion.
Actual broker source remains `ab608327e63727f66ffb1aa7b3200c2865307cf5`.
No newer Windows source, broker/native code, parent registry, module pin,
workflow, frozen record or other layer change is adopted.

## Source and corrections

Original product `b72ad94c60eb38d7c4ebe70fafd10b21a3289037` was rejected by
independent review `7fdfc9bedb66dd7a4cdb67eb798a936d24a06ace`. Its actual report
and small controls remain unchanged in M3-HelperAssets-Review--001 and
Evidence/Runtime-HelperAssets-Review--001. Original CI34068851101 attempt2
was 18 success/2 failures; canceled attempts are not evidence.

- `02bb1c8` corrects H70-M03: kernel final-path resolution supports selected
  Windows root junctions and rejects child redirection outside that physical
  root. Errors include the failing selected path and resolution stage.
- `a9559d99` corrects H70-M02: the existing command prepares pinned dependencies
  with Go's checksum mechanism in temporary byte-identical go.mod/go.sum copies.
  Any attempted pin update refuses. Fresh builders need no undocumented module
  metadata cache. Build-only preparation follows the existing proxy settings;
  both actual helper builds remain offline. This source passed all 20 CI jobs
  in [run34070628791](https://github.com/Hans-Einar/gh-tree/actions/runs/34070628791),
  including the exact canonical checker and native Windows source-closure test.
- `9a7d3262316670340458a387fedf38823020c197` is the tested initial H70-H01
  correction checkpoint. The current correction additionally retains native
  no-write/no-delete input-file handles and no-delete directory handles until
  each isolated build exits. Native controls distinguish actual retained-handle
  protection from read-only attributes.

## Enforced build contract

The interface remains `go run ./internal/runtime/cmd/helpergen -check` on native
Windows amd64 Go1.25.0, including actual IsWow64Process2 and executable admission.
Ordinary build/install/run never invokes the generator or downloads a helper or
compiler. Exactly broker-amd64.gz, broker-arm64.gz and manifest.json are generated;
all twelve public assets and the pure architecture-selected loader are preserved.

Actual target `go list -deps -json` selects repository/API/Domain, pinned modules
and standard Go/assembly/header/object/embed inputs. Repository text is normalized;
embedded and external bytes remain raw. Selected module buffers are independently
bound to root h1 pins by hashing the complete module directory, comparing each
captured selected buffer and checking the module go.mod pin. A cached ziphash or
an earlier verify is insufficient. Offline module metadata is derived from those
verified values and included in provenance. Recursive literal assembly includes
capture nested inputs, including previously omitted runtime/cgo ABI headers;
absolute/traversing include operands refuse. Compiler-generated go_asm.h derives
only from captured source. Parent Runtime, assets and generator cannot enter the
broker dependency closure; recipe source is separate provenance.

Each of two independent builds creates owned source/GOROOT/module/build/work
roots using only captured buffers. The copied Go/compiler/linker/assembler and
support files are used; no live module cache or GOROOT bytes are consumed.
Captured hashes are checked before materialization and again through retained
input handles. Target selection on the copied offline tree rejects unrecorded or
changed files. GOENV/GOWORK/GOFLAGS/CGO/experiments/FIPS/target baselines are fixed,
PGO is off and linking internal, with trimpath/buildvcs=false/empty build ID.
Build caches are fresh. The two PE32+ executable images must match for each
machine. Deterministic gzip uses best compression, zero timestamp, OS255 and no
name/comment. Recapture additionally detects source drift; it supplies no
substitute for isolated consumed inputs. Check mode compares all three exact
outputs and never enters the writing branch.

The loader's in-memory JSON/PE parsing, owned return bytes, corruption checks,
Windows386 dual payload and Windowsamd64 ARM64-only embedding are unchanged.
The separate policy prerequisite `4b31ac4c34df370b852a1626107c102ff300c6dd`
remains unchanged: only the authorized assets JSON/PE imports are admitted;
typed pe.Open aliases/references, native I/O, printing and goroutines refuse.
Unchanged independent parser/policy/WOW64 and twelve-build proof may be reused.

## Executed correction evidence

Local native toolchain: Go1.25.0 windows/amd64 at
`C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe`.
External observer: Python3.14, explicit UTF-8. All mutations are owned temporary
fixtures; shared modules/toolchains, existing native fixtures and user data are
untouched. Tests are checked in beside the generator; no duplicate archive added.

- Full helpergen, brokerassets and Composition architecture package tests pass;
  generator/assets vet pass. After native input guards, full helpergen tests and
  vet pass again. Existing parser/policy coverage is unchanged.
- H01 regression changes an owned selected module, standard source, compile.exe
  and go.exe during BOTH clean builds, restores all before recapture, and proves
  identical before/after manifest plus repeated amd64/ARM64 images containing
  only the recorded module literal. Both builds run with GOPROXY=off.
- Extra selected files injected into a materialized tree, unrecorded plan inputs,
  corrupt captured buffers, inconsistent actual module bytes and external
  assembly includes refuse. Actual retained handles block writes/deletion and
  directory renaming; released handles permit the same operations.
- Owned local-proxy controls pass for fresh, selected-only and populated offline
  caches. Inconsistent pins refuse and preserve the original bad pin bytes.
  At a9559d99, real generation/checker passed using distinct empty caches; an
  external observer confirmed all 45 Runtime files' hashes, lengths, mtimes and
  file set unchanged by the exact checker.
- Owned selected-root junction positive and redirected-child negative controls
  pass, including failing path/stage diagnostics.
- Final guarded regeneration passes: 915 recorded inputs, source digest
  `8b7e1ecac4c0510dd99b8bb5e09ea1f5ad1488858fbc979efae8d565ef4e53f6`.
  Both helper images are byte-identical to the existing fixed ab608 outputs:

| Target | Image bytes / SHA256 | Gzip bytes / SHA256 |
|---|---|---|
| amd64 / 0x8664 | 3493376 / `9188bd5063040a2a39b7d5d550b06cd22bebcad73c9406a14f53314b54b44c5b` | 2038259 / `437f614f3ae04390c47f547fd0a7f60608870fb1f63cb136379f0a1b725f0183` |
| ARM64 / 0xAA64 | 3308544 / `2676e8efdcd53f1a87de64790d22ec2fa2f190419005eba25213978e29696e5d` | 1902326 / `c9952ce13f6f4846c7445226d2bfc91a2826f135623ce0fac965452b2d90c8f4` |

At exact clean technical source `732fd9bb`, the unchanged checker also passed
with BOTH an empty owned GOMODCACHE and an owned junction to the canonical Go
installation. It rebuilt both architectures twice from guarded offline snapshots.
The external observer verified all 54 observed files (the complete Runtime file
set plus go.mod/go.sum): SHA256, length, nanosecond mtime and file set unchanged.
The emitted digest is the final 915-input digest above. Final brokerassets tests,
diff whitespace/scope and clean-tree checks pass. The toolchain junction itself
was removed. Automatic approval review rejected the proposed PowerShell recursive
cleanup of these three owned caches with "blocked by policy" before executing
the command; no bypass or retry was attempted. They remain locally under
`C:/Users/hanse/AppData/Local/Temp/`:
`gh-tree-helper-fresh-43f6e3a638ca4bf5afc5ec390cb2f542`,
`gh-tree-helper-check-fresh-krknen5u`, and `gh-tree-helper-final-xdx2pbor`.
These are reproducible downloaded dependency caches, not unpublished product or
verification evidence. No source change remains unpublished after this report
checkpoint; the rejected cleanup is not the unrelated preserved ACL fixture.

## Remaining gates

The existing independent reviewer must reopen the corrected actual source.
Final candidate CI and that bounded review are still required. Master
alone coordinates later reviewed Windows-source adoption and regenerated assets;
Windows candidate5e964336 is not included. Sessions/parent/native ABI/ConPTY,
ARM64/emulation, full V-RUN/Slice, serial integration and release gates remain
separate mandatory work. The blocked Git review is untouched.
