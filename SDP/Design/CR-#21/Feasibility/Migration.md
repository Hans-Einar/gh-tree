# Migration ownership and capability feasibility review

State: PROPOSED DESIGN INPUT — not implementation authority or final migration map
Authority: #52 under #21; Sprint-004-v04 / I-02 / G-07
Product baseline: v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b
Design worktree HEAD inspected: 58f1cb9eda941db0941cbb8e04e6a0559a3ca896
Date: 2026-09-06
Role: fresh bounded migration/traceability design reviewer

This review independently inventories actual baseline files, reads the draft
REFDES--001 and accepted architecture/review ownership records, and proposes the
single-owner decomposition and parity gates below. Master owns reconciliation
into MigrationMap.yaml, Slices--001 and FindingDisposition; none is created here.
The draft and feasibility records were untracked at inspection, so this is not an
acceptance of a frozen design SHA. No product edit, test execution, commit, push,
BC freeze or implementation branch is performed by this review.

## Evidence and scope

The baseline-to-design-HEAD diff is empty for cmd, internal, go.mod, go.sum,
.github and README.md. `git ls-tree -r --name-only` at the baseline inventories
**62 Go files: 38 production and 24 test files**. There are eight other tracked
baseline files. Every one appears below; the inventory is 70 unique paths.
No layer-local AGENTS/README adds another rule to these product paths.

Authority inspected includes root AGENTS/developmentInstructions; full #52 and
#21 comments; the complete user run contract; APS22/BR21; accepted focused
responsibility, migration and Slice records, especially Composition's serial
order, Domain's corrected unborn-HEAD policy, and the TUI State/View extraction
rules. The canonical Application IDs remain APP29-H03 confirmation and APP29-H06
reconciliation despite the transposed Git report headings. Accepted findings
remain open; source inventory is not evidence of their resolution.

Source inspection included each production file's responsibility/import/function
inventory, concrete service and TUI call sites, provider/storage implementations,
Git graph/diff/stash/checkout mechanisms, the standalone graph controller, every
test-file/test-name inventory, and current README/workflows/console/release notes.
The test reassignment table classifies existing assertions and names useful
replacement proof; it does not assert that those tests prove the new contracts.

Important source anchors are:

| Baseline anchor | Consequence for the migration |
|---|---|
| app/service.go Service, Load, LoadGraph, CreatePRWorktree; app/v3.go GraphHistory and DiffPullRequest | Preserve use cases; replace concrete fields/facades. Graph joins belong in Application, lane/text formatting in View, explicit Git reads in Git. |
| tui/model.go beginDeployment/updateDeployDialog, executeCommit, beginBranchFromCommit, saveSelectedLaunch | Configured target selection, stage-all-then-commit, branch creation from selected history, and save-as-default must receive explicit parity coverage. |
| tui/runtime_v0314.go handle314Key, executeDeploy314, activateSelectedWorktree314 | PR/branch/commit Enter deployment differs from Ctrl+Enter activation. Preserve both while removing branch-name substitution and stale confirmations. |
| worktree/worktree.go Deploy, pr.go CheckoutRef, branch.go ensureBranchStart | Configured branch deployment and general retargeting have different safety today. Do not reproduce reset/-C/-f as unreviewed compatibility. Exact selected Revision and conservative history protection govern all replacements. |
| graph/graph.go Load; graph/rows.go Rows; graphui/model.go Update | A Git file currently joins remote/worktree facts; graph lanes are pure display. Both graph entries need one accumulated graph generation before legacy deletion. |
| diff/reader.go Worktree/run | Actual working-tree diff runs `git diff` (index to working tree), despite its current "vs HEAD" label. Preserve the useful source behavior and make its label/API precise. |
| launch/config.go SaveCandidate, npm.go Build, make.go Build/findMakefile | Alias/default, explicit command override, exact script bytes, ordered targets and selected Makefile/cwd must survive the split. First save becomes default if none exists. |
| terminal/shell.go and docs/interactive-console.md | GH_TREE_SHELL override, ancestry choice, platform fallback, original active cwd, keyboard forwarding and CR/backspace output are retained capability. |
| .github/workflows/ci.yml and release.yml | Five builds do not cover twelve release assets; codereview-21 pushes lack CI. Bootstrap branch publication and clobber-capable publication cannot be used for testing. |

## One physical edit owner and coexistence

**MC** below is the sole owner of every old physical file: the Master-managed
Composition/shared-cutover contribution, implemented by one explicitly assigned
fresh worker and reviewed by a separate fresh reviewer. It is not permission for
Master to implement product code directly. At most one such writer is active for
any shared file; a later contribution receives an explicit serial handoff.

