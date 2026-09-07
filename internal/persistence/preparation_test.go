package persistence

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type shortWriter struct{ calls int }

func (w *shortWriter) Write(p []byte) (int, error) {
	w.calls++
	return len(p) - 1, nil
}

func TestPreparationChecksWriteLengthAndCancellation(t *testing.T) {
	w := &shortWriter{}
	if err := writeComplete(context.Background(), w, []byte("private")); !errors.Is(err, io.ErrShortWrite) || w.calls != 1 {
		t.Fatalf("short write: %v, calls %d", err, w.calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := writeComplete(ctx, w, []byte("private")); !errors.Is(err, context.Canceled) || w.calls != 1 {
		t.Fatalf("canceled write: %v, calls %d", err, w.calls)
	}
	raw := bytes.Repeat([]byte("payload"), 20000)
	var complete bytes.Buffer
	if err := writeComplete(context.Background(), &complete, raw); err != nil || !bytes.Equal(raw, complete.Bytes()) {
		t.Fatalf("complete multi-chunk write: %v", err)
	}
}

func TestPreparationRecoveryNamespaceMatchesNativeEquality(t *testing.T) {
	for _, pair := range [][2]string{{"Config.JSON", "config.json"}, {"\u03c2.json", "\u03c3.json"}, {"\u212a.json", "K.json"}} {
		if (recoveryPrefix(pair[0]) == recoveryPrefix(pair[1])) != nativeSameName(pair[0], pair[1]) {
			t.Fatalf("namespace and native equality disagree for %q", pair)
		}
	}
	if len(recoveryPrefix(strings.Repeat("x", 230))) != 42 {
		t.Fatal("recovery prefix grows with document name")
	}
}

func TestPreparationExclusiveCreationAndNativeMetadata(t *testing.T) {
	root := t.TempDir()
	// Constructor tests already select a physical root on macOS /var aliases.
	root, err := nativeResolveExplicit(root)
	if err != nil {
		t.Fatal(err)
	}
	c, err := nativeAcquire(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.close(); err != nil {
			t.Error(err)
		}
	}()
	if err := nativeInspectDirectory(c.parent()); err != nil {
		t.Fatal(err)
	}
	child, err := nativeCreateDirectory(c.parent(), "owned", true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := child.close(); err != nil {
			t.Error(err)
		}
	}()
	if duplicate, err := nativeCreateDirectory(c.parent(), "owned", true); err == nil {
		duplicate.close()
		t.Fatal("directory creation reused existing entry")
	}
	file, err := nativeCreateFile(child, "payload", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeComplete(context.Background(), file.file, []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := file.file.Sync(); err != nil {
		t.Fatal(err)
	}
	if _, err := nativeInspectMetadata(file); err != nil {
		t.Fatal(err)
	}
	if err := file.close(); err != nil {
		t.Fatal(err)
	}
	if duplicate, err := nativeCreateFile(child, "payload", false); err == nil {
		duplicate.close()
		t.Fatal("file creation reused existing entry")
	}
	if got, err := os.ReadFile(filepath.Join(root, "owned", "payload")); err != nil || string(got) != "original" {
		t.Fatalf("exclusive target changed: %q, %v", got, err)
	}
}

func TestPreparationInventoryCountsActualCrashResidue(t *testing.T) {
	for _, tc := range []struct {
		name      string
		maxRecord int
		maxBytes  int64
		wantError bool
	}{
		{"fits", 3, 200, false},
		{"actual-byte-limit", 3, 90, true},
		{"orphan-record-limit", 1, 200, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, err := nativeResolveExplicit(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			prefix := recoveryPrefix("config.json")
			nonce, err := operationNonce()
			if err != nil {
				t.Fatal(err)
			}
			for name, raw := range map[string]string{
				prefix + nonce + ".manifest": "header",
				prefix + nonce + ".payload":  strings.Repeat("x", 100),
				prefix + "partial":           "incomplete writer",
				"unrelated":                  strings.Repeat("u", 1000),
			} {
				if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0600); err != nil {
					t.Fatal(err)
				}
			}
			c, err := nativeAcquire(context.Background(), root)
			if err != nil {
				t.Fatal(err)
			}
			names, records, size, err := inventoryRecovery(context.Background(), c.parent(), "config.json", tc.maxRecord, tc.maxBytes)
			if closeErr := c.close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			if tc.wantError {
				if !errors.Is(err, errRecoveryCapacity) {
					t.Fatalf("expected capacity refusal, got %v", err)
				}
			} else if err != nil || len(names) != 3 || records != 2 || size != 123 {
				t.Fatalf("inventory: names %d records %d bytes %d error %v", len(names), records, size, err)
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 4 {
				t.Fatalf("inventory changed files: %v %d", err, len(entries))
			}
		})
	}
}
