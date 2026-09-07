# M3 helper/assets independent review — Issue #70

Disposition: **CHANGES_REQUIRED**. One HIGH and two MEDIUM findings block this
bounded helper generator/assets/policy candidate. No product changes were made
by this reviewer, and no Runtime contribution, Slice, integration or release is
accepted.

Date: 2026-09-07. Role: fresh independent Reviewer, separate from the author.
Authority: #70 under #69/#65/#21; Sprint-004-v04 / I-03 / M3; Master review
dispatch `30651a81490524a8937c50417640dff4ec5241ad`, ledger CR21-0082.
The mandatory repository instructions, complete Issues and comments, frozen
Application--Runtime/BoundaryTypes 1.0.0 and BCFreeze, relevant accepted
REFDES/WindowsBroker/CwdAcquisition/Runtime feasibility, Verification--001, M1
workflow/interface and Runtime/architecture instructions governed this review.

Frozen branch: `codex/cr21-runtime-helper-assets`; clean reviewed HEAD
`d110a068db8ea68cb2d848774d79b7b9d71b0557`; technical candidate
`b72ad94c60eb38d7c4ebe70fafd10b21a3289037`; policy prerequisite
`4b31ac4c34df370b852a1626107c102ff300c6dd`. Actual reviewed delta starts at
`ab608327e63727f66ffb1aa7b3200c2865307cf5`. Its 18 paths are the author report,
helpergen, brokerassets and the three expressly delegated Composition files.
No broker native source change is included in this acceptance scope.

## Findings

### H70-H01 — build bytes are not bound to captured external provenance (HIGH / P1)

Locations: `internal/runtime/cmd/helpergen/main.go:384`, especially the
`repoPath == ""` exclusion at line 392, and the recapture comparison at line 510.
Capture records the bytes/hashes of module and toolchain inputs, but build copies
only repository files. Both supposedly independent builds continue reading the
live module cache and GOROOT. A before/after recapture does not bind intervening
reads to those recorded bytes.

Independent executable control `TestReviewerLiveModuleBuildInputs` uses only an
owned fixture module/cache. Admission and capture see the original source and
correct go.sum pin. The reviewer then changes its literal value during both
actual clean builds and restores the original before recapture. Both amd64 and
ARM64 repeated images are byte-identical, contain the changed unrecorded value,
and pass the generator's PE checks; the entire before/after manifest is identical.
`outputs` also accepts that plan/image pair. This exercises the same build and
recapture sequence used by generation, without a timing-dependent race or any
shared-cache edit. A committed result from that sequence can claim the original
source provenance for different executable bytes.

Fix: make the actual build consume verified immutable dependency/toolchain inputs
bound to the captured provenance, through an isolated snapshot or equivalent
enforceable retained ownership. Do not rely on another before/after stat/hash
check. Reverify the complete selected source, assembly/include, embed, module,
toolchain and options closure of that consumed input set. Add the reproduced
intervening-change case to meaningful regression coverage and regenerate the
reviewed manifest/assets when the recipe changes.

### H70-M02 — canonical checker requires undeclared prepopulated module metadata (MEDIUM / P2)

Locations: `internal/runtime/cmd/helpergen/main.go:121` and line 162.
The environment forces `GOPROXY=off`, then unconditional `go mod verify` loads the
whole build-list graph before selected dependency capture. The unchanged accepted
CI command has no separate dependency-preparation step. A fresh canonical builder
therefore cannot complete the required interface.

Real CI attempt 2 at the technical SHA fails helper job 101583102130 with
`github.com/charmbracelet/x/termios@v0.1.1: module lookup disabled by GOPROXY=off`.
An independent local invocation of the exact `go run ... -check` interface with
an owned empty GOMODCACHE fails similarly on the first unavailable graph module.
All 40 Runtime files retain their bytes, lengths, mtimes and file set in that
negative run. The populated local cache passing does not establish a fresh CI
builder. The author report's claim that the earlier unrelated metadata problem
was eliminated is contradicted by this source and executed CI.

Fix the generator's pinned dependency admission/preparation within the existing
command, including fresh/selected-only cache coverage. Keep go.sum verification,
no checkout rewrite, native/toolchain admission and the existing workflow gate;
do not require an undocumented developer cache or weaken CI to pass.

### H70-M03 — selected source resolution fails through the canonical GOROOT junction (MEDIUM / P2)

Location: `internal/runtime/cmd/helpergen/main.go:364`, particularly its
`filepath.EvalSymlinks` calls at lines 369/373.
Native Windows CI job 101583039279 fails the helper closure test at
`closure_test.go:24` with `The system cannot find the path specified.` Its setup-go
log records a junction from `C:\hostedtoolcache\windows\go\1.25.0\x64` to the
actual D: installation. The asset tests in that job pass.

The reviewer independently created an owned junction to the very same local
canonical Go tree. `os.Stat` successfully reads a standard-library source through
it, while `contained` rejects that existing file with the identical error.
The actual `TestActualSelectedDependencyClosure` also fails identically when
GOROOT names the junction, and passes with the physical installation path.
This local reproduction supports the CI root-cause attribution; no alternate
toolchain download or CI-only source modification was needed.

