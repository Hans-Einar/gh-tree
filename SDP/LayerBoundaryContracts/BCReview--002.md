# BCReview--002 — initial review findings and correction

State: CHANGES_REQUIRED at initial candidate; correction awaits exact-HEAD re-review.
Authority: #55/#21; accepted design merge 4a42222f7bfedc1d80693effbb25a1a82fcff65e.
Reviewed candidate: c291606c2ecfaae15f1bef7b5155d82aa6966712, PR #56.
Reviewer: fresh independent bc_whole_set_review, read-only; not a contract author.
Configured candidate CI: 18/18 SUCCESS, runs 34026364910 and 34026398215.

## Independent disposition

CHANGES_REQUIRED: two MEDIUM cross-contract vocabulary findings before freeze.

| ID | Finding at reviewed candidate | Proposed correction |
|---|---|---|
| BC55-TYPES-01 | B5's closed EffectFacet omits upstream/local configuration although Git G6 requires it; G7 ReconcileRequest names undefined GitEffectFacet. | Add LocalConfiguration; use shared EffectFacet with an explicit Git reconciliation subset and typed upstream observations. Independently retain verified push and failed/uncertain upstream effects. |
| BC55-TYPES-02 | StorageRecovery has no RecoveryID/shared record, so B5 recovery references and the operation union cannot resolve its artifacts. | Wrap shared RecoveryRecord, define stable persisted artifact IDs and exact owner/store/family/version mapping, and require lossless family-detail/result-plus-error normalization. |

One LOW editorial correction: BCReview--001's prose now formats
`Some[T](value)` as code, removing its unintended Markdown link.

Reviewer independently found the commit specialization consistent with accepted
design: native HEAD-first locking, actual hook context, later message-hook staging
separate from the commit tree and protected index publication. The Windows batch
carrier is also a consistent specialization. Neither needs design-change handling
on the reviewed evidence. Real signing, package-manager integration, native
platform and failure matrices remain implementation verification gates.

Reviewer checked all seven DRAFT 1.0.0 contracts/shared annex, unchanged accepted
design and all 70 baseline blobs/90 test mappings, all 143 open product findings,
preserved ledger prefix, verification references, balanced fences and seven
archived-file plus three available upstream-source hashes. No files changed and
no production tests/probes ran during that independent review.

## Master correction and next gate

The shared/Git/Persistence corrections above are contract vocabulary/normalization
specializations under #55, with no new layer, native publication mechanism or
accepted-design change. They add implementation proof obligations to existing
V-GIT/V-PER/V-APP checks; this document does not claim those tests already exist.
All seven contracts remain DRAFT 1.0.0. BC55-GIT-01's mechanical specialization is
accepted by the reviewer, but whole-set acceptance still requires re-review of
the corrected exact HEAD. Then, and only then, prepare explicit freeze metadata,
review that exact HEAD, require configured CI and merge PR #56. All 143 baseline
product findings remain open; there is no product implementation authority yet.
