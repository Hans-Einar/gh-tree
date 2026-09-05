# Program traceability

Issue #21 / #35 authorizes this minimal gh-tree traceability surface.
CurrentIndex.yaml is the current state; Relations.yaml links authority, reports
and gates; Ledger.ndjson is append-only event history. Each ledger line is one
JSON object with eventId, ts, role, action, sprint, iteration, slice, links,
filesPlanned/filesChanged, outcome, status and notes as relevant.

Use corrections/follow-up events instead of rewriting prior ledger events.
Record exact source/review/integration SHA and evidence links for each gate.
An accepted report describes open product findings; it does not resolve them.
The full original user run contract lives in the sprint folder.
