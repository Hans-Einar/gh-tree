# Native Windows anchor and converted-parent followup

Scope: #52 design feasibility only, Windows 10.0.26200 / local NTFS / 64-bit
Python 3.14 ctypes native calls. All fixtures are exclusive OS temporary
directories. Original archived Persistence sources/logs and frozen repository
are untouched. The three scripts and three JSON outputs in this new subdirectory
are the evidence; SHA256.json records their bytes.

## Result: actual data-access anchor works in the tested native profile

An anchor opened with `GENERIC_READ` and share 0 or READ|WRITE without DELETE
prevents both tested ways of emptying cwd:

- `NtSetInformationFile(FileRenameInformationEx=65)` with flags
  REPLACE_IF_EXISTS|POSIX_SEMANTICS (3) returns STATUS_SHARING_VIOLATION
  (0xc0000043), without replacing the anchor.
- An independent DELETE-access open fails ERROR_SHARING_VIOLATION (32), so the
  attacker cannot issue FileDispositionInformationEx against that object.
  DeleteFileW / RemoveDirectoryW also fails with error 32.

Both ordinary file and directory anchors passed. After each attempted attack,
FSCTL_SET_REPARSE_POINT on the containing cwd fails ERROR_DIR_NOT_EMPTY (145).
The native POSIX replacement and deletion attacks are tested independently on
fresh fixtures in `probe-data-anchor.py`, not after deleting the original first.

The opposite metadata-only case is a concrete failure: an anchor opened with
only FILE_READ_ATTRIBUTES|SYNCHRONIZE permits a new DELETE handle even with share
0. FileDispositionInformationEx DELETE|POSIX_SEMANTICS|IGNORE_READONLY succeeds.
Alternatively class65 replaces the held anchor; deleting its replacement succeeds.
The directory then accepts a junction retarget to the outside marker directory.
Retaining that original metadata handle therefore does not pin a named entry.

This corrects the earlier Persistence proposal's overly broad assertion that
FILE_TRAVERSE|FILE_READ_ATTRIBUTES|SYNCHRONIZE with no delete share is a general
Windows directory interlock. Require real FILE_LIST_DIRECTORY/data-read access
(tested here as GENERIC_READ) for directory guards, and FILE_READ_DATA/data-read
access for an anchor. Do not infer interlock strength merely from a share mask.

## Narrow cwd mechanism

Open/bind each accepted path directory with real directory read/list access and
sharing READ|WRITE, excluding DELETE. Retain one actual child of final cwd with
data-read access and no DELETE sharing, or create a unique temporary child
atomically with NtCreateFile(FILE_CREATE), data read/write/DELETE access and
READ|WRITE sharing. Metadata-only child handles are forbidden. Refuse if neither
safe existing-child retention nor exclusive anchor creation is possible.

The anchor keeps final cwd nonempty; the path's child directories keep ancestors
nonempty. Keep all identities/guards and the anchor through CreateProcess.
`probe-chain.py` uses an exclusively created native anchor, proves both final cwd
and its ancestor refuse reparse conversion (145) and independent DELETE opens
(32), and successfully launches a real subprocess at the expected path while
those handles remain held. Cleanup must use exact owned identity and cannot
delete a replaced user entry. The test does not settle user-code effects after
launch or require keeping the anchor for the whole session. No volume/mount,
privileged handle duplication, malicious filesystem/filter-driver, or universal
Windows-version guarantee is established.

## Storage: converted parent is refused without external write

`probe.py` converts an empty pinned parent directory into a junction in place.
Subsequent single-component native operations using that already-held handle as
RootDirectory both fail with STATUS_REPARSE_POINT_NOT_RESOLVED (0xc0000280):

- NtCreateFile using OBJ_DONT_REPARSE|FILE_OPEN_REPARSE_POINT;
- NtSetInformationFile class65 for a separately prepared payload, RootDirectory
  equal to that converted parent, target a simple basename.

The outside directory contains only its original marker; neither destination
name is created there. Thus this tested in-place conversion does not redirect
the proposed native storage operations. The observed behavior is refusal, not
successful access to hidden original-directory contents. Treat that status as
InvalidScope/NotCommitted for a call known not to have published; never retry
with absolute paths, disable no-reparse handling, or claim a committed write.

`probe.py` also includes exploratory metadata-only anchor cases where deletion
preceded rename. Use the independent matrix in `probe-data-anchor.py` for the
actual rename-over conclusion. All calls and status codes are preserved so a
fresh reviewer can inspect that distinction.
