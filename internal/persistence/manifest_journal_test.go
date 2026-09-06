package persistence

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func TestManifestJournalPartialPreparationRetainsPersistedIDs(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	load, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := load.Observation().Data().Version.Value()
	injected := errors.New("stop after payload snapshot")
	s.hook = func(stage string) error {
		if stage == "prepare.publication" {
			return injected
		}
		return nil
	}
	result, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "prepared"))
	if !errors.Is(err, injected) || !result.Valid() || result.Data().Outcome != api.NotCommitted || len(result.Data().Recovery) != 2 {
		t.Fatalf("partial preparation facts: %v %v", result, err)
	}
	fresh := newTestStore(t, root)
	after, err := fresh.LoadUserConfig(context.Background())
	if !errors.Is(err, errIncompletePreparation) || !after.Valid() || after.Observation().Data().State != api.LoadAbsent {
		t.Fatalf("partial load lost absence/error: %v %v", after, err)
	}
	if !reflect.DeepEqual(result.Data().Recovery, after.Observation().Data().Recovery) {
		t.Fatal("partial restart lost or reminted persisted recovery")
	}
}

func TestManifestJournalTornTailPreservesEarlierFacts(t *testing.T) {
	root := physicalStoreTemp(t)
	s := newTestStore(t, root)
	load, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v, _ := load.Observation().Data().Version.Value()
	result, err := s.CommitUserConfig(context.Background(), userProposal(t, v, "committed"))
	assertCommitted(t, result, err)
	var name string
	for _, record := range result.Data().Recovery {
		if record.Data().Kind == api.Manifest {
			name = record.Data().Locator
		}
	}
	file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte(`{"Body":{"Sequence":`)); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	loaded, err := newTestStore(t, root).LoadUserConfig(context.Background())
	if err == nil || !loaded.Valid() || loaded.Observation().Data().State != api.ValidCurrent || len(loaded.Observation().Data().Recovery) < 2 {
		t.Fatalf("torn tail erased committed current or earlier records: %v %v", loaded, err)
	}
	ids := map[api.RecoveryID]bool{}
	for _, record := range result.Data().Recovery {
		ids[record.Data().Record.Data().RecoveryID] = true
	}
	for _, record := range loaded.Observation().Data().Recovery {
		if !ids[record.Data().Record.Data().RecoveryID] {
			t.Fatal("torn tail reminted recovery ID")
		}
	}
}
