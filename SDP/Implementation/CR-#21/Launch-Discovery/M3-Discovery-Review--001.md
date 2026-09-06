# M3 Launch Discovery independent implementation review (#64)

Verdict: **CHANGES_REQUIRED** — one MEDIUM finding, M364-M01.
Reviewed technical SHA: `2a18fdd5b884568d832afeb7fab54bb4d4b50682`.
Reviewed final/report SHA: `8e2d2d1218ef6dcb70cbe06d55678e0420a20b36`.
Branch: `codereview-21/layer-launchdiscovery`.
Reviewer: fresh independent `m3_discovery_review`, separate from implementation.
Date: 2026-09-07. Parent #21; continuation authority is its current comment and
canonical `30a78d632299cd9ab4eb26e81c029c32b10dcb52`. Serial integration remains held.

## Scope and authority

Read repository AGENTS/developmentInstructions, current #64/#21 authority and
comments, #36's accepted review, relevant accepted REFDES/API/Storage/CwdAcquisition,
MigrationMap/Slices/Verification and frozen Application--LaunchDiscovery 1.0.0,
Runtime/Persistence and BoundaryTypes/BCFreeze clauses. Inspected every actual
`internal/launchdiscovery` source/test/README, its author report and the accepted
Domain/API constructors and binding checks used by both ports. Author summaries
are not the basis of this verdict. No other adapter review or full Slice acceptance
is attempted; the unrelated held review was not retrieved or retried.

Initial worktree was clean at the final SHA. `git diff --stat 2a18fdd 8e2d2d1`
shows exactly the author report changed; the executable source/tests are identical.
The contribution adds only its owned package and report; shared prerequisites
remain separately governed. This review changes only this independent report.

## Finding M364-M01 — inline Make comments erase valid simple targets

**MEDIUM**, `internal/launchdiscovery/providers.go:221` (`parseMake`). The parser
checks the entire physical line for `$`, `%` or backslash before stripping its
inline `#` comment at line225. Consequently these ordinary supported rules produce
zero members and `make-profile-limitation`:

```make
all: dep # costs $5
all: dep # 100% done
```

The same rule with `# ordinary comment` produces the expected `all` member.
Comment prose is being interpreted as dynamic target syntax. This hides a valid
launch from Discover and makes an otherwise valid saved `all` intent unavailable.
It violates the accepted simple textual target/profile behavior; the generic
partial diagnostic does not preserve this independently valid fact. Existing
`TestSimpleMakeGrammar` does not cover inline comments containing these characters.

Independent overlay test `TestIndependentSimpleCommentOperands` at the exact
technical source returned exit1 with both failures: `members=[]`,
`notices=[make-profile-limitation]`, `error=<nil>`. The ordinary-comment positive
control passed. A separate `$(TARGET): dep` plus `all: dep` control retained only
`all` and still diagnosed the actual dynamic target. No Make/provider process ran.

Required correction: classify ordinary inline comment text before interpreting
expansion/pattern syntax in the rule; preserve legitimate simple targets and
independent project results. Distinguish actual escaped comment markers and
backslash-newline continuations from ordinary comment prose. Unsupported real
continuation/escape/dynamic constructs must stay explicitly limited/refused;
naively stripping every `#` or constructing a Make interpreter is not required.
Add parser and Discover/saved-resolution regression controls for `$`/`%` comments,
ordinary comments, genuine dynamic rules and the applicable continuation/escape
profile. Freeze a corrected source for independent re-review.

## Independent source assessment and evidence

Apart from M364-M01, no additional actionable defect was established in this pass.
Both ports remain passive: source has no provider/process invocation, cwd change,
project write, run.json read or cross-adapter call. Registration is copied/unique;
Domain owns length-framed IDs. API admission binds worktree/storage/closed selection
and copies nested values. Source reobservation binds manifest identity/content,
root/project identities, relevant regular lock observations, manager and override;
saved aliases and complete ordered Make member lists remain distinct authorities.

Strict JSON checks duplicate keys, UTF-8/surrogates and depth; invalid scripts remain
unavailable. Npm keeps colon/whitespace in one literal argument, uses only colocated
regular pnpm/yarn precedence and diagnoses conflicts/redirects. Make binds its
chosen precedence lookup with `-f`, preserves selected order and rejects unsafe
operands. Scan/read/candidate/directory/line limits, exclusions, cancellation and
partial independent projects have concrete controls. Native Windows relative
no-reparse observation and Unix retained no-follow descriptors are interval facts;
no handle crosses the port or substitutes for Runtime's later actual-cwd acquisition.

Commands/evidence actually inspected or run:

- Supplied Go1.25.0 Windows/amd64 toolchain: `go test ./internal/launchdiscovery
  -count=1 -v` PASS, all17 applicable test functions. Actual junction/conversion
  and root replacement controls passed; local symlink creation explicitly lacked
  privilege, so that path is not claimed as native Windows symbolic-link proof.
- Independent overlay `go test -overlay=<controls-overlay.json>
  ./internal/launchdiscovery -run TestIndependentReplacementAndExactSavedControls
  -count=1 -v` PASS. Fresh successful pnpm/colon resolution; identical-content
  manifest object replacement and removed manager lock each refused; two saved
  aliases retained one stable underlying ID and the exact selected spaced alias.
