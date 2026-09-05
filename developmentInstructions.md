# Development Instructions

This file defines the default development, review, refactor and integration process for `gh-tree`.

The purpose is to make work reproducible across ChatGPT/Codex sessions, keep architecture in the repository instead of chat memory, and support parallel implementation without avoidable merge conflicts.

## 1. Terminology

### Slice
A **Slice** is a vertical unit of user-visible or externally observable behavior that may cross several horizontal architectural layers.

Examples: deploy a selected commit into Active worktree; stop a launch process; render local-vs-upstream commit state.

A Slice has an end-to-end acceptance contract.

### Layer
A **Layer** is a horizontal architectural responsibility. A layer owns one coherent kind of logic and exposes explicit interfaces/contracts to adjacent layers.

The accepted architecture/refactor design decides the actual layer set. Do not create new layers opportunistically during implementation.

### Architecture Pre-study
An **Architecture Pre-study** is a target-architecture design activity performed before a broad code review when the desired logical layering itself needs deliberate thought.

It answers what layers we *want*, what each should own, what dependency direction should exist, and what boundaries should be explicit. It deliberately avoids deep implementation review. The later Broad Review measures the current code against this target model and determines how to migrate toward it.

### Layer Boundary Contract (BC)
A **Boundary Contract** describes how two layers communicate and lives at:

`SDP/LayerBoundaryContracts/BC--<LayerA>--<LayerB>.md`

A BC describes ownership, inputs/outputs, state authority, lifecycle, error semantics, concurrency/event rules, safety invariants and forbidden responsibilities.

Boundary contract states are `DRAFT`, `REVIEWED`, `FROZEN`, and `SUPERSEDED`.

A `FROZEN` BC is implementation authority. Layer workers may rely on it but may not edit it. A change requires an explicit contract-change finding and Master/Integrator approval, followed by affected re-review/re-verification.

## 2. Source of truth and handoff

The repository is the source of truth. Chat history is not.

Every significant work item must be reconstructable from:

- GitHub Issue;
- branch/ref and exact commit SHA;
- accepted SDP architecture/review/design/verification records;
- frozen boundary contracts;
- tests/CI evidence;
- PR discussion and final disposition.

At handoff, record at minimum repository/Issue, role, branch/exact HEAD, owned layer(s), Slice IDs, contracts read/changed, tests/results, unresolved findings/risks, and exact next permitted action.

## 3. Issue-driven work

No implementation starts without an Issue.

### 3.1 Small isolated Issue
Use this when the change is bounded, architecture is already clear and no cross-layer redesign is required.

1. Read Issue + mandatory instructions.
2. Branch from the current release/integration line as `issue-#<N>/<slug>`.
3. Implement only Issue scope.
4. Add/update tests and docs.
5. Obtain independent review when practical.
6. Open PR back to the current release line.
7. Run exact-head CI and merge gate.
8. Increment patch/release number according to project policy.

Escalate to the review/refactor path if implementation reveals architectural ownership ambiguity, cross-layer state races, repeated duplication or contract changes.

### 3.2 Code Review / Refactor Program
A dedicated Issue is the parent authority. Stages are separated so session/tool limits do not force shallow work.

The default sequence is:

1. Architecture Pre-study when target layering is not already accepted.
2. Broad Code Review against the accepted pre-study target model.
3. Focused horizontal Layer Reviews.
4. Refactor Design.
5. Boundary Contract review/freeze.
6. Layer-local Slice implementation in parallel.
7. Serial integration.
8. Verification.
9. Product PR/release-line integration.

## 4. Architecture Pre-study stage

Use a Pre-study when a large review/refactor needs an explicit target logical architecture before examining implementation details deeply.

The Pre-study is **design, not code review**. It may inspect repository structure and representative files to understand the product, but should not produce exhaustive defect findings or implementation prescriptions.

It should answer:

