# M3 Persistence contribution — Issue #63

State: IN PROGRESS; codec/version, native mechanisms/security and bound typed-load
milestones; public commit/recovery and adapter acceptance remain pending.
Branch: `codereview-21/layer-persistence`.
Initial base: `412f33e477cec03cb6eafe7b846c9bcdd02c0a25`.
Worker owns only `internal/persistence/` and this report. Master supplies separate
independent review and serial integration after Git/GitHub. No legacy/shared/
frozen files, product entry point, global configuration or user stores changed.

Authority: full #21/#63 resumed assignment, UserRunContract resumed policy,
M3-Adapters/M3-Assignments, accepted Persistence LR and preceding Domain/Discovery
conclusions, REFDES/API/Storage/Slices/Verification/MigrationMap/FindingDisposition,
FROZEN Application--Persistence/Discovery, shared BoundaryTypes and State--App
contracts plus BCFreeze. Actual accepted Domain/API/ports are the type authority.

## Milestone P1 — strict documents and versions

Source: `a95d112aa493d3899d34728843b59090ab3c6f17`, pushed and clean at P1 handoff.
Files: `codec.go`,
`codec_documents.go`, `document.go`, corresponding tests and layer README.

| Clause / related checks | Actual implemented evidence |
|---|---|
| V-PER-01/03 codec/migration inputs | All three full schema0/schema1 codecs and constructors; exact strings/ordered targets, only prefixes-null, copied raw originals; all known object levels and noncurrent scope entries retained. Round-trip fixtures plus wrong-shape/null/duplicate/UTF-8/surrogate/depth/size/forward-integer tests. No disk migration or business intent choice. |
| V-PER-01 unknown preservation | Guarded-document retention primitive checks every retained object path. Independent member-removal controls reject loss, including enclosing-object deletion. Comparison preserves exact decimal values without float rounding or unbounded exponentiation. |
| V-PER-02/04 content/scope foundation | Complete-byte length/SHA256, presence, family, store, run WorktreeID/root binding; deterministic missing-anchor/literal-component tokens. Tests distinguish foreign family/root/worktree/parent/basename, raw/unknown changes and same-byte restoration. Native binding/absence proof is still pending. |

Executed locally: Windows amd64, Go1.25.0; targeted package tests and vet PASS,
gofmt and `git diff --check` PASS. Bounded pre-final numeric-normalization fuzz
run: 549186 executions in11.547s, PASS (23 new interesting inputs, no failing
corpus). Subsequent exact-decimal normalization change passed the updated package
tests/vet, including large positive/negative exponent carry/borrow controls.
Tests mutate only in-memory values; they perform no store or configuration I/O.
These are author tests, not independent review or native product verification.

## Milestone P2a — Windows acquired objects and permanent locks

Source: `19df0a110cfb91aa6ff7d6cda5b1dd20dc5ae0bb`, pushed/clean at P2a handoff.
`acquire_windows.go` adds request-owned NtCreateFile ancestor guards
with actual list/data-read/no-delete access, handle-relative no-reparse opens,
aligned full FileIdInfo/native birth identity, local-NTFS detection, coherent
bounded reads, missing-anchor observation without writes, and cancellable
permanent LockFileEx byte0/length1 locking. Native NTStatus plus errno are retained.

`acquire_windows_test.go` executes native acquisition/read/absence, parent movement
refusal, data-read child deletion exclusion, empty-parent junction conversion
and safe retained-parent relative refusal, multi-handle/two-store locking,
cancellation and kernel lock release after a real child process is killed.
Initial controls incorrectly assumed an empty parent could not convert and let
the child exit via Go deadlock detection; corrected tests follow the accepted
windows-anchor/Followup.md and keep the child alive until its owned parent kills/
joins it. No product contract was weakened or changed.

Exact candidate local tests: Windows amd64 full Persistence package PASS, vet
PASS; native Windows386/WOW64 native selectors PASS (0.654s); Windows ARM64
package build PASS, explicitly compile-only. Native FileIdInfo assertions:24
bytes, FileID offset8 on tested amd64/386. Formatting/diff checks PASS. All
filesystem/process fixtures are beneath test-owned temporary directories.
No API Storage method/public constructor/publication/recovery profile is claimed
implemented by these private primitives. P2a is a recoverable source checkpoint.

## Milestone P2b — Unix descriptor acquisition and flock

