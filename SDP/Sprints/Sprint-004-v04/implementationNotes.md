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

Re-review of8c109b1e4308eb301d7c1418292044d714c2f33b accepted design H03/M01/M02
but retained H04 HIGH. Native FSCTL reparse worked through metadata/share guards;
CreateProcess returned before actual cwd acquisition. Strong data-read anchors
and a target-specific initial-breakpoint/FileIdInfo barrier were then proved.
Direct386-to-native64 debug creation fails50, so the complete proposed correction
uses an embedded native Runtime broker, preserving capability and12 public assets.
Actual386 -> extracted amd64 broker -> amd64 user/descendant with ConPTY, input/
resize, normal/forced outer Job0 and deterministic helper rebuild passed bounded
native probes. Native ARM64 implementation/emulation remains required, now with
the documented windows-11-arm hosted CI route. Full source/log reports are archived;
WindowsBroker--001 specifies the final topology/protocol/build contract. Fresh
exact-HEAD review must accept this correction before any design/BC/product gate.

Fresh independent complete design ACCEPT at
664f0c051344e3abdfd7d3c5698e4fbd3f584a83; H04 and prior design findings resolved
at design level. All18 CI checks pass in runs34022527812/34022530503. Master posted
technical acceptance on #52 and recorded DesignReview--004. Acceptance labels
become effective after final metadata exact-HEAD review/CI and PR54 merge. The next
authorized program gate is separate BC review/freeze; product findings stay open.

G-07 COMPLETE: independent final metadata ACCEPT at
e63547771d89f892ebe8e13f66a1ed8ef0807ba2, final18 CI checks SUCCESS in
34022968145/34022971176, expected-HEAD merge PR54 produced
4a42222f7bfedc1d80693effbb25a1a82fcff65e. #52 closed; main fast-forwarded cleanly.
Master recorded completion on #21 and opened #55 for the separate seven-BC1.0.0
review/freeze. New codex/cr21-boundary-contracts worktree is based on exact4a42222.
I-02/G-08 kickoff updates sprint/index/relations/append-only ledger before authoring.
All143 product findings remain open; no product implementation branch or BC freeze.

BC55 initial whole-set draft complete: seven contracts + shared types and review
candidate. Separate authors covered Git/GitHub and Runtime/Persistence; Master
reconciled versions/private wrappers, optional unavailable facts, compound receipts,
active context/intent and pure presentation. A bounded separate native commit
specialization produced25 captured PASS cases; Root read protocol/probe/bridge/
fixture and checked results/hashes. Signing fixture is noncryptographic. Root's
native Windows batch argv carrier probe passed exact supported operands. Sources
and outcomes are archived below LayerBoundaryContracts/Feasibility as evidence,
not product code. No accepted design file or product source changed. Independent
whole-set review of the committed draft precedes freeze.

Initial independent BC review of c291606c2ecfaae15f1bef7b5155d82aa6966712 found
BC55-TYPES-01/02 (both MEDIUM): missing configuration effect/reconciliation type
and missing shared identity for StorageRecovery. Candidate CI18/18 passed in
34026364910/34026398215; it did not override CHANGES_REQUIRED. Master recorded
the review and proposed correction in BCReview--002, added exact Git effect subset
and StorageRecovery shared-record mapping, and fixed one LOW prose link. Re-review
the corrected exact HEAD before freeze. No product/design change or finding closure.

BC55 complete technical set independently ACCEPTED at
7685494e45c0ef44fbccf9b49a589a90a78026d0; BC55-TYPES-01/02 resolved and
all18 CI checks SUCCESS in34027057276/34027058992. Master records REVIEWED then
FROZEN1.0.0 metadata for seven contracts/shared annex in BCFreeze--001, effective
only after final metadata exact-HEAD review/CI and PR56 merge. One reference-only
Git method-name correction is included for that review. G-08 remains active;
no implementation branch/Slice is active and all143 product findings stay open.

Metadata review at a4d160fe0273f7cdfa75794169a679bbc9321929 returned
BC55-META-01 (MEDIUM), despite18 successful checks in34027303143/34027304790.
Master corrected current accepted reviewedHead versus historical initialReviewedHead
and separate historical/current Relations. BCFreeze--001 records the actual review;
technical acceptance stands, corrected metadata exact-HEAD review/CI is next.

G-08/M0 COMPLETE: final BC freeze metadata independently ACCEPTED at
3ac7f20cedd8d14f131b2a15a602b233ef8ca5dc; CI18/18 in34027493442/34027495276;
PR56 expected-HEAD squash merge97cc0d8257603766dd741b49b7d8005857b421a9.
BC55-META-01 resolved; #55 closed, source/freeze/merge recorded on#21/#55.
Root main fast-forwarded cleanly. Created canonical codereview-21/refactor at the
merge and opened #57 for first fresh-worker Composition M1 CI/bootstrap contribution.
Plan--001 and I-03/M1 index/relations/ledger kickoff precede worker dispatch.
No product code changed yet; all143 findings remain open. All11 existing worktrees
were audited clean before branch creation; old checkpoint local801280a is retained.

M1 worker caught stale top I-02/G-08 header after kickoff. Root traced it to an
implicit Windows cp1252 read before UTF-8 write, which also altered Unicode dash
bytes in the two sprint files and prevented the intended header match. Root
restored those UTF-8 bytes and set explicit current I-03/M1 header. Future reads
must specify encoding=utf-8. Product-finding/release-baseline data matched the
prior accepted objects exactly; no product/design/BC content changed. Main CI at
actual BC merge97cc0d8 passed all9 jobs in34027623200. Worker implementation
continues under #57/Plan; this corrective metadata commit is supplied separately.

