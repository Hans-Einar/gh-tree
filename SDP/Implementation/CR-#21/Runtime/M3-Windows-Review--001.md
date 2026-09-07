# M3 Windows native review -- 001

Disposition: **CHANGES_REQUIRED** — W69-M01 and W69-M02, both MEDIUM (P2).
Date: 2026-09-07. Reviewer: fresh independent `m3_runtime_windows_review`;
not the Windows author or Master. Authority: #69 under #65/#21, Sprint-004-v04 /
I-03 / M3; Master dispatch `81cce6f278162bf2f8c71f60621c0cdde915a07d`, ledger85.

Frozen source: `codex/cr21-runtime-windows` at
`5e9643364f62dcc51a967fbf94775e5350cd2bc9`, initially clean and equal to origin,
worktree `C:/Users/hanse/GIT/gh-tree-wt/runtime-windows`. Delta base:
`46fe59c5eae3f08df6ebbcc7bcd41536f3fd6864`. The broker subtree object is
`6325b039ff3d29f2033d8ffd7890e89f3a95dc8a`. This review changes only this report
and five small evidence files. No product source, test, shared contract, module,
workflow, generated image or parent Sessions file was changed.

## Findings

### W69-M01 — canceled controls can execute and lose their delivery facts

Location: `internal/runtime/broker/client_windows.go:655` (`request`, especially
lines 667–676), and the delivery retirement at lines 488–513.

`request` sends a copied control frame before considering `ctx.Done()`. An
already canceled context therefore still causes native input, resize or ETX.
After sending, its cancellation and `done` branches return an empty
`WindowsDelivery` even when an effect is possible or a buffered completion is
available. The later `Delivered` handler deletes the pending entry and sends
the final accounting to the abandoned reply channel. Neither `NextFact` nor
`Wait` exposes that later delivery or its error.

Independent real execution, reproduced on amd64 and on an executing 386 parent
using its extracted native amd64 test helper:

- Pre-canceled Write returns `{Accepted:0 Delivered:0 Completed:false}` and
  `context canceled`, but the user process prints `INPUT CANCELED_CONTROL_EFFECT`.
- Pre-canceled Resize returns the same empty canceled result, but subsequent
  uncanceled input observes the real ConPTY geometry `SIZE 101 31`.
- After one 65,484-byte pipe fill, a second blocked Write times out. Its public
  result is empty, while the later native reply is
  `{Accepted:65484 Delivered:0 Completed:true}` plus `short write`. The probe
  observes that reply only through the unexported pending channel after the
  caller has returned. Stop completes, the pending map is empty, and final
  `Err` is nil: eventual native acceptance/delivery evidence was lost.

This is an effect-accounting defect, not an observed native handle leak. Zero
delivered bytes is the actual partial outcome in these captures; the test does
not claim that the second write delivered a positive count. `Completed:false`
also must not be interpreted as proof of no effect.

The frozen Application--Runtime cancellation and result clauses require refusal
before admission/delivery and preservation of accepted/known effects afterward;
late partial-write facts must remain available without replay. The separate
#71 parent has its own copied queue acceptance, but cannot recover native counts
discarded by this seam. This finding does not claim a tested defect in that
unassembled parent adapter.

Required correction: check cancellation at the serialized dispatch admission
point, including while waiting for control serialization and when Stop wins.
After a frame may have been sent, retain bounded ownership and expose truthful
pending/unknown and eventual known delivery facts to the parent; prefer a known
completion when cancellation/completion race. Do not repair this only with an
entry context check or report zero/no effect after possible delivery. No automatic
input replay. Cover pre-cancel, cancel while awaiting serialization, post-send
blocked input, completion/cancel races, and Stop joining/retiring the retained
request. The exact private mechanism belongs to the correcting author.

### W69-M02 — native failures erase their contract classification

Location: `internal/runtime/broker/engine_windows.go:136` (`fail`) and
`internal/runtime/broker/client_windows.go:526` (`Failure`).

The engine ignores the supplied native error and sends the identical generic
failure payload for acquisition, cwd, command lookup, ABI/debug and cleanup
failures. The client replaces even that payload with another untyped generic
error. Meaningful cause and stage information is unavailable at the parent seam.

