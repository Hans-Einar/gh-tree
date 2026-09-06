package persistence

import (
	"crypto/sha256"
	"sync"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func testDirectory(t testing.TB, id byte) api.DirectoryIdentity {
	t.Helper()
	v, err := api.NewDirectoryIdentity(api.DirectoryUnix, 7, [16]byte{id}, "birth-test")
	return must(t, v, err)
}
func testScope(t testing.TB, key string, root api.DirectoryIdentity) api.WorktreeScope {
	t.Helper()
	r, err := domain.NewRepositoryID(domain.LocalCommon, "/test/common/.git")
	repo := must(t, r, err)
	w, err := domain.NewWorktreeID(repo, key)
	worktree := must(t, w, err)
	v, err := api.NewSourceVersion("git", "test-common", "test-observer", "observation1")
	source := must(t, v, err)
	s, err := api.NewWorktreeScope(api.WorktreeScopeData{ID: worktree, RootLocator: "/test/root", RootIdentity: root, Source: source})
	return must(t, s, err)
}

func TestWholeContentVersionsAndBoundAbsence(t *testing.T) {
	parent := testDirectory(t, 1)
	key, err := bindingToken(parent, nil, "run.json")
	store := must(t, key, err)
	scope := testScope(t, "main", parent)
	old := []byte(`{"schemaVersion":1,"default":"a","future":1}`)
	v, err := contentVersion(api.RunConfig, store, scope, true, old)
	original := must(t, v, err)
	if original.ByteLength() != uint64(len(old)) || original.SHA256() != sha256.Sum256(old) || !original.MatchesRunScope(scope) {
		t.Fatal("version omitted complete byte/scope binding")
	}
	for _, raw := range [][]byte{[]byte(`{"schemaVersion":1,"default":"a","future":2}`), append(append([]byte{}, old...), '\n')} {
		v, err := contentVersion(api.RunConfig, store, scope, true, raw)
		changed := must(t, v, err)
		if changed.Equal(original) || !changed.SameBinding(original) {
			t.Fatal("unknown/raw content drift not versioned")
		}
	}
	v, err = contentVersion(api.RunConfig, store, scope, true, append([]byte{}, old...))
	restored := must(t, v, err)
	if restored != original {
		t.Fatal("same-byte restoration incorrectly implies a content conflict")
	}
	v, err = contentVersion(api.RunConfig, store, scope, false, nil)
	absent := must(t, v, err)
	if absent.Present() || absent.ByteLength() != 0 || absent.SHA256() != ([32]byte{}) || absent.Equal(original) {
		t.Fatal("absence is not explicit")
	}
	if _, err := contentVersion(api.RunConfig, store, scope, false, []byte("x")); err == nil {
		t.Fatal("absence with bytes")
	}
	for _, otherScope := range []api.WorktreeScope{testScope(t, "linked", parent), testScope(t, "main", testDirectory(t, 2))} {
		v, err := contentVersion(api.RunConfig, store, otherScope, true, old)
		other := must(t, v, err)
		if other.SameBinding(original) || original.MatchesRunScope(otherScope) {
			t.Fatal("foreign worktree/root version accepted")
		}
	}
	for _, family := range []api.StorageFamily{api.UserConfig, api.Preferences} {
		v, err := contentVersion(family, store, api.WorktreeScope{}, true, old)
		foreign := must(t, v, err)
		if foreign.SameBinding(original) || foreign.Equal(original) {
			t.Fatal("foreign family version accepted")
		}
	}
	seen := map[string]bool{store: true}
	for _, remaining := range [][]string{{".gh-tree"}, {"a", "bc"}, {"ab", "c"}, {"a bc"}} {
		v, err := bindingToken(parent, remaining, "run.json")
		anchor := must(t, v, err)
		if seen[anchor] {
			t.Fatal("absence anchor/components aliased another binding")
		}
		seen[anchor] = true
		v, err = bindingToken(parent, remaining, "run.json")
		if err != nil || v != anchor {
			t.Fatal("binding not deterministic", err)
		}
	}
	other, err := bindingToken(testDirectory(t, 2), nil, "run.json")
	if err != nil || other == store {
		t.Fatal("replacement parent did not change binding", err)
	}
	other, err = bindingToken(parent, nil, "state.json")
	if err != nil || other == store {
		t.Fatal("fixed basename did not change binding", err)
	}
}

func TestNoMutableCodecStateAcrossRequests(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, fixture := range codecFixtures {
				d, err := decodeDocument(fixture.family, []byte(fixture.raw))
				if err != nil {
					t.Error(err)
					return
				}
				before := d.raw
				copy := promoted(t, d)
				raw, err := copy.encode()
				if err != nil {
					t.Error(err)
					return
				}
				if len(raw) > 0 {
					raw[0] = '!'
				}
				if d.raw != before {
					t.Error("output aliases stored original")
				}
			}
		}()
	}
	wg.Wait()
}

func TestRetainedObjectRemovalCannotEraseUnknowns(t *testing.T) {
	for _, fixture := range codecFixtures {
		v, err := decodeDocument(fixture.family, []byte(fixture.raw))
		original := must(t, v, err)
		d, err := emptyDocument(fixture.family)
		empty := must(t, d, err)
		if verifyRetention(original, empty) == nil {
			t.Fatal("parent removal erased retained values")
		}
	}
}
