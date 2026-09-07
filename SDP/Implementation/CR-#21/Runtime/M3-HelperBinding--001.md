# M3 committed helper-image binding — Issue #72

Status: frozen committed-image binding candidate, ready for separate independent review. Technical source cf160729581911c1cb02a24a458ec703337fb063 is clean/pushed and all20 exact-source CI jobs pass. Earlier milestone statements below are historical; final evidence and next action appear at the end. Worker owns only the three generated assets, exact Runtime root helper_binding_windows_test.go exception and this report. No product host cutover or Runtime/Slice acceptance.

Authority: #72 under #65/#69/#70/#71/#21, Sprint-004-v04/I-03/M3; Master e43297acbbff385315062bbaec7b5a58c79678a8 / ledger94. Clean preparation 7e19d617ae430cd2c7ad414b9ea4622eb1e00617 combines independently accepted Windows review9b06f8bd (technical bd78deafd4dd36e22d5b106eb7ef9c4edcd2e832) and helper review6f385a9c (technical b6f161f5189d70b66b95129237b51f9984d58e35). Broker tree is 04afda252be325beaa6bf1f22c154a094b8daed9. No reviewed native, generator, loader, module, workflow or frozen source was changed.

Native Windows amd64 Go1.25.0 at C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe executed `go run ./internal/runtime/cmd/helpergen`, then the exact `go run ./internal/runtime/cmd/helpergen -check`. Both succeeded, with two independent builds per target each. Checker exit0 in28.264s. An external Python3.14 UTF-8 observer verified all75 Runtime/go.mod/go.sum files retained their complete set, bytes, lengths and nanosecond mtimes across -check, and independently matched each normalized repository source entry to the manifest.

The unchanged accepted generator captures the complete selected source/assembly/include/embed/module/toolchain recipe into guarded isolated snapshots, with continuous one-shot native name invalidation through actual compiler exit and joined cancellation. CGO0, explicit microarchitecture, trimpath/buildvcsfalse/PGOoff/emptybuildid/internal-link and deterministic gzip remain intact. Ordinary build/run/install uses committed inputs and performs no helper compilation or download; there are three generated outputs and twelve public assets.

Source closure:955 inputs; digest23bf82b051123cd1aa31c5a2368d1cc732f4b09cbc33ea2c9abf4f08f0cfdde5. Manifest SHA25603c9db1f329aa29a4d5b4c279dcc2fc17ac0cd42967d7b4527c2bef5c482a3e8. No self-referential commit SHA is embedded.

| Target | Image length / SHA256 | Gzip length / SHA256 |
|---|---|---|
| amd64 0x8664 |3612160 / bfdb8eb2ec496222b8033bbeca2331c319fd1a0cfcacb6bb7adf3e79c138781c|2109197 / b62f559568d69b369ea5cf2ecef93d56f65ae0c5f3d03ab4241f7b64a447f44b|
| ARM64 0xAA64 |3422208 / 617b7b8a06b5333e146af86a63d5c93f89548f98ab9ab240e9a6600aa34dc5bf|1968006 / 54ecab0a18d9ca2bc762cb75baf8e9d6e07190c1370efdd86df41d6206b032c7|

At the asset-only milestone, the next permitted action was to publish the coherent partial result, then add real committed-image native parent/extraction/client binding tests in the explicitly owned root file. Actual amd64/386 and ARM64/emulated execution, typed failure/retained receipt compatibility, clean owned teardown, exact CI/race/vet/architecture/twelve builds and separate independent binding review remain open. Earlier native ABI/loader/TLS/ACL and generator adversarial reviews are reused at their exact unchanged source; their standalone acceptance is not this binding proof. Policy-rejected residues and unrelated review/access controls remain untouched.

## Committed-image execution candidate

Generated-asset milestone b8271a2c11bfdc024fe859cc82ada43274a87da2 was pushed before test implementation. The sole additional source file is internal/runtime/helper_binding_windows_test.go; it contains ordinary selected test parent/user fixtures, with no TestMain or product/private dispatcher change. It imports both private packages from the legal Runtime root; broker never imports assets. Compiled fixture binaries act only as parent or user process, never as the broker under proof.

The native matrix driver verifies its actual/compiled machine, independently reads the repository manifest and passes its target hash to each executing emulated parent. Native amd64 runs386; native ARM64 runs amd64 and386 and additionally asserts its own binary embeds neither payload. The parent uses actual MachineRoute, brokerassets.Load, ExtractWindowsImage and StartWindows. It hashes the selected in-memory and retained disk bytes, observes the uniquely extracted executing broker through a retained query/synchronize handle, and checks native helper architecture. Started facts and the actual native user process verify target machine and target-specific initial breakpoint. The fixture reads relative cwd data and verifies no startup anchor/debugger survives into user code.

Five cases exercise actual committed images: ConPTY resize101x31/input and normal close; blocked pipe input with65,484-byte delivery then canceled pending receipt, second-input refusal, coalesced Stop before receipt observation and repeated known partial receipt through an already canceled context; over1MiB raw output drained into the real256KiB Runtime ring without a UI consumer; missing executable with closed NotFound/ProcessContainment failure; stale cwd with distinct Cwd/CwdAcquisition failure. Every owner joins; final cleanup must have no residuals, exact extracted image and empty nonce directory must be absent, retained helper and user handles must signal exit. Failure cases must not execute user markers and retain their original classification after cleanup. The old generic-failure image cannot pass these typed-failure expectations.

