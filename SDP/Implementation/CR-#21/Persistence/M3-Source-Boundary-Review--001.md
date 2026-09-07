# M3 Persistence preparation-source boundary review

Disposition: **CHANGES REQUIRED to the draft's scope/remaining-gate wording before
user decision. A separate Windows prepared-byte defect is reproduced.** No user
approval, contract freeze, complete adapter acceptance or integration is granted.

Fresh independent reviewer `m3_persistence_source_boundary_review`, 2026-09-07,
#63 under #21 / Sprint-004-v04 / I-03 / M3. Reviewed BC-CHANGE-PER-SOURCE-01 at the
end of `M3-Adapters--001.md` in canonical
`1b61e977c17479b5891f6ed86d8b1884193844af` (ledger100). Persistence branch
`codereview-21/layer-persistence` was clean/pushed at
`82610b891a82e7b12eb0d13d6be82bc15342f1c4`; Go source/tests equal technical
`ff40e32dcebf36c7182f757b0bc5cb2bdfe08cf8`. Current complete Persistence tree,
including its later README, is `8cc553a579d373754aab1d4ad9aff9d02e1049fd`.

Read AGENTS/developmentInstructions, full #21/#63 bodies/comments, frozen
Application--Persistence/BoundaryTypes/BCFreeze, selected Storage/native feasibility
and Verification, relevant REFDES/API clauses, local README, existing independent
protocol/identity/fault-resource/remaining-gates records and actual source/tests.
`SDP/Instructions` is absent. Summaries were navigation, not evidence. Master
explicitly extended this review's report-only scope to preserve the bounded native
probe in the existing protocol evidence folder after its concrete finding.

## Boundary assessment

