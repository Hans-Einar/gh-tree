# Layer Boundary Contracts

Layer Boundary Contracts (BCs) are the implementation authority for communication between horizontal architectural layers.

Naming:

`BC--<LayerA>--<LayerB>.md`

Example:

`BC--TUI--Application.md`

## Required fields

Each BC should contain at minimum:

- Contract ID/name
- State: `DRAFT | REVIEWED | FROZEN | SUPERSEDED`
- Parent review/refactor Issue
- Applicable refactor/release line
- Layer A responsibility
- Layer B responsibility
- Allowed dependency direction
- Inputs/commands crossing the boundary
- Outputs/events/results crossing the boundary
- State ownership: which layer is authoritative for each relevant state
- Lifecycle/resource ownership
- Concurrency/ordering rules
- Error/cancellation semantics
- Safety invariants
- Forbidden responsibilities / calls
- Tests or verification clauses that prove conformance
- Change history and superseding contract, if any

## Freeze rule

A `FROZEN` BC is read-only for ordinary layer workers.

If it proves insufficient, the worker must raise a boundary-contract change finding rather than implementing around it. The Master/Integrator owns contract-change integration and must ensure all affected layers are re-reviewed/re-verified.

## Template

```markdown
# BC--LayerA--LayerB

State: DRAFT
Parent Issue: #<N>
Applies to: <branch/release/refactor program>

## Responsibilities
### LayerA
...
### LayerB
...

## Dependency direction
LayerA -> LayerB

## Commands / Inputs
...

## Events / Outputs
...

## State authority
...

## Lifecycle ownership
...

## Ordering / concurrency
...

## Errors / cancellation
...

## Safety invariants
...

## Forbidden behavior
...

## Verification
...

## Change history
...
```
