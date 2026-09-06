# Program handoff

ACTIVE: Sprint-004-v04 / I-03 / M3. Explicit user resumption supersedes the pause.
Repo Hans-Einar/gh-tree; canonical worktree C:/Users/hanse/GIT/gh-tree-wt/refactor,
branch codereview-21/refactor. Follow the resumed policy in UserRunContract.md:
coherent verified commits pushed promptly, targeted tests, strong independent
critical review, concise traceability and no duplicate evidence archives.

## Verified baseline

Recovery commit3e259652797295504e34bafe2063a8022a503965 and annotated tag
checkpoint/gh-tree-m3-preparation/20260906-1725Z matched local/origin at resume.
All16 worktrees were clean and remote heads matched the checkpoint inventory;
M3 Issue bodies were unchanged. Main remains97cc0d8. Historical checkpoint801280a
is already pushed on codex/cr21-checkpoints. Earlier pause/inventory details remain
in the immutable recovery tag and existing checkpoint evidence archive.

M0: ten focused reviews, accepted design #52/PR54 and seven BCs/shared types
FROZEN1.0.0 (#55/PR56, effective97cc0d8). M1 #57 accepted05982a4. M2 #58 Domain6113ca5,
#59 API finald4c3bfd/technical6e289c4 and #60 viewmodeld279c3f are accepted and closed.
API integrationecf12e9 CI34046178481 and combined M2 integrationd60afe8 CI34046727344
passed18 applicable jobs plus explicit pre-Runtime helper skip. Checkpoint3e25965
CI34050258198 also passed. Full review histories/evidence are already indexed;
do not recopy them. M259/M260 implementation findings are resolved.

## Current work

M3 Issues61 Git,62 GitHub,63 Persistence,64 Discovery,65 Runtime remain the exact
work contracts. Read M3-Adapters--001.md, M3-Assignments.json, the complete Issue/#21,
frozen referenced authority and actual M2 code. First authors: Git/GitHub/Persistence;
Discovery/Runtime queued. All five worker branches/worktrees are created/pushed at
base412f33e477cec03cb6eafe7b846c9bcdd02c0a25; exact paths are in M3-Assignments. Workers
own only their adapter folder/tests/bounded report; no legacy, shared workflow/module
or frozen-contract edits without Master coordination. Separate reviewers inspect
frozen candidates; serial integration is Git→GitHub→Persistence→Discovery→Runtime.

Current pushed milestones: Gitde9b2c8 (nine read methods/native foundations,
no mutations); GitHub technical7382617/reviewccbc843 independently ACCEPTED,
CI34055294036 passes18/helper skip, held for Git-first integration. Persistence
c146bb5 has private native publication primitives/Windows metadata; latest verified
CI34055865289 at0630ded passes18/helper skip. Public Storage/Unix metadata/recovery
remain in progress. #66 FreeBSD worker branch codex/cr21-freebsd-ci is ready at
base22217f9, worktree C:/Users/hanse/GIT/gh-tree-wt/freebsd-ci.

#67 API correction technical675dfff/review0bb8f28 is independently ACCEPTED,
integrated/pushed4d1b5486a029eeabcd874306d7a7ece4ae629ff3. Source CI34055882003 and
integrated CI34056365373 pass18/helper skip; exact local integration race/vet/arch
pass. Git may adopt that exact prerequisite at a clean boundary and resume status.
Git contract1.1.0 froze at10f687e; #67 is complete as a bounded prerequisite.
#68 queues a small ports PlanSummaryDigest accessor needed by already-frozen
approval semantics; it does not change the contract. #66 CI author is running.
No M3 adapter is integrated and no full Slice closed. Authority: M3-Adapters--001.

Local-only test residue: Persistence report atc146bb5 records the owned Windows
metadata fixture TestWindowsNativeMetadataCopy2880129540/001 under the user's
Local/Temp directory; automatic approval review rejected its ACL restoration/
removal as blocked by policy. Preserve it; do not retry/bypass. It is test data,
not unpushed product source. Current later tests clean through retained handles.

No M3 product work existed at resume. Before the next substantial step ensure the
previous milestone is verified, committed and pushed. A partial recovery checkpoint
is not completion. Canonical docs-only commits may skip CI; executable/build changes
and meaningful acceptance/integration gates keep the required verification.

## Required continuation

All13 Slices/all143 baseline findings remain selected/open. CLI still uses legacy.
After M3: M4 Application/State/View, M5 real new-stack harness, M6 host/cutover,
M7 mapped retirement/strict tree, M8 full native/platform/E2E/product PR/main/
twelve assets/controlled install-upgrade/release gates. No scope reduction or
completion inferred from adapter folders. Original conditional release authority
remains governed by the complete verification contract.

Runtime helper interface: go run ./internal/runtime/cmd/helpergen -check on native
Windows/amd64 Go1.25.0; brokerassets/broker-amd64.gz, broker-arm64.gz and manifest.json,
with broker/ including broker/cmd/ and cmd/helpergen/ source closure. First Runtime
input makes missing prerequisites fail. No runtime compiler/download/extra asset.
Tooling-Preparation--001 and native feasibility are inputs, not product proof.
FreeBSD read-only action source is preserved in checkpoint evidence, not executed
or accepted CI. Shared native CI requirements are Master-coordinated.

Go1.25 executable: C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe.
Use explicit UTF-8 and task workdirs. Original temporary fixtures/keys/tools/caches
remain local, inventoried at checkpoint; no unpushed product work at resume.
No technical external blocker is known. Consult CurrentIndex and ledger for exact
new source/dispatch/integration SHAs rather than narrative history.
