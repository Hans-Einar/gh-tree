package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/scanner"
)

func assemblyIncludes(b []byte) ([]string, error) {
	var s scanner.Scanner
	s.Init(bytes.NewReader(b))
	s.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments | scanner.SkipComments
	var scanErr error
	s.Error = func(_ *scanner.Scanner, msg string) { scanErr = fmt.Errorf("assembly include scan: %s", msg) }
	var names []string
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if tok != '#' {
			continue
		}
		if s.Scan() != scanner.Ident || s.TokenText() != "include" {
			continue
		}
		if s.Scan() != scanner.String {
			return nil, fmt.Errorf("assembly include requires literal filename")
		}
		name, err := strconv.Unquote(s.TokenText())
		if err != nil || !filepath.IsLocal(name) || strings.ContainsAny(name, "\\:\r\n\x00") {
			return nil, fmt.Errorf("assembly include outside declared closure: %q", name)
		}
		for _, part := range strings.Split(name, "/") {
			if part == ".." {
				return nil, fmt.Errorf("assembly include outside declared closure: %q", name)
			}
		}
		names = append(names, name)
	}
	return names, scanErr
}

func captureIncludes(q pkg, root string, p *plan, selected, included map[string]bool, add func(string, string, string, bool) error) error {
	var scope, prefix string
	switch {
	case q.Standard:
		scope, prefix = filepath.Join(runtime.GOROOT(), "src"), "toolchain/src/"
	case q.Module != nil && q.Module.Main:
		scope, prefix = root, "repo/"
	case q.Module != nil:
		scope, prefix = q.Module.Dir, "module/"+q.Module.Path+"@"+q.Module.Version+"/"
	default:
		return fmt.Errorf("unattributed assembly package: %s", q.ImportPath)
	}
	seen := map[string]bool{}
	var visit func(string, string, string) error
	visit = func(path, base, keyPrefix string) error {
		rel, err := contained(base, path)
		if err != nil {
			return err
		}
		key := keyPrefix + rel
		if seen[key] {
			return nil
		}
		seen[key] = true
		repoPath := ""
		if keyPrefix == "repo/" {
			repoPath = rel
		}
		if err := add(key, path, repoPath, repoPath != ""); err != nil {
			return err
		}
		selected[key] = true
		included[key] = true
		names, err := assemblyIncludes(p.files[key].bytes)
		if err != nil {
			return fmt.Errorf("selected assembly/header %q: %w", path, err)
		}
		for _, name := range names {
			// Go's assembler tries package cwd, source dirname, generated work
			// headers and GOROOT/pkg/include, in that order. All lexical include
			// paths stay within those captured roots; absolute/traversal refuse.
			found := false
			for _, dir := range []string{q.Dir, filepath.Dir(path)} {
				candidate := filepath.Join(dir, filepath.FromSlash(name))
				if _, err := os.Stat(candidate); err == nil {
					if err := visit(candidate, base, keyPrefix); err != nil {
						return err
					}
					found = true
					break
				} else if !os.IsNotExist(err) {
					return err
				}
			}
			if found || name == "go_asm.h" {
				continue
			} // compiler-derived from captured Go sources
			inc := filepath.Join(runtime.GOROOT(), "pkg", "include")
			if err := visit(filepath.Join(inc, filepath.FromSlash(name)), inc, "toolchain/pkg/include/"); err != nil {
				return fmt.Errorf("resolve include %q in %q: %w", name, path, err)
			}
		}
		return nil
	}
	for _, name := range q.SFiles {
		if err := visit(filepath.Join(q.Dir, name), scope, prefix); err != nil {
			return err
		}
	}
	return nil
}
