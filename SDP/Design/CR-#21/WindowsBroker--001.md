# WindowsBroker--001 — native startup barrier and retained capability

State: DRAFT correction for DES52-H04 under #52; normative REFDES appendix.
This selects the mechanisms demonstrated in Feasibility/Evidence/windows-startup
and windows-broker, with the complete build/ownership/verification rules below.
It supersedes release of cwd guards merely when CreateProcess returns.

## Ownership and native architecture

Use one Windows Runtime broker protocol for launch and interactive sessions.
The main Runtime registry alone allocates SessionIDs and owns public lifecycle,
operation correlation, output history and final cleanup acceptance. A broker is
a private Runtime OS worker for one session, never another registry/provider or
an Application-facing service. Git and other adapters may not use it.

Choose the broker architecture from IsWow64Process2 nativeMachine and the actual
extension image machine. When already native, launch the same trusted executable
in Runtime-private broker mode. An emulated extension uses an embedded native
broker: amd64 for386 on amd64; ARM64 for386 or amd64 emulation on ARM64. Native32
Windows uses the current386 executable's private broker mode. This retains386
calling native64 programs; the real DEBUG creation error50 is not a feature-removal
policy. Unknown architectures cannot guess a broker or startup ABI.

Composition's early private-mode dispatch remains tiny and precedes normal CLI,
config/auth/default-path bootstrap. Runtime validates inherited anonymous endpoints,
protocol version, role, nonce, parent identity and session binding before any user
invocation. No public PID/signal/debugger API or network listener is introduced.
The broker does no cwd-dependent project work before the authenticated Start.

The parent creates a noninheritable kill-on-close outer Job and the suspended
broker. Assign it before first Resume, with no breakaway. Never give the broker
an inherited/duplicated outer-Job handle that would prevent last-parent-handle
cleanup after parent death. Broker creates a nested user Job before the actual
user root resumes. HPCON remains broker-local; it is not a duplicable kernel
handle or pointer that may be sent to another process.

## Scope guards and actual startup barrier

1. Broker independently acquires and validates the canonical effective directory
   chain against CwdObservation. Use actual directory read/list access, tested as
   GENERIC_READ, with READ|WRITE sharing excluding DELETE. Metadata-only handles
   are not interlocks. Inspect no-reparse attributes and volume/file identities.
2. Keep final cwd nonempty through actual acquisition. Pin a suitable existing
   direct child with real FILE_READ_DATA/GENERIC_READ (or directory list/read)
   access and no DELETE sharing, or create an exclusive random Runtime anchor
   with actual data access. Record its exact identity/ownership. The strong pin
   resists native POSIX replacement/deletion; all effective ancestors also have
   pinned child entries. Refuse if this cannot be established. Never overwrite
   a user entry or retain a normal manifest lock for the entire session.
3. On one runtime.LockOSThread owner, create the user root with
   DEBUG_ONLY_THIS_PROCESS|CREATE_SUSPENDED plus the appropriate Unicode,
   restricted-handle-list or ConPTY attributes. Assign the inner Job before Resume.
   Begin output draining and retain all cwd guards/anchor. CreateProcess return
   is not cwd acquisition: native tests show its handle is still zero then.
4. Drive bounded WaitForDebugEventEx on the creating OS thread. Own/close debug
   image/DLL handles and track process/thread handles independently. Continue
   only the supported architecture's expected loader events. Startup deadline is
   initially30s, cancellable through owned Job teardown; it is not a fixed sleep
   used as proof. Do not continue an unknown exception to find a later breakpoint.
5. Stop at the target runtime's initial breakpoint while its threads remain
   suspended. Native64/native32 and WOW64 have explicit event/PEB profiles. For
   WOW64 under native64, the first0x80000003 event is too early; the later
   0x4000001f event is the demonstrated barrier. ARM64/emulation must have its
   own verified sequence, never an “any first breakpoint” rule.
6. Read only PEB -> ProcessParameters -> CurrentDirectory.Handle using the target
   ABI. Derive/assert supported layouts from pinned x/sys types; a native64 broker
   selects WOW64 PEB through ProcessWow64Information. Require complete reads and
   a nonzero handle, DuplicateHandle with SAME_ACCESS, and compare FileIdInfo
   with the retained selected-directory guard. The duplicate may lack attribute
   query rights: inspect reparse/type on the guard and bind via FileIdInfo equality.
   No process-memory write, untagged handle injection or guessed private flags.
7. While that debug event is still pending, release the exact owned temporary
   anchor and directory guards, then DebugActiveProcessStop. Do not continue the
   event and detach later after user code runs. Native tests establish successful
   pending-event detach, no debugger/anchor visible to the user fixture, and an
   application immediately changing its own cwd afterward. Only successful barrier
   and detach permit Started. A missing/mismatched identity, unsupported profile,
   unexpected exception or detach failure is failed-start cleanup, not success.

The barrier establishes the actual initial cwd object. It does not sandbox later
chdir or absolute pathnames synthesized from cached cwd text. Native post-barrier
reparse experiments show this distinction: relative access through the acquired
cwd refused rather than selecting the replacement, while an explicit absolute
path could follow the changed namespace. Do not promise immutable future project
contents or continuous pathname ancestry. Transient startup debugging must pass
normal tool/shell, static DLL/TLS, debug-heap and immediate-chdir compatibility
tests; IsDebuggerPresent=false alone is not the entire compatibility proof.

## Control, output and cleanup

Use separate control and output transports. Control frames are length-prefixed,
at most64KiB, versioned and bound to fresh nonce plus monotonic sequence. Closed
opcodes are Start, WriteInput, Resize, Interrupt, Stop, Abort, Release and the
corresponding facts/errors. Runtime allocates the public IDs before sending them;
broker cannot invent IDs or accept provider-based routing. Reject malformed,
replayed, foreign or out-of-order frames. Keep secrets/nonce off argv/environment.

