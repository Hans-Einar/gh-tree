# Composition architecture checks

Issue #57 implements the M1 prerequisite in accepted CR-21 REFDES/API,
MigrationMap, WindowsBroker and the seven FROZEN 1.0.0 boundary contracts.
This build tool is owned by Composition. Product packages cannot import it.
There is no generic shared process, filesystem or types layer.

From the repository root, with the module's Go 1.25 toolchain:

```sh
go test ./internal/composition/architecture -count=1
go run ./internal/composition/architecture
go run ./internal/composition/architecture -target windows/arm64
go run ./internal/composition/architecture -mode strict
go run ./internal/composition/architecture -targets
go run ./internal/composition/architecture -runtime-prerequisite
```

The default command checks all twelve release selections. CI invokes the same
checker separately for every matrix target. `-targets` compares the complete
name/pair inventory with the independent accepted Verification--001 table before
printing its JSON matrix; a missing/changed/duplicate target fails the job.
All selected Go metadata uses `CGO_ENABLED=0`, default architecture baselines,
`GOWORK=off` and `GOFLAGS=-mod=readonly`. No module file or product code is changed.
`go list -deps -export -json ./...` must succeed for the chosen target. Go's
target export data and go/types resolve the actual module-qualified package and
public type identities; no host-size or import-substring shortcut is used.

## Migration and package ownership

`MigrationMap.yaml` is JSON-compatible YAML and is read with the standard library.
Every remaining old Go file must match its exact path and baseline Git blob.
The checker reads path-specific text attributes and core.autocrlf. Explicit text
and ordinary automatic text profiles preserve legitimate CRLF checkout conversion;
`-text` and autocrlf=false retain the actual bytes. Automatic conversion with a
CRLF/binary index, filter/encoding/ident/eol attributes, any explicit historical
`crlf` attribute (including `-crlf`), and unknown profiles refuse
explicitly. Git clean filters are never executed. Source archives with no Git
metadata get raw-byte comparison, with no inferred Windows normalization.
The checker does not trust folder prefixes or a moved copy of the same blob. New files in
old folders fail; importing an unchanged legacy package from new code fails too.
Deleting an allowance's file does not authorize its replacement or retirement:
those decisions still require the Master-managed M5/M6/M7 gates.

The unchanged `cmd/gh-tree/main.go` has a distinct exact shared-entry allowance,
read from MigrationMap's retained shared-file record, while the CLI starts the
old stack. Once rewritten at M6, it gets the strict entry imports immediately.
`internal/version` always gets strict Composition-value policy. No unchanged
baseline exception carries into `-mode strict`; that mode also requires the final
named layer, leaf, host and private broker/helper package inventory. It is expected
to fail on the M1 tree. A custom-tag-only or unsupported-platform-only production
file cannot hide outside all twelve default selections.

The selected root may be an alias and is bound to its physical path. Selected
source paths must retain that physical containment and declared ownership; a child
alias cannot relabel an outside, another-layer or legacy source. Windows uses a
temporary, fully shared read-attributes handle and
[GetFinalPathNameByHandleW](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getfinalpathnamebyhandlew)
because Go1.25 EvalSymlinks does not resolve every junction. Other hosts resolve
symbolic links. Handles close before returning. This is inspection of a stable
checkout, not a retained capability or protection against concurrent tree changes.
Native Windows junction tests require no symlink privilege and exercise both
selected-root positives and outside/cross-layer/legacy child-alias refusals.

Private subpackages remain under the same layer policy. Cross-layer imports end
at published roots; the explicit coordinator/usecases and Runtime broker/asset
edges preserve their accepted private decomposition. Broker code cannot import
the registry, assets or generator. Private short-command helpers in Git/GitHub
remain independently owned; neither may call Runtime or the other adapter.

API/ports, Domain, State/viewmodel/View and version use a bounded pure standard
library list. `time` is limited to supplied values and deterministic operations;
clock/lifecycle/environment symbols, operational fmt calls, OS I/O, channels and
goroutine starts fail. Domain also rejects JSON tags. API may use only the pure
`json.Valid` predicate to implement the frozen Storage BC's OpaqueJSON validation;
the JSON naming/encoding schema stays Persistence-owned. Rendering libraries are
explicitly listed; automatic/global Lip Gloss renderer configuration is rejected.
New dependencies or policy changes need the applicable Issue/BC authority and
independent review, not an in-source suppression comment.

