package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrUnknownPeer = errors.New("p2p: unknown peer")
	ErrClosed      = errors.New("p2p: closed")
	ErrBadEnvelope = errors.New("p2p: bad envelope")
)

type Envelope struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	From      string          `json:"from"`
	To        string          `json:"to,omitempty"`
	Region    string          `json:"region,omitempty"`
	Parents   []string        `json:"parents,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	CreatedAt int64           `json:"created_at"`
	Hops      uint32          `json:"hops"`
}

func (e Envelope) Clone() Envelope {
	out := e
	out.Parents = append([]string(nil), e.Parents...)
	if e.Payload != nil {
		out.Payload = append([]byte(nil), e.Payload...)
	}
	return out
}

func (e Envelope) CanonicalBytes() ([]byte, error) {
	wire := struct {
		Type      string   `json:"type"`
		From      string   `json:"from"`
		To        string   `json:"to,omitempty"`
		Region    string   `json:"region,omitempty"`
		Parents   []string `json:"parents,omitempty"`
		Payload   []byte   `json:"payload,omitempty"`
		CreatedAt int64    `json:"created_at"`
		Hops      uint32   `json:"hops"`
	}{
		Type:      e.Type,
		From:      e.From,
		To:        e.To,
		Region:    e.Region,
		Parents:   append([]string(nil), e.Parents...),
		Payload:   append([]byte(nil), e.Payload...),
		CreatedAt: e.CreatedAt,
		Hops:      e.Hops,
	}
	return json.Marshal(wire)
}

func (e *Envelope) Seal() error {
	b, err := e.CanonicalBytes()
	if err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	e.ID = hex.EncodeToString(sum[:])
	return nil
}

type Transport interface {
	LocalID() string
	Send(ctx context.Context, peerID string, env Envelope) error
	Broadcast(ctx context.Context, env Envelope) error
	Inbound() <-chan Envelope
	Close() error
}

type MemoryBus struct {
	mu    sync.RWMutex
	peers map[string]*MemoryTransport
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		peers: make(map[string]*MemoryTransport),
	}
}

func (b *MemoryBus) Register(id string) *MemoryTransport {
	b.mu.Lock()
	defer b.mu.Unlock()

	if existing, ok := b.peers[id]; ok {
		return existing
	}

	mt := &MemoryTransport{
		id:      id,
		bus:     b,
		inbound: make(chan Envelope, 256),
		closed:  make(chan struct{}),
	}
	b.peers[id] = mt
	return mt
}

func (b *MemoryBus) unregister(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.peers, id)
}

func (b *MemoryBus) route(peerID string) (*MemoryTransport, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	mt, ok := b.peers[peerID]
	return mt, ok
}

type MemoryTransport struct {
	id      string
	bus     *MemoryBus
	inbound chan Envelope
	closed  chan struct{}
	once    sync.Once
}

func (t *MemoryTransport) LocalID() string { return t.id }

func (t *MemoryTransport) Send(ctx context.Context, peerID string, env Envelope) error {
	target, ok := t.bus.route(peerID)
	if !ok {
		return ErrUnknownPeer
	}
	return target.deliver(ctx, env)
}

func (t *MemoryTransport) Broadcast(ctx context.Context, env Envelope) error {
	t.bus.mu.RLock()
	ids := make([]string, 0, len(t.bus.peers))
	for id := range t.bus.peers {
		if id == t.id {
			continue
		}
		ids = append(ids, id)
	}
	t.bus.mu.RUnlock()

	for _, id := range ids {
		if err := t.Send(ctx, id, env); err != nil {
			return err
		}
	}
	return nil
}

func (t *MemoryTransport) Inbound() <-chan Envelope { return t.inbound }

func (t *MemoryTransport) Close() error {
	t.once.Do(func() {
		close(t.closed)
		t.bus.unregister(t.id)
		close(t.inbound)
	})
	return nil
}

func (t *MemoryTransport) deliver(ctx context.Context, env Envelope) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.closed:
		return ErrClosed
	default:
	}

	select {
	case t.inbound <- env.Clone():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-t.closed:
		return ErrClosed
	default:
		return fmt.Errorf("p2p: inbound buffer full for peer %s", t.id)
	}
}

func NewEnvelope(msgType, from, to, region string, parents []string, payload []byte, hops uint32) (Envelope, error) {
	env := Envelope{
		Type:      msgType,
		From:      from,
		To:        to,
		Region:    region,
		Parents:   append([]string(nil), parents...),
		Payload:   append([]byte(nil), payload...),
		CreatedAt: time.Now().UnixNano(),
		Hops:      hops,
	}
	if err := env.Seal(); err != nil {
		return Envelope{}, err
	}
	return env, nil
}