Independent real missing-executable and stale-root-identity starts both return
only `native broker reported failure` (`*errors.joinError`); neither preserves
`ErrCwd` or `os.ErrNotExist`. Both remain unestablished and clean completely.
This positive cleanup result does not restore the lost error classification.
The source also routes native permission/unsupported-profile/cleanup errors
through the same erasure; those additional native errors were not separately
injected by this reviewer.

The frozen Errors clause distinguishes stale observation, unavailable/not-found,
unsupported profile, permission, process failure and incomplete cleanup. #71
cannot make that distinction reliably by inspecting this result or rerunning
filesystem observations after the failure.

Required correction: preserve bounded, closed, safe native failure classification
and stage through the private protocol/client facts. Retain useful original
causes locally where appropriate, without exposing paths, argv, environment,
handles or other private details in public diagnostics. Validate the failure
payload and exercise distinct stale-cwd, missing-command, permission/profile and
cleanup outcomes. Shared wire framing can remain unchanged; coordinate any
actual shared-interface change through Master. This does not authorize changing
frozen public contracts or guessing a public classification from generic text.

## Independent inspection and verification

Read root AGENTS/developmentInstructions, full #21/#65/#69/#70/#71 issue bodies
and comments, frozen Application--Runtime/BoundaryTypes/BCFreeze, accepted
WindowsBroker/CwdAcquisition/Runtime feasibility/Verification, local Runtime
README, actual shared wire/start source, and every Windows candidate source/test
including the C fixtures. The author report was used for navigation and its
claims checked against source and execution, not treated as independent proof.

Actual code review traced:

- Native-machine versus mapped-image routing; native and WOW64 PEB layouts;
  first-chance expected breakpoint sequence, exact duplicate/FileIdInfo check,
  and release/detach at the pending event before user initialization.
- Retained relative no-reparse directory acquisition, real list/data access,
  existing-child or exclusive owned nonempty anchor, and exact guard cleanup.
- Suspended broker/outer Job and user/inner Job assignment before resume;
  explicit output transfer, independent parent drain, single broker input owner,
  real ConPTY lifetime and borrowed versus independently owned debug handles.
- Inner zero/root wait/terminal join, Quiescent/Release/broker exit/outer zero
  ordering; abnormal broker/protocol/parent death forces retained containment;
  timeout keeps ownership and copied typed residuals instead of false cleanup.
- Protected exact current-user/SYSTEM ACL, exclusive image creation, full write,
  flush, retained read-only guard, full identity/hash/PE validation, and exact
  nonrecursive cleanup preserving unexpected/replaced objects.

No additional blocking finding was established in those mechanisms. Native
handles/PIDs stay private; numerical IDs identify debugger events or acquired
test process observations, not production census-to-signal authority. The
Windows delta does not import assets recursively or add a production stub.
Shared `broker/protocol.go`, `broker/start.go`, Runtime root Go files/README,
other layers, module/workflow and frozen governance are unchanged from the base.

Local environment: Windows NT 10.0.26200.0, Go1.25.0 windows/amd64; Python3.14;
installed MSVC19.43.34810 and linker14.43.34810.0 for the additional x86 fixture.
No installation, global environment change or model/access-setting change.

| Independently executed check | Result |
|---|---|
| `go test -race ./internal/runtime/broker -count=1 -timeout=180s` at frozen source | PASS, 40.347s, exit0. Includes actual startup/cwd/ABI, extraction/fault/cancel, root/descendant/parent-death, raw output, blocked input/ConPTY close, Stop and native resource-count controls. |
| `go vet ./internal/runtime/broker` | PASS, exit0. |
| Additive reviewer controls, native amd64, CGO0 | Expected FAIL, exit1, 0.972s: both findings reproduced; all created sessions joined. |
| Same additive controls, executing 386/WOW64 parent, CGO0 | Expected FAIL, exit1, 4.592s: both findings reproduced through actual native helper routing; all created sessions joined. |
| Additional real x86 static DLL/exe TLS and debug-CRT fixture, native broker to WOW64 user | PASS, exit0, 3.669s; mapped machine014c, `NATIVE_DLL_TLS_DEBUG_HEAP exe=0 dll=0`. This extends the existing native-machine loader fixture to a real WOW64 C program. |
| Product diff isolation, `git diff --check`, evidence UTF-8 decoding | PASS. |

