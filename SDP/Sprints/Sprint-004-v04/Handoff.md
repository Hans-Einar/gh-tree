# Program handoff

State: ACTIVE — focused review gates, no product implementation
Repository: Hans-Einar/gh-tree
Parent: #21, reopened under explicit user authority
Sprint/iteration/gate: Sprint-004-v04 / I-01 / G-02
Master owns coordination; Launch Discovery #36 is authorized.

## Exact current evidence

- Product: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b.
- Main governance: 11cfef79009bb3956f0ba970470df14843e27e07.
- Domain #33 / PR #34 / codereview-21/layer-domain-review:
  corrected HEAD 86d2a4c0cc0eb8f427d597617246e85473a6cf55, accepted and merged
  at 5b1cf181fd3245a65337f1db28ee8f0fb95c200f; #33 closed.
- Checkpoint Issue #35 / PR #37: accepted at 8b5514e5e69ed6de603ed5f34e6e530a6a8b76f3,
  merged at 11cfef79009bb3956f0ba970470df14843e27e07; #35 closed.
- No accepted REFDES or FROZEN BCs. All reported product safety findings remain open.

## Continue

Read AGENTS.md, developmentInstructions.md, full #21 comments, the full
UserRunContract.md and the current index/ledger before acting. Inspect exact
remote refs and clean/dirty state; preserve any unpublished or user work.
Domain acceptance and #35 scaffolding are complete. Execute Launch Discovery review Issue #36
in a dedicated review-only worktree/branch using LayerReview.md; consume all
accepted earlier reviews. Follow the ordered gate sequence in ScrumIterations.md.

No actual blocker is known. Go 1.25.0 setup and baseline Windows tests passed;
the absolute bootstrap path is in implementationNotes.md.
Never infer success from this handoff: actual files, Issues, SHAs and evidence win.
