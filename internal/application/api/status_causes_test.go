package api_test

import (
	"fmt"
	"testing"

	a "github.com/Hans-Einar/gh-tree/internal/application/api"
	d "github.com/Hans-Einar/gh-tree/internal/domain"
)

func statusPath(s string) a.GitPath { return rvMust(a.NewGitPath(s)) }
func statusFile(p a.GitPath) a.PresentFile {
	return rvMust(a.NewPresentFile(a.PresentFileData{Path: p, ObjectIdentity: rvSource("object"), Kind: a.RegularFile, Mode: 0100644, Content: rvSource("content"), ParentIdentity: rvSource("parent")}))
}
func statusEntry(p a.GitPath, stage uint8, flags ...a.IndexFlag) a.IndexEntryFact {
	return rvMust(a.NewIndexEntryFact(a.IndexEntryFactData{Path: p, Stage: stage, Object: rvRev(rvRepo("status"), "2").OID(), Mode: 0100644, SemanticFlags: flags}))
}
func statusRow(p a.GitPath, cause a.ChangeCause, kind a.ChangeKind, entries ...a.IndexEntryFact) a.ChangeFactData {
	return a.ChangeFactData{Path: p, Cause: cause, Kind: kind, IndexEntries: entries, WorktreeState: statusFile(p)}
}
func statusData(rows ...a.ChangeFact) a.StatusFactsData {
	v := rvStatus(rvWork(rvRepo("status"), "primary")).Data()
	v.Changes = rows
	return v
}
func rejectChange(t *testing.T, v a.ChangeFactData) {
	t.Helper()
	got, err := a.NewChangeFact(v)
	if err == nil || got.Valid() {
		t.Fatal("invalid change admitted or returned valid value", got.Data(), err)
	}
}
func rejectStatus(t *testing.T, rows ...a.ChangeFact) {
	t.Helper()
	got, err := a.NewStatusFacts(statusData(rows...))
	if err == nil || got.Valid() {
		t.Fatal("contradictory status admitted or returned valid value", err)
	}
}

func TestStatusIndependentComparisons(t *testing.T) {
	p := statusPath("changed")
	index := rvMust(a.NewChangeFact(statusRow(p, a.IndexChangeCause, a.Modified, statusEntry(p, 0))))
	worktree := rvMust(a.NewChangeFact(statusRow(p, a.WorktreeChangeCause, a.Modified, statusEntry(p, 0))))
	// Identical current index and filesystem values are deliberately shared.
	// Only the observer knows whether HEAD-to-index, index-to-filesystem or both
	// changed; admission must retain that evidence without decoding content tokens.
	for _, tc := range []struct {
		name string
		rows []a.ChangeFact
		want []a.ChangeCause
	}{
		{"staged-only", []a.ChangeFact{index}, []a.ChangeCause{a.IndexChangeCause}},
		{"unstaged-only", []a.ChangeFact{worktree}, []a.ChangeCause{a.WorktreeChangeCause}},
		{"both", []a.ChangeFact{index, worktree}, []a.ChangeCause{a.IndexChangeCause, a.WorktreeChangeCause}},
		{"both-reversed", []a.ChangeFact{worktree, index}, []a.ChangeCause{a.WorktreeChangeCause, a.IndexChangeCause}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := rvMust(a.NewStatusFacts(statusData(tc.rows...)))
			envelope := rvMust(a.NewObserveStatusResult(a.ObserveStatusResultData{Status: a.Some(s), Observation: a.Some(s.Data().Observation), Transport: rvTransport()}))
			retained, _ := envelope.Data().Status.Value()
			if len(retained.Data().Changes) != len(tc.want) {
				t.Fatal("comparison evidence collapsed")
			}
			for i, row := range retained.Data().Changes {
				if row.Data().Cause != tc.want[i] {
					t.Fatal("comparison evidence changed", row.Data())
				}
			}
		})
	}
}

