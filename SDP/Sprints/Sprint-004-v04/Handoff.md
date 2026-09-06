# Program handoff

State: PAUSED BY USER - Sprint-004-v04 / I-03 / M3 before dispatch.
Full long-running goal #21 remains open. M1/M2 Issues #57..60 are complete/closed.
Canonical codereview-21/refactor worktree:
C:/Users/hanse/GIT/gh-tree-wt/refactor.

## User-requested checkpoint

Work is stopped. Resume requires an explicit instruction from the user. Do not
start M3 implementation, create its worktrees/agents, or perform a merge as part
of this checkpoint. The full original objective is unfinished and unchanged.

Recovery tag: `checkpoint/gh-tree-m3-preparation/20260906-1725Z` (annotated).
It is a recovery marker, not a claim that M3, a Slice or v0.4.0 is complete.
The tag's peeled commit identifies the exact checkpoint HEAD. This record cannot
embed its own commit hash; use Git metadata and the final Issue checkpoint comment.

Repository: Hans-Einar/gh-tree, origin https://github.com/Hans-Einar/gh-tree.git.
Worktree: C:/Users/hanse/GIT/gh-tree-wt/refactor.
Dedicated branch: codereview-21/refactor; no new branch or switch was needed.
Default worktree C:/Users/hanse/GIT/gh-tree remains clean main97cc0d8.
Pre-checkpoint HEAD was3cf264141764385eb6afe3f7e4e25a7504020690, already pushed.
The exact pending M3 preparation was first preserved without simplification in
5cdd4367af786f0f2a18e500eb621c6867939f63. This following checkpoint preserves the
pause status and recovery inventory. Both commits are included in the tag.

All16 registered worktrees were inspected. Only this task worktree was dirty:
8 unstaged SDP files and2 untracked M3 contract/assignment files, nothing staged
and no product edits. Their original bytes/hashes and tracked patch are in
SDP/Verification/CR-#21/Evidence/Checkpoint-20260906-1725Z. Its ArchiveManifest.json
maps exact archived files. workspace-before.json records every worktree/ref/status,
89 commits captured across refs since resumption and72 current ancestors since the
product baseline. Git history retains the complete commits.

The previously local-only older merge801280a7073fa87c04fdc4d06b3a9fbd4d7cd7fc was
pushed on its existing codex/cr21-checkpoints branch. Its tree equals previously
published historical main58f1cb9eda941db0941cbb8e04e6a0559a3ca896. No new merge was
performed. That historical branch is separate from this tag's ancestry; recover
it from origin by its recorded ref/SHA when needed, without merging it.

Completed: ten focused reviews, accepted design, seven frozen BCs/shared types,
M1 CI/architecture bootstrap, and independently reviewed/integrated M2 Domain,
API/ports and viewmodel. Verified combined productd60afe8fb763269964bcdc05958c8cdbc7849dea
passed integrationCI34046727344; API integrationecf12e9 passedCI34046178481.
Latest pre-checkpoint metadata3cf2641 passedCI34047530046. Each run has18 applicable
SUCCESS jobs and an explicit pre-Runtime M3 skip. Earlier M1/M259/M260 failures
were corrected; original failures and deliberately failing mutation controls are
preserved in their review archives. No new product tests are claimed for this
documentation-only checkpoint; JSON/hash/scope checks verify preservation.

Partial: M3 Issues61..65 and complete contracts exist and remain open, paused.
No M3 branch, worktree or implementation agent was created. All prior agents
were confirmed completed. Full Issue bodies/comments, all M3 drafts and read-only
FreeBSD source/provenance/license are archived. The FreeBSD action was not run or
accepted as infrastructure. Native signing experiments are tooling evidence only.

Remaining: M3 adapters; M4 Application/State/View; M5 real new-stack harness;
M6 host/cutover; M7 mapped retirement; M8 full native/platform verification,
product PR/main/twelve assets/controlled install-upgrade/release. All13 Slices and
143 baseline findings remain selected/open. CLI still uses legacy behavior.
No technical external blocker is known; the current dependency is user resumption.

Local state outside the tag remains intact: generated test repos/review clones,
compiled binaries, temporary signing keys, downloaded tools, Go caches and local
worktree registrations/reflogs. local-only-roots.json lists24 task-named roots,
20555 files/about420MB and4 reparse entries that were not followed, with no read
errors. Necessary accepted evidence is already in Git; this archive adds exact
M3 preparation and original pending documents. Temporary keys/caches are not
project source and are not published. Installed tools/credentials are external;
no credentials were copied. There is no uncommitted local product work.

If instructed to resume, first verify the tag/branch and read this Handoff,
CurrentIndex, ledger55 and full Issues21/61..65 plus frozen authority. Update the
pause/dispatch records, create the planned worktrees from the verified current
checkpoint and record exact bases. Then start fresh Git/GitHub/Persistence workers
with separate reviewers; Discovery/Runtime remain queued. Serial integration is
Git -> GitHub -> Persistence -> Discovery -> Runtime. No implementation follows
the checkpoint without that user instruction.

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

## Current stage and exact next action

M2 is COMPLETE. API/ports#59 finald4c3bfdd2038bbc5921488fbcffad1b8736c7460
(technical6e289c4) is independently ACCEPTED in M2-API-Review--003. All eight M259
groups resolved;99 primary/77 bounded raw blobs and53 archived evidence files
verified. SourceCI34044482946/finalCI34045071509 each18/explicitM3skip. Actual API
integrationecf12e91f5a1223c9d97b404db024e1ce05372df passed integrationCI34046178481.