The Unix namespace counterexample is sound: `preparation_unix.go:37–41` verifies a
source entry, then `publication_unix.go:15–22` supplies its **name**, not the retained
payload descriptor, to Renameat/Linkat. Parent retention does not bind that name
to the checked object through the syscall. Rename moves a source symlink itself;
Linkat flags0 likewise does not dereference it. No write through a substituted
symlink's referent is asserted. [rename(2)](https://man7.org/linux/man-pages/man2/rename.2.html),
[link(2)](https://man7.org/linux/man-pages/man2/link.2.html).

An explicit cooperative preparation-ownership condition is a coherent possible
**guarantee change**, distinct from the already accepted target-editor gap. Normal
gh-tree instances hold the permanent lock; an ordinary editor restricted to the
fixed target need not know the generated names. Tools that alter preparation
objects/names concurrently would be outside the proposed condition. Random names,
exclusive creation and advisory locking do not enforce it against directory/file
writers. The draft correctly requires user choice and keeps the stronger option
blocked pending a demonstrated mechanism.

Retain the existing PR01 observed-substitution refusals, exact no-follow/identity/
bytes/metadata checks, immutable raw backup, late-written retained original, and
stable recovery observations. A later target edit can make CurrentVersion differ
from ProposedVersion after a real successful publication; it cannot erase that
known effect. An issued operation whose **proposal attribution** is actually
uncertain must instead use the existing valid Indeterminate shape:
PublicationKnown=false, storage EffectIndeterminate, actual known current/proposed
and recovery facts, and precise diagnostics. The API validator at
`storage_validation.go:260–295` forbids Indeterminate plus PublicationKnown=true.
A successful namespace return can be retained as a diagnostic fact without
mislabeling intended bytes AppliedVerified. Neither another postcheck nor blanket
demotion of genuine commits supplies the missing proof. No API change is needed
to express these distinctions.

## Required corrections

**M363-BCS-M01 — MEDIUM, draft scope clarity.** The proposal includes both Unix
source-name ownership and payload byte/metadata noninterference, but its Windows
sentence can leave the separate unprotected content obligation hidden. Retained
handle publication binds the Windows **object**, not immutable bytes/security.
Before asking the user, state explicitly:

> The proposed Unix condition covers generated publication-source names and the
> prepared object's bytes/metadata before publication. Windows retained-handle
> source identity is unchanged; no Windows content/metadata trust exception is
> proposed. Windows prepared-byte integrity remains a separate product gate
> (M363-SRC-H01). Approval of this Unix condition cannot close that gate or #63.

If Master instead proposes a namespace-only choice, remove the Unix byte/metadata
exception and explicitly leave that integrity obligation open on Unix too. Do not
present that narrower choice as closing all preparation-source risks.

Clarify the lifetime as well: the condition ends when the prepared object becomes
the published target at the native publication point. Writes through the now-
published target remain permitted even before syscall-return delivery or request
cleanup completes. The retained `.payload` alias (and `.publication` alias after
absent Linkat) names that same object; its changed bytes are then a recovery
observation, not necessarily prepublication interference. Raw backup and manifest
retention/integrity obligations remain separate. This avoids reading “through the
selected publication call” as a new ban on already permitted target edits.

The affected refreeze must also qualify the existing native-success table and
PublicationKnown prose: success must be attributable to the prepared proposal
under the stated boundary. Merely observing an issued namespace call's success
cannot simultaneously force PublicationKnown=true and permit Indeterminate for
uncertain source attribution. This clarifies the semantic condition without
changing the public Go result shape.

**M363-SRC-H01 — HIGH, Windows prepared bytes can change before publication while
the result claims the intended proposal committed.** `preparation_windows.go:105`
creates owned artifacts with winShareAll; the publisher reopen at line130 also
shares writes. The last byte/policy checks are `commit.go:586–593`; after the
before-publication seam, the native call at598 is followed by known success at601
and AppliedVerified at188. No held resource prevents a peer with existing ordinary
write access from changing the same prepared object in that interval. Windows
sharing controls write opens; access to attributes/extended attributes is a
separate concern. [Microsoft CreateFile sharing contract](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew).

The independent native control writes different valid document bytes through a
separate READ|WRITE/share-all handle to the owned `.publication` object at that
seam, flushes/closes it, and lets the actual class65 publication run. Both absence
and presence produce valid CommittedDurabilityUncertain, PublicationKnown=true,
nil error, different Proposed/Current versions and the foreign bytes at the
target. This is a prepublication payload mutation, not an edit of the published
target or retained original. No privilege, source-name swap or external/user store
is involved. The validated known result requires AppliedVerified under the API.

Required bounded product correction: preserve the prepared bytes through actual
publication under the existing Windows guarantee. Any native guard must cover the
whole relevant lifetime, including earlier acquired writable handles/mappings;
acquiring it only after a final read does not suffice. Keep retained-handle class65,
original-reader/late-write behavior, permissions, independent metadata checks,
truthful outcomes and cleanup. A share-write guard alone is not proof of all
security-metadata exclusion. Choose/prove the concrete mechanism in the author
pass; do not adopt a Windows trust exception, blanket Indeterminate results or a
postcheck-only repair. Fresh independent confirmation remains required.

## Executed evidence and remaining scope

Native Windows NT10.0.26200 / local NTFS, Go1.25.0 windows/amd64, CGO_ENABLED=0,
GOTOOLCHAIN=local. Only owned `t.TempDir` stores were mutated. An additive Go overlay
maps the nonexistent package path `internal/persistence/review_source_windows_test.go`
to the preserved `.go.txt` source; no tracked test/product file is replaced.
Run `go test -overlay <absolute-overlay.json> ./internal/persistence -run
'^TestReviewWindowsPreparationBytesBeforePublication$' -count=1 -v` at the recorded
source. The probe deliberately fails its contract assertion after establishing
all counterexample prerequisites.

| Evidence in existing M3-Persistence-Protocol--001 | Result / SHA256 of Git blob |
|---|---|
| source-boundary-review-windows_test.go.txt | Exact corrected probe; e1357d907e1fd044d77e8afa650dbd365470f6b37dc09283e1218fe68793e2d1 |
| source-boundary-review-windows.log | Both presence modes reproduce; exit1, package0.590s; 010ff26bcbd979230b5f63708c351fd8e1ae2fadaf1f3f9dce9f46a1d556c9e1 |
| source-boundary-review-windows-initial_test.go.txt | Initial access-mask fixture; 82e3c5ba3b521b3e323ad9c6f7f36651f47a4ffd65bde2a81bad7ace5dea1ad6 |
| source-boundary-review-windows-initial.log | Exit1 before mutation: write-only handle lacked access for winOpen's observation; not product evidence; 067aaa93c729e47d06fc51cd1767ef6854e326ccc0570ceb72d4747c8cce11cc |
| source-boundary-review-windows-positive.log | Existing all-family/restart, target-gap/absence competitor, detected target conflict, independent current/cancel and late-original tests PASS, exit0, package0.764s; 5967427184f290d04da14e1a863cedd9c21d999448ecffa536570c1c1a639348 |

Evidence folder: [existing protocol evidence](../../../Verification/CR-%2321/Evidence/M3-Persistence-Protocol--001).
Initial source is preserved by reversing the sole subsequent access-mask edit;
its original log is untouched locally. Git applies ordinary text EOL normalization;
these hashes identify pushed blobs. No new matrix/architecture/FreeBSD execution is
claimed. Previously reviewed ff40e32 CI remains18 PASS / FreeBSD FAIL / helper SKIP.

The separate no-birth inventory makes no false full-profile claim: returned Linux
birth identity, APFS birth identity and accepted NTFS artifact identity evidence
are distinguished from acquisition whitelists and the supplied change-stamp test.
An explicit native no-birth refusal/profile-compatibility gate remains; this review
does not execute or close it. FreeBSD's pending authority/native gate cannot become
read-only support through this inventory. Blocked Git review/access and canonical
serial order are separate and untouched.

Exact next permitted action: Master corrects the DRAFT wording and records the
separate Windows finding; a fresh #63 worker corrects only the bounded product
defect with native tests, push and independent confirmation. The Unix trust choice
still requires explicit user approval and normal affected review/refreeze before
implementation relies on it. This report/evidence are pushed with `[skip ci]`;
product tree, frozen records and all prior work remain unchanged.

## Independent bounded product confirmation — 015cd1a

Disposition: **ACCEPT M363-SRC-H01's Windows prepared-byte correction. Separate
Windows metadata requirement remains OPEN/HIGH (M363-META-H01 below).** This is
not full adapter acceptance. The same independent reviewer reopened frozen
`015cd1a437b1e7fd914383b7943a702f1227ccd8` under canonical0f805a6 / ledger105;
author stopped and branch matched origin. Complete Persistence tree is
`28988772b77247caf6654652c042f6cae20edc73`.

Actual delta from f6d07fb: three product/plumbing files, one Windows test file,
author report and the already recorded metadata probe/log. The Windows payload
uses FILE_CREATE and READ|DELETE sharing from its exclusive birth; it never adopts
an existing object. `commit.go` keeps that exact creator in the request-owned list
beside the independent publisher until class65 returns. Thus both payload aliases
remain covered without reopening a replacement guard or releasing it for the
precommit-close loop. All failure/cancel paths still drain owned resources. The
successful path removes and closes the creator once, before result observation/
delivery; the publisher independently retains object identity. Close diagnostics
join the known result instead of erasing it. Unchanged Unix creation is reached
through a delegating function and constant-false ownership selector: no Unix
guard, close-order, native primitive or boundary exception was introduced.

The continuous creator is material: writable handles/mappings cannot precede an
exclusively created object, and later write-data/append opens conflict while it
exists. A read-only file handle cannot obtain a writable mapping. Ordinary reads,
class65/class11 identity operations and later writes after the creator closes
retain their native behavior. Genuine sharing refusal during the short publication
guard lifetime remains possible under the existing availability contract; no
universal simultaneous-reader/writer guarantee is claimed. Metadata access is
separate from this byte proof.

Native independent verification used the same Windows26200/NTFS and Go1.25.0
host. Additive reviewer and existing metadata overlays replace no product/tests.
Selectors and compact logs are preserved in the same existing protocol evidence
folder. Automatic GC was disabled for the new resource assertions.

| Verification | Independent result |
|---|---|
| Native amd64 existing 70 fault/cancellation cases, external target-gap/absence competitor and retained-original restart controls | PASS in the initial run. That aggregate run exits1 solely because the new reviewer fixture wrongly required nil error on reload after intentionally changing the published payload; its source/log remain preserved. |
| Native amd64 `-race`, corrected reviewer controls plus prepared guard/mapping tests, late-payload recovery and unchanged metadata probe | PASS, exit0, package2.162s. The correction requires the entire reload error tree to contain only the typed retained-payload-change notice/its precise diagnostic, with usable current/document/version; it admits no other error. |
| Same bounded selector, GOARCH=386 native WOW64 | PASS, exit0, package0.668s. |
| Guard lifetime and earlier mapping | 51 absent +63 present actual write/append/generic-write refusals across nine phases; both aliases covered, read mapping retained, postpublication write allowed. A writable view survives closure of its file/section handles; a later deny-write guard refuses, and exclusive creation cannot adopt/truncate that prior mapped object. |
| Added reviewer four controls: both presence modes, successful versus injected creator-close delivery failure | Both prepared aliases reject a separate writer. The first close.artifact callback after publication can write the target before later outcome hooks. Known commit, distinct current/proposal, exact target bytes, stable recovery IDs and exact File-handle baseline survive; injected close failure remains identifiable with CleanupIncomplete. Fresh load preserves the legitimate changed-payload notice. |
| `go vet ./internal/persistence`; diff/source scope | PASS; no product, frozen contract, workflow/module or prior evidence changes by reviewer. |

Replay the corrected reviewer source by mapping the nonexistent
`internal/persistence/review_prepared_bytes_windows_test.go` to
`source-byte-confirmation-015cd1a-windows_test.go.txt`. The separate metadata overlay
uses its existing recorded source/path. Bounded selector:
`TestReviewPreparedByteCorrectionAndGuardClose|TestWindowsPreparedBytesGuardLifetime|TestWindowsEarlierWritableMappingCannotBecomePreparedPayload|TestPreparedByteGuardSeparateMetadataBoundary|TestCommitLatePayloadWriteKeepsIDAndAllowsFreshExpectedIntent`.

| Preserved file (same evidence folder) | SHA256 of Git blob |
|---|---|
| source-byte-confirmation-015cd1a-windows_test.go.txt | 014e097338f8868560af020f9ac88b29dbeae3a94b0dc05f4413b12da4f14d9f |
| source-byte-confirmation-015cd1a-amd64-race.log | 99c67b1779b1fcb14f823b50ecd23b2cbf78d17f027c82642a46305d1e4772b5 |
| source-byte-confirmation-015cd1a-386.log | b57bb6314407cf95a397a10c239b02317372133d7121d6015e8b27da849da445 |
| source-byte-confirmation-015cd1a-initial_test.go.txt | c193d49c8623f13373467dd8e4945181b22ac7301aa69117cf0ef4458f921a4f |
| source-byte-confirmation-015cd1a-initial.log | 30bfeff090df48d3853c1fb33f896ba76a54a50745470b605f9da1e3ff0fc6b0 |

Independently reopened exact [CI34085358603](https://github.com/Hans-Einar/gh-tree/actions/runs/34085358603):
**18 SUCCESS / native FreeBSD FAILURE / Runtime helper SKIP**. The aggregate fails.
[ARM64 job101628239080](https://github.com/Hans-Einar/gh-tree/actions/runs/34085358603/job/101628239080)
confirms source015cd1a, actual Windows/ARM64, Go1.25.0 windows/arm64,
GOHOSTARCH=GOARCH=arm64 and Persistence14.305s; full tests/vet/build pass.
Linux/race/macOS and all twelve build/architecture jobs also succeed. No new full
matrix rerun was needed for these report-only changes.

### M363-META-H01 — HIGH: prepared Windows DACL remains mutable before publication

The reviewer independently executed the preserved
`TestPreparedByteGuardSeparateMetadataBoundary` on native amd64/race and386.
Both absence/presence permit an ordinary READ|WRITE_DAC/share-all handle to change
the prepared object's ordered DACL after `commit.go`'s final metadata comparison,
then publish that exact changed DACL with a valid known committed result and nil
commit error. The probe adds a redundant allow ACE for the same user; it does not
grant a different principal access, change privileges or touch denied fixtures.
Its exit0 means the counterexample prerequisites were proved, not metadata
conformance. The actual scope is the owned preparation object before publication,
not the original/target editor gap.

This is an unresolved **frozen metadata-preservation obligation**, not a request
for a new feature or a side effect introduced by the byte fix.
Application--Persistence's Preparation/publication table and following Windows
paragraph require preserved/verified ordered DACL/protection/inheritance and
supported security metadata; Storage--001 requires supported permissions or
refusal. The present-mode changed DACL no longer matches the observed original's
required policy. A native handle and FILE_SHARE_WRITE exclusion do not freeze
WRITE_DAC; Microsoft also explicitly separates attribute access from share-mode
checks. The independently accepted identity profile and immutable document bytes
therefore do not close this condition. No Windows preparation-metadata trust
exception is frozen, and the pending Unix choice does not create one.

Master must retain this gate until a demonstrated, reviewed preservation/refusal
mechanism meets the existing contract, or an explicit guarantee change receives
the required user authority and affected review/refreeze. Another last-minute
comparison, retaining the old inode alone, or silently treating same-user metadata
changes as out of scope is insufficient. This bounded review chooses no new
mechanism, privilege/profile change or weakened outcome. Wider unexecuted metadata
mutation cases are not claimed proven by the single ordered-DACL counterexample.

M363-BCS-M01's separate wording confirmation remains RESOLVED at canonical
21b512d; the Unix trust choice remains unapproved. No-birth and FreeBSD decisions,
blocked Git review, serial integration and downstream complete adapter/Slice gates
remain unchanged. Exact next permitted action: Master records bounded byte
acceptance and the separate metadata gate, then coordinates only its authorized
next decision/work. This append/evidence are committed/pushed with `[skip ci]`,
leaving the reviewed product tree unchanged and worktree clean.

## Bounded wording confirmation

Independent reviewer m3_persistence_source_boundary_review inspected corrected
canonical21b512d0ca0befcc0a8be0ec7fbb8b032b45d535 read-only and returned:
**M363-BCS-M01 RESOLVED; corrected text is ready for user decision.**
Unix namespace/content assumptions, publication lifetime, permitted later edits,
attribution semantics and the separate Windows defect are distinguished; no
remaining material wording defect was found. No files/tests were changed/rerun in
that confirmation. Master preserves this direct reviewer disposition here.
M363-SRC-H01 remains open. This is wording confirmation only, not user approval,
contract freeze, Windows product acceptance or complete #63 acceptance.

## Bounded Windows metadata proposal review

Independent reviewer m3_persistence_source_boundary_review inspected exact
canonical99dbc840a557e89f0f361b4180b9219895c9f814 read-only and returned
**READY FOR USER DECISION** for BC-CHANGE-PER-METADATA-01. The Windows metadata
condition, publication lifetime, later edits, retained byte/object guards and valid
uncertainty semantics are clear; separate user authority from Unix is explicit.
No material wording defect was found. No files/tests were changed/rerun by this
review. Master records that direct disposition here. M363-META-H01 stays open;
no trust exception, refreeze or adapter acceptance is granted.
