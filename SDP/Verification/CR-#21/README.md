# v0.4 verification evidence

CurrentMatrix.yaml tracks execution of the accepted Verification--001 contract
and all thirteen vertical Slices. It is a current-state index, not a replacement
for the accepted design, frozen BC supplements or actual source/test evidence.
Every check starts not_executed. Baseline/design/feasibility CI and probes remain
distinct from v0.4 implementation evidence.

Master records actual evidence after reviewed contributions and integrated runs.
An evidence entry identifies source and integration SHA, command/scenario, actual
platform/tool versions, exit/disposition, fixture/native/cross-build distinction,
raw report/log link and material limits. A command that never ran is not evidence.
Failure/rework history remains in linked reports and the append-only program ledger.

A partial layer contribution may support a check without completing it. Mark a
check complete only when its entire accepted clause, BC supplements, relevant
platform profiles and required integrated behavior are established. Final evidence
must cover the exact frozen integrated HEAD; earlier green source snapshots do not
automatically verify later changes. Domain primitives do not prove Git/Runtime
behavior, cross-builds do not prove native behavior, and baseline CI does not prove
new implementation. Unsupported or environment-limited cases remain explicit.

V-E2E-01..13 map to SLC-01..13. Full new-stack headless and actual host scenarios
remain required at M5/M6 and final verification. SLC-13 prepublication staging
precedes the final product PR; published download/install/upgrade proof follows
the gated main/tag/draft-release/publication sequence. Do not create circular
publication prerequisites or mark delivered behavior from staged builds alone.

All143 baseline findings remain open until their mapped implementation and full
verification are independently established. FindingDisposition.yaml, current
index/relations/ledger and this matrix must agree with actual evidence, not author
summaries. The full long-running goal remains the complete verified v0.4 release.
