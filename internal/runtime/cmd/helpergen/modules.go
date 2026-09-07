package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func pinnedSum(p plan, path, version string) (string, error) {
	var sum string
	for _, line := range strings.Split(string(p.files["repo/go.sum"].bytes), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == path && fields[1] == version {
			if sum != "" && sum != fields[2] {
				return "", fmt.Errorf("conflicting checksum pins for %s@%s", path, version)
			}
			sum = fields[2]
		}
	}
	if !strings.HasPrefix(sum, "h1:") {
		return "", fmt.Errorf("missing h1 pin for %s@%s", path, version)
	}
	return sum, nil
}

func sumSummary(rows []string) string {
	// Sort by filename, not the file hash at the beginning of a summary row.
	sort.Slice(rows, func(i, j int) bool { return rows[i][66:] < rows[j][66:] })
	h := sha256.New()
	for _, row := range rows {
		h.Write([]byte(row))
	}
	return "h1:" + base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Bind captured bytes to go.sum using Go's documented h1 directory algorithm.
// A cached ziphash or an earlier go mod verify is not authority for bytes read
// later. This independently hashes the full selected module and checks that
// every selected byte buffer is exactly the corresponding authenticated content.
func verifyCapturedModule(m listedModule, p *plan) error {
	expected, err := pinnedSum(*p, m.Path, m.Version)
	if err != nil {
		return err
	}
	modExpected, err := pinnedSum(*p, m.Path, m.Version+"/go.mod")
	if err != nil {
		return err
	}
	prefix := "module/" + m.Path + "@" + m.Version + "/"
	mod := p.files[prefix+"go.mod"].bytes
	if expected != m.Sum || modExpected != m.GoModSum || sumSummary([]string{hash(mod) + "  go.mod\n"}) != modExpected {
		return fmt.Errorf("module metadata checksum mismatch: %s@%s", m.Path, m.Version)
	}
	var rows []string
	seen := map[string]bool{}
	err = filepath.WalkDir(m.Dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("redirected module input: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular module input: %s", path)
		}
		rel, err := contained(m.Dir, path)
		if err != nil {
			return err
		}
		if strings.ContainsAny(rel, "\r\n") {
			return fmt.Errorf("unsupported module filename %q", path)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if f, ok := p.files[prefix+rel]; ok && !bytes.Equal(f.bytes, b) {
			return fmt.Errorf("selected module input changed during capture: %s", path)
		}
		seen[rel] = true
		rows = append(rows, hash(b)+"  "+m.Path+"@"+m.Version+"/"+rel+"\n")
		return nil
	})
	if err != nil {
		return err
	}
	if sumSummary(rows) != expected {
		return fmt.Errorf("module source checksum mismatch: %s@%s", m.Path, m.Version)
	}
	for key := range p.files {
		if strings.HasPrefix(key, prefix) && key != prefix+"go.mod" && !seen[strings.TrimPrefix(key, prefix)] {
			return fmt.Errorf("unverified selected module input: %s", key)
		}
	}
	// Go's offline module loader needs only these derived records alongside the
	// authenticated selected source tree. They introduce no mutable cache input.
	for ext, b := range map[string][]byte{"mod": mod, "ziphash": []byte(expected), "info": jsonBytes(struct{ Version string }{m.Version})} {
		key := "modulemeta/" + escapeModule(m.Path) + "/@v/" + escapeModule(m.Version) + "." + ext
		p.files[key] = captured{source: source{key, hash(b), len(b)}, bytes: append([]byte(nil), b...)}
	}
	return nil
}

// The Go module-cache encoding escapes ASCII uppercase letters with '!'.
func escapeModule(s string) string {
	var out strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			out.WriteByte('!')
			out.WriteRune(r + 'a' - 'A')
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}
