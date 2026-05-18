package model

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
)

func ValidateEvent(event Event) error {
	if event.EventID == "" {
		return errors.New("event_id is required")
	}
	if event.EntityID == "" {
		return errors.New("entity_id is required")
	}
	if event.RegionID == "" {
		return errors.New("region_id is required")
	}
	if event.SpaceID == "" {
		return errors.New("space_id is required")
	}
	if event.TimeCreated.IsZero() {
		return errors.New("time_created is required")
	}
	if event.SignerID == "" {
		return errors.New("signer_id is required")
	}
	if event.EventHash == "" {
		return errors.New("event_hash is required")
	}
	if !ValidateVectorType(event.VectorType) {
		return fmt.Errorf("invalid vector_type: %s", event.VectorType)
	}
	if !ValidateOperation(event.Operation) {
		return fmt.Errorf("invalid operation: %s", event.Operation)
	}
	if len(event.ParentHashes) > 0 {
		seen := make(map[string]struct{}, len(event.ParentHashes))
		for _, parent := range event.ParentHashes {
			if parent == "" {
				return errors.New("parent_hashes contains an empty hash")
			}
			if parent == event.EventHash {
				return errors.New("event cannot reference itself as a parent")
			}
			if _, ok := seen[parent]; ok {
				return errors.New("parent_hashes contains duplicates")
			}
			seen[parent] = struct{}{}
		}
	}

	computedHash, payload, err := CanonicalEventHash(event.EventCore)
	if err != nil {
		return fmt.Errorf("unable to compute canonical event hash: %w", err)
	}
	if computedHash != event.EventHash {
		return fmt.Errorf("event hash mismatch: got %s want %s", event.EventHash, computedHash)
	}

	pubProvided := event.SignerPublicKey != ""
	sigProvided := event.Signature != ""
	if pubProvided != sigProvided {
		return errors.New("signer_public_key and signature must either both be set or both be empty")
	}
	if pubProvided {
		pubBytes, err := hex.DecodeString(event.SignerPublicKey)
		if err != nil {
			return fmt.Errorf("invalid signer public key: %w", err)
		}
		if len(pubBytes) != ed25519.PublicKeySize {
			return errors.New("invalid ed25519 public key size")
		}
		sigBytes, err := hex.DecodeString(event.Signature)
		if err != nil {
			return fmt.Errorf("invalid signature encoding: %w", err)
		}
		if !ed25519.Verify(ed25519.PublicKey(pubBytes), payload, sigBytes) {
			return errors.New("signature verification failed")
		}
	} else {
		return errors.New("signature and signer_public_key are required")
	}

	if event.Operation == OpNormalize && event.InputVector.IsZero() {
		return errors.New("cannot normalize zero vector")
	}

	return nil
}