func TestStatusRenameDeletionAndCopyIdentity(t *testing.T) {
	aPath, bPath, cPath := statusPath(" old\t\n\xff"), statusPath("new\n\t\xfe "), statusPath("next\xff")
	rename := statusRow(bPath, a.IndexChangeCause, a.Renamed, statusEntry(bPath, 0))
	rename.OldPath = a.Some(aPath)
	for _, kind := range []a.ChangeKind{a.Modified, a.Deleted} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			second := statusRow(bPath, a.WorktreeChangeCause, kind, statusEntry(bPath, 0))
			first := rename
			if kind == a.Deleted {
				absent := rvMust(a.NewAbsentFile(a.AbsentFileData{Path: bPath}))
				first.WorktreeState, second.WorktreeState = absent, absent
			}
			s := rvMust(a.NewStatusFacts(statusData(rvMust(a.NewChangeFact(second)), rvMust(a.NewChangeFact(first)))))
			rows := s.Data().Changes
			old, present := rows[1].Data().OldPath.Value()
			if !present || old != aPath || rows[1].Data().Path != bPath || rows[0].Data().Kind != kind || rows[0].Data().OldPath.Present() {
				t.Fatal("staged rename and independent edit/deletion identity lost")
			}
		})
	}
	// A later worktree rename B->C has its own destination, with no index entry.
	chain := statusRow(cPath, a.WorktreeChangeCause, a.Renamed)
	chain.OldPath = a.Some(bPath)
	s := rvMust(a.NewStatusFacts(statusData(rvMust(a.NewChangeFact(rename)), rvMust(a.NewChangeFact(chain)))))
	last := s.Data().Changes[1].Data()
	old, _ := last.OldPath.Value()
	if last.Path != cPath || old != bPath || len(last.IndexEntries) != 0 {
		t.Fatal("rename chain replaced with aggregate source/destination")
	}
	// Distinct causes may have distinct OldPath even at the same destination.
	copyRow := rename
	copyRow.Cause, copyRow.Kind, copyRow.OldPath = a.WorktreeChangeCause, a.Copied, a.Some(cPath)
	if _, err := a.NewStatusFacts(statusData(rvMust(a.NewChangeFact(rename)), rvMust(a.NewChangeFact(copyRow)))); err != nil {
		t.Fatal("per-cause copy source refused", err)
	}
	deleted := statusRow(bPath, a.IndexChangeCause, a.Deleted)
	untracked := statusRow(bPath, a.UntrackedChangeCause, a.Untracked)
	replacement := rvMust(a.NewStatusFacts(statusData(rvMust(a.NewChangeFact(deleted)), rvMust(a.NewChangeFact(untracked)))))
	for _, row := range replacement.Data().Changes {
		if _, ok := row.Data().WorktreeState.(a.PresentFile); !ok || len(row.Data().IndexEntries) != 0 {
			t.Fatal("staged deletion lost present untracked replacement")
		}
	}
}

