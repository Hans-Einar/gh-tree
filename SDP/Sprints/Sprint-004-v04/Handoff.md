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

## Published contribution checkpoints

Worker source, reports and detailed evidence live on the named branches. Use
CurrentIndex plus git show <HEAD>:<report-path> when a report is not on canonical.
Worktree directory names below are under C:/Users/hanse/GIT/gh-tree-wt/.

| Contribution / worktree | Pushed branch | Exact HEAD | Disposition / remaining gate |
|---|---|---|---|
| Git#61 / git-implementation | codereview-21/layer-git | 25240dd2acc40cb93003674844c248cdceac5920 | 11 facts/registry/locks; CI34063202321 19PASS; required review blocked, mutations incomplete |
| GitHub#62 / github-implementation | codereview-21/layer-github | ccbc8430618c24fd2e490d7f7c6565063dd35bbe | technical7382617 independently accepted, held for Git-first integration |
| Persistence#63 / persistence-implementation | codereview-21/layer-persistence | 53de3e1daea2c7ffd0fe6c6bdf365a8c2160357f | bounded Windows identity correction bfdc79c accepted; full adapter/native gates open |
| Discovery#64 / discovery-implementation | codereview-21/layer-launchdiscovery | 4e9f4793de6b438aec0a0d5e77e37631a829a1d6 | technical20bd8ac accepted; CI34065291612 19PASS; held for prior serial gates |
| Unix Runtime#65 / runtime-implementation | codereview-21/layer-runtime | d0a7682e761a7c076feeaf2e4d1ab0476edf3ae7 | corrected source71b95a3/product4df4385; CI34076324950 6PASS; M65-U01/U02 confirmation pending |
| Windows Runtime#69 / runtime-windows | codex/cr21-runtime-windows | 9b06f8bd0e8bc4baaaabb7dccd1054f38d74fec8 | standalone bd78dea independently accepted; committed helper binding pending |
| Helper#70 / runtime-helper-assets | codex/cr21-runtime-helper-assets | 6f385a9cab6798661970c4dc2a0aa56edba6c97b | b6f161f generator/assets/policy accepted, CI34075549875 20PASS; images bind older ab608327 |
| Parent#71 / runtime-sessions | codex/cr21-runtime-sessions | 78a67e1198466e05bc5f2544b3c4a495dc45a52c | common engine fff5ee4,12 methods/unexported test-owner construction; fresh review and real bridges pending |

Standalone Windows recovery tag:
checkpoint/m3-windows-native-reviewed/20260907-0206Z at review9b06f8bd.
No M3 adapter is integrated into canonical yet. Component acceptance is not full
Runtime or vertical Slice acceptance.

## Verification and open work

- Windows W69-M01/M02 independently resolved: canceled admission, retained bounded
  delivery receipts and closed native cause/stage facts. Actual amd64/386 controls,
  native ARM64, race/vet and12-target builds pass. CI34073256923 has6 native/race
  successes; only missing helper inventory fails, dependent cross/helper jobs skip.
- Helper H70-H01/M02/M03 independently resolved: captured/guarded input bytes,
  one-shot native name-change invalidation through actual compilation, fresh-cache
  admission and canonical Go-root junctions. Real changed-source PE compilations
  are rejected; exact checker preserves all60 Runtime+pin files. Accepted images
  still derive from ab608327e63727f66ffb1aa7b3200c2865307cf5. Regenerate and test
  actual committed images against reviewed Windows source before native binding.
- Unix M65-U01/U02 author corrections add actual anonymous endpoint profiles and
  startup-bound cleanup periods. Product defaults remain2s/3s. Original named-FIFO,
  natural-exit-budget and malformed FreeBSD fixture failures are retained. Native
  Linux/macOS/FreeBSD and race pass at71b95a3; independent confirmation remains.
- Parent common-engine tests cover lifecycle/cancel, late output/finals/ACK,
 64KiB split/pending receipts, restart, Shutdown, cleanup repair and counter limits.
  Root race/vet and all12 architecture/build checks pass; source CI34076409741
  has6 successes and expected helper prerequisite failure. Controlled owners are
  common-state evidence only; there is no production constructor or native stub.
