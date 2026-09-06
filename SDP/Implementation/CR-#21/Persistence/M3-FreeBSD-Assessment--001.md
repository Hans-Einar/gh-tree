# M3 FreeBSD metadata feasibility assessment — #63 / #21

State: RESERVED for Master disposition; assessment complete, no product acceptance.
Role: fresh independent native-profile assessor, 2026-09-07.
Canonical authority: `c7eef51efc1a387da512bc7f87a31f5906011597`, ledger76,
Sprint-004-v04 / I-03 / M3. Immutable product inspected with `git show`:
`6fc6c2c16f0bf53bcddbb0234a4f3c57ed110f12`, chiefly
`internal/persistence/{metadata_freebsd,metadata_unix}.go` and its README/report.
No authoring worktree, product, workflow, module, frozen contract or user-system
change. This report is uncommitted for Master preservation. The blocked Git review
was neither retried nor substituted, and no withheld output was retrieved.

## Conclusion

An ordinary unprivileged process on the observed FreeBSD 15/ZFS profile cannot
positively enumerate, prove absent, or copy arbitrary protected system extended
attributes through the selected Extattr interfaces. Native ACL access does work
independently and can support an ordinary-user ACL implementation, but does not
resolve the generic system-metadata visibility gap. No examined native replacement
or block-copy mechanism transfers that inaccessible metadata onto an independent
prepared payload. Therefore the unchanged full detection/preservation requirement
has no demonstrated successful ordinary ZFS write profile. Refusing before
publication is truthful safety behavior, but is not completion of FreeBSD support,
#63, any full Slice, or v0.4.

This is a conditional technical contradiction: arbitrary system metadata must be
detected/preserved, the process lacks its required privilege, the input may be an
existing ordinary file with that metadata, and publication replaces it with an
independent inode. It is not a claim that every possible FreeBSD filesystem or
future kernel mechanism is impossible. No alternative filesystem/profile has
been proved here; switching to one is not an accepted product decision.

## Authority and existing native evidence

Read root AGENTS/developmentInstructions, full #63 and parent #21 authority,
FROZEN Application--Persistence 1.0.0, accepted Storage--001 and
Verification--001, selected Persistence feasibility protocol, local README and
existing CI review/evidence. The boundary requires positive ACL/xattr inspection,
copy/verification or unchanged refusal, retained original plus separate raw backup,
and prepared same-directory Renameat/Linkat publication. The original inode's
retention explicitly does not claim preservation on the new payload.

