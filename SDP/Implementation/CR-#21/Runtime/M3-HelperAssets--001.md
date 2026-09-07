# M3 Runtime helper assets — Issue #70

Correction checkpoint (2026-09-07): independent review `7fdfc9b` requires
H70-H01/M02/M03; the original candidate below is rejected, not accepted.
M03 now resolves Windows selected roots through kernel final paths, accepting
an owned junction to the canonical GOROOT while rejecting a child junction
outside the selected root. Both that positive/negative native control and
`TestActualSelectedDependencyClosure` pass on native Go1.25.0 amd64. Regeneration
retains the exact ab608 broker images and updates recipe provenance to 902 inputs,
digest `b22395a0763d7a41f1f4681d1f171aa0090af2e63a1ae13afd8c6214d6db5fce`.
M02 now prepares pinned modules with Go's checksum mechanism inside the existing
command using temporary byte-identical go.mod/go.sum copies; attempted pin
rewrites refuse. Fresh, selected-only, populated/offline and inconsistent-pin
controls pass with an owned file proxy. Actual generation and the exact `go run
./internal/runtime/cmd/helpergen -check` both passed with distinct empty owned
GOMODCACHE directories. An external Python observer confirmed all 45 Runtime
files' bytes, lengths, mtimes and file set unchanged in check mode. The current
903-input digest is `5c8f02ba57b335162667bf656d043ebbe83ec562243230e912e2e413162dd292`.
Exact M02/M03 source `a9559d99` passes all 20 CI jobs in run 34070628791,
including the unchanged canonical helper checker and native Windows tests.

H01 now builds from captured repository, selected module and toolchain buffers
in isolated read-only input trees, copied Go/compiler/linker/assembler executables,
fresh work/build/module caches and offline subprocesses. Captured selected module
bytes are independently bound to root h1 pins by a full module-directory checksum;
cache ziphash and earlier verification are insufficient. Recursive assembly
includes bind actual nested headers (including Runtime ABI headers), rejecting
absolute/traversing include operands. Auto-PGO is off and linking is internal.
The real regression replaces an owned selected module, standard source,
compile.exe and go.exe during BOTH clean builds, restores all before recapture,
and proves identical original provenance plus images containing only recorded
module bytes for amd64/ARM64. Corrupt captured buffers, extra unrecorded inputs,
and inconsistent selected-module content refuse. Full generator/assets/architecture
package tests and generator/assets vet pass. Actual regeneration passes with
913 inputs, digest `1573f6fef6e93eb8c9608e358505a771bd671f7aad7eef47330828d2c4cf2a99`;
both broker images remain the exact ab608 bytes listed below.

This is a tested H01 correction checkpoint, pending final snapshot ownership
hardening/controls, exact fresh-cache/root-junction no-rewrite checker and bounded
independent re-review. It is not helper, Runtime or Slice acceptance. The historical
initial-candidate description below is superseded by these correction details.

Disposition: bounded candidate frozen for independent review; final Windows
source adoption/regeneration and integrated/native Runtime acceptance remain open.
Product candidate: `b72ad94c60eb38d7c4ebe70fafd10b21a3289037`.
Worker branch: `codex/cr21-runtime-helper-assets`, dedicated
`C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets`. Broker source base:
`ab608327e63727f66ffb1aa7b3200c2865307cf5`. No other worker changes are integrated.

Authority: full #21/#65/#69/#70, root instructions, frozen Application--Runtime
and BoundaryTypes 1.0.0, accepted WindowsBroker/CwdAcquisition/Runtime feasibility
and Verification--001. Sprint-004-v04 / I-03 / M3. Master owns completion gates.
The #70 parser-policy exception is recorded at canonical
`9aa504640c04b6f67778c19a47fae8e046d023ad`; separate prerequisite commit
`4b31ac4c34df370b852a1626107c102ff300c6dd` changes only the three authorized
Composition policy/fixture files. Assets may use in-memory JSON/PE parsers;
typed `pe.Open` references, native I/O, printing and goroutines remain forbidden.

## Implemented contract

`go run ./internal/runtime/cmd/helpergen -check` admits native Windows amd64
Go1.25.0 only, including an actual IsWow64Process2 check. It selects the pinned
Go executable from its compiler GOROOT and fixes GOENV/GOWORK/GOFLAGS/CGO,
target baselines, experiments/FIPS and external-cache settings. Module lookups
and toolchain downloads are disabled in its subprocesses. Dependencies must
already be available as verified pinned source modules, as after ordinary Go
dependency preparation; unused module archives are not required.

Both helper target closures come from actual `go list -deps -json`. All selected
Go/header/assembly/object/embed inputs are read; C/cgo/other compiler inputs and
unattributed/replaced modules refuse. Root go.mod/go.sum and selected module
versions/h1 sums/go.mod files are included. The source record contains sorted
normalized repository text hashes, raw embedded/binary/dependency hashes, selected
standard-library sources, actual Go/compiler/linker/assembler binaries and their
support headers. Recipe source is separately identified by a `recipe/` prefix.
Protocol is parsed from the actual broker constant. No generated input, registry
or asset package is in the helper dependency closure; no commit SHA is embedded.

Each invocation builds both targets twice in separate temporary normalized source
trees and fresh build caches, using CGO0, explicit GOAMD64=v1/GOARM64=v8.0,
trimpath, buildvcs=false and empty build ID. PE32+ executable machines and the
two independently built byte arrays must match. Source/provenance is recaptured
afterward to reject concurrent edits. Deterministic gzip has best compression,
zero timestamp, OS255 and no name/comment. Check mode compares all three exact
output byte arrays and never calls the checkout-writing branch. Temporary build
directories are owned and removed by the generator.

