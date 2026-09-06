# Windows actual-cwd startup barrier feasibility — #52 / DES52-H04

Design-only bounded followup, 2026-09-06. Reviewed worktree:
`C:\Users\hanse\GIT\gh-tree-wt\refactor-design`, frozen
`8c109b1e4308eb301d7c1418292044d714c2f33b` / PR54. No product, frozen
worktree, Issue, branch or ref was changed. Native environment: Windows
10.0.26200, local NTFS, Python3.14 amd64; the separate 386 executable was
built using the existing Go1.25.0 toolchain, stdlib only.

## Disposition

A startup debug barrier plus actual child cwd-handle identity verification is
concretely feasible on the tested 64->64, 64->32, and 32->32 launch profiles.
It permits immediate application chdir and removes the temporary nonempty anchor
before application code runs. It is **not yet a complete selected replacement**
for arbitrary Windows launch: a 32-bit Runtime cannot create a debugged native
64-bit child; the native call fails ERROR_NOT_SUPPORTED (50). The nondebugged
32->64 control succeeds. Preserving that existing launch capability would require
a native 64-bit Runtime helper (with a separately designed trusted delivery,
control and Job/ConPTY lifecycle) or another proven barrier. A same-executable
32-bit shim cannot remove this restriction. No helper artifact policy is selected
or implicitly approved by this report.

Neither a fixed sleep, sampled path recheck, CreateProcess return nor the first
unqualified breakpoint is a sufficient acquisition barrier. Ordinary inherited
handles and untagged CurrentDirectory.Handle injection both failed natively.

## Actual results

All source and the complete per-case stdout/stderr/return codes are archived
beside this report; `results.json` is the consolidated capture.

| Case | Observed evidence |
|---|---|
| 64-bit Python debugger -> 64-bit Python child, reparse after barrier | CREATE_PROCESS event cwd handle 0; first runtime initial breakpoint handle nonzero. DuplicateHandle + FileIdInfo exactly matches the retained intended directory. Child output absent while stopped. Anchor removed, directory junction conversion succeeds, debugger detached while event pending. Child sees IsDebuggerPresent=0. Python open('marker') refuses EINVAL; explicit absolute-path open made from os.getcwd() reads the new outside pathname. |
| 64-bit Python child immediately changes its cwd | Correct accepted initial handle identity at breakpoint, then detach; child's first recorded cwd is selected target, immediate os.chdir(outside) and relative read work. No need to wait for a later marker or retain a session-long anchor. |
| 64-bit debugger -> native64 cmd.exe | Correct bound FileIdInfo at first initial breakpoint, no output before barrier; detached cmd reports selected cwd. |
| 64-bit debugger -> WOW64 cmd.exe | First native 0x80000003 breakpoint still has **zero WOW64 cwd handle**. The later WOW64 0x4000001f breakpoint has the correct nonzero cwd handle/FileIdInfo. Detached cmd reports selected cwd. |
| Actual Go386 debugger -> WOW64 cmd.exe | 32-bit initial 0x80000003 breakpoint has the correct cwd identity; child output absent then reports intended cwd after detach. Go debugger locks its OS thread. |
| Actual Go386 debugger -> native64 cmd.exe | CreateProcessW with DEBUG_ONLY_THIS_PROCESS|CREATE_SUSPENDED fails ERROR_NOT_SUPPORTED50 before the child exists. This is a capability limitation, not a timeout or memory-reader bug. |
| Actual Go386 -> native64 cmd.exe, no debug flag control | CreateProcess succeeds, NtWow64QueryInformationProcess64/NtWow64ReadVirtualMemory64 successfully read the >4GiB PEB and show zero cwd handle before Resume. With the anchor retained through this short fixture's exit, cmd runs at intended cwd. This control does not propose session-long anchors. |
| lpCurrentDirectory=NULL, bInheritHandles=true | Suspended child's cwd handle remains zero. Closing anchor, then converting to junction before Resume makes child read outside. |
| Untagged handle injection | DuplicateHandle places the accepted real directory in suspended child. WriteProcessMemory writes that remote handle into its zero CurrentDirectory.Handle. After anchor removal/reparse/Resume, child still reads outside. This is a rejected experiment, not a proposed process-memory-writing design. No private low-bit tags/flags were guessed. |

The relative/absolute distinction is material: the barrier binds the child's
actual initial cwd object; it cannot sandbox later pathname operations or the
child's own chdir. An absolute locator synthesized from cached cwd text can follow
a later namespace change. Do not claim this protocol protects arbitrary future
relative-path libraries that first convert their inputs into absolute pathnames.

## Implementable protocol for the proven debug-capable profiles

1. Acquire and validate the entire effective path chain using actual directory
   list/read access and no DELETE sharing. Retain an actual data-read child entry
   of final cwd or an exclusive Runtime temporary anchor with no DELETE sharing.
   Keep final cwd nonempty; metadata-only pins are forbidden. Use the separate
   anchor followup evidence for class65 POSIX rename-over/deletion resistance.
2. One Runtime owner goroutine calls runtime.LockOSThread before native process
   creation and keeps that thread through debug detachment/failed-start teardown.
   Own all normal output/ConPTY resources and the session Job first. Create with
   DEBUG_ONLY_THIS_PROCESS|CREATE_SUSPENDED plus existing Unicode/environment and
   handle-list/ConPTY flags. Assign Job before ResumeThread; no breakaway fallback.
