package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// This supports the common Git text/autocrlf clean profiles without invoking
// Git's clean conversion engine, which could execute a repository-defined filter.
// Unsupported encodings, ident/filter/eol attributes are explicit refusals.
func (c *checker) prepareCleanPolicy() error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = c.root
	if _, err := cmd.Output(); err != nil {
		if _, statErr := os.Stat(c.root + "/.git"); statErr == nil {
			return fmt.Errorf("cannot read selected repository Git metadata: %w", err)
		}
		// Source archives have no clean metadata. Only their actual raw blobs
		// count; they receive no inferred Windows newline conversion.
		c.gitClean = false
		return nil
	}
	c.gitClean = true
	cmd = exec.Command("git", "config", "--get", "core.autocrlf")
	cmd.Dir = c.root
	b, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); !ok || e.ExitCode() != 1 {
			return fmt.Errorf("read core.autocrlf: %w", err)
		}
	}
	c.autoCRLF = strings.ToLower(strings.TrimSpace(string(b)))
	if c.autoCRLF != "" && c.autoCRLF != "false" && c.autoCRLF != "true" && c.autoCRLF != "input" {
		return fmt.Errorf("unsupported core.autocrlf profile %q", c.autoCRLF)
	}
	return nil
}

func (c *checker) cleanBlobHash(path string, b []byte) (string, error) {
	if !c.gitClean {
		return blobHash(b), nil
	}
	cmd := exec.Command("git", "check-attr", "-z", "text", "eol", "filter", "working-tree-encoding", "ident", "crlf", "--", path)
	cmd.Dir = c.root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("read path-specific clean attributes for %s: %w", path, err)
	}
	fields := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	if len(fields) != 18 {
		return "", fmt.Errorf("incomplete clean attribute metadata for %s", path)
	}
	attrs := map[string]string{}
	for i := 0; i < len(fields); i += 3 {
		attrs[fields[i+1]] = fields[i+2]
	}
	// The historical crlf attribute predates text and still changes Git's
	// conversion behavior. In particular -crlf is not an inert unset attribute.
	// Refuse every explicit legacy setting rather than guessing its precedence.
	if attrs["crlf"] != "unspecified" {
		return "", fmt.Errorf("unsupported Git legacy crlf profile for %s: crlf=%s", path, attrs["crlf"])
	}
	for _, attr := range []string{"filter", "working-tree-encoding", "ident", "eol"} {
		if value := attrs[attr]; value != "unspecified" && value != "unset" {
			return "", fmt.Errorf("unsupported Git clean profile for %s: %s=%s (no clean filter is executed)", path, attr, value)
		}
	}
	normalize, automatic := false, false
	switch attrs["text"] {
	case "set":
		normalize = true
	case "unset": // -text preserves actual bytes, even under core.autocrlf.
	case "auto":
		normalize, automatic = true, true
	case "unspecified":
		normalize = c.autoCRLF == "true" || c.autoCRLF == "input"
		automatic = normalize
	default:
		return "", fmt.Errorf("unsupported Git text profile for %s: %q", path, attrs["text"])
	}
	if normalize && bytes.Contains(b, []byte("\r\n")) {
		if automatic {
			// Git intentionally preserves an existing CRLF/binary index entry
			// under automatic conversion. Refuse that less common profile rather
			// than silently granting the LF baseline's allowance.
			cmd = exec.Command("git", "ls-files", "--eol", "--", path)
			cmd.Dir = c.root
			indexed, err := cmd.Output()
			if err != nil {
				return "", err
			}
			if f := strings.Fields(string(indexed)); len(f) > 0 && f[0] != "i/lf" && f[0] != "i/none" {
				return "", fmt.Errorf("unsupported automatic clean/index profile for %s: %s", path, f[0])
			}
			if bytes.IndexByte(b, 0) >= 0 {
				return "", fmt.Errorf("binary bytes cannot use a text legacy allowance: %s", path)
			}
		}
		b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	}
	return blobHash(b), nil
}
