package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func recoveryRecords(t testing.TB, records []api.StorageRecovery) map[api.RecoveryID]api.StorageRecoveryData {
	t.Helper()
	out := make(map[api.RecoveryID]api.StorageRecoveryData, len(records))
	for _, record := range records {
		data := record.Data()
		id := data.Record.Data().RecoveryID
		if _, exists := out[id]; exists {
			t.Fatal("duplicate recovery identity")
		}
		out[id] = data
	}
	return out
}

func storeFileBytes(t testing.TB, root string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out[entry.Name()] = string(raw)
	}
	return out
}

// A smaller admission setting must not erase the facts left by prior writers
// using a larger setting. The same applies to later growth of retained files.
func TestAdmissionRefusalPreservesAllExistingRecoveryFacts(t *testing.T) {
	for _, limit := range []string{"records", "bytes"} {
		t.Run(limit, func(t *testing.T) {
			root := physicalStoreTemp(t)
			s := newTestStore(t, root)
			for _, value := range []string{"first", "second", "third"} {
				loaded, err := s.LoadUserConfig(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				version, _ := loaded.Observation().Data().Version.Value()
				result, err := s.CommitUserConfig(context.Background(), userProposal(t, version, value))
				assertCommitted(t, result, err)
			}
			baseline, err := s.LoadUserConfig(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			want := recoveryRecords(t, baseline.Observation().Data().Recovery)
			if len(want) < 10 {
				t.Fatalf("fixture did not establish three complete operations: %d", len(want))
			}
			files := storeFileBytes(t, root)
			options := s.options
			if limit == "records" {
				options.RecoveryMaxRecords = 1
			} else {
				options.RecoveryMaxBytes = 1
			}
			limited, err := New(context.Background(), options)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := limited.LoadUserConfig(context.Background())
			if !errors.Is(err, errRecoveryCapacity) || !loaded.Valid() || !reflect.DeepEqual(loaded.Document(), baseline.Document()) || loaded.Observation().Data().Version != baseline.Observation().Data().Version {
				t.Fatalf("capacity load lost usable current facts: %v", err)
			}
			if got := recoveryRecords(t, loaded.Observation().Data().Recovery); !reflect.DeepEqual(got, want) {
				t.Errorf("capacity load lost recovery facts: got %d, want %d", len(got), len(want))
			}
			version, _ := baseline.Observation().Data().Version.Value()
			result, err := limited.CommitUserConfig(context.Background(), userProposal(t, version, "refused"))
			if !errors.Is(err, errRecoveryCapacity) || !result.Valid() || result.Data().Outcome != api.NotCommitted || result.Data().PublicationKnown || result.Data().CurrentVersion != baseline.Observation().Data().Version || !result.Data().ProposedVersion.Present() {
				t.Fatalf("capacity commit lost refusal/version facts: %+v %v", result.Data(), err)
			}
			if got := recoveryRecords(t, result.Data().Recovery); !reflect.DeepEqual(got, want) {
				t.Errorf("capacity commit lost recovery facts: got %d, want %d", len(got), len(want))
			}
			if !reflect.DeepEqual(storeFileBytes(t, root), files) {
				t.Fatal("admission refusal changed current or retained bytes/names")
			}
		})
	}
}
