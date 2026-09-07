# M3 adapter contribution contracts

State: ACTIVE after explicit user resumption from checkpoint3e259652.
Apply the resumed execution policy in UserRunContract.md; frozen behavioral contracts remain unchanged.
Authority: #61 Git, #62 GitHub, #63 Persistence, #64 Launch Discovery and #65 Runtime
under full program #21. I-03 / M3; all SLC-01..13 remain selected.

Verified combined M2 product integration: d60afe8fb763269964bcdc05958c8cdbc7849dea.
Exact integrationCI34046727344 passed18 applicable jobs/explicit pre-Runtime skip.
API source d4c3bfdd2038bbc5921488fbcffad1b8736c7460 (technical6e289c4),
viewmodel d279c3f2af59e48029e77d845f5a55271c7889cf and Domain6113ca5 are accepted.
Issues57..60 are complete/closed; all143 baseline program findings remain open.
Preparation checkpoint3cf264141764385eb6afe3f7e4e25a7504020690 also passed
CI34047530046 with18/explicitM3skip. Superseded docCI34047480826 was canceled,
not counted passing. No adapter, workflow, full Slice or release proof is implied.

## Ownership and dispatch

| Serial order | Issue / layer | Branch / worktree suffix | Sole worker product scope | Report |
|---|---|---|---|---|
| 1 | #61 Git | codereview-21/layer-git / git-implementation | internal/git/ | SDP/Implementation/CR-#21/Git/M3-Git--001.md |
| 2 | #62 GitHub | codereview-21/layer-github / github-implementation | internal/github/adapter/ | SDP/Implementation/CR-#21/GitHub/M3-GitHub--001.md |
| 3 | #63 Persistence | codereview-21/layer-persistence / persistence-implementation | internal/persistence/ | SDP/Implementation/CR-#21/Persistence/M3-Persistence--001.md |
| 4 | #64 Launch Discovery | codereview-21/layer-launchdiscovery / discovery-implementation | internal/launchdiscovery/ | SDP/Implementation/CR-#21/Launch-Discovery/M3-Discovery--001.md |
| 5 | #65 Runtime | codereview-21/layer-runtime / runtime-implementation | internal/runtime/ | SDP/Implementation/CR-#21/Runtime/M3-Runtime--001.md |

All worktrees are below C:/Users/hanse/GIT/gh-tree-wt. Master creates them from
the kickoff commit containing this contract; exact base is recorded in the
dispatch ledger/Issue comment before each worker starts. Initial parallel authors
are Git, GitHub and Persistence. Discovery and Runtime are authorized queued
contributions and start as slots become available. Each imports only accepted
Domain/API/ports and its permitted low-level libraries, so no concrete adapter
implementation dependency or cross-folder authoring is introduced. Integration
is strictly Git → GitHub → Persistence → Discovery → Runtime; accepted later
results wait for preceding serial gates. Starting a worker does not pass its gate.

Fresh worker contexts read AGENTS, full developmentInstructions and their full
Issue/#21, all accepted/frozen referenced documents and actual current M2 code.
Mandatory interface declarations are now real source in application/ports; pure
constructor validity is not native observation or resource authority. Workers own
only their row plus local tests/README and that bounded report. Legacy sources/
tests, module/workflow files, other layers, frozen design/BCs and Master metadata
are not delegated. Master performs no product implementation; it supplies separate
fresh reviewers and coordinates any explicit shared Composition/CI assignment.

## Complete behavior and verification

Each full authorizing Issue is the exact behavioral contract, read with its frozen
BC and normative design appendices. Implement every declared method and family:
GitFacts11+GitMutations12, RemoteFacts4+RemoteMutations1, Storage6, Discovery2,
Sessions12. No old-code facade, generic process/shared-types layer or empty
placeholder can substitute for the accepted architecture. Shared API/ports are
read-only to these workers; a deficient frozen contract requires explicit
BC-CHANGE/impact/review/refreeze and affected re-verification.

