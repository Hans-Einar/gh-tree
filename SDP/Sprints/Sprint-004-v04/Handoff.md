# Program handoff

State: ACTIVE - Sprint-004-v04 / I-03 / M2 parallel contract leaves.
Full long-running goal #21 remains open. Current Issues #59 (API/ports) and
#60 (State-owned viewmodel). Canonical codereview-21/refactor worktree:
C:/Users/hanse/GIT/gh-tree-wt/refactor.

## Accepted source and gates

Baseline v0.3.14: f626077ca0e59fbe9ede7ba1116982bb94b2eb6b. All ten focused
reviews and design #52/PR54 accepted/merged. Seven BCs and BoundaryTypes--001
FROZEN1.0.0 under #55/PR56, final reviewed3ac7f20cedd8d14f131b2a15a602b233ef8ca5dc,
merged97cc0d8257603766dd741b49b7d8005857b421a9. Frozen clauses remain read-only.

M1 #57 COMPLETE: accepted source/review/integration
05982a4adb35e39d7b7ba371e0d16d83b2b3674c; source CI34031420582 and canonical
CI34031833446 each18 SUCCESS/one explicit M3skip. M1-Review--001/002 preserve
failures; M1-Review--003 accepts all six corrected findings. Architecture tooling,
canonical/native ARM64/twelve-target CI and safe manual staging are integrated.

Domain #58 COMPLETE: accepted source/review/integration
6113ca55b57cb628b016c483070dbb52cd5dd79a; source CI34033181820 and canonical
CI34033772732 each18 SUCCESS/one explicit M3skip. Independent review found no
issue; M2-Domain-Review--001 archives actual constructor/mutation/compiler probes,
logs and verified source-copy hashes. Private Head/ExactTarget structs and every
API A1 Domain identity are implemented; constructors/accessors are in its README.
The CLI still uses the legacy stack. All13 Slices selected; all143 baseline
findings remain open. Full native/vertical/release proof is not supplied by leaves.

## Current contracts and dispatch

Read AGENTS, full developmentInstructions, #21/#59/#60 and their prerequisites,
UserRunContract, accepted REFDES/API/MigrationMap/Slices/Verification and all seven
frozen BCs/BoundaryTypes. Current contract:
SDP/Implementation/CR-#21/M2-Surfaces--001.md. API-Inventory--001 is a read-only
independent completeness review, not new authority. It found no material signature
or ownership contradiction:5 Client methods,28 commands,13 queries,6 events and
7 port interfaces/48 exact methods. Actual frozen signatures/fields control.

Master dispatches fresh independent workers from this kickoff commit:
- #59 branch codereview-21/layer-application-api, worktree application-api-ports:
  only new internal/application/api and ports plus tests/README and bounded
  SDP/Implementation/CR-#21/Application/M2-API--001.md report.
- #60 branch codereview-21/layer-tuistate-viewmodel, worktree state-viewmodel:
  only new internal/tuistate/viewmodel plus tests/README and bounded
  SDP/Implementation/CR-#21/TUI-State/M2-Viewmodel--001.md report.

Both worktrees are under C:/Users/hanse/GIT/gh-tree-wt. Viewmodel imports Domain
and pure values, never API/ports or parent State, so work can overlap. No other
layer, legacy, module/workflow, frozen contract/design or program metadata edits
are delegated. Master coordinates and does not implement product code.

Each worker implements its complete frozen surface with validated closed values,
immutable copies and meaningful tests, then freezes/pushes exact source. A separate
fresh reviewer reads actual source/tests/contracts, not author summaries. Correct/
re-review and require applicable current branch CI. Master serially integrates
API/ports then viewmodel and records integrated CI before declaring M2 complete.
No adapter work before required M2 surfaces; a faulty BC triggers BC-CHANGE.

## Traceability and remaining scope

CurrentIndex/Relations/Ledger through47 record accepted predecessors, M2
correction/acceptance states and bounded native tool preparation. CurrentMatrix has65 accepted checks; M1 and Domain
contributions are partial evidence for full clauses. It does not close a Slice,
Runtime allocator/exhaustion or native Git identity requirement. Original failed
assumptions/probes remain in verification archives; use exact source/hash evidence.

