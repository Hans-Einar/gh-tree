# Implementation and verification notes

No product refactor has begun. No BC is frozen. No v0.4 verification is claimed.

Reconstruction verified main 3531fd165ff2b35036928328685d45d84965e9ef against
v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b: only governance/review records
changed. Initial user checkout was clean and fast-forwarded from f2e1bd05a8b123ce2234b0bce96672bbde9a6eba.

Initial PR #34 observation: one new Domain artifact, HEAD ec482795a82371996370ec52e68c5c24b4e240b1,
draft, all nine configured checks green (three platform test/vet/build jobs,
Linux race, five release-target cross-builds). These are existing baseline CI
results, not proof of future refactor behavior. Independent Domain gate review
was pending at that observation; its completed correction/acceptance follows.

Domain gate completed: fresh reviewer required an explicit unborn-HEAD boundary
policy; corrected report independently accepted at
86d2a4c0cc0eb8f427d597617246e85473a6cf55, merged via #34 at
5b1cf181fd3245a65337f1db28ee8f0fb95c200f. All nine corrected-HEAD CI jobs passed:
https://github.com/Hans-Einar/gh-tree/actions/runs/33997317388 . Report counts
remain 4 HIGH / 8 MEDIUM / 2 LOW. The report correction is resolved; its product
findings are still unimplemented. Issue #33 is closed.

Environment: Windows PowerShell; Git/GitHub CLI/Python available. Go bootstrap
located at C:/Users/hanse/.codex/tmp/ponsse-r9-go1.23.12/go/bin/go.exe; its automatic
toolchain selection obtained Go 1.25.0. Full local Windows `go test ./...` passed
in the Domain-review worktree at the corrected HEAD. Product source equals the
frozen v0.3.14 baseline. This is baseline evidence only.

Checkpoint scaffolding #35 / PR #37: accepted by a fresh independent reviewer
at 8b5514e5e69ed6de603ed5f34e6e530a6a8b76f3 and merged at
11cfef79009bb3956f0ba970470df14843e27e07. All nine jobs passed on both push and
PR CI. JSON/YAML/NDJSON and links/provenance verified. Historical wording nits
are clarified in this follow-up. The 91 finding IDs across the six accepted
reports are now explicitly indexed as open; related reports overlap, so this
count is not a claim of 91 distinct defects.

Finding-index checkpoint PR #38 accepted independently at
9f33ea31535612d6bd2d7d8d1af167db3d19585f and merged at
e0fa9c8743c0b68e936169b7997dd0c841e9dc98 with all push/PR CI green.

Release baseline read-only inventory: v0.3.14 publishes 12 assets, not merely the
five CI cross-build targets. Preserve darwin amd64/arm64; freebsd 386/amd64/arm64;
linux 386/amd64/arm/arm64; windows 386/amd64/arm64 (.exe). Exact names are in
CurrentIndex.yaml. Evidence: `gh release view v0.3.14 --repo Hans-Einar/gh-tree`
and https://github.com/Hans-Einar/gh-tree/releases/tag/v0.3.14 . Composition and
REFDES must include all targets in packaging parity and cross-build verification.
`gh extension list` reports installed gh tree v0.3.13; no upgrade performed.

Launch Discovery #36 / PR #39 accepted after fresh independent source/artifact
review at 859eab7af576b6b0eb689da6d7ed7c6d235d5853; merged at
3b8923e0fad29eec24a2d8fa1811dba24c2a626f. All nine CI jobs passed in run33998243250.
Master independently matched the four probe source hashes and reran all five
nonexecuting probes, reproducing reported baseline behavior. Report author also
ran 7 launch and 4 TUI tests. No report correction required. Its 2 HIGH / 9 MEDIUM /
1 LOW findings remain open product work; total indexed IDs now 103 (32H/58M/13L),
with overlap across reports. Next focused review: Persistence #40. No BC frozen.

G-03 checkpoint PR #41 independently accepted at
0339fd59170289ef39c247e559445789896464e4; merged at
7113e4111ff9e0e4de04f86cacc42a209c011727 with green push/PR CI.

Persistence #40 / PR #42 accepted by Master independently of the fresh report
author: complete actual report/source read, five copied hashes matched, six
nonexecuting storage probes rerun PASS, Go 1.25.0 library claims checked locally.
Report HEAD e13639926ca1b34e5021d1a038b83517cb7eec65; merge
db07857550cdfd56c114eb42c045b896b229d747. All nine CI checks passed in run
33999062918; no report correction. 2H/6M/1L product findings remain open. Total
112 report IDs (34H/64M/14L), with overlap. GitHub #43 is next. No BC frozen.

G-04 checkpoint PR #44 independently accepted at
d611e7df49877fb9865a5bd6614749ac1ff34d4e; merged at
a0ef95bae4ad27c23323b428b400c97f49b89d3a with green push/PR CI.

