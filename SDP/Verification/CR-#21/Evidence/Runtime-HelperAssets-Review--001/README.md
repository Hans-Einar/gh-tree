# Independent helper/assets review controls

Source: frozen d110a068db8ea68cb2d848774d79b7b9d71b0557, technical b72ad94c60eb38d7c4ebe70fafd10b21a3289037.
Disposition and evidence limits: `SDP/Implementation/CR-#21/Runtime/M3-HelperAssets-Review--001.md`.

These small reviewer-authored controls use the actual helpergen functions through
a Go overlay. No product source is changed. The `.go.txt` suffix keeps evidence
outside Go package discovery. Logs distinguish a passing checksum-admission
control from two regression failures. Earlier exploration initially asserted that
capture would accept a forged cache pin; actual `go list` rejected it, so the
preserved final control correctly expects rejection and that hypothesis is not a
review finding.

The local owned fixture root is
`C:/Users/hanse/.codex/tmp/cr21-helper-review-001`. Copy the three source scripts
there before replay. `prepare_probes.py` creates isolated tiny pinned/forged
module caches and synthetic broker entries, writes the overlay, and runs the
empty-cache exact checker control. The required owned toolchain junction was
created with PowerShell:

```powershell
New-Item -ItemType Junction -Path C:/Users/hanse/.codex/tmp/cr21-helper-review-001/go-link -Target C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64
```

`run_record.py` asserts the frozen HEAD before observing the real canonical
checker and executing the overlay tests. It also retrieves compact run/job proof.
Its paths deliberately identify the exact reviewed worktree/toolchain and owned
fixture root; use a separate frozen worktree and adjust those constants for
independent replay after this report commit. Do not alter the product candidate
or any shared module cache. The live-input control restores its owned module
source in a defer, also before final recapture; it does not modify module metadata
or the original go.sum pin. It performs two actual fresh builds per target.

Separate observed native commands were the full three-package helper/assets/
architecture test run, actual Windows386 asset tests and expected windows/386
generator refusal. The direct-root source-closure test passes. With GOROOT set
to the owned junction, the same source-closure test fails identically to CI.
The small direct `contained` control in the preserved source isolates this cause.
Their exact results and scope are in the review report; they are not fabricated
as additional raw logs here.

CI diagnostic files are selected raw timestamped context, labeled as excerpts.
Canceled attempt 1 and report-only CI are not successful evidence. The manifest
hashes evidence files, excluding itself. No generated executable, toolchain,
module-cache tree, original source archive or unrelated review is duplicated.

## Final bounded confirmation

The `final-` evidence belongs to frozen4d59b27 / technicalb6f161f; the manifest's
`currentReview` is its bounded ACCEPT. Earlier failed evidence remains unchanged.
For replay, place `final-acceptance_test.go.txt` at the new owned fixture path
`C:/Users/hanse/.codex/tmp/cr21-helper-review-003/reviewer_acceptance_test.go.txt`
and `final-observer.py` at `.../record_review.py`. Use the frozen worktree or
adjust that script's explicit REPO constant to another checkout of the same SHA.
Run `record_review.py controls` for the independent overlay plus focused native
controls, and `record_review.py check` for the observed exact checker. Both
commands use explicit UTF-8 and assert the source SHA before execution.

`final-controls.log` is direct captured process output. Its corresponding JSON
records command/exit/time. The independently adapted controls distinguish a
compiler producing an image from the generator accepting it: successful real
amd64/ARM64 PE images containing inserted source are rejected by permanent
invalidation. The 64-cycle control completes insertion/removal before release,
then requires rejection through cancellation/join. `final-check.json` captures
the 60-file hashes/lengths/mtime/file-set no-rewrite assertion. `final-ci.json`
records the exact technical20-job successful run. No new module cache or junction
is required for this final confirmation; the earlier fresh-cache/junction proof
remains in `recheck-` evidence.
