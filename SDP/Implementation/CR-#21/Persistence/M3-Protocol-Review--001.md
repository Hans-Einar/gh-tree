# M3 Persistence early protocol audit — #63 under #21

Disposition: **BOUNDED AUDIT COMPLETE / THREE CHANGES REQUIRED; NOT ACCEPTANCE**.
Fresh independent Reviewer; source is the explicitly incomplete P3d checkpoint
`5e8aaebce381b92856b1abb837443de0ce1c0728`, detached at
`C:/Users/hanse/GIT/gh-tree-wt/persistence-protocol-review`. Source worktree remained
clean and unchanged. Findings were sent promptly to Master for concurrent author
correction. No implementation, commit, merge, push, or user-data mutation by Reviewer.

Read authority: AGENTS.md, complete developmentInstructions.md and Issues #21/#63
including comments; FROZEN Application--Persistence and BoundaryTypes 1.0.0;
accepted Storage--001/Verification--001; Persistence README and actual milestone
record. Master M3-Adapters--001's bounded Windows security and mutable-observation/
stable-RecoveryID interpretations were read. Actual commit, store/load, binding,
manifest/journal, native preparation/publication/metadata helpers and public result
constructors were inspected; author summaries were navigation, not proof.

## Findings

### M363-PR01 — P1: Unix publishes a substituted staging entry after verifying another object

At `internal/persistence/commit.go:512` the final checks verify the retained
`publisher` object's identity/content/policy. `preparation_unix.go:26` discards that
object and `publication_unix.go:20` publishes the `.publication` pathname with
Renameat/Linkat. Its current directory entry is never compared with the verified
object during those final checks.

Independent `TestReviewUnixStagingSubstitution` changes only the staging entry at
the **final-check hook, before final checks**: move the generated `.publication`
entry to another owned name, then put a symlink to a separate owned external
sentinel at the old staging name. Both expected-absent and expected-present cases
return a valid **Committed / PublicationKnown=true / AppliedVerified** result while
the destination becomes that symlink. The supplemental current-read ELOOP does not
correct the false claim that the verified proposal was applied. The external
sentinel's bytes remain unchanged; the demonstrated mutation is the wrong target
object and false successful proposal effect.

Required correction: bind the entry consumed by the selected Unix publisher to
the verified regular proposal, reject observable staging substitution before
publication, and retain truthful actual effects if an issued publication consumes
an unexpected object. Include absent/present regular-file and symlink substitutions.
This reproduction is detectable drift **before** final checks. It does not rely
on, or widen, the accepted arbitrary external editor's target comparison/publication
gap; any remaining source-name race after the check needs explicit reasoning.

### M363-PR02 — P2: metadata verification precedes the write that changes it

`commit.go:413` applies original metadata before `writeComplete` at line424.
The post-write inspection at line427 discards the resulting metadata value;
the later payload-policy snapshot is compared only with itself, never with the
required original policy. On ordinary unprivileged Linux, writing clears set-ID
bits that `unixApplyMetadata` previously copied and verified.

Independent `TestReviewUnixModePreservedAfterWriting` writes a valid existing
configuration, sets and independently confirms mode04750, then performs an
ordinary public commit without fault injection. Actual result: valid Committed,
PublicationKnown=true, **nil error**, destination mode0750. The supported mode
profile silently changes. This is an existing implementation defect, not a missing
test matrix or unimplemented profile.

Required correction: establish the final required metadata after the byte writes,
preserve the mandated ownership-before-mode order, verify exact final policy
against the original before publication, and refuse unsupported preservation.
Add the real unprivileged set-ID control alongside ordinary mode/ACL/xattr controls.

### M363-PR03 — P2: recovery refusal erases an independently valid current load

`store.go:183` changes the load state to UnsupportedProfile for any joined native
unsupported error, including errors from the later recovery scan. The current
document has already been read and decoded, and remains nonnil. Passing this
state/document combination to `ports.NewLoadedUserConfig` at store.go:77 violates
the public load invariant and returns a zero invalid result.

Independent `TestReviewUnixRecoveryRefusalPreservesCurrent` successfully commits,
then adds an unsupported symlink entry in that store's reserved recovery prefix.
The current configuration is untouched. Fresh LoadUserConfig returns ELOOP plus
`invalid port value: user config load`, with **Valid=false, no document, no version**.
An auxiliary recovery failure has erased the independently usable current facts.

Required correction: distinguish acquisition/document-profile failure from later
recovery-profile errors. Preserve the established current state/document/version
and independent recoveries with supplemental diagnostics; continue refusing to
follow the unsupported entry. Apply the same behavior to all three typed loads.

## Executed evidence

Go1.25.0, CGO_ENABLED=0. Test-only Go overlays added independent tests without
changing frozen source. Native Windows amd64 and actual WSL openSUSE-Leap-15.5,
Linux6.18.33.2-microsoft-standard-WSL2 x86_64, UID/GID65534 (`nobody`), ext4
(`/usr/bin/findmnt` verified). Linux binary/fixtures were under owned
`/tmp/gh-tree-protocol-review.ct3r62`; Windows fixtures used test-owned TEMP paths.
No old denied-ACL fixture was touched. Test-created fixture cleanup completed.

