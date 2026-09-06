# Independent API/ports review — consolidated findings

Candidate: 7e3dd0a3545cc259af6dfe67885b20c8ddc57749.
Product tree: identical to 3e45f313d3e953f3293b9d42405f6843264656e0.
Base: 29a12cbd6f2a0e189439b9ca25889f9f566afd6c.
Reviewer: m2_api_ports_review, independent of product author.
Disposition: CHANGES_REQUIRED at the complete exact candidate above. The separate fresh bounded Runtime/Storage review independently agrees with M259-M06/M07; its exact-byte rerun and additional Storage cases are consolidated below.
Eight MEDIUM/P2 groups below. No product/BC edits or integration performed.

## Evidence and review scope

Read the actual complete frozen authority set during preparation. Inspected all seven port interfaces, all API record fields and validation/copy bodies, helper/primitive/union/evidence/port code, and author tests. An independent structural audit checked all 363 record constructor bodies plus Valid/Data/Clone patterns and every direct mutable-container copy path. All 32 union validators use exact concrete switches. All 5 Client/48 port signatures conform; complete command/query/event/result inventory exists. No finding concerns mere declaration counts or ordinary constructor layout.

All 226 mutable-container clone cases pass with the race detector; independent nested document/output copies, invalid JSON/surrogate-equivalent duplicate/depth cases and changed-root version rejection pass. The main gaps are semantic relationships between otherwise individually valid fields.

Owned exact clone: exact-source/ at candidate HEAD, core.autocrlf=false, no tracked modifications. Only reviewer tests are untracked there. Product raw Git blob identity is checked (including primitives.go 94eda35e6d680579b85f6c266c29e36cd69aeb5d). Initial git archive inherited Windows line conversion and lacked Git metadata; its full suite correctly failed legacy blob checks. That initial result is a snapshot limitation, not a product finding. All relevant tests were rerun in the exact clone.

Native Go 1.25.0 windows/amd64:
- go test ./... -skip TestReview -count=1: PASS, including the original API/ports and M1 checker/baseline tests.
- go test -race ./internal/application/... -skip 'TestReview(Git|Request|Public|Effect|Storage|Runtime)' -count=1: PASS, original tests and independent nested copy/bounds control.
- go test -race ./internal/application/api -run TestReviewCloneAllContainers -count=1: PASS, 226 direct-container cases.
- go vet ./internal/application/...: PASS.
- Primary independent semantic probes: 24 negative cases incorrectly admitted; positive controls succeed. The bounded review reproduces four of those and adds two distinct Storage association cases, totaling 26 concrete invalid admissions. These fail by expecting constructor rejection, not by unrelated compilation or fixture errors.

Current source/final CI and all12 selections remain Master gates. No whole workflow, native adapter or Slice proof is claimed.

## M259-M01 — bind known request/selection scope before publishing valid values

MEDIUM/P2. Primary discovery_records.go:438; related :610/:721, application_records.go:855/:1615/:1666 and validation.go:386.

A SavedLaunch for worktree A admits a RunConfig StorageVersion explicitly bound to worktree B. DiscoveryRequest/ResolveLaunchRequest likewise accept B's version while receiving A's WorktreeScope; SaveLaunchCommand accepts it as ExpectedStorage, and CurrentDefaultLaunch can carry it into StartLaunchCommand. The validators only check family, despite StorageVersion already exposing Worktree/MatchesRunScope. Separately, RetargetWorktreeCommand accepts an exact LocalCommon target from another clone than its bound destination/Expected.

Frozen Discovery's exact saved/version/worktree binding, Application's invalid scope admission, and B1/B3 distinct local scopes require rejecting these representable contradictions. Check full root identity where a scope is present, worktree identity where only an ID is available, and all supplied local exact-target/mode/expected relationships. Preserve explicitly supported remote-to-local association rather than rejecting valid fork targets.

Repro: TestReviewRequestVersionBindings (one positive, six failing negatives).

## M259-M02 — validate Git facts and successful mutation postconditions together

MEDIUM/P2. Primary git_records.go:1249; related :699/:4625/:4927.

