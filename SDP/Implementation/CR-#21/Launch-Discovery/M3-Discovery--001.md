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
Ten local Windows test functions pass, including nested collision controls,
partial malformed projects, exact whitespace/colon argv, manager/member drift,
Make file precedence and semantic target order, saved overrides/unknown providers,
candidate limits, cancellation and concurrent immutable copies. On ordinary
case-insensitive Windows filesystems a physical Makefile also answers the earlier
makefile lookup; the selected lookup is explicitly bound in argv and source.
One test initially expected case-sensitive lookup there; corrected to the actual
native Windows policy. Source changes remain incomplete pending additional bounded
enumeration, native junction/root replacement, Linux execution/all12 builds and
configured native CI. This is a durable partial checkpoint, not a review candidate.