M3 integrates Git -> GitHub -> Persistence -> Discovery -> Runtime; M4 Application
-> State -> View; M5 complete real new-stack headless Slices; M6 host/entry cutover;
M7 mapped retirement/strict tree; M8 full verification/product PR/main/tag/twelve
assets/controlled install-upgrade. No scope reduction or premature release claim.

Accepted Runtime helper hook from M1: go run ./internal/runtime/cmd/helpergen -check
on native Windows/amd64 Go1.25.0, with real source-closure/image/compression/manifest
verification. Required committed brokerassets/broker-amd64.gz, broker-arm64.gz,
manifest.json plus broker/broker-cmd/helpergen Go sources. First Runtime source/image
makes missing inputs fail; until then conformance is explicitly NOT RUN. Pass this
actual interface to the M3 Runtime worker; no runtime compiler/download/extra asset.

## Workspace and limits

Root C:/Users/hanse/GIT/gh-tree remains main97cc0d8. Current implementation is the
canonical worktree above. M1 composition-bootstrap preserves accepted05982a4;
Domain worktree domain-implementation preserves clean accepted6113ca5. Older
review/design worktrees remain; program-checkpoints retains local801280a7073fa87c04fdc4d06b3a9fbd4d7cd7fc.
Inspect actual HEAD/dirty/unpublished state before integration; never discard it.
Use explicit UTF-8 in scripts; exact native evidence has scoped byte-preserving
attributes. No external blocker is known. Go1.25/native tools are available as
recorded in implementationNotes. Installed extension was observed at v0.3.13 and
is unchanged. No main product PR, tag, publication or install-upgrade occurred.

## Latest M2 state and exact next action

API/ports #59: CHANGES_REQUIRED at7e3dd0a3545cc259af6dfe67885b20c8ddc57749,
product3e45f313d3e953f3293b9d42405f6843264656e0. Independent primary/bounded reviews
record eight MEDIUM groups M259-M01..M08 and26 invalid admissions, with passing
copy/positive controls. Evidence: M2-API-Review--001. Product/source final CI18/
explicitM3skip does not override findings. Corrected productc96ca9244a30e16808743a0cd3e0b557e75080d7
and Master312b534 merge are frozen at4700ea72f080cae1f0a22cf21e80876d88617754,
clean at dispatch. Both independent reviewers are re-reviewing immutable source;
worker is completing the bounded report/current CI34041668920. Any product change
requires new exact review. All eight findings remain unresolved pending disposition;
author-reported passing local gates do not establish acceptance.

Viewmodel #60: independent ACCEPT atd279c3f2af59e48029e77d845f5a55271c7889cf;
M260-M01 resolved, CI34038307701 passed18/explicitM3skip. Complete independent
controls and original failure history are in M2-Viewmodel-Review--001/002. Preserve
this frozen source until accepted API/ports integrates first. Neither leaf is
integrated and neither Issue is closed. Master serially integrates accepted API,
then viewmodel, verifying the actual integrated SHA/CI and recording both gates.
Only then may M2 complete and M3 adapter implementation Issues be dispatched.

Tooling-Preparation--001 supplies only a local native signing baseline: Git2.48.1,
real SSH/OpenPGP signatures and tamper rejection in SHA-1/SHA-256 temporary repos.
All4 final profiles pass; original agent-home/header probe failures are retained.
This does not verify the actual Git adapter or G8a transparent bridge. Archive
contains scripts/logs/results/hashes only, no private keys. No CurrentMatrix check
or baseline finding is closed. Canonical04eb90a CI34039754376 and prior5e3aa2e
CI34039134247 each passed18 applicable jobs plus the explicit Runtime M3 skip.

All13 Slices/all143 baseline findings remain open. M3..M8/full native verification,
final product PR, main CI, release assets and isolated extension install/upgrade
remain required. No external blocker is known and no scope reduction is authorized.
