# DesignReview--001 — early bounded design reconciliation

State: CORRECTIONS REQUIRED; not final frozen-HEAD design acceptance.
Authority: #52 under #21; base 58f1cb9eda941db0941cbb8e04e6a0559a3ca896.
Artifacts were uncommitted drafts. Independent reviewer design_safety_review read
actual REFDES and feasibility records; migration_design_review independently read
baseline source and inventoried physical ownership. Master records their findings
here and owns the normative corrections. Product findings remain all open.

## Safety findings

| ID | Severity | Finding and required correction | Current disposition |
|---|---|---|---|
| DES52-H01 | HIGH | Native switch/read-tree/stash cleanup can overwrite an intervening edit and still produce expected clean post-state. External-writer quiescence plus pre/post checks cannot uphold no silent dirty/untracked loss. Native stash reference-transaction hooks are also a concrete write-between-capture-and-cleanup path. | Proposed correction incorporated; independent re-review pending. REFDES selects the superseding scratch preparation/retained capture/no-replace publisher, native index/ref guards and explicit partial recovery. Eighteen bounded NTFS followup cases support feasibility, including native Git2.48.1 symref protocol. V-GIT-06/07 and native platform/crash proof remain required. |
| DES52-H02 | HIGH | Darwin/FreeBSD census followed by numeric PID/PGID signaling has a reuse gap; retaining the SID leader does not pin every other group. This is ordinary process churn, independent of deliberate daemon escape. | Proposed correction incorporated; independent re-review pending. RTF-02 selects a retained private SID supervisor whose helpers acquire same-session group membership before signaling their own kernel-held group. Native Linux mechanism probe passed as UID65534; Master read actual source/log. Full protocol and native implementation proof required by V-RUN-04. |

The reviewer found no additional blocker in the explicit Unix SID ownership scope,
provided ordinary foreground/background job-control groups are covered and observed
escape yields residual failure. Deliberate setsid/daemon/external-service escape
remains an explicit portable containment limit. It is not evidence that descendants
inside the owned session may survive a successful Stop.

The narrow exact files-backend stash deletion/refusal protocol was considered
justified, with native lock-before-selection, survivor preservation, journal and
unsupported-backend refusal. Retained captured originals for tracked restore were
also considered appropriate; Python models do not prove platform filesystem APIs.

Reviewed input SHA-256 values, recorded by the independent early reviewer:

| Draft artifact | SHA-256 |
|---|---|
| REFDES--001.md | 459FFF66F42EF054B873767CE3BC15794A5F4DDFBADB15A786DDCF0686DF03DE |
| Feasibility/Runtime.md | 2C0897800455895E63A06464BE4F67192F3A7216D1F7E1CCC627C2111B743844 |
| Feasibility/Git-Safety.md | 8E4BF8C5444A88C698C50A49BE8423B30EB6F89FC816CD30CFC2AF11CB231031 |

These hashes identify the reviewed earlier drafts, not the current corrected
artifacts. The final complete design must receive a separate exact-commit review.

## Migration findings

The full source evidence and seven MEDIUM design inputs are in
Feasibility/Migration.md. Master has incorporated proposed corrections below;
their final adequacy remains subject to complete frozen-HEAD review.

| ID | Correction in normative design |
|---|---|
| MIG52-M01 | Final new GitHub package internal/github/adapter; unchanged legacy parent owned solely by MC until M7. |
| MIG52-M02 | Slices--001 M5 test-only full-stack Composition harness before cmd cutover, followed by host/CLI proof at M6. |
| MIG52-M03 | API/Slices explicitly preserve configured targets, StageAllAndCommit, unstage, historical branch creation, latest-stash identity resolution, launch overrides/first default, Ctrl+Enter activation and shell policy. |
| MIG52-M04 | MigrationMap inventories every old test file/function and pending replacement evidence; M7 cannot retire meaningful assertions without new proof or reviewed unsafe/obsolete disposition. |
| MIG52-M05 | Index-to-worktree diff retains actual behavior with truthful label; configured non-fast-forward branch deployment refuses with detached/new-branch alternatives. |
| MIG52-M06 | Machine-readable inventory names all 70 baseline paths, blob identities, sole edit owner and exact temporary legacy allowance; final allowance empty. |
| MIG52-M07 | M1 first Composition contribution enables canonical branch CI/twelve builds and disables unsafe bootstrap publication before product workers rely on it. |

Master structural checks: MigrationMap has 70 unique baseline paths, 38 production
Go files, 24 test files, eight shared files, and 90 baseline Test/Fuzz/Example
functions. Slices contains all 13 selected capabilities. FindingDisposition maps
all 143 canonical source findings to planned design clauses, Slice IDs and valid
verification IDs; none is marked resolved. These are document consistency checks,
not implementation or release verification.

Master independently read all of the RTF-02 probe source and recorded log, matched
probe SHA-256 0451B733D4D0429D73355BBEEC7F779B37A7F32B96ECBDA8553DDF65E5603503
and final Runtime appendix hash
7D3A6E51F45DFCBF7BF951FF18123B3A7CF82EF5F89ADAAA3B0C9C3789ADFC22,
then reran the native Linux executable through WSL as nobody. Abort, ordinary
groups, early-root exit, foreign/missing/acquired group, nonce/cancel and real PTY
cases again printed PASS with zero live session remnants and the unrelated fixture
alive until its owner's cleanup. This reproduces the bounded mechanism evidence;
it does not prove the future complete framed protocol or macOS/FreeBSD runtime.
