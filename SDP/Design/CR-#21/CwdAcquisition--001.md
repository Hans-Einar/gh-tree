# CwdAcquisition--001 — bind launch scope through native Start

State: ACCEPTED DESIGN (effective after PR #54 merge) correction for DES52-H04 under #52. Normative part of REFDES.
Runtime and Launch Discovery retain separate ownership; no shared filesystem
adapter or adapter-to-adapter invocation is introduced.

## Observation and authority

API CwdObservation contains selected WorktreeID, physically canonical root locator,
literal relative project components, root/project directory identity observations
and source version. Identity observations are platform-tagged values: Windows
volume/file ID, Unix device/inode plus available birth/change stamp. They are
short-lived observation tokens, not permanent Domain IDs or open OS handles.
Discovery obtains them while resolving the selected manifest/member/override and
refuses unsupported or redirected project components. Domain remains pure.

Runtime independently opens the current root and project without following child
links, validates the supplied scope/identity, then retains the acquired directory
object or effective path guards through child creation. It never trusts a token
as a reusable filesystem capability. Deleted/recreated objects, unavailable
identity, observed stamp/path drift or unsupported filesystem semantics refuse
Start before user code. Revalidate observed identities after acquiring guards;
checking a path and then opening it is insufficient. No portable token promises
identity across arbitrary deletion/reuse, filesystem rollback or hostile privileged
namespace mutation; an unprovable observation must be resolved afresh.

Scope authorizes the acquired directory object. On Unix another actor can move
that same object after acquisition; it cannot substitute a different directory
for it. Refuse observed relocation before user launch, but do not claim continuous
current-path ancestry or a sandbox against project code. Session facts retain the
accepted identity and observed actual cwd locator/diagnostic, not a fabricated
claim that a replaced old path is still its cwd.

## Unix acquisition

Open the selected root, then each relative project component using retained
directory descriptors and Openat(O_DIRECTORY|O_NOFOLLOW|O_CLOEXEC), checking Fstat
and no-follow parent entries. Reject `..`, absolute components and symlinks. The
resulting descriptor refers to the selected actual directory even if its old
pathname is replaced afterward.

Pass this descriptor on one explicitly designated private inherited descriptor
to the Runtime-owned SID supervisor, separately from control/reply pipes. The
versioned handshake binds its role, expected directory identity, nonce, SessionID
and immutable invocation. Before any user root or relative-path helper work,
supervisor validates the descriptor with Fstat, invokes `unix.Fchdir(fd)` exactly
once, closes its explicit descriptor and then creates the user process with an
empty exec.Cmd.Dir, inheriting the supervisor's acquired cwd object. The process
cwd reference itself keeps that object bound. Never chdir the multithreaded main
application process. No `/proc/self/fd` or `/dev/fd` pathname dependency is needed.

Construct the user command and resolve executable/PATH policy after supervisor
cwd acquisition; do not resolve a relative executable against the main app's cwd.
Use the copied environment and correct observed PWD; user commands inherit neither
control pipes nor cwd capability descriptors. Signal helpers use the absolute
trusted gh-tree executable and their existing private protocol. Failed Fchdir,
identity check, child Start or handshake follows the full partial-acquisition
cleanup contract. Only the supervisor, not user shell/provider strings, owns it.

Local Go1.25 source confirms an empty Cmd.Dir inherits the calling process cwd;
its Linux/FreeBSD/Darwin child setup invokes pathname chdir only for nonempty Dir.
Pinned x/sys0.44 supplies Fchdir on Linux and both Darwin/three FreeBSD target
architectures. Native macOS/FreeBSD implementation tests still must execute it.

## Windows acquisition

Select WindowsBroker--001 in full. Metadata-only no-delete handles and a return
from CreateProcess are both rejected barriers by native evidence. A native-architecture
Runtime broker holds actual directory list/read guards plus a real data-read child
anchor to keep final cwd nonempty, blocking both native POSIX removal and in-place
reparse conversion. It creates the user root debugged+suspended, assigns its user
Job before Resume, then stops at the target ABI's initial runtime breakpoint.
While that event remains pending, read and duplicate the actual child cwd handle
and match FileIdInfo to the selected guard. Only then remove the exact anchor,
release guards and detach, before user initialization proceeds. No additional
path check, sleep, ordinary handle inheritance or handle injection substitutes.

The native broker preserves386-to-native64 launch and owns ConPTY locally; same-
native sessions use the same protocol. The main registry retains final outer-Job
cleanup authority. Embedded helper source/build/extraction and all architecture-
specific breakpoint/PEB profiles are governed by WindowsBroker--001. Unknown
profiles require implementation evidence, not removal of supported release targets.

The guarantee binds actual initial cwd, not future absolute-path operations or
the program's own chdir. A cached cwd string can follow a later namespace change;
do not promise a filesystem sandbox. Unsupported volume/identity/guard or startup
behavior fails with owned cleanup instead of executing an unverified fallback.

## Bounded mechanism evidence and required proof

Master created only disposable temporary fixtures and ran:

| Native probe | Observed result |
|---|---|
| Linux / existing WSL openSUSE, Python3.6, UID65534 | Open directory no-follow, rename original and create a different project at its former path, pass original descriptor to private child, Fchdir and spawn grandchild with inherited cwd: child reads selected original, replacement remains untouched; symlink acquisition refuses. PASS. |
| Windows10.0.26200 / NTFS, Python3.14 | Hold no-delete-share guards on project and effective ancestors, attempt project and ancestor rename before native subprocess creation: both fail sharing violation32; CreateProcess child reads selected original marker. PASS. |

Source snapshots are archived in Feasibility/Evidence. SHA-256:
Linux `B44892E8968A32767BD0C4EE4160774D16EEFF5B2F7169D2D44C96F836C609DC`;
Windows `B1D5A511E99B63E1DC3605882023AA6C0B74A7196F68F9ADE2F0F8214F96F9B1`.
Persistence's independent native Go probes additionally establish directory pin
and reparse-object behavior, but Runtime implements and verifies its own guards.

The original Windows rename-only probe is insufficient by itself; later in-place
reparse and suspended-start failures are preserved in DesignReview--003. The
windows-anchor/startup/broker archives establish the corrected complete mechanism,
including actual386-to-native64 ConPTY and normal/forced outer-Job cleanup.
These probes establish OS mechanisms, not the future full launch implementation.
V-RUN-01/03/04 and V-LCH-02/03 must test replacement specifically after validation
and before CreateProcess/user-root acquisition, all guard/descriptor failure
unwinds, no descriptor inheritance, rejected links, relocated original objects,
stale identity, executable resolution and actual PTY/ConPTY startup on each native
required platform. Cross-compilation alone is insufficient.
