package launchdiscovery

import (
	"context"
	"reflect"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

var makeCommentCases = []struct {
	name, source       string
	available, limited bool
}{
	{"ordinary", "all: dep # ordinary comment\n", true, false},
	{"dollar prose", "all: dep # costs $5\n", true, false},
	{"percent prose", "all: dep # 100% done\n", true, false},
	{"backslash prose", "all: dep # C:\\tools costs $5, 100%\n", true, false},
	{"paired comment backslashes", "all: dep # comment \\\\\n", true, false},
	{"comment backslash then space", "all: dep # comment \\ \n", true, false},
	{"escaped marker", "all: dep \\# literal\n", false, true},
	{"dynamic target", "$(TARGET): dep # ordinary\n", false, true},
	{"pattern target", "all%: dep # ordinary\n", false, true},
	{"dynamic prerequisites", "all: $(DEPS) # costs $5\n", false, true},
	{"rule continuation", "all: dep \\\n tail: dep\n", false, true},
	{"inline comment continuation", "all: dep # costs $5 \\\n tail: dep\n", false, true},
	{"comment continuation hides rule", "# comment \\\nall: dep\n", false, true},
	{"recipe continuation hides rule", "\t echo \\\nall: dep\n", false, true},
	{"chained CRLF continuation", "# comment \\\r\n another \\\r\nall: dep\r\n", false, true},
}

func TestMakeCommentClassification(t *testing.T) {
	for _, tc := range makeCommentCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := parseMake(context.Background(), []byte(tc.source+"safe: dep\n"), 1024)
			if err != nil {
				t.Fatal(err)
			}
			want := []member{{"safe", true}}
			if tc.available {
				want = append([]member{{"all", true}}, want...)
			}
			if !reflect.DeepEqual(p.members, want) || (len(p.notices) > 0) != tc.limited {
				t.Fatalf("members=%v notices=%v; want %v limitation=%v", p.members, p.notices, want, tc.limited)
			}
		})
	}
	// A final comment backslash without a newline is only comment prose.
	p, err := parseMake(context.Background(), []byte("all: dep # prose \\"), 1024)
	if err != nil || len(p.members) != 1 || len(p.notices) != 0 {
		t.Fatal(p, err)
	}
}

func TestMakeCommentsDiscoverAndSavedResolve(t *testing.T) {
	for _, tc := range makeCommentCases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			put(t, root, "GNUmakefile", "all: dep\nsafe: dep\n")
			put(t, root, "independent/package.json", `{"scripts":{"dev":"never executed"}}`)
			a := must(New(Config{}))
			scope := fixtureScope(t, root)
			entries := []api.SavedLaunchEntry{savedEntry("selected", "make", "", "", "", []string{"all"})}
			baseline := discover(t, a, scope, entries...).Saved[0].Data()
			baselineID, _ := baseline.LaunchPointID.Value()
			baselineSource, _ := baseline.SourceVersion.Value()
			baselineSelection := must(api.NewSavedLaunch(api.SavedLaunchData{Alias: baseline.Alias, LaunchPointID: baselineID, StorageVersion: baseline.StorageVersion, SourceExpectation: baselineSource}))
			put(t, root, "GNUmakefile", tc.source+"safe: dep\n")
			r := discover(t, a, scope, entries...)
			pick(t, r.Definitions, api.Make, "", "safe")
			pick(t, r.Definitions, api.Npm, "independent", "dev")
			found := false
			for _, def := range r.Definitions {
				if def.Data().Provider == api.Make && def.Data().Member == "all" {
					found = true
				}
			}
			if found != tc.available || r.Saved[0].Data().Definition.Present() != tc.available {
				t.Fatal("incorrect target availability", r.Definitions, r.Saved)
			}
			if (r.Observation.Data().Completeness != api.Complete) != tc.limited {
				t.Fatal("incorrect profile completeness", r.Observation.Data())
			}
			if !tc.available {
				if len(r.Saved[0].Data().Diagnostics) == 0 {
					t.Fatal("missing saved refusal")
				}
				out, err := resolve(t, a, scope, baselineSelection, entries, api.Some(baseline.StorageVersion))
				if err == nil || out.Data().Invocation.Present() {
					t.Fatal("unsupported source authorized saved target")
				}
				return
			}
			saved := r.Saved[0].Data()
			id, _ := saved.LaunchPointID.Value()
			version, _ := saved.SourceVersion.Value()
			selection := must(api.NewSavedLaunch(api.SavedLaunchData{Alias: saved.Alias, LaunchPointID: id, StorageVersion: saved.StorageVersion, SourceExpectation: version}))
			out, err := resolve(t, a, scope, selection, entries, api.Some(saved.StorageVersion))
			if err != nil {
				t.Fatal(err)
			}
			inv, ok := out.Data().Invocation.Value()
			if !ok {
				t.Fatal("missing invocation")
			}
			if got := inv.Data().Execution.(api.ArgvExecution).Data().Arguments; !reflect.DeepEqual(got, []string{"-f", "GNUmakefile", "all"}) {
				t.Fatal(got)
			}
		})
	}
}
