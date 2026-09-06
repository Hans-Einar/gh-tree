# M3 Persistence contribution — Issue #63

State: IN PROGRESS; durable codec/version and Windows acquisition milestones,
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

Source is the commit adding this subsection, separately reported by exact SHA
after push. `acquire_windows.go` adds request-owned NtCreateFile ancestor guards
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

SLICE(S): SLC-01/04/05/09/10/12/13 foundations only. REVIEW: pending fresh reviewer.
INTEGRATION: none. TAG: none. All full Slices and baseline findings remain open.
NEXT: commit/push P2a, then Unix acquisition/locking/read and complete constructor/
port wiring plus metadata/publication/recovery milestones;
freeze the complete adapter later for independent review and full native gates.
