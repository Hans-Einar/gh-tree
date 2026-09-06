# M2 parallel leaf contribution contracts

State: COMPLETE atd60afe8fb763269964bcdc05958c8cdbc7849dea.
Authority: completed #59 and #60 under #21; I-03 / M2.
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

Second API/ports review at246e2a6b02e6812cc7579770e5aac973e19b3073 remains
CHANGES_REQUIRED for M259-M02/M03/M04/M05/M06:9 additional invalid admissions and
one legitimate recovery false refusal. All26 original cases and226 copy controls
pass; M01/M07/M08 resolved. Master verified primary88/bounded66 exact source blobs
and archived complete raw reports/tests/logs/manifests in M2-API-Review--002.
SourceCI34041943143 and prior4700 CI34041668920 each18/explicitM3skip. Worker
corrects existing groups with no-op/compound/complementary-subject controls before
new frozen re-review/currentCI. Missing-parent anchor interpretation is no finding;
original Expected remains in native manifest, guarded Original may use actual parent.
No API/VM integration or M2/full Slice/baseline/native/release closure.

Successor6e289c4b47a7816493f376c886c00bf0df4835e8 is frozen/pushed for independent
re-review of M02/M03/M04/M05/M06. Root verified the20-file API/ports-only diff and
whitespace; product authoring stopped, only the bounded report was dirty. Both
primary and separate Runtime/Storage reviewers reopen immutable source, old/new
probes and legitimate no-op/partial/compound/complementary-subject controls.
CI34044482946 and full author gates/report were pending at dispatch. Standalone
partial Push may retain unassociated remote recovery; supplied request/summary/
returned bindings must match known repository/branch facts without mutating them.
Primary resolved this under existing B5/G6/G11; no token decoding or BC change.
Known findings remain unresolved until actual re-review. No M2 or Slice closure.

API/ports#59 independently ACCEPTED atd4c3bfdd2038bbc5921488fbcffad1b8736c7460,
technical6e289c4b47a7816493f376c886c00bf0df4835e8. All M259-M01..M08 resolved;
complete primary/bounded reports/raw probes and99/77 exact blobs verified/archived
in M2-API-Review--003. Full/replay/race/copy and meaningful mutation controls pass.
Final report-only scope/identical product tree verified; sourceCI34044482946 and
finalCI34045071509 each18/explicitM3skip. Master integrates accepted API first,
verifies exact integratedCI and recordsSHA, then accepted viewmodeld279c3f. No M2
completion before both serial gates; all143 baseline findings/full13 Slices remain.

API/ports#59 COMPLETE contribution: accepted sourced4c3bfdd2038bbc5921488fbcffad1b8736c7460,
technical6e289c4, integratedecf12e91f5a1223c9d97b404db024e1ce05372df. Exact
integrationCI34046178481 passed18 applicable jobs/explicitM3skip. Master checked
all product/shared build inputs equal the accepted source; no conflicts. Review003
resolves all M259 groups. Accepted viewmodeld279c3f may now integrate serially;
M2/full Slice/native/baseline/release closure is not implied by the API leaf.

M2 COMPLETE atd60afe8fb763269964bcdc05958c8cdbc7849dea. Viewmodel#60 accepted
sourced279c3f2af59e48029e77d845f5a55271c7889cf was integrated after verified API
integrationecf12e9; its exact integrationCI34046727344 passed18/explicitM3skip.
Only accepted viewmodel files were added, with no other product/build change or
merge conflict. Issues59/60 are CLOSED, verified on GitHub. Domain6113ca5, API
finald4c3bfd/technical6e289c4 and viewmodeld279c3f are complete M2 dependencies.
All13 Slices/all143 baseline findings remain open; CLI still uses legacy stack.
M3 preparation now opens bounded adapter Issues under#21, records exact contracts/
traceability before fresh implementation, and uses separate reviewers. Accepted
serial adapter integration remains Git→GitHub→Persistence→Discovery→Runtime.
M4..M8/full native/vertical/packaging/release gates remain intact.
