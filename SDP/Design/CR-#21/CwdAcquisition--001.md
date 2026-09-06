# CwdAcquisition--001 — bind launch scope through native Start

State: DRAFT correction for DES52-H04 under #52. Normative part of REFDES.
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

Resolve the selected physical root to a canonical volume path and acquire guards
for every effective pathname component through the project directory. Open
directory handles with FILE_TRAVERSE|FILE_READ_ATTRIBUTES|SYNCHRONIZE,
FILE_FLAG_BACKUP_SEMANTICS|FILE_FLAG_OPEN_REPARSE_POINT, sharing READ|WRITE but
not DELETE; inspect directory/file identities and reject reparse objects. Relative
NtCreateFile operations with OBJ_DONT_REPARSE may implement the component walk.
Retain ancestor guards too: guarding only the final directory does not establish
the whole pathname chain. Unsupported volume/network/path profiles or sharing
conflicts return explicit refusal, never weaker check-then-CreateProcess fallback.

With guards held, independently revalidate the component chain and root/project
observation. Resolve the executable explicitly under the copied cwd/environment
policy and pass an absolute application name plus the guarded canonical cwd to
CreateProcessW. Create the process suspended, assign its Job before ResumeThread,
and keep guards until native creation establishes the child's own cwd reference.
The selected object cannot be renamed/replaced through ordinary namespace operations
in that interval because DELETE sharing is denied on the effective chain.
ConPTY uses the same acquisition, not another cwd shortcut. Release every guard
on success or failed-start cleanup; guards are noninheritable.

No guarantee covers privileged volume/mount changes or hostile same-UID process
manipulation outside these OS access semantics. Directory movement after the child
has its cwd reference does not retarget that child into a replacement project.

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

These probes establish OS mechanisms, not the future full launch implementation.
V-RUN-01/03/04 and V-LCH-02/03 must test replacement specifically after validation
and before CreateProcess/user-root acquisition, all guard/descriptor failure
unwinds, no descriptor inheritance, rejected links, relocated original objects,
stale identity, executable resolution and actual PTY/ConPTY startup on each native
required platform. Cross-compilation alone is insufficient.
