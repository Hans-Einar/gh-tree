Repo: `Hans-Einar/gh-tree`

You are the long-running program agent responsible for completing the entire remaining `gh-tree` v0.4 architecture/refactor program from the repository's current authoritative state through a verified, release-ready — and, if safe, published — `v0.4.0`.

This is intentionally a large multi-stage assignment intended to test your ability to maintain architectural and governance coherence over a long-running task.

## USER AUTHORITY FOR THIS RUN

The user explicitly authorizes you to act sequentially as needed as:

* SteeringGroup
* Master / Integrator
* Focused Reviewer
* Refactor Designer
* Boundary Contract author/reviewer
* Worker
* Implementation Reviewer
* Verifier
* Release integrator

You may:

* create the remaining GitHub Issues required by the accepted SDP process;
* create branches/worktrees;
* write and merge governance/review/design PRs after their gates pass;
* create/freeze Boundary Contracts after the Refactor Design is accepted;
* implement the full v0.4 refactor;
* review and correct your own implementation when no independent agent is available;
* integrate accepted layer work;
* run tests/CI;
* open and merge the final product PR when all required gates pass;
* tag and publish `v0.4.0` only if the complete release contract is actually verified.

This is an explicit same-agent multi-role exception for this experiment.

However, preserve the repository's **role boundaries and staged evidence discipline**. When changing roles, independently reopen and inspect the actual artifact/code/exact SHA rather than accepting your own prior summary as evidence.

Where repository governance normally prefers independent reviewers, record that this run is an authorized same-agent multi-role execution and compensate with:

* frozen exact SHAs;
* fresh re-review of actual code/artifacts;
* explicit review findings;
* strong automated verification;
* CI/platform evidence;
* durable GitHub comments at every gate.

Do not weaken safety requirements merely because one agent performs multiple roles.

---

# 1. READ AUTHORITY BEFORE ACTING

Start by reading completely:

1. `AGENTS.md`
2. `developmentInstructions.md`
3. GitHub Issue #21 in full, including **all comments**
4. GitHub Issue #22 and:
   `SDP/Architecture/AS-#22/APS--001.md`
5. Accepted Broad Review:
   `SDP/Reviews/CR-#21/Broad/BR--001.md`
6. Accepted Runtime review:
   `SDP/Reviews/CR-#21/Layers/Runtime/LR--001.md`
7. Accepted TUI State review:
   `SDP/Reviews/CR-#21/Layers/TUI-State/LR--001.md`
8. Accepted Application review:
   `SDP/Reviews/CR-#21/Layers/Application/LR--001.md`
9. Accepted Git review:
   `SDP/Reviews/CR-#21/Layers/Git/LR--001.md`
10. GitHub Issue #33 completely
11. Current Domain review artifact:
    `SDP/Reviews/CR-#21/Layers/Domain/LR--001.md`
    from draft PR #34
12. all SDP templates relevant to subsequent stages
13. any governance/design/BC records referenced by the above

Do not use chat summaries as authority.

Repository, Issues, accepted SDP artifacts, exact SHAs, frozen BCs, tests and CI are authority.

---

# 2. CURRENT KNOWN START STATE — VERIFY, DO NOT BLINDLY TRUST

Product baseline remains:

* release: `v0.3.14`
* product SHA:
  `f626077ca0e59fbe9ede7ba1116982bb94b2eb6b`

Governance/review commits after this SHA are process/design authority and must not initially be mistaken for product-code changes.

Known current `main` before Domain acceptance:

`3531fd165ff2b35036928328685d45d84965e9ef`

Known current Domain review:

* Issue: #33
* branch: `codereview-21/layer-domain-review`
* HEAD:
  `ec482795a82371996370ec52e68c5c24b4e240b1`
* report:
  `SDP/Reviews/CR-#21/Layers/Domain/LR--001.md`
* draft PR: #34
* expected scope: exactly one review artifact
* findings: 4 HIGH / 8 MEDIUM / 2 LOW
* reviewer disposition:
  `ACCEPTED — FOCUSED DOMAIN REVIEW COMPLETE / REFACTOR NOT AUTHORIZED`

First independently verify all of this.

If Domain LR--001 is complete and consistent with Issue #33 and preceding accepted reviews:

1. post SteeringGroup/Master acceptance on Issue #33;
2. merge PR #34;
3. record exact merge SHA;
4. continue immediately.

If it has substantive defects, correct them through an explicit review/correction cycle before accepting it.

