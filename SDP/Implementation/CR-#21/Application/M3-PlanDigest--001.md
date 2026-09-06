# M3 plan digest accessor — #68

Role: bounded Application ports Worker under #21/#61/#68.
Branch: `codex/cr21-plan-digest`; base: `64750cb26a0b8e3088e82ced57f5a2bf99b3ac1b`.
Source milestone: the commit containing this report (resolve with `git log -1 --format=%H -- internal/application/ports/plan_digest_test.go`).

Authority read: root AGENTS/developmentInstructions; #68 and current #21/#61
authority/comments; frozen Application--Git 1.1.0 G5/G6 (unchanged from 1.0.0),
shared BoundaryTypes 1.0.0 B1/B2; actual ports README, plan identity, variants,
accessors, ApprovalIssuer and existing tests. No contract was changed.

`PlanSummaryDigest(PreparedGitPlan) ([32]byte, error)` now returns the existing
issued digest through `identity`, using the same invalid-plan error as other
accessors. The array is returned by value. It does not compute a hash, issue
approval, change issuer signatures or introduce native state.

Evidence on Windows/amd64 with Go 1.25.0:

- `go test ./internal/application/... -count=1`: PASS.
- `go test -race ./internal/application/... -count=1`: PASS.
- `go vet ./internal/application/...`: PASS.
- `go run ./internal/composition/architecture -target windows/amd64`: PASS.
- `git diff --check`: PASS.

`TestPlanSummaryDigestOpaqueConsumer` receives only a plan from its fixture;
the fixture returns neither PlanSpec nor digest. Successful ApprovalIssuer
issuance verifies the retrieved bytes match the exact original digest. Controls
reject changed digest, different plan, canceled/disallowed/invalid choices and
missing/zero confirmation; changing the returned array leaves the plan intact.
`TestPlanSummaryDigestRejectsInvalidPlans` checks nil, all eight zero plans,
all eight typed nil pointers and foreign embedding (including a valid embedded
plan), requiring zero return plus the same error as PlanSummary.
The existing continuation test now obtains root approval through the accessor,
gives the child a distinct digest, verifies that exact child digest is returned,
rejects independent child approval, and retains root-not-executable and valid
root-bound child approval controls. Existing issuer/operation/version/receipt
tests run unchanged apart from these bounded continuation additions.

Scope is ports source/tests/README plus this report. No other layer, API DTO,
frozen contract, module/workflow or Master metadata changed. Native Git registry,
replay/release enforcement and coordinator intent/one-use confirmation remain
their existing owners' later implementation gates; no native execution, full
Slice completion or baseline finding closure is claimed.

Next permitted action: fresh independent reviewer inspects the pushed exact
source SHA and tests, then Master gates integration on accepted review and
exact-head configured CI. CI is requested by the source push, not claimed PASS
in this report. Worker performs no merge or Issue closure.
