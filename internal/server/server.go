package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"vnodex/internal/debug"
	"vnodex/internal/model"
	"vnodex/internal/peer"
	"vnodex/internal/store"
	"vnodex/internal/sync"
)

type Server struct {
	Store  *store.Store
	Peers  *peer.Registry
	Sync   *sync.Manager
	NodeID string
	Public bool
}

func New(store *store.Store, peers *peer.Registry, syncMgr *sync.Manager, nodeID string, public bool) *Server {
	return &Server{
		Store:  store,
		Peers:  peers,
		Sync:   syncMgr,
		NodeID: nodeID,
		Public: public,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/v1/state", s.getState)
	mux.HandleFunc("/v1/state/", s.getEntity)
	mux.HandleFunc("/v1/events", s.events)
	mux.HandleFunc("/v1/events/", s.getEvent)
	mux.HandleFunc("/v1/snapshots", s.snapshots)
	mux.HandleFunc("/v1/peers", s.peersHandler)

	if s.Sync != nil {
		mux.HandleFunc("/v1/sync/pull", s.Sync.PullHandler)
		mux.HandleFunc("/v1/sync/push", s.Sync.PushHandler)
		mux.HandleFunc("/v1/sync/ack", s.Sync.AckHandler)
		mux.HandleFunc("/v1/sync/peer", s.Sync.SetPeerHandler)
	}

	if !s.Public {
		mux.HandleFunc("/debug/state", debug.StateDump(s.Store))
		mux.HandleFunc("/debug/peers", debug.PeersDump(s.Peers))
		mux.HandleFunc("/debug/events", debug.EventsDump(s.Store))
	}

	return mux
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"node_id": s.NodeID,
	})
}

func (s *Server) getState(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	s.writeJSON(w, http.StatusOK, s.Store.State())
}

func (s *Server) getEntity(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/state/")
	if id == "" {
		http.Error(w, "entity id required", http.StatusBadRequest)
		return
	}

	entity, ok := s.Store.Entity(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	s.writeJSON(w, http.StatusOK, entity)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		limit := 128
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 {
				limit = n
			}
		}
		s.writeJSON(w, http.StatusOK, s.Store.ListEvents(limit))

	case http.MethodPost:
		var event model.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		appended, err := s.Store.Append(event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.writeJSON(w, http.StatusCreated, appended)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	hash := strings.TrimPrefix(r.URL.Path, "/v1/events/")
	if hash == "" {
		http.NotFound(w, r)
		return
	}

	ev, ok := s.Store.GetEvent(hash)
	if !ok {
		http.NotFound(w, r)
		return
	}

	s.writeJSON(w, http.StatusOK, ev)
}

func (s *Server) snapshots(w http.ResponseWriter, r *http.Request) {
	if s.Store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, http.StatusOK, s.Store.Snapshots())

	case http.MethodPost:
		scope := r.URL.Query().Get("scope")
		if scope == "" {
			scope = "global"
		}

		snap, err := s.Store.Snapshot(scope)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.writeJSON(w, http.StatusCreated, snap)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) peersHandler(w http.ResponseWriter, r *http.Request) {
	if s.Peers == nil {
		http.Error(w, "peer registry unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.writeJSON(w, http.StatusOK, s.Peers.List())

	case http.MethodPost:
		var peerInfo model.PeerInfo
		if err := json.NewDecoder(r.Body).Decode(&peerInfo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.Store.RegisterPeer(peerInfo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.Peers.Add(peerInfo)
		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