GitHub #43 / PR #45 independently accepted by Master at
ea52f8c5c50e178c6e415cbba2284296e2d9b910; merge
e16650036c10f764c1b99c76d98d2bd37c3ffbb0. Actual report/source/probe code read;
two copied source hashes matched; six probes plus copied baseline tests/helper
rerun PASS. Native gh 2.96.0 source matched tagged GitHub blob
1a3eba7d88261a4291e988eb3ff288c8efd409e6 and confirmed head scope/no-push behavior.
All nine CI checks passed in run 33999849098; no correction. 1H/8M/1L product
findings remain open; total 122 report IDs (35H/72M/15L), overlapping reports.
TUI View #46 is next. No product implementation or BC freeze.

G-05 checkpoint PR #47 independently accepted at
9e7fcca9d98b3bcd9e10d03ca1051d5213c623fd; merged at
61224e88f3102ca3604df7067400f61f1d1c3dff with green push/PR CI.

TUI View #46 / PR #48 independently accepted by Master at
108a3bd0ceb782c6a29d980ce0abcce1eae7a284; merge
00584a57021dda767ead5e3996cc92477e0a2d3c. Full report and source/probe code read;
all 61 copied internal Go hashes matched; seven render probes rerun PASS. All
nine CI checks passed in run 34000698430. No report correction; 0H/11M/1L
product findings remain open. Total 134 report IDs (35H/83M/16L), overlapping
reports. Composition #49 is next; all nine preceding layers accepted. No BC frozen.

G-06 checkpoint PR #50 independently accepted at
4d15a31e287f9b1757c8ae76716f477f56feaf1e; merged at
c9ae9851fdd789c13e8da4ebf4b34c603b724987. Verified CI runs 34000926362 (PR) and
34000895523 (push) succeeded; initial comment run-number references were corrected
transparently before merge and are not verification evidence.

Composition #49 / PR #51 independently accepted by Master at
e2ffc1e72c06270bbaf2a20f2c21b949a9b3985c; merge
e0e716fd44421c221d00375fc0c0d0b8ea85bdff. Complete actual report/root/workflow
source read; pinned action script Git blob verified, Bubble Tea shutdown source
inspected, all twelve local build files/Windows metadata inspected. Master rebuilt
bootstrap independently and reproduced all six native pre-UI probes. All nine CI
checks passed in run 34001756011; no report correction. 1H/6M/2L product findings
remain open. Total 143 report IDs (36H/89M/18L), overlapping reports.

All ten focused reviews are accepted/merged; I-01 is complete. Refactor Design
Issue #52 authorizes I-02/G-07. No product refactor or v0.4 verification has begun,
and no BC is frozen. Required design resolves APIs, physical ownership, full Slice
and compatibility plan, verification and safe twelve-target release gates.

I-02 checkpoint #53 merged reviewed source
d62d112bd897f63dd14b186356335b3d65170b33 at
58f1cb9eda941db0941cbb8e04e6a0559a3ca896. Remote PR metadata rechecked:
CI runs 34002140473/34002264047 have all 18 checks successful.

Design #52: Master reread all accepted reports afresh. Bounded separate reviewers
examined Git/Runtime feasibility, actual migration ownership and early safety.
Early design findings DES52-H01/H02 reject precheck-only live Git overwrites and
numeric Unix signal revalidation. Detailed correction/evidence state is in
SDP/Design/CR-#21/DesignReview--001.md and its feasibility appendices.
Normative drafts now cover API/ports, all 70 old physical files, 90 baseline test
functions, 13 vertical capabilities, the complete platform/verification/release
contract and all 143 planned finding dispositions. Every product finding remains
open; all implementation/review/integration evidence fields remain empty.

Master has made governance/design edits only. No product implementation, BC
freeze, product integration branch, tag or release mutation has occurred. Final
design review must inspect an exact complete committed HEAD after both safety
protocols are reconciled; bounded probe results are not v0.4 product verification.

PR #54 initial design HEAD f1dec60cbd3892f068410f2d7f3caa7f855ba52e passed
all 18 CI checks in runs 34005955138/34005970811. Fresh full independent review
returned 2 HIGH/2 MEDIUM design findings (DES52-H03/H04/M01/M02), while accepting
the earlier H01/H02 mechanism directions. DesignReview--002 preserves the findings.
Master added selected native Persistence protocol/schema migration, cwd identity
through child acquisition, missing exact PR/stash reads and bounded reliable
Runtime-to-App cleanup transfer. Separate Persistence feasibility source/logs and
Root native Windows/Linux cwd experiments are archived. Corrected exact HEAD
review/CI is next; no product finding closed, no BC frozen, no product code changed.
