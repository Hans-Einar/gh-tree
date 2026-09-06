# Bounded design probe evidence

These are exact source/log snapshots supporting the #52 feasibility appendices.
They are not product code, production tests or implementation authority. Source
suffixes deliberately end in `.txt` so the product Go package inventory cannot
accidentally compile a disposable probe. Preserve bytes with the local attributes
file; recorded hashes refer to these exact bytes, not platform newline conversion.

For reproduction, copy a source into a new temporary directory outside product
worktrees, remove its final `.txt` suffix, use the dependency versions recorded in
Runtime.md, and run only its self-created disposable fixtures. The Git scripts
create new repositories beneath their own directory and isolate global/system
Git config. The scratch script's two symref tests require an isolated MinGit
2.48.1 at the relative location stated in that script; when absent those two
cases are omitted, so a sixteen-case run is not the recorded eighteen-case run.
Do not run these models against a user repository or copy their incomplete
publication/protocol code wholesale into production.

Runtime Linux evidence uses native Linux execution through existing WSL, finally
as UID65534. Its census implementation is Linux-specific; successful compilation
for nine Unix targets does not prove Darwin/FreeBSD execution. Windows logs name
native amd64 and WOW64 386 runs. The complete appendices state omitted resource,
protocol, filesystem, configuration and crash cases and required later proof.

| Source snapshot | SHA-256 |
|---|---|
| git-initial-probe.py.txt | 99BBBC22F3BEEEF65A7073DDF30D67535CCAF443066D170DEEC6CCE2A3514411 |
| git-scratch-probe.py.txt | E326EE32342DDE76E083DAEE210792D5E55AF3F4050E5D8FBD5CD947DAD66BE3 |
| runtime-windows-probe.go.txt | 811BD94B105705762C1BFD25CBCC935EC3BD5417A2D02A1798AC151ACF16C7AB |
| runtime-unix-probe.go.txt | 0451B733D4D0429D73355BBEEC7F779B37A7F32B96ECBDA8553DDF65E5603503 |

Git result records retain their original temporary absolute fixture locators as
provenance. Reproduction naturally creates different paths/OIDs/PIDs and result
hashes. The eighteen followup results and source are preserved here so a future
reviewer can inspect the evidence without relying on ephemeral chat or temp files.

The correction pass adds Persistence sources/modules/logs and its original proposal.
`persistence-source-hashes.json` retains original names; corresponding archive names
are prefixed `persistence-` (probe_windows_test.go -> persistence-windows-probe.go.txt,
probe_linux.py -> persistence-linux-probe.py.txt, go.mod/go.sum ->
persistence-probe-go.mod.txt/persistence-probe-go.sum.txt, Persistence-proposed.md ->
persistence-proposal.md.txt). Windows/Linux result names carry the same prefix.
All seven copies matched their recorded hashes. Cwd probe source hashes and native
observations are in CwdAcquisition--001; these remain bounded mechanism experiments.
