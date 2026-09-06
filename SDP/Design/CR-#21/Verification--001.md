# Verification--001 — v0.4 acceptance contract

State: ACCEPTED DESIGN (effective after PR #54 merge), normative design appendix under #52. No check below is claimed run
against v0.4. Accepted review and feasibility evidence applies only to its recorded
source/probe, not to future implementation. Test names may be chosen by workers;
the behavior, platforms, failure injection and proof obligations are mandatory.

## Evidence and gates

Each verification record in SDP/Verification/CR-#21 names contract version,
source and integration SHA, clean/dirty status, toolchain, OS/architecture,
command or test selector, fixture, expected/actual result, exit status, log/artifact
hash and CI URL/job/attempt. Re-run relevant checks after any affected change;
final full verification uses one frozen integrated product HEAD. CI checkout SHA
and merge-ref versus source-ref evidence must be distinguished. Skipped tests,
compile-only probes, mocks, cross-builds and native execution are separate facts.

Layer review inspects actual source/tests and BC conformance at frozen Worker HEAD.
Author summaries never satisfy independent review. Master records exact accepted
source and resulting serial integration SHAs. V-E2E-01..12 run through the complete
new stack in a test-only Composition harness before cmd cutover, then through
actual host/CLI transport after cutover. GitHub may use transport fixtures for
controlled mutation tests; Git, storage, discovery and Runtime must have real
temporary infrastructure. No test mutates the user's working repositories/config.

All V-DOM, APP, GIT, RUN, STATE, VIEW, PER, LCH, GH, COMP, E2E-01..12 and the
prepublication portions of E2E-13/REL must pass before final product PR. Final
main SHA is separately tested before tag/release. Delivered install/upgrade and
published-asset checks follow staged artifact approval, not a circular prerequisite
for creating the product PR. If publication cannot be safely completed, report
the exact release-ready external action; never label unpublished delivery tested.

## Platforms and tooling

Native minimum: Windows amd64 and ARM64 (Job + ConPTY + actual-cwd startup), Linux amd64 (session/PTY), macOS on
an available hosted native architecture, and FreeBSD amd64 for its selected census/
signal mechanism. Use native Windows 386/WOW64 for handle/ABI contract tests where
the implementation has architecture-sensitive native structures. Exercise native
ARM64 and x86/x64 emulation broker routing on the documented windows-11-arm runner;
verify the actual runner/image/profile at execution. All twelve
release targets cross-build from the exact source. No cross-build claims native
Runtime parity. Additional architecture-specific syscall layouts require compile-
time size/offset assertions plus native evidence where layout differs materially.

GitHub Actions retains native Windows/Linux/macOS full `go test ./...`, `go vet
./...` and build. Linux runs `go test -race ./...`; additional supported native
race jobs may strengthen evidence. FreeBSD execution may use a reviewed pinned
VM action on a hosted runner; log guest version and prove tests actually execute
inside the guest. Native evidence cannot be replaced by successful packaging.
ConPTY close tests include a supported Windows environment exercising blocking
close behavior or an explicitly injected native-close blocking path with an
independently tested real ConPTY lifecycle. Record this distinction.

Tooling checks: gofmt produces no diff; full test/vet/build, bounded-race suite,
dependency/API/package checks for every GOOS/GOARCH selection, and no generated
or untracked dependency on developer-local paths. Release builds use CGO_ENABLED=0
as existing packaging requires. Record actual Go, Git, gh and dependency versions.

| Target | Exact asset name |
|---|---|
| darwin/amd64 | darwin-amd64 |
| darwin/arm64 | darwin-arm64 |
| freebsd/386 | freebsd-386 |
| freebsd/amd64 | freebsd-amd64 |
| freebsd/arm64 | freebsd-arm64 |
| linux/386 | linux-386 |
| linux/amd64 | linux-amd64 |
| linux/arm | linux-arm |
| linux/arm64 | linux-arm64 |
| windows/386 | windows-386.exe |
| windows/amd64 | windows-amd64.exe |
| windows/arm64 | windows-arm64.exe |

## Domain

| ID | Required proof |
|---|---|
| V-DOM-01 | Equality/zero/invalid constructors; scoped repo/worktree/local-vs-remote branch/PR/stash/launch/session identities; linked worktrees versus clones; mutable display/path/position excluded where required. |
| V-DOM-02 | Full SHA-1/SHA-256 OIDs, canonical case, nonhex/zero/whitespace/abbreviation rejection, commit Revision versus latest ref; exact target scope consistency; Attached/Detached/Unborn and absent revision refusal. |
| V-DOM-03 | Pure imports and public types, no DTO/JSON/path/URL/clock/process behavior; value copies and SessionID allocation exhaustion contract in Runtime. |

## Application

| ID | Required proof |
|---|---|
| V-APP-01 | Unique monotonic OperationIDs; acceptance before progress/terminal; exactly one terminal under success, errors, panic, cancel, supersede and shutdown. Shared operation/session reservation budget refuses before acceptance. Fill App queue, naturally exit multiple sessions and detach host: Runtime ACK occurs only after reserved downstream transfer, normalized cleanup remains reliable, no premature eviction/unbounded queue, and shutdown drains without deadlock. |
| V-APP-02 | Query slot/generation replacement and mutation lifetime independence; canceled-before-start, canceled-during-native-call and applied-then-refresh-failed have truthful distinct effects; immutable nested snapshots under race detector. |
| V-APP-03 | Confirmation token exact target/plan/allowed choice binding, single consumption, wrong/expired token, stale HEAD/index/content/occupancy/ref/storage; no locks during user delay; dirty stash/deploy preserves original exact target and returned created stash after later failure. |
| V-APP-04 | Active preference commit-before-publish; failed write leaves old context; committed durability uncertainty and indeterminate reconciliation; concurrent activations ordered; saved/current/fallback with missing/partial inventory; no terminal cwd retarget after activation. |
| V-APP-05 | Canonical local/upstream/GitHub reconciliation with explicit host/repo/head/base scopes, forks, multiple PRs, detached/unborn, none/resolved/gone/unresolved, stale cache and incomplete lists; optional ahead/behind never zero-filled when unknown. Graph joins preserve commit messages. |
| V-APP-06 | Narrow port fakes prove use-case sequencing, normalized adapter errors and facet effects, runtime operation-vs-session event distinction, default/saved launch and configured worktree targets; no concrete DTO import or facade delegation to legacy. |

## Git — real temporary repositories, SHA-1 and SHA-256 where supported

Use local bare remotes and at least two linked worktrees/clones. Test literal
paths containing spaces, colon, dash, non-ASCII, rename and newline where the
platform permits. Disabling Git features in a test fixture must be recorded;
it does not prove behavior with those features enabled.

| ID | Required proof |
|---|---|
| V-GIT-01 | Physically scoped common-dir/worktree identity, primary/current/linked worktrees, clean/staged/unstaged/untracked/conflicted states, Detached/Unborn, malformed/locked/missing worktree records and branch occupancy. |
| V-GIT-02 | Exact PR/branch/commit resolve/fetch/checkout; mutable ref advances between selection and fetch; stale local branch differs from selected remote revision; full expected-OID mismatch refuses, no fallback to current tip. Preserve departing detached/local history, ignored/untracked collisions and occupied/primary/current restrictions. |
| V-GIT-03 | Upstream none/resolved/gone/unresolved/not-applicable, equal/ahead/behind/diverged endpoints, scoped freshness generations across linked worktrees; remote-tracking refs remain cached facts. Fetch partial/error/prune and remote replacement do not fake freshness. |
| V-GIT-04 | Stash identity survives prepend/reflog shift, duplicate OID ambiguity, managed/unmanaged legacy metadata, SHA-256; exact apply restores index, clean-target refusal/conflict state; pop apply followed by drop failure yields AppliedStashRetained and is never automatically replayed. |
| V-GIT-05 | Exact files-backend stash deletion under competing store/drop, packed refs, last-entry A-B-A reinsertion, lock order, survivor metadata/chain preservation, malformed/symbolic/reftable refusal. Failure/crash injection at each journal publication stage and restart recovery; never age-delete another writer's lock. Recovery refs survive object cleanup. |
| V-GIT-06 | Confirmed restore from index with HEAD=A/index=B/worktree=C; same-size/mtime-preserved edits, captured-object mismatch, recreated destination, held-open writer after capture, sharing denial, conversion/attributes/CRLF, path redirects and index replacement. Every original/new object remains accessible; no untracked deletion. |
| V-GIT-07 | All checkout/fast-forward/stash worktree publication uses the accepted no-loss protocol. Inject edits in precheck/write gaps, reference-transaction hooks, path replacement after capture and failures at every multi-path/index/ref stage. Prove retained actual originals, no-replace publication, no silent reset/clean overwrite, truthful partial/indeterminate recovery and no destructive retry. Quiescent-writer assumptions alone fail this gate. |
| V-GIT-08 | Stage/unstage literal paths and StageAll; index lock/conflict and changing files; commit existing index versus explicit stage-all workflow, literal message, failed hooks/signing, new exact Revision and partial staged result. Fast-forward pull only, dirty/detached/gone/diverged refusal; pinned nonforce push and separate set-upstream outcome. |
| V-GIT-09 | Canceled before acquisition/during fetch/worktree/index/ref/remote mutation; child reaping and bounded reconciliation independent of canceled request; facet effects distinguish known applied, unchanged, partial, indeterminate and auxiliary recovery refs. Unsupported features refuse before unsafe publication. |
| V-GIT-10 | Exact commit/parent/root/history, accumulated graph refs/parents, PR merge-base range from exact fetched base/head and recorded merge-base, explicit missing/ambiguous base diagnostics; StashPatch by exact OID/parent choice after reflog shifts, staged vs unstaged diff, binary/rename/large truncated patch and invalid text. Local-only facts never require GitHub; no stderr prose as sole semantic parser. |

## Runtime — native helpers, resource and descendant evidence

Fixtures record root/child/grandchild identities and actual handles/session/group
membership. A killed root alone is never sufficient proof. Time bounds use a
controlled clock or deliberately bounded real waits and retain diagnostic output.

| ID | Required proof |
|---|---|
| V-RUN-01 | Unique IDs across launch/terminal/concurrent starts, exhaustion refusal, deep copied specs, admission bound; failure injection after every pipe/PTY/process/job/registry acquisition proves no orphan or leaked handle/goroutine. Cancellation during start establishes/cleans ownership. |
| V-RUN-02 | Running/exit/stopping/cleanup facts, Stop idempotency and concurrent natural exit; exactly one cleaned lifecycle result; root exits before descendants; graceful deadline then force; successful Stop proves all required owned descendants and I/O are gone. Restart refuses failed cleanup and uses new ID only after old barrier. |
| V-RUN-03 | Windows suspended root assigned to nonbreakaway Job before resume; failure never resumes uncontained child; child/grandchild/tree termination, Job active count zero and root handle wait; nonPTY hidden launch, no broadcast Ctrl signal, ConPTY ETX delivery separate from Stop. |
| V-RUN-04 | Unix dedicated SID, normal shell foreground/background PGIDs and PTY job control; identity-safe cleanup under disappearing/reused groups, stopped jobs and root-exit races; native Linux/macOS/FreeBSD mechanism evidence. Explicit new-session escape scope is documented, observed escape is residual failure. No numeric precheck-to-kill assumption can pass the identity gate. |
| V-RUN-05 | Real PTY/ConPTY input/resize/close; blocked readers/writers and terminal close join, old blocking-close behavior; invalid/late resize/input after stopping; latest accepted geometry at restart; resource counts before/after repeated cycles. |
| V-RUN-06 | 256KiB ring, byte/stream/offset order, overflow gap and truncation, coalesced hints recover via ReadOutput; slow/absent UI cannot block draining; copied bounded input; cleaned history evicts only cleaned records. |
| V-RUN-07 | Concurrent Stop/Restart/Shutdown and consumer detach, complete aggregate residuals, bounded overall shutdown, no timer/subscription/goroutine leak; cleanup timeout never reported Cleaned and retained resource prevents replacement/eviction. |
| V-RUN-08 | Full WindowsBroker/CwdAcquisition contract: actual data-read nonempty anchors versus metadata-only failure, in-place FSCTL reparse before/after CreateProcess and before Resume, target-specific breakpoint/PEB/FileIdInfo proof, pending-event detach and no user code/anchor/debugger before approved startup. Test386-to-native64 via embedded broker, native64-to-WOW64 and native ARM64/emulation profiles, static DLL/TLS/debug-heap and immediate-chdir compatibility, helper extraction tamper/identity/permissions, all partial failures, framed control/output ownership, inner quiescence -> Release -> broker exit -> outer Job0. No outer-membership1 circular wait or private handle-injection fallback. |

## State and View

| ID | Required proof |
|---|---|
| V-STATE-01 | One deterministic mode/focus/modal authority; PR/branch namespace navigation, pane/subpane cycle, search/clear/back, Alt mnemonics/Alt-number, console selection/input ownership and deterministic semantic fallback. |
| V-STATE-02 | Async Start then Alt/Tab navigation, console A restart while B selected, worktree change during status/stash/launch query, stale branch/diff/page result and older modal completion; OperationID plus slot generation and global intent prevent focus/selection/modal theft. |
| V-STATE-03 | Confirmation only through matching token/operation; cancel/quit/root failure and late events cannot reopen disposed state; session sequences accepted independently of current tab; active context is App projection. |
| V-STATE-04 | One graph reducer across both entry modes, accumulated pages/selection, launch ordered Make selection/edit forms, key applicability and declarative effects; no adapters, timers, clocks or hidden workflow in reducer. |
| V-VIEW-01 | Known widths/heights including 1x1, narrow, short, standard and wide; pure layout/measurement agreement; focused compact pane access, scrollable modal with exact target/choices inspectable at adequate size, bounded resize/read-only notice with no hidden approval when too small, help from applicable keymap. Cell/grapheme widths, combining/wide text, invalid UTF-8 and control injection. |
| V-VIEW-02 | Graph DAG multiple roots/merges/page continuation, full exact detail, diff/source/truncation labels, semantic unknown/upstream states, console CR/backspace/control rendering and supplied activity frame; no rewritten commit facts or process-existence working inference. |
| V-VIEW-03 | Immutable viewmodel/deep copy, rendering changes no State, modal overlays no interaction authority, no backend imports/callbacks/I/O/clock; host resize uses actual content rectangle with SessionID/viewport correlation. |

## Persistence, Discovery and GitHub

| ID | Required proof |
|---|---|
| V-PER-01 | Existing config/state/run.json fixtures: missing vs explicit-empty prefixes, aliases/default and command overrides, configured targets, last namespace/worktree, unknown fields, corrupt and forward versions, lossless backup/migration. |
| V-PER-02 | Whole-document expected version, permanent LockFileEx/flock lock lifetime including crash/PID reuse and concurrent cooperating writers; selected handle-relative Windows class65 replacement/class11 retention and Unix Renameat/no-replace publication, complete reader visibility, exclusive payload and supported permission/ACL/label profiles. Inject short-write/metadata/flush/commit/close/crash stages, preserve exact commit-point/durability/indeterminate outcomes. Explicitly test detected external-editor conflict, retained observed originals and the documented unsupported check/replace race; no universal CAS or corrupt-to-default claim. |
| V-PER-03 | Host/owner/name key normalization, case collision and ambiguous legacy maps, common-dir vs clones/relocation, linked worktrees, unavailable saved path retained; platform-native Windows/macOS/XDG paths and explicit overrides without default-path evaluation. |
| V-PER-04 | Project run.json no-follow/no-reparse object acquisition, Windows directory/ancestor pins, Unix pinned-directory movement and substitution, observed drift refusal and explicit object-scope/current-path limit; same-document alias/default commit, concurrent source-version change and cancellation; storage never chooses active target or starts a provider. |
| V-LCH-01 | Passive nested npm/pnpm/yarn/Make discovery, exclusions/limits/incomplete diagnostics, exact `dev:wan` and whitespace bytes, colocated lock precedence/conflict, GNUmakefile/makefile/Makefile and explicit chosen file. No provider execution. |
| V-LCH-02 | Ordered same-project/source/executable Make targets, option/assignment grammar refusal, repository-relative cwd, symlink/reparse/escape/source-replacement race, immutable identity vs source version, cancellation and limit diagnostics. Replace cwd/ancestor after validation but before actual Start: Unix acquired descriptor/inherited cwd cannot select replacement; Windows effective directory pins block substitution. Test observed relocation refusal and fixed-object/current-path limit explicitly. |
| V-LCH-03 | Saved/default/first-save alias, executable override, reopen run.json and source/member disappearance, stale selection before Runtime Start, foreign worktree/source refusal; no persistence or Runtime ownership in Discovery. |
| V-GH-01 | Mocked raw API/GraphQL parsing, explicit host/repo scope and exact PR head/base/fork identity, SHA formats, absent/null/deleted head fields, branch mapping, immutable DTO isolation. |
| V-GH-02 | Observation intervals, bounded pagination and More/Unknown, partial per-source errors, rate/auth/not-found/timeout/cancel/malformed data; capped list cannot prove absence, no fake freshness from local fetch. |
| V-GH-03 | Explicit head/base remote create-PR request and expected published revision checks; no implicit local push/fork/repository guessing; created/failed/indeterminate remote effect and safe reconciliation without duplicate retry. |

## Composition and vertical capability proof

| ID | Required proof |
|---|---|
| V-COMP-01 | Flag parsing/help/version before auth/storage/process bootstrap; unknown args/operand diagnostics; explicit --repo/--state/--config, remote-only mode and mismatched local repo capability; lazy platform defaults and dependency failure messages. |
| V-COMP-02 | Partial constructor unwind, root context into Tea/App, actual terminal restoration, one host event bridge/timer and output caches, primary plus cleanup error propagation, nonzero residual exit; main is sole os.Exit owner. |
| V-COMP-03 | Every baseline file/test disposition, no new code importing legacy, exact temporary allowance then empty after cutover; all platform selected imports/public signatures; no relocated facade/wrapper chain or recreated generic process layer. |
| V-COMP-04 | CI trigger proves canonical integration/layer branch source; native test/race/vet/build and twelve packaging targets; shared paths single owner; release bootstrap branch cannot publish and no clobber action hidden behind tag trigger. |

V-E2E-01 through V-E2E-13 are one-to-one with SLC-01 through SLC-13 in
Slices--001. Every retained key/form/capability there must have an executable
scenario identifier, fixture and observed result. Required failure interleavings
include dirty deploy with stale confirmation, exact stash after position shift,
applied mutation plus failed refresh, async launch followed by navigation and
shutdown with active launch and interactive sessions. Headless scenarios drive
State -> App -> real adapters -> event -> State -> View; CLI cases add real host
input/resize/quit. A collection of isolated folder tests is not E2E evidence.

## Release proof

| ID | Required proof |
|---|---|
| V-REL-01 | Before product PR, build/stage all twelve exact names from verified integration SHA; nonempty executable format/architecture, version/module/source metadata, hashes and complete manifest. Rebuild embedded Runtime helper source closure with canonical pinned toolchain; require exact image/compression/manifest hashes and native machines, normal clean-checkout go build without generation, and no helper download/compiler/extra asset at install or run. Fail on omitted/extra/unsupported targets; no publication side effects. |
| V-REL-02 | Fresh exact product PR readiness review and green current HEAD, zero blocking threads/findings, then exact main release commit full CI; version/release notes in that tested commit. Tag v0.4.0 must point exactly there, never fabricated by default-branch release behavior. |
| V-REL-03 | Draft release assets match approved manifest/hash/source/workflow SHA, no existing asset clobber; publish only after inspection. Read back published tag/release/assets and verify downloads correspond to staged bytes. |
| V-REL-04 | Isolated GH_CONFIG_DIR/config/state environment, preserved user's installed extension: install delivered artifact with correct extension name/OS/arch, noninteractive --version/help, controlled upgrade from baseline binary/state and migrated app startup. Record any environment-limited live transport step exactly; never claim the user's installation was upgraded. |

All baseline HIGH findings and both early design HIGH findings require corrected
design/implementation plus independent review and real proof. Lower-severity
findings receive resolved or explicitly accepted documented disposition under
their Issue; no finding closes solely because a file moved or a report merged.
If a mandatory check cannot run, its gate remains incomplete with exact cause.
