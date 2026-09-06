package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

// A document is request-private, contains exactly one family, and retains the
// complete observed bytes independently of later re-encoding or caller buffers.
// Neither it nor a content version carries a native handle across port calls.
type document struct {
	family      api.StorageFamily
	raw         string
	user        api.UserConfigDocument
	preferences api.PreferencesDocument
	run         api.RunConfigDocument
}

func decodeDocument(family api.StorageFamily, raw []byte) (document, error) {
	d := document{family: family, raw: string(raw)}
	var err error
	switch family {
	case api.UserConfig:
		d.user, err = decodeUserConfig(raw)
	case api.Preferences:
		d.preferences, err = decodePreferences(raw)
	case api.RunConfig:
		d.run, err = decodeRunConfig(raw)
	default:
		return document{}, errors.New("invalid storage family")
	}
	if err != nil {
		var detail *codecError
		if errors.As(err, &detail) {
			return document{}, err
		}
		return document{}, corrupt("invalid known document shape: " + err.Error())
	}
	return d, nil
}
func (d document) encode() ([]byte, error) {
	switch d.family {
	case api.UserConfig:
		return encodeUserConfig(d.user)
	case api.Preferences:
		return encodePreferences(d.preferences)
	case api.RunConfig:
		return encodeRunConfig(d.run)
	default:
		return nil, errors.New("invalid storage family")
	}
}
func (d document) schema() uint32 {
	switch d.family {
	case api.UserConfig:
		return d.user.Data().SchemaVersion
	case api.Preferences:
		return d.preferences.Data().SchemaVersion
	case api.RunConfig:
		return d.run.Data().SchemaVersion
	default:
		return 0
	}
}
func emptyDocument(family api.StorageFamily) (document, error) {
	return decodeDocument(family, []byte(`{"schemaVersion":1}`))
}

// contentVersion accepts an already acquired binding. It never decodes a token
// into a filename and is not a substitute for native acquisition/revalidation.
func contentVersion(family api.StorageFamily, store string, scope api.WorktreeScope, present bool, raw []byte) (api.StorageVersion, error) {
	var digest [32]byte
	var length uint64
	if present {
		digest = sha256.Sum256(raw)
		length = uint64(len(raw))
	} else if len(raw) != 0 {
		return api.StorageVersion{}, errors.New("absence cannot contain bytes")
	}
	if family == api.RunConfig {
		return api.NewRunStorageVersion(scope, store, present, length, digest)
	}
	return api.NewStorageVersion(family, store, present, length, digest)
}

// bindingToken includes the actual parent, or the nearest existing anchor and
// every still-missing literal directory component. The basename is independent
// of the parent chain. Its hash is an equality value, never path authority.
func bindingToken(parent api.DirectoryIdentity, remaining []string, basename string) (string, error) {
	if !parent.Valid() || !singleName(basename) {
		return "", errors.New("invalid physical binding")
	}
	for _, name := range remaining {
		if !singleName(name) {
			return "", errors.New("invalid missing component")
		}
	}
	fileID := parent.FileID()
	parts := []string{"gh-tree-storage-binding-v1", fmt.Sprint(parent.Platform()), fmt.Sprint(parent.Device()), hex.EncodeToString(fileID[:]), parent.Stamp()}
	parts = append(parts, remaining...)
	parts = append(parts, "", basename)
	raw, err := json.Marshal(parts)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return "storage-v1:" + hex.EncodeToString(digest[:]), nil
}
func singleName(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}
	for _, c := range []byte(s) {
		if c == 0 || c == '/' || runtime.GOOS == "windows" && (c == '\\' || c == ':') {
			return false
		}
	}
	return true
}

