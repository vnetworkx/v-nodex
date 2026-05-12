package peer

import (
	"sort"
	"sync"
	"time"

	"vnodex/internal/model"
)

type Registry struct {
	mu     sync.RWMutex
	nodeID string
	peers  map[string]model.PeerInfo
}

func NewRegistry(nodeID string) *Registry {
	return &Registry{nodeID: nodeID, peers: make(map[string]model.PeerInfo)}
}

func (r *Registry) NodeID() string { return r.nodeID }

func (r *Registry) Add(peer model.PeerInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	peer.LastSeen = time.Now().UTC()
	r.peers[peer.PeerID] = peer
}

func (r *Registry) Remove(peerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, peerID)
}

func (r *Registry) List() []model.PeerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.PeerInfo, 0, len(r.peers))
	for _, peer := range r.peers {
		out = append(out, peer)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PeerID < out[j].PeerID })
	return out
}
