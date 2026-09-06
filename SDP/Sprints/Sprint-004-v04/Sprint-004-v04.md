# Sprint 004 — gh-tree v0.4 architecture program

State: ACTIVE
Authority: Issues #21 and #35; [full user run contract](UserRunContract.md).
Operative workflow: [developmentInstructions.md](../../../developmentInstructions.md).

Deliver the accepted ten-layer architecture, resolve the reviewed safety defects,
preserve v0.3.14 capabilities, verify every planned vertical Slice and release
v0.4.0 only when the exact release commit satisfies the complete release gate.

## Phases and exit gates

1. Finish Domain acceptance and the five remaining focused reviews in accepted order.
2. Accept/merge REFDES--001, then review and freeze the implementation BC set.
3. Implement the accepted vertical Slices through layer-owned branches/worktrees.
   Fresh workers implement; separate fresh reviewers inspect exact source HEADs.
   Master integrates accepted contributions serially and owns shared cutover.
4. Freeze integration HEAD; execute the accepted verification contract including
   real Git repositories, process descendants, deterministic state tests and CI.
5. Merge exact tested final product PR, verify main CI, publish/tag only when
   all release criteria and extension packaging checks actually pass.

The user's same-agent multi-role exception is available but does not waive
independent rereading, findings, SHA freeze or verification. Prefer fresh agents
when available. No product code implementation by Master while workers exist.

## Invariants and scope

- Product evidence baseline: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b.
- Exact selected Revision is authoritative over mutable locators.
- Stash identity is repository plus OID; confirmations revalidate exact state.
- Runtime owns process trees and reports stopped only after proven cleanup.
- Async completions cannot overwrite newer TUI focus/selection/modal intent.
- Preserve all capabilities listed in UserRunContract.md; no cosmetic-only
  package migration, unrelated features, or silent feature removal.
- No frozen BC edits by workers; follow explicit BC-CHANGE/review/refreeze.
- No destructive loss of user files, force push, unverified merge or release.

The active iteration and gate are in [ScrumIterations.md](ScrumIterations.md),
with repository-wide current state in [CurrentIndex.yaml](../../Traceability/CurrentIndex.yaml).
Acceptance of a review report does not mean its product findings are resolved.

## Planned iterations

I-01: remaining focused reviews and program checkpoint establishment.
I-02: reconciled Refactor Design and BC review/freeze.
I-03: complete selected refactor implementation, M1 #57/Domain #58 complete, M2 API/ports #59 and viewmodel #60 active; contract in
SDP/Implementation/CR-#21/Plan--001.md follows accepted REFDES/Slices--001.
The later verification/release iteration executes the full M8 gate.

This is gh-tree-specific process scaffolding. The farmStatistics SDP supplied by
the user was consulted for sprint/ledger shape only; its product rules do not apply.