For output, parent owns the single reader/ring after an explicit kernel-pipe
handle transfer; broker closes its transferred reader copy after acknowledgment.
No two readers compete for the same byte stream. Parent drains independently of
UI progress, including while broker closes HPCON. Input/resize remain bounded
serialized broker controls; HPCON creation/resize/close stays in its owning process.
Protocol/I/O closure and parent EOF are lifecycle inputs, not rollback claims.

Normal Stop: broker stops/forces the inner user Job, proves zero user membership,
waits the exact root, closes terminal/input resources and establishes its I/O
barriers. It reports Quiescent. Parent sends Release and joins broker exit, then
terminates any remaining members through the retained outer Job, proves outer
membership zero and joins its control/output owners. Only then is Session Cleaned
and its one reliable cleanup event transferable under API A6.

Do not require outer Job membership to equal only the broker before Release.
The actual ConPTY fixture retained terminal auxiliaries until broker exit; that
condition caused a circular wait. Query-restricted auxiliary image names do not
weaken the kernel Job barrier. Quiescent is not Stopped/Cleaned. Forced parent
abort may terminate the whole outer Job, then requires broker/root handles,
outer membership and parent I/O joins; it does not pretend a graceful broker
ClosePseudoConsole callback ran after that process was killed.

Every partial start/debug/pipe/Job/extraction acquisition is retained until closed
or reported residual. Restart waits the complete old outer barrier before a new
SessionID/broker. No UI disappearance, caller timeout or full event queue abandons
cleanup. Parent crash closes its last outer-Job handle; broker death is observed
and parent forces the remaining outer membership. Native proof must cover these
paths, not only the three passing normal/forced fixture captures.

## Physical helper inputs and ordinary builds

Runtime owns these private subpackages and derived inputs, in its layer directory:

- `internal/runtime/broker`: native OS worker, startup/debug/ConPTY/inner-Job engine
  and private control types. It imports API/Domain contract values as needed but
  never the parent registry, embedded assets, Application coordinator or adapters.
- `internal/runtime/broker/cmd`: Windows-only private broker build entry. It calls
  the broker engine and exits with its already-cleaned result; no public CLI/UI.
- `internal/runtime/brokerassets`: architecture-selected embedded images and
  provenance manifest; main Runtime client uses them for emulated-host routing.
- `internal/runtime/cmd/helpergen`: build-only generator/verifier for those Runtime
  images. Composition owns its CI/release invocation and asset packaging gates.

The helper dependency graph excludes brokerassets, so generating a helper cannot
embed itself recursively. Ordinary product builds include already committed
generated payload inputs: `broker-amd64.gz`, `broker-arm64.gz` plus manifest. A386
Windows product includes both host-native images; amd64 includes ARM64 for native-
ARM64 emulation; native ARM64 uses its same-executable broker. There are still
exactly twelve public assets, with their existing names. No helper download,
runtime compiler, extra published asset or manual post-install step is required.

Generate on a canonical Windows/amd64 Go1.25.0 toolchain with pinned modules,
CGO_ENABLED=0, explicit target/default microarchitecture settings,
`-trimpath -buildvcs=false -ldflags=-buildid=`. Deterministic gzip uses fixed
compression/header fields, zero timestamp and no source pathname. The manifest
records protocol/schema, machine, byte lengths, compressed/uncompressed SHA256,
toolchain/modules/options and a sorted source-closure digest. Derive that closure
from target-selected helper dependencies, normalize repository source line endings,
and exclude generated outputs by dependency topology. Do not embed a self-referential
commit SHA in a generated artifact belonging to that same commit.

CI independently rebuilds both helper targets on the canonical builder, verifies
source closure, PE machine and exact bytes, then tests ordinary clean-checkout
go build without generation/download. Runtime changes affecting the helper closure
require regenerated reviewed inputs. Cross-host reproducibility is tested before
claiming it; the canonical builder is the acceptance source. Final release evidence
ties these embedded bytes to the exact tested product commit and workflow.

Extraction uses an exclusive cryptographic nonce directory with protected
current-user/SYSTEM DACL, complete native no-reparse parent/child acquisition and
actual data-read guards. Create image with CREATE_NEW, fully write/flush, reopen
read-only with READ sharing only, verify original FileIdInfo, machine/protocol and
embedded hash, and retain that no-write/no-delete image guard through broker exit.
Never execute a cache path merely because its name/version matches. Runtime's
private helper code performs no untrusted cwd-dependent bootstrap before handshake.
After complete outer cleanup, delete only exact owned image/empty directories;
preserve unexpected added/replaced entries and report residual cleanup. No broad
recursive removal or filename/PID/age-only trust is authorized.

## Evidence and platform acceptance

Archived native sources/logs demonstrate samebit and WOW64 startup barriers,
actual386 -> embedded native64 broker -> native64 root/descendant, Job assignment
before both resumes, ConPTY input/output/100x30 resize, normal broker release and
forced outer cleanup to zero. Two controlled helper builds match byte-for-byte;
the ARM64 helper compiles as machine0xAA64. These are bounded mechanism proofs,
not finished Runtime, full ABI/ConPTY/fault tests or native ARM64 execution.

Require native Windows amd64 with386/WOW64 combinations and native Windows ARM64
with supported x86/x64 emulation paths. GitHub currently documents windows-11-arm
as a hosted runner, making that a concrete planned CI route. Retain native32 ABI
tests and all twelve cross-builds. Unknown event/layout profiles must be resolved
and tested before release, not silently used to remove an existing architecture.
The required detailed checks are V-RUN-01/03/05/07/08 and V-REL-01/04.

[GitHub hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
