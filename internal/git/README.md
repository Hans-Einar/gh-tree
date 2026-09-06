# Git adapter — M3 in progress

Issue #61 owns this folder under Application--Git and BoundaryTypes FROZEN 1.0.0.
The current checkpoint supplies ResolveLocal/ListWorktrees, private bounded command
transport, physical common/administrative identities and scoped remote mappings.
It is not a complete GitFacts/GitMutations candidate. No legacy package is called
and no live rewrite path exists. Remaining ports and native publication follow.

`New` snapshots its environment/current directory and validates bounded options.
Commands receive literal argv, explicit cwd and adapter-owned Git location/pathspec
settings. Config and signer/SSH inputs remain private. Machine stdout and stderr
are separately capped at 16 MiB and 256 KiB, with 30s reads, 120s mutations and
2s forced pipe drainage by default. Limits are configurable downward. Native error
diagnostics do not publish raw stderr or inherited credentials.

One owner waits each root and joins the Go command pipe copiers. Output overflow
cancels immediately. Canceled/killed roots and forced pipe drainage return explicit
cleanup uncertainty; this transport makes no Runtime descendant guarantee.
Windows subprocesses use hidden windows. Read commands disable optional index
writes, fsmonitor, untracked cache, automatic maintenance and replace-object lookup.
Their empty native protocol allowlist also prevents implicit promisor fetches.
Unproved partial-clone object-read profiles return an explicit diagnostic.

ReadCommits returns native commit messages/parents/signatures by exact endpoint,
with bounded pages whose authenticated cursors bind repository, full root OID,
traversal and shallow/graft source context. MergeBase returns Unique,
NoCommonAncestor or every ambiguous candidate. Neither read joins GitHub facts.

ListRefs preserves literal branch suffixes, native tag object versus peeled
commit, symbolic targets and cached-remote freshness. Mutable pages require the
returned source-bound cursor. ResolveExact checks full commit identity, selected
branch/ref equality and original remote/configuration binding. Equal OID bytes
without an observed mapping do not establish a remote association. Private fetch
root registration remains part of the upcoming Fetch implementation.

ListStashes binds full stash OID and the observed reflog occurrence independently
of display position. It validates native base/index/untracked parents and retains
duplicate occurrences instead of coalescing OIDs. Files refs use bounded original
log bytes; reftable uses native read-only reflog metadata and native pseudo-HEAD
queries. Legacy managed metadata remains descriptive. Actual Git2.43 files and
Git2.48.1 reftable controls pass for SHA-1/SHA-256; this proves no stash deletion.
Tests optionally select an explicit executable/ref backend with GH_TREE_TEST_GIT
and GH_TREE_TEST_REF_STORAGE, with no developer-local path in source.

Physical Windows observation uses a fully shared read-attributes handle, native
final path, full FileIdInfo uint64 volume/native16 file ID and creation stamp
`birth-filetime:<decimal uint64>`. Aligned native storage is required; a raw byte
array failed with ERROR_NOACCESS in the actual native control. Corrected native
Windows/amd64 and386 tests pass. Unix uses uint64 native device plus LE uint64
inode/zero-high8 FileID, `birth:<sec>:<nsec>` when proved, otherwise explicit
`change:<sec>:<nsec>`. Linux uses statx birth availability; Darwin/FreeBSD use
available native birth stamps. These are short-lived read facts, not retained
mutation handles. Protected publication must independently acquire parents.

Tests use owned temporary repositories/configuration and a self-executing native
test process for literal argv, separate output bounds, before/during cancellation,
root reap and descendant-held pipe controls. Native SHA-1/SHA-256 status tests
verify exact index bytes, object identity and mtime remain unchanged. Inventory tests
cover unborn/attached/detached, linked scopes versus clones, current/primary, locks,
missing roots with retained independent administrative Head, and changed remotes.
Remote identity convention is `lower-host/lower-owner/lower-name`, independently
coordinated with GitHub. URLs in bindings remove credentials. Native file remotes
have separate physical Remote scopes; relative/ambiguous transport profiles return
explicit diagnostics and partial completeness. No user/global configuration is
modified. Run `go test ./internal/git -count=1`.