- [Run34056731199/job101549884000](https://github.com/Hans-Einar/gh-tree/actions/runs/34056731199/job/101549884000),
  attempt1, source `489a731557ca699992386a0fc07fcbe57e792070`: actual
  FreeBSD15.0-RELEASE-p12 GENERIC amd64, kernel `7b527b9b97ba`, ZFS zroot/home,
  uid/gid1001 ghci, Go1.25.0 CGO0. Reopened existing local artifact: user list
  exit0, system list exit1/EPERM, getfacl exit0 showing owner/group/everyone
  NFSv4 entries. This was infrastructure/M2 evidence, with Persistence absent.
  `native.log` SHA256 `d68928b96ec5668c9067bc10b27c1362b0bbecdb69cb536be0a843d0bbae14bd`;
  retained location `C:/Users/hanse/.codex/tmp/gh-tree-freebsd-review/artifact/`.
- [Run34063842833/job101569041029](https://github.com/Hans-Einar/gh-tree/actions/runs/34063842833/job/101569041029),
  attempt1, actual adapter source `6fc6c2c16f0bf53bcddbb0234a4f3c57ed110f12`:
  same native OS/user/filesystem profile; job FAILURE. Reopened run metadata and
  downloaded existing artifact9998337208 (61,945 bytes; archive digest
  `93bbab10dd839da9651aec7df386bb38e539fbc5d83cbaee31fa901ee4a234f6`).
  Actual positive failures include TestUnixNativeMetadataCopy,
  TestPreparationExclusiveCreationAndNativeMetadata and public commit/recovery
  cases reporting `unsupported native BSD file flags`. This is the first observed
  refusal; the log does not identify the actual flag value. Do not relabel those
  failures as observed ACL or extattr syscall failures. Further barriers below
  follow from independently inspected source. `tests.jsonl` SHA256
  `d774a431cb59b29774bf44ece935c946cc3c56c5d2bfd530490d9f362fd7e04e`,
  `native.log` SHA256 `cc88e434f4fdb30296772498565b2542ee3f42e528a30c4a90b4519f63dee1d8`;
  local directory `C:/Users/hanse/.codex/tmp/gh-tree-freebsd-assessment-34063842833/`.

No new VM, native probe, privilege experiment or product test was run. Cross-build
success is separate from these failed native positives.

## What FreeBSD actually permits

Primary source inspected at current official releng/15.0 commit
`af58d0db156a4036624d7f8a0bdd4b739a5418d2`; the essential privilege check and ZFS
ACL-type behavior were also checked at the actual guest kernel ref `7b527b9b97ba`.

1. **Direct namespace access:** `extattr_check_cred` requires
   `PRIV_VFS_EXTATTR_SYSTEM` for system namespace; user namespace instead checks
   vnode access. ZFS list/get/set invokes it before attribute lookup. Even a
   size-only system list is denied before establishing whether anything exists.
   Ownership, mode0600, a newly created file, or successful user listing cannot
   turn EPERM into absence. The kernel's internal NOCRED allowance is not a
   userspace credential option. [Kernel credential rule](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/sys/kern/vfs_subr.c#L5716),
   [ZFS list/get/set](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/sys/contrib/openzfs/module/os/freebsd/zfs/zfs_vnops_os.c#L6569),
   [extattr(9)](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/share/man/man9/extattr.9).
2. **ACL access is different:** ZFS native getacl accepts ACL_TYPE_NFS4 and
   rejects a POSIX access-ACL request with EINVAL. ACL retrieval/copy checks
   ACE_READ_ACL/ACE_WRITE_ACL through the ACL machinery, not blanket system-EA
   enumeration privilege. The observed getfacl success is consistent with this.
   Current adapter code checks flags first, then POSIX before NFS4, propagates
   EINVAL, and refuses every successful NFS4 ACL. Thus correcting flags alone
   cannot enable it. A supported implementation can select the actual ACL model,
   validate complete ACL semantics and copy/requery on its owned payload. It must
   not suppress arbitrary EINVAL or equate three NFS4 entries with mode-only
   semantics. Native acl_is_trivial_np defines model-specific triviality.
   [ZFS ACL conversion](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/sys/contrib/openzfs/module/os/freebsd/zfs/zfs_vnops_os.c#L6615),
   [ACL authorization](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/sys/contrib/openzfs/module/os/freebsd/zfs/zfs_acl.c#L1792),
   [libc descriptor model selection](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/lib/libc/posix1e/acl_get.c),
   [triviality definition](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/lib/libc/posix1e/acl_is_trivial_np.3).
3. **Generic system metadata is not the ACL:** successful ACL retrieval says
   nothing about other protected names or policy labels. The two extattr
   namespaces have independent names/values. A finite list of known ACL/label
   queries cannot prove arbitrary system metadata absent. Disabling xattr access
   is not evidence that preexisting stored values vanished; no such profile is
   adopted. BSD stat flags are another separate, observable surface: current
   blanket nonzero refusal needs an explicit supported-flags decision and native
   evidence. The present logs do not justify guessing which flag failed.
   [extattr(2)](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/lib/libsys/extattr_get_file.2),
   [flag definitions](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/sys/sys/stat.h).
4. **Native preservation mechanisms:** descriptor ACL and user-EA copies can
   preserve metadata they can actually read/write, with verification. Renameat
   publishes the prepared inode; it does not merge the replaced inode's metadata.
   A hardlink preserves the original inode and its metadata, but writing the new
   document through that alias would mutate the retained original and violate
   old/new visibility. FreeBSD copy_file_range, including COPY_FILE_RANGE_CLONE
   and ZFS block cloning, copies a byte range between existing files; it is not
   an opaque complete security/EA clone. None supplies the missing generic
   system-metadata transfer. [rename(2)](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/lib/libsys/rename.2),
   [copy_file_range(2)](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/lib/libsys/copy_file_range.2),
   [ZFS implementation](https://github.com/freebsd/freebsd-src/blob/af58d0db156a4036624d7f8a0bdd4b739a5418d2/sys/contrib/openzfs/module/os/freebsd/zfs/zfs_vnops_os.c#L6904).

## Exact next permitted action

Master must resolve this contradiction before accepting an ordinary FreeBSD write
profile. To retain the present requirement, a concrete supported mechanism must
establish/transfer otherwise inaccessible metadata onto an independent payload,
with unchanged descriptor scope, native publication and recovery guarantees.
No such mechanism is established here. Adding a privileged helper would change
the execution/trust model and requires explicit user/contract authority; it is
not an implementation shortcut.

Alternatively, an explicit reviewed BC-CHANGE must define the exact FreeBSD
preservation set, treatment of unreadable system metadata, and a positively
established supported environment/profile with honest limitations. Any reduction
from current metadata requirements or ordinary-platform capability requires the
user's concrete scope decision, followed by affected design/BC review/refreeze
and native verification. This report chooses none of those reductions.

ACL/model selection and supported ordinary flags are technically addressable
implementation work, but cannot independently close the protected-namespace
gap. When a real solution/approved contract is ready, use Master's existing
owned CI route for positive public commits, existing files, ACL/user-EA copy and
refusal controls, flags, native barriers and recovery. Add exact flags/model
diagnostics to that meaningful run; do not schedule another VM solely to repeat
the already conclusive privilege probe. All intended platforms, full v0.4 scope,
independent review and integration/release gates remain open and unchanged.

## Master disposition

Assessment evidence is accepted as a scoped feasibility finding. No implementation,
privilege expansion or capability reduction is approved. A user scope decision is
required before changing the ordinary FreeBSD trust/preservation profile; any
selected change then needs concrete design, affected independent BC review and
native verification. Windows identity correction and other independent M3 work
continue. The full original requirement remains open until explicitly resolved.
