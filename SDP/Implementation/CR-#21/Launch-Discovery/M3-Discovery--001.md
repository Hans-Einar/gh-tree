# M3 Launch Discovery contribution (#64)

Status: M364-M01 CORRECTION COMPLETE / FROZEN FOR BOUNDED RE-REVIEW; integration not authorized.
Technical candidate: 20bd8acf3576ca19f08b68205289f25cfe9dd3d9.
Previously reviewed technical candidate: 2a18fdd5b884568d832afeb7fab54bb4d4b50682.
Branch: codereview-21/layer-launchdiscovery.
Worktree: C:/Users/hanse/GIT/gh-tree-wt/discovery-implementation.
Role: implementation Worker, separate fresh Reviewer required.
Scope: internal/launchdiscovery plus this report only, SLC-09/10 and cancellation12.
Authority read: full resumed #21/#64, AGENTS/developmentInstructions, accepted #36
review and REFDES/API/CwdAcquisition/Storage/MigrationMap/Slices/Verification,
frozen Discovery/Runtime/Persistence/BoundaryTypes1.0.0 and actual accepted API/ports.
No frozen contract, legacy product, module, workflow or Master metadata was edited.
Master merged verified prerequisites at4bfcf7f; adapter source was unchanged then.

## Bounded M364-M01 correction (2026-09-07)

Independent review591cb46 returned CHANGES_REQUIRED for inline Make comment prose
being misclassified as dynamic rule syntax. This correction changes only
providers.go and the new make_comments_test.go; the independent review is unchanged.
Ordinary comment text is classified before rule expansion/pattern checks. Escapes
before a comment marker remain unsupported. Actual backslash-newline logical lines
are diagnosed and skipped in full, including comment/recipe continuation tails;
no line joining, tool execution or Make interpreter was introduced.

Two regression tests each exercise15 cases: ordinary/$/%/backslash comment prose,
paired versus actual trailing continuation backslashes, escaped markers, actual
dynamic/pattern rules, rule/comment/recipe continuation and chained CRLF tails.
Parser controls retain an independent safe rule; Discover controls also retain an
independent npm project. Saved resolution succeeds with literal -f GNUmakefile all
for ordinary comments and refuses a previously valid saved target after unsupported
source changes. A final comment backslash without newline is separately covered.

At the correction bytes committed as20bd8ac, supplied Go1.25.0 Windows/amd64:
`go test ./internal/launchdiscovery -count=1`, corresponding `-race`, `go vet`, and
`git diff --check` all PASS (exit0). Product correction was committed/pushed with
source CI enabled. Root watches the exact-source native CI; no current native CI
success is inferred from local tests. The previous candidate's native CI34059784449
is independently verified SUCCESS in the unchanged review report, superseding
the historical pending status below. Exact next action is the same independent
reviewer's bounded confirmation of20bd8ac and current CI, then Master's gated
disposition. This author stops after the report checkpoint; no integration occurs.

## Original candidate implementation and evidence (historical where SHA-specific)

Constructor `launchdiscovery.New(Config{})` supplies both exact immutable ports.
Private provider/native helpers are independently authored, with no concrete
adapter dependency, provider process, cwd change, file write or run.json load.
README documents construction limits, profile/version rules and native guarantees.

| Layer-owned clauses | Actual source / tests |
|---|---|
| Passive operation and dependency boundary | discovery.go/resolve.go; TestPassiveOperationsAndProductionSurface checks owned tree bytes/cwd, deliberately executable-looking sources and forbidden production calls/imports; all12 architecture checker separately passes. |
| Unique providers and stable exact identity | New, Domain.NewLaunchPointID, length-framed source fingerprints; TestDiscoverIdentityPartialAndLiteralResolve covers provider and slash/member collisions, leading spaces, colon argv, generated exclusions and independent malformed project. |
| Strict npm Unicode/duplicates/invalid members | providers.go strictValue/parseNpm; TestStrictManifestAndLiteralMembers covers escaped duplicate keys, nested unknown duplicates, invalid UTF-8/surrogates, malformed shapes and exact members. |
| Manager/override/source drift | project/savedDefinitions; TestManagerObservationsAndDrift, TestSavedOverridesUnknownAndOrderedBinding and TestSameSizeSourceEditAndNewPreferredMakefile cover locks, conflicts, explicit overrides, same-size restored-mtime edits and manifest substitution. |
| Make profile, precedence, order | parseMake, project, Resolve; TestSimpleMakeGrammar and TestMakePrecedenceOrderAndExactVersion cover unsafe operands, profile limitations, selected -f and exact semantic target order. Saved observations retain complete ordered argv. |
| Bounded scan/read/parse/diagnostics | Limits, bounded child heap, observeFile; TestExactDefaultBoundsAndDeterministicPartialScan verifies default constants, reduced breadth, actual depth5/6 and10,001 members; TestManifestAndMakeLineByteCaps tests exact/over4MiB and1MiB. |
| Cancellation, partial facts and copied results | TestCancellationPreservesIndependentRootFacts cancels during observation while preserving root definitions; construction/concurrent-copy tests plus actual Windows race run. No detached goroutine is used. |
| Foreign/stale/saved binding | TestForeignSourcesSavedAliasesAndOrderRefuse covers foreign issuer/worktree, cross-project Make, duplicate alias, exact StorageVersion/alias, unknown provider and saved order/override drift. Discovery does not assert current store freshness/default authority. |
| Native root/project/file/no-link revalidation | native files and openProject/check; native retained identity/byte-cap/profile tests, Windows junction/in-place conversion/root replacement, Unix symlink/permission/moved original/root replacement and supplied change-profile tests. |
| Literal invocation and current cwd DTO | Resolve builds ArgvExecution plus fresh CwdObservation using supplied scope, exact components and observed identity. Tests inspect complete argv/cwd and copied nested values; actual Runtime execution remains a separate gate. |

