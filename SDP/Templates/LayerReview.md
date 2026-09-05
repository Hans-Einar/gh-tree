# Focused Layer Review — <Layer>

State: DRAFT
Parent Review Issue: #<N>
Broad Review: <path/id>
Layer: <name>
Frozen review SHA: <sha>
Reviewer role/session: <role>
Date: <YYYY-MM-DD>

## Layer responsibility
State what this layer should own and what it must not own.

## Physical scope reviewed
List packages/directories/files inspected.

## Cohesion / ownership findings
| ID | Severity | File/function | Finding | Recommended owner/layer |
|---|---|---|---|---|

## State authority
Document authoritative state, mutations, duplicate caches and stale/async overwrite risks.

## Lifecycle / concurrency
Document resource ownership, start/stop/cancel behavior, ordering and race risks.

## Error and safety semantics
Document error propagation, destructive-operation guards and recovery behavior.

## Incoming boundary changes required
Describe changes needed from adjacent layers.

## Outgoing boundary changes required
Describe changes this layer should expose to adjacent layers.

## Slice impact
Map findings to affected vertical Slices.

## Tests / observability gaps
List missing unit/integration/race/platform coverage.

## Disposition
`ACCEPTED | CHANGES_REQUIRED | BLOCKED`

State the exact next permitted action.