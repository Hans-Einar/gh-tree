# M3 Persistence contribution — Issue #63

State: IN PROGRESS; first durable codec/version milestone, not adapter acceptance.
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

Source is the commit introducing this milestone; its exact SHA is reported in
the Master dispatch/ledger after commit and push. Files: `codec.go`,
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

Master coordination requested: exact Git-issued DirectoryIdentity representation
for cross-adapter comparison; native FreeBSD metadata/mechanism job. Existing
x/sys0.44.0 is sufficient to begin native helpers, and current shared CI provides
Linux, macOS ARM64, Windows amd64/ARM64 and Linux race. No workflow/module edit
was made here. Native evidence is distinct from architecture cross-builds.

SLICE(S): SLC-01/04/05/09/10/12/13 foundations only. REVIEW: pending fresh reviewer.
INTEGRATION: none. TAG: none. All full Slices and baseline findings remain open.
NEXT: commit/push P1, then implement native acquisition/locking/read milestone;
freeze the complete adapter later for independent review and full native gates.
