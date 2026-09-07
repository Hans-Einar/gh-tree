package persistence

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

func must[T any](t testing.TB, value T, err error) T {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

var codecFixtures = []struct {
	name   string
	family api.StorageFamily
	raw    string
}{
	{"user", api.UserConfig, ` {
 "stripPrefixes":[],
 "repos":{"OWNER/repo":{"worktrees":{" A ":{"path":" /tmp/ space ","branch":"Dev:branch","futureTarget":{"integer":9007199254740993,"array":[true,null]}},"a":{"path":"","futureEmpty":[]}},"futureRepo":{"b":2,"a":1}},"owner/REPO":{"worktrees":{}}},
 "scopedRepos":{"remote:aG9zdC9vd25lci9yZXBv":{"worktrees":{"outside context":{"path":"C:\\ untrimmed ","branch":"main","futureScopedTarget":false}},"futureScoped":null}},
 "futureDocument":[{"exact":"\uD83D\uDE00","a":1e999999999999999999999999999999999999999999}],"":false
 } `},
	{"preferences", api.Preferences, `{
 "lastFolders":{"Owner/repo":" Feature/ ","owner/REPO":"other"},
 "lastWorktrees":{"Owner/repo":" /tmp/lost path ","owner/REPO":""},
 "scopedPreferences":{"remote:aG9zdC9vd25lci9yZXBv":{"lastFolder":"","futureRemote":[1,2]},"local-common:L3RtcC9jbG9uZS8uZ2l0":{"activeWorktree":{"administrativeKey":"linked:name","lastKnownPath":" /tmp/ exact ","futureActive":{"a":false}},"futureLocal":null}},
 "futureDocument":{"keep":true}
 }`},
	{"run", api.RunConfig, `{
 "default":" build ",
 "launch":{" build ":{"provider":"make","dir":" sub dir ","targets":["second","first","second"],"command":" /tools/custom-make ","futureMake":{"escaped":"\u0061","number":1.234567890123456789}},"Build":{"provider":"npm","script":" dev:wan ","dir":"","command":"custom-npm","futureNpm":null},"future":{"provider":"unknown-provider-v900","script":" literal\u0000\n bytes ","targets":[],"futureProvider":{"nested":[{"x":1}]}}},
 "futureDocument":true
 }`},
}

func promoted(t testing.TB, d document) document {
	t.Helper()
	switch d.family {
	case api.UserConfig:
		v := d.user.Data()
		v.SchemaVersion = 1
		x, err := api.NewUserConfigDocument(v)
		d.user = must(t, x, err)
	case api.Preferences:
		v := d.preferences.Data()
		v.SchemaVersion = 1
		x, err := api.NewPreferencesDocument(v)
		d.preferences = must(t, x, err)
	case api.RunConfig:
		v := d.run.Data()
		v.SchemaVersion = 1
		x, err := api.NewRunConfigDocument(v)
		d.run = must(t, x, err)
	}
	return d
}
func TestAllFamiliesRetainExactIntentAndOriginal(t *testing.T) {
	for _, fixture := range codecFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			input := []byte(fixture.raw)
			decoded, err := decodeDocument(fixture.family, input)
			original := must(t, decoded, err)
			if original.schema() != 0 || original.raw != fixture.raw {
				t.Fatal("legacy schema/raw bytes changed")
			}
			input[0] = '!'
			if original.raw != fixture.raw {
				t.Fatal("raw original aliases caller buffer")
			}
			if _, err := original.encode(); err == nil {
				t.Fatal("implicit legacy migration")
			}
			proposed := promoted(t, original)
			encoded, err := proposed.encode()
			raw := must(t, encoded, err)
			decoded, err = decodeDocument(fixture.family, raw)
			again := must(t, decoded, err)
			if again.schema() != 1 {
				t.Fatal("schema1 not emitted")
			}
			if !reflect.DeepEqual(proposed.user.Data(), again.user.Data()) || !reflect.DeepEqual(proposed.preferences.Data(), again.preferences.Data()) || !reflect.DeepEqual(proposed.run.Data(), again.run.Data()) {
				t.Fatal("known or retained document values changed")
			}
			if err := verifyRetention(original, again); err != nil {
				t.Fatal(err)
			}
			for _, m := range retainedMembers(original) {
				for _, entry := range m.Entries() {
					if !bytes.Contains(raw, entry.Value().Bytes()) {
						t.Fatalf("unknown raw value lost: %s", entry.Name())
					}
				}
			}
			if original.raw != fixture.raw {
				t.Fatal("encoding changed raw original")
			}
		})
	}
}