Source: `4c5e5bed6b38860dc7a823780475e264605cbf26`, pushed/clean at P2b handoff.
`acquire_unix.go`, platform profile helpers and native tests implement no-follow
Openat/Fstat/statx/BSD birth observations, no inherited descriptors, bounded
double-read consistency, missing-anchor observation, explicit moved-object versus
substituted-path revalidation, special-object refusal before blocking reads,
reference-counted inode mutex plus stable flock, and request-owned cleanup.

Native Linux execution PASS: Go1.25.0 crosscompiled CGO0 test binary, then actual
WSL openSUSE-Leap-15.5 / Linux6.18.33.2-microsoft-standard-WSL2 x86_64,
UID/GID65534 nobody, ext4 fixtures entirely beneath task-owned Linux `/tmp`.
Complete package cases and fuzz seeds pass, including real child-process kill/
flock release and directory rename followed by substituted symlink refusal.
All nine Unix target test binaries compile; Windows P2a source is unchanged.
Filesystem recognition is acquisition only, not native metadata/durability proof.

Bounded local evidence: Windows temporary directory
`C:/Users/hanse/AppData/Local/Temp/gh-tree-persistence-p2b-aa36504819f0492097fff66545082146`,
Linux staging `/tmp/gh-tree-persistence-p2b.oF0eSB`. `persistence-linux.test` SHA256
`b3b839e7dc56cd532f2e43650ea1e9cbf22867e7dc7288415f473466cd8d720c`;
`native-linux.log` SHA256
`19f9d6dc38c0254b43d2514ab49b274384a42fdd0a50017df464dc171ff4ead9`.
Three harness-only attempts initially failed from WSL backslash/command-path/
PowerShell dotted-argument transport, before tests ran. Final reproducible route:
WSL `--exec /usr/sbin/runuser -u nobody -- /usr/bin/env TMPDIR=<owned>/tmp
<owned>/persistence-linux.test -test.v=true -test.timeout=30s`, with literal quoted
PowerShell test flags. No Linux toolchain installation/global change was made.

## Remaining native and acceptance work

P2f now supplies constructor/run binding and three typed read paths with native
reacquisition and observed missing-parent adoption. Private lock/security/native
publication helpers are tested foundations. Public commit methods and load-time
recovery observation remain pending. Preserve the originally
supplied Expected absence anchor in the native manifest through parent creation;
shared recovery Original can be the guarded established-parent absence.

P3: exclusive payload/manifest/raw backup and exact retained original, bounded
admission, selected publication/barriers, complete outcomes and stable persisted
recovery IDs/restart observation. No replay/eviction/fallback. Fault/crash/race/
late-writer/resource controls and all required native profiles remain mandatory.
These are forthcoming implementation milestones, not Unsupported placeholders.

Master supplied the exact common DirectoryIdentity convention; Git's separate
native source checkpoint is `dd80cdecc0c835d119dc21671fbba2efacc23644`, read-only
convention evidence with no import/call. Native FreeBSD metadata/mechanism job
remains Master-coordinated. Existing
x/sys0.44.0 is sufficient to begin native helpers, and current shared CI provides
Linux, macOS ARM64, Windows amd64/ARM64 and Linux race. No workflow/module edit
was made here. Native evidence is distinct from architecture cross-builds.

## Milestone P2c — selected native publication and retention primitives

Exact source is the commit adding this subsection, reported after push. This is
a private mechanism checkpoint; no public Storage commit path is claimed.
Windows class65 rename/class11 no-replace hardlink use retained handles and
pointer-aligned native layouts, with explicit386/amd64 offsets. Unix selects
Renameat for presence and Linkat for absence. A successful absence Linkat leaves
the owned payload name so subsequent cleanup cannot erase the known commit point.
No replacement fallback, replay, truncate or retention deletion is implemented.

Native author tests: Windows amd64 complete Persistence package and vet PASS;
native386/WOW64 selected publication tests PASS; WSL Linux/ext4 UID/GID65534
selected publication tests PASS. Cases cover existing-absence competitor refusal,
no-replace original retention, late writes through an old held handle versus the
independent raw backup, and three fresh native readers through30 replacements.
The initial Windows test used an ordinary Go reader that did not share DELETE;
its actual sharing failure was corrected to the required native reader profile,
without changing the publication primitive. Linux file/parent fsync executed.
Windows namespace power-loss durability and full request outcomes remain unproved.
Bounded test staging: Windows temporary `gh-tree-persistence-p2c-27053b08c0d844bc9c1fb47731ec1c76`;
Linux `/tmp/gh-tree-persistence-p2c.iVD5TK`, all fixtures owned under its tmp folder.

