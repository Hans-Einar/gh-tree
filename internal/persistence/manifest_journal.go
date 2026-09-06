package persistence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

const manifestJournalMagic = "gh-tree-manifest-journal-v1\n"
const maxManifestFrames = 8
const maxManifestJournalBytes = maxManifestFrames * maxManifestBytes

var errIncompletePreparation = errors.New("retained storage preparation is incomplete")

type manifestFrameBody struct {
	Sequence uint32
	Previous [32]byte
	Snapshot recoveryManifest
}
type manifestFrame struct {
	Body   manifestFrameBody
	Digest [32]byte
}
type manifestJournal struct {
	object *nativeObject
	frames uint32
	last   [32]byte
	bytes  int
}

// Each flushed snapshot is independently complete and hash-chained. A torn
// trailing append cannot destroy earlier recorded IDs/native facts. The final
// Ready snapshot is still mandatory before any publication. No in-place rewrite
// or replay of the destination/manifest is performed.
func (j *manifestJournal) append(ctx context.Context, snapshot recoveryManifest) error {
	if j.frames >= maxManifestFrames {
		return errors.New("manifest frame limit")
	}
	body := manifestFrameBody{j.frames + 1, j.last, snapshot}
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	frame := manifestFrame{body, sha256.Sum256(bodyRaw)}
	raw, err := json.Marshal(frame)
	if err != nil || len(raw) >= maxManifestBytes {
		return errors.Join(err, errors.New("manifest frame size limit"))
	}
	raw = append(raw, '\n')
	if j.frames == 0 {
		raw = append([]byte(manifestJournalMagic), raw...)
	}
	if len(raw) > maxManifestJournalBytes-j.bytes {
		return errors.New("manifest journal size limit")
	}
	if err := writeComplete(ctx, j.object.file, raw); err != nil {
		return err
	}
	if err := j.object.file.Sync(); err != nil {
		return err
	}
	j.frames, j.last, j.bytes = body.Sequence, frame.Digest, j.bytes+len(raw)
	return nil
}

func decodeManifestJournal(raw []byte) (last recoveryManifest, resultErr error) {
	if len(raw) > maxManifestJournalBytes || !bytes.HasPrefix(raw, []byte(manifestJournalMagic)) {
		return last, errors.New("invalid manifest journal framing")
	}
	data := raw[len(manifestJournalMagic):]
	var digest [32]byte
	sequence := uint32(0)
	for len(data) != 0 {
		end := bytes.IndexByte(data, '\n')
		if end < 0 || end >= maxManifestBytes || sequence >= maxManifestFrames {
			return last, errors.New("incomplete or excessive manifest journal tail")
		}
		line := data[:end]
		data = data[end+1:]
		if _, err := api.NewOpaqueJSON(line); err != nil {
			return last, err
		}
		var frame manifestFrame
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&frame); err != nil {
			return last, err
		}
		canonical, err := json.Marshal(frame)
		if err != nil || !sameJSONValue(line, canonical) {
			return last, errors.New("manifest frame shape changed")
		}
		body, err := json.Marshal(frame.Body)
		if err != nil || frame.Body.Sequence != sequence+1 || frame.Body.Previous != digest || frame.Digest != sha256.Sum256(body) {
			return last, errors.New("manifest frame history mismatch")
		}
		if sequence != 0 {
			if err := manifestSuccessor(last, frame.Body.Snapshot); err != nil {
				return last, err
			}
		}
		sequence, digest, last = frame.Body.Sequence, frame.Digest, frame.Body.Snapshot
	}
	if sequence == 0 {
		return last, errors.New("empty manifest journal")
	}
	return last, nil
}

func manifestSuccessor(old, next recoveryManifest) error {
	if !old.Preparing || len(old.Artifacts) != len(next.Artifacts) {
		return errors.New("manifest history altered completed preparation")
	}
	a, b := old, next
	a.Preparing, b.Preparing = false, false
	a.Artifacts, b.Artifacts = nil, nil
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	if !bytes.Equal(x, y) {
		return errors.New("manifest history changed request binding")
	}
	for i, before := range old.Artifacts {
		after := next.Artifacts[i]
		if before.Identity == (diskIdentity{}) {
			before.Identity = after.Identity
		}
		if before != after {
			return errors.New("manifest history changed artifact identity or ID")
		}
	}
	return nil
}
