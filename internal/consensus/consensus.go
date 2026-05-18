package consensus

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"

	identity "vnodex/internal/crypto"
)

var (
	ErrUnknownValidator = errors.New("consensus: unknown validator")
	ErrInvalidVote      = errors.New("consensus: invalid vote")
	ErrDuplicateVote    = errors.New("consensus: duplicate vote")
)

type Validator struct {
	ID        string `json:"id"`
	Weight    uint64 `json:"weight"`
	PublicKey []byte `json:"public_key,omitempty"`
}

type Quorum struct {
	Validators  map[string]Validator `json:"validators"`
	TotalWeight uint64               `json:"total_weight"`
	Threshold   float64              `json:"threshold"`
}

func NewQuorum(validators []Validator, threshold float64) *Quorum {
	if threshold <= 0 {
		threshold = 2.0 / 3.0
	}
	q := &Quorum{
		Validators: make(map[string]Validator, len(validators)),
		Threshold:  threshold,
	}
	for _, v := range validators {
		if v.Weight == 0 {
			v.Weight = 1
		}
		q.Validators[v.ID] = v
		q.TotalWeight += v.Weight
	}
	return q
}

func (q *Quorum) ThresholdWeight() uint64 {
	if q == nil || q.TotalWeight == 0 {
		return 0
	}
	return uint64(math.Ceil(float64(q.TotalWeight) * q.Threshold))
}

func (q *Quorum) Validator(id string) (Validator, bool) {
	if q == nil {
		return Validator{}, false
	}
	v, ok := q.Validators[id]
	return v, ok
}

type Vote struct {
	RecordID    string `json:"record_id"`
	ValidatorID string `json:"validator_id"`
	Approve     bool   `json:"approve"`
	Height      uint64 `json:"height"`
	Signature   []byte `json:"signature,omitempty"`
}

func (v Vote) CanonicalBytes() ([]byte, error) {
	wire := struct {
		RecordID    string `json:"record_id"`
		ValidatorID string `json:"validator_id"`
		Approve     bool   `json:"approve"`
		Height      uint64 `json:"height"`
	}{
		RecordID:    v.RecordID,
		ValidatorID: v.ValidatorID,
		Approve:     v.Approve,
		Height:      v.Height,
	}
	return json.Marshal(wire)
}

type Decision struct {
	RecordID        string `json:"record_id"`
	ApproveWeight   uint64 `json:"approve_weight"`
	RejectWeight    uint64 `json:"reject_weight"`
	ThresholdWeight uint64 `json:"threshold_weight"`
	TotalWeight     uint64 `json:"total_weight"`
	Finalized       bool   `json:"finalized"`
}

type Tracker struct {
	mu        sync.RWMutex
	quorum    *Quorum
	votes     map[string]map[string]Vote
	finalized map[string]bool
}

func NewTracker(q *Quorum) *Tracker {
	return &Tracker{
		quorum:    q,
		votes:     make(map[string]map[string]Vote),
		finalized: make(map[string]bool),
	}
}

func (t *Tracker) CastVote(v Vote) error {
	if v.RecordID == "" || v.ValidatorID == "" {
		return ErrInvalidVote
	}
	if t.quorum == nil {
		return errors.New("consensus: missing quorum")
	}

	validator, ok := t.quorum.Validator(v.ValidatorID)
	if !ok {
		return ErrUnknownValidator
	}

	msg, err := v.CanonicalBytes()
	if err != nil {
		return err
	}
	if len(validator.PublicKey) > 0 && len(v.Signature) > 0 {
		if !identity.Verify(ed25519.PublicKey(validator.PublicKey), msg, v.Signature) {
			return ErrInvalidVote
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.votes[v.RecordID]; !ok {
		t.votes[v.RecordID] = make(map[string]Vote)
	}
	if _, exists := t.votes[v.RecordID][v.ValidatorID]; exists {
		return ErrDuplicateVote
	}
	t.votes[v.RecordID][v.ValidatorID] = v

	if t.Decision(v.RecordID).Finalized {
		t.finalized[v.RecordID] = true
	}

	return nil
}

func (t *Tracker) Decision(recordID string) Decision {
	t.mu.RLock()
	defer t.mu.RUnlock()

	dec := Decision{
		RecordID:        recordID,
		ThresholdWeight: t.quorum.ThresholdWeight(),
		TotalWeight:     t.quorum.TotalWeight,
	}

	votes := t.votes[recordID]
	for validatorID, v := range votes {
		val, ok := t.quorum.Validator(validatorID)
		if !ok {
			continue
		}
		if v.Approve {
			dec.ApproveWeight += val.Weight
		} else {
			dec.RejectWeight += val.Weight
		}
	}

	dec.Finalized = dec.ApproveWeight >= dec.ThresholdWeight && dec.ThresholdWeight > 0
	return dec
}

func (t *Tracker) IsFinalized(recordID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.finalized[recordID]
}

type HeadCandidate struct {
	RecordID          string `json:"record_id"`
	Height            uint64 `json:"height"`
	TimestampUnixNano int64  `json:"timestamp_unix_nano"`
	Weight            uint64 `json:"weight"`
	Finalized         bool   `json:"finalized"`
}

func CanonicalHead(candidates []HeadCandidate) (HeadCandidate, bool) {
	if len(candidates) == 0 {
		return HeadCandidate{}, false
	}

	out := append([]HeadCandidate(nil), candidates...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Finalized != out[j].Finalized {
			return out[i].Finalized
		}
		if out[i].Height != out[j].Height {
			return out[i].Height > out[j].Height
		}
		if out[i].Weight != out[j].Weight {
			return out[i].Weight > out[j].Weight
		}
		if out[i].TimestampUnixNano != out[j].TimestampUnixNano {
			return out[i].TimestampUnixNano > out[j].TimestampUnixNano
		}
		return out[i].RecordID < out[j].RecordID
	})

	return out[0], true
}

func (d Decision) String() string {
	return fmt.Sprintf(
		"Decision(record=%s approve=%d reject=%d threshold=%d finalized=%t)",
		d.RecordID, d.ApproveWeight, d.RejectWeight, d.ThresholdWeight, d.Finalized,
	)
}
