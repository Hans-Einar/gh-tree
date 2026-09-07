# M3 Unix integration correction review — #65 / #21

Disposition: **CHANGES_REQUIRED — one MEDIUM finding (M65-I01).** The bounded
read-side correction is sound; the adjacent acquisition-side disappearance is
still unhandled. No signaling, cleanup or cross-layer redesign is requested.

Reviewer: fresh `m3_runtime_linux_integration_review`, separate from the author.
Master dispatch62215da / ledger101. Exact clean/pushed reviewed source:
`99af6bf2217f91c1f3b870a1f7389bb71269c669`, branch
`codereview-21/layer-runtime`, worktree
`C:/Users/hanse/GIT/gh-tree-wt/runtime-implementation`.
Compared with `fd3abeac967a695bcdad15918a723ff8bab751df`: only
`census_linux.go`, new `census_linux_test.go` and the author report differ.
No product/test changes were made by this reviewer.

Authority: full #21/#65 including comments; repository governance; frozen
Application--Runtime1.0.0/BCFreeze; RTF-02 ownership, complete-census and no numeric
signal rules; V-RUN-01/02/04/07. Existing accepted Unix/shared, Windows, helper
and binding reviews remain applicable to unchanged source. This pass reopens
only the Linux census delta and its startup/cleanup effects.

## M65-I01 — acquisition ESRCH still rejects a disappeared proc record

`internal/runtime/broker/census_linux.go:31` opens each `/proc/PID/stat` and omits
only `os.IsNotExist(err)`. An ESRCH acquisition error is returned as an incomplete
census, although it can mean precisely that the enumerated task disappeared.
Supervisor startup calls this census before executable lookup, so ordinary
process turnover can still prevent a valid session start or replace its intended
typed NotFound result with the census I/O failure. This is an availability/error
classification defect; the current code does not falsely claim cleanup success.

Independent native evidence, as UID/GID65534: open an owned live child's proc
directory and stat descriptor; close its input and wait/reap that exact child
with exit0. A fresh pathname open returns ENOENT; `openat(retainedDir, "stat")`
returns ESRCH3. Wrapped as the actual Go open PathError, ESRCH is not recognized
by the current `os.IsNotExist` predicate. The separately retained stat read is
correctly omitted and closed by the reviewed correction. A second live owned
child remains visible in the full census and is subsequently joined by its owner.

