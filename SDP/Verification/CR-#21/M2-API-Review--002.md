# M2 API/ports correction review — second disposition

Disposition: CHANGES_REQUIRED at246e2a6b02e6812cc7579770e5aac973e19b3073.
Authority: #59 under #21; all seven BCs/shared types remain FROZEN1.0.0.
Initial correction4700ea72f080cae1f0a22cf21e80876d88617754 includes productc96ca924
and Master312b534 metadata. The inspected246e successor adds only21 lines in two
files for admitted/newer restart replacement identity. Neither source is integrated.

Independent primary reviewer:m2_api_ports_review; separate Runtime/Storage reviewer:
m2_runtime_storage_api_review. All26 original malformed cases now refuse and226
copy controls pass. M259-M01/M07/M08 are resolved at this source. Five MEDIUM
implementation finding groups remain; all143 original program findings stay open.

| Finding | Remaining reproduced case |
|---|---|
| M259-M02 | A repo-only BranchCreated(A) with no checkout worktree admits unrelated repositoryB Git recovery. Bind every independently known subject, including repository-only outcomes. |
| M259-M03 | Diff A→B accepts A→C result; explicit PR baseB accepts RequestedBaseC; Graph rootsA accepts rootsC. Preserve actual requested semantics even with valid partial/failed payloads and identical correlation. |
| M259-M04 | Complete Stage summary accepts StashThenDeploy. Allowed choice must belong to its operation/root sequence, not merely be a valid enum. |
| M259-M05 | Known BranchCreated or established Running session accepts a sole corresponding explicit NotStarted/VerifiedNoTargetChange facet. Reject the contradictory complete assertion without inventing mandatory missing facets or losing distinct mixed stage facts. |
| M259-M06 | Distinct same-store/family/worktree/root recovery artifacts reject when one omits optional Repository and another supplies its consistent value. Preserve complementary observations and both exact records; reject actual conflicts. |

The [primary report](Evidence/M2-API-Review--002/primary/review-246e.md) contains
seven independently reproduced remaining invalid admissions with positive controls.
The [bounded original report](Evidence/M2-API-Review--002/bounded/bounded-rereview.md)
records one false refusal; its [separate M05 revision](Evidence/M2-API-Review--002/bounded/bounded-rereview--002.md)
adds two Runtime explicit-effect invalid admissions. Original report/manifest
bytes remain preserved, with the supplementary files and hashes separate.

Scope controls matter: empty/unasserted effects are not a finding. A successful
already-up-to-date push can correctly report VerifiedNoTargetChange; stage, restore,
retarget and apply may also be no-ops. Distinct applied/unstarted compound steps
must survive. The missing-parent version hypothesis is not a finding: the request/
native manifest preserves its original Expected anchor, while shared recovery
Original can be the guarded prepublication absence under the established parent.
Future native Persistence proof must retain that authorization history.

Master read actual reports/new probe sources/failed and positive logs, verified88
primary plus66 bounded source/authority Git blobs and exact source copies, and
verified both independent manifest sets. The [archive manifest](Evidence/M2-API-Review--002/ArchiveManifest.json)
maps all retained scripts, source archives, outputs and original hashes. Reviewer
Go sources are inert text. Original preliminary fixture failures remain archived
and are distinguished from product findings.

Full candidate tests and independent positive race/copy/vet controls pass. Source
[CI34041668920](https://github.com/Hans-Einar/gh-tree/actions/runs/34041668920) at4700
and [CI34041943143](https://github.com/Hans-Einar/gh-tree/actions/runs/34041943143)
at246e each passed18 applicable jobs and the explicit Runtime M3 helper skip.
Master checkpointa752367 [CI34041821029](https://github.com/Hans-Einar/gh-tree/actions/runs/34041821029)
also passed18/explicitM3skip. Green CI does not override the independent failures.

Worker corrects M02/M03/M04/M05/M06 within the existing issue, freezes a successor
and supplies exact source/local/current-CI evidence for independent re-review.
Any report-only final checkpoint must preserve reviewed product blobs. Accepted
viewmodeld279c3f remains held for API-first integration. No M2/full Slice/native/
release acceptance, frozen contract change or baseline finding closure occurred.