// Retention is checked against the guarded current document at commit. Known
// Application intent is not merged here. Removing an enclosing known object
// cannot silently discard the unknown members retained beneath that object.
func retainedMembers(d document) map[string]api.JSONMembers {
	result := map[string]api.JSONMembers{}
	add := func(path []string, members api.JSONMembers) {
		key, _ := json.Marshal(path)
		result[string(key)] = members
	}
	configured := func(prefix []string, c api.ConfiguredRepository) {
		v := c.Data()
		add(prefix, v.UnknownMembers)
		if items, present := v.Worktrees.Value(); present {
			for _, item := range items {
				v := item.Data()
				path := append(append([]string{}, prefix...), "worktrees", v.Name)
				add(path, v.UnknownMembers)
			}
		}
	}
	switch d.family {
	case api.UserConfig:
		v := d.user.Data()
		add(nil, v.UnknownMembers)
		if items, present := v.LegacyRepos.Value(); present {
			for _, item := range items {
				v := item.Data()
				configured([]string{"repos", v.Key}, v.Value)
			}
		}
		if items, present := v.ScopedRepos.Value(); present {
			for _, item := range items {
				v := item.Data()
				key, _ := scopeKey(v.RepositoryID)
				configured([]string{"scopedRepos", key}, v.Value)
			}
		}
	case api.Preferences:
		v := d.preferences.Data()
		add(nil, v.UnknownMembers)
		if items, present := v.ScopedPreferences.Value(); present {
			for _, item := range items {
				v := item.Data()
				key, _ := scopeKey(v.RepositoryID)
				prefix := []string{"scopedPreferences", key}
				add(prefix, v.UnknownMembers)
				if active, present := v.ActiveWorktree.Value(); present {
					add(append(prefix, "activeWorktree"), active.Data().UnknownMembers)
				}
			}
		}
	case api.RunConfig:
		v := d.run.Data()
		add(nil, v.UnknownMembers)
		if items, present := v.Launch.Value(); present {
			for _, item := range items {
				v := item.Data()
				add([]string{"launch", v.Alias}, v.Definition.Data().UnknownMembers)
			}
		}
	}
	return result
}
func verifyRetention(original, proposed document) error {
	if original.family != proposed.family {
		return errors.New("retention family mismatch")
	}
	newObjects := retainedMembers(proposed)
	for path, old := range retainedMembers(original) {
		newValues := map[string]api.OpaqueJSON{}
		for _, m := range newObjects[path].Entries() {
			newValues[m.Name()] = m.Value()
		}
		for _, m := range old.Entries() {
			other, ok := newValues[m.Name()]
			if !ok || !sameJSONValue(m.Value().Bytes(), other.Bytes()) {
				return errors.New("proposed document discards or changes retained unknown members")
			}
		}
	}
	return nil
}

// No float conversion: arbitrary JSON numbers compare by their exact decimal
// value. Object ordering, string escaping and whitespace may normalize; arrays
// and all primitive values retain their exact meaning.
func sameJSONValue(a, b []byte) bool {
	canonical := func(raw []byte) ([]byte, error) {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		return json.Marshal(normalizeJSONNumbers(value))
	}
	x, err := canonical(a)
	if err != nil {
		return false
	}
	y, err := canonical(b)
	return err == nil && bytes.Equal(x, y)
}

func normalizeJSONNumbers(v any) any {
	switch x := v.(type) {
	case json.Number:
		return json.Number(canonicalNumber(string(x)))
	case []any:
		for i := range x {
			x[i] = normalizeJSONNumbers(x[i])
		}
		return x
	case map[string]any:
		for key, value := range x {
			x[key] = normalizeJSONNumbers(value)
		}
		return x
	default:
		return v
	}
}

func canonicalNumber(s string) string {
	sign := ""
	if s[0] == '-' {
		sign = "-"
		s = s[1:]
	}
	exponent := "0"
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exponent = s[i+1:]
		s = s[:i]
	}
	adjust := 0
	if i := strings.IndexByte(s, '.'); i >= 0 {
		adjust -= len(s) - i - 1
		s = s[:i] + s[i+1:]
	}
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	trimmed := strings.TrimRight(s, "0")
	adjust += len(s) - len(trimmed)
	return sign + trimmed + "e" + addExponent(exponent, adjust)
}

// JSON permits exponents much larger than machine integers. Decimal addition
// of this bounded coefficient adjustment is linear, even for a 4 MiB exponent;
// neither enormous exponentiation nor a float conversion is needed.
func addExponent(s string, adjustment int) string {
	negative := false
	if s[0] == '-' || s[0] == '+' {
		negative = s[0] == '-'
		s = s[1:]
	}
	s = strings.TrimLeft(s, "0")
	if len(s) <= 8 {
		n := 0
		if s != "" {
			n, _ = strconv.Atoi(s)
		}
		if negative {
			n = -n
		}
		return strconv.Itoa(n + adjustment)
	}
	if negative {
		adjustment = -adjustment
	}
	out := []byte(s)
	if adjustment >= 0 {
		carry := adjustment
		for i := len(out) - 1; i >= 0 && carry > 0; i-- {
			n := int(out[i]-'0') + carry
			out[i] = byte(n%10) + '0'
			carry = n / 10
		}
		if carry > 0 {
			out = append([]byte(strconv.Itoa(carry)), out...)
		}
	} else {
		borrow := -adjustment
		for i := len(out) - 1; i >= 0 && borrow > 0; i-- {
			n := int(out[i]-'0') - borrow%10
			borrow /= 10
			if n < 0 {
				n += 10
				borrow++
			}
			out[i] = byte(n) + '0'
		}
	}
	result := strings.TrimLeft(string(out), "0")
	if negative {
		result = "-" + result
	}
	return result
}
