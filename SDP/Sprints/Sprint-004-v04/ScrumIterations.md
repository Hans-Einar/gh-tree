# Scrum iterations

State: ACTIVE
Sprint: Sprint-004-v04
Active iteration: I-03 - complete refactor implementation
Active gate: M2 - API/ports (#59) and State-owned viewmodel (#60)
All SLC-01..13 selected; active parallel leaves supply API/ports and presentation values.

## I-01 contract

State: COMPLETE — all ten focused layer reports accepted and merged.

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
- [x] G-05: TUI View #46 / PR #48 accepted at
  108a3bd0ceb782c6a29d980ce0abcce1eae7a284; merge 00584a57021dda767ead5e3996cc92477e0a2d3c.
- [x] G-06: Composition #49 / PR #51 accepted at
  e2ffc1e72c06270bbaf2a20f2c21b949a9b3985c; merge e0e716fd44421c221d00375fc0c0d0b8ea85bdff.

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

## G-05 completed gate contract

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

## G-06 completed gate contract

Goal: settle CLI/root/host wiring, root lifecycle and cleanup errors, identity,
dependency enforcement and release packaging after all nine preceding reviews.
Authority: #49; inputs are all accepted layer reports, BR21, APS22 and the full
user contract. Files: SDP/Reviews/CR-#21/Layers/Composition/LR--001.md only on
codereview-21/layer-composition-review; Master owns checkpoint metadata.
Invariants: frozen v0.3.14 evidence, preserve twelve release targets, no tag or
release publication/product changes/BC freeze, no workflow logic moved to root.
Verification: actual source/report/workflow inspection, complete question and
finding coverage, bounded probes/build evidence as needed, green exact-HEAD CI.
Completion: Master acceptance/merge recorded; close I-01 only when all ten reports
are accepted, then open dedicated Refactor Design Issue and begin I-02.

## I-02 / G-07 completed contract (historical execution below)

Historical kickoff state: ACTIVE (gate now COMPLETE). Authority: #52 and full user run contract, stages 4-5 and all
downstream capability/safety/verification/release requirements.
Goal: reconcile the ten accepted reviews into implementable REFDES--001,
complete API/ownership/vertical Slice and verification contracts. Read accepted
reports afresh; resolve safety-critical technical feasibility using source or
bounded temporary probes rather than unverified promises.
Files: SDP/Design/CR-#21/REFDES--001.md plus necessary API/ownership/verification
appendices; Master checkpoint metadata. Branch: codex/cr21-refactor-design.
Inputs: APS22, BR21, all ten LR records, 143 open finding IDs (36H/89M/18L,
overlapping reports). No product file change or BC freeze permitted by this gate.
Verification: exact frozen design HEAD, fresh independent artifact review,
cross-contract/API/ownership/compatibility/verification completeness and feasibility
evidence, correction and re-review, design-only diff and green configured CI.
Completion: explicit Master design acceptance/merge. Only then authorize required
BC creation/review/freeze at G-08. Product implementation remains gated until
design and required BCs are accepted/frozen; no implementation Slice is active.

Current design draft: REFDES/API, MigrationMap (70 exact baseline paths and 90
test functions), Slices--001 (all 13 capabilities), Verification--001 and complete
FindingDisposition (143 open IDs). DesignReview--001 records independent early
DES52-H01/H02 corrections and seven migration completeness inputs. Git scratch
publication and Unix supervisor corrections are incorporated, supported by bounded
native probes with source/log snapshots retained in Feasibility/Evidence. The
complete frozen-HEAD review is next. No product finding is closed by this progress.

Initial complete review of PR #54 at f1dec60cbd3892f068410f2d7f3caa7f855ba52e
returned two HIGH/two MEDIUM corrections, detailed in DesignReview--002. Proposed
Storage/cwd/native-evidence and API read/event-transfer corrections are complete;
the corrected exact HEAD must be independently re-reviewed. The initial 18 CI
checks passed, but G-07 remains incomplete until substantive acceptance and merge.

Correction re-review at8c109b1 resolved H03/M01/M02 at design level; H04 remains
pending a complete native startup correction. WindowsBroker--001 now supplies
native-bitness helper delivery, actual cwd breakpoint/identity, ConPTY and final
outer-Job barriers with bounded evidence. Review the new exact committed draft;
do not infer acceptance from its author or earlier18 successful CI checks.

Complete technical design independently ACCEPTED at
664f0c051344e3abdfd7d3c5698e4fbd3f584a83, all18 CI checks successful. G-07 awaits
only final acceptance-metadata exact-HEAD review/CI and PR54 merge. G-08 BC authority
is opened after that merge; no implementation Slice is active yet.

## I-02 / G-08 completed contract (historical execution below)

Historical kickoff state: ACTIVE (gate now COMPLETE). Authority: #55 under #21. G-07 is COMPLETE: PR54 final reviewed
e63547771d89f892ebe8e13f66a1ed8ef0807ba2 merged at
4a42222f7bfedc1d80693effbb25a1a82fcff65e; all18 final CI checks passed in
34022968145/34022971176. #52 is closed. Accepted design is now effective.

Goal: turn accepted REFDES/API/Storage/CwdAcquisition/WindowsBroker and complete
Slice/verification maps into seven mutually consistent implementation BCs1.0.0.
Files: SDP/LayerBoundaryContracts/BC--TUIState--TUIView.md,
BC--TUIState--Application.md, BC--Application--Git.md,
BC--Application--GitHub.md, BC--Application--Runtime.md,
BC--Application--LaunchDiscovery.md, BC--Application--Persistence.md; necessary
shared type/signature annex and review record; Master checkpoint metadata.
Branch/worktree: codex/cr21-boundary-contracts / gh-tree-wt/boundary-contracts,
base4a42222f7bfedc1d80693effbb25a1a82fcff65e.

Invariants: no product changes, no implementation branch/Slice, no silent design
change; all143 product findings remain open. Boundaries keep exact identities,
ownership, operation/session terminal separation, reliable transfer/ACK, acquired
cwd/native broker, storage commit outcomes, pure State/View and full capability.
Master authors/reconciles contracts; fresh independent reviewer reads the whole
frozen set. Findings require correction and exact-HEAD re-review. Only consistent
reviewed contracts receive FROZEN1.0.0 labels, followed by final metadata review,
configured CI and expected-HEAD merge. Record freeze/source/merge on #21 and ledger.
Next gate after completion: issue-authorized Composition CI bootstrap and reviewed
layer Slice contributions on codereview-21/refactor; no premature product work.

G-08 draft progress: all seven contracts and BoundaryTypes--001 are written,
including selected commit continuation/hook/index/ref mechanics and explicit
Windows batch carrier. BCReview--001 records the candidate; full independent
frozen-HEAD review is next. Every contract remains DRAFT1.0.0.

Initial whole-set review at c291606c2ecfaae15f1bef7b5155d82aa6966712 returned
CHANGES_REQUIRED: BC55-TYPES-01/02, two MEDIUM vocabulary/recovery findings.
All18 candidate CI checks passed (34026364910/34026398215). BCReview--002 records
the actual review and proposed shared/Git/Persistence correction. Exact corrected-
HEAD re-review is next; all seven contracts stay DRAFT1.0.0 until accepted.

BC55 complete technical set independently ACCEPTED at
7685494e45c0ef44fbccf9b49a589a90a78026d0; BC55-TYPES-01/02 resolved and
all18 CI checks SUCCESS in34027057276/34027058992. Master records REVIEWED then
FROZEN1.0.0 metadata for seven contracts/shared annex in BCFreeze--001, effective
only after final metadata exact-HEAD review/CI and PR56 merge. One reference-only
Git method-name correction is included for that review. G-08 remains active;
no implementation branch/Slice is active and all143 product findings stay open.

Metadata review at a4d160f returned BC55-META-01 (MEDIUM) for mixed initial/current
review fields. Technical acceptance stands. Corrected index/relations preserve
both exact review SHAs distinctly; metadata re-review/current CI is next.

## I-03 / M1 completed contribution

I-02/G-08 and M0 COMPLETE: BC55/PR56 final independently reviewed freeze
3ac7f20cedd8d14f131b2a15a602b233ef8ca5dc merged at
97cc0d8257603766dd741b49b7d8005857b421a9. Final18 CI checks SUCCESS in
34027493442/34027495276. #55 closed; all seven BCs/shared annex FROZEN1.0.0.

Authority #57 under #21; complete contract in SDP/Implementation/CR-#21/Plan--001.md.
Canonical codereview-21/refactor starts from actual merge above. Fresh worker on
codereview-21/layer-composition-bootstrap owns only the two CI/release workflows,
new internal/composition/architecture tool/tests/README and bounded worker report.
Master keeps sprint/index/relations/ledger current before broad work and integrates
only independently reviewed exact sources with passing branch CI. M1 supplies
SLC-13 build prerequisites; all13 Slices remain selected and none is complete.
No cmd cutover/legacy retirement/other layer/publication in this contribution.
Next after M1 acceptance/integration: issue-bounded reviewed Domain/API/ports/
viewmodel leaves at M2. All143 product findings remain open.

M1 candidate6685b4f76ca071be0125289925e7c9aa466dc10a received independent
CHANGES_REQUIRED for M157-M01..05 (five MEDIUM); actual review/source/probe/CI
records are in SDP/Verification/CR-#21/M1-Review--001.md and its evidence archive.
CI34029160810 finished17 success/1 macOS failure/1 explicit M3 skip. Worker is
correcting root aliases, nested callbacks, private helper boundaries, Git clean
semantics and trimpath toolchain identity. No product contribution is integrated.
Re-review corrected exact source and current CI before M1 acceptance; all143
baseline findings remain open and full Slice verification remains ahead.

M1 corrected candidateb1f533e062c8ab94b81e9309aaffadee096a5c7d has green
CI34030099746 (18 success/1 explicit M3 skip), including real Windows ARM64 and
macOS test/vet/build, but independent re-review is CHANGES_REQUIRED. M01/M02/M05
regressions pass; M03 private-decomposition wrapper traversal and M04 legacy crlf
profile remain, and M06 external child-junction source ownership is new. Actual
sources/results/CI/native evidence is in M1-Review--002 and its archive. Worker
corrects these three before a new exact-source re-review/CI; no integration or
baseline finding/Slice closure. All143 baseline findings remain open.

## I-03 / M2 completed Domain contribution

M1 COMPLETE at accepted source/review/integration
05982a4adb35e39d7b7ba371e0d16d83b2b3674c. Independent M1-Review--003 accepts
all six corrections. Source CI34031420582 and integration CI34031833446 each
passed18 applicable jobs, with explicit M3 helper skip. No baseline finding closed.

Active #58, full contract SDP/Implementation/CR-#21/M2-Domain--001.md. Fresh worker
on codereview-21/layer-domain owns only new internal/domain values/tests/README
and bounded Domain/M2--001.md report. All13 Slices selected; Domain is a shared
foundation and completes no Slice. Separate fresh exact-source review and current
CI precede Master serial integration, then separate API/ports and viewmodel Issues.
No legacy/other-layer/module/frozen-contract/program metadata edit by the worker.

## I-03 / M2 active parallel leaves

Domain #58 COMPLETE: independent ACCEPT at6113ca55b57cb628b016c483070dbb52cd5dd79a,
source CI34033181820 and canonical integration CI34033772732 passed18 applicable
jobs each plus explicit M3skip. All13 Slices and143 open baseline findings retained.
Current contract SDP/Implementation/CR-#21/M2-Surfaces--001.md; full Issues59/60.

Fresh Application worker owns new api/ports only on layer-application-api; fresh
State-viewmodel worker owns only tuistate/viewmodel on layer-tuistate-viewmodel.
Both start from the Master kickoff after Domain integration. Shared Domain values
are the only product dependency of viewmodel; it cannot import API or State parent.
Each source freezes for separate fresh review/current CI; Master integrates API/
ports then viewmodel serially and verifies integrated CI before declaring M2 done.
No workflow/reducer/renderer/adapter/legacy/module/frozen-contract edits delegated.

Current M2 leaves: API/ports#59 final7e3dd0a3545cc259af6dfe67885b20c8ddc57749
(product3e45f31) is under independent review/current CI34038064464. Viewmodel#60
atdbb89ea received M260-M01 MEDIUM despite CI34037022233 green18/M3skip: disjoint
modes reject retained simultaneous Navigator/Branch cockpit. Independent copies,
identity/output/stash/geometry controls pass; rejected vacant-Deploy hypothesis
adds no scope. Actual report/test/control/raw/hash evidence is archived in
M2-Viewmodel-Review--001. Worker corrects mode/content representation under#60;
new exact re-review/CI required. No M2 integration or baseline finding closure.

Viewmodel#60 source d279c3f2af59e48029e77d845f5a55271c7889cf independently
ACCEPTED; M260-M01 resolved, source CI34038307701 green18/explicitM3skip. Full
correction/review/control/raw/hash evidence is in M2-Viewmodel-Review--002. This
leaf waits for accepted API/ports#59 before serial integration; do not integrate
or close60 yet. API#59 current7e3dd0a is still independently reviewed, with a
separate Runtime/Storage sub-review consolidating semantic findings. No M2/full
Slice/baseline finding closure; all143 baseline findings remain open.

API/ports#59 independent complete review at7e3dd0a3545cc259af6dfe67885b20c8ddc57749
is CHANGES_REQUIRED for eight MEDIUM groups M259-M01..M08 (26 invalid admissions).
Separate Runtime/Storage reviewer confirms six cases in M06/M07; positive/race/
copy tests pass, but cross-field semantics do not. Actual full reports/probes/logs/
source hashes are in M2-API-Review--001 and its archive. Source/final CI passed18/
explicitM3skip; this does not override findings. Worker corrects all groups before
new frozen review/current CI. Viewmodeld279c3f is accepted and waits unintegrated
for API/ports. All143 baseline findings/full13 Slices remain; no M2 completion.
