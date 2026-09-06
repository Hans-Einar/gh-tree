# Git adapter — M3 in progress

Issue #61 owns this folder under Application--Git and BoundaryTypes FROZEN 1.0.0.
The first checkpoint supplies the private bounded command transport and physical
directory observations needed by the adapter. It is not a GitFacts/GitMutations
implementation candidate. No legacy package is called and no live rewrite path
exists in this checkpoint. Remaining port and native publication work follows.

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

Physical Windows observation uses a fully shared read-attributes handle, native
final path, volume/file identity and creation stamp. Unix observes an open directory
and compares its native device/inode to the current resolved name; birth/change
stamps remain source observations. These are short-lived read facts, not retained
mutation handles. Protected publication must independently acquire parents.

Tests use owned temporary repositories/configuration and a self-executing native
test process for literal argv, separate output bounds, before/during cancellation,
root reap and descendant-held pipe controls. Native SHA-1/SHA-256 status tests
verify exact index bytes, object identity and mtime remain unchanged. No user/global
configuration is modified. Run `go test ./internal/git -count=1`.
