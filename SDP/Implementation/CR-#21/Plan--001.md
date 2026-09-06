# Implementation plan — complete v0.4 refactor

State: ACTIVE, I-03 / M1. Authority: #21 and bounded contribution Issue #57.
Integration branch: codereview-21/refactor.
Integration base: 97cc0d8257603766dd741b49b7d8005857b421a9.
Product baseline: f626077ca0e59fbe9ede7ba1116982bb94b2eb6b (v0.3.14).

M0 is complete: accepted design #52 / PR54 and all seven BCs/shared annex FROZEN
1.0.0 under #55 / PR56. Technical BC source7685494e45c0ef44fbccf9b49a589a90a78026d0;
final reviewed freeze3ac7f20cedd8d14f131b2a15a602b233ef8ca5dc; actual merge is the
integration base above. Final18 CI checks passed in34027493442/34027495276.

All SLC-01..13 in accepted Slices--001 remain selected. This plan does not narrow
the user goal, declare a folder a completed Slice or waive native verification.
All143 product findings remain open. M1..M8 retain the accepted serial order;
M1 precedes Domain/API/ports/viewmodel leaves at M2. Final verification, product
PR/main, source/tag/twelve-asset and controlled install/upgrade gates remain ahead.

## Active contribution contract: M1 / #57

Purpose: establish branch CI and architecture checks before dependent product
workers rely on them; disable unsafe automatic bootstrap publication. This is
primarily a SLC-13 build prerequisite and supports the entire selected Slice set.
It does not complete SLC-13 or authorize another layer, entry cutover or release.

Fresh worker owns codereview-21/layer-composition-bootstrap in the separate
composition-bootstrap worktree. Master delegates exactly .github/workflows/ci.yml,
.github/workflows/release.yml, new internal/composition/architecture/ tool/tests/
fixtures/README, and its bounded worker report Composition/M1--001.md. Shared
module changes require a separate explicit Master assignment. No other old source,
frozen BC or accepted design file may change. Master owns sprint/traceability.

Required output and proof are the full #57 contract: existing native test/vet/build
and Linux race retained; canonical-branch push/manual CI; all twelve cross-builds;
native Windows ARM64 infrastructure evidence; staged exact-path/blob allowances,
strict new production/import/exported-type policy across all target selections,
separate external-test policy, meaningful negative fixtures, empty-allowance final
mode, explicit future Runtime helper provenance/reproducibility hook and safe
publication separation. Missing M3 helper inputs are a named prerequisite, never
a fake passing conformance check. No tag, release or published-asset mutation.

Worker freezes/pushes exact source HEAD with actual commands/evidence/limits.
Separate fresh reviewer reads the real source, tests, workflows and authority;
author summaries never count as review. Corrections require re-review at exact
HEAD. All configured branch CI must pass before Master serially integrates the
accepted source and records source/review/integration SHAs and remaining findings.
No product PR to main is permitted before the full accepted verification gate.

## Subsequent gated order

M2 Domain -> API/ports and State-owned viewmodel; M3 Git -> GitHub -> Persistence
-> Launch Discovery -> Runtime; M4 Application -> State -> View; M5 complete
new-stack headless vertical harness; M6 Composition host/entry cutover; M7 mapped
legacy retirement and strict final tree; M8 complete verification/product PR/main/
release. Each contribution gets its own Issue, layer-owned worker and independent
review. See accepted Slices--001 for actual end-to-end behavior and Verification--001
for required real Git, storage, native Runtime, State/View, migration and release
evidence. A failing frozen boundary triggers BC-CHANGE before affected work proceeds.
