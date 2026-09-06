# M3 status cause API independent review

Disposition: **ACCEPT** of exact source; no blocking findings.
Role: fresh independent Reviewer, 2026-09-06; #67 / #61 / #21,
Sprint-004-v04 / I-03 / M3. Master retains the source-CI/integration gate.

## Frozen source and scope

Reviewed `codex/cr21-status-api` at
`675dfffb6d055caa1074fe3feaccdfbdd8b21f74`, relative to frozen-contract base
`10f687ea7df4cde1746b50ef4b5537ecdeb13c39`.
Local HEAD and remote branch matched; worktree was clean before this report.
Read AGENTS.md/developmentInstructions.md, #21/#67 bodies/comments including
freeze/dispatch, Application--Git 1.1.0 G1..G4/G11/G12 and BoundaryTypes--001,
relevant API design/README, M3-Adapters--001 decision and accepted
M3-Status-BC-Review--001. Inspected actual changed source/tests and surrounding
constructors, sealed FileState variants, immutable copies and status envelopes.
The worker report was inspected for scope/claims; it was not review evidence.

The complete eight-file diff is exactly seven API files plus M3-Status--001:
README, enums, git_records, git_consistency, status_consistency and two status
test files. No port, Domain, adapter, native import/operation, module, workflow,
frozen-contract or unrelated change. `git diff --check` passed.

## Source assessment and executed evidence

ChangeFact validates constituent values before indexing fixed stage/cause arrays.
Its closed cause/kind, exclusive distinct OldPath and unique allowed stage rules
match G3. StatusFacts keys exact GitPath and Cause, rejects duplicate causes and
same-path conflict coexistence, and compares every current FileState field plus
stage-bound path/object/mode and unordered semantic flag sets. The four admitted
IndexFlag values fit the bitset. No ordering/token interpretation supplies cause.
Per-cause Kind/OldPath and literal bytes survive unchanged; private immutable
nested values and copied slices do not expose stored flags/entries/rows.

Independently ran `go test ./internal/application/... -count=1`: API and ports
PASS on Windows amd64 with direct Go 1.25.0. Inspected and executed checked-in
status tests covering staged-only/unstaged-only/both, rename plus modification or
deletion, B-to-C rename without destination index, independent copy sources,
staged deletion plus present untracked replacement, every conflict stage subset
with present/absent files, cause/kind/path/stage negatives, same-path fact drift,
all duplicate causes, partial/unknown/complete envelopes, unborn Head, preserved
versions, foreign-worktree refusal and admission/getter/Clone nested copies.
The helper test separately discriminates reordered stages from swapped side OIDs.

The additional public-constructor probe reproduced below PASSed against the same
source using a Go overlay; it uses only pre-existing M2 `rv*` fixture builders.
It independently checks all 256 cause tags; all 15 cause subsets in both orders
plus duplicate refusal; every pair of the 16 semantic-flag sets in both orders,
with reordered/repeated flags; and all 12 individual opaque-version component
changes across object/content/parent identities. These are positive and adverse
contract controls, not native status fixtures. No product file was edited.

