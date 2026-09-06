# Passive Launch Discovery

Issue #64 implements `ports.LaunchDiscovery` through `New(Config{})`. The immutable
constructor registers npm and Make by default; explicit closed provider keys must
be unique. The adapter imports Application API/ports, Domain, standard libraries
and its own native read-only filesystem primitives. It never executes tools,
changes cwd, writes project files, reads run.json or calls another adapter.

`Discover` returns separately valid project definitions, supplied saved-entry
observations, scoped diagnostics and explicit completeness. Generated directories
are excluded case-insensitively. Limits default to depth5, 10,000 visited
directories, 10,000 definitions, 4MiB per manifest and 1MiB per Make line. Saved
input/member lists are separately bounded by the same candidate ceiling. Smaller
construction limits are supported; raising the frozen ceilings refuses. Directory
entries are read in batches64; retained child names use a bounded maximum heap,
so truncation retains the lexicographically earliest directories deterministically.
Diagnostic retention is capped at directory+candidate budgets plus one limit
notice. Enumeration can still inspect many ordinary entries; it checks context
between batches and owns no background scan. A native filesystem call itself is
not instantaneously interruptible.

Npm JSON rejects invalid UTF-8, unpaired surrogate escapes, duplicate object keys
at every depth, ambiguous scripts and nesting beyond64. Exact script bytes survive;
empty/invalid members remain diagnosed unavailable. Colocated regular pnpm lock,
then yarn lock, then npm determine executable policy. Conflicts and packageManager
disagreement are diagnosed. Redirected or nonregular locks never qualify; actual
permission/I/O failures remain failures. Lock content is not interpreted and does
not choose a manager. Source versions bind their observed regularity/object
identity, manifest identity/content, exact project/worktree and executable policy.

Make selects GNUmakefile, makefile, Makefile in order and binds the selected lookup
with `-f`. On case-insensitive filesystems a physical Makefile also answers the
earlier case-insensitive makefile lookup. Only simple textual targets comprising
ASCII letters/digits/underscore/dash/dot are supported; leading dot/dash, assignment,
slash, macros, patterns and control syntax refuse. Recipes are not evaluated.
Includes/conditionals/dynamic syntax produce profile limitations; simple textual
members can remain available without claiming exhaustive native Make semantics.

`Resolve` reconstructs and reobserves the exact source; ID, source version,
worktree/root/project, manifest precedence, manager and saved override must match.
Ordered Make members share physical project/source/executable policy and preserve
every selected member in order. Saved selections retain exact alias, supplied
StorageVersion, member order and override presence. Duplicate aliases refuse;
unknown providers remain unavailable observations. Discovery cannot establish
freshness of the caller's saved store: Application reloads/checks it. No default
alias is chosen here. Two aliases may refer to the same underlying LaunchPointID;
a saved multi-target Make entry identifies its first member and separately binds
the complete order in its source version and resolved member list.

Invocation carries literal `npm run <single-script>` or `make -f <file> <targets>`
argv, inherited-base environment intent, Pipes mode and supplied positive geometry.
No shell source, executable lookup or native batch transport is constructed here.
Runtime alone performs actual argv/cwd execution and reviewed Windows batch handling.

Native Windows observation opens local DOS-volume roots and each child relative
to retained handles, rejecting reparse objects. NT relative opens use
OBJ_DONT_REPARSE/FILE_OPEN_REPARSE_POINT; directory handles deny delete sharing.
Identity is full aligned FileIdInfo volume+16-byte ID and unsigned creation FILETIME.
UNC/device/network roots and ambiguous Windows component spellings refuse this
initial observation profile. Native Unix walks canonical absolute roots and
project components through Openat/O_NOFOLLOW directory descriptors. Identity is
device plus uint64 inode in low8 little-endian bytes, then8 zero bytes. Linux uses
proved statx birth stamps, otherwise ctime; Darwin/FreeBSD use native birth stamps.
Consumers observe the supplied birth/change profile without silently upgrading it.

Manifest reads compare bounded bytes twice plus native identity/metadata and the
newly named object. Resolve rechecks every project component and the named root
before publishing a new CwdObservation; no native handle crosses the boundary.
These are interval observations. Unix may retain an original object moved by
another actor, but an observed path replacement refuses. Runtime independently
acquires actual cwd at Start. This is neither immutable future project code nor
a post-start filesystem sandbox, and no ordinary no-delete Windows guard is
claimed as Runtime's stronger startup barrier.

Tests cover real temporary sources, passivity, exact Unicode/whitespace/colon
identity, partial failures/cancellation, all default ceilings, duplicate/foreign
selection, saved overrides/order, same-size restored-mtime edits, manager and
preferred-Makefile drift, Windows native junction/conversion/replacement and
Unix nofollow/permissions/moved objects/stamp profiles. Race and all12 architecture
checks are distinct from native tests. Application/State/View/default roundtrip
and actual Runtime argv/cwd/batch launch remain downstream integration gates.
