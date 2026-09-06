# M2 State-owned viewmodel worker report

State: IMPLEMENTED — candidate for separate fresh independent review.
Authority: #60 under #21; Sprint-004-v04 / I-03 / M2.
Role: fresh TUI State viewmodel Worker, m2_viewmodel_worker. Date: 2026-09-06.
Branch: codereview-21/layer-tuistate-viewmodel.
Worktree: C:/Users/hanse/GIT/gh-tree-wt/state-viewmodel.
Kickoff HEAD: 29a12cbd6f2a0e189439b9ca25889f9f566afd6c.
Accepted Domain dependency: 6113ca55b57cb628b016c483070dbb52cd5dd79a.
Frozen candidate is the commit containing this report; Master records its exact
source/review/integration SHAs separately. No author acceptance is claimed.

## Authority and scope

Read fresh root AGENTS/developmentInstructions, full #60/#21 and comments,
predecessor authority, accepted REFDES/API/MigrationMap/Slices/Verification,
current sprint/index/ledger and M2-Surfaces/Plan, accepted State/View reviews,
baseline presentation/form inventory and actual accepted Domain source/README.
The full State/View and State/Application BCs plus shared BoundaryTypes are
effective FROZEN1.0.0 under BCFreeze--001, merge97cc0d8257603766dd741b49b7d8005857b421a9.
M1 architecture guard remains unchanged. No boundary change was required.

Only new internal/tuistate/viewmodel values/tests/README and this bounded report
change. No parent State/reducer, View layout/render/interpreter, API/ports/adapters,
Composition/legacy/module/workflow, accepted design/frozen BC or program metadata
is edited. Existing CLI behavior remains on the legacy stack. All13 selected
Slices and143 open baseline findings remain; this shared leaf closes none of them.

## Concrete implementation

The leaf implements Snapshot, Viewport, FocusPath/PanePath, closed ElementID,
PaneModel with all ten typed panes, semantic branch/worktree/PR/history/stash/
graph/diff/launch/console rows and full details, local ModalID/OwnerKey, all20
retained modal purposes and closed body variants, semantic ActionBinding/status,
Rect/PaneRect/ConsoleRect/Measurement, OutputInput, ByteRange and safe copied
ConsolePresentation. internal/tuistate/viewmodel/README.md contains the consumer
inventory, constructor/access pattern, modal fields and correlation semantics.

Every published family has private representation and validating constructors.
Spec records are copied admission/access values, never owned state exposed by
reference. All nested slices/bytes and optional record payloads are recursively
copied on construction, Fields/access and Clone, including nested SourceStatus
notices, graph annotation/commit metadata, modal target/fields/choices, console
presentation lines/summary and key chords. There are no public mutable maps,
callbacks, interface variant payloads, readers, channels or native/backend objects.
Generic Optional represents presence only; containing families validate and clone.

ElementID is a comparable private closed union of namespace/key and exact Domain
repository/PR/branch/revision/worktree/stash/launch+optional saved alias/session
values. Full40/64 revisions/stashes/targets remain inspectable; labels, alias case,
positions and scopes are not silently normalized. Branch/upstream/PR/worktree
relations and partial/unknown badges are supplied facts; constructors perform no
source association or business relation inference. Commit messages and ordinary
text/patch facts remain verbatim for later View sanitation. ConsoleLine is the
interpreter-result exception: valid UTF-8, no C0/DEL/C1 control escapes.

PresentationGeneration, ViewportGeneration, ContentGeneration and SourceGeneration
are distinct supplied types. Known0/tiny dimensions remain actual bounds;
UnknownViewport alone represents initial fallback eligibility. Snapshot receives
animation/version facts. Console source/presentation/content and SessionID must
match copied presentation; output offsets are overflow-checked with explicit gaps.
No clock, environment, OS, process, network, storage, invocation builder, style
global or effect execution exists. Imports are Domain, errors, unicode/utf8 and
pure supplied time values. Source timestamps retain original explicit offsets
independently of observation/freshness. Stash patch detail covers all five frozen
views, exact two/three parent OIDs and actual tree OIDs without fabricated commit
Revisions. Duplicate stash-object observations remain available. The final inventory
audit added these retained fields before freeze; no frozen contract was changed.