Do not skip this gate.

---

# 3. COMPLETE ALL REMAINING FOCUSED LAYER REVIEWS

Broad Review #21 established this focused-review order:

1. Runtime — complete
2. TUI State — complete
3. Application — complete
4. Git — complete
5. Domain — currently pending acceptance
6. Launch Discovery
7. Persistence
8. GitHub
9. TUI View
10. Composition

After Domain acceptance, create the missing focused-review Issues yourself as required.

Use the accepted Broad Review, earlier focused reviews and `SDP/Templates/LayerReview.md` to construct each work contract.

For each remaining layer:

* open a dedicated GitHub Issue;
* explicitly reference Issue #21 and all prerequisite accepted artifacts;
* freeze the same v0.3.14 product baseline for review evidence unless governance explicitly determines otherwise;
* perform the focused review;
* create:
  `SDP/Reviews/CR-#21/Layers/<Layer>/LR--001.md`
* use a dedicated review branch;
* open a draft review PR containing only the review artifact / strictly necessary metadata;
* independently inspect the finished artifact and PR scope as SteeringGroup/Master;
* post acceptance or exact corrections on the Issue;
* merge the review PR only when acceptable;
* record exact review HEAD and merge SHA;
* proceed to the next layer without waiting for user intervention.

Do not implement product fixes during these reviews.

## Required remaining reviews

### Launch Discovery

At minimum settle:

* discovery/provider ownership;
* normalized launch-point semantics;
* npm/pnpm/yarn/Make discovery;
* repository-relative cwd semantics;
* provider-private DTOs versus Domain/Application types;
* `.gh-tree/run.json` relationship to Launch Discovery versus Persistence;
* no Runtime/process ownership;
* cancellation/error semantics;
* target `internal/launchdiscovery/`;
* Application--LaunchDiscovery BC requirements.

Consume Domain's conclusion that only `LaunchPointID` is currently proven Domain-stable; decide whether further pure launch definitions are justified.

### Persistence

At minimum settle:

* repository identity/keying;
* user config versus repository-local config versus ephemeral UI state;
* active-worktree preference;
* last namespace/folder;
* `.gh-tree/run.json` ownership;
* filesystem path policy;
* atomic writes/crash safety;
* migration/backward compatibility;
* Application ownership of durable intent versus Persistence storage authority;
* target `internal/persistence/`;
* Application--Persistence BC requirements.

### GitHub

At minimum settle:

* remote repository/PR/branch facts;
* adapter-neutral semantic mapping;
* freshness/observation semantics;
* PR identity versus head/base mutable refs versus exact revisions;
* remote mutations currently supported;
* GitHub must not own local Git;
* Application owns local/remote reconciliation;
* target `internal/github/`;
* Application--GitHub BC requirements.

Preserve the accepted Ref/Revision Authority rule.

### TUI View

At minimum settle:

* rendering only;
* immutable supplied view model;
* no Application/Git/Runtime calls;
* no focus/state authority;
* layout, modal compositing, styles, colors, help rows, graph rendering and activity animation;
* relationship with TUI State;
* convergence/removal of duplicate `graphui` architecture;
* target `internal/tuiview/`;
* TUIState--TUIView BC requirements.

### Composition

At minimum settle:

* CLI/bootstrap/wiring only;
* repository resolution/prerequisites;
* adapter construction;
* root context;
* application startup/shutdown;
* Runtime cleanup failure propagation;
* `cmd/gh-tree/main.go`;
* optional `internal/composition/`;
* dependency-injection wiring;
* CI forbidden-import checks;
* release/version/bootstrap concerns;
* Master-owned files and integration sequence.

Composition is reviewed last because it consumes all other layer boundaries.

---

# 4. CREATE AND ACCEPT THE FINAL REFACTOR DESIGN

After all ten focused Layer Reviews are accepted and merged, open a dedicated Refactor Design Issue under parent #21.

Read all accepted reviews afresh.

Create:

`SDP/Design/CR-#21/REFDES--001.md`

using the repository's design template/process.

The Refactor Design must reconcile all reviews into one implementable architecture and explicitly define:

