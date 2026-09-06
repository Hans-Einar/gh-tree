# M3 Persistence contribution — Issue #63

State: IN PROGRESS; durable codec/version and Windows/Unix acquisition milestones,
not adapter acceptance.
Branch: `codereview-21/layer-persistence`.
Initial base: `412f33e477cec03cb6eafe7b846c9bcdd02c0a25`.
Worker owns only `internal/persistence/` and this report. Master supplies separate
independent review and serial integration after Git/GitHub. No legacy/shared/
frozen files, product entry point, global configuration or user stores changed.

Authority: full #21/#63 resumed assignment, UserRunContract resumed policy,
M3-Adapters/M3-Assignments, accepted Persistence LR and preceding Domain/Discovery
conclusions, REFDES/API/Storage/Slices/Verification/MigrationMap/FindingDisposition,
FROZEN Application--Persistence/Discovery, shared BoundaryTypes and State--App
contracts plus BCFreeze. Actual accepted Domain/API/ports are the type authority.

## Milestone P1 — strict documents and versions

Source: `a95d112aa493d3899d34728843b59090ab3c6f17`, pushed and clean at P1 handoff.
Files: `codec.go`,
`codec_documents.go`, `document.go`, corresponding tests and layer README.

| Clause / related checks | Actual implemented evidence |
|---|---|
| V-PER-01/03 codec/migration inputs | All three full schema0/schema1 codecs and constructors; exact strings/ordered targets, only prefixes-null, copied raw originals; all known object levels and noncurrent scope entries retained. Round-trip fixtures plus wrong-shape/null/duplicate/UTF-8/surrogate/depth/size/forward-integer tests. No disk migration or business intent choice. |
| V-PER-01 unknown preservation | Guarded-document retention primitive checks every retained object path. Independent member-removal controls reject loss, including enclosing-object deletion. Comparison preserves exact decimal values without float rounding or unbounded exponentiation. |
| V-PER-02/04 content/scope foundation | Complete-byte length/SHA256, presence, family, store, run WorktreeID/root binding; deterministic missing-anchor/literal-component tokens. Tests distinguish foreign family/root/worktree/parent/basename, raw/unknown changes and same-byte restoration. Native binding/absence proof is still pending. |

Executed locally: Windows amd64, Go1.25.0; targeted package tests and vet PASS,
gofmt and `git diff --check` PASS. Bounded pre-final numeric-normalization fuzz
run: 549186 executions in11.547s, PASS (23 new interesting inputs, no failing
corpus). Subsequent exact-decimal normalization change passed the updated package
tests/vet, including large positive/negative exponent carry/borrow controls.
Tests mutate only in-memory values; they perform no store or configuration I/O.
These are author tests, not independent review or native product verification.

## Milestone P2a — Windows acquired objects and permanent locks

Source: `19df0a110cfb91aa6ff7d6cda5b1dd20dc5ae0bb`, pushed/clean at P2a handoff.
`acquire_windows.go` adds request-owned NtCreateFile ancestor guards
with actual list/data-read/no-delete access, handle-relative no-reparse opens,
aligned full FileIdInfo/native birth identity, local-NTFS detection, coherent
bounded reads, missing-anchor observation without writes, and cancellable
permanent LockFileEx byte0/length1 locking. Native NTStatus plus errno are retained.

`acquire_windows_test.go` executes native acquisition/read/absence, parent movement
refusal, data-read child deletion exclusion, empty-parent junction conversion
and safe retained-parent relative refusal, multi-handle/two-store locking,
cancellation and kernel lock release after a real child process is killed.
Initial controls incorrectly assumed an empty parent could not convert and let
the child exit via Go deadlock detection; corrected tests follow the accepted
windows-anchor/Followup.md and keep the child alive until its owned parent kills/
joins it. No product contract was weakened or changed.

Exact candidate local tests: Windows amd64 full Persistence package PASS, vet
PASS; native Windows386/WOW64 native selectors PASS (0.654s); Windows ARM64
package build PASS, explicitly compile-only. Native FileIdInfo assertions:24
bytes, FileID offset8 on tested amd64/386. Formatting/diff checks PASS. All
filesystem/process fixtures are beneath test-owned temporary directories.
No API Storage method/public constructor/publication/recovery profile is claimed
implemented by these private primitives. P2a is a recoverable source checkpoint.

## Milestone P2b — Unix descriptor acquisition and flock

Source: `4c5e5bed6b38860dc7a823780475e264605cbf26`, pushed/clean at P2b handoff.
`acquire_unix.go`, platform profile helpers and native tests implement no-follow
Openat/Fstat/statx/BSD birth observations, no inherited descriptors, bounded
double-read consistency, missing-anchor observation, explicit moved-object versus
substituted-path revalidation, special-object refusal before blocking reads,
reference-counted inode mutex plus stable flock, and request-owned cleanup.

