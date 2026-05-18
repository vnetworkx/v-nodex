package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	ErrNoHandler = errors.New("p2p: no handler configured")
)

type GossipHandler func(context.Context, Envelope) error

type GossipManager struct {
	mu        sync.Mutex
	selfID    string
	transport Transport
	registry  *Registry
	seen      map[string]time.Time
	ttl       time.Duration
	fanout    int
	Handler   GossipHandler
}

func NewGossipManager(selfID string, transport Transport, registry *Registry, fanout int, ttl time.Duration) *GossipManager {
	if fanout <= 0 {
		fanout = 3
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &GossipManager{
		selfID:    selfID,
		transport: transport,
		registry:  registry,
		seen:      make(map[string]time.Time),
		ttl:       ttl,
		fanout:    fanout,
	}
}

func (g *GossipManager) Publish(ctx context.Context, msgType string, payload []byte, region string, parents []string) (Envelope, error) {
	env, err := NewEnvelope(msgType, g.selfID, "", region, parents, payload, 0)
	if err != nil {
		return Envelope{}, err
	}
	if err := g.process(ctx, env, true); err != nil {
		return Envelope{}, err
	}
	return env, nil
}

func (g *GossipManager) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case env, ok := <-g.transport.Inbound():
			if !ok {
				return nil
			}
			_ = g.process(ctx, env, false)
		}
	}
}

func (g *GossipManager) process(ctx context.Context, env Envelope, local bool) error {
	if env.ID == "" {
		if err := env.Seal(); err != nil {
			return err
		}
	}

	if !g.markSeen(env.ID) {
		return nil
	}

	if env.From != "" && env.From != g.selfID {
		g.registry.MarkSeen(env.From)
	}

	if g.Handler != nil {
		if err := g.Handler(ctx, env.Clone()); err != nil {
			return err
		}
	} else if !local {
		return ErrNoHandler
	}

	if env.Hops > 32 {
		return nil
	}

	forward := env.Clone()
	forward.Hops++

	peers := g.registry.PickFanout(g.fanout, g.selfID, env.From)
	for _, peer := range peers {
		forward.To = peer.ID
		_ = g.transport.Send(ctx, peer.ID, forward)
	}

	return nil
}

func (g *GossipManager) markSeen(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	g.pruneLocked(now)

	if _, ok := g.seen[id]; ok {
		return false
	}
	g.seen[id] = now
	return true
}

func (g *GossipManager) pruneLocked(now time.Time) {
	for id, seenAt := range g.seen {
		if now.Sub(seenAt) > g.ttl {
			delete(g.seen, id)
		}
	}
}

func EnvelopeID(msgType, from, region string, payload []byte, parents []string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(msgType))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(from))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(region))
	_, _ = h.Write([]byte("|"))
	for _, p := range parents {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte(";"))
	}
	_, _ = h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
