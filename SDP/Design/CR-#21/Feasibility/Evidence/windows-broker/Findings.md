# Native Windows broker + actual cwd + ConPTY followup — #52

Design-only bounded feasibility, 2026-09-06. Product/worktree/Issues/refs unchanged;
previous `cr52-cwd-startup-feasibility` archive remains unchanged. Native execution:
Windows10.0.26200 / NTFS, existing Go1.25.0 and x/sys0.44.0. This followup addresses
the earlier real Windows386 -> native64 debug-creation limitation by adding a
native64 Runtime broker fixture. It does not select a capability refusal.

## Result and coherent design direction

The proposed native broker is feasible in the tested 386 -> amd64 -> amd64
configuration, including real ConPTY input/output/resize, correct cwd acquisition,
debug detachment, user descendant containment, and complete outer-Job cleanup.
The native64 -> native64 configuration also passes. A native64 broker was
successfully embedded in a plain-built 386 executable, extracted into an exclusive
private nonce directory, identity/hash checked through a retained read handle,
and launched by that 386 executable without downloads or runtime compilation.

Prefer one Windows broker protocol for consistent ownership. A process already
running in the host's native architecture may invoke the same trusted executable's
Runtime-private broker mode; an emulated extension invokes its embedded host-native
broker. Main Runtime retains registry/SessionID/operation/lifecycle authority.
Broker owns only native OS launch/debug/terminal resources and correlated facts.
This is Runtime-internal execution, not a provider adapter or extra public asset.

The tested executable is a small *fixture implementation* of that boundary, not
production code or a full frozen contract. The actual user root in this test is
its native64 `user-root` mode, which starts one ordinary descendant, reports cwd /
debugger / anchor visibility, accepts terminal input and reports observed geometry.

## Captured native evidence

`results.json` and three per-case files preserve all stdout/stderr and results.

| Case | Evidence |
|---|---|
| Actual 386 parent -> embedded/extracted amd64 broker -> amd64 user root + descendant | PASS. Parent outer Job had exactly one member and broker user CPU=0 before first Resume. Broker created real ConPTY and an inner Job; user root was debugged+suspended, with exactly one inner member and user CPU=0 before Resume. |
| Actual cwd barrier combined with ConPTY | PASS. User cwd handle was zero before Resume and nonzero at the target initial breakpoint; duplicated handle FileIdInfo exactly matched the broker's retained selected directory. No USER_READY output before barrier. The pending breakpoint stayed controlled while the exact anchor and guard were removed and DebugActiveProcessStop detached. |
| Normal user behavior after detachment | PASS. USER_READY reports selected cwd, debugged=false and anchor=false. ConPTY input `probe-input` reaches the user. Resize to100x30 is reflected by GetConsoleScreenBufferInfo in the user process and output returns through ConPTY. Inner Job contains root plus ordinary descendant. |
| Normal Stop / resource join / broker release | PASS. Inner Job terminated to0; exact root process waited; input closed; ClosePseudoConsole and independent output reader joined. Quiescent is reported, parent sends Release, broker exits0, and parent proves outer Job0. The parent also issues TerminateJobObject on that retained outer Job before accepting its final zero membership, covering any remaining auxiliary process without numeric-PID signaling. |
| Forced whole-outer-Job cleanup while broker/user/descendant/terminal are active | PASS. Parent's retained outer Job termination kills broker and all ordinary contained descendants, broker retained handle signals with requested exit55, final outer Job count0. This case proves containment cleanup, not a graceful broker-owned ConPTY close callback after forced broker death. |
| Native64 parent -> native64 broker | PASS same normal cwd/terminal/containment/cleanup sequence; supports one broker protocol for same-native and cross-bitness paths. |
| Native ARM64 broker | Compile-only PASS, PE machine0xAA64. No native ARM64, CHPE, x64-on-ARM64 or 32-bit-Windows execution claim. |

## Necessary teardown correction found by the native test

