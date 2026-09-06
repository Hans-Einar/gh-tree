# M1 initial independent implementation review

Disposition: CHANGES_REQUIRED. Authority: #57 under #21; I-03 / M1.
Reviewed source: 6685b4f76ca071be0125289925e7c9aa466dc10a.
Base: 47ad595703b2d779dd21d62630d47a5683b754c6.
Independent reviewer: m1_composition_review; separate from Worker and Master.
Review scope: two workflows, Composition architecture checker/tests/README and
bounded worker report; nine files. No product entry cutover or legacy retirement.

## Findings

| ID / severity | Actual evidence at reviewed source | Required correction |
|---|---|---|
| M157-M01 / MEDIUM (P2) | macOS CI job101475414670 in run34029160810 failed because Go reports temporary roots through /private/var while fixtures supplied /var. Lexical relative-path checking refused the same physical repository. | Canonicalize the actual root consistently; real alias-root regression on supported native systems, preserving outside-root rejection. |
| M157-M02 / MEDIUM (P2) | check.go publicTypes signature traversal accepted API Result.Callbacks as []func() and a map[string]func() return, including a Windows ARM64-only source. Actual checker exit0. []any negative control refused. | Track callback position throughout nested field/argument/result type graphs; legitimate declared API functions/methods stay callable. |
| M157-M03 / MEDIUM (P2) | check.go applied public DTO rules to private Git native helpers. OpenLock returning *os.File solely to its owning Git root incorrectly refused, while equivalent private Runtime broker type passed. | Permit accepted private native decomposition; recursively reject private helper-defined native/callback wrapper values that escape through a published owner boundary. |
| M157-M04 / MEDIUM (P2) | Unconditional CRLF removal accepted61 exemptions after path-specific -text/core.autocrlf=false changed the actual clean blob of internal/app/service.go to f9c880865942fcf2d20e179bf941f34751b84174, instead of accepted4f8be20c2beff7db3f180a102fe146310c96a3aa. +text was a positive control. | Respect actual path-specific Git clean semantics or refuse unsupported profiles; retain legitimate CRLF without overlooking a changed blob or implicitly executing arbitrary filters. |
| M157-M05 / MEDIUM (P2) | go run -trimpath and a trimpath-built standalone checker, with no explicit GOROOT, rejected standard testing as forbidden. Explicit actual GOROOT made the same source pass. | Resolve standard-library ownership from the actual selected Go toolchain metadata, not the checker binary's path-bearing embedded GOROOT. |

The reviewer ran existing checker fixtures, vet, formatting, all12 checker
selections and the native Windows full suite against the initial frozen source;
these passed. Independent additional compiled fixtures exposed the findings above.
The tested standalone checker SHA256 was
5AB4FFFD71768B045E6D7F7D4556D30B2A20762F66142836BCBC710A40BA9646.
After authoring resumed for the CI correction, additional probes used that retained
exact binary and an owned frozen source snapshot; no candidate files were edited
by the reviewer. Release staging had no publishing/clobber route.

## Actual CI and archived evidence

[CI34029160810](https://github.com/Hans-Einar/gh-tree/actions/runs/34029160810)
at exact6685b4f completed17 SUCCESS,1 FAILURE (macOS),1 SKIPPED (M3 Runtime helper
inputs absent). Successful jobs include all12 builds/architecture checks, Linux
race and native Linux/Windows amd64/Windows ARM64 tests. The M3 prerequisite
classification is distinct from native helper conformance, which did not run.

Master read actual source/workflows, the macOS failed job log and the independent
probe source/results. Exact temporary probe source and three result files are
archived under [Evidence/M1-Review--001](Evidence/M1-Review--001/ArchiveManifest.json),
with source/destination hashes and preserved bytes. They are review fixtures, not
product implementation or proof that a correction passes. The compiled checker
and complete frozen source ZIP remain temporary; their absence from the repository
does not change the identified source SHA or archived fixture inputs/results.

## Next permitted action

Worker corrects M157-M01..05 with actual regressions, records failures/limits and
freezes a new exact source. Separate reviewer reopens the complete corrected
candidate and checks private-wrapper escapes as well as the original negative
controls. All current applicable branch CI must pass before Master integration.
No acceptance, integration, product-finding closure, completed Slice or release
follows from this first candidate. All143 baseline product findings remain open.
