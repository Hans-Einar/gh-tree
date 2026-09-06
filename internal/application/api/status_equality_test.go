package api

import "testing"

func TestStatusIndexFactsIgnoreOrderButRetainStageIdentity(t *testing.T) {
	p := must(NewGitPath("conflict"))
	entry := func(stage uint8, digit string, flags ...IndexFlag) IndexEntryFact {
		return must(NewIndexEntryFact(IndexEntryFactData{Path: p, Stage: stage, Object: revision(local("status"), digit).OID(), Mode: 0100644, SemanticFlags: flags}))
	}
	a := []IndexEntryFact{entry(1, "1", IntentToAdd, SkipWorktree), entry(2, "2"), entry(3, "3")}
	b := []IndexEntryFact{entry(3, "3"), entry(1, "1", SkipWorktree, IntentToAdd), entry(2, "2")}
	if !sameIndexEntries(a, b) || !sameIndexEntries(b, a) {
		t.Fatal("entry/flag order invented a difference")
	}
	// Identical OID multisets assigned to different conflict sides are different
	// evidence; unordered comparison must key on stage rather than object alone.
	b = []IndexEntryFact{entry(3, "2"), entry(1, "1", SkipWorktree, IntentToAdd), entry(2, "3")}
	if sameIndexEntries(a, b) || sameIndexEntries(b, a) {
		t.Fatal("conflict side object identities interchanged")
	}
	if sameIndexEntries(a, a[:2]) || sameIndexEntries(a[:1], []IndexEntryFact{entry(2, "1", IntentToAdd, SkipWorktree)}) {
		t.Fatal("missing/replaced stage treated as the same evidence")
	}
}
