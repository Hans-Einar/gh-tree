# M3 Persistence remaining identity and source-publication gates

State: **ASSESSMENT COMPLETE; Master disposition pending. No adapter acceptance.**
Authority: #63 under #21, Sprint-004-v04 / I-03 / M3; canonical
`d55d56f0a50e7c27e8c8b15e22e397826f7601f3`, ledger98. Fresh read-only assessor
`m3_persistence_remaining_assessment`, 2026-09-07. Inspected clean, pushed
`codereview-21/layer-persistence` at `5380629ec23552cdb750f15aeb50bc5260438c5a`;
executable/test source `ff40e32dcebf36c7182f757b0bc5cb2bdfe08cf8`, Persistence tree
`8cc553a579d373754aab1d4ad9aff9d02e1049fd`. Locations below refer to that source.

Read AGENTS/developmentInstructions, full #21/#63 including comments, frozen
Application--Persistence/BoundaryTypes/BCFreeze, accepted Storage/native feasibility/
Verification, relevant design/review authority, current author and independent
protocol/identity/fault-resource reports, and canonical M3-Adapters/Handoff/FreeBSD
assessment. Existing native evidence is reused; source/API reasoning below adds no
new execution claim. No product, tests, frozen records or Master surfaces changed.

## 1. No-birth association: define the supported profile before adding a mechanism

**There is no unconditional frozen requirement to write on every filesystem lacking
birth time.** BoundaryTypes B3:103–105 defines Unix device/inode plus a supported
birth/change stamp as a **short-lived observation**. It does not make ctime a
persistent incarnation identifier. Conversely, Application--Persistence:273–289,
473–480 requires actual parent/store/Expected association and once-persisted
RecoveryIDs that survive restart; matching a path, inode alone or mutable
SourceVersion cannot satisfy this. M3-Adapters' recovery interpretation:195–202
requires proof before claiming no-birth association; it supplies no such proof.

The BC's selected Unix row:402 says **supported local filesystem**, with positive
metadata inspection:415–419 and native proof for claimed profiles:561–569.
Storage--001's scope/publication sections and REFDES:35–41 permit explicit
correctness refusals for unsupported state. Neither the inspected scope authority
nor the twelve release-architecture list commits to every Linux/BSD filesystem or
kernel capability. An acquisition whitelist is not acceptance of every member's
write/recovery semantics (`profile_linux.go:15`, `profile_bsd.go:17`; README's Unix
acquisition paragraph explicitly makes that distinction).

