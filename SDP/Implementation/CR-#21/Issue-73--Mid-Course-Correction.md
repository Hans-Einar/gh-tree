# Issue #73 — Mid-Course Correction

Status: PROPOSED EXECUTION CORRECTION
Parent program: #21
Canonical integration branch: `codereview-21/refactor`
Snapshot basis: user-requested WIP stop `checkpoint/gh-tree-wip-stop/20260907-0532Z`

## 1. Purpose

The v0.4 refactor has produced strong architecture, contracts and several useful implementation foundations, but the execution model has become disproportionately expensive in reasoning usage, review orchestration, evidence duplication and worktree/process overhead.

This correction does not discard the accepted architecture, frozen contracts, completed M0-M2 work, accepted M3 contributions, or the selected vertical SLC-01..13 behavior contract. It changes how the remaining work is executed.

Primary goals from this point:

1. maximize working-product progress per engineering/reasoning cost;
2. preserve safety-critical review for Git mutation/recovery, Runtime/process lifecycle, Persistence integrity and final cutover;
3. commit and push every coherent recoverable milestone;
4. use targeted unit/integration tests during implementation and reserve full expensive matrices for meaningful integration gates;
5. stop treating every minor correction as a new independent review program;
6. allow a ChatGPT Master to author/integrate pushed code while a local lower-cost Codex worker runs native/local tests and reports exact failures.

## 2. Recoverability snapshot

At the 2026-09-07 WIP stop all 30 registered worktrees were clean, all local branch heads matched published origin refs, there were zero stashes/unpublished commits, and required source/review evidence was pushed. Local-only state was limited to reproducible build/test caches, fixtures, probe logs and explicitly retained policy-rejected test residue. No known unpushed gh-tree product source existed at that checkpoint.

The current stop therefore provides a safe remote-first restart point. Any resumed lane must still verify its current remote/local head before modification.

## 3. Important terminology

`M0..M8` are integration milestones/gates. `SLC-01..13` are the selected vertical product behaviors.

A layer contribution can implement prerequisites for several slices without completing those slices end-to-end. This document therefore distinguishes:

- **implemented contribution coverage** — source exists for the relevant layer portion;
- **accepted contribution coverage** — that layer portion has passed its required review/test gate;
- **vertical slice complete** — the full new stack passes the slice end-to-end; this is primarily M5/M6/M8 evidence.

No SLC-01..13 is declared vertically complete merely because M1/M2/M3 source exists.

## 4. Snapshot: existing milestone and slice coverage

### M0 — architecture/design/boundary freeze — COMPLETE

Implemented/accepted:
- accepted v0.4 architecture/design;
- seven frozen layer boundary contracts plus shared boundary types;
- migration map, finding disposition and verification design;
- all SLC-01..13 selected and retained.

Slice effect: defines contracts for SLC-01..13; no vertical product slice completed.

### M1 — Composition CI/bootstrap safety — COMPLETE

Implemented/accepted under #57:
- architecture/dependency guard;
- branch CI improvements;
- native Windows ARM64 and twelve-target build preparation;
- safer publication/bootstrap staging.

Primary slice coverage: SLC-13 prerequisite plus global architecture/build protection for SLC-01..12.

No vertical slice completed.

### M2 — Domain + Application API/ports + State viewmodel — COMPLETE

Implemented/accepted:
- Domain foundation #58;
- Application API/ports #59;
- State-owned viewmodel #60;
- verified serial integration on canonical refactor branch.

Contribution coverage:
- SLC-01..12 receive Domain/API/port/viewmodel foundations as applicable;
- SLC-13 receives shared Domain/config/version/build-facing prerequisites where applicable.

No vertical slice completed because concrete adapters, coordinator/use cases, host and end-to-end new-stack path are not yet fully integrated.

### M3 — concrete adapters — PARTIAL

Existing planned adapter-to-slice mapping from accepted design:

| Adapter | Existing vertical slice coverage |
|---|---|
| Git | SLC-01..08 plus worktree/scope support used by SLC-09..11 |
| GitHub | SLC-01..03, SLC-05, SLC-08 |
| Persistence | SLC-01, SLC-04, SLC-05, SLC-09, SLC-10, SLC-13 |
| Launch Discovery | SLC-09, SLC-10 |
| Runtime | SLC-10, SLC-11, SLC-12 |

