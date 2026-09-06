# Git mutation safety feasibility — Issue #52

State: DRAFT — design input; not an implementation or verification acceptance.
Authority: #52 under #21, Sprint-004-v04 / I-02 / G-07.
Author role: fresh bounded Git safety feasibility reviewer.
Product evidence: v0.3.14 / `f626077ca0e59fbe9ede7ba1116982bb94b2eb6b`.
Design checkout when inspected: `codex/cr21-refactor-design` at
`e0e716fd44421c221d00375fc0c0d0b8ea85bdff`.
Date: 2026-09-06.

**Superseding safety decision:** the final addendum below replaces every earlier
recommendation to run switch/read-tree/stash cleanup against the live worktree or
rely on quiescent external writers. Those earlier paragraphs are retained as
rejected feasibility history, not implementation authority. All file-changing
operations must use scratch preparation plus protected publication.

Read: AGENTS.md, complete developmentInstructions.md, full Issues #52/#21,
#22/#33, UserRunContract, APS22, BR21 and all ten accepted focused reports.
Git/Application/Domain/Persistence are the principal inputs. APP29-H03 denotes
confirmation identity and APP29-H06 denotes reconciliation; the Git report's
transposed headings do not change their meanings. No layer-local instructions
apply to this new appendix. The checkout's earlier G-06 metadata is a historical
checkpoint; #52 authorizes this design work and Master owns its update.

## Decisions for the integrated design

| Requirement | Implementable recommendation | Limit that must remain explicit |
|---|---|---|
| Exact stash apply | `git stash apply --index <full stash OID>`; pin the object before applying. | Worktree/index changes can conflict or partially complete. Applying is not deletion. |
| Exact stash pop | Application sequences exact apply, verifies result, then exact drop. | A successful apply followed by refused/failed drop is `AppliedStashRetained`; never apply again automatically. |
| Exact stash drop with shifting reflog | A narrowly scoped **files ref backend** deletion primitive locks `refs/stash` before reading the log and selects by OID under that lock. Use the source-aligned journaled protocol below. | Native CLI has no expected-stash-OID predicate for reflog deletion. Do not silently substitute a positional CLI command. Unknown/reftable storage requires typed unsupported deletion. |
| Confirmed tracked restore | Bind full file/index/HEAD state, then capture the actual original by a same-filesystem rename before inspecting/replacing it; install without overwriting a newly created destination. | Filesystem capture is not a transaction against arbitrary open descriptors or directory replacement. Retain captured originals and report recovery state. |
| Exact detached deployment | Resolve/fetch and verify selected OID, pin departure/target history, use full OID and `--no-overwrite-ignore`, verify post-state. | Native checkout is not an atomic worktree snapshot against arbitrary external editors. |
| Attached branch advancement | Explicit fast-forward-only plan, expected-old-OID native ref transaction, native index/worktree transition and post-state verification. | A ref transaction is not a worktree occupancy lock or filesystem transaction. |
| Cancellation | Cancel before mutation; once effects may begin, retain ownership through reap/reconciliation and classify effects separately from cancellation intent. | Killing Git is not rollback; current observed state may not prove the complete historical effect. |

The files-backend stash writer is a deliberate, isolated exception to preferring
native Git commands. It is justified by the specific public-API gap below, not
permission for generic hand editing of Git databases. Everything else uses native
Git object/ref/index machinery or separately specified worktree-file recovery.
Its implementation must pass dedicated independent review before BC conformance
can be claimed. A native-only implementation must refuse exact drop/pop deletion
where it cannot establish the required interlock; it cannot claim the same safety
by adding another status check.

## Native API evidence and rejected shortcuts

