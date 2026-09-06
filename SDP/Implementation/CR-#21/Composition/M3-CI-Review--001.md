# M3 FreeBSD CI independent review — #66

Disposition: **ACCEPT — bounded native CI infrastructure; no blocking findings.**
Role: fresh independent Composition workflow/security/correctness Reviewer.
Sprint-004-v04 / I-03 / shared M3 prerequisite under #21, #63 and #65.

Reviewed clean, pushed source `489a731557ca699992386a0fc07fcbe57e792070`
on `codex/cr21-freebsd-ci`, isolated base
`22217f9c31b33ab78f030a8dd36bd3624bcf7485`. Reviewed tree:
`b5c424db36e91db42e4e8b9d5affd7b7b4d7531e`.
This report is a later documentation-only addition; it does not change that
executable source or claim another native run. M3-CI--001 remains the author's
initial pre-execution snapshot; this record supplies the subsequent review and
observed execution disposition.

## Scope and authority

Read root AGENTS/developmentInstructions, Issue #66 and relevant complete #21,
#63/#65 authority, resumed UserRunContract policy, M3-Adapters--001, existing M1
CI/helper interface, normative Verification--001 and relevant frozen Persistence,
Runtime and native-helper obligations. No new contract interpretation or waiver.
Inspected actual four-file base-to-candidate diff: appended 57-line CI job,
freebsd-native.sh, freebsd-tests.sh and author's report. No product, legacy,
module, architecture guard, release or frozen-contract edit. Existing workflow
jobs and exact canonical Windows/amd64 Go1.25.0 helpergen `-check` invocation
are unchanged. No layer-local instructions apply to the new report.

## Independent checks and findings

- Workflow uses one ubuntu-24.04 disposable host, contents:read and no publication
  operation. Checkout `fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09`, FreeBSD action
  `f0552d3b69211736abd97f02ff3d4674c56b73b1` and artifact action
  `ea165f8d65b6e75b540449e92b4886f43607fa02` are full revision pins.
  Checkout credentials are not persisted; no secret or arbitrary envs input is
  supplied. No pull_request_target execution or cache-sharing path is introduced.
- Guest setup checks FreeBSD 15.0/amd64, creates ghci, and copies only the checkout
  into its private home. All Go operations run as this ordinary user through
  `env -i`, named local toolchain/module settings and private temporary storage.
  Native host/target/toolchain, exact Git HEAD and clean tracked/untracked source
  are asserted; source and temporary filesystems must be UFS or ZFS. A second
  clean-source assertion follows the suite. This is trustworthy-runner source
  identity evidence, not attestation against a malicious VM image.
- Five required M2 roots are enumerated with `go list`. Every present M3 root
  (Git, GitHub adapter, Persistence, Discovery, Runtime) is recursively enumerated;
  a present root with no selected package fails. Missing roots print NOT RUN.
  All resulting native-selected packages go into the same uncached Go suite;
  no test-name filter or silent per-package failure catch is present. Build-tag
  excluded platforms, absent adapters and later Application/State/View work are
  not thereby verified. Legacy tests remain outside this declared native scope.
- Inspected the shell error path and the pinned action's SSH runner/final catch:
  Go nonzero -> run_tests return -> guest shell -> su status -> root script exit
  -> SSH rejection -> core.setFailed/process.exit(1). The real failing control
  confirms the Go/helper/su portion; action-level failure propagation is supported
  by source inspection, not falsely described as an observed failed whole job.
- JSONL keeps complete suite output, including skips; human logs expose terminal
  pass/fail/skip events. Ten-minute per-package Go timeout and 25-minute job bound
  constrain execution. Always-run retrieval copies only nine named evidence files,
  records missing evidence explicitly, checks host rules/listeners and uploads a
  seven-day artifact. Named scope and time bounds are not a strict log-byte cap;
  a killed runner/job can prevent final evidence retrieval. Neither condition
  converts a failed suite into success.
- Independently reran both Git Bash syntax checks and base-to-source
  `git diff --check`: PASS. Inspected appended-only workflow diff and existing
  job/helper interface directly. Author actionlint result is corroborative only;
  no independent actionlint rerun or duplicate full CI was needed. No product
  test modification or fabricated local FreeBSD execution was performed.

## Actual dependency and network boundary

Independently fetched pinned upstream index.js and compared it with the existing
Checkpoint-20260906-1725Z/freebsd-source archive (normalized line endings/trailing
archive whitespace): identical. Read the actual action inputs/config selection,
HTTPS downloader, sudo apt setup, AnyVM invocation, SSH environment/exit handling,
workspace synchronization, prepare/run, cache/copyback gates and final failure
handler. The actual run selects AnyVM0.6.5 and FreeBSD builder2.2.6.

Independently downloaded current AnyVM0.6.5 source without executing it locally;
SHA256 `24fbcf739fc07d7fd655f493481cb77ed022b4baa81b039227080e1a7aaaa921`.
Inspected image/key/profile acquisition and release-search fallback, QEMU address
selection/hostfwd, monitor/serial binding and conditional reverse SSH authorization.
This matches the author's observed source hash; the CI log records its download
URL but no runtime content hash, so byte-identical controller execution is not
claimed. AnyVM Python, builder image/key/profile release assets, host apt and guest
pkg remain mutable trusted dependencies. Action revision pinning does not freeze
those assets or authenticate guest-reported identity. The Go tarball alone has
an enforced content hash, independently matched to the official go.dev manifest:
`86e6fe0a29698d7601c4442052dac48bd58d532c51cccb8f1917df648138730b`.