* final horizontal layer set;
* target physical directory/package ownership;
* exact files/packages to split, move, remove or temporarily shim;
* allowed dependency graph;
* Domain value-object ownership;
* Application command/query/use-case API;
* Application-owned ports;
* Runtime unified session model;
* Git fact/mutation model;
* GitHub observation model;
* Launch Discovery model;
* Persistence model;
* TUI State reducer/controller model;
* TUI View model/rendering boundary;
* Composition root;
* final Layer Boundary Contract inventory;
* operation/cancellation/event semantics;
* exact Ref vs Revision Authority rule;
* stable Stash identity;
* exact precondition/confirmation model;
* active-worktree authority;
* canonical branch reconciliation model;
* process-tree lifecycle guarantees;
* async user-intent supersession;
* migration/backward-compatibility strategy;
* Master-owned migration hotspots;
* vertical Slice migration plan;
* integration order;
* CI dependency rules;
* verification contract;
* final release-line decision.

Unless new evidence strongly contradicts it, v0.4 should be the architecture/refactor line.

Review the Refactor Design at a frozen exact SHA.

Correct any findings before accepting it.

Merge the accepted design artifact before BC freeze.

No product implementation before this stage passes.

---

# 5. CREATE, REVIEW AND FREEZE REQUIRED LAYER BOUNDARY CONTRACTS

From the accepted REFDES, create/update the required BCs under:

`SDP/LayerBoundaryContracts/`

Expected set includes at least:

* `BC--TUIState--TUIView.md`
* `BC--TUIState--Application.md`
* `BC--Application--Git.md`
* `BC--Application--GitHub.md`
* `BC--Application--Runtime.md`
* `BC--Application--LaunchDiscovery.md`
* `BC--Application--Persistence.md`

Domain normally remains governed by its design/API invariants rather than a traditional adapter BC unless the accepted design provides a strong reason otherwise.

Composition dependency rules may be better enforced in CI than represented as a large runtime BC.

For each BC:

1. define ownership;
2. inputs/outputs;
3. state authority;
4. identity semantics;
5. lifecycle;
6. errors;
7. cancellation;
8. concurrency/event ordering;
9. safety invariants;
10. forbidden responsibilities.

Review all BCs for cross-contract consistency.

Resolve contradictions before implementation.

Then explicitly mark the implementation-required BC versions `FROZEN` according to repository governance.

Record the freeze commit and versions on the parent program Issue.

---

# 6. IMPLEMENT THE COMPLETE v0.4 REFACTOR

After design + required BCs are frozen:

Create canonical integration branch:

`codereview-21/refactor`

from the accepted governance/design base.

Use layer-owned branches/worktrees as required by `developmentInstructions.md`, even though one Astra instance performs the work.

Recommended worker branches:

* `codereview-21/layer-domain`
* `codereview-21/layer-application`
* `codereview-21/layer-git`
* `codereview-21/layer-github`
* `codereview-21/layer-runtime`
* `codereview-21/layer-launchdiscovery`
* `codereview-21/layer-persistence`
* `codereview-21/layer-tuistate`
* `codereview-21/layer-tuiview`

Composition/shared cutover remains Master-owned unless REFDES explicitly decides otherwise.

Do not let multiple layer branches independently edit/move the same legacy file.

Implement by accepted vertical Slices and layer boundaries, using separately traceable commits where practical.

## Core architectural objectives

The finished production tree should physically reflect the accepted architecture, approximately:

```text
cmd/
  gh-tree/

internal/
  domain/
  application/
    api/
    ports/
    usecases/
  git/
  github/
  runtime/
  launchdiscovery/
  persistence/
  tuistate/
  tuiview/
  composition/        # only if accepted design needs it
  version/
```

Do not merely rename/move the old vertical packages wholesale.

The legacy architecture must actually be decomposed.

## Mandatory behavior/safety outcomes

### Domain

Implement the accepted small pure semantic model.

At minimum preserve:

* exact OID/Revision distinction;
* SHA-1 and SHA-256 full OID validation if accepted;
* attached versus detached HEAD as different semantic states;
* stable repository/worktree/branch/PR/stash/launch/session identities;
* typed operation-safe targets with expected revision where required;
* no adapter/TUI/Application imports;
* no serialization/persistence concerns unless explicitly justified by design.

### Ref vs Revision Authority

This is a hard safety rule:

> When user intent selects an exact commit OID, that immutable Revision is authoritative. Mutable branch/ref/PR names are locator metadata only and MUST resolve/fetch/verify to the expected OID before mutation.

Freshness alone is not enough.

Never silently replace an exact selected revision with the current tip of a mutable name.

Generalize the strong existing PR exact-ref + selected-SHA verification pattern.

Fix the known branch-deploy stale-local-tip failure.

### Git

Implement:

