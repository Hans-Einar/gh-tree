# AGENTS.md

This repository uses an issue-driven SDP development process designed for safe handoff between ChatGPT/Codex sessions and for parallel implementation without hidden conversational state.

## Mandatory reading order

Before reviewing, designing, implementing, testing, merging, or releasing changes, every agent/session MUST read:

1. this `AGENTS.md`;
2. [`developmentInstructions.md`](developmentInstructions.md) completely;
3. the full GitHub Issue that authorizes the work;
4. all SDP review/design/boundary-contract documents referenced by that Issue;
5. any layer-local README/AGENTS instructions for files the agent is allowed to change.

Conversation summaries are navigation aids only. Repository state, Issues, accepted SDP records, frozen boundary contracts, tests and exact commit SHAs are the authority.

## Core rules

- No implementation without an Issue that defines scope and authority.
- Small isolated fixes use an issue branch and PR.
- Broad review/refactor programs follow the staged review process in `developmentInstructions.md`.
- A **Slice** is vertical behavior through multiple horizontal layers.
- A **Layer** is a horizontal architectural responsibility with explicit interfaces to adjacent layers.
- During parallel refactor implementation, a worker/session owns one layer folder unless the accepted refactor design explicitly says otherwise.
- Cross-layer behavior changes require an accepted Layer Boundary Contract change before implementation.
- Frozen boundary contracts MUST NOT be edited opportunistically by layer workers.
- Dirty worktrees, unpublished commits, test failures, review findings and unresolved uncertainties must be reported explicitly at handoff.
- Never treat implementation summaries from the authoring worker as review evidence.

## Roles

A session may act as one or more of the following only when the Issue permits it:

- **SteeringGroup** — scopes/opens Issues and selects release/refactor strategy.
- **Master / Integrator** — owns the program branch, layer integration order, contract freeze state and verification gate.
- **Worker** — implements bounded scope in one layer/slice branch.
- **Reviewer** — independently evaluates frozen products/reports; does not implement the item being reviewed.
- **Verifier** — runs the accepted verification contract against exact integrated SHAs.

When separate workers/reviewers are unavailable, a ChatGPT session may implement directly, but should request an independent review before integration whenever practical.

## Repository governance locations

- `developmentInstructions.md` — operative development workflow.
- `SDP/` — durable review/design/verification/handoff records.
- `SDP/LayerBoundaryContracts/` — contracts between horizontal layers.

If instructions conflict, the full authorizing Issue and the most recently accepted/frozen SDP governance record control, unless the user explicitly overrides them.