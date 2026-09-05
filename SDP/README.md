# SDP

This directory contains durable Software Development Process records for `gh-tree`.

The SDP exists so that architecture, review findings, contracts, verification evidence and handoff state survive across different ChatGPT/Codex sessions without depending on chat history.

## Structure

- `Architecture/` — target-architecture pre-studies performed before broad code review when desired layering is not yet accepted.
- `Reviews/` — broad code reviews and focused horizontal layer reviews.
- `Design/` — accepted refactor/design records produced from architecture and review evidence.
- `LayerBoundaryContracts/` — explicit contracts between horizontal layers.
- `Verification/` — verification plans, exact-head evidence and final dispositions.
- `Handoffs/` — optional bounded handoff records when a program spans multiple sessions.
- `Templates/` — durable templates for architecture/review records.

The operative workflow is defined in [`../developmentInstructions.md`](../developmentInstructions.md).

## Program naming

For an architecture pre-study Issue `#22` and later code-review Issue `#21`, records may look like:

```text
SDP/
  Architecture/
    AS-#22/
      APS--001.md
  Reviews/
    CR-#21/
      Broad/
      Layers/
        TUI/
        Process/
        Git/
  Design/
    CR-#21/
  Verification/
    CR-#21/
```

Use exact Issue numbers and exact product/refactor SHAs in documents.

## Record state

Every durable SDP record should declare its state near the top, for example:

- `DRAFT`
- `IN_REVIEW`
- `ACCEPTED`
- `CHANGES_REQUIRED`
- `VERIFIED`
- `SUPERSEDED`

Boundary contracts use the more specific states described in `LayerBoundaryContracts/README.md`.

## Evidence rule

Architecture Pre-study is explicitly target-design/hypothesis work and should mark assumptions. Broad/focused reviews and verification must cite the code/contracts/tests/SHAs they actually inspected. Implementation summaries are never review evidence.