Current published contribution status at the stop:

- **Shared M3 prerequisites #66/#67/#68:** COMPLETE/accepted and integrated.
- **Git #61:** facts/read foundation pushed; mutation work remains incomplete; required review is unresolved/blocked.
- **GitHub #62:** all five methods independently accepted; held for Git-first integration.
- **Persistence #63:** substantial implementation/review exists; Windows byte-integrity correction accepted; FreeBSD/profile/source-binding/metadata-integrity/no-birth questions remain open.
- **Launch Discovery #64:** both methods independently accepted and pushed.
- **Runtime #65/#71-related work:** native components and helper/binding portions accepted; combined Runtime base has passing CI; Unix bridge is pushed with local native/race/build proof but exact CI failed; Windows production bridge remains pushed WIP and unverified.

No complete M3 adapter set has yet been serially integrated into canonical `codereview-21/refactor`.

### M4 — Application coordinator -> State -> View — PLANNED

Existing plan:
- implement Application coordinator/use cases/operations against the accepted ports;
- integrate State behavior against Application;
- integrate View against State;
- use fake-port and real-boundary proof.

Slice coverage: SLC-01..12 as applicable.

### M5 — complete new-stack headless vertical harness — PLANNED

Existing plan:
- construct a test-only Composition harness;
- drive State actions -> Application -> real Git/Persistence/Discovery/Runtime -> events -> State -> View;
- use real repositories/worktrees/remotes and actual storage/discovery/runtime;
- identify GitHub transport fixtures explicitly;
- leave production CLI on legacy stack until this passes.

Slice coverage: full vertical verification candidate for SLC-01..12.

### M6 — production host/entry cutover — PLANNED

Existing plan:
- Composition constructor/root/host;
- serial cmd entry cutover;
- both normal and graph entry modes use one implementation;
- real key/event/resize/quit behavior;
- full selected capability matrix through production entry.

Slice coverage: production-host proof for SLC-01..12, especially host/runtime aspects of SLC-10..12.

### M7 — mapped legacy retirement — PLANNED

Existing plan:
- retire legacy source/tests only after M5/M6 succeed;
- mapped replacement tests must exist;
- zero remaining production imports of retired stack;
- strict final dependency/public API/platform inventory;
- freeze integrated candidate for final verification.

Slice coverage: structural replacement confirmation for SLC-01..12.

### M8 — final verification/product PR/main/release — PLANNED

Existing plan:
- full exact-SHA verification;
- product PR/main gate;
- version/source/artifact identity;
- twelve release assets;
- controlled extension install/upgrade;
- publication only after explicit release authority.

Slice coverage:
- final acceptance for SLC-01..12;
- SLC-13 completes only here after delivered install/upgrade evidence.

## 5. Mid-course execution proposal

### M3 — KEEP architecture, SIMPLIFY execution

#### Reuse unchanged
- existing frozen adapter boundaries;
- existing Issues #61..65 scope;
- accepted GitHub and Launch Discovery source/reviews;
- accepted Runtime native/helper/binding components;
- accepted Persistence byte-integrity correction and other already accepted technical corrections;
- existing Git facts/read foundation;
- requirement that unsafe Git mutations/recovery be proven rather than copied from legacy behavior.

#### Change
- stop requiring a fresh author + fresh reviewer + re-review chain for every small subcomponent;
- finish each adapter as one coherent implementation lane;
- one focused technical review per completed adapter, with re-review only for material corrections to review findings;
- integrate accepted adapters as soon as predecessor compatibility allows; do not create extra status-only gates;
- targeted adapter unit/native tests during development; one meaningful M3 integration matrix after all five adapters are integrated;
- use a single concise evidence/status record per adapter instead of repeated source snapshots and duplicate narrative archives.

#### Remove/avoid
- duplicate review archives where Git history + CI + concise review result already prove the same thing;
- full CI for documentation-only status updates;
- review-of-review cycles that make no product change;
- new worktrees unless isolation materially reduces merge risk;
- status commits between two product commits when no durable decision changed.

#### Proposed M3 completion sequence