func TestKnownNullAndEmptyPresence(t *testing.T) {
	for _, tc := range []struct {
		raw      string
		presence api.FieldPresence
	}{{`{}`, api.FieldAbsent}, {`{"stripPrefixes":null}`, api.FieldNull}, {`{"stripPrefixes":[]}`, api.FieldPresent}} {
		v, err := decodeUserConfig([]byte(tc.raw))
		d := must(t, v, err)
		if d.Data().StripPrefixes.Presence() != tc.presence {
			t.Fatal("prefix presence changed")
		}
	}
	for _, family := range []api.StorageFamily{api.UserConfig, api.Preferences, api.RunConfig} {
		d, err := emptyDocument(family)
		value := must(t, d, err)
		if value.schema() != 1 {
			t.Fatal("absent candidate schema")
		}
		raw, err := value.encode()
		if err != nil || string(raw) != `{"schemaVersion":1}` {
			t.Fatal("empty candidate fabricated fields", err)
		}
	}
	knownNulls := map[api.StorageFamily][]string{
		api.UserConfig:  {`{"repos":null}`, `{"scopedRepos":null}`, `{"repos":{"x":null}}`, `{"repos":{"x":{"worktrees":null}}}`, `{"repos":{"x":{"worktrees":{"a":null}}}}`, `{"repos":{"x":{"worktrees":{"a":{"path":null}}}}}`, `{"repos":{"x":{"worktrees":{"a":{"branch":null}}}}}`},
		api.Preferences: {`{"lastFolders":null}`, `{"lastWorktrees":null}`, `{"lastFolders":{"x":null}}`, `{"scopedPreferences":null}`, `{"scopedPreferences":{"remote:eA":null}}`, `{"scopedPreferences":{"remote:eA":{"lastFolder":null}}}`, `{"scopedPreferences":{"local-common:eA":{"activeWorktree":null}}}`},
		api.RunConfig:   {`{"default":null}`, `{"launch":null}`, `{"launch":{"x":null}}`, `{"launch":{"x":{"provider":null}}}`, `{"launch":{"x":{"provider":"future","dir":null}}}`, `{"launch":{"x":{"provider":"future","script":null}}}`, `{"launch":{"x":{"provider":"future","command":null}}}`, `{"launch":{"x":{"provider":"future","targets":null}}}`, `{"launch":{"x":{"provider":"future","targets":[null]}}}`},
	}
	for family, fixtures := range knownNulls {
		for _, raw := range fixtures {
			if _, err := decodeDocument(family, []byte(raw)); err == nil {
				t.Errorf("accepted known null: %s", raw)
			}
		}
	}
	// Present nil slices are not null, including arrays nested beneath providers.
	d, err := decodeUserConfig([]byte(`{"schemaVersion":1}`))
	v := must(t, d, err).Data()
	v.StripPrefixes = api.PresentField[[]string](nil)
	d, err = api.NewUserConfigDocument(v)
	raw, err := encodeUserConfig(must(t, d, err))
	if err != nil || !bytes.Contains(raw, []byte(`"stripPrefixes":[]`)) {
		t.Fatal("present nil became null", err)
	}
}

