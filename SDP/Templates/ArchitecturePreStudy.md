# Architecture Pre-study

State: DRAFT
Parent Issue: #<N>
Repository: Hans-Einar/gh-tree
Product baseline: <tag/ref>
Baseline SHA: <sha>
Architecture session/role: <role>
Date: <YYYY-MM-DD>

## Purpose
Define the desired logical horizontal architecture before the Broad Code Review. This is target-design work, not an exhaustive code review.

## Product capabilities considered
List the major vertical user-visible Slices/capabilities that the architecture must support.

## Proposed target layer map
| Layer | Responsibility | Must not own | Notes |
|---|---|---|---|

## Target physical package/directory map
Show the intended production-code directories/packages. Prefer one horizontal layer per folder/package family.

## Dependency direction
Document the allowed layer dependency graph and any composition root.

## State authority map
| State / concept | Authoritative layer | Consumers | Mutation rule |
|---|---|---|---|

## Lifecycle/resource authority
| Resource | Owner layer | Start | Stop/cancel | Cleanup invariant |
|---|---|---|---|---|

## Vertical Slice orchestration
Describe where end-to-end Slice orchestration belongs and how it crosses layer boundaries without mixing layer implementation code.

## Candidate Layer Boundary Contracts
List candidate `BC--LayerA--LayerB.md` contracts and what each must govern. Do not freeze contracts here.

## Parallel-development ownership
Identify layer-local ownership, shared integration/composition files, and likely conflict hotspots.

## Architecture questions for Broad Review
List hypotheses and uncertainties the Broad Review must challenge against current code evidence.

## Preliminary release-line impact
State whether the target architecture appears compatible with the current minor line or likely warrants a new architecture/minor line. This is preliminary only.

## Disposition
`ACCEPTED | CHANGES_REQUIRED | BLOCKED`

State the exact next permitted action.