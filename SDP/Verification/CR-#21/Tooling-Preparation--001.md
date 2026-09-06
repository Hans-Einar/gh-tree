# Native Git signing tool preparation

State: bounded tool baseline PASS; no v0.4 product verification or M3 start.
Authority: #21 user verification contract and accepted Git BC G8a/Verification--001.
Observed canonical checkpoint: 04eb90a54fed3acdc6f87614af7f919ffedb167e.
The Master/Verifier executed only fresh owned temporary repositories and identities.
No gh-tree adapter, private commit bridge, user repository/configuration, installed
extension, public remote, tag or release was changed. M2 #59 corrections continue.

Native Windows 11 build 26200 amd64, Git 2.48.1.windows.1, GnuPG 2.2.41-unknown,
Windows OpenSSH ssh-keygen 9.5.6.1 and Python 3.14.0 were used. For SHA-1 and SHA-256,
native Git produced actual Ed25519 SSH and OpenPGP signed commits. All four
`git verify-commit --raw` checks returned 0. Four objects with altered messages
retained their original signatures and were rejected with exit 1. This establishes
that these local tools can supply positive and adverse cryptographic controls for
future adapter tests. It does not verify G8a context mapping/publication or support
X.509, SSH agent/defaultKeyCommand, alternate key/program contexts or other systems.

All original attempts are retained in the [evidence archive](Evidence/Tooling-Preparation--001/ArchiveManifest.json):

- Attempt 01 failed before signature cases: a Windows-form long GNUPGHOME caused
  the bundled MSYS GPG agent connection to report invalid IPC value.
- Attempt 02 used a shorter Windows-form home and still reported IPC connect
  failure. Shortening the path alone did not fix this environment.
- Attempt 03 used an MSYS `/c/...` home. Both SHA-1 profiles passed; SHA-256 SSH
  verification also returned 0, then the probe incorrectly demanded `gpgsig`.
  The captured native SHA-256 object has `gpgsig-sha256`; this was a probe defect.
- Attempt 04 corrected the format-specific header assertion in a fresh fixture.
  All four valid signatures and all four tamper controls passed. The scoped
  temporary GPG agent kill returned 0 after every attempt.

The archive preserves only each script, command log, results and original manifest,
with exact byte/hash checks. Temporary private keys and repository object stores
remain outside the repository. No signing identity or key is a user credential.
The scripts clear Git/GPG/SSH context, isolate global/system config and create local
fixtures; they are inert `.txt` evidence here. Native program paths are recorded
for reproducibility, not a developer-local dependency in product tests.

M1/Domain remain the latest integrated product contributions. API/ports #59 needs
M259-M01..M08 correction/re-review/current CI; accepted viewmodel d279c3f waits for
API-first integration. No CurrentMatrix check, full Slice or baseline finding is
closed by this preparation. Upcoming M3 Git must implement/review actual real
signing bridge behavior and the complete accepted native contract independently.