func TestChangeCauseAdmission(t *testing.T) {
	p := statusPath("entry")
	for _, cause := range []a.ChangeCause{0, a.IndexChangeCause, a.WorktreeChangeCause, a.UntrackedChangeCause, a.ConflictChangeCause, 255} {
		for _, kind := range []a.ChangeKind{0, a.Added, a.Modified, a.Deleted, a.Renamed, a.Copied, a.TypeChanged, a.Untracked, a.Unmerged, 255} {
			t.Run(fmt.Sprintf("cause-%d-kind-%d", cause, kind), func(t *testing.T) {
				row := statusRow(p, cause, kind)
				if kind == a.Renamed || kind == a.Copied {
					row.OldPath = a.Some(statusPath("source"))
				}
				if cause == a.ConflictChangeCause {
					row.IndexEntries = []a.IndexEntryFact{statusEntry(p, 1)}
				}
				want := ((cause == a.IndexChangeCause || cause == a.WorktreeChangeCause) && kind >= a.Added && kind <= a.TypeChanged) ||
					(cause == a.UntrackedChangeCause && kind == a.Untracked) || (cause == a.ConflictChangeCause && kind == a.Unmerged)
				got, err := a.NewChangeFact(row)
				if (err == nil) != want || got.Valid() != want {
					t.Fatalf("admission=%v validity=%v; want %v", err, got.Valid(), want)
				}
				if cause.Valid() != (cause >= a.IndexChangeCause && cause <= a.ConflictChangeCause) {
					t.Fatal("cause validity")
				}
			})
		}
	}
	for name, mutate := range map[string]func(*a.ChangeFactData){
		"rename-without-source":     func(v *a.ChangeFactData) { v.Kind = a.Renamed },
		"copy-without-source":       func(v *a.ChangeFactData) { v.Kind = a.Copied },
		"extraneous-source":         func(v *a.ChangeFactData) { v.OldPath = a.Some(statusPath("old")) },
		"self-rename":               func(v *a.ChangeFactData) { v.Kind, v.OldPath = a.Renamed, a.Some(p) },
		"self-copy":                 func(v *a.ChangeFactData) { v.Kind, v.OldPath = a.Copied, a.Some(p) },
		"invalid-source":            func(v *a.ChangeFactData) { v.Kind, v.OldPath = a.Renamed, a.Some(a.GitPath{}) },
		"zero-path":                 func(v *a.ChangeFactData) { v.Path = a.GitPath{} },
		"missing-filesystem":        func(v *a.ChangeFactData) { v.WorktreeState = nil },
		"different-filesystem-path": func(v *a.ChangeFactData) { v.WorktreeState = statusFile(statusPath("other")) },
		"different-index-path":      func(v *a.ChangeFactData) { v.IndexEntries = []a.IndexEntryFact{statusEntry(statusPath("other"), 0)} },
		"invalid-index-entry":       func(v *a.ChangeFactData) { v.IndexEntries = []a.IndexEntryFact{{}} },
		"duplicate-stage-zero":      func(v *a.ChangeFactData) { v.IndexEntries = []a.IndexEntryFact{statusEntry(p, 0), statusEntry(p, 0)} },
		"ordinary-conflict-stage":   func(v *a.ChangeFactData) { v.IndexEntries = []a.IndexEntryFact{statusEntry(p, 2)} },
		"untracked-index-entry":     func(v *a.ChangeFactData) { v.Cause, v.Kind = a.UntrackedChangeCause, a.Untracked },
	} {
		t.Run(name, func(t *testing.T) {
			row := statusRow(p, a.IndexChangeCause, a.Modified, statusEntry(p, 0))
			mutate(&row)
			rejectChange(t, row)
		})
	}
}

func TestStatusConflictStageSubsets(t *testing.T) {
	p := statusPath("conflict")
	for mask := 1; mask < 8; mask++ {
		for _, absent := range []bool{false, true} {
			t.Run(fmt.Sprintf("stages-%03b-absent-%v", mask, absent), func(t *testing.T) {
				row := statusRow(p, a.ConflictChangeCause, a.Unmerged)
				// Reverse native entry order is equally valid and must stay intact.
				for stage := uint8(3); stage > 0; stage-- {
					if mask&(1<<(stage-1)) != 0 {
						row.IndexEntries = append(row.IndexEntries, statusEntry(p, stage))
					}
				}
				if absent {
					row.WorktreeState = rvMust(a.NewAbsentFile(a.AbsentFileData{Path: p}))
				}
				conflict := rvMust(a.NewChangeFact(row))
				other := rvMust(a.NewChangeFact(statusRow(statusPath("Conflict"), a.IndexChangeCause, a.Modified)))
				s := rvMust(a.NewStatusFacts(statusData(conflict, other)))
				got := 0
				for _, entry := range s.Data().Changes[0].Data().IndexEntries {
					got |= 1 << (entry.Data().Stage - 1)
				}
				if got != mask {
					t.Fatal("conflict sides fabricated or erased", got)
				}
				for _, cause := range []a.ChangeCause{a.IndexChangeCause, a.WorktreeChangeCause, a.UntrackedChangeCause} {
					ordinary := statusRow(p, cause, a.Modified)
					if cause == a.UntrackedChangeCause {
						ordinary.Kind = a.Untracked
					}
					other := rvMust(a.NewChangeFact(ordinary))
					rejectStatus(t, conflict, other)
					rejectStatus(t, other, conflict)
				}
			})
		}
	}
	for _, stages := range [][]uint8{nil, {0}, {0, 1}, {1, 1}, {2, 3, 2}} {
		row := statusRow(p, a.ConflictChangeCause, a.Unmerged)
		for _, stage := range stages {
			row.IndexEntries = append(row.IndexEntries, statusEntry(p, stage))
		}
		rejectChange(t, row)
	}
}

