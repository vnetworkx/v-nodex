package p2p

import (
	"sort"
	"sync"
	"time"
)

type Peer struct {
	ID       string    `json:"id"`
	Address  string    `json:"address,omitempty"`
	Region   string    `json:"region,omitempty"`
	Score    float64   `json:"score"`
	LastSeen time.Time `json:"last_seen"`
	Healthy  bool      `json:"healthy"`
}

type Registry struct {
	mu    sync.RWMutex
	peers map[string]Peer
}

func NewRegistry() *Registry {
	return &Registry{
		peers: make(map[string]Peer),
	}
}

func (r *Registry) Upsert(peer Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if peer.ID == "" {
		return
	}
	if peer.LastSeen.IsZero() {
		peer.LastSeen = time.Now().UTC()
	}
	r.peers[peer.ID] = peer
}

func (r *Registry) MarkSeen(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer, ok := r.peers[id]
	if !ok {
		peer = Peer{ID: id}
	}
	peer.LastSeen = time.Now().UTC()
	peer.Healthy = true
	r.peers[id] = peer
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, id)
}

func (r *Registry) Get(id string) (Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	peer, ok := r.peers[id]
	return peer, ok
}

func (r *Registry) List() []Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Peer, 0, len(r.peers))
	for _, peer := range r.peers {
		out = append(out, peer)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if !out[i].LastSeen.Equal(out[j].LastSeen) {
			return out[i].LastSeen.After(out[j].LastSeen)
		}
		return out[i].ID < out[j].ID
	})

	return out
}

func (r *Registry) PickFanout(n int, excludeIDs ...string) []Peer {
	if n <= 0 {
		return nil
	}

	excluded := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		excluded[id] = struct{}{}
	}

	all := r.List()
	out := make([]Peer, 0, n)
	for _, peer := range all {
		if _, ok := excluded[peer.ID]; ok {
			continue
		}
		out = append(out, peer)
		if len(out) >= n {
			break
		}
	}
	return out
}
