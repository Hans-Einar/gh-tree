# M3 Launch Discovery contribution (#64)

Status: PARTIAL AUTHOR CHECKPOINT; neither review candidate nor integration authority.
Worker branch: codereview-21/layer-launchdiscovery, base412f33e477cec03cb6eafe7b846c9bcdd02c0a25.
Authority: resumed #21/#64; frozen Discovery/Runtime/Persistence and shared types1.0.0;
accepted REFDES/API/CwdAcquisition/Storage/Slices/MigrationMap/Verification.
Ownership: internal/launchdiscovery and this report only. Legacy remains untouched.

First milestone establishes private strict npm JSON and bounded textual Make
parsers, native no-follow directory-relative reads, full Windows FileIdInfo and
Unix device/inode/birth-or-change identity profiles. No provider executable,
process, storage read/write or cross-adapter dependency is introduced.

Actual local Windows/amd64 Go1.25.0 targeted `go test ./internal/launchdiscovery
-count=1 -v` passed four tests: strict Unicode/duplicate JSON, literal members,
Make grammar/line limit/cancel, native retained observation/byte cap/invalid profile
and directory replacement. Windows symbolic-link creation lacked privilege;
that negative is explicitly NOT PROVED by the successful test. Native junction
controls, Unix execution, full ports, scan/selection tests and all12 builds remain.

Next permitted worker action: complete bounded Discover/Resolve and saved binding,
add native reparse/symlink/adverse and legitimate controls, run actual local/Linux
and configured native CI, freeze/push the candidate for a fresh independent review.
Application/State/View/default persistence roundtrip and actual Runtime argv/cwd
Start remain downstream integration gates. No Slice or baseline finding closes.

Second partial milestone implements both exact ports, registered immutable
constructor policy, nested partial scans, manager/version policy, fresh exact
discovered/ordered/saved resolution, literal Invocation and native cwd observation.
Nine local Windows test functions pass, including nested collision controls,
partial malformed projects, exact whitespace/colon argv, manager/member drift,
Make file precedence and semantic target order, saved overrides/unknown providers,
candidate limits, cancellation and concurrent immutable copies. On ordinary
case-insensitive Windows filesystems a physical Makefile also answers the earlier
makefile lookup; the selected lookup is explicitly bound in argv and source.
One test initially expected case-sensitive lookup there; corrected to the actual
native Windows policy. Source changes remain incomplete pending additional bounded
enumeration, native junction/root replacement, Linux execution/all12 builds and
configured native CI. This is a durable partial checkpoint, not a review candidate.

Third partial milestone follows Master's clean prerequisite merge4bfcf7f of
canonicald5eeaac (Discovery source unchanged in that merge). Corrected actual
Darwin x/sys stat fields and 32-bit timestamp casts; all12 Discovery test binaries
now cross-compile with Go1.25/CGO0. Windows/amd64 executes11 tests including real
mount-point junction and already-open directory reparse conversion controls,
redirected lock/saved path refusal and same-path root replacement.
Linux/amd64 executes12 tests in existing WSL openSUSE-Leap-15.5 as actual
UID/GID65534, using /tmp/gh-tree-discovery.NBF6Ai on its native filesystem
(stat reports ext2/ext3 magic, not a /mnt fixture). Actual moved retained object,
symlinks, permission-denied partial facts and supplied change-stamp drift pass.
Two initial runner script attempts failed before tests due Windows argument
quoting/CRLF; normalized UTF-8/LF script then executed successfully.

The adapter now bounds diagnostic retention, retains deterministic smallest child
names at breadth caps, refuses duplicate saved alias authority, reopens the root
after scanning, double-reads bounded manifest bytes and issues distinct observation
IDs atomically. Remaining worker work: targeted bounds/passivity/foreign-selection
tests, architecture/race/vet, README and exact native CI candidate gate.