The original suite passing does not supersede the new adverse failures. Reviewer
controls use additive external Go overlays; production files are never replaced.
The loader overlay derives a separately named test from the existing owned C
compiler/runner and selects x86. Only owned temporary directories/processes are
used. The specifically excluded earlier denied-ACL directory was untouched.

Reproduction from this checkout, with an explicit native Go path:

```powershell
$reviewGo = 'C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe'
& 'C:/Python314/python.exe' 'SDP/Verification/CR-#21/Evidence/M3-Windows-Review--001-run.py' --go $reviewGo --arch amd64
& 'C:/Python314/python.exe' 'SDP/Verification/CR-#21/Evidence/M3-Windows-Review--001-run.py' --go $reviewGo --arch 386
& 'C:/Python314/python.exe' 'SDP/Verification/CR-#21/Evidence/M3-Windows-Review--001-run.py' --go $reviewGo --wow64-loader
```

Evidence is in the existing `SDP/Verification/CR-#21/Evidence/` folder. Log text
normalizes line endings and trailing whitespace only; hashes name the committed
LF blob bytes (a Windows checkout may use CRLF).

| File | SHA-256 |
|---|---|
| M3-Windows-Review--001-controls.go.txt | B07DF85B959FCBCC52967F470109F56ED3C88CA12CA5D990A523194B69F2CF95 |
| M3-Windows-Review--001-run.py | 513BB099B59DA165AA4147C778171032794FF4EE0C3F9F2C8C0C7DAE04BEF51D |
| M3-Windows-Review--001-controls-amd64.log | 5D19D5D68BA53092CACC671A0E824C631339D8213312D64D54BDC719E4372327 |
| M3-Windows-Review--001-controls-386.log | F75D721D9DC4BEB97BCA5E20E00398779C77493E6AF523B712D416B099B6C65D |
| M3-Windows-Review--001-wow64-loader.log | 78961F3967BAFA30A7E29E74A69D580249CBDB89F689B537A0C8E0A9B264CB1D |

## Exact CI and acceptance boundary

Independently queried [run34071023665](https://github.com/Hans-Einar/gh-tree/actions/runs/34071023665),
attempt1, completed at exact source5e964336. Six independent native/race jobs
succeeded: Windows amd64101588261669, Windows ARM64101588261610, Linux101588261655,
macOS101588261569, FreeBSD101588261641 and race101588261493. The ARM64 job log
records Go1.25.0 windows/arm64 and actual broker tests PASS54.148s. This is actual
native CI, not local ARM64 execution or a cross-build inference.

The run's overall conclusion is **failure**: inventory job101588261594 reports
`M3 prerequisite missing: internal/runtime/cmd/helpergen Go source`. Dependent
cross-build job101588328603 and helper reproducibility101588328776 are skipped.
This expected separate #70 prerequisite remains explicit, not an additional
undisclosed finding against this standalone source candidate. Author-reported
local twelve-target builds are separate evidence, not independently repeated here.

No standalone candidate acceptance is granted while W69-M01/M02 are open.
After correction, reuse unaffected review/native evidence and perform bounded
independent review of the exact corrected source, focused adverse controls and
affected native CI. Then #70's separately reviewed generator/assets must regenerate
against that final broker closure and execute the committed images. Old
ab608-based bytes cannot satisfy this gate. Parent #71, full Sessions/native
assembly, actual package-manager/Discovery/host scenarios, aggregate registry
behavior, all Slices and M8 release/install proof remain separate required gates.

Next permitted action: Master dispatches a separate author for these two bounded
corrections, then freezes the corrected candidate for bounded independent re-review.
Master alone records integration. This report closes neither #65/#69 nor a Slice,
M3, the Runtime adapter or release. The unrelated blocked Git foundation review
was not retrieved, retried, substituted or rephrased.