Fix Windows physical source resolution so a supported selected-root alias binds
to the actual root and source, while child redirection outside the declared
closure remains rejected. Preserve the containment requirement and add both the
root-junction positive and redirected-child negative controls. Include failing
path/stage context in errors so future acquisition failures are diagnosable.

## Independently executed evidence

Toolchain: native Windows amd64 Go1.25.0 at
`C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe`.
External observer: Python 3.14 at `C:/Python314/python.exe`; UTF-8 source/log and
subprocess decoding. All fixtures belong to
`C:/Users/hanse/.codex/tmp/cr21-helper-review-001`; no existing native fixture,
shared module-cache bytes, other layer or frozen record was modified.

| Command/control | Actual result |
|---|---|
| Exact native `go run ./internal/runtime/cmd/helpergen -check`, populated cache | PASS; two fresh builds per target; 900-input digest `e3a5f8b97f9b5e02fc389097429c98226146d202ba49fe398a843fd6cf06045c`. Independently repeated under an external before/after observer: all 40 Runtime file hashes, lengths, mtimes and file set unchanged. |
| `go test ./internal/runtime/cmd/helpergen ./internal/runtime/brokerassets ./internal/composition/architecture -count=1 -v` | PASS; actual selected source/CRLF/non-Go embed/recursive dependency controls, both real PE payloads, malformed/provenance/length/gzip/PE corruption and complete architecture fixture suite. |
| `GOARCH=386 go test ./internal/runtime/brokerassets -count=1` | PASS, actual local WOW64 execution loads both native images and runs the parser/corruption controls. |
| `GOARCH=386 go run ./internal/runtime/cmd/helpergen -check` | Expected refusal before generation: noncanonical windows/386 builder. |
| Forged owned module directory plus changed cache ziphash, original go.sum | Positive integrity control: admission alone passes, but subsequent capture's actual `go list` rejects the checksum mismatch. **No claim that an already inconsistent go.sum pin is accepted.** |
| Intervening owned module change during both builds, restored before recapture | Expected regression FAIL; H70-H01 reproduced on both target images. |
| Existing source through owned Go root junction | Expected regression FAIL; H70-M03 reproduced directly and through the actual source-closure test. |
| Exact checker with empty owned GOMODCACHE | Expected refusal reproduces H70-M02; no checkout rewrite. |

The actual delta inspection supports the narrow policy boundary: JSON/PE imports
are added only to assets; type-resolved `pe.Open` including aliases/references is
rejected; assets acquire the existing pure-symbol/goroutine/channel/linkname
restrictions. Native I/O, printing, broker-to-assets recursion and other-layer
permission expansion remain rejected by the executed fixtures.

The in-memory loader checks unique canonical JSON, provenance digests, target
order, 2 MiB manifest / 64 MiB image bounds, gzip metadata/single-member/CRC and
canonical compression, exact hashes/lengths and PE32+ machine/executable type.
Owned return bytes and Windows386/amd64/other embed selection remain correct for
the committed inputs. The helper closure includes the actual broker/API/Domain
dependencies, standard selected Go/assembly/header inputs and canonical build
binaries; assets and generator are excluded from executable dependencies. These
positive observations do not repair H70-H01's live external input consumption.

Small reusable evidence is in
`SDP/Verification/CR-#21/Evidence/Runtime-HelperAssets-Review--001/`: reviewer
source `.go.txt`, fixture preparation/observer scripts, exact positive/no-rewrite
result, regression log, compact CI metadata and the two relevant job diagnostic
excerpts. `Manifest.json` records each evidence file's length and SHA256. The
minimal synthetic fixture is provenance-mechanism evidence, not native broker
behavior proof. No source tree, binaries, module cache or broad archive is copied.

## CI and handoff

[CI run 34068851101, attempt 2](https://github.com/Hans-Einar/gh-tree/actions/runs/34068851101)
is terminal **FAILURE**, technical HEAD `b72ad94c60eb38d7c4ebe70fafd10b21a3289037`:
18 jobs succeed, 2 fail. All twelve architecture/cross-build jobs, inventory,
Linux/macOS/Windows ARM64 tests, Linux race and native FreeBSD job succeed.
Native Windows amd64 and the canonical helper job fail as above. Attempt 1 and
the redundant report-only run 34068966078 were canceled and are not passing proof.
No complete native ABI/ConPTY/Runtime acceptance is inferred from these jobs.

The images are correctly reviewed only against this branch's actual ab608 broker
closure. Newer Windows source `fb1a9ec5a46f8b886ff226e32072abdef593648b` and
`40e7e103` are not adopted here; their later coordinated adoption requires new
reviewed regeneration. The separate blocked Git review was not retried, replaced
or inspected. Serial integration authority remains with Master.

Exact next permitted action: the separate author corrects H70-H01/M02/M03 in
the authorized files, regenerates the manifest/images as required, runs the fixed
canonical interface and targeted adverse controls, then freezes/pushes a new
candidate for bounded independent re-review and exact-source CI. Reuse unaffected
parser/policy/build evidence where valid. No integration or acceptance is
permitted from this report. The reviewed product tree was clean and identical
to d110 before this report/evidence-only checkpoint; the report commit uses
`[skip ci]` and contains no product/policy/generated-asset changes.