| Layer-owned verification | Required implementation evidence |
|---|---|
| Git / V-GIT-01..10 | Real owned SHA-1/SHA-256 repos, linked worktrees/clones/remotes; exact facts/targets, common-scope scheduler+lock, preparations/receipts; full G8/G8a/G9/G10 native retained publication, commit hooks/signing, stash occurrence deletion, errors/cancel/recovery and Windows/Linux/macOS proof. Capability-specific2.43 commit/read versus newer attach support, no local-path hardcoding/fallback. |
| GitHub / V-GH-01..03 | Registered explicit remote scope, strict private DTO parsing, bounded provenance/paging/fork/unknown facts; exact pre/post-create expectations and all creation outcomes; labeled raw transport fixtures and actual bounded subprocess resource tests, no test-created live remote PRs. |
| Persistence / V-PER-01..04 | Strict whole schema/unknown-member retention and exact versions; permanent cooperative lock, nofollow object binding, selected native NT/Unix publication/metadata/recovery protocols, real concurrent/failure/crash/late-writer controls and claimed platform profiles including FreeBSD metadata. |
| Discovery / V-LCH-01..03 | Passive bounded nested npm/Make sources, exact member/manager/override/ordered selection, copied saved versions, native root/project no-link observation, cancellation/partial diagnostics; no process or run.json ownership. |
| Runtime / V-RUN-01..08 | Sole bounded registry/admission/event/ACK/output/input/restart/shutdown ownership; actual acquired cwd, complete Unix acquired-helper and Windows native-broker protocols, generated helper closure/extraction; native Windows/ARM64/emulation/WOW64/Linux/macOS/FreeBSD lifecycle/PTY/ConPTY/resource evidence. |

Every report maps each layer-owned clause to exact source, fixture, command,
platform, exit/log/hash and CI evidence. Neighboring boundary checks use accepted
values and declared ports; full Application/State/View and combined vertical paths
remain M4/M5/M6/final gates. This separates verification scope without waiving
native adapter-owned requirements. All12 architecture/build selections and current
configured source CI must pass. Cross-builds, mocks and feasibility prototypes
remain distinct from native execution. Never close a baseline finding or full
Slice solely because its adapter folder compiles or isolated happy paths pass.

Source/tool preparation is available in Tooling-Preparation--001 and accepted
native feasibility archives. Real local signatures are tooling controls only,
not Git bridge acceptance. Persistence retains original Expected anchor evidence
even when recovery Original is reobserved at a newly established parent. Standalone
unknown Push recovery is preserved; known request/summary/returned bindings must
validate represented facts without changing them. These existing-contract
interpretations do not supply native implementation proof.

Runtime must implement the exact accepted M1 check:
go run ./internal/runtime/cmd/helpergen -check, canonical Windows/amd64 Go1.25.0.
Required internal/runtime/brokerassets/{broker-amd64.gz,broker-arm64.gz,manifest.json};
sources internal/runtime/broker/ (including broker/cmd/), cmd/helpergen/ and actual
dependencies. Broker does not import brokerassets. Real independent regeneration
and source/PE/compression/hash verification, no checkout rewriting or runtime
compiler/download/extra public asset. First Runtime source/image activates the
mandatory check and missing inputs fail. Master owns workflow invocation changes.

## Acceptance and continuation

Worker freezes/pushes exact source and reports actual local/current CI and limits.
Fresh independent reviewer reads real source/tests/contracts and frozen SHA,
reproduces adverse and legitimate controls, and records any findings outside the
candidate. Corrections require a new exact review. Master verifies evidence and
current CI, integrates accepted sources serially, checks integrated CI and updates
implementationNotes/Handoff/CurrentIndex/Relations/append-only Ledger. No product
PR, entry cutover, legacy retirement, main merge, tag or release is authorized by
these adapter Issues. M3 ends only after all five contribution gates pass; the
full M4..M8 objective and all13 Slices remain afterward.

## Native value interoperability conventions

Implementation detail within frozen B3/Cwd/remote scope semantics, coordinated
between the emitting and consuming workers; no new layer or authority is added.
- Remote RepositoryID: domain.Remote with lowercase DNS-host/owner/name token,
  e.g. github.com/hans-einar/gh-tree. Transport .git normalization is separate;
  adapters still verify/register associations, never decode tokens into authority.
- Windows DirectoryIdentity: DirectoryWindows; full FileIdInfo uint64 volume and
  native16 FileID bytes; stamp birth-filetime:<unsigned CreationTime FILETIME>.
  Use native aligned storage and verified ABI; legacy64 file IDs are insufficient.
- Unix: DirectoryUnix; native device as uint64, inode uint64 little-endian in
  FileID bytes0..7 and zero bytes8..15. Prefer proved-available birth:<signed-sec>:<nsec>;
  otherwise explicitly use change:<signed-sec>:<nsec>. Decimal is canonical without
  padding, nsec0..999999999. Never fabricate availability or a clock stamp.
