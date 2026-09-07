// Package persistence owns the private storage schema and native publication
// mechanics. Decoding a stored identity never observes or authorizes a repository.
package persistence

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/domain"
)

// codecError keeps unsupported integer text without narrowing it to uint32.
// Native loads attach the actual independently observed version and raw backup.
type codecError struct {
	state     api.StorageLoadState
	reason    string
	schema    api.Optional[uint32]
	schemaRaw string
}

func (e *codecError) Error() string { return "storage codec: " + e.reason }
func corrupt(reason string) error   { return &codecError{state: api.Corrupt, reason: reason} }

type jsonMember struct {
	name string
	raw  []byte
}
type jsonObject []jsonMember

// documentObject applies the whole-input bound and strict predicate before any
// decoder can replace invalid UTF-8/surrogates or choose a duplicate-key winner.
func documentObject(raw []byte) (jsonObject, uint32, error) {
	if len(raw) > api.MaxDocumentBytes {
		return nil, 0, corrupt("document exceeds 4 MiB")
	}
	if _, err := api.NewOpaqueJSON(raw); err != nil {
		return nil, 0, corrupt("invalid JSON, UTF-8, duplicate member, or depth above 64")
	}
	o, err := object(raw)
	if err != nil {
		return nil, 0, err
	}
	v, present := o.take("schemaVersion")
	if !present {
		return o, 0, nil
	}
	s := string(bytes.TrimSpace(v))
	if s == "" || strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return nil, 0, corrupt("schemaVersion must be a nonnegative integer")
	}
	if s == "0" || s == "1" {
		return o, uint32(s[0] - '0'), nil
	}
	e := &codecError{state: api.UnsupportedVersion, reason: "unsupported schemaVersion " + s, schemaRaw: s}
	if n, err := strconv.ParseUint(s, 10, 32); err == nil {
		e.schema = api.Some(uint32(n))
	}
	return nil, 0, e
}

// object is private to already strict, bounded input. It preserves member order
// and each raw value; map decoding would lose ordered unknown-member retention.
func object(raw []byte) (jsonObject, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, corrupt("expected object")
	}
	o := jsonObject{}
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return nil, corrupt("invalid object member")
		}
		name, ok := t.(string)
		if !ok {
			return nil, corrupt("invalid object name")
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return nil, corrupt("invalid object value")
		}
		o = append(o, jsonMember{name, value})
	}
	if _, err := d.Token(); err != nil {
		return nil, corrupt("invalid object end")
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, corrupt("trailing JSON")
	}
	return o, nil
}

func (o *jsonObject) take(name string) ([]byte, bool) {
	for i, m := range *o {
		if m.name == name {
			*o = append((*o)[:i], (*o)[i+1:]...)
			return m.raw, true
		}
	}
	return nil, false
}
func (o jsonObject) unknown() (api.JSONMembers, error) {
	entries := make([]api.JSONMember, 0, len(o))
	for _, m := range o {
		v, err := api.NewOpaqueJSON(m.raw)
		if err != nil {
			return api.JSONMembers{}, err
		}
		entry, err := api.NewJSONMember(m.name, v)
		if err != nil {
			return api.JSONMembers{}, err
		}
		entries = append(entries, entry)
	}
	return api.NewJSONMembers(entries)
}

