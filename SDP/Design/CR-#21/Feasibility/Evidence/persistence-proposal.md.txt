# Persistence commit feasibility proposal (PERF-52)

Design-only contribution for #52/#21; inspected frozen design
f1dec60cbd3892f068410f2d7f3caa7f855ba52e. This is a proposal to reconcile into
REFDES/API/Verification, not an accepted BC or product implementation. Scope is
one document; schema details and semantic migration mappings remain in the
Master's separate codec decision. No product file, Issue, ref or commit changed.

## 1. Scope is a bound directory object

Project storage receives Git/Application's selected WorktreeID, physical root
locator and expected root identity. Acquire the root and compare its actual file
identity with that authority. Open the one literal child `.gh-tree` relative to
that root, rejecting every symlink/reparse child; create a missing child through
the root handle and inspect it before use. Bind its identity for the request.
All subsequent names are single fixed or cryptographically generated basenames
relative to that retained parent handle, never a concatenated absolute path.
Reject a reparse/symlink/nonregular document or lock. Do not truncate the existing
document even when it is a hard link. Refuse observed identity/parent-entry drift
before commit; after a possibly completed commit, report drift with actual effect.

On Unix use `unix.Openat` with `O_DIRECTORY|O_NOFOLLOW|O_CLOEXEC` for the child,
`Fstat`/`Fstatat(AT_SYMLINK_NOFOLLOW)` for identities, and `Openat` with
`O_NOFOLLOW|O_CLOEXEC|O_CREAT|O_EXCL` for new regular files. Root canonicalization
belongs to the supplied registered-root authority. Do not recursively create or
follow a caller-supplied project path. Go 1.25 `os.Root` is traversal resistant but
follows in-root links and is not by itself the selected stricter no-child-link
policy. A pinned root/child can move: a noncooperating process may move the same
authorized directory outside its old pathname between observations. The contract
authorizes that originally opened object and never follows its substituted
pathname to another object; it does not promise continuous current-path ancestry.
Privileged mount changes are outside this file-storage contract.

Windows uses x/sys/windows `NtCreateFile`, `OBJECT_ATTRIBUTES.RootDirectory`,
`OBJ_DONT_REPARSE`, `FILE_OPEN_REPARSE_POINT`, and handle attribute/identity checks.
The reparse-open flag can successfully open the reparse object itself, so checking
`FILE_ATTRIBUTE_REPARSE_POINT` is mandatory. Keep root/child guards opened with
`FILE_TRAVERSE|FILE_READ_ATTRIBUTES|SYNCHRONIZE`, sharing READ|WRITE but not DELETE.
These guards prevent ordinary directory rename/deletion while held; READ-only
sharing is unsuitable because destination rename internally needs directory
write access. Relative single-name operations remain bound to the original
object if its metadata changes. No privileged namespace/volume-mutation sandbox
is claimed. All handles are noninheritable and deterministically closed.

Explicit --state/--config remains user-selected scope, including external or
linked parent directories. Composition resolves relative overrides once against
startup cwd. Persistence opens that selected parent, obtains its physical identity
and binds it for the operation; it does not impose project ancestry. A final
explicit file symlink may be resolved once as explicit user scope, then the
resolved ordinary target parent/name is bound; a later link is never followed.
Unsupported special/device/network profiles refuse writes without changing bytes.

## 2. Cooperative lock and expected version

Use one stable sibling `<basename>.lock` inode/file per document. Never unlink,
replace or age-expire that file: doing so can split waiters between lock objects.
Open/create without truncation and reject links/nonregular files. Windows opens
the lock denying DELETE sharing and takes `LockFileEx(EXCLUSIVE|FAIL_IMMEDIATELY)`
on byte 0 length 1. Unix uses `flock(LOCK_EX|LOCK_NB)` with an in-process keyed
mutex as well. Retry acquisition on a bounded cancellable backoff; never use a
PID/mtime lease. Kernel release on handle close/process exit establishes stale
ownership cleanup, including PID reuse. A diagnostic nonce/PID in the lock is
informational only. Cooperating implementations agree never to remove lock files.
Unexpected lock identity changes refuse the operation; arbitrary lock-file
tampering is not a supported cooperating-writer scenario.

Under the exclusive lock, reload the entire document including unknown bytes and
compare the opaque version with the request. Version includes missing/present
state, bound store identity, length and SHA-256 of exact complete document bytes;
it is not a field-specific timestamp. Same bytes are the same content version;
an intervening edit restored byte-for-byte is not a semantic conflict. The codec
must reject duplicate/invalid/unsupported shapes and preserve unknown members.
A stale version returns Conflict/NotCommitted; never merge a stale whole snapshot.
Application owns any fresh-read/revalidated bounded-intent retry.

Immediately before publication reopen the target without following links and
recompare identity/content/security metadata with the observed original. Reject
detected external changes. This remains cooperative serialization, not universal
CAS: an arbitrary editor can write or replace the target after that comparison.
Post-observation cannot prove there was no such race. Document this limit rather
than claiming every externally replaced object is captured. A busy external
writer can also make a read non-snapshot; inconsistent bounded reads are refused.