3. Resume once while every guard/anchor remains held. Drive WaitForDebugEventEx
   on that same OS thread with bounded waits and a startup deadline. Close debug
   image/DLL file handles; track debug-created process/thread handles separately
   from ProcessInformation handles. Continue only the expected loader events.
   Output drains and cancellation remain independently live.
4. Stop at the **target runtime's initial breakpoint**: native64/native32 in the
   appropriate debugger, or the WOW64 initial event after its native bootstrap
   event. Do not accept the first numeric breakpoint without architecture state,
   and do not continue past a missing/mismatched identity to hunt for a later
   application breakpoint. Unsupported sequence, exception or architecture fails
   startup while the target remains controlled.
5. With all target threads suspended by the pending debug event, read only the
   required PEB -> ProcessParameters -> CurrentDirectory.Handle fields for the
   target ABI. The tested offsets are x64 PEB+0x20/parameters+0x48; x86
   PEB+0x10/parameters+0x2c. Pinned x/sys0.44 defines the matching named structures;
   derive/assert layouts, validate reads and reject unsupported layouts. A 64-bit
   debugger selects the WOW64 PEB with NtQueryInformationProcess class26.
6. Require nonzero handle; DuplicateHandle from the child with SAME_ACCESS;
   GetFileInformationByHandleEx(FileIdInfo) must match the expected retained
   directory's volume/file ID. Close the duplicate. The child's cwd handle lacks
   FILE_READ_ATTRIBUTES in these native runs: querying FileAttributeTagInfo on
   that duplicate fails access-denied. Inspect directory/reparse attributes on
   the retained Runtime guard and use FileIdInfo equality to bind that object.
7. While still at that pending event, remove/close the exact owned temporary
   anchor and release directory guards, then DebugActiveProcessStop. Do not
   ContinueDebugEvent and later attempt a detach after letting user code run.
   Native tests show pending-event detach resumes normally and the child reports
   no attached debugger. A detach failure is failed-start cleanup, not success.
   Retain the ordinary Job/process/output owners for the existing session contract.
8. Cancellation, missing/mismatched cwd, deadline, resume or detach failure
   terminates the owned Job/controlled root and drains/detaches the debug owner;
   join exact handles and I/O using the full Runtime cleanup contract. Failure to
   prove cleanup retains a residual session. The prototype is not a complete
   every-failure production unwind or debug-handle leak test.

## Evidence limits and source support

Microsoft documents that the debugger's initial breakpoint precedes static DLL
initialization, and that a pending debug event suspends all affected process
threads. The source does **not** promise the private cwd-field layout forever;
actual handle reads and identity checks are necessary, with explicit supported
ABI/platform refusal. The WOW64 second-breakpoint requirement is concrete native
evidence, not inferred from the broad initial-breakpoint sentence.

- [Microsoft initial breakpoint](https://learn.microsoft.com/en-us/windows-hardware/drivers/debugger/initial-breakpoint)
- [Microsoft debugging events](https://learn.microsoft.com/en-us/windows/win32/debug/debugging-events)
- [Microsoft WaitForDebugEvent thread/handle contract](https://learn.microsoft.com/en-us/windows/win32/api/debugapi/nf-debugapi-waitfordebugevent)
- [Microsoft process debug creation flags](https://learn.microsoft.com/en-us/windows/win32/debug/process-functions-for-debugging)
- [Microsoft detachment](https://learn.microsoft.com/en-us/windows/win32/api/debugapi/nf-debugapi-debugactiveprocessstop)
- [Microsoft debugger architecture selection](https://learn.microsoft.com/en-us/windows-hardware/drivers/debugger/choosing-a-32-bit-or-64-bit-debugger-package)
- [Microsoft PEB internal-layout warning](https://learn.microsoft.com/en-us/windows/win32/api/winternl/ns-winternl-peb)
- [Pinned x/sys0.44 structures](https://cs.opensource.google/go/x/sys/+/v0.44.0:windows/types_windows.go)

All successful debug cases used Job assignment before Resume. No combined ConPTY
startup-debug run, Windows ARM64 or emulation matrix, older Windows profile,
loader fallback/ACL failure, full cleanup fault matrix, static DLL/TLS fixture,
or transient debugging/debug-heap compatibility certification was executed here.
Those limits prevent claiming complete product/platform readiness. The existing
Runtime ConPTY/Job evidence does not automatically prove their combination with
this new barrier. Transient debugging can affect loader/heap state even after
detachment; IsDebuggerPresent=0 alone is not complete compatibility evidence.

Harness corrections are disclosed: x86 CurrentDirectory.Handle offset corrected
from 0x24 to 0x2c before passing captures; FileAttributeTagInfo access was removed
from the child duplicate after native access-denied (FileIdInfo still succeeds);
cmd command-line quoting was corrected to cmd's native command-string grammar.
Failed harness iterations terminated/waited their owned processes through the
Job and detached before temporary-directory teardown. Final sources/logs are the
ones hashed in SHA256.json. No earlier failed attempt is counted as a pass.

Exact next action: Master reconciles the confirmed debug-capable protocol and
the 32->64 capability/ConPTY gaps. DES52-H04 must remain HIGH until a complete
supported-profile protocol preserves required launch capability and independent
review accepts its actual evidence. No product implementation or BC freeze is
authorized by this bounded report.
