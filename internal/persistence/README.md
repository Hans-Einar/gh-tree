# Persistence

Issue #63 implements the FROZEN Application--Persistence 1.0.0 boundary in
isolated milestones. Current milestones supply private schema/document/version
primitives, native Windows/Unix mechanisms, an explicit-location constructor and
three typed load paths. Public commits and manifest recovery are still pending;
this package is not connected to the product entry point.

The codec maps all specified schema0/schema1 user configuration, preferences,
and run configuration fields to the accepted immutable API documents. Only
`stripPrefixes` accepts null. Missing fields and explicit empty lists/strings
remain distinct. Loads retain schema0 without migrating; encoding requires an
Application-proposed schema1 document. Empty document candidates preserve absent
fields for Application's default policy. No codec selects active worktree,
default alias, migration winner, provider executable or repository association.

Every input first passes the bounded strict JSON predicate: 4 MiB, depth64,
valid UTF-8 and surrogate escapes, unique decoded member names at every depth,
one object envelope and a supported integer schema. Forward integer text is
retained without overflow into uint32. All known shape/presence checks use the
API constructors, and final encoded bytes pass the exact byte/depth limits again.

Unknown values retain their raw bytes and order at every known object level.
`verifyRetention` checks the guarded current document's retained members against
the proposal, including members beneath removed known objects. JSON comparison
permits equivalent object order, escapes, whitespace and exact decimal notation
without float rounding or huge-exponent arithmetic. Entire original document
bytes are a separate immutable request-private string for later raw backup.

Private canonical scoped keys are `remote:` or `local-common:` followed by
canonical unpadded base64url of the exact opaque RepositoryID token bytes. This
also preserves non-UTF-8 local tokens. Decoding a key establishes no observed
remote/local association or filesystem authority. Legacy keys retain case and
all exact values; noncurrent scoped records remain in the document.

Versions hash every original byte and distinguish family, physical store,
presence and selected run WorktreeID/root identity. The private physical-store
token hashes the native parent identity (or nearest existing absence anchor),
literal remaining component sequence and fixed basename. Tokens are equality
evidence; native requests must independently acquire/revalidate the binding and
never decode a token into a filesystem path. Byte-for-byte restoration in the
same binding has the same content version; native file/security comparison is
a separate publication precondition.

Verification for this milestone: `go test ./internal/persistence -count=1`,
`go vet ./internal/persistence`, and the bounded `FuzzDocumentCodec` target.
Tests exercise all nested known/unknown families, exact ordered/whitespace intent,
null/empty/default distinctions, forward/overflow schemas, malformed/reserved
shapes, escaped duplicate keys, UTF-8/surrogates, limits, lossless retained values,
copied buffers/concurrency, whole-byte and foreign-store/root versions.

Windows acquisition retains every ancestor with actual directory list/data-read
access and no-delete sharing, opens descendants through NtCreateFile RootDirectory
and no-reparse flags, checks native file IDs/reparse attributes, and reports a
missing anchor without writing. It implements local NTFS selection, bounded
double-read consistency checks, and permanent LockFileEx byte0/length1 locks
with context cancellation and at most a five-second retry budget. Read handles
share deletion; lock and directory handles do not. All resources stay private
to the request and close explicitly. Exact Git-issued directory identity matches
the shared native FileIdInfo/birth-filetime convention.

Native Windows amd64 and386 tests exercise real acquisition/absence/read,
ancestor rename exclusion, actual data-read child protection, in-place junction
conversion followed by retained-parent relative refusal, independent handle/
store lock exclusion, cancellation and killed-process kernel lock release.
Empty directory conversion can succeed while its directory guard is held;
relative operations then refuse, preserving the accepted storage limitation.
Windows ARM64 also executed these helpers in exact-source CI34054966709.

Unix acquisition uses no-follow, close-on-exec, nonblocking Openat descriptors
and native Fstat/statx or BSD birth observations; nonblocking opens let special
objects such as FIFOs refuse without hanging before their type can be checked.
Requests retain ancestor descriptors, distinguish the missing anchor/components,
and can revalidate every named edge. Moved original descriptors remain bound to
that original object; an observed substituted pathname refuses. Filesystem
inspection excludes unknown/network/FUSE profiles. Recognition for acquisition
does not certify a filesystem's later metadata/publication/durability profile.

