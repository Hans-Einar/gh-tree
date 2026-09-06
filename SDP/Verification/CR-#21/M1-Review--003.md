# M1 independent source acceptance

Disposition: ACCEPT at 05982a4adb35e39d7b7ba371e0d16d83b2b3674c.
Authority: #57 under #21; I-03 / M1. Reviewer: m1_composition_review, separate
from author and Master. Contribution base: 0b91d92506434d1079fd715b623d8aa0eab8a76f.
All M157-M01..M06 are resolved; no remaining or new blocking finding.

## Independent evidence

Reviewer reopened actual complete source, corrections, workflows and report at
the frozen SHA; the clean contribution contained12 authorized files. Existing
45 JSON fixtures plus standalone regressions passed in39.583s, with vet/gofmt
clean and all12 selected-source/export checks passing with61 exact allowances.

Ten retained independent type fixtures now produce their required results:
private plumbing/pure methods pass; nested callbacks and Git/Runtime-broker/
Application-usecase wrapper escapes reject. Actual retained outside child junction
rejects and selected-root junction passes. Path-specific +text/-text, all three
legacy crlf forms and configured-filter controls produce expected acceptance or
explicit refusal without creating a filter-execution marker. Actual local Windows
root/outside/cross-layer/legacy junction regressions run without a privilege skip.

Reviewer inspected the Windows metadata-handle path inspector and checked native
flags/signature against pinned x/sys0.44.0. CloseHandle is deferred on every
post-open return and its error propagates. Actual Windows386 execution also passed
selected-root alias and outside-child rejection. The inspection is tool-local,
not a Runtime/Persistence acquisition protocol or concurrent-tree safety claim.

A first uncontrolled process-wide handle count106->119 was inconclusive because
Go thread initialization changed the process. It is retained in the evidence.
The controlled probe sets GOMAXPROCS2, warms500 directory/500 file/1000 rejected
inspections, then repeats five such rounds. Counts stayed95 handles/6 threads
after every round. This establishes the bounded inspector behavior tested; it is
not product Runtime cleanup or cross-platform resource proof.

The exact tested checker SHA256 is
D42FF68986597A48A0BE8DE4805E9C21AC2780282914F578F625691BEA5EDC83.
Master read all actual retained results and handle-probe source and verified its
binary/archive hashes. [Archive manifest](Evidence/M1-Review--003/ArchiveManifest.json)
preserves five native source/result captures. Earlier failed assumptions and
findings remain in M1-Review--001/002 and their archives.

## CI and serial integration

Master independently revalidated [source CI34031420582](https://github.com/Hans-Einar/gh-tree/actions/runs/34031420582)
at exact05982a4: all18 applicable jobs SUCCESS, sole M3 helper conformance SKIPPED
because its inputs do not exist. Jobs cover four native test/vet/build profiles,
Linux race, all12 build/architecture targets and accepted inventory. Windows native
junction tests have no skip branch; the native suites execute them. M3 helper/ABI/
Runtime and full vertical behavior remain unverified.

After this independent ACCEPT and current source CI, Master fast-forwarded the
clean canonical codereview-21/refactor branch from0b91d925 to exact05982a4 and
pushed it. Source and integration SHA are identical. The separately triggered
integration-branch CI run34031833446 completed successfully on the same SHA:
18 SUCCESS and the sole explicitly deferred M3 helper SKIP. Master verified the
actual push event and codereview-21/refactor branch. M1 is complete; #58 authorizes
the next Domain contribution after its Master kickoff checkpoint.

No baseline product finding or complete Slice is resolved by this infrastructure
gate. All143 baseline findings, M2..M8 implementation/verification, final product
PR/main and full twelve-asset release/install-upgrade requirements remain active.
No main product PR, tag, publication or user installation change occurred.