func TestStatusSamePathCurrentFactConsistency(t *testing.T) {
	p := statusPath("path")
	first := statusRow(p, a.IndexChangeCause, a.Modified, statusEntry(p, 0, a.IntentToAdd, a.SkipWorktree))
	for name, change := range map[string]func(*a.ChangeFactData){
		"index-absent": func(v *a.ChangeFactData) { v.IndexEntries = nil },
		"index-object": func(v *a.ChangeFactData) {
			e := v.IndexEntries[0].Data()
			e.Object = rvRev(rvRepo("status"), "3").OID()
			v.IndexEntries = []a.IndexEntryFact{rvMust(a.NewIndexEntryFact(e))}
		},
		"index-mode": func(v *a.ChangeFactData) {
			e := v.IndexEntries[0].Data()
			e.Mode = 0100755
			v.IndexEntries = []a.IndexEntryFact{rvMust(a.NewIndexEntryFact(e))}
		},
		"index-flags": func(v *a.ChangeFactData) {
			e := v.IndexEntries[0].Data()
			e.SemanticFlags = []a.IndexFlag{a.IntentToAdd}
			v.IndexEntries = []a.IndexEntryFact{rvMust(a.NewIndexEntryFact(e))}
		},
		"filesystem-absent": func(v *a.ChangeFactData) { v.WorktreeState = rvMust(a.NewAbsentFile(a.AbsentFileData{Path: p})) },
		"filesystem-object": func(v *a.ChangeFactData) {
			f := statusFile(p).Data()
			f.ObjectIdentity = rvSource("other-object")
			v.WorktreeState = rvMust(a.NewPresentFile(f))
		},
		"filesystem-content": func(v *a.ChangeFactData) {
			f := statusFile(p).Data()
			f.Content = rvSource("other-content")
			v.WorktreeState = rvMust(a.NewPresentFile(f))
		},
		"filesystem-parent": func(v *a.ChangeFactData) {
			f := statusFile(p).Data()
			f.ParentIdentity = rvSource("other-parent")
			v.WorktreeState = rvMust(a.NewPresentFile(f))
		},
		"filesystem-mode": func(v *a.ChangeFactData) {
			f := statusFile(p).Data()
			f.Mode = 0100755
			v.WorktreeState = rvMust(a.NewPresentFile(f))
		},
		"filesystem-kind": func(v *a.ChangeFactData) {
			f := statusFile(p).Data()
			f.Kind = a.OtherFile
			v.WorktreeState = rvMust(a.NewPresentFile(f))
		},
	} {
		t.Run(name, func(t *testing.T) {
			second := first
			second.Cause = a.WorktreeChangeCause
			change(&second)
			x, y := rvMust(a.NewChangeFact(first)), rvMust(a.NewChangeFact(second))
			rejectStatus(t, x, y)
			rejectStatus(t, y, x)
		})
	}
	second := first
	second.Cause = a.WorktreeChangeCause
	second.IndexEntries = []a.IndexEntryFact{statusEntry(p, 0, a.SkipWorktree, a.IntentToAdd, a.IntentToAdd)}
	if _, err := a.NewStatusFacts(statusData(rvMust(a.NewChangeFact(first)), rvMust(a.NewChangeFact(second)))); err != nil {
		t.Fatal("flag order/repetition invented a contradiction", err)
	}
	// A duplicate is invalid even if its supplied facts differ; no last-writer-wins.
	for _, cause := range []a.ChangeCause{a.IndexChangeCause, a.WorktreeChangeCause, a.UntrackedChangeCause, a.ConflictChangeCause} {
		row := statusRow(p, cause, a.Modified)
		if cause == a.UntrackedChangeCause {
			row.Kind = a.Untracked
		}
		if cause == a.ConflictChangeCause {
			row.Kind, row.IndexEntries = a.Unmerged, []a.IndexEntryFact{statusEntry(p, 1)}
		}
		x := rvMust(a.NewChangeFact(row))
		rejectStatus(t, x, x)
		row.WorktreeState = rvMust(a.NewAbsentFile(a.AbsentFileData{Path: p}))
		rejectStatus(t, x, rvMust(a.NewChangeFact(row)))
	}
	link := statusFile(p).Data()
	link.Kind, link.LinkTarget = a.SymlinkFile, a.Some("one")
	first.WorktreeState = rvMust(a.NewPresentFile(link))
	link.LinkTarget = a.Some("two")
	second.WorktreeState = rvMust(a.NewPresentFile(link))
	rejectStatus(t, rvMust(a.NewChangeFact(first)), rvMust(a.NewChangeFact(second)))
}

