# Commit specialization feasibility for BC Issue #55

Bounded design evidence only; no product/repository/Issue/ref edits. The inspected
BC worktree is `1de3dca25d71b3411f499d83835fce55d3887cca`, accepted-design merge
base `4a42222f7bfedc1d80693effbb25a1a82fcff65e`. Authority read: AGENTS,
developmentInstructions, full #55/#52/#21 and relevant accepted REFDES/API,
Git-Safety superseding publisher and Storage. Existing dirty BC drafts belong to
other authors. All executable fixtures and source downloads are below this new
temporary directory. No production protocol acceptance is claimed.

## Proposed mechanical sequence

1. Bind exact HeadState, full original index bytes/version/identity, observed
   StageAll path/content set when selected, effective commit/conversion/signing
   configuration and hook source versions. Established attached and detached
   commit parent is exactly P; unborn is exactly the selected symbolic branch
   with absence of its ref and no fabricated parent. Use accepted full-index,
   files-backend, no-sequencer supported profile and existing scope/guard policy.

2. Build an operation-owned candidate index natively. Existing-index Commit starts
   from the confirmed complete index; observed StageAll starts from the accepted
   snapshot staging preparation, never a later unconstrained live `git add -A`.
   StageAllAndCommit can sequence the existing protected StageAll publication as
   a separately journaled stage effect before the following Commit substep.
   Then a failed commit truthfully retains the completed stage effect. Do not
   report one inseparable stage+commit success or automatically unstage on failure.

3. Create a private standalone repository with exact source HEAD/branch name,
   independent refs/object-write directory and copied candidate index. Native
   alternates provide read access to pinned real objects. Native ordinary
   `git commit -m <literal>` (no amend/-a/pathspec/allow-empty) remains the message,
   tree, parent and signing engine. Snapshot/map effective supported configuration;
   private automatic maintenance and unneeded rerere are disabled. Do not copy
   real storage paths or silently drop unsupported config. This engine changes no
   actual worktree paths or public branch; it commits only its private HEAD.

4. Private core.hooksPath contains only adapter bridge wrappers for pre-commit,
   prepare-commit-msg and commit-msg, which immediately dispatch native
   `git hook run --ignore-missing <hook> -- <args>` in the ACTUAL worktree root.
   Restore actual GIT_DIR, GIT_COMMON_DIR and GIT_WORK_TREE, original effective
   config/environment policy, native author variables and GIT_EDITOR=:; give
   every hook the same absolute operation-owned GIT_INDEX_FILE. Convert the native
   message-file argument to its absolute owned location, preserving the remaining
   native arguments (for literal -m, prepare source is `message`). Hook Git commands
   thus resolve the real repository/HEAD and write candidate index/object data.
   Pre-hook worktree edits remain actual hook effects; gh-tree performs no later
   worktree reset/checkout/cleanup. Scratch must NOT dispatch the user's
   reference-transaction/post-commit hooks.

   This preserves the documented hook operands, normal cwd/Git context, ordering,
   author environment, hook failure propagation and native message cleanup. It
   does not claim a private Git directory is equivalent to the real one. A hook
   depending on a hardcoded COMMIT_EDITMSG pathname instead of its passed filename,
   ignoring GIT_INDEX_FILE, or other unmapped private-context detail needs explicit
   supported handling/capability refusal, never silent suppression. Arbitrary
   helper side effects cannot be statically sandboxed or universally classified.

5. A configured signing program also needs an adapter-owned transparent launcher:
   execute the ORIGINAL configured program in the actual cwd, restored real Git
   environment, candidate index and preserved stdin/stdout/stderr/exit semantics.
   Git still supplies the final exact commit payload and native format-specific
   arguments; never synthesize signature headers or strip signing configuration.
   Resolve relative program/key paths against their real context. OpenPGP/X.509,
   SSH literal/path keys and SSH defaultKeyCommand need their appropriate explicit
   mappings and native tests. If a profile cannot be reproduced, refuse it before
   target publication. The included fixture proves OpenPGP invocation and native
   embedding, not cryptographic validity or complete signing-provider support.

6. On native success, preserve exact N, expected P/no-parent, final message and
   final candidate index as distinct outputs. IMPORTANT: native commit rereads the
   index after pre-commit, then builds its tree before prepare-commit-msg and
   commit-msg. A later message hook's `git add` can leave a newer staged index that
   is NOT in N. Preserve it; do not force final index tree to equal N's tree.
   Final whole-index bytes/semantic flags need accepted native refresh against
   live files before freezing its publication payload. Transfer/pin N and every
   final candidate index object root; N alone may not retain later staged blobs.
   Retain/report candidate changes on failure too. A policy that publishes valid
   hook-created staging on failed commit must use the same guarded stage substep;
   never publish an unclassified failed native index or silently discard it.

7. Prepare a native transaction through HEAD, not an explicit symref-verify plus
   branch update:

   ```text
   git update-ref --stdin --create-reflog -m <commit reflog reason>
   start
   update HEAD <N> <P-or-format-null>
   prepare
   ```

   For expected detached HEAD add --no-deref. For attached/unborn use normal
   dereferencing. AFTER prepare succeeds, inspect exact actual raw HEAD mode/name
   under its native HEAD.lock and require the expected HeadState. Native prepare
   also holds the selected referent's lock and tests expected P/null. If HEAD
   switched to another branch at the same P before preparation, this transaction
   locks that other branch but the locked symbolic-identity comparison aborts it
   BEFORE index/ref publication. Revalidate scope/config/occupancy as required.
   This profile passed on Git2.43.0.windows.1 and2.48.1.windows.1; it does not need
   the attach-operation symref-update profile. Unknown backends need proof.