- What logical horizontal layers should the product have?
- What responsibility belongs exclusively to each layer?
- What responsibilities are explicitly forbidden in each layer?
- What dependency direction should exist?
- What state is authoritative in which layer?
- Which layer owns process/resource lifecycle, persistence, Git mutation, application orchestration, UI interaction state and rendering?
- Which boundaries need explicit BCs?
- What physical top-level/package layout would make layer ownership obvious and parallel work safe?
- Where should vertical Slice orchestration live?
- Which shared integration/composition files must remain Master-owned?
- What architecture questions remain intentionally open for Broad Review evidence?

The Pre-study should prefer **separate physical directories/packages per horizontal layer**. A production code file should belong to one layer. Cross-layer behavior belongs in explicit orchestration and BC-defined interfaces, not files that mix horizontal responsibilities.

Report location:

`SDP/Architecture/AS-#<Issue>/APS--<seq>.md`

The report ends with:

- proposed target layer map;
- responsibility and forbidden-responsibility table;
- target physical directory/package map;
- allowed dependency graph;
- state/lifecycle authority map;
- candidate BC set;
- architecture questions for Broad Review;
- exact next permitted action.

The Pre-study does **not** freeze BCs. It creates a target architecture hypothesis that Broad Review must challenge with code evidence.

## 5. Broad code review stage

The Broad Review maps the whole current system against the accepted Architecture Pre-study and identifies architectural problems without trying to deeply inspect every horizontal layer.

It must distinguish:

- **Target logical layers** from the Pre-study;
- **current de-facto layers** found in code;
- **migration gaps** between them.

It should answer:

- Which files/packages mix responsibilities from multiple target layers?
- Which target responsibilities are missing, duplicated, or in the wrong layer?
- Where is state authority ambiguous or duplicated?
- Which async/event paths can overwrite newer state?
- Where do process/resource lifecycles cross boundaries incorrectly?
- Which interfaces leak implementation details?
- Which vertical Slice failures are symptoms of horizontal design problems?
- Which target boundaries/contracts need confirmation or revision?
- What focused Layer Reviews are required next?

The Broad Review may identify vertical and horizontal findings, but must not consume the whole session doing exhaustive line-level review inside one layer.

Report location:

`SDP/Reviews/CR-#<Issue>/Broad/BR--<seq>.md`

It ends with finding IDs, current-vs-target layer map, migration problem map, focused-review plan, preliminary boundary changes, release-line recommendation and exact next permitted action.

## 6. Focused horizontal Layer Reviews

After Broad Review, inspect one target horizontal layer per focused review session.

Report location:

`SDP/Reviews/CR-#<Issue>/Layers/<Layer>/LR--<seq>.md`

Each Layer Review covers responsibility/cohesion, physical scope, logic to move in/out, state authority, concurrency/event/lifecycle concerns, error/safety semantics, tests/observability gaps, required boundary changes, and affected vertical Slices.

A Layer Review does not implement fixes unless explicitly authorized.

## 7. Refactor Design stage

When required Layer Reviews are accepted, create:

`SDP/Design/CR-#<Issue>/REFDES--<seq>.md`

The Refactor Design consumes the Pre-study, Broad Review and all Layer Reviews and defines:

- final horizontal layer map;
- target physical directory ownership;
- files to split/move/remove;
- dependency direction;
- layer APIs/interfaces;
- final BC set;
- Slice migration plan;
- integration order;
- verification strategy;
- release-line decision.

A physical implementation file must belong to one horizontal layer. Vertical orchestration belongs in a designated application/orchestration layer. Cross-layer communication occurs only through explicit interfaces/contracts.

After design review, create/update BC files and freeze those required for the chosen implementation round.

## 8. Choosing implementation scope

Before implementation begins, Master/Steering selects one Slice, a bounded subset, or all planned refactor Slices. Record the decision in the parent Issue/design and do not silently expand it.

## 9. Refactor branches and parallel sessions

Canonical integration branch:

`codereview-#<Issue>/refactor`

Create it from the chosen release-line base after design/contracts are accepted.

### 9.1 Per-layer worker branches
Parallel sessions MUST NOT all push directly to the integration ref.

Each layer receives a dedicated branch/worktree:

`codereview-#<Issue>/layer-<layer-name>`

Rules:

