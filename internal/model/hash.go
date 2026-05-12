package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func CanonicalizeMetadata(meta []KeyValue) []KeyValue {
	if len(meta) == 0 {
		return nil
	}
	out := make([]KeyValue, len(meta))
	copy(out, meta)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Key == out[j].Key {
			return out[i].Value < out[j].Value
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func CanonicalEventHash(core EventCore) (string, []byte, error) {
	core.Metadata = CanonicalizeMetadata(core.Metadata)
	payload, err := json.Marshal(core)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), payload, nil
}