RefFact admits a local branch locator/observation in clone A and a Revision from clone B. StatusFacts admits an Observation for a different worktree in the same repository. WorktreeCreated admits a WorktreeFacts from another repository than its exact resolution, or a MissingWorktree with no Scope/Head at all. StashApplied similarly admits stash A and resulting status B. Each field is valid alone, but the fact or successful closed outcome is false as a whole.

G3 requires explicit exact observed source; G3/G6 require successful mutation scope/Head postconditions rather than using optional inventory fields to avoid proof. Add family-specific cross-field checks throughout the Git facts/outcomes/result envelopes. Do not require unavailable observations to fabricate facts or reject independently valid partial records.

Repro: TestReviewGitCrossFieldScope (positive ref, five failing negatives).

## M259-M03 — preserve exact subjects and true semantics of public result alternatives

MEDIUM/P2. Primary application_records.go:2008; related :2329/:4403/:4478.

StashPopCompleted accepts two RefusedMutation results whose Kind is PushMutation. DeployResult accepts Target revision A while its Resolution.Requested/Local name revision B. PullRequestDiffResult accepts an ordinary CommitTarget rather than a PR target; WorktreeStatusResult labels worktree A but carries B's StatusFacts. These all become valid results and can pass the closed Result family switch.

The frozen public protocol defines matching workflow-specific outcomes and immutable expected targets, not merely sealed record names. Validate applicable result kind/operation/subject/target and compound-step success relations while retaining permitted error-plus-known-effect states. Apply this systematically to public wrappers/projections rather than fixing only the exemplars.

Repro: TestReviewPublicResultSemantics except the separately identified summary case (four failing negatives).

## M259-M04 — require the exact destructive target in confirmation summaries

MEDIUM/P2. Primary git_records.go:4287 (final validation near :4359).

MutationPlanSummary with Kind=RetargetMutation, ConfirmationRequired=true and Choices=[Proceed] is valid with Target absent; its Expected only describes departure state. Nothing in that value identifies the revision the user is approving. Port issuance accepts summaries after matching only operation/kind/version, so later confirmation values can wrap this incomplete semantic summary.

G5 says the safe immutable summary represents the exact same intent and cannot omit a material destructive identity. Require the operation-specific target/subject/precondition facts and check applicable duplications against Expected before a summary is valid. This is structural validation of supplied data, not native registry or UI approval implementation.

Repro: TestReviewPublicResultSemantics/retarget_confirmation_omits_exact_target_and_worktree.

## M259-M05 — retain known effect facts in common operation envelopes

MEDIUM/P2. Primary evidence.go:37; related evidence_records.go:1223 and application_records.go:2624/:5054.

The evidence traversal collects nested effects but checks only dangling IDs/recovery consistency. A StageAllResult containing Git Index=AppliedVerified accepts an empty Outcome.Effects. A successful OperationTerminal containing the correctly populated stage result accepts Effects=[Index=NotStarted]. This creates contradictory common terminal truth despite retaining the nested facts.

B5 and Application's common terminal contract require known applied/partial/indeterminate facets to survive normalization independently of disposition/error/cancellation. Validate or construct the aggregate effect representation so it cannot erase or contradict known child effects. Keep exact subject/stage distinctions and do not invent ordinal effect ordering or force all compound stages into simple equality.

Repro: TestReviewEffectNormalization (positive stage, two failing negatives).

## M259-M06 — bind Storage versions and recovery to the exact run subject/store

MEDIUM/P2. Primary storage_validation.go:224-228; related :292-295 and ports/storage.go:60/:76/:152.

A StorageRecovery with shared subject Worktree A and Original StorageRecoveryVersion explicitly bound to Worktree B is valid when Family=RunConfig and the opaque Store strings match. The validator checks only Family and Store. This mismatches the artifact's recovery identity and exact original document scope while normalization faithfully preserves the wrong association.

Persistence's recovery contract requires one exact subject/store/family plus WorktreeID for RunConfig, and typed document versions. Validate the supplied run-version worktree binding (and compatible complete scope where represented) against the record subject for Original/Proposed; retain the original RecoveryID and typed detail.

The separate bounded reviewer independently reproduced that case and two additional associations: StorageCommitResult accepts Proposed run version A/root1 plus Current B/root2 when Family/Store strings match; LoadedRunConfig(scope A) accepts a Corrupt observation with Version A and Recovery for store B/worktree B. Compare complete supplied run bindings and every retained recovery record to the relevant load/commit subject, without requiring equal current/proposed byte versions. Its valid same-store divergent-current-bytes/committed-with-error/normalized-detail controls pass.