1. Finish Git #61 mutation/recovery subset required by selected slices.
2. Run targeted Git tests plus safety-critical native/race controls.
3. One focused Git review; correct concrete findings only.
4. Integrate Git into canonical refactor branch.
5. Integrate already-accepted GitHub #62; run targeted compatibility tests.
6. Finish Persistence #63 using explicit supported-profile decisions; unresolved requirements that genuinely require user authority remain explicit rather than spawning speculative implementation.
7. One focused Persistence review and integration.
8. Integrate already-accepted Discovery #64.
9. Finish Runtime production Unix/Windows bridges; investigate current Unix CI failure; run Runtime-native/process tests.
10. One focused Runtime review and integration.
11. Run one complete M3 adapter integration gate.
12. Commit/tag a durable M3 completion point.

### M4 — CONSOLIDATE aggressively

#### Reuse
- Application -> State -> View dependency order;
- accepted APIs, viewmodel and frozen boundaries;
- all SLC-01..12 behavior requirements.

#### Change
Replace many tiny per-feature implementation/review units with three natural contributions:

- **M4-A Application orchestration** — coordinators/use cases/operations for SLC-01..12.
- **M4-B State integration** — state transitions, intent generations, async outcome handling and terminal ownership-facing state.
- **M4-C View integration** — rendering/actions/help/selection behavior against accepted State.

Each contribution gets normal unit tests and one review after the contribution is coherent.

#### Remove
- separate independent review ceremonies for every individual use case/key path;
- repeated full-stack test campaigns before M5.

### M5 — KEEP, but use it as the main integration test stage

M5 is high value and should remain.

#### Reuse
- real new-stack headless harness;
- real Git/worktree/remote/storage/discovery/runtime where required;
- explicit GitHub fixtures;
- SLC-01..12 vertical behavior matrix.

#### Change
Organize tests by coherent behavior families rather than one governance package per slice:

- navigation/read-only: SLC-01..04;
- local mutation/stash/commit/sync: SLC-05..08;
- launch/discovery/session/terminal/cleanup: SLC-09..12.

One complete M5 review after the harness and all three families pass.

### M6 — KEEP, narrow to actual production cutover risk

#### Reuse
- Composition host/constructor/root;
- shared implementation for normal and graph entry modes;
- real key/resize/quit/runtime behavior.

#### Change
Treat M6 as a single cutover contribution plus a focused cutover review.

Unit tests should focus on entry wiring and host lifecycle; do not repeat adapter unit tests already covered by M3 or vertical behavior already covered by M5 except for production-entry-specific regressions.

### M7 — MERGE retirement into a small number of mapped cleanup commits

#### Reuse
- migration map;
- requirement for zero legacy production imports;
- replacement-test requirement;
- final architecture guard.

#### Change
Retire legacy code in 2-4 coherent commits grouped by ownership/package dependency rather than one review process per retired file/finding.

Run architecture/import tests after each retirement group and one final M7 review.

#### Remove
- independent reviewer for mechanical file deletion where replacement mapping and tests are already explicit;
- standalone evidence archive for each retired legacy path.

### M8 — KEEP final rigor, remove redundant repetition

M8 is the right place for the expensive complete matrix.

#### Reuse
- exact-SHA full verification;
- native/cross-platform coverage;
- twelve builds/assets;
- controlled install/upgrade;
- product PR/main/release identity checks.

#### Change
- perform one release-candidate verification campaign on the frozen candidate;
- if corrections are required, rerun only impacted targeted checks during correction and then one final complete campaign;
- one final product review, not separate reviews for every verification family.

Release publication remains explicitly user-authorized.

## 6. Proposed review policy

### Review required

A dedicated technical review is required for:
- Git mutations/recovery/safety boundary;
- Persistence corruption/recovery/metadata integrity boundary;
- Runtime process/session/Windows lifecycle boundary;
- M4 integrated Application/State/View behavior at contribution boundaries;
- M5 complete new-stack vertical harness;
- M6 production cutover;
- M8 release candidate.

### Review normally not required

A separate reviewer is not required for:
- documentation-only status changes;
- mechanical file/package moves with unchanged behavior and passing architecture/tests;
- straightforward wiring already covered by compile-time interfaces and unit tests;
- each individual baseline finding when a shared implementation/test validly resolves a group;
- every small correction after an accepted review unless the correction materially changes the reviewed behavior.

### Re-review rule