Unix locks combine a reference-counted, inode-keyed process mutex with flock,
recheck the named lock object under ownership, and never unlink the permanent
file. Context/timeout/error exits close resources before releasing local ownership.
Native unprivileged WSL Linux/ext4 tests execute read/absence, noninheritance,
FIFO/link refusal, moved-object/path distinction, concurrent handle/store locks,
cancellation, mutex-map release and killed-process kernel lock release. All nine
Unix architecture test binaries compile. CI34054966709 at P2b source4c5e5be
executed the package on macOS26 ARM64, Linux/race, Windows amd64/ARM64 with18
SUCCESS jobs and the sole pre-Runtime helper skip. FreeBSD execution remains pending.

Private publication primitives now select class65/class11 with pointer-aligned
native layouts on Windows, and Renameat/Linkat on Unix. They require the future
request owner to complete metadata, version and manifest barriers before use.
Native Windows amd64/386 and unprivileged Linux/ext4 tests prove no-replace
competitor refusal, retained-object late writes alongside an unchanged raw
backup, and complete fresh-reader visibility during repeated replacement.
These mechanism tests do not establish a complete Storage commit/recovery path.

Windows metadata inspection queries owner/group/DACL plus mandatory labels,
resource attributes, CAP scopes, process trust labels and access filters. Only
mandatory labels are currently supported among those access-affecting SACL ACEs;
the other nonempty or unreadable profiles refuse. Ordered complete DACL ACEs,
protection/inheritance and label policy are copied and independently re-queried.
Read-only/special attributes, alternate data streams and native EAs refuse.
Per-user exclusive creation accepts a protected current-user-only ACL before
writing bytes. These private functions have native amd64/386 evidence including
actual resource attribute refusal, low-label and deny/allow ordered DACL copies.
The accepted ordinary-account profile excludes audit-only SACL replication;
limited READ_CONTROL queries do not prove audit ACE absence or a full security
descriptor copy. There is no public audit option or privilege escalation.

Unix metadata helpers perform bounded native attribute enumeration, double-read
comparison, Fchown then Fchmod, exact attribute copy and post-copy verification.
Linux copies user attributes and POSIX access ACLs, rejects other returned
security namespaces and unsupported inode flags; extents allocation is allowed.
Native unprivileged Linux/ext4 tests prove named-user ACL/uid/gid/mode and xattr
copying, extra inherited metadata refusal and native-query error propagation.
Darwin separately queries fgetattrlist extended security without RETURNED_ATTRS:
XNU reports unsupported requested ACL metadata as EINVAL and supported NULL ACL
as a zero-length reference. Complete filesec/no-ACL sentinel validation follows;
extended ACLs refuse. The first bitmap-based probe refused ordinary APFS absence
in CI34056596058; this corrected selector awaits its new exact-source native run.
FreeBSD queries native POSIX/NFS4 ACLs and direct ExtattrListFd for both namespaces,
preserving errors that the pinned convenience wrappers hide. Darwin/FreeBSD
metadata mechanisms currently have cross-build evidence only; FreeBSD unprivileged
system-namespace enumeration must be proved before claiming ordinary-user support.

New binds explicit absolute configuration/preferences locations once, including
an explicit link/junction selection, and rejects coincident file objects or
document/permanent-lock namespaces across families. Loads retain no native handles
after returning; each reacquires the binding and rejects a replaced observed
parent. Newly appeared missing components can be adopted and thereafter cannot
be silently replaced. Run loads validate the supplied Git root identity/profile,
then acquire only `.gh-tree/run.json` without following child redirects. They
also refuse overlap with either bound user store. No load creates directories,
lock files, migrations or defaults on disk. Three typed load methods return the
strict documents/whole-byte versions or explicit absent/corrupt/unsupported/
unavailable observations; recovery manifest observation remains pending.

Pending native milestones retain the complete selected contract: request-owned
no-follow acquisition, supported metadata, permanent cooperative locks, missing
parent revalidation, exact class65/class11 or Renameat/Linkat publication,
immutable raw plus late-write-retained originals, bounded recovery admission,
stable persisted recovery IDs and truthful commit/error/cancel outcomes. Native
Windows/Linux/macOS/FreeBSD execution, faults/crashes/concurrent processes and
all twelve builds remain separate mandatory acceptance evidence. Arbitrary
external edits in the final-check/publication gap remain outside cooperative
CAS; the implementation must test and report this accepted limitation.