- a worker owns its assigned layer directory plus explicitly authorized layer-local tests/docs;
- it does not modify another layer's implementation files;
- it does not modify frozen BC files;
- insufficient BCs cause a contract-change finding, not a workaround;
- shared composition/integration files are Master-owned unless explicitly delegated.

### 9.2 Slice commits
When implementing multiple Slices, commit each completed layer/Slice contribution separately, e.g.:

`refactor(layer:tui slice:SLC-04): route focus through UI state authority`

`refactor(layer:process slice:SLC-02): terminate launch process groups`

Do not combine unrelated layer/Slice work.

### 9.3 Integration
Only Master/Integrator merges/cherry-picks accepted layer commits into `codereview-#<Issue>/refactor`.

For every integration step record source branch, exact source SHA, Slice(s), BC version/state, integration SHA and tests/gates. Conflicts between layer implementation files are treated as architecture/ownership defects, not merely merge inconvenience.

## 10. Review of implementation

Preferred flow: Master assigns Worker per layer; Worker implements bounded Slices; fresh Reviewer evaluates exact frozen Worker HEAD independently; Worker corrects; Reviewer re-reviews; Master integrates accepted product.

For ChatGPT-only work, an implementing session should request independent review when another instance is available. If unavailable, mark self-review explicitly and require stronger integration verification.

Reviewer must inspect source/tests/contracts/exact SHA and never treat Worker/Master summaries as evidence.

## 11. Verification stage

Freeze integrated refactor HEAD and verify:

- unit/integration tests;
- race/concurrency tests where relevant;
- supported platform builds;
- Slice acceptance contracts;
- BC conformance;
- forbidden dependency checks;
- process/resource cleanup;
- state/focus/event ordering;
- migration/backward compatibility.

Artifacts live under `SDP/Verification/CR-#<Issue>/`.

No product PR is opened until the accepted verification gate is complete unless the Issue explicitly authorizes an earlier draft integration PR.

## 12. Product PR and release line

After verification, PR `codereview-#<Issue>/refactor` to the chosen release line.

Stay on the current minor line when external behavior remains compatible, migration risk is bounded, and internal restructuring can be safely delivered as a compatible evolution.

Open the next minor architecture line when restructuring is broad, interfaces/state/lifecycle semantics change materially, several Slices migrate together, or the current line should remain available for small fixes during stabilization.

Steering/Master records the release-line decision before implementation branches are created.

## 13. Layer directory policy

Long-term target: production code is physically grouped by horizontal layer.

Do not perform cosmetic moves before responsibility and BC design. Once accepted:

- files in a layer directory implement that layer only;
- vertical orchestration lives in the designated orchestration/application layer;
- cross-layer imports follow documented dependency direction;
- CI should enforce forbidden dependencies where practical.

## 14. Boundary-contract change control

If implementation discovers a frozen BC is wrong:

1. stop affected cross-layer implementation;
2. record/open a BC-CHANGE finding;
3. update design impact analysis;
4. review proposed BC change;
5. freeze revised BC;
6. notify/rebase affected layer workers;
7. re-run affected verification.

Never silently work around a frozen BC.

## 15. Safety and quality gates

- exact HEAD/SHA matters;
- do not merge failing required CI;
- do not hide environment/test limitations;
- destructive Git/worktree operations require explicit safe contracts;
- uncommitted user work must never be silently discarded;
- process/session cleanup must be deterministic;
- async completion must not overwrite newer user intent unless the owning contract explicitly permits it;
- reviews distinguish observed facts, inferred risks and recommendations.

## 16. Current gh-tree stabilization sequence

For the post-v0.3.14 architecture program, the intended sequence is:

1. accept/merge this governance bootstrap;
2. run an Architecture Pre-study to establish the desired target horizontal layers, physical folder layout and candidate BC graph;
3. run the dedicated Broad Code Review against v0.3.14 using that target model;
4. open/perform focused Layer Reviews;
5. create Refactor Design and freeze the required BC set;
6. select implementation Slices and perform parallel layer-local implementation;
7. integrate, verify and choose the product release line.