# M3 committed helper binding independent review — Issue #72

Disposition: **ACCEPT — reviewed Windows/helper source-image/native binding
milestone.** No finding remains in this bounded scope. This does not accept the
complete Runtime adapter, parent Sessions, integrated Slices or release.

Date: 2026-09-07. Role: fresh independent Reviewer, separate from the author and
Master. Authority: #72 under #65/#69/#70/#71/#21, Sprint-004-v04 / I-03 / M3;
Master dispatch e9e2d0b3606a58a3b3dee3bcc02e6c0ee5e29ff3, ledger CR21-0095.
Read root AGENTS/developmentInstructions, full authorizing Issues/comments,
frozen Application--Runtime/BoundaryTypes/BCFreeze, accepted WindowsBroker,
CwdAcquisition, Runtime feasibility and Verification, Runtime/evidence READMEs,
the current author report and both accepted independent native/helper reviews.
The author report supplied navigation; actual source, bytes and execution supplied
the review evidence.

Frozen branch: `codex/cr21-runtime-helper-binding`, initially clean and equal to
origin at `b9def1ff383bad3bcd2da53a1062eee93fae7255`, worktree
`C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-binding`. Technical source:
`cf160729581911c1cb02a24a458ec703337fb063`. Their sole difference is the author
report. From preparation `7e19d617ae430cd2c7ad414b9ea4622eb1e00617`, the complete
five-path delta is the three generated assets, exactly
`internal/runtime/helper_binding_windows_test.go`, and that report.

The unchanged broker tree `04afda252be325beaa6bf1f22c154a094b8daed9` equals accepted
Windows technical `bd78deafd4dd36e22d5b106eb7ef9c4edcd2e832` / review
`9b06f8bd0e8bc4baaaabb7dccd1054f38d74fec8`. The unchanged helpergen tree
`ee05f85bd3b3f3ed6a63535ea41bc3e873821dd7` equals accepted helper technical
`b6f161f5189d70b66b95129237b51f9984d58e35` / review
`6f385a9cab6798661970c4dc2a0aa56edba6c97b`. Loader, policy, module pins, workflow
and frozen contracts are unchanged. Earlier comprehensive ABI/DLL/TLS/ACL/native
failure and generator adversarial evidence is reused at those exact trees.

## Source, recipe and committed bytes

The reviewer traced the actual dependency-selected capture, root go.mod/go.sum,
13 recipe sources, canonical compiler/linker/assembler and selected pinned
`golang.org/x/sys v0.44.0` into the isolated guarded build. Existing byte/path
guards and continuous one-shot directory name invalidation remain active through
actual copied-Go exit and joined release. The generator checks both independent
builds per target, source recapture and every generated output byte. It excludes
Runtime root, brokerassets and helpergen from the executable dependency graph;
the recipe is provenance, without recursive assets or a self-referential SHA.

The exact native Windows amd64 Go1.25.0 command
`go run ./internal/runtime/cmd/helpergen -check` independently **PASSed**, exit0,
27.851 seconds, at the frozen candidate. A separate UTF-8 Python observer proved
the complete set of 76 Runtime/go.mod/go.sum files retained bytes, lengths and
nanosecond mtimes. It separately matched all 56 normalized repository source
entries and 13 recipe entries and decoded/hash-checked both actual PE images.
The final native test run also left those inputs unchanged.

Manifest SHA256:
`03c9db1f329aa29a4d5b4c279dcc2fc17ac0cd42967d7b4527c2bef5c482a3e8`.
The 955-input source digest is
`23bf82b051123cd1aa31c5a2368d1cc732f4b09cbc33ea2c9abf4f08f0cfdde5`.
These assets attest the accepted corrected Windows source, superseding the
preparation checkpoint's deliberately stale ab608-based images.

