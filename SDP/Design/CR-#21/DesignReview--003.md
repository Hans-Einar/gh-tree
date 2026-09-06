# DesignReview--003 — correction re-review and remaining Windows cwd gap

Disposition: CHANGES_REQUIRED at 8c109b1e4308eb301d7c1418292044d714c2f33b.
Authority #52/#21, PR #54. Fresh independent reviewer frozen_design_review reopened
the complete correction diff, Storage/Cwd/Persistence records, all new archived
sources/logs and their hashes. Product baseline and migration/Slice maps unchanged;
ledger prefix/reference checks pass. Reviewer executed no probes and changed no files.

Resolved at design level: DES52-H03 concrete storage/commit/schema/limits;
DES52-M01 exact PR merge-base and StashPatch reads; DES52-M02 reliable bounded
cleanup transfer. Full implementation/platform evidence remains required.

DES52-H04 remains HIGH. No-delete-sharing directory handles prevent ordinary
rename but do not prevent an emptied directory's in-place reparse conversion.
Microsoft's FSCTL_SET_REPARSE_POINT contract permits FILE_WRITE_ATTRIBUTES access,
which sharing flags do not exclude. Root's native fixture created an outside
target under its own temporary tree, changed the held empty cwd into a junction,
then observed the child read the other-directory marker. ShareREAD alone also
failed when the setter requested only write-attributes.

Independent bounded followup found actual data-read/list-read child anchors with
no DELETE sharing block native class65 POSIX replacement and independent deletion,
keeping the parent nonempty so reparse conversion fails145. Metadata-only anchors
are bypassed and cannot be described as interlocks. Handle-relative Persistence
operations refuse converted parents with STATUS_REPARSE_POINT_NOT_RESOLVED instead
of following them outside. These facts require precise native rights in the final
contract; they do not themselves prove when a Windows child has acquired cwd.

Root separately tested CreateProcess(CREATE_SUSPENDED), removed the temporary
anchor after its return, then set the junction before ResumeThread. The child
still read the other-directory marker. Native process-parameter observation showed
CurrentDirectory.Handle zero before Resume, both with explicit cwd and null cwd
inheritance. Consequently, anchor release at CreateProcess return is insufficient.

The remaining correction must bind the selected cwd through actual native startup,
not another timed check, sleep or assumed return-point barrier. A separate bounded
Windows startup reviewer is testing mechanisms. No corrected mechanism has yet
been accepted in this record, and no merge/BC/product authority follows from it.

## Proposed complete correction — exact-HEAD re-review pending

WindowsBroker--001 selects a Runtime-private native-architecture broker, strong
data-read/list-read guards/nonempty anchor, actual target initial-breakpoint /
PEB/cwd-handle FileIdInfo check, and pending-event detachment. Native386 cannot
debug native64 directly (error50); an embedded native broker preserves that launch
capability without downloads, runtime compilation or extra public release assets.
The ordinary-build and reproducible source-closure/payload contract is explicit.

Bounded native evidence now includes64->64/64->32/32->32 barriers, actual386 parent
to embedded/extracted native64 broker to native64 user/descendant with ConPTY,
debugger/anchor absent to user code, input/output/100x30 resize, inner quiescence,
broker Release and final outer Job0, plus forced whole-outer cleanup. Two controlled
helper rebuilds match; ARM64 is compile-only evidence and now has an explicit
native hosted-runner acceptance gate. Waiting for outer membership1 before Release
was disproved and is forbidden by the selected teardown sequence.

Master read all three followup reports and actual combined broker/embed source and
captured results. All43 declared source/log/build-output hashes across the three
followups matched;37 source/report files were archived, while temporary executable
build outputs were deliberately not committed. ArchiveManifest files distinguish
the two. Earlier failed reparse/suspended/inheritance/injection experiments remain
available as rejected evidence. This is mechanism evidence, not production proof.

Independent review must now assess the complete correction, helper source/build/
extraction topology, native profiles and no new cross-contract contradiction.
DES52-H04 is not marked resolved by its author. The reviewed8c109b1 commit has
all18 CI checks successful in runs34007362962/34007365156; a new corrected commit
needs its own exact-HEAD review and CI.