Re-review only the changed risk surface plus affected tests. A full fresh review is required only when the correction materially changes architecture, safety semantics or a previously accepted contract.

## 7. Proposed test policy

### During implementation
- focused package/unit tests;
- compile/build for affected packages/targets;
- race tests only where concurrency/lifecycle is relevant;
- native tests only where platform/native semantics are touched.

### At contribution integration
- affected package tests;
- architecture/dependency guard;
- relevant cross-platform compile selection;
- one or two integration paths that prove the contribution is wired correctly.

### Full expensive matrices
Reserve complete cross-platform/native/vertical matrices for:
- M3 completion;
- M5 completion;
- M6 cutover candidate;
- M8 release candidate.

Do not rerun complete matrices for pure documentation/traceability commits.

## 8. Durable commit/push policy

Preferred unit: one commit per completed slice-sized engineering step.

Multiple vertical slices may share one commit when they are implemented by one coherent primitive or when splitting would produce artificial/non-buildable states.

Large steps may use explicit partial checkpoint commits.

Rules:
- test -> coherent commit -> push before the next substantial step;
- never leave substantial completed source only locally;
- completion tags mark actual accepted milestones only;
- checkpoint tags are recovery markers, not completion claims.

## 9. ChatGPT Master + local Codex worker model

This is the preferred lower-cost execution model after Issue #73 approval.

### ChatGPT Master
- owns the milestone/slice plan and keeps scope bounded;
- authors or edits code directly on dedicated pushed branches where practical;
- groups work into coherent recoverable commits;
- performs code review from GitHub diffs/source and decides integration readiness;
- integrates accepted contributions into `codereview-21/refactor`;
- keeps only concise durable status in GitHub/SDP;
- does not create extra review bureaucracy unless risk requires it.

### Local Codex worker (lower reasoning effort, e.g. Spark/standard coding agent)
- checks out/pulls the exact pushed branch;
- builds and runs requested local/native tests;
- reports exact failing command, platform, output and reproducible condition;
- may implement narrowly scoped corrections when explicitly assigned;
- commits/pushes corrections as coherent checkpoints;
- does not independently expand scope, redesign contracts or create new governance programs.

### Iteration loop

1. Master implements one or more coherent slice contributions and pushes.
2. Local worker pulls exact SHA and runs the specified test set.
3. Worker reports pass/fail with minimal useful evidence.
4. Master fixes or delegates a narrow correction.
5. Repeat targeted tests until green.
6. Perform the one required review for that risk boundary.
7. Integrate and move on.

This model deliberately separates expensive reasoning/architecture from cheap local execution/testing.

## 10. Proposed remaining milestone/slice summary

| Milestone | Existing slice set reused | Proposed change |
|---|---|---|
| M3 | Existing adapter coverage of SLC-01..12 + Persistence SLC-13 support | Keep technical scope; collapse review/evidence overhead; complete/integrate five adapters |
| M4 | SLC-01..12 | Three coherent contributions: Application, State, View |
| M5 | SLC-01..12 | Three behavior-family harness groups; one full M5 review |
| M6 | SLC-01..12 | One production cutover contribution/review |
| M7 | SLC-01..12 | 2-4 mapped retirement groups; one final M7 review |
| M8 | SLC-01..13 | One release-candidate full campaign + one final review |

No accepted vertical slice is removed. The correction removes process duplication, not product behavior.

## 11. Explicit non-goals

Issue #73 does NOT:
- weaken frozen Git mutation/recovery safety;
- waive Persistence integrity requirements without explicit user/boundary authority;
- waive Runtime process/session cleanup correctness;
- declare unfinished M3 work accepted;
- declare SLC-01..13 complete early;
- authorize release publication;
- require preservation of every historical review ritual if equivalent engineering evidence is available more cheaply.

## 12. Immediate next action after approval

1. Verify the exact WIP-stop checkpoint and current remote refs.
2. Inspect the failed Unix Runtime CI before further Runtime edits.
3. Resume M3 using the simplified sequence in section 5.
4. Finish Git #61 first because it remains the serial integration predecessor.
5. Do not start new review programs or new milestone planning documents unless a concrete engineering dependency requires them.

This Issue #73 artifact becomes the execution snapshot and process correction for the remaining #21 program while preserving the accepted v0.4 architecture and SLC-01..13 product scope.
