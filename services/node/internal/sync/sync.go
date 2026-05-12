package sync

import (
	"encoding/json"
	"net/http"

	"vnodex/node/internal/peer"
	"vnodex/node/internal/store"
)

type Manager struct {
	store *store.Storage
	peers *peer.Registry
}

func NewManager(store *store.Storage, peers *peer.Registry) *Manager {
	return &Manager{store: store, peers: peers}
}

type SyncMessage struct {
	Type    string          `json:"type"`
	ChainID string          `json:"chain_id"`
	NodeID  string          `json:"node_id"`
	Payload json.RawMessage `json:"payload"`
}

func (m *Manager) PullHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte("sync pull not implemented"))
}

func (m *Manager) PushHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte("sync push not implemented"))
}

func (m *Manager) AckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte("sync ack not implemented"))
}