Modal purpose/body/exact target/field/choice and focus/mode consistency are
validated, with enabled cancel retained. ModalID/OwnerKey are local presentation
keys and cannot be converted to Application approval capabilities. Measurement
validates supplied geometry within known bounds and correlates viewport, outer
presentation, modal, pane membership and console identity/content. It does not
compute layout or prove adequate positive confirmation geometry; the later pure
View/State/host gate owns that proof and native resize/approval behavior.

## Local verification

Native Windows/amd64; Go go1.25.0 from
C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin,
added only to each command environment. Git2.43.0.windows.1; gh2.96.0.
No installed extension, user configuration or external product resource changed.

| Command / control | Actual result |
|---|---|
| gofmt -w internal/tuistate/viewmodel; final gofmt -l | New files formatted; final list empty. |
| go test ./internal/tuistate/viewmodel -count=1 | PASS:12 top-level tests, all ten panes/all20 modal purposes, full identity/variant/scope/copy/generation/measurement controls. |
| go test ./internal/tuistate/viewmodel -cover | PASS;78.1% statement coverage, not independent acceptance. |
| go test -race ./internal/tuistate/viewmodel -count=1 | PASS native Windows/amd64. |
| go test ./... -count=1 | PASS, unchanged baseline and M1 architecture tests included. |
| go vet ./...; go build ./... | PASS. |
| go run ./internal/composition/architecture | PASS all12 selections and public/external-test types;61 exact legacy/shared-entry allowances. No checker suppression/change. |
| go run ./internal/composition/architecture -targets | PASS exact twelve accepted asset inventory. |
| go run ./internal/composition/architecture -runtime-prerequisite | pending-m3; helper conformance explicitly NOT RUN. |
| External compiler overlay positive | Public immutable constructor/access compiles, exit0. |
| Five external compiler negatives | Expected diagnostic/exit1 for private ElementID tag, private Snapshot validity, nil ModalBody, scalar ModalID cast and callback payload. |
| git diff --check | PASS. |

External tests construct realistic nested values, mutate admission/access copies
and recursively compare actual slice backing addresses through snapshots/modal
clones. They preserve full OIDs/messages and partial-source facts; reject foreign
identity, invalid enum/option/body fields, missing exact confirmations, unknown
upstream counts fabricated as zero, stale modal/session/generations and overflowing
or out-of-bound geometry/byte ranges. Package-local adversarial tests forge
contradictory private variants and require rejection. No product/unit/guard test
failure occurred during this authoring pass. Native/platform behavior beyond the
local value tests is not inferred from their success.

Compiler controls used go test -overlay <file.json>
./internal/tuistate/viewmodel -run '^$' replacing closed_test.go only. Candidate
files were untouched; snippets/overlays/logs/results are retained at
C:/Users/hanse/AppData/Local/Temp/gh-tree-m2-viewmodel-4b7f2db236e94f9d9a592738a1b95a7d.
Each negative checked its specific diagnostic and a separate positive compiled,
so unrelated build failure is not counted as proof. These are worker controls,
not separate independent review evidence.

## Frozen handoff and remaining gates

Commit/push the candidate; configured branch CI must pass at its exact HEAD.
CI has not started when this report is authored. Send actual run URL/job results
and source SHA to Master after push and stop authoring for a separate fresh
reviewer. Corrections require explicit re-review. No author summary substitutes
for the reviewer's actual source/tests/contracts inspection.

This is partial M2/V-VIEW-03 value/purity evidence. Layout/Render/InterpretOutput,
reducer/event ordering, real terminal/visual geometry, native resize, integrated
State/View/E2E, complete Slices and release remain later gates. M2 itself waits
for accepted API/ports and viewmodel contributions plus serial integration CI.
Master alone integrates API/ports then this accepted leaf and records sprint,
CurrentIndex/Relations/Ledger/verification evidence. No PR, merge, Issue closure,
tag, publication or release-ready claim is made here.