## 3. Preparation and retained originals

Encode and validate the complete new document before publication. Alias/default
is one payload. Reserve an exclusive random same-directory operation manifest,
raw original-byte backup and payload; do not reuse predictable `.tmp` names.
Write fully, checking count and errors, apply supported permissions by handle,
flush payload/backup, and verify the data/metadata. A short write, flush or
precommit close failure refuses publication. Keep a native payload handle through
rename on Windows; its close is a postcommit step, not a fictional precondition.

When an original exists, also retain a no-replace hard link to its observed file
object before replacing its normal name. On Unix `linkat` uses the pinned parent,
then verifies the linked inode matches the opened original; on Windows
`NtSetInformationFile(FileLinkInformation=11)` uses the actual original handle
and retained parent. The x/sys constant named FileLinkInformation is 72 (extended
class), so use the explicitly reviewed native class 11/layout, as Go 1.25's own
`internal/syscall/windows/at_windows.go` does. Do not accidentally use 72 with the
ordinary layout. Retain that link after success: writes through an existing
editor handle continue to reach it. A separate flushed raw-byte backup remains
the immutable read/migration input even if the retained object receives late
writes. Refuse before commit if safe original retention is unavailable.

Manifests identify schema/operation nonce, store/parent/original/payload identities,
expected/proposed versions and generated basenames. They contain no executable
recovery instructions. A complete manifest must be flushed before publication;
Unix also flushes the directory entries. Keep the old bytes/object and payload
evidence across uncertain outcomes; never overwrite/delete the live target to
make recovery match an old manifest. Restart holds the same lock and reads current
identities/versions. It can report old/new/changed/uncertain facts, but never
automatically reap an artifact by filename/PID/mtime alone or replay a commit.
Abandoned uncommitted payloads are disposable only after exact ownership and
unreferenced identity are proved; retained originals/migration backups are not
automatically purged. Bound retention space/admission and report concrete paths;
space exhaustion refuses a new write instead of evicting recovery data.

## 4. Native replacement and permissions

Windows selects local NTFS with successfully supported native FileRenameInformationEx
operations. Use `NtSetInformationFile` class 65 on the retained payload handle,
with `RootDirectory=retainedParent`, a single UTF-16 basename and flags
`FILE_RENAME_REPLACE_IF_EXISTS|FILE_RENAME_POSIX_SEMANTICS`. Missing-file creation
uses flags 0, so a concurrently created destination is not replaced. Construct
the architecture-correct variable buffer using Go `unsafe.Offsetof`; require
DELETE access on the payload, not a path-based reopen. This is the same modern
native mechanism selected first by Go 1.25 Root.Rename. Do not take Go's broad
fallback to an older replacement operation in the storage safety path: unsupported
capability refuses with the old file intact. File readers use native sharing
READ|WRITE|DELETE; a normal external reader that denies sharing may cause a
truthful SharingViolation/NotCommitted instead of a retry that changes mechanism.

Prepublication metadata preparation uses `GetSecurityInfo`/`SetSecurityInfo` on
the actual old/payload handles, then queries and compares the result. Preserve
owner, group, ordered DACL, DACL protection/inheritance, mandatory label and
resource attributes; query CAP scope and refuse nonempty unsupported CAP rather
than dropping it. LABEL/ATTRIBUTE/SCOPE queries require READ_CONTROL, while setting
label requires WRITE_OWNER; requesting those rights only on owned payloads avoids
modifying the original. Keep old read-only behavior: refuse an existing read-only
destination; do not set IGNORE_READONLY. Creation uses a protected user-only DACL
for per-user state; project payloads preserve existing access, and new project
files use normal inherited project access. Go mode 0600 is not a Windows ACL.
This bounded profile does not promise audit-SACL/security-policy replication,
EFS, compression, alternate streams or other special metadata: identify/refuse
unsupported access-affecting metadata, and retain the entire original object.
Custom audit-preservation requirements need an explicitly supported native
profile; do not claim full security-descriptor preservation from the DACL probe.

`ReplaceFileW` is not selected for final publication. Although it merges DACLs
and several file attributes, its documented error states can leave the replaced
file under another name or absent, and REPLACEFILE_WRITE_THROUGH is unsupported.
Those semantics do not justify a blanket atomic-visibility/success claim.

Unix uses one same-directory `renameat` of a fully prepared payload to the target.
For expected absence use no-replace publication (`linkat` of the completed owned
payload followed by removal of only its owned temporary name, or a proved native
RENAME_NOREPLACE), never an overwriting rename. Preserve uid/gid/mode through
`Fchown` then `Fchmod`, verify with `Fstat`, and flush the file. Refuse permission
changes that cannot be reproduced. Linux ACL/xattr metadata can be inspected/copied
through Flistxattr/Fgetxattr/Fsetxattr; Darwin has Flistxattr and native fgetattrlist
extended-security queries; FreeBSD has Extattr*Fd and native ACL queries. Initial
simple-file support must positively establish its supported metadata profile,
not assume Unix mode bits describe an ACL-bearing file. Copy verified supported
metadata or refuse unchanged. Those additional platform metadata paths still
require native implementation evidence; the Linux probe proves ordinary mode
preservation only. This is a narrow helper in Persistence, not a storage framework.

