# M3 Persistence Windows artifact identity implementation review

Disposition: **ACCEPT the bounded P3j identity correction. No implementation findings in this scope.** This is not acceptance of the complete #63 adapter, a Slice, integration, or release.

Authority: #63 under #21; Sprint-004-v04 / I-03 / M3; Master's decision `69c4db248d0cf87fb318be54d770dc62be7c3159`, ledger 75, and accepted [M3-Identity-Review--001](M3-Identity-Review--001.md). Fresh independent reviewer `m3_persistence_identity_code_review`, 2026-09-07 local. Reviewed clean `codereview-21/layer-persistence` at **`bfdc79c2289274636f66574a3c86f467bb993615`**, against `c515d31e0a1c241622104291c49e3335e3f5927e`. Product implementation is `0058949bf0f9bdc81021e53e8ef4f9f1e6713968`; bfdc79c adds tests/report. Worktree: `C:/Users/hanse/GIT/gh-tree-wt/persistence-implementation`.

Read AGENTS/developmentInstructions, full current #63/#21 bodies/comments, frozen Application--Persistence and BoundaryTypes/BCFreeze, Storage/native feasibility and the scoped accepted design/review clauses, the accepted identity proposal, prior protocol review/correction evidence, actual changed source/tests, acquisition/publication/manifest/journal context, README and author report. Source and execution below are review evidence; author summaries are not. The blocked Git foundation review was neither retried, substituted, nor accessed.

## Source assessment

- `preparation_windows.go:91` creates only through exclusive FILE_CREATE with FILE_NON_DIRECTORY_FILE, then performs OS CREATE_OR_GET before any artifact bytes, flush, hardlink or publication. Failure closes the handle, reports the abandoned created name and refuses publication. Production CREATE_OR_GET has this single call site; originals, directories, permanent locks and reopened/hardlinked entries do not take it.
- `artifact_identity_windows.go:22/45/57/72` retains the full volume/128-bit FileIdInfo and all 64 returned ObjectID bytes, requires exactly 64 output bytes and nonzero first 16 bytes, and preserves native errors. The sole initial-original fallback is the exact native ERROR_FILE_NOT_FOUND value. Recorded ObjectID profiles use GET only and never fall back. Recorded birth profiles compare their exact birth tuple even on ID-bearing objects. `winObserve`'s layout, FileID/reparse validation and directory/lock birth semantics remain intact.
- `commit.go:328/430/503/550` captures the original before preparation, carries that identity into retention/final checks, and validates the publisher and both payload names against the persisted identity. It does not compare two freshly selected profiles after a change. Byte/security checks, Expected/whole-version binding, native class 65/class 11 publication/retention and truthful outcome/error/cancel behavior remain separate and unchanged.
- `manifest.go:269` validates the manifest self and each artifact using that artifact's recorded profile. The four-field disk identity shape is unchanged. Journal successor checks reject an identity upgrade; flat/journal readers preserve old shapes/hashes and historical RecoveryIDs/records. Late payload bytes change the observation SourceVersion while preserving the shared historical record. No timestamp repair or postpublication identity rewrite exists.
- The handle test counts all current-process native File-type handles, including directory guards and locks. It disables automatic GC, requires exact baseline restoration after every one of 12 public cycles, and detects a deliberate duplicate. This measures the request-owned native file resources the regression concerns. It does not establish complete process/thread/event-resource or every failure-path cleanup coverage.

## Independently executed evidence

Native Windows NT 10.0.26200 / local NTFS; direct Go 1.25.0 at `C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe`, CGO_ENABLED=0. All fixtures are reviewer/test-owned. No preserved denied-ACL fixture was touched. No product/frozen file was edited.