## Actual verification

All commands use Go1.25.0. Windows executable is the supplied toolchain under
C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin.
Tests create only owned temporary projects. No user configuration/tooling changes.

- At exact2a18fdd: `go test ./internal/launchdiscovery -count=1` PASS on Windows/amd64;
  all12 `GOOS/GOARCH CGO_ENABLED=0 go test -c ./internal/launchdiscovery` PASS.
- Exact2a18fdd Windows/386 binary executes17 test functions natively under WOW64,
  exit0. Exact Linux/amd64 binary executes18 test functions in existing WSL
  openSUSE-Leap-15.5 as actual UID/GID65534, exit0, with TMPDIR
  /tmp/gh-tree-discovery.DY7aAS. `findmnt -T /tmp` identifies ext4 /dev/sdd;
  stat's ext2/ext3 label denotes its filesystem magic. No /mnt fixture proof claimed.
- At937ae7e, before the bounded native-profile correction: Windows/amd64
  `go test -race ./internal/launchdiscovery -count=1 -v`, `go vet` and
  `go run ./internal/composition/architecture -mode staged` all PASS, exit0.
  Architecture selects all12 targets and61 exact legacy allowances. The subsequent
  production change is Unix EMLINK classification; Windows production is unchanged.
- Local Windows symbolic-link creation lacks privilege, explicitly logged; actual
  mount-point junction, already-open directory conversion and root replacement
  controls execute successfully. Unix symbolic-link controls execute successfully.
- Current source CI: https://github.com/Hans-Einar/gh-tree/actions/runs/34059784449
  at2a18fdd is IN PROGRESS at handoff. Native macOS/ARM64/FreeBSD and complete
  configured suite acceptance remain required; Master must inspect actual results.

Bounded local logs are retained in C:/Users/hanse/.codex/tmp/gh-tree-discovery-native:

| Log / source | SHA256 |
|---|---|
| 2a18fdd-linux.log | BEA2081C485915DF850D0995A47D95F23A27EE9B79F80F4A35D4D9E8315D9DB2 |
| 2a18fdd-windows-386.log | AD3D1439779D157EF0E4DE47A60A32F8032AF17CD2352A344F9B1F51FFE16399 |
| final-windows-race.log /937ae7e | 6F8BBE4CB389A2081A46055A00586C00FA2CE530C1D732F3DC249B40861976DE |
| final-architecture.log /937ae7e | E774071C66C98CD68D0DB01DBC8231FDECB7D0A57F525DA00A157DF5A1815395 |

The executable tests and native fixtures are checked in and reproducible; these
local captures do not replace independent source inspection or native CI.

## Corrections, limits and next action

Verified milestones were committed/pushed before subsequent work:427e254 parsers/
observations, cac6470 both ports,3a9e009 native/bounded observations,937ae7e adverse
verification, then2a18fdd native-profile corrections. Prior CI34059223809 and
34059579603 are FAIL, not accepted evidence. Actual earlier macOS evidence exposed
fixture assumptions about case-insensitive Makefile lookup and /var canonical
paths; tests now use actual filesystem lookup and supplied canonical scope.
The downloaded FreeBSD tests.jsonl isolated O_NOFOLLOW's EMLINK result: it had
incorrectly erased valid npm facts when a lock was a symlink. Native classification
now treats EMLINK as redirected, preserving the valid npm project. Current native
CI must prove the correction. Earlier local runner attempts failed before test
execution on quoting/CRLF or unquoted PowerShell test flags; successful reruns
are separately recorded above. No known uncommitted/local-only product work remains.

Observation is passive and bounded, not an immutable-code or post-start sandbox
claim. Runtime independently acquires cwd and executes literal argv/batch transport.
Application/State/View/current-default/full storage roundtrip and V-E2E gates remain
later program work; this contribution closes no full Slice or baseline finding.

Exact next permitted action: fresh independent review of2a18fdd product source and
this report, actual current native CI inspection, corrections/re-review if needed;
Master alone may serially integrate after preceding adapters and record acceptance.
The author stops editing the candidate after this report is pushed.
