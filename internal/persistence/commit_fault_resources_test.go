package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

var commitPreparationStages = []string{
	"prepare.manifest", "prepare.manifest.journal.write", "prepare.manifest.journal.flush",
	"prepare.payload", "prepare.payload.write", "prepare.payload.flush", "prepare.payload.journal", "prepare.payload.journal.write", "prepare.payload.journal.flush",
	"prepare.raw", "prepare.raw.write", "prepare.raw.flush", "prepare.raw.journal", "prepare.raw.journal.write", "prepare.raw.journal.flush",
	"prepare.original", "prepare.original.journal", "prepare.original.journal.write", "prepare.original.journal.flush",
	"prepare.publication", "prepare.publication.journal", "prepare.publication.journal.write", "prepare.publication.journal.flush",
	"prepare.ready.journal.write", "prepare.ready.journal.flush", "manifest-flushed", "prepare.close", "final-check", "before-publication",
	"native-return-lost", "directory-flush", "outcome-delivery", "close.artifact", "close.lock", "close.parents",
}

// These are delivery failures at real request boundaries, not claims to cause
// disk-full or power loss. Native I/O/refusal controls are separate. Every case
// keeps the actual protocol and resource owners, without GC-assisted cleanup.
func TestCommitFaultAndCancellationResources(t *testing.T) {
	previousGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previousGC)
	checkResources := testRequestResources(t)
	for _, mode := range []string{"failure", "cancellation"} {
		for _, stage := range commitPreparationStages {
			t.Run(mode+"/"+stage, func(t *testing.T) {
				root := physicalStoreTemp(t)
				s := newTestStore(t, root)
				initial, err := s.LoadUserConfig(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				absent, _ := initial.Observation().Data().Version.Value()
				seed, err := s.CommitUserConfig(context.Background(), userProposal(t, absent, "old"))
				assertCommitted(t, seed, err)
				loaded, err := s.LoadUserConfig(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				baseline := recoveryRecords(t, loaded.Observation().Data().Recovery)
				expected, _ := loaded.Observation().Data().Version.Value()
				old, err := os.ReadFile(filepath.Join(root, "config.json"))
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				injected := errors.New("boundary delivery failed")
				hit := false
				s.hook = func(at string) error {
					if at != stage || hit {
						return nil
					}
					hit = true
					if mode == "cancellation" {
						cancel()
						return nil
					}
					return injected
				}
				result, err := s.CommitUserConfig(ctx, userProposal(t, expected, "new"))
				if !hit || !result.Valid() || (mode == "failure" && !errors.Is(err, injected)) || result.Data().CancellationAsked != (mode == "cancellation") {
					t.Fatalf("lost fault/cancellation: hit=%v result=%+v err=%v", hit, result.Data(), err)
				}
				checkResources(t)
				post := stage == "native-return-lost" || stage == "directory-flush" || stage == "outcome-delivery" || strings.HasPrefix(stage, "close.")
				unknown := mode == "failure" && stage == "native-return-lost"
				data := result.Data()
				if data.PublicationKnown != (post && !unknown) || (data.Outcome == api.NotCommitted) != !post || (data.Outcome == api.StorageIndeterminate) != unknown || !data.ProposedVersion.Present() || !data.CurrentVersion.Present() {
					t.Fatalf("wrong native/version facts: %+v", data)
				}
				if mode == "failure" && (stage == "prepare.close" || strings.HasPrefix(stage, "close.")) {
					found := false
					for _, diagnostic := range data.Diagnostics {
						found = found || (diagnostic.Data().Code == api.CleanupIncomplete && diagnostic.Data().Reason == "storage-"+stage)
					}
					if !found {
						t.Fatal("close error lost its separate cleanup diagnostic")
					}
				}
				for id, want := range baseline {
					if got := recoveryRecords(t, data.Recovery)[id]; !reflect.DeepEqual(got, want) {
						t.Fatal("fault discarded an earlier independent recovery fact")
					}
				}
				current, readErr := os.ReadFile(filepath.Join(root, "config.json"))
				if readErr != nil || (string(current) == string(old)) == post {
					t.Fatalf("wrong target bytes at native boundary: %q %v", current, readErr)
				}
				// A fresh request must reacquire the same permanent lock, preserve
				// the complete usable current value, and recover every returned
				// artifact still named after publication. Partial preparation may
				// independently contribute diagnostics or additional proved facts.
				restarted, loadErr := newTestStore(t, root).LoadUserConfig(context.Background())
				if !restarted.Valid() || restarted.Observation().Data().State != api.ValidCurrent || restarted.Observation().Data().Version != data.CurrentVersion {
					t.Fatalf("restart lost current facts: %v", loadErr)
				}
				recovered := recoveryRecords(t, restarted.Observation().Data().Recovery)
				for id, want := range recoveryRecords(t, data.Recovery) {
					if post && strings.HasSuffix(want.Locator, ".publication") {
						continue
					}
					if got := recovered[id]; !reflect.DeepEqual(got, want) {
						t.Fatalf("restart lost or altered returned recovery %s: %v", want.Locator, loadErr)
					}
				}
				checkResources(t)
			})
		}
	}
}
