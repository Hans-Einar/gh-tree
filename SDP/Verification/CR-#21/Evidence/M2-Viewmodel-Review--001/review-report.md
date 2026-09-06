# M2 viewmodel independent review — dbb89ea

Disposition: CHANGES_REQUIRED. One MEDIUM finding, M260-M01.
Reviewer: m2_viewmodel_review, fresh and separate from the worker.
Authority: Issue #60 under #21; frozen State/View, State/Application and
BoundaryTypes 1.0.0; accepted REFDES/API/Slices/Verification and State/View reviews.
Source: codereview-21/layer-tuistate-viewmodel at
dbb89ea813a954bb8bbb276bea2bb46647826810, base
29a12cbd6f2a0e189439b9ca25889f9f566afd6c. Actual worktree was clean before and
after review. All 18 changed files were read, including full source, six test
files, consumer README and bounded worker report. Scope is authorized.

## M260-M01 — retain simultaneous Navigator and Branch context panes

Location: internal/tuistate/viewmodel/value.go:258-261, enforced for both
focus and every pane by snapshot.go:36 and :49.

NavigatorPane is accepted only in PullRequestsMode or BranchesMode, while
BranchPane is accepted only in BranchContextMode. Therefore no valid Mode can
publish the retained wide cockpit containing both panes. NewSnapshot rejects a
160x70 snapshot with these two otherwise valid panes when either Navigator or
Branch is focused. Future State/View cannot render this cockpit, move focus with
Alt+N/Alt+B/C/M or retain its pane data without discarding a pane or bypassing the
published validation.

This is a retained capability, not a new design preference: baseline
internal/tui/runtime_v0314.go:546-574 constructs the Navigator on the left and
Branch context on the right together, plus Active/Console; accepted TUI View
LR describes that cockpit; SLC-02 retains branch context/navigation, and the
frozen State/View layout contract says the wide cockpit retains the accepted
panes. The independent TestReviewRetainedBranchCockpit reproduces both rejected
focus choices against the exact frozen source.

Correct the mode/pane representation so one valid snapshot can retain both
Navigator and Branch context and shift focus between them. Preserve explicit
Navigator PR/branch content intent through this representation; a selected row,
title, or the presence of cached rows must not become the only authority for
which navigator content to render. Add mixed-pane and focus-transition value
regressions, including independently retained selection/scroll and partial rows.
Continue rejecting genuinely incompatible focus/modal/input combinations.
This needs no frozen BC change or renderer/reducer implementation.

## Independent checks

Native Windows/amd64; Go1.25.0 executable:
C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe.
All probes are isolated under this report's temporary directory; no candidate
source, issue, branch, PR, tag, publication or user resource was changed.

* `go test ./internal/tuistate/viewmodel -count=1` on the frozen actual worktree:
  PASS independently. Author tests do not cover simultaneous Navigator/Branch.
* `go test ./internal/tuistate/viewmodel -run '^TestReview' -count=1 -json`
  on a Git-blob-verified temporary source copy: exit1, only
  TestReviewRetainedBranchCockpit fails. Final evidence is
  independent-final-unit-raw.json.
* `go test -race ./internal/tuistate/viewmodel -run
  '^TestReview(Copy|Geometry|Stash|Output|ConfiguredMissing)' -count=1 -json`:
  PASS, five top-level independent positive/negative-control tests, including
  58 recursive-copy subcases and 84 known viewport width/height combinations.
  Evidence: independent-positive-race-raw.json.
* Copy checks inspect overlapping slice storage through specs, private records,
  nested option/variant payloads, every substantial aggregate, getters and
  clones. No shared mutable backing found. Three independent source overlays
  intentionally remove scalar-slice copying, nested element Clone, or optional
  Clone. Each fails this review's CopyInventory test with the expected
  `shared mutable interval` diagnostic; the original source passes.
* One external compilation positive passes. Four distinct negative controls fail
  with their intended diagnostics: assigning ContentGeneration to
  ViewportGeneration, converting string to ModalID, setting Snapshot's private
  validity field, and nil ModalBody. See control-results.json/run-controls.ps1.
* Identity controls preserve distinct real 40/64 OIDs, stash/base/index/untracked
  parent OIDs, actual differing tree OIDs, parent index2, explicit absent
  untracked data, and source timestamp instants/original offsets. Nested PR/fork,
  launch alias, optional unknown counts, partial selections and full text values
  are retained in the copy fixture.
* Output controls retain raw bytes and ordered explicit gaps, reject overflow and
  overlapping gaps, distinguish safe interpreted Unicode from C0/DEL/C1/invalid
  UTF-8, and keep source/content/console-presentation/outer-presentation domains
  separate. Measurement checks keep known0/tiny dimensions and refuse stale
  presentation/viewport generations and duplicate geometry.

Source integrity: initial `git archive` inherited core.autocrlf=true and produced
CRLF text, so a raw blob comparison correctly failed. The initial evidence is
retained and is not cited as exact-byte proof. A fresh archive with
`git -c core.autocrlf=false archive` produced 31/31 matching Git blobs (module,
Domain, and viewmodel files) in source-blob-checks.raw.json. All final tests,
race tests and overlay controls above were rerun on that exact-byte copy.
source-raw.tar SHA256:
EF602428608B2ECE7A8F9C0FC27B14F5E6CC3B6F84A3B0D170AE319A048E1890.

Rejected investigation: an initial probe expected ConfirmDeploy for a vacant
configured destination without WorktreeID. Baseline worktree.go:124-153 requires
an observed registered destination, and frozen authority does not require a new
vacant-destination Deploy workflow. This hypothesis is not a finding. The original
failed experiment remains in independent-unit.json; final test names/expectations
explicitly retain missing-WorktreeID refusal. Explicit CreateWorktree remains a
distinct form/plan family.

## Limits and next action

No other substantive leaf defect found after full source/test/contract inventory
and the controls above. No API/ports/parent-State/backend import, callbacks, I/O,
clock read, mutable style global, session allocator, layout/render/parser or
workflow implementation was found. Ordinary supplied text remains verbatim for
future rendering sanitation; copied safe console lines have their own validation.

Master separately reports source CI34037022233 as18 SUCCESS plus explicit M3skip;
this reviewer does not replace that gate or claim full architecture/native/Slice
verification from leaf controls. All143 baseline findings and13 Slices remain
unchanged. Worker should correct M260-M01 within the authorized leaf/report,
freeze a new SHA, then obtain fresh actual-source re-review and current branch CI.
No integration acceptance at dbb89ea.