Initialized CurrentMatrix.yaml for all52 detailed checks plus13 E2E checks from
accepted Verification--001/Slices--001. All start not_executed; prior baseline
CI is separate. Fresh M1 reviewer preparation found no blocking authority/ownership
conflict and identified LOW historical ACTIVE labels; Master clarified them without
changing the current M1 contract. Product worker continues; no test result invented.

M1 candidate6685b4f76ca071be0125289925e7c9aa466dc10a received independent
CHANGES_REQUIRED for M157-M01..05 (five MEDIUM); actual review/source/probe/CI
records are in SDP/Verification/CR-#21/M1-Review--001.md and its evidence archive.
CI34029160810 finished17 success/1 macOS failure/1 explicit M3 skip. Worker is
correcting root aliases, nested callbacks, private helper boundaries, Git clean
semantics and trimpath toolchain identity. No product contribution is integrated.
Re-review corrected exact source and current CI before M1 acceptance; all143
baseline findings remain open and full Slice verification remains ahead.

M1 corrected candidateb1f533e062c8ab94b81e9309aaffadee096a5c7d has green
CI34030099746 (18 success/1 explicit M3 skip), including real Windows ARM64 and
macOS test/vet/build, but independent re-review is CHANGES_REQUIRED. M01/M02/M05
regressions pass; M03 private-decomposition wrapper traversal and M04 legacy crlf
profile remain, and M06 external child-junction source ownership is new. Actual
sources/results/CI/native evidence is in M1-Review--002 and its archive. Worker
corrects these three before a new exact-source re-review/CI; no integration or
baseline finding/Slice closure. All143 baseline findings remain open.

M1 COMPLETE: independent ACCEPT at05982a4adb35e39d7b7ba371e0d16d83b2b3674c;
all M157-M01..M06 resolved. Master verified raw retained controls, binary/archive
hashes, bounded Windows386/handle evidence and source CI34031420582 (18success,
1explicitM3skip). Clean canonical branch fast-forwarded to same SHA; actual push
integration CI34031833446 also passed18/1skip. M1-Review--003 preserves acceptance
and5 exact source/result captures. No product Runtime or whole Slice proof implied.
Opened #58 for M2 Domain; new contract/index/relations/matrix/ledger precede fresh
worker dispatch. All143 baseline findings remain open; seven BCs stay FROZEN1.0.0.

Domain #58 COMPLETE at6113ca55b57cb628b016c483070dbb52cd5dd79a, source/review/
integration identical. Independent ACCEPT without findings; constructor/value/
mutation/compiler probes and53-entry/12-source-copy hashes verified and archived
in M2-Domain-Review--001. Source CI34033181820 and canonical push CI34033772732
passed18 applicable jobs each/oneM3skip. No baseline finding/full Slice closed.
Read-only API inventory review found no material contradiction:5 Client methods,
28 commands,13 queries,6 events,7 ports/48 methods. API-Inventory--001 records the
checklist under frozen authority, not implementation acceptance. Opened59/60 for
independent API/ports and State-viewmodel leaves; M2-Surfaces/index/relations/matrix/
ledger kickoff is recorded before fresh worker dispatch. Master retains shared scope.

Current M2 leaves: API/ports#59 final7e3dd0a3545cc259af6dfe67885b20c8ddc57749
(product3e45f31) is under independent review/current CI34038064464. Viewmodel#60
atdbb89ea received M260-M01 MEDIUM despite CI34037022233 green18/M3skip: disjoint
modes reject retained simultaneous Navigator/Branch cockpit. Independent copies,
identity/output/stash/geometry controls pass; rejected vacant-Deploy hypothesis
adds no scope. Actual report/test/control/raw/hash evidence is archived in
M2-Viewmodel-Review--001. Worker corrects mode/content representation under#60;
new exact re-review/CI required. No M2 integration or baseline finding closure.

Viewmodel#60 source d279c3f2af59e48029e77d845f5a55271c7889cf independently
ACCEPTED; M260-M01 resolved, source CI34038307701 green18/explicitM3skip. Full
correction/review/control/raw/hash evidence is in M2-Viewmodel-Review--002. This
leaf waits for accepted API/ports#59 before serial integration; do not integrate
or close60 yet. API#59 current7e3dd0a is still independently reviewed, with a
separate Runtime/Storage sub-review consolidating semantic findings. No M2/full
Slice/baseline finding closure; all143 baseline findings remain open.

API/ports#59 independent complete review at7e3dd0a3545cc259af6dfe67885b20c8ddc57749
is CHANGES_REQUIRED for eight MEDIUM groups M259-M01..M08 (26 invalid admissions).
Separate Runtime/Storage reviewer confirms six cases in M06/M07; positive/race/
copy tests pass, but cross-field semantics do not. Actual full reports/probes/logs/
source hashes are in M2-API-Review--001 and its archive. Source/final CI passed18/
explicitM3skip; this does not override findings. Worker corrects all groups before
new frozen review/current CI. Viewmodeld279c3f is accepted and waits unintegrated
for API/ports. All143 baseline findings/full13 Slices remain; no M2 completion.

Native tool preparation under #21: Tooling-Preparation--001 records actual SSH/
OpenPGP signed commits and tamper rejection for SHA-1/SHA-256 on local Windows
Git2.48.1. All four profiles pass after recorded MSYS home/header probe corrections.
Only bounded script/log/result/manifest files are archived; keys remain outside
the repo. No adapter/bridge/Slice check is satisfied. Canonical checkpoint CI
34039754376 at04eb90a and34039134247 at5e3aa2e each passed18/explicitM3skip.
API#59 correction continues; accepted viewmodel remains held for API-first merge.
