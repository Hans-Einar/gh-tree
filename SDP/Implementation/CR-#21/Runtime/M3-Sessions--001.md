# M3 Sessions common-parent milestone — #71

Author checkpoint, 2026-09-07. Role: fresh Runtime Sessions worker. Authority:
#71 and its first-milestone dispatch comment under #65/#21;
Sprint-004-v04 / I-03 / M3. This is a partial adapter contribution awaiting separate
independent review, not Runtime/Slice/integration/release acceptance.

Branch `codex/cr21-runtime-sessions`, worktree
`C:/Users/hanse/GIT/gh-tree-wt/runtime-sessions`.
Verified clean pushed development base `c3f23b89610a8695125fcd881711a929c7828374`.
Common foundations remain from `6be865b`; the development base's Unix source is
not accepted native proof and carries the separately assigned M65-U01/U02 fixes.

Source checkpoints, all pushed before proceeding:

- `c3b0bf9`: common engine, private native ownership seam, exact twelve-port
  conformance and seven initial controlled-owner tests.
- `1fac62d`: final numerical reservation, retained canceled replacement identity,
  shutdown restart producers and three additional adverse cases.
- `fff5ee4a30f3b2d67997fcf74fb864a1174fead0`: final technical candidate; preserve
  known native cleanup with an accompanying error, correct project display
  locator and parent cleanup residual stages; eleven engine tests total.

## Implemented bounded behavior

Only root Runtime Go files/tests and the authorized local documentation changed.
`sessions` implements the exact accepted `ports.Sessions` interface with no
exported production constructor, platform fallback or placeholder native owner.
The unexported starter receives the copied immutable invocation/environment and
a bounded output callback. It returns a retained owner for every native partial
acquisition; nil means provably no acquired native resource. Actual platform
clients are deliberately not imported until Master adopts their reviewed source.

The common registry remains the only identity and membership authority. It
reserves 64 live / 256 final-event capacity, retains 256 cleaned records and
preserves never-wrapped IDs, session versions, global events and raw offsets.
Memory-only registry changes follow registry -> session -> event lock order;
native startup, delivery, receipt waits and Stop calls occur outside those locks.

Start copies constructor environment and policy into private restart state,
deduplicates retained lifecycle keys and retains failed admitted identity/finals.
Natural root exit initiates stopping without completing cleanup. Output captures
raw late bytes until the native callback join. Parent input/control producers
must also finish before publishing the one reliable final. The 64KiB copied
input reservation spans in-flight native work and 65484-byte frame splits.
Canceled receipt callers cannot abandon eventual delivery observation or replay
accepted bytes. Native partial/unknown delivery remains a bounded diagnostic.

Stop coalesces native intent, including Starting sessions. CleanupFailed retains
capacity and can receive later repaired facts. Restart serializes one transition
per retained old ID, rejects changed-key input and competing replacements, and
does not acquire the replacement before complete old cleanup. Its canceled
result retains any already admitted replacement ID. Shutdown closes admission
atomically, includes pending startups/restarts and parent receipt producers, and
reports final ACKs as EventTransfer residuals. Its deadline does not manufacture
cleanup, discard resources or terminate receipt ownership.

## Verification at the final technical source

Native host: Windows amd64, Go1.25.0 at the explicitly supplied module-toolchain
path; Python3.14 used only to orchestrate read-only cross-build environment
selection. No tool installs, global environment changes, module/workflow edits,
foreign process fixtures or production native startup occurred.

| Check | Result / scope |
|---|---|
| `go test -race ./internal/runtime -count=1 -timeout=60s` | PASS, includes all existing foundations and eleven deterministic engine tests. |
| `go vet ./internal/runtime` | PASS. |
| `go run ./internal/composition/architecture` | PASS, all twelve target selections, staged policy and existing exact legacy allowances. |
| `CGO_ENABLED=0 GOTOOLCHAIN=local GOWORK=off go build ./...` | PASS independently for all twelve accepted GOOS/GOARCH pairs. Compilation is separate from native runtime proof. |
| `git diff --check` | PASS. |
| `go run ./internal/composition/architecture -runtime-prerequisite` | Expected FAIL: this unbound base lacks `internal/runtime/broker/cmd` Go source. CI/native helper prerequisites remain incomplete; no workflow was changed. |

Meaningful barriers exercised: cancellation before admission versus during
startup versus after establishment; root exit before trailing NUL/invalid-UTF8
bytes; native cleanup before an outstanding parent receipt; 65536-byte queue
ownership and exact 65484+52 split; partial delivery without replay; canceled
resize with eventual geometry applied before final/restart; duplicate/changed
restart keys; cancellation after replacement admission; Shutdown racing blocked
startup plus concurrent Stop; CleanupFailed then repaired cleanup; final sequence
reservation at MaxUint64; and complete native facts accompanied by an error.
Existing registry/event tests retain the broader capacity/ACK/replay/history and
raw-ring boundary evidence. Controlled owners are common-engine evidence only.

## Remaining work and exact next action

Master supplies reviewed Unix fixes, reviewed Windows client and matching
verified helper assets. Read-only Windows seam source was inspected at accepted
technical `bd78deaf` / review `9b06f8bd`; it was not copied into this branch.
The real binding must normalize closed failure causes/stages and stable native
receipts, preserve private shell/main-ancestry resolution and environment rules,
and provide actual safe startup/display/cwd evidence. `native.go` states the
parent-side ownership contract; native receipt observation does not replace the
clients' own resource joins.

Then implement the private platform bridges and authorized production constructor,
exercise all twelve methods against real owned native processes/PTYs/ConPTYs,
and satisfy full V-RUN/native ABI/FreeBSD/ARM64/emulation/helper/source-binding
gates. Fresh independent parent review and Master integration are still required.
The separate Git review hold remains unchanged. No native acceptance, complete
adapter, whole Slice, canonical merge or release is claimed here.