Target-layer workers create replacement code and tests only in their own new
folders. They may read old files as evidence, but may not move, delete, rewrite or
alias them. Distinguish replacement responsibility from the physical old-file
owner: many new layers can consume one file's behavior without sharing its edit
ownership. MC removes the old file once its mapped proofs pass. This choice avoids
an extraction phase in which each layer rewrites the same facade or wrapper.

The old product remains buildable and unchanged alongside new layers until final
entry cutover. New production code must not import or call legacy implementations.
There are no new compatibility shims planned. Any later shim needs an explicit
path, owner, permitted imports, bounded purpose and removal gate accepted before
creation; a wrapper hidden behind a new port is not an acceptable shim.

GitHub needs special treatment because its target folder already exists. Reserve
`internal/github/client.go` and `client_test.go` to MC, unchanged. The GitHub worker
creates the new implementation under **internal/github/adapter/**, a GitHub-layer
subpackage. It implements the new ports directly with private transport/DTOs and
does not import the parent legacy package. New Composition imports that subpackage;
old app/graph/tui/cmd keep importing the unchanged parent package. Delete the two
parent files only at the final retirement gate. Keep adapter as the final package
location to avoid a second move/API rename. The physical layer remains
internal/github; internal/github/adapter is not an eleventh layer.

This also makes staged enforcement feasible: exact old file paths, including the
parent GitHub files, form the temporary legacy set. Checks apply to every new file
and every platform-selected package; no blanket exception for internal/github/**
is allowed. A new GitHub file in the parent package would inherit its legacy
process import and must be rejected. Final legacy allowance is empty.

## Gate vocabulary and accepted integration order

The labels here are proposed migration gates, not new implementation Slices.

| Gate | Required evidence before progressing |
|---|---|
| M0 | Accepted/merged complete design; required seven BC versions reviewed/frozen; complete Slice selection and path ownership recorded. |
| M1 | Reviewed Composition CI/bootstrap-safety contribution: codereview-21 push CI, twelve cross-builds, staged architecture checks, automatic bootstrap-branch publication disabled; shared dependency edits serialized. Obtain exact-HEAD CI before relying on it. |
| M2 | Reviewed Domain values, then Application api/ports and State-owned viewmodel leaf integrated. Compile-time conformance/public-API purity and all platform selections checked. |
| M3 | Reviewed adapters serially integrated **Git -> GitHub -> Persistence -> Launch Discovery -> Runtime**. Native Runtime proof precedes any entry cutover. New folders cannot call old code. |
| M4 | Reviewed Application coordinator/usecases/operations integrated with fake-port and real-boundary proof, followed by State and then View. This is a horizontal integration order, not a claim that a folder is a completed Slice. |
| M5 | Full new layer stack passes each applicable vertical Slice through a headless composition harness, with real Git repositories/worktrees/remotes, actual storage/discovery and native Runtime; GitHub transport fixtures are explicitly identified. CLI still selects the legacy product. |
| M6 | Reviewed Composition constructor/root/host and the serial cmd entry cutover; repeat both entry modes, real key/event/resize/quit behavior and the full selected capability matrix. Both graph entries share one implementation. |
| M7 | MC retires legacy sources/tests only after M5/M6, mapped replacement tests and zero remaining production imports; strict dependencies/public API/platform inventory with empty legacy allowance; freeze integrated HEAD for full accepted verification. |
| M8 | Full exact-SHA verification, product PR/main gates, version/source/artifact identity, staged twelve-asset manifest and controlled extension install/upgrade; publication only after release authority/evidence gate. |

At M5, a **test-only composition harness** in internal/composition (external test
package permitted to wire layers) constructs the new adapters/coordinator and
drives State actions, Application events and pure View rendering. It must not call
old app/tui or substitute fake ports for all infrastructure. A later host-specific
test exercises Tea decoding/dispatch with API fakes and stays within restricted
host imports. Harness imports belong to Composition test policy, never a reason
to let State/View/host call adapters. Runtime native session fixtures use bounded
temporary helper processes and inspect descendants/resources. Windows/Linux/macOS
and the accepted additional-platform mechanisms need their real evidence.

M5 can construct the new stack from public constructors in tests without exposing
a second production CLI or premature cmd rewrite. The headless harness proves
the behavior boundary through View; M6 adds actual host/CLI transport, terminal
restore and root error propagation. Neither replaces M7/M8 verification. Record
exact worker/review/integration/harness SHAs at each gate and re-run affected
proof when a dependency or frozen contract changes.

## Production file mapping — 38 exact baseline paths

All rows have exactly one physical owner, MC. "Split/rebuild" means target workers
create the stated responsibility afresh, then MC deletes the old file at M7.
No row authorizes a wholesale wrapper/facade relocation. SLC references below are
provisional capability IDs for Master to reconcile with the final Slice appendix.

| Baseline file | Owner/action and exact replacement responsibility | Gate / capability |
|---|---|---|
| cmd/gh-tree/main.go | MC rewrite: tiny entry/exit; constructor/flags/context/root lifecycle to internal/composition; input/event transport to restricted composition/host. | M6; SLC-01/03/12/13 |
| internal/app/service.go | MC split/rebuild/delete: Application repository/graph/worktree/PR use cases and API; concrete construction to Composition; Git/GitHub mechanics behind ports. No Service facade retained. | M7; SLC-01..08 |
| internal/app/v3.go | MC split/rebuild/delete: Application graph/diff/launch/stash orchestration; View graph decoration/label prose; Discovery resolution; Persistence launch bytes; Runtime lifecycle. | M7; SLC-03/06/07/09/10 |
| internal/app/v312.go | MC split/rebuild/delete: Application launch/session/shutdown workflows; Runtime one registry; selected console entirely State. Remove two-manager routing and SelectConsole. | M7; SLC-09/10/12 |
| internal/app/v313.go | MC split/rebuild/delete: Application terminal start/input/resize by explicit Runtime capability, no membership/provider routing. | M7; SLC-11/12 |
| internal/app/v314.go | MC split/rebuild/delete: Application stash/deploy use cases; Git exact primitives. Preserve selected Revision/StashID in new API. | M7; SLC-05/06 |
| internal/config/config.go | MC split/rebuild/delete: Persistence codecs/state/config keys/bytes; Composition lazy default/override selection; Application namespace/default/active intent and configured-target interpretation. Domain receives identities only. | M7; SLC-01/04/05/09 |
| internal/diff/reader.go | MC split/rebuild/delete: Git diff source/file/stat/patch primitives; API semantic diff facts; View source labels/truncation notice. | M7; SLC-03/07 |
| internal/github/client.go | MC retire after coexistence: new internal/github/adapter transport/repository/PR/branch/error mapping owned by GitHub; Domain scoped IDs and Application facts through ports. | M7; SLC-01/02/03/08; no new parent-package symbols |
| internal/graph/graph.go | MC split/rebuild/delete: Git commits/parents/refs/HEAD reads; Application PR/worktree joins/completeness; API graph projection; View human freshness/labels. | M7; SLC-02/03 |
| internal/graph/rows.go | MC rebuild/delete: pure accumulated-DAG lane formatting in tuiview; no Git/GitHub imports or rewritten commit messages. | M7; SLC-03 |
| internal/graphui/model.go | MC split/rebuild/delete: common State graph navigation/generations, View graph rendering and Composition initial graph mode/host. | M7 only after both entry modes at M6; SLC-03 |
| internal/launch/config.go | MC split/rebuild/delete: Persistence run.json paths/codecs/read/write/version; Discovery provider validation/intent resolution; Application SaveCandidate alias/default workflow. | M7; SLC-09/10 |
| internal/launch/make.go | MC rebuild/delete: launchdiscovery private passive Make provider, ordered compatible members and exact selected file binding. | M7; SLC-09/10 |
| internal/launch/npm.go | MC rebuild/delete: launchdiscovery private npm/pnpm/yarn provider, exact scripts/command override/manager policy. | M7; SLC-09/10 |
| internal/launch/session.go | MC rebuild/delete: Runtime registry/resources/output/lifecycle; API session facts; no manager-current-session selection retained. | M7; SLC-10/12 |
| internal/launch/types.go | MC split/rebuild/delete: Discovery registry/private provider records; API/ports definitions and execution specifications; Domain LaunchPointID only. Do not move Candidate/Invocation wholesale into Domain. | M7; SLC-09/10 |
| internal/process/process.go | MC retire: separate private bounded short-command transports in Git and GitHub, Runtime private native session creation. Do not recreate generic shared process layer or call old ExecRunner. | M7 after all old importers removed; SLC-01..03/05..08/10..12 |
| internal/terminal/session.go | MC split/rebuild/delete: Runtime unified PTY/ConPTY resource ownership/byte transport; View pure CR/backspace/control interpretation; host owns disposable presentation cache. | M7; SLC-10/11/12 |
| internal/terminal/shell.go | MC rebuild/delete: Runtime private shell/ancestry/environment selection; preserve override/fallback; Composition supplies environment/start context only. | M7; SLC-11 |
| internal/tree/tree.go | MC split/rebuild/delete: Application namespace/strip-prefix/durable fallback policy; State pure immediate hierarchy/filter/selection; View leaf labels. No Domain namespace tree. | M7; SLC-01/02 |
| internal/tui/model.go | MC split/rebuild/delete: State navigation/search/forms/correlation; Application every product workflow/persistence change; View formatting; host Tea decoding/effect execution. | M7; SLC-01..12 |
| internal/tui/runtime_v039.go | MC split/rebuild/delete: State create-review/edit and ordered launch selection; App/Discovery compatibility validation; View create/launch picker appearance. | M7; SLC-05/09 |
| internal/tui/runtime_v0310.go | MC split/rebuild/delete: State Branch/Commits/Message focus and dirty-path selection; App branch relations/history/restore/checkout; View direction/status/detail formatting. | M7; SLC-02/03/05/07 |
| internal/tui/runtime_v0311.go | MC split/rebuild/delete: State active chooser/mnemonics/root-to-launch transitions; App discovery/activation; View headings/overlay/layout. | M7; SLC-02/04/09 |
| internal/tui/runtime_v0312.go | MC split/rebuild/delete: State console tabs/input focus and conditional follow; App lifecycle/event projection; View tabs/activity; host disposable timer. | M7; SLC-10/12 |
| internal/tui/runtime_v0313.go | MC split/rebuild/delete: State terminal input authority; host key bytes/measurement dispatch; App terminal commands; View version/console layout and shared measurement. | M7; SLC-11/12/13 |
| internal/tui/runtime_v0314.go | MC split/rebuild/delete: State single stash/deploy modal/selection; App exact deploy/stash/activation and relations; View panel/help/confirmation/stash/output. | M7; SLC-02/04/05/06/09/10 |
| internal/tui/view.go | MC split/rebuild/delete: tuiview render/theme/layout/dialog/diff/text; semantic relations to App and intent to State; remove Model receiver/backend access. | M7; SLC-01..12 |
| internal/version/version.go | MC retain path/rewrite: one Composition build/version identity with noninteractive output; value included in tested release commit. | M6/M8; SLC-13 |
| internal/worktree/branch.go | MC split/rebuild/delete: Git exact fetch/resolve/branch transition mechanics; App source/target choice and creation/checkout workflow. Existing local branch is never an exact-target shortcut. | M7; SLC-02/05/08 |
| internal/worktree/cleanup.go | MC rebuild/delete: Git confirmed tracked restore/recovery; App prepares/consumes confirmation; State/View prompt identity. | M7; SLC-07 |
| internal/worktree/history.go | MC split/rebuild/delete: App PR/branch source resolution; Git explicit pinned history primitive. | M7; SLC-02/03 |
| internal/worktree/ops.go | MC split/rebuild/delete: Git status/upstream/create/checkout/fetch/stage/commit/push/pull/history mechanics; Domain exact types; API/ports facts; App sequencing. | M7; SLC-02..05/07/08 |
| internal/worktree/pr.go | MC split/rebuild/delete: Git generic exact ref fetch/verify/guarded branch transition; App PR scope/private-ref choice/deploy/worktree workflow. | M7; SLC-03/05/08 |
| internal/worktree/staging.go | MC split/rebuild/delete: Git status path parsing/index/stash mechanics; App stage-all/commit and stash operation sequencing. | M7; SLC-06/07 |
| internal/worktree/stash.go | MC split/rebuild/delete: Git exact stash facts, legacy metadata codec and safe mutation/recovery; Domain scoped StashID; App confirmation/read model. Positional ref and metadata stay nonauthoritative. | M7; SLC-06 |
| internal/worktree/worktree.go | MC split/rebuild/delete: Git inventory/root/common identity/path/ref/dirty/history guards; Domain values; App configured-target and general deploy sequencing; API outcomes. | M7; SLC-01/04/05 |

## Existing test files — 24 exact paths

MC alone retires the old tests at M7. The target owner listed creates new tests in
its own folder before deletion; no worker moves an old file. Mixed test files are
decomposed by assertion, including their fixtures/fake backends. In particular,
worktree fixture helpers used across old tests cannot be deleted early to make
the source migration look complete. Every assertion gets a replacement reference
or an explicit obsolete/unsafe-behavior disposition in the final mapping.

| Baseline test file | Owner/action; replacement test responsibilities and required correction |
|---|---|
| internal/config/config_test.go | MC retire; Persistence missing/default/explicit-empty/case/order/legacy codec tests; App durable folder/worktree scope and publication tests. Do not retain owner/name as universal active-worktree identity. |
| internal/diff/reader_test.go | MC retire; Git NUL rename parsing, bounded patch/stats, first-parent/root semantics; View renders separate truncation/source facts. |
| internal/github/client_test.go | MC retire; GitHub adapter parsing/required exact identities/bounded pages/auth-context/argv tests; App host/fork creation binding. Current test forbidding all pagination becomes bounded-completeness evidence, not a legacy invariant. |
| internal/graph/graph_test.go | MC split/retire; Git graph/ref/HEAD limits/classification tests; App remote/worktree annotation tests; composition full graph fixture. |
| internal/graph/rows_test.go | MC retire; View merge/multiple-root/lane tests against immutable graph DTOs and accumulated page boundaries. |
| internal/launch/launch_test.go | MC split/retire; Discovery exact colon script/nested manifest/root-lock-only/cwd/provider-grammar tests; Persistence run.json round-trip; App save/default/reopen/invalid intent. |
| internal/launch/make_test.go | MC split/retire; Discovery simple targets/native ordered stack/selected manifest; Persistence serialized ordered targets; App named saved Make default. |
| internal/launch/session_test.go | MC retire; Runtime bounded-output/exit/failure/concurrent launch tests using owned helper processes, strengthened to descendants/barriers/idempotent Stop rather than root-only stopped flag. |
| internal/terminal/terminal_test.go | MC split/retire; View CR/backspace/control parser tests; Runtime environment replacement and GH_TREE_SHELL override; native PTY/ConPTY lifecycle tests newly required. |
| internal/tree/tree_test.go | MC split/retire; App strip-prefix/default/namespace intent tests; State hierarchy/filter/nearest permitted ancestor tests using complete versus incomplete projection. |
| internal/tui/model_test.go | MC split/retire; State namespace/search/mode tests; App confirmation/exact deployment tests; View full 40/64-hex detail tests; composition retained configured-target flow. |
| internal/tui/model_v2_test.go | MC split/retire; State history focus/interactive-worktree fallback/pull key; App scoped exact PR/worktree association; View dirty/upstream/narrow layout with unknown facts. |
| internal/tui/model_v3_test.go | MC split/retire; State ordered Make selection and diff return; Discovery compatibility validation; View file/patch/help tests from actual action keymap. |
| internal/tui/runtime_v039_test.go | MC split/retire; State create review `e` editing and Make group selection; View bounded create dialog/project labels; App/Discovery enforce cross-project refusal independently of cursor UX. |
| internal/tui/runtime_v0310_test.go | MC split/retire; App scoped PR head/base direction; State branch subpane cycling; View commit message/dirty-cause rendering. |
| internal/tui/runtime_v0311_test.go | MC split/retire; View cell-bounded overlay and heading/mnemonic tests; State Alt+C/M and active-worktree chooser. |
| internal/tui/runtime_v0312_test.go | MC split/retire; State Alt+O/Ctrl+C exit modal/root-to-launch/Alt-number selection; View console tabs/supplied activity frames, no provider-inferred working/idle. |
| internal/tui/runtime_v0313_test.go | MC split/retire; Composition host key-byte and Alt-number decoding tests; View version display; Composition noninteractive binary version test compared to release identity. |
| internal/worktree/cleanup_test.go | MC retire; Git tracked restore-from-index/untracked refusal with post-confirm edit, captured-original and retained-recovery cases; App single-use confirmation test. |
| internal/worktree/ops_test.go | MC split/retire; Git status/history/upstream/create/retarget mechanics; App PR prepare/create/reuse and source-versus-target branch workflows; real composition worktree fixtures. |
| internal/worktree/pr_v2_test.go | MC retire; Git exact guarded fast-forward and refusal to lose local commits; App expected-selected-PR propagation. |
| internal/worktree/staging_test.go | MC split/retire; Git selective staging/unstaging/exact path grammar/stash contents and index restoration; App stash sequencing/effect classification. |
| internal/worktree/stash_test.go | MC split/retire; Git managed-metadata compatibility/parser tests and positional-locator display validation; Domain StashID tests; actual destructive authority tested by OID after reflog reorder. |
| internal/worktree/worktree_test.go | MC split/retire; Git porcelain/clean/dirty/primary/current/occupancy/detached-history/exact-fetch guards; App configured target deploy workflow; composition real exact deployment. Replace reset/force compatibility assertions with documented safe fast-forward/new-branch/detached outcomes. |

There are **no baseline app, cmd, version, graphui or runtime_v0314 tests**.
Their absence is not a coverage waiver. M5/M6 require new Application operation,
Composition lifecycle and latest stash/deploy/state/view tests independent of the
old test count. The baseline tests named HelperProcess and old fake runners are
test infrastructure, not separate production capabilities or publishable binaries.

## Other baseline files — eight exact paths

| Baseline path | Sole owner and disposition |
|---|---|
| .github/workflows/ci.yml | MC serial rewrite at M1; retain Windows/Linux/macOS test/vet/build and race, expand twelve cross-builds and new-branch push evidence, staged then strict dependency/API/package checks. |
| .github/workflows/release.yml | MC disable unsafe automatic bootstrap publication at M1; replace publication with exact source/version/tag/evidence/staged-assets gate at M8. No clobber. |
| go.mod | MC retain/serial dependency edits; layer workers request reviewed additions. Preserve Go/toolchain/platform support and remove obsolete PTY dependencies only after both new Runtime and legacy cutover compile on every target. |
| go.sum | MC synchronized with reviewed go.mod/toolchain resolution; keep integrity and exact dependency evidence. Never independent tidy commits from multiple workers. |
| README.md | MC revise at M6/M8 to actual accepted keymap/diff source/safety/limits/version/install guarantees; every existing capability maps below. |
| docs/interactive-console.md | MC retain historical v0.3.13 record; current README/new release documentation describes v0.4 shell/input/output/cleanup accurately. If updated, label historical versus current behavior explicitly. |
| docs/v0.3.14.md | MC retain historical release record; do not rewrite old release facts into v0.4 claims. Current docs explicitly explain safe activity semantics and selection/deploy/activation. |
| .gitignore | MC retain; revise only for reviewed build/staging output location, with no ignored source/verification evidence. |

## Proposed complete vertical capability coverage

These SLC IDs are provisional design inputs. They describe externally observable
behavior crossing layers, not a list of implementation folders. Domain/API/BC/CI
work is prerequisite contribution work, not an independent product Slice.
For each row, all applicable layer contributions are reviewed, serially integrated
in the M2-M4 order, then exercised through M5 before M6 cutover. Master assigns
final acceptance-test IDs and every open finding to these or more finely bounded
final Slices. No subset is silently selected.

| Proposed Slice / retained capability | Full vertical path | Concrete acceptance contract and baseline hooks |
|---|---|---|
| SLC-01 Start, restore and refresh navigator | Composition -> App -> Git/GitHub/Persistence -> App -> State -> View | `--repo`, help/overrides, remote-only browsing, same remote/local capability diagnostics; restore strip-prefix namespaces and saved folder, PR/branch switch, hierarchy/search/clear/back navigation, stable selection. Partial/capped or failed source cannot prove absence or erase intent. model namespace/search tests + config/GitHub/tree fixtures. |
| SLC-02 Branch context and canonical relations | State -> App -> Git/GitHub -> App -> State -> View | Enter branch opens Branch/Commits/Message, preserves full OID/message/detail scroll and Alt+B/C/M/Shift+Tab behavior. Head/base PR direction, multiplicity/forks and local occupancy marker derive from scoped facts; local/upstream/GitHub distinction includes none/resolved/gone/unresolved, detached/unborn and fresh/cached/unknown. |
| SLC-03 Graph, history and diff/review | Both Composition entry modes -> State -> App -> Git/GitHub -> App -> State -> View | In-app `g` and `--graph` use same accumulated DAG/selection/renderer; refs/tags/HEAD/PR/worktree labels and `L` continuation without lane reset. PR/branch/worktree history, selected first-parent/root commit diff, exact PR merge-base range, working-tree/index and staged/HEAD diff with rename/binary/bounded patch. Return mode/scroll and full identity remain available; stale pages/diff responses do not redirect intent. |
| SLC-04 Active worktree selection and durable context | State -> App -> Git/Persistence -> App -> State -> View | Worktree pane Enter/a, Alt+A chooser and Ctrl+Enter selected-object activation remain distinct from deploy. Validate current inventory, serialize activation, durable commit then publish; scoped saved/current/deterministic fallback; linked worktrees share common scope, independent clones do not. Existing terminal retains original cwd after activation. |
| SLC-05 Create, retarget and deploy exact PR/branch/commit | State -> App -> Git/Persistence -> App confirmation -> State/View -> App -> Git -> App -> State/View | Create from PR/branch/selected history with editable path/local branch; retarget secondary worktree; configured legacy targets remain usable via mapped target intent. Enter PR/branch/commit deploys exact selected Revision detached; dirty path offers bound stash-and-deploy. Exact target is reverified before mutation, no stale local-tip substitution, protected primary/current/history/occupancy/ignored/untracked state, single-use stale-confirmation refusal and truthful effects. No force-rewind parity requirement. |
| SLC-06 Repository stash cockpit | State -> App -> Git -> App -> State -> View | List managed/unmanaged repository stashes across worktrees, show exact identity/origin/patch/files, Alt+S/left-right navigation, push tracked+untracked, apply, confirmed pop/drop and legacy latest-pop shortcut resolved to stable OID before confirmation. Clean target required, index restored, conflict surfaced. Shift positions/create another stash between selection and execution: never act on another OID. Applied-but-retained result prevents automatic reapply. |
| SLC-07 Inspect changes, stage/unstage, commit and tracked restore | State -> App -> Git -> App -> State -> View | Dirty cause distinguishes staged/worktree/untracked/conflicted/rename; selected path staging and mutable-diff staging/unstaging; normal `m` commit preserves existing explicit stage-all-then-commit semantics in its request/prompt. `r` confirms exact tracked path/index/content; edits after confirmation refuse/recover without losing original/new bytes. No silent untracked deletion/conflict resolution. Mutation applied + refresh failure stays a known applied effect. |
| SLC-08 Fetch, fast-forward pull, non-force push, new branch, draft PR | State -> App -> Git/GitHub -> App -> State -> View | Fetch/prune updates scoped freshness; pull exact observed upstream fast-forward with dirty/detached/gone/diverged refusal; push pinned OID to explicit remote ref with separately checked upstream setup. New branch at current or historical exact commit. `o` form preserves base/title/body/draft; explicit host/head/base/fork and expected remote head prevent same-name substitution/implicit push. Interrupted local/remote effects reconcile before retry. |
| SLC-09 Discover, choose, save and reopen launch intent | State -> App -> Git/Discovery/Persistence -> App -> State -> View | Active list, Ctrl+F5 chooser, nested projects/exclusions, exact `dev:wan`, npm/pnpm/yarn/command overrides, ordered Make group from one project/source/executable. Save named alias/default including first-save policy; reopen same selected worktree run.json losslessly. Collision/version/link/source changes yield explicit conflict/stale/unsupported results; no provider execution while discovering. |
| SLC-10 Launch, multi-console stop and restart | State -> App -> Git/Discovery/Persistence/Runtime -> App -> State -> View | F5 in Launch/Enter starts selected item, F5 elsewhere starts saved default; each new session has unique ID/tab and bound cwd/spec. Output bounded/ordered; Alt-number and left/right navigation preserved for noninteractive tabs. Shift+F5/Ctrl+C launch stop is proved cleanup, F6 old barrier then new ID; failed cleanup forbids replacement. Launch then Alt/Tab navigation retains newer focus when completion arrives. |
| SLC-11 Interactive shell and terminal input/resize | State/host -> App -> Git/Runtime -> App -> State -> View/host | Alt+T opens PTY/ConPTY in Active cwd using GH_TREE_SHELL/ancestry/fallback policy; printable/Enter/arrows/Tab/control input forwarded only while selected terminal owns input. Reserved global mnemonics still navigate; Ctrl+C sends ETX and is not whole-tree stopped proof. One measured positive viewport by SessionID/generation drives resize; bounded safe CR/backspace output; native close/restart/stop resource cleanup. |
| SLC-12 Quit, cancel and aggregate cleanup | State/host/Composition -> App -> all accepted operations/Runtime -> root result | Outside-console Ctrl+C confirmation, accepted quit, root cancel, UI/constructor failure all quiesce submissions and retained operation outcomes; stop every owned session, join output/readers/bridges/timers, combine primary and cleanup failures, nonzero residual exit. Late confirmation/terminal/query/timer cannot reopen disposed/newer State. Real active-session shutdown before M6, host terminal-restore/error paths at M6. |
| SLC-13 Install/upgrade delivered v0.4 | Verified source/tag -> CI/staging/release -> gh extension -> binary Composition -> App/Persistence | All twelve assets, exact source/version/workflow/artifact identity, noninteractive version/help without auth/storage, controlled install and upgrade from preserved v0.3 state, correct OS/architecture executable naming. No automatic existing-asset replacement. This Slice finishes only at M8; pre-cutover build/staging assertions do not imply published install success. |

All navigation keys/forms/modals in a capability row need a State test and the
actual applicable action/help projection in View. All async behavior needs
OperationID, projection source/generation and global user-intent interleavings,
including selecting console B while A restarts, switching worktrees before
launch/stash/status results, stale branch/diff pages, and modal cancellation.
SLC-01..12 must be proven across the new full stack before final legacy retirement.
SLC-13's release/publication portion is deliberately later than headless/CLI parity.

## Twelve assets and provenance gate

Keep the existing exact installed-asset names. CI and release staging use this
same reviewed manifest, while native runtime evidence is recorded separately.

| GOOS | GOARCH | Exact asset |
|---|---|---|
| darwin | amd64 | darwin-amd64 |
| darwin | arm64 | darwin-arm64 |
| freebsd | 386 | freebsd-386 |
| freebsd | amd64 | freebsd-amd64 |
| freebsd | arm64 | freebsd-arm64 |
| linux | 386 | linux-386 |
| linux | amd64 | linux-amd64 |
| linux | arm | linux-arm |
| linux | arm64 | linux-arm64 |
| windows | 386 | windows-386.exe |
| windows | amd64 | windows-amd64.exe |
| windows | arm64 | windows-arm64.exe |

Reject omitted, extra, skipped, differently named or conflicting assets. Preserve
`gh extension install Hans-Einar/gh-tree` and `gh extension upgrade tree`; record
the delivered binary/architecture/version and controlled storage migration.
Baseline observed installed v0.3.13 is inventory only; this review neither changes
that installation nor claims an upgrade. Cross-build success alone does not
prove FreeBSD/Windows/Unix Runtime behavior or extension delivery.

## Design review findings and reconciliation requirements

| ID / severity | Gap or risk grounded in current draft/source | Proposed disposition required before design acceptance |
|---|---|---|
| MIG52-M01 / MEDIUM | REFDES reserves internal/github as new layer while legacy client uses old process and is imported by old app/graph/tui/cmd. Letting the GitHub worker replace it at M3 breaks coexistence or forces shared edits before cutover. | Accept the internal/github/adapter coexistence map or another equally explicit noncolliding final subpackage. Assign old two-file deletion to MC and prohibit parent-package allowance from covering new code. |
| MIG52-M02 / MEDIUM | "Verify all retained Slices before retiring old paths" alone can become circular if the new stack is only constructible after cmd cutover. Accepted State/Composition records require behavior proof before the switch. | Name M5 test-only full-stack composition harness and exact evidence; keep cmd legacy until its proofs pass. Host/CLI transport re-verification then precedes deletion. |
| MIG52-M03 / MEDIUM | Broad capability labels can miss source-backed operations: configured targets, stage-all commit, historical new branch, latest-pop shortcut, launch command overrides/first default, Ctrl+Enter activation and shell override/ancestry. No V314 tests exist to catch several losses. | Preserve them explicitly in final API/Slices/capability and baseline-test disposition. Safety changes such as configured-branch non-FF refusal must be documented with usable detached/new-branch alternatives, not silently dropped. |
| MIG52-M04 / MEDIUM | Old tests mix Git mechanisms, use cases, rendering and fixture helpers. Moving a whole test to a layer would import other legacy adapters and retain false contracts such as positional stash identity, incomplete-list absence or unproved stopped flags. | Final map records one old-file owner plus per-assertion new test/reference or obsolete-behavior rationale. Retain old tests until all their importers retire; add missing App/host/V314 coverage independently. |
| MIG52-M05 / MEDIUM | Current Working-tree diff label says HEAD, but the actual command compares index to working tree. Configured branch deploy also force-moves branches while general PR checkout refuses history loss. Textual parity can accidentally copy unsafe/inaccurate semantics. | Final API/verification state exact diff endpoints and branch mode/history policy; tests use staged-plus-unstaged fixtures and non-FF configured target. Preserve useful capability while correcting labels and refusing unsafe mutation. |
| MIG52-M06 / MEDIUM | Draft says exact unchanged legacy paths may remain but does not yet supply a complete machine-checkable map. CI path-prefix exemptions could hide new Windows code, adapter public DTOs or a relocated wrapper. | Final inventory covers all 62 Go + eight shared baseline files and planned target packages; each physical path one owner. CI checks all twelve platform selections, public signatures and new/changed files separately from exact old files; final allowance empty. |
| MIG52-M07 / MEDIUM | Safe release bootstrap must precede reliance on codereview-21 CI; M8 cannot be the first point where release/v*-bootstrap auto-publication is disabled. | First reviewed Composition product contribution at M1 disables that trigger and establishes new branch CI without an early final product PR. Release build/staging and later publication are separate exact-SHA gates. |

These findings are design completeness inputs, not newly counted baseline product
defects or a final design rejection. Master must link their resolution into the
normative appendices and re-review frozen artifacts. Every accepted report
finding remains independently traceable: final FindingDisposition needs canonical
ID, decision/clause, affected BC versions, implementation layer(s)/Slice(s), exact
implementation/review/integration SHAs and verification evidence. Multiple related
IDs may share proof, but report acceptance or a migration-map row cannot close one.

Recommended next action: Master reconcile this proposal with API--001 and the
final Slice/verification/finding appendices, validate the inventory mechanically,
then freeze the complete design HEAD for fresh independent review. No downstream
implementation or BC freeze follows from this feasibility record alone.

Inventory verification executed for this record: compare the first column of
the file-mapping tables against `git ls-tree -r --name-only` at the stated
baseline. Result: **70 baseline / 70 mapped / 70 unique; 38 production Go and
24 test Go; zero missing and zero duplicate paths**. Shared worktree status remains
untracked SDP/Design artifacts; no staged or committed product edits were made.
This structural check is not a product test or proof of the proposed migration.
