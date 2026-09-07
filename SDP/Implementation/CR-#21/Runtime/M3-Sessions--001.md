# M3 Sessions common-parent milestone — #71

Current checkpoint, 2026-09-07: startup-publication correction is clean and pushed
at technical `5a14f7daed121c5d8f357d81d45c8ee8db1a599d`; bounded independent
confirmation is pending. The reviewer resolved M71-H01/M01 and the original resize
case at prior technical `4495635`, retaining M71-M02 for the startup interaction
corrected below. Original rejected sources/reports/evidence remain unchanged.

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

## Independent-review corrections

Authority is Master's bounded correction dispatch under #71/#65/#21, recorded at
canonical `e9e2d0b3606a58a3b3dee3bcc02e6c0ee5e29ff3` / ledger 95. Correction base
was clean pushed `c436773ea9f6debb70918c18d48c0c6a7d89302e`. Only owned root Go
files/tests, README and this author report changed. No broker/native/assets,
helper binding test, public API/ports, frozen contract or Master/reviewer artifact
was edited or merged.

Coherent source commits were tested and pushed before proceeding:

- `700668b`: H01 retains the admitted predecessor pointer through Stop,
  replacement admission and canceled-result attribution after history eviction.
- `ac89811`: M01 gives Stop and Shutdown the same retained observer ownership.
  One coalesced external recovery intent survives an incomplete-error handoff;
  acquiring/active observers cannot overlap. A persistent failure does not poll.
- `7c74e43`: M02 reserves one session publication slot before each optional
  native control. Output cannot spend it. Unavoidable observed facts accumulate
  privately until a fresh publication/final; published snapshots are immutable.
  New input admission also refuses exhausted unreserved sequence space, while
  already accepted input/receipt effects remain owned and are never replayed.
- `4495635`: H01 also retains active restart keys and aggregate subjects outside
  cleaned-history eviction. Competing Start/Restart cannot steal an admitted key;
  Shutdown captures/waits that transition. At most 256 active restart subjects
  are retained, with Busy before additional transition admission. Existing
  64-live/256-final/256-history and byte limits are unchanged.

Root tests add nine deterministic regressions. They cover actual 256-session
cleaned/ACKed eviction, the no-eviction control, cancellation after replacement
admission, colliding lifecycle keys and aggregate capture; before/after observation
error with repeated/concurrent Stop/Shutdown and outstanding receipts; one retry
per new request on a permanently failed owner; native-effect refusal at Max-1,
an in-flight resize racing output and cleanup, supplemental delivery errors,
Max final geometry/exit/restart truth and retained exhausted residual repair.
Only numerical counters are positioned directly for overflow fixtures; the
effects and lifecycle use actual common code and controlled native owners.

Verification at clean technical `4495635` (Windows amd64, explicit pinned Go1.25.0):

| Check | Result |
|---|---|
| Root `go test -race ./internal/runtime -count=1 -timeout=60s` | PASS, all foundations and 20 engine/correction tests. |
| Root `go vet ./internal/runtime` | PASS. |
| Each affected correction group repeated with race instrumentation, count=10 | PASS at its coherent source checkpoint. |
| Original unchanged independent `TestReview*` probes via Go overlay, race count=10 | PASS, including all three original adversaries and original receipt/error/reentrant-native positive controls. This author replay is not independent acceptance. |
| `go run ./internal/composition/architecture` | PASS, all 12 target selections and 61 existing exact allowances. |
| `CGO_ENABLED=0 GOTOOLCHAIN=local GOWORK=off go build ./...` | PASS for all 12 accepted targets. |
| Source `git diff --check` / owned-path scope / clean pushed branch | PASS. |
| Source CI [34080252451](https://github.com/Hans-Einar/gh-tree/actions/runs/34080252451) | Queued when frozen; not claimed passed. The unbound development base still lacks the separately owned helper prerequisite. |

Python subprocess capture explicitly used UTF-8. No installations, global settings,
native user-process fixtures, forbidden cleanup or duplicate evidence archives.
The sole temporary reviewer overlay JSON points at the original committed probe.

## Startup publication follow-up

Master `73a8005f4a0fc366d359ff9f33705321978eb8cd` / ledger97 authorized the remaining
M71-M02 correction from clean review-only `15eed03fed36f844d6f7d50001a62760c5dddec0`.
The original startup overlay demonstrates that output refusal followed by Stop
spent the last nonfinal version, preventing acquired cwd publication before a
known Established result. That failed candidate remains preserved.

Technical `5a14f7d` makes startup publication an explicit reservation observed by
every hint producer. Acquisition consumes it atomically with publishing its full
native result, independently of whether the caller still waits. Early output,
concurrent Stop and cancellation cannot spend the slot; previously published
snapshots stay immutable. No panic handling or fabricated unestablished result
was added. Existing resolved lifecycle/control ownership is unchanged.

Two new tests cover six successful/error/partial/no-resource/canceled startup
cases at Max-2 with early output and concurrent Stop, plus startup at Max-3 handing
off to accepted input and terminal-unknown control receipts. They check admitted
identity, genuine Established/cwd plus errors, repeated Start, fresh versions,
unchanged old snapshots, receipt ownership and final truth without wrap.

At clean technical `5a14f7d`: targeted startup cases and both unchanged independent
confirmation probes PASS with race count=10; complete root race suite (22 engine/
correction tests plus foundations), root vet, architecture/all12 selections and
all12 ordinary cross-builds PASS. The original independent startup probe now
returns a valid Established result; the two-receipt positive remains passing.
This is author execution evidence, not independent acceptance or native proof.
Pinned Go1.25.0/UTF-8 tooling and ownership constraints remain unchanged.
Source CI [34081419821](https://github.com/Hans-Einar/gh-tree/actions/runs/34081419821)
is in progress at freeze; no CI completion is claimed. Separately reviewed native
source/assets were not adopted into this common-parent branch.

## Remaining work and exact next action

Existing independent reviewer confirms the bounded corrections against exact
technical `5a14f7d` and this report, then Master records the disposition. Common
engine confirmation cannot close #71/#65/Runtime, a Slice or an integration gate.

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