8. Acquire real index.lock exclusively/nonblockingly while native ref locks remain
   owned; compare the full original index version. Native reference-transaction
   prepared hooks have already run, so their real index edits also invalidate the
   expected index. Construct the final native image through an alternate path,
   never by a nested command that tries to reuse the owned real lock. Apply the
   accepted journal/capture/revalidate/no-replace publisher to the actual index.
   Retain both the captured old object and installed payload inode; a direct late
   index writer or recreated destination cannot silently lose its bytes. No
   worktree-file publication is needed for Commit. A logically unchanged index
   should avoid unnecessary replacement. Missing initial index uses expected-
   absence/no-replace semantics, never a fabricated old index object.

9. Release the owned index interlock, commit the already prepared native ref
   transaction, and invoke real post-commit exactly once after known public commit
   success with real cwd/environment and the now-real index. The native public
   transaction preserves real reference-transaction hooks; recovery/pin refs are
   separately reported native transactions and may have their own callbacks.
   Prepare-hook rejection occurs before index publication. Committed reference
   callbacks and post-commit occur after index publication; NO index/worktree
   replacement or cleanup follows them. Their resulting edits remain visible.
   Post-hook failure does not erase a known committed N; retain its diagnostic.

10. Reconcile exact actual HEAD/ref/index/status and record separate stage,
    hook/helper, object/recovery, index, ref and callback effects. Cancellation
    before owned publication does not undo real hook side effects. Cancellation
    after index publication but before ref commit leaves a known staged/partial
    effect with original ref; cancellation after known ref commit still reports
    the known new Revision. Lost process/ref result is indeterminate unless actual
    evidence suffices. Keep the journal/candidate/recovery, reap the owned command
    and never automatically reset, unstage or replay a consumed commit.

## Native evidence

`python probe_commit.py` uses only new NTFS fixture repos, isolated Git global/
system config, test author identity and disabled auto maintenance. Results and
native stdout/stderr/facts are in results.json. Source hashes are in hashes.json.

- symref-verify HEAD + branch update in one transaction exits128 in Git2.48.1:
  `multiple updates for 'HEAD' (including one via its referent ...)`.
- Dereferencing `update HEAD` prepared transactions hold both HEAD.lock and the
  actual branch lock; detached holds HEAD.lock; unborn holds HEAD plus absent
  branch lock. Real competing symbolic-ref fails. Exact same-OID wrong-branch
  identity is detected while locked and aborted, both Git versions.
- SHA1/SHA256 native ordinary, initial and detached bridged commits keep real
  pre/message-hook context and edits, invoke post-commit only after publication,
  retain exact parent/no-parent and preserve native cleanup. Hook rejection leaves
  public Head/index unchanged while retaining actual worktree and candidate edits.
- A commit-msg-hook added file remains staged but outside N, proving the native
  distinction above. Native no-empty default rejects an unchanged index.
- OpenPGP fixture sees actual cwd, real Git dir, candidate index, exact native
  `--status-fd=2 -bsau FIXTURE-KEY` and tree/parent/message payload. Git embeds the
  fixture output itself. This is explicitly NOT a valid cryptographic signature.
- Native add is blocked by index.lock; prelock external staging is preserved and
  rejected; direct last-check index writes are captured and restored exclusively;
  recreated index prevents replacement. Cancellation is injected before index,
  after index and after ref, with distinct verified effects and retained originals.

Primary source anchors (downloaded tagged source retained alongside the probe):

- [commit.c v2.48.1](https://github.com/git/git/blob/v2.48.1/builtin/commit.c):
  determine_author_info619–673, prehook759, index reread/tree1080–1091,
  prepare/message hooks1093–1111, cleanup1838, commit construction1872–1875,
  public update1879, postcommit1900.
- [files-backend.c v2.48.1](https://github.com/git/git/blob/v2.48.1/refs/files-backend.c):
  split_head_update2384–2430 and split_symref_update2446–2495 explain the rejected
  explicit HEAD composition and the native HEAD-first split/lock path.
- [gpg-interface.c v2.48.1](https://github.com/git/git/blob/v2.48.1/gpg-interface.c):
  native OpenPGP/X.509 and SSH invocation973–1115.
- [Documented hooks](https://git-scm.com/docs/githooks) and
  [native hook runner](https://git-scm.com/docs/git-hook).

Limits: this is an ordinary-file Python feasibility model, not a production
publisher, process-lifecycle implementation or crash/permission/path proof. Full
index flags/extensions, linked-worktree path binding, Unicode/path corner cases,
hooks/config drift, legitimate real signing providers, absent index publication,
actual killed/descendant/crash states and Windows/Linux/macOS product-native tests
remain implementation gates. Final source/hash review must independently reopen
actual artifacts. The first run's negative symref case hit a harness broken-pipe
cleanup error after Git refused; results-first-run.json preserves that failed
harness result. The correction captures the one-shot native failure directly.