Local Windows amd64 Go1.25.0 targeted execution passed twice after harness compilation correction (SessionSequence uses its real validated constructor). Latest run: `go test ./internal/runtime -run '^TestWindowsCommittedHelperBinding$' -count=1 -timeout=180s -v`, exit0/package5.440s; executing386 parent3.54s, all five cases PASS. Actual parent image014c/native8664, helper native8664/emulation0000, manifest/image hashes above. Windowsamd64 architecture and Runtime vet passed before the final retained-user-handle assertion; complete final-source gates and native ARM64 CI remain pending. No ARM64 local execution or full public CLI cutover is claimed.

At source5d0216f777e180fff65a8ef28ea5beb6f8f172e1, Runtime race tests passed9.704s; vet and Windows386 architecture passed; exact canonical checker passed28.462s, independently verifying two builds per target and unchanged complete76-file bytes/size/mtime/set. Asset-only CI34078342524 at b8271a2 completed all20 jobs successfully, including canonical helper verification, native Windowsamd64/ARM64 and all12 architecture/build targets. This is the pre-binding-test milestone, not ARM64 binding-test proof.

A final harness-only refinement marks the matrix driver nonapplicable under already emulated or native32 callers (native32 has no embedded-helper route), matching the accepted native broker matrix pattern. Nativeamd64/ARM64 drivers still require every listed emulated parent and every case, with no fallback/skip. The exact selected fixture runs those cases inside each child. Shared bounded-output rendering was deduplicated. Targeted native execution then passed5.354s and Runtime vet passed. No generated/native/loader/recipe source changed, so source closure/image hashes remain those above. Final exact-source CI is required after this refinement; no earlier run is relabeled as its execution evidence.


## Final frozen evidence and handoff

Technical source: cf160729581911c1cb02a24a458ec703337fb063, branch codex/cr21-runtime-helper-binding, worktree C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-binding. The final report checkpoint changes only this Markdown file with [skip ci]; executable/source equivalence is checked. Complete diff from preparation is exactly the three generated assets, the root binding test and this report. Broker tree04afda252be325beaa6bf1f22c154a094b8daed9 and helpergen treeee05f85bd3b3f3ed6a63535ea41bc3e873821dd7 remain independently accepted source. No interface/source gap or native correction was needed.

[CI34078800342](https://github.com/Hans-Einar/gh-tree/actions/runs/34078800342), attempt1, completed SUCCESS at exact cf16072 with all20 jobs SUCCESS:

| Gate | Exact evidence |
|---|---|
| Native Windows amd64 | Job101610045608 logs Go1.25.0 windows/amd64 and exact source; full Runtime42.659s, broker77.129s, assets6.796s, helpergen152.423s; full test/vet/ordinary product build succeed. The native root test compiles and executes386 parent with committed amd64 bytes. |
| Native Windows ARM64 | Job101610045459 logs host Arm64, Go1.25.0 windows/arm64 and exact source; full Runtime14.210s, broker33.782s, assets2.630s, helpergen4.501s; full test/vet/build succeed. At that architecture the exact root test requires BOTH executing amd64 and386 parent cases with committed ARM64 bytes; native driver asserts neither payload is embedded in its own ARM64 executable. No applicable case is skipped. |
| Canonical helper binding | Job101610119625 executes the unchanged exact native Go1.25.0 -check interface: two clean builds per target;955 inputs/digest23bf82b0 above; clean checkout and ordinary committed-input build pass. |
| Other native/race | Linux, macOS, ordinary-user FreeBSD and Linux full race jobs succeed. |
| All12 targets | Every accepted architecture policy and exact public-asset cross-build job succeeds. Cross-builds remain distinct from the Windows execution evidence. |
| Local exact-source race | `go test -race ./internal/runtime -count=1 -timeout=180s -v` at cf16072: exit0/package8.990s, wall10.998s. Includes registry/queue/ring/event races plus actual committed-image matrix. Parent386 is CGO0 (386 has no race instrumentation); native driver/user fixture is race-built. |

The successful default CI `go test` log prints package outcomes, not individual passing test logs. The exact test body and native-environment records establish the required matrix ran; the separate local verbose log exposes its actual five-case execution, machine/hash values and completion marker. The immediately preceding5d0216f CI34078666284 was canceled by the later source push and is not passing evidence. Preparation7e19d61 CI34077765643 failed with stale images; the explicit partial boundary is preserved. Asset-only b827 CI34078342524 passed20 before binding tests.

Small raw UTF-8 logs (only CRLF normalized to LF) and final CI metadata are retained locally at C:/Users/hanse/.codex/tmp/cr21-helper-binding-cf16072/. They are ordinary test logs, not additional product outputs or a duplicate binary/source archive. CI also retains their public source run above.

| Log | SHA256 |
|---|---|
| native-root-race.log |5cb53fd566ff4d025ff46469b65466f2809a3167f067e14093609d1d80c507b6|
| ci-amd64.log |148c2028f094343c76c935f5c423ffd87d6bed3df9b125bea07d6c87d1179493|
| ci-arm64.log |1472def138c41d322dc81840785152833936836032539c40ca44c14ed1cf3b0a|
| ci-canonical-helper.log |ed500f224112be06b7024c2dd10fae02613632714518515c9000dfb38ffc06c0|

Next permitted action: Master dispatches a separate fresh independent binding reviewer against this clean/pushed candidate and actual source/CI/image evidence. The worker stops; no source update, native merge, canonical/main integration, PR, Slice closure or release is authorized by this handoff. The unchanged comprehensive accepted native engine/ABI/DLL/TLS/ACL/extraction and generator invalidation evidence is reused; this report does not independently accept it again. #71 parent Sessions adoption and complete Runtime/native/full-stack/host/private-mode cutover remain later gates, with Master-only serial integration after Git/GitHub/Persistence/Discovery. The current public CLI still runs legacy behavior. No policy-rejected residue, unrelated worktree or blocked review/access/model control was touched. No dirty work or unpublished commit remains after the report push.