func TestStrictEnvelopeAndReservedShapes(t *testing.T) {
	badCommon := []string{``, `null`, `[]`, `true`, `0`, `"x"`, `{} {}`, `{"x":0,"x":1}`, `{"x":{"a":1,"\u0061":2}}`, `{"x":[{"a":1,"a":2}]}`, `{"x":"\uD800"}`, `{"x":"\uDC00"}`, `{"x":"\uD800x"}`, "{\"x\":\"\xff\"}", `{"schemaVersion":-1}`, `{"schemaVersion":-0}`, `{"schemaVersion":1.0}`, `{"schemaVersion":1e0}`, `{"schemaVersion":"1"}`, `{"schemaVersion":null}`, `{"schemaVersion":true}`, `{"schemaVersion":1,"schema\u0056ersion":1}`}
	for _, family := range []api.StorageFamily{api.UserConfig, api.Preferences, api.RunConfig} {
		for _, raw := range badCommon {
			if _, err := decodeDocument(family, []byte(raw)); err == nil {
				t.Errorf("family %v accepted malformed input %q", family, raw)
			}
		}
	}
	bad := map[api.StorageFamily][]string{
		api.UserConfig:  {`{"stripPrefixes":[1]}`, `{"repos":[]}`, `{"repos":{"x":{"worktrees":{"a":[]}}}}`, `{"scopedRepos":{"Owner/repo":{}}}`, `{"scopedRepos":{"local-common:eA":{}}}`, `{"repos":{"":{"worktrees":{}}}}`, `{"repos":{"x":{"worktrees":{"":{}}}}}`},
		api.Preferences: {`{"lastFolders":[]}`, `{"lastFolders":{"x":{}}}`, `{"scopedPreferences":{"remote:eA":{"activeWorktree":{"administrativeKey":"a","lastKnownPath":"/p"}}}}`, `{"scopedPreferences":{"local-common:eA":{"lastFolder":"x"}}}`, `{"scopedPreferences":{"local-common:eA":{"activeWorktree":{"administrativeKey":"a","lastKnownPath":""}}}}`, `{"scopedPreferences":{"local-common:eA":{"activeWorktree":{"lastKnownPath":"/p"}}}}`},
		api.RunConfig:   {`{"launch":[]}`, `{"launch":{"x":{}}}`, `{"launch":{"":{"provider":"npm"}}}`, `{"launch":{"x":{"provider":""}}}`, `{"launch":{"x":{"provider":99}}}`, `{"launch":{"x":{"provider":"npm","targets":"a"}}}`, `{"launch":{"x":{"provider":"npm","script":{}}}}`},
	}
	for family, fixtures := range bad {
		for _, raw := range fixtures {
			if _, err := decodeDocument(family, []byte(raw)); err == nil {
				t.Errorf("accepted invalid shape %s", raw)
			}
		}
	}
}

func TestForwardVersionNeverWrapsOrCreatesDocument(t *testing.T) {
	for _, rawVersion := range []string{"2", "4294967295", "4294967296", "18446744073709551616", strings.Repeat("9", 1000)} {
		for _, family := range []api.StorageFamily{api.UserConfig, api.Preferences, api.RunConfig} {
			d, err := decodeDocument(family, []byte(`{"schemaVersion":`+rawVersion+`,"future":{"state":true}}`))
			var detail *codecError
			if !errors.As(err, &detail) || detail.state != api.UnsupportedVersion || detail.schemaRaw != rawVersion || d.family.Valid() {
				t.Fatal("lost unsupported raw version", rawVersion, err)
			}
			if len(rawVersion) > 10 && detail.schema.Present() {
				t.Fatal("overflow wrapped into uint32")
			}
		}
	}
}

func TestDepthAndInputBounds(t *testing.T) {
	for _, depth := range []int{63, 64, 65} {
		raw := `{"unknown":` + strings.Repeat("[", depth-1) + `0` + strings.Repeat("]", depth-1) + `}`
		_, err := decodeDocument(api.RunConfig, []byte(raw))
		if (depth <= 64) != (err == nil) {
			t.Fatalf("depth %d: %v", depth, err)
		}
	}
	raw := []byte("{}" + strings.Repeat(" ", api.MaxDocumentBytes-2))
	if _, err := decodeDocument(api.Preferences, raw); err != nil {
		t.Fatal("exact 4 MiB input refused", err)
	}
	if _, err := decodeDocument(api.Preferences, append(raw, ' ')); err == nil {
		t.Fatal("oversized input accepted")
	}
}

