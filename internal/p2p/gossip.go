package p2p

import (
    "context"
    "sync"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Gossip struct {
    seen map[string]struct{}
    mu   sync.Mutex
}

func NewGossip() *Gossip {
    return &Gossip{seen: map[string]struct{}{}}
}

func (g *Gossip) Seen(hash string) bool {
    g.mu.Lock()
    defer g.mu.Unlock()
    _, ok := g.seen[hash]
    return ok
}

func (g *Gossip) Mark(hash string) {
    g.mu.Lock()
    g.seen[hash] = struct{}{}
    g.mu.Unlock()
}

func (g *Gossip) Propagate(ctx context.Context, transport Transport, peers []model.Peer, rec model.EventRecord) int {
    pushed := 0
    for _, p := range peers {
        if p.BaseURL == "" {
            continue
        }
        if err := transport.PushEvent(ctx, p.BaseURL, rec); err == nil {
            pushed++
        }
    }
    return pushed
}