Do **not** wait for the outer Job to contain only the broker before sending
Release. The actual ConPTY run had three outer members after inner Job0 and
joined ClosePseudoConsole/output: broker, conhost.exe and another member whose
image-name query returned access-denied. They disappeared after broker exit.
Requiring outer membership1 before allowing broker exit therefore creates a
circular wait on this native profile.

Correct sequence: broker proves user inner Job0, root wait and its terminal/I/O
barriers; sends Quiescent; parent sends Release; joins the broker; terminates any
remaining members through the retained outer Job; verifies outer Job0 and joins
its own control/output owners; only then may Runtime publish Cleaned. Unknown
auxiliary image names cannot substitute for the kernel Job membership barrier.
Do not call Quiescent itself Stopped/Cleaned, and do not treat HPCON close alone
as complete OS-process teardown. The broker is the last Runtime control actor;
OS terminal auxiliaries may still exist until it exits.

## Resource and authority boundary exercised

- Parent supplies only explicitly inherited anonymous kernel pipe endpoints in
  HANDLE_LIST. Versioned length-prefixed JSON frames cap payloads at64KiB and bind
  nonce and monotonic sequence across Start, Exercise, Stop, Release. No nonce is
  in argv/environment. The inherited endpoints are the authority capability;
  this fixture is not the full production malformed-frame/deadline/auth matrix.
- Parent creates a noninheritable kill-on-close outer Job and assigns the
  suspended broker before broker code can run. Broker creates a separate inner
  kill-on-close user Job; no breakaway flags are enabled. Inner assignment happens
  before user Resume, inside the inherited outer containment.
- Broker owns HPCON and its own os.File input/output endpoints. Main passes no
  HPCON pointer across processes. The fixture exercises input/resize through
  control frames and sends output as bounded diagnostic data; production may
  duplicate appropriate kernel pipe handles for independent raw stream transport,
  while keeping HPCON and ClosePseudoConsole in the broker process.
- Broker holds runtime.LockOSThread across CreateProcess/debug wait/detach.
  The startup proof reads the PEB/parameters using pinned x/sys named structures,
  duplicates the actual child cwd handle, and compares FileIdInfo. No remote
  memory write or private cwd-handle injection is used.
- The parent registry must bind Started to the actual native barrier/detachment
  and own every partial-start residual. A trusted broker is a Runtime worker,
  not a second registry or an entity allowed to invent public SessionIDs.

## Embedded helper build/extraction feasibility

The fixture uses a separate `runtimebroker` build tag to exclude embedded payloads
when building a broker, avoiding recursive embedding. `embedded_386.go` embeds the
resulting amd64 PE only for the ordinary386 extension build. A subsequent plain
`GOARCH=386 go build` succeeds without generation or runtime tools because its
payload input is already present. The ordinary extension executable remains the
sole distributed asset in this experiment.

Two broker builds with identical Go1.25.0, x/sys0.44.0 source, `-tags runtimebroker
-trimpath -buildvcs=false -ldflags=-buildid=` produced identical SHA256:
`8E4AC278DF353AAA965AC8DCB99E5B7609FD3352FBD53E6BE0FB335CB89E7B44`.
`build-metadata.json` records all binary sizes, hashes and PE machine fields.

Important build-plan requirement for Master: go:embed does **not** run generators.
For a clean checkout to support normal go build, generated helper payloads (or
equivalent generated byte-source files) must already be durable repository inputs.
Alternatively requiring a preparation command would change that build contract.
CI can rebuild helper source with pinned toolchain/modules/options and require
identical payload bytes, then package the same twelve existing asset names.
Record the helper source-closure digest separately from generated outputs, avoiding
self-referential "this commit's SHA embedded inside this commit's generated binary"
requirements. Final generation/provenance layout is Master-owned.

The extraction fixture uses an exclusive random directory with a protected DACL
granting current user and SYSTEM access; creates the payload with CREATE_NEW,
writes and syncs, reopens it with actual GENERIC_READ and READ sharing only,
checks FileIdInfo against the originally created object and SHA256 against embedded
bytes, and retains that no-write/no-delete read guard through the broker lifetime.
No shell, download or compiler runs during extraction/launch. These are owned
temporary files and are removed after broker/outer cleanup.

