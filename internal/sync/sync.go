package sync

import (
	"encoding/json"
	"net/http"
	"strconv"

	"vnodex/internal/model"
	"vnodex/internal/peer"
	"vnodex/internal/store"
)

type Manager struct {
	store *store.Store
	peers *peer.Registry
}

func NewManager(st *store.Store, peers *peer.Registry) *Manager {
	return &Manager{store: st, peers: peers}
}

type pullRequest struct {
	Cursor uint64 `json:"cursor"`
	Limit  int    `json:"limit"`
}

type pullResponse struct {
	NodeID      string           `json:"node_id"`
	HeadHash    string           `json:"head_hash"`
	LatestOrder uint64           `json:"latest_order"`
	Events      []model.Event    `json:"events"`
	Snapshot    *model.Snapshot  `json:"snapshot,omitempty"`
	Peers       []model.PeerInfo `json:"peers"`
}

type pushRequest struct {
	NodeID   string          `json:"node_id"`
	Events   []model.Event   `json:"events"`
	Snapshot *model.Snapshot `json:"snapshot,omitempty"`
}

type ackResponse struct {
	Accepted    int    `json:"accepted"`
	Rejected    int    `json:"rejected"`
	CurrentHead string `json:"current_head"`
}

func (m *Manager) PullHandler(w http.ResponseWriter, r *http.Request) {
	var req pullRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.Limit <= 0 {
		req.Limit = 128
	}
	state := m.store.State()
	resp := pullResponse{
		NodeID:      m.peers.NodeID(),
		HeadHash:    state.LatestHash,
		LatestOrder: state.LatestOrder,
		Events:      m.store.EventsSince(req.Cursor, req.Limit),
		Peers:       m.peers.List(),
	}
	if snap := state.Snapshot; snap != nil {
		resp.Snapshot = snap
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (m *Manager) PushHandler(w http.ResponseWriter, r *http.Request) {
	var req pushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	accepted, rejected := 0, 0
	for _, ev := range req.Events {
		if _, err := m.store.Append(ev); err != nil {
			rejected++
			continue
		}
		accepted++
	}
	if req.Snapshot != nil {
		_, _ = m.store.Snapshot(req.Snapshot.Scope)
	}
	resp := ackResponse{Accepted: accepted, Rejected: rejected, CurrentHead: m.store.State().LatestHash}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (m *Manager) AckHandler(w http.ResponseWriter, r *http.Request) {
	state := m.store.State()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"node_id":      m.peers.NodeID(),
		"current_head": state.LatestHash,
		"latest_order": state.LatestOrder,
		"peers":        m.peers.List(),
	})
}

func (m *Manager) SetPeerHandler(w http.ResponseWriter, r *http.Request) {
	var peerInfo model.PeerInfo
	if err := json.NewDecoder(r.Body).Decode(&peerInfo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if peerInfo.PeerID == "" {
		peerInfo.PeerID = r.URL.Query().Get("peer_id")
	}
	if peerInfo.PeerID == "" {
		http.Error(w, "peer_id is required", http.StatusBadRequest)
		return
	}
	if err := m.store.RegisterPeer(peerInfo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.peers.Add(peerInfo)
	w.WriteHeader(http.StatusCreated)
}

func (m *Manager) PullSinceHandler(w http.ResponseWriter, r *http.Request) {
	cursor, _ := strconv.ParseUint(r.URL.Query().Get("cursor"), 10, 64)
	limit := 128
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	state := m.store.State()
	resp := pullResponse{
		NodeID:      m.peers.NodeID(),
		HeadHash:    state.LatestHash,
		LatestOrder: state.LatestOrder,
		Events:      m.store.EventsSince(cursor, limit),
		Peers:       m.peers.List(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
