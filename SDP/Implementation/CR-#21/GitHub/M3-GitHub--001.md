# M3 GitHub adapter — Issue #62

State: PARTIAL CHECKPOINT; independent review and integration not requested yet.
Worker branch: codereview-21/layer-github; base412f33e477cec03cb6eafe7b846c9bcdd02c0a25.
Authority: #62/#21, resumed UserRunContract, M3-Adapters--001, frozen1.0.0
Application--GitHub GH1..GH8 and shared BoundaryTypes. Product ownership is solely
internal/github/adapter; legacy parent and all neighboring contracts are unchanged.

First checkpoint supplies repository resolution/registered scope associations,
bounded branch pages with immutable per-request intervals and cursor scope/cap,
strict private JSON/HTTP parsing, and the native noninteractive gh command transport.
Windows suspends before nonbreakaway Job assignment/resume and checks Job active0;
Unix uses a dedicated short-command process group with conservative residuals.
This is not Runtime session ownership or remote rollback evidence.

Evidence: `go test ./internal/github/adapter -count=1 -v`, native Windows/amd64,
Go1.25.0, passes five substantive tests plus its helper. Tests cover exact scope,
duplicate/malformed records, independent survivors, conflict/cap/cursor binding,
private404/rate metadata, actual stdout/stderr separation, output bounds and timed
command cancellation/root reap. Initial fixture startup budget was too short and
the output control exposed an embedded bytes.Buffer ReadFrom bypass; both were
corrected and the complete targeted command rerun successfully. No live PRs or
local Git product mutations were used. Source/test references are in this commit.

Second checkpoint adds all four RemoteFacts methods and CreatePullRequest:
private fork/base/head/unavailable endpoint mapping, exact expectation comparison,
qualified literal POST payload, repeated preflight endpoint checks, all six closed
creation alternatives and independently bounded post-create observation. Known
creation survives nonzero transport, cancellation and failed refresh; no replay,
causal adoption, local push/fork or corrective mutation occurs. Nine controlled
creation scenarios, deleted-fork expectation preservation and concurrent value-copy
controls pass in the same targeted native Windows command. These are raw transport
fixtures, not live publication evidence.

Remaining: broader adverse transport/cancellation controls, full configured platform CI,
independent exact-source review and serial Master integration after accepted Git.
No Slice or baseline finding is closed. This checkpoint is not an acceptance
candidate and carries no completion tag.