- Independent comment overlay command uses `-overlay=<overlay.json>` and
  `-run TestIndependentSimpleCommentOperands -count=1 -v`; exit1 as above.
  Overlay source files are `.go.txt` outside the repository; actual source unchanged.
- [CI34059784449](https://github.com/Hans-Einar/gh-tree/actions/runs/34059784449)
  independently queried at exact `2a18fdd`: COMPLETED/SUCCESS,19 successful jobs,
  solely expected pre-Runtime helper SKIP. Actual Windows ARM64/macOS test logs
  contain this package's PASS; Linux race log records package PASS (9.143s).
  Native FreeBSD log records all18 Discovery functions PASS, including unprivileged
  permission/redirected-lock, moved-object and supplied-change-profile controls.
  Twelve build/import jobs are distinct compilation evidence, not native execution.
- Author local `2a18fdd-linux.log` and `2a18fdd-windows-386.log` bytes were inspected;
  their SHA256 matches the author report. These remain supplemental WSL/Windows386
  captures, not this reviewer's reruns. Earlier failed macOS/FreeBSD runs are not
  passing evidence. Final source includes the EMLINK classification correction,
  and final native FreeBSD controls prove the formerly failing path.
- `git diff --check` PASS before this report. No broad matrix rerun was needed.

Reproducible local overlays: `C:/Users/hanse/.codex/tmp/gh-tree-discovery-review-001`.
`providers_test.go.txt` SHA256
`6D10A0794F2D297440BE99EA6BF993BA311302137B271299186C1B7DBAC6F7EE`;
`adverse_test.go.txt` SHA256
`EC061321001D05AEE50FE2B3A42B63ADEB8CEAB716005010B6C7F8F6C8E8DE2A`.
They append the named independent tests to the corresponding unchanged source test
files. Their essential fixtures/assertions/results are recorded above; no duplicated
large archive or additional repository package was created.

## Limits and next permitted action

The author report's CI IN PROGRESS statement is historical and superseded by the
actual terminal CI result above. Green CI does not resolve M364-M01. Return the
bounded parser/test correction to a worker, then independently re-review its exact
source and appropriate CI/evidence. Master alone may accept and integrate in the
Git -> GitHub -> Persistence -> Discovery -> Runtime order when preceding gates permit.

This review proves neither actual Runtime argv/cwd/batch execution nor Application/
State/View saved/default/storage roundtrip. Those integration/native/E2E gates,
all full Slices and baseline program findings remain open. Product source was
preserved unchanged; only this report is committed/pushed, with no integration,
release, user-project execution or user-configuration mutation.
## Bounded correction re-review — 2026-09-07

Current verdict: **ACCEPT — M364-M01 RESOLVED**, no remaining review finding.
This supersedes the initial technical verdict above without erasing its evidence.
Corrected technical SHA: `20bd8acf3576ca19f08b68205289f25cfe9dd3d9`.
Final author/report SHA: `67156305b23d557aad370ad62d6eb0026d298b02`.

Independently inspected the complete correction: only `providers.go` and new
`make_comments_test.go` changed in product source/tests. The technical-to-final
delta is author-report-only. Native helpers, Discover/Resolve, dependencies and
frozen contracts are unchanged, so the initial native/boundary assessment carries
forward without a redundant full review.

The parser now removes ordinary comment prose before expansion/pattern checks;
escaped markers remain outside the simple profile. Continuation state skips the
unsupported logical line and its tails, preserving independent simple rules.
This respects the GNU distinction between ordinary comment text and an unescaped
trailing backslash that continues it; see [GNU Makefile contents](https://www.gnu.org/s/make/manual/html_node/Makefile-Contents.html).
No native Make/provider execution or full Make interpreter is introduced.

Actual independent rerun on the supplied Go1.25.0 Windows/amd64 toolchain:
`go test -overlay=<original overlay.json> ./internal/launchdiscovery
-run 'Test(IndependentSimpleCommentOperands|MakeCommentClassification|MakeCommentsDiscoverAndSavedResolve)'
-count=1 -v` PASS, exit0. The original independent overlay hash is unchanged from
above; both formerly failing ordinary `$`/`%` comment controls now pass, as do its
ordinary-comment positive and actual-dynamic-target adverse control. The two new
15-case tests each pass: ordinary/dollar/percent/backslash prose, paired/comment
backslashes, escaped markers, actual dynamic/pattern/prerequisite syntax, rule/
comment/recipe continuations and chained CRLF tails. EOF comment-backslash control
also passes. Actual Discover/saved tests preserve an independent npm project and
safe Make rule, resolve ordinary-comment `all` with exact `-f GNUmakefile all`, and
refuse previously valid saved intent after unsupported source changes.

[CI34065291612](https://github.com/Hans-Einar/gh-tree/actions/runs/34065291612)
was independently queried at exact `20bd8ac`: COMPLETED/SUCCESS,19 successful jobs
(native Windows amd64/ARM64, Linux/macOS/FreeBSD, race, inventory and twelve builds)
and solely expected pre-Runtime helper SKIP. `git diff --check` PASS; source stayed
unchanged throughout re-review. Only this appended report is committed/pushed.

Master may now assess acceptance from the exact source/report/CI and this independent
review. Serial integration still requires Git -> GitHub -> Persistence prerequisites;
no held unrelated review is replaced. Runtime execution, Application/default/storage
roundtrip, full Slices, baseline finding closure and release remain later gates.