Production extraction must additionally use the accepted complete native
no-reparse/ancestor-guard/identity/cleanup contract for its private directory and
executable path. This small fixture assumes its controlled temporary ancestors;
it is not an adversarial proof of the complete filesystem acquisition implementation.
It must not justify deleting an unexpected replacement or arbitrary added file.

## Architecture coverage and native ARM64 profile

Choose the broker from the host native machine value, not from an assumption
that any WOW64 process runs on AMD64. Pinned x/sys0.44 exposes IsWow64Process2;
Microsoft defines its nativeMachine output explicitly. A386 extension on ARM64,
or an amd64 extension under ARM64 emulation, therefore needs the corresponding
ARM64 broker payload rather than the amd64 payload used on this test host.

The compiled ARM64 broker has machine0xAA64 and hash
`1653D7711CB5BAF02BA60CB36207FE8E40949274C2E42AAED90CC7008324ED99`.
This proves APIs/build layout are available, not native startup semantics. The
actual prototype's debug-event loop is intentionally native64-target-only; it
must not be copied as the entire architecture matrix. The earlier archive
already proves WOW64 needs its later0x4000001f breakpoint/32-bit PEB, while a
native32 debugger has a different DEBUG_EVENT/PEB layout.

Microsoft documents native ARM64 debugger support for x86/CHPE context switching
and warns that emulation can raise internally handled exceptions. Production
requires explicit target-ABI event/PEB profiles and native tests on supported ARM64
and emulation combinations. Do not silently add an "any first breakpoint" or
unknown-exception-continue fallback, or infer native ARM64 success from cross-build.
Native32 Windows also needs its ABI-specific broker path; preserving the386 asset
includes that profile, not just running386 under this host's WOW64.

- [Microsoft IsWow64Process2](https://learn.microsoft.com/en-us/windows/win32/api/wow64apiset/nf-wow64apiset-iswow64process2)
- [Microsoft ARM64 debugging](https://learn.microsoft.com/en-us/windows-hardware/drivers/debugger/debugging-arm64)
- [Microsoft machine constants](https://learn.microsoft.com/en-us/windows/win32/sysinfo/image-file-machine-constants)
- [Microsoft nested Jobs](https://learn.microsoft.com/en-us/windows/win32/procthread/nested-jobs)
- [Microsoft ConPTY lifetime guidance](https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session)

## Limits and handoff

This is bounded feasibility, not a product implementation or accepted BC. It does
not execute the full twelve release builds, native ARM64/old-Windows/native32 OS
matrix, DLL/TLS/anti-debug compatibility matrix, every partial-allocation failure,
parent crash/last-handle-close case, all malformed/control cancellation paths,
or full malicious namespace/fallback tests. Existing prior evidence remains
separate, including successful same/WOW64 cwd-handle proofs and rejected inheritance
or injection shortcuts. No normal launch capability needs to be removed to address
the formerly blocking386->native64 debug restriction.

The failed initial outer-membership1 fixture gate and diagnostic access-denied
are disclosed above. They were not passing cases. Owned outer Job termination
cleaned their processes. An empty private temporary directory from an early
failed attempt was discovered in final inventory. Automatic policy rejected a
recursive cleanup command before execution; subsequent read-only inspection
confirmed the exact owned directory was empty, and nonrecursive Remove-Item
removed it successfully. Production must wait the process/resource barrier
before directory cleanup and report failed cleanup rather than ignoring a
deferred RemoveAll error. Final passing source/logs and binary provenance are
hashed in SHA256.json; environment-cleanup.json records final fixture inventory.

Exact next action: Master integrates the native broker design, its corrected
Quiescent/Release/outer-Job final barrier, complete helper build/extraction policy
and required native platform gates; fresh independent review reopens this source
and evidence alongside the earlier cwd/anchor artifacts. No frozen repository
file, product file, Issue or ref was modified by this followup.
