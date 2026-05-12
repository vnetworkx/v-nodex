package debug

import (
	"encoding/json"
	"net/http"

	"vnodex/internal/peer"
	"vnodex/internal/store"
)

func StateDump(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st.State())
	}
}

func PeersDump(reg *peer.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(reg.List())
	}
}

func EventsDump(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 128
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st.ListEvents(limit))
	}
}