* authoritative local facts;
* explicit upstream states;
* freshness/fetch generation;
* exact expected-OID verification;
* stable stash identity by object identity rather than `stash@{n}`;
* safe precondition fingerprints;
* worktree branch occupancy;
* detached HEAD truth;
* typed mutation results/errors;
* mutation cancellation != rollback;
* explicit indeterminate/reconciliation path when necessary;
* safe pull/push semantics;
* no cross-source GitHub joins.

Never silently destroy dirty/untracked user work.

### Runtime

Implement one authoritative Runtime session registry.

Mandatory:

* collision-free opaque SessionID;
* unified launch + interactive terminal lifecycle model;
* explicit capabilities;
* process-tree ownership;
* Windows Job Object or equivalently strong descendant ownership;
* Unix process-group/session-group semantics;
* truthful `stopping` versus confirmed `stopped`;
* bounded graceful -> force stop;
* idempotent Stop;
* old session fully quiescent before Restart replacement;
* new SessionID after restart;
* deterministic PTY/ConPTY cleanup;
* exactly-once correlated terminal lifecycle events/results;
* aggregate Shutdown/StopAll result;
* no provider-string routing;
* no UI inference of `working` merely from process existence.

### Application

Implement coherent use cases rather than a concrete-adapter facade.

Application owns:

* Application-owned ports;
* OperationID;
* operation lifecycle;
* exactly-once terminal results;
* cancellation policy;
* safe supersession;
* confirmation/precondition identity;
* active-worktree session context;
* canonical local/upstream/GitHub reconciliation;
* graph/PR/worktree joins;
* Runtime event normalization;
* stable API/read models.

Move deploy/stash/checkout/launch/runtime/persistence workflows out of TUI.

### TUI State

Replace the V039/V310→V314 wrapper chain with one coherent deterministic interaction authority.

Own only ephemeral user intent such as:

* mode;
* pane/subpane focus;
* selection;
* modal state;
* search/filter;
* console tab/input focus;
* view-local scroll/cursor;
* request generation / user-intent correlation.

Background completions must not steal newer focus/selection/modal intent.

Use OperationID + request generations + intent tokens according to accepted contracts.

No direct Git/GitHub/Runtime/Persistence orchestration.

### TUI View

Rendering only.

Own:

* layout;
* visual styling;
* colors;
* help;
* modal composition;
* graph rendering;
* textual formatting;
* animation frames.

Do not own product or interaction authority.

### Launch Discovery

Discovery/normalization only.

No Runtime session ownership.

Provider-specific implementation details remain private unless explicitly made semantic in the accepted design.

### Persistence

Own storage mechanics, not product workflow decisions.

Use appropriate atomic/crash-safe semantics accepted by BC/design.

Preserve/migrate existing user state/config where required.

### GitHub

Remote facts/mutations only.

No local Git mutation and no canonical local/remote reconciliation.

### Composition

Wire layers and own root startup/shutdown only.

No business logic.

---

# 7. KEEP v0.3.14 USER CAPABILITY WORKING

v0.4 is an architectural refactor and correctness hardening, not a feature purge.

Preserve existing accepted gh-tree functionality unless REFDES explicitly removes or changes something for a documented reason.

At minimum retain working behavior for:

* namespace PR/branch navigator;
* branch context / commits;
* worktree discovery and activation;
* create/retarget worktree;
* dirty-worktree safety;
* stash cockpit;
* deploy selected PR/branch/commit;
* exact revision display;
* Git graph;
* diff/review;
* staging/commit/push/pull/fetch/new branch/create PR flows currently supported;
* launch discovery/config;
* F5 launch;
* multi-console behavior;
* interactive PTY/ConPTY terminal;
* focus/mnemonic navigation;
* persisted repository UI state;
* cross-platform GitHub CLI extension packaging.

Do not add unrelated new features simply because the architecture now makes them easy.

---

# 8. REVIEW EACH IMPLEMENTED LAYER BEFORE INTEGRATION

For each layer implementation:

1. freeze exact Worker HEAD;
2. switch role to Reviewer;
3. inspect source, tests, BCs and diff independently;
4. do not use the Worker summary as evidence;
5. create durable review findings;
6. correct findings on the Worker branch;
7. re-review exact corrected HEAD;
8. only accept when no blocking findings remain.

Record:

* branch;
* exact SHA;
* Slices;
* BC versions;
* tests;
* findings;
* final disposition.

Then switch to Master and serially integrate accepted commits into:

