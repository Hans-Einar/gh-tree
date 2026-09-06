package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const MaxDocumentBytes = 4 * 1024 * 1024
const MaxJSONDepth = 64

// OpaqueJSON retains an unknown JSON value. This predicate-only constructor does
// not decode a document, assign schema names, marshal, or certify a native load.
// Persistence still owns complete byte/schema/unknown-preservation validation.
type OpaqueJSON struct {
	raw   string
	depth uint8
}

func NewOpaqueJSON(raw []byte) (OpaqueJSON, error) {
	if len(raw) == 0 || len(raw) > MaxDocumentBytes || !utf8.Valid(raw) || !json.Valid(raw) {
		return OpaqueJSON{}, invalid("opaque JSON bytes")
	}
	depth, ok := jsonShape(string(raw))
	if !ok {
		return OpaqueJSON{}, invalid("JSON duplicate name/depth/string")
	}
	return OpaqueJSON{string(raw), uint8(depth)}, nil
}
func (v OpaqueJSON) Valid() bool             { _, e := NewOpaqueJSON([]byte(v.raw)); return e == nil }
func (v OpaqueJSON) Bytes() []byte           { return []byte(v.raw) }
func (v OpaqueJSON) Depth() uint8            { return v.depth }
func (v OpaqueJSON) Equal(w OpaqueJSON) bool { return v.raw == w.raw }

type JSONMember struct {
	name  string
	value OpaqueJSON
}

func NewJSONMember(name string, value OpaqueJSON) (JSONMember, error) {
	if !utf8.ValidString(name) || !value.Valid() {
		return JSONMember{}, invalid("JSON member")
	}
	return JSONMember{name, value}, nil
}
func (v JSONMember) Valid() bool       { return utf8.ValidString(v.name) && v.value.Valid() }
func (v JSONMember) Name() string      { return v.name }
func (v JSONMember) Value() OpaqueJSON { return v.value }

type JSONMembers struct {
	entries     []JSONMember
	initialized bool
}

func NewJSONMembers(entries []JSONMember) (JSONMembers, error) {
	seen := map[string]bool{}
	size := 2
	for _, m := range entries {
		if !m.Valid() || seen[m.name] {
			return JSONMembers{}, invalid("unknown member duplicate/value")
		}
		seen[m.name] = true
		size += quotedSize(m.name) + len(m.value.raw) + 2
	}
	if size > MaxDocumentBytes {
		return JSONMembers{}, invalid("unknown members size")
	}
	return JSONMembers{cloneSlice(entries), true}, nil
}
func (v JSONMembers) Valid() bool {
	if !v.initialized {
		return false
	}
	_, e := NewJSONMembers(v.entries)
	return e == nil
}
func (v JSONMembers) Entries() []JSONMember { return cloneSlice(v.entries) }
func (v JSONMembers) Excludes(names ...string) bool {
	for _, m := range v.entries {
		for _, n := range names {
			if m.name == n {
				return false
			}
		}
	}
	return true
}
func (v JSONMembers) WithinDepth(containingDepth int) bool {
	for _, m := range v.entries {
		if containingDepth+int(m.value.depth) > MaxJSONDepth {
			return false
		}
	}
	return true
}
func (v JSONMembers) Size() int {
	n := 0
	for _, m := range v.entries {
		n += quotedSize(m.name) + len(m.value.raw) + 2
	}
	return n
}

// jsonShape is a structural predicate over already syntactically valid JSON.
// It computes depth and rejects repeated decoded member names at every object.
// It produces no values or storage DTOs. Escaped key equivalence (including
// surrogate pairs) is checked so {"a":1,"\u0061":2} cannot evade uniqueness.
func jsonShape(s string) (int, bool) {
	type frame struct {
		object bool
		keys   map[string]bool
	}
	stack := []frame{}
	maxDepth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{', '[':
			stack = append(stack, frame{object: s[i] == '{', keys: map[string]bool{}})
			if len(stack) > maxDepth {
				maxDepth = len(stack)
			}
			if maxDepth > MaxJSONDepth {
				return 0, false
			}
		case '}', ']':
			stack = stack[:len(stack)-1]
		case '"':
			end := i + 1
			for end < len(s) {
				if s[end] == '\\' {
					end += 2
					continue
				}
				if s[end] == '"' {
					break
				}
				end++
			}
			key, ok := jsonStringIdentity(s[i+1 : end])
			if !ok {
				return 0, false
			}
			next := end + 1
			for next < len(s) && strings.ContainsRune(" \r\n\t", rune(s[next])) {
				next++
			}
			if next < len(s) && s[next] == ':' {
				f := &stack[len(stack)-1]
				if !f.object || f.keys[key] {
					return 0, false
				}
				f.keys[key] = true
			}
			i = end
		}
	}
	return maxDepth, true
}
func jsonStringIdentity(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case '"', '\\', '/':
			b.WriteByte(s[i])
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'u':
			n, e := strconv.ParseUint(s[i+1:i+5], 16, 16)
			if e != nil {
				return "", false
			}
			i += 4
			r := rune(n)
			if utf16.IsSurrogate(r) {
				if r >= 0xDC00 || i+6 >= len(s) || s[i+1:i+3] != "\\u" {
					return "", false
				}
				m, e := strconv.ParseUint(s[i+3:i+7], 16, 16)
				if e != nil || m < 0xDC00 || m > 0xDFFF {
					return "", false
				}
				r = utf16.DecodeRune(r, rune(m))
				i += 6
			}
			b.WriteRune(r)
		default:
			return "", false
		}
	}
	return b.String(), true
}
func quotedSize(s string) int {
	n := 2
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' || c == '\\':
			n += 2
		case c == '\b' || c == '\f' || c == '\n' || c == '\r' || c == '\t':
			n += 2
		case c < 32:
			n += 6
		default:
			n++
		}
	}
	return n
}
