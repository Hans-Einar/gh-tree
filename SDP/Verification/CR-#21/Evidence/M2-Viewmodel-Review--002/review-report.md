# M2 viewmodel independent re-review — d279c3f

Disposition: ACCEPT. M260-M01 is resolved; no remaining review finding.
Reviewer: m2_viewmodel_review, separate from the worker; no implementation edits.
Authority: Issue #60 under #21, effective FROZEN 1.0.0 State/View,
State/Application and BoundaryTypes; accepted REFDES/API/Slices/Verification
and State/View reviews, unchanged from the initial independent review.
Reviewed source: codereview-21/layer-tuistate-viewmodel at
d279c3f2af59e48029e77d845f5a55271c7889cf. Previous rejected candidate:
dbb89ea813a954bb8bbb276bea2bb46647826810. Worktree clean before and after review.

## Source and correction assessment

Inspected the complete nine-file correction, new cockpit_test.go, updated consumer
README/report, current source around every changed value/validator and actual
#60 correction authority. The original complete value/source/test inventory was
read at dbb89ea; exact Git comparison verifies that all other product source,
Domain and frozen design/contracts remain unchanged. The total contribution is
19 files: 11 production Go files, six test files, README and worker report.
Inventory erratum: the initial reviewer report said six test files; that candidate
actually had five. This correction adds the sixth, cockpit_test.go. No source was
omitted from the initial inspection.

Only value.go and pane.go change production behavior. Mode now selects
Cockpit/History/Graph/Diff, both Navigator and Branch panes belong to Cockpit,
and NavigatorModelSpec.Content independently requires the closed
PullRequestsContent/BranchesContent value. Copies retain that scalar with the
complete record. Titles, cached rows, selected identity and current focus are not
used to infer or overwrite content. This directly resolves the contradictory
mode predicates without a renderer/reducer, backend, new capability, authority
leak or frozen BC change. No accepted downstream State/View implementation
depended on the rejected candidate's conflated enum values.

The correction retains the existing private values, options, exact Domain
identities, copied records, scope/tag checks, console source/content/presentation
correlation, known-zero dimensions, modal/input validation and measurement rules.
No new forbidden import, callback, I/O, clock, mutable style state or allocator
is introduced. The bounded worker report accurately records the earlier failure
and does not claim independent acceptance from its own tests.

## Independent correction proof

The original TestReviewRetainedBranchCockpit now passes for both focused panes.
Its fixture was adapted only from the three rejected cockpit-mode enum names to
CockpitMode and explicit Navigator content, preserving its required simultaneous
pane behavior.

New TestReviewCorrectionRetainsIndependentCockpit constructs Navigator + Branch
+ Active + Console at160x70. It exercises80 successive focus/content/cache
combinations: two explicit navigator content types, empty/both/branch-only/PR-only
caches, and ten pane/subpane focus destinations. Constant branch-like title and
the same folder selection are used for both content types. Four independent
semantic selections, filters, cursors, list/message/detail scroll, partial source
facts, actual session identity and output generations survive unchanged. These
are supplied next values, not an implemented reducer or simulated key transport.

The fixture proves invalid History/Graph/Diff/zero/out-of-range modes reject this
cockpit, foreign-mode focus rejects, and unowned terminal-input focus rejects.
Supplying input ownership for the actual selected running console succeeds;
moving focus elsewhere without clearing that ownership rejects. The modal
regression separately confirms current measurements, stale modal refusal,
single modal focus, invalid form focus and zero-area confirmation refusal.

Three semantic overlays independently challenge the new tests:

* deny BranchPane: fails with `retained simultaneous cockpit rejected`;
* overwrite Content with PullRequestsContent: fails with
  `explicit navigator content lost`;
* remove Content.Valid checking: fails with the specific invalid NavigatorModel
  admission diagnostic.

These expected failures are in correction-control-results.json, corresponding
logs/overlays and run-correction-controls.ps1. All three controls pass; the actual
unmodified source passes. No candidate source was edited for the controls.

## Actual verification and evidence

Native Windows/amd64, Go1.25.0 executable:
C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe.

* On actual clean d279c3f: `go test ./internal/tuistate/viewmodel -count=1`,
  `go vet ./internal/tuistate/viewmodel`, `go build ./internal/tuistate/viewmodel`
  and `git diff --check dbb89ea..HEAD`: PASS independently.
* Exact source closure copied with `git -c core.autocrlf=false archive`;32/32
  files (module/Domain/viewmodel) have matching raw Git blobs in
  source-blob-checks.json. source-raw.tar SHA256:
  C92BF6A14933E8EC39A83A0CCCCCA9DC3A6064D03FCFD16817377B5575BB38D0.
* `go test ./internal/tuistate/viewmodel -run '^TestReview' -count=1 -json`:
  PASS, all eight independent top-level tests; independent-unit.json.
* `go test -race ./internal/tuistate/viewmodel -run '^TestReview' -count=1 -json`:
  PASS, all eight independent top-level tests; independent-race.json.
* Original 58 recursive-copy subcases remain, covering nested specs, optional
  records, closed variants, access and clones through significant aggregates.
  Three shallow-copy mutants still fail with the expected overlapping-storage
  diagnostic. Copy overlays/control-results.json retain the actual failures.
* Original exact40/64 OID/stash/indexed-parent/real-tree/timestamp/offset controls,
  raw output/gap/control-text checks and84 known viewport combinations all pass.
  No prior negative assertion was weakened to accept the new candidate.
* One external compiler positive and four negatives pass: generation-domain
  assignment, scalar ModalID conversion, private Snapshot field and nil body.
  Corrected controls replace both independent test files for isolated compiler
  snippets, preventing unrelated missing-fixture-symbol failures.

All tests and overlays are confined to this temporary review directory. One first
attempt had an unused import in the new reviewer fixture; that compile error is
retained as independent-fixture-build-initial.json, corrected only in the temporary
test, then the complete tests/race/controls reran successfully. It is not a product
failure or passing verification result. The initial review's rejected hypothetical
vacant configured Deploy requirement remains rejected; the negative missing
registered WorktreeID control continues to pass.

## Acceptance limits and next action

ACCEPT applies to the complete M2 viewmodel leaf at exact d279c3f and resolves
M260-M01 at this source. Master must still verify current source CI, integrate
accepted API/ports first and viewmodel second, and obtain exact integrated CI
before completing M2. This review does not accept State reducer, View layout/
render/parser, actual measured terminal behavior, any whole Slice, native Runtime
or release. All143 baseline findings and13 selected Slices retain their existing
status. No PR, merge, issue closure, tag or publication was performed by reviewer.

Master owns durable archive/verification/index/relations/ledger integration.
Initial failed review/evidence remains historical and must not be overwritten
by this accepted correction.
