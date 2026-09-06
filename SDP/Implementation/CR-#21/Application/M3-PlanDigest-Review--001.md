# M3 plan digest independent review — #68

Disposition: **ACCEPT**; no findings in the bounded contribution.
Role: fresh independent Reviewer under #68 / parent #21; Sprint-004-v04 / M3 prerequisite.
Repository: Hans-Einar/gh-tree. Branch: `codex/cr21-plan-digest`.
Exact technical SHA: `70d1719d7f0ff834b4d59245ea775ea7b2c2d66c`.
Compared base: `64750cb26a0b8e3088e82ced57f5a2bf99b3ac1b`.

Read root AGENTS/developmentInstructions, complete #68 and #21 body/comments,
current frozen Application--Git 1.1.0 G5/G6 and shared BoundaryTypes 1.0.0
validity/copy/identity authority, actual ports README, plans, closed variants,
approval/receipt implementation, existing and changed tests. The worker report
was inspected for scope/claims; its summary was not used as test evidence.

The actual five-file diff contains one ten-line product addition: the accessor
calls unchanged `identity`, returns the existing `[32]byte` digest by value, and
returns zero plus `invalid("plan")` on rejection. No hashing convention, public
API record, issuer signature, native registry, contract or other layer changed.
Exact type switches reject nil, all eight zero values, all eight typed nils and
foreign wrappers even when embedding a valid plan. The getter does not call the
interface's potentially foreign `Valid` method or infer authority from it.

Inspected tests and independent execution establish opaque-only consumption
without returning the minting PlanSpec/digest to the consumer; correct approval
binding; copy isolation; digest mismatch and foreign-plan rejection; allowed
choice and required confirmation checks. The changed continuation control uses
separate root/child digests, reads the child's own digest, rejects independent
child approval and direct root execution, and accepts the original root-bound
approval for its child. Existing issuer/operation/version/group/step/receipt
checks remain intact. Native admission and consumed-confirmation enforcement
remain separate responsibilities exactly as documented.

Independent verification on Windows/amd64, direct Go 1.25.0 (all PASS):

- `go test ./internal/application/... -count=1`.
- `go vet ./internal/application/...`.
- `go run ./internal/composition/architecture -target windows/amd64`.
- `go test -race -overlay <review-overlay> ./internal/application/... -count=1`.
- `git diff --check`; candidate worktree clean before and after tests.

The reviewer-only overlay adds `TestReviewOpaqueDigestControls`: producer-local
cryptographically random digest, opaque-only return, confirming/non-confirming
approval, exact retained choice/ConfirmationID, foreign coordinator rejection,
same plan metadata with a different digest rejected, all 32 single-byte digest
mutations rejected, unchanged getter value/existing approval after copy mutation,
and zero approval rejected after observation. It reuses only the existing public
semantic fixture constructor, not private fields or a returned digest oracle.
Temporary source and overlay remain at
`C:/Users/hanse/.codex/tmp/gh-tree-plan-digest-review-70d1719/`;
source SHA256 `91c5f879434deb10c7a9924c18c629c214932d23504a180fbdb851aff59d8096`.
No product files were edited for review and no redundant native archive created.

Limits/next action: this acceptance covers the mechanical portable accessor and
its regressions only. Source CI is separately watched/gated by Master; it is not
claimed passed here. No native Git, coordinator one-use confirmation, full Slice,
baseline finding, release or integration completion is claimed. Master may gate
serial integration on this exact accepted source and successful configured CI,
then supply the prerequisite to #61. Reviewer performs no merge or Issue closure.
This report-only commit intentionally uses `[skip ci]`; product SHA stays above.