Pure `brokerassets.Load(arch)` returns owned bytes, SHA256, machine and protocol.
It verifies canonical unique JSON, schema/provenance/target consistency, bounded
lengths, gzip/header/CRC/single-member integrity, image hashes, PE machine/type and
exact deterministic compression. Windows386 embeds amd64+ARM64; Windowsamd64
embeds ARM64; other products embed neither helper. Unsupported/unavailable
selection refuses. A local .gitattributes fixes manifest LF and binary gzip bytes.
There are exactly three generated files and no extra published asset.

The Runtime and Windows authors agreed this API directly. Runtime parent decodes
the hash and passes plain values to `broker.ExtractWindowsImage`; extraction,
ACL/native identity/path guards, process creation and cleanup are the Windows
worker's responsibility. Broker imports no assets. Existing host/private-mode
composition and Sessions integration are untouched.

## Evidence at this source checkpoint

Local native toolchain: Go1.25.0 Windows/amd64, CGO disabled for helper generation.
The initial complete module-graph metadata lookup failed because unrelated unused
module metadata was absent offline; the implementation now selects only actual
helper dependencies while retaining exact root module pins. This earlier failed
attempt produced no accepted assets. No prerequisite/CI gate was weakened.

- Actual generation: two independent clean builds per architecture matched.
  Final source-closure digest:
  `e3a5f8b97f9b5e02fc389097429c98226146d202ba49fe398a843fd6cf06045c`
  (900 inputs). The manifest records individual source/module/options/toolchain
  and target hashes. The final added assembler-support-header provenance changes
  the manifest, not either executable image.
- amd64: PE machine0x8664, 3,493,376 image bytes,
  SHA256 `9188bd5063040a2a39b7d5d550b06cd22bebcad73c9406a14f53314b54b44c5b`;
  2,038,259 gzip bytes,
  SHA256 `437f614f3ae04390c47f547fd0a7f60608870fb1f63cb136379f0a1b725f0183`.
- ARM64: PE machine0xAA64, 3,308,544 image bytes,
  SHA256 `2676e8efdcd53f1a87de64790d22ec2fa2f190419005eba25213978e29696e5d`;
  1,902,326 gzip bytes,
  SHA256 `c9952ce13f6f4846c7445226d2bfc91a2826f135623ce0fac965452b2d90c8f4`.
- Asset/generator tests and vet passed on native amd64. Real go-list fixture
  detects changed API source, added dependency and binary embedded data; CRLF
  source is equivalent, Linux-only input is excluded, recursive assets refuse.
  Pure controls reject malformed/duplicate/trailing JSON, stale provenance,
  bad schema/protocol/toolchain/machines/hashes/lengths, concatenated gzip members,
  noncanonical compression/header, corruption and truncation. Check comparison
  preserves corrupt/missing bytes and modification timestamps.
- Native Windows386/WOW64 pure asset tests passed, loading/validating both images
  and testing the same corruption cases. An actual 386 `helpergen -check` invocation
  refused the noncanonical builder before generation. This is no broker execution
  or ARM64 native-runtime claim.
- Full Composition architecture test package passed. All twelve actual target
  selections passed architecture/type/dependency checks. All twelve ordinary
  public cmd/gh-tree cross-builds passed and outputs were hashed, then removed
  from the owned temporary output directory. CLI still uses its legacy entry;
  those builds establish packaging availability, not Runtime cutover.
- Final exact-product `b72ad94` `go run ./internal/runtime/cmd/helpergen -check`
  passed with the complete 900-input digest above, after two new independent
  clean builds per target. An external Python observer captured every Runtime
  file's SHA256, length and nanosecond modification time before/after, and checked
  the complete file set: all 40 files were unchanged. Checkout stayed clean.
- An owned clean `git archive b72ad94` extraction, with GOPROXY=off/GOSUMDB=off,
  GOTOOLCHAIN=local/GOWORK=off/GOFLAGS=-mod=readonly/CGO_ENABLED=0, passed ordinary
  `go build -trimpath -o <owned-temp>/gh-tree.exe ./cmd/gh-tree`,
  `go build ./internal/runtime/...`, and `go test ./internal/runtime/brokerassets
  -count=1`. No generator ran in that clean tree, and no dependency download was
  permitted. The owned temporary archive extraction/output was removed afterward.

Exact final check output (exit0):

```text
helpers verified: canonical windows/amd64 go1.25.0; two clean builds per target; source closure e3a5f8b97f9b5e02fc389097429c98226146d202ba49fe398a843fd6cf06045c (900 inputs)
No-rewrite exact bytes/lengths/mtimes/file set: True files= 40
```

## Remaining gates and exact next action

Master requests a fresh independent reviewer for the frozen helper candidate,
including the narrow Composition policy delta. Windows source has advanced to
`fb1a9ec5a46f8b886ff226e32072abdef593648b` after this branch's base; these assets
must not be represented as current for that source. Wait for a coherent Windows
freeze, obtain Master-authorized adoption, then regenerate/check reviewed assets
against that exact closure. No transient source synchronization is performed.
The current candidate is frozen for review against its fixed real ab608 closure;
the author stops after this report-only evidence checkpoint is pushed.

Runtime Sessions/parent integration, extraction native tests, native Windows ARM64
and emulation, full ABI/fault/ConPTY/lifecycle tests, other native platforms,
independent review, serial/integrated CI and all V-RUN/V-REL/full Slice gates
remain separate required work. The blocked Git review is untouched. This report
does not accept a complete Runtime contribution, Slice, M3 stage or release.
