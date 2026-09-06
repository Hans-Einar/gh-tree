# M3 GitHub independent review — Issue #62

Disposition: **CHANGES_REQUIRED**, one P2 finding, no acceptance or integration.
Reviewer: fresh `m3_github_review`, separate from the author.
Reviewed technical SHA: `926b2b70900fa5f0c2277b086fda0d9b6aad1194`.
Branch: `codereview-21/layer-github`; base `412f33e477cec03cb6eafe7b846c9bcdd02c0a25`.
Worktree: `C:/Users/hanse/GIT/gh-tree-wt/github-implementation`, clean before review.

Authority: AGENTS/developmentInstructions, full #21/#62 including explicit resumed
policy, accepted GitHub #43 review/design, Application--GitHub GH1..GH8 and shared
BoundaryTypes FROZEN1.0.0, accepted API/ports, and canonical M3 interoperability
notes. The parent Git/status-cause proposal is outside this source and review.
Inspected all 17 candidate files (2,650 added lines), including actual source,
tests, native implementations, README and worker report. Legacy/shared product,
workflow, module and frozen-contract files are unchanged. Author summaries were
navigation only. This review does not close a vertical Slice or baseline finding.

## M3GH-R01 — P2: malformed UTF-8 record drops independent valid siblings

`internal/github/adapter/parsing.go:73`, `decodeList`, checks `utf8.Valid(raw)`
for the entire page before separating its `json.RawMessage` records. An invalid
byte inside one branch name or PR title therefore rejects every record, including
independently well-formed neighboring objects. Both public list methods return
zero facts and Unknown. The invalid scalar must remain rejected, but GH4/#62
requires keeping independently valid records with diagnostics and Unknown.

Independent controlled inputs, HTTP200 with no Link header:

- Branches: `[bad, good]`, where bad is
  `{"name":"bad<FF>","commit":{"sha":"1111111111111111111111111111111111111111"}}`
  and good is
  `{"name":"good","commit":{"sha":"2222222222222222222222222222222222222222"}}`.
  `<FF>` denotes the single raw byte `0xff`, not literal text or a JSON escape.
- PRs: use existing `prJSON("base", "fork", oidB)` for both objects, replacing
  only the first object's `"title":"literal title"` with `"title":"bad<FF>"`.
  Resolve `base/project` and list with FilterAll, limit100, InitialPage.

Expected: one unchanged valid sibling, Unknown completeness, malformed-record
diagnostic. Actual: zero siblings, Unknown, `remote protocol: invalid UTF-8`.
The framing remains separable into raw objects; this is not permission to recover
arbitrary broken JSON syntax or accept replacement-decoded identity bytes.

Correct by separating structurally framed records without globally rejecting
record-local UTF-8 failures, then retaining strict byte/shape/identity validation
on each record. Add branch and PR regressions with valid neighbors; preserve
whole-envelope rejection for structurally unparseable input. Re-review this
bounded change and run relevant source CI before any integration.

## Independent verification and limits

Go1.25.0, direct installed toolchain, native Windows amd64:
`go test ./internal/github/adapter -count=1 -v` **PASS**, including real separate
streams, bounded overflow, literal stdin, timeout/canceled admission, descendant
pipe joins, repeated resources, failed Job assignment, opened-thread owner
validation and accounting ABI. Opt-in live read was skipped; no live mutation.

Small independent tests were injected with Go `-overlay`, leaving the candidate
unchanged. `go test -overlay <overlay.json> ./internal/github/adapter -run
TestIndependent -count=1 -v` exits1 solely for the two malformed-UTF8 cases above.
Passing controls: valid Japanese branch bytes with independently rejected wrong
type/unpaired surrogate; second-preflight base OID drift sends no POST; postcheck
head branch identity drift with equal OID retains CreatedWithDrift, AppliedVerified
and the original requested endpoint. The existing nine creation scenarios cover
all six outcomes, response loss/partial identity, refresh failure and cancellation.