Repro: TestReviewStorageRecoveryScope/artifact_subject_and_version_different_worktree; bounded TestRSStorageWorktreeRecoveryBinding, TestRSStorageCommitVersionWorktreeBinding and TestRSStorageRecoveryLoadBinding. Normative Persistence clauses include exact bound token rejection at lines116-117 and RecoveryRecord/typed detail at470-480/493-495.

## M259-M07 — close contradictory Runtime admission and restart values

MEDIUM/P2. Primary runtime_records.go:660-664/:855-868.

SessionStartResult admits absent Session with RuntimeResources=AppliedVerified, even though absence means refusal before registry admission/acquisition. It also admits a Running snapshot with Established=false. SessionRestartResult accepts a cleaned old worktree A session and replacement worktree B/cwd, as long as the IDs differ and RestartOf points to the old ID.

Runtime requires retained identity after admission, established startup before Running, and replacement from the original copied specification/cwd with only permitted geometry override. Enforce those supplied factual relationships without claiming native cleanup or comparing observation freshness as permanent identity.

Repro: TestReviewRuntimeConsistency (three failing negatives).

The fresh bounded reviewer independently reproduced all three with controls for refused admission, failed admitted cleanup, immediately Cleaned successful establishment, same-worktree replacement and no replacement after cancellation. Runtime BC lines197-206 and319-322 require the supplied identity/establishment/original-spec relationships. Those controls ensure this finding does not ask constructors to erase valid failed-start facts or claim native proof.

## M259-M08 — validate PR base identity even when its repository field is absent

MEDIUM/P2. Primary remote_records.go:453-459; helper validation.go:358.

An unavailable Base endpoint with KnownRepository absent but KnownBranch in repository B is accepted under PRID/base repository A. endpointRepository considers only KnownRepository for unavailable endpoints and ignores the independently explicit branch/revision scopes. Separately, a PR with an Available base repository can supply an unrelated host/repository/number URL and still be valid.

GH3 requires qualified base/head scope, independently validated unavailable identities and validated returned identity/URL. Infer/check available scope evidence from every retained endpoint field, preserve a fork head's distinct scope, and validate the PR URL against known qualified locator/number when available. No repository existence/authentication or opaque-token decoding is requested.

Repro: TestReviewRemoteEndpointIdentity (one positive, two failing negatives).

## Retained repro files

- exact-source/internal/application/api/review_independent_test.go
- exact-source/internal/application/api/review_runtime_test.go
- exact-source/internal/application/api/review_remote_test.go
- exact-source/internal/application/api/review_clone_test.go
- exact-invariants.log, exact-remote.log, exact-clone-all.log
- exact-full-suite.log, exact-race-positive.log, exact-vet.log
- audit.ps1

ReviewManifest.json records raw Git blobs and SHA-256 hashes for all29 frozen product files and this primary evidence set. The bounded review's separate evidence root is C:/Users/hanse/GIT/gh-tree-review-evidence/m2-runtime-storage-7e3dd0a. Read actual review_runtime_storage_test.go and independent-probes-final.log; source-copy-verification.json reports43 source/authority files, zero raw blob mismatches and clean candidate7e3dd0a. independent-probes-final-result.json records expected exit1; positive-controls-result.json records exit0 after exact-byte regeneration. Its initial archive line-conversion problem was corrected before this final consolidation.

Next permitted action: Master authorizes bounded worker corrections, then freezes a new exact candidate for independent re-review of actual changes, complete semantic family coverage and current CI. No acceptance/integration from declaration counts or passing original tests. All143 baseline findings and full Slice/native gates remain unchanged.

Bounded final report: C:/Users/hanse/GIT/gh-tree-review-evidence/m2-runtime-storage-7e3dd0a/bounded-review.md; SHA-256 30DE6C5B4A70C5F5ABD4A7ADFAB9B6B70D6404F68E7F885823C2EA4F8D9015FF. Primary reviewer read the actual complete bounded report, source probes and exact-byte replay results before final consolidation. Candidate worktree rechecked clean at7e3dd0a.