## 5. Visibility, crashes and truthful results

Visibility means fresh cooperating opens see a complete old or complete new
document; existing open handles can continue seeing the old object. It does not
mean all readers switch simultaneously, multiple stores form a transaction, or
an editor cannot later modify the new file. Readers may also get honest sharing/
availability errors. Local supported filesystems are required; network/FUSE/cloud
providers do not inherit guarantees from a drive letter or a cross-build.

| Last established stage | Outcome/evidence |
|---|---|
| Conflict/cancel/prepare/metadata/flush error before native publication | NotCommitted; target not modified by this operation, owned temporary artifacts may remain after a process crash |
| Native publication returned success | Committed effect; cancellation cannot roll it back |
| Unix file and following parent-directory fsync succeeded | Committed, with the supported filesystem's crash-durability contract; not a hardware power-loss guarantee |
| Windows native publication succeeded | CommittedDurabilityUncertain: data FlushFileBuffers was required, but no portable unprivileged directory/namespace durability barrier is claimed |
| Publication succeeded but directory flush/close/final metadata observation failed | Known committed effect with durability/cleanup/observation diagnostics; do not turn it into NotCommitted |
| Publication may have run, result lost/process died, or subsequent external changes prevent attribution | Indeterminate with current observations and retained recovery locations; reload before any semantic retry |

The process-crash stages are: acquired lock; original retained; payload complete;
manifest flushed; publication not yet issued; native publication issued; success
known; directory flush; outcome published. Inject each boundary in final tests.
After restarting, same target bytes alone do not prove historical causality.
Keep commit effect, observed current version, cleanup errors and durability as
distinct facts in the Application-owned result. Do not promise fsync withstands
every drive/controller failure. No global disk/volume flush requiring privileges.

## 6. Evidence and limits

`probe_windows_test.go` uses Go 1.25.0/x/sys 0.44.0 on Windows
10.0.26200/NTFS. Six tests pass (directory test has two sharing subcases): native
relative replace and no-replace, 100 replacements with complete concurrent native
reads, owner/group/protected-DACL equality, original hardlink late writes,
exclusive create, no-delete directory/ancestor pins, junction-object rejection,
cross-process byte lock and ungraceful-exit lock release. The initial test draft
needed SYNCHRONIZE on NtCreateFile and WRITE_OWNER on payload metadata handles;
final source/log record the corrected native access contract. Directory flush on
the ordinary guard returns access denied, as recorded, rather than durability.
An additional security-component probe queries label/resource/CAP using
READ_CONTROL and copies an explicit low-integrity label without privilege; the
observed descriptor is `S:AI(ML;;NW;;;LW)`. Nonempty resource attributes/CAP and
audit-SACL preservation are not established by this case.

`probe_linux.py` runs as UID65534 on native WSL2 Linux /tmp in existing
openSUSE-Leap-15.5. It proves no-follow rejection, pinned-object behavior after a
real outside rename and substituted symlink, 100 atomic replacements with
complete readers, file/directory fsync, mode preservation, late writes to the
retained original, cross-process flock and process-exit release, and crash exits
immediately before/after rename producing complete old/new targets. It is not a
Go implementation test, power-cut test, universal filesystem test or macOS/FreeBSD
proof. Native tests for resource attributes/CAP refusal, all metadata
profiles, Windows crash/short-write failures, reparse mutation, process races,
restart recovery and final Application outcomes remain required at exact product
SHA. All raw evidence is outside the frozen worktree and must be archived/hash
checked if adopted into the design.

Primary references (local Go source inspected, online native semantics checked):

- [Go 1.25 os.Root](https://github.com/golang/go/blob/go1.25.0/src/os/root.go) and [Windows handle-relative primitives](https://github.com/golang/go/blob/go1.25.0/src/internal/syscall/windows/at_windows.go).
- [Native FILE_RENAME_INFORMATION](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_rename_information): retained handle and target-name visibility semantics.
- [CreateFile sharing](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew): delete sharing and handle lifetime.
- [ReplaceFileW](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-replacefilew): supported metadata and nontransactional error states.
- [FlushFileBuffers](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers): file flushing and privileged volume flushing.
- [SECURITY_INFORMATION](https://learn.microsoft.com/en-us/windows/win32/secauthz/security-information), [GetSecurityInfo](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-getsecurityinfo), [SetSecurityInfo](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-setsecurityinfo): explicit security components/access rights and races.
- [rename(2)](https://man7.org/linux/man-pages/man2/rename.2.html): same-filesystem atomic replacement and relative-directory operations.

Suggested verification refinements: V-PER-02 must distinguish content version,
cooperative serialization, retained observed originals, external-editor gap,
native visibility and durability. V-PER-04 must assert never following a substituted
object, refusal on observed drift and Windows pins, while explicitly testing and
documenting the Unix directory-object/current-path distinction.