- Persistence early protocol findings are independently corrected at91acac5;
  Windows repeated-publication identity is independently accepted at bfdc79c.
  Native amd64/386/ARM64 and Linux confirmation pass. CI34066675310 still fails
  native FreeBSD (18 successes/helper skip). Full fault/resource, no-birth
  association and Unix source-name/check-to-syscall interval gates remain open.

Detailed source/test/CI records are the existing M3 reports and review files;
reuse them rather than duplicate histories or infer review from author summaries.

## Blockers and exact next actions

The required Git foundation review failed with a cybersecurity service/access
block at source855a144; no completed accepted review exists. Its already-delivered
provisional findings are in M3-Foundation-Review--001. Do not retrieve withheld
output, retry/relabel that review, switch models to bypass controls or waive it.
Resolve approved review access/false positive first. Official guidance:
https://learn.chatgpt.com/docs/cyber-safety . Independent M3 work may continue.

FreeBSD Persistence requires a user trust/scope decision after the independent
M3-FreeBSD-Assessment--001: ordinary ZFS cannot inspect arbitrary protected system
EAs through the selected profile. No privileged helper, metadata-loss/read-only
substitute or platform reduction is approved. The pending user choice is to retain
full requirements and design a privileged mechanism, explicitly revise the profile,
or keep that part blocked. Elapsed time is not an answer.

Next Master actions: request bounded Unix correction confirmation and fresh parent
common-engine review; define a bounded helper-binding contribution, combine only
reviewed Windows/helper sources in its own worktree, regenerate exact assets and
verify real committed-image execution. Then integrate accepted native work into
#71 and implement actual platform bridges/production construction with full native
verification and independent review. No unreviewed merge or placeholder shortcut.

Canonical adapter order stays Git → GitHub → Persistence → Discovery → Runtime.
Then M4 Application/State/View, M5 real harness, M6 host/cutover, M7 mapped retirement,
M8 native/E2E, product PR/main and release only with explicit authority/all gates.
The full goal continues automatically; this recovery point does not cancel it.

## Local-only state and working rules

No unpublished product work or uncommitted implementation exists at this snapshot.
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

## Work after the recovery checkpoint

#72 has separate pushed preparation7e19d617ae430cd2c7ad414b9ea4622eb1e00617 on
codex/cr21-runtime-helper-binding, worktree runtime-helper-binding. It combines
reviewed Windows/helper sources; payloads are still ab608-based until its worker
regenerates and proves actual committed-image binding. Only its exact root test
file is excepted from #71 ownership. No canonical product integration occurred.
Unix corrected d0a7682 is queued for existing reviewer confirmation; parent common
78a67e1 is queued for fresh independent review. This work follows the clean29-tree
recovery inventory, so inspect the now-active worktrees instead of reusing that
historical clean-state assertion.

Latest follow-up: Unix/shared native review f85294d independently accepts corrected
71b95a3/product4df4385. Binding #72 freezes b9def1f (technical cf16072), CI34078800342
20PASS, for fresh review of actual committed-image/source execution. Parent review
c436773 requires an eviction-triggered Restart panic fix, Shutdown reobservation
and truthful geometry/sequence-boundary handling; root-only correction follows.
Persistence #63 also resumes concrete remaining fault/resource coverage from
53de3e1 while all explicit platform/trust blockers stay open. CurrentIndex holds
exact refs; prior recovery inventory remains a historical clean snapshot.

Native integration510139d18da16a0d42c62ee88fba8053813f2b13 is pushed on
codereview-21/layer-runtime: reviewed Unix plus accepted binding f926195f (Windows
and helper sources/images). Local Windows race, all12 architecture, vet and
unchanged955-input helper rebuild pass; exact integrated native CI is pending.
Canonical program still has no M3 product integration. Parent corrected
f4e48b3/technical4495635 is frozen for independent confirmation of M71 findings.

Current rework gates: parent review15eed03 resolves H01/M01 but keeps M02 for an
established-start sequence/cwd publication panic. Native integration510139d
CI34080586354 has19 successes and one FreeBSD blocked-output fixture failure;
actual final fact is clean/error nil, so its historical-timeout synchronization
is being checked without weakening native ownership. Persistence sourceff40e32/
report3772faf awaits fresh bounded fault/resource review; CI18PASS/known FreeBSD
failure. CurrentIndex and separate branch reports contain exact refs.
