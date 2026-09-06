# Program handoff

State: ACTIVE — final boundary freeze metadata review; product implementation gated.
Repository: Hans-Einar/gh-tree. Parent #21 remains open under UserRunContract.md.
Sprint/iteration/gate: Sprint-004-v04 / I-02 / G-08. Active Issue: #55 / PR #56.

## Current authoritative evidence

- Product baseline: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b.
- Main and BC base: 4a42222f7bfedc1d80693effbb25a1a82fcff65e.
- All ten focused reviews accepted/merged; exact report/merge/CI records are in
  CurrentIndex.yaml and Ledger.ndjson. All 143 overlapping product findings remain
  open (36 HIGH, 89 MEDIUM, 18 LOW). No product finding is closed by design/BC work.
- Complete design #52 / PR54: technical664f0c051344e3abdfd7d3c5698e4fbd3f584a83,
  final reviewed e63547771d89f892ebe8e13f66a1ed8ef0807ba2, merged at main above.
  Final18 CI checks SUCCESS in34022968145/34022971176. #52 is closed.
- BC initial candidate c291606c2ecfaae15f1bef7b5155d82aa6966712 received independent
  CHANGES_REQUIRED for BC55-TYPES-01/02 (MEDIUM), despite18 successful CI checks.
  BCReview--002 records the actual findings and corrections.
- Corrected complete BC technical set independently ACCEPTED at
  7685494e45c0ef44fbccf9b49a589a90a78026d0. Both findings and the prior LOW prose
  link resolved. All18 CI checks SUCCESS in34027057276/34027058992.
- BCFreeze--001 records REVIEWED then FROZEN1.0.0 metadata for seven contracts
  and BoundaryTypes--001. It becomes implementation authority only after its own
  exact-HEAD independent metadata review, current CI and PR56 merge. The sole
  technical-text correction is the reviewer's ObserveStatus/StatusFacts reference.
- Accepted design, product source and archived mechanism evidence are unchanged.
  The freeze does not prove implementation, real signing, native platform/fault
  matrices, package-manager integration or any delivered release.

## Exact next action

Read AGENTS.md, developmentInstructions.md, full #21/#52/#55, UserRunContract.md,
accepted REFDES/API/Storage/CwdAcquisition/WindowsBroker/Slices/Verification and the
seven BCs/shared annex/BCFreeze--001. Repository and exact SHAs are authority.

Worktree: C:/Users/hanse/GIT/gh-tree-wt/boundary-contracts.
Branch: codex/cr21-boundary-contracts. Freeze the metadata commit, request the
independent reviewer's exact-HEAD metadata inspection, require all configured CI,
then ready/merge PR56 with expected HEAD. Record source/freeze/actual merge SHAs
on #55/#21 and next repository checkpoint; close #55 only after the gate passes.
Do not implement or create the implementation branch before that merge.

After merge, create canonical codereview-21/refactor from the actual merge and
record I-03/M1 with all SLC-01..13 selected. Open the bounded Composition CI/
bootstrap-safety Issue, explicitly delegate its shared paths to a fresh worker,
and use a separate reviewer. Establish canonical-branch CI, all twelve builds,
native Windows ARM64 coverage and staged architecture checks; disable automatic
bootstrap publication. Exact reviewed branch CI precedes serial integration and
M2 Domain/API/ports/viewmodel work. Master coordinates; workers implement product
code. Full M0..M8 and vertical acceptance are in Slices--001, not folder completion.

## Workspace and release limits

Root worktree C:/Users/hanse/GIT/gh-tree remains on main at the BC base above.
Review/design worktrees preserve accepted source HEADs. The program-checkpoints
worktree retains local801280a7073fa87c04fdc4d06b3a9fbd4d7cd7fc after PR53; do not
reset or discard it. Inspect dirty/unpublished state before integration; exact
current worktree HEADs are discoverable with git worktree list and git status.

No external blocker is known. Native feasibility sources/logs and failed earlier
assumptions are preserved in design/BC Feasibility; compile evidence is distinct
from native proof. Windows Go1.25 is available via the toolchain recorded in
implementationNotes.md. Required native platforms and all12 exact asset names
are in Verification--001 and CurrentIndex.releaseBaseline, including FreeBSD and
32-bit/Windows ARM64. Installed gh-tree was observed at v0.3.13 and is unchanged.
No tag, publication or controlled install/upgrade has occurred in this program.
