# Broad Code Review

State: DRAFT
Parent Issue: #<N>
Repository: Hans-Einar/gh-tree
Architecture Pre-study: <path/id>
Reviewed product/ref: <branch/tag>
Frozen review SHA: <sha>
Reviewer role/session: <role>
Date: <YYYY-MM-DD>

## Review purpose
Describe the system-level review contract and exclusions. Treat the accepted Architecture Pre-study as a target hypothesis to challenge with code evidence, not as unquestionable truth.

## Target logical layer map
Summarize the desired layers, responsibilities, physical folder/package intent and dependency direction from the Architecture Pre-study.

## Current de-facto layer map
Identify horizontal responsibilities and the physical files/packages currently implementing them.

## Current-to-target migration gaps
| ID | Current location/owner | Target layer | Gap | Likely migration/refactor |
|---|---|---|---|---|

Explicitly identify files/packages that mix multiple target layers and therefore need split/move/redesign.

## Vertical Slice findings
| ID | Slice / behavior | Severity | Observation | Likely layer/boundary cause |
|---|---|---|---|---|

## Horizontal architecture findings
| ID | Layer / boundary | Severity | Observation | Required focused review |
|---|---|---|---|---|

## Ownership / state-authority findings
Identify duplicated or ambiguous state authority, async overwrites and lifecycle ownership.

## Boundary-contract candidates
Compare target candidate boundaries with evidence from current code. List missing/incorrect boundaries and preliminary contract changes. Do not freeze contracts in the Broad Review.

## Physical layer-separation findings
Assess how production code should be redistributed into separate layer directories/packages. Flag vertical cross-layer grab-bag files and shared composition files requiring Master ownership.

## Focused layer review plan
List each target layer requiring a separate deep review and why.

## Release-line recommendation
State whether the likely refactor belongs on the current minor line or should open a new minor architecture line, with rationale.

## Findings disposition
Summarize blocker/high/medium/low findings, current-vs-target architecture status, and the exact next permitted action.