Current `commit.go:232–238,285–290` refuses a change-only native anchor/parent;
reads may still return coherent current bytes. This is an honest unsupported write
profile, not demonstrated corruption. Removing those guards is insufficient:
`binding.go:89–125` and `manifest.go:126–148,355–404` compare parent/Expected/root
observations; `preparation_unix.go:74–86` also records the artifact stamp. Directory
creation, journal writes, hardlinks and publication change ctime. Merely updating
the recorded stamp can lose replacement discrimination and stable restart binding.
Linux birth availability must be established from the returned `STATX_BTIME` mask;
a requested field is not necessarily returned. [Linux statx(2)](https://man7.org/linux/man-pages/man2/statx.2.html)

Existing positive evidence covers birth-capable Linux/ext4, native macOS/APFS
controls, and the separately accepted Windows/NTFS artifact profiles. The supplied
change-stamped run-root test (`store_unix_test.go:55`) is **not a no-birth control**:
it deliberately changes the public observation profile while the native birth
anchor remains available; `manifest.go:400` requires that independent birth proof.
FreeBSD has no accepted successful complete adapter profile and its separate
protected-system-EA decision remains pending.

**Smallest closure:** Master can explicitly inventory the proven write profiles
and their required native identities/metadata, retain exact no-birth refusal and
read behavior, and verify that no separately accepted ordinary-platform/profile
capability is removed. Reuse positive restart/replacement/own-effect evidence and
add only any missing targeted refusal control. This is ordinary technical profile
selection under the existing BC if those conditions hold; it need not become a
project to support every imagined filesystem. This assessment itself selects no
profile. If the inventory excludes a specifically accepted environment or reduces
ordinary platform capability, user scope authority and affected BC review are
required; packaging the same binary is insufficient. In particular, it cannot be
used to close the outstanding FreeBSD gate as read-only support.

If a concrete required no-birth profile remains, a private association must cover
**directories/absence anchors and each artifact**, survive own changes and process
restart, distinguish replacement, preserve original Expected and all historical
RecoveryIDs, and keep public root/change observations exact. Linux offers a bounded
candidate: compare opaque `name_to_handle_at(fd,"",AT_EMPTY_PATH)` handles obtained
from already acquired objects. Its documented identity-comparison use and
filesystem-specific inode/generation encoding warrant investigation on a named
required profile; they do not prove universal support or reboot-stable mount IDs.
No privileged `open_by_handle_at`, inode-only fallback, ID mutation or new public
identity is necessary merely to investigate comparison. **No complete no-birth
mechanism is demonstrated here.** [Linux handle API](https://man7.org/linux/man-pages/man2/open_by_handle_at.2.html),
[Linux v6.18 export encoding](https://github.com/torvalds/linux/blob/v6.18/fs/exportfs/expfs.c#L325)

## 2. Unix source interval: a concrete conditional publication/attribution defect

The accepted target-editor gap is explicit: Storage--001:81–90 and BC:531–536
permit an unobserved **target** write/replacement after final comparison and retain
only the observed original. They do not explicitly authorize publishing a foreign
**source** as the prepared proposal. BC:379–402 requires the fully prepared payload;
BC:321–328 permits Application to publish only known committed intent;
BC:421–427,494–507 requires truthful effects and Indeterminate for uncertain
attribution. B5 independently preserves actual effects/recovery. These obligations
do not require importing Git's stronger target-capture transaction.

The prior independent PR01 correction remains valid: all eight observed regular/
symlink substitutions before final checks or at the before-publication hook were
refused. Actual current code rechecks the source in `preparation_unix.go:37–41`
using `commit.go:629–645`, then calls `publication_unix.go:15–22`. The verified
entry is closed; Renameat/Linkat receive parent descriptors and **names**, not the
verified publisher descriptor or an expected source identity.

The following interleaving uses the same source-substitution capability as PR01,
at a later point; it adds no mount, privilege or hardware threat assumption:

| Step | Established fact / consequence |
|---|---|
| Final source check | Generated publication name names verified regular payload A. |
| Before native name lookup | Another actor moves that entry aside and puts regular B or a symlink at the source name. Retained publisher fd and payload hardlink still identify A. |
| Native call succeeds | Renameat replaces the present target with B; Linkat with flags0 creates the absent target as B. Parent scope is retained and an absence competitor still cannot be overwritten. The source object, however, is wrong. |
| Result construction | `commit.go:598–618,187–194` reports Committed/PublicationKnown/AppliedVerified with ProposedVersion(A) after successful barriers. A later current-read error or different CurrentVersion does not correct the false attribution of applied intent. |

This is a source/API-derived counterexample, **not a newly executed reproduction**.
Linux documents that rename moves the source symlink itself and leaves held fds
unaffected; Linkat flags0 does not dereference it. Apple/FreeBSD also define these
as pathname operations. A successful syscall establishes that a namespace
operation occurred; without source binding it does not establish publication of A.
No write through the symlink to its referent is asserted. [Linux rename(2)](https://man7.org/linux/man-pages/man2/rename.2.html),
[Linux link(2)](https://man7.org/linux/man-pages/man2/link.2.html),
[Apple rename(2)](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/man/man2/rename.2),
[FreeBSD rename(2)](https://github.com/freebsd/freebsd-src/blob/releng/15.0/lib/libsys/rename.2)

**Disposition:** a concrete conformance defect if source substitution remains in
scope, and an unresolved boundary assumption today: the frozen text contains no
source-namespace exception allowing this result. It is not evidence that ordinary
cooperating commits fail. Another precheck, random basename, retained fd, flock,
extra hardlink or postcheck cannot establish atomic source binding. Postcheck
equality cannot exclude an intervening substitution/restoration, and mismatch may
also follow a genuine publication then an allowed target edit. Blindly demoting
those genuine known commits would violate the other half of the outcome contract.

**Smallest next action is a bounded Master boundary decision, not another broad
test campaign.** To keep source-substitution protection, identify a demonstrated
mechanism that binds publication to the prepared object or excludes source-name
mutation through the syscall, while preserving existing permissions, directory
scope and both presence modes. None is established by the selected ordinary Unix
calls. Linux Linkat AT_EMPTY_PATH binds a source fd but requires capability and
only addresses absence; its procfs alternative changes the selected no-follow
route. Renameat2 no-replace/exchange flags still take source names. A protected
staging directory changes the same-directory protocol and is not automatically
protection from every actor already able to edit it. None is adopted here.

Alternatively, an explicit reviewed BC-CHANGE must define who may mutate the
generated staging/recovery namespace, separately from the permitted target-editor
race, with observed drift refusal and precise uncertain-attribution results.
Changing that trust/guarantee boundary requires user scope authority followed by
Master-coordinated affected design/BC review/refreeze; it cannot be inferred from
cooperative target CAS or silently added by a worker. Any detected issued-call
ambiguity must preserve actual namespace effects, current/proposed facts and
recovery, never claim NotCommitted, roll back or replay. Making every ordinary
write Indeterminate is also a capability reduction, not completion.
The existing API (`storage_validation.go:260–263`) ties PublicationKnown to the
two committed outcomes; a correction cannot invent an incompatible mixed result.

## Handoff

Master can resolve the no-birth **profile inventory** without redesign where the
existing scope permits it, and must settle the distinct **source publication
boundary** before claiming that gate complete. No extra native probe was needed
to distinguish these API properties. Accepted Windows identity and bounded
protocol/fault-resource reviews remain intact. Their exact ff40e32 CI has
18 PASS / FreeBSD FAIL / helper SKIP; it does not close full #63. FreeBSD's existing
user decision, blocked Git review/access controls, serial integration and all
full Slice/release gates remain separate and open. This report is the sole change,
committed/pushed with `[skip ci]`; source-tree equality and clean remote handoff
are verified separately.
