# Slices--001 — complete vertical migration contract

State: DRAFT, normative appendix to REFDES--001 under #52.
Selection: all SLC-01..13. No subset is authorized by omission or folder completion.
The per-Slice behavior below is source-grounded in the complete baseline inventory
in MigrationMap.yaml. API--001 defines its requests/results; Verification--001
defines the required failure, platform and evidence gates.

## Integration prerequisites and ownership

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

M0..M8 are integration gates, not product Slices. For every SLC-01..12, integrate
the relevant reviewed Domain/API/ports/viewmodel leaves at M2; relevant reviewed
adapter contributions at M3 in the listed order; App use cases, then State and
View at M4. Every contribution records layer, Slice ID, Issue, exact source SHA,
review disposition, BC version and resulting integration SHA. Different Slices
may share a primitive; its evidence is linked rather than counted as new behavior.

MC constructs a test-only new-stack Composition harness at M5, driving State
actions -> API -> real Git/Storage/Discovery/Runtime -> events -> State -> View.
GitHub transport fixtures are labeled. The legacy cmd remains untouched until
these tests pass. M6 separately validates actual Tea host/key/resize/root behavior.
No dependency on a second product CLI or fake-only full stack is accepted.
SLC-13 prepublication build/staging checks run before final product PR; published
install/upgrade completes after the exact main/tag/draft/publication gates. This
ordering avoids making publication a prerequisite of product integration.

All old physical paths remain solely MC-owned, including test helpers. Every
product contribution, including MC changes, is implemented by a fresh worker and
reviewed independently. Adapter workers cannot retire old files. Unsafe baseline
force/reset/positional-stash behavior is corrected according to REFDES, not copied
as a compatibility requirement. Configured branch targets that cannot safely
advance return an explicit refusal with detached/new-branch alternatives.

## Selected vertical Slices

| Slice / retained capability | Full vertical path | Concrete acceptance contract and baseline hooks |
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

## Slice evidence and boundary coverage

Each Slice SLC-NN has acceptance proof V-E2E-NN plus the applicable detailed checks
in Verification--001. Every action/form/key named above requires State and actual
View action/help proof. Every asynchronous source requires operation, query-slot
generation and user-intent interleavings; successful happy paths alone fail parity.

| Contributions | Required boundary versions after the later freeze gate |
|---|---|
| Every UI Slice SLC-01..12 | BC--TUIState--Application 1.0.0 and BC--TUIState--TUIView 1.0.0 |
| Local facts/mutations SLC-01..08 plus worktree scope for 09..11 | BC--Application--Git 1.0.0 |
| Remote facts/PR SLC-01..03/05/08 | BC--Application--GitHub 1.0.0 |
| Saved intent/config SLC-01/04/05/09/10/13 | BC--Application--Persistence 1.0.0 |
| Discovery/resolution SLC-09/10 | BC--Application--LaunchDiscovery 1.0.0 |
| Sessions/terminal/shutdown SLC-10..12 | BC--Application--Runtime 1.0.0 |

These version numbers are planned, not frozen by this appendix. SLC-12 also
requires cancellation/join obligations of every active port. Domain and Composition
are governed by design/API/import rules, not an invented omnibus adapter contract.

Completion state is currently planned for every Slice. M5/M6 proof, replacement
test dispositions, integrated exact SHA and final verification must be recorded
before SLC-01..12 can become verified. SLC-13 requires delivered-asset evidence;
review acceptance and cross-build success cannot mark it complete.