The [Linux v6.18 source](https://github.com/torvalds/linux/blob/v6.18/fs/proc/base.c#L719)
explains the acquisition result: `proc_pid_permission` returns ESRCH when its
inode no longer resolves to a task. Absolute `/proc/PID/stat` traversal uses that
same permission check; its task can disappear after lookup. The retained-directory
control makes that native condition deterministic. It does **not** claim to have
interposed the production absolute-path syscall or identified the unavailable
native errno in original CI34082980191. That original attribution remains
plausible, not proven, and its failing source/log/assertion are unchanged.

Required bounded correction: recognize exact native disappearance at acquisition
as well as read, preserving all other open/read/close errors and all existing
census bounds, parser, membership and quiescence rules. Add a direct native
acquisition control plus unrelated-error controls for the actual predicate used
by census; retain the live positive and read/close negatives. No retry, numeric
signaling, permission-as-empty or cleanup relaxation is authorized by this finding.

## Accepted behavior and independent evidence

`readLinuxStat` owns the descriptor and closes it exactly once. ENOENT/ESRCH read
disappearance with a successful close yields no record. Close failure takes
precedence and preserves both errors; EACCES/EIO/EBADF, malformed/empty/oversized
records and valid bytes with failed close still refuse. The 8192-byte record and
65536-record limits, parser/identity fields, session membership, retained waiters,
helper acquisition and Quiescent/Release conditions are unchanged. The syscall
source also confirms `proc_single_show` returns ESRCH only when the inode's task
is unavailable before calling the stat producer.

Independent Windows/amd64 Go1.25.0 builder, Linux/amd64 CGO0 test binary with one
external reviewer-test overlay. Native execution: existing openSUSE-Leap-15.5,
Linux6.18.33.2-microsoft-standard-WSL2 amd64, UID/GID65534, owned ext4 directory
`/tmp/gh-tree-census-review-1_2m5lic`. Full broker suite plus reviewer native
control **PASS**, exit0. This includes the unchanged typed failed-start gate,
read/error controls, foreign/disappearing/acquired helper groups, root-before-
grandchildren, PTY job control, escape/EOF residuals and partial acquisition.
Twelve resource cycles close102 registered descriptors; fd count6→6 and
goroutines2→2. A subsequent exact executable-path inventory finds no live fixture.

Minimal reproducible review evidence:
`SDP/Verification/CR-#21/Evidence/M65-Unix-Integration-Review--001/`
contains the external `.go.txt` overlay, native log and commands/hash manifest.
The test intentionally demonstrates the remaining open predicate gap; its PASS
means the native evidence was reproduced, not that M65-I01 is resolved.
An initial inline shell harness failed before test execution during directory
selection; the explicit Python staging harness then ran successfully. Reproducible
binaries remain in the named Linux temp directory, `/broker.test` from that failed
staging attempt, and the Windows temp directory named in the manifest; no unique
product state exists there. No unrelated fixture was removed.

Independently retrieved [CI34084151911 attempt1](https://github.com/Hans-Einar/gh-tree/actions/runs/34084151911):
exact source99af6bf, **all20 jobs SUCCESS**, including Linux native/race,
macOS/FreeBSD/Windowsamd64/ARM64, helper source closure/reproducibility and all12
build/architecture targets. This does not override the native/source finding.
The Windows/shared/helper product closure is unchanged by the two Linux-selected
files; no generated-asset replacement is required for this delta.

Next permitted action: same bounded author corrects M65-I01, verifies and pushes
coherent source; this reviewer confirms only that edge and affected native
controls. Master then records the exact integrated-source CI gate. Parent
Sessions/production binding, full #65, canonical serial M3 integration, all
Slices/M8/product PR/release remain separate gates. No merge or acceptance of
those contributions is performed here.

## Bounded M65-I01 confirmation

**Final disposition: ACCEPT the bounded Linux census correction; M65-I01 is
resolved.** Exact clean/pushed source
`e39ff4c72a4ec3f175d36f3c436ed3aa2a28030c` was independently compared with the
original reviewbf8df249. Only the acquisition predicate, focused Linux tests and
author report changed. The original rejected source/report/evidence remain
preserved; accepted read/close, signaling, native cleanup and shared/Windows/helper
source are unchanged.

The actual census call now uses `linuxProcEntryGone`: only ENOENT/ESRCH, direct
or under the single os.Open PathError wrapper, count as disappearance. Nil,
permission/I/O/descriptor/not-directory and joined errors do not. No generic
substring, broad error-tree match or retry can hide an unrelated acquisition
failure. The native retained-directory control now calls that actual predicate;
fresh ENOENT and native ESRCH are recognized, with live/read/closed-owner facts
still required.

Independent targeted execution **PASS**, exit0, on the same Go1.25.0/Linux/amd64
profile as UID/GID65534 in owned ext4
`/tmp/gh-tree-census-confirm-q394y91h`: native acquisition/read disappearance,
all acquisition/error and read/close negatives, unchanged typed NotFound startup,
and the unchanged reviewer live-peer control. Both children join; exact executable
inventory is empty. `linux-confirmation.log` and `confirmation.json` in the
existing evidence directory retain commands/hashes. The historical reviewer
overlay deliberately still checks the original os.IsNotExist expression; its
printed "current open predicate" describes the original rejected source. The
corrected production predicate is independently exercised by the updated native
and classification tests. No full native suite or expensive matrix was repeated
locally; the accepted unaffected evidence above is reused.

[CI34084970413](https://github.com/Hans-Einar/gh-tree/actions/runs/34084970413)
names exact corrected sourcee39ff4c and is still running at confirmation. Master
must inspect its terminal result before accepting the combined integration gate.
This acceptance closes only M65-I01 and the bounded census correction; parent
binding, full Runtime, canonical serial integration and later program gates are
unchanged. No merge, product edit or broad review restart was performed.