- Consumers compare the supplied stamp profile, without silently upgrading it or
  ignoring drift. Birth stamps remain stable under child creation; change fallback
  is short-lived and must honor frozen stale/held-object/own-effect rules. Test
  child creation/rename versus replacement and unsupported/unavailable profiles.

Value vectors: Unix device42/inode0x0102030405060708 has FileID
08070605040302010000000000000000 and stamp birth:1:2 (or explicitly change:1:2).
Windows device42/FileID000102030405060708090a0b0c0d0e0f uses the same16 native bytes
and stamp birth-filetime:116444736000000000. These are layout vectors, not native
proof. Native observation, nofollow acquisition and process/resource gates remain.

Shared CI #66 (Composition) is accepted/integrated atd5eeaac, CI34058149301
passes19/helper skip. Native FreeBSD supplements existing Linux/macOS/Windows/
ARM64 and all12 architecture/build jobs; no adapter proof from absent sources.

## BC-CHANGE-67: independent status causes (FROZEN 1.1.0, #67)

Concrete G3/API omission: one Kind without a cause cannot carry separate staged,
unstaged/untracked/conflict changes from current index/filesystem facts. The full
DRAFT and impact proposal is preserved at980513b2b126842ab1d50558931a84441c253172.
Fresh independent [review](Application/M3-Status-BC-Review--001.md) ACCEPTS it,
with no blocking findings and explicit legitimate-case/API/native/projection tests.
Master freezes the corresponding [Git G3](../../LayerBoundaryContracts/BC--Application--Git.md)
delta as1.1.0 in the commit adding this decision. Required cause-tagged rows retain
per-cause kind/rename identity and consistent current facts; other contracts and
all native mutation protocols remain1.0.0/unchanged. The original whole-set freeze
history stays in BCFreeze--001; this bounded correction does not reopen that set.

API correction675dfff/review0bb8f28 is independently ACCEPTED, integrated and
pushed at4d1b5486a029eeabcd874306d7a7ece4ae629ff3. Source CI34055882003 and
integrated CI34056365373 each pass18 applicable jobs/helper skip; integrated
Application race/vet/windows architecture and source-tree match pass. Git may
adopt this exact prerequisite at a clean boundary and resume status. Impact is
Application--Git G3, API value/consistency/tests/README; no port, Domain, adapter,
viewmodel or legacy change. M4/M5 must preserve causes in public/presentation
projection. The freeze is contract authority, not implementation acceptance.

## Windows storage security profile interpretation (#63)

Master read frozen Persistence404..412 with accepted Feasibility/Persistence148..155:
its bounded profile excludes audit-SACL/security-policy replication and requires
an explicitly supported profile for custom audit-preservation requirements.
The BC's explicit refusal/profile evidence obligation does not claim that limited
READ_CONTROL queries prove full audit-SACL absence. Ordinary profile preserves/
verifies owner/group/ordered DACL/protection/inheritance and supported label/
resource security; unsupported or unreadable access-affecting policy still refuses.
Never infer absent audit data, full descriptor equality or an unsupported security
policy from those limited observations. No automatic privilege changes or new
public configuration/API authority are introduced. Any access-affecting SACL class
that cannot be detected must be raised before the affected publication proceeds.