Exact-source [CI run 34055882003](https://github.com/Hans-Einar/gh-tree/actions/runs/34055882003)
was independently inspected: HEAD matches `675dfffb6d055caa1074fe3feaccdfbdd8b21f74`;
17 applicable jobs SUCCESS, Windows ARM64 test still RUNNING at this report's
snapshot, Runtime helper explicitly SKIPPED because M3 Runtime is not integrated.
Successful jobs include Linux race, Linux/macOS/Windows amd64 tests and all twelve
build/architecture jobs. This review does not claim the pending job passed.
Master must inspect final source CI and run the integration gate before Git
consumes the correction. Repeating the whole native/build matrix locally adds
no necessary evidence for this bounded pure API change.

## Reproduction and limits

At the reviewed source, save the following as a UTF-8 temporary `review67_test.go`
and use Go's overlay JSON `Replace` to map a nonexistent absolute path
`<checkout>/internal/application/api/review67_external_test.go` to that file.
Run `go test -overlay <overlay.json> ./internal/application/api -run
TestReview67PublicContract -count=1 -v`. The exact executed temporary source is
also retained locally at
`C:/Users/hanse/.codex/tmp/gh-tree-status-review-675dfff/review67_test.go`, SHA-256
`a47320d85083b66dff1a7cd513cfdc614c335c9a093c866449a7a886eb4282a4`.
No source archive or extra committed test artifact is needed.

Native Git status acquisition, no-index-write/source-drift proof and actual
SHA-1/SHA-256/intent-to-add fixtures remain #61. Lossless Application/State
projection remains M4/M5. No Slice or baseline finding is closed here.
Next permitted action: Master completes exact-source CI, accepts/integrates this
prerequisite, verifies the integrated source, and supplies the explicit commit to
Git. This report-only commit/push changes no executable input; no merge performed.
No local verification failed and no unresolved implementation finding was found.

```go
package api_test

import (
 "testing"
 a "github.com/Hans-Einar/gh-tree/internal/application/api"
)

// Independent contract oracle; only rv* fixture builders predate this change.
func TestReview67PublicContract(t *testing.T) {
 p := rvMust(a.NewGitPath(" literal\t\n\xff "))
 absent := rvMust(a.NewAbsentFile(a.AbsentFileData{Path:p}))
 makeRow := func(c a.ChangeCause, flags ...a.IndexFlag) a.ChangeFactData {
  row := a.ChangeFactData{Path:p, Cause:c, Kind:a.Deleted, WorktreeState:absent}
  if c == a.UntrackedChangeCause { row.Kind = a.Untracked }
  if c == a.ConflictChangeCause {
   row.Kind = a.Unmerged
   row.IndexEntries = []a.IndexEntryFact{rvMust(a.NewIndexEntryFact(a.IndexEntryFactData{Path:p, Stage:1, Object:rvRev(rvRepo("r"),"1").OID(), Mode:0100644}))}
  } else if flags != nil {
   row.IndexEntries = []a.IndexEntryFact{rvMust(a.NewIndexEntryFact(a.IndexEntryFactData{Path:p, Object:rvRev(rvRepo("r"),"1").OID(), Mode:0100644, SemanticFlags:flags}))}
  }
  return row
 }
 check := func(want bool, rows ...a.ChangeFact) {
  t.Helper()
  d := rvStatus(rvWork(rvRepo("r"),"primary")).Data(); d.Changes = rows
  s, err := a.NewStatusFacts(d)
  if (err == nil) != want || s.Valid() != want { t.Fatalf("status admission %v, want %v",err,want) }
  if want { for i, r := range s.Data().Changes { if r.Data().Cause != rows[i].Data().Cause { t.Fatal("cause/row order lost") } } }
 }
 for raw:=0; raw<256; raw++ {
  v, err := a.NewChangeFact(makeRow(a.ChangeCause(raw)))
  want := raw==1 || raw==2 || raw==3 || raw==4
  if (err==nil)!=want || v.Valid()!=want { t.Fatalf("cause %d admitted=%v",raw,err) }
 }
 // Every subset of causes and both enumeration directions; conflict stands alone.
 for mask:=1; mask<16; mask++ {
  rows:=[]a.ChangeFact{}
  for c:=1; c<=4; c++ { if mask&(1<<(c-1))!=0 { rows=append(rows,rvMust(a.NewChangeFact(makeRow(a.ChangeCause(c))))) } }
  want:=mask<8 || mask==8
  check(want,rows...)
  for left,right:=0,len(rows)-1; left<right; left,right=left+1,right-1 { rows[left],rows[right]=rows[right],rows[left] }
  check(want,rows...)
  for _,row:=range rows { check(false,row,row) }
 }
 // All 16 semantic flag sets against all others; reverse and repeat the second.
 flags := func(mask int, reverse bool) []a.IndexFlag {
  out:=make([]a.IndexFlag,0)
  for n:=1; n<=4; n++ { f:=n; if reverse { f=5-n }; if mask&(1<<(f-1))!=0 { out=append(out,a.IndexFlag(f)); if reverse { out=append(out,a.IndexFlag(f)) } } }
  return out
 }
 for x:=0; x<16; x++ { for y:=0; y<16; y++ {
  left:=rvMust(a.NewChangeFact(makeRow(a.IndexChangeCause,flags(x,false)...)))
  right:=rvMust(a.NewChangeFact(makeRow(a.WorktreeChangeCause,flags(y,true)...)))
  check(x==y,left,right); check(x==y,right,left)
 } }
 // Equality must retain every opaque version component, not merely token bytes.
 file:=a.PresentFileData{Path:p,ObjectIdentity:rvSource("id"),Kind:a.RegularFile,Mode:0100644,Content:rvSource("bytes"),ParentIdentity:rvSource("parent")}
 left:=makeRow(a.IndexChangeCause); left.WorktreeState=rvMust(a.NewPresentFile(file))
 for field:=0; field<3; field++ { for component:=0; component<4; component++ {
  values:=[]string{"git","scope","reviewer",[]string{"id","bytes","parent"}[field]}; values[component]+="-different"
  changed:=rvMust(a.NewSourceVersion(values[0],values[1],values[2],values[3]))
  f:=file; if field==0 { f.ObjectIdentity=changed }; if field==1 { f.Content=changed }; if field==2 { f.ParentIdentity=changed }
  right:=makeRow(a.WorktreeChangeCause); right.WorktreeState=rvMust(a.NewPresentFile(f))
  check(false,rvMust(a.NewChangeFact(left)),rvMust(a.NewChangeFact(right)))
 } }
 t.Log("PASS: all 256 causes; 15 cause subsets in both orders and duplicates; 256 flag-set pairs in both orders; 12 opaque-version component negatives")
}
```
