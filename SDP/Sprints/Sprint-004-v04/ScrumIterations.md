# Scrum iterations

State: ACTIVE
Sprint: Sprint-004-v04
Active iteration: I-01 — focused reviews
Active gate: G-05 — TUI View review (#46)
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
- [x] G-03: Persistence #40 / PR #42 accepted at
  e13639926ca1b34e5021d1a038b83517cb7eec65; merge db07857550cdfd56c114eb42c045b896b229d747.
- [x] G-04: GitHub #43 / PR #45 accepted at
  ea52f8c5c50e178c6e415cbba2284296e2d9b910; merge e16650036c10f764c1b99c76d98d2bd37c3ffbb0.
- [ ] G-05: accept TUI View review, authorized by #46.
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

## G-03 completed gate contract

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

## G-04 completed gate contract

Goal: settle remote facts/mutations, host/PR/branch identity, observation and
expected-revision semantics before TUI View/Composition and final design.
Authority: #43; consume all seven accepted preceding reviews, BR21 and APS22.
Files: only SDP/Reviews/CR-#21/Layers/GitHub/LR--001.md on
codereview-21/layer-github-review; Master owns checkpoint metadata.
Invariants: frozen v0.3.14 evidence, no live mutation probes/product changes/BC
freeze; GitHub owns no local Git; Application reconciles remote/local sources.
Verification: actual source/report/diff inspection independent of author summary,
complete Issue/template coverage and counts, exact accepted HEAD with green CI.
Completion: Master acceptance/merge recorded; then TUI View review.

## G-05 active gate contract

Goal: settle rendering-only immutable projections, responsive/modal/graph/help/
animation behavior and host measurements without interaction/backend authority.
Authority: #46, all eight preceding accepted layer reviews, BR21 and APS22.
Files: SDP/Reviews/CR-#21/Layers/TUI-View/LR--001.md only on
codereview-21/layer-tui-view-review; Master owns checkpoint metadata.
Invariants: frozen v0.3.14 evidence, preserve accepted UI capability, no product
edits/BC freeze; TUI State owns intent, Application supplies semantic facts.
Verification: actual source/report/render evidence and exact scope, requirement
coverage and finding counts; accepted frozen HEAD must pass configured CI.
Completion: Master acceptance/merge recorded; then Composition review.
