package model

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
)

func ValidateEvent(event Event) error {
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
	if event.EventHash == "" {
		return errors.New("event_hash is required")
	}
	if event.SignerID == "" {
		return errors.New("signer_id is required")
	}
	if event.SignerPublicKey != "" && event.Signature != "" {
		pubBytes, err := hex.DecodeString(event.SignerPublicKey)
		if err != nil {
			return fmt.Errorf("invalid signer public key: %w", err)
		}
		sigBytes, err := hex.DecodeString(event.Signature)
		if err != nil {
			return fmt.Errorf("invalid signature encoding: %w", err)
		}
		if len(pubBytes) != ed25519.PublicKeySize {
			return errors.New("invalid ed25519 public key size")
		}
		if !ed25519.Verify(ed25519.PublicKey(pubBytes), []byte(event.EventHash), sigBytes) {
			return errors.New("signature verification failed")
		}
	}
	if event.Operation == OpNormalize && event.InputVector.IsZero() {
		return errors.New("cannot normalize zero vector")
	}
	return nil
}