P2b CI correction: actual4c5e5be CI34054966709 passed18 SUCCESS jobs and the sole
pre-Runtime helper skip, executing this package on Windows amd64/ARM64,
macOS26 ARM64, Linux and Linux/race. That source CI proves P1/P2a/P2b, not these
new publication primitives or pending metadata/recovery. FreeBSD remains pending.

Master confirmed the already accepted bounded Windows security profile excludes
audit-only SACL replication; READ_CONTROL queries must never claim audit absence.
Owner/group/ordered DACL/protection/inheritance and access-affecting label/resource/
CAP queries/copy/refusal remain mandatory before public publication. No new public
option, privilege escalation or contract change is introduced by this clarification.

## Milestone P2d — bounded native Windows security profile

Exact source is the commit adding this subsection, reported after push. Private
security helpers query OWNER/GROUP/DACL plus LABEL/ATTRIBUTE/SCOPE and native
trust/filter selectors0x80/0x100. Only label ACEs are copy-supported; other
access-affecting SACL profiles refuse. Comparison preserves complete ordered ACE
bytes and relevant protection/inheritance bits. No SDDL normalization substitutes
for policy comparison. Read-only/special flags, alternate streams and nonzero NT
EA size refuse. Exclusive NtCreateFile supports a supplied protected current-user
DACL at creation, before any sensitive bytes. No public port is wired yet.

Native Windows amd64 package and vet PASS; native386/WOW64 targeted metadata
tests PASS. Real protected/unprotected DACL, ordered deny/allow ACE and explicit
low-label copies pass re-query verification. Real resource-attribute-bearing file,
read-only flag and alternate stream controls refuse unchanged. User-only creation
checks the sole ACE's actual SID and protected mask. Empty label input originally
created a different empty SACL when unnecessarily set; fixed by avoiding that
write and still verifying exact source/destination policy. Genuine sharing errors
in ordinary ADS fixture creation were fixed by the specified native share-all
test profile. No metadata/publication fallback was added.