The builder-provided SSH key is publicly downloadable. AnyVM disables host-key
checking and binds guest SSH to loopback plus detected host private addresses.
The workflow installs first-position IPv4/IPv6 INPUT denial of NEW non-loopback
TCP before the VM starts and checks both rules afterward. Actual evidence shows
QEMU SSH on 127.0.0.1, 10.1.0.159 and 172.17.0.1 port10022; monitor4444 and serial7000
are loopback. The pre-existing host SSH service listens on wildcard port22 and is
also covered by those ingress rules. No VNC listener is shown; actual invocation
has `--vnc off`, no public/nat/debug/accept-vm-ssh options. Reverse authorized_keys
writes and guest host-alias injection require sshfs or accept-vm-ssh; neither is
selected. These are source/rule/listener observations, not an external packet
penetration test or proof of isolation from a compromised dependency.

Actual rsync copies the runner's `~/work/` subtree, excluding _actions and
_PipelineMapping, onto guest disk; this can include runner temporary workflow
files. The ordinary-user source copy is restricted to GITHUB_WORKSPACE. The action
implicitly forwards GITHUB_* and CI through SSH, not the whole host environment;
this job supplies no secret and tests clear the inherited metadata environment.
Copyback/cache/prepared-cache are disabled. The controller still executes with
host Actions privileges and available sudo. Guest code/controller and dependency
maintainers remain inside the CI trust boundary: the firewall limits off-runner
ingress, not malicious guest-to-host access or hostile host control code. This
bounded dependency choice is acceptable for this disposable, read-only CI job;
it is not approval for secrets, deployment authority or persistent self-hosted use.

## Independently observed run and artifact

Official [run34056731199](https://github.com/Hans-Einar/gh-tree/actions/runs/34056731199),
attempt1, event push, head489a731557ca699992386a0fc07fcbe57e792070: completed SUCCESS.
All19 applicable jobs succeeded, including the unchanged18; the sole expected
pre-Runtime helper job is explicitly SKIPPED.
[Native job101549884000](https://github.com/Hans-Einar/gh-tree/actions/runs/34056731199/job/101549884000)
actually ran 2026-09-06T20:02:53Z..20:05:00Z. Independently queried official run/job
metadata and downloaded artifact9996216974, named
`freebsd-native-489a731557ca699992386a0fc07fcbe57e792070-1`, 44,330 archived bytes;
GitHub-reported archive digest
`sha256:d5d6c49cc8e930c7792a119cc6074a63b64fc4410f986e6ec3394fdef245ca6d`.

Observed kernel/userland15.0-RELEASE-p12, GENERIC amd64; uid/gid1001 ghci with no
wheel membership; Go1.25.0 native freebsd/amd64, CGO0/GOTOOLCHAINlocal; Git/git-lite
2.55.0 and ca_root_nss3.128. Source SHA/tree match the reviewed objects. Both source
and temporary storage are zroot/home ZFS, ghci-owned private directories.
Five packages PASS: application/api, application/ports, composition/architecture,
domain, tuistate/viewmodel. JSONL contains494 test/subtest pass events, zero skip
and zero fail events; suite exit0. All five M3 roots are absent and explicitly
NOT RUN. Separate expected failure JSONL records TestExpectedFailure/package fail,
marker CI_EXPECTED_FAILURE_CONTROL and helper exit1; parent log records its
validated failure control PASS.

Ordinary-user user-xattr enumeration exit0; system-xattr enumeration exit1 with
Operation not permitted; native getfacl exit0 reports the actual NFSv4 ACL.
No system-xattr absence, metadata preservation, native publication/census/PTY or
V-PER/V-RUN acceptance is inferred. Persistence must handle/prove this actual
profile or refuse it under its unchanged contract. UFS is allowed but not executed
in this run. Infrastructure availability is the only gate accepted here.

Reused full job log SHA256:
`12611d19c016a77e43f5e37502afc417a3d49f76705d7560758bf88e1a140c0b`.
Independently downloaded artifact file SHA256s:

| File | SHA256 |
|---|---|
| host.txt | 8cb16145bf06653cdafc5c43978133db8fc72382659683ea0f338d78292de320 |
| native.log | d68928b96ec5668c9067bc10b27c1362b0bbecdb69cb536be0a843d0bbae14bd |
| tests.jsonl | 19f63d1426b99f30790a547622e203056c7314830355bfb0c68fa2eae2598f8c |
| failure-control.jsonl | 54afa3320711970e03f3d41c57c9b66e2efee088353c5935149d8afafec9d2f4 |
| packages.txt | e1d53c45f75c51ec4d5330c1a06927529800b9011d1e554c4853a8c17f2e7e23 |

Local audit downloads remain in C:/Users/hanse/.codex/tmp/gh-tree-freebsd-review;
existing action archive and official artifact are reused rather than duplicated.
Seven-day artifact expiration remains explicit; Master may retain the concise
needed evidence with its normal integration checkpoint.

## Handoff

No blocking or nonblocking implementation finding. No product/workflow correction,
BC change, gate waiver or full Slice/baseline-finding closure. Reviewer adds only
this UTF-8 report, commits/pushes with [skip ci], and hands the resulting report
HEAD to Master. Exact next permitted action: Master integrates this accepted
infrastructure source plus review, records the actual integration SHA and checks
its applicable CI. Future adapter native acceptance must execute their actual
integrated packages/fixtures and inspect skips/profile failures; this M2-only
run cannot substitute. Full #21/native/vertical/release gates remain open.
