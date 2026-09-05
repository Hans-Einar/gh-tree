# Scrum iterations

State: ACTIVE
Sprint: Sprint-004-v04
Active iteration: I-01 — focused reviews
Active gate: G-03 — Persistence review (#40)
No implementation Slice is active before design and BC freeze.

## I-01 contract

Goal: complete the focused-review inputs required before designing v0.4.
Kickoff state: Runtime, TUI State, Application and Git were accepted; Domain was
in PR #34. Domain and checkpoint scaffolding have since passed their gates below.
Authority: #21, #33; checkpoint scaffolding authorized by #35.

Planned changes: focused review artifacts only, plus strictly necessary SDP
checkpoint metadata. Product code remains frozen at v0.3.14. No BC freeze.
Shared product files are read-only. There are no gh-tree SharedUI rules.

Dependencies: APS22, BR21, LR25, LR26, LR29, LR31; G-01 reviews LR33.
Verification: complete authorizing Issues/comments and accepted inputs; independently
inspect actual report and cited product source; compare finding IDs/counts and
required questions; check exact PR scope/HEAD, review findings and all configured CI.
Completion signal per gate: explicit Master acceptance, merge exact reviewed HEAD,
record review/merge SHAs and CI URL on Issue and in ledger/index/handoff.

## Ordered TODO

- [x] G-01: Domain #33 / PR #34 accepted at corrected HEAD
  86d2a4c0cc0eb8f427d597617246e85473a6cf55; merge 5b1cf181fd3245a65337f1db28ee8f0fb95c200f.
- [x] Establish/merge #35 checkpoint surfaces before broad new work (PR #37,
  merge 11cfef79009bb3956f0ba970470df14843e27e07).
- [x] G-02: Launch Discovery #36 / PR #39 accepted at
  859eab7af576b6b0eb689da6d7ed7c6d235d5853; merge 3b8923e0fad29eec24a2d8fa1811dba24c2a626f.
- [ ] G-03: accept Persistence review, authorized by #40.
- [ ] G-04: authorize and accept GitHub review.
- [ ] G-05: authorize and accept TUI View review.
- [ ] G-06: authorize and accept Composition review, consuming all preceding reviews.

Carry forward: no product finding is resolved by review acceptance. The next
iteration cannot start design acceptance before all ten reports are accepted.

Editorial erratum: Git LR's challenge headings APP29-H03 and APP29-H06 are
transposed. Canonical reconciliation is APP29-H06; confirmation identity is
APP29-H03. Cite the actual finding IDs and meanings when producing REFDES.

## G-02 completed gate contract

Goal: settle normalized launch discovery/provider identity, cwd, saved/default
launch intent and storage boundary before Persistence review. Exact contract: #36.
Files: only SDP/Reviews/CR-#21/Layers/Launch-Discovery/LR--001.md on
codereview-21/layer-launch-discovery-review; Master updates checkpoint metadata.
Inputs: all five accepted preceding layer reviews, BR21 and APS22.
Invariants: frozen v0.3.14 source, no product edits/BC freeze/process ownership.
Verification: independent actual report/source review, exact one-artifact diff,
finding counts and question coverage, green configured CI at frozen report HEAD.
Completion: Master acceptance/merge and recorded exact SHAs, then authorize Persistence.

## G-03 active gate contract

Goal: settle storage/keying/atomicity/migration for user state, active-worktree
preferences and .gh-tree/run.json, consuming accepted Discovery semantics.
Authority and exact question set: #40. Inputs: all six accepted preceding layer
reviews, BR21 and APS22. Files: only
SDP/Reviews/CR-#21/Layers/Persistence/LR--001.md on
codereview-21/layer-persistence-review; Master owns checkpoint updates.
Invariants: frozen v0.3.14 source, no product edits, no adapter-to-adapter calls,
no BC freeze, preserve stored user intent and make failures observable.
Verification: actual report/source inspection independent of author summaries,
question coverage/counts, exact artifact diff and green configured CI.
Completion: Master acceptance and merge with exact SHA recorded; then GitHub review.
