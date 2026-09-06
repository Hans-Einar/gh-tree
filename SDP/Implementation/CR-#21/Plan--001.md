# Implementation plan — complete v0.4 refactor

State: ACTIVE, I-03 / M3. Authority: #21 and adapter Issues #61..65; M1/M2 #57..60 complete.
M1 #57 is complete; its original contribution contract remains below as history.
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

## Completed contribution contract: M1 / #57

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

M1 completion: independent ACCEPT/source/integration05982a4; source CI34031420582
and canonical integration CI34031833446 passed18 applicable jobs each, with explicit
M3 helper skip. All six M1 review findings resolved; all143 baseline findings open.
M2 Domain #58 is accepted/integrated at6113ca5 with source/canonical CI18/1skip.
Completed surface contract: [M2-Surfaces--001](M2-Surfaces--001.md), #59/#60.
API-Inventory--001 is the independent completeness aid; full frozen BCs govern.

M2 COMPLETE atd60afe8fb763269964bcdc05958c8cdbc7849dea. Viewmodel#60 accepted
sourced279c3f2af59e48029e77d845f5a55271c7889cf was integrated after verified API
integrationecf12e9; its exact integrationCI34046727344 passed18/explicitM3skip.
Only accepted viewmodel files were added, with no other product/build change or
merge conflict. Issues59/60 are CLOSED, verified on GitHub. Domain6113ca5, API
finald4c3bfd/technical6e289c4 and viewmodeld279c3f are complete M2 dependencies.
All13 Slices/all143 baseline findings remain open; CLI still uses legacy stack.
M3 preparation now opens bounded adapter Issues under#21, records exact contracts/
traceability before fresh implementation, and uses separate reviewers. Accepted
serial adapter integration remains Git→GitHub→Persistence→Discovery→Runtime.
M4..M8/full native/vertical/packaging/release gates remain intact.

M3 adapter Issues61 Git,62 GitHub,63 Persistence,64 Launch Discovery and65 Runtime
are authorized under#21 after complete M2. Contract M3-Adapters--001 and current
index/relations/ledger54 precede fresh worker dispatch. First parallel authors:
Git/GitHub/Persistence; Discovery/Runtime queued. Exact worker base is the Master
kickoff commit, recorded in dispatch ledger/Issue comments after worktree creation.
No product source is changed by this kickoff; serial integration order stays
Git→GitHub→Persistence→Discovery→Runtime. Complete native layer-owned gates and
M4..M8/full13-Slice/release requirements remain. Current3cf2641 CI34047530046
passed18/explicitM3skip; superseded docCI34047480826 was canceled, not passing.