| Target | Image bytes / SHA256 | Gzip bytes / SHA256 |
|---|---|---|
| amd64 / 0x8664 | 3612160 / `bfdb8eb2ec496222b8033bbeca2331c319fd1a0cfcacb6bb7adf3e79c138781c` | 2109197 / `b62f559568d69b369ea5cf2ecef93d56f65ae0c5f3d03ab4241f7b64a447f44b` |
| ARM64 / 0xAA64 | 3422208 / `617b7b8a06b5333e146af86a63d5c93f89548f98ab9ab240e9a6600aa34dc5bf` | 1968006 / `54ecab0a18d9ca2bc762cb75baf8e9d6e07190c1370efdd86df41d6206b032c7` |

There remain exactly three generated outputs, with CGO0, explicit target/
microarchitecture, trimpath, buildvcs=false, PGO-off, empty build ID, internal
linking and deterministic gzip. Build-time pinned module preparation in the
canonical generator remains distinct from ordinary helper extraction/execution;
the latter performs no generator, compiler or helper download. Current ordinary
product builds still use the legacy CLI. Their success does not prove the later
public binary has adopted Runtime or contains its required final helper routing.

## Actual committed-image execution

The root test imports assets and broker through the allowed Runtime graph.
Imported broker test files/TestMain do not enter the root test executable. The
only compiled temporary executables are parent/user fixtures. The selected
`brokerassets.Load` bytes are manifest/hash checked, extracted by
`ExtractWindowsImage`, retained and disk-hash checked, and passed to
`StartWindows`. Its production CreateProcess uses that exact image with the
private mode and inherited authenticated endpoints; the committed broker/cmd
entry calls the real `RunWindowsPrivate`. No test broker, regenerated temporary
helper image, source hook or alternate dispatcher supplies the behavior.

The reviewer independently ran:

`go test -race ./internal/runtime -run '^TestWindowsCommittedHelperBinding$' -count=1 -timeout=180s -v`

**PASS**, exit0, package8.179s / wall9.999s on Windows11 build26200,
Go1.25.0 windows/amd64. The actually executing386 parent (PE014c/native8664)
passed every case through the committed native amd64 helper. The native driver/
user is race-built; the executing386 parent is CGO0 without race instrumentation.

| Case | Independently checked assertion and outcome |
|---|---|
| Terminal | Real native target startup breakpoint/cwd, relative cwd marker, no startup anchor/debugger in user code; ConPTY resize101x31, exact input and clean natural exit. Pre-canceled input is refused without dispatch or observed user effect. |
| Blocked input | A real65484-byte first write fills the pipe; the next dispatched blocked write returns a retained receipt on timeout, and another input is refused. Coalesced Stop cleans before receipt observation; repeated waits through an already canceled context preserve the same known partial accepted/delivered result and short-write error without replay. |
| Flood | More than1MiB of raw bytes drains without a UI consumer into the real bounded256KiB Runtime ring; its end/retained-size/final marker and natural cleanup pass. Detailed ring byte/offset correctness remains supported by the unchanged existing ring tests. |
| Missing command | Actual committed broker returns NotFound/ProcessContainment with os.ErrNotExist compatibility, no user marker, and complete cleanup retaining the original failure. |
| Stale cwd | A distinct directory identity returns Cwd/CwdAcquisition with ErrCwd compatibility, no user marker, and complete cleanup retaining the original failure. |

All five owners prove CleanupComplete, no residuals and disappearance of the exact
owned image and empty nonce directory. Terminal/blocked cases additionally retain
query/synchronize handles to the exact extracted helper and reported native user
image, verify actual machine identity while running, and require both handles
signaled after cleanup. The process census is read-only observation; it grants
no numerical termination authority. Full native descendant/ABI/partial-unwind
coverage remains the accepted unchanged Windows review's evidence, not a new
claim that these five tests duplicate it.

## Exact native CI and applicability

