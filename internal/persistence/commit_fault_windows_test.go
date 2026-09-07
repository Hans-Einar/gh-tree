package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"golang.org/x/sys/windows"
)

func TestWindowsCommitNativeSharingFailure(t *testing.T) {
	root := physicalStoreTemp(t)
	original := []byte(`{"schemaVersion":1}`)
	if err := os.WriteFile(filepath.Join(root, "config.json"), original, 0600); err != nil {
		t.Fatal(err)
	}
	s := newTestStore(t, root)
	loaded, err := s.LoadUserConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	version, _ := loaded.Observation().Data().Version.Value()
	checkResources := testRequestResources(t)
	var reader windows.Handle
	defer func() {
		if reader != 0 {
			windows.CloseHandle(reader)
		}
	}()
	s.hook = func(stage string) error {
		if stage == "lock" {
			name, err := windows.UTF16PtrFromString(filepath.Join(root, "config.json"))
			if err != nil {
				t.Fatal(err)
			}
			reader, err = windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
			if err != nil {
				t.Fatal(err)
			}
		}
		return nil
	}
	result, err := s.CommitUserConfig(context.Background(), userProposal(t, version, "new"))
	if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) || !result.Valid() || result.Data().Outcome != api.NotCommitted || result.Data().PublicationKnown || result.Data().CurrentVersion != loaded.Observation().Data().Version {
		t.Fatalf("native sharing refusal lost current/effect: %+v %v", result.Data(), err)
	}
	if err := windows.CloseHandle(reader); err != nil {
		t.Fatal(err)
	}
	reader = 0
	checkResources(t)
	current, err := os.ReadFile(filepath.Join(root, "config.json"))
	if err != nil || string(current) != string(original) {
		t.Fatal("native sharing refusal changed current")
	}
	restarted, err := newTestStore(t, root).LoadUserConfig(context.Background())
	if err != nil || restarted.Observation().Data().Version != loaded.Observation().Data().Version {
		t.Fatalf("sharing refusal stranded request ownership: %v", err)
	}
}
