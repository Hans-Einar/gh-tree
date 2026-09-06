# M2 parallel leaf contribution contracts

State: ACTIVE kickoff. Authority: #59 and #60 under #21; I-03 / M2.
Accepted Domain source/review/integration6113ca55b57cb628b016c483070dbb52cd5dd79a.
Source CI34033181820 and integration push CI34033772732 each passed18 applicable
jobs, with the one explicit M3 helper conformance skip. Domain review had no findings.
All13 Slices stay selected; all143 baseline findings remain open.

| Contribution | Dedicated branch/worktree | Sole worker-owned files |
|---|---|---|
| #59 Application API/ports | codereview-21/layer-application-api / application-api-ports | New internal/application/api and ports values/interfaces/tests/README; SDP/Implementation/CR-#21/Application/M2-API--001.md |
| #60 State-owned viewmodel | codereview-21/layer-tuistate-viewmodel / state-viewmodel | New internal/tuistate/viewmodel values/tests/README; SDP/Implementation/CR-#21/TUI-State/M2-Viewmodel--001.md |

Both worktrees are under C:/Users/hanse/GIT/gh-tree-wt and begin from the Master
kickoff after Domain completion. These leaves are independent: viewmodel imports
Domain/pure standard values only, not API/ports or parent State. Each gets a fresh
worker and separate fresh reviewer. No reducer/View renderer, Application workflow,
adapter, legacy retirement, module/workflow, frozen design/BC or program metadata
edit is delegated. Master coordinates shared changes and serial integration.

#59 implements every family and exact signature in all seven frozen BCs/shared
BoundaryTypes, with [API-Inventory--001](API-Inventory--001.md) as a completeness
checklist only. Five Client methods,28 commands/results,13 queries,six events and
seven ports/48 methods must be concrete and conforming, with validated immutable
values, proper ownership/private wrappers and result-plus-error/recovery semantics.
The actual accepted Domain constructors/tagged values are the dependency.

#60 implements the full independent viewmodel shape in State/View BC: snapshot,
viewport/focus/element/panes, exact semantic rows/graph/diff/launch/console models,
local modal/action/notice values, measurements/rectangles and output presentation
inputs/results. Preserve local generation domains, exact target detail, unknown/
partial facts and copied data. No Application operation/confirmation capability,
backend callback, I/O, state reducer, layout renderer or actual effect execution.

Each worker freezes/pushes exact source with actual tests/limits. Separate reviewer
inspects real source/tests/contracts and meaningful negative/positive controls;
corrections require re-review. Current applicable branch CI must pass. Master
integrates accepted API/ports then viewmodel serially, verifies integrated CI and
records source/review/integration SHAs before M2 completion and adapter Issues.
No whole Slice or full V-APP/V-STATE/V-VIEW/native proof follows from leaf tests.
Any insufficient frozen boundary triggers BC-CHANGE before affected code proceeds.

M2 API correction source frozen for independent re-review:4700ea72f080cae1f0a22cf21e80876d88617754,
productc96ca9244a30e16808743a0cd3e0b557e75080d7. Master verified clean freeze, permitted
API/ports/report scope and no product change in the312b534 metadata merge. Worker
reports full local test/race/vet/build/all12 gates pass; CI34041668920 and bounded
report were pending at dispatch. Primary and separate Runtime/Storage reviewers
reopen immutable source, all26 original cases, positive partial/compound controls
and related new semantics. M259-M01..M08 remain unresolved until actual acceptance;
viewmodeld279c3f remains held for API-first integration. No M2/Slice/native closure.
