# M3 GitHub adapter — Issue #62

State: CANDIDATE; configured source CI and separate independent review pending.
Worker branch: codereview-21/layer-github; base412f33e477cec03cb6eafe7b846c9bcdd02c0a25.
Authority: #62/#21, resumed UserRunContract, M3-Adapters--001, frozen1.0.0
Application--GitHub GH1..GH8 and shared BoundaryTypes. Product ownership is solely
internal/github/adapter; legacy parent and all neighboring contracts are unchanged.

Checkpoint0a45c563 supplies repository resolution/registered scope associations,
bounded branch pages with immutable per-request intervals and cursor scope/cap,
strict private JSON/HTTP parsing, and the native noninteractive gh command transport.
Windows suspends before nonbreakaway Job assignment/resume and checks Job active0;
The final Unix profile uses root-only termination and a dedicated group solely
for conservative residual observation; it never signals numeric PGIDs.
This is not Runtime session ownership or remote rollback evidence.

Evidence: `go test ./internal/github/adapter -count=1 -v`, native Windows/amd64,
Go1.25.0, passes five substantive tests plus its helper. Tests cover exact scope,
duplicate/malformed records, independent survivors, conflict/cap/cursor binding,
private404/rate metadata, actual stdout/stderr separation, output bounds and timed
command cancellation/root reap. Initial fixture startup budget was too short and
the output control exposed an embedded bytes.Buffer ReadFrom bypass; both were
corrected and the complete targeted command rerun successfully. No live PRs or
local Git product mutations were used. Source/test references are in this commit.

Checkpoint235b32a2 adds all four RemoteFacts methods and CreatePullRequest:
private fork/base/head/unavailable endpoint mapping, exact expectation comparison,
qualified literal POST payload, repeated preflight endpoint checks, all six closed
creation alternatives and independently bounded post-create observation. Known
creation survives nonzero transport, cancellation and failed refresh; no replay,
causal adoption, local push/fork or corrective mutation occurs. Nine controlled
creation scenarios, deleted-fork expectation preservation and concurrent value-copy
controls pass in the same targeted native Windows command. These are raw transport
fixtures, not live publication evidence.

Final candidate strengthens bounded wait/drain, Windows opened-thread identity,
pre-resume failed assignment and immediate descendant controls, actual literal stdin,
101-row pagination, malformed/duplicate/foreign-link/provider-format/cancel/partial
creation controls and bounded cursor digests. Local Windows/amd64 targeted and
race commands pass. A full local test/vet/build/twelve-selection architecture pass
also ran during hardening; final source is gated by the configured full CI matrix.
An actual read-only gh observation passed for Hans-Einar/gh-tree (55 branches,
31 PRs, Complete, plus exact ObservePullRequest). No live mutation was performed.

| Clauses / checks | Actual source and proof |
|---|---|
| GH1/GH2/GH7; V-GH-01; V-COMP-03 | adapter.go/create.go conformance declarations, owned package imports; twelve-selection architecture check; unchanged legacy/shared files |
| GH3/GH4; V-GH-01/02 | parsing.go/facts.go/paging.go/pullrequests.go; facts_test.go/adversity_test.go/create_test.go exact scopes, forks/nulls, full OIDs/provider profile, pages/intervals/copies and stale expectations |
| GH5; V-GH-03 | create.go; TestCreateSixAlternativesAndFailurePreservation (nine paths), truncated/admission controls, qualified literal payload and actual native stdin preservation |
| GH6; V-GH-02/03/SLC-12 | transport.go/command_windows.go/command_unix.go; TestNativeCommandStreamsLimitsAndCancel, TestNativeDescendantPipeJoinAndRepeatedResources, Windows invalid-assignment/opened-thread/ABI controls |
| GH8 | complete local targeted/race proof and configured exact-source CI, followed by separate independent actual-source review; no author summary is acceptance evidence |

Remaining: exact-source CI, independent review/corrections and serial Master
integration after accepted Git. SLC-01/02/03/05/08/12/13 contributions do not close
any Slice or baseline finding. Full Application/State/View/Composition and release
proof remains later gates. No completion tag, product PR or integration is claimed.

## Bounded M3GH-R01 correction

Independent review of926b2b70900fa5f0c2277b086fda0d9b6aad1194 required one P2
correction: globally rejecting invalid UTF-8 discarded valid neighboring records.
`decodeList` now validates array framing into unchanged RawMessage bytes;
`strictJSON` still rejects invalid UTF-8 before decoding each record. Valid branch
and PR siblings survive with Unknown completeness and malformed-record diagnostics.
Broken envelopes still refuse entirely; no replacement-character coercion occurs.

`record_utf8_test.go` contains both raw0xff sibling regressions, exact surviving
Unicode/identity/OID assertions, raw-byte preservation and broken-envelope controls.
Native Windows/amd64 Go1.25.0 package test, race and vet pass. The reviewer's existing
independent overlay also passes all four test groups, including the previously
failing branch/PR cases and unchanged creation/Unicode controls. Only parsing,
these regressions and this worker report changed. Native lifecycle source and the
reviewer report are unchanged. Corrected-source CI and independent bounded re-review
remain required; the author does not self-accept the finding or integration.