Native Linux execution PASS: Go1.25.0 crosscompiled CGO0 test binary, then actual
WSL openSUSE-Leap-15.5 / Linux6.18.33.2-microsoft-standard-WSL2 x86_64,
UID/GID65534 nobody, ext4 fixtures entirely beneath task-owned Linux `/tmp`.
Complete package cases and fuzz seeds pass, including real child-process kill/
flock release and directory rename followed by substituted symlink refusal.
All nine Unix target test binaries compile; Windows P2a source is unchanged.
Filesystem recognition is acquisition only, not native metadata/durability proof.

Bounded local evidence: Windows temporary directory
`C:/Users/hanse/AppData/Local/Temp/gh-tree-persistence-p2b-aa36504819f0492097fff66545082146`,
Linux staging `/tmp/gh-tree-persistence-p2b.oF0eSB`. `persistence-linux.test` SHA256
`b3b839e7dc56cd532f2e43650ea1e9cbf22867e7dc7288415f473466cd8d720c`;
`native-linux.log` SHA256
`19f9d6dc38c0254b43d2514ab49b274384a42fdd0a50017df464dc171ff4ead9`.
Three harness-only attempts initially failed from WSL backslash/command-path/
PowerShell dotted-argument transport, before tests ran. Final reproducible route:
WSL `--exec /usr/sbin/runuser -u nobody -- /usr/bin/env TMPDIR=<owned>/tmp
<owned>/persistence-linux.test -test.v=true -test.timeout=30s`, with literal quoted
PowerShell test flags. No Linux toolchain installation/global change was made.

## Remaining native and acceptance work

P2: native constructor/run binding, no-follow retained object acquisition, stable
cooperative locks, coherent bounded reads, missing-parent revalidation/adoption,
supported metadata and all six declared Storage methods. Preserve the originally
supplied Expected absence anchor in the native manifest through parent creation;
shared recovery Original can be the guarded established-parent absence.

P3: exclusive payload/manifest/raw backup and exact retained original, bounded
admission, selected publication/barriers, complete outcomes and stable persisted
recovery IDs/restart observation. No replay/eviction/fallback. Fault/crash/race/
late-writer/resource controls and all required native profiles remain mandatory.
These are forthcoming implementation milestones, not Unsupported placeholders.

Master supplied the exact common DirectoryIdentity convention; Git's separate
native source checkpoint is `dd80cdecc0c835d119dc21671fbba2efacc23644`, read-only
convention evidence with no import/call. Native FreeBSD metadata/mechanism job
remains Master-coordinated. Existing
x/sys0.44.0 is sufficient to begin native helpers, and current shared CI provides
Linux, macOS ARM64, Windows amd64/ARM64 and Linux race. No workflow/module edit
was made here. Native evidence is distinct from architecture cross-builds.

## Milestone P2c — selected native publication and retention primitives

Exact source is the commit adding this subsection, reported after push. This is
a private mechanism checkpoint; no public Storage commit path is claimed.
Windows class65 rename/class11 no-replace hardlink use retained handles and
pointer-aligned native layouts, with explicit386/amd64 offsets. Unix selects
Renameat for presence and Linkat for absence. A successful absence Linkat leaves
the owned payload name so subsequent cleanup cannot erase the known commit point.
No replacement fallback, replay, truncate or retention deletion is implemented.

Native author tests: Windows amd64 complete Persistence package and vet PASS;
native386/WOW64 selected publication tests PASS; WSL Linux/ext4 UID/GID65534
selected publication tests PASS. Cases cover existing-absence competitor refusal,
no-replace original retention, late writes through an old held handle versus the
independent raw backup, and three fresh native readers through30 replacements.
The initial Windows test used an ordinary Go reader that did not share DELETE;
its actual sharing failure was corrected to the required native reader profile,
without changing the publication primitive. Linux file/parent fsync executed.
Windows namespace power-loss durability and full request outcomes remain unproved.
Bounded test staging: Windows temporary `gh-tree-persistence-p2c-27053b08c0d844bc9c1fb47731ec1c76`;
Linux `/tmp/gh-tree-persistence-p2c.iVD5TK`, all fixtures owned under its tmp folder.

P2b CI correction: actual4c5e5be CI34054966709 passed18 SUCCESS jobs and the sole
pre-Runtime helper skip, executing this package on Windows amd64/ARM64,
macOS26 ARM64, Linux and Linux/race. That source CI proves P1/P2a/P2b, not these
new publication primitives or pending metadata/recovery. FreeBSD remains pending.

Master confirmed the already accepted bounded Windows security profile excludes
audit-only SACL replication; READ_CONTROL queries must never claim audit absence.
Owner/group/ordered DACL/protection/inheritance and access-affecting label/resource/
CAP queries/copy/refusal remain mandatory before public publication. No new public
option, privilege escalation or contract change is introduced by this clarification.

SLICE(S): SLC-01/04/05/09/10/12/13 foundations only. REVIEW: pending fresh reviewer.
INTEGRATION: none. TAG: none. All full Slices and baseline findings remain open.
NEXT: commit/push P2c; complete native metadata and constructor/port wiring with
version/manifest/retention/admission/restart barriers around these primitives;
freeze the complete adapter later for independent review and full native gates.