The existing selected profile explicitly excludes audit-only SACL replication;
limited queries never establish audit absence. No new public API option or
privilege escalation. Primary query-rights authority:
[Microsoft SECURITY_INFORMATION](https://learn.microsoft.com/en-us/windows/win32/secauthz/security-information).
CAP/trust/filter nonempty native fixture execution and all-platform full request
fault/concurrency/recovery evidence remain future acceptance gates; querying the
empty native fields here is not proof of every possible security-policy profile.

Local residue: an earlier deny-ACL test left only its owned
`C:/Users/hanse/AppData/Local/Temp/TestWindowsNativeMetadataCopy2880129540/001/copy-d`
and `source-d`. Attempted cleanup first checked the absolute fixture root beneath
the current TEMP, then attempted Set-Acl to `D:P(A;;FA;;;OW)` on those two literal
children and Remove-Item -LiteralPath on the exact test root with Recurse/Force.
Automatic approval review rejected that command as `blocked by policy` before
execution. It was not retried or bypassed. Root/user were informed; the old local
fixture remains intact. Later tests restore DACLs through already-owned native
handles before close, and their cleanup passes. No user store/global ACL changed.

## Milestone P2e — native Unix metadata inspection and copy

Exact source is the commit adding this subsection, reported after push. Private
helpers bound attribute lists/values to1MiB, read them twice around native stat,
and apply Fchown before Fchmod and native xattrs, followed by complete verification.
Additional inherited attributes refuse instead of being silently removed. Linux
inspects inode flags and copies user attributes and system.posix_acl_access;
other returned security namespaces or unsupported flags refuse. Extents allocation
is supported. Native unprivileged Linux/ext4 UID/GID65534 full Persistence package
and Linux-target vet PASS, including real named-user POSIX ACL, mode/uid/gid/xattr
copies, changed attributes, inherited extra metadata and invalid-descriptor errors.

Darwin fgetattrlist explicitly asks for returned-attribute bitmap plus extended
security, validates bounded native attrreference/filesec representation and accepts
only actual absent ACL or KAUTH_FILESEC_NOACL, never a valid empty deny-all ACL.
FreeBSD direct ExtattrListFd preserves both namespace errors. This avoids pinned
x/sys xattr_bsd.go Flistxattr's intentional system EPERM suppression and the
FlistxattrNS error path's nil error. Native __acl_get_fd inspects current POSIX and
NFS4 types; nontrivial POSIX and NFS4 ACLs refuse. Native header:24+254*16=4088
bytes across target ABIs. Unsupported EINVAL is propagated, not treated absent.

Darwin ARM64 and all three FreeBSD test binaries cross-build; native metadata
execution is pending and no profile acceptance is claimed. The Darwin test adds
a real ACL through /bin/chmod only in its owned fixture. FreeBSD ordinary-user
system attribute enumeration may deny; Root has this concrete concern and #66
must record the actual native result. No error is discarded to make that profile
pass. Source references: [Apple native ACL layout](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/kauth.h)
and [FreeBSD native ACL layout](https://github.com/freebsd/freebsd-src/blob/releng/14.3/sys/sys/acl.h).

P2c exact0630ded CI34055865289 passed18 SUCCESS plus the pre-Runtime helper skip,
independently confirmed by Master. It proves those mechanisms on current native
CI platforms, not P2d/P2e metadata or the forthcoming complete request protocol.

## Milestone P2f — explicit binding and typed read paths

Exact source is the commit adding this subsection, reported after push. `New`
accepts Composition-selected absolute locations and bounded construction settings,
resolves explicitly selected links/junctions once and binds actual parent/absence
anchor identities. No ambient default location, directory creation or migration.
Three typed load methods acquire/revalidate native objects for each request and
return complete documents or explicit absence/corruption/unsupported/unavailable
observations. Recovery manifest observation is not yet wired into these methods.
No Commit method stub or port-conformance assertion claims completed Storage.

Binding checks reject user/preference same location or hardlink identity,
document/other-family permanent-lock collision and run/user overlap. Missing
components may be observed/adopted; a small in-process protected identity record
then rejects later replacement of the adopted parent. The original construction
anchor/literal remainder remains separately retained for forthcoming expected-
absence transition evidence. Run scope rechecks Git's exact root/profile before
the sole literal .gh-tree child. Unix honors a supplied change stamp without
silently substituting its available birth stamp; observed change then refuses.

Native author verification: Windows amd64 complete package/vet PASS; Windows386
typed binding/link tests PASS; WSL Linux/ext4 ordinary UID/GID65534 package PASS
before the final additional cases, then all typed-load/link/change-profile cases
PASS and Linux-target vet PASS. Tests prove no-write absence, full families/raw
versions, legacy/current/corrupt/forward/oversized inputs, exact file/lock family
overlap, replaced parent/adoption, cancellation/concurrency, explicit external
link selection versus project-child refusal, and no request-held parent guards
after return. All filesystem changes remain in test-owned temporary directories.

Windows filepath.EvalSymlinks does not resolve the fixture's directory junction;
explicit selection now uses one CreateFile/GetFinalPathNameByHandle resolution,
followed by the required no-reparse native binding/reacquisition. This is confined
to constructor-owned explicit user scope, never run scope or publication fallback.
One existing child lock fixture produced an EOF without native cause; targeted
rerun passed. Its uncontended child initialization now uses a1s budget within the
frozen5s cap and captures actual stderr/wait failure. Separate contention/bounded-
cancel tests retain their short deadlines. The updated complete Windows run passes.

Native CI history: P2d c146bb5 CI34056256194 SUCCESS. P2e cb1f625
CI34056596058 failed only macOS TestUnixNativeMetadataCopy: bitmap-based extended
security detection refused an ordinary no-ACL APFS file. This failure is retained.
Apple XNU attr_pack_common omits the returned bit for a supported NULL ACL as
well as unavailable metadata; vfs_attr_pack_internal with RETURNED_ATTRS omitted
instead returns EINVAL for unsupported requested metadata. The corrected single
selector distinguishes native zero-length ACL absence from unsupported requests,
with bounded12-byte header/reference and complete filesec/NOACL validation.
[Exact source functions](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/vfs/vfs_attrlist.c)
establish this correction; corrected Darwin ARM64 cross-build passes but its
new native CI result remains pending at this checkpoint. No bitmap/error/xattr
absence is silently accepted. Native extended-ACL refusal test remains required.

Root's #66 infrastructure run34056731199 independently observed ordinary FreeBSD
uid1001 system-EA enumeration EPERM on its actual filesystem, while user namespace
enumeration succeeded. This is not adapter proof; the direct FreeBSD helper
preserves that error. Ordinary-user supported-profile resolution remains with
Master and future native adapter evidence, without weakening guards.

## Milestone P3a — request preparation foundations

Fresh request-protocol worker, exact source is the commit adding this subsection.
Private preparation helpers now perform full-write length/cancel checks, native
exclusive file/directory creation with initial per-user or inherited project
access, native parent-policy inspection and retained-parent directory enumeration.
Recovery namespace is fixed-size and agrees with the current native basename
equivalence. Actual file sizes, incomplete/orphan matching artifacts and operation
nonces consume admission; excessive enumeration refuses instead of allocating an
unbounded list. Enumeration never deletes, replays or trusts manifest size claims.

Windows amd64 full Persistence package/vet and diff checks PASS. Native Linux/ext4
UID/GID65534 targeted preparation tests PASS through existing WSL, Go1.25.0 CGO0
test binary, fixtures solely beneath owned `/tmp/gh-tree-persistence-p3a.doq7dO/tmp`.
Controls cover short nil-error writes, cancellation before I/O, multi-chunk bytes,
exclusive creation refusal, actual metadata inspection, actual byte/record limits
including incomplete crash residue, and no inventory changes. This checkpoint
does not publish a document or assemble the public commit/recovery protocol.

The fallback directory change-stamp needs explicit retained-object own-effect
transitions: lock/artifact creation changes ctime, while shared API observations
require common document/store/root binding. Original Expected must remain intact;
no binding rewriting or dropped incarnation checks is authorized. Master has the
concrete issue; stable logical binding and live acquisition evidence are distinct.
The ordinary FreeBSD system-EA EPERM profile remains unresolved without a waiver.

## Exact successor work

P3b private manifest checkpoint (the commit adding this paragraph) now supplies
bounded strict manifest decoding, family/store/parent/root/Expected/Original/
Proposed and generated-name structural checks, per-artifact cryptographic IDs,
native artifact identity checks, and shared typed recovery construction. Native
Windows amd64 full package/vet PASS; Linux/ext4 UID65534 targeted manifest tests
PASS in owned `/tmp/gh-tree-persistence-p3b.Fs2rkC/tmp`. Fresh acquire/lock/close
cycles preserve all IDs; late writes to the retained original leave raw backup
unchanged; replacing a raw backup with identical bytes refuses its old ID, while
independent unaffected records survive. Malformed/foreign bindings refuse.
These private helpers are not yet connected to loads or public commits and do
not establish the complete restart/publication protocol or profile acceptance.

Master verified P2f87506ee CI34057603393 and P3a8775a80 CI34058262124 SUCCESS;
the corrected Darwin selector therefore has actual native CI evidence. Accepted
shared FreeBSD CI prerequisite canonicald5eeaac CI34058149301 is awaiting its
Master-owned merge at this clean P3b boundary. New native adapter execution must
preserve the ordinary system-EA EPERM error. No FreeBSD profile waiver is made.

Manifest identity is a recorded observation, distinct from the stable RecoveryID;
the frozen contract does not demand recursively embedding a manifest's final
ctime. Native no-birth association still needs explicit supported-profile proof;
changed inode observation alone cannot prove replacement absence. Existing
birth-stamped native controls pass. Full own-effect directory/root transitions,
native ancestor verification of missing-parent Expected evidence, persistent
payload publication/recovery naming, and constructor namespace exclusions remain
part of the next complete-protocol work, not accepted by this partial checkpoint.

Stop clean after this pushed milestone for the Master-requested fresh request-
protocol worker. Existing private helpers are not an assembled safety protocol.
Implement all three public commits and wire recovery into loads: expected whole-
version validation under permanent lock; safe missing-parent creation/adoption
with original Expected anchor preserved; supported parent/new-project/per-user
metadata and exclusive payload/manifest/raw backup; exact original hardlink;
bounded256/1GiB actual admission; manifest/artifact identities and per-artifact
persisted RecoveryIDs; all required write/flush/close/prepublication comparisons;
selected native commit and barriers; truthful known/indeterminate/cancel/error
results and independent current observation. Native helpers currently expect
their caller to establish those barriers. Never call them merely because encoding
succeeds. Extend construction namespace overlap for the final recovery-name scheme.

Restart observation must reacquire the stable lock and safely report exact retained
facts, preserving same IDs and family/worktree/store/document-version association;
no replay/eviction/age cleanup or attribution from equal bytes. Explicitly prove
missing-parent Original/Proposed rebinding versus the original request Expected,
including own-effect updates for change-stamp profiles. Full native fault/crash/
late-writer/cooperative multi-process/external-gap/outcome-delivery tests remain.
Obtain corrected macOS and supported FreeBSD metadata evidence, then separate
fresh review of the complete candidate. No integration or Slice closure here.

SLICE(S): SLC-01/04/05/09/10/12/13 foundations only. REVIEW: pending fresh reviewer.
INTEGRATION: none. TAG: none. All full Slices and baseline findings remain open.
NEXT: push P2f and hand the clean exact source to the fresh request-protocol worker;
native CI remains pending. Complete public commits/recovery before final review.
