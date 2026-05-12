package peer

import "sync"

type Peer struct {
	ID         string `json:"id"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	Confirmed  bool   `json:"confirmed"`
	LastHead   string `json:"last_head"`
	LastSeenAt string `json:"last_seen_at"`
}

type Registry struct {
	mu    sync.RWMutex
	self  string
	peers map[string]Peer
}

func NewRegistry(self string) *Registry {
	return &Registry{self: self, peers: map[string]Peer{}}
}

func (r *Registry) Add(peer Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers[peer.ID] = peer
}

func (r *Registry) List() []Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Peer, 0, len(r.peers))
	for _, p := range r.peers {
		out = append(out, p)
	}
	return out
}
