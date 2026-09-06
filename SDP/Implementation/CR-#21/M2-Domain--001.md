# M2 Domain contribution contract

State: ACTIVE kickoff, I-03 / M2. Authority: #58 under #21.
Predecessor: M1 #57, accepted source/review/integration
05982a4adb35e39d7b7ba371e0d16d83b2b3674c. Source CI34031420582 and canonical
integration CI34031833446 each passed18 applicable jobs; M3 helper explicitly
not run. Independent M1-Review--003 ACCEPT resolved all six infrastructure findings.

All SLC-01..13 remain selected. Domain is a shared foundation contribution to
those Slices, not a completed vertical behavior. All143 baseline findings remain
open. Accepted REFDES/API A1, corrected Domain review and frozen BoundaryTypes/BC
consumers define the actual semantic contract; full #58 supplies scope/acceptance.

Fresh worker branch: codereview-21/layer-domain, from the Master kickoff following
the M1 completion above. Worktree: C:/Users/hanse/GIT/gh-tree-wt/domain-implementation.
Write ownership: only new internal/domain/ values/tests/README and its bounded
worker report SDP/Implementation/CR-#21/Domain/M2--001.md. No legacy, other-layer,
module/workflow, accepted design, frozen BC or program metadata edits. Master
coordinates; a separate fresh reviewer inspects frozen exact source and evidence.

Implement immutable comparable private-field values and validating constructors:
Remote/LocalCommon RepositoryID, WorktreeID, Local/RemoteHead BranchID, PRID,
full nonzero canonical SHA-1/SHA-256 OID/ObjectFormat/Revision, closed Attached/
Detached/Unborn Head, StashID, unambiguous length-delimited LaunchPointID and
positive opaque SessionID. ExactTarget is closed Commit/Branch/PR intent with
immutable expected Revision. Preserve different PR base/head/fork repository
scopes and never fabricate an unborn Revision or turn latest-ref into exact intent.

Domain owns no path/URL canonicalization, clocks, context, storage/observation/
transport DTOs, serialization tags, process behavior or identity allocator.
Adapters mint/verify repository tokens and native branch validity; Runtime owns
SessionID allocation/exhaustion. Pure branch-name validation preserves exact bytes
and cannot normalize or expand previous-checkout shorthand into another identity.

Deterministic V-DOM-01..03 tests cover invalid/zero/foreign scopes, equality,
40/64 full OID/case/zero/nonhex/whitespace, closed variant construction, exact
target/Head/fork consistency, copy and unambiguous launch identity boundaries.
No Git executable, environment or network dependency in Domain unit tests.
Run appropriate unit/race/full-suite, M1 architecture/public-type/all-platform
checks and configured branch CI at exact source. Preserve all baseline tests.

Worker freezes/pushes source with actual commands/results/limits. Separate reviewer
reads real source/tests/contracts, returns ACCEPT or findings, and re-reviews any
correction. Master alone integrates after current branch CI and records source/
review/integration evidence. Then API/ports and State-owned viewmodel leaves may
proceed under separate Issues; no adapter work before their required M2 surfaces.
An insufficient frozen boundary triggers BC-CHANGE before affected implementation.
