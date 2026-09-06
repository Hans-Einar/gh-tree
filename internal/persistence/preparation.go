package persistence

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

var errRecoveryCapacity = errors.New("storage recovery capacity exhausted")

// The fixed-size prefix keeps names below native component limits even when
// Composition chose a long document basename. Native names, not caller paths,
// are the sole arguments used for every artifact operation.
func recoveryPrefix(basename string) string {
	h := sha256.Sum256([]byte(nativeNameKey(basename)))
	return ".gh-tree-" + hex.EncodeToString(h[:16]) + "-"
}

func operationNonce() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// writeComplete never treats a short, nil-error write as success. Sync and
// explicit close are separate barriers owned by the request protocol.
func writeComplete(ctx context.Context, w io.Writer, raw []byte) error {
	for len(raw) != 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		n := len(raw)
		if n > 64*1024 {
			n = 64 * 1024
		}
		written, err := w.Write(raw[:n])
		if written < 0 || written > n {
			return errors.Join(err, errors.New("invalid native write length"))
		}
		if err != nil {
			return err
		}
		if written != n {
			return io.ErrShortWrite
		}
		raw = raw[n:]
	}
	return ctx.Err()
}

// inventoryRecovery is called once on a fresh retained parent while owning the
// permanent store lock. Every matching entry is opened no-follow and measured;
// malformed/orphan names also consume budget, never authorize cleanup or vanish
// merely because a complete manifest was not yet flushed when a writer crashed.
// Bounded enumeration refuses excessive unrelated entries rather than allocating
// an unbounded directory listing. No artifact data is read to measure its size.
func inventoryRecovery(ctx context.Context, parent *nativeObject, basename string, maxRecords int, maxBytes int64) ([]string, int, int64, error) {
	prefix := recoveryPrefix(basename)
	var names []string
	records := map[string]struct{}{}
	var size int64
	var observedErr error
	entries := 0
	for {
		if err := ctx.Err(); err != nil {
			return names, len(records), size, err
		}
		batch, err := parent.file.Readdirnames(128)
		entries += len(batch)
		if entries > 16384 {
			return names, len(records), size, errRecoveryCapacity
		}
		for _, name := range batch {
			if !strings.HasPrefix(nativeNameKey(name), prefix) {
				continue
			}
			if !singleName(name) {
				return names, len(records), size, errors.New("invalid recovery basename")
			}
			key := strings.TrimPrefix(nativeNameKey(name), prefix)
			// A valid operation has one 256-bit nonce, followed by a suffix.
			// Malformed matching names each count as an independent record.
			if len(key) > 64 && key[64] == '.' {
				if nonce, e := hex.DecodeString(key[:64]); e == nil && len(nonce) == 32 {
					key = key[:64]
				}
			}
			records[key] = struct{}{}
			names = append(names, name)
			if len(records) > maxRecords || len(names) > 5*maxRecords {
				return names, len(records), size, errors.Join(observedErr, errRecoveryCapacity)
			}
			object, openErr := nativeOpenDocument(parent, name)
			if openErr != nil {
				observedErr = errors.Join(observedErr, openErr)
				continue
			}
			n, sizeErr := nativeObjectSize(object)
			sizeErr = errors.Join(sizeErr, object.close())
			if sizeErr != nil {
				observedErr = errors.Join(observedErr, sizeErr)
				continue
			}
			if n < 0 || n > maxBytes-size {
				return names, len(records), size, errors.Join(observedErr, errRecoveryCapacity)
			}
			size += n
		}
		if errors.Is(err, io.EOF) {
			return names, len(records), size, observedErr
		}
		if err != nil {
			return names, len(records), size, errors.Join(observedErr, err)
		}
	}
}
