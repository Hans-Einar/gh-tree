package tree

import (
	"reflect"
	"testing"
)

func TestNormalizeBranch(t *testing.T) {
	t.Parallel()
	prefixes := []string{"steering/", "codex", "worker", "review", "agent", "fix", "feature"}
	tests := map[string]string{
		"steering/Concept1/ui-box":               "Concept1/ui-box",
		"codex/MVP1/machine-service/slc003":      "MVP1/machine-service/slc003",
		"review/emulator/timer2-slc015":          "emulator/timer2-slc015",
		"codex/worker/Nested/service/change":     "Nested/service/change",
		"custom/namespace/ordinary-branch":       "custom/namespace/ordinary-branch",
		"ordinary-unstructured-branch":           "misc/ordinary-unstructured-branch",
		"/feature//Geometry/overlay/persistence": "Geometry/overlay/persistence",
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeBranch(input, prefixes); got != want {
				t.Fatalf("NormalizeBranch(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestEntriesBuildImmediateHierarchy(t *testing.T) {
	t.Parallel()
	items := []PathItem{
		{ID: "1", Path: "MVP1/machine-service/slc003", Label: "#1 machine"},
		{ID: "2", Path: "MVP1/simulator/slc004", Label: "#2 simulator"},
		{ID: "3", Path: "Concept1/ui-box", Label: "#3 UI"},
	}
	root := Entries(items, "", "")
	got := []string{root[0].Label, root[1].Label}
	want := []string{"Concept1/", "MVP1/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("root labels = %v, want %v", got, want)
	}

	mvp := Entries(items, "MVP1", "simulator")
	if len(mvp) != 1 || !mvp[0].IsFolder || mvp[0].Path != "MVP1/simulator" {
		t.Fatalf("filtered MVP1 entries = %#v", mvp)
	}
}

func TestResolveFolderFallsBackToNearestExistingAncestor(t *testing.T) {
	t.Parallel()
	items := []PathItem{{ID: "1", Path: "Concept1/ui-box", Label: "UI"}}
	tests := map[string]string{
		"Concept1":                "Concept1",
		"Concept1/stale/deeper":   "Concept1",
		"removed/folder":          "",
		"":                        "",
		"/Concept1/stale/deeper/": "Concept1",
	}
	for saved, want := range tests {
		if got := ResolveFolder(items, saved); got != want {
			t.Errorf("ResolveFolder(%q) = %q, want %q", saved, got, want)
		}
	}
}
