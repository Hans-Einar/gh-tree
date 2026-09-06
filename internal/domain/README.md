# Domain values

Issue #58 / M2 implements API--001 A1 and the FROZEN 1.0.0 boundary vocabulary.
All values have private fields, validating constructors, `Valid`, copied getters
and `Equal`; ordinary Go `==` and map keys use the same exact structural equality.
The zero of every identity, Head, ExactTarget and discriminator is invalid.
Equality does not imply validity: two invalid zero values still compare equal.
Every failed constructor returns invalid zero and a validation error. Boundary
consumers must reject invalid values before use. No Domain value proves that an
object exists, is currently observed or authorizes a mutation.

| Value / constructor | Identity and validation |
|---|---|
| RepositoryID / NewRepositoryID | Remote or LocalCommon plus nonempty opaque token. Tokens retain all bytes; adapters mint and canonicalize them. |
| WorktreeID / NewWorktreeID | LocalCommon repository and nonempty administrative key. Paths, display, active/current status and HEAD are observations. |
| BranchID / NewBranchID | Local in LocalCommon or RemoteHead in Remote, with exact stored name. Cached remote-tracking refs stay adapter-owned. |
| PRID / NewPRID | Remote base repository and positive uint64 number. |
| OID / NewOID | Exactly 40 SHA-1 or 64 SHA-256 hex digits, nonzero; format and fixed bytes retained. String is canonical lowercase; Bytes returns a copy. |
| Revision / NewRevision | Explicit repository and exact OID. Adapter verifies commit type/existence; equal OID bytes cannot associate repositories. |
| Head / NewAttachedHead, NewDetachedHead, NewUnbornHead | Closed tagged Attached(branch, revision), Detached(revision), Unborn(branch). All repositories local; attached endpoints agree. |
| StashID / NewStashID | LocalCommon repository and exact stash OID; reflog position, origin, label and source worktree are excluded. |
| LaunchPointID / NewLaunchPointID | Worktree and length-delimited provider/project/member bytes. Empty project identifies root; provider/member are nonempty. |
| SessionID / NewSessionID | Positive uint64 supplied by Runtime. No allocation, kind routing, uniqueness claim or wrap operation. |
| ExactTarget / NewCommitTarget, NewBranchTarget, NewPullRequestTarget | Closed tagged family; every alternative preserves expected Revision. Branch endpoints share a scope. PRID's base and expected head's explicit remote repository may differ. |

Head and ExactTarget are concrete comparable private tagged structs, as permitted
by BoundaryTypes B1. Their exported kind constants select inspection only; there
is no public constructor accepting a discriminator plus optional payloads, no
open interface that consumers can extend by embedding, and no typed-nil variant.
Head.Branch/Revision and target Branch/PullRequest return `(value, present)`;
absence returns invalid zero and false. Invalid values expose no optional payload.
Unborn cannot produce a Revision or enter an exact-target operation through its
absent Revision. ExpectedRevision on an invalid target returns invalid zero.

Head.MatchesWorktree checks the local common-repository scope. Linked worktrees
may share that scope; the Git/API observation constructor must additionally bind
the actual observed Head to the intended worktree. Domain keeps no inventory.
PR targets retain both the PR base scope and the exact remote head scope through
their getters. Application separately verifies fork/host/local associations; it
must not replace either scope merely because OID bytes match.

## Exact branch identity rule

The name is the entire literal suffix beneath Git's `refs/heads/` namespace.
The pure byte rule follows stored ref syntax in
[git-check-ref-format](https://git-scm.com/docs/git-check-ref-format): reject empty
names/components, components starting with dot or ending `.lock`, consecutive
dots, trailing dot, ASCII bytes 0..32 and DEL, `~ ^ : ? * [ \\`, and `@{`.
Names cannot start/end with slash or contain doubled slash. Every other byte is
preserved, including case, Unicode and non-UTF-8 bytes; there is no trimming,
Unicode normalization, label-prefix stripping, refspec parsing or shorthand
expansion. `@{-n}` is always invalid. This is exact identity, not creation intent.

Literal `main`, `-topic`, `HEAD`, `@`, `origin/main` and `refs/heads/main` are valid
distinct suffixes. In particular the complete stored ref for suffix `@` is
`refs/heads/@`, not the forbidden single-character full ref `@`. Likewise suffix
`refs/heads/main` means `refs/heads/refs/heads/main`; it never becomes `main`.
The Git adapter adds the native namespace prefix once, separately revalidates
the actual native ref and applies operation-specific porcelain restrictions
(including leading dash/HEAD rules). It never passes a suffix as an unchecked
revision expression or expands mutable previous-checkout shorthand.

## Launch identity and ownership

Launch key encoding is decimal **byte** length, colon, exact bytes for each of
provider, project, member in that order. For example root npm member `a/b` is
`3:npm0:3:a/b`, while project `a` member `b` is `3:npm1:a1:b`. The decoder can
consume exactly each indicated length, so delimiters inside components cannot
alias boundaries. `Key` is a copied immutable identity value; there is no raw-key
constructor. Project is an opaque adapter-minted key, not a Domain filesystem
path. Project component validation and canonicalization stay in Discovery.
Labels, aliases, executable overrides and content/source versions are absent
from the identity. Saved aliases and source expectations remain API values.

Domain performs no environment, URL/path, clock, context, filesystem, network,
process, storage, transport, serialization or allocator work. Standard-library
imports are limited to errors, strings, strconv and encoding/hex. A constructor
validates supplied identity; Runtime owns SessionID allocation, no reuse and
exhaustion before wrap. Independent local clones and relocated repositories get
distinct adapter-minted local tokens; linked worktrees share common scope with
different administrative keys.

## Verification

`go test ./internal/domain -count=1` runs deterministic constructor/byte/scope/
copy/closed-value tests without native Git, network or environment dependencies.
`go test -race ./internal/domain -count=1` checks the same tests with race support.
External-consumer tests require private comparable structs and no mutable/open
representation; package-local tests reject contradictory tag/payload states.
`go run ./internal/composition/architecture` checks imports/public value surfaces
and unchanged legacy sources across all twelve target selections. Full-suite
and configured source-branch CI supplement these tests. These are Domain M2
foundation evidence for V-DOM-01..03, not completed Slices, adapter observations,
Runtime allocation/exhaustion, native Git validation or release proof.
