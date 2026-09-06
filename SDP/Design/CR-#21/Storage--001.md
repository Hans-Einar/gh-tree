# Storage--001 — concrete persistence and migration contract

State: ACCEPTED DESIGN (effective after PR #54 merge) correction for DES52-H03 under #52; normative REFDES appendix.
Select Feasibility/Persistence.md sections 1–5 as the native mechanism, with the
schema, semantic ownership and explicit limits below. Its section 6 records
bounded evidence and remaining product/native-platform proof. No BC is frozen.

## Store scope and supported publication

Persistence binds the authorized directory object and fixed document basename.
Project run.json is one literal `.gh-tree` child below the selected Git-issued
worktree root. Open it through retained root handles with no-follow/no-reparse
checks; never follow a substituted child. Windows retains actual directory list/
data-read guards with READ|WRITE/no-DELETE sharing, not metadata-only handles.
Unix uses directory-relative Openat/Fstatat/Renameat operations
on retained descriptors. Refuse observed identity/parent drift. Unix authority is
the originally opened directory object, not a promise of continuous current-path
ancestry after another actor moves that same object. This is a precise capability
scope, not permission to resolve a replacement path outside the worktree.

Explicit --config/--state resolves once against startup cwd into the selected
physical parent/name, including an explicitly selected external/link target.
Defaults remain the existing platform locations and are evaluated lazily after
flags/help/version. Create missing application-owned path components relative to
the nearest validated existing parent; never follow newly substituted components.
Project scope remains stricter and accepts only its literal `.gh-tree` child.
Reads distinguish unavailable/unsupported/corrupt from absent; writes do not
repair corruption automatically. Network/cloud/FUSE/special-object profiles do
not receive local-filesystem guarantees without their own proven capability.

Use a permanent never-unlinked sibling `<basename>.lock` with nonblocking bounded
LockFileEx byte lock on Windows or flock plus keyed in-process mutex on Unix.
Kernel handle/process lifetime releases ownership; PID/age is never authority to
delete a lock. Under that lock compare the whole-document version, prepare an
exclusive same-directory manifest/payload/raw backup and retain a hardlink to the
observed original object. All writes and required file flushes complete before
publication. Preserve the original observed bytes separately from its live inode,
which may receive later writes through another handle. Refuse if retention or
supported permissions cannot be established; never truncate the destination.

Windows local NTFS publication uses the retained payload handle and parent with
NtSetInformationFile class65 FileRenameInformationEx, REPLACE_IF_EXISTS|POSIX
semantics for expected presence and flags0 for expected absence. Use reviewed
architecture-correct native layout; unsupported native behavior refuses rather
than falling back. The retained original hardlink uses native class11, not the
x/sys class72 constant with the wrong layout. Native readers share READ|WRITE|
DELETE or report genuine sharing failures. Preserve/verify owner/group/DACL and
supported label/resource security; refuse unsupported CAP or metadata profiles.
New per-user files use a protected user-only ACL; mode0600 alone is insufficient
on Windows. Existing read-only destinations refuse without clearing attributes.
ReplaceFileW is not the selected commit primitive because its failure states can
rename/remove the target and its WRITE_THROUGH flag is unsupported.

Unix uses same-directory Renameat for expected presence, and no-replace Linkat
publication (or proved native equivalent) for absence. Preserve uid/gid/mode and
positively detect supported ACL/xattr metadata before publishing; copy and verify
supported metadata or refuse unchanged. Fchown precedes Fchmod, followed by file
flush and parent-directory fsync. No fallback discards ACLs because a worker only
implemented mode bits. The platform-specific inspection APIs and limits are
enumerated in Feasibility/Persistence.md and require native implementation proof.

The native publication return is the commit point. Before invocation, errors or
cancel are NotCommitted, although owned preparation artifacts may remain. Native
success is a known committed effect even if later close/flush/observation fails.
Unix successful file/directory fsync yields Committed under the supported local
filesystem's crash contract. Windows success yields CommittedDurabilityUncertain:
payload flushing does not prove namespace/power-loss durability. Lost native
result or ambiguous subsequent state is Indeterminate. Exact current bytes after
restart do not prove historical causality. Application publishes actual committed
intent, preserves diagnostics and reconciles indeterminate state before retry.

All journal/recovery names carry a cryptographic operation nonce and recorded
store/parent/payload/original identities. Restart reacquires the stable lock and
observes facts; it does not automatically replay or overwrite the current target.
Keep original/migration backups until conscious cleanup, never age-purge them;
space limits refuse further commits. Report recovery locations as part of the
outcome. Codec/version errors cannot turn an existing document into empty defaults.

## Explicit concurrency and durability limits

Cooperating writers serialize and compare whole-document versions, preventing
stale gh-tree snapshots from overwriting each other. Immediately before commit,
recompare target content/identity/security and reject detected external changes.
This is not universal CAS against an arbitrary editor's write/atomic replacement
after the last check. The retained original is the observed object; this protocol
does not claim to capture every unobserved externally replaced object. Such writers
are outside its cooperative commit guarantee. Never use a post-check to claim no
intervening effect. Preserve detected conflicts/recovery and report uncertain
attribution. The stronger Git worktree publisher is a separate contract.

Fixed directory-object authority prevents following a substituted pathname; it
does not stop Unix peers moving the authorized object elsewhere. Windows guards
add ordinary rename/deletion exclusion. Privileged mount/volume manipulation and
hardware power-loss guarantees are not promised. These explicit limits replace
the earlier unspecific phrase “reviewed platform replacement.”

## Codec version 1 and backward compatibility

All three document families use a top-level integer `schemaVersion: 1`. Missing
schemaVersion or explicit integer0 is legacy version0; any unknown positive version is read as
UnsupportedVersion and never rewritten. Null/array/nonobject documents, duplicate
JSON object keys at any depth, invalid known-field types and noninteger versions
are corrupt/unsupported, not defaults. Negative versions and invalid UTF-8 also
refuse; JSON decoding must not silently replace path/script bytes. Maximum document input is 4MiB, with a
bounded nesting depth of64 and explicit limit diagnostic. Missing files alone
produce defaults. Parsing preserves exact strings rather than trimming paths,
script names or aliases into different identities.

Persistence codecs retain unknown members at every known object level as opaque
JSON values and preserve unknown provider definitions. A write patches only its
authorized known fields and retains other known/unknown values. Formatting may
normalize in the new document; the complete original bytes are retained in the
backup. Duplicate/ambiguous fields refuse instead of choosing a destructive winner.
An unknown provider can be preserved and shown unavailable; Discovery, not the
storage codec, decides whether that entry can resolve into an invocation.

| Family | Version1 known fields and retained version0 fields |
|---|---|
| User config | `stripPrefixes`, legacy `repos`, new `scopedRepos`. scopedRepos maps canonical remote RepositoryID tokens to objects with `worktrees` keyed by exact configured-target name; each target retains path/branch and unknown members. Legacy repos is retained intact as migration input. |
| UI preferences | Legacy `lastFolders`/`lastWorktrees`, new `scopedPreferences`. Remote-scope entries may contain `lastFolder`; LocalCommon-scope entries may contain `activeWorktree:{administrativeKey,lastKnownPath}`. Parent key supplies repository scope. lastKnownPath is recovery/display metadata; Git-issued WorktreeID is active authority. Legacy maps remain intact. |
| Worktree run | Existing `default` alias and `launch` alias map, with existing provider/dir/script/targets/command fields and retained unknown members. Alias/default commit in one document. The shape remains familiar; provider semantics remain Discovery/Application-owned. |

Canonical scope keys serialize opaque API/Domain scope tokens as unambiguous
strings; codecs do not infer scope by case-folding arbitrary keys or filesystem
paths. The token carries its Remote/LocalCommon namespace. Application/Git/GitHub
establish that scope under API A1; Domain has no JSON tags or path/URL parser.
Unknown noncurrent scope entries are retained, not pruned because today's query
cannot observe their repositories. A legacy field colliding with the reserved
new field in an ambiguous shape is a migration conflict, never overwritten.

Compatibility details:

- Missing or null legacy stripPrefixes retains the existing default list;
  explicit `[]` remains an intentionally empty list. New snapshots copy defaults.
- Configured target names and launch aliases remain exact case-sensitive names.
  Display order uses folded label, exact bytes and stable identity as a total
  order; label ordering never defines repository equality.
- Absent/empty run default means none. First save becomes default only if none
  exists or the user chose make-default. Existing aliases need explicit version-
  bound replacement intent. Command overrides and ordered Make targets survive.
- Missing/empty project dir means root; nonempty bytes are preserved exactly and
  Discovery validates them. Empty saved active-worktree paths are invalid intent,
  never `filepath.Clean("") == "."` as another target.
- Only absolute valid XDG/default roots are used. Explicit relative overrides
  become absolute once against startup cwd, not whichever worktree becomes active.

## Semantic migration and commit ownership

Load returns decoded legacy/scoped values, source version and diagnostics without
writing. Application chooses unambiguous mappings and commits through the same
versioned native protocol on an authorized durable intent change. No eager write
is needed for startup/help/version. The exact original bytes are backed up before
the first version1 commit and retained thereafter with ordinary recovery records.

Prefer an existing validated canonical-scope entry. For legacy owner/name keys,
collect all exact/case-equivalent candidates; conflicting values or uncertain host
association remain diagnosed legacy candidates, never an arbitrary map-iteration
winner. A legacy active path can migrate only if fresh Git inventory uniquely
matches it to a WorktreeID in the current LocalCommon scope. Namespace/config
mapping needs an explicit unambiguous remote host/repository association; absent
host information is not silently invented. If that cannot be established, retain
legacy bytes/intent and use the documented current/deterministic fallback until
new navigation/selection writes an explicit canonical intent. No ambiguous entries
are deleted, and an incomplete source cannot justify removing saved intent.

Application supplies one validated intent patch against an expected whole-document
version. Persistence never chooses active worktree, alias/default or a migration
winner. A version conflict requires a fresh load and revalidated Application
intent; no stale full-map overwrite or automatic broad merge. Active context is
published only for its known committed result; indeterminate storage blocks the
dependent context transition until reconciliation. Defaults, migration, retained
unknown fields and every commit-stage result are required V-PER/APP proof.
