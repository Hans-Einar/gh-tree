# GitHub remote adapter

Implements `ports.RemoteFacts` and `ports.RemoteMutations` under frozen GH1..GH8.
Construct with `New(Config{Host: "github.com"})`; resolve a validated locator
first. Config snapshots authentication environment/profile and limits without
login, refresh, local Git, global configuration writes or a checkout dependency.
The legacy parent client is never imported. All concrete wire DTOs remain private.

Remote ID token encoding is lowercase `host/owner/name`, in `domain.Remote`.
`ParseLocator` removes one terminal `.git` transport suffix and rejects credentials,
query/fragment/ambiguous components. The provider profile accepts GitHub ASCII
login/repository component syntax; branch names preserve exact validated bytes.
Observed URLs and repository response identity must agree. Resolution registers
the verified association; other calls reject unregistered tokens. A rename/transfer
resolution explicitly returns a new association and diagnostic. SHA-1 is the current
provider capability; Domain's SHA-256 support is not silently asserted remotely.

Each list call reads one bounded REST page, limit1..100 (callers use100 by default).
PR ordering is created ascending; branch ordering is provider default. Opaque
continuations bind repository/filter/order/limit and ten-page refresh scope. The
registry holds at most256 ten-minute cursors and4096 repository associations.
History stores bounded identity/fact digests; identical duplicates are diagnosed
and deduplicated, conflicts survive with Unknown completeness. Every returned fact
retains its page's acquisition interval/version; separate pages are not a snapshot.
Malformed independent records preserve valid neighbors. Only exhaustive valid
response evidence yields Complete; More/Unknown never proves absence. Missing PR
forks never inherit base scope; missing body uses explicit field absence.

Creation reobserves exact published endpoints, checks qualified existing candidates,
and reobserves immediately before the sole explicit REST POST. Payload fields are
literal JSON on stdin: qualified head owner/ref plus head_repo, base, title/body,
draft and maintainer intent. No implicit push/fork or `gh pr create` occurs. A201
with validated identity is retained across transport errors/cancellation and a
separately bounded post-observation. Drift/refresh failure does not erase creation;
unproved responses retain indeterminate request/identity evidence. Existing facts
never establish causal ownership and no mutation is automatically replayed/reversed.

Native transport owns separate bounded stdout/stderr (16MiB/256KiB), a single
waiter and joined pipe copies. Read/mutation budgets are30s/120s, with separately
bounded2s forced join. Shorter constructor bounds are supported. Diagnostics never
include raw stderr, payloads, credentials or inherited environment. HTTP status,
request/rate/reset/retry facts are structured; private404 remains ambiguous.

Windows starts hidden/suspended, assigns nonbreakaway kill-on-close Job, validates
the opened initial thread's process identity, then resumes. Failure cannot resume
an unassigned root; active-process count0 and root/pipe joins establish cleanup.
Unix owns only the root and pipes. Its dedicated group is observed for residuals;
**it never signals a numeric PGID**. Root/process-group escape or remaining group
members are outside a stronger containment claim. Remaining members produce
CleanupKnown=false; caller cancellation never implies remote rollback. This is a
short noninteractive gh transport, not Runtime session supervision. A root/pipe
join deadline failure remains an explicit residual with a retained waiter.

Run `go test ./internal/github/adapter` and `go test -race ./internal/github/adapter`.
`TestNative*` uses real owned subprocesses; Windows controls include immediate child
spawn, invalid Job assignment, wrong opened-thread identity, accounting ABI and
actual child process exit. Unix's retained-child fixture has a bounded self-exit
lease and checks conservative residual reporting. Create tests use raw transport
fixtures, never live PRs. An explicit `GH_TREE_ADAPTER_LIVE_READ=owner/name` enables
the separately labeled read-only gh compatibility test. Full workflows, State/View,
serial integration and independent acceptance remain separate program gates.
