# Development Instructions

This file defines the default development, review, refactor and integration process for `gh-tree`.

The purpose is to make work reproducible across ChatGPT/Codex sessions, to keep architecture visible in the repository instead of in chat memory, and to support parallel implementation without avoidable merge conflicts.

## 1. Terminology

### Slice
A **Slice** is a vertical unit of user-visible or externally observable behavior that may cross several horizontal architectural layers.

Examples: deploy a selected commit into Active worktree; stop a launch process; render local-vs-upstream commit state.

A Slice should have an end-to-end acceptance contract.

### Layer
A **Layer** is a horizontal architectural responsibility. A layer owns one coherent kind of logic and exposes explicit interfaces/contracts to adjacent layers.

Examples may include domain/model, Git/worktree operations, process/session runtime, application/service orchestration, TUI state/navigation, rendering/presentation, persistence/configuration.

The accepted refactor design decides the actual layer set. Do not infer or create new layers opportunistically during implementation.

### Layer Boundary Contract (BC)
A **Boundary Contract** describes how two layers communicate. Contracts live in:

`SDP/LayerBoundaryContracts/BC--<LayerA>--<LayerB>.md`

A BC describes ownership, inputs/outputs, state authority, lifecycle, error semantics, concurrency/event rules and forbidden responsibilities.

Boundary contracts have states:

- `DRAFT` — design may still change.
- `REVIEWED` — reviewed but implementation has not started.
- `FROZEN` — implementation authority. Layer workers may rely on it but not change it.
- `SUPERSEDED` — replaced by a newer accepted contract.

Changing a `FROZEN` contract requires an explicit contract-change finding/Issue and Master approval. The affected layer implementations must then be re-verified.

## 2. Source of truth and handoff

The repository is the source of truth. Chat history is not.

Every significant work item must be reconstructable from:

- GitHub Issue;
- branch/ref and exact commit SHA;
- SDP review/design/verification records;
- frozen boundary contracts;
- tests/CI evidence;
- PR discussion and final disposition.

At handoff, record at minimum:

- repository and Issue;
- role performed;
- branch and exact HEAD;
- changed/owned layer(s);
- Slice ID(s);
- contracts read/changed;
- tests run and result;
- unresolved findings/risks;
- exact next permitted action.

## 3. Issue-driven work

No implementation starts without an Issue.

There are two primary paths.

### 3.1 Small isolated Issue

Use this when the change is bounded, architecture is already clear and no cross-layer redesign is required.

Flow:

1. Read Issue + required instructions.
2. Branch from the current release/integration line:
   `issue-#<N>/<slug>`
3. Implement only Issue scope.
4. Add/update tests and docs.
5. Request/perform independent review when practical.
6. Open PR back to the current release line.
7. Run exact-head CI and merge gate.
8. Increment the patch/release number according to the project release policy.

A small Issue must be escalated into the review/refactor path if implementation reveals architectural ownership ambiguity, cross-layer state races, repeated duplication or contract changes.

### 3.2 Code Review / Refactor Program

A broad review is triggered by a dedicated Issue. The review Issue is the parent authority for the program.

The stages are intentionally separated so that tool/session limits do not force shallow review.

## 4. Broad code review stage

The Broad Review maps the whole system and identifies architectural problems without trying to deeply inspect every horizontal layer.

It should answer:

- What are the current de-facto layers?
- Which files mix responsibilities from multiple layers?
- Where is state authority ambiguous or duplicated?
- Which async/event paths can overwrite newer state?
- Where do process/resource lifecycles cross layer boundaries incorrectly?
- Which current interfaces leak implementation details?
- Which vertical Slice failures are symptoms of horizontal design problems?
- Which logic appears to belong in another layer?
- Which boundaries/contracts are missing or need change?
- What focused layer reviews are required next?

The Broad Review may identify both:

- **Vertical findings** — a Slice is broken end-to-end.
- **Horizontal findings** — a layer has weak cohesion/ownership or a boundary is wrong.

It should not consume the whole session doing exhaustive line-level review inside one layer.

Report location:

`SDP/Reviews/CR-#<Issue>/Broad/BR--<date-or-seq>.md`

The report ends with:

- findings list with IDs;
- proposed layer map;
- focused-review plan;
- preliminary boundary changes;
- recommended release-line strategy.

## 5. Focused horizontal layer reviews

After Broad Review, inspect one horizontal layer per focused review session.

Each focused review gets enough time/token budget to go deep inside that layer while treating adjacent layers through their current/proposed interfaces.

Report location:

`SDP/Reviews/CR-#<Issue>/Layers/<Layer>/LR--<seq>.md`

Each Layer Review covers:

- responsibility/cohesion;
- files/functions currently belonging to the layer;
- logic that should move out;
- logic missing from the layer;
- state ownership and mutation rules;
- concurrency/event/lifecycle concerns;
- error/safety semantics;
- tests and observability gaps;
- required incoming/outgoing boundary changes;
- Slice impacts.

A Layer Review does not implement fixes unless the parent Issue explicitly combines review and repair for that layer.

## 6. Refactor design stage

When all required Layer Reviews are accepted, create a Refactor Design that consumes the Broad Review plus every Layer Review.

Report/design location:

`SDP/Design/CR-#<Issue>/REFDES--<seq>.md`

The Refactor Design defines:

- final horizontal layer map;
- target physical directory ownership;
- which existing files are split/moved/removed;
- dependency direction;
- layer APIs/interfaces;
- boundary contract set;
- Slice migration plan;
- integration order;
- verification strategy;
- release-line decision.

The design must prevent implementation files from becoming vertical cross-layer grab-bags. A physical code file should belong to one horizontal layer. Cross-layer communication happens through explicit interfaces/contracts.