Export checks follow aliases, pointers, collections, generic arguments, all
boundary-reachable module named types, embedded fields, methods and inferred variables.
They distinguish declared API functions/interface methods from callback signatures
inside value graphs, rejecting nested callbacks, native/adapter types, channels
and any DTOs on public boundary surfaces. Generic `Optional[T any]` constraints
are permitted. Private implementation packages may exchange native resources with
their owner. Their wrapper types are recursively checked if a public owner surface
exposes them, so a private helper cannot hide native handles or nested callbacks.
Composition wiring and Runtime's private broker/asset/build tools have different
public implementation surfaces; the public product DTO rules are enforced at
their public roots/API/ports/Domain boundaries and all import edges remain checked.
Test-only standard-library ownership comes from the actual selected `go list std`
metadata, so trimpath-built tools do not depend on an embedded GOROOT pathname.

## Tests and proof limits

`_test.go` files use an explicit separate policy: standard library fixture I/O and
the listed test libraries are permitted; internal imports obey the layer graph.
External test packages have no blanket production exemption. Only Composition's
test harness may assemble the complete new stack. It still cannot import legacy
code or private build tools. Existing baseline tests require their exact blobs.

The JSON fixtures are written to isolated temporary modules, compiled through
actual target `go list`, and passed through the real checker. Negative tests assert
the intended policy diagnostic, so an unrelated compile failure does not pass.
They cover direct/import-prefix/cross-layer restrictions, Windows/FreeBSD selected
sources, public aliases and inferred types, legacy edits/renames/new callers,
separate external-test policy, and legitimate private/broker decompositions.

This checker is a mechanical guard, not proof of semantic ownership, immutable
copying, deterministic renderer internals, absence of hidden mutable global state,
or the meaning of arbitrary callbacks/integers. It cannot prove that rewritten
code is not a cosmetic old facade, or that a permitted library invocation performs
no implicit I/O through its internals. Independent source review is mandatory for
those properties, dependency additions and the full BC/Slice behavior. Cross-builds
and target type checking do not establish native Runtime/ABI or release parity.

## Runtime M3 hook

Until Runtime production inputs exist, `-runtime-prerequisite` prints `pending-m3`
and explicitly reports **conformance NOT RUN**. The corresponding real CI check
is skipped, never represented as a successful helper rebuild. The first Runtime
Go/image input makes these prerequisites mandatory:

- Go sources under `internal/runtime/broker`, `broker/cmd` and `cmd/helpergen`;
- nonempty `internal/runtime/brokerassets/broker-amd64.gz`, `broker-arm64.gz` and
  `manifest.json`.

With complete inputs it prints `ready`. CI then requires the Runtime-owned command
`go run ./internal/runtime/cmd/helpergen -check` on native Windows/amd64 Go1.25.0,
CGO disabled and pinned modules, followed by an ordinary product build without
generation. The verifier must exit nonzero on every source-closure, machine,
manifest, image or deterministic-compression mismatch, rebuild both native helper
targets independently with WindowsBroker--001's exact options, and compare actual
bytes. It must not rewrite the checkout, download a helper or compile at runtime.
CI rejects tracked or untracked changes to Runtime inputs after verification.
These are explicit inputs to the M3 worker; M1 implements no helper or fake verifier.

## Release staging

`release.yml` is manual compilation staging with read-only repository permissions.
It checks out a full specified SHA, verifies its identity, builds exactly twelve
accepted names, writes lengths/hashes/source/workflow/toolchain manifest and uploads
a temporary workflow artifact. It has no tag/push trigger, release action, release
creation, asset clobber or publishing permission. The manifest explicitly states
`compilation_only_M8_not_verified`. Native conformance, helper rebuild provenance,
full Slice gates, executable/version inspection and exact-SHA publication remain
the independently reviewed M8 contribution. Do not use publication to test staging.
