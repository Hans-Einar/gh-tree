# API/ports independent correction review at 246e2a6

Disposition: CHANGES_REQUIRED.
Issue #59 under #21. Primary reviewer: m2_api_ports_review.
Rejected predecessor: 7e3dd0a3545cc259af6dfe67885b20c8ddc57749.
First correction reviewed: 4700ea72f080cae1f0a22cf21e80876d88617754 (product c96ca9244a30e16808743a0cd3e0b557e75080d7).
Frozen correction disposition: 246e2a6b02e6812cc7579770e5aac973e19b3073.
246e adds only the inspected Runtime replacement identity correction to4700. The worker's report-only final checkpoint/current CI remain later gates. Subsequent author edits do not change this exact-source disposition.

## Authority and actual review

Read current #59 body/comments and repository sprint/handoff/ledger checkpoint. Root instructions, developmentInstructions, accepted design and all seven frozen BC/BoundaryTypes bodies are unchanged from the previously fully read authority set (actual Git diff checked). Master312b534 metadata/evidence merge is distinct from worker product scope. No frozen contract, signature or legacy/module/workflow change was made by the correction.

Reopened actual new consistency helpers, every changed production hunk/record hook, source tests and adapted regressions. Added MutationPlanSummary fields specialize the frozen semantic plan without adding an adapter or native authority. Complete Client/ports/closed inventory files are unchanged. API/ports own the same immutable values; no native registry, workflow or resource proof is inferred.

Own exact Git clone uses core.autocrlf=false. source-hashes-246e.json verifies88 original source/Domain/contract/module/instruction files against exact raw Git blobs, zero mismatches. Tracked clone is pristine; only clearly named independent reviewer tests are untracked. Candidate was frozen by Master; author worktree may contain subsequent authorized corrections/report edits while this snapshot remains246e.

## Original finding disposition

| Finding | At246e |
|---|---|
| M259-M01 | Original request/version/selection scope cases resolved. Full WorktreeScope root checks and ID-only worktree checks preserve permitted remote/fork target scopes. |
| M259-M02 | Original fact/postcondition cases resolved; repository-only recovery case below remains. |
| M259-M03 | Original result-family/target contradictions resolved; exact query/result matching cases below remain. |
| M259-M04 | Original missing destructive target resolved; operation-inappropriate allowed choice remains. |
| M259-M05 | Child-effect inclusion/union correction resolves original erasure probes and preserves mixed stage facts; explicit known-outcome/contrary own-effect cases remain. |
| M259-M06 | All original foreign version/recovery associations resolved; one overrestriction of equivalent partial subjects remains. |
| M259-M07 | Original failed-admission/establishment/restart continuity cases and246e's present-unadmitted/non-newer replacement corrections pass. Related explicit effect contradiction is consolidated under M05. |
| M259-M08 | Original unavailable base identity/qualified PR URL cases resolved; independent unavailable fork controls remain valid. |

The original26 malformed semantic cases now correctly refuse: primary24 plus the two distinct bounded additions. Primary replay starts from the retained original test sources, not the worker's rewritten assertions; the one shared RecoveryRecord now rejects earlier and the replay accepts that explicit rejection while preserving exactly the malformed subject/version inputs. The bounded reviewer likewise verified its two earlier nested rejections against original semantics. No blanket panic recovery or changed negative input masks a failure. Original positive controls and all226 direct copy cases pass.

## Remaining M259-M02 — repository-only mutation recovery

MEDIUM/P2. git_consistency.go:298-318, particularly the conditional recoveryWorktree call at316.

A valid BranchCreated outcome names a newly created local branch in repository A and intentionally has no checked-out Worktree. GitMutationResult accepts a Git-owned RecoveryRecord whose Subject.Repository is explicitly B. consistentGitMutationResult validates recovery only when an outcome Worktree is available; known repository-only outcomes fall through. This loses exact recovery subject association under G3/G6 and BoundaryTypes B5, without any need to establish native facts in the API.

TestIndependentRepoOnlyMutationRecovery proves same-repository positive and foreign-repository negative. The latter is admitted at246e. Check every independently supplied recovery repository even when worktree is absent; preserve explicitly unavailable subjects and valid partial outcomes.

## Remaining M259-M03 — bind actual query semantics, not only family/repository

MEDIUM/P2. request_result_consistency.go:20 and query switch around160-168.

ValidateTerminalFor accepts all of:
- DiffQuery CommitPair A->B paired with DiffResult A->C;
- a PR diff query requesting exact base B paired with a result whose RequestedBase is C;
- GraphQuery exact roots[A] paired with GraphResult roots[C] in the same repository.

All values are individually valid and the operation correlation is identical. A correct same-comparison control passes. The query switch omits Diff entirely and compares only PR Target or Graph Repository. Frozen Application/Git read contracts retain exact comparison endpoints, explicit base choice and root/traversal/source selection even on partial/failed results. Result family equality is insufficient.

TestIndependentExactRequestResults reproduces the three accepted mismatches. Add matching checks for represented exact query semantics, preserving explicit absence and legitimate partial facts; do not replace the original choice with observed current state.

## Remaining M259-M04 — allowed choices must belong to the operation

MEDIUM/P2. plan_consistency.go:17-24 and the operation switch.

A complete StageMutation summary with StageAction=Stage accepts Choices=[StashThenDeploy]. Its ordinary Proceed/Cancel positive succeeds too. Validation checks enum membership and uniqueness but not operation applicability. Frozen G5/G6 reserve StashThenDeploy for the bound retarget/deploy sequence; an unrelated executable stage plan cannot advertise that continuation.