Git 2.43.0's `get_stash_info_assert` requires a stash reflog reference for pop and
drop; apply accepts a stash-like commit directly. `do_drop_stash` invokes
`reflog_delete` with rewrite/updateref and prints the OID resolved earlier. Thus
its printed OID is not an atomic predicate on the record removed.
[Git v2.43.0 stash.c:694-772](https://github.com/git/git/blob/v2.43.0/builtin/stash.c#L694).

`reflog_delete` turns a numeric selector into a record count by traversing the log
**before** invoking the ref backend. The files backend then acquires the ref lock
and traverses the log again. Native locking protects its write, but cannot make an
earlier application OID lookup—or even that earlier record count—a stable
selection under an external stash mutation.
[Git v2.43.0 reflog.c:394-441](https://github.com/git/git/blob/v2.43.0/reflog.c#L394).

The backend explicitly uses the **reference lock** to exclude reflog writers.
`logs/refs/stash.lock` alone is not that interlock. Its expiration implementation
writes a new log, optionally writes the surviving top OID into the ref lock,
publishes the log, then commits the ref. This is two publications, with a real
crash window between them.
[Git v2.43.0 files-backend.c:3136-3270](https://github.com/git/git/blob/v2.43.0/refs/files-backend.c#L3136).

Native `update-ref --stdin` supports expected-old-OID checks and prepared
transactions: after `prepare` succeeds, referenced locks remain held until
commit/abort. EOF without explicit commit aborts a started transaction. This is
useful for holding a native ref lock while performing a bounded inspection, and
for final deletion of an empty stash ref including its packed representation.
It does not provide a reflog-entry predicate or guard file contents/occupancy.
[git-update-ref 2.43](https://git-scm.com/docs/git-update-ref/2.43.0).

Rejected as insufficient:

- Resolve OID to `stash@{n}`, then invoke drop/pop: external writers can shift it.
- Invoke drop/pop with the full OID: native Git rejects it as not a stash ref.
- Verify only `refs/stash` with old-OID CAS: middle entries can change without
  changing the top; even a top A→B→A sequence defeats an OID-only check.
- Hold a prepared native ref lock, then invoke native reflog delete: the nested
  command cannot acquire the same lock. The probe confirms refusal, not a
  lock-transfer API.
- Delete/rebuild the whole stash stack, use date selectors, or add logical
  tombstones: these change unrelated records/order, introduce ambiguity or leave
  a different native stash stack. They do not preserve the existing capability.
- Rely on native drop's stdout to recover an accidentally deleted stash: its
  displayed OID was resolved before the backend's locked deletion.
- Hold only an application mutex or custom lock: external Git does not honor it.

## Coordination, plans and effect vocabulary

The Git adapter owns a scheduler/guard keyed by **local common-repository ID**,
not remote URL. Serialize all gh-tree mutations in one common repository for the
first implementation, including its linked worktrees. A cooperating cross-process
application lock uses that same scope. This intentionally favors a simple safety
boundary over parallel mutations. Reads use immutable observations; a query
generation is not a repository version.

Native Git's own ref/index locks remain necessary. Never precreate an index/ref
lock and then expect an ordinary nested Git mutation to reuse it. Use a native
prepared transaction where available, or the narrowly specified writer that owns
the relevant lock. Release locks in a documented order; do not hold an operation
guard while waiting for UI confirmation. Confirmation consumption reacquires the
guard and revalidates all required facts.

Application owns `OperationID`, `ConfirmationID`, choices, single-use consumption
and terminal publication. Git supplies opaque, operation-specific plans rather
than a generic authorization boolean. Suggested boundary semantics:

```text
PrepareRestore(worktree, exact paths) -> RestorePlan + display facts
RestoreConfirmed(plan) -> MutationResult
PrepareCheckout(worktree, exact target, attached/detached intent) -> CheckoutPlan
CheckoutConfirmed(plan) -> MutationResult
ApplyStash(worktree plan, StashID) -> MutationResult
DropStash(StashID, expected occurrence observation) -> MutationResult
```

A plan binds repository/worktree identities, the operation kind and version,
exact expected target Revision, established/unborn HEAD representation,
symbolic HEAD branch identity, relevant refs/config/freshness, exact selected
paths and their file/index facts, and relevant worktree availability/occupancy.
Its display facts must be the same facts represented by the opaque plan.
Cross-repository use or unsupported plan versions refuse before mutation.

Recommended result separates facts that are often falsely collapsed:

```text
Effect = NotStarted | VerifiedNoTargetChange | AppliedVerified |
         Partial | Indeterminate
Result = {
  target effect, auxiliary/local-ref/index/worktree/remote effect facets,
  cancellationRequested, failure category, completed steps,
  authoritative post-observation or observation failure,
  recovery locations/refs and required reconciliation
}
```

`VerifiedNoTargetChange` is scoped: a failed fetch/prepare may have downloaded
objects or created recovery refs while leaving the chosen worktree unchanged.
Do not report global rollback from a HEAD comparison. `Partial` requires positive
evidence of completed steps; `Indeterminate` means the evidence cannot classify
all required effects. A known completed mutation plus failed refresh is not
reported as an unchanged failure.

External Git generally honors its native locks, but can start operations whose
read phases predate a lock. External editors, filesystem tools, hooks and manual
lock-file deletion need not honor either lock protocol. The application does not
provide repository-wide serializability against those actors. Detectable drift
causes refusal or partial/indeterminate recovery; never automatically overwrite
unexpected post-state to make a success assertion true.

## Stable stash identity and exact deletion protocol

`StashID = CommonRepositoryID + full OID`. Reflog selector, subject, timestamp and
managed-message text are observation/display data. The adapter verifies that the
object is a stash-like commit through native Git. Query patch/files by full OID,
not a cached position. A missing current log entry is `StashMissing` for a new
destructive operation even if a recovery ref still retains that object.

An OID may appear more than once after store/import. Do not choose an occurrence
by its index. Return `AmbiguousStashOccurrence` when the operation cannot select
one occurrence uniquely from its expected record observation; retain all entries
and refresh. The command must not silently delete all duplicates. Occurrence
metadata is an adapter precondition, not a second Domain identity.

Recommended exact-drop implementation for the baseline files backend:

1. Establish object/ref format with native Git. Support SHA-1 and SHA-256 files
   storage; reject symbolic `refs/stash`, malformed logs, unknown repository
   extensions/backends, redirected storage paths, or inability to enforce the
   tested local-filesystem lock/replace rules. No loose-file assumptions for
   reftable. Git 2.43 has the required primitives and is the tested baseline;
   new ref backends require explicit additional capability coverage.
2. Acquire the app common-repository guard, then exclusively create the actual
   files-backend `refs/stash.lock` at the Git-reported common directory. Refuse an
   existing lock; never unlink another process's lock. Validate administrative
   directory identity and no redirected child path before touching it.
3. **After acquiring that lock**, read the full current log and resolve its top.
   Validate the bounded format, full OIDs, non-symbolic ref and top consistency.
   Locate the exact selected OID/unique expected occurrence now. New pushes that
   happened before this lock are retained; pushes while it is held fail/wait at
   Git's interlock. Missing/ambiguous/mismatched input aborts before publication.
4. Create native recovery refs to the selected and surviving observed objects
   under `refs/gh-tree/recovery/<operation-nonce>/...` with create-only updates.
   Refs and a private journal preserve the original full log bytes, old resolved
   top/loose-ref state, intended new bytes/top, storage format, file identities,
   lock ownership and stage. Make originals/journal durable to the supported
   crash guarantee **before** altering the live log. Failure here aborts deletion;
   report any recovery refs already created as auxiliary effects.
5. Construct exactly the original record vector minus the selected occurrence.
   Preserve order, new OIDs and author/message bytes of survivors; rewrite only
   old-OID chaining as native `--rewrite` does (first old OID is the format's
   null protocol OID). Protocol null is not a Domain Revision. Do not normalize
   messages, invent records or deserialize/reserialize identities lossily.
6. Write/flush the complete new log into the owned log lock, and the new top
   (when nonempty) into the owned ref lock. Publish log then ref using tested
   same-directory replacement, matching native ordering. A new loose top safely
   shadows a packed old ref; do not edit `packed-refs` by hand. Record stages and
   verify exact surviving vector/ref while ownership is retained where possible.
7. If no record survives, keep the old ref temporarily and publish the empty log,
   as native expiration does. Release the manual ref lock, then use a native
   `update-ref --stdin` **delete with expected old OID; prepare** transaction.
   While its locks are held, inspect that the current log is still empty. Commit
   only then; otherwise abort and retain an external writer's new entries. This
   additional locked inspection protects A→B→A reinsertion that OID CAS alone
   misses. Native Git handles any packed-ref deletion. Failure to perform this
   cleanup is an explicit auxiliary cleanup result, not a reason to drop a new
   stack.
8. Retain recovery refs/journal until verified completion and the accepted
   recovery-retention policy permits removal. Recovery cleanup is scoped to the
   operation's owned refs and uses expected-OID native deletion; never expire
   unrelated reflogs or run GC as a cleanup shortcut.

This writer must remain private to Git and limited to `refs/stash` and its reflog.
Do not copy Git source wholesale or introduce a generic ref implementation. Its
format/lock assumptions must have direct tests against every supported Git storage
configuration. The probe below demonstrates the mechanism, not a production
crash-safe writer or support for every filesystem.

Crash/recovery rules are essential because the ref/log pair is not atomic:

- Before live publication: old live state remains; owned staged files may remain.
- New log / old ref: retain the ref lock plus journal; native writers are blocked.
  Recovery can finish the intended ref only if actual log/ref/lock identities and
  hashes match the journal's exact expected intermediate state.
- New log / new ref: classify completed deletion and finish journal bookkeeping.
- Missing/changed lock, mismatched bytes, unknown owner liveness, manual external
  repair or any unrecognized state: do not replay old snapshots over live data.
  Keep recovery objects/bytes, mark `RecoveryRequired`, refuse further destructive
  stash operations and expose the exact diagnostic/recovery path.
- No automatic stale-lock deletion by age/PID alone. A journal-owned lock needs
  verified ownership and a dead originating process identity; uncertainty refuses.
  Do not promise survival of disk corruption or unflushed power-loss state.

Apply uses the pinned OID with `--index`, preserving baseline staged restoration.
Require a clean target and no unresolved conflicts. Re-read index/worktree after
the command; a conflict keeps the stash. Pop drops only after verified apply,
using this fresh exact deletion step. If drop is missing/unsupported/blocked,
publish `AppliedStashRetained` (or the precise missing-entry fact), not an apply
failure inviting duplicate replay. App shutdown/cancellation never silently
starts the delete step after an unclassified apply.

Creation must also return the stash actually created by the operation. Do not
resolve `refs/stash` after `stash push` and assume that its current tip is ours.
Use an operation-unique creation marker in the supported stash message, inspect
the resulting log/commit facts and pin the uniquely matching OID. Preserve the
existing origin metadata as descriptive data. If no unique result can be proved,
return partial/indeterminate creation and reconcile; never substitute a newer
external stash. `stash create` can return a stable OID without cleanup for tracked
changes, but is not by itself a replacement for baseline `push -u` capability.
Native stash cleanup shares the external-file-writer limit stated below.

## Confirmed restore and post-confirmation edits

Baseline `RestorePaths` invokes `git restore --worktree`, whose source is the
index. Preserve that meaning: a staged version B with worktree edits C restores
B, not HEAD A, and leaves staging intact. Retain untracked and conflicted-path
refusal. Use literal, NUL-delimited path operands and preserve spaces; do not trim
valid names into different paths. A selected rename needs old/new identity facts;
gitlinks, sparse/skip-worktree/intent-to-add and special files require explicit
supported handling or typed refusal, not recursive guesswork.

The fingerprint includes exact path bytes and parent/object identity, presence,
type/mode/link target, full content digest, index entry stages/OIDs/modes/flags,
the full relevant index version, HEAD mode/ref/OID, and the restore source blob.
Status letters, size or mtime alone are insufficient: the probe changes bytes
while keeping status, length and timestamp identical. A double-read can detect
some unstable observations but cannot establish a universal snapshot.

Before prompting, retain the observed expected file bytes and index facts needed
for comparison. On confirmation, acquire application coordination and the native
index interlock for this worktree, and validate the plan. Read-only Git queries
must not refresh the index while the guard is held. Do not invoke ordinary
`git restore` against the original path after the final hash check: it can overwrite
an edit made in the remaining check-to-write interval.

Use a per-path recovery protocol instead:

1. Resolve the confirmed index blob/mode into a staged replacement using native
   Git conversion mechanics in an owned scratch location; preserve filters,
   attributes, CRLF and symlink semantics where supported. Bind relevant
   conversion inputs and verify them. Raw `cat-file blob` is not a compatible
   replacement for checkout conversion. Never invoke a configured converter and
   then claim it was a side-effect-free pure read.
2. Prepare a journal and a unique **same-filesystem** recovery location with no
   overwrite capability. A worktree-adjacent owned recovery directory outside the
   tracked tree avoids creating dirty backup files in the user's checkout. If a
   path is on a nested mount/different filesystem or safe capture cannot be
   established, refuse before changing it. Do not fall back to copy-then-delete.
3. Atomically rename the current directory entry into recovery. This captures
   the actual original at the mutation boundary; it does not unlink its content.
   Inspect the captured file itself against the confirmation. If bytes/identity
   differ, do not discard them: put it back only with a no-replace operation, or
   leave both existing destination and recovery intact with a recovery notice.
4. For a matching capture, install the prepared replacement at the original
   name using a no-replace primitive. Never truncate an existing destination.
   If an editor recreates that name, refuse replacement and retain both versions.
   For an expected-absent deleted tracked file, the same exclusive installation
   prevents overwriting a newly created user file. Preserve the index unchanged.
5. Verify source/index/HEAD and resulting path facts; return completed per-path
   effects. A multi-path restore is not all-or-nothing. Do not overwrite an
   unexpected path during rollback of earlier completed members.

Platform implementations must use handle/directory-relative operations and a
tested no-replace primitive, without following an unexpected symlink/reparse
target. The Python model below proves only ordinary NTFS rename/collision behavior;
it does not prove Go/Windows APIs, nested mounts or adversarial parent replacement.
Normal restore capability remains; a refused unusual storage state is explicit.

An editor holding the original open may continue writing to the captured file on
systems that permit that rename. **Keep the captured file inode/object**, not just
an earlier byte copy, until conscious recovery cleanup. A sharing denial on
Windows is a pre-publication refusal. There is no portable point at which the app
can prove all noncooperating open handles are finished; never automatically erase
that captured object merely because a post-read matched. The user-visible outcome
must name retained recovery when it exists. This preserves late bytes rather than
promising atomic CAS against arbitrary editors. Disk space/size limits refuse new
recovery-dependent mutations instead of silently evicting retained originals.

## Exact revision, worktree and history transitions

Application preserves the selected full `Revision` through every branch/PR/commit
command. Mutable source locators are resolved/fetched to an operation-owned ref,
then compared with that Revision before the target worktree changes. Use
`refs/gh-tree/operations/<unique nonce>/source`, not one shared PR-number ref.
Validate expected object format/type and scope. A moved source is a typed mismatch;
never silently use the cached local branch or the newly fetched different tip.

Pin selected and departing established HEADs to create-only recovery refs before
retargeting, including detached commits not reachable from a user branch. Do not
silently reset/delete the user's branch to protect a deployment convenience name.
Keep the existing refusal when required preservation cannot be established.
Primary/current-worktree restrictions remain operation-specific: generic retarget
refuses primary; configured legacy deploy also refuses its current worktree.
Availability, locked/prunable state and occupancy come from fresh Git inventory.

For detached deploy, `git switch --no-overwrite-ignore --detach <full OID>`
preserves the exact target independently of later locator movement. New attached
branches use create-only semantics at the same OID; duplicate names refuse. An
existing attached branch may be selected directly only after its exact current
tip matches the expected target. Disable branch guessing; do not use `-C`, `-B`,
`branch -f`, `reset --hard` or `--ignore-other-worktrees` as convenience fallbacks.

`--no-overwrite-ignore` matters: native checkout defaults can overwrite ignored
files that collide with the target. The probe demonstrates both behaviors. Scan
all untracked/ignored collisions and retain native guards; never call `clean` to
make checkout possible. Disable recursive submodule worktree mutation unless an
explicit supported submodule contract exists.
[git-switch 2.43](https://git-scm.com/docs/git-switch/2.43.0).

Preserve fast-forward attached retarget/pull capability without unguarded force:

1. Freeze old attached branch OID B and selected target T; prove B is an ancestor
   of T and revalidate clean target/occupancy. Divergence/rewind refuses.
2. For an existing different local branch, safely attach that branch at its
   observed old OID first, preserving departure history; verify that exact
   intermediate state. Any drift/failure is a reported partial transition. Never
   move an inactive branch underneath a worktree that may own it.
3. At the now attached target worktree run a native started transaction containing
   `update refs/heads/<branch> T B`, then `prepare`. Git's files backend acquires
   the branch and associated HEAD locks for its HEAD reflog update. Recheck
   symbolic HEAD identity, expected B, occupancy and supported lock behavior
   before changing the index/worktree. Abort if they do not match.
4. Run `git read-tree -m -u B T` with its normal index interlock and no ignore
   exclusion relaxation, then verify index/worktree state and commit the prepared
   ref transaction. Preserve an independent reconciliation lifetime if interrupted.
   `read-tree` applies Git's two-tree safety rules; it is not a file snapshot.
5. Verify exact HEAD T, expected attached identity, index and dirty facts. On
   failure after index/files changed, preserve B/T and the journal, abort any
   uncommitted ref transaction and return partial/indeterminate recovery. Never
   force a reset to B/T as automatic rollback.

The probe establishes this sequence on the files backend. Adding a separate
`verify HEAD` to the same transaction is **invalid** in Git 2.43: it collides with
Git's implicit HEAD log update. This is why the protocol revalidates HEAD under
the locks acquired by that single branch update. Ref backends/Git versions must
pass this exact lock-behavior contract or refuse that path; there is no generic
promise of it from an application mutex.

Native `update-ref` does not check worktree occupancy: the probe moves a ref
checked out elsewhere and leaves that worktree's files at the previous revision.
Do not infer safe checkout from successful ref CAS. Occupancy protection is exact
for cooperating app operations and checked through Git's normal checkout guards;
arbitrary external checkout can race the global inventory. Locks/verification
bound this risk but do not form a global worktree transaction.

External file writes can also race native switch/read-tree after Git's own
uptodate checks. For these native multi-file operations the supported concurrency
contract requires external writers to leave the affected worktree quiescent
during mutation. Preflight drift refuses; observed post-drift yields partial or
indeterminate recovery with retained history/original snapshots. A baseline
snapshot alone does not preserve an arbitrarily timed later edit. Do not describe
this as universal external-editor safety. If final REFDES instead requires that
stronger guarantee for checkout/stash cleanup as well, it must extend the
capture/no-replace file protocol to every overwritten path; native Git alone
cannot satisfy it. This is a contract limit, not a completed stronger proof.

Worktree creation uses exact OID, create-only path/branch semantics and no force.
Keep native worktree registration; after interruption inspect registration,
administrative directory, path, HEAD and files. Do not recursively remove a
partially created path whose ownership/content is uncertain. Unborn inventory
remains a no-Revision observation; operations needing an established departure
commit refuse without inventing an all-zero Domain OID.

## Fetch, pull, push and cancellation

Fetch into a bounded explicit scope, preferably `--atomic` for its local ref
updates; fetched object retention is still an auxiliary effect. A successful
fetch generation records the exact observed ref/OID scope, not permanent remote
truth. Failed/incomplete fetch never publishes exhaustive/fresh observations.

Implement pull as Application's fetch + exact-upstream resolution + the guarded
fast-forward transition above, avoiding a second implicit mutable-ref resolution
inside a broad `pull` command. Retain detached/no-upstream/gone/unknown/dirty and
non-fast-forward refusals. Upstream absence, deletion and inability to inspect are
different facts; unknown ahead/behind is not zero.

Push an explicit pinned source OID to an explicit remote branch, non-force and
without a `+` refspec. This prevents a moving local branch from replacing the
intended source. Server rejection and remote observation remain authoritative;
no force-with-lease shortcut is introduced. Setting upstream after a successful
OID push is a separate checked native-config step because an OID is not a local
branch name for `push -u`. Report pushed-but-upstream-configuration-failed if
necessary. Preserve a changed existing upstream instead of silently replacing it.

For reads, ordinary context cancellation stops observation and returns cancelled
without a mutation claim. For writes, distinguish pre-start cancellation from a
request to cancel an already started command. Once started, a navigation change
does not kill or forget it. Use a bounded adapter-owned execution/reconciliation
context separate from the cancelled UI/root wait. A hard timeout may terminate
the subprocess tree, but the adapter must reap it, inspect effects and retain its
journal/recovery before final classification. Exactly one Application terminal
outcome records both cancellation intent and actual effect; no second success
event later overwrites an earlier false rollback claim.

Classify each completed step: object acquisition, recovery-ref creation, stash
creation, file/index change, ref/HEAD transition, remote push and configuration
update. A fetch may be known partial; a checkout may have updated files before a
ref failure; a push may have succeeded remotely before transport loss. A later
remote tip equal to the requested OID proves current state, not which actor
caused it or absence of intervening effects. Do not retry uncertain destructive
mutations automatically. Reconciliation is a fresh correlated operation that
preserves the original terminal record.

Native hooks/filters/configured helpers may have effects beyond refs/index/files
the adapter observes. Preserve supported user behavior and report unclassified
effects honestly; do not claim a subprocess sandbox or silently disable hooks to
manufacture rollback guarantees.

## Executed probes and reproducible evidence

Environment: Windows 10.0.26200, Git `2.43.0.windows.1`, Python 3.14 executable
`C:/Python314/python.exe`. All fixture repositories/remotes are newly created
under `C:/Users/hanse/.codex/tmp/gh-tree-design-git-safety-52/`. Git global/system
config is isolated; fixtures have a local test identity and automatic GC disabled.
No real user files or external remotes were mutated. Only this appendix is written
in the design worktree; no product implementation, add/commit/push is performed.

Run: `python C:/Users/hanse/.codex/tmp/gh-tree-design-git-safety-52/probe.py`.
Final fixture directory: `run-_z0r6f22`; `results.json` stores all 15 results.

SHA-256 evidence: `probe.py`
`99BBBC22F3BEEEF65A7073DDF30D67535CCAF443066D170DEEC6CCE2A3514411`;
`results.json`
`D42D85E42C3A25CF0DB0A3F585AD1AF4D07B39F8A0F81984C54190D5B02B852D`.
Downloaded native v2.43.0 source hashes: `stash.c`
`BA905AE75D4689222D717578B280738CBA8F566DB1A0BF21EBDB118B938A7D8D`;
`reflog.c`
`0165CE9AAD2CEFDFF1FE2904FB375D5CF6B2206D32F71A392F4FE8693CA9E831`;
`refs/files-backend.c`
`A76C52BBF3A3EBD1CC4C46CA7AD1CA926F57C8733993498DB49C08A7F127653A`.

| Probe | Actual result |
|---|---|
| exact-oid-sha1 / exact-oid-sha256 | Push A then B; apply A by 40/64-digit OID restores A and retains [B,A]. OID pop/drop both exit 1. |
| precheck-gap | Observe A at zero, insert B, drop cached zero: B is removed and A remains. Demonstrates the application check-to-command gap. |
| native-ref-transaction-lock | Prepared verify on refs/stash blocks native stash store and drop; both exit 1 and stack is unchanged. |
| restore-gap | Edit confirmed bytes to different same-length bytes, reset mtime, retain identical status; native restore removes the later bytes. |
| capture-mismatch | Rename captures later bytes; digest mismatch refuses restore; no-replace link-back and retained recovery preserve those bytes. |
| capture-collision | After capture, external creation at original name causes exclusive installation to refuse; both original and new data survive. |
| ignored-collision | `switch --no-overwrite-ignore` exits 1 with ignored user bytes intact; ordinary detached switch overwrites them. |
| ref-cas-and-occupancy | Stale old-OID update exits 128. Native branch -f refuses occupied branch, but update-ref CAS moves it and leaves other worktree files mismatched. |
| locked-oid-delete-sha1 / locked-oid-delete-sha256 | Source-aligned temporary files-backend model locks first, blocks native external store, deletes exact middle OID, retains correct survivor order and native recovery refs. |
| log-ref-crash-window | Deliberately pause at new-log/old-ref/ref-lock state; external native store fails; completing ref publication restores the intended [A] state. This is injected intermediate-state inspection, not an actual killed-process/power-loss test. |
| empty-ref-cleanup-race | Construct A→B→A while old empty-ref cleanup was pending. Prepared native delete plus locked nonempty-log check aborts and preserves [A,B]. |
| exact-remote | Local bare remote advances after selection; fetched mismatch is detected before target change; subsequent verified full-OID detached switch reaches exactly the new selection. |
| prepared-fastforward | Prepared native branch update acquires branch and HEAD locks; two-tree read-tree plus commit produces exact new HEAD and clean index/worktree. |

Probe corrections are recorded rather than hidden: an initial cleanup-race
fixture assumed `stash store` of the current identical OID appends a record; Git
treats that as a no-op. The corrected fixture uses A→B→A and retains both entries.
Its expected survivor assertion was then corrected to [A,B]. A first fast-forward
probe added explicit HEAD verify and hit Git's duplicate-HEAD-update rejection;
the passing protocol uses the native branch update's associated HEAD lock.

The temporary model intentionally omits production codec/path/permissions/fsync,
journal ownership/restart, format gates and cancellation machinery. Passing it is
feasibility evidence only. No full product test suite, race run, Linux/macOS native
execution, reftable support, adversarial editor test, actual crash or power-loss
test was performed by this bounded reviewer.

## Required implementation and verification gates

The accepted Git BC must map at least GIT31-H01/H03/H04/H05/H06, DOM33-H01/H02/H04,
APP29-H03, GIT31-M03/M04 and related verification rows to these mechanisms.

- Real repositories in both object formats; linked worktree common scope,
  independent clone isolation, full/literal/unusual path bytes and unborn cases.
- Exact stash apply after arbitrary prior shifts; deletion under native external
  push/drop contention; missing/duplicate OID; top/middle/last deletion; packed
  stash ref; failed native empty-ref cleanup; unsupported/symbolic/malformed
  storage; full survivor byte/order verification, not merely selected absence.
- Kill/fault injection at every journal/log/ref stage; restart with owned and
  foreign locks, PID reuse/unknown liveness, changed hashes, disk-full/close/flush/
  replacement failures; no overwrite of unexpected recovery state. Verify retained
  objects through native Git after cleanup and GC eligibility separately.
- Restore staged B versus HEAD A/worktree C; edit after prompt, edit after last
  precheck, open-descriptor edit after capture, recreated destination, missing
  path, symlink/reparse/ancestor replacement, permissions, CRLF/filters, index
  modification, multi-path partial failure and cancellation. Confirm index and
  every later byte remain intact or explicitly recoverable.
- Exact branch/PR fetch mismatch; stale local branch; occupied/primary/current/
  locked/prunable worktree; ignored/untracked collisions; detached orphan history
  preservation; native prepared-lock behavior, fast-forward/divergence/ref-CAS
  failures and interrupted two-tree transition. Verify full post HEAD/mode/index.
- Fetch partial scope, ff pull, explicit non-force OID push, rejection/response
  loss/upstream-write failure, source movement and remote ABA; no automatic retry
  or unsupported unchanged claim.
- Windows/Linux/macOS native filesystem/ref-lock tests and all twelve required
  release cross-builds. Cross-compilation is not proof of native lock/rename/
  process cleanup. Tests must assert documented external-writer limits instead
  of claiming an impossible global filesystem/ref CAS.

Master must reconcile the explicitly bounded native checkout/editor concurrency
contract with the final safety wording, and freeze neither a stronger unsupported
promise nor a positional stash deletion shortcut. Exact next action: incorporate
these mechanisms, ownership and limitations into REFDES/API/verification; freeze
the complete design HEAD for independent review. No product change is authorized
by this appendix alone.

## Superseding addendum — protected publication for every worktree mutation

State: DRAFT, corrected after early independent safety review under #52.
Followup checkout HEAD: `58f1cb9eda941db0941cbb8e04e6a0559a3ca896`.

The review rejection is accepted. The earlier quiescent-writer assumption cannot
satisfy the user's no-silent-loss rule, even when external editors are quiet:
native stash creation invokes ref hooks before its later `clean`/`reset --hard`.
A hook can create new bytes in that interval. Git v2.43.0's own stash source marks
the reset collision as a known untracked-file problem.
[stash.c:1577-1658](https://github.com/git/git/blob/v2.43.0/builtin/stash.c#L1577).

The selected design therefore makes **native Git a preparation engine for file
changes**, and uses one protected publisher for checkout/fast-forward, stash
push/apply/pop and restore. No native switch, read-tree with `-u`, restore, stash
push/apply/pop, reset or clean may rewrite/remove actual worktree paths. Native
commands continue to own Git object encoding, merge/index semantics, and ref
transactions. This is a Git-private mechanism behind the same Application ports,
not an additional layer or a second product workflow authority.

### Preparation isolation and supported input

Use a private **standalone** scratch repository of the same object format, with
its own git directory, HEAD, refs/reflogs, index and object-write directory.
`GIT_ALTERNATE_OBJECT_DIRECTORIES` may supply read-only access to already pinned
real objects. A linked scratch worktree alone is insufficient: it shares the real
stash ref and other refs. Merely setting `GIT_WORK_TREE`/`GIT_INDEX_FILE` while
leaving `GIT_DIR` real also fails to isolate stash/ref effects.
[Git environment variables](https://git-scm.com/docs/git#Documentation/git.txt-GITINDEXFILE).

Copy bytes into scratch; **never hardlink real inputs into scratch**, because
native scratch reset/filters could then mutate a live inode through its alias.
Establish the exact source HEAD in private refs, including an equivalent private
branch name when attached source naming matters. Copy the confirmed index and
relevant file snapshots. Re-stat/revalidate inputs while preparing; a changing
read is not an atomic snapshot, but no real path has changed and final capture
will validate the actual originals before publication.

The first supported mutation profile is ordinary files-backend Git repositories
with understood normal full indexes, regular files and supported Git symlinks,
known object format, normal attributes/ignore/configuration and no active
merge/rebase/cherry-pick/sequencer. Do not strip unknown index features into a
simple index and call it compatible. Sparse/split indexes, skip-worktree,
assume-unchanged, intent-to-add, gitlinks/submodules, special files, unknown
mandatory extensions and platform link/reparse behavior need explicit proven
handling; otherwise return `UnsupportedMutationState` before any live publication.
Reads remain usable. A native scratch merge may produce ordinary unmerged index
stages as an **output**; that conflict result is supported below.

Native Git can normalize an operation-owned alternate index; an adapter must
never run that normalization on the real index. Preserve entry modes, object
IDs, stages and supported semantic flags. Refresh scratch stat information using
native Git before two-tree operations: blindly copying real stat data made the
probe's scratch `read-tree` refuse a clean file. Do not treat that failure as
permission to use force. The probe reconstructs ordinary stage entries with
native `read-tree --empty` plus `update-index -z --index-info`; that limited probe
is not authority to erase unhandled flags or extensions in production.

Production retains the complete understood native scratch index image. Before
publication, normalize/refresh its **alternate** image against installed live
files using native read-only worktree inspection (`update-index --really-refresh`
with an operation-owned absolute `GIT_INDEX_FILE`). This writes that alternate
index, never live file contents. Dirty-result exit status is meaningful, not an
invitation to overwrite files. Do not publish scratch stat/cache assumptions as
proof that live files are clean. Unsupported inspection/normalization refuses;
never write binary index fields by hand.

Snapshot the effective conversion/merge inputs: applicable tracked and untracked
`.gitattributes`/ignore files, info/global attributes and exclusions, object format,
filemode/symlink/case/EOL policy, and supported filter/merge settings. Bind their
versions to the plan. Disable scratch fsmonitor/untracked-cache shortcuts and
automatic maintenance. A command-specific adapter mapping chooses supported
effective settings; do not copy `core.worktree`, real git paths, hook paths or
repository extensions wholesale into scratch. External configuration drift
invalidates preparation rather than silently changing conversion rules.

Ordinary native EOL/attribute and byte content semantics remain required. A
configured external filter/merge driver whose execution depends on the real
physical cwd or cannot be reproduced safely in the private context is an explicit
unsupported preparation case until tested; do not silently drop it, run a live
checkout instead, or claim scratch isolation sandboxes arbitrary user scripts.
Known supported converters run against versioned copied inputs. Any helper's
out-of-scratch effects remain observable external changes which the publisher
must preserve. No later cleanup may discard them.

The mirror may be a bounded relevant snapshot rather than a copy of every ignored
build artifact. Its affected-path closure must include original/target index and
tree paths, all stash parent trees, selected included untracked paths, relevant
attributes/ignore sources, and parent/descendant shapes needed for directory/file
collisions. Missing mirror context must cause a safe preparation/publication
refusal, not erase a real collision. In particular, an entry absent from the
mirror is **not** authority to delete any real file at that name.

### One operation manifest and publication state machine

Preparation produces an immutable manifest containing:

- expected repository/worktree/HEAD/ref/index/configuration identities;
- exact source/target OIDs and native resulting index image/stages;
- explicit per-path old presence/type/identity/content and desired state;
- native preparation result, supported conflict classification and completeness;
- retained object roots, recovery directory identities and journal version;
- required ref transaction, hook invocations and single-use Application identity.

Only paths with an actual prepared semantic change are published. Unchanged
tracked files and unrelated ignored/untracked entries are never rewritten. New
untracked files appearing after the snapshot are not added to a cleanup set.
Renames are explicit old-path removal/new-path creation entries. Directory/file
transitions operate on named entries, never recursive deletion: remove an empty
directory only with atomic `rmdir` semantics, which refuses if new contents exist.
Refuse unsupported shapes/mounts before capture wherever possible.

Publication proceeds under the app common-repository guard:

1. Transfer and pin all new native objects required by the output. Scratch can
   send a native pack (`pack-objects --stdout --revs`) into real `index-pack --stdin`
   without touching worktree files. Include every required commit/tree/blob root,
   including merged index entries; a conflict index cannot be represented by
   `write-tree` alone. Bound and stream the transfer. Create-only recovery refs
   preserve source/target/departure history before any original is relocated.
2. Perform operation-specific prepublication ref effects, notably exact public
   stash store/new-branch creation. Their real hooks may change the worktree/index;
   the following guards/capture must detect or retain those changes. Prepare the
   final native ref transaction where required, with exact expected identities.
3. Acquire the actual per-worktree `index.lock` with exclusive creation. This is
   the native index **interlock**, not a claim of a public Git index-CAS API. Check
   the current entire index against the expected version under the lock. A native
   competing `git add` now refuses; a change committed before lock acquisition
   returns `StalePrecondition` without file publication. Do not invoke a native
   command expecting to reuse that lock; all image construction uses alternate
   indexes. Acquire additional guards nonblockingly; avoid lock-order deadlocks.
4. Persist the manifest, originals' retention locations and stage records before
   each irreversible directory-entry change. For every expected-present affected
   path, atomically rename the actual entry into a unique same-filesystem recovery
   location, then validate **that captured object** against the manifest. A newer
   edit or unexpected type is preserved and causes refusal/partial recovery.
   Expected-absent paths are not captured or deleted; they must remain available
   for an exclusive new-entry installation.
5. Prefer capturing/validating all required originals before installing outputs.
   On capture failure, restore an absent original name only via no-replace
   operation; never replace a new destination. Keep recovery aliases to captured
   objects. A partially captured set is a real effect, even if no new output was
   installed. Report it; do not label it unchanged merely because HEAD is unchanged.
6. Freeze each scratch output into a **separate retained payload object** on the
   destination filesystem, never an inode that native scratch Git will touch
   again. For supported regular files, retain that object in recovery and create
   the live name by a no-replace hardlink. This both prevents replacement of a
   recreated destination and keeps a name for the installed inode from its first
   visibility. A platform no-replace equivalent must prove the same retention
   property; unsupported filesystems refuse, never downgrade to copy-overwrite.
   Symlink publication is no-follow/exclusive with retained original link state.
7. For operations that change the index, capture the actual original index object
   under its lock and revalidate it, then publish the complete native-generated
   image with **no replacement of a newly created index name**. Retain both old
   and installed index objects in the journal. The temporary model hardlinks the
   completed lock image to the absent index path and to recovery, then releases
   the owned lock. No binary index storage is synthesized by this publisher.
   Worktree-only restore retains the exact original index instead of publishing
   an unnecessary replacement.
8. Release the index guard after file/index publication; commit the prepared
   ref transaction. A failed/unknown ref commit leaves a recoverable partial
   operation; do not run reset/switch to make files agree. Native hooks at the
   ref commit and applicable post-checkout/post-merge hooks run with their real
   operation context. **No cleanup or file/index replacement follows those
   callbacks.** Refresh/verify and report any hook-produced or concurrent changes.
9. Verify exact HEAD/mode/ref, index stages/semantic identity, occupancy and
   resulting files, returning per-facet effects and all recovery locations. A
   clean exact result is success. Known publication plus later edits/conflicts
   is a corresponding changed result, never a fabricated clean success or rollback.

The publisher does not promise all-or-nothing visibility across several paths,
index and refs. Readers may observe intermediate state; application commands are
serialized and a journal permits reconstruction. Native Git interlocks protect
cooperating Git writers. Arbitrary direct index editors are additionally protected
from silent byte loss by index capture/no-replace, but can still make the result
indeterminate. Directory-relative no-follow operations retain handles to validated
parents; unexpected scope/parent replacement aborts or reports captured recovery,
never follows a substituted symlink to overwrite a different tree. The platform
implementation must prove these operations, or refuse that filesystem/path state.

### Late handles, retry and recovery are part of the guarantee

Renaming an original does not mean every external handle is finished. An editor
may write into the captured object **after** its content was validated. Keep that
same file object reachable, not merely a byte copy. Windows sharing denial is a
safe refusal; the successful handle-sharing case is explicitly probed below.

Likewise, a user may edit a newly installed file before an interrupted operation
is retried. The retained payload link now contains those newer bytes. Its original
manifest digest remains immutable evidence; the payload must never be trusted as
unchanged just because gh-tree created it. **Every retry/recovery replacement must
capture the currently named object anew and install a fresh payload exclusively.**
Never truncate/delete an installed file on the basis of an ownership marker,
matching old hash or previous failed operation. Prefer a fresh plan/confirmation;
there is no automatic replay of a consumed plan.

No automatic retention expiry may unlink captured/published objects while an
unobservable external handle might still be writing them. Keep them until an
explicit conscious recovery-cleanup decision under the accepted product policy.
Space/size limits refuse additional publication rather than evict retained
originals. Completion must explicitly expose retained originals/recovery; an
editor's late bytes living there are not silently discarded. Unexpected recovery
state is preserved, and automatic rollback never overwrites live destinations.

### Operation-specific native preparation and terminal effects

| Operation | Scratch/native preparation | Real publication and outcome |
|---|---|---|
| Detached checkout | Exact target OID; native scratch switch with ignored/untracked protection. | Publish prepared delta/index, then native expected-old HEAD transaction. Retain departure history. |
| Attached fast-forward / pull | Native scratch two-tree read-tree from exact B to T; prove ancestry. | Publish files/index, then prepared native branch CAS B→T; no live `read-tree -u`. |
| New branch / attach existing | Prepare exact T in scratch; validate fresh occupancy. | Create new branch separately with native create-only OID update if needed; conditionally attach HEAD using the native symref protocol below. |
| Stash push with `-u` | Native scratch stash push computes W commit, index parent and untracked parent, and cleans only scratch. | Transfer exact OID; real `stash store <OID>` before capture. Publish only the captured snapshot's cleanup delta and index. Later user/hook files stay untouched or cause refusal. |
| Stash apply | Native scratch `stash apply --index <OID>` from confirmed clean target. | Publish exact prepared working/staged versions and untracked additions with exclusive entry creation. Existing real collisions are preserved. |
| Stash pop | Same apply preparation/publication. | Only after verified publication, execute the separately guarded exact-OID drop. Conflict/partial publication retains stash; drop failure is `AppliedStashRetained`. |
| Tracked restore | Native scratch restore from the confirmed index. | Publish only selected worktree paths; leave index bytes unchanged. Staged B/HEAD A/worktree C becomes B with B still staged. |

A scratch apply's ordinary merge conflict can produce supported conflict-marker
files and multi-stage index. Publish those with the same safeguards and return
`AppliedWithConflicts`, retaining the stash. A generic native preparation failure
with an unclassified partial scratch state does **not** authorize that state's
publication; return preparation failure with the real target unchanged.

If a real `reference-transaction` hook on stash store changes tracked or untracked
bytes, the subsequent capture either matches the original manifest or preserves
the newer object and refuses cleanup. A successfully stored stash can therefore
coexist with an unchanged/newer dirty worktree: report `StashCreatedCleanupRefused`
with the exact stash identity and recovery. Never apply/store again automatically.
The deterministic probe exercises precisely this case.

Scratch hooks are disabled, preventing premature/double lifecycle callbacks.
Real native ref commands retain their reference-transaction hooks. Where a
worktree operation is assembled from plumbing, invoke its applicable real
post-checkout/post-merge hook exactly once at the corresponding completed commit
point through native `git hook run`, with correct arguments and real context.
Do not invoke post-checkout for stash apply/restore. Hook failures and resulting
edits are explicit outcomes; no later cleanup attempts to undo them.
[git-hook 2.43](https://git-scm.com/docs/git-hook/2.43.0.html).

### Native symbolic HEAD transactions and branch occupancy

Use native symref transactions instead of introducing a manual HEAD writer.
Tagged Git **2.46.0** contains `symref-update`/`symref-verify`; this followup
executes the full attach/ref-log sequence using isolated official Git for Windows
**2.48.1.windows.1**. These are distinct source and execution evidence. Gate the
affected capability on the documented native command set and a supported semantic
probe/version profile, not a guessed unknown-option exit status. The supported
tested baseline for these attached transitions is 2.48.1; 2.46 is source evidence
of introduction, not a claimed tested minimum.
[Git v2.46 update-ref parser](https://github.com/git/git/blob/v2.46.0/builtin/update-ref.c#L278),
[Git v2.48 transaction tests](https://github.com/git/git/blob/v2.48.0/t/t1400-update-ref.sh#L1867).

Older Git receives `UnsupportedGitCapability` **for the affected attached/new-
branch transition**, with the prerequisite stated. Safe reading and supported
detached/other primitives remain available. No unguarded `symbolic-ref` or manual
HEAD fallback is permitted. This preserves the capability under an explicit tool
prerequisite; it does not silently remove the feature or modify installed Git.

Proven files-backend attach sequence:

1. Preserve departure/target objects. If creating a branch, use native create-only
   update to create it at T first. This is a separately reported preparatory
   effect; do not silently delete the branch on subsequent failure. Creating the
   branch and attaching HEAD in one transaction did not produce the expected
   HEAD reflog in this Git version because the target did not yet resolve.
2. Use `git update-ref --stdin --no-deref --create-reflog -m <operation reason>`:

   ```text
   start
   symref-update HEAD refs/heads/new ref refs/heads/old
   verify refs/heads/new <T>
   prepare
   ```

   For a previously detached HEAD, use the supported `oid <B>` old condition.
   After prepare, inspect actual HEAD mode/name/OID under its native lock.
   A same-OID switch to another old branch is a stale identity, not equivalence.
3. The old attached branch's OID also needs protection. Git rejects adding a
   `verify refs/heads/old B` to that same transaction: its implicit HEAD log update
   conflicts with explicit HEAD update. In the tested **files-backend** profile,
   exclusively acquire that old branch's `.lock` as a **verify-only guard** after
   native prepare, then resolve/compare B while it is held. It never writes or
   commits ref bytes. A competing native update is blocked. Failure aborts before
   file capture. This reuses the source-established files ref interlock, not a
   manual HEAD/ref writer. Unknown backend/path semantics refuse this profile.
4. With native HEAD/target locks and the old-branch guard held, execute protected
   file/index publication, release its index guard, commit the native transaction,
   and release only the owned old-branch guard. Verify exact symbolic HEAD/T and
   native HEAD reflog. The probe retains both new and departure entries.

For fast-forward on the already attached branch, the single native B→T branch
transaction acquires the associated HEAD lock as in the earlier probe; no explicit
duplicate HEAD update is added. Its filesystem work now occurs only in scratch.
Detached HEAD updates use no-deref expected-OID native transactions plus the
required departure-branch guard when starting attached. A guard that causes a
mutating hook to encounter contention is surfaced; never silently skip the hook
or rerun a destructive phase to obtain a clean result.

Ref locks are not occupancy locks. Revalidate Git's worktree inventory before
publication and after commit. Existing occupied branches refuse; no force flags
or writes to another worktree are allowed. An external checkout can race the
inventory, so no universal one-worktree-per-branch transaction is claimed. If
occupancy changes, return the known partial/changed state with preserved originals
and refs rather than success or automatic foreign-worktree cleanup. Unlike the
rejected live-checkout design, that race cannot make this publisher silently
overwrite another actor's file bytes: it touches only captured target entries,
leaves other worktrees untouched, and never rewinds an existing branch.

### Followup evidence and remaining verification boundary

Only new temporary repositories/files and this appendix changed. System Git stays
2.43.0.windows.1. Official MinGit 2.48.1 is extracted under
`C:/Users/hanse/.codex/tmp/gh-tree-design-git-safety-52/MinGit-2.48.1/`; it was not
installed or added to user PATH/configuration.

Run `python C:/Users/hanse/.codex/tmp/gh-tree-design-git-safety-52/probe_scratch.py`.
Final run: `scratch-bx4nok3u`; `scratch-results.json` records **18 PASS** results:

Followup SHA-256: `probe_scratch.py`
`E326EE32342DDE76E083DAEE210792D5E55AF3F4050E5D8FBD5CD947DAD66BE3`;
`scratch-results.json`
`586C7CF6E5C077D16C070E5E67E06CC4066E2440540FE3CD0866B0ACADD2980C`;
official `MinGit-2.48.1-64-bit.zip`
`11E8F462726827ACCCC7ECDAD541F2544CBE5506D70FEF4FA1FFAC7C16288709`.

- SHA-1/SHA-256 scratch stash push preserves native staged B/worktree C/untracked
  U parent semantics; real target remains unchanged during preparation; native
  pack transfer/exact store/protected publication succeeds and retains ignored I.
- Both formats' exact scratch apply with `--index` restores B/C/U and retains stash.
- Real stash reference-transaction hook writes late tracked and new untracked
  files; cleanup refuses at captured mismatch with all those bytes preserved.
- Scratch detached switch and prepared fast-forward publish exact clean HEAD/index
  without a native command writing live paths.
- Native git add during index guard exits 128; earlier external index change is
  refused before capture; restore leaves exact original index bytes unchanged.
- Last-check edit, recreated destination and installed-file edit followed by stale
  retry all preserve actual current objects. Installed payload links retain edits.
- Native `stash apply --index` conflict yields safely published conflict markers
  and unmerged index, with stash retained.
- Ignored-file collision refuses in scratch without changing real ignored bytes.
- Git 2.48.1 native symref sequence yields exact new branch/clean index and retained
  native HEAD reflog; old-branch guard blocks updates. Same-OID old-branch identity
  change is rejected by native symref CAS (exit 128).
- A Windows `CreateFileW` handle allowing delete sharing writes to the original
  **after capture and installation**: live output remains A, while the captured
  original inode contains the later handle-written bytes.

Disclosed probe corrections: Python's first Windows lock writer omitted binary
mode, inserting CRLF bytes into a binary index; the corrected writer uses
`O_BINARY`. Copying real index stat data without native scratch refresh caused a
two-tree uptodate refusal; refresh corrected the fixture. Native symref attempts
first omitted no-deref, then combined old-branch verify with explicit HEAD update;
both correctly failed. Creating the target branch and symref in one transaction
did not append the expected HEAD reflog; the final sequence creates it first and
uses explicit native reflog creation. These failures explain the final protocol;
they are not concealed passing evidence.

This ordinary-file Python model omits production path/codec/journal/permissions,
resource bounds and crash recovery. It proves the mechanisms in normal NTFS
fixtures, not every supported platform/filesystem. Future native Windows/Linux/
macOS conformance must exercise capture/payload/index/ref crash boundaries,
same-filesystem no-replace/link behavior, directory/symlink/reparse replacement,
space/permission faults, semantic index flags/config/attributes/filters, hooks,
late handles, concurrent index/ref writers, ambiguity and cancellation. Include
all twelve release builds; compilation alone is not native safety evidence.

**Final design rule:** all supported file-changing Git operations use this
protected publisher. Unsupported native/storage semantics refuse explicitly.
Neither external-editor quiescence nor pre/post hashes authorizes an in-place
overwrite. Earlier live switch/read-tree/stash-cleanup recommendations are
superseded. Master can now incorporate this concrete protocol and capability
prerequisites into REFDES/API/Slices/verification and request independent review.