`codereview-21/refactor`

Record every integration SHA.

If integration conflicts reveal unclear ownership, treat that as architecture evidence. Do not paper over a genuine BC/design defect.

If a FROZEN BC proves wrong:

* stop affected implementation;
* record a BC-CHANGE finding;
* update design impact;
* review/re-freeze BC;
* re-review affected layers;
* continue only after the contract is coherent.

---

# 9. VERIFICATION GATE

When integration is complete, freeze exact integrated refactor HEAD.

Create verification artifacts under:

`SDP/Verification/CR-#21/`

Execute the full accepted verification contract.

At minimum run/verify:

## Go/tooling

* `go test ./...`
* `go test -race ./...` where platform/support permits
* `go vet ./...`
* full build
* formatting/static checks
* forbidden-import/dependency-boundary checks

## Domain

* value equality;
* invalid constructor cases;
* SHA-1 OID;
* SHA-256 OID;
* ref versus revision;
* attached/detached HEAD;
* stable stash identity;
* SessionID;
* package purity/import rules.

## Application

* operation IDs;
* exactly-once completion;
* cancellation;
* safe supersession;
* confirmation stale/revalidation;
* active-worktree persistence coordination;
* adapter error normalization;
* cross-source branch reconciliation;
* stale async results.

## Git

Use temporary real repositories/worktrees/remotes where practical:

* clean/dirty worktrees;
* detached HEAD;
* exact revision checkout;
* stale local branch versus selected remote revision;
* expected-OID mismatch refusal;
* upstream none/resolved/gone/unresolved;
* freshness generations;
* ahead/behind/diverged;
* branch occupied in another worktree;
* stash identity after reflog position shift;
* destructive precondition changes;
* restore after post-confirmation edit;
* fetch/pull/push;
* interrupted mutation effect-state classification;
* graph/diff local facts.

## Runtime

Cross-platform contract tests:

* unique SessionIDs;
* start failure cleanup;
* Stop idempotency;
* stopping/stopped truth;
* restart uses new ID;
* exactly-once lifecycle completion;
* output bounds/order;
* shutdown aggregate result;
* no descendants survive successful Stop.

Platform-specific:

### Windows

* Job Object/equivalent process-tree containment;
* ConPTY;
* Ctrl+C/interrupt behavior where applicable;
* restart/stop/shutdown with descendants.

### Unix

* process groups/session groups;
* PTY;
* graceful/force escalation;
* descendant cleanup.

## TUI State

Deterministic reducer/state-machine tests:

* focus transitions;
* modal ownership;
* pane/subpane navigation;
* async console start followed by Alt/Tab navigation;
* stale completion cannot steal focus;
* request generation replacement;
* stale diff/branch/stash/launch responses;
* console tab selection;
* active-worktree projection update;
* deterministic fallback when selected semantic item disappears.

## TUI View

* layout degradation at narrow terminal sizes;
* modal overlays do not change controller authority;
* context help;
* graph/detail rendering;
* no backend imports.

## Persistence

* migration from current state/config;
* atomic write behavior;
* invalid/corrupt data handling;
* repository key normalization;
* platform path behavior.

## Launch Discovery

* nested project discovery;
* npm script names such as `dev:wan` remain one exact script;
* Make ordered targets;
* repository-relative cwd;
* saved/default launch semantics according to accepted ownership.

## GitHub

* mocked/parsing tests;
* stable PR/branch semantic mapping;
* exact head revision;
* observation/freshness behavior;
* API errors.

## End-to-end Slices

Exercise at least:

* refresh navigator;
* branch/PR selection;
* deploy exact selected revision;
* dirty deploy requiring stash confirmation;
* stash apply/pop/drop identity;
* active worktree;
* fetch/pull/push;
* graph/commit/diff;
* launch start/stop/restart;
* interactive terminal open/write/resize/stop;
* user navigates away while async operation completes;
* application shutdown with active sessions.

---

# 10. CROSS-PLATFORM CI AND RELEASE PACKAGING

Do not consider verification complete solely because tests pass on the current machine.

Require GitHub CI for supported platforms consistent with repository policy, including Windows/Linux/macOS builds/tests where configured/required.

Confirm release/precompile workflow remains compatible with:

`gh extension install Hans-Einar/gh-tree`

and:

`gh extension upgrade tree`

Verify the v0.4 binary still has the correct extension naming/packaging.

Do not silently remove architectures currently supported by release workflow.

---

# 11. FINAL PRODUCT PR

