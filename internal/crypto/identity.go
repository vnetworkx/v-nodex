package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	ErrInvalidSeed = errors.New("identity: seed must be 32 bytes")
)

type Keypair struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
	NodeID  string
}

func Generate() (*Keypair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	k := &Keypair{
		Private: priv,
		Public:  pub,
		NodeID:  NodeIDFromPublicKey(pub),
	}
	return k, nil
}

func FromSeed(seed []byte) (*Keypair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, ErrInvalidSeed
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	return &Keypair{
		Private: priv,
		Public:  pub,
		NodeID:  NodeIDFromPublicKey(pub),
	}, nil
}

func NodeIDFromPublicKey(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])
}

func Sign(priv ed25519.PrivateKey, msg []byte) []byte {
	return ed25519.Sign(priv, msg)
}

func Verify(pub ed25519.PublicKey, msg, sig []byte) bool {
	return ed25519.Verify(pub, msg, sig)
}

func (k *Keypair) RefreshNodeID() {
	if k == nil || len(k.Public) == 0 {
		return
	}
	k.NodeID = NodeIDFromPublicKey(k.Public)
}

func (k *Keypair) Sign(msg []byte) []byte {
	if k == nil || len(k.Private) == 0 {
		return nil
	}
	return ed25519.Sign(k.Private, msg)
}

func (k *Keypair) Verify(msg, sig []byte) bool {
	if k == nil || len(k.Public) == 0 {
		return false
	}
	return ed25519.Verify(k.Public, msg, sig)
}