| Selector/control | Windows amd64 | Linux amd64 ordinary account |
|---|---|---|
| Independent staging substitution, absent/present | Not run; Unix mechanism | FAIL, M363-PR01 |
| Independent original04750 mode preservation | Not applicable | FAIL, M363-PR02 |
| Independent recovery refusal/current preservation | Not run | FAIL, M363-PR03 |
| Independent known commit + delivery error + cancellation retains effect/recovery | PASS | PASS |
| Independent postcommit external edit keeps distinct Proposed/Current versions and known publication | PASS | PASS |
| Existing stale/foreign-store, seven fault boundaries, partial journal, torn tail, stable IDs/late original, same-byte artifact replacement, malformed/foreign manifest controls | PASS | PASS |

Final Windows selector:
`go test -overlay <overlay.json> ./internal/persistence -count=1 -run='TestReviewKnown|TestManifest|TestCommitFaults|TestCommitStale' -v`, exit0.
Linux: crosscompile with `go test -overlay <overlay.json> -c -o <binary> ./internal/persistence`,
then `runuser -u nobody -- env TMPDIR=<owned>/tmp <binary> -test.run='TestReview|TestManifest|TestCommitFaults|TestCommitStale' -test.v=true -test.timeout=30s`,
exit1 for exactly the three finding tests (PR01 has two failing subcases).

The first Windows external-edit fixture used os.WriteFile and encountered genuine
sharing refusal; the next native fixture omitted read/query access and failed its
observation. Both failed attempt logs remain. Final fixture uses native READ|WRITE
and READ|WRITE|DELETE sharing; the complete Windows selector above passes. Neither
fixture failure is presented as a product defect or hidden as passing evidence.

Compact local evidence root:
`C:/Users/hanse/.codex/tmp/persistence-protocol-review-5e8aaeb/`.
The independent source, overlay mapping, native binary and UTF-8 logs remain there;
no large duplicate archive was created. Exact final SHA256:

| File | SHA256 |
|---|---|
| review_linux_test.go | 0d59895df4ed23c938698cabd52c4b820fe578097da4cce01cf03d15b841d8b6 |
| review_protocol_test.go | 68f139486ae948462037fa2fb372d82df06db912636f431c71394e3f7c93b17a |
| review_windows_test.go | 9eeeb87fd5bcde45f3c8502341e3188f0c627dc7420ae5ec2e4d3ec134de7bcc |
| review-linux.test | c3f56dacecd2787468ead9793a5a1dab3b3931dfcb284dfbe2822e308edffeeb |
| linux-results.log | 113e8b460aa42c028e6bd8611fe039e22ed0a14d3b42ebc500495e189c4ea000 |
| windows-results.log | f9ba1de544ab6e4704b41f2fef316369f0e1c2ec7dfd93e7f2006cb0586e11b5 |

## Limits and next permitted action

This is one early defect-discovery audit of already implemented behavior, not a
complete candidate review, layer/Slice acceptance, native platform certification,
or integration gate. No destructive replay/cleanup was found in the inspected
protocol; journal snapshots preserve stable IDs and prior verified facts in the
executed bounded torn-tail controls. These positives do not prove every crash stage.

Known TODOs were not reported as new findings: actual original Expected-ancestor
validation, remaining partial/crash artifact stages, payload late-edit semantics,
full fault/multiprocess/resource matrix, no-birth own-effect/restart proof and
ordinary FreeBSD system-EA EPERM profile. Master reports a later author checkpoint
82df1d2568ab48a4ac6a542d41a402b287391790 for ancestor/current-on-exit work; that SHA
was not reviewed here. macOS/FreeBSD and other architectures were not executed by
this reviewer. No CI or complete adapter acceptance is claimed.

Master should preserve this report, coordinate worker corrections to M363-PR01..03,
and carry the independent controls into later exact-candidate verification. Finish
the already tracked protocol/native TODOs before separate full acceptance review.
All full Slices and baseline program findings remain open under existing governance.

## Master preservation and correction status

The three small independent Go overlays and final Windows/Linux logs are preserved
in SDP/Verification/CR-#21/Evidence/M3-Persistence-Protocol--001 (byte-identical to
the local originals before Git text normalization); source evidence uses .go.txt
so it cannot become an unintended module package. No binaries, source checkout, module cache or large archive
was copied. Recreate an overlay JSON mapping nonexistent
internal/persistence/review_{protocol,linux,windows}_test.go paths in the chosen
exact review checkout to those absolute .go.txt evidence files, then use the commands
above. Native platform suffixes select the matching test file.

Author reports corrections atca758efafec7cdbfb5ca6d838b56772bf4cabae9. They are
pending independent bounded confirmation; this report's original three findings
and early-only scope remain unchanged. No full adapter acceptance is implied.

Preservation check: the initial .go evidence suffix caused `go list ./...` to
consider the SDP path a package and reject its # character. Renaming those three
unchanged source artifacts to .go.txt restores package discovery; this is evidence
packaging only, not a product or review-test correction.
