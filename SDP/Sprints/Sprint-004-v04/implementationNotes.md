# Implementation and verification notes

No product refactor has begun. No BC is frozen. No v0.4 verification is claimed.

Reconstruction verified main 3531fd165ff2b35036928328685d45d84965e9ef against
v0.3.14 / f626077ca0e59fbe9ede7ba1116982bb94b2eb6b: only governance/review records
changed. Initial user checkout was clean and fast-forwarded from f2e1bd05a8b123ce2234b0bce96672bbde9a6eba.

Initial PR #34 observation: one new Domain artifact, HEAD ec482795a82371996370ec52e68c5c24b4e240b1,
draft, all nine configured checks green (three platform test/vet/build jobs,
Linux race, five release-target cross-builds). These are existing baseline CI
results, not proof of future refactor behavior. Independent Domain gate review
was pending at that observation; its completed correction/acceptance follows.

Domain gate completed: fresh reviewer required an explicit unborn-HEAD boundary
policy; corrected report independently accepted at
86d2a4c0cc0eb8f427d597617246e85473a6cf55, merged via #34 at
5b1cf181fd3245a65337f1db28ee8f0fb95c200f. All nine corrected-HEAD CI jobs passed:
https://github.com/Hans-Einar/gh-tree/actions/runs/33997317388 . Report counts
remain 4 HIGH / 8 MEDIUM / 2 LOW. The report correction is resolved; its product
findings are still unimplemented. Issue #33 is closed.

Environment: Windows PowerShell; Git/GitHub CLI/Python available. Go bootstrap
located at C:/Users/hanse/.codex/tmp/ponsse-r9-go1.23.12/go/bin/go.exe; its automatic
toolchain selection obtained Go 1.25.0. Full local Windows `go test ./...` passed
in the Domain-review worktree at the corrected HEAD. Product source equals the
frozen v0.3.14 baseline. This is baseline evidence only.

Checkpoint scaffolding #35 / PR #37: accepted by a fresh independent reviewer
at 8b5514e5e69ed6de603ed5f34e6e530a6a8b76f3 and merged at
11cfef79009bb3956f0ba970470df14843e27e07. All nine jobs passed on both push and
PR CI. JSON/YAML/NDJSON and links/provenance verified. Historical wording nits
are clarified in this follow-up. The 91 finding IDs across the six accepted
reports are now explicitly indexed as open; related reports overlap, so this
count is not a claim of 91 distinct defects.

Finding-index checkpoint PR #38 accepted independently at
9f33ea31535612d6bd2d7d8d1af167db3d19585f and merged at
e0fa9c8743c0b68e936169b7997dd0c841e9dc98 with all push/PR CI green.

Release baseline read-only inventory: v0.3.14 publishes 12 assets, not merely the
five CI cross-build targets. Preserve darwin amd64/arm64; freebsd 386/amd64/arm64;
linux 386/amd64/arm/arm64; windows 386/amd64/arm64 (.exe). Exact names are in
CurrentIndex.yaml. Evidence: `gh release view v0.3.14 --repo Hans-Einar/gh-tree`
and https://github.com/Hans-Einar/gh-tree/releases/tag/v0.3.14 . Composition and
REFDES must include all targets in packaging parity and cross-build verification.
`gh extension list` reports installed gh tree v0.3.13; no upgrade performed.

Launch Discovery #36 / PR #39 accepted after fresh independent source/artifact
review at 859eab7af576b6b0eb689da6d7ed7c6d235d5853; merged at
3b8923e0fad29eec24a2d8fa1811dba24c2a626f. All nine CI jobs passed in run33998243250.
Master independently matched the four probe source hashes and reran all five
nonexecuting probes, reproducing reported baseline behavior. Report author also
ran 7 launch and 4 TUI tests. No report correction required. Its 2 HIGH / 9 MEDIUM /
1 LOW findings remain open product work; total indexed IDs now 103 (32H/58M/13L),
with overlap across reports. Next focused review: Persistence #40. No BC frozen.
