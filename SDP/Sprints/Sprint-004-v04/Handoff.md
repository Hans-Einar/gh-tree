# Program handoff

State: ACTIVE — Boundary Contracts; product implementation gated
Repository: Hans-Einar/gh-tree
Parent: #21, open under full user run authority
Sprint/iteration/gate: Sprint-004-v04 / I-02 / G-08
Active authority: #55 (Boundary Contract review/freeze)

## Exact evidence

- Product baseline: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b.
- Main/contract base observed: 4a42222f7bfedc1d80693effbb25a1a82fcff65e.
- All ten focused layers accepted/merged: Runtime #25, TUI State #26,
  Application #29, Git #31, corrected Domain #33, Launch Discovery #36,
  Persistence #40, GitHub #43, TUI View #46, Composition #49.
- Exact report/merge SHAs, CI and all 143 open finding IDs are in CurrentIndex.yaml
  and append-only Ledger.ndjson; findings overlap across reports. Counts 36H/89M/18L.
- Last report: Composition PR #51, reviewed e2ffc1e72c06270bbaf2a20f2c21b949a9b3985c,
  merged e0e716fd44421c221d00375fc0c0d0b8ea85bdff; nine CI checks passed.
- Complete technical REFDES accepted at664f0c051344e3abdfd7d3c5698e4fbd3f584a83;
  final e63547771d89f892ebe8e13f66a1ed8ef0807ba2 accepted/merged in PR54 at
  4a42222f7bfedc1d80693effbb25a1a82fcff65e. No FROZEN BC, implementation
  Slice or v0.4 product change. Design acceptance is not product-finding resolution.
- Release parity requires all twelve v0.3.14 assets, including FreeBSD, 32-bit
  and Windows ARM64; exact names in CurrentIndex.releaseBaseline. Installed
  extension was observed at v0.3.13 and remains unchanged.

## Continue

Read AGENTS.md, developmentInstructions.md, full #21/#52/#55, UserRunContract.md,
index/relations/ledger and all accepted reports afresh as #52 requires. Inspect
actual remote refs and dirty/unpublished work before acting. Preserve user work.

Design work is on codex/cr21-refactor-design in
C:/Users/hanse/GIT/gh-tree-wt/refactor-design. Draft REFDES/API, all 70 baseline
file mappings/90 test-function dispositions, 13 vertical Slices, verification and
143 open-finding mappings are present. DesignReview--001 records two early HIGH
corrections: no-loss publication for every Git worktree rewrite and identity-safe
Unix group signaling. Both proposed corrections are now incorporated: standalone
scratch preparation/retained exclusive publication and a private SID supervisor
with acquired-group signal helpers. Exact probe sources/logs are preserved below
Feasibility/Evidence; complete implementation/platform/crash proof remains future
work. Freeze the complete design HEAD and review it independently. These are design
findings, not resolved product findings.

Initial full review at f1dec60cbd3892f068410f2d7f3caa7f855ba52e / PR #54
returned DES52-H03/H04/M01/M02 (2 HIGH/2 MEDIUM) despite all 18 CI checks passing.
DesignReview--002 records exact findings and proposed corrections. New Storage--001
and CwdAcquisition--001 supply concrete native protocols; API adds PR/stash reads
and downstream lifecycle reservations/ACK transfer. Raw bounded Persistence/cwd
evidence is archived. Re-review of8c109b1e4308eb301d7c1418292044d714c2f33b
accepted H03/M01/M02 at design level, but H04 remained HIGH after real in-place
reparse and pre-Resume tests disproved the guard-only assumption. Its18 CI checks
passed in runs34007362962/34007365156; CI did not override the finding.

WindowsBroker--001 now proposes the complete native startup correction, including
embedded host-native helpers to preserve386-to-native64, the actual cwd-handle
breakpoint barrier, ConPTY and final outer-Job cleanup. Three followup archives
preserve43 checked source/log/build-output hashes and37 committed-text candidates;
temporary executable probes are not product implementation. DesignReview--003
records failed assumptions and the proposed correction. The next permitted action
was independent review of the complete native correction. DesignReview--004 now
records ACCEPT at664f0c051344e3abdfd7d3c5698e4fbd3f584a83, with H04 resolved at
design level and all18 CI checks passing (34022527812/34022530503). All143 product
findings remain open. Final acceptance labels are effective only after PR54 merge.

Current next action: author/reconcile seven BCs and necessary shared type/signature
annex on codex/cr21-boundary-contracts at
C:/Users/hanse/GIT/gh-tree-wt/boundary-contracts. #55 authorizes this document-only
gate; sprint/index/ledger kickoff is recorded before broad authoring. Fresh frozen-
set review/correction, FROZEN1.0.0 metadata review/CI and merge precede product work.

All seven drafts plus BoundaryTypes--001 now exist. BCReview--001 records the
whole-set candidate and mechanical specializations, including25 temporary native
commit cases and Windows batch argv evidence. Next: exact committed candidate
independent review, corrections and re-review; no BC is frozen yet.
Initial review of c291606c2ecfaae15f1bef7b5155d82aa6966712 is now complete:
CHANGES_REQUIRED for BC55-TYPES-01/02, two MEDIUM effect/recovery vocabulary gaps.
BCReview--002 contains the actual review and proposed correction. All18 candidate
CI checks passed in34026364910/34026398215. Current permitted action is independent
exact-HEAD re-review of the correction, not first review or product implementation.
No product implementation before
both gates. Master owns metadata/contracts/integration and delegates bounded
product work to fresh workers with separate reviewers when implementation begins.

No external blocker is known. Windows Go 1.25.0 is selected automatically through
C:/Users/hanse/.codex/tmp/ponsse-r9-go1.23.12/go/bin/go.exe. Baseline tests/builds,
probes and limitations are recorded in implementationNotes.md and actual reports.
Do not infer future conformance from baseline evidence or this handoff.

Checkpoint PR #53 merged source d62d112bd897f63dd14b186356335b3d65170b33
at 58f1cb9eda941db0941cbb8e04e6a0559a3ca896; verified runs
34002140473 and 34002264047 passed all 18 push/PR checks. No design acceptance,
implementation authority or BC freeze follows from that checkpoint.