Reusable local fixture: `C:/Users/hanse/.codex/tmp/gh-tree-m3-github-review/independent_test.go`;
SHA256 `B2C6B429028B3F89A4F8C1C16A9B039961B4C849CC494E22293809B5EFF02555`.
Sibling `overlay.json` maps the nonexistent candidate
`internal/github/adapter/review_independent_test.go` to that fixture.
`independent.log` SHA256
`1793D78CA6FB27566C6D8AA6FC52FB86D2AF0E1EE6EAF8CF8438ACF2C4DC0A9B`.
Inputs and assertions above preserve reproduction without a full source archive.

Native Windows386/WOW64: compiled the unchanged package with GOARCH=386,
CGO_ENABLED=0 and executed the test binary with `-test.run=^TestNative -test.v`;
all six Native entries **PASS**. This exercises the 48-byte accounting layout,
Job-before-resume, immediate child termination and failure controls. The initial
PowerShell unquoted option invocation exited2 before tests; quoting complete
arguments corrected the invocation. Logs remain beside the independent fixture.

Native audit: Windows opens the retained suspended root, assigns kill-on-close
nonbreakaway Job before Resume, validates opened thread owner, kills on failed
ownership, owns one Wait, and combines joined readers/root reap with actual Job0.
Unix only kills the owned root; numeric PGID is used for kill(0) observation and
never a termination signal. Remaining ordinary group members report cleanup
uncertainty. Group escape is explicitly outside that profile. No remote rollback
or Runtime containment claim is accepted from command termination.

Independently queried [CI34054297560](https://github.com/Hans-Einar/gh-tree/actions/runs/34054297560):
completed SUCCESS at the exact reviewed technical SHA, 18 SUCCESS jobs covering
Windows amd64/ARM64, Linux/macOS native tests, Linux race and twelve selected
architecture/build targets. The sole Runtime M3 helper job is explicitly skipped
before Runtime exists. CI and its actual native tests supplement local evidence;
no local Unix execution or final workflow/release proof is claimed. No redundant
full matrix was rerun for unchanged code.

Next permitted action: author corrects M3GH-R01 in the owned adapter/tests and
updates its report, freezes/pushes corrected technical SHA; independent bounded
re-review checks sibling retention plus strict rejection/positive controls.
Master holds GitHub for Git-first serial integration and exact integrated CI.

## Bounded re-review — M3GH-R01 resolved

Current technical disposition: **ACCEPT** at corrected frozen/pushed SHA
`738261756b86930a5f2d7c8374f5eaf3deb56bf0`; no blocking review findings remain.
The initial rejection and original candidate above remain historical evidence.
The worktree was clean at this SHA. Actual correction changes only decodeList's
global UTF-8 admission, adds `record_utf8_test.go`, and updates the worker report.
Native supervision, transport, creation and remote fact methods are unchanged.

Independently inspected the complete correction: RawMessage preserves each
record's bytes; unchanged strictJSON rejects malformed UTF-8 before typed scalar
decoding. Both valid siblings now survive unchanged with Unknown and a record
diagnostic. Framing still refuses broken syntax, trailing values, wrong root,
null and invalid bytes outside a scalar; a valid empty array remains accepted.
No replacement-decoded identity is admitted.

Native Windows amd64, Go1.25.0, independently executed with exit0:

```text
go test -overlay C:/Users/hanse/.codex/tmp/gh-tree-m3-github-review/overlay.json ./internal/github/adapter -run TestIndependent -count=1 -v
go test ./internal/github/adapter -run 'TestMalformed(Branch|PullRequest)UTF8|TestListFraming|TestRepositoryScopeAndStrictParsing|TestBranchMalformedRecord' -count=1 -v
```

All four original independent groups (including both previously failing cases)
and five targeted parser/framing/regression tests PASS. git diff --check passes.
Prior exact-source native evidence is reused for unchanged native code. Corrected
source CI is a separate Master gate and is not claimed passed by this record.
No integration, whole Slice closure or release acceptance occurs here. Next:
Master checks corrected source CI, holds GitHub until accepted Git integration,
then performs the required serial integration and exact integrated-SHA gate.