func TestCanonicalScopeSerialization(t *testing.T) {
	seen := map[string]domain.RepositoryID{}
	for _, scope := range []domain.RepositoryScope{domain.Remote, domain.LocalCommon} {
		for _, token := range []string{"host/o/r", "/a:b/c", " A ", "\x00\xff", "é", "local-common:eA"} {
			v, err := domain.NewRepositoryID(scope, token)
			id := must(t, v, err)
			key, err := scopeKey(id)
			if err != nil {
				t.Fatal(err)
			}
			if other, exists := seen[key]; exists && other != id {
				t.Fatal("scope serialization collision")
			}
			seen[key] = id
			v, err = parseScopeKey(key)
			if err != nil || v != id {
				t.Fatal("scope identity bytes changed", err)
			}
		}
	}
	for _, key := range []string{"x", "remote:", "local:", "REMOTE:eA", "remote:eA==", "remote:eB", "remote:e A", "remote:eA\n", "remote:eA:extra"} {
		if _, err := parseScopeKey(key); err == nil {
			t.Fatal("noncanonical scope accepted", key)
		}
	}
}

func TestRetentionChecksEveryObjectAndLosslessNumbers(t *testing.T) {
	for _, fixture := range codecFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			v, err := decodeDocument(fixture.family, []byte(fixture.raw))
			original := must(t, v, err)
			// Remove every unknown field independently, including those under known
			// repository/target/scoped/active/provider objects. All must refuse.
			var tree map[string]any
			d := json.NewDecoder(strings.NewReader(fixture.raw))
			d.UseNumber()
			if err := d.Decode(&tree); err != nil {
				t.Fatal(err)
			}
			for path, members := range retainedMembers(original) {
				for _, member := range members.Entries() {
					var components []string
					if err := json.Unmarshal([]byte(path), &components); err != nil {
						t.Fatal(err)
					}
					parent := tree
					for _, name := range components {
						parent = parent[name].(map[string]any)
					}
					previous := parent[member.Name()]
					delete(parent, member.Name())
					raw, err := json.Marshal(tree)
					if err != nil {
						t.Fatal(err)
					}
					changed, err := decodeDocument(fixture.family, raw)
					proposed := must(t, changed, err)
					if verifyRetention(original, proposed) == nil {
						t.Fatalf("lost retention under %s", path)
					}
					parent[member.Name()] = previous
				}
			}
			// Reserialization of objects/escapes keeps unknown values semantically
			// equal without float64 rounding enormous integers or exponents.
			raw, err := json.Marshal(tree)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := decodeDocument(fixture.family, raw)
			proposed := must(t, changed, err)
			if err := verifyRetention(original, proposed); err != nil {
				t.Fatal(err)
			}
		})
	}
	if sameJSONValue([]byte(`9007199254740992`), []byte(`9007199254740993`)) {
		t.Fatal("large integers rounded equal")
	}
	if sameJSONValue([]byte(`[1,2]`), []byte(`[2,1]`)) {
		t.Fatal("array order ignored")
	}
	if !sameJSONValue([]byte(`{"a":1,"b":"\u0061"}`), []byte(` { "b":"a", "a":1 } `)) {
		t.Fatal("JSON value normalization refused")
	}
	for _, pair := range [][2]string{{"1", "1.0"}, {"-0.0e999999999999999999999", "0"}, {"1000", "1e3"}, {"123.40e+0003", "123400"}, {"10e999999999999999999999", "1e1000000000000000000000"}, {"0.1e-999999999999999999999", "1e-1000000000000000000000"}, {"100e-1000000000", "1e-999999998"}, {"0.01e1000000000", "1e999999998"}} {
		if !sameJSONValue([]byte(pair[0]), []byte(pair[1])) {
			t.Fatal("equal exact decimals refused", pair)
		}
	}
}

func FuzzDocumentCodec(f *testing.F) {
	for _, fixture := range codecFixtures {
		f.Add(uint8(fixture.family), []byte(fixture.raw))
	}
	f.Add(uint8(api.UserConfig), []byte(`{"schemaVersion":1}`))
	f.Add(uint8(api.RunConfig), []byte(`{"x":"\ud800"}`))
	f.Fuzz(func(t *testing.T, familyByte uint8, raw []byte) {
		family := api.StorageFamily(familyByte)
		if !family.Valid() {
			return
		}
		d, err := decodeDocument(family, raw)
		if err != nil {
			return
		}
		d = promoted(t, d)
		encoded, err := d.encode()
		if err != nil {
			return
		}
		again, err := decodeDocument(family, encoded)
		if err != nil {
			t.Fatal("encoded invalid document", err)
		}
		if err := verifyRetention(d, again); err != nil {
			t.Fatal(err)
		}
	})
}