| Exact frozen-source control | Result |
|---|---|
| amd64 `go test ./internal/persistence -run 'TestWindowsArtifact\|TestWindowsPublicCommitAndLoadsReleaseAllRequestHandles\|TestManifest\|TestCommitFaults\|TestCommitStale' -count=1 -v` | PASS (3.688s). Includes 12-cycle resource regression; four fresh-store plus four fresh-process commits per presence case, all 30/30 absent-first and 32/32 present-first recovery records, with four observed tunneled payload births each; child kill/join before outcome; all-role same-byte substitutions; malformed/full-tuple/native-error controls; unchanged originals; old flat/journal profiles; journal/fault/stale controls. |
| New independent test-only overlay, amd64 `-run '^TestReview'` | PASS (0.819s). Native ID deletion after original capture before retention, or at final checks on original/payload/publication, yields valid NotCommitted, no publication, VerifiedNoTargetChange and unchanged target bytes/security. Eight old-birth drift cases cover every artifact role in flat/journal records, retaining independent records and exact disk bytes. |
| Same overlay plus identity/original/late-payload/outcome controls, native 386/WOW64 | PASS (0.918s). Confirms actual 32-bit execution, precise ID-loss refusal, historical/current record separation and unchanged tagged/untagged originals. |
| Independent directory/lock counter negative controls, amd64 and 386 | PASS. Acquiring the native chain adds exactly its guard count; lock adds 1; duplicating directory and lock handles each adds 1 and explicit closure restores baseline with GC disabled. The revised counter detects these resource leaks. |
| Existing independent 91acac5 confirmation overlays at bfdc79c, native WSL Linux/ext4 as UID/GID65534 | PASS. All 8 staging substitutions at final-check/before-publication refuse with unchanged target/absence and sentinel; 04750 survives a known commit; valid current and 3 recovery IDs survive an unsupported entry; known effect survives delivery error/cancel; Proposed/Current remain distinct. Existing all-family recovery and set-ID controls also pass. M363-PR01/02/03 do not regress in their reviewed cases. |

Windows overlays add tests through `-overlay` without replacing product source. The Linux run crosscompiled the unchanged product plus the already preserved independent `.go.txt` controls, then executed under `runuser -u nobody` in owned `/tmp/gh-tree-identity-code-review.5WyxS0`; `findmnt` confirmed ext4. An initial inline WSL command failed before tests because quoting lost its temporary-path variable. Its accidentally staged `/review.test` binary was hash-matched to this run's binary and removed; the corrected file-backed script executed only the owned /tmp fixture. That failed command log remains separate and is not passing evidence.

## CI and limits

Independently reopened [CI34066675310](https://github.com/Hans-Einar/gh-tree/actions/runs/34066675310) at bfdc79c: **18 successful jobs, 1 native FreeBSD failure, 1 expected Runtime-helper skip**. Windows amd64/native ARM64, Linux/macOS, Linux race and all 12 architecture/build jobs succeeded. The [ARM64 job](https://github.com/Hans-Einar/gh-tree/actions/runs/34066675310/job/101576538227) log confirms exact source, runner/host ARM64, Go 1.25.0 windows/arm64 with GOHOSTARCH=GOARCH=arm64, actual package tests (Persistence 8.233s), vet and build. This is native execution, distinct from cross-compilation. The aggregate run is failing; no complete CI/adapter gate is declared passed.

Compact local evidence: `C:/Users/hanse/.codex/tmp/gh-tree-m3-identity-code-review-20260907`. New source uses `.go.txt`; no duplicate checkout/archive is added. SHA256:

| File | SHA256 |
|---|---|
| reviewer_identity_windows_test.go.txt | 0FF9E1F548B8C7ABD3A44A5010087C11705DD91FBB4A1EA30047888748714E68 |
| independent-amd64.log | 1805E8DCC410A22087C34D9F3CFC29C6909AEE1C6246925B30DC2493E7697703 |
| independent-386.log | A35599AFC5FCB4597551D271CC4AB57A28A3752020C82504261ABBDB29C1B91D |
| independent-linux.log | F232802D3516780BD211D411CE4A6F2733AAD1E13EFEEEFBF376F09E68F19AEB |
| ci-windows-arm64.log | DE45295D9F55ED26027ACE93FEAA7F7D3BFEB1C01A7EE749AEF07F2F486179E0 |

The source was clean at bfdc79c through verification; `git diff --check` passed. The Persistence tree is `ddcb2a69aec9dd98052d15407aef0985a01ee3de`. This report is the only repository change made by this reviewer and is committed/pushed separately with `[skip ci]`.

Exact next permitted action: Master may record acceptance of this bounded private correction and preserve its evidence. **Full #63 remains open** for ordinary-user FreeBSD metadata/publication, no-birth association, the Unix source-name check/syscall interval, and complete native preparation/fault/crash/resource gates. No reduced read-only/profile substitute, broader adapter/Slice/finding closure, product integration or release is authorized. Git-first serial integration remains held behind its separate unresolved review and prior gates.
