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

Remaining: PR facts/endpoint mapping, exact pre/post creation and all six outcomes,
broader adverse transport/cancellation controls, full configured platform CI,
independent exact-source review and serial Master integration after accepted Git.
No Slice or baseline finding is closed. This checkpoint is not an acceptance
candidate and carries no completion tag.
