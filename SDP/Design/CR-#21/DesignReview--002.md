# DesignReview--002 — complete frozen design review

Disposition: CHANGES_REQUIRED.
Authority: #52 under #21; PR #54.
Reviewed HEAD: f1dec60cbd3892f068410f2d7f3caa7f855ba52e.
Reviewer: fresh independent frozen_design_review, read-only.
CI at reviewed HEAD: all 18 checks successful, runs 34005955138/34005970811.
CI does not override substantive design findings.

The reviewer reread mandatory authority, all ten accepted focused reports and
complete normative design artifacts, inspected archived probe sources/logs and
checked accepted source identities. No author summary was used as review evidence.
No probe/product test, file edit or external write was performed by the reviewer.

| ID | Severity | Frozen source anchor and finding | Required correction |
|---|---|---|---|
| DES52-H03 | HIGH | REFDES:388 / API:296 defer Persistence's native replacement and scoped path mechanism to a future review/test. Commit points, existing/missing destination, native locks/permissions/recovery and supported filesystem profile remain undecided. | Select concrete native mechanisms and bounded evidence; define exact visibility/durability/indeterminate classifications and path authority. Exhaustive implementation faults remain future proof, not a substitute for the design. |
| DES52-H04 | HIGH | REFDES:374 / API:280 / Runtime feasibility:35 validate cwd by path before later native child creation. A directory/link replacement can select another project before CreateProcess/chdir consumes the path. | Hold effective path guards or acquire the directory object through actual child creation; define unsupported/refusal and object-scope limits. Test replacement in that precise handoff interval. |
| DES52-M01 | MEDIUM | API:124/126/211 closed inventory lacks PR merge-base diff and exact stash patch/files inspection, although Slices:52/55 require them and baseline DiffPullRequest/StashPatch implement them. | Add typed requests/results/Git primitives, exact PR head/base and StashID semantics, linked SLC-03/06 checks. |
| DES52-M02 | MEDIUM | API:90 reserves only operation events; API:271 retains Runtime cleanup until App ACK without bounding reliable normalized-event transfer. | Reserve downstream session lifecycle capacity before admission or equivalent bounded transfer; no early ACK/loss/eviction/unbounded queue. Test full App queue, natural exits and host detach. |

Earlier DES52-H01/H02 corrective mechanisms are accepted design directions:
retained capture/no-replace Git publication and acquired-group Unix helpers address
their original races. Full native filesystem/protocol/crash/platform implementation
evidence remains mandatory; these are not resolved product findings.

Independent structural checks passed: all 70 baseline paths, 90 test functions,
143 open findings/counts, accepted report blob identities, references, append-only
ledger, acyclic dependency graph and seven planned BCs. Product code is unchanged.

## Correction pass — independent re-review pending

| Finding | Incorporated correction and bounded evidence |
|---|---|
| DES52-H03 | Storage--001 selects native lock/replacement/no-replace/path/permission/recovery/commit-point protocols and schema1 migration. Feasibility/Persistence provides a separate fresh reviewer's concrete proposal, six native Windows and six unprivileged native Linux cases. Root read the actual full proposal, both probe sources/logs and verified all seven archived source/module/log/proposal hashes. External-editor and Unix directory-object/current-path limits are explicit. |
| DES52-H04 | CwdAcquisition--001 selects Unix acquired-FD -> supervisor Fchdir -> inherited child cwd and Windows effective component no-delete-share guards through CreateProcess. Root native temporary Linux UID65534 and Windows probes passed replacement-in-the-handoff cases; sources are archived. Native full Runtime/platform/fault proof remains required. |
| DES52-M01 | API adds PullRequestDiff and StashPatch typed reads, exact endpoint/parent semantics and Git MergeBase/ReadStashPatch primitives; V-GIT-10 and SLC-03/06 cover execution. |
| DES52-M02 | API reserves shared App operation/session capacity before Start/Restart, transfers cleanup into its reserved downstream slot before explicit Runtime AckEvents, deduplicates replay and retains bounded pending acknowledgments; V-APP-01 covers full queues/natural exits/host detach. |

These are proposed corrections, not self-acceptance. Freeze the corrected complete
HEAD, independently re-review all changes and require green CI before acceptance.
No BC freeze or product implementation is authorized by this correction record.
