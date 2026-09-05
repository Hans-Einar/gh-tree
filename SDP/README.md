# SDP

This directory contains durable Software Development Process records for `gh-tree`.

The SDP exists so that architecture, review findings, contracts, verification evidence and handoff state survive across different ChatGPT/Codex sessions without depending on chat history.

## Structure

- `Reviews/` — broad code reviews and focused horizontal layer reviews.
- `Design/` — accepted refactor/design records produced from review findings.
- `LayerBoundaryContracts/` — explicit contracts between horizontal layers.
- `Verification/` — verification plans, exact-head evidence and final dispositions.
- `Handoffs/` — optional bounded handoff records when a program spans multiple sessions.

The operative workflow is defined in [`../developmentInstructions.md`](../developmentInstructions.md).

## Review program naming

For a review Issue `#19`, records should normally live under a program path such as:

```text
SDP/
  Reviews/
    CR-#19/
      Broad/
      Layers/
        TUI/
        Process/
        Git/
  Design/
    CR-#19/
  Verification/
    CR-#19/
```

Use exact Issue numbers and exact product/refactor SHAs in the documents.

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

Implementation summaries are not evidence. Reviews and verification should cite the code/contracts/tests/SHAs they actually inspected.
