// Package launchdiscovery implements passive, bounded launch-source observation.
package launchdiscovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

const npmProfile = "npm-colocated-locks-v1"
const makeProfile = "gnu-make-simple-text-v1"

type member struct {
	name  string
	valid bool
}
type parsed struct {
	members []member
	manager string
	notices []string
}

// strictValue checks every object, including unowned members, before decoding
// provider fields. encoding/json alone silently accepts duplicates and replaces
// malformed Unicode. The token walk is bounded and observes cancellation.
func strictValue(ctx context.Context, b []byte) error {
	if !utf8.Valid(b) {
		return errors.New("invalid UTF-8")
	}
	// OpaqueJSON is not used as a parser; validate JSON string escapes here so
	// unpaired UTF-16 surrogates cannot become replacement identity characters.
	for i := 0; i < len(b); i++ {
		if b[i] != '"' {
			continue
		}
		i++
		for ; i < len(b) && b[i] != '"'; i++ {
			if b[i] != '\\' {
				continue
			}
			i++
			if i >= len(b) {
				break
			}
			if b[i] != 'u' {
				continue
			}
			if i+4 >= len(b) {
				return errors.New("invalid Unicode escape")
			}
			var u uint16
			if _, e := fmt.Sscanf(string(b[i+1:i+5]), "%04x", &u); e != nil {
				return e
			}
			i += 4
			if u >= 0xDC00 && u <= 0xDFFF {
				return errors.New("unpaired surrogate")
			}
			if u >= 0xD800 && u <= 0xDBFF {
				if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
					return errors.New("unpaired surrogate")
				}
				var v uint16
				if _, e := fmt.Sscanf(string(b[i+3:i+7]), "%04x", &v); e != nil || v < 0xDC00 || v > 0xDFFF {
					return errors.New("unpaired surrogate")
				}
				i += 6
			}
		}
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > 64 {
			return errors.New("JSON depth limit")
		}
		t, err := d.Token()
		if err != nil {
			return err
		}
		c, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch c {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || seen[s] {
					return errors.New("duplicate or ambiguous JSON member")
				}
				seen[s] = true
				if e = value(depth + 1); e != nil {
					return e
				}
			}
		case '[':
			for d.More() {
				if e := value(depth + 1); e != nil {
					return e
				}
			}
		default:
			return errors.New("unexpected JSON delimiter")
		}
		_, err = d.Token()
		return err
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}

func parseNpm(ctx context.Context, b []byte) (parsed, error) {
	p := parsed{}
	if err := strictValue(ctx, b); err != nil {
		return p, err
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil || obj == nil {
		return p, errors.New("package manifest must be an object")
	}
	if raw, ok := obj["packageManager"]; ok {
		if err := json.Unmarshal(raw, &p.manager); err != nil || string(raw) == "null" {
			p.notices = append(p.notices, "invalid-package-manager")
		}
	}
	raw, ok := obj["scripts"]
	if !ok {
		return p, nil
	}
	var scripts map[string]json.RawMessage
	if err := json.Unmarshal(raw, &scripts); err != nil || scripts == nil {
		return p, errors.New("scripts must be an object")
	}
	for name, raw := range scripts {
		if err := ctx.Err(); err != nil {
			return p, err
		}
		var body string
		e := json.Unmarshal(raw, &body)
		valid := e == nil && string(raw) != "null" && body != "" && safeScript(name) && !strings.ContainsRune(body, 0)
		if name == "" {
			p.notices = append(p.notices, "invalid-script-member")
			continue
		}
		p.members = append(p.members, member{name, valid})
	}
	sort.Slice(p.members, func(i, j int) bool { return p.members[i].name < p.members[j].name })
	return p, nil
}

func safeScript(s string) bool {
	return s != "" && utf8.ValidString(s) && !strings.ContainsAny(s, "\x00\r\n") && !strings.HasPrefix(s, "-")
}
func safeTarget(s string) bool {
	if s == "" || s[0] == '-' || s[0] == '.' {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

func parseMake(ctx context.Context, b []byte, maxLine int) (parsed, error) {
	p := parsed{}
	if !utf8.Valid(b) || bytes.IndexByte(b, 0) >= 0 {
		return p, errors.New("invalid Make text")
	}
	seen := map[string]bool{}
	limited := false
	continuing := false
	for len(b) > 0 {
		if err := ctx.Err(); err != nil {
			return p, err
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			i = len(b)
		}
		line := string(b[:i])
		hasNewline := i < len(b)
		if hasNewline {
			b = b[i+1:]
		} else {
			b = nil
		}
		if len(line) > maxLine {
			return p, fmt.Errorf("%w: Make line limit", errLimit)
		}
		line = strings.TrimSuffix(line, "\r")
		// Continuations apply even to comment and recipe lines. Refuse the whole
		// logical line, including its tails, rather than discovering a tail as a
		// separate rule. Escaped backslashes and spaces after a backslash do not
		// continue the physical line. No joining or Make evaluation occurs here.
		continued := hasNewline && makeContinues(line)
		if continuing || continued {
			limited = true
			continuing = continued
			continue
		}
		if strings.HasPrefix(line, "\t") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			// Any escape before this marker is outside the simple profile, so an
			// escaped # cannot be mistaken for an ordinary comment delimiter.
			if strings.ContainsRune(line[:i], '\\') {
				limited = true
				continue
			}
			line = line[:i]
		}
		// Comment prose has no expansion/pattern meaning. Actual rule escapes,
		// directives and expansion remain explicit profile limitations.
		if strings.ContainsAny(line, "$%\\") {
			limited = true
			continue
		}
		i = strings.IndexByte(line, ':')
		if i < 0 {
			if strings.Contains(line, "=") {
				continue
			}
			limited = true
			continue
		}
		lhs, rhs := line[:i], line[i+1:]
		if strings.ContainsAny(lhs, "=;") || strings.HasPrefix(rhs, "=") {
			limited = true
			continue
		}
		if strings.ContainsAny(rhs, ";:") {
			limited = true
			continue
		}
		for _, name := range strings.Fields(lhs) {
			if !safeTarget(name) {
				limited = true
				continue
			}
			if !seen[name] {
				seen[name] = true
				p.members = append(p.members, member{name, true})
			}
		}
	}
	if limited {
		p.notices = append(p.notices, "make-profile-limitation")
	}
	sort.Slice(p.members, func(i, j int) bool { return p.members[i].name < p.members[j].name })
	return p, nil
}

func makeContinues(line string) bool {
	n := 0
	for i := len(line) - 1; i >= 0 && line[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}