Independently queried [CI34078800342](https://github.com/Hans-Einar/gh-tree/actions/runs/34078800342),
attempt1, exact technical cf16072: **all20 jobs SUCCESS**. All twelve individual
architecture/public-asset build jobs, native Linux/macOS/ordinary-user FreeBSD,
Linux race, inventory and both Windows jobs are successful. Workflow review
confirms full native test/vet/ordinary build, unchanged canonical helper -check
and a clean Runtime checkout assertion; no workflow gate was relaxed.

| Gate | Directly inspected evidence |
|---|---|
| Windows amd64 / job101610045608 | Source cf16072, hostX64, Go1.25.0 windows/amd64; Runtime42.659s, broker77.129s, assets6.796s, helpergen152.423s; test/vet/build all succeed. |
| Windows ARM64 / job101610045459 | Source cf16072, hostArm64, runnerWindows/ARM64, imagewin11-arm64 version20260830.155.1, Go1.25.0 windows/arm64; Runtime14.210s, broker33.782s, assets2.630s, helpergen4.501s; test/vet/build all succeed. |
| Canonical helper / job101610119625 | Native Go1.25.0 -check reports two clean builds per target and the same955-input digest; clean-checkout assertion and ordinary build succeed. |

CI prints package success, not verbose passing test bodies. The inspected exact
test has mandatory amd64 AND386 executing parents on the confirmed native ARM64
driver, each running all five cases with the committed ARM64 image. It checks
that native ARM64 itself embeds neither image. No applicable parent/case can skip
or fall back: emulated/native32 matrix-driver skips avoid recursively applying
the driver where the committed-helper route does not exist, while the actual
selected parent fixture insists it is emulated. Native same-executable paths and
native32 public dispatch remain separate gates, not silently omitted proof.
No local ARM64 execution is claimed.

The author's three final raw CI logs were independently fetched from their job
log API endpoints; after CRLF-to-LF normalization they exactly match the recorded
files/hashes. Its native-root-race log hash also matches. The asset-only b827 run
and canceled5d run are not used as final binding evidence. Direct final-source
execution above supersedes dependence on earlier no-rewrite/native observations.

## Small evidence and handoff

Evidence is in the existing `SDP/Verification/CR-#21/Evidence/` directory.
Logs are UTF-8/LF. The CI excerpt preserves selected original gh-rendered lines;
CI JSON distinguishes those rendered-log hashes from raw job API hashes. An
initial console print of the already saved CI excerpt hit Windows cp1252 on its
BOM; UTF-8 display corrected that observer-only output issue. All actual native
and checker commands completed successfully. No binaries, sources, caches or
broad duplicate archive are added.

| File (prefix M3-HelperBinding-Review--001-) | SHA256 |
|---|---|
| run.py | `7c7fde606e5613bd16702af65c5b6391e67801d9e34d457f8ae8002c9492ca2d` |
| checker.log | `89ef567022716524d1891709b095be23fed181d54970c82d5cf9cceb28ee4084` |
| native.log | `5baabc27073e8f2915ee4204e4a141e16fb8209c3ef935bdd07c6dd3740144a4` |
| result.json | `decc6ce983d133267ad660f8c4878ef9124f658167b4b634fd4c9ecefb4c7252` |
| ci.json | `3819e12d60f0a1fbdcaf80e486b61d839f52c75c747f517f0da9ac63dbab8b69` |
| ci-excerpts.log | `8ad3a339185d1bb54780f59ebb71af3458ccc8c156d9cc22a5ec35255d90a0aa` |

Only this report and the six small evidence files are reviewer-owned changes;
product/test/generated/module/workflow/frozen trees remain byte-identical to the
frozen candidate. gofmt and Git whitespace/isolation checks pass. The review
checkpoint is committed/pushed with [skip ci], with clean worktree/origin checked
at handoff. No merge, contract change or correction was performed.

Exact next permitted action: Master may supply this accepted Windows/helper
binding milestone to #71 for reviewed parent/native assembly. Unix, complete
Sessions and Runtime verification, canonical serial Git→GitHub→Persistence→
Discovery→Runtime integration, M5/M6 public CLI/host/private self-entry cutover,
full vertical/native/package-manager and M8 release/install gates remain open.
No Slice or baseline finding closes from this bounded acceptance. The separate
blocked Git review/access issue, policy-rejected residues, unrelated worktrees
and access/model controls were untouched.
