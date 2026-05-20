package crypto

import (
    "crypto/ed25519"
    crand "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "strings"
)

type Identity struct {
    PrivateKey ed25519.PrivateKey
    PublicKey  ed25519.PublicKey
    Path       string
}

type storedIdentity struct {
    PrivateKey string `json:"private_key"`
    PublicKey  string `json:"public_key"`
}

func LoadOrCreate(path string) (*Identity, error) {
    if b, err := os.ReadFile(path); err == nil {
        var s storedIdentity
        if err := json.Unmarshal(b, &s); err != nil {
            return nil, err
        }
        priv, err := base64.StdEncoding.DecodeString(s.PrivateKey)
        if err != nil {
            return nil, err
        }
        pub, err := base64.StdEncoding.DecodeString(s.PublicKey)
        if err != nil {
            return nil, err
        }
        if len(priv) != ed25519.PrivateKeySize || len(pub) != ed25519.PublicKeySize {
            return nil, errors.New("invalid identity file")
        }
        return &Identity{PrivateKey: ed25519.PrivateKey(priv), PublicKey: ed25519.PublicKey(pub), Path: path}, nil
    }

    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return nil, err
    }

    pub, priv, err := ed25519.GenerateKey(crand.Reader)
    if err != nil {
        return nil, err
    }
    s := storedIdentity{
        PrivateKey: base64.StdEncoding.EncodeToString(priv),
        PublicKey:  base64.StdEncoding.EncodeToString(pub),
    }
    b, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return nil, err
    }
    if err := os.WriteFile(path, b, 0o600); err != nil {
        return nil, err
    }
    return &Identity{PrivateKey: priv, PublicKey: pub, Path: path}, nil
}

func (i *Identity) PeerID() string {
    sum := sha256.Sum256(i.PublicKey)
    return "peer-" + strings.ToLower(base64.RawURLEncoding.EncodeToString(sum[:16]))
}

func (i *Identity) Sign(msg []byte) []byte {
    return ed25519.Sign(i.PrivateKey, msg)
}

func Verify(pub ed25519.PublicKey, msg, sig []byte) bool {
    return ed25519.Verify(pub, msg, sig)
}

func (i *Identity) PublicKeyString() string {
    return base64.StdEncoding.EncodeToString(i.PublicKey)
}