After design review, create/update the files in `SDP/LayerBoundaryContracts/` and freeze the contracts required for the chosen implementation round.

## 7. Choosing implementation scope

Before implementation begins, Master/Steering selects which Slices are in the round:

- one Slice;
- a bounded subset;
- all planned refactor Slices.

This decision is recorded in the parent Issue or accepted Refactor Design.

Do not silently expand the implementation round.

## 8. Refactor branches and parallel sessions

The canonical refactor integration branch is:

`codereview-#<Issue>/refactor`

It is created from the chosen release-line base after the design/contracts are accepted.

### 8.1 Per-layer worker branches

Parallel workers/sessions MUST NOT all push directly to the same integration ref.

Each horizontal layer receives a dedicated branch/worktree, for example:

`codereview-#<Issue>/layer-<layer-name>`

Rules:

- worker owns only its assigned layer directory plus explicitly authorized layer-local tests/docs;
- worker does not modify another layer's implementation files;
- worker does not modify frozen BC files;
- if a BC is insufficient, worker stops and raises a contract-change finding instead of bypassing it;
- shared top-level integration files are owned by Master/Integrator unless explicitly delegated.

This minimizes merge conflicts and makes concurrent work safe.

### 8.2 Slice commits

If all Slices are being implemented, each layer worker commits after each completed Slice contribution.

Commit message must identify both layer and Slice, for example:

`refactor(layer:tui slice:SLC-04): route focus through UI state authority`

or

`refactor(layer:process slice:SLC-02): terminate launch process groups`

Do not combine unrelated layer/Slice work in one commit.

### 8.3 Integration

Only Master/Integrator merges/cherry-picks accepted layer commits into:

`codereview-#<Issue>/refactor`

Integration is serial and contract-aware even when implementation is parallel.

For each integration step Master records:

- source layer branch;
- exact source SHA;
- Slice(s);
- BC version/state;
- integration SHA;
- tests/gates run.

If two layer branches conflict in implementation files, that is treated as an architecture/ownership defect, not merely a merge inconvenience.

## 9. Review of implementation

Preferred Codex flow:

- Master assigns Worker per layer;
- Worker implements bounded Slice(s);
- fresh Reviewer evaluates exact frozen Worker HEAD independently;
- Worker corrects findings;
- Reviewer re-reviews exact corrected HEAD;
- Master integrates accepted product.

ChatGPT-only flow:

- session may implement its assigned layer;
- it should request an independent review when another instance is available;
- if independent review is unavailable, clearly mark self-review and require stronger integration verification before release.

Reviewer must never accept Worker/Master summary as evidence. Review source, tests, contracts and exact SHA directly.

## 10. Verification stage

When all selected layers/Slices are integrated into the refactor branch, freeze the refactor HEAD and run verification.

Verification covers:

- all unit/integration tests;
- race/concurrency tests where relevant;
- supported platform builds;
- Slice acceptance contracts;
- BC conformance;
- no forbidden cross-layer imports/dependencies;
- process/resource cleanup;
- state/focus/event ordering where applicable;
- migration/backward compatibility as required.

Verification artifacts live under:

`SDP/Verification/CR-#<Issue>/`

No product PR is opened until the accepted verification gate is complete unless the Issue explicitly requires a draft integration PR earlier.

## 11. Product PR and release line

After verification, open a PR from:

`codereview-#<Issue>/refactor`

to the chosen product release line.

Examples:

- current line: `v1.3`
- next architectural line: `v1.4`

The exact numbering is project-specific; these are examples.

### Stay on current minor line when

- external behavior remains compatible;
- internal boundaries changed but public/user contracts are preserved;
- migration risk is bounded and verification is strong.

### Open next minor line when

- architectural restructuring is broad enough that parallel stabilization is useful;
- interfaces/state ownership/lifecycle semantics change materially;
- multiple Slices are migrated together;
- the current release line should remain available for small fixes while refactor work continues.

Steering/Master records the decision before implementation branches are created.

## 12. Layer directory policy

Long-term target: implementation code is physically grouped by horizontal layer.

The Refactor Design defines the actual directories. Do not perform cosmetic moves before ownership and BCs are designed.

Once a layer layout is accepted:

- files in a layer directory implement that layer only;
- vertical orchestration belongs in a designated orchestration/application layer rather than being spread through every layer;
- cross-layer imports follow the documented dependency direction;
- CI should eventually enforce forbidden dependencies where practical.

## 13. Boundary-contract change control

If implementation discovers a frozen BC is wrong:

1. stop affected cross-layer implementation;
2. open/document a `BC-CHANGE` finding under the parent review Issue;
3. update design impact analysis;
4. review the proposed BC change;
5. freeze the revised BC;
6. notify/rebase affected layer workers;
7. re-run affected verification.

Never make an implementation workaround that silently violates the BC.

## 14. Safety and quality gates

For every path:

- exact HEAD/SHA matters;
- do not merge with failing required CI;
- do not hide test/environment limitations;
- destructive Git/worktree operations require explicit safe contracts;
- user data/uncommitted work must never be silently discarded;
- process/session cleanup must be deterministic;
- async completion must not overwrite newer user intent unless the owning contract explicitly permits it;
- reviews distinguish observed facts, inferred risks and recommendations.

## 15. Recommended next step for current gh-tree state

After this governance bootstrap is merged, open a dedicated Broad Code Review Issue against the current product HEAD.

That review should start with the structural symptoms already observed around:

- layered `V310`..`V314` TUI wrappers and overlapping state authority;
- console/process-tree lifecycle and stop semantics;
- async focus ownership;
- local vs upstream branch/commit state;
- current package/directory boundaries.

The Broad Review should define the proposed layer map before any large refactor begins.