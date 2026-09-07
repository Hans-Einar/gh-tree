# Program handoff — M3 prebinding checkpoint

Sprint-004-v04 / I-03 / M3. Full Issue #21 v0.4 objective remains active.
Repository: Hans-Einar/gh-tree. Canonical worktree:
C:/Users/hanse/GIT/gh-tree-wt/refactor, branch codereview-21/refactor.
Recovery tag: checkpoint/m3-runtime-prebinding/20260907-0247Z (checkpoint d7a62935; subsequent work may be newer).
This is a recovery marker, not adapter, Slice, stage or release completion.

## Authoritative state

M0/M1/M2 and shared prerequisites #66/#67/#68 are accepted and integrated.
Last integrated executable HEAD is d5eeaac4acb11894bbb2209aa7e04104ee634f1d,
CI34058149301: 19 successful jobs and expected pre-Runtime helper skip.
Canonical changes since then are SDP/status/evidence only. Main remains
97cc0d8257603766dd741b49b7d8005857b421a9; public CLI/install still uses legacy.
Git BC1.1.0 froze at10f687ea7df4cde1746b50ef4b5537ecdeb13c39;
other frozen contracts/shared types remain1.0.0. All13 selected Slices and143
baseline findings remain in scope; none closes from these adapter foundations.

The original resume checkpoint3e259652797295504e34bafe2063a8022a503965 and tag
checkpoint/gh-tree-m3-preparation/20260906-1725Z were rechecked unchanged.
At this checkpoint, all29 registered worktrees were clean, with no stashes and
no local branch commits absent from origin. Existing inventory format is reused
at SDP/Verification/CR-#21/Evidence/Checkpoint-20260907-0242Z/workspace-before.json.
That inventory precedes this status commit; it records all exact worktree/ref
states and post-resume commit history. All authors/reviewers were stopped at
published boundaries. Before modifying a resumed lane, inspect its actual state.

## Current published contributions

Paths are under C:/Users/hanse/GIT/gh-tree-wt/. Exact reports/evidence are on the
named branches and indexed in CurrentIndex; historical evidence is not repeated.

| Contribution/worktree | Branch | Latest pushed HEAD | Disposition |
|---|---|---|---|
| Git/git-implementation | codereview-21/layer-git | 25240dd2acc40cb93003674844c248cdceac5920 | Facts foundation only; required review blocked, mutation methods incomplete |
| GitHub/github-implementation | codereview-21/layer-github | ccbc8430618c24fd2e490d7f7c6565063dd35bbe | All5 methods independently accepted; held for Git-first integration |
| Persistence/persistence-implementation | codereview-21/layer-persistence | 82610b891a82e7b12eb0d13d6be82bc15342f1c4 | Identity and bounded protocol/fault-resource deltas accepted; full adapter gates open |
| Discovery/discovery-implementation | codereview-21/layer-launchdiscovery | 4e9f4793de6b438aec0a0d5e77e37631a829a1d6 | All2 methods independently accepted; prior serial gates held |
| Native Runtime/runtime-implementation | codereview-21/layer-runtime | fd3abeac967a695bcdad15918a723ff8bab751df | Reviewed native/helper components combined; integrated Linux gate failing |
| Windows/runtime-windows | codex/cr21-runtime-windows | 9b06f8bd0e8bc4baaaabb7dccd1054f38d74fec8 | Native sourcebd78dea independently accepted |
| Helper/runtime-helper-assets | codex/cr21-runtime-helper-assets | 6f385a9cab6798661970c4dc2a0aa56edba6c97b | Generator/assets/policy independently accepted |
| Parent/runtime-sessions | codex/cr21-runtime-sessions | 01641986c702b9c6a1abd176586b4d58b39f13fa | Common12-method engine accepted; production bridges pending |
| Binding/runtime-helper-binding | codex/cr21-runtime-helper-binding | f926195ffebc5b99921f6b0127bbe23ad7555a5a | Actual committed helpers independently accepted; integrated into native branch |

No M3 adapter has been integrated into canonical. Common-parent/native component
acceptance does not close a full Runtime adapter or vertical Slice.

## Verification and exact next actions

- Native integration510139d combines Unix reviewf85294d and binding reviewf926195f
  (including Windows/helper). Product closure remains unchanged through test2bc27e0.
  CI34082980191:19PASS, Linux FAIL. Both Windows native architectures, FreeBSD,
  macOS/race, exact955-input helper reproduction and all12 builds pass. Linux
  TestNativeUnixClientFailedStartPreservesTypedFacts returned code12/stage2 instead
  of NotFound; fresh Unix worker investigates before further integration.
