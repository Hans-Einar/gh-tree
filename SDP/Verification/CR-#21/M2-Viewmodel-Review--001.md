# M2 viewmodel initial independent review

Disposition: CHANGES_REQUIRED, one MEDIUM finding M260-M01.
Reviewed source: dbb89ea813a954bb8bbb276bea2bb46647826810.
Base: 29a12cbd6f2a0e189439b9ca25889f9f566afd6c. Authority: #60 under #21.
Reviewer: m2_viewmodel_review, fresh and separate from the worker. All18 changed
files were inspected; the candidate remained clean and unchanged during review.

## M260-M01 — simultaneous navigator and branch context

value.go:258-261 permits NavigatorPane only in PullRequestsMode/BranchesMode,
but BranchPane only in BranchContextMode. Snapshot validates both focus and every
supplied pane through these disjoint predicates. Independent tests demonstrate
that no snapshot can retain the existing160x70 navigator-plus-branch cockpit with
either pane focused. Baseline runtime_v0314.go:546-574 explicitly renders those
panes together; frozen wide-layout/SLC-02 behavior retains them.

Correct the representation so both panes and their independent selection, scroll
and partial facts survive focus changes. Navigator PR-versus-branch content intent
must remain explicit, not inferred from a title, selected row or cached content.
No renderer/reducer, new workflow or frozen contract change is needed. Worker is
correcting this under the existing #60 ownership.

## Actual evidence and controls

Candidate package tests independently passed. On the exact Git-blob source copy,
the final independent unit run failed only TestReviewRetainedBranchCockpit.
Five positive independent tests passed under race, including58 recursive-copy
subcases and84 known viewport dimension combinations. Three deliberate copy
mutants were rejected by the specific shared-mutable-storage assertion. Positive
external compiler control passed; four private/type/generation/nil controls failed
as required. Exact40/64 identities, stash parent/tree/index semantics, timestamp
offsets, copied output/gaps and unknown/partial facts passed their controls.

An initial configured-vacant Deploy hypothesis was rejected after actual baseline
and frozen-authority inspection: existing Deploy requires a registered destination,
and CreateWorktree is separate. It is not a finding or new capability requirement.
The earlier failed experiment remains labeled in raw evidence.

The first archive used CRLF conversion and did not match Git blobs; that failed
comparison is retained. Fresh `git -c core.autocrlf=false archive` matched31/31
module/Domain/viewmodel blobs, and final tests were rerun on it. Master verified
all40 original manifest files, final raw test actions, controls, baseline/code
reproducer, report hash and raw archive hash. [Evidence manifest](Evidence/M2-Viewmodel-Review--001/ArchiveManifest.json)
preserves the independent report, source/control/test/log/hash captures and both
labeled archive attempts. Probe source is inert text, not product implementation.
Original reviewer report SHA256:
BED544631856B9634702836D85B8D44379C95E7142FAEE0E9C5968FB704A5AF6.
Verified raw source tar SHA256:
EF602428608B2ECE7A8F9C0FC27B14F5E6CC3B6F84A3B0D170AE319A048E1890.

[Source CI34037022233](https://github.com/Hans-Einar/gh-tree/actions/runs/34037022233)
passed18 applicable jobs/one explicit M3 helper SKIP. Green CI does not override
the independent retained-capability finding. Re-review the corrected exact source
and its current CI; API/ports #59 must integrate before this leaf. No integration,
full Slice,143 baseline finding or State/View/native/E2E/release closure is claimed.