func TestStatusUnbornPartialEnvelopesAndCopies(t *testing.T) {
	p := statusPath(" file\t\n\xff ")
	flags := []a.IndexFlag{a.IntentToAdd, a.SkipWorktree}
	entry := statusEntry(p, 0, flags...)
	entries := []a.IndexEntryFact{entry}
	row := rvMust(a.NewChangeFact(statusRow(p, a.IndexChangeCause, a.Added, entries...)))
	rows := []a.ChangeFact{row}
	v := statusData(rows...)
	wd := v.Worktree.Data()
	branch := rvMust(d.NewBranchID(wd.ID.Repository(), d.Local, "main"))
	wd.Head = a.Some(rvMust(d.NewUnbornHead(branch)))
	v.Worktree = rvMust(a.NewWorktreeFacts(wd))
	s := rvMust(a.NewStatusFacts(v))
	flags[0], entries[0], rows[0] = a.AssumeUnchanged, a.IndexEntryFact{}, a.ChangeFact{}
	for _, value := range []a.StatusFacts{s, s.Clone()} {
		out := value.Data()
		out.Changes[0] = a.ChangeFact{}
		change := value.Data().Changes[0].Data()
		change.IndexEntries[0] = a.IndexEntryFact{}
		flagOut := value.Data().Changes[0].Data().IndexEntries[0].Data()
		flagOut.SemanticFlags[0] = a.AssumeUnchanged
		retained := value.Data().Changes[0].Data()
		if retained.Path.String() != p.String() || retained.Cause != a.IndexChangeCause || retained.IndexEntries[0].Data().SemanticFlags[0] != a.IntentToAdd {
			t.Fatal("nested admission/getter/clone copy was mutable")
		}
		head, _ := value.Data().Worktree.Data().Head.Value()
		if _, present := head.Revision(); present {
			t.Fatal("unborn revision fabricated")
		}
	}
	for _, completeness := range []a.Completeness{a.Complete, a.Partial, a.Unknown} {
		for _, empty := range []bool{false, true} {
			v := s.Data()
			obs := v.Observation.Data()
			obs.Completeness = completeness
			v.Observation = rvMust(a.NewGitObservation(obs))
			if empty {
				v.Changes = nil
			}
			status := rvMust(a.NewStatusFacts(v))
			diagnostics := []a.Diagnostic{rvDiag()}
			result := rvMust(a.NewObserveStatusResult(a.ObserveStatusResultData{Status: a.Some(status), Observation: a.Some(v.Observation), Diagnostics: diagnostics, Transport: rvTransport()}))
			diagnostics[0] = a.Diagnostic{}
			out := result.Data()
			out.Diagnostics[0] = a.Diagnostic{}
			retained, _ := result.Data().Status.Value()
			if retained.Data().Observation.Data().Completeness != completeness || len(retained.Data().Changes) != len(v.Changes) || !result.Data().Diagnostics[0].Valid() {
				t.Fatal("incomplete status or diagnostics collapsed to clean")
			}
			if retained.Data().IndexVersion != v.IndexVersion || retained.Data().WorktreeVersion != v.WorktreeVersion || retained.Data().ConfigurationVersion != v.ConfigurationVersion {
				t.Fatal("status version identity changed")
			}
			wrong := result.Data()
			wrong.Observation = a.Some(rvObservation(rvWork(wd.ID.Repository(), "other")))
			if _, err := a.NewObserveStatusResult(wrong); err == nil {
				t.Fatal("foreign worktree envelope admitted")
			}
		}
	}
}
