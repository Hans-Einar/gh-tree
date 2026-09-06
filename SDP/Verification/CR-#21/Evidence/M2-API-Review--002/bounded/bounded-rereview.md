# M259-M06/M07 bounded correction re-review

Disposition at exact 246e2a6b02e6812cc7579770e5aac973e19b3073: CHANGES_REQUIRED for one remaining M259-M06 overrestriction. M259-M07's bounded original and additional correction cases pass; no full API, native adapter, Slice or CI acceptance is claimed.

The review began at frozen4700ea72f080cae1f0a22cf21e80876d88617754 (correctionc96ca9244a30e16808743a0cd3e0b557e75080d7) and explicitly advanced to246e2a6 for its two-file restart correction. New code, actual changed tests, mandatory instructions and FROZEN 1.0.0 Runtime/Persistence/BoundaryTypes contracts were reopened. Those governance/contract files were unchanged from the prior reviewed7e3dd0a. Author summaries were not used as review proof.

All 66 original files in this owned snapshot match their246e2a6 Git blob IDs byte-for-byte. Archive explicitly used core.autocrlf=false. source-copy-hashes.json and source-copy-verification.json preserve verification. Actual worktree HEAD was246e2a6; only SDP/Implementation/CR-#21/Application/M2-API--001.md was dirty (worker-owned report). No product or frozen-document edits were made by this reviewer. No parallel State source was read.

## Remaining M259-M06 correction

NewStorageLoadObservation rejects two independently valid StorageRecovery artifacts with the same Family=RunConfig, Store, Worktree and Original StorageVersion/root when only the second's Subject adds Repository=that same Worktree.Repository. The first explicitly omits that optional redundant field. Both have distinct RecoveryIDs and their own locators/identities. The optional absence does not conflict with any supplied identity, but storage_consistency.go:75-76 compares the complete RecoverySubject data structs and returns `storage recovery subject association`.

TestRCStorageEquivalentSubjectObservation reproduces the rejection at correction-final-246e.log with `invalid boundary value: storage recovery subject association`. Correct compatibility checking should retain both original records and their exact optional values while validating overlapping supplied semantic subjects. It must continue rejecting genuinely conflicting store/worktree/root/family identities. The contract does not require identical optional field presence across distinct artifacts: BoundaryTypes--001 B1 explicitly models absence; B5 requires retaining independent recovery facts and rejects inconsistent duplicate IDs, whereas these IDs differ. Persistence BC:473-489 requires distinct artifact IDs, their same physically bound store/family/worktree and lossless propagation. Both supplied records satisfy that exact store/family/worktree contract.

This is a residual of M259-M06, not a new duplicate finding ID. Root, primary reviewer and worker were informed with the actual probe path.

## Corrected and controlled cases

All six prior malformed cases now refuse. Shared NewRecoveryRecord rejects the foreign RunConfig Original worktree before StorageRecovery construction; NewStorageLoadObservation rejects the foreign version/recovery store association before the ports wrapper. These earlier rejections preserve the original malformed semantic triggers and correctly satisfy the original regression tests. Four original positive/control groups all pass: valid admission/failure/cleanup/restart, output and copied byte bounds, JSON/unknown/provider/Null/depth/string retention, and committed-with-error/recovery normalization.

Added controls cover distinct worktree, root, store and document-family versions; compatible content/presence drift; SourceVersion refusal as a document version; multiple distinct artifacts with consistent subjects; corrupt/documentless foreign-root/worktree refusal; unknown root/versions retained; unavailable current and all-version absence; cancellation and known publication; failed admitted starts before cwd acquisition; immediate Cleaned establishment; immutable restart argv/executable/terminal/project/root consistency; positive geometry/source freshness changes; acquired cwd subject mismatch; and246e2a6's present-unadmitted replacement and non-newer ID refusals. Positive canceled-before-replacement and admitted-failed replacement remain valid; failed old cleanup still prevents replacement.

The additional restart diff at246e2a6 was independently inspected: runtime_consistency.go now requires a present replacement's admitted Session and a strictly newer SessionID; the original old-cleanup, predecessor, worktree/cwd/spec checks remain. These checks correspond to frozen Runtime BC:197-206, :227-236, :292-304 and :312-322.

## Expected-absence transition interpretation (not a finding)

An exploratory probe injects the original request's absence-anchor token as RecoveryRecord.Original alongside a Proposed version bound to the newly established parent; the constructor rejects the store difference. That does not establish a defect. Persistence BC:278-288 says result versions bind the established actual store, while :392-394 separately requires the native manifest to preserve the original Expected token. RecoveryRecord.Original may be the guarded prepublication document absence observed after the parent is established. The same probe proves a RunConfigCommit retains the original Expected anchor unchanged and that the recovery record accepts this freshly observed absence plus Proposed under the actual parent. Primary reviewer independently agreed with this reading. No API broadening or BC change is prescribed from this probe; native implementation must still preserve the manifest's input Expected evidence.

## Evidence

correction-final-246e.log/result.json run `go test ./internal/application/api -run '^(TestRS|TestRC)' -count=1 -v` with direct Windows Go1.25.0. All groups pass except the one optional-Repository overrestriction (exit1). Reviewer source files are snapshot/internal/application/api/review_runtime_storage_test.go and review_correction_coverage_test.go. The earlier4700 archive/logs remain at sibling m2-runtime-storage-4700ea7; its initial compile used the wrong reviewer constant name CommitIndeterminate, corrected to StorageIndeterminate before all final runs. No candidate product failure was attributed to that fixture error.

evidence-manifest.json hashes exact source archive, original blob records, final raw log/result, this report and both independent probes. Exact next action is the bounded M06 correction followed by fresh exact-source re-review and the primary complete disposition/current CI gate. Existing source CI cannot override the reproduced remaining review failure.
