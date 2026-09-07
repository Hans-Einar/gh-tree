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