TestIndependentPlanChoiceKind reproduces the invalid accepted summary. Validate operation-appropriate choices while retaining the original root approval for correctly derived sequence children. No native approval consumption is requested here.

## Remaining M259-M05 — explicit outcome/effect contradictions

MEDIUM/P2. Git validation.go:246 and git_consistency.go:298; Runtime runtime_consistency.go:11-22/runtime_records.go:657-668.

Primary TestIndependentKnownOutcomeEffect shows BranchCreated plus a sole LocalRefsHead=NotStarted or VerifiedNoTargetChange accepted in GitMutationResult. The same known creation with AppliedVerified passes. PrepareBranch is create-only and BranchCreated identifies the created branch under G5/G6, so the known outcome and a sole explicit no-change fact contradict one another.

The bounded reviewer independently reproduced the Runtime analogue at246e: a valid acquired-cwd Running snapshot with Established=true accepts a sole RuntimeResources=NotStarted or VerifiedNoTargetChange. Actual source, log and supplementary report were read by the primary reviewer.

This is a check of supplied facts, not proof that native work occurred. No finding is made merely because an optional/unasserted effect report is empty. Do not blanket-ban an unstarted sibling stage: branch creation followed by refused attachment, or App compound operations, may truthfully retain distinct Applied and NotStarted facts. Partial/unknown evidence and the complete union/provenance correction remain necessary.

Important no-op control: Pushed does not necessarily imply remote ref change. A pinned non-force push can return already-up-to-date success with proven RemoteRefsPR=VerifiedNoTargetChange. Do not expand a guaranteed-change predicate to Pushed, stage, restore, retarget or apply solely because the operation succeeded. This interpretation was communicated to Worker and Master before correction.

## Remaining M259-M06 — equivalent optional recovery subjects are overrestricted

MEDIUM/P2. storage_consistency.go:75-76.

The bounded review's TestRCStorageEquivalentSubjectObservation constructs two distinct recovery artifacts with identical RunConfig/store/worktree/root/version binding. One omits optional Repository; the other supplies that same worktree's Repository. NewStorageLoadObservation rejects the pair because complete RecoverySubject data structs differ. There is no contradictory fact or repeated RecoveryID.

B1 explicit absence and B5/Persistence recovery preservation require compatibility of independently supplied subjects, not identical optional-field presence across different artifacts. Retain both original records and exact presence values; continue rejecting conflicting repository/store/worktree/root/family evidence.

Primary read actual bounded source, log and complete report. Evidence lives separately at C:/Users/hanse/GIT/gh-tree-review-evidence/m2-runtime-storage-246e2a6.

## No finding: initial absence anchor and newly established parent

A hypothesis used the request's original Expected absence-anchor token as RecoveryRecord.Original while Proposed used the newly created actual parent identity. The frozen contract permits Original to be the guarded prepublication absence after parent establishment; the native manifest separately preserves the original Expected token (Persistence BC278-288/392-394). The bounded positive control retains Expected unchanged in RunConfigCommit and constructs valid actual-parent absence/proposal recovery. No API broadening or BC-CHANGE follows. Future Persistence implementation must retain that original Expected evidence.

## Verification and evidence

Native toolchain: C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe, Windows/amd64 Go1.25.0.

- Original candidate API/ports tests: PASS at4700.
- Independent primary original24 semantic cases and226 clone cases: PASS. Rejection moved only for the explicitly preserved Storage malformed nesting.
- Full existing candidate suite at246e, with independently authored failing probes run separately: PASS, including baseline/M1 tests. See full-246e.log for actual command output; additional independent probes were authored separately during review.
- API/ports race run covering original/correction tests and independent replay/copy controls: PASS.
- API/ports vet: PASS.
- Remaining primary negatives at246e: three comparison mismatches, one repository-only recovery mismatch, one invalid stage choice and two explicit BranchCreated effect contradictions. Each has a valid positive control.
- Separate bounded246e: original six malformed cases and four original controls pass; additional binding/Runtime/continuity controls pass; one equivalent-subject positive wrongly refuses, and two supplementary explicit established Runtime effect cases wrongly admit.
- Client5/ports48 and32 closed union declarations remain unchanged. All12 platform selections/current branch CI are Master gates; this report makes no unrun native/whole-Slice claim.

Primary files: source-4700-product.tar, source-246e-product.tar, source-hashes-246e.json, independent-original-replay.log, independent-246e-remaining.log, independent-246e-plan-choice.log, independent-246e-outcome-effect.log, full-246e.log, race-246e-positive.log, vet-246e.log and exact-source/internal/application/api/independent*.go. ReviewManifest-246e.json binds exact hashes.

Bounded records: bounded-rereview.md (SHA2569935BED8E87FD033731D04D03A8901D89310C659482F7922F302C4924A4F38E3), bounded-rereview--002.md (SHA256B71D24AA70A3D8849A7E7CA29FFD1596D43BCE91187C9D52672834936A90297D), original and supplementary manifests/probes/logs. Both are independent of the product author.

Next action: correct remaining M02/M03/M04/M05/M06 within existing #59 scope; preserve these exact rejected snapshots/evidence; freeze successor and independently re-review actual changes, valid no-op/partial/compound controls, report-only checkpoint and current CI. All143 baseline findings, full M2/Slice/native/release gates remain unchanged.