func textValue(raw []byte) (string, error) {
	if len(raw) == 0 || raw[0] != '"' {
		return "", corrupt("expected string")
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", corrupt("invalid string")
	}
	return s, nil
}
func stringsValue(raw []byte) ([]string, error) {
	if len(raw) == 0 || raw[0] != '[' {
		return nil, corrupt("expected string array")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, corrupt("invalid string array")
	}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		v, err := textValue(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
func requiredText(o *jsonObject, name string) (string, error) {
	raw, present := o.take(name)
	if !present {
		return "", corrupt("missing required " + name)
	}
	return textValue(raw)
}
func field[T any](o *jsonObject, name string, decode func([]byte) (T, error)) (api.StoredField[T], error) {
	raw, present := o.take(name)
	if !present {
		return api.AbsentField[T](), nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return api.StoredField[T]{}, corrupt("null " + name)
	}
	v, err := decode(raw)
	if err != nil {
		return api.StoredField[T]{}, err
	}
	return api.PresentField(v), nil
}
func entries[T any](decode func(string, []byte) (T, error)) func([]byte) ([]T, error) {
	return func(raw []byte) ([]T, error) {
		o, err := object(raw)
		if err != nil {
			return nil, err
		}
		result := make([]T, 0, len(o))
		for _, m := range o {
			v, err := decode(m.name, m.raw)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
		return result, nil
	}
}

// Scope keys encode namespace plus opaque token bytes, including non-UTF-8 local
// tokens. Canonical unpadded base64url prevents alternate textual encodings from
// selecting the same semantic map entry. This is a codec, never path resolution.
func scopeKey(id domain.RepositoryID) (string, error) {
	if !id.Valid() {
		return "", corrupt("invalid scope")
	}
	prefix := "remote:"
	if id.Scope() == domain.LocalCommon {
		prefix = "local-common:"
	}
	return prefix + base64.RawURLEncoding.EncodeToString([]byte(id.Token())), nil
}
func parseScopeKey(key string) (domain.RepositoryID, error) {
	prefix, suffix, ok := strings.Cut(key, ":")
	var scope domain.RepositoryScope
	switch prefix {
	case "remote":
		scope = domain.Remote
	case "local-common":
		scope = domain.LocalCommon
	default:
		return domain.RepositoryID{}, corrupt("invalid canonical scope namespace")
	}
	if !ok {
		return domain.RepositoryID{}, corrupt("missing scope token")
	}
	token, err := base64.RawURLEncoding.Strict().DecodeString(suffix)
	if err != nil || len(token) == 0 || base64.RawURLEncoding.EncodeToString(token) != suffix {
		return domain.RepositoryID{}, corrupt("invalid canonical scope token")
	}
	return domain.NewRepositoryID(scope, string(token))
}

type objectWriter struct {
	members jsonObject
	err     error
}

func (w *objectWriter) add(name string, raw []byte, err error) {
	if w.err == nil {
		w.err = err
	}
	w.members = append(w.members, jsonMember{name, raw})
}
func (w *objectWriter) retained(m api.JSONMembers) {
	if !m.Valid() {
		w.err = errors.New("invalid retained members")
		return
	}
	for _, entry := range m.Entries() {
		w.add(entry.Name(), entry.Value().Bytes(), nil)
	}
}
func jsonValue[T any](v T) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte{'\n'}), nil
}
func put[T any](w *objectWriter, name string, f api.StoredField[T], encode func(T) ([]byte, error)) {
	switch f.Presence() {
	case api.FieldAbsent:
		return
	case api.FieldNull:
		w.add(name, []byte("null"), nil)
	case api.FieldPresent:
		v, _ := f.Value()
		raw, err := encode(v)
		w.add(name, raw, err)
	default:
		w.err = errors.New("invalid field presence")
	}
}
func (w *objectWriter) finish() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	var b bytes.Buffer
	b.WriteByte('{')
	seen := map[string]bool{}
	for i, m := range w.members {
		if seen[m.name] {
			return nil, corrupt("duplicate output member")
		}
		seen[m.name] = true
		if i > 0 {
			b.WriteByte(',')
		}
		key, err := jsonValue(m.name)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')
		b.Write(m.raw)
		if b.Len() > api.MaxDocumentBytes {
			return nil, corrupt("encoded document exceeds 4 MiB")
		}
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}
func mapValue[T any](values []T, encode func(T) (string, []byte, error)) ([]byte, error) {
	w := objectWriter{}
	for _, v := range values {
		name, raw, err := encode(v)
		w.add(name, raw, err)
	}
	return w.finish()
}
func checkedDocument(raw []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	if len(raw) > api.MaxDocumentBytes {
		return nil, corrupt("encoded document exceeds 4 MiB")
	}
	if _, err := api.NewOpaqueJSON(raw); err != nil {
		return nil, fmt.Errorf("invalid encoded document: %w", err)
	}
	return raw, nil
}