After verification passes:

1. open the final product PR from:
   `codereview-21/refactor`
   to the release target selected in accepted REFDES, normally `main`;
2. include:

   * accepted design;
   * frozen BC versions;
   * integrated exact SHA;
   * layer review dispositions;
   * verification artifact;
   * CI evidence;
   * migration notes;
   * known limitations, if any;
3. perform a fresh merge-readiness review of the exact PR HEAD;
4. ensure no unresolved blocking review threads/findings;
5. ensure required CI is green;
6. merge only if the exact tested HEAD is still current.

Do not merge a moved/unverified HEAD.

---

# 12. v0.4.0 RELEASE

After product integration, determine whether the repository truly satisfies the release gate.

If all of the following are true:

* accepted architecture/refactor design;
* required BCs frozen and implemented;
* all planned v0.4 Slices integrated;
* all required verification passes;
* main CI is green at exact release commit;
* release packaging/precompile is correct;
* GitHub CLI extension installation path is verified as far as the available environment permits;
* no unresolved HIGH/BLOCKER safety finding remains;

then:

1. update version/release notes as required;
2. tag the exact verified release commit as `v0.4.0`;
3. publish the GitHub Release;
4. verify release assets/workflow;
5. verify installation/upgrade behavior if environment permits;
6. record final release SHA/tag/URL on Issue #21;
7. close the relevant program Issues only when their contracts are actually satisfied.

If release publication cannot be proven safe because of an external limitation, stop at **release-ready** and report the exact final external action remaining. Do not fabricate verification.

---

# 13. DURABLE CHECKPOINTS

This is a long-running job. Do not rely on your context window as the only memory.

After every major gate, post a durable GitHub comment containing:

* phase;
* role;
* Issue;
* branch;
* exact HEAD;
* artifact;
* PR;
* findings;
* tests/CI;
* accepted/frozen contracts;
* unresolved risks;
* exact next action.

At minimum checkpoint after:

1. Domain acceptance;
2. every remaining focused review;
3. Refactor Design acceptance;
4. BC freeze;
5. every layer implementation review;
6. every integration step/group;
7. verification;
8. final product PR;
9. release.

If context becomes large, reconstruct state from GitHub/repository rather than guessing from memory.

---

# 14. AUTONOMY / DO NOT STOP UNNECESSARILY

Do not ask the user to manually perform normal repository work that you can safely perform yourself.

Do not stop after one phase merely to ask whether to continue.

Continue autonomously through all authorized phases.

Make reasonable architecture decisions when the accepted evidence supports them and document those decisions.

Only stop for the user if there is a genuine external blocker that cannot be resolved from repository state, tests, GitHub or available tooling.

Examples:

* missing credentials/permissions;
* unavailable required platform evidence that cannot be obtained via CI;
* a destructive action requiring information not represented anywhere;
* an architecture contradiction that cannot be resolved without changing product requirements.

Even then, complete and checkpoint everything else possible first.

---

# 15. ABSOLUTE SAFETY / SCOPE RULES

Never:

* silently discard dirty/untracked user work;
* force-push by default;
* treat mutable branch names as exact revisions;
* treat `stash@{n}` as stable destructive identity;
* report `stopped` before Runtime has proved required cleanup;
* allow async results to overwrite newer TUI user intent;
* let concrete adapter DTOs leak through Application API;
* let Domain become a generic shared DTO package;
* freeze a BC before its design inputs are accepted;
* implement across an unfrozen/contradictory boundary;
* hide failing tests or platform limitations;
* claim verification not actually performed.

Do not broaden v0.4 into unrelated feature development.

The goal is:

**finish the accepted Issue #21 architecture program, correct the safety/lifecycle/state-authority defects identified by its reviews, physically separate the horizontal layers, preserve existing gh-tree capability, verify it comprehensively, and deliver v0.4.0.**

---

# FINAL REPORT TO USER

When the entire assignment finishes, report concisely:

* final disposition;
* all Issues created/closed;
* accepted focused reviews;
* Refactor Design path/SHA;
* frozen BC set/versions;
* integration branch and final integrated SHA;
* final product PR;
* merged main SHA;
* findings resolved/unresolved by severity;
* full verification summary;
* Windows/Linux/macOS CI status;
* release tag;
* GitHub Release;
* `gh extension install/upgrade` verification;
* architecture package map actually delivered;
* any remaining known limitations;
* exact next action, if anything remains.

Do not report success unless the evidence supports it.