Viewmodel#60 sourced279c3f2af59e48029e77d845f5a55271c7889cf is independently
ACCEPTED in M2-Viewmodel-Review--002; M260-M01 resolved. SourceCI34038307701 was
rechecked18/skip. After API passed, viewmodel integrated at
**d60afe8fb763269964bcdc05958c8cdbc7849dea**, integrationCI34046727344 passed18/skip.
Product trees match accepted sources and no unexpected shared/build inputs changed.
API#59 and viewmodel#60 are CLOSED, verified on GitHub. M1/M2 are complete; no
full Slice, baseline finding, coordinator/native behavior or release gate is closed.

Active stage: I-03 / M3 adapter implementation. Read AGENTS/developmentInstructions, current
full#21, accepted REFDES/API/Storage/CwdAcquisition/WindowsBroker/MigrationMap/
Slices/Verification/FindingDisposition, relevant full frozen BCs/shared types and
actual accepted Domain/API/ports. M3 Issues61/62/63/64/65 and Master contract
SDP/Implementation/CR-#21/M3-Adapters--001.md are created. Index/relations/ledger54
record scope before dispatch. Only after explicit user resumption may Master
create the worktrees, record their actual base and start fresh Git/GitHub/
Persistence authors; Discovery/Runtime remain queued. Every worker owns only its
one declared adapter folder/tests/report; separate fresh reviewers and Master
serial integration remain required. No M3 product author has started yet at this
kickoff; exact base/dispatch follow in the authoritative records.

M3 serial integration: Git → GitHub → Persistence → Launch Discovery → Runtime.
Then M4 Application → State → View; M5 full real new-stack headless Slices; M6
host/entry cutover; M7 mapped retirement/strict empty-allowance tree; M8 full
verification/product PR/main/tag/twelve assets/controlled extension install-upgrade.
All13 Slices and143 baseline findings remain selected/open. No scope reduction.

Draft work descriptions (proposals only, not implementation authority) are in
C:/Users/hanse/.codex/tmp/gh-tree-m3-preparation: Git-Issue-Draft-v2.md,
GitHub-Issue-Draft.md, Persistence-Issue-Draft-v2.md, Discovery-Issue-Draft.md and
Runtime-Issue-Draft.md. Git draft received independent planning ACCEPT with G7
scope locking, capability-specific2.43/2.48 handling and contribution-versus-full-
stack evidence clarifications. Actual frozen contracts and current M2 source win.

## Native preparation and required limits

Tooling-Preparation--001 records bounded local Windows Git2.48.1 real SSH/OpenPGP
signatures and tamper rejection in SHA-1/SHA-256;4 profiles pass. It is tool setup
evidence only, not actual Git adapter/G8a bridge/native publication proof. All
failed environment/header fixture attempts are preserved; keys remain outside repo.
Missing-parent Storage interpretation: original Expected anchor stays in native
manifest, while shared recovery Original may describe guarded absence after actual
parent establishment. Unknown standalone Push recovery is retained without asserting
association; known request/summary/returned bindings reject conflicting facts.
These are existing-contract interpretations, not BC changes or native acceptance.

Accepted M1 Runtime hook: go run ./internal/runtime/cmd/helpergen -check on native
Windows/amd64 Go1.25.0. Required brokerassets/broker-amd64.gz, broker-arm64.gz,
manifest.json plus broker/broker-cmd/helpergen source closure; real independent
rebuild/PE/hash/compression/source verification, no checkout rewrite. First Runtime
source/image makes missing inputs fail; until then the job is explicitly skipped.
No runtime compiler/download/extra public release asset is authorized.

Read-only FreeBSD CI preparation inspected vmactions/freebsd-vm v1.5.5 at
f0552d3b69211736abd97f02ff3d4674c56b73b1, with15.0/14.3 configs using AnyVM0.6.5/
builder2.2.6. Local source snapshot:
C:/Users/hanse/AppData/Local/Temp/gh-tree-freebsd-ci-source-b72818e2fe6f48bfb37b757e9735968c.
No guest was executed and no source/runtime/image dependency is accepted merely
from this read. Future native CI must be reviewed/pinned and record actual guest
OS/arch/toolchain/source/tests/artifacts, not cross-build claims or full env dumps.

## Workspace and continuity

Canonical worktree C:/Users/hanse/GIT/gh-tree-wt/refactor now contains both accepted
leaves and latest verified product integrationd60afe8. Main remains97cc0d8 and CLI
still uses the legacy stack. API worker preserves cleand4c3bfd; viewmodel worker
preserves cleand279c3f; Domain6113ca5 and M1 composition05982a4 are unchanged.
Older worktrees remain; program-checkpoints retains local801280a7073fa87c04fdc4d06b3a9fbd4d7cd7fc.
Inspect actual HEAD/dirty/unpublished state before each next action. No product
PR/tag/release/install-upgrade has occurred; installed extension was observedv0.3.13
and is unchanged. No external blocker is known. Use explicit UTF-8 for scripts and
scoped byte-preserving attributes for native evidence. CurrentIndex/Relations/
Ledger through55 and CurrentMatrix65 checks record M2 completion; no whole check
is manufactured complete from leaf-only proof.
