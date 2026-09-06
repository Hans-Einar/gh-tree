# Persistence

Issue #63 implements the FROZEN Application--Persistence 1.0.0 boundary in
isolated milestones. The current milestone supplies private schema/document/
version primitives. The six-method native Storage adapter is still pending;
this package is not connected to the product entry point.

The codec maps all specified schema0/schema1 user configuration, preferences,
and run configuration fields to the accepted immutable API documents. Only
`stripPrefixes` accepts null. Missing fields and explicit empty lists/strings
remain distinct. Loads retain schema0 without migrating; encoding requires an
Application-proposed schema1 document. Empty document candidates preserve absent
fields for Application's default policy. No codec selects active worktree,
default alias, migration winner, provider executable or repository association.

Every input first passes the bounded strict JSON predicate: 4 MiB, depth64,
valid UTF-8 and surrogate escapes, unique decoded member names at every depth,
one object envelope and a supported integer schema. Forward integer text is
retained without overflow into uint32. All known shape/presence checks use the
API constructors, and final encoded bytes pass the exact byte/depth limits again.

Unknown values retain their raw bytes and order at every known object level.
`verifyRetention` checks the guarded current document's retained members against
the proposal, including members beneath removed known objects. JSON comparison
permits equivalent object order, escapes, whitespace and exact decimal notation
without float rounding or huge-exponent arithmetic. Entire original document
bytes are a separate immutable request-private string for later raw backup.

Private canonical scoped keys are `remote:` or `local-common:` followed by
canonical unpadded base64url of the exact opaque RepositoryID token bytes. This
also preserves non-UTF-8 local tokens. Decoding a key establishes no observed
remote/local association or filesystem authority. Legacy keys retain case and
all exact values; noncurrent scoped records remain in the document.

Versions hash every original byte and distinguish family, physical store,
presence and selected run WorktreeID/root identity. The private physical-store
token hashes the native parent identity (or nearest existing absence anchor),
literal remaining component sequence and fixed basename. Tokens are equality
evidence; native requests must independently acquire/revalidate the binding and
never decode a token into a filesystem path. Byte-for-byte restoration in the
same binding has the same content version; native file/security comparison is
a separate publication precondition.

Verification for this milestone: `go test ./internal/persistence -count=1`,
`go vet ./internal/persistence`, and the bounded `FuzzDocumentCodec` target.
Tests exercise all nested known/unknown families, exact ordered/whitespace intent,
null/empty/default distinctions, forward/overflow schemas, malformed/reserved
shapes, escaped duplicate keys, UTF-8/surrogates, limits, lossless retained values,
copied buffers/concurrency, whole-byte and foreign-store/root versions.

Pending native milestones retain the complete selected contract: request-owned
no-follow acquisition, supported metadata, permanent cooperative locks, missing
parent revalidation, exact class65/class11 or Renameat/Linkat publication,
immutable raw plus late-write-retained originals, bounded recovery admission,
stable persisted recovery IDs and truthful commit/error/cancel outcomes. Native
Windows/Linux/macOS/FreeBSD execution, faults/crashes/concurrent processes and
all twelve builds remain separate mandatory acceptance evidence. Arbitrary
external edits in the final-check/publication gap remain outside cooperative
CAS; the implementation must test and report this accepted limitation.
