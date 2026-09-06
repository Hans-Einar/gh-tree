# M3 status cause API implementation

Worker handoff under #67 / #61 / #21, Sprint-004-v04 / I-03 / M3.
Disposition: implemented and locally verified; independent source review and
source/integration CI acceptance remain required. No Slice/finding completion.

Branch: `codex/cr21-status-api`.
Worktree: `C:/Users/hanse/GIT/gh-tree-wt/status-api`.
Exact starting HEAD: `10f687ea7df4cde1746b50ef4b5537ecdeb13c39`.
The coherent implementation milestone is the commit adding this report; its
exact SHA is supplied in the worker handoff for the separate reviewer to freeze.

Authority read: AGENTS.md, full developmentInstructions.md, Issues #21/#67 and
comments, frozen Application--Git 1.1.0 G3/G4 with ownership/observation context,
BoundaryTypes--001, M3-Adapters--001 BC-CHANGE-67 decision, accepted adjacent
M3-Status-BC-Review--001, and the API README/source. Scope is API and this report
only; no ports, Domain, viewmodel, adapter, legacy, module, workflow, frozen
contract or shared traceability changes.

## Implemented behavior

`ChangeFactData.Cause` is required. The closed enum uses IndexChangeCause,
WorktreeChangeCause, UntrackedChangeCause and ConflictChangeCause because the
short names already belong to other semantic families. Constructors validate
cause/kind, required/exclusive/distinct OldPath and unique allowed index stages.
Conflict retains every nonempty subset of stages 1..3; ordinary changes permit
stage 0 or absence, and Untracked permits no current index entry.

StatusFacts admits at most one row per exact (Path,Cause), excludes other causes
at the exact conflict path, and compares same-path current facts. The comparison
retains every FileState field, stage-to-object/mode identity and unordered semantic
flag sets. Existing IndexEntryFact admits repeated flags; repetition does not
invent an additional semantic fact. Inputs and returned entry/row/flag ordering
are preserved rather than rewritten. There was no existing FileState or index
entry semantic equality helper to reuse in the starting API tree.

Per-cause Kind/OldPath remain independent. Staged rename plus worktree modification
or deletion, subsequent B-to-C rename, different copy sources, and staged deletion
plus present untracked replacement are admitted. No destination index entry or
absent filesystem state is invented from Kind. API uses no native I/O, reflection,
cause inference or opaque SourceVersion decoding.

## Verification evidence

Local environment: Windows amd64, direct Go 1.25.0 toolchain at
`C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin`.

| Check | Result |
|---|---|
| `go test ./internal/application/api -run 'Test(Status\|ChangeCause)' -count=1` | PASS initial targeted cases |
| `go test ./... -count=1` | PASS all repository packages; final added stage-order equality test subsequently included in the race run below |
| `go test -race ./internal/application/... -count=1` | PASS, including final equality test |
| `go vet ./...` and final `go vet ./internal/application/...` | PASS |
| `go run ./internal/composition/architecture` | PASS all 12 target selections, 61 explicit existing legacy/shared-entry allowances |
| `go build -trimpath -o <temporary-output>/gh-tree-status-api.exe ./cmd/gh-tree` | PASS Windows amd64 build outside worktree |
| gofmt and `git diff --check` | PASS |

`status_causes_test.go` exercises staged-only/unstaged-only/both with identical
current index/filesystem values; independent rename/modification/deletion/copy
and chain identity; literal whitespace/newline/non-UTF8 path bytes; all seven
conflict subsets with both present and absent FileState; unknown/zero/incompatible
causes; source-path and stage negatives; duplicates; exact-path-only conflict
exclusion; and contradictory object/content/parent/mode/kind/link/index facts in
both row orders. Valid flag reordering/repetition does not cause rejection.
Nested admission/getter/Clone controls mutate outer changes, entries and flags.
Unborn status preserves absent Revision, and Complete/Partial/Unknown result
envelopes retain rows, emptiness, diagnostic copies, source versions and worktree
binding. Existing API tests had no actual ChangeFact constructor fixtures to
migrate; their existing clone-shape and status scope tests still pass.

`status_equality_test.go` additionally distinguishes stage-to-object association
from unordered OID sets: reordering all three entries/flags passes, swapped side
objects or omitted/reassigned stages fail. This is a focused helper control,
complementing constructor-facing tests rather than claiming native observation.

## Limits and next permitted action

Push this coherent source milestone and freeze its exact SHA for a fresh separate
implementation reviewer. Master watches applicable source CI and integrates only
after independent acceptance and required gates. Native status acquisition,
intent-to-add interpretation, SHA-1/SHA-256 real repositories, no-index-write and
source-drift proof remain #61. Lossless public/presentation projection remains
M4/M5. This worker does not merge, close #67, or mark any Slice complete.

No failed local verification or unresolved implementation finding is known.
Cross-platform runtime/race and all twelve release cross-builds are source CI
gates, not claimed by the local architecture check. Runtime helper verification
is still the existing explicit pending-M3 gate, unrelated to this correction.
