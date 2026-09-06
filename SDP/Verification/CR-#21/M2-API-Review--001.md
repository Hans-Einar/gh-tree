# M2 API/ports initial independent review

Disposition: CHANGES_REQUIRED at7e3dd0a3545cc259af6dfe67885b20c8ddc57749.
Product source:3e45f313d3e953f3293b9d42405f6843264656e0; final commit adds report only.
Base:29a12cbd6f2a0e189439b9ca25889f9f566afd6c. Authority:#59 under #21.
Primary reviewer:m2_api_ports_review; bounded second reviewer:m2_runtime_storage_api_review.
Both are independent of author. No candidate edits or integration were performed.

## Consolidated findings — all MEDIUM/P2

| ID | Actual invalid admission and required correction |
|---|---|
| M259-M01 | Saved/discovered/default launch and save values admit a foreign RunConfig version; Retarget accepts a foreign-clone local target. Bind represented worktree/root/version/expected/target/mode relationships, preserving legitimate remote association. |
| M259-M02 | Git RefFact, StatusFacts and successful mutation outcomes admit inconsistent repo/worktree/target/postcondition facts, including absent successful-created Scope/Head. Validate the complete fact/outcome while retaining valid partial/unavailable evidence. |
| M259-M03 | Public StashPop, Deploy, PR diff and status result alternatives admit wrong operation kind, target or labeled worktree. Bind workflow-specific result semantics and exact supplied subjects. |
| M259-M04 | A destructive Retarget confirmation summary can omit its exact destination/target. Require material operation-specific intent/preconditions before valid summary/port issuance. |
| M259-M05 | Common result/terminal effects can erase or contradict known child effects. Preserve facet/subject/stage truth without inventing ordinal effect ordering or simplistic compound-stage equality. |
| M259-M06 | Storage recovery subject and version, Proposed/Current versions, and loaded run recovery can refer to different stores/worktrees/roots. Bind complete represented associations, preserving same-scope content drift, unavailable current and retained evidence. |
| M259-M07 | Runtime result accepts applied resources without admitted Session, Running with Established=false, or restart into another worktree/cwd. Enforce supplied admission/establishment/prior-spec consistency while retaining failed admitted/immediate-cleaned/canceled-before-replacement outcomes. |
| M259-M08 | Unavailable PR base ignores independently known branch/revision scope; a known PR URL can identify another host/repository/number. Validate every retained identity/qualified URL without opaque-token decoding or network lookup, preserving fork heads and unknowns. |

[Primary complete findings](Evidence/M2-API-Review--001/primary/review-findings.md)
contain exact source lines, frozen contract references and26 concrete invalid
admissions. [Bounded Runtime/Storage review](Evidence/M2-API-Review--001/bounded/bounded-review.md)
independently reproduces six related cases with four passing positive/control
groups; these consolidate into M06/M07, not duplicate findings.

The primary reviewer inspected all363 constructor/Valid/Data/Clone record paths,
32 exact union switches and the complete five Client/48 port signatures. All226
direct mutable-container copy probes pass with race; declaration counts, closed
switches and copying do not establish the missing cross-field semantics. Original
suite/vet/race positives pass. Both reviewers retained exact-source probes and
corrected initial archive CRLF/fixture limitations before final reruns.

Master read both complete reports, actual failed/positive logs and source audits,
verified29 product SHA256/Git blobs and the original evidence manifests. The
[archive manifest](Evidence/M2-API-Review--001/ArchiveManifest.json) preserves
independent tests/scripts/reports/logs and the bounded exact-source/authority
snapshot with hashes. Probe sources are inert text, not implementation edits.
Full reports retain initial non-passing fixture/archive attempts as history.

[Product CI34037756152](https://github.com/Hans-Einar/gh-tree/actions/runs/34037756152)
and [final candidate CI34038064464](https://github.com/Hans-Einar/gh-tree/actions/runs/34038064464)
each passed18 applicable jobs and one explicit M3 helper SKIP. This does not
override independent CHANGES_REQUIRED or prove workflow/native/Slice behavior.

Worker is authorized to correct all eight groups systematically within #59,
preserve positive partial/unavailable/compound controls and the complete surface,
then freeze a new exact source for independent re-review/current CI. No frozen
contract, other-layer, shared module/workflow or program-metadata edit is delegated.
Viewmodel #60 is independently accepted atd279c3f but remains unintegrated until
API/ports passes first. All143 baseline findings/full13 Slices and M3..M8 remain.