[Microsoft's query-rights table](https://learn.microsoft.com/en-us/windows/win32/secauthz/security-information)
and [GetSecurityInfo](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-getsecurityinfo)
confirm READ_CONTROL for owner/group/DACL/label/resource/CAP queries versus
ACCESS_SYSTEM_SECURITY for audit-SACL. Adapter code/report/native tests must state
the selected profile and its limits, and the critical independent review must
check both preservation and refusal. This is an interpretation of existing
bounded authority, not native proof, a contract relaxation or adapter acceptance.

## Plan summary digest plumbing (#68, complete)

Existing frozen Git G5 requires coordinator approval bound to the issued summary
digest. Actual ports stores PlanSpec.SummaryDigest privately, yet ApprovalIssuer
requires the consumer to provide it; only same-author test fixtures know it.
#68 authorizes a bounded read-only PlanSummaryDigest accessor through the existing
sealed identity path, value copy and proper invalid-plan rejection. No new hash
format, confirmation rule, approval privilege or frozen boundary change. Fresh worker/reviewer accepted source70d1719/review736129a; source CI34057016256
and integrated396a629 CI34057368177 pass18/helper skip. Exact integrated ports
race/vet/windows architecture pass. Git may adopt at a clean boundary; original
approval and native-registry obligations remain.

Current dispatch: Discovery#64 starts on its already-pushed planned base412f33e
and owned branch/worktree from M3-Assignments; unchanged Discovery/Runtime/shared
contracts and current native identity conventions apply. Persistence successor
starts from pushed87506ee, three public loads present; three commits/recovery
protocol remains. #66 source489a731/reviewbe94b44 is ACCEPTED and integrated atd5eeaac with
source/integrated CI19PASS. Active Git/Persistence/Discovery branches adopted it
at clean boundaries, preserving their owned source. No adapter integration or
complete Slice claim.

Recovery identity interpretation: frozen Persistence392..395 does not require a
manifest to embed its own final ctime. Stable persisted RecoveryID is distinct
from mutable artifact observation SourceVersion (468..480); DirectoryIdentity
is short-lived (BoundaryTypes103..105). Reobserve current facts and preserve
proven store/artifact association without equating changed observations or
waiving replacement checks. No-birth profiles still need actual restart proof;
unknown association cannot become authority to adopt/replay/delete an artifact.

Git constructor wiring for M4/Composition: create one immutable coordinator
ports.ApprovalIssuer and supply the same value to coordinator and Git
Options.ApprovalAuthority. Git verifies both Issued and ValidFor; it never issues
approvals. Zero authority provides facts only and cannot prepare executable plans.
This implements existing G5 ownership, not a new boundary API or confirmation rule.

Persistence early audit findings M363-PR01/02/03 are independently resolved at
91acac5 in bounded native controls; complete acceptance remains pending. New actual
Windows12-cycle regression at a147a1d shows the same FileIdInfo file ID with a changed
CreationTime after native replacement, consistent with MS-FSA FileRenameInformation
tunneling. A file creation timestamp cannot silently be treated as immutable.
No timestamp-writing workaround, identity-check removal, or selected native
primitive change is accepted. Fresh finishing worker must establish a safe native
profile and crash/restart association; any frozen primitive change needs BC review.
Shared DirectoryIdentity conventions concern directories and remain unchanged.

## External review-access hold

Required Git foundation review stopped with a service cybersecurity content/access
block. M3-Foundation-Review--001 is a Master incomplete-status record, not an
acceptance. Active authors stopped at clean pushed boundaries; no retry/rephrased
review/model switch or withheld-output recovery was attempted. No additional
implementation begins until the approved review access/surface is resolved.
All original scope and engineering gates remain; the program is not complete.

Git technical8c2584b/report25240dd adds verified common scheduling/permanent locks,
with CI34063202321 19PASS/helper skip, but no native mutation assembly. Persistence
mechanism699a69f/report3a5a8e7 proves owned Windows ObjectID/FileIdInfo behavior,
including negative ID-transfer and pre-outcome process-death controls; proposal
remains unadopted in product. Its Windows failing regression remains, FreeBSD
profile is unresolved. Mechanical6fc6c2c only repairs native test file selection,
with all12 test binaries compiling; no native/full-matrix acceptance implied.
Discovery2a18fdd/report8e2d2d1 is frozen with successful source CI, review pending.
Current exact refs and next steps are in Handoff/CurrentIndex and worker reports.

Continuation clarification: current goal instruction requires taking available safe
work. The external hold remains on the failed Git review and its dependent serial
integration; independent authorized M3 work can proceed. #64 review of the separate
frozen passive Discovery adapter is dispatched without changing/retrying the Git
audit or substituting for its required verdict. No model/access controls changed.

Independent #63 proposal assessment may also proceed: native owned-artifact
ObjectID/FileIdInfo probes at699a69f (platform selection corrected6fc6c2c), actual
frozen artifact identity contract and unchanged class65/class11 publication.
This is separate from the blocked Git audit. Product adoption remains withheld
until the private protocol's scope, negative controls and crash association are
reviewed; no directory convention, timestamp repair or original-ID mutation waiver.

#65 Runtime dispatch is independent of concrete adapter implementation: the Issue
requires accepted M2 values/ports and frozen Runtime authority, already present.
Its clean planned branch was fast-forwarded to26d5f4f (verified executable baseline
d5eeaac and current shared native CI/governance), then pushed before authoring.
Fresh worker owns only Runtime and its report. No Git-review retry/substitution,
serial integration, concrete adapter coupling or native safety waiver occurs.
The first Runtime input activates the existing helper conformance gate; incomplete
checkpoints remain explicit. No permanent placeholder implementation is accepted.

## Reviewed Windows artifact identity adoption (#63)

Master accepts M3-Identity-Review--001's bounded proposal assessment at6fc6c2c:
full volume/128-bit FileIdInfo plus all64 native ObjectID-buffer bytes for newly
and exclusively FILE_CREATE-created regular artifacts. CREATE_OR_GET precedes
first journal/data flush; all later validation is GET-only and profile-directed.
Existing originals never receive an ID for proof; only exact missing-ID errno2
permits initial strict birth selection. Recorded ObjectID profiles never fall
back, and old birth manifests stay exact without automatic upgrade. Directory,
lock, root/Expected-anchor, whole-version, class65/class11, byte/security and
outcome rules remain unchanged. See the independent report's complete adoption
and product-verification requirements. No frozen BC change is needed or made.

A fresh bounded worker may implement this private correction under#63. Native
mechanism acceptance is not product acceptance: public repeated commit/recovery,
process-death/error/mismatch/legacy profiles, native amd64/386/ARM64 and full
required gates still need proof and a separate fresh implementation review.
Git review remains blocked; no serial integration is authorized by this decision.

Discovery20bd8ac/review4e9f479 is independently ACCEPTED, M364-M01 resolved,
CI34065291612 19PASS/helper skip. Candidate remains held for prior serial gates.

Runtime parallel ownership #69: original #65 author owns Unix private owners and
parent registry/Sessions assembly; fresh Windows author owns only broker Windows
files, broker/cmd, brokerassets and cmd/helpergen plus M3-Windows--001 report, on
codex/cr21-runtime-windows at46fe59c. Shared broker/protocol.go/start.go and common
tests are coordinated interfaces. This explicit Issue ownership exception remains
inside Runtime; no new layer/BC or neighboring edits. Master integrates only after
separate Windows review and exact source gates, then verifies combined Runtime.

## Current bounded milestones and helper ownership (#63/#65/#69/#70)

Master accepts the Windows identity correction at bfdc79c, separate review
53de3e1 / Persistence/M3-Identity-Code-Review--001: no bounded findings. Independent
Windows amd64/386, Linux protocol reconfirmation and actual native ARM64 CI pass.
CI34066675310 remains failing only native FreeBSD (18 success, helper skip);
full #63/no-birth/source-name/fault/resource gates remain open. New small review
controls/logs are preserved in existing M3-Persistence-Protocol--001 evidence.

Runtime Unix checkpoint da29569 and Windows live-startup checkpoint 9c9a2f7 are
pushed partial implementations. #70 now owns only cmd/helpergen and brokerassets
plus M3-HelperAssets--001; #69 retains Windows broker/cmd and native extraction.
#70 starts the real committed broker.RunWindowsPrivate entry at 9c9a2f7 on
codex/cr21-runtime-helper-assets. Shared files remain coordinated. Exact pinned
helper interface/source closure and fresh review are required; ongoing broker
changes require regeneration. No complete Runtime or serial integration claim.

Before dispatch, untouched #70 base advanced to ab608327 after #69 pushed its
real parent-client/protocol tests (pipes/ConPTY/force stop/root-first cleanup).
Broker package must not import assets through client/extraction code; parent
Runtime supplies private asset values. No native/full candidate acceptance.

#70 narrow shared prerequisite: Master permits only Composition architecture
policy.go/check.go/testdata/cases.json to admit encoding/json and debug/pe for
assets, retaining in-memory use and typed rejection of pe.Open (also referenced
or aliased). Existing dependency/inventory/CI/frozen rules stay unchanged.
Standard parsers avoid handwritten JSON/PE validation and extra generated output.
Positive byte-reader and negative path/native-I/O/recursive-import fixtures use
the existing test format. Separate helper candidate review includes this small
delta; no extra review round or broader Composition ownership is authorized.

#71 queues a fresh Runtime parent Sessions worker after a slot and actual native
client seam are available. Exact common root foundations: 6be865b. #65 author
now retains only Unix broker/client/native hardening and its report; #71 owns
root Runtime Go/tests/README plus M3-Sessions--001. No new branch/worker yet.
The Issue preserves all12 methods, ordering/reservations, input split accounting,
retained cleanup ownership and separate fresh native/parent reviews. Test-only
owners may support early state proofs; final acceptance requires real clients.
No unreviewed native merge, production placeholder or public contract change.

#70 source b72ad94 has CHANGES_REQUIRED at review7fdfc9b: H01 external input
snapshot binding, M02 fresh-cache admission and M03 canonical Go root junctions.
Fresh author corrections stay in existing #70 paths, then regenerate/push and
independently confirm. Build-only pinned Go dependency preparation is allowed
within -check; no product runtime download/compiler, pin mutation or CI waiver.