- Prior FreeBSD blocked-output fixture now observes actual join-timeout fact before
  releasing its callback. Windows extraction fixture now proves full native cleanup
  and admits only the exact repaired historical STATUS_CANNOT_DELETE; deterministic
  owned SEC_IMAGE refusal/release control confirms it. Original failures remain in
  reports; neither test correction changed product code or waived cleanup.
- Parent technical5a14f7d/review01641986 resolves all M71 findings (Restart retention,
  Shutdown repair, sequence/startup publication). Controlled-owner race/vet/all12
  tests pass; CI34081419821 has6 successes and expected missing-helper prerequisite
  failure/dependent skips. After native gate passes, Master integrates reviewed
  native sources into this branch; fresh #71 worker supplies real platform bridges,
  production constructor and early private entry, then native/independent review.
- Persistence technicalff40e32/review5380629 has bounded admission/fault/resource
  ACCEPT; earlier identity/protocol corrections remain accepted. CI34080421099:
  18PASS/native FreeBSDFAIL/helperSKIP. Remaining assessment82610b8 distinguishes
  explicit proven write-profile selection from the unresolved source-name binding
  interval. No no-birth mechanism or boundary waiver is claimed.

## Blockers and authority boundaries

Git foundation review failed with a cybersecurity service/access block. No accepted
review exists. Do not retrieve withheld output, retry/relabel/substitute the blocked
review, switch model to bypass or waive it. Resolve approved access/false positive.
Official guidance: https://learn.chatgpt.com/docs/cyber-safety . Separately authorized
independent M3 work continues.

FreeBSD ordinary ZFS cannot inspect protected system EAs through the selected
profile. User decision remains pending: retain full requirements and design an
explicit privileged mechanism, explicitly revise the profile, or keep it blocked.
No privileged helper, metadata-loss/read-only substitute or platform reduction is
approved. The existing M3-FreeBSD-Assessment--001 is authoritative evidence.

Persistence's remaining source-publication interval can consume a substituted
source after the last check. Frozen authority permits a target-editor race but no
source-namespace exception. Master must identify a demonstrated binding mechanism
or obtain an explicit reviewed boundary decision; neither is adopted. Draft
BC-CHANGE-PER-SOURCE-01 in M3-Adapters defines preparation-source cooperative
ownership, awaits fresh independent review, then user decision. It changes no
frozen authority. No-birth
write refusal may be scoped to proven profiles only if no accepted capability is
removed. See M3-Remaining-Gates-Assessment--001 for exact clauses and options.

Canonical order remains Git → GitHub → Persistence → Discovery → Runtime, then M4
Application/State/View, M5 real harness, M6 host/cutover, M7 mapped retirement, M8
native/E2E, product PR/main and release only with explicit authority/all gates.
The full objective remains active. No CLI/legacy cutover has occurred.

## Local-only state and working rules

The latest author handoffs listed above were clean and pushed. Inspect active lanes
before reuse; the older29-worktree inventory is historical, not current live status.
Ordinary compiler/module caches, owned native fixtures/binaries and WSL /tmp runs
remain local; material reviewer sources/logs are preserved in existing SDP evidence.
No broad source/binary/cache archive is required. Two detached review worktrees
(Git855a144, Persistence91acac5) remain reachable through published histories.

Automatic approval review rejected cleanup as "blocked by policy" for these owned
local residues; no retry/bypass is authorized:
- C:/Users/hanse/AppData/Local/Temp/TestWindowsNativeMetadataCopy2880129540/001
  (copy-d/source-d denied-ACL fixture).
- C:/Users/hanse/AppData/Local/Temp/gh-tree-helper-fresh-43f6e3a638ca4bf5afc5ec390cb2f542
- C:/Users/hanse/AppData/Local/Temp/gh-tree-helper-check-fresh-krknen5u
- C:/Users/hanse/AppData/Local/Temp/gh-tree-helper-final-xdx2pbor
The helper toolchain junction was removed before the cache-cleanup rejection.
These residues are reproducible test/cache state, not missing source/evidence.

Native Go1.25.0:
C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe.
Use explicit UTF-8 including subprocess decoding, exact workdirs and existing
scope/traceability. Test → coherent commit → push before the next substantial step.
Use [skip ci] for documentation-only commits. Master does not implement product
code; separate workers author and separate reviewers verify. Preserve all work.
