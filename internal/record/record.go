package record

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

var (
	ErrEmptyNodeID    = errors.New("record: empty node id")
	ErrEmptyOperation = errors.New("record: empty operation")
	ErrInvalidID      = errors.New("record: invalid id")
	ErrInvalidVector  = errors.New("record: invalid vector")
	ErrInvalidHash    = errors.New("record: invalid hash")
)

type Record struct {
	ID                string     `json:"id"`
	Parents           []string   `json:"parents,omitempty"`
	Height            uint64     `json:"height"`
	NodeID            string     `json:"node_id"`
	EntityID          string     `json:"entity_id"`
	RegionID          string     `json:"region_id"`
	Operation         string     `json:"operation"`
	Vector            [3]float64 `json:"vector"`
	Payload           []byte     `json:"payload,omitempty"`
	TimestampUnixNano int64      `json:"timestamp_unix_nano"`
	Signature         []byte     `json:"signature,omitempty"`
	PublicKey         []byte     `json:"public_key,omitempty"`
}

func New(nodeID, entityID, regionID, operation string, vector [3]float64, parents []string, payload []byte, height uint64, timestampUnixNano int64) Record {
	return Record{
		Parents:           append([]string(nil), parents...),
		Height:            height,
		NodeID:            nodeID,
		EntityID:          entityID,
		RegionID:          regionID,
		Operation:         operation,
		Vector:            vector,
		Payload:           append([]byte(nil), payload...),
		TimestampUnixNano: timestampUnixNano,
	}
}

func (r Record) Clone() Record {
	out := r
	out.Parents = append([]string(nil), r.Parents...)
	out.Payload = append([]byte(nil), r.Payload...)
	out.Signature = append([]byte(nil), r.Signature...)
	out.PublicKey = append([]byte(nil), r.PublicKey...)
	return out
}

func (r Record) CanonicalBytes() ([]byte, error) {
	parents := append([]string(nil), r.Parents...)
	sort.Strings(parents)

	wire := struct {
		Parents           []string   `json:"parents,omitempty"`
		Height            uint64     `json:"height"`
		NodeID            string     `json:"node_id"`
		EntityID          string     `json:"entity_id"`
		RegionID          string     `json:"region_id"`
		Operation         string     `json:"operation"`
		Vector            [3]float64 `json:"vector"`
		Payload           []byte     `json:"payload,omitempty"`
		TimestampUnixNano int64      `json:"timestamp_unix_nano"`
	}{
		Parents:           parents,
		Height:            r.Height,
		NodeID:            r.NodeID,
		EntityID:          r.EntityID,
		RegionID:          r.RegionID,
		Operation:         r.Operation,
		Vector:            r.Vector,
		Payload:           append([]byte(nil), r.Payload...),
		TimestampUnixNano: r.TimestampUnixNano,
	}

	return json.Marshal(wire)
}

func (r Record) Hash() (string, error) {
	b, err := r.CanonicalBytes()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func (r *Record) Seal() error {
	h, err := r.Hash()
	if err != nil {
		return err
	}
	r.ID = h
	return nil
}

func (r Record) Validate() error {
	if strings.TrimSpace(r.NodeID) == "" {
		return ErrEmptyNodeID
	}
	if strings.TrimSpace(r.Operation) == "" {
		return ErrEmptyOperation
	}

	for _, v := range r.Vector {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return ErrInvalidVector
		}
	}

	if r.ID != "" {
		if len(r.ID) != 64 {
			return fmt.Errorf("%w: expected 64 hex chars, got %d", ErrInvalidID, len(r.ID))
		}
		for _, ch := range r.ID {
			switch {
			case ch >= '0' && ch <= '9':
			case ch >= 'a' && ch <= 'f':
			case ch >= 'A' && ch <= 'F':
			default:
				return ErrInvalidID
			}
		}
	}

	return nil
}

func SortAndCopyStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
