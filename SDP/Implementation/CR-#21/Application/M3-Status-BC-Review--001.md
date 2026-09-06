# M3 status cause contract review

Disposition: **ACCEPT** of the BC-CHANGE-67 proposal; no blocking findings.
Role: fresh independent Contract Reviewer, 2026-09-06, under #67 / #61 / #21.
Sprint-004-v04 / I-03 / M3. This is design review, not implementation acceptance.

## Exact reviewed authority and source

Canonical `codereview-21/refactor` at
`980513b2b126842ab1d50558931a84441c253172`; clean at review entry.
Read AGENTS.md and developmentInstructions.md completely, Issue bodies/comments,
the actual DRAFT BC-CHANGE-67 section of M3-Adapters--001, frozen
Application--Git 1.0.0 G3/G4 and their ownership/observation context, accepted
API/SLC-07 references, and the relevant actual API/viewmodel source and READMEs.
The following Git blob IDs bind the principal inspected evidence without another
source archive:

| Path | Reviewed blob |
|---|---|
| SDP/Implementation/CR-#21/M3-Adapters--001.md | cac8e5f13ab17b93bfad0fe65b7ac628facd78c7 |
| SDP/LayerBoundaryContracts/BC--Application--Git.md | a3787cd165289795c1d1ff629cec44b41f563b27 |
| internal/application/api/git_records.go | 637524288477c608bbfd3afb60d0a675d153b883 |
| internal/application/api/git_consistency.go | 836a8e9e2c421b311170b5f7afdf53dee8690a14 |
| internal/tuistate/viewmodel/rows.go | bc3830b9119b4b1f8686bb40806fefddd01b91ec |

## Independent assessment

`ChangeFactData` (git_records.go:617) cannot express which comparison produced its
single Kind. Current index/filesystem facts alone cannot recover HEAD-to-index
change, and consistentStatusFacts currently checks scope rather than cause
identity. FileChangeSpec (rows.go:338) already has two independent status fields.
The contradiction with G3 and SLC-07 is real. Required cause-tagged rows with
per-cause Kind/OldPath are a sufficient minimal correction; no new port or layer
is needed.

Git's [status documentation](https://git-scm.com/docs/git-status#_short_format)
distinguishes ordinary index/worktree comparisons from unmerged-side status and
untracked paths. Its porcelain format preserves literal rename target/source
paths and does not promise entry order. These support explicit cause identity,
unordered rows and conflict stages rather than interpreting conflict XY as
ordinary staged/unstaged changes.

The proposed restrictions admit the important cases:

- Staged deletion plus untracked replacement: two rows at the same Path,
  Index/Deleted and Untracked/Untracked; both carry an empty current index and
  the same present replacement FileState. Deleted must not require absent
  WorktreeState: it describes the selected comparison.
- Staged rename A to B plus an edit or deletion of B: Index/Renamed at B with
  OldPath A, plus Worktree/Modified or Deleted at B. Their current facts agree.
  A subsequent independently observed worktree rename B to C instead uses a
  Worktree/Renamed row at C with OldPath B. Do not impose a shared OldPath or
  require a destination index entry for every renamed row.
- All seven nonempty subsets of stages 1..3 fit Conflict/Unmerged, including
  stage 1 alone for both-deleted and stages 2+3 for both-added. Absent sides
  remain absent; conflict FileState can be present or absent. Suppressing
  ordinary causes only for that exact conflicting Path is appropriate.
- Unborn comparison uses the empty tree without inventing a Revision. Partial
  observations retain independently known rows; empty Partial/Unknown is not
  clean. A row whose mandatory current facts cannot be established is omitted
  with incomplete/diagnostic evidence, not supplied with fabricated facts.

These are representation checks, not claims of executed adapter fixtures.
No native mutation, product edit, implementation test or CI run was needed for
this documentation review. Viewmodel's single OldPath/Kind means later projection
must preserve distinct cause records when one merged display row would lose
information; that remains the expressly pending M4/M5 gate.

## Required implementation evidence

The API worker must cover staged-only/unstaged-only/both, the valid cases above,
all conflict stage subsets, exact unusual path bytes and immutable nested copies.
Negative cases must reject zero/unknown causes, incompatible cause/kind,
missing/extraneous/self OldPath, duplicate (Path,Cause), duplicate stages,
stage 0 mixed with conflict stages, conflicting same-path current facts and
conflict coexistence. Compare current index entries by stage and semantic flags
as facts; ordering must not invent a contradiction. Preserve successful and
partial result envelopes and existing repository/worktree/version bindings.
Do not infer cause from token bytes or strengthen admission into nonexistent
native observations.

#61 must independently prove the same cases in owned native repositories,
including intent-to-add, SHA-1/SHA-256, no index refresh/write, rename/copy
source/destination interpretation and source drift during acquisition. Incomplete
or failed reads must not report clean. API constructor tests cannot prove these
adapter obligations. M4/M5 must prove lossless public and presentation projection.

Next permitted action: Master records this review and freezes the narrowly
corresponding Application--Git 1.1.0 delta, then dispatches the fresh API worker
and separate implementation reviewer described in #67. Frozen 1.0.0 remains
effective until that freeze; ObserveStatus-dependent work remains paused until
the reviewed prerequisite is integrated. No Slice or baseline finding is closed.
This new report is the only review edit; it is uncommitted for Master integration.
