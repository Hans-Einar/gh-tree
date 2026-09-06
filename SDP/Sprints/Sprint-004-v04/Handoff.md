# Program handoff

State: ACTIVE — Refactor Design; product implementation gated
Repository: Hans-Einar/gh-tree
Parent: #21, open under full user run authority
Sprint/iteration/gate: Sprint-004-v04 / I-02 / G-07
Active authority: #52 (Refactor Design)

## Exact evidence

- Product baseline: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b.
- Main observed: e0e716fd44421c221d00375fc0c0d0b8ea85bdff.
- All ten focused layers accepted/merged: Runtime #25, TUI State #26,
  Application #29, Git #31, corrected Domain #33, Launch Discovery #36,
  Persistence #40, GitHub #43, TUI View #46, Composition #49.
- Exact report/merge SHAs, CI and all 143 open finding IDs are in CurrentIndex.yaml
  and append-only Ledger.ndjson; findings overlap across reports. Counts 36H/89M/18L.
- Last report: Composition PR #51, reviewed e2ffc1e72c06270bbaf2a20f2c21b949a9b3985c,
  merged e0e716fd44421c221d00375fc0c0d0b8ea85bdff; nine CI checks passed.
- No accepted REFDES, no FROZEN BC, no active implementation Slice, no v0.4
  product changes. Review acceptance is not product-finding resolution.
- Release parity requires all twelve v0.3.14 assets, including FreeBSD, 32-bit
  and Windows ARM64; exact names in CurrentIndex.releaseBaseline. Installed
  extension was observed at v0.3.13 and remains unchanged.

## Continue

Read AGENTS.md, developmentInstructions.md, full #21/#52, UserRunContract.md,
index/relations/ledger and all accepted reports afresh as #52 requires. Inspect
actual remote refs and dirty/unpublished work before acting. Preserve user work.

Author design on codex/cr21-refactor-design: REFDES--001 plus necessary API,
physical file ownership/migration and verification appendices. Resolve technical
feasibility with bounded temporary evidence; no real-user-file mutations. Fresh
independent review at frozen exact SHA, corrections/re-review and green CI precede
acceptance/merge. Then authorize BC review/freeze; no product implementation before
both gates. Master owns metadata/contracts/integration and delegates bounded
product work to fresh workers with separate reviewers when implementation begins.

No external blocker is known. Windows Go 1.25.0 is selected automatically through
C:/Users/hanse/.codex/tmp/ponsse-r9-go1.23.12/go/bin/go.exe. Baseline tests/builds,
probes and limitations are recorded in implementationNotes.md and actual reports.
Do not infer future conformance from baseline evidence or this handoff.
