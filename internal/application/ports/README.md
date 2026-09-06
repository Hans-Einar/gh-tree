# Application dependency interfaces

Issue #59 implements the seven exact FROZEN 1.0.0 interfaces: GitFacts (11),
GitMutations (12), RemoteFacts (4), RemoteMutations (1), LaunchDiscovery (2),
Storage (6), and Sessions (12), totaling 48 methods. API owns semantic data;
ports owns interfaces and private preparation/load/commit/receipt wrappers.
SourceVersion, StorageVersion and RuntimeEventCursor alias their API-owned types.
API never imports ports. No concrete adapter DTO, native object, callback,
reader, channel or dispatcher enters these surfaces.

Storage load wrappers retain complete typed documents only for usable load
observations and enforce family/run-root binding. Commit wrappers require a
whole current-schema document and expected typed version, including absence.
Returned observations and documents are immutable API values.

PlanIssuer takes an adapter-generated lifetime; IssueCreate/Retarget/Stage/Commit/
Restore/Stash/Branch/Push produce eight private plan alternatives. PlanSpec supplies
operation, version, unpredictable token, group, step, role, immutable summary and
summary digest. Origin binds a derived child to its sequence root. Exact switches
reject nil, typed nil and foreign embedded plans. Roots are not executable;
continuation kind/group/step and supplied before/after receipt versions remain
distinct. Summaries and receipt version getters retain independent copies.
PlanSummaryDigest returns the exact adapter-issued digest by value through the
same closed identity validation. An opaque-plan consumer can supply it to
ApprovalIssuer without decoding versions or inventing another hash convention;
observing the digest grants no approval or native execution authority.

ApprovalIssuer takes a coordinator-generated lifetime. Issue binds the original
root/digest/allowed choice and a required ConfirmationID when the plan requires
one. ValidFor checks portable original/root-child binding. GitPrepareContext
retains the original operation/root and only accepts a verified, uncanceled
matching first-step predecessor. ExecutedGitMutation preserves API facts plus an
optional matching receipt even when the port returns a nonnil error. Failed,
conflicted, indeterminate and cancellation facts cannot become successful
continuation evidence through the wrapper.

These constructors allocate no entropy or monotonic IDs and own no native
registry/resource. They cannot prove that supplied tokens are unpredictable,
that an issuer is registered, that a confirmation was consumed, or that a native
step succeeded. Git must perform bounded atomic registry admission, release,
replay prevention, own-step reconciliation and native revalidation; Application
must consume approval and sequence the frozen workflows. Portable metadata
validation does not grant user permission or replace those later gates.

Contextless ReleasePlan/AckEvents, the Prepare triple returns, exact specific plan
types, Runtime direct SessionID calls and aggregate Shutdown result are retained.
External conformance fakes prove all signatures compile. They intentionally
implement no behavior and do not constitute workflow or native acceptance.

M259 corrections require complete operation-specific summaries before issuance,
match sequence-root roles to declared compound intent, and retain the original
subject/target in derived summaries. Executed receipts bind every supplied result
subject as well as operation/kind/version. Storage loads validate all represented
versions and retained recovery records, including documentless failures. These
checks reject explicit contradictions; native registry and publication proof
remain the implementing adapter's duty